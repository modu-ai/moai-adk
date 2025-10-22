# moai-lang-c - Working Examples

_Last updated: 2025-10-22_

## Example 1: Basic C Project Setup with Unity and CMake

### Project Structure
```
my-c-project/
├── CMakeLists.txt
├── src/
│   ├── calculator.c
│   └── calculator.h
└── tests/
    ├── CMakeLists.txt
    ├── unity/           # Unity framework files
    │   ├── unity.c
    │   └── unity.h
    └── test_calculator.c
```

### calculator.h
```c
#ifndef CALCULATOR_H
#define CALCULATOR_H

int add(int a, int b);
int subtract(int a, int b);

#endif
```

### calculator.c
```c
#include "calculator.h"

int add(int a, int b) {
    return a + b;
}

int subtract(int a, int b) {
    return a - b;
}
```

### tests/test_calculator.c
```c
#include "unity/unity.h"
#include "../src/calculator.h"

void setUp(void) {
    // This runs before each test
}

void tearDown(void) {
    // This runs after each test
}

void test_add_positive_numbers(void) {
    TEST_ASSERT_EQUAL(5, add(2, 3));
}

void test_add_negative_numbers(void) {
    TEST_ASSERT_EQUAL(-5, add(-2, -3));
}

void test_subtract(void) {
    TEST_ASSERT_EQUAL(1, subtract(3, 2));
}

int main(void) {
    UNITY_BEGIN();
    RUN_TEST(test_add_positive_numbers);
    RUN_TEST(test_add_negative_numbers);
    RUN_TEST(test_subtract);
    return UNITY_END();
}
```

### Root CMakeLists.txt
```cmake
cmake_minimum_required(VERSION 3.31)
project(Calculator C)

set(CMAKE_C_STANDARD 17)
set(CMAKE_C_STANDARD_REQUIRED ON)

# Main executable
add_executable(calculator src/calculator.c)

# Enable testing
enable_testing()
add_subdirectory(tests)
```

### tests/CMakeLists.txt
```cmake
# Unity test framework
add_library(unity unity/unity.c)

# Test executable
add_executable(test_calculator
    test_calculator.c
    ../src/calculator.c
)

target_link_libraries(test_calculator unity)
target_include_directories(test_calculator PRIVATE unity)

add_test(NAME CalculatorTests COMMAND test_calculator)
```

### Build and Run
```bash
# Setup build directory
mkdir build && cd build

# Configure
cmake ..

# Build
cmake --build .

# Run tests
ctest --output-on-failure

# Or run directly
./tests/test_calculator
```

## Example 2: TDD Workflow (RED-GREEN-REFACTOR)

### RED Phase: Write Failing Test First
```c
// tests/test_string_utils.c
#include "unity/unity.h"
#include "../src/string_utils.h"

void test_string_reverse(void) {
    char input[] = "hello";
    char expected[] = "olleh";

    string_reverse(input);

    TEST_ASSERT_EQUAL_STRING(expected, input);
}

int main(void) {
    UNITY_BEGIN();
    RUN_TEST(test_string_reverse);
    return UNITY_END();
}
```

```bash
# Run test - should FAIL
cmake --build build
./build/tests/test_string_utils

# Output: test_string_reverse:FAIL (function doesn't exist yet)
```

### GREEN Phase: Implement Minimum Code to Pass
```c
// src/string_utils.h
#ifndef STRING_UTILS_H
#define STRING_UTILS_H

void string_reverse(char* str);

#endif
```

```c
// src/string_utils.c
#include "string_utils.h"
#include <string.h>

void string_reverse(char* str) {
    int len = strlen(str);
    for (int i = 0; i < len / 2; i++) {
        char temp = str[i];
        str[i] = str[len - 1 - i];
        str[len - 1 - i] = temp;
    }
}
```

```bash
# Run test again - should PASS
cmake --build build
./build/tests/test_string_utils

# Output: test_string_reverse:PASS
```

