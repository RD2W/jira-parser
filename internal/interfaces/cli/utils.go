package cli

import (
	"github.com/fatih/color"
)

// getColorForStatus возвращает цвет в зависимости от статуса тестирования
func getColorForStatus(status string) *color.Color {
	switch status {
	case "Fixed", "OK", "Passed", "Verified", "Resolved":
		return color.New(color.FgGreen)
	case "Not Fixed", "NOK", "Failed", "Blocked":
		return color.New(color.FgRed)
	case "Partially Fixed", "Partially OK":
		return color.New(color.FgHiYellow)
	case "Could not test", "Pending":
		return color.New(color.FgBlue)
	default:
		return color.New(color.Reset)
	}
}
