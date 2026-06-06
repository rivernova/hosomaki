// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// validation logic for JSON output schemas from LLM providers

type ValidationResult struct {
	structuralErrors []string
	semanticErrors   []string
}

func (r *ValidationResult) Valid() bool {
	return len(r.structuralErrors) == 0 && len(r.semanticErrors) == 0
}

func (r *ValidationResult) StructurallyValid() bool {
	return len(r.structuralErrors) == 0
}

func (r *ValidationResult) Errors() []string {
	out := make([]string, 0, len(r.structuralErrors)+len(r.semanticErrors))
	out = append(out, r.structuralErrors...)
	out = append(out, r.semanticErrors...)
	return out
}

func (r *ValidationResult) StructuralErrors() []string { return r.structuralErrors }

func (r *ValidationResult) SemanticErrors() []string { return r.semanticErrors }

func (r *ValidationResult) Error() string {
	return "validation failed:\n  - " + strings.Join(r.Errors(), "\n  - ")
}

type StructValidator[T any] struct {
	SemanticCheck func(T) []string
}

func (v StructValidator[T]) Validate(raw string) *ValidationResult {
	result := &ValidationResult{}

	// syntactic
	var value T
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		result.structuralErrors = append(result.structuralErrors,
			fmt.Sprintf("JSON syntax error: %v", err))
		return result
	}

	// structural
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &rawMap); err != nil {
		result.structuralErrors = append(result.structuralErrors,
			fmt.Sprintf("top-level must be a JSON object: %v", err))
		return result
	}

	validateStruct(reflect.TypeOf(value), rawMap, "", result)

	// semantic
	if result.StructurallyValid() && v.SemanticCheck != nil {
		result.semanticErrors = append(result.semanticErrors, v.SemanticCheck(value)...)
	}

	return result
}

func (v StructValidator[T]) Decode(raw string) (T, error) {
	var value T
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return value, fmt.Errorf("decode: %w", err)
	}
	return value, nil
}

func validateStruct(t reflect.Type, rawMap map[string]json.RawMessage, path string, result *ValidationResult) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	expected := make(map[string]struct{}, t.NumField())

	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonName, omitempty := parseTag(field.Tag.Get("json"), field.Name)
		if jsonName == "-" {
			continue
		}
		expected[jsonName] = struct{}{}

		fieldPath := jsonName
		if path != "" {
			fieldPath = path + "." + jsonName
		}

		rawVal, present := rawMap[jsonName]
		if !present {
			if !omitempty {
				result.structuralErrors = append(result.structuralErrors,
					fmt.Sprintf("field %q is missing", fieldPath))
			}
			continue
		}

		if !omitempty && strings.TrimSpace(string(rawVal)) == "null" && field.Type.Kind() != reflect.Slice {
			result.structuralErrors = append(result.structuralErrors,
				fmt.Sprintf("field %q cannot be null", fieldPath))
			continue
		}

		validateField(field.Type, rawVal, fieldPath, result)
	}

	for key := range rawMap {
		if _, ok := expected[key]; !ok {
			fieldPath := key
			if path != "" {
				fieldPath = path + "." + key
			}
			result.structuralErrors = append(result.structuralErrors,
				fmt.Sprintf("unexpected field %q", fieldPath))
		}
	}
}

func validateField(ft reflect.Type, rawVal json.RawMessage, path string, result *ValidationResult) {
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}

	trimmed := strings.TrimSpace(string(rawVal))

	switch ft.Kind() {
	case reflect.String:
		var s string
		if err := json.Unmarshal(rawVal, &s); err != nil {
			result.structuralErrors = append(result.structuralErrors,
				fmt.Sprintf("field %q: expected string, got %s", path, jsonKindOf(trimmed)))
		}

	case reflect.Bool:
		var b bool
		if err := json.Unmarshal(rawVal, &b); err != nil {
			result.structuralErrors = append(result.structuralErrors,
				fmt.Sprintf("field %q: expected bool, got %s", path, jsonKindOf(trimmed)))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		var n json.Number
		if err := json.Unmarshal(rawVal, &n); err != nil {
			result.structuralErrors = append(result.structuralErrors,
				fmt.Sprintf("field %q: expected number, got %s", path, jsonKindOf(trimmed)))
		}

	case reflect.Slice:
		validateSliceField(ft, rawVal, trimmed, path, result)

	case reflect.Struct:
		validateObjectField(ft, rawVal, trimmed, path, result)
	default:
		panic("unhandled default case")
	}
}

func validateSliceField(ft reflect.Type, rawVal json.RawMessage, trimmed, path string, result *ValidationResult) {
	if trimmed == "null" {
		return
	}
	if !strings.HasPrefix(trimmed, "[") {
		result.structuralErrors = append(result.structuralErrors,
			fmt.Sprintf("field %q: expected array, got %s", path, jsonKindOf(trimmed)))
		return
	}

	var elems []json.RawMessage
	if err := json.Unmarshal(rawVal, &elems); err != nil {
		result.structuralErrors = append(result.structuralErrors,
			fmt.Sprintf("field %q: invalid array: %v", path, err))
		return
	}

	elemType := ft.Elem()
	isStruct := elemType.Kind() == reflect.Struct ||
		(elemType.Kind() == reflect.Pointer && elemType.Elem().Kind() == reflect.Struct)

	for idx, elem := range elems {
		elemPath := fmt.Sprintf("%s[%d]", path, idx)
		if isStruct {
			var elemMap map[string]json.RawMessage
			if err := json.Unmarshal(elem, &elemMap); err != nil {
				result.structuralErrors = append(result.structuralErrors,
					fmt.Sprintf("%s: invalid object: %v", elemPath, err))
				continue
			}
			validateStruct(elemType, elemMap, elemPath, result)
		} else {
			validateField(elemType, elem, elemPath, result)
		}
	}
}

func validateObjectField(ft reflect.Type, rawVal json.RawMessage, trimmed, path string, result *ValidationResult) {
	if !strings.HasPrefix(trimmed, "{") {
		result.structuralErrors = append(result.structuralErrors,
			fmt.Sprintf("field %q: expected object, got %s", path, jsonKindOf(trimmed)))
		return
	}
	var nested map[string]json.RawMessage
	if err := json.Unmarshal(rawVal, &nested); err != nil {
		result.structuralErrors = append(result.structuralErrors,
			fmt.Sprintf("field %q: invalid object: %v", path, err))
		return
	}
	validateStruct(ft, nested, path, result)
}

func parseTag(tag, fieldName string) (name string, omitempty bool) {
	if tag == "" {
		return fieldName, false
	}
	parts := strings.SplitN(tag, ",", 2)
	name = parts[0]
	if name == "" {
		name = fieldName
	}
	omitempty = len(parts) == 2 && strings.Contains(parts[1], "omitempty")
	return name, omitempty
}

func jsonKindOf(s string) string {
	if len(s) == 0 {
		return "empty"
	}
	switch s[0] {
	case '"':
		return "string"
	case '{':
		return "object"
	case '[':
		return "array"
	case 't', 'f':
		return "bool"
	case 'n':
		return "null"
	default:
		return "number"
	}
}
