# moai-lang-julia - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with Pkg.jl

```julia
# Create a new package
using Pkg
Pkg.generate("MyPackage")
cd("MyPackage")

# Activate the package environment
Pkg.activate(".")

# Add Test.jl (stdlib, already available)
# Add development dependencies if needed
Pkg.add("Aqua")  # For quality assurance tests
```

**Project.toml structure**:
```toml
name = "MyPackage"
uuid = "12345678-1234-1234-1234-123456789012"
authors = ["Your Name <you@example.com>"]
version = "0.1.0"

[deps]

[compat]
julia = "1.11"

[extras]
Test = "8dfed614-e22c-5e08-85e1-65c5234f0b40"
Aqua = "4c88cf16-eb10-579e-8560-4a9242c79595"

[targets]
test = ["Test", "Aqua"]
```

## Example 2: TDD Workflow with Test.jl

**RED: Write failing test** (`test/runtests.jl`):
```julia
using MyPackage
using Test

@testset "Calculator Tests" begin
    @testset "Addition" begin
        calc = Calculator()

        @test add(calc, 2, 3) == 5
        @test add(calc, -1, -2) == -3
        @test add(calc, 0, 5) == 5

        # Test floating point with tolerance
        @test add(calc, 0.1, 0.2) ≈ 0.3 atol=1e-10
    end
end
```

**GREEN: Implement feature** (`src/MyPackage.jl`):
```julia
module MyPackage

export Calculator, add

struct Calculator end

function add(calc::Calculator, a::Number, b::Number)
    return a + b
end

end # module
```

**REFACTOR: Add type constraints and documentation**:
```julia
module MyPackage

export Calculator, add

"""
    Calculator

A calculator struct for basic arithmetic operations.
"""
struct Calculator end

"""
    add(calc::Calculator, a::Number, b::Number) -> Number

Add two numbers and return their sum.

# Arguments
- `calc::Calculator`: The calculator instance
- `a::Number`: First operand
- `b::Number`: Second operand

# Returns
- `Number`: The sum of a and b

# Examples
```jldoctest
julia> calc = Calculator()
Calculator()

julia> add(calc, 2, 3)
5
```
"""
function add(calc::Calculator, a::T, b::T) where T<:Number
    return a + b
end

end # module
```

## Example 3: Running Tests

```bash
# Run all tests
julia --project=. -e 'using Pkg; Pkg.test()'

# Or use the REPL
julia --project=.
```

```julia
# In Julia REPL
] test  # Shortcut for Pkg.test()

# Run tests with coverage
using Pkg
Pkg.test(coverage=true)

# Generate coverage report
using Coverage
coverage = process_folder()
LCOV.writefile("coverage.info", coverage)
```

**Expected output**:
```
Test Summary:   | Pass  Total
Calculator Tests|    4      4
    Testing MyPackage tests passed
```

## Example 4: Quality Assurance with Aqua.jl

```julia
# test/aqua_tests.jl
using MyPackage
using Test
using Aqua

@testset "Aqua.jl quality assurance" begin
    Aqua.test_all(MyPackage)
end
```

**What Aqua checks**:
- Method ambiguities
- Undefined exports
- Unbound type parameters
- Missing compatibility constraints
- Persistent tasks
- Project extras and stale dependencies

## Example 5: Floating Point Comparisons

```julia
using Test

@testset "Floating Point Tests" begin
    # Use ≈ (isapprox) for floating point comparisons
    @test 0.1 + 0.2 ≈ 0.3

    # Specify tolerance
    @test 0.1 + 0.2 ≈ 0.3 atol=1e-10
    @test 1.0 ≈ 1.0001 rtol=0.01

    # Test for NaN and Inf
    @test isnan(0/0)
    @test isinf(1/0)
    @test !isfinite(1/0)
end
```

## Example 6: Test Organization with Nested @testset

```julia
using Test

@testset "MyPackage" begin
    @testset "Core Functionality" begin
        @testset "Addition" begin
            @test 1 + 1 == 2
        end

        @testset "Subtraction" begin
            @test 2 - 1 == 1
        end
    end

    @testset "Edge Cases" begin
        @testset "Overflow" begin
            @test_throws OverflowError typemax(Int) + 1
        end

        @testset "Division by zero" begin
            @test_throws DivideError 1 ÷ 0
        end
    end
end
```

## Example 7: Setup and Teardown

```julia
using Test

@testset "Database Tests" begin
    # Setup
    db = connect_database()

    try
        @testset "Insert" begin
            result = insert!(db, "test_data")
            @test result.success
        end

        @testset "Query" begin
            data = query(db, "test_data")
            @test !isempty(data)
        end
    finally
        # Teardown
        close(db)
    end
end
```

---

_For complete API reference and configuration options, see reference.md_
