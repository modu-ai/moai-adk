---
name: moai-lang-c
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: C17/C23 best practices with Unity 2.6.0, cppcheck 2.16.0, CMake 3.31.0, and modern tooling. Use when writing or reviewing C code in project workflows.
keywords: [c, c17, c23, unity, cppcheck, cmake, make, gcc, clang, valgrind, memory-safety]
allowed-tools:
  - Read
  - Bash
---

# C17/C23 Expert Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-c |
| **Version** | 2.0.0 (2025-10-22) |
| **C Support** | C23 (latest), C17 (stable), C11 (maintenance) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when language keywords detected |
| **Trigger cues** | `.c`, `.h` files, C frameworks, TDD discussions, memory safety |
| **Tier** | Language / 23 (comprehensive coverage) |

---

## What It Does

Provides **C17/C23 expertise** for modern systems programming with:

- ✅ **Testing Framework**: Unity 2.6.0 (embedded-friendly unit testing)
- ✅ **Static Analysis**: cppcheck 2.16.0 (defect detection)
- ✅ **Build Systems**: CMake 3.31.0, Make 4.4 (cross-platform builds)
- ✅ **Memory Safety**: Valgrind 3.24.0, AddressSanitizer (leak detection)
- ✅ **C23 Features**: `nullptr`, `typeof`, `constexpr`, `[[attributes]]`
- ✅ **Compiler Support**: GCC 14.2.0, Clang 19.1.7 (modern standards)
- ✅ **Code Quality**: clang-format 19.1.7 (consistent style)

---

## When to Use

**Automatic triggers**:
- C code discussions, `.c`/`.h` files, systems programming
- "Writing C tests", "Memory safety", "C build systems"
- C SPEC implementation (`/alfred:2-run`)
- Embedded/IoT/kernel development requests

**Manual invocation**:
- Review C code for TRUST 5 compliance
- Design systems-level components (drivers, libraries)
- Upgrade from C11/C17 to C23
- Refactor code for memory safety (ASAN, Valgrind)

---

## How It Works (Best Practices)

### 1. TDD Framework (Unity 2.6.0)

**Why Unity?**: Lightweight, no dependencies, perfect for embedded/kernel development

```c
// tests/test_calculator.c
#include "unity.h"
#include "../src/calculator.h"

void setUp(void) {
    // Run before each test
}

void tearDown(void) {
    // Run after each test
}

void test_add_positive_numbers(void) {
    TEST_ASSERT_EQUAL_INT(5, add(2, 3));
}

void test_add_negative_numbers(void) {
    TEST_ASSERT_EQUAL_INT(-5, add(-2, -3));
}

void test_add_overflow_detection(void) {
    TEST_ASSERT_EQUAL_INT(INT_MAX, add(INT_MAX, 1));  // Should saturate
}

int main(void) {
    UNITY_BEGIN();
    RUN_TEST(test_add_positive_numbers);
    RUN_TEST(test_add_negative_numbers);
    RUN_TEST(test_add_overflow_detection);
    return UNITY_END();
}
```

**Key Points**:
- ✅ Use Unity 2.6.0+ (not CUnit/Check)
- ✅ One assertion per test (clarity)
- ✅ `setUp()`/`tearDown()` for cleanup
- ✅ `TEST_ASSERT_*` macros for validation
- ✅ Coverage ≥85% enforced by quality gate
- ✅ Test isolation (no shared state)

**CLI Commands**:
```bash
# Build and run tests
gcc -o tests/test_runner tests/*.c src/*.c -Ithird_party/unity -DUNITY_INCLUDE_DOUBLE
./tests/test_runner                 # Run all tests

# With Valgrind (memory leak detection)
valgrind --leak-check=full ./tests/test_runner

# With AddressSanitizer (heap corruption)
gcc -fsanitize=address -g -o tests/test_runner tests/*.c src/*.c
./tests/test_runner
```

### 2. Static Analysis (cppcheck 2.16.0)

**cppcheck**: Fast, zero-config static analyzer for C/C++

```bash
# Run cppcheck
cppcheck --enable=all --suppress=missingIncludeSystem src/

# Check specific issues
cppcheck --enable=warning,performance,portability src/

# Generate XML report
cppcheck --enable=all --xml --xml-version=2 src/ 2> report.xml
```

**Configuration** (`.cppcheck`):
```ini
# Enable all checks except missing includes
--enable=all
--suppress=missingIncludeSystem
--inline-suppr
--std=c17
--platform=unix64
```

**Common Issues Detected**:
- Memory leaks (`malloc` without `free`)
- Null pointer dereferences
- Buffer overflows (array bounds)
- Uninitialized variables
- Resource leaks (file handles, sockets)

### 3. Build Systems (CMake 3.31.0)

**Modern CMake Configuration**:

