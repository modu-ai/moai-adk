---
name: moai-lang-cpp
description: C++ best practices with Google Test, clang-format, and modern C++ (C++17/20)
allowed-tools:
- Read
- Bash
---

# C++ Expert

## What it does

Provides C++-specific expertise for TDD development, including Google Test framework, clang-format formatting, and modern C++ (C++17/20) features.

## When to use

- "C++ 테스트 작성", "Google Test 사용법", "모던 C++"
- Automatically invoked when working with C++ projects
- C++ SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with Google Test
User: "/alfred:2-build CACHE-001"
Claude: (creates RED test with gtest, GREEN implementation with smart pointers, REFACTOR)

### Example 2: Modern C++ refactoring
User: "C++17 기능으로 리팩토링"
Claude: (refactors code to use std::optional, structured bindings)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C++-specific review)
- alfred-performance-optimizer (C++ profiling)
