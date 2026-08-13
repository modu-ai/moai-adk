package cli

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestExtractProgressToken covers the three request shapes a tool handler can
// receive: no _meta at all, _meta without a progressToken (client did not opt
// into progress reporting), and _meta carrying a progressToken.
func TestExtractProgressToken(t *testing.T) {
	cases := []struct {
		name      string
		req       mcp.CallToolRequest
		wantToken bool // whether a non-nil token is expected
	}{
		{
			name:      "nil_meta",
			req:       mcp.CallToolRequest{},
			wantToken: false,
		},
		{
			name: "meta_without_token",
			req: mcp.CallToolRequest{
				Params: mcp.CallToolParams{Meta: &mcp.Meta{}},
			},
			wantToken: false,
		},
		{
			name: "meta_with_string_token",
			req: mcp.CallToolRequest{
				Params: mcp.CallToolParams{Meta: &mcp.Meta{ProgressToken: "tok-abc"}},
			},
			wantToken: true,
		},
		{
			name: "meta_with_int_token",
			req: mcp.CallToolRequest{
				Params: mcp.CallToolParams{Meta: &mcp.Meta{ProgressToken: 42}},
			},
			wantToken: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractProgressToken(c.req)
			if c.wantToken && got == nil {
				t.Errorf("extractProgressToken(%s) = nil; want non-nil token", c.name)
			}
			if !c.wantToken && got != nil {
				t.Errorf("extractProgressToken(%s) = %v; want nil (no progress opt-in)", c.name, got)
			}
		})
	}
}

// TestNotifyMCPProgress_NoOpOutsideServer verifies notifyMCPProgress is a safe
// no-op when called outside an MCP server context (ServerFromContext returns
// nil — the case in unit tests and direct CLI use). It must never panic or
// block, whether or not a token is supplied.
func TestNotifyMCPProgress_NoOpOutsideServer(t *testing.T) {
	// A plain background context carries no MCPServer, so both notification
	// paths bail at the ServerFromContext guard. This guards the fail-open
	// contract: progress is advisory-only and never panics without a server.
	notifyMCPProgress(context.Background(), nil, 0.5, "no server, no token — must no-op")
	notifyMCPProgress(context.Background(), mcp.ProgressToken("tok"), 0.5, "token but no server — must no-op")
}
