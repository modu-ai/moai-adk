# progress.md — SPEC-GIT-STATUS-FIXTURE-001

Card: t474 · Branch: `WT-git-status-fixture` · Base: `25a3212a9` (origin/develop tip)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-GIT-STATUS-FIXTURE-001
card: t474
branch: WT-git-status-fixture
base: 25a3212a9
status: draft
verified_facts:
  - "Two internal/core/git tests (TestStatusAheadBehindFromHeader :269, TestStatusBranchHeaderShapes :122/:128) are deterministically red on develop CI (dual-job: Test (ubuntu-latest) + Race Test) and green locally."
  - "Single root cause, measured: unpinned 'git init -q --bare' at status_branch_test.go:99 and :235 leaves the bare remote HEAD at ambient init.defaultBranch (main on dev machines via Xcode system gitconfig, master on CI) → unborn-HEAD clones → wrong-branch push exiting 0 → all three CI signatures reproduced with byte-identical fatal signature."
  - "Identity hypothesis eliminated: gitFixture injects GIT_AUTHOR_*/GIT_COMMITTER_* env (status_branch_test.go:22-27)."
  - "Repair precedent in-repo: helpers_test.go:42 'init --bare -b main'."
  - "Local RED not achievable in lane (worktree-scoped config forcing measured ineffective; system/global/env forms rejected for lane safety) — RED-now evidence is the L4/L5 mechanism probe + live CI red; the deciding fix verdict is the merged-tree CI run (external dependency)."
evidence: .moai/reports/t474/repro-probe.md
plan_status: audit-ready
plan_complete_at: "2026-09-03T19:04:13Z"
```

Artifacts: spec.md (4 REQ, 6 AC — Tier S inline), plan.md (M1 kickoff gate on optional :315 pin → M2 repair+local sweep → M3 external CI confirmation).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
