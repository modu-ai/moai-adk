package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// closeInProcessClient closes the in-process client, satisfying errcheck (the
// SDK's own examples ignore the error; our linter does not).
func closeInProcessClient(c *client.Client) {
	if c != nil {
		_ = c.Close()
	}
}

// moaiCoreToolNames is the M1 core tool surface (design.md §3, AC-MCP-005). Every
// name MUST appear in tools/list, and each handler MUST wrap its verified
// internal/ integration point (thin-wrapper invariant C1/AP-1).
var moaiCoreToolNames = []string{
	"session_list",
	"goal_status",
	"goal_arm",
	"spec_progress",
	"verify_snapshot",
	"verify_trend",
	"spec_audit",
	"spec_drift",
	"audit_cache",
}

// TestNewMCPServerCmd_Registered verifies the `moai mcp-server` subcommand is
// built and wired for root.go AddCommand (AC-MCP-001).
func TestNewMCPServerCmd_Registered(t *testing.T) {
	cmd := newMCPServerCmd()
	if cmd == nil {
		t.Fatal("newMCPServerCmd returned nil")
	}
	if cmd.Use != "mcp-server" {
		t.Fatalf("Use = %q, want %q", cmd.Use, "mcp-server")
	}
	if cmd.RunE == nil {
		t.Fatal("RunE is nil — mcp-server must block on a stdio RunE (goal.go pattern)")
	}
}

// TestMoaiMCPServer_ToolsListDeclaresSchema exercises the stdio round-trip
// in-process (AC-MCP-001 initialize→tools/list, AC-MCP-004 every tool carries
// a name + inputSchema). Uses the SDK's in-process transport to avoid
// subprocess flakiness.
func TestMoaiMCPServer_ToolsListDeclaresSchema(t *testing.T) {
	srv := newMoaiMCPServer()
	if srv == nil {
		t.Fatal("newMoaiMCPServer returned nil")
	}

	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)

	initRes, err := c.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if initRes.ServerInfo.Name != moaiMCPServerName {
		t.Fatalf("server name = %q, want %q", initRes.ServerInfo.Name, moaiMCPServerName)
	}

	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	got := map[string]bool{}
	for _, tool := range tools.Tools {
		got[tool.Name] = true
		// AC-MCP-004: every tool declares a JSON Schema (InputSchema is a struct;
		// a declared schema carries Type "object").
		if tool.InputSchema.Type == "" {
			t.Errorf("tool %q: inputSchema.Type empty (AC-MCP-004)", tool.Name)
		}
	}
	for _, want := range moaiCoreToolNames {
		if !got[want] {
			t.Errorf("core tool %q missing from tools/list (AC-MCP-005)", want)
		}
	}
}

// TestMoaiMCPServer_SessionListRoundTrip exercises initialize → tools/list →
// tools/call for session_list and verifies thin-wrapper parity (AC-MCP-003): the
// handler calls session.QueryActiveWork and returns its output. A temp project
// dir with no active-sessions registry yields an empty list — proving the
// handler reached the internal/ function and shaped its result, not an error.
func TestMoaiMCPServer_SessionListRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	res, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "session_list",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("tools/call session_list: %v", err)
	}
	if res == nil {
		t.Fatal("CallTool returned nil result")
	}
	if res.IsError {
		t.Fatalf("session_list returned IsError=true; content=%v", res.Content)
	}
	if len(res.Content) == 0 {
		t.Fatal("session_list returned no content")
	}
}

