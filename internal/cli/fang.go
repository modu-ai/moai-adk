package cli

import (
	"context"
	"errors"
	"image/color"
	"io"
	"os"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// runFang runs the root command through charm.land/fang/v2 (SPEC-CLI-TUX-V3-001
// M1c, REQ-CTX-018): styled help, styled errors, --version, and completions.
//
// fang.Execute returns the original error from ExecuteContext (fang.go:176), so
// the ExitCoder chain consumed by cmd/moai/main.go — the `moai worktree verify`
// 0/1/2/3 custom exit codes — is preserved verbatim (REQ-CTX-020,
// TestFangExitCoderCharacterization). The trivial fast-path lazy-init branch
// stays in Execute(), OUTSIDE this fang wrapper (REQ-CTX-019, design.md §E).
//
// fang mutates the passed command (help func, silence flags, version). In
// production Execute() runs once and the process exits, so the mutation is
// harmless; snapshot+restore keeps the shared global rootCmd pristine for the
// in-process cli test suite (the tui-based renderRootHelp set at package init
// stays the effective help func for the ~20 sibling tests that call
// rootCmd.Execute() directly).
// isRootHelpArgs reports whether the CLI args indicate the explicit root-help
// surface (moai --help / moai -h / moai help) that should carry the restored
// large logo BEFORE fang renders its help body (REQ-TUXIU-055, AC-TUXIU-024).
// It inspects ONLY os.Args[1:] shape.
//
// The three matched tokens (--help / -h / help) are a subset of the root.go
// trivialCommands token map. The empty arg vector MUST NOT match: no-args `moai`
// already prints the logo via rootCmd.Run→PrintBanner, so matching [] here would
// double-print the logo on the most-visible surface (§A.1 L4 HARD invariant).
// Subcommand-help shapes (`moai help <sub>`, `moai <sub> --help`) MUST NOT match
// either — subcommand help is logo-free by design.
func isRootHelpArgs(args []string) bool {
	if len(args) == 0 {
		return false // no-args is covered by PrintBanner in rootCmd.Run
	}
	switch args[0] {
	case "--help", "-h":
		return true
	case "help":
		return len(args) == 1 // "moai help" only; "moai help <sub>" is subcommand help
	default:
		return false
	}
}

func runFang(ctx context.Context, cmd *cobra.Command) error {
	origHelp := cmd.HelpFunc()
	origSilenceUsage := cmd.SilenceUsage
	origSilenceErrors := cmd.SilenceErrors
	origVersion := cmd.Version
	defer func() {
		cmd.SetHelpFunc(origHelp)
		cmd.SilenceUsage = origSilenceUsage
		cmd.SilenceErrors = origSilenceErrors
		cmd.Version = origVersion
	}()
	// Explicit root-help surface (moai --help / -h / help): print the restored
	// large logo through the same Printer/stdout gateway as PrintBanner, ABOVE
	// fang's styled help body (REQ-TUXIU-055). fang v2 has no header hook, so the
	// logo is emitted here before fang.Execute — no fang change / no go.mod change.
	if isRootHelpArgs(os.Args[1:]) {
		uikit.PrintLogo()
	}
	return fang.Execute(ctx, cmd, fangOptions()...)
}

// fangOptions is the M1c fang configuration:
//   - WithVersion pins fang's version surface to the single existing source
//     (pkg/version), avoiding a duplicate --version implementation
//     (acceptance.md §C "fang --version 자동화 vs 기존 version.GetVersion()").
//   - WithoutManpage keeps the current behavior (no `man` subcommand), and also
//     avoids a duplicate `man` add across repeated in-process Execute() calls.
//   - WithColorSchemeFunc feeds fang's palette from internal/tui tokens only
//     (REQ-CTX-022 — no hex literal outside internal/tui).
//   - WithErrorHandler suppresses the styled error box for ExitCoder carriers
//     (see moaiErrorHandler), preserving their SilenceErrors intent.
func fangOptions() []fang.Option {
	return []fang.Option{
		fang.WithVersion(version.GetVersion()),
		fang.WithoutManpage(),
		fang.WithColorSchemeFunc(moaiColorScheme),
		fang.WithErrorHandler(moaiErrorHandler),
	}
}

// moaiColorScheme derives fang's help/error palette entirely from internal/tui
// design tokens (REQ-CTX-022) — no hex literal lives outside internal/tui. The
// lipgloss.LightDarkFunc picks the light or dark token per the terminal's
// detected background, matching fang's own DefaultColorScheme pattern. Under
// NO_COLOR / non-TTY, fang's colorprofile writer strips the resulting ANSI, so
// AC-CTX-021 holds regardless of the tokens chosen here.
func moaiColorScheme(c lipgloss.LightDarkFunc) fang.ColorScheme {
	light := tui.LightTheme()
	dark := tui.DarkTheme()
	pick := func(lightTok, darkTok string) color.Color {
		return c(lipgloss.Color(lightTok), lipgloss.Color(darkTok))
	}
	return fang.ColorScheme{
		Base:           pick(light.Fg, dark.Fg),
		Title:          pick(light.Accent, dark.Accent),
		Description:    pick(light.Body, dark.Body),
		Codeblock:      pick(light.Panel, dark.Panel),
		Program:        pick(light.Info, dark.Info),
		DimmedArgument: pick(light.Dim, dark.Dim),
		Comment:        pick(light.Faint, dark.Faint),
		Flag:           pick(light.Success, dark.Success),
		FlagDefault:    pick(light.Dim, dark.Dim),
		Command:        pick(light.AccentDeep, dark.AccentDeep),
		QuotedString:   pick(light.Info, dark.Info),
		Argument:       pick(light.Body, dark.Body),
		Help:           pick(light.Dim, dark.Dim),
		Dash:           pick(light.Faint, dark.Faint),
		// ErrorHeader is {fg, bg}: a bright foreground on the danger background.
		ErrorHeader:  [2]color.Color{lipgloss.Color(dark.Fg), lipgloss.Color(light.Danger)},
		ErrorDetails: pick(light.Danger, dark.Danger),
	}
}

// exitCoder mirrors the cmd/moai/main.go ExitCoder interface: an error carrying
// a custom process exit code (the `moai worktree verify` 0/1/2/3 chain).
type exitCoder interface{ ExitCode() int }

// moaiErrorHandler styles genuine errors through fang's default renderer, but
// stays silent for ExitCoder carriers. Those are not user-facing failures — the
// subcommand already emitted its structured output and set SilenceErrors; fang
// would otherwise print a spurious styled "ERROR" box for a clean 0/1/2/3 exit
// (fang's DefaultErrorHandler runs unconditionally, ignoring SilenceErrors).
func moaiErrorHandler(w io.Writer, styles fang.Styles, err error) {
	var ec exitCoder
	if errors.As(err, &ec) {
		return
	}
	fang.DefaultErrorHandler(w, styles, err)
}
