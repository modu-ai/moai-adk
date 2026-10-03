package cli

import (
	"strings"
	"testing"
)

// findingLinesUnder returns the indented finding lines printed beneath the
// `todo list` row whose id is cardID.
func findingLinesUnder(out, cardID string) []string {
	var lines []string
	inside := false
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "\t") {
			inside = strings.HasPrefix(line, cardID+"\t")
			continue
		}
		if inside {
			lines = append(lines, line)
		}
	}
	return lines
}

// TestTodoFindingLineSuggestsDroppingTheSubjectNotTheRow — card t1470,
// GitHub #1732: a near-duplicate finding is printed beneath BOTH cards of
// the pair, and the drop/edit suggestion it carries must name the card the
// finding is about (the newly admitted subject), never the row it happens
// to be rendered under. Before the fix, the line under the ORIGINAL card
// suggested `moai todo drop <original>` — acting on it drops the card the
// operator meant to keep.
func TestTodoFindingLineSuggestsDroppingTheSubjectNotTheRow(t *testing.T) {
	todoFixture(t)
	if _, _, err := runTodo(t, "add", "Rework the auth middleware error paths"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	if _, _, err := runTodo(t, "add", "Rework auth middleware error paths"); err != nil {
		t.Fatalf("near-duplicate add: %v", err)
	}

	out, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	under := findingLinesUnder(out, "t1")
	if len(under) == 0 {
		t.Fatalf("no finding line under the original card t1:\n%s", out)
	}
	for _, line := range under {
		assertFindingSuggestsSubject(t, "list", line, "t1", "t2")
	}

	why, _, err := runTodo(t, "why", "t1")
	if err != nil {
		t.Fatalf("why: %v", err)
	}
	whyLines := 0
	for _, line := range strings.Split(strings.TrimSpace(why), "\n") {
		if !strings.Contains(line, "near-duplicate") {
			continue
		}
		whyLines++
		assertFindingSuggestsSubject(t, "why", line, "t1", "t2")
	}
	if whyLines == 0 {
		t.Fatalf("todo why t1 printed no near-duplicate finding:\n%s", why)
	}
}

func assertFindingSuggestsSubject(t *testing.T, surface, line, original, subject string) {
	t.Helper()
	for _, wrong := range []string{"moai todo drop " + original, "moai todo edit " + original} {
		if strings.Contains(line, wrong) {
			t.Errorf("%s: finding line under %s suggests %q — that acts on the original, not the near-duplicate:\n%s",
				surface, original, wrong, line)
		}
	}
	if !strings.Contains(line, "moai todo drop "+subject) {
		t.Errorf("%s: finding line does not name the subject %s in its drop suggestion:\n%s", surface, subject, line)
	}
}