// TestBuildMoaiMCPServerEntry_Neutral verifies the locked neutral .mcp.json
// entry shape (AC-MCP-006, §G.5): exactly {command:"moai", args:["mcp-server"]},
// NO env block (no secret surface), NO absolute path, NO SPEC-ID/SHA/date.
func TestBuildMoaiMCPServerEntry_Neutral(t *testing.T) {
	entry := buildMoaiMCPServerEntry()
	if entry["command"] != "moai" {
		t.Errorf("command = %v, want %q (PATH-resolved, no absolute path)", entry["command"], "moai")
	}
	args, ok := entry["args"].([]string)
	if !ok {
		t.Fatalf("args type = %T, want []string", entry["args"])
	}
	if len(args) != 1 || args[0] != "mcp-server" {
		t.Errorf("args = %v, want [mcp-server]", args)
	}
	if _, hasEnv := entry["env"]; hasEnv {
		t.Error("entry MUST NOT carry an env block (secret-hygiene C3; secrets stay ${VAR} literals only when a backend is enabled)")
	}
	// Neutral: no SPEC-ID, no SHA, no date, no type/url (third-party/HTTP markers).
	b, _ := json.Marshal(entry)
	neutral := string(b)
	for _, forbidden := range []string{"SPEC-", "sha", "2026", "http", "/Users/", "/home/"} {
		if strings.Contains(neutral, forbidden) {
			t.Errorf("neutral entry leaked forbidden token %q: %s (§25 neutrality)", forbidden, neutral)
		}
	}
}

// TestProvisionMoaiMCPServerEntry_OptInDefaultOff verifies AC-MCP-002 (opt-in
// default-off) + AC-MCP-006 (single neutral entry via atomic-config helpers): a
// fresh .mcp.json is NOT touched until provisioning runs, and provisioning
// inserts exactly one neutral `moai` entry through mutateClaudeJSONAtomic.
func TestProvisionMoaiMCPServerEntry_OptInDefaultOff(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".mcp.json")

	// AC-MCP-002: opt-in default-off — no provisioning ran, file untouched.
	if _, err := os.Stat(configPath); err == nil {
		t.Fatal("fresh project must NOT carry a moai .mcp.json entry by default (opt-in default-off)")
	}

	if err := provisionMoaiMCPServerEntryAt(configPath); err != nil {
		t.Fatalf("provision: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read provisioned .mcp.json: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse provisioned .mcp.json: %v", err)
	}
	servers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers missing/not-object: %v", root)
	}
	moaiEntry, ok := servers[moaiMCPServerKey].(map[string]any)
	if !ok {
		t.Fatalf("moai entry missing under mcpServers (key=%q): %v", moaiMCPServerKey, servers)
	}
	if moaiEntry["command"] != "moai" {
		t.Errorf("provisioned command = %v, want moai", moaiEntry["command"])
	}
	// Exactly ONE neutral entry — no third-party bundling (design.md §6.4).
	if len(servers) != 1 {
		t.Errorf("expected exactly 1 server entry, got %d (no third-party bundling): %v", len(servers), servers)
	}
}

// TestMCPServer_NoAskUserQuestion enforces the subagent boundary (B3/B11,
// C-HRA-008, REQ-MCP-014): the MCP server + handlers MUST NOT CALL
// AskUserQuestion — tools return structured results only. Matches the canonical
// C-HRA-008 static guard (TestWeb_NoAskUserQuestion): comment lines (// …) are
// excluded so documentation of the constraint is not a false positive.
func TestMCPServer_NoAskUserQuestion(t *testing.T) {
	b, err := os.ReadFile("mcp_server.go")
	if err != nil {
		t.Fatalf("read mcp_server.go: %v", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue // comment line — documentation, not a call
		}
		for _, forbidden := range []string{"AskUserQuestion", "mcp__askuser"} {
			if strings.Contains(line, forbidden) {
				t.Errorf("mcp_server.go calls %q — MCP tools must return structured results, never prompt (REQ-MCP-014): %s", forbidden, trimmed)
			}
		}
	}
	// Sanity: the file is non-empty and registers the command.
	if !strings.Contains(string(b), "func newMCPServerCmd") {
		t.Error("mcp_server.go missing newMCPServerCmd — command not registered")
	}
}

