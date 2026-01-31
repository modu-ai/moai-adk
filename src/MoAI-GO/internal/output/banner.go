package output

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Claude Code official terra cotta color
const ColorTerraCotta = "#DA7756"

// MoAI banner ASCII art
const MoaiBanner = `
███╗   ███╗          █████╗ ██╗       █████╗ ██████╗ ██╗  ██╗
████╗ ████║ ██████╗ ██╔══██╗██║      ██╔══██╗██╔══██╗██║ ██╔╝
██╔████╔██║██║   ██║███████║██║█████╗███████║██║  ██║█████╔╝
██║╚██╔╝██║██║   ██║██╔══██║██║╚════╝██╔══██║██║  ██║██╔═██╗
██║ ╚═╝ ██║╚██████╔╝██║  ██║██║      ██║  ██║██████╔╝██║  ██╗
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝      ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝
`

var (
	// BannerStyle for the ASCII art banner
	BannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTerraCotta)).
			Bold(true)

	// VersionStyle for version text
	VersionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))

	// SubtitleStyle for subtitle text
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))

	// WelcomeStyle for welcome message
	WelcomeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInfo)).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	// WizardInfoStyle for wizard info text
	WizardInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))
)

// PrintBanner prints the MoAI-ADK banner with version
func PrintBanner(version string) {
	fmt.Println(BannerStyle.Render(MoaiBanner))
	fmt.Println(SubtitleStyle.Render("  Modu-AI's Agentic Development Kit w/ SuperAgent 🎩 Alfred"))
	fmt.Println()
	fmt.Println(VersionStyle.Render(fmt.Sprintf("  Version: %s", version)))
	fmt.Println()
}

// PrintWelcomeMessage prints the welcome message
func PrintWelcomeMessage() {
	fmt.Println(WelcomeStyle.Render("🚀 Welcome to MoAI-ADK Project Initialization!"))
	fmt.Println()
	fmt.Println(WizardInfoStyle.Render("This wizard will guide you through setting up your MoAI-ADK project."))
	fmt.Println(WizardInfoStyle.Render("You can press Ctrl+C at any time to cancel."))
	fmt.Println()
}
