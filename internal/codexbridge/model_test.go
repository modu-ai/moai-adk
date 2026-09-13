package codexbridge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

// An idle conversation may switch models: the transition repins the barrier
// record and the next turn/start carries the new model upstream. Waiting
// transitions stay rejected so a switch can never land on the wrong turn; a
// conversation is pinned to the model of its creating request before that.
func TestIdleModelChangeRepinsAndCarriesNewModel(t *testing.T) {
	e, rpc, store := fixture(t)
	completeConversation(t, e, "switch")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("switch")
	q.Model = "gpt-6-astra"
	q.ExpectedPrefix = "prefix-2"
	q.PrefixDigest = "prefix-3"
	q.Input = []any{map[string]string{"type": "text", "text": "continue"}}
	first, err := e.Step(ctx, q)
	if err != nil || first.Tool == nil {
		t.Fatal("idle model change rejected", err)
	}
	rpc.mu.Lock()
	got := append([]string(nil), rpc.models...)
	rpc.mu.Unlock()
	if len(got) != 2 || got[0] != "gpt-5.6-sol" || got[1] != "gpt-6-astra" {
		t.Fatal("turn/start model sequence", got)
	}
	q.Input = nil
	q.ExpectedPrefix = "prefix-3"
	q.PrefixDigest = "prefix-4"
	q.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "inputText", Text: "ok"}}, Success: true}}
	if last, err := e.Step(ctx, q); err != nil || !last.Done {
		t.Fatal(last, err)
	}
	e.Close()

	// The switched model is now the pinned one: a fresh process resumes with it,
	// while the pre-switch model is rejected by the pinning contrast.
	freshRPC := &fakeRPC{events: make(chan codexapp.Message, 8)}
	fresh, err := New(context.Background(), freshRPC, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	rq := resumeRequest("switch")
	rq.Model = "gpt-6-astra"
	rq.ExpectedPrefix = "prefix-4"
	rq.PrefixDigest = "prefix-5"
	seg, err := fresh.Step(ctx, rq)
	if err != nil || seg.Tool == nil {
		t.Fatal("switched model resume rejected", err)
	}
	if len(freshRPC.resumes) != 1 || freshRPC.next != 0 {
		t.Fatal("resume did not reuse the owned thread", freshRPC.resumes, freshRPC.next)
	}
	rq.Input = nil
	rq.ExpectedPrefix = "prefix-5"
	rq.PrefixDigest = "prefix-6"
	rq.Results = []ToolResult{{ID: seg.Tool.ID, Content: []Content{{Type: "inputText", Text: "ok"}}, Success: true}}
	if last, err := fresh.Step(ctx, rq); err != nil || !last.Done {
		t.Fatal(last, err)
	}
	fresh.Close()

	staleRPC := &fakeRPC{events: make(chan codexapp.Message, 8)}
	stale, err := New(context.Background(), staleRPC, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer stale.Close()
	old := resumeRequest("switch")
	old.ExpectedPrefix = "prefix-6"
	old.PrefixDigest = "prefix-7"
	if _, err = stale.Step(ctx, old); !errors.Is(err, ErrScope) {
		t.Fatal("pre-switch model accepted after the pin moved", err)
	}
	if len(staleRPC.resumes) != 0 || staleRPC.next != 0 {
		t.Fatal("rejected resume reached the App Server")
	}
}

func TestWaitingAndNewPhaseModelChangeStayRejected(t *testing.T) {
	t.Run("waiting", func(t *testing.T) {
		e, rpc, _ := fixture(t)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		q := request("wait")
		first, err := e.Step(ctx, q)
		if err != nil || first.Tool == nil {
			t.Fatal(err)
		}
		q.Model = "gpt-6-astra"
		q.Input = nil
		q.ExpectedPrefix = "prefix-1"
		q.PrefixDigest = "prefix-2"
		q.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "inputText", Text: "ok"}}, Success: true}}
		if _, err = e.Step(ctx, q); !errors.Is(err, ErrScope) {
			t.Fatal("waiting model change accepted", err)
		}
		if rpc.responses != 0 {
			t.Fatal("switched model reached the pending turn RPC")
		}
	})
}
