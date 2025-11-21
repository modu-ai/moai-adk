# C++ Advanced Patterns & Template Metaprogramming

_Part of moai-lang-cpp Skill - Advanced Techniques_

---

## Template Metaprogramming (TMP)

### Compile-Time Computation

```cpp
// Factorial at compile time
template<int N>
struct Factorial {
    static constexpr int value = N * Factorial<N-1>::value;
};

template<>
struct Factorial<0> {
    static constexpr int value = 1;
};

int main() {
    constexpr int fact5 = Factorial<5>::value;  // 120, computed at compile time
    std::array<int, Factorial<3>::value> arr;   // Array of size 6
}
```

### Type Traits & SFINAE

```cpp
#include <type_traits>

// Detect if type has a serialize() method
template<typename T, typename = void>
struct is_serializable : std::false_type {};

template<typename T>
struct is_serializable<T, std::void_t<decltype(std::declval<T>().serialize())>>
    : std::true_type {};

// Use in overload resolution
template<typename T>
typename std::enable_if<is_serializable<T>::value, std::string>::type
to_string(const T& obj) {
    return obj.serialize();
}

template<typename T>
typename std::enable_if<!is_serializable<T>::value, std::string>::type
to_string(const T& obj) {
    return "Non-serializable object";
}
```

### Variadic Templates

```cpp
// Type-safe printf-like function
template<typename... Args>
void print(const Args&... args) {
    (std::cout << ... << args);  // C++17 fold expression
}

// Example
print("Hello", " ", "World", "\n");

// Tuple-like structure for arbitrary types
template<typename... Ts>
struct Tuple {
    // Empty specialization
};

template<typename T, typename... Ts>
struct Tuple<T, Ts...> {
    T value;
    Tuple<Ts...> rest;
};

// Access element at index
template<size_t I, typename T>
auto& get(T& tuple) {
    if constexpr (I == 0) {
        return tuple.value;
    } else {
        return get<I-1>(tuple.rest);
    }
}
```

### CRTP (Curiously Recurring Template Pattern)

```cpp
// Static polymorphism without virtual functions
template<typename Derived>
class Base {
public:
    void interface() {
        static_cast<Derived*>(this)->implementation();
    }
};

class Derived : public Base<Derived> {
public:
    void implementation() {
        std::cout << "Derived implementation\n";
    }
};

int main() {
    Derived obj;
    obj.interface();  // No virtual call overhead
}
```

### Expression Templates

```cpp
// Lazy evaluation of mathematical expressions
template<typename E>
class Vector {
public:
    double operator[](size_t i) const {
        return static_cast<const E&>(*this)[i];
    }

    size_t size() const {
        return static_cast<const E&>(*this).size();
    }
};

class VectorConstant : public Vector<VectorConstant> {
    double value_;
    size_t size_;

public:
    VectorConstant(double v, size_t s) : value_(v), size_(s) {}

    double operator[](size_t) const { return value_; }
    size_t size() const { return size_; }
};

// Expression template for scalar multiplication
template<typename E>
class VectorScaled : public Vector<VectorScaled<E>> {
    const E& vec_;
    double scale_;

public:
    VectorScaled(const E& v, double s) : vec_(v), scale_(s) {}

    double operator[](size_t i) const { return scale_ * vec_[i]; }
    size_t size() const { return vec_.size(); }
};

template<typename E>
auto operator*(const Vector<E>& vec, double scale) {
    return VectorScaled<E>(static_cast<const E&>(vec), scale);
}
```

---

## Advanced OOP Patterns

### Pimpl (Pointer to Implementation)

```cpp
// Header file
class Widget {
public:
    Widget();
    ~Widget();

    void operation();

    // Non-copyable
    Widget(const Widget&) = delete;
    Widget& operator=(const Widget&) = delete;

private:
    class Impl;
    std::unique_ptr<Impl> pimpl_;
};

// Implementation file
class Widget::Impl {
public:
    void operation() {
        // Hidden implementation
    }

private:
    std::vector<int> data_;
    std::string state_;
};

Widget::Widget() : pimpl_(std::make_unique<Impl>()) {}
Widget::~Widget() = default;  // Compiler-generated
void Widget::operation() { pimpl_->operation(); }
```

### Strategy Pattern

