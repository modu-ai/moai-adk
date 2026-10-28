//go:build unix

// FIFO fixture for the instruction reader (sync-audit F2, follow-up F6):
// mkfifo exists only on unix, so the test carries the repo's unix build
// tag (the reconcile_fifo_unix_test.go pattern) and the windows-side test
// tree stays free of syscall references — a runtime t.Skip cannot prevent
// compilation.
package hook

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestInstructionReaderSkipsFIFO pins sync-audit F2: a writer-less FIFO at
// the local instruction-file path must not hang the session-start hook —
// the resolver treats it as absent and the aggregate never opens it
// (os.Stat passes FIFOs; an unguarded ReadFile would block forever).
func TestInstructionReaderSkipsFIFO(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude", "rules", "moai"), 0o750); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"CLAUDE.md":                         "# Project\n",
		".claude/rules/moai/always-core.md": "# Always-loaded core rule\n",
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(root, rel), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	fifo := filepath.Join(root, "CLAUDE.local.md")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skipf("mkfifo unavailable: %v", err)
	}

	// The resolver reads the FIFO as absent (canonical shared file answers).
	gotPath, gotLabel := resolveLocalInstructionFileDetail(root)
	if gotPath != "" || gotLabel != "" {
		t.Errorf("local resolution = (%q, %q), want empty — a FIFO is not a readable instruction file", gotPath, gotLabel)
	}

	// The aggregate enumerates without opening the FIFO; it must complete
	// (a hang fails this test via the test timeout) and exclude it.
	set := instructionFileSet(root)
	for _, p := range set {
		if p == fifo {
			t.Error("instructionFileSet enumerated the FIFO — the aggregate would block on read")
		}
	}
	if aggregateInstructionChars(root) == 0 {
		t.Fatal("fixture aggregate is zero; fixture is broken")
	}
}
