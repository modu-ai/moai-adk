---

name: moai-lang-julia
description: Julia best practices with Test stdlib, Pkg manager, and scientific computing patterns. Use when writing or reviewing Julia code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Julia Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Julia code discussions, framework guidance, or file extensions such as .jl. |
| Tier | 3 |

## What it does

Provides Julia-specific expertise for TDD development, including Test standard library, Pkg package manager, and high-performance scientific computing patterns.

## When to use

- Engages when the conversation references Julia work, frameworks, or files like .jl.
- "Writing Julia tests", "How to use Test stdlib", "Scientific computing"
- Automatically invoked when working with Julia projects
- Julia SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **Test**: Built-in testing library (@test, @testset)
- **Coverage.jl**: Test coverage analysis
- **BenchmarkTools.jl**: Performance benchmarking

**Package Management**:
- **Pkg**: Built-in package manager
- **Project.toml**: Package configuration
- **Manifest.toml**: Dependency lock file

**Code Quality**:
- **JuliaFormatter.jl**: Code formatting
- **Lint.jl**: Static analysis
- **JET.jl**: Type inference analysis

**Scientific Computing**:
- **Multiple dispatch**: Method specialization on argument types
- **Type stability**: Performance optimization
- **Broadcasting**: Element-wise operations (. syntax)
- **Linear algebra**: Built-in BLAS/LAPACK

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Type annotations for performance-critical code
- Prefer abstract types for function arguments
- Use @inbounds for performance (after bounds checking)
- Profile before optimizing

## Examples
```bash
julia --project -e 'using Pkg; Pkg.test()'
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
- Julia Language. "Documentation." https://docs.julialang.org/en/v1/ (accessed 2025-03-29).
- JuliaFormatter.jl. "JuliaFormatter Documentation." https://domluna.github.io/JuliaFormatter.jl/stable/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Julia-specific review)
- alfred-performance-optimizer (Julia profiling)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
