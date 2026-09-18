package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeAudit_ProjectPathIsLiteralData_AC_CLA_004(t *testing.T) {
	base := t.TempDir()
	marker := "MOAI_CLAUDE_PATH_INJECTION"
	root := filepath.Join(base, "repo;$(touch "+marker+")")
	originalRoot := newClaudeReviewTree(t, true)
	if err := os.Rename(originalRoot, root); err != nil {
		t.Fatal(err)
	}

	runner := validClaudeRunner()
	out := performClaudeAuditWith(
		context.Background(),
		claudeAuditRequest{Target: codexTargetUncommitted, ProjectRoot: root},
		"/test/bin/claude",
		runner,
		os.Environ(),
	)
	if out.Verdict != "pass" {
		t.Fatalf("literal project path audit = %+v, want pass", out)
	}
	if runner.auditDir != root {
		t.Fatalf("audit dir = %q, want literal %q", runner.auditDir, root)
	}
	for _, candidate := range []string{
		filepath.Join(base, marker),
		filepath.Join(root, marker),
		filepath.Join(mustGetwd(t), marker),
	} {
		if _, err := os.Stat(candidate); err == nil {
			t.Fatalf("project path executed as shell syntax; marker exists at %q", candidate)
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect marker %q: %v", candidate, err)
		}
	}
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
