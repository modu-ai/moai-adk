# Evaluation Report — SPEC-PROJECT-CONTINUATION-KEY-001 (sync-audit, independent re-verification)

SPEC: SPEC-PROJECT-CONTINUATION-KEY-001 — `workflow.project.continuation`, a three-value `/moai project` completion-continuation key
Card: **t191** · Worktree `.claude/worktrees/t191` · Branch `WT-project-continuation`
Audited at: 2026-09-02, HEAD `f5a8a716d` (sync commit `c7ac03fe8` + §E.4 backfill `f5a8a716d`), base `2660bcd09`
Auditor: sync-auditor (independent, skeptical, fresh-judgment)
Scoring model: flat weighted (no `evaluator_mode: hierarchical` in harness.yaml)
Profile: built-in default (no `evaluator_profile` in spec frontmatter)

**Overall Verdict: PASS — weighted harmonic 94.0**

Both must-pass dimensions (Functionality, Security) clear their thresholds independently. **Blocking findings: 0.** Seven findings recorded, all classified `optional`.

---

## Dimension Scores

| Dimension | Score | Verdict | Evidence (re-executed this run, verbatim) |
|-----------|-------|---------|-------------------------------------------|
| Functionality (40%) | 96/100 | PASS | All six scoped packages green: `go test ./internal/config/... ./internal/settings/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/web/... -count=1` → `ok …/internal/config 2.594s` · `ok …/internal/settings 0.620s` · `ok …/internal/cli/wizard 5.689s` · `ok …/internal/core/project 3.300s` · `ok …/internal/web 5.655s` (+ 3 subpackages). 14/14 ACs independently re-verified (per-AC table below), incl. all 13 blocking. **Three non-vacuity mutants reproduced with correct attribution** (§ Mutants). Pre-P1 option set verified against the actual `e91def4ca^` text, not against the SPEC's assertion. AC-008 count re-measured `3` vs pinned baseline `1`. Mirror parity re-measured `0/0/0/1`. |
| Security (25%) | 95/100 | PASS | No trust boundary crossed. Value origin is closed: `QuestionTypeSelect` with three `config.ProjectContinuation*` constants (`questions.go:409-411`); `grep -rn "project-continuation\|projectContinuation" --include='*.go' .` → **no match** (no CLI flag path). Read side fails safe: `ProjectContinuation()` nil-receiver-safe, absent→`card`, unmatched→`card` + reported, never applied (`project_continuation.go:56-71`), verified by `--- PASS: TestProjectContinuation/unmatched_value_falls_back_to_card_and_is_reported` and `--- PASS: TestProjectContinuationNilReceiver`. No secrets, no credentials, no network surface, no dependency-manifest change (`go.mod`/`go.sum` absent from `git diff --name-only 2660bcd09..HEAD`). The `applyI18n` `.opt.` guard is untouched: `git diff --stat 2660bcd09..HEAD -- internal/web/assets/app.js` → empty; `app.js:273` still reads `var str = (key.indexOf(".opt.") >= 0 ? enDict : dict)[key];`. One latent (unreachable) note: F6. |
| Craft (20%) | 88/100 | PASS | `golangci-lint run` over the five touched packages → `0 issues.` (rc 0). `go build ./...` → rc 0. `GOOS=windows GOARCH=amd64 go build ./...` → rc 0. Coverage: `internal/web` `coverage: 66.8% of statements` and `internal/config` 80.5% both sit under the 85% package target — **attribution established, both pre-existing** (F4/F5). Documentation and evidence discipline are exceptional: the `progress.md` §E matrix carries per-AC commands and outputs, every claim I spot-checked reproduced, and three findings the plan phase had not measured were self-reported rather than absorbed. Deductions: two coverage packages under target and no regression guard on the prose file (F2). |
| Consistency (15%) | 96/100 | PASS | Every established repo idiom followed: wizard select shape matches its `audit_model` neighbour (`questions.go`, same group, same `Label/Value/Desc` triple); schema uses the house `withOptionDesc(closedSeam(...))` composition (`schema_sections.go:374-376`); writer mirrors `writeWorkflowTodoYAML` incl. the empty-string "never asked" convention (`initializer_expansion.go:83-108`); inventory class **P = prose-consumed** is a precedented classification from SPEC-CONFIG-KEY-HONESTY-001, not an escape hatch invented here — the only sibling P row is `workflow.worktree.auto_cleanup` (`shipped_key_inventory.yaml:2926`), classified identically with its prose consumer named; new test follows the per-SPEC enforcer pattern and reuses the `i18nKeyInAllLocales` helper; Template-First mirror parity holds; Conventional Commits with SPEC-ID scope and card id on all 9 commits. Korean comment in the local `workflow.yaml` matches that file's own established style (`documentation: ko`), English in the template per neutrality §2.1. |

