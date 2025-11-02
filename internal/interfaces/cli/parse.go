package cli

import (
	"fmt"
	"log"
	"time"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/spf13/cobra"
)

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
					// Parse comment creation date
					commentTime, err := time.Parse(time.RFC3339, comment.Created)
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
	fmt.Printf("Issue: %s\n", issue.Key)
	fmt.Printf("Found %d QA comments:\n\n", len(issue.Comments))

	for i, comment := range issue.Comments {
		fmt.Printf("Comment #%d:\n", i+1)
		fmt.Printf("  Version: %s\n", comment.SoftwareVersion)
		fmt.Printf("  Result: %s\n", comment.TestResult)
		if comment.Comment != "" {
			fmt.Printf("  Comment: %s\n", comment.Comment)
		}
		fmt.Println()
	}
}
