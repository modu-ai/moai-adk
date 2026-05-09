package cli

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

// resolveTheme returns a Theme based on the NO_COLOR env and MOAI_THEME env.
// NO_COLOR=1 → MonochromeTheme; MOAI_THEME=dark → DarkTheme; otherwise LightTheme.
func resolveTheme() tui.Theme {
	if os.Getenv("NO_COLOR") == "1" {
		return tui.MonochromeTheme()
	}
	if strings.ToLower(os.Getenv("MOAI_THEME")) == "dark" {
		return tui.DarkTheme()
	}
	return tui.LightTheme()
}

// goVersion returns a short Go version string (e.g. "1.21.5" from "go1.21.5").
// MOAI_GO_VERSION_OVERRIDE env var allows pinning the value for deterministic
// test output across Go toolchain versions (CI vs local).
func goVersion() string {
	if v := os.Getenv("MOAI_GO_VERSION_OVERRIDE"); v != "" {
		return v
	}
	v := runtime.Version()
	return strings.TrimPrefix(v, "go")
}

// claudeVersion returns the CLAUDE_CODE_VERSION env var, or "claude" if unset.
func claudeVersion() string {
	if v := os.Getenv("CLAUDE_CODE_VERSION"); v != "" {
		return v
	}
	return "claude"
}

// gitVersionOverride returns the MOAI_GIT_VERSION_OVERRIDE env var (empty if unset).
// When set, doctor.go's checkGit uses this string as the message instead of
// executing `git --version`. Lesson NEW (M4-S4a) extension for CI portability:
// ubuntu-latest runner ships git 2.53.x while macOS runners ship Apple Git 2.50.x,
// causing golden test drift on otherwise-identical doctor output.
func gitVersionOverride() string {
	return os.Getenv("MOAI_GIT_VERSION_OVERRIDE")
}

// ghVersionOverride returns the MOAI_GH_VERSION_OVERRIDE env var (empty if unset).
// When set, doctor.go's checkGitHubCLI uses this string as the message instead
// of executing `gh --version`. Lesson NEW (M4-S4a) extension for CI portability:
// ubuntu-latest ships newer gh releases than macOS-latest, causing drift.
func ghVersionOverride() string {
	return os.Getenv("MOAI_GH_VERSION_OVERRIDE")
}

// goosArch returns the platform string used in doctor output ("goos/goarch").
// MOAI_GOOS_OVERRIDE / MOAI_GOARCH_OVERRIDE env vars allow pinning the values
// for deterministic test output across CI runners (linux/amd64 vs darwin/arm64
// vs windows/amd64). Lesson NEW (M4-S4a) extension for CI portability.
func goosArch() string {
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

// @MX:NOTE: [AUTO] CLI 배너 출력 — root/init/update/version 4+ entry point에서 호출됨
// PrintBanner displays the MoAI ASCII art banner with version information.
// The banner uses MoAI's deep teal accent colour from internal/tui Theme.Accent
// and includes the provided version string. If version is empty, it displays "".
// Three tui.Pill badges are rendered below the banner: version, go version, claude version.
func PrintBanner(version string) {
	th := resolveTheme()
	bannerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent))
	dimStyle := lipgloss.NewStyle().Faint(true)

	// Print the ASCII art banner
	fmt.Println(bannerStyle.Render(moaiBanner))

	// Print description
	fmt.Println(dimStyle.Render("  Modu-AI's Agentic Development Kit w/ SuperAgent MoAI"))
	fmt.Println()

	// Print version
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Version: %s", version)))
	fmt.Println()

	// Pill row: version (primary solid), go version (ok outline), claude version (info outline)
	// Design source: screens.jsx:180-182 (ScreenBanner)
	p1 := tui.Pill(tui.PillOpts{Kind: tui.PillPrimary, Solid: true, Label: fmt.Sprintf("v%s", version), Theme: &th})
	p2 := tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("go %s", goVersion()), Theme: &th})
	p3 := tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Solid: false, Label: claudeVersion(), Theme: &th})
	pillRow := lipgloss.JoinHorizontal(lipgloss.Top, p1, " ", p2, " ", p3)
	fmt.Println("  " + pillRow)
	fmt.Println()
}

// PrintWelcomeMessage displays a friendly welcome message for new users.
// It provides basic usage instructions and reminds users they can exit anytime
// with Ctrl+C. The title uses MoAI's deep teal accent colour from internal/tui Theme.Accent.
func PrintWelcomeMessage() {
	th := resolveTheme()
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Accent)).
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
