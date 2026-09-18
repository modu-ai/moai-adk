package mission

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testGitBinary resolves git from PATH so the fixtures run on every CI
// platform instead of only where a host-specific install path exists.
func testGitBinary(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git not found on PATH: %v", err)
	}
	return path
}

func gitFixtureRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	all := append([]string{"-C", dir}, args...)
	cmd := exec.Command(testGitBinary(t), all...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.invalid", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.invalid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestGitOwnerCommitExplicitPathsAndReadback(t *testing.T) {
	repo := t.TempDir()
	gitFixtureRun(t, repo, "init", "-q", "-b", "WT-card")
	gitFixtureRun(t, repo, "config", "user.name", "Test")
	gitFixtureRun(t, repo, "config", "user.email", "test@example.invalid")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("base"), 0600); err != nil {
		t.Fatal(err)
	}
	gitFixtureRun(t, repo, "add", "tracked.txt")
	gitFixtureRun(t, repo, "commit", "-qm", "base")
	if err := os.WriteFile(filepath.Join(repo, "tracked.txt"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "foreign.txt"), []byte("foreign"), 0600); err != nil {
		t.Fatal(err)
	}
	owner := GitOwnerAdapter{GitBinary: testGitBinary(t), Effect: GitEffect{Action: ActionCommit, Repository: repo, WorktreeBranch: "WT-card", ExplicitPaths: []string{"tracked.txt"}, CommitMessage: "feat: card"}}
	if applied, err := owner.Readback(context.Background(), "op-commit"); err != nil || applied {
		t.Fatalf("pre readback=%v err=%v", applied, err)
	}
	if err := owner.Apply(context.Background(), "op-commit"); err != nil {
		t.Fatal(err)
	}
	if applied, err := owner.Readback(context.Background(), "op-commit"); err != nil || !applied {
		t.Fatalf("post readback=%v err=%v", applied, err)
	}
	if got := gitFixtureRun(t, repo, "status", "--porcelain"); got != "?? foreign.txt" {
		t.Fatalf("explicit staging violated: %q", got)
	}
	if err := owner.Apply(context.Background(), "op-commit"); err == nil {
		t.Fatal("duplicate commit invocation allowed")
	}
}

func TestGitOwnerLocalDevelopNoFFLeaseAndCrashReadback(t *testing.T) {
	repo := t.TempDir()
	gitFixtureRun(t, repo, "init", "-q", "-b", "develop")
	gitFixtureRun(t, repo, "config", "user.name", "Test")
	gitFixtureRun(t, repo, "config", "user.email", "test@example.invalid")
	if err := os.WriteFile(filepath.Join(repo, "base"), []byte("base"), 0600); err != nil {
		t.Fatal(err)
	}
	gitFixtureRun(t, repo, "add", "base")
	gitFixtureRun(t, repo, "commit", "-qm", "base")
	base := gitFixtureRun(t, repo, "rev-parse", "HEAD")
	gitFixtureRun(t, repo, "checkout", "-qb", "WT-card")
	if err := os.WriteFile(filepath.Join(repo, "card"), []byte("card"), 0600); err != nil {
		t.Fatal(err)
	}
	gitFixtureRun(t, repo, "add", "card")
	gitFixtureRun(t, repo, "commit", "-qm", "card")
	card := gitFixtureRun(t, repo, "rev-parse", "HEAD")
	gitFixtureRun(t, repo, "checkout", "-q", "develop")
	lease := filepath.Join(repo, ".git", "moai-integration-lease.json")
	if err := WriteIntegrationLease(lease, IntegrationLease{SessionID: "018f4f4a-7b7c-7a11-8f4d-777777777777", BaseSHA: base}); err != nil {
		t.Fatal(err)
	}
	owner := GitOwnerAdapter{GitBinary: testGitBinary(t), Effect: GitEffect{Action: ActionLocalMerge, Repository: repo, IntegrationWorktree: repo, WorktreeBranch: "WT-card", CardSHA: card, BaseSHA: base, LeasePath: lease, SessionID: "018f4f4a-7b7c-7a11-8f4d-777777777777"}}
	if err := owner.Apply(context.Background(), "op-merge"); err != nil {
		t.Fatal(err)
	}
	if applied, err := owner.Readback(context.Background(), "op-merge"); err != nil || !applied {
		t.Fatalf("readback=%v err=%v", applied, err)
	}
	parents := strings.Fields(gitFixtureRun(t, repo, "show", "-s", "--format=%P", "HEAD"))
	if len(parents) != 2 {
		t.Fatalf("merge was not --no-ff: %v", parents)
	}
	if err := owner.Apply(context.Background(), "op-merge"); err == nil {
		t.Fatal("crash readback must prevent caller from applying twice")
	}
	bad := owner
	bad.Effect.LeasePath = filepath.Join(repo, "missing")
	if err := bad.Apply(context.Background(), "other"); err == nil {
		t.Fatal("missing integration lease allowed")
	}
}

