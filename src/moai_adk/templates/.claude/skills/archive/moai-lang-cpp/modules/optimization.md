# C++ Performance Optimization & Tuning

_Part of moai-lang-cpp Skill - Performance Strategies_

---

## Cache-Friendly Data Structures

### Array of Structures vs. Structure of Arrays

```cpp
// ❌ AoS (Array of Structures) - poor cache locality
struct Particle {
    float x, y, z;      // Position
    float vx, vy, vz;   // Velocity
    float mass;
    float temperature;
};
std::vector<Particle> particles(1000000);

void update_particles_aos(std::vector<Particle>& particles, float dt) {
    for (auto& p : particles) {
        p.x += p.vx * dt;
        p.y += p.vy * dt;
        p.z += p.vz * dt;
        // Cache misses due to scattered memory access
    }
}

// ✅ SoA (Structure of Arrays) - excellent cache locality
struct ParticlesSoA {
    std::vector<float> x, y, z;        // Positions
    std::vector<float> vx, vy, vz;     // Velocities
    std::vector<float> mass;
    std::vector<float> temperature;
    size_t count{0};

    void update(float dt) {
        for (size_t i = 0; i < count; ++i) {
            x[i] += vx[i] * dt;
            y[i] += vy[i] * dt;
            z[i] += vz[i] * dt;
            // Sequential memory access - cache-friendly
        }
    }
};
```

### Cache-Line Optimization

```cpp
// Align data to cache line boundaries (usually 64 bytes)
alignas(64) struct CacheAlignedData {
    int value;
    // Padding ensures next object starts on new cache line
};

// Prevent false sharing in multi-threaded code
struct ThreadLocalCounter {
    alignas(64) std::atomic<long long> value{0};
    // Each thread's counter on separate cache line
};

std::vector<ThreadLocalCounter> counters(num_threads);

void worker(int thread_id) {
    for (int i = 0; i < 1000000; ++i) {
        counters[thread_id].value.fetch_add(1);
        // No contention between threads
    }
}
```

---

## Compiler Optimization Hints

### Branch Prediction Optimization

```cpp
// ❌ Branch misprediction
void slow_sort(std::vector<int>& data) {
    for (int i = 0; i < data.size(); ++i) {
        if (data[i] < 128)  // Often mispredicted
            sum += data[i];
    }
}

// ✅ Predictable branches
void fast_sort(std::vector<int>& data) {
    // Pre-process to improve branch prediction
    std::vector<int> less_128, greater_equal_128;
    for (int v : data) {
        if (v < 128)
            less_128.push_back(v);
        else
            greater_equal_128.push_back(v);
    }
}

// Use compiler hints
#define likely(x)   __builtin_expect(!!(x), 1)
#define unlikely(x) __builtin_expect(!!(x), 0)

void process_with_hints(const std::vector<int>& data) {
    for (int v : data) {
        if (likely(v > 0)) {
            // Hot path
        } else if (unlikely(v < 0)) {
            // Error handling - rare
        }
    }
}
```

### Inline Optimization

```cpp
// ✅ Aggressive inlining for performance-critical code
[[gnu::always_inline]] inline int fast_add(int a, int b) {
    return a + b;
}

// ✅ Prevent inlining for large functions
[[gnu::noinline]] void complex_algorithm() {
    // Large function - avoid code bloat
}

// ✅ Virtual function optimization
class Base {
public:
    [[gnu::always_inline]] virtual void hot_path() {
        // Performance-critical virtual function
    }
};
```

### SIMD Vectorization

```cpp
#include <immintrin.h>

// Manual SIMD for vector operations
void add_vectors_simd(float* c, const float* a, const float* b, int n) {
    for (int i = 0; i < n; i += 8) {
        __m256 va = _mm256_loadu_ps(&a[i]);
        __m256 vb = _mm256_loadu_ps(&b[i]);
        __m256 vc = _mm256_add_ps(va, vb);
        _mm256_storeu_ps(&c[i], vc);
    }
}

// Compiler auto-vectorization hints
void dot_product(float* result, const float* a, const float* b, int n) {
    *result = 0;
    // Mark loop as vectorizable
    #pragma omp simd reduction(+:*result)
    for (int i = 0; i < n; ++i) {
        *result += a[i] * b[i];
    }
}
```

---

## Memory Optimization

### Efficient String Handling

