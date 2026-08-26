---
title: Configuration Sections Reference
weight: 71
draft: false
description: "Key configuration file keys in .moai/config/sections/ (handoff/delegation/llm/statusline/security/workflow/crosssession)."
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
  profile: "medium"            # high | medium | low (active matrix column; max read as high)
  performance_tier: "medium"   # legacy alias (read when profile absent; same vocabulary)
  profiles:                    # profile column → 11 agents → {model, effort}
    high: { ... }              # detailed table: Profile Matrix page
    medium: { ... }
    low: { ... }
  agent_overrides: {}          # per-agent {model, effort} override (optional)
  glm:
    base_url: "https://api.z.ai/api/anthropic"
    models:
      high: "glm-5.3"          # 1M context — Opus slot
      medium: "glm-5.3"        # 1M context   — Sonnet slot
      low: "glm-5.3"          # 1M context   — lightweight slot
      fable: "glm-5.3"
```

| Key | Description |
|-----|-------------|
| `profile` | Active profile matrix column (`high`/`medium`/`low`; the former `max` is read as an alias of `high`). An empty value is interpreted as `medium`. The model+effort source for every subagent spawn |
| `performance_tier` | Legacy alias field. Read only when `profile` is absent; shares the same `high`/`medium`/`low` vocabulary, so no normalization step is needed |
| `profiles` | The per-agent → `{model, effort}` matrix per profile column (11 agents × 3 columns = 33 cells). The Go default (`template.DefaultProfileMatrix`) is the authoritative fallback for missing cells |
| `agent_overrides` | Per-canonical-agent-name `{model, effort}` override. Takes precedence over the active profile's agent cell (catalog+enum validated) |
| `glm.base_url` | Z.AI Anthropic-compatible proxy endpoint |
| `glm.models` | Per-slot GLM model mapping. GLM collapses Claude's 5-step effort into 3 reasoning states (thinking-off / reasoning-high / reasoning-max) |

Related: [Profile Matrix](/en/advanced/profile-matrix), [3-Tier Agent Architecture](/en/advanced/no-haiku-3tier).

## statusline.yaml — status line

Controls the statusline theme, which hosting service the `github` segment counts against, and 16 segment toggles.

```yaml
statusline:
  theme: "catppuccin-mocha"   # catppuccin-mocha | catppuccin-latte
  # forge: gitlab             # github | gitlab | none (unset detects from the origin host)
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
| `forge` | One of `github` · `gitlab` · `none` — which hosting service the `github` segment counts open work against. Leave it unset and the value is decided from the origin remote's host: `github.com` selects `gh`, `gitlab.com` selects `glab`. A self-hosted instance carries no signal in its name, so set it explicitly there. An unrecognised value renders nothing rather than falling back to detection, so a typo shows as an absent segment instead of a wrong count |
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

## workflow.yaml — branch_guard

An opt-in guard that protects branch state in the primary checkout. When several sessions share one checkout, a `git switch`, `git checkout`, `git reset --hard`, `git stash` or `git rebase` run by one of them changes another session's working tree with no signal to either side. This guard refuses those commands, and only in the primary checkout.

```yaml
workflow:
    branch_guard:
        enabled: false   # distributed default
```

| Key | Value | Description |
|-----|-------|-------------|
| `enabled` | `false` (default) | The guard is fully inert. It does not even run the `git rev-parse` needed to classify the checkout, so it costs nothing |
| `enabled` | `true` | Branch-state-changing commands are refused in the primary checkout. Inside a worktree they are allowed as before |

**Why it ships off.** The hazard this guard addresses only exists when several sessions share one checkout. It does not arise in a single-developer repository, so the distributed build ships with the guard disabled. A maintainer running several sessions at once writes the key above to turn it on.

**Scope.** The guard distinguishes the primary checkout from a worktree and does not block branch operations inside a worktree. Read-only commands such as `git status`, `git log`, `git diff` and `git fetch`, along with `git stash list` and `git merge-base`, pass even while it is enabled.

**Exemptions and failure direction.** The git agent that has to create branches is exempted by identity, and the `MOAI_BRANCH_GUARD_EXEMPT=1` environment variable also bypasses it. When classification is uncertain (not a git repository, `git rev-parse` failing, and so on) the command is allowed through and only an audit-log entry is written — the guard refuses only on positive evidence.

Work that needs a different branch belongs in a worktree rather than behind a refusal. For the procedure see [moai worktree](/en/cli-reference/worktree/).

