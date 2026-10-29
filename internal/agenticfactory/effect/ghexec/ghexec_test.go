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

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
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
  comment) if [ "$2" = "-X" ]; then printf '{"id":55}'; else printf '{"id":5,"body":"hello MARK"}\n'; fi;;
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
	src := ghpr.Source{Repo: "o/r", Run: ex().Run, Timeout: time.Second, Number: func(int) (int, bool, error) { return 9, true, nil }}
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

// --- the context ledger's comment writes: Enforcer first, then this wrapper ----------------------------

const me, other = int64(1001), int64(2002)

type ident struct {
	user     int64
	userErr  error
	issues   map[int]effect.IssueFacts
	comments map[int64]effect.CommentFacts
}

func (f ident) CurrentUserID(context.Context) (int64, error) { return f.user, f.userErr }
func (f ident) IssueFacts(_ context.Context, _ string, n int) (effect.IssueFacts, error) {
	v, ok := f.issues[n]
	if !ok {
		return effect.IssueFacts{}, errors.New("lookup failed")
	}
	return v, nil
}
func (f ident) CommentFacts(_ context.Context, _ string, c int64) (effect.CommentFacts, error) {
	v, ok := f.comments[c]
	if !ok {
		return effect.CommentFacts{}, errors.New("lookup failed")
	}
	return v, nil
}

// issue 7 is ours, issue 8 is someone else's; comment 501 is ours, 502 is theirs (both on issue 8).
func world() ident {
	return ident{
		user: me,
		issues: map[int]effect.IssueFacts{
			7: {AuthorID: me, ActualRepo: "o/r"}, 8: {AuthorID: other, ActualRepo: "o/r"},
		},
		comments: map[int64]effect.CommentFacts{
			501: {AuthorID: me, Issue: 8, ActualRepo: "o/r"}, 502: {AuthorID: other, Issue: 8, ActualRepo: "o/r"},
		},
	}
}

func remote(write bool, id effect.Identity) ctxledger.GHRemote {
	e := ex()
	e.AllowWrite = write
	var enf *effect.Enforcer
	if id != nil {
		enf = &effect.Enforcer{Identity: id}
	}
	return ctxledger.GHRemote{Repo: "o/r", Run: e.Run, Enforcer: enf}
}

// Own Issue, someone else's Issue: a comment may be created on both (comments are the one write the policy
// gives on another author's Issue); editing is for our own comments only; the process starts only when the
// Enforcer approved and the wrapper may write.
func TestLedgerCommentWritesPassTheEnforcerThenRunThroughTheWrapper(t *testing.T) {
	f := newFake(t, "comment")
	r := remote(true, world())
	c, err := r.Create(t.Context(), 7, "a ledger body")
	if err != nil || c.ID != 55 || !f.spawned() {
		t.Fatalf("own issue: comment=%+v err=%v spawned=%v", c, err, f.spawned())
	}
	got := f.args()
	if strings.Join(got[:4], " ") != "api -X POST repos/o/r/issues/7/comments" || got[4] != "-f" || got[5] != "body=a ledger body" {
		t.Errorf("args = %q", got)
	}
	f = newFake(t, "comment")
	if _, err := r.Create(t.Context(), 8, "on their issue"); err != nil || !f.spawned() {
		t.Errorf("another author's issue: err=%v spawned=%v", err, f.spawned())
	}
	f = newFake(t, "comment")
	if err := r.Update(t.Context(), 8, 501, "our comment"); err != nil || !f.spawned() {
		t.Errorf("edit of our comment: err=%v spawned=%v", err, f.spawned())
	}
	f = newFake(t, "comment")
	err = r.Update(t.Context(), 8, 502, "their comment")
	var re *effect.RefusedError
	if !errors.As(err, &re) || f.spawned() {
		t.Errorf("edit of their comment: err=%v spawned=%v, want a refusal and no process", err, f.spawned())
	}
	f = newFake(t, "comment")
	if _, found, err := r.Find(t.Context(), 7, "MARK"); err != nil || !found || !f.spawned() {
		t.Errorf("a read needs no approval: found=%v err=%v spawned=%v", found, err, f.spawned())
	}
}

