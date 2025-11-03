package jira

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/rd2w/jira-parser/internal/infrastructure/auth"
)

type JiraClient struct {
	client        *jira.Client
	parsingConfig domain.ParsingConfig
}

func NewJiraClient(baseURL, username, token string, parsingConfig domain.ParsingConfig) (*JiraClient, error) {
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

	// Validate the token before returning the client
	validator := auth.NewJiraTokenValidator()
	if err := validator.ValidateToken(baseURL, username, token); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return &JiraClient{client: client, parsingConfig: parsingConfig}, nil
}

func (jc *JiraClient) GetIssueInfo(issueKey string) (*domain.IssueInfo, error) {
	issue, _, err := jc.client.Issue.GetWithContext(context.Background(), issueKey, &jira.GetQueryOptions{
		Expand: "names,renderedFields", // Расширяем, чтобы получить больше информации о полях
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %s: %w", issueKey, err)
	}

	assigneeEmail := ""
	if issue.Fields.Assignee != nil {
		assigneeEmail = issue.Fields.Assignee.EmailAddress
	}

	// В JIRA нет специального поля "QA Owner", но мы можем определить QA владельца как
	// пользователя, оставившего последний QA комментарий
	qaOwnerEmail := jc.getQaOwnerEmail(issue)

	return &domain.IssueInfo{
		Key:           issue.Key,
		Summary:       issue.Fields.Summary,
		AssigneeEmail: assigneeEmail,
		QaOwnerEmail:  qaOwnerEmail,
	}, nil
}

func (jc *JiraClient) GetIssueComments(issueKey string) ([]domain.QAComment, error) {
	return jc.getComments(context.Background(), issueKey)
}

func (jc *JiraClient) GetLastQAComment(issueKey string) (*domain.QAComment, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key cannot be empty")
	}

	issue, _, err := jc.client.Issue.GetWithContext(context.Background(), issueKey, &jira.GetQueryOptions{
		Expand: "comments",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %s: %w", issueKey, err)
	}

	if issue.Fields.Comments == nil {
		log.Printf("No comments found for issue %s", issueKey)
		return nil, nil
	}

	// Process comments in reverse order to find the last QA comment
	for i := len(issue.Fields.Comments.Comments) - 1; i >= 0; i-- {
		comment := issue.Fields.Comments.Comments[i]
		if jc.isQAComment(comment.Body) {
			qaComment, err := jc.parseQAComment(comment.Body, comment.Created)
			if err != nil {
				log.Printf("Error parsing QA comment for issue %s: %v", issueKey, err)
				// Continue processing other comments even if one fails
				continue
			}

			// Добавляем email автора комментария
			qaComment.AuthorEmail = comment.Author.EmailAddress

			// Only return the comment if it has meaningful data
			if qaComment.SoftwareVersion != "" || qaComment.TestResult != "" || qaComment.Comment != "" {
				return &qaComment, nil
			}
		}
	}

	return nil, nil
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
	// Process comments to find QA comments
	for _, comment := range issue.Fields.Comments.Comments {
		if jc.isQAComment(comment.Body) {
			qaComment, err := jc.parseQAComment(comment.Body, comment.Created)
			if err != nil {
				log.Printf("Error parsing QA comment for issue %s: %v", issueKey, err)
				// Continue processing other comments even if one fails
				continue
			}

			// Добавляем email автора комментария
			qaComment.AuthorEmail = comment.Author.EmailAddress

			// Only add the comment if it has meaningful data
			if qaComment.SoftwareVersion != "" || qaComment.TestResult != "" || qaComment.Comment != "" {
				qaComments = append(qaComments, qaComment)
			}
		}
	}

	log.Printf("Found %d QA comments for issue %s", len(qaComments), issueKey)
	return qaComments, nil
}

func (jc *JiraClient) isQAComment(body string) bool {
	normalized := jc.removeJiraFormatting(body)
	normalized = strings.ToLower(normalized)

	// Check for various QA comment indicators using configurable patterns
	for _, indicator := range jc.parsingConfig.QAIndicators {
		// First, check for exact string matches
		if strings.Contains(normalized, strings.ToLower(indicator)) {
			return true
		}
		// Then, check if the indicator is a regex pattern
		if strings.Contains(indicator, ".*") {
			matched, err := regexp.MatchString("(?is)"+indicator, normalized)
			if err == nil && matched {
				return true
			}
		}
	}

	// Additional check for specific regex patterns that might span multiple words
	// For example, "test.*result" should match "test scenario: login\nresult: success"
	for _, indicator := range jc.parsingConfig.QAIndicators {
		if strings.Contains(indicator, ".*") {
			// Replace spaces with .* to match patterns across multiple words
			// This is a special case for patterns like "test.*result" to match "test X result"
			spaceAwarePattern := strings.ReplaceAll(indicator, " ", ".*")
			matched, err := regexp.MatchString("(?is)"+spaceAwarePattern, normalized)
			if err == nil && matched {
				return true
			}
		}
	}

	return false
}

func (jc *JiraClient) parseQAComment(body string, created string) (domain.QAComment, error) {
	var comment domain.QAComment
	normalizedBody := jc.removeJiraFormatting(body)

	// Extract version with configurable patterns
	for _, pattern := range jc.parsingConfig.VersionPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(normalizedBody); len(matches) > 1 {
			comment.SoftwareVersion = matches[1]
			break
		}
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

	// Extract result with configurable patterns
	for _, pattern := range jc.parsingConfig.ResultPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(normalizedBody); len(matches) > 1 {
			result := strings.TrimSpace(matches[1])
			// Normalize common variations using configurable mapping
			if normalized, exists := jc.parsingConfig.ResultNormalization[strings.ToLower(result)]; exists {
				result = normalized
			}
			comment.TestResult = result
			break
		}
	}

	// Extract comment with configurable patterns
	for _, pattern := range jc.parsingConfig.CommentPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(normalizedBody); len(matches) > 1 {
			comment.Comment = strings.TrimSpace(matches[1])
			break
		}
	}

	// Set created date
	comment.Created = created

	// If we still don't have a result but found "could not test" somewhere, set it
	if comment.TestResult == "" && strings.Contains(strings.ToLower(normalizedBody), "could not test") {
		comment.TestResult = "Could not test"
	}

	// If we still don't have a result, try to infer from other common keywords
	if comment.TestResult == "" {
		lowerBody := strings.ToLower(normalizedBody)
		for indicator, normalizedResult := range jc.parsingConfig.ResultNormalization {
			if strings.Contains(lowerBody, indicator) {
				comment.TestResult = normalizedResult
				break
			}
		}
	}

	// Additional fallback for common result indicators not covered by patterns
	if comment.TestResult == "" {
		lowerBody := strings.ToLower(normalizedBody)
		if strings.Contains(lowerBody, "not fixed") {
			comment.TestResult = "not fixed"
		} else if strings.Contains(lowerBody, "partially fixed") {
			comment.TestResult = "partially fixed"
		} else if strings.Contains(lowerBody, "fixed") {
			comment.TestResult = "fixed"
		} else if strings.Contains(lowerBody, "passed") {
			comment.TestResult = "passed"
		} else if strings.Contains(lowerBody, "failed") {
			comment.TestResult = "failed"
		} else if strings.Contains(lowerBody, "could not test") {
			comment.TestResult = "could not test"
		} else if strings.Contains(lowerBody, "verified") {
			comment.TestResult = "verified"
		} else if strings.Contains(lowerBody, "resolved") {
			comment.TestResult = "resolved"
		} else if strings.Contains(lowerBody, "blocked") {
			comment.TestResult = "blocked"
		} else if strings.Contains(lowerBody, "pending") {
			comment.TestResult = "pending"
		}
	}

	// Final normalization using configurable mapping
	if comment.TestResult != "" {
		if normalized, exists := jc.parsingConfig.ResultNormalization[strings.ToLower(comment.TestResult)]; exists {
			comment.TestResult = normalized
		}
	}

	return comment, nil
}

// removeJiraFormatting удаляет JIRA-разметку из текста
func (jc *JiraClient) removeJiraFormatting(text string) string {
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

// getQaOwnerEmail определяет QA владельца как пользователя, оставившего последний QA комментарий
func (jc *JiraClient) getQaOwnerEmail(issue *jira.Issue) string {
	if issue.Fields.Comments == nil {
		return ""
	}

	// Ищем последний QA комментарий и возвращаем email его автора
	for i := len(issue.Fields.Comments.Comments) - 1; i >= 0; i-- {
		comment := issue.Fields.Comments.Comments[i]
		if jc.isQAComment(comment.Body) && comment.Author.EmailAddress != "" {
			return comment.Author.EmailAddress
		}
	}

	return ""
}
