# moai-lang-cpp - Working Examples

_Last updated: 2025-11-22_

## Quick Start - Project Setup

### CMake Configuration

```cmake
cmake_minimum_required(VERSION 3.28)
project(MyApp VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include(FetchContent)
FetchContent_Declare(
    googletest
    GIT_REPOSITORY https://github.com/google/googletest.git
    GIT_TAG v1.15.0
)
FetchContent_MakeAvailable(googletest)

add_library(mylib src/core.cpp)
add_executable(myapp src/main.cpp)
target_link_libraries(myapp PRIVATE mylib)
```

### Clang-Format Configuration

```yaml
# .clang-format
BasedOnStyle: LLVM
IndentWidth: 4
ColumnLimit: 100
AllowShortFunctionsOnASingleLine: None
AlwaysBreakTemplateDeclarations: Yes
```

---

## Example 1: Modern C++23 - Structured Bindings & Optional

### Problem: Safe User Lookup

```cpp
#include <optional>
#include <string>
#include <map>

struct User {
    std::string name;
    int age;
    std::string email;
};

class UserRepository {
    std::map<std::string, User> users_;

public:
    std::optional<User> findUserByName(const std::string& name) const {
        auto it = users_.find(name);
        if (it != users_.end()) {
            return it->second;
        }
        return std::nullopt;
    }

    void addUser(const User& user) {
        users_[user.name] = user;
    }
};

int main() {
    UserRepository repo;
    repo.addUser(User{"Alice", 30, "alice@example.com"});
    repo.addUser(User{"Bob", 25, "bob@example.com"});

    // Structured bindings with optional
    if (auto user = repo.findUserByName("Alice")) {
        auto [name, age, email] = user.value();
        std::cout << name << " is " << age << " years old\n";
    }
}
```

---

## Example 2: Concepts - Compile-Time Type Constraints

### Problem: Generic Container Processing

```cpp
#include <concepts>
#include <vector>
#include <iostream>

// Concept: types that have a size() method
template<typename T>
concept Sized = requires(T t) {
    { t.size() } -> std::convertible_to<std::size_t>;
};

// Concept: Numeric types
template<typename T>
concept Numeric = std::is_arithmetic_v<T>;

template<Sized T>
void printSize(const T& container) {
    std::cout << "Size: " << container.size() << "\n";
}

template<Numeric T>
T calculateSum(T a, T b) {
    return a + b;
}

// Custom concept with multiple requirements
template<typename T>
concept Serializable = requires(T t) {
    { t.serialize() } -> std::same_as<std::string>;
    { T::deserialize("") } -> std::same_as<T>;
};

class Document {
public:
    std::string serialize() const {
        return "document_data";
    }

    static Document deserialize(const std::string& data) {
        return Document{};
    }
};

int main() {
    std::vector<int> vec{1, 2, 3, 4, 5};
    printSize(vec);  // Works: vector has size()

    int sum = calculateSum(10, 20);  // Works: int is Numeric

    Document doc;
    // doc is Serializable
}
```

---

## Example 3: Ranges - Functional Programming

### Problem: Data Pipeline Processing

```cpp
#include <ranges>
#include <vector>
#include <iostream>

struct Product {
    std::string name;
    double price;
    int quantity;
};

int main() {
    std::vector<Product> products{
        {"Laptop", 999.99, 5},
        {"Mouse", 29.99, 100},
        {"Keyboard", 79.99, 50},
        {"Monitor", 299.99, 10}
    };

    // Ranges-based data pipeline
    auto expensive_items = products
        | std::views::filter([](const auto& p) { return p.price > 100; })
        | std::views::transform([](const auto& p) {
            return std::pair{p.name, p.price * p.quantity};
        });

    std::cout << "Expensive inventory value:\n";
    for (const auto& [name, total] : expensive_items) {
        std::cout << name << ": $" << total << "\n";
    }

    // Chaining multiple operations
    auto summary = products
        | std::views::filter([](const auto& p) { return p.quantity > 0; })
        | std::views::transform([](const auto& p) { return p.price; })
        | std::views::take(3);

    for (double price : summary) {
        std::cout << "$" << price << " ";
    }
}
```

---

## Example 4: Coroutines - Async Operations

### Problem: Asynchronous Data Fetching Simulation

```cpp
#include <coroutine>
#include <iostream>
#include <string>

struct Task {
    struct promise_type {
        std::string result;

        Task get_return_object() {
            return Task{std::coroutine_handle<promise_type>::from_promise(*this)};
        }

        std::suspend_never initial_suspend() { return {}; }
        std::suspend_never final_suspend() noexcept { return {}; }

        void return_value(std::string value) {
            result = std::move(value);
        }

        void unhandled_exception() {}
    };

    std::coroutine_handle<promise_type> handle;

    ~Task() {
        if (handle) handle.destroy();
    }

    std::string get_result() const {
        return handle.promise().result;
    }
};

struct Awaitable {
    std::string value;

    bool await_ready() const { return false; }
    void await_suspend(std::coroutine_handle<>) {}
    std::string await_resume() const { return value; }
};

Task fetchUserData(const std::string& userId) {
    std::cout << "Fetching data for " << userId << "...\n";

    Awaitable fetch{userId + "_data"};
    auto result = co_await fetch;

    std::cout << "Got result: " << result << "\n";
    co_return result;
}

int main() {
    auto task = fetchUserData("user123");
    std::cout << "Final: " << task.get_result() << "\n";
}
```

