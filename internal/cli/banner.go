package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// MoAI ASCII art banner
const moaiBanner = `
███╗   ███╗          █████╗ ██╗       █████╗ ██████╗ ██╗  ██╗
████╗ ████║ ██████╗ ██╔══██╗██║      ██╔══██╗██╔══██╗██║ ██╔╝
██╔████╔██║██║   ██║███████║██║█████╗███████║██║  ██║█████╔╝
██║╚██╔╝██║██║   ██║██╔══██║██║╚════╝██╔══██║██║  ██║██╔═██╗
██║ ╚═╝ ██║╚██████╔╝██║  ██║██║      ██║  ██║██████╔╝██║  ██╗
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝      ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝
`

// @MX:NOTE: [AUTO] CLI banner output - called from 4+ entry points including init, update, and version
// PrintBanner displays the MoAI ASCII art banner with version information.
// The banner uses MoAI's adaptive brand color (#C45A3C light, #DA7756 dark)
// and includes the provided version string. If version is empty, it displays "unknown".
func PrintBanner(version string) {
	// Create a style with terra cotta color (adaptive for light/dark terminals)
	bannerStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C45A3C", Dark: "#DA7756"})
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
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#5B21B6", Dark: "#7C3AED"}).
		Bold(true)
	dimStyle := lipgloss.NewStyle().Faint(true)

	// Print welcome title
	fmt.Println(titleStyle.Render("Welcome to MoAI-ADK Project Initialization!"))
	fmt.Println()

	// Print guide message
	fmt.Println(dimStyle.Render("This wizard will guide you through setting up your MoAI-ADK project."))
	fmt.Println(dimStyle.Render("You can press Ctrl+C at any time to cancel."))
	fmt.Println()
}
