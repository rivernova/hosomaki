// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"strings"
	"testing"
)

// unit testing for validation of LLM output against a schema

type validatorItem struct {
	Label  string `json:"label"`
	Active bool   `json:"active"`
}

type validatorSchema struct {
	Name  string          `json:"name"`
	Count int             `json:"count"`
	Items []validatorItem `json:"items"`
	Note  string          `json:"note,omitempty"`
}

func newValidator() StructValidator[validatorSchema] {
	return StructValidator[validatorSchema]{}
}

func TestValidator_Valid(t *testing.T) {
	raw := `{"name":"foo","count":3,"items":[{"label":"a","active":true}]}`
	r := newValidator().Validate(raw)
	if !r.Valid() {
		t.Fatalf("expected valid, got errors: %v", r.Errors())
	}
}

func TestValidator_EmptyArrayIsValid(t *testing.T) {
	raw := `{"name":"x","count":0,"items":[]}`
	r := newValidator().Validate(raw)
	if !r.Valid() {
		t.Fatalf("empty array should be valid structurally, got: %v", r.Errors())
	}
}

func TestValidator_OmitemptyAbsentIsValid(t *testing.T) {
	raw := `{"name":"bar","count":0,"items":[]}`
	r := newValidator().Validate(raw)
	if !r.Valid() {
		t.Fatalf("absent omitempty field should be valid, got: %v", r.Errors())
	}
}

func TestValidator_MissingRequiredField(t *testing.T) {
	raw := `{"count":1,"items":[]}`
	r := newValidator().Validate(raw)
	if r.Valid() {
		t.Fatal("expected invalid (missing 'name'), got valid")
	}
	if r.StructurallyValid() {
		t.Fatal("StructurallyValid() must be false when a required field is missing")
	}
	if !containsError(r.StructuralErrors(), `"name" is missing`) {
		t.Fatalf("expected missing-field error for 'name', got: %v", r.StructuralErrors())
	}
}

func TestValidator_ArrayItemMissingField(t *testing.T) {
	raw := `{"name":"x","count":1,"items":[{"active":true}]}`
	r := newValidator().Validate(raw)
	if r.Valid() {
		t.Fatal("expected invalid (items[0].label missing), got valid")
	}
	if !containsError(r.StructuralErrors(), `"items[0].label" is missing`) {
		t.Fatalf("expected items[0].label error, got: %v", r.StructuralErrors())
	}
}

func TestValidator_StringFieldGotNumber(t *testing.T) {
	raw := `{"name":42,"count":1,"items":[]}`
	r := newValidator().Validate(raw)
	if r.Valid() || r.StructurallyValid() {
		t.Fatal("expected structural invalid (name is number), got valid")
	}
}

func TestValidator_EmptyStringIsInvalid(t *testing.T) {
	raw := `{"name":"","count":1,"items":[]}`
	r := newValidator().Validate(raw)
	if r.Valid() || r.StructurallyValid() {
		t.Fatal("expected structural invalid (name is empty string), got valid")
	}
}

func TestValidator_BoolFieldGotString(t *testing.T) {
	raw := `{"name":"x","count":1,"items":[{"label":"a","active":"yes"}]}`
	r := newValidator().Validate(raw)
	if r.Valid() || r.StructurallyValid() {
		t.Fatal("expected structural invalid (active is string not bool), got valid")
	}
}

func TestValidator_SyntaxError(t *testing.T) {
	r := newValidator().Validate(`{broken`)
	if r.Valid() || r.StructurallyValid() {
		t.Fatal("expected invalid for syntax error, got valid")
	}
}

func TestValidator_EmptyInput(t *testing.T) {
	r := newValidator().Validate(``)
	if r.Valid() || r.StructurallyValid() {
		t.Fatal("expected invalid for empty input, got valid")
	}
}

func TestValidator_SemanticCheckFires(t *testing.T) {
	v := StructValidator[validatorSchema]{
		SemanticCheck: func(s validatorSchema) []string {
			if len(s.Items) == 0 {
				return []string{"items must not be empty"}
			}
			return nil
		},
	}
	raw := `{"name":"x","count":0,"items":[]}`
	r := v.Validate(raw)

	if r.Valid() {
		t.Fatal("expected invalid from semantic check, got valid")
	}

	if !r.StructurallyValid() {
		t.Fatal("StructurallyValid() must be true when only semantic check failed")
	}
	if !containsError(r.SemanticErrors(), "items must not be empty") {
		t.Fatalf("expected semantic error, got: %v", r.SemanticErrors())
	}
	if len(r.StructuralErrors()) != 0 {
		t.Fatalf("StructuralErrors() must be empty for a semantic-only failure, got: %v", r.StructuralErrors())
	}
}

func TestValidator_SemanticErrorsNotInStructuralErrors(t *testing.T) {
	v := StructValidator[validatorSchema]{
		SemanticCheck: func(s validatorSchema) []string {
			return []string{"custom semantic rule violated"}
		},
	}
	r := v.Validate(`{"name":"x","count":0,"items":[]}`)

	if len(r.StructuralErrors()) != 0 {
		t.Fatalf("structural errors must be empty for semantic-only failure, got: %v", r.StructuralErrors())
	}
	if !containsError(r.SemanticErrors(), "custom semantic rule violated") {
		t.Fatalf("semantic error missing, got: %v", r.SemanticErrors())
	}
}

func TestValidator_SemanticCheckNotCalledOnStructuralFailure(t *testing.T) {
	called := false
	v := StructValidator[validatorSchema]{
		SemanticCheck: func(s validatorSchema) []string {
			called = true
			return nil
		},
	}
	r := v.Validate(`{"count":1,"items":[]}`)
	if r.Valid() {
		t.Fatal("expected invalid")
	}
	if called {
		t.Fatal("SemanticCheck must not be called when structural validation fails")
	}
}

func TestValidator_ErrorsCombinesStructuralAndSemantic(t *testing.T) {
	r := &ValidationResult{
		structuralErrors: []string{"structural error"},
		semanticErrors:   []string{"semantic error"},
	}
	errs := r.Errors()
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
	if errs[0] != "structural error" {
		t.Fatalf("structural errors must come first, got: %v", errs)
	}
}

func TestValidationResult_Error(t *testing.T) {
	r := &ValidationResult{structuralErrors: []string{"field \"x\" is missing"}}
	msg := r.Error()
	if !strings.Contains(msg, "validation failed") {
		t.Fatalf("Error() missing expected prefix, got: %q", msg)
	}
}

func TestValidationResult_ValidIsTrue(t *testing.T) {
	r := &ValidationResult{}
	if !r.Valid() {
		t.Fatal("empty errors should be valid")
	}
	if !r.StructurallyValid() {
		t.Fatal("empty errors should be structurally valid")
	}
}

func TestValidator_Decode(t *testing.T) {
	raw := `{"name":"decoded","count":7,"items":[]}`
	v, err := newValidator().Decode(raw)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if v.Name != "decoded" || v.Count != 7 {
		t.Fatalf("Decode() unexpected value: %+v", v)
	}
}

func containsError(errs []string, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
