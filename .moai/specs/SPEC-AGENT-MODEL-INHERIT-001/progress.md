# Progress — SPEC-AGENT-MODEL-INHERIT-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1246. Branch `WT-agent-model-inherit`, base develop `d6992e3a0`. Tier L.
- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md; run-entry touch set `.moai/reports/t1246/touch-set.{sh,txt}`.
- Inventory measured 2026-09-26 (research.md). Budget baseline: 77539 / 77600 tokens, headroom 61.
- Run gate: `git merge-base --is-ancestor WT-rules-diet develop` exit 0 (REQ-AMI-001; exit 1 on 2026-09-26); t1257 overlap re-measured at run entry (REQ-AMI-002; C-collated intersection 0 at `024b95f77`, touch set 272 paths).
- Operator questions — RESOLVED 2026-09-26 (answered by the operator in the lane window, relayed by the coordinator):
  - Q1 — web console: delete the agent-settings tab entirely (UI + API). → REQ-AMI-011, design D7.
  - Q2 — leftover user config keys: `moai update` removes them and lists them in the update report. → REQ-AMI-014, design D8/D14.
  - Q3 — extra scope included: remove `workflow_agents`, `model_routing`, `model_routing_profiles`, `performance_tier`, and dynamic-workflow `agent()` model/effort values. → REQ-AMI-007, REQ-AMI-013, design D12.
  - Q4 — `init`/`update --profile`: keep the flag, no-op, emit a deprecation warning. → REQ-AMI-016, design D10; extended by measurement to `--model-policy` / `--high` / `--medium-alias` / `--low` (design D13).
  - Q5 — harness v4 manifest specialist `model`/`effort`: optional; `/moai:harness` stops emitting them on generation; existing manifests still parse. → REQ-AMI-003, REQ-AMI-006, design D11.
  - Q6 — docs-site: in scope as M8 of this SPEC. → REQ-AMI-025.
  - Q7 — RESOLVED 2026-09-26 (operator, lane window, relayed by the coordinator): the init/update wizard "agent model policy" question is DELETED; the main-session policy stays in `moai profile setup`. → REQ-AMI-015, design D13.
- plan_status: audit-ready — plan phase closed 2026-09-26 on plan-audit iter-5 PASS-WITH-DEBT.
- plan-audit verdict history: iter-1 FAIL 0.76 → iter-2 FAIL 0.82 → iter-3 FAIL 0.83 → iter-4 FAIL 0.86 → iter-5 **PASS-WITH-DEBT 0.90** (Tier L threshold 0.85). Iterations 4 and 5 ran beyond the Tier L 3-iteration cap on explicit operator approval.
- plan-audit iter-5: PASS-WITH-DEBT 0.90 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-5.md`, commit `b46ea465d`). No SPEC body edit after this verdict; the debt below is carried into run entry.
- plan-audit iter-4: FAIL 0.86 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-4.md`, commit `4cc702dcf`) — blocked by V1/V2; V1, V2 and V5 addressed in commit `077c4f8f6`.
- Run-entry debt (resolve at M0, before the first edit commit; record resolution in §E.2):
  - W1 — `internal/cli/init_test.go:419` `TestValidateInitFlags_InvalidProfile` and the invalid-value half of the `ModelPolicyVocabulary` test at `:476` assert rejection of an invalid `--profile` / `--model-policy` value, which design D10 retires (the value is no longer validated); both will fail at M4. At run entry add `internal/cli/init_test.go` to the touch set (`.moai/reports/t1246/touch-set.sh`) and to the plan §G test-file table under M4 (adapt: expect the deprecation warning and exit 0).
  - W2 — `research.md:L314` states the host-(b) version-match predicate without the `verr == nil` conjunct; the authoritative form is `verr == nil && packageVersion == projectVersion && !forceUpdate` (update_template_sync.go:771-773, as design D14 states). Implement against D14.
  - W3 — `plan.md:L93` cites `wizard/translations.go:89-91` / `:179-181` / `:269-271` for the init-wizard question translations; those ranges stop at the opening lines — each `"model_policy": {` block runs to its closing brace. Remove the whole block per locale.
  - Optional carry-over (plan-audit iter-4): V3 — dead agentfm selectors in `internal/web/assets/app.js:390-480` (`wireProfileMatrix`, `reapplyHaikuLocks`, `wireHaikuEffortLock`) are residue once M2 removes the panel; remove them in M2 if in scope, otherwise note in §E.2. V4 — AC-AMI-006's "no finding for `user-pinned.md`" depends on fixture fields (LR-05 warns on a missing `isolation:`); give the fixture the fields that avoid LR-05, or read the clause as "no LR-03/LR-12/LR-13 finding and no error".
