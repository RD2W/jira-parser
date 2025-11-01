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

	log.Printf("Successfully parsed %d comments for issue %s", len(comments), issueKey)
	return &domain.Issue{
		Key:      issueKey,
		Summary:  "", // Summary will be populated if needed from JIRA
		Comments: comments,
	}, nil
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
