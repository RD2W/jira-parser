package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rd2w/jira-parser/internal/domain"
	"github.com/rd2w/jira-parser/internal/infrastructure/config"
	"github.com/spf13/cobra"
)

func NewExportCommand() *cobra.Command {
	var ticketsFile string
	var outputFormat string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "export [issue-key...]",
		Short: "Export all QA comments as JSON or HTML",
		Long: `Export all QA comments as JSON or HTML.
If tickets are provided as arguments, they will be used instead of the tickets file.
If no arguments are provided, loads tickets from the specified file or from ./configs/tickets.yaml by default.
Example: jira-parser export TOS-30690 TOS-30692
Example: jira-parser export --tickets-file ./my-tickets.yaml
Example: jira-parser export --tickets-file ./my-tickets.yaml --format html --output-dir ./QA_comments`,
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

			// Получаем имя файла без расширения для формирования имени выходного файла
			baseFileName := "export"
			if ticketsFile != "" {
				baseFileName = strings.TrimSuffix(filepath.Base(ticketsFile), filepath.Ext(ticketsFile))
			}

			// Используем указанную директорию или по умолчанию QA_comments
			if outputDir == "" {
				outputDir = "./QA_comments"
			}

			// Создаем директорию, если она не существует
			if _, err := os.Stat(outputDir); os.IsNotExist(err) {
				err := os.MkdirAll(outputDir, 0755)
				if err != nil {
					log.Fatalf("Failed to create output directory: %v", err)
				}
			}

			// Формируем имя файла с датой и временем
			currentTime := strings.ReplaceAll(time.Now().Format("2006-01-02_15:04:05"), ":", "-")
			outputFileName := fmt.Sprintf("%s/%s_%s", outputDir, baseFileName, currentTime)

			// Определяем формат вывода
			switch strings.ToLower(outputFormat) {
			case "html":
				exportToHTML(issuesList, outputFileName)
			case "json":
				pretty, _ := cmd.Flags().GetBool("pretty")
				exportToJSON(issuesList, outputFileName, pretty)
			default:
				// По умолчанию экспортируем в JSON
				pretty, _ := cmd.Flags().GetBool("pretty")
				exportToJSON(issuesList, outputFileName, pretty)
			}
		},
	}

	cmd.Flags().BoolP("pretty", "p", false, "Pretty print JSON output")
	cmd.Flags().StringVarP(&ticketsFile, "tickets-file", "f", "", "Path to the YAML file containing the list of tickets (default: ./configs/tickets.yaml)")
	cmd.Flags().StringVarP(&outputFormat, "format", "F", "json", "Output format: json or html")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for exported files (default: ./QA_comments)")
	return cmd
}

func exportToJSON(issuesList *domain.IssuesList, fileName string, pretty bool) {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(issuesList, "", "  ")
	} else {
		output, err = json.Marshal(issuesList)
	}

	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Добавляем расширение .json к имени файла
	fileNameWithExt := fileName + ".json"
	err = os.WriteFile(fileNameWithExt, output, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON file: %v", err)
	}

	fmt.Printf("Exported results to %s\n", fileNameWithExt)
}

func exportToHTML(issuesList *domain.IssuesList, fileName string) {
	htmlContent := generateHTMLReport(issuesList)

	// Добавляем расширение .html к имени файла
	fileNameWithExt := fileName + ".html"
	err := os.WriteFile(fileNameWithExt, []byte(htmlContent), 0644)
	if err != nil {
		log.Fatalf("Error writing HTML file: %v", err)
	}

	fmt.Printf("Exported results to %s\n", fileNameWithExt)
}

