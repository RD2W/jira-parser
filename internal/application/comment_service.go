package application

import (
	"github.com/rd2w/jira-parser/internal/domain"
)

type CommentService struct {
	repo domain.CommentRepository
}

func NewCommentService(repo domain.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) ParseComments(issueKey string) (*domain.Issue, error) {
	comments, err := s.repo.GetIssueComments(issueKey)
	if err != nil {
		return nil, err
	}

	return &domain.Issue{
		Key:      issueKey,
		Comments: comments,
	}, nil
}

func (s *CommentService) GetLastComment(issueKey string) (*domain.QAComment, error) {
	return s.repo.GetLastQAComment(issueKey)
}
