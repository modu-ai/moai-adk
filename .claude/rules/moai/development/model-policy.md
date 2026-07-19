---
paths: "**/.claude/agents/**"
---

# Model Policy

Rules for agent model field values and multi-model architecture.

## Valid Model Field Values

Agent definition `model` field accepts only these values:
- inherit: Uses parent session's model (default)
- opus: Claude Opus (highest capability)
- sonnet: Claude Sonnet (balanced)
- fable: Claude Fable (current generation; added to the model enum per CC v2.1.196 model-priority update)
- haiku: Claude Haiku (fastest, lowest cost)

Current model generation mapping:
- opus = Opus 4.8 (default effort: high across all surfaces incl. Claude Code; set xhigh explicitly for coding/agentic work)
- sonnet = Sonnet 5 (current generation; single 1M-token context window — no 200K variant)
- fable = Fable (current generation; added to the model enum per CC v2.1.196 model-priority update)
- haiku = Haiku (current generation; retired from MoAI agent routing per the No-Haiku policy — value remains valid for documentation/example YAML)

Opus 4.8 serves the full 1M token context window by default (no beta header, no long-context premium). Fast mode (speed: "fast") is a research preview for higher output throughput. Explore (Anthropic built-in) inherits the session model per CC v2.1.198 — no separate deployment or model pin needed.

Invalid values (NEVER use):
- glm: Not a model field value (GLM is configured via environment variables)
- high/medium/low: These are CLI policy flags, not model field values
- Pinned old versions (opus-4-0, opus-4-1, sonnet-4-5): Auto-migrated to current generation
- Full model-ID form (e.g., `claude-opus-4-8`): **official-but-intentionally-disallowed in MoAI.** Claude Code itself accepts a full model-ID string in the `model` field, but MoAI intentionally restricts agents to the five alias values above (`inherit` / `opus` / `sonnet` / `fable` / `haiku`). The reason is the `[1m]` context-entitlement inheritance bug (see § Inherit-by-Default Convention): a subagent that pins a concrete full model ID — like an explicit `model: sonnet` / `model: opus` — does not inherit the parent session's `[1m]` entitlement and fails to spawn. The alias `inherit` sidesteps this. This restriction being a deliberate MoAI policy (not a stale gap) means "MoAI is outdated relative to Claude Code" is a misreading — the full-ID form is omitted on purpose.

## Inherit-by-Default Convention

[ZONE:Evolvable] [HARD] All MoAI agents SHOULD declare `model: inherit` unless explicitly assigned `haiku` for speed-critical tasks. The previous practice of declaring `model: sonnet` or `model: opus` is deprecated for new agents.

Rationale (Claude Code session inheritance bug):
- When the parent session uses an `[1m]` context variant (e.g., `claude-opus-4-8[1m]`, `claude-sonnet-5[1m]`) and a spawned subagent declares an explicit `model: sonnet` or `model: opus` in its frontmatter, the parent's 1M context entitlement does NOT propagate to the subagent.
- Result: subagent spawn fails with `API Error: Usage credits required for 1M context · run /usage-credits to turn them on, or /model to switch to standard context`.
- Sonnet 1M specifically requires extra usage credits on every plan (including Max), so the entitlement mismatch is unrecoverable mid-spawn.

