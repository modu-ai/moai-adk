// Package harness — Context-Governance Axis weight-recording tests.
// SPEC-V3R6-CONTEXT-GOV-AXIS-001 REQ-CGA-001/002/003.
//
// Covers: (1) new weight fields present on fresh v2.1 lines; (2) legacy v1 AND v2
// lines parse without error and weight fields resolve to sentinel; (3) fail-open
// path — estimation skip leaves weight fields at zero value while schema_version
// is still stamped "v2.1".
package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestLogSchemaVersionBumpedToV21 verifies the schema_version constant was
// bumped from "v2" to "v2.1" per REQ-CGA-002 (NOT v1→v1.1 — v2 is the baseline).
func TestLogSchemaVersionBumpedToV21(t *testing.T) {
	t.Parallel()
	if LogSchemaVersion != "v2.1" {
		t.Errorf("LogSchemaVersion: got=%q, want=%q (REQ-CGA-002 bump v2→v2.1)", LogSchemaVersion, "v2.1")
	}
}

// TestEvent_WeightFieldsPresentOnV21 verifies that a fresh v2.1 Event carrying
// the weight fields serializes all three (eager_context_weight,
// on_demand_context_weight, weight_unit) per REQ-CGA-001.
func TestEvent_WeightFieldsPresentOnV21(t *testing.T) {
	t.Parallel()

	evt := Event{
		EagerContextWeight:    42000,
		OnDemandContextWeight: 0,
		WeightUnit:            WeightUnitBytes,
		SchemaVersion:         LogSchemaVersion,
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	for _, field := range []string{"eager_context_weight", "weight_unit"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("weight field %q missing from v2.1 serialization (REQ-CGA-001)", field)
		}
	}
	// on_demand_context_weight is 0 here and omitempty drops it — that is the
	// documented sentinel behavior (schema_version=v2.1 still distinguishes it).
	if _, ok := raw["on_demand_context_weight"]; ok {
		t.Errorf("on_demand_context_weight should be omitempty-dropped when 0")
	}
}

// TestParseOldLogLinesNoCrash verifies the backward-compatibility gate (AC-CGA-002):
// legacy v1 AND v2 fixture lines (neither has weight fields) MUST parse without
// error under the extended reader, and the weight fields MUST resolve to the Go
// zero-value sentinel (0/"") on BOTH legacy schemas.
func TestParseOldLogLinesNoCrash(t *testing.T) {
	t.Parallel()

	// Fixtures captured from the live .moai/harness/usage-log.jsonl (plan-phase
	// 2026-06-18): 508 v1 lines + 100 v2 lines. These three representative lines
	// exercise both legacy schemas.
	fixtures := []struct {
		name string
		line string
		wantSchemaVersion string
	}{
		{
			name: "legacy v1 user_prompt line",
			line: `{"timestamp":"2026-05-19T09:22:34.7666Z","event_type":"user_prompt","subject":"","context_hash":"","tier_increment":0,"schema_version":"v1","prompt_hash":"ced7535932b4ffa3","prompt_len":383,"prompt_lang":"en"}`,
			wantSchemaVersion: "v1",
		},
		{
			name: "legacy v1 session_stop line",
			line: `{"timestamp":"2026-05-20T10:00:00Z","event_type":"session_stop","subject":"SPEC-X","context_hash":"","tier_increment":0,"schema_version":"v1","last_assistant_message_hash":"abc","last_assistant_message_len":100}`,
			wantSchemaVersion: "v1",
		},
		{
			name: "legacy v2 user_prompt line",
			line: `{"timestamp":"2026-06-10T09:22:34.7666Z","event_type":"user_prompt","subject":"","context_hash":"","tier_increment":0,"schema_version":"v2","prompt_hash":"fa160429b1c38a5e","prompt_len":1040,"prompt_lang":"en"}`,
			wantSchemaVersion: "v2",
		},
		{
			name: "legacy v2 subagent_stop line",
			line: `{"timestamp":"2026-06-11T10:00:00Z","event_type":"subagent_stop","subject":"unknown","context_hash":"","tier_increment":0,"schema_version":"v2","agent_id":"ag-123"}`,
			wantSchemaVersion: "v2",
		},
	}

	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var evt Event
			if err := json.Unmarshal([]byte(tc.line), &evt); err != nil {
				t.Fatalf("legacy %s line failed to parse (AC-CGA-002 backward-compat): %v", tc.wantSchemaVersion, err)
			}
			if evt.SchemaVersion != tc.wantSchemaVersion {
				t.Errorf("schema_version: got=%q, want=%q", evt.SchemaVersion, tc.wantSchemaVersion)
			}
			// Weight fields MUST resolve to sentinel on legacy lines.
			if evt.EagerContextWeight != 0 {
				t.Errorf("legacy %s line: eager_context_weight should be sentinel 0, got %d", tc.wantSchemaVersion, evt.EagerContextWeight)
			}
			if evt.OnDemandContextWeight != 0 {
				t.Errorf("legacy %s line: on_demand_context_weight should be sentinel 0, got %d", tc.wantSchemaVersion, evt.OnDemandContextWeight)
			}
			if evt.WeightUnit != "" {
				t.Errorf("legacy %s line: weight_unit should be sentinel empty, got %q", tc.wantSchemaVersion, evt.WeightUnit)
			}
		})
	}
}

