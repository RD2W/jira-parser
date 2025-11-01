package jira

import (
	"testing"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestIsQAComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		expected bool
	}{
		{
			name:     "contains Tested on",
			body:     "Tested on v1.2.3\nResult: Fixed",
			expected: true,
		},
		{
			name:     "contains Could not test on SW",
			body:     "Could not test on SW v1.2.3 due to environment issues",
			expected: true,
		},
		{
			name:     "contains QA Comment",
			body:     "QA Comment: Tested functionality\nResult: Passed",
			expected: true,
		},
		{
			name:     "contains Test and Result",
			body:     "Test scenario: Login\nResult: Success",
			expected: true,
		},
		{
			name:     "regular comment",
			body:     "This is a regular comment without QA keywords",
			expected: false,
		},
		{
			name:     "JIRA formatted text",
			body:     "*Tested* on _v2.0.0_ ??Result:?? *Fixed*",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isQAComment(tt.body)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseQAComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		expected domain.QAComment
	}{
		{
			name: "full comment with version and result",
			body: "Tested on v1.2.3\nResult: Fixed\nComment: All tests passed",
			expected: domain.QAComment{
				SoftwareVersion: "v1.2.3",
				TestResult:      "Fixed",
				Comment:         "All tests passed",
			},
		},
		{
			name: "comment with alternative version format",
			body: "Tested on SW v2.0.0\nResult: Not Fixed\nComment: Issue still exists",
			expected: domain.QAComment{
				SoftwareVersion: "v2.0.0",
				TestResult:      "Not Fixed",
				Comment:         "Issue still exists",
			},
		},
		{
			name: "could not test scenario",
			body: "Could not test on SW v1.0.0\nResult: Could not test",
			expected: domain.QAComment{
				SoftwareVersion: "v1.0.0",
				TestResult:      "Could not test",
			},
		},
		{
			name: "comment with notes instead of comment",
			body: "Tested on v3.0.0\nResult: Fixed\nNotes: Additional validation required",
			expected: domain.QAComment{
				SoftwareVersion: "v3.0.0",
				TestResult:      "Fixed",
				Comment:         "Additional validation required",
			},
		},
		{
			name: "comment with passed/failed result",
			body: "Tested on v1.1\nResult: Passed\nComment: Functionality works",
			expected: domain.QAComment{
				SoftwareVersion: "v1.1",
				TestResult:      "Fixed", // Should normalize "Passed" to "Fixed"
				Comment:         "Functionality works",
			},
		},
		{
			name:     "empty comment",
			body:     "",
			expected: domain.QAComment{},
		},
		{
			name: "JIRA formatted comment",
			body: "*Tested* on {{v2.0.0}}\n*Result:* Fixed\n*Comment:* ??All good??",
			expected: domain.QAComment{
				SoftwareVersion: "v2.0.0",
				TestResult:      "Fixed",
				Comment:         "All good",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQAComment(tt.body)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.SoftwareVersion, result.SoftwareVersion)
			assert.Equal(t, tt.expected.TestResult, result.TestResult)
			assert.Equal(t, tt.expected.Comment, result.Comment)
		})
	}
}

func TestRemoveJiraFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove bold formatting",
			input:    "*bold text*",
			expected: "bold text",
		},
		{
			name:     "remove italic formatting",
			input:    "_italic text_",
			expected: "italic text",
		},
		{
			name:     "remove monospace formatting",
			input:    "??monospace text??",
			expected: "monospace text",
		},
		{
			name:     "remove color formatting",
			input:    "{color:red}colored text{color}",
			expected: "colored text",
		},
		{
			name:     "remove link formatting",
			input:    "[link text|http://example.com]",
			expected: "link text",
		},
		{
			name:     "complex formatting",
			input:    "*Tested* on {{v2.0.0}} with _result_ ??Passed?? [link|http://test.com]",
			expected: "Tested on v2.0.0 with result Passed link",
		},
		{
			name:     "multiple formatting types",
			input:    "{color:blue}*Important* _notice_{color}",
			expected: "Important notice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeJiraFormatting(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveJiraFormattingWithLinks(t *testing.T) {
	t.Parallel()

	// Test link removal specifically
	input := "Tested on v1.0.0, see [results|https://jira.example.com/results] for details"
	expected := "Tested on v1.0.0, see results for details"
	result := removeJiraFormatting(input)
	assert.Equal(t, expected, result)

	// Test link without URL part
	input2 := "See [this link] for more info"
	expected2 := "See this link for more info"
	result2 := removeJiraFormatting(input2)
	assert.Equal(t, expected2, result2)
}

func TestParseQACommentResultNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "normalize passed to fixed",
			body:     "Tested on v1.0.0\nResult: Passed",
			expected: "Fixed",
		},
		{
			name:     "normalize failed to not fixed",
			body:     "Tested on v1.0.0\nResult: Failed",
			expected: "Not Fixed",
		},
		{
			name:     "preserve could not test",
			body:     "Could not test on SW v1.0",
			expected: "Could not test",
		},
		{
			name:     "infer result from keywords",
			body:     "Tested on v1.0.0\nFixed after validation",
			expected: "Fixed",
		},
		{
			name:     "infer not fixed from keywords",
			body:     "Tested on v1.0\nNot fixed yet",
			expected: "Not Fixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQAComment(tt.body)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.TestResult)
		})
	}
}