---

## Example 5: RAII - Resource Management

### Problem: Database Connection Management

```cpp
#include <iostream>
#include <memory>
#include <stdexcept>

class DatabaseConnection {
    int connection_id_;

public:
    explicit DatabaseConnection(int id) : connection_id_(id) {
        std::cout << "Connecting to DB [" << id << "]\n";
    }

    ~DatabaseConnection() {
        std::cout << "Disconnecting from DB [" << connection_id_ << "]\n";
    }

    void execute(const std::string& query) {
        std::cout << "Executing: " << query << "\n";
    }

    // Non-copyable
    DatabaseConnection(const DatabaseConnection&) = delete;
    DatabaseConnection& operator=(const DatabaseConnection&) = delete;
};

class Transaction {
    std::unique_ptr<DatabaseConnection> db_;
    bool committed_{false};

public:
    Transaction(int db_id)
        : db_(std::make_unique<DatabaseConnection>(db_id)) {
        std::cout << "Transaction started\n";
    }

    ~Transaction() {
        if (!committed_) {
            std::cout << "Transaction rolled back (auto-cleanup)\n";
        }
    }

    void executeQuery(const std::string& query) {
        db_->execute(query);
    }

    void commit() {
        std::cout << "Transaction committed\n";
        committed_ = true;
    }
};

int main() {
    {
        Transaction txn(1);
        txn.executeQuery("INSERT INTO users VALUES (...)");

        // Even if exception occurs, destructor ensures cleanup
        txn.commit();
    }  // Auto cleanup here

    std::cout << "\nAnother transaction:\n";
    {
        Transaction txn2(2);
        txn2.executeQuery("UPDATE users SET ...");
        // No commit called - auto rollback in destructor
    }  // Auto rollback here
}
```

---

## Example 6: CRTP - Static Polymorphism

### Problem: Shape Area Calculation Without Virtual Functions

```cpp
#include <iostream>
#include <cmath>
#include <vector>

// Base class using CRTP
template<typename Derived>
class Shape {
public:
    double area() const {
        return static_cast<const Derived*>(this)->area_impl();
    }

    void printArea() const {
        std::cout << "Area: " << area() << "\n";
    }
};

class Circle : public Shape<Circle> {
    double radius_;

public:
    explicit Circle(double r) : radius_(r) {}

    double area_impl() const {
        return 3.14159 * radius_ * radius_;
    }
};

class Rectangle : public Shape<Rectangle> {
    double width_, height_;

public:
    Rectangle(double w, double h) : width_(w), height_(h) {}

    double area_impl() const {
        return width_ * height_;
    }
};

class Triangle : public Shape<Triangle> {
    double base_, height_;

public:
    Triangle(double b, double h) : base_(b), height_(h) {}

    double area_impl() const {
        return 0.5 * base_ * height_;
    }
};

template<typename T>
void processShape(const Shape<T>& shape) {
    shape.printArea();  // No virtual function call overhead
}

int main() {
    Circle circle(5.0);
    Rectangle rect(4.0, 6.0);
    Triangle tri(3.0, 4.0);

    processShape(circle);   // Area: 78.5398
    processShape(rect);     // Area: 24
    processShape(tri);      // Area: 6
}
```

---

## Example 7: Smart Pointers & Memory Safety

### Problem: Ownership Transfer

```cpp
#include <memory>
#include <iostream>

class Resource {
    std::string name_;

public:
    explicit Resource(const std::string& name) : name_(name) {
        std::cout << "Resource [" << name << "] created\n";
    }

    ~Resource() {
        std::cout << "Resource [" << name_ << "] destroyed\n";
    }

    void use() const {
        std::cout << "Using " << name_ << "\n";
    }
};

// Factory function
std::unique_ptr<Resource> createResource(const std::string& name) {
    return std::make_unique<Resource>(name);
}

class Manager {
    std::vector<std::shared_ptr<Resource>> resources_;

public:
    void addResource(std::shared_ptr<Resource> res) {
        resources_.push_back(res);
    }

    void useAll() {
        for (const auto& res : resources_) {
            if (res) {
                res->use();
            }
        }
    }
};

int main() {
    // Unique ownership
    auto res1 = createResource("File");
    res1->use();
    // res1 automatically destroyed here

    std::cout << "\n";

    // Shared ownership
    Manager manager;

    {
        auto res2 = std::make_shared<Resource>("Database");
        auto res3 = std::make_shared<Resource>("Cache");

        manager.addResource(res2);
        manager.addResource(res3);
        manager.useAll();

        // res2, res3 still alive (held by manager)
    }

    std::cout << "\nResources still alive (held by manager)\n";
    manager.useAll();

    // All resources destroyed when manager goes out of scope
}
```

---

## Example 8: Google Test - TDD Workflow

