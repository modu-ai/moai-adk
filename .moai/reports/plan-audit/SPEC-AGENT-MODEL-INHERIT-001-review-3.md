# SPEC Review Report: SPEC-AGENT-MODEL-INHERIT-001
Iteration: 3/3 (Tier L ceiling 3 — FINAL; escalation report)
Verdict: FAIL
Overall Score: 0.83 (Tier L PASS threshold 0.85; iter-1 0.76 → iter-2 0.82 → iter-3 0.83, not regressing — no STOP signal, but the iteration cap is reached)

Card: t1246. Audited tree: `.claude/worktrees/t1246` (`git rev-parse --show-toplevel`), branch `WT-agent-model-inherit`,
HEAD pinned `b084d67898c370b58926c6adaa63eb8b28888210` (read first; matches the requested b084d6789).
Artifacts read in full: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, `.moai/reports/t1246/touch-set.{sh,txt}`, review-2.md.
Reasoning context ignored per M1 Context Isolation. The HISTORY 0.4.0 row was treated as a claim and every N1–N9 fix was re-measured against code at this HEAD (no Go code differs from the plan base; the branch carries SPEC/report commits only).

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: spec.md:L55-L97 → extracted `001 … 025`, contiguous, no duplicate; `grep -rnoE 'REQ-AMI-0(2[6-9]|[3-9][0-9])'` over the six artifacts → no output.
- [PASS] MP-2 GEARS (requirement layer; ACs are Given-When-Then in acceptance.md, graded under Group 4): all 25 REQs carry a pattern label and matching form, e.g. REQ-AMI-014 "(Event-driven) **When** `moai update` runs … the update shall remove each such key" (L80), REQ-AMI-006 "(Where)" (L63), REQ-AMI-009 "shall not observe" (L69).
- [PASS] MP-3 YAML frontmatter: spec.md:L2-L13 — all 12 canonical fields, `version: "0.4.0"` quoted, `status: draft`, ISO dates, `priority: P1`, `lifecycle: spec-anchored`, `tags` string; no rejected alias.
- [N/A] MP-4 language neutrality: no programming-language-specific tooling named; REQ-AMI-023 carries the template-neutrality obligation.
- [PASS] MP-5 D7: D7 verb over spec.md → `SPEC-MODEL-PROFILE-MATRIX-002 status: superseded`, reconciled at spec.md:L103 ("This SPEC reverses the target of … SPEC-MODEL-PROFILE-MATRIX-002 (superseded …)" + sync-phase closure). Others: ENFORCE-001 completed, PROFILE-MATRIX-001 completed, AUDIT-MODEL-PIN-001 completed, MATRIX-CORE-001 in-progress, MATRIX-DOCS-001 in-progress, MATRIX-CONFIG/SURFACES-001 draft. SHOULD only: `SPEC-ALWAYS-LOADED-DIET-002`, `SPEC-ROLE-NAMING-DOCS-001` not found (unmerged sibling branches, explained at L101-L102).
- [PASS] MP-6 D8: `grep -c syscall` → 0 in all six artifacts.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output, exit 1.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75-1.0 | D14 now names measured hosts (design.md:L33, lines verified). But REQ-AMI-020's premise "consumed … by configuration validation" (spec.md:L89) is false after M5 (R1), and spec §D Out of Scope says the `moai profile setup` wizard is "not changed" (L108) while H24/M4 rewrite its labels (R5). |
| Completeness | 0.80 | 0.75-1.0 | N3/N7/N8/N9 inventory gaps closed; touch set 252 reproduced. Remaining: `ModelEffort` retention unstated (R1), `web-i18n-agent-descriptions` / i18n.js absent (R2), effort_level copy and ja/zh profile-setup lines (R6), ja/zh multi-llm residue (R7), Q7 test surface (R8). |
| Testability | 0.80 | 0.75-1.0 | AC-AMI-007(d) RED today with exactly 3 files (reproduced); AC-AMI-002 gate reproduced → 0. But AC-AMI-020 cannot pass at M5 with the §F `shipped-key-inventory` "site kept" row (R2); AC-AMI-006's "non-zero count of files checked" names output the linter does not print (R4); AC-AMI-014's user-cancel row has no seam (R9). |
| Traceability | 0.93 | 1.0 minus nits | 25 REQ ↔ 25 AC, every AC maps an existing REQ (acceptance.md:L15-L39). Nits: Q7 folded into REQ-AMI-015 but spec §E, plan §I and acceptance §D.3 still say Q1–Q6; REQ/AC-025 point to "research.md §I" for the 8-page list, which lives in §F (research.md:L193-L198). |

