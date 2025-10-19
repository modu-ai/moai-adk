---
name: moai-lang-shell
description: Shell scripting best practices with bats, shellcheck, and POSIX compliance
allowed-tools:
- Read
- Bash
---

# Shell Expert

## What it does

Provides shell scripting expertise for TDD development, including bats testing framework, shellcheck linting, and POSIX compliance for portable scripts.

## When to use

- "Shell 스크립트 작성", "bats 테스트", "POSIX 호환성"
- Automatically invoked when working with shell script projects
- Shell SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **bats**: Bash Automated Testing System
- **shunit2**: xUnit-style shell testing
- **assert.sh**: Shell assertion library
- Test-driven shell development

**Code Quality**:
- **shellcheck**: Static analysis for shell scripts
- **shfmt**: Shell script formatting
- **bashate**: Style checker

**POSIX Compliance**:
- Portable shell features (sh vs bash)
- Avoid bashisms for portability
- Use `[ ]` instead of `[[ ]]` for POSIX
- Standard utilities (no GNU extensions)

**Shell Patterns**:
- **Error handling**: set -e, set -u, set -o pipefail
- **Exit codes**: Proper use of 0 (success) and non-zero
- **Quoting**: Always quote variables ("$var")
- **Functions**: Modular script organization

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Use `#!/bin/sh` for POSIX, `#!/bin/bash` for Bash
- Check command existence with `command -v`
- Use `$()` over backticks
- Validate input arguments

## Examples

### Example 1: TDD with bats
User: "/alfred:2-build DEPLOY-001"
Claude: (creates RED test with bats, GREEN implementation, REFACTOR with error handling)

### Example 2: Shellcheck validation
User: "shellcheck 실행"
Claude: (runs shellcheck *.sh and reports issues)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Shell-specific review)
- devops-expert (Deployment scripts)
