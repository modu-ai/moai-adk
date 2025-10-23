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

# MoAI Essentials Debug v2.0

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Version | 2.0.0 |
| Created | 2025-10-22 |
| Last Updated | 2025-10-22 |
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

## Debugger Matrix (23 Languages)

### Systems Programming

#### C
- **Primary**: `gdb` (GNU Debugger 14.x)
- **Alternative**: `lldb` (LLVM 17.x)
- **VSCode**: C/C++ Extension (ms-vscode.cpptools)
- **CLI**: `gdb ./program`, `break main`, `run`, `bt`, `print var`

#### C++
- **Primary**: `gdb` 14.x with C++23 support
- **Alternative**: `lldb` 17.x
- **GUI**: CLion Debugger, VS Code C++ extension
- **CLI**: `gdb --args ./app arg1 arg2`
- **Tools**: AddressSanitizer, Valgrind 3.22

#### Rust
- **Primary**: `rust-lldb` (shipped with rustup)
- **Alternative**: `rust-gdb`, CodeLLDB (VS Code)
- **CLI**: `rust-gdb target/debug/myapp`
- **Panic traces**: `RUST_BACKTRACE=1 cargo run`
- **Tools**: `rust-analyzer` debugger integration

#### Go
- **Primary**: Delve 1.22.x (`dlv`)
- **VSCode**: Go Extension (golang.go) with Delve
- **CLI**: `dlv debug`, `dlv test`, `dlv attach <pid>`
- **Remote**: `dlv debug --headless --listen=:2345`
- **Goroutine debugging**: `goroutines`, `goroutine <id>`

### JVM Ecosystem

#### Java
- **Primary**: `jdb` (built-in with JDK)
- **IDE**: IntelliJ IDEA Debugger, Eclipse Debugger
- **Remote**: `-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005`
- **CLI**: `jdb -attach 5005`
- **Tools**: VisualVM, JProfiler, Java Flight Recorder

#### Kotlin
- **Primary**: IntelliJ IDEA Kotlin Debugger
- **Coroutines**: Enable "Async stack traces" in IDE
- **CLI**: Same as Java (`jdb` with Kotlin bytecode)
- **Tools**: Kotlin Coroutines Debugger, Android Studio profiler

#### Scala
- **Primary**: IntelliJ IDEA Scala Plugin debugger
- **sbt**: `sbt -jvm-debug 5005`
- **Remote**: Same JDWP protocol as Java
- **Tools**: Metals language server debugging

#### Clojure
- **Primary**: CIDER debugger (Emacs), Cursive (IntelliJ)
- **REPL-based**: `(clojure.tools.trace/trace-ns 'my-ns)`
- **CLI**: Leiningen with `jdb` attachment
- **Tools**: `clojure.tools.logging`, `timbre` logging

### Scripting Languages

#### Python
- **Primary**: `pdb` (built-in), `debugpy` 1.8.0
- **Enhanced**: `pudb` 2024.1 (TUI), `ipdb` (IPython)
- **VSCode**: Python Extension with debugpy
- **CLI**: `python -m pdb script.py`
- **Async**: `await` stepping, asyncio task inspection
- **Remote**: `debugpy.listen(5678)`, `debugpy.wait_for_client()`

#### Ruby
- **Primary**: `debug` gem (Ruby 3.2+ built-in)
- **Alternative**: `byebug`, `pry-byebug`
- **VSCode**: Ruby LSP + debug gem
- **CLI**: `ruby -r debug script.rb`
- **Rails**: `binding.break` in views/controllers

#### PHP
- **Primary**: Xdebug 3.3.x
- **Alternative**: `phpdbg` (built-in CLI)
- **VSCode**: PHP Debug Extension (xdebug.php-debug)
- **CLI**: `php -dxdebug.mode=debug -dxdebug.start_with_request=yes script.php`
- **Laravel**: Laravel Debugbar, Telescope

#### Lua
- **Primary**: ZeroBrane Studio debugger
- **Alternative**: MobDebug, lua-debug (VS Code)
- **CLI**: `luadebug` module
- **Tools**: LuaRocks debug packages

#### Shell (Bash)
- **Primary**: `bash -x script.sh` (trace mode)
- **Interactive**: `set -x`, `set +x` toggle
- **VSCode**: Bash Debug Extension
- **Tools**: ShellCheck 0.9.x for static analysis
- **Trace**: `PS4='+(${BASH_SOURCE}:${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'`

### Web & Mobile

