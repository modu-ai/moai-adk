package cli

// managed_factory_session_test.go — SPEC-FACTORY-MANAGED-SESSION-001 M1
// acceptance coverage: AC-MS-003 (launch-pending rollback), AC-MS-004
// (metadata-only inbox prompt), AC-MS-006 (receipt acknowledges), AC-MS-007
// (owned stream flags), AC-MS-008 (operator/inbox serial queue). The Codex
// App Server owner (AC-MS-001/002) is M2 scope; the loopback round trip
// (AC-MS-015) is M5 scope.

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// activateManagedRun provisions the factory state row ValidateActiveRun
// demands, so launch-pending registration has a live run to join.
func activateManagedRun(t *testing.T, root, run string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: run, Backend: "claude", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

// seedManagedInbox registers a bound lane peer owned by the test process plus
// a leader sender, sends one status request carrying body, and returns the
// store, the bound lane peer, and the sent envelope.
func seedManagedInbox(t *testing.T, root, run, body string) (*factorymsg.Store, factorymsg.Peer, factorymsg.Envelope) {
	t.Helper()
	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	start := homestate.CurrentProcessFingerprint()
	if start == "" {
		t.Fatal("test process identity unavailable")
	}
	lane := factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "claude",
		Role: kanban.RoleLane, Slot: kanban.FactoryLaneLabel(1), SessionUUID: "managed-lane",
		Generation: 1, PID: os.Getpid(), ProcessStart: start,
	}
	lane, err = store.RegisterPeer(context.Background(), lane)
	if err != nil {
		t.Fatal(err)
	}
	senderStart, state := homestate.ProbeProcessIdentity(os.Getppid())
	if state != homestate.ProcessIdentityLive || senderStart == "" {
		t.Fatal("test parent process identity unavailable")
	}
	leader := lane
	leader.Role, leader.Slot = kanban.RoleLeader, kanban.RoleLeader
	leader.SessionUUID = "managed-leader"
	leader.PID, leader.ProcessStart = os.Getppid(), senderStart
	leader.Generation = 1
	leader, err = store.RegisterPeer(context.Background(), leader)
	if err != nil {
		t.Fatal(err)
	}
	env, err := store.Send(context.Background(), factorymsg.SendRequest{
		From: leader, To: lane, Kind: factorymsg.KindStatusRequest,
		IdempotencyKey: "managed-once", TaskRef: "t1", CorrelationID: "c1",
		TTL: time.Minute, Payload: []byte(body),
	})
	if err != nil {
		t.Fatal(err)
	}
	return store, lane, env
}

