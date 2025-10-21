---

name: moai-lang-cpp
description: C++ best practices with Google Test, clang-format, and modern C++ (C++17/20). Use when writing or reviewing C++ code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# C++ Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | C++ code discussions, framework guidance, or file extensions such as .cpp/.hpp. |
| Tier | 3 |

## What it does

Provides C++-specific expertise for TDD development, including Google Test framework, clang-format formatting, and modern C++ (C++17/20) features.

## When to use

- Engages when the conversation references C++ work, frameworks, or files like .cpp/.hpp.
- "Writing C++ tests", "How to use Google Test", "Modern C++"
- Automatically invoked when working with C++ projects
- C++ SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **Google Test (gtest)**: Unit testing framework
- **Google Mock (gmock)**: Mocking framework
- **Catch2**: Alternative testing framework
- Test coverage with gcov/lcov

**Build Tools**:
- **CMake**: Cross-platform build system
- **Make**: Traditional build tool
- **Conan/vcpkg**: Package managers

**Code Quality**:
- **clang-format**: Code formatting
- **clang-tidy**: Static analysis
- **cppcheck**: Additional static analysis

**Modern C++ Features**:
- **Smart pointers**: unique_ptr, shared_ptr, weak_ptr
- **Move semantics**: std::move, rvalue references
- **Lambda expressions**: Inline functions
- **auto keyword**: Type inference
- **constexpr**: Compile-time evaluation
- **std::optional**: Nullable types (C++17)
- **Concepts**: Type constraints (C++20)

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- RAII (Resource Acquisition Is Initialization)
- Rule of Five (destructor, copy/move constructors/assignments)
- Prefer stack allocation over heap
- Const correctness

## Examples
```bash
cmake --build build --target test && clang-tidy src/*.cpp
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
- ISO. "ISO/IEC 14882:2020(E) Programming Language C++." (accessed 2025-03-29).
- LLVM Project. "clang-tidy Documentation." https://clang.llvm.org/extra/clang-tidy/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C++-specific review)
- alfred-performance-optimizer (C++ profiling)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