```cpp
// ❌ Inefficient - creates temporary strings
std::string slow_concat() {
    std::string result = "";
    for (int i = 0; i < 1000; ++i) {
        result += "item_";
        result += std::to_string(i);
        result += "\n";
    }
    return result;
}

// ✅ Efficient - pre-allocates memory
std::string fast_concat() {
    std::string result;
    result.reserve(50000);  // Estimate final size
    for (int i = 0; i < 1000; ++i) {
        result += "item_";
        result += std::to_string(i);
        result += "\n";
    }
    return result;
}

// ✅ Best - uses stringstream
std::string best_concat() {
    std::ostringstream oss;
    for (int i = 0; i < 1000; ++i) {
        oss << "item_" << i << "\n";
    }
    return oss.str();
}
```

### Move Semantics & Return Value Optimization (RVO)

```cpp
// RVO - compiler optimizes away copy
class LargeBuffer {
    std::vector<char> data{1000000};
};

LargeBuffer create_buffer() {
    LargeBuffer buf;
    // Compiler applies RVO - no copy constructor called
    return buf;
}

auto buf = create_buffer();  // Efficient

// Explicit move for clarity
std::vector<int> combine_vectors(std::vector<int> a, std::vector<int> b) {
    a.insert(a.end(),
             std::make_move_iterator(b.begin()),
             std::make_move_iterator(b.end()));
    return a;  // RVO or move
}
```

### Pool Allocators for High-Frequency Operations

```cpp
template<typename T, size_t PoolSize = 1000>
class ObjectPool {
    std::array<T, PoolSize> objects_;
    std::array<bool, PoolSize> in_use_{};
    size_t next_free_{0};

public:
    T* allocate() {
        for (size_t i = 0; i < PoolSize; ++i) {
            if (!in_use_[i]) {
                in_use_[i] = true;
                return &objects_[i];
            }
        }
        throw std::bad_alloc();
    }

    void deallocate(T* ptr) {
        size_t index = ptr - objects_.data();
        in_use_[index] = false;
    }
};

// Usage
class Connection {
    // ... connection data
};

ObjectPool<Connection> connection_pool;

auto conn = connection_pool.allocate();
// ... use connection
connection_pool.deallocate(conn);
```

---

## Algorithm Optimization

### Parallel STL Algorithms

```cpp
#include <execution>
#include <algorithm>

std::vector<int> data(1000000);

// ✅ Parallel sorting (C++17)
std::sort(std::execution::par, data.begin(), data.end());

// ✅ Parallel for_each
std::for_each(std::execution::par,
              data.begin(), data.end(),
              [](int& val) { val = val * 2; });

// ✅ Transform-reduce pattern
int sum_of_squares = std::transform_reduce(
    std::execution::par,
    data.begin(), data.end(),
    0,
    std::plus<>(),
    [](int x) { return x * x; }
);
```

### Early Exit Optimizations

```cpp
// ❌ Searches entire collection
bool contains_negative(const std::vector<int>& vec) {
    for (const auto& v : vec) {
        if (v < 0) return true;
    }
    return false;
}

// ✅ Uses std::any_of for early exit
bool has_negative(const std::vector<int>& vec) {
    return std::any_of(vec.begin(), vec.end(),
                       [](int v) { return v < 0; });
}

// ✅ Custom iterator-based approach
std::vector<int> find_all_negatives(const std::vector<int>& vec) {
    std::vector<int> result;
    auto it = std::find_if(vec.begin(), vec.end(),
                          [](int v) { return v < 0; });
    if (it != vec.end()) {
        result.push_back(*it);  // Early exit on first match
    }
    return result;
}
```

---

## Compilation & Link-Time Optimization

### Link-Time Code Generation (LTO)

```bash
# Compile with LTO enabled
g++ -O3 -flto main.cpp lib.cpp -o app

# Parallel LTO
g++ -O3 -flto=auto main.cpp lib.cpp -o app
```

### Profile-Guided Optimization (PGO)

```bash
# Instrumented build
g++ -O2 -fprofile-generate main.cpp -o app_instr

# Run with typical workload
./app_instr < typical_input.txt

# Optimized build using profiling data
g++ -O2 -fprofile-use main.cpp -o app_optimized
```

---

## Runtime Profiling

### Benchmark with Google Benchmark

```cpp
#include <benchmark/benchmark.h>

static void BenchmarkVectorCreation(benchmark::State& state) {
    for (auto _ : state) {
        std::vector<int> vec(1000);
        benchmark::DoNotOptimize(vec);
    }
}
BENCHMARK(BenchmarkVectorCreation);

static void BenchmarkMemcpy(benchmark::State& state) {
    std::vector<char> src(10000), dst(10000);
    for (auto _ : state) {
        std::memcpy(dst.data(), src.data(), src.size());
    }
}
BENCHMARK(BenchmarkMemcpy);

BENCHMARK_MAIN();
```

