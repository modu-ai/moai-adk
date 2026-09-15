package codexbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

func TestLongToolWaitRetainsBurstWithoutPoisoningOtherOwners(t *testing.T) {
	e, rpc, _ := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := request("long-tool-wait")
	first, err := e.Step(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 256; i++ {
		for _, method := range []string{"item/reasoning/textDelta", "item/agentMessage/delta", "thread/tokenUsage/updated"} {
			params := map[string]any{"threadId": "thread-1", "turnId": "turn-thread-1", "delta": "x"}
			if method == "thread/tokenUsage/updated" {
				u := map[string]int{"inputTokens": i + 1, "cachedInputTokens": 0, "outputTokens": i + 1}
				params["tokenUsage"] = map[string]any{"total": u, "last": u}
			}
			raw, _ := json.Marshal(params)
			select {
			case rpc.events <- codexapp.Message{Method: method, Params: raw}:
			case <-e.done:
				t.Fatalf("long tool wait poisoned shared reader: %v", e.failure())
			case <-ctx.Done():
				t.Fatal("event reader blocked RPC transport")
			}
		}
	}
	other, err := e.Step(ctx, request("other-owner"))
	if err != nil || other.Tool == nil {
		t.Fatalf("unrelated owner unavailable: %v", err)
	}
	q.Input = nil
	q.ExpectedPrefix = q.PrefixDigest
	q.PrefixDigest = "returned-tool"
	q.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "inputText", Text: "done"}}, Success: true}}
	last, err := e.Step(ctx, q)
	if err != nil || !last.Done || last.Text != strings.Repeat("x", 256)+"answer" {
		t.Fatalf("lost ordered semantic output: %+v %v", last, err)
	}
	if last.Usage == nil || last.Usage.InputTokens != 256 || last.Usage.OutputTokens != 256 {
		t.Fatalf("lost usage updates: %+v", last.Usage)
	}
}

func TestEventMailboxBudgetOrderAndWakeReuse(t *testing.T) {
	q := newEventQueue(64 << 10)
	var admitted []codexapp.Message
	for i := 0; ; i++ {
		event := codexapp.Message{Method: "unknown/future", ID: json.RawMessage(`"request"`), Params: json.RawMessage(`{"sequence":` + fmt.Sprint(i) + `}`)}
		if !q.push(event) {
			break
		}
		admitted = append(admitted, event)
	}
	if len(admitted) == 0 || q.bytes > q.limit {
		t.Fatal("invalid budget", q.bytes, q.limit)
	}
	for i, want := range admitted {
		<-q.ready
		got, ok := q.pop()
		if !ok || !reflect.DeepEqual(got, want) {
			t.Fatalf("event %d changed/reordered: %+v", i, got)
		}
	}
	if q.bytes != 0 || q.events.Len() != 0 {
		t.Fatal("drain retained accounted payload", q.bytes)
	}
	if !q.push(admitted[0]) {
		t.Fatal("drained queue not reusable")
	}
	select {
	case <-q.ready:
	default:
		t.Fatal("reused mailbox lost wake")
	}
	if _, ok := q.pop(); !ok {
		t.Fatal("lost reused event")
	}
	oversized := codexapp.Message{Params: json.RawMessage(strings.Repeat("x", q.limit))}
	if q.push(oversized) {
		t.Fatal("payload-only budget ignored event overhead")
	}
}

func TestInformationalFilterPreservesRequestsAndUnknownEvents(t *testing.T) {
	for _, method := range []string{"item/agentMessage/delta", "item/tool/call", "thread/tokenUsage/updated", "turn/completed", "turn/started", "error", "unknown/future"} {
		if informationalEvent(codexapp.Message{Method: method}) {
			t.Fatalf("semantic/unknown notification filtered: %s", method)
		}
	}
	for _, method := range []string{"item/reasoning/textDelta", "item/reasoning/summaryTextDelta", "item/reasoning/summaryPartAdded", "thread/status/changed", "item/started", "item/completed"} {
		if !informationalEvent(codexapp.Message{Method: method}) {
			t.Fatalf("informational event not filtered: %s", method)
		}
		if informationalEvent(codexapp.Message{Method: method, ID: json.RawMessage(`"request"`)}) {
			t.Fatalf("RPC request filtered: %s", method)
		}
	}
}

func TestOverflowFailureWinsOverAlreadySelectedToolBoundary(t *testing.T) {
	e, rpc, _ := fixture(t)
	q := request("overflow-boundary")
	first, err := e.Step(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	c, err := e.get(q)
	if err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	c.failure = eventQueueLimitError{}
	c.stopOnce.Do(func() { close(c.stopped) })
	// Model a quiet-boundary winner that was already selected as the reader
	// failed the owner: phase/deferred must never override terminal failure.
	c.phase = "waiting"
	c.deferred = &codexapp.Message{Method: "turn/completed", Params: json.RawMessage(`{"threadId":"thread-1","turn":{"id":"turn-thread-1","status":"completed"}}`)}
	c.mu.Unlock()
	q.Input = nil
	q.ExpectedPrefix = q.PrefixDigest
	q.PrefixDigest = "after-overflow"
	q.Results = []ToolResult{{ID: first.Tool.ID, Content: []Content{{Type: "inputText", Text: "done"}}, Success: true}}
	last, err := e.Step(context.Background(), q)
	if !errors.Is(err, ErrLimit) || last.Done || rpc.responses != 0 {
		t.Fatalf("terminal overflow overwritten: done=%v responses=%d err=%v", last.Done, rpc.responses, err)
	}
}
