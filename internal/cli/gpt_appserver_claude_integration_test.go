package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
	"golang.org/x/net/websocket"
)

// Real Claude Code drives its Agent tool through the production gateway,
// bridge and shared transport. Only the final App Server peer is simulated;
// no provider account, user profile or external model request is used.
func TestClaudeCodeProductionGatewayAgentRoundTrip(t *testing.T) {
	if os.Getenv("MOAI_CLAUDE_INTEGRATION") != "1" {
		t.Skip("set MOAI_CLAUDE_INTEGRATION=1 for installed Claude Code local integration")
	}
	if runtime.GOOS == "windows" {
		t.Skip("shared transport integration requires Unix")
	}
	claude, err := exec.LookPath("claude")
	if err != nil {
		t.Fatal(err)
	}
	claude, err = filepath.Abs(claude)
	if err != nil {
		t.Fatal(err)
	}
	home, err := filepath.EvalSymlinks(sharedGPTFixtureHome(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", home)
	binDir, _ := installProductionWiringFakeCodex(t)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	const family = "320262b9-f176-4b4d-b72c-509adb15a34a"
	const token = "synthetic-local-gateway-token"
	receiptDir := filepath.Join(home, "receipts")
	if err := os.Mkdir(receiptDir, 0700); err != nil {
		t.Fatal(err)
	}
	ledger, err := receipt.OpenStore(context.Background(), receiptDir, family, true)
	if err != nil {
		t.Fatal(err)
	}
	_ = ledger.Close()
	fixture := startClaudeGatewayPeer(t, home)
	payload, _ := json.Marshal(gatewayPrivatePayload{Version: 1, SessionToken: token, ModelIDs: []string{"gpt-5.6-sol"}, Conversation: &gatewayPrivateConversation{FamilyID: family, SessionID: family, ReceiptDir: receiptDir, CWD: home}})
	handler, err := productionGatewayHandlerFactory(payload)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = handler.(io.Closer).Close() })
	var requestMu sync.Mutex
	mainRequests, agentRequests := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMu.Lock()
		if r.Header.Get("X-Claude-Code-Agent-Id") == "" {
			mainRequests++
		} else {
			agentRequests++
		}
		requestMu.Unlock()
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, claude, "--print", "--output-format", "json", "--session-id", family, "--model", "gpt-5.6-sol", "--tools", "Agent", "--allowedTools", "Agent", "--agents", `{"probe":{"description":"local integration child","prompt":"MOAI_CHILD_PROBE: reply CHILD_GATEWAY_OK","tools":[]}}`, "--max-turns", "5", "Use the probe agent and reply MAIN_GATEWAY_OK")
	cmd.Dir = home
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "CLAUDE_CONFIG_DIR=" + filepath.Join(home, "native"), "ANTHROPIC_API_KEY=local-probe", "ANTHROPIC_AUTH_TOKEN=" + token, "ANTHROPIC_BASE_URL=" + server.URL, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1", "DISABLE_AUTOUPDATER=1"}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil || !strings.Contains(stdout.String(), "MAIN_GATEWAY_OK") {
		t.Fatalf("Claude local integration: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	var result struct {
		Usage struct {
			Input  int64 `json:"input_tokens"`
			Output int64 `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Usage.Input <= 0 || result.Usage.Output <= 0 {
		t.Fatalf("Claude did not consume streamed usage: err=%v output=%q", err, stdout.String())
	}
	requestMu.Lock()
	mainCount, agentCount := mainRequests, agentRequests
	requestMu.Unlock()
	fixture.mu.Lock()
	threads, toolResults, children := fixture.threads, fixture.toolResults, fixture.children
	fixture.mu.Unlock()
	if mainCount < 2 || agentCount < 1 || threads < 2 || toolResults != 1 || children < 1 {
		t.Fatalf("incomplete roundtrip: main=%d agent=%d threads=%d toolResults=%d children=%d", mainCount, agentCount, threads, toolResults, children)
	}
	t.Logf("actual Claude Code -> production gateway -> App Server fixture: main=%d agent=%d threads=%d toolResults=%d children=%d result=MAIN_GATEWAY_OK", mainCount, agentCount, threads, toolResults, children)
}

type claudeGatewayPeer struct {
	mu                             sync.Mutex
	threads, toolResults, children int
	rejectedArguments              int
}

func startClaudeGatewayPeer(t *testing.T, home string, invalidAgentFirst ...bool) *claudeGatewayPeer {
	t.Helper()
	profile := filepath.Join(home, "gpt-appserver")
	if err := os.Mkdir(profile, 0700); err != nil {
		t.Fatal(err)
	}
	path := codexapp.SharedSocketPath(profile)
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = listener.Close()
		t.Fatal(err)
	}
	peer := &claudeGatewayPeer{}
	var connectionsMu sync.Mutex
	connections := map[*websocket.Conn]bool{}
	server := &http.Server{ReadHeaderTimeout: time.Second, Handler: websocket.Handler(func(ws *websocket.Conn) {
		connectionsMu.Lock()
		connections[ws] = true
		connectionsMu.Unlock()
		defer func() { _ = ws.Close(); connectionsMu.Lock(); delete(connections, ws); connectionsMu.Unlock() }()
		childThreads := map[string]bool{}
		turns := map[string]int{}
		pendingThread, pendingTurn := "", ""
		send := func(v any) bool { return websocket.JSON.Send(ws, v) == nil }
		usageTurns := map[string]int{}
		complete := func(thread, turn, text string) bool {
			usageTurns[thread]++
			count := usageTurns[thread]
			usage := map[string]any{"method": "thread/tokenUsage/updated", "params": map[string]any{"threadId": thread, "turnId": turn, "tokenUsage": map[string]any{"last": map[string]int{"inputTokens": 100, "cachedInputTokens": 20, "outputTokens": 10}, "total": map[string]int{"inputTokens": 100 * count, "cachedInputTokens": 20 * count, "outputTokens": 10 * count}}}}
			return send(map[string]any{"method": "item/agentMessage/delta", "params": map[string]any{"threadId": thread, "turnId": turn, "delta": text}}) && send(usage) && send(map[string]any{"method": "turn/completed", "params": map[string]any{"threadId": thread, "turn": map[string]string{"id": turn, "status": "completed"}}})
		}
		for {
			var raw []byte
			if err := websocket.Message.Receive(ws, &raw); err != nil {
				return
			}
			var message struct {
				ID     json.RawMessage            `json:"id"`
				Method string                     `json:"method"`
				Params map[string]json.RawMessage `json:"params"`
				Result json.RawMessage            `json:"result"`
			}
			if json.Unmarshal(raw, &message) != nil {
				return
			}
			if message.Method == "" && len(message.Result) > 0 {
				var feedback struct {
					Success bool `json:"success"`
				}
				if json.Unmarshal(message.Result, &feedback) != nil {
					return
				}
				if len(invalidAgentFirst) > 0 && invalidAgentFirst[0] && !feedback.Success {
					peer.mu.Lock()
					peer.rejectedArguments++
					peer.mu.Unlock()
					if !send(map[string]any{"id": "corrected-agent-rpc", "method": "item/tool/call", "params": map[string]any{"threadId": pendingThread, "turnId": pendingTurn, "callId": "corrected-agent-call", "tool": "Agent", "arguments": map[string]string{"description": "Inspect repository", "prompt": "Reply CHILD_GATEWAY_OK"}}}) {
						return
					}
					continue
				}
				peer.mu.Lock()
				peer.toolResults++
				peer.mu.Unlock()
				if pendingThread == "" || !complete(pendingThread, pendingTurn, "MAIN_GATEWAY_OK") {
					return
				}
				pendingThread, pendingTurn = "", ""
				continue
			}
			value := func(key string) string { var s string; _ = json.Unmarshal(message.Params[key], &s); return s }
			var result any
			switch message.Method {
			case "initialize":
				result = map[string]string{"userAgent": "claude-gateway-offline-peer"}
			case "initialized":
				continue
			case "account/read":
				result = map[string]any{"account": map[string]string{"type": "chatgpt", "planType": "test"}}
			case "thread/start":
				peer.mu.Lock()
				peer.threads++
				thread := fmt.Sprintf("thread-%d", peer.threads)
				peer.mu.Unlock()
				result = map[string]any{"thread": map[string]string{"id": thread}}
			case "turn/start":
				thread := value("threadId")
				var mode struct {
					Settings struct {
						DeveloperInstructions string `json:"developer_instructions"`
					} `json:"settings"`
				}
				if err := json.Unmarshal(message.Params["collaborationMode"], &mode); err != nil {
					t.Errorf("turn collaboration mode: %v", err)
					return
				}
				childThreads[thread] = strings.Contains(mode.Settings.DeveloperInstructions, "MOAI_CHILD_PROBE")
				turns[thread]++
				turn := fmt.Sprintf("turn-%s-%d", thread, turns[thread])
				if !send(map[string]any{"id": message.ID, "result": map[string]any{"turn": map[string]string{"id": turn}}}) {
					return
				}
				if childThreads[thread] {
					peer.mu.Lock()
					peer.children++
					peer.mu.Unlock()
					if !complete(thread, turn, "CHILD_GATEWAY_OK") {
						return
					}
				} else if turns[thread] == 1 {
					pendingThread, pendingTurn = thread, turn
					if len(invalidAgentFirst) > 0 && invalidAgentFirst[0] {
						if !send(map[string]any{"method": "item/agentMessage/delta", "params": map[string]any{"threadId": thread, "turnId": turn, "delta": "PRE_TOOL_PROGRESS"}}) {
							return
						}
						if !send(map[string]any{"id": "invalid-agent-rpc", "method": "item/tool/call", "params": map[string]any{"threadId": thread, "turnId": turn, "callId": "missing-description", "tool": "Agent", "arguments": map[string]string{"prompt": "Reply CHILD_GATEWAY_OK"}}}) {
							return
						}
						continue
					}
					if !send(map[string]any{"id": "native-agent-rpc", "method": "item/tool/call", "params": map[string]any{"threadId": thread, "turnId": turn, "callId": "agent-call", "tool": "Agent", "arguments": map[string]string{"description": "local child", "prompt": "Reply CHILD_GATEWAY_OK", "subagent_type": "probe"}}}) {
						return
					}
				} else if !complete(thread, turn, "MAIN_GATEWAY_OK") {
					return
				}
				continue
			case "thread/resume":
				result = map[string]any{"thread": map[string]string{"id": value("threadId")}}
			case "turn/interrupt":
				result = map[string]any{}
			default:
				return
			}
			if !send(map[string]any{"id": message.ID, "result": result}) {
				return
			}
		}
	})}
	done := make(chan struct{})
	go func() { defer close(done); _ = server.Serve(listener) }()
	t.Cleanup(func() {
		_ = server.Close()
		connectionsMu.Lock()
		for ws := range connections {
			_ = ws.Close()
		}
		connectionsMu.Unlock()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("offline peer did not stop")
		}
	})
	return peer
}
