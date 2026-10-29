package gitexec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ghintake"
)

// ghintake.Git takes this wrapper's Run as its Runner.
var _ ghintake.Runner = Exec{}.Run

// A fake `git` shell script in a temp directory is put first on PATH. It records its arguments (NUL
// separated), whether it ran at all, and behaves as FAKE_GIT_MODE says. The real git is never started and
// no remote is contacted.
const fakeGit = `#!/bin/sh
dir="$FAKE_GIT_DIR"
echo spawned > "$dir/spawned"
: > "$dir/args"
for a in "$@"; do printf '%s\0' "$a" >> "$dir/args"; done
case "$FAKE_GIT_MODE" in
  remote)   printf 'abc123\trefs/heads/WT-x\n'; exit 0;;
  noremote) exit 0;;
  tree)     printf 'worktree /w/main\nHEAD abc\nbranch refs/heads/main\n\nworktree /w/card\nHEAD def\nbranch refs/heads/WT-x\n'; exit 0;;
  ok)       printf 'OUT'; printf 'ERR' >&2; exit 0;;
  fail)     printf 'partial'; printf 'fatal: could not read from remote repository' >&2; exit 128;;
  sleep)    echo $$ > "$dir/pid"; exec sleep 30;;
  big)      head -c 100000 /dev/zero | tr '\0' 'x'; exit 0;;
  env)      printf '%s|%s' "$GIT_TERMINAL_PROMPT" "$EXTRA";;
esac
`

type fake struct {
	t   *testing.T
	dir string
}

func newFake(t *testing.T, mode string) *fake {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(fakeGit), 0o755); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GIT_DIR", dir)
	t.Setenv("FAKE_GIT_MODE", mode)
	return &fake{t: t, dir: dir}
}

func (f *fake) spawned() bool { _, err := os.Stat(filepath.Join(f.dir, "spawned")); return err == nil }

func (f *fake) args() []string {
	raw, err := os.ReadFile(filepath.Join(f.dir, "args"))
	if err != nil {
		f.t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(raw), "\x00"), "\x00")
}

func ex() Exec { return Exec{MaxOutput: 50_000, WaitDelay: time.Second} }

func TestTheTwoReadFormsRunWithTheirArgumentsUnchanged(t *testing.T) {
	for _, want := range [][]string{
		{"ls-remote", "--heads", "origin", "WT-card-branch"},
		{"ls-remote", "--heads", "origin"},
		{"ls-remote", "--heads", "upstream.2", "feature/a b"}, // a pattern is one argument, spaces and all
		{"worktree", "list"},
		{"worktree", "list", "--porcelain"},
		{"worktree", "list", "--porcelain", "-z"},
	} {
		f := newFake(t, "ok")
		out, err := ex().Run(t.Context(), want...)
		if err != nil || string(out) != "OUT" {
			t.Errorf("%v: out=%q err=%v", want, out, err)
			continue
		}
		got := f.args()
		if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Errorf("args = %q, want %q", got, want)
		}
	}
}

// Everything else is refused before a process exists: other subcommands, the options that make ls-remote or
// worktree do something else, a remote that is not a plain remote name (a URL, a path, the ext:: transport),
// and git's own leading options (-c, -C), which could change what a read does.
func TestAnythingElseIsRefusedAndNoProcessIsStarted(t *testing.T) {
	for _, a := range [][]string{
		{"ls-remote", "origin"}, // no --heads
		{"ls-remote", "--tags", "origin"},
		{"ls-remote", "--heads"}, // no remote
		{"ls-remote", "--heads", "https://example.invalid/r.git"},
		{"ls-remote", "--heads", "git@example.invalid:o/r.git"},
		{"ls-remote", "--heads", "ext::sh -c 'touch pwned'"},
		{"ls-remote", "--heads", "/local/path"},
		{"ls-remote", "--heads", "../other"},
		{"ls-remote", "--heads", "-oPROXY=x"},
		{"ls-remote", "--heads", "--upload-pack=touch pwned", "origin"},
		{"ls-remote", "--heads", "origin", "--upload-pack=x"},
		{"ls-remote", "--heads", "origin", "-o", "x"},
		{"ls-remote", "--heads", "origin", "a", "b"}, // one pattern at most
		{"ls-remote", "--heads", "origin", ""},
		{"ls-remote", "--heads", "origin", "bad\x00pattern"},
		{"worktree", "add", "/w/x"}, {"worktree", "remove", "/w/x"}, {"worktree", "prune"}, {"worktree", "lock", "/w/x"},
		{"worktree", "move", "a", "b"}, {"worktree", "repair"},
		{"worktree", "list", "--verbose"}, {"worktree", "list", "--expire=now"}, {"worktree", "list", "/w/x"},
		{"worktree", "list", "--porcelain", "--porcelain"},
		{"-c", "core.sshCommand=touch pwned", "ls-remote", "--heads", "origin"},
		{"-C", "/other", "worktree", "list"},
		{"--git-dir=/x", "worktree", "list"},
		{"fetch", "origin"}, {"push", "origin", "x"}, {"pull"}, {"clone", "x"}, {"checkout", "x"}, {"reset", "--hard"},
		{"config", "--global", "x", "y"}, {"branch", "-D", "x"}, {"remote", "add", "x", "y"}, {"status"}, {"log"}, {"ls-files"},
		{"ls-remote"}, {"worktree"}, {},
	} {
		f := newFake(t, "ok")
		if _, err := ex().Run(t.Context(), a...); !errors.Is(err, ErrNotAllowed) || f.spawned() {
			t.Errorf("%q: err=%v spawned=%v, want a refusal with no process", a, err, f.spawned())
		}
	}
}