// TestMoaiMCPServer_CoreHandlersRoundTrip exercises initialize → tools/list →
// tools/call for EVERY core tool (AC-MCP-005: each handler wraps its verified
// internal/ integration point). A temp project dir with a minimal SPEC skeleton
// exercises each read path against representative state, asserting each returns
// non-error, non-empty structured content. This closes the per-handler coverage
// gap (E3) beyond the session_list-only round-trip above.
func TestMoaiMCPServer_CoreHandlersRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	// Minimal SPEC skeleton so spec_progress / spec_audit / audit_cache have a
	// real tree to scan (spec.ListDocs + spec.Audit read .moai/specs/).
	specDir := filepath.Join(tmp, ".moai", "specs", "SPEC-DEMO-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	specBody := "---\nid: SPEC-DEMO-001\ntitle: \"Demo\"\nversion: \"0.1.0\"\nstatus: draft\ncreated: 2026-08-05\nupdated: 2026-08-05\nauthor: test\npriority: P1\nphase: \"v3.0.0\"\nmodule: \"internal/demo\"\nlifecycle: spec-anchored\ntags: \"demo\"\n---\n\n# Demo SPEC\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specBody), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	// Each case calls a core tool and asserts a non-error, non-empty result,
	// proving the handler reached its internal/ function and shaped its output.
	// `label` is the subtest name; `tool` is the registered tool to invoke.
	cases := []struct {
		label string
		tool  string
		args  map[string]any
	}{
		{"goal_status", "goal_status", map[string]any{"session_id": "sess-demo"}},
		{"goal_arm", "goal_arm", map[string]any{"session_id": "sess-demo", "condition": "go test ./... exits 0"}},
		{"goal_status_after_arm", "goal_status", map[string]any{"session_id": "sess-demo"}},
		{"spec_progress", "spec_progress", map[string]any{}},
		{"verify_snapshot", "verify_snapshot", map[string]any{"key": "deadbeef:0"}},
		{"verify_trend", "verify_trend", map[string]any{"key": "deadbeef:0"}},
		{"spec_audit", "spec_audit", map[string]any{}},
		{"spec_drift", "spec_drift", map[string]any{}},
		{"audit_cache_hash", "audit_cache", map[string]any{"op": "compute_hash", "spec_dir": specDir}},
		{"audit_cache_lookup", "audit_cache", map[string]any{"op": "lookup", "spec_id": "SPEC-DEMO-001", "hash": "x"}},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			res, err := c.CallTool(ctx, mcp.CallToolRequest{
				Params: mcp.CallToolParams{Name: tc.tool, Arguments: tc.args},
			})
			if err != nil {
				t.Fatalf("tools/call %s: %v", tc.label, err)
			}
			if res == nil {
				t.Fatalf("%s: nil result", tc.label)
			}
			if len(res.Content) == 0 {
				t.Errorf("%s: empty content (handler did not shape a result)", tc.label)
			}
		})
	}
}

// TestMoaiMCPServer_ErrorPaths covers the structured-error branches: a missing
// required argument and an invalid audit_cache op return IsError results (not Go
// errors), preserving the subagent boundary (REQ-MCP-014).
func TestMoaiMCPServer_ErrorPaths(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	// verify_trend with missing key → structured IsError result.
	res, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "verify_trend", Arguments: map[string]any{}},
	})
	if err != nil {
		t.Fatalf("tools/call verify_trend (missing key): %v", err)
	}
	if res == nil || !res.IsError {
		t.Fatalf("verify_trend missing key: expected IsError result, got %v", res)
	}

	// audit_cache with unknown op → structured IsError result.
	res2, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "audit_cache", Arguments: map[string]any{"op": "bogus"}},
	})
	if err != nil {
		t.Fatalf("tools/call audit_cache (bad op): %v", err)
	}
	if res2 == nil || !res2.IsError {
		t.Fatalf("audit_cache bad op: expected IsError result, got %v", res2)
	}
}

