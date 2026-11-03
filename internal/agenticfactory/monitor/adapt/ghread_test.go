package adapt

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

type fakeGH struct {
	out   map[string]string
	calls [][]string
	err   error
}

func (f *fakeGH) run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	if f.err != nil {
		return nil, f.err
	}
	return []byte(f.out[args[0]+" "+args[1]]), nil
}

const prJSON = `[
 {"number":54,"state":"CLOSED","headRefOid":"` + head + `","statusCheckRollup":[],"closingIssuesReferences":[{"number":7}]},
 {"number":55,"state":"OPEN","headRefOid":"` + head + `","statusCheckRollup":[{"name":"build","conclusion":"SUCCESS","status":"COMPLETED"},{"name":"lint","conclusion":"","status":"IN_PROGRESS"}],"closingIssuesReferences":[{"number":7}]},
 {"number":56,"state":"OPEN","headRefOid":"x","statusCheckRollup":[],"closingIssuesReferences":[{"number":8}]}
]`

func TestReadPRPicksNewestOpenPRForTheIssue(t *testing.T) {
	g := &fakeGH{out: map[string]string{"pr list": prJSON, "issue view": `{"state":"OPEN"}`}}
	r := GHReader{PRListLimit: 100, Run: g.run, Verdict: func(int, string) string { return "PASS" }, Repair: func(int) bool { return false }}
	info, found, err := r.ReadPR(t.Context(), 7)
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	if info.Number != 55 || !info.Open || info.Merged || info.IssueClosed || info.HeadSHA != head || info.ReviewVerdict != "PASS" {
		t.Fatalf("info = %+v", info)
	}
	if len(info.Checks) != 2 || info.Checks[0].Conclusion != "success" || info.Checks[1].Conclusion == "success" {
		t.Errorf("checks = %+v (an unfinished check must not read as success)", info.Checks)
	}
}

func TestReadPRNoneAndMerged(t *testing.T) {
	g := &fakeGH{out: map[string]string{"pr list": `[]`}}
	if _, found, err := (GHReader{PRListLimit: 100, Run: g.run}).ReadPR(t.Context(), 7); err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	g = &fakeGH{out: map[string]string{
		"pr list":    `[{"number":60,"state":"MERGED","headRefOid":"` + head + `","statusCheckRollup":[],"closingIssuesReferences":[{"number":7}]}]`,
		"issue view": `{"state":"CLOSED"}`,
	}}
	info, found, err := (GHReader{PRListLimit: 100, Run: g.run}).ReadPR(t.Context(), 7)
	if err != nil || !found || !info.Merged || !info.IssueClosed {
		t.Fatalf("info=%+v found=%v err=%v", info, found, err)
	}
}

func TestCountOpenPRs(t *testing.T) {
	g := &fakeGH{out: map[string]string{"pr list": `[{"number":1},{"number":2},{"number":3}]`}}
	n, err := (GHReader{PRListLimit: 100, Run: g.run}).CountOpenPRs(t.Context())
	if err != nil || n != 3 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestGHErrorsPropagateAsErrors(t *testing.T) {
	g := &fakeGH{err: errors.New("429")}
	if _, _, err := (GHReader{PRListLimit: 100, Run: g.run}).ReadPR(t.Context(), 7); err == nil {
		t.Error("gh error swallowed")
	}
	g = &fakeGH{out: map[string]string{"pr list": `not json`}}
	if _, _, err := (GHReader{PRListLimit: 100, Run: g.run}).ReadPR(t.Context(), 7); err == nil {
		t.Error("bad json accepted")
	}
}

func TestOnlyReadOnlyGHCommandsAreIssued(t *testing.T) {
	g := &fakeGH{out: map[string]string{"pr list": prJSON, "issue view": `{"state":"OPEN"}`}}
	r := GHReader{PRListLimit: 100, Run: g.run}
	_, _, _ = r.ReadPR(t.Context(), 7)
	_, _ = r.CountOpenPRs(t.Context())
	for _, c := range g.calls {
		if !readOnlyGH(c) {
			t.Errorf("non-read-only gh call: %v", c)
		}
	}
	for _, bad := range [][]string{{"pr", "merge", "1"}, {"pr", "close", "1"}, {"issue", "close", "1"}, {"api", "-X", "POST", "x"}, {"pr", "list", "--web"}} {
		if readOnlyGH(bad) {
			t.Errorf("allowed %v", bad)
		}
	}
	if !strings.Contains(strings.Join(g.calls[0], " "), "--json") {
		t.Error("expected machine-readable output")
	}
	var _ monitor.PRReader = r
}

// The list limit is configuration: without it nothing is asked, and a list that reaches it is not trusted.
func TestAListLimitIsRequiredAndAListThatReachesItIsAnError(t *testing.T) {
	g := &fakeGH{out: map[string]string{"pr list": `[{"number":1}]`}}
	if _, _, err := (GHReader{Run: g.run}).ReadPR(t.Context(), 7); err == nil {
		t.Error("ReadPR ran without a configured limit")
	}
	if _, err := (GHReader{Run: g.run}).CountOpenPRs(t.Context()); err == nil {
		t.Error("CountOpenPRs ran without a configured limit")
	}
	if len(g.calls) != 0 {
		t.Errorf("gh was called %d times with no limit configured", len(g.calls))
	}
	g = &fakeGH{out: map[string]string{"pr list": `[{"number":1},{"number":2},{"number":3}]`}}
	r := GHReader{PRListLimit: 3, Run: g.run}
	if _, err := r.CountOpenPRs(t.Context()); err == nil {
		t.Error("a count at the limit was reported as exact")
	}
	if _, found, err := r.ReadPR(t.Context(), 7); err == nil || found {
		t.Errorf("a list at the limit gave found=%v err=%v: it may have been cut short", found, err)
	}
	joined := strings.Join(g.calls[0], " ")
	if !strings.Contains(joined, "--limit 3") {
		t.Errorf("the configured limit was not passed to gh: %v", g.calls[0])
	}
}
