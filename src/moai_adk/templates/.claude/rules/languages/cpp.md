---
paths:
  - "**/*.cpp"
  - "**/*.hpp"
  - "**/*.h"
  - "**/*.cc"
  - "**/CMakeLists.txt"
---

# C++ Rules

Version: C++23/C++20

## Tooling

- Build: CMake 3.20+
- Linting: clang-tidy
- Formatting: clang-format
- Testing: Google Test or Catch2

## Preferred Patterns

- Use smart pointers over raw pointers
- Use concepts and ranges
- Apply RAII for resource management

## MoAI Integration

- Use Skill("moai-lang-cpp") for detailed patterns
- Follow TRUST 5 quality gates
