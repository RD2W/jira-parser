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

	return &cfg, nil
}
