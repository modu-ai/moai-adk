package e2e

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
)

// Per-Issue context ledger regression on lane-20's ctxledger. The remote is a
// counting fake: no real GitHub write exists in these tests.

type ctxRemote struct {
	mu       sync.Mutex
	comments []ctxledger.Comment
	nextID   int64
	Finds    int
	Creates  int
	Updates  int

	// loseNextCreate applies the create and then reports a transport failure.
	loseNextCreate bool

	// updateErr, when set, is returned by every Update without applying it.
	updateErr error
}

// statusErr is a typed HTTP failure: the caller can read the status code
// instead of parsing the message.
type statusErr struct {
	Status int
	Msg    string
}

func (e *statusErr) Error() string { return fmt.Sprintf("HTTP %d: %s", e.Status, e.Msg) }

func (r *ctxRemote) Find(_ context.Context, _ int, marker string) (ctxledger.Comment, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Finds++
	for _, c := range r.comments {
		if strings.Contains(c.Body, marker) {
			return c, true, nil
		}
	}
	return ctxledger.Comment{}, false, nil
}

func (r *ctxRemote) Create(_ context.Context, _ int, body string) (ctxledger.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Creates++
	r.nextID++
	c := ctxledger.Comment{ID: r.nextID, Body: body}
	r.comments = append(r.comments, c)
	if r.loseNextCreate {
		r.loseNextCreate = false
		return ctxledger.Comment{}, errors.New("response lost after the comment was created")
	}
	return c, nil
}

func (r *ctxRemote) Update(_ context.Context, _ int, id int64, body string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Updates++
	if r.updateErr != nil {
		return r.updateErr
	}
	for i := range r.comments {
		if r.comments[i].ID == id {
			r.comments[i].Body = body
			return nil
		}
	}
	// An unknown comment id: the documented typed error of the Remote contract,
	// not a message the caller has to read.
	return fmt.Errorf("%w: HTTP 404", ctxledger.ErrGone)
}

func (r *ctxRemote) calls() int { return r.Finds + r.Creates + r.Updates }

func newCtxLedger(r ctxledger.Remote, st effect.IntentStore) *ctxledger.Ledger {
	return &ctxledger.Ledger{Remote: r, Store: st, Timeouts: effect.Timeouts{LedgerComment: time.Second}}
}

func ctxBody(text string) ctxledger.Body {
	return ctxledger.Body{Issue: 1948, Text: text, MaskPassed: true}
}

// The same update sent twice has one comment effect.
func TestCtxLedger_SameUpdateTwice_OneEffect(t *testing.T) {
	r := &ctxRemote{}
	l := newCtxLedger(r, effect.NewMemIntentStore())
	ctx := context.Background()
	for range 2 {
		if _, err := l.Submit(ctx, ctxBody("audit: PASS head aaaa")); err != nil {
			t.Fatal(err)
		}
		if err := l.Flush(ctx); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("MEASURED same-update-twice finds=%d creates=%d updates=%d comments=%d", r.Finds, r.Creates, r.Updates, len(r.comments))
	if r.Creates != 1 || r.Updates != 0 || len(r.comments) != 1 {
		t.Errorf("creates=%d updates=%d comments=%d, want 1/0/1", r.Creates, r.Updates, len(r.comments))
	}
}

// The intake ACK does not wait for the upload: Submit never touches the remote.
func TestCtxLedger_Submit_NeverTouchesRemote(t *testing.T) {
	r := &ctxRemote{}
	l := newCtxLedger(r, effect.NewMemIntentStore())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = l.Submit(context.Background(), ctxBody("queued for upload"))
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Submit did not return promptly")
	}
	if r.calls() != 0 {
		t.Errorf("remote calls during Submit = %d, want 0", r.calls())
	}
}

