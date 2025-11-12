# Compiled Languages Validation Patterns

> Referenced from: `moai-validation-quality/SKILL.md`

This document covers validation patterns for compiled languages: **C, C++, Rust, Go, C#**.

---

## C / C++

### Technology Stack
- **Compiler**: GCC 13.x, Clang 18.x
- **Build System**: CMake 3.27+, Make
- **Package Manager**: Conan 2.x, vcpkg
- **Testing**: Google Test (GTest), Catch2, CppUnit

### Validation Tools

#### Compilation (with strict warnings)
```bash
# GCC - Treat warnings as errors
gcc -Wall -Wextra -Werror -pedantic src/*.c -o program

# Clang - With additional checks
clang -Wall -Wextra -Werror -fsanitize=address src/*.c -o program

# C++
g++ -Wall -Wextra -Werror -std=c++17 src/*.cpp -o program
clang++ -Wall -Wextra -Werror -std=c++20 src/*.cpp -o program
```

#### Static Analysis (CPPCheck)
```bash
# Comprehensive static analysis
cppcheck --enable=all --error-exitcode=1 src/

# With specific checks
cppcheck --enable=all --suppress=missingIncludeSystem src/ --error-exitcode=1
```

#### Memory Analysis (Valgrind - optional)
```bash
# Memory leak detection
valgrind --leak-check=full --error-exitcode=1 ./program arg1 arg2

# With verbose output
valgrind --leak-check=full --show-leak-kinds=all --track-origins=yes ./program
```

#### Code Formatting (clang-format)
```bash
# Check formatting
clang-format --dry-run -Werror src/*.{c,cpp,h}

# With custom config
clang-format -i src/*.{c,cpp,h}  # Auto-format
clang-format --dry-run src/*.{c,cpp,h} || exit 1
```

#### Testing
```bash
# GTest (Google Test)
g++ -o test test.cpp src/*.cpp -lgtest -lgtest_main
./test

# Catch2
catch_tests --success
```

### Fallback Chain
```yaml
c:
  compiler: [gcc, clang]
  static_analysis: [cppcheck, splint]
  memory_check: [valgrind, address-sanitizer]
  formatter: [clang-format]

cpp:
  compiler: [g++, clang++]
  static_analysis: [cppcheck, clang-tidy, cpplint]
  memory_check: [valgrind, thread-sanitizer]
  formatter: [clang-format]
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/cpp-2025",
    topic="C++ cmake conan gtest memory safety sanitizers patterns"
)
```

---

## Rust

### Technology Stack
- **Language**: Rust 1.91.1 (November 2025)
- **Async Runtime**: Tokio 1.48+
- **Testing**: Rust built-in test framework + criterion
- **Linting**: Clippy (built-in)
- **Formatting**: Rustfmt (built-in)

### Validation Tools

#### Testing + Coverage
```bash
# Run all tests
cargo test

# Run with output
cargo test -- --nocapture

# Coverage (with tarpaulin)
cargo tarpaulin --out Html --output-dir coverage

# With exclusions
cargo tarpaulin --exclude-files tests/* --out Xml
```

#### Linting (Clippy)
```bash
# Clippy checks (warnings as errors)
cargo clippy -- -D warnings

# Specific lint groups
cargo clippy -- -D clippy::all -D clippy::pedantic
```

#### Formatting (Rustfmt)
```bash
# Check format
cargo fmt --check

# Auto-format
cargo fmt

# Check with error exit
cargo fmt --check || exit 1
```

#### Security Scanning (Audit)
```bash
# Check for security vulnerabilities
cargo audit || true  # Non-blocking

# With strict mode
cargo audit --deny warnings
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🦀 Rust Validation Starting..."

echo "▶ Testing"
cargo test

echo "▶ Linting"
cargo clippy -- -D warnings

echo "▶ Formatting"
cargo fmt --check

echo "▶ Security Audit"
cargo audit

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
rust:
  test: [cargo-test]
  linter: [cargo-clippy]
  formatter: [rustfmt]
  security: [cargo-audit]
  coverage: [tarpaulin, llvm-cov]
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/rust-2025",
    topic="Rust cargo clippy rustfmt tokio testing patterns safety"
)
```

---

## Go

