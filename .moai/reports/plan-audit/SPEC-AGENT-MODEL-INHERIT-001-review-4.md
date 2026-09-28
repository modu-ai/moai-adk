# SPEC Review Report: SPEC-AGENT-MODEL-INHERIT-001
Iteration: 4/3 (one round beyond the Tier L ceiling, operator-approved — progress.md §E.1 line "plan-audit iter-3 … approved ONE extra revision round")
Verdict: FAIL
Overall Score: 0.86 (Tier L PASS threshold 0.85; iter-1 0.76 → iter-2 0.82 → iter-3 0.83 → iter-4 0.86, rising — no STOP signal)

Card: t1246. Audited tree: `.claude/worktrees/t1246`, branch `WT-agent-model-inherit`, HEAD pinned
`f50ca57dafc97efe797f70bbc0b2c4f268995672` (read first; starts with the requested `f50ca57da`).
Artifacts read in full: spec.md, plan.md, design.md, acceptance.md (diff + changed rows), research.md (diff + section map), progress.md, `.moai/reports/t1246/touch-set.{sh,txt}`, review-3.md.
Reasoning context ignored per M1 Context Isolation. The HISTORY 0.5.0 row and the commit message were treated as claims; every R1–R9 fix was re-measured against code at this HEAD. `git diff --name-only d6992e3a0 HEAD | grep -vc '^\.moai/'` → `0`, so all code citations are equally valid at the plan base.

Verdict basis: the aggregate clears 0.85, but one **major blocking** defect introduced by this revision's R9 fix (V1) prescribes a change that breaks a completed SPEC's requirement and its mechanical guard. Per M6, blocking findings are fixed before the verdict is revisited; the score alone does not carry a PASS over a design instruction that turns M1 red.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `grep -oE 'REQ-AMI-[0-9]+' spec.md | sort -u` → 001…025 contiguous, no duplicate (spec.md:L56-L98).
- [PASS] MP-2 GEARS (requirement layer only; ACs are Given-When-Then in acceptance.md and graded under Group 4): every REQ carries its pattern label and form, e.g. REQ-AMI-014 "(Event-driven) **When** `moai update` runs … the update shall remove each such key" (L81), REQ-AMI-006 "(Where) **Where** a user-authored agent file…" (L64), REQ-AMI-020 "(Ubiquitous) The retained-agent roster … shall have exactly one source of truth" (L90).
- [PASS] MP-3 YAML frontmatter: spec.md:L2-L13 — all 12 canonical fields, `version: "0.5.0"` quoted, `status: draft`, ISO dates, `priority: P1`, `lifecycle: spec-anchored`, `tags` string; no rejected alias.
- [N/A] MP-4 language neutrality: no programming-language-specific tooling named; REQ-AMI-023 carries the template-neutrality obligation.
- [PASS] MP-5 D7: verb re-run over spec.md → `SPEC-MODEL-PROFILE-MATRIX-002 status: superseded`, reconciled at spec.md:L104 ("This SPEC reverses the target of … SPEC-MODEL-PROFILE-MATRIX-002 (superseded …)" + sync-phase closure). Others completed/in-progress/draft. SHOULD only: `SPEC-ALWAYS-LOADED-DIET-002`, `SPEC-ROLE-NAMING-DOCS-001` NOTFOUND (sibling branches, explained L102-L103). No D7 BLOCKING. (V1 below is a cross-SPEC conflict with a SPEC this one does not reference, so it falls outside the D7 verb and is graded as a consistency defect.)
- [PASS] MP-6 D8: `grep -c syscall` → 0 in all six artifacts.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no output, exit 1.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.82 | 0.75 | R1/R5 premises corrected (spec.md:L90, L109). But design D14 (design.md:L33) and plan M1 (plan.md:L62-L64) instruct a `runTemplateSyncWithProgress` return-type change that contradicts completed SPEC-UPDATE-MIRROR-HEAL-001 REQ-UMH-003 ("with its present `(skipped, err)` contract") — V1. |
| Completeness | 0.83 | 0.75 | Touch set 271 reproduced; R2/R6/R7/R8 inventory closed. Missing: 5 cli test files calling `runTemplateSyncWithProgress` plus `update_codex_wiring_test.go` (V1), `internal/web/i18n_governance_test.go` (V2), `internal/web/assets/app.js` (V3). |
| Testability | 0.88 | 0.75-1.0 | AC-AMI-006 now observable (must-fail LR-01 fixture, `agent_lint.go:449`; "No agent files found" string at :204); AC-AMI-014 gained the clean-reinstall row; AC-AMI-020 now a clean "prints nothing" grep. Residual: AC-AMI-006 "no finding for user-pinned.md" depends on unspecified fixture content (LR-05 warning, V4). |
| Traceability | 0.95 | 1.0 minus nit | 25 REQ ↔ 25 AC, each AC maps an existing REQ (acceptance.md "maps REQ-AMI-001…025" verified in order). Q7 now traced in spec §E (L131), plan §I (L184), DoD; section pointers corrected (REQ-025 → research §F, which holds the list at research.md:L193-L199). |

