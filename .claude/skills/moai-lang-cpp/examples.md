# moai-lang-cpp - Working Examples

_Last updated: 2025-10-22_

## Example 1: Basic C++ Project Setup with Google Test and CMake

### Project Structure
```
my-cpp-project/
├── CMakeLists.txt
├── src/
│   ├── calculator.cpp
│   └── calculator.h
└── tests/
    ├── CMakeLists.txt
    └── test_calculator.cpp
```

### calculator.h
```cpp
#ifndef CALCULATOR_H
#define CALCULATOR_H

#include <stdexcept>

class Calculator {
public:
    int add(int a, int b);
    int subtract(int a, int b);
    double divide(double a, double b);
};

#endif
```

### calculator.cpp
```cpp
#include "calculator.h"

int Calculator::add(int a, int b) {
    return a + b;
}

int Calculator::subtract(int a, int b) {
    return a - b;
}

double Calculator::divide(double a, double b) {
    if (b == 0.0) {
        throw std::invalid_argument("Division by zero");
    }
    return a / b;
}
```

### tests/test_calculator.cpp
```cpp
#include <gtest/gtest.h>
#include "../src/calculator.h"

// Test fixture for Calculator tests
class CalculatorTest : public ::testing::Test {
protected:
    Calculator calc;

    void SetUp() override {
        // Runs before each test
    }

    void TearDown() override {
        // Runs after each test
    }
};

TEST_F(CalculatorTest, AddPositiveNumbers) {
    EXPECT_EQ(5, calc.add(2, 3));
}

TEST_F(CalculatorTest, AddNegativeNumbers) {
    EXPECT_EQ(-5, calc.add(-2, -3));
}

TEST_F(CalculatorTest, SubtractNumbers) {
    EXPECT_EQ(1, calc.subtract(3, 2));
}

TEST_F(CalculatorTest, DivideNumbers) {
    EXPECT_DOUBLE_EQ(2.5, calc.divide(5.0, 2.0));
}

TEST_F(CalculatorTest, DivideByZeroThrows) {
    EXPECT_THROW(calc.divide(5.0, 0.0), std::invalid_argument);
}

// Parameterized test example
class AdditionTest : public ::testing::TestWithParam<std::tuple<int, int, int>> {};

TEST_P(AdditionTest, ParameterizedAdd) {
    Calculator calc;
    auto [a, b, expected] = GetParam();
    EXPECT_EQ(expected, calc.add(a, b));
}

INSTANTIATE_TEST_SUITE_P(
    AddTestCases,
    AdditionTest,
    ::testing::Values(
        std::make_tuple(1, 2, 3),
        std::make_tuple(0, 0, 0),
        std::make_tuple(-1, 1, 0),
        std::make_tuple(100, 200, 300)
    )
);
```

### Root CMakeLists.txt
```cmake
cmake_minimum_required(VERSION 3.22)
project(Calculator CXX)

# Set C++ standard (C++20 or C++23)
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Compiler warnings
if(MSVC)
    add_compile_options(/W4 /WX)
else()
    add_compile_options(-Wall -Wextra -Wpedantic -Werror)
endif()

# Main library (not executable)
add_library(calculator_lib src/calculator.cpp)
target_include_directories(calculator_lib PUBLIC src/)

# Main executable (optional)
add_executable(calculator_main src/main.cpp)
target_link_libraries(calculator_main PRIVATE calculator_lib)

# Enable testing
include(CTest)
enable_testing()

# Add tests subdirectory
add_subdirectory(tests)
```

### tests/CMakeLists.txt
```cmake
# Fetch Google Test from GitHub
include(FetchContent)
FetchContent_Declare(
  googletest
  GIT_REPOSITORY https://github.com/google/googletest.git
  GIT_TAG        v1.15.2  # Latest stable as of 2025-10-22
)

# For Windows: Prevent overriding the parent project's compiler/linker settings
set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(googletest)

# Test executable
add_executable(calculator_tests test_calculator.cpp)
target_link_libraries(calculator_tests
    PRIVATE
    calculator_lib
    GTest::gtest_main
)

# Discover tests automatically
include(GoogleTest)
gtest_discover_tests(calculator_tests)
```

