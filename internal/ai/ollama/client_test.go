// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// unit tests for Ping

func TestPing_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("Ping: unexpected path %q, want /api/tags", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL, "gemma3:4b", 30*time.Second)
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() returned unexpected error: %v", err)
	}
}

func TestPing_Unreachable(t *testing.T) {
	c := New("http://localhost:1", "gemma3:4b", 30*time.Second)
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("Ping() expected error for unreachable endpoint, got nil")
	}
	if !strings.Contains(err.Error(), "ollama serve") {
		t.Fatalf("Ping() error should hint at 'ollama serve', got: %q", err.Error())
	}
}

func TestPing_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := New(srv.URL, "gemma3:4b", 30*time.Second)
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("Ping() expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Fatalf("Ping() error should mention the status code, got: %q", err.Error())
	}
}

func TestPing_RespectsShortTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := New(srv.URL, "gemma3:4b", 30*time.Second)
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("Ping() expected timeout error for hanging server, got nil")
	}
}
