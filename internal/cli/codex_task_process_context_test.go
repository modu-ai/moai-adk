//go:build !windows

package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
)

// This fixture is an actual child process. Canned codexConn fixtures cannot
// observe exec.CommandContext killing the app-server when an MCP request ends.
func TestCodexTaskBackgroundOwnsProcessPastRequest(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	bin := filepath.Join(root, "codex-app-server")
	script := `#!/bin/sh
while IFS= read -r line; do
  case "$line" in
    *'"id":1'*) printf '%s\n' '{"id":1,"result":{}}' ;;
    *'"id":2'*) printf '%s\n' '{"id":2,"result":{"thread":{"id":"tid-process"}}}' ;;
    *'"id":3'*)
      sleep 0.1
      printf '%s\n' '{"id":3,"result":{"turn":{"id":"turn-process","status":"inProgress"}}}'
      printf '%s\n' '{"method":"turn/started","params":{"threadId":"tid-process","turn":{"id":"turn-process"}}}'
      printf '%s\n' '{"method":"item/completed","params":{"threadId":"tid-process","turnId":"turn-process","item":{"type":"agentMessage","id":"m1","text":"process survived"}}}'
      printf '%s\n' '{"method":"turn/completed","params":{"threadId":"tid-process","turn":{"id":"turn-process","status":"completed"}}}'
      ;;
  esac
done
`
	if err := os.WriteFile(bin, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	prevLook, prevSession := codexLookPath, codexSession
	codexLookPath = func(string) (string, error) { return bin, nil }
	codexSession = realCodexSessionRunner{}
	t.Cleanup(func() { codexLookPath, codexSession = prevLook, prevSession })

	ctx, cancel := context.WithCancel(context.Background())
	res, err := handleCodexTask(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]any{"prompt": "say ready", "background": true},
	}})
	if err != nil {
		t.Fatal(err)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatal("no background job id")
	}
	cancel() // MCP handler returned; its request context is now over.

	start := time.Now()
	rec := awaitTerminalJob(t, newCodexJobRegistry(root), jobID)
	if rec.Status != codexJobStatusCompleted || !strings.Contains(rec.Output, "process survived") {
		t.Fatalf("status=%s output=%q error=%q after %s", rec.Status, rec.Output, rec.Error, time.Since(start))
	}
}

func TestCodexTaskServerExitStopsBackgroundProcess(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexTaskTimeout(t, 3*time.Second)
	bin := filepath.Join(root, "codex-app-server")
	script := `#!/bin/sh
while IFS= read -r line; do
  case "$line" in
    *'"id":1'*) printf '%s\n' '{"id":1,"result":{}}' ;;
    *'"id":2'*) printf '%s\n' '{"id":2,"result":{"thread":{"id":"tid-stall"}}}' ;;
    *'"id":3'*)
      printf '%s\n' '{"id":3,"result":{"turn":{"id":"turn-stall","status":"inProgress"}}}'
      printf '%s\n' '{"method":"turn/started","params":{"threadId":"tid-stall","turn":{"id":"turn-stall"}}}'
      ;;
  esac
done
`
	if err := os.WriteFile(bin, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	prevLook, prevSession := codexLookPath, codexSession
	codexLookPath = func(string) (string, error) { return bin, nil }
	codexSession = realCodexSessionRunner{}
	t.Cleanup(func() { codexLookPath, codexSession = prevLook, prevSession })

	res, err := handleCodexTask(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]any{"prompt": "stall", "background": true},
	}})
	if err != nil {
		t.Fatal(err)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatal("no background job id")
	}
	stopped := make(chan struct{})
	go func() { stopCodexBackgroundJobs(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("server shutdown did not stop the background process")
	}
	rec := awaitTerminalJob(t, newCodexJobRegistry(root), jobID)
	if rec.Status != codexJobStatusFailed {
		t.Fatalf("status=%s error=%q, want failed after shutdown", rec.Status, rec.Error)
	}
	if _, live := codexLiveJobSessions.Load(jobID); live {
		t.Fatal("session remained live after shutdown")
	}
}

func TestCodexTaskBackgroundProcessBound(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexTaskTimeout(t, 750*time.Millisecond)
	bin := filepath.Join(root, "codex-app-server")
	script := `#!/bin/sh
while IFS= read -r line; do
  case "$line" in
    *'"id":1'*) printf '%s\n' '{"id":1,"result":{}}' ;;
    *'"id":2'*) printf '%s\n' '{"id":2,"result":{"thread":{"id":"tid-bound"}}}' ;;
    *'"id":3'*) printf '%s\n' '{"id":3,"result":{"turn":{"id":"turn-bound","status":"inProgress"}}}' ;;
  esac
done
`
	if err := os.WriteFile(bin, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	prevLook, prevSession := codexLookPath, codexSession
	codexLookPath = func(string) (string, error) { return bin, nil }
	codexSession = realCodexSessionRunner{}
	t.Cleanup(func() { codexLookPath, codexSession = prevLook, prevSession })

	res, err := handleCodexTask(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{
		Arguments: map[string]any{"prompt": "stall", "background": true},
	}})
	if err != nil {
		t.Fatal(err)
	}
	jobID, _ := structuredMap(t, res)["job_id"].(string)
	if jobID == "" {
		t.Fatal("no background job id")
	}
	start := time.Now()
	rec := awaitTerminalJob(t, newCodexJobRegistry(root), jobID)
	if rec.Status != codexJobStatusFailed || !strings.Contains(rec.Error, "timed out after "+config.DefaultCodexTaskTimeout.String()) {
		t.Fatalf("status=%s error=%q elapsed=%s", rec.Status, rec.Error, time.Since(start))
	}
}
