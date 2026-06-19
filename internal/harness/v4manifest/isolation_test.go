package v4manifest

import (
	"strings"
	"testing"
)

// TestDecideIsolation_ReadOnlySpecialistReturnsNone covers AC-HV4-007a: the
// ANALYZE phase is read-only, so its specialists MUST get isolation=none —
// zero worktrees are created for read-only analysis.
func TestDecideIsolation_ReadOnlySpecialistReturnsNone(t *testing.T) {
	input := IsolationInput{
		Role:          "codebase-explorer",
		TargetPaths:   []string{"internal/cli/", "internal/harness/"},
		Parallel:      true,
		ReadOnly:      true,
		Risky:         false,
	}
	got := DecideIsolation(input)
	if got.Isolation != IsolationNone {
		t.Errorf("DecideIsolation(read-only specialist) = %q, want %q (AC-HV4-007a: 0 worktrees for read-only ANALYZE)", got.Isolation, IsolationNone)
	}
	if !strings.Contains(got.Rationale, "read-only") {
		t.Errorf("DecideIsolation(read-only) rationale = %q, want substring 'read-only'", got.Rationale)
	}
}

// TestDecideIsolation_ConflictProneOverlappingPathsReturnsWorktree covers
// AC-HV4-007b: conflict-prone parallel generation targeting OVERLAPPING paths
// MUST get isolation=worktree (>= 1 worktree for conflict-prone GENERATE).
func TestDecideIsolation_ConflictProneOverlappingPathsReturnsWorktree(t *testing.T) {
	input := IsolationInput{
		Role:          "template-neutrality-auditor",
		TargetPaths:   []string{"internal/template/templates/.claude/", "internal/template/templates/.claude/skills/"},
		Parallel:      true,
		ReadOnly:      false,
		Risky:         false,
		OverlapsPeer:  true,
	}
	got := DecideIsolation(input)
	if got.Isolation != IsolationWorktree {
		t.Errorf("DecideIsolation(conflict-prone overlapping-paths) = %q, want %q (AC-HV4-007b: >= 1 worktree for conflict-prone GENERATE)", got.Isolation, IsolationWorktree)
	}
	if !strings.Contains(got.Rationale, "overlap") {
		t.Errorf("DecideIsolation(conflict-prone) rationale = %q, want substring 'overlap'", got.Rationale)
	}
}

// TestDecideIsolation_SequentialSinglePathReturnsNone covers the sequential
// single-path case: a specialist that runs alone (no parallelism, no overlap)
// targets a single path gets isolation=none — no write conflict is possible.
func TestDecideIsolation_SequentialSinglePathReturnsNone(t *testing.T) {
	input := IsolationInput{
		Role:          "manifest-writer",
		TargetPaths:   []string{".claude/commands/harness/dev/manifest.json"},
		Parallel:      false,
		ReadOnly:      false,
		Risky:         false,
		OverlapsPeer:  false,
	}
	got := DecideIsolation(input)
	if got.Isolation != IsolationNone {
		t.Errorf("DecideIsolation(sequential single-path) = %q, want %q (no parallel overlap → no worktree)", got.Isolation, IsolationNone)
	}
	if !strings.Contains(got.Rationale, "sequential") && !strings.Contains(got.Rationale, "no overlap") {
		t.Errorf("DecideIsolation(sequential) rationale = %q, want substring 'sequential' or 'no overlap'", got.Rationale)
	}
}

// TestDecideIsolation_RiskyChangeReturnsWorktree covers the risky-change
// carve-out: even a sequential specialist making a risky change (e.g. touching
// shared infrastructure) gets isolation=worktree per REQ-HV4-007 "risky changes".
func TestDecideIsolation_RiskyChangeReturnsWorktree(t *testing.T) {
	input := IsolationInput{
		Role:          "shared-config-refactorer",
		TargetPaths:   []string{".moai/config/sections/workflow.yaml"},
		Parallel:      false,
		ReadOnly:      false,
		Risky:         true,
		OverlapsPeer:  false,
	}
	got := DecideIsolation(input)
	if got.Isolation != IsolationWorktree {
		t.Errorf("DecideIsolation(risky change) = %q, want %q (REQ-HV4-007: risky changes → worktree)", got.Isolation, IsolationWorktree)
	}
	if !strings.Contains(got.Rationale, "risky") {
		t.Errorf("DecideIsolation(risky) rationale = %q, want substring 'risky'", got.Rationale)
	}
}