Aggregate (harmonic mean of 0.82 / 0.83 / 0.88 / 0.95) ≈ 0.86.

## Regression Check (iteration-3 defects R1–R9, verified against code)

- **R1 — RESOLVED.** `grep -rn 'retainedAgentNames' --include='*.go' internal cmd pkg` → `config/profile.go:138,142,182`, `rosterguard/registry.go:52`, and a comment in `profile_test.go:168`; sole reader `validateAgentOverrides` (profile.go:179/182). REQ-AMI-020 (spec.md:L90) now names rosterguard as the only surviving consumer; D4 (design.md:L23), §E config row (L98), §F row (L118) and plan M5 (L101-L104) remove the variable and its site together at M5. `ModelEffort` (profile.go:73) is kept and relocated: consumers re-measured `config/audit_models.go:67,75,82`, `config/defaults.go:1120`, `cli/mcp_claude.go:181-182`, `cli/mcp_codex.go:211-254`, `cli/mcp_glm.go:172` — all match §E. Lint baseline run by this auditor: `golangci-lint run ./internal/config/...` → `0 issues.`, exit 0; `.golangci.yml:30` `- unused` confirmed. (First attempt returned exit 3 "parallel golangci-lint is running"; the retry completed.)
- **R2 — RESOLVED, with a new defect (V2).** (a) `shipped-key-inventory` now "Removed" at M5 (design.md:L121); re-measured: registry.go:207 Note "Derived llm.profiles.* key inventory — one key pair per matrix row"; after filtering `profiles|harness_agents|agent_overrides|model_routing` the fixture names 7 roster agents (e2e-tester, manager-design, manager-develop, manager-docs, manager-spec, plan-auditor, super-advisor). (b) `web-i18n-agent-descriptions` (registry.go:137-142, `internal/web/assets/i18n.js`, ClaimMembership) now has a §F row (L122), M2 carries it (plan.md:L74-L77), i18n.js and the allowlist test are in the touch set; `grep -cE 'agentfm|agentdesc' i18n.js` → 149 (37 `agentdesc.` lines). But removing the sole exempt prefix empties `i18nEnExemptPrefixes`, which `TestI18nKeyCoverageReverse` forbids — V2.
- **R3 — RESOLVED.** D14(c) (design.md:L33) switches `update_clean_install.go:511` from `backup.RestoreMoaiConfig` to `RestoreMoaiConfigRetained` and filters the advisory; verified the wrapper prints each non-`KeptOverDefault` ref to `retainedKeySink` (restore.go:64-75). New §C row (design.md:L52) and AC-AMI-014 clean-reinstall clause (acceptance.md:L28); routing verified at update.go:409-424 (`fingerprint.IsV2 && isMoAIProject(cwd)` → `runCleanReinstall`).
- **R4 — RESOLVED.** AC-AMI-006 (acceptance.md:L20) now uses "does not contain `No agent files found`" (printed at agent_lint.go:204) and a must-fail fixture with a literal `AskUserQuestion` yielding LR-01 (rule emitted at agent_lint.go:449). Residual optional V4.
- **R5 — RESOLVED.** spec.md:L109 "the behaviour of the `moai profile setup` wizard (its `model_policy` and `effort_level` wording is rewritten to main-session terms — design H24) … not changed" — consistent with H24 and M4.
- **R6 — RESOLVED.** H24 (design.md:L88) lines verified: `wizard/translations.go:444,456,468,480` (model_policy title+description on one line each) and `:445,457,469,481` (effort_level; en carries "Per-agent effort comes from the agent model policy instead."); `profile_setup_translations.go:165-166, 260-261, 356-357, 452-453` (ModelPolicyTitle/Desc) and `:181, 277, 373, 469` (EffortLevelDesc). 4 + 8 + 4 + 4 = 20 lines, as stated.
- **R7 — RESOLVED.** `ja/multi-llm/_index.md:81` "エージェント別モデル割り当て表", `zh/multi-llm/_index.md:80` "各代理的模型分配表" confirmed and in the touch set; research §F now lists 10 pages; REQ-AMI-025 and AC-AMI-025 point at §F. "62 at `d6992e3a0`" holds (touch set has 62 `docs-site/content` paths; no non-`.moai` file differs from d6992e3a0).
- **R8 — RESOLVED.** Q7 in spec §E (L131), plan §I (L184), DoD (acceptance §D.3); `wizard/wizard.go:484` `case "model_policy":` and 13 wizard files in the touch set; HISTORY rows now ordered 0.1.0 → 0.5.0 (spec.md:L23-L27).
- **R9 — RESOLVED as asked (seam and skip reason named, design.md:L33, plan.md:L62-L64), but the named seam is a regression → V1.**

