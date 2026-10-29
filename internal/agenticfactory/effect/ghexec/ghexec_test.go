package ghexec

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ghintake"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/adapt"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/notify/ghpr"
)

// One Exec serves every component's Runner type.
var (
	_ ghpr.Runner      = Exec{}.Run
	_ ghintake.Runner  = Exec{}.Run
	_ adapt.GHRunner   = Exec{}.Run
	_ ctxledger.Runner = Exec{}.Run
)

// A fake `gh` shell script in a temp directory is put first on PATH. It records its arguments (NUL
// separated), its environment values and whether it ran at all, and behaves as FAKE_GH_MODE says. The real
// gh is never started and there is no network.
const fakeGH = `#!/bin/sh
dir="$FAKE_GH_DIR"
echo spawned > "$dir/spawned"
: > "$dir/args"
for a in "$@"; do printf '%s\0' "$a" >> "$dir/args"; done
case "$FAKE_GH_MODE" in
  ok)    printf 'OUT'; printf 'ERR' >&2; exit 0;;
  fail)  printf 'partial'; printf 'gh: Not Found (HTTP 404)' >&2; exit 1;;
  pend)  printf '[{"bucket":"pending","name":"build"}]'; printf 'some checks pending' >&2; exit 8;;
  sleep) echo $$ > "$dir/pid"; exec sleep 30;;
  big)   head -c 100000 /dev/zero | tr '\0' 'x'; exit 0;;
  env)   printf '%s|%s|%s|%s' "$GH_PROMPT_DISABLED" "$GH_NO_UPDATE_NOTIFIER" "$NO_COLOR" "$EXTRA";;
  pr)    case "$1 $2" in
           "pr view")    printf '{"headRefOid":"%s","number":9,"reviewDecision":"","state":"OPEN"}' "$(printf 'a%.0s' $(seq 1 40))";;
           "pr checks")  printf '[{"bucket":"fail","name":"build"}]'; exit 1;;
           "issue view") printf '{"state":"OPEN"}';;
         esac;;
esac
`

type fake struct {
	t   *testing.T
	dir string
}

func newFake(t *testing.T, mode string) *fake {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(fakeGH), 0o755); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_DIR", dir)
	t.Setenv("FAKE_GH_MODE", mode)
	return &fake{t: t, dir: dir}
}

func (f *fake) spawned() bool {
	_, err := os.Stat(filepath.Join(f.dir, "spawned"))
	return err == nil
}

func (f *fake) args() []string {
	raw, err := os.ReadFile(filepath.Join(f.dir, "args"))
	if err != nil {
		f.t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(raw), "\x00"), "\x00")
}

func ex() Exec { return Exec{MaxOutput: 50_000, WaitDelay: time.Second} }

