package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:     "mcp",
	Short:   "MCP server commands",
	GroupID: "tools",
}

var mcpLSPCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start LSP MCP stdio server for Claude Code integration",
	Long: `Starts an MCP (Model Context Protocol) server over stdio that provides
LSP-based code intelligence tools (goto definition, find references, hover,
document symbols, diagnostics, rename) to Claude Code agents.

The server communicates via newline-delimited JSON-RPC 2.0 on stdin/stdout.
Add it to .mcp.json to make it available in Claude Code sessions.`,
	RunE: runMCPLSP,
}

func init() {
	mcpCmd.AddCommand(mcpLSPCmd)
	rootCmd.AddCommand(mcpCmd)
}

// runMCPLSP starts the MCP LSP stdio server and blocks until the process
// receives SIGINT or SIGTERM, or stdin is closed by the MCP client.
func runMCPLSP(_ *cobra.Command, _ []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	handler := mcp.NewLSPHandler(cwd)
	server := mcp.NewServer(handler)

	return server.Serve(ctx, os.Stdin, os.Stdout)
}
