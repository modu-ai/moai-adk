---
name: moai-workflow-loop
description: >
  Ralph Engine - Automated feedback loop with LSP diagnostics and AST-grep
  integration for continuous code quality improvement. Use when implementing
  error-driven development, automated fixing, or continuous quality validation
  workflows.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.2.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-11"
  tags: "lsp, ast-grep, feedback-loop, code-quality, automation, diagnostics, ralph"
---

# Ralph Engine

Automated feedback loop system integrating LSP diagnostics, AST-grep security scanning, and test validation for continuous code quality improvement.

## Quick Reference

Core Capabilities:

- LSP Integration: Real-time diagnostics from language servers
- AST-grep Scanning: Structural code analysis and security checks
- Feedback Loop: Iterative error correction until completion conditions met
- Hook System: PostToolUse and Stop hooks for seamless Claude Code integration

Key Components:

- post_tool__lsp_diagnostic: LSP diagnostics after Write/Edit operations
- stop__loop_controller: Loop iteration control
- ralph.yaml: Configuration settings

Commands:

- /moai: One-click Plan-Run-Sync automation (default)
- /moai loop: Start feedback loop
- /moai fix: One-time auto-fix

When to Use:

- Implementing features with zero-error goal
- Automated code quality improvement
- Continuous integration workflows
- Error-driven development patterns

## Implementation Guide

### Architecture Overview

The Ralph Engine follows a layered architecture. User commands such as /moai:loop, /moai:fix, and /moai enter the Command Layer. The Command Layer invokes the Hook System, which contains the PostToolUse Hook for LSP diagnostics and the Stop Hook for loop control. The Hook System connects to Backend Services including the LSP Client (MoAILSPClient), AST-grep Scanner, and Test Runner. Backend Services feed into Completion Check which evaluates whether errors are zero, tests pass, and coverage is met. Based on the Completion Check result, the system either continues the loop or completes.

### Configuration

The ralph.yaml configuration file contains the following sections and settings.

Under the ralph section, enabled controls whether Ralph is active (true by default).

Under the lsp section, auto_start controls automatic language server startup (true by default), timeout_seconds sets the connection timeout (30 seconds default), and graceful_degradation enables fallback to linters when LSP unavailable (true by default).

Under the ast_grep section, enabled controls AST-grep integration (true by default), security_scan enables security rule checking (true by default), and quality_scan enables code quality rule checking (true by default).

Under the loop section, max_iterations sets the maximum loop iterations (10 by default), auto_fix controls automatic fix application (false by default requiring confirmation), and require_confirmation requires user approval before fixes (true by default).

Under the completion subsection of loop, zero_errors requires no LSP or compiler errors (true by default), zero_warnings requires no warnings (false by default as optional), tests_pass requires all tests to pass (true by default), and coverage_threshold sets minimum coverage percentage (85 by default).

Under the hooks section, post_tool_lsp has enabled (true by default) and severity_threshold (error by default). The stop_loop_controller has enabled set to true by default.

### Hook Integration

#### PostToolUse Hook

The PostToolUse hook is triggered after Write and Edit operations. When invoked, Claude Code provides hook input containing the tool_name (such as Write) and tool_input containing the file_path and content.

The hook processes diagnostics and returns hook output with hookSpecificOutput containing the hookEventName (PostToolUse) and additionalContext describing the diagnostic results. For example, the context might report LSP found 2 errors and 3 warnings in file.py, with specific error messages including line numbers.

Exit code 0 indicates no action needed. Exit code 2 indicates attention needed due to errors found.

#### Stop Hook for Loop Controller

The Stop hook is triggered after each Claude response. The hook reads the loop state file located at .moai/cache/.moai_loop_state.json. This state contains active status (true or false), current iteration number, max_iterations limit, last_error_count from previous iteration, and completion_reason when finished.

