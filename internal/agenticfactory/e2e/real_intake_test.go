package e2e

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/intake"
)

// Background intake on lane-18's intake package and claim store (S51 stages).
// Publisher, Brancher and Preparer are counting fakes: no GitHub, git or
// filesystem effect happens here. Limits and timeouts are test inputs.

type fakePub struct {
	mu           sync.Mutex
	publishes    int
	finds        int
	created      map[string]int // request key -> issue that now exists on the "remote"
	next         int
	publishErr   error // returned by PublishIssue
	createOnErr  bool  // the Issue DOES get created although PublishIssue returns an error (lost response)
	findErr      error
	block        bool // PublishIssue waits for its context
	hangForeverC chan struct{}
}

func (p *fakePub) PublishIssue(ctx context.Context, req claim.Request) (int, string, error) {
	p.mu.Lock()
	p.publishes++
	if p.created == nil {
		p.created = map[string]int{}
	}
	if p.next == 0 {
		p.next = 2000
	}
	if p.block {
		p.mu.Unlock()
		<-ctx.Done()
		return 0, "", ctx.Err()
	}
	if p.publishErr != nil {
		if p.createOnErr {
			p.next++
			p.created[req.Key] = p.next
		}
		err := p.publishErr
		p.mu.Unlock()
		return 0, "", err
	}
	p.next++
	p.created[req.Key] = p.next
	n := p.next
	p.mu.Unlock()
	return n, fmt.Sprintf("issue/%d", n), nil
}

func (p *fakePub) FindIssue(_ context.Context, req claim.Request) (int, string, bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.finds++
	if p.findErr != nil {
		return 0, "", false, p.findErr
	}
	n, ok := p.created[req.Key]
	return n, fmt.Sprintf("issue/%d", n), ok, nil
}

func (p *fakePub) counts() (int, int) { p.mu.Lock(); defer p.mu.Unlock(); return p.publishes, p.finds }

type fakeBrancher struct {
	mu    sync.Mutex
	calls int
}

func (b *fakeBrancher) RegisterBranch(_ context.Context, req claim.Request) (claim.Branch, error) {
	b.mu.Lock()
	b.calls++
	b.mu.Unlock()
	return claim.Branch{Ref: "WT-" + req.Key, BaseOID: headA}, nil
}

func (b *fakeBrancher) FindBranch(_ context.Context, req claim.Request) (claim.Branch, bool, error) {
	return claim.Branch{Ref: "WT-" + req.Key, BaseOID: headA}, true, nil
}

type fakePrep struct {
	mu            sync.Mutex
	calls         int
	writerStopped bool
}

func (p *fakePrep) PrepareWorktree(_ context.Context, req claim.Request, slot string) (intake.Ready, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return intake.Ready{Host: "host-1", Path: "/work/" + slot + "/" + req.Key, Head: headA, WriterStopped: p.writerStopped}, nil
}

type intakeRig struct {
	dir  string
	st   *claim.Store
	in   *intake.Intake
	pub  *fakePub
	br   *fakeBrancher
	prep *fakePrep
	cfg  intake.Config
}

