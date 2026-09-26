---
id: SPEC-AGENT-MODEL-INHERIT-001
title: "Research — measured inventory for subagent model/effort inheritance"
created: 2026-09-26
---

# Research — SPEC-AGENT-MODEL-INHERIT-001

Every count in this file was measured in the card worktree `.claude/worktrees/t1246`
(branch `WT-agent-model-inherit`, base develop `d6992e3a0`) on 2026-09-26. Nothing is carried
over from memory or from another tree. Run-phase M0 re-measures every count before any edit
(REQ-AMI-001), because develop moves and card t1175 lands first.

## §A. The premise under the whole card: does "no field" mean "inherit"?

The card's intent is that removing `model:` / `effort:` makes every subagent take the main
session's model and effort. That is only true if Claude Code resolves an absent field to the
main session. It was measured, not assumed:

- Local reference `.claude/skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md:79,164`
  says: "If omitted, uses configured default (usually sonnet)." — **stale**.
- Official doc `https://code.claude.com/docs/en/sub-agents` (fetched 2026-09-26) states the
  resolution order: (1) per-invocation `model` parameter, (2) the definition's `model`
  frontmatter (`inherit` = main model), (3) `CLAUDE_CODE_SUBAGENT_MODEL` env, (4) the main
  conversation's model. For `effort`: "Default: inherits from session".
- `grep -rnE 'CLAUDE_CODE_SUBAGENT_MODEL' internal cmd pkg .claude/settings.json internal/template/templates/.claude/settings.json.tmpl`
  → one hit, `internal/cli/mcp_claude.go:233`, an env **scrub** list entry for the `claude_audit`
  child process. MoAI never sets the variable.

Conclusion: with no frontmatter field and no spawn-time `model`, a subagent lands on step 4
(main model) unless the user has exported `CLAUDE_CODE_SUBAGENT_MODEL` themselves. That residual
is recorded as a documented user-environment exception (spec.md §F), not a MoAI defect. The
stale local reference is itself a correction target (plan.md M7).

## §B. Agent frontmatter inventory (includes files whose only hit is a `model:`/`effort:` line)

Command: `grep -rnE '^(model|effort):' .claude/agents` and
`grep -rnE '^(model|effort):' internal/template/templates/.claude/agents`;
Codex: `grep -nE '^(model|model_reasoning_effort)' internal/template/templates/.codex/agents/moai/*.toml`.

| Copy | Files with hits | Lines | Detail |
|---|---|---|---|
| C1 local `.claude/agents/moai/*.md` | 12 of 12 | 24 | `model: inherit` ×11, `model: sonnet` ×1 (manager-git); effort high ×6, medium ×3, low ×3 |
| C1 local `.claude/agents/harness/*.md` | 10 of 10 | 20 | all `model: opus` + `effort: high` (hns-github, hns-oss-docs-{content-author,locale-translator,structure-curator}, hns-release, hns-release-update, cli-template, hook-ci, quality, workflow) |
| C2 template `internal/template/templates/.claude/agents/moai/*.md` | 12 of 12 | 24 | same values as C1 moai |
| C3 codex `internal/template/templates/.codex/agents/moai/*.toml` | 12 of 12 | 12 | `model_reasoning_effort` only (high ×6, medium ×3, low ×3); **zero** `model =` lines — the manifest already sets `fields.model.emit: false` |
| **Total** | **46 files** | **80 lines** | local total `.md` = 22, template `.md` = 12, toml = 12 (+ directory entries) |

Emitter coupling (measured, `internal/template/agentemit/writer.go:123-131`): with
`fields.model_reasoning_effort.emit: true` an empty `effort` value is looked up in the map, is
not found, and **fails emission** ("refusing to guess"). Stripping C2 effort therefore requires a
manifest change (`agents-codex.yaml` `model_reasoning_effort.emit: false` plus its rationale row
at `:227-232`) in the same milestone, or `make agents-emit` fails closed.

Agent-lint coupling (`internal/cli/agentlint/agent_lint.go:123-125,324-330,558,613-621`):
LR-03 is an **error** on a missing `effort:`; LR-12 rejects drift from the canonical effort
matrix; LR-13 validates the enum. After stripping, LR-03 fires on every agent. LR-03 and LR-12
must retire; LR-13 becomes vacuous for MoAI agents but still validates a user-authored agent that
chooses to declare `effort`.

