// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// normalizes the LLM responses

func normalizeJSON(raw string, t reflect.Type, enums map[string][]string) string {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	normalized := normalizeValue(v, t, "", enums)
	out, err := json.Marshal(normalized)
	if err != nil {
		return raw
	}
	return string(out)
}

func normalizeValue(v any, t reflect.Type, field string, enums map[string][]string) any {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok {
			return v
		}
		for i := range t.NumField() {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			name, _ := parseTag(f.Tag.Get("json"), f.Name)
			if name == "-" {
				continue
			}
			if cur, present := m[name]; present {
				m[name] = normalizeValue(cur, f.Type, name, enums)
			}
		}
		return m

	case reflect.Slice:
		arr, ok := v.([]any)
		if !ok {
			return v
		}
		elem := t.Elem()
		for i, e := range arr {
			arr[i] = normalizeValue(e, elem, field, enums)
		}
		return arr

	case reflect.String:
		return normalizeString(v, field, enums)

	case reflect.Bool:
		return coerceBool(v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return coerceNumber(v)

	default:
		return v
	}
}

func normalizeString(v any, field string, enums map[string][]string) any {
	var s string
	switch x := v.(type) {
	case string:
		s = strings.TrimSpace(x)
	case bool:
		s = strconv.FormatBool(x)
	case float64:
		s = strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return v
	}
	if allowed, ok := enums[field]; ok {
		if canon, matched := canonicalizeEnum(s, allowed); matched {
			s = canon
		}
	}
	return s
}

func coerceBool(v any) any {
	if s, ok := v.(string); ok {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "true":
			return true
		case "false":
			return false
		}
	}
	return v
}

func coerceNumber(v any) any {
	if s, ok := v.(string); ok {
		if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
			return f
		}
	}
	return v
}

func canonicalizeEnum(value string, allowed []string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return value, false
	}
	for _, a := range allowed {
		if trimmed == a {
			return a, trimmed != value
		}
	}
	lower := strings.ToLower(trimmed)
	for _, a := range allowed {
		if lower == strings.ToLower(a) {
			return a, true
		}
	}
	for _, a := range allowed {
		la := strings.ToLower(a)
		if strings.HasPrefix(la, lower) || strings.HasPrefix(lower, la) {
			return a, true
		}
	}
	return value, false
}
