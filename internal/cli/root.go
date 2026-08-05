package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/agentlint"
	"github.com/modu-ai/moai-adk/internal/cli/preference"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "moai",
	Short: "MoAI-ADK: Agentic Development Kit for Claude Code",
	Long: `MoAI-ADK (Go Edition) is a high-performance development toolkit
that serves as the runtime backbone for the MoAI framework within Claude Code.

It provides CLI tooling, configuration management, LSP integration,
Git operations, quality gates, and autonomous development loop capabilities.

Use 'moai cc', 'moai cg', or 'moai glm' to launch Claude Code.`,
	Version: version.GetVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		uikit.PrintBanner(version.GetVersion())
		_ = cmd.Help()
	},
}

// trivialCommands lists subcommands/flags that are handled entirely by cobra's
// built-in machinery and do NOT access the `deps` global. These skip the full
// InitDependencies() call (REQ-PERF-003-A — CLI lazy initialization).
//
// R3 nil-access mitigation: only commands verified to NOT touch `deps` are listed
// here. Adding a command that accesses deps without InitDependencies would panic.
// The launch group (cc/cg/glm) is listed for a second reason: those commands
// end in syscall.Exec, replacing this process with the claude binary, so the
// whole graph would be assembled and immediately discarded. The only deps read
// on that path is loadGLMConfig's nil-safe branch (glm.go), which falls back to
// reading llm.yaml from disk — its documented live runtime path.
var trivialCommands = map[string]bool{
	"--version":  true,
	"version":    true,
	"-v":         true,
	"help":       true,
	"--help":     true,
	"-h":         true,
	"completion": true, // cobra built-in
	"cc":         true, // launcher: exec's claude, discards the graph
	"cg":         true, // launcher: exec's claude, discards the graph
	"glm":        true, // launcher: exec's claude, discards the graph
}

// @MX:ANCHOR: [AUTO] Execute is the main entry point for the moai CLI
// @MX:REASON: [AUTO] fan_in=3, called from cmd/moai/main.go, root_test.go, integration_test.go
// Execute initializes dependencies and runs the root command.
//
// REQ-PERF-003-A: trivial subcommands (--version, help, completion) skip the
// full dependency graph initialization. Heavy components (hook registry, LSP,
// security scanner, astgrep) are not initialized for these commands.
// REQ-PERF-003-B: all other subcommands (hook, spec, update, etc.) receive
// the full dependency graph via InitDependencies() — handler completeness preserved.
func Execute() error {
	initConsole()
	args := os.Args[1:]
	// Logging is configured here, for every subcommand and ahead of the branch
	// below, so that both paths share one decision. configureLogging is the only
	// place the CLI installs the default logger; InitDependencies deliberately
	// does not install one of its own.
	configureLogging(args)
	if !isTrivialCommand(args) {
		InitDependencies()
	}
	// SPEC-CLI-TUX-V3-004 M4d (REQ-TUX4-007, keep-fang verdict): order each
	// help group's rows by usage frequency before fang renders Commands().
	reorderRootHelpCommands(rootCmd)
	return executeRoot(context.Background(), rootCmd)
}

// executeRoot is the single root-execution seam. M1a/M1b ran the raw cobra path;
// SPEC-CLI-TUX-V3-001 M1c routes it through charm.land/fang/v2 (styled help,
// errors, --version, completions) while preserving the ExitCoder error chain
// consumed by cmd/moai/main.go and the trivial fast-path lazy-init in Execute().
// See fang.go for the fang wiring and TestFangExitCoderCharacterization for the
// exit-code characterization that must hold across the swap.
func executeRoot(ctx context.Context, cmd *cobra.Command) error {
	return runFang(ctx, cmd)
}

