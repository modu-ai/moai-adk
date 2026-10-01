package cli

// managed_loopback_test.go — SPEC-FACTORY-MANAGED-SESSION-001 M5 acceptance
// coverage: AC-MS-015 (loopback round trip). No second host and no network:
// the broker is a temp store, the session is a fake model that acts on the
// injected prompt the way the real model is instructed to.

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
	for _, reexec := range []bool{false, true} {
		reexec := reexec
		name := "in-process"
		if reexec {
			name = "re-exec"
		}
		t.Run(name, func(t *testing.T) { runLoopbackArms(t, name, body, reexec) })
	}
}

// runLoopbackArms runs the real path and both mutants against one session
// kind (an in-process fake model, or a re-exec of the test binary).
func runLoopbackArms(t *testing.T, tag, body string, reexec bool) {
	if reexec && runtime.GOOS == "windows" {
		t.Skip("POSIX sh re-exec fixture; the Windows lane proves the owner by cross-build")
	}
	t.Run("real path", func(t *testing.T) {
		out := runManagedLoopback(t, "managed-loop-real-"+tag, body, reexec, nil, nil)
		if v := out.violations(body); len(v) != 0 {
			t.Fatalf("round trip violated: %v\n%+v", v, out)
		}
	})
	t.Run("mutant skipping the claim is caught", func(t *testing.T) {
		skip := func(*factorymsg.Store) ([]factorymsg.Claim, error) { return nil, nil }
		out := runManagedLoopback(t, "managed-loop-noclaim-"+tag, body, reexec, skip, nil)
		if len(out.violations(body)) == 0 {
			t.Fatalf("a session that never claims passed the round trip: %+v", out)
		}
	})
	t.Run("mutant injecting the body is caught", func(t *testing.T) {
		leak := func(runID string, claims []factorymsg.Claim) string {
			return managedFactoryInboxPrompt(runID, claims) + body
		}
		out := runManagedLoopback(t, "managed-loop-leak-"+tag, body, reexec, nil, leak)
		if len(out.violations(body)) == 0 {
			t.Fatalf("a body-bearing prompt passed the round trip: %+v", out)
		}
	})
}

// Env names the re-exec child reads; the child runs only when they are set.
const (
	loopbackEnvRoot = "MOAI_TEST_LOOPBACK_ROOT"
	loopbackEnvRun  = "MOAI_TEST_LOOPBACK_RUN"
	loopbackEnvBody = "MOAI_TEST_LOOPBACK_BODY_OUT"
)

// TestManagedLoopbackChild is the re-exec fake session (AC-MS-015): the test
// binary relaunched as a stream-json child under the managed owner. It reads
// each injected prompt from stdin, then does what the model is told to do —
// body lookup by claim token and receipt — by opening the same on-disk broker
// store from this second process.
func TestManagedLoopbackChild(t *testing.T) {
	root, run := os.Getenv(loopbackEnvRoot), os.Getenv(loopbackEnvRun)
	if root == "" || run == "" {
		t.Skip("re-exec helper: runs only as the loopback child")
	}
	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOnCleanup(t, "factory message broker", store)
	lane, err := store.Peer(context.Background(), "managed-lane")
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64*1024), 1<<20)
	for scanner.Scan() {
		var in struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &in); err != nil {
			t.Fatalf("child input is not stream-json: %v", err)
		}
		if err := actOnInboxPrompt(store, lane, in.Message.Content, func(b []byte) {
			if err := os.WriteFile(os.Getenv(loopbackEnvBody), b, 0o600); err != nil {
				t.Fatal(err)
			}
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Fprintln(os.Stdout, `{"type":"result","is_error":false}`); err != nil {
			t.Fatal(err)
		}
	}
}

