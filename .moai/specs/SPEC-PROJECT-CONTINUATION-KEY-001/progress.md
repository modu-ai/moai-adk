# SPEC-PROJECT-CONTINUATION-KEY-001 — Progress

Card **t191** · Tier **M** · Branch `WT-project-continuation` · Baseline tree `2660bcd09`

| Field | Value |
|---|---|
| plan_status | revised (v0.3.0 — plan-audit iter-2 delta fix; Tier M ceiling exhausted, no iteration 3) |
| run_status | not started |
| sync_status | not started |

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.

**v0.2.0** — 12 GEARS requirements (`REQ-PCK-001..012`); 15 acceptance criteria (`AC-PCK-001..015`, 13 blocking).

Iteration-1 audit (`.moai/reports/t191/plan-audit.md`, FAIL 0.78) disposition:

| Audit defect | Class | Disposition |
|---|---|---|
| D1 `pipeline` ≡ `card`, gate clause unconstrained | blocking | **Accepted.** `spec.md` §3 D1.1-D1.3 rewritten; `REQ-PCK-006` states the delta positively (carry distance) + the kickoff-clause obligation; `AC-PCK-006` rebuilt with three conjuncts, conjunct 1 red on a `card`-behaving `pipeline`. |
| D2 wizard-localization gap false, fallback breaks a test | blocking | **Accepted in full.** `spec.md` §5 corrected (precedent is `audit_model`); `plan.md` §B item 1 withdrawn; `acceptance.md` §D.5 folding branch deleted; `AC-PCK-010` strengthened to 3 conjuncts naming the enforcing test. `REQ-PCK-010` unchanged — it was already deliverable. |
| D3 `AC-PCK-008` vacuous | blocking | **Accepted.** Replaced with the report's drafted per-branch kickoff-clause criterion; diff-stat re-filed as non-blocking `AC-PCK-014` and struck from `REQ-PCK-008`'s verification path. |
| D4 `REQ-PCK-011` "descriptions" ambiguous | minor | **Resolved toward the stronger reading** — `withOptionDesc`, 32 entries. Measurement added a finding the audit did not have: the `.opt.` guard keeps labels English by design, so `REQ-PCK-012` / `AC-PCK-015` were added to protect it. |
| D5 inventory citation off by one | minor | **REJECTED with evidence.** `grep -n` and `awk` both place the row at 2922-2924, as v0.1.0 stated. See `spec.md` §3 D5. |
| D6 `REQ-PCK-004` silent on P1-added options | minor | **Accepted.** `REQ-PCK-004` now enumerates the four options and omits `Create SPEC later`; `AC-PCK-004` asserts it. |

Coordinator-proposed progression-mode reading for `pipeline`: **evaluated and rejected in writing** (`spec.md` §3 D1.3) — `autonomous` is already the default (`goal.md:112-113`), so it reproduces the synonym defect; carry distance adopted instead.

New-in-v0.2.0 measurements, all on tree `2660bcd09`: `grep -c "QuestionTypeSelect" questions.go` → `12`; `translations_completeness_test.go:120-124` `continue` ⇒ 3 errors not 9; `f.workflow.audit.model.opt.*` present and translated in all four i18n maps while `applyI18n` resolves `.opt.` against English; inventory row at 2922-2924 by two methods.

## §E.1b Iteration-2 disposition (v0.3.0)

Iteration-2 audit (`.moai/reports/t191/plan-audit-iter2.md`, PASS-WITH-DEBT 0.80). The Tier M iteration ceiling is exhausted; this is a scoped delta fix, not a third audit round.

| Audit defect | Class | Disposition |
|---|---|---|
| iter2-D1 `REQ-PCK-007` contradicts `REQ-PCK-006` | critical | **Accepted.** `REQ-PCK-007` rewritten as invariant + three-item permitted-change list (recommended option / carry distance / wording); `AC-PCK-007` asserts both halves; `plan.md` M1 instructs the same sentence. Gate invariant strengthened, not weakened — it gained "pre-filled" and "bypassed". |
| iter2-D5 `REQ-PCK-006` omits the clarification precondition | major | **Accepted.** `REQ-PCK-006` extended with the `plan.md:53,73` ordering; `AC-PCK-006` gains conjunct 4; `plan.md` M1 `pipeline` bullet and §D carry it. |
| iter2-D2 `AC-PCK-011`'s `withOptionDesc` half has no red | major | **Accepted.** Verified in source: `allOptionDescFields` skips fields with no `OptionDesc` (`option_desc_test.go:27-28`), coverage tests compare locale maps to each other. M5 adds `TestProjectContinuationI18nKeysInAllLocales` asserting a non-empty `OptionDesc` on all three options — that assertion is the red. |
| iter2-D3 `AC-PCK-015` cites an audit-scoped map | major | **Accepted.** Citation replaced with `TestEveryOptionDescKeyAvoidsOptGuard` (`option_desc_test.go:50`), driven by `allOptionDescFields` → `settings.AllSections()`. The `app.js` tripwire citation is retained — it asserts on the asset, not a field set. |
| iter2-D4 header claim vs `AC-PCK-014` | minor | **Accepted.** Header qualified; the unfalsifiable criterion relocated to `plan.md` §D Constraints; ACs renumbered gapless (15 → 14). |
| iter2-D6 `card` tightened beyond P1 | minor | **Accepted.** §3 D1.1 now states the gap is widened from both ends and labels it a decision. |

Additional (not defects): the `AC-PCK-005`/`AC-PCK-006` differential pair is now named explicitly in `acceptance.md`; the three-segment settings write path is **RESOLVED** (measured — `schema_sections.go:380-384` ships 3- and 4-segment paths in this same section) and dropped from §5 Gaps; the 3-vs-9 error count is restated as scenario-dependent (3 = no `Options` slice, 9 = length-3 slice with empty `Desc`s, neither executed); `REQ-PCK-012`'s "guard shall not be modified" recorded as a scope observation.

Both v0.2.0 disputes were upheld by the auditor: D5 (inventory line) recorded as auditor error, and the "3, not 9" control-flow reading confirmed.

Ready for Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