**Weighted harmonic mean**: `1 / (0.40/96 + 0.25/95 + 0.20/88 + 0.15/96)` = `1 / 0.01063347` ≈ **94.04 → 94.0**

Must-pass firewall: Functionality 96 ≥ 85 PASS · Security 95 ≥ 85 PASS. Neither forces a FAIL.

---

## AC-by-AC re-verification (14 criteria, 13 blocking)

| AC | Blocking | Verdict | Independent evidence (this run) |
|---|---|---|---|
| 001 | yes | PASS | `--- PASS: TestValidProjectContinuations (0.00s)` |
| 002 | yes | PASS | `--- PASS: TestProjectContinuation/project_block_absent`, `/continuation_key_absent_inside_a_present_project_block`, `--- PASS: TestProjectContinuationNilReceiver` |
| 003 | yes | PASS | `--- PASS: TestProjectContinuation/unmatched_value_falls_back_to_card_and_is_reported`; source confirms `return ProjectContinuationCard, configured` (`project_continuation.go:70`) — no error, no panic |
| 004 | yes | PASS (prose) | **Verified against the measured pre-P1 baseline, not the SPEC's claim.** `git show e91def4ca^:…/doc-generation.md` Step 4.2 lists exactly: `Create SPEC (Recommended)` · `Review and Edit Documentation` · `Generate project-specific harness` · `Done`. The shipped `none` block reproduces all four **verbatim**, in order, and carries no `Create SPEC later`. Cap clause present: "the four-option cap is therefore satisfied with **nothing routed to `Other`**". Issuance skip explicit at Step 4.1.5 step 0: "Steps 1-5 below do not run, and no `moai todo add` invocation is reachable on this path." |
| 005 | yes | PASS (prose) | Row names `Create the SPEC and start now`. Terminal imperative: "**The session stops when `/moai plan` returns**; run-phase entry is a separately-initiated operator action in some later turn, and this branch neither proceeds to it nor emits its gate." Standing-source steps 1-5 unmodified — the file's complete removed-line set across `2660bcd09..HEAD` is 5 lines (the `[HARD]` issuance preamble, the Step 4.2 header, the old `card` option, the cap clause, the no-branch clause); **none of steps 1-5 appears in it**. |
| 006 | yes | PASS (prose) | (1) "**The session does not stop when `/moai plan` returns**: it continues past plan completion to the run-phase boundary and **emits the Implementation Kickoff Approval gate in this same session**". (2) `grep -c` restricted to the row = 1 ≥ 1. (3) Positive assertion present: "**emits that gate and stops for the operator's answer**; it never selects, answers, pre-fills, or defaults it, and if no answer comes back nothing starts." (4) Ordering clause present and correctly conditional: "the gate is emitted only once the plan phase's `[NEEDS CLARIFICATION]` markers are resolved (`plan.md:53`, `plan.md:73`); where markers remain open, this branch stops at their resolution rather than at the gate." |
| 007 | yes | PASS (prose) | Both clauses present and explicitly universal. `doc-generation.md:396` gained "**This clause binds at every value of `workflow.project.continuation`**". The new `[HARD]` sentence carries **both** halves: invariant ("never skipped, auto-answered, pre-filled, defaulted-on-no-answer, or bypassed") and **three** permitted changes ("which option is recommended, how far that option carries the session, and how that option is worded"). Neither is scoped to a subset. |
| 008 | yes | PASS | `grep -c "Implementation Kickoff Approval" .claude/skills/moai/workflows/project/doc-generation.md` → **`3`**; same grep against `git show 2660bcd09:…` → **`1`**. Hits at :367 (carry-distance table), :380 (`card` row), :388 (`pipeline` row). Both run-phase-offering options carry the sentence **verbatim identically** ("Run-phase entry still passes the Implementation Kickoff Approval gate, where the `ac_converge` goal is offered as the progression-mode axis and armed only after the gate passes — this branch never bypasses that gate."); `none` offers no run-phase entry and correctly carries none. 3 ≥ 2. |
| 009 | yes | PASS | `--- PASS: TestShippedConfigKeysHaveReaders (0.89s)` (4 subtests). Mutant M1 below confirms non-vacuity. Row present at `shipped_key_inventory.yaml:2871-2873`, `class: P`, `evidence: .claude/skills/moai/workflows/project/doc-generation.md`. |
| 010 | yes | PASS | `--- PASS: TestProjectContinuationQuestion` · `TestProjectContinuationTranslationsExist` · `TestProjectContinuationNotOptionTranslationExempt` · `TestWizardQuestionTranslationCompleteness`. Source read: `Type: QuestionTypeSelect`, `Default: config.ProjectContinuationCard`, option values are the three `config.ProjectContinuation*` constants. ko/ja/zh each carry non-empty `Title` + `Description` + a 3-element `Options` slice with non-empty `Label` **and** `Desc` (`translations.go:137-145`, `:305-313`, `:473-481`). |
| 011 | yes | PASS | `--- PASS: TestProjectContinuationI18nKeysInAllLocales`. `grep -c "f.workflow.project.continuation" internal/web/assets/i18n.js` → **`32`**; per-key sweep confirms the distribution is exactly 8 keys × 4 locales (title 4, desc 4, opt.{none,card,pipeline} 4 each, option.{none,card,pipeline} 4 each). Empty-value scan → `0`. Mutant M2 below confirms the wrapper assertion is the live red. |
| 012 | yes | PASS | `cmp` matrix re-measured: `doc-generation.md` rc=0 · `todo.md` rc=0 · `tab_schema.json` rc=0 · `workflow.yaml` rc=1. **Key present on both sides with the same value**: local `:135-136` and template `:46-47` both `project:` / `continuation: card`. Divergence delta base→HEAD is +23 lines, accounted for entirely by the deliberately-asymmetric comment blocks (3-line Korean local, 16-line English template) plus hunk realignment — no substantive divergence introduced. |
| 013 | **no** | PASS (prose) | `doc-generation.md:357` carries the line `workflow.project.continuation: "<offending value>" is not one of none \| card \| pipeline — resolved to card.` and the negative case: "When no offending value was recorded, that line is absent." |
| 014 | yes | PASS | Conjunct 1 (prose): schema row uses `"f.workflow.project.continuation.opt."` for labels. Conjunct 2: `"f.workflow.project.continuation.option."` — no `.opt.` substring (`opti` ≠ `opt.`); mutant M3 below confirms the enforcer fires. Conjunct 3: `app.js` absent from the entire change diffstat; guard intact at `:273`. `--- PASS: TestOptionLabelsStayEnglishStillGuarded`, `TestEveryOptionDescKeyAvoidsOptGuard`, `TestEveryOptionDescKeyTranslatedInAllLocales`, `TestEveryOptionDescRendersOnConsolePage`. |

