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

// this file contains the default configuration

const DefaultTimeout = 120 * time.Second
const DefaultTemperature = 0.1
const DefaultNumPredict = 2048

type Config struct {
	AI     AIConfig     `mapstructure:"ai"`
	Output OutputConfig `mapstructure:"output"`
}

type AIConfig struct {
	Provider    string        `mapstructure:"provider"`
	Endpoint    string        `mapstructure:"endpoint"`
	Model       string        `mapstructure:"model"`
	Timeout     time.Duration `mapstructure:"timeout"`
	Temperature float64       `mapstructure:"temperature"`
	NumPredict  int           `mapstructure:"num_predict"`
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
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
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
	if cfg.AI.NumPredict == 0 {
		cfg.AI.NumPredict = DefaultNumPredict
	}

	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("ai.provider", "ollama")
	viper.SetDefault("ai.endpoint", "http://localhost:11434")
	viper.SetDefault("ai.model", "llama3.1:8b")
	viper.SetDefault("ai.timeout", DefaultTimeout)
	viper.SetDefault("ai.temperature", DefaultTemperature)
	viper.SetDefault("ai.num_predict", DefaultNumPredict)
	viper.SetDefault("output.color", true)
	viper.SetDefault("output.language", "en")
}
