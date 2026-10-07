---
id: SPEC-CUTOVER-RESIDUE-001
title: "Acceptance criteria — cutover residue cleanup"
version: "0.1.0"
created: 2026-10-07
updated: 2026-10-07
author: t1564 card worker
priority: P1
module: "internal/cli/worktree"
tier: M
---

# SPEC-CUTOVER-RESIDUE-001 — Acceptance

Measurement tree at plan time: `cb2a011d0` (branch `WT-cutover-residue` cut from `origin/main`). Evidence ledger lives in `.moai/reports/t1564/verdict.md`; each criterion below cites its ledger entry id.

| AC | Requirement | Assertion (green path) | RED-now (pre-implementation tree `cb2a011d0`) | M | Status |
|----|-------------|------------------------|------------------------------------------------|---|--------|
| AC-CR-001 | REQ-CR-002/003 | `go test ./internal/cli/worktree/ -run '^TestSweepEffectiveBase$' -count=1` exits 0 with all four cases (fallback / present / probe-error / head-error) green | NEW test file — fails to compile (sweepEffectiveBase undefined): the gap is real, the test is the instrument | M1 | open |
| AC-CR-002 | REQ-CR-001 | `git ls-remote --exit-code origin refs/heads/develop` exits 2 and `git ls-remote --symref origin HEAD` names `refs/heads/main` on this repo's remote | Same command — observation, already true at plan time (measured 2026-10-07, exit 2 / `ref: refs/heads/main`) — regression-guard against the remote changing back | M1 | observed |
| AC-CR-003 | REQ-CR-003 | Post-fix binary: `./bin/moai worktree sweep --json` reports 0 verdicts with `cause=fetch-failed`; fallback notice on stderr names both bases | Pre-fix binary on the same tree: `--json | grep -c 'cause=fetch-failed'` > 0 (187 at tick 3; re-measured this run at RED step) | M1 | open |
| AC-CR-004 | REQ-CR-004 | Explicit `--base origin/develop` on the post-fix binary still reports `cause=fetch-failed` (no silent switch) | No unit seam test contradicts it; verified live at M4 | M1 | open |
| AC-CR-005 | REQ-CR-005 (revised) | The config's `lead_push_threshold` counting comment names no branch (`grep -c 'origin/<develop>'` → 0, `grep -c 'develop'` inside the comment block → 0); `worktree_base_branch: main` present; `manual.workflow` still `git-flow` (guard-arming coupling, spec §A amendment); `go test ./internal/config/... -count=1` green and `TestGitHubFlowSweepGuard` green (disarmed) | Pre-state: comment named `origin/develop..develop`; the flip attempt was measured RED (`--- FAIL: TestGitHubFlowSweepGuard ` on the flipped tree, run-phase M2) | M2 | open |
| AC-CR-006 | REQ-CR-006/007 | Survival doc §2.3 names `git restore --source=main` + the two verification greps; template comment carries no branch name; `go test ./internal/template/... -count=1` green and `make build` succeeds | Pre-state: doc instructs `--source=develop`, `worktree_base_branch: develop`; template comment names `origin/<develop>..<develop>` | M2 | open |
| AC-CR-007 | REQ-CR-008 | `grep -rn 'branches:.*develop' .github/workflows/` → 0 matches; docs-i18n-check.yml has no develop trigger entry | Pre-state: 8 files × 10 filter lines match (measured `grep -n` at plan time) | M3 | open |
| AC-CR-008 | all | `golangci-lint run ./internal/cli/worktree/... ./internal/config/... ./internal/template/...` clean; `go vet` clean; affected-package suites green with `-timeout 30m` | Baseline lint at plan time: to be recorded at M4 (pre-existing findings reported separately from NEW) | M4 | open |

## Two-cell discipline notes

- AC-CR-001's RED-now is the compile-failure form (the behavior does not exist yet); the unit RED output at M1 step 2 is recorded verbatim in the ledger before GREEN.
- AC-CR-002 is an observation criterion about the remote, not the work: classified regression-guard (its red/green depends on the remote, not on this change).
- AC-CR-003's pre/post pair is the audit-verification corollary (`.claude/rules/moai/development/verification-completeness.md` §5): both forms run on the same tree, differing only in the binary.