### Profiling with Perf (Linux)

```bash
# Compile with debug symbols
g++ -O3 -g main.cpp -o app

# Profile CPU usage
perf record -g ./app

# Generate flamegraph
perf report

# Annotate with source code
perf annotate --source
```

### Memory Profiling

```bash
# Compile with Address Sanitizer
g++ -fsanitize=address -g main.cpp -o app

# Run with memory profiling
ASAN_OPTIONS=detect_leaks=1 ./app

# Valgrind memory profiling
valgrind --leak-check=full --show-leak-kinds=all ./app
```

---

## Caching Strategies

### Function Result Caching

```cpp
template<typename Func, typename... Args>
class CachedFunction {
    std::map<std::tuple<Args...>, std::invoke_result_t<Func, Args...>> cache_;
    Func func_;

public:
    explicit CachedFunction(Func f) : func_(f) {}

    auto operator()(const Args&... args) {
        auto key = std::make_tuple(args...);
        auto it = cache_.find(key);
        if (it != cache_.end()) {
            return it->second;
        }
        auto result = func_(args...);
        cache_[key] = result;
        return result;
    }
};

// Usage
int fibonacci(int n) {
    if (n <= 1) return n;
    return fibonacci(n-1) + fibonacci(n-2);
}

auto cached_fib = CachedFunction(fibonacci);
cached_fib(40);  // Computed once, cached thereafter
```

### LRU Cache Implementation

```cpp
template<typename Key, typename Value, size_t MaxSize = 100>
class LRUCache {
    std::list<std::pair<Key, Value>> list_;
    std::unordered_map<Key, typename std::list<std::pair<Key, Value>>::iterator> map_;

public:
    Value get(const Key& key) {
        auto it = map_.find(key);
        if (it == map_.end()) {
            throw std::out_of_range("Key not found");
        }
        // Move to front (most recently used)
        list_.splice(list_.begin(), list_, it->second);
        return it->second->second;
    }

    void put(const Key& key, const Value& value) {
        auto it = map_.find(key);
        if (it != map_.end()) {
            // Update existing
            it->second->second = value;
            list_.splice(list_.begin(), list_, it->second);
            return;
        }

        // Add new entry
        if (map_.size() >= MaxSize) {
            // Evict least recently used (back of list)
            auto last = list_.back();
            list_.pop_back();
            map_.erase(last.first);
        }

        list_.emplace_front(key, value);
        map_[key] = list_.begin();
    }
};
```

---

## Lazy Loading & Deferred Computation

### Lazy Evaluation

```cpp
template<typename T>
class Lazy {
    std::function<T()> initializer_;
    mutable std::optional<T> value_;

public:
    explicit Lazy(std::function<T()> init) : initializer_(init) {}

    const T& get() const {
        if (!value_) {
            value_ = initializer_();
        }
        return *value_;
    }

    const T& operator*() const { return get(); }
    const T* operator->() const { return &get(); }
};

// Usage
Lazy<std::vector<int>> expensive_data([]() {
    std::cout << "Computing...\n";
    return std::vector<int>(1000000);
});

// Not computed yet
auto& data = expensive_data.get();  // Computed on first access
```

---

## Memory Layout Optimization

### Struct Packing & Alignment

```cpp
// ❌ Inefficient packing - 24 bytes due to alignment
struct BadLayout {
    char a;         // 1 byte + 7 padding
    double b;       // 8 bytes
    int c;          // 4 bytes + 4 padding
};

// ✅ Efficient packing - 16 bytes
struct GoodLayout {
    double b;       // 8 bytes
    int c;          // 4 bytes
    char a;         // 1 byte + 3 padding
};

// ✅ Bit fields for compact storage
struct CompactFlags {
    bool flag1 : 1;
    bool flag2 : 1;
    bool flag3 : 1;
    // ... 5 more flags
    // Total: 1 byte for 8 flags
};
```

---

## Best Practices Summary

1. **Profile First**: Use perf, valgrind, or benchmark before optimizing
2. **Cache-Friendly**: Use SoA over AoS, align data to cache lines
3. **Move Semantics**: Leverage move constructors to avoid copies
4. **Algorithm Choice**: Use parallel STL when appropriate
5. **Compiler Hints**: Use likely/unlikely for branch prediction
6. **Memory Pooling**: Pre-allocate for high-frequency operations
7. **Lazy Evaluation**: Defer expensive computations
8. **SIMD**: Use vectorization for computationally intensive code
9. **Const-Correctness**: Enables compiler optimizations
10. **Compile with -O3**: Always use optimization flags in production

---

_See advanced-patterns.md for template metaprogramming and design patterns_