### REFACTOR Phase: Improve Code Quality
```c
// src/string_utils.c (refactored)
#include "string_utils.h"
#include <string.h>

// @CODE:STRUTIL-001 | SPEC: SPEC-STRUTIL-001.md | TEST: tests/test_string_utils.c
static void swap_chars(char* a, char* b) {
    char temp = *a;
    *a = *b;
    *b = temp;
}

void string_reverse(char* str) {
    if (str == NULL) return;  // Guard clause

    size_t len = strlen(str);
    for (size_t i = 0; i < len / 2; i++) {
        swap_chars(&str[i], &str[len - 1 - i]);
    }
}
```

```bash
# Run test after refactor - should still PASS
./build/tests/test_string_utils

# Run static analysis
cppcheck --enable=all --suppress=missingIncludeSystem src/string_utils.c
```

## Example 3: Quality Gate with Coverage and Static Analysis

### Setup Project with Quality Checks
```bash
# Install gcov for coverage
sudo apt-get install gcov lcov

# Configure CMake with coverage flags
cmake -DCMAKE_C_FLAGS="--coverage" -B build

# Build
cmake --build build

# Run tests
cd build && ctest

# Generate coverage report
lcov --capture --directory . --output-file coverage.info
lcov --remove coverage.info '/usr/*' --output-file coverage.info
genhtml coverage.info --output-directory coverage_report

# View coverage (should be ≥85%)
firefox coverage_report/index.html
```

### Static Analysis with Cppcheck
```bash
# Run cppcheck on entire project
cppcheck --enable=all \
         --suppress=missingIncludeSystem \
         --language=c \
         --std=c17 \
         --error-exitcode=1 \
         src/

# Expected output: No errors found
```

### Full Quality Gate Script
```bash
#!/bin/bash
# quality-gate.sh

set -e

echo "=== Building project ==="
cmake --build build

echo "=== Running tests ==="
cd build && ctest --output-on-failure

echo "=== Checking coverage ==="
lcov --capture --directory . --output-file coverage.info
COVERAGE=$(lcov --summary coverage.info | grep lines | awk '{print $2}' | cut -d'%' -f1)

if (( $(echo "$COVERAGE < 85" | bc -l) )); then
    echo "❌ Coverage $COVERAGE% is below 85%"
    exit 1
fi

echo "✅ Coverage $COVERAGE% passed"

echo "=== Running static analysis ==="
cd ..
cppcheck --enable=all --suppress=missingIncludeSystem --error-exitcode=1 src/

echo "✅ All quality gates passed!"
```

## Example 4: C23 Modern Features

### Using New C23 Keywords
```c
// src/modern_c23.c
#include <stdio.h>
#include <stdbit.h>    // C23: Bit utilities
#include <stdckdint.h> // C23: Checked arithmetic

// C23: nullptr instead of NULL
void process_data(int* data) {
    if (data == nullptr) {
        return;
    }
    printf("Processing: %d\n", *data);
}

// C23: bool, true, false built-in
bool is_power_of_two(unsigned int n) {
    if (n == 0) return false;
    return stdc_count_ones(n) == 1;  // C23: stdbit.h
}

// C23: Checked integer addition
bool safe_add(int a, int b, int* result) {
    return !ckd_add(result, a, b);  // Returns true if no overflow
}

int main(void) {
    // C23: nullptr
    int* ptr = nullptr;
    process_data(ptr);

    // C23: bit utilities
    printf("Is 16 power of 2? %s\n", is_power_of_two(16) ? "true" : "false");

    // C23: checked arithmetic
    int result;
    if (safe_add(2147483647, 1, &result)) {
        printf("Addition succeeded: %d\n", result);
    } else {
        printf("Addition would overflow\n");
    }

    return 0;
}
```

### Compile with C23
```bash
# GCC with C23 support
gcc -std=c23 -o modern modern_c23.c

# Clang with C23 support
clang -std=c23 -o modern modern_c23.c
```

---

_For more details on best practices and TRUST principles, see SKILL.md_
