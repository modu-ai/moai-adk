package cli

// managed_loopback_test.go — SPEC-FACTORY-MANAGED-SESSION-001 M5 acceptance
// coverage: AC-MS-015 (loopback round trip). No second host and no network:
// the broker is a temp store, the session is a fake model that acts on the
// injected prompt the way the real model is instructed to.

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestManagedSessionLoopbackRoundTrip(t *testing.T) {
	const body = "LOOPBACK_SECRET_BODY"
	t.Run("real path", func(t *testing.T) {
		out := runManagedLoopback(t, "managed-loop-real", body, nil, nil)
		if v := out.violations(body); len(v) != 0 {
			t.Fatalf("round trip violated: %v\n%+v", v, out)
		}
	})
	t.Run("mutant skipping the claim is caught", func(t *testing.T) {
		skip := func(*factorymsg.Store) ([]factorymsg.Claim, error) { return nil, nil }
		out := runManagedLoopback(t, "managed-loop-noclaim", body, skip, nil)
		if len(out.violations(body)) == 0 {
			t.Fatalf("a session that never claims passed the round trip: %+v", out)
		}
	})
	t.Run("mutant injecting the body is caught", func(t *testing.T) {
		leak := func(runID string, claims []factorymsg.Claim) string {
			return managedFactoryInboxPrompt(runID, claims) + body
		}
		out := runManagedLoopback(t, "managed-loop-leak", body, nil, leak)
		if len(out.violations(body)) == 0 {
			t.Fatalf("a body-bearing prompt passed the round trip: %+v", out)
		}
	})
}

// loopbackOutcome is what one driven round trip observed.
type loopbackOutcome struct {
	prompt       string // the injected inbox prompt ("" when none was injected)
	bodyRead     string // what the claim-token body lookup returned
	acked        int
	pending      int
	sessionError error
}

// violations lists every way the outcome departs from the REQ-MS-013/015
// round trip: claim -> metadata-only injection -> body lookup -> receipt,
// ending acknowledged+1 / pending-1.
func (o loopbackOutcome) violations(body string) []string {
	var v []string
	if o.sessionError != nil {
		v = append(v, "session error: "+o.sessionError.Error())
	}
	if o.prompt == "" {
		v = append(v, "no inbox prompt was injected")
	}
	if strings.Contains(o.prompt, body) {
		v = append(v, "prompt carries the message body")
	}
	if o.bodyRead != body {
		v = append(v, fmt.Sprintf("body lookup returned %q", o.bodyRead))
	}
	if o.acked != 1 || o.pending != 0 {
		v = append(v, fmt.Sprintf("acknowledged=%d pending=%d, want 1/0", o.acked, o.pending))
	}
	return v
}

// loopbackModel is a fake session that acts on an inbox prompt as the real
// model is told to: read each body by claim token, then record disposition
// and receipt. Receipt is delivery evidence only (REQ-MS-015).
type loopbackModel struct {
	store *factorymsg.Store
	lane  factorymsg.Peer
	done  chan struct{}
	once  sync.Once
	mu    sync.Mutex
	out   loopbackOutcome
}

func (m *loopbackModel) Start() error { return nil }
func (m *loopbackModel) Close() error { return nil }

func (m *loopbackModel) DeliverTurn(prompt string) error {
	if prompt == managedPrimingPrompt {
		return nil
	}
	ctx := context.Background()
	m.mu.Lock()
	m.out.prompt = prompt
	m.mu.Unlock()
	for _, line := range strings.Split(prompt, "\n") {
		var id, token string
		for _, f := range strings.Fields(line) {
			switch {
			case strings.HasPrefix(f, "message_id="):
				id = strings.TrimPrefix(f, "message_id=")
			case strings.HasPrefix(f, "claim_token="):
				token = strings.TrimPrefix(f, "claim_token=")
			}
		}
		if id == "" || token == "" {
			continue
		}
		b, err := m.store.ReadBody(ctx, m.lane, id, token)
		if err != nil {
			return err
		}
		m.mu.Lock()
		m.out.bodyRead = string(b)
		m.mu.Unlock()
		if err := m.store.RecordDisposition(ctx, m.lane, id, token, factorymsg.DispositionAccepted); err != nil {
			return err
		}
		if err := m.store.Receipt(ctx, m.lane, id, token); err != nil {
			return err
		}
	}
	m.once.Do(func() { close(m.done) })
	return nil
}

// runManagedLoopback drives the real delivery loop (driveManagedFactorySession
// over a temp broker) with the fake model. claimOverride / promptOverride
// replace the real claim and prompt steps to build the two-armed mutants.
func runManagedLoopback(t *testing.T, run, body string, claimOverride func(*factorymsg.Store) ([]factorymsg.Claim, error), promptOverride func(string, []factorymsg.Claim) string) loopbackOutcome {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	activateManagedRun(t, root, run)
	store, lane, _ := seedManagedInbox(t, root, run, body)
	defer closeOnCleanup(t, "factory message broker", store)

	model := &loopbackModel{store: store, lane: lane, done: make(chan struct{})}
	var calls atomic.Int64
	claim := func() ([]factorymsg.Claim, error) {
		calls.Add(1)
		if claimOverride != nil {
			return claimOverride(store)
		}
		return claimManagedFactoryInbox(store, os.Getpid(), homestate.CurrentProcessFingerprint())
	}
	toPrompt := func(c []factorymsg.Claim) string {
		if promptOverride != nil {
			return promptOverride(run, c)
		}
		return managedFactoryInboxPrompt(run, c)
	}

	pr, pw := io.Pipe()
	t.Cleanup(func() { _ = pw.Close() })
	idle := make(chan time.Time)
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case idle <- time.Now():
			case <-stop:
				return
			}
		}
	}()
	errCh := make(chan error, 1)
	go func() { errCh <- driveManagedFactorySession(model, pr, idle, claim, toPrompt) }()

	// Finished when the model completed its turn, or — for a mutant that
	// never injects — after enough idle polls to prove none will come.
	deadline := time.After(10 * time.Second)
	for settled := false; !settled; {
		select {
		case <-model.done:
			settled = true
		case err := <-errCh:
			close(stop)
			model.mu.Lock()
			model.out.sessionError = err
			model.mu.Unlock()
			return model.finish(t)
		case <-deadline:
			t.Fatal("round trip did not settle within 10s")
		case <-time.After(10 * time.Millisecond):
			settled = calls.Load() >= 5
		}
	}
	_, _ = pw.Write([]byte("/exit\n"))
	err := <-errCh
	close(stop)
	model.mu.Lock()
	model.out.sessionError = err
	model.mu.Unlock()
	return model.finish(t)
}

func (m *loopbackModel) finish(t *testing.T) loopbackOutcome {
	t.Helper()
	st, err := m.store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.out.acked, m.out.pending = st.Acknowledged, st.Pending
	return m.out
}
