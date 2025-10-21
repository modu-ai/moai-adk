---

name: moai-lang-lua
description: Lua best practices with busted, luacheck, and embedded scripting patterns. Use when writing or reviewing Lua code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Lua Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Lua code discussions, framework guidance, or file extensions such as .lua. |
| Tier | 3 |

## What it does

Provides Lua-specific expertise for TDD development, including busted testing framework, luacheck linting, and embedded scripting patterns for game development and system configuration.

## When to use

- Engages when the conversation references Lua work, frameworks, or files like .lua.
- "Writing Lua tests", "How to use busted", "Embedded scripting"
- Automatically invoked when working with Lua projects
- Lua SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **busted**: Elegant Lua testing framework
- **luassert**: Assertion library
- **lua-coveralls**: Coverage reporting
- BDD-style test writing

**Code Quality**:
- **luacheck**: Lua linter and static analyzer
- **StyLua**: Code formatting
- **luadoc**: Documentation generation

**Package Management**:
- **LuaRocks**: Package manager
- **rockspec**: Package specification

**Lua Patterns**:
- **Tables**: Versatile data structure
- **Metatables**: Operator overloading
- **Closures**: Function factories
- **Coroutines**: Cooperative multitasking

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Use `local` for all variables
- Prefer tables over multiple return values
- Document public APIs
- Avoid global variables

## Examples
```bash
luacheck src && busted
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
- Lua.org. "Programming in Lua." https://www.lua.org/pil/contents.html (accessed 2025-03-29).
- Olivine Labs. "busted." https://olivinelabs.com/busted/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Lua-specific review)
- cli-tool-expert (Lua scripting utilities)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
