# SPEC Review Report: SPEC-AGENT-MODEL-INHERIT-001
Iteration: 5/3 (partial scope, operator-approved 2026-09-26 in the lane window — confirming re-audit limited to the iter-4 defect delta V1/V2/V5 plus regressions introduced by `4cc702dcf..077c4f8f6`)
Verdict: PASS (PASS-WITH-DEBT — no blocking defect remains; optional items V3, V4, W1, W2, W3 carried)
Overall Score: 0.90 (Tier L PASS threshold 0.85; iter-1 0.76 → iter-2 0.82 → iter-3 0.83 → iter-4 0.86 → iter-5 0.90, rising — no STOP signal)

Card: t1246. Audited tree: `.claude/worktrees/t1246`, branch `WT-agent-model-inherit`, HEAD pinned
`077c4f8f62630e3b662feffc81bbbc61eab9c809` (read first; starts with the requested `077c4f8f6`).
Reasoning context ignored per M1 Context Isolation. The revision's commit message was treated as a claim; each fix was re-measured against code at this HEAD.
Scope: `git diff --stat 4cc702dcf..077c4f8f6` → 7 files (touch-set.sh/.txt, acceptance.md, design.md, plan.md, progress.md, research.md), 28+/14−. Items resolved in earlier iterations were not reopened; no diff hunk touches them.

## Must-Pass Results

