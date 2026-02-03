---
paths:
  - "**/*.cpp"
  - "**/*.hpp"
  - "**/*.h"
  - "**/*.cc"
  - "**/CMakeLists.txt"
---

# C++ Rules

Version: C++23 / C++20

## Tooling

- Build: CMake 3.28+
- Linting: clang-tidy
- Formatting: clang-format
- Testing: GoogleTest or Catch2
- Package management: vcpkg or Conan

## MUST

- Use smart pointers (unique_ptr, shared_ptr)
- Use RAII for resource management
- Use std::span for array views
- Use std::optional for nullable values
- Enable compiler warnings (-Wall -Wextra -Werror)
- Document public APIs with Doxygen comments

## MUST NOT

- Use raw new/delete (use smart pointers)
- Use C-style casts (use static_cast, etc.)
- Ignore compiler warnings
- Use macros when constexpr works
- Leave uninitialized variables
- Use using namespace in headers

## File Conventions

- *_test.cpp for test files
- .hpp for C++ headers, .h for C headers
- Use PascalCase for classes and types
- Use snake_case for functions and variables
- Use SCREAMING_CASE for macros

## Testing

- Use GoogleTest or Catch2
- Use GMock for mocking
- Use sanitizers (ASan, UBSan) in CI
- Use benchmark tests for performance

## MoAI Integration

- Use Skill("moai-lang-cpp") for detailed patterns
- Follow TRUST 5 quality gates
