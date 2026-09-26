package cli

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryMsgCodexWakeLive is an opt-in gate for an already-running Codex
// TUI with the MoAI MCP configured. It verifies that an MCP send wakes that
// session and that the recipient, without terminal input, claims and receipts
// the persisted message. The target must be an isolated test session.
func TestFactoryMsgCodexWakeLive(t *testing.T) {
	thread := os.Getenv("MOAI_FACTORY_LIVE_WORKER_THREAD")
	pidText := os.Getenv("MOAI_FACTORY_LIVE_WORKER_PID")
	root := os.Getenv("MOAI_FACTORY_LIVE_PROJECT_ROOT")
	if thread == "" || pidText == "" || root == "" {
		t.Skip("set MOAI_FACTORY_LIVE_WORKER_THREAD, _PID, and _PROJECT_ROOT for the live gate")
	}
	pid, err := strconv.Atoi(pidText)
	if err != nil {
		t.Fatal(err)
	}
	start, state := homestate.ProbeProcessIdentity(pid)
	if state != homestate.ProcessIdentityLive {
		t.Fatalf("Codex recipient PID %d is not live", pid)
	}
	const run = "pushprobe"
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv(config.EnvClaudeCodeSessionID, "live-test-lead")
	factoryDB, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := factoryDB.RecordRun(context.Background(), homestate.FactoryRun{RunID: run, LeadSessionID: "live-test-lead", Backend: "glm", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	if err := factoryDB.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	lead := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "glm", Role: "lead", Slot: "lead", SessionUUID: "live-test-lead", Generation: 1, PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint()}
	worker := factorymsg.Peer{ProjectKey: lead.ProjectKey, RunID: run, Backend: "codex", Role: "worker", Slot: "worker-1", SessionUUID: thread, Generation: 1, PID: pid, ProcessStart: start}
	for _, peer := range []factorymsg.Peer{lead, worker} {
		if _, err := s.RegisterPeer(context.Background(), peer); err != nil {
			t.Fatal(err)
		}
	}
	before, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	c, err := client.NewInProcessClient(newMoaiMCPServer())
	if err != nil {
		t.Fatal(err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatal(err)
	}
	key := "live-wake-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if mcpBinary := os.Getenv("MOAI_FACTORY_LIVE_GLM_MCP_BINARY"); mcpBinary != "" {
		launcher := os.Getenv("MOAI_FACTORY_LIVE_GLM_LAUNCHER")
		if launcher == "" {
			t.Fatal("set MOAI_FACTORY_LIVE_GLM_LAUNCHER with _GLM_MCP_BINARY")
		}
		mcpConfig := map[string]any{"mcpServers": map[string]any{"moai": map[string]any{"command": mcpBinary, "args": []string{"mcp-server"}, "env": map[string]string{
			"MOAI_HOME": os.Getenv("MOAI_HOME"), "CLAUDE_PROJECT_DIR": root,
			"MOAI_SESSION_PID": strconv.Itoa(os.Getpid()), "CLAUDE_CODE_SESSION_ID": lead.SessionUUID,
			"MOAI_KANBAN_ID": run, "MOAI_KANBAN_BACKEND": "glm", "MOAI_FACTORY_WORKERS": "1",
		}}}}
		configBytes, err := json.Marshal(mcpConfig)
		if err != nil {
			t.Fatal(err)
		}
		configPath := filepath.Join(t.TempDir(), "mcp.json")
		if err := os.WriteFile(configPath, configBytes, 0600); err != nil {
			t.Fatal(err)
		}
		prompt := "Call the moai factory_msg_send MCP tool exactly once with run_id=pushprobe, to_slot=worker-1, kind=status_request, idempotency_key=" + key + ", task_ref=t1, correlation_id=" + key + ", ttl_seconds=120, body=LIVE_GLM_TO_CODEX_RECEIVED. Report the tool result briefly."
		callCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(callCtx, launcher, "glm", "--", "-p", "--dangerously-skip-permissions", "--mcp-config", configPath, "--strict-mcp-config", prompt)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), config.EnvMoaiSessionPID+"="+strconv.Itoa(os.Getpid()), config.EnvMoaiKanbanID+"="+run, config.EnvMoaiKanbanBackend+"=glm", config.EnvClaudeProjectDir+"="+root)
		out, callErr := cmd.CombinedOutput()
		if callErr != nil {
			t.Fatalf("GLM MCP send failed: %v, output=%q", callErr, string(out))
		}
		logOutput := string(out)
		if key := os.Getenv("GLM_API_KEY"); key != "" {
			logOutput = strings.ReplaceAll(logOutput, key, "[redacted]")
		}
		if len(logOutput) > 1600 {
			logOutput = logOutput[len(logOutput)-1600:]
		}
		t.Logf("GLM MCP send output tail: %s", logOutput)
	} else {
		res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "factory_msg_send", Arguments: map[string]any{
			"run_id": run, "to_slot": worker.Slot, "kind": factorymsg.KindStatusRequest,
			"idempotency_key": key, "task_ref": "t1", "correlation_id": key,
			"ttl_seconds": 120, "body": "LIVE_MCP_PUSH_RECEIVED; report this marker and receipt the message",
		}}})
		if err != nil || res.IsError {
			t.Fatalf("send=%+v err=%v", res, err)
		}
		t.Logf("MCP send result: %+v", res.StructuredContent)
	}
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		status, err := s.Status(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if status.Acknowledged > before.Acknowledged {
			t.Logf("broker acknowledged=%d pending=%d claimed=%d", status.Acknowledged, status.Pending, status.Claimed)
			return
		}
		time.Sleep(time.Second)
	}
	status, _ := s.Status(context.Background())
	t.Fatalf("Codex did not receipt pushed message: %+v", status)
}