#### JavaScript
- **Primary**: Chrome DevTools, Node.js `inspect`
- **VSCode**: Built-in JavaScript Debugger
- **CLI**: `node --inspect script.js`, `node --inspect-brk script.js`
- **Browser**: `debugger;` statement, breakpoints in DevTools
- **Async**: Async stack traces in Chrome DevTools

#### TypeScript
- **Primary**: Chrome DevTools (via source maps), VS Code TS Debugger
- **CLI**: `node --inspect -r ts-node/register script.ts`
- **Source Maps**: Ensure `"sourceMap": true` in `tsconfig.json`
- **VSCode**: Auto-attaches to Node.js processes

#### Dart/Flutter
- **Primary**: Flutter DevTools, VS Code Dart extension
- **CLI**: `flutter run --observe`, `dart run --observe`
- **Hot reload**: `r` (reload), `R` (hot restart)
- **Observatory**: `http://localhost:port/` for VM profiling

#### Swift
- **Primary**: LLDB (Xcode integrated)
- **CLI**: `lldb ./MyApp`
- **Xcode**: Breakpoints, memory graph debugger
- **SwiftUI**: View hierarchy debugger
- **Tools**: Instruments for performance profiling

### Functional & Concurrency

#### Haskell
- **Primary**: GHCi debugger (`:break`, `:step`, `:trace`)
- **CLI**: `ghci -fbreak-on-exception script.hs`
- **Tools**: `Debug.Trace` module, `eventlog2html`
- **Profiling**: `ghc -prof -fprof-auto`, `hp2ps`

#### Elixir
- **Primary**: IEx debugger (`:debugger.start()`)
- **Observer**: `:observer.start()` for live system inspection
- **Tools**: `:sys.trace/2`, `:sys.get_status/1`
- **Phoenix**: Phoenix LiveDashboard for web apps

#### Julia
- **Primary**: Debugger.jl, Infiltrator.jl
- **CLI**: `using Debugger; @enter myfunction(args)`
- **VSCode**: Julia Extension with debug support
- **Profiling**: `@profile`, ProfileView.jl

#### R
- **Primary**: `browser()`, `debug()`, `debugonce()`
- **RStudio**: Built-in debugger with breakpoints
- **CLI**: `traceback()`, `recover()`
- **Tools**: `profvis` for profiling

### Enterprise & Data

#### C#
- **Primary**: Visual Studio Debugger, Rider
- **CLI**: `dotnet run` with VS Code C# extension
- **Remote**: `vsdbg` for Linux/macOS
- **Tools**: PerfView, dotTrace, WinDbg
- **Async**: Async call stack inspection

#### SQL
- **PostgreSQL**: `\set VERBOSITY verbose`, EXPLAIN ANALYZE
- **MySQL**: `SHOW WARNINGS`, slow query log
- **Tools**: pgAdmin debugger, MySQL Workbench
- **Profiling**: `pg_stat_statements`, Performance Schema

## Container & Kubernetes Debugging

### Docker Debugging
```bash
# Attach to running container
docker exec -it <container> /bin/sh

# Debug with debugger ports exposed
docker run -p 5005:5005 -e JAVA_TOOL_OPTIONS='-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005' myapp

# Python remote debugging
docker run -p 5678:5678 -e DEBUGPY_ENABLE=true myapp

# View logs with timestamps
docker logs --timestamps --follow <container>
```

### Kubernetes Debugging
```bash
# Port-forward debugger port
kubectl port-forward pod/myapp-pod 5005:5005

# Exec into pod
kubectl exec -it myapp-pod -- /bin/bash

# Debug with ephemeral container (K8s 1.23+)
kubectl debug -it myapp-pod --image=busybox --target=myapp

# Stream logs
kubectl logs -f deployment/myapp --all-containers=true
```

### Debug Container Images
- **Distroless debugging**: Use `gcr.io/distroless/base:debug` variants
- **Scratch debugging**: `kubectl debug` with busybox/alpine

## Distributed Tracing & Observability

### OpenTelemetry 1.24.0+
```python
# Python instrumentation
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

provider = TracerProvider()
processor = BatchSpanProcessor(OTLPSpanExporter(endpoint="http://localhost:4317"))
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

tracer = trace.get_tracer(__name__)
with tracer.start_as_current_span("operation"):
    # Your code here
    pass
```

