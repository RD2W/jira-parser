package application

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/stretchr/testify/assert"
)

// MockCommentRepository implements domain.CommentRepository for testing
type MockCommentRepository struct {
	GetIssueCommentsFunc func(issueKey string) ([]domain.QAComment, error)
	GetLastQACommentFunc func(issueKey string) (*domain.QAComment, error)
}

func (m *MockCommentRepository) GetIssueComments(issueKey string) ([]domain.QAComment, error) {
	if m.GetIssueCommentsFunc != nil {
		return m.GetIssueCommentsFunc(issueKey)
	}
	return nil, nil
}

func (m *MockCommentRepository) GetLastQAComment(issueKey string) (*domain.QAComment, error) {
	if m.GetLastQACommentFunc != nil {
		return m.GetLastQACommentFunc(issueKey)
	}
	return nil, nil
}

func TestCommentService_ParseComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		issueKey      string
		mockComments  []domain.QAComment
		mockError     error
		expectError   bool
		expectedCount int
	}{
		{
			name:     "successful parsing",
			issueKey: "TEST-123",
			mockComments: []domain.QAComment{
				{
					SoftwareVersion: "v1.0.0",
					TestResult:      "Fixed",
					Comment:         "Test passed successfully",
				},
				{
					SoftwareVersion: "v1.0.1",
					TestResult:      "Not Fixed",
					Comment:         "Issue still exists",
				},
			},
			mockError:     nil,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "error from repository",
			issueKey:      "TEST-456",
			mockComments:  nil,
			mockError:     errors.New("repository error"),
			expectError:   true,
			expectedCount: 0,
		},
		{
			name:          "no comments found",
			issueKey:      "TEST-789",
			mockComments:  []domain.QAComment{},
			mockError:     nil,
			expectError:   false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockCommentRepository{
				GetIssueCommentsFunc: func(issueKey string) ([]domain.QAComment, error) {
					assert.Equal(t, tt.issueKey, issueKey)
					return tt.mockComments, tt.mockError
				},
			}

			service := NewCommentService(mockRepo)
			result, err := service.ParseComments(tt.issueKey)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.issueKey, result.Key)
				assert.Equal(t, tt.expectedCount, len(result.Comments))

				if tt.expectedCount > 0 {
					assert.Equal(t, tt.mockComments, result.Comments)
				}
			}
		})
	}
}

func TestCommentService_GetLastComment(t *testing.T) {
	t.Parallel()

	lastComment := &domain.QAComment{
		SoftwareVersion: "v1.0.0",
		TestResult:      "Fixed",
		Comment:         "Latest test result",
	}

	tests := []struct {
		name        string
		issueKey    string
		mockComment *domain.QAComment
		mockError   error
		expectError bool
		expectNil   bool
	}{
		{
			name:        "successful get last comment",
			issueKey:    "TEST-123",
			mockComment: lastComment,
			mockError:   nil,
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "no comment found",
			issueKey:    "TEST-456",
			mockComment: nil,
			mockError:   nil,
			expectError: false,
			expectNil:   true,
		},
		{
			name:        "error from repository",
			issueKey:    "TEST-789",
			mockComment: nil,
			mockError:   errors.New("repository error"),
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockCommentRepository{
				GetLastQACommentFunc: func(issueKey string) (*domain.QAComment, error) {
					assert.Equal(t, tt.issueKey, issueKey)
					return tt.mockComment, tt.mockError
				},
			}

			service := NewCommentService(mockRepo)
			result, err := service.GetLastComment(tt.issueKey)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)

				if tt.expectNil {
					assert.Nil(t, result)
				} else {
					assert.Equal(t, tt.mockComment, result)
				}
			}
		})
	}
}

func TestCommentService_ParseMultipleTickets(t *testing.T) {
	t.Parallel()

	firstIssueComments := []domain.QAComment{
		{
			SoftwareVersion: "v1.0.0",
			TestResult:      "Fixed",
			Comment:         "Test passed successfully",
		},
	}

	secondIssueComments := []domain.QAComment{
		{
			SoftwareVersion: "v1.0.1",
			TestResult:      "Not Fixed",
			Comment:         "Issue still exists",
		},
	}

	tests := []struct {
		name             string
		ticketKeys       []string
		mockCommentsFunc func(issueKey string) ([]domain.QAComment, error)
		expectError      bool
		expectedIssues   int
	}{
		{
			name:       "successful parsing multiple tickets",
			ticketKeys: []string{"TEST-123", "TEST-456"},
			mockCommentsFunc: func(issueKey string) ([]domain.QAComment, error) {
				switch issueKey {
				case "TEST-123":
					return firstIssueComments, nil
				case "TEST-456":
					return secondIssueComments, nil
				default:
					return nil, fmt.Errorf("unexpected issue key: %s", issueKey)
				}
			},
			expectError:    false,
			expectedIssues: 2,
		},
		{
			name:             "empty ticket list",
			ticketKeys:       []string{},
			mockCommentsFunc: nil,
			expectError:      false,
			expectedIssues:   0,
		},
		{
			name:       "error on one ticket continues processing",
			ticketKeys: []string{"TEST-123", "TEST-789", "TEST-456"},
			mockCommentsFunc: func(issueKey string) ([]domain.QAComment, error) {
				switch issueKey {
				case "TEST-123":
					return firstIssueComments, nil
				case "TEST-789":
					return nil, fmt.Errorf("error parsing issue")
				case "TEST-456":
					return secondIssueComments, nil
				default:
					return nil, fmt.Errorf("unexpected issue key: %s", issueKey)
				}
			},
			expectError:    false,
			expectedIssues: 2, // Should return 2 successful issues
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockCommentRepository{
				GetIssueCommentsFunc: tt.mockCommentsFunc,
				GetLastQACommentFunc: func(issueKey string) (*domain.QAComment, error) {
					return nil, nil
				},
			}

			service := NewCommentService(mockRepo)
			result, err := service.ParseMultipleTickets(tt.ticketKeys)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedIssues, len(result.Issues))
			}
		})
	}
}
