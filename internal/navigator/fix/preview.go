package fix

// preview.go — M3.4 patch-preview generation (SPEC-NAVIGATOR-SYNC-005,
// AC-NS5-008b Go-engine half). Given the layer-2 draft subtrees at
// fix-drafts/<draft-id>/draft/ + the current live doc surfaces, generates ONE
// *.patch unified-diff file per in-scope stale subtree at
// fix-drafts/<draft-id>/draft/<doc-surface>.patch (design.md §D.1 layout).
//
// The AskUserQuestion approval-gate `preview` field consumes these patch files
// (the orchestrator renders them + truncates to ~12 lines for the preview
// pane — askuser-protocol.md § Preview Field Standards). The Go engine writes
// the FULL patch (no truncation); the orchestrator owns the ~12-line render.
//
// The approval-gate AskUserQuestion call itself is orchestrator-side and lives
// OUTSIDE this package (AC-NS5-008b — the gate is "verified at the
// orchestrator-integration level"); this package produces the previews only.
//
// Stdlib-only diff: go.mod carries NO diff dependency (sergi/go-diff /
// diffmatchpatch are absent), so UnifiedDiff implements a minimal LCS-based
// line-level unified diff using only the Go standard library.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PreviewSubtree identifies one in-scope draft subtree + its live doc surface,
// for patch-preview generation (AC-NS5-008b Go-half). The draft subtree's new
// (AI-drafted) content is read from DraftPath; the current live content is
// read from LiveDocPath. The patch is written to <draftDir>/<DocSurface>.patch.
type PreviewSubtree struct {
	// DocSurface is the doc-surface basename, e.g. "capability-map.md" or
	// "audit-report.json" (used for the --- a/ / +++ b/ headers + the .patch
	// filename).
	DocSurface string
	// DraftPath is the absolute path to the AI-drafted new content for this
	// subtree (layer-2 output under fix-drafts/<id>/draft/).
	DraftPath string
	// LiveDocPath is the absolute path to the current live doc surface (the
	// old/baseline content the patch is computed against).
	LiveDocPath string
}

// GeneratePreviews reads each in-scope subtree's draft (new) + live (old)
// content, computes the unified diff, and writes ONE *.patch file per subtree
// at <draftDir>/<DocSurface>.patch (design.md §D.1). Returns the list of patch
// file paths written (sorted lexicographically for determinism).
//
// Fail-open (REQ-NS5-009 spirit): a subtree whose live doc OR draft content is
// absent/unreadable is skipped (a diagnostic is written to the navigator-sync
// log via logFix), and the remaining subtrees still produce their patches —
// the function returns nil error on the skip path. A non-nil error is returned
// only for a fundamentally broken draftDir (MkdirAll failure).
func GeneratePreviews(draftDir string, subtrees []PreviewSubtree) ([]string, error) {
	if err := os.MkdirAll(draftDir, 0o755); err != nil {
		return nil, fmt.Errorf("preview: mkdir %s: %w", draftDir, err)
	}
	// rootForLog derives a project-root-ish path so logFix writes the
	// skip diagnostic alongside the other navigator-sync logs. draftDir is
	// <root>/.moai/project/navigator/fix-drafts/<id>/draft — climb 5 parents
	// to reach <root>. This is best-effort: a misaligned draftDir degrades
	// only the log location, not the preview output.
	rootForLog := deriveRootForLog(draftDir)

	written := make([]string, 0, len(subtrees))
	for _, sub := range subtrees {
		oldBytes, err := os.ReadFile(sub.LiveDocPath)
		if err != nil {
			logFix(rootForLog, fmt.Sprintf(
				"navigator-fix: preview skipped (live doc unreadable) for %s: %v",
				sub.DocSurface, err))
			continue
		}
		newBytes, err := os.ReadFile(sub.DraftPath)
		if err != nil {
			logFix(rootForLog, fmt.Sprintf(
				"navigator-fix: preview skipped (draft unreadable) for %s: %v",
				sub.DocSurface, err))
			continue
		}
		patch := UnifiedDiff(sub.DocSurface, string(oldBytes), string(newBytes))
		if patch == "" {
			// Identical old/new → no patch needed (no-op subtree). Still
			// counted as written=0 for this subtree; not an error.
			continue
		}
		patchPath := filepath.Join(draftDir, sub.DocSurface+".patch")
		if err := atomicWriteFile(patchPath, []byte(patch)); err != nil {
			logFix(rootForLog, fmt.Sprintf(
				"navigator-fix: preview write failed (fail-open) for %s: %v",
				sub.DocSurface, err))
			continue
		}
		written = append(written, patchPath)
	}
	return written, nil
}

