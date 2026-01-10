---
description: "Cancel an active Ralph feedback loop"
argument-hint: "[--force]"
type: utility
allowed-tools: Bash, Read, Write, AskUserQuestion
model: inherit
---

## Pre-execution Context

!git status --porcelain
!cat .moai/cache/.moai_loop_state.json 2>/dev/null || echo "No active loop"

## Essential Files

@.moai/config/sections/ralph.yaml

---

# /moai:cancel-loop - Cancel Active Feedback Loop

Stop an active Ralph Engine feedback loop and optionally preserve or discard state.

## Command Purpose

Gracefully terminates an active `/moai:loop` session:

1. Stops the loop iteration cycle
2. Clears loop state file
3. Reports final status
4. Optionally preserves work in progress

Arguments: $ARGUMENTS

## Usage Examples

Cancel loop with confirmation:

```
/moai:cancel-loop
```

Force cancel without confirmation:

```
/moai:cancel-loop --force
```

## Command Options

- `--force`: Skip confirmation prompt and cancel immediately
- `--preserve-state`: Keep the state file for debugging

## Cancellation Process

```
START: /moai:cancel-loop

CHECK: Is loop active?
  |
  NO --> Report "No active loop" and EXIT
  |
  YES
  |
  v
CONFIRM (unless --force)
  |-- Show current loop status
  |-- Ask for confirmation
  |-- IF rejected: EXIT without canceling
  |
  v
CANCEL
  |-- Clear MOAI_LOOP_ACTIVE environment
  |-- Remove state file (unless --preserve-state)
  |-- Log cancellation
  |
  v
REPORT
  |-- Show final iteration count
  |-- Show issues remaining
  |-- Suggest next steps

END: Loop cancelled
```

## State File Location

The loop state is stored at:

```
.moai/cache/.moai_loop_state.json
```

Contents include:

- `active`: Whether loop is active
- `iteration`: Current iteration number
- `max_iterations`: Maximum allowed iterations
- `last_error_count`: Errors from last check
- `last_warning_count`: Warnings from last check
- `files_modified`: List of modified files
- `start_time`: When loop started

## Output Format

### Normal Cancellation

```markdown
## Loop Cancelled

### Final Status

- Iterations completed: 4
- Errors remaining: 2
- Warnings remaining: 3

### Work in Progress

Files modified during this loop:

- src/auth.py
- tests/test_auth.py

### Next Steps

1. Review modified files for partial fixes
2. Run `/moai:fix` to address remaining issues
3. Or start fresh with `/moai:loop`
```

### No Active Loop

```markdown
## No Active Loop

There is no active feedback loop to cancel.

To start a new loop:
```

/moai:loop

```

```

### Force Cancellation

```markdown
## Loop Force-Cancelled

Loop terminated without confirmation.

State file removed: .moai/cache/.moai_loop_state.json

Note: Any unsaved progress may be lost.
```

## When to Cancel

Cancel the loop when:

- Loop appears stuck (same errors repeating)
- Need to make manual changes
- Want to change loop configuration
- Encountered unexpected errors
- Need to switch to different task

## Recovery After Cancellation

If you cancelled mid-fix:

1. **Check Git Status**

   ```
   git status
   git diff
   ```

2. **Review Changes**
   Look at modified files for partial fixes

3. **Decide Next Action**
   - Commit working changes
   - Revert problematic changes
   - Resume with `/moai:loop`

## Environment Variables

The loop uses these environment variables:

- `MOAI_LOOP_ACTIVE`: Set to "true" when loop is active
- `MOAI_LOOP_ITERATION`: Current iteration number

These are cleared on cancellation.

## Related Commands

- `/moai:loop`: Start a new feedback loop
- `/moai:fix`: One-time fix without loop
- `/moai:alfred`: Full workflow automation

---

## Implementation

When this command runs:

1. Check for state file existence
2. If exists and active, proceed with cancellation
3. Clear environment variables
4. Remove or archive state file
5. Report status to user

```python
# Pseudo-implementation
state_path = Path(".moai/cache/.moai_loop_state.json")

if state_path.exists():
    state = json.load(state_path)
    if state.get("active"):
        # Show status and confirm
        # Then clear state
        state_path.unlink()
        print("Loop cancelled")
    else:
        print("Loop already inactive")
else:
    print("No active loop found")
```

---

Version: 1.0.0
Last Updated: 2026-01-10
Pattern: State Management
Integration: Ralph Engine Loop Controller
