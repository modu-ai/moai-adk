package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/feedback"
	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
)

// ---- fixtures ----

// payloadFixture returns a valid payload and the queue item the drain
// would have created from it (outbox.RenderReport — the one render
// function).
func payloadFixture(t *testing.T) (bugreport.Payload, feedback.QueueItem) {
	t.Helper()
	payload, err := bugreport.Build(bugreport.KindPanic, []string{"internal/cli.Execute"}, nil, bugreport.BuildIdentity{
		Version: "v3.2.0",
		Commit:  "abcdef1234567",
	})
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	title, body := outbox.RenderReport(payload)
	item := feedback.QueueItem{
		ID:          "f1",
		Title:       title,
		Body:        body,
		QueuedAt:    time.Now().UTC().Format(time.RFC3339),
		Fingerprint: payload.Fingerprint,
		Kind:        string(payload.Kind),
	}
	return payload, item
}

// seedQueue writes items into the user-scoped bugreport queue.
func seedQueue(t *testing.T, items ...feedback.QueueItem) {
	t.Helper()
	store := outbox.BugreportQueueStore()
	err := store.Mutate(func(rec *feedback.QueueRecord) error {
		for _, it := range items {
			rec.LastSeq++
			next := it
			next.ID = it.ID
			rec.Items = append(rec.Items, next)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed queue: %v", err)
	}
}

func queuedItems(t *testing.T) []feedback.QueueItem {
	t.Helper()
	rec, err := outbox.BugreportQueueStore().Load()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	return rec.Items
}

// consentOn/Off write the user-scoped consent file under the test's
// MOAI_HOME.
func consentOn(t *testing.T) {
	t.Helper()
	seedConsentFile(t, "participation:\n  enabled: true\n  asked: true\n")
}

func consentOff(t *testing.T) {
	t.Helper()
	seedConsentFile(t, "participation:\n  enabled: false\n  asked: true\n")
}

func seedConsentFile(t *testing.T, body string) {
	t.Helper()
	home := os.Getenv("MOAI_HOME")
	if home == "" {
		home = t.TempDir()
		t.Setenv("MOAI_HOME", home)
	}
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("seed consent: %v", err)
	}
}

// stubRunner is the gh seam's test double: scripted search results, every
// call recorded, ctx respected (the production runner is ctx-bound too).
type stubRunner struct {
	mu        sync.Mutex
	available bool
	issues    []RemoteIssue
	searchErr error

	searches []string
	repos    []string
	creates  []createCall
	comments []commentCall

	onSearch func(s *stubRunner)
	block    chan struct{} // when non-nil, SearchIssues parks until ctx is done
}

type createCall struct {
	Repo  string
	Title string
	Body  string
}

type commentCall struct {
	Repo   string
	Number int
	Body   string
}

func newStubRunner(available bool) *stubRunner {
	return &stubRunner{available: available}
}

func (s *stubRunner) Available(_ context.Context) bool { return s.available }

func (s *stubRunner) SearchIssues(ctx context.Context, repo, token string) ([]RemoteIssue, error) {
	if s.block != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-s.block:
		}
	}
	s.mu.Lock()
	s.searches = append(s.searches, token)
	hook := s.onSearch
	s.mu.Unlock()
	if hook != nil {
		hook(s)
	}
	return s.issues, s.searchErr
}

func (s *stubRunner) CreateIssue(_ context.Context, repo, title string, body io.Reader) error {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(body)
	s.mu.Lock()
	s.repos = append(s.repos, repo)
	s.creates = append(s.creates, createCall{Repo: repo, Title: title, Body: buf.String()})
	s.mu.Unlock()
	return nil
}

func (s *stubRunner) CommentIssue(_ context.Context, repo string, number int, body io.Reader) error {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(body)
	s.mu.Lock()
	s.repos = append(s.repos, repo)
	s.comments = append(s.comments, commentCall{Repo: repo, Number: number, Body: buf.String()})
	s.mu.Unlock()
	return nil
}

