---
id: SPEC-MODEL-PROFILE-MATRIX-002
title: "Agent-direct profile matrix + effort actualization"
version: "0.1.0"
status: draft
created: 2026-07-23
updated: 2026-07-23
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/template, internal/config, internal/cli, internal/web, internal/harness/v4manifest, docs-site"
lifecycle: spec-anchored
tags: "model-profile, agent-matrix, effort-injection, tokenomics, documentation"
tier: L
era: V3R6
related_specs: [SPEC-MODEL-PROFILE-MATRIX-001, SPEC-MODEL-TIER-PLANTYPE-001, SPEC-AGENT-ARCH-V2-001, SPEC-WEBCONF-SIMPLIFY-001]
---

# SPEC-MODEL-PROFILE-MATRIX-002 — Agent-direct profile matrix + effort actualization

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-23 | manager-spec | Initial draft. Supersedes the 6-group abstraction of SPEC-MODEL-PROFILE-MATRIX-001 with a direct 33-cell profile→agent matrix; revives frontmatter `effort:` rewrite in a narrowed form; retires the 36-cell Tier×Phase axis and the `llm.yaml profiles:` mirror; corrects SPEC-001 DECISION-001's effort-injection claim. |

---

## §A Context

### §A.1 Predecessor and supersession scope

`SPEC-MODEL-PROFILE-MATRIX-001` (completed 2026-07-20) introduced a 3-column profile axis (`max` / `medium` / `low`) resolved through **six agent groups** (`spec_auditors`, `develop`, `advisor`, `design_harness_e2e`, `docs`, `git`). This SPEC supersedes that group abstraction. It does **not** supersede SPEC-001 wholesale: the profile axis itself, the `llm.profile` config key, the `EffectiveProfile()` alias chain, the resolver precedence order, the `inherit` sentinel for unmapped agents, and the No-Haiku policy all survive unchanged.

Two independent forces drive the change.

### §A.2 Force 1 — the leaderboard inverted the naming semantics

The DeepSWE leaderboard v1.1 (113 tasks) reports per-model, per-effort rows. User-supplied readings for the three models of interest:

| Model [effort] | Per-task cost | Pass@1 | Output tokens | Agent steps |
|---|---|---|---|---|
| Fable 5 [low] | $3.76 | 60% | 25k | 38 |
| Opus 4.8 [high] | $4.28 | 52% | 50k | 73 |
| GLM 5.2 [max] | $3.92 | 42% *or* 45% (conflicting readings) | 78k | 129 |

**These four-metric readings are UNVERIFIED user-supplied inputs.** They are recorded here as the SPEC's design input, not as observed fact. A blocking precondition (§B.0 / milestone S0) confirms them against the live source before any documentation number is committed. See `research.md` §A for the partial probe result and why it does not discharge S0.

If the readings hold, Fable 5 at `low` effort leads on **all three cost axes** (cheapest per task, fewest tokens, fewest steps) while scoring highest. Combined with Fable 5's inclusion in the subscription plan, the consequence is a semantic inversion: the `max` / `medium` / `low` profile names imply an ordering "higher = stronger and costlier", and that ordering is now **false** — the Max profile is simultaneously cheaper and higher-scoring than the Medium profile. The axis is a **subscription-tier access axis**, not a performance-grade axis.

This SPEC does not rename the profiles (renaming would break `llm.profile` values, the CLI flag, the wizard, and every doc surface). It requires the inversion to be **disclosed in documentation** rather than silently shipped.

### §A.3 Force 2 — the group abstraction broke

The user-assigned per-agent cells (§A.4) split two of the six existing groups:

- `spec_auditors` splits: under `medium`, `manager-spec` takes `opus/high` while `plan-auditor` and `sync-auditor` take `opus/xhigh`.
- `design_harness_e2e` splits three ways: under `max`, `manager-design` takes `fable/low`, `builder-harness` takes `opus/high`, `e2e-tester` takes `fable/low`; under `low`, all three take `sonnet/medium` but diverge from the other columns' shape.

