# t584 — Card Verdict: SPEC-AUT-PERMMODES-001

> Card **t584** · branch `WT-autonomy-perm-modes` · judged tree `d34ba9ec4` (clean) · 2026-09-14.
> Evidence-bearing completion report per `verification-claim-integrity.md` §3. Final PASS/FAIL judgment is the lead's, read from this file and `.moai/reports/t584/sync-audit.md`.

## Claim

- SPEC-AUT-PERMMODES-001 delivered and closed: init wizard autonomy question redefined as Claude Code's real permission modes — 3 options (`Accept edits on (Recommended)` / `Auto mode` / `Bypass permissions`), `acceptEdits` written as USER-scope `defaultMode` default, REQ-007 re-scoped to zero-delta (bounded-delta invariant), REQ-006 downgrade regression set (5 named tests) preserved green; unknown tier → `"default"` MOST-restrictive fail-safe.
- Lifecycle: `status: completed`, `sync_commit_sha` backfilled, §E.4 signal authored, CHANGELOG entry emitted.
- Run-phase AC: 12/12 PASS. Independent sync-audit: **PASS 8.8/10 — 0 blocking findings** (0 S1, 0 S2, 5 S3 non-gating).
- Disposition: ready for develop merge — integration window requested (queue order t747 → t741 → t584).

## Evidence

Per-phase commits on this branch (all local, unpushed — lanes never push):

| Phase | Commit(s) | What |
|---|---|---|
| plan | `0ddd1a282` (v0.1.0) → iter2 v0.1.1 | plan-audit iter1 FAIL 0.75 → D1-D7 applied (`.moai/reports/t584/plan-audit{,-iter2}.md`) |
| run | `e8deabc21` M1 / `1c30ff881` M2 / `92f2e4c7f` M3 | version-floor evidence; TierDefaultMode mapping + labels + tests; downgrade-guard re-scope + §E.3 signal |
| absorb | `54ea53fb1` | develop merge absorbed pre-audit; tree clean |
| sync | `07eed817c` + `d34ba9ec4` | §E.4 + status completed + CHANGELOG; sync_commit_sha backfill |

Key measured outputs cited in `.moai/reports/t584/sync-audit.md` (17 rows E1-E17), re-measured by the independent auditor: affected packages green (wizard 6.167s / config 9.671s / core/project 7.082s); AC-008 5-named regression set 5/5 PASS; `TestTierDefaultMode` mapping + unknown→`"default"` fail-safe PASS; `./bin/moai init --help` prints the AC-011 string verbatim; AC-012 godocs state the new mapping and bounded-delta invariant; coverage re-measure wizard 93.6% / config 82.0% / core/project 88.8% (matches §E.3); `golangci-lint` changed packages **0 issues**; `GOOS=windows` build exit 0; kill-switch-trumps-proof gate verified in source + tests; official docs live cross-check (E15); M3 comment-only verified by diff (E16).

## Baseline-attribution

- Audit figures: re-measured by sync-auditor **this run, this tree @ `d34ba9ec4`** (its own commands, quoted verbatim in sync-audit.md).
- Full `internal/cli` suite figure: run-phase export `.moai/state/verify/t584/t584-green-cli-full.txt` (launched against the `1c30ff881` tree; the later M3 delta was comment-only — verified E16 — and the config package re-verified green post-M3).
- §E.3 build/cross-build evidence: run-phase, per `progress.md` §E.3.

## Gaps

- config package coverage 82.0% < 85% target; baseline unmeasured → pre-existing vs new attribution open (touched functions 100%; audit F2).
- Interactive TTY wizard not exercised end-to-end (golden + units + source reads cover it).
- Full `internal/cli` suite not re-run at audit (load discipline; run-phase export stands).
- docs-site `getting-started/init-wizard` 4-locale pages stale vs new labels — not read this phase; declared separate-card disposition (audit F4).
- Two wiring-test bodies + darwin full build attested by package-green runs, not individually quoted.

## Residual-risk / non-gating follow-ups

- **F1** — IAM reference `:43-44` says both kill switches are settable in "any settings file"; live official-docs check shows `disableBypassPermissionsMode` is **managed settings** only. One-file doc fix, owes `make build` (template rule). Recommended fast follow-up.
- **F2** — coverage baseline measurement card for the config package.
- **F4** — docs-site init-wizard refresh card (4 locales).
- **F3** — `sync_commit_sha` recorded as short SHA (mechanical predicate passes; full SHA documented in §E.4 comment).
- CHANGELOG entry is long by house convention; lead may trim at release.