// TestDecideIsolation_ParallelNoOverlapReturnsNone covers the parallel-but-
// disjoint case: parallel specialists targeting DISJOINT paths (no overlap) get
// isolation=none — parallelism alone does not trigger worktree; overlap does.
func TestDecideIsolation_ParallelNoOverlapReturnsNone(t *testing.T) {
	input := IsolationInput{
		Role:          "backend-specialist",
		TargetPaths:   []string{"internal/backend/"},
		Parallel:      true,
		ReadOnly:      false,
		Risky:         false,
		OverlapsPeer:  false,
	}
	got := DecideIsolation(input)
	if got.Isolation != IsolationNone {
		t.Errorf("DecideIsolation(parallel disjoint paths) = %q, want %q (parallel but no overlap → no worktree)", got.Isolation, IsolationNone)
	}
}

// TestDecideIsolation_RationaleAlwaysNonEmpty ensures the rationale field is
// never empty (NFR-HV4-002 observability — every isolation decision is auditable).
func TestDecideIsolation_RationaleAlwaysNonEmpty(t *testing.T) {
	cases := []IsolationInput{
		{Role: "a", TargetPaths: []string{"x"}, ReadOnly: true},
		{Role: "b", TargetPaths: []string{"x", "x/sub"}, Parallel: true, OverlapsPeer: true},
		{Role: "c", TargetPaths: []string{"y"}, Parallel: false},
		{Role: "d", TargetPaths: []string{"z"}, Risky: true},
	}
	for i, in := range cases {
		got := DecideIsolation(in)
		if strings.TrimSpace(got.Rationale) == "" {
			t.Errorf("case %d: DecideIsolation rationale is empty (NFR-HV4-002 violated)", i)
		}
	}
}

// TestDecideIsolation_ReturnsOnlyNoneOrWorktree ensures the function never
// returns an isolation value outside the canonical 2-value set (design §C.2).
func TestDecideIsolation_ReturnsOnlyNoneOrWorktree(t *testing.T) {
	cases := []IsolationInput{
		{ReadOnly: true},
		{Parallel: true, OverlapsPeer: true},
		{Parallel: false},
		{Risky: true},
		{Parallel: true, OverlapsPeer: false},
	}
	for i, in := range cases {
		got := DecideIsolation(in)
		if got.Isolation != IsolationNone && got.Isolation != IsolationWorktree {
			t.Errorf("case %d: DecideIsolation = %q, want none|worktree only", i, got.Isolation)
		}
	}
}

// TestEmitCleanupDirective_FiresOnlyWhenWorktreeSpecialistPresent verifies
// M5.3's Runner cleanup-directive predicate: the directive fires ONLY when
// >= 1 specialist declared isolation=worktree. All-none rosters do NOT fire.
func TestEmitCleanupDirective_FiresOnlyWhenWorktreeSpecialistPresent(t *testing.T) {
	cases := []struct {
		name        string
		specialists []Specialist
		want        bool
	}{
		{
			name:        "empty roster does not fire",
			specialists: nil,
			want:        false,
		},
		{
			name: "all-none roster does not fire",
			specialists: []Specialist{
				{Role: "a", Isolation: IsolationNone},
				{Role: "b", Isolation: IsolationNone},
			},
			want: false,
		},
		{
			name: "one worktree specialist fires (AC-HV4-007b)",
			specialists: []Specialist{
				{Role: "a", Isolation: IsolationNone},
				{Role: "b", Isolation: IsolationWorktree},
			},
			want: true,
		},
		{
			name: "all-worktree roster fires",
			specialists: []Specialist{
				{Role: "a", Isolation: IsolationWorktree},
				{Role: "b", Isolation: IsolationWorktree},
			},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EmitCleanupDirective(tc.specialists)
			if got != tc.want {
				t.Errorf("EmitCleanupDirective(%v) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
