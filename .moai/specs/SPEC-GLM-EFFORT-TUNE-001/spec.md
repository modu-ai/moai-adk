---
id: SPEC-GLM-EFFORT-TUNE-001
title: "GLM effort overlay configuration tune-up (P1/P2/P4)"
version: "0.1.0"
status: in-progress
created: 2026-07-14
updated: 2026-07-14
author: manager-spec
priority: P2
phase: "v3.x config-tune"
module: "internal/template/glm_effort_overlay.go + .moai/config/sections/llm.yaml"
lifecycle: spec-anchored
tags: "glm, effort, overlay, config, template-mirror, reasoning-effort"
related_specs: [SPEC-MODEL-TIER-PLANTYPE-001]
---

# SPEC-GLM-EFFORT-TUNE-001 — GLM effort overlay configuration tune-up

## HISTORY

| Version | Date | Author | Notes |
|---------|------|--------|-------|
| 0.1.0 | 2026-07-14 | manager-spec | Initial plan-phase draft. Carries forward 3 of 4 findings (P1/P2/P4) from the 2026-07-14 GLM effort overlay configuration review. P3 (GLM wire-effectiveness live validation) is explicitly Out-of-Scope — it is owned by the SPEC-MODEL-TIER-PLANTYPE-001 follow-up backlog. |

## §A. Overview

This SPEC tunes the GLM effort overlay subsystem that was introduced by **SPEC-MODEL-TIER-PLANTYPE-001** (the plan_type × tier × agent {model, effort} redesign). The user proposal that triggered this work — Opus 4.8 [1m] default effort high; GLM-5.2 `reasoning_effort=high` for standard-work agents; `reasoning_effort=max` for advanced reasoning/audit agents — was verified by the configuration review to be **already largely aligned** with the overlay implemented in `internal/template/glm_effort_overlay.go`. The review surfaced four findings (P1–P4); this SPEC implements three (P1, P2, P4) and references the fourth (P3) as a separate backlog concern owned by SPEC-MODEL-TIER-PLANTYPE-001's follow-up list.

The three in-scope findings are small, surgical configuration-hygiene changes:

- **P1** — `builder-harness` is removed from the `glmCodingMaxOverrideAgents` override set; `manager-develop` stays. The harness agent uses `CollapseClaudeEffortToGLM(high) = reasoning-high`.
- **P2** — the GLM reasoning-control mapping (3 reachable states + collapse table) is exposed as a documentation block inside `llm.yaml` AND its template mirror, without changing runtime behavior (the Go overlay remains the SSOT).
- **P4** — the framing of GLM as a "2-tier high/max" system is corrected to the actual 3-state reality (`thinking-off` / `reasoning-high` / `reasoning-max`) in the overlay file's comments, the `llm.yaml` exposure block, and any docs-site references that framing it as 2-tier.

The truth these findings rest on was read directly from `internal/template/glm_effort_overlay.go` (lines 26–33, 88–100, 102–110), `internal/template/model_policy.go` (lines 140–154, 294–323), `.claude/rules/moai/development/agent-authoring.md` (lines 334–374), and `internal/template/glm_effort_overlay_test.go` (lines 114–135). It is NOT based on memory-summary prose (per `feedback_defect_claim_verification`).

## §B. Background — the GLM effort overlay (implementation truth)

The overlay exists because z.ai does NOT implement Claude's 5-level effort vocabulary (`low/medium/high/xhigh/max`). z.ai's `reasoning_effort` request field accepts only `{high, max}`, the `thinking` field is a binary enabled/disabled toggle, and `reasoning_effort` is moot when `thinking` is disabled. The overlay therefore **collapses** the 5 Claude effort levels onto z.ai's 3 reachable reasoning states, with a coding-max override for code-producing agents.

### §B.1 The 3 reachable GLM reasoning states (named constants in code)

From `glm_effort_overlay.go:26-33`:

| Constant | Name | ThinkingEnabled | reasoning_effort | Reached by |
|----------|------|-----------------|------------------|------------|
| `GLMStateThinkingOff` | `thinking-off` | false | (moot) | Claude `low` effort |
| `GLMStateReasoningHigh` | `reasoning-high` | true | `high` | Claude `medium` or `high` effort |
| `GLMStateReasoningMax` | `reasoning-max` | true | `max` | Claude `xhigh` or `max` effort, OR the coding-max override |

This is a **3-state system**, not a 2-tier system. The user proposal's "high/max" axis describes the two REASONING-ON tiers; the third reachable state (`thinking-off`) is the mechanical tier below them (occupied by `manager-git` whose Claude effort is `low` across all tiers — `model_policy.go:306,320`). P4 corrects the framing wherever it has drifted to "2-tier".

