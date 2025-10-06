package cli

import (
	"fmt"
	"log"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/spf13/cobra"
)

func NewParseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "parse <issue-key>",
		Short: "Parse all QA comments for an issue",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service, err := createCommentService()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			issue, err := service.ParseComments(args[0])
			if err != nil {
				log.Fatalf("Failed to parse comments: %v", err)
			}

			printIssueComments(issue)
		},
	}
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
