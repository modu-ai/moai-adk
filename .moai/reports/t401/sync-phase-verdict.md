# t401 sync-phase verdict (continuation session, 2026-09-08)

card: t401 · SPEC-JUDGMENT-FIRST-MODE-001 v0.2.4 · Tier L · gitflow lane (develop integration via lead's window)
worktree `.claude/worktrees/t401` · branch `WT-analysis-pull`

**The close, in one sentence:** **21/23 acceptance criteria PASS** (corrected from an earlier
22/23 at the sync-audit — AC-JFM-023 had been carried as a pass without its required export and
calls_issued contrast); **AC-JFM-018 (vacuity falsifier) and AC-JFM-023 (positive control) are
both RED by measurement and stay RED at close** — the falsifier's pull-mode denominator is empty
(0 rows) and its ≥20-row collection window opened at close, and the positive control's required
baseline export + session-split `calls_issued` contrast was not performed; **both are owned by
follow-up card t547 (issued with operator approval, entry-gated on the develop merge that turns
the primary checkout's config to `pull`)**, and no release relying on the falsifier evidence may
precede t547's close. The release gate (t204) reads this path: verdict Gaps → t547.

## Claim

The SPEC delivers `interview.recommendation_mode` (default `push`, template default `push`,
this repo dogfooding `pull`) — a judgment-first axis from issue #1683's "Analysis is Pull, Not
Push" — with the mode branch across six surfaces, a 6-test config contract, a six-check CI
consistency guard demonstrated in both directions, a closed sweep ledger, and a fully backfilled
progress record. One criterion (AC-JFM-018) is carried RED into close on the lead's
operator-backed (a) path, explicitly marked, never passed.

## Evidence (deciding commands + observed output, this run, this tree)

| Dimension | Command (verbatim) | Observed |
|---|---|---|
| Config contract (AC-JFM-002/003) | `go test ./internal/config/... -run 'RecommendationMode' -count=1` | 3/3 packages ok; 7 tests incl. `TestRecommendationModeUnrecognizedFallsBackToPush` PASS; `[no tests to run]` absent from the config package line |
| Affected packages (AC-JFM-022) | `go test ./internal/config/... ./internal/hook/... ./internal/template/... -count=1` | every line `ok` (hook 98.1s, mx 30.4s, perf 84.1s, quality 34.6s, security 17.2s) |
| Observer (AC-JFM-014/015/016) | `go test ./internal/hook/ -run 'AskUserQuestionObserver' -count=1 -v` | 7/7 PASS (RowShape, NeverDenies, FailOpen, ModeResolution, …) |
| Neutrality (AC-JFM-020) | `grep -rln 'SPEC-JUDGMENT-FIRST-MODE\|REQ-JFM\|/Users/\|CLAUDE.local' internal/template/templates/` | 0 matches (rc=1); `TestTemplateNeutralityAudit` ok |
| Mirror parity (AC-JFM-019) | per-pair `diff -q` across the 13 swept paths | 10 pairs byte-identical; `rules/local/` ×2 no-mirror by documented intent; `settings.json`'s twin is `.json.tmpl` (+11 both sides) |
| Hook wrapper pairs (AC-JFM-021) | corrected 4-pair drift loop (swept count printed) | `swept=4 drift=0` — agent-hook, stop-goal, task-completed, teammate-idle all SYNC |
| CI guard (AC-JFM-017) | run block extracted verbatim, executed locally | clean tree: `all checks pass` exit 0; scratch mutation `86ddb89d0` (one `interview.recommendation_mode` ref dropped): `references dropped below 5 (got 4)` exit 1; scratch removed before merge; clean-tree re-run exit 0 |
| Token budget | `go test ./internal/config/ -run TestAlwaysLoadedTokenBudget -count=1 -v` | surface 78,211 ≤ raised budget 78,500 (headroom 289) |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| M6 falsifier (AC-JFM-018) | `jq -s '[.[] | select(.mode=="push" and .label_present==true)] | length'` over the primary observer log | baseline 20 rows (push, labeled — AC-JFM-023 green); **pull rows: 0** → denominator empty → **RED, window open** |

## Baseline-attribution

Every figure above was measured in this session, in this worktree, against the branch tip
`c82c47f66` (sync-artifact commits land after it: this verdict, the CHANGELOG entry, the
frontmatter transition, and the sync commit). The RED/GREEN guard demonstration used a scratch
commit on this branch only; it was removed (`reset --soft` + reversal edit — the `reset --hard`
form was denied by the permission guard and not retried) before any merge surface.

## Gaps — what was explicitly NOT observed

1. **AC-JFM-018 has no falsifier sample.** 0 pull-mode rows exist; nothing was exported because
   exporting now would manufacture the empty-set `violations: 0` reading plan.md §F M6 warns
   against. **Owner: follow-up card t547** (queued, operator-approved; entry-gated on the develop
   merge — the observer follows the asking session's own tree's config, so rows stay `push` until
   the merge lands `pull` at the primary). The export + REQ-JFM-025 provenance (`rows_recorded`
   vs `calls_issued`) is t547's first act on entry.
2. **AC-JFM-023 was originally carried as a pass — that was wrong, and the sync-audit (F1)
   caught it.** The criterion requires the M0 baseline sample exported to
   `baseline-push-window.jsonl` with its provenance and a `calls_issued` four-way contrast; none
   of that was performed. The 20-row primary log is a mixed-session sample (8+ session_ids) —
   the criterion's own rule says a positive control read from an unattributed sample asserts
   nothing, and the rows must be split by session_id and contrasted before any reading. The
   count in every surface was corrected 22/23 → 21/23, and the baseline export + session-split
   contrast joined t547's scope.
2. **The re-measurement at the integration window.** Behind = 1,058+ vs `origin/develop` at
   dispatch; every doctrine coordinate this SPEC cites must be re-taken when the lead's window
   merges this branch into `develop` (the lead absorbs `origin/develop` at that time). The
   M1–M2 re-measurement habit is the standing guard against merge-tree drift.
3. **Remote CI has not seen this branch.** Lane-local verification only; `origin/develop`'s run
   after the lead's batch push is the integration verdict surface.

## Residual-risk

- **AC-JFM-016's strength depends on config plumbing that M3 added late.** The observer's mode
  resolution is unit-tested (`TestAskUserQuestionObserverModeResolution` PASS), but the
  end-to-end path (loader → hook process → row) has no integration test; a wiring defect there
  would be silent until the falsifier window produces its first rows — which is itself the
  window's purpose.
- **The budget raise (76,400 → 78,500) is the second raise this quarter.** The standing root fix
  (the large always-loaded rule diet) remains unlanded; a third raise without it would signal
  the diet is overdue, not that raises are cheap.
- **The "any other (S1) row" SPEC-body wording defect is open debt**, recorded in the ledger
  resolution; a future literal reading of AC-JFM-013 by a fresh auditor could trip on it before
  the manager-spec wording repair lands.
- **M2's missing AC matrix was found by continuation luck**, not by the process: had the branch
  been merged without this session's re-verification pass, AC-JFM-007 and AC-JFM-009 would have
  landed unmeasured. The guard against recurrence is recorded practice, not a mechanism.

## Records the lead asked to be visible in the verdict

- **The M2 AC-matrix gap (the session's core finding).** M2 committed with no AC re-verification
  table in progress.md; the re-run found AC-JFM-007 FAIL (implemented as the re-sort prohibition
  REQ-JFM-008 explicitly rejects) and AC-JFM-009 FAIL (behavior present, named citation absent,
  0 references at `1fb802d09`). Both repaired and re-measured. A milestone that skips its matrix
  is not passed; it is unmeasured.
- **The M4 guard's first GREEN run caught its author's `\+` escape bug** (`Abort\+preserve` in
  BRE means "one or more t" and never matches the literal `+`) — the sequence assertion was
  vacuous until fixed to the literal form the criterion's Verify text prescribes. The
  vacuous-guard direction, observed in the wild before the workflow reached CI.
- **The kickoff-pass record states its evidence honestly**: lead confirmation at dispatch + the
  run commits' existence; no independent timestamp of the gate event exists, and the record says
  so rather than manufacturing one.
