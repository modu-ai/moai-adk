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
- haiku: Claude Haiku (fastest, lowest cost)

Current model generation mapping:
- opus = Opus 4.8 (default effort: high across all surfaces incl. Claude Code; set xhigh explicitly for coding/agentic work)
- sonnet = Sonnet (current generation)
- haiku = Haiku (current generation)

Opus 4.8 serves the full 1M token context window by default (no beta header, no long-context premium). Fast mode (speed: "fast") is a research preview for higher output throughput.

Invalid values (NEVER use):
- glm: Not a model field value (GLM is configured via environment variables)
- high/medium/low: These are CLI policy flags, not model field values
- Pinned old versions (opus-4-0, opus-4-1, sonnet-4-5): Auto-migrated to current generation
- Full model-ID form (e.g., `claude-opus-4-8`): **official-but-intentionally-disallowed in MoAI.** Claude Code itself accepts a full model-ID string in the `model` field, but MoAI intentionally restricts agents to the four alias values above (`inherit` / `opus` / `sonnet` / `haiku`). The reason is the `[1m]` context-entitlement inheritance bug (see § Inherit-by-Default Convention): a subagent that pins a concrete full model ID — like an explicit `model: sonnet` / `model: opus` — does not inherit the parent session's `[1m]` entitlement and fails to spawn. The alias `inherit` sidesteps this. This restriction being a deliberate MoAI policy (not a stale gap) means "MoAI is outdated relative to Claude Code" is a misreading — the full-ID form is omitted on purpose.

## Inherit-by-Default Convention

[ZONE:Evolvable] [HARD] All MoAI agents SHOULD declare `model: inherit` unless explicitly assigned `haiku` for speed-critical tasks. The previous practice of declaring `model: sonnet` or `model: opus` is deprecated for new agents.

Rationale (Claude Code session inheritance bug):
- When the parent session uses an `[1m]` context variant (e.g., `claude-opus-4-7[1m]`, `claude-sonnet-4-6[1m]`) and a spawned subagent declares an explicit `model: sonnet` or `model: opus` in its frontmatter, the parent's 1M context entitlement does NOT propagate to the subagent.
- Result: subagent spawn fails with `API Error: Usage credits required for 1M context · run /usage-credits to turn them on, or /model to switch to standard context`.
- Sonnet 1M specifically requires extra usage credits on every plan (including Max), so the entitlement mismatch is unrecoverable mid-spawn.

