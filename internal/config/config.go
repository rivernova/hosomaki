// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// C is the global config, available after Init() is called.
var C Config

type Config struct {
	AI     AIConfig     `mapstructure:"ai"`
	Output OutputConfig `mapstructure:"output"`
}

type AIConfig struct {
	Provider string `mapstructure:"provider"`
	Endpoint string `mapstructure:"endpoint"`
	Model    string `mapstructure:"model"`
}

type OutputConfig struct {
	Color    bool   `mapstructure:"color"`
	Language string `mapstructure:"language"`
}

// Init loads config from file and environment. Called by cobra.OnInitialize.
func Init() {
	viper.SetDefault("ai.provider", "ollama")
	viper.SetDefault("ai.endpoint", "http://localhost:11434")
	viper.SetDefault("ai.model", "llama3")
	viper.SetDefault("output.color", true)
	viper.SetDefault("output.language", "en")

	// Config file locations, in order of preference
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(filepath.Join(home, ".config", "hosomaki"))
			viper.AddConfigPath(home)
		}
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Allow env overrides: HOSOMAKI_AI_MODEL, HOSOMAKI_AI_ENDPOINT, etc.
	viper.SetEnvPrefix("HOSOMAKI")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "config error:", err)
		}
		// No config file found — defaults apply, that's fine.
	}

	if err := viper.Unmarshal(&C); err != nil {
		fmt.Fprintln(os.Stderr, "config parse error:", err)
	}
}
