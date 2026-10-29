package ctxledger

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// fakeRemote is the only "GitHub" the tests see: an in-memory comment list with write counters.
type fakeRemote struct {
	comments         map[int][]Comment
	nextID           int64
	creates, updates int
	lostAfterCreate  bool // create lands, response is lost
	lostAfterUpdate  bool
	findErr          error
	finds            int
	plain404         bool           // Update reports a missing comment as a plain error, not ErrGone
	gone             map[int64]bool // ids that no longer exist
}

func newRemote() *fakeRemote { return &fakeRemote{comments: map[int][]Comment{}, nextID: 100} }

func (f *fakeRemote) Find(_ context.Context, issue int, marker string) (Comment, bool, error) {
	f.finds++
	if f.findErr != nil {
		return Comment{}, false, f.findErr
	}
	for _, c := range f.comments[issue] {
		if strings.Contains(c.Body, marker) {
			return c, true, nil
		}
	}
	return Comment{}, false, nil
}

func (f *fakeRemote) Create(_ context.Context, issue int, body string) (Comment, error) {
	f.creates++
	f.nextID++
	c := Comment{ID: f.nextID, Body: body}
	f.comments[issue] = append(f.comments[issue], c)
	if f.lostAfterCreate {
		return Comment{}, errors.New("connection reset")
	}
	return c, nil
}

func (f *fakeRemote) Update(_ context.Context, issue int, id int64, body string) error {
	f.updates++
	if f.gone[id] {
		return ErrGone
	}
	if f.plain404 {
		found := false
		for _, c := range f.comments[issue] {
			if c.ID == id {
				found = true
			}
		}
		if !found {
			return errors.New("HTTP 404: comment not found")
		}
	}
	if f.lostAfterUpdate {
		f.lostAfterUpdate = false
		return errors.New("timeout")
	}
	for i, c := range f.comments[issue] {
		if c.ID == id {
			f.comments[issue][i].Body = body
		}
	}
	return nil
}

func newLedger(r Remote, max int) (*Ledger, *effect.MemIntentStore) {
	st := effect.NewMemIntentStore()
	return &Ledger{Remote: r, Store: st, MaxBytes: max, Timeouts: effect.Timeouts{LedgerComment: time.Second}}, st
}

func body(text string) Body { return Body{Issue: 7, Text: text, MaskPassed: true} }

func TestUnmaskedBodyIsNeverRecordedOrSent(t *testing.T) {
	r := newRemote()
	l, st := newLedger(r, 4096)
	if _, err := l.Submit(t.Context(), Body{Issue: 7, Text: "plain", MaskPassed: false}); !errors.Is(err, ErrNotMasked) {
		t.Fatalf("err = %v", err)
	}
	if all, _ := st.List(); len(all) != 0 || r.creates+r.updates != 0 {
		t.Errorf("intents=%d writes=%d", len(all), r.creates+r.updates)
	}
}

func TestSubmitOnlyRecordsAndNeverWaitsForUpload(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	in, err := l.Submit(t.Context(), body("context v1"))
	if err != nil || in.Status != effect.StatusIntended || r.creates != 0 || len(r.comments[7]) != 0 {
		t.Fatalf("in=%+v err=%v creates=%d", in, err, r.creates)
	}
}