---

## Mutants — non-vacuity established by execution, not by citation

Each mutant was applied, measured, and reverted; `git status --short` confirmed a clean tree after each.

**M1 — AC-009. Remove the inventory row alone** (`sed -i '' '2871,2873d'`):
```
FAIL	github.com/modu-ai/moai-adk/internal/config	1.249s
          workflow.project.continuation
```
The failure names the key by path. Reverted; `git status --short internal/config/testdata/` → empty.

**M2 — AC-011. Drop the `withOptionDesc` wrapper** (bare `closedSeam`, i18n dictionary left intact):
```
--- FAIL: TestProjectContinuationI18nKeysInAllLocales (0.00s)
    project_continuation_i18n_test.go:50: field "workflow.project.continuation" option "none" carries no OptionDesc — the withOptionDesc wrapper is missing, and no existing sweep can catch that (they skip fields with no OptionDesc)
    … same for "card" and "pipeline"
FAIL	github.com/modu-ai/moai-adk/internal/web	3.680s
ok  	github.com/modu-ai/moai-adk/internal/settings	0.565s
```
**This measurement also settles hunt 6 (see § Hunt adjudications).** The new test was the *only* failing test in `internal/web` under this mutant — every other test in that package, `TestI18nUntranslatedValues` included, passed. Reverted; green restored.

