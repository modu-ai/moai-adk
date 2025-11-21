---
name: moai-lang-cpp
description: C++23 best practices with Google Test 1.15, clang-format 19, and modern
  C++ standards.
---

## Quick Reference (30 seconds)

# C++ Systems Programming — Enterprise

## References (Latest Documentation)

_Documentation links updated 2025-11-22_

---

---

## Implementation Guide

## What It Does

C++23 best practices with Google Test 1.15, clang-format 19, and modern C++ standards.

**Key capabilities**:
- ✅ Best practices enforcement for language domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-11-22)
- ✅ TDD workflow support
- ✅ Modern C++ features (C++17/20/23)
- ✅ Performance optimization patterns
- ✅ STL algorithms and containers mastery

---

## When to Use

**Automatic triggers**:
- Related code discussions and file patterns
- SPEC implementation (`/alfred:2-run`)
- Code review requests

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new features
- Troubleshoot issues

---

## Tool Version Matrix (2025-11-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **C++** | 23 | Primary | ✅ Current |
| **GCC** | 14.2.0 | Compiler | ✅ Current (2024-11) |
| **Clang** | 19.1.7 | Compiler | ✅ Current (2024-11) |
| **Google Test** | 1.15.0 | Testing | ✅ Current |
| **clang-format** | 19.1.7 | Formatting | ✅ Current |
| **CMake** | 3.30+ | Build system | ✅ Current |
| **Catch2** | 3.7.0 | Testing | ✅ Current |

---

## Modern C++ Features (C++17/20/23)

### C++17 Features

**Structured Bindings**:
```cpp
std::map<std::string, int> user_scores{{"Alice", 95}, {"Bob", 87}};
for (const auto& [name, score] : user_scores) {
    std::cout << name << ": " << score << "\n";
}
```

**std::optional for Null Safety**:
```cpp
std::optional<User> findUser(const std::string& username) {
    auto it = users.find(username);
    if (it != users.end()) {
        return it->second;
    }
    return std::nullopt;
}

// Usage with value_or
auto user = findUser("alice").value_or(User{"guest", 0});
```

**std::variant for Type-Safe Unions**:
```cpp
using Response = std::variant<Success, Error>;

Response processRequest(const Request& req) {
    if (validate(req)) {
        return Success{processData(req)};
    }
    return Error{"Invalid request"};
}

// Pattern matching with visit
std::visit([](auto&& result) {
    using T = std::decay_t<decltype(result)>;
    if constexpr (std::is_same_v<T, Success>) {
        std::cout << "Success: " << result.data << "\n";
    } else {
        std::cerr << "Error: " << result.message << "\n";
    }
}, response);
```

### C++20 Features

**Concepts for Compile-Time Constraints**:
```cpp
template<typename T>
concept Numeric = std::is_arithmetic_v<T>;

template<Numeric T>
T add(T a, T b) {
    return a + b;
}

// Custom concept
template<typename T>
concept Serializable = requires(T t) {
    { t.serialize() } -> std::same_as<std::string>;
    { T::deserialize(std::string{}) } -> std::same_as<T>;
};
```

**Ranges for Functional Programming**:
```cpp
#include <ranges>

std::vector<int> numbers{1, 2, 3, 4, 5};

// Filter even numbers and square them
auto result = numbers 
    | std::views::filter([](int n) { return n % 2 == 0; })
    | std::views::transform([](int n) { return n * n; });

for (int val : result) {
    std::cout << val << " ";  // Output: 4 16
}
```

**Coroutines for Async Programming**:
```cpp
#include <coroutine>

struct Task {
    struct promise_type {
        Task get_return_object() { return {}; }
        std::suspend_never initial_suspend() { return {}; }
        std::suspend_never final_suspend() noexcept { return {}; }
        void return_void() {}
        void unhandled_exception() {}
    };
};

Task async_operation() {
    std::cout << "Starting\n";
    co_await std::suspend_always{};
    std::cout << "Resuming\n";
}
```

**Modules (Replacement for Headers)**:
```cpp
// math_utils.cppm
export module math_utils;

export namespace math {
    int add(int a, int b) { return a + b; }
    int multiply(int a, int b) { return a * b; }
}

// main.cpp
import math_utils;

int main() {
    return math::add(3, 4);
}
```

### C++23 Features

