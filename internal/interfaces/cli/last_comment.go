package cli

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/rd2w/jira-parser/internal/infrastructure/config"
	"github.com/spf13/cobra"
)

func NewLastCommentCommand() *cobra.Command {
	var ticketsFile string

	cmd := &cobra.Command{
		Use:   "last-comment [issue-keys...]",
		Short: "Get the last QA comment for issue(s)",
		Long: `Get the last QA comment for issue(s).
If no issue keys are provided, reads tickets from the specified file or from configs/tickets.yaml by default.
Example: jira-parser last-comment TOS-30690 TOS-30692
Example: jira-parser last-comment --tickets-file ./my-tickets.yaml`,
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			service, err := createCommentService()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			var ticketKeys []string

			// If arguments are provided, use them as ticket keys
			if len(args) > 0 {
				ticketKeys = args
			} else {
				// Otherwise, load ticket list from the specified file or default
				ticketsFilePath := ticketsFile
				if ticketsFilePath == "" {
					ticketsFilePath = "./configs/tickets.yaml"
				}

				ticketsConfig, err := config.LoadTickets(ticketsFilePath)
				if err != nil {
					log.Fatalf("Failed to read tickets file: %v", err)
				}

				ticketKeys = ticketsConfig.Tickets
			}

			if len(ticketKeys) == 0 {
				log.Fatalf("No tickets provided either as arguments or in tickets file")
			}

			// Process each ticket and get the last comment
			for _, ticketKey := range ticketKeys {
				comment, err := service.GetLastComment(ticketKey)
				if err != nil {
					log.Printf("Failed to get last comment for %s: %v", ticketKey, err)
					continue
				}

				printLastComment(ticketKey, comment)
				fmt.Println(strings.Repeat("-", 30)) // separator between tickets
			}
		},
	}

	cmd.Flags().StringVarP(&ticketsFile, "tickets-file", "f", "", "Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)")

	return cmd
}

func printLastComment(issueKey string, comment *domain.QAComment) {
	if comment == nil {
		fmt.Printf("Last QA Comment for %s:\n", issueKey)
		fmt.Println("No QA comments found")
		return
	}

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
		}
	}

	// Print header with author and date
	if comment.AuthorEmail != "" && createdTime != "" {
		fmt.Printf("Last QA comment on %s by %s (%s)\n", issueKey, comment.AuthorEmail, createdTime)
	} else if comment.AuthorEmail != "" {
		fmt.Printf("Last QA comment on %s by %s\n", issueKey, comment.AuthorEmail)
	} else if createdTime != "" {
		fmt.Printf("Last QA comment on %s (%s)\n", issueKey, createdTime)
	} else {
		fmt.Printf("Last QA comment on %s\n", issueKey)
	}

	if comment.SoftwareVersion != "" {
		fmt.Printf("Version: %s\n", comment.SoftwareVersion)
	}

	// Use colored output for test result
	resultColor := getColorForStatus(comment.TestResult)
	_, _ = resultColor.Printf("Result: %s\n", comment.TestResult)

	if comment.Comment != "" {
		fmt.Printf("Comment: %s\n", comment.Comment)
	}
}
