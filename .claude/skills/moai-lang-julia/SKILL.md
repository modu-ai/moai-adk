---
name: moai-lang-julia
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Julia 1.11+ best practices with Test stdlib, Pkg manager, and scientific computing patterns.
keywords: ['julia', 'test', 'pkg', 'scientific']
allowed-tools:
  - Read
  - Bash
---

# Lang Julia Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-julia |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language |

---

## What It Does

Julia 1.11+ best practices with Test stdlib, Pkg manager, and scientific computing patterns. Provides comprehensive guidance for writing high-performance scientific computing code with strong type safety and multiple dispatch.

**Key capabilities**:
- ✅ Best practices enforcement for Julia domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (Julia 1.11.0, 2025-10-22)
- ✅ TDD workflow support with Test.jl
- ✅ Performance optimization strategies
- ✅ Multiple dispatch patterns
- ✅ Type system mastery

---

## When to Use

**Automatic triggers**:
- `.jl` file patterns detected
- `Project.toml` or `Manifest.toml` present
- SPEC implementation (`/alfred:2-run`)
- Code review requests for Julia codebases

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new scientific computing features
- Troubleshoot performance issues
- Optimize numerical algorithms

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Julia** | 1.11.0 | Compiler/Runtime | ✅ Current |
| **Test.jl** | stdlib | Unit Testing | ✅ Current |
| **Pkg.jl** | stdlib | Package Manager | ✅ Current |
| **BenchmarkTools.jl** | 1.5.0 | Performance Testing | ✅ Current |
| **Documenter.jl** | 1.7.0 | Documentation | ✅ Current |
| **Revise.jl** | 3.5.18 | Interactive Development | ✅ Current |

---

## Inputs

- `src/` directory with Julia source files
- `test/` directory with test suites
- `Project.toml` with dependencies
- `Manifest.toml` with locked versions
- Benchmark scripts and datasets

## Outputs

- Test execution plan with coverage reports
- Performance benchmarking results
- TRUST 5 compliance review
- Type stability analysis
- Migration guidance for version updates

## Failure Modes

- When Julia runtime is not installed or outdated
- When required packages are missing from Project.toml
- When test coverage falls below 85%
- When type instabilities cause performance degradation
- When allocations exceed expected thresholds

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates
- Integration with `moai-essentials-perf` for performance optimization

---

## Core Language Principles

### 1. Multiple Dispatch

Julia's core paradigm - functions specialize on all argument types:

```julia
# ✅ GOOD: Multiple dispatch with type specialization
function process(data::Vector{Float64})
    # Specialized for Float64 vectors
    sum(data) / length(data)
end

function process(data::Vector{Int})
    # Specialized for Int vectors
    sum(data) ÷ length(data)
end

function process(data::Matrix{Float64})
    # Specialized for Float64 matrices
    mean(data, dims=1)
end

# ❌ BAD: Single dispatch with type checking
function process(data)
    if isa(data, Vector{Float64})
        # Manual type checking
    elseif isa(data, Vector{Int})
        # ...
    end
end
```

**Best practices**:
- Define methods for specific type combinations
- Avoid type checks inside functions
- Use abstract types for generic interfaces
- Leverage compiler optimizations via dispatch

### 2. Type System Mastery

**Type annotations for clarity and performance**:

```julia
# ✅ GOOD: Concrete types in struct definitions
struct Point{T<:Real}
    x::T
    y::T
end

# Type-stable function
function distance(p1::Point{T}, p2::Point{T}) where T
    sqrt((p1.x - p2.x)^2 + (p1.y - p2.y)^2)
end

# ❌ BAD: Abstract types in struct fields (type instability)
struct BadPoint
    x::Real  # Abstract type causes boxing
    y::Real
end

# ✅ GOOD: Parametric types for flexibility
struct Container{T}
    items::Vector{T}
    count::Int
end

function add_item!(c::Container{T}, item::T) where T
    push!(c.items, item)
    c.count += 1
end
```

**Type hierarchy design**:

```julia
# Abstract type hierarchy
abstract type Shape end
abstract type Polygon <: Shape end

struct Triangle <: Polygon
    vertices::NTuple{3,Point{Float64}}
end

struct Rectangle <: Polygon
    width::Float64
    height::Float64
end

# Generic functions work on abstract types
area(s::Shape) = error("Not implemented for $(typeof(s))")
area(t::Triangle) = # ... compute triangle area
area(r::Rectangle) = r.width * r.height
```

### 3. Performance Optimization

