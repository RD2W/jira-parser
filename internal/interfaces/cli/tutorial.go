package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewTutorialCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tutorial",
		Short: "Interactive tutorial for jira-parser",
		Long:  `Interactive tutorial that guides new users through the basic usage of jira-parser.`,
		Run: func(cmd *cobra.Command, args []string) {
			runTutorial()
		},
	}
}

func runTutorial() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("🚀 Welcome to jira-parser Interactive Tutorial!")
	fmt.Println("")
	fmt.Println("This tutorial will guide you through the basic usage of jira-parser.")
	fmt.Println("")

	// Шаг 1: Объяснение назначения инструмента
	fmt.Println("Step 1: Understanding jira-parser")
	fmt.Println("==================================")
	fmt.Println("jira-parser is a tool for parsing and analyzing QA comments from JIRA issues.")
	fmt.Println("It extracts structured information from comments about testing results,")
	fmt.Println("software versions, and additional notes.")
	fmt.Println("")

	pause(scanner)

	// Шаг 2: Конфигурация
	fmt.Println("Step 2: Configuration")
	fmt.Println("=====================")
	fmt.Println("Before using jira-parser, you need to configure it with your JIRA credentials.")
	fmt.Println("Create a config.yaml file in the configs/ directory with the following structure:")
	fmt.Println("")
	fmt.Println("jira:")
	fmt.Println("  base_url: \"https://your-domain.atlassian.net\"")
	fmt.Println("  username: \"your-email@example.com\"  # for Atlassian Cloud use email")
	fmt.Println(" token: \"your-api-token\"             # API token for authentication")
	fmt.Println("")
	fmt.Println("For Atlassian Cloud, use your email as username and an API token.")
	fmt.Println("For self-hosted JIRA, you can use username and password or API token.")
	fmt.Println("")

	pause(scanner)

	// Шаг 3: Основные команды
	fmt.Println("Step 3: Main Commands")
	fmt.Println("=====================")
	fmt.Println("jira-parser has several main commands:")
	fmt.Println("")
	fmt.Println("1. parse - Parse all QA comments for an issue")
	fmt.Println("   Usage: jira-parser parse TOS-30690")
	fmt.Println("")
	fmt.Println("2. last-comment - Get the last QA comment for an issue")
	fmt.Println("   Usage: jira-parser last-comment TOS-30690")
	fmt.Println("")
	fmt.Println("3. export - Export all QA comments as JSON")
	fmt.Println("   Usage: jira-parser export TOS-30690 --pretty")
	fmt.Println("")
	fmt.Println("4. parse-multiple - Parse QA comments for multiple tickets")
	fmt.Println("   Usage: jira-parser parse-multiple TOS-30690 TOS-30692")
	fmt.Println("")
	fmt.Println("5. version - Print the version number of jira-parser")
	fmt.Println("   Usage: jira-parser version")
	fmt.Println("")

	pause(scanner)

	// Шаг 4: Фильтрация
	fmt.Println("Step 4: Filtering Results")
	fmt.Println("=========================")
	fmt.Println("You can filter results by test result or date:")
	fmt.Println("")
	fmt.Println("Filter by result:")
	fmt.Println("  jira-parser parse TOS-30690 --result=\"Fixed\"")
	fmt.Println("")
	fmt.Println("Filter by date range:")
	fmt.Println("  jira-parser parse TOS-30690 --date-from=2023-01-01 --date-to=2023-12-31")
	fmt.Println("")
	fmt.Println("Combine filters:")
	fmt.Println("  jira-parser parse TOS-30690 --result=\"Fixed\" --date-from=2023-01-01")
	fmt.Println("")

	pause(scanner)

	// Шаг 5: Практическое задание
	fmt.Println("Step 5: Hands-on Exercise")
	fmt.Println("=========================")
	fmt.Println("Now it's time to try jira-parser yourself!")
	fmt.Println("")
	fmt.Println("1. Make sure you have configured your JIRA credentials in configs/config.yaml")
	fmt.Println("2. Try running: jira-parser parse <your-issue-key>")
	fmt.Println("3. Try filtering: jira-parser parse <your-issue-key> --result=\"Fixed\"")
	fmt.Println("")
	fmt.Println("Would you like to try parsing a test issue now? (y/n): ")

	if scanner.Scan() {
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if input == "y" || input == "yes" {
			fmt.Println("")
			fmt.Print("Enter an issue key to parse (e.g., TOS-30690): ")
			if scanner.Scan() {
				issueKey := strings.TrimSpace(scanner.Text())
				if issueKey != "" {
					fmt.Printf("\nTo parse this issue, run:\n")
					fmt.Printf("  jira-parser parse %s\n", issueKey)
					fmt.Println("")
					fmt.Printf("To parse with result filter, run:\n")
					fmt.Printf("  jira-parser parse %s --result=\"Fixed\"\n", issueKey)
					fmt.Println("")
				}
			}
		}
	}

	fmt.Println("")
	fmt.Println("🎉 Congratulations! You've completed the jira-parser tutorial.")
	fmt.Println("")
	fmt.Println("Additional resources:")
	fmt.Println("- Run 'jira-parser --help' to see all commands")
	fmt.Println("- Check the README.md file for detailed documentation")
	fmt.Println("- Visit the project repository for more examples")
}

func pause(scanner *bufio.Scanner) {
	fmt.Print("Press Enter to continue... ")
	scanner.Scan()
	fmt.Println("")
}