The hook returns output with hookSpecificOutput containing hookEventName (Stop) and additionalContext reporting loop status. For example, it might report Ralph Loop CONTINUE at Iteration 3 of 10 with 2 Errors, and next actions to fix the remaining errors.

Exit code 0 indicates loop complete or inactive. Exit code 1 indicates continue loop with more work needed.

### LSP Client Usage

The Go LSP client is integrated into the hook system. LSP diagnostics are automatically collected via the post-tool hook (moai hook post-tool-use).

To get diagnostics for a file, call the get_diagnostics method asynchronously with the file path.

Process the returned diagnostics by iterating through each diagnostic object. Check the severity property against DiagnosticSeverity.ERROR to identify errors. Access the line number from diag.range.start.line and the message from diag.message.

### Completion Conditions

The loop completes when all enabled conditions are met.

The zero_errors condition (default true) requires no LSP or compiler errors.

The zero_warnings condition (default false) optionally requires no warnings.

The tests_pass condition (default true) requires all tests to pass.

The coverage_threshold condition (default 85) requires minimum coverage percentage.

## Advanced Patterns

### Custom Completion Conditions

Extend the loop controller with custom conditions by implementing a check function. For example, create a function to count TODO comments in the codebase and return true only when the count reaches zero.

### Integration with CI/CD

For GitHub Actions integration, create a workflow step that runs Claude with the /moai:loop command and max-iterations flag. Set the MOAI_LOOP_ACTIVE environment variable to true to enable loop mode.

### Graceful Degradation

When LSP is unavailable, the system falls back to linter-based diagnostics using tools like ruff or eslint, then to compiler error detection, and finally to test failure detection.

## Troubleshooting

### Loop Not Starting

Check that ralph.enabled is set to true in configuration. Verify MOAI_DISABLE_LOOP_CONTROLLER environment variable is not set. Ensure the state file location is writable.

### LSP Diagnostics Missing

Check LSP server configuration in .lsp.json file. Verify the language server is installed for your language. Check that MOAI_DISABLE_LSP_DIAGNOSTIC environment variable is not set.

### Loop Stuck

Review the max_iterations setting to ensure it allows sufficient iterations. Review completion conditions to verify they are achievable. Send any message to interrupt the loop, or delete the state file (.moai/cache/.moai_loop_state.json) to reset.

## Works Well With

Skills:

- moai-foundation-quality: TRUST 5 validation
- moai-tool-ast-grep: Security scanning patterns
- moai-workflow-testing: DDD integration
- moai-lang-python: Python-specific patterns
- moai-lang-typescript: TypeScript patterns

Agents:

- manager-ddd: DDD implementation
- manager-quality: Quality validation
- expert-debug: Complex debugging

Commands:

- /moai:2-run: DDD implementation
- /moai:3-sync: Documentation sync

## Reference

### Environment Variables

MOAI_DISABLE_LSP_DIAGNOSTIC disables the LSP hook when set.

MOAI_DISABLE_LOOP_CONTROLLER disables the loop hook when set.

MOAI_LOOP_ACTIVE indicates whether the loop is currently active.

MOAI_LOOP_ITERATION contains the current iteration number.

CLAUDE_PROJECT_DIR contains the project root path.

### File Locations

Configuration is stored at .moai/config/sections/ralph.yaml.

Loop state is stored at .moai/cache/.moai_loop_state.json.

The LSP hook is located at .claude/hooks/moai/post_tool__lsp_diagnostic.

The loop hook is located at .claude/hooks/moai/stop__loop_controller.

### Supported Languages

LSP diagnostics are available for Python using pyright or pylsp, TypeScript and JavaScript using tsserver, Go using gopls, Rust using rust-analyzer, Java using jdtls, and additional languages via .lsp.json configuration.

---

Version: 1.2.0
Last Updated: 2026-01-11
Status: Active
Integration: Claude Code Hooks, LSP Protocol, AST-grep
Skill Name: moai-workflow-loop (formerly moai-ralph)
