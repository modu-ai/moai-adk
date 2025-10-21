---

name: moai-alfred-performance-optimizer
description: Performance analysis and optimization suggestions with profiling, bottleneck detection, and language-specific optimizations. Use when planning performance improvements or regressions checks.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Performance Optimizer

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:2-run performance lane |
| Trigger cues | Performance regressions surfaced in Alfred workflows, profiling or tuning requests. |

## What it does

Performance analysis and optimization with profiling tools, bottleneck detection, and language-specific optimization techniques.

## When to use

- Activates when Alfred must diagnose or remediate performance bottlenecks.
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

## Best Practices
- The text shown to the user is written using TUI/report expressions.
- When running the tool, a summary of commands and results are recorded.

## Examples
```markdown
- Call this skill inside the /alfred command to generate a report.
- Add summary to Completion Report.
```

## Inputs
- MoAI-ADK project context (`.moai/project/`, `.claude/` templates, etc.).
- Parameters passed from user commands or higher commands.

## Outputs
- Reports, checklists or recommendations for your Alfred workflow.
- Structured data for subsequent subagent calls.

## Failure Modes
- When required input documents are missing or permissions are limited.
- When disruptive changes are required without user approval.

## Dependencies
- Cooperation with higher-level agents such as cc-manager and project-manager is required.

## References
- Google SRE. "The Four Golden Signals." https://sre.google/sre-book/monitoring-distributed-systems/ (accessed 2025-03-29).
- Dynatrace. "Application Performance Monitoring Best Practices." https://www.dynatrace.com/resources/ebooks/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-code-reviewer
- alfred-debugger-pro
