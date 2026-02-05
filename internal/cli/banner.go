package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Claude Code official terra cotta color
const claudeTerraCotta = "#DA7756"

// MoAI ASCII art banner
const moaiBanner = `
███╗   ███╗          █████╗ ██╗       █████╗ ██████╗ ██╗  ██╗
████╗ ████║ ██████╗ ██╔══██╗██║      ██╔══██╗██╔══██╗██║ ██╔╝
██╔████╔██║██║   ██║███████║██║█████╗███████║██║  ██║█████╔╝
██║╚██╔╝██║██║   ██║██╔══██║██║╚════╝██╔══██║██║  ██║██╔═██╗
██║ ╚═╝ ██║╚██████╔╝██║  ██║██║      ██║  ██║██████╔╝██║  ██╗
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝      ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝
`

// PrintBanner displays the MoAI ASCII art banner with version information.
// The banner uses MoAI's brand color (claude terra cotta #DA7756) and includes
// the provided version string. If version is empty, it displays "unknown".
func PrintBanner(version string) {
	// Create a style with terra cotta color
	bannerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeTerraCotta))
	dimStyle := lipgloss.NewStyle().Faint(true)

	// Print the ASCII art banner
	fmt.Println(bannerStyle.Render(moaiBanner))

	// Print description
	fmt.Println(dimStyle.Render("  Modu-AI's Agentic Development Kit w/ SuperAgent MoAI"))
	fmt.Println()

	// Print version
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Version: %s", version)))
	fmt.Println()
}

// PrintWelcomeMessage displays a friendly welcome message for new users.
// It provides basic usage instructions and reminds users they can exit anytime
// with Ctrl+C.
func PrintWelcomeMessage() {
	cyanBoldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true)
	dimStyle := lipgloss.NewStyle().Faint(true)

	// Print welcome title
	fmt.Println(cyanBoldStyle.Render("Welcome to MoAI-ADK Project Initialization!"))
	fmt.Println()

	// Print guide message
	fmt.Println(dimStyle.Render("This wizard will guide you through setting up your MoAI-ADK project."))
	fmt.Println(dimStyle.Render("You can press Ctrl+C at any time to cancel."))
	fmt.Println()
}
