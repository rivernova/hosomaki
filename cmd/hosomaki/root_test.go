// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hosomaki

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rivernova/hosomaki/internal/config"
)

// unit testing for root command setup

func TestExecuteExposesVersionFlag(t *testing.T) {
	var out bytes.Buffer

	provider = nil
	version = ""
	rootCmd.Version = ""
	rootCmd.SetOut(&out)
	rootCmd.SetArgs([]string{"--version"})
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
		rootCmd.Version = ""
		version = ""
		provider = nil
	})

	Execute("test-version")

	got := out.String()
	if !strings.Contains(got, "test-version") {
		t.Fatalf("expected version output to contain %q, got %q", "test-version", got)
	}
	if provider != nil {
		t.Fatal("expected --version to skip provider initialization")
	}
}

func TestRootCmd_HasProviderFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("provider")
	if f == nil {
		t.Fatal("expected --provider persistent flag to be registered")
	}
	if f.Usage != "override ai.provider from config at runtime" {
		t.Fatalf("unexpected --provider usage: %q", f.Usage)
	}
}

func TestApplyAIProviderOverride(t *testing.T) {
	cfg := config.Config{AI: config.AIConfig{Provider: "from-file"}}

	applyAIProviderOverride(&cfg, "ollama", true)
	if cfg.AI.Provider != "ollama" {
		t.Fatalf("expected override to ollama, got %q", cfg.AI.Provider)
	}

	cfg = config.Config{AI: config.AIConfig{Provider: "from-file"}}
	applyAIProviderOverride(&cfg, "ignored", false)
	if cfg.AI.Provider != "from-file" {
		t.Fatalf("expected config provider unchanged, got %q", cfg.AI.Provider)
	}
}

func TestRootCmd_HasModelFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("model")
	if f == nil {
		t.Fatal("expected --model persistent flag to be registered")
	}
}

func TestApplyAIModelOverride(t *testing.T) {
	cfg := config.Config{AI: config.AIConfig{Model: "llama3.2:3b"}}

	applyAIModelOverride(&cfg, "qwen2.5:7b", true)
	if cfg.AI.Model != "qwen2.5:7b" {
		t.Fatalf("expected model override, got %q", cfg.AI.Model)
	}

	cfg = config.Config{AI: config.AIConfig{Model: "llama3.2:3b"}}
	applyAIModelOverride(&cfg, "ignored", false)
	if cfg.AI.Model != "llama3.2:3b" {
		t.Fatalf("expected config model unchanged, got %q", cfg.AI.Model)
	}
}
