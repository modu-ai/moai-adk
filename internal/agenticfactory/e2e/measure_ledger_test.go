package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mirror"
)

// Measurement harness: a counting fake of `gh` so ledger-write cost and safety
// are observed without any real GitHub write. Counts are what the code under
// test actually issued in this run; they are not extrapolated.

type fakeGH struct {
	mu       sync.Mutex
	comments []fakeComment
	nextID   int64
	Gets     int
	Posts    int
	Patches  int
	Bodies   []string // every body written, in order

	// failAfterApply makes the next write apply its effect and then report a
	// transport failure, the lost-response case.
	failAfterApply bool
}

type fakeComment struct {
	ID   int64
	Body string
}

func (f *fakeGH) Run(_ context.Context, args ...string) ([]byte, []byte, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	joined := strings.Join(args, " ")
	body := ""
	for _, a := range args {
		if b, ok := strings.CutPrefix(a, "body="); ok {
			body = b
		}
	}
	switch {
	case strings.Contains(joined, "-X POST"):
		f.Posts++
		f.nextID++
		f.comments = append(f.comments, fakeComment{ID: f.nextID, Body: body})
		f.Bodies = append(f.Bodies, body)
	case strings.Contains(joined, "-X PATCH"):
		f.Patches++
		f.Bodies = append(f.Bodies, body)
		path := ""
		for _, a := range args {
			if strings.HasPrefix(a, "repos/") {
				path = a // the endpoint, wherever the flags put it
			}
		}
		for i := range f.comments {
			if strings.HasSuffix(path, "/"+fmt.Sprint(f.comments[i].ID)) {
				f.comments[i].Body = body
			}
		}
	default:
		f.Gets++
		var out strings.Builder
		for _, c := range f.comments {
			line, _ := json.Marshal(map[string]any{"id": c.ID, "body": c.Body})
			out.Write(line)
			out.WriteByte('\n')
		}
		return []byte(out.String()), nil, 0, nil
	}
	if f.failAfterApply {
		f.failAfterApply = false
		return nil, []byte("HTTP 502 bad gateway (response lost)"), 1, nil
	}
	return nil, nil, 0, nil
}

func (f *fakeGH) calls() int { return f.Gets + f.Posts + f.Patches }

func ledgerAt(seq int64, state core.State) core.Ledger {
	ref, _ := core.NewIssueRef(1948)
	return core.Ledger{Issue: ref, State: state, StoreSeq: seq, Head: core.HeadSHA(headA)}
}

func newMirror(f *fakeGH) *mirror.Mirror {
	return &mirror.Mirror{Repo: "o/r", Exec: f, AllowWrite: true}
}

