---
name: moai-lang-cpp
description: C++23 best practices with Google Test 1.15, clang-format 19, and modern C++ standards.
version: 1.0.0
modularized: true
last_updated: 2025-11-22
compliance_score: 70
auto_trigger_keywords:
  - cpp
  - lang
  - testing
category_tier: 1
---

## Quick Reference (30 seconds)

# C++ Systems Programming — Enterprise

**Primary Focus**: C++23 with modern standards, high-performance systems, and template metaprogramming
**Best For**: Systems programming, game engines, financial systems, real-time applications
**Key Libraries**: C++23 STL, Google Test 1.15, Clang 19, CMake 3.28+, Catch2 3.7
**Auto-triggers**: C++, C++20/23, Google Test, CMake, performance-critical code

| Version | Release | Support |
|---------|---------|---------|
| C++23   | 2023-11 | Active  |
| GCC 14  | 2024-11 | ✅      |
| Clang 19| 2024-11 | ✅      |
| Google Test 1.15 | 2024-10 | ✅ |

---

## What It Does

C++23 systems programming with modern features, zero-cost abstractions, and production-grade best practices.

**Key capabilities**:
- ✅ C++23 features (Concepts, Ranges, Coroutines, Modules, std::expected)
- ✅ Template metaprogramming and compile-time computation
- ✅ Advanced memory management (RAII, smart pointers, move semantics)
- ✅ Zero-cost abstractions with CRTP, concepts, and static polymorphism
- ✅ High-performance optimization strategies
- ✅ Multi-threaded programming patterns
- ✅ Test-driven development with Google Test and Catch2

---

## When to Use

**Automatic triggers**:
- C++ code files (*.cpp, *.h, *.hpp)
- Performance-critical applications
- System-level programming
- Real-time systems development

**Manual invocation**:
- Design high-performance architecture
- Optimize bottlenecks
- Review code for best practices
- Troubleshoot memory or threading issues

---

## Three-Level Learning Path

### Level 1: Fundamentals (See examples.md)

Core C++23 concepts with practical patterns:
- **Modern C++ Features**: Structured bindings, optional, variant, ranges, coroutines
- **Smart Pointers**: unique_ptr, shared_ptr for memory safety
- **RAII Pattern**: Resource management without garbage collection
- **Basic Testing**: Google Test fixtures and parameterized tests
- **Build System**: CMake configuration and FetchContent dependencies

### Level 2: Advanced Patterns (See modules/advanced-patterns.md)

Production-ready enterprise patterns:
- **Template Metaprogramming**: Compile-time computation, type traits, concepts
- **Advanced OOP**: CRTP, Pimpl, Strategy, Visitor, Observer patterns
- **Concurrency**: Thread-safe singletons, producer-consumer, double-checked locking
- **Policy-Based Design**: Flexible compile-time configuration
- **Recursive Lambdas**: C++23 deducing this feature
- **Zero-Cost Abstractions**: Static polymorphism without virtual overhead

### Level 3: Performance (See modules/optimization.md)

Production deployment and optimization:
- **Cache Optimization**: Array of Structures vs. Structure of Arrays, cache alignment
- **Compiler Hints**: Branch prediction, inlining, SIMD vectorization
- **Profiling**: Perf, valgrind, Google Benchmark, Address Sanitizer
- **Parallel Algorithms**: C++17 parallel STL execution policies
- **Memory Pooling**: Custom allocators, object pools, lazy loading
- **Link-Time Optimization**: LTO, Profile-Guided Optimization (PGO)

---

## Best Practices

✅ **DO**:
- Use modern C++ features (concepts, ranges, coroutines)
- Apply RAII for all resource management
- Prefer smart pointers over raw pointers
- Write move semantics for large objects
- Maintain const-correctness throughout
- Test coverage ≥85% minimum
- Use CMake for build configuration
- Profile before optimizing

❌ **DON'T**:
- Use raw new/delete without RAII wrapper
- Ignore compiler warnings
- Mix C and C++ idioms in the same codebase
- Use `using namespace std` in headers
- Implement manual memory management
- Skip const correctness
- Optimize without profiling data
- Mix threading primitives inconsistently

---

## Tool Versions (2025-11-22)

| Tool | Version | Purpose |
|------|---------|---------|
| **C++** | 23 | Standard |
| **GCC** | 14.2.0 | Compiler |
| **Clang** | 19.1.7 | Compiler |
| **Google Test** | 1.15.0 | Testing |
| **clang-format** | 19.1.7 | Code formatting |
| **CMake** | 3.30+ | Build system |
| **Catch2** | 3.7.0 | BDD testing |

---

## Installation & Setup

```bash
# Ubuntu/Debian
sudo apt-get install g++-14 clang-19 cmake google-test

# macOS with Homebrew
brew install gcc cmake catch2

# CMake with Google Test
cmake_minimum_required(VERSION 3.28)
project(MyApp CXX)
set(CMAKE_CXX_STANDARD 23)

# Project dependencies via FetchContent
include(FetchContent)
FetchContent_Declare(googletest
  GIT_REPOSITORY https://github.com/google/googletest.git
  GIT_TAG v1.15.0
)
FetchContent_MakeAvailable(googletest)
```

---

## Works Well With

- `moai-essentials-debug` (debugging complex issues)
- `moai-essentials-perf` (performance profiling)
- `moai-domain-backend` (backend architecture)
- `moai-foundation-trust` (TRUST 5 quality gates)

---

## Learn More

- **Practical Examples**: See `examples.md` for 10 real-world patterns
- **Advanced Patterns**: See `modules/advanced-patterns.md` for TMP, CRTP, concurrency
- **Performance Tuning**: See `modules/optimization.md` for cache optimization, profiling
- **Official Documentation**: https://en.cppreference.com/w/cpp/
- **Google Test**: https://google.github.io/googletest/
- **CMake**: https://cmake.org/documentation/

---

## Changelog

- **v3.1.0** (2025-11-22): Modularized structure with advanced patterns and optimization modules
- **v3.0.0** (2025-11-22): Comprehensive C++23 features, STL deep dive, performance optimization
- **v2.0.0** (2025-10-22): Major update with latest tool versions, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial release

---

## Context7 Integration

### Related Libraries & Tools
- [C++](/cplusplus): Modern C++ standards and references
- [Boost](/boostorg/boost): Comprehensive C++ libraries
- [Google Test](/google/googletest): Testing framework
- [CMake](/Kitware/CMake): Build system generator

### Official Documentation
- [C++ Reference](https://en.cppreference.com/)
- [C++23 Features](https://en.cppreference.com/w/cpp/23)
- [Google Test Guide](https://google.github.io/googletest/)
- [CMake Documentation](https://cmake.org/documentation/)

---

**Skills**: Skill("moai-essentials-debug"), Skill("moai-essentials-perf"), Skill("moai-domain-backend")
**Auto-loads**: C++ projects with Google Test, CMake, modern C++ features