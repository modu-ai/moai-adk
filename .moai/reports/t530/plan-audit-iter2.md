# SPEC Review Report: SPEC-DOCS-TABCOUNT-DRIFT-001 (card t530)

Iteration: 2/3
Verdict: **FAIL**
Overall Score: **0.775** (Tier M threshold 0.80)
Tree: `.claude/worktrees/t530`, branch `WT-web-tab-docs`, HEAD `28c1ea062`, base `1d150a27d`
Auditor note: reasoning context from the SPEC author was not consulted; the spawn prompt's framing was treated as claims to test, not as findings to confirm (M1 Context Isolation). All `mcp__moai__*` calls carried `project_root=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t530`.

No score regression (iter1 0.66 → iter2 0.775), so no STOP signal is raised. One iteration remains.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `spec.md:108,115,123,129,134,146,152,157,162` give REQ-TCD-001..009, sequential, no gaps, no duplicates, consistent zero-padding.
- **[PASS] MP-2 GEARS format compliance** (judged against the **requirement layer**, `REQ-XXX` in `spec.md`; the `Given/When/Then` bodies in `acceptance.md` are the verification layer and were graded under Group 4, not here) — 001/002/006 Ubiquitous; 003/004 `Where`; 005 Event-driven (`When the guard scans a document, it shall …`); 007 State-driven (`While any one of the ko / en / ja / zh copies … is changed`); 008/009 Unwanted (`shall not`). Two heading labels disagree with the body pattern (003 labelled Ubiquitous but written as `Where`; 009 labelled Ubiquitous but written as Unwanted) — cosmetic, not a pattern failure.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`spec.md:2-13`): `id`, `title`, `version: "0.2.0"`, `status: draft`, `created`/`updated` ISO, `author`, `priority: P2`, `phase: "v3.1.4 target"` (not a prohibited lifecycle-stage value), `module`, `lifecycle: spec-anchored`, `tags` comma-string. No rejected snake_case alias. `plan.md` / `acceptance.md` correctly carry no `status:` (statelessness on the status axis).
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to this repository's own README set and docs-site content, not to template-bound or multi-language tooling. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — references resolved: `SPEC-WEB-CODEX-PANEL-001` → `status: completed`, `SPEC-PRECOMMIT-GATE-SCOPE-001` → `status: completed`. Neither is retired/superseded/archived, so no reconciliation is owed. `mcp__moai__spec_audit` returned `modern_era_clean: 1` with one INFO `EraAutoDetected` finding only.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rc 'syscall'` over all four artifacts returns `0` for each. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-DOCS-TABCOUNT-DRIFT-001/` → no matches. `plan.md §C` records the llm-label question as resolved and deliberately omits the marker token; `research.md` does not exist (Tier M).