Upstream tracking (Anthropic claude-code repository):
- [Issue #45847](https://github.com/anthropics/claude-code/issues/45847): skill with `model: sonnet` frontmatter fails from Opus 4.6/4.7 [1m] parent
- [Issue #51060](https://github.com/anthropics/claude-code/issues/51060): subagent with `model: opus` ignores parent's Extra Usage flag
- [Issue #36670](https://github.com/anthropics/claude-code/issues/36670): Team teammates do not inherit the `[1m]` context variant from leader

Workaround pattern (`model: inherit`):
- The subagent fully inherits the parent's model + context entitlement, eliminating the mismatch.
- Reference implementation: `.claude/agents/moai/plan-auditor.md` has used `model: inherit`.
- All package agents under `.claude/agents/moai/` (7 retained agents per FLAT v.2.x baseline) declare `model: inherit`, except `manager-docs` and `manager-git` which use `model: haiku`.

Exceptions (do NOT migrate to inherit):
- `model: haiku` agents (`manager-docs`, `manager-git`) — Haiku has no `[1m]` variant, so the bug does NOT apply. Speed-critical agents should remain on `haiku` for cost and latency.
- Documentation/example YAML inside skill bodies (`.claude/skills/moai-foundation-cc/reference/**/*.md`) — these mirror official Claude Code documentation and MUST show all valid values (`sonnet`, `opus`, `haiku`, `inherit`) for educational purposes.

## `[1m]` Constraint Re-Verification (CC 2.1.178)

The `[1m]` entitlement-inheritance constraint was re-verified against CC 2.1.178 (2026-06-16, M1 research milestone). **Verdict: STILL-ACTIVE (conservative).** Per-agent `model:` pins remain forbidden regardless of this verdict (the re-verification SPEC records per-agent pinning as out-of-scope).

Evidence fetched 2026-06-16 via the GitHub issue API + the canonical CC CHANGELOG:

- Issue #45847 (skill with `model:` fails from `[1m]` parent): **closed** (2026-04-13), labeled `duplicate` — no explicit "fixed" resolution.
- Issue #51060 (subagent `model: opus` spawn fails): **closed** (2026-05-26), labeled `bug, area:model, area:agents, stale` — no CHANGELOG entry fixes the spawn-time entitlement-inheritance root cause.
- Issue #36670 (Team teammates don't inherit `[1m]` from leader): **OPEN** (updated 2026-06-02) — the Team-mode path is confirmed unfixed at CC 2.1.178.
- CC 2.1.172 fixes ("1M context stuck session", "doubled `[1m]` suffix") address the *symptom* and *suffix normalization*, NOT the *spawn-time entitlement mismatch*. CC 2.1.173/2.1.174 are Fable-5-suffix and background-env-inheritance fixes — orthogonal.

Because the Team-mode path (#36670) is open and no CHANGELOG resolves the single-spawn root cause, the constraint is treated as still-active. A follow-up SPEC (conditional) MAY re-enable per-agent pinning only when #36670 is closed-with-fixed AND a CHANGELOG confirms Team `[1m]` inheritance for explicit `model:` teammates.

## Default-Model Cost Lever (CC 2.1.175)

[ZONE:Evolvable] [HARD] The `[1m]`-safe cost lever is the **Default-model** routed via `availableModels` + `enforceAvailableModels`, NOT per-agent `model:` pins. The deployed `settings.json` template (`.claude/settings.json.tmpl`) sets:

```json
"model": "sonnet",
"availableModels": ["sonnet", "opus", "haiku"],
"enforceAvailableModels": true
```

Semantics (CC 2.1.175 CHANGELOG verbatim): _"Added `enforceAvailableModels` managed setting — when enabled, the `availableModels` allowlist also constrains the Default model (a Default that would resolve to a disallowed model now falls back to the first allowed model)"_. CC 2.1.176 further tightens enforcement: alias model picks can no longer be redirected to a blocked model via `ANTHROPIC_DEFAULT_*_MODEL` env vars.

Why this is `[1m]`-safe: the lever operates on the **Default** model resolution at the settings level, not on per-agent explicit pins, so it does not trigger the spawn-time entitlement-inheritance failure (#45847/#51060/#36670). The cost-routing thesis (route the busy-agent cost through Sonnet, not Opus) flows through the Default; deep-reasoning exceptions use per-spawn `Agent(model: "opus")` only for the 5-10% of tasks where Opus wins (architecture, complex perf) — and even those inherit the parent `[1m]` entitlement because they are spawned without a frontmatter `model:` pin (the per-spawn `model` parameter is a runtime arg, distinct from the frontmatter field that triggers the bug).

## Model Policy Tiers

Model policy is set via `moai init --model-policy <tier>`:

| Tier | Description | Opus Agents | Sonnet Agents | Haiku Agents |
|------|-------------|-------------|---------------|--------------|
| high | Maximum quality | spec, strategy, security | backend, frontend, ddd, tdd | quality, git, researcher |
| medium | Balanced (default) | spec, strategy, security | backend, frontend, ddd, tdd | quality, git, researcher |
| low | Cost optimized | None | spec, strategy, security | All others |

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

Note: `ultrathink` is a Claude Code one-turn keyword that requests deeper reasoning for that prompt; MoAI standardizes it to `effort: xhigh` (the coding/agentic level above) for that turn.

## Rules

- Agent `model` field must be one of: inherit, opus, sonnet, haiku
- [ZONE:Evolvable] [HARD] New agent definitions SHOULD use `model: inherit` (default); explicit `sonnet`/`opus` are deprecated due to Claude Code Issue #45847/#51060 (see Inherit-by-Default Convention)
- `model: haiku` remains valid for speed-critical agents (immune to the [1m] entitlement bug)
- GLM is configured via env vars in settings.json, never via model field
- Model policy tier is a CLI concern, not an agent definition concern
- CG Mode uses tmux session-level env isolation for model routing
- Old model versions are auto-migrated: do not pin to specific version IDs
