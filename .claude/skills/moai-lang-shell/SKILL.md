---

name: moai-lang-shell
description: Shell scripting best practices with bats, shellcheck, and POSIX compliance. Use when writing or reviewing Shell scripts code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Shell Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Shell code discussions, framework guidance, or file extensions such as .sh/.bash. |
| Tier | 3 |

## What it does

Provides shell scripting expertise for TDD development, including bats testing framework, shellcheck linting, and POSIX compliance for portable scripts.

## When to use

- Engages when the conversation references Shell work, frameworks, or files like .sh/.bash.
- "Writing shell scripts", "bats testing", "POSIX compatibility"
- Automatically invoked when working with shell script projects
- Shell SPEC implementation (`/alfred:2-run`)

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
```bash
bats tests && shellcheck scripts/*.sh
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- GNU. "Bash Reference Manual." https://www.gnu.org/software/bash/manual/bash.html (accessed 2025-03-29).
- koalaman. "ShellCheck." https://www.shellcheck.net/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Shell-specific review)
- devops-expert (Deployment scripts)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
