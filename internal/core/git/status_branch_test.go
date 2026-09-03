package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The header shapes below are asserted against REAL git output from fixture
// repositories, not against a hand-written list: TestStatusBranchHeaderShapes
// drives git into each state and checks the header it actually prints, and
// TestParseStatusBranchHeader pins the parser against the same strings. A
// parser agreeing with an imagined format would prove nothing.

// gitFixture runs git in dir, failing the test on error.
func gitFixture(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"LC_ALL=C",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimRight(string(out), "\n")
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mkFixtureDir(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	return path
}

// newFixtureRepo builds a repository with one commit under t.TempDir().
func newFixtureRepo(t *testing.T) string {
	t.Helper()
	dir := mkFixtureDir(t, filepath.Join(t.TempDir(), "repo"))
	gitFixture(t, dir, "init", "-q", "-b", "main")
	writeFixtureFile(t, filepath.Join(dir, "tracked.txt"), "one\n")
	gitFixture(t, dir, "add", "tracked.txt")
	gitFixture(t, dir, "commit", "-qm", "one")
	return dir
}

// statusHeader returns the `## ` header line git prints for dir.
func statusHeader(t *testing.T, dir string) string {
	t.Helper()
	out := gitFixture(t, dir, "status", "--porcelain", "--branch")
	line, _, found := strings.Cut(out, "\n")
	_ = found // a single-line status has no separator; the line IS the header.
	if !strings.HasPrefix(line, "## ") {
		t.Fatalf("first status line is not a header: %q", line)
	}
	return line
}

// TestStatusBranchHeaderShapes records the header shape git actually emits in
// each repository state, so the parser table below is anchored to observed
// output rather than to an assumption about the format.
func TestStatusBranchHeaderShapes(t *testing.T) {
	base := t.TempDir()
	repo := mkFixtureDir(t, filepath.Join(base, "repo"))
	gitFixture(t, repo, "init", "-q", "-b", "main")

	if got, want := statusHeader(t, repo), "## No commits yet on main"; got != want {
		t.Errorf("fresh repo header = %q, want %q", got, want)
	}

	writeFixtureFile(t, filepath.Join(repo, "tracked.txt"), "one\n")
	gitFixture(t, repo, "add", "tracked.txt")
	gitFixture(t, repo, "commit", "-qm", "one")
	if got, want := statusHeader(t, repo), "## main"; got != want {
		t.Errorf("no-upstream header = %q, want %q", got, want)
	}

	gitFixture(t, repo, "checkout", "-q", "--detach", "HEAD")
	if got, want := statusHeader(t, repo), "## HEAD (no branch)"; got != want {
		t.Errorf("detached header = %q, want %q", got, want)
	}
	gitFixture(t, repo, "checkout", "-q", "main")

	remote := filepath.Join(base, "remote.git")
	gitFixture(t, base, "init", "-q", "--bare", remote)
	gitFixture(t, repo, "remote", "add", "origin", remote)
	gitFixture(t, repo, "push", "-q", "-u", "origin", "main")
	if got, want := statusHeader(t, repo), "## main...origin/main"; got != want {
		t.Errorf("in-sync header = %q, want %q", got, want)
	}

	writeFixtureFile(t, filepath.Join(repo, "second.txt"), "two\n")
	gitFixture(t, repo, "add", "second.txt")
	gitFixture(t, repo, "commit", "-qm", "two")
	if got, want := statusHeader(t, repo), "## main...origin/main [ahead 1]"; got != want {
		t.Errorf("ahead header = %q, want %q", got, want)
	}

	// A clone that pushes ahead of the remote makes the first repo diverge.
	clone := filepath.Join(base, "clone")
	gitFixture(t, base, "clone", "-q", remote, clone)
	writeFixtureFile(t, filepath.Join(clone, "third.txt"), "three\n")
	gitFixture(t, clone, "add", "third.txt")
	gitFixture(t, clone, "commit", "-qm", "three")
	gitFixture(t, clone, "push", "-q")
	gitFixture(t, repo, "fetch", "-q")
	if got, want := statusHeader(t, repo), "## main...origin/main [ahead 1, behind 1]"; got != want {
		t.Errorf("diverged header = %q, want %q", got, want)
	}

	// A clone rewound one commit is behind and nothing else.
	behind := filepath.Join(base, "behind")
	gitFixture(t, base, "clone", "-q", remote, behind)
	gitFixture(t, behind, "reset", "-q", "--hard", "HEAD~1")
	if got, want := statusHeader(t, behind), "## main...origin/main [behind 1]"; got != want {
		t.Errorf("behind header = %q, want %q", got, want)
	}
}

func TestParseStatusBranchHeader(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantBranch string
		wantAhead  int
		wantBehind int
	}{
		{"diverged", "## main...origin/main [ahead 1, behind 2]", "main", 1, 2},
		{"ahead only", "## main...origin/main [ahead 3]", "main", 3, 0},
		{"behind only", "## main...origin/main [behind 4]", "main", 0, 4},
		{"in sync", "## main...origin/main", "main", 0, 0},
		{"no upstream", "## main", "main", 0, 0},
		{"fresh repo", "## No commits yet on main", "main", 0, 0},
		{"fresh repo with upstream", "## No commits yet on main...origin/main", "main", 0, 0},
		{"detached", "## HEAD (no branch)", "", 0, 0},
		{"slashed branch", "## feat/x...origin/feat/x [ahead 2]", "feat/x", 2, 0},
		{"dotted branch", "## v1.2.x...origin/v1.2.x", "v1.2.x", 0, 0},
		{"unrecognised shape", "## something odd here", "", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			branch, ahead, behind := parseStatusBranchHeader(tc.line)
			if branch != tc.wantBranch || ahead != tc.wantAhead || behind != tc.wantBehind {
				t.Errorf("parseStatusBranchHeader(%q) = (%q, %d, %d), want (%q, %d, %d)",
					tc.line, branch, ahead, behind, tc.wantBranch, tc.wantAhead, tc.wantBehind)
			}
		})
	}
}

