package codexbridge

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

// newStore and engineOn build the multi-process fork fixture: one durable
// store, several sequential Engine instances over it.
func newStore(t *testing.T) *FileStore {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	store, err := OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func engineOn(t *testing.T, store *FileStore) (*Engine, *fakeRPC) {
	t.Helper()
	rpc := &fakeRPC{events: make(chan codexapp.Message, 32)}
	e, err := New(context.Background(), rpc, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { e.Close() })
	return e, rpc
}

// A fork child is a brand-new conversation that starts a brand-new App Server
// thread while carrying only the inherited boundary prefix — it never adopts
// the parent thread identity and never resumes it.
func TestForkChildStartsNewThreadAtInheritedPrefix(t *testing.T) {
	store := newStore(t)
	parent, _ := engineOn(t, store)
	completeConversation(t, parent, "parent")
	parent.Close()

	fresh, freshRPC := engineOn(t, store)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("child")
	q.Fork = true
	q.ExpectedPrefix = "prefix-2"
	q.PrefixDigest = "prefix-3"
	seg, err := fresh.Step(ctx, q)
	if err != nil || seg.Tool == nil {
		t.Fatal("fork child rejected", err)
	}
	if freshRPC.next != 1 || len(freshRPC.resumes) != 0 {
		t.Fatal("fork child did not start a new thread", freshRPC.next, freshRPC.resumes)
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-3"
	q.PrefixDigest = "prefix-4"
	q.Results = []ToolResult{{ID: seg.Tool.ID, Content: []Content{{Type: "inputText", Text: "ok"}}, Success: true}}
	if last, err := fresh.Step(ctx, q); err != nil || !last.Done {
		t.Fatal(last, err)
	}
	fresh.Close()

	// The child owns its own barrier: a new process resumes it independently.
	third, rpc3 := engineOn(t, store)
	rq := resumeRequest("child")
	rq.ExpectedPrefix = "prefix-4"
	rq.PrefixDigest = "prefix-5"
	if _, err = third.Step(ctx, rq); err != nil {
		t.Fatal("fork-child resume rejected", err)
	}
	if len(rpc3.resumes) != 1 || rpc3.next != 0 {
		t.Fatal("fork-child resume did not reuse its own thread", rpc3.resumes, rpc3.next)
	}
}

func TestForkRequestRejectionGroups(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t.Run("missing-boundary-prefix", func(t *testing.T) {
		e, rpc, _ := fixture(t)
		q := request("orphan")
		q.Fork = true
		if _, err := e.Step(ctx, q); !errors.Is(err, ErrScope) {
			t.Fatal("fork without an inherited prefix accepted", err)
		}
		if rpc.next != 0 {
			t.Fatal("rejected fork started a thread")
		}
	})
	t.Run("fork-and-resume", func(t *testing.T) {
		store := newStore(t)
		e, _ := engineOn(t, store)
		completeConversation(t, e, "done")
		e.Close()
		fresh, freshRPC := engineOn(t, store)
		q := resumeRequest("done")
		q.Fork = true
		if _, err := fresh.Step(ctx, q); !errors.Is(err, ErrScope) {
			t.Fatal("fork combined with resume accepted", err)
		}
		if len(freshRPC.resumes) != 0 || freshRPC.next != 0 {
			t.Fatal("rejected fork reached the App Server")
		}
	})
}
