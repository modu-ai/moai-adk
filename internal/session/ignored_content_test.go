package session

// ignored_content_test.go — SPEC-WORKTREE-REAPER-001 post-audit repair F3.
// The ignored-content decision now has two consumers (the PR-merge sweep and
// `worktree clean --stale`), so the predicate is bound here, once, rather than
// through either caller's fixtures.

import (
	"reflect"
	"testing"
)

func TestIrreplaceableIgnoredEntries(t *testing.T) {
	tests := []struct {
		name      string
		porcelain string
		want      []string
	}{
		{
			name:      "agent memory has no regenerator",
			porcelain: "!! .claude/agent-memory/\n",
			want:      []string{".claude/agent-memory"},
		},
		{
			name:      "an allowlisted tree is regenerable at any depth",
			porcelain: "!! .moai/state/verify/t209/\n!! bin/moai\n!! .claude/settings.local.json\n",
			want:      nil,
		},
		{
			name:      "tracked and untracked lines are not this guard's business",
			porcelain: " M internal/foo.go\n?? scratch.txt\n",
			want:      nil,
		},
		{
			name:      "a mixed listing reports only the irreplaceable entries",
			porcelain: "!! .moai/state/\n!! .moai/specs/SPEC-X/.moai/\n!! bin/\n",
			want:      []string{".moai/specs/SPEC-X/.moai"},
		},
		{
			name:      "quoted and trailing-slash forms normalise the same way",
			porcelain: "!! \".claude/agent-memory/\"\n",
			want:      []string{".claude/agent-memory"},
		},
		{
			name:      "a prefix that is not a path boundary is not allowlisted",
			porcelain: "!! bin-archive/\n",
			want:      []string{"bin-archive"},
		},
		{
			name:      "empty porcelain reports nothing",
			porcelain: "",
			want:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IrreplaceableIgnoredEntries(tt.porcelain)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IrreplaceableIgnoredEntries(%q) = %#v, want %#v", tt.porcelain, got, tt.want)
			}
		})
	}
}

func TestIsRegenerableIgnoredPath(t *testing.T) {
	for _, p := range RegenerableIgnoredPaths {
		if !IsRegenerableIgnoredPath(p) {
			t.Errorf("allowlisted path %q must be regenerable", p)
		}
		if !IsRegenerableIgnoredPath(p + "/nested/file") {
			t.Errorf("a path below allowlisted %q must be regenerable", p)
		}
	}
	// Negative control: an unclassified path preserves the tree (fail-closed).
	if IsRegenerableIgnoredPath(".claude/agent-memory") {
		t.Error("agent memory must never be classified regenerable")
	}
}