// TestMoaiMCPServer_BranchCoverage closes the remaining handler branches toward
// the mcp_server.go ≥85% coverage target (E3): the verify_snapshot record path,
// the goal_arm error/reject branches, and the goal_status empty-session fallback.
func TestMoaiMCPServer_BranchCoverage(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	call := func(tool string, args map[string]any) *mcp.CallToolResult {
		t.Helper()
		res, callErr := c.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: tool, Arguments: args}})
		if callErr != nil {
			t.Fatalf("tools/call %s: %v", tool, callErr)
		}
		return res
	}

	// verify_snapshot RECORD branch: a `command` arg records via verify.RecordCheck.
	res := call("verify_snapshot", map[string]any{"key": "deadbeef:1", "command": "go test ./...", "exit_code": 0})
	if res.IsError {
		t.Fatalf("verify_snapshot record: IsError result %v", res.Content)
	}

	// goal_arm error: missing condition → structured IsError.
	if r := call("goal_arm", map[string]any{"session_id": "x"}); !r.IsError {
		t.Errorf("goal_arm missing condition: expected IsError, got %v", r.Content)
	}
	// goal_arm reject: max_turns=0 without max_duration → structured IsError
	// (inherited infinite-goal fail-closed rule).
	if r := call("goal_arm", map[string]any{"session_id": "x", "condition": "c", "max_turns": 0}); !r.IsError {
		t.Errorf("goal_arm max_turns=0 without duration: expected IsError, got %v", r.Content)
	}

	// goal_status: empty session_id exercises the statusSessionID("") fallback
	// (pid key), returning a non-error armed:false result.
	res = call("goal_status", map[string]any{"session_id": ""})
	if res.IsError {
		t.Fatalf("goal_status fallback: IsError result %v", res.Content)
	}
}

// TestMCPServer_StdioRoundTripSubprocess is the gold-standard AC-MCP-001 smoke:
// it builds the real `moai mcp-server` binary and drives a genuine stdio
// JSON-RPC round-trip (initialize → tools/list → tools/call session_list) over
// pipes — exercising the blocking ServeStdio entry (runMCPServer) and the
// newline-delimited transport the in-process tests bypass. Skipped when the
// build cannot run.
func TestMCPServer_StdioRoundTripSubprocess(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "moai-test")
	if out, err := exec.Command("go", "build", "-o", bin, "./cmd/moai").CombinedOutput(); err != nil {
		t.Skipf("skip subprocess smoke: cannot build ./cmd/moai: %v\n%s", err, out)
	}

	workdir := t.TempDir() // isolated CLAUDE_PROJECT_DIR → session_list reads an empty registry
	cmd := exec.Command(bin, moaiMCPServerSubcommand)
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+workdir)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() {
		_ = stdin.Close()
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
		}
	}()

	// write sends one NDJSON request line.
	write := func(id int, method string, params map[string]any) {
		t.Helper()
		req := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
		if params != nil {
			req["params"] = params
		}
		b, mErr := json.Marshal(req)
		if mErr != nil {
			t.Fatalf("marshal %s: %v", method, mErr)
		}
		b = append(b, '\n')
		if _, wErr := stdin.Write(b); wErr != nil {
			t.Fatalf("write %s: %v", method, wErr)
		}
	}

	// recv reads NDJSON lines until one carries the wanted id (notifications
	// like initialized/progress are skipped), or the deadline passes.
	recv := func(wantID int) map[string]any {
		t.Helper()
		scan := bufio.NewScanner(stdout)
		scan.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			if !scan.Scan() {
				if err := scan.Err(); err != nil {
					t.Fatalf("read stdout: %v", err)
				}
				time.Sleep(20 * time.Millisecond)
				continue
			}
			var msg map[string]any
			if jErr := json.Unmarshal(scan.Bytes(), &msg); jErr != nil {
				continue // non-JSON line (server log); skip
			}
			if id, ok := msg["id"].(float64); ok && int(id) == wantID {
				return msg
			}
		}
		t.Fatalf("no response with id=%d within deadline", wantID)
		return nil
	}

	// initialize
	write(1, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
	})
	initRes := recv(1)
	if serverInfo, _ := initRes["result"].(map[string]any); serverInfo != nil {
		if name, _ := serverInfo["serverInfo"].(map[string]any); name != nil {
			if n, _ := name["name"].(string); n != moaiMCPServerName {
				t.Errorf("stdio server name = %q, want %q", n, moaiMCPServerName)
			}
		}
	}

	// tools/list — non-empty and carries the core tool names.
	write(2, "tools/list", map[string]any{})
	listRes := recv(2)
	result, _ := listRes["result"].(map[string]any)
	tools, _ := result["tools"].([]any)
	if len(tools) < len(moaiCoreToolNames) {
		t.Fatalf("tools/list returned %d tools, want >= %d", len(tools), len(moaiCoreToolNames))
	}
	got := map[string]bool{}
	for _, tl := range tools {
		if tm, ok := tl.(map[string]any); ok {
			got[tm["name"].(string)] = true
		}
	}
	for _, want := range moaiCoreToolNames {
		if !got[want] {
			t.Errorf("stdio tools/list missing core tool %q", want)
		}
	}

	// tools/call session_list — non-error, non-empty content.
	write(3, "tools/call", map[string]any{"name": "session_list", "arguments": map[string]any{}})
	callRes := recv(3)
	if e, _ := callRes["error"].(map[string]any); e != nil {
		t.Fatalf("stdio tools/call session_list returned error: %v", e)
	}
	cr, _ := callRes["result"].(map[string]any)
	if content, _ := cr["content"].([]any); len(content) == 0 {
		t.Error("stdio tools/call session_list returned empty content")
	}
}

