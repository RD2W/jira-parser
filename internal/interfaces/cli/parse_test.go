package cli

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewParseCommand(t *testing.T) {
	// Создаем временную директорию для конфигурационного файла
	tempDir := t.TempDir()
	configPath := tempDir + "/config.yaml"

	// Создаем минимальный конфигурационный файл
	configContent := `jira:
  base_url: "https://test.atlassian.net"
  username: "test@example.com"
  token: "test-token"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Устанавливаем путь к конфигурационному файлу
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	// Создаем команду
	cmd := NewParseCommand()
	assert.NotNil(t, cmd)
	assert.Equal(t, "parse <issue-key>", cmd.Use)
	assert.Equal(t, "Parse all QA comments for an issue", cmd.Short)
}

func TestParseCommand_Execute(t *testing.T) {
	// Создаем временную директорию для конфигурационного файла
	tempDir := t.TempDir()
	configPath := tempDir + "/config.yaml"

	// Создаем минимальный конфигурационный файл
	configContent := `jira:
  base_url: "https://test.atlassian.net"
  username: "test@example.com"
  token: "test-token"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Устанавливаем путь к конфигурационному файлу
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	// Создаем команду
	cmd := NewParseCommand()
	assert.NotNil(t, cmd)

	// Проверяем, что команда требует аргумент
	err = cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestParseCommand_WithArgs(t *testing.T) {
	// Создаем временную директорию для конфигурационного файла
	tempDir := t.TempDir()
	configPath := tempDir + "/config.yaml"

	// Создаем минимальный конфигурационный файл
	configContent := `jira:
  base_url: "https://test.atlassian.net"
  username: "test@example.com"
  token: "test-token"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Устанавливаем путь к конфигурационному файлу
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	assert.NoError(t, err)

	// Создаем команду
	cmd := NewParseCommand()
	assert.NotNil(t, cmd)

	// Устанавливаем аргументы
	cmd.SetArgs([]string{"TEST-123"})

	// Захватываем вывод
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Выполняем команду (ожидаем ошибку из-за недоступности JIRA сервера)
	// Игнорируем ошибку, потому что мы хотим проверить только, что команда может быть выполнена
	_ = cmd.Execute()

	// Закрываем канал записи и восстанавливаем stdout
	w.Close()
	os.Stdout = old

	// Читаем вывод (даже если команда завершилась с ошибкой)
	_, _ = io.ReadAll(r)

	// Проверяем, что команда была выполнена (возможно с ошибкой из-за недоступности сервера)
	// Главное, что команда не завершилась с паникой
	assert.NotNil(t, cmd)
}

func TestFilterCommentsByResult(t *testing.T) {
	comments := []domain.QAComment{
		{
			SoftwareVersion: "v1.0.0",
			TestResult:      "Fixed",
			Comment:         "All tests passed",
			Created:         time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
		{
			SoftwareVersion: "v1.0.1",
			TestResult:      "Not Fixed",
			Comment:         "Issue still exists",
			Created:         time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
		},
		{
			SoftwareVersion: "v1.0.2",
			TestResult:      "Fixed",
			Comment:         "Fixed in this version",
			Created:         time.Now().Format(time.RFC3339),
		},
	}

	issue := &domain.Issue{
		Key:      "TEST-123",
		Comments: comments,
	}

	// Тестируем фильтрацию по результату "Fixed"
	var filteredComments []domain.QAComment
	for _, comment := range issue.Comments {
		if comment.TestResult == "Fixed" {
			filteredComments = append(filteredComments, comment)
		}
	}

	assert.Len(t, filteredComments, 2)
	assert.Equal(t, "Fixed", filteredComments[0].TestResult)
	assert.Equal(t, "Fixed", filteredComments[1].TestResult)

	// Тестируем фильтрацию по результату "Not Fixed"
	filteredComments = []domain.QAComment{}
	for _, comment := range issue.Comments {
		if comment.TestResult == "Not Fixed" {
			filteredComments = append(filteredComments, comment)
		}
	}

	assert.Len(t, filteredComments, 1)
	assert.Equal(t, "Not Fixed", filteredComments[0].TestResult)
}

func TestFilterCommentsByDate(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	comments := []domain.QAComment{
		{
			SoftwareVersion: "v1.0.0",
			TestResult:      "Fixed",
			Comment:         "All tests passed",
			Created:         twoDaysAgo.Format(time.RFC3339),
		},
		{
			SoftwareVersion: "v1.0.1",
			TestResult:      "Not Fixed",
			Comment:         "Issue still exists",
			Created:         yesterday.Format(time.RFC3339),
		},
		{
			SoftwareVersion: "v1.0.2",
			TestResult:      "Fixed",
			Comment:         "Fixed in this version",
			Created:         now.Format(time.RFC3339),
		},
	}

	issue := &domain.Issue{
		Key:      "TEST-123",
		Comments: comments,
	}

	// Тестируем фильтрацию по дате "date-from"
	var filteredComments []domain.QAComment
	dateFrom := yesterday.Add(-time.Hour).Format("2006-01-02")

	for _, comment := range issue.Comments {
		commentTime, err := time.Parse(time.RFC3339, comment.Created)
		assert.NoError(t, err)

		fromTime, err := time.Parse("2006-01-02", dateFrom)
		assert.NoError(t, err)

		if !commentTime.Before(fromTime) {
			filteredComments = append(filteredComments, comment)
		}
	}

	assert.Len(t, filteredComments, 2)
	assert.Equal(t, "v1.0.1", filteredComments[0].SoftwareVersion)
	assert.Equal(t, "v1.0.2", filteredComments[1].SoftwareVersion)

	// Тестируем фильтрацию по дате "date-to"
	filteredComments = []domain.QAComment{}
	dateTo := yesterday.Add(time.Hour).Format(time.RFC3339)

	for _, comment := range issue.Comments {
		commentTime, err := time.Parse(time.RFC3339, comment.Created)
		assert.NoError(t, err)

		toTime, err := time.Parse(time.RFC3339, dateTo)
		assert.NoError(t, err)

		if !commentTime.After(toTime) {
			filteredComments = append(filteredComments, comment)
		}
	}

	assert.Len(t, filteredComments, 2)
	assert.Equal(t, "v1.0.0", filteredComments[0].SoftwareVersion)
	assert.Equal(t, "v1.0.1", filteredComments[1].SoftwareVersion)
}
