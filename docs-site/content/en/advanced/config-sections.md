---
title: Configuration Sections Reference
weight: 71
draft: false
description: "Key configuration file keys in .moai/config/sections/ (handoff/delegation/llm/statusline/security)."
---

MoAI-ADK project settings are split into several YAML files under `.moai/config/sections/`. While the [settings.json guide](/en/advanced/settings-json) covers Claude Code runtime settings, this page documents the keys of the main section files that control MoAI-ADK's own behavior.

{{< callout type="info" >}}
**One-line summary**: `settings.json` defines what to allow Claude Code, while `.moai/config/sections/*.yaml` defines how MoAI-ADK orchestrates.
{{< /callout >}}

## handoff.yaml — auto-resume handoff

Controls how saved handoffs are handled across session boundaries.

```yaml
handoff:
    mode: manual   # manual | auto
    guide: false
```

| Key | Value | Description |
|-----|-------|-------------|
| `mode` | `manual` (default) | Do not auto-inject saved handoffs (opt-in baseline UX) |
| `mode` | `auto` | Inject saved handoff into session context on `/clear`, then move to audit-trail copy |
| `guide` | `false` (default) | When `true`, emit a best-effort stderr hint about a pending handoff on non-`/clear` session starts (startup/resume/compact). Informational only, does not block the session |

Related: [Autonomous Continuation Loops](/en/advanced/autonomous-loops), [moai handoff](/en/cli-reference/handoff).

## delegation.yaml — agent routing SSOT

This is the default skill/agent assignment map per `/moai` subcommand. When orchestrator builds an execution plan (Analyze-First), it reads this map to decide which agents to spawn and which skills to inject.

```yaml
delegation:
    version: 1
    learning:
        observe: routing-ledger
        propose_via: harness-tier-ladder
        auto_apply: false          # Tier-4 gate — requires user approval
    subcommands:
        plan:
            agents: [manager-spec, plan-auditor, Explore]
            skills: [moai-workflow-spec, moai-foundation-thinking]
        # run / sync / project / fix / loop / ...
    domain_skills:
        backend:  [moai-ref-api-patterns, moai-domain-backend]
        security: [moai-ref-owasp-checklist, moai-ref-llm-security, ...]
    agents:
        manager-spec: [moai-workflow-spec, moai-foundation-thinking]
```

| Block | Description |
|-------|-------------|
| `learning` | Manages routing usage as an append-only ledger (`.moai/state/routing-ledger.jsonl`, opt-in·fail-open), and the harness learning subsystem proposes updates via a 4-tier ladder. `auto_apply: false` — Tier-4 changes require `AskUserQuestion` user approval |
| `subcommands` | Per-subcommand `agents` (11 retained agents to spawn) + `skills` (workflow skills to inject at spawn). 0 assignments is valid (orchestrator executes directly) |
| `domain_skills` | Skills to inject per mission domain (0-3 per spawn). Matched against domain signals |
| `agents` | Per-agent conditional skills (loaded on-demand when trigger fires) |

Related: [Agent Guide](/en/advanced/agent-guide), [Skill Guide](/en/advanced/skill-guide).

## llm.yaml — backend·profile matrix

Defines the profile, the profile matrix, per-agent overrides, and GLM model mappings.

```yaml
llm:
  profile: "medium"            # max | medium | low (active matrix column)
  performance_tier: "medium"   # legacy alias (read when profile absent, high→max)
  profiles:                    # profile column → 6 groups → {model, effort}
    max: { ... }               # detailed table: Profile Matrix page
    medium: { ... }
    low: { ... }
  agent_overrides: {}          # per-agent {model, effort} override (optional)
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.2"          # 1M context — Opus slot
      medium: "glm-4.7"        # 202K context — Sonnet slot
      low: "glm-4.5-air"       # 128K context — lightweight slot
      fable: "glm-5.2"
```

| Key | Description |
|-----|-------------|
| `profile` | Active profile matrix column (`max`/`medium`/`low`). An empty value is interpreted as `medium`. The model+effort source for every subagent spawn |
| `performance_tier` | Legacy alias field. Read only when `profile` is absent, normalized `high`→`max` |
| `profiles` | The group → `{model, effort}` matrix per profile column. The Go default (`template.DefaultProfileMatrix`) is the authoritative fallback for missing cells |
| `agent_overrides` | Per-canonical-agent-name `{model, effort}` override. Takes precedence over the active profile's group cell (catalog+enum validated) |
| `glm.base_url` | Z.AI Anthropic-compatible proxy endpoint |
| `glm.models` | Per-slot GLM model mapping. GLM collapses Claude's 5-step effort into 3 reasoning states (thinking-off / reasoning-high / reasoning-max) |

Related: [Profile Matrix](/en/advanced/profile-matrix), [3-Tier Agent Architecture](/en/advanced/no-haiku-3tier).

## statusline.yaml — status line

Controls statusline theme and 16 segment toggles.

```yaml
statusline:
  theme: "catppuccin-mocha"   # catppuccin-mocha | catppuccin-latte
  segments:
    model: true
    context: true
    # ... 16 segments total (all on by default)
    task: true
    pr: true
```

| Key | Description |
|-----|-------------|
| `theme` | Exactly 2 themes exist: `catppuccin-mocha` (default) or `catppuccin-latte` |
| `segments` | 16 individual segment toggles (the only runtime lever). All on by default; inactive states are handled gracefully with no output |

Segments are placed across 3 lines — line 1 (model·version·session meta), line 2 (context window·API usage bar), line 3 (directory·git·workflow·PR).

Related: [Statusline System & PR Segment](/en/advanced/statusline).

## security.yaml — security hardening

Additional security settings that **extend** (not replace) the built-in `DefaultSecurityPolicy` pattern. Follows SOLID's open-closed principle — config-only extension without core modifications.

```yaml
security:
  extra_dangerous_bash_patterns:
    - 'curl\s+.*\|\s*(ba)?sh'
    - 'rm\s+-rf\s+/[^.]'
  extra_deny_patterns: []
  extra_ask_patterns: []
  permission:
    strict_mode: true
    session_rules: []
  sandbox:
    required: false
    network_allowlist: []
    env_scrub_extra: []
    docker_image: "alpine:latest"
```

| Key | Description |
|-----|-------------|
| `extra_dangerous_bash_patterns` | Additional dangerous Bash command regex patterns (case-insensitive) **added to** built-in deny patterns |
| `extra_deny_patterns` / `extra_ask_patterns` | Additional file deny/ask patterns |
| `permission.strict_mode` | When `true`, rejects agent spawns in bypassPermissions mode |
| `sandbox.required` | When `true`, rejects `sandbox: none` agents without `sandbox.justification` (default false) |
| `sandbox.network_allowlist` | Network hosts **added to** the default 8 hosts |
| `sandbox.env_scrub_extra` | Env variable names **added to** default scrub list (AWS_*, GITHUB_TOKEN, etc.) |
| `sandbox.docker_image` | Default image for docker backend |

Related: [Security Notes](/en/advanced/security-notes), [settings.json Guide](/en/advanced/settings-json).

## Related Documentation

- [settings.json Guide](/en/advanced/settings-json) — Claude Code runtime settings
- [Harness Profiles & Evaluation](/en/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/en/cli-reference/doctor) — `moai doctor config` for merged configuration inspection
