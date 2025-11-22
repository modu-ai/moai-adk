# Rust Performance Optimization — Memory, Async, Systems

_Last updated: 2025-11-22_

## Memory Optimization

### Stack vs Heap Allocation

```rust
// Stack allocation (fast, limited size)
fn stack_example() {
    let buffer = [0u8; 1024];  // 1KB on stack - very fast
    process_buffer(&buffer);
}

// Heap allocation (slower, unlimited)
fn heap_example() {
    let buffer = vec![0u8; 1024];  // 1KB on heap - slower
    process_buffer(&buffer);
}

// Rule of thumb: Keep hot data on stack
// Use heap for large allocations, collections, owned data

// Stack-allocated array for small fixed sizes
fn small_buffer() {
    let buffer: [u8; 64] = [0; 64];  // Stack
    // Fast access, no allocation
}

// Heap-allocated for dynamic sizes
fn large_buffer(size: usize) {
    let buffer = vec![0u8; size];  // Heap
    // Scalable, slightly slower
}
```

### Zero-Cost Abstractions

```rust
// Monomorphization - compiler generates specialized code
fn generic_add<T: std::ops::Add<Output=T>>(a: T, b: T) -> T {
    a + b
}

// Compiler creates:
// fn generic_add_i32(a: i32, b: i32) -> i32 { a + b }
// fn generic_add_f64(a: f64, b: f64) -> f64 { a + b }
// No runtime overhead - specialized assembly

// Inlining hints
#[inline]
fn small_function() -> u32 {
    42
}

#[inline(never)]
fn large_function() {
    // Compiler won't inline
}

// Compile-time evaluation
const BUFFER_SIZE: usize = compute_size();
const fn compute_size() -> usize {
    1024 * 1024  // Evaluated at compile time
}

// Generic enum without overhead
enum Result<T, E> {
    Ok(T),
    Err(E),
}
// Zero overhead - no boxing, discriminant is minimal
```

### String Handling

```rust
// ❌ INEFFICIENT: String concatenation with +
fn inefficient_concat() -> String {
    let mut result = String::new();
    result = result + "Hello";   // Allocates
    result = result + " ";       // Allocates again
    result = result + "World";   // Allocates again
    result
}

// ✅ EFFICIENT: String builder
fn efficient_concat() -> String {
    let mut result = String::with_capacity(20);
    result.push_str("Hello");
    result.push(' ');
    result.push_str("World");
    result
}

// ✅ OPTIMAL: Format macro
fn optimal_concat() -> String {
    format!("Hello World")  // Single allocation
}

// Avoid allocations with references
fn process_str(s: &str) {
    // Takes reference, no allocation
    println!("{}", s);
}

// Copy small strings on stack (small string optimization)
use smallvec::SmallVec;

let mut vec: SmallVec<[u32; 4]> = SmallVec::new();
// Stores up to 4 items on stack, spills to heap if needed
```

## Async Performance Optimization

### Tokio Tuning

```rust
// Configure Tokio runtime for workload type
use tokio::runtime;

// CPU-bound work
let rt = runtime::Builder::new_multi_thread()
    .worker_threads(num_cpus::get())  // Match CPU cores
    .build()
    .unwrap();

// I/O-bound work (default is good for this)
let rt = tokio::runtime::Runtime::new()?;

// Single-threaded for light workloads
let rt = runtime::Builder::new_current_thread()
    .build()
    .unwrap();

// Spawn blocking tasks correctly
async fn async_with_blocking() {
    let result = tokio::task::spawn_blocking(|| {
        // CPU-intensive work here
        // Won't block other tasks
        heavy_computation()
    })
    .await
    .unwrap();
}
```

### Buffer Management

```rust
// Pre-allocate buffers to avoid reallocations
fn process_with_buffer() {
    let mut buffer = Vec::with_capacity(1024 * 1024);

    for chunk in read_chunks() {
        buffer.clear();  // Reuse allocation
        buffer.extend_from_slice(&chunk);
        process(&buffer);
    }
}

// Ring buffer for fixed-size message processing
use heapless::RingBuffer;

let mut ring: RingBuffer<u8, 256> = RingBuffer::new();
// Fixed memory, no allocation

// ArrayVec for small collections
use smallvec::SmallVec;
let mut items: SmallVec<[Item; 10]> = SmallVec::new();
// Stack allocation for small sizes, heap for large
```

