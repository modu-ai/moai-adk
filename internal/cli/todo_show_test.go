// todo_show_test.go — SPEC-TODO-SURFACE-POLISH-001 M1 (card t1349): the
// `moai todo show <id|n>` single-card read surface.
//
// AC-TSP-001 exercises the three scenarios (live line with the full text,
// absent id with the issued-mark qualifier, flattened body) through the same
// lookup machine `history` uses; AC-TSP-002 pins the read-only contract —
// the store file's bytes and mtime survive the call, and stdout is
// byte-identical whether or not stderr carries a disclosure.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// showFields splits a show stdout line into its tab-separated fields.
func showFields(t *testing.T, line string) []string {
	t.Helper()
	fields := strings.Split(strings.TrimSuffix(line, "\n"), "\t")
	return fields
}

// AC-TSP-001 — the three show scenarios. The parent test carries the
// scenario subtests acceptance.md names; run them with
// -run '^TestTodoShow$/^<scenario>$'.
func TestTodoShow(t *testing.T) {
	t.Run("live line carries the full text", func(t *testing.T) {
		todoFixture(t)
		const text = "polish the todo surface: show verb, add -f misparse, list limit"
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("add: %v", err)
		}

		out, _, err := runTodo(t, "show", "t1")
		if err != nil {
			t.Fatalf("show t1: %v", err)
		}
		lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
		if len(lines) != 1 {
			t.Fatalf("show t1 stdout = %d lines, want exactly one: %q", len(lines), out)
		}
		fields := showFields(t, lines[0])
		// id, live, state, landing, picked_at, dropped_at, text (text LAST).
		if len(fields) != 7 {
			t.Fatalf("show t1 line = %d fields, want 7: %q", len(fields), lines[0])
		}
		if fields[0] != "t1" || fields[1] != "live" || fields[2] != string(factory.BacklogStateQueued) {
			t.Errorf("show t1 id/live/state = %q/%q/%q, want t1/live/queued", fields[0], fields[1], fields[2])
		}
		if !strings.HasPrefix(fields[3], "landing=") {
			t.Errorf("show t1 landing field = %q, want the landing=<...> cell", fields[3])
		}
		if got := fields[6]; got != todoPRCell(text) {
			t.Errorf("show t1 text field = %q, want the verbatim body %q (flattened)", got, todoPRCell(text))
		}
	})

	t.Run("absent id prints absent and the issued-mark qualifier", func(t *testing.T) {
		_, _ = seedFiveAndDeleteT3(t)

		out, errOut, err := runTodo(t, "show", "t3")
		if err != nil {
			t.Fatalf("show t3: %v (stderr %q)", err, errOut)
		}
		if out != "t3\tabsent\n" {
			t.Errorf("show t3 stdout = %q, want exactly the single absent line", out)
		}
		if !strings.Contains(errOut, "t3") || !strings.Contains(errOut, todoHistoryPreArchiveNote) {
			t.Errorf("show t3 stderr = %q, want the at-or-below-the-mark qualifier naming t3", errOut)
		}
	})

	t.Run("body with tabs and newlines flattens without truncation", func(t *testing.T) {
		todoFixture(t)
		const text = "line one\nline two\twith a tab\nline three"
		if _, _, err := runTodo(t, "add", text); err != nil {
			t.Fatalf("add: %v", err)
		}

		out, _, err := runTodo(t, "show", "t1")
		if err != nil {
			t.Fatalf("show t1: %v", err)
		}
		lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
		if len(lines) != 1 {
			t.Fatalf("show t1 stdout = %d lines, want one (the body flattened): %q", len(lines), out)
		}
		fields := showFields(t, lines[0])
		if got := fields[len(fields)-1]; got != todoPRCell(text) {
			t.Errorf("show t1 body field = %q, want the flattened full body %q (len %d vs %d)",
				got, todoPRCell(text), len(got), len(todoPRCell(text)))
		}
		if strings.Contains(out, "\t\t") {
			t.Errorf("show t1 stdout = %q, want no doubled tabs from an empty body cell", out)
		}
	})
}

// AC-TSP-002 — show is read-only and its stdout is disclosure-independent.
func TestTodoShowReadOnly(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "read-only show fixture card"); err != nil {
		t.Fatalf("add: %v", err)
	}

	// The engine artifact is the store the read touches; the backlog.json
	// NAME is only the store's address (the SQLite sibling is the data).
	queuePath := store.EnginePath()
	digest := func() (string, os.FileInfo) {
		raw, err := os.ReadFile(queuePath)
		if err != nil {
			t.Fatalf("read queue %s: %v", queuePath, err)
		}
		sum := sha256.Sum256(raw)
		info, err := os.Stat(queuePath)
		if err != nil {
			t.Fatalf("stat queue: %v", err)
		}
		return hex.EncodeToString(sum[:]), info
	}
	beforeSum, beforeInfo := digest()

	out1, _, err := runTodo(t, "show", "t1")
	if err != nil {
		t.Fatalf("show t1 (first): %v", err)
	}
	afterSum, afterInfo := digest()
	if afterSum != beforeSum {
		t.Errorf("queue sha256 changed across show: %s -> %s", beforeSum, afterSum)
	}
	if !afterInfo.ModTime().Equal(beforeInfo.ModTime()) {
		t.Errorf("queue mtime changed across show: %v -> %v", beforeInfo.ModTime(), afterInfo.ModTime())
	}

	// A non-authoritative backlog.json beside the store fires the layout
	// disclosure on stderr; stdout must stay byte-identical (REQ-BJD-004
	// lineage, restated for show by REQ-TSP-003).
	ghostPath := store.Path()
	if err := os.WriteFile(ghostPath, []byte(`{"version":1,"items":[]}`), 0o600); err != nil {
		t.Fatalf("plant disclosure-provoking backlog.json: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(ghostPath) })

	out2, _, err := runTodo(t, "show", "t1")
	if err != nil {
		t.Fatalf("show t1 (second): %v", err)
	}
	if out1 != out2 {
		t.Errorf("show stdout changed across a disclosure-bearing run:\nfirst:  %q\nsecond: %q", out1, out2)
	}
	if !strings.Contains(out1, "t1") {
		t.Errorf("show stdout = %q, want it to name the card id", out1)
	}
	_ = root
}
