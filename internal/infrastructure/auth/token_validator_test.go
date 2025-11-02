package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJiraTokenValidator_ValidateToken(t *testing.T) {
	// Тестирование валидатора токенов
	validator := NewJiraTokenValidator()

	// Тест с неправильными данными для проверки возврата ошибки
	// В реальных условиях, для полного тестирования потребуется подключение к JIRA
	err := validator.ValidateToken("https://invalid-domain.atlassian.net", "invalid-user", "invalid-token")
	assert.Error(t, err)

	// Проверка Bearer токена
	err = validator.ValidateToken("https://invalid-domain.atlassian.net", "invalid-user", "Bearer invalid-token")
	assert.Error(t, err)

	// Проверка Basic токена
	err = validator.ValidateToken("https://invalid-domain.atlassian.net", "invalid-user", "Basic invalid-token")
	assert.Error(t, err)
}