```typescript
// TypeScript instrumentation
import { NodeTracerProvider } from '@opentelemetry/sdk-trace-node';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';

const provider = new NodeTracerProvider();
provider.addSpanProcessor(new BatchSpanProcessor(new OTLPTraceExporter()));
provider.register();
```

### Prometheus 2.48.x Integration
```yaml
# Scrape config
scrape_configs:
  - job_name: 'my-app'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

```python
# Python metrics (prometheus-client 0.19.0)
from prometheus_client import Counter, Histogram, start_http_server

request_count = Counter('app_requests_total', 'Total requests')
request_duration = Histogram('app_request_duration_seconds', 'Request duration')

@request_duration.time()
def handle_request():
    request_count.inc()
    # Handle request
```

### Cloud Debuggers

#### AWS X-Ray
```python
from aws_xray_sdk.core import xray_recorder
from aws_xray_sdk.core import patch_all

patch_all()  # Auto-instrument AWS SDK calls

@xray_recorder.capture('my_function')
def my_function():
    pass
```

#### GCP Cloud Debugger
```python
try:
    import googleclouddebugger
    googleclouddebugger.enable(
        module='myapp',
        version='v1.0'
    )
except ImportError:
    pass  # Not available in dev environment
```

## Stack Trace Analysis Patterns

### Python
```python
# Traceback example
Traceback (most recent call last):
  File "app.py", line 42, in process_data
    result = data['missing_key']
KeyError: 'missing_key'

# Analysis
📍 Location: app.py:42 in process_data()
🔍 Root Cause: Accessing dictionary key without validation
💡 Fix Suggestions:
   1. Use data.get('missing_key', default_value)
   2. Add key existence check: if 'missing_key' in data
   3. Use try-except for KeyError handling
```

### TypeScript/JavaScript
```typescript
// Error example
TypeError: Cannot read properties of undefined (reading 'name')
    at getUserName (user.service.ts:15:23)
    at processUser (user.controller.ts:42:10)

// Analysis
📍 Location: user.service.ts:15 in getUserName()
🔍 Root Cause: Object is undefined before property access
💡 Fix Suggestions:
   1. Add null check: user?.name
   2. Type guard: if (user && 'name' in user)
   3. Use optional chaining throughout call chain
```

### Java
```java
// Exception example
java.lang.NullPointerException: Cannot invoke "String.length()" because "text" is null
    at com.example.TextProcessor.process(TextProcessor.java:23)
    at com.example.Main.main(Main.java:15)

// Analysis
📍 Location: TextProcessor.java:23 in process()
🔍 Root Cause: Null reference passed to method
💡 Fix Suggestions:
   1. Use Optional<String> parameter type
   2. Add Objects.requireNonNull(text) guard
   3. Add @NonNull annotation with null-checking framework
```

### Go
```go
// Panic example
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x10a4b20]

goroutine 1 [running]:
main.processData(0x0)
    /app/main.go:42 +0x20
main.main()
    /app/main.go:15 +0x40

// Analysis
📍 Location: main.go:42 in processData()
🔍 Root Cause: Nil pointer dereference
💡 Fix Suggestions:
   1. Add nil check: if data != nil { ... }
   2. Initialize pointer before use
   3. Use defensive programming with early returns
```

### Rust
```rust
// Panic example
thread 'main' panicked at 'called `Option::unwrap()` on a `None` value', src/main.rs:42:23
stack backtrace:
   0: rust_begin_unwind
   1: core::panicking::panic_fmt
   2: myapp::process_data
             at ./src/main.rs:42

// Analysis
📍 Location: src/main.rs:42 in process_data()
🔍 Root Cause: Unwrapping None Option without checking
💡 Fix Suggestions:
   1. Use match expression: match opt { Some(v) => ..., None => ... }
   2. Use if let: if let Some(v) = opt { ... }
   3. Use .unwrap_or(default) or .unwrap_or_else(|| ...)
```

## Common Error Patterns by Language

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

## Debugging Workflow

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

## VSCode Launch Configuration Examples

### Python (debugpy)
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Python: Current File",
      "type": "debugpy",
      "request": "launch",
      "program": "${file}",
      "console": "integratedTerminal",
      "justMyCode": false
    },
    {
      "name": "Python: Remote Attach",
      "type": "debugpy",
      "request": "attach",
      "connect": {
        "host": "localhost",
        "port": 5678
      },
      "pathMappings": [
        {
          "localRoot": "${workspaceFolder}",
          "remoteRoot": "/app"
        }
      ]
    }
  ]
}
```

