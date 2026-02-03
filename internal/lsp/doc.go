// Package lsp provides a Language Server Protocol (LSP) client system
// for communicating with language servers using JSON-RPC 2.0 over stdio or TCP.
//
// The package implements a layered architecture:
//
//   - models.go: Core LSP data types (Diagnostic, Position, Range, Location, etc.)
//   - protocol.go: JSON-RPC 2.0 transport and connection management
//   - client.go: LSP client interface for single-server communication
//   - server.go: ServerManager for multi-server lifecycle management
//
// Basic usage:
//
//	mgr := lsp.NewServerManager(launcher, lsp.WithMaxParallel(4))
//	err := mgr.StartServer(ctx, "go")
//	client, err := mgr.GetClient("go")
//	diagnostics, err := client.Diagnostics(ctx, "file:///path/to/file.go")
//	err = mgr.StopServer(ctx, "go")
//
// Error handling uses sentinel errors that can be checked with errors.Is:
//
//	if errors.Is(err, lsp.ErrServerNotRunning) {
//	    // server is not running
//	}
//
// All long-running operations accept context.Context as the first parameter
// and respect cancellation and timeout signals.
package lsp
