// Package transport defines the Transport interface for LSP JSON-RPC communication
// and provides a powernap-backed implementation.
//
// The Transport interface abstracts the JSON-RPC 2.0 wire layer, allowing
// upstream packages (core.Client, document sync) to be tested independently
// using mock implementations.
//
// The concrete powernapTransport delegates to github.com/charmbracelet/x/powernap/pkg/transport.Connection
// for actual subprocess stdio communication (REQ-LC-001).
//
// Intended usage:
//
//	stream := transport.NewStreamTransport(proc.Stdin, proc.Stdout, proc.Stdin)
//	tr := transport.NewPowernapTransport(stream)
//	defer tr.Close()
//
//	var result protocol.InitializeResult
//	if err := tr.Call(ctx, "initialize", params, &result); err != nil {
//	    // handle
//	}
package transport