- plan-audit iter-3: FAIL 0.83 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-3.md`, commit `d8261313c`). The operator explicitly approved ONE extra revision round beyond the Tier L 3-iteration cap (2026-09-26, lane window, relayed by the coordinator). R1–R9 addressed in spec v0.5.0.
- plan-audit iter-2: FAIL 0.82 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-2.md`, commit `c87803153`). N1–N9 addressed in spec v0.4.0.
- plan-audit iter-1: FAIL 0.76 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-1.md`, commit `5faf93bb8`). D1–D16 addressed in spec v0.3.0 (see spec.md HISTORY).
- Counts: 25 requirements (REQ-AMI-001..025, contiguous), 25 acceptance criteria (AC-AMI-001..025) — both at the Tier L ceiling of 25.

## §E.2 Run-phase Evidence

### M0 — run entry and baseline (2026-09-27, measured on HEAD `5509ea71e`)

Tree: branch `WT-agent-model-inherit`, HEAD `5509ea71e` = merge of `dc8d18bcd` (plan close) and
develop `7fe658815` (the absorbed develop SHA). No file edited before these measurements; raw
outputs under `.moai/state/verify/t1246/` (local, not committed).

Run gate (REQ-AMI-001 / AC-AMI-001):

| Command | Output |
|---|---|
| `git merge-base --is-ancestor WT-rules-diet develop; echo $?` | `rules-diet-ancestor exit=0` (`WT-rules-diet` = `8fb81c948`, develop = `7fe658815`) |

Overlap gate (REQ-AMI-002 / AC-AMI-002):

| Command | Output |
|---|---|
| `sh .moai/reports/t1246/touch-set.sh > .moai/state/verify/t1246/touch-set.now.txt` | exit 0, **272** paths; `LC_ALL=C diff` against the committed `touch-set.txt` → exit 0 (byte-identical); `LC_ALL=C sort -c` → exit 0 |
| `git merge-base HEAD WT-role-naming-docs` | `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c` (`WT-role-naming-docs` tip `637513578`) |
| `git diff --name-only b59a5d69c… WT-role-naming-docs \| LC_ALL=C sort` | 40 paths: 34 under `.moai/reports/t1257/`, 6 under `.moai/specs/SPEC-ROLE-NAMING-DOCS-001/` |
| `LC_ALL=C comm -12 t1257.txt touch-set.now.txt \| wc -l` | **0** — gate open |

Touch-set distribution (`cut -d/ -f1-2 | LC_ALL=C sort | uniq -c`): 22 `.claude/agents`,
2 `.claude/commands`, 1 `.claude/hooks`, 12 `.claude/rules`, 8 `.claude/skills`, 5 `.claude/workflows`,
1 `.moai/docs`, 2 `.moai/project`, 1 `CHANGELOG.md`, 62 `docs-site/content`, 38 `internal/cli`,
12 `internal/config`, 10 `internal/harness`, 6 `internal/hook`, 5 `internal/settings`, 2 `internal/spec`,
54 `internal/template`, 29 `internal/web` — identical to the plan-time distribution (research.md §I).

Re-measured inventory (research.md §B–§K), this tree:

| Item | Command (abridged) | Now | Plan-time |
|---|---|---|---|
| local moai agent frontmatter | `grep -rlE / -rnE '^(model\|effort):' .claude/agents/moai` | 12 files / 24 lines | 12 / 24 |
| local harness agent frontmatter | same over `.claude/agents/harness` | 10 / 20 | 10 / 20 |
| template agent frontmatter | same over `internal/template/templates/.claude/agents` | 12 / 24 | 12 / 24 |
| codex toml `model` / `model_reasoning_effort` lines | `grep -cE '^model ' / '^model_reasoning_effort'` over the 12 toml | 0 / 12 | 0 / 12 |
| `ResolveAgentModelEffort` files (non-test / test, definer included) | `grep -rlF … --include='*.go' internal cmd pkg` | 9 / 9 | 8 consumers + definer |
| `ResolveHarnessAgentModelEffort` | same | 3 / 2 | 3 |
| `DefaultProfileMatrix` | same | 4 / 7 | 3 + definer |
| `ProfileMatrixAgents` | same | 7 / 4 | 6 + definer |
| `AgentGroup(` | same | 3 / 0 | 2 + definer |
| `ApplyProfile` | same | 5 / 1 | 4 + definer |
| `EffectiveProfile` | same | 6 / 1 | 5 + definer |
| `ResolveGLMReasoningForModel` | same | 3 / 2 | 2 + definer |
| `retainedAgentNames` | same | 2 / 1 | profile.go + registry.go |
| guard surface `.go` files | `grep -rlE 'agent_model_guard\|AgentModelGuard\|agent-model-audit' --include='*.go'` | 11 | 11 (research §C list) |
| `agent_model_guard` in template YAML | `grep -rnE … --include='*.yaml' internal/template/templates` | 0 | 0 |
| resolver/guard test pattern | `grep -rlE 'ProfileMatrix\|ResolveAgentModelEffort\|agent_overrides\|AgentModelGuard\|agentModel\|profile_matrix' --include='*_test.go'` | 29 | 29 |
| `CLAUDE_CODE_SUBAGENT_MODEL` non-test hits | `grep -rnE … internal cmd pkg \| grep -v _test.go` | 1 (scrub list) | 1 |
| AC-AMI-007(d) `judge_effort` files | `grep -rlE 'judge_effort\|JUDGE_EFFORT' …` | 3 | 3 |
| AC-AMI-007(c) prose hits | research §F regex | 6 | 6 |
| docs-site pages (AC-AMI-025 pattern) | research §F narrowed pattern | 52 | 52 |
| `[HARD]` lines naming effort/model | `grep -rnE '\[HARD\].*(effort\|model)' .claude/rules` | 17 | 17 |

Always-loaded budget baseline (REQ-AMI-024 / AC-AMI-024 "before"):

```
$ unset MOAI_KANBAN … && go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v
    token_budget_guard_test.go:70: always-loaded surface = 65591 tokens (budget 77600, headroom 12009, 16 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.152s
```

Before-headroom re-baselined after t1175: **12009** (plan-time 61 at `d6992e3a0`). The test lives in
`internal/config`; `go test ./internal/template/ -run TestAlwaysLoadedTokenBudget` reports
`[no tests to run]`.

Run-entry debt resolution:

- **W1 — resolved by record.** `internal/cli/init_test.go` is added to the run touch set as an
  explicit addendum (touch set for this run = the 272 generated paths + `internal/cli/init_test.go`
  = 273; `grep -c 'internal/cli/init_test.go' touch-set.now.txt` → 0 before the addendum). The
  generator script and plan.md §G are plan artifacts and are not edited in run; the M4 test-file
  row is carried here instead: at M4, `TestValidateInitFlags_InvalidProfile` (:419), the invalid
  half of `TestValidateInitFlags_ModelPolicyVocabulary` (:451) and `TestInitCmd_ProfilePersistence`
  (:490, asserts `profile:` persisted) are adapted to expect the deprecation warning, exit 0, and no
  `profile:` / `performance_tier:` write (D10/D13).
- **W2** — carried to M1 host (b): implemented against design D14's
  `verr == nil && packageVersion == projectVersion && !forceUpdate`.
- **W3** — carried to M4 (whole `"model_policy": {…}` block per locale).
- **V3 / V4** — carried to M2 / the AC-AMI-006 fixture.

### M1 — retention seams, characterisation, strip step (2026-09-27)

Commits: `152ba5545` (characterisation, test-only), `dc51913ee` (roster SSOT), `267a484ba` (D5),
`0d19b6107` (strip step). Raw outputs under `.moai/state/verify/t1246/` (local).

| Unit | RED evidence | GREEN evidence |
|---|---|---|
| Characterisation (REQ-AMI-012/017/018/019) | n/a — pins current behaviour; recorded before any removal | `TestCharacterize_{LaunchEffortFromPreferenceProfile,GLMAliasMapping,GLMSessionReasoning,AuditPinPrecedenceAndBackendDefault}` (cli) and `TestCharacterize_{PreferenceProfileCreateRenameDelete,MainSessionEffortSaveWritesPreferencesOnly}` (web): PASS |
| Roster SSOT (REQ-AMI-020, D4) | `undefined: RetainedAgents` (build failed); rosterguard sweep `undeclared roster listing: internal/template/retained_agents.go` | `TestRetainedAgents_IsTheModelFreeRosterSSOT` PASS; sweep failure gone; registry row `retained-agent-roster` added, `profile-matrix-order` asserted against the SSOT |
| D5 pin > backend default (REQ-AMI-019) | `codex task = {Model:gpt-5-codex Effort:high}, want the zero value`; `glm task default = "glm-4.6", want "glm-5.3-flash"` | `Test{Codex,GLM}Resolution_IgnoresPerAgentLLMCells` PASS; `ResolveAgentModelEffort` gone from `cli/{glm_task,mcp_codex,mcp_glm}.go` and their tests |
| Strip step (REQ-AMI-014, D14 a/b/c) | `undefined: StripRetiredModelKeys` / `undefined: stripRetiredModelConfig … confirmViaPreviewFn`; mutation replays: flow-mapping fix removed → `stripped llm.yaml no longer parses`; shipped gate removed → `RetiredModelKeysPresent with every key shipped = true` | 10 template tests + 7 cli tests PASS (report once as removed, backup holds originals, version-match own backup, cancel byte-identical via `confirmViaPreviewFn`, host placement, clean-reinstall filtered advisory) |

Design deviation recorded for manager-spec: the strip leaves a retired key in place while the
embedded template still ships it (`template.ShippedRetiredModelKeys`). Without that gate, every
update would strip `profile`/`agent_overrides`/… that code in the same build still reads and
that the next deploy re-adds; the existing llm-preserve tests (AC-LCP-001/002/005) fail on it.
Today only `workflow.agent_model_guard` (never shipped in YAML) is stripped; the rest start
stripping automatically when M5 removes them from the template.

Package verification (env-scrubbed, one compound call each):

| Command | Result |
|---|---|
| `go test ./internal/template/... -count=1` | exit 0 (3 packages ok) |
| `go test ./internal/web/... -count=1 -timeout 25m` | exit 0 |
| `go test ./internal/cli/... -count=1 -timeout 25m` | exit 1 — only `TestCodexAuditMCPTool` and `TestMCPToolCatalogueFiguresMatchRegistry`, both reading `moai-mcp-tools.md` (not touched by this card; pre-existing from the absorbed develop) |
| `go test ./internal/harness/rosterguard/` | exit 1 — `TestNumeralResidualArithmeticCloses`, `TestNumeralBreadthSetEqualsTheDeclaredUnion`, `TestRegisteredSitesMatchTheirDeclaredAxis` (CLAUDE.md §4 CountPattern); identical failing set on HEAD before the roster edit — pre-existing |
| `golangci-lint run` (v2.1.6) on cli, template, rosterguard, web | `0 issues.` |
| `go build ./...` | ok |
| `git diff --name-only 5509ea71e HEAD -- internal/hook internal/config` | 0 files — t1282-frozen surfaces untouched |

### M2 — web console removal (2026-09-27, commit `384eb3460`)

RED (before any removal), `go test ./internal/web/ -run 'TestAgentSettings(Tab_IsNotRendered|Fields_ArePostedWithoutEffect)'`:
`GET /settings still renders "sec.agentfm.title"`, `… "name=\"performance_tier\""`,
`… "id=\"moai-profile-matrix\""`, and `llm.yaml changed:` after a POST carrying
`performance_tier` + `agentfm.manager-develop.{model,effort}`. GREEN: both PASS.

Removed: `internal/web/agentfm.go`, `internal/settings/agentfm/`, the agentfm templ blocks
(`*_templ.go` regenerated with `go run github.com/a-h/templ/cmd/templ generate -path ./internal/web`,
`templ version` → `v0.3.1020`), app seams `listAgentFMs` / `patchAgentFM` / `applyPerfTierEdits`,
pageView fields `AgentFMs` / `PerfTier*` / `LLM`, the settings tab entry, the v4manifest
tier/model-colour helpers and settings re-exports (`TierForAgent`, `TierSuggestedModelEffort`,
`V4EffortValues`, `V4ModelValues`), the app.js `wireProfileMatrix` / haiku-lock code (V3 resolved),
orphaned CSS, 148 i18n lines (`agentfm.*`, `fieldDesc.agentfm.*`, `agentdesc.*`, `sec.agentfm.*`; count
after = 0), the `agentdesc.` exemption and the empty-registry assertion, and rosterguard sites
`v4manifest-agent-tiers`, `v4manifest-tier-test`, `web-i18n-agent-descriptions`,
`web-agentfm-display-rank(-test)`, exemption `web-agentfm-subset-count`. No dedicated agentfm
route existed; the former fields rode `/save` and are now ignored.

| Command | Result |
|---|---|
| `go test ./internal/web/... -count=1 -timeout 25m -v` | exit 1 — only `TestDocsTabContract` (README ×4 and docs-site `moai-web-console.md` ×4 still list 14 tabs incl. Agents); 509 PASS incl. the M1 web characterisation tests |
| `go test ./internal/settings/... ./internal/harness/v4manifest/...` | ok |
| `go test ./internal/template/... -count=1` | exit 0 |
| `go test ./internal/harness/rosterguard/` | exit 1 — same three tests; failure lines byte-identical to M1 (`diff rg1.txt rg2.txt` exit 0, 7 lines) |
| `go test ./internal/cli/ -run 'TestCharacterize_(…)$'` | 4 PASS |
| `golangci-lint run` (v2.1.6) web, settings, harness | `0 issues.` |
| `go build ./...` / `go vet` on cli, web, settings, harness, template | ok / exit 0 |

Blocker for the lead: `TestDocsTabContract` binds README.md (4 locales, manager-docs territory) and
the docs-site web-console pages (M8) to the rendered tab list; it stays red until those drop
the Agents tab. **Resolved** by operator decision in `961631da3` (Agents entry removed from the
README tab lists ×4 and the numbered list in `advanced/moai-web-console.md` ×4, renumbered 9–13);
`go test ./internal/web/ -run TestDocsTabContract -v` → literals / allowlist / names PASS.

### M3 — agent_model_guard key and deny (lead boundary ① only, commit `d76728a0a`)

RED: `TestAgentModelGuardKey_LoadsAndHasNoEffect` —
`decision: got "deny", want allow fall-through` and `with key "deny|AGENT_MODEL_VIOLATION: Explore
was spawned with model \"haiku\" …"`. GREEN: PASS (the key loads; with/without key decisions equal).

Removed: `config.AgentModelGuardConfig`, `WorkflowConfig.AgentModelGuard`, its default and
`TestDefaultAgentModelGuardDisabled`; `agentModelGuardEnabled`, `SentinelAgentModelViolation`, the deny
branch of `checkAgentModel` (now returns the advisory only) and the deny call site in `pre_tool.go`;
the deny-matrix / gate-on hook tests. Sibling comments reworded (`agent_stop_guard.go`,
`subagent_write_guard.go`, `types.go`). The gitignore-artifact row and rosterguard row name the audit
jsonl (frozen) and stay.

Deferred to post-t1282 (frozen, verdict.md §2 ②): `agentModelAuditFileName`, `appendAgentModelAudit`
and the `.moai/logs/agent-model-audit.jsonl` writes, `prune_logs` retention, `resolveAgentModel`,
`llmConfig`, `classifyAgentModel` / `agentModelAdvisory` / `extractAgentSpawn` (the observation layer
that feeds the frozen writes), the rest of `agent_model_guard.go` + its observation tests, and the
audit-receipt path in `pre_tool.go` (unchanged; its line moved from :724 to :719 because 5 lines above
it were removed). REQ-AMI-009 (no observation, no audit file) is therefore not yet met.

| Command | Result |
|---|---|
| `go test ./internal/hook/... -count=1 -timeout 25m` | exit 0 (11 packages ok) |
| `go test ./internal/config/... -count=1` | exit 0 |
| `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` | `always-loaded surface = 65591 tokens (budget 77600, headroom 12009, 16 entries)` PASS |
| `go test ./internal/web/... ./internal/template/... -count=1` | exit 0 (web now fully green) |
| `golangci-lint run` (v2.1.6) hook, config | `0 issues.` |
| `go build ./...` | ok |

### M4 — CLI, lint and guard consumers (commit `41cf11c4d`)

RED before any change:
- `TestModelCommandIsRemoved`: `` `moai model` is still registered (Model-routing profile inspection) ``.
- `TestInitDeprecatedModelFlags_WarnAndAcceptAnyValue`: no warning for any flag, and
  `--profile=bogus must be accepted, got: invalid --profile value "bogus": …`.
- `TestUpdateDeprecatedProfileFlag_WarnsAndIsAccepted`: no warning.
- `TestInitDeprecatedModelFlags_WriteNothing`: `init with map[profile:high] wrote a different llm.yaml…` (also model-policy, high).
- `TestWizardsDoNotAskTheAgentModelPolicy`: `the default wizard still asks model_policy` (also reconfigure).
- `TestEffortRules_LR03AndLR12Retired`: LR-03 and LR-12 fired.
- `TestWorkflowLint_ModelRoutingBlockIsNoLongerChecked`: `lint violations detected`.
- `TestHaikuResidualRule_RoutingSurfacesRetired`: routing surfaces still yield findings.
All GREEN after the change.

Removed / changed: `moai model` (`model.go`, `model_test.go`, root registration); retired flags now
warn to stderr (`Warning: --<flag> is deprecated and has no effect: subagents now inherit the main
session's model and effort. Set the main-session model policy with `moai profile setup`.`), accept any
value and write nothing (`resolveModelPolicy`, the init ApplyPerformanceTier/ApplyProfile block,
`applyUpdateProfile` and both update call sites removed; `InitOptions.{ModelPolicy,Profile}` removed);
wizard `model_policy` question, capture branch, ko/ja/zh blocks and `WizardResult.ModelPolicy`; the
update-wizard ApplyProfile + system.yaml `model_policy` write; H24 wording (20 lines, 4 locales);
agentlint LR-03/LR-12 + `canonicalEffortMatrix`; workflow-lint routing check + `SentinelModelRoutingInvalid`
(the command keeps its parse/exit-code contract); HaikuResidual surfaces 3–4; `internal/harness/cellguard`;
`normalizeLLMSectionMaps` retired-map lines; MCP tool-description resolver citations. The
`agentlint-section-marker` numeral exemption stays (its comment survives).

W1 resolved: `TestValidateInitFlags_InvalidProfile`, the invalid half of
`TestValidateInitFlags_ModelPolicyVocabulary`, and `TestInitCmd_ProfilePersistence` now expect the
warning, a nil error, and no `profile: high` write.

| Command | Result |
|---|---|
| `go test ./internal/cli/... -count=1 -timeout 25m` (slot `internal-cli-suite`) | 17 packages ok; `internal/cli` FAIL on `TestCodexAuditMCPTool`, `TestMCPToolCatalogueFiguresMatchRegistry` (inherited), `TestProfileWizardGolden_LocaleFrames` (H24 wording → goldens regenerated with `-update-golden`, now PASS), `TestCodexTaskBackgroundHandshakeHonorsTaskBound` (100 ms handshake race: 1/3 then 0/10 fails on re-run; code path untouched) |
| `go test ./internal/cli/agentlint/ ./internal/cli/wizard/... ./internal/spec/ -count=1` | all ok |
| `go test ./internal/harness/rosterguard/ -v` | the same three failures; failure lines identical to M1 (`diff rg1.txt rg4.txt` exit 0) |
| `golangci-lint run` (v2.1.6) cli, spec, core | `0 issues.` |
| `go build ./...` | ok |

### M6 — agent frontmatter, Codex emission, harness manifests, workflow scripts (commits `ee77b7d4d`, `ddd9ad83a`)

RED: `TestAgentsDeclareNoModelOrEffort` (replaces the haiku-effort guard) — 34 files reported
`AGENT_MODEL_EFFORT_DECLARED` (12 template + 12 local moai + 10 local harness);
`TestValidate_SpecialistModelAndEffortAreOptional` — `specialists[0].effort "" is not low|medium|high|xhigh|max`.
Both GREEN.

- Commit 1 (template): 24 frontmatter lines stripped from the 12 C2 agents; `agents-codex.yaml`
  `model_reasoning_effort.emit: false`, class row `effort` → `disposition: omit` ("Omitted — inherit the
  parent"); AC-008 emitter tests inverted; `make agents-emit` exit 0 (12 toml, −12 lines);
  `make agents-emit-check` exit 0 (`ok … internal/template/agentemit`); `make build` exit 0 (catalog.yaml hashes).
- Commit 2 (local + harness + scripts): 44 frontmatter lines stripped (12 moai + 10 harness);
  v4manifest model/effort optional (`omitempty`, present values still checked); local manifests drop
  them; `agent()` effort options removed from template + local `codemaps-extract.js`,
  `plan-research-fanout.js`, `sync-audit-4dim.js` and the local Runners (7 lines); local
  `judge_effort` / `JUDGE_EFFORT` channel and `test-judge-effort-contract.sh` removed; rosterguard
  `agentemit-golden` site removed (its 12-name effort map is gone) and the numeral live-candidate line
  re-measured (52 → 50).

| Command | Result |
|---|---|
| `make build` (after template script edits) / `make embed-check` | exit 0 / exit 0, `Agent Emit Embed 12/12 embedded agent-emit artifacts match the committed set` |
| `go test ./internal/template/... ./internal/cli/agentlint/... -count=1` | exit 0 |
| `go test ./internal/harness/... -count=1` | all ok except rosterguard: the same three; failure lines identical to M1 (`diff rg1.txt rg6.txt` exit 0) |
| AC-AMI-003(a) `grep -rnE '^(model\|effort):' .claude/agents internal/template/templates/.claude/agents` | no output (exit 1) |
| AC-AMI-004 `grep -nE '^(model\|model_reasoning_effort)' …/.codex/agents/moai/*.toml` | no output (exit 1) |
| AC-AMI-007(b) scripts grep | no output (exit 1) |
| AC-AMI-005 `bin/moai agent lint` | exit 0, 0 LR-03/LR-12 findings (25 LR-08 warnings, pre-existing kind) |
| `golangci-lint run` (v2.1.6) template, v4manifest, rosterguard | `0 issues.` |

Deferred to M7 (doctrine): `.claude/rules/moai/workflow/verify-judge-effort-contract.md` (H22 rule file —
the last `judge_effort` hit for AC-AMI-007(d)); the `/moai:harness` generation instructions
(harness-builder / harness-build-entry / builder-harness body, H12 — AC-AMI-003 second grep).

Deferred to post-t1282: `internal/hook/agent_model_guard.go` still calls
`template.ResolveAgentModelEffort` (`resolveAgentModel`, frozen) — the last non-test consumer
outside `internal/config` / `internal/template`, so M5 cannot delete the resolver until it moves.

### M7 — doctrine (design §D H1–H20, H22, H23; H24b wording) (commits `238219302`, `d8d164d90`, `29081aaf3`)

Template first, then local (commit 1 template, commit 2 local + Go + tests, commit 3 label restore).

- H1/H2: `agent-common-protocol.md` § Per-Spawn Model Injection → "### Subagent Model and Effort" (one
  plain sentence; the PreToolUse hook is described neutrally as an observation log that never blocks —
  the `.moai/logs/agent-model-audit.jsonl` surface stays until t1282); reference rationale removed.
- H3–H6 + D9: `model-policy.md` rewritten to the inheritance rule; § Model Policy Tiers → § Model
  Policy (main session); § Per-Agent Profile Resolver removed; § Harness-Agent Model Policy reduced to
  the no-field rule; per-agent effort paragraph → session effort. Rosterguard rows
  `model-policy-profile-matrix-size` + `-mirror` removed with it.
- H7/H8/H19/H23: agent-authoring § Effort-Level Calibration Matrix removed (its general effort-default
  paragraph moved into the `effort` field note); constitution pointer removed; constitution-detail
  pointer reworded; `model: "haiku"` dropped from the team spawn example; prompt-craft line and the
  per-spawn cost-axis cross-reference rewritten/removed.
- H9/H10 cache-aware d5/d10; H11 agent-patterns (history kept, "no field" line added; the
  canonical per-spawn example also loses `model: "opus"`); H12 harness-builder / harness-build-entry
  (`Agent(model/effort)` arguments, the cost-leak sentence, per-specialist model/effort assignment
  removed); H13 token-optimization; H14 sub-agents reference (measured resolution order); H15 agent
  bodies' escalation line (all 12 agents, both trees) + builder-harness generation policy +
  manager-design / super-advisor model text; H16 agent-lint LR-03/LR-12 rows retired; H17
  dynamic-workflows purpose taxonomy section replaced by one plain sentence; H18 settings-management;
  H20 archived-agent-rejection `model:` arguments; H22 `verify-judge-effort-contract.md` deleted.
- H24b (wording): `hint.effort.go_unbound` (templ + i18n.js, 4 locales) now states the backend-only
  scope; `EffortLevelXHigh` labels (4 locales) drop the agent-matrix clause; profile-wizard goldens
  regenerated (ko/ja/zh). The `ModelPolicyHigh/Medium/Low` labels were rewritten and then restored:
  `TestModelPolicyLabels_AgreeWithProfileMatrix` and `TestGetProfileText_OpusAliasValues` derive the
  expected label text from `template.DefaultProfileMatrix`, so the labels move with the matrix (M5,
  after t1282).
- agentlint doc tests adapted: `TestAuthoringDocHasEffortMatrix` → `TestAuthoringDocHasNoEffortMatrix`,
  `TestConstitutionCrossReference` → `TestConstitutionHasNoPerAgentEffortPointer`.

Design §D row note (not edited in design.md): AC-AMI-007(a) had four residual hits outside §D —
maintainer-owned `.claude/agents/harness/{hook-ci,cli-template}-specialist.md` and
`.claude/skills/hns-moaiadk-{patterns,best-practices}/SKILL.md` quoted the archived-agent `§C`
per-spawn pattern with `model: opus`. Edited under REQ-AMI-007 so they match the H20 row they cite.

| Command | Result |
|---|---|
| `go test ./internal/config/ -run TestAlwaysLoadedTokenBudget -v` before | `always-loaded surface = 65591 tokens (budget 77600, headroom 12009, 16 entries)` PASS |
| same, after | `always-loaded surface = 65307 tokens (budget 77600, headroom 12293, 16 entries)` PASS |
| `go test ./internal/template/ -count=1` (neutrality, internal-content-leak, rule mirror, catalog) | `ok … internal/template 112.639s` |
| `go test ./internal/cli/agentlint/ ./internal/web/ ./internal/harness/... ./internal/template/agentemit/...` | all ok except rosterguard |
| `go test ./internal/harness/rosterguard/` | the same three failures; failure lines equal M6's `rg6.txt` after sorting (`sort … \| diff` exit 0) |
| `make agents-emit` / `make agents-emit-check` / `make build` / `go build ./...` | exit 0 / exit 0 / exit 0 / exit 0 |
| `golangci-lint run` (v2.1.6) agentlint, web, rosterguard | `0 issues.` |
| AC-AMI-007 (a)(b)(c)(d) | no output each |
| HARD markers (`grep -c '\[HARD\]'`, both trees) | agent-common-protocol 16→15, dynamic-workflows 1→0, reference 5→4, all others unchanged |
| zone-registry check | no test or command found for it |
| `go test ./internal/cli/ -count=1 -timeout 40m` (slot `internal-cli-suite`; a second session ran cli tests concurrently, load ~25) | FAIL, 1630.772s: 14 failures = inherited `TestCodexAuditMCPTool`, `TestMCPToolCatalogueFiguresMatchRegistry`, flaky `TestCodexTaskBackgroundHandshakeHonorsTaskBound`; 9 codex-launch tests (M6 regression, below); 2 label tests from the first H24b draft (fixed by restoring the model-policy labels) |
| after the restore: `go test ./internal/cli/ -run TestModelPolicyLabels_AgreeWithProfileMatrix` / `-run TestGetProfileText` / `-run Profile` | PASS / PASS / `ok … 22.069s` |
| `golangci-lint run ./internal/cli/` / `GOOS=windows GOARCH=amd64 go build ./...` | `0 issues.` / exit 0 |

AC-AMI-021 note: the reference file's −1 is not a marker — the removed H2 section's relocation note
contained the literal text "The [HARD] rule and the four operative bullets", which `grep -c` counts.

Found, not fixed (outside M7): the M6 codex emission change (`model_reasoning_effort.emit: false`)
broke `internal/cli/codex_audit_launch.go`, which requires `model_reasoning_effort` in each role toml
(`role "sync-auditor" file carries no usable model_reasoning_effort`). Nine `TestCodexAuditLaunch*` /
`TestCodexAuditVerbRunsInCallerWorktree` tests fail. The SPEC inventory (research §F) does not list
this consumer.

Deferred to post-t1282: the frozen hook (`resolveAgentModel`, the audit log) is described neutrally
in doctrine; no text claims a deny or opt-in exists or that the log is gone.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — see §E.2 succession block for the current state>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

### Succession session — develop absorption + stranded-consumer repairs (2026-09-28, worker-66; commits `3b29dfcea`·`b822c7d00`·`9eb125a73`)

**Absorption**: develop 404 commits absorbed (merge `3b29dfcea`), four conflict families resolved with documented rationale — ① `AgentModelGuardConfig` dropped (M3 removed the block layer; the hook file's own header says the opt-in key is gone and a leftover key is ignored) while `ServedModelGateConfig` is kept (develop/t1282 addition with live consumers: `internal/hook/served_model.go`, `internal/auditreceipt/store.go`, `internal/config/served_model_gate_test.go`); ② i18n.js `agentdesc.*` deleted with the card's M2 panel removal (develop's t1256 additions to a deleted surface are dead); ③ catalog.yaml hash conflicts regenerated via `go run ./internal/template/scripts/gen-catalog-hashes.go --all`; ④ no other content conflicts.

**Stranded consumer 1 — the audit launcher (`b822c7d00`)**: M6's emission change (no `model_reasoning_effort` in role TOMLs) left `codex_audit_launch.go` as the sole remaining reader, requiring the key — 8 `TestCodexAuditLaunch*` red. Fixed per the card design: the launcher reads effort from the audit pin (`workflow.audit.codex.effort`), an unpinned launch **omits** the `-c model_reasoning_effort` directive rather than fabricating a value, and an unlaunchable pin refuses before spawning. Tests: the Argv fixture writes the pin and asserts the token; the argv judge treats an unpinned launch as must-be-absent; the ceiling fixture's handwritten key is inert by design.

**Stranded consumer 2 — the explicit model argument (`9eb125a73`)**: t1284's served-model tests (absorbed from develop) failed because the card's `resolveCodexModelEffort` rewrite (explicit-only) left `handleCodexTask` building `turnParams` without the caller's `model` argument — the deleted SSOT cell path had been the only source. Fixed by piping a non-empty explicit model into `turnParams` (feeding both thread/start and `result.Model`); the three configured-model subtests now pass the model explicitly, with the llm.yaml fixture retained as a leftover-cell control proving the dead cell does not leak.

**Verification (this tree, this session)**:

```
$ go build ./...                                              → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                    → exit 0
$ go test ./internal/cli/ -run 'ModelPolicy|Profile|AgentModel|ServedModel|Init|CodexAuditLaunch' -count=1 -timeout 20m
ok  github.com/modu-ai/moai-adk/internal/cli  95.274s         (slot-leased)
$ go test ./internal/config/ ./internal/harness/rosterguard/ ./internal/web/ -count=1
ok ×3 (4.160s / 18.171s / 29.423s)
$ go test ./internal/template/... -count=1                     → ok ×3
$ go test ./internal/hook/ -count=1 -timeout 15m
ok  github.com/modu-ai/moai-adk/internal/hook  343.026s
$ golangci-lint run ./internal/cli/                           → 0 issues.
```

Tests the M7 close recorded as red, all now green after absorption + the two fixes: `TestCodexAuditMCPTool`, `TestMCPToolCatalogueFiguresMatchRegistry` (develop's 404 commits), `TestCodexTaskBackgroundHandshakeHonorsTaskBound` (t1288), rosterguard ×3, `TestCodexAuditLaunch*` ×8, `TestCodexTask_ServedModelUnknown` ×3-subtests.

**M5 blocker — unchanged after absorption**: `internal/hook/agent_model_guard.go:106` still calls `template.ResolveAgentModelEffort` (the observation layer's expected-model source), so M5 (resolver/matrix deletion) remains blocked on that call site's rework — recorded as "deferred to post-t1282" at M6 close, and t1282's landing did not remove it. **M8 (docs-site) not started.** Both are the card's remaining run-phase work; §E.3 stays pending until they land.

## §F Phase 4 Mode Selection (worker-71 continuation, 2026-09-28)

- **Input parameters**: tier L continuation (M5 only); scope ~15 files, mostly deletions; domains = Go tests + template YAML + rosterguard registry + project docs; coding-heavy; concurrency benefit LOW; one uncommitted predecessor tree inherited (worker-66, build-verified).
- **Mode evaluation**: direct — no (multi-file semantic edits) · serial — **selected** · fanout — no (coding-heavy; one writer per tree) · sweep — no (not mechanical-uniform, has inter-file deps).
- **Decision: serial** — one manager-develop spawn completes M5 on top of the inherited uncommitted work.
- **Justification**: continuation of an in-flight uncommitted tree admits exactly one writer; coding-heavy per Anthropic caveat; the lead's prescription (declared-model logging + producer/schema deletion) is already half-landed in the tree, so fan-out would split one coherent change.
- **Gate**: Implementation Kickoff Approval granted by the operator in the lane window, 2026-09-28 (AskUserQuestion, this session).

### M5 completion (worker-66 → committed 3fa8bd2ab, 2026-09-28)

**Producer/schema deletion landed as one set with the observation-layer switch** (declared/inherit
verdicts in agent_model_guard.go, declaration-only served-model expectations). 39 files,
+319/−4213. Deletions: config profile.go + model_routing.go + 3 dead test files;
template profile_matrix.go + embedded_llm_yaml_test.go; model_policy.go perf-tier helpers;
orphaned per-agent GLM helpers (coding-max override set, ResolveGLMReasoning*); ModelEffort
relocated beside audit_models.go (audit pins keep it). Templates: llm.yaml drops
profile/performance_tier/profiles/harness_agents/agent_overrides; workflow.yaml drops
workflow_agents/model_routing/model_routing_profiles; shipped_key_inventory regenerated
977→795; NFR-CKH-002 floor re-derived 875→700 against the measured 730-key surface.
Rosterguard: 8 M5 sites removed, new retained-agents-test-expectations site registered
(the guard itself caught the unregistered roster literal — positive control observed);
axis.go doc re-anchored. product.md/tech.md rewritten.

**§E.2 evidence (this run, this tree 3fa8bd2ab)**:
- `go build ./...` → exit 0 (no output) · `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet` config/template/settings/harness/cli → clean
- `golangci-lint run --timeout=2m` (v2.1.6, CI version) config/template/harness/settings → `0 issues.`
- `go test ./internal/config/ ./internal/template/ -count=1` → both `ok`
- `go test ./internal/hook/ -run 'ServedModel|AgentModel|Validate|Section|LLM' -count=1` → `ok 5.003s`
- `go test ./internal/cli/ -run 'ModelPolicy|Profile|ServedModel|AgentModel|CodexAuditLaunch|CodexTask|GlmModel|InitQuietWizard|ProfileSetupSchema' -count=1` → `ok 36.466s`, re-run → `ok 21.378s`
- `go test ./internal/harness/rosterguard/ -count=1` → `ok 18.540s`

**Gaps**: full `internal/hook` suite re-run still executing in background at commit time
(targeted observation-layer family already green; pre-M5 full hook suite was green at
343.026s). Full `internal/cli` suite not re-run locally — CI on develop push owns it.
**Not in this commit**: M8 (docs-site) — separate card per lead instruction.
