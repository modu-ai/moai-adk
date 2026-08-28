package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// stubGuardLivenessStore points the render at a throwaway persistence
// directory, so no test reads or writes the operator's real one.
func stubGuardLivenessStore(t *testing.T) *guardliveness.Store {
	t.Helper()
	store := guardliveness.NewStore(t.TempDir())
	orig := newGuardLivenessStore
	t.Cleanup(func() { newGuardLivenessStore = orig })
	newGuardLivenessStore = func() (*guardliveness.Store, error) { return store, nil }
	return store
}

// nonCleanResult carries one clean and one non-clean entry under a vocabulary
// invented here, per the seam contract (spec.md §B.1).
func nonCleanResult() guardliveness.Result {
	return guardliveness.Result{
		Clean: &guardliveness.Designation{Values: []string{"alpha"}},
		Entries: []guardliveness.Entry{
			{Subject: "subject-clean", Classifications: []string{"alpha"}, Surface: "settled"},
			{Subject: "subject-fires", Classifications: []string{"beta"}, Surface: "settled"},
		},
	}
}

// failingProducer stands for a forge that cannot be reached, and counts every
// arrival at the query layer so a render that reached it is observable.
type failingProducer struct {
	mu    sync.Mutex
	calls int
}

func (p *failingProducer) Produce(context.Context, guardliveness.Activation) (guardliveness.Result, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return guardliveness.Result{}, context.DeadlineExceeded
}

func (p *failingProducer) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// stubUnreachableProducer installs the unreachable forge for the duration of a
// test.
//
// The handler tests below use it rather than a succeeding producer for a
// reason that is itself a property of the design: in production the refresh is
// abandoned unawaited and the render reads a verdict from an EARLIER
// activation, but under test the refresh runs inline, so a succeeding producer
// would overwrite the seeded verdict before the render read it and the test
// would be asserting about this activation's own result. An unreachable forge
// leaves the seeded verdict standing, which is the arrangement the render
// actually meets.
func stubUnreachableProducer(t *testing.T) *failingProducer {
	t.Helper()
	orig := guardLivenessProducer
	t.Cleanup(func() { guardLivenessProducer = orig })
	p := &failingProducer{}
	guardLivenessProducer = p
	return p
}

// AC-GDL-005 — the advisory arrives with the operator issuing no query and
// supplying no guard identifier, and AC-GDL-010(a) — it arrives through the
// EXISTING session-start additional-context block.
func TestSessionStart_GuardLivenessAdvisoryArrivesWithNoOperatorInput(t *testing.T) {
	store := stubGuardLivenessStore(t)
	stubUnreachableProducer(t)
	root := t.TempDir()
	if err := store.Save(root, nonCleanResult(), time.Now().Add(-20*time.Minute-30*time.Second)); err != nil {
		t.Fatalf("seed the persisted result: %v", err)
	}

	h := NewSessionStartHandler(nil)
	// An ordinary session start. Nothing here names a guard, a workflow file,
	// or the liveness feature.
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:     "sess-guard-liveness-render",
		CWD:           root,
		ProjectDir:    root,
		HookEventName: "SessionStart",
		Source:        "startup",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// (a) the advisory rendered.
	if out.HookSpecificOutput == nil {
		t.Fatal("session start produced no hook-specific output — the advisory has no surface to arrive on")
	}
	ctxText := out.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ctxText, "subject-fires") {
		t.Fatalf("the advisory did not reach the session-start additional context:\n%s", ctxText)
	}
	if strings.Contains(ctxText, "subject-clean") {
		t.Errorf("the advisory names a clean subject:\n%s", ctxText)
	}
	if got, want := out.HookSpecificOutput.HookEventName, string(EventSessionStart); got != want {
		t.Errorf("advisory joined a block named %q, want %q", got, want)
	}

	// (b) the rendering path consumed no operator-supplied identifier or query
	// string: varying every operator-authored input field leaves the advisory
	// byte-identical.
	loud, err := h.Handle(context.Background(), &HookInput{
		SessionID:          "sess-guard-liveness-render",
		CWD:                root,
		ProjectDir:         root,
		HookEventName:      "SessionStart",
		Source:             "startup",
		Prompt:             "moai guard liveness docs-i18n-check.yml",
		CustomInstructions: "report every guard by name",
		AgentType:          "manager-develop",
	})
	if err != nil {
		t.Fatalf("Handle with operator input: %v", err)
	}
	if loud.HookSpecificOutput == nil {
		t.Fatal("second session start produced no hook-specific output")
	}
	if loud.HookSpecificOutput.AdditionalContext != ctxText {
		t.Fatalf("the advisory changed with operator-supplied input — the rendering path consumed it:\n%s\n---\n%s",
			ctxText, loud.HookSpecificOutput.AdditionalContext)
	}
}

// AC-GDL-010(a), call-site half — the advisory is emitted through the shared
// contributor helper from inside the already-registered session-start handler,
// which is a property of the deliverable's own wiring rather than of any run.
func TestGuardLivenessJoinsThroughTheExistingContributorHelper(t *testing.T) {
	src, err := os.ReadFile("session_start.go")
	if err != nil {
		t.Fatalf("read the handler: %v", err)
	}
	if !strings.Contains(string(src), "appendAdditionalContext(out, guardLivenessAdvisory(") {
		t.Fatal("the advisory does not join through appendAdditionalContext from inside sessionStartHandler.Handle")
	}
}

