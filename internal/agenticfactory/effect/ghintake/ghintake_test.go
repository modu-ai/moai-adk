package ghintake

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// fakeExec is the only gh/git the tests see. It records every call and never leaves the process.
type fakeExec struct {
	calls                  [][]string
	created                []string // bodies of issues "created" by the fake
	failCreateAfterLanding bool
	listErr                error
	out                    map[string]string
}

func (f *fakeExec) run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	key := args[0] + " " + args[1]
	switch key {
	case "issue create":
		body := ""
		for i, a := range args {
			if a == "--body" {
				body = args[i+1]
			}
		}
		f.created = append(f.created, body)
		if f.failCreateAfterLanding {
			return nil, errors.New("connection reset")
		}
		return []byte("https://github.com/o/r/issues/321\n"), nil
	case "issue list":
		if f.listErr != nil {
			return nil, f.listErr
		}
		var sb strings.Builder
		sb.WriteString("[")
		for i, b := range f.created {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(`{"number":321,"body":` + jsonStr(b) + `}`)
		}
		sb.WriteString("]")
		return []byte(sb.String()), nil
	}
	return []byte(f.out[key]), nil
}

func jsonStr(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)
	return `"` + r.Replace(s) + `"`
}

func newExec(f *fakeExec, allow bool) *effect.Executor {
	g := GH{Repo: "o/r", Run: f.run, AllowWrite: allow, ListLimit: 100}
	return &effect.Executor{
		Store: effect.NewMemIntentStore(), Send: g, Lookup: Lookup{GH: g, Git: Git{Run: f.run}},
		Timeouts: effect.Timeouts{IssueCreate: time.Second, BranchPrepare: time.Second, WorktreePrepare: time.Second},
	}
}

func TestWritesAreOffByDefaultAndNothingIsExecuted(t *testing.T) {
	f := &fakeExec{}
	in, err := newExec(f, false).Execute(t.Context(), IssueRequest(7, "o/r", "title", "body"))
	if err != nil {
		t.Fatal(err)
	}
	if in.Status != effect.StatusFailed || !strings.Contains(in.LastError, "disabled") {
		t.Errorf("intent = %+v", in)
	}
	if len(f.calls) != 0 {
		t.Errorf("gh/git were called %d times with writes disabled: %v", len(f.calls), f.calls)
	}
}

func TestEnabledCreateSendsOneIssueWithMarker(t *testing.T) {
	f := &fakeExec{}
	in, err := newExec(f, true).Execute(t.Context(), IssueRequest(7, "o/r", "title", "body"))
	if err != nil || in.Status != effect.StatusConfirmed || in.ProviderID != "321" {
		t.Fatalf("in=%+v err=%v", in, err)
	}
	if len(f.created) != 1 || effect.KeyFromMarker(f.created[0]) != in.Key {
		t.Errorf("created = %v", f.created)
	}
}

func TestLostCreateResponseIsFoundByBodyMarker(t *testing.T) {
	f := &fakeExec{failCreateAfterLanding: true}
	in, _ := newExec(f, true).Execute(t.Context(), IssueRequest(7, "o/r", "title", "body"))
	if in.Status != effect.StatusConfirmed || in.ProviderID != "321" || len(f.created) != 1 {
		t.Errorf("in=%+v created=%d", in, len(f.created))
	}
}

func TestAbsentAfterErrorIsRetryableAndUnsureIsParked(t *testing.T) {
	g := GH{Repo: "o/r", Run: func(ctx context.Context, args ...string) ([]byte, error) {
		if args[1] == "create" {
			return nil, errors.New("HTTP 502")
		}
		return []byte("[]"), nil
	}, AllowWrite: true, ListLimit: 100}
	now := time.Unix(1_000_000, 0)
	ex := &effect.Executor{Store: effect.NewMemIntentStore(), Send: g, Lookup: Lookup{GH: g}, Timeouts: effect.Timeouts{IssueCreate: time.Second}, Now: func() time.Time { return now }}
	in, _ := ex.Execute(t.Context(), IssueRequest(7, "o/r", "t", "b"))
	if in.Status != effect.StatusSent {
		t.Errorf("absent inside the window: %s, want sent (the create may still land)", in.Status)
	}
	now = now.Add(time.Hour)
	if in, _ := ex.Reconcile(t.Context(), in.Key); in.Status != effect.StatusFailed {
		t.Errorf("absent past the window: %s", in.Status)
	}
	f2 := &fakeExec{failCreateAfterLanding: true, listErr: errors.New("HTTP 429")}
	if in, _ := newExec(f2, true).Execute(t.Context(), IssueRequest(8, "o/r", "t", "b")); in.Status != effect.StatusParked {
		t.Errorf("unsure: %s", in.Status)
	}
	n := 0
	for _, c := range f2.calls {
		if c[1] == "create" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("creates after a 429 lookup = %d, want exactly 1 (no retry loop)", n)
	}
}

