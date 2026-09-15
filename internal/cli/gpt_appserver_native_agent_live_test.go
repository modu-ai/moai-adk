package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

// Inspect only the isolation config, never credentials or full request bodies.
type nativeAgentIsolationRPC struct {
	*codexapp.Client
	starts int
}

func (r *nativeAgentIsolationRPC) Call(ctx context.Context, method string, params, result any) error {
	if method == "thread/start" {
		raw, err := json.Marshal(params)
		if err != nil {
			return err
		}
		var request struct {
			Config map[string]json.RawMessage `json:"config"`
		}
		if err := json.Unmarshal(raw, &request); err != nil {
			return err
		}
		if string(request.Config["agents.enabled"]) != "false" {
			return fmt.Errorf("thread/start missing agents.enabled=false isolation")
		}
		r.starts++
	}
	return r.Client.Call(ctx, method, params, result)
}

// This proves the bridge-supplied Agent tool survives native-agent isolation.
// The result is synthetic: it does not claim a real Claude child was launched,
// or that non-use alone proves every Codex native-agent tool was unavailable.
func TestSharedGPTLiveNativeAgentIsolationPreservesDynamicAgent(t *testing.T) {
	if os.Getenv("MOAI_GPT_LIVE") != "1" {
		t.Skip("subscription agent isolation probe requires MOAI_GPT_LIVE=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 110*time.Second)
	defer cancel()
	profile, err := managedGPTProfile()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	client, err := codexapp.ConnectShared(ctx, codexapp.Config{Binary: binary, Home: profile})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }() // Best-effort teardown; primary assertions own failures.
	if _, err := client.Initialize(ctx, "moai-native-agent-isolation-live", "1"); err != nil {
		t.Fatal(err)
	}
	account, err := client.Account(ctx)
	if err != nil || account.Account == nil || account.Account.Type != "chatgpt" {
		t.Fatalf("subscription account unavailable: %v", err)
	}
	rpc := &nativeAgentIsolationRPC{Client: client}
	store := managedGPTTestStore(t)
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	const marker = "DYNAMIC_AGENT_ISOLATION_OK"
	q := codexbridge.Request{
		Owner: codextools.Binding{ConversationID: fmt.Sprintf("agent-isolation-%d", time.Now().UnixNano()), AccountScope: "agent-isolation-live"},
		Model: "gpt-5.6-sol", Effort: "low", CWD: t.TempDir(), PrefixDigest: "agent-isolation-1",
		Tools: []codextools.Definition{{Name: "Agent", Description: "Delegate a synthetic task to the outer Claude Code agent interface and return its result. The task is only an echo probe.", InputSchema: json.RawMessage(`{"type":"object","properties":{"prompt":{"type":"string"}},"required":["prompt"],"additionalProperties":false}`)}},
		Input: []any{map[string]string{"type": "text", "text": "Delegate exactly once using the available Agent tool with prompt " + marker + ". Wait for its result, then report that result verbatim. Do not run shell commands, native agents, or any other tools."}},
	}
	started := time.Now()
	seg, err := engine.Step(ctx, q)
	if err != nil {
		t.Fatalf("dynamic Agent start: %v", err)
	}
	if rpc.starts != 1 || seg.Tool == nil || seg.Tool.Name != "Agent" {
		t.Fatalf("expected isolated thread and dynamic Agent: starts=%d tool=%+v done=%v", rpc.starts, seg.Tool, seg.Done)
	}
	var args struct {
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal(seg.Tool.Arguments, &args); err != nil || args.Prompt != marker {
		t.Fatalf("dynamic Agent did not carry the exact synthetic task: %v", err)
	}
	firstTool := time.Since(started)
	q.Input, q.ExpectedPrefix, q.PrefixDigest = nil, q.PrefixDigest, "agent-isolation-2"
	q.Results = []codexbridge.ToolResult{{ID: seg.Tool.ID, Success: true, Content: []codexbridge.Content{{Type: "inputText", Text: marker}}}}
	seg, err = engine.Step(ctx, q)
	if err != nil || !seg.Done || !strings.Contains(seg.Text, marker) {
		t.Fatalf("dynamic Agent result completion: done=%v text=%q err=%v", seg.Done, seg.Text, err)
	}
	barrier, _, err := store.Barrier(q.Owner)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("model=%s phase=%s agents.enabled=false dynamic_Agent_returned=true first_tool=%s total=%s result=%q", q.Model, barrier.Phase, firstTool.Round(time.Millisecond), time.Since(started).Round(time.Millisecond), seg.Text)
}
