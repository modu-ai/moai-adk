package intake

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

var head = strings.Repeat("c", 40)

// fakePub is a code host that, like GitHub, makes no promise of idempotency:
// every PublishIssue that gets through creates a NEW issue. A test that sees
// created > 1 has found a duplicate.
type fakePub struct {
	mu       sync.Mutex
	base     int
	created  map[string]int // request key -> issue
	n        int            // issues created in total
	publish  int            // PublishIssue calls
	finds    int            // FindIssue calls
	err      error          // returned by PublishIssue
	lost     bool           // the issue IS created but the response is lost (err returned)
	block    bool           // PublishIssue waits for the context
	findErr  error
	wrongNum int  // when set, PublishIssue claims this number instead of the real one
	crash    bool // the issue IS created, then the process dies before anything is recorded

	noUser   bool        // the authenticated user cannot be established
	dups     []Candidate // issues that already do the same work
	dupErr   error
	dupLooks int
}

func (f *fakePub) AuthenticatedUser(context.Context) (int64, error) {
	if f.noUser {
		return 0, errors.New("whoami failed")
	}
	return 7, nil
}

func (f *fakePub) FindDuplicates(context.Context, claim.Request) ([]Candidate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dupLooks++
	return f.dups, f.dupErr
}

func (f *fakePub) PublishIssue(ctx context.Context, req claim.Request) (int, string, error) {
	f.mu.Lock()
	f.publish++
	block, err, lost := f.block, f.err, f.lost
	f.mu.Unlock()
	if block {
		<-ctx.Done()
		return 0, "", ctx.Err()
	}
	if err != nil && !lost {
		return 0, "", err
	}
	f.mu.Lock()
	if f.created == nil {
		f.created = map[string]int{}
	}
	f.n++
	issue := f.base + f.n - 1
	f.created[req.Key] = issue
	wrong := f.wrongNum
	crash := f.crash
	f.mu.Unlock()
	if crash {
		panic("process died after the issue was created")
	}
	if lost {
		return 0, "", err
	}
	if wrong != 0 {
		return wrong, "ref", nil
	}
	return issue, "issue-ref", nil
}

func (f *fakePub) FindIssue(_ context.Context, req claim.Request) (int, string, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.finds++
	if f.findErr != nil {
		return 0, "", false, f.findErr
	}
	n, ok := f.created[req.Key]
	return n, "issue-ref", ok, nil
}

func (f *fakePub) createdCount() int { f.mu.Lock(); defer f.mu.Unlock(); return f.n }
func (f *fakePub) publishCount() int { f.mu.Lock(); defer f.mu.Unlock(); return f.publish }

type fakeBr struct {
	mu         sync.Mutex
	registered map[string]claim.Branch
	calls      int
	err        error
	lost       bool // the branch IS registered but the response is lost
	crash      bool // the branch IS registered, then the process dies before anything is recorded
	oid        string
}

func (f *fakeBr) RegisterBranch(_ context.Context, req claim.Request) (claim.Branch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil && !f.lost {
		return claim.Branch{}, f.err
	}
	if f.registered == nil {
		f.registered = map[string]claim.Branch{}
	}
	oid := f.oid
	if oid == "" {
		oid = head
	}
	b := claim.Branch{Ref: "WT-" + req.Key, BaseOID: oid}
	f.registered[req.Key] = b
	if f.crash {
		panic("process died after the branch was registered")
	}
	if f.lost {
		return claim.Branch{}, f.err
	}
	return b, nil
}

func (f *fakeBr) FindBranch(_ context.Context, req claim.Request) (claim.Branch, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.registered[req.Key]
	return b, ok, nil
}

type fakePrep struct {
	calls   atomic.Int32
	stopped bool
	host    string
	err     error
	slots   []string
}

func (f *fakePrep) PrepareWorktree(_ context.Context, req claim.Request, slot string) (Ready, error) {
	f.calls.Add(1)
	f.slots = append(f.slots, slot)
	if f.err != nil {
		return Ready{}, f.err
	}
	return Ready{Host: f.host, Path: "/wt/" + slot, Head: head, WriterStopped: f.stopped}, nil
}

type rig struct {
	dir string
	in  *Intake
	pub *fakePub
	br  *fakeBr
	pr  *fakePrep
}