func generateHTMLReport(issuesList *domain.IssuesList) string {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>QA Comments Report</title>
	<meta charset="UTF-8">
	<style>
		body {
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			margin: 20px;
			background-color: #f9f9f9;
		}
		.container {
			max-width: 1200px;
			margin: 0 auto;
			background-color: white;
			padding: 20px;
			border-radius: 8px;
			box-shadow: 0 2px 10px rgba(0,0,0,0.1);
		}
		h1 {
			color: #333;
			border-bottom: 2px solid #007acc;
			padding-bottom: 10px;
		}
		.issue {
			border: 1px solid #ddd;
			margin: 20px 0;
			padding: 20px;
			border-radius: 8px;
			background-color: #ffffff;
		}
		.issue-key {
			font-weight: bold;
			font-size: 1.4em;
			color: #007acc;
			margin-bottom: 10px;
		}
		.issue-summary {
			color: #666;
			margin: 10px 0;
			font-style: italic;
		}
		.issue-info {
			display: flex;
			gap: 20px;
			margin: 10px 0;
			color: #555;
		}
		.comment {
			margin: 15px 0;
			padding: 15px;
			background-color: #f9f9f9;
			border-left: 4px solid #007acc;
			border-radius: 0 4px 4px 0;
			box-shadow: 0 1px 3px rgba(0,0,0,0.1);
		}
		.comment-header {
			font-weight: bold;
			margin-bottom: 8px;
			color: #333;
			font-size: 1.1em;
		}
		.comment-details {
			margin-top: 10px;
		}
		.comment-field {
			display: flex;
			align-items: center;
			margin-bottom: 8px;
			font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
			font-size: 1em;
		}
		.comment-label {
			font-weight: bold;
			color: #555;
			min-width: 80px;
		}
		.comment-value {
			color: #333;
			flex: 1;
		}
		.result-fixed { color: green; }
		.result-not-fixed { color: red; }
		.result-partially-fixed { color: orange; }
		.result-could-not-test { color: blue; }
		.result-blocked { color: red; }
		.result-pending { color: gray; }
		.result-passed { color: green; }
		.result-failed { color: red; }
		.result-verified { color: green; }
		.result-resolved { color: green; }
		.result-ok { color: green; }
		.result-nok { color: red; }
	</style>
</head>
<body>
	<div class="container">
		<h1>QA Comments Report</h1>`

	for _, issue := range issuesList.Issues {
		html += fmt.Sprintf(`
		<div class="issue">
			<div class="issue-key">%s</div>
			<div class="issue-summary">%s</div>`, issue.Key, issue.Summary)

		if issue.AssigneeEmail != "" || issue.QaOwnerEmail != "" {
			html += `<div class="issue-info">`
			if issue.AssigneeEmail != "" {
				html += fmt.Sprintf("<div><strong>Assigned:</strong> %s</div>", issue.AssigneeEmail)
			}
			if issue.QaOwnerEmail != "" {
				html += fmt.Sprintf("<div><strong>QA Owner:</strong> %s</div>", issue.QaOwnerEmail)
			}
			html += `</div>`
		}

		html += fmt.Sprintf("<div><strong>Found %d QA comments:</strong></div>", len(issue.Comments))

		for j, comment := range issue.Comments {
			resultClass := ""
			switch comment.TestResult {
			case "Fixed", "OK", "Passed", "Verified", "Resolved":
				resultClass = "result-fixed"
			case "Not Fixed", "NOK", "Failed", "Blocked":
				resultClass = "result-not-fixed"
			case "Partially Fixed", "Partially OK":
				resultClass = "result-partially-fixed"
			case "Could not test", "Pending":
				resultClass = "result-could-not-test"
			default:
				resultClass = ""
			}

			// Parse the timestamp and format it as "YYYY-MM-DD HH:MM:SS"
			createdTime := comment.Created
			if comment.Created != "" {
				var t time.Time
				var err error

				// Try multiple formats since JIRA can return different timestamp formats
				t, err = time.Parse("2006-01-02T15:04:05.000-0700", comment.Created)
				if err != nil {
					t, err = time.Parse(time.RFC3339, comment.Created)
				}
				if err != nil {
					t, err = time.Parse("2006-01-02T15:04:05-0700", comment.Created)
				}

				if err == nil {
					createdTime = t.Format("2006-01-02 15:04:05")
				}
			}

			// Start building comment HTML
			html += fmt.Sprintf(`
			<div class="comment">
				<div class="comment-header">Comment #%d (%s) from %s</div>
				<div class="comment-details">
					<div class="comment-field">
						<span class="comment-label">Version:</span>
						<span class="comment-value">%s</span>
					</div>
					<div class="comment-field">
						<span class="comment-label">Result:</span>
						<span class="comment-value %s">%s</span>
					</div>`, j+1, createdTime, comment.AuthorEmail, comment.SoftwareVersion, resultClass, comment.TestResult)

			// Add Note field only if comment.Comment is not empty
			if comment.Comment != "" {
				html += fmt.Sprintf(`
					<div class="comment-field">
						<span class="comment-label">Note:</span>
						<span class="comment-value">%s</span>
					</div>`, comment.Comment)
			}

			// Close the comment
			html += `
				</div>
			</div>`
		}

		html += `
		</div>`
	}

	html += `
	</div>
</body>
</html>`

	return html
}