## §C. Go producers and consumers

Symbol fan-in (`grep -rlF "<sym>" internal cmd pkg | grep -v _test.go`):

| Symbol | Non-test consumers |
|---|---|
| `ResolveAgentModelEffort` | cli/glm_task.go, cli/mcp_codex.go, cli/mcp_glm.go, cli/mcp_server.go (tool description text), cli/model.go, hook/agent_model_guard.go, web/agentfm.go, web/handlers.go (comment) |
| `ResolveHarnessAgentModelEffort` | template/profile_matrix.go, harness/cellguard/cellguard.go, config/types.go (comment) |
| `DefaultProfileMatrix` | cli/agentlint/agent_lint.go, harness/cellguard/cellguard.go, config/types.go |
| `ProfileMatrixAgents` | cli/model.go, harness/rosterguard/{axis,check,numeral,registry}.go, web/agentfm.go |
| `AgentGroup` | cli/model.go, web/agentfm.go |
| `ApplyProfile` | cli/init.go, cli/update.go, cli/update_wizard.go, web/agentfm.go |
| `AgentOverrides` / `HarnessAgents` | cli/glm.go (nil-map normalisation), config/{profile,types,validation}.go, web/{agentfm,app,schemaform}.go |
| `EffectiveProfile` | cli/model.go, config/profile.go, template/profile_matrix.go, web/agentfm.go, web/schemaform.go |
| `ResolveGLMReasoningForModel` (per-agent GLM) | cli/model.go, web/agentfm.go |

Guard surface (`grep -rlE 'agent_model_guard|AgentModelGuard|agent-model-audit' --include=*.go`):
`internal/hook/agent_model_guard.go` (273 lines) + `_test.go`, `internal/hook/pre_tool.go:645`,
`internal/hook/prune_logs.go:59,81`, `internal/config/types.go:477-488,781-788`,
`internal/config/defaults.go:1058`, plus comment-only mentions in `hook/agent_stop_guard.go`
and `hook/subagent_write_guard.go`, and test mentions in `config/defaults_test.go`,
`harness/rosterguard/rosterguard_test.go`, `template/gitignore_local_artifacts_test.go`.
`grep -rnE 'agent_model_guard' --include=*.yaml` → **0 hits in either tree**: the key lives only
as a Go default, so no shipped YAML needs editing for it.

Web surface: `internal/web/agentfm.go` (561 lines), `internal/settings/agentfm/agentfm.go`
(frontmatter read/patch layer), `internal/web/handlers.go:74-93,441-579`, `internal/web/app.go:81-92,154-157`,
`internal/web/settings_shell.go:154-268` (`agentfm` tab), `internal/web/fieldsets.templ`
(29 hits) + generated `fieldsets_templ.go`, `internal/web/root.templ` (1 hit) + `root_templ.go`,
`internal/web/schemaform.go`, `internal/settings/schema_sections.go`,
`internal/harness/v4manifest/schema.go` (`ModelColor`, tier suggestions — consumed by
web/agentfm.go and settings/schema_sections.go only). The `/profile/*` routes in
`internal/web/app.go:198-201` are the **user preference profile**, a different concept; they stay.

CLI: `internal/cli/model.go` (`moai model profile [--json]`, registered at `root.go:268`);
`init --profile` (`init.go:113,367-371,568-571,996-1009`); `update --profile`
(`update.go:49-90,526-574,647-660`); `update_wizard.go:310` writes `llm.profile` from the
main-session model policy via `ApplyProfile(NormalizeToTier(...))`.

Doc-binding guards: `internal/harness/cellguard/cellguard.go` binds docs-site cell tables to
`defaultProfileMatrix`; with the matrix gone the guard has nothing to compare against and
retires. `internal/harness/rosterguard` uses `ProfileMatrixAgents()` as the **canonical
retained roster** — that roster is not model policy and must survive under a new home.

