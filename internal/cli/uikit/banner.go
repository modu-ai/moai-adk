package uikit

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/modu-ai/moai-adk/internal/tui"
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

// ResolveTheme returns a Theme appropriate for the current terminal.
//
// Delegates to tui.ResolveOS() which applies the canonical priority chain:
//  1. NO_COLOR set → MonochromeTheme
//  2. MOAI_THEME="light" → LightTheme
//  3. MOAI_THEME="dark"  → DarkTheme
//  4. MOAI_THEME="auto"/unset → lipgloss.HasDarkBackground() auto-detect
//  5. invalid MOAI_THEME → DarkTheme (safe default)
//
// Moved from internal/cli/banner.go (SPEC-CLI-UIKIT-KERNEL-001). Package cli
// retains a local resolveTheme() wrapper (cli/theme.go) so its 13 callers are
// unchanged.
func ResolveTheme() tui.Theme {
	return tui.ResolveOS()
}

// GoVersion returns a short Go version string (e.g. "1.21.5" from "go1.21.5").
// MOAI_GO_VERSION_OVERRIDE env var allows pinning for deterministic test output.
func GoVersion() string {
	if v := os.Getenv("MOAI_GO_VERSION_OVERRIDE"); v != "" {
		return v
	}
	v := runtime.Version()
	return strings.TrimPrefix(v, "go")
}

// ClaudeVersion returns the CLAUDE_CODE_VERSION env var, or "claude" if unset.
func ClaudeVersion() string {
	if v := os.Getenv("CLAUDE_CODE_VERSION"); v != "" {
		return v
	}
	return "claude"
}

// GitVersionOverride returns the MOAI_GIT_VERSION_OVERRIDE env var (empty if unset).
func GitVersionOverride() string {
	return os.Getenv("MOAI_GIT_VERSION_OVERRIDE")
}

// GhVersionOverride returns the MOAI_GH_VERSION_OVERRIDE env var (empty if unset).
func GhVersionOverride() string {
	return os.Getenv("MOAI_GH_VERSION_OVERRIDE")
}

// GoosArch returns the platform string used in doctor output ("goos/goarch").
func GoosArch() string {
	goos := runtime.GOOS
	if v := os.Getenv("MOAI_GOOS_OVERRIDE"); v != "" {
		goos = v
	}
	goarch := runtime.GOARCH
	if v := os.Getenv("MOAI_GOARCH_OVERRIDE"); v != "" {
		goarch = v
	}
	return goos + "/" + goarch
}

// @MX:NOTE: [AUTO] CLI banner output — called from root/init/update/version entry points
// PrintBanner displays the MoAI ASCII art banner with version information.
func PrintBanner(version string) {
	th := ResolveTheme()
	bannerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent))
	// Secondary text uses the theme Body colour (readable, subordinate to the
	// coral accent) rather than the ANSI Faint attribute, which renders at
	// reduced intensity and is illegible on many dark terminal backgrounds.
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Body))

	fmt.Println(bannerStyle.Render(moaiBanner))
	fmt.Println(dimStyle.Render("  Modu-AI's Agentic Development Kit w/ SuperAgent MoAI"))
	fmt.Println()
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Version: %s", version)))
	fmt.Println()

	p1 := tui.Pill(tui.PillOpts{Kind: tui.PillPrimary, Solid: true, Label: fmt.Sprintf("v%s", version), Theme: &th})
	p2 := tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("go %s", GoVersion()), Theme: &th})
	p3 := tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Solid: false, Label: ClaudeVersion(), Theme: &th})
	pillRow := lipgloss.JoinHorizontal(lipgloss.Top, p1, " ", p2, " ", p3)
	fmt.Println("  " + pillRow)
	fmt.Println()
}

// PrintWelcomeMessage displays a friendly welcome message for new users.
func PrintWelcomeMessage() {
	th := ResolveTheme()
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Accent)).
		Bold(true)
	// Body colour instead of the ANSI Faint attribute — see PrintBanner.
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Body))

	fmt.Println(titleStyle.Render("Welcome to MoAI-ADK Project Initialization!"))
	fmt.Println()
	fmt.Println(dimStyle.Render("This wizard will guide you through setting up your MoAI-ADK project."))
	fmt.Println(dimStyle.Render("You can press Ctrl+C at any time to cancel."))
	fmt.Println()
}