```cmake
# CMakeLists.txt
cmake_minimum_required(VERSION 3.31)
project(my_c_project VERSION 2.0.0 LANGUAGES C)

# Set C23 standard
set(CMAKE_C_STANDARD 23)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_EXTENSIONS OFF)

# Enable strict warnings
add_compile_options(-Wall -Wextra -Wpedantic -Werror)

# Library target
add_library(calculator src/calculator.c)
target_include_directories(calculator PUBLIC include)

# Executable target
add_executable(main src/main.c)
target_link_libraries(main calculator)

# Test target
add_executable(test_runner tests/test_calculator.c)
target_link_libraries(test_runner calculator unity)

# Enable testing
enable_testing()
add_test(NAME CalculatorTests COMMAND test_runner)

# Optional: AddressSanitizer
option(ENABLE_ASAN "Enable AddressSanitizer" OFF)
if(ENABLE_ASAN)
    add_compile_options(-fsanitize=address -fno-omit-frame-pointer)
    add_link_options(-fsanitize=address)
endif()
```

**CLI Commands**:
```bash
# Configure build
cmake -B build -DCMAKE_BUILD_TYPE=Release

# Build project
cmake --build build

# Run tests
ctest --test-dir build --output-on-failure

# Enable ASAN
cmake -B build -DENABLE_ASAN=ON
cmake --build build
```

### 4. Memory Safety (Valgrind 3.24.0 + ASAN)

**Valgrind** (leak detection):
```bash
# Detect memory leaks
valgrind --leak-check=full --show-leak-kinds=all ./program

# Track uninitialized reads
valgrind --track-origins=yes ./program

# Generate suppression file (for known issues)
valgrind --gen-suppressions=all ./program 2> suppressions.txt
```

**AddressSanitizer** (heap corruption):
```bash
# Compile with ASAN
gcc -fsanitize=address -g -O1 -fno-omit-frame-pointer src/*.c -o program

# Run with ASAN options
ASAN_OPTIONS=detect_leaks=1:halt_on_error=0 ./program
```

**Common Memory Issues**:
- ❌ Leaks: `malloc` without `free`
- ❌ Double-free: calling `free` twice
- ❌ Use-after-free: accessing freed memory
- ❌ Buffer overflow: writing past array bounds
- ❌ Uninitialized reads: using uninitialized variables

**Safe Patterns**:
```c
// RAII pattern (initialize on declaration)
int *ptr = malloc(sizeof(int) * 10);
if (ptr == NULL) {
    return -1;  // Handle allocation failure
}

// Use, then immediately free
*ptr = 42;
free(ptr);
ptr = NULL;  // Prevent double-free

// Always check pointers before use
if (ptr != NULL) {
    *ptr = 42;
}
```

### 5. C23 New Features

**`nullptr` (type-safe null)**:
```c
// C17 (old)
int *ptr = NULL;  // NULL is just (void*)0

// C23 (new)
int *ptr = nullptr;  // Type-safe, not just a macro
```

**`typeof` (type inference)**:
```c
// C17 (old)
int x = 10;
int y = x;  // Must repeat type

// C23 (new)
int x = 10;
typeof(x) y = x;  // Type inferred
```

**`constexpr` (compile-time constants)**:
```c
// C23
constexpr int BUFFER_SIZE = 1024;
char buffer[BUFFER_SIZE];  // Guaranteed compile-time
```

**`[[attributes]]` (standard annotations)**:
```c
// C23
[[nodiscard]] int process_data(void) {
    return 42;  // Compiler warns if return value ignored
}

[[deprecated("Use new_function instead")]]
void old_function(void) { }
```

### 6. Code Formatting (clang-format 19.1.7)

**Configuration** (`.clang-format`):
```yaml
BasedOnStyle: LLVM
IndentWidth: 4
ColumnLimit: 100
PointerAlignment: Right
AlignConsecutiveAssignments: true
AllowShortFunctionsOnASingleLine: Empty
```

**CLI Commands**:
```bash
# Format single file
clang-format -i src/calculator.c

# Format all C files
find src/ -name "*.c" -o -name "*.h" | xargs clang-format -i

# Check formatting (CI/CD)
clang-format --dry-run --Werror src/*.c
```

---

## C23 Standard Features

### Key C23 Improvements

| Feature | Description | Example |
|---------|-------------|---------|
| `nullptr` | Type-safe null pointer | `int *ptr = nullptr;` |
| `typeof` | Type inference | `typeof(x) y = x;` |
| `constexpr` | Compile-time constants | `constexpr int SIZE = 10;` |
| `[[attributes]]` | Standard annotations | `[[nodiscard]] int func();` |
| `#embed` | Binary file inclusion | `#embed "data.bin"` |
| `_BitInt(N)` | Arbitrary-width integers | `_BitInt(24) x = 0x123456;` |
| `_Decimal*` | Decimal floating-point | `_Decimal64 price = 19.99;` |

---

## Example Workflow

### RED Phase (Failing Test)

```c
// tests/test_calculator.c
#include "unity.h"
#include "../src/calculator.h"

void test_divide_by_zero_returns_error(void) {
    int result;
    int status = divide(10, 0, &result);
    TEST_ASSERT_EQUAL_INT(-1, status);  // Expect error code
}
```

**Run**: `gcc -o test_runner tests/*.c src/*.c && ./test_runner`
**Result**: ❌ FAIL (function not implemented)

