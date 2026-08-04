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
      high: "glm-5.2"          # 1M context — Opus slot
      medium: "glm-4.7"        # 202K context — Sonnet slot
      low: "glm-4.5-air"       # 128K context — lightweight slot
      fable: "glm-5.2"
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

## ralph.yaml — Ralph Engine

Controls the diagnostic-driven iterative fix loop (Ralph Engine) used by `/moai loop`.

```yaml
ralph:
  max_iterations: 5         # iteration ceiling (default 5; CLI --max takes precedence)
  auto_converge: true       # auto-converge on stall detection
  human_review: true        # halt for human inspection at the review step
  lint_as_instruction: true # inject LSP diagnostics as next-turn guidance
  warn_as_instruction: false # also inject warnings when there are no errors
```

| Key | Description |
|-----|-------------|
| `max_iterations` | Iteration ceiling (default 5). Priority: CLI `--max` flag > `ralph.max_iterations` > `workflow.yaml` `loop_prevention.max_iterations` |
| `auto_converge` | Auto-converge after N consecutive no-progress turns |
| `human_review` | Halt at the review step for human inspection |
| `lint_as_instruction` | Inject LSP diagnostics as a `systemMessage` so the AI receives them as the next prompt (default true) |
| `warn_as_instruction` | Also inject warnings when there are no errors (default false) |

Related: [/moai loop](/en/utility-commands/moai-loop), [Autonomous Continuation Loops](/en/advanced/autonomous-loops).

## harness.yaml — harness depth·evaluation

Defines the harness quality pipeline depth (minimal/standard/thorough), auto-detection, evaluator memory scope, and escalation.

```yaml
harness:
  default_profile: "default"  # default | strict | lenient | frontend
  evaluator:
    memory_scope: per_iteration  # FROZEN — cannot be changed (design-constitution §11.4.1)
  mode_defaults:
    solo: auto
    team: auto
    cg: thorough                 # CG mode is always thorough
  auto_detection:
    enabled: true
    rules:
      minimal:  # file_count <= 3 AND single_domain, ...
      standard: # file_count > 3 OR multi_domain, ...
      thorough: # security/payment keywords, critical priority, ...
  escalation:
    enabled: true
```

| Block | Description |
|-------|-------------|
| `default_profile` | Default evaluation profile when the SPEC has no `evaluator_profile` |
| `evaluator.memory_scope` | Evaluator memory scope. Frozen at `per_iteration` |
| `mode_defaults` | Default depth per execution mode (solo/team/cg) |
| `auto_detection.rules` | Conditions under which the Complexity Estimator auto-classifies into minimal/standard/thorough |
| `escalation` | Escalate to a higher depth on failure |

Related: [Harness Profiles & Evaluation](/en/advanced/harness-profiles), [moai harness](/en/cli-reference/harness).

## quality.yaml — quality gates·development methodology

Controls development mode (DDD/TDD), coverage targets, LSP quality-gate thresholds, and quality-gate enforcement.

```yaml
constitution:
  development_mode: tdd        # tdd | ddd | off
  enforce_quality: true
  test_coverage_target: 85
  lsp_quality_gates:
    enabled: true
    plan: { require_baseline: true }
    run:  { max_errors: 0, max_type_errors: 0, max_lint_errors: 0, allow_regression: false }
    sync: { max_errors: 0, max_warnings: 10, require_clean_lsp: true }
```

| Block | Description |
|-------|-------------|
| `development_mode` | `tdd` / `ddd` / `off`. Determines the default `cycle_type` of `/moai run` |
| `enforce_quality` | When `true`, run-phase does not go GREEN on a quality-gate violation |
| `test_coverage_target` | Minimum package coverage target (%). Critical packages (cli/template/hook) should be 90%+ |
| `lsp_quality_gates.plan` | plan-phase: whether to capture an LSP baseline |
| `lsp_quality_gates.run` | run-phase: 0 errors/type-errors/lint-errors, no regression |
| `lsp_quality_gates.sync` | sync-phase: 0 errors, warnings ≤ 10, requires clean LSP |

Related: [SPEC Workflow](/en/advanced/spec-workflow), [LSP Gates](/en/advanced/lsp-gates).

## workflow.yaml — workflow state

Controls workflow thresholds, branch-state guard opt-in, session worktree opt-in, and agentic loop ceilings.

