package web

// intake_requests_test.go — the background intake request panel (Issue #1948,
// S51). Requests come from lane-18's claim store through its read-only
// ViewRequests, encoded as JSON and decoded by the web; the fixtures are real
// store records, not hand-written JSON.
//
// Defect answer: these tests catch a blocked or effect-unknown request shown as
// a plain stage, a parked request shown as running, a derived started flag
// shown as a stored stage, retired stage names reappearing, an agent display
// name shown as its session id, and an unreadable store shown as empty.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
)

func intakeStore(t *testing.T) *claim.Store {
	t.Helper()
	s, err := claim.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	pol := claim.Policy{MaxInFlight: 3, MaxParked: 2}
	for i, k := range []string{"req-accepted", "req-confirmed", "req-blocked", "req-unknown"} {
		if i == 3 {
			pol = claim.Policy{MaxInFlight: 4, MaxParked: 2}
		}
		if _, _, err := s.PutRequest(k, "hash-"+k, i, pol); err != nil {
			t.Fatalf("put %s: %v", k, err)
		}
	}
	// req-accepted..req-unknown are in flight (MaxInFlight 4 for the last);
	// one more request is parked behind the limit.
	if _, _, err := s.PutRequest("req-parked", "hash-parked", 9, claim.Policy{MaxInFlight: 4, MaxParked: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateRequest("req-confirmed", func(r *claim.Request) error {
		r.Stage = claim.StageIssueConfirmed
		r.Issue = 77
		r.Agent, r.AgentName = "sess-1111-aaaa", "agent-2"
		r.Evidence = append(r.Evidence, claim.StageEvidence{Stage: claim.StageIssueConfirmed, Ref: "issue#77"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateRequest("req-blocked", func(r *claim.Request) error {
		r.Blocked = &claim.Blocked{Step: claim.StageAccepted, Detail: "gh returned 429", Retryable: true, Resume: "wait for the rate limit"}
		r.Attempts = 2
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateRequest("req-unknown", func(r *claim.Request) error {
		r.Blocked = &claim.Blocked{Step: claim.StageAccepted, Detail: "send reply lost", Unknown: true, Resume: "FindIssue decides"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return s
}

func requestsApp(t *testing.T, s *claim.Store) *app {
	t.Helper()
	return newApp(Config{ProjectRoot: t.TempDir(), IntakeRequests: IntakeRequestsFrom(s)})
}

func TestIntakeRequests_RoundTripFromRealStore(t *testing.T) {
	vm := requestsApp(t, intakeStore(t)).buildIntakeRequests(context.Background())
	if !vm.Known || len(vm.Rows) != 5 {
		t.Fatalf("vm = %+v", vm)
	}
	by := map[string]IntakeRequestVM{}
	for _, r := range vm.Rows {
		by[r.Key] = r
	}
	if r := by["req-accepted"]; r.Status != "ACCEPTED" || r.Parked != "" || r.Started {
		t.Errorf("accepted = %+v", r)
	}
	if r := by["req-confirmed"]; r.Status != "ISSUE_CONFIRMED" || r.Issue != "#77" || r.AgentName != "agent-2" || r.Agent != "sess-1111-aaaa" || len(r.Evidence) != 2 {
		t.Errorf("confirmed = %+v", r)
	}
	if r := by["req-blocked"]; r.Status != "BLOCKED" || r.Retryable != "yes" || r.Attempts != 2 || r.BlockDetail != "gh returned 429" {
		t.Errorf("blocked must not read as its stage: %+v", r)
	}
	if r := by["req-unknown"]; r.Status != "EFFECT_UNKNOWN" || r.Retryable != "no" {
		t.Errorf("effect-unknown must be distinct from blocked: %+v", r)
	}
	if r := by["req-parked"]; r.Parked != "backpressure" || r.Status != "ACCEPTED" {
		t.Errorf("parked = %+v", r)
	}
	if vm.ParkedCount != 1 || vm.BlockedCount != 2 {
		t.Errorf("counts parked=%d blocked=%d", vm.ParkedCount, vm.BlockedCount)
	}
}

func TestIntakeRequests_RetiredStageNamesNeverAppear(t *testing.T) {
	vm := requestsApp(t, intakeStore(t)).buildIntakeRequests(context.Background())
	for _, r := range vm.Rows {
		switch r.Status {
		case "ACK", "ISSUE_PUBLISHED", "WT_READY", "STARTED":
			t.Errorf("retired stage name shown: %+v", r)
		}
	}
}

func TestIntakeRequests_StartedIsDerivedAndUnknownStageStaysVerbatim(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir(), IntakeRequests: func(context.Context) ([]byte, error) {
		return []byte(`[{"key":"k1","stage":"READY_ACKED","started":true,"started_ref":"attempt-1","evidence":[]},
		                 {"key":"k2","stage":"FUTURE_STAGE","evidence":[]}]`), nil
	}})
	vm := a.buildIntakeRequests(context.Background())
	if r := vm.Rows[0]; r.Status != "READY_ACKED" || !r.Started || r.StartedRef != "attempt-1" {
		t.Errorf("k1 = %+v", r)
	}
	if r := vm.Rows[1]; r.Status != "FUTURE_STAGE" || r.Started {
		t.Errorf("unknown stage must stay verbatim: %+v", r)
	}
}

func TestIntakeRequests_NotProvided(t *testing.T) {
	for name, cfg := range map[string]Config{
		"unwired":    {ProjectRoot: t.TempDir()},
		"read error": {ProjectRoot: t.TempDir(), IntakeRequests: func(context.Context) ([]byte, error) { return nil, errors.New("down") }},
		"bad json":   {ProjectRoot: t.TempDir(), IntakeRequests: func(context.Context) ([]byte, error) { return []byte("{"), nil }},
		"nil store":  {ProjectRoot: t.TempDir(), IntakeRequests: IntakeRequestsFrom(nil)},
	} {
		vm := newApp(cfg).buildIntakeRequests(context.Background())
		if vm.Known || vm.Note == "" || len(vm.Rows) != 0 {
			t.Errorf("%s: want not provided, got %+v", name, vm)
		}
	}
}

func TestMonitorRendersIntakeRequests(t *testing.T) {
	a := requestsApp(t, intakeStore(t))
	m, _ := a.buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	for _, want := range []string{
		`id="intake-requests"`, "BLOCKED", "EFFECT_UNKNOWN", "parked backpressure", "agent agent-2", "session sess-1111-aaaa",
		"retryable yes", "gh returned 429", "attempts 2", "issue#77",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor missing %q", want)
		}
	}
}
