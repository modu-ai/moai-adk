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
//
// After SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 this is the ONLY default path —
// no preset participates in defaulting (REQ-SPR-009).
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

// TestSyncStatuslineThemeOnlyPreservesSegments verifies REQ-SPR-008: a theme-only
// save (no segments submitted) preserves the existing persisted segments map
// untouched. This is the profile-level guard backing the web theme-only
// round-trip characterization gate (integration_test.go:124-165).
//
// The seed YAML carries a legacy `preset: custom` key which the loader now
// silently ignores (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 REQ-SPR-021); the
// explicit `segments:` block is the source of truth and must survive verbatim.
func TestSyncStatuslineThemeOnlyPreservesSegments(t *testing.T) {
	projectRoot := t.TempDir()
	setupProjectConfig(t, projectRoot)

	// Seed an existing statusline.yaml with a hand-picked custom segment map.
	// The legacy preset: key is retained in the seed to exercise silent-ignore.
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	existing := "statusline:\n  preset: custom\n  theme: catppuccin-mocha\n  segments:\n    model: true\n    context: false\n    git_status: true\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "statusline.yaml"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write statusline.yaml: %v", err)
	}

	// Theme-only save: no segments submitted.
	if err := SyncToProjectConfig(projectRoot, ProfilePreferences{StatuslineTheme: "catppuccin-latte"}); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	segs := readStatuslineSegments(t, projectRoot)
	// The existing hand-picked segments must be preserved verbatim (REQ-SPR-008).
	if !segs["model"] || segs["context"] || !segs["git_status"] {
		t.Errorf("theme-only save did not preserve existing segments (REQ-SPR-008): got %v", segs)
	}
	if len(segs) != 3 {
		t.Errorf("theme-only save altered the segment key count: got %d keys, want 3 (preserved)", len(segs))
	}
}
