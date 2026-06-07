package profile

import "testing"

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