### Technology Stack
- **Language**: Go 1.25.4 (November 2025)
- **Package Manager**: go modules (go.mod)
- **Testing**: `testing` package (built-in)
- **Linting**: staticcheck, golint
- **Formatting**: gofmt (built-in)

### Validation Tools

#### Testing + Coverage
```bash
# Run tests
go test ./...

# With coverage
go test ./... -cover

# Coverage profile
go test ./... -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out
go tool cover -func=coverage.out | grep total
```

#### Static Analysis (staticcheck)
```bash
# Install staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# Run staticcheck
staticcheck ./...

# Or use go vet (built-in)
go vet ./...
```

#### Linting (golint)
```bash
# Install golint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
golangci-lint run ./...

# With config
golangci-lint run --config .golangci.yml
```

#### Formatting (gofmt)
```bash
# Check formatting
gofmt -l . | grep . && exit 1 || exit 0

# Auto-format
gofmt -w .

# Using goimports (includes import organization)
goimports -w .
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🐹 Go Validation Starting..."

echo "▶ Testing"
go test ./... -cover

echo "▶ Static Analysis"
go vet ./...
staticcheck ./... || true

echo "▶ Formatting"
gofmt -l . | grep . && exit 1 || echo "Format OK"

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
go:
  test: [go-test]
  static_analysis: [staticcheck, go-vet, golint]
  linter: [golangci-lint, gometalinter]
  formatter: [gofmt, goimports]
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/go-2025",
    topic="Go testing modules staticcheck gofmt coverage patterns"
)
```

---

## C#

### Technology Stack
- **Language**: C# 12.0+
- **Runtime**: .NET 9.0 LTS (November 2025)
- **Build System**: .NET CLI (dotnet)
- **Testing**: xUnit, NUnit, MSTest
- **Code Analysis**: Roslyn Analyzers

### Validation Tools

#### Build + Compilation
```bash
# Restore dependencies
dotnet restore

# Build with strict warnings
dotnet build /p:TreatWarningsAsErrors=true

# Release build
dotnet build --configuration Release
```

#### Testing
```bash
# Run tests
dotnet test

# With coverage
dotnet test --collect:"XPlat Code Coverage"

# Coverage report
dotnet test /p:CollectCoverage=true /p:CoverageFormat=cobertura
```

#### Code Analysis (Roslyn)
```bash
# Enable static analysis
dotnet build /p:EnableNETAnalyzers=true /p:TreatWarningsAsErrors=true

# Style enforcement
dotnet build /p:EnforceCodeStyleInBuild=true

# Specific rules
dotnet format --verify-no-changes --verbosity diagnostic
```

#### Formatting (dotnet format)
```bash
# Check formatting
dotnet format --verify-no-changes

# Auto-format
dotnet format

# With specific settings
dotnet format --include src/**/*.cs
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🔷 C# Validation Starting..."

echo "▶ Build"
dotnet build /p:TreatWarningsAsErrors=true

echo "▶ Testing"
dotnet test /p:CollectCoverage=true

echo "▶ Code Analysis"
dotnet format --verify-no-changes

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
csharp:
  compiler: [dotnet]
  test: [xunit, nunit, mstest]
  analyzer: [roslyn-analyzers]
  formatter: [dotnet-format]
  style_check: [editorconfig]
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/csharp-2025",
    topic="C# .NET testing xunit roslyn static analysis patterns"
)
```

---

## Installation Commands

### C/C++ Tools
```bash
# macOS
brew install gcc clang cmake cppcheck clang-format valgrind

# Linux (Ubuntu)
sudo apt-get install build-essential clang cppcheck clang-format valgrind

# Google Test
git clone https://github.com/google/googletest.git
cd googletest && mkdir -p build && cd build && cmake .. && make
```

### Rust Tools
```bash
# Install Rust (includes cargo, rustfmt, clippy)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Update
rustup update

# Tarpaulin (coverage)
cargo install cargo-tarpaulin
```

### Go Tools
```bash
# Go already includes: go test, gofmt, go vet
# Install additional tools:
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### C# Tools
```bash
# .NET SDK (includes all C# tools)
# macOS
brew install dotnet

# Linux
sudo apt-get install dotnet-sdk-9.0

# Windows
# Download from https://dotnet.microsoft.com/download
```

---

**Last Updated**: 2025-11-12
**Related**: [SKILL.md](SKILL.md), [TOOL-REFERENCE.md](TOOL-REFERENCE.md)