// A later update edits the one comment in place; nothing is deleted, and the
// remote interface has no delete capability to begin with.
func TestCtxLedger_NewerText_UpdatesInPlace_NoDeleteCapability(t *testing.T) {
	r := &ctxRemote{}
	st := effect.NewMemIntentStore()
	l := newCtxLedger(r, st)
	ctx := context.Background()
	for _, text := range []string{"entry audit head=aaaa state=current", "entry audit head=bbbb; previous head=aaaa now stale"} {
		if _, err := l.Submit(ctx, ctxBody(text)); err != nil {
			t.Fatal(err)
		}
		if err := l.Flush(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if r.Creates != 1 || r.Updates != 1 || len(r.comments) != 1 {
		t.Errorf("creates=%d updates=%d comments=%d, want 1/1/1", r.Creates, r.Updates, len(r.comments))
	}
	if !strings.Contains(r.comments[0].Body, "head=bbbb") {
		t.Errorf("comment not brought to the newest text: %q", r.comments[0].Body)
	}
	rt := reflect.TypeOf((*ctxledger.Remote)(nil)).Elem()
	for i := range rt.NumMethod() {
		if strings.Contains(strings.ToLower(rt.Method(i).Name), "delete") {
			t.Errorf("Remote exposes %s; the ledger must have no deletion path", rt.Method(i).Name)
		}
	}
}

// Of two pending texts for one Issue only the newest is sent.
func TestCtxLedger_SupersededPending_NeverSent(t *testing.T) {
	r := &ctxRemote{}
	l := newCtxLedger(r, effect.NewMemIntentStore())
	ctx := context.Background()
	for _, text := range []string{"v1 pending", "v2 pending"} {
		if _, err := l.Submit(ctx, ctxBody(text)); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if r.Creates != 1 || strings.Contains(r.comments[0].Body, "v1 pending") {
		t.Errorf("creates=%d body=%q; the superseded text must not be sent", r.Creates, r.comments[0].Body)
	}
}

// Secrets never get recorded or sent: an unmasked body and a body that still
// holds a secret are both refused before anything is stored.
func TestCtxLedger_Unmasked_And_SecretBodies_PublishNothing(t *testing.T) {
	secret := "gh" + "p_" + strings.Repeat("b", 36)
	r := &ctxRemote{}
	st := effect.NewMemIntentStore()
	l := newCtxLedger(r, st)
	ctx := context.Background()

	if _, err := l.Submit(ctx, ctxledger.Body{Issue: 1948, Text: "plain", MaskPassed: false}); !errors.Is(err, ctxledger.ErrNotMasked) {
		t.Errorf("unmasked body: %v, want ErrNotMasked", err)
	}
	if _, err := l.Submit(ctx, ctxBody("leaked "+secret)); !errors.Is(err, ctxledger.ErrNotMasked) {
		t.Errorf("secret body: %v, want ErrNotMasked", err)
	}
	if err := l.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if all, _ := st.List(); len(all) != 0 {
		t.Errorf("stored intents = %d, want 0", len(all))
	}
	if r.calls() != 0 {
		t.Errorf("remote calls = %d, want 0", r.calls())
	}
}

// A lost response is reconciled by the marker: the comment exists once and a
// second flush does not create another.
func TestCtxLedger_LostResponse_ReconciledByMarker(t *testing.T) {
	r := &ctxRemote{loseNextCreate: true}
	l := newCtxLedger(r, effect.NewMemIntentStore())
	ctx := context.Background()
	if _, err := l.Submit(ctx, ctxBody("evidence summary")); err != nil {
		t.Fatal(err)
	}
	if err := l.Flush(ctx); err != nil {
		t.Fatalf("flush after a lost response: %v", err)
	}
	if err := l.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	t.Logf("MEASURED lost-response finds=%d creates=%d updates=%d comments=%d", r.Finds, r.Creates, r.Updates, len(r.comments))
	if r.Creates != 1 || len(r.comments) != 1 {
		t.Errorf("creates=%d comments=%d, want 1/1", r.Creates, len(r.comments))
	}
}

// Restart: a new Ledger over the same durable store sends what the old one had
// only recorded, exactly once.
func TestCtxLedger_Restart_SendsRecordedIntentOnce(t *testing.T) {
	st := effect.NewMemIntentStore()
	first := newCtxLedger(&ctxRemote{}, st)
	if _, err := first.Submit(context.Background(), ctxBody("recorded before the restart")); err != nil {
		t.Fatal(err)
	}
	r := &ctxRemote{}
	second := newCtxLedger(r, st)
	for range 2 {
		if err := second.Flush(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if r.Creates != 1 || len(r.comments) != 1 {
		t.Errorf("creates=%d comments=%d, want 1/1", r.Creates, len(r.comments))
	}
}