// The queue is arrival-order FIFO, one turn at a time (REQ-MS-006). Each
// arrangement is proven both ways — operator arriving first and inbox
// arriving first — at the queue and at the driver, so a queue that prefers
// either source (or picks the last turn) fails this test.
func TestManagedQueueSerializesOperatorAndInbox(t *testing.T) {
	t.Run("queue serves arrival order: operator then inbox", func(t *testing.T) {
		q := &managedTurnQueue{}
		q.PushOperator("operator-asks")
		q.PushOperator("   ") // blank operator lines are dropped, never delivered
		q.PushInboxBatch("inbox-batch-prompt")
		first, ok := q.Next()
		if !ok || first.prompt != "operator-asks" || first.fromInbox {
			t.Fatalf("first turn=%+v ok=%v, want the first-arrived turn (operator) first", first, ok)
		}
		second, ok := q.Next()
		if !ok || second.prompt != "inbox-batch-prompt" || !second.fromInbox {
			t.Fatalf("second turn=%+v ok=%v, want the later-arrived turn (inbox) second", second, ok)
		}
		if _, ok := q.Next(); ok {
			t.Fatal("queue still holds a turn after both were delivered")
		}
	})

	t.Run("queue serves arrival order: inbox then operator", func(t *testing.T) {
		q := &managedTurnQueue{}
		q.PushInboxBatch("inbox-batch-prompt")
		q.PushOperator("operator-asks")
		first, ok := q.Next()
		if !ok || first.prompt != "inbox-batch-prompt" || !first.fromInbox {
			t.Fatalf("first turn=%+v ok=%v, want the first-arrived turn (inbox) first", first, ok)
		}
		second, ok := q.Next()
		if !ok || second.prompt != "operator-asks" || second.fromInbox {
			t.Fatalf("second turn=%+v ok=%v, want the later-arrived turn (operator) second", second, ok)
		}
		if _, ok := q.Next(); ok {
			t.Fatal("queue still holds a turn after both were delivered")
		}
	})

	t.Run("driver claims only when idle and serves operator then inbox in arrival order", func(t *testing.T) {
		sess := &fakeManagedSession{began: make(chan fakeTurnSignal, 8)}
		pr, pw := io.Pipe()
		t.Cleanup(func() { _ = pw.Close() })
		// Buffered: ticks fired while a held turn keeps the driver busy must
		// not block the test on delivery.
		idle := make(chan time.Time, 8)
		var claimCalls atomic.Int64
		var haveMsg atomic.Bool
		claim := func() ([]factorymsg.Claim, error) {
			claimCalls.Add(1)
			if !haveMsg.Load() {
				return nil, nil
			}
			return []factorymsg.Claim{{
				Envelope:   factorymsg.Envelope{ID: "m1", Kind: factorymsg.KindStatusRequest, SenderSlot: kanban.FactoryLaneLabel(1), TaskRef: "t1"},
				ClaimToken: "tok-1",
			}}, nil
		}
		errCh := make(chan error, 1)
		go func() {
			errCh <- driveManagedFactorySession(sess, pr, idle, claim, func(c []factorymsg.Claim) string { return managedFactoryInboxPrompt("managed-driver-run", c) })
		}()

		// The priming turn is the first delivery; release it so the driver
		// reaches its idle select.
		if got := sess.waitBegan(t); got != managedPrimingPrompt {
			t.Fatalf("priming turn=%q", got)
		}
		sess.release()

		if _, err := pw.Write([]byte("operator-asks\n")); err != nil {
			t.Fatal(err)
		}
		idle <- time.Time{}
		if got := sess.waitBegan(t); got != "operator-asks" {
			t.Fatalf("turn after input=%q, want the first-arrived operator line", got)
		}
		// Busy: the driver is blocked inside DeliverTurn, so idle ticks must
		// not produce a claim. The baseline is the count at hold entry — an
		// idle tick may legitimately have claimed before the operator line
		// was processed.
		claimsAtHold := claimCalls.Load()
		for i := 0; i < 3; i++ {
			idle <- time.Time{}
		}
		if n := claimCalls.Load(); n != claimsAtHold {
			t.Fatalf("claim ran while the session was busy: %d -> %d", claimsAtHold, n)
		}
		sess.release()

		// Idle: the first claim finds nothing; only after the inbox message
		// appears does the claimed batch become the next turn.
		haveMsg.Store(true)
		idle <- time.Time{}
		inbox := sess.waitBegan(t)
		// The claim's lease means the next idle poll yields nothing — the
		// store owns redelivery, the driver never re-sends on its own.
		haveMsg.Store(false)
		if !strings.Contains(inbox, "m1") {
			t.Fatalf("inbox turn %q does not carry the claimed message metadata", inbox)
		}
		sess.release()

		if _, err := pw.Write([]byte("/exit\n")); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("driver: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("driver did not exit on /exit within 5s")
		}
		sess.mu.Lock()
		turns := append([]string(nil), sess.turns...)
		sess.mu.Unlock()
		if len(turns) != 3 || turns[0] != managedPrimingPrompt || turns[1] != "operator-asks" || !strings.Contains(turns[2], "m1") {
			t.Fatalf("turn order=%q, want arrival order (operator, then inbox) after the priming turn", turns)
		}
	})
	t.Run("driver serves inbox then operator when the claim returns with an operator line already waiting", func(t *testing.T) {
		sess := &fakeManagedSession{began: make(chan fakeTurnSignal, 8)}
		pr, pw := io.Pipe()
		t.Cleanup(func() { _ = pw.Close() })
		idle := make(chan time.Time, 8)
		var calls atomic.Int64
		claim := func() ([]factorymsg.Claim, error) {
			if calls.Add(1) != 1 {
				return nil, nil
			}
			// The operator line reaches the driver's input channel while the
			// claim is still running, then the claim returns its batch: the
			// batch (claimed first) is queued before the line is absorbed.
			if _, err := pw.Write([]byte("operator-late\n")); err != nil {
				return nil, err
			}
			time.Sleep(150 * time.Millisecond)
			return []factorymsg.Claim{{
				Envelope:   factorymsg.Envelope{ID: "m2", Kind: factorymsg.KindStatusRequest, SenderSlot: kanban.FactoryLaneLabel(1), TaskRef: "t1"},
				ClaimToken: "tok-2",
			}}, nil
		}
		errCh := make(chan error, 1)
		go func() {
			errCh <- driveManagedFactorySession(sess, pr, idle, claim, func(c []factorymsg.Claim) string { return managedFactoryInboxPrompt("managed-driver-fifo", c) })
		}()
		if got := sess.waitBegan(t); got != managedPrimingPrompt {
			t.Fatalf("priming turn=%q", got)
		}
		sess.release()
		idle <- time.Time{}

		first := sess.waitBegan(t)
		if !strings.Contains(first, "m2") {
			t.Fatalf("first turn after priming = %q, want the first-arrived inbox batch", first)
		}
		// Busy: no new claim while the inbox turn is in flight.
		claimsAtHold := calls.Load()
		for i := 0; i < 3; i++ {
			idle <- time.Time{}
		}
		if n := calls.Load(); n != claimsAtHold {
			t.Fatalf("claim ran while the session was busy: %d -> %d", claimsAtHold, n)
		}
		sess.release()
		if second := sess.waitBegan(t); second != "operator-late" {
			t.Fatalf("second turn = %q, want the later-arrived operator line", second)
		}
		sess.release()
		if _, err := pw.Write([]byte("/exit\n")); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("driver: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("driver did not exit on /exit within 5s")
		}
	})
}

