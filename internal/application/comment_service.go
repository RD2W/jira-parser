package application

import (
	"fmt"
	"log"

	"github.com/rd2w/jira-parser/internal/domain"
)

type CommentService struct {
	repo domain.CommentRepository
}

func NewCommentService(repo domain.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) ParseComments(issueKey string) (*domain.Issue, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key cannot be empty")
	}

	log.Printf("Starting to parse comments for issue %s", issueKey)
	comments, err := s.repo.GetIssueComments(issueKey)
	if err != nil {
		log.Printf("Error getting comments for issue %s: %v", issueKey, err)
		return nil, fmt.Errorf("failed to get comments for issue %s: %w", issueKey, err)
	}

	// Get issue info to populate the Summary
	issueInfo, err := s.repo.GetIssueInfo(issueKey)
	if err != nil {
		log.Printf("Warning: Could not get issue info for %s: %v", issueKey, err)
		// Continue with empty summary if we can't get the issue info
		issueInfo = &domain.IssueInfo{
			Key:     issueKey,
			Summary: "",
		}
	}

	log.Printf("Successfully parsed %d comments for issue %s", len(comments), issueKey)
	return &domain.Issue{
		Key:      issueInfo.Key,
		Summary:  issueInfo.Summary,
		Comments: comments,
	}, nil
}

func (s *CommentService) ParseMultipleTickets(ticketKeys []string) (*domain.IssuesList, error) {
	if len(ticketKeys) == 0 {
		return &domain.IssuesList{Issues: []domain.Issue{}}, nil
	}

	issues := make([]domain.Issue, 0, len(ticketKeys))

	for _, ticketKey := range ticketKeys {
		issue, err := s.ParseComments(ticketKey)
		if err != nil {
			log.Printf("Error parsing comments for ticket %s: %v", ticketKey, err)
			// Продолжаем обработку других тикетов даже если один не удался
			continue
		}
		issues = append(issues, *issue)
	}

	return &domain.IssuesList{Issues: issues}, nil
}

func (s *CommentService) GetLastComment(issueKey string) (*domain.QAComment, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key cannot be empty")
	}

	log.Printf("Getting last QA comment for issue %s", issueKey)
	comment, err := s.repo.GetLastQAComment(issueKey)
	if err != nil {
		log.Printf("Error getting last comment for issue %s: %v", issueKey, err)
		return nil, fmt.Errorf("failed to get last comment for issue %s: %w", issueKey, err)
	}

	if comment == nil {
		log.Printf("No QA comment found for issue %s", issueKey)
	} else {
		log.Printf("Found last QA comment for issue %s with result: %s", issueKey, comment.TestResult)
	}

	return comment, nil
}
