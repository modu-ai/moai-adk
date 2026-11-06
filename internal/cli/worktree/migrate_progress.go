package worktree

// migrate_progress.go — the live lines of the migration apply.
//
// Card t2053 added the "Migrating worktrees (i/N) <tree>" line: the apply used
// to print nothing until the whole engine returned, and a project with hundreds
// of registered trees looked hung. Card t2070 closes the other silent stretch:
// the plan, which checks every registered tree before the first one moves, ran
// for minutes with no output at all. It now announces how many trees it is about
// to check and that this can take a long time, shows a throttled "Checking
// worktrees (i/N)" line, and the move line carries the elapsed time and an
// estimate of what is left.
//
// The lines reuse tui.ProgressLine — the in-place line `moai update` already
// uses for every other long step — so the existing conventions hold unchanged:
// on a terminal each tree rewrites one line in place, on a non-TTY writer (a
// pipe, a CI log, a test buffer) each update gets one plain line with no escape
// sequences. The final rows are printed by the callers exactly as before.

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// scanRefreshInterval is the least time between two "Checking worktrees" lines.
// A plan over hundreds of trees would otherwise write hundreds of lines to a
// log, and rewrite a terminal faster than it can be read.
const scanRefreshInterval = time.Second

// treeProgress renders the apply's live lines. A nil *treeProgress is a no-op,
// which is how the dry run and the --json report ask for silence.
type treeProgress struct {
	out   io.Writer
	total int
	line  *tui.ProgressLineHandle
	now   func() time.Time // a test's clock; nil reads the wall clock

	scanTotal int
	scanLine  *tui.ProgressLineHandle
	lastScan  time.Time

	started time.Time
}

// newTreeProgress returns a progress renderer writing to out, or nil when out
// is nil.
func newTreeProgress(out io.Writer) *treeProgress {
	if out == nil {
		return nil
	}
	return &treeProgress{out: out}
}

func (p *treeProgress) clock() time.Time {
	if p.now != nil {
		return p.now()
	}
	return time.Now()
}

// planBegin announces, before the plan does any work, how many registered trees
// it is about to check and that this can take a long time. A plan with nothing
// to check prints nothing.
func (p *treeProgress) planBegin(total int) {
	if p == nil || total <= 0 {
		return
	}
	p.scanTotal = total
	_, _ = fmt.Fprintf(p.out,
		"Worktree migration: checking %d registered tree(s). With many trees this can take a long time; nothing is moved until the checks finish.\n", total)
}

// scan reports that the index-th tree (1-based) of total is being checked. The
// first and the last tree are always shown; the ones between are shown at most
// once per scanRefreshInterval.
func (p *treeProgress) scan(index, total int) {
	if p == nil || p.scanTotal <= 0 {
		return
	}
	now := p.clock()
	label := fmt.Sprintf("Checking worktrees (%d/%d)", index, total)
	if p.scanLine == nil {
		p.scanLine = tui.ProgressLine(p.out, label, nil)
		p.lastScan = now
		return
	}
	if index == total || now.Sub(p.lastScan) >= scanRefreshInterval {
		p.scanLine.Update(label)
		p.lastScan = now
	}
}

// planEnd closes the scan line so what follows starts on a fresh line. It is
// safe to call more than once.
func (p *treeProgress) planEnd() {
	if p == nil || p.scanLine == nil {
		return
	}
	p.scanLine.Done(fmt.Sprintf("Worktree migration checked %d tree(s)", p.scanTotal))
	p.scanLine = nil
}

// begin announces how many trees the apply is about to move. A run with
// nothing to move prints nothing.
func (p *treeProgress) begin(total int) {
	if p == nil || total <= 0 {
		return
	}
	p.total = total
	p.started = p.clock()
	_, _ = fmt.Fprintf(p.out, "Worktree migration: %d tree(s) to move\n", total)
}

// tree reports that the index-th tree (1-based) is about to be processed. From
// the second tree on it adds the time spent so far and an estimate of what is
// left, taken from the average of the trees finished; before a whole second has
// passed there is nothing worth saying.
func (p *treeProgress) tree(index int, source string) {
	if p == nil || p.total <= 0 {
		return
	}
	label := fmt.Sprintf("Migrating worktrees (%d/%d) %s", index, p.total, filepath.Base(source))
	if index > 1 {
		spent := p.clock().Sub(p.started)
		if spent.Round(time.Second) >= time.Second {
			done := index - 1
			left := (spent / time.Duration(done) * time.Duration(p.total-done)).Round(time.Second)
			label += fmt.Sprintf(" (elapsed %s, about %s left)", spent.Round(time.Second), left)
		}
	}
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
