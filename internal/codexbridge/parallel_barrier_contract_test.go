package codexbridge

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

func TestParallelToolBarrierRequiresEveryDistinctResult(t *testing.T) {
	e, rpc, store := fixture(t)
	rpc.transform = func(first codexapp.Message) codexapp.Message {
		rpc.events <- first
		return codexapp.Message{ID: json.RawMessage(`"second-rpc"`), Method: "item/tool/call", Params: json.RawMessage(`{"threadId":"thread-1","turnId":"turn-thread-1","callId":"second-call","tool":"echo","arguments":{"text":"second"}}`)}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("parallel-barrier")
	seg, err := e.Step(ctx, q)
	if err != nil || len(seg.Tools) != 2 {
		t.Fatalf("parallel tools=%+v err=%v", seg.Tools, err)
	}
	if seg.Tools[0].ID == seg.Tools[1].ID || string(seg.Tools[1].Arguments) != `{"text":"second"}` {
		t.Fatalf("tool identities/arguments lost: %+v", seg.Tools)
	}
	barrier, found, err := store.Barrier(q.Owner)
	if err != nil || !found || barrier.Phase != "waiting" {
		t.Fatalf("barrier=%+v found=%v err=%v", barrier, found, err)
	}
	q.Input = nil
	q.ExpectedPrefix = q.PrefixDigest
	q.Results = []ToolResult{{ID: seg.Tools[0].ID, Success: true, Content: []Content{{Type: "inputText", Text: "first result"}}}}
	if _, err = e.Step(ctx, q); !errors.Is(err, ErrScope) {
		t.Fatalf("partial batch accepted: %v", err)
	}
	q.Results = append(q.Results, q.Results[0])
	if _, err = e.Step(ctx, q); !errors.Is(err, ErrScope) {
		t.Fatalf("duplicate result accepted: %v", err)
	}
	if rpc.responses != 0 {
		t.Fatal("invalid batch crossed tool response boundary")
	}
}
