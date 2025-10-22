# moai-lang-julia - CLI Reference

_Last updated: 2025-10-22_

## Official Documentation Links

- **Julia**: https://docs.julialang.org/
- **Test.jl**: https://docs.julialang.org/en/v1/stdlib/Test/
- **Pkg.jl**: https://pkgdocs.julialang.org/
- **Aqua.jl**: https://github.com/JuliaTesting/Aqua.jl

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Status |
|------|---------|--------------|--------|
| **Julia** | 1.11.0 | 2024-10 | ✅ Latest |
| **Test.jl** | stdlib | Built-in | ✅ Stable |
| **Pkg.jl** | 1.11.0 | Built-in | ✅ Stable |

## Installation

### Install Julia

```bash
# macOS (Homebrew)
brew install julia

# Ubuntu/Debian
sudo apt-get install julia

# Or download from official site
# https://julialang.org/downloads/

# Verify installation
julia --version  # Should show v1.11 or higher
```

### Package Management

```julia
# Enter package mode in REPL (press ])
]

# Or use Pkg directly
using Pkg

# Activate project environment
Pkg.activate(".")

# Add package
Pkg.add("PackageName")

# Add development dependency
Pkg.add("Aqua"; target="test")

# Update all packages
Pkg.update()

# Remove package
Pkg.rm("PackageName")

# Check package status
Pkg.status()
```

## Common Commands

### Running Tests

```bash
# Run tests from command line
julia --project=. -e 'using Pkg; Pkg.test()'

# Run tests with coverage
julia --project=. -e 'using Pkg; Pkg.test(coverage=true)'

# Run specific test file
julia --project=. test/specific_test.jl
```

```julia
# In Julia REPL
using Pkg
Pkg.test()  # Run all tests

# Package mode shortcut (press ])
] test
```

### Project Management

```julia
# Create new package
using Pkg
Pkg.generate("MyPackage")

# Initialize existing directory as package
Pkg.activate(".")
Pkg.add("SomePackage")  # This creates Project.toml

# Instantiate dependencies (like npm install)
Pkg.instantiate()

# Build packages that need compilation
Pkg.build()

# Pin package to specific version
Pkg.pin("PackageName", v"1.2.3")
```

### REPL Tips

```julia
# Enter package mode: ]
# Enter help mode: ?
# Enter shell mode: ;
# Back to Julia mode: backspace

# Multi-line input: end with incomplete expression
julia> function foo()
       # Press Enter, continues...
       end

# Clear REPL
julia> Base.Banner.banner()  # Or Ctrl+L in some terminals
```

## Test.jl API Reference

### Basic Test Macros

```julia
# Basic assertion
@test condition

# Test approximate equality
@test a ≈ b
@test isapprox(a, b, atol=1e-10, rtol=0.01)

# Test exception throwing
@test_throws ExceptionType expression
@test_throws DivideError 1 ÷ 0

# Skip test conditionally
@test_skip condition  # Mark as skipped
@test_broken condition  # Known to fail, doesn't affect test suite

# Test with custom error message
@test condition "Custom error message"
```

### Test Organization

```julia
# Group tests
@testset "Description" begin
    @test condition1
    @test condition2
end

# Nested test sets
@testset "Outer" begin
    @testset "Inner 1" begin
        @test true
    end
    @testset "Inner 2" begin
        @test true
    end
end

# Parameterized tests
@testset "Parametric" for x in 1:5
    @test x^2 ≥ x
end
```

### Test Output

```julia
# Verbose output
@testset verbose=true "Tests" begin
    @test true
end

# Custom test set type
@testset CustomTestSet "Tests" begin
    @test true
end
```

## Best Practices

### Code Organization
✅ Follow Julia style guide (snake_case for functions, PascalCase for types)
✅ Keep functions small and focused
✅ Use multiple dispatch appropriately
✅ Avoid type instability in hot paths
✅ Add docstrings to exported functions

### Type System
✅ Use abstract types for interfaces
✅ Parametric types for generic code
✅ Type annotations for documentation, not performance
✅ Let Julia infer types when possible

### Testing
✅ Write tests before implementation (TDD)
✅ Test independence: each test should be self-contained
✅ Use `@testset` to organize related tests
✅ Use `≈` or `isapprox` for floating point comparisons
✅ Maintain ≥85% code coverage
✅ Add Aqua.jl tests for quality assurance

### Performance
✅ Profile before optimizing (`@time`, `@btime`, `@profile`)
✅ Avoid global variables or declare them `const`
✅ Use `@inbounds` only after verification
✅ Prefer immutable structs when possible
✅ Use `StaticArrays.jl` for small fixed-size arrays

### Documentation
✅ Write docstrings in Markdown format
✅ Include examples in docstrings with `jldoctest`
✅ Document function arguments and return values
✅ Use `"""` for multi-line docstrings

## Project Structure

```
MyPackage/
├── Project.toml          # Package metadata and dependencies
├── Manifest.toml         # Exact dependency versions (gitignore for libraries)
├── src/
│   ├── MyPackage.jl     # Main module file
│   └── submodule.jl     # Additional source files
├── test/
│   ├── runtests.jl      # Main test file
│   └── submodule_tests.jl
├── docs/
│   ├── make.jl          # Documentation build script
│   └── src/
│       └── index.md
├── examples/
│   └── example.jl
└── README.md
```

## Coverage Reporting

```julia
# Install Coverage.jl
using Pkg
Pkg.add("Coverage")

# Run tests with coverage
using Pkg
Pkg.test(coverage=true)

# Generate coverage report
using Coverage

# Process coverage files
coverage = process_folder()

# Calculate coverage percentage
covered_lines, total_lines = get_summary(coverage)
percentage = covered_lines / total_lines * 100

# Write LCOV file (for CI tools)
LCOV.writefile("coverage.info", coverage)

# Clean coverage files
clean_folder(".")
```

## Common Test Patterns

### Testing Exceptions

```julia
@testset "Exception Handling" begin
    # Test that function throws specific exception
    @test_throws ArgumentError my_function(-1)

    # Test exception message
    @test_throws "invalid input" my_function(-1)

    # Catch and inspect exception
    ex = try
        my_function(-1)
    catch e
        e
    end
    @test ex isa ArgumentError
    @test occursin("negative", ex.msg)
end
```

### Testing Output

```julia
@testset "Output Testing" begin
    # Capture stdout
    output = @capture_out println("Hello")
    @test output == "Hello\n"

    # Capture stderr
    error_output = @capture_err @warn "Warning message"
    @test occursin("Warning", error_output)
end
```

### Testing Performance

```julia
using BenchmarkTools

@testset "Performance" begin
    # Simple timing
    result = @timed my_function()
    @test result.time < 1.0  # Less than 1 second

    # Detailed benchmarking
    b = @benchmark my_function()
    @test median(b).time < 1e6  # Less than 1ms in nanoseconds
end
```

## Common Issues & Solutions

### Issue: Package not found after adding
**Solution**: Make sure you're in the correct environment with `Pkg.activate(".")`

### Issue: Tests fail with "UndefVarError"
**Solution**: Check that all required modules are imported in test files with `using`

### Issue: Manifest.toml conflicts in Git
**Solution**: For libraries, add `Manifest.toml` to `.gitignore`; for applications, commit it

### Issue: Slow first-time execution
**Solution**: Julia compiles on first run (JIT); subsequent runs are fast. Use PackageCompiler.jl for AOT compilation

### Issue: Type instability warnings
**Solution**: Use `@code_warntype` to identify issues, add type annotations to help inference

---

_For working examples and use cases, see examples.md_