**std::expected for Error Handling**:
```cpp
#include <expected>

std::expected<User, Error> loadUser(const std::string& id) {
    if (id.empty()) {
        return std::unexpected(Error{"Empty user ID"});
    }
    // Load user logic
    return User{id, "John Doe"};
}

// Usage
auto result = loadUser("user123");
if (result) {
    std::cout << "Loaded: " << result->name << "\n";
} else {
    std::cerr << "Error: " << result.error().message << "\n";
}
```

**Deducing this for Recursive Lambdas**:
```cpp
auto factorial = [](this auto self, int n) -> int {
    return n <= 1 ? 1 : n * self(n - 1);
};

std::cout << factorial(5);  // Output: 120
```

---

## Advanced OOP Patterns

### CRTP (Curiously Recurring Template Pattern)

```cpp
// Static polymorphism without virtual functions
template<typename Derived>
class Shape {
public:
    double area() const {
        return static_cast<const Derived*>(this)->area_impl();
    }
};

class Circle : public Shape<Circle> {
    double radius_;
public:
    explicit Circle(double r) : radius_(r) {}
    double area_impl() const { return 3.14159 * radius_ * radius_; }
};

class Rectangle : public Shape<Rectangle> {
    double width_, height_;
public:
    Rectangle(double w, double h) : width_(w), height_(h) {}
    double area_impl() const { return width_ * height_; }
};

// Zero-cost abstraction
template<typename T>
void print_area(const Shape<T>& shape) {
    std::cout << shape.area() << "\n";
}
```

### PImpl Idiom (Pointer to Implementation)

```cpp
// widget.h
class Widget {
public:
    Widget();
    ~Widget();
    
    void do_something();
    
private:
    class Impl;  // Forward declaration
    std::unique_ptr<Impl> pImpl;
};

// widget.cpp
class Widget::Impl {
public:
    void do_something() {
        // Implementation details hidden from header
    }
};

Widget::Widget() : pImpl(std::make_unique<Impl>()) {}
Widget::~Widget() = default;  // Required for unique_ptr<Impl>

void Widget::do_something() {
    pImpl->do_something();
}
```

### Strategy Pattern with std::function

```cpp
class PaymentProcessor {
    using PaymentStrategy = std::function<bool(double)>;
    PaymentStrategy strategy_;
    
public:
    void set_strategy(PaymentStrategy strategy) {
        strategy_ = std::move(strategy);
    }
    
    bool process(double amount) {
        return strategy_ ? strategy_(amount) : false;
    }
};

// Usage
PaymentProcessor processor;

// Credit card strategy
processor.set_strategy([](double amount) {
    std::cout << "Processing $" << amount << " via credit card\n";
    return true;
});

// PayPal strategy
processor.set_strategy([](double amount) {
    std::cout << "Processing $" << amount << " via PayPal\n";
    return amount < 1000.0;  // PayPal limit
});
```

---

## STL Deep Dive

### Custom Allocators for Memory Control

```cpp
template<typename T>
class PoolAllocator {
    std::vector<T*> pool_;
    size_t next_index_{0};
    
public:
    using value_type = T;
    
    PoolAllocator(size_t pool_size = 1000) {
        pool_.reserve(pool_size);
        for (size_t i = 0; i < pool_size; ++i) {
            pool_.push_back(static_cast<T*>(::operator new(sizeof(T))));
        }
    }
    
    T* allocate(size_t n) {
        if (n != 1) throw std::bad_alloc();
        if (next_index_ >= pool_.size()) throw std::bad_alloc();
        return pool_[next_index_++];
    }
    
    void deallocate(T* p, size_t) noexcept {
        // Return to pool (simplified)
    }
};

// Usage
std::vector<int, PoolAllocator<int>> fast_vector;
```

### Smart Pointers Best Practices

```cpp
// Factory function returning unique_ptr
std::unique_ptr<Resource> create_resource() {
    return std::make_unique<Resource>("config.json");
}

// Ownership transfer
void process(std::unique_ptr<Resource> res) {
    // Takes ownership
}

// Shared ownership with weak observers
class Observer {
    std::weak_ptr<Subject> subject_;
    
public:
    void update() {
        if (auto subject = subject_.lock()) {
            // Subject still exists
            subject->notify();
        }
    }
};
```

### Algorithm Composition

```cpp
#include <algorithm>
#include <numeric>

std::vector<int> data{1, 2, 3, 4, 5};

// Parallel algorithms (C++17)
std::sort(std::execution::par, data.begin(), data.end());

// Transform-reduce pattern
int sum_of_squares = std::transform_reduce(
    std::execution::par,
    data.begin(), data.end(),
    0,  // Initial value
    std::plus<>(),  // Reduce operation
    [](int x) { return x * x; }  // Transform operation
);
```

