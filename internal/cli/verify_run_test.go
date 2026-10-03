package cli

// verify_run_test.go — SPEC-VERIFY-RUN-REUSE-001 acceptance tests for
// `moai verify run`. The commands under test are the test binary itself
// re-executed as a helper process (no shell, no echo — portable to Windows),
// switched on by MOAI_VERIFY_RUN_HELPER so a normal test run skips it.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/verify"
)

const verifyRunHelperEnv = "MOAI_VERIFY_RUN_HELPER"

// TestVerifyRunHelperProcess is the helper process, not a test: modes follow
// the "--" in os.Args. It exits itself so the test framework prints nothing
// onto the stdout a verify-run assertion reads.
func TestVerifyRunHelperProcess(t *testing.T) {
	if os.Getenv(verifyRunHelperEnv) == "" {
		t.Skip("helper process for verify run tests; runs only under " + verifyRunHelperEnv)
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 {
		os.Exit(90)
	}
	switch args[0] {
	case "count": // count <file> [ignored...]: append one line, print "counted"
		verifyRunHelperAppend(args[1])
		_, _ = os.Stdout.WriteString("counted\n")
	case "count-exit": // count-exit <file> <N>
		verifyRunHelperAppend(args[1])
		n, _ := strconv.Atoi(args[2])
		os.Exit(n)
	case "count-flaky": // fails on its first run, passes afterwards
		if verifyRunHelperAppend(args[1]) == 1 {
			os.Exit(1)
		}
	case "exit":
		n, _ := strconv.Atoi(args[1])
		os.Exit(n)
	case "sleep":
		d, _ := time.ParseDuration(args[1])
		time.Sleep(d)
	case "spawn-sleep": // spawn-sleep <d> <marker>: a grandchild touches marker after 2s
		child := exec.Command(os.Args[0], "-test.run=^TestVerifyRunHelperProcess$", "--", "touch-after", "2s", args[2])
		if err := child.Start(); err != nil {
			os.Exit(91)
		}
		d, _ := time.ParseDuration(args[1])
		time.Sleep(d)
	case "touch-after": // touch-after <d> <file>
		d, _ := time.ParseDuration(args[1])
		time.Sleep(d)
		_ = os.WriteFile(args[2], []byte("alive"), 0o644)
	case "mutate": // mutate <tracked file>
		f, err := os.OpenFile(args[1], os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			os.Exit(92)
		}
		_, _ = f.WriteString("mutated\n")
		_ = f.Close()
	case "print":
		if len(args) > 1 {
			_, _ = os.Stdout.WriteString(args[1])
		}
	case "eprint":
		if len(args) > 1 {
			_, _ = os.Stderr.WriteString(args[1])
		}
	default:
		os.Exit(93)
	}
	os.Exit(0)
}

// verifyRunHelperAppend appends a line to path and returns the line count.
func verifyRunHelperAppend(path string) int {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		os.Exit(94)
	}
	_, _ = f.WriteString("x\n")
	_ = f.Close()
	data, _ := os.ReadFile(path)
	return strings.Count(string(data), "\n")
}

// helperArgv returns the argv that re-executes this test binary in a helper mode.
func helperArgv(mode string, args ...string) []string {
	return append([]string{os.Args[0], "-test.run=^TestVerifyRunHelperProcess$", "--", mode}, args...)
}

// helperRepo is a throwaway repo plus the counter file kept outside it.
type helperRepo struct {
	t       *testing.T
	dir     string
	counter string
}

func newHelperRepo(t *testing.T) *helperRepo {
	t.Helper()
	t.Setenv(verifyRunHelperEnv, "1")
	return &helperRepo{t: t, dir: initVerifyTestRepo(t), counter: filepath.Join(t.TempDir(), "counter")}
}

func (r *helperRepo) count() int {
	data, err := os.ReadFile(r.counter)
	if err != nil {
		return 0
	}
	return strings.Count(string(data), "\n")
}

// runResult is one `verify run` invocation's observable outcome.
type runResult struct {
	stdout, stderr string
	code           int
}

// run invokes `verify run <flags> -- <argv...>` against the repo.
func (r *helperRepo) run(flags []string, argv []string) runResult {
	r.t.Helper()
	cmd := newVerifyCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	args := append([]string{"run", "--project-root", r.dir}, flags...)
	args = append(args, "--")
	args = append(args, argv...)
	cmd.SetArgs(args)
	err := cmd.Execute()
	res := runResult{stdout: out.String(), stderr: errBuf.String()}
	var ec *exitCodeError
	switch {
	case err == nil:
	case errors.As(err, &ec):
		res.code = ec.ExitCode()
	default:
		res.code = 1
		res.stderr += err.Error()
	}
	return res
}

