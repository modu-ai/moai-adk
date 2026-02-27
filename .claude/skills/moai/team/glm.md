---
name: moai-workflow-team
description: >
  CG (Claude + GLM) hybrid mode for MoAI-ADK. Uses tmux pane-level environment
  isolation so the Leader session runs on Claude API while Agent Teams teammates
  spawn in new tmux panes that inherit GLM env vars from the tmux session.
  60-70% cost reduction for implementation-heavy tasks.
user-invocable: false
metadata:
  version: "3.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-22"
  tags: "team, glm, tmux, cost-effective, agent-teams, cg"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["--team", "team mode", "glm worker", "cg mode", "tmux"]
  agents: ["moai"]
  phases: ["run"]
---

# MoAI CG Mode (Claude + GLM via tmux Agent Teams)

## Overview

CG mode combines Claude (leader) with GLM (teammates) using tmux pane-level
environment isolation. The leader session stays on Claude API for high-quality
planning and review, while Agent Teams teammates spawn in new tmux panes that
inherit GLM env vars — routing them through Z.AI's cost-effective GLM API.

```
User runs: moai glm sk-xxx     (save API key, once)
User runs: tmux new -s moai    (start tmux session)
User runs: moai cg             (configure CG mode)
    │
    ├── Validates tmux session (required)
    ├── Removes GLM env from settings.local.json
    │   → Leader (this pane) uses Claude API
    ├── Injects GLM env into tmux session env
    │   → New panes inherit GLM env → Z.AI API
    ├── Sets CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
    ├── Sets CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
    └── Saves team_mode: cg to llm.yaml

User runs: claude              (leader on Claude, in THIS pane)
User runs: /moai --team "task"
    │
    ├── PHASE 1: PLAN (Leader on Claude)
    │   └── manager-spec creates SPEC
    │
    ├── PHASE 2: RUN (Agent Teams — teammates on GLM)
    │   ├── TeamCreate → teammates spawn in new tmux panes
    │   ├── New panes inherit GLM env from tmux session
    │   ├── Teammates run on Z.AI GLM API
    │   ├── File ownership prevents write conflicts
    │   └── Teammates complete tasks and report
    │
    ├── PHASE 3: QUALITY (Leader on Claude)
    │   └── manager-quality validates TRUST 5
    │
    └── PHASE 4: SYNC (Leader on Claude)
        └── Documentation and PR
```

## How It Works

The key mechanism is **tmux session-level environment variables**:

1. `tmux set-environment` injects GLM vars at the session level
2. The CURRENT pane is NOT affected (leader stays on Claude)
3. Only NEW panes inherit session-level env vars
4. Agent Teams with `CLAUDE_CODE_TEAMMATE_DISPLAY=tmux` spawns teammates in new panes
5. Result: Leader = Claude API, Teammates = Z.AI GLM API

This is NOT headless mode. Teammates run as full interactive Claude Code
sessions in their own tmux panes, visible via `tmux list-panes`.

## Cost Benefit

| Phase | Tokens | Model | Cost |
|-------|--------|-------|------|
| Plan | 30K | Leader (Claude) | Standard |
| Run | 180K | Teammates (GLM) | Cost effective |
| Quality | 20K | Leader (Claude) | Standard |
| Sync | 40K | Leader (Claude) | Standard |

**Result**: Run phase uses ~70% of tokens. GLM for Run = 60-70% overall cost reduction.

## LLM Mode Detection

Read `.moai/config/sections/llm.yaml` for `team_mode` value:

| team_mode | Execution Mode | Leader | Teammates |
|-----------|---------------|--------|-----------|
| (empty) | Sub-agent | Current session | Task() subagents |
| cg | CG Mode | Claude (this pane) | GLM (new tmux panes) |
| glm | GLM-only | GLM | GLM |

Detection steps:
1. Read `.moai/config/sections/llm.yaml`
2. If `team_mode == "cg"`: Activate CG mode (this skill)
3. If `team_mode == "glm"`: All-GLM mode (no hybrid)
4. If `team_mode == ""`: Fall back to sub-agent mode

## Prerequisites

1. **Save GLM API key** (once):
   ```bash
   moai glm sk-your-glm-api-key
   ```
   Or set `GLM_API_KEY` environment variable.

2. **Start tmux session** (required for CG mode):
   ```bash
   tmux new -s moai
   ```

3. **Enable CG mode** (inside tmux):
   ```bash
   moai cg
   ```

4. **Start Claude Code in the SAME pane** (critical):
   ```bash
   claude
   ```
   Starting Claude in a NEW pane would make it inherit GLM env.

5. **Run workflow**:
   ```
   /moai --team "Your task description"
   ```

## tmux Environment Variables

`moai cg` injects these into the tmux session:

| Variable | Value | Purpose |
|----------|-------|---------|
| ANTHROPIC_AUTH_TOKEN | GLM API key | Z.AI authentication |
| ANTHROPIC_BASE_URL | https://api.z.ai/api/anthropic | Z.AI endpoint |
| ANTHROPIC_DEFAULT_OPUS_MODEL | glm-5 | Opus model override |
| ANTHROPIC_DEFAULT_SONNET_MODEL | glm-4.7 | Sonnet model override |
| ANTHROPIC_DEFAULT_HAIKU_MODEL | glm-4.5-air | Haiku model override |

These are set via `tmux set-environment` (session-level, not global).

## Agent Teams Integration

CG mode uses standard Agent Teams with tmux display:

- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` enables Agent Teams
- `CLAUDE_CODE_TEAMMATE_DISPLAY=tmux` makes teammates spawn in tmux panes
- Each teammate gets its own pane → inherits GLM env → uses Z.AI API
- Leader stays in the original pane → no GLM env → uses Claude API

Agent model mapping in CG mode:

| Agent | Pane | API | Model |
|-------|------|-----|-------|
| Leader (MoAI) | Original | Claude | User's choice (Opus/Sonnet) |
| team-backend-dev | New pane | Z.AI | glm-5 / glm-4.7 |
| team-frontend-dev | New pane | Z.AI | glm-5 / glm-4.7 |
| team-tester | New pane | Z.AI | glm-5 / glm-4.7 |
| team-researcher | New pane | Z.AI | glm-4.5-air |
| team-quality | New pane | Z.AI | glm-4.5-air |

## Error Recovery

| Failure | Recovery |
|---------|----------|
| Not in tmux | Error: "CG mode requires a tmux session" |
| No API key | Error: "Run moai glm <api-key> first" |
| Teammate spawn failure | Falls back to sub-agent mode |
| tmux env injection failure | Fatal for CG mode (retry tmux session) |
| Quality gate failure | Leader creates fix task or manual intervention |

## Comparison with Other Modes

| Aspect | CG Mode | GLM-only Mode | Sub-agent Mode | Agent Teams Mode |
|--------|---------|---------------|----------------|------------------|
| APIs | Claude + GLM | GLM only | Claude only | Claude only |
| Cost | Lowest | Low | Medium | Highest |
| Quality | Highest (Claude leads) | Medium | High | High |
| Parallelism | Parallel (Agent Teams) | Parallel (Agent Teams) | Sequential | Parallel |
| Requires tmux | Yes | No (recommended) | No | No |
| Isolation | tmux panes | tmux panes / in-process | None | File ownership |

## Cleanup

When done with CG mode:

```bash
moai cc
```

This command:
- Removes GLM env from settings.local.json
- Unsets tmux session GLM env vars
- Resets team_mode to empty in llm.yaml
- Restores standard Claude-only operation

---

Version: 3.0.0 (tmux Agent Teams CG Mode)
Last Updated: 2026-02-22
