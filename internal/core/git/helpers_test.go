package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a temporary Git repository with an initial commit on "main".
// The returned path has symlinks resolved (important for macOS /var -> /private/var).
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := resolveSymlinks(t, t.TempDir())
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	writeTestFile(t, filepath.Join(dir, "README.md"), "# Test\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Initial commit")
	return dir
}

// resolveSymlinks resolves all symlinks in a path (e.g., macOS /var -> /private/var).
func resolveSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve symlinks for %q: %v", path, err)
	}
	return resolved
}

// initTestRepoWithRemote creates a local repo with a bare remote and upstream tracking.
// Returns (localDir, remoteDir).
func initTestRepoWithRemote(t *testing.T) (string, string) {
	t.Helper()

	// Create bare remote.
	remoteDir := resolveSymlinks(t, t.TempDir())
	runGit(t, remoteDir, "init", "--bare", "-b", "main")

	// Create local repo.
	localDir := resolveSymlinks(t, t.TempDir())
	runGit(t, localDir, "init", "-b", "main")
	runGit(t, localDir, "config", "user.email", "test@example.com")
	runGit(t, localDir, "config", "user.name", "Test User")

	// Add remote, initial commit, push with tracking.
	runGit(t, localDir, "remote", "add", "origin", remoteDir)
	writeTestFile(t, filepath.Join(localDir, "README.md"), "# Test\n")
	runGit(t, localDir, "add", ".")
	runGit(t, localDir, "commit", "-m", "Initial commit")
	runGit(t, localDir, "push", "-u", "origin", "main")

	return localDir, remoteDir
}

// runGit executes a git command in dir and returns trimmed stdout.
// Fails the test on error.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %s: %v", args, dir, string(out), err)
	}
	return strings.TrimSpace(string(out))
}

// writeTestFile creates a file with the given content, creating parent directories as needed.
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
