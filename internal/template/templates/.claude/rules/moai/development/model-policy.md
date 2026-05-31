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
