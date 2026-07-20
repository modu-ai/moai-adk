---
id: SPEC-MODEL-TIER-PLANTYPE-001
title: "Research — GLM backend reasoning-control & effort-collapse design"
version: "0.3.1"
status: completed
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.1.x model-policy redesign"
module: "internal/config + internal/template + internal/cli + internal/tmux"
lifecycle: spec-anchored
tier: L
tags: "research, glm, reasoning-effort, thinking-mode, effort-overlay, z.ai"
---

# Research — GLM backend reasoning-control & effort-collapse design

> Scope note: this artifact was added at spec.md v0.3.0 to satisfy the Tier-L artifact set
> for the GLM backend effort-overlay dimension (§B.8). It documents ONLY the GLM
> reasoning-control findings and the collapse-design rationale. The plan_type / tier-profile
> research (v0.1.0–v0.2.0) is carried inline in plan.md §A and the design SSOT
> (`.moai/reports/model-tier-redesign-20260712.md`); it is not restated here.

## §A Problem framing

MoAI's plan_type tier profiles (spec.md §B.6/§B.7) assign each retained agent an atomic
`{model, effort}` pair, where `effort` uses the Claude 5-level vocabulary
(`low` / `medium` / `high` / `xhigh` / `max`; the shipped matrices use low/medium/high/xhigh).
Under a Claude backend this maps directly to Claude Code's per-sub-agent effort control.

Under a **GLM backend** (`moai glm` or `moai cg`), sub-agent model calls route to z.ai through
the Anthropic-compatibility shim (`base_url = https://api.z.ai/api/anthropic`, the constant
`DefaultGLMBaseURL` in `internal/config/defaults.go`). GLM (GLM-5.2 via z.ai) does **NOT**
implement Claude's 5-level effort vocabulary. Its reasoning control is a different, smaller
model. Sending Claude effort values verbatim to z.ai is therefore semantically undefined —
hence the need for a collapse overlay.

## §B GLM reasoning-control facts (z.ai official)

The following facts are drawn from z.ai's official documentation. They are the design input
for the collapse mapping; they are cited, not inferred.

1. **Thinking toggle is binary.** z.ai exposes `thinking: {type: enabled | disabled}` — a
   binary on/off switch for the model's reasoning phase. There is no graduated thinking
   budget analogous to Claude's effort levels.
   - Source: `https://docs.z.ai/guides/capabilities/thinking-mode` (z.ai — Thinking Mode).

2. **`reasoning_effort` has exactly two levels.** When thinking is enabled, `reasoning_effort`
   accepts ONLY `{high, max}`:
   - `max` is the DEFAULT (omitting `reasoning_effort` = `max`).
   - `high` must be requested EXPLICITLY (`reasoning_effort="high"`).
   - `reasoning_effort` has **NO effect when thinking is disabled** (the reasoning phase is
     off, so its depth control is moot).
   - Source: the GLM-5.2 OpenAI-compatible `reasoning_effort` guide (z.ai).

3. **z.ai vendor recommendation.** z.ai recommends `reasoning_effort: max` for **coding
   tasks**, and `reasoning_effort: high` for **fast, economical agent loops** where latency /
   token economy dominates.
   - Source: the GLM-5.2 `reasoning_effort` guide (z.ai) + Thinking Mode guide.

> Both z.ai source URLs are recorded here for the run-phase to re-verify (WebFetch) before
> relying on them as live facts. This artifact encodes the facts as of authoring (2026-07-12);
> it does not assert they are current at run time (verification-claim integrity — a cited fact
> is a hypothesis until re-verified).

## §C Collapse-design rationale (Claude 5-level → GLM 3-state)

The GLM canonical reasoning state has exactly **three** reachable values, from z.ai §B:
`thinking-off`, `reasoning-high`, `reasoning-max`. The collapse maps the 5 Claude effort
levels onto these three:

