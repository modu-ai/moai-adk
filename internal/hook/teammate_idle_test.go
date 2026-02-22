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
		name         string
		input        *HookInput
		setupDir     func(t *testing.T) string // returns projectDir to inject; "" means skip
		wantExitCode int
	}{
		{
			name: "no team mode - always allow idle",
			input: &HookInput{
				SessionID:    "sess-ti-1",
				TeammateName: "worker-1",
				// TeamName is empty: not in team mode
			},
			wantExitCode: 0,
		},
		{
			name: "team mode with no project dir - allow idle (graceful degradation)",
			input: &HookInput{
				SessionID:    "sess-ti-2",
				TeamName:     "team-alpha",
				TeammateName: "worker-1",
				// CWD and ProjectDir both empty
			},
			wantExitCode: 0,
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
			wantExitCode: 0,
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
			wantExitCode: 0,
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
			wantExitCode: 2,
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
			wantExitCode: 0,
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
			wantExitCode: 2,
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
			wantExitCode: 0,
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
			}
			if got.ExitCode != tt.wantExitCode {
				t.Errorf("ExitCode = %d, want %d", got.ExitCode, tt.wantExitCode)
			}
		})
	}
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
	memDir := filepath.Join(projectDir, ".moai", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{
		"coverage_percent": percent,
		"updated_at":       "2026-02-19T10:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(memDir, "coverage.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeBaseline writes a diagnostics baseline JSON file.
// filesSeverities maps file path (relative key, doesn't need to be real) to
// a list of severity strings per diagnostic.
func writeBaseline(t *testing.T, projectDir string, filesSeverities map[string][]string) {
	t.Helper()
	memDir := filepath.Join(projectDir, ".moai", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
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
	if err := os.WriteFile(filepath.Join(memDir, "diagnostics-baseline.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}