func (s *stubRunner) recorded() (searches int, creates int, comments int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.searches), len(s.creates), len(s.comments)
}

// ---- AC-015: sender discipline ----

func TestSenderChecksConsentPerItem(t *testing.T) {
	consentOn(t)
	_, item1 := payloadFixture(t)
	item2 := item1
	item2.ID = "f2"
	seedQueue(t, item1, item2)

	stub := newStubRunner(true)
	stub.onSearch = func(s *stubRunner) {
		// The user withdraws consent the moment the first item is processed.
		consentOff(t)
	}

	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}

	searches, creates, comments := stub.recorded()
	if searches != 1 || creates != 1 || comments != 0 {
		t.Fatalf("after the consent flip: searches=%d creates=%d comments=%d, want exactly item one's search+create and nothing for item two", searches, creates, comments)
	}
	rest := queuedItems(t)
	if len(rest) != 1 || rest[0].ID != "f2" {
		t.Fatalf("item two did not stay queued: %+v", rest)
	}
}

func TestSenderQuietWithoutGh(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	stub := newStubRunner(false) // gh missing or unauthenticated
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send with gh absent must be quiet, got: %v", err)
	}
	searches, creates, comments := stub.recorded()
	if searches != 0 || creates != 0 || comments != 0 {
		t.Fatalf("gh absent: searches=%d creates=%d comments=%d, want zero", searches, creates, comments)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the item did not stay queued: %+v", rest)
	}
}

func TestSenderNeverRunsOnHookPath(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)
	t.Setenv(config.EnvHookDispatch, "1")

	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send on the hook path must be refused quietly, got: %v", err)
	}
	searches, creates, comments := stub.recorded()
	if searches != 0 || creates != 0 || comments != 0 {
		t.Fatalf("hook path: searches=%d creates=%d comments=%d, want zero", searches, creates, comments)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the item did not stay queued: %+v", rest)
	}
}

