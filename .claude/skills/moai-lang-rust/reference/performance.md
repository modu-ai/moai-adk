# Rust Performance Optimization

## Core Question

**What's the bottleneck, and is optimization worth it?**

Before optimizing:
- Have you measured? (Don't guess)
- What's the acceptable performance?
- Will optimization add complexity?

## Performance Decision → Implementation

| Goal | Design Choice | Implementation |
|------|---------------|----------------|
| Reduce allocations | Pre-allocate, reuse | `Vec::with_capacity`, object pools |
| Improve cache | Contiguous data | `Vec`, `SmallVec` |
| Parallelize | Data parallelism | `rayon`, threads |
| Avoid copies | Zero-copy | References, `Cow<T>` |
| Reduce indirection | Inline data | `smallvec`, arrays |

## Thinking Prompt

Before optimizing:

1. **Have you measured?**
   - Profile first → flamegraph, perf
   - Benchmark → criterion, cargo bench
   - Identify actual hotspots

2. **What's the priority?**
   - Algorithm (10x-1000x improvement)
   - Data structure (2x-10x)
   - Allocation (2x-5x)
   - Cache (1.5x-3x)

3. **What's the trade-off?**
   - Complexity vs speed
   - Memory vs CPU
   - Latency vs throughput

## Optimization Priority

```
1. Algorithm choice     (10x - 1000x)
2. Data structure       (2x - 10x)
3. Allocation reduction (2x - 5x)
4. Cache optimization   (1.5x - 3x)
5. SIMD/Parallelism     (2x - 8x)
```

## Common Techniques

| Technique | When | How |
|-----------|------|-----|
| Pre-allocation | Known size | `Vec::with_capacity(n)` |
| Avoid cloning | Hot paths | Use references or `Cow<T>` |
| Batch operations | Many small ops | Collect then process |
| SmallVec | Usually small | `smallvec::SmallVec<[T; N]>` |
| Inline buffers | Fixed-size data | Arrays over Vec |

## Quick Reference

| Tool | Purpose |
|------|---------|
| `cargo bench` | Micro-benchmarks |
| `criterion` | Statistical benchmarks |
| `perf` / `flamegraph` | CPU profiling |
| `heaptrack` | Allocation tracking |
| `valgrind` / `cachegrind` | Cache analysis |

## Common Mistakes

| Mistake | Why Wrong | Better |
|---------|-----------|--------|
| Optimize without profiling | Wrong target | Profile first |
| Benchmark in debug mode | Meaningless | Always `--release` |
| Use LinkedList | Cache unfriendly | `Vec` or `VecDeque` |
| Hidden `.clone()` | Unnecessary allocs | Use references |
| Premature optimization | Wasted effort | Make it work first |

## Anti-Patterns

| Anti-Pattern | Why Bad | Better |
|--------------|---------|--------|
| Clone to avoid lifetimes | Performance cost | Proper ownership |
| Box everything | Indirection cost | Stack when possible |
| HashMap for small sets | Overhead | Vec with linear search |
| String concat in loop | O(n^2) | `String::with_capacity` or `format!` |

## Release Build Optimization

In `Cargo.toml`:

```toml
[profile.release]
lto = true              # Link-time optimization
codegen-units = 1       # Better optimization
panic = "abort"         # Smaller binary
strip = true            # Remove symbols
```

## Zero-Cost Abstractions

Rust provides zero-cost abstractions at compile time:

### Iterators

```rust
// Zero-cost iterator
let sum: i32 = numbers.iter().map(|x| x * 2).sum();

// Compiler optimizes to simple loop
```

### Const Generics

```rust
struct Array<T, const N: usize> {
    data: [T; N],
}

// Size known at compile time
```

## Memory Layout

### Struct Field Ordering

Order fields to minimize padding:

```rust
// Bad: 7 bytes padding after 'a'
struct Bad {
    a: u8,    // 1 byte
    b: u64,   // 8 bytes
    c: u8,    // 1 byte
} // 24 bytes total

// Good: minimal padding
struct Good {
    b: u64,   // 8 bytes
    a: u8,    // 1 byte
    c: u8,    // 1 byte
} // 16 bytes total
```

### Enum Optimization

Rust optimizes enums with no data:

```rust
enum Option {
    Some,
    None,
}
// Uses single byte
```

## Allocation Reduction

### Reuse Allocations

```rust
// Reuse buffer
let mut buf = Vec::with_capacity(1024);
for item in items {
    buf.clear();
    // Fill buffer...
    process(&buf);
}
```

### String vs &str

Use `&str` for string slices when ownership isn't needed:

```rust
// Prefer
fn process(s: &str) { }

// Over
fn process(s: String) { }
```

### Cow for Conditional Cloning

```rust
use std::borrow::Cow;

fn maybe_uppercase(s: &str) -> Cow<str> {
    if s.chars().all(|c| c.is_lowercase()) {
        Cow::Borrowed(s)  // No allocation
    } else {
        Cow::Owned(s.to_uppercase())  // Allocate only when needed
    }
}
```

## Parallelism

### Rayon for Data Parallelism

```rust
use rayon::prelude::*;

let sum: i32 = numbers.par_iter()
    .map(|x| x * 2)
    .sum();
```

### Async for I/O Bound

Use tokio for I/O-bound operations:

```rust
async fn fetch_multiple(urls: Vec<&str>) -> Vec<String> {
    let futures: Vec<_> = urls.into_iter()
        .map(|url| fetch(url))
        .collect();

    futures::future::join_all(futures).await
}
```

## SIMD

Use portable SIMD for vector operations:

```rust
use std::simd::*;

fn add_simd(a: [f32; 4], b: [f32; 4]) -> [f32; 4] {
    let a = f32x4::from_array(a);
    let b = f32x4::from_array(b);
    (a + b).to_array()
}
```

## Cache-Friendly Data Structures

### Contiguous Memory

Prefer `Vec` over linked lists for cache efficiency:

```rust
// Good: contiguous memory
let mut data = Vec::new();

// Bad: scattered allocations
let mut data = LinkedList::new();
```

### SoA (Struct of Arrays)

For better cache utilization:

```rust
// AoS (Array of Structures)
struct ParticleAoS {
    position: [f32; 3],
    velocity: [f32; 3],
    mass: f32,
}

// SoA (Struct of Arrays) - better cache usage
struct ParticleSoA {
    positions: Vec<[f32; 3]>,
    velocities: Vec<[f32; 3]>,
    masses: Vec<f32>,
}
```

## Measurement

### Cargo Bench

```bash
# Run benchmarks
cargo bench

# With specific filter
cargo bench --bench my_benchmark

# Save benchmark data
cargo bench -- --save-baseline main
```

### Criterion

Statistical benchmarking with better analysis:

```rust
use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn bench_function(c: &mut Criterion) {
    c.bench_function("my_function", |b| {
        b.iter(|| {
            my_function(black_box(input_data))
        })
    });
}

criterion_group!(benches, bench_function);
criterion_main!(benches);
```

### Flamegraph

```bash
# Install flamegraph
cargo install flamegraph

# Generate flamegraph
cargo flamegraph --bin my_binary

# View flamegraph
firefox flamegraph.svg
```

## Profiling Tools

| Tool | Use Case |
|------|----------|
| `valgrind` | Memory profiling, cache analysis |
| `perf` | CPU profiling on Linux |
| `flamegraph` | Visualize call stacks |
| `heaptrack` | Memory allocation tracking |
| `dhat` | Heap profiling |

Remember: Always measure before optimizing. Profile to find actual bottlenecks, not assumed ones.
