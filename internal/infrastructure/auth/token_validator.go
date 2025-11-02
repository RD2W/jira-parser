package auth

import (
	"fmt"
	"strings"

	jira "github.com/andygrunwald/go-jira"
)

// TokenValidator интерфейс для проверки валидности токенов аутентификации
type TokenValidator interface {
	ValidateToken(baseURL, username, token string) error
}

// JiraTokenValidator реализация валидатора токенов для JIRA
type JiraTokenValidator struct{}

// NewJiraTokenValidator создает новый экземпляр валидатора токенов JIRA
func NewJiraTokenValidator() *JiraTokenValidator {
	return &JiraTokenValidator{}
}

// ValidateToken проверяет валидность токена аутентификации
func (v *JiraTokenValidator) ValidateToken(baseURL, username, token string) error {
	var client *jira.Client
	var err error

	// Определяем метод аутентификации на основе формата токена
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		// Bearer token аутентификация
		tp := jira.BearerAuthTransport{
			Token: strings.TrimPrefix(token, "bearer "),
		}
		client, err = jira.NewClient(tp.Client(), baseURL)
	} else if strings.HasPrefix(strings.ToLower(token), "basic ") {
		// Basic token аутентификация
		tp := jira.BasicAuthTransport{
			Username: username,
			Password: strings.TrimPrefix(token, "basic "),
		}
		client, err = jira.NewClient(tp.Client(), baseURL)
	} else {
		// Предполагаем, что это personal access token или пароль
		tp := jira.BasicAuthTransport{
			Username: username,
			Password: token,
		}
		client, err = jira.NewClient(tp.Client(), baseURL)
	}

	if err != nil {
		return fmt.Errorf("failed to create JIRA client: %w", err)
	}

	// Выполняем проверку валидности токена через запрос к API
	_, _, err = client.User.GetSelf()
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	return nil
}
