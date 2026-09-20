# SPEC-JEV-OPTIN-MEASURE-001 — Progress

Card: t1020 · Tier L · plan-phase artifacts authored 2026-09-20. Split from `SPEC-JEV-INTEGRATION-001` on the M2+M3 seam.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | L — REQ 21 / ceiling 25; AC 20 / ceiling 25 |
| Artifact set | spec.md · plan.md · acceptance.md · design.md · research.md · progress.md |
| Requirements | 21 (REQ-JEVO-001 … REQ-JEVO-021) |
| Acceptance criteria | 20 (AC-JEVO-001 … AC-JEVO-020) |
| Predecessor | `SPEC-JEV-CORE-001` |
| Successor | `SPEC-JEV-CONSUMERS-001` |
| Operator decision encoded | Q1 resolved — wizard question is init-only (`InitQuestions`, five); consequence and escape hatch are REQ-JEVO-006 / REQ-JEVO-007 |
| Status transition | (none) → draft |

Open questions carried to the Implementation Kickoff Approval gate: Q3 (labelled-set size per consumer), R1 (which surface carries the `moai web` pointer).

R1 resolved at implementation time: the **wizard question `Description`** carries the pointer, in all four locales. It reaches the user at the moment of the decision, and it is one surface to keep translated rather than two.

## §E.2 Run-phase Evidence

Run-phase baseline: the tree at `2ce0294dd` (`WT-jev-init-optin`), measured in this run. Every row below names the command that produced it.

### Milestones

| Milestone | Delivered |
|---|---|
| M2a — wizard question | `jev_enabled` joined `Page3Questions` (init-only), four-locale translations, confirm-answer capture |
| M2b — web section | Jev sub-section on `/settings` inside the workflow panel: enable toggle + credential field + privacy note, four locales |
| M2c — shared seam | `settings.SetJevEnabled` + `settings.JevEnabledField` over the existing `ApplySchemaEdits` writer |
| M3a/b — harness | `internal/jevmeasure`: two-arm runner, constant-answer baseline, fitted thresholds, absence-with-control |
| M3c — gate demonstration | `TestGateDemonstration_ArtifactSatisfiesAC010` runs the gate end to end against an injected answerer |

### AC matrix

