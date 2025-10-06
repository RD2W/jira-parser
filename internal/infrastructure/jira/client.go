package jira

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/rd2w/jira-parser/internal/domain"
)

type JiraClient struct {
	client *jira.Client
}

func NewJiraClient(baseURL, username, token string) (*JiraClient, error) {
	tp := jira.BearerAuthTransport{
		Token: token,
	}

	client, err := jira.NewClient(tp.Client(), baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create JIRA client: %w", err)
	}

	return &JiraClient{client: client}, nil
}

func (jc *JiraClient) GetIssueComments(issueKey string) ([]domain.QAComment, error) {
	return jc.getComments(context.Background(), issueKey)
}

func (jc *JiraClient) GetLastQAComment(issueKey string) (*domain.QAComment, error) {
	comments, err := jc.getComments(context.Background(), issueKey)
	if err != nil {
		return nil, err
	}

	if len(comments) == 0 {
		return nil, nil
	}

	return &comments[len(comments)-1], nil
}

func (jc *JiraClient) getComments(ctx context.Context, issueKey string) ([]domain.QAComment, error) {
	issue, _, err := jc.client.Issue.GetWithContext(ctx, issueKey, &jira.GetQueryOptions{
		Expand: "comments",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	var qaComments []domain.QAComment
	for _, comment := range issue.Fields.Comments.Comments {
		if isQAComment(comment.Body) {
			qaComment, err := parseQAComment(comment.Body)
			if err != nil {
				continue // или логировать ошибку
			}
			qaComments = append(qaComments, qaComment)
		}
	}

	return qaComments, nil
}

func isQAComment(body string) bool {
	normalized := removeJiraFormatting(body)
	return strings.Contains(normalized, "Tested on") ||
		strings.Contains(normalized, "Could not test on SW")
}

func parseQAComment(body string) (domain.QAComment, error) {
	var comment domain.QAComment
	normalizedBody := removeJiraFormatting(body)

	versionRe := regexp.MustCompile(`Tested on (?:SW )?(v[\d.]+)`)
	resultRe := regexp.MustCompile(`Result:\s*([\w\s]+)`)
	commentRe := regexp.MustCompile(`Comment:\s*(.+)`)

	if matches := versionRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.SoftwareVersion = matches[1]
	} else if strings.Contains(normalizedBody, "Could not test on SW") {
		comment.TestResult = "Could not test"
	}

	if matches := resultRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.TestResult = strings.TrimSpace(matches[1])
	}

	if matches := commentRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.Comment = strings.TrimSpace(matches[1])
	}

	return comment, nil
}

// removeJiraFormatting удаляет JIRA-разметку из текста
func removeJiraFormatting(text string) string {
	// Удаляем базовое форматирование JIRA
	replacements := map[string]string{
		"*":   "", // жирный
		"_":   "", // курсив
		"-":   "", // зачеркивание
		"??":  "", // моноширинный
		"{{":  "",
		"}}":  "",
		"{*}": "",
	}

	// Удаляем цветовую разметку {color}
	colorRe := regexp.MustCompile(`\{color[^\}]*\}(.*?)\{color\}`)
	text = colorRe.ReplaceAllString(text, "$1")

	// Удаляем другие элементы форматирования
	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	// Удаляем ссылки [текст|url]
	linkRe := regexp.MustCompile(`\[([^\|\]]+)(?:\|[^\]]+)?\]`)
	text = linkRe.ReplaceAllString(text, "$1")

	return strings.TrimSpace(text)
}
