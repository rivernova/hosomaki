// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// configuration loading and parsing logic

const DefaultTimeout = 120 * time.Second

type Config struct {
	AI     AIConfig     `mapstructure:"ai"`
	Output OutputConfig `mapstructure:"output"`
}

type AIConfig struct {
	Provider string        `mapstructure:"provider"`
	Endpoint string        `mapstructure:"endpoint"`
	Model    string        `mapstructure:"model"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type OutputConfig struct {
	Color    bool   `mapstructure:"color"`
	Language string `mapstructure:"language"`
}

func Init(cfgFile string) (Config, error) {
	viper.Reset()

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
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return Config{}, fmt.Errorf("config: read: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("config: parse: %w", err)
	}

	if cfg.AI.Timeout == 0 {
		cfg.AI.Timeout = DefaultTimeout
	}

	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("ai.provider", "ollama")
	viper.SetDefault("ai.endpoint", "http://localhost:11434")
	viper.SetDefault("ai.model", "llama3.1:8b")
	viper.SetDefault("ai.timeout", DefaultTimeout)
	viper.SetDefault("output.color", true)
	viper.SetDefault("output.language", "en")
}