Test files matching the resolver/guard pattern set: **29** (`grep -rln
'ProfileMatrix|ResolveAgentModelEffort|agent_overrides|AgentModelGuard|agentModel|profile_matrix'
--include=*_test.go`), plus frontmatter-reading tests found separately:
`template/haiku_effort_guard_test.go`, `template/agentemit/{agentemit,agentemit_edge,golden}_test.go`,
`web/{agent_settings,agentfm_policy,agentfm_polish,m5_agentfm,webux_followup,console_ux_fix}_test.go`,
`cli/{codex_role_fingerprint,harness_profile_transition,update_mirror_notice}_test.go`,
`spec/lint_haiku_residual_test.go`. The full set is re-measured at M0.

## §D. Retention premises (measured, not assumed)

| Keep | Evidence that it is independent of the removed code |
|---|---|
| Main-session model/effort (`moai cc --model`, `model_policy`, `effort_level`) | `grep -nE 'EffectiveProfile|\.Profile\b|ResolveAgent' internal/cli/launcher.go internal/cli/launch_effort_settings.go` → 0 hits; main-session effort resolves through `resolveLaunchEffort(prefs.EffortLevel, prefs.ModelPolicy)` (`launcher.go:763`, `launch_effort_settings.go:68`). |
| GLM model alias mapping (`llm.glm.models.{high,medium,low,fable}`) | Resolved by `resolveGLMModels` in `cli/glm.go`; no dependency on `profiles`. |
| GLM session reasoning (`SessionGLMReasoningState*`, `CollapseClaudeEffortToGLM*`) | Session-global, consumed by `cli/glm.go:419-463`; `CollapseClaudeEffortToGLM` also by `config/audit_models.go`, `cli/mcp_glm.go`. Only the per-agent helpers (`ResolveGLMReasoning*`, `IsGLMCodingMaxOverrideAgent`, `GLMCodingMaxOverrideAgents`) lose their last consumers. |
| Cross-model audit pins `workflow.audit.{claude,codex,glm}` | These pick the model of a separate process (`claude -p --model`, codex, z.ai HTTP) — not a Claude Code subagent. The pin already outranks the matrix cell. |
| Codex/GLM task & audit model when no pin | Today: pin > matrix cell > backend default. The Go matrix holds only `opus`/`sonnet`; `codexServableModel` (`mcp_codex.go:183`) rejects both, so on the default config the matrix contributes nothing to codex. For GLM, `resolveGLMModelForAgent` returns the matrix model only under a GLM backend. Removing the matrix collapses both to pin > backend default (`glmAuditDefaultModel = config.DefaultGLMHigh`, codex default empty). A user who put a codex/GLM id into `llm.agent_overrides` for the audit key loses that path — documented in migration notes. |
| Retained-agent roster | Needed by rosterguard and `config.retainedAgentNames`; moves to a model-free home. |
| `llm.yaml` decode of leftover keys | `grep -rnE 'KnownFields\(true\)|DisallowUnknown' internal/config internal/cli/*.go internal/settings` → only `cli/mcp_claude_protocol.go:119` (an unrelated protocol decoder). Config decoding is non-strict, so leftover `profile`/`profiles`/… keys in a user's `llm.yaml` are ignored rather than rejected once the struct fields go. |

## §E. Configuration keys

| File | Key | Tree | Disposition |
|---|---|---|---|
| `internal/template/templates/.moai/config/sections/llm.yaml` (274 lines) | `profile`, `profiles` (39 cells), `harness_agents` (21 cells), `agent_overrides`, plus their comment blocks (~lines 26-215) | template | remove |
| same | `performance_tier` | template | open question Q3 (it also feeds `model_routing_profiles`) |
| same | `glm.models`, `glm.effort`, `glm.context_windows`, `claude_bin`, `harness`, `mode`, `team_mode` | template | keep |
| `.moai/config/sections/llm.yaml` | `performance_tier`, `profile`, `profiles`, `harness_agents`, `agent_overrides` (lines 6-187) | **primary checkout only, untracked** (`.gitignore:287`) | operator's own file — not committed by this SPEC; `moai update` migration applies |
| `workflow.yaml` (template :187-240, local :54-130) | `workflow_agents`, `model_routing`, `model_routing_profiles` | both | open question Q3 |
| `workflow.yaml` | `audit.{claude,codex,glm}` | both | keep (not subagents) |
| (Go default only) | `workflow.agent_model_guard.enabled` | neither YAML | remove from Go; leftover user key ignored |

