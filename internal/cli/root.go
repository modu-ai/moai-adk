package cli

import (
	"fmt"
	"io"
	"log/slog"
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
var trivialCommands = map[string]bool{
	"--version": true,
	"version":   true,
	"-v":        true,
	"help":      true,
	"--help":    true,
	"-h":        true,
	"completion": true, // cobra built-in
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
	if isTrivialCommand(os.Args[1:]) {
		initLightDeps()
	} else {
		InitDependencies()
	}
	return rootCmd.Execute()
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

// initLightDeps initializes only the lightweight logger for trivial subcommands.
// This avoids the full InitDependencies() cost (hook registry, LSP config load,
// security scanner creation, astgrep analyzer) for commands like --version.
func initLightDeps() {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	slog.SetDefault(logger)
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

	// SPEC-TELEMETRY-001: register telemetry subcommand
	rootCmd.AddCommand(telemetryCmd)

	// SPEC-V3R2-CON-001: register constitution subcommand
	rootCmd.AddCommand(newConstitutionCmd())

	// SPEC-V3R2-RT-004: register state subcommand
	rootCmd.AddCommand(newStateCmd())

	// SPEC-V3R2-RT-004 REQ-031: register clean subcommand
	rootCmd.AddCommand(newCleanCmd())

	// SPEC-V3R2-RT-007: register migration subcommand group
	rootCmd.AddCommand(migrationCmd)

	// NOTE: newHarnessCmd is intentionally NOT registered per SPEC-V3R4-HARNESS-001
	// (BC-V3R4-HARNESS-001-CLI-RETIREMENT). The harness lifecycle verbs
	// (status / apply / rollback / disable) are retired and owned by the
	// /moai:harness slash command + skill workflow surface (no Go binary invocation).
	// See internal/cli/harness_retirement_test.go for the CI guard test.

	// SPEC-V3R2-HRN-001: Register the NEW harness routing command (route + validate verbs).
	// This is DISTINCT from the retired newHarnessCmd() factory.
	// The CI guard in harness_retirement_test.go allows 'route' and 'validate' verbs
	// but continues to block the retired lifecycle verbs (status/apply/rollback/disable).
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
}