func TestOneCommentPerIssueUpdatedInPlaceAndSameContentNeverResent(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	for _, text := range []string{"v1", "v1", "v2", "v2", "v3"} {
		if _, err := l.Submit(t.Context(), body(text)); err != nil {
			t.Fatal(err)
		}
		if err := l.Flush(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	if len(r.comments[7]) != 1 {
		t.Fatalf("comments = %d, want exactly 1", len(r.comments[7]))
	}
	if r.creates != 1 || r.updates != 2 {
		t.Errorf("creates=%d updates=%d, want 1 and 2 (v1 repeat and v2 repeat send nothing)", r.creates, r.updates)
	}
	if !strings.Contains(r.comments[7][0].Body, "v3") || !strings.Contains(r.comments[7][0].Body, Marker(7)) {
		t.Errorf("body = %q", r.comments[7][0].Body)
	}
}

func TestLostCreateResponseIsReconciledByMarker(t *testing.T) {
	r := newRemote()
	r.lostAfterCreate = true
	l, st := newLedger(r, 4096)
	_, _ = l.Submit(t.Context(), body("v1"))
	_ = l.Flush(t.Context())
	_ = l.Flush(t.Context())
	if r.creates != 1 || len(r.comments[7]) != 1 {
		t.Errorf("creates=%d comments=%d", r.creates, len(r.comments[7]))
	}
	for _, in := range mustList(st) {
		if in.Status != effect.StatusConfirmed {
			t.Errorf("status = %s", in.Status)
		}
	}
}

func TestLostUpdateResponseRetriesUpdateNotCreate(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	_, _ = l.Submit(t.Context(), body("v1"))
	_ = l.Flush(t.Context())
	r.lostAfterUpdate = true
	_, _ = l.Submit(t.Context(), body("v2"))
	_ = l.Flush(t.Context())
	_ = l.Flush(t.Context())
	if r.creates != 1 || len(r.comments[7]) != 1 || !strings.Contains(r.comments[7][0].Body, "v2") {
		t.Errorf("creates=%d comments=%v", r.creates, r.comments[7])
	}
}

func TestUnsureLookupParksAndWritesNothing(t *testing.T) {
	r := newRemote()
	r.findErr = errors.New("gh 502")
	l, st := newLedger(r, 4096)
	_, _ = l.Submit(t.Context(), body("v1"))
	_ = l.Flush(t.Context())
	if r.creates+r.updates != 0 {
		t.Errorf("wrote %d times while the lookup was failing", r.creates+r.updates)
	}
	for _, in := range mustList(st) {
		if in.Status == effect.StatusConfirmed {
			t.Error("confirmed without evidence")
		}
	}
}

func TestOversizeBodyIsSummarised(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 300)
	_, _ = l.Submit(t.Context(), body(strings.Repeat("line of context\n", 200)))
	_ = l.Flush(t.Context())
	got := r.comments[7][0].Body
	if len(got) > 300 || !strings.Contains(got, "summary") {
		t.Errorf("len=%d body=%q", len(got), got)
	}
}

func TestSecretSlippingPastTheFlagIsRefusedBeforeSend(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	_, _ = l.Submit(t.Context(), body("token ghp_abcdefghijklmnopqrstuvwxyz0123456789"))
	_ = l.Flush(t.Context())
	if r.creates+r.updates != 0 {
		t.Error("a body that still holds a secret was sent")
	}
}

func mustList(st *effect.MemIntentStore) []effect.Intent {
	l, _ := st.List()
	return l
}

// GHRemote issues only comment create/update/list calls, all through the injected runner.
type fakeRun struct{ calls [][]string }

func (f *fakeRun) run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	if args[0] == "api" && len(args) > 2 && args[1] == "-X" {
		return []byte(`{"id":501,"body":"x"}`), nil
	}
	return []byte(`[]`), nil
}

func TestGHRemoteOnlyIssuesCommentCalls(t *testing.T) {
	fr := &fakeRun{}
	g := GHRemote{Repo: "o/r", Run: fr.run}
	if _, err := g.Create(t.Context(), 7, "b"); err != nil {
		t.Fatal(err)
	}
	if err := g.Update(t.Context(), 7, 501, "b"); err != nil {
		t.Fatal(err)
	}
	_, _, _ = g.Find(t.Context(), 7, "m")
	for _, c := range fr.calls {
		joined := strings.Join(c, " ")
		if !strings.Contains(joined, "/issues/7/comments") && !strings.Contains(joined, "/issues/comments/501") {
			t.Errorf("unexpected gh call %q", joined)
		}
		if strings.Contains(joined, "issue close") || strings.Contains(joined, "pr merge") {
			t.Errorf("forbidden call %q", joined)
		}
	}
	if err := (GHRemote{Repo: "o/r", Run: fr.run}).Update(t.Context(), 7, 0, "b"); err == nil {
		t.Error("update without a comment id accepted")
	}
}

