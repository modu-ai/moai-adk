// Package v4manifest — conditional worktree-isolation decision (REQ-HV4-007).
//
// This file provides the Go-side DecideIsolation helper that the Builder PLAN
// phase consults when deciding each specialist's `isolation` field (none vs
// worktree). The decision is sub-agent-granular and conditional (design §G):
//
//   - read-only specialists (ANALYZE fan-out) → isolation=none (AC-HV4-007a:
//     0 worktrees for read-only analysis — no write conflict is possible)
//   - conflict-prone parallel generation targeting OVERLAPPING paths →
//     isolation=worktree (AC-HV4-007b: >= 1 worktree for conflict-prone GENERATE)
//   - sequential single-path specialists → isolation=none (no parallel overlap)
//   - risky changes (shared-infrastructure touches) → isolation=worktree
//     (REQ-HV4-007: "risky changes" carve-out)
//
// The decision is advisory: per C-HV4-006, L1 Agent(isolation:"worktree") is
// runtime-autonomous — DecideIsolation RECOMMENDS; the Claude Code runtime
// materializes (or not). The harness logic does NOT mandate worktree creation.
//
// The Runner template (runner_template.go) emits a worktree-cleanup directive
// at end-of-run when >= 1 specialist had isolation=worktree (EmitCleanupDirective).
package v4manifest

// IsolationInput characterizes a specialist's write-target profile so the
// isolation decision can be made from observable signals (design §G).
// The Builder PLAN phase populates this per specialist before calling
// DecideIsolation (REQ-HV4-007 / AC-HV4-007a/b).
type IsolationInput struct {
	// Role is the specialist's responsibility (e.g. "template-neutrality-auditor").
	// Recorded for audit; does not affect the decision.
	Role string

	// TargetPaths are the file/directory paths the specialist writes to.
	// Used to detect overlap with peer specialists (conflict-prone generation).
	TargetPaths []string

	// Parallel is true when the specialist runs concurrently with peers
	// (Fan-out/Fan-in, Expert Pool patterns). Sequential specialists (Pipeline)
	// set this false.
	Parallel bool

	// ReadOnly is true when the specialist performs only read operations
	// (ANALYZE Explore fan-out). Read-only specialists ALWAYS get isolation=none
	// — no write conflict is possible (AC-HV4-007a).
	ReadOnly bool

	// Risky is true when the specialist's change is risky (shared-infrastructure
	// touches, broad refactors). Risky changes get isolation=worktree per the
	// REQ-HV4-007 "risky changes" carve-out, even when sequential.
	Risky bool

	// OverlapsPeer is true when this specialist's TargetPaths overlap with at
	// least one other specialist's TargetPaths (conflict-prone parallel
	// generation). Overlap + Parallel → isolation=worktree (AC-HV4-007b).
	OverlapsPeer bool
}

// IsolationDecision is the result of the per-specialist isolation decision.
// It is returned by DecideIsolation and recorded in the manifest's specialist
// entry so the Runner dispatches verbatim (REQ-HV4-005 / AC-HV4-005b).
type IsolationDecision struct {
	// Isolation is "none" (main-tree) or "worktree"
	// (Agent(isolation:"worktree") sub-agent). Always one of the canonical
	// 2-value set (design §C.2).
	Isolation string

	// Rationale records the reason for the decision. Always non-empty
	// (NFR-HV4-002 observability — every isolation decision is auditable).
	Rationale string
}

// DecideIsolation applies the conditional worktree-isolation rule (REQ-HV4-007
// / AC-HV4-007a/b). The decision tree (design §G):
//
//  1. Read-only specialist → isolation=none (AC-HV4-007a: 0 worktrees for
//     read-only ANALYZE — no write conflict is possible).
//  2. Risky change → isolation=worktree (REQ-HV4-007 "risky changes" carve-out;
//     isolates blast radius even when sequential).
//  3. Parallel + overlaps peer → isolation=worktree (AC-HV4-007b: conflict-prone
//     parallel generation targeting overlapping paths).
//  4. Otherwise (sequential OR parallel-disjoint) → isolation=none (no overlap
//     → no write conflict).
//
// The decision is advisory (C-HV4-006): the orchestrator does NOT mandate
// worktree creation; L1 Agent(isolation:"worktree") is runtime-autonomous.
func DecideIsolation(in IsolationInput) IsolationDecision {
	// (1) Read-only → none. ANALYZE fan-out is always read-only (AC-HV4-007a).
	if in.ReadOnly {
		return IsolationDecision{
			Isolation: IsolationNone,
			Rationale: "read-only specialist, no write conflict possible (isolation=none per AC-HV4-007a)",
		}
	}

	// (2) Risky change → worktree (REQ-HV4-007 "risky changes" carve-out).
	if in.Risky {
		return IsolationDecision{
			Isolation: IsolationWorktree,
			Rationale: "risky change, blast-radius isolation (isolation=worktree per REQ-HV4-007 risky-changes carve-out)",
		}
	}

	// (3) Parallel + overlaps peer → worktree (AC-HV4-007b).
	if in.Parallel && in.OverlapsPeer {
		return IsolationDecision{
			Isolation: IsolationWorktree,
			Rationale: "conflict-prone parallel generation, target paths overlap with peer (isolation=worktree per AC-HV4-007b)",
		}
	}

	// (4) Default → none. Sequential specialists or parallel-disjoint specialists
	// have no write conflict (no overlap → no worktree).
	note := "sequential specialist, no parallel overlap"
	if in.Parallel {
		note = "parallel specialist, but target paths do not overlap with peers"
	}
	return IsolationDecision{
		Isolation: IsolationNone,
		Rationale: note + " (isolation=none — no write conflict possible)",
	}
}

// EmitCleanupDirective reports whether the Runner should emit a worktree-cleanup
// directive at end-of-run. Returns true when >= 1 specialist in the manifest
// declared isolation=worktree (REQ-HV4-007 / design §F + §G). L1 worktree
// cleanup itself is runtime-autonomous (C-HV4-006); this helper only decides
// whether the Runner EMITS the directive.
//
// The Runner template (runner_template.go) mirrors this logic in JS as
// `manifest.specialists.some(s => s.isolation === "worktree")`.
func EmitCleanupDirective(specialists []Specialist) bool {
	for _, s := range specialists {
		if s.Isolation == IsolationWorktree {
			return true
		}
	}
	return false
}
