# SPEC-UPDATE-CI-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- Baseline tree: `d5336214e` on `plan/epic-update-config-audit` (merged with `origin/main`).
  Authored at worktree `9426bf49b`, whose merge-base with `origin/main` was `76d9a8f3b`; the local
  branch `main` at `1d4e4f7da` was a stale divergent branch, not this branch's base.
- Findings F1-F5 each re-verified against this tree while authoring. Four drifts recorded
  (spec.md §A.6): the `ci.yml:222` job attribution (test-integration, not build); the `internal/cli`
  75.6% / 75.7% two-tree coverage split — **RESOLVED** by the `origin/main` merge, which collapsed
  the two trees into one baseline measuring 75.7%; the neutrality workflow running three isolated
  test targets rather than one; and two independent `C<N>` class-numbering schemes across the two
  guard files.
- F3 re-measured at `d5336214e`: `go test -cover ./internal/cli/... ./internal/config/...` —
  `internal/cli` 75.7%, `internal/config` 80.5%, `internal/cli/specid` 58.3%, six update
  subpackages 88.6-97.5%. No `codecov.yml` / `.codecov.yml` at repo root.
- F4 independently re-measured: `go tool cover -func` over `internal/cli/update/backup/` — all five
  `merge.go` functions at 100.0%, package total 88.6%.
- F5 pattern set verified against implementation: `internal_content_leak_test.go:171` C1 =
  `SPEC-(V3R[2-6]|AGENCY|WORKTREE)-`, `:282` C1c = `SPEC-(DB-SYNC-RELOC|PROJECT-DB-HINT)-[0-9]{3}`;
  `template_neutrality_audit_test.go:168` C6 = `PR #[0-9]+`. All three §A.5 shapes escape.
- SPEC ID regex self-check executed as Bash: `PASS`.
- Requirements: 18 REQ + 4 NFR across §B.1-§B.6. Acceptance criteria: 23 (14 behavioural,
  9 presence-only, every presence criterion paired). Falsification procedures: 5.
- Status: `draft`. Awaiting plan-audit and Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