// Arguments reach the process exactly as given - spaces, quotes, dollar signs, newlines - because no shell
// is involved.
func TestArgumentsArriveVerbatimWithoutAShell(t *testing.T) {
	f := newFake(t, "ok")
	want := []string{"issue", "view", "7", "--repo", "o/r; touch pwned", "--json", "$(echo x) `y` \"q\" 'z'\nnewline"}
	if _, err := ex().Run(t.Context(), want...); err != nil {
		t.Fatal(err)
	}
	got := f.args()
	if len(got) != len(want) {
		t.Fatalf("args = %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("arg %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestStdoutAndStderrAreKeptApart(t *testing.T) {
	newFake(t, "ok")
	out, err := ex().Run(t.Context(), "pr", "view", "9")
	if err != nil || string(out) != "OUT" {
		t.Errorf("out=%q err=%v: stderr must not mix into stdout", out, err)
	}
}

// A non-zero exit is an error that carries the exit code and the standard error, and the standard output
// is still returned (gh pr checks prints its JSON and exits non-zero while checks fail or are pending).
func TestANonZeroExitIsAnErrorThatKeepsTheOutput(t *testing.T) {
	newFake(t, "pend")
	out, err := ex().Run(t.Context(), "pr", "checks", "9", "--required", "--json", "name,bucket")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 8 || !strings.Contains(ee.Stderr, "pending") {
		t.Fatalf("err = %v", err)
	}
	if !bytes.Contains(out, []byte(`"pending"`)) {
		t.Errorf("the output was lost with the error: %q", out)
	}
}

// gh api reports an HTTP failure on standard error; the wrapper returns it as the typed forge.StatusError
// the context ledger decides on.
func TestAnHTTPFailureIsTheTypedStatusError(t *testing.T) {
	newFake(t, "fail")
	_, err := ex().Run(t.Context(), "api", "repos/o/r/issues/comments/5")
	var se *forge.StatusError
	if !errors.As(err, &se) || se.Status != 404 {
		t.Fatalf("err = %v, want a *forge.StatusError with status 404", err)
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 1 {
		t.Errorf("the exit code is lost: %v", err)
	}
}

// The error names the verb (the first two arguments) and never the rest of the call: a comment body or a
// title may carry text that must not travel into logs.
func TestAnErrorDoesNotEchoTheBodyOfTheCall(t *testing.T) {
	newFake(t, "fail")
	e := ex()
	e.AllowWrite = true
	_, err := e.Run(t.Context(), "issue", "create", "--repo", "o/r", "--title", "t", "--body", "PRIVATE-ISSUE-TEXT")
	var ee *ExitError
	if !errors.As(err, &ee) || strings.Contains(err.Error(), "PRIVATE-ISSUE-TEXT") {
		t.Errorf("err = %v", err)
	}
}

func TestBinaryNotFoundIsAnErrorNotAPanic(t *testing.T) {
	e := ex()
	e.Bin = filepath.Join(t.TempDir(), "no-such-gh")
	if _, err := e.Run(t.Context(), "pr", "view", "9"); err == nil {
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

// Cancelling the context stops the process, and Run returns the context's error.
func TestCancellingTheContextStopsTheProcess(t *testing.T) {
	f := newFake(t, "sleep")
	t.Cleanup(func() { // only the recorded PID, only if it is still there
		if pid, ok := readPID(f.dir); ok && alive(pid) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { _, err := ex().Run(ctx, "pr", "view", "9"); done <- err }()
	var pid int
	deadline := time.Now().Add(10 * time.Second)
	for {
		if p, ok := readPID(f.dir); ok {
			pid = p
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the fake gh never started")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !alive(pid) {
		t.Fatal("the process is not running before the cancel")
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
	t.Cleanup(func() {
		if pid, ok := readPID(f.dir); ok && alive(pid) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := ex().Run(ctx, "pr", "view", "9")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 10*time.Second {
		t.Errorf("err=%v after %v", err, time.Since(start))
	}
}

// Output beyond the limit stops the process and is an error; nothing partial is passed on as complete.
// A limit that is not configured refuses the call before any process exists.
func TestOutputBeyondTheLimitIsAnErrorAndAMissingLimitRefusesTheCall(t *testing.T) {
	f := newFake(t, "big")
	e := ex()
	e.MaxOutput = 1000
	out, err := e.Run(t.Context(), "pr", "view", "9")
	if !errors.Is(err, ErrOutputTooLarge) || out != nil {
		t.Errorf("out=%d bytes err=%v", len(out), err)
	}
	f2 := newFake(t, "ok")
	e.MaxOutput = 0
	if _, err := e.Run(t.Context(), "pr", "view", "9"); err == nil || f2.spawned() {
		t.Errorf("err=%v spawned=%v: an unconfigured limit must stop the call first", err, f2.spawned())
	}
	_ = f
}

func TestTheProcessGetsTheNonInteractiveEnvironmentAndTheCallersExtras(t *testing.T) {
	newFake(t, "env")
	e := ex()
	e.Env = []string{"EXTRA=yes"}
	out, err := e.Run(t.Context(), "pr", "view", "9")
	if err != nil || string(out) != "1|1|1|yes" {
		t.Errorf("env seen = %q err=%v", out, err)
	}
}

// The allowlist: reads run; writes and everything unknown are refused before a process exists, and the
// writes this code base makes need AllowWrite and nothing else ever runs.
func TestOnlyReadsRunByDefaultAndNoProcessIsStartedForARefusedCall(t *testing.T) {
	reads := [][]string{
		{"pr", "view", "9", "--json", "number"},
		{"pr", "list", "--state", "open", "--limit", "10"},
		{"pr", "checks", "9", "--required", "--json", "name,bucket"},
		{"issue", "view", "7", "--json", "state"},
		{"issue", "list", "--search", "x"},
		{"api", "repos/o/r/issues/7/comments", "--paginate", "--jq", ".[] | {id: .id}"},
		{"api", "-X", "GET", "repos/o/r/issues/7/comments"},
	}
	for _, a := range reads {
		f := newFake(t, "ok")
		if _, err := ex().Run(t.Context(), a...); err != nil || !f.spawned() {
			t.Errorf("read %v: err=%v spawned=%v", a, err, f.spawned())
		}
	}
	writes := [][]string{
		{"issue", "create", "--repo", "o/r", "--title", "t", "--body", "b"},
		{"api", "-X", "POST", "repos/o/r/issues/7/comments", "-f", "body=x"},
		{"api", "-X", "PATCH", "repos/o/r/issues/comments/5", "-f", "body=x"},
		{"pr", "merge", "9"}, {"pr", "close", "9"}, {"pr", "create"}, {"pr", "comment", "9"}, {"pr", "review", "9"},
		{"issue", "close", "7"}, {"issue", "edit", "7"}, {"issue", "comment", "7"},
		{"api", "repos/o/r/issues/7/comments", "-f", "body=x"}, // -f turns a bare call into a POST
		{"api", "-X", "DELETE", "repos/o/r/issues/comments/5"},
		{"api", "-X", "POST", "repos/o/r/pulls/9/merge"},
		{"api", "--unknown-flag", "repos/o/r/issues"},
		{"pr", "view", "9", "--web"}, {"auth", "login"}, {"repo", "delete", "o/r"}, {"pr"},
	}
	for _, a := range writes {
		f := newFake(t, "ok")
		if _, err := ex().Run(t.Context(), a...); !errors.Is(err, ErrNotAllowed) || f.spawned() {
			t.Errorf("%v: err=%v spawned=%v, want a refusal with no process", a, err, f.spawned())
		}
	}
}

func TestAllowWriteOpensCreatingAnIssueAndNothingElse(t *testing.T) {
	e := ex()
	e.AllowWrite = true
	f := newFake(t, "ok")
	if _, err := e.Run(t.Context(), "issue", "create", "--repo", "o/r", "--title", "t", "--body", "b"); err != nil || !f.spawned() {
		t.Errorf("issue create: err=%v spawned=%v", err, f.spawned())
	}
	for _, a := range [][]string{
		{"pr", "merge", "9"}, {"issue", "close", "7"}, {"issue", "edit", "7"},
		{"api", "-X", "POST", "repos/o/r/pulls/9/merge"}, {"api", "-X", "DELETE", "repos/o/r/issues/comments/5"},
		{"api", "-X", "PATCH", "repos/o/r/issues/7"}, // the Issue itself
		// comment writes need the author policy's identity decision, which this layer cannot make
		{"api", "-X", "POST", "repos/o/r/issues/7/comments", "-f", "body=x"},
		{"api", "-X", "PATCH", "repos/o/r/issues/comments/5", "-f", "body=x"},
	} {
		f := newFake(t, "ok")
		if _, err := e.Run(t.Context(), a...); !errors.Is(err, ErrNotAllowed) || f.spawned() {
			t.Errorf("%v with AllowWrite: err=%v spawned=%v", a, err, f.spawned())
		}
	}
}

// End to end through a real component: ghpr reads a PR through this wrapper and the fake gh, including the
// non-zero exit of `pr checks` that still carries its JSON.
func TestTheGhprSourceReadsThroughTheWrapper(t *testing.T) {
	newFake(t, "pr")
	src := ghpr.Source{Repo: "o/r", Run: ex().Run, Number: func(int) (int, bool, error) { return 9, true, nil }}
	v, found, err := src.PR(t.Context(), 7)
	if err != nil || !found || v.Number != 9 || v.CI != "failure" || v.Review != "" {
		t.Errorf("view=%+v found=%v err=%v", v, found, err)
	}
}

// The last gate also runs the protection guard: a write aimed at a protected path or without a repository
// to check is refused before a process exists.
func TestWritesPassTheProtectionGuardAtTheLastGate(t *testing.T) {
	e := ex()
	e.AllowWrite = true
	for _, a := range [][]string{
		{"issue", "create", "--repo", "o/r/branches/main/protection/x", "--title", "t", "--body", "b"},
		{"issue", "create", "--title", "t", "--body", "b"}, // no repository: nothing to check
	} {
		f := newFake(t, "ok")
		if _, err := e.Run(t.Context(), a...); !errors.Is(err, ErrNotAllowed) || f.spawned() {
			t.Errorf("%v: err=%v spawned=%v", a, err, f.spawned())
		}
	}
	f := newFake(t, "ok")
	if _, err := e.Run(t.Context(), "issue", "create", "--repo=o/r", "--title", "t", "--body", "b"); err != nil || !f.spawned() {
		t.Errorf("a plain write to a plain repository: err=%v spawned=%v", err, f.spawned())
	}
}
