//go:build !windows

package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
)

// The production stdio entry point is driven through pipes so closing stdin
// exercises StdioServer.Listen's worker shutdown and runMCPServer's job cleanup.
type codexTaskMCPStdio struct {
	in       *io.PipeWriter
	messages chan map[string]json.RawMessage
	done     chan error
	exited   chan struct{}
}

type mcpReaderFunc func([]byte) (int, error)

func (f mcpReaderFunc) Read(p []byte) (int, error) { return f(p) }

func TestMCPEOFReaderPreservesFinalBytes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	payload := []byte("{\"id\":1}\n")
	sent := false
	reader := &mcpEOFReader{Reader: mcpReaderFunc(func(p []byte) (int, error) {
		if sent {
			return 0, io.EOF
		}
		sent = true
		return copy(p, payload), io.EOF // legal n > 0, err == EOF shape
	}), cancel: cancel}
	buf := make([]byte, 32)
	n, err := reader.Read(buf)
	if err != nil || string(buf[:n]) != string(payload) {
		t.Fatalf("final data read = (%q, %v), want (%q, nil)", buf[:n], err, payload)
	}
	if ctx.Err() != nil {
		t.Fatal("reader cancelled before delivering the final bytes")
	}
	n, err = reader.Read(buf)
	if n != 0 || err != io.EOF || ctx.Err() != context.Canceled {
		t.Fatalf("next read = (%d, %v), context=%v; want EOF and cancellation", n, err, ctx.Err())
	}
}

func startCodexTaskMCPStdio(t *testing.T, root string) *codexTaskMCPStdio {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	f := &codexTaskMCPStdio{
		in: inWriter, messages: make(chan map[string]json.RawMessage, 16),
		done: make(chan error, 1), exited: make(chan struct{}),
	}
	go func() {
		f.done <- runMCPServerWithIO(inReader, outWriter)
		_ = outWriter.Close()
		close(f.exited)
	}()
	go func() {
		defer close(f.messages)
		scanner := bufio.NewScanner(outReader)
		scanner.Buffer(make([]byte, 0, 4096), 8*1024*1024)
		for scanner.Scan() {
			var message map[string]json.RawMessage
			if json.Unmarshal(scanner.Bytes(), &message) == nil {
				f.messages <- message
			}
		}
	}()
	t.Cleanup(func() {
		_ = inWriter.Close()
		select {
		case <-f.exited:
		case <-time.After(5 * time.Second):
			t.Error("MCP stdio server remained live after test cleanup")
		}
		_ = outReader.Close()
	})
	f.send(t, 1, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "codex-task-test", "version": "1.0"},
	})
	f.receive(t, 1)
	return f
}

func (f *codexTaskMCPStdio) send(t *testing.T, id int, method string, params map[string]any) {
	t.Helper()
	request, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": id, "method": method, "params": params,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.in.Write(append(request, '\n')); err != nil {
		t.Fatal(err)
	}
}

func (f *codexTaskMCPStdio) receive(t *testing.T, id int) map[string]json.RawMessage {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case message, ok := <-f.messages:
			if !ok {
				t.Fatal("MCP stdout closed before response")
			}
			var got int
			if json.Unmarshal(message["id"], &got) == nil && got == id {
				return message
			}
		case err := <-f.done:
			t.Fatalf("MCP server exited before response: %v", err)
		case <-deadline:
			t.Fatalf("MCP response id=%d timed out", id)
		}
	}
}

