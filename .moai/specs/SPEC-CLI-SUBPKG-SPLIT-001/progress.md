# Progress — SPEC-CLI-SUBPKG-SPLIT-001

> Canonical §E lifecycle skeleton. Plan-phase emits placeholder headings only; §E.2/§E.3
> are populated by manager-develop (run-phase) and §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex self-check: `decomposition: SPEC ✓ | CLI ✓ | SUBPKG ✓ | SPLIT ✓ | 001 ✓ → PASS` (canonical `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`; last segment digit-only, no alpha suffix).
- Frontmatter: 12 canonical fields present; `status: draft`; `priority: P2`; ISO `created`/`updated`; `tags` comma-separated string; `tier: L`; `era: V3R6`. Validated.
- Artifacts emitted: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md (Tier L 5-file set + progress).
- Tier: **L** (25,838 LOC flat package, 93 files, 54,756 test LOC, cross-package structural refactor with import-cycle + deps coupling → constitutional-scale).
- Out of Scope: present with `### Out of Scope —` H3 sub-headings (big-bang split / tiny single-file clusters / deps-platform-tangled clusters / functional change / new tests / existing-subpackage reorg / CHANGELOG-README).
- Measurements grounded (NOT assumed): all §A research metrics from `find`/`wc`/`grep`/`go build` against the working tree; build baseline `go build ./internal/cli/...` observed exit 0.
- Cluster map: 26 candidate clusters + shared kernel, reconciled to ~25,730 LOC (residual ~108 = `worktree_validation.go`, belongs to existing worktree subpackage). Exact per-cluster counts re-derived at run-phase, not hardcoded.
- Two dominant risks documented: (1) test-migration surface (147 files / 54,756 test LOC white-box) — dominant regression risk; (2) import-cycle hazard (root imports subpackages for AddCommand → subpackages cannot import package cli back) — blocks kernel-dependent clusters on a prior `uikit` leaf extraction.
- deps coupling quantified: exactly 8 root files touch the global `deps`; provider-injection pattern (worktree.WorktreeProvider precedent) is the resolution.
- Honest value/risk framing recorded in spec.md §A: refactor of WORKING code, gains real-but-incremental + non-user-observable; big-bang REJECTED; phased lowest-risk-first RECOMMENDED with post-M4 checkpoint stop-condition (REQ-CSS-010).
- Status: plan-phase artifacts final (status: draft); ready for Implementation Kickoff Approval human gate.
- **Drift re-verification (2026-07-07, pre-audit)**: re-measured against the current working tree. Δ +5 root non-test files / +602 LOC, +6 test files / +1,265 test LOC, +2 build-tag files; deps-coupled count unchanged at 8; build baseline still exit 0 (both `go build ./...` and `GOOS=windows GOARCH=amd64`). All M1-M7 cluster files remain cohesive; 6 existing subpackages intact; kernel helpers still in `package cli`. Cluster LOC measured (M1 migrate 947 / M2 profile 1,527 / M3 agentlint 1,392 / M6 doctor 2,073 / M7 update 5,184) — within ±15% of plan.md claims; plan still valid. `moai spec lint` PASS (0 findings); `moai spec audit` PASS (0 drift findings, 1 modern-era clean). Plan-audit-ready confirmed.

## §E.2 Run-phase Evidence

_<pending run-phase — populated by manager-develop (cycle_type=ddd)>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs; carries sync_commit_sha>_
