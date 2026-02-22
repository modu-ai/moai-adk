package cli

import (
	"context"
	"io"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// --- Spy types for verifying event dispatch wiring ---

// spyRegistry records which EventTypes are dispatched.
type spyRegistry struct {
	dispatched []hook.EventType
}

func (s *spyRegistry) Register(_ hook.Handler)                  {}
func (s *spyRegistry) Handlers(_ hook.EventType) []hook.Handler { return nil }
func (s *spyRegistry) Dispatch(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
	s.dispatched = append(s.dispatched, event)
	return &hook.HookOutput{}, nil
}

// spyProtocol returns a valid HookInput without reading real stdin.
type spyProtocol struct{}

func (s *spyProtocol) ReadInput(_ io.Reader) (*hook.HookInput, error) {
	return &hook.HookInput{}, nil
}

func (s *spyProtocol) WriteOutput(_ io.Writer, _ *hook.HookOutput) error {
	return nil
}

// --- Test 1: Verify all 7 new subcommands exist ---

func TestHookSubcommands_AllNewEventsRegistered(t *testing.T) {
	t.Parallel()

	expectedSubcommands := []string{
		"post-tool-failure",
		"notification",
		"subagent-start",
		"user-prompt-submit",
		"permission-request",
		"teammate-idle",
		"task-completed",
	}

	registeredNames := make(map[string]bool)
	for _, cmd := range hookCmd.Commands() {
		registeredNames[cmd.Name()] = true
	}

	for _, name := range expectedSubcommands {
		if !registeredNames[name] {
			t.Errorf("hook subcommand %q not found; registered commands: %v",
				name, registeredNames)
		}
	}
}

// --- Test 2: Verify each subcommand maps to the correct EventType ---
//
// This test uses a spy registry and protocol to intercept the dispatch call
// made by runHookEvent, allowing us to verify the EventType mapping without
// accessing the unexported hookSubcommands variable.
//
// Note: subtests are NOT parallel because they share the global deps variable.

func TestHookSubcommands_EventTypeMapping(t *testing.T) {
	// Not parallel: subtests write to global deps sequentially.
	origDeps := deps
	defer func() { deps = origDeps }()

	expected := map[string]hook.EventType{
		"post-tool-failure":  hook.EventPostToolUseFailure,
		"notification":       hook.EventNotification,
		"subagent-start":     hook.EventSubagentStart,
		"user-prompt-submit": hook.EventUserPromptSubmit,
		"permission-request": hook.EventPermissionRequest,
		"teammate-idle":      hook.EventTeammateIdle,
		"task-completed":     hook.EventTaskCompleted,
	}

	for subcmdName, wantEvent := range expected {
		t.Run(subcmdName, func(t *testing.T) {
			spy := &spyRegistry{}
			deps = &Dependencies{
				HookRegistry: spy,
				HookProtocol: &spyProtocol{},
			}

			// Find the subcommand by name and execute with a valid context.
			var found bool
			for _, cmd := range hookCmd.Commands() {
				if cmd.Name() == subcmdName {
					found = true
					cmd.SetContext(context.Background())
					err := cmd.RunE(cmd, []string{})
					if err != nil {
						t.Fatalf("RunE for %q returned error: %v", subcmdName, err)
					}
					break
				}
			}
			if !found {
				t.Fatalf("subcommand %q not found in hookCmd", subcmdName)
			}

			if len(spy.dispatched) != 1 {
				t.Fatalf("expected 1 dispatch call, got %d", len(spy.dispatched))
			}
			if spy.dispatched[0] != wantEvent {
				t.Errorf("subcommand %q dispatched EventType %q, want %q",
					subcmdName, spy.dispatched[0], wantEvent)
			}
		})
	}
}

// --- Test 3: Verify all 14 handlers are wired via InitDependencies ---

func TestHookDepsWiring_AllHandlersRegistered(t *testing.T) {
	// Not parallel: calls InitDependencies which writes to global deps.
	origDeps := deps
	defer func() { deps = origDeps }()

	InitDependencies()

	allEvents := []hook.EventType{
		hook.EventSessionStart,
		hook.EventSessionEnd,
		hook.EventPreToolUse,
		hook.EventPostToolUse,
		hook.EventStop,
		hook.EventPreCompact,
		hook.EventSubagentStop, // registered via RankSessionHandler conditionally
		// New events:
		hook.EventPostToolUseFailure,
		hook.EventNotification,
		hook.EventSubagentStart,
		hook.EventUserPromptSubmit,
		hook.EventPermissionRequest,
		hook.EventTeammateIdle,
		hook.EventTaskCompleted,
	}

	// Events that may not have a handler (conditionally registered).
	optionalEvents := map[hook.EventType]bool{
		hook.EventSubagentStop: true, // RankSessionHandler is conditional
	}

	for _, event := range allEvents {
		handlers := deps.HookRegistry.Handlers(event)
		if len(handlers) == 0 && !optionalEvents[event] {
			t.Errorf("no handlers registered for event %q; expected at least 1", event)
		}
	}
}

// --- Test 4: ValidEventTypes returns exactly 16 event types ---

func TestHookValidEventTypes_Complete(t *testing.T) {
	t.Parallel()

	allNewEvents := []hook.EventType{
		hook.EventPostToolUseFailure,
		hook.EventNotification,
		hook.EventSubagentStart,
		hook.EventUserPromptSubmit,
		hook.EventPermissionRequest,
		hook.EventTeammateIdle,
		hook.EventTaskCompleted,
		hook.EventWorktreeCreate,
		hook.EventWorktreeRemove,
	}

	validTypes := hook.ValidEventTypes()

	// Verify exact count.
	if got := len(validTypes); got != 16 {
		t.Errorf("ValidEventTypes() returned %d types, want 16", got)
	}

	// Build a lookup set for quick membership checks.
	validSet := make(map[hook.EventType]bool, len(validTypes))
	for _, et := range validTypes {
		validSet[et] = true
	}

	// Verify all 7 new events are included.
	for _, event := range allNewEvents {
		if !validSet[event] {
			t.Errorf("ValidEventTypes() missing new event %q", event)
		}
	}

	// Also verify no duplicates.
	seen := make(map[hook.EventType]bool, len(validTypes))
	for _, et := range validTypes {
		if seen[et] {
			t.Errorf("ValidEventTypes() contains duplicate: %q", et)
		}
		seen[et] = true
	}
}

// --- Test 5: All 7 new event types pass IsValidEventType ---

func TestHookNewEventTypes_AreValid(t *testing.T) {
	t.Parallel()

	newEvents := []struct {
		name  string
		event hook.EventType
	}{
		{"PostToolUseFailure", hook.EventPostToolUseFailure},
		{"Notification", hook.EventNotification},
		{"SubagentStart", hook.EventSubagentStart},
		{"UserPromptSubmit", hook.EventUserPromptSubmit},
		{"PermissionRequest", hook.EventPermissionRequest},
		{"TeammateIdle", hook.EventTeammateIdle},
		{"TaskCompleted", hook.EventTaskCompleted},
	}

	for _, tc := range newEvents {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !hook.IsValidEventType(tc.event) {
				t.Errorf("IsValidEventType(%q) = false, want true", tc.event)
			}
		})
	}

	// Verify an invalid event type returns false.
	if hook.IsValidEventType(hook.EventType("BogusEvent")) {
		t.Error("IsValidEventType(\"BogusEvent\") = true, want false")
	}
}