---

## Memory Management & RAII

### Move Semantics Optimization

```cpp
class Buffer {
    char* data_;
    size_t size_;
    
public:
    // Constructor
    Buffer(size_t size) : data_(new char[size]), size_(size) {}
    
    // Destructor
    ~Buffer() { delete[] data_; }
    
    // Move constructor
    Buffer(Buffer&& other) noexcept 
        : data_(other.data_), size_(other.size_) {
        other.data_ = nullptr;
        other.size_ = 0;
    }
    
    // Move assignment
    Buffer& operator=(Buffer&& other) noexcept {
        if (this != &other) {
            delete[] data_;
            data_ = other.data_;
            size_ = other.size_;
            other.data_ = nullptr;
            other.size_ = 0;
        }
        return *this;
    }
    
    // Delete copy operations
    Buffer(const Buffer&) = delete;
    Buffer& operator=(const Buffer&) = delete;
};
```

### Perfect Forwarding

```cpp
template<typename T, typename... Args>
std::unique_ptr<T> make_unique_wrapper(Args&&... args) {
    return std::make_unique<T>(std::forward<Args>(args)...);
}

// Preserves value categories
struct Widget {
    Widget(int x, const std::string& s) { /* ... */ }
};

auto widget = make_unique_wrapper<Widget>(42, "hello");
```

### RAII for Resource Management

```cpp
class FileGuard {
    FILE* file_;
    
public:
    explicit FileGuard(const char* filename) 
        : file_(fopen(filename, "r")) {
        if (!file_) {
            throw std::runtime_error("Failed to open file");
        }
    }
    
    ~FileGuard() {
        if (file_) {
            fclose(file_);
        }
    }
    
    FILE* get() const { return file_; }
    
    // Non-copyable
    FileGuard(const FileGuard&) = delete;
    FileGuard& operator=(const FileGuard&) = delete;
};

// Usage - automatic cleanup even on exceptions
void process_file(const char* filename) {
    FileGuard file(filename);
    // Use file.get()
    // Automatically closed when leaving scope
}
```

---

## Performance Optimization

### Cache-Friendly Data Structures

```cpp
// Structure of Arrays (SoA) for better cache locality
struct ParticlesSoA {
    std::vector<float> x, y, z;  // Positions
    std::vector<float> vx, vy, vz;  // Velocities
    
    void update(float dt) {
        for (size_t i = 0; i < x.size(); ++i) {
            x[i] += vx[i] * dt;  // Sequential memory access
            y[i] += vy[i] * dt;
            z[i] += vz[i] * dt;
        }
    }
};

// Array of Structures (AoS) - less cache-friendly
struct Particle {
    float x, y, z, vx, vy, vz;
};
std::vector<Particle> particles;  // Interleaved data
```

### Compiler Optimization Hints

```cpp
// Branch prediction hints
#define likely(x)   __builtin_expect(!!(x), 1)
#define unlikely(x) __builtin_expect(!!(x), 0)

void process(int value) {
    if (likely(value > 0)) {
        // Hot path
    } else {
        // Cold path
    }
}

// Force inline for performance-critical functions
[[gnu::always_inline]] inline int fast_add(int a, int b) {
    return a + b;
}

// No inline for large functions
[[gnu::noinline]] void large_function() {
    // ...
}
```

### Profiling with Perf

```bash
# Compile with debug symbols
g++ -O3 -g -std=c++23 main.cpp -o app

# Profile CPU usage
perf record -g ./app

# Generate report
perf report

# Annotate source code with performance data
perf annotate
```

---

## Testing with Google Test & Catch2

### Google Test Patterns

```cpp
#include <gtest/gtest.h>

// Test fixture for setup/teardown
class DatabaseTest : public ::testing::Test {
protected:
    void SetUp() override {
        db = std::make_unique<Database>("test.db");
    }
    
    void TearDown() override {
        db->close();
    }
    
    std::unique_ptr<Database> db;
};

TEST_F(DatabaseTest, InsertAndRetrieve) {
    User user{"alice", 30};
    db->insert(user);
    
    auto result = db->find("alice");
    ASSERT_TRUE(result.has_value());
    EXPECT_EQ(result->name, "alice");
    EXPECT_EQ(result->age, 30);
}

// Parameterized tests
class MathTest : public ::testing::TestWithParam<std::pair<int, int>> {};

TEST_P(MathTest, Addition) {
    auto [a, b] = GetParam();
    EXPECT_EQ(add(a, b), a + b);
}

INSTANTIATE_TEST_SUITE_P(
    AdditionTests,
    MathTest,
    ::testing::Values(
        std::make_pair(1, 2),
        std::make_pair(5, 10),
        std::make_pair(-1, 1)
    )
);
```

