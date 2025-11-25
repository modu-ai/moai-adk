# C Performance Optimization Guide

**Focus**: Compiler optimization, memory efficiency, profiling
**Standards**: C17 with GCC/Clang
**Last Updated**: 2025-11-22

---

## Optimization Level 1: Compiler Flags

**Development build** (fast compilation, debugging):
```bash
gcc -std=c17 -O0 -g -Wall -Wextra -fno-omit-frame-pointer
```

**Production build** (optimized, no debugging):
```bash
gcc -std=c17 -O3 -DNDEBUG -Wall -Wextra -march=native -flto
```

**Performance analysis**:
```bash
gcc -std=c17 -O2 -fprofile-generate -o prog prog.c
./prog < test_input    # Generate profile data
gcc -std=c17 -O2 -fprofile-use -o prog prog.c  # Rebuild with profile
```

---

## Optimization Level 2: Memory Layout and Cache

**Structure packing for cache efficiency**:
```c
// Bad: Wasted space, cache misses
typedef struct {
    char a;      // 1 byte
    // 7 bytes padding
    long b;      // 8 bytes
    char c;      // 1 byte
    // 7 bytes padding
} Inefficient;  // 32 bytes total

// Good: Aligned fields, cache friendly
typedef struct {
    long b;      // 8 bytes, naturally aligned
    char a;      // 1 byte
    char c;      // 1 byte
    // 6 bytes padding (but contiguous)
} Efficient;    // 16 bytes total

// Best: Use attributes (GCC/Clang)
typedef struct {
    long b;
    char a;
    char c;
} __attribute__((packed)) PackedStruct;  // 10 bytes exact

// Verify sizes
void check_sizes(void) {
    printf("Inefficient: %zu\n", sizeof(Inefficient));
    printf("Efficient: %zu\n", sizeof(Efficient));
    printf("PackedStruct: %zu\n", sizeof(PackedStruct));
}
```

**Cache-friendly access patterns**:
```c
// Poor cache locality
void poor_access(int matrix[1000][1000]) {
    for (int j = 0; j < 1000; j++) {
        for (int i = 0; i < 1000; i++) {
            // Accesses column-wise - cache misses
            matrix[i][j]++;
        }
    }
}

// Good cache locality
void good_access(int matrix[1000][1000]) {
    for (int i = 0; i < 1000; i++) {
        for (int j = 0; j < 1000; j++) {
            // Row-wise access - cache hits
            matrix[i][j]++;
        }
    }
}
```

---

## Optimization Level 3: Profiling with Valgrind

**Memcheck for memory leaks**:
```bash
gcc -std=c17 -g -O0 -o prog prog.c
valgrind --leak-check=full --show-leak-kinds=all ./prog
```

**Cachegrind for cache analysis**:
```bash
valgrind --tool=cachegrind ./prog
cg_annotate cachegrind.out.* prog.c
```

**Callgrind for profiling**:
```bash
valgrind --tool=callgrind ./prog
callgrind_annotate callgrind.out.* | head -50
```

---

## Optimization Level 4: Fast Path Optimization

**Branch prediction hints**:
```c
#ifndef likely
#define likely(x)   __builtin_expect(!!(x), 1)
#define unlikely(x) __builtin_expect(!!(x), 0)
#endif

int process_input(int value) {
    // Common case first
    if (likely(value > 0)) {
        return value * 2;
    }

    // Rare case
    if (unlikely(value < -100)) {
        return -999;
    }

    return 0;
}

// Hot path optimization
void hot_loop(int *arr, int n) {
    // Minimize work in inner loop
    for (int i = 0; i < n; i++) {
        arr[i] *= 2;  // Simple operation
    }
}

// Cold path
void setup_expensive_state(void) {
    __attribute__((cold))
    static void setup(void) {
        // Expensive initialization
    }
    setup();
}
```

---

## Optimization Level 5: Loop Unrolling

**Manual unrolling**:
```c
// Original (slow)
void sum_naive(int *arr, int n) {
    int sum = 0;
    for (int i = 0; i < n; i++) {
        sum += arr[i];
    }
}

// Manually unrolled (2x speedup)
void sum_unrolled_2(int *arr, int n) {
    int sum = 0;
    int i;
    for (i = 0; i < n - 1; i += 2) {
        sum += arr[i] + arr[i + 1];
    }
    if (i < n) sum += arr[i];
}

// Compiler-assisted unrolling
void sum_compiler(int *arr, int n) {
    int sum = 0;
    #pragma GCC unroll 4
    for (int i = 0; i < n; i++) {
        sum += arr[i];
    }
}
```

---

## Optimization Level 6: SIMD/Vectorization

