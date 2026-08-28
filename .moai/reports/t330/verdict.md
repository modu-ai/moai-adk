# Card t330 — verdict

**SPEC**: SPEC-TODO-DESTRUCTIVE-GUARD-001 — reversibility for `moai todo done`
**Branch**: `WT-todo-destructive-guard` · **Worktree**: `.claude/worktrees/t330`
**Card**: t330 · **Tier**: M · **Lane**: lane-5 · **Date**: 2026-08-28

## Claim

The card is complete through sync-phase and ready for the lead's integration window. It is **not** integrated: nothing was pushed, no PR exists, and no CI run covers this branch.

## Evidence

This lane inherited a worktree whose run-phase had already closed (9 commits, 16/16 AC). Everything below was measured in **this** session, on **this** tree.

### Base absorption

```
$ git fetch origin develop
From https://github.com/modu-ai/moai-adk
 * branch                develop    -> FETCH_HEAD

$ git rev-list --count --left-right origin/develop...HEAD
38	9

$ git merge origin/develop
(merge commit c7445353f — no conflict; the 9 pre-existing commits preserved, no reset, no force)
```

### Re-verification on the merged tree

```
$ go build ./...                                  -> rc=0
$ go vet ./...                                    -> rc=0
$ go test ./internal/kanban/ -count=1 -cover
ok  github.com/modu-ai/moai-adk/internal/kanban  13.256s  coverage: 87.0% of statements
$ go test ./internal/cli/ -count=1 -cover
ok  github.com/modu-ai/moai-adk/internal/cli    264.156s  coverage: 79.9% of statements

$ diff -q .claude/skills/moai/workflows/todo.md \
          internal/template/templates/.claude/skills/moai/workflows/todo.md
-> rc=0 (byte-identical mirror)
```

The full suite was NOT run locally, per CLAUDE.local.md §4 — the whole-module verdict is CI's, on a clean environment against the integrated head.

### Commits added by this lane

| SHA | What |
|---|---|
| `c7445353f` | merge — absorb `origin/develop` |
| `0048e33cd` | sync-phase close: CHANGELOG entry, docs-site 4-locale, `progress.md` §E.4, `spec.md` frontmatter `in-progress → completed` |
| `0566d5394` | backfill `sync_commit_sha` (a commit cannot name its own SHA) |
| `7382ce247` | correct the CHANGELOG close sentence to the transitions that actually happened |
| `<this commit>` | audit report + verdict + F1-F3 corrections |

### Independent sync audit

**PASS-WITH-DEBT, harmonic 0.890** — Functionality 0.95 / Security 0.92 / Craft 0.82 / Consistency 0.88. Zero blocking findings, ten debt findings. Full report: `.moai/reports/t330/sync-audit.md`.

The auditor built the pre-archive binary from `812ee01fc` and drove six isolated fixtures rather than reading the run-phase claims back. It confirmed the strict round trip is byte-identical, that a pre-archive binary opens and writes a database carrying archived rows without dropping them, that the id allocator refuses a restore into a reissued id even with `last_seq` hand-lowered to 0, and that both freeze tests were extended rather than loosened. It also moved the JSON downgrade loss from inference to measurement.

## Baseline-attribution

Every figure above was produced by the quoted command, in this session, against `.claude/worktrees/t330` on branch `WT-todo-destructive-guard`, at the HEAD each command names. The run-phase figures in `progress.md` §E.3 are cited there as run-phase measurements and are not restated here as fresh ones — except `internal/cli` coverage, which this session re-measured on the merged tree (79.9%, vs 79.8% at run-phase).

## Corrections this lane made

Two came from the audit, one from verifying the sync agent's own report:

1. **F1** — `progress.md` and `plan.md` §D both carried a commit SHA (`3cb258d62`) that `spec.md` v0.2.2 had already corrected. Re-derived independently: `git log 812ee01fc --grep='t306' --oneline | wc -l` → 13, earliest `3030df58b` — the *plan-phase* commit, before any of that card's code existed. Both stale copies fixed.
2. **F2 / F3** — the CHANGELOG carried an unattributed coverage figure and an uncovered-arm miscount ("two arms" where §E.3 records four). Both corrected and attributed.
3. **The sync agent's report claimed** it closed "spec.md / plan.md / acceptance.md frontmatter" with `updated` refreshed. Measured: only `spec.md` carries a frontmatter block, and `updated` was left unchanged (already same-day). The CHANGELOG sentence was rewritten to the transitions that actually happened.

## Gaps — what was NOT observed

- **Nothing pushed; no CI run exists for this branch.** The repository-wide verdict belongs to whoever integrates it.
- **Full test suite not run locally** (see above). Only `internal/kanban` and `internal/cli` were exercised.
- **Windows tests not compiled or run.** Only `GOOS=windows go build` was verified at run-phase; a green cross-build does not establish the tests compile there.
- **No `-race` run.** The archive rides the existing `Mutate` lock and adds no goroutine — an argument, not a measurement.
- **`moai spec lint` / `moai spec audit` not run** after the frontmatter transition.
- **docs-site `hugo build` not invoked.** The 4-locale checks were grep-based: line-count parity (227 each), heading-count parity (9 each), URL-blacklist clean.
- **`internal/cli` coverage (79.9%) is below the 85% package target** and was not measured at a base tree, so no claim is made about whether this change moved it.

## Residual risk

- **`--require-landed` currently has no practical discriminating power in this repository** (audit F6). Shipped opt-in and documented in the help text and both `todo.md` copies, but its fail-open is indistinguishable from a pass on stdout (`rc=0`, `done t2`), and with the develop-integrated ref the predicate is structurally always-true. Turning the flag on does **not** prevent a t306-class incident; closing that needs t331's persisted landing-state field.
- **The archive is rewritten in full on every queue mutation** (audit F4, `backlog_migrate.go:292`), not only on `done`. Unbounded growth is declared out of scope in `spec.md` §D, but that the growth is a *per-write* cost is recorded nowhere in the SPEC.
- **Restore position is exact only for the immediate round trip**; a later restore is clamped into range.
- **`export-json` now writes an `archived` key** — additive, but the artifact's shape changed.
- **The docs-site prose was derived** from the doctrine and `progress.md`, not from a fresh line-by-line read of the Go implementation at sync-phase.

## Verdict

**Ready for the integration window.** Sync-phase closed, independent audit PASS-WITH-DEBT 0.890 with zero blocking findings, all three cheap findings corrected before integration. Integration into local `develop` awaits the lead's reply — this lane starts no merge on its own.