// Editing the body of an Issue is not something this path does: the policy refuses it on someone else's
// Issue and the wrapper refuses it everywhere, whoever asks.
func TestAnIssueBodyEditNeverRuns(t *testing.T) {
	f := newFake(t, "comment")
	enf := &effect.Enforcer{Identity: world()}
	body := []byte(`{"body":"changed"}`)
	if _, err := enf.Approve(t.Context(), "PATCH", "/repos/o/r/issues/8", nil, body); err == nil {
		t.Error("the policy approved editing someone else's Issue body")
	}
	e := ex()
	e.AllowWrite = true
	ctx, err := enf.Approve(t.Context(), "PATCH", "/repos/o/r/issues/7", nil, body) // ours: the policy allows it
	if err != nil {
		t.Fatal(err)
	}
	for _, issue := range []string{"7", "8"} {
		if _, err := e.Run(ctx, "api", "-X", "PATCH", "repos/o/r/issues/"+issue, "-f", "body=changed"); !errors.Is(err, ErrNotAllowed) {
			t.Errorf("issue %s: err=%v, want the wrapper to refuse", issue, err)
		}
	}
	if f.spawned() {
		t.Error("a process was started for an Issue body edit")
	}
}

// Without an identity nothing is written: no lookup answer, no Enforcer, no identity provider.
func TestUnknownIdentityRefusesTheWriteAndStartsNoProcess(t *testing.T) {
	unknownUser := world()
	unknownUser.user, unknownUser.userErr = 0, errors.New("no user")
	unknownAuthor := world()
	unknownAuthor.issues = map[int]effect.IssueFacts{}
	for name, r := range map[string]ctxledger.GHRemote{
		"the user cannot be established": remote(true, unknownUser),
		"the Issue author is unknown":    remote(true, unknownAuthor),
		"no Enforcer":                    remote(true, nil),
		"an Enforcer with no identity":   {Repo: "o/r", Run: ex().Run, Enforcer: &effect.Enforcer{}},
	} {
		f := newFake(t, "comment")
		if _, err := r.Create(t.Context(), 7, "x"); err == nil || f.spawned() {
			t.Errorf("%s: err=%v spawned=%v", name, err, f.spawned())
		}
	}
}

// AllowWrite is the wrapper's own switch: off, nothing starts even for a write the policy approved.
func TestWithAllowWriteOffNoProcessStartsEvenForAnApprovedWrite(t *testing.T) {
	f := newFake(t, "comment")
	_, err := remote(false, world()).Create(t.Context(), 7, "x")
	if !errors.Is(err, ErrNotAllowed) || f.spawned() {
		t.Errorf("err=%v spawned=%v", err, f.spawned())
	}
}

// The proof is for one write: it does not open another issue, another verb, or a call with no proof.
func TestAnApprovalOpensOnlyTheWriteItWasGivenFor(t *testing.T) {
	enf := &effect.Enforcer{Identity: world()}
	e := ex()
	e.AllowWrite = true
	ctx, err := enf.Approve(t.Context(), "POST", "/repos/o/r/issues/7/comments", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		ctx  context.Context
		args []string
	}{
		"another issue":   {ctx, []string{"api", "-X", "POST", "repos/o/r/issues/8/comments", "-f", "body=x"}},
		"another verb":    {ctx, []string{"api", "-X", "PATCH", "repos/o/r/issues/7/comments", "-f", "body=x"}},
		"another repo":    {ctx, []string{"api", "-X", "POST", "repos/x/y/issues/7/comments", "-f", "body=x"}},
		"no proof at all": {t.Context(), []string{"api", "-X", "POST", "repos/o/r/issues/7/comments", "-f", "body=x"}},
	} {
		f := newFake(t, "comment")
		if _, err := e.Run(tc.ctx, tc.args...); !errors.Is(err, ErrNotAllowed) || f.spawned() {
			t.Errorf("%s: err=%v spawned=%v", name, err, f.spawned())
		}
	}
	f := newFake(t, "comment")
	if _, err := e.Run(ctx, "api", "-X", "POST", "repos/o/r/issues/7/comments", "-f", "body=x"); err != nil || !f.spawned() {
		t.Errorf("the approved write itself: err=%v spawned=%v", err, f.spawned())
	}
}