// Measured: gh calls per ledger update, and the duplicate read each update does.
func TestMeasure_LedgerWrite_CallsPerUpdate(t *testing.T) {
	f := &fakeGH{}
	m := newMirror(f)
	states := []core.State{core.StateClaimed, core.StateImplementing, core.StateVerifying, core.StateReady, core.StateLanding}
	for i, s := range states {
		if err := m.WriteLedger(context.Background(), ledgerAt(int64(i+1), s)); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("MEASURED updates=%d gets=%d posts=%d patches=%d total=%d", len(states), f.Gets, f.Posts, f.Patches, f.calls())
	if len(f.comments) != 1 {
		t.Fatalf("comment count = %d, want exactly 1 ledger comment", len(f.comments))
	}
}

// The state-ledger comment is deleted on GitHub between two updates. The first
// update after the deletion must leave exactly one comment (or report an error).
func TestMeasure_LedgerWrite_DeletedRemotely_Observed(t *testing.T) {
	f := &fakeGH{}
	m := newMirror(f)
	ctx := context.Background()
	if err := m.WriteLedger(ctx, ledgerAt(1, core.StateClaimed)); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	f.comments = nil // deleted by a person
	f.mu.Unlock()
	err := m.WriteLedger(ctx, ledgerAt(2, core.StateImplementing))
	t.Logf("MEASURED deleted-remotely first update: err=%v comments=%d gets=%d posts=%d patches=%d",
		err, len(f.comments), f.Gets, f.Posts, f.Patches)
	if err != nil || len(f.comments) != 1 {
		t.Errorf("after a remote deletion the first update must leave exactly one comment or report an error: err=%v comments=%d", err, len(f.comments))
	}
}

// A person edits the ledger comment, then the SAME ledger is written again. The
// comment must be brought back to the authoritative content: a shortcut that
// skips an unchanged-content write must not leave the person's edit in place.
func TestMeasure_LedgerWrite_ExternalEdit_Observed(t *testing.T) {
	f := &fakeGH{}
	m := newMirror(f)
	ctx := context.Background()
	l := ledgerAt(1, core.StateClaimed)
	if err := m.WriteLedger(ctx, l); err != nil {
		t.Fatal(err)
	}
	authoritative := f.comments[0].Body
	f.mu.Lock()
	f.comments[0].Body = mirror.Marker + "\n```json\n{\"state\":\"done\"}\n```\n" // edited by a person
	f.mu.Unlock()
	callsBefore := f.calls()
	err := m.WriteLedger(ctx, l)
	t.Logf("MEASURED external edit then same ledger: err=%v restored=%v extra-calls=%d", err, f.comments[0].Body == authoritative, f.calls()-callsBefore)
	if err != nil || f.comments[0].Body != authoritative {
		t.Errorf("a person's edit survived a rewrite of the same ledger: err=%v", err)
	}
}

// Safety: the same update twice never creates a second comment. The cost of the
// repeat (gh calls spent on an unchanged ledger) is measured, not asserted.
func TestMeasure_LedgerWrite_SameUpdateTwice(t *testing.T) {
	f := &fakeGH{}
	m := newMirror(f)
	l := ledgerAt(1, core.StateClaimed)
	for range 2 {
		if err := m.WriteLedger(context.Background(), l); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("MEASURED same-update-twice gets=%d posts=%d patches=%d (body identical across writes: %v)",
		f.Gets, f.Posts, f.Patches, len(f.Bodies) == 2 && f.Bodies[0] == f.Bodies[1])
	if len(f.comments) != 1 {
		t.Errorf("comment count = %d, want 1", len(f.comments))
	}
}

// Safety: a write whose response was lost is reconciled by the marker on retry,
// never duplicated.
func TestMeasure_LedgerWrite_LostResponse_ReconciledByMarker(t *testing.T) {
	f := &fakeGH{failAfterApply: true}
	m := newMirror(f)
	l := ledgerAt(1, core.StateClaimed)
	if err := m.WriteLedger(context.Background(), l); err == nil {
		t.Fatal("injected response loss did not surface as an error")
	}
	if err := m.WriteLedger(context.Background(), l); err != nil {
		t.Fatalf("retry: %v", err)
	}
	t.Logf("MEASURED lost-response retry gets=%d posts=%d patches=%d comments=%d", f.Gets, f.Posts, f.Patches, len(f.comments))
	if len(f.comments) != 1 || f.Posts != 1 {
		t.Errorf("comments=%d posts=%d, want 1 and 1 (marker reconciliation)", len(f.comments), f.Posts)
	}
}

// Safety: secrets in the ledger never reach the comment body. The fixtures are
// assembled at run time and are not real credentials.
func TestMeasure_LedgerWrite_SecretsNeverPublished(t *testing.T) {
	tokenLike := "gh" + "p_" + strings.Repeat("a", 36)
	pwLike := strings.Repeat("h", 12)
	f := &fakeGH{}
	m := newMirror(f)
	l := ledgerAt(1, core.StateBlocked)
	l.NextAction = "retry with token=" + tokenLike + " and pass" + "word: " + pwLike
	// Design 4.4 (lane-18 7196d0b5ea): a ledger that holds a secret is refused with
	// core.ErrUnsanitized and a RESTRICTED record is posted in its place; it used to be masked and posted.
	// The safety claim is unchanged: no published body carries the value.
	err := m.WriteLedger(context.Background(), l)
	if !errors.Is(err, core.ErrUnsanitized) {
		t.Fatalf("WriteLedger = %v, want core.ErrUnsanitized", err)
	}
	if len(f.Bodies) != 1 || !strings.Contains(f.Bodies[0], `"status": "RESTRICTED"`) {
		t.Fatalf("published bodies = %q, want exactly one RESTRICTED record", f.Bodies)
	}
	for _, b := range f.Bodies {
		for _, secret := range []string{tokenLike, pwLike} {
			if strings.Contains(b, secret) {
				t.Errorf("published body contains the secret-like value %.6s…", secret)
			}
		}
	}
}

// Safety: writes are off unless explicitly allowed, and then the fake sees zero calls.
func TestMeasure_LedgerWrite_DisabledMakesNoCalls(t *testing.T) {
	f := &fakeGH{}
	m := &mirror.Mirror{Repo: "o/r", Exec: f}
	if err := m.WriteLedger(context.Background(), ledgerAt(1, core.StateClaimed)); !errors.Is(err, mirror.ErrWriteDisabled) {
		t.Fatalf("err = %v, want ErrWriteDisabled", err)
	}
	if f.calls() != 0 {
		t.Errorf("gh calls with writes disabled = %d, want 0", f.calls())
	}
}
