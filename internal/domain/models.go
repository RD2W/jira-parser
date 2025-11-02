package domain

// QAComment представляет структурированный комментарий QA
type QAComment struct {
	SoftwareVersion string
	TestResult      string // "Fixed", "Not Fixed", "Partially Fixed", "Could not test"
	Comment         string
	Created         string // Дата создания комментария в формате RFC3339
}

// Issue представляет JIRA тикет с комментариями
type Issue struct {
	Key      string
	Summary  string
	Comments []QAComment
}

// IssuesList представляет список JIRA тикетов с комментариями
type IssuesList struct {
	Issues []Issue
}

// ParsingConfig содержит настройки для парсинга комментариев
type ParsingConfig struct {
	VersionPatterns     []string          `mapstructure:"version_patterns"`
	ResultPatterns      []string          `mapstructure:"result_patterns"`
	CommentPatterns     []string          `mapstructure:"comment_patterns"`
	QAIndicators        []string          `mapstructure:"qa_indicators"`
	ResultNormalization map[string]string `mapstructure:"result_normalization"`
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
	ParseMultipleTickets(ticketKeys []string) (*IssuesList, error)
}
