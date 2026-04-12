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
