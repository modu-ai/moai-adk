package core

import (
	"bytes"
	"log/slog"
	"sync"
	"testing"
)

// logBuffer는 slog 출력을 캡처하기 위한 thread-safe 버퍼.
type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (lb *logBuffer) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.buf.Write(p)
}

func (lb *logBuffer) String() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.buf.String()
}

func newTestLogger(buf *logBuffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// TestStateMachine_InitialState verifies that a new StateMachine starts in StateSpawning.
func TestStateMachine_InitialState(t *testing.T) {
	sm := NewStateMachine(nil)
	if got := sm.Current(); got != StateSpawning {
		t.Errorf("expected initial state %q, got %q", StateSpawning, got)
	}
}

// TestStateMachine_AllowedTransitions verifies each allowed transition succeeds.
func TestStateMachine_AllowedTransitions(t *testing.T) {
	tests := []struct {
		name string
		from ClientState
		to   ClientState
	}{
		{"spawning→initializing", StateSpawning, StateInitializing},
		{"spawning→shutdown", StateSpawning, StateShutdown},
		{"initializing→ready", StateInitializing, StateReady},
		{"initializing→degraded", StateInitializing, StateDegraded},
		{"initializing→shutdown", StateInitializing, StateShutdown},
		{"ready→degraded", StateReady, StateDegraded},
		{"ready→shutdown", StateReady, StateShutdown},
		{"degraded→ready", StateDegraded, StateReady},
		{"degraded→shutdown", StateDegraded, StateShutdown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 각 테스트는 독립된 StateMachine을 사용
			sm := newStateMachineAt(tt.from, nil)
			if err := sm.Transition(tt.to); err != nil {
				t.Errorf("Transition(%q→%q) unexpected error: %v", tt.from, tt.to, err)
			}
			if got := sm.Current(); got != tt.to {
				t.Errorf("after Transition: expected state %q, got %q", tt.to, got)
			}
		})
	}
}

// TestStateMachine_DeniedTransitions verifies invalid transitions return ErrInvalidTransition.
func TestStateMachine_DeniedTransitions(t *testing.T) {
	tests := []struct {
		name string
		from ClientState
		to   ClientState
	}{
		{"spawning→ready", StateSpawning, StateReady},
		{"spawning→degraded", StateSpawning, StateDegraded},
		{"initializing→spawning", StateInitializing, StateSpawning},
		{"ready→spawning", StateReady, StateSpawning},
		{"ready→initializing", StateReady, StateInitializing},
		{"degraded→spawning", StateDegraded, StateSpawning},
		{"degraded→initializing", StateDegraded, StateInitializing},
		{"shutdown→anything", StateShutdown, StateSpawning},
		{"shutdown→ready", StateShutdown, StateReady},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := newStateMachineAt(tt.from, nil)
			err := sm.Transition(tt.to)
			if err == nil {
				t.Errorf("Transition(%q→%q): expected ErrInvalidTransition, got nil", tt.from, tt.to)
				return
			}
			if !isErrInvalidTransition(err) {
				t.Errorf("Transition(%q→%q): expected ErrInvalidTransition, got %v", tt.from, tt.to, err)
			}
		})
	}
}

// TestStateMachine_ShutdownTerminal verifies that shutdown is a terminal state
// and any further transition returns ErrInvalidTransition.
func TestStateMachine_ShutdownTerminal(t *testing.T) {
	sm := NewStateMachine(nil)
	if err := sm.Transition(StateShutdown); err != nil {
		t.Fatalf("Transition to shutdown: %v", err)
	}
	// shutdown 이후 어떤 상태로도 전환 불가
	for _, next := range []ClientState{StateSpawning, StateInitializing, StateReady, StateDegraded} {
		err := sm.Transition(next)
		if err == nil {
			t.Errorf("from shutdown→%q: expected error, got nil", next)
		}
	}
}

// TestStateMachine_LogsOnTransition verifies that slog.Debug is called on allowed transitions.
func TestStateMachine_LogsOnTransition(t *testing.T) {
	buf := &logBuffer{}
	logger := newTestLogger(buf)

	sm := NewStateMachine(logger)
	if err := sm.Transition(StateInitializing); err != nil {
		t.Fatalf("Transition: %v", err)
	}

	got := buf.String()
	if got == "" {
		t.Error("expected slog output after Transition, got empty string")
	}
}

// TestStateMachine_ConcurrentCurrentReads verifies that concurrent Current() calls are race-free.
func TestStateMachine_ConcurrentCurrentReads(t *testing.T) {
	t.Parallel()

	sm := NewStateMachine(nil)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			_ = sm.Current()
		}()
	}
	wg.Wait()
}

// TestStateMachine_ConcurrentTransitionAndRead verifies that mixed concurrent
// reads and a single writer are race-free.
func TestStateMachine_ConcurrentTransitionAndRead(t *testing.T) {
	t.Parallel()

	sm := NewStateMachine(nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 단일 writer: spawning→initializing
		_ = sm.Transition(StateInitializing)
	}()

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sm.Current()
		}()
	}
	wg.Wait()
}
