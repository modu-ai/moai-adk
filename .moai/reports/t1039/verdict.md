# t1039 — FINAL VERDICT (lane)

**Card**: t1039 — repair the evidence-path contradiction
**SPEC**: `SPEC-EVIDENCE-PATH-EXCEPTION-001`, Tier L, `status: completed`
**Branch**: `WT-evidence-path`, merge commit `382ba9b71`, 13 commits ahead of `origin/develop`
**Base**: `116820f40` → sync `619774c89` → `sync_commit_sha` backfill `fcc4e12f5` → develop absorbed `382ba9b71`

This file is the first artifact tracked through the exception this card opened. Before it landed,
a path of this shape was ignored; the four probes in §2 are what changed that.

---

## 1. Claim

The card is complete and ready for the integration window.

1. Doctrine no longer asserts something the ignore rules make false: the tracked-citation claim is
   narrowed to verdicts, across four repository files, four template mirrors, one emitted `.codex`
   TOML, `catalog.yaml`, and twelve docs-site files in four locales.
2. `.gitignore` carries a narrow exception implementing the operator's reading (B) — `verdict.md`
   only, at one card-directory level — and the exception behaves exactly as decided.
3. The dead 2026-09-01 negation at the old `:109` is retracted, along with the comment that read as
   though it had fixed something.
4. `SPEC-EVIDENCE-CITATION-CANON-001` records the narrowing additively, with zero deletions to its
   requirement text.
5. Plan-phase closed PASS 0.9375 (Tier L threshold 0.85) after four iterations; run-phase closed
   23/23 AC PASS.

---

## 2. Evidence

**The exception's behaviour, measured in the merged tree at `382ba9b71`** —
`git check-ignore --no-index -q <path>`, where `rc=0` means ignored and `rc=1` means the path can
be tracked:

| path | rc | meaning |
|---|---|---|
| `.moai/reports/tZZZ/verdict.md` | **1** | not ignored — the exception fires |
| `.moai/reports/tZZZ/report.md` | 0 | ignored — the width held |
| `.moai/reports/plan-audit/verdict.md` | 0 | ignored — the leak plan-phase predicted does not occur |
| `.moai/reports/tZZZ/sub/verdict.md` | 0 | ignored — depth limited to one level, as decided |

These four are identical **before and after** absorbing 30 commits of `develop`, which is the
measurement that matters: rule order decides the outcome in this file, and the pre-merge result
does not establish the post-merge one. No absorbed commit touched `.gitignore`
(`git log HEAD..develop -- .gitignore` → empty), but the prediction was checked rather than
substituted for the measurement.

**Independent reproduction.** The same four probes were run by the lane at `7939e38b9`, by the lead
in its own session, and by `manager-docs` at `4cfd112fa`. Four observations, three actors, same
results.

**Tree state at `382ba9b71`**: `git status --porcelain` empty; `wc -l .gitignore` → 450
(417 + 36 insertions − 3 deletions); `develop...HEAD` → `0 13`, so develop is fully absorbed.

**Discriminator note.** Every ignore check above uses `--no-index -q`. Two traps make the obvious
form unreliable and both bit this card: `-v` exits 0 on a *negation* match, and without
`--no-index` the command silently skips already-tracked files. Details and the controls:
`lane-measurements.md` §0.

---

## 3. Baseline-attribution

Every figure below names the tree it was taken in, because this card was repeatedly misled by
counts quoted without one.

| figure | value | tree | date |
|---|---|---|---|
| `verdict.md` at one card-directory level | **319** | primary checkout | 2026-09-21 |
| same command | **11** | worktree `t1039` (of 50 files at that depth) | 2026-09-21 |
| excluded: depth 3+ `verdict.md` | 19 (18 archive, 1 live) | primary checkout | 2026-09-21 |
| excluded: audit-verdict family at depth 2 | 155 | primary checkout | 2026-09-21 |
| `.gitignore` | 450 lines | worktree `t1039` | at `382ba9b71` |

The dispatch carried 304. Re-running the *original* command form today returns 319, and the only
structural difference between the two forms — a depth-1 `verdict.md` — is 0, so the gap is real
growth in a tree other sessions write to, not a command artifact. **No acceptance criterion asserts
an exact count against that tree**, verified across all iterations.

