package domain

// QAComment представляет структурированный комментарий QA
type QAComment struct {
	SoftwareVersion string
	TestResult      string // "Fixed", "Not Fixed", "Partially Fixed", "Could not test"
	Comment         string
}

// Issue представляет JIRA тикет с комментариями
type Issue struct {
	Key      string
	Summary  string
	Comments []QAComment
}

// CommentRepository интерфейс для работы с комментариями
type CommentRepository interface {
	GetIssueComments(issueKey string) ([]QAComment, error)
	GetLastQAComment(issueKey string) (*QAComment, error)
}

// CommentService интерфейс для бизнес-логики
type CommentService interface {
	ParseComments(issueKey string) (*Issue, error)
	GetLastComment(issueKey string) (*QAComment, error)
}
