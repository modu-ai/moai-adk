---
name: moai-ralph
description: Ralph Engine - Automated feedback loop with LSP diagnostics and AST-grep integration for continuous code quality improvement. Use when implementing error-driven development, automated fixing, or continuous quality validation workflows.
version: 1.0.0
user-invocable: false
---

# Ralph Engine

Automated feedback loop system integrating LSP diagnostics, AST-grep security scanning, and test validation for continuous code quality improvement.

## Quick Reference (30 seconds)

Core Capabilities:
- LSP Integration: Real-time diagnostics from language servers
- AST-grep Scanning: Structural code analysis and security checks
- Feedback Loop: Iterative error correction until completion conditions met
- Hook System: PostToolUse and Stop hooks for seamless Claude Code integration

Key Components:
- `post_tool__lsp_diagnostic.py`: LSP diagnostics after Write/Edit
- `stop__loop_controller.py`: Loop iteration control
- `ralph.yaml`: Configuration settings

Commands:
- `/all-is-well`: One-click Plan-Run-Sync automation
- `/moai-loop`: Start feedback loop
- `/moai-fix`: One-time auto-fix
- `/cancel-loop`: Stop active loop

When to Use:
- Implementing features with zero-error goal
- Automated code quality improvement
- Continuous integration workflows
- Error-driven development patterns

## Implementation Guide (5 minutes)

### Architecture Overview

```
User Command
     |
     v
Command Layer (/moai-loop, /moai-fix, /all-is-well)
     |
     v
Hook System
     |-- PostToolUse Hook (LSP diagnostics)
     |-- Stop Hook (Loop controller)
     |
     v
Backend Services
     |-- LSP Client (MoAILSPClient)
     |-- AST-grep Scanner
     |-- Test Runner
     |
     v
Completion Check
     |-- Zero errors?
     |-- Tests pass?
     |-- Coverage met?
     |
     v
Continue or Complete
```

### Configuration (ralph.yaml)

```yaml
ralph:
  enabled: true

  lsp:
    auto_start: true
    timeout_seconds: 30
    graceful_degradation: true

  ast_grep:
    enabled: true
    security_scan: true
    quality_scan: true

  loop:
    max_iterations: 10
    auto_fix: false
    require_confirmation: true
    completion:
      zero_errors: true
      zero_warnings: false
      tests_pass: true
      coverage_threshold: 85

  hooks:
    post_tool_lsp:
      enabled: true
      severity_threshold: "error"
    stop_loop_controller:
      enabled: true
```

### Hook Integration

#### PostToolUse Hook

Triggered after Write/Edit operations:

```python
# Hook input (from Claude Code)
{
    "tool_name": "Write",
    "tool_input": {
        "file_path": "/path/to/file.py",
        "content": "..."
    }
}

# Hook output
{
    "hookSpecificOutput": {
        "hookEventName": "PostToolUse",
        "additionalContext": "LSP: 2 error(s), 3 warning(s) in file.py\n  - [ERROR] Line 45: undefined name 'x'"
    }
}
```

Exit codes:
- 0: No action needed
- 2: Attention needed (errors found)

#### Stop Hook (Loop Controller)

Triggered after each Claude response:

```python
# Loop state file (.moai/cache/.moai_loop_state.json)
{
    "active": true,
    "iteration": 3,
    "max_iterations": 10,
    "last_error_count": 2,
    "completion_reason": null
}

# Hook output
{
    "hookSpecificOutput": {
        "hookEventName": "Stop",
        "additionalContext": "Ralph Loop: CONTINUE | Iteration: 3/10 | Errors: 2\nNext actions: Fix 2 error(s)"
    }
}
```

Exit codes:
- 0: Loop complete or inactive
- 1: Continue loop (more work needed)

### LSP Client Usage

```python
from moai_adk.lsp.client import MoAILSPClient
from moai_adk.lsp.models import DiagnosticSeverity

# Initialize client
client = MoAILSPClient(project_root="/path/to/project")

# Get diagnostics
diagnostics = await client.get_diagnostics("src/auth.py")

# Process results
for diag in diagnostics:
    if diag.severity == DiagnosticSeverity.ERROR:
        print(f"Error at line {diag.range.start.line}: {diag.message}")
```

### Completion Conditions

The loop completes when all enabled conditions are met:

| Condition | Default | Description |
|-----------|---------|-------------|
| zero_errors | true | No LSP/compiler errors |
| zero_warnings | false | No warnings (optional) |
| tests_pass | true | All tests pass |
| coverage_threshold | 85 | Minimum coverage % |

## Advanced Patterns

### Custom Completion Conditions

Extend the loop controller to add custom conditions:

```python
def check_custom_conditions(config: dict) -> bool:
    """Add project-specific completion checks."""
    # Example: Check for TODO comments
    todos = count_todo_comments()
    return todos == 0
```

### Integration with CI/CD

```yaml
# GitHub Actions example
- name: Run Ralph Loop
  run: |
    claude -p "/moai-loop --max-iterations 5"
  env:
    MOAI_LOOP_ACTIVE: "true"
```

### Graceful Degradation

When LSP is unavailable, the system falls back to:
1. Linter-based diagnostics (ruff, eslint, etc.)
2. Compiler error detection
3. Test failure detection

## Troubleshooting

### Loop Not Starting
- Check `ralph.enabled` in config
- Verify `MOAI_DISABLE_LOOP_CONTROLLER` is not set
- Ensure state file is writable

### LSP Diagnostics Missing
- Check LSP server configuration in `.lsp.json`
- Verify language server is installed
- Check `MOAI_DISABLE_LSP_DIAGNOSTIC` is not set

### Loop Stuck
- Check max_iterations setting
- Review completion conditions
- Use `/cancel-loop` to reset

## Works Well With

Skills:
- moai-foundation-quality: TRUST 5 validation
- moai-tool-ast-grep: Security scanning patterns
- moai-workflow-testing: TDD integration
- moai-lang-python: Python-specific patterns
- moai-lang-typescript: TypeScript patterns

Agents:
- manager-tdd: TDD implementation
- manager-quality: Quality validation
- expert-debug: Complex debugging

Commands:
- /moai:2-run: TDD implementation
- /moai:3-sync: Documentation sync

## Reference

### Environment Variables

| Variable | Description |
|----------|-------------|
| MOAI_DISABLE_LSP_DIAGNOSTIC | Disable LSP hook |
| MOAI_DISABLE_LOOP_CONTROLLER | Disable loop hook |
| MOAI_LOOP_ACTIVE | Loop active flag |
| MOAI_LOOP_ITERATION | Current iteration |
| CLAUDE_PROJECT_DIR | Project root path |

### File Locations

| File | Purpose |
|------|---------|
| .moai/config/sections/ralph.yaml | Configuration |
| .moai/cache/.moai_loop_state.json | Loop state |
| .claude/hooks/moai/post_tool__lsp_diagnostic.py | LSP hook |
| .claude/hooks/moai/stop__loop_controller.py | Loop hook |

### Supported Languages

LSP diagnostics available for:
- Python (pyright, pylsp)
- TypeScript/JavaScript (tsserver)
- Go (gopls)
- Rust (rust-analyzer)
- Java (jdtls)
- And more via .lsp.json configuration

---

Version: 1.0.0
Last Updated: 2025-01-09
Status: Active
Integration: Claude Code Hooks, LSP Protocol, AST-grep