func TestSenderTimeBox(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	stub := newStubRunner(true)
	stub.block = make(chan struct{}) // never closed: a stub wedged like a hung gh

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := NewSender(stub).Send(ctx); err != nil {
		t.Fatalf("send: %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("send ran %s on a blocking gh, want it to end within the box", elapsed)
	}
}

// ---- AC-016: existing fingerprint ----

func TestExistingFingerprintGetsOccurrenceComment(t *testing.T) {
	consentOn(t)
	payload, item := payloadFixture(t)
	seedQueue(t, item)

	stub := newStubRunner(true)
	stub.issues = []RemoteIssue{{
		Number: 7,
		Title:  item.Title, // the exact title key
		State:  "open",
		Comments: []RemoteComment{
			{Body: OccurrenceMarkerForTest(payload)},
			{Body: OccurrenceMarkerForTest(payload)},
			{Body: OccurrenceMarkerForTest(payload)},
		},
	}}

	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, comments := stub.recorded()
	if creates != 0 {
		t.Fatalf("an existing issue must never be re-created (%d creates)", creates)
	}
	if comments != 1 {
		t.Fatalf("comments = %d, want exactly one occurrence comment", comments)
	}
	if stub.comments[0].Number != 7 {
		t.Fatalf("commented issue #%d, want #7", stub.comments[0].Number)
	}
	if !IsOccurrenceComment(stub.comments[0].Body) {
		t.Fatalf("the comment does not carry the occurrence marker: %q", stub.comments[0].Body)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("a sent item stayed queued: %+v", rest)
	}
}

func TestExactTitleKeyOnly(t *testing.T) {
	consentOn(t)
	payload, item := payloadFixture(t)
	seedQueue(t, item)

	stub := newStubRunner(true)
	stub.issues = []RemoteIssue{
		{Number: 1, Title: item.Title + " (duplicate report)", State: "open"},
		{Number: 2, Title: "re: " + item.Title, State: "open"},
		{Number: 3, Title: strings.Replace(item.Title, string(payload.Kind), "other", 1), State: "open"},
	}

	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, comments := stub.recorded()
	if comments != 0 {
		t.Fatalf("a near-title match was treated as the same issue (%d comments)", comments)
	}
	if creates != 1 {
		t.Fatalf("creates = %d, want a new issue for the exact-key miss", creates)
	}
}

func TestClosedIssueStillCounts(t *testing.T) {
	consentOn(t)
	payload, item := payloadFixture(t)
	seedQueue(t, item)

	stub := newStubRunner(true)
	stub.issues = []RemoteIssue{{
		Number: 9,
		Title:  item.Title,
		State:  "closed",
		Comments: []RemoteComment{
			{Body: OccurrenceMarkerForTest(payload)},
		},
	}}

	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, comments := stub.recorded()
	if creates != 0 || comments != 1 {
		t.Fatalf("a closed issue still counts: creates=%d comments=%d, want 0/1", creates, comments)
	}
}

func TestOccurrenceCapSkipsComment(t *testing.T) {
	consentOn(t)
	payload, item := payloadFixture(t)
	seedQueue(t, item)

	comments := make([]RemoteComment, config.DefaultBugreportOccurrenceCommentsPerIssue)
	for i := range comments {
		comments[i] = RemoteComment{Body: OccurrenceMarkerForTest(payload)}
	}
	stub := newStubRunner(true)
	stub.issues = []RemoteIssue{{Number: 11, Title: item.Title, State: "open", Comments: comments}}

	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, gotComments := stub.recorded()
	if creates != 0 || gotComments != 0 {
		t.Fatalf("at the cap: creates=%d comments=%d, want the sender to add nothing", creates, gotComments)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the capped item must stay queued: %+v", rest)
	}
}

// ---- review-gate P1: send-time re-validation of the stored body ----

// TestSenderRevalidatesStoredBodyBeforeCreate pins the trust boundary: the
// queue file is a local file, so the stored body is untrusted input. An
// intact item creates with EXACTLY the stored bytes; a body replaced on
// disk after enqueue never reaches gh — the item fails re-validation,
// stays queued, and no search runs either (the refusal precedes every gh
// call).
func TestSenderRevalidatesStoredBodyBeforeCreate(t *testing.T) {
	payload, item := payloadFixture(t)
	freshHome := func() {
		t.Setenv("MOAI_HOME", t.TempDir())
		consentOn(t)
	}

	// (a) Intact: the create carries exactly the stored title and body.
	freshHome()
	stub := newStubRunner(true)
	seedQueue(t, item)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if _, creates, _ := stub.recorded(); creates != 1 {
		t.Fatalf("creates = %d, want the one create for the intact item", creates)
	}
	if stub.creates[0].Body != item.Body || stub.creates[0].Title != item.Title {
		t.Fatalf("the intact item's create drifted from the stored bytes:\n--- create ---\n%s\n--- stored ---\n%s", stub.creates[0].Body, item.Body)
	}

	// (b) Tampered: a body replaced on disk after enqueue.
	freshHome()
	stub = newStubRunner(true)
	tampered := item
	tampered.Body = "private notes with /Users/x/secret paths"
	seedQueue(t, tampered)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if searches, creates, comments := stub.recorded(); searches != 0 || creates != 0 || comments != 0 {
		t.Fatalf("a tampered body reached gh: searches=%d creates=%d comments=%d", searches, creates, comments)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the tampered item did not stay queued: %+v", rest)
	}

	// (c) Spliced: a valid render of ANOTHER payload under this item's
	// title — the marker fields disagree with the title key.
	freshHome()
	stub = newStubRunner(true)
	spliced := item
	other := payload
	other.Fingerprint = "fedcba9876543210"
	_, spliced.Body = outbox.RenderReport(other)
	seedQueue(t, spliced)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if searches, creates, _ := stub.recorded(); searches != 0 || creates != 0 {
		t.Fatalf("a spliced body reached gh: searches=%d creates=%d", searches, creates)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the spliced item did not stay queued: %+v", rest)
	}
}

// TestSenderRevalidatesStoredBodyBeforeComment extends the boundary to the
// comment path: a tampered body must not become a public comment either.
func TestSenderRevalidatesStoredBodyBeforeComment(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	tampered := item
	tampered.Body = "private notes with /Users/x/secret paths"
	seedQueue(t, tampered)

	stub := newStubRunner(true)
	stub.issues = []RemoteIssue{{Number: 7, Title: tampered.Title, State: "open"}}
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if _, creates, comments := stub.recorded(); creates != 0 || comments != 0 {
		t.Fatalf("a tampered body reached the comment path: creates=%d comments=%d", creates, comments)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the tampered item did not stay queued: %+v", rest)
	}
}

// ---- review-gate P2: per-item cross-process send ownership ----

// TestSendSkipsItemOwnedByLiveClaim pins the ownership contract: from the
// duplicate lookup through the outcome record, one item belongs to ONE
// flush. A second flush whose snapshot still carries the item must skip it
// — the live claim says another flush owns it — and send it only after the
// owner releases. Without ownership, two concurrent flushes each searched
// and EACH created the issue (searches=2, creates=2).
func TestSendSkipsItemOwnedByLiveClaim(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	// Another flush holds the item's claim, live (this is exactly the
	// artifact ClaimItemSend leaves behind — the store-relative path shape
	// is pinned here by construction, not by importing the helper).
	claimPath := filepath.Join(os.Getenv("MOAI_HOME"), filepath.FromSlash(bugreport.BugreportStoreDir), "send-"+item.ID+".claim")
	release, err := atomicfile.ClaimSection(context.Background(), claimPath, 0o600, 0, time.Millisecond)
	if err != nil {
		t.Fatalf("the test flush could not claim the item: %v", err)
	}

	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if searches, creates, comments := stub.recorded(); searches != 0 || creates != 0 || comments != 0 {
		t.Fatalf("a flush sent an item another flush owned: searches=%d creates=%d comments=%d", searches, creates, comments)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the owned item did not stay queued: %+v", rest)
	}

	// The owning flush finishes and releases: the next flush sends it.
	_ = release()
	stub = newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send after release: %v", err)
	}
	if _, creates, _ := stub.recorded(); creates != 1 {
		t.Fatalf("creates = %d, want the released item sent", creates)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("the sent item stayed queued: %+v", rest)
	}
}