func newRig(t *testing.T, cfg intake.Config) *intakeRig {
	t.Helper()
	dir := t.TempDir()
	st, err := claim.OpenWithLimit(dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	r := &intakeRig{dir: dir, st: st, pub: &fakePub{}, br: &fakeBrancher{}, prep: &fakePrep{writerStopped: true}, cfg: cfg}
	r.in = r.intakeOver(t, st)
	return r
}

func (r *intakeRig) intakeOver(t *testing.T, st *claim.Store) *intake.Intake {
	t.Helper()
	in, err := intake.New(st, r.pub, r.br, r.prep, r.cfg)
	if err != nil {
		t.Fatal(err)
	}
	return in
}

var rigCfg = intake.Config{MaxInFlight: 2, MaxParked: 1, Slots: []string{"slot-1", "slot-2"}, StepTimeout: 2 * time.Second}

func advance(t *testing.T, in *intake.Intake, key string, want claim.Stage) claim.Request {
	t.Helper()
	req, err := in.Advance(context.Background(), key)
	if err != nil {
		t.Fatalf("advance %s: %v", key, err)
	}
	if req.Stage != want {
		t.Fatalf("advance %s: stage = %s, want %s", key, req.Stage, want)
	}
	return req
}

func registerAgent(t *testing.T, st *claim.Store, n int) core.Agent {
	t.Helper()
	a, err := st.RegisterAgent(sessionID(100 + n))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// A repeated submission of the same request is the same request; another
// payload under the same key is a conflict; neither touches any effect.
func TestIntake_DuplicateSubmit_Idempotent_ConflictOnOtherPayload(t *testing.T) {
	r := newRig(t, rigCfg)
	first, err := r.in.Accept("req-1", []byte("payload A"), 1)
	if err != nil {
		t.Fatal(err)
	}
	again, err := r.in.Accept("req-1", []byte("payload A"), 1)
	if err != nil || again.Seq != first.Seq {
		t.Fatalf("repeat = %+v, %v; want the same request (seq %d)", again, err, first.Seq)
	}
	if _, err := r.in.Accept("req-1", []byte("payload B"), 1); !errors.Is(err, core.ErrConflict) {
		t.Errorf("same key, other payload: %v, want ErrConflict", err)
	}
	all, _ := r.st.ListRequests()
	if len(all) != 1 {
		t.Errorf("requests stored = %d, want 1", len(all))
	}
	if p, f := r.pub.counts(); p != 0 || f != 0 {
		t.Errorf("effects touched by Accept: publishes=%d finds=%d", p, f)
	}
}

// The leader's ACK does not wait for the work: Accept returns at once even when
// publishing would hang, and nothing has been published.
func TestIntake_Accept_DoesNotWaitForPublishing(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.block = true
	done := make(chan error, 1)
	go func() { _, err := r.in.Accept("req-1", []byte("p"), 1); done <- err }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Accept did not return promptly")
	}
	if p, _ := r.pub.counts(); p != 0 {
		t.Errorf("publishes during Accept = %d, want 0", p)
	}
}

// The stages are separate and in order, each with evidence; the agent cannot
// skip one, and "started" is read from the claim, never stored.
func TestIntake_StagesAreSeparate_AndOrdered_WithEvidence(t *testing.T) {
	r := newRig(t, rigCfg)
	key := "req-1"
	if req, err := r.in.Accept(key, []byte("p"), 1); err != nil || req.Stage != claim.StageAccepted {
		t.Fatalf("accept: %+v, %v", req, err)
	}
	advance(t, r.in, key, claim.StageIssueConfirmed)
	advance(t, r.in, key, claim.StageBranchRegistered)
	if _, err := r.in.Advance(context.Background(), key); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("advance without an agent reservation: %v, want ErrTransitionInvalid", err)
	}
	agent := registerAgent(t, r.st, 1)
	if req, err := r.in.Reserve(key, agent); err != nil || req.Stage != claim.StageAgentReserved {
		t.Fatalf("reserve: %+v, %v", req, err)
	}
	advance(t, r.in, key, claim.StageWTPrepared)
	if _, err := r.in.Advance(context.Background(), key); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("advance before the agent acknowledged the receipt: %v", err)
	}
	acked, err := r.in.AckReady(key, agent)
	if err != nil || acked.Stage != claim.StageReadyAcked {
		t.Fatalf("ack: %+v, %v", acked, err)
	}
	view, err := r.in.View(key)
	if err != nil {
		t.Fatal(err)
	}
	if view.Started {
		t.Error("started is true before any attempt exists")
	}
	stages := map[claim.Stage]bool{}
	for _, ev := range view.Evidence {
		if ev.Ref == "" {
			t.Errorf("stage %s has empty evidence", ev.Stage)
		}
		stages[ev.Stage] = true
	}
	for _, want := range []claim.Stage{claim.StageAccepted, claim.StageIssueConfirmed, claim.StageBranchRegistered, claim.StageWTPrepared} {
		if !stages[want] {
			t.Errorf("no evidence recorded for %s", want)
		}
	}
}

// The agent writes only after the preparer has stopped: a receipt that says the
// preparer is still writing leaves the request at AGENT_RESERVED, blocked.
func TestIntake_PreparerStillWriting_AgentCannotAck(t *testing.T) {
	r := newRig(t, rigCfg)
	r.prep.writerStopped = false
	key := "req-1"
	_, _ = r.in.Accept(key, []byte("p"), 1)
	advance(t, r.in, key, claim.StageIssueConfirmed)
	advance(t, r.in, key, claim.StageBranchRegistered)
	agent := registerAgent(t, r.st, 1)
	if _, err := r.in.Reserve(key, agent); err != nil {
		t.Fatal(err)
	}
	req, err := r.in.Advance(context.Background(), key)
	if err == nil {
		t.Fatal("a receipt without writer_stopped was accepted")
	}
	if req.Stage != claim.StageAgentReserved || req.Blocked == nil || !req.Blocked.Retryable {
		t.Errorf("request = stage %s blocked %+v; want AGENT_RESERVED, blocked and retryable", req.Stage, req.Blocked)
	}
	if _, err := r.in.AckReady(key, agent); err == nil {
		t.Error("the agent acknowledged a worktree that was never prepared")
	}
}