// --- Bonus: Verify ValidEventTypes is sorted consistently ---

func TestHookValidEventTypes_Deterministic(t *testing.T) {
	t.Parallel()

	// Call twice and verify identical order.
	first := hook.ValidEventTypes()
	second := hook.ValidEventTypes()

	if len(first) != len(second) {
		t.Fatalf("ValidEventTypes() returned different lengths: %d vs %d", len(first), len(second))
	}

	for i := range first {
		if first[i] != second[i] {
			t.Errorf("ValidEventTypes()[%d] = %q on first call, %q on second call",
				i, first[i], second[i])
		}
	}
}

// --- Bonus: Verify all ValidEventTypes have a corresponding hookCmd subcommand ---

func TestHookValidEventTypes_AllHaveSubcommands(t *testing.T) {
	t.Parallel()

	// Events that do NOT have a direct hookCmd subcommand.
	// (SubagentStop now has a direct subcommand in addition to the "agent" pattern.)
	excludedEvents := map[hook.EventType]bool{}

	// Build a mapping from EventType to expected subcommand name.
	eventToSubcmd := map[hook.EventType]string{
		hook.EventSessionStart:       "session-start",
		hook.EventSessionEnd:         "session-end",
		hook.EventPreToolUse:         "pre-tool",
		hook.EventPostToolUse:        "post-tool",
		hook.EventStop:               "stop",
		hook.EventPreCompact:         "compact",
		hook.EventPostToolUseFailure: "post-tool-failure",
		hook.EventNotification:       "notification",
		hook.EventSubagentStart:      "subagent-start",
		hook.EventUserPromptSubmit:   "user-prompt-submit",
		hook.EventPermissionRequest:  "permission-request",
		hook.EventTeammateIdle:       "teammate-idle",
		hook.EventTaskCompleted:      "task-completed",
		hook.EventSubagentStop:       "subagent-stop",
		hook.EventWorktreeCreate:     "worktree-create",
		hook.EventWorktreeRemove:     "worktree-remove",
	}

	registeredNames := make(map[string]bool)
	for _, cmd := range hookCmd.Commands() {
		registeredNames[cmd.Name()] = true
	}

	for _, event := range hook.ValidEventTypes() {
		if excludedEvents[event] {
			continue
		}
		subcmd, ok := eventToSubcmd[event]
		if !ok {
			t.Errorf("no expected subcommand mapping for event %q", event)
			continue
		}
		if !registeredNames[subcmd] {
			t.Errorf("event %q expects subcommand %q but it is not registered", event, subcmd)
		}
	}

	// Reverse check: verify all hook event subcommands map to a valid event.
	subcmdToEvent := make(map[string]hook.EventType, len(eventToSubcmd))
	for event, subcmd := range eventToSubcmd {
		subcmdToEvent[subcmd] = event
	}

	// Collect event subcommand names (exclude utility subcommands like "list", "agent", "pre-push").
	utilitySubcmds := map[string]bool{
		"list":     true,
		"agent":    true,
		"pre-push": true,
	}

	for _, cmd := range hookCmd.Commands() {
		name := cmd.Name()
		if utilitySubcmds[name] {
			continue
		}
		if _, ok := subcmdToEvent[name]; !ok {
			t.Errorf("subcommand %q has no corresponding EventType mapping", name)
		}
	}
}