No must-pass criterion fails. The FAIL verdict rests on the aggregate score and on D1 (a release-blocking criterion that is red for reasons the card cannot fix).

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Two-layer table (`spec.md:97-102`) and the explicit self-corrections (`plan.md:46-50`, `plan.md:69-71`) are unusually clear. Deducted for D4 (`tab-count-sites.md:14` contradicts `spec.md:48/51` **and its own line 167**), D9 (`spec.md:48` "원천 두 개" vs `:51` "원천 1 + 미러 1"), D7 (`allowed 1` unit undefined), D8 (two dangling AC ids). |
| Completeness | 0.80 | 0.75-1.0 | All sections present; `§6` carries four `### Out of Scope — <topic>` H3 sub-headings each with specific `-` bullets (`spec.md:220-243`); the enumeration deliverable exists and reproduces. Deducted because `progress.md §E.1` Gaps and `spec.md §7` are each incomplete in a specific, checkable way — see D2 and D1. |
| Testability | 0.70 | 0.50-0.75 | Every AC carries a command plus expected output, and three genuine empty-sweep defences are present (`acceptance.md:142-143` asserts the absence of `no tests to run`; `:105` counts the pattern file's lines; `:228` turns an all-zero diff into FAIL). Deducted for D1 (wrong-reason red), D5, D6, D7 — four criteria admit a mutant that satisfies them while violating their requirement. |
| Traceability | 0.85 | 0.75-1.0 | The REQ↔AC map at `acceptance.md:55-58` was checked in both directions and holds: every REQ-TCD-001..009 has at least one AC or RG, and every AC names an existing REQ. `RG-TCD-002` declares its REQ absence explicitly rather than hiding it. Deducted for D8 (two citations to AC ids that do not exist) and because REQ-TCD-001's only enumeration-completeness teeth sit in a regression guard (D5). |

Aggregate (arithmetic mean) = 0.775; harmonic mean = 0.772. Both are below the Tier M threshold of 0.80.

---

## What the repair got right (stated so it is not re-litigated)

Three of the prompt's claims were tested and hold:

- **Set equality, both directions.** `grep -rnF -f count-literals.txt` over the 12 files returns exactly 20 lines, and their `file:line` set is identical to the A/B/C tables: `web.md:53` ×4, `moai-web-console.md:25` ×4, `:128` ×4, `README*:414` ×4, `README*:750` ×4. This is a real fix of iteration 1's D2 — the previous window regex produced 20 hits of a *different* set.
- **The literal set survives absorption.** Re-run against `develop`'s current versions of the same 12 files (`git grep -F -f … develop -- …`): still 20 hits, same set, with only README line numbers shifted 414→419 / 750→754. AC-TCD-011 deliberately asserts paths and not line numbers, so it is unaffected. The acceptance layer's independence from line drift is real, not merely asserted.
- **Code-side premises verified.** `consoleTabs()` yields 14 (`sed -n '30,90p' internal/web/schemaform.go | grep -c 'LabelKey:'` → `14`); `internal/web/assets/i18n.js:229,1096,1852,2608` carry `GLM Settings` / `GLM 설정` / `GLM設定` / `GLM设置`; `schemaform.go:39` `Baseline: "GLM Settings"` agrees with `i18n.js:229`. `plan.md §C`'s decision rests on a sound premise.

The AC set is also mutually reinforcing in a way single-AC mutant probing understates: several mutants that satisfy AC-005, AC-007 or AC-009 in isolation are caught by AC-TCD-006's V2/V3 mutations, which land in a different file class and a different locale. That is good design and is credited in the Testability score.

---

## Defects Found

**D1. AC-TCD-010 becomes wrong-reason red the moment the card absorbs `develop`** — `acceptance.md:222-243` — Severity: **critical** — Class: **blocking**

The criterion pins its left endpoint to the literal base `1d150a27d` and its pathspec to `README.md README.ko.md README.ja.md README.zh.md docs-site/content` — the whole content tree, not the 12 target files. Measured on this tree:

```
$ git rev-list --count 1d150a27d..develop                → 40
$ git diff --name-only 1d150a27d develop -- README.md README.ko.md README.ja.md README.zh.md docs-site/content | wc -l   → 155
```

Applying the criterion's own third pipeline to that list, four distinct relative paths violate the "exactly 4 times" rule and none of them belongs to this card:

```
1 advanced/hooks-reference.md
2 advanced/settings-json.md
2 cost-optimization/_index.md
2 workflow-commands/moai-run.md
```

This is not hypothetical fragility. `.claude/rules/local/gitflow-lane-protocol.md` §2/§8 and `CLAUDE.local.md` §4.1 make absorbing `develop` inside the card worktree and **re-measuring in the merged tree** a [HARD] precondition of integration, so the mandated procedure is what turns this criterion red. Its green path then runs through "someone fixes the unrelated files", which `.claude/rules/moai/development/verification-completeness.md` §2 disqualifies explicitly. The criterion's own rationale (`acceptance.md:242-243`) defends the pin against a *different* hazard — `origin/develop` advancing under a moving ref — and never addresses foreign churn entering a fixed-base range.

**Required fix**: (a) restrict the pathspec to the enumerated 12 files rather than `docs-site/content`; and (b) replace the literal base with `CARD_BASE=$(git merge-base develop HEAD)`, evaluated at read time and not pinned, per gitflow-lane-protocol §8 — after absorption the merge-base is the absorbed develop commit, so the range contains only this card's work. Keep the existing zero-changes-means-FAIL control group; it is correct and should be stated against the same recomputed base.

**D2. AC-TCD-002 admits a one-word mutant on the six word-form sites, and both layers are silent** — `acceptance.md:84-105`, `spec.md:95-102`, `plan.md:64` — Severity: **major** — Class: **blocking**

The SPEC claims the layers cover one another (`plan.md:64` "셋이 서로를 덮는다"; `spec.md:97-102`). For the 14 numeral sites that is true — N1's numeric sweep backstops a literal that stops matching. For the six word-form sites (A2 `The nine settings tabs`, A4 `设置九个标签页`, B4 `unfolds fourteen tabs`, C1 `fourteen tabs`, C2 `열네 개 탭`, C4 `十四个标签页`) there is no second layer at all, because N1 is `[0-9]+`-only. Measured, on the real line `docs-site/content/en/advanced/moai-web-console.md:128`:

| line | literal layer (AC-TCD-002) | guard layer N1 |
|---|---|---|
| `…unfolds fourteen tabs below it…` (original) | 1 | 0 |
| `…unfolds fourteen settings tabs below it…` (mutant) | **0** | **0** |

Inserting one word makes AC-TCD-002 green while a hand-written count survives — violating REQ-TCD-003 and REQ-TCD-008. The Korean case is more likely to occur by accident than by malice: `열네 개 탭` → `탭 열네 개` is an ordinary rewrite and produces the same result. This is the mutant probe of `verification-completeness.md` §2 landing on the class of site the SPEC itself identifies as the one previous cleanups missed (`spec.md:86-87`).

**Required fix**: add a per-site positive control for the six word-form sites — for each, assert the numeral word token (`nine`, `fourteen`, `열네`, `十四`, `九个`) is absent from that file, scoped to the file rather than the line so edits cannot move it; or close the word axis in the guard (see D3, which shows the stated obstacle is not there).

**D3. `plan.md §A.4`'s measurement does not support the decision not to open the word axis** — `plan.md:79-83`, `spec.md:256-259` — Severity: **major** — Class: **blocking**

`plan.md:79` justifies excluding word numerals from the guard by naming three false positives: ko `탭 두 곳`, en `two tabs`, zh `两个标签页`. All three are on line 149 of `advanced/moai-web-console.md`, and all four locales of that line contain the token `codex`:

```
$ sed -n '149p' docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md | grep -ci codex   → 1 each
```

The allowlist the plan itself adopts (`acceptance.md:191`, `plan.md:59-60`) exempts exactly "a line containing `codex` inside `advanced/moai-web-console.md`". So every false positive cited as the reason not to open the word axis is already exempt by the device already chosen. A word-numeral sweep over the 12 target files returns 8 hits: 6 enumerated word-form sites, 1 allowlisted codex line, and 1 trivially excludable artifact (`zh/advanced/moai-web-console.md:165`, matched only because `第三方` contains `三`). `spec.md:256-259` states the axis "열 수 없다 … 실측으로 확인했다" — a premise claim the measurement, taken on the guard's actual scan scope, does not support (`verification-claim-integrity.md` §1.1 surface 4: a recommendation premise carries the same evidence burden as a defect claim).

**Limit of this evidence, stated**: my sweep used a hand-built numeral-word class. It shows the *stated* false-positive obstacle is not established on the 12-file scope; it does not prove a production word-axis regex would stay false-positive-free across future edits. The decision may still be correct — but it must be re-argued from a measurement that accounts for the allowlist, or restated as a scope choice rather than a measured impossibility.

**Required fix**: either (a) re-measure the word axis with the allowlist applied and open it if the un-allowed false-positive count is zero, which would also close D2; or (b) keep the axis closed but rewrite `plan.md §A.4`'s last row and `spec.md §7`'s third bullet to state the real reason (locale numeral-class maintenance cost) and to disclose the D2 exposure.

**D4. The enumeration deliverable contradicts the SPEC and itself on the "three independent measurements" claim** — `.moai/reports/t530/tab-count-sites.md:14` vs `:167` and `spec.md:48-52` — Severity: **major** — Class: **blocking**

`spec.md:48-52` and `progress.md:9-10` deliberately downgrade the claim to "one source + one mirror + one agreement test", and `spec.md` HISTORY D8 records that repair as done. The deliverable the SPEC cites as its enumeration artifact still opens with `**14.** 서로 독립인 세 측정이 일치한다` at line 14, while its own reproduction block at line 167 carries the corrected wording. A reader who opens the artifact first reads the overstatement the SPEC exists to disown — and `spec.md:52` names precisely this ("독립성을 부풀려 적지 않는다 … 과장이야말로 이 카드가 다루는 결함이다"). This is the cross-layer revision sweep of `verification-completeness.md` §3: the revision did not end in the file it started in.

Secondary: `spec.md:48` itself says "원천 **두 개**" where `:51` says "원천 1 + 미러 1" — the passage written to stop overstatement is internally inconsistent (logged separately as D9).

**Required fix**: rewrite `tab-count-sites.md:14` to the `:167` wording, and reconcile `spec.md:48` with `:51`.

**D5. AC-TCD-011 is satisfied by a mutant that deletes an entire enumeration table** — `acceptance.md:245-258` — Severity: minor — Class: **optional**

The criterion extracts backticked paths from anywhere in `tab-count-sites.md` and asserts the unique count is 12. The B-group rows abbreviate their paths (`` `.../ko/advanced/moai-web-console.md` ``), so they contribute no matches; the four `moai-web-console.md` paths come from the D-group and §7-B tables instead. Measured on a mutant with the whole B-group table (8 enumerated sites) deleted:

```
unique paths: 12   (unchanged → AC-TCD-011 PASS)
| [ABCD]N rows: 24 (32 → 24 → RG-TCD-003 FAIL)
```

So REQ-TCD-001's completeness is actually enforced by a regression guard, not by the acceptance criterion named for it. The 12 also holds only because §7-A writes its glob rows with a literal `*`; expanding those rows to per-locale paths would push the count above 12 and fail the criterion spuriously.

**Required fix**: scope the extraction to the A/B/C/D table rows (e.g. lines beginning `| [ABCD][0-9]`) and, if the enumeration is to be asserted, assert row count and path count from the same restricted region.

**D6. AC-TCD-005's `swept 12 files` can be printed without reading anything** — `acceptance.md:156-164` — Severity: minor — Class: **optional**

The criterion greps the guard's own output for `swept 12 files`. A guard that emits `t.Logf("swept %d files", len(targetFiles))` against its declared literal slice satisfies it while opening one file or none — the exact confusion the criterion's own rationale (`:163-164`) says it exists to prevent. AC-TCD-006's V2/V3 mutations do backstop the underlying hazard, which is why this is optional rather than blocking.

**Required fix**: specify that the counter increments on a successful read, after the read, and that a read failure is `t.Fatal` (consistent with `acceptance.md:292`).

**D7. `allowed 1` has no defined unit, and bounds rules rather than the exemption surface** — `acceptance.md:184-193` vs `spec.md:255` — Severity: minor — Class: **optional**

`acceptance.md:190-193` asserts `allowed 1` and fails at 0 or ≥2; `spec.md:255` describes the same allowlist as "허용 항목이 **4줄**"; `tab-count-sites.md:125` calls it "1항목, 4로케일". An implementer can read "allowed" as rule-entries (1) or as exempted lines (1 today, because only `ja:149` is a numeric-sweep hit) — both give 1 on this tree, so the ambiguity is invisible until a future edit makes `ko:149` numeric and the count becomes 2, producing a spurious FAIL. Separately, `allowed 1` bounds the number of rules, not the exemption surface: `grep -ci codex` returns 4 per file across 4 files, so the single rule exempts up to 16 lines from the numeric sweep. `spec.md §7` acknowledges the residual honestly, but the criterion does not measure it.

**Required fix**: state the unit explicitly ("one allowlist **rule**"), and add a second assertion bounding the number of *lines* the rule actually exempted on this run.

**D8. Two citations name AC ids that do not exist** — `.moai/reports/t530/tab-count-sites.md:164`, `plan.md:97` — Severity: minor — Class: **optional**

`tab-count-sites.md:164` cites `AC-TCD-012`; `acceptance.md` defines AC-TCD-001..011 only. `plan.md:97` attributes the swept-file-count assertion to `AC-TCD-011`, but that is the path-completeness criterion — the swept count is `AC-TCD-005`. Both were introduced by the repair's renumbering.

**Required fix**: `tab-count-sites.md:164` → `AC-TCD-011`; `plan.md:97` → `AC-TCD-005`.

**D9. `spec.md §1.1` is internally inconsistent about its own measurement count** — `spec.md:48` vs `:51` — Severity: minor — Class: **optional** — **Required fix**: change `:48` "원천 두 개와 그 둘의 일치를 확인하는 테스트 하나" to match `:51` "원천 1 + 미러 1 + 둘의 일치 테스트 1".

**D10. REQ-TCD-005 hardcodes line numbers into a requirement** — `spec.md:139-142` — Severity: minor — Class: **optional**

The requirement names `README.md:422`, `README.md:418`, `README.md:698`, and `plugins.md:100`. The acceptance layer deliberately identifies the same sites by phrase (`acceptance.md:202-205`) and identifies the allowlist by content (`:191`), citing line movement as the reason. The requirement layer should follow the same discipline — more so because this card edits `README.md`, and because develop has already moved the README line numbers by five (`414→419`, `750→754`).

**Required fix**: restate REQ-TCD-005's negative cases by phrase, matching `acceptance.md:202-205`.

**D12. The plan-phase commit carries no `Authored-By-Agent:` trailer, so the `(none) → draft` transition is unattributable** — commit `c205eeeff` — Severity: minor — Class: **optional** — informational only

`moai spec lint` reports `INFO OwnershipTransitionUnmeasured` against `spec.md:1`: the expected owner for `(none) → draft` is `manager-spec`, but `c205eeeff1793157c7d99163998e286988f14a92` ("feat(SPEC-DOCS-TABCOUNT-DRIFT-001): plan-phase artifacts (Tier M, 4 artifacts) (t530)") carries no trailer, so the transition cannot be attributed. Per `spec-frontmatter-schema.md` § OwnershipTransitionRule this is "a statement about measurement state, not a violation", it is Info severity, and `--strict` never escalates it. It is recorded here only so the absence is on the record rather than unmentioned; the transition commit is already landed, so there is nothing to repair in this revision round.

**Required fix**: none for this card. Add the `Authored-By-Agent: manager-spec` trailer on future plan-phase commits.

**D11. RG-TCD-001 carries the same fixed-base exposure as D1** — `acceptance.md:267` — Severity: minor — Class: **optional**

`git diff --quiet 1d150a27d -- assets/images/` returns 0 today, and `git diff --name-only 1d150a27d develop -- assets/images` is empty, so there is no current exposure. It would become a false FAIL attributing another card's image change to t530 the moment develop touches that path and t530 absorbs it.

**Required fix**: use the recomputed `merge-base` endpoint from D1's fix here as well.

---

## Regression Check (iteration 1 defects)

Iteration 1's report enumerated its findings as D1-D11 within `.moai/reports/t530/plan-audit.md`. Judged on the current artifacts:

- **iter1 D1 (the window regex cannot see `The nine settings tabs`)** — **RESOLVED**. The acceptance layer no longer uses a window regex; `head -4 count-literals.txt` sweeps A-group and returns 4 hits including `docs-site/content/en/cli-reference/web.md:53`.
- **iter1 D2 (20 hits of the wrong set)** — **RESOLVED**. Set equality verified in both directions against the A/B/C tables; no enumerated site is missing and no non-enumerated line is included.
- **iter1 D5 (screenshot / hugo classified as acceptance criteria)** — **RESOLVED, and the reclassification is honest.** RG-TCD-001 guards a prohibition (REQ-TCD-009 is a `shall not`), which is a regression guard by construction. RG-TCD-002 is evaluated post-edit (`acceptance.md:275` "문서 수정이 끝난 트리에서"), so it is not a burden dodged. RG-TCD-003 is already-green because the plan phase produced the artifact it guards. No criterion appears to have been relabelled to escape a red-now burden.
- **iter1 D4 (parity AC treated an all-zero diff as a pass)** — **PARTIALLY RESOLVED**. The control group was added and is correct (`acceptance.md:228`), but the same repair introduced D1 above.

The remaining iteration-1 findings are not re-adjudicated here; this audit judged the artifacts on their own merits, as instructed, and D1-D4 above are all defects the repair itself introduced or left in the deliverable.

No defect has survived unchanged across all iterations, so no stagnation flag is raised.

---

## What I could not check

Stated rather than passed over silently:

- **The guard does not exist.** `internal/web/docs_tab_contract_test.go` is absent, so AC-TCD-004 through AC-TCD-009 were probed as specifications only. Their mutant analysis (D5, D6, D7) is analytic plus targeted grep; no red was executed against a real guard. The `no tests to run` RED-now cell was re-executed and reproduces (`ok … [no tests to run]`, exit 0).
- ~~`moai spec lint` did not complete within a 120s budget~~ — **gap closed after the run finished in the background.** Result: exactly one row names this SPEC, at INFO severity (`OwnershipTransitionUnmeasured`, see below); zero WARNING and zero ERROR rows. Project-wide the run exited 0 with `0 error(s), 3134 warning(s)`, none of them attributed to SPEC-DOCS-TABCOUNT-DRIFT-001. Notably the lint emitted **no `CoverageIncomplete` row for any REQ-TCD-\***, which independently corroborates the Traceability finding above — the lint's own coverage check also sees every requirement referenced by an acceptance criterion. `mcp__moai__spec_audit` (with `project_root` set to this worktree) returned `modern_era_clean: 1` with an INFO `EraAutoDetected` finding only.
- **RG-TCD-002 was not executed.** No `hugo` build was run; the warning-free claim is carried forward from the artifacts, not re-measured here.
- **D1's post-absorption failure was simulated, not executed.** I measured `1d150a27d..develop` rather than performing a merge, because a worktree-isolated audit must not mutate the card's tree. The 155-file and four-violating-path figures are therefore a faithful projection of what the mandated absorption produces, not an observation of a merged tree.
- **The word-axis sweep in D3 used a hand-built numeral class**, so it establishes that the stated obstacle is unproven, not that a production word-axis regex is clean.

---

## Recommendation

FAIL at 0.775 against a 0.80 threshold. Four blocking findings, each with a concrete and cheap fix. In priority order:

1. **Fix AC-TCD-010 (D1).** Restrict the pathspec to the 12 enumerated files and derive the left endpoint as `git merge-base develop HEAD` at read time. Apply the same endpoint to RG-TCD-001 (D11). This is the one finding that would otherwise make a release-blocking criterion red for foreign reasons after the repo's mandated integration step.
2. **Close the word-form gap (D2 + D3) together.** Re-measure the word axis with the existing `codex`-line allowlist applied; if the un-allowed false-positive count is zero, open the axis in the guard, which closes D2 and makes D3's rationale true. If the axis stays closed, restate `plan.md §A.4` and `spec.md §7` in terms of maintenance cost rather than a measured impossibility, and add per-file numeral-word absence controls for the six word-form sites.
3. **Sweep the deliverable (D4).** Rewrite `tab-count-sites.md:14` to match its own line 167 and `spec.md:51`; reconcile `spec.md:48`.
4. Optional findings D5-D11 are surfaced for the orchestrator's discretion. D5 and D7 are the two most worth taking, because each currently lets a criterion report a bound it does not actually measure; D8, D9 and D10 are one-line corrections that cost nothing to fold into the same revision.

Iteration 3 is available. A re-audit should be scoped to this enumerated defect delta plus a regression check over D1-D4, not a from-scratch full re-audit.