func (r *helperRepo) runCount(flags ...string) runResult {
	r.t.Helper()
	return r.run(flags, helperArgv("count", r.counter))
}

func (r *helperRepo) expectCount(want int, what string) {
	r.t.Helper()
	if got := r.count(); got != want {
		r.t.Fatalf("%s: executions = %d, want %d", what, got, want)
	}
}

func (r *helperRepo) git(args ...string) {
	r.t.Helper()
	out, err := exec.Command("git", append([]string{"-C", r.dir}, args...)...).CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func (r *helperRepo) write(name, content string) {
	r.t.Helper()
	if err := os.WriteFile(filepath.Join(r.dir, name), []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// snapshotExists reports whether any snapshot is recorded for the repo's tree.
func (r *helperRepo) snapshotExists() bool {
	r.t.Helper()
	key, err := verify.Key(r.t.Context(), r.dir)
	if err != nil {
		r.t.Fatal(err)
	}
	snap, err := verify.Load(r.dir, key)
	if err != nil {
		r.t.Fatal(err)
	}
	return snap != nil
}

func TestVerifyRunHitExecutesZeroTimes(t *testing.T) {
	r := newHelperRepo(t)
	first := r.runCount("--env", "FOO")
	if first.code != 0 || !strings.Contains(first.stdout, "counted") {
		t.Fatalf("first run: code=%d stdout=%q stderr=%q", first.code, first.stdout, first.stderr)
	}
	r.expectCount(1, "after the first run")
	for i := 2; i <= 5; i++ {
		res := r.runCount("--env", "FOO")
		if res.code != 0 {
			t.Fatalf("run %d: exit %d, want 0 (stderr %q)", i, res.code, res.stderr)
		}
		if res.stdout != "" {
			t.Errorf("run %d: a reuse must print no command output, got %q", i, res.stdout)
		}
		for _, want := range []string{"reuse", "key=", "recorded_at=", "duration_ms="} {
			if !strings.Contains(res.stderr, want) {
				t.Errorf("run %d: reuse notice lacks %q: %q", i, want, res.stderr)
			}
		}
	}
	r.expectCount(1, "after five runs")
}

func TestVerifyRunMissOnTreeChange(t *testing.T) {
	r := newHelperRepo(t)
	r.runCount()
	r.runCount()
	r.expectCount(1, "baseline")

	r.write("tracked.txt", "v2\n") // (a) tracked edit
	r.runCount()
	r.expectCount(2, "tracked file edited")
	r.runCount()
	r.expectCount(2, "same edited tree reuses")

	r.write("new.txt", "fresh\n") // (b) new untracked file
	r.runCount()
	r.expectCount(3, "untracked file added")

	r.git("add", "tracked.txt", "new.txt") // (c) new commit
	r.git("commit", "-q", "-m", "second")
	r.runCount()
	r.expectCount(4, "new commit")
	r.runCount()
	r.expectCount(4, "committed tree reuses")
}

func TestVerifyRunMissOnCommandBytes(t *testing.T) {
	r := newHelperRepo(t)
	variants := [][]string{
		{"a", "b"},
		{"a b"},
		{"b", "a"},
		{"a", "b", ""},
		{"a  b"},
		{"a", "b "},
	}
	for i, extra := range variants {
		argv := helperArgv("count", append([]string{r.counter}, extra...)...)
		r.run(nil, argv)
		r.expectCount(i+1, "variant "+strconv.Itoa(i)+" runs once")
		r.run(nil, argv)
		r.expectCount(i+1, "variant "+strconv.Itoa(i)+" repeats as a hit")
	}
}

func TestVerifyRunNeverReusesFailure(t *testing.T) {
	r := newHelperRepo(t)
	argv := helperArgv("count-flaky", r.counter)
	if res := r.run(nil, argv); res.code != 1 {
		t.Fatalf("first run: exit %d, want 1", res.code)
	}
	if res := r.run(nil, argv); res.code != 0 {
		t.Fatalf("second run: exit %d, want 0 (the failure must not be reused)", res.code)
	}
	if got := r.count(); got != 2 {
		t.Fatalf("executions after fail then pass = %d, want 2", got)
	}
	if res := r.run(nil, argv); res.code != 0 || r.count() != 2 {
		t.Fatalf("third run: exit %d, executions %d; the passing result must now be reused", res.code, r.count())
	}
}

func TestVerifyRunMissOnTTL(t *testing.T) {
	r := newHelperRepo(t)
	r.runCount()
	r.runCount()
	r.expectCount(1, "default TTL reuses")
	time.Sleep(5 * time.Millisecond)
	res := r.runCount("--ttl", "1ns")
	r.expectCount(2, "a 1ns TTL never reuses")
	if !strings.Contains(res.stderr, "miss") {
		t.Errorf("miss notice absent: %q", res.stderr)
	}
}

func TestVerifyRunMissOnEnvChange(t *testing.T) {
	r := newHelperRepo(t)
	t.Setenv("MOAI_VRR_BOUND", "-a")
	t.Setenv("MOAI_VRR_UNLISTED", "x")
	flags := []string{"--env", "MOAI_VRR_BOUND"}

	r.runCount(flags...)
	r.runCount(flags...)
	r.expectCount(1, "same value reuses")

	t.Setenv("MOAI_VRR_UNLISTED", "y")
	r.runCount(flags...)
	r.expectCount(1, "an unlisted variable never breaks a reuse")

	t.Setenv("MOAI_VRR_BOUND", "-b")
	r.runCount(flags...)
	r.expectCount(2, "changed value")

	if err := os.Unsetenv("MOAI_VRR_BOUND"); err != nil {
		t.Fatal(err)
	}
	r.runCount(flags...)
	r.expectCount(3, "unset")
	r.runCount(flags...)
	r.expectCount(3, "unset repeats as a hit")

	t.Setenv("MOAI_VRR_BOUND", "")
	r.runCount(flags...)
	r.expectCount(4, "empty string differs from unset")
}

// toolVersionFlags binds the tool identity to a helper-printed word.
func toolVersionFlags(mode string, args ...string) []string {
	var flags []string
	for _, e := range helperArgv(mode, args...) {
		flags = append(flags, "--tool-version-cmd", e)
	}
	return flags
}

func TestVerifyRunToolVersionBinding(t *testing.T) {
	r := newHelperRepo(t)
	v1 := toolVersionFlags("print", "v1")
	r.runCount(v1...)
	r.runCount(v1...)
	r.expectCount(1, "same identity reuses")

	r.runCount(toolVersionFlags("print", "v2")...)
	r.expectCount(2, "identity v2")
	r.runCount()
	r.expectCount(3, "no flag binds the fixed marker, not v1/v2")
	r.runCount()
	r.expectCount(3, "unversioned repeats as a hit")

	// Unbound identities: no reuse, no record, the command still runs and its
	// own exit code decides.
	unbound := map[string][]string{
		"non-zero exit":     toolVersionFlags("exit", "3"),
		"stderr only":       toolVersionFlags("eprint", "v9"),
		"empty output":      toolVersionFlags("print"),
		"cannot start":      {"--tool-version-cmd", filepath.Join(r.dir, "no-such-program")},
		"whitespace output": toolVersionFlags("print", " \n"),
	}
	for name, flags := range unbound {
		fresh := newHelperRepo(t)
		res := fresh.run(flags, helperArgv("count-exit", fresh.counter, "5"))
		if res.code != 5 {
			t.Errorf("%s: exit %d, want the command's own 5 (stderr %q)", name, res.code, res.stderr)
		}
		if !strings.Contains(res.stderr, "tool version") {
			t.Errorf("%s: stderr must name the cause: %q", name, res.stderr)
		}
		fresh.run(flags, helperArgv("count-exit", fresh.counter, "5"))
		if fresh.count() != 2 {
			t.Errorf("%s: executions = %d, want 2 (never reused)", name, fresh.count())
		}
		if fresh.snapshotExists() {
			t.Errorf("%s: a snapshot was recorded for an unbound tool identity", name)
		}
	}
}

func TestVerifyRunToolVersionTimeout(t *testing.T) {
	r := newHelperRepo(t)
	flags := append(toolVersionFlags("sleep", "30s"), "--tool-version-timeout", "200ms")
	start := time.Now()
	res := r.runCount(flags...)
	if elapsed := time.Since(start); elapsed > 20*time.Second {
		t.Errorf("tool-version timeout not enforced: took %s", elapsed)
	}
	if res.code != 0 || r.count() != 1 {
		t.Fatalf("command must still run once: exit %d, executions %d", res.code, r.count())
	}
	if !strings.Contains(res.stderr, "tool version") || !strings.Contains(res.stderr, "timed out") {
		t.Errorf("stderr must name the timeout: %q", res.stderr)
	}
	r.runCount(flags...)
	r.expectCount(2, "an unbound identity is never reused")
	if r.snapshotExists() {
		t.Error("a snapshot was recorded under an unbound tool identity")
	}
}

func TestVerifyRunExitCodePassthrough(t *testing.T) {
	r := newHelperRepo(t)
	for _, want := range []int{0, 3, 1} {
		res := r.run(nil, helperArgv("exit", strconv.Itoa(want)))
		if res.code != want {
			t.Errorf("exit %d: verify run exited %d", want, res.code)
		}
	}
}

func TestVerifyRunTimeoutAndNotFound(t *testing.T) {
	t.Run("ExitCodes124And127", func(t *testing.T) {
		r := newHelperRepo(t)
		res := r.run([]string{"--timeout", "300ms"}, helperArgv("sleep", "30s"))
		if res.code != 124 {
			t.Errorf("timeout: exit %d, want 124 (stderr %q)", res.code, res.stderr)
		}
		if r.snapshotExists() {
			t.Error("a timed-out run was recorded")
		}
		res = r.run(nil, []string{filepath.Join(r.dir, "no-such-program")})
		if res.code != 127 {
			t.Errorf("not found: exit %d, want 127 (stderr %q)", res.code, res.stderr)
		}
		if r.snapshotExists() {
			t.Error("a command that could not start was recorded")
		}
	})
	t.Run("GrandchildKilled", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("process-group termination is Unix only; residual risk spec §D-5")
		}
		r := newHelperRepo(t)
		marker := filepath.Join(t.TempDir(), "grandchild-alive")
		res := r.run([]string{"--timeout", "500ms"}, helperArgv("spawn-sleep", "30s", marker))
		if res.code != 124 {
			t.Fatalf("exit %d, want 124 (stderr %q)", res.code, res.stderr)
		}
		time.Sleep(3 * time.Second)
		if _, err := os.Stat(marker); err == nil {
			t.Error("the grandchild survived the timeout: its marker file exists")
		}
	})
}

func TestVerifyRunTreeMovedNotRecorded(t *testing.T) {
	r := newHelperRepo(t)
	before, err := verify.Key(t.Context(), r.dir)
	if err != nil {
		t.Fatal(err)
	}
	res := r.run(nil, helperArgv("mutate", filepath.Join(r.dir, "tracked.txt")))
	if res.code != 0 {
		t.Fatalf("exit %d, want the command's own 0 (stderr %q)", res.code, res.stderr)
	}
	if !strings.Contains(res.stderr, "tree changed") {
		t.Errorf("tree-moved notice absent: %q", res.stderr)
	}
	after, err := verify.Key(t.Context(), r.dir)
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("test premise broken: the mutate helper did not move the tree key")
	}
	for _, key := range []string{before, after} {
		if snap, err := verify.Load(r.dir, key); err != nil || snap != nil {
			t.Errorf("a snapshot exists for key %s (err %v)", key, err)
		}
	}
}

