package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rd2w/jira-parser/internal/application"
	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/rd2w/jira-parser/internal/infrastructure/config"
	"github.com/rd2w/jira-parser/internal/infrastructure/jira"
	"github.com/rd2w/jira-parser/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "jira-parser",
	Short: "Parse QA comments from JIRA issues",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(NewParseCommand())
	rootCmd.AddCommand(NewLastCommentCommand())
	rootCmd.AddCommand(NewExportCommand())
	rootCmd.AddCommand(NewVersionCommand())
	rootCmd.AddCommand(NewParseMultipleCommand())

	// Настройка конфигурации
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
}

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of jira-parser",
		Run: func(cmd *cobra.Command, args []string) {
			if version.App.Commit != "" && version.App.Date != "" {
				fmt.Printf("jira-parser %s (commit: %s, built: %s)\n", version.App.Version, version.App.Commit, version.App.Date)
			} else if version.App.Commit != "" {
				fmt.Printf("jira-parser %s (commit: %s)\n", version.App.Version, version.App.Commit)
			} else if version.App.Date != "" {
				fmt.Printf("jira-parser %s (built: %s)\n", version.App.Version, version.App.Date)
			} else {
				fmt.Printf("jira-parser %s\n", version.App.Version)
			}
		},
	}
}

func NewParseMultipleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "parse-multiple [tickets...]",
		Short: "Parse QA comments for multiple tickets from tickets file or command line arguments",
		Long: `Parse QA comments for multiple tickets.
If tickets are provided as arguments, they will be used instead of the tickets file.
Example: jira-parser parse-multiple TOS-30690 TOS-30692`,
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			service, err := createCommentService()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			var ticketKeys []string

			// Если переданы аргументы, используем их как тикеты
			if len(args) > 0 {
				ticketKeys = args
			} else {
				// Иначе загружаем список тикетов из отдельного файла
				ticketsConfig, err := config.LoadTickets("./configs/tickets.yaml")
				if err != nil {
					log.Fatalf("Failed to read tickets file: %v", err)
				}

				ticketKeys = ticketsConfig.Tickets
			}

			if len(ticketKeys) == 0 {
				log.Fatalf("No tickets provided either as arguments or in tickets file")
			}

			issuesList, err := service.ParseMultipleTickets(ticketKeys)
			if err != nil {
				log.Fatalf("Failed to parse multiple tickets: %v", err)
			}

			printMultipleIssues(issuesList)
		},
	}
}

func printMultipleIssues(issuesList *domain.IssuesList) {
	fmt.Printf("Found %d issues with QA comments:\n\n", len(issuesList.Issues))

	for _, issue := range issuesList.Issues {
		fmt.Printf("Issue: %s\n", issue.Key)
		fmt.Printf("Found %d QA comments:\n\n", len(issue.Comments))

		for i, comment := range issue.Comments {
			fmt.Printf("Comment #%d:\n", i+1)
			fmt.Printf(" Version: %s\n", comment.SoftwareVersion)
			fmt.Printf(" Result: %s\n", comment.TestResult)
			if comment.Comment != "" {
				fmt.Printf(" Comment: %s\n", comment.Comment)
			}
			fmt.Println()
		}
		fmt.Println(strings.Repeat("-", 50))
	}
}

func createCommentService() (*application.CommentService, error) {
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(viper.ConfigFileUsed())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	jiraClient, err := jira.NewJiraClient(cfg.BaseURL, cfg.Username, cfg.Token, cfg.Parsing)
	if err != nil {
		return nil, fmt.Errorf("failed to create JIRA client: %w", err)
	}

	return application.NewCommentService(jiraClient), nil
}
