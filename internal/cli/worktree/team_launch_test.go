// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M1 pattern dispatcher scaffolding tests.
//
// Covers `decidePattern` pure function and `Pattern` enum stringer (M1).
// POSIX-specific launchP3 execution path tests live in team_launch_posix_test.go
// (build tag !windows). Tmux pane spawn tests (P1/P2) land in M3.

package worktree

import "testing"

// TestDecidePattern verifies the pure pattern decision matrix from the spec
// (REQ-WTL-001..005). No I/O, no stubs — pure boolean truth table.
func TestDecidePattern(t *testing.T) {
	tests := []struct {
		name     string
		teamFlag bool
		inTmux   bool
		cgMode   bool
		want     Pattern
	}{
		{
			name:     "no team flag → P4 handoff",
			teamFlag: false,
			inTmux:   true,
			cgMode:   true,
			want:     PatternP4Handoff,
		},
		{
			name:     "no team flag + no tmux → P4 handoff",
			teamFlag: false,
			inTmux:   false,
			cgMode:   false,
			want:     PatternP4Handoff,
		},
		{
			name:     "team flag + no tmux → P3 in-progress",
			teamFlag: true,
			inTmux:   false,
			cgMode:   false,
			want:     PatternP3InProgress,
		},
		{
			name:     "team flag + tmux + no CG → P2 tmux CC",
			teamFlag: true,
			inTmux:   true,
			cgMode:   false,
			want:     PatternP2TmuxCC,
		},
		{
			name:     "team flag + tmux + CG → P1 tmux GLM",
			teamFlag: true,
			inTmux:   true,
			cgMode:   true,
			want:     PatternP1TmuxGLM,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decidePattern(tt.teamFlag, tt.inTmux, tt.cgMode)
			if got != tt.want {
				t.Errorf("decidePattern(team=%v, tmux=%v, cg=%v) = %v, want %v",
					tt.teamFlag, tt.inTmux, tt.cgMode, got, tt.want)
			}
		})
	}
}

// TestPatternString verifies the Pattern.String() stringer returns stable
// human-readable identifiers used in logs and registry entries.
func TestPatternString(t *testing.T) {
	tests := []struct {
		p    Pattern
		want string
	}{
		{PatternP4Handoff, "P4-handoff"},
		{PatternP3InProgress, "P3-in-progress"},
		{PatternP2TmuxCC, "P2-tmux-cc"},
		{PatternP1TmuxGLM, "P1-tmux-glm"},
		{Pattern(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("Pattern(%d).String() = %q, want %q", int(tt.p), got, tt.want)
			}
		})
	}
}

// TestTeamLaunchConfig_ZeroValue verifies the zero-value TeamLaunchConfig is
// sensible (P4Handoff = 0, empty paths). M2+ extends with execution tests.
func TestTeamLaunchConfig_ZeroValue(t *testing.T) {
	var cfg TeamLaunchConfig
	if cfg.Pattern != PatternP4Handoff {
		t.Errorf("zero-value Pattern = %v, want PatternP4Handoff", cfg.Pattern)
	}
	if cfg.WorktreePath != "" {
		t.Errorf("zero-value WorktreePath = %q, want empty", cfg.WorktreePath)
	}
	if cfg.LLM != "" {
		t.Errorf("zero-value LLM = %q, want empty", cfg.LLM)
	}
}
