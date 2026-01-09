---
description: One-click automation - From SPEC generation to documentation sync
user-invocable: true
---

# /all-is-well Command

Execute complete Plan -> Run -> Sync workflow in a single command.

## Command Syntax

```
/moai:all-is-well "<features>" [options]
```

## Arguments

- `<features>`: Space-separated list of feature descriptions (required)

## Options

| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--worktree` | | false | Use git worktrees for parallel development |
| `--parallel` | `-p` | 1 | Number of parallel workers |
| `--no-branch` | | false | Skip creating feature branches |
| `--no-pr` | | false | Skip creating pull requests |
| `--auto-merge` | | false | Auto-merge PRs after approval |
| `--model` | `-m` | glm | Default model (glm, opus) |

## Execution Flow

### PHASE 0: Configuration
- Load git-strategy.yaml settings
- Parse command flags and validate options
- Verify dependencies and prerequisites

### PHASE 1: Planning (Opus Model)
- Analyze feature requirements
- Create SPEC documents for each feature
- **CHECKPOINT**: User reviews and approves SPECs
- Create worktrees if --worktree enabled

### PHASE 2: Implementation (GLM Model)
- Execute /moai:2-run for each SPEC
- Run TDD implementation cycles
- Monitor progress via WebSocket events
- Wait for all implementations to complete

### PHASE 3: Sync and PR
- Execute moai-adk worktree sync --all
- Run /moai:3-sync for each SPEC
- Create pull requests if enabled
- Auto-merge if --auto-merge specified

### PHASE 4: Completion
- Generate workflow summary report
- Display cost analysis
- Cleanup worktrees if used

## Examples

### Single Feature
```
/moai:all-is-well "user authentication system"
```

### Multiple Features
```
/moai:all-is-well "user auth" "dashboard" "api endpoints"
```

### With Parallel Execution
```
/moai:all-is-well "feature1" "feature2" "feature3" --parallel 3 --worktree
```

### Full Options
```
/moai:all-is-well "complex feature" --worktree --parallel 2 --auto-merge --model opus
```

## Checkpoint Approval

During PHASE 1, the workflow pauses for SPEC review:

1. SPECs are displayed for review
2. User can approve to continue
3. User can reject to modify requirements
4. Workflow proceeds only after approval

## Cost Tracking

The workflow tracks and reports:
- Total tokens used per model
- Cost breakdown by phase
- Per-SPEC resource usage
- Cumulative session cost

## Error Handling

- Recoverable errors trigger retry with exponential backoff
- Non-recoverable errors halt workflow and report status
- Partial completions are preserved for manual recovery

## Integration

### With Worktree Manager
```bash
moai-adk worktree list    # View active worktrees
moai-adk worktree sync    # Sync changes to main
moai-adk worktree clean   # Remove completed worktrees
```

### With SPEC System
- Creates SPEC-XXX documents automatically
- Tracks implementation progress
- Links to PR when created

## Requirements

- Git repository initialized
- MoAI-ADK configured (.moai/config/)
- Valid API credentials for AI providers

## Output

On completion, displays:
- Workflow summary with all SPECs
- Implementation status per feature
- Total cost and token usage
- Links to created PRs (if any)

## Related Commands

- `/moai:1-plan` - SPEC planning only
- `/moai:2-run` - TDD implementation only
- `/moai:3-sync` - Documentation sync only
- `/moai:moai-loop` - Continuous feedback loop
