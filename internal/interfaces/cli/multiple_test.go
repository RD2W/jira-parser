package cli

import (
	"io"
	"os"
	"testing"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewParseMultipleCommand(t *testing.T) {
	// Создаем временный конфигурационный файл
	tempDir := t.TempDir()
	configPath := tempDir + "/config.yaml"

	configContent := `jira:
  base_url: "https://test.atlassian.net"
  username: "test@example.com"
  token: "test-token"
  tickets:
    - "TOS-30690"
    - "TOS-30692"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Устанавливаем путь к конфигурационному файлу
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	// Создаем команду
	cmd := NewParseMultipleCommand()
	assert.NotNil(t, cmd)
	assert.Equal(t, "parse-multiple", cmd.Use)
	assert.Equal(t, "Parse QA comments for multiple tickets from tickets file", cmd.Short)
}

func TestPrintMultipleIssues(t *testing.T) {
	issuesList := &domain.IssuesList{
		Issues: []domain.Issue{
			{
				Key: "TOS-30690",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.0",
						TestResult:      "Fixed",
						Comment:         "All tests passed",
					},
				},
			},
			{
				Key: "TOS-30692",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.1",
						TestResult:      "Not Fixed",
						Comment:         "Issue still exists",
					},
					{
						SoftwareVersion: "v1.0.2",
						TestResult:      "Fixed",
					},
				},
			},
		},
	}

	// Захватываем вывод
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printMultipleIssues(issuesList)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	output := string(out)

	// Проверяем, что вывод содержит ожидаемые элементы
	assert.Contains(t, output, "Found 2 issues with QA comments:")
	assert.Contains(t, output, "Issue: TOS-30690")
	assert.Contains(t, output, "Issue: TOS-30692")
	assert.Contains(t, output, "Found 1 QA comments:")
	assert.Contains(t, output, "Found 2 QA comments:")
	assert.Contains(t, output, "Version: v1.0.0")
	assert.Contains(t, output, "Result: Fixed")
	assert.Contains(t, output, "Comment: All tests passed")
	assert.Contains(t, output, "Version: v1.0.1")
	assert.Contains(t, output, "Result: Not Fixed")
	assert.Contains(t, output, "Comment: Issue still exists")
	assert.Contains(t, output, "Version: v1.0.2")
}

func TestParseMultipleCommand_Execute(t *testing.T) {
	// Создаем команду
	cmd := &cobra.Command{}

	// Проверяем, что команда может быть создана без ошибок
	cmd = NewParseMultipleCommand()
	assert.NotNil(t, cmd)

	// Проверяем, что команда имеет правильные параметры
	assert.Equal(t, "parse-multiple", cmd.Use)
}