Aggregate (harmonic) ≈ 0.83, below the Tier L threshold 0.85, with two major blocking defects, one of them a partially unresolved iteration-2 defect (N2).

## Regression Check (iteration-2 defects)

- N1 (strip placement) — RESOLVED. Every cited line re-read: `update_template_sync.go:539` `backup.BackupMoaiConfig` in the "Backup" step; `:620` `RestoreMoaiConfigRetained`; `:638` `renderRetainedKeyAdvisory`; `:773-776` version match `return true, nil`; `:795-799` "Merge cancelled by user" `return true, nil`; `update.go:388` pre-backup deny strip with comment :377-381; `update.go:524-533` `if syncSkipped { … return nil }`; `update_clean_install.go:511` `RestoreMoaiConfig` then `:554` `stripRetiredV2DenyEntries`. Backup (`.moai-backups/<ts>`, `defs/dirs.go:12`) precedes all three hosts; version-matched and user-cancel behaviour stated (REQ-AMI-014 L80, design §C rows 7-8). Residual on host (c) → R3.
- N2 (rosterguard bindings) — PARTIALLY RESOLVED (counts as unresolved). The §F table exists and most rows are right, but: `shipped-key-inventory` is marked "Fixture regenerated … site kept" (design.md:L120) although its roster membership comes from the `llm.profiles.*` rows (registry.go:207 Note "one key pair per matrix row"); with the removed keys filtered out the fixture names 7 of 13 roster names (measured: only e2e-tester, manager-design, manager-develop, manager-docs, manager-spec, plan-auditor, super-advisor remain), so check.go:197-209 reports a membership gap at M5. `config-retained-agent-names` "re-pointed" leaves an orphan (R1). `web-i18n-agent-descriptions` (registry.go:137) is missing (R2).
- N3 (verify-judge channel) — RESOLVED. H22 (design.md:L85) + M6 (plan.md:L113-L115) + AC-AMI-007(d); `grep -rlE 'judge_effort|JUDGE_EFFORT' .claude/rules .claude/skills .claude/workflows .claude/hooks internal/template/templates/.claude` → exactly `verify-judge-effort-contract.md`, `sync-audit-4dim.js`, `test-judge-effort-contract.sh` (RED as claimed); template `sync-audit-4dim.js` has no `judge_effort`; the rule file has 0 `[HARD]` markers, so H22's "Marker none" holds.
- N4 (test-layer order) — RESOLVED. plan.md:L149-L162 test table reproduced symbol by symbol with `grep -rlE <sym> --include='*_test.go' internal cmd pkg`: ResolveAgentModelEffort 9 files, ProfileMatrixAgents 4, DefaultProfileMatrix|ResolveHarness… 7, ApplyProfile… 2, AgentModelGuard 2, WorkflowAgents|ModelRouting 5 — all match the table.
- N5 (AC-006 vacuous) — RESOLVED as asked (`--path` flag exists, `agent_lint.go:160`); the added anti-vacuity clause introduces R4.
- N6 (touch-set breakdown/collation) — RESOLVED. `sh touch-set.sh` → 252 lines, `diff` vs committed file → identical; `LC_ALL=C sort -c` → exit 0; per-directory counts equal research.md:L253-L257 exactly (8 `.claude/skills`, 54 `internal/template`, 60 `docs-site/content` = en 17 / ko 17 / ja 13 / zh 13).
- N7 (docs residue) — RESOLVED for the 8 cited pages; residual R7 (optional).
- N8 (agent-authoring residue) — RESOLVED. H23 (design.md:L86); agent-authoring.md:219 and :333 confirmed in both trees.
- N9 (profile-wizard copy) — RESOLVED in intent (H24, M4); residuals R5 (blocking consistency) and R6 (optional).

