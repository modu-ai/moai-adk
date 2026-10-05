package graph

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// AC-GSA-006 / AC-GSA-007 — the graph-freshness reachability guard, executed
// rather than read. The step's own shell script is extracted from the workflow
// and run against synthetic git histories under the three event shapes, so the
// assertion is the guard's exit code and not a substring of the YAML.
//
// A text scan would answer "does TARGET=HEAD appear in the push branch"; it
// would NOT answer whether the release conditional is reachable at all once
// the push branch stops exiting early, which is the behaviour this SPEC
// changes. Running the script covers both.
//
// Both event shapes judge the checked-out tree. A PR checks its merge preview;
// a push checks the landed commit and catches a squash-orphaned stamp:
//
//	push          → HEAD (stamp off HEAD fails even though the object exists)
//	ordinary PR   → HEAD (a new stamp on the PR branch passes)
//	release/* PR  → HEAD (the checkout is the merge preview)
const graphFreshnessWorkflowRelPath = ".github/workflows/graph-freshness.yml"

const reachabilityStepName = "Guard codemaps stamp reachability (pre-merge)"

func TestGraphFreshnessReachabilityGuard_TargetSelection(t *testing.T) {
	requireBinaries(t, "bash", "jq", "git")
	script := reachabilityStepScript(t)

	t.Run("push rejects an object-present non-ancestor stamp", func(t *testing.T) {
		repo := newGuardRepo(t)
		stamp := guardSideCommit(t, repo)
		writeGuardProvenance(t, repo, stamp)
		// The object resolves; only ancestry fails. Before the fix the push
		// branch exited 0 right after the cat-file probe.
		code := runGuard(t, script, repo, map[string]string{"GITHUB_BASE_REF": "", "GITHUB_HEAD_REF": ""})
		if code == 0 {
			t.Fatalf("push guard exited 0 on a stamp that is not an ancestor of HEAD — object presence was accepted as reachability")
		}
	})

	t.Run("push accepts an ancestor stamp", func(t *testing.T) {
		repo := newGuardRepo(t)
		head := guardGit(t, repo, "rev-parse", "HEAD")
		writeGuardProvenance(t, repo, head)
		if code := runGuard(t, script, repo, map[string]string{"GITHUB_BASE_REF": "", "GITHUB_HEAD_REF": ""}); code != 0 {
			t.Fatalf("push guard exited %d on an ancestor stamp — the guard must not fail a reachable tree", code)
		}
	})

	t.Run("ordinary PR accepts a new stamp in merge preview", func(t *testing.T) {
		repo := newGuardRepo(t)
		// Stamp reachable from HEAD but not origin/main. The PR cannot
		// already contain its new stamp in the target base branch.
		stamp := guardHeadOnlyCommit(t, repo)
		writeGuardProvenance(t, repo, stamp)
		code := runGuard(t, script, repo, map[string]string{"GITHUB_BASE_REF": "main", "GITHUB_HEAD_REF": "feature/x"})
		if code != 0 {
			t.Fatalf("ordinary PR guard exited %d on a reachable new stamp — the target must be merge-preview HEAD", code)
		}
	})

	t.Run("ordinary PR rejects a resolvable stamp outside merge preview", func(t *testing.T) {
		repo := newGuardRepo(t)
		stamp := guardSideCommit(t, repo)
		writeGuardProvenance(t, repo, stamp)
		if code := runGuard(t, script, repo, map[string]string{"GITHUB_BASE_REF": "main", "GITHUB_HEAD_REF": "feature/x"}); code == 0 {
			t.Fatal("ordinary PR guard accepted a stamp that merge-preview HEAD cannot reach")
		}
	})

	t.Run("release PR judges the merge-preview HEAD", func(t *testing.T) {
		repo := newGuardRepo(t)
		stamp := guardHeadOnlyCommit(t, repo)
		writeGuardProvenance(t, repo, stamp)
		// Same fixture as the ordinary-PR row; only the head ref differs. The
		// release lane merges with a merge commit, so a head-reachable stamp
		// survives and the guard must pass.
		code := runGuard(t, script, repo, map[string]string{"GITHUB_BASE_REF": "main", "GITHUB_HEAD_REF": "release/v1.0.0"})
		if code != 0 {
			t.Fatalf("release PR guard exited %d on a head-reachable stamp — the merge-preview target was lost", code)
		}
	})

	t.Run("anchorless provenance skips with a reason", func(t *testing.T) {
		repo := newGuardRepo(t)
		writeGuardProvenanceRaw(t, repo, `{"schema_version":1,"dirty":true,"content_fingerprint":"abc"}`)
		if code := runGuard(t, script, repo, map[string]string{"GITHUB_BASE_REF": "", "GITHUB_HEAD_REF": ""}); code != 0 {
			t.Fatalf("anchorless provenance must skip with exit 0, got %d — the pre-existing contract is preserved", code)
		}
	})

	t.Run("missing object fails on every event shape", func(t *testing.T) {
		for _, env := range []map[string]string{
			{"GITHUB_BASE_REF": "", "GITHUB_HEAD_REF": ""},
			{"GITHUB_BASE_REF": "main", "GITHUB_HEAD_REF": "feature/x"},
			{"GITHUB_BASE_REF": "main", "GITHUB_HEAD_REF": "release/v1.0.0"},
		} {
			repo := newGuardRepo(t)
			writeGuardProvenance(t, repo, strings.Repeat("0", 40))
			if code := runGuard(t, script, repo, env); code == 0 {
				t.Fatalf("missing-object guard exited 0 for env %v", env)
			}
		}
	})
}

