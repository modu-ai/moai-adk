---

name: moai-lang-elixir
description: Elixir best practices with ExUnit, Mix, and OTP patterns. Use when writing or reviewing Elixir code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Elixir Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Elixir code discussions, framework guidance, or file extensions such as .ex/.exs. |
| Tier | 3 |

## What it does

Provides Elixir-specific expertise for TDD development, including ExUnit testing, Mix build tool, and OTP (Open Telecom Platform) patterns for concurrent systems.

## When to use

- Engages when the conversation references Elixir work, frameworks, or files like .ex/.exs.
- "Writing Elixir tests", "How to use ExUnit", "OTP patterns"
- Automatically invoked when working with Elixir/Phoenix projects
- Elixir SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **ExUnit**: Built-in test framework
- **Mox**: Mocking library
- **StreamData**: Property-based testing
- Test coverage with `mix test --cover`

**Build Tools**:
- **Mix**: Build tool and project manager
- **mix.exs**: Project configuration
- **Hex**: Package manager

**Code Quality**:
- **Credo**: Static code analysis
- **Dialyzer**: Type checking
- **mix format**: Code formatting

**OTP Patterns**:
- **GenServer**: Generic server behavior
- **Supervisor**: Process supervision
- **Application**: Application behavior
- **Task**: Async/await operations

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Pattern matching over conditionals
- Pipe operator (|>) for data transformations
- Immutable data structures
- "Let it crash" philosophy with supervisors

## Examples
```bash
mix test && mix credo --strict
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
- Elixir Lang. "Getting Started." https://elixir-lang.org/getting-started/introduction.html (accessed 2025-03-29).
- Credo. "Credo — The Elixir Linter." https://hexdocs.pm/credo/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Elixir-specific review)
- web-api-expert (Phoenix API development)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