## Zero-Cost Abstraction Patterns

### Iterator Efficiency

```rust
// ❌ INEFFICIENT: Collecting intermediate results
fn inefficient() -> usize {
    let vec = vec![1, 2, 3, 4, 5];
    let doubled: Vec<_> = vec.iter().map(|x| x * 2).collect();
    let evens: Vec<_> = doubled.iter().filter(|x| x % 2 == 0).collect();
    evens.len()
}

// ✅ EFFICIENT: Iterator chains (lazy evaluation)
fn efficient() -> usize {
    let vec = vec![1, 2, 3, 4, 5];
    vec.iter()
        .map(|x| x * 2)
        .filter(|x| x % 2 == 0)
        .count()
    // No intermediate allocations
}

// Compiler optimizations: Loop unrolling, SIMD
fn simd_candidate() {
    let a = [1, 2, 3, 4];
    let b = [5, 6, 7, 8];
    let result: Vec<_> = a.iter()
        .zip(b.iter())
        .map(|(x, y)| x + y)
        .collect();
    // Compiler may vectorize this
}
```

## Profiling & Diagnostics

### Flamegraph Profiling

```bash
# Install flamegraph
cargo install flamegraph

# Profile with flamegraph
cargo flamegraph --bin myapp -- arg1 arg2

# Generates flamegraph.svg showing hot functions
```

### Benchmarking

```rust
#[cfg(test)]
mod benches {
    use super::*;

    #[bench]
    fn bench_algorithm(b: &mut Bencher) {
        b.iter(|| algorithm_under_test());
    }

    #[bench]
    fn bench_with_setup(b: &mut Bencher) {
        b.iter_with_setup(
            || vec![1; 1000],
            |v| process(&v)
        );
    }
}

// Run with: cargo bench
```

### Memory Profiling with Valgrind

```bash
# Install valgrind
sudo apt-get install valgrind

# Profile memory
valgrind --leak-check=full --show-leak-kinds=all \
    ./target/debug/myapp

# Massif for memory usage over time
valgrind --tool=massif ./target/debug/myapp
```

## CPU-Bound Optimization

### Parallel Processing

```rust
use rayon::prelude::*;

// Sequential
fn sequential() {
    let data = (0..1_000_000).collect::<Vec<_>>();
    let result: Vec<_> = data.iter()
        .map(|x| x * x)
        .collect();
}

// Parallel (automatic work distribution)
fn parallel() {
    let data = (0..1_000_000).collect::<Vec<_>>();
    let result: Vec<_> = data.par_iter()
        .map(|x| x * x)
        .collect();
    // Automatically uses all CPU cores
}

// Custom thread pool for control
fn custom_parallel() {
    let pool = rayon::ThreadPoolBuilder::new()
        .num_threads(4)
        .build()
        .unwrap();

    pool.install(|| {
        let result: Vec<_> = (0..1_000_000).par_iter()
            .map(|x| x * x)
            .collect();
    });
}
```

### SIMD Operations

```rust
use packed_simd::SimdVector;

// Auto-vectorization hint
#[inline]
fn vectorizable_sum(data: &[f32]) -> f32 {
    data.iter().sum()
    // Compiler may vectorize with SIMD
}

// Explicit SIMD (nightly Rust)
#[cfg(target_arch = "x86_64")]
unsafe fn fast_sum(data: &[f32]) -> f32 {
    // Manual SIMD implementation
    // Much faster for large data
}
```

## Best Practices

- [ ] Use stack allocation for small, fixed-size data
- [ ] Pre-allocate vectors with capacity to avoid reallocations
- [ ] Chain iterators instead of collecting intermediate results
- [ ] Use references to avoid unnecessary clones
- [ ] Profile actual bottlenecks before optimizing
- [ ] Leverage compiler optimizations (release builds)
- [ ] Use unsafe sparingly, document invariants
- [ ] Monitor async runtime behavior
- [ ] Test on target architecture (different CPUs)
- [ ] Avoid unnecessary boxing and trait objects

---

**Last Updated**: 2025-11-22
**Related**: moai-lang-rust/SKILL.md, modules/advanced-patterns.md

