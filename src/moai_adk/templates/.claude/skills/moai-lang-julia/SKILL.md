---
name: moai-lang-julia
description: Julia best practices with Test stdlib, Pkg manager, and scientific computing
  patterns
allowed-tools:
- Read
- Bash
---

# Julia Expert

## What it does

Provides Julia-specific expertise for TDD development, including Test standard library, Pkg package manager, and high-performance scientific computing patterns.

## When to use

- "Julia 테스트 작성", "Test stdlib 사용법", "과학 컴퓨팅"
- Automatically invoked when working with Julia projects
- Julia SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with Test stdlib
User: "/alfred:2-build COMPUTE-001"
Claude: (creates RED test with @testset, GREEN implementation with type stability, REFACTOR)

### Example 2: Performance optimization
User: "Julia 성능 최적화"
Claude: (profiles code and suggests type-stable refactoring)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Julia-specific review)
- alfred-performance-optimizer (Julia profiling)