**Key principles for high-performance code**:

```julia
# ✅ GOOD: Type-stable, non-allocating code
function sum_of_squares!(result::Vector{Float64}, x::Vector{Float64})
    @assert length(result) == length(x)
    @inbounds for i in eachindex(x)
        result[i] = x[i]^2
    end
    sum(result)
end

# ❌ BAD: Type-unstable, allocating code
function sum_of_squares(x)
    result = []  # Type-unstable array
    for val in x
        push!(result, val^2)  # Repeated allocations
    end
    sum(result)
end

# ✅ GOOD: Vectorized operations with broadcasting
function transform!(output::Vector{Float64}, input::Vector{Float64}, scale::Float64)
    output .= scale .* sin.(input) .+ cos.(input)
end

# Use @views to avoid allocations
function process_subarray(data::Matrix{Float64}, row::Int)
    @views return sum(data[row, :])  # No allocation
end
```

**Performance annotations**:

```julia
# @inbounds: Skip bounds checking (use carefully!)
function fast_sum(arr::Vector{Float64})
    total = 0.0
    @inbounds for i in 1:length(arr)
        total += arr[i]
    end
    total
end

# @simd: Enable SIMD vectorization
function simd_sum(arr::Vector{Float64})
    total = 0.0
    @simd for x in arr
        total += x
    end
    total
end

# @fastmath: Aggressive floating-point optimizations
function fast_math_compute(x::Float64)
    @fastmath sqrt(x^2 + 1.0) / (x + 1.0)
end
```

---

## Project Structure

### Standard Julia Package Layout

```
MyPackage/
├── Project.toml          # Package metadata and dependencies
├── Manifest.toml         # Locked dependency versions
├── README.md
├── LICENSE
├── src/
│   ├── MyPackage.jl     # Main module file
│   ├── types.jl         # Type definitions
│   ├── algorithms.jl    # Core algorithms
│   └── utils.jl         # Utility functions
├── test/
│   ├── runtests.jl      # Test entry point
│   ├── test_types.jl    # Type tests
│   └── test_algorithms.jl
├── docs/
│   ├── make.jl          # Documentation build script
│   └── src/
│       ├── index.md
│       └── api.md
└── benchmark/
    └── benchmarks.jl    # Performance benchmarks
```

### Project.toml Example

```toml
name = "MyPackage"
uuid = "12345678-1234-5678-1234-567812345678"
authors = ["Your Name <your.email@example.com>"]
version = "0.1.0"

[deps]
LinearAlgebra = "37e2e46d-f89d-539d-b4ee-838fcccc9c8e"
Statistics = "10745b16-79ce-11e8-11f9-7d13ad32a3b2"
Test = "8dfed614-e22c-5e08-85e1-65c5234f0b40"

[compat]
julia = "1.11"

[extras]
Test = "8dfed614-e22c-5e08-85e1-65c5234f0b40"
BenchmarkTools = "6e4b80f9-dd63-53aa-95a3-0cdb28fa8baf"

[targets]
test = ["Test", "BenchmarkTools"]
```

---

## Testing with Test.jl

### Basic Test Structure

```julia
# test/runtests.jl
using MyPackage
using Test

@testset "MyPackage.jl" begin
    @testset "Type constructors" begin
        @test Point(1.0, 2.0) isa Point{Float64}
        @test Point(1, 2) isa Point{Int}
    end

    @testset "Distance calculations" begin
        p1 = Point(0.0, 0.0)
        p2 = Point(3.0, 4.0)
        @test distance(p1, p2) ≈ 5.0
        @test distance(p1, p1) ≈ 0.0
    end

    @testset "Error handling" begin
        @test_throws ArgumentError invalid_operation()
        @test_throws DomainError sqrt(-1.0)
    end
end
```

### Advanced Testing Patterns

```julia
# Property-based testing
@testset "Commutative operations" begin
    for _ in 1:100
        a, b = rand(Float64, 2)
        @test my_add(a, b) ≈ my_add(b, a)
    end
end

# Approximate equality for floating-point
@testset "Numerical stability" begin
    result = compute_integral(f, 0.0, 1.0)
    expected = 0.333333333
    @test result ≈ expected atol=1e-6 rtol=1e-6
end

# Testing performance (non-regression)
@testset "Performance" begin
    data = rand(1000)
    @test (@elapsed fast_algorithm(data)) < 0.001

    # Allocation testing
    allocs = @allocated process_data(data)
    @test allocs < 1000  # Less than 1KB allocated
end

# Testing type stability
using Test: @inferred

@testset "Type stability" begin
    x = rand(Float64, 100)
    @inferred sum_of_squares(x)  # Fails if return type not inferred
end
```