### Build and Run
```bash
# Configure
cmake -B build -S . -DCMAKE_BUILD_TYPE=Release

# Build
cmake --build build

# Run tests
cd build && ctest --output-on-failure

# Or run directly
./build/tests/calculator_tests
```

---

## Example 2: Modern C++20 with Concepts and Ranges

### math_utils.h
```cpp
#ifndef MATH_UTILS_H
#define MATH_UTILS_H

#include <concepts>
#include <ranges>
#include <vector>
#include <numeric>

// Concept for numeric types
template<typename T>
concept Numeric = std::is_arithmetic_v<T>;

// Template function with concepts
template<Numeric T>
T square(T value) {
    return value * value;
}

// Range-based sum function
template<std::ranges::input_range R>
    requires Numeric<std::ranges::range_value_t<R>>
auto sum(R&& range) {
    using T = std::ranges::range_value_t<R>;
    return std::accumulate(std::ranges::begin(range),
                          std::ranges::end(range),
                          T{});
}

#endif
```

### test_math_utils.cpp
```cpp
#include <gtest/gtest.h>
#include "math_utils.h"
#include <vector>

TEST(MathUtilsTest, SquareInteger) {
    EXPECT_EQ(16, square(4));
    EXPECT_EQ(25, square(-5));
}

TEST(MathUtilsTest, SquareDouble) {
    EXPECT_DOUBLE_EQ(2.25, square(1.5));
}

TEST(MathUtilsTest, SumVector) {
    std::vector<int> numbers = {1, 2, 3, 4, 5};
    EXPECT_EQ(15, sum(numbers));
}

TEST(MathUtilsTest, SumEmptyVector) {
    std::vector<int> empty;
    EXPECT_EQ(0, sum(empty));
}

// Using C++20 ranges
TEST(MathUtilsTest, SumWithRanges) {
    std::vector<int> numbers = {10, 20, 30, 40, 50};
    auto evens = numbers | std::views::filter([](int n) { return n % 20 == 0; });
    EXPECT_EQ(60, sum(std::vector<int>(evens.begin(), evens.end())));
}
```

---

## Example 3: RAII and Smart Pointers

### resource_manager.h
```cpp
#ifndef RESOURCE_MANAGER_H
#define RESOURCE_MANAGER_H

#include <memory>
#include <string>
#include <fstream>

// RAII File Handler
class FileHandler {
private:
    std::unique_ptr<std::ifstream> file_;
    std::string path_;

public:
    explicit FileHandler(const std::string& path);
    ~FileHandler() = default;

    // Prevent copying, allow moving
    FileHandler(const FileHandler&) = delete;
    FileHandler& operator=(const FileHandler&) = delete;
    FileHandler(FileHandler&&) = default;
    FileHandler& operator=(FileHandler&&) = default;

    bool isOpen() const;
    std::string readLine();
};

// Factory function
std::unique_ptr<FileHandler> createFileHandler(const std::string& path);

#endif
```

### resource_manager.cpp
```cpp
#include "resource_manager.h"
#include <stdexcept>

FileHandler::FileHandler(const std::string& path)
    : file_(std::make_unique<std::ifstream>(path)), path_(path) {
    if (!file_->is_open()) {
        throw std::runtime_error("Failed to open file: " + path);
    }
}

bool FileHandler::isOpen() const {
    return file_ && file_->is_open();
}

std::string FileHandler::readLine() {
    std::string line;
    if (file_ && std::getline(*file_, line)) {
        return line;
    }
    return "";
}

std::unique_ptr<FileHandler> createFileHandler(const std::string& path) {
    return std::make_unique<FileHandler>(path);
}
```