### GREEN Phase (Minimal Implementation)

```c
// src/calculator.c
#include "calculator.h"
#include <stddef.h>

int divide(int a, int b, int *result) {
    if (b == 0) {
        return -1;  // Error: division by zero
    }
    if (result == NULL) {
        return -1;  // Error: null pointer
    }
    *result = a / b;
    return 0;  // Success
}
```

**Run**: `gcc -o test_runner tests/*.c src/*.c && ./test_runner`
**Result**: ✅ PASS

### REFACTOR Phase (Improve Quality)

```c
// src/calculator.h
#ifndef CALCULATOR_H
#define CALCULATOR_H

/**
 * Divide two integers safely.
 * @param a Dividend
 * @param b Divisor
 * @param result Output parameter (must not be NULL)
 * @return 0 on success, -1 on error
 */
[[nodiscard]]
int divide(int a, int b, int *result);

#endif  // CALCULATOR_H
```

**Run**: `clang-format -i src/* && cppcheck src/ && valgrind ./test_runner`
**Result**: ✅ Clean (no warnings, no leaks)

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Release Date | Status |
|------|---------|---------|--------------|--------|
| **GCC** | 14.2.0 | Compiler | 2024-08-01 | ✅ Current |
| **Clang** | 19.1.7 | Compiler | 2024-12-02 | ✅ Current |
| **Unity** | 2.6.0 | Testing | 2024-03-10 | ✅ Current |
| **cppcheck** | 2.16.0 | Static Analysis | 2024-11-30 | ✅ Current |
| **CMake** | 3.31.0 | Build System | 2024-11-20 | ✅ Current |
| **Valgrind** | 3.24.0 | Memory Profiler | 2024-10-24 | ✅ Current |
| **clang-format** | 19.1.7 | Code Formatter | 2024-12-02 | ✅ Current |

---

## Compiler Flags Reference

### Development Flags (Debug)
```bash
gcc -Wall -Wextra -Wpedantic -Werror \
    -g -O0 -DDEBUG \
    -fsanitize=address -fsanitize=undefined \
    src/*.c -o program
```

### Production Flags (Release)
```bash
gcc -Wall -Wextra -Wpedantic -Werror \
    -O3 -march=native -DNDEBUG \
    -fstack-protector-strong -D_FORTIFY_SOURCE=2 \
    src/*.c -o program
```

### Security Flags
```bash
gcc -Wall -Wextra -Wpedantic -Werror \
    -fstack-protector-all \
    -D_FORTIFY_SOURCE=2 \
    -Wformat -Wformat-security \
    -fPIE -pie \
    src/*.c -o program
```

---

## TRUST 5 Compliance (C-specific)

### T — Test First
- ✅ Use Unity 2.6.0 for unit tests
- ✅ Achieve ≥85% code coverage
- ✅ Test edge cases (NULL, overflow, underflow)
- ✅ Memory leak testing (Valgrind/ASAN)

**Example**:
```bash
# Run tests with coverage
gcc -fprofile-arcs -ftest-coverage tests/*.c src/*.c -o test_runner
./test_runner
gcov src/*.c  # Generate coverage report
```

### R — Readable
- ✅ Follow clang-format style guide
- ✅ Document public APIs (Doxygen style)
- ✅ Use descriptive names (`calculate_sum` not `calc_s`)
- ✅ Limit function complexity (≤10 cyclomatic)

**Example**:
```c
/**
 * @brief Add two integers with overflow detection.
 * @param a First operand
 * @param b Second operand
 * @return Sum, saturated at INT_MAX on overflow
 */
int safe_add(int a, int b) {
    if (a > 0 && b > INT_MAX - a) {
        return INT_MAX;  // Saturate on overflow
    }
    return a + b;
}
```

### U — Unified (Type Safety)
- ✅ Use C23 `nullptr` instead of `NULL`
- ✅ Enable `-Werror` (treat warnings as errors)
- ✅ Validate all pointer arguments
- ✅ Use `const` for read-only parameters

**Example**:
```c
// C23 type safety
int process(const char *input, int *output) {
    if (input == nullptr || output == nullptr) {
        return -1;  // Validate inputs
    }
    *output = strlen(input);
    return 0;
}
```

### S — Secured
- ✅ Run cppcheck with `--enable=all`
- ✅ Use `-D_FORTIFY_SOURCE=2` (buffer checks)
- ✅ Enable stack protection (`-fstack-protector-all`)
- ✅ Avoid unsafe functions (`strcpy`, `sprintf`)

**Secure Alternatives**:
```c
// ❌ UNSAFE
char buf[10];
strcpy(buf, user_input);  // Buffer overflow risk

// ✅ SAFE
char buf[10];
strncpy(buf, user_input, sizeof(buf) - 1);
buf[sizeof(buf) - 1] = '\0';  // Ensure null-termination
```

### T — Trackable
- ✅ Add `@TAG` comments in code
- ✅ Link tests to specs (`@TEST:CALC-001`)
- ✅ Document HISTORY in SPEC files
- ✅ Tag commits with TAG references

