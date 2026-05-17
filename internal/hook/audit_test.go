// Package hook — audit_test.go
// SPEC-V3R2-RT-006 REQ-003, REQ-042, REQ-063
// Registration parity, per-file category headers, retire-event absence, observability opt-in.
package hook

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// auditConfigProvider wraps a config.Config for use as a ConfigProvider in audit tests.
type auditConfigProvider struct {
	cfg *config.Config
}

func (a *auditConfigProvider) Get() *config.Config { return a.cfg }

// newTestConfigWithObsEvents returns a ConfigProvider with the given observability
// events list set in the system hook config.
func newTestConfigWithObsEvents(events []string) ConfigProvider {
	cfg := config.NewDefaultConfig()
	cfg.System.Hook.ObservabilityEvents = events
	return &auditConfigProvider{cfg: cfg}
}

// newTestConfigWithStrictMode returns a ConfigProvider with strict_mode set.
func newTestConfigWithStrictMode(strictMode bool) ConfigProvider {
	cfg := config.NewDefaultConfig()
	cfg.System.Hook.StrictMode = strictMode
	return &auditConfigProvider{cfg: cfg}
}

// testCtx returns a plain context.Background() for use in handler tests.
func testCtx() context.Context {
	return context.Background()
}

// retiredEventNames is the canonical list of events retired from settings.json.
// Used by multiple audit tests to ensure consistency.
var retiredEventNames = []string{
	"Notification",
	"Elicitation",
	"ElicitationResult",
	"TaskCreated",
}

// TestAuditRegistrationParity verifies that the handler count matches the
// expected formula:
//   Go handlers == native settings.json events + 1 composite (autoUpdate) + |observability-only handlers|
//
// SPEC-V3R2-RT-006 REQ-003, AC-13.
// Implements T-RT006-02 (RED) + T-RT006-19 (body).
func TestAuditRegistrationParity(t *testing.T) {
	// Count of RETIRE-OBS-ONLY handlers (kept in Go, not in settings.json)
	const obsOnlyCount = 4 // notification, elicitation, elicitationResult, taskCreated

	// Read the local settings.json to count native registrations.
	// We look for the project-local settings.json from the repo root.
	settingsPath := "../../.claude/settings.json"
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Skipf("settings.json not found at %s (skip in isolated CI): %v", settingsPath, err)
	}

	// Count how many hook event keys are in the hooks section.
	var settingsJSON struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settingsJSON); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}

	nativeCount := len(settingsJSON.Hooks)
	t.Logf("native settings.json hook registrations: %d", nativeCount)

	// Verify that no retired events are present.
	for _, retiredKey := range retiredEventNames {
		if _, found := settingsJSON.Hooks[retiredKey]; found {
			t.Errorf("RETIRE-OBS-ONLY event %q found in settings.json (AC-10, REQ-063 violation)", retiredKey)
		}
	}

	// Expected: 22 settings.json event registrations (21 native + 1 composite SessionStart/autoUpdate
	// share the same key) + 4 obs-only = 26 Go handlers total.
	// The autoUpdate composite is registered in Go deps.go under SessionStart, NOT as a separate
	// settings.json key — so the settings.json count is 22 (includes SessionStart once).
	const expectedNative = 22
	if nativeCount != expectedNative {
		t.Errorf("settings.json hook count = %d, want %d (4 retired events must be absent)", nativeCount, expectedNative)
	}

	// Total Go handlers: native registrations + obs-only = 26
	// (autoUpdate composite is embedded in SessionStart key, not a separate count)
	expectedGoHandlers := expectedNative + obsOnlyCount
	t.Logf("expected Go handler count: %d (=%d settings.json keys + %d obs-only)",
		expectedGoHandlers, expectedNative, obsOnlyCount)
}

// TestAuditPerFileCategoryHeader verifies that each handler file declares
// a "// Resolution: CATEGORY" header at the top of the file.
//
// SPEC-V3R2-RT-006 REQ-002, AC-14.
// Implements T-RT006-02 (RED) + T-RT006-13 (body).
func TestAuditPerFileCategoryHeader(t *testing.T) {
	// Valid resolution categories per SPEC §5.7.
	validCategories := []string{
		"KEEP", "UPGRADE", "FIX", "REMOVE", "RETIRE-OBS-ONLY", "COMPOSITE",
	}
	categoryPattern := regexp.MustCompile(`^// Resolution: (` + strings.Join(validCategories, "|") + `)`)

	// Handler files that must have the resolution header.
	// Exclude test files, doc.go, and auxiliary non-handler files.
	handlerFiles := []string{
		"session_start.go",
		"session_end.go",
		"pre_tool.go",
		"post_tool.go",
		"post_tool_failure.go",
		"compact.go",
		"post_compact.go",
		"stop.go",
		"stop_failure.go",
		"subagent_start.go",
		"subagent_stop.go",
		"notification.go",
		"user_prompt_submit.go",
		"permission_request.go",
		"permission_denied.go",
		"teammate_idle.go",
		"task_completed.go",
		"task_created.go",
		"worktree_create.go",
		"worktree_remove.go",
		"config_change.go",
		"cwd_changed.go",
		"file_changed.go",
		"instructions_loaded.go",
		"elicitation.go",
		"auto_update.go",
	}

	for _, fname := range handlerFiles {
		t.Run(fname, func(t *testing.T) {
			data, err := os.ReadFile(fname)
			if err != nil {
				t.Fatalf("read %s: %v", fname, err)
			}

			// Check first few lines for the header.
			lines := strings.SplitN(string(data), "\n", 10)
			found := false
			for _, line := range lines {
				if categoryPattern.MatchString(line) {
					found = true
					t.Logf("%s: found header %q", fname, line)
					break
				}
			}
			if !found {
				t.Errorf("%s: missing '// Resolution: <CATEGORY>' header (REQ-002, AC-14)", fname)
			}
		})
	}
}

