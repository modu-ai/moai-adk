# moai-lang-cpp - CLI Reference

_Last updated: 2025-10-22_

## Quick Reference

### Installation

```bash
# GCC/G++ (via package manager)
# Debian/Ubuntu
sudo apt-get install g++ clang cmake clang-format clang-tidy

# macOS
brew install gcc llvm cmake clang-format

# Verify installation
g++ --version          # Should show 14.2.0+
clang++ --version      # Should show 19.1.7+
cmake --version        # Should show 3.22.0+
clang-format --version # Should show 19.1.7+
```

### Common Commands

```bash
# Compile (G++)
g++ -std=c++20 -Wall -Wextra -Wpedantic -Werror -o program main.cpp

# Compile (Clang++)
clang++ -std=c++20 -Wall -Wextra -Wpedantic -Werror -o program main.cpp

# Format code (clang-format)
clang-format -i src/*.cpp src/*.h

# Static Analysis (clang-tidy)
clang-tidy src/*.cpp -- -std=c++20

# Build with CMake
cmake -B build -S . -DCMAKE_BUILD_TYPE=Release
cmake --build build

# Run Google Test
./build/tests/my_tests
```

## Tool Versions (2025-10-22)

- **C++ Standard**: C++20 (stable), C++23 (experimental)
- **GCC**: 14.2.0 - Primary compiler (GNU Compiler Collection)
- **Clang**: 19.1.7 - Alternative compiler with better diagnostics
- **Google Test**: 1.15.2 - Unit testing and mocking framework
- **CMake**: 3.22+ - Cross-platform build system generator
- **clang-format**: 19.1.7 - Code formatter
- **clang-tidy**: 19.1.7 - Static analysis tool

## Official Documentation Links

- **C++ Standards**: https://isocpp.org/
- **cppreference**: https://en.cppreference.com/
- **Google Test**: https://google.github.io/googletest/
- **CMake**: https://cmake.org/cmake/help/latest/
- **Clang Tools**: https://clang.llvm.org/docs/
- **GCC**: https://gcc.gnu.org/onlinedocs/

## C++ Standards Overview

### C++20 (ISO/IEC 14882:2020)
- **Concepts**: Constraints for template parameters
- **Ranges**: Composable algorithms and views
- **Coroutines**: Stackless coroutine support
- **Modules**: Alternative to header files (experimental)
- **Three-way comparison**: `operator<=>`
- **Designated initializers**: `Point{.x=1, .y=2}`
- **`constexpr` enhancements**: Virtual functions, `std::vector`
- **`std::format`**: Type-safe string formatting
- **Calendar and timezone**: `<chrono>` extensions

### C++23 (ISO/IEC 14882:2024)
- **`std::expected`**: Error handling without exceptions
- **`std::mdspan`**: Multi-dimensional array view
- **`std::flat_map`**: Flat associative containers
- **`if consteval`**: Compile-time conditional execution
- **Deducing `this`**: Explicit object parameter
- **`std::print`**: Simplified output formatting
- **Relaxed `constexpr`**: More constexpr capabilities

## Google Test Framework

### Core Components
- **gtest**: Unit testing framework
- **gmock**: Mocking framework for interfaces
- Portable: works across platforms (Linux, macOS, Windows)

### Test Structure
```cpp
#include <gtest/gtest.h>

// Simple test
TEST(TestSuiteName, TestName) {
    EXPECT_EQ(expected, actual);
}

// Test with fixture
class MyTest : public ::testing::Test {
protected:
    void SetUp() override { /* setup code */ }
    void TearDown() override { /* cleanup code */ }
};

TEST_F(MyTest, TestName) {
    EXPECT_TRUE(condition);
}

// Parameterized test
class ParamTest : public ::testing::TestWithParam<int> {};

TEST_P(ParamTest, TestName) {
    int value = GetParam();
    EXPECT_GT(value, 0);
}

INSTANTIATE_TEST_SUITE_P(
    TestSuite,
    ParamTest,
    ::testing::Values(1, 2, 3, 5, 8)
);
```

### Common Assertions

#### Basic Assertions
```cpp
// Fatal assertions (stop on failure)
ASSERT_TRUE(condition);
ASSERT_FALSE(condition);
ASSERT_EQ(expected, actual);
ASSERT_NE(val1, val2);
ASSERT_LT(val1, val2);  // Less than
ASSERT_LE(val1, val2);  // Less than or equal
ASSERT_GT(val1, val2);  // Greater than
ASSERT_GE(val1, val2);  // Greater than or equal

// Non-fatal assertions (continue on failure)
EXPECT_TRUE(condition);
EXPECT_FALSE(condition);
EXPECT_EQ(expected, actual);
EXPECT_NE(val1, val2);
EXPECT_LT(val1, val2);
EXPECT_LE(val1, val2);
EXPECT_GT(val1, val2);
EXPECT_GE(val1, val2);
```

