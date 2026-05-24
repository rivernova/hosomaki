// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

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
		{"AI model default", "ai.model", "llama3"},
		{"AI timeout default", "ai.timeout", DefaultTimeout},
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
}

func TestDefaultTimeout(t *testing.T) {
	expected := 120 * time.Second
	if DefaultTimeout != expected {
		t.Errorf("DefaultTimeout = %v, want %v", DefaultTimeout, expected)
	}
}

func TestInitWithDefaults(t *testing.T) {
	viper.Reset()
	setDefaults()

	cfg, err := Init("")
	if err != nil {
		t.Errorf("Init() with no config file returned error: %v", err)
		return
	}

	if cfg.AI.Provider != "ollama" {
		t.Errorf("Init() AI.Provider = %v, want %v", cfg.AI.Provider, "ollama")
	}
	if cfg.AI.Timeout != DefaultTimeout {
		t.Errorf("Init() AI.Timeout = %v, want %v", cfg.AI.Timeout, DefaultTimeout)
	}
	if cfg.Output.Color != true {
		t.Errorf("Init() Output.Color = %v, want %v", cfg.Output.Color, true)
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		AI: AIConfig{
			Provider: "test-provider",
			Endpoint: "http://test:11434",
			Model:    "test-model",
			Timeout:  60 * time.Second,
		},
		Output: OutputConfig{
			Color:    false,
			Language: "cs",
		},
	}

	if cfg.AI.Provider != "test-provider" {
		t.Errorf("Config struct AI.Provider = %v, want %v", cfg.AI.Provider, "test-provider")
	}
	if cfg.Output.Language != "cs" {
		t.Errorf("Config struct Output.Language = %v, want %v", cfg.Output.Language, "cs")
	}
}
