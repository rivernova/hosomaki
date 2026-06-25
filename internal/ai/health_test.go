// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ai

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type healthStubProvider struct {
	err error
}

func (h *healthStubProvider) Generate(context.Context, string) (string, error) { return "", nil }
func (h *healthStubProvider) GenerateJSON(context.Context, string, func()) (string, error) {
	return "", nil
}
func (h *healthStubProvider) GenerateStream(context.Context, string, func(), io.Writer) (string, error) {
	return "", nil
}
func (h *healthStubProvider) HealthCheck(context.Context) error { return h.err }

func TestCheckProviderHealth_OK(t *testing.T) {
	p := &healthStubProvider{}
	if err := CheckProviderHealth(context.Background(), p, "ollama"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCheckProviderHealth_FailureWrapsProviderName(t *testing.T) {
	want := errors.New("connection refused")
	p := &healthStubProvider{err: want}
	err := CheckProviderHealth(context.Background(), p, "ollama")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "ollama") {
		t.Fatalf("expected provider name in error, got %q", err.Error())
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error %v, got %v", want, err)
	}
}
