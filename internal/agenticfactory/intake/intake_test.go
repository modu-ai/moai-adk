package intake

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

type fakePub struct {
	mu    sync.Mutex
	calls []string
	issue int
	err   error
	block bool
}

func (f *fakePub) PublishIssue(ctx context.Context, req claim.Request) (int, string, error) {
	f.mu.Lock()
	f.calls = append(f.calls, req.Key)
	err, block, n := f.err, f.block, f.issue+len(f.calls)-1
	f.mu.Unlock()
	if block {
		<-ctx.Done()
		return 0, "", ctx.Err()
	}
	if err != nil {
		return 0, "", err
	}
	return n, "issue-ref", nil
}

func (f *fakePub) count() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.calls) }

type fakePrep struct {
	calls   atomic.Int32
	stopped bool
	err     error
	slots   []string
}

func (f *fakePrep) PrepareWorktree(_ context.Context, req claim.Request, slot string) (Ready, error) {
	f.calls.Add(1)
	f.slots = append(f.slots, slot)
	if f.err != nil {
		return Ready{}, f.err
	}
	return Ready{Path: "/wt/" + slot, Head: head, WriterStopped: f.stopped}, nil
}

type rig struct {
	dir string
	in  *Intake
	pub *fakePub
	pr  *fakePrep
}

func newRig(t *testing.T, cfg Config) *rig {
	t.Helper()
	dir := t.TempDir()
	return rigAt(t, dir, cfg, &fakePub{issue: 1948}, &fakePrep{stopped: true})
}

