package cli

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/spf13/cobra"
)

func NewExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <issue-key>",
		Short: "Export all QA comments as JSON",
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

			pretty, _ := cmd.Flags().GetBool("pretty")
			printIssueCommentsJSON(issue, pretty)
		},
	}

	cmd.Flags().BoolP("pretty", "p", false, "Pretty print JSON output")
	return cmd
}

func printIssueCommentsJSON(issue *domain.Issue, pretty bool) {
	if pretty {
		output, err := json.MarshalIndent(issue, "", " ")
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		output, err := json.Marshal(issue)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}
		fmt.Println(string(output))
	}
}
