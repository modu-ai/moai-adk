// memory_drain.go — `moai memory drain` (SPEC-AGENT-MEMORY-DRAIN-001).
//
// The one-shot backfill half of the agent-memory drain: enumerate every
// registered worktree, report its agent-memory content, and with --yes copy
// the topic files into the primary checkout's store under the shared
// reconciliation rules (internal/hook/agentmemory.go — copy never move,
// never overwrite, append exactly one index line per landed topic).
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/spf13/cobra"
)

// Function-variable seams for test injection. The real implementations are
// git-backed; tests point them at t.TempDir() fixtures so no real sibling
// tree is ever enumerated or written.
var (
	// memoryDrainPrimaryRoot resolves the primary checkout root for a
	// starting directory.
	memoryDrainPrimaryRoot = hook.PrimaryRootOf

	// memoryDrainListWorktrees enumerates the linked worktree roots of a
	// primary checkout, excluding the primary itself.
	memoryDrainListWorktrees = listWorktreesReal
)

// listWorktreesReal enumerates linked worktree roots via
// `git worktree list --porcelain`. The first porcelain stanza is the main
// worktree (the primary); every other entry is a linked tree.
func listWorktreesReal(primary string) ([]string, error) {
	worktrees, err := git.NewWorktreeManager(primary).List()
	if err != nil {
		return nil, fmt.Errorf("memory drain: list worktrees: %w", err)
	}
	cleanPrimary := filepath.Clean(primary)
	var out []string
	for _, wt := range worktrees {
		p := filepath.Clean(wt.Path)
		if p == cleanPrimary {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

// newMemoryDrainCmd — `moai memory drain [--yes] [--json] [--from DIR]`.
func newMemoryDrainCmd() *cobra.Command {
	var apply bool
	var jsonOutput bool
	var fromDir string

	cmd := &cobra.Command{
		Use:   "drain",
		Short: "Copy worktree agent-memory into the primary store (preview by default)",
		Long: `Copy worktree agent-memory into the primary store.

A worktree is its own project root and .claude/agent-memory/ is gitignored, so
memory written inside a worktree never reaches the primary checkout through
git. drain enumerates every registered worktree, reports the agent-memory
content it finds, and with --yes copies the topic files into the primary
store: never overwriting (collisions land as <name>.wt-<worktree>.md), never
deleting anything, appending one index line per copied topic.

Without --yes the command previews the copy set and writes nothing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := fromDir
			if start == "" {
				start = resolveProjectDir()
			}
			primary, _, err := memoryDrainPrimaryRoot(start)
			if err != nil {
				return fmt.Errorf("memory drain: %w", err)
			}
			trees, err := memoryDrainListWorktrees(primary)
			if err != nil {
				return fmt.Errorf("memory drain: %w", err)
			}

			records := make([]hook.TreeDrainRecord, 0, len(trees))
			cleanPrimary := filepath.Clean(primary)
			for _, tree := range trees {
				if filepath.Clean(tree) == cleanPrimary {
					// The primary checkout is never its own drain source.
					continue
				}
				rec, drainErr := hook.DrainTree(primary, tree, apply)
				if drainErr != nil {
					// One unreadable tree must not abort the backfill; the
					// record still names the tree so the gap is visible.
					drainPrintf(cmd.ErrOrStderr(), "warning: %s: %v\n", tree, drainErr)
				}
				records = append(records, rec)
			}

			if jsonOutput {
				return renderMemoryDrainJSON(cmd.OutOrStdout(), records)
			}
			renderMemoryDrainText(cmd.OutOrStdout(), primary, records, apply)
			return nil
		},
	}
	cmd.Flags().BoolVar(&apply, "yes", false, "Apply the copy (default is a write-nothing preview)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit per-tree records as a JSON array on stdout")
	cmd.Flags().StringVar(&fromDir, "from", "", "Resolve the repository from this directory instead of the project dir")
	return cmd
}

// renderMemoryDrainJSON emits the machine-readable per-tree records
// (REQ-AM-010).
func renderMemoryDrainJSON(out io.Writer, records []hook.TreeDrainRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("memory drain: marshal records: %w", err)
	}
	drainPrintf(out, "%s\n", string(data))
	return nil
}

// drainPrintf writes a report line. A write error to the command's output
// writer is not actionable from a renderer; the error is taken into a named
// value and deliberately discarded.
func drainPrintf(out io.Writer, format string, args ...any) {
	_, err := fmt.Fprintf(out, format, args...)
	_ = err
}

// renderMemoryDrainText prints the human report: the copy set per tree and
// a summary line.
func renderMemoryDrainText(out io.Writer, primary string, records []hook.TreeDrainRecord, apply bool) {
	drainPrintf(out, "primary store: %s\n", filepath.Join(primary, ".claude", "agent-memory"))
	var totals hook.TreeDrainRecord
	withContent := 0
	for _, rec := range records {
		mode := ""
		if !apply {
			mode = " [preview]"
		}
		drainPrintf(out, "\n%s (%d agent(s), %d topic file(s))%s\n",
			rec.Path, len(rec.Agents), rec.Files, mode)
		for _, a := range rec.Agents {
			indexNote := ""
			if a.HasIndex {
				indexNote = ", index present"
			}
			drainPrintf(out, "  %s: %d file(s)%s\n", a.Name, a.Files, indexNote)
		}
		for _, action := range rec.Actions {
			drainPrintf(out, "    %s\n", action)
		}
		if rec.Missing {
			drainPrintf(out, "    (tree or agent-memory store missing)\n")
		}
		if rec.Files > 0 {
			withContent++
		}
		totals.Files += rec.Files
		totals.Copied += rec.Copied
		totals.Collided += rec.Collided
		totals.Skipped += rec.Skipped
		totals.IndexLinesAdded += rec.IndexLinesAdded
	}
	if apply {
		drainPrintf(out, "\nsummary: %d tree(s) with content, %d copied, %d collided, %d skipped, %d index line(s) added\n",
			withContent, totals.Copied, totals.Collided, totals.Skipped, totals.IndexLinesAdded)
	} else {
		drainPrintf(out, "\nsummary: %d tree(s) with content, %d to copy, %d to collide, %d already present, %d index line(s) to add — preview, nothing written; pass --yes to apply\n",
			withContent, totals.Copied, totals.Collided, totals.Skipped, totals.IndexLinesAdded)
	}
}