| AC | Status | Evidence (command → observation) |
|---|---|---|
| AC-JEVO-001 | PASS | `go test ./internal/web/ -run TestJevPanel_CarriesToggleAndCredential` → ok. Exactly one `data-section="jev"`; both `workflow.jev.enabled` and `jev_api_key` render inside it and nowhere else |
| AC-JEVO-002 | PASS | `go test ./internal/settings/ -run TestSetJevEnabled_AndConsoleEditAgree` → ok. The wizard wrapper and the console edit map produce byte-identical files; both call `ApplySchemaEdits` |
| AC-JEVO-003 | PASS | `go test ./internal/settings/ -run TestSetJevEnabled_SiblingFieldsByteIdentical` → ok. Exactly one line changed; the sibling comment and scalar survive (both asserted by positive control) |
| AC-JEVO-004 | PASS | `go test ./internal/cli/wizard/ -run TestJevQuestion_PrivacyAndPointerInEveryLocale` (4 sub-cases) + `go test ./internal/web/ -run TestJevI18nKeys_PresentInFourLocales` → ok |
| AC-JEVO-005 | PASS-WITH-DEBT | `go test ./internal/cli/wizard/ -run TestJevQuestion_TranslatedInFourLocales` → ok. All four locales covered, but through THREE map keys — see the blocker row below |
| AC-JEVO-006 | PASS | `go test ./internal/web/ -run TestConsoleRoutes_NoInitRoute` → ok; `git status --porcelain` shows `internal/web/app.go` unmodified. Positive control: the scan finds `"/settings"` |
| AC-JEVO-007 | PASS | `go test ./internal/cli/wizard/ -run TestInitQuestions_CarriesExactlyOneJevQuestion` → ok. Five questions, exactly one Jev id, `QuestionTypeConfirm` |
| AC-JEVO-008 | PASS | `go test ./internal/cli/wizard/ -run TestJevQuestion_AbsentFromDefaultAndReconfigure` → ok. Neither set carries the id; the reconfigure membership is pinned as an explicit 12-id sequence |
| AC-JEVO-009 | PASS | `go test ./internal/cli/wizard/ -run TestJevQuestion_PrivacyAndPointerInEveryLocale` → ok. `moai web` appears verbatim in all four locales |
| AC-JEVO-010 | PASS (shape) | `go test ./internal/jevmeasure/ -run TestGateDemonstration_ArtifactSatisfiesAC010` → ok. The artifact carries the pinned id, both arms with per-arm results, the delta, and the constant baseline. **The figures are fixtures** — see Gaps |
| AC-JEVO-011 | PASS | `go test ./internal/jevmeasure/ -run TestReport_RenderCitesPinnedModelAndCarriesNoAlias` → ok. Zero `jev-latest` in the rendered artifact, with an in-test positive control proving the search fires |
| AC-JEVO-012 | PASS (vacuous-by-design) | `go test ./internal/jevmeasure/ -run TestNoConsumerCallPathShips` → ok. No consumer call path exists in `internal/**`; positive control confirms the walk read this package's own source. No consumer was measured-and-withheld, because none exists yet |
| AC-JEVO-013 | PASS | `go test ./internal/jevmeasure/ -run TestThreshold_RejectsVendorCopyAndCrossShapeTransfer` → ok. Both vendor figures refused without provenance; Noul→Choice and cross-consumer transfer refused |
| AC-JEVO-014 | PASS | `go test ./internal/jevmeasure/ -run TestReport_AbsenceRequiresPositiveControl` → ok. `Validate()` refuses an absence with no control, and refuses a control with no observation |
| AC-JEVO-015 | PASS | `git status --porcelain \| grep -E "internal/(kanban\|epic)/"` → zero rows. Positive control: `grep -rl "type Card" internal/kanban/` → 3 files, so the filter addresses a real surface |
| AC-JEVO-016 | PASS | `go test ./internal/jevmeasure/ -run TestBuildRequest_RejectsComputationQuestions` → ok. Six computation shapes refused; positive control shows a judgment question over computed fields passes |
| AC-JEVO-017 | PASS | `go test ./internal/jevmeasure/ -run TestBuildRequest_ChoiceNeedsNoMatch` → ok. Mutant `if false` on the guard → test FAILED (measured) |
| AC-JEVO-018 | PASS | `go test ./internal/jevmeasure/ -run TestBuildRequest_StateCarriesOnlyReadFields` → ok. Mutant `if false` on the guard → test FAILED (measured) |
| AC-JEVO-019 | PASS | `go test ./internal/jevmeasure/ -run TestBuildRequest_StateIsDataNotInstruction` → ok. An injected imperative round-trips as a JSON value and never enters a question text |
| AC-JEVO-020 | PASS | `go test ./internal/jevmeasure/ -run 'TestReadNoul_BothPolaritiesRead\|TestNoComplementDerivationInSource'` → ok. Both polarities read; source scan finds no `1 - …Probability`, with a matcher positive control |

### Mutation checks (the tests were measured, not assumed)

Four mutants were introduced and reverted. The first one SURVIVED and exposed a vacuous test, which was then rewritten:

| Mutant | Result |
|---|---|
| `Verdict`: baseline comparison → `if false` | **SURVIVED** — `TestVerdict_WithholdsWhenBaselineNotBeaten` never reached the gate (the injected-source guard returns first). Test rewritten to construct a `SourceLive` report; re-measured below |
| `Verdict`: `<=` → `<` (boundary flip) | killed — `TestVerdict_WithholdsWhenBaselineNotBeaten/equal_to_the_constant` + `TestVerdict_EqualAccuracyDoesNotShip` FAILED |
| `Run`: unavailable answers counted correct | killed — `TestRun_UnavailableAnswersAreRecordedNotCountedCorrect` FAILED |
| `BuildRequest`: no-match + unread-state guards → `if false` | killed — `TestBuildRequest_ChoiceNeedsNoMatch` + `TestBuildRequest_StateCarriesOnlyReadFields` FAILED |

### Verification commands

| Command | Result |
|---|---|
| `go test ./internal/settings/... ./internal/jevmeasure/... ./internal/config/...` | ok (7 packages) |
| `go test ./internal/web/ ./internal/jev/... ./internal/jevcred/...` | ok (3 packages) |
| `go test ./internal/cli/wizard/` | ok |
| `go test ./internal/cli/ -run TestApplyJevFromWizard` | ok |
| `go test ./internal/cli/... -timeout 22m` | exit 0; every package line `ok` (filtered with `grep -vE "^ok"`, which emitted nothing but the trailing marker) |
| `golangci-lint run ./internal/jevmeasure/... ./internal/settings/... ./internal/web/...` | `0 issues.` |
| `golangci-lint run ./internal/cli/...` | `0 issues.` |
| `go vet ./internal/cli/ ./internal/cli/wizard/ ./internal/settings/ ./internal/jevmeasure/` | clean |
| `make build` | succeeded (catalog regenerated, binary linked) |