Carried from iter-4 (spec.md is not in this revision's diff, so MP-1/2/3/5/6 inputs are unchanged):
- [PASS] MP-1 REQ number consistency — spec.md unchanged since iter-4 (REQ-AMI-001…025 contiguous).
- [PASS] MP-2 GEARS (requirement layer) — spec.md unchanged; the only AC edit (acceptance.md AC-AMI-014) is verification layer and stays Given-When-Then.
- [PASS] MP-3 YAML frontmatter — spec.md unchanged (`version: "0.5.0"`, 12 fields).
- [N/A] MP-4 language neutrality — no programming-language tooling named.
- [PASS] MP-5 D7 — no new SPEC reference introduced except SPEC-UPDATE-MIRROR-HEAL-001 (design.md:L33), which is cited to *keep* its REQ-UMH-003 contract, not to reverse it — no reconciliation needed.
- [PASS] MP-6 D8 — `syscall` absent from the diff.
- [PASS] MP-7 clarification gate — no `[NEEDS CLARIFICATION` in the diff; plan.md/research.md carried none at iter-4.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | 0.75-1.0 | D14(b) (design.md:L33) now names one mechanism — re-evaluate `verr == nil && packageVersion == projectVersion && !forceUpdate` — and states the signature does not change; plan.md:L62-L65 matches. D10 now states invalid-value behaviour. Nits W2/W3. |
| Completeness | 0.87 | 0.75 | Touch set 272, `i18n_governance_test.go` added; §G gains the `IsValidProfile` row (plan.md:L157). Gaps: `internal/cli/init_test.go` (W1), app.js (V3, carried). |
| Testability | 0.89 | 0.75-1.0 | AC-AMI-014 (acceptance.md:L28) now names the `confirmViaPreview` seam for the cancel fixture. V4 carried. |
| Traceability | 0.95 | 1.0 minus nit | Unchanged mapping (25 REQ ↔ 25 AC). |

Aggregate (harmonic mean of 0.90 / 0.87 / 0.89 / 0.95) ≈ 0.90.

## Regression Check (iteration-4 defects)

- **V1 — RESOLVED.** Verified against code:
  - `update_template_sync.go:758` `func runTemplateSyncWithProgress(cmd *cobra.Command) (skipped bool, err error)`; only two `return true, nil` in the function — :776 (version match, guarded by :771-773 `packageVersion := version.GetVersion()` / `projectVersion, verr := plan.GetProjectConfigVersion(projectRoot)` / `if verr == nil && packageVersion == projectVersion && !forceUpdate {`) and :799 (user cancel after `confirmViaPreview` :791). No third skip reason exists, so re-evaluating the :773 predicate partitions the two exactly; `--yes` never reaches the cancel branch. design.md:L33 quotes the predicate and line range correctly.
  - Guards: `update_mirror_heal_test.go:337-339` asserts the literal signature; :341 looks for `packageVersion == projectVersion && !forceUpdate` in `update_template_sync.go` and requires `return true, nil` before `runTemplateSyncWithReporter(`. The SPEC now keeps the signature and places the re-evaluation in `update.go` (host (b), inside the `syncSkipped` branch at update.go:524), so both guards stay green by design. `update_codex_wiring_test.go:85` and `update_mirror_heal_wiring_test.go:30` both locate `if syncSkipped {` — the SPEC does not rename it.
  - Seam: `confirmViaPreview` (:849) is a plain function today (non-TTY → error); D14 and plan M1 make it a package-level function variable; `grep confirmViaPreview --include=*_test.go` finds no source-text guard on it (iter-4 measurement, unchanged). AC-AMI-014 names the seam.
  - The 8 call sites of `runTemplateSyncWithProgress` in cli tests (iter-4 V1 list) now need no change.
- **V2 — RESOLVED.** `i18n_governance_test.go:437` `if len(i18nEnExemptPrefixes) == 0 {`, :438 `t.Error("en-exempt prefix registry is empty")`, :439 `}`; :440-444 the per-member justification loop — exactly as design.md §F row (L122) and plan.md M2 (L78-L79) cite. The registry's only member is `agentdesc.` (`i18n_untranslated_allowlist_test.go:264-269`). `i18nMatchesExemptPrefix` (:449-456) iterates the registry harmlessly when empty. The file is now in the touch set (touch-set.sh adds it explicitly; touch-set.txt line 259).
- **V5 — RESOLVED, with two new optional nits.** (a) `IsValidProfile` §G row (plan.md:L157): non-test consumers re-measured — `cli/update.go:55` (validateUpdateFlags), `:657` (applyUpdateProfile), `cli/init.go:363` (--model-policy), `:370` (--profile); `template/profile_matrix.go:77` is a comment only. Exact match. (b) D10 (design.md:L29) now states "its value is no longer validated (any value, valid or not, gets the same warning)" — consistent with AC-AMI-016 ("each exits 0"). But see W1. (c) M4 now names `wizard/translations.go:89-91` (ko), `:179-181` (ja), `:269-271` (zh); the en text lives in `wizard/questions.go:103-125`, already in M4. See W3.
- **V3, V4 — carried unchanged (optional, not in this revision's scope).**

## Diff regressions (this revision only)

- design.md D10/D14, plan.md M1/M2/M4/§G, acceptance.md AC-AMI-014, research.md §I/§J, progress.md run-gate line, touch-set.{sh,txt}: all edits consistent with each other (272 paths stated in plan.md:L11, research.md:L257, progress.md:L8; 29 `internal/web` in research.md:L262).
- New findings W1-W3 below; none blocking.

## Gates reproduced now

- `sh .moai/reports/t1246/touch-set.sh > …/ts.txt` → exit 0, 272 lines; `diff` vs committed `touch-set.txt` → identical; `LC_ALL=C sort -c touch-set.txt` → exit 0; `cut -d/ -f1-2 | LC_ALL=C sort | uniq -c` → 22 .claude/agents, 2 .claude/commands, 1 .claude/hooks, 12 .claude/rules, 8 .claude/skills, 5 .claude/workflows, 1 .moai/docs, 2 .moai/project, 1 CHANGELOG.md, 62 docs-site/content, 38 internal/cli, 12 internal/config, 10 internal/harness, 6 internal/hook, 5 internal/settings, 2 internal/spec, 54 internal/template, 29 internal/web — equal to research.md:L259-L262.
- REQ-AMI-001: `git merge-base --is-ancestor WT-rules-diet develop` → exit 1 (develop `b4f798dcc`, WT-rules-diet `3a48485af`) — gate still closed, as recorded; run may not start yet.
- REQ-AMI-002: `git merge-base HEAD WT-role-naming-docs` → `e62c3e183`; `git diff --name-only e62c3e183 WT-role-naming-docs` → 32 files, all under `.moai/reports/t1257/` and `.moai/specs/SPEC-ROLE-NAMING-DOCS-001/`; the touch set's only `.moai/` entries are `.moai/docs` (1) and `.moai/project` (2), so the C-collated intersection is 0 (tip `024b95f77`).

## Defects Found (structured defect-list)

V3. DEAD-AGENTFM-JS — carried from iter-4 unchanged — Severity: minor — Class: optional — Required fix: add `internal/web/assets/app.js` to M2 and the touch set, or record it as accepted residue.

V4. AC-006-FIXTURE-CONTENT — carried from iter-4 unchanged (acceptance.md AC-AMI-006) — Severity: minor — Class: optional — Required fix: "no LR-03/LR-12/LR-13 finding and no error", or state the fixture is otherwise lint-clean.

W1. INIT-TEST-INVALID-VALUE — design.md:L29 (D10, new clause) vs `internal/cli/init_test.go:419` (`TestValidateInitFlags_InvalidProfile`, asserts an error containing `invalid --profile`, :431) and :451 (`TestValidateInitFlags_ModelPolicyVocabulary`, invalid half asserts `invalid --model-policy`, :476). D10's new "value is no longer validated" makes both assertions fail at M4. `init_test.go` is not in the touch set and not in plan §G's test-file table (the §G rule is keyed on deleted *symbols*; these tests pin behaviour, not `IsValidProfile`). The behaviour change is explicit, so the adaptation is determined and M4 would catch it mechanically; t1257 does not touch the file, so the overlap gate result is unaffected. — Severity: minor — Class: optional — Required fix: add `internal/cli/init_test.go` to the touch set and name the two tests' removal/adaptation in M4 (by the M4 author at run entry is sufficient).

W2. PREDICATE-TEXT-DRIFT — research.md:L314 quotes the predicate as `packageVersion == projectVersion && !forceUpdate`, dropping the `verr == nil` conjunct that design.md:L33 (correctly) carries. Design is the implementation authority, so no behaviour risk. — Severity: minor — Class: optional — Required fix: add `verr == nil &&` to research.md:L314, or leave as is.

W3. TRANSLATION-RANGE-NARROW — plan.md:L93 cites `wizard/translations.go:89-91`, `:179-181`, `:269-271`; each `model_policy` map entry runs to its closing brace (ko :89-97, with the `Options` sub-block at :92-96; ja/zh likewise). Removing only the cited lines would not compile, so an implementer will remove the whole entry — no correctness risk. — Severity: minor — Class: optional — Required fix: cite the full entries (e.g. ko :89-97) or say "the `model_policy` entry".

## Recommendation

PASS. Rationale per must-pass: MP-1/2/3 inputs (spec.md) are unchanged since iter-4's PASS; MP-5 introduces only a keep-the-contract citation; MP-6/MP-7 clean. Both iter-4 blocking defects are resolved against code: V1 keeps `(skipped bool, err error)` and separates version match from cancel by re-evaluating the exact :773 predicate, which is a complete partition because the function has only those two `(true, nil)` returns; V2 names the :437-439 drop and keeps :440-444, with the file added to the touch set. Aggregate 0.90 ≥ 0.85.

Debt for the orchestrator to route (all optional, M6): W1 is the one worth doing at run entry (touch-set completeness for `init_test.go`); V3, V4, W2, W3 are wording/residue.

Reminder, not a defect: REQ-AMI-001 remains closed (exit 1 above). This PASS does not open the run; Implementation Kickoff Approval and the WT-rules-diet gate still apply.

## Cross-model second opinion

Not run. No `audit_model` key is set; iter-3/iter-4 recorded codex `inconclusive` (usage limit until 2026-09-28) and GLM `inconclusive`. This verdict is Claude-only.

## Gaps

- V1 "guards stay green by design" is established by reading the guard assertions (`update_mirror_heal_test.go:337-348`, `update_codex_wiring_test.go:85`, `update_mirror_heal_wiring_test.go:30`) against the prescribed change, not by implementing host (b) and running `go test ./internal/cli/...`.
- V2 is established by reading `i18n_governance_test.go:423-456`, not by removing the prefix and running the test.
- W1 is established by reading `init_test.go:416-480`; the tests were not run against a modified `validateInitFlags`.
- Only files in this revision's diff and the code they cite were re-read; spec.md and unchanged sections of the other artifacts were not re-audited (partial scope as instructed).
- `moai spec lint` not run (installed binary lags the tree).

## Residual risk

- If the implementer extracts the version-match predicate into a helper and rewrites the early return at :773 to call it, `TestUpdateMirrorHeal_EarlyReturnPreserved` (:341) could lose its literal anchor; D14 does not forbid that refactor. Keeping the predicate literal at :773 (duplicated in update.go) avoids it.
- Host (b) adds a `BackupMoaiConfig` on the version-matched path; tests asserting no backup directory on an "already up to date" run were searched only by the strings `moai-backups` / `BackupMoaiConfig` / `backup` in `update_skip_sync_test.go`, `update_hooks_guidance_test.go`, `update_repair_test.go` — none found.