// TestStatusSkipsBranchHeaderEntry is the regression this change most plausibly
// introduces: the header line is not a file. Its first two columns are '#' and
// '#', and '#' is neither ' ' nor '?', so an unskipped header would be counted
// as a Staged file and every staged count would be one too high.
func TestStatusSkipsBranchHeaderEntry(t *testing.T) {
	repo := newFixtureRepo(t)
	m, err := NewRepository(repo)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status.Staged) != 0 || len(status.Modified) != 0 || len(status.Untracked) != 0 {
		t.Fatalf("clean tree reported staged=%v modified=%v untracked=%v",
			status.Staged, status.Modified, status.Untracked)
	}
	if status.Branch != "main" {
		t.Errorf("Branch = %q, want %q", status.Branch, "main")
	}
}

// TestStatusCategorization pins the staged/modified/untracked split that
// Status reported before the header was added, over a tree holding one of each.
func TestStatusCategorization(t *testing.T) {
	repo := newFixtureRepo(t)

	// tracked.txt: staged AND then modified again in the working tree.
	writeFixtureFile(t, filepath.Join(repo, "tracked.txt"), "staged\n")
	gitFixture(t, repo, "add", "tracked.txt")
	writeFixtureFile(t, filepath.Join(repo, "tracked.txt"), "worktree\n")
	// added.txt: staged only.
	writeFixtureFile(t, filepath.Join(repo, "added.txt"), "added\n")
	gitFixture(t, repo, "add", "added.txt")
	// loose.txt: untracked.
	writeFixtureFile(t, filepath.Join(repo, "loose.txt"), "loose\n")

	m, err := NewRepository(repo)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	assertSet(t, "Staged", status.Staged, []string{"added.txt", "tracked.txt"})
	assertSet(t, "Modified", status.Modified, []string{"tracked.txt"})
	assertSet(t, "Untracked", status.Untracked, []string{"loose.txt"})
	if status.Branch != "main" {
		t.Errorf("Branch = %q, want main", status.Branch)
	}
	if status.Ahead != 0 || status.Behind != 0 {
		t.Errorf("no upstream: Ahead=%d Behind=%d, want 0/0", status.Ahead, status.Behind)
	}
}