// TestDeadOwnerSendClaimIsReclaimed: a flush that died holding its claim
// must not wedge the item — the dead owner's claim is reclaimed through
// the same verified-dead rule, and the item sends.
func TestDeadOwnerSendClaimIsReclaimed(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	claimPath := filepath.Join(os.Getenv("MOAI_HOME"), filepath.FromSlash(bugreport.BugreportStoreDir), "send-"+item.ID+".claim")
	identity := atomicfile.BootIDIdentity()
	if identity == "" {
		t.Skip("no boot identity on this platform")
	}
	dead := atomicfile.LockOwner{PID: os.Getpid(), BootID: "previous-boot-" + identity}
	raw, err := json.Marshal(dead)
	if err != nil {
		t.Fatalf("marshal dead owner: %v", err)
	}
	if err := os.WriteFile(claimPath, raw, 0o600); err != nil {
		t.Fatalf("write dead claim: %v", err)
	}

	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if _, creates, _ := stub.recorded(); creates != 1 {
		t.Fatalf("creates = %d, want the dead owner's claim reclaimed and the item sent", creates)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("the sent item stayed queued: %+v", rest)
	}
}

// TestStaleSnapshotDoesNotResendADeletedItem pins the residual of the
// ownership fix: the claim excludes concurrent holders, but a second flush
// that LOADED its snapshot before the first finished still iterates the
// item — and after the first flush's release it acquires the now-free
// claim and re-sends the already-sent (deleted) item. After acquiring the
// claim the item must be re-checked against the LIVE queue: gone means
// skip.
func TestStaleSnapshotDoesNotResendADeletedItem(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	// Flush B parks just before it takes the item's claim; flush A fully
	// sends the item in that window. When B proceeds on its stale
	// snapshot, the item is already gone.
	arrived := make(chan struct{})
	proceed := make(chan struct{})
	prev := beforeClaimForTest
	t.Cleanup(func() { beforeClaimForTest = prev })
	// Only the FIRST caller parks — flush B, whose goroutine reaches the
	// seam before the main test invokes flush A (the main test blocks on
	// the arrival signal before starting A).
	firstCall := true
	beforeClaimForTest = func() {
		if !firstCall {
			return
		}
		firstCall = false
		close(arrived)
		<-proceed
	}

	stubB := newStubRunner(true)
	doneB := make(chan error, 1)
	go func() { doneB <- NewSender(stubB).Send(context.Background()) }()
	<-arrived

	// Flush A runs to completion: the item is sent and removed.
	stubA := newStubRunner(true)
	if err := NewSender(stubA).Send(context.Background()); err != nil {
		t.Fatalf("flush A: %v", err)
	}
	if _, createsA, _ := stubA.recorded(); createsA != 1 {
		t.Fatalf("flush A creates = %d, want the one send", createsA)
	}
	close(proceed)

	if err := <-doneB; err != nil {
		t.Fatalf("flush B: %v", err)
	}
	if _, createsB, _ := stubB.recorded(); createsB != 0 {
		t.Fatalf("flush B re-sent the item flush A already sent (creates=%d) — the stale snapshot was published twice", createsB)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("the queue is not empty after both flushes: %+v", rest)
	}
}