func newRig(t *testing.T, cfg Config) *rig {
	t.Helper()
	return rigAt(t, t.TempDir(), cfg, &fakePub{base: 1948}, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
}

func rigAt(t *testing.T, dir string, cfg Config, pub *fakePub, br *fakeBr, pr *fakePrep) *rig {
	t.Helper()
	st, err := claim.OpenWithLimit(dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.StepTimeout == 0 {
		cfg.StepTimeout = time.Second
	}
	if cfg.MaxInFlight == 0 {
		cfg.MaxInFlight = 4
	}
	if cfg.Slots == nil {
		cfg.Slots = []string{"wt-1", "wt-2"}
	}
	in, err := New(st, pub, br, pr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return &rig{dir: dir, in: in, pub: pub, br: br, pr: pr}
}

// ownIssues is a host where every Issue was opened by the acting user.
type ownIssues struct{}

func (ownIssues) IssueFacts(context.Context, core.IssueRef) (runner.IssueFacts, error) {
	return runner.IssueFacts{RequestedRepo: "acme/widgets", ActualRepo: "acme/widgets", UserID: 7, AuthorID: 7, Owner: core.OwnerNone}, nil
}

func ctx() context.Context { return context.Background() }

var agentSeq atomic.Int32

func newAgent(t *testing.T, r *rig) core.Agent {
	t.Helper()
	n := agentSeq.Add(1)
	a, err := r.in.Store.RegisterAgent(fmt.Sprintf("aaaaaaaa-0000-0000-0000-%012d", n))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func agentOf(req claim.Request) core.Agent {
	n, _ := strconv.Atoi(strings.TrimPrefix(req.AgentName, "agent-"))
	return core.Agent{SessionID: req.Agent, Number: n}
}

// advanceTo accepts a request and drives it to the stage, reserving a fresh
// agent where needed. It returns the request and the agent.
func advanceTo(t *testing.T, r *rig, key string, want claim.Stage) (claim.Request, core.Agent) {
	t.Helper()
	if _, err := r.in.Store.GetRequest(key); errors.Is(err, core.ErrNotHeld) {
		if _, err := r.in.Accept(key, []byte("payload-"+key), 1); err != nil {
			t.Fatal(err)
		}
	}
	var a core.Agent
	rank := map[claim.Stage]int{claim.StageAccepted: 0, claim.StageIssueConfirmed: 1, claim.StageBranchRegistered: 2, claim.StageAgentReserved: 3, claim.StageWTPrepared: 4, claim.StageReadyAcked: 5}
	cur, _ := r.in.Store.GetRequest(key)
	for rank[cur.Stage] < rank[want] {
		var err error
		switch cur.Stage {
		case claim.StageAccepted, claim.StageIssueConfirmed:
			cur, err = r.in.Advance(ctx(), key)
		case claim.StageBranchRegistered:
			a = newAgent(t, r)
			cur, err = r.in.Reserve(key, a)
		case claim.StageAgentReserved:
			cur, err = r.in.Advance(ctx(), key)
		case claim.StageWTPrepared:
			cur, err = r.in.AckReady(key, agentOf(cur))
		}
		if err != nil {
			t.Fatalf("advance %s from %s: %v", key, cur.Stage, err)
		}
	}
	if a.SessionID == "" && cur.Agent != "" {
		a = agentOf(cur)
	}
	return cur, a
}

// beginAttempt is what the reserved agent does next: claim the Issue and start.
func beginAttempt(t *testing.T, r *rig, req claim.Request, a core.Agent) error {
	t.Helper()
	rn := runner.New(r.in.Store)
	rn.Facts = ownIssues{}
	ref, _ := core.NewIssueRef(req.Issue)
	tk, err := rn.ClaimAs(ref, a, "disp-"+req.Key, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = rn.StartAttempt(tk, "attempt-1", core.DeliveryExecuted)
	return err
}

func TestConfigIsValidated(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 1)
	ok := Config{MaxInFlight: 1, MaxParked: 0, Slots: []string{"wt-1"}, StepTimeout: time.Second}
	good := func() (*claim.Store, Publisher, Brancher, Preparer) {
		return st, &fakePub{}, &fakeBr{}, &fakePrep{}
	}
	if s, p, b, w := good(); true {
		if _, err := New(s, p, b, w, ok); err != nil {
			t.Fatalf("valid config: %v", err)
		}
	}
	for name, mut := range map[string]func(*Config){
		"in-flight 0":   func(c *Config) { c.MaxInFlight = 0 },
		"negative park": func(c *Config) { c.MaxParked = -1 },
		"no slots":      func(c *Config) { c.Slots = nil },
		"bad slot":      func(c *Config) { c.Slots = []string{"Bad Slot"} },
		"no timeout":    func(c *Config) { c.StepTimeout = 0 },
	} {
		c := ok
		mut(&c)
		s, p, b, w := good()
		if _, err := New(s, p, b, w, c); !errors.Is(err, core.ErrLimitInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	s, _, b, w := good()
	if _, err := New(s, nil, b, w, ok); err == nil {
		t.Error("nil publisher accepted")
	}
	s, p, _, w := good()
	if _, err := New(s, p, nil, w, ok); err == nil {
		t.Error("nil brancher accepted")
	}
}

func TestAcceptIsDurableIdempotentAndConflicting(t *testing.T) {
	r := newRig(t, Config{})
	a, err := r.in.Accept("req-1", []byte(`{"task":"x"}`), 3)
	if err != nil || a.Stage != claim.StageAccepted || a.Evidence[0].Stage != claim.StageAccepted {
		t.Fatalf("accept: %+v %v", a, err)
	}
	b, err := r.in.Accept("req-1", []byte(`{"task":"x"}`), 3)
	if err != nil || b.Seq != a.Seq {
		t.Errorf("same payload: %+v %v", b, err)
	}
	if _, err := r.in.Accept("req-1", []byte(`{"task":"y"}`), 3); !errors.Is(err, core.ErrConflict) {
		t.Errorf("other payload: %v", err)
	}
	if r.pub.publishCount() != 0 {
		t.Error("accepting must not create anything")
	}
}

func TestPayloadIsStoredAsHashOnly(t *testing.T) {
	r := newRig(t, Config{})
	if _, err := r.in.Accept("req-1", []byte("token=ghp_supersecretvalue1234567890"), 1); err != nil {
		t.Fatal(err)
	}
	var all string
	_ = filepath.Walk(r.dir, func(p string, i os.FileInfo, err error) error {
		if err == nil && !i.IsDir() {
			b, _ := os.ReadFile(p)
			all += string(b)
		}
		return nil
	})
	if strings.Contains(all, "supersecret") || strings.Contains(all, "ghp_") {
		t.Errorf("payload text reached the store")
	}
}

func TestFullPathThroughAllSixStagesWithEvidence(t *testing.T) {
	r := newRig(t, Config{})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)

	req, err := r.in.Advance(ctx(), "req-1")
	if err != nil || req.Stage != claim.StageIssueConfirmed || req.Issue != 1948 {
		t.Fatalf("issue: %+v %v", req, err)
	}
	req, err = r.in.Advance(ctx(), "req-1")
	if err != nil || req.Stage != claim.StageBranchRegistered || req.Branch == nil || req.Branch.Ref != "WT-req-1" || req.Branch.BaseOID != head {
		t.Fatalf("branch: %+v %v", req, err)
	}
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("advance must wait for an agent reservation: %v", err)
	}
	a := newAgent(t, r)
	req, err = r.in.Reserve("req-1", a)
	if err != nil || req.Stage != claim.StageAgentReserved || req.Agent != a.SessionID || req.Slot == "" {
		t.Fatalf("reserve: %+v %v", req, err)
	}
	if r.pr.calls.Load() != 0 {
		t.Error("reserving an agent must not prepare anything")
	}
	req, err = r.in.Advance(ctx(), "req-1")
	if err != nil || req.Stage != claim.StageWTPrepared || req.Ready == nil || req.Ready.Host != "host-1" || !req.Ready.WriterStopped {
		t.Fatalf("prepare: %+v %v", req, err)
	}
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("advance must wait for the agent's ack: %v", err)
	}
	req, err = r.in.AckReady("req-1", a)
	if err != nil || req.Stage != claim.StageReadyAcked || req.AckedBy != a.SessionID {
		t.Fatalf("ack: %+v %v", req, err)
	}
	if v, _ := r.in.View("req-1"); v.Started {
		t.Error("not started before the first attempt")
	}
	if err := beginAttempt(t, r, req, a); err != nil {
		t.Fatalf("attempt after the ack: %v", err)
	}
	v, _ := r.in.View("req-1")
	if !v.Started || v.StartedRef == "" || v.Stage != claim.StageReadyAcked {
		t.Fatalf("view after the attempt: %+v", v)
	}
	var stages []claim.Stage
	for _, e := range v.Evidence {
		stages = append(stages, e.Stage)
	}
	want := []claim.Stage{claim.StageAccepted, claim.StageIssueConfirmed, claim.StageBranchRegistered, claim.StageAgentReserved, claim.StageWTPrepared, claim.StageReadyAcked}
	if fmt.Sprint(stages) != fmt.Sprint(want) {
		t.Errorf("evidence stages = %v, want %v", stages, want)
	}
	if r.pub.createdCount() != 1 || r.br.calls != 1 || r.pr.calls.Load() != 1 {
		t.Errorf("effects: issues=%d branches=%d prepares=%d", r.pub.createdCount(), r.br.calls, r.pr.calls.Load())
	}
}

// BUILDING (the agent's first attempt) cannot begin before the whole chain, and
// not for an agent other than the one that acked.
func TestAttemptRefusedUntilReadyAckedByTheSameAgent(t *testing.T) {
	for _, stage := range []claim.Stage{claim.StageIssueConfirmed, claim.StageBranchRegistered, claim.StageAgentReserved, claim.StageWTPrepared} {
		r := newRig(t, Config{})
		req, a := advanceTo(t, r, "req-1", stage)
		if a.SessionID == "" {
			a = newAgent(t, r)
		}
		if err := beginAttempt(t, r, req, a); !errors.Is(err, core.ErrTransitionInvalid) {
			t.Errorf("attempt at %s: %v", stage, err)
		}
		if v, _ := r.in.View("req-1"); v.Started {
			t.Errorf("a refused attempt must not read as started (%s)", stage)
		}
	}
	r := newRig(t, Config{})
	req, _ := advanceTo(t, r, "req-1", claim.StageReadyAcked)
	other := newAgent(t, r)
	if err := beginAttempt(t, r, req, other); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("an agent that did not ack: %v", err)
	}
}

func TestOnlyTheReservedAgentCanAck(t *testing.T) {
	r := newRig(t, Config{})
	_, _ = advanceTo(t, r, "req-1", claim.StageWTPrepared)
	if _, err := r.in.AckReady("req-1", newAgent(t, r)); !errors.Is(err, core.ErrForeign) {
		t.Errorf("ack by another agent: %v", err)
	}
}

func TestPreparerThatDidNotStopWritingIsNotPrepared(t *testing.T) {
	r := newRig(t, Config{})
	r.pr.stopped = false
	_, _ = advanceTo(t, r, "req-1", claim.StageAgentReserved)
	req, err := r.in.Advance(ctx(), "req-1")
	if err == nil {
		t.Fatalf("not-stopped preparer accepted: %+v", req)
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	if cur.Stage != claim.StageAgentReserved || cur.Ready != nil || cur.Attempts != 1 || cur.Blocked == nil || !cur.Blocked.Retryable || cur.Blocked.Unknown {
		t.Errorf("record after a refused receipt: %+v blocked=%+v", cur, cur.Blocked)
	}
	r.pr.stopped = true
	cur, err = r.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageWTPrepared || cur.Blocked != nil {
		t.Errorf("retry: %+v %v", cur, err)
	}
	if len(r.pr.slots) != 2 || r.pr.slots[0] != r.pr.slots[1] {
		t.Errorf("a retry must reuse the reserved slot: %v", r.pr.slots)
	}
}

func TestPreparerMustNameTheHost(t *testing.T) {
	r := newRig(t, Config{})
	r.pr.host = ""
	_, _ = advanceTo(t, r, "req-1", claim.StageAgentReserved)
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("a receipt without a host: %v", err)
	}
}

// Losing the response of an Issue creation must never lead to a second Issue.
func TestLostCreationResponseIsLookedUpNotRepeated(t *testing.T) {
	pub := &fakePub{base: 1948, lost: true, err: errors.New("connection reset after the request was sent")}
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, err := r.in.Advance(ctx(), "req-1")
	if !errors.Is(err, core.ErrEffectUnknown) {
		t.Fatalf("want an unknown effect, got %v", err)
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	if cur.Stage != claim.StageAccepted || cur.Blocked == nil || !cur.Blocked.Unknown || cur.Blocked.Step != claim.StageAccepted {
		t.Fatalf("record: %+v blocked=%+v", cur, cur.Blocked)
	}
	cur, err = r.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageIssueConfirmed || cur.Issue != 1948 || cur.Blocked != nil {
		t.Fatalf("the lookup must find and bind the issue: %+v %v", cur, err)
	}
	if pub.publishCount() != 1 || pub.createdCount() != 1 {
		t.Errorf("publish calls=%d issues created=%d, want 1 and 1", pub.publishCount(), pub.createdCount())
	}
}

func TestNothingFoundStaysUnknownUntilNotCreatedIsConfirmed(t *testing.T) {
	pub := &fakePub{base: 1948, err: errors.New("HTTP 429 secondary rate limit")} // not created, but not proven
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrEffectUnknown) {
		t.Fatalf("first attempt: %v", err)
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	if cur.Attempts != 1 || cur.Blocked == nil || !strings.Contains(cur.Blocked.Detail, "429") {
		t.Errorf("the failure must be recorded: %+v", cur.Blocked)
	}
	// A lookup that finds nothing is not proof; the request stays unknown and nothing is sent.
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrEffectUnknown) {
		t.Errorf("second advance: %v", err)
	}
	if pub.publishCount() != 1 {
		t.Fatalf("the creation was repeated on a guess: %d calls", pub.publishCount())
	}
	if _, err := r.in.ConfirmNotCreated("req-1", ""); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("confirming without evidence: %v", err)
	}
	if _, err := r.in.ConfirmNotCreated("req-1", "gh issue list --search marker returned 0 rows at the time of the check"); err != nil {
		t.Fatal(err)
	}
	cur, _ = r.in.Store.GetRequest("req-1")
	if cur.Blocked.Unknown || !cur.Blocked.Retryable {
		t.Fatalf("after the confirmation a retry is allowed: %+v", cur.Blocked)
	}
	pub.mu.Lock()
	pub.err = nil
	pub.mu.Unlock()
	cur, err := r.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageIssueConfirmed || pub.createdCount() != 1 || pub.publishCount() != 2 {
		t.Errorf("retry: %+v %v issues=%d calls=%d", cur, err, pub.createdCount(), pub.publishCount())
	}
	if _, err := r.in.ConfirmNotCreated("req-1", "late"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("nothing is unknown any more: %v", err)
	}
}

func TestAnErrorThatProvesNothingWasSentMayBeRetried(t *testing.T) {
	pub := &fakePub{base: 1948, err: fmt.Errorf("%w: refused before any call", ErrNotSent)}
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, err := r.in.Advance(ctx(), "req-1")
	if err == nil || errors.Is(err, core.ErrEffectUnknown) {
		t.Fatalf("a proven non-send is a plain failure: %v", err)
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	if cur.Blocked == nil || !cur.Blocked.Retryable || cur.Blocked.Unknown {
		t.Fatalf("blocked: %+v", cur.Blocked)
	}
	pub.mu.Lock()
	pub.err = nil
	pub.mu.Unlock()
	if cur, err = r.in.Advance(ctx(), "req-1"); err != nil || cur.Stage != claim.StageIssueConfirmed {
		t.Errorf("retry: %+v %v", cur, err)
	}
	if pub.createdCount() != 1 {
		t.Errorf("issues created = %d", pub.createdCount())
	}
}

func TestACreationMustBeConfirmedByALookup(t *testing.T) {
	pub := &fakePub{base: 1948, wrongNum: 77} // the publisher names an issue the lookup does not find
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrEffectUnknown) {
		t.Fatalf("disagreeing lookup: %v", err)
	}
	if cur, _ := r.in.Store.GetRequest("req-1"); cur.Stage != claim.StageAccepted || cur.Issue != 0 {
		t.Errorf("an unconfirmed issue was bound: %+v", cur)
	}
}

func TestTimeoutDuringCreationIsAnUnknownEffect(t *testing.T) {
	pub := &fakePub{base: 1948, block: true}
	r := rigAt(t, t.TempDir(), Config{StepTimeout: 20 * time.Millisecond}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, err := r.in.Advance(ctx(), "req-1")
	if !errors.Is(err, core.ErrEffectUnknown) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want an unknown effect caused by the deadline, got %v", err)
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	if cur.Blocked == nil || !cur.Blocked.Unknown || !strings.Contains(cur.Blocked.Detail, "timeout") || cur.Attempts != 1 {
		t.Errorf("timeout record: %+v", cur.Blocked)
	}
}

func TestBranchReceiptNeedsAFullBaseOIDAndRef(t *testing.T) {
	br := &fakeBr{oid: "cb30e05b3e"}
	r := rigAt(t, t.TempDir(), Config{}, &fakePub{base: 1948}, br, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, _ = r.in.Advance(ctx(), "req-1") // ISSUE_CONFIRMED
	_, err := r.in.Advance(ctx(), "req-1")
	if !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("short base OID: %v", err)
	}
	if cur, _ := r.in.Store.GetRequest("req-1"); cur.Stage != claim.StageIssueConfirmed || cur.Branch != nil {
		t.Errorf("a malformed receipt advanced the request: %+v", cur)
	}
}

func TestLostBranchRegistrationIsLookedUpNotRepeated(t *testing.T) {
	br := &fakeBr{lost: true, err: errors.New("push timed out after the ref was created")}
	r := rigAt(t, t.TempDir(), Config{}, &fakePub{base: 1948}, br, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, _ = r.in.Advance(ctx(), "req-1")
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrEffectUnknown) {
		t.Fatalf("lost registration: %v", err)
	}
	cur, err := r.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageBranchRegistered || cur.Branch == nil {
		t.Fatalf("lookup: %+v %v", cur, err)
	}
	if br.calls != 1 {
		t.Errorf("the branch was registered %d times", br.calls)
	}
}

func TestBackpressureParksThenRejectsAndPromotesByPriority(t *testing.T) {
	r := newRig(t, Config{MaxInFlight: 2, MaxParked: 2, Slots: []string{"wt-1", "wt-2", "wt-3"}})
	_, _ = r.in.Accept("a", []byte("1"), 1)
	_, _ = r.in.Accept("b", []byte("2"), 1)
	lo, err := r.in.Accept("lo", []byte("3"), 1)
	if err != nil || lo.Parked != claim.ParkBackpressure {
		t.Fatalf("lo: %+v %v", lo, err)
	}
	hi, _ := r.in.Accept("hi", []byte("4"), 9)
	if hi.Parked != claim.ParkBackpressure {
		t.Fatalf("hi: %+v", hi)
	}
	if _, err := r.in.Accept("over", []byte("5"), 99); !errors.Is(err, core.ErrIntakeFull) {
		t.Fatalf("over: %v", err)
	}
	if _, err := r.in.Advance(ctx(), "hi"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("a parked request must not advance: %v", err)
	}
	// Drive a all the way to its first attempt; that frees one in-flight place.
	req, ag := advanceTo(t, r, "a", claim.StageReadyAcked)
	if err := beginAttempt(t, r, req, ag); err != nil {
		t.Fatal(err)
	}
	got, err := r.in.Pump()
	if err != nil || len(got) != 1 || got[0].Key != "hi" {
		t.Fatalf("pump: %+v %v", got, err)
	}
	if cur, _ := r.in.Store.GetRequest("lo"); cur.Parked != claim.ParkBackpressure {
		t.Errorf("lo must still be parked: %+v", cur)
	}
}

func TestConcurrentAcceptNeverExceedsTheLimits(t *testing.T) {
	r := newRig(t, Config{MaxInFlight: 3, MaxParked: 5})
	var wg sync.WaitGroup
	var admitted, parked, full atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req, err := r.in.Accept("req-"+string(rune('a'+i)), []byte("p"+string(rune('a'+i))), i)
			switch {
			case errors.Is(err, core.ErrIntakeFull):
				full.Add(1)
			case err != nil:
				t.Errorf("accept %d: %v", i, err)
			case req.Parked != "":
				parked.Add(1)
			default:
				admitted.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if admitted.Load() != 3 || parked.Load() != 5 || full.Load() != 12 {
		t.Errorf("admitted=%d parked=%d rejected=%d", admitted.Load(), parked.Load(), full.Load())
	}
	recs, _ := r.in.Store.ListRequests()
	if len(recs) != 8 {
		t.Errorf("records = %d, rejected requests must leave none", len(recs))
	}
}

func TestSameKeyConcurrentlyYieldsOneRecordAndOneConflict(t *testing.T) {
	r := newRig(t, Config{})
	var wg sync.WaitGroup
	var ok, conflict atomic.Int32
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := r.in.Accept("same", []byte("payload-"+string(rune('a'+i%2))), 1)
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, core.ErrConflict):
				conflict.Add(1)
			default:
				t.Errorf("accept: %v", err)
			}
		}(i)
	}
	wg.Wait()
	recs, _ := r.in.Store.ListRequests()
	if len(recs) != 1 || ok.Load()+conflict.Load() != 8 || conflict.Load() == 0 {
		t.Errorf("records=%d ok=%d conflict=%d", len(recs), ok.Load(), conflict.Load())
	}
}

