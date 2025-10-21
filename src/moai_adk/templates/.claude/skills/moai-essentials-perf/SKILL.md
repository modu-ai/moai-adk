---

name: moai-essentials-perf
description: Performance optimization with profiling, bottleneck detection, and tuning strategies. Use when performing baseline performance reviews.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Performance Optimizer

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | On demand during Run stage (performance triage) |
| Trigger cues | Latency complaints, profiling requests, throughput bottleneck analysis. |

## What it does

Performance analysis and optimization with profiling tools, bottleneck detection, and language-specific optimization techniques.

## When to use

- Loads when diagnosing performance regressions or planning optimization work.
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
```markdown
- Checks the current diff and lists items that can be modified immediately.
- Schedule follow-up tasks with TodoWrite.
```

## Inputs
- A snapshot of the code/tests/documentation you are currently working on.
- Ongoing agent status information.

## Outputs
- Immediately actionable checklists or improvement suggestions.
- Recommendations on whether to take next steps or not.

## Failure Modes
- If you cannot find the required files or test results.
- When the scope of work is excessively large and cannot be resolved with simple support.

## Dependencies
- Mainly used in conjunction with `tdd-implementer`, `quality-gate`, etc.

## References
- Google SRE. "The Four Golden Signals." https://sre.google/sre-book/monitoring-distributed-systems/ (accessed 2025-03-29).
- Dynatrace. "Application Performance Monitoring Best Practices." https://www.dynatrace.com/resources/ebooks/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Overhauled input/output definitions for Essentials skills.

## Works well with

- moai-essentials-refactor

## Best Practices
- Record results, even for simple improvements, to increase traceability.
- Clearly mark items that require human review to distinguish them from automation.
