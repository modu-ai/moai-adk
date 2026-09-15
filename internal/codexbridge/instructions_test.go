package codexbridge

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexapp"
)

type instructionRPC struct {
	*fakeRPC
	captureMu          sync.Mutex
	updates            []string
	turnInstructions   []string
	staticInstructions bool
}

func (r *instructionRPC) Call(ctx context.Context, method string, p any, out any) error {
	if method == "thread/start" {
		_, r.staticInstructions = p.(map[string]any)["developerInstructions"]
	}
	if method == "turn/start" {
		if mode, ok := p.(map[string]any)["collaborationMode"].(map[string]any); ok {
			settings := mode["settings"].(map[string]any)
			r.turnInstructions = append(r.turnInstructions, settings["developer_instructions"].(string))
		}
	}
	if method == "thread/resume" {
		if value, ok := p.(map[string]any)["developerInstructions"].(string); ok {
			r.captureMu.Lock()
			r.updates = append(r.updates, value)
			r.captureMu.Unlock()
		}
	}
	err := r.fakeRPC.Call(ctx, method, p, out)
	if err == nil && method == "turn/start" {
		thread := p.(map[string]any)["threadId"].(string)
		r.mu.Lock()
		turn := r.currentTurn[thread]
		r.mu.Unlock()
		raw, _ := json.Marshal(map[string]any{"threadId": thread, "turn": map[string]string{"id": turn, "status": "completed"}})
		r.events <- codexapp.Message{Method: "turn/completed", Params: raw}
	}
	return err
}

func TestInstructionsUseTurnOverrideNotIgnoredLoadedThreadResume(t *testing.T) {
	_, _, store := fixture(t)
	rpc := &instructionRPC{fakeRPC: &fakeRPC{events: make(chan codexapp.Message, 32), suppressTool: true}}
	engine, err := New(context.Background(), rpc, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	q := request("turn-instructions")
	for _, instructions := range []string{"PHASE_ALPHA", "PHASE_BETA", ""} {
		q.Instructions = instructions
		if _, err := engine.Step(context.Background(), q); err != nil {
			t.Fatal(err)
		}
		q.ExpectedPrefix = q.PrefixDigest
		q.PrefixDigest += "-next"
	}
	if rpc.staticInstructions {
		t.Error("mutable instructions pinned to static thread config")
	}
	if len(rpc.updates) != 0 {
		t.Error("instruction update used ignored loaded-thread resume override")
	}
	if len(rpc.turnInstructions) != 3 {
		t.Fatalf("turn instruction overrides = %v", rpc.turnInstructions)
	}
	for i, want := range []string{"PHASE_ALPHA", "PHASE_BETA", ""} {
		if rpc.turnInstructions[i] != want {
			t.Errorf("turn %d: got %q want %q", i, rpc.turnInstructions[i], want)
		}
	}
}

func TestIdleInstructionsUpdateOnceAndPersistAcrossResume(t *testing.T) {
	_, _, store := fixture(t)
	rpc := &instructionRPC{fakeRPC: &fakeRPC{events: make(chan codexapp.Message, 32), suppressTool: true}}
	engine, err := New(context.Background(), rpc, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	q := request("instructions")
	q.Instructions = "Answer concisely"
	if _, err := engine.Step(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	q.ExpectedPrefix = q.PrefixDigest
	q.PrefixDigest = "prefix-two"
	if _, err := engine.Step(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	rpc.captureMu.Lock()
	count := len(rpc.updates)
	rpc.captureMu.Unlock()
	if count != 0 {
		t.Fatal("unchanged instructions caused extra RPC")
	}
	q.ExpectedPrefix = q.PrefixDigest
	q.PrefixDigest = "prefix-three"
	q.Instructions = "Answer with code examples"
	if _, err := engine.Step(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	rpc.captureMu.Lock()
	updates := append([]string(nil), rpc.updates...)
	rpc.captureMu.Unlock()
	if len(updates) != 0 || len(rpc.turnInstructions) != 3 || rpc.turnInstructions[2] != q.Instructions {
		t.Fatalf("updated instructions not applied at idle: resume=%v turn=%v", updates, rpc.turnInstructions)
	}
	engine.Close()
	secondRPC := &instructionRPC{fakeRPC: &fakeRPC{events: make(chan codexapp.Message, 32), suppressTool: true}}
	resumed, err := New(context.Background(), secondRPC, Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer resumed.Close()
	q.Resume = true
	q.ExpectedPrefix = q.PrefixDigest
	q.PrefixDigest = "prefix-four"
	if _, err := resumed.Step(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	secondRPC.captureMu.Lock()
	count = len(secondRPC.updates)
	secondRPC.captureMu.Unlock()
	if count != 0 {
		t.Fatal("durable instruction digest lost across resume")
	}
}