Gates reproduced now:
- `sh .moai/reports/t1246/touch-set.sh` → exit 0, 271 lines, `diff` vs committed file → identical; `LC_ALL=C sort -c` → exit 0; per-directory counts equal research.md:L259-L262 exactly (62 docs-site/content, 38 internal/cli, 12 internal/config, 28 internal/web …).
- `git merge-base --is-ancestor WT-rules-diet develop` → exit 1 (develop `b4f798dcc`, WT-rules-diet `3a48485af`) — REQ-AMI-001 gate still closed, as recorded.
- `git diff --name-only e62c3e183 WT-role-naming-docs` (merge-base of HEAD and the branch; tip `024b95f77`) → 32 files; `LC_ALL=C sort | LC_ALL=C comm -12 - touch-set.txt | wc -l` → 0.

## Defects Found (structured defect-list)

V1. SKIP-REASON-SEAM-BREAKS-UMH-CONTRACT — design.md:L33 (D14 "Skip reason and seam"), plan.md:L62-L64 (M1) — The R9 fix prescribes that `runTemplateSyncWithProgress` "returns a skip reason (`none` / `versionMatch` / `userCancel`) instead of a bare bool". Completed SPEC-UPDATE-MIRROR-HEAL-001 REQ-UMH-003 (spec.md:L81-L83 there): "The template-sync step shall retain its version-match early return unchanged, in its present position, with its present `(skipped, err)` contract." It is mechanically guarded: `internal/cli/update_mirror_heal_test.go:337-339` asserts the literal `func runTemplateSyncWithProgress(cmd *cobra.Command) (skipped bool, err error)` and :345-348 that the version-match branch returns `(true, nil)`; `update_codex_wiring_test.go:85` pins `if syncSkipped {` in update.go. The signature change also breaks 8 call sites in `coverage_improvement_test.go:2376,3690`, `target_coverage_test.go:260,317`, `update_hooks_guidance_test.go:86`, `update_mirror_heal_test.go:375`, `update_skip_sync_test.go:86,163`. None of these six files is in the touch set, SPEC-UPDATE-MIRROR-HEAL-001 is not referenced or reconciled anywhere in the SPEC, and M1 ("Additive changes … Nothing is deleted") would close red on `go test ./internal/cli/...`. — Severity: major — Class: blocking — Required fix: keep the `(skipped bool, err error)` contract and have host (b) tell the version match from the cancel without changing it — e.g. re-evaluate the same predicate the early return uses (`version.GetVersion() == plan.GetProjectConfigVersion(root) && !forceUpdate`) inside the `syncSkipped` branch, or record the reason in a package-level/unexported side channel; state the chosen form in D14 and M1. If the signature change is kept instead, reconcile with SPEC-UPDATE-MIRROR-HEAL-001 explicitly (supersede REQ-UMH-003 with a rationale) and add the six test files to the touch set and to plan §G.

V2. EMPTY-EXEMPT-REGISTRY — design.md:L122 (§F `web-i18n-agent-descriptions`), plan.md:L75-L77 (M2) — Removing the `agentdesc.` entry empties `i18nEnExemptPrefixes` (`i18n_untranslated_allowlist_test.go:264-269`, its only member), and `TestI18nKeyCoverageReverse` (`internal/web/i18n_governance_test.go:423`) fails with "en-exempt prefix registry is empty" (:437-439). `i18n_governance_test.go` is not in the touch set, and the SPEC does not say whether the non-empty assertion (from completed SPEC-I18N-GOVERNANCE-001, REQ-I18NGOV-020 "sole initial member is `agentdesc.`") is dropped or the registry is kept. M2's "M1 characterisation tests stay green" would hold, but `go test ./internal/web` would not. — Severity: minor — Class: blocking — Required fix: add one sentence to §F/M2 naming the disposition (drop the `len == 0` assertion at i18n_governance_test.go:437-439, keeping the justification check for future members) and add the file to the touch set.

