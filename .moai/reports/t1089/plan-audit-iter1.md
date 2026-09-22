# SPEC Review Report: SPEC-MODEL-OPUS55-001
Iteration: 1/2 (Tier M ceiling 2)
Verdict: FAIL
Overall Score: 0.78 (Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation (the authoring decisions (a)/(b)/(c) passed in the prompt were treated as claims to verify, not as accepted rationale).

Audited tree: `.claude/worktrees/t1089`, branch `WT-opus-55-default`, HEAD `366ad4ee0`. `git diff --name-only 6e75b74db 366ad4ee0` lists only the four SPEC files plus `.moai/reports/t1089/p4-rednow-6e75b74db.txt`, so the RED-now pins taken on `6e75b74db` are comparable to this tree.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `grep -oE '^- \*\*REQ-OP55-[0-9]{3}' spec.md` → REQ-OP55-001..016, sequential, no gaps or duplicates, 3-digit padding (spec.md:L47-L71). AC-OP55-001..016 are sequential too (acceptance.md:L28-L133).
- [PASS] MP-2 EARS/GEARS format compliance, judged on the requirement layer (spec.md §B, REQ-XXX entries) only: 001/004/005/006/007/009/016 are ubiquitous ("The … shall"); 002/008/011 are event-driven ("When …, the … shall"); 012 is a `Where` clause; 003/013/014/015 use the canonical `shall not` form. The ACs are Given-When-Then plus commands in acceptance.md, which is the correct verification-layer format and is not graded here. `moai spec lint .moai/specs/SPEC-MODEL-OPUS55-001` → `✓ No findings`, rc=0.
- [PASS] MP-3 YAML frontmatter validity: all 12 canonical fields are present at spec.md:L2-L14 (`id`, `title`, `version: "0.1.0"` quoted, `status: draft`, `created`/`updated: 2026-09-23`, `author`, `priority: P1`, `phase: "v3.2.0"`, `module`, `lifecycle: spec-anchored`, `tags` as a comma string). No rejected aliases. `tier: M` is present. `mcp__moai__spec_audit(filter_spec, project_root=<worktree>)` → `modern_era_clean: 1`, INFO `EraAutoDetected` only.
- [PASS/N/A] MP-4 language neutrality: the SPEC does not enumerate multi-language tooling. The template-neutrality obligation is carried by REQ-OP55-016 → AC-OP55-015 (`TestTemplateNoInternalContentLeak`, `TestLanguageNeutrality`, `TestLeakClassNoDateShaInDefaultTier`; I confirmed all three exist at internal/template/internal_content_leak_test.go:1535/1761 and lang_boundary_audit_test.go:190).
- [PASS] MP-5 D7 cross-SPEC reconciliation: the extracted IDs are `SPEC-MODEL-OPUS55-001` (self), `SPEC-OPUS47-COMPAT-001` (status `completed`, not retired/superseded/archived), and `SPEC-MODEL-OPUS-5`, a regex truncation of the rejected dispatch ID named in HISTORY at spec.md:L24. That third ID does not exist, which is a SHOULD finding by design and not BLOCKING. No BLOCKING finding.
- [PASS] MP-6 D8 cross-platform discipline: `grep -c syscall` → 0 in all four artifacts, so D8 auto-passes.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md` matches plan.md:L142 `- 0 × [NEEDS CLARIFICATION].`. The same statement appears at progress.md:L10. This is a count statement, not a marker in the `[NEEDS CLARIFICATION: <topic>]` form, so no marker is open. research.md is absent (Tier M). The mechanical verb still fires on it, which is recorded as D13.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.70 | 0.75 band, docked | REQ-OP55-007 (spec.md:L56) contradicts the launch-effort fallback that the SPEC itself cites at spec.md:L83 (D1). REQ-OP55-009 (spec.md:L61) is over-broad as written (D7). The other REQs are single-interpretation, e.g. REQ-OP55-001 (L47) and REQ-OP55-011 (L63). |
| Completeness | 0.80 | 0.75-1.0 | All sections and the frontmatter are present, with three `### Out of Scope — …` H3s and bullets (spec.md:L75-L91). Gaps: the operator's conflict-enumeration requirement is unmet (D2), and the `.moai/project/*` current-behavior surfaces are neither in scope nor excluded (D6). |
| Testability | 0.75 | 0.75 | I re-measured every RED-now cell and each reproduces (below). The P4 probe is correct and non-vacuous. Gaps: AC-OP55-009 cannot be made binary by a single mutant (D5), REQ-010 is positively asserted in only one file (D4), and the AC-008 regex rejects valid labels (D10). |
| Traceability | 0.85 | 0.75-1.0 | The acceptance.md:L7-L24 table maps all 16 REQs and every AC traces back to a REQ. Coverage is partial for REQ-009 (3 of 7 facts), REQ-010 (1 of 3 locations) and REQ-008 (one half only). |

Aggregate 0.78 is below the Tier M threshold of 0.80. Independently of the score, two blocking-class defects (D1, D2) prevent PASS.

## Independent re-measurement (evidence)

| Probe | Plan claim | My result on `366ad4ee0` |
|---|---|---|
| B.1 `grep -rli 'opus-5\|opus 5' internal pkg cmd --include='*.go' --exclude-dir=templates \| wc -l` | 15 | 15 |
| B.1 same, all file types | 19 | 19 |
| B.1 templates tree | 15 | 15 (identical file list to plan.md:L26) |
| B.1 local `.claude CLAUDE.md AGENTS.md .moai/config` | 15 | 15 (identical list to plan.md:L28) |
| B.4 P4 probe | 147 lines, exit 0 | 147 lines, rc=0; the `sort`ed output diffs **IDENTICAL** against `.moai/reports/t1089/p4-rednow-6e75b74db.txt` |
| AC-001 grep | exit 1 | rc=1 |
| AC-002 grep | exit 1 | rc=1 |
| AC-003 `-c` | 4/1/4/1/3 | model_policy.go:4, glm_slot_test.go:1, model_policy_test.go:4, glm_slot_effort_test.go:1, launcher_test.go:3 |
| AC-005 `opus-55-…` count | 0/0 | 0/0 |
| AC-006 (all three) | 0/0 | 0/0 each |
| AC-007 effort-medium Recommended | 0 | 0 |
| AC-008 EffortLevelMedium Recommended | 0 | 0 |
| AC-010 five `cmp` pairs | SAME | all five SAME; constitution DIFF, as stated |
| AC-011 settings.json.tmpl effort | rc=1 | rc=1 |

Planted-sentence test of the P4 regex (a scratch file run through P4's exact awk):

- Caught: `Opus 5 defaults to high effort`, `` use `claude-opus-5` ``, `the current model is Opus 5.`, `Opus 5. Next`, `Opus 5를 사용`, `claude-opus-5[1m]`, `Opus 5/4.8 philosophy`, `Opus 5.x family`, `Opus 5.5 and Opus 5 both`.
- Not caught, correctly: `Opus 5.5 defaults to medium`, `claude-opus-5-5`, `opus-5-48-prompt`, and a `// superseded` deprecated row.
- Not caught, gap: `Opus5`, and `Opus 5 (superseded prompts) is the default` (D11).

Verdict on the probe: non-vacuous, and it does not falsely fire on "Opus 5.5".

Premise checks on the authoring decisions:

- (a) `ModelAliasFromCanonicalID` (model_policy.go) consults `ModelDeprecatedCanonicalIDs` only. Without the row, `claude-opus-5` would pass through unnormalized, so the premise holds. The consumers are launcher.go:1240, profile_setup.go:77/119 and glm_effort_overlay.go:341. The GLM slot outcome for `claude-opus-5` is unchanged: it resolves to `opus` both before and after. One caveat is recorded in D15: the 4.8→5 precedent kept a named constant (`ModelIDOpus48`, model_policy.go:52-56).
- (b) settings-management.md:93 reads "Intentionally NOT shipped in `settings.json.tmpl`", byte-identical in both copies. I fetched `https://code.claude.com/docs/en/model-config.md` directly. Line 540 reads: "A top-level `effortLevel` in project, local, or managed settings, or one passed with `--settings`, applies to every model." Line 64 reads: "Opus 5.5 requires Claude Code v2.1.280 or later." Line 538 reads: "except that Opus 5.5 defaults to `medium`". Line 609 reads: "`max` isn't accepted as a level in either key", which supports K4. The premise holds.
- (c) The regex at internal/spec/lint.go:1162, `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`, rejects the `-5-` segment, so the premise holds. Platform facts were fetched from `https://platform.claude.com/docs/en/models/opus-5-5/overview.md`: model ID `claude-opus-5-5`, 1M context, 128K output, $4 / $20, "Adaptive (always on)", default effort `medium`, released September 22, 2026. All match spec.md:L34.
- Matrix left untouched: profile_matrix.go:285-298 reads "settled operator input … Do NOT re-derive". The premise holds.

## Defects Found (structured defect-list)

D1. REQ-007-CONTRADICTION: spec.md:L56 against spec.md:L83. Severity: BLOCKING (major). Class: blocking.
- REQ-OP55-007 has the web effort empty option state that "the runtime default applies and … is `medium` on Opus 5.5". That is false for any profile with a stored `model_policy`.
- `resolveLaunchEffort` (internal/cli/launcher.go:1171-1180) and `applyLaunchEffort` (internal/cli/launch_effort_settings.go:52) inject `MapModelPolicyToEffort(prefs.ModelPolicy)` as `effortLevel` in the `--settings` payload when `effort_level` is empty. The policy mapping is high→high, medium→medium, low→low (model_policy.go:279-289).
- The TUI wizard sets `model_policy`, which is `profile.Preferences.ModelPolicy`, preferences.go:25. The web console no longer shows it (schema.go:350-352).
- So a user who picked "Max" gets `high` while the label says "medium". The SPEC documents that fallback itself at L83, so the requirement contradicts the SPEC's own §C.
- Required fix: rewrite REQ-OP55-007 so the empty-option wording matches the real resolution. Either state "profile model policy effort if set, otherwise the runtime default (medium on Opus 5.5)", or render the label from the stored policy. Add an AC that pins the chosen wording against `resolveLaunchEffort`, for example a test that stores `model_policy: high` with an empty effort and asserts the resolved launch effort next to the label text.

D2. CONFLICT-ENUMERATION-MISSING: spec.md:L62, plan.md §C. Severity: BLOCKING (major). Class: blocking, because the operator's card criterion (3) explicitly requires it.
- The card says "enumerate conflicts with current high/xhigh prose". REQ-OP55-010 rewrites only three "defaults to high" statements, and plan.md has no conflict inventory.
- These unenumerated statements conflict with, or must be reconciled against, a medium recommendation:
  - moai-constitution.md:57-60: "Use `xhigh` for coding and agentic work, keep a minimum of `high` for intelligence-sensitive work".
  - agent-authoring.md:380: "raise to `xhigh` for coding/agentic work".
  - model-policy.md:19: "set xhigh explicitly for coding/agentic work".
  - model-policy.md:230: "high: … minimum for intelligence-sensitive work".
  - dynamic-workflows.md:126: "`high` (default)". This line names no Opus version, so neither P4 nor any inventory in the SPEC sees it, but it is false on Opus 5.5.
  - CLAUDE.md:141 and moai-foundation-thinking/SKILL.md:294: `ultrathink` → xhigh. Likely keep, but state it.
  - profile_matrix.go:280: "`xhigh` is retired from the matrix", which already contradicts constitution:59.
- All locations are given in both copies where paired.
- Required fix: add a plan.md table listing every high/xhigh effort statement in both copies (file:line, current text, keep/rewrite decision, reason). Extend REQ-OP55-010 or add a REQ for the dynamic-workflows.md:126 default statement, and give each rewrite an AC.

D3. WEB-LABEL-INVARIANT-CONFLICT: acceptance.md:L82 against internal/web/console_ux_fix_test.go:192-209. Severity: SHOULD-FIX (major). Class: blocking.
- `TestModelOptLabelsEnglishUnified` requires the `f.model.opt.opus[1m]` label to be one English string, byte-identical in all 4 locales. It currently asserts `"Opus 5"` × 4, a design from commit 66d83609a (#1171 "simplify model picker…").
- AC-OP55-007 requires a locale-specific recommendation marker (`Recommended|권장|推奨|推荐`) inside that same label. That breaks the English-unified invariant.
- The SPEC never names this test or reconciles the prior design; plan.md:L30 lists it only as an "Opus-5 hit" file, and K3 classifies similar files as fixtures to leave.
- Required fix: decide explicitly. Option A keeps English-unified labels (`Opus 5.5 (Recommended)` in all four locales), and the AC-007 regex is changed to match. Option B uses localized markers, rewrites `TestModelOptLabelsEnglishUnified` with the reason stated, and names that test in M2.

D4. REQ-010-POSITIVE-GAP: acceptance.md:L70-L75. Severity: SHOULD-FIX (major). Class: blocking.
- Only the constitution gets a positive "Opus 5.5 … medium" check (AC-006 third command).
- For agent-authoring.md:380 and model-policy.md:19/230, which REQ-010 names, the only guard is P4 removing the bare "Opus 5". A rewrite reading "Opus 5.5 defaults to `effort: high`" passes every AC, because P4 is blind to "5.5" by design and AC-006's model-policy checks only look for `2.1.280` / `claude-opus-5-5`.
- Required fix: add, for both copies of agent-authoring.md and model-policy.md, `grep -cE 'Opus 5\.5[^|]*medium'` ≥ 1. Add a negative probe, for example `grep -nE 'Opus 5\.5[^.|]*default[^.|]*high'` → empty, across the rule trees.

D5. AC-009-NOT-BINARY: acceptance.md:L93-L96. Severity: SHOULD-FIX (major). Class: blocking.
- "When one locale's `ModelOpus` label (or one model-policy description) is mutated … Then [both test commands] each fail." One mutant cannot make both fail.
- `TestGetProfileText_OpusAliasValues` (profile_setup_translations_test.go:85-100) reads `profileSetupText`. The wizard agreement test (model_policy_matrix_agreement_test.go:64-80) reads `DefaultQuestions(...)` model_policy descriptions. These are different source files and different surfaces.
- Required fix: specify two mutants, M-a on `profile_setup_translations.go` (a ModelOpus label) → `./internal/cli/` fails, and M-b on `wizard/questions.go` or `wizard/translations.go` (one description) → `./internal/cli/wizard/` fails. Each needs its own expected failing test name and a reverted-tree PASS.

D6. PROJECT-DOCS-UNSCOPED: .moai/project/product.md:151, .moai/project/tech.md:13-15. Severity: SHOULD-FIX (minor). Class: blocking, because it is a scope-completeness gap for card requirement (2).
- `git grep -il 'opus[ _-]\?5'` outside historical paths returns these two files, which neither B.1 nor any §C bullet mentions.
- tech.md:14 reads "Anchor model: `claude-opus-5` (1M-context, the default Opus as of Claude Code 2.1.219)". tech.md:15 reads "`high` (default)". product.md:151 reads "anchored on Opus 5". These are current-behavior claims.
- P4's `find` roots exclude `.moai/project`, so AC-004 cannot see them.
- Required fix: either add them to M3 and to the P4 roots, or add an explicit `### Out of Scope` bullet. Suggested wording: "`.moai/project/*` is regenerated by /moai project; refreshed by the follow-up docs card."

D7. REQ-009-AMBIGUOUS/UNDER-TESTED: spec.md:L61. Severity: SHOULD-FIX (minor). Class: blocking.
- As written, every piece of rule, skill and config prose "that describes current Opus behavior" must state all seven facts (id, medium, v2.1.280, 1M, 128K, $4/$20, adaptive always on). That is 15+15 files, which is almost certainly not intended.
- AC-006 verifies only three facts, in two files. 1M / 128K / $4/$20 / adaptive-always-on have no AC.
- Required fix: name the canonical location of the full fact block, e.g. model-policy.md "Current model generation mapping" in both copies. Say that other prose only needs to be consistent with it. Add grep ACs there for `128K`, `\$4`, `\$20`, and `always on` / `always-on`.

D8. LAUNCH-BEHAVIOR-UNDISCLOSED: spec.md:L81 (§C heading) and plan.md §E. Severity: SHOULD-FIX (minor). Class: optional.
- REQ-OP55-001 changes what the Claude-backend launcher passes as `--model`: `expandModelString` (launcher.go:1139-1152) expands `opus[1m]` to `claude-opus-5-5[1m]`. That is a runtime behavior change, not a label change.
- On Claude Code < v2.1.280 that id is not a known model, and moai has no CC version floor in Go (no `"2.1.NNN"` literal in internal/pkg/cmd non-test code). The §C heading "behavior changes beyond labels and defaults display" reads as if the SPEC changes labels only.
- Required fix: add a K6 residual-risk entry (profile opus launches require CC ≥ 2.1.280; older CC sees an unknown model). Optionally reword the §C heading so it does not imply the change is labels-only.

D9. GOLDEN-COUNT-MISMATCH: acceptance.md:L140. Severity: MINOR. Class: optional.
- "The four golden files are regenerated". `ls internal/cli/testdata/profilewizard/` shows three (groups-ja/ko/zh.golden), consistent with spec.md:L40 and plan.md:L19.
- Required fix: change "four" to "three". Optionally name the regeneration flag `-update-golden` (profile_setup_golden_test.go:90).

D10. AC-007/008-REGEX-FRAGILITY: acceptance.md:L81-L89. Severity: MINOR. Class: optional.
- `ModelOpus1M:[^,]*Opus 5\.5[^,]*(Recommended…)` rejects a valid label containing an ASCII comma before the marker, e.g. `"opus[1m] (Opus 5.5 + 1M context, Recommended)"`.
- The `== 4` counts in AC-007/008 are satisfied by English "Recommended" in all four locales, so they do not prove per-locale wording.
- `(medium|중간|中)` accepts any `中` character.
- Required fix: anchor on the field value (`"[^"]*…"`) instead of `[^,]*`. If per-locale markers are intended (see D3), use per-locale-block assertions.

D11. P4-ALLOWLIST-LOOPHOLE: acceptance.md:L57, plan.md:L57. Severity: MINOR. Class: optional.
- The exemption `superseded` matches anywhere on the line, so a planted `Opus 5 (superseded prompts) is the default` passes.
- `Opus5` and `Opus 5-era` are not matched.
- The local `find` roots omit `.claude/agents`, `.claude/commands`, `.claude/output-styles` and `AGENTS.md`, which have zero hits today, so this is a regression-guard gap only.
- Required fix: tighten the allow-list, e.g. `// superseded` for Go and `(superseded)` / `measured on Opus 5` for prose, and add those roots. Or state these as known limits.

D12. REQ-015-REPORTS-WORDING: spec.md:L70 against commits 6e75b74db / 366ad4ee0. Severity: MINOR. Class: optional.
- REQ-015 says `.moai/reports/` "shall not be modified", yet this card commits new evidence under `.moai/reports/t1089/`. AC-014 (acceptance.md:L122) does not check `.moai/reports` at all.
- Required fix: say "pre-existing files under `.moai/reports/` (this card's `.moai/reports/t1089/` excepted)". Optionally add `.moai/reports ':!.moai/reports/t1089'` to AC-014.

D13. MP-7-VERB-FALSE-POSITIVE: plan.md:L142, progress.md:L10. Severity: MINOR. Class: optional.
- The literal `[NEEDS CLARIFICATION]` in "0 × [NEEDS CLARIFICATION]" trips the clarification-gate grep.
- Required fix: reword as "0 open clarification markers".

D14. REQ-008-HALF-UNTESTED: spec.md:L57, acceptance.md:L91. Severity: MINOR. Class: optional.
- REQ-008 has two halves: the guards fail when the alias target changes until the labels follow, and they reject a bare "Opus 5". AC-009 exercises only the second half.
- Required fix: add a mutant that points the `opus` alias at a hypothetical `claude-opus-6` and expect both guards to fail. The K2 hyphen→dot derivation must also produce `6`.

D15. PRECEDENT/CARD-WORDING: plan.md:L91, model_policy.go:52-56. Severity: MINOR. Class: optional.
- C.1 calls the deprecated-row approach "the established pattern at every previous Opus bump". The 4.8→5 bump actually kept a named constant (`ModelIDOpus48`) as well as the row.
- The card says "remove Opus 5 constant/row", while decision (a) adds a row.
- Both are defensible under operator direction, but the SPEC should state the divergence. The suggested statement is "the constant is removed and the alias-table row is repointed; a deprecated-map row is added so stored `claude-opus-5` normalizes", so the operator can confirm that "row" meant the alias-table row.
- Required fix: add one sentence to C.1. The `ModelIDOpus48` doc comment ("now replaced by ModelIDOpus5") must also be updated. AC-003's `-w` grep will catch it, so no new AC is needed.

## Regression Check

N/A (iteration 1).

## Recommendation

FAIL. Route to manager-spec for these fixes, in order:

1. D1: rewrite REQ-OP55-007 (spec.md:L56) so the empty-effort label matches `resolveLaunchEffort` (model-policy fallback, then runtime default), and add an AC.
2. D2: add a plan.md high/xhigh conflict table covering constitution:57-60, agent-authoring:380, model-policy:19/230, dynamic-workflows:126, CLAUDE.md:141, thinking SKILL:294 and profile_matrix.go:280, with both copies where paired. Extend REQ-010 with ACs.
3. D3: decide English-unified versus localized "Recommended" for the web model label, and name `TestModelOptLabelsEnglishUnified` in M2.
4. D4 and D7: add positive `Opus 5.5 … medium` ACs for agent-authoring.md and model-policy.md (both copies), plus a negative `Opus 5.5 … default … high` probe. Pin REQ-009's fact block to one canonical location, with ACs for 128K, $4/$20 and always-on.
5. D5: split AC-009 into two mutants, one per guard.
6. D6: scope or explicitly exclude `.moai/project/{product,tech}.md`.
7. D8-D15: fold in the minor fixes (K6 residual risk, golden count 3, regex anchoring, allow-list tightening, REQ-015 wording, MP-7 wording, REQ-008 second mutant, C.1 precedent sentence).

The re-audit (iteration 2, the Tier M ceiling) will be scoped to this defect delta.

## Gaps (not observed in this audit)

- No cross-model backend (`audit_multi` / codex / GLM) was run. No `audit_model` key exists under `.moai/config/sections/`, and the orchestrator did not request one.
- No Go test or build was executed. This is a plan-phase audit, and every AC RED-now cell was reproduced by grep/cmp only.
- The official doc facts were fetched with `curl` of the `.md` endpoints on 2026-09-23. They were not cross-checked against the release-notes page.
