---
description: "Run Phase 0 — Context loading, mode dispatch, harness level routing, and worktree path rules"
user-invocable: false
metadata:
  parent: moai-workflow-run
  phase: "Phase 0: Context Loading and Mode Dispatch"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai run Phase <N> <enter|exit> -->

# Run Workflow Orchestration

## Purpose

Implement SPEC requirements using the configured development methodology.

For methodology details (DDD ANALYZE-PRESERVE-IMPROVE and TDD RED-GREEN-REFACTOR cycles, success criteria, brownfield enhancement), see: .claude/rules/moai/workflow/spec-workflow.md (Run Phase section)

## Scope

- Implements Step 3 of MoAI's 4-step workflow (Task Execution)
- Receives SPEC documents created by /moai plan
- Hands off to /moai sync for documentation and PR

## Input

- $ARGUMENTS: SPEC-ID to implement (e.g., SPEC-AUTH-001)
- Resume: Re-running /moai run SPEC-XXX resumes from last successful phase checkpoint
- --team: Enable team-based implementation (see ${CLAUDE_SKILL_DIR}/team/run.md for parallel implementation team)

## Mode Dispatch (Multi-Mode Router)

<!-- @MX:NOTE - Multi-Mode Router per the mode-dispatch policy; --mode {autopilot,loop,team,pipeline} dispatch with harness-based default. See spec-workflow.md#mode-dispatch-cross-reference. -->

`/moai run` participates in the `--mode` axis with 4 valid values: `autopilot`, `loop`, `team`, `pipeline`. Each value selects a distinct execution style.

### Mode Values

- **`autopilot` (default for harness `minimal` / `standard`)**: Single-lead orchestration via Phase 0.95 Scale-Based Mode Selection (Fix / Focused / Standard / Full Pipeline) → Phase 2A/2B per `quality.yaml development_mode`. Behaves as today's default `/moai run` invocation.
- **`loop`**: Delegate to `Skill("moai-workflow-loop")` with the SPEC-ID and remaining args. Bypasses Phase 2A/2B and enters the Ralph engine per-iteration cycle (see `loop.md` Steps 1-9). `/moai loop SPEC-XXX` is an alias resolving to `/moai run --mode loop SPEC-XXX` with identical behavior.
- **`team` (default for harness `thorough` AND prerequisites met)**: Routes to existing Team Mode Routing (Phase 0.95 row 5 + the Team Mode Routing section). Requires `workflow.team.enabled: true` AND `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` env var.
- **`pipeline`**: REJECTED on `/moai run`. Pipeline mode is reserved for utility subcommands (`fix`, `coverage`, `mx`, `codemaps`, `clean`). Passing `--mode pipeline` here triggers `MODE_PIPELINE_ONLY_UTILITY` (the same error key the utility subcommands share).

### Mode Resolver

Precedence order (hard-coded):

1. CLI flag `--mode <value>` — highest priority. Wins regardless of config / harness.
2. Config field `workflow.default_mode` in `.moai/config/sections/workflow.yaml` — used when CLI flag is absent.
3. Harness auto-selection — fallback when both CLI and config are absent.

Pseudocode:

<!-- @MX:WARN - Mode resolver validation sensitive to new mode value additions. Mode-set lock-in. @MX:REASON - Expanding the mode set without updating validation set breaks runtime error path. -->
```
mode = cli_mode_flag or config.workflow.default_mode or harness_auto_select(harness_level)
if mode not in {autopilot, loop, team, pipeline}:
    emit MODE_UNKNOWN listing 4 valid values; abort
if mode == pipeline:
    emit MODE_PIPELINE_ONLY_UTILITY pointing to {fix, coverage, mx, codemaps, clean}; abort
if mode == team and not (workflow.team.enabled and CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS):
    if cli_mode_flag == team:                      # explicit request
        emit MODE_TEAM_UNAVAILABLE suggesting --mode autopilot; abort
    else:                                          # harness auto-selected team
        log [mode-auto-downgrade] info; mode = autopilot  # silent downgrade per the relevant requirement
dispatch(mode)
```

### Harness-Based Default Selection

| Harness level | Team prerequisites | Default mode |
|---------------|--------------------|--------------|
| `minimal` | (any) | `autopilot` |
| `standard` | (any) | `autopilot` |
| `thorough` | satisfied | `team` |
| `thorough` | not satisfied | `autopilot` (downgraded with `[mode-auto-downgrade]` info log) |

### Sentinel Error Keys

This skill emits the following sentinel error keys when mode dispatch fails. A CI audit verifies each literal sentinel remains present in this skill body.