func TestOnlyAllowedCommandsPassTheGuard(t *testing.T) {
	ok := [][]string{{"issue", "list", "--repo", "o/r"}, {"issue", "view", "3"}, {"ls-remote", "--heads", "origin", "b"}, {"worktree", "list", "--porcelain"}}
	for _, a := range ok {
		if !readOnly(a) {
			t.Errorf("read-only call refused: %v", a)
		}
	}
	for _, a := range [][]string{{"issue", "close", "3"}, {"pr", "merge", "1"}, {"push", "origin", "b"}, {"api", "-X", "POST", "x"}, {"issue", "create"}, {"worktree", "remove", "x"}, {"ls-remote", "--upload-pack=evil", "origin"}} {
		if readOnly(a) {
			t.Errorf("non-read-only call allowed: %v", a)
		}
	}
}

func TestBranchAndWorktreeFactLookups(t *testing.T) {
	f := &fakeExec{out: map[string]string{
		"ls-remote --heads": "abc123\trefs/heads/WT-x\n",
		"worktree list":     "worktree /w/a\nHEAD abc\nbranch refs/heads/WT-a\n\nworktree /w/b\nHEAD def\nbranch refs/heads/WT-b\n",
	}}
	lk := Lookup{GH: GH{Repo: "o/r", Run: f.run, ListLimit: 100}, Git: Git{Run: f.run}}
	r, err := lk.Lookup(t.Context(), effect.Intent{Kind: effect.BranchPrepare, Target: "WT-x"})
	if err != nil || !r.Found {
		t.Fatalf("branch found: %+v err=%v", r, err)
	}
	r, _ = lk.Lookup(t.Context(), effect.Intent{Kind: effect.BranchPrepare, Target: "WT-none"})
	if r.Found {
		t.Error("branch with another name found")
	}
	r, err = lk.Lookup(t.Context(), effect.Intent{Kind: effect.WorktreePrepare, Target: "/w/b"})
	if err != nil || !r.Found || !r.Authoritative {
		t.Fatalf("worktree: %+v err=%v", r, err)
	}
	if r, _ = lk.Lookup(t.Context(), effect.Intent{Kind: effect.WorktreePrepare, Target: "/w/zzz"}); r.Found {
		t.Error("unknown worktree found")
	}
	if _, err := lk.Lookup(t.Context(), effect.Intent{Kind: effect.CheckObserve}); err == nil {
		t.Error("unsupported kind accepted")
	}
	for _, c := range f.calls {
		if !readOnly(c) {
			t.Errorf("lookup issued non-read-only call %v", c)
		}
	}
}

// The list limit is configuration: without it no list is asked for, and a list that reaches it proves
// nothing about what is absent.
func TestTheIssueListLimitIsRequiredAndALimitReachedIsNotAbsence(t *testing.T) {
	calls := 0
	run := func(_ context.Context, args ...string) ([]byte, error) {
		calls++
		return []byte(`[{"number":1,"body":"x"},{"number":2,"body":"y"}]`), nil
	}
	in := effect.Intent{Kind: effect.IssueCreate, Key: "k-not-there"}
	if _, err := (Lookup{GH: GH{Repo: "o/r", Run: run}}).Lookup(t.Context(), in); err == nil || calls != 0 {
		t.Errorf("without a ListLimit: err=%v calls=%d", err, calls)
	}
	res, err := (Lookup{GH: GH{Repo: "o/r", Run: run, ListLimit: 2}}).Lookup(t.Context(), in)
	if err == nil || res.Authoritative {
		t.Errorf("a list at its limit with no match was taken as absence: %+v %v", res, err)
	}
	res, err = (Lookup{GH: GH{Repo: "o/r", Run: run, ListLimit: 3}}).Lookup(t.Context(), in)
	if err != nil || !res.Authoritative || res.Found {
		t.Errorf("a list below its limit is authoritative absence: %+v %v", res, err)
	}
}

// B4: nothing sensitive leaves in an Issue title or body, masked or not — it is refused before gh runs.
func TestASensitiveTitleOrBodyIsRefusedBeforeGHRuns(t *testing.T) {
	for name, tc := range map[string]struct{ title, body string }{
		"token in the body":   {"title", "see ghp_abcdefghijklmnopqrstuvwxyz0123456789 for access"},
		"email in the body":   {"title", "reported by someone@example.com"},
		"home path in body":   {"title", "log at /Users/someone/project/out.log"},
		"credential in title": {"password=hunter2hunter2", "body"},
	} {
		t.Run(name, func(t *testing.T) {
			f := &fakeExec{}
			in, err := newExec(f, true).Execute(t.Context(), IssueRequest(7, "o/r", tc.title, tc.body))
			if err != nil {
				t.Fatal(err)
			}
			if len(f.calls) != 0 || len(f.created) != 0 {
				t.Errorf("gh ran for a sensitive issue: calls=%v", f.calls)
			}
			if in.Status == effect.StatusConfirmed {
				t.Errorf("intent = %+v, want it not confirmed", in)
			}
			if strings.Contains(in.LastError, "ghp_") || strings.Contains(in.LastError, "someone@") || strings.Contains(in.LastError, "hunter2") {
				t.Errorf("the refusal repeats the sensitive value: %q", in.LastError)
			}
		})
	}
}
