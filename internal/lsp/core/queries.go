package core

import (
	"context"
	"encoding/json"
	"fmt"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// LSP query method constants (prevent hard-coding).
const (
	methodReferences = "textDocument/references"
	methodDefinition = "textDocument/definition"
)

// GetDiagnostics returns the latest diagnostics for path from the push-model cache (REQ-LC-002b).
//
// Behavior:
//   - The handler registered in Start() stores server-pushed textDocument/publishDiagnostics notifications in the cache.
//   - GetDiagnostics returns immediately from that cache (no blocking I/O).
//   - Returns ErrFileNotOpen if the URI is not in the document cache.
//
// @MX:ANCHOR: [AUTO] GetDiagnostics — public API for LSP diagnostic queries
// @MX:REASON: fan_in >= 3 — called by Ralph engine, Quality Gates, LOOP command, Manager, and integration tests
func (c *client) GetDiagnostics(_ context.Context, path string) ([]lsp.Diagnostic, error) {
	uri := pathToURI(path)

	// ErrFileNotOpen: returned when the URI is not in the document cache (REQ-LC-002b rule: caller must call OpenFile first).
	snap := c.docs.snapshot()
	if _, ok := snap[uri]; !ok {
		return nil, ErrFileNotOpen
	}

	c.diagnosticsMu.RLock()
	diags := c.diagnostics[uri]
	c.diagnosticsMu.RUnlock()

	// nil slice → empty slice (consistency).
	if diags == nil {
		return []lsp.Diagnostic{}, nil
	}
	out := make([]lsp.Diagnostic, len(diags))
	copy(out, diags)
	return out, nil
}

// FindReferences returns all reference locations for the symbol at pos in path.
//
// Precondition check:
//   - Returns ErrCapabilityUnsupported if serverCaps.Supports("textDocument/references") is false.
//
// @MX:ANCHOR: [AUTO] FindReferences — public API for LSP references queries
// @MX:REASON: fan_in >= 3 — called by Ralph engine, Quality Gates, LOOP command, and integration tests
func (c *client) FindReferences(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error) {
	if !c.serverCaps.Supports(methodReferences) {
		return nil, fmt.Errorf("FindReferences: %w", ErrCapabilityUnsupported)
	}

	uri := pathToURI(path)
	params := map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     pos,
		"context":      map[string]any{"includeDeclaration": true},
	}

	var raw json.RawMessage
	if err := transport.CallWithTimeout(ctx, c.tr, methodReferences, params, &raw, c.cfg.Language); err != nil {
		return nil, transport.WrapCallError(methodReferences, uri, c.cfg.Language, err)
	}

	return parseLocations(raw)
}

// GotoDefinition returns definition locations for the symbol at pos in path.
//
// Precondition check:
//   - Returns ErrCapabilityUnsupported if serverCaps.Supports("textDocument/definition") is false.
//
// The LSP response may be a Location, []Location, or LocationLink[].
// The v1 implementation handles []Location or a single Location (tolerant decoder).
//
// @MX:ANCHOR: [AUTO] GotoDefinition — public API for LSP definition queries
// @MX:REASON: fan_in >= 3 — called by Ralph engine, Quality Gates, LOOP command, and integration tests
func (c *client) GotoDefinition(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error) {
	if !c.serverCaps.Supports(methodDefinition) {
		return nil, fmt.Errorf("GotoDefinition: %w", ErrCapabilityUnsupported)
	}

	uri := pathToURI(path)
	params := map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     pos,
	}

	var raw json.RawMessage
	if err := transport.CallWithTimeout(ctx, c.tr, methodDefinition, params, &raw, c.cfg.Language); err != nil {
		return nil, transport.WrapCallError(methodDefinition, uri, c.cfg.Language, err)
	}

	return parseLocations(raw)
}

// parseLocations decodes a raw LSP response JSON into []Location.
//
// Decoding strategy (tolerant decoder):
//  1. Try array: unmarshal as []lsp.Location.
//  2. On failure, try single object: unmarshal as lsp.Location → wrap in []lsp.Location{}.
//  3. null response: return an empty slice.
func parseLocations(raw json.RawMessage) ([]lsp.Location, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []lsp.Location{}, nil
	}

	// Try array first.
	var locs []lsp.Location
	if err := json.Unmarshal(raw, &locs); err == nil {
		return locs, nil
	}

	// Try single object (tolerant fallback).
	var single lsp.Location
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, fmt.Errorf("parseLocations: unable to decode as array or single location: %w", err)
	}
	return []lsp.Location{single}, nil
}
