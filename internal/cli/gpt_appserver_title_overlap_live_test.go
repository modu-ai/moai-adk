package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

type titleOverlapRPC struct {
	*codexapp.Client
	started chan struct{}
	once    sync.Once
}

func (r *titleOverlapRPC) Call(ctx context.Context, method string, params any, out any) error {
	err := r.Client.Call(ctx, method, params, out)
	if method == "turn/start" && err == nil {
		r.once.Do(func() { close(r.started) })
	}
	return err
}

func TestSharedGPTLiveTitleMainSameClient(t *testing.T) {
	if os.Getenv("MOAI_GPT_TITLE_OVERLAP_LIVE") != "1" {
		t.Skip("explicit authenticated title overlap probe")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	profile, err := managedGPTProfile()
	if err != nil {
		t.Fatal("profile unavailable")
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal("codex unavailable")
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal("codex path unavailable")
	}
	client, err := codexapp.ConnectShared(ctx, codexapp.Config{Binary: binary, Home: profile})
	if err != nil {
		t.Fatal("shared connection unavailable")
	}
	defer func() { _ = client.Close() }()
	if _, err = client.Initialize(ctx, "moai-title-overlap-probe", "1"); err != nil {
		t.Fatal("initialize failed")
	}
	rpc := &titleOverlapRPC{Client: client, started: make(chan struct{})}
	engine, err := codexbridge.New(ctx, rpc, codexbridge.Config{Store: managedGPTTestStore(t)})
	if err != nil {
		t.Fatal("engine unavailable")
	}
	defer engine.Close()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal("cwd unavailable")
	}
	if override := os.Getenv("MOAI_GPT_TITLE_PROBE_CWD"); override != "" {
		if !filepath.IsAbs(override) {
			t.Fatal("probe cwd must be absolute")
		}
		cwd = override
	}
	title := codexbridge.Request{Owner: codextools.Binding{ConversationID: fmt.Sprintf("title-overlap-%d:title", time.Now().UnixNano()), AccountScope: "title-overlap-live"}, Model: "gpt-5.6-luna", Effort: "low", CWD: cwd, PrefixDigest: "title", Ephemeral: true,
		Instructions: "Generate a short session title and nothing else.", Input: []any{map[string]string{"type": "text", "text": "A synthetic concurrent gateway diagnostic"}}, OutputSchema: json.RawMessage(`{"type":"object","properties":{"title":{"type":"string"}},"required":["title"],"additionalProperties":false}`)}
	titleResult := make(chan error, 1)
	go func() { _, err := engine.Step(ctx, title); titleResult <- err }()
	select {
	case <-rpc.started:
	case err := <-titleResult:
		t.Fatalf("title never active; error_type=%T", err)
	case <-ctx.Done():
		t.Fatal("title start deadline")
	}
	main := title
	main.Owner.ConversationID += "-main"
	main.Model = "gpt-5.6-sol"
	main.PrefixDigest = "main"
	main.Ephemeral = false
	main.OutputSchema = nil
	main.Instructions = "Reply exactly MAIN_OVERLAP_OK. Do not call tools."
	toolCount := 15
	if os.Getenv("MOAI_GPT_TITLE_PROBE_NO_TOOLS") == "1" {
		toolCount = 0
	}
	for i := 0; i < toolCount; i++ {
		main.Tools = append(main.Tools, codextools.Definition{Name: fmt.Sprintf("probe_%d", i), Description: "Synthetic unused tool", InputSchema: json.RawMessage(`{"type":"object"}`)})
	}
	if os.Getenv("MOAI_GPT_TITLE_PROBE_MCP_TOOL") == "1" {
		main.Tools = []codextools.Definition{{Name: "mcp__synthetic__probe", Description: "Synthetic unused MCP tool", InputSchema: json.RawMessage(`{"type":"object"}`)}}
	}
	for i := 0; i < 4; i++ {
		if _, err := client.Account(ctx); err != nil {
			t.Fatal("interleaved account read failed")
		}
	}
	seg, mainErr := engine.Step(ctx, main)
	for name, err := range map[string]error{"main": mainErr, "title": <-titleResult} {
		if err != nil {
			var rpcErr *codexapp.RPCError
			if errors.As(err, &rpcErr) {
				t.Errorf("%s rpc_code=%d", name, rpcErr.Code)
			} else {
				t.Errorf("%s error_type=%T", name, err)
			}
		}
	}
	if mainErr == nil && !seg.Done {
		t.Fatal("main did not complete")
	}
}
