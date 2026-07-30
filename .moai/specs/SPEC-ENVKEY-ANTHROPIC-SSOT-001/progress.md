---
spec: SPEC-ENVKEY-ANTHROPIC-SSOT-001
phase: run
---

# Progress - SPEC-ENVKEY-ANTHROPIC-SSOT-001

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored at base commit `76d9a8f3b` on branch
`plan/SPEC-ENVKEY-ANTHROPIC-SSOT-001`.

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- SPEC ID regex pre-write check: executed as Bash, observed `PASS`
- Duplicate SPEC ID check: observed `NO_DUPLICATE_SPEC_ID`
- `moai spec lint` on `spec.md`: observed `✓ No findings — all SPEC documents are valid`
- Baseline measurements re-verified in this worktree (not carried over):
  83 bare literals / 10 files / 9 unique values / 6 existing constants /
  46 reachable + 37 unreachable by the existing guard / 291 `_test.go` refs /
  12 template refs / `go vet ./...` exit 0
- 15 acceptance criteria authored, each carrying either an executed verdict
  command with its observed pre-change baseline, or an explicit
  `not runnable at base` marking with the reason
- 4 open questions (OQ-1..OQ-4) recorded with recommendations, pending
  resolution at the Implementation Kickoff Approval gate

### Audit iteration 2 (plan-auditor FAIL 0.72 -> remediation)

An independent plan-auditor returned FAIL (0.72) with 5 MUST-FIX and 6 SHOULD-FIX
findings. Remediation landed in `spec.md` v0.2.0, `plan.md`, and `acceptance.md`.
Every remediating code fact was re-verified by executed command in this worktree:

- `lifecycle` corrected `spec-first` -> `spec-anchored` (canonical enum,
  `.claude/rules/moai/development/spec-frontmatter-schema.md:48`)
- AC-EAS-006 rebuilt as a **runtime** banned-set assertion; the prior source-grep
  baseline was fabricated (executed: exit 0 with 6 lines, not exit 2) and
  contradicted plan.md's derive-from-constants design
- AC-EAS-013 lint verdict rebuilt to capture the exit code before the pipe;
  vacuity proven by executed `bash -c 'false | tail -5; echo $?'` -> `0`
- M6 falsification retargeted `internal/hook/session_start.go` ->
  `internal/sandbox/env.go:32` (the sole production `ANTHROPIC_API_KEY` site,
  confirmed by repo-wide grep)
- Guard root-finder retargeted to `findProjectRoot(t)` at
  `internal/config/required_checks_test.go:162`; the
  `findRepoRoot`-redeclaration hazard against
  `internal/config/token_budget_guard.go:45` recorded in plan.md F/M2
- AC-EAS-007 gained a positive scope assertion (sibling-file plant); AC-EAS-012
  gained a `cmd/` plant, a 9-name falsification table, and residual-diff checks
- Missing baselines executed and recorded: per-package `^FAIL` counts (all `0`),
  `golangci-lint run --timeout=3m` -> `0 issues.` / `lint_exit=0`, template
  changed-file count `0`
- Census re-measurement gate added as a [HARD] run-phase entry precondition
  (spec.md A.5, plan.md C.1)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
