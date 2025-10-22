# moai-lang-c - CLI Reference

_Last updated: 2025-10-22_

## Quick Reference

### Installation

```bash
# GCC/Clang (via package manager)
# Debian/Ubuntu
sudo apt-get install gcc clang cmake cppcheck

# macOS
brew install gcc llvm cmake cppcheck

# Verify installation
gcc --version        # Should show 14.2.0+
clang --version      # Should show 19.1.7+
cmake --version      # Should show 3.31.0+
cppcheck --version   # Should show 2.16.0+
```

### Common Commands

```bash
# Compile (GCC)
gcc -std=c17 -Wall -Wextra -pedantic -o program main.c

# Compile (Clang)
clang -std=c17 -Wall -Wextra -pedantic -o program main.c

# Static Analysis (cppcheck)
cppcheck --enable=all --suppress=missingIncludeSystem src/

# Build with CMake
mkdir build && cd build
cmake ..
cmake --build .

# Run Unity tests
./build/tests/test_runner
```

## Tool Versions (2025-10-22)

- **GCC**: 14.2.0 - Primary compiler (GNU Compiler Collection)
- **Clang**: 19.1.7 - Alternative compiler with better diagnostics
- **cppcheck**: 2.16.0 - Static analysis tool for C/C++
- **CMake**: 3.31.0 - Cross-platform build system generator

## Official Documentation Links

- **C17 Standard**: ISO/IEC 9899:2018 (draft N2176 available free)
- **C23 Standard**: ISO/IEC 9899:2024 (draft N3299 available free)
- **cppreference**: https://en.cppreference.com/w/c/17.html
- **Unity Test Framework**: https://www.throwtheswitch.org/unity
- **Cppcheck**: https://cppcheck.sourceforge.io/
- **CMake**: https://cmake.org/cmake/help/latest/

## C Standards Overview

### C17 (ISO/IEC 9899:2018)
- Bug-fix release for C11
- No new language features
- Published July 5, 2018

### C23 (ISO/IEC 9899:2024)
- New keywords: `bool`, `true`, `false`, `nullptr`, `static_assert`, `thread_local`
- Bit utility headers: `<stdbit.h>` with `stdc_count_ones`
- Checked integer arithmetic: `<stdckdint.h>`
- Unified `[[attribute]]` syntax: `[[likely]]`, `[[unlikely]]`, `[[deprecated]]`
- Published October 2024

## Unity Test Framework

### Core Components
- **unity.h** - Main header
- **unity.c** - Test runner implementation
- Portable: works from 8-bit microcontrollers to 64-bit systems

### Test Structure
```c
#include "unity.h"
#include "module_to_test.h"

void setUp(void) {
    // Run before each test
}

void tearDown(void) {
    // Run after each test
}

void test_example(void) {
    TEST_ASSERT_EQUAL(expected, actual);
}

int main(void) {
    UNITY_BEGIN();
    RUN_TEST(test_example);
    return UNITY_END();
}
```

### Common Assertions
- `TEST_ASSERT_EQUAL(expected, actual)`
- `TEST_ASSERT_TRUE(condition)`
- `TEST_ASSERT_NULL(pointer)`
- `TEST_ASSERT_EQUAL_STRING(expected, actual)`

## Cppcheck Configuration

### Basic Usage
```bash
# Analyze all C files
cppcheck --enable=all --language=c src/

# Suppress false positives
cppcheck --enable=all --inline-suppr src/

# Check against MISRA C 2023
cppcheck --addon=misra.json src/
```

### Inline Suppression
```c
// cppcheck-suppress warningId
int risky_code() {
    // Code that triggers false positive
}
```

### Supported Standards
- MISRA C 2023
- CERT C
- AUTOSAR C++14
- Top 25 CWE

## CMake Best Practices

### Minimum CMakeLists.txt
```cmake
cmake_minimum_required(VERSION 3.31)
project(MyProject C)

set(CMAKE_C_STANDARD 17)
set(CMAKE_C_STANDARD_REQUIRED ON)

add_executable(myapp main.c)

# Add Unity tests
enable_testing()
add_subdirectory(tests)
```

### Build Commands
```bash
# Configure (Unix Makefiles)
cmake -S . -B build

# Build
cmake --build build

# Run tests
cd build && ctest
```

---

_For detailed usage and best practices, see SKILL.md_
