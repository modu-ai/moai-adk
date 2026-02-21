---
name: moai-workflow-team
description: >
  GLM Worker mode for MoAI-ADK. Enables Leader (current Claude model) + Worker (GLM-5)
  architecture with git worktree isolation. Leader creates SPEC on your selected model,
  Worker implements on GLM-5, Leader merges and documents. 60-70% cost
  reduction for implementation-heavy tasks.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "team, glm, worktree, cost-effective, parallel"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["--team", "team mode", "glm worker", "worktree"]
  agents: ["moai"]
  phases: ["run"]
---

# MoAI GLM Worker Mode

## Overview

GLM Worker mode combines your selected Claude model with GLM-5 (cost effective) for optimal development workflow:

```
User runs: moai cg
    │
    ├── team_mode="cg" saved to llm.yaml
    │
User runs: claude (starts on your selected model)
    │
User runs: /moai --team "Add user auth"
    │
    ├── PHASE 1: PLAN (Leader)
    │   └── manager-spec creates SPEC
    │
    ├── PHASE 2: RUN (GLM-5 Worker)
    │   ├── Leader spawns Worker via Bash
    │   │   env ANTHROPIC_*=... claude -w worker-SPEC-XXX -p "..."
    │   ├── Worker implements in worktree
    │   └── Worker commits changes
    │
    ├── PHASE 3: MERGE (Leader)
    │   ├── git merge worktree branch
    │   ├── Run tests, lint
    │   └── Quality validation
    │
    └── PHASE 4: SYNC (Leader)
        ├── Documentation
        └── Worktree cleanup
```

## Cost Benefit

| Phase | Tokens | Model | Cost |
|-------|--------|-------|------|
| Plan | 30K | Leader (your model) | Standard |
| Run | 180K | GLM-5 | Cost effective |
| Merge | 20K | Leader (your model) | Standard |
| Sync | 40K | Leader (your model) | Standard |

**Result**: Run phase uses ~70% of tokens. GLM-5 for Run = 60-70% overall cost reduction.

## LLM Mode Detection

Read `.moai/config/sections/llm.yaml` for `team_mode` value:

| team_mode | Execution Mode | Leader | Worker |
|-----------|---------------|--------|--------|
| (empty) | Sub-agent | Current session | Task() subagents |
| glm | GLM Worker | Current Claude model | GLM-5 in worktree |

Detection steps:
1. Read `.moai/config/sections/llm.yaml`
2. If `team_mode == "glm"`: Activate GLM Worker mode (this skill)
3. If `team_mode == ""`: Fall back to sub-agent mode

## Prerequisites

Before using `/moai --team`:

1. **Save GLM API key** (CLI):
   ```bash
   moai glm sk-your-glm-api-key
   ```
   Or set `GLM_API_KEY` environment variable.

2. **Enable CG mode** (CLI):
   ```bash
   moai cg
   ```
   This saves `team_mode: cg` to llm.yaml and configures worktree-based isolation.

3. **Start Claude Code**:
   ```bash
   claude
   ```
   This starts on your selected Claude model because settings.local.json was NOT modified.

4. **Run workflow**:
   ```
   /moai --team "Your task description"
   ```

## Worker Spawning

The Leader spawns a GLM-5 worker via Bash:

```bash
env \
  ANTHROPIC_AUTH_TOKEN="$GLM_API_KEY" \
  ANTHROPIC_BASE_URL="https://api.z.ai/api/anthropic" \
  ANTHROPIC_DEFAULT_OPUS_MODEL="glm-5" \
  ANTHROPIC_DEFAULT_SONNET_MODEL="glm-4.7" \
  ANTHROPIC_DEFAULT_HAIKU_MODEL="glm-4.7-flashx" \
  API_TIMEOUT_MS="3000000" \
  claude -w worker-SPEC-XXX \
    -p "Implement SPEC at .moai/specs/SPEC-XXX/spec.md
        Follow TDD methodology.
        Commit when done." \
    --output-format text \
    > .moai/team/worker-SPEC-XXX.log 2>&1 &
```

**Key concepts**:
- `env VAR=val command` sets environment ONLY for that command
- `claude -w NAME` creates worktree at `.claude/worktrees/NAME/`
- `-p "..."` runs headless with prompt
- `&` backgrounds the process

## Worktree Isolation

Each worker operates in its own git worktree:

```
project-root/
├── .claude/
│   └── worktrees/
│       ├── worker-SPEC-AUTH-001/    # Worker 1 worktree
│       │   └── ... (implementation files)
│       └── worker-SPEC-API-002/     # Worker 2 worktree
│           └── ... (implementation files)
└── .moai/
    └── team/
        ├── worker-SPEC-AUTH-001.log  # Worker 1 log
        └── worker-SPEC-API-002.log   # Worker 2 log
```

**Branch naming**: `worktree-worker-SPEC-XXX`

**Cleanup**: Automatic on `moai cc` or manual via `git worktree remove`

## Merge Process

After worker completes:

1. **Fetch and merge**:
   ```bash
   git fetch origin worktree-worker-SPEC-XXX
   git merge worktree-worker-SPEC-XXX --no-ff
   ```

2. **Quality gates**:
   ```bash
   go test -race ./...
   golangci-lint run
   ```

3. **Conflict resolution**: Leader (Opus) handles or presents to user

## Multiple Workers

For large SPECs with independent components:

```
/moai --team "Add user auth with API and UI"
    │
    ├── SPEC split: AUTH-API, AUTH-UI, AUTH-TESTS
    │
    ├── Spawn workers in parallel:
    │   ├── worker-SPEC-AUTH-API    (GLM-5)
    │   ├── worker-SPEC-AUTH-UI     (GLM-5)
    │   └── worker-SPEC-AUTH-TESTS  (GLM-5)
    │
    └── Sequential merge by dependency
```

## Error Recovery

| Failure | Recovery |
|---------|----------|
| Worker timeout | Extend, check log, or abort |
| Worker failure | Check log, retry with refined prompt, or sub-agent fallback |
| Merge conflict | Opus auto-resolves or user choice |
| Quality gate failure | Create fix task or manual intervention |

## Comparison with Other Modes

| Aspect | GLM Worker | Sub-agent | Agent Teams |
|--------|-----------|-----------|-------------|
| APIs | Claude + GLM | Single | Single |
| Cost | Lowest | Medium | Highest |
| Quality | High (Leader review) | High | High |
| Parallelism | Sequential | Sequential | Parallel |
| Isolation | Worktree | None | File ownership |

## Cleanup

When done with team mode:

```bash
moai cc
```

This command:
- Removes GLM env from settings.local.json
- Resets team_mode to empty
- Cleans up moai worktrees
- Deletes worker logs

---

Version: 2.0.0 (GLM Worker Mode)
Last Updated: 2026-02-20
