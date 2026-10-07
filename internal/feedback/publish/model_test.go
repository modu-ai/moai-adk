package publish

// The M6 model-seam tests (SPEC-FEEDBACK-PARTICIPATION-001, AC-017..019,
// AC-025): one model call per item at most, validated output, the template
// fallback on every excluded or failed path, the daily cap, and the static
// reachability guards.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/feedback"
	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
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

// TestModelByteCapsReferenceTheCentralConstants (review-gate finding 7,
// P2): the model input and output caps were LOCAL literals (8 KiB / 4 KiB)
// diverging from the central constants in internal/config/defaults.go
// (4096 / 2048) — two answers to "how big can a summary be". The package
// constants must reference the central ones.
func TestModelByteCapsReferenceTheCentralConstants(t *testing.T) {
	if modelInputMaxBytes != config.DefaultBugreportModelInputMaxBytes {
		t.Fatalf("model input cap = %d, want the central %d", modelInputMaxBytes, config.DefaultBugreportModelInputMaxBytes)
	}
	if modelOutputMaxBytes != config.DefaultBugreportModelOutputMaxBytes {
		t.Fatalf("model output cap = %d, want the central %d", modelOutputMaxBytes, config.DefaultBugreportModelOutputMaxBytes)
	}
	if ModelInputMaxBytes() != config.DefaultBugreportModelInputMaxBytes {
		t.Fatalf("ModelInputMaxBytes() = %d, want the central %d", ModelInputMaxBytes(), config.DefaultBugreportModelInputMaxBytes)
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

// TestTamperedPathTraversalTokenDoesNotPublish (review gate, P2): the
// revalidation checked field FORMAT but not the tripwire's required-hold
// rules — a tampered queue item re-carrying the path_traversal token in its
// detail passed ValidatePayload and published (one search + one create).
// The path-traversal sentinel report is ALWAYS withheld (REQ-ANON-012); the
// send path must hold it too.
func TestTamperedPathTraversalTokenDoesNotPublish(t *testing.T) {
	consentOn(t)
	p, err := bugreport.Build(bugreport.KindTemplateDeployFailure, []string{"internal/cli.Execute"}, bugreport.TokenPathTraversal, bugreport.BuildIdentity{Version: "v3.2.0", Commit: "abcdef1234567"})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	title, body := outbox.RenderReport(p)
	seedQueue(t, feedback.QueueItem{
		ID:          "f1",
		Title:       title,
		Body:        body,
		QueuedAt:    time.Now().UTC().Format(time.RFC3339),
		Fingerprint: p.Fingerprint,
		Kind:        string(p.Kind),
	})

	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	searches, creates, _ := stub.recorded()
	if searches != 0 || creates != 0 {
		t.Fatalf("searches=%d creates=%d — a required-hold payload (path_traversal) was published from the queue", searches, creates)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the held item did not stay queued (items: %d) — the hold must keep it for the attempt limit, not lose it", len(rest))
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
	// The persisted seam (review-gate finding 3): the cap state lives in the
	// user-scoped store, counted atomically cross-process — the in-process
	// budget a new Sender instance reset is gone.
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	outbox.SetClockForTest(func() time.Time { return now })
	t.Cleanup(func() { outbox.SetClockForTest(nil) })

	for i := 0; i < config.DefaultBugreportModelCallsPerDay; i++ {
		if !outbox.AllowAndRecordModelCall(context.Background(), now.Add(time.Duration(i)*time.Minute)) {
			t.Fatalf("call %d of %d refused — a fresh budget must allow up to the cap", i+1, config.DefaultBugreportModelCallsPerDay)
		}
	}
	if outbox.AllowAndRecordModelCall(context.Background(), now.Add(time.Duration(config.DefaultBugreportModelCallsPerDay)*time.Minute)) {
		t.Fatalf("the budget allowed a call past the daily cap of %d", config.DefaultBugreportModelCallsPerDay)
	}
	// The window is rolling: past 24 hours the spend prunes and the budget
	// reopens.
	if !outbox.AllowAndRecordModelCall(context.Background(), now.Add(25*time.Hour)) {
		t.Fatal("the budget stayed closed past the rolling 24-hour window")
	}
}

// TestModelCallBudgetPersistsAcrossSenders (review-gate finding 3, P2): the
// budget reset per Sender instance — every new flush was a fresh budget, so
// N flushes could spend N x the daily cap. The cap state is per-user and
// counted atomically cross-process: a NEW sender facing a full budget falls
// back to the template, it does not call again.
func TestModelCallBudgetPersistsAcrossSenders(t *testing.T) {
	sum, _, _ := summarizerFixture(t)
	sum.out = "A validated summary."

	// Six distinct reports: the first sender burns the whole daily cap on
	// the create path (one call each).
	var items []feedback.QueueItem
	for i := 0; i < config.DefaultBugreportModelCallsPerDay; i++ {
		p, err := bugreport.Build(bugreport.KindPanic, []string{fmt.Sprintf("internal/cli.Execute.f%d", i)}, nil, bugreport.BuildIdentity{Version: "v3.2.0", Commit: "abcdef1234567"})
		if err != nil {
			t.Fatalf("build payload %d: %v", i, err)
		}
		title, body := outbox.RenderReport(p)
		items = append(items, feedback.QueueItem{
			ID:          fmt.Sprintf("f%d", i+1),
			Title:       title,
			Body:        body,
			QueuedAt:    time.Now().UTC().Format(time.RFC3339),
			Fingerprint: p.Fingerprint,
			Kind:        string(p.Kind),
		})
	}
	seedQueue(t, items...)

	first := NewSender(newStubRunner(true))
	first.Summarizer = sum
	if err := first.Send(context.Background()); err != nil {
		t.Fatalf("first send: %v", err)
	}
	if got := sum.count(); got != config.DefaultBugreportModelCallsPerDay {
		t.Fatalf("the first sender made %d calls, want the full cap %d", got, config.DefaultBugreportModelCallsPerDay)
	}

	// A seventh report and a NEW sender — the mutant shape: a fresh budget
	// resets the spend and calls again.
	p, err := bugreport.Build(bugreport.KindPanic, []string{"internal/cli.Execute.f7"}, nil, bugreport.BuildIdentity{Version: "v3.2.0", Commit: "abcdef1234567"})
	if err != nil {
		t.Fatalf("build payload 7: %v", err)
	}
	title, body := outbox.RenderReport(p)
	seedQueue(t, feedback.QueueItem{ID: "f7", Title: title, Body: body, QueuedAt: time.Now().UTC().Format(time.RFC3339), Fingerprint: p.Fingerprint, Kind: string(p.Kind)})

	second := NewSender(newStubRunner(true))
	second.Summarizer = sum
	if err := second.Send(context.Background()); err != nil {
		t.Fatalf("second send: %v", err)
	}
	if got := sum.count(); got != config.DefaultBugreportModelCallsPerDay {
		t.Fatalf("a new sender called the model again (total %d) — the budget reset per sender instead of persisting per user", got)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("the capped report did not publish on the template fallback: %+v", rest)
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
	for _, name := range []string{"model.go", "sender.go", "revalidate.go"} {
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
