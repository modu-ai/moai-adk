package security

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Card t560 (GH #1691) — runGit is the guardian's only subprocess.
//
// It sets cmd.Dir and leaves cmd.Env nil. git exports GIT_DIR, and on the
// commit path GIT_INDEX_FILE, into every hook it runs, and those OUTRANK the
// working directory — so under a hook the guardian inspects the caller's
// repository rather than the project root it was given. commitChangedFiles and
// commitDiff both resolve through runGit, which means the secret scan runs over
// another repository's commit and reports on a diff nobody asked about.
//
// The two fixtures carry different commit subjects on purpose: an assertion
// that only checked for a non-empty result, or for ok == true, would pass while
// reading the leaked repository.

// guardianCleanGitEnv returns the current environment with every GIT_ variable
// removed, so a fixture-side command reads the repository it names.
func guardianCleanGitEnv() []string {
	out := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "GIT_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// guardianGitIn runs git in dir with a scrubbed environment and returns trimmed stdout.
func guardianGitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = guardianCleanGitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// guardianNewRepo creates a repository with one commit whose subject is marker+"-seed".
func guardianNewRepo(t *testing.T, dir, marker string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	guardianGitIn(t, dir, "init", "-q", "-b", "main")
	guardianGitIn(t, dir, "config", "user.name", "t560")
	guardianGitIn(t, dir, "config", "user.email", "t560@test.invalid")
	if err := os.WriteFile(filepath.Join(dir, marker+".txt"), []byte(marker+"\n"), 0o644); err != nil {
		t.Fatalf("write %s.txt: %v", marker, err)
	}
	guardianGitIn(t, dir, "add", marker+".txt")
	guardianGitIn(t, dir, "commit", "-q", "-m", marker+"-seed")
	return dir
}

// guardianLeakPreCommit reproduces the environment git hands a pre-commit hook.
// GIT_WORK_TREE is deliberately absent: git does not export it to hooks, and a
// leak model stronger than the real one makes children fail outright rather
// than quietly reading the wrong repository.
func guardianLeakPreCommit(t *testing.T, hostRepo string) {
	t.Helper()
	gitDir := filepath.Join(hostRepo, ".git")
	t.Setenv("GIT_DIR", gitDir)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(gitDir, "index"))
}

// runGit must answer about the root it was given.
func TestRunGitReadsGivenRootNotLeakedGitDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	base := t.TempDir()
	host := guardianNewRepo(t, filepath.Join(base, "host"), "host")
	target := guardianNewRepo(t, filepath.Join(base, "target"), "target")

	// Fixture sanity: each repository reports its own commit subject, so a
	// failure below means the wrong repository was read rather than that
	// nothing was read.
	if got := guardianGitIn(t, host, "log", "-1", "--format=%s"); got != "host-seed" {
		t.Fatalf("fixture broken: host subject %q, want %q", got, "host-seed")
	}
	if got := guardianGitIn(t, target, "log", "-1", "--format=%s"); got != "target-seed" {
		t.Fatalf("fixture broken: target subject %q, want %q", got, "target-seed")
	}

	guardianLeakPreCommit(t, host)

	out, ok := runGit(target, "log", "-1", "--format=%s")
	if !ok {
		t.Fatalf("runGit(%s, log) reported failure", target)
	}
	if got := strings.TrimSpace(out); got != "target-seed" {
		t.Errorf("runGit(%s) answered about the LEAKED repository: got %q, want %q\nGIT_DIR=%s GIT_INDEX_FILE=%s",
			target, got, "target-seed", os.Getenv("GIT_DIR"), os.Getenv("GIT_INDEX_FILE"))
	}
}

// commitChangedFiles is the guardian's actual entry into runGit. It exercises
// the same leak one layer up, so a fix confined to a single call site — the
// failure mode that produced this card — cannot satisfy both tests.
//
// The filenames differ per repository: the leaked answer is "host.txt" and the
// correct one "target.txt", so a count-based assertion could not separate them.
func TestCommitChangedFilesReadsGivenRootNotLeakedGitDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	base := t.TempDir()
	host := guardianNewRepo(t, filepath.Join(base, "host"), "host")
	target := guardianNewRepo(t, filepath.Join(base, "target"), "target")

	if got := commitChangedFiles(target); len(got) != 1 || got[0] != "target.txt" {
		t.Fatalf("fixture broken: commitChangedFiles(target) = %v, want [target.txt]", got)
	}

	guardianLeakPreCommit(t, host)

	got := commitChangedFiles(target)
	if len(got) != 1 || got[0] != "target.txt" {
		t.Errorf("commitChangedFiles(%s) answered about the LEAKED repository: got %v, want [target.txt]\nGIT_DIR=%s GIT_INDEX_FILE=%s",
			target, got, os.Getenv("GIT_DIR"), os.Getenv("GIT_INDEX_FILE"))
	}
}
