package cli

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codexbridge"
	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

type managedGPTTestAccount struct{}

func (managedGPTTestAccount) Account(context.Context) (codexapp.AccountResult, error) {
	return codexapp.AccountResult{Account: &codexapp.Account{Type: "chatgpt", PlanType: "test"}}, nil
}

func managedGPTTestStore(t *testing.T) *codexbridge.FileStore {
	t.Helper()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	store, err := codexbridge.OpenStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestManagedGPTLiveSimpleTurn(t *testing.T) {
	if os.Getenv("MOAI_GPT_LIVE") != "1" {
		t.Skip("set MOAI_GPT_LIVE=1 for a subscription App Server turn")
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	client, err := codexapp.Start(ctx, codexapp.Config{Binary: binary, Home: profile})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err = client.Initialize(ctx, "moai-gpt-live-test", "1"); err != nil {
		t.Fatal(err)
	}
	authority, err := gateway.NewAppServerAuthority(client, "chatgpt", func(context.Context) (string, error) { return "live-scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	storeDir := t.TempDir()
	if err = os.Chmod(storeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	store, err := codexbridge.OpenStore(storeDir)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := codexbridge.New(ctx, client, codexbridge.Config{Store: store})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	for _, model := range []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna"} {
		t.Run(model, func(t *testing.T) {
			prepare := newManagedGPTPrepare("live-simple-"+model, t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxEventBytes: 1 << 20, MaxOutputBytes: 1 << 20}, store)
			adapter, adapterErr := gateway.NewAppServerAdapter(gateway.AppServerAdapterConfig{Engine: engine, Authority: authority, Prepare: prepare})
			if adapterErr != nil {
				t.Fatal(adapterErr)
			}
			entry := gateway.ModelEntry{RouteID: model, UpstreamID: model, Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
			grant, authorizeErr := authority.Authorize(ctx, entry)
			if authorizeErr != nil {
				t.Fatal(authorizeErr)
			}
			body, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 64, "stream": false, "thinking": map[string]any{"type": "adaptive"}, "output_config": map[string]any{"effort": "low"}, "messages": []any{map[string]string{"role": "user", "content": "Reply with exactly LIVE_APP_SERVER_OK."}}, "tools": []any{}})
			response, sendErr := adapter.Send(ctx, gateway.RoutedRequest{Entry: entry, Body: body, Managed: grant})
			if sendErr != nil {
				t.Fatalf("live adapter send: %T %v", sendErr, sendErr)
			}
			raw, readErr := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if readErr != nil || response.StatusCode != 200 || !strings.Contains(string(raw), "LIVE_APP_SERVER_OK") {
				t.Fatalf("live response status=%d body=%s err=%v", response.StatusCode, raw, readErr)
			}
		})
	}

	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(ctx, entry)
	if err != nil {
		t.Fatal(err)
	}
	toolPrepare := newManagedGPTPrepare("live-tool", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxEventBytes: 1 << 20, MaxOutputBytes: 1 << 20}, store)
	toolAdapter, err := gateway.NewAppServerAdapter(gateway.AppServerAdapterConfig{Engine: engine, Authority: authority, Prepare: toolPrepare})
	if err != nil {
		t.Fatal(err)
	}
	tool := map[string]any{"name": "echo", "description": "Return the supplied text", "input_schema": map[string]any{"type": "object", "properties": map[string]any{"text": map[string]string{"type": "string"}}, "required": []string{"text"}, "additionalProperties": false}}
	toolBody, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 128, "stream": false, "thinking": map[string]any{"type": "adaptive"}, "output_config": map[string]any{"effort": "low"}, "messages": []any{map[string]string{"role": "user", "content": "Call the echo tool exactly once with text ping, then report its result."}}, "tools": []any{tool}})
	toolResponse, err := toolAdapter.Send(ctx, gateway.RoutedRequest{Entry: entry, Body: toolBody, Managed: grant})
	if err != nil {
		t.Fatalf("live tool start: %T %v", err, err)
	}
	var toolMessage struct {
		Content []struct {
			Type, ID, Name string
			Input          map[string]any
		}
	}
	if err = json.NewDecoder(toolResponse.Body).Decode(&toolMessage); err != nil {
		_ = toolResponse.Body.Close()
		t.Fatal(err)
	}
	_ = toolResponse.Body.Close()
	var called struct {
		Type, ID, Name string
		Input          map[string]any
	}
	for _, block := range toolMessage.Content {
		if block.Type == "tool_use" {
			called = block
			break
		}
	}
	if called.ID == "" || called.Name != "echo" {
		t.Fatalf("live tool boundary=%+v", toolMessage.Content)
	}
	continuation, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 128, "stream": false, "thinking": map[string]any{"type": "adaptive"}, "output_config": map[string]any{"effort": "low"}, "messages": []any{
		map[string]string{"role": "user", "content": "Call the echo tool exactly once with text ping, then report its result."},
		map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": called.ID, "name": "echo", "input": called.Input}}},
		map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": called.ID, "content": "pong"}}},
	}, "tools": []any{tool}})
	continued, err := toolAdapter.Send(ctx, gateway.RoutedRequest{Entry: entry, Body: continuation, Managed: grant})
	if err != nil {
		t.Fatalf("live tool continuation: %T %v", err, err)
	}
	defer continued.Body.Close()
	continuedRaw, err := io.ReadAll(continued.Body)
	if err != nil || continued.StatusCode != 200 || !strings.Contains(strings.ToLower(string(continuedRaw)), "pong") {
		t.Fatalf("live continuation status=%d body=%s err=%v", continued.StatusCode, continuedRaw, err)
	}
}

