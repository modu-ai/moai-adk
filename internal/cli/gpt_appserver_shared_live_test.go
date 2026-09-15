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

// Run only after the product has started its shared owner. This never spawns a
// second authentication owner or changes the operator's login state.
func TestSharedGPTLiveConcurrentToolRouting(t *testing.T) {
	if os.Getenv("MOAI_GPT_LIVE") != "1" {
		t.Skip("authenticated shared routing requires MOAI_GPT_LIVE=1")
	}
	profile, err := managedGPTProfile()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	binary, _ = filepath.Abs(binary)
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	clients := make([]*codexapp.Client, 2)
	engines := make([]*codexbridge.Engine, 2)
	for i := range clients {
		clients[i], err = codexapp.ConnectShared(ctx, codexapp.Config{Binary: binary, Home: profile})
		if err != nil {
			t.Fatalf("start product shared owner first: %v", err)
		}
		t.Cleanup(func() { _ = clients[i].Close() })
		if _, err := clients[i].Initialize(ctx, fmt.Sprintf("moai-shared-live-%d", i), "1"); err != nil {
			t.Fatal(err)
		}
		account, err := clients[i].Account(ctx)
		if err != nil || account.Account == nil || account.Account.Type != "chatgpt" {
			t.Fatalf("subscription account unavailable: %v", err)
		}
		engines[i], err = codexbridge.New(ctx, clients[i], codexbridge.Config{Store: managedGPTTestStore(t)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { engines[i].Close() })
	}
	models := []string{"gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna", "gpt-6-astra"}
	for pair := 0; pair < 2; pair++ {
		results := make(chan error, 2)
		for i := range engines {
			model := models[pair*2+i]
			marker := fmt.Sprintf("SHARED_CLIENT_%d_PAIR_%d", i, pair)
			go func() { results <- sharedLiveToolRoundtrip(ctx, engines[i], model, marker, profile) }()
		}
		for range engines {
			if err := <-results; err != nil {
				t.Fatal(err)
			}
		}
	}
	engines[0].Close()
	if err := clients[0].Close(); err != nil {
		t.Fatal(err)
	}
	q := codexbridge.Request{Owner: codextools.Binding{ConversationID: "shared-peer-survives", AccountScope: "shared-live"}, Model: "gpt-5.6-sol", Effort: "low", CWD: profile, PrefixDigest: "peer-1", Input: []any{map[string]string{"type": "text", "text": "Reply exactly SHARED_PEER_ALIVE."}}}
	seg, err := engines[1].Step(ctx, q)
	if err != nil || !seg.Done || !strings.Contains(seg.Text, "SHARED_PEER_ALIVE") {
		t.Fatalf("peer shutdown interrupted remaining client: %+v %v", seg, err)
	}
}

func sharedLiveToolRoundtrip(ctx context.Context, engine *codexbridge.Engine, model, marker, cwd string) error {
	q := codexbridge.Request{Owner: codextools.Binding{ConversationID: marker, AccountScope: "shared-live"}, Model: model, Effort: "low", CWD: cwd, PrefixDigest: marker + "-1",
		Tools: []codextools.Definition{{Name: "echo", Description: "Return the supplied text unchanged", InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)}},
		Input: []any{map[string]string{"type": "text", "text": "Call echo exactly once with text " + marker + ". Then report its result verbatim."}}}
	seg, err := engine.Step(ctx, q)
	if err != nil {
		return fmt.Errorf("%s tool start: %w", model, err)
	}
	if seg.Tool == nil || seg.Tool.Name != "echo" {
		return fmt.Errorf("%s missing native echo call", model)
	}
	var args struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(seg.Tool.Arguments, &args) != nil || args.Text != marker {
		return fmt.Errorf("%s tool arguments crossed client boundary", model)
	}
	q.Input, q.ExpectedPrefix, q.PrefixDigest = nil, q.PrefixDigest, marker+"-2"
	q.Results = []codexbridge.ToolResult{{ID: seg.Tool.ID, Content: []codexbridge.Content{{Type: "inputText", Text: marker}}, Success: true}}
	seg, err = engine.Step(ctx, q)
	if err != nil {
		return fmt.Errorf("%s tool continuation: %w", model, err)
	}
	if !seg.Done || !strings.Contains(seg.Text, marker) {
		return fmt.Errorf("%s tool completion did not preserve client marker", model)
	}
	return nil
}
