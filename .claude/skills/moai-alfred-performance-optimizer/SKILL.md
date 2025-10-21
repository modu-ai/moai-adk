---
name: moai-alfred-performance-optimizer
description: Performance analysis and optimization suggestions with profiling, bottleneck detection, and language-specific optimizations
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Performance Optimizer

## What it does

Performance analysis and optimization with profiling tools, bottleneck detection, and language-specific optimization techniques.

## When to use

- “Improve performance”, “Find slow parts”, “How to optimize?”
- “Profiling”, “Bottleneck”, “Memory leak”

## How it works

**Profiling Tools**:
- **Python**: cProfile, memory_profiler
- **TypeScript**: Chrome DevTools, clinic.js
- **Java**: JProfiler, VisualVM
- **Go**: pprof
- **Rust**: flamegraph, criterion

**Common Performance Issues**:
- **N+1 Query Problem**: Use eager loading/joins
- **Inefficient Loop**: O(n²) → O(n) with Set/Map
- **Memory Leak**: Remove event listeners, close connections

**Optimization Checklist**:
- [ ] Current performance benchmark
- [ ] Bottleneck identification
- [ ] Profiling data collected
- [ ] Algorithm complexity improved (O(n²) → O(n))
- [ ] Unnecessary operations removed
- [ ] Caching applied
- [ ] Async processing introduced
- [ ] Post-optimization benchmark
- [ ] Side effects checked

**Language-specific Optimizations**:
- **Python**: List comprehension, generators, @lru_cache
- **TypeScript**: Memoization, lazy loading, code splitting
- **Java**: Stream API, parallel processing
- **Go**: Goroutines, buffered channels
- **Rust**: Zero-cost abstractions, borrowing

**Performance Targets**:
- API response time: <200ms (P95)
- Page load time: <2s
- Memory usage: <512MB
- CPU usage: <70%

## Examples

User: "Please solve the N+1 query problem"
Claude: (identifies N+1 queries, suggests eager loading, provides optimized code)

## Works well with

- alfred-code-reviewer
- alfred-debugger-pro