Only two same-cell pairs survive across all three columns (`plan-auditor` ≡ `sync-auditor`; `manager-docs` ≡ `Explore`). A grouping layer that partitions 11 agents into 6 groups, of which 4 are singletons and 2 no longer hold, is dead weight: it adds an indirection hop, a second literal (`agentGroupMembership`), and a `group` display column that no longer carries information.

### §A.4 The 33-cell matrix (authoritative — settled design input, MUST NOT be re-derived)

| agent | max | medium | low |
|---|---|---|---|
| manager-spec | fable / low | opus / high | opus / low |
| plan-auditor | fable / low | opus / xhigh | opus / low |
| sync-auditor | fable / low | opus / xhigh | opus / low |
| manager-develop | opus / xhigh | opus / high | sonnet / medium |
| super-advisor | fable / medium | fable / low | opus / medium |
| manager-design | fable / low | opus / high | sonnet / medium |
| builder-harness | opus / high | opus / high | sonnet / medium |
| e2e-tester | fable / low | opus / high | sonnet / medium |
| manager-docs | sonnet / medium | sonnet / medium | sonnet / medium |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / medium | sonnet / medium | sonnet / medium |
| *(unmapped user agents)* | inherit | inherit | inherit |

Matrix properties (asserted as acceptance criteria in §B.1):

- 11 mapped agents × 3 profiles = **33 cells**.
- Zero `haiku` anywhere.
- Models ⊆ `{fable, opus, sonnet}`.
- Efforts ⊆ `{low, medium, high, xhigh}` (no `max` effort cell).
- `inherit` does **not** appear inside the matrix — it survives only as the unmapped-agent fallback.
- Three agents are **profile-invariant**: `manager-docs`, `manager-git`, `Explore`.
- The Max column's Opus assignments (`manager-develop` opus/xhigh, `builder-harness` opus/high) are a **deliberate quality-first choice for high-failure-cost work**, NOT a benchmark optimum. This rationale is a documentation requirement (§B.6).

### §A.5 `Explore` changes the resolver contract

Today `Explore` is unmapped: `ResolveAgentModelEffort` returns `{inherit, ""}, hasGroup=false`, and `TestResolveAgentModelEffort_Inherit` (`internal/template/profile_matrix_test.go:144`) pins `Explore` and `some-user-agent` to that same behavior.

Under the new matrix `Explore` is an **explicit** mapping (`sonnet/medium` in all three columns) while unmapped user agents still resolve to `inherit`. The test must be amended to split those two cases.

`Explore` is an Anthropic built-in with **no agent file on disk**. Per-spawn `model` injection is therefore the only applicable channel for it; its `medium` effort is documented intent only and is never written to a file.

### §A.6 Effort injection is path-dependent (SPEC-001 DECISION-001 correction)

`SPEC-MODEL-PROFILE-MATRIX-001` DECISION-001 states that effort cannot be injected per-spawn and characterizes the Workflow channel as "prompt-level". **That is inaccurate.** Verified from the live tool schemas:

| Channel | Orchestration mode | `model` | `effort` | Consequence |
|---|---|---|---|---|
| `Agent` tool | Mode 5 (sub-agent delegation, the standard path) | runtime arg, `enum: sonnet\|opus\|haiku\|fable` | **no parameter exists** | frontmatter `effort:` is the effective effort → rewrite required |
| `Workflow` tool `agent()` | Mode 6 (dynamic-workflow) | `opts.model` | `opts.effort ∈ {low, medium, high, xhigh, max}` | structured parameter, injectable directly |

The Workflow form is `agent(prompt, {agentType: 'manager-develop', model: 'opus', effort: 'xhigh'})` — a **structured option**, not prompt-level steering. Corroborated by `.claude/rules/moai/workflow/dynamic-workflows.md:91`, which validates workflow `agent()` effort against that closed set.

The matrix effort is therefore consumed through **two channels**, and requirements must cover both: (a) frontmatter rewrite for the Agent-tool path, (b) direct `opts.effort` injection for the Workflow path, which needs a matrix-lookup route reachable from workflow scripts.

### §A.7 Frontmatter rewrite revival — why it is safe this time

