package config

import (
	"github.com/spf13/viper"
)

type JiraConfig struct {
	BaseURL  string `mapstructure:"base_url"`
	Username string `mapstructure:"username"`
	Token    string `mapstructure:"token"`
}

func LoadConfig(path string) (*JiraConfig, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg JiraConfig
	if err := viper.UnmarshalKey("jira", &cfg); err != nil {
		return nil, err
	}

	// Validate required fields
	if cfg.BaseURL == "" {
		return nil, &ConfigError{Field: "base_url", Message: "base_url is required"}
	}
	if cfg.Username == "" {
		return nil, &ConfigError{Field: "username", Message: "username is required"}
	}
	if cfg.Token == "" {
		return nil, &ConfigError{Field: "token", Message: "token is required"}
	}

	return &cfg, nil
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
