---
description: "Start Ralph-style feedback loop for automated error correction"
argument-hint: "[--max-iterations N] [--auto-fix]"
type: utility
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git diff --name-only HEAD

## Essential Files

@.moai/config/sections/ralph.yaml

---

# /moai:loop - Ralph Engine Feedback Loop

Start an automated feedback loop that continuously checks for errors and guides fixes until all conditions are satisfied.

## Command Purpose

Implements the Ralph-style "continuous improvement" pattern:

1. Check current state (LSP errors, test failures, coverage)
2. If issues exist, provide guidance to fix them
3. After each fix, re-check conditions
4. Repeat until all conditions are met or max iterations reached

Arguments: $ARGUMENTS

## Usage Examples

Start basic feedback loop:

```
/moai:loop
```

With iteration limit:

```
/moai:loop --max-iterations 5
```

With auto-fix enabled (applies safe fixes automatically):

```
/moai:loop --auto-fix
```

## Command Options

- `--max-iterations N`: Override max iterations (default: 10 from ralph.yaml)
- `--auto-fix`: Enable automatic application of safe fixes
- `--errors-only`: Only check for errors, ignore warnings
- `--include-coverage`: Include test coverage in completion conditions

## Loop Behavior

### Completion Conditions (from ralph.yaml)

The loop completes when ALL enabled conditions are met:

1. **zero_errors**: No LSP/compiler errors
2. **zero_warnings**: No LSP warnings (optional, default: false)
3. **tests_pass**: All tests pass
4. **coverage_threshold**: Test coverage meets minimum (default: 85%)

### Loop Iteration Cycle

```
ITERATION START
  |
  v
CHECK CONDITIONS
  |-- LSP Diagnostics (errors/warnings)
  |-- Test Execution (pass/fail)
  |-- Coverage Report (percentage)
  |
  v
ALL MET? --YES--> COMPLETE
  |
  NO
  |
  v
GENERATE GUIDANCE
  |-- List specific issues
  |-- Suggest fixes
  |-- Prioritize by severity
  |
  v
APPLY FIXES (manual or auto)
  |
  v
INCREMENT ITERATION
  |
  v
MAX REACHED? --YES--> STOP (with summary)
  |
  NO
  |
  v
ITERATION START (repeat)
```

## Integration with Hooks

When the loop is active, these hooks provide real-time feedback:

### PostToolUse Hook (post_tool\_\_lsp_diagnostic.py)

- Runs after every Write/Edit operation
- Provides immediate LSP diagnostic feedback
- Exit code 2 signals errors needing attention

### Stop Hook (stop\_\_loop_controller.py)

- Runs after each Claude response
- Checks completion conditions
- Exit code 1 continues the loop
- Exit code 0 completes the loop

## Loop State Management

State is persisted in `.moai/cache/.moai_loop_state.json`:

```json
{
  "active": true,
  "iteration": 3,
  "max_iterations": 10,
  "last_error_count": 2,
  "last_warning_count": 5,
  "files_modified": ["src/auth.py", "tests/test_auth.py"],
  "start_time": 1704067200.0,
  "completion_reason": null
}
```

## Auto-Fix Capabilities

When `--auto-fix` is enabled, safe fixes are applied automatically:

### Safe Auto-Fixes (applied without confirmation)

- Import sorting and organization
- Whitespace and formatting issues
- Unused import removal
- Simple type annotation additions

### Unsafe Fixes (require confirmation)

- Logic changes
- API modifications
- Test modifications
- Security-related changes

## Output Format

### During Loop

```markdown
## Ralph Loop: Iteration 3/10

### Current Status

- Errors: 2
- Warnings: 5
- Tests: 23/25 passing
- Coverage: 82%

### Issues to Address

1. [ERROR] src/auth.py:45 - undefined name 'jwt_token'
2. [ERROR] src/auth.py:67 - missing return statement
3. [WARNING] tests/test_auth.py:12 - unused variable 'result'

### Suggested Actions

1. Import jwt_token or define it locally
2. Add return statement to validate_token function
3. Use or remove the 'result' variable
```

### On Completion

```markdown
## Ralph Loop: COMPLETE

### Final Status

- Iterations: 4
- Errors: 0
- Warnings: 0
- Tests: 25/25 passing
- Coverage: 87%

### Changes Made

- Fixed 3 errors
- Resolved 5 warnings
- Added 2 test cases

### Duration

- Start: 10:30:00
- End: 10:35:42
- Total: 5m 42s
```

## Cancellation

To cancel an active loop:

- Use `/moai:cancel-loop` command
- Or set environment variable: `MOAI_LOOP_ACTIVE=false`
- Or delete state file: `.moai/cache/.moai_loop_state.json`

## Best Practices

1. **Start Small**: Begin with `--max-iterations 3` for new projects
2. **Review Auto-Fixes**: Even with `--auto-fix`, review changes before committing
3. **Use with TDD**: Combine with `/moai:2-run` for test-driven development
4. **Monitor Progress**: Watch iteration count to detect stuck loops

## Error Recovery

If the loop gets stuck:

1. Check LSP server status
2. Verify test framework is working
3. Review recent changes for syntax errors
4. Use `/moai:cancel-loop` and fix manually if needed

---

## Execution Directive

When this command is invoked:

1. Parse arguments for options
2. Initialize or resume loop state
3. Set environment variables:
   - `MOAI_LOOP_ACTIVE=true`
   - `MOAI_LOOP_ITERATION=N`
4. Check initial conditions
5. If issues exist, provide first iteration guidance
6. Let hooks handle subsequent iterations

The Stop hook (`stop__loop_controller.py`) manages the loop continuation.

---

Version: 1.0.0
Last Updated: 2026-01-10
Pattern: Continuous Feedback Loop
Integration: LSP Diagnostics, AST-grep, Test Runner
