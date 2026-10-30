package web

// af_branch_coverage_test.go — the branches of the Agentic Factory web files that
// no other test reached (Issue #1948, B10). Each test asserts what the branch
// shows, not only that it ran.
//
// Defect answer: these tests catch a PR phase mapped to the wrong track, a check
// conclusion counted in the wrong bucket, a missing limit or backlog shown as a
// count, a blocked/parked intake item given the wrong status, a record's park,
// gate or refusal lost on the way to the row, and a restricted context item shown
// as linkable.

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

func TestDetailTextAndResumeText(t *testing.T) {
	if got := detailText(true, true, func() []string { return []string{"", ""} }); got != prGap {
		t.Errorf("a recorded detail with every part empty must read as a gap, got %q", got)
	}
	if got := detailText(true, true, func() []string { return []string{"a", "", "b"} }); got != "a · b" {
		t.Errorf("parts = %q", got)
	}
	if resumeText("") != "" || resumeText("x") != "resume: x" {
		t.Errorf("resumeText: %q / %q", resumeText(""), resumeText("x"))
	}
}

func TestPRTrackMapsEveryPhase(t *testing.T) {
	for phase, want := range map[string]string{
		"checks-pending": "checks", "checks-failed": "checks", "review-pending": "review",
		"repair-needed": "repair", "repair-dispatched": "repair",
		"ready-to-merge": "merge", "merged-issue-open": "merge", "issue-closed": "closed",
		"": "gap", "none": "gap", "something-new": "gap",
	} {
		if got := prTrack(phase); got != want {
			t.Errorf("prTrack(%q) = %q, want %q", phase, got, want)
		}
	}
}

func TestSummarizeChecksBuckets(t *testing.T) {
	row := func(conclusions ...string) monRow {
		var r monRow
		r.Checks.Confirmed = true
		for _, c := range conclusions {
			r.Checks.Value = append(r.Checks.Value, struct{ Name, Conclusion, HeadSHA string }{"n", c, "h"})
		}
		return r
	}
	for name, tc := range map[string]struct {
		r    monRow
		want string
	}{
		"mixed":         {row("success", "failure", "cancelled", "timed_out", "queued", ""), "1 passed, 3 failed, 2 pending"},
		"all pass":      {row("success", "success"), "2 passed"},
		"none required": {row(), "none required"},
	} {
		if got := summarizeChecks(tc.r); got != tc.want {
			t.Errorf("%s: %q, want %q", name, got, tc.want)
		}
	}
}

func TestPRMetaStates(t *testing.T) {
	if got := prMeta(PRTrackingVM{}); got != "no data" {
		t.Errorf("unavailable = %q", got)
	}
	if got := prMeta(PRTrackingVM{Available: true}); got != "backlog unknown" {
		t.Errorf("no backlog = %q", got)
	}
	vm := PRTrackingVM{Available: true, Backlog: PRBacklogVM{Known: true, Open: 4, State: "no-limit-set"}}
	if got := prMeta(vm); !strings.Contains(got, "open PRs 4") || !strings.Contains(got, "no limit set") || !strings.Contains(got, "repair not provided") {
		t.Errorf("no limit = %q", got)
	}
}

