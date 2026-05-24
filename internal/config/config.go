// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const DefaultTimeout = 120 * time.Second

// Config is the top-level configuration structure for hosomaki.
type Config struct {
	AI     AIConfig     `mapstructure:"ai"`
	Output OutputConfig `mapstructure:"output"`
}

// AIConfig holds provider-level settings.
type AIConfig struct {
	Provider string        `mapstructure:"provider"`
	Endpoint string        `mapstructure:"endpoint"`
	Model    string        `mapstructure:"model"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// OutputConfig holds presentation settings.
type OutputConfig struct {
	Color    bool   `mapstructure:"color"`
	Language string `mapstructure:"language"`
}

// Init loads configuration from file and environment variables.
// It returns the populated Config and any error encountered.
// Callers decide whether a config error is fatal — Init does not print to
// stderr or call os.Exit, keeping it testable and composable.
func Init(cfgFile string) (Config, error) {
	setDefaults()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(filepath.Join(home, ".config", "hosomaki"))
			viper.AddConfigPath(home)
		}
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("HOSOMAKI")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// A file was found but could not be read — that is a real error.
			return Config{}, fmt.Errorf("config: read: %w", err)
		}
		// No config file found — defaults and env vars apply, which is fine.
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("config: parse: %w", err)
	}

	// Apply the sentinel: if timeout was not set, use the default.
	if cfg.AI.Timeout == 0 {
		cfg.AI.Timeout = DefaultTimeout
	}

	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("ai.provider", "ollama")
	viper.SetDefault("ai.endpoint", "http://localhost:11434")
	viper.SetDefault("ai.model", "llama3")
	viper.SetDefault("ai.timeout", DefaultTimeout)
	viper.SetDefault("output.color", true)
	viper.SetDefault("output.language", "en")
}