func TestANonZeroExitIsAnErrorThatKeepsTheOutputAndNamesOnlyTheVerb(t *testing.T) {
	newFake(t, "fail")
	out, err := ex().Run(t.Context(), "ls-remote", "--heads", "origin", "SECRET-BRANCH-NAME")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 128 || !strings.Contains(ee.Stderr, "could not read") || string(out) != "partial" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	if strings.Contains(err.Error(), "SECRET-BRANCH-NAME") {
		t.Errorf("the error echoes the call's arguments: %v", err)
	}
}

func TestStdoutAndStderrAreKeptApart(t *testing.T) {
	newFake(t, "ok")
	out, err := ex().Run(t.Context(), "worktree", "list")
	if err != nil || string(out) != "OUT" {
		t.Errorf("out=%q err=%v", out, err)
	}
}

func TestTheProcessGetsANonInteractiveEnvironmentAndTheCallersExtras(t *testing.T) {
	newFake(t, "env")
	e := ex()
	e.Env = []string{"EXTRA=yes"}
	out, err := e.Run(t.Context(), "worktree", "list")
	if err != nil || string(out) != "0|yes" {
		t.Errorf("env seen = %q err=%v", out, err)
	}
}

func TestOutputBeyondTheLimitIsAnErrorAndAMissingLimitRefusesTheCall(t *testing.T) {
	newFake(t, "big")
	e := ex()
	e.MaxOutput = 1000
	if out, err := e.Run(t.Context(), "worktree", "list"); !errors.Is(err, ErrOutputTooLarge) || out != nil {
		t.Errorf("out=%d bytes err=%v", len(out), err)
	}
	f := newFake(t, "ok")
	e.MaxOutput = 0
	if _, err := e.Run(t.Context(), "worktree", "list"); err == nil || f.spawned() {
		t.Errorf("err=%v spawned=%v: an unconfigured limit must stop the call first", err, f.spawned())
	}
}

func TestBinaryNotFoundIsAnErrorNotAPanic(t *testing.T) {
	e := ex()
	e.Bin = filepath.Join(t.TempDir(), "no-such-git")
	if _, err := e.Run(t.Context(), "worktree", "list"); err == nil {
		t.Error("a missing program gave no error")
	}
}

func alive(pid int) bool { return syscall.Kill(pid, 0) == nil }

func readPID(dir string) (int, bool) {
	raw, err := os.ReadFile(filepath.Join(dir, "pid"))
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	return pid, err == nil && pid > 0
}

func cleanupPID(t *testing.T, f *fake) {
	t.Cleanup(func() { // only the recorded PID, only if it is still there
		if pid, ok := readPID(f.dir); ok && alive(pid) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})
}

func TestCancellingTheContextStopsTheProcess(t *testing.T) {
	f := newFake(t, "sleep")
	cleanupPID(t, f)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { _, err := ex().Run(ctx, "ls-remote", "--heads", "origin"); done <- err }()
	var pid int
	for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(10 * time.Millisecond) {
		if p, ok := readPID(f.dir); ok {
			pid = p
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the fake git never started")
		}
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return after the cancel")
	}
	for i := 0; alive(pid) && i < 200; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if alive(pid) {
		t.Errorf("process %d is still running after the cancel", pid)
	}
}

func TestAnExpiredDeadlineIsReportedAsSuch(t *testing.T) {
	f := newFake(t, "sleep")
	cleanupPID(t, f)
	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := ex().Run(ctx, "worktree", "list")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 10*time.Second {
		t.Errorf("err=%v after %v", err, time.Since(start))
	}
}

// Through the component that uses it: ghintake's branch and worktree lookups run on this wrapper and the
// fake git, and read what it prints.
func TestGhintakeLookupsRunThroughTheWrapper(t *testing.T) {
	lk := ghintake.Lookup{Git: ghintake.Git{Run: ex().Run}}
	newFake(t, "remote")
	res, err := lk.Lookup(t.Context(), effect.Intent{Kind: effect.BranchPrepare, Target: "WT-x"})
	if err != nil || !res.Found || !res.Authoritative {
		t.Errorf("branch present: %+v %v", res, err)
	}
	newFake(t, "noremote")
	res, err = lk.Lookup(t.Context(), effect.Intent{Kind: effect.BranchPrepare, Target: "WT-x"})
	if err != nil || res.Found || !res.Authoritative {
		t.Errorf("branch absent: %+v %v", res, err)
	}
	newFake(t, "tree")
	res, err = lk.Lookup(t.Context(), effect.Intent{Kind: effect.WorktreePrepare, Target: "/w/card"})
	if err != nil || !res.Found {
		t.Errorf("worktree present: %+v %v", res, err)
	}
	res, err = lk.Lookup(t.Context(), effect.Intent{Kind: effect.WorktreePrepare, Target: "/w/none"})
	if err != nil || res.Found || !res.Authoritative {
		t.Errorf("worktree absent: %+v %v", res, err)
	}
}
