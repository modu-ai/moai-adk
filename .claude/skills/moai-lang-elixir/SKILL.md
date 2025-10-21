---
name: moai-lang-elixir
description: Elixir best practices with ExUnit, Mix, and OTP patterns
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Elixir Expert

## What it does

Provides Elixir-specific expertise for TDD development, including ExUnit testing, Mix build tool, and OTP (Open Telecom Platform) patterns for concurrent systems.

## When to use

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

### Example 1: TDD with ExUnit
User: "/alfred:2-run SERVER-001"
Claude: (creates RED test with ExUnit, GREEN GenServer implementation, REFACTOR)

### Example 2: Credo analysis
User: "Run Credo Analysis"
Claude: (runs mix credo --strict and reports issues)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Elixir-specific review)
- web-api-expert (Phoenix API development)
