package cli

// navigator-tiers CLI: BAS 4-tier overlay step for /moai project
// (SPEC-NAVIGATOR-SYNC-003, M4.6). Sibling to M0's navigator-sync Hidden
// subcommand. Reads M0's nav-graph.json as a READ-ONLY consumer (plus the
// blueprint/contracts/ADR/astx surfaces) and emits the additive
// .moai/project/navigator/tiers.json overlay. The two artifacts are JOINed
// by consumers on the composite key (entity_type, identifier); tiers.json
// NEVER overwrites nav-graph.json (REQ-NS3-018).
//
// Fail-open: exit 0 always. Every error mode (nav-graph absent, astx error,
// narrative file absent, per-tier timeout, write failure) logs ONE line to
// .moai/logs/navigator-sync.log and is swallowed — the calling /moai project
// step never aborts (REQ-NS3-020). Provenance is git-sourced
// (rev-parse HEAD + committer-date); NO wall-clock is used in the deterministic
// emission path, so two runs on the same HEAD produce byte-identical output
// (REQ-NS3-019).
//
// Wiring: this subcommand is the M4 sibling invoked AFTER M0's navigator-sync
// during /moai project (plan.md §F M4.6). It mirrors navigator_sync.go's
// shape verbatim — same Hidden flag, same fail-open RunE contract, same
// project-root resolution — so the orchestrator-side sequencing treats them
// as a uniform pair.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/navigator/tiers"
)

// newNavigatorTiersCmd creates the navigator-tiers subcommand. Hidden from
// the top-level `moai --help` surface (mirrors navigator-sync + navigator-enrich).
func newNavigatorTiersCmd() *cobra.Command {
	var projectRoot string

	cmd := &cobra.Command{
		Use:    "navigator-tiers",
		Short:  "Emit the 4-tier tiers.json overlay (BAS M4 sibling of navigator-sync)",
		Hidden: true,
		Long: `BAS 4-tier overlay step (sibling of navigator-sync).

Reads M0's nav-graph.json (READ-ONLY) plus the blueprint/contracts/ADR/astx
surfaces and emits .moai/project/navigator/tiers.json — an additive overlay
that consumers JOIN with nav-graph.json on (entity_type, identifier). Never
overwrites nav-graph.json (REQ-NS3-018).

Fail-open: exit 0 always. Every error mode logs one diagnostic line to
.moai/logs/navigator-sync.log and is swallowed (REQ-NS3-020). Provenance is
git-sourced (no wall-clock) — two runs on the same HEAD are byte-identical
(REQ-NS3-019).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigatorTiers(projectRoot)
		},
	}

	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"project root (defaults to $CLAUDE_PROJECT_DIR then $PWD)")

	return cmd
}

// runNavigatorTiers is the fail-open core. It never returns a non-nil error
// to the cobra RunE so /moai project never aborts on the tier step
// (REQ-NS3-020). Every error path logs one diagnostic line to
// .moai/logs/navigator-sync.log and yields exit 0.
func runNavigatorTiers(projectRoot string) error {
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
	root, _ = filepath.Abs(root)

	if err := tiers.Enrich(root); err != nil {
		// tiers.Enrich is itself fail-open and returns nil on every internal
		// error; this branch is the belt-and-suspenders guard so a future
		// contract change can never leak a non-nil return to the calling
		// /moai project step (REQ-NS3-020).
		fmt.Fprintf(os.Stderr, "navigator-tiers: %v\n", err)
	}
	return nil
}