Gates re-run now: `git merge-base --is-ancestor WT-rules-diet develop` → exit 1 (t1175 not merged, gate holds; develop `b4f798dcc`); `git diff --name-only e62c3e183 WT-role-naming-docs | LC_ALL=C sort | LC_ALL=C comm -12 - touch-set.txt | wc -l` → 0 (branch tip `024b95f77`).

## Defects Found (structured defect-list)

R1. ROSTER-CONSUMER-PREMISE — spec.md:L89 (REQ-AMI-020), design.md:L23 (D4), design.md:L97 (§E config row), design.md:L117 (§F `config-retained-agent-names`) — `config.retainedAgentNames` (internal/config/profile.go:142) has exactly one consumer, `validateAgentOverrides` (profile.go:182; `grep -rn retainedAgentNames internal | grep -v _test` → profile.go:138/142/182 + the registry anchor only), which M5 deletes. After M5 no configuration validation consumes the roster, so REQ-AMI-020's "consumed … by configuration validation" is false, and a "re-pointed" `retainedAgentNames` is an unused package var — `.golangci.yml:30` enables `unused`, failing the per-milestone lint gate (plan §E) — or, if deleted, the §F row's Path points at a file §E removes. The same removal of `profile.go` (§E "Remove: `profile.go` (profile enum, overrides validation)") would take `config.ModelEffort` (profile.go:73), which the RETAINED audit pins use (`config/audit_models.go:67/75/82`, `config/defaults.go:1120`, `cli/mcp_claude.go:181`, `cli/mcp_codex.go:211+`, `cli/mcp_glm.go:172`); §E's Keep column does not retain it. — Severity: major — Class: blocking — Required fix: in REQ-AMI-020 drop "and by configuration validation" (or name the validation that still consumes it); in D4/§E/§F dispose of `retainedAgentNames` and the `config-retained-agent-names` site explicitly ("removed in M5 with its last consumer"); in §E Keep, retain `ModelEffort` (relocated beside `audit_models.go` or kept in a surviving file).

R2. ROSTERGUARD-DISPOSITION-ERRORS (iteration-2 N2 residual) — design.md:L120, design.md:L106-L126 (§F), plan.md:L65-L74 (M2), acceptance.md:L34 (AC-AMI-020) — (a) `shipped-key-inventory` "site kept" is wrong: membership of `internal/config/testdata/shipped_key_inventory.yaml` comes from the `llm.profiles/harness_agents/agent_overrides` and `workflow.model_routing` rows (registry.go:207 Note); filtering those out leaves 7 of 13 retained names (measured), so rosterguard's ClaimMembership check (check.go:197-209) fails at M5 and AC-AMI-020 cannot pass; (b) `web-i18n-agent-descriptions` (registry.go:137, `internal/web/assets/i18n.js`, ClaimMembership, "one agentdesc.* key per definition file") is absent from §F and i18n.js is absent from the touch set, yet the `agentdesc.*` keys' only consumer is `agentFMRow` (fieldsets.templ:694), which M2 deletes, together with ~40 `agentfm.*`/`fieldDesc.agentfm.*`/`agentdesc.*` keys (`grep -cE 'agentfm|agentdesc' i18n.js` → 149 lines); REQ-AMI-011 removes "its UI" but no milestone disposes of these keys or the site, and `i18n_untranslated_allowlist_test.go:260-268` justifies the `agentdesc.` exemption by the panel. — Severity: major — Class: blocking — Required fix: change the `shipped-key-inventory` row to "Removed (the fixture no longer enumerates the roster)" at M5; add a §F row for `web-i18n-agent-descriptions` (kept, with the agentdesc keys retained and justified, or removed with the keys at M2), add `internal/web/assets/i18n.js` to the touch set and to M2, and state the disposition of the `agentfm.*` i18n keys and the `agentdesc.` exemption entry.