// TestRecordedSendIsReconciledNotResent pins the completion-state rule: a
// send whose queue cleanup failed left the item in the queue with a sent
// row already recorded — the next flush must NOT publish again (the first
// run created the issue; a resent run would only add an occurrence comment
// for the same report). The sent history is the completion record: the
// surviving item is reconciled (removed), never re-sent.
func TestRecordedSendIsReconciledNotResent(t *testing.T) {
	consentOn(t)
	payload, item := payloadFixture(t)
	seedQueue(t, item)

	// The post-cleanup-failure state: the issue EXISTS remotely, the sent
	// row is in the history, and the item survived in the queue.
	if err := outbox.AppendOutbox(outbox.OutboxRow{
		Outcome:  "sent",
		Reason:   "issue created (queue cleanup failed)",
		Title:    item.Title,
		Body:     item.Body,
		Fingerpr: payload.Fingerprint,
	}); err != nil {
		t.Fatalf("seed sent row: %v", err)
	}

	stub := newStubRunner(true)
	stub.issues = []RemoteIssue{{Number: 12, Title: item.Title, State: "open"}}
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	if _, creates, comments := stub.recorded(); creates != 0 || comments != 0 {
		t.Fatalf("a recorded send was published again: creates=%d comments=%d — the surviving item must be reconciled, not re-sent", creates, comments)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("the surviving item was not reconciled out of the queue: %+v", rest)
	}
}

