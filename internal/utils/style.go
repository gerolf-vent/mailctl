package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	BlackStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "0",
		Dark:  "8",
	})
	RedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "1",
		Dark:  "9",
	})
	GreenStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "2",
		Dark:  "10",
	})
	YellowStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "3",
		Dark:  "11",
	})
	BlueStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "4",
		Dark:  "12",
	})
	MagentaStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "5",
		Dark:  "13",
	})
	CyanStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "6",
		Dark:  "14",
	})
	WhiteStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "7",
		Dark:  "15",
	})

	TableHeaderStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	TableRowStyle    = lipgloss.NewStyle().Padding(0, 1)
)

var (
	ErrorPrefix   = RedStyle.Bold(true).Render("Error: ")
	WarningPrefix = YellowStyle.Bold(true).Render("Warning: ")
)

func PrintError(err error) {
	fmt.Println(ErrorPrefix + RedStyle.Render(err.Error()))
}

func PrintErrorWithMessage(message string, err error) {
	if err != nil {
		message = message + ": " + err.Error()
	}
	fmt.Println(ErrorPrefix + RedStyle.Render(message))
}

func PrintWarning(message string) {
	fmt.Println(WarningPrefix + YellowStyle.Render(message))
}

func PrintSuccess(message string) {
	fmt.Println(GreenStyle.Bold(true).Render(message))
}