SPEC-001 REQ-MPM-024 retired `ApplyTierProfile` because it rewrote each shipped agent's `model:` **and** `effort:`, re-introducing a concrete-model pin (the `[1m]` hazard) and producing large diffs across both the deployed and template trees. The revival is deliberately narrower; four mitigations (§B.3) bound it:

1. `effort:` line only — `model:` never written by this path.
2. Deployed tree only — the template tree is immutable and keeps a fixed Medium-profile baseline.
3. Reapply **after** template deploy on `moai update`, so a profile survives an update.
4. Agents carrying an `llm.agent_overrides[<agent>].effort` are excluded, matching resolver precedence.

The precedent for (1) and (2) already exists in-tree: `internal/settings/agentfm/agentfm.go:89` `Patch(path, model, effort string, deleteEffort bool)` is a frontmatter-only, live-files-only editor with no template dual-write.

---

## §B Requirements (GEARS)

### §B.0 S0 — leaderboard verification (blocking precondition)

- **REQ-MPM2-001** — **While** the v1.1 leaderboard readings in §A.2 remain unconfirmed against the live source, the implementation shall not commit any benchmark figure to a documentation surface.
- **REQ-MPM2-002** — The verification record shall resolve the GLM 5.2 Pass@1 conflict (42% vs 45%) to a single canonical value, stating the effort level the row belongs to and the reported error bar.
- **REQ-MPM2-003** — The verification record shall pin, per model, all four metrics (per-task cost, Pass@1, output tokens, agent steps) **together with the effort level of the row they were read from**, and this record shall be the single source for every number written in milestone M6.
- **REQ-MPM2-004** — **When** a confirmed metric differs from the §A.2 user-supplied reading, the implementation shall use the confirmed value and record the delta in `progress.md` rather than silently substituting it.

### §B.1 M1 — 33-cell matrix redesign

- **REQ-MPM2-010** — The profile matrix shall map `profile → agent → {model, effort}` directly, containing exactly the 33 cells transcribed in §A.4.
- **REQ-MPM2-011** — The group constants `GroupSpecAuditors`, `GroupDevelop`, `GroupAdvisor`, `GroupDesignHarnessE2E`, `GroupDocs`, `GroupGit` shall not exist after M1.
- **REQ-MPM2-012** — The `agentGroupMembership` agent→group table shall not exist after M1.
- **REQ-MPM2-013** — The `AgentGroup(agent)` accessor and every consumer of its return value shall be removed or retargeted to the direct per-agent lookup.
- **REQ-MPM2-014** — The matrix shall carry an explicit `Explore` row resolving to `sonnet / medium` in all three profile columns.
- **REQ-MPM2-015** — **When** the resolver is queried for an agent absent from the matrix, it shall return the `inherit` sentinel with a false membership flag, so the caller skips model injection.
- **REQ-MPM2-016** — The resolver shall preserve the existing precedence order: `llm.agent_overrides[agent]` → active-profile matrix cell → unknown-profile fallback to the `medium` column → `inherit` for unmapped agents.
- **REQ-MPM2-017** — The matrix shall contain zero cells whose model is `haiku`.
- **REQ-MPM2-018** — Every matrix cell's model shall be a member of `{fable, opus, sonnet}`.
- **REQ-MPM2-019** — Every matrix cell's effort shall be a member of `{low, medium, high, xhigh}`.
- **REQ-MPM2-020** — The literal `inherit` shall not appear as a model value inside the matrix itself.
- **REQ-MPM2-021** — The agents `manager-docs`, `manager-git`, and `Explore` shall resolve to an identical `{model, effort}` pair across all three profile columns.
- **REQ-MPM2-022** — **When** the matrix is amended, the existing resolver tests pinning the retired group vocabulary shall be amended in the same change so that no test references a removed group constant.
- **REQ-MPM2-023** — The `Explore`-and-unmapped-agent inherit test shall be split so that `Explore` asserts the explicit `sonnet / medium` mapping while an arbitrary user-agent name asserts the `inherit` fallback.

### §B.2 M2 — retire the 36-cell axis and the config mirror