// UnifiedDiff computes a line-level unified diff between oldText and newText
// using a minimal LCS-based algorithm (stdlib only — no external diff
// dependency per the go.mod audit). Returns a unified-diff string with:
//
//	--- a/<path>
//	+++ b/<path>
//	@@ -<oldStart>,<oldCount> +<newStart>,<newCount> @@
//	 <context line>
//	-<removed line>
//	+<added line>
//
// If oldText and newText are identical, returns "" (nothing to preview — the
// no-op case). The diff is emitted as a single hunk covering the full file,
// which is valid unified-diff format and appropriate for the approval-gate
// preview (the orchestrator truncates to ~12 lines for the preview pane).
//
// Pure function — no I/O, no side effects, deterministic.
func UnifiedDiff(path, oldText, newText string) string {
	oldLines := splitLines(oldText)
	newLines := splitLines(newText)

	// Identity fast-path: nothing to diff.
	if linesEqual(oldLines, newLines) {
		return ""
	}

	lcs := computeLCS(oldLines, newLines)
	ops := diffOps(oldLines, newLines, lcs)

	var b strings.Builder
	// Unified-diff headers.
	b.WriteString("--- a/" + path + "\n")
	b.WriteString("+++ b/" + path + "\n")

	// Single-hunk header: old range starts at line 1, new range starts at
	// line 1 (1-based per unified-diff convention). Counts are the line
	// totals of each side (context+removed for old; context+added for new).
	// A zero-count side is rendered as "0,0" per GNU diff convention for an
	// empty file (the header line number is then 0).
	oldCount, newCount := hunkCounts(ops)
	fmt.Fprintf(&b, "@@ -%s +%s @@\n", rangeStr(1, oldCount), rangeStr(1, newCount))

	for _, op := range ops {
		switch op.kind {
		case opEqual:
			b.WriteString(" " + op.line + "\n")
		case opRemove:
			b.WriteString("-" + op.line + "\n")
		case opAdd:
			b.WriteString("+" + op.line + "\n")
		}
	}
	return b.String()
}

// --- LCS-based line diff (minimal stdlib implementation) ---
//
// The classic dynamic-programming LCS table over line slices. For the doc
// surfaces M3 diffs (capability-map.md, audit-report.json — hundreds to low
// thousands of lines), the O(n*m) table is acceptable; a future optimization
// (Myers) is out of scope for M3.4. The LCS yields a stable, minimal edit
// script: equal lines are context, old-only lines are removals, new-only
// lines are additions.

const (
	opEqual = iota
	opRemove
	opAdd
)

type diffOp struct {
	kind int
	line string
}

// computeLCS returns the LCS DP table for a/b line slices.
func computeLCS(a, b []string) [][]int {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	return dp
}

// diffOps walks the LCS DP table to emit the ordered edit script (equal /
// remove / add ops).
func diffOps(a, b []string, dp [][]int) []diffOp {
	var ops []diffOp
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			ops = append(ops, diffOp{opEqual, a[i]})
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			ops = append(ops, diffOp{opRemove, a[i]})
			i++
		} else {
			ops = append(ops, diffOp{opAdd, b[j]})
			j++
		}
	}
	for i < len(a) {
		ops = append(ops, diffOp{opRemove, a[i]})
		i++
	}
	for j < len(b) {
		ops = append(ops, diffOp{opAdd, b[j]})
		j++
	}
	return ops
}

// hunkCounts computes the old-side and new-side line counts for the single
// hunk header: old = context+removed, new = context+added.
func hunkCounts(ops []diffOp) (oldCount, newCount int) {
	for _, op := range ops {
		switch op.kind {
		case opEqual:
			oldCount++
			newCount++
		case opRemove:
			oldCount++
		case opAdd:
			newCount++
		}
	}
	return oldCount, newCount
}

// rangeStr renders a unified-diff hunk range. GNU diff emits "<start>,<count>"
// when count != 1, and just "<start>" when count == 1 (and "0,0" / "<n+1>" for
// the empty-add-before / empty-file edge cases). For M3's full-file single
// hunk, the "<start>,<count>" form is the clearest + most parser-friendly.
func rangeStr(start, count int) string {
	if count == 0 {
		return "0,0"
	}
	return fmt.Sprintf("%d,%d", start, count)
}

// splitLines splits text into lines WITHOUT the trailing newline. A trailing
// newline yields no empty final element (matching `git diff` line semantics).
// An empty string yields an empty slice.
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	// strings.Split on "\n" leaves a trailing "" when text ends with "\n";
	// trim exactly one trailing newline so the line set matches git's view.
	trimmed := strings.TrimSuffix(text, "\n")
	return strings.Split(trimmed, "\n")
}

// linesEqual is the deep-equality check for the identity fast-path.
func linesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// deriveRootForLog climbs draftDir's parents to find a directory containing
// `.moai/logs/` (or its ancestor `.moai/project/navigator/`). Best-effort: if
// the climb finds nothing, returns the draftDir's parent so logFix at least
// writes somewhere reasonable. Used only for fail-open skip diagnostics.
func deriveRootForLog(draftDir string) string {
	dir := filepath.Clean(draftDir)
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".moai")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Dir(filepath.Clean(draftDir))
}