// AC-GDL-010(b) — the count of session-start handlers is unchanged from its
// measured baseline of 4. A handler is a type whose EventType() returns
// EventSessionStart; the scan mirrors the criterion's own two-line-context
// grep so both declaration forms are captured.
func TestSessionStartHandlerCountIsUnchanged(t *testing.T) {
	const baseline = 4

	names, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	var declaring []string
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if countsAsSessionStartHandler(t, name) {
			declaring = append(declaring, name)
		}
	}
	if len(declaring) != baseline {
		t.Fatalf("session-start handlers = %d %v, want %d — a second advisory surface was opened", len(declaring), declaring, baseline)
	}
}

func countsAsSessionStartHandler(t *testing.T, name string) bool {
	t.Helper()
	src, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}

	lines := strings.Split(string(src), "\n")
	for i, line := range lines {
		if !strings.Contains(line, "EventType() EventType") {
			continue
		}
		// Two lines of trailing context, matching the criterion's own
		// `grep -A2`, so the single-line and multi-line declaration forms are
		// both captured. Reading the whole file rather than sliding a window
		// over it: a window loses a declaration that sits in the last two
		// lines of a file, which is exactly where a hastily added second
		// surface would sit.
		end := min(i+3, len(lines))
		if strings.Contains(strings.Join(lines[i:end], "\n"), "return EventSessionStart") {
			return true
		}
	}
	return false
}

// AC-GDL-011 (a)+(b) — with the forge unreachable the advisory still renders
// from the persisted result, and the render path issues zero forge calls.
//
// The producer IS the query layer (evaluator.go), so counting arrivals there
// while exercising only the render is the same count the criterion asks for.
func TestGuardLivenessRenderIssuesNoForgeQuery(t *testing.T) {
	store := stubGuardLivenessStore(t)
	root := t.TempDir()
	if err := store.Save(root, nonCleanResult(), time.Now().Add(-3*time.Hour)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	unreachable := &failingProducer{}
	orig := guardLivenessProducer
	t.Cleanup(func() { guardLivenessProducer = orig })
	guardLivenessProducer = unreachable

	text := guardLivenessAdvisory(root)
	if !strings.Contains(text, "subject-fires") {
		t.Fatalf("the render did not produce an advisory from the persisted result:\n%s", text)
	}
	if got := unreachable.count(); got != 0 {
		t.Fatalf("the render path issued %d forge call(s), want 0 — session start waits on the network", got)
	}
}

// AC-GDL-011 (c)+(d) — the deliverable declares its own render join bound, the
// declared value is at most 250 ms, and the render completes within it.
func TestGuardLivenessRenderJoinBoundIsDeclaredAndHonoured(t *testing.T) {
	if guardLivenessJoinBound <= 0 {
		t.Fatalf("declared render join bound = %v — an undeclared bound is an unbounded one", guardLivenessJoinBound)
	}
	if guardLivenessJoinBound > 250*time.Millisecond {
		t.Fatalf("declared render join bound = %v, want at most 250ms", guardLivenessJoinBound)
	}

	store := stubGuardLivenessStore(t)
	root := t.TempDir()
	if err := store.Save(root, nonCleanResult(), time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	start := time.Now()
	text := guardLivenessAdvisory(root)
	elapsed := time.Since(start)
	if text == "" {
		t.Fatal("the render produced nothing on a result carrying a non-clean entry")
	}
	if elapsed > guardLivenessJoinBound {
		t.Fatalf("render took %v, beyond the declared %v bound", elapsed, guardLivenessJoinBound)
	}
}

// Before any refresh has completed there is no persisted result, and silence is
// the honest outcome — an advisory would be reporting a set nothing evaluated.
func TestGuardLivenessRenderIsSilentWithNoPersistedResult(t *testing.T) {
	stubGuardLivenessStore(t)
	if text := guardLivenessAdvisory(t.TempDir()); text != "" {
		t.Fatalf("the render spoke with nothing persisted: %q", text)
	}
}

// AC-GDL-013 at the host surface — a contract-violating persisted result
// reaches the operator as a named violation, never as silence.
func TestSessionStart_GuardLivenessReportsAContractViolation(t *testing.T) {
	store := stubGuardLivenessStore(t)
	stubUnreachableProducer(t)
	root := t.TempDir()

	violating := guardliveness.Result{
		Clean:   nil,
		Entries: []guardliveness.Entry{{Subject: "subject-1", Classifications: []string{"alpha"}, Surface: "settled"}},
	}
	if err := store.Save(root, violating, time.Now().Add(-10*time.Minute)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	h := NewSessionStartHandler(nil)
	out, err := h.Handle(context.Background(), &HookInput{
		SessionID:     "sess-guard-liveness-violation",
		CWD:           root,
		ProjectDir:    root,
		HookEventName: "SessionStart",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out.HookSpecificOutput == nil || !strings.Contains(out.HookSpecificOutput.AdditionalContext, "no clean-value designation") {
		t.Fatalf("the contract violation did not reach the operator:\n%#v", out.HookSpecificOutput)
	}
}