- **REQ-MPM2-030** — The `model_routing_profiles` block shall be removed from both `workflow.yaml` copies (project tree and template tree).
- **REQ-MPM2-031** — The `RouteModelFor` accessor, its `ModelRoutingProfiles` type, and its config-load validators shall be removed.
- **REQ-MPM2-032** — **Where** removal of `RouteModelFor` would orphan its tests, those tests shall be deleted rather than retargeted, since the axis has no production consumer.
- **REQ-MPM2-033** — The `profiles:` block shall be removed from both `llm.yaml` copies, leaving the Go constant as the sole matrix SSOT.
- **REQ-MPM2-034** — After M2 the matrix literal shall exist in exactly two places: the Go constant and its fidelity test.
- **REQ-MPM2-035** — **When** `moai update` runs against a project whose `llm.yaml` still carries a non-empty `profiles:` block, the updater shall detect it and shall not discard the user's customization silently.
- **REQ-MPM2-036** — **When** a detected `profiles:` customization is representable as per-agent overrides, the updater shall migrate it into `llm.agent_overrides`; **when** it is not representable, the updater shall emit a warning naming the affected profile column and cells.
- **REQ-MPM2-037** — The config loader shall not fail on a legacy `llm.yaml` that still carries `profiles:` or a legacy `workflow.yaml` that still carries `model_routing_profiles`; the unknown block shall be tolerated as inert.

### §B.3 M3 — effort actualization (both channels)

- **REQ-MPM2-040** — The system shall provide an agent-effort application function that writes each deployed MoAI agent's frontmatter `effort:` value from the active profile's matrix cell.
- **REQ-MPM2-041** — The effort application function shall rewrite the `effort:` line only and shall not write the `model:` line.
- **REQ-MPM2-042** — The effort application function shall not modify any file under `internal/template/templates/.claude/agents/`.
- **REQ-MPM2-043** — The template tree's agent frontmatter shall carry the Medium-profile effort values as a fixed baseline independent of any project's active profile.
- **REQ-MPM2-044** — **When** `moai update` deploys templates, the effort application shall run after deployment completes, so a user's non-default profile survives the update.
- **REQ-MPM2-045** — **Where** an agent has an `llm.agent_overrides[<agent>].effort` entry, the effort application function shall exclude that agent from rewrite.
- **REQ-MPM2-046** — The effort application shall be invoked at the four existing profile-application seams: `moai init`, the `moai update` profile flag path, the `moai update` wizard path, and the web-console profile save path.
- **REQ-MPM2-047** — **When** the effort application encounters an agent name with no corresponding file on disk, it shall skip that agent without error.
- **REQ-MPM2-048** — The effort application shall not modify agent body bytes; only the frontmatter `effort:` value shall change.
- **REQ-MPM2-049** — The system shall expose a machine-readable route by which a dynamic-workflow script can obtain an agent's resolved `{model, effort}` under the active profile, for injection as the `Workflow` tool's `opts.model` / `opts.effort`.
- **REQ-MPM2-050** — The documentation shall state that effort reaches the agent through two distinct channels — frontmatter for the `Agent` tool path (which has no `effort` parameter) and `opts.effort` for the `Workflow` tool path — and shall record that SPEC-MODEL-PROFILE-MATRIX-001 DECISION-001's "effort cannot be injected per-spawn / Workflow is prompt-level" wording is superseded.
- **REQ-MPM2-051** — The documentation shall state that `Explore` has no agent file, so its matrix effort is documented intent consumed only through the Workflow path.

### §B.4 M4 — init wizard question

- **REQ-MPM2-060** — The `moai init` model-policy question shall present subscription-tier framing rather than model-class framing.
- **REQ-MPM2-061** — The question's option text shall not reference `haiku`.
- **REQ-MPM2-062** — **Where** the wizard's option values feed the profile normalizer, the resulting persisted `llm.profile` shall be one of `max`, `medium`, `low`.
- **REQ-MPM2-063** — The question title, description, and every option label and description shall have `ko`, `ja`, and `zh` translations.
- **REQ-MPM2-064** — **When** the wizard renders under a non-English conversation language, the model-policy question shall not fall back to English text.
- **REQ-MPM2-065** — The option descriptions shall state the subscription tier each profile targets and shall not assert a performance ordering contradicted by §A.2.

### §B.5 M5 — web console cleanup

