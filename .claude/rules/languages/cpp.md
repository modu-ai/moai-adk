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

- Build: CMake 3.28+
- Linting: clang-tidy
- Formatting: clang-format
- Testing: Google Test or Catch2

## Best Practices (2026)

- Use smart pointers over raw pointers
- Use concepts for template constraints
- Use ranges for collection operations
- Apply RAII for resource management
- Use modules where supported (C++20)

## CMake Configuration

```cmake
cmake_minimum_required(VERSION 3.28)
project(MyProject LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Use cxx_std_23 for target
target_compile_features(mytarget PUBLIC cxx_std_23)

# Enable Unity builds for faster compilation
set(CMAKE_UNITY_BUILD ON)
```

## C++23 Features

```cpp
// std::expected for error handling
std::expected<int, Error> parse(std::string_view sv);

// std::print for formatted output
std::print("Hello, {}!\n", name);

// Deducing this
struct Widget {
    template<typename Self>
    auto&& value(this Self&& self) { return self.m_value; }
};
```

## Performance

- Use Unity builds to reduce compilation time
- Use precompiled headers (PCH) for large projects
- Profile with Valgrind/AddressSanitizer
- Use `-O2` or `-O3` for release builds

## MoAI Integration

- Use Skill("moai-lang-cpp") for detailed patterns
- Follow TRUST 5 quality gates