### Test Coverage

```bash
# Run tests with coverage
julia --project=. --code-coverage=user test/runtests.jl

# Generate coverage report
using Coverage
coverage = process_folder()
covered_lines, total_lines = get_summary(coverage)
percentage = covered_lines / total_lines * 100
@assert percentage >= 85.0 "Coverage below 85%: $(round(percentage, digits=2))%"
```

---

## Package Management with Pkg.jl

### Basic Package Operations

```julia
using Pkg

# Activate project environment
Pkg.activate(".")

# Add dependencies
Pkg.add("DataFrames")
Pkg.add(["Plots", "StatsBase"])

# Add development dependency
Pkg.add("BenchmarkTools", target="test")

# Update dependencies
Pkg.update()

# Remove packages
Pkg.rm("UnusedPackage")

# Install all dependencies
Pkg.instantiate()

# Check for outdated packages
Pkg.status()

# Resolve dependency conflicts
Pkg.resolve()
```

### Version Constraints

```julia
# Add specific version
Pkg.add(name="DataFrames", version="1.6.0")

# Add version range
Pkg.add(name="Plots", version="1.38.0-1.39.0")

# Add from URL (development)
Pkg.add(url="https://github.com/username/Package.jl")

# Add local package
Pkg.develop(path="/path/to/local/package")
```

### Environment Management

```bash
# Create new environment
julia --project=@MyProject

# Switch between environments
julia --project=.           # Current directory
julia --project=@.         # Nearest Project.toml
julia --project=@MyEnv     # Named environment

# Shared environments
export JULIA_PROJECT=/path/to/project
```

---

## Performance Optimization

### Benchmarking with BenchmarkTools.jl

```julia
using BenchmarkTools

# Basic benchmark
@benchmark sum($data) setup=(data=rand(1000))

# Compare implementations
data = rand(10000)

@btime my_sum_v1($data)      # 15.2 μs
@btime my_sum_v2($data)      # 8.4 μs  ← Winner

# Profile allocations
@btime process_data($data) evals=1 samples=1000

# Benchmark suite
suite = BenchmarkGroup()
suite["sum"] = @benchmarkable sum($data) setup=(data=rand(1000))
suite["mean"] = @benchmarkable mean($data) setup=(data=rand(1000))

results = run(suite, verbose=true)
```

### Type Stability Analysis

```julia
using Test: @inferred

# Check type stability
function analyze_stability()
    @code_warntype my_function(args...)  # Manual inspection

    # Automated check
    @inferred my_function(args...)  # Throws if type-unstable
end

# Common type instabilities
function type_unstable(x)
    if x > 0
        return x          # Returns Int/Float
    else
        return "negative" # Returns String - TYPE UNSTABLE!
    end
end

# Fixed version
function type_stable(x)
    x > 0 ? Float64(x) : -1.0  # Always returns Float64
end
```

### Memory Optimization

```julia
# Pre-allocate arrays
function efficient_computation(n::Int)
    result = Vector{Float64}(undef, n)  # Pre-allocated
    @inbounds for i in 1:n
        result[i] = sin(i) + cos(i)
    end
    result
end

# Avoid global variables (use const)
const GLOBAL_CONSTANT = 42  # Type-stable global

# Use views instead of slices
function process_columns(matrix::Matrix{Float64})
    for col in eachcol(matrix)  # Returns views, not copies
        process_vector(col)
    end
end

# Stack allocation with StaticArrays
using StaticArrays

function fast_3d_transform(v::SVector{3,Float64})
    # Stack-allocated, no heap allocation
    rotation = SMatrix{3,3,Float64}([1 0 0; 0 1 0; 0 0 1])
    rotation * v
end
```

---

## Scientific Computing Patterns

### Linear Algebra

```julia
using LinearAlgebra

# Matrix operations
A = rand(100, 100)
b = rand(100)

# Solve linear system
x = A \ b  # Efficient solver selection

# Eigenvalues/eigenvectors
λ, V = eigen(A)

# Singular value decomposition
U, Σ, Vt = svd(A)

# Specialized matrix types
using LinearAlgebra: Symmetric, Diagonal, UpperTriangular

S = Symmetric(A)      # Exploit symmetry
D = Diagonal(1:10)    # Diagonal matrix
U = UpperTriangular(A) # Upper triangular

# In-place operations (avoid allocations)
mul!(C, A, B)         # C = A * B (in-place)
ldiv!(A, b)          # Solve Ax = b, overwriting b
```

