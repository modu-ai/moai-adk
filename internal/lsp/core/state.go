package core

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// ClientState represents the lifecycle state of a language server client.
// States progress through a strict transition graph — see Transition for allowed paths.
type ClientState string

const (
	// StateSpawning is the initial state: the subprocess is being launched.
	StateSpawning ClientState = "spawning"

	// StateInitializing indicates the subprocess is running and the LSP
	// initialize handshake is in progress.
	StateInitializing ClientState = "initializing"

	// StateReady indicates the client has successfully initialized and is
	// ready to serve queries.
	StateReady ClientState = "ready"

	// StateDegraded indicates the client is partially functional (e.g., the
	// subprocess crashed or a capability is missing). Recovery to StateReady
	// is allowed.
	StateDegraded ClientState = "degraded"

	// StateShutdown is the terminal state. Once reached, no further transitions
	// are permitted and all resources have been released.
	StateShutdown ClientState = "shutdown"
)

// ErrInvalidTransition is returned by StateMachine.Transition when the requested
// state change is not in the allowed transition table.
//
// Callers can use errors.Is to detect invalid transitions:
//
//	if errors.Is(err, ErrInvalidTransition) { ... }
var ErrInvalidTransition = errors.New("state: invalid transition")

// allowedTransitions defines the valid state transition graph.
//
// Allowed transitions:
//   - spawning     → initializing, shutdown
//   - initializing → ready, degraded, shutdown
//   - ready        → degraded, shutdown
//   - degraded     → ready, shutdown
//   - shutdown     → (terminal — no transitions allowed)
var allowedTransitions = map[ClientState][]ClientState{
	StateSpawning:     {StateInitializing, StateShutdown},
	StateInitializing: {StateReady, StateDegraded, StateShutdown},
	StateReady:        {StateDegraded, StateShutdown},
	StateDegraded:     {StateReady, StateShutdown},
	StateShutdown:     {}, // terminal
}

// StateMachine manages ClientState transitions with thread-safety and logging.
//
// @MX:NOTE: [AUTO] StateMachine.Transition — 상태 전환은 slog.Debug로 기록되어 디버깅 가능
type StateMachine struct {
	mu      sync.RWMutex
	current ClientState
	logger  *slog.Logger
}

// NewStateMachine creates a StateMachine starting in StateSpawning.
// If logger is nil, slog.Default() is used.
//
// @MX:ANCHOR: [AUTO] NewStateMachine — creates the client state machine; consumed by NewClient and tests
// @MX:REASON: fan_in >= 3 — NewClient, test helpers, and future Manager all construct StateMachine
func NewStateMachine(logger *slog.Logger) *StateMachine {
	if logger == nil {
		logger = slog.Default()
	}
	return &StateMachine{
		current: StateSpawning,
		logger:  logger,
	}
}

// newStateMachineAt creates a StateMachine with an arbitrary initial state.
// For testing only — not exported.
func newStateMachineAt(initial ClientState, logger *slog.Logger) *StateMachine {
	if logger == nil {
		logger = slog.Default()
	}
	return &StateMachine{
		current: initial,
		logger:  logger,
	}
}

// isErrInvalidTransition reports whether err wraps ErrInvalidTransition.
// For testing only — not exported.
func isErrInvalidTransition(err error) bool {
	return errors.Is(err, ErrInvalidTransition)
}

// Transition attempts to advance the state machine to the given state.
// Returns ErrInvalidTransition if the transition is not in the allowed table.
// On success, logs the transition at slog.Debug.
//
// Thread-safe: safe to call from multiple goroutines.
func (sm *StateMachine) Transition(to ClientState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	from := sm.current
	allowed := allowedTransitions[from]

	for _, s := range allowed {
		if s == to {
			sm.current = to
			sm.logger.Debug("lsp client state transition",
				slog.String("from", string(from)),
				slog.String("to", string(to)),
			)
			return nil
		}
	}

	return fmt.Errorf("state: %q → %q: %w", from, to, ErrInvalidTransition)
}

// Current returns the current state.
//
// Thread-safe: may be called concurrently with Transition.
func (sm *StateMachine) Current() ClientState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}
