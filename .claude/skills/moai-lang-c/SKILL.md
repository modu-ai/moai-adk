---

name: moai-lang-c
description: C best practices with Unity test framework, cppcheck, and Make build system. Use when writing or reviewing C code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# C Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | C code discussions, framework guidance, or file extensions such as .c/.h. |
| Tier | 3 |

## What it does

Provides C-specific expertise for TDD development, including Unity test framework, cppcheck static analysis, and Make build system for system programming.

## When to use

- Engages when the conversation references C work, frameworks, or files like .c/.h.
- "Writing C tests", "Unity test framework", "Embedded C"
- Automatically invoked when working with C projects
- C SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **Unity**: Lightweight C test framework
- **CMock**: Mocking framework for C
- **Ceedling**: Build automation for C
- Test coverage with gcov

**Build Tools**:
- **Make**: Standard build automation
- **CMake**: Modern build system
- **GCC/Clang**: C compilers

**Code Quality**:
- **cppcheck**: Static code analysis
- **Valgrind**: Memory leak detection
- **splint**: Secure programming lint

**C Patterns**:
- **Opaque pointers**: Information hiding
- **Function pointers**: Callback mechanisms
- **Error codes**: Integer return values
- **Manual memory management**: malloc/free discipline

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Always check malloc return values
- Free every malloc
- Avoid buffer overflows (use strncpy, snprintf)
- Use const for read-only parameters
- Initialize all variables

## Examples
```bash
make test && cppcheck src
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
- ISO. "ISO/IEC 9899:2018 Programming Languages — C." (accessed 2025-03-29).
- Cppcheck. "Cppcheck Manual." http://cppcheck.sourceforge.net/manual.pdf (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C-specific review)
- alfred-debugger-pro (C debugging)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
