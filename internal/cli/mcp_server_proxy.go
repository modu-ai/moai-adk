// mcp_server_proxy.go — stale-server write proxy and the hidden one-shot call
// form of `moai mcp-server`.
//
// A long-lived MCP server keeps running the build it started with. When the
// installed `moai` binary moves on (a reinstall), a write-capable call used to
// be refused until the operator reconnected the server by hand. Instead the
// stale server now re-executes that single call through the INSTALLED build:
// `<installed> mcp-server --call <tool> --input-json -` (arguments on stdin).
//
// The one-shot form builds the same tool registry the serving loop builds,
// looks the handler up by name, invokes it once and prints the
// *mcp.CallToolResult as JSON. The serving loop injects no per-call context
// values of its own; per-call attribution (session, factory role, project dir)
// comes from the inherited process environment and working directory, which the
// proxy passes through unchanged. The one-shot context is the cobra command
// context (background when unset).
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	// mcpOneShotEnv marks a process running the one-shot form. The guard serves
	// locally while it is set, so the installed build never proxies again.
	mcpOneShotEnv = "MOAI_MCP_ONESHOT"
	// mcpProxyCallTimeout bounds one proxied call.
	mcpProxyCallTimeout = 10 * time.Minute
	// mcpInputFromStdin is the --input-json value that reads the arguments from stdin.
	mcpInputFromStdin = "-"
)

// mcpOneShotActive reports whether this process is the one-shot form.
func mcpOneShotActive() bool { return os.Getenv(mcpOneShotEnv) == "1" }

// runMCPServerOneShot invokes the named tool once and writes the result JSON.
// It errors only on registry/parse failure; a tool-level error is carried in
// the result (isError) with a nil error.
func runMCPServerOneShot(ctx context.Context, stdin io.Reader, stdout io.Writer, name, inputJSON, exposure string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := os.Setenv(mcpOneShotEnv, "1"); err != nil {
		return err
	}
	raw := []byte(inputJSON)
	if inputJSON == mcpInputFromStdin {
		b, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("read --input-json from stdin: %w", err)
		}
		raw = b
	}
	var args map[string]any
	if trimmed := bytes.TrimSpace(raw); len(trimmed) > 0 {
		if err := json.Unmarshal(trimmed, &args); err != nil {
			return fmt.Errorf("parse --input-json: %w", err)
		}
	}
	s := newMoaiMCPServerWithExposure(exposure)
	return callMCPToolOnce(ctx, s, stdout, name, args)
}

// callMCPToolOnce looks the tool up on s, invokes it once and encodes the result.
func callMCPToolOnce(ctx context.Context, s *server.MCPServer, stdout io.Writer, name string, args map[string]any) error {
	tool := s.GetTool(name)
	if tool == nil {
		return fmt.Errorf("unknown or disabled tool %q", name)
	}
	var req mcp.CallToolRequest
	req.Params.Name = name
	req.Params.Arguments = args
	res, err := tool.Handler(ctx, req)
	if err != nil {
		res = toolErr(name, err)
	}
	if res == nil {
		return fmt.Errorf("tool %q returned no result", name)
	}
	return json.NewEncoder(stdout).Encode(res)
}

// proxyMCPCallToInstalled runs the call through the installed build and decodes
// its result. Any exec or decode failure is returned so the caller can fail closed.
func proxyMCPCallToInstalled(ctx context.Context, installed, name string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if installed == "" {
		return nil, errors.New("no installed moai binary path")
	}
	body, err := json.Marshal(req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("marshal arguments: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, mcpProxyCallTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, installed, moaiMCPServerSubcommand, "--call", name, "--input-json", mcpInputFromStdin)
	cmd.Stdin = bytes.NewReader(body)
	cmd.Env = append(os.Environ(), mcpOneShotEnv+"=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec %s: %w (%s)", installed, err, strings.TrimSpace(mcpProxyStderrTail(stderr.String(), 400)))
	}
	raw := json.RawMessage(bytes.TrimSpace(stdout.Bytes()))
	res, err := mcp.ParseCallToolResult(&raw)
	if err != nil {
		return nil, fmt.Errorf("decode result from %s: %w", installed, err)
	}
	return res, nil
}

// mcpProxyStderrTail returns at most the last n bytes of s.
func mcpProxyStderrTail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