| Claude effort | GLM canonical state | Rationale |
|---------------|---------------------|-----------|
| `low` | `thinking: disabled` | `low` signals "fastest, least thorough" — the closest GLM analogue is to skip the reasoning phase entirely (thinking off). |
| `medium` | `reasoning_effort: high` | `medium` is the balanced default; GLM `high` is the economical-loop level — the lower of the two enabled-thinking levels. |
| `high` | `reasoning_effort: high` | `high` maps to GLM `high` (both are "deep but not maximal"); z.ai has no distinct level between `high` and `max`, so `medium` and `high` converge here. |
| `xhigh` | `reasoning_effort: max` | `xhigh` (Opus 4.7+ extended reasoning) is the strongest Claude level in the matrices; GLM `max` is its analogue. |
| `max` | `reasoning_effort: max` | `max` maps to GLM `max` directly (both the ceiling). |

Design consequences:

- The mapping is **lossy** (5 → 3): `medium`/`high` both collapse to `reasoning-high`, and
  `xhigh`/`max` both collapse to `reasoning-max`. This is inherent — z.ai offers fewer levels
  than Claude. Documented as an accepted limitation, not a defect.
- `thinking: disabled` is only reachable from Claude `low`. All other levels keep thinking on
  (z.ai fact §B.2: `reasoning_effort` is moot when thinking is off, so any enabled-thinking
  level requires thinking on).
- The mapping is **total** over the 5 Claude levels; an unrecognized effort string maps to a
  documented GLM default state (defensive completeness — the profile only emits the 5 levels,
  but the collapse guards against drift).

### Coding-max override

z.ai's explicit vendor recommendation (§B.3) is `reasoning_effort: max` for coding tasks.
MoAI's two code-producing retained agents are `manager-develop` (run-phase implementation) and
`builder-harness` (dynamic specialist generation). The overlay therefore forces these two
agents to `reasoning_effort: max` under a GLM backend, OVERRIDING their collapse result — e.g.
`manager-develop` at api/medium (opus / high) would collapse to `reasoning-high`, but the
override lifts it to `reasoning-max` per the z.ai coding recommendation. The override set is
exactly `{manager-develop, builder-harness}`; no other agent is a code-producer in the retained
catalog, so no other agent is overridden.

## §D The shim-injection risk (the unverified mechanism)

The collapse LOGIC above is deterministic and unit-testable. The open risk is the **delivery
mechanism**: does the collapsed GLM reasoning state actually reach z.ai?

`moai glm` / `moai cg` inject GLM configuration at the launch path
(`internal/cli/glm.go` — `setGLMEnv` for the direct process env, `injectGLMEnvForTeam` for the
`.claude/settings.local.json` `env` block). Today these inject `ANTHROPIC_BASE_URL` and the
`ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` model mapping, but **NO** reasoning/effort
control (`grep -c 'reasoning_effort' internal/cli/glm.go` → 0; `grep -c 'thinking' ...` → 0).

It is **UNVERIFIED** whether Claude Code's per-sub-agent `effort:` frontmatter is forwarded to
z.ai as `reasoning_effort` / `thinking` through the Anthropic-compat shim. Two branches:

- **Branch A — shim passthrough.** Claude Code translates the sub-agent effort frontmatter into
  the outbound request's `reasoning_effort` / `thinking` fields when `ANTHROPIC_BASE_URL` points
  at z.ai. If so, the overlay adjusts the effort representation pre-launch and relies on the
  passthrough. Lowest cost.
- **Branch B — explicit write.** Claude Code does NOT translate effort→reasoning control for the
  z.ai shim. Then MoAI must write `reasoning_effort` and/or the `thinking` toggle explicitly at
  the GLM launch path (env var and/or settings.local.json `env`), mirroring the existing
  `ANTHROPIC_DEFAULT_*` injection. On Branch B the new key(s) must be added to
  `buildTmuxClearVars()` so `moai cc` teardown clears them (the inject↔clear parity invariant,
  REQ-CGH-009).