func rigAt(t *testing.T, dir string, cfg Config, pub *fakePub, pr *fakePrep) *rig {
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
	in, err := New(st, pub, pr, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return &rig{dir: dir, in: in, pub: pub, pr: pr}
}

// startAgent claims the request's Issue for a fresh agent and begins its
// attempt: the moment the request becomes STARTED.
func startAgent(t *testing.T, r *rig, key string, n int) (core.Agent, error) {
	t.Helper()
	req, err := r.in.Store.GetRequest(key)
	if err != nil {
		t.Fatal(err)
	}
	ag, err := r.in.Store.RegisterAgent("aaaaaaaa-0000-0000-0000-" + strings.Repeat("0", 11) + string(rune('0'+n)))
	if err != nil {
		t.Fatal(err)
	}
	rn := runner.New(r.in.Store)
	ref, _ := core.NewIssueRef(req.Issue)
	tk, err := rn.ClaimAs(ref, ag, "disp-"+key, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = rn.StartAttempt(tk, "attempt-1", core.DeliveryExecuted)
	return ag, err
}

// acceptRetry is what a caller does with ErrBusy: the store said "retry".
func acceptRetry(in *Intake, key string, payload []byte, pri int) (claim.Request, error) {
	for i := 0; i < 100; i++ {
		req, err := in.Accept(key, payload, pri)
		if !errors.Is(err, core.ErrBusy) {
			return req, err
		}
		time.Sleep(5 * time.Millisecond)
	}
	return in.Accept(key, payload, pri)
}

func ctx() context.Context { return context.Background() }

func TestConfigIsValidated(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 1)
	ok := Config{MaxInFlight: 1, MaxParked: 0, Slots: []string{"wt-1"}, StepTimeout: time.Second}
	if _, err := New(st, &fakePub{}, &fakePrep{}, ok); err != nil {
		t.Fatalf("valid config: %v", err)
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
		if _, err := New(st, &fakePub{}, &fakePrep{}, c); !errors.Is(err, core.ErrLimitInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := New(st, nil, &fakePrep{}, ok); err == nil {
		t.Error("nil publisher accepted")
	}
}

func TestAcceptIsDurableIdempotentAndConflicting(t *testing.T) {
	r := newRig(t, Config{})
	a, err := r.in.Accept("req-1", []byte(`{"task":"x"}`), 3)
	if err != nil || a.Stage != claim.StageACK || a.Evidence[0].Stage != claim.StageACK {
		t.Fatalf("accept: %+v %v", a, err)
	}
	b, err := r.in.Accept("req-1", []byte(`{"task":"x"}`), 3)
	if err != nil || b.Seq != a.Seq {
		t.Errorf("same payload: %+v %v", b, err)
	}
	if _, err := r.in.Accept("req-1", []byte(`{"task":"y"}`), 3); !errors.Is(err, core.ErrConflict) {
		t.Errorf("other payload: %v", err)
	}
	if r.pub.count() != 0 {
		t.Error("accepting must not publish anything")
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

func TestFullPathWithEvidencePerStage(t *testing.T) {
	r := newRig(t, Config{})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	req, err := r.in.Advance(ctx(), "req-1")
	if err != nil || req.Stage != claim.StageIssuePublished || req.Issue != 1948 {
		t.Fatalf("publish step: %+v %v", req, err)
	}
	req, err = r.in.Advance(ctx(), "req-1")
	if err != nil || req.Stage != claim.StageWTReady || req.Slot == "" || req.Ready == nil || !req.Ready.WriterStopped {
		t.Fatalf("worktree step: %+v %v", req, err)
	}
	if owner, _ := r.in.Store.SlotOwner(req.Slot); owner != "req-1" {
		t.Errorf("slot owner = %q", owner)
	}
	if v, _ := r.in.View("req-1"); v.Stage != claim.StageWTReady {
		t.Errorf("not started yet, view says %v", v.Stage)
	}
	if _, err := startAgent(t, r, "req-1", 1); err != nil {
		t.Fatal(err)
	}
	v, err := r.in.View("req-1")
	if err != nil || v.Stage != claim.StageStarted {
		t.Fatalf("view after attempt: %+v %v", v, err)
	}
	var stages []claim.Stage
	for _, e := range v.Evidence {
		stages = append(stages, e.Stage)
	}
	want := []claim.Stage{claim.StageACK, claim.StageIssuePublished, claim.StageWTReady, claim.StageStarted}
	if len(stages) != 4 || stages[0] != want[0] || stages[1] != want[1] || stages[2] != want[2] || stages[3] != want[3] {
		t.Errorf("evidence stages = %v", stages)
	}
	if r.pub.count() != 1 || r.pr.calls.Load() != 1 {
		t.Errorf("effects: publish=%d prepare=%d", r.pub.count(), r.pr.calls.Load())
	}
}

// The agent may not begin writing before the preparer's READY receipt.
func TestAttemptRefusedUntilWorktreeReady(t *testing.T) {
	r := newRig(t, Config{})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, _ = r.in.Advance(ctx(), "req-1") // ISSUE_PUBLISHED
	if _, err := startAgent(t, r, "req-1", 1); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Fatalf("attempt before WT_READY: %v", err)
	}
	if v, _ := r.in.View("req-1"); v.Stage == claim.StageStarted {
		t.Error("a refused attempt must not read as started")
	}
	req, _ := r.in.Store.GetRequest("req-1")
	ref, _ := core.NewIssueRef(req.Issue)
	if rec, err := r.in.Store.Load(ref); err != nil || rec.State != core.StateClaimed {
		t.Errorf("claim must stay claimed: %+v %v", rec, err)
	}
}

func TestStageStartedCannotBeWritten(t *testing.T) {
	r := newRig(t, Config{})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, _ = r.in.Advance(ctx(), "req-1")
	_, _ = r.in.Advance(ctx(), "req-1")
	if _, err := r.in.Store.UpdateRequest("req-1", func(q *claim.Request) error { q.Stage = claim.StageStarted; return nil }); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("writing STARTED: %v", err)
	}
	if _, err := r.in.Advance(ctx(), "req-1"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("advance at WT_READY has nothing to do: %v", err)
	}
}

func TestPreparerThatDidNotStopWritingIsNotReady(t *testing.T) {
	r := newRig(t, Config{})
	r.pr.stopped = false
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, _ = r.in.Advance(ctx(), "req-1")
	req, err := r.in.Advance(ctx(), "req-1")
	if err == nil {
		t.Fatalf("not-stopped preparer accepted: %+v", req)
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	if cur.Stage != claim.StageIssuePublished || cur.Ready != nil || cur.Attempts != 1 || !cur.Retryable {
		t.Errorf("record after refused ready: %+v", cur)
	}
	r.pr.stopped = true
	cur, err = r.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageWTReady || cur.Retryable {
		t.Errorf("retry: %+v %v", cur, err)
	}
	if len(r.pr.slots) != 2 || r.pr.slots[0] != r.pr.slots[1] {
		t.Errorf("a retry must reuse the slot it holds: %v", r.pr.slots)
	}
}

func TestTimeoutIsARetryableFailure(t *testing.T) {
	pub := &fakePub{issue: 1948, block: true}
	r := rigAt(t, t.TempDir(), Config{StepTimeout: 20 * time.Millisecond}, pub, &fakePrep{stopped: true})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, err := r.in.Advance(ctx(), "req-1")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want deadline error, got %v", err)
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	if cur.Stage != claim.StageACK || !cur.Retryable || cur.Attempts != 1 || !strings.Contains(cur.LastError, "timeout") {
		t.Errorf("timeout record: %+v", cur)
	}
	pub.mu.Lock()
	pub.block = false
	pub.mu.Unlock()
	cur, err = r.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageIssuePublished || cur.Retryable || cur.LastError != "" {
		t.Errorf("recovered: %+v %v", cur, err)
	}
	if cur.Attempts != 1 {
		t.Errorf("attempt history must survive recovery, got %d", cur.Attempts)
	}
}

func TestBackpressureParksThenRejectsAndPromotesByPriority(t *testing.T) {
	r := newRig(t, Config{MaxInFlight: 2, MaxParked: 2})
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
	// Drive a to STARTED; that frees one in-flight place.
	_, _ = r.in.Advance(ctx(), "a")
	_, _ = r.in.Advance(ctx(), "a")
	if _, err := startAgent(t, r, "a", 1); err != nil {
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
			req, err := acceptRetry(r.in, "req-"+string(rune('a'+i)), []byte("p"+string(rune('a'+i))), i)
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
			_, err := acceptRetry(r.in, "same", []byte("payload-"+string(rune('a'+i%2))), 1)
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

func TestWorktreeNeedsAFreeExclusiveSlot(t *testing.T) {
	r := newRig(t, Config{Slots: []string{"wt-1"}})
	_, _ = r.in.Accept("a", []byte("1"), 1)
	_, _ = r.in.Accept("b", []byte("2"), 1)
	for _, k := range []string{"a", "b"} {
		if _, err := r.in.Advance(ctx(), k); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := r.in.Advance(ctx(), "a"); err != nil {
		t.Fatal(err)
	}
	before := r.pr.calls.Load()
	_, err := r.in.Advance(ctx(), "b")
	if !errors.Is(err, core.ErrNoSlot) {
		t.Fatalf("b without a free slot: %v", err)
	}
	if r.pr.calls.Load() != before {
		t.Error("the worktree was prepared without a slot")
	}
	cur, _ := r.in.Store.GetRequest("b")
	if cur.Parked != claim.ParkNoSlot || cur.Slot != "" || cur.Stage != claim.StageIssuePublished {
		t.Errorf("b record: %+v", cur)
	}
	if owner, _ := r.in.Store.SlotOwner("wt-1"); owner != "a" {
		t.Errorf("slot owner = %q", owner)
	}
	// a is done with its tree: slot returns, b is promoted and prepared in it.
	if _, err := startAgent(t, r, "a", 1); err != nil {
		t.Fatal(err)
	}
	if err := r.in.ReleaseWorktree("a"); err != nil {
		t.Fatal(err)
	}
	if got, err := r.in.Pump(); err != nil || len(got) != 1 || got[0].Key != "b" {
		t.Fatalf("pump: %+v %v", got, err)
	}
	cur, err = r.in.Advance(ctx(), "b")
	if err != nil || cur.Stage != claim.StageWTReady || cur.Slot != "wt-1" {
		t.Errorf("b after promotion: %+v %v", cur, err)
	}
}

func TestReleaseWorktreeOnlyAfterStart(t *testing.T) {
	r := newRig(t, Config{})
	_, _ = r.in.Accept("a", []byte("1"), 1)
	_, _ = r.in.Advance(ctx(), "a")
	_, _ = r.in.Advance(ctx(), "a")
	if err := r.in.ReleaseWorktree("a"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("release before the agent started: %v", err)
	}
}

func TestRestartResumesWithoutRepeatingEffects(t *testing.T) {
	dir := t.TempDir()
	r1 := rigAt(t, dir, Config{}, &fakePub{issue: 1948}, &fakePrep{stopped: true})
	_, _ = r1.in.Accept("req-1", []byte("p"), 1)
	if _, err := r1.in.Advance(ctx(), "req-1"); err != nil {
		t.Fatal(err)
	}
	// The process restarts: new store, new intake, same directory.
	r2 := rigAt(t, dir, Config{}, &fakePub{issue: 1948}, &fakePrep{stopped: true})
	cur, err := r2.in.Store.GetRequest("req-1")
	if err != nil || cur.Stage != claim.StageIssuePublished || cur.Issue != 1948 {
		t.Fatalf("durable record: %+v %v", cur, err)
	}
	cur, err = r2.in.Advance(ctx(), "req-1")
	if err != nil || cur.Stage != claim.StageWTReady {
		t.Fatalf("resume: %+v %v", cur, err)
	}
	if r2.pub.count() != 0 {
		t.Errorf("the issue was published again after restart")
	}
	if _, err := r2.in.Accept("req-1", []byte("p"), 1); err != nil {
		t.Errorf("replayed request after restart: %v", err)
	}
	if r2.pub.count() != 0 {
		t.Error("a replayed request caused a new effect")
	}
}

func TestPublisherErrorsAreRecordedNotLost(t *testing.T) {
	pub := &fakePub{err: errors.New("HTTP 429 secondary rate limit")}
	r := rigAt(t, t.TempDir(), Config{}, pub, &fakePrep{stopped: true})
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if _, err := r.in.Advance(ctx(), "req-1"); err == nil {
		t.Fatal("want error")
	}
	cur, _ := r.in.Store.GetRequest("req-1")
	if cur.Stage != claim.StageACK || cur.Attempts != 1 || !strings.Contains(cur.LastError, "429") {
		t.Errorf("failure not recorded: %+v", cur)
	}
}