func TestVerifyRunStoreFailOpen(t *testing.T) {
	// A regular file where the snapshot directory should be: loads and saves
	// both fail on every platform without relying on permission bits.
	r := newHelperRepo(t)
	blocker := filepath.Join(r.dir, filepath.FromSlash(verify.SnapshotDir))
	if err := os.MkdirAll(filepath.Dir(blocker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	argv := helperArgv("count-exit", r.counter, "4")
	for i := 1; i <= 2; i++ {
		res := r.run(nil, argv)
		if res.code != 4 {
			t.Errorf("run %d: exit %d, want the command's own 4 (stderr %q)", i, res.code, res.stderr)
		}
		if !strings.Contains(res.stderr, "snapshot") {
			t.Errorf("run %d: stderr must name the store failure: %q", i, res.stderr)
		}
	}
	r.expectCount(2, "a failing store re-executes, never fabricates a hit")

	// Key computation failing (not a git repository) fails open the same way.
	nonGit := &helperRepo{t: t, dir: t.TempDir(), counter: filepath.Join(t.TempDir(), "c2")}
	res := nonGit.run(nil, helperArgv("count-exit", nonGit.counter, "6"))
	if res.code != 6 || nonGit.count() != 1 {
		t.Errorf("non-git root: exit %d, executions %d, want 6 and 1 (stderr %q)", res.code, nonGit.count(), res.stderr)
	}
}

func TestVerifyRunIgnoresHandRecordedEntry(t *testing.T) {
	r := newHelperRepo(t)
	argv := helperArgv("count", r.counter)
	if _, err := runVerifyCmd(t, "record", "--project-root", r.dir,
		"--check-id", "test", "--command", verify.CanonicalCommand(argv), "--exit", "0"); err != nil {
		t.Fatal(err)
	}
	r.run(nil, argv)
	r.expectCount(1, "a hand-recorded entry has no config_digest/tool_version and must not be reused")
	r.run(nil, argv)
	r.expectCount(1, "the verb's own record is reused")
}

func TestVerifyRunUsageErrors(t *testing.T) {
	r := newHelperRepo(t)
	if res := r.run(nil, nil); res.code != 2 {
		t.Errorf("no command after --: exit %d, want 2", res.code)
	}
	cmd := newVerifyCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"run", "--project-root", r.dir, "somecommand"})
	err := cmd.Execute()
	var ec *exitCodeError
	if !errors.As(err, &ec) || ec.ExitCode() != 2 {
		t.Errorf("command without --: err %v, want exit code 2", err)
	}
}