**SSE2 vectorization**:
```c
#include <emmintrin.h>

// SIMD sum of 32 integers
int simd_sum(const int *arr, int n) {
    __m128i sum = _mm_setzero_si128();

    for (int i = 0; i < n; i += 4) {
        __m128i v = _mm_loadu_si128((__m128i *)&arr[i]);
        sum = _mm_add_epi32(sum, v);
    }

    // Horizontal sum
    __m128i s1 = _mm_shuffle_epi32(sum, _MM_SHUFFLE(2, 3, 0, 1));
    sum = _mm_add_epi32(sum, s1);
    s1 = _mm_shuffle_epi32(sum, _MM_SHUFFLE(1, 0, 3, 2));
    sum = _mm_add_epi32(sum, s1);

    return _mm_cvtsi128_si32(sum);
}

// Compiler auto-vectorization hint
void vector_add(float *a, float *b, float *c, int n) {
    #pragma omp simd
    for (int i = 0; i < n; i++) {
        c[i] = a[i] + b[i];
    }
}
```

---

## Optimization Level 7: Reducing Function Call Overhead

**Inline functions**:
```c
// Inline small functions
static inline int max(int a, int b) {
    return a > b ? a : b;
}

// Force inline
static inline __attribute__((always_inline)) int min(int a, int b) {
    return a < b ? a : b;
}

// Prevent inlining
__attribute__((noinline)) void expensive_function(void) {
    // Complex operation
}

// Tail call optimization hint
int factorial_tail(int n, int acc) {
    if (n <= 1) return acc;
    return factorial_tail(n - 1, acc * n);  // Will be optimized to loop
}
```

---

## Optimization Level 8: Memory Allocation Patterns

**Stack allocation for small objects**:
```c
#include <alloca.h>

void process_small_array(int n) {
    // Stack allocation (fast, scope-limited)
    int *arr = alloca(n * sizeof(int));
    // Use arr...
}  // Automatic cleanup

// Reusable buffer pool
typedef struct {
    char *buffers[100];
    int available;
} BufferPool;

char *pool_acquire(BufferPool *pool) {
    if (pool->available > 0) {
        return pool->buffers[--pool->available];
    }
    return malloc(4096);
}

void pool_release(BufferPool *pool, char *buffer) {
    if (pool->available < 100) {
        pool->buffers[pool->available++] = buffer;
    } else {
        free(buffer);
    }
}
```

---

## Optimization Level 9: Benchmarking

**Simple benchmark framework**:
```c
#include <time.h>

#define BENCHMARK_START() \
    clock_t start = clock();

#define BENCHMARK_END(label) do { \
    clock_t end = clock(); \
    double elapsed = (double)(end - start) / CLOCKS_PER_SEC; \
    printf("%s: %.6f seconds\n", label, elapsed); \
} while(0)

void benchmark_functions(void) {
    int arr[100000];
    for (int i = 0; i < 100000; i++) arr[i] = i;

    BENCHMARK_START();
    int sum1 = sum_naive(arr, 100000);
    BENCHMARK_END("Naive sum");

    BENCHMARK_START();
    int sum2 = sum_unrolled_2(arr, 100000);
    BENCHMARK_END("Unrolled sum");

    printf("Results: %d vs %d\n", sum1, sum2);
}
```

---

## Optimization Level 10: Common Performance Pitfalls

**Avoid**:
```c
// DON'T: Repeated strlen in loop
char *str = "Hello world";
for (int i = 0; i < strlen(str); i++) {  // O(n²)
    printf("%c\n", str[i]);
}

// DO: Cache length
int len = strlen(str);
for (int i = 0; i < len; i++) {  // O(n)
    printf("%c\n", str[i]);
}

// DON'T: Unnecessary pointer indirection
int arr[1000];
int *p = arr;
for (int i = 0; i < 1000; i++) {
    p[i]++;  // Extra indirection
}

// DO: Direct access when possible
for (int i = 0; i < 1000; i++) {
    arr[i]++;
}

// DON'T: Malloc every iteration
void slow_loop(int n) {
    for (int i = 0; i < n; i++) {
        int *p = malloc(sizeof(int));  // Expensive!
        *p = i;
        free(p);
    }
}

// DO: Allocate once
void fast_loop(int n) {
    int *arr = malloc(n * sizeof(int));
    for (int i = 0; i < n; i++) {
        arr[i] = i;
    }
    free(arr);
}
```

---

## Compiler Optimization Options

**GCC/Clang flags**:
```bash
-O0  # No optimization (debugging)
-O1  # Basic optimization
-O2  # Balanced optimization (recommended)
-O3  # Aggressive optimization
-Os  # Optimize for size
-Ofast # Unsafe optimizations

# Vectorization
-march=native       # CPU-specific optimizations
-ftree-vectorize    # Enable loop vectorization

# Link-time optimization
-flto              # Full program optimization

# Profiling
-fprofile-generate # Generate profile data
-fprofile-use      # Use profile data
```

---

**Last Updated**: 2025-11-22 | **Production Ready**
