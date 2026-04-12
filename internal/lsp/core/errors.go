package core

import "errors"

// ErrNotImplemented is returned by Client methods that are stubbed in Sprint 3
// and will be implemented in later sprints (T-011..T-014).
//
// Callers can check:
//
//	if errors.Is(err, core.ErrNotImplemented) { /* skip in Sprint 3 */ }
var ErrNotImplemented = errors.New("lsp: not implemented in this sprint")

// ErrCapabilityUnsupported is returned when a query method is called but the
// server does not advertise the required capability in its initialize response
// (REQ-LC-033).
//
// The error message includes the unsupported method name.
var ErrCapabilityUnsupported = errors.New("lsp: capability unsupported by server")

// ErrFileNotOpen is returned by GetDiagnostics when the requested file has not
// been opened via OpenFile (T-013).
//
// @MX:ANCHOR: [AUTO] ErrFileNotOpen — sentinel used across GetDiagnostics, Manager, and test assertions
// @MX:REASON: fan_in >= 3 — GetDiagnostics, Manager routing, integration tests, and upstream callers all branch on this sentinel
var ErrFileNotOpen = errors.New("lsp: file not open")