Upstream tracking (Anthropic claude-code repository):
- [Issue #45847](https://github.com/anthropics/claude-code/issues/45847): skill with `model: sonnet` frontmatter fails from Opus 4.6/4.7 [1m] parent
- [Issue #51060](https://github.com/anthropics/claude-code/issues/51060): subagent with `model: opus` ignores parent's Extra Usage flag
- [Issue #36670](https://github.com/anthropics/claude-code/issues/36670): Team teammates do not inherit the `[1m]` context variant from leader

Workaround pattern (`model: inherit`):
- The subagent fully inherits the parent's model + context entitlement, eliminating the mismatch.
- Reference implementation: `.claude/agents/moai/plan-auditor.md` has used `model: inherit`.
- All 10 MoAI-custom retained agents under `.claude/agents/moai/` declare `model: inherit` (per the 11-agent catalog: 10 MoAI-custom + 1 Anthropic built-in `Explore`, aligned with CLAUDE.md §4). The No-Haiku policy (SPEC-AGENT-ARCH-V2-001 §D) retired the former `model: haiku` exception — `manager-docs` and `manager-git` moved from `model: haiku` to `model: sonnet` with `effort: low` (cost reduction via effort tiering, not model-class substitution).

Exceptions (do NOT migrate to inherit):
- Documentation/example YAML inside skill bodies (`.claude/skills/moai-foundation-cc/reference/**/*.md`) — these mirror official Claude Code documentation and MUST show all valid values (`sonnet`, `opus`, `haiku`, `inherit`) for educational purposes.

## Baseline-Refill Breaker (team sonnet — second failure mode; Sonnet 5 / Opus 4.8-resolved)

[ZONE:Evolvable] The `[1m]` entitlement bug in § Inherit-by-Default Convention is the *spawn-time* failure mode (a frontmatter `model: sonnet` pin → spawn fails with a 1M credit error). A **distinct second failure mode** historically affected team-mode teammates spawned via per-spawn `model: "sonnet"` override:

| failure mode | trigger | symptom | mitigation |
|--------------|---------|---------|------------|
| `[1m]` credit-fail | frontmatter `model: sonnet` pin | spawn fails immediately (`Usage credits required for 1M`) | use `model: inherit` |
| baseline-refill breaker | per-spawn `model: "sonnet"` in team mode on a **200K-variant** model | spawn succeeds, but the 200K window saturates under the heavy baseline → autocompact → rapid-refill → runtime circuit breaker → zero output | historical only — see resolution below |

**Resolution (Sonnet 5 / Opus 4.8 era):** the breaker required a teammate to fall back to a **200K context variant** after the `[1m]` suffix was stripped on teammate spawn (Anthropic issues #36670 / #34421; the suffix-strip mechanism is still OPEN upstream). The fallback target — a 200K context variant — **no longer exists in the current default lineup**: Sonnet 5 ships a single 1M-token context window (1M is both default and maximum; no smaller variant — per platform.claude.com Sonnet 5 model docs), and Opus 4.8 likewise serves the full 1M window by default. With no 200K variant to fall back to, a teammate spawned as `sonnet` / `opus` operates at 1M regardless of suffix stripping, and the rapid-refill → circuit-breaker → zero-output cascade cannot trigger. The mechanism (#36670) is technically still open but its observable impact on the current default lineup is zero.

The breaker therefore remains documented only as a **historical hazard for legacy 200K-variant models** (Sonnet 4.x, Opus 4.6, and Haiku 4.5 which is still 200K): on those models a teammate can still fall back to 200K. For the current default lineup the operational mitigation (single `manager-develop` + Milestone split over team mode) is no longer forced by the breaker — though team mode is additionally disabled by default per the Phase 4 re-design (`.claude/rules/moai/workflow/orchestration-mode-selection.md`), in favor of subagent fanout (Mode 4) for multi-domain research/review and sequential sub-agent (Mode 5) for coding.

## `[1m]` Constraint Re-Verification (CC 2.1.178; Sonnet 5 / Opus 4.8 practical-impact update)

The `[1m]` entitlement-inheritance constraint was re-verified against CC 2.1.178. **Verdict: STILL-ACTIVE mechanism, but ZERO PRACTICAL IMPACT on the current default lineup (Sonnet 5 / Opus 4.8).** Per-agent `model:` pins remain forbidden regardless (the re-verification records per-agent pinning as out-of-scope).

Evidence fetched via the GitHub issue API + the canonical CC CHANGELOG:

- Issue #45847 (skill with `model:` fails from `[1m]` parent): **closed**, labeled `duplicate` — no explicit "fixed" resolution.
- Issue #51060 (subagent `model: opus` spawn fails): **closed**, labeled `bug, area:model, area:agents, stale` — no CHANGELOG entry fixes the spawn-time entitlement-inheritance root cause.
- Issue #36670 (Team teammates don't inherit `[1m]` from leader): **OPEN** — the Team-mode path is confirmed unfixed at CC 2.1.178.
- CC 2.1.172 fixes ("1M context stuck session", "doubled `[1m]` suffix") address the *symptom* and *suffix normalization*, NOT the *spawn-time entitlement mismatch*. CC 2.1.173/2.1.174 are Fable-5-suffix and background-env-inheritance fixes — orthogonal.

**Why STILL-ACTIVE mechanism ≠ practical impact (Sonnet 5 / Opus 4.8 era):** the #36670 mechanism strips the `[1m]` suffix on teammate spawn, which historically forced a fallback to a 200K variant. Sonnet 5 and Opus 4.8 have **no 200K variant** — their context window is 1M as both default and maximum (per platform.claude.com Sonnet 5 model docs; Opus 4.8 serves the full 1M window by default). The stripped-suffix teammate therefore still resolves to 1M; there is no smaller variant to degrade into. The mechanism is unfixed upstream, but on the current default lineup it has nothing to break. The `model: inherit` convention is retained as defense-in-depth and for legacy-200K-variant models (Haiku 4.5 still ships 200K).

A follow-up SPEC (conditional) MAY re-enable per-agent pinning only when #36670 is closed-with-fixed AND a CHANGELOG confirms Team `[1m]` inheritance for explicit `model:` teammates — though for Sonnet 5 / Opus 4.8 the practical case for that re-enablement has dissolved.

## Default-Model Cost Lever (Default = sonnet, no allowlist enforcement)

[ZONE:Evolvable] [HARD] The `[1m]`-safe cost lever is the **Default model** set at the settings level, NOT per-agent `model:` pins. The deployed `settings.json` template (`.claude/settings.json.tmpl`) sets ONLY:

```json
"model": "sonnet"
```

The template deliberately does **NOT** set `availableModels` or `enforceAvailableModels`. A closed `availableModels` allowlist combined with `enforceAvailableModels: true` hides any model not in the list from the `/model` picker (CC v2.1.172 behavior), which caused two problems:

1. **New-model lockout** — every new Claude model (for example a new `fable` generation, or any future tier) was invisible in `/model` until an operator manually appended it to the allowlist. This recurred on every model release.
2. **GLM allowlist maintenance** — enforcement forced every GLM swap target (`glm-5.2` and the other GLM tiers) to be enumerated in the allowlist, or the swap was declined (see GLM-mode reconciliation below).

Dropping `enforceAvailableModels` resolves both at once: all Claude models (current and future) auto-appear in the picker with no maintenance, and the GLM swap is admitted without an allowlist. Only the Default-model cost lever is retained — `"model": "sonnet"` alone still routes the busy-agent cost through Sonnet by default.

Why this is `[1m]`-safe: the lever operates on the **Default** model resolution at the settings level, not on per-agent explicit pins, so it does not trigger the spawn-time entitlement-inheritance failure (#45847/#51060/#36670). The cost-routing thesis (route the busy-agent cost through Sonnet, not Opus) flows through the Default; deep-reasoning exceptions use per-spawn `Agent(model: "opus")` only for the 5-10% of tasks where Opus wins (architecture, complex perf) — and even those inherit the parent `[1m]` entitlement because they are spawned without a frontmatter `model:` pin (the per-spawn `model` parameter is a runtime arg, distinct from the frontmatter field that triggers the bug).

### GLM-mode reconciliation

[ZONE:Evolvable] [HARD] With `enforceAvailableModels` unset, GLM mode needs no allowlist reconciliation. When GLM mode is active (`moai glm` whole-session, or the GLM teammate panes of `moai cg`), the GLM activation sets `ANTHROPIC_DEFAULT_OPUS_MODEL` to the configured GLM high model (default `glm-5.2`), surfaced in the model UI as the Opus-slot alias. The CC 2.1.176 redirect-blocking semantics — which decline an `ANTHROPIC_DEFAULT_*_MODEL` redirect to a model NOT in `availableModels` — apply ONLY when `enforceAvailableModels` is `true`. Because the template no longer sets that flag, the GLM swap is never checked against an allowlist and is admitted directly; the session runs on the configured GLM model instead of silently falling back to Sonnet.

This supersedes the earlier approach of enumerating the GLM model ids in `availableModels` (the `[1m]`-variant + raw-GLM-id expansion). That expansion existed only to satisfy `enforceAvailableModels: true`; removing the enforcement flag removes the need for the expansion entirely. The Default model stays `sonnet` — a non-GLM (`moai cc` / plain Claude) session still resolves its Default to Sonnet; the only change is that no model is hidden and no swap is declined.

Scope note: this is a **static template change** in `.claude/settings.json.tmpl` (removal of the `availableModels` + `enforceAvailableModels` keys). It touches no Go runtime code (`glm.go` / `launcher.go` / `settings.go` unchanged) and writes nothing to `settings.local.json` — so the solo `moai glm` "settings.local.json clean" design (no GLM env leak to subsequent plain-`claude` invocations) is preserved.

## Model Policy Tiers (3-tier — max/medium/low)

Model policy is set via `moai init --model-policy <tier>`. The 3-tier system (`max` / `medium` / `low`) replaces the legacy `high` / `medium` / `low` model-class tiers. Under the No-Haiku policy (SPEC-AGENT-ARCH-V2-001 §D), all workers are Sonnet 5 fixed across all tiers; the tier governs only two axes — (a) where Opus is deployed and (b) how aggressively Sonnet effort is lowered. The authoritative agent×tier matrix is the **§2-B table** (design.md §D.3); the Sonnet effort criteria is the **§2-C table** (design.md §D.4). The `model_routing_profiles.{max,medium,low}` matrices in `workflow.yaml` are the 3-tier config SSOT.

| Tier | Philosophy | Opus deployment | Sonnet effort baseline |
|------|------------|-----------------|------------------------|
| `max` | Opus-centric reasoning + Sonnet workers — quality first | Orchestrator · super-advisor · manager-spec(plan) · plan/sync-auditor | Implementation xhigh; procedural tasks low/medium |
| `medium` (default) | Sonnet-centric, minimal Opus — balanced | super-advisor + Tier L plan — 2 points only (on-demand) | Implementation high~xhigh; docs/procedural low~medium |
| `low` | Sonnet single + effort tiering — cost minimum | None (Opus 0) — consultation also Sonnet xhigh | One tier down across the board: implementation high, docs/procedural low |

The per-agent model+effort values for each cell are in the §2-B agent×tier matrix (design.md §D.3). Former haiku slots (docs sync, mx tagging, git procedures) are replaced by `sonnet / low` — cost reduction via effort tiering, not model-class substitution.

## Per-Agent Profile Resolver (model injection source)

The per-agent model+effort **profile** (config `llm.profile` ∈ {max, medium, low}) is the runtime-arg **model** injection source the orchestrator reads at spawn time. Query it with the read-only accessor:

```bash
moai model profile          # human table
moai model profile --json   # machine-readable
```

The resolver maps the active profile + optional `llm.agent_overrides` to each retained agent's `{model, effort}` via the agent-GROUP matrix (`spec_auditors` / `develop` / `advisor` / `design_harness_e2e` / `docs` / `git`); `Explore` and any ungrouped agent resolve to `inherit`. Precedence: `agent_overrides[agent]` → active profile group cell → Go-default group cell → `inherit`.

The resolved **model** is the value the orchestrator injects as a per-spawn `Agent(model: <alias>)` runtime arg — `[1m]`-safe and distinct from the frontmatter `model:` field (see § Inherit-by-Default Convention), so a profile change never re-introduces the concrete-frontmatter-`model:` spawn-failure risk. Agent `.md` frontmatter stays at `model: inherit`; no init / update / web save mutates it.

The resolved **effort** is *documented intent* for a NAMED subagent: the Agent/Task tool accepts a per-spawn `model` arg but NOT a per-spawn `effort` arg for a named subagent, so the effort is consumed only through (a) the agent-frontmatter effort default (session-scoped), (b) the GLM effort overlay (`CollapseClaudeEffortToGLM` + the `manager-develop` coding-max override), and (c) Workflow-script / `Agent(general-purpose)` prompt-level effort steering. Where the profile effort diverges from the frontmatter doc-canonical effort, the frontmatter effort remains the effective named-subagent spawn effort; the divergence is intentional, not a config or lint error.

## CG Mode

CG Mode (Claude + GLM) uses environment variable overrides, not model field changes:
- Leader session: Uses Claude models (no GLM env)
- Teammate sessions: Inherit GLM env from tmux session
- Activation: `moai cg` (requires tmux)

## Effort Levels

Claude models support effort levels that control reasoning depth (Opus 4.8 calibration):
- xhigh: best setting for coding and agentic use cases
- high: default on Opus 4.8 across all surfaces; minimum for intelligence-sensitive work
- medium: cost-sensitive work that can trade off intelligence
- low: short, scoped, latency-sensitive tasks

The per-agent effort defaults across the 3-tier system (max/medium/low) are governed by the §2-B agent×tier matrix (design.md §D.3) and the §2-C Sonnet effort criteria (design.md §D.4). The `model_routing_profiles` cells in `workflow.yaml` carry the `{model, effort}` pair for each Tier-Phase × perfTier combination.

Note: `ultrathink` is a Claude Code one-turn keyword that requests deeper reasoning for that prompt; MoAI standardizes it to `effort: xhigh` (the coding/agentic level above) for that turn.

## Rules

- Agent `model` field must be one of: inherit, opus, sonnet, fable, haiku
- [ZONE:Evolvable] [HARD] New agent definitions SHOULD use `model: inherit` (default); explicit `sonnet`/`opus` are deprecated due to Claude Code Issue #45847/#51060 (see Inherit-by-Default Convention)
- `model: haiku` is retired from MoAI agent routing per the No-Haiku policy (SPEC-AGENT-ARCH-V2-001 §D); the HaikuResidualRule lint enforces 0 haiku references in agent frontmatter, claude_models, model_routing_profiles, workflow_agents, and role_profiles. Former haiku slots use `sonnet` with `effort: low`.
- GLM is configured via env vars in settings.json, never via model field
- Model policy tier (max/medium/low) is a CLI concern, not an agent definition concern
- CG Mode uses tmux session-level env isolation for model routing
- Old model versions are auto-migrated: do not pin to specific version IDs