func TestManagedGPTLiveResume(t *testing.T) {
	if os.Getenv("MOAI_GPT_LIVE") != "1" {
		t.Skip("set MOAI_GPT_LIVE=1 for a subscription App Server resume")
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	store := managedGPTTestStore(t)
	cwd := t.TempDir()
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	start := func() (*codexapp.Client, *codexbridge.Engine, *gateway.AppServerAdapter, *gateway.ManagedGrant) {
		client, startErr := codexapp.Start(ctx, codexapp.Config{Binary: binary, Home: profile})
		if startErr != nil {
			t.Fatal(startErr)
		}
		if _, startErr = client.Initialize(ctx, "moai-gpt-resume-test", "1"); startErr != nil {
			_ = client.Close()
			t.Fatal(startErr)
		}
		authority, startErr := gateway.NewAppServerAuthority(client, "chatgpt", func(context.Context) (string, error) { return "live-resume-scope", nil })
		if startErr != nil {
			_ = client.Close()
			t.Fatal(startErr)
		}
		engine, startErr := codexbridge.New(ctx, client, codexbridge.Config{Store: store})
		if startErr != nil {
			_ = client.Close()
			t.Fatal(startErr)
		}
		prepare := newManagedGPTPrepare("live-resume", cwd, translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxEventBytes: 1 << 20, MaxOutputBytes: 1 << 20}, store)
		adapter, startErr := gateway.NewAppServerAdapter(gateway.AppServerAdapterConfig{Engine: engine, Authority: authority, Prepare: prepare})
		if startErr != nil {
			engine.Close()
			_ = client.Close()
			t.Fatal(startErr)
		}
		grant, startErr := authority.Authorize(ctx, entry)
		if startErr != nil {
			engine.Close()
			_ = client.Close()
			t.Fatal(startErr)
		}
		return client, engine, adapter, grant
	}
	firstClient, firstEngine, firstAdapter, firstGrant := start()
	firstBody, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 64, "stream": false, "messages": []any{map[string]string{"role": "user", "content": "Reply with exactly LIVE_RESUME_BASE_OK."}}, "tools": []any{}})
	first, err := firstAdapter.Send(ctx, gateway.RoutedRequest{Entry: entry, Body: firstBody, Managed: firstGrant})
	if err != nil {
		firstEngine.Close()
		_ = firstClient.Close()
		t.Fatal(err)
	}
	firstRaw, readErr := io.ReadAll(first.Body)
	_ = first.Body.Close()
	firstEngine.Close()
	if closeErr := firstClient.Close(); readErr != nil || closeErr != nil || !strings.Contains(string(firstRaw), "LIVE_RESUME_BASE_OK") {
		t.Fatalf("first body=%s read=%v close=%v", firstRaw, readErr, closeErr)
	}

	secondClient, secondEngine, secondAdapter, secondGrant := start()
	defer secondEngine.Close()
	defer secondClient.Close()
	secondBody, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 64, "stream": false, "messages": []any{
		map[string]string{"role": "user", "content": "Reply with exactly LIVE_RESUME_BASE_OK."},
		map[string]string{"role": "assistant", "content": "LIVE_RESUME_BASE_OK"},
		map[string]string{"role": "user", "content": "Reply with exactly LIVE_RESUME_OK."},
	}, "tools": []any{}})
	second, err := secondAdapter.Send(ctx, gateway.RoutedRequest{Entry: entry, Body: secondBody, Managed: secondGrant})
	if err != nil {
		t.Fatalf("resume send: %T %v", err, err)
	}
	defer second.Body.Close()
	secondRaw, err := io.ReadAll(second.Body)
	if err != nil || !strings.Contains(string(secondRaw), "LIVE_RESUME_OK") {
		t.Fatalf("resume body=%s err=%v", secondRaw, err)
	}
}