#### String Assertions
```cpp
EXPECT_STREQ(str1, str2);      // C-strings equal
EXPECT_STRNE(str1, str2);      // C-strings not equal
EXPECT_STRCASEEQ(str1, str2);  // Case-insensitive equal
EXPECT_STRCASENE(str1, str2);  // Case-insensitive not equal
```

#### Floating-Point Assertions
```cpp
EXPECT_FLOAT_EQ(val1, val2);   // ~4 ULP tolerance
EXPECT_DOUBLE_EQ(val1, val2);  // ~4 ULP tolerance
EXPECT_NEAR(val1, val2, abs_error);  // Custom tolerance
```

#### Exception Assertions
```cpp
EXPECT_THROW(statement, exception_type);
EXPECT_ANY_THROW(statement);
EXPECT_NO_THROW(statement);
```

### Google Mock

```cpp
#include <gmock/gmock.h>

class MockClass : public InterfaceClass {
public:
    MOCK_METHOD(ReturnType, MethodName, (ArgTypes...), (override));
};

// Usage
MockClass mock;
EXPECT_CALL(mock, MethodName(arg_matchers...))
    .Times(n)
    .WillOnce(Return(value))
    .WillRepeatedly(Return(value));
```

## CMake Integration

### Basic CMakeLists.txt
```cmake
cmake_minimum_required(VERSION 3.22)
project(MyProject CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Compiler flags
if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang")
    add_compile_options(-Wall -Wextra -Wpedantic -Werror)
elseif(MSVC)
    add_compile_options(/W4 /WX)
endif()

# Source files
add_library(mylib src/file1.cpp src/file2.cpp)
target_include_directories(mylib PUBLIC include/)

# Executable
add_executable(myapp main.cpp)
target_link_libraries(myapp PRIVATE mylib)

# Testing
include(CTest)
enable_testing()
add_subdirectory(tests)
```

### Testing with CMake
```cmake
# tests/CMakeLists.txt
include(FetchContent)
FetchContent_Declare(
  googletest
  GIT_REPOSITORY https://github.com/google/googletest.git
  GIT_TAG        v1.15.2
)
FetchContent_MakeAvailable(googletest)

add_executable(my_tests test_file1.cpp test_file2.cpp)
target_link_libraries(my_tests PRIVATE mylib GTest::gtest_main)

include(GoogleTest)
gtest_discover_tests(my_tests)
```

## clang-format Configuration

### .clang-format (Example)
```yaml
---
BasedOnStyle: Google
IndentWidth: 4
ColumnLimit: 100
PointerAlignment: Left
ReferenceAlignment: Left
NamespaceIndentation: All
AllowShortFunctionsOnASingleLine: Empty
AllowShortIfStatementsOnASingleLine: Never
AllowShortLoopsOnASingleLine: false
```

### Usage
```bash
# Format single file
clang-format -i src/myfile.cpp

# Format all files in directory
find src/ -name "*.cpp" -o -name "*.h" | xargs clang-format -i

# Check formatting (CI)
clang-format --dry-run --Werror src/*.cpp
```

## clang-tidy Configuration

### .clang-tidy (Example)
```yaml
---
Checks: >
  -*,
  clang-analyzer-*,
  cppcoreguidelines-*,
  modernize-*,
  performance-*,
  readability-*,
  bugprone-*
CheckOptions:
  - key: readability-identifier-naming.ClassCase
    value: CamelCase
  - key: readability-identifier-naming.FunctionCase
    value: camelBack
  - key: readability-identifier-naming.VariableCase
    value: lower_case
```

### Usage
```bash
# Run clang-tidy
clang-tidy src/myfile.cpp -- -std=c++20 -Iinclude/

# Fix issues automatically
clang-tidy -fix src/myfile.cpp -- -std=c++20 -Iinclude/

# Use with CMake compile_commands.json
cmake -B build -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
clang-tidy -p build src/myfile.cpp
```

## Code Coverage (gcov/lcov)

### Enable Coverage in CMake
```cmake
# Add to CMakeLists.txt
if(CMAKE_BUILD_TYPE STREQUAL "Coverage")
    add_compile_options(--coverage)
    add_link_options(--coverage)
endif()
```

