package worktree

// Card t2106: appendMoveListLine's duplicate check.
//
// The function documents the rule as "no append when the line equals the LATEST
// entry naming either of its two paths". It compared the new line with the
// FIRST line of the file instead (revLines is documented as "last first" and
// returned the lines in file order). Two visible effects:
//   - a new `pending` line whose text equalled the file's first line was dropped
//     even though a `reversed` record for the same tree sat between them. Every
//     line carries a `ts` of whole seconds, so the texts are equal only when
//     both were made within the same second: TestApply_PendingIntentButNeverMoved_
//     IsSettledThenMoved failed 4/10 (RC42), 8/10 (RC43), 1/10 (this tree, t2106);
//   - a true duplicate of the latest record was appended again whenever another
//     tree's record happened to be first in the file.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mlTestFile(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "move-list.jsonl")
	if len(lines) > 0 {
		if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func mlLineCount(t *testing.T, p string) int {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, l := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(l) != "" {
			n++
		}
	}
	return n
}

// Control (passes before and after): an identical re-run right after the line
// adds nothing.
func TestAppendMoveListLine_IdenticalRerunRightAfterAddsNothing(t *testing.T) {
	line := m6MoveLine("pending", "/r/.moai/worktrees/a", "/h/a", "a:1")
	p := mlTestFile(t, line)
	if err := appendMoveListLine(p, line); err != nil {
		t.Fatal(err)
	}
	if got := mlLineCount(t, p); got != 1 {
		t.Errorf("an identical re-run appended: %d lines, want 1", got)
	}
}

// RED before the repair: the same text after a record of the same paths that
// differs (a `reversed` settlement) is a new record, not a duplicate.
func TestAppendMoveListLine_SameTextAfterASettlementOfTheSamePathsIsAppended(t *testing.T) {
	pending := m6MoveLine("pending", "/r/.moai/worktrees/a", "/h/a", "a:1")
	reversed := m6MoveLine("reversed", "/h/a", "/r/.moai/worktrees/a", "a:1")
	p := mlTestFile(t, pending, reversed)
	if err := appendMoveListLine(p, pending); err != nil {
		t.Fatal(err)
	}
	if got := mlLineCount(t, p); got != 3 {
		t.Errorf("pending, reversed, pending must stay three records; got %d lines", got)
	}
}

// RED before the repair: the latest record naming these paths is the line
// itself, so appending it again is a duplicate - whatever another tree's record
// is first in the file.
func TestAppendMoveListLine_DuplicateOfTheLatestRecordIsDroppedWhateverComesFirst(t *testing.T) {
	other := m6MoveLine("moved", "/r/.moai/worktrees/z", "/h/z", "z:1")
	mine := m6MoveLine("pending", "/r/.moai/worktrees/a", "/h/a", "a:1")
	p := mlTestFile(t, other, mine)
	if err := appendMoveListLine(p, mine); err != nil {
		t.Fatal(err)
	}
	if got := mlLineCount(t, p); got != 2 {
		t.Errorf("a duplicate of the latest record for these paths was appended: %d lines, want 2", got)
	}
}

// Control: another tree's record in between does not hide the latest record of
// THESE paths, and a record that differs is appended.
func TestAppendMoveListLine_AnotherTreesRecordInBetweenDoesNotChangeTheComparison(t *testing.T) {
	mine := m6MoveLine("pending", "/r/.moai/worktrees/a", "/h/a", "a:1")
	other := m6MoveLine("moved", "/r/.moai/worktrees/z", "/h/z", "z:1")
	p := mlTestFile(t, mine, other)
	if err := appendMoveListLine(p, mine); err != nil {
		t.Fatal(err)
	}
	if got := mlLineCount(t, p); got != 2 {
		t.Errorf("the latest record naming these paths is the same line; it was appended again: %d lines, want 2", got)
	}
	differing := m6MoveLine("moved", "/r/.moai/worktrees/a", "/h/a", "a:1")
	if err := appendMoveListLine(p, differing); err != nil {
		t.Fatal(err)
	}
	if got := mlLineCount(t, p); got != 3 {
		t.Errorf("a record that differs from the latest one for these paths must be appended: %d lines, want 3", got)
	}
}

// Control: a line that is not a record cannot be matched by path; it is judged
// against the latest line only, so an immediate repeat still adds nothing.
func TestAppendMoveListLine_ANonRecordLineIsComparedWithTheLatestLineOnly(t *testing.T) {
	garbage := `{"result":"moved"}`
	other := m6MoveLine("moved", "/r/.moai/worktrees/z", "/h/z", "z:1")
	p := mlTestFile(t, other, garbage)
	if err := appendMoveListLine(p, garbage); err != nil {
		t.Fatal(err)
	}
	if got := mlLineCount(t, p); got != 2 {
		t.Errorf("an immediate repeat of a non-record line was appended: %d lines, want 2", got)
	}
	p2 := mlTestFile(t, garbage, other)
	if err := appendMoveListLine(p2, garbage); err != nil {
		t.Fatal(err)
	}
	if got := mlLineCount(t, p2); got != 3 {
		t.Errorf("a non-record line equal only to the FIRST line must be appended: %d lines, want 3", got)
	}
}