**M3 — AC-014 conjunct 2. Rename the description prefix to `…continuation.opt.desc.`**:
```
--- FAIL: TestEveryOptionDescKeyAvoidsOptGuard (0.00s)
    option_desc_test.go:54: field "workflow.project.continuation" option "none" OptionDesc "f.workflow.project.continuation.opt.desc.none" contains ".opt." — the G1-2 guard resolves it against the English dictionary, so its ko/ja/zh translation never renders
    … same for "card" and "pipeline"
```
Confirms the iter2-D3 citation correction (`option_desc_test.go:50`, all-sections scope) is the enforcer that actually fires. Reverted.

---

## Hunt adjudications

### Hunt 1 — the six prose ACs

**Adjudication: all six hold on the shipped text, and the two I was pointed at hold on their hardest reading.** I read the shipped prose myself rather than accepting the record's reading, and I resolved AC-004 against a source the record did not use.

- **AC-006 conjunct 4** (the `[NEEDS CLARIFICATION]` ordering clause) is present, correctly **conditional**, and correctly directional: the row does not merely mention clarification, it states where the branch stops when markers remain open ("stops at their resolution rather than at the gate"). A row that carried the gate unconditionally — the failure the conjunct was written for — would read differently. Satisfied.
- **AC-004** is the one I would not accept on the record's word, because it asserts fidelity to a *historical* option set. I recovered that set from `git show e91def4ca^` and compared it to the shipped `none` block: four options, same order, same wording, and no `Create SPEC later`. The `Other`-routing clause and the issuance-skip clause are both explicit. Satisfied against the measured baseline.
- **AC-005/006/007/013** are satisfied on their literal terms, all quoted in the AC table above.

The prose exemption is disclosed in the SPEC's own §D.3 rather than hidden, and the disclosure is accurate — but disclosure is not verification, so I verified them. What the exemption genuinely costs is **durability, not correctness**: see hunt 4 / F2.

### Hunt 3 — `pipeline` has no mechanical carrier

**Adjudication: the carrier is defensible for this SPEC, and the honest framing is a disclosed limitation rather than a missing blocking criterion. No blocking finding. But the record's framing understates one asymmetry, which I record as F1.**

Three things make it defensible:

1. **The classification is honest and precedented, not invented.** `class: P` (prose-consumed) is an existing category from SPEC-CONFIG-KEY-HONESTY-001 with a documented legend (`shipped_key_inventory.yaml:4` — `W=wire, P=prose-consumed, R=reserved, D=delete`), and the row names its prose consumer in the `evidence` field exactly as the legend requires. There is one sibling P row (`workflow.worktree.auto_cleanup`, `:2926`) classified the same way. The key is not masquerading as wired: `grep -rn "ProjectContinuation(\|ProjectContinuationForRoot(" --include='*.go' internal/ cmd/ | grep -v _test.go` returns **only the definitions** — no production caller branches on it, and the source says so in a `@MX:NOTE` at `project_continuation.go:19`.
2. **The mechanical alternative was evaluated and rejected in writing** (`spec.md` §3 D1.3), with its measurement — the `progression_mode` reading reproduces the synonym defect because `autonomous` is already the default. Rejecting an alternative with a reason is a design decision; rejecting it silently would be the defect.
3. **A blocking criterion for a mechanical carrier would have been unsatisfiable at this tier.** `/moai project` Phase 14 is orchestrator-executed prose end to end; there is no Go execution path for the criterion to attach to. Demanding one would have required a different SPEC, not a stricter AC.

