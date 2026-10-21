package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// codexCmd — the launcher-family sibling of cc/glm/cg (same group, same
// DisableFlagParsing discipline: --spawn and -- passthrough are handled by
// the launcher itself). SilenceErrors/SilenceUsage keep every diagnostic
// byte-identical to the named constants above (cobra would otherwise prefix
// "Error: " and append the usage block, breaking the exact-match cells).
var codexCmd = &cobra.Command{
	Use:   "codex [cli | status | app]",
	Short: "Codex launcher: launch the Codex CLI, or print the readiness readout",
	Long: "Launch Codex, or ask what is installed and wired.\n" +
		"\n" +
		"Called with no verb, this launches the Codex CLI at the project root.\n" +
		"The readiness readout moved to an explicit alias: moai codex status\n" +
		"reports six rows and starts nothing - the codex binary, CODEX_HOME,\n" +
		"the auth provider, the project wiring, the generated agent TOMLs, and\n" +
		"the harness entry. An incomplete wiring row is informational, not an\n" +
		"error: moai init --llm gpt generates the .codex wiring files.\n" +
		"Common local guidance shared with Claude, then Codex-specific local\n" +
		"guidance: both non-empty project-root files are injected as\n" +
		"developer instructions, in that order.\n" +
		"\n" +
		"  moai codex            launch the Codex CLI at the project root\n" +
		"  moai codex cli        the same launch, named explicitly\n" +
		"  moai codex status     print the readiness readout (starts nothing)\n" +
		"  moai codex app        launch the Codex desktop app (codex app)\n" +
		"  -w <worktree>         launch in an existing worktree (never creates one)\n" +
		"  -l, --lane            join the active factory as the next lane (CLI only)\n" +
		"  Factory sessions show launch_pending until their first prompt binds a session UUID.\n" +
		"  --spawn               open the launch in a new tmux window\n" +
		"  -d, --debug           launcher debug: trace the pre-exec path to stderr\n" +
		"                        (launcher-owned; the token never reaches codex)\n" +
		"  -- <codex-args...>    arguments after -- pass to codex verbatim",
	Example: "  # Launch the Codex CLI here\n" +
		"  moai codex\n" +
		"\n" +
		"  # Show what is installed, resolved, and wired (starts nothing)\n" +
		"  moai codex status\n" +
		"\n" +
		"  # Launch, passing arguments through verbatim\n" +
		"  moai codex -- --model o3\n" +
		"\n" +
		"  # Enter an existing worktree\n" +
		"  moai codex -w feat-login\n" +
		"\n" +
		"  # Launch the desktop app in a new tmux window\n" +
		"  moai codex app --spawn",
	GroupID:            "launch",
	DisableFlagParsing: true,
	SilenceErrors:      true,
	SilenceUsage:       true,
	RunE:               runCodex,
}

// runCodex routes the invocation: --help first (DisableFlagParsing means
// cobra never intercepts it), then --spawn stripping, then the closed-set
// verb lookup, then the readout or launch path.
func runCodex(cmd *cobra.Command, args []string) error {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return cmd.Help()
		}
		if a == "--" {
			break
		}
	}

	args, spawn := stripSpawnFlag(args)
	head, tail, hasTail := splitCodexDashDash(args)
	// The debug tokens are launcher-owned on this launcher (SPEC-CODEX-DEBUG-
	// MODE-001 REQ-001): the codex CLI rejects them, so the head's tokens are
	// stripped here — never forwarded — and activate the launcher debug trace
	// downstream. Tokens after -- are codex's own and were never inspected
	// (REQ-002).
	head, debugRequested := stripCodexDebugFlag(head)
	// The factory entries classify before anything else is read or written
	// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-003/-004, narrowing
	// REQ-CFR-001..005, SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-001/002/006): `-l` /
	// `--lane` routes to the per-card relaunch, every other factory shape
	// prints its one refusal line. Tokens after -- are codex's own.
	entry, diag := codexFactoryEntryClassify(head)
	if diag != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diag)
		return &exitCodeError{code: 1}
	}
	head, factoryEntry, err := parseCodexFactoryEntry(head)
	if err != nil {
		return err
	}
	// -w is consumed before the verb lookup so its tokens can never be
	// mistaken for a verb, and so the verb position keeps its one-token shape.
	head, worktree := stripCodexWorktreeFlag(head)
	if entry == codexFactoryEntryLane {
		// The relaunch stays the parent and selects each card's worktree
		// itself: no --spawn, no -w, and no verb or passthrough composes
		// with it — the closed-set discipline refuses the combination
		// rather than adapting it. The debug token (already stripped above)
		// composes: the relaunch loop traces its own pre-exec steps under
		// debug mode.
		if spawn || worktree.present || hasTail || len(stripCodexFactoryTokens(head)) > 0 {
			return codexUsageFailure(cmd)
		}
		return runCodexFactoryLane(cmd, factoryEntry, debugRequested)
	}
	if len(head) > 1 {
		return codexUsageFailure(cmd)
	}
	verb := ""
	if len(head) == 1 {
		verb = head[0]
	}
	kind, ok := codexVerbRouting[verb]
	if !ok {
		return codexUsageFailure(cmd)
	}
	if factoryEntry.Enabled && kind == codexVerbLaunchApp {
		return fmt.Errorf("-f/--factory requires the Codex CLI launch")
	}

	if kind.launches() {
		return runCodexLaunch(cmd, kind, tail, spawn, worktree, factoryEntry, debugRequested)
	}
	if worktree.present || factoryEntry.Enabled {
		// A readout starts no process, so it has no working directory to
		// point anywhere.
		return codexUsageFailure(cmd)
	}
	if spawn {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexSpawnReadoutDiag)
		return &exitCodeError{code: 1}
	}
	if debugRequested {
		// REQ-003: the debug token on a readout refuses with the named
		// diagnostic — a readout starts nothing there is nothing to trace.
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexDebugReadoutDiag)
		return &exitCodeError{code: 1}
	}
	if hasTail {
		return codexUsageFailure(cmd)
	}
	return runCodexReadout(cmd)
}
