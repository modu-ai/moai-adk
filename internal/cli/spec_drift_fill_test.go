package cli

// spec_drift_fill_test.go — AC-DCF-013 (the card's headline behaviour) and
// AC-DCF-004(a) (the fill exits at its deadline).
//
// The fill child is the ONLY writer of the drift cache that outlives the
// SessionStart hook process, so "does a fill actually leave a HEAD-matching
// cache behind" is the release-blocking carrier of this card. The field
// measurement (AC-DCF-015) is corroboration, not the gate.

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// driftFillGitRepo builds a real one-commit git repository carrying one SPEC
// directory, and makes it the process working directory for the test (t.Chdir
// forces the test serial, which a cwd-sensitive test wants anyway).
func driftFillGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	specDir := filepath.Join(dir, ".moai", "specs", "SPEC-DFILL-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	body := "---\nid: SPEC-DFILL-001\nstatus: completed\n---\n\n# fixture\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "--quiet", "--initial-branch=main")
	run("add", "-A")
	run("commit", "--quiet", "-m", "fixture")

	t.Chdir(dir)
	return dir
}

func driftFillHeadSHA(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return string(bytes.TrimSpace(out))
}

func runSpecDriftCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newSpecDriftCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// TestSpecDriftFillWritesHeadKeyedCache is AC-DCF-013. The fill runs to
// completion against the REAL drift computation on a real (tiny) git
// repository; afterwards the cache exists, parses, and its head_sha equals the
// repository's HEAD.
//
// The assertion is on cache CONTENT keyed to a HEAD the test computes itself —
// a mutant writing an empty or garbage file satisfies mere existence.
func TestSpecDriftFillWritesHeadKeyedCache(t *testing.T) {
	dir := driftFillGitRepo(t)

	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	cachePath := filepath.Join(dir, ".moai", "state", "drift-cache.json")
	if _, err := os.Stat(cachePath); err == nil {
		t.Fatal("precondition: the cache already exists before the fill ran")
	}

	out, err := runSpecDriftCmd(t, "--fill-cache")
	if err != nil {
		t.Fatalf("fill returned an error: %v\n%s", err, out)
	}

	raw, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("the fill left no drift cache behind: %v", err)
	}
	var payload struct {
		HeadSHA string `json:"head_sha"`
		Records []any  `json:"records"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("the cache the fill wrote does not parse: %v\n%s", err, raw)
	}
	if want := driftFillHeadSHA(t, dir); payload.HeadSHA != want {
		t.Fatalf("cache head_sha = %q, want the repository HEAD %q", payload.HeadSHA, want)
	}
}

// TestSpecDriftFillIsSilent asserts the detach contract's output half at the
// child's end: the fill writes nothing to stdout, so a child whose streams were
// ever wired to a parent could not hold a pipe open with content.
func TestSpecDriftFillIsSilent(t *testing.T) {
	dir := driftFillGitRepo(t)
	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	out, err := runSpecDriftCmd(t, "--fill-cache")
	if err != nil {
		t.Fatalf("fill returned an error: %v", err)
	}
	if out != "" {
		t.Fatalf("fill wrote %q to its output, want nothing", out)
	}
}

// TestSpecDriftFillExitsAtItsDeadline is AC-DCF-004(a): the entry point bounds
// ITSELF. The seam sleeps 5s and ignores every cancellation signal — the shape
// of the real drift computation, which is a synchronous git+in-memory pass with
// no context awareness — and the command must still return at its 50ms
// deadline.
//
// The positive control that excludes the "returns instantly by never computing"
// mutant is TestSpecDriftFillWritesHeadKeyedCache above: the same command
// against the real computation writes a complete, parseable, HEAD-matching
// cache.
func TestSpecDriftFillExitsAtItsDeadline(t *testing.T) {
	dir := driftFillGitRepo(t)
	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	origCompute := driftFillComputeFn
	started := make(chan struct{})
	driftFillComputeFn = func(string) {
		close(started)
		time.Sleep(5 * time.Second)
	}
	t.Cleanup(func() { driftFillComputeFn = origCompute })

	start := time.Now()
	if _, err := runSpecDriftCmd(t, "--fill-cache", "--fill-timeout", "50ms"); err != nil {
		t.Fatalf("fill returned an error: %v", err)
	}
	elapsed := time.Since(start)

	select {
	case <-started:
	default:
		t.Fatal("the fill returned without ever entering the computation — the elapsed bound below would be vacuous")
	}
	if elapsed >= 250*time.Millisecond {
		t.Fatalf("fill returned after %v with a 50ms deadline, want < 250ms", elapsed)
	}
}

// TestSpecDriftFillFlagsAreHidden pins the surface decision: the fill flags are
// a machine-to-machine channel between the SessionStart handler and the child
// it starts, not a user-facing verb.
func TestSpecDriftFillFlagsAreHidden(t *testing.T) {
	cmd := newSpecDriftCmd()
	for _, name := range []string{"fill-cache", "fill-timeout"} {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Fatalf("flag --%s is not registered", name)
		}
		if !f.Hidden {
			t.Errorf("flag --%s is not hidden", name)
		}
	}
}

// TestSpecDriftWithoutFillFlagIsUnchanged guards the untouched path: the plain
// verb still reports, and still does not take the fill branch.
func TestSpecDriftWithoutFillFlagIsUnchanged(t *testing.T) {
	dir := driftFillGitRepo(t)
	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	out, err := runSpecDriftCmd(t)
	if err != nil {
		t.Fatalf("plain drift returned an error: %v\n%s", err, out)
	}
	if !bytes.Contains([]byte(out), []byte("Summary:")) {
		t.Fatalf("plain drift output lost its report:\n%s", out)
	}
}