// --- Bonus: Verify handler count matches expected after InitDependencies ---

func TestHookDepsWiring_HandlerCounts(t *testing.T) {
	// Not parallel: calls InitDependencies which writes to global deps.
	origDeps := deps
	defer func() { deps = origDeps }()

	InitDependencies()

	// Each new event should have exactly 1 handler registered via InitDependencies.
	singleHandlerEvents := []hook.EventType{
		hook.EventPostToolUseFailure,
		hook.EventNotification,
		hook.EventSubagentStart,
		hook.EventUserPromptSubmit,
		hook.EventPermissionRequest,
		hook.EventTeammateIdle,
		hook.EventTaskCompleted,
		hook.EventWorktreeCreate,
		hook.EventWorktreeRemove,
	}

	for _, event := range singleHandlerEvents {
		handlers := deps.HookRegistry.Handlers(event)
		if len(handlers) != 1 {
			t.Errorf("event %q: got %d handlers, want exactly 1", event, len(handlers))
		}
	}

	// SessionStart may have multiple handlers (session start + auto-update + optional rank).
	sessionStartHandlers := deps.HookRegistry.Handlers(hook.EventSessionStart)
	if len(sessionStartHandlers) < 2 {
		t.Errorf("event %q: got %d handlers, want at least 2 (session start + auto-update)",
			hook.EventSessionStart, len(sessionStartHandlers))
	}

	// PreToolUse should have exactly 1 handler.
	preToolHandlers := deps.HookRegistry.Handlers(hook.EventPreToolUse)
	if len(preToolHandlers) != 1 {
		t.Errorf("event %q: got %d handlers, want 1", hook.EventPreToolUse, len(preToolHandlers))
	}

	// PostToolUse should have exactly 1 handler.
	postToolHandlers := deps.HookRegistry.Handlers(hook.EventPostToolUse)
	if len(postToolHandlers) != 1 {
		t.Errorf("event %q: got %d handlers, want 1", hook.EventPostToolUse, len(postToolHandlers))
	}
}

// --- Bonus: Verify original (pre-existing) subcommands also map correctly ---

func TestHookSubcommands_OriginalEventTypeMapping(t *testing.T) {
	// Not parallel: subtests write to global deps sequentially.
	origDeps := deps
	defer func() { deps = origDeps }()

	expected := map[string]hook.EventType{
		"session-start": hook.EventSessionStart,
		"session-end":   hook.EventSessionEnd,
		"pre-tool":      hook.EventPreToolUse,
		"post-tool":     hook.EventPostToolUse,
		"stop":          hook.EventStop,
		"compact":       hook.EventPreCompact,
	}

	for subcmdName, wantEvent := range expected {
		t.Run(subcmdName, func(t *testing.T) {
			spy := &spyRegistry{}
			deps = &Dependencies{
				HookRegistry: spy,
				HookProtocol: &spyProtocol{},
			}

			var found bool
			for _, cmd := range hookCmd.Commands() {
				if cmd.Name() == subcmdName {
					found = true
					cmd.SetContext(context.Background())
					err := cmd.RunE(cmd, []string{})
					if err != nil {
						t.Fatalf("RunE for %q returned error: %v", subcmdName, err)
					}
					break
				}
			}
			if !found {
				t.Fatalf("subcommand %q not found in hookCmd", subcmdName)
			}

			if len(spy.dispatched) != 1 {
				t.Fatalf("expected 1 dispatch call, got %d", len(spy.dispatched))
			}
			if spy.dispatched[0] != wantEvent {
				t.Errorf("subcommand %q dispatched EventType %q, want %q",
					subcmdName, spy.dispatched[0], wantEvent)
			}
		})
	}
}