func TestGitOwnerSafetyMutants(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	gitFixtureRun(t, repo, "init", "-q", "-b", "main")
	gitFixtureRun(t, repo, "config", "user.name", "Test")
	gitFixtureRun(t, repo, "config", "user.email", "test@example.invalid")
	if err := os.WriteFile(filepath.Join(repo, "x"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	for name, effect := range map[string]GitEffect{"wrong-branch": {Action: ActionCommit, Repository: repo, WorktreeBranch: "WT-card", ExplicitPaths: []string{"x"}, CommitMessage: "x"}, "unsafe-path": {Action: ActionCommit, Repository: repo, WorktreeBranch: "main", ExplicitPaths: []string{"../x"}, CommitMessage: "x"}, "empty-paths": {Action: ActionCommit, Repository: repo, WorktreeBranch: "main", CommitMessage: "x"}, "unsupported": {Action: ActionForcePush, Repository: repo}} {
		t.Run(name, func(t *testing.T) {
			owner := GitOwnerAdapter{GitBinary: testGitBinary(t), Effect: effect}
			if err := owner.Apply(ctx, "op"); err == nil {
				t.Fatal("unsafe effect allowed")
			}
		})
	}
	if _, _, err := (GitOwnerAdapter{Effect: GitEffect{Repository: "relative"}}).Snapshot(ctx); err == nil {
		t.Fatal("relative repo accepted")
	}
	if err := WriteIntegrationLease(filepath.Join(repo, "bad"), IntegrationLease{}); err == nil {
		t.Fatal("invalid lease written")
	}
	leasePath := filepath.Join(repo, ".git", "lease.json")
	session := "018f4f4a-7b7c-7a11-8f4d-999999999999"
	if err := WriteIntegrationLease(leasePath, IntegrationLease{SessionID: session, BaseSHA: "base"}); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(leasePath, 0644); err != nil {
		t.Fatal(err)
	}
	owner := GitOwnerAdapter{Effect: GitEffect{Action: ActionLocalMerge, Repository: repo, IntegrationWorktree: repo, WorktreeBranch: "WT-card", CardSHA: "card", BaseSHA: "base", LeasePath: leasePath, SessionID: session}}
	if err := owner.ValidateIntegrationLease(); err == nil {
		t.Fatal("weak lease mode accepted")
	}
}

func TestGitOwnerReadbackLeaseAndSnapshotRefusals(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	gitFixtureRun(t, repo, "init", "-q", "-b", "develop")
	gitFixtureRun(t, repo, "config", "user.name", "Test")
	gitFixtureRun(t, repo, "config", "user.email", "test@example.invalid")
	if err := os.WriteFile(filepath.Join(repo, "base"), []byte("base"), 0600); err != nil {
		t.Fatal(err)
	}
	gitFixtureRun(t, repo, "add", "base")
	gitFixtureRun(t, repo, "commit", "-qm", "base")
	base := gitFixtureRun(t, repo, "rev-parse", "HEAD")
	owner := GitOwnerAdapter{GitBinary: testGitBinary(t), Effect: GitEffect{Repository: repo}}
	if head, branch, err := owner.Snapshot(ctx); err != nil || head != base || branch != "develop" {
		t.Fatalf("snapshot head=%q branch=%q err=%v", head, branch, err)
	}
	if applied, err := owner.Readback(ctx, "op"); err == nil || applied {
		t.Fatalf("unsupported readback=%v err=%v", applied, err)
	}
	owner.Effect.Action = ActionLocalMerge
	if applied, err := owner.Readback(ctx, "op"); err == nil || applied {
		t.Fatalf("missing sha readback=%v err=%v", applied, err)
	}

	session := "018f4f4a-7b7c-7a11-8f4d-888888888888"
	lease := filepath.Join(repo, ".git", "lease.json")
	if err := WriteIntegrationLease(lease, IntegrationLease{SessionID: session, BaseSHA: base}); err != nil {
		t.Fatal(err)
	}
	if err := WriteIntegrationLease(filepath.Join(repo, ".git", "lease-link"), IntegrationLease{SessionID: session, BaseSHA: base}); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(repo, ".git", "lease-symlink")
	if err := os.Symlink(lease, link); err != nil {
		t.Fatal(err)
	}
	if err := WriteIntegrationLease(link, IntegrationLease{SessionID: session, BaseSHA: base}); err == nil {
		t.Fatal("symlink lease overwritten")
	}
	baseEffect := GitEffect{Action: ActionLocalMerge, Repository: repo, IntegrationWorktree: repo, WorktreeBranch: "WT-card", CardSHA: base, BaseSHA: base, LeasePath: lease, SessionID: session}
	wrongSession := GitOwnerAdapter{GitBinary: owner.GitBinary, Effect: baseEffect}
	wrongSession.Effect.SessionID = "018f4f4a-7b7c-7a11-8f4d-999999999999"
	if err := wrongSession.ValidateIntegrationLease(); err == nil {
		t.Fatal("lease session mismatch accepted")
	}
	outside := GitOwnerAdapter{GitBinary: owner.GitBinary, Effect: baseEffect}
	outside.Effect.LeasePath = filepath.Join(t.TempDir(), "lease")
	if err := os.WriteFile(outside.Effect.LeasePath, []byte(`{"session_id":"`+session+`","base_sha":"`+base+`"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := outside.ValidateIntegrationLease(); err == nil {
		t.Fatal("outside lease accepted")
	}
	stale := GitOwnerAdapter{GitBinary: owner.GitBinary, Effect: baseEffect}
	stale.Effect.BaseSHA = strings.Repeat("0", 40)
	if err := WriteIntegrationLease(lease, IntegrationLease{SessionID: session, BaseSHA: stale.Effect.BaseSHA}); err != nil {
		t.Fatal(err)
	}
	if err := stale.Apply(ctx, "stale-op"); err == nil || !strings.Contains(err.Error(), "stale integration base") {
		t.Fatalf("stale base err=%v", err)
	}
	if err := ValidateMissionSessionID("NOT-A-UUID"); err == nil {
		t.Fatal("invalid session accepted")
	}
}
