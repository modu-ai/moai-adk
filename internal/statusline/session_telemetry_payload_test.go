// Package statusline tests for SPEC-SESSION-TELEMETRY-001 §B.1: the model and
// effort the per-session record carries, its tolerance of a record written at
// the previous schema version, and the write throttle under the widened
// payload (plan.md §F item 1).
package statusline

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildWith renders one statusline for sessionID with an explicit model display
// name and effort level, and returns the record the render persisted.
func buildWith(t *testing.T, projDir, sessionID, displayName, effort string) *SessionTelemetryRecord {
	t.Helper()
	t.Setenv("MOAI_STATUSLINE_CONTEXT_SIZE", "")

	in := StdinData{
		SessionID: sessionID,
		Workspace: &WorkspaceInfo{CurrentDir: projDir},
		ContextWindow: &ContextWindowInfo{
			ContextWindowSize: 256000,
			UsedPercentage:    new(90.0),
		},
	}
	if displayName != "" {
		in.Model = &ModelInfo{DisplayName: displayName}
	}
	if effort != "" {
		in.Effort = &EffortInfo{Level: effort}
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	b := &defaultBuilder{renderer: NewRenderer("default", true, nil), mode: ModeDefault}
	if _, err := b.Build(context.Background(), bytes.NewReader(raw)); err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	rec, err := ReadSessionTelemetry(usagePath(projDir, sessionID))
	if err != nil {
		t.Fatalf("ReadSessionTelemetry: %v", err)
	}
	return rec
}

// TestRecordedModelIsBackendResolved — AC-ST-003 (REQ-ST-004).
// The recorded model is the model the session actually runs: under a
// non-Claude backend the z.ai model name replaces the Claude display name, and
// the [1m] suffix is stripped (D-5 — resolveGLMModelName, the same resolver the
// render path already calls).
func TestRecordedModelIsBackendResolved(t *testing.T) {
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.3")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")

	rec := buildWith(t, t.TempDir(), "sess-glm", "Opus 5 (1M context)[1m]", "medium")
	if rec.Model != "glm-5.3" {
		t.Errorf("model = %q, want glm-5.3 (the backend-resolved model)", rec.Model)
	}
	if strings.Contains(rec.Model, "[1m]") {
		t.Errorf("model = %q, want no [1m] suffix", rec.Model)
	}
	if strings.Contains(strings.ToLower(rec.Model), "opus") {
		t.Errorf("model = %q, want the backend model rather than the Claude display name", rec.Model)
	}
	if rec.Effort != "medium" {
		t.Errorf("effort = %q, want medium", rec.Effort)
	}
}

// TestModelAndEffortRoundTripAndOmit — AC-ST-004 (REQ-ST-003).
// Present values round-trip unchanged; absent values are omitted from the
// marshalled JSON rather than defaulted, and the write still succeeds.
func TestModelAndEffortRoundTripAndOmit(t *testing.T) {
	t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_SONNET_MODEL", "")
	t.Setenv("ANTHROPIC_DEFAULT_HAIKU_MODEL", "")

	proj := t.TempDir()
	rec := buildWith(t, proj, "sess-present", "Opus 5 (1M context)", "xhigh")
	if rec.Model != "Opus 5 (1M context)" {
		t.Errorf("model = %q, want the display name round-tripped unchanged", rec.Model)
	}
	if rec.Effort != "xhigh" {
		t.Errorf("effort = %q, want xhigh", rec.Effort)
	}

	// Absent: a payload carrying neither omits both keys.
	proj2 := t.TempDir()
	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}
	writeContextUsage(proj2, "sess-absent", 7, m, handoffStageSoft, "", "")
	raw, err := os.ReadFile(usagePath(proj2, "sess-absent"))
	if err != nil {
		t.Fatalf("write must still succeed when model and effort are absent: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("record is not valid JSON: %v", err)
	}
	for _, key := range []string{"model", "effort"} {
		if _, ok := obj[key]; ok {
			t.Errorf("key %q present; an absent value must be omitted, not defaulted", key)
		}
	}
}

// TestReadsPreviousSchemaRecord — AC-ST-010 (REQ-ST-003).
// A record produced by the pre-change writer — marshalled here from the
// pre-change field set, so key order and indentation are the marshaller's own
// output rather than a hand-authored fixture — reads back with its context
// values intact and its model and effort reported as not recorded.
func TestReadsPreviousSchemaRecord(t *testing.T) {
	t.Parallel()

	// The pre-change record: schema v1, no model, no effort.
	type preChangeRecord struct {
		SchemaVersion     int     `json:"schema_version"`
		SessionID         string  `json:"session_id"`
		WriterPID         int     `json:"writer_pid"`
		CapturedAt        string  `json:"captured_at"`
		ContextWindowSize int     `json:"context_window_size"`
		TokensUsed        int     `json:"tokens_used"`
		RawPct            float64 `json:"raw_pct"`
		Stage             string  `json:"stage"`
		Band              string  `json:"band"`
	}
	data, err := json.MarshalIndent(&preChangeRecord{
		SchemaVersion:     1,
		SessionID:         "sess-v1",
		WriterPID:         42,
		CapturedAt:        "2026-08-17T00:00:00Z",
		ContextWindowSize: 1_000_000,
		TokensUsed:        500_000,
		RawPct:            50.0,
		Stage:             "soft",
		Band:              "large",
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "sess-v1.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	rec, err := ReadSessionTelemetry(path)
	if err != nil {
		t.Fatalf("a record at the previous schema version must still read: %v", err)
	}
	if rec.RawPct != 50.0 || rec.ContextWindowSize != 1_000_000 || rec.Stage != "soft" {
		t.Errorf("context values lost: %+v", rec)
	}
	if rec.Model != "" || rec.Effort != "" {
		t.Errorf("model/effort = %q/%q, want both reported as not recorded", rec.Model, rec.Effort)
	}
}

// TestThrottleUnaffectedByModelAndEffort — plan.md §F item 1, measured.
// Adding model and effort to the record could in principle turn every render
// into a write. Count writes across N renders whose context values, model, and
// effort are all unchanged: the throttle must still skip. The remaining halves
// are why the two fields stay IN the throttle payload — a changed model or
// effort must reach disk rather than sit stale behind an unrelated context
// value, a state REQ-ST-003's "not recorded" path does not cover.
func TestThrottleUnaffectedByModelAndEffort(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}
	path := usagePath(proj, "sess-throttle-me")

	writeContextUsage(proj, "sess-throttle-me", 1, m, handoffStageSoft, "Opus 5", "high")
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	old := st.ModTime().Add(-time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	const renders = 5
	for range renders {
		writeContextUsage(proj, "sess-throttle-me", 1, m, handoffStageSoft, "Opus 5", "high")
	}
	st2, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !st2.ModTime().Equal(old) {
		t.Errorf("%d unchanged renders rewrote the record; the widened payload defeated the throttle", renders)
	}

	// A changed effort must reach disk.
	writeContextUsage(proj, "sess-throttle-me", 1, m, handoffStageSoft, "Opus 5", "low")
	rec, err := ReadSessionTelemetry(path)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Effort != "low" {
		t.Errorf("effort = %q, want low — a changed value must not sit stale behind the throttle", rec.Effort)
	}

	// A changed model must reach disk too.
	writeContextUsage(proj, "sess-throttle-me", 1, m, handoffStageSoft, "glm-5.3", "low")
	rec, err = ReadSessionTelemetry(path)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Model != "glm-5.3" {
		t.Errorf("model = %q, want glm-5.3 — a changed value must not sit stale behind the throttle", rec.Model)
	}
}