### §B.2 The collapse function (5 → 3, total)

From `glm_effort_overlay.go:88-100` (`CollapseClaudeEffortToGLM`):

| Claude effort | → GLM state |
|---------------|-------------|
| `low` | `thinking-off` |
| `medium`, `high` | `reasoning-high` |
| `xhigh`, `max` | `reasoning-max` |
| (unrecognized) | `reasoning-max` (totality clause — never under-reason) |

### §B.3 The coding-max override set (the P1 target)

From `glm_effort_overlay.go:102-110`:

```go
var glmCodingMaxOverrideAgents = map[string]bool{
    "manager-develop": true,
    "builder-harness": true,
}
```

Agents in this set force `reasoning-max` REGARDLESS of their collapse result (z.ai's "reasoning_effort: max for coding tasks" recommendation). **P1 removes `builder-harness` from this set.**

### §B.4 Why P1 is defensible — the builder-harness role profile

`builder-harness` effort is `high` across all three tiers in BOTH plan_type profiles (`model_policy.go:303,317`):

```go
"builder-harness": {{"opus", "high"}, {"opus", "medium"}, {"opus", "medium"}},   // api
"builder-harness": {{"sonnet", "high"}, {"sonnet", "medium"}, {"sonnet", "medium"}}, // subscription
```

The agent-authoring Effort-Level Calibration Matrix (`agent-authoring.md:350`) classifies builder-harness default effort as `high` with rationale **"artifact scaffolding (agents/skills/plugins/hooks)"** — generation volume + metadata accuracy, NOT deep coding reasoning. By contrast, `manager-develop` is `xhigh` with rationale **"run-phase implementation (coding/agentic)"** (`agent-authoring.md:343`). z.ai's "max for coding" recommendation's PRIMARY beneficiary is manager-develop.

Forcing builder-harness to `reasoning-max` also contradicts the stated CG-mode cost-reduction goal (CLAUDE.md §15 CG Mode — "60-70% cost reduction on implementation-heavy tasks"). builder-harness DOES generate code (dynamic specialist agents), so a "coding" classification is defensible — design.md §A records BOTH options and the tradeoff, and recommends EXCLUDE with this cost/role rationale.

### §B.5 Why P2 is needed — config transparency

`.moai/config/sections/llm.yaml` (local) and `internal/template/templates/.moai/config/sections/llm.yaml` (template) both expose `glm.base_url` and `glm.models` (tier→GLM model map), but neither carries ANY surface describing the reasoning-effort mapping. A user reading `llm.yaml` cannot see the overlay; P2 adds a documentation block that describes the 3 states and the collapse table without changing runtime behavior (the Go overlay remains the SSOT). Per CLAUDE.local.md §2 [HARD] Template-First, the template source is edited first; `make build` recompiles. Mirror parity (rule_template_mirror_test.go pattern) must hold.

### §B.6 Why P4 is needed — framing drift

Some comments and docs describe GLM as a "2-tier high/max" system. The code constants (`glm_effort_overlay.go:26-33`) already encode 3 reachable states correctly — P4 is a *framing correction* in prose (comments + the new llm.yaml exposure block + any docs-site references), not a code-constant change. The user's high/max axis is preserved as the two REASONING-ON tiers; `thinking-off` is documented as the mechanical tier below them.

## §C. Requirements (GEARS notation)

### P1 — builder-harness removed from the coding-max override set

**REQ-GET-001** (Ubiquitous): The `glmCodingMaxOverrideAgents` set in `internal/template/glm_effort_overlay.go` shall contain exactly one entry, `manager-develop`, after this SPEC is implemented.

**REQ-GET-002** (Ubiquitous): The `builder-harness` agent shall NOT be a member of the `glmCodingMaxOverrideAgents` set.

**REQ-GET-003** (Event-driven): **When** `ResolveGLMReasoning("builder-harness", "high")` is invoked under a GLM backend, the overlay shall return the GLM canonical state `reasoning-high` (the collapse of Claude `high`), NOT `reasoning-max`.

**REQ-GET-004** (Capability gate): **Where** an agent is `manager-develop`, the overlay shall continue to force the GLM canonical state `reasoning-max` regardless of the collapse result (z.ai coding-task recommendation preserved).

**REQ-GET-005** (Ubiquitous): The test `TestGLMCodingMaxOverrideAgents_ExactlyTwo` in `internal/template/glm_effort_overlay_test.go` shall be updated — renamed and reworded — to assert the override set is exactly `{manager-develop}` (one member), and the doc comments on `glmCodingMaxOverrideAgents`, `IsGLMCodingMaxOverrideAgent`, and `GLMCodingMaxOverrideAgents` shall be updated to remove the "two code-producing retained agents" framing.

### P2 — GLM reasoning-effort mapping exposed in llm.yaml

**REQ-GET-006** (State-driven): **While** a project ships the GLM backend, the file `.moai/config/sections/llm.yaml` AND its template mirror `internal/template/templates/.moai/config/sections/llm.yaml` shall carry a documentation block under the existing `glm:` key describing the 3 reachable reasoning states (`thinking-off` / `reasoning-high` / `reasoning-max`) and the 5→3 collapse table, so that the GLM effort mapping is visible to a user reading the config file.

**REQ-GET-007** (Ubiquitous): The `llm.yaml` exposure block shall state explicitly that the Go overlay (`internal/template/glm_effort_overlay.go`) is the runtime SSOT and that the YAML block is documentation-only — the block shall NOT introduce a parallel runtime path.

**REQ-GET-008** (Ubiquitous): The `llm.yaml` exposure block shall describe the overlay's wire state as **"implemented and wired; live validation of z.ai wire-effectiveness pending"** — it shall NOT claim the overlay "works", "is guaranteed", or "is validated" (per `verification-claim-integrity.md` §1.1 surface 3 + the MODEL-TIER-PLANTYPE-001 honesty caveat carried forward by this SPEC).

**REQ-GET-009** (State-driven): **While** the template mirror and the local config both exist, byte-identity of the exposure block text shall hold between them where the local file has not been intentionally diverged (Template-First mirror parity; CLAUDE.local.md §2 + the rule_template_mirror_test.go pattern).

### P4 — 3-state framing corrected in comments and docs

**REQ-GET-010** (Ubiquitous): The package-level doc comment and the `glmCodingMaxOverrideAgents` / `CollapseClaudeEffortToGLM` doc comments in `internal/template/glm_effort_overlay.go` shall frame z.ai's reasoning control as 3 reachable states (`thinking-off` / `reasoning-high` / `reasoning-max`), and shall NOT frame it as a 2-tier `high`/`max` system.

**REQ-GET-011** (Ubiquitous): The `llm.yaml` exposure block introduced by REQ-GET-006 shall use the same 3-state framing, presenting `high`/`max` as the two REASONING-ON tiers and `thinking-off` as the mechanical tier below them (occupied by `manager-git` at Claude effort `low`).

**REQ-GET-012** (Event-detected): **When** a docs-site reference (README, docs-site content, or any Markdown file under the repo) is found by grep to frame GLM as a 2-tier `high`/`max` system, the SPEC shall correct it to the 3-state framing; **when** no such reference is found, the SPEC shall record the absence (absence-grep evidence) and make no docs-site change.

## §D. Constraints

- **C-1 Template-First (CLAUDE.local.md §2 [HARD])**: any template edit happens in `internal/template/templates/` FIRST, then `make build` recompiles. The local `.moai/config/sections/llm.yaml` is the rendered result and is updated in the same commit.
- **C-2 no runtime behavior change for P2**: the YAML exposure block MUST NOT become a parallel runtime path. The Go overlay (`glm_effort_overlay.go`) remains the SSOT for the collapse and override. If plan.md selects the "real struct field" option (design.md §B), the loader wiring MUST satisfy the `YAML_SECTION_NO_LOADER` + `CONFIG_STRUCT_YAML_MISMATCH` CI guards (settings-management.md § MoAI Configuration); if plan.md selects "comments-only", no struct/loader change is made.
- **C-3 honesty caveat (verification-claim-integrity.md §1.1 surface 3)**: the overlay's wire-effectiveness is "implemented + wired; live validation pending" — carried forward from MODEL-TIER-PLANTYPE-001. The P2 exposure wording MUST NOT overclaim.
- **C-4 P3 is Out-of-Scope**: the GLM wire-effectiveness live validation (the actual end-to-end test that the overlay's resolved state reaches z.ai and produces the expected behavior) is owned by SPEC-MODEL-TIER-PLANTYPE-001 follow-up #1 (`project_model_tier_plantype_001_followups.md`). This SPEC references it; it does NOT implement it.
- **C-5 case-sensitive agent names**: agent-name tokens in the override set are case-sensitive string-literal keys (`"manager-develop"`, `"builder-harness"`). The P1 edit removes the literal `"builder-harness"` key; no case-folding.
- **C-6 no frontmatter effort pin change**: this SPEC does NOT change any agent's `effort:` frontmatter or any cell in the `tierProfiles` matrix (`model_policy.go:294-323`). Those are settled design input owned by MODEL-TIER-PLANTYPE-001.

## §E. Out of Scope

### Out of Scope — P3 GLM wire-effectiveness live validation

- The end-to-end live validation that the overlay's resolved GLM reasoning state actually reaches z.ai and produces the expected wire behavior is **Out of Scope** for this SPEC. It is owned by **SPEC-MODEL-TIER-PLANTYPE-001 follow-up #1** (memory: `project_model_tier_plantype_001_followups.md`). This SPEC's `llm.yaml` exposure block (REQ-GET-006..009) carries forward the honesty caveat describing the wire as "implemented + wired; live-validation pending", but does NOT perform the live validation itself.

### Out of Scope — `tierProfiles` cell changes

- No cell in the `tierProfiles` matrix (`model_policy.go:294-323`) is changed. The P1 change is to the GLM overlay's override SET, not to the agent's Claude-side effort (which stays `high` for builder-harness). The matrix is settled design input owned by MODEL-TIER-PLANTYPE-001 and is referenced, not edited.

### Out of Scope — agent frontmatter `effort:` changes

- No agent file's `effort:` frontmatter field is touched. The P1 outcome is achieved entirely by removing one key from the GLM-side override map; the Claude-side effort that feeds the collapse is unchanged.

### Out of Scope — adding `builder-harness` to a different override tier

- The SPEC does not introduce a new "reasoning-medium override" or any new tier. Removing `builder-harness` from the coding-max set places it under the standard collapse (`high` → `reasoning-high`); no further classification is added.

### Out of Scope — per-spawn runtime channel for reasoning_effort

- The "delivery-granularity limitation" (per-agent overlay collapses to a session-global value; see `SessionGLMReasoningState()` at `glm_effort_overlay.go:172-174`) is documented but NOT addressed by this SPEC. A per-spawn wire channel is a separate concern.

## §F. References (sources read directly — not memory summaries)

- `internal/template/glm_effort_overlay.go:1-213` — full overlay implementation (CollapseClaudeEffortToGLM, glmCodingMaxOverrideAgents, ResolveGLMReasoning, ApplyGLMEffortOverlay, SessionGLMReasoningState, SessionGLMReasoningStateForEffort, IsGLMBackend)
- `internal/template/glm_effort_overlay_test.go:85-135` — existing test asserting the override set is exactly 2 members (the test P1 must update)
- `internal/template/model_policy.go:140-154` — EffortLevel constants
- `internal/template/model_policy.go:294-323` — tierProfiles map (builder-harness = high across all tiers; manager-git = low across all tiers)
- `.claude/rules/moai/development/agent-authoring.md:334-374` — Effort-Level Calibration Matrix
- `.moai/config/sections/llm.yaml` — current local config (no reasoning_effort surface)
- `internal/template/templates/.moai/config/sections/llm.yaml` — current template mirror (no reasoning_effort surface)
- `internal/cli/glm.go:242` + `internal/cli/launcher.go:820` — comment-only references to "coding-max default"; use `SessionGLMReasoningState()` (no set-membership assertion; not affected by P1)
- Memory: `reference_glm_reasoning_effort_2026_07.md`, `project_model_tier_plantype_001_completed.md`, `project_model_tier_plantype_001_followups.md`, `project_agent_token_cost_color_tiers.md`
- Cross-SPEC: SPEC-MODEL-TIER-PLANTYPE-001 (the parent overlay SPEC; P3 owner)

## §G. Acceptance Criteria Summary

Acceptance criteria live in `acceptance.md`. The minimum make-or-break AC set:

- **AC-GET-001**: `glmCodingMaxOverrideAgents` set == `{manager-develop}` (grep evidence, exactly one key)
- **AC-GET-003**: `ResolveGLMReasoning("builder-harness", "high")` returns state with `.Name == "reasoning-high"` (Go test evidence)
- **AC-GET-006**: `llm.yaml` exposure block present in BOTH `.moai/config/sections/llm.yaml` AND `internal/template/templates/.moai/config/sections/llm.yaml` (file-existence + content grep, both surfaces)
- **AC-GET-010**: overlay doc comments frame 3 reachable states, not 2-tier (content grep on `glm_effort_overlay.go`)

## §H. Cross-References

- `verification-claim-integrity.md` §1.1 surface 3 — the honesty caveat binding the P2 exposure wording (no unobserved-defect / no unobserved-success claim about wire-effectiveness)
- CLAUDE.local.md §2 — Template-First Rule binding P2
- CLAUDE.local.md §15 — CG Mode cost-reduction goal (the rationale for P1's exclude recommendation)
- `feedback_defect_claim_verification` — the methodological constraint that produced this SPEC's line-number citations
- SPEC-MODEL-TIER-PLANTYPE-001 — the parent SPEC (P3 owner; honesty-caveat origin)