### Numerical Integration and Differentiation

```julia
# Numerical integration
using QuadGK

f(x) = sin(x) * exp(-x)
integral, error = quadgk(f, 0, Inf)

# Automatic differentiation
using ForwardDiff

f(x) = x[1]^2 + x[2]^2
gradient(f, [1.0, 2.0])  # [2.0, 4.0]

# Symbolic computation
using Symbolics

@variables x y
expr = x^2 + 2x*y + y^2
expand(expr)
substitute(expr, Dict(x => 1.0, y => 2.0))
```

### Statistical Computing

```julia
using Statistics, StatsBase

# Descriptive statistics
data = rand(1000)
μ = mean(data)
σ = std(data)
med = median(data)

# Weighted statistics
weights = Weights(rand(1000))
wmean = mean(data, weights)

# Histograms and binning
hist = fit(Histogram, data, nbins=50)

# Correlation
using Statistics: cor, cov
correlation_matrix = cor(data_matrix)
```

---

## Code Quality Standards

### Naming Conventions

```julia
# ✅ GOOD: Clear, descriptive names
function calculate_average(values::Vector{Float64})::Float64
    sum(values) / length(values)
end

const MAX_ITERATIONS = 1000
const DEFAULT_TOLERANCE = 1e-6

struct DataPoint
    timestamp::Float64
    value::Float64
end

# Type names: CamelCase
# Functions: snake_case
# Constants: UPPER_CASE
# Modules: CamelCase
```

### Documentation Standards

```julia
"""
    distance(p1::Point, p2::Point) -> Float64

Compute the Euclidean distance between two points.

# Arguments
- `p1::Point`: First point
- `p2::Point`: Second point

# Returns
- `Float64`: Distance between points

# Examples
```julia
julia> p1 = Point(0.0, 0.0)
julia> p2 = Point(3.0, 4.0)
julia> distance(p1, p2)
5.0
```

# Notes
This function is type-stable and does not allocate memory.

# See also
- [`Point`](@ref): Point type definition
"""
function distance(p1::Point{T}, p2::Point{T}) where T
    sqrt((p1.x - p2.x)^2 + (p1.y - p2.y)^2)
end
```

### Error Handling

```julia
# ✅ GOOD: Specific error types with messages
function divide(a::Float64, b::Float64)
    if b == 0.0
        throw(DomainError(b, "Division by zero"))
    end
    a / b
end

# Use @assert for debugging checks
function process_data(data::Vector{Float64})
    @assert !isempty(data) "Data cannot be empty"
    @assert all(x -> x >= 0, data) "All values must be non-negative"
    # ... processing
end

# Custom error types
struct DataProcessingError <: Exception
    msg::String
end

Base.showerror(io::IO, e::DataProcessingError) = print(io, "DataProcessingError: ", e.msg)
```

---

## TRUST 5 Principles (Julia Edition)

### T - Test First

**Test.jl workflow**:

```julia
# test/test_calculator.jl
@testset "Calculator operations" begin
    @testset "Addition" begin
        @test add(2, 3) == 5
        @test add(-1, 1) == 0
        @test add(0.1, 0.2) ≈ 0.3
    end

    @testset "Division" begin
        @test divide(6, 2) == 3.0
        @test_throws DomainError divide(1, 0)
    end
end

# Run specific test file
include("test/test_calculator.jl")

# Coverage requirement: ≥85%
using Coverage
coverage = process_folder("src")
covered_lines, total_lines = get_summary(coverage)
coverage_pct = covered_lines / total_lines * 100
@test coverage_pct >= 85.0
```

### R - Readable Code

**JuliaFormatter.jl for consistent style**:

```bash
# Install formatter
julia -e 'using Pkg; Pkg.add("JuliaFormatter")'

# Format entire project
julia -e 'using JuliaFormatter; format(".")'

# Format single file
julia -e 'using JuliaFormatter; format_file("src/MyModule.jl")'

# CI integration (.github/workflows/format.yml)
- name: Check formatting
  run: |
    julia -e 'using JuliaFormatter; @assert format(".", overwrite=false)'
```

**.JuliaFormatter.toml**:

```toml
indent = 4
margin = 92
always_for_in = true
whitespace_typedefs = true
whitespace_ops_in_indices = true
remove_extra_newlines = true
```

### U - Unified Architecture

**Type-stable interfaces**:

