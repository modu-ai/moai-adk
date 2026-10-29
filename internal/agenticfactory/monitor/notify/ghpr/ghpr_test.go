package ghpr

import (
	"context"
	"errors"
	"strings"
	"testing"
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
	return Source{Repo: "o/r", Run: f.run, Number: func(issue int) (int, bool, error) {
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
