# Progress — SPEC-V3R6-CG-MODE-HARDENING-001

## §E.1 Plan-phase Audit-Ready Signal

- **Tier**: M (justified in plan.md §A — multi-file + detector redesign + conditional template sync + security validation; not L, not S).
- **Requirements**: 10 (REQ-CGH-001 .. REQ-CGH-010), grouped: launch-safety (001), atomicity cluster (002/003/005), detector SSOT headline (006), doc (004), precondition (008), security (007), coverage (009), regression-safety (010).
- **Acceptance criteria**: 10 AC groups (AC-CGH-001 .. AC-CGH-010), each mechanically verifiable; 6 supporting edge cases (EC-1..EC-6).
- **Defect verification**: all CONFIRMED + POTENTIAL findings re-verified against cited source during plan-phase (spec.md §A.1 table). AGENT-REPORTED security finding (REQ-CGH-007) confirmed real at `glm.go:742-778` + `validation.go:349-352`. Disproven process-env-pollution hypothesis explicitly excluded (§A.2 / §H).
- **Sibling-SPEC reconciliation**: `cg_detect.go` / `REQ-WTL-009` owned by SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001; REQ-CGH-006 reconciles (does not delete), enforced by C-7 + AC-CGH-006 Scenario 6b.
- **Artifacts**: spec.md, plan.md, acceptance.md, design.md, progress.md (5-file Tier M set).
- **Frontmatter**: 12 canonical fields present; `status: draft`; `id` regex self-check PASS (decomposition printed in agent response).
- **Out of Scope**: 5 `### Out of Scope —` H3 sub-headings present in spec.md §H (satisfies OutOfScopeRule).

_Run-phase (§E.2/§E.3) and sync-phase (§E.4) sections below are placeholder headings only at plan-phase._

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
