---
name: moai-lang-julia
description: Julia best practices with Test stdlib, Pkg manager, and scientific computing patterns
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Julia Expert

## What it does

Provides Julia-specific expertise for TDD development, including Test standard library, Pkg package manager, and high-performance scientific computing patterns.

## When to use

- "Julia 테스트 작성", "Test stdlib 사용법", "과학 컴퓨팅", "데이터 과학", "수치 계산", "머신러닝"
- "Flux.jl", "DifferentialEquations.jl", "Plots.jl", "DataFrames.jl"
- "고성능 컴퓨팅", "병렬 처리", "GPU 프로그래밍"
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

## Modern Julia (1.10+)

**Recommended Version**: Julia 1.10+ for production, 1.6+ LTS for stability

**Modern Features**:
- **Multiple dispatch**: Dynamic dispatch on all argument types
- **Metaprogramming**: Macros and generated functions
- **Parallel computing** (1.3+): Multi-threading with `Threads.@threads`
- **Package Extensions** (1.9+): Conditional dependency loading
- **Public keyword** (1.11+): Distinguish public vs exported API
- **Triple-quoted string dedent** (1.10+): Automatic indentation removal

**Version Check**:
```bash
julia --version
```

## Package Management Commands

### Using Pkg (Built-in)
```julia
# Enter Pkg REPL mode (press ])
using Pkg

# Add packages
Pkg.add("DataFrames")
Pkg.add("Plots")
Pkg.add(["CSV", "StatsBase"])

# Add specific version
Pkg.add(name="DataFrames", version="1.5")

# Add from GitHub
Pkg.add(url="https://github.com/user/MyPackage.jl")

# Remove packages
Pkg.rm("DataFrames")

# Update packages
Pkg.update()
Pkg.update("DataFrames")

# Check status
Pkg.status()

# Test package
Pkg.test("MyPackage")

# Build package
Pkg.build("MyPackage")

# Garbage collect unused packages
Pkg.gc()
```

**Project.toml Example**:
```toml
name = "MyProject"
uuid = "..."
version = "0.1.0"

[deps]
DataFrames = "a93c6f00-e57d-5684-b7b6-d8193f3e46c0"
Plots = "91a5bcdd-55d7-5caf-9e0b-520d859cae80"
CSV = "336ed68f-0bac-5ca0-87d4-7b16caf5d00b"

[compat]
julia = "1.10"
DataFrames = "1.5"
```

### Common Development Commands
```bash
# Run Julia script
julia script.jl

# Run with arguments
julia script.jl arg1 arg2

# Interactive REPL
julia

# Run tests
julia --project -e 'using Pkg; Pkg.test()'

# Instantiate project (install dependencies)
julia --project -e 'using Pkg; Pkg.instantiate()'

# Generate project
julia -e 'using Pkg; Pkg.generate("MyPackage")'

# Format code
julia -e 'using JuliaFormatter; format(".")'

# Run with multiple threads
julia -t auto script.jl
julia -t 4 script.jl  # 4 threads
```

### Project Structure
```
MyPackage.jl/
├── Project.toml
├── Manifest.toml
├── src/
│   └── MyPackage.jl
├── test/
│   └── runtests.jl
├── docs/
│   ├── make.jl
│   └── src/
│       └── index.md
└── README.md
```

### Testing Example
```julia
# test/runtests.jl
using Test
using MyPackage

@testset "MyPackage Tests" begin
    @testset "Basic functionality" begin
        @test my_function(1, 2) == 3
        @test_throws ArgumentError my_function(-1, 2)
    end

    @testset "Performance" begin
        result = my_function(100, 200)
        @test result isa Int
    end
end
```

## Examples

### Example 1: TDD with Test stdlib
User: "/alfred:2-run COMPUTE-001"
Claude: (creates RED test with @testset, GREEN implementation with type stability, REFACTOR)

### Example 2: Performance optimization
User: "Julia 성능 최적화"
Claude: (profiles code and suggests type-stable refactoring)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Julia-specific review)
- alfred-performance-optimizer (Julia profiling)