- **REQ-MPM2-070** — The web console agent-frontmatter model selector shall not offer `haiku`.
- **REQ-MPM2-071** — The v4manifest lightblue-tier suggestion shall be `sonnet / low`.
- **REQ-MPM2-072** — The `agentfm.tier.desc` i18n string shall be re-worded in all four locales to describe the effort-reapplication behavior that M3 restores, rather than deleted.
- **REQ-MPM2-073** — The orphaned `mp.*` i18n key family shall be removed from all four locale files.
- **REQ-MPM2-074** — **When** a locale file is edited in M5, all four locale files shall be edited in the same change so no key exists in a subset of locales.

### §B.6 M6 — 4-locale documentation

- **REQ-MPM2-080** — The README benchmark table shall be replaced with S0-confirmed v1.1 figures in all four locale files.
- **REQ-MPM2-081** — The `advanced/profile-matrix.md` page shall present the 33-cell per-agent matrix in all four locales, and its agent-group table shall be removed.
- **REQ-MPM2-082** — The `advanced/no-haiku-3tier.md` benchmark table shall be replaced with S0-confirmed v1.1 figures in all four locales.
- **REQ-MPM2-083** — The `multi-llm/model-policy.md` page shall no longer assert the retired "every worker agent is pinned to Sonnet 5" policy, in all four locales.
- **REQ-MPM2-084** — The `multi-llm/model-policy.md` Chinese copy shall not contain a Haiku column or any per-agent haiku assignment.
- **REQ-MPM2-085** — The four locale copies of `multi-llm/model-policy.md` shall share one table shape, resolving the Japanese copy's divergence.
- **REQ-MPM2-086** — The stale "all workers Sonnet fixed" tier table in `.claude/rules/moai/development/model-policy.md` shall be reconciled with the modern resolver description in the same file.
- **REQ-MPM2-087** — **When** `.claude/rules/moai/development/model-policy.md` is edited, its byte-identical template twin shall be edited in the same commit.
- **REQ-MPM2-088** — The model enum in `agent-authoring.md` and in `dynamic-workflows.md` shall include `fable`.
- **REQ-MPM2-089** — Every documentation reference to the retired `--model-policy` flag name shall be updated to the current `--profile` flag name.
- **REQ-MPM2-090** — The documentation shall state explicitly that the `max` / `medium` / `low` names denote **subscription-tier access**, not performance grade, and that the Max profile is both cheaper and higher-scoring than the Medium profile under the S0-confirmed data.
- **REQ-MPM2-091** — The documentation shall state that the Max profile's Opus assignments are a deliberate quality-first choice for high-failure-cost work and not a benchmark optimum.
- **REQ-MPM2-092** — **When** any documentation surface in this milestone is edited, all four locale copies of that surface shall be edited in the same change.

### §B.7 M7 — guard realignment and verification

- **REQ-MPM2-100** — The haiku-residual lint rule's surface list shall include the web-console agent-frontmatter model option set.
- **REQ-MPM2-101** — The haiku-residual lint rule's surface list shall include the v4manifest tier-suggestion table.
- **REQ-MPM2-102** — **When** the haiku-residual rule's surfaces change, its retired-surface entries (`model_routing_profiles`, `validRoutingModels`) shall be removed in the same change so the rule does not scan for a block that no longer exists.
- **REQ-MPM2-103** — A property test shall assert the matrix invariants of §B.1 (33 cells, closed model set, closed effort set, no `haiku`, no in-matrix `inherit`, profile-invariant trio).
- **REQ-MPM2-104** — The full Go test suite shall pass.
- **REQ-MPM2-105** — The build shall succeed for the project's release target platforms.

### §B.8 Cross-cutting risk requirements

- **REQ-MPM2-110** — **Where** the `fable` model is unavailable in the runtime environment, the system shall document the fallback the user is expected to take, since the Max profile depends on Fable for six of its eleven cells.
- **REQ-MPM2-111** — The documentation shall record that under a GLM backend the increased Fable usage collapses to `glm-5.2`, whose observed step count is the least step-efficient of the three models, and that CG-mode profile pairings warrant review on that basis.
- **REQ-MPM2-112** — **When** the effort application runs inside this repository itself, it shall be a no-op at the default `medium` profile, because the deployed agent frontmatter values already equal the template baseline.
- **REQ-MPM2-113** — The system shall not silently drop a user's existing `llm.yaml profiles:` customization; removal without either migration or a warning is prohibited.

