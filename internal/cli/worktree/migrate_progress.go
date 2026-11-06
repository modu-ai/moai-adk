package worktree

// migrate_progress.go — card t2053: the live "Migrating worktrees (i/N) <tree>"
// line of the migration apply.
//
// The apply used to print nothing until the whole engine returned; a project
// with hundreds of registered trees looked hung. The line reuses
// tui.ProgressLine — the in-place line `moai update` already uses for every
// other long step — so the existing conventions hold unchanged: on a terminal
// each tree rewrites one line in place, on a non-TTY writer (a pipe, a CI log,
// a test buffer) each tree gets one plain line with no escape sequences. The
// final rows are printed by the callers exactly as before.

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// treeProgress renders the apply's live lines. A nil *treeProgress is a no-op,
// which is how the dry run and the --json report ask for silence.
type treeProgress struct {
	out   io.Writer
	total int
	line  *tui.ProgressLineHandle
}

// newTreeProgress returns a progress renderer writing to out, or nil when out
// is nil.
func newTreeProgress(out io.Writer) *treeProgress {
	if out == nil {
		return nil
	}
	return &treeProgress{out: out}
}

// begin announces how many trees the apply is about to move. A run with
// nothing to move prints nothing.
func (p *treeProgress) begin(total int) {
	if p == nil || total <= 0 {
		return
	}
	p.total = total
	_, _ = fmt.Fprintf(p.out, "Worktree migration: %d tree(s) to move\n", total)
}

// tree reports that the index-th tree (1-based) is about to be processed.
func (p *treeProgress) tree(index int, source string) {
	if p == nil || p.total <= 0 {
		return
	}
	label := fmt.Sprintf("Migrating worktrees (%d/%d) %s", index, p.total, filepath.Base(source))
	if p.line == nil {
		p.line = tui.ProgressLine(p.out, label, nil)
		return
	}
	p.line.Update(label)
}

// end closes the live line so the rows that follow start on a fresh line. It
// is safe to call more than once.
func (p *treeProgress) end() {
	if p == nil || p.line == nil {
		return
	}
	p.line.Done(fmt.Sprintf("Worktree migration processed %d tree(s)", p.total))
	p.line = nil
}