Consumers of `model_routing*`: `config/model_routing.go` — `RouteModelFor` referenced only from a
comment in `config/types.go`; `ValidateModelRoutingProfiles` from `cli/agentlint/workflow_lint.go`.
No runtime reader was found; the orchestrator consumes the block as prose via `model-policy.md`.

## §F. Doctrine and document inventory

Pattern set A (`Per-Spawn Model Injection|moai model profile|agent_model_guard|agent-model-audit|Effort-Level Calibration|Calibration Matrix|profile matrix|Profile Resolver|ResolveAgentModelEffort|agent_overrides|harness_agents|model_reasoning_effort|Inherit the session model|explicit \`model\` argument|spawn-time parameter`):

- Local (10): `.claude/agents/moai/manager-design.md`, `rules/moai/core/{agent-common-protocol,agent-common-protocol-reference,moai-constitution,moai-constitution-detail}.md`, `rules/moai/development/{agent-authoring,model-policy}.md`, `rules/moai/workflow/{cache-aware-execution,cache-aware-execution-reference}.md`, `skills/moai-foundation-core/modules/token-optimization.md`.
- Template (10 mirrors of the above + 12 codex toml carrying the manager-design text + `llm.yaml`).
- `CLAUDE.md` / `AGENTS.md` (root and template): **0 hits** — the card listed CLAUDE.md, but no sentence there carries the removed doctrine.

Spawn-time `Agent(... model:/effort: ...)` directives (18 lines local):
`rules/moai/development/{agent-authoring,model-policy}.md`, `rules/moai/workflow/archived-agent-rejection.md`,
`skills/moai/workflows/{harness-builder,harness-build-entry}.md` (all mirrored in template), local-only
`skills/hns-moaiadk-{best-practices,patterns}/SKILL.md`, and bodies of local-only
`agents/harness/{cli-template,hook-ci}-specialist.md`.

Dynamic-workflow `agent()` literals: local `.claude/workflows/{codemaps-extract,hns-release-update-run,plan-research-fanout,hns-oss-docs-run,sync-audit-4dim}.js`; template `{plan-research-fanout,codemaps-extract,sync-audit-4dim}.js` (Q3).

Harness v4 manifests carry per-specialist `model`/`effort` (`.claude/commands/harness/{oss-docs,release-update}/manifest.json`), validated as REQUIRED by `internal/harness/v4manifest/validate.go:66-70` (Q5).

HARD clauses in scope (grep `\[HARD\]`):

| File:line | Clause (head) | Zone-registry entry |
|---|---|---|
| `agent-common-protocol.md:161` | "When spawning a subagent, pass the model the active profile resolves…" | none (measured: `zone-registry.md` carries no Per-Spawn entry) |
| `model-policy.md:34` | "All MoAI agents SHOULD declare `model: inherit` unless…" | none |
| `model-policy.md:245` | "New agent definitions SHOULD use `model: inherit` (default)…" | none |
| `model-policy.md:93` | "The `[1m]`-safe cost lever is the Default model… NOT per-agent pins" | none — kept, reworded |
| `model-policy.md:110` | GLM allowlist reconciliation | none — kept |
| `cache-aware-execution.md:27` (directive 10) | mid-session model/effort switch busts the cache | none — kept (main session) |

HARD inventory with an effort/model-inclusive pattern (plan-audit iter-1 D2), at `d6992e3a0`:
`grep -rnE '\[HARD\].*(effort|model)' .claude/rules` → 17 lines. In scope: agent-common-protocol.md:161,
model-policy.md:34/93/110/245, cache-aware-execution.md:27 (directive 10, cites the removed
section), dynamic-workflows.md:124 ("the script author SHALL set `effort` explicitly…"). Out of
scope (main-session or unrelated): skill-authoring.md:187 (triggers), cache-aware-execution.md:19
(@-mention), context-window-management.md:17/67/72/81 (model-specific thresholds),
cross-session-messaging.md:89, repo-local-pr-policy.md:10, session-handoff-examples.md:303/304.
`grep -cE 'dynamic-workflows|cache-aware-execution|model-policy' .claude/rules/moai/core/zone-registry.md` → 0.