// TestSendHonorsDeadlineOnResultSaves: the result saves (complete, drop,
// fail) used plain Mutate — a contended queue lock made a 25ms-deadline
// Send return after ~1.22s at the gate. The whole flush honors its
// deadline: the result saves select on the caller's context like every
// other lock wait.
func TestSendHonorsDeadlineOnResultSaves(t *testing.T) {
	consentOn(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	// A LIVE queue-lock holder: the result save contends against it.
	store := outbox.BugreportQueueStore()
	release, err := atomicfile.ClaimSection(context.Background(), store.LockPath(), 0o600, 0, time.Millisecond)
	if err != nil {
		t.Fatalf("claim the live holder lock: %v", err)
	}
	defer func() { _ = release() }()

	stub := newStubRunner(true)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	_ = NewSender(stub).Send(ctx)
	if elapsed := time.Since(start); elapsed > 700*time.Millisecond {
		t.Fatalf("a 200ms-deadline Send ran %s on a contended result save — the deadline was not honored", elapsed)
	}
}

// TestExpiredWindowRecurrenceIsNotSuppressedByAnOldSentRow pins the
// reconcile rule's scope: a past sent record suppresses only within the
// DEC-3 duplicate-suppression window. An expired-window recurrence is a
// NEW report — it publishes (here: an occurrence comment on the family
// issue) instead of disappearing into the reconcile path.
func TestExpiredWindowRecurrenceIsNotSuppressedByAnOldSentRow(t *testing.T) {
	consentOn(t)
	payload, item := payloadFixture(t)
	seedQueue(t, item)

	// A sent row stamped 8 days ago — one day past the 7-day window.
	outbox.SetClockForTest(func() time.Time { return time.Now().Add(-8 * 24 * time.Hour) })
	if err := outbox.AppendOutbox(outbox.OutboxRow{
		Outcome:  "sent",
		Reason:   "issue created",
		Title:    item.Title,
		Body:     item.Body,
		Fingerpr: payload.Fingerprint,
	}); err != nil {
		t.Fatalf("seed sent row: %v", err)
	}
	outbox.SetClockForTest(nil)

	stub := newStubRunner(true)
	stub.issues = []RemoteIssue{{Number: 12, Title: item.Title, State: "open"}}
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, comments := stub.recorded()
	if creates != 0 || comments != 1 {
		t.Fatalf("an expired-window recurrence published nothing: creates=%d comments=%d — the old sent row suppressed a new report past the window", creates, comments)
	}
	if rest := queuedItems(t); len(rest) != 0 {
		t.Fatalf("the recurrence did not process out of the queue: %+v", rest)
	}
}

// ---- AC-003: the off/tracked-only arms at the sender ----

func TestSenderNoopWhenParticipationOff(t *testing.T) {
	consentOff(t)
	_, item := payloadFixture(t)
	seedQueue(t, item)

	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	searches, creates, comments := stub.recorded()
	if searches != 0 || creates != 0 || comments != 0 {
		t.Fatalf("participation off: searches=%d creates=%d comments=%d, want zero gh calls", searches, creates, comments)
	}
	if rest := queuedItems(t); len(rest) != 1 {
		t.Fatalf("the item did not stay queued: %+v", rest)
	}
}

func TestSenderIgnoresTrackedFileConsentAndRepository(t *testing.T) {
	// (a) A cloned repository ships participation: true and
	// repository: attacker/x in its tracked section file; the user's own
	// consent file is ABSENT. Nothing may run.
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	project := t.TempDir()
	sections := filepath.Join(project, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	hostile := "feedback:\n  repository: attacker/x\n  participation: true\n  participation_asked: true\n"
	if err := os.WriteFile(filepath.Join(sections, "feedback.yaml"), []byte(hostile), 0o644); err != nil {
		t.Fatalf("write hostile file: %v", err)
	}
	// The user tier section file carries the hostile repository too: the
	// reader resolves participation ONLY from participation.yaml.
	userSections := filepath.Join(home, "config", "sections")
	if err := os.MkdirAll(userSections, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(userSections, "feedback.yaml"), []byte(hostile), 0o600); err != nil {
		t.Fatalf("write hostile user file: %v", err)
	}
	t.Chdir(project)

	_, item := payloadFixture(t)
	seedQueue(t, item)
	stub := newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	searches, creates, comments := stub.recorded()
	if searches != 0 || creates != 0 || comments != 0 {
		t.Fatalf("tracked-only consent: searches=%d creates=%d comments=%d, want zero", searches, creates, comments)
	}

	// (b) The user consents via participation.yaml (no repository key);
	// the hostile repository keys stay where they are. The target must be
	// the compiled default, never attacker/x.
	consentOn(t)
	stub = newStubRunner(true)
	if err := NewSender(stub).Send(context.Background()); err != nil {
		t.Fatalf("send: %v", err)
	}
	_, creates, _ = stub.recorded()
	if creates != 1 {
		t.Fatalf("creates = %d, want the consenting item sent", creates)
	}
	for _, repo := range stub.repos {
		if repo != config.DefaultFeedbackRepository {
			t.Fatalf("gh targeted %q, want only %q (a cloned repository must not redirect the user's account)", repo, config.DefaultFeedbackRepository)
		}
	}
}
