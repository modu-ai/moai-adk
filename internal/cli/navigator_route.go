package cli

// navigator_route CLI: BAS Route step for /moai project (SPEC-NAVIGATOR-
// SYNC-004, M2). Sibling to M0's navigator-sync and M4's navigator-tiers
// Hidden subcommands. Reads the 002 audit-report.json, M1 detect JSONL,
// and M0 nav-graph.json as READ-ONLY consumer inputs and emits
// .moai/project/navigator/work-items.{md,json} — actionable work items
// with owners bound to code paths (never persons).
//
// Fail-open: exit 0 always. Every error mode (audit absent, detect absent,
// nav-graph absent, unparseable JSON, schema-invalid, owner-resolution
// error, all-inputs-absent, timeout) logs at most one diagnostic line to
// .moai/logs/navigator-sync.log and is swallowed — the calling /moai
// project step never aborts (REQ-NS4-009). Provenance is git-sourced
// (rev-parse HEAD + committer-date); NO wall-clock is used, so two runs
// on the same HEAD produce byte-identical output (REQ-NS4-008).

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/navigator/route"
)

// newNavigatorRouteCmd creates the Route subcommand. Hidden from the
// top-level `moai --help` surface (mirrors navigator-sync +
// navigator-tiers).
func newNavigatorRouteCmd() *cobra.Command {
	var projectRoot string

	cmd := &cobra.Command{
		Use:    "navigator-route",
		Short:  "Promote audit + detect findings into owner-bound work items",
		Hidden: true,
		Long: `BAS Route step (sibling of navigator-sync + navigator-tiers).

Reads 002's audit-report.json (advisory), M1's detect JSONL state, and M0's
nav-graph.json (READ-ONLY) and promotes missing/orphan/detect findings into
actionable work items at .moai/project/navigator/work-items.{md,json}. Each
work item's owner is bound to a code path or design-doc path — never a
person (falconer binding, REQ-NS4-004).

Fail-open: exit 0 always. Every error mode logs one diagnostic line to
.moai/logs/navigator-sync.log and is swallowed (REQ-NS4-009). Provenance is
git-sourced (no wall-clock) — two runs on the same HEAD are byte-identical
(REQ-NS4-008).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNavigatorRoute(projectRoot)
		},
	}

	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"project root (defaults to $CLAUDE_PROJECT_DIR then $PWD)")

	return cmd
}

// runNavigatorRoute is the fail-open core. It never returns a non-nil error
// to the cobra RunE so /moai project never aborts on the Route step
// (REQ-NS4-009). Every error path is swallowed inside route.RunDefault.
func runNavigatorRoute(projectRoot string) error {
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

	if err := route.RunDefault(root); err != nil {
		// route.RunDefault is itself fail-open and returns nil on every
		// internal error; this branch is the belt-and-suspenders guard so
		// a future contract change can never leak a non-nil return to the
		// calling /moai project step.
		fmt.Fprintf(os.Stderr, "navigator-route: %v\n", err)
	}
	return nil
}
