# Progress — SPEC-CODEX-E2E-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

plan_phase:
  spec: SPEC-CODEX-E2E-GUARD-001
  base_sha: "ace1c5440"
  branch: WT-codex-e2e-guard
  worktree: .claude/worktrees/t500
  tier: M
  measurements: spec.md §A + §F (commands and outputs in spec.md HISTORY)
  premise_refutations:
    - axis-2 card premise REFUTED (spec.md §F.1) — existing TestStatusLineDefaultSubsetOfAllowlist
    - lane "41 matches no population" PARTIALLY SUPERSEDED — 41 is the repo-wide codex-named
      test-file population at this base (spec.md §F.2)
  baseline_greens:
    - "go test ./internal/codexwiring/ -run TestStatusLine -count=1 → ok 0.617s"
    - "go test ./internal/cli/ -run 'TestCodexSpecFiles|TestCodexCommand_NeutralityScan|TestCodexSpawn_TmuxDiagnosticSingleSource' -count=1 -timeout 600s → ok 0.929s"
  lint: pending run-phase (moai spec lint at sync)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
