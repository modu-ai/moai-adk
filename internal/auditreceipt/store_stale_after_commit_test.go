package auditreceipt

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Receipt freshness is SPEC-MCP-NEUTRALITY-001 REQ-MN-020 and nothing more: a receipt
// whose bound head differs from the current HEAD is stale, whatever the commits in
// between touched. SPEC-AUDIT-ENGINE-CONVERGENCE-001 excludes the evidence-only
// relaxation (B3) by operator decision (card t2017), so a commit that only files the
// audit's own verdict or edits a progress.md leaves the receipt stale like any other.

const staleProgress = "# P\n\n## §E.1 One\n\na\n\n## §E.2 Run-phase Evidence\n\nrow\n\n## §E.3 Three\n\nc\n"

func staleWrite(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func staleCommit(t *testing.T, root, msg string) {
	t.Helper()
	if _, err := runGit(t, root, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := runGit(t, root, "-c", "user.email=fx@example.com", "-c", "user.name=fx", "commit", "-q", "--allow-empty", "-m", msg); err != nil {
		t.Fatal(err)
	}
}

// staleFixture returns a repository with an audited SPEC progress.md and a report
// directory committed, and the receipt written for an audit that judged that commit.
func staleFixture(t *testing.T) (root string, r Receipt, start *StartMarker) {
	t.Helper()
	root = initContentRepo(t)
	staleWrite(t, root, ".gitignore", ".moai/state/\n")
	staleWrite(t, root, ".moai/specs/SPEC-X-001/progress.md", staleProgress)
	staleWrite(t, root, ".moai/reports/t1/verdict.md", "verdict: PASS\n")
	staleCommit(t, root, "base")
	id, err := WriteReceipt(root, &Receipt{Tool: ToolAuditMulti, TreeRoot: root, CreatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	r, err = ReadReceipt(root, id)
	if err != nil {
		t.Fatal(err)
	}
	return root, r, &StartMarker{AgentID: "a", TreeRoot: root, StartedAt: time.Now().UTC().Add(-time.Hour)}
}

func TestCheckCitedReceipts_AnyCommitAfterTheAuditLeavesTheReceiptStale(t *testing.T) {
	rows := []struct {
		name string
		do   func(t *testing.T, root string)
	}{
		{"a verdict file in a report directory", func(t *testing.T, root string) {
			staleWrite(t, root, ".moai/reports/t1/verdict.md", "verdict: PASS\nrevised\n")
		}},
		{"a sync-audit report and a second-review record", func(t *testing.T, root string) {
			staleWrite(t, root, ".moai/reports/t1/sync-audit-1.md", "sync\n")
			staleWrite(t, root, ".moai/reports/t1/second-review.jsonl", "{}\n")
		}},
		{"section E.3 of the audited progress.md", func(t *testing.T, root string) {
			staleWrite(t, root, ".moai/specs/SPEC-X-001/progress.md", staleProgress+"\nrewritten\n")
		}},
		{"section E.2 of the audited progress.md", func(t *testing.T, root string) {
			staleWrite(t, root, ".moai/specs/SPEC-X-001/progress.md", "# P\n\n## §E.1 One\n\na\n\n## §E.2 Run-phase Evidence\n\nrow edited\n\n## §E.3 Three\n\nc\n")
		}},
		{"a code file", func(t *testing.T, root string) {
			staleWrite(t, root, "a.go", "package a\n\nvar X = 1\n")
		}},
		{"nothing at all (an empty commit)", func(t *testing.T, root string) {}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			root, r, start := staleFixture(t)
			row.do(t, root)
			staleCommit(t, root, row.name)
			if ok, cause := CheckCitedReceiptsSince(root, start, time.Time{}, []string{r.ReceiptID}); ok || cause != CauseReceiptTreeStateChanged {
				t.Errorf("CheckCitedReceiptsSince = (%v, %q), want (false, %q)", ok, cause, CauseReceiptTreeStateChanged)
			}
		})
	}

	t.Run("control: the head did not move", func(t *testing.T) {
		root, r, start := staleFixture(t)
		if ok, cause := CheckCitedReceiptsSince(root, start, time.Time{}, []string{r.ReceiptID}); !ok {
			t.Errorf("CheckCitedReceiptsSince = (%v, %q), want the receipt fresh", ok, cause)
		}
	})
	t.Run("control: an uncommitted edit at the same head", func(t *testing.T) {
		root, r, start := staleFixture(t)
		staleWrite(t, root, "a.go", "package a\n\nvar Y = 2\n")
		if ok, cause := CheckCitedReceiptsSince(root, start, time.Time{}, []string{r.ReceiptID}); ok || cause != CauseReceiptTreeStateChanged {
			t.Errorf("CheckCitedReceiptsSince = (%v, %q), want (false, %q)", ok, cause, CauseReceiptTreeStateChanged)
		}
	})
}