func TestReservationNeedsAFreeExclusiveSlot(t *testing.T) {
	r := newRig(t, Config{Slots: []string{"wt-1"}})
	_, _ = advanceTo(t, r, "a", claim.StageAgentReserved)
	_, _ = advanceTo(t, r, "b", claim.StageBranchRegistered)
	before := r.pr.calls.Load()
	_, err := r.in.Reserve("b", newAgent(t, r))
	if !errors.Is(err, core.ErrNoSlot) {
		t.Fatalf("b without a free slot: %v", err)
	}
	if r.pr.calls.Load() != before {
		t.Error("a worktree was prepared without a slot")
	}
	cur, _ := r.in.Store.GetRequest("b")
	if cur.Parked != claim.ParkNoSlot || cur.Slot != "" || cur.Stage != claim.StageBranchRegistered || cur.Agent != "" {
		t.Errorf("b record: %+v", cur)
	}
	if owner, _ := r.in.Store.SlotOwner("wt-1"); owner != "a" {
		t.Errorf("slot owner = %q", owner)
	}
	// a finishes preparing, its agent acks and begins; the slot returns; b is promoted.
	req, ag := advanceTo(t, r, "a", claim.StageReadyAcked)
	if err := beginAttempt(t, r, req, ag); err != nil {
		t.Fatal(err)
	}
	if err := r.in.ReleaseWorktree("a"); err != nil {
		t.Fatal(err)
	}
	if got, err := r.in.Pump(); err != nil || len(got) != 1 || got[0].Key != "b" {
		t.Fatalf("pump: %+v %v", got, err)
	}
	cur, err = r.in.Reserve("b", newAgent(t, r))
	if err != nil || cur.Stage != claim.StageAgentReserved || cur.Slot != "wt-1" {
		t.Errorf("b after promotion: %+v %v", cur, err)
	}
}