---

## §C Exclusions

This SPEC deliberately excludes the following. Each item is out of scope for this SPEC and is either owned elsewhere or deferred.

### Out of Scope — profile renaming

- Renaming the `max` / `medium` / `low` profile vocabulary to names that match the inverted cost/quality ordering. §A.2 requires disclosure, not renaming; a rename would break the `llm.profile` value set, the CLI flag, the wizard values, and every documentation surface at once.
- Adding a fourth profile column.
- Reintroducing the retired `plan_type` axis in any form.

### Out of Scope — model behavior and routing beyond the matrix

- Changing the GLM effort-collapse table (`low`→thinking-off, `medium`/`high`→reasoning-high, `xhigh`/`max`→reasoning-max) or the `manager-develop` coding-max singleton override.
- Changing the GLM model-alias map (`fable`→`glm-5.2`).
- Live z.ai wire-effectiveness validation of the GLM effort overlay — inherited as pending from SPEC-MODEL-PROFILE-MATRIX-001 and unchanged here.
- Orchestrator spawn-time routing policy (which agent handles which task). This SPEC changes what `{model, effort}` an agent resolves to, not who is spawned.

### Out of Scope — agent catalog and tiering

- Adding, removing, or renaming any of the 11 retained agents.
- Changing the v4manifest name→tier badge assignment table. Only the lightblue tier's suggested `{model, effort}` pair changes (REQ-MPM2-071).
- Changing agent `tools:`, `skills:`, or body content.

### Out of Scope — infographic regeneration

- Regenerating `assets/images/readme/tokenomics-harness-{en,ko,ja,zh}.png`. The images may embed superseded benchmark figures; assessing and regenerating them is deferred to a follow-up, and M6 records the assessment result rather than performing the regeneration.

### Out of Scope — separate quality surfaces

- The `moai web` console's non-agentfm panels.
- Wizard questions other than `model_policy`.
- The `haiku` alias in the GLM model map (`glm.models.haiku`), which is an exempt surface of the haiku-residual rule by design.

---

## §D Constraints

| # | Constraint | Source |
|---|---|---|
| C-1 | Template tree content must remain internally neutral — no SPEC IDs, REQ tokens, internal dates, or commit SHAs in `internal/template/templates/**`. | CLAUDE.local.md §25 |
| C-2 | Template-First: every template change is made in `internal/template/templates/` first, then `make build`. | CLAUDE.local.md §2 |
| C-3 | `.claude/rules/.../model-policy.md` and its template twin are byte-parity-checked; both sides are edited in one commit. | `internal/template/rule_template_mirror_test.go` |
| C-4 | 4-locale parity is mandatory for every docs-site page and for the README set. | CLAUDE.local.md §17 |
| C-5 | The No-Haiku policy is a HARD gate not skippable via SPEC `lint.skip`. | `internal/spec/lint_haiku_residual.go` |
| C-6 | `settings.local.json` and machine-specific values are never written by this work. | CLAUDE.local.md §2 |
| C-7 | The `Agent` tool model enum is `sonnet\|opus\|haiku\|fable`; `inherit` is not an injectable runtime value, so the unmapped-agent path must skip injection rather than pass `inherit`. | live tool schema |
| C-8 | A parallel session may be active on this shared checkout; changes are committed with explicit pathspecs. | project operating practice |

---

## §E Decisions