```yaml
workflow:
  agentic_loop:
    max_iterations: 10        # pipeline-level completion-loop ceiling
  branch_guard:
    enabled: false            # distributed default — hook is INERT (SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001)
  session_worktree:
    enabled: false            # distributed default — auto-isolation INERT
  loop_prevention:
    max_iterations: 5         # per-operation diagnostic fix-loop ceiling (separate axis from ralph.max_iterations)
```

| Block | Description |
|-------|-------------|
| `agentic_loop.max_iterations` | Pipeline-level completion-loop ceiling (`AgenticLoopConfig`) |
| `branch_guard.enabled` | Enables the Main-Checkout Branch-State Guard. **Distributed default `false`** — the hook does not evaluate in single-developer (non-shared-checkout) environments. Maintainers of shared multi-session checkouts opt in |
| `session_worktree.enabled` | Enables automatic worktree isolation on `moai init` / `moai profile` / `moai web`. **Distributed default `false`** (SPEC-SESSION-WORKTREE-001). Overridable via the `MOAI_SESSION_WORKTREE` env var |
| `loop_prevention.max_iterations` | Per-operation diagnostic fix-loop ceiling. A separate axis from `ralph.max_iterations` — ralph.yaml takes precedence |

> Both `branch_guard` and `session_worktree` ship with a distributed default of `false`. In a single-user repo these being `false` is intentional behavior, not a defect.

Related: [Main-Checkout Branch Guard](/en/advanced/branch-guard), [Worktree Integration](/en/advanced/worktree).

## mx.yaml — @MX tag system

Defines the tag kinds, per-language detection, and validation rules for the `@MX` code-annotation tag system.

```yaml
mx:
  version: "2.1"
  languages:
    go:      { enabled: auto, patterns: ["*.go"], exclude: ["*_generated.go", "vendor/**"] }
    python:  { enabled: auto, patterns: ["*.py"], exclude: ["**/__pycache__/**"] }
    # ... 16 languages treated equally (go, python, typescript, javascript, rust, java, kotlin,
    #     csharp, ruby, php, elixir, cpp, scala, r, flutter, swift)
  tags:
    # per-tag metadata and activation
  reason_required:  # tags that require @MX:REASON
    - WARN
    - DEBT
```

| Block | Description |
|-------|-------------|
| `version` | MX tag system schema version |
| `languages` | Lists 16 languages as equals. Each is auto-detected (`enabled: auto`) via project markers (`go.mod`, `pyproject.toml`, `Cargo.toml`, etc.) |
| `tags` | Per-tag metadata for `@MX:NOTE` / `@MX:WARN` / `@MX:ANCHOR` / `@MX:TODO` / `@MX:SPEC` / `@MX:DEBT` / `@MX:LEGACY`, etc. `@MX:SPEC` records SPEC associations (SPEC-MX-ASSOCIATION-001) |
| `reason_required` | Tags that require `@MX:REASON` (default: WARN, DEBT) |

> Do not demote a specific language to "PRIMARY" or mark only some as "enabled" — all 16 languages are equal.

Related: [@MX Tag Protocol](/en/advanced/mx-tag-protocol), [moai mx](/en/cli-reference/mx).

## Environment variable overrides

Some YAML section values can be overridden via environment variables. Environment variables take precedence over file values (`applyEnvOverrides` in `internal/config/manager.go`).

| Env var | Target | Description |
|---------|--------|-------------|
| `MOAI_DEVELOPMENT_MODE` | `constitution.development_mode` | Forces `tdd` / `ddd` / `off` |
| `MOAI_LOG_LEVEL` | Log level | `debug` / `info` / `warn` / `error` |
| `MOAI_LOG_FORMAT` | Log format | `text` / `json` |
| `MOAI_NO_COLOR` | Color output | `1` / `true` forces color off |
| `MOAI_CONFIG_DIR` | Config directory location | Use a different path instead of `.moai/config/` |

> These 5 are the complete set of environment-variable overrides the config manager actually reads. The constants are defined in `internal/config/envkeys.go`. `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` are not currently implemented — name and conversation language only read from `user.yaml` / `language.yaml`.

## Related Documentation

- [settings.json Guide](/en/advanced/settings-json) — Claude Code runtime settings
- [Harness Profiles & Evaluation](/en/advanced/harness-profiles) — harness.yaml / evaluator-profiles
- [moai doctor](/en/cli-reference/doctor) — `moai doctor config` for merged configuration inspection