```julia
# Abstract interface
abstract type DataProcessor end

# Concrete implementations with consistent types
struct BatchProcessor <: DataProcessor
    batch_size::Int
    buffer::Vector{Float64}
end

struct StreamProcessor <: DataProcessor
    chunk_size::Int
    buffer::Vector{Float64}
end

# Unified interface function
function process(p::DataProcessor, data::Vector{Float64})::Vector{Float64}
    # Implementation specific to processor type
end

# Type assertions
@assert process(BatchProcessor(10, []), Float64[]) isa Vector{Float64}
```

### S - Secured

**Security best practices**:

```julia
# Input validation
function load_file(path::String)
    if !isfile(path)
        throw(ArgumentError("File does not exist: $path"))
    end
    if !endswith(path, ".jl")
        throw(ArgumentError("Only .jl files allowed"))
    end
    # Sanitize path
    safe_path = abspath(path)
    # Read with error handling
    try
        return read(safe_path, String)
    catch e
        @error "Failed to read file" path exception=e
        rethrow(e)
    end
end

# Avoid eval() in production
# ❌ BAD
eval(Meta.parse(user_input))

# ✅ GOOD: Use safe alternatives
allowed_functions = Dict("sin" => sin, "cos" => cos)
func = get(allowed_functions, user_input, nothing)
if func !== nothing
    result = func(x)
end
```

### T - Trackable

**TAG system integration**:

```julia
# @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: test/test_calculator.jl
"""
Calculator module implementing basic arithmetic operations.

TAG: @CODE:CALC-001
"""
module Calculator

# @CODE:CALC-001:ADD
function add(a::T, b::T) where T<:Number
    a + b
end

# @CODE:CALC-001:DIV
function divide(a::T, b::T) where T<:Number
    if b == zero(T)
        throw(DomainError(b, "Division by zero"))
    end
    a / b
end

end # module

# Test file with TAG reference
# @TEST:CALC-001 | CODE: src/Calculator.jl | SPEC: SPEC-CALC-001.md
```

---

## Development Workflow

### Interactive Development with Revise.jl

```julia
# Add to ~/.julia/config/startup.jl
using Revise

# Start REPL with Revise enabled
julia

# Load package (changes auto-reload)
using MyPackage

# Edit source files - changes apply immediately
# No need to restart REPL
```

### Debugging with Debugger.jl

```julia
using Debugger

# Enter debugger
@enter my_function(args...)

# Breakpoints
@breakpoint my_function(args...)

# Step through code
# n: next line
# s: step into function
# c: continue
# q: quit debugger
```

### REPL Tips

```julia
# Help mode (?)
?sum  # Show documentation

# Shell mode (;)
;ls  # Run shell commands

# Package mode (])
]add Package  # Add packages
]status       # Show installed packages
]test         # Run tests

# Search mode (Ctrl+R)
# Search command history
```

---

## Common Patterns and Idioms

### Broadcasting

```julia
# Apply function element-wise
data = [1, 2, 3, 4, 5]
result = sin.(data)  # [sin(1), sin(2), ...]

# Multi-argument broadcasting
a = [1, 2, 3]
b = [4, 5, 6]
c = a .+ b  # [5, 7, 9]

# Broadcasting with scalars
scaled = data .* 2  # [2, 4, 6, 8, 10]

# Custom functions
function my_func(x, y)
    x^2 + y
end
results = my_func.(a, b)
```

### Splatting and Slurping

```julia
# Splat operator (...)
args = [1, 2, 3]
result = my_function(args...)  # my_function(1, 2, 3)

# Slurp operator
function print_all(first, rest...)
    println("First: $first")
    println("Rest: $rest")
end
print_all(1, 2, 3, 4)  # rest = (2, 3, 4)
```

### Generators and Comprehensions

```julia
# List comprehension
squares = [x^2 for x in 1:10]

# Generator expression (lazy)
sum_of_squares = sum(x^2 for x in 1:1000000)  # No intermediate array

# Filtered comprehension
evens = [x for x in 1:10 if iseven(x)]

# Nested comprehension
matrix = [i+j for i in 1:3, j in 1:3]
```

### Pipe and Composition

```julia
# Pipe operator
result = data |>
         x -> filter(iseven, x) |>
         x -> map(y -> y^2, x) |>
         sum

# Function composition
transform = sqrt ∘ abs ∘ sin
result = transform(x)  # sqrt(abs(sin(x)))
```

---

## Advanced Topics

### Metaprogramming