| # | Decision | Rationale |
|---|---|---|
| D-1 | Group system abolished; direct `profile → agent → {model, effort}` mapping. | User-confirmed. 4 of 6 groups were already singletons; the remaining 2 are split by the new cells. |
| D-2 | Frontmatter `effort:` rewrite revived, narrowed to one line, deployed tree only, post-deploy, override-excluded. | User-confirmed. Without it the matrix effort is inert on the `Agent` tool path, which has no `effort` parameter. |
| D-3 | 36-cell Tier×Phase axis retired entirely rather than migrated. | User-confirmed. Zero non-test production call sites, so runtime impact is nil. |
| D-4 | `llm.yaml profiles:` removed; the Go constant is the single SSOT. | User-confirmed. Drops matrix literal duplication from 4 copies to 2. |
| D-5 | Profile names retained; the inverted ordering is disclosed in documentation. | User-confirmed. Renaming has a far larger blast radius than the problem it solves. |
| D-6 | `Explore` becomes an explicit matrix row while unmapped user agents keep `inherit`. | The two cases were conflated by a single test; they are semantically distinct. |
| D-7 | S0 leaderboard confirmation is a blocking precondition, not a parallel task. | Every M6 number depends on it; committing unverified figures would violate the no-unobserved-claim invariant. |
| D-8 | Legacy `profiles:` / `model_routing_profiles` blocks are tolerated as inert rather than rejected on load. | A hard load failure would break existing projects on upgrade. |

---

## §F Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | Removing `llm.yaml profiles:` silently drops an existing user override. | High | REQ-MPM2-035/036/113 — detect, migrate to `agent_overrides`, or warn. |
| R-2 | The Max profile leans on Fable for 6 of 11 cells; environments without Fable access break. | High | REQ-MPM2-110 — documented fallback path. |
| R-3 | Under GLM, heavier Fable usage collapses to `glm-5.2`, the least step-efficient of the three models. | Medium | REQ-MPM2-111 — record and flag CG-mode pairings for review. |
| R-4 | Documentation scale: 4 locales × (README + 3 doc page families + 2 rules files). One missed locale breaks parity. | Medium | REQ-MPM2-092 + routing M6 through the `oss-docs` harness. |
| R-5 | This repository's own `.claude/agents/moai/` becomes a rewrite target. | Low | REQ-MPM2-112 — no-op at the default `medium` profile; divergence only under a non-default profile, which is a maintainer choice. |
| R-6 | S0 confirmation contradicts the §A.2 readings, invalidating the framing in §A.2. | Medium | REQ-MPM2-004 — use the confirmed value, record the delta; the structural work (M1-M5, M7) does not depend on the numbers. |
| R-7 | Frontmatter rewrite reintroduces large diffs, the failure mode that retired the predecessor. | Medium | REQ-MPM2-041/042/048 — one line per file, ≤10 files, deployed tree only. |
| R-8 | Parallel session race on the shared checkout during a broad multi-file change. | Medium | C-8 — pathspec commits, pre-spawn fetch. |

---

## §G Traceability

| Milestone | REQ range | AC range |
|---|---|---|
| S0 | REQ-MPM2-001 … 004 | AC-MPM2-001 … 003 |
| M1 | REQ-MPM2-010 … 023 | AC-MPM2-010 … 019 |
| M2 | REQ-MPM2-030 … 037 | AC-MPM2-020 … 026 |
| M3 | REQ-MPM2-040 … 051 | AC-MPM2-030 … 040 |
| M4 | REQ-MPM2-060 … 065 | AC-MPM2-050 … 054 |
| M5 | REQ-MPM2-070 … 074 | AC-MPM2-060 … 064 |
| M6 | REQ-MPM2-080 … 092 | AC-MPM2-070 … 082 |
| M7 | REQ-MPM2-100 … 105 | AC-MPM2-090 … 095 |
| Cross-cutting | REQ-MPM2-110 … 113 | AC-MPM2-100 … 103 |

Full AC↔REQ mapping: `acceptance.md` §D.

---

## §H Cross-References

- `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-001/` — predecessor; group abstraction and DECISION-001 superseded here.
- `.moai/specs/SPEC-AGENT-ARCH-V2-001/` — origin of the No-Haiku 3-tier policy and the 36-cell axis retired here.
- `.moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/` — origin of the GLM effort overlay retained unchanged.
- `.moai/specs/SPEC-WEBCONF-SIMPLIFY-001/` — origin of the v4manifest tier badge table whose lightblue suggestion changes here.
- `.claude/rules/moai/development/model-policy.md` — self-contradictory rule file reconciled in M6.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — Workflow `agent()` effort closed set corroborating §A.6.
