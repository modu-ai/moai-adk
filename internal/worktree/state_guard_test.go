package worktree

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

// initTestRepo creates a temporary git repo with an initial commit on branch main.
// Tests must NEVER modify the host repository — all git operations happen in tmpDir.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runOrFail(t, dir, "git", "init", "-q", "-b", "main")
	runOrFail(t, dir, "git", "config", "user.email", "test@example.com")
	runOrFail(t, dir, "git", "config", "user.name", "Test User")
	runOrFail(t, dir, "git", "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("initial\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runOrFail(t, dir, "git", "add", "README.md")
	runOrFail(t, dir, "git", "commit", "-q", "-m", "initial")
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "specs", "SPEC-X"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runOrFail(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v in %s: %v\n%s", name, args, dir, err, out)
	}
}

func TestSnapshot_NoDiff(t *testing.T) {
	repo := initTestRepo(t)
	ctx := context.Background()
	pre, err := Capture(ctx, CaptureOptions{RepoDir: repo})
	if err != nil {
		t.Fatalf("Capture pre: %v", err)
	}
	post, err := Capture(ctx, CaptureOptions{RepoDir: repo})
	if err != nil {
		t.Fatalf("Capture post: %v", err)
	}
	div := Diff(pre, post)
	if div.IsDivergent() {
		t.Errorf("expected no divergence, got: %+v", div)
	}
	if pre.HeadSHA == "" {
		t.Errorf("expected pre.HeadSHA non-empty (initial commit exists)")
	}
	if pre.Branch != "main" {
		t.Errorf("expected pre.Branch=main, got %q", pre.Branch)
	}
}

func TestSnapshot_UntrackedAdded(t *testing.T) {
	repo := initTestRepo(t)
	ctx := context.Background()
	pre, err := Capture(ctx, CaptureOptions{RepoDir: repo})
	if err != nil {
		t.Fatalf("Capture pre: %v", err)
	}
	newFile := filepath.Join(repo, ".moai", "specs", "SPEC-X", "new.md")
	if err := os.WriteFile(newFile, []byte("new content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	post, err := Capture(ctx, CaptureOptions{RepoDir: repo})
	if err != nil {
		t.Fatalf("Capture post: %v", err)
	}
	div := Diff(pre, post)
	if !div.IsDivergent() {
		t.Errorf("expected divergence (untracked added), got none: %+v", div)
	}
	if !slices.Contains(div.UntrackedAdded, ".moai/specs/SPEC-X/new.md") {
		t.Errorf("expected UntrackedAdded to contain .moai/specs/SPEC-X/new.md, got: %v", div.UntrackedAdded)
	}
}

func TestSnapshot_UntrackedRemoved(t *testing.T) {
	repo := initTestRepo(t)
	ctx := context.Background()
	extraFile := filepath.Join(repo, ".moai", "specs", "SPEC-X", "extra.md")
	if err := os.WriteFile(extraFile, []byte("extra\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pre, err := Capture(ctx, CaptureOptions{RepoDir: repo})
	if err != nil {
		t.Fatalf("Capture pre: %v", err)
	}
	if err := os.Remove(extraFile); err != nil {
		t.Fatal(err)
	}
	post, err := Capture(ctx, CaptureOptions{RepoDir: repo})
	if err != nil {
		t.Fatalf("Capture post: %v", err)
	}
	div := Diff(pre, post)
	if !div.IsDivergent() {
		t.Errorf("expected divergence (untracked removed), got none: %+v", div)
	}
	if len(div.UntrackedRemoved) == 0 {
		t.Errorf("expected UntrackedRemoved non-empty, got: %+v", div.UntrackedRemoved)
	}
}

func TestSnapshot_BranchChanged(t *testing.T) {
	repo := initTestRepo(t)
	ctx := context.Background()
	pre, err := Capture(ctx, CaptureOptions{RepoDir: repo})
	if err != nil {
		t.Fatalf("Capture pre: %v", err)
	}
	runOrFail(t, repo, "git", "checkout", "-q", "-b", "wave-5-test")
	post, err := Capture(ctx, CaptureOptions{RepoDir: repo})
	if err != nil {
		t.Fatalf("Capture post: %v", err)
	}
	div := Diff(pre, post)
	if !div.IsDivergent() {
		t.Errorf("expected divergence (branch changed), got none: %+v", div)
	}
	if !div.BranchChanged {
		t.Errorf("expected BranchChanged=true, got false; pre=%s post=%s", pre.Branch, post.Branch)
	}
}

func TestDivergence_IsDivergent_AllFalse(t *testing.T) {
	d := Divergence{}
	if d.IsDivergent() {
		t.Error("expected IsDivergent=false for zero-value Divergence")
	}
}

func TestDivergence_IsDivergent_HeadChanged(t *testing.T) {
	d := Divergence{HeadChanged: true}
	if !d.IsDivergent() {
		t.Error("expected IsDivergent=true when HeadChanged")
	}
}
