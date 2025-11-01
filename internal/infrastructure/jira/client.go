package jira

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/rd2w/jira-parser/internal/domain"
)

type JiraClient struct {
	client *jira.Client
}

func NewJiraClient(baseURL, username, token string) (*JiraClient, error) {
	var client *jira.Client
	var err error

	// Determine authentication method based on token format
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		// Bearer token authentication
		tp := jira.BearerAuthTransport{
			Token: strings.TrimPrefix(token, "bearer "),
		}
		client, err = jira.NewClient(tp.Client(), baseURL)
	} else if strings.HasPrefix(strings.ToLower(token), "basic ") {
		// Basic token authentication
		tp := jira.BasicAuthTransport{
			Username: username,
			Password: strings.TrimPrefix(token, "basic "),
		}
		client, err = jira.NewClient(tp.Client(), baseURL)
	} else {
		// Assume personal access token or password
		tp := jira.BasicAuthTransport{
			Username: username,
			Password: token,
		}
		client, err = jira.NewClient(tp.Client(), baseURL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create JIRA client: %w", err)
	}

	return &JiraClient{client: client}, nil
}

func (jc *JiraClient) GetIssueComments(issueKey string) ([]domain.QAComment, error) {
	return jc.getComments(context.Background(), issueKey)
}

func (jc *JiraClient) GetLastQAComment(issueKey string) (*domain.QAComment, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key cannot be empty")
	}

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
	if issueKey == "" {
		return nil, fmt.Errorf("issue key cannot be empty")
	}

	issue, _, err := jc.client.Issue.GetWithContext(ctx, issueKey, &jira.GetQueryOptions{
		Expand: "comments",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %s: %w", issueKey, err)
	}

	if issue.Fields.Comments == nil {
		log.Printf("No comments found for issue %s", issueKey)
		return []domain.QAComment{}, nil
	}

	var qaComments []domain.QAComment
	for _, comment := range issue.Fields.Comments.Comments {
		if isQAComment(comment.Body) {
			qaComment, err := parseQAComment(comment.Body)
			if err != nil {
				log.Printf("Error parsing QA comment for issue %s: %v", issueKey, err)
				// Continue processing other comments even if one fails
				continue
			}

			// Only add the comment if it has meaningful data
			if qaComment.SoftwareVersion != "" || qaComment.TestResult != "" || qaComment.Comment != "" {
				qaComments = append(qaComments, qaComment)
			}
		}
	}

	log.Printf("Found %d QA comments for issue %s", len(qaComments), issueKey)
	return qaComments, nil
}

func isQAComment(body string) bool {
	normalized := removeJiraFormatting(body)
	normalized = strings.ToLower(normalized)

	// Check for various QA comment indicators
	return strings.Contains(normalized, "tested on") ||
		strings.Contains(normalized, "could not test on sw") ||
		strings.Contains(normalized, "qa comment") ||
		strings.Contains(normalized, "qa verification") ||
		strings.Contains(normalized, "qa tested") ||
		(strings.Contains(normalized, "test") &&
			(strings.Contains(normalized, "result") ||
				strings.Contains(normalized, "passed") ||
				strings.Contains(normalized, "failed") ||
				strings.Contains(normalized, "status")))
}