R3. CLEAN-INSTALL-DOUBLE-REPORT — design.md:L33 (D14 host c), spec.md:L80 (REQ-AMI-014) — REQ-AMI-014 requires each removed key to appear "once … as removed (not as retained)". On host (c), `backup.RestoreMoaiConfig` (update_clean_install.go:511) prints a retained-key advisory line for every retained key through `retainedKeySink` (update/backup/restore.go:64-75) before the strip at :554, so each stripped key is reported first as retained. D14 states the retained-key filter for host (a) only; AC-AMI-014 tests only §C rows, none of which is the clean-install path. — Severity: minor — Class: blocking — Required fix: state in D14(c) that the clean-install host calls `RestoreMoaiConfigRetained` (or filters the strip's key set before the advisory), and add the clean-install path to AC-AMI-014 or §C.

R4. AC-006-UNOBSERVABLE-COUNT — acceptance.md:L20 (AC-AMI-006) — "reports a non-zero count of files checked": `moai agent lint` prints no files-checked count; text mode prints "No agent files found in …" or "No violations found", and JSON `Total` counts violations (agent_lint.go:204, :244-258). The clause cannot be observed as written. — Severity: minor — Class: blocking — Required fix: replace with "does not print `No agent files found`" (or name an existing output field).

R5. OUT-OF-SCOPE-CONTRADICTION — spec.md:L108 vs design.md:L87 (H24), plan.md:L87-L88 — §D Out of Scope says "the `moai profile setup` wizard … are not changed", while H24 and M4 rewrite that wizard's `model_policy` title/description in four locales. — Severity: minor — Class: blocking — Required fix: amend the Out of Scope bullet to "not changed in behaviour; its `model_policy` wording is rewritten (design H24)".

R6. PROFILE-COPY-RESIDUE — design.md:L87 (H24) — H24 cites `profile_setup_translations.go:165/260` only (en, ko) though it says four locales; ja `:356-357` and zh `:452-453` carry the same text. The sibling `effort_level` descriptions — "Per-agent effort comes from the agent model policy instead." (`wizard/translations.go:445/457/469/481`, `profile_setup_translations.go:181/277/373/469`) — become false and are not listed. — Severity: minor — Class: optional — Required fix: cite all four locales and add the eight `effort_level` lines to H24.

R7. DOCS-RESIDUE-LOCALES — research.md:L193-L198, spec.md:L97, acceptance.md:L39 — `docs-site/content/ja/multi-llm/_index.md:81` ("エージェント別モデル割り当て表") and `zh/multi-llm/_index.md:80` ("各代理的模型分配表") carry the same stale link description as `en/multi-llm/_index.md:95`, which is in the set; neither is in the touch set. REQ-AMI-025 and AC-AMI-025 cite "research.md §I" for the 8-page list, which is in research §F (L193-L198). — Severity: minor — Class: optional — Required fix: add the two pages (making the explicit list 10) and correct the section pointer.

R8. Q7-TRACE-AND-SURFACE — spec.md:L122-L129 (§E lists Q1–Q6), plan.md:L176-L177, acceptance.md:L56-L57, progress.md:L16 — Q7 is recorded and folded into REQ-AMI-015, but spec §E, plan §I and the DoD still enumerate Q1–Q6. Q7's code surface is only partly in the touch set: `internal/cli/wizard/wizard.go:484` (`case "model_policy"`) and the init-wizard tests that name the question (`wizard/{model_policy_default,questions,restructure,unified_form,wizard,agent_wiring_question}_test.go`) are absent (they fail at M4 `go test`, not at compile). HISTORY rows are out of order (0.4.0 above 0.3.0, spec.md:L25-L26). — Severity: minor — Class: optional — Required fix: add Q7 to §E/§I/DoD; add the wizard files to the touch set; reorder HISTORY.

R9. MIGRATION-TEST-SEAMS — acceptance.md:L28 (AC-AMI-014), design.md:L33 (D14 b) — the user-cancel fixture needs a cancel path, but `confirmViaPreview` (update_template_sync.go:849-853) is a plain function that returns an error in a non-TTY; and host (b) must tell version-match from cancel although `runTemplateSyncWithProgress` returns the same `(true, nil)` for both. Neither the seam nor the signature change is named. — Severity: minor — Class: optional — Required fix: one sentence in M1 naming the test seam and the skip-reason return value.

## Recommendation

FAIL at the final iteration allowed by the Tier L ceiling. All seven must-pass criteria pass; seven of nine iteration-2 defects are fully resolved against the code (N1, N3, N4, N5, N6, N8 cleanly; N7/N9 with optional residue). The verdict rests on the aggregate 0.83 < 0.85 and on two major blocking defects, R2 being the unresolved part of N2.

Escalation per the Retry Loop Contract (iter 3 FAIL → user intervention). Defect history: iter-1 16 defects (D1–D16) → all resolved; iter-2 9 defects (N1–N9) → 7 resolved, N2 partially, N9 with a new consistency defect; iter-3 9 defects (R1–R9), 2 major blocking, 3 minor blocking, 4 optional. No defect appeared unchanged across all three iterations (no stagnation), and the score rose each round.

The blocking fixes are all small text edits in design.md §E/§F, spec.md L89/L108, acceptance.md L20/L28 and the touch-set script:
1. R1 — REQ-AMI-020 premise; dispose of `retainedAgentNames` + its site; retain `ModelEffort`.
2. R2 — `shipped-key-inventory` → removed at M5; add `web-i18n-agent-descriptions` + i18n.js to §F, touch set and M2.
3. R3 — clean-install host uses the retained-returning restore or filters; cover it in AC-AMI-014.
4. R4 — replace the files-checked clause.
5. R5 — amend the Out of Scope bullet.

Options for the orchestrator's user question: (1) PASS-with-debt, carrying R1–R5 as run-phase M1/M2/M5 obligations recorded in progress.md (each is mechanically checkable at the milestone boundary — rosterguard and golangci-lint would catch R1/R2 anyway); (2) one more revision with a confirming re-audit scoped to R1–R5 (explicit user override of the cap); (3) scope reduction is not indicated — the defects are local, not structural.

## Cross-model second opinion

`mcp__moai__audit_multi` (project_root = this worktree, target baseBranch, session `t1246-plan-audit-iter3`): codex `inconclusive` (usage limit until Sep 28th, 2026 2:37 PM), glm `inconclusive` ("z.ai response carried no text content"); both fail-open, participant_count 1, overall = this in-session verdict (fail). The tool reported build lag: installed moai `a8a9b9376` is an ancestor of HEAD `b084d6789`.

## Gaps

- `moai spec lint` was not run (installed binary lags HEAD).
- R2(a) was established by filtering the fixture with the removed-key pattern and counting roster names (7/13), not by running rosterguard against a regenerated fixture; the exact regenerated content is the M5 implementer's.
- R1's lint failure is inferred from `unused` being enabled in `.golangci.yml`; golangci-lint was not run on a trial deletion.
- R3 was established by reading restore.go and update_clean_install.go call order, not by running a clean-install update.
- `TestAlwaysLoadedTokenBudget` was not re-measured (unchanged plan-phase baseline 61; re-baselined after t1175 by design).

## Residual risk

- Other rosterguard sites with whole-file bodies (no BlockStart) could lose roster names when prose is rewritten in M7/M8; only the §F rows and the sites reached above were checked.
- The docs-site probes remain pattern-based; locale pages that paraphrase the removed model without any pattern token can survive M8.
