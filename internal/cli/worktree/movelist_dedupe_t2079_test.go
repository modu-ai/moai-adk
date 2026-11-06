package worktree

// movelist_dedupe_t2079_test.go — card t2079.
//
// appendMoveListLine drops a line identical to the LATEST entry (an idempotent
// re-run). It compared against the FIRST line of the file instead, because
// revLines kept file order while documenting "last first". A line identical to
// the first entry was then dropped even when other lines came after it, which
// loses the second `pending` of a settled-then-moved tree whenever both
// `pending` lines fall in the same RFC3339 second (the clock-dependent failure
// of TestApply_PendingIntentButNeverMoved_IsSettledThenMoved).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func t2079ReadLines(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read move list: %v", err)
	}
	return strings.Split(strings.TrimSpace(string(raw)), "\n")
}

// A line equal to an EARLIER entry, with a different entry after it, is new
// information and is appended.
func TestAppendMoveListLine_EqualToEarlierEntryButNotLatest_IsAppended(t *testing.T) {
	mlPath := filepath.Join(t.TempDir(), "move-list.jsonl")
	a, b := `{"result":"pending"}`, `{"result":"reversed"}`
	for _, line := range []string{a, b, a} {
		if err := appendMoveListLine(mlPath, line); err != nil {
			t.Fatalf("append %s: %v", line, err)
		}
	}
	got := t2079ReadLines(t, mlPath)
	if want := []string{a, b, a}; strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("move list = %v, want %v", got, want)
	}
}

// A line equal to the latest entry is still the idempotent re-run: not appended.
func TestAppendMoveListLine_EqualToLatestEntry_IsNotAppended(t *testing.T) {
	mlPath := filepath.Join(t.TempDir(), "move-list.jsonl")
	a, b := `{"result":"pending"}`, `{"result":"moved"}`
	for _, line := range []string{a, b, b} {
		if err := appendMoveListLine(mlPath, line); err != nil {
			t.Fatalf("append %s: %v", line, err)
		}
	}
	got := t2079ReadLines(t, mlPath)
	if want := []string{a, b}; strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("move list = %v, want %v", got, want)
	}
}
