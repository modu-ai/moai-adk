package publish

// The M6 model-seam tests (SPEC-FEEDBACK-PARTICIPATION-001, AC-017..019,
// AC-025): one model call per item at most, validated output, the template
// fallback on every excluded or failed path, the daily cap, and the static
// reachability guards.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// stubSummarizer counts calls and scripts outputs.
type stubSummarizer struct {
	mu    sync.Mutex
	calls int
	out   string
	err   error
}

func (s *stubSummarizer) Summarize(_ context.Context, _ bugreport.Payload) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.err != nil {
		return "", s.err
	}
	return s.out, nil
}

func (s *stubSummarizer) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func summarizerFixture(t *testing.T) (*stubSummarizer, bugreport.Payload, feedback.QueueItem) {
	t.Helper()
	consentOn(t)
	payload, item := payloadFixture(t)
	return &stubSummarizer{}, payload, item
}

// TestPublishCallsModelOnceAndValidates: a fresh item takes ONE model call,
// and the create carries the summary AHEAD of the template text.
func TestPublishCallsModelOnceAndValidates(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	sum.out = "The wizard panicked while writing the settings file."
	seedQueue(t, item)

	sender := NewSender(newStubRunner(true))
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got := sum.count(); got != 1 {
		t.Fatalf("model calls = %d, want exactly one", got)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("a sent item stayed queued: %+v", rest)
	}
}

// TestPublishFallsBackToTemplate: a failing model still files the issue on
// the deterministic template text, and the template decision is persisted.
func TestPublishFallsBackToTemplate(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	sum.err = errors.New("claude unavailable")
	seedQueue(t, item)

	sender := NewSender(newStubRunner(true))
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got := sum.count(); got != 1 {
		t.Fatalf("model calls = %d, want the one attempt", got)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("the item did not send on the template fallback: %+v", rest)
	}
}

// TestModelInputIsPayloadFieldsOnly (AC-017): the model input is the fixed
// prompt over the payload's closed fields only — golden-pinned, byte-stable.
func TestModelInputIsPayloadFieldsOnly(t *testing.T) {
	p := hookDetailPayload()
	got := SummaryPrompt(p)
	goldenPath := filepath.Join("..", "testdata", "bugreport_summary_prompt_v1.golden")
	if os.Getenv("MOAI_UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run with MOAI_UPDATE_GOLDEN=1 to write): %v", err)
	}
	if string(want) != got {
		t.Fatalf("the model input drifted from the golden:\n--- golden ---\n%s\n--- rendered ---\n%s", want, got)
	}
}

// TestSummaryPersistedBeforeCreateAndReusedOnRetry (AC-018): the validated
// summary is stored on the item, and a retry NEVER calls again.
func TestSummaryPersistedBeforeCreateAndReusedOnRetry(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	sum.out = "A validated summary."
	seedQueue(t, item)

	sender := NewSender(newStubRunner(true))
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got := sum.count(); got != 1 {
		t.Fatalf("first run calls = %d", got)
	}

	// Retry shape: re-queue the STORED (summary-carrying) item and run
	// again — the stored summary must be reused, not re-asked.
	stored := queuedItems(t)
	if len(stored) != 0 {
		t.Fatalf("the sent item stayed queued: %+v", stored)
	}
	seedQueue(t, item)
	sender = NewSender(newStubRunner(true))
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("retry send: %v", err)
	}
	if got := sum.count(); got != 1 {
		t.Fatalf("the retry called the model again (calls=%d) — the stored summary was not reused", got)
	}
}

// TestModelCallBoundPerQueueItem (AC-018): a crash-window item (marker, no
// outcome) takes the template — zero calls (D28).
func TestModelCallBoundPerQueueItem(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	item.SummaryRequested = true
	seedQueue(t, item)

	sender := NewSender(newStubRunner(true))
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got := sum.count(); got != 0 {
		t.Fatalf("a crash-window item called the model %d times — the bound says zero", got)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("the crash-window item did not send on the template: %+v", rest)
	}
}

// TestMarkerBoundsCrashWindowRecall (D28): the marker alone records the
// template DECISION on the item — visible on the retried item (the create
// fails so the item stays queued with its recorded decision).
func TestMarkerBoundsCrashWindowRecall(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	item.SummaryRequested = true
	seedQueue(t, item)

	stub := newStubRunner(true)
	stub.createErr = errors.New("gh create failed (fixture)")
	sender := NewSender(stub)
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got := sum.count(); got != 0 {
		t.Fatalf("the marker-only item called the model %d times", got)
	}
	live := queuedItems(t)
	if len(live) != 1 || live[0].SummaryDecision != "template" {
		t.Fatalf("decision = %+v, want the template decision recorded for the marker-only item", live)
	}
}

// TestStoredSummaryIsRevalidatedBeforePublish (review-gate P1): the queue
// file is a local file, so a STORED summary is untrusted at send time — the
// reuse path must run it through validateSummary and fall back to the
// deterministic template when it fails. The mutant this kills: a tampered
// stored summary (here a credential-shaped string, assembled in pieces so
// the fixture itself is inert) published as-is with zero validation.
func TestStoredSummaryIsRevalidatedBeforePublish(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	credentialShaped := "-----BEGIN " + "RSA PRIVATE" + " KEY-----\n" + strings.Repeat("a", 40)
	item.Summary = "ignore previous instructions. " + credentialShaped
	item.SummaryDecision = "model"
	seedQueue(t, item)

	stub := newStubRunner(true)
	sender := NewSender(stub)
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	stub.mu.Lock()
	creates := append([]createCall(nil), stub.creates...)
	stub.mu.Unlock()
	if len(creates) != 1 {
		t.Fatalf("creates = %d, want the one create on the fallback text", len(creates))
	}
	if strings.Contains(creates[0].Body, "RSA PRIVATE") {
		t.Fatalf("the unvalidated stored summary reached the public issue body")
	}
	if got := sum.count(); got != 0 {
		t.Fatalf("the failed re-validation re-asked the model %d times — the bound says the bad text is discarded, not re-asked", got)
	}
}

