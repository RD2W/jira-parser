package cli

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/spf13/cobra"
)

func init() {
	// Ensure colors are enabled
	color.NoColor = false
}

func NewParseCommand() *cobra.Command {
	var resultFilter string
	var dateFrom string
	var dateTo string

	cmd := &cobra.Command{
		Use:   "parse <issue-key>",
		Short: "Parse all QA comments for an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := createCommentService()
			if err != nil {
				return fmt.Errorf("error creating comment service: %w", err)
			}

			issue, err := service.ParseComments(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse comments: %w", err)
			}

			// Apply filters if specified
			var filteredComments []domain.QAComment
			for _, comment := range issue.Comments {
				// Apply result filter
				if resultFilter != "" && comment.TestResult != resultFilter {
					continue
				}

				// Apply date filters
				if dateFrom != "" || dateTo != "" {
					// Parse comment creation date - try multiple formats since JIRA can return different timestamp formats
					var commentTime time.Time
					var err error

					// Try the JIRA format with milliseconds and timezone offset: 2025-08-12T16:35:38.514+0300
					commentTime, err = time.Parse("2006-01-02T15:04:05.000-0700", comment.Created)
					if err != nil {
						// Try standard RFC3339 format
						commentTime, err = time.Parse(time.RFC3339, comment.Created)
					}
					if err != nil {
						// Try another common format
						commentTime, err = time.Parse("2006-01-02T15:04:05-0700", comment.Created)
					}

					if err != nil {
						log.Printf("Warning: Could not parse comment creation date: %s", comment.Created)
						// Skip date filtering for this comment if we can't parse the date
						filteredComments = append(filteredComments, comment)
						continue
					}

					// Apply date-from filter
					if dateFrom != "" {
						fromTime, err := time.Parse("2006-01-02", dateFrom)
						if err != nil {
							log.Printf("Warning: Invalid date-from format: %s", dateFrom)
						} else if commentTime.Before(fromTime) {
							continue
						}
					}

					// Apply date-to filter
					if dateTo != "" {
						toTime, err := time.Parse("2006-01-02", dateTo)
						if err != nil {
							log.Printf("Warning: Invalid date-to format: %s", dateTo)
						} else if commentTime.After(toTime) {
							continue
						}
					}
				}

				filteredComments = append(filteredComments, comment)
			}

			// Update issue with filtered comments
			issue.Comments = filteredComments

			printIssueComments(issue)
			return nil
		},
	}

	cmd.Flags().StringVar(&resultFilter, "result", "", "Filter comments by test result (e.g., Fixed, Not Fixed, etc.)")
	cmd.Flags().StringVar(&dateFrom, "date-from", "", "Filter comments created after specified date (format: YYYY-MM-DD)")
	cmd.Flags().StringVar(&dateTo, "date-to", "", "Filter comments created before specified date (format: YYYY-MM-DD)")

	return cmd
}

func printIssueComments(issue *domain.Issue) {
	if issue.Summary != "" {
		fmt.Printf("\n%s: %s\n", issue.Key, issue.Summary)
	} else {
		fmt.Printf("\n%s\n", issue.Key)
	}

	// Выводим информацию о назначенном и QA владельце
	if issue.AssigneeEmail != "" {
		fmt.Printf("Assigned: %s\n", issue.AssigneeEmail)
	}
	if issue.QaOwnerEmail != "" {
		fmt.Printf("QA Owner: %s\n", issue.QaOwnerEmail)
	}

	fmt.Printf("Found %d QA comments:\n\n", len(issue.Comments))

	for i, comment := range issue.Comments {
		// Format the creation date for display
		createdTime := ""
		if comment.Created != "" {
			// Parse the timestamp and format it as "YYYY-MM-DD HH:MM:SS"
			// Try multiple formats since JIRA can return different timestamp formats
			var t time.Time
			var err error

			// Try the JIRA format with milliseconds and timezone offset: 2025-08-12T16:35:38.514+0300
			t, err = time.Parse("2006-01-02T15:04:05.000-0700", comment.Created)
			if err != nil {
				// Try standard RFC3339 format
				t, err = time.Parse(time.RFC3339, comment.Created)
			}
			if err != nil {
				// Try another common format
				t, err = time.Parse("2006-01-02T15:04:05-0700", comment.Created)
			}

			if err == nil {
				createdTime = t.Format("2006-01-02 15:04:05")
				if comment.AuthorEmail != "" {
					fmt.Printf("Comment #%d (%s) from %s:\n", i+1, createdTime, comment.AuthorEmail)
				} else {
					fmt.Printf("Comment #%d (%s):\n", i+1, createdTime)
				}
			} else {
				if comment.AuthorEmail != "" {
					fmt.Printf("Comment #%d from %s:\n", i+1, comment.AuthorEmail)
				} else {
					fmt.Printf("Comment #%d:\n", i+1)
				}
			}
		} else {
			fmt.Printf("Comment #%d:\n", i+1)
		}
		fmt.Printf("  Version: %s\n", comment.SoftwareVersion)

		// Use colored output for test result
		resultColor := getColorForStatus(comment.TestResult)
		_, _ = resultColor.Printf("  Result: %s\n", comment.TestResult)

		if comment.Comment != "" {
			fmt.Printf("  Info: %s\n", comment.Comment)
		}
		fmt.Println()
	}
}