### Catch2 BDD Style

```cpp
#include <catch2/catch_test_macros.hpp>

SCENARIO("User authentication") {
    GIVEN("A registered user") {
        User user{"alice", "password123"};
        
        WHEN("They provide correct credentials") {
            auto result = authenticate(user.name, "password123");
            
            THEN("Authentication succeeds") {
                REQUIRE(result.success);
                REQUIRE(result.token.has_value());
            }
        }
        
        WHEN("They provide incorrect password") {
            auto result = authenticate(user.name, "wrong");
            
            THEN("Authentication fails") {
                REQUIRE_FALSE(result.success);
                REQUIRE_FALSE(result.token.has_value());
            }
        }
    }
}
```

---

## Build System (CMake 3.28+)

### Modern CMake Configuration

```cmake
cmake_minimum_required(VERSION 3.28)
project(MyApp VERSION 1.0.0 LANGUAGES CXX)

# C++23 standard
set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# FetchContent for dependencies
include(FetchContent)

FetchContent_Declare(
    googletest
    GIT_REPOSITORY https://github.com/google/googletest.git
    GIT_TAG v1.15.0
)
FetchContent_MakeAvailable(googletest)

# Library target
add_library(mylib STATIC
    src/core.cpp
    src/utils.cpp
)

target_include_directories(mylib PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)

# Executable target
add_executable(myapp src/main.cpp)
target_link_libraries(myapp PRIVATE mylib)

# Test target
enable_testing()
add_executable(tests tests/test_core.cpp)
target_link_libraries(tests PRIVATE mylib GTest::gtest_main)
gtest_discover_tests(tests)

# Compiler warnings
target_compile_options(mylib PRIVATE
    $<$<CXX_COMPILER_ID:GNU,Clang>:-Wall -Wextra -Wpedantic>
    $<$<CXX_COMPILER_ID:MSVC>:/W4>
)
```

---

## Inputs

- Language-specific source directories
- Configuration files
- Test suites and sample data

## Outputs

- Test/lint execution plan
- TRUST 5 review checkpoints
- Migration guidance

## Failure Modes

- When required tools are not installed
- When dependencies are missing
- When test coverage falls below 85%

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates

---

## Changelog

- **v3.1.0** (2025-11-22): Updated to C++23, removed C++14 compatibility, added Modules practical examples, GCC 14.2, Clang 19.1, CMake 3.30+ patterns
- **v3.0.0** (2025-11-22): Massive expansion with C++17/20/23 features, advanced patterns, STL deep dive, performance optimization
- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)
- `moai-essentials-perf` (performance optimization)

---

## Best Practices

✅ **DO**:
- Use modern C++ features (concepts, ranges, coroutines)
- Apply RAII for all resource management
- Prefer smart pointers over raw pointers
- Use move semantics for large objects
- Write const-correct code
- Maintain test coverage ≥85%
- Use CMake for build configuration
- Profile before optimizing

❌ **DON'T**:
- Use raw `new`/`delete` without RAII wrapper
- Ignore compiler warnings
- Mix C and C++ idioms
- Use `using namespace std` in headers
- Implement manual memory management
- Use deprecated features (auto_ptr, bind1st)
- Skip const correctness
- Optimize without profiling data

---

## Advanced Patterns




---

## Context7 Integration

### Related Libraries & Tools
- [C++](/cplusplus): Modern C++ standards
- [Boost](/boostorg/boost): Comprehensive C++ libraries
- [Qt](/qt/qt): Cross-platform framework
- [Google Test](/google/googletest): Testing framework
- [CMake](/Kitware/CMake): Build system generator

### Official Documentation
- [Documentation](https://en.cppreference.com/w/)
- [API Reference](https://en.cppreference.com/w/cpp/header)
- [Google Test Guide](https://google.github.io/googletest/)
- [CMake Documentation](https://cmake.org/documentation/)

### Version-Specific Guides
Latest stable version: C++23
- [C++23 Release](https://en.cppreference.com/w/cpp/23)
- [C++20 Features](https://en.cppreference.com/w/cpp/20)
- [Migration Guide](https://en.cppreference.com/w/cpp/language/history)
