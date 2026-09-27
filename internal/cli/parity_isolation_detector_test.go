package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestParityIsolationDetector exercises the AC-HPR-020 detector offline — the
// two mutations acceptance.md names: a Codex process started without the
// temporary CODEX_HOME, and a Claude run whose working directory is the
// repository root. A clean run yields no violation.
func TestParityIsolationDetector(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(string(filepath.Separator), "work", "repo")
	project := filepath.Join(tmp, "scratch")
	home := filepath.Join(tmp, "codex-home")

	if v := parityIsolationViolations([]string{"CODEX_HOME=" + home}, project, tmp, repo, true); len(v) != 0 {
		t.Fatalf("a clean Codex run was flagged: %v", v)
	}
	if v := parityIsolationViolations(nil, project, tmp, repo, false); len(v) != 0 {
		t.Fatalf("a clean Claude run was flagged: %v", v)
	}

	// Mutation 1: Codex without the temporary CODEX_HOME.
	if v := parityIsolationViolations([]string{"PATH=/bin"}, project, tmp, repo, true); len(v) != 1 || !strings.Contains(v[0], "CODEX_HOME") {
		t.Fatalf("a Codex process without a temporary CODEX_HOME was not flagged: %v", v)
	}
	if v := parityIsolationViolations([]string{"CODEX_HOME=/home/op/.codex"}, project, tmp, repo, true); len(v) != 1 {
		t.Fatalf("a Codex process on the operator's CODEX_HOME was not flagged: %v", v)
	}
	// Mutation 2: a Claude run with cwd at the repository root.
	v := parityIsolationViolations(nil, repo, tmp, repo, false)
	if len(v) != 2 || !strings.Contains(strings.Join(v, "\n"), "inside the repository") {
		t.Fatalf("a run at the repository root was not flagged on both counts: %v", v)
	}
	// A path that only shares a prefix is not "under" the root.
	if parityUnder(tmp+"-sibling", tmp) {
		t.Fatal("a sibling directory sharing a name prefix read as inside the root")
	}
}
