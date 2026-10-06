# SPEC-HOOK-BACKSLASH-SYMLINK-001 — Progress Record

status: draft (plan phase)
card: t1556 (factory lane-12)
worktree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1556 (branch WT-backslash-symlink, HEAD cad44a751, clean at authoring)

## Plan-phase Evidence (what was read, what was decided)

Read and verified in this tree (cad44a751):

- `internal/hook/pre_tool.go:1357-1480` — `resolveThroughExistingParent` /
  `resolvePhysicalWalk`: the t1530 physical-walk repair. Defect line confirmed at
  1397: `strings.Split(strings.ReplaceAll(filepath.ToSlash(p), "\\", "/"), "/")` —
  unconditional backslash→slash conversion before segmentation.
- `internal/hook/pre_tool.go:1487-1593` — `absoluteUncleaned` + `checkFileAccess`:
  boundary check, NFC normalization, project-root EvalSymlinks symmetry — untouched
  surfaces.
- `internal/hook/protected_zone_path.go:34-49` — `zoneSlash` carries the same
  unconditional conversion; classified OUT OF SCOPE (PowerShell display-form matching,
  separate follow-up), recorded in spec.md §F so its silence is deliberate.
- Defect evidence: `.moai/worktrees/t1533/.moai/reports/t1533/codex-review-gate-1.md`
  (card t1533 codex review gate round 1, "[P1] POSIX 경로의 실제 백슬래시를 보존 —
  internal/hook/pre_tool.go:1397" — reproduced twice; external write of `"escaped"`
  observed pre-fix; blocking works with the base interpretation function).

Decisions made at plan phase:

1. The defect is an input-normalization error INSIDE the t1530 walking pattern; the
   walk's structure (probe existing components, rejoin missing tail, physical `..`
   pop, depth-bound fail-closed) is retained. The leader hint (does
   `resolveThroughExistingParent` apply?) evaluated to: it IS the defective code; the
   repair narrows to platform-appropriate segmentation. Concrete mechanism (runtime
   GOOS check vs build tags) left to the repair agent, constrained by
   REQ-HBS-001/REQ-HBS-004.
2. Tier M artifact set (spec/plan/acceptance/progress). SPEC ID
   `SPEC-HOOK-BACKSLASH-SYMLINK-001` — regex check PASS, no collision in
   `.moai/specs/`.
3. Scope: ONLY the pre_tool.go:1397 bypass. The other 12 findings of the t1533 round-1
   review belong to the t1454 residual ledger (out of scope, spec.md §F).

## §E.1 Plan-phase Audit-Ready Signal

_<plan artifacts complete (spec.md / plan.md / acceptance.md / progress.md); awaiting
plan-auditor verdict — auditor to verify the RED-now cell procedure of AC-HBS-001 is
executable as written and that scope boundaries (§F) match the diff when run-phase
lands>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