### Delivery-granularity sub-risk

The overlay is designed to be **per-agent** (each agent gets its own collapsed effort, and the
coding-max override is per-agent). But env vars and the settings.local.json `env` block are
**session-global**. If Branch B holds AND no per-agent reasoning-control channel exists through
the shim, the overlay's per-agent granularity is not deliverable — the run-phase would then
record a documented limitation: the session-level `reasoning_effort` is derived from the overlay
(a coding agent as the active spawn ⇒ session `reasoning_effort: max` via the override), and the
per-agent collapse logic remains defined and unit-tested even though the wire only carries a
session-level value.

### Why this is a run-phase gate, not a plan-phase clarification

Distinguishing Branch A from Branch B requires **running a GLM session and observing the
outbound request to z.ai** — an empirical measurement, not a user preference. The SPEC states
both branches; the run-phase MUST determine which holds and MUST record the observed evidence
before claiming the overlay is effective (verification-claim integrity: "the function is called"
is NOT "the effort reached z.ai as reasoning_effort"). This is the D7 decision (plan.md) and the
AC-MTP-032b gate (acceptance.md).

## §E Codebase anchors (measured 2026-07-12)

| Concern | Location | Note |
|---------|----------|------|
| GLM `team_mode` field (REAL signal) | `internal/config/types.go` `LLMConfig.TeamMode` (yaml `llm.team_mode`) | `moai glm` writes `"glm"` AND `moai cg` writes `"cg"` — both via `persistTeamMode` (glm.go:519 `llmCfg.TeamMode = mode`); the M5 predicate's primary signal; godoc comment stale — see plan.md §A.10 |
| GLM `mode` field (DORMANT) | `internal/config/types.go` `LLMConfig.Mode` (yaml `llm.mode`) | **No writer / no reader** in non-test Go (verified 2026-07-12) — `moai glm` does NOT write `mode="glm"` (it writes `team_mode="glm"`). Reserved; the M5 predicate keeps `mode=="glm"` only as a defensive OR for this dormant/future field |
| Runtime CG detector (SSOT) | `internal/tmux/cg_detect.go` `IsCGMode` | reads `team_mode == "cg"` + tmux GLM markers; the M5 predicate cross-references it |
| Claude effort constants | `internal/template/model_policy.go` `EffortLevelLow/Medium/High/XHigh/Max` | collapse INPUT domain |
| GLM env injection (direct) | `internal/cli/glm.go` `setGLMEnv` | process `os.Setenv`; proposed overlay wire point |
| GLM env injection (team) | `internal/cli/glm.go` `injectGLMEnvForTeam` | settings.local.json `env`; proposed overlay wire point |
| `moai cc` teardown | `internal/cli/glm.go` `buildTmuxClearVars` | Branch-B parity target |
| Shim base URL | `internal/config/defaults.go` `DefaultGLMBaseURL` | `https://api.z.ai/api/anthropic` |
| tier→GLM model mapping | `setGLMEnv` / `injectGLMEnvForTeam` `ANTHROPIC_DEFAULT_*` | model dimension — overlay does NOT touch this |

## §F Cross-references

- spec.md §B.8 (GLM overlay REQ-MTP-026..030) + §A.5 (baseline invariants).
- plan.md §A.10 (injection-surface evidence) + §B D7 (injection mechanism decision) + §F M5.
- acceptance.md M5 (AC-MTP-028..032) + Scenario 5.
- `.claude/rules/moai/core/glm-web-tooling.md` § CG Mode (`team_mode` SSOT + field
  disambiguation).
- Adjacent: SPEC-GLM-MODEL-ALLOWLIST-001 (completed — GLM model allowlist; untouched here).
- z.ai sources (re-verify at run time via WebFetch):
  `https://docs.z.ai/guides/capabilities/thinking-mode`; GLM-5.2 OpenAI-compat
  `reasoning_effort` guide.
