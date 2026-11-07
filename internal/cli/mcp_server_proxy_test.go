package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestOneShot_FakeToolResultRoundTrips(t *testing.T) {
	s := server.NewMCPServer("fake", "0")
	s.AddTool(mcp.NewTool("echo_tool", mcp.WithString("msg")), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return toolJSON("echo_tool", map[string]any{"msg": req.GetString("msg", ""), "n": 2}), nil
	})
	var out bytes.Buffer
	if err := callMCPToolOnce(context.Background(), s, &out, "echo_tool", map[string]any{"msg": "hi"}); err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(bytes.TrimSpace(out.Bytes()))
	res, err := mcp.ParseCallToolResult(&raw)
	if err != nil {
		t.Fatalf("decode %q: %v", out.String(), err)
	}
	m, _ := res.StructuredContent.(map[string]any)
	if res.IsError || m["msg"] != "hi" {
		t.Fatalf("round trip lost the payload: %s", out.String())
	}
}

func TestOneShot_UnknownToolAndBadJSONAreCommandErrors(t *testing.T) {
	t.Setenv(mcpOneShotEnv, "")
	var out bytes.Buffer
	if err := runMCPServerOneShot(context.Background(), strings.NewReader("{}"), &out, "no_such_tool", "-", ""); err == nil {
		t.Fatal("unknown tool must be a command error")
	}
	if err := runMCPServerOneShot(context.Background(), strings.NewReader("{"), &out, "goal_status", "-", ""); err == nil {
		t.Fatal("malformed input JSON must be a command error")
	}
}

func TestOneShot_CobraFlagsAreHidden(t *testing.T) {
	cmd := newMCPServerCmd()
	for _, name := range []string{"call", "input-json"} {
		f := cmd.Flags().Lookup(name)
		if f == nil || !f.Hidden {
			t.Fatalf("flag --%s must exist and be hidden", name)
		}
	}
}

// stubInstalledBuild writes a /bin/sh stub that records its argv and stdin and
// prints canned (stdout) result JSON; it returns the stub path and the record path.
func stubInstalledBuild(t *testing.T, cannedStdout string, exitCode int) (stub, record string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("sh stub")
	}
	dir := t.TempDir()
	stub = filepath.Join(dir, "moai")
	record = filepath.Join(dir, "record")
	canned := filepath.Join(dir, "canned.json")
	if err := os.WriteFile(canned, []byte(cannedStdout), 0o644); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\necho \"$@\" > " + record + "\ncat >> " + record + "\ncat " + canned + "\nexit " + string(rune('0'+exitCode)) + "\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return stub, record
}

func staleAgainst(t *testing.T, installed string) {
	t.Helper()
	// Captured identity differs in size from the stub, so the server reads stale.
	withStalenessSeams(t, binaryIdentity{Size: 1, ModTime: time.Now().Add(-time.Hour), Captured: true}, installed)
	t.Setenv(mcpOneShotEnv, "")
}

func guardedLocal(called *bool) server.ToolHandlerFunc {
	return func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		*called = true
		return toolJSON("local", map[string]any{"from": "local"}), nil
	}
}

func TestStalenessGuard_WriteCallProxiesToInstalledBuild(t *testing.T) {
	stub, record := stubInstalledBuild(t, `{"content":[{"type":"text","text":"{\"from\":\"installed\"}"}],"structuredContent":{"from":"installed"}}`, 0)
	staleAgainst(t, stub)
	var local bool
	var req mcp.CallToolRequest
	req.Params.Arguments = map[string]any{"condition": "x"}
	res, err := mcpStalenessGuard("goal_arm", guardedLocal(&local))(context.Background(), req)
	if err != nil || res == nil {
		t.Fatalf("res=%v err=%v", res, err)
	}
	m, _ := res.StructuredContent.(map[string]any)
	if local || m["from"] != "installed" || m["stale_server"] != true || res.IsError {
		t.Fatalf("expected the installed build's result with stale markers, local=%v got %+v", local, m)
	}
	var notes string
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			notes += tc.Text + "\n"
		}
	}
	if !strings.Contains(notes, "proxied to the installed build") {
		t.Fatalf("missing proxied note: %s", notes)
	}
	rec, _ := os.ReadFile(record)
	if !strings.Contains(string(rec), "mcp-server --call goal_arm --input-json -") || !strings.Contains(string(rec), `"condition":"x"`) {
		t.Fatalf("stub did not receive the call and arguments: %s", rec)
	}
}

func TestStalenessGuard_ReadCallServedLocally(t *testing.T) {
	stub, record := stubInstalledBuild(t, `{}`, 0)
	staleAgainst(t, stub)
	var local bool
	res, err := mcpStalenessGuard("goal_status", guardedLocal(&local))(context.Background(), mcp.CallToolRequest{})
	if err != nil || res == nil || !local {
		t.Fatalf("read call must be served locally: local=%v res=%v err=%v", local, res, err)
	}
	if _, statErr := os.Stat(record); statErr == nil {
		t.Fatal("read call must not exec the installed build")
	}
	if m, _ := res.StructuredContent.(map[string]any); m["stale_server"] != true {
		t.Fatalf("stale metadata missing on a local read: %+v", m)
	}
}

func TestStalenessGuard_ProxyFailureKeepsRefusal(t *testing.T) {
	stub, _ := stubInstalledBuild(t, `not json`, 0)
	staleAgainst(t, stub)
	var local bool
	res, _ := mcpStalenessGuard("goal_arm", guardedLocal(&local))(context.Background(), mcp.CallToolRequest{})
	if local || res == nil || !res.IsError {
		t.Fatalf("decode failure must fail closed: local=%v res=%+v", local, res)
	}
	var text string
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			text += tc.Text
		}
	}
	if !strings.Contains(text, "stale_server") || !strings.Contains(text, "proxy to the installed build failed") {
		t.Fatalf("refusal text missing: %s", text)
	}

	failing, _ := stubInstalledBuild(t, ``, 3)
	staleAgainst(t, failing)
	res, _ = mcpStalenessGuard("goal_arm", guardedLocal(&local))(context.Background(), mcp.CallToolRequest{})
	if res == nil || !res.IsError {
		t.Fatalf("exec failure must fail closed: %+v", res)
	}
}

func TestStalenessGuard_OneShotNeverProxies(t *testing.T) {
	stub, record := stubInstalledBuild(t, `{}`, 0)
	staleAgainst(t, stub)
	t.Setenv(mcpOneShotEnv, "1")
	var local bool
	res, _ := mcpStalenessGuard("goal_arm", guardedLocal(&local))(context.Background(), mcp.CallToolRequest{})
	if !local || res == nil {
		t.Fatalf("one-shot must serve locally: local=%v", local)
	}
	if _, err := os.Stat(record); err == nil {
		t.Fatal("one-shot must not exec again")
	}
}