// actOnInboxPrompt performs the model's instructed steps for every message in
// an injected prompt: read the body by claim token, record disposition, then
// the receipt. Shared by the in-process and re-exec fake sessions.
func actOnInboxPrompt(store *factorymsg.Store, lane factorymsg.Peer, prompt string, sawBody func([]byte)) error {
	ctx := context.Background()
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
		b, err := store.ReadBody(ctx, lane, id, token)
		if err != nil {
			return err
		}
		sawBody(b)
		if err := store.RecordDisposition(ctx, lane, id, token, factorymsg.DispositionAccepted); err != nil {
			return err
		}
		if err := store.Receipt(ctx, lane, id, token); err != nil {
			return err
		}
	}
	return nil
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
	m.mu.Lock()
	m.out.prompt = prompt
	m.mu.Unlock()
	if err := actOnInboxPrompt(m.store, m.lane, prompt, func(b []byte) {
		m.mu.Lock()
		m.out.bodyRead = string(b)
		m.mu.Unlock()
	}); err != nil {
		return err
	}
	m.once.Do(func() { close(m.done) })
	return nil
}

// runManagedLoopback drives the real delivery loop (driveManagedFactorySession
// over a temp broker). The session is an in-process fake model, or — with
// reexec — the test binary relaunched as a stream-json child owned by a real
// managedStreamSession. claimOverride / promptOverride replace the real claim
// and prompt steps to build the two-armed mutants.
func runManagedLoopback(t *testing.T, run, body string, reexec bool, claimOverride func(*factorymsg.Store) ([]factorymsg.Claim, error), promptOverride func(string, []factorymsg.Claim) string) loopbackOutcome {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	activateManagedRun(t, root, run)
	store, lane, _ := seedManagedInbox(t, root, run, body)
	defer closeOnCleanup(t, "factory message broker", store)

	var (
		session  managedSession
		settled  func() bool
		bodyRead func() string
		mu       sync.Mutex
		prompt   string
	)
	if reexec {
		bodyOut := filepath.Join(t.TempDir(), "body.out")
		script := "#!/bin/sh\nexec \"" + os.Args[0] + "\" -test.run='^TestManagedLoopbackChild$' -- \"$@\"\n"
		bin := filepath.Join(t.TempDir(), "loopback-child.sh")
		if err := os.WriteFile(bin, []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
		env := append(os.Environ(), loopbackEnvRoot+"="+root, loopbackEnvRun+"="+run, loopbackEnvBody+"="+bodyOut)
		stream, err := newManagedStreamSession(BackendClaude, bin, []string{bin}, env)
		if err != nil {
			t.Fatal(err)
		}
		if err := stream.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = stream.Close() }()
		session = stream
		settled = func() bool {
			st, err := store.Status(context.Background())
			return err == nil && st.Acknowledged >= 1
		}
		bodyRead = func() string { b, _ := os.ReadFile(bodyOut); return string(b) }
	} else {
		model := &loopbackModel{store: store, lane: lane, done: make(chan struct{})}
		session = model
		settled = func() bool {
			select {
			case <-model.done:
				return true
			default:
				return false
			}
		}
		bodyRead = func() string { model.mu.Lock(); defer model.mu.Unlock(); return model.out.bodyRead }
	}

	var calls atomic.Int64
	claim := func() ([]factorymsg.Claim, error) {
		calls.Add(1)
		if claimOverride != nil {
			return claimOverride(store)
		}
		return claimManagedFactoryInbox(store, os.Getpid(), homestate.CurrentProcessFingerprint())
	}
	toPrompt := func(c []factorymsg.Claim) string {
		p := managedFactoryInboxPrompt(run, c)
		if promptOverride != nil {
			p = promptOverride(run, c)
		}
		mu.Lock()
		prompt = p
		mu.Unlock()
		return p
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
	go func() { errCh <- driveManagedFactorySession(session, pr, idle, claim, toPrompt) }()

	// Finished when the session completed its turn, or — for a mutant that
	// never injects — after enough idle polls to prove none will come.
	deadline := time.After(20 * time.Second)
	var driveErr error
	for done := false; !done; {
		select {
		case driveErr = <-errCh:
			done = true
		case <-deadline:
			t.Fatal("round trip did not settle within 20s")
		case <-time.After(10 * time.Millisecond):
			if settled() || calls.Load() >= 5 {
				_, _ = pw.Write([]byte("/exit\n"))
				driveErr = <-errCh
				done = true
			}
		}
	}
	close(stop)

	st, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	return loopbackOutcome{prompt: prompt, bodyRead: bodyRead(), acked: st.Acknowledged, pending: st.Pending, sessionError: driveErr}
}
