package worktree

// Hoist — evidence retrieval before worktree disposal
// (SPEC-REPORTS-LIFECYCLE-001 REQ-RLC-005 / REQ-RLC-006).
//
// Card trees carry their run-phase evidence under <tree>/.moai/reports/
// (gitignored, machine-local). Disposing the tree destroys that evidence —
// the worktree is the only copy (AGENTS.md §3), so hoisting is the one
// retrieval step that must complete BEFORE any disposal. Card trees are L1
// (SPEC-WORKTREE-DONE-TIER-001: moai worktree done refuses them outright),
// so the exposure surface has two entries:
//  1. this independent verb — the only execution mechanism for L1 trees,
//     invoked by the session-end disposal flow (worktree-integration.md
//     hoist-before-dispose obligation); and
//  2. moai worktree done itself, which calls the same routine before
//     removing an L2 tree (done.go wiring).

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// HoistResult reports what one hoist pass did.
type HoistResult struct {
	Count   int64    // files copied
	Bytes   int64    // bytes copied
	Skipped []string // relative paths NOT copied: destination held different content
	Dest    string   // destination directory (main root / .moai/reports/worktrees/<tree>)
}

// hoistWorktreeReports copies treePath/.moai/reports/ content into
// mainRoot/.moai/reports/worktrees/<tree-name>/, preserving relative paths.
// Conflict policy (REQ-RLC-006): an existing destination file with different
// content is NEVER overwritten — its relative path is reported in Skipped
// and the copy proceeds for everything else. A tree without .moai/reports/
// is a no-op success (nothing to hoist is a completed retrieval).
//
// The source tree path is trusted here; the verb resolves and validates it
// before calling, and done passes a worktree path it just resolved from the
// provider.
//
// @MX:ANCHOR: [AUTO] hoist routine — the single evidence-retrieval core shared by the verb and done
// @MX:REASON: both the moai worktree hoist verb and the done L2 removal path copy evidence through this function; weakening its conflict policy or destination layout silently destroys card evidence on the next disposal
// @MX:SPEC: SPEC-REPORTS-LIFECYCLE-001
func hoistWorktreeReports(treePath, mainRoot string) (HoistResult, error) {
	res := HoistResult{
		Dest: filepath.Join(mainRoot, ".moai", "reports", "worktrees", filepath.Base(treePath)),
	}
	srcRoot := filepath.Join(treePath, ".moai", "reports")
	if _, err := os.Stat(srcRoot); err != nil {
		if os.IsNotExist(err) {
			return res, nil
		}
		return res, fmt.Errorf("stat %s: %w", srcRoot, err)
	}

	err := filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return fmt.Errorf("rel %s: %w", path, err)
		}
		dst := filepath.Join(res.Dest, rel)
		if existing, statErr := os.ReadFile(dst); statErr == nil {
			incoming, readErr := os.ReadFile(path)
			if readErr != nil {
				return fmt.Errorf("read %s: %w", path, readErr)
			}
			if string(existing) == string(incoming) {
				return nil // identical destination — nothing to do
			}
			res.Skipped = append(res.Skipped, rel)
			return nil // REQ-RLC-006: never silently overwrite
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", rel, err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
		res.Count++
		res.Bytes += info.Size()
		return nil
	})
	if err != nil {
		return res, fmt.Errorf("hoist %s: %w", treePath, err)
	}
	return res, nil
}

// hoistTargetMainRoot resolves the project root the hoist destination hangs
// off FROM THE TARGET tree path — the same target-derived resolution the
// done tier guard uses, so running the verb from inside another worktree
// lands evidence in the primary checkout, not the current tree.
func hoistTargetMainRoot(treePath string) (string, error) {
	return gitMainRootFromTargetFunc(treePath)
}

// isInsideRoot reports whether path is root itself or under it, after
// symlink-canonicalizing both sides (macOS /tmp vs /private/tmp).
func isInsideRoot(path, root string) bool {
	canonicalPath := canonicalTierPath(path)
	canonicalRoot := canonicalTierPath(root)
	if canonicalPath == canonicalRoot {
		return true
	}
	return strings.HasPrefix(canonicalPath, canonicalRoot+string(os.PathSeparator))
}

func newHoistCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hoist <tree-path>",
		Short: "Copy a worktree's evidence reports into the project root before disposal",
		Long: `Copy <tree-path>/.moai/reports/ content into the project root's
.moai/reports/worktrees/<tree-name>/, preserving relative paths.

Card worktrees hold their run-phase evidence under .moai/reports/ —
gitignored, machine-local, and destroyed with the tree. The worktree is
the only copy of that evidence, so run this verb BEFORE disposing a tree
(session-end keep/remove flow, or any manual removal). moai worktree
done runs the same routine automatically before removing an L2 tree.

Conflict policy: an existing destination file with different content is
never overwritten — the path is reported as skipped and everything else
still hoists. A tree without .moai/reports/ is a no-op.`,
		Args: cobra.ExactArgs(1),
		RunE: runHoist,
	}
}

// runHoist resolves the target, refuses paths outside the project root, and
// runs the shared hoist routine.
//
// @MX:SPEC: SPEC-REPORTS-LIFECYCLE-001
func runHoist(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	treePath, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("resolve %s: %w", args[0], err)
	}
	if info, statErr := os.Stat(treePath); statErr != nil || !info.IsDir() {
		return fmt.Errorf("not a directory: %s", treePath)
	}

	mainRoot, err := hoistTargetMainRoot(treePath)
	if err != nil {
		return fmt.Errorf("resolve project root from %s: %w (is %s inside a git worktree?)", treePath, err, treePath)
	}
	if !isInsideRoot(treePath, mainRoot) {
		return fmt.Errorf("refusing %s: outside the project root %s", treePath, mainRoot)
	}

	res, err := hoistWorktreeReports(treePath, mainRoot)
	if err != nil {
		return err
	}
	printHoistResult(out, res, treePath)
	return nil
}

// printHoistResult renders the hoist outcome (REQ-RLC-005 output clause:
// count and byte total; REQ-RLC-006: every unresolved path).
func printHoistResult(w io.Writer, res HoistResult, treePath string) {
	if res.Count == 0 && len(res.Skipped) == 0 {
		_, _ = fmt.Fprintf(w, "Nothing to hoist: %s has no .moai/reports/ content\n", treePath)
		return
	}
	_, _ = fmt.Fprintf(w, "Hoisted %d file(s), %d byte(s) -> %s\n", res.Count, res.Bytes, res.Dest)
	for _, rel := range res.Skipped {
		_, _ = fmt.Fprintf(w, "Skipped (differs at destination): %s\n", rel)
	}
}

// hoistBeforeDisposal is the done-side entry: resolve the project root from
// the target, hoist its evidence, and report to w. A root-resolution failure
// is a no-op (nothing resolvable to hoist — the mock/unregistered-tree
// shape); a hoist COPY failure is returned so the caller blocks the removal
// (the tree is the only copy of the evidence).
func hoistBeforeDisposal(w io.Writer, targetPath string) error {
	mainRoot, err := hoistTargetMainRoot(targetPath)
	if err != nil {
		// Unresolvable main root — no destination to hoist into. Not a
		// removal blocker: there is nothing this routine can act on.
		_, _ = fmt.Fprintf(w, "hoist skipped (project root unresolved from %s)\n", targetPath)
		return nil
	}
	res, err := hoistWorktreeReports(targetPath, mainRoot)
	if err != nil {
		return fmt.Errorf("hoist evidence before disposal: %w (rerun with --no-hoist to remove anyway)", err)
	}
	printHoistResult(w, res, targetPath)
	return nil
}