// --- fixture plumbing ----------------------------------------------------

func requireBinaries(t *testing.T, names ...string) {
	t.Helper()
	for _, n := range names {
		if _, err := exec.LookPath(n); err != nil {
			t.Fatalf("%s is required to execute the workflow guard; a skipped run would be a vacuous pass", n)
		}
	}
}

// reachabilityStepScript pulls the guard step's `run:` body out of the
// workflow. Parsing the YAML rather than slicing the file keeps the test
// pinned to the step by NAME, so reordering the job does not silently make it
// read a different step.
func reachabilityStepScript(t *testing.T) string {
	t.Helper()
	path := filepath.Join(findRepoRootForWorkflowTest(t), graphFreshnessWorkflowRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", graphFreshnessWorkflowRelPath, err)
	}
	var wf struct {
		Jobs map[string]struct {
			Steps []struct {
				Name string `yaml:"name"`
				Run  string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatalf("parse %s: %v", graphFreshnessWorkflowRelPath, err)
	}
	for _, job := range wf.Jobs {
		for _, step := range job.Steps {
			if step.Name == reachabilityStepName {
				if strings.TrimSpace(step.Run) == "" {
					t.Fatalf("step %q carries an empty run body", reachabilityStepName)
				}
				return step.Run
			}
		}
	}
	t.Fatalf("step %q not found in %s", reachabilityStepName, graphFreshnessWorkflowRelPath)
	return ""
}

// findRepoRootForWorkflowTest walks up from the test's working directory until
// the workflow is found, so the test works from any package directory.
func findRepoRootForWorkflowTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, graphFreshnessWorkflowRelPath)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate %s above %s", graphFreshnessWorkflowRelPath, dir)
		}
		dir = parent
	}
}

// newGuardRepo builds a two-commit repo whose `origin/main` remote-tracking
// ref points at the initial commit, so base-vs-head targets can diverge.
func newGuardRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	guardGit(t, root, "init", "-q", "-b", "main")
	guardGit(t, root, "config", "user.email", "t688@example.test")
	guardGit(t, root, "config", "user.name", "t688")
	guardGit(t, root, "config", "gc.auto", "0")
	guardGit(t, root, "config", "maintenance.auto", "false")
	writeGuardFile(t, root, "README.md", "# base\n")
	guardGit(t, root, "add", "README.md")
	guardGit(t, root, "commit", "-q", "-m", "base")
	base := guardGit(t, root, "rev-parse", "HEAD")
	guardGit(t, root, "update-ref", "refs/remotes/origin/main", base)
	return root
}

// guardSideCommit creates a commit on a side branch and returns to the base
// branch, so the returned sha is a resolvable object that HEAD cannot reach.
func guardSideCommit(t *testing.T, root string) string {
	t.Helper()
	guardGit(t, root, "switch", "-q", "-c", "side")
	writeGuardFile(t, root, "side.txt", "side\n")
	guardGit(t, root, "add", "side.txt")
	guardGit(t, root, "commit", "-q", "-m", "side")
	sha := guardGit(t, root, "rev-parse", "HEAD")
	guardGit(t, root, "switch", "-q", "main")
	return sha
}

// guardHeadOnlyCommit advances HEAD past origin/main and returns the new sha:
// reachable from HEAD, not reachable from origin/main.
func guardHeadOnlyCommit(t *testing.T, root string) string {
	t.Helper()
	writeGuardFile(t, root, "head-only.txt", "head only\n")
	guardGit(t, root, "add", "head-only.txt")
	guardGit(t, root, "commit", "-q", "-m", "head only")
	return guardGit(t, root, "rev-parse", "HEAD")
}

func writeGuardProvenance(t *testing.T, root, sha string) {
	t.Helper()
	writeGuardProvenanceRaw(t, root, fmt.Sprintf(`{"schema_version":1,"commit_sha":%q,"dirty":false}`, sha))
}

func writeGuardProvenanceRaw(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "provenance.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeGuardFile(t *testing.T, root, rel, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func guardGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// runGuard executes the extracted step body with bash in the fixture repo and
// returns its exit code. stdout/stderr are surfaced on failure only.
func runGuard(t *testing.T, script, repo string, env map[string]string) int {
	t.Helper()
	cmd := exec.Command("bash", "-c", script)
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	out, err := cmd.CombinedOutput()
	t.Logf("guard output (env=%v):\n%s", env, out)
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if ok := asExitError(err, &exitErr); ok {
		return exitErr.ExitCode()
	}
	t.Fatalf("running the guard failed for a reason other than its exit code: %v", err)
	return -1
}

func asExitError(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	return false
}
