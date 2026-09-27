package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
)

func TestFactoryMCPIdentityAttribution(t *testing.T) {
	root := t.TempDir()
	s, err := factorymsg.Open(root, "run")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	pid := os.Getpid()
	fingerprint := "test-process-start"
	oldProbe := factoryProbeProcessIdentity
	factoryProbeProcessIdentity = func(gotPID int) (string, homestate.ProcessIdentityState) {
		if gotPID == pid {
			return fingerprint, homestate.ProcessIdentityLive
		}
		return "", homestate.ProcessIdentityDead
	}
	t.Cleanup(func() { factoryProbeProcessIdentity = oldProbe })
	peer, err := s.RegisterPeer(context.Background(), factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: "run", Backend: "codex", Role: "worker", Slot: "agent-1", SessionUUID: "owner-session", Generation: 1, PID: pid, ProcessStart: fingerprint})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(pid))
	t.Setenv(config.EnvClaudeCodeSessionID, "foreign-authoritative-session")
	if got, err := currentFactoryPeer(context.Background(), s); err == nil {
		t.Fatalf("authoritative lookup failure fell back to PID peer: %+v", got)
	}
	t.Setenv(config.EnvClaudeCodeSessionID, "")
	t.Setenv(config.EnvClaudeProjectDir, t.TempDir()) // no shared side-channel attribution
	got, err := currentFactoryPeer(context.Background(), s)
	if err != nil || got.SessionUUID != peer.SessionUUID || got.Generation != peer.Generation {
		t.Fatalf("owner attribution=%+v err=%v", got, err)
	}
}

func operationalStatusClient(t *testing.T, root, run string) *client.Client {
	t.Helper()
	t.Setenv(config.EnvClaudeProjectDir, root)
	r, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.RecordRun(context.Background(), homestate.FactoryRun{RunID: run, LeadSessionID: "leader", Backend: "codex", ManifestJSON: "{}"}); err != nil {
		t.Fatal(err)
	}
	_ = r.Close()
	c, err := client.NewInProcessClient(newMoaiMCPServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeInProcessClient(c) })
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatal(err)
	}
	return c
}

func factoryBrokerSnapshot(t *testing.T, root, run string) string {
	t.Helper()
	path, err := factorymsg.BrokerPath(root, run)
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var snapshot []any
	for _, query := range []string{"SELECT * FROM peers ORDER BY slot", "SELECT * FROM messages ORDER BY id", "SELECT * FROM dead_letters ORDER BY id"} {
		rows, err := db.Query(query)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			values := make([]any, len(columns))
			ptrs := make([]any, len(values))
			for i := range values {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			snapshot = append(snapshot, values)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		_ = rows.Close()
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestFactoryMsgStatusReadOnlyRoster(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "ops-readonly"
	c := operationalStatusClient(t, root, run)
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "factory_msg_status", Arguments: map[string]any{"run_id": run}}}
	// An active run without a broker must not be silently initialized by a read.
	res, err := c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("status initialized absent broker instead of preserving query failure")
	}
	path, err := factorymsg.BrokerPath(root, run)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("status created broker state")
	}
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	p := factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "codex", Role: "lead", Slot: "lead", SessionUUID: "own-session", Generation: 1, PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint()}
	p, err = s.RegisterPeer(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	w := p
	w.Slot = "agent-1"
	w.Role = "worker"
	w.SessionUUID = "worker-session"
	w, err = s.RegisterPeer(context.Background(), w)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Send(context.Background(), factorymsg.SendRequest{From: p, To: w, Kind: factorymsg.KindStatusRequest, IdempotencyKey: "status-read", TaskRef: "t1074", CorrelationID: "status", TTL: time.Hour, Payload: []byte("private-message")}); err != nil {
		t.Fatal(err)
	}
	foreign, err := factorymsg.Open(t.TempDir(), run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = foreign.Close() })
	foreignPeer := p
	foreignPeer.SessionUUID = "foreign-sentinel"
	foreignPeer.ProjectKey = "project"
	if _, err := foreign.RegisterPeer(context.Background(), foreignPeer); err != nil {
		t.Fatal(err)
	}
	before := factoryBrokerSnapshot(t, root, run)
	for range 2 {
		res, err = c.CallTool(context.Background(), request)
		if err != nil || res.IsError {
			t.Fatalf("status=%+v err=%v", res, err)
		}
		body, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "own-session") || !strings.Contains(string(body), "\"lanes\"") || strings.Contains(string(body), "foreign-sentinel") || strings.Contains(string(body), "private-message") {
			t.Fatalf("roster response=%s", body)
		}
		if after := factoryBrokerSnapshot(t, root, run); after != before {
			t.Fatal("read-only status changed persisted rows")
		}
	}
	request.Params.Arguments = map[string]any{"run_id": "not-active"}
	res, err = c.CallTool(context.Background(), request)
	if err != nil || !res.IsError {
		t.Fatalf("inactive run accepted: %v", err)
	}
}