func TestRevertToOlderTextIsSentAgain(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	for _, text := range []string{"v1", "v2", "v1"} {
		if _, err := l.Submit(t.Context(), body(text)); err != nil {
			t.Fatal(err)
		}
		if err := l.Flush(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	if len(r.comments[7]) != 1 || !strings.HasSuffix(r.comments[7][0].Body, "v1") {
		t.Errorf("remote holds %v, want the latest submitted text v1", r.comments[7])
	}
}

func submitFlush(t *testing.T, l *Ledger, text string) {
	t.Helper()
	if _, err := l.Submit(t.Context(), body(text)); err != nil {
		t.Fatal(err)
	}
	if err := l.Flush(t.Context()); err != nil {
		t.Fatal(err)
	}
}

// S52 M1: after the first write the comment id and body hash are kept, so later writes skip the list GET.
func TestCachedIDSkipsTheListLookup(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	for _, text := range []string{"v1", "v2", "v3", "v4"} {
		submitFlush(t, l, text)
	}
	if r.finds != 1 {
		t.Errorf("list GETs = %d, want 1 (only the first write needs the marker lookup)", r.finds)
	}
	if r.creates != 1 || r.updates != 3 || len(r.comments[7]) != 1 || !strings.HasSuffix(r.comments[7][0].Body, "v4") {
		t.Errorf("creates=%d updates=%d comments=%v", r.creates, r.updates, r.comments[7])
	}
}

func TestSameBodyAsCachedIsNotSentAtAll(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	submitFlush(t, l, "v1")
	in := mustList(l.Store.(*effect.MemIntentStore))[0]
	before := r.finds + r.creates + r.updates
	id, err := (sender{r, l}).Send(t.Context(), in)
	if err != nil || id == "" {
		t.Fatalf("id=%q err=%v", id, err)
	}
	if after := r.finds + r.creates + r.updates; after != before {
		t.Errorf("remote calls %d -> %d for an unchanged body", before, after)
	}
}

func TestCommentGoneFallsBackToMarkerLookupAndStaysOneComment(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	submitFlush(t, l, "v1")
	gone := r.comments[7][0].ID
	r.gone = map[int64]bool{gone: true}
	r.comments[7] = nil // someone deleted it
	submitFlush(t, l, "v2")
	if len(r.comments[7]) != 1 || !strings.HasSuffix(r.comments[7][0].Body, "v2") || r.creates != 2 {
		t.Errorf("comments=%v creates=%d", r.comments[7], r.creates)
	}
}

func TestFreshLedgerWithoutCacheStillFindsByMarker(t *testing.T) {
	r := newRemote()
	l1, st := newLedger(r, 4096)
	submitFlush(t, l1, "v1")
	l2 := &Ledger{Remote: r, Store: st, MaxBytes: 4096, Timeouts: effect.Timeouts{LedgerComment: time.Second}} // restart: no cache
	submitFlush(t, l2, "v2")
	if len(r.comments[7]) != 1 || r.creates != 1 {
		t.Errorf("restart made a second comment: %v creates=%d", r.comments[7], r.creates)
	}
}

func TestAmbiguousUpdateDropsTheCacheSoReconcileReadsTheRemote(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	submitFlush(t, l, "v1")
	r.lostAfterUpdate = true
	_, _ = l.Submit(t.Context(), body("v2"))
	_ = l.Flush(t.Context()) // update lost: reported, not silent
	submitFlush(t, l, "v3")
	if len(r.comments[7]) != 1 || !strings.HasSuffix(r.comments[7][0].Body, "v3") {
		t.Errorf("comments = %v", r.comments[7])
	}
}

// A person deletes the comment; the remembered id now answers 404 in whatever words the host uses.
// The first Flush must leave one comment at the newest text, or return an error — never look fine
// with the comment missing.
func TestDeletedCommentIsRecreatedInTheSameFlushEvenWithoutErrGone(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	submitFlush(t, l, "first")
	r.comments[7] = nil
	r.plain404 = true // Update answers errors.New("HTTP 404: comment not found"), not ErrGone
	if _, err := l.Submit(t.Context(), body("second")); err != nil {
		t.Fatal(err)
	}
	err := l.Flush(t.Context())
	if err == nil && (len(r.comments[7]) != 1 || !strings.HasSuffix(r.comments[7][0].Body, "second")) {
		t.Errorf("Flush returned nil but comments = %v", r.comments[7])
	}
}

func TestFlushReportsAnIntentThatDidNotLand(t *testing.T) {
	r := newRemote()
	r.findErr = errors.New("gh 502")
	l, _ := newLedger(r, 4096)
	_, _ = l.Submit(t.Context(), body("v1"))
	if err := l.Flush(t.Context()); err == nil {
		t.Error("Flush returned nil although the comment did not land (parked)")
	}
}

func TestEmailAndHomePathNeverPublishEvenWithTheFlag(t *testing.T) {
	email := "person" + "@" + "example.org"
	for _, text := range []string{
		"owner " + email,
		"worktree /Us" + "ers/someone/project",
		"worktree /ho" + "me/someone/project",
		`worktree C:\Us` + `ers\someone\project`,
	} {
		r := newRemote()
		l, st := newLedger(r, 4096)
		if _, err := l.Submit(t.Context(), body(text)); !errors.Is(err, ErrNotMasked) {
			t.Errorf("%q: submit err = %v", text, err)
		}
		// Even if an intent reached the store by another path, the sender's last check refuses it.
		in := effect.Intent{Key: "k", Kind: effect.LedgerComment, Issue: 7, Payload: Marker(7) + "\n" + text, Status: effect.StatusIntended, Seq: 1}
		_ = st.Put(in)
		_ = l.Flush(t.Context())
		if r.creates+r.updates != 0 {
			t.Errorf("%q: published", text)
		}
	}
}

func TestRelativeEvidencePathsStillPass(t *testing.T) {
	r := newRemote()
	l, _ := newLedger(r, 4096)
	submitFlush(t, l, "see .moai/evidence/issue-7/audit.md and internal/agenticfactory/effect")
	if len(r.comments[7]) != 1 {
		t.Errorf("relative path text was refused: %v", r.comments[7])
	}
}