**Example**:
```c
// @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: tests/test_calculator.c
int add(int a, int b) {
    return a + b;
}
```

---

## Common Patterns

### Error Handling (Return Codes)
```c
// Return 0 on success, negative on error
int read_file(const char *path, char **buffer) {
    if (path == nullptr || buffer == nullptr) {
        return -1;  // EINVAL (invalid argument)
    }

    FILE *f = fopen(path, "r");
    if (f == nullptr) {
        return -2;  // ENOENT (file not found)
    }

    // Read logic...
    fclose(f);
    return 0;  // Success
}
```

### Resource Cleanup (RAII-like)
```c
// Allocate and free in the same scope
void process_data(void) {
    int *data = malloc(sizeof(int) * 100);
    if (data == nullptr) {
        return;
    }

    // Use data...

    free(data);  // Always free before return
}
```

### Defensive Programming
```c
// Always validate inputs
int divide(int a, int b, int *result) {
    // Check for division by zero
    if (b == 0) {
        return -1;
    }

    // Check for null output pointer
    if (result == nullptr) {
        return -1;
    }

    *result = a / b;
    return 0;
}
```

---

## Anti-Patterns (Avoid)

### Memory Leaks
```c
// ❌ BAD: Memory leak
void leak_example(void) {
    int *ptr = malloc(sizeof(int) * 10);
    // Forgot to free(ptr)
}

// ✅ GOOD: Always free
void no_leak_example(void) {
    int *ptr = malloc(sizeof(int) * 10);
    if (ptr != nullptr) {
        // Use ptr...
        free(ptr);
    }
}
```

### Buffer Overflows
```c
// ❌ BAD: No bounds check
void overflow_example(char *dest, const char *src) {
    strcpy(dest, src);  // Unsafe!
}

// ✅ GOOD: Bounds check
void safe_copy(char *dest, const char *src, size_t dest_size) {
    strncpy(dest, src, dest_size - 1);
    dest[dest_size - 1] = '\0';
}
```

### Uninitialized Variables
```c
// ❌ BAD: Undefined behavior
void uninit_example(void) {
    int x;  // Not initialized
    printf("%d\n", x);  // Undefined behavior!
}

// ✅ GOOD: Always initialize
void init_example(void) {
    int x = 0;  // Initialized
    printf("%d\n", x);
}
```

---

## Inputs

- `.c` source files, `.h` headers
- `CMakeLists.txt`, `Makefile` build configs
- Unity test suites (`test_*.c`)

## Outputs

- Compiled binaries (executables, libraries)
- Test execution reports (Unity output)
- Static analysis reports (cppcheck XML)
- Memory leak reports (Valgrind output)

## Failure Modes

- **Missing tools**: Unity, cppcheck, Valgrind not installed
- **Compiler errors**: Syntax errors, type mismatches
- **Test failures**: Failing Unity assertions
- **Coverage gaps**: <85% code coverage
- **Memory leaks**: Detected by Valgrind/ASAN

**Mitigation**:
```bash
# Install tools (Ubuntu/Debian)
sudo apt install gcc clang cmake cppcheck valgrind

# Verify installation
gcc --version
cppcheck --version
valgrind --version
```

---

## Dependencies

- **Foundation Skills**: `moai-foundation-trust` (quality gates), `moai-foundation-tags` (traceability)
- **Essentials Skills**: `moai-essentials-debug` (debugging), `moai-essentials-review` (code review)
- **Domain Skills**: `moai-domain-security` (secure coding), `moai-domain-backend` (systems design)

---

## References (Latest Documentation)

_Documentation links updated 2025-10-22_

