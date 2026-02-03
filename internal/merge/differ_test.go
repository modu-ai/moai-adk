package merge

import (
	"strings"
	"testing"
)

func TestDiffLines_IdenticalInput(t *testing.T) {
	t.Parallel()

	a := []string{"line1", "line2", "line3"}
	b := []string{"line1", "line2", "line3"}

	edits := DiffLines(a, b)

	if len(edits) != 0 {
		t.Errorf("expected 0 edits for identical input, got %d: %v", len(edits), edits)
	}
}

func TestDiffLines_AllInserted(t *testing.T) {
	t.Parallel()

	a := []string{}
	b := []string{"new1", "new2"}

	edits := DiffLines(a, b)

	insertCount := 0
	for _, e := range edits {
		if e.Op == OpInsert {
			insertCount++
		}
	}
	if insertCount != 2 {
		t.Errorf("expected 2 inserts, got %d", insertCount)
	}
}

func TestDiffLines_AllDeleted(t *testing.T) {
	t.Parallel()

	a := []string{"old1", "old2"}
	b := []string{}

	edits := DiffLines(a, b)

	deleteCount := 0
	for _, e := range edits {
		if e.Op == OpDelete {
			deleteCount++
		}
	}
	if deleteCount != 2 {
		t.Errorf("expected 2 deletes, got %d", deleteCount)
	}
}

func TestDiffLines_Modification(t *testing.T) {
	t.Parallel()

	a := []string{"A", "B", "C"}
	b := []string{"A", "B_mod", "C"}

	edits := DiffLines(a, b)

	// A modification is represented as a delete + insert pair.
	hasDelete := false
	hasInsert := false
	for _, e := range edits {
		if e.Op == OpDelete && e.OldLine == 1 {
			hasDelete = true
		}
		if e.Op == OpInsert && e.NewText == "B_mod" {
			hasInsert = true
		}
	}

	if !hasDelete || !hasInsert {
		t.Errorf("expected delete+insert for modification, got edits: %v", edits)
	}
}

func TestDiffLines_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		a           []string
		b           []string
		wantInserts int
		wantDeletes int
	}{
		{
			name:        "empty to empty",
			a:           []string{},
			b:           []string{},
			wantInserts: 0,
			wantDeletes: 0,
		},
		{
			name:        "add one line at end",
			a:           []string{"A"},
			b:           []string{"A", "B"},
			wantInserts: 1,
			wantDeletes: 0,
		},
		{
			name:        "remove one line from start",
			a:           []string{"A", "B"},
			b:           []string{"B"},
			wantInserts: 0,
			wantDeletes: 1,
		},
		{
			name:        "swap two lines",
			a:           []string{"A", "B"},
			b:           []string{"B", "A"},
			wantInserts: 1,
			wantDeletes: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			edits := DiffLines(tt.a, tt.b)

			inserts := 0
			deletes := 0
			for _, e := range edits {
				switch e.Op {
				case OpInsert:
					inserts++
				case OpDelete:
					deletes++
				}
			}

			if inserts != tt.wantInserts {
				t.Errorf("inserts: got %d, want %d (edits: %v)", inserts, tt.wantInserts, edits)
			}
			if deletes != tt.wantDeletes {
				t.Errorf("deletes: got %d, want %d (edits: %v)", deletes, tt.wantDeletes, edits)
			}
		})
	}
}

func TestUnifiedDiff_NoDifference(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB\nC")
	current := []byte("A\nB\nC")

	result := UnifiedDiff("file.txt", base, current)

	if result != "" {
		t.Errorf("expected empty diff for identical files, got:\n%s", result)
	}
}

func TestUnifiedDiff_WithChanges(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB\nC")
	current := []byte("A\nB_mod\nC")

	result := UnifiedDiff("file.txt", base, current)

	if result == "" {
		t.Fatal("expected non-empty diff for changed files")
	}
	if !strings.Contains(result, "-B") {
		t.Errorf("expected diff to contain deleted line '-B', got:\n%s", result)
	}
	if !strings.Contains(result, "+B_mod") {
		t.Errorf("expected diff to contain added line '+B_mod', got:\n%s", result)
	}
}

func TestUnifiedDiff_HeaderFormat(t *testing.T) {
	t.Parallel()

	base := []byte("old")
	current := []byte("new")

	result := UnifiedDiff("test.txt", base, current)

	if !strings.Contains(result, "--- a/test.txt") {
		t.Errorf("expected unified diff header with '--- a/test.txt', got:\n%s", result)
	}
	if !strings.Contains(result, "+++ b/test.txt") {
		t.Errorf("expected unified diff header with '+++ b/test.txt', got:\n%s", result)
	}
}

func TestSplitLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", []string{}},
		{"single line", "hello", []string{"hello"}},
		{"two lines", "a\nb", []string{"a", "b"}},
		{"trailing newline", "a\nb\n", []string{"a", "b"}},
		{"only newline", "\n", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := splitLines(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitLines(%q): got %d lines %v, want %d lines %v", tt.input, len(got), got, len(tt.want), tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitLines(%q)[%d]: got %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}
