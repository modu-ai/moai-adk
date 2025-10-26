---
name: moai-essentials-debug
description: Advanced debugging with stack trace analysis, error pattern detection, and fix suggestions. Use when delivering quick diagnostic support for everyday issues.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# MoAI Essentials Debug v2.1

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Version | 2.1.0 |
| Created | 2025-10-22 |
| Last Updated | 2025-10-27 |
| Language Coverage | 23 languages + containers + distributed systems |
| Allowed tools | Read, Write, Edit, Bash, TodoWrite |
| Auto-load | On demand during Run stage (debug-helper) |
| Trigger cues | Runtime error triage, stack trace analysis, root cause investigation requests |

## What it does

Comprehensive debugging support across all 23 MoAI-ADK languages with:
- Language-specific debugger integration
- Stack trace analysis and error pattern detection
- Container and Kubernetes debugging
- Distributed tracing with OpenTelemetry
- Cloud debugger integration (AWS X-Ray, GCP Cloud Debugger)
- Performance profiling with Prometheus

## When to use

- Runtime errors, exceptions, crashes
- Stack trace analysis requests
- "Why is this failing?", "Debug this error"
- Container/K8s debugging scenarios
- Distributed system tracing
- Performance bottleneck investigation
- Automatically invoked via debug-helper sub-agent

---

## Quick Reference: Debugger by Language

### Systems Programming
- **C/C++**: gdb 14.x, lldb 17.x, AddressSanitizer
- **Rust**: rust-lldb, rust-gdb, RUST_BACKTRACE=1
- **Go**: Delve 1.22.x, goroutine debugging

### JVM Ecosystem
- **Java**: jdb, IntelliJ IDEA, Remote JDWP
- **Kotlin**: IntelliJ Kotlin Debugger, Coroutines debugger
- **Scala**: IntelliJ Scala Plugin, sbt debug mode
- **Clojure**: CIDER, Cursive, REPL-based debugging

### Scripting Languages
- **Python**: pdb, debugpy 1.8.0, pudb (TUI)
- **Ruby**: debug gem (built-in), byebug, pry-byebug
- **PHP**: Xdebug 3.3.x, phpdbg
- **Lua**: ZeroBrane Studio, MobDebug
- **Shell**: bash -x, set -x toggle

### Web & Mobile
- **JavaScript**: Chrome DevTools, node --inspect
- **TypeScript**: Chrome DevTools with source maps, VS Code debugger
- **Dart/Flutter**: Flutter DevTools, hot reload
- **Swift**: LLDB (Xcode), Instruments profiling

### Functional & Concurrency
- **Haskell**: GHCi debugger, Debug.Trace
- **Elixir**: IEx debugger, :observer.start()
- **Julia**: Debugger.jl, Infiltrator.jl
- **R**: browser(), debug(), RStudio debugger

### Enterprise & Data
- **C#**: Visual Studio Debugger, Rider, vsdbg
- **SQL**: EXPLAIN ANALYZE, pg_stat_statements

> **Complete debugger matrix with CLI commands**: See [reference.md](reference.md)

---

## Debugging Workflow (6-Step Process)

### 1. Reproduce
- [ ] Minimal reproducible example (MRE)
- [ ] Consistent reproduction steps
- [ ] Document environment (OS, language version, dependencies)

### 2. Isolate
- [ ] Binary search the code (comment out sections)
- [ ] Check recent changes (git diff, git log)
- [ ] Verify input data and edge cases

### 3. Investigate
- [ ] Read stack trace from bottom (entry point) to top (error site)
- [ ] Add logging at key decision points
- [ ] Use debugger breakpoints before error location
- [ ] Check variable state in debugger

### 4. Hypothesize
- [ ] Form theory about root cause
- [ ] Identify 2-3 most likely culprits
- [ ] Design experiment to test hypothesis

### 5. Fix
- [ ] Implement minimal fix first
- [ ] Add regression test (RED → GREEN)
- [ ] Refactor if needed (REFACTOR stage)
- [ ] Update documentation

### 6. Verify
- [ ] Run full test suite
- [ ] Test edge cases explicitly
- [ ] Verify fix in production-like environment
- [ ] Monitor for recurrence

---

## Common Error Patterns by Language Category

### Memory Safety
- **C/C++**: Buffer overflow, use-after-free, memory leaks
  - Tools: Valgrind, AddressSanitizer (`-fsanitize=address`)
- **Rust**: Ownership violations (prevented at compile time)
- **Go**: Goroutine leaks, improper channel usage

### Null/Nil Handling
- **Java**: NullPointerException → Use Optional<T>
- **Kotlin**: NullPointerException → Leverage null safety (?.)
- **TypeScript**: undefined access → Optional chaining (?.)
- **Go**: Nil pointer → Early nil checks
- **Rust**: Option<T> unwrap → Pattern matching