// TestStatusAheadBehindFromHeader checks that the counts the removed rev-list
// spawn used to supply now arrive from the header, with the same values.
func TestStatusAheadBehindFromHeader(t *testing.T) {
	base := t.TempDir()
	repo := mkFixtureDir(t, filepath.Join(base, "repo"))
	gitFixture(t, repo, "init", "-q", "-b", "main")
	writeFixtureFile(t, filepath.Join(repo, "a.txt"), "a\n")
	gitFixture(t, repo, "add", "a.txt")
	gitFixture(t, repo, "commit", "-qm", "one")

	remote := filepath.Join(base, "remote.git")
	gitFixture(t, base, "init", "-q", "--bare", remote)
	gitFixture(t, repo, "remote", "add", "origin", remote)
	gitFixture(t, repo, "push", "-q", "-u", "origin", "main")

	m, err := NewRepository(repo)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	inSync, err := m.Status()
	if err != nil {
		t.Fatalf("Status (in sync): %v", err)
	}
	if inSync.Ahead != 0 || inSync.Behind != 0 {
		t.Errorf("in sync: Ahead=%d Behind=%d, want 0/0", inSync.Ahead, inSync.Behind)
	}

	writeFixtureFile(t, filepath.Join(repo, "b.txt"), "b\n")
	gitFixture(t, repo, "add", "b.txt")
	gitFixture(t, repo, "commit", "-qm", "two")

	clone := filepath.Join(base, "clone")
	gitFixture(t, base, "clone", "-q", remote, clone)
	writeFixtureFile(t, filepath.Join(clone, "c.txt"), "c\n")
	gitFixture(t, clone, "add", "c.txt")
	gitFixture(t, clone, "commit", "-qm", "three")
	gitFixture(t, clone, "push", "-q")
	gitFixture(t, repo, "fetch", "-q")

	diverged, divergedErr := m.Status()
	if divergedErr != nil {
		t.Fatalf("Status (diverged): %v", divergedErr)
	}
	if diverged.Ahead != 1 || diverged.Behind != 1 {
		t.Errorf("diverged: Ahead=%d Behind=%d, want 1/1", diverged.Ahead, diverged.Behind)
	}
	if diverged.Branch != "main" {
		t.Errorf("Branch = %q, want main", diverged.Branch)
	}
}

// TestStatusDetachedHEADLeavesBranchEmpty pins the contract the statusline
// collector relies on to fall back: a detached HEAD yields no branch name here,
// and CurrentBranch keeps owning ErrDetachedHEAD.
func TestStatusDetachedHEADLeavesBranchEmpty(t *testing.T) {
	repo := newFixtureRepo(t)
	gitFixture(t, repo, "checkout", "-q", "--detach", "HEAD")

	m, err := NewRepository(repo)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Branch != "" {
		t.Errorf("detached HEAD Branch = %q, want empty", status.Branch)
	}
	branch, branchErr := m.CurrentBranch()
	if branchErr == nil {
		t.Errorf("CurrentBranch on detached HEAD returned %q and no error", branch)
	}
}

// TestNewRepositoryErrorTaxonomy pins the two distinct failures the collapsed
// single-spawn rev-parse must keep telling apart.
func TestNewRepositoryErrorTaxonomy(t *testing.T) {
	base := t.TempDir()

	nonRepo := mkFixtureDir(t, filepath.Join(base, "plain"))
	plain, plainErr := NewRepository(nonRepo)
	if plainErr == nil {
		t.Fatalf("NewRepository on a non-repository returned %v and no error", plain)
	}
	if !strings.Contains(plainErr.Error(), "not a git repository") {
		t.Errorf("non-repository error = %v, want ErrNotRepository", plainErr)
	}

	barePath := filepath.Join(base, "bare.git")
	gitFixture(t, base, "init", "-q", "--bare", barePath)
	bare, bareErr := NewRepository(barePath)
	if bareErr == nil {
		t.Fatalf("NewRepository on a bare repository returned %v and no error", bare)
	}
	if !strings.Contains(bareErr.Error(), "get repository root") {
		t.Errorf("bare-repository error = %v, want the distinct root-resolution error", bareErr)
	}
	if strings.Contains(bareErr.Error(), "not a git repository") {
		t.Errorf("bare repository misreported as not-a-repository: %v", bareErr)
	}
}

// TestNewRepositoryResolvesRoot checks the toplevel is still read off the right
// line now that one rev-parse prints two.
func TestNewRepositoryResolvesRoot(t *testing.T) {
	repo := newFixtureRepo(t)
	m, err := NewRepository(repo)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	// The fixture path may traverse a symlink (/var vs /private/var on macOS),
	// which git resolves; compare against git's own answer.
	want := filepath.Clean(gitFixture(t, repo, "rev-parse", "--show-toplevel"))
	if m.Root() != want {
		t.Errorf("Root() = %q, want %q", m.Root(), want)
	}
}

func assertSet(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s = %v, want %v", label, got, want)
		return
	}
	seen := make(map[string]bool, len(got))
	for _, g := range got {
		seen[g] = true
	}
	for _, w := range want {
		if !seen[w] {
			t.Errorf("%s = %v, want %v", label, got, want)
			return
		}
	}
}
