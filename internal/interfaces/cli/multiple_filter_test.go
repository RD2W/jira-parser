package cli

import (
	"testing"
	"time"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestFilterMultipleIssuesByResult(t *testing.T) {
	issuesList := &domain.IssuesList{
		Issues: []domain.Issue{
			{
				Key: "TEST-123",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.0",
						TestResult:      "Fixed",
						Comment:         "All tests passed",
						Created:         time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
					},
					{
						SoftwareVersion: "v1.0.1",
						TestResult:      "Not Fixed",
						Comment:         "Issue still exists",
						Created:         time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
					},
				},
			},
			{
				Key: "TEST-124",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.2",
						TestResult:      "Fixed",
						Comment:         "Fixed in this version",
						Created:         time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	// Тестируем фильтрацию по результату "Fixed"
	resultFilter := "Fixed"
	for i := range issuesList.Issues {
		var filteredComments []domain.QAComment
		for _, comment := range issuesList.Issues[i].Comments {
			if comment.TestResult == resultFilter {
				filteredComments = append(filteredComments, comment)
			}
		}
		issuesList.Issues[i].Comments = filteredComments
	}

	// Проверяем, что остались только комментарии с результатом "Fixed"
	assert.Len(t, issuesList.Issues[0].Comments, 1)
	assert.Equal(t, "Fixed", issuesList.Issues[0].Comments[0].TestResult)
	assert.Len(t, issuesList.Issues[1].Comments, 1)
	assert.Equal(t, "Fixed", issuesList.Issues[1].Comments[0].TestResult)

	// Тестируем фильтрацию по результату "Not Fixed"
	// Создаем новый issuesList для второго теста, чтобы не модифицировать исходные данные
	issuesList2 := &domain.IssuesList{
		Issues: []domain.Issue{
			{
				Key: "TEST-123",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.0",
						TestResult:      "Fixed",
						Comment:         "All tests passed",
						Created:         time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
					},
					{
						SoftwareVersion: "v1.0.1",
						TestResult:      "Not Fixed",
						Comment:         "Issue still exists",
						Created:         time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
					},
				},
			},
			{
				Key: "TEST-124",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.2",
						TestResult:      "Fixed",
						Comment:         "Fixed in this version",
						Created:         time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	resultFilter = "Not Fixed"
	for i := range issuesList2.Issues {
		var filteredComments []domain.QAComment
		for _, comment := range issuesList2.Issues[i].Comments {
			if comment.TestResult == resultFilter {
				filteredComments = append(filteredComments, comment)
			}
		}
		issuesList2.Issues[i].Comments = filteredComments
	}

	// Проверяем, что остались только комментарии с результатом "Not Fixed"
	assert.Len(t, issuesList2.Issues[0].Comments, 1)
	assert.Equal(t, "Not Fixed", issuesList2.Issues[0].Comments[0].TestResult)
	assert.Len(t, issuesList2.Issues[1].Comments, 0)
}

func TestFilterMultipleIssuesByDate(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	issuesList := &domain.IssuesList{
		Issues: []domain.Issue{
			{
				Key: "TEST-123",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.0",
						TestResult:      "Fixed",
						Comment:         "All tests passed",
						Created:         twoDaysAgo.Format(time.RFC3339),
					},
					{
						SoftwareVersion: "v1.0.1",
						TestResult:      "Not Fixed",
						Comment:         "Issue still exists",
						Created:         yesterday.Format(time.RFC3339),
					},
				},
			},
			{
				Key: "TEST-124",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.2",
						TestResult:      "Fixed",
						Comment:         "Fixed in this version",
						Created:         now.Format(time.RFC3339),
					},
				},
			},
		},
	}

	// Тестируем фильтрацию по дате "date-from"
	dateFrom := yesterday.Add(-time.Hour).Format("2006-01-02")
	for i := range issuesList.Issues {
		var filteredComments []domain.QAComment
		for _, comment := range issuesList.Issues[i].Comments {
			commentTime, err := time.Parse(time.RFC3339, comment.Created)
			assert.NoError(t, err)

			fromTime, err := time.Parse("2006-01-02", dateFrom)
			assert.NoError(t, err)

			if !commentTime.Before(fromTime) {
				filteredComments = append(filteredComments, comment)
			}
		}
		issuesList.Issues[i].Comments = filteredComments
	}

	// Проверяем, что остались только комментарии, созданные после указанной даты
	assert.Len(t, issuesList.Issues[0].Comments, 1)
	assert.Equal(t, "v1.0.1", issuesList.Issues[0].Comments[0].SoftwareVersion)
	assert.Len(t, issuesList.Issues[1].Comments, 1)
	assert.Equal(t, "v1.0.2", issuesList.Issues[1].Comments[0].SoftwareVersion)

	// Тестируем фильтрацию по дате "date-to"
	// Создаем новый issuesList для второго теста, чтобы не модифицировать исходные данные
	// Используем фиксированные даты для тестирования
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	twoDaysAgoFixed := fixedTime.Add(-48 * time.Hour)
	yesterdayFixed := fixedTime.Add(-24 * time.Hour)
	nowFixed := fixedTime

	issuesList2 := &domain.IssuesList{
		Issues: []domain.Issue{
			{
				Key: "TEST-123",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.0",
						TestResult:      "Fixed",
						Comment:         "All tests passed",
						Created:         twoDaysAgoFixed.Format(time.RFC3339),
					},
					{
						SoftwareVersion: "v1.0.1",
						TestResult:      "Not Fixed",
						Comment:         "Issue still exists",
						Created:         yesterdayFixed.Format(time.RFC3339),
					},
				},
			},
			{
				Key: "TEST-124",
				Comments: []domain.QAComment{
					{
						SoftwareVersion: "v1.0.2",
						TestResult:      "Fixed",
						Comment:         "Fixed in this version",
						Created:         nowFixed.Format(time.RFC3339),
					},
				},
			},
		},
	}

	// dateTo - это "вчера плюс один час"
	dateToTime := yesterdayFixed.Add(time.Hour)
	dateTo := dateToTime.Format(time.RFC3339)
	for i := range issuesList2.Issues {
		var filteredComments []domain.QAComment
		for _, comment := range issuesList2.Issues[i].Comments {
			commentTime, err := time.Parse(time.RFC3339, comment.Created)
			assert.NoError(t, err)

			toTime, err := time.Parse(time.RFC3339, dateTo)
			assert.NoError(t, err)

			if !commentTime.After(toTime) {
				filteredComments = append(filteredComments, comment)
			}
		}
		issuesList2.Issues[i].Comments = filteredComments
	}

	// Проверяем, что остались только комментарии, созданные до указанной даты
	assert.Len(t, issuesList2.Issues[0].Comments, 2)
	assert.Equal(t, "v1.0.0", issuesList2.Issues[0].Comments[0].SoftwareVersion)
	assert.Equal(t, "v1.0.1", issuesList2.Issues[0].Comments[1].SoftwareVersion)
	assert.Len(t, issuesList2.Issues[1].Comments, 0)
}
