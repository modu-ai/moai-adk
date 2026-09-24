//go:build !windows

// codex_audit_confine_test.go — write-time confinement of the launcher's two
// writes (the verdict and the launch record). Each test reproduces one
// sync-audit finding: a report directory swapped for an outside symlink while
// the audit runs (F1), an upper-case alias of the record directory or of
// `.git` (F2), and a record directory that is already a symlink (F3).
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// auditFilesUnder lists every non-directory entry below dir.
func auditFilesUnder(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			out = append(out, p)
		}
		return nil
	})
	return out
}

// auditReadRecord parses the launch record named by the single LAUNCH_RECORD line.
func auditReadRecord(t *testing.T, root, stderr string) map[string]any {
	t.Helper()
	lines := launchRecordLines(stderr)
	if len(lines) != 1 {
		t.Fatalf("want one LAUNCH_RECORD line, got %q", stderr)
	}
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(lines[0])))
	if err != nil {
		t.Fatalf("launch record unreadable: %v", err)
	}
	var rec map[string]any
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("launch record is not JSON: %v", err)
	}
	return rec
}

// F1: a middle component of the destination is replaced by a symlink to a
// directory outside the worktree after validation and before the write.
func TestCodexAuditLaunchWriteRaceConfined(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	msg := "RACED VERDICT\n"
	fake.setExec(msg, 0)
	outside := filepath.Join(repo.base, "outside")
	mid := filepath.Join(repo.a1, ".moai", "reports", "x")
	for _, d := range []string{outside, filepath.Join(mid, "y")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	dest := filepath.Join(mid, "y", "v.md")
	fake.write(t, "swap.dir", mid)
	fake.write(t, "swap.to", outside)

	r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: dest})

	if fi, err := os.Lstat(mid); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the race was not created: %s is not a symlink (%v)", mid, err)
	}
	if n := len(fake.execCalls(t)); n != 1 {
		t.Fatalf("audit process started %d times, want 1 (the swap happens while it runs)", n)
	}
	if escaped := auditFilesUnder(t, outside); len(escaped) != 0 {
		t.Fatalf("verdict escaped the worktree through the swapped component: %v", escaped)
	}
	rec := auditReadRecord(t, repo.a1, r.stderr)
	vp, _ := rec["verdict_path"].(string)
	if r.res.ExitCode == 0 {
		// A confined write is acceptable only when the recorded path is the
		// real, symlink-free path that holds the returned text.
		abs := filepath.Join(repo.a1, filepath.FromSlash(vp))
		real, err := filepath.EvalSymlinks(abs)
		if vp == "" || err != nil || real != abs {
			t.Fatalf("success recorded a path that is not the real written path: %q (%v)", vp, err)
		}
		if got, _ := os.ReadFile(abs); string(got) != msg {
			t.Fatalf("recorded path does not hold the returned text: %q", got)
		}
		return
	}
	if vp != "" {
		t.Fatalf("failed write still recorded verdict_path %q", vp)
	}
	if fr, _ := rec["failure_reason"].(string); strings.TrimSpace(fr) == "" {
		t.Fatalf("failed write recorded no failure_reason: %v", rec)
	}
}

// F2: the record directory and `.git` are matched case-insensitively.
func TestCodexAuditLaunchRecordAliasRefused(t *testing.T) {
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	fake.setExec("genuine verdict\n", 0)
	legal := filepath.Join(repo.a1, ".moai", "reports", "x", "v.md")
	first := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: legal})
	if first.res.ExitCode != 0 {
		t.Fatalf("genuine audit failed: %s", first.stderr)
	}
	recRel := launchRecordLines(first.stderr)[0]
	recAbs := filepath.Join(repo.a1, filepath.FromSlash(recRel))
	recSHA := auditFileSHA(t, recAbs)
	reports := filepath.Join(repo.a1, ".moai", "reports")

	fake.setExec("FORGED-RECORD", 0)
	dests := map[string]string{
		"CODEX-AUDIT record alias": filepath.Join(reports, "CODEX-AUDIT", filepath.Base(recAbs)),
		"Codex-Audit record alias": filepath.Join(reports, "Codex-Audit", filepath.Base(recAbs)),
		".GIT component":           filepath.Join(reports, "x", ".GIT", "v.md"),
		".Git component":           filepath.Join(reports, ".Git", "v.md"),
	}
	for name, dest := range dests {
		t.Run(name, func(t *testing.T) {
			before := len(fake.calls(t))
			snap := auditSnapshotTree(t, repo.base)
			r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1, Out: dest})
			if r.res.ExitCode == 0 {
				t.Fatalf("alias destination accepted: %s", dest)
			}
			if n := len(fake.calls(t)) - before; n != 0 {
				t.Fatalf("alias destination called codex %d times", n)
			}
			if got := auditFileSHA(t, recAbs); got != recSHA {
				t.Fatalf("launch record %s was overwritten", recRel)
			}
			if d := auditDiffSnapshots(snap, auditSnapshotTree(t, repo.base)); len(d) > 0 {
				t.Fatalf("refused alias changed files: %v", d)
			}
		})
	}
}

// F3: the launch record never follows a symlinked record or report directory.
func TestCodexAuditLaunchSymlinkedRecordDir(t *testing.T) {
	cases := map[string]struct {
		link string // path under the root replaced by a symlink to outside
		out  bool   // name a destination (true) or return on stdout (false)
	}{
		"record dir symlink":          {link: ".moai/reports/codex-audit", out: true},
		"reports dir symlink, stdout": {link: ".moai/reports", out: false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			repo := newAuditRepo(t)
			fake := installFakeCodex(t)
			outside := filepath.Join(repo.base, "outside")
			if err := os.MkdirAll(outside, 0o755); err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(repo.a1, filepath.FromSlash(c.link))
			if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, link); err != nil {
				t.Fatal(err)
			}
			req := codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, CallerDir: repo.a1, Root: repo.a1}
			if c.out {
				req.Out = filepath.Join(repo.a1, ".moai", "reports", "x", "v.md")
			}
			r := runAudit(t, req)
			if written := auditFilesUnder(t, outside); len(written) != 0 {
				t.Fatalf("launcher wrote through the symlinked directory: %v", written)
			}
			if r.res.ExitCode == 0 {
				t.Fatal("launch through a symlinked record directory succeeded")
			}
			if n := len(fake.execCalls(t)); n != 0 {
				t.Fatalf("audit process started %d times despite the refused record", n)
			}
		})
	}
}