### Type Errors
- **Python**: TypeError, AttributeError → Type hints + mypy
- **JavaScript**: Type coercion bugs → Use TypeScript
- **Ruby**: NoMethodError → Duck typing checks

### Concurrency Issues
- **Go**: Data races → `go build -race`, proper channel usage
- **Java**: ConcurrentModificationException → Use concurrent collections
- **Rust**: Data races (prevented by borrow checker)
- **Python**: GIL limitations → Use multiprocessing for CPU-bound tasks

### Async/Await Pitfalls
- **Python**: `RuntimeError: Event loop is closed` → Proper asyncio usage
- **JavaScript**: Unhandled promise rejections → Always catch async errors
- **Rust**: Send/Sync trait violations → Understand thread safety

> **Detailed error pattern analysis**: See [examples.md](examples.md)

---

## Container & Kubernetes Debugging

### Docker Debugging
```bash
# Attach to running container
docker exec -it <container> /bin/sh

# Debug with debugger ports exposed
docker run -p 5005:5005 -e JAVA_TOOL_OPTIONS='-agentlib:jdwp=...' myapp

# Python remote debugging
docker run -p 5678:5678 -e DEBUGPY_ENABLE=true myapp
```

### Kubernetes Debugging
```bash
# Port-forward debugger port
kubectl port-forward pod/myapp-pod 5005:5005

# Exec into pod
kubectl exec -it myapp-pod -- /bin/bash

# Debug with ephemeral container (K8s 1.23+)
kubectl debug -it myapp-pod --image=busybox --target=myapp
```

> **Complete container debugging guide**: See [reference.md](reference.md)

---

## Distributed Tracing

### OpenTelemetry 1.24.0+
- Python: opentelemetry-api, opentelemetry-sdk
- TypeScript: @opentelemetry/sdk-trace-node
- Exporters: Jaeger, Zipkin, OTLP

### Prometheus 2.48.x Integration
- Metrics: Counter, Gauge, Histogram, Summary
- Python: prometheus-client 0.19.0
- Go: prometheus/client_golang

### Cloud Debuggers
- **AWS X-Ray**: aws-xray-sdk, auto-instrumentation
- **GCP Cloud Debugger**: googleclouddebugger Python package

> **Complete distributed tracing setup**: See [reference.md](reference.md)

---

## Stack Trace Analysis Quick Guide

### Reading Stack Traces

**Direction**: Bottom (entry point) → Top (error site)

**Key information**:
- 📍 **Location**: File and line number
- 🔍 **Context**: Function name and arguments
- 💡 **Error type**: Exception class or error message

### Analysis Pattern

1. **Identify error type**: What exception/error?
2. **Locate error site**: Top of stack trace
3. **Trace execution path**: Follow stack from bottom
4. **Identify root cause**: Where did bad data/state originate?
5. **Suggest fix**: 1-3 actionable options

> **Language-specific stack trace examples**: See [examples.md](examples.md)

---

## Performance Profiling Integration

### CPU Profiling
- **Python**: cProfile, py-spy (production-safe)
- **Go**: pprof, go test -cpuprofile
- **Rust**: flamegraph crate
- **Java**: Java Flight Recorder (JFR)

### Memory Profiling
- **Python**: memory_profiler, tracemalloc
- **Go**: pprof memory profiling
- **C/C++**: Valgrind massif
- **Rust**: heaptrack

> **Complete profiling guide**: See [reference.md](reference.md)

---

## Inputs
- Stack traces, error messages, logs
- Code context (relevant files)
- Environment information (versions, config)
- Reproduction steps

## Outputs
- Root cause analysis with evidence
- Actionable fix suggestions (1-3 options)
- Debugging checklist tailored to error type
- Code snippets demonstrating fix

---

## Failure Modes
- Insufficient stack trace or log information
- Unable to reproduce error locally
- Complex distributed system failures requiring multi-service tracing
- Race conditions or timing-dependent bugs

---

## Dependencies
- Works with: tdd-implementer, debug-helper, quality-gate
- Requires: Language-specific debugger tools installed
- Optional: OpenTelemetry, Prometheus, cloud debugger SDKs

---

## Related Skills
- moai-essentials-refactor (clean up code after debugging)
- moai-essentials-perf (performance bottleneck investigation)
- moai-alfred-debugger-pro (advanced debugging strategies)
- moai-foundation-trust (ensure debugging doesn't skip tests)

---

## Best Practices
- Always create regression test after fixing bug (TDD cycle)
- Log debugging insights in code comments with @TAG references
- Use language-appropriate debugger (don't force Python workflow on Go)
- Enable source maps for compiled/transpiled languages
- Set up distributed tracing early in microservices projects
- Use production-safe profilers (py-spy, async-profiler) in live systems
- Document reproduction steps in issue tracker or SPEC HISTORY

---

**For complete 23-language debugger matrix**: [reference.md](reference.md)  
**For language-specific debugging examples**: [examples.md](examples.md)

---

**End of Skill** | Refactored 2025-10-27