### Problem: Unit Testing a Calculator

```cpp
// calculator.h
#ifndef CALCULATOR_H
#define CALCULATOR_H

class Calculator {
public:
    double add(double a, double b) const { return a + b; }
    double subtract(double a, double b) const { return a - b; }
    double multiply(double a, double b) const { return a * b; }
    double divide(double a, double b) const;
};

#endif

// calculator.cpp
#include "calculator.h"
#include <stdexcept>

double Calculator::divide(double a, double b) const {
    if (b == 0.0) {
        throw std::invalid_argument("Division by zero");
    }
    return a / b;
}

// calculator_test.cpp
#include <gtest/gtest.h>
#include "calculator.h"

class CalculatorTest : public ::testing::Test {
protected:
    Calculator calc;
};

TEST_F(CalculatorTest, Addition) {
    EXPECT_DOUBLE_EQ(5.0, calc.add(2.0, 3.0));
}

TEST_F(CalculatorTest, Subtraction) {
    EXPECT_DOUBLE_EQ(1.0, calc.subtract(3.0, 2.0));
}

TEST_F(CalculatorTest, Multiplication) {
    EXPECT_DOUBLE_EQ(6.0, calc.multiply(2.0, 3.0));
}

TEST_F(CalculatorTest, DivisionByZero) {
    EXPECT_THROW(calc.divide(1.0, 0.0), std::invalid_argument);
}

// Parameterized test
class CalculatorParametrizedTest : public ::testing::TestWithParam<std::tuple<double, double, double>> {};

TEST_P(CalculatorParametrizedTest, AdditionVariants) {
    auto [a, b, expected] = GetParam();
    EXPECT_DOUBLE_EQ(expected, calc.add(a, b));
}

INSTANTIATE_TEST_SUITE_P(
    AdditionTests,
    CalculatorParametrizedTest,
    ::testing::Values(
        std::make_tuple(1.0, 2.0, 3.0),
        std::make_tuple(0.0, 0.0, 0.0),
        std::make_tuple(-1.0, 1.0, 0.0),
        std::make_tuple(0.5, 0.5, 1.0)
    )
);
```

---

## Example 9: Move Semantics - Performance Optimization

### Problem: Efficient Buffer Transfer

```cpp
#include <iostream>
#include <vector>

class Buffer {
    std::vector<char> data_;
    size_t size_;

    static int copy_count;
    static int move_count;

public:
    explicit Buffer(size_t size) : data_(size, 0), size_(size) {}

    // Copy constructor (expensive)
    Buffer(const Buffer& other) : data_(other.data_), size_(other.size_) {
        copy_count++;
        std::cout << "Copy constructor called (count: " << copy_count << ")\n";
    }

    // Copy assignment
    Buffer& operator=(const Buffer& other) {
        if (this != &other) {
            data_ = other.data_;
            size_ = other.size_;
            copy_count++;
            std::cout << "Copy assignment called\n";
        }
        return *this;
    }

    // Move constructor (cheap)
    Buffer(Buffer&& other) noexcept
        : data_(std::move(other.data_)), size_(other.size_) {
        move_count++;
        std::cout << "Move constructor called (count: " << move_count << ")\n";
        other.size_ = 0;
    }

    // Move assignment
    Buffer& operator=(Buffer&& other) noexcept {
        if (this != &other) {
            data_ = std::move(other.data_);
            size_ = other.size_;
            other.size_ = 0;
            move_count++;
            std::cout << "Move assignment called\n";
        }
        return *this;
    }

    static void printStats() {
        std::cout << "Total copies: " << copy_count
                  << ", moves: " << move_count << "\n";
    }
};

int Buffer::copy_count = 0;
int Buffer::move_count = 0;

Buffer createLargeBuffer() {
    return Buffer(1000);  // Calls move constructor (RVO)
}

int main() {
    std::cout << "=== Without std::move ===\n";
    Buffer buf1(100);
    Buffer buf2 = buf1;  // Copy constructor
    Buffer::printStats();

    std::cout << "\n=== With std::move ===\n";
    Buffer buf3(100);
    Buffer buf4 = std::move(buf3);  // Move constructor
    Buffer::printStats();

    std::cout << "\n=== Function return (RVO) ===\n";
    Buffer buf5 = createLargeBuffer();  // Move elision
    Buffer::printStats();
}
```

---

## Example 10: Build and Test Commands

### Compilation

```bash
# Simple compilation
g++ -std=c++23 -O3 -Wall -Wextra main.cpp -o app

# With CMake
mkdir build
cd build
cmake -DCMAKE_BUILD_TYPE=Release ..
cmake --build . --parallel $(nproc)
```

### Testing

```bash
# Run unit tests
ctest --output-on-failure

# Run with verbose output
ctest --verbose

# Run specific test
ctest -R "CalculatorTest" --verbose
```

### Profiling

```bash
# Compile with profiling
g++ -std=c++23 -O2 -g -pg main.cpp -o app

# Profile execution
./app
gprof ./app gmon.out | head -20
```

---

_See modules/advanced-patterns.md for template metaprogramming, modules/optimization.md for performance tuning_