### test_resource_manager.cpp
```cpp
#include <gtest/gtest.h>
#include "resource_manager.h"
#include <fstream>

class ResourceManagerTest : public ::testing::Test {
protected:
    const std::string test_file_ = "test_data.txt";

    void SetUp() override {
        // Create test file
        std::ofstream out(test_file_);
        out << "Line 1\n";
        out << "Line 2\n";
        out.close();
    }

    void TearDown() override {
        // Clean up test file
        std::remove(test_file_.c_str());
    }
};

TEST_F(ResourceManagerTest, OpenFileSuccessfully) {
    auto handler = createFileHandler(test_file_);
    EXPECT_TRUE(handler->isOpen());
}

TEST_F(ResourceManagerTest, ThrowsOnInvalidFile) {
    EXPECT_THROW(createFileHandler("nonexistent.txt"), std::runtime_error);
}

TEST_F(ResourceManagerTest, ReadLines) {
    auto handler = createFileHandler(test_file_);
    EXPECT_EQ("Line 1", handler->readLine());
    EXPECT_EQ("Line 2", handler->readLine());
    EXPECT_EQ("", handler->readLine()); // EOF
}

TEST_F(ResourceManagerTest, MoveSemantics) {
    auto handler1 = createFileHandler(test_file_);
    auto handler2 = std::move(handler1);

    EXPECT_TRUE(handler2->isOpen());
    // handler1 is now in a moved-from state
}
```

---

## Example 4: TDD Workflow - String Utilities

### Step 1: RED - Write Failing Test

```cpp
// tests/test_string_utils.cpp
#include <gtest/gtest.h>
#include "string_utils.h"  // Doesn't exist yet!

TEST(StringUtilsTest, TrimLeadingSpaces) {
    EXPECT_EQ("hello", trim("  hello"));
}

TEST(StringUtilsTest, TrimTrailingSpaces) {
    EXPECT_EQ("world", trim("world  "));
}

TEST(StringUtilsTest, TrimBothSides) {
    EXPECT_EQ("test", trim("  test  "));
}

TEST(StringUtilsTest, NoSpacesToTrim) {
    EXPECT_EQ("nochange", trim("nochange"));
}

TEST(StringUtilsTest, EmptyString) {
    EXPECT_EQ("", trim(""));
}

TEST(StringUtilsTest, OnlySpaces) {
    EXPECT_EQ("", trim("    "));
}
```