```cpp
class DataProcessor {
    using ProcessingStrategy = std::function<void(const std::vector<int>&)>;
    ProcessingStrategy strategy_;

public:
    void setStrategy(ProcessingStrategy strat) {
        strategy_ = std::move(strat);
    }

    void process(const std::vector<int>& data) {
        if (strategy_) {
            strategy_(data);
        }
    }
};

int main() {
    DataProcessor processor;

    // Sort strategy
    processor.setStrategy([](const std::vector<int>& data) {
        auto sorted = data;
        std::sort(sorted.begin(), sorted.end());
        std::cout << "Sorted: ";
        for (int v : sorted) std::cout << v << " ";
        std::cout << "\n";
    });

    std::vector<int> numbers{5, 2, 8, 1};
    processor.process(numbers);  // Outputs: Sorted: 1 2 5 8

    // Different strategy
    processor.setStrategy([](const std::vector<int>& data) {
        std::cout << "Sum: " << std::accumulate(data.begin(), data.end(), 0) << "\n";
    });

    processor.process(numbers);  // Outputs: Sum: 16
}
```

### Observer Pattern

```cpp
class Subject {
    std::vector<std::weak_ptr<class Observer>> observers_;

public:
    void attach(std::shared_ptr<Observer> obs) {
        observers_.push_back(obs);
    }

    void notify(const std::string& message) {
        for (auto& obs : observers_) {
            if (auto shared = obs.lock()) {
                shared->update(message);
            }
        }
    }
};

class Observer : public std::enable_shared_from_this<Observer> {
public:
    virtual ~Observer() = default;
    virtual void update(const std::string& message) = 0;
};

class ConcreteObserver : public Observer {
    std::string name_;

public:
    explicit ConcreteObserver(const std::string& n) : name_(n) {}

    void update(const std::string& message) override {
        std::cout << name_ << " received: " << message << "\n";
    }
};
```

### Visitor Pattern

```cpp
class Circle;
class Square;
class Triangle;

class ShapeVisitor {
public:
    virtual ~ShapeVisitor() = default;
    virtual void visit(Circle& c) = 0;
    virtual void visit(Square& s) = 0;
    virtual void visit(Triangle& t) = 0;
};

class Shape {
public:
    virtual ~Shape() = default;
    virtual void accept(ShapeVisitor& visitor) = 0;
};

class Circle : public Shape {
    double radius_;
public:
    explicit Circle(double r) : radius_(r) {}
    void accept(ShapeVisitor& visitor) override { visitor.visit(*this); }
    double getRadius() const { return radius_; }
};

class AreaCalculator : public ShapeVisitor {
    double totalArea_{0};
public:
    void visit(Circle& c) override {
        totalArea_ += 3.14159 * c.getRadius() * c.getRadius();
    }
    // ... other visits
    double getTotalArea() const { return totalArea_; }
};
```

---

## Advanced Memory Management

### Move Semantics & Perfect Forwarding

```cpp
// Perfect forwarding factory
template<typename T, typename... Args>
std::unique_ptr<T> makeUnique(Args&&... args) {
    return std::make_unique<T>(std::forward<Args>(args)...);
}

// Demonstrates preserving value categories
class Widget {
public:
    Widget(int value, const std::string& name) {}
};

int main() {
    auto w1 = makeUnique<Widget>(42, "widget");          // Passes by value

    std::string name = "widget2";
    auto w2 = makeUnique<Widget>(43, name);              // Passes lvalue ref
    auto w3 = makeUnique<Widget>(44, std::move(name));   // Passes rvalue ref
}
```

### Custom Allocators

```cpp
template<typename T>
class PoolAllocator {
    std::vector<T*> available_;
    std::vector<T*> all_;

public:
    using value_type = T;

    PoolAllocator(size_t pool_size = 1000) {
        all_.reserve(pool_size);
        for (size_t i = 0; i < pool_size; ++i) {
            auto* ptr = static_cast<T*>(::operator new(sizeof(T)));
            available_.push_back(ptr);
            all_.push_back(ptr);
        }
    }

    ~PoolAllocator() {
        for (auto* ptr : all_) {
            ::operator delete(ptr);
        }
    }

    T* allocate(size_t n) {
        if (n != 1 || available_.empty()) {
            throw std::bad_alloc();
        }
        T* result = available_.back();
        available_.pop_back();
        return result;
    }

    void deallocate(T* p, size_t) {
        available_.push_back(p);
    }
};

// Usage
std::vector<int, PoolAllocator<int>> fast_vector;
```

---

## Concurrency Patterns

### Thread-Safe Singleton