// fakeTurnSignal pairs a began turn with that turn's OWN release gate, so a
// gate can never be latched by the wrong turn when the test re-arms between
// turns (a shared mutable gate field raced here once — per-turn binding
// removes the race structurally).
type fakeTurnSignal struct {
	prompt string
	gate   chan struct{}
}

// fakeManagedSession is a started session from the driver's point of view:
// the driver owns only the delivery loop, never the spawn (Start is the
// owner's lifecycle edge and is covered at the owner level).
type fakeManagedSession struct {
	mu           sync.Mutex
	turns        []string
	began        chan fakeTurnSignal
	held         chan struct{} // the gate of the turn waitBegan delivered last
	failOnPrompt string
}

func (f *fakeManagedSession) Start() error { return nil }

func (f *fakeManagedSession) DeliverTurn(prompt string) error {
	f.mu.Lock()
	f.turns = append(f.turns, prompt)
	f.mu.Unlock()
	if prompt == f.failOnPrompt {
		return errors.New("stream died")
	}
	if f.began == nil {
		// No test observer: the turn completes immediately (the failure-path
		// subtests never wait on turns).
		return nil
	}
	gate := make(chan struct{})
	f.began <- fakeTurnSignal{prompt: prompt, gate: gate}
	<-gate
	return nil
}

func (f *fakeManagedSession) Close() error { return nil }

func (f *fakeManagedSession) waitBegan(t *testing.T) string {
	t.Helper()
	select {
	case sig := <-f.began:
		f.held = sig.gate
		return sig.prompt
	case <-time.After(5 * time.Second):
		t.Fatal("no turn began within 5s")
		return ""
	}
}