// isTrivialCommand checks whether the CLI args indicate a trivial subcommand
// that does not require the full dependency graph.
func isTrivialCommand(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if trivialCommands[arg] {
				return true
			}
			continue
		}
		// First non-flag arg is the subcommand
		return trivialCommands[arg]
	}
	return false
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("moai-adk %s\n", version.GetVersion()))

	// M6-S5: Install tui-based help renderer for rootCmd.
	// SetHelpFunc applies to "moai --help" and "moai help" (cobra's built-in help subcommand).
	// Subcommand-level help (e.g. "moai doctor --help") still uses cobra's default template.
	rootCmd.SetHelpFunc(renderRootHelp)

	// Command groups
	rootCmd.AddGroup(
		&cobra.Group{ID: "launch", Title: "Launch Commands:"},
		&cobra.Group{ID: "project", Title: "Project Commands:"},
		&cobra.Group{ID: "tools", Title: "Tools:"},
	)

	// Wire worktree subcommand with lazy Git initialization
	worktree.WorktreeCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if deps == nil {
			return fmt.Errorf("dependencies not initialized")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		if err := deps.EnsureGit(cwd); err != nil {
			return fmt.Errorf("initialize git: %w", err)
		}
		worktree.WorktreeProvider = deps.GitWorktree
		return nil
	}

	// Register worktree subcommand tree
	rootCmd.AddCommand(worktree.WorktreeCmd)

	// SPEC-CLI-SUBPKG-SPLIT-001 M1: register agentlint subpackage. The "agent"
	// and "workflow" parent commands are exported by the subpackage; their lint
	// subcommands are wired on in package init().
	rootCmd.AddCommand(agentlint.AgentCmd)
	rootCmd.AddCommand(agentlint.WorkflowCmd)

	// Register statusline command
	rootCmd.AddCommand(StatuslineCmd)

	// ASTG-UPGRADE-001: register astgrep command
	rootCmd.AddCommand(NewAstGrepCmd())

	// SPEC-ASTGREP-EDIT-001: register ast-edit, the write-side counterpart.
	// Kept separate from ast-grep so a read-only permission grant stays read-only.
	rootCmd.AddCommand(NewAstEditCmd())

	// SPEC-TELEMETRY-001: register telemetry subcommand
	rootCmd.AddCommand(telemetryCmd)

	// SPEC-V3R2-CON-001: register constitution subcommand
	rootCmd.AddCommand(newConstitutionCmd())

	// SPEC-V3R2-RT-004: register state subcommand
	rootCmd.AddCommand(newStateCmd())

	// SPEC-V3R2-RT-004 REQ-031: register clean subcommand
	rootCmd.AddCommand(newCleanCmd())

	// SPEC-PROJECT-NAVIGATOR-003: AST enrichment entry point for /moai codemaps.
	rootCmd.AddCommand(newNavigatorEnrichCmd())

	// SPEC-NAVIGATOR-SYNC-001: BAS integration-layer join step for /moai project.
	rootCmd.AddCommand(newNavigatorSyncCmd())

	// SPEC-V3R2-RT-007: register migration subcommand group
	rootCmd.AddCommand(migrationCmd)

	// NOTE: newHarnessCmd is a superseded factory and is intentionally NOT
	// registered here. It remains compilable only as a deprecation marker
	// (see TestHarnessFactoryStillCompiles in harness_retirement_test.go).
	// The harness lifecycle verbs (status / apply / rollback / disable) were
	// retired by SPEC-V3R4-HARNESS-001 and later un-retired by
	// SPEC-V3R5-HARNESS-AUTONOMY-001: they are registered live below, under
	// the unified newHarnessRouterCmd() tree.

	// newHarnessRouterCmd() (below) is the live, unified registration site for
	// every harness verb: the routing verbs (route/validate), the un-retired
	// lifecycle verbs (status/apply/rollback/disable), the proposal-management
	// verbs (mute/mute-list/unmute/verify), and the v4 harness verbs
	// (list/edit/remove/doctor). This is DISTINCT from the superseded
	// newHarnessCmd() factory above, which is never added to the command tree.
	// See internal/cli/harness_retirement_test.go TestHarnessV3R5VerbSurface
	// for the CI guard enforcing this registration.
	rootCmd.AddCommand(newHarnessRouterCmd())

	// SPEC-V3R6-TOOL-POLICY-SSOT-001: register tool-policy subcommand
	// (build + list). The YAML at .moai/config/sections/tool-policy.yaml is the
	// SSOT from which the settings.json permissions block is generated.
	rootCmd.AddCommand(newToolPolicyCmd())

	// SPEC-DIVECC-INVENTORY-VIEW-001: register inventory subcommand — a
	// read-only unified view composing sessions / worktrees / harnesses.
	rootCmd.AddCommand(newInventoryCmd())

	// SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 M4: register the preference
	// subtree (parent + decay-scan child). M5 will add `toggle` as a sibling.
	rootCmd.AddCommand(preference.PreferenceCmd)

	// SPEC-MODEL-PROFILE-MATRIX-001 M2: register the read-only `moai model
	// profile` resolver — the per-agent model+effort profile injection surface.
	rootCmd.AddCommand(newModelCmd())
}