// The first attempt is gated: not before READY_ACKED, not by another agent.
func TestIntake_ReadyGate_FirstAttemptNeedsTheAckingAgent(t *testing.T) {
	r := newRig(t, rigCfg)
	key := "req-1"
	_, _ = r.in.Accept(key, []byte("p"), 1)
	req := advance(t, r.in, key, claim.StageIssueConfirmed)
	issue := req.Issue
	advance(t, r.in, key, claim.StageBranchRegistered)
	owner, other := registerAgent(t, r.st, 1), registerAgent(t, r.st, 2)
	if _, err := r.in.Reserve(key, owner); err != nil {
		t.Fatal(err)
	}
	advance(t, r.in, key, claim.StageWTPrepared)
	if err := r.st.ReadyGate(issue, owner.SessionID); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("gate before READY_ACKED: %v, want ErrTransitionInvalid", err)
	}
	if _, err := r.in.AckReady(key, owner); err != nil {
		t.Fatal(err)
	}
	if err := r.st.ReadyGate(issue, owner.SessionID); err != nil {
		t.Errorf("gate for the acknowledging agent: %v", err)
	}
	if err := r.st.ReadyGate(issue, other.SessionID); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("gate for another agent: %v, want ErrTransitionInvalid", err)
	}
	if err := r.st.ReadyGate(issue+999, other.SessionID); err != nil {
		t.Errorf("an Issue no intake request names must not be gated: %v", err)
	}
}

// Publishing whose result is lost: the Issue exists but PublishIssue errored.
// The request becomes EFFECT_UNKNOWN, Advance only looks it up, and the lookup
// finds the Issue: exactly one Issue is ever created.
func TestIntake_LostPublishResponse_FoundByLookup_NeverPublishedTwice(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.publishErr, r.pub.createOnErr = errors.New("connection reset after the request was sent"), true
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	req, err := r.in.Advance(context.Background(), "req-1")
	if !errors.Is(err, core.ErrEffectUnknown) || req.Blocked == nil || !req.Blocked.Unknown {
		t.Fatalf("lost response: %+v, %v; want an unknown effect", req, err)
	}
	req, err = r.in.Advance(context.Background(), "req-1")
	if err != nil || req.Stage != claim.StageIssueConfirmed || req.Issue == 0 {
		t.Fatalf("lookup: %+v, %v; want ISSUE_CONFIRMED with the existing Issue", req, err)
	}
	if p, _ := r.pub.counts(); p != 1 {
		t.Errorf("PublishIssue calls = %d, want 1: an unknown effect must never be repeated", p)
	}
	if len(r.pub.created) != 1 {
		t.Errorf("Issues created = %d, want 1", len(r.pub.created))
	}
}

// Issue published but the lookup cannot confirm it (partial failure): still
// unknown, repeated Advance only looks, and a new attempt needs explicit evidence
// that nothing was created.
func TestIntake_UnconfirmedPublish_OnlyLooksUntilNotCreatedIsProven(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.findErr = errors.New("lookup unavailable")
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if _, err := r.in.Advance(context.Background(), "req-1"); !errors.Is(err, core.ErrEffectUnknown) {
		t.Fatalf("unconfirmed publish: %v, want ErrEffectUnknown", err)
	}
	for range 3 {
		_, _ = r.in.Advance(context.Background(), "req-1")
	}
	if p, _ := r.pub.counts(); p != 1 {
		t.Fatalf("PublishIssue calls after repeated Advance = %d, want 1", p)
	}
	if _, err := r.in.ConfirmNotCreated("req-1", ""); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("not-created without evidence: %v, want ErrEvidenceIncomplete", err)
	}
	r.pub.findErr = nil
	delete(r.pub.created, "req-1") // the proven case: nothing exists on the remote
	if _, err := r.in.ConfirmNotCreated("req-1", "lookup by marker returned nothing at t0"); err != nil {
		t.Fatal(err)
	}
	if req, err := r.in.Advance(context.Background(), "req-1"); err != nil || req.Stage != claim.StageIssueConfirmed {
		t.Fatalf("after proof: %+v, %v", req, err)
	}
	if p, _ := r.pub.counts(); p != 2 {
		t.Errorf("PublishIssue calls = %d, want 2 (one retry, only after proof)", p)
	}
}

// A refusal that proves nothing was sent is retryable, not unknown.
func TestIntake_NotSentFailure_IsRetryableNotUnknown(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.publishErr = fmt.Errorf("%w: refused before any call", intake.ErrNotSent)
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	req, err := r.in.Advance(context.Background(), "req-1")
	if err == nil || errors.Is(err, core.ErrEffectUnknown) || req.Blocked == nil || req.Blocked.Unknown || !req.Blocked.Retryable {
		t.Fatalf("not-sent: %+v, %v; want a retryable block", req.Blocked, err)
	}
	r.pub.publishErr = nil
	if req, err := r.in.Advance(context.Background(), "req-1"); err != nil || req.Stage != claim.StageIssueConfirmed {
		t.Errorf("retry after a not-sent refusal: %+v, %v", req, err)
	}
}