// release lets the last delivered turn complete.
func (f *fakeManagedSession) release() { close(f.held) }

func TestManagedSessionOwnsStreamFlags(t *testing.T) {
	for _, bad := range []string{"-p", "--print", "--input-format", "--output-format", "--input-format=text", "--output-format=json"} {
		args := append([]string{"claude", "--model", "x"}, bad)
		if managedStreamFlagViolation(args) == nil {
			t.Errorf("args %v: owned flag %q was not refused", args, bad)
		}
	}
	for _, okArgs := range [][]string{
		{"claude", "--model", "x"},
		{"claude", "--verbose", "--settings", "s.json"},
		{"claude"},
	} {
		if err := managedStreamFlagViolation(okArgs); err != nil {
			t.Errorf("args %v: unexpected refusal %v", okArgs, err)
		}
	}
	// The refusal names the offending flag so the operator reads which
	// argument to move, not just that something was rejected.
	err := managedStreamFlagViolation([]string{"claude", "--print"})
	if err == nil || !strings.Contains(err.Error(), "--print") {
		t.Fatalf("refusal %v does not name the offending flag", err)
	}

	// The forced stream format precedes every operator argument.
	s, err := newManagedStreamSession("claude", "/bin/claude", []string{"/bin/claude", "--model", "m"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := s.cmd.Args
	wantHead := []string{"/bin/claude", "--print", "--verbose", "--input-format", "stream-json", "--output-format", "stream-json"}
	if len(got) < len(wantHead) {
		t.Fatalf("managed argv=%v shorter than the forced prefix", got)
	}
	for i, want := range wantHead {
		if got[i] != want {
			t.Fatalf("managed argv=%v: position %d = %q, want forced %q", got, i, got[i], want)
		}
	}
	if got[len(wantHead)] != "--model" || got[len(wantHead)+1] != "m" {
		t.Fatalf("managed argv=%v: operator args lost after the forced prefix", got)
	}
}

func TestManagedInboxPromptMetadataOnly(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-prompt"
	activateManagedRun(t, root, run)
	store, _, env := seedManagedInbox(t, root, run, "SECRET_BODY")
	defer closeOnCleanup(t, "factory message broker", store)

	claims, err := claimManagedFactoryInbox(store, os.Getpid(), homestate.CurrentProcessFingerprint())
	if err != nil {
		t.Fatal(err)
	}
	if len(claims) != 1 {
		t.Fatalf("claims=%d, want 1", len(claims))
	}
	prompt := managedFactoryInboxPrompt(run, claims)
	for _, want := range []string{env.ID, claims[0].ClaimToken, claims[0].Kind, claims[0].SenderSlot, claims[0].TaskRef, run} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt misses metadata %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "SECRET_BODY") {
		t.Fatalf("prompt leaked the message body:\n%s", prompt)
	}
	for _, tool := range []string{"factory_msg_body", "factory_msg_receipt"} {
		if !strings.Contains(prompt, tool) {
			t.Errorf("prompt does not name the %s tool:\n%s", tool, prompt)
		}
	}
	if !strings.Contains(prompt, "untrusted") {
		t.Errorf("prompt carries no untrusted-data warning:\n%s", prompt)
	}
}

func TestManagedReceiptAcknowledges(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-receipt"
	activateManagedRun(t, root, run)
	store, lane, env := seedManagedInbox(t, root, run, "receipt-probe-body")
	defer closeOnCleanup(t, "factory message broker", store)

	before, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if before.Pending != 1 {
		t.Fatalf("pending=%d before delivery, want 1", before.Pending)
	}
	claims, err := claimManagedFactoryInbox(store, os.Getpid(), homestate.CurrentProcessFingerprint())
	if err != nil || len(claims) != 1 || claims[0].ID != env.ID {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}
	// The receipt half the injected prompt instructs the model to run:
	// disposition first, then the receipt that turns the claim into delivery
	// evidence (never a completion judgment — REQ-MS-015).
	if err := store.RecordDisposition(context.Background(), lane, claims[0].ID, claims[0].ClaimToken, factorymsg.DispositionAccepted); err != nil {
		t.Fatal(err)
	}
	if err := store.Receipt(context.Background(), lane, claims[0].ID, claims[0].ClaimToken); err != nil {
		t.Fatal(err)
	}
	after, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if after.Acknowledged != 1 || after.Pending != 0 {
		t.Fatalf("after receipt: acknowledged=%d pending=%d, want 1/0", after.Acknowledged, after.Pending)
	}
	// A receipt without a matching claim is refused — the claim token gates
	// the whole settle path.
	if err := store.Receipt(context.Background(), lane, claims[0].ID, "bogus-token"); err == nil {
		t.Fatal("receipt with a stale claim token was accepted")
	}
}

func TestManagedLaunchPendingRollback(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-rollback"
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	activateManagedRun(t, root, run)
	start := homestate.CurrentProcessFingerprint()
	if start == "" {
		t.Fatal("test process identity unavailable")
	}

	// Positive control on the roster read: a directly registered
	// launch-pending row is visible, so the zero assertion below is not an
	// empty sweep.
	env := []string{
		config.EnvMoaiKanbanID + "=" + run,
		config.EnvMoaiKanbanBackend + "=claude",
		config.EnvMoaiFactoryWorker + "=" + kanban.FactoryLaneLabel(1),
	}
	pending, err := registerFactoryLaunchPending(context.Background(), root, env, os.Getpid(), start)
	if err != nil {
		t.Fatal(err)
	}
	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOnCleanup(t, "factory message broker", store)
	roster, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(roster.Lanes) != 1 || roster.Lanes[0].BindingState != factorymsg.BindingLaunchPending {
		t.Fatalf("control roster=%+v, want one launch-pending lane", roster.Lanes)
	}
	if _, err := store.RollbackLaunchPending(context.Background(), pending); err != nil {
		t.Fatal(err)
	}

	// The owner refuses a missing binary, and the failed start leaves no
	// launch-pending row behind (AC-MS-003).
	launchErr := runManagedFactoryClaude("missing-managed-binary", []string{"missing-managed-binary"}, env)
	if launchErr == nil {
		t.Fatal("managed launch with a missing binary returned nil")
	}
	after, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, lane := range after.Lanes {
		if lane.BindingState == factorymsg.BindingLaunchPending {
			t.Fatalf("launch-pending row survived a failed start: %+v", after.Lanes)
		}
	}
}

func TestManagedStreamTurnPump(t *testing.T) {
	t.Run("assistant text forwarded, result ends the turn", func(t *testing.T) {
		out := strings.NewReader(strings.Join([]string{
			`{"type":"system","subtype":"init","session_id":"s-1"}`,
			`{"type":"assistant","message":{"content":[{"type":"text","text":"hello"},{"type":"tool_use","id":"x"}]}}`,
			`{"type":"result","subtype":"success","result":"done","is_error":false}`,
		}, "\n") + "\n")
		var stdout strings.Builder
		var in strings.Builder
		if err := pumpManagedStreamTurn(out, &stdout, &in, "turn-prompt"); err != nil {
			t.Fatalf("pump: %v", err)
		}
		if !strings.Contains(stdout.String(), "hello") {
			t.Errorf("assistant text not forwarded: %q", stdout.String())
		}
		if strings.Contains(stdout.String(), "tool_use") {
			t.Errorf("non-text block leaked to stdout: %q", stdout.String())
		}
		wire := in.String()
		if !strings.Contains(wire, `"type":"user"`) || !strings.Contains(wire, "turn-prompt") {
			t.Errorf("stream-json user message malformed: %s", wire)
		}
	})
	t.Run("failed result fails the turn", func(t *testing.T) {
		out := strings.NewReader(`{"type":"result","result":"boom","is_error":true}` + "\n")
		if err := pumpManagedStreamTurn(out, io.Discard, io.Discard, "p"); err == nil {
			t.Fatal("error result accepted as turn completion")
		}
	})
	t.Run("closed output fails the turn", func(t *testing.T) {
		out := strings.NewReader(`{"type":"system","subtype":"init"}` + "\n")
		err := pumpManagedStreamTurn(out, io.Discard, io.Discard, "p")
		if err == nil || !errors.Is(err, errManagedStreamClosed) {
			t.Fatalf("output close = %v, want errManagedStreamClosed", err)
		}
	})
}

// writeManagedFakeBackend writes a POSIX sh script that answers every stdin
// line with one stream-json event, then sleeps past teardown so Close
// exercises the kill path. POSIX-only: the Windows lane skips the live
// child (crossbuild coverage is the build itself).
func writeManagedFakeBackend(t *testing.T, event string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh fixture backend")
	}
	script := "#!/bin/sh\nwhile IFS= read -r line; do\n  printf '%s\\n' '" + event + "'\ndone\nsleep 30\n"
	path := filepath.Join(t.TempDir(), "fake-managed-backend.sh")
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestManagedStreamSessionLifecycle(t *testing.T) {
	okBackend := writeManagedFakeBackend(t, `{"type":"result","is_error":false}`)
	badBackend := writeManagedFakeBackend(t, `{"type":"result","result":"boom","is_error":true}`)

	t.Run("missing binary fails Start with the backend named", func(t *testing.T) {
		s, err := newManagedStreamSession("claude", "missing-managed-bin", []string{"missing-managed-bin"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		err = s.Start()
		if err == nil || !strings.Contains(err.Error(), "claude") {
			t.Fatalf("start error %v does not name the backend", err)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("close of an unstarted session: %v", err)
		}
	})
	t.Run("turns deliver until teardown kills the child", func(t *testing.T) {
		s, err := newManagedStreamSession("glm", okBackend, []string{okBackend}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Start(); err != nil {
			t.Fatal(err)
		}
		for _, prompt := range []string{"first-turn", "second-turn"} {
			if err := s.DeliverTurn(prompt); err != nil {
				t.Fatalf("DeliverTurn(%q): %v", prompt, err)
			}
		}
		if err := s.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		// The child answered one result per turn and slept past teardown, so
		// Close reaped it by kill; a second Close must not panic or hang.
		if err := s.Close(); err != nil {
			t.Fatalf("second close: %v", err)
		}
	})
	t.Run("failed result fails the live turn", func(t *testing.T) {
		s, err := newManagedStreamSession("claude", badBackend, []string{badBackend}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = s.Close() })
		err = s.DeliverTurn("doomed-turn")
		if err == nil || !strings.Contains(err.Error(), "turn failed") {
			t.Fatalf("live error-result turn = %v, want a turn failure", err)
		}
	})
}

func TestManagedQueueDriverFailurePaths(t *testing.T) {
	// The driver-level start-failure property lives at the owner level now:
	// TestManagedLaunchPendingRollback covers a failed spawn with rollback.
	t.Run("inbox claim error recovers on the next tick", func(t *testing.T) {
		sess := &fakeManagedSession{began: make(chan fakeTurnSignal, 8)}
		pr, pw := io.Pipe()
		t.Cleanup(func() { _ = pw.Close() })
		idle := make(chan time.Time, 4)
		var calls int
		var haveMsg atomic.Bool
		claim := func() ([]factorymsg.Claim, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("broker busy once")
			}
			if !haveMsg.Load() {
				return nil, nil
			}
			return []factorymsg.Claim{{Envelope: factorymsg.Envelope{ID: "m1"}, ClaimToken: "tok"}}, nil
		}
		errCh := make(chan error, 1)
		go func() {
			errCh <- driveManagedFactorySession(sess, pr, idle, claim,
				func(c []factorymsg.Claim) string { return managedFactoryInboxPrompt("managed-retry-run", c) })
		}()
		if got := sess.waitBegan(t); got != managedPrimingPrompt {
			t.Fatalf("priming turn=%q", got)
		}
		sess.release()
		haveMsg.Store(true)
		idle <- time.Time{} // first claim fails, the driver logs and retries
		idle <- time.Time{}
		if inbox := sess.waitBegan(t); !strings.Contains(inbox, "m1") {
			t.Fatalf("turn after claim error=%q, want the claimed message", inbox)
		}
		// Lease discipline again: without this the next idle claim would
		// re-deliver the same message and the driver would never reach /exit.
		haveMsg.Store(false)
		sess.release()
		if _, err := pw.Write([]byte("/exit\n")); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("driver: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("driver did not exit within 5s")
		}
		if calls < 2 {
			t.Fatalf("claim ran %d times, want the retry after the first error", calls)
		}
	})
	t.Run("turn failure ends the driver", func(t *testing.T) {
		sess := &fakeManagedSession{failOnPrompt: "doomed"}
		want := errors.New("stream died")
		err := driveManagedFactorySession(sess, strings.NewReader("doomed\n"), make(chan time.Time), func() ([]factorymsg.Claim, error) { return nil, nil }, func(c []factorymsg.Claim) string { return managedFactoryInboxPrompt("managed-fail-run", c) })
		if err == nil || !strings.Contains(err.Error(), "stream died") {
			t.Fatalf("driver error = %v, want %v", err, want)
		}
	})
}

func TestClaimManagedFactoryInboxWithoutEndpoint(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-empty-claim"
	activateManagedRun(t, root, run)
	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOnCleanup(t, "factory message broker", store)
	claims, err := claimManagedFactoryInbox(store, os.Getpid(), homestate.CurrentProcessFingerprint())
	if err != nil || claims != nil {
		t.Fatalf("claim without a registered endpoint = %v, %v; want no inbox and no error", claims, err)
	}
}

// TestManagedStreamSessionOwnerRuns drives the real owner entry over a fake
// stream backend: the priming turn completes, the operator's /exit ends the
// session cleanly, and the launch-pending row the owner registered stays —
// the bind half belongs to the child session's hook (REQ-MS-002).
func TestManagedStreamSessionOwnerRuns(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh fixture backend")
	}
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-owner-run"
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	activateManagedRun(t, root, run)
	backend := writeManagedFakeBackend(t, `{"type":"result","is_error":false}`)
	env := []string{
		config.EnvMoaiKanbanID + "=" + run,
		config.EnvMoaiKanbanBackend + "=claude",
		config.EnvMoaiFactoryWorker + "=" + kanban.FactoryLaneLabel(1),
	}
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stdinW.WriteString("/exit\n"); err != nil {
		t.Fatal(err)
	}
	if err := stdinW.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stdinR.Close() })

	if err := runManagedFactoryStreamSession(BackendClaude, backend, []string{backend}, env, stdinR); err != nil {
		t.Fatalf("managed owner run: %v", err)
	}
	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOnCleanup(t, "factory message broker", store)
	roster, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(roster.Lanes) != 1 || roster.Lanes[0].BindingState != factorymsg.BindingLaunchPending ||
		roster.Lanes[0].Slot != kanban.FactoryLaneLabel(1) {
		t.Fatalf("roster after a clean run = %+v, want the lane-1 launch-pending row", roster.Lanes)
	}

	// The GLM wrapper with no factory run refuses before any child work and
	// still tears the session down (empty-run-id guard).
	glmErr := runManagedFactoryGlm(backend, []string{backend}, nil)
	if glmErr == nil || !strings.Contains(glmErr.Error(), "factory run id") {
		t.Fatalf("glm owner without a run = %v, want the run-id refusal", glmErr)
	}
}