V3. DEAD-AGENTFM-JS — `internal/web/assets/app.js:390-480` (`wireProfileMatrix`, `reapplyHaikuLocks`, `wireHaikuEffortLock`) query `select[name^="agentfm."]` and the `performance_tier` radios; after M2 they match nothing. app.js is not in the touch set and M2 does not mention it. Harmless at runtime (no elements → no-op), but it is residue of the removed UI. — Severity: minor — Class: optional — Required fix: add app.js to M2 and the touch set, or note it as accepted residue.

V4. AC-006-FIXTURE-CONTENT — acceptance.md:L20 — "it reports no finding for `user-pinned.md`" is stricter than the prior "no error": LR-05 warns on missing `isolation:` for standalone agents (agent_lint.go:126), so whether it passes depends on fixture fields the AC does not state. — Severity: minor — Class: optional — Required fix: say "no LR-03/LR-12/LR-13 finding and no error", or state that the fixture is otherwise lint-clean.

V5. RESIDUAL-CONSUMERS-UNNAMED — plan.md:L144-L154 (§G) — `config.IsValidProfile` (removed with profile.go in M5) has non-test consumers `cli/update.go:55,657` and `cli/init.go:363,370`; M4's "flag handling reduced to deprecation warnings" implies their removal but §G does not list them, and REQ-AMI-016 does not say whether an invalid `--profile` value is still rejected. The init-wizard question's own translations (`wizard/translations.go:89-91` ko, `:179-181` ja, `:269-271` zh, plus en) are not named for removal with the Q7 question. — Severity: minor — Class: optional — Required fix: add an `IsValidProfile` row to §G (M4), state invalid-value behaviour in D10, and name the question translations in M4.

## Recommendation

FAIL on V1. Must-pass criteria all pass; all nine iteration-3 defects are resolved against the code (R1 including an auditor-run `golangci-lint` baseline, R2 with the new V2, R9 with the new V1). The aggregate rose to ≈ 0.86, above the Tier L threshold. The FAIL rests on V1 alone being major and blocking: the SPEC now instructs a signature change that violates a completed SPEC's requirement and trips its guard test at the first milestone.

Fix instructions for manager-spec:
1. V1 — design.md:L33 D14 "Skip reason and seam" and plan.md:L62-L64: keep `(skipped bool, err error)`; host (b) re-evaluates the version-match predicate (or reads a side channel) to separate version match from cancel. Keep the `confirmViaPreview` function-variable seam (it has no source-text guard; `grep confirmViaPreview --include=*_test.go` → none).
2. V2 — design.md:L122 and plan.md:L75-L77: name the `i18n_governance_test.go:437-439` disposition; add the file to the touch set.
3. V3–V5 optional.

Options for the orchestrator (the iteration budget, including the operator's extra round, is spent): (1) PASS-with-debt, carrying V1 and V2 as M1/M2 obligations recorded in progress.md — both are mechanically caught at the milestone boundary (`TestUpdateMirrorHeal_EarlyReturnPreserved`, `TestI18nKeyCoverageReverse`) and each fix is a sentence; (2) one more scoped revision with a confirming re-audit limited to V1/V2 (explicit operator override); (3) scope reduction is not indicated — the defects are local.

Defect history: iter-1 16 (D1–D16) → resolved; iter-2 9 (N1–N9) → 7 resolved; iter-3 9 (R1–R9) → all 9 resolved at iter-4; iter-4 5 new (V1 major blocking, V2 minor blocking, V3–V5 optional), two of them introduced by the R2/R9 fixes. No defect has persisted unchanged across iterations (no stagnation).

## Cross-model second opinion

Not run. No `audit_model` key is set in `.moai/config/sections/*.yaml`; iteration 3's `audit_multi` call returned codex `inconclusive` (usage limit until 2026-09-28) and GLM `inconclusive`. This verdict is Claude-only.

## Gaps

- `moai spec lint` not run (installed binary `a8a9b9376` lags the tree).
- V1's test breakage is established by reading the guard assertions and call sites, not by applying the signature change and running `go test`.
- V2 is established by reading `i18n_governance_test.go:423-445`, not by deleting the prefix and running the test.
- The roster-name count used for R2(a) came from a 12-name pattern (the fixture's 13th name was not identified); the 7 remaining after filtering match research.md §K.
- acceptance.md and research.md were read as diff plus changed/cited rows rather than re-read line by line in full; spec.md, design.md and plan.md were read in full.

## Residual risk

- Other source-text guard tests over `update.go` / `update_template_sync.go` / `update_clean_install.go` may pin shapes D14 changes; only the `runTemplateSyncWithProgress`, `syncSkipped` and `confirmViaPreview` strings were searched.
- The clean-reinstall advisory renderer (legacy stderr text vs `renderRetainedKeyAdvisory`) is left to the implementer; the `KeptOverDefault` skip the legacy wrapper applies must survive the switch.
