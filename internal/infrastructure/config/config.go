package config

import (
	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/spf13/viper"
)

type JiraConfig struct {
	BaseURL  string               `mapstructure:"base_url"`
	Username string               `mapstructure:"username"`
	Token    string               `mapstructure:"token"`
	Parsing  domain.ParsingConfig `mapstructure:"parsing"`
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

	// Load parsing configuration
	var parsingCfg domain.ParsingConfig
	if err := viper.UnmarshalKey("parsing", &parsingCfg); err != nil {
		// If parsing config is not found, use default values
		parsingCfg = domain.ParsingConfig{
			VersionPatterns: []string{
				`(?i)Tested on (?:SW )?(v?[\d.]+(?:-[\w.]+)?)`,
				`(?i)version.*?(v?[\d.]+(?:-[\w.]+)?)`,
				`(?i)sw.*?(v?[\d.]+(?:-[\w.]+)?)`,
			},
			ResultPatterns: []string{
				`(?i)Result:\s*([^\n\r]+)`,
				`(?i)Status:\s*([^\n\r]+)`,
				`(?i)(Fixed|Not Fixed|Partially Fixed|Could not test|Passed|Failed|Blocked|Resolved|Verified|Re-Test|Pending|In Progress|N/A)`,
			},
			CommentPatterns: []string{
				`(?i)Comment:\s*(.+)`,
				`(?i)Notes?:\s*(.+)`,
				`(?i)Observations?:\s*(.+)`,
			},
			QAIndicators: []string{
				"tested on",
				"could not test on sw",
				"qa comment",
				"qa verification",
				"qa tested",
				"test.*result",
				"test.*passed",
				"test.*failed",
				"test.*status",
			},
			ResultNormalization: map[string]string{
				"passed":         "Fixed",
				"verified":       "Fixed",
				"resolved":       "Fixed",
				"re-test":        "Fixed",
				"failed":         "Not Fixed",
				"blocked":        "Not Fixed",
				"pending":        "Not Fixed",
				"in progress":    "Not Fixed",
				"n/a":            "N/A",
				"not applicable": "N/A",
			},
		}
	}

	cfg.Parsing = parsingCfg

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