// TestEstimateContextWeight_PopulatesFields verifies that the estimator
// populates the three weight fields on a real project tree (REQ-CGA-001).
func TestEstimateContextWeight_PopulatesFields(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Create the eager source files the estimator walks.
	if err := os.MkdirAll(filepath.Join(root, ".claude", "output-styles", "moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("CLAUDE body"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".claude", "output-styles", "moai", "moai.md"), []byte("moai body"), 0o644); err != nil {
		t.Fatal(err)
	}

	evt := &Event{}
	EstimateContextWeight(evt, root)

	if evt.EagerContextWeight <= 0 {
		t.Errorf("eager_context_weight should be > 0 after estimation, got %d", evt.EagerContextWeight)
	}
	if evt.WeightUnit != WeightUnitBytes {
		t.Errorf("weight_unit: got=%q, want=%q", evt.WeightUnit, WeightUnitBytes)
	}
	if evt.OnDemandContextWeight != 0 {
		t.Errorf("on_demand_context_weight should be 0 (§X.6 deferred), got %d", evt.OnDemandContextWeight)
	}
}

// TestEstimateContextWeight_FailOpenOnEmptyRoot verifies the fail-open path
// (REQ-CGA-003): when estimation cannot proceed (empty project root), the
// weight fields are left at the Go zero value (sentinel). The caller still
// stamps schema_version="v2.1" — distinguishing estimation-skipped from legacy.
func TestEstimateContextWeight_FailOpenOnEmptyRoot(t *testing.T) {
	t.Parallel()

	evt := &Event{}
	EstimateContextWeight(evt, "") // empty root → fail-open

	if evt.EagerContextWeight != 0 {
		t.Errorf("fail-open: eager_context_weight should be sentinel 0, got %d", evt.EagerContextWeight)
	}
	if evt.OnDemandContextWeight != 0 {
		t.Errorf("fail-open: on_demand_context_weight should be sentinel 0, got %d", evt.OnDemandContextWeight)
	}
	if evt.WeightUnit != "" {
		t.Errorf("fail-open: weight_unit should be sentinel empty, got %q", evt.WeightUnit)
	}
}

// TestEstimateContextWeight_NilEventSafe verifies the estimator does not panic
// on a nil Event pointer (defensive fail-open).
func TestEstimateContextWeight_NilEventSafe(t *testing.T) {
	t.Parallel()
	// Must not panic.
	EstimateContextWeight(nil, "/tmp")
}

// TestEstimateContextWeight_SkipsMissingFile verifies EC-1: a missing eager
// source file is skipped (not an error); the remaining sources are still summed.
func TestEstimateContextWeight_SkipsMissingFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Create only CLAUDE.md; leave moai.md + MEMORY.md absent.
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("only file"), 0o644); err != nil {
		t.Fatal(err)
	}

	evt := &Event{}
	EstimateContextWeight(evt, root)

	if evt.EagerContextWeight != len("only file") {
		t.Errorf("EC-1 skip-missing: eager weight should be %d (only CLAUDE.md), got %d", len("only file"), evt.EagerContextWeight)
	}
}

// TestRecordExtendedEvent_V21SchemaStamp verifies that RecordExtendedEvent
// stamps schema_version="v2.1" even when the caller leaves the weight fields
// unset (fail-open sentinel case: v2.1 + zero-value weight fields).
func TestRecordExtendedEvent_V21SchemaStamp(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	logPath := filepath.Join(root, "usage-log.jsonl")
	obs := NewObserver(logPath)

	// Event with NO weight fields set — simulates estimation-skipped fail-open.
	evt := Event{
		EventType: EventTypeSessionStop,
		Subject:   "SPEC-TEST",
	}
	if err := obs.RecordExtendedEvent(evt); err != nil {
		t.Fatalf("RecordExtendedEvent failed: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}
	if decoded.SchemaVersion != "v2.1" {
		t.Errorf("schema_version: got=%q, want=v2.1 (fail-open must still stamp v2.1)", decoded.SchemaVersion)
	}
	// Weight fields at sentinel — reader distinguishes via schema_version branching.
	if decoded.EagerContextWeight != 0 || decoded.OnDemandContextWeight != 0 || decoded.WeightUnit != "" {
		t.Errorf("fail-open: weight fields should be sentinel, got eager=%d on_demand=%d unit=%q",
			decoded.EagerContextWeight, decoded.OnDemandContextWeight, decoded.WeightUnit)
	}
}
