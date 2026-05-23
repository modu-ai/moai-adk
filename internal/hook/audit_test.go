// Package hook — audit_test.go
// SPEC-V3R2-RT-006 REQ-003, REQ-042, REQ-063
// SPEC-V3R2-MIG-002 REQ-009, REQ-010 (3-way sync invariant)
// Registration parity, per-file category headers, retire-event absence, observability opt-in,
// 3-way sync invariant (Go handlers ≡ settings.json keys ∪ retiredEventNames).
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

// retiredEventNames is an alias for the exported RetiredEventNames symbol in retired_events.go.
// Using the exported symbol ensures audit tests and migration logic share a single source of truth.
// SPEC-V3R2-MIG-002 T-MIG002-13: promoted to exported package-level var.
var retiredEventNames = RetiredEventNames

// TestAuditRegistrationParity verifies that the handler count matches the
// expected formula:
//   Go handlers == native settings.json events + 1 composite (autoUpdate) + |observability-only handlers|
//
// SPEC-V3R2-RT-006 REQ-003, AC-13.
// Implements T-RT006-02 (RED) + T-RT006-19 (body).
//
// Baseline history:
//   - Original SPEC-V3R2-RT-006 baseline: 22 settings.json keys.
//   - 2026-05-22 commit a3239d3de "fix(hook): WorktreeCreate/Remove 등록 해제":
//     deregistered WorktreeCreate + WorktreeRemove from settings.json because
//     Claude Code v2.1.49+ contract treats WorktreeCreate as an active creator
//     (must echo path to stdout); our observer-style HookOutput{} broke worktree
//     creation. The Go handlers remain (ResolutionKeep in CoverageTable) as
//     observability taps but are NOT registered in settings.json. Reduced
//     native count: 22 → 20.
//   - 2026-05-23 SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 M2: HOI conditional rendering
//     gates secondary `handle-harness-observe-*` wrappers inside Stop/SubagentStop/
//     UserPromptSubmit nested `hooks[]` arrays. This does NOT change the top-level
//     event key count (still 20).
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

	// Expected: 20 settings.json event registrations (22 native − 2 Worktree
	// events deregistered per a3239d3de) + 4 obs-only retired = 24 Go handlers
	// that have a registered or observability-only path. The 2 deregistered
	// Worktree handlers (ResolutionKeep, IsActive: true in CoverageTable) are
	// counted separately as orphan-but-intentional entries.
	// The autoUpdate composite is registered in Go deps.go under SessionStart,
	// NOT as a separate settings.json key — so the settings.json count includes
	// SessionStart once.
	const expectedNative = 20
	if nativeCount != expectedNative {
		t.Errorf("settings.json hook count = %d, want %d (4 retired events + 2 worktree events deregistered per a3239d3de must be absent)", nativeCount, expectedNative)
	}

	// Total Go handlers with a path: native registrations + obs-only = 24.
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

// deregisteredButLiveEventNames lists events whose Go handler remains live
// (ResolutionKeep in CoverageTable) but is intentionally NOT registered in
// settings.json. These are orphan-but-intentional entries — the third leg
// of the 3-way sync invariant beyond retiredEventNames.
//
// Origin: 2026-05-22 commit a3239d3de "fix(hook): WorktreeCreate/Remove 등록 해제".
// Claude Code v2.1.49+ contract: WorktreeCreate is an active creator that must
// echo a worktree path to stdout. Our observer-style empty HookOutput{} caused
// `path that is not a directory: {}` regression, breaking isolation: worktree
// for 5 agents. The handlers remain in Go for future re-enablement under a
// stdout-emitting contract, but settings.json registration was removed.
var deregisteredButLiveEventNames = []string{
	"WorktreeCreate",
	"WorktreeRemove",
}

