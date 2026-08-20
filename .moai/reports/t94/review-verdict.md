# t94 review verdict — concurrency-cap rationale rebuild

- Reviewer: lead session (operator-confirmed direct verdict; t96 lead-verified precedent)
- Card: t94 · Worktree: `.claude/worktrees/t85` · Branch: `WT-t85` @ `4c5994237` (reviewed delta base `59dd62d02`)
- Delta reviewed: `59dd62d02..4c5994237` (18 files, +231/−38; 16 doc files + 2 evidence files)
- Lens: `--deep` (distributed template docs + neutrality-sensitive surface)
- Evidence read: `.moai/reports/t94/t94-evidence.md` (129 lines, 5-section)

## Verdict: PASS (delta-scoped — see §6 scope note)

## 1. Claims reviewed (evidence vs this review's direct reads)

| # | Claim | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | Commit 4c5994237 stacked on 59dd62d02 | `git log` + `git merge-base --is-ancestor` → ANCESTOR-OK | PASS |
| 2 | 18 files +231/−38 | `git diff --stat` — exact match; "16 files, +62" in evidence = doc-only count excluding the 2 evidence files (169 lines) — reconciles exactly | PASS |
| 3 | §C.2 rebuilt as three-limit SSOT table | Full diff read: subagent 20 HARD / workflow 16 (Mode 6) / team-size 3-5 ADVISORY (Mode 3); coincidence explicitly named; write fan-out never-two-writers clause retained | PASS |
| 4 | §C.4 added (factory workers ≠ Mode 4 fan-out) | Present; correctly references the registry/free-slot discipline t85 implemented | PASS |
| 5 | §E anti-pattern replaced, §F citations annotated, §A row + §B tree aligned | All confirmed in diff | PASS |
| 6 | t92 experimental frame untouched | §C.1, Mode 3 row, §G appear only as unchanged context in diff hunks | PASS |
| 7 | Pointing texts (6 + CLAUDE.md) | `moai-constitution.md` Mode 4 bullet read directly — new phrasing + §C.2 SSOT anchor resolves to the rebuilt heading | PASS |
| 8 | Mirror parity (8 pairs + CLAUDE.md pair) | Lane: `ALL 8 PAIRS IDENTICAL` + drift test ok. Reviewer shasum spot-check: 4/4 pairs byte-identical (orchestration-mode-selection, moai-constitution, CLAUDE.md, cache-aware-execution) | PASS |
| 9 | Template neutrality of added text | `git diff 59dd62d02..4c5994237 -- internal/template/ \| grep '^+' \| grep -E 'SPEC-…\|tNN\|2026-…\|sha'` → empty (forbidden classes 0) | PASS |
| 10 | Workflow 16-concurrent grounding | Lane gap ("runtime-documented, not observable in-session") — **upgraded by this review**: the Workflow tool description in THIS session documents `min(16, available CPUs − 2)` verbatim, corroborating the §C.2 row | PASS |

## 2. Evidence (this review's commands)

- `git log/merge-base/diff --stat/--name-only 59dd62d02..4c5994237`
- Full diff read of `orchestration-mode-selection.md` (the rebuild) + selected pointing files
- `git show @4c5994237 | shasum` mirror pairs ×4
- Template-side added-line neutrality grep
- Read: `.moai/reports/t94/t94-evidence.md` + `.moai/reports/t94/t85-run-evidence.md`

## 3. Baseline attribution

- Doc reads: tree @ `4c5994237`. Neutrality grep: template-side added lines of the same delta.
- Test outputs (mirror-drift / sanitized-pair / no-internal-content-leak, `ok 2.566s`/`5.987s`): lane-attributed, not re-run (same rationale as t85 verdict §3 — lane-local discipline, CI owns the full judgment).

## 4. Gaps

- 5 of 9 mirror pairs accepted from lane's `diff` claim + drift test (reviewer spot-checked 4).
- Test outputs lane-attributed (see §3).

## 5. Residual risks

- 8 skill-file sites still say "per the Mode 4 ceiling (§C.2)" — deliberate scope boundary (SSOT delegation; references resolve); cosmetic rename candidate for a follow-up card.
- Pre-existing real SPEC IDs in template `spec-workflow.md:331,347` — antecedent content, flagged by the lane for a possible future neutrality tightening; unchanged by t94.
- Semantic relaxation risk ("3-5 ceiling" → advisory under 20) closed by the retained never-two-writers clause + stagger-spawn forfeiture wording — verified present.

## 6. Scope note (post-verdict development)

After this review's delta, the operator's §6 final scope correction moved WT-t85's base to `9794f293d` (merge of origin/release/v3.1.1 = t96 lineage) with rework commits stacking: `-f`/`--factory` entry flag fully retired in favor of `-k N` unification, card-class path branching (A/B fanout vs C serial 3-phase), t96 overlap removal, and `-f` mention updates in t94's own files. **This verdict binds the reviewed delta only** (the three-limit logic is invariant); the rework commits will be reviewed as a fresh delta (`9794f293d..<new>`) per the run lane's re-review request. t85's 1st-pass PASS is likewise held as a non-integration trigger pending that re-review.