### Coverage

| Package | Coverage |
|---|---|
| `internal/jevmeasure` (new) | 96.1% |
| `internal/settings` | 90.6% |
| `internal/cli/wizard` | 94.0% |
| `internal/cli` — `applyJevFromWizard` | 100.0% (per-function) |
| `internal/web` | 67.8% (package-wide, pre-existing level) |

### Implementation decisions taken at run time

1. **The Jev section is a sub-section of the workflow panel, not a console tab.** A tab is coupled by `TestDocsTabContract` to eight documentation surfaces (four README locales, four docs-site locales), and README is outside run-phase ownership. The sub-section satisfies REQ-JEVO-001 without reaching into documentation this SPEC has no mandate over.
2. **The wizard question carries its own group label** (`Judgment Capability`), so it renders on its own page. Appending it to `Agents & Autonomy` was measured to overflow the 80x40 test viewport and scroll the step indicator off the top — a privacy statement the user must read must not depend on the page fitting.
3. **`applyJevFromWizard` takes an explicit `wizardRan` flag.** The wizard's zero value and its declined answer are the same bool, so without the flag a non-interactive `moai init --force` would write `false` over a `true` the user set earlier.
4. **`Report.Source` is inferred from the concrete client type, never declared.** A caller-declared source would be the one field a hurried change could set to `live` to make a stub's numbers shippable.

### Tests modified (each required by a SPEC decision, none weakened)

`TestInitQuestions_QuietSet`, `TestPage3QuestionsStructure`, `TestTotalVisibleQuestions_Page3AlwaysCounted`, `TestInitPages_Membership`, `TestStepperTotal_DynamicDenominator`, `TestInitStepper_Denominator4`, `TestInitRegroup_TwoPages`, `TestGroupLabel_NotRendered`, `TestInitRegroup_SecondGroupGolden`, `TestProfileWizardStepper_SameFormatAsInit` — all pinned the four-question init set that REQ-JEVO-005 takes to five. Five goldens regenerated; the only byte delta in each is the stepper denominator (`1 / 4` → `1 / 5`, `3 / 4` → `3 / 5`), verified by reading the diff.

`TestSchemaParity_EditableFieldsHaveRenderHome` — the new field was registered as having a render home (a dedicated component, like `codexAuthBlock`), NOT added to the exemption map. The field is rendered and editable; an exemption would have been false.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-20
run_commit_sha: 132324cbb
run_status: complete-with-gaps
ac_pass_count: 19
ac_fail_count: 0
ac_pass_with_debt_count: 1          # AC-JEVO-005, see blocker row
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a           # no push in run phase
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0 # golangci-lint: 0 issues on every changed package
cross_platform_build:
  darwin: pass                      # make build succeeded on darwin/arm64
  windows: unmeasured               # CI owns the matrix verdict
  linux: unmeasured
