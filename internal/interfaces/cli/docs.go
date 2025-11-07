package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewDocsCommand() *cobra.Command {
	var outputDir string
	var format string

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate CLI documentation",
		Long: `Generate documentation for jira-parser CLI commands.
Supports multiple output formats: markdown, man, rest, and yaml.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create output directory if it doesn't exist
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Printf("Error creating output directory: %v\n", err)
				return
			}

			switch format {
			case "markdown", "md":
				generateOfflineMarkdownDocs(outputDir)
			case "offline-help":
				// Generate a help file with all command information
				generateOfflineHelp(outputDir)
			default:
				fmt.Println("This command requires the github.com/spf13/cobra/doc package which is not available in offline mode.")
				fmt.Println("To generate documentation, run:")
				fmt.Println(" go get github.com/spf13/cobra/doc")
				fmt.Println("  jira-parser docs --format=markdown --output=./docs")
				fmt.Println("")
				fmt.Println("Alternatively, use the offline mode:")
				fmt.Println("  jira-parser docs --format=offline-help --output=./docs")
			}
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "./docs", "Output directory for documentation")
	cmd.Flags().StringVarP(&format, "format", "F", "markdown", "Output format (markdown, offline-help)")

	return cmd
}

func generateOfflineMarkdownDocs(outputDir string) {
	content := `# jira-parser CLI Documentation

## Commands

### parse
Parse all QA comments for an issue

Usage: jira-parser parse <issue-key>

Flags:
  -r, --result string     Filter comments by test result (e.g., Fixed, Not Fixed, etc.)
  -d, --date-from string  Filter comments created after specified date (format: YYYY-MM-DD)
  -t, --date-to string    Filter comments created before specified date (format: YYYY-MM-DD)

### last-comment
Get the last QA comment for an issue

Usage: jira-parser last-comment [issue-key...]

Flags:
  -f, --tickets-file Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)

### export
Export all QA comments as JSON or HTML

Usage: jira-parser export [issue-key...]

Flags:
  -p, --pretty       Pretty print JSON output
  -F, --format       Output format (json or html) (default "json")
  -o, --output-dir   Output directory for exported files (default "./QA_comments")
  -f, --tickets-file Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)

### parse-multiple
Parse QA comments for multiple tickets from tickets file or command line arguments

Usage: jira-parser parse-multiple [tickets...]

Flags:
  -r, --result string     Filter comments by test result (e.g., Fixed, Not Fixed, etc.)
  -d, --date-from string  Filter comments created after specified date (format: YYYY-MM-DD)
  -t, --date-to string    Filter comments created before specified date (format: YYYY-MM-DD)
  -f, --tickets-file      Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)

### version
Print the version number of jira-parser

Usage: jira-parser version

### docs
Generate CLI documentation

Usage: jira-parser docs

Flags:
  -F, --format string   Output format (markdown, offline-help) (default "markdown")
  -o, --output string   Output directory for documentation (default "./docs")

### tutorial
Interactive tutorial for jira-parser

Usage: jira-parser tutorial

## Examples

Parse an issue:
  jira-parser parse TOS-30690

Parse with result filter:
  jira-parser parse TOS-30690 --result="Fixed"

Parse with date range:
  jira-parser parse TOS-30690 --date-from=2023-01-01 --date-to=2023-12-31

Export as JSON:
  jira-parser export TOS-30690 --pretty

Export as HTML:
  jira-parser export TOS-30690 --format html --output-dir ./reports

Parse multiple tickets:
 jira-parser parse-multiple TOS-30690 TOS-30692

Parse multiple tickets from file:
 jira-parser parse-multiple --tickets-file ./my-tickets.yaml

Get last comment:
  jira-parser last-comment TOS-30690

Get last comment for multiple issues:
  jira-parser last-comment TOS-30690 TOS-30692

Get last comment from tickets file:
  jira-parser last-comment --tickets-file ./my-tickets.yaml
`

	filePath := filepath.Join(outputDir, "jira-parser.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing documentation file: %v\n", err)
		return
	}
	fmt.Printf("Offline documentation generated successfully in %s\n", filePath)
}

func generateOfflineHelp(outputDir string) {
	content := `jira-parser Help Documentation
============================

NAME:
   jira-parser - Parse QA comments from JIRA issues

USAGE:
   jira-parser [global options] command [command options] [arguments...]

COMMANDS:
   parse           Parse all QA comments for an issue
   last-comment    Get the last QA comment for an issue
   export          Export all QA comments as JSON
   parse-multiple  Parse QA comments for multiple tickets from tickets file or command line arguments
   version         Print the version number of jira-parser
   docs            Generate CLI documentation
   tutorial        Interactive tutorial for jira-parser
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help

COMMAND SPECIFICS:

parse command:
  Usage: jira-parser parse <issue-key>
  Flags:
    -r, --result string     Filter comments by test result (e.g., Fixed, Not Fixed, etc.)
    -d, --date-from string  Filter comments created after specified date (format: YYYY-MM-DD)
    -t, --date-to string    Filter comments created before specified date (format: YYYY-MM-DD)

last-comment command:
 Usage: jira-parser last-comment <issue-key>

export command:
   Usage: jira-parser export [issue-key...]
   Flags:
     --pretty, -p        Pretty print JSON output
     --format, -F        Output format (json or html) (default: "json")
     --output-dir, -o    Output directory for exported files (default: "./QA_comments")
     --tickets-file, -f  Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)

last-comment command:
   Usage: jira-parser last-comment [issue-key...]
   Flags:
     -f, --tickets-file      Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)

parse-multiple command:
  Usage: jira-parser parse-multiple [tickets...]
   Flags:
     -r, --result string     Filter comments by test result (e.g., Fixed, Not Fixed, etc.)
     -d, --date-from string  Filter comments created after specified date (format: YYYY-MM-DD)
     -t, --date-to string    Filter comments created before specified date (format: YYYY-MM-DD)
     -f, --tickets-file      Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)

docs command:
 Usage: jira-parser docs
  Flags:
    --format value, -F value   Output format (markdown, offline-help) (default: "markdown")
    --output value, -o value   Output directory for documentation (default: "./docs")

tutorial command:
  Usage: jira-parser tutorial
`

	filePath := filepath.Join(outputDir, "jira-parser-help.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing help file: %v\n", err)
		return
	}
	fmt.Printf("Offline help file generated successfully in %s\n", filePath)
}