### Generate Coverage Report
```bash
# Build with coverage
cmake -B build -DCMAKE_BUILD_TYPE=Coverage
cmake --build build

# Run tests
cd build && ctest

# Generate coverage report
lcov --capture --directory . --output-file coverage.info
lcov --remove coverage.info '/usr/*' --output-file coverage.info
genhtml coverage.info --output-directory coverage_html
```

## Compiler Flags Reference

### Warning Flags
```bash
# Essential warnings
-Wall           # Enable common warnings
-Wextra         # Enable extra warnings
-Wpedantic      # Strict ISO C++ compliance
-Werror         # Treat warnings as errors

# Additional useful warnings
-Wshadow        # Warn on variable shadowing
-Wconversion    # Warn on implicit type conversions
-Wsign-conversion  # Warn on sign conversions
-Wnon-virtual-dtor  # Warn on non-virtual destructors
```

### Optimization Flags
```bash
-O0    # No optimization (debug)
-O1    # Basic optimization
-O2    # Recommended for release
-O3    # Aggressive optimization
-Og    # Optimize for debugging
-Os    # Optimize for size
```

### Debug Flags
```bash
-g      # Generate debug information
-ggdb   # GDB-specific debug info
-g3     # Maximum debug information
```

### Sanitizers (Runtime Checks)
```bash
# Address Sanitizer (memory errors)
-fsanitize=address

# Undefined Behavior Sanitizer
-fsanitize=undefined

# Thread Sanitizer (data races)
-fsanitize=thread

# Memory Sanitizer (uninitialized reads)
-fsanitize=memory
```

## Best Practices

### Project Structure
```
project/
├── CMakeLists.txt
├── .clang-format
├── .clang-tidy
├── include/
│   └── mylib/
│       └── header.h
├── src/
│   └── implementation.cpp
├── tests/
│   ├── CMakeLists.txt
│   └── test_*.cpp
├── docs/
└── README.md
```

### Naming Conventions
- **Classes/Structs**: `PascalCase`
- **Functions/Methods**: `camelCase` or `snake_case` (consistent)
- **Variables**: `snake_case` or `camelCase` (consistent)
- **Constants**: `UPPER_SNAKE_CASE` or `kPascalCase`
- **Namespaces**: `lowercase`

### Modern C++ Guidelines
1. Use RAII for resource management
2. Prefer `std::unique_ptr` over raw pointers
3. Use `const` and `constexpr` liberally
4. Avoid manual memory management (`new`/`delete`)
5. Use `auto` for type deduction when clear
6. Prefer range-based for loops
7. Use smart pointers (`unique_ptr`, `shared_ptr`)
8. Avoid C-style casts (use `static_cast`, etc.)
9. Use `override` and `final` keywords
10. Initialize all variables

### Testing Guidelines
- **Test Granularity**: One test per behavior/condition
- **Test Independence**: Tests should not depend on each other
- **Test Naming**: Descriptive names (`TestWhat_WhenCondition_ThenExpectation`)
- **Fixtures**: Use for shared setup/teardown
- **Mocks**: Test interfaces, not implementations
- **Coverage Target**: ≥85% code coverage

## Common Pitfalls

### Pitfall 1: Memory Leaks
```cpp
// ❌ BAD
int* ptr = new int(42);
// ... forget to delete

// ✅ GOOD
auto ptr = std::make_unique<int>(42);
// Automatically cleaned up
```

### Pitfall 2: Dangling References
```cpp
// ❌ BAD
const std::string& getName() {
    std::string name = "temp";
    return name;  // Returns reference to local variable!
}

// ✅ GOOD
std::string getName() {
    return "temp";  // Return by value (move semantics)
}
```

### Pitfall 3: Uninitialized Variables
```cpp
// ❌ BAD
int x;
int y = x + 5;  // Undefined behavior

// ✅ GOOD
int x = 0;
int y = x + 5;  // Or use: int x{};
```

### Pitfall 4: Iterator Invalidation
```cpp
// ❌ BAD
std::vector<int> vec = {1, 2, 3};
for (auto it = vec.begin(); it != vec.end(); ++it) {
    vec.push_back(*it);  // Invalidates iterators!
}

// ✅ GOOD
std::vector<int> vec = {1, 2, 3};
size_t size = vec.size();
for (size_t i = 0; i < size; ++i) {
    vec.push_back(vec[i]);
}
```

---

_For working examples, see [examples.md](examples.md)_