- **`MODE_UNKNOWN`**: Emitted when `--mode <value>` is supplied but `<value>` is not in the 4-value valid set `{autopilot, loop, team, pipeline}`. The error message MUST enumerate the 4 valid values to guide the user.
- **`MODE_TEAM_UNAVAILABLE`**: Emitted when an EXPLICIT `--mode team` request cannot be honored because either `workflow.team.enabled: false` in workflow.yaml OR the `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var is unset. The error message MUST suggest `--mode autopilot` as the supported fallback. Note: when `team` is auto-selected by harness (not explicit CLI), the system silently downgrades to `autopilot` with an info log instead of raising this error.
- **`MODE_PIPELINE_ONLY_UTILITY`**: Preserved from the utility-subcommand baseline. Emitted when `--mode pipeline` is passed to this multi-agent subcommand. The error message MUST point the user to the utility subcommand set `{fix, coverage, mx, codemaps, clean}`.

See [Subcommand Classification matrix](../../rules/moai/workflow/spec-workflow.md#subcommand-classification) for the cross-skill mode dispatch contract.

## UltraThink Auto-Activation

When the run phase begins, evaluate whether to activate deep analysis mode for the strategy phase:

**Activation condition** (any of):
- SPEC spans >= 2 distinct domains (backend + frontend, auth + database, etc.)
- SPEC plan.md lists >= 8 files to create or modify
- SPEC involves architectural patterns (new module, service, middleware layer)
- User explicitly includes `ultrathink` keyword

**UltraThink (primary deep reasoning trigger)**:
- `ultrathink`: Extended reasoning within the current agent (Adaptive Thinking on Opus 4.7+) — deeper strategy analysis, more thorough trade-off evaluation

When activated: Apply to Phase 1 (Strategy) for deeper architectural analysis. Log: "UltraThink mode activated for strategy phase: [reason]"

## Harness Level Routing

At Run phase entry, determine the pipeline depth:

1. Receive harness level from orchestrator (moai.md Complexity Estimator) or default to standard
2. Apply level-specific phase configuration:
   - **minimal**: Skip phases [0, 0.6, 2.0, 2.5, 2.75, 2.8a, 2.9, 2.10]. Direct implementation only.
     Note: Phase 0.5 (Plan Audit Gate) is NEVER skipped, not even in minimal harness.
   - **standard**: Execute all phases. sync-auditor in final-pass mode (Phase 2.8a only).
   - **thorough**: Execute all phases. sync-auditor in per-sprint mode (Phase 2.0 + 2.8a). Sprint contract enabled.
3. Load SPEC context (token-efficient):
   - If `.moai/specs/SPEC-{ID}/spec-compact.md` exists: Load spec-compact.md (~30% token savings)
   - Otherwise: Load full spec.md (backward compatible)
4. Log harness level to progress.md for traceability

Escalation: If a quality gate fails during execution, escalate harness level:
- minimal → standard (on Phase 2.5 fail)
- standard → thorough (on Phase 2.8a CRITICAL)
- Maximum 2 escalations per SPEC run

## Context Loading

Before execution, load these essential files:

- .moai/config/config.yaml (git strategy, automation settings)
- .moai/config/sections/quality.yaml (coverage targets, TRUST 5 settings, development_mode)
- .moai/config/sections/harness.yaml (harness depth levels, auto-detection rules)
- .moai/config/sections/git-strategy.yaml (auto_branch, branch creation policy)
- .moai/config/sections/language.yaml (git_commit_messages setting)
- .moai/specs/SPEC-{ID}/ directory (spec-compact.md preferred, or spec.md, plan.md, acceptance.md)
- .moai/specs/SPEC-{ID}/progress.md (session resume context: if exists, load to identify completed phases and skip them; if absent, will be created at Phase 1 start)
- .moai/specs/SPEC-{ID}/tasks.md (task decomposition with planned files, if exists)
- .moai/project/structure.md (architecture context for implementation decisions)
- .moai/project/tech.md (technology stack context)
- .moai/project/codemaps/ directory listing (architecture maps for dependency and module understanding)

Pre-execution commands: git status, git branch, git log, git diff.

### Lessons Loading

Before spawning implementation agents, load relevant lessons from auto-memory:

1. Read `~/.claude/projects/{project-hash}/memory/lessons.md` if it exists
2. Filter lessons by domain relevance:
   - Match lesson categories against SPEC domain keywords
   - Match lesson tags against modified file paths (from SPEC scope)
   - Limit to top 5 most recent matching lessons
3. Include filtered lessons in agent spawn prompt as "Previous lessons learned" context
4. Maximum 2000 tokens for lesson injection
5. If lessons.md does not exist or no relevant lessons found, skip silently

### Resume Check

Before Phase 1, check if `.moai/specs/SPEC-{ID}/progress.md` exists:
- If it exists: Load content, identify last completed phase checkpoint, skip all completed phases, resume from the next pending phase. Log: "Resuming SPEC-{ID} from Phase {N}"
- If it does not exist: Create the file now with initial entry:
  ```
  ## SPEC-{ID} Progress

  - Started: {current timestamp}
  ```
- The progress.md file persists across sessions and enables seamless resume after interruption.

---

## Worktree Path Rules [HARD] (All Modes)

When delegating to ANY agent with `isolation: "worktree"` (sub-agent mode or team mode):

- [HARD] Reference all write-target files by project-root-relative paths (e.g., `src/auth/handler.go`)
- [HARD] Do NOT include absolute paths (e.g., `$HOME/project/src/auth/handler.go`) in agent prompts
- [HARD] Do NOT include `cd /absolute/path &&` in any Bash commands within agent prompts
- [HARD] SPEC files: use `.moai/specs/SPEC-XXX/spec.md` (relative), not absolute paths
- [HARD] The agent's CWD is automatically set to the worktree root by Claude Code — all relative paths resolve correctly

Anti-patterns that bypass worktree isolation:
```
# WRONG: Absolute path bypasses worktree
"Read $HOME/project/src/auth/handler.go and fix the bug"

# WRONG: cd to main project in Bash command
"Run: cd $HOME/project && go test ./..."

# CORRECT: Relative path — agent resolves from its own CWD (worktree root)
"The bug is in src/auth/handler.go. Read the file and fix it."

# CORRECT: No cd prefix — agent CWD is already worktree root
"Run: go test ./..."
```

See `.claude/rules/moai/workflow/worktree-integration.md` for complete path rules.
