package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `jira:
   base_url: "https://test.atlassian.net"
   username: "test@example.com"
   token: "test-token"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	t.Run("successful config loading", func(t *testing.T) {
		cfg, err := LoadConfig(configPath)
		assert.NoError(t, err)
		assert.Equal(t, "https://test.atlassian.net", cfg.BaseURL)
		assert.Equal(t, "test@example.com", cfg.Username)
		assert.Equal(t, "test-token", cfg.Token)
	})

	t.Run("missing base_url", func(t *testing.T) {
		invalidConfig := `jira:
     username: "test@example.com"
     token: "test-token"
  `
		invalidConfigPath := filepath.Join(tempDir, "invalid_config.yaml")
		err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
		assert.NoError(t, err)

		_, err = LoadConfig(invalidConfigPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "base_url is required")
	})

	t.Run("missing username", func(t *testing.T) {
		invalidConfig := `jira:
  base_url: "https://test.atlassian.net"
  token: "test-token"
`
		invalidConfigPath := filepath.Join(tempDir, "invalid_config2.yaml")
		err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
		assert.NoError(t, err)

		_, err = LoadConfig(invalidConfigPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username is required")
	})

	t.Run("missing token", func(t *testing.T) {
  invalidConfig := `jira:
  base_url: "https://test.atlassian.net"
  username: "test@example.com"
  `
  invalidConfigPath := filepath.Join(tempDir, "invalid_config3.yaml")
  err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
  assert.NoError(t, err)
  
  _, err = LoadConfig(invalidConfigPath)
  assert.Error(t, err)
  assert.Contains(t, err.Error(), "token is required")
	})

}
