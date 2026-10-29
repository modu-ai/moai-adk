package ghpr

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// The outputs below are hand-written to the shape the gh manual documents (fields number, headRefOid,
// state, reviewDecision for `pr view`; name and bucket for `pr checks`; state for `issue view`). They are
// not captured from a real gh run, and this package's tests make no call to GitHub.

var head = strings.Repeat("a", 40)

func prView(state, decision string) []byte {
	return []byte(`{"number":9,"headRefOid":"` + head + `","state":"` + state + `","reviewDecision":"` + decision + `"}`)
}

const (
	allPass   = `[{"name":"build","bucket":"pass"},{"name":"lint","bucket":"skipping"}]`
	oneFail   = `[{"name":"build","bucket":"fail"},{"name":"lint","bucket":"pass"}]`
	onePend   = `[{"name":"build","bucket":"pending"},{"name":"lint","bucket":"pass"}]`
	cancel    = `[{"name":"build","bucket":"cancel"}]`
	issueOpen = `{"state":"OPEN"}`
)

func TestParseReadsCiReviewMergeAndIssueState(t *testing.T) {
	for name, tc := range map[string]struct {
		pr, checks, issue string
		ci, review        string
		merged, closed    bool
	}{
		"all pass":              {string(prView("OPEN", "")), allPass, issueOpen, "success", "", false, false},
		"a failing check":       {string(prView("OPEN", "")), oneFail, issueOpen, "failure", "", false, false},
		"failure beats pending": {string(prView("OPEN", "")), `[{"name":"a","bucket":"pending"},{"name":"b","bucket":"fail"}]`, issueOpen, "failure", "", false, false},
		"pending":               {string(prView("OPEN", "")), onePend, issueOpen, "pending", "", false, false},
		"cancelled":             {string(prView("OPEN", "")), cancel, issueOpen, "failure", "", false, false},
		"no required checks":    {string(prView("OPEN", "")), `[]`, issueOpen, "", "", false, false},
		"changes requested":     {string(prView("OPEN", "CHANGES_REQUESTED")), allPass, issueOpen, "success", "changes_requested", false, false},
		"approved":              {string(prView("OPEN", "APPROVED")), allPass, issueOpen, "success", "approved", false, false},
		"review required":       {string(prView("OPEN", "REVIEW_REQUIRED")), allPass, issueOpen, "success", "", false, false},
		"merged":                {string(prView("MERGED", "APPROVED")), allPass, `{"state":"CLOSED"}`, "success", "approved", true, true},
		"closed unmerged":       {string(prView("CLOSED", "")), allPass, issueOpen, "success", "", false, false},
	} {
		v, err := Parse([]byte(tc.pr), []byte(tc.checks), []byte(tc.issue))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if v.Number != 9 || v.Head != head || v.CI != tc.ci || v.Review != tc.review || v.Merged != tc.merged || v.IssueClosed != tc.closed {
			t.Errorf("%s: %+v", name, v)
		}
	}
}