func parseQAComment(body string) (domain.QAComment, error) {
	var comment domain.QAComment
	normalizedBody := removeJiraFormatting(body)

	// Version patterns - more comprehensive
	versionRe := regexp.MustCompile(`(?i)Tested on (?:SW )?(v?[\d.]+(?:-[\w.]+)?)`)
	// More flexible version pattern
	altVersionRe := regexp.MustCompile(`(?i)version.*?(v?[\d.]+(?:-[\w.]+)?)`)
	// Alternative version pattern for different formats
	swVersionRe := regexp.MustCompile(`(?i)sw.*?(v?[\d.]+(?:-[\w.]+)?)`)

	// Result patterns - expanded
	resultRe := regexp.MustCompile(`(?i)Result:\s*([^\n\r]+)`)
	statusRe := regexp.MustCompile(`(?i)Status:\s*([^\n\r]+)`)
	// Alternative result patterns with more options
	altResultRe := regexp.MustCompile(`(?i)(Fixed|Not Fixed|Partially Fixed|Could not test|Passed|Failed|Blocked|Resolved|Verified|Re-Test|Pending|In Progress|N/A)`)

	// Comment patterns
	commentRe := regexp.MustCompile(`(?i)Comment:\s*(.+)`)
	noteRe := regexp.MustCompile(`(?i)Notes?:\s*(.+)`)
	// Additional comment patterns
	observationRe := regexp.MustCompile(`(?i)Observations?:\s*(.+)`)

	// Extract version with multiple patterns
	if matches := versionRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.SoftwareVersion = matches[1]
	} else if matches := altVersionRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.SoftwareVersion = matches[1]
	} else if matches := swVersionRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.SoftwareVersion = matches[1]
	}

	// Handle "could not test" case - extract version if possible
	if strings.Contains(strings.ToLower(normalizedBody), "could not test on sw") {
		// Extract version from "could not test on SW vX.X.X" pattern
		couldNotTestVersionRe := regexp.MustCompile(`(?i)could not test on sw (v?[\d.]+(?:-[\w.]+)?)`)
		if matches := couldNotTestVersionRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
			comment.SoftwareVersion = matches[1]
		}
		comment.TestResult = "Could not test"
	}

	// Extract result with multiple patterns
	if matches := resultRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.TestResult = strings.TrimSpace(matches[1])
	} else if matches := statusRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.TestResult = strings.TrimSpace(matches[1])
	} else if matches := altResultRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		result := strings.TrimSpace(matches[1])
		// Normalize common variations
		switch strings.ToLower(result) {
		case "passed", "verified", "resolved", "re-test":
			result = "Fixed"
		case "failed", "blocked", "pending", "in progress":
			result = "Not Fixed"
		case "n/a", "not applicable":
			result = "N/A"
		}
		comment.TestResult = result
	}

	// Extract comment with multiple patterns
	if matches := commentRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.Comment = strings.TrimSpace(matches[1])
	} else if matches := noteRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.Comment = strings.TrimSpace(matches[1])
	} else if matches := observationRe.FindStringSubmatch(normalizedBody); len(matches) > 1 {
		comment.Comment = strings.TrimSpace(matches[1])
	}

	// If we still don't have a result but found "could not test" somewhere, set it
	if comment.TestResult == "" && strings.Contains(strings.ToLower(normalizedBody), "could not test") {
		comment.TestResult = "Could not test"
	}

	// If we still don't have a result, try to infer from other common keywords
	if comment.TestResult == "" {
		lowerBody := strings.ToLower(normalizedBody)
		if strings.Contains(lowerBody, "not fixed") {
			comment.TestResult = "Not Fixed"
		} else if strings.Contains(lowerBody, "partially fixed") {
			comment.TestResult = "Partially Fixed"
		} else if strings.Contains(lowerBody, "fixed") {
			comment.TestResult = "Fixed"
		} else if strings.Contains(lowerBody, "passed") {
			comment.TestResult = "Fixed"
		} else if strings.Contains(lowerBody, "failed") {
			comment.TestResult = "Not Fixed"
		} else if strings.Contains(lowerBody, "could not test") {
			comment.TestResult = "Could not test"
		} else if strings.Contains(lowerBody, "verified") {
			comment.TestResult = "Fixed"
		} else if strings.Contains(lowerBody, "resolved") {
			comment.TestResult = "Fixed"
		} else if strings.Contains(lowerBody, "blocked") {
			comment.TestResult = "Not Fixed"
		} else if strings.Contains(lowerBody, "pending") {
			comment.TestResult = "Not Fixed"
		}
	}

	// Normalize result if needed
	if comment.TestResult == "Passed" {
		comment.TestResult = "Fixed"
	} else if comment.TestResult == "Failed" {
		comment.TestResult = "Not Fixed"
	} else if strings.ToLower(comment.TestResult) == "not fixed" {
		// Ensure proper capitalization
		comment.TestResult = "Not Fixed"
	} else if comment.TestResult == "Verified" {
		comment.TestResult = "Fixed"
	} else if comment.TestResult == "Resolved" {
		comment.TestResult = "Fixed"
	}

	return comment, nil
}

// removeJiraFormatting удаляет JIRA-разметку из текста
func removeJiraFormatting(text string) string {
	// Удаляем базовое форматирование JIRA
	replacements := map[string]string{
		"*":               "", // жирный
		"_":               "", // курсив
		"-":               "", // зачеркивание
		"??":              "", // моноширинный
		"{{":              "",
		"}}":              "",
		"{*}":             "",
		"{code}":          "",
		"{code:":          "",
		"{noformat}":      "",
		"{quote}":         "",
		"{panel}":         "",
		"{panel:bgcolor=": "",
		"{color:":         "",
		"{color}":         "",
	}

	// Удаляем цветовую разметку {color}
	colorRe := regexp.MustCompile(`\{color[^\}]*\}(.*?)\{color\}`)
	text = colorRe.ReplaceAllString(text, "$1")

	// Удаляем код-блоки {code}...{code}
	codeRe := regexp.MustCompile(`\{code[^\}]*\}(.*?)\{code\}`)
	text = codeRe.ReplaceAllString(text, "$1")

	// Удаляем панели {panel}...{panel}
	panelRe := regexp.MustCompile(`\{panel[^\}]*\}(.*?)\{panel\}`)
	text = panelRe.ReplaceAllString(text, "$1")

	// Удаляем другие элементы форматирования
	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	// Удаляем ссылки [текст|url]
	linkRe := regexp.MustCompile(`\[([^\|\]]+)(?:\|[^\]]+)?\]`)
	text = linkRe.ReplaceAllString(text, "$1")

	// Удаляем встроенные изображения !image.png!
	imageRe := regexp.MustCompile(`!\S*!`)
	text = imageRe.ReplaceAllString(text, "")

	// Удаляем упоминания пользователей [~username]
	mentionRe := regexp.MustCompile(`\[~[^\]]+\]`)
	text = mentionRe.ReplaceAllString(text, "")

	// Удаляем метки {noformat}...{noformat}
	noformatRe := regexp.MustCompile(`\{noformat\}(.*?)\{noformat\}`)
	text = noformatRe.ReplaceAllString(text, "$1")

	// Удаляем цитаты {quote}...{quote}
	quoteRe := regexp.MustCompile(`\{quote\}(.*?)\{quote\}`)
	text = quoteRe.ReplaceAllString(text, "$1")

	return strings.TrimSpace(text)
}