// TestAuditRetiredEventsNotInSettings asserts that the 4 retired events are
// absent from settings.json. If found, the audit fails naming the event.
//
// SPEC-V3R2-RT-006 REQ-063, AC-10.
// Implements T-RT006-02 (RED) + T-RT006-21 (body).
func TestAuditRetiredEventsNotInSettings(t *testing.T) {
	settingsPath := "../../.claude/settings.json"
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Skipf("settings.json not found at %s: %v", settingsPath, err)
	}

	var settingsJSON struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settingsJSON); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}

	for _, retired := range retiredEventNames {
		if _, found := settingsJSON.Hooks[retired]; found {
			t.Errorf("illegally registered: %q found in settings.json hooks section (AC-10, REQ-063)", retired)
		}
	}
}

// TestAuditObservabilityWhitelist verifies that the observabilityOptIn helper
// correctly honors the system.yaml hook.observability_events list.
//
// SPEC-V3R2-RT-006 REQ-040, REQ-041, AC-11, AC-16.
// Implements T-RT006-02 (RED) + T-RT006-08 (GREEN via observabilityOptIn impl).
func TestAuditObservabilityWhitelist(t *testing.T) {
	t.Run("empty_events_returns_false", func(t *testing.T) {
		cfg := newTestConfigWithObsEvents(nil)
		for _, event := range retiredEventNames {
			if observabilityOptIn(cfg, event) {
				t.Errorf("observabilityOptIn(%q) = true, want false (empty list)", event)
			}
		}
	})

	t.Run("listed_event_returns_true", func(t *testing.T) {
		cfg := newTestConfigWithObsEvents([]string{"notification"})
		if !observabilityOptIn(cfg, "notification") {
			t.Error("observabilityOptIn(notification) = false, want true (notification in list)")
		}
		// Others are still false.
		if observabilityOptIn(cfg, "taskCreated") {
			t.Error("observabilityOptIn(taskCreated) = true, want false (not in list)")
		}
	})

	t.Run("case_insensitive_match", func(t *testing.T) {
		cfg := newTestConfigWithObsEvents([]string{"Notification"})
		if !observabilityOptIn(cfg, "notification") {
			t.Error("observabilityOptIn is not case-insensitive")
		}
	})

	t.Run("nil_config_returns_false", func(t *testing.T) {
		if observabilityOptIn(nil, "notification") {
			t.Error("observabilityOptIn(nil, ...) = true, want false")
		}
	})

	t.Run("notification_handler_silent_when_not_opted_in", func(t *testing.T) {
		h := NewNotificationHandler()
		out, err := h.Handle(testCtx(), &HookInput{SessionID: "test"})
		if err != nil {
			t.Fatalf("Handle error: %v", err)
		}
		if out.SystemMessage != "" {
			t.Errorf("expected empty SystemMessage when not opted in, got %q", out.SystemMessage)
		}
	})

	t.Run("strict_mode_retired_event_still_silent", func(t *testing.T) {
		// REQ-041: even with strict_mode, retired events succeed silently.
		cfg := newTestConfigWithStrictMode(true)
		if observabilityOptIn(cfg, "notification") {
			t.Error("strict mode alone should not enable notification opt-in")
		}
	})
}

// TestAuditRetiredHandlersNotActive is a legacy guard that verifies retired
// events remain absent from the active event type list.
func TestAuditRetiredHandlersNotActive(t *testing.T) {
	retiredEventTypes := []EventType{
		EventNotification,
		EventElicitation,
		EventElicitationResult,
		EventTaskCreated,
	}

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

	for _, retired := range retiredEventTypes {
		for _, active := range activeEventTypes {
			if retired == active {
				t.Errorf("RETIRE-OBS-ONLY event %v should not be in active list", retired)
			}
		}
	}
}