```julia
# Macros
macro time_it(expr)
    quote
        t0 = time()
        result = $expr
        elapsed = time() - t0
        println("Elapsed: $elapsed seconds")
        result
    end
end

@time_it expensive_computation()

# Generated functions
@generated function my_generic(x::T) where T
    if T <: Integer
        return :(x + 1)
    else
        return :(x * 2)
    end
end
```

### Parallel and Distributed Computing

```julia
using Distributed

# Add worker processes
addprocs(4)

# Parallel map
@everywhere function slow_function(x)
    sleep(1)
    x^2
end

results = pmap(slow_function, 1:100)

# Parallel for loop
@distributed (+) for i in 1:1000
    compute_value(i)
end

# Shared arrays
using SharedArrays
arr = SharedArray{Float64}(1000)
@distributed for i in 1:1000
    arr[i] = compute_value(i)
end
```

### GPU Computing

```julia
using CUDA

# Transfer to GPU
x_gpu = cu(x)
y_gpu = cu(y)

# GPU operations
z_gpu = x_gpu .+ y_gpu

# Transfer back to CPU
z = Array(z_gpu)

# GPU kernels
function gpu_kernel!(output, input)
    index = threadIdx().x
    output[index] = sin(input[index])
end
```

---

## Migration Guide

### From Julia 1.9 to 1.11

**Breaking changes**:
- Package loading performance improvements
- Enhanced inference capabilities
- New Base.@constprop annotation

**Recommended updates**:
```julia
# Update Project.toml
[compat]
julia = "1.11"

# Update packages
using Pkg
Pkg.update()

# Check for deprecation warnings
julia --depwarn=yes test/runtests.jl
```

---

## References (Latest Documentation)

_Documentation links updated 2025-10-22_

**Official Documentation**:
- Julia Manual: https://docs.julialang.org/en/v1.11/
- Julia Packages: https://julialang.org/packages/
- JuliaHub: https://juliahub.com/

**Style Guides**:
- Official Style Guide: https://docs.julialang.org/en/v1.11/manual/style-guide/
- BlueStyle: https://github.com/invenia/BlueStyle

**Learning Resources**:
- Think Julia: https://benlauwens.github.io/ThinkJulia.jl/latest/
- Julia Academy: https://juliaacademy.com/
- JuliaLang Discourse: https://discourse.julialang.org/

**Performance**:
- Performance Tips: https://docs.julialang.org/en/v1.11/manual/performance-tips/
- BenchmarkTools.jl: https://github.com/JuliaCI/BenchmarkTools.jl

---

## Changelog

- **v2.0.0** (2025-10-22): Comprehensive expansion to 1,200+ lines with Julia 1.11.0, advanced patterns, performance optimization, TRUST 5 integration, complete testing and package management workflows
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)
- `moai-essentials-perf` (performance optimization)
- `moai-domain-data-science` (scientific computing context)
- `moai-domain-ml` (machine learning workflows)

---

## Best Practices Summary

✅ **DO**:
- Use multiple dispatch as the primary paradigm
- Write type-stable code for performance
- Pre-allocate arrays when possible
- Use `@inbounds` and `@simd` judiciously
- Maintain test coverage ≥85%
- Document all public APIs with docstrings
- Use broadcasting (`.`) for element-wise operations
- Leverage standard library (LinearAlgebra, Statistics)
- Profile before optimizing
- Use BenchmarkTools for performance testing

❌ **DON'T**:
- Use abstract types in struct fields
- Write type-unstable functions
- Use global variables without `const`
- Allocate inside tight loops
- Mix tabs and spaces
- Ignore compiler warnings
- Use `eval()` in production code
- Overuse macros when functions suffice
- Skip bounds checking without benchmarks
- Use `Any` type unnecessarily

---

## Quick Reference Card

```julia
# Package management
using Pkg
Pkg.activate(".")
Pkg.add("Package")
Pkg.test()

# Testing
using Test
@testset "Name" begin
    @test expr
    @test_throws ErrorType expr
end

# Benchmarking
using BenchmarkTools
@benchmark function_call(args)
@btime optimized_function($args)

# Type stability
@code_warntype function_call(args)
@inferred function_call(args)

# Documentation
"""
    func(arg) -> ReturnType

Description.
"""
function func(arg)
    # ...
end

# Broadcasting
result = func.(array)
output .= func.(input)

# Performance
@inbounds for i in 1:n
    # Skip bounds checking
end

@simd for x in array
    # SIMD vectorization
end
```

---

**End of Julia Language Skill v2.0.0**
