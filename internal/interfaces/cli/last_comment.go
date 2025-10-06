package cli

import (
	"fmt"
	"log"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/spf13/cobra"
)

func NewLastCommentCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "last-comment <issue-key>",
		Short: "Get the last QA comment for an issue",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			service, err := createCommentService()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			comment, err := service.GetLastComment(args[0])
			if err != nil {
				log.Fatalf("Failed to get last comment: %v", err)
			}

			printLastComment(args[0], comment)
		},
	}
}

func printLastComment(issueKey string, comment *domain.QAComment) {
	fmt.Printf("Last QA Comment for %s:\n", issueKey)
	if comment == nil {
		fmt.Println("No QA comments found")
		return
	}

	if comment.SoftwareVersion != "" {
		fmt.Printf("Version: %s\n", comment.SoftwareVersion)
	}
	fmt.Printf("Result: %s\n", comment.TestResult)
	if comment.Comment != "" {
		fmt.Printf("Comment: %s\n", comment.Comment)
	}
}