total_run_phase_files: 30
m1_to_mN_commit_strategy: single commit covering M2a/M2b/M2c/M3a/M3b/M3c
```

### Gaps (explicitly NOT observed)

1. **No live measurement was taken, and no accuracy figure in this repository is a measurement.** There is no TypeSafe credential in this environment and no test contacts `api.typesafe.ai`. Every figure the harness produced came from an injected answerer. This is enforced rather than merely stated: `Report.Source` is inferred from the concrete client type and `Verdict()` returns `VerdictWithhold` for any non-live source, so a fixture's numbers cannot become a shipping decision. **AC-JEVO-010 is marked PASS on the artifact's SHAPE only** — the gate has been demonstrated, not exercised against the vendor.
2. **AC-JEVO-012 is vacuously satisfiable.** No consumer exists, so "a consumer that failed its baseline is absent from the build" could not be observed happening. What was observed is that no consumer call path ships.
3. **The full test suite was not run locally** (CLAUDE.local.md §6 — parallel lanes running `go test ./...` drove machine load to 413). The full-suite verdict is CI's, on the pushed head. The `internal/cli/...` tree WAS run to completion (exit 0, all `ok`); everything outside the affected packages was not.
   - Attribution caveat on that run: it was started before two comment-only edits to `internal/cli/wizard/questions.go` landed. `go test ./internal/cli/wizard/` was re-run afterwards and returned `ok`, so the post-edit tree is covered for the package that changed; the rest of the `internal/cli` tree was not re-measured after those comment edits.
4. **Cross-platform build is darwin-only here.** Windows and Linux are unmeasured; CI owns that matrix.
5. **`TestTodoCommitsStillRefresh/held_WAL_writer` failed once and passed on two subsequent runs** of the same unchanged tree. It touches a WAL-lock contention path unrelated to this change. Recorded as an observed flake, not diagnosed.

### Blocker report — a requirement the tree contradicts

| Item | AC-JEVO-005 — "an entry for the Jev question id exists under each of the four locale keys" |
|---|---|
| What the tree does | `internal/cli/wizard/translations.go` has THREE locale keys: `ko`, `ja`, `zh`. `GetLocalizedQuestion` returns the base `Question` unchanged for `"en"` and `""` **before** it consults the map (`translations.go:317-321`), so an `en` entry is data no code path can read |
| What was implemented | All four locales ARE covered — en by the base `Question` value, the other three by map entries. The test reads all four through `GetLocalizedQuestion`, which is the surface the wizard actually renders from, and additionally asserts each non-en title is not byte-identical to the English base (so a missing entry cannot pass by silent fallback) |
| Why not the literal reading | Adding an `"en"` key would satisfy the AC's wording with dead data, added solely to be counted. That is the shape of an unobserved claim: the artifact would assert a coverage mechanism that does not run |
| Decision owner | Not mine. The AC wording, not the tree, is what needs amending — routed here rather than resolved unilaterally |

### Residual risk

- The harness's threshold fit is the median observed probability. It is deliberately simple because the rejected task measured the alternative (a flat sweep across 0.30-0.80, with the two better-looking cells resting on 5 and 4 samples). A consumer author could still read the fitted number as more precise than its sample size supports; `Provenance` states the caveat in the artifact, but nothing enforces that a reader reads it.
- `labelOf` maps a Noul onto `Labels[0]`/`Labels[1]` by position. A consumer that declares its labels in the other order would score itself exactly inverted, and the inversion would look like a very bad model rather than a wiring error. The convention is documented; it is not mechanically checked.
- The computation-verb guard is a substring matcher over English question text. It will not catch a computation phrased in a way the list does not cover, and it will reject an innocent question containing one of those words. The first direction is the one that ships a defect.

## §E.4 Sync-phase Audit-Ready Signal

Sync-phase baseline: the tree at `904c97ed5` (`WT-jev-init-optin`), measured in this run, inside the worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1020`. Every row below names the command that produced it and what that command printed. No figure is carried over from the run phase except where the row says so explicitly.

### What this sync commit changes

| File | Change |
|---|---|
| `CHANGELOG.md` | One entry at the head of `[Unreleased] → Added`, covering the two opt-in entrances, the shared writer, the credential handling, and the measurement gate — including the explicit statement that no live measurement was taken and that no accuracy figure in this release measures anything |
| `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md` | Frontmatter only: `status: in-progress → completed`, `updated: 2026-09-20` (already the sync date) |
| `.moai/specs/SPEC-JEV-OPTIN-MEASURE-001/progress.md` | This section |

No production code, no test, and no template file is touched by the sync commit.

### Status transition

`in-progress → implemented → completed`, landing on this single sync commit (3-phase close).

**Deviation, stated rather than worked around:** this SPEC's artifact set carries a frontmatter block on `spec.md` ONLY. `plan.md`, `acceptance.md`, and `progress.md` open with an H1 and have no frontmatter — verified by `head -5` on each, and the predecessor `SPEC-JEV-CORE-001` has the identical shape (`head -1` across its six artifacts: five H1s, one `---`). The transition therefore lands on the one artifact that has a frontmatter block. Adding a block to `plan.md` or `acceptance.md` would be a body edit, which the sync phase does not own; adding one to `progress.md` alone would make the set inconsistent with itself and with the predecessor.

### Evidence

