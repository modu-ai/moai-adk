package profile

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestDefaultStatuslineSegments verifies that defaultStatuslineSegments returns
// the canonical 15-key segment map (SLM-5 fix). Before the fix the seed emitted
// only 11 of 15 segments (missing effort_thinking, worktree, task, pr), so a
// statusline.yaml created on first sync silently dropped those four segments.
func TestDefaultStatuslineSegments(t *testing.T) {
	got := defaultStatuslineSegments()

	want := []string{
		"model", "context", "output_style", "claude_version", "moai_version",
		"session_time", "effort_thinking", "usage_5h", "usage_7d", "directory",
		"git_status", "git_branch", "worktree", "task", "pr",
	}

	if len(got) != len(want) {
		t.Fatalf("defaultStatuslineSegments() returned %d keys, want %d", len(got), len(want))
	}

	for _, key := range want {
		val, ok := got[key]
		if !ok {
			t.Errorf("defaultStatuslineSegments() missing canonical key %q", key)
			continue
		}
		if !val {
			t.Errorf("defaultStatuslineSegments()[%q] = false, want true (all default-on)", key)
		}
	}

	// SegmentRepo ("repo") is intentionally outside the 15-key schema (SLM-7).
	if _, ok := got["repo"]; ok {
		t.Errorf("defaultStatuslineSegments() includes \"repo\" — SegmentRepo is intentionally excluded (SLM-7)")
	}
}

// readStatuslineSegments reads .moai/config/sections/statusline.yaml and returns
// the persisted segments map (nil if absent).
func readStatuslineSegments(t *testing.T, projectRoot string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectRoot, ".moai", "config", "sections", "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}
	var wrapper statuslineFileWrapper
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("unmarshal statusline.yaml: %v", err)
	}
	return wrapper.Statusline.Segments
}

// TestSyncStatuslinePresetExpand verifies the SLR-3 fix: saving a non-custom
// preset (compact/minimal) expands it into the full 15-key segments map at write
// time, so the preset selection takes effect at render time instead of silently
// no-opping. Before the fix the segments map was left unchanged, so the runtime
// (which reads segments, not preset) ignored the preset.
func TestSyncStatuslinePresetExpand(t *testing.T) {
	tests := []struct {
		name        string
		preset      string
		wantEnabled map[string]bool // subset assertions over the 15-key map
	}{
		{
			name:   "compact expands to essentials + workflow context",
			preset: "compact",
			wantEnabled: map[string]bool{
				"model": true, "context": true, "git_status": true,
				"git_branch": true, "task": true, "pr": true,
				// off in compact:
				"output_style": false, "directory": false, "usage_5h": false,
			},
		},
		{
			name:   "minimal expands to model + context only",
			preset: "minimal",
			wantEnabled: map[string]bool{
				"model": true, "context": true,
				"git_status": false, "task": false, "pr": false, "directory": false,
			},
		},
		{
			name:   "full expands to all 15 enabled",
			preset: "full",
			wantEnabled: map[string]bool{
				"model": true, "context": true, "pr": true, "task": true,
				"worktree": true, "effort_thinking": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := t.TempDir()
			setupProjectConfig(t, projectRoot)

			// Save a non-custom preset with NO explicit segments (the SLR-3 path).
			prefs := ProfilePreferences{StatuslinePreset: tt.preset}
			if err := SyncToProjectConfig(projectRoot, prefs); err != nil {
				t.Fatalf("SyncToProjectConfig: %v", err)
			}

			segs := readStatuslineSegments(t, projectRoot)
			if segs == nil {
				t.Fatal("segments map is nil — preset write-effective did not materialize a map (SLR-3 no-op)")
			}
			if len(segs) != 15 {
				t.Errorf("segments map has %d keys, want 15 (preset must expand to the full canonical map)", len(segs))
			}
			for key, want := range tt.wantEnabled {
				got, ok := segs[key]
				if !ok {
					t.Errorf("segments[%q] missing — preset %q did not expand the full key set", key, tt.preset)
					continue
				}
				if got != want {
					t.Errorf("preset %q: segments[%q] = %v, want %v", tt.preset, key, got, want)
				}
			}
		})
	}
}

// TestStatuslinePresetEffectiveReloadRoundTrip verifies the SLR-3 fix survives a
// reload: a compact preset saved without explicit segments persists the expanded
// 15-key map, and re-reading the file confirms the preset is no longer a no-op.
func TestStatuslinePresetEffectiveReloadRoundTrip(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	if err := SyncToProjectConfig(projectRoot, ProfilePreferences{StatuslinePreset: "compact"}); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	segs := readStatuslineSegments(t, projectRoot)
	// compact disables output_style; full would have it true. This proves the
	// preset materially shaped the persisted map (not a full-default no-op).
	if segs["output_style"] {
		t.Error("compact preset did not take effect: output_style is true (expected false in compact)")
	}
	if !segs["task"] || !segs["pr"] {
		t.Error("compact preset lost workflow-context segments (task/pr should be true)")
	}
}

// TestSyncStatuslineThemeOnlyPreservesSegments verifies HARD-7: a theme-only save
// (no preset submitted) preserves the existing persisted segments map untouched.
// This is the profile-level guard backing the web theme-only round-trip
// characterization gate (integration_test.go:124-165).
func TestSyncStatuslineThemeOnlyPreservesSegments(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	// Seed an existing statusline.yaml with a hand-picked custom segment map.
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	existing := "statusline:\n  preset: custom\n  theme: catppuccin-mocha\n  segments:\n    model: true\n    context: false\n    git_status: true\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "statusline.yaml"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write statusline.yaml: %v", err)
	}

	// Theme-only save: no preset, no segments submitted.
	if err := SyncToProjectConfig(projectRoot, ProfilePreferences{StatuslineTheme: "catppuccin-latte"}); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	segs := readStatuslineSegments(t, projectRoot)
	// The existing hand-picked segments must be preserved verbatim (HARD-7).
	if !segs["model"] || segs["context"] || !segs["git_status"] {
		t.Errorf("theme-only save did not preserve existing segments (HARD-7): got %v", segs)
	}
	if len(segs) != 3 {
		t.Errorf("theme-only save altered the segment key count: got %d keys, want 3 (preserved)", len(segs))
	}
}