func (f *codexTaskMCPStdio) awaitExit(t *testing.T) {
	t.Helper()
	select {
	case err := <-f.done:
		if err != nil {
			t.Fatalf("MCP server exit: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("MCP server did not exit after stdin EOF")
	}
}

func mcpCodexJobID(t *testing.T, message map[string]json.RawMessage) string {
	t.Helper()
	var result struct {
		StructuredContent struct {
			JobID string `json:"job_id"`
		} `json:"structuredContent"`
	}
	if err := json.Unmarshal(message["result"], &result); err != nil {
		t.Fatal(err)
	}
	if result.StructuredContent.JobID == "" {
		t.Fatalf("codex_task returned no job id: %s", message["result"])
	}
	return result.StructuredContent.JobID
}

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
	withCodexTaskTimeout(t, 30*time.Second) // natural expiry cannot satisfy the 5s shutdown check
	bin := filepath.Join(root, "codex-app-server")
	pidFile := filepath.Join(root, "child.pid")
	t.Setenv("MOAI_TEST_CHILD_PID_PATH", pidFile)
	script := `#!/bin/sh
printf '%s\n' "$$" > "$MOAI_TEST_CHILD_PID_PATH"
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
	stdio := startCodexTaskMCPStdio(t, root)
	t.Cleanup(func() {
		if raw, err := os.ReadFile(pidFile); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(raw))); err == nil {
				if syscall.Kill(pid, 0) == nil {
					if child, err := os.FindProcess(pid); err == nil {
						_ = child.Kill() // fixture only, even on test failure
						_ = child.Release()
					}
				}
			}
		}
	})

	stdio.send(t, 2, "tools/call", map[string]any{
		"name": "codex_task", "arguments": map[string]any{"prompt": "stall", "background": true},
	})
	jobID := mcpCodexJobID(t, stdio.receive(t, 2))
	_ = stdio.in.Close() // real MCP stdin EOF, not a direct call to the cleanup helper
	stdio.awaitExit(t)
	rec := awaitTerminalJob(t, newCodexJobRegistry(root), jobID)
	if rec.Status != codexJobStatusFailed {
		t.Fatalf("status=%s error=%q, want failed after shutdown", rec.Status, rec.Error)
	}
	if _, live := codexLiveJobSessions.Load(jobID); live {
		t.Fatal("session remained live after shutdown")
	}
	if raw, err := os.ReadFile(pidFile); err != nil {
		t.Fatal(err)
	} else if pid, err := strconv.Atoi(strings.TrimSpace(string(raw))); err != nil {
		t.Fatal(err)
	} else if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
		t.Fatalf("app-server pid %d remains live after EOF: %v", pid, err)
	}
}

func TestCodexTaskMCPEOFAbortsStalledHandshake(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	bin := filepath.Join(root, "codex-app-server")
	pidFile := filepath.Join(root, "handshake-child.pid")
	t.Setenv("MOAI_TEST_CHILD_PID_PATH", pidFile)
	// The child accepts initialize but never answers it. Without EOF-driven
	// request cancellation, the stdio server waits forever for this tool worker.
	script := `#!/bin/sh
printf '%s\n' "$$" > "$MOAI_TEST_CHILD_PID_PATH"
while IFS= read -r line; do :; done
`
	if err := os.WriteFile(bin, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	prevLook, prevSession := codexLookPath, codexSession
	codexLookPath = func(string) (string, error) { return bin, nil }
	codexSession = realCodexSessionRunner{}
	t.Cleanup(func() { codexLookPath, codexSession = prevLook, prevSession })
	stdio := startCodexTaskMCPStdio(t, root)
	t.Cleanup(func() {
		if raw, err := os.ReadFile(pidFile); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(raw))); err == nil && syscall.Kill(pid, 0) == nil {
				if child, err := os.FindProcess(pid); err == nil {
					_ = child.Kill()
					_ = child.Release()
				}
			}
		}
	})
	stdio.send(t, 2, "tools/call", map[string]any{
		"name": "codex_task", "arguments": map[string]any{"prompt": "stall handshake", "background": true},
	})
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(pidFile); err == nil {
			break // tool worker reached the real app-server child
		}
		if time.Now().After(deadline) {
			t.Fatal("codex child never reached initialize handshake")
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = stdio.in.Close()
	stdio.awaitExit(t)
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
		t.Fatalf("handshake child pid %d remains live after EOF: %v", pid, err)
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
