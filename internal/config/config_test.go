// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

// uni testing for default config

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	setDefaults()

	tests := []struct {
		name      string
		key       string
		wantValue interface{}
	}{
		{"AI provider default", "ai.provider", "ollama"},
		{"AI endpoint default", "ai.endpoint", "http://localhost:11434"},
		{"AI model default", "ai.model", "gemma3:4b"},
		{"Output color default", "output.color", true},
		{"Output language default", "output.language", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := viper.Get(tt.key)
			if got != tt.wantValue {
				t.Errorf("setDefaults() %s = %v, want %v", tt.key, got, tt.wantValue)
			}
		})
	}

	t.Run("AI timeout default", func(t *testing.T) {
		got := viper.GetDuration("ai.timeout")
		if got != DefaultTimeout {
			t.Errorf("setDefaults() ai.timeout = %v, want %v", got, DefaultTimeout)
		}
	})
}

func TestInitWithDefaults(t *testing.T) {
	cfg, err := Init("")
	if err != nil {
		t.Fatalf("Init() with no config file returned error: %v", err)
	}
	if cfg.AI.Provider != "ollama" {
		t.Errorf("Init() AI.Provider = %v, want ollama", cfg.AI.Provider)
	}
	if cfg.AI.Timeout != DefaultTimeout {
		t.Errorf("Init() AI.Timeout = %v, want %v", cfg.AI.Timeout, DefaultTimeout)
	}
	if !cfg.Output.Color {
		t.Errorf("Init() Output.Color = false, want true")
	}
}

func TestInitWithConfigFile(t *testing.T) {
	tmp := t.TempDir() + "/config.yaml"
	content := []byte("ai:\n  model: mistral\n  timeout: 60s\noutput:\n  language: es\n")
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Init(tmp)
	if err != nil {
		t.Fatalf("Init() with config file returned error: %v", err)
	}
	if cfg.AI.Model != "mistral" {
		t.Errorf("Init() AI.Model = %v, want mistral", cfg.AI.Model)
	}
	if cfg.AI.Timeout != 60*time.Second {
		t.Errorf("Init() AI.Timeout = %v, want 60s", cfg.AI.Timeout)
	}
	if cfg.Output.Language != "es" {
		t.Errorf("Init() Output.Language = %v, want es", cfg.Output.Language)
	}
}
