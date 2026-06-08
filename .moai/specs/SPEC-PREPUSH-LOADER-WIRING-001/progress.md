# Progress — SPEC-PREPUSH-LOADER-WIRING-001

> Status: draft (plan-phase complete). 4-phase lifecycle scaffold.
> Tier S · `internal/config` · READ-only loader wiring for the `git_strategy` section.

## Phase Tracker

| Phase | Status | Commit SHA | Notes |
|-------|--------|-----------|-------|
| Plan  | complete | _(pending)_ | spec.md + plan.md + acceptance.md + progress.md authored |
| Run   | pending  | _(pending)_ | M1 wrapper → M2 loader+wired call → M3 tests → M4 full-suite reconcile |
| Sync  | pending  | _(pending)_ | CHANGELOG + status → implemented |
| Mx    | pending  | _(pending)_ | status → completed + 4-phase close |

## Milestone Status (run-phase)

- [ ] M1 — `gitStrategyFileWrapper` struct in `types.go` (mirror `gitConventionFileWrapper`,
      wrap local-package `GitStrategyConfig`)
- [ ] M2 — `loadGitStrategySection` method in `loader.go` + wired call in `Load()`
- [ ] M3 — unit tests covering AC-PLW-001..008 + 2 edge cases (`t.TempDir()`, table-driven)
- [ ] M4 — full `go test ./internal/config/...` reconcile per plan.md §E.1

## §E.2 Run-phase Audit-Ready Signal

_(populated by manager-develop at run-phase completion)_

- run_commit_sha: _(pending)_

## §E.3 Run-phase Evidence

_(populated by manager-develop — AC verification output, coverage delta)_

## §E.4 Sync-phase Audit-Ready Signal

_(populated by manager-docs at sync-phase completion)_

- sync_commit_sha: _(pending)_

## §E.5 Mx-phase Audit-Ready Signal

_(populated at Mx-phase close)_

- mx_commit_sha: _(pending)_

## Notes / Open Questions

- Scope is FIXED READ-only (user-approved this session via AskUserQuestion). No open questions.
- WRITE/Save path deferred (see spec.md §C Exclusions) — candidate for a future SPEC if a
  production caller that mutates `cfg.GitStrategy` and saves it ever materializes.
- Regression to anticipate: existing loader fixtures that include `git-strategy.yaml` now
  load real values (plan.md §E.1) — expected, reconcile not revert.
