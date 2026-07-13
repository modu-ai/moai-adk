# Worktree Workflow Patterns

Git worktree provides isolated working directories per SPEC for parallel development without context switching.

## Worktree Concept

- Independent working directories for multiple branches
- Each SPEC gets isolated development environment
- No branch switching needed for parallel work
- Reduced merge conflicts through feature isolation

## Worktree Creation

```bash
# Create parallel SPECs with separate worktrees
/moai plan "login feature" "signup feature" --worktree
```

Result: creates `project-worktrees` directory with SPEC-specific subdirectories.

## Worktree Benefits

| Benefit | Detail |
|---------|--------|
| Parallel Development | Multiple features developed simultaneously |
| Team Collaboration | Clear ownership boundaries per SPEC |
| Dependency Isolation | Different library versions per feature |
| Risk Reduction | Unstable code does not affect other features |

## Integration Examples

### Sequential Workflow

```bash
# Step 1: PLAN
/moai plan "user authentication system"

# Step 2: RUN
/moai run SPEC-001

# Step 3: SYNC
/moai sync SPEC-001
```

### Parallel Workflow

```bash
# Create multiple SPECs with worktrees
/moai plan "backend API" "frontend UI" "database schema" --worktree

# Session 1 (backend API worktree)
/moai run SPEC-001

# Session 2 (frontend UI worktree, separate terminal)
/moai run SPEC-002

# Session 3 (database schema worktree, separate terminal)
/moai run SPEC-003
```

## Worktree Isolation Rules (Advisory — 2026-05-17 Policy)

Per user policy 2026-05-17, L2/L3 worktree usage is user opt-in. L1 `Agent(isolation: "worktree")` is Claude Code runtime autonomous — MoAI orchestrator does not mandate isolation.

See [moai-workflow-worktree](../../moai-workflow-worktree/SKILL.md) for the canonical worktree management skill.
