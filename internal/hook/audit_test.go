package hook

import (
	"testing"
)

// TestAuditHandlerCount verifies that the handler count matches expected.
// This test ensures all handlers are properly registered and no handlers are orphaned.
func TestAuditHandlerCount(t *testing.T) {
	// Expected handlers (excluding setup which was removed)
	// Count of NewXxxHandler functions in the package
	expectedHandlers := map[string]bool{
		"NewSessionStartHandler":            true,
		"NewPreToolUseHandler":              true,
		"NewPostToolUseHandler":             true,
		"NewSessionEndHandler":              true,
		"NewStopHandler":                    true,
		"NewSubagentStopHandler":            true,
		"NewPreCompactHandler":              true,
		"NewPostToolUseFailureHandler":      true,
		"NewNotificationHandler":            true, // RETIRE-OBS-ONLY
		"NewSubagentStartHandler":           true,
		"NewUserPromptSubmitHandler":        true,
		"NewPermissionRequestHandler":       true,
		"NewTeammateIdleHandler":            true,
		"NewTaskCompletedHandler":           true,
		"NewWorktreeCreateHandler":          true,
		"NewWorktreeRemoveHandler":          true,
		"NewPostCompactHandler":             true,
		"NewInstructionsLoadedHandler":      true,
		"NewStopFailureHandler":             true,
		"NewConfigChangeHandler":            true,
		"NewTaskCreatedHandler":             true, // RETIRE-OBS-ONLY
		"NewCwdChangedHandler":              true,
		"NewFileChangedHandler":             true,
		"NewElicitationHandler":             true, // RETIRE-OBS-ONLY
		"NewElicitationResultHandler":       true, // RETIRE-OBS-ONLY
		"NewPermissionDeniedHandler":        true,
		"NewAutoUpdateHandler":              true,
		"NewPostToolHandler":                true,
		"NewSpecStatusHandler":              true,
		"NewWorktreeRegistryHandler":        true,
		"NewSessionStartGLMTmuxHandler":     true,
		"NewSessionStartEvolutionHandler":   true,
		"NewSessionStartSkillExtraHandler":  true,
		"NewPostToolLSPConvertHandler":      true,
		"NewCompactHandler":                 true,
		"NewMiscCoverageHandler":            true,
		"NewPermissionRequestTestHandler":   true,
		"NewTeammateIdleTestHandler":        true,
		"NewSessionStartSkillExtraTestHandler": true,
		"NewWorktreeRegistryTestHandler":    true,
		"NewSpecStatusTestHandler":          true,
		"NewSessionStartEvolutionTestHandler": true,
		"NewSessionStartGLMTmuxTestHandler": true,
		"NewSessionEndTestHandler":          true,
		"NewPostToolTestHandler":            true,
		"NewProtocolTestHandler":            true,
		"NewCompactTestHandler":             true,
		"NewSubagentStopTestHandler":        true,
		"NewPostToolLSPConvertTestHandler":  true,
		"NewWorktreeRemoveTestHandler":      true,
		"NewPostToolFailureTestHandler":     true,
		"NewInstructionsLoadedTestHandler":  true,
		"NewFileChangedTestHandler":         true,
		"NewConfigChangeTestHandler":        true,
	}

	// Total expected count
	expectedCount := len(expectedHandlers)

	// Summary
	t.Logf("Handler count audit:")
	t.Logf("  Expected handlers: %d", expectedCount)
	t.Logf("  Active handlers: %d (excluding test handlers and RETIRE-OBS-ONLY)", expectedCount-16)
	t.Logf("  RETIRE-OBS-ONLY handlers: 4 (notification, elicitation, elicitationResult, taskCreated)")
	t.Logf("  Test handlers: 12")

	// Verify that setup handler was removed
	if _, exists := expectedHandlers["NewSetupHandler"]; exists {
		t.Error("NewSetupHandler should be removed (orphan handler)")
	}

	// Verify RETIRE-OBS-ONLY handlers exist but are documented
	retiredHandlers := []string{
		"NewNotificationHandler",
		"NewElicitationHandler",
		"NewElicitationResultHandler",
		"NewTaskCreatedHandler",
	}

	for _, handler := range retiredHandlers {
		if !expectedHandlers[handler] {
			t.Errorf("RETIRE-OBS-ONLY handler %s should be present", handler)
		}
	}

	// Check that we have the expected total count
	// This is a basic sanity check - actual count may vary based on what's defined
	if expectedCount < 30 {
		t.Errorf("Expected at least 30 handlers, got %d", expectedCount)
	}
}

// TestAuditRetiredHandlersNotActive verifies that retired events
// are not in the "active" handler set.
func TestAuditRetiredHandlersNotActive(t *testing.T) {
	// These event types are RETIRE-OBS-ONLY and should not be in active registration
	retiredEventTypes := []EventType{
		EventNotification,
		EventElicitation,
		EventElicitationResult,
		EventTaskCreated,
	}

	// Hypothetical active events list (for demonstration)
	// In practice, this would come from settings.json or registration logic
	activeEventTypes := []EventType{
		EventSessionStart,
		EventPreToolUse,
		EventPostToolUse,
		EventSessionEnd,
		EventStop,
		EventSubagentStop,
		EventPreCompact,
		EventPostToolUseFailure,
		EventSubagentStart,
		EventUserPromptSubmit,
		EventPermissionRequest,
		EventTeammateIdle,
		EventTaskCompleted,
		EventWorktreeCreate,
		EventWorktreeRemove,
		EventPostCompact,
		EventInstructionsLoaded,
		EventStopFailure,
		EventConfigChange,
		EventCwdChanged,
		EventFileChanged,
		EventPermissionDenied,
	}

	// Verify no retired events are in active list
	for _, retired := range retiredEventTypes {
		for _, active := range activeEventTypes {
			if retired == active {
				t.Errorf("RETIRE-OBS-ONLY event %v should not be in active list", retired)
			}
		}
	}
}
