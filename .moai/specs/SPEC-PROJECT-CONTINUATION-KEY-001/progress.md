# SPEC-PROJECT-CONTINUATION-KEY-001 — Progress

Card **t191** · Tier **M** · Branch `WT-project-continuation` · Baseline tree `2660bcd09`

| Field | Value |
|---|---|
| plan_status | revised (v0.3.0 — plan-audit iter-2 delta fix; Tier M ceiling exhausted, no iteration 3) |
| run_status | complete (M1-M6, 14/14 AC PASS; base `2660bcd09` → HEAD `7ea775a19`, unpushed) |
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

Run-phase tree: worktree `.claude/worktrees/t191`, branch `WT-project-continuation`, base `2660bcd09`, HEAD `7ea775a19`. Six milestone commits (`aed638b8b` M1 · `554652a19` M2 · `54390e4db` M3 · `29c580f27` M4 · `c46f109d3` M5 · `7ea775a19` M6). Not pushed — this lane does not push.

### AC matrix (14 criteria, 13 blocking)

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-PCK-001 | PASS | `go test ./internal/config/ -run TestValidProjectContinuations` | `--- PASS: TestValidProjectContinuations (0.00s)` / `ok … 0.223s` — asserts members in order, not length |
| AC-PCK-002 | PASS | `go test ./internal/config/ -run TestProjectContinuation` | `--- PASS: TestProjectContinuation/project_block_absent`, `/continuation_key_absent_inside_a_present_project_block`, and `--- PASS: TestProjectContinuationNilReceiver` |
| AC-PCK-003 | PASS | same | `--- PASS: TestProjectContinuation/unmatched_value_falls_back_to_card_and_is_reported`. Mutant (`return configured, ""`) reds BOTH conjuncts: `value = "pipelien", want "card"` + `unmatched = "", want "pipelien"` |
| AC-PCK-004 | PASS (prose) | read Step 4.1.5 step 0 + the `none` block of the Step 4.2 table | `none` skips issuance ("Steps 1-5 below do not run, and no `moai todo add` invocation is reachable on this path"); recommends `Create SPEC` in the pre-P1 wording; lists exactly four options; omits `Create SPEC later`; states the cap is met "with **nothing routed to `Other`**" |
| AC-PCK-005 | PASS (prose) | read the `card` block + `git diff 2660bcd09..HEAD` over the file | Recommends `Create the SPEC and start now`; terminal instruction is `/moai plan` ("**The session stops when `/moai plan` returns**; … this branch neither proceeds to it nor emits its gate"). The diff removes 5 lines, none of them Step 4.1.5 steps 1-5 — the five standing-source steps are unmodified |
| AC-PCK-006 | PASS (prose) | read the `pipeline` block | (1) "**The session does not stop when `/moai plan` returns**: it continues past plan completion … and **emits the Implementation Kickoff Approval gate in this same session**"; (2) `grep -c` restricted to the row = `1`; (3) "**emits that gate and stops for the operator's answer**; it never selects, answers, pre-fills, or defaults it"; (4) "the gate is emitted only once the plan phase's `[NEEDS CLARIFICATION]` markers are resolved … where markers remain open, this branch stops at their resolution rather than at the gate" |
| AC-PCK-007 | PASS (prose) | `grep -n "No branch is taken on the operator's behalf"` + read the new invariant sentence | Both clauses present and value-independent (the no-branch clause now adds "**This clause binds at every value**"). The new [HARD] sentence carries both halves: the invariant ("asked … never skipped, auto-answered, pre-filled, defaulted-on-no-answer, or bypassed") and **three** permitted changes (recommended option / carry distance / wording) |
| AC-PCK-008 | PASS | `grep -c "Implementation Kickoff Approval" .claude/skills/moai/workflows/project/doc-generation.md` | `3` (pre-change baseline `1`, re-measured on this tree at pre-flight). Per-branch: `card` row = 1, `pipeline` row = 1 — both run-phase-offering options carry it, `none` does not |
| AC-PCK-009 | PASS | `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders` | `ok … 1.354s`. Non-vacuity established as the RED: with the template key added and the row withheld, `REQ-CKH-008 anti-rot: 1 shipped config key(s) are NOT in the triage inventory: workflow.project.continuation` |
| AC-PCK-010 | PASS | `go test ./internal/cli/wizard/` | `ok … 3.339s`. Covers `TestProjectContinuationQuestion` (type/default/values), `TestProjectContinuationTranslationsExist` (3 locales × Title/Description/3 options with Label+Desc), `TestProjectContinuationNotOptionTranslationExempt`, and `TestWizardQuestionTranslationCompleteness` |
| AC-PCK-011 | PASS | `go test ./internal/web/ -run TestProjectContinuationI18nKeysInAllLocales` | `--- PASS`. `grep -c '"f.workflow.project.continuation'` = `32` (8 keys × 4 locales). RED executed first — see §E.2 gap 1 below |
| AC-PCK-012 | PASS | `make build`, then `cmp` over the four pairs | `doc-generation.md` rc=0 · `todo.md` rc=0 · `tab_schema.json` rc=0 · `workflow.yaml` rc=1 (`differ: char 17, line 2`). The `workflow.yaml` diff carries `continuation: card` on **both** sides, differing only in comment prose |
| AC-PCK-013 | PASS (prose, non-blocking) | read the Step 4.2 report contract | Carries the line `workflow.project.continuation: "<offending value>" is not one of none \| card \| pipeline — resolved to card.` and states "When no offending value was recorded, that line is absent" |
| AC-PCK-014 | PASS | `go test ./internal/web/` | Conjunct 1 (prose): labels use the `.opt.` prefix, descriptions use `.option.` (schema row read). Conjuncts 2-3: `ok`. Mutant — renaming the prefix to `…continuation.opt.desc.` reds `TestEveryOptionDescKeyAvoidsOptGuard` (`option_desc_test.go:50`) naming all three options |