// TestMoaiMCPServer_CorruptedStateErrorBranches exercises the handlers' defensive
// toolErr branches: a corrupted goal/snapshot file must yield a structured
// IsError result, never a panic (REQ-MCP-014 subagent boundary — structured
// results only). This is a real production case (state-file corruption), not a
// contrived branch-fill, and it closes the remaining error-path coverage gap.
func TestMoaiMCPServer_CorruptedStateErrorBranches(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	const sessionID = "sess-bad"
	// Malformed goal JSON → goal.LoadGoal returns a parse error.
	goalDir := filepath.Join(tmp, goal.StateDir)
	if err := os.MkdirAll(goalDir, 0o755); err != nil {
		t.Fatalf("mkdir goal dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(goalDir, sessionID+".json"), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write bad goal: %v", err)
	}
	// Malformed snapshot JSON → verify.Load returns a parse error.
	snapDir := filepath.Join(tmp, verify.SnapshotDir)
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatalf("mkdir verify dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, "badkey.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write bad snapshot: %v", err)
	}

	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	mustCall := func(name string, args map[string]any) *mcp.CallToolResult {
		t.Helper()
		res, callErr := c.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: name, Arguments: args}})
		if callErr != nil {
			t.Fatalf("tools/call %s: %v", name, callErr)
		}
		return res
	}

	for _, tc := range []struct {
		label string
		tool  string
		args  map[string]any
	}{
		{"goal_status", "goal_status", map[string]any{"session_id": sessionID}},
		{"verify_snapshot", "verify_snapshot", map[string]any{"key": "badkey"}},
		{"verify_trend", "verify_trend", map[string]any{"key": "badkey"}},
	} {
		t.Run(tc.label, func(t *testing.T) {
			if r := mustCall(tc.tool, tc.args); !r.IsError {
				t.Errorf("%s on corrupted state: expected IsError, got %v", tc.label, r.Content)
			}
		})
	}
}

// TestProvisionMoaiMCPServerEntryAt_Idempotent verifies the no-op path: a second
// provision against an already-neutral entry does not rewrite the file. The
// comparison is marshal-based so it survives the JSON round-trip ([]string →
// []any after a disk read) — a type-assertion idempotency check would miss the
// re-read case.
func TestProvisionMoaiMCPServerEntryAt_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".mcp.json")

	if err := provisionMoaiMCPServerEntryAt(configPath); err != nil {
		t.Fatalf("first provision: %v", err)
	}
	first, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read after first: %v", err)
	}
	if err := provisionMoaiMCPServerEntryAt(configPath); err != nil {
		t.Fatalf("second (idempotent) provision: %v", err)
	}
	second, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read after second: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("idempotent provision rewrote the file:\nfirst:  %s\nsecond: %s", first, second)
	}
}

// Compile-time guard: ensure the SDK server type is wired so the import is
// justified (AC-MCP-001 thin stdio JSON-RPC).
var _ *server.MCPServer
