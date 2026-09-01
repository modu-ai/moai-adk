# Sync Close — SPEC-STATUS-TRANSITION-VALIDITY-001 (card t376)

Date: 2026-09-01. Tree: `.claude/worktrees/t376`, branch `WT-status-transition-gap`,
base HEAD at close start: `b43d4cb56`.

## Claim

The 3-phase close (plan→run→sync) of SPEC-STATUS-TRANSITION-VALIDITY-001 is complete:
§E.3 `run_commit_sha` backfilled, §E.4 written citing the already-run sync-audit
(PASS 91.2/100), `spec.md` frontmatter `status: in-progress → completed` on the single
sync commit, `updated: 2026-09-01`. No CHANGELOG / docs-site / README entry, per this
branch's zero-doc-file precedent and the audit's own disposition (all 4 findings
optional and closed).

## Evidence

Commands run in this close, this tree, in this order:

```
$ git rev-parse --short HEAD        → b43d4cb56
$ git branch --show-current         → WT-status-transition-gap
$ sed -n '3,8p' .../spec.md          → status: completed / updated: 2026-09-01 (after edit)
$ git status --porcelain | grep -v '^??'
  → M .moai/specs/SPEC-STATUS-TRANSITION-VALIDITY-001/progress.md
    M .moai/specs/SPEC-STATUS-TRANSITION-VALIDITY-001/spec.md
$ git show --stat --oneline b1bcce4f4  → F1/F3/F4 fixes (gitquery_cache.go, lint_ownership.go, lint_transition.go, 2 tests, progress.md)
$ git show --stat --oneline ab114b2cf  → F2 fix (spec.md §C)
$ grep -n '91.2\|Overall verdict' .moai/reports/t376/sync-audit.md → PASS, weighted 91.2/100, 0 blocking
```

Commit plan (this close): commit 1 = SPEC artifacts + this report
(`docs(SPEC-...-001): sync-phase artifacts — 3-phase close (t376)`); commit 2 = backfill
of `sync_commit_sha` in §E.4 to commit 1's SHA.

## Baseline-attribution

All claims above were measured in this run against the tree at `b43d4cb56` and its
committed history (`git show` reads of `b1bcce4f4`, `ab114b2cf`, `b43d4cb56`).
The carried sync-audit figures (PASS 91.2, dimension scores, findings F1-F4) are
attributed to the sync-auditor's run recorded in `.moai/reports/t376/sync-audit.md`
(audited HEAD `ff8a7dcba`, binary provenance `strings bin/moai | grep -c ff8a7dcba → 4`)
— cited, not re-measured here.

## Gaps

- No tests, lint, vet, or build were executed in this close. §E.2/§E.3 figures are
  citations of the run-phase and sync-audit runs, not re-measurements.
- No CI verdict exists for the new commits — nothing was pushed.
- No CHANGELOG/docs-site/README sweep was performed (deemed not applicable per branch
  precedent and dispatch constraint; not independently re-verified by grep here).
- The 4 pre-existing untracked files under `.moai/reports/t376/` (lint-after-m1.err,
  lint-after-spec.json, lint-baseline.err, lint-sync-audit.err) were left untouched
  and untracked.

## Residual-risk

- `run_commit_sha: 73bfba170` is derived from §E.2's own provenance statement (the
  measurement binary was built from that commit); the F1 fix later modified
  `lint_transition.go`, so `git diff --stat 73bfba170 -- internal/spec/` is non-empty
  at close. The corpus totals were re-measured after F1 with findings byte-identical
  (1200), and the sync-audit independently re-measured at `ff8a7dcba`, so the
  attribution chain is double-covered — but a reader wanting a single-tree proof must
  use `ff8a7dcba` (the audit's audited HEAD), not `73bfba170`.
- If the lead's merge window lands after further commits on this branch, the
  `sync_commit_sha` backfilled here still names the correct sync commit, but the
  branch HEAD will have moved past it.