Stale non-HARD sentences (D15): settings-management.md:21 ("MoAI resolves per-spawn models
through its own model pipeline"), agent-authoring.md:165 (`Agent(... model: "haiku")`),
archived-agent-rejection.md:86-91 (`Agent(general-purpose, model: …)` rows).

Prose-instruction grep for AC-AMI-007(c), RED-now at `d6992e3a0`:
`grep -rnE '(set|pass|inject)[^.]{0,40}\b(effort|model)\b[^.]{0,80}(explicit|per spawn|per-spawn|agent\(\))' .claude/rules .claude/skills internal/template/templates/.claude`
→ 6 hits in 3 file pairs: agent-common-protocol.md, settings-management.md, dynamic-workflows.md.

docs-site (re-measured, plan-audit iter-1 D9/D10): the first pattern's `[Pp]er-[Ss]pawn` token
over-matched unrelated per-spawn specialisation prose. Narrowed pattern plus the Q3 keys and the
deprecated flags:
`grep -rlE -- 'model profile|profile matrix|agent_overrides|harness_agents|agent_model_guard|[Pp]er-[Ss]pawn [Mm]odel|agent-model-audit|workflow_agents|model_routing|performance_tier|--model-policy|moai (init|update)[^|]{0,60}--profile' docs-site/content | wc -l`
→ **52** (en 14, ko 14, ja 12, zh 12). Without the Q3 keys and flags the narrowed pattern gives 40.

Docs residue outside that pattern (plan-audit iter-2 N7), added to the touch set explicitly:
`docs-site/content/{en,ja,ko,zh}/advanced/harness-v4-builder.md` (generated specialists shown with
`model:`/`effort:`, e.g. en:70), `docs-site/content/{en,ko}/cost-optimization/_index.md`,
`docs-site/content/en/multi-llm/_index.md` ("the per-agent model (and effort) assignment table"),
`docs-site/content/ko/claude-code/foundations/features-overview.md` — 8 pages. `claude-code/agentic/sub-agents.md`
hits are Claude Code feature examples and stay out of scope.

Local verify-judge effort channel (plan-audit iter-2 N3), local-only (no template mirror):
`.claude/rules/moai/workflow/verify-judge-effort-contract.md:8-9`, `.claude/workflows/sync-audit-4dim.js:201-208`
(`args.judge_effort` → `JUDGE_EFFORT` → `effort:` on four `agent()` calls), `.claude/hooks/tests/test-judge-effort-contract.sh`.
`grep -rlE 'judge_effort|JUDGE_EFFORT' .claude internal/template/templates/.claude` (worktree mirrors excluded) → exactly those
3 files; the template `sync-audit-4dim.js` carries none. `\b(effort|model)\b` cannot see `judge_effort` (`_` is a
word character), hence AC-AMI-007(d). A `\b`-free variant of the (c) regex was tried and rejected: the
shell's `grep` (ugrep) aborts it with "exceeds complexity limits".

agent-authoring residue (N8): `agent-authoring.md:219` ("Control reasoning depth with `effort` …") and
`:333` (cross-reference to dynamic-workflows § Purpose-driven model+effort selection), both trees.

Profile-setup wording (N9): `internal/cli/wizard/translations.go:444/456/468/480` and
`internal/cli/profile_setup_translations.go:165/260` label the retained `model_policy` preference
"Agent model policy — … assigning optimal models to each agent".

## §G. Always-loaded budget baseline

`go test -count=1 -run TestAlwaysLoadedTokenBudget -v ./internal/config/` at `d6992e3a0`:
`always-loaded surface = 77539 tokens (budget 77600, headroom 61, 16 entries)`. The removal
shrinks two always-loaded files (`agent-common-protocol.md`, `cache-aware-execution.md`) and the
constitution pointer line, so headroom is expected to rise; the figure is re-measured after
t1175 lands, because that card rewrites the same files.

## §H. Overlap with sibling branches

`git diff --name-only d6992e3a0...WT-rules-diet` (card t1175, HEAD `3a48485af`): 84 paths.
Intersection with this SPEC's text touch set — **21 paths**:

- agents (6): `{.claude,internal/template/templates/.claude}/agents/moai/{plan-auditor,super-advisor,sync-auditor}.md`
- rules (10): `{local,template}` × `core/{agent-common-protocol,agent-common-protocol-reference,moai-constitution}.md`, `workflow/{cache-aware-execution,cache-aware-execution-reference}.md`
- codex (3): `internal/template/templates/.codex/agents/moai/{plan-auditor,super-advisor,sync-auditor}.toml`
- docs (2): `.moai/docs/agent-lint.md`, `internal/template/templates/.moai/docs/agent-lint.md`

t1175 also edits `CLAUDE.md` (root and template) and `AGENTS.md`; this SPEC does not (0 hits).

Superseded at run entry by the mechanical gate of §I; the figures below are the plan-time snapshot.

`git diff --name-only d6992e3a0...WT-role-naming-docs` (card t1257, HEAD `ffc83b3b1`): 25 paths,
all under `.moai/reports/t1257/` and `.moai/specs/SPEC-ROLE-NAMING-DOCS-001/` — **0 overlap
today**. Its planned run surface (`.moai/reports/t1257/raw/per-file-class.tsv`, 377 paths)
intersects this SPEC's touch set in **90 paths**: 57 agent/rule/skill/codex/config paths and 33
docs-site pages. That branch is plan-only, so the overlap is re-measured at this SPEC's run
entry (REQ-AMI-002).

## §I. Committed touch set and the run-entry gate (plan-audit iter-1 D5)

The touch set is a committed file produced by a committed, read-only script:

- Generator: `.moai/reports/t1246/touch-set.sh` (greps only; frontmatter hits, codex toml,
  doctrine/docs pattern, Go/templ symbol pattern, workflow scripts, harness manifests, and an
  explicit list for files no pattern reaches).
- Output: `.moai/reports/t1246/touch-set.txt` — `sh .moai/reports/t1246/touch-set.sh > .moai/reports/t1246/touch-set.txt`
  → exit 0, **252** paths at `d6992e3a0` (revision after plan-audit iter-2; iter-1 revision had 238).
  `cut -d/ -f1-2 touch-set.txt | LC_ALL=C sort | uniq -c`: 22 `.claude/agents`, 2 `.claude/commands`,
  1 `.claude/hooks`, 12 `.claude/rules`, 8 `.claude/skills`, 5 `.claude/workflows`, 1 `.moai/docs`,
  2 `.moai/project`, 1 `CHANGELOG.md`, 60 `docs-site/content`, 24 `internal/cli`, 11 `internal/config`,
  10 `internal/harness`, 6 `internal/hook`, 5 `internal/settings`, 2 `internal/spec`,
  54 `internal/template`, 26 `internal/web`.
- Collation: the script sorts with `LC_ALL=C`; `LC_ALL=C sort -c touch-set.txt` → exit 0. The gate
  must therefore run `LC_ALL=C sort` and `LC_ALL=C comm -12`.

`<merge-base>` is defined as `git merge-base HEAD WT-role-naming-docs` → `e62c3e183` on
2026-09-26 (branch tip `024b95f77`). Gate as run today:
`git diff --name-only e62c3e183 WT-role-naming-docs | LC_ALL=C sort | LC_ALL=C comm -12 - .moai/reports/t1246/touch-set.txt | wc -l` → **0**.

t1175 predicate: `git merge-base --is-ancestor WT-rules-diet develop` → exit **1** on 2026-09-26
(not merged; the gate holds).

## §J. Update merge and main-session persistence (plan-audit iter-1 D3/D4)

Merge behaviour for keys absent from the new template:

- `internal/cli/update/backup/node_merge.go:374` — "Second pass: old-only keys (absent from the
  new template). REQ-UYP-006 retains ALL of them".