**Where I depart from the record.** The residual-risk paragraph treats all three values symmetrically. They are not. `none` and `card` fail *safe* — `card` is the status quo, `none` is a subtraction, and an orchestrator that ignores either lands on defensible behaviour. `pipeline` is the only value that asks the orchestrator to do **more** than the default, so a non-compliant orchestrator degrades `pipeline` silently into `card`, and the operator who configured `pipeline` gets `card` behaviour **with no signal anywhere** — no error, no report line, nothing. The unmatched-value path got a report line precisely because a silent fallback "would hide the typo for the life of the project" (`project_continuation.go:50-55`); the same reasoning applies to a silently-unhonoured `pipeline` and was not extended to it. That is F1: worth a follow-up card, not a blocking defect here.

### Hunt 4 — could a future edit collapse the differential pair?

**Adjudication: yes, and nothing would catch it. The pair is a plan-time guarantee, not a regression guard.** This is F2.

The pair works exactly as claimed *against an implementer shipping this SPEC*: I confirmed no single row text satisfies both, and the shipped rows are genuinely differential — `card` says "**The session stops when `/moai plan` returns**", `pipeline` says "**The session does not stop when `/moai plan` returns**". Direct negations; a synonym implementation is impossible without one of them going red.

But `AC-PCK-005` and `AC-PCK-006` conjunct 1 are prose criteria evaluated once, at audit time. `grep -rn "doc-generation" --include='*_test.go' internal/` → **`0`**: no test in the repository asserts anything about this file's contents. After merge, an editor who rewrites the `pipeline` row to terminate at `/moai plan` — collapsing it back into `card` — trips nothing. The anti-synonym mechanism protects the implementation, not the file.

A cheap guard exists and is stated as the required fix in F2. I classify it `optional` deliberately: the SPEC did not require it, adding it now would be scope expansion, and the Enforce Simplicity brake says an evaluator-found improvement is not automatically work.

### Hunt 7 — in-place amendment of a self-authored CHANGELOG entry

**Adjudication: within the sync lane's authoring boundary, and the resulting entry is accurate. No finding against the card.**

Measured facts:
- `git show 7ea775a19 -- CHANGELOG.md` — this branch's own M6 run-phase commit authored the bullet, labelled "run-phase (card t191)".
- `git show c7ac03fe8 -- CHANGELOG.md` — the sync commit amended that same bullet: **1 insertion / 1 deletion**, no new bullet, no adjacent entry touched.
- `git diff 2660bcd09..HEAD -- CHANGELOG.md` — the branch's total contribution to the file is **exactly one added line**.
- `grep -c 'SPEC-PROJECT-CONTINUATION-KEY-001' CHANGELOG.md` → **`1`**.

Three reasons the amendment is within boundary:

