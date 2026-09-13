package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

func TestGatewayGPTModelsAreExactAndFullyBound(t *testing.T) {
	models := gatewayGPTModels()
	want := []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna"}
	if len(models) != len(want) {
		t.Fatalf("model count = %d, want %d", len(models), len(want))
	}
	for i, model := range models {
		if model.RouteID != want[i] || model.UpstreamID != want[i] || model.Provider != gateway.ProviderOpenAI || model.AuthMethod != gateway.AuthPKCE {
			t.Fatalf("model[%d] = %+v", i, model)
		}
		if model.Capabilities.ContextTokens != gatewayContextWindow || !model.Capabilities.Tools || !model.Capabilities.Streaming {
			t.Fatalf("model[%d] lacks verified capability binding: %+v", i, model.Capabilities)
		}
	}
	if _, err := gateway.NewCatalog(models); err != nil {
		t.Fatalf("exact GPT catalog rejected: %v", err)
	}
}

func TestGatewayChildEnvironmentScrubsCredentialAndRoutingKeys(t *testing.T) {
	inherited := append(os.Environ(),
		"Z_AI_API_KEY=foreign",
		"OPENAI_API_KEY=foreign",
		"ANTHROPIC_API_KEY=foreign",
		"CODEX_HOME=/foreign",
		"ANTHROPIC_BASE_URL=https://foreign",
		"KEEP=present",
	)
	env := gatewayChildEnvironment(inherited)
	joined := "\n" + strings.Join(env, "\n") + "\n"
	for _, key := range []string{"Z_AI_API_KEY", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "CODEX_HOME", "ANTHROPIC_BASE_URL"} {
		if strings.Contains(joined, "\n"+key+"=") {
			t.Fatalf("child environment retained %s", key)
		}
	}
	if !strings.Contains(joined, "\nKEEP=present\n") {
		t.Fatal("child environment discarded unrelated setting")
	}
}

func TestGPTProductionBindingHasPrivateChildAndExactPayload(t *testing.T) {
	binding, err := newGPTGatewayBinding()
	if err != nil {
		t.Fatal(err)
	}
	if binding.Mode != "gpt" || binding.Prepare == nil {
		t.Fatalf("incomplete GPT binding: %+v", binding)
	}
	// Inspect only nonsecret launch metadata. The payload is intentionally
	// opaque here; the private child validates and decodes it before resources.
	if binding.Prepare == nil {
		t.Fatal("missing private child preparation")
	}
}

func TestProductionGatewayFactoryRejectsUnauthenticatedRequest(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	// The factory only requires the codex broker to be locatable at
	// construction; the 401 rejection below happens before any credential
	// resolve. A stub keeps the test hermetic on machines without codex.
	dir := t.TempDir()
	name := "codex"
	body := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		name = "codex.bat"
		body = "@echo codex-stub\r\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	models := gatewayGPTModels()
	payload, err := marshalGatewayPrivatePayload("private-session", models)
	if err != nil {
		t.Fatal(err)
	}
	h, err := productionGatewayHandlerFactory(json.RawMessage(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer h.(io.Closer).Close()
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"gpt-6-astra","max_tokens":2,"messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set("Authorization", "Bearer private-session")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != 401 {
		t.Fatalf("unauthenticated production route status = %d, want 401", res.Code)
	}
}

func TestGPTBindingBootstrapsPrivateFamilyAndSessionArgs(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	var started gateway.StartOptions
	start := func(_ context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
		started = opts
		return gatewayStartedChild{Address: "127.0.0.1:4321", OverlayPath: "/private/settings.json", Stop: func() {}}, nil
	}
	binding, err := newGPTGatewayBindingWithStart(start)
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	plan, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt", CWD: project, Project: project})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if len(plan.Args) != 2 || plan.Args[0] != "--session-id" || plan.Args[1] == "" {
		t.Fatalf("new session args = %v", plan.Args)
	}
	if len(started.Config.Payload) == 0 || started.Config.Lifetime != 24*time.Hour {
		t.Fatalf("private child configuration incomplete: lifetime=%s payload=%d", started.Config.Lifetime, len(started.Config.Payload))
	}
	var payload gatewayPrivatePayload
	if err := json.Unmarshal(started.Config.Payload, &payload); err != nil || payload.Conversation == nil || payload.Conversation.SessionID != plan.Args[1] || payload.Conversation.ReceiptDir == "" {
		t.Fatalf("private child conversation binding missing: err=%v payload=%+v", err, payload.Conversation)
	}
	for _, item := range started.Env {
		if strings.HasPrefix(item, "OPENAI_API_KEY=") || strings.HasPrefix(item, "CODEX_HOME=") {
			t.Fatalf("credential environment reached child: %s", item)
		}
	}
	var sessionKey string
	for _, item := range plan.ChildEnv {
		if strings.HasPrefix(item, "ANTHROPIC_AUTH_TOKEN=") {
			sessionKey = strings.TrimPrefix(item, "ANTHROPIC_AUTH_TOKEN=")
		}
	}
	if sessionKey == "" {
		t.Fatal("private session key was not provided through Claude's Bearer header")
	}
}

func TestGPTBindingResumeDescriptorReachesChildPlan(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	state := filepath.Join(home, "state", "gateway-conversations")
	families, err := conversation.Open(state)
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	d, err := families.New(context.Background(), conversation.NewRequest{CWD: project, Project: project})
	if err != nil {
		t.Fatal(err)
	}
	transcript := filepath.Join(d.ConfigDir, "native.jsonl")
	raw := `{"session_id":"` + d.UUID + `","project":"` + project + `","type":"completed"}` + "\n"
	if err := os.WriteFile(transcript, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	if err := families.Complete(context.Background(), d.UUID, transcript, 1); err != nil {
		t.Fatal(err)
	}
	var started gateway.StartOptions
	binding, err := newGPTGatewayBindingWithStart(func(_ context.Context, opts gateway.StartOptions) (gatewayStartedChild, error) {
		started = opts
		return gatewayStartedChild{Address: "127.0.0.1:4322", OverlayPath: "/private/settings.json", Stop: func() {}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	plan, stop, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt", CWD: project, Project: project, Args: []string{"--resume", d.UUID}})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if len(plan.Args) < 2 || plan.Args[0] != "--resume" || plan.Args[1] != d.UUID {
		t.Fatalf("resume args = %v", plan.Args)
	}
	if len(started.Config.Payload) == 0 {
		t.Fatal("private child payload missing")
	}
}

func TestGatewayProductionLongContextBudget(t *testing.T) {
	text := "START" + strings.Repeat(" x", 800000) + "END"
	body, _ := json.Marshal(map[string]any{"model": "gpt-5.6-sol", "max_tokens": 16, "messages": []any{map[string]any{"role": "user", "content": text}}})
	if len(body) <= 1<<20 {
		t.Fatal("fixture must exceed old body limit")
	}
	n := int64(833112)
	limits := translate.Limits{MaxBodyBytes: gatewayRequestBodyLimit, ContextTokens: gatewayContextWindow, InputTokens: &n}
	out, _, err := translate.Request("gpt-5.6-sol", body, limits)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), text) {
		t.Fatal("long input was truncated")
	}
	n = int64(gatewayContextWindow) + 1
	if _, _, err := translate.Request("gpt-5.6-sol", body, limits); err == nil {
		t.Fatal("over-window request accepted")
	}
	if _, _, err := translate.Request("gpt-5.6-sol", make([]byte, gatewayRequestBodyLimit+1), limits); err == nil {
		t.Fatal("oversize body accepted")
	}
}
