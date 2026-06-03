// Package cli — doctor_hook_test.go
// Tests for "moai doctor hook" subcommand.
// SPEC-V3R2-RT-006 REQ-050, REQ-051, AC-12.
package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestDoctorHook_27EventTableCount verifies that the coverage table has the expected
// number of entries after SPEC-V3R2-MIG-002 EventSetup retirement and the
// SPEC-HOOK-EVENT-REGISTRY-001 observe-only additions.
// Post-registry: 29 events + 1 composite (autoUpdate) = 30 entries total.
// AC-MIG002-A9: counts consistent with post-cleanup state.
func TestDoctorHook_27EventTableCount(t *testing.T) {
	// CoverageTable has 30 entries: 29 events + 1 composite (autoUpdate).
	// EventSetup row removed by SPEC-V3R2-MIG-002 M2.1; 3 observe-only rows
	// added by SPEC-HOOK-EVENT-REGISTRY-001.
	entries := buildDoctorHookEntries(false)
	if len(entries) != len(hook.CoverageTable) {
		t.Errorf("buildDoctorHookEntries() count = %d, want %d", len(entries), len(hook.CoverageTable))
	}

	// Verify exactly 29 canonical hook events (excluding composite).
	// EventSetup retired: SPEC-V3R2-MIG-002 M2.1 removed the REMOVE-resolution row;
	// 3 observe-only events added by SPEC-HOOK-EVENT-REGISTRY-001.
	eventCount := 0
	for _, e := range hook.CoverageTable {
		if e.Resolution != hook.ResolutionComposite {
			eventCount++
		}
	}
	if eventCount != 29 {
		t.Errorf("non-composite event count = %d, want 29 (3 observe-only events added by SPEC-HOOK-EVENT-REGISTRY-001)", eventCount)
	}
}

// TestDoctorHook_DefaultObservabilityEventsEmpty verifies that default
// CoverageTable entries have ObservabilityOptIn=false (empty list default).
// AC-16: "empty observability_events → handler returns HookOutput{} without logging".
func TestDoctorHook_DefaultObservabilityEventsEmpty(t *testing.T) {
	for _, e := range hook.CoverageTable {
		if e.ObservabilityOptIn {
			t.Errorf("event %q has ObservabilityOptIn=true in default table, want false", e.EventName)
		}
	}
}

// TestDoctorHook_JSONOutputParseable verifies that --json flag produces valid JSON.
// AC-12.
func TestDoctorHook_JSONOutputParseable(t *testing.T) {
	var buf bytes.Buffer
	entries := buildDoctorHookEntries(false)
	summary := hook.Summarize()

	if err := printDoctorHookJSON(&buf, entries, summary); err != nil {
		t.Fatalf("printDoctorHookJSON: %v", err)
	}

	output := buf.String()
	if !json.Valid([]byte(output)) {
		t.Errorf("JSON output is not valid: %s", output)
	}

	// Parse and verify coverage_table is non-empty.
	var parsed doctorHookOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(parsed.CoverageTable) == 0 {
		t.Error("coverage_table should be non-empty")
	}
	if parsed.Summary.Total == 0 {
		t.Error("summary.total should be non-zero")
	}
}

// TestDoctorHook_TraceNonexistentEventNoOp verifies that --trace for a
// nonexistent event name produces a graceful "no log lines found" message.
// AC-12: "non-existent event no-op".
func TestDoctorHook_TraceNonexistentEventNoOp(t *testing.T) {
	var buf bytes.Buffer
	// Use /dev/null as log path by passing a temp dir without hook.log.
	err := runDoctorHookTrace(&buf, "nonexistent-event-xyz")
	if err != nil {
		// err from runDoctorHookTrace is expected only for I/O errors, not missing file.
		t.Fatalf("runDoctorHookTrace: %v", err)
	}
	// Should contain "No log lines" or "No hook.log found".
	output := buf.String()
	if !strings.Contains(strings.ToLower(output), "no ") {
		t.Errorf("expected 'no ...' message for missing trace, got: %s", output)
	}
}

// TestDoctorHook_ObservabilityFilter verifies that --observability flag
// returns only RETIRE-OBS-ONLY events.
func TestDoctorHook_ObservabilityFilter(t *testing.T) {
	entries := buildDoctorHookEntries(true)
	for _, e := range entries {
		if e.Resolution != string(hook.ResolutionRetireObsOnly) {
			t.Errorf("observability filter returned non-RETIRE-OBS-ONLY entry: %s (%s)", e.EventName, e.Resolution)
		}
	}
	// Should have exactly 7 retired events (4 prior + 3 observe-only added by
	// SPEC-HOOK-EVENT-REGISTRY-001).
	if len(entries) != 7 {
		t.Errorf("observability filter count = %d, want 7", len(entries))
	}
}

// TestDoctorHook_SummaryCountsConsistent verifies that Summarize() produces
// consistent counts matching the CoverageTable.
func TestDoctorHook_SummaryCountsConsistent(t *testing.T) {
	summary := hook.Summarize()

	total := summary.Keep + summary.Upgrade + summary.Fix +
		summary.RetireObsOnly + summary.Remove + summary.Composite
	if total != summary.Total {
		t.Errorf("summary parts sum=%d, total=%d (inconsistent)", total, summary.Total)
	}

	// Verify specific expectations from SPEC §5.7.
	// RetireObsOnly: 4 prior + 3 observe-only added by SPEC-HOOK-EVENT-REGISTRY-001 = 7.
	if summary.RetireObsOnly != 7 {
		t.Errorf("RetireObsOnly count = %d, want 7", summary.RetireObsOnly)
	}
	if summary.Fix != 1 {
		t.Errorf("Fix count = %d, want 1 (subagentStop P-H02)", summary.Fix)
	}
	if summary.Remove != 0 {
		t.Errorf("Remove count = %d, want 0 (Setup row removed from CoverageTable by SPEC-V3R2-MIG-002 M2.1)", summary.Remove)
	}
	if summary.Composite != 1 {
		t.Errorf("Composite count = %d, want 1 (autoUpdate)", summary.Composite)
	}
}