// A step that exceeds its timeout is recorded as such, and since the request may
// have gone out it is an unknown effect, not a retry.
func TestIntake_StepTimeout_IsRecordedAndTreatedAsUnknown(t *testing.T) {
	cfg := rigCfg
	cfg.StepTimeout = 40 * time.Millisecond
	r := newRig(t, cfg)
	r.pub.block = true
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	start := time.Now()
	req, err := r.in.Advance(context.Background(), "req-1")
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("a blocked publisher held Advance for %v", elapsed)
	}
	if !errors.Is(err, core.ErrEffectUnknown) || req.Blocked == nil || !req.Blocked.Unknown {
		t.Fatalf("timeout: %+v, %v; want an unknown effect", req.Blocked, err)
	}
	if !strings.Contains(req.Blocked.Detail, "timeout") {
		t.Errorf("detail %q does not say it was a timeout", req.Blocked.Detail)
	}
	r.pub.block = false
	_, _ = r.in.Advance(context.Background(), "req-1")
	if p, _ := r.pub.counts(); p != 1 {
		t.Errorf("PublishIssue calls = %d, want 1 (a timeout is never retried blindly)", p)
	}
}

// After a restart (a new Store and Intake over the same directory) the request
// continues from its recorded stage and does not publish again.
func TestIntake_Restart_ContinuesFromTheRecordedStage(t *testing.T) {
	r := newRig(t, rigCfg)
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	advance(t, r.in, "req-1", claim.StageIssueConfirmed)

	st2, err := claim.OpenWithLimit(r.dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	in2 := r.intakeOver(t, st2)
	view, err := in2.View("req-1")
	if err != nil || view.Stage != claim.StageIssueConfirmed || view.Issue == 0 {
		t.Fatalf("after restart: %+v, %v", view, err)
	}
	advance(t, in2, "req-1", claim.StageBranchRegistered)
	if p, _ := r.pub.counts(); p != 1 {
		t.Errorf("PublishIssue calls = %d, want 1", p)
	}
}

// Concurrency bound and back-pressure: requests beyond MaxInFlight park, beyond
// MaxParked are refused and record nothing.
func TestIntake_InFlightCap_ParksThenRefusesWithoutRecording(t *testing.T) {
	r := newRig(t, rigCfg) // MaxInFlight 2, MaxParked 1
	for i, wantParked := range []bool{false, false, true} {
		req, err := r.in.Accept(fmt.Sprintf("req-%d", i+1), []byte("p"), 1)
		if err != nil {
			t.Fatalf("accept %d: %v", i+1, err)
		}
		if (req.Parked != "") != wantParked {
			t.Errorf("request %d parked=%q, want parked=%v", i+1, req.Parked, wantParked)
		}
	}
	if _, err := r.in.Accept("req-4", []byte("p"), 1); !errors.Is(err, core.ErrIntakeFull) {
		t.Fatalf("beyond the parked bound: %v, want ErrIntakeFull", err)
	}
	all, _ := r.st.ListRequests()
	if len(all) != 3 {
		t.Errorf("requests stored = %d, want 3 (the refused one records nothing)", len(all))
	}
	if _, err := r.in.Advance(context.Background(), "req-3"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("a parked request advanced: %v", err)
	}
}

// Parked requests are released highest priority first, ties by arrival.
func TestIntake_PromotionOrder_HigherPriorityFirst_ThenArrival(t *testing.T) {
	cfg := intake.Config{MaxInFlight: 1, MaxParked: 4, Slots: []string{"slot-1"}, StepTimeout: time.Second}
	r := newRig(t, cfg)
	_, _ = r.in.Accept("req-0", []byte("p"), 1) // takes the only in-flight place
	for _, c := range []struct {
		key  string
		prio int
	}{{"low", 1}, {"high-first", 9}, {"mid", 5}, {"high-second", 9}} {
		if req, err := r.in.Accept(c.key, []byte("p"), c.prio); err != nil || req.Parked == "" {
			t.Fatalf("accept %s: %+v, %v; want it parked", c.key, req, err)
		}
	}
	// Make room: in-flight is full while req-0 exists, so lift the cap for the pump
	// the way a freed in-flight place would.
	promoted, err := r.st.PromoteParked(claim.Policy{MaxInFlight: 5, MaxParked: 4}, 4)
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, p := range promoted {
		order = append(order, p.Key)
	}
	want := []string{"high-first", "high-second", "mid", "low"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("promotion order = %v, want %v", order, want)
	}
}
