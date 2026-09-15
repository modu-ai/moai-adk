package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codextools"
	"github.com/modu-ai/moai-adk/internal/gateway/translate"
)

func TestClaudeLiveEmptyToolsNativeShape(t *testing.T) {
	if os.Getenv("MOAI_CLAUDE_TOOLS_PROBE") != "1" {
		t.Skip("explicit local Claude request capture")
	}
	for _, empty := range []bool{true, false} {
		t.Run(map[bool]string{true: "empty", false: "default"}[empty], func(t *testing.T) {
			var mu sync.Mutex
			var bodies [][]byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, err := io.ReadAll(io.LimitReader(r.Body, 16<<20))
				if err != nil {
					w.WriteHeader(400)
					return
				}
				if strings.Contains(r.URL.Path, "messages") && !strings.Contains(r.URL.Path, "count_tokens") {
					mu.Lock()
					bodies = append(bodies, raw)
					mu.Unlock()
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"probe\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"gpt-5.6-sol\",\"content\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\nevent: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"LOCAL_OK\"}}\n\nevent: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":1}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			args := []string{"--print", "--no-session-persistence", "--model", "gpt-5.6-sol", "--settings", `{"disableAllHooks":true}`, "Reply LOCAL_OK."}
			mcpConfig := os.Getenv("MOAI_CLAUDE_TOOLS_MCP_CONFIG")
			if mcpConfig != "" {
				args = append(args, "--mcp-config", mcpConfig)
			}
			if empty {
				args = append(args, "--tools", "")
			}
			cmd := exec.CommandContext(ctx, "claude", args...)
			cmd.Dir = t.TempDir()
			cmd.Env = append(gatewayChildEnvironment(os.Environ()), "CLAUDE_CONFIG_DIR="+t.TempDir(), "ANTHROPIC_BASE_URL="+server.URL, "ANTHROPIC_AUTH_TOKEN=local-synthetic", "ENABLE_TOOL_SEARCH=false", "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1")
			cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
			runErr := cmd.Run()
			mu.Lock()
			captured := append([][]byte(nil), bodies...)
			mu.Unlock()
			if len(captured) == 0 {
				t.Fatalf("no request captured; command_error_type=%T", runErr)
			}
			for index, body := range captured {
				if managedGPTTitleRequest(body) {
					continue
				}
				_, _, err := translate.RequestContext(ctx, "gpt-5.6-sol", body, translate.Limits{PolicyProfile: translate.PolicyGPTNative, MaxBodyBytes: 16 << 20, MaxEventBytes: 1 << 20, MaxOutputBytes: 8 << 20, NativeReceiptAuthorize: func(context.Context, []byte, translate.NativePolicy) error { return nil }})
				if err != nil {
					t.Fatalf("translation error_type=%T", err)
				}
				defs, _, _, _, err := managedGPTPublicDelta(body)
				if err != nil {
					t.Fatal("delta rejected")
				}
				if empty && mcpConfig == "" && len(defs) != 0 || !empty && len(defs) == 0 {
					t.Fatalf("unexpected tool mode projection: empty=%v count=%d", empty, len(defs))
				}
				registry, err := codextools.New(codextools.Binding{ConversationID: "shape-probe", AccountScope: "synthetic"}, defs)
				if err != nil {
					t.Fatal("registry rejected")
				}
				names := []string{}
				for _, tool := range registry.NativeTools() {
					if !json.Valid(tool.InputSchema) {
						t.Fatal("invalid native schema")
					}
					names = append(names, tool.Name)
				}
				if len(names) != len(defs)+1 {
					t.Fatal("native registry declaration count drift")
				}
				mcpCount, reservedCount := 0, 0
				for _, def := range defs {
					if def.Name == "mcp" || strings.HasPrefix(def.Name, "mcp__") {
						mcpCount++
					}
				}
				for _, name := range names {
					if name == "mcp" || strings.HasPrefix(name, "mcp__") {
						reservedCount++
					}
				}
				t.Logf("public_tools=%d public_mcp=%d native_tools=%d native_reserved=%d", len(defs), mcpCount, len(names), reservedCount)
				if empty && index == 0 && os.Getenv("MOAI_CLAUDE_TOOLS_NATIVE_LIVE") == "1" {
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
					if _, err = client.Initialize(ctx, "moai-captured-schema-probe", "1"); err != nil {
						t.Fatal("initialization failed")
					}
					var reply struct {
						Thread struct {
							ID string `json:"id"`
						} `json:"thread"`
					}
					err = client.Call(ctx, "thread/start", map[string]any{"model": "gpt-5.6-sol", "cwd": cmd.Dir, "dynamicTools": registry.NativeTools(), "approvalPolicy": "never", "sandbox": "read-only", "environments": []any{}, "config": map[string]any{"agents.enabled": false, "features.multi_agent": false, "features.multi_agent_v2": false}}, &reply)
					if err != nil {
						var rpcErr *codexapp.RPCError
						if errors.As(err, &rpcErr) {
							t.Fatalf("captured MCP catalog rpc_code=%d", rpcErr.Code)
						}
						t.Fatalf("captured MCP catalog error_type=%T", err)
					}
					if reply.Thread.ID == "" {
						t.Fatal("missing allocated thread")
					}
					t.Log("captured_mcp_catalog_native_thread_start=accepted")
				}
			}
		})
	}
}