| Claim | Command | Observed output |
|---|---|---|
| Build green after the sync edits | `go build ./...` | no output; `build_exit=0` |
| Affected packages green, uncached | `go test -count=1 -timeout 30m ./internal/cli/wizard/... ./internal/web/... ./internal/settings/... ./internal/jevmeasure/...` | `exit=0`; six `ok` lines — `internal/cli/wizard 4.077s`, `internal/web 22.573s`, `internal/settings 0.561s`, `internal/settings/agentfm 0.352s`, `internal/settings/yamlpatch 0.353s`, `internal/jevmeasure 0.461s` |
| No duplicate CHANGELOG entry existed (B12 pre-emission grep) | `grep -c 'SPEC-JEV-OPTIN-MEASURE-001' CHANGELOG.md` | `0` — emission proceeded |
| AC count (B12 AC-count match) | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md \| sort -u \| wc -l` | `20` — non-zero, and equal to the AC matrix's 20 rows in §E.2 |
| Every path named in the CHANGELOG entry resolves (B12 path verification) | `ls -d internal/cli/wizard/questions.go internal/jevmeasure internal/settings/jev.go .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md` | all four listed, no `No such file` |
| Symbols named in the entry exist | `grep -rlo 'func SetJevEnabled\|JevEnabledField' internal/settings/*.go`; `grep -rl 'TestDocsTabContract' internal/web/*_test.go`; `grep -rl 'func ApplySchemaEdits' internal/settings/*.go` | `internal/settings/jev.go` (+ its test); `internal/web/docs_tab_contract_test.go`; `internal/settings/sectionapply.go` |

### README and docs-site: the decision NOT to edit, and its evidence

**README (4 locales) — unchanged.** Two measurements.

1. `grep -rniE '\bjev\b|typesafe' README.md README.ko.md README.ja.md README.zh.md docs-site/content | wc -l` → `0`. Positive control on the same corpus: `grep -rl "moai web" README.ko.md docs-site/content | wc -l` → `41`, so the search apparatus fires on these files. Nothing in the README corpus says anything about Jev that this SPEC could have made false.
2. The one README claim this SPEC could plausibly have invalidated is the settings-tab enumeration at `README.ko.md:445` (`Identity·Language·LLM·GLM Settings·Workflow·Git & Worktree·Audit·Codex·Agents·Report·MCP·Cross-Session·Feedback·Quality Gate`). The Jev surface was implemented as a **sub-section inside the Workflow panel**, not a tab (§E.2 decision 1), so the tab list is still exactly right. Had a tab been added, `TestDocsTabContract` would have required all eight documentation surfaces in the same change.

The README also does not enumerate the wizard's question count or question list — `grep -n "마법사\|질문" README.ko.md` returns generic prose at line 290 and an FAQ heading, nothing that counts questions. So the init set moving from four to five makes no README sentence wrong.

**docs-site (4 locales) — unchanged.** The same zero-hit grep above covers `docs-site/content` entirely. One page does enumerate wizard questions — `docs-site/content/<locale>/getting-started/init-wizard.md` — and it IS wrong; the measurement is that it was **already** wrong before this SPEC, for a different cause:

- The page's structure table (`docs-site/content/ko/getting-started/init-wizard.md:44`) states `Page 3 — 품질 및 워크플로우 | LSP 통합, 품질 게이트 강제, 프로젝트 모드, 디자인 워크플로우, Claude Design 연동`.
- Reading `internal/cli/wizard/questions.go` at `c79afb760` — the commit BEFORE this SPEC's first implementation commit — and enumerating its `ID:` fields lists exactly: `conversation_language`, `user_name`, `project_name`, `model_policy`, `report_format`, `git_mode`, `git_provider`, `gitlab_instance_url`, `github_username`, `github_token`, `gitlab_username`, `gitlab_token`, `agent_wiring`, `autonomy_tier`. **None of the five questions the page's Page-3 row names exists in that tree.** They were retired by `SPEC-INIT-QUIET-WIZARD-001` REQ-IQW-002, as the `Page3Questions` doc comment in `questions.go` records.

So the page's divergence is pre-existing and not caused by `jev_enabled`. Repairing it means re-deriving the whole page from the tree across four locales — a different SPEC's scope, and one whose correctness depends on facts this SPEC did not measure. Editing only a Jev line into an already-false table would make the page read as maintained while staying wrong. **Recommended follow-up card: re-derive `getting-started/init-wizard.md` (ko canonical → en/ja/zh) from the live `InitQuestions` set.** This is a recommendation, not a claim that anything else on that page was checked — see Gaps 4.

### Gaps (explicitly NOT observed)

1. **The measurement was NOT run, and this sync phase produced no accuracy figure, confidence value, threshold result, or sample count.** The live-measurement path requires a call to the vendor; there is no TypeSafe credential in this tree and nothing in this sync phase contacted `api.typesafe.ai`. No number appearing in the CHANGELOG entry, in this section, or anywhere in this commit is a measurement of model accuracy. What would close this: a run of the `internal/jevmeasure` harness against a labelled set with a real credential, under the pinned model id, whose `Report.Source` is `SourceLive` and whose rendered artifact is committed as evidence.
2. **AC-JEVO-005 remains unresolved, and its claimed resolution has no on-disk record.** Recorded as required:
   - **The AC**: `AC-JEVO-005` — "Given the `translations` map, When it is enumerated, Then an entry for the Jev question id exists under each of the four locale keys."
   - **The claimed operator judgement**: that during the run phase the operator judged this AC to be carried over to the sync phase for disposition.
   - **The fact**: **no on-disk record of that judgement exists.** Searching this SPEC's artifact set finds only the run phase's own blocker report (§E.2 "Blocker report — a requirement the tree contradicts"), which states the opposite posture — "Decision owner: Not mine … routed here rather than resolved unilaterally". A judgement asserted in conversation and absent from disk is indistinguishable from an unanswered blocker, and this sync phase treats it as unanswered. **It is NOT resolved here.**
   - **What would make it verifiable**: the operator's decision written into the SPEC artifact set — either (a) an amendment to the AC's wording in `acceptance.md` accepting coverage via `GetLocalizedQuestion` rather than via a literal four-key map (a plan-phase edit owned by `manager-spec`, not by sync), or (b) an explicit recorded acceptance of the debt with its reason. Either lands as a commit; the commit is the record.
3. **The full test suite was not run.** Per `CLAUDE.local.md` §6 this machine does not run the whole-module test command locally; the four affected package trees were run uncached and are green. Cross-platform (windows, linux) is unmeasured here. Both verdicts belong to CI on the pushed head, and nothing has been pushed (Gap 5).
4. **The docs-site `init-wizard.md` page was inspected only on the single axis above.** The Page-3 row was measured against the pre-SPEC tree; the Page-1 and Page-2 rows, the per-step prose, the non-interactive-flag block, and the ja/en/zh translations of the same page were NOT checked against the tree. The follow-up recommendation must not be read as a claim that the rest of that page is correct, or that the Page-3 row is its only defect.
5. **Nothing was pushed, no PR was opened, and no merge was performed.** The branch `WT-jev-init-optin` is local; CI has rendered no verdict on this tree.
6. **`SPEC-JEV-CORE-001`, the predecessor that landed on this same branch, is still `status: in-progress`** (reading `^status:` from its `spec.md`, line 5). Its sync phase is outside this card's scope and was not performed here; observed and reported rather than silently closed.
7. **One tool call was REFUSED rather than executed** during this sync phase: a compound command combining a heredoc body and an inline script was refused by the worktree-isolation guard, which could not statically verify it stayed inside this worktree. It was re-issued as separate plain steps and ran to completion; no measurement was replaced by inference as a result. Recorded per the refused-tool-degradation rule.

### Residual risk

- The CHANGELOG entry states that the settings-tab list is unchanged. That is true of the tab list, but the Workflow panel's *content* is now larger by a sub-section, and no documentation surface describes that panel's contents at this granularity — so a reader looking for the Jev switch finds it only by opening the console. The init question names `moai web`; nothing in the published documentation does.
- Leaving `init-wizard.md` untouched keeps a page that is wrong about the wizard, and this sync phase has now recorded that it is wrong. The record makes the defect findable; it does not make the page correct, and a reader who never reaches this file still meets the stale page.
- The `sync_commit_sha` below is a placeholder at commit time — a commit cannot cite its own hash — and is backfilled in a following `chore:` commit. Between those two commits the field reads `pending-backfill-sync`, which is the intended transient state, not a missing value.

```yaml
sync_complete_at: 2026-09-20
sync_commit_sha: 52c460d30
sync_status: complete-with-gaps
changelog_entry_position: "[Unreleased] -> Added, first bullet"
b12_self_test_a_pre_emission_grep: "grep -c 'SPEC-JEV-OPTIN-MEASURE-001' CHANGELOG.md -> 0 (no duplicate; emission proceeded)"
b12_self_test_b_ac_count_match: "20 distinct AC ids in acceptance.md; E.2 AC matrix carries 20 rows"
b12_self_test_c_path_verification: "every path named in the entry resolved via ls -d; zero misses"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "n/a - no frontmatter block in this artifact set"
  acceptance_md: "n/a - no frontmatter block in this artifact set"
  progress_md: "n/a - no frontmatter block in this artifact set"
readme_touched: false
docs_site_touched: false
measurement_executed: false
ac_jevo_005_resolved: false
pushed: false
pr_opened: false
```
