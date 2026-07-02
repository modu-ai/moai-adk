package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTeammateIdleHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewTeammateIdleHandler()

	if got := h.EventType(); got != EventTeammateIdle {
		t.Errorf("EventType() = %q, want %q", got, EventTeammateIdle)
	}
}

func TestTeammateIdleHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *HookInput
		setupDir func(t *testing.T) string // returns projectDir to inject; "" means skip
		// wantBlock: true = keep-working (official decision:"block" + reason JSON, exit 0)
		wantBlock bool
	}{
		{
			name: "no team mode - always allow idle",
			input: &HookInput{
				SessionID:    "sess-ti-1",
				TeammateName: "worker-1",
				// TeamName is empty: not in team mode
			},
			wantBlock: false,
		},
		{
			name: "team mode with no project dir - allow idle (graceful degradation)",
			input: &HookInput{
				SessionID:    "sess-ti-2",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
				// CWD and ProjectDir both empty
			},
			wantBlock: false,
		},
		{
			name: "team mode with project dir but no baseline - allow idle",
			input: &HookInput{
				SessionID:    "sess-ti-3",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				// Create .moai/config/sections/ but no baseline file.
				if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantBlock: false,
		},
		{
			name: "team mode with baseline containing zero errors - allow idle",
			input: &HookInput{
				SessionID:    "sess-ti-4",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeQualityConfig(t, dir, true)
				writeBaseline(t, dir, map[string][]string{
					"file.go": {"warning", "information"},
				})
				return dir
			},
			wantBlock: false,
		},
		{
			name: "team mode with baseline containing errors exceeding threshold - block idle",
			input: &HookInput{
				SessionID:    "sess-ti-5",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeQualityConfig(t, dir, true)
				writeBaseline(t, dir, map[string][]string{
					"file.go":  {"error"},
					"other.go": {"error", "error"},
				})
				return dir
			},
			wantBlock: true,
		},
		{
			name: "team mode with coverage data meeting threshold - allow idle",
			input: &HookInput{
				SessionID:    "sess-ti-6",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeQualityConfig(t, dir, true)
				writeBaseline(t, dir, map[string][]string{
					"file.go": {"warning"},
				})
				writeCoverageData(t, dir, 90.0)
				return dir
			},
			wantBlock: false,
		},
		{
			name: "team mode with coverage data below threshold - block idle",
			input: &HookInput{
				SessionID:    "sess-ti-7",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeQualityConfig(t, dir, true)
				writeBaseline(t, dir, map[string][]string{
					"file.go": {"warning"},
				})
				writeCoverageData(t, dir, 50.0)
				return dir
			},
			wantBlock: true,
		},
		{
			name: "team mode with no coverage data - allow idle (graceful)",
			input: &HookInput{
				SessionID:    "sess-ti-8",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
			},
			setupDir: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				writeQualityConfig(t, dir, true)
				writeBaseline(t, dir, map[string][]string{
					"file.go": {"warning"},
				})
				// No coverage.json written - graceful degradation
				return dir
			},
			wantBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input := tt.input
			if tt.setupDir != nil {
				projectDir := tt.setupDir(t)
				// Clone input to avoid mutating shared struct.
				clone := *input
				clone.CWD = projectDir
				input = &clone
			}

			h := NewTeammateIdleHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			} else {
				gotBlock := got.Decision == DecisionBlock
				if gotBlock != tt.wantBlock {
					t.Errorf("block = %v (decision=%q), want %v", gotBlock, got.Decision, tt.wantBlock)
				}
				if tt.wantBlock && got.Reason == "" {
					t.Error("keep-working block output must carry a reason")
				}
				// Official channel is JSON with exit 0 — never ExitCode=2.
				if got.ExitCode != 0 {
					t.Errorf("ExitCode = %d, want 0 (block rides JSON)", got.ExitCode)
				}
			}
		})
	}
}

// TestTeammateIdleHandler_OfficialAgentTypeGate verifies the official stdin
// contract: TeammateIdle sends agent_type (the teammate name) — team_name is
// a legacy MoAI field the official runtime never sends. The quality gate must
// key on agent_type with team_name as legacy fallback.
func TestTeammateIdleHandler_OfficialAgentTypeGate(t *testing.T) {
	t.Parallel()

	h := NewTeammateIdleHandler()

	t.Run("agent_type alone activates team gate (official payload)", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeQualityConfig(t, dir, true)
		writeBaseline(t, dir, map[string][]string{
			"file.go": {"error", "error"},
		})
		got, err := h.Handle(context.Background(), &HookInput{
			SessionID: "sess-official-1",
			AgentType: "researcher", // official teammate-name field
			CWD:       dir,
			// TeamName intentionally absent (official runtime never sends it)
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || got.Decision != DecisionBlock {
			t.Fatalf("Decision = %+v, want block (agent_type must activate the gate)", got)
		}
		if got.Reason == "" {
			t.Error("block output must carry a reason")
		}
	})

	t.Run("legacy team_name still activates team gate", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeQualityConfig(t, dir, true)
		writeBaseline(t, dir, map[string][]string{
			"file.go": {"error", "error"},
		})
		got, err := h.Handle(context.Background(), &HookInput{
			SessionID:    "sess-legacy-1",
			TeamName:     "team-alpha",
			TeammateName: "worker-1",
			CWD:          dir,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || got.Decision != DecisionBlock {
			t.Fatalf("Decision = %+v, want block (legacy team_name fallback)", got)
		}
	})

	t.Run("neither field present - gate inactive", func(t *testing.T) {
		t.Parallel()
		got, err := h.Handle(context.Background(), &HookInput{
			SessionID: "sess-nogate-1",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || got.Decision != "" {
			t.Fatalf("Decision = %+v, want empty output (no team context)", got)
		}
	})
}

// writeQualityConfig writes a minimal quality.yaml that enables blocking on errors.
func writeQualityConfig(t *testing.T, projectDir string, blockOnError bool) {
	t.Helper()
	dir := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// blockOnError is derived from lsp_quality_gates.enabled and max_errors=0.
	var enabled string
	if blockOnError {
		enabled = "true"
	} else {
		enabled = "false"
	}
	content := "constitution:\n  lsp_quality_gates:\n    enabled: " + enabled + "\n    run:\n      max_errors: 0\n"
	if err := os.WriteFile(filepath.Join(dir, "quality.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeCoverageData writes a coverage.json file with the given coverage percentage.
func writeCoverageData(t *testing.T, projectDir string, percent float64) {
	t.Helper()
	stateDir := filepath.Join(projectDir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{
		"coverage_percent": percent,
		"updated_at":       "2026-02-19T10:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "coverage.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeBaseline writes a diagnostics baseline JSON file.
// filesSeverities maps file path (relative key, doesn't need to be real) to
// a list of severity strings per diagnostic.
func writeBaseline(t *testing.T, projectDir string, filesSeverities map[string][]string) {
	t.Helper()
	stateDir := filepath.Join(projectDir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	type diagEntry struct {
		Severity string `json:"severity"`
	}
	type fileEntry struct {
		Diagnostics []diagEntry `json:"diagnostics"`
	}
	type baselineDoc struct {
		Files map[string]fileEntry `json:"files"`
	}

	doc := baselineDoc{Files: make(map[string]fileEntry)}
	for path, severities := range filesSeverities {
		diags := make([]diagEntry, 0, len(severities))
		for _, s := range severities {
			diags = append(diags, diagEntry{Severity: s})
		}
		doc.Files[path] = fileEntry{Diagnostics: diags}
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "diagnostics-baseline.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}