### The three plan-phase gaps, closed by execution

1. **`option_desc_test.go:27-28` skips a bare field — CONFIRMED, and stronger than the SPEC states.** With the row declared as a bare `closedSeam` and only its five label keys in `i18n.js`, `go test ./internal/web/... ./internal/settings/...` was **fully green**: all three `allOptionDescFields` sweeps passed vacuously and `TestI18nKeyCoverageForward/Reverse` passed. The new `TestProjectContinuationI18nKeysInAllLocales` then failed on exactly that field, naming all three options plus the three absent `.option.` keys. Adding `withOptionDesc` + the 3 keys turned it green.
2. **`option_desc_test.go:50` is open-scope and reaches this field — CONFIRMED.** Mutant: `withOptionDesc(..., "f.workflow.project.continuation.opt.desc.")` → `TestEveryOptionDescKeyAvoidsOptGuard` FAIL on `none`/`card`/`pipeline`. Reverted; green restored. The iter2-D3 citation correction is validated by execution.
3. **`make build` — run, exit 0.** It produced **zero** tracked-file changes (templates are embedded via `//go:embed all:templates`; there is no generated file to commit). The built binary carries the edits: `strings bin/moai | grep -c "workflow.project.continuation"` = `39`, and `grep -c "Create the SPEC and continue to the kickoff gate"` = `2`.

### Contradiction between execution and the SPEC's written claims

None on the three gaps — all three executions confirmed what the SPEC read from source.

One **unanticipated finding**, reported rather than absorbed: a fourth existing test, `TestI18nUntranslatedValues` (`i18n_governance_test.go:221`), also fires on this field — but on a **different axis** (identical `.opt.` label VALUES across locales), not the missing `OptionDesc`. It was satisfied by localizing the three `.opt.` labels per locale following the `audit.gate.opt.*` precedent (`Off`/`끔`/`オフ`/`关闭`), which `applyI18n` still resolves from English at render. Satisfying it leaves the `OptionDesc` hole completely open, so it is **not** a substitute enforcer and does not weaken `AC-PCK-011`'s case for the new test. The SPEC did not predict this test; it did not contradict the SPEC either.

### Scope note

Adding a question to the wizard's Quality & Workflow group shifted three pre-existing hardcoded counters (`wizard_test.go` page-3 denominator 18→19; `expansion_test.go` `TotalVisibleQuestions` 17→18 and group count 11→12, plus the `Page3Questions` want-list; `restructure_test.go` membership list). Updated as the mechanical consequence of the SPEC's own deliverable, each with a comment naming this SPEC.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-02
run_commit_sha: 7ea775a19
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0   # the three gate-owning rule files (plan.md §D) untouched
l44_pre_commit_fetch: n/a          # lane does not push; no remote interaction this run
l44_post_push_fetch: n/a           # no push performed (lane constraint)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: exit 0            # go build ./...
  windows_amd64: exit 0   # GOOS=windows GOARCH=amd64 go build ./...
coverage_touched_packages:
  internal/config: 80.5%          # project_continuation.go — ProjectContinuation 88.9%, ForRoot 83.3%
  internal/settings: 90.3%
  internal/cli/wizard: 92.0%
  internal/core/project: 88.2%
  internal/web: 66.8%
lint: "0 issues"                  # golangci-lint run --timeout=5m over the five touched packages
total_run_phase_files: 20
m1_to_mN_commit_strategy: "one commit per milestone, M1..M6, each naming card t191"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