**Result**: Compilation fails (string_utils.h doesn't exist)

### Step 2: GREEN - Minimal Implementation

```cpp
// src/string_utils.h
#ifndef STRING_UTILS_H
#define STRING_UTILS_H

#include <string>
#include <algorithm>
#include <cctype>

std::string trim(const std::string& str);

#endif
```

```cpp
// src/string_utils.cpp
#include "string_utils.h"

std::string trim(const std::string& str) {
    auto start = std::find_if_not(str.begin(), str.end(),
                                   [](unsigned char ch) { return std::isspace(ch); });
    auto end = std::find_if_not(str.rbegin(), str.rend(),
                                 [](unsigned char ch) { return std::isspace(ch); }).base();

    return (start < end) ? std::string(start, end) : std::string();
}
```

**Result**: All tests pass ✓

### Step 3: REFACTOR - Improve Code Quality

```cpp
// src/string_utils.h (improved)
#ifndef STRING_UTILS_H
#define STRING_UTILS_H

#include <string>
#include <string_view>

namespace utils {

// Trim whitespace from both ends
std::string trim(std::string_view str);

// Individual trim functions for flexibility
std::string trimLeft(std::string_view str);
std::string trimRight(std::string_view str);

} // namespace utils

#endif
```

```cpp
// src/string_utils.cpp (improved)
#include "string_utils.h"
#include <algorithm>
#include <cctype>

namespace utils {

std::string trimLeft(std::string_view str) {
    auto start = std::find_if_not(str.begin(), str.end(),
                                   [](unsigned char ch) { return std::isspace(ch); });
    return std::string(start, str.end());
}

std::string trimRight(std::string_view str) {
    auto end = std::find_if_not(str.rbegin(), str.rend(),
                                 [](unsigned char ch) { return std::isspace(ch); }).base();
    return std::string(str.begin(), end);
}

std::string trim(std::string_view str) {
    return trimLeft(trimRight(str));
}

} // namespace utils
```

**Improvements**:
- Added namespace for organization
- Used `std::string_view` for efficiency (no unnecessary copies)
- Split into focused functions (single responsibility)
- Maintained backward compatibility

**Result**: All tests still pass ✓

---

## Example 5: Mock Testing with Google Mock

### email_service.h (Interface)
```cpp
#ifndef EMAIL_SERVICE_H
#define EMAIL_SERVICE_H

#include <string>

// Abstract interface for email sending
class IEmailService {
public:
    virtual ~IEmailService() = default;
    virtual bool sendEmail(const std::string& to,
                          const std::string& subject,
                          const std::string& body) = 0;
};

// User registration that depends on email service
class UserRegistration {
private:
    IEmailService* emailService_;

public:
    explicit UserRegistration(IEmailService* service)
        : emailService_(service) {}

    bool registerUser(const std::string& email, const std::string& name);
};

#endif
```

### email_service.cpp
```cpp
#include "email_service.h"

bool UserRegistration::registerUser(const std::string& email,
                                    const std::string& name) {
    // Business logic
    if (email.empty() || name.empty()) {
        return false;
    }

    // Send welcome email
    std::string subject = "Welcome " + name;
    std::string body = "Thank you for registering!";

    return emailService_->sendEmail(email, subject, body);
}
```

### test_email_service.cpp
```cpp
#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include "email_service.h"

using ::testing::Return;
using ::testing::_;
using ::testing::Eq;

// Mock implementation
class MockEmailService : public IEmailService {
public:
    MOCK_METHOD(bool, sendEmail,
                (const std::string& to,
                 const std::string& subject,
                 const std::string& body),
                (override));
};

TEST(UserRegistrationTest, RegistersUserAndSendsEmail) {
    MockEmailService mockEmail;
    UserRegistration registration(&mockEmail);

    // Expect sendEmail to be called once and return true
    EXPECT_CALL(mockEmail, sendEmail(
        Eq("user@example.com"),
        Eq("Welcome Alice"),
        Eq("Thank you for registering!")
    )).WillOnce(Return(true));

    bool result = registration.registerUser("user@example.com", "Alice");
    EXPECT_TRUE(result);
}

TEST(UserRegistrationTest, FailsOnEmptyEmail) {
    MockEmailService mockEmail;
    UserRegistration registration(&mockEmail);

    // sendEmail should NOT be called
    EXPECT_CALL(mockEmail, sendEmail(_, _, _)).Times(0);

    bool result = registration.registerUser("", "Alice");
    EXPECT_FALSE(result);
}

TEST(UserRegistrationTest, HandlesEmailServiceFailure) {
    MockEmailService mockEmail;
    UserRegistration registration(&mockEmail);

    // Email service fails
    EXPECT_CALL(mockEmail, sendEmail(_, _, _))
        .WillOnce(Return(false));

    bool result = registration.registerUser("user@example.com", "Bob");
    EXPECT_FALSE(result);
}
```

---

## Common Testing Patterns

### Death Tests (Testing Fatal Errors)
```cpp
TEST(DeathTest, FunctionDiesOnInvalidInput) {
    EXPECT_DEATH(function_that_aborts(), "Expected error message");
}
```

### Typed Tests (Generic Testing)
```cpp
template <typename T>
class TypedTest : public ::testing::Test {};

using NumericTypes = ::testing::Types<int, float, double>;
TYPED_TEST_SUITE(TypedTest, NumericTypes);

TYPED_TEST(TypedTest, WorksWithDifferentTypes) {
    TypeParam value = 42;
    EXPECT_GT(value, 0);
}
```

### Test Timing
```cpp
#include <chrono>

TEST(PerformanceTest, RunsInReasonableTime) {
    auto start = std::chrono::high_resolution_clock::now();

    expensiveOperation();

    auto end = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);

    EXPECT_LT(duration.count(), 100); // Should complete in < 100ms
}
```

---

_For detailed CLI commands and tool versions, see [reference.md](reference.md)_