- `internal/cli/update/backup/merge.go:78` — "key only in old → absent from new template →
  retain + report (REQ-UYP-006/007)".
- Post-merge strip precedent: `stripRetiredV2DenyEntries` called at `internal/cli/update.go:388`
  and `internal/cli/update_clean_install.go:554`.

So the merge never drops the removed keys; the explicit strip step (design D14) is required for
every row of design §C.

Update ordering (plan-audit iter-2 N1), measured in `internal/cli/update.go` and
`internal/cli/update_template_sync.go`:

- `update.go:388` `stripRetiredV2DenyEntries` — comment :377-381: "after the --binary / --dry-run
  early-returns … but BEFORE the version-match short-circuit". It runs before any backup or merge,
  so it is **not** a valid host for a key strip whose backup must hold the originals.
- `update.go:493` `runTemplateSyncWithProgress` → `update_template_sync.go:773-776` returns `true`
  on a version match (no backup, no merge); `:795-799` returns the same `true` when the user cancels
  the merge prompt; otherwise `runTemplateSyncWithReporter`.
- Inside the sync: "Backup" step `BackupMoaiConfig` (:539, backup dir `.moai-backups/<timestamp>`,
  `defs.BackupsDir`), then "Restore Settings" (:610-638): `RestoreMoaiConfigRetained` (:620) returns
  the retained-key refs, `renderRetainedKeyAdvisory` (:638) prints them.
