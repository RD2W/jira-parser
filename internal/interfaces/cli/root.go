package cli

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	rootCmd.AddCommand(NewDocsCommand())
	rootCmd.AddCommand(NewTutorialCommand())

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
	var resultFilter string
	var dateFrom string
	var dateTo string
	var ticketsFile string

	cmd := &cobra.Command{
		Use:   "parse-multiple [tickets...]",
		Short: "Parse QA comments for multiple tickets from tickets file or command line arguments",
		Long: `Parse QA comments for multiple tickets.
If tickets are provided as arguments, they will be used instead of the tickets file.
If no arguments are provided, loads tickets from the specified file or from ./configs/tickets.yaml by default.
Example: jira-parser parse-multiple TOS-30690 TOS-30692
Example: jira-parser parse-multiple --tickets-file ./my-tickets.yaml`,
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
				// Иначе загружаем список тикетов из файла
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

			issuesList, err := service.ParseMultipleTickets(ticketKeys)
			if err != nil {
				log.Fatalf("Failed to parse multiple tickets: %v", err)
			}

			// Apply filters if specified
			if resultFilter != "" || dateFrom != "" || dateTo != "" {
				for i := range issuesList.Issues {
					var filteredComments []domain.QAComment
					for _, comment := range issuesList.Issues[i].Comments {
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
					issuesList.Issues[i].Comments = filteredComments
				}
			}

			printMultipleIssues(issuesList)
		},
	}

	cmd.Flags().StringVarP(&resultFilter, "result", "r", "", "Filter comments by test result (e.g., Fixed, Not Fixed, etc.)")
	cmd.Flags().StringVarP(&dateFrom, "date-from", "d", "", "Filter comments created after specified date (format: YYYY-MM-DD)")
	cmd.Flags().StringVarP(&dateTo, "date-to", "t", "", "Filter comments created before specified date (format: YYYY-MM-DD)")
	cmd.Flags().StringVarP(&ticketsFile, "tickets-file", "f", "", "Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)")

	return cmd
}

func printMultipleIssues(issuesList *domain.IssuesList) {
	fmt.Printf("\nChecked %d issues with QA comments:\n\n", len(issuesList.Issues))

	for _, issue := range issuesList.Issues {
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
				if comment.AuthorEmail != "" {
					fmt.Printf("Comment #%d from %s:\n", i+1, comment.AuthorEmail)
				} else {
					fmt.Printf("Comment #%d:\n", i+1)
				}
			}
			fmt.Printf("  Version: %s\n", comment.SoftwareVersion)

			// Use colored output for test result
			resultColor := getColorForStatus(comment.TestResult)
			_, _ = resultColor.Printf("  Result: %s\n", comment.TestResult)

			if comment.Comment != "" {
				fmt.Printf("  Comment: %s\n", comment.Comment)
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
