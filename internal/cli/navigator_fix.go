package cli

// navigator_fix CLI: BAS Fix layer-1 entry point (SPEC-NAVIGATOR-SYNC-005,
// M3.2). Sibling to M0's navigator-sync, M2's navigator-route, and M4's
// navigator-tiers Hidden subcommands. On-demand ONLY — it is the deterministic
// draft-request producer (layer 1). It does NOT generate the AI draft (layer 2,
// orchestrator-mediated via manager-develop) and does NOT apply it (layer 3,
// post-approval).
//
// The CLI loads the four read-only inputs (work-items.json + detect JSONL +
// git-diff + nav-graph.json), computes the diff-scope, and emits
// .moai/project/navigator/fix-drafts/<draft-id>/request.json plus a stdout JSON
// signal ({"draft_request_path":..., "status":..., "draft_id":...}) — the
// design.md §A.4 handoff contract the orchestrator consumes to spawn the
// AI-draft delegation.
//
// On-demand ONLY (REQ-NS5-001 / AC-NS5-001b): there is NO PostToolUse hook and
// NO handle-navigator-fix.sh wrapper. The sole entry point is this Hidden cobra
// subcommand. Fail-open (REQ-NS5-009): exit 0 always. Provenance is git-sourced
// (no wall-clock) so two runs on the same HEAD are byte-identical.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/navigator/fix"
)

// newNavigatorFixCmd creates the navigator-fix subcommand. Hidden from the
// top-level `moai --help` surface (mirrors navigator-sync / navigator-route /
// navigator-tiers).
func newNavigatorFixCmd() *cobra.Command {
	var (
		projectRoot string
		compareTo   string
	)

	cmd := &cobra.Command{
		Use:   "navigator-fix",
		Short: "Produce the draft-request manifest for stale doc subtrees",
		Hidden: true,
		Long: `BAS Fix layer-1 producer (sibling of navigator-sync/route/tiers).

Loads the four read-only inputs (M2 work-items.json + M1 detect JSONL +
git-diff + M0 nav-graph.json), computes the stale doc-subtree diff-scope, and
emits .moai/project/navigator/fix-drafts/<draft-id>/request.json plus a stdout
JSON signal — the handoff contract the orchestrator consumes to spawn the
AI-draft delegation (layer 2). It does NOT generate the draft and does NOT
apply it.

On-demand ONLY (REQ-NS5-001): no PostToolUse hook, no handle wrapper. Fail-open
(REQ-NS5-009): exit 0 always. Provenance is git-sourced (no wall-clock) — two
runs on the same HEAD are byte-identical.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigatorFix(projectRoot, compareTo)
		},
	}

	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"project root (defaults to $CLAUDE_PROJECT_DIR then $PWD)")
	cmd.Flags().StringVar(&compareTo, "compare-to", "",
		"baseline commit override (default: nav-graph provenance, then HEAD~1)")

	return cmd
}

// runNavigatorFix is the fail-open core. It never returns a non-nil error to
// the cobra RunE (REQ-NS5-009): every failure mode degrades inside fix.Run to an
// exit-0 stdout signal. The Result signal is printed to stdout so the
// orchestrator consumes the §A.4 handoff contract.
func runNavigatorFix(projectRoot, compareTo string) error {
	root := projectRoot
	if root == "" {
		root = os.Getenv("CLAUDE_PROJECT_DIR")
	}
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			root = "."
		}
	}
	if absRoot, err := filepath.Abs(root); err == nil {
		root = absRoot
	}

	res := fix.Run(fix.Options{ProjectRoot: root, CompareTo: compareTo})
	if line, err := res.SignalJSON(); err == nil {
		_, _ = fmt.Fprint(os.Stdout, string(line))
	}
	return nil
}