Tool provenance: `./bin/moai` was built at `429a7b3d4`; zero Go files separate it from the tree
under measurement (positive control: 5 files in range, all markdown), so the lint verdict
`0 error(s), 31 warning(s)` was produced by this tree's lint code — though not by a same-commit
build.

---

## 4. Gaps

Explicitly not observed, or observed and left open.

1. **The two `[HARD]` export mandates stay false.** `plan-auditor.md:601` and `sync-auditor.md:108`
   require a verdict to be exported; under reading (B) their filenames stay ignored. This is a
   **knowing operator choice**, not an oversight, and card **t1059** owns the doctrine-text repair.
   t1059 has not landed. The exception was not widened to make the mandates true — silently
   repairing a decided trade-off is the failure this card exists to prevent.
2. **Four of this card's own five evidence files remain ignored.** `plan-audit.md`,
   `plan-audit-iter2.md`, `plan-audit-iter3.md`, `plan-audit-iter4.md` and `lane-measurements.md`
   are outside (B)'s width. **This worktree holds their only copies** — disposing of it destroys
   them.
3. **No push, no CI.** Everything here is a local early signal: no clean-environment full suite, no
   darwin/windows matrix. The binding verdict is CI's, after the lead pushes develop.
4. **`internal/spec` tests fail, inherited from develop.** The sole failing assertion is a file
   t1036 removed (`8f87f2359`, an ancestor of this card's base). Attribution: 0 commits from this
   branch touch that directory; positive control on this card's own SPEC directory → 2.
5. **The AC-counter disagreement is unexplained.** The regex counter reads 24 where the criteria
   number 23; the 24th token is `AC-001` inside a verbatim probe string. The Go counter's internals
   were never read, and whether its pattern matches that token is unmeasured. Nothing was edited to
   satisfy either count; the lead carries it as a separate card. Detail and grading:
   `lane-measurements.md` §9.6.
6. **`REQ-EPE-005`'s `CoverageIncomplete` lint warning did not clear** after naming it in two ACs —
   the rule keys on something that edit does not satisfy. Recorded, not chased; it is plan-phase
   baseline. Net lint introduced by this card: 0.
7. **`SPEC-EVIDENCE-CITATION-CANON-001`'s `updated:` is stale** at 2026-08-31 despite its new
   HISTORY entry. A non-transition frontmatter correction, outside the phases that ran here.
8. **`git show main:.gitignore` stays refused** by the worktree guard, so every line number here is
   worktree-attributed only. The refusal is recorded rather than routed around.
9. **The docs-site build was not re-run** in sync-phase; M5's `hugo` exit 0 with zero warnings is
   reused, justified by this run touching no docs-site file.

---

## 5. Residual-risk

- **Rule order is the live hazard in this file.** `.gitignore` carries five clusters over 23 lines,
  and a rule *below* the insertion point decides a path the card's criteria test. A future edit that
  inserts near the reports rules can change the outcome without touching the exception's own lines.
  The in-place measurement in `progress.md` §E.2 is the check to repeat, not the conclusion to cite.
- **The exception is depth-limited by decision.** One live verdict at depth 3+
  (`lead/merge-batch-20260913/verdict.md`) stays ignored. Widening to depth 2 would pull in an
  18-file rescue archive and dissolve the width — which is why the limit exists, and why the
  `.gitignore` comment names the shape it misses.
- **This card's subject recurred inside the card four times.** A fabricated RED-now figure, a
  vacuous check, a carrier broken by markdown escaping, and an inference that measured numbers
  proved a relation they did not. Each was caught by a different mechanism and none by the same one
  twice. The authoring position — writing a mechanical check for a docs-site inventory — failed in
  three consecutive iterations; iter4 moved the carrier rather than substituting a fourth regex, and
  that is the repair whose durability is untested.
- **Locale placeholders are the blind spot most likely to recur.** Korean translates the placeholder
  itself (`.moai/reports/<카드-id>/`, `M<마일스톤>`), so an English-token grep cannot reach `ko` in
  principle. An English-only inventory sees 3 files where there are 12. Any future check over
  docs-site must be locale-tolerant or it is measuring the wrong population.

---

## Verdict

**PASS — ready for the integration window.**

Not ready for: dispose of this worktree, or treat the local result as a CI verdict.