## workflow.yaml — audit

Pins which model and effort the cross-model audit backends (`codex_audit`, `glm_audit`, `audit_multi`) actually run on. Each backend takes one `{model, effort}` pair, and the distributed defaults are empty.

```yaml
workflow:
    audit:
        codex:
            model: ""   # e.g. gpt-5.6-sol — a codex-servable model id
            effort: ""  # e.g. high — low | medium | high | xhigh | max
        glm:
            model: ""   # e.g. glm-5.3
            effort: ""  # e.g. max — low | high | max (z.ai reasoning-state names)
```

| Key | Description |
|-----|-------------|
| `audit.codex.{model, effort}` | The pair the codex audit carries on session start and review turns. An empty model, or one codex cannot serve, discards the pin and falls back to the existing SSOT resolution |
| `audit.glm.{model, effort}` | The pair the GLM audit carries on the z.ai request. Effort accepts only the z.ai reasoning-state names `low`, `high`, `max`; any other value drops the reasoning directive while the model pin still applies |
| Both pairs empty | The resolvers behave exactly as before this key existed — a project that never writes a pin changes nothing |

The pin applies to the audit entry points only. Model resolution on the task-delegation paths (`codex_task`, `glm_task`) is unaffected, and the same fields are editable in the web console's Audit panel.

## workflow.yaml — todo

The switch that turns the backlog queue's **guidance surfaces** off. The session-start summary line, the statusline TODO segment, and the skill's inference routing of natural-language requests into the todo workflow — all three go quiet under this one key.

```yaml
workflow:
    todo:
        enabled: false   # explicit off — an absent key reads as on
```

| Key | Value | Meaning |
|-----|-------|---------|
| `todo.enabled` | (key absent, default) | On. The shipped template carries no todo block at all, which is the state most projects live in |
| `todo.enabled` | `false` | The session-start summary, the statusline TODO segment, and the skill's automatic routing turn off |

**Turning it off does not remove the command.** The `moai todo` CLI stays registered with every verb working, and an explicit `/moai todo` invocation still runs. That boundary is intended — the switch silences only the surfaces that show the queue to someone who did not ask for it, and leaves the path of someone actually using it untouched. A config file that cannot be read also resolves to on (fail-open).

Reach for this key when the per-session backlog summary reads as noise on a small, one-off project. How to operate the queue itself is on the [moai todo](/en/utility-commands/moai-todo/) page.

## crosssession.yaml — cross-session messaging

Decides how this session treats messages from your other Claude Code sessions. The `moai cc` · `moai glm` · `moai cg` launchers translate these values into a transient `--settings` file at launch, and the web console edits this file through the settings seam. A session launched without the launcher — a bare `claude` command — does not read this file.

```yaml
crosssession:
  inbound: ""             # "" | accept | hold | refuse
  isolate_machines: false # default — no approval required before a message leaves this machine
  dialog_expiry: ""       # "" | 60s | 5m | 10m | never
```

| Key | Value | Description |
|-----|-------|-------------|
| `inbound` | `""` (default) | Claude Code decides per message from the two sessions' permission-mode classes |
| `inbound` | `accept` | Delivers the message. A worker session meant to take messages unattended needs this value |
| `inbound` | `hold` | Parks each message for approval. A message held this way does not expire; it is delivered when an `accept` later applies |
| `inbound` | `refuse` | Drops the message |
| `isolate_machines` | `false` (default) | **No approval is required before a message goes to a session beyond this machine.** Messages between sessions on the same machine never leave the machine whatever the value is, but a message to a session beyond it travels through Anthropic servers — so the default permits that path without asking. Decide whether to keep it rather than inheriting it unread |
| `isolate_machines` | `true` | Requires your explicit approval before any message leaves the machine, even in `bypassPermissions` mode. A `true` from ANY settings scope applies, so a checked-in project file can turn the requirement on but not off — turning it off means removing the `true` from every scope that sets it |
| `dialog_expiry` | `""` (default) | Keeps Claude Code's five-minute default |
| `dialog_expiry` | `60s` · `5m` · `10m` · `never` | Deadline for the approval dialog on a **default**-held message; `never` holds until the session ends. It does not apply to a message held by an explicit `inbound: hold` |

## Related Documentation

- [settings.json Guide](/en/advanced/settings-json) — Claude Code runtime settings
- [Harness Profiles & Evaluation](/en/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/en/cli-reference/doctor) — `moai doctor config` for merged configuration inspection