- [C23 Standard (ISO/IEC 9899:2023)](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n3220.pdf)
- [GCC 14 Manual](https://gcc.gnu.org/onlinedocs/gcc-14.2.0/gcc/)
- [Clang 19 Documentation](https://clang.llvm.org/docs/)
- [Unity Testing Framework](https://github.com/ThrowTheSwitch/Unity)
- [cppcheck Manual](https://cppcheck.sourceforge.io/manual.pdf)
- [CMake 3.31 Documentation](https://cmake.org/cmake/help/v3.31/)
- [Valgrind Quick Start](https://valgrind.org/docs/manual/quick-start.html)
- [AddressSanitizer Documentation](https://github.com/google/sanitizers/wiki/AddressSanitizer)

---

## Changelog

- **v2.0.0** (2025-10-22): Major update with C23 support, Unity 2.6.0, comprehensive memory safety guide, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release with C17 support

---

## Works Well With

- `moai-foundation-trust` (TRUST 5 quality gates)
- `moai-alfred-code-reviewer` (automated code review)
- `moai-essentials-debug` (debugging workflows)
- `moai-domain-security` (secure coding patterns)
- `moai-lang-cpp` (C++ interoperability)

---

## Best Practices Summary

✅ **DO**:
- Use Unity 2.6.0 for unit testing
- Enable all compiler warnings (`-Wall -Wextra -Wpedantic -Werror`)
- Run cppcheck and Valgrind on every commit
- Document public APIs with Doxygen comments
- Validate all pointer arguments for `nullptr`
- Use C23 features (`nullptr`, `typeof`, `constexpr`)
- Follow clang-format style guide
- Maintain ≥85% test coverage
- Tag all code with `@CODE:ID`

❌ **DON'T**:
- Use unsafe functions (`strcpy`, `sprintf`, `gets`)
- Ignore compiler warnings
- Skip memory leak testing
- Mix testing frameworks (Unity only)
- Use `NULL` instead of `nullptr` (C23)
- Forget to `free()` allocated memory
- Access freed memory (use-after-free)
- Skip input validation

---

## Advanced Topics

### Multi-threading (POSIX Threads)

**pthread basics**:
```c
#include <pthread.h>
#include <stdio.h>

// Thread function
void *worker(void *arg) {
    int *id = (int *)arg;
    printf("Thread %d running\n", *id);
    return nullptr;
}

int main(void) {
    pthread_t threads[4];
    int ids[4] = {1, 2, 3, 4};

    // Create threads
    for (int i = 0; i < 4; i++) {
        pthread_create(&threads[i], nullptr, worker, &ids[i]);
    }

    // Wait for threads
    for (int i = 0; i < 4; i++) {
        pthread_join(threads[i], nullptr);
    }

    return 0;
}
```

**Thread synchronization (mutex)**:
```c
#include <pthread.h>

typedef struct {
    int counter;
    pthread_mutex_t lock;
} SharedData;

void *increment(void *arg) {
    SharedData *data = (SharedData *)arg;

    pthread_mutex_lock(&data->lock);
    data->counter++;  // Critical section
    pthread_mutex_unlock(&data->lock);

    return nullptr;
}

int main(void) {
    SharedData data = {0, PTHREAD_MUTEX_INITIALIZER};
    pthread_t threads[10];

    for (int i = 0; i < 10; i++) {
        pthread_create(&threads[i], nullptr, increment, &data);
    }

    for (int i = 0; i < 10; i++) {
        pthread_join(threads[i], nullptr);
    }

    printf("Final counter: %d\n", data.counter);  // Should be 10
    pthread_mutex_destroy(&data.lock);
    return 0;
}
```

**Condition variables**:
```c
#include <pthread.h>
#include <stdbool.h>

typedef struct {
    bool ready;
    pthread_mutex_t lock;
    pthread_cond_t cond;
} Signal;

void *producer(void *arg) {
    Signal *sig = (Signal *)arg;

    pthread_mutex_lock(&sig->lock);
    sig->ready = true;
    pthread_cond_signal(&sig->cond);  // Wake up waiting thread
    pthread_mutex_unlock(&sig->lock);

    return nullptr;
}

void *consumer(void *arg) {
    Signal *sig = (Signal *)arg;

    pthread_mutex_lock(&sig->lock);
    while (!sig->ready) {
        pthread_cond_wait(&sig->cond, &sig->lock);  // Wait for signal
    }
    printf("Data ready!\n");
    pthread_mutex_unlock(&sig->lock);

    return nullptr;
}
```

### File I/O Best Practices

**Safe file reading**:
```c
#include <stdio.h>
#include <stdlib.h>
#include <errno.h>

int read_file(const char *path, char **buffer, size_t *size) {
    if (path == nullptr || buffer == nullptr || size == nullptr) {
        return -EINVAL;
    }

    FILE *f = fopen(path, "rb");
    if (f == nullptr) {
        return -errno;
    }

    // Get file size
    fseek(f, 0, SEEK_END);
    long fsize = ftell(f);
    fseek(f, 0, SEEK_SET);

    if (fsize < 0) {
        fclose(f);
        return -EIO;
    }

    // Allocate buffer
    *buffer = malloc(fsize + 1);
    if (*buffer == nullptr) {
        fclose(f);
        return -ENOMEM;
    }

    // Read data
    size_t read = fread(*buffer, 1, fsize, f);
    (*buffer)[read] = '\0';
    *size = read;

    fclose(f);
    return 0;
}

// Usage
int main(void) {
    char *content = nullptr;
    size_t size = 0;

    int result = read_file("input.txt", &content, &size);
    if (result == 0) {
        printf("Read %zu bytes\n", size);
        free(content);
    }

    return 0;
}
```

**Buffered writing**:
```c
#include <stdio.h>

int write_log(const char *message) {
    FILE *f = fopen("app.log", "a");  // Append mode
    if (f == nullptr) {
        return -1;
    }

    fprintf(f, "[LOG] %s\n", message);
    fclose(f);
    return 0;
}
```

### Dynamic Memory Management

**Custom allocators**:
```c
#include <stdlib.h>
#include <string.h>

// Arena allocator (fast, no individual frees)
typedef struct {
    char *buffer;
    size_t size;
    size_t offset;
} Arena;

Arena *arena_create(size_t size) {
    Arena *arena = malloc(sizeof(Arena));
    if (arena == nullptr) {
        return nullptr;
    }

    arena->buffer = malloc(size);
    if (arena->buffer == nullptr) {
        free(arena);
        return nullptr;
    }

    arena->size = size;
    arena->offset = 0;
    return arena;
}

void *arena_alloc(Arena *arena, size_t size) {
    if (arena->offset + size > arena->size) {
        return nullptr;  // Out of memory
    }

    void *ptr = arena->buffer + arena->offset;
    arena->offset += size;
    return ptr;
}

void arena_destroy(Arena *arena) {
    if (arena != nullptr) {
        free(arena->buffer);
        free(arena);
    }
}

// Usage
int main(void) {
    Arena *arena = arena_create(1024);

    int *nums = arena_alloc(arena, sizeof(int) * 10);
    char *str = arena_alloc(arena, 100);

    // Use allocations...

    arena_destroy(arena);  // Free everything at once
    return 0;
}
```

**Pool allocator (fixed-size objects)**:
```c
#include <stdlib.h>

typedef struct Block {
    struct Block *next;
} Block;

typedef struct {
    Block *free_list;
    size_t block_size;
    size_t block_count;
    void *memory;
} Pool;

Pool *pool_create(size_t block_size, size_t block_count) {
    Pool *pool = malloc(sizeof(Pool));
    if (pool == nullptr) {
        return nullptr;
    }

    pool->block_size = block_size;
    pool->block_count = block_count;
    pool->memory = malloc(block_size * block_count);

    if (pool->memory == nullptr) {
        free(pool);
        return nullptr;
    }

    // Build free list
    pool->free_list = pool->memory;
    Block *current = pool->free_list;

    for (size_t i = 1; i < block_count; i++) {
        current->next = (Block *)((char *)pool->memory + i * block_size);
        current = current->next;
    }
    current->next = nullptr;

    return pool;
}

void *pool_alloc(Pool *pool) {
    if (pool->free_list == nullptr) {
        return nullptr;  // Pool exhausted
    }

    Block *block = pool->free_list;
    pool->free_list = block->next;
    return block;
}

void pool_free(Pool *pool, void *ptr) {
    Block *block = (Block *)ptr;
    block->next = pool->free_list;
    pool->free_list = block;
}

void pool_destroy(Pool *pool) {
    if (pool != nullptr) {
        free(pool->memory);
        free(pool);
    }
}
```

### Data Structures

**Dynamic array (vector)**:
```c
#include <stdlib.h>
#include <string.h>

typedef struct {
    int *data;
    size_t size;
    size_t capacity;
} Vector;

Vector *vector_create(void) {
    Vector *vec = malloc(sizeof(Vector));
    if (vec == nullptr) {
        return nullptr;
    }

    vec->capacity = 16;
    vec->size = 0;
    vec->data = malloc(sizeof(int) * vec->capacity);

    if (vec->data == nullptr) {
        free(vec);
        return nullptr;
    }

    return vec;
}

int vector_push(Vector *vec, int value) {
    if (vec->size >= vec->capacity) {
        // Grow capacity (2x)
        size_t new_capacity = vec->capacity * 2;
        int *new_data = realloc(vec->data, sizeof(int) * new_capacity);

        if (new_data == nullptr) {
            return -1;  // Allocation failed
        }

        vec->data = new_data;
        vec->capacity = new_capacity;
    }

    vec->data[vec->size++] = value;
    return 0;
}

int vector_get(const Vector *vec, size_t index, int *out) {
    if (index >= vec->size || out == nullptr) {
        return -1;
    }

    *out = vec->data[index];
    return 0;
}

void vector_destroy(Vector *vec) {
    if (vec != nullptr) {
        free(vec->data);
        free(vec);
    }
}
```

**Hash table (open addressing)**:
```c
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

typedef struct {
    char *key;
    int value;
    bool occupied;
} Entry;

typedef struct {
    Entry *entries;
    size_t capacity;
    size_t size;
} HashMap;

// FNV-1a hash
static uint32_t hash_string(const char *key) {
    uint32_t hash = 2166136261u;
    while (*key) {
        hash ^= (uint8_t)(*key++);
        hash *= 16777619;
    }
    return hash;
}

HashMap *hashmap_create(size_t capacity) {
    HashMap *map = malloc(sizeof(HashMap));
    if (map == nullptr) {
        return nullptr;
    }

    map->capacity = capacity;
    map->size = 0;
    map->entries = calloc(capacity, sizeof(Entry));

    if (map->entries == nullptr) {
        free(map);
        return nullptr;
    }

    return map;
}

int hashmap_put(HashMap *map, const char *key, int value) {
    if (map->size >= map->capacity * 0.75) {
        return -1;  // Need to resize (not implemented here)
    }

    uint32_t hash = hash_string(key);
    size_t index = hash % map->capacity;

    // Linear probing
    while (map->entries[index].occupied) {
        if (strcmp(map->entries[index].key, key) == 0) {
            // Update existing key
            map->entries[index].value = value;
            return 0;
        }
        index = (index + 1) % map->capacity;
    }

    // Insert new entry
    map->entries[index].key = strdup(key);
    map->entries[index].value = value;
    map->entries[index].occupied = true;
    map->size++;
    return 0;
}

int hashmap_get(const HashMap *map, const char *key, int *out) {
    uint32_t hash = hash_string(key);
    size_t index = hash % map->capacity;

    while (map->entries[index].occupied) {
        if (strcmp(map->entries[index].key, key) == 0) {
            *out = map->entries[index].value;
            return 0;
        }
        index = (index + 1) % map->capacity;
    }

    return -1;  // Not found
}

void hashmap_destroy(HashMap *map) {
    if (map != nullptr) {
        for (size_t i = 0; i < map->capacity; i++) {
            if (map->entries[i].occupied) {
                free(map->entries[i].key);
            }
        }
        free(map->entries);
        free(map);
    }
}
```

### Performance Optimization

**Compiler optimizations**:
```c
// 1. Inline functions (avoid call overhead)
static inline int fast_add(int a, int b) {
    return a + b;
}

// 2. Restrict pointers (no aliasing)
void process(int * restrict a, int * restrict b, int n) {
    for (int i = 0; i < n; i++) {
        a[i] += b[i];  // Compiler knows a and b don't overlap
    }
}

// 3. Loop unrolling
void sum_unrolled(const int *arr, int n, int *result) {
    int sum = 0;
    int i;

    // Process 4 elements at a time
    for (i = 0; i <= n - 4; i += 4) {
        sum += arr[i];
        sum += arr[i + 1];
        sum += arr[i + 2];
        sum += arr[i + 3];
    }

    // Handle remaining elements
    for (; i < n; i++) {
        sum += arr[i];
    }

    *result = sum;
}

// 4. Branch prediction hints (GCC)
int process_with_hints(int x) {
    if (__builtin_expect(x > 0, 1)) {  // Likely branch
        return x * 2;
    } else {
        return x / 2;  // Unlikely branch
    }
}
```

**Cache-friendly code**:
```c
// BAD: Poor cache locality (column-major access)
void bad_matrix_sum(int matrix[1000][1000]) {
    int sum = 0;
    for (int col = 0; col < 1000; col++) {
        for (int row = 0; row < 1000; row++) {
            sum += matrix[row][col];  // Cache misses!
        }
    }
}

// GOOD: Cache-friendly (row-major access)
void good_matrix_sum(int matrix[1000][1000]) {
    int sum = 0;
    for (int row = 0; row < 1000; row++) {
        for (int col = 0; col < 1000; col++) {
            sum += matrix[row][col];  // Sequential access
        }
    }
}
```

### Error Handling Patterns

**Errno-style errors**:
```c
#include <errno.h>
#include <string.h>

int divide_safe(int a, int b, int *result) {
    if (result == nullptr) {
        errno = EINVAL;
        return -1;
    }

    if (b == 0) {
        errno = EDOM;  // Math domain error
        return -1;
    }

    *result = a / b;
    return 0;
}

// Usage
int main(void) {
    int result;
    if (divide_safe(10, 0, &result) != 0) {
        fprintf(stderr, "Error: %s\n", strerror(errno));
        return 1;
    }
    return 0;
}
```

**Result type pattern**:
```c
typedef enum {
    OK,
    ERR_NULL_POINTER,
    ERR_OUT_OF_MEMORY,
    ERR_INVALID_INPUT,
    ERR_IO_ERROR
} ErrorCode;

typedef struct {
    ErrorCode code;
    int value;
    const char *message;
} Result;

Result divide_result(int a, int b) {
    if (b == 0) {
        return (Result){ERR_INVALID_INPUT, 0, "Division by zero"};
    }
    return (Result){OK, a / b, nullptr};
}

// Usage
int main(void) {
    Result res = divide_result(10, 2);
    if (res.code == OK) {
        printf("Result: %d\n", res.value);
    } else {
        fprintf(stderr, "Error: %s\n", res.message);
    }
    return 0;
}
```

### Testing Strategies

**Parameterized tests (Unity)**:
```c
#include "unity.h"

typedef struct {
    int a;
    int b;
    int expected;
} TestCase;

void test_add_cases(void) {
    TestCase cases[] = {
        {2, 3, 5},
        {-1, 1, 0},
        {0, 0, 0},
        {100, 200, 300}
    };

    for (size_t i = 0; i < sizeof(cases) / sizeof(cases[0]); i++) {
        int result = add(cases[i].a, cases[i].b);
        TEST_ASSERT_EQUAL_INT_MESSAGE(cases[i].expected, result,
            "Failed for case index");
    }
}
```

**Mock objects**:
```c
// Interface (function pointer)
typedef struct {
    int (*read)(void *ctx, char *buffer, size_t size);
    int (*write)(void *ctx, const char *data, size_t size);
    void *context;
} IO;

// Mock implementation (for testing)
typedef struct {
    const char *mock_data;
    size_t mock_size;
    size_t mock_offset;
} MockContext;

int mock_read(void *ctx, char *buffer, size_t size) {
    MockContext *mock = (MockContext *)ctx;
    size_t remaining = mock->mock_size - mock->mock_offset;
    size_t to_read = (size < remaining) ? size : remaining;

    memcpy(buffer, mock->mock_data + mock->mock_offset, to_read);
    mock->mock_offset += to_read;
    return to_read;
}

// Test using mock
void test_with_mock(void) {
    const char *test_data = "Hello, World!";
    MockContext mock = {test_data, strlen(test_data), 0};

    IO io = {mock_read, nullptr, &mock};

    char buffer[100];
    int bytes_read = io.read(io.context, buffer, sizeof(buffer));

    TEST_ASSERT_EQUAL_INT(13, bytes_read);
    TEST_ASSERT_EQUAL_STRING("Hello, World!", buffer);
}
```

### Debugging Techniques

**Assertions**:
```c
#include <assert.h>

int divide_checked(int a, int b) {
    assert(b != 0 && "Division by zero");  // Runtime check
    return a / b;
}

// Custom assertion with message
#define ASSERT_MSG(cond, msg) \
    do { \
        if (!(cond)) { \
            fprintf(stderr, "Assertion failed: %s (%s:%d)\n", \
                    msg, __FILE__, __LINE__); \
            abort(); \
        } \
    } while (0)
```

**Debug macros**:
```c
#ifdef DEBUG
#define DEBUG_PRINT(fmt, ...) \
    fprintf(stderr, "[DEBUG] %s:%d: " fmt "\n", \
            __FILE__, __LINE__, ##__VA_ARGS__)
#else
#define DEBUG_PRINT(fmt, ...) ((void)0)
#endif

// Usage
void process_data(int *data, size_t size) {
    DEBUG_PRINT("Processing %zu items", size);

    for (size_t i = 0; i < size; i++) {
        DEBUG_PRINT("Item[%zu] = %d", i, data[i]);
        // Process...
    }
}
```

### Portable Code Patterns

**Platform detection**:
```c
#if defined(_WIN32) || defined(_WIN64)
    #define PLATFORM_WINDOWS
#elif defined(__linux__)
    #define PLATFORM_LINUX
#elif defined(__APPLE__)
    #define PLATFORM_MACOS
#else
    #error "Unsupported platform"
#endif

// Platform-specific code
#ifdef PLATFORM_WINDOWS
    #include <windows.h>
    void platform_sleep(int ms) {
        Sleep(ms);
    }
#else
    #include <unistd.h>
    void platform_sleep(int ms) {
        usleep(ms * 1000);
    }
#endif
```

**Endianness handling**:
```c
#include <stdint.h>

// Detect endianness at runtime
bool is_little_endian(void) {
    uint32_t x = 1;
    return *(uint8_t *)&x == 1;
}

// Byte swap functions
uint16_t bswap16(uint16_t x) {
    return (x << 8) | (x >> 8);
}

uint32_t bswap32(uint32_t x) {
    return ((x & 0xFF000000) >> 24) |
           ((x & 0x00FF0000) >> 8) |
           ((x & 0x0000FF00) << 8) |
           ((x & 0x000000FF) << 24);
}

// Convert to/from network byte order
uint32_t htonl_portable(uint32_t x) {
    return is_little_endian() ? bswap32(x) : x;
}

uint32_t ntohl_portable(uint32_t x) {
    return is_little_endian() ? bswap32(x) : x;
}
```

### Security Hardening

**Input validation**:
```c
#include <ctype.h>
#include <limits.h>

// Validate string is alphanumeric
bool is_valid_username(const char *username) {
    if (username == nullptr || *username == '\0') {
        return false;
    }

    size_t len = 0;
    while (*username) {
        if (!isalnum(*username) && *username != '_') {
            return false;  // Invalid character
        }
        username++;
        len++;

        if (len > 32) {
            return false;  // Too long
        }
    }

    return len >= 3;  // Minimum length
}

// Safe integer parsing
int parse_int_safe(const char *str, int *out) {
    if (str == nullptr || out == nullptr) {
        return -1;
    }

    char *endptr;
    errno = 0;
    long val = strtol(str, &endptr, 10);

    if (errno == ERANGE || val < INT_MIN || val > INT_MAX) {
        return -1;  // Overflow
    }

    if (endptr == str || *endptr != '\0') {
        return -1;  // Invalid format
    }

    *out = (int)val;
    return 0;
}
```

**Secure memory operations**:
```c
#include <string.h>

// Securely clear sensitive data
void secure_zero(void *ptr, size_t size) {
    volatile unsigned char *p = ptr;
    while (size--) {
        *p++ = 0;
    }
}

// Constant-time string comparison (timing-attack resistant)
bool secure_strcmp(const char *a, const char *b, size_t len) {
    unsigned char diff = 0;
    for (size_t i = 0; i < len; i++) {
        diff |= a[i] ^ b[i];
    }
    return diff == 0;
}

// Usage: compare passwords
bool verify_password(const char *input, const char *stored) {
    size_t len = strlen(stored);
    if (strlen(input) != len) {
        return false;
    }
    return secure_strcmp(input, stored, len);
}
```

---

**End of moai-lang-c Skill v2.0.0**
