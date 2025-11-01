package cli

import (
	"fmt"
	"os"

	"github.com/rd2w/jira-parser/internal/application"
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

func createCommentService() (*application.CommentService, error) {
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	baseURL := viper.GetString("jira.base_url")
	username := viper.GetString("jira.username")
	token := viper.GetString("jira.token")

	if baseURL == "" {
		return nil, fmt.Errorf("jira.base_url is required in config")
	}
	if username == "" {
		return nil, fmt.Errorf("jira.username is required in config")
	}
	if token == "" {
		return nil, fmt.Errorf("jira.token is required in config")
	}

	jiraClient, err := jira.NewJiraClient(baseURL, username, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create JIRA client: %w", err)
	}

	return application.NewCommentService(jiraClient), nil
}
