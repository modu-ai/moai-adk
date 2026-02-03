package hook

import "errors"

// Sentinel errors for hook operations.
var (
	// ErrHookTimeout indicates hook execution exceeded the configured timeout.
	ErrHookTimeout = errors.New("hook: execution timed out")

	// ErrHookContractFail indicates the hook execution contract was violated.
	// See ADR-012 for the complete contract specification.
	ErrHookContractFail = errors.New("hook: execution contract violated")

	// ErrHookInvalidInput indicates the JSON input from stdin was invalid
	// or missing required fields.
	ErrHookInvalidInput = errors.New("hook: invalid JSON input")

	// ErrHookBlocked indicates an action was blocked by a hook handler.
	ErrHookBlocked = errors.New("hook: action blocked by hook")
)
