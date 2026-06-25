// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"strings"
	"testing"
)

func TestNewShellIntegrationCmd(t *testing.T) {
	cmd := newShellIntegrationCmd()
	if cmd.Use != "shell-integration" {
		t.Errorf("expected Use 'shell-integration', got %q", cmd.Use)
	}
	if !strings.Contains(strings.ToLower(cmd.Short), "shell") {
		t.Error("expected Short to mention shell")
	}
}

func TestShellIntegrationShellFlag(t *testing.T) {
	cmd := newShellIntegrationCmd()
	f := cmd.Flags().Lookup("shell")
	if f == nil {
		t.Fatal("expected --shell flag")
	}
	if f.DefValue != "" {
		t.Errorf("expected default empty shell, got %q", f.DefValue)
	}
}

func TestShellIntegrationNoArgs(t *testing.T) {
	cmd := newShellIntegrationCmd()
	if cmd.Args == nil {
		t.Error("expected Args to be set (NoArgs)")
	}
}

func TestShellSnippets_Bash(t *testing.T) {
	s, ok := shellSnippets["bash"]
	if !ok {
		t.Fatal("expected bash snippet")
	}
	if !strings.Contains(s.code, "hosomaki explain") {
		t.Error("bash snippet should reference hosomaki explain")
	}
}

func TestShellSnippets_Zsh(t *testing.T) {
	if _, ok := shellSnippets["zsh"]; !ok {
		t.Fatal("expected zsh snippet")
	}
}

func TestShellSnippets_Fish(t *testing.T) {
	s, ok := shellSnippets["fish"]
	if !ok {
		t.Fatal("expected fish snippet")
	}
	if !strings.Contains(s.code, "function explain") {
		t.Error("fish snippet should define function explain")
	}
}

func TestDetectShell_Bash(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	if got := detectShell(); got != "bash" {
		t.Fatalf("expected bash, got %q", got)
	}
}

func TestDetectShell_Zsh(t *testing.T) {
	t.Setenv("SHELL", "/usr/bin/zsh")
	if got := detectShell(); got != "zsh" {
		t.Fatalf("expected zsh, got %q", got)
	}
}

func TestDetectShell_Fish(t *testing.T) {
	t.Setenv("SHELL", "/usr/bin/fish")
	if got := detectShell(); got != "fish" {
		t.Fatalf("expected fish, got %q", got)
	}
}

func TestDetectShell_Fallback(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")
	if got := detectShell(); got != "bash" {
		t.Fatalf("expected bash fallback, got %q", got)
	}
}

func TestDetectShell_Empty(t *testing.T) {
	t.Setenv("SHELL", "")
	if got := detectShell(); got != "bash" {
		t.Fatalf("expected bash fallback for empty SHELL, got %q", got)
	}
}

func TestShellIntegration_UnsupportedShell(t *testing.T) {
	cmd := newShellIntegrationCmd()
	cmd.SetArgs([]string{"--shell", "tcsh"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for unsupported shell tcsh")
	} else if !strings.Contains(err.Error(), "unsupported shell") {
		t.Fatalf("unexpected error: %v", err)
	}
}