func TestReleaseWorktreeOnlyAfterTheAgentBegan(t *testing.T) {
	r := newRig(t, Config{})
	_, _ = advanceTo(t, r, "a", claim.StageReadyAcked)
	if err := r.in.ReleaseWorktree("a"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("release before the agent began: %v", err)
	}
}

func TestRestartResumesWithoutRepeatingEffects(t *testing.T) {
	dir := t.TempDir()
	pub := &fakePub{base: 1948}
	r1 := rigAt(t, dir, Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r1.in.Accept("req-1", []byte("p"), 1)
	if _, err := r1.in.Advance(ctx(), "req-1"); err != nil {
		t.Fatal(err)
	}
	// The process restarts: new store, new intake, same directory, same host state.
	r2 := rigAt(t, dir, Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	cur, err := r2.in.Store.GetRequest("req-1")
	if err != nil || cur.Stage != claim.StageIssueConfirmed || cur.Issue != 1948 {
		t.Fatalf("durable record: %+v %v", cur, err)
	}
	cur, err = r2.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageBranchRegistered {
		t.Fatalf("resume: %+v %v", cur, err)
	}
	if pub.publishCount() != 1 || pub.createdCount() != 1 {
		t.Errorf("the issue was created again after restart (calls=%d issues=%d)", pub.publishCount(), pub.createdCount())
	}
	if _, err := r2.in.Accept("req-1", []byte("p"), 1); err != nil {
		t.Errorf("replayed request after restart: %v", err)
	}
	if pub.publishCount() != 1 {
		t.Error("a replayed request caused a new effect")
	}
}

// ErrBusy is returned to the caller, never retried inside the product path: the
// store waits a bounded time for the lock and then says "retry".
func TestAcceptReturnsBusyToTheCallerAfterABoundedWait(t *testing.T) {
	r := newRig(t, Config{})
	lock := filepath.Join(r.dir, "requests.lock")
	if err := os.WriteFile(lock, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err := r.in.Accept("req-1", []byte("p"), 1)
	waited := time.Since(start)
	if !errors.Is(err, core.ErrBusy) {
		t.Fatalf("want ErrBusy, got %v", err)
	}
	// The wait is the store's budget: a sleep never returns early, so it lasts at
	// least that long, and it must not run on without end (five times the budget
	// is the slack this test has always allowed for a loaded machine).
	budget := claim.LockWaitBudget()
	if waited < budget {
		t.Errorf("the wait for a held lock ended before its budget: %s < %s", waited, budget)
	}
	if waited > 5*budget {
		t.Errorf("the wait for a held lock is not bounded: %s (budget %s)", waited, budget)
	}
	if _, gerr := r.in.Store.GetRequest("req-1"); !errors.Is(gerr, core.ErrNotHeld) {
		t.Error("a Busy accept left a record")
	}
	if r.pub.publishCount() != 0 {
		t.Error("a Busy accept caused an effect")
	}
	if err := os.Remove(lock); err != nil {
		t.Fatal(err)
	}
	if _, err := r.in.Accept("req-1", []byte("p"), 1); err != nil {
		t.Errorf("accept after the lock cleared: %v", err)
	}
}

// The publish intent is recorded before the send (t2072), so a store that is Busy
// stops the step before anything is sent: the Busy is reported and no effect
// happened. (Before t2072 the send went out and the failure then could not be
// recorded; the join of both errors is still what fail() does when the store goes
// Busy after the intent was written.)
func TestAdvanceFailureWhileBusyKeepsBothErrors(t *testing.T) {
	pub := &fakePub{base: 1948, err: fmt.Errorf("%w: HTTP 429 secondary rate limit", ErrNotSent)}
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if err := os.WriteFile(filepath.Join(r.dir, "requests.lock"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := r.in.Advance(ctx(), "req-1")
	if err == nil || !errors.Is(err, core.ErrBusy) {
		t.Fatalf("the Busy must be reported: %v", err)
	}
	if pub.publishCount() != 0 {
		t.Error("an intent that could not be recorded was sent anyway")
	}
}

// Same work already exists as another author's Issue: nothing is created in its
// place. The request is blocked and names the Issue, so the work can go there.
func TestDuplicateOfAnotherAuthorsIssueCreatesNothing(t *testing.T) {
	pub := &fakePub{base: 1948, dups: []Candidate{{Issue: 3000, AuthorID: 99, Ref: "issue-3000"}}}
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, err := r.in.Advance(ctx(), "req-1")
	if !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Fatalf("want a refusal, got %v", err)
	}
	if pub.publishCount() != 0 || pub.createdCount() != 0 {
		t.Fatalf("an issue was created over someone else's: calls=%d created=%d", pub.publishCount(), pub.createdCount())
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	b := cur.Blocked
	if cur.Stage != claim.StageAccepted || b == nil || b.Candidate != 3000 || b.Reason != core.ReasonNotIssueAuthor || b.Retryable || b.Unknown || cur.Issue != 0 {
		t.Fatalf("record: %+v blocked=%+v", cur, b)
	}
	// Asking again changes nothing: still blocked, still nothing created.
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Errorf("second advance: %v", err)
	}
	if pub.publishCount() != 0 {
		t.Error("the second advance created an issue")
	}
}

// The user's own earlier Issue for the same work is reused, not duplicated.
func TestOwnDuplicateIsReusedNotRecreated(t *testing.T) {
	pub := &fakePub{base: 1948, dups: []Candidate{{Issue: 2500, AuthorID: 7, Ref: "issue-2500"}}}
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	cur, err := r.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageIssueConfirmed || cur.Issue != 2500 || cur.Blocked != nil {
		t.Fatalf("reuse: %+v %v", cur, err)
	}
	if pub.publishCount() != 0 {
		t.Error("an own duplicate must not lead to another issue")
	}
	if last := cur.Evidence[len(cur.Evidence)-1]; last.Stage != claim.StageIssueConfirmed || !strings.Contains(last.Ref, "2500") {
		t.Errorf("evidence: %+v", last)
	}
}

func TestOwnDuplicateWinsOverAForeignOne(t *testing.T) {
	pub := &fakePub{base: 1948, dups: []Candidate{{Issue: 3000, AuthorID: 99}, {Issue: 2500, AuthorID: 7}}}
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if cur, err := r.in.Advance(ctx(), "req-1"); err != nil || cur.Issue != 2500 {
		t.Errorf("got %+v %v", cur, err)
	}
}

// If it cannot be told whose Issue it is, or who the user is, nothing is created.
func TestUnknownIdentitiesBlockCreation(t *testing.T) {
	cases := map[string]*fakePub{
		"user unknown":     {base: 1948, noUser: true},
		"candidate author": {base: 1948, dups: []Candidate{{Issue: 3000, AuthorID: 0}}},
		"duplicate lookup": {base: 1948, dupErr: errors.New("search failed")},
	}
	for name, pub := range cases {
		t.Run(name, func(t *testing.T) {
			r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
			_, _ = r.in.Accept("req-1", []byte("p"), 1)
			if _, err := r.in.Advance(ctx(), "req-1"); err == nil {
				t.Fatal("want a refusal")
			}
			if pub.publishCount() != 0 {
				t.Fatalf("an issue was created without knowing whose work exists")
			}
			cur, _ := r.in.Store.GetRequest("req-1")
			if cur.Blocked == nil || !cur.Blocked.Retryable || cur.Blocked.Unknown || cur.Stage != claim.StageAccepted {
				t.Errorf("a read-only failure is retryable and not an unknown effect: %+v", cur.Blocked)
			}
		})
	}
}

func TestNoDuplicateMeansTheIssueIsCreatedAfterTheLookup(t *testing.T) {
	pub := &fakePub{base: 1948}
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakeBr{}, &fakePrep{stopped: true, host: "host-1"})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	cur, err := r.in.Advance(ctx(), "req-1")
	if err != nil || cur.Issue != 1948 || pub.dupLooks != 1 || pub.createdCount() != 1 {
		t.Errorf("got %+v %v lookups=%d created=%d", cur, err, pub.dupLooks, pub.createdCount())
	}
}