func TestFactoryLeadNoticeUsesOperationalStatus(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	root, run := t.TempDir(), "ops-notice"
	c := operationalStatusClient(t, root, run)
	t.Setenv(config.EnvMoaiKanbanID, run)
	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvMoaiKanbanBackend, "codex")
	t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(os.Getpid()))
	start, state := homestate.ProbeProcessIdentity(os.Getpid())
	if state != homestate.ProcessIdentityLive || start == "" {
		t.Fatal("test process identity unavailable")
	}
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if _, err := s.RegisterLaunchPending(context.Background(), factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "codex",
		Role: "lead", Slot: "lead", PID: os.Getpid(), ProcessStart: start,
	}); err != nil {
		t.Fatal(err)
	}
	output, err := hook.NewSessionStartHandler(nil, hook.WithSynchronousDeferredScans()).Handle(context.Background(), &hook.HookInput{SessionID: "notice-lead", ProjectDir: root, CWD: root, Source: "startup"})
	if err != nil || output.HookSpecificOutput == nil {
		t.Fatalf("SessionStart: %v", err)
	}
	text := output.HookSpecificOutput.AdditionalContext
	if !strings.Contains(text, `factory_msg_status({"run_id":"`+run+`"})`) || strings.Contains(text, "factory_msg_list") {
		t.Fatalf("notice has no read-only operational call: %s", text)
	}
	registered, err := c.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tool := range registered.Tools {
		if tool.Name == "factory_msg_status" {
			found = true
		}
	}
	if !found {
		t.Fatal("notice names an unregistered tool")
	}
	res, err := c.CallTool(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "factory_msg_status", Arguments: map[string]any{"run_id": run}}})
	if err != nil || res.IsError {
		t.Fatalf("notice handler=%+v err=%v", res, err)
	}
	encoded, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), "notice-lead") {
		t.Fatalf("SessionStart peer missing: %s", encoded)
	}
}

func TestFactoryMCPIdentityFallsBackToDirectParent(t *testing.T) {
	root := t.TempDir()
	s, err := factorymsg.Open(root, "run")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	pid := os.Getppid()
	fingerprint := "parent-process-start"
	oldProbe := factoryProbeProcessIdentity
	factoryProbeProcessIdentity = func(gotPID int) (string, homestate.ProcessIdentityState) {
		if gotPID == pid {
			return fingerprint, homestate.ProcessIdentityLive
		}
		return "", homestate.ProcessIdentityDead
	}
	t.Cleanup(func() { factoryProbeProcessIdentity = oldProbe })
	peer, err := s.RegisterPeer(context.Background(), factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: "run", Backend: "codex", Role: "worker", Slot: "agent-1", SessionUUID: "parent-owned-session", Generation: 1, PID: pid, ProcessStart: fingerprint})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvMoaiSessionPID, "")
	t.Setenv(config.EnvClaudeCodeSessionID, "")
	t.Setenv(config.EnvClaudeProjectDir, t.TempDir())
	got, err := currentFactoryPeer(context.Background(), s)
	if err != nil || got.SessionUUID != peer.SessionUUID {
		t.Fatalf("parent attribution=%+v err=%v", got, err)
	}
}