### Node.js/TypeScript
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "node",
      "request": "launch",
      "name": "TypeScript: Current File",
      "runtimeArgs": ["-r", "ts-node/register"],
      "args": ["${file}"],
      "cwd": "${workspaceFolder}",
      "protocol": "inspector"
    },
    {
      "type": "node",
      "request": "attach",
      "name": "Attach to Remote",
      "address": "localhost",
      "port": 9229,
      "localRoot": "${workspaceFolder}",
      "remoteRoot": "/app"
    }
  ]
}
```

### Go (Delve)
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Package",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}"
    },
    {
      "name": "Attach to Process",
      "type": "go",
      "request": "attach",
      "mode": "local",
      "processId": "${command:pickProcess}"
    }
  ]
}
```

### Rust (CodeLLDB)
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "lldb",
      "request": "launch",
      "name": "Debug executable",
      "cargo": {
        "args": ["build", "--bin=myapp", "--package=myapp"]
      },
      "args": [],
      "cwd": "${workspaceFolder}"
    }
  ]
}
```

### Java
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "java",
      "name": "Debug (Launch)",
      "request": "launch",
      "mainClass": "com.example.Main"
    },
    {
      "type": "java",
      "name": "Debug (Attach)",
      "request": "attach",
      "hostName": "localhost",
      "port": 5005
    }
  ]
}
```

## Performance Profiling Integration

### Python (cProfile + py-spy)
```bash
# CPU profiling
python -m cProfile -o output.prof script.py
python -m pstats output.prof

# Sampling profiler (production-safe)
py-spy top --pid <pid>
py-spy record -o profile.svg --pid <pid>
```

### Go (pprof)
```bash
# CPU profiling
go test -cpuprofile cpu.prof -bench .
go tool pprof cpu.prof

# Memory profiling
go test -memprofile mem.prof -bench .
go tool pprof mem.prof
```

### Rust (flamegraph)
```bash
cargo install flamegraph
cargo flamegraph --bin myapp
```

### Java (JFR)
```bash
# Start with JFR enabled
java -XX:StartFlightRecording=duration=60s,filename=recording.jfr -jar app.jar

# Analyze with JMC or jfr CLI
jfr print recording.jfr
```

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

## Failure Modes
- Insufficient stack trace or log information
- Unable to reproduce error locally
- Complex distributed system failures requiring multi-service tracing
- Race conditions or timing-dependent bugs

## Dependencies
- Works with: tdd-implementer, debug-helper, quality-gate
- Requires: Language-specific debugger tools installed
- Optional: OpenTelemetry, Prometheus, cloud debugger SDKs

## References
- OpenTelemetry Documentation. https://opentelemetry.io/docs/ (accessed 2025-10-22)
- Prometheus Documentation. https://prometheus.io/docs/ (accessed 2025-10-22)
- Microsoft. "Debugging in Visual Studio Code." https://code.visualstudio.com/docs/editor/debugging (accessed 2025-10-22)
- JetBrains. "Debugging Code." https://www.jetbrains.com/help/idea/debugging-code.html (accessed 2025-10-22)
- AWS. "AWS X-Ray Developer Guide." https://docs.aws.amazon.com/xray/ (accessed 2025-10-22)
- Google Cloud. "Cloud Debugger Documentation." https://cloud.google.com/debugger/docs (accessed 2025-10-22)
- Delve Debugger Documentation. https://github.com/go-delve/delve (accessed 2025-10-22)
- Rust Debugging Documentation. https://doc.rust-lang.org/book/ch09-00-error-handling.html (accessed 2025-10-22)

## Changelog
- 2025-10-22: v2.0.0 - Complete rewrite with 23-language coverage, container debugging, distributed tracing, cloud debuggers
- 2025-03-29: v1.0.0 - Initial version with Python/TypeScript/Java support

## Works well with
- moai-essentials-refactor (clean up code after debugging)
- moai-essentials-perf (performance bottleneck investigation)
- moai-alfred-debugger-pro (advanced debugging strategies)
- moai-foundation-trust (ensure debugging doesn't skip tests)

## Best Practices
- Always create regression test after fixing bug (TDD cycle)
- Log debugging insights in code comments with @TAG references
- Use language-appropriate debugger (don't force Python workflow on Go)
- Enable source maps for compiled/transpiled languages
- Set up distributed tracing early in microservices projects
- Use production-safe profilers (py-spy, async-profiler) in live systems
- Document reproduction steps in issue tracker or SPEC HISTORY
