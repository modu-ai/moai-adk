package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

func migrationGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func migrationFixture(t *testing.T) (string, string, string) {
	t.Helper()
	previousCWDProbe := activeProcessCWDs
	activeProcessCWDs = func() ([]string, error) { return nil, nil }
	t.Cleanup(func() { activeProcessCWDs = previousCWDProbe })
	root := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	migrationGit(t, root, "init", "-q")
	migrationGit(t, root, "config", "user.name", "Migration Test")
	migrationGit(t, root, "config", "user.email", "migration-test@example.com")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored.txt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tracked.txt"), []byte("initial\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	migrationGit(t, root, "add", ".gitignore", "tracked.txt")
	migrationGit(t, root, "commit", "-q", "-m", "seed")
	old := filepath.Join(root, ".claude", "worktrees", "card")
	modern := filepath.Join(root, ".moai", "worktrees", "card")
	migrationGit(t, root, "worktree", "add", "-q", "-b", "WT-migrate", old)
	return root, old, modern
}

func TestLegacyWorktreeMigrationSkipsActiveProcessWithoutRegistry(t *testing.T) {
	root, old, _ := migrationFixture(t)
	if err := os.Mkdir(filepath.Join(old, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	activeProcessCWDs = func() ([]string, error) { return []string{filepath.Join(old, "nested")}, nil }
	plans, err := planLegacyWorktreeMigration(root, time.Now())
	if err != nil || len(plans) != 1 || plans[0].skip != "an active process is inside this worktree" {
		t.Fatalf("active process was not protected: %#v, %v", plans, err)
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatalf("source lost: %v", err)
	}
}

func TestLegacyWorktreeMigrationFailsClosedWhenProcessProbeFails(t *testing.T) {
	root, old, _ := migrationFixture(t)
	activeProcessCWDs = func() ([]string, error) { return nil, os.ErrPermission }
	plans, err := planLegacyWorktreeMigration(root, time.Now())
	if err != nil || len(plans) != 1 || !strings.Contains(plans[0].skip, "cannot inspect active process") {
		t.Fatalf("unmeasured process state was not protected: %#v, %v", plans, err)
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatalf("source lost: %v", err)
	}
}

func TestLegacyWorktreeMigrationPreservesGitAndWorkingFiles(t *testing.T) {
	root, old, modern := migrationFixture(t)
	for name, body := range map[string]string{
		"tracked.txt": "changed\n", "untracked.txt": "untracked\n", "ignored.txt": "ignored\n",
	} {
		if err := os.WriteFile(filepath.Join(old, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	before, err := gitForWorktreeMigration(old, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		t.Fatal(err)
	}
	plans, err := planLegacyWorktreeMigration(root, time.Now())
	if err != nil || len(plans) != 1 || plans[0].skip != "" {
		t.Fatalf("plan = %#v, %v", plans, err)
	}
	if err := moveLegacyWorktree(root, plans[0]); err != nil {
		t.Fatalf("move: %v", err)
	}
	if _, err := os.Lstat(old); !os.IsNotExist(err) {
		t.Fatalf("old tree still exists: %v", err)
	}
	if got := migrationGit(t, modern, "branch", "--show-current"); got != "WT-migrate" {
		t.Fatalf("branch = %q", got)
	}
	after, err := gitForWorktreeMigration(modern, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil || after != before {
		t.Fatalf("status changed: before %q, after %q, err %v", before, after, err)
	}
	for name, want := range map[string]string{
		"tracked.txt": "changed\n", "untracked.txt": "untracked\n", "ignored.txt": "ignored\n",
	} {
		got, err := os.ReadFile(filepath.Join(modern, name))
		if err != nil || string(got) != want {
			t.Fatalf("%s = %q, %v", name, got, err)
		}
	}
	if again, err := planLegacyWorktreeMigration(root, time.Now()); err != nil || len(again) != 0 {
		t.Fatalf("rerun must be idempotent: %#v, %v", again, err)
	}
}

func TestLegacyWorktreeMigrationSkipsLockAndCollision(t *testing.T) {
	for _, reason := range []string{"locked", "destination collision"} {
		t.Run(reason, func(t *testing.T) {
			root, old, modern := migrationFixture(t)
			if reason == "locked" {
				migrationGit(t, root, "worktree", "lock", "--reason", "active", old)
			} else if err := os.MkdirAll(modern, 0o755); err != nil {
				t.Fatal(err)
			}
			plans, err := planLegacyWorktreeMigration(root, time.Now())
			if err != nil || len(plans) != 1 || plans[0].skip == "" {
				t.Fatalf("unsafe tree not skipped: %#v, %v", plans, err)
			}
			if err := moveLegacyWorktree(root, plans[0]); err == nil {
				t.Fatal("skipped tree moved")
			}
			if _, err := os.Stat(old); err != nil {
				t.Fatalf("source lost: %v", err)
			}
		})
	}
}

func TestLegacyWorktreeMigrationRechecksOnlySelectedTree(t *testing.T) {
	root, old, _ := migrationFixture(t)
	other := filepath.Join(root, ".claude", "worktrees", "other")
	migrationGit(t, root, "worktree", "add", "-q", "-b", "WT-other", other)
	plans, err := planLegacyWorktreeMigration(root, time.Now())
	if err != nil || len(plans) != 2 {
		t.Fatalf("full migration plan = %#v, %v", plans, err)
	}
	migrationGit(t, root, "worktree", "lock", other)
	focused, err := planLegacyWorktreeMigrationFor(root, time.Now(), old, nil)
	if err != nil || len(focused) != 1 || canonicalTreePath(focused[0].source) != canonicalTreePath(old) || focused[0].skip != "" {
		t.Fatalf("selected tree recheck = %#v, %v", focused, err)
	}
	migrationGit(t, root, "worktree", "lock", old)
	if err := moveLegacyWorktree(root, focused[0]); err == nil || !strings.Contains(err.Error(), "changed since migration planning") {
		t.Fatalf("lock added after planning was not detected: %v", err)
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatalf("source lost after failed recheck: %v", err)
	}
}

func TestUpdateWorktreeMigrationDryRunAndRetry(t *testing.T) {
	root, old, modern := migrationFixture(t)
	var output bytes.Buffer
	if err := runUpdateWorktreeMigration(root, true, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Worktree migration planned") {
		t.Fatalf("dry run did not report plan: %q", output.String())
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatalf("dry run moved source: %v", err)
	}
	output.Reset()
	if err := runUpdateWorktreeMigration(root, false, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "1 legacy worktree detected") ||
		!strings.Contains(output.String(), "Worktree migration safety check: 1/1 inspected") {
		t.Fatalf("live migration did not announce its size: %q", output.String())
	}
	if !strings.Contains(output.String(), "1 moved, 0 skipped, 0 failed") {
		t.Fatalf("update did not move tree: %q", output.String())
	}
	if _, err := os.Stat(modern); err != nil {
		t.Fatalf("destination missing: %v", err)
	}
	output.Reset()
	if err := runUpdateWorktreeMigration(root, false, &output); err != nil || output.Len() != 0 {
		t.Fatalf("idempotent retry: %v, %q", err, output.String())
	}
}

func TestUpdateWorktreeMigrationReportsProgress(t *testing.T) {
	root, _, _ := migrationFixture(t)
	for _, name := range []string{"two", "three", "four", "five"} {
		old := filepath.Join(root, ".claude", "worktrees", name)
		migrationGit(t, root, "worktree", "add", "-q", "-b", "WT-"+name, old)
	}
	var output bytes.Buffer
	if err := runUpdateWorktreeMigration(root, false, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "5 legacy worktrees detected") ||
		!strings.Contains(output.String(), "Worktree migration safety check: 5/5 inspected") ||
		!strings.Contains(output.String(), "Worktree migration progress: 5/5 checked, 5 moved, 0 skipped, 0 failed") {
		t.Fatalf("live migration progress missing: %q", output.String())
	}
}

func TestLegacyWorktreeMigrationSkipsActiveSession(t *testing.T) {
	root, old, _ := migrationFixture(t)
	registry := session.NewRegistry(filepath.Join(root, session.DefaultRegistryPath), nil)
	if err := registry.Register("active-session", session.SpecIDNone, session.PhaseNone); err != nil {
		t.Fatal(err)
	}
	if err := registry.RelocateSession("active-session", old); err != nil {
		t.Fatal(err)
	}
	plans, err := planLegacyWorktreeMigration(root, time.Now())
	if err != nil || len(plans) != 1 || !strings.Contains(plans[0].skip, "active session") {
		t.Fatalf("active tree was not skipped: %#v, %v", plans, err)
	}
	var output bytes.Buffer
	if err := runUpdateWorktreeMigration(root, false, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "1 skipped") {
		t.Fatalf("active tree skip not reported: %q", output.String())
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatalf("active source lost: %v", err)
	}
}

func TestRunUpdateMigratesRegisteredLegacyWorktree(t *testing.T) {
	root, old, modern := migrationFixture(t)
	writeV3ProjectFixture(t, root)
	output := runUpdateInFixture(t, root)
	if !strings.Contains(output, "Worktree migration: 1 moved, 0 skipped, 0 failed") {
		t.Fatalf("moai update did not report migration: %q", output)
	}
	if _, err := os.Lstat(old); !os.IsNotExist(err) {
		t.Fatalf("legacy location remains after update: %v", err)
	}
	if got := migrationGit(t, modern, "branch", "--show-current"); got != "WT-migrate" {
		t.Fatalf("migrated branch = %q", got)
	}
}