```cpp
class DatabaseConnection {
public:
    static DatabaseConnection& getInstance() {
        static DatabaseConnection instance;  // Thread-safe (C++11)
        return instance;
    }

    void execute(const std::string& query) {
        std::lock_guard<std::mutex> lock(mutex_);
        // Perform operation
    }

private:
    DatabaseConnection() = default;
    ~DatabaseConnection() = default;

    DatabaseConnection(const DatabaseConnection&) = delete;
    DatabaseConnection& operator=(const DatabaseConnection&) = delete;

    mutable std::mutex mutex_;
};
```

### Producer-Consumer Pattern

```cpp
template<typename T>
class Queue {
    std::deque<T> data_;
    mutable std::mutex mutex_;
    std::condition_variable cond_;

public:
    void push(T value) {
        {
            std::lock_guard<std::mutex> lock(mutex_);
            data_.push_back(std::move(value));
        }
        cond_.notify_one();
    }

    bool tryPop(T& value) {
        std::lock_guard<std::mutex> lock(mutex_);
        if (data_.empty()) return false;
        value = std::move(data_.front());
        data_.pop_front();
        return true;
    }

    T waitAndPop() {
        std::unique_lock<std::mutex> lock(mutex_);
        cond_.wait(lock, [this] { return !data_.empty(); });
        T value = std::move(data_.front());
        data_.pop_front();
        return value;
    }
};
```

### Double-Checked Locking

```cpp
class Lazy {
    mutable std::unique_ptr<Expensive> instance_;
    mutable std::mutex mutex_;

public:
    Expensive& get() const {
        // First check without lock
        if (!instance_) {
            std::lock_guard<std::mutex> lock(mutex_);
            // Check again with lock
            if (!instance_) {
                instance_ = std::make_unique<Expensive>();
            }
        }
        return *instance_;
    }

private:
    struct Expensive {
        Expensive() { std::cout << "Expensive construction\n"; }
    };
};
```

---

## Policy-Based Design

```cpp
// Storage policies
struct OnStackStorage {
    char data[256];
};

struct OnHeapStorage {
    std::unique_ptr<char[]> data{std::make_unique<char[]>(256)};
};

// Logging policies
struct NoLogging {
    void log(const std::string&) {}
};

struct ConsoleLogging {
    void log(const std::string& msg) {
        std::cout << "[LOG] " << msg << "\n";
    }
};

// Combine policies
template<typename StoragePolicy, typename LoggingPolicy>
class DataStore : private StoragePolicy, private LoggingPolicy {
public:
    void store(const std::string& data) {
        LoggingPolicy::log("Storing data");
        // Use StoragePolicy::data for storage
    }
};

int main() {
    using FastStore = DataStore<OnStackStorage, NoLogging>;
    using SafeStore = DataStore<OnHeapStorage, ConsoleLogging>;
}
```

---

## Advanced Lambda Features

### Stateful Lambdas with Captures

```cpp
class EventEmitter {
    std::map<std::string, std::function<void(const std::string&)>> handlers_;

public:
    template<typename Handler>
    void on(const std::string& event, Handler&& handler) {
        handlers_[event] = std::forward<Handler>(handler);
    }

    void emit(const std::string& event, const std::string& data) {
        if (auto it = handlers_.find(event); it != handlers_.end()) {
            it->second(data);
        }
    }
};

int main() {
    EventEmitter emitter;

    int counter = 0;
    emitter.on("click", [&counter](const std::string& data) {
        ++counter;
        std::cout << "Click #" << counter << ": " << data << "\n";
    });

    emitter.emit("click", "button1");  // Counter incremented
    emitter.emit("click", "button2");  // Counter incremented again
}
```

### Recursive Lambdas (C++23)

```cpp
auto factorial = [](this auto self, int n) -> int {
    return n <= 1 ? 1 : n * self(n - 1);
};

std::cout << factorial(5);  // Output: 120
```

---

## Zero-Cost Abstractions

### Compile-Time Dispatch

```cpp
template<typename Implementation>
class Interface {
public:
    void doSomething() {
        static_cast<Implementation*>(this)->doSomething_impl();
    }
};

class FastImpl : public Interface<FastImpl> {
public:
    void doSomething_impl() { /* optimized */ }
};

class SlowImpl : public Interface<SlowImpl> {
public:
    void doSomething_impl() { /* general */ }
};

// No virtual overhead - all resolved at compile time
```

### Inline Variables & Constexpr

```cpp
template<typename T>
class Singleton {
public:
    static constexpr std::string_view name = "Singleton";

    static auto& getInstance() {
        static T instance;
        return instance;
    }
};

// Compile-time constant
constexpr auto singleton_name = Singleton<int>::name;
```

---

_See optimization.md for performance tuning techniques_