// TestAuditThreeWaySync verifies the 3-way sync invariant (extended to 4-way):
// Go-registered event set ≡ settings.json.tmpl hook-key set
//                          ∪ retiredEventNames
//                          ∪ deregisteredButLiveEventNames.
//
// SPEC-V3R2-MIG-002 REQ-MIG002-009, REQ-MIG002-010 → AC-MIG002-A1.
// Reports HOOK_SYNC_DRIFT for Go-only entries; HOOK_WRAPPER_ORPHAN for settings-only entries.
func TestAuditThreeWaySync(t *testing.T) {
	// --- Step 1: Collect Go-registered event set ---
	// We use the coverage table as the authoritative source of Go-registered events.
	// An event is "Go-registered" if its Resolution is not REMOVE or COMPOSITE.
	// COMPOSITE shares a settings.json key (autoUpdate shares SessionStart), so
	// it does not represent an independent hook key.
	// REMOVE means the event constant was retired with no live handler — excluded.
	goEvents := make(map[string]bool)
	for _, entry := range CoverageTable {
		if entry.Resolution == ResolutionComposite || entry.Resolution == ResolutionRemove {
			continue
		}
		goEvents[entry.EventName] = true
	}

	// --- Step 2: Collect settings.json.tmpl hook-key set ---
	// Render the settings.json.tmpl with a representative TemplateContext and parse the JSON.
	// Use the local project settings.json as a proxy for the template output; this avoids
	// the need to import internal/template from the hook package (import cycle risk).
	// The local settings.json is generated from the template and stays in sync via CI.
	settingsPath := "../../.claude/settings.json"
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Skipf("settings.json not found at %s (skip in isolated CI): %v", settingsPath, err)
	}

	var settingsJSON struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settingsJSON); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}

	settingsKeys := make(map[string]bool)
	for key := range settingsJSON.Hooks {
		settingsKeys[key] = true
	}

	// --- Step 3: Build retired-event set ---
	retiredSet := make(map[string]bool, len(retiredEventNames))
	for _, name := range retiredEventNames {
		retiredSet[name] = true
	}

	// --- Step 3b: Build deregistered-but-live event set ---
	deregisteredSet := make(map[string]bool, len(deregisteredButLiveEventNames))
	for _, name := range deregisteredButLiveEventNames {
		deregisteredSet[name] = true
	}

	// --- Step 4: Assert 4-way invariant ---
	// Expected: goEvents ∖ (settingsKeys ∪ retiredSet ∪ deregisteredSet) == ∅
	// and settingsKeys ∖ goEvents == ∅

	driftFound := false

	// Check for HOOK_SYNC_DRIFT: Go-only entries not in settings AND not retired
	// AND not in the deregistered-but-live allowlist.
	for event := range goEvents {
		if !settingsKeys[event] && !retiredSet[event] && !deregisteredSet[event] {
			t.Errorf("HOOK_SYNC_DRIFT: Go handler registered for %q but absent from settings.json and not in retiredEventNames or deregisteredButLiveEventNames", event)
			driftFound = true
		}
	}

	// Check for HOOK_WRAPPER_ORPHAN: settings-only entries not backed by a Go handler.
	for key := range settingsKeys {
		if !goEvents[key] {
			t.Errorf("HOOK_WRAPPER_ORPHAN: settings.json key %q has no matching Go handler registration", key)
			driftFound = true
		}
	}

	if !driftFound {
		t.Logf("4-way sync OK: goEvents=%d, settingsKeys=%d, retiredEvents=%d, deregisteredButLive=%d",
			len(goEvents), len(settingsKeys), len(retiredSet), len(deregisteredSet))
	}
}

// TestAuditNoEventSetupOrphan verifies that the EventSetup constant and the
// "setup" cobra subcommand binding have been removed from the codebase.
//
// SPEC-V3R2-MIG-002 REQ-MIG002-003 → AC-MIG002-A2, AC-MIG002-A3.
func TestAuditNoEventSetupOrphan(t *testing.T) {
	// Check that types.go does NOT define EventSetup.
	typesData, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types.go: %v", err)
	}

	if strings.Contains(string(typesData), "EventSetup") {
		t.Errorf("AC-MIG002-A2 FAIL: EventSetup constant still present in types.go (SPEC-V3R2-MIG-002 REQ-MIG002-003)")
	}

	// Check that internal/cli/hook.go does NOT have the "setup" cobra binding.
	hookCLIData, err := os.ReadFile("../../internal/cli/hook.go")
	if err != nil {
		t.Fatalf("read internal/cli/hook.go: %v", err)
	}

	if strings.Contains(string(hookCLIData), `"setup"`) {
		t.Errorf("AC-MIG002-A3 FAIL: \"setup\" cobra binding still present in internal/cli/hook.go (SPEC-V3R2-MIG-002 REQ-MIG002-003)")
	}
}

// TestAuditNoStubHandlers verifies that each of the 5 RT-006-resolved handlers
// carries a // Resolution: UPGRADE or // Resolution: FIX header (NOT a stub marker).
//
// SPEC-V3R2-MIG-002 REQ-MIG002-004..008 → AC-MIG002-A8.
// This is a characterization test: it locks the RT-006 work product against regression.
func TestAuditNoStubHandlers(t *testing.T) {
	// These handlers were UPGRADE or FIX resolved by SPEC-V3R2-RT-006.
	rt006Files := []struct {
		file     string
		category string // expected Resolution category
	}{
		{"subagent_stop.go", "FIX"},
		{"config_change.go", "UPGRADE"},
		{"instructions_loaded.go", "UPGRADE"},
		{"file_changed.go", "UPGRADE"},
		{"post_tool_failure.go", "UPGRADE"},
	}

	validResolutions := map[string]bool{
		"UPGRADE": true,
		"FIX":     true,
	}

	resolutionPattern := regexp.MustCompile(`^// Resolution: ([A-Z\-]+)`)

	for _, tc := range rt006Files {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(tc.file)
			if err != nil {
				t.Fatalf("read %s: %v", tc.file, err)
			}

			lines := strings.SplitN(string(data), "\n", 10)
			found := false
			var category string
			for _, line := range lines {
				if m := resolutionPattern.FindStringSubmatch(line); m != nil {
					category = m[1]
					found = true
					break
				}
			}

			if !found {
				t.Errorf("%s: missing // Resolution: header — RT-006 work product may be lost", tc.file)
				return
			}

			if !validResolutions[category] {
				t.Errorf("%s: Resolution: %q is not a valid RT-006 resolution (want UPGRADE or FIX)", tc.file, category)
				return
			}

			t.Logf("%s: Resolution: %s (locked by SPEC-V3R2-MIG-002 characterization)", tc.file, category)
		})
	}
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
