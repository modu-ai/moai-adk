package hook

// cwd_changed_relocate_anchor_test.go — t1058 / SPEC-SESSION-REGISTRY-READ-ANCHOR-001
// scope S1 (the R2 repair).
//
// The defect: relocateSessionCwd walked up from each candidate working
// directory and stopped at the FIRST .moai/state/active-sessions.json it
// found. When that file did not carry the session, the loop advanced to the
// next CANDIDATE WORKING DIRECTORY rather than to the parent directory — so
// where every candidate sits inside one linked worktree carrying its own
// orphan registry, there was structurally no path to the repository's primary
// registry and the relocation silently did nothing.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// --- fixture helpers -------------------------------------------------------

func requireGitHook(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func runGitHook(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

// canonicalHook resolves symlinks so comparisons hold on macOS, where
// t.TempDir() hands back /var/... while git reports /private/var/....
func canonicalHook(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", path, err)
	}
	return resolved
}

// newRepoWithWorktree builds a git repository with one linked worktree and
// returns both canonical roots.
func newRepoWithWorktree(t *testing.T) (primary, worktree string) {
	t.Helper()
	requireGitHook(t)

	primary = canonicalHook(t, t.TempDir())
	runGitHook(t, primary, "init", "-q", "-b", "main")
	runGitHook(t, primary, "-c", "user.email=t@example.com", "-c", "user.name=t",
		"commit", "-q", "--allow-empty", "-m", "seed")

	worktree = filepath.Join(canonicalHook(t, t.TempDir()), "wt")
	runGitHook(t, primary, "worktree", "add", "-q", "-b", "wt-branch", worktree)
	return primary, worktree
}

// writeRegistryEntries seeds <root>/.moai/state/active-sessions.json with the
// given entries verbatim and returns the file path.
func writeRegistryEntries(t *testing.T, root string, entries []session.Entry) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	path := filepath.Join(dir, "active-sessions.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	return path
}

func entryFor(t *testing.T, sessionID, cwd string) session.Entry {
	t.Helper()
	host, _ := os.Hostname()
	now := time.Now().UTC()
	return session.Entry{
		SessionID:     sessionID,
		SpecID:        "(none)",
		Phase:         "(none)",
		StartedAt:     now,
		LastHeartbeat: now,
		PID:           os.Getpid(),
		Host:          host,
		CWD:           cwd,
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

// cwdOf returns the recorded CWD of sessionID in the registry file at path.
func cwdOf(t *testing.T, path, sessionID string) string {
	t.Helper()
	var entries []session.Entry
	if err := json.Unmarshal(mustReadFile(t, path), &entries); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	for _, e := range entries {
		if e.SessionID == sessionID {
			return e.CWD
		}
	}
	t.Fatalf("session %s absent from %s", sessionID, path)
	return ""
}

// --- AC-RAR-002 ------------------------------------------------------------

// TestRelocateReachesPrimaryRegistryWhenEveryCandidateIsInsideOneWorktree is
// the load-bearing case: the worktree carries its own orphan registry that
// does NOT hold the session, every candidate working directory sits inside
// that worktree, and the session's entry lives only in the repository's
// primary registry. The relocation must still reach the primary registry, and
// the worktree-local file must be byte-unchanged.
//
// Mutation control (recorded in the run-phase evidence): reverting
// relocateSessionCwd to the first-file-found walk makes this test fail.
func TestRelocateReachesPrimaryRegistryWhenEveryCandidateIsInsideOneWorktree(t *testing.T) {
	primary, worktree := newRepoWithWorktree(t)
	const sid = "sess-t1058-anchored"

	primaryReg := writeRegistryEntries(t, primary, []session.Entry{entryFor(t, sid, primary)})
	// The orphan file: present, readable, and about a DIFFERENT session.
	orphanReg := writeRegistryEntries(t, worktree, []session.Entry{entryFor(t, "sess-other", worktree)})
	orphanBefore := mustReadFile(t, orphanReg)

	inner := mustMkdirAllHook(t, filepath.Join(worktree, "internal", "pkg"))

	h := NewCwdChangedHandler()
	input := &HookInput{SessionID: sid, CWD: inner, OldCwd: worktree, NewCwd: inner}
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if got := cwdOf(t, primaryReg, sid); got != inner {
		t.Errorf("primary registry CWD = %q, want %q (the relocation never reached the primary registry)", got, inner)
	}
	if got := mustReadFile(t, orphanReg); string(got) != string(orphanBefore) {
		t.Errorf("worktree-local registry was modified:\nbefore: %s\nafter:  %s", orphanBefore, got)
	}
}

// --- AC-RAR-003 ------------------------------------------------------------

// TestRelocateFailOpen pins the three fail-open cases: an absent registry, an
// unreadable one, and one carrying no matching session. In each case no
// registry file changes and the hook returns no error.
func TestRelocateFailOpen(t *testing.T) {
	t.Run("absent registry creates nothing", func(t *testing.T) {
		root := canonicalHook(t, t.TempDir()) // not a git repository, no .moai anywhere
		inner := mustMkdirAllHook(t, filepath.Join(root, "nested"))

		h := NewCwdChangedHandler()
		input := &HookInput{SessionID: "sess-t1058-absent", CWD: inner, OldCwd: root, NewCwd: inner}
		if _, err := h.Handle(context.Background(), input); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		// Mutation control: dropping the not-found guard makes RelocateSession
		// run against an absent file, and withLock CREATES it.
		if _, err := os.Stat(filepath.Join(root, session.DefaultRegistryPath)); !os.IsNotExist(err) {
			t.Errorf("a registry file was created at %s (stat err = %v)", root, err)
		}
	})

	t.Run("unreadable registry is skipped, not claimed", func(t *testing.T) {
		primary, worktree := newRepoWithWorktree(t)
		const sid = "sess-t1058-unreadable"

		primaryReg := writeRegistryEntries(t, primary, []session.Entry{entryFor(t, sid, primary)})
		corruptDir := mustMkdirAllHook(t, filepath.Join(worktree, ".moai", "state"))
		corruptReg := filepath.Join(corruptDir, "active-sessions.json")
		if err := os.WriteFile(corruptReg, []byte("{not json"), 0o644); err != nil {
			t.Fatalf("write corrupt registry: %v", err)
		}
		corruptBefore := mustReadFile(t, corruptReg)

		inner := mustMkdirAllHook(t, filepath.Join(worktree, "sub"))
		h := NewCwdChangedHandler()
		input := &HookInput{SessionID: sid, CWD: inner, OldCwd: worktree, NewCwd: inner}
		if _, err := h.Handle(context.Background(), input); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		if got := mustReadFile(t, corruptReg); string(got) != string(corruptBefore) {
			t.Errorf("unreadable registry was rewritten: %s", got)
		}
		// Mutation control: dropping the fail-open guards (the read-error
		// `continue` AND the not-found `continue`) makes the walk claim the
		// unreadable file and return, so the primary registry is never reached.
		// The read-error guard alone is not observable here — with the
		// not-found guard still standing, a failed read yields no entries and
		// the walk continues anyway.
		if got := cwdOf(t, primaryReg, sid); got != inner {
			t.Errorf("primary registry CWD = %q, want %q", got, inner)
		}
	})

	t.Run("entry-less registries are left byte-identical", func(t *testing.T) {
		primary, worktree := newRepoWithWorktree(t)

		primaryReg := writeRegistryEntries(t, primary, []session.Entry{entryFor(t, "sess-someone-else", primary)})
		orphanReg := writeRegistryEntries(t, worktree, []session.Entry{})
		primaryBefore := mustReadFile(t, primaryReg)
		orphanBefore := mustReadFile(t, orphanReg)

		inner := mustMkdirAllHook(t, filepath.Join(worktree, "sub"))
		h := NewCwdChangedHandler()
		input := &HookInput{SessionID: "sess-t1058-nomatch", CWD: inner, OldCwd: worktree, NewCwd: inner}
		if _, err := h.Handle(context.Background(), input); err != nil {
			t.Fatalf("Handle() error = %v", err)
		}

		// Mutation control: dropping the not-found guard rewrites both files
		// through withLock's MarshalIndent, which is not byte-identical.
		if got := mustReadFile(t, primaryReg); string(got) != string(primaryBefore) {
			t.Errorf("primary registry rewritten:\nbefore: %s\nafter:  %s", primaryBefore, got)
		}
		if got := mustReadFile(t, orphanReg); string(got) != string(orphanBefore) {
			t.Errorf("worktree registry rewritten:\nbefore: %s\nafter:  %s", orphanBefore, got)
		}
	})
}