func TestManagedGPTPrepareToolContinuation(t *testing.T) {
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(context.Background(), entry)
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare("conversation", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxEventBytes: 1 << 20, MaxOutputBytes: 1 << 20}, managedGPTTestStore(t))
	tool := map[string]any{"name": "Read", "description": "read", "input_schema": map[string]any{"type": "object", "properties": map[string]any{"file_path": map[string]string{"type": "string"}}, "required": []string{"file_path"}, "additionalProperties": false}}
	first, _ := json.Marshal(map[string]any{
		"model":         entry.RouteID,
		"max_tokens":    1024,
		"stream":        true,
		"thinking":      map[string]any{"type": "adaptive"},
		"output_config": map[string]any{"effort": "high"},
		"messages":      []any{map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": "read config"}}}},
		"tools":         []any{tool},
	})
	q1, err := prepare(context.Background(), gateway.RoutedRequest{Entry: entry, Body: first, Managed: grant})
	if err != nil || len(q1.Input) != 1 || len(q1.Results) != 0 || q1.Effort != "high" {
		t.Fatalf("first prepare=%+v err=%v", q1, err)
	}
	continuation, _ := json.Marshal(map[string]any{"model": entry.RouteID, "max_tokens": 1024, "stream": true, "thinking": map[string]any{"type": "adaptive"}, "output_config": map[string]any{"effort": "high"}, "messages": []any{
		map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": "read config"}}},
		map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": "toolu_moai_fixture", "name": "Read", "input": map[string]string{"file_path": "config"}}}},
		map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "toolu_moai_fixture", "content": []any{map[string]string{"type": "text", "text": "language: ko"}}}}},
		map[string]any{"role": "user", "content": []any{map[string]string{"type": "text", "text": "conditional skill attachment"}}},
	}, "tools": []any{tool}})
	q2, err := prepare(context.Background(), gateway.RoutedRequest{Entry: entry, Body: continuation, Managed: grant})
	if err != nil || len(q2.Input) != 0 || len(q2.Results) != 1 || len(q2.Results[0].Content) != 2 || q2.Results[0].ID != "toolu_moai_fixture" || q2.ExpectedPrefix != "" {
		t.Fatalf("continuation prepare=%+v err=%v", q2, err)
	}
}

func TestManagedGPTPrepareKeepsPromptBeforeTrailingAttachments(t *testing.T) {
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	entry := gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}
	grant, err := authority.Authorize(context.Background(), entry)
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare("conversation", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxEventBytes: 1 << 20, MaxOutputBytes: 1 << 20}, managedGPTTestStore(t))
	body, _ := json.Marshal(map[string]any{
		"model": entry.RouteID, "max_tokens": 1024, "stream": true, "tools": []any{},
		"messages": []any{
			map[string]string{"role": "user", "content": "environment attachment"},
			map[string]string{"role": "user", "content": "Reply with exactly UI_MAIN_OK."},
			map[string]string{"role": "user", "content": "output style attachment"},
		},
	})
	q, err := prepare(context.Background(), gateway.RoutedRequest{Entry: entry, Body: body, Managed: grant})
	if err != nil || len(q.Input) != 3 {
		t.Fatalf("prepare input=%+v err=%v", q.Input, err)
	}
	got := q.Input[1].(map[string]string)["text"]
	if got != "Reply with exactly UI_MAIN_OK." {
		t.Fatalf("middle prompt=%q", got)
	}
}

func TestManagedGPTPrepareSeparatesTitleAndMainThreads(t *testing.T) {
	authority, err := gateway.NewAppServerAuthority(managedGPTTestAccount{}, "chatgpt", func(context.Context) (string, error) { return "scope", nil })
	if err != nil {
		t.Fatal(err)
	}
	grant, err := authority.Authorize(context.Background(), gateway.ModelEntry{Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer})
	if err != nil {
		t.Fatal(err)
	}
	prepare := newManagedGPTPrepare("conversation", t.TempDir(), translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 1 << 20, MaxEventBytes: 1 << 20, MaxOutputBytes: 1 << 20}, managedGPTTestStore(t))
	titleBody, _ := json.Marshal(map[string]any{
		"model":      "gpt-5.6-luna",
		"max_tokens": 128,
		"stream":     true,
		"messages":   []any{map[string]string{"role": "user", "content": "name this session"}},
		"tools":      []any{},
		"output_config": map[string]any{"format": map[string]any{
			"type":   "json_schema",
			"schema": map[string]any{"type": "object", "properties": map[string]any{"title": map[string]string{"type": "string"}}, "required": []string{"title"}, "additionalProperties": false},
		}},
	})
	title, err := prepare(context.Background(), gateway.RoutedRequest{Entry: gateway.ModelEntry{RouteID: "gpt-5.6-luna", UpstreamID: "gpt-5.6-luna", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}, Body: titleBody, Managed: grant})
	if err != nil {
		t.Fatal(err)
	}
	mainBody, _ := json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 128, "stream": true, "messages": []any{map[string]string{"role": "user", "content": "implement the task"}}, "tools": []any{}})
	main, err := prepare(context.Background(), gateway.RoutedRequest{Entry: gateway.ModelEntry{RouteID: "gpt-5.6-sol", UpstreamID: "gpt-5.6-sol", Provider: gateway.ProviderOpenAI, AuthMethod: gateway.AuthAppServer}, Body: mainBody, Managed: grant})
	if err != nil {
		t.Fatal(err)
	}
	if title.Owner.ConversationID == main.Owner.ConversationID || main.ExpectedPrefix != "" || title.ExpectedPrefix != "" {
		t.Fatalf("title owner=%q prefix=%q main owner=%q prefix=%q", title.Owner.ConversationID, title.ExpectedPrefix, main.Owner.ConversationID, main.ExpectedPrefix)
	}
}
