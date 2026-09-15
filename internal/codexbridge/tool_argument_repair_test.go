package codexbridge

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

type argumentRepairRPC struct {
	*fakeRPC
	rejected      int
	alwaysInvalid bool
}

func (r *argumentRepairRPC) Respond(ctx context.Context, id json.RawMessage, value any) error {
	if value.(map[string]any)["success"] == false {
		r.rejected++
		contents := value.(map[string]any)["contentItems"].([]Content)
		if len(contents) != 1 || !strings.Contains(contents[0].Text, "schema") {
			return ErrProtocol
		}
		raw, _ := json.Marshal(map[string]any{"threadId": "thread-1", "turnId": "turn-thread-1", "callId": "corrected", "tool": "Agent", "arguments": map[string]string{"description": "Inspect repository", "prompt": "public fixture"}})
		if r.alwaysInvalid {
			raw = []byte(`{"threadId":"thread-1","turnId":"turn-thread-1","callId":"bad-again","tool":"Agent","arguments":{"prompt":"public fixture"}}`)
		}
		r.events <- codexapp.Message{ID: json.RawMessage(`"corrected-rpc"`), Method: "item/tool/call", Params: raw}
		return nil
	}
	return r.fakeRPC.Respond(ctx, id, value)
}

func TestToolArgumentRecoveryBudgetAndAuthorityRemainClosed(t *testing.T) {
	for _, kind := range []string{"budget", "unknown-tool", "foreign-thread"} {
		t.Run(kind, func(t *testing.T) {
			_, _, store := fixture(t)
			fake := &fakeRPC{events: make(chan codexapp.Message, 32)}
			fake.transform = func(event codexapp.Message) codexapp.Message {
				tool, thread := "Agent", "thread-1"
				if kind == "unknown-tool" {
					tool = "unregistered"
				}
				if kind == "foreign-thread" {
					thread = "thread-foreign"
				}
				event.Params, _ = json.Marshal(map[string]any{"threadId": thread, "turnId": "turn-thread-1", "callId": "bad", "tool": tool, "arguments": map[string]string{"prompt": "public fixture"}})
				return event
			}
			rpc := &argumentRepairRPC{fakeRPC: fake, alwaysInvalid: true}
			engine, err := New(context.Background(), rpc, Config{Store: store})
			if err != nil {
				t.Fatal(err)
			}
			defer engine.Close()
			q := request("closed-" + kind)
			q.Tools = []codextools.Definition{{Name: "Agent", InputSchema: json.RawMessage(`{"type":"object","properties":{"description":{"type":"string"},"prompt":{"type":"string"}},"required":["description","prompt"],"additionalProperties":false}`)}}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err = engine.Step(ctx, q)
			if err == nil {
				t.Fatal("invalid operation accepted")
			}
			if kind == "budget" {
				if rpc.rejected != 3 || !errors.Is(err, ErrProtocol) {
					t.Fatalf("unbounded feedback or wrong error: count=%d err=%v", rpc.rejected, err)
				}
			} else if rpc.rejected != 0 {
				t.Fatal("authority violation received repair feedback")
			}
		})
	}
}

func TestSchemaInvalidAgentCallCanCorrectWithoutPoisoningConversation(t *testing.T) {
	_, _, store := fixture(t)
	fake := &fakeRPC{events: make(chan codexapp.Message, 32)}
	fake.transform = func(event codexapp.Message) codexapp.Message {
		event.Params = []byte(`{"threadId":"thread-1","turnId":"turn-thread-1","callId":"missing-description","tool":"Agent","arguments":{"prompt":"public fixture"}}`)
		return event
	}
	rpc := &argumentRepairRPC{fakeRPC: fake}
	engine, err := New(context.Background(), rpc, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	q := request("argument-repair")
	q.Tools = []codextools.Definition{{Name: "Agent", InputSchema: json.RawMessage(`{"type":"object","properties":{"description":{"type":"string"},"prompt":{"type":"string"}},"required":["description","prompt"],"additionalProperties":false}`)}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	segment, err := engine.Step(ctx, q)
	if err != nil {
		t.Fatalf("model argument validation poisoned session: %v", err)
	}
	if rpc.rejected != 1 || len(segment.Tools) != 1 || segment.Tools[0].Name != "Agent" || !strings.Contains(string(segment.Tools[0].Arguments), "Inspect repository") {
		t.Fatalf("invalid call exposed or correction lost: rejected=%d segment=%+v", rpc.rejected, segment)
	}
	barrier, ok, err := store.Barrier(q.Owner)
	if err != nil || !ok || barrier.Phase != "waiting" {
		t.Fatalf("corrected call barrier=%+v found=%v err=%v", barrier, ok, err)
	}
	fake.mu.Lock()
	interrupts := len(fake.interrupts)
	fake.mu.Unlock()
	if interrupts != 0 {
		t.Fatal("correctable arguments caused turn interrupt")
	}
}
