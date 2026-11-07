package homestate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeReport(t *testing.T, dir, card, name string) string {
	t.Helper()
	p := filepath.Join(dir, ".moai", "reports", card, name)
	frWrite(t, p, "x\n")
	return p
}

func TestSelectCardVerdictFile(t *testing.T) {
	t.Run("verdict.md wins over sync-audit files", func(t *testing.T) {
		dir := t.TempDir()
		writeReport(t, dir, "c1", "sync-audit-iter9.md")
		want := writeReport(t, dir, "c1", "verdict.md")
		got, err := SelectCardVerdictFile(dir, "c1")
		if err != nil || got != want {
			t.Fatalf("got %q, %v; want %q", got, err, want)
		}
	})
	t.Run("highest iteration wins, bare name is iteration 1", func(t *testing.T) {
		dir := t.TempDir()
		writeReport(t, dir, "c1", "sync-audit.md")
		writeReport(t, dir, "c1", "sync-audit-iter2.md")
		want := writeReport(t, dir, "c1", "sync-audit-iter10.md")
		got, err := SelectCardVerdictFile(dir, "c1")
		if err != nil || got != want {
			t.Fatalf("got %q, %v; want %q", got, err, want)
		}
	})
	t.Run("equal iteration breaks to newest mtime", func(t *testing.T) {
		dir := t.TempDir()
		older := writeReport(t, dir, "c1", "sync-audit-a.md")
		newer := writeReport(t, dir, "c1", "sync-audit-b.md")
		old := time.Now().Add(-time.Hour)
		if err := os.Chtimes(older, old, old); err != nil {
			t.Fatal(err)
		}
		got, err := SelectCardVerdictFile(dir, "c1")
		if err != nil || got != newer {
			t.Fatalf("got %q, %v; want %q", got, err, newer)
		}
	})
	t.Run("none present names both forms", func(t *testing.T) {
		_, err := SelectCardVerdictFile(t.TempDir(), "c1")
		if !errors.Is(err, ErrEvidence) || !strings.Contains(err.Error(), "no verdict file (verdict.md or sync-audit*.md)") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("unsafe card id refuses", func(t *testing.T) {
		if _, err := SelectCardVerdictFile(t.TempDir(), "../x"); !errors.Is(err, ErrEvidence) {
			t.Fatalf("err = %v", err)
		}
	})
}

// The T13 and rebind readers both switch to the shared selector for the
// sync-audit phase: a lane that wrote only verdict.md is no longer refused.
func TestSyncAuditReadersAcceptVerdictMd(t *testing.T) {
	repo := frNewRepo(t, false)
	frWriteVerdict(t, repo.Dir, "c1", "verdict.md", "PASS", repo.Merge)
	v, err := readAuditVerdict(repo.Dir, "c1", "sync-audit", repo.Merge)
	if err != nil || !strings.HasSuffix(v.Path, "verdict.md") {
		t.Fatalf("T13 reader: %v %+v", v, err)
	}
	full, err := ResolveRebindEvidence(context.Background(), Card{CardID: "c1", WorktreePath: repo.Dir}, "sync-audit")
	if err != nil || full != repo.Merge {
		t.Fatalf("rebind reader: %q %v; want %q", full, err, repo.Merge)
	}
	// plan-audit keeps its own glob.
	if _, err := readAuditVerdict(repo.Dir, "c1", "plan-audit", repo.Merge); !errors.Is(err, ErrEvidence) {
		t.Fatalf("plan-audit must not see verdict.md: %v", err)
	}
}

func commitFile(t *testing.T, dir, rel string) string {
	t.Helper()
	frWrite(t, filepath.Join(dir, rel), "change "+rel+"\n")
	frGit(t, dir, "add", "-A")
	frGit(t, dir, "commit", "-q", "-m", "add "+rel)
	return frGit(t, dir, "rev-parse", "HEAD")
}

func TestAuditedToEvidenceBinding(t *testing.T) {
	repo := frNewRepo(t, false)
	const card = "c1"

	t.Run("same sha passes", func(t *testing.T) {
		frWriteVerdict(t, repo.Dir, card, "verdict.md", "PASS", repo.Merge)
		if _, err := readAuditVerdict(repo.Dir, card, "sync-audit", repo.Merge); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("reports-only offset passes", func(t *testing.T) {
		frWriteVerdict(t, repo.Dir, card, "verdict.md", "PASS", repo.Merge)
		head := commitFile(t, repo.Dir, ".moai/reports/"+card+"/extra.md")
		if _, err := readAuditVerdict(repo.Dir, card, "sync-audit", head); err != nil {
			t.Fatalf("reports-only offset refused: %v", err)
		}
	})

	t.Run("code offset refuses and names the path", func(t *testing.T) {
		head := commitFile(t, repo.Dir, "internal/code.go")
		_, err := readAuditVerdict(repo.Dir, card, "sync-audit", head)
		if !errors.Is(err, ErrEvidence) || !strings.Contains(err.Error(), "internal/code.go") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("audited commit not an ancestor refuses", func(t *testing.T) {
		frGit(t, repo.Dir, "checkout", "-q", "-b", "side", repo.Commit)
		side := commitFile(t, repo.Dir, "side.txt")
		frGit(t, repo.Dir, "checkout", "-q", frIntegration)
		frWriteVerdict(t, repo.Dir, card, "verdict.md", "PASS", side)
		_, err := readAuditVerdict(repo.Dir, card, "sync-audit", repo.Merge)
		if !errors.Is(err, ErrEvidence) || !strings.Contains(err.Error(), "not an ancestor") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("injected git runner failure refuses", func(t *testing.T) {
		frWriteVerdict(t, repo.Dir, card, "verdict.md", "PASS", repo.Commit)
		saved := evidenceGitRead
		t.Cleanup(func() { evidenceGitRead = saved })
		evidenceGitRead = func(context.Context, string, ...string) (string, error) {
			return "", errors.New("injected")
		}
		if _, err := readAuditVerdict(repo.Dir, card, "sync-audit", repo.Merge); !errors.Is(err, ErrEvidence) {
			t.Fatalf("err = %v", err)
		}
	})
}
