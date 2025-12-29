---
name: moai-lang-cpp
description: Modern C++ (C++23/C++20) development specialist covering RAII, smart pointers, concepts, ranges, modules, and CMake. Use when developing high-performance applications, games, system software, or embedded systems.
version: 1.0.0
updated: 2025-12-07
status: active
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

## Quick Reference (30 seconds)

Modern C++ (C++23/C++20) Development Specialist - RAII, smart pointers, concepts, ranges, modules, and CMake.

Auto-Triggers: `.cpp`, `.hpp`, `.h`, `CMakeLists.txt`, `vcpkg.json`, `conanfile.txt`, modern C++ discussions

Core Capabilities:
- C++23 Features: std::expected, std::print, std::generator, deducing this
- C++20 Features: Concepts, Ranges, Modules, Coroutines, std::format
- Memory Safety: RAII, Rule of 5, smart pointers (unique_ptr, shared_ptr, weak_ptr)
- STL: Containers, Algorithms, Iterators, std::span, std::string_view
- Build Systems: CMake 3.28+, FetchContent, presets
- Concurrency: std::thread, std::jthread, std::async, atomics, std::latch/barrier
- Testing: Google Test, Catch2
- Package Management: vcpkg, Conan 2.0

### Quick Patterns

Smart Pointer Factory:
```cpp
#include <memory>

class Widget {
public:
    static auto create(int value) -> std::unique_ptr<Widget> {
        return std::make_unique<Widget>(value);
    }
    explicit Widget(int v) : value_(v) {}
private:
    int value_;
};

auto widget = Widget::create(42);
```

Concepts Constraint:
```cpp
#include <concepts>

template<typename T>
concept Numeric = std::integral<T> || std::floating_point<T>;

template<Numeric T>
auto square(T value) -> T {
    return value * value;
}
```

Ranges Pipeline:
```cpp
#include <ranges>
#include <vector>

auto result = std::views::iota(1, 100)
    | std::views::filter([](int n) { return n % 2 == 0; })
    | std::views::transform([](int n) { return n * n; })
    | std::views::take(10);
```

---

## Implementation Guide (5 minutes)

### C++23 New Features

std::expected for Error Handling:
```cpp
#include <expected>
#include <string>

enum class ParseError { InvalidFormat, OutOfRange };

auto parse_int(std::string_view str) -> std::expected<int, ParseError> {
    try {
        int value = std::stoi(std::string(str));
        return value;
    } catch (const std::invalid_argument&) {
        return std::unexpected(ParseError::InvalidFormat);
    } catch (const std::out_of_range&) {
        return std::unexpected(ParseError::OutOfRange);
    }
}

// Usage
auto result = parse_int("42");
if (result) {
    std::println("Value: {}", *result);
} else {
    std::println("Error occurred");
}
```

std::print for Type-Safe Output:
```cpp
#include <print>

int main() {
    std::println("Hello, {}!", "World");
    std::print("Value: {}, Hex: {:#x}\n", 255, 255);
    std::println("Formatted: {:.2f}", 3.14159);
}
```

Deducing This (Explicit Object Parameter):
```cpp
class Builder {
    std::string data_;
public:
    template<typename Self>
    auto append(this Self&& self, std::string_view s) -> Self&& {
        self.data_ += s;
        return std::forward<Self>(self);
    }

    auto build() const -> std::string { return data_; }
};

auto result = Builder{}.append("Hello").append(" World").build();
```

### C++20 Features

Concepts and Constraints:
```cpp
#include <concepts>

template<typename T>
concept Hashable = requires(T a) {
    { std::hash<T>{}(a) } -> std::convertible_to<std::size_t>;
};

template<typename Container>
    requires std::ranges::range<Container>
auto make_default_filled(std::size_t count) -> Container {
    Container c;
    c.resize(count);
    return c;
}

void process(std::integral auto value) {
    // value is constrained to integral types
}
```

Modules (C++20):
```cpp
// math.cppm - Module interface
export module math;

export namespace math {
    template<typename T>
    constexpr auto add(T a, T b) -> T {
        return a + b;
    }
}

// main.cpp - Consumer
import math;
import std;

int main() {
    auto result = math::add(10, 20);
    std::println("Result: {}", result);
}
```

Ranges Library:
```cpp
#include <ranges>
#include <vector>

struct Person {
    std::string name;
    int age;
};

void process_people(std::vector<Person>& people) {
    auto adults = people
        | std::views::filter([](const Person& p) { return p.age >= 18; })
        | std::views::transform([](const Person& p) { return p.name; });

    for (const auto& name : adults) {
        std::println("{}", name);
    }

    std::ranges::sort(people, {}, &Person::age);
}
```

### RAII and Resource Management

