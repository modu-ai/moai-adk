// Package core implements the LSP client lifecycle and query operations
// for the multi-language LSP client (SPEC-LSP-CORE-002).
//
// The primary entry point is NewClient, which creates a Client that manages
// the full lifecycle of a language server subprocess: spawn, initialize,
// query, and shutdown.
//
// Consumers import this package to obtain LSP diagnostics, references, and
// definition locations for any language server described by a config.ServerConfig.
package core