func TestIntakeRequestsFromNilAndEmptyStore(t *testing.T) {
	if _, err := IntakeRequestsFrom(nil)(context.Background()); err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Errorf("a nil store is not configured: %v", err)
	}
	s, dir := recordStore(t, 3)
	// A requests path that is a file, not a directory, is a store that cannot be read.
	broken, bdir := recordStore(t, 3)
	if err := os.WriteFile(filepath.Join(bdir, "requests"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := IntakeRequestsFrom(broken)(context.Background()); err == nil {
		t.Errorf("an unreadable request store must be an error, never an empty list")
	}
	_ = dir
	raw, err := IntakeRequestsFrom(s)(context.Background())
	if err != nil || strings.TrimSpace(string(raw)) != "[]" && strings.TrimSpace(string(raw)) != "null" {
		t.Errorf("an empty store reads as no requests: %q, %v", raw, err)
	}
}

func TestBuildIntakeBlockedAndParkedStatus(t *testing.T) {
	a := prApp(t, `{"rows":[],"intake":{"value":[
 {"key":"u","stage":"ACCEPTED","block":{"step":"s","detail":"d","unknown":true,"resume":"r"}},
 {"key":"b","stage":"ACCEPTED","block":{"step":"s","detail":"d","retryable":true,"resume":"r"}},
 {"key":"p","stage":"ACCEPTED","parked":"no-slot"},
 {"key":"w","stage":"ACCEPTED","state":"WT_PREPARED","parked":"backpressure"}
],"confirmed":true}}`, nil)
	by := map[string]IntakeRequestVM{}
	vm := a.buildPRTracking(context.Background()).Intake
	for _, r := range vm.Rows {
		by[r.Key] = r
	}
	if by["u"].Status != "EFFECT_UNKNOWN" || by["b"].Status != "BLOCKED" || by["p"].Status != "PARKED" {
		t.Errorf("statuses: u=%q b=%q p=%q", by["u"].Status, by["b"].Status, by["p"].Status)
	}
	if by["b"].Retryable != "yes" || by["u"].Retryable != "no" {
		t.Errorf("retryable: b=%q u=%q", by["b"].Retryable, by["u"].Retryable)
	}
	if by["w"].Status != "WT_PREPARED" {
		t.Errorf("a state the monitor reported is kept even when parked: %q", by["w"].Status)
	}
	if vm.BlockedCount != 2 || vm.ParkedCount != 2 {
		t.Errorf("counts: blocked %d parked %d", vm.BlockedCount, vm.ParkedCount)
	}
}

// A record's park, gate and refusal reach the row (the branches that fill them).
func TestRecordRowsCarryParkGateRefusal(t *testing.T) {
	src := fakeRecords{recs: []claim.Record{{
		Issue: 99, Holder: "s", AgentName: "a", State: core.StateLanding, Stage: core.StageMerging,
		Park:    &core.Park{Reason: core.BlockReason("backpressure"), Evidence: "6/5", Resume: "below limit"},
		Gate:    &core.Gate{Reason: core.BlockReason("checks-failed"), Detail: "lint red", Resume: "checks green"},
		Refusal: &core.Refusal{Reason: core.BlockReason("stale-head"), Detail: "head moved"},
	}}}
	r := recordApp(t, src).buildPRTracking(context.Background()).Rows[0]
	if r.Park != "backpressure · 6/5 · resume: below limit" {
		t.Errorf("park = %q", r.Park)
	}
	if r.Gate != "checks-failed · lint red · resume: checks green" {
		t.Errorf("gate = %q", r.Gate)
	}
	if r.Refusal != "stale-head · head moved" {
		t.Errorf("refusal = %q", r.Refusal)
	}
}

func TestContextModelRestrictedItemIsNotLinkable(t *testing.T) {
	a := prApp(t, cmRows, nil)
	a.contextModel = func(context.Context, int) ([]byte, error) {
		l := &ctxmodel.Ledger{Issue: 7, Items: []ctxmodel.Item{{Section: ctxmodel.SectionAudit, Title: "t", Owner: "o", Restricted: true, State: ctxmodel.StateRemoteVerified}}}
		return l.Marshal()
	}
	items := a.buildPRTracking(context.Background()).Rows[0].Ledger.Items
	if len(items) != 1 || items[0].Status != "restricted" {
		t.Errorf("a restricted item is restricted, never available: %+v", items)
	}
}

func TestMonitorHandlerServesAndIsReadOnly(t *testing.T) {
	a := prApp(t, prRowsJSON, nil)
	if rec := serveGet(t, a.routes(), "/monitor"); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `id="pr-tracking"`) {
		t.Fatalf("GET /monitor = %d, pr-tracking panel present: %v", rec.Code, strings.Contains(rec.Body.String(), `id="pr-tracking"`))
	}
	if rec := servePost(t, a.routes(), "/monitor", url.Values{}); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /monitor = %d, want 405", rec.Code)
	}
}