// What cannot be read is an error, never a guess toward success.
func TestParseRefusesWhatItCannotRead(t *testing.T) {
	good := string(prView("OPEN", ""))
	for name, in := range map[string][3]string{
		"unknown bucket":   {good, `[{"name":"a","bucket":"weird"}]`, issueOpen},
		"short head":       {`{"number":9,"headRefOid":"abc","state":"OPEN"}`, allPass, issueOpen},
		"no number":        {`{"headRefOid":"` + head + `"}`, allPass, issueOpen},
		"pr not json":      {`nope`, allPass, issueOpen},
		"checks not json":  {good, `nope`, issueOpen},
		"issue not json":   {good, allPass, `nope`},
		"checks not array": {good, `{"name":"a"}`, issueOpen},
	} {
		if _, err := Parse([]byte(in[0]), []byte(in[1]), []byte(in[2])); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

type fakeRun struct {
	calls  [][]string
	out    map[string]string // by the first two args
	checks error             // returned beside the checks output
	fail   map[string]error
}

func (f *fakeRun) run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	key := args[0] + " " + args[1]
	if err := f.fail[key]; err != nil {
		return nil, err
	}
	if key == "pr checks" {
		return []byte(f.out[key]), f.checks
	}
	return []byte(f.out[key]), nil
}

func newSource(f *fakeRun) Source {
	return Source{Repo: "o/r", Run: f.run, Timeout: time.Second, Number: func(issue int) (int, bool, error) {
		if issue == 7 {
			return 9, true, nil
		}
		return 0, false, nil
	}}
}

func fixtures() *fakeRun {
	return &fakeRun{out: map[string]string{"pr view": string(prView("OPEN", "CHANGES_REQUESTED")), "pr checks": oneFail, "issue view": issueOpen}}
}

func TestSourceReadsThroughThreeReadOnlyCalls(t *testing.T) {
	f := fixtures()
	v, found, err := newSource(f).PR(t.Context(), 7)
	if err != nil || !found || v.CI != "failure" || v.Review != "changes_requested" || v.Number != 9 {
		t.Fatalf("view=%+v found=%v err=%v", v, found, err)
	}
	if len(f.calls) != 3 {
		t.Fatalf("calls = %v", f.calls)
	}
	for _, c := range f.calls {
		if !readOnly(c) {
			t.Errorf("a call outside the read-only list: %v", c)
		}
		joined := strings.Join(c, " ")
		for _, verb := range []string{"create", "edit", "merge", "close", "comment", "review", "ready", "reopen"} {
			if strings.Contains(" "+joined+" ", " "+verb+" ") {
				t.Errorf("a write verb %q in %v", verb, c)
			}
		}
		if !strings.Contains(joined, "--repo o/r") {
			t.Errorf("a call without the repository: %v", c)
		}
	}
}

func TestAnIssueWithoutARecordedPRMakesNoCall(t *testing.T) {
	f := fixtures()
	_, found, err := newSource(f).PR(t.Context(), 8)
	if found || err != nil || len(f.calls) != 0 {
		t.Errorf("found=%v err=%v calls=%v", found, err, f.calls)
	}
}

// gh exits non-zero while checks fail or are pending and still prints the JSON: that output is used, for
// that one call. With no output it is an error, and every other call's error stays an error.
func TestTheChecksExitCodeAlongsideOutputIsAcceptedOnlyForTheChecksCall(t *testing.T) {
	f := fixtures()
	f.checks = errors.New("exit status 8")
	if v, found, err := newSource(f).PR(t.Context(), 7); err != nil || !found || v.CI != "failure" {
		t.Errorf("checks error with output: %+v %v %v", v, found, err)
	}
	f = fixtures()
	f.out["pr checks"], f.checks = "", errors.New("exit status 1")
	if _, _, err := newSource(f).PR(t.Context(), 7); err == nil {
		t.Error("a checks error with no output was accepted")
	}
	for _, key := range []string{"pr view", "issue view"} {
		f = fixtures()
		f.fail = map[string]error{key: errors.New("boom")}
		if _, _, err := newSource(f).PR(t.Context(), 7); err == nil {
			t.Errorf("%s failure was swallowed", key)
		}
	}
}

func TestAnAnswerForAnotherPullRequestIsRefused(t *testing.T) {
	f := fixtures()
	f.out["pr view"] = `{"number":10,"headRefOid":"` + head + `","state":"OPEN"}`
	if _, _, err := newSource(f).PR(t.Context(), 7); err == nil {
		t.Error("an answer for PR #10 was taken for PR #9")
	}
}

func TestAWriteCallNeverReachesTheRunner(t *testing.T) {
	f := fixtures()
	s := newSource(f)
	for _, args := range [][]string{{"pr", "merge", "9"}, {"issue", "close", "7"}, {"pr", "comment", "9"}, {"api", "-X", "POST", "x"}, {"pr"}} {
		if _, err := s.read(t.Context(), args...); err == nil {
			t.Errorf("%v reached the runner", args)
		}
	}
	if len(f.calls) != 0 {
		t.Errorf("runner calls: %v", f.calls)
	}
}

func TestAnUnconfiguredSourceIsRefused(t *testing.T) {
	if _, _, err := (Source{}).PR(t.Context(), 7); err == nil {
		t.Error("a source with no runner was accepted")
	}
}

// The shapes below are what a real `gh` printed for open and merged pull requests of this project, read
// only, as measured and relayed by the leader (this lane did not run gh). The head is padded to 40
// characters here; the point is the shape: keys in gh's order, and a reviewDecision that is the EMPTY
// string while no review decision exists. That is "no decision", not an error.
func TestParseAcceptsTheShapesGhActuallyPrinted(t *testing.T) {
	pr := `{"headRefOid":"a1d7ba49` + strings.Repeat("0", 32) + `","number":1939,"reviewDecision":"","state":"OPEN"}`
	checks := `[{"bucket":"pass","name":"Lint"},{"bucket":"pass","name":"Build (linux/amd64)"},{"bucket":"pass","name":"Analyze (Go) (go)"}]`
	v, err := Parse([]byte(pr), []byte(checks), []byte(`{"state":"OPEN"}`))
	if err != nil {
		t.Fatalf("a real open PR with no review decision was refused: %v", err)
	}
	if v.Number != 1939 || v.Review != "" || v.CI != "success" || v.Merged || v.IssueClosed || !strings.HasPrefix(v.Head, "a1d7ba49") {
		t.Errorf("view = %+v", v)
	}
	merged := `{"headRefOid":"` + strings.Repeat("b", 40) + `","number":1931,"reviewDecision":"","state":"MERGED"}`
	if v, err := Parse([]byte(merged), []byte(`[]`), []byte(`{"state":"CLOSED"}`)); err != nil || !v.Merged || !v.IssueClosed || v.Review != "" {
		t.Errorf("a merged PR: %+v %v", v, err)
	}
}

// Every gh read of one PR lookup gets its own time limit: a hung call ends at the limit, and the next
// call starts with a fresh one rather than inheriting what is left.
func TestEachGHReadHasItsOwnTimeLimit(t *testing.T) {
	var deadlines []time.Duration
	run := func(ctx context.Context, args ...string) ([]byte, error) {
		d, ok := ctx.Deadline()
		if !ok {
			t.Errorf("%v ran without a deadline", args)
			return nil, errors.New("no deadline")
		}
		deadlines = append(deadlines, time.Until(d))
		return nil, errors.New("stop after the first call")
	}
	s := Source{Repo: "o/r", Run: run, Timeout: 10 * time.Second, Number: func(int) (int, bool, error) { return 9, true, nil }}
	if _, _, err := s.PR(t.Context(), 7); err == nil {
		t.Fatal("want the runner's error")
	}
	if len(deadlines) != 1 || deadlines[0] <= 0 || deadlines[0] > 10*time.Second {
		t.Errorf("deadlines = %v, want one within (0, 10s]", deadlines)
	}
}

func TestAHungGHReadEndsAtTheLimit(t *testing.T) {
	run := func(ctx context.Context, _ ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	s := Source{Repo: "o/r", Run: run, Timeout: 30 * time.Millisecond, Number: func(int) (int, bool, error) { return 9, true, nil }}
	start := time.Now()
	_, _, err := s.PR(t.Context(), 7)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 5*time.Second {
		t.Errorf("err = %v after %v, want a deadline error well under the test bound", err, time.Since(start))
	}
}

func TestAZeroTimeLimitIsRefused(t *testing.T) {
	s := Source{Repo: "o/r", Run: func(context.Context, ...string) ([]byte, error) { return nil, nil }, Number: func(int) (int, bool, error) { return 9, true, nil }}
	if _, _, err := s.PR(t.Context(), 7); err == nil {
		t.Error("a Source with no Timeout must be refused, not run unbounded")
	}
}