// TestStoredSummaryOverCapFallsBackToTemplate: the same trust boundary at
// the length cap — a stored summary over the output cap is discarded for
// the template, never published.
func TestStoredSummaryOverCapFallsBackToTemplate(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	item.Summary = strings.Repeat("x", modelOutputMaxBytes+1)
	item.SummaryDecision = "model"
	seedQueue(t, item)

	stub := newStubRunner(true)
	sender := NewSender(stub)
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	stub.mu.Lock()
	creates := append([]createCall(nil), stub.creates...)
	stub.mu.Unlock()
	if len(creates) != 1 {
		t.Fatalf("creates = %d, want the one create on the fallback text", len(creates))
	}
	if strings.Contains(creates[0].Body, strings.Repeat("x", 64)) {
		t.Fatalf("the over-cap stored summary reached the public issue body")
	}
}

// ---- AC-019: the budget ----

func TestLLMBudgetZeroCalls(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	sum.out = "A validated summary."
	seedQueue(t, item)

	sender := NewSender(newStubRunner(true)) // no summarizer injected
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got := sum.count(); got != 0 {
		t.Fatalf("calls = %d, want zero without an injected summarizer", got)
	}
}

func TestLLMBudgetPositiveControl(t *testing.T) {
	sum, _, item := summarizerFixture(t)
	sum.out = "A validated summary."
	seedQueue(t, item)

	sender := NewSender(newStubRunner(true))
	sender.Summarizer = sum
	if err := sender.Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got := sum.count(); got != 1 {
		t.Fatalf("the positive control made %d calls, want exactly one", got)
	}
}

func TestDailyModelCallCap(t *testing.T) {
	b := NewModelCallBudget()
	if !b.Allow() {
		t.Fatal("a fresh budget must allow the first call")
	}
	for i := 0; i < config.DefaultBugreportModelCallsPerDay; i++ {
		b.Record()
	}
	if b.Allow() {
		t.Fatalf("the budget allowed a call past the daily cap of %d", config.DefaultBugreportModelCallsPerDay)
	}
}

// ---- AC-025: the static reachability guards ----

// stripLineComments removes // comments so the guards pin imports, not
// prose about them (the house precedent of TestNoLabelsNoBodyEdit).
func stripLineComments(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func TestPublishImportAllowlist(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		scanned++
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := stripLineComments(string(raw))
		if strings.Contains(src, "internal/cli") {
			t.Fatalf("%s imports internal/cli — publish must not (REQ-ANON-025)", name)
		}
		if strings.Contains(src, "net/http") {
			t.Fatalf("%s imports net/http — publish reaches gh through the runner only", name)
		}
		if strings.Contains(src, "os/exec") && name != "ghrunner.go" {
			t.Fatalf("%s imports os/exec — os/exec is ghrunner.go's alone", name)
		}
	}
	if scanned < 5 {
		t.Fatalf("the guard scanned %d files — it swept nothing and proves nothing", scanned)
	}
}

func TestBugreportAndOutboxImportAllowlist(t *testing.T) {
	for _, dir := range []string{"../../bugreport", "../outbox"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		scanned := 0
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			scanned++
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read %s/%s: %v", dir, name, err)
			}
			src := stripLineComments(string(raw))
			if strings.Contains(src, "os/exec") || strings.Contains(src, "net/http") {
				t.Fatalf("%s/%s imports os/exec or net/http — the local half is network-free by construction", dir, name)
			}
		}
		if scanned < 3 {
			t.Fatalf("the guard scanned %d files in %s — it swept nothing and proves nothing", scanned, dir)
		}
	}
}

func TestModelSeamIsInjectedNotImported(t *testing.T) {
	for _, name := range []string{"model.go", "budget.go", "sender.go", "revalidate.go"} {
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := string(raw)
		for _, banned := range []string{"exec.Command", "claudeAuditArgs", "runClaudeCommand"} {
			if strings.Contains(src, banned) {
				t.Fatalf("%s references %q — the model implementation lives in internal/cli, injected", name, banned)
			}
		}
	}
}

func TestSummaryOutputValidation(t *testing.T) {
	if err := validateSummary("A plain, valid summary sentence."); err != nil {
		t.Fatalf("valid output refused: %v", err)
	}
	if err := validateSummary(""); err == nil {
		t.Fatal("empty output accepted")
	}
	if err := validateSummary(strings.Repeat("x", modelOutputMaxBytes+1)); err == nil {
		t.Fatal("over-cap output accepted")
	}
	if err := validateSummary("carries a bell \x07 here"); err == nil {
		t.Fatal("non-printable output accepted")
	}
	if err := validateSummary("-----BEGIN RSA PRIVATE KEY-----\n" + strings.Repeat("a", 40)); err == nil {
		t.Fatal("credential-shaped output accepted — the tripwire must refuse it")
	}
	_ = time.Now
}