- `update.go:524-533` `if syncSkipped { … return nil }` — everything after it ("Post-sync steps")
  never runs on a version-matched update.
- Clean-install path: `update_clean_install.go:511` `RestoreMoaiConfig`, then `:554`
  `stripRetiredV2DenyEntries` — post-merge.

Hence design D14's three hosts: inside "Restore Settings" between :620 and :638 (filtering the
strip's keys out of the retained list), inside the `syncSkipped` branch for the version-match reason
only (own backup first), and the clean-install site.

Where the model policy is persisted and read:

| Surface | Writer (measured) | Reader (measured) |
|---|---|---|
| `llm.yaml performance_tier` | `init.go:984-993` via `resolveModelPolicy` (`--model-policy`, `--high`, `--medium-alias`, `--low`, `init.go:422-436`) or the init wizard `ModelPolicy`; `ApplyPerformanceTier` | web agentfm, agentlint, schemaform, `EffectiveProfile` alias — all subagent-side |
| `llm.yaml profile` | `init.go:995-1012`, `update.go:647-660`, `update_wizard.go:307-310` (`ApplyProfile`) | profile matrix resolver — subagent-side |
| system.yaml `moai.model_policy` | `update_wizard.go:332` | none: `grep -rnE 'model_policy' --include=*.go internal cmd pkg` shows no reader of this key |
| preference profile `~/.moai/claude-profiles/<name>/preferences.yaml` `model_policy` | `profile_setup.go:425` `WritePreferences`, `moai web` | `launcher.go:763` and `launch_effort_settings.go:68` via `resolveLaunchEffort` → `MapModelPolicyToEffort` — **main session** |

`TemplateContext.WithModelPolicy` (`template/context.go:260`) has no non-test caller and no
template file references `.ModelPolicy` (`grep -rln '\.ModelPolicy' internal/template/templates` → 0).
Conclusion: the init/update flags and the init/update wizard question reach only subagent-side
keys; the main-session policy lives in the preference profile and is untouched.

## §K. Rosterguard registry sites bound to removed surfaces (plan-audit iter-2 N2)

`grep -nE 'ID: +"|Path: +"' internal/harness/rosterguard/registry.go` at `d6992e3a0`, filtered to
paths this SPEC removes or rewrites: `profile-matrix-order` (:23) and `profile-matrix-group-membership`
(:35) → `internal/template/profile_matrix.go`; `config-retained-agent-names` (:48) → `internal/config/profile.go`;
`profile-matrix-test-expectations` (:91); `v4manifest-agent-tiers` (:107) and `v4manifest-tier-test` (:116);
`shipped-key-inventory` (:203) → `internal/config/testdata/shipped_key_inventory.yaml`; `template-llm-yaml`
(:210); `product-md-profile-matrix-size` (:239, CountPattern `(\d+) retained agents x 3 model tiers`)
→ `.moai/project/product.md:149`; `tech-md-profile-matrix-size` (:254) → `.moai/project/tech.md:17`;
`web-agentfm-display-rank` (:358) and `-test` (:371); `model-policy-profile-matrix-size` (:432) and
`-mirror` (:440); numeral exemptions `agentlint-section-marker` (:817) and `web-agentfm-subset-count` (:822).
Sites on files this SPEC touches but whose anchors it does not remove: `agent-authoring-catalog` (:164,
BlockStart `### Retained MoAI-custom Agents (` — agent-authoring.md:130, a different section from the
calibration matrix), `docs-truth-catalog` (:217), `manager-design-catalog-citation` (:575),
`manager-docs-then-8` / `manager-spec-then-8` (:736-737).

Failure modes if a site outlives its anchor: `check.go:77` "block anchor … not found … (the anchor moved
or the block was deleted)"; `check.go:168` CountPattern "matched %d times …, want exactly 1".
Disposition per site and milestone: design.md §F.