Rule of Five:
```cpp
class Resource {
    int* data_;
    std::size_t size_;

public:
    explicit Resource(std::size_t size)
        : data_(new int[size]), size_(size) {}

    ~Resource() { delete[] data_; }

    Resource(const Resource& other)
        : data_(new int[other.size_]), size_(other.size_) {
        std::copy(other.data_, other.data_ + size_, data_);
    }

    auto operator=(const Resource& other) -> Resource& {
        if (this != &other) {
            Resource temp(other);
            swap(temp);
        }
        return *this;
    }

    Resource(Resource&& other) noexcept
        : data_(std::exchange(other.data_, nullptr))
        , size_(std::exchange(other.size_, 0)) {}

    auto operator=(Resource&& other) noexcept -> Resource& {
        if (this != &other) {
            delete[] data_;
            data_ = std::exchange(other.data_, nullptr);
            size_ = std::exchange(other.size_, 0);
        }
        return *this;
    }

    void swap(Resource& other) noexcept {
        std::swap(data_, other.data_);
        std::swap(size_, other.size_);
    }
};
```

Smart Pointer Patterns:
```cpp
#include <memory>

class Connection {
public:
    static auto create(std::string_view host) -> std::unique_ptr<Connection> {
        return std::make_unique<Connection>(host);
    }
    explicit Connection(std::string_view host) : host_(host) {}
private:
    std::string host_;
};

class Node : public std::enable_shared_from_this<Node> {
public:
    std::vector<std::shared_ptr<Node>> children;
    std::weak_ptr<Node> parent;  // Weak to break cycle

    void add_child(std::shared_ptr<Node> child) {
        child->parent = weak_from_this();
        children.push_back(std::move(child));
    }
};
```

### CMake Modern Patterns

CMakeLists.txt (C++23 Project):
```cmake
cmake_minimum_required(VERSION 3.28)
project(MyProject VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

add_compile_options(
    $<$<CXX_COMPILER_ID:GNU,Clang>:-Wall -Wextra -Wpedantic>
    $<$<CXX_COMPILER_ID:MSVC>:/W4>
)

include(FetchContent)

FetchContent_Declare(fmt
    GIT_REPOSITORY https://github.com/fmtlib/fmt
    GIT_TAG 10.2.1)
FetchContent_Declare(googletest
    GIT_REPOSITORY https://github.com/google/googletest
    GIT_TAG v1.14.0)
FetchContent_MakeAvailable(fmt googletest)

add_library(mylib STATIC src/core.cpp)
target_include_directories(mylib PUBLIC include)
target_link_libraries(mylib PUBLIC fmt::fmt)

add_executable(myapp src/main.cpp)
target_link_libraries(myapp PRIVATE mylib)

enable_testing()
add_executable(mylib_tests tests/core_test.cpp)
target_link_libraries(mylib_tests PRIVATE mylib GTest::gtest_main)
include(GoogleTest)
gtest_discover_tests(mylib_tests)
```

### Concurrency

std::jthread and Stop Tokens:
```cpp
#include <thread>
#include <stop_token>

void worker(std::stop_token stoken) {
    while (!stoken.stop_requested()) {
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
}

int main() {
    std::jthread worker_thread(worker);
    std::this_thread::sleep_for(std::chrono::seconds(1));
    worker_thread.request_stop();
}
```

Synchronization Primitives:
```cpp
#include <latch>
#include <barrier>
#include <semaphore>

void parallel_init(std::latch& ready, int id) {
    ready.count_down();
}

void parallel_compute(std::barrier<>& sync, int id) {
    for (int i = 0; i < 10; ++i) {
        sync.arrive_and_wait();
    }
}

std::counting_semaphore<4> pool(4);
void limited_resource() {
    pool.acquire();
    pool.release();
}
```

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:
- Template metaprogramming patterns
- Advanced concurrency (thread pools, lock-free data structures)
- Memory management and custom allocators
- Performance optimization (SIMD, cache-friendly patterns)
- Production patterns (dependency injection, factories)
- Extended testing with Google Test and Catch2

See:
- [Advanced Patterns](modules/advanced-patterns.md) - Complete advanced patterns guide

---

## Context7 Library Mappings

```
/microsoft/vcpkg - Package manager
/conan-io/conan - Conan package manager
/google/googletest - Google Test framework
/catchorg/Catch2 - Catch2 testing framework
/fmtlib/fmt - {fmt} formatting library
/nlohmann/json - JSON for Modern C++
/gabime/spdlog - Fast logging library
```

---

## Works Well With

- `moai-lang-rust` - Systems programming comparison and interop
- `moai-domain-backend` - Backend service architecture
- `moai-workflow-testing` - TDD and testing strategies
- `moai-essentials-debug` - Debugging and profiling
- `moai-foundation-quality` - TRUST 5 quality principles

---

## Troubleshooting

Version Check:
```bash
g++ --version  # GCC 13+ for C++23
clang++ --version  # Clang 17+ for C++23
cmake --version  # CMake 3.28+
```

Common Compilation Flags:
```bash
g++ -std=c++23 -Wall -Wextra -Wpedantic -O2 main.cpp -o main
g++ -std=c++23 -fsanitize=address,undefined -g main.cpp -o main
```

vcpkg Integration:
```bash
git clone https://github.com/microsoft/vcpkg
./vcpkg/bootstrap-vcpkg.sh
./vcpkg/vcpkg install fmt nlohmann-json gtest
cmake -B build -DCMAKE_TOOLCHAIN_FILE=./vcpkg/scripts/buildsystems/vcpkg.cmake
```

---

Last Updated: 2025-12-07
Status: Active (v1.0.0)