1. **It preserves the invariant B12 protects.** B12's halt exists to stop a second session appending a duplicate bullet for one SPEC. The hit here was self-authored on the same unpushed branch. Appending would have produced `grep -c == 2` — the precise state B12 forbids. Amending kept it at 1. The agent read the rule's *purpose*, not just its trigger, which is the correct reading when the two diverge.
2. **Scope discipline held.** The edit is confined to one line; `git show` confirms no restructuring of surrounding entries, and the sync record explains the confinement in merge terms (t315's unrelated CHANGELOG edit touches a different line, so both survive the develop merge — neither is to be resolved by deletion).
3. **It was surfaced, not buried.** §E.4 opens with a heading naming the non-clean grep, states the attribution, and states the reasoning. Under "manage confusion actively", surfacing a rule whose literal trigger fired for a reason the rule did not anticipate is the required behaviour.

**Accuracy of the resulting entry.** I spot-checked its load-bearing claims: `internal/config/testdata/shipped_key_inventory.yaml` class-P row — present at `:2871-2873`, and the corrected full path resolves (the run-phase version's bare basename did not). "8 i18n keys × 4 locales" — re-measured `32`. "with the row declared bare, the whole `internal/web` + `internal/settings` suite was green" — this is a counterfactual claim about a state that no longer exists, so I reproduced it: under mutant M2 the *only* failing test in `internal/web` was the new enforcer, and `internal/settings` was `ok`. Absent the new test, the suite would indeed have been green. The claim is corroborated. The phase label "sync-phase close (3-phase plan→run→sync)" matches the artifacts (`spec.md` `status: completed`, three phases in `progress.md`).

One doctrine observation, against the rule rather than the card: B12 as written offers no carve-out for a self-authored same-branch hit, so the lane had to reason its way out of a halt. That gap will recur on every SPEC whose run phase writes its own CHANGELOG entry. Recorded as F7.

---

## Findings

Ranked by severity. **Blocking: 0.** All seven are `optional` — none affects correctness or a requirement the SPEC states, and per the finding-consumption brake none is auto-routed into fixes.

- **F1** [Medium] [optional] `.claude/skills/moai/workflows/project/doc-generation.md:388` + `internal/config/project_continuation.go:39-55` — **`pipeline` degrades to `card` silently.** It is the only value asking the orchestrator to do more than the default, so a non-compliant orchestrator produces `card` behaviour with no signal reaching the operator. The unmatched-value path got a report line for exactly this reasoning ("a silent fallback would hide the typo for the life of the project"); it was not extended to a silently-unhonoured `pipeline`. Confidence: high that the gap exists; medium that it is worth closing. — **Required fix (follow-up card, not this one):** either give `pipeline` a mechanical carrier and re-derive AC-006 conjunct 1 against it (as `acceptance.md` §D.5 already anticipates), or add one line to the Step 4.2 report contract naming the resolved value, so an operator can see which branch the session actually took.
- **F2** [Medium] [optional] `.claude/skills/moai/workflows/project/doc-generation.md:380,388` — **the AC-005/006 differential pair has no regression guard.** `grep -rn "doc-generation" --include='*_test.go' internal/` → `0`. The pair is a plan-time argument, evaluated once; a later edit collapsing `pipeline` back into `card` trips nothing. Confidence: high (measured absence of any guard). — **Required fix:** add a test in `internal/template` or `internal/web` asserting both `.claude/skills/moai/workflows/project/doc-generation.md` and its template mirror contain the literal `The session stops when \`/moai plan\` returns` **and** `The session does not stop when \`/moai plan\` returns`. Two `strings.Contains` assertions; it makes the differential pair survive merge.
- **F3** [Low] [optional] `spec.md` §3 D6 — **the mirror table omits `.claude/skills/moai/workflows/project.md`**, which M1 edited. Confirmed exactly as self-reported: `cmp` rc=1, `diff` shows a single divergence `102d101 < Last Updated: 2026-02-21`, and I verified that divergence is **pre-existing** — the identical `102d101` line separates the pair at base `2660bcd09`. The substantive Phase 14 row edit is byte-identical on both sides. Documentation-completeness gap only; no parity defect. Confidence: high. — **Required fix:** none required; if D6 is ever revised, add the fourth pair with its known rc=1 and the reason.
- **F4** [Low] [optional] `internal/web` — **coverage 66.8%, under the 85% package target.** Attribution established rather than asserted: `git diff --name-only 2660bcd09..HEAD -- 'internal/web/*.go' | grep -v _test.go | wc -l` → **`0`**. The change added zero non-test Go statements to this package, so the denominator is unchanged and the new test can only have raised the figure. **Pre-existing, not introduced.** Confidence: high. — **Required fix:** none in this card's scope; a separate coverage card owns `internal/web`.
- **F5** [Low] [optional] `internal/config` — **coverage 80.5%, under the 85% target.** Also pre-existing in direction: the new file's own functions measure 88.9% / 83.3% (`progress.md` §E.3), above the package figure, so `project_continuation.go` raises rather than lowers it. Confidence: medium — I did not re-measure the package's base figure, only established the new code's direction. Recorded as such. — **Required fix:** none in this card's scope.
- **F6** [Low] [optional] `internal/core/project/initializer_expansion.go:91` — **latent YAML interpolation in the no-deployer fallback.** `fmt.Sprintf("workflow:\n    project:\n        continuation: %s\n", opts.ProjectContinuation)` interpolates the value into YAML without validating it against `ValidProjectContinuations()`. **Currently unreachable:** the only writer is the closed wizard select and there is no CLI flag (verified). It matches the sibling writers' established pattern, so it is not a regression, and the read side falls back safely regardless. Recorded for defense-in-depth only. Confidence: high that the pattern is as described; high that it is presently unreachable. — **Required fix:** if a non-wizard caller is ever added to `InitOptions.ProjectContinuation`, validate against the closed set before this write.
- **F7** [Info] [optional] Doctrine, not the card — **B12's CHANGELOG halt has no self-authored carve-out.** The grep trigger fires identically for a parallel-session duplicate (which must halt) and a same-branch self-authored entry (where amending is correct and appending is the defect). The lane reasoned correctly but had to reason at all, and the situation recurs on every SPEC whose run phase writes its own entry. Confidence: high. — **Required fix (doctrine owner, not this card):** add a clause: where `git log --oneline -- CHANGELOG.md` attributes the hit to a commit on the current unpushed branch, amend in place rather than halting; a hit attributable to any other branch still halts.

---

## Gaps — what I did NOT observe

- **`make build` was not re-run by this audit.** I verified source-level mirror parity by `cmp` (rc 0/0/0/1) and that `go build ./...` compiles the embedded tree, but I did not reproduce the record's `strings bin/moai | grep -c "workflow.project.continuation"` → `39`, and I did not inspect any installed binary.
- **No full test suite.** `go test ./...` is prohibited on this shared machine. I ran the six scoped packages only. **CI owns the full-suite verdict**; nothing in this report claims one.
- **`internal/config` base coverage was not measured.** I established the new file's direction (its functions score above the package figure) but did not check out `2660bcd09` to read the package's prior number — a branch switch is forbidden in this checkout. F5's attribution is therefore directional, not exact.
- **No runtime execution of `/moai project`.** The six prose criteria were verified by reading the shipped instruction text, which is what the orchestrator consumes; no Phase 14 run was performed under any of the three values. This is the SPEC's own §D.3 boundary and is the sole verification path available.
- **No cross-model audit backend was invoked.** `audit_multi` / `codex_audit` / `glm_audit` were not called; this verdict is the Claude anchor alone.
- **CI, PR, and push state not inspected.** The lane does not push, no PR exists, and I requested no CI run.
- **`plan.md` §D constraints were not independently swept.** The record claims `preserve_list_post_run_count: 0` over three gate-owning rule files; the change's diffstat contains none of them, which is consistent, but I did not enumerate the preserve list itself.
- Nothing else was left unobserved. Every claim in the Dimension Scores and AC tables above was produced by a command run against HEAD `f5a8a716d` in this worktree during this audit.

## Residual risk

- **The verdict on six criteria rests on reading, and reading is not repeatable by a machine.** A later reader disagreeing with my reading of a prose row has no test to settle it. F2 narrows this for the two rows that matter most; it does not eliminate it.
- **`pipeline` may never have behaved as specified in any real session.** Nothing in this audit, or in the run phase, executed it. Its correctness is the correctness of an instruction, verified as an instruction.
- **Mirror parity is verified at source, not at install.** A `make build` regression between this tree and a released binary would not be visible to the `cmp` matrix.
- **The three mutants were reverted by file restore and confirmed clean by `git status --short`.** If any mutant left an artifact outside the paths I checked (`internal/config/testdata/`, `internal/settings/`), it would be invisible to this audit — though the pre-commit `git status --short` for the report commit is a second check on that.
- **Coverage on two packages sits under target and this card does not move it.** The debt is recorded and attributed, not repaid.

---

## Recommendations

1. **Open a follow-up card for F1 + F2 together.** They share a root: `pipeline` is prose-carried, so it can be neither observed at runtime nor protected at rest. The `doc-generation.md` differential-string test (F2) is a few lines and is the higher-value half — it makes the SPEC's central anti-synonym argument survive the merge.
2. **Raise F7 with the doctrine owner.** It costs one clause and removes a recurring judgment call from every future sync lane.
3. **Leave F3-F6 as recorded observations.** F3 is already self-reported and factually closed; F4/F5 belong to coverage cards that own those packages; F6 is unreachable and pattern-consistent. Chasing them would add scope without reducing risk.
4. **Where the implementation is right, it is unusually right.** The three mutants, the honest `class: P` declaration with a `@MX:NOTE` saying the enum has no Go caller, the rejected-in-writing `progression_mode` alternative, and the self-reported §E.4 findings mean this card's weakest point is a disclosed property of a deliberate design choice — not something an audit had to uncover.

---

Auditor: sync-auditor · Card t191 · HEAD `f5a8a716d` · 2026-09-02
