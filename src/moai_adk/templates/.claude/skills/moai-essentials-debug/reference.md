# moai-essentials-debug — Technical Reference

> **Version**: 2.1.0
> **Last Updated**: 2025-10-27

This document provides detailed technical specifications, debugger configurations for 23 languages, and container/distributed system debugging guides for the moai-essentials-debug Skill.

---

## 23 Language Debugger Matrix

### Systems Programming

#### C/C++
**Debuggers**: GDB 14.x, LLDB 17.x, AddressSanitizer

**CLI Commands**:
```bash
# GDB basic usage
gdb ./myapp
(gdb) break main
(gdb) run
(gdb) next
(gdb) print variable
(gdb) backtrace

# LLDB usage
lldb ./myapp
(lldb) b main
(lldb) run
(lldb) n
(lldb) p variable
(lldb) bt

# AddressSanitizer (memory error detection)
gcc -fsanitize=address -g myapp.c -o myapp
./myapp
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "C++ Debug",
      "type": "cppdbg",
      "request": "launch",
      "program": "${workspaceFolder}/build/myapp",
      "args": [],
      "stopAtEntry": false,
      "cwd": "${workspaceFolder}",
      "environment": [],
      "externalConsole": false,
      "MIMode": "gdb",
      "setupCommands": [
        {
          "description": "Enable pretty-printing",
          "text": "-enable-pretty-printing",
          "ignoreFailures": true
        }
      ]
    }
  ]
}
```

#### Rust
**Debuggers**: rust-lldb, rust-gdb, RUST_BACKTRACE

**CLI Commands**:
```bash
# Enable backtrace
RUST_BACKTRACE=1 cargo run
RUST_BACKTRACE=full cargo run  # Full backtrace

# rust-lldb usage
rust-lldb target/debug/myapp
(lldb) b main
(lldb) run

# rust-gdb usage
rust-gdb target/debug/myapp
(gdb) break main
(gdb) run
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Rust Debug",
      "type": "lldb",
      "request": "launch",
      "program": "${workspaceFolder}/target/debug/${workspaceFolderBasename}",
      "args": [],
      "cwd": "${workspaceFolder}",
      "env": {
        "RUST_BACKTRACE": "1"
      }
    }
  ]
}
```

#### Go
**Debugger**: Delve 1.22.x

**CLI Commands**:
```bash
# Delve debugging
dlv debug
(dlv) break main.main
(dlv) continue
(dlv) next
(dlv) print variable
(dlv) goroutines  # List goroutines
(dlv) goroutine 1  # Switch to specific goroutine

# Attach to running process
dlv attach <pid>

# Debug tests
dlv test -- -test.run TestName
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Go Debug",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "env": {},
      "args": []
    }
  ]
}
```

---

### JVM Ecosystem

#### Java
**Debuggers**: jdb, IntelliJ IDEA, Remote JDWP

**CLI Commands**:
```bash
# jdb usage
jdb -classpath . MyApp
> stop at MyClass:42
> run
> step
> print variable

# Enable remote debugging
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005 MyApp
```

**Remote Debugging Setup**:
```bash
# Connect IntelliJ/VSCode remote debugger
Host: localhost
Port: 5005
```

#### Kotlin
**Debugger**: IntelliJ Kotlin Debugger, Coroutines Debugger

**Coroutine Debugging**:
```kotlin
// Enable coroutine debug info
kotlinOptions {
    freeCompilerArgs += ["-Xdebug"]
}
```

**IntelliJ Coroutines Viewer**: View → Tool Windows → Kotlin Coroutines

#### Scala
**Debugger**: IntelliJ Scala Plugin, sbt debug mode

**sbt Debug Mode**:
```bash
# Run sbt debug mode
sbt -jvm-debug 5005 run
```

#### Clojure
**Debugger**: CIDER, Cursive, REPL-based debugging

**CIDER Debugging**:
```clojure
;; Set breakpoint
#break
(defn my-function [x]
  #break  ; Break here
  (+ x 1))
```

---

### Scripting Languages

#### Python
**Debuggers**: pdb, debugpy 1.8.0, pudb (TUI)

**CLI Commands**:
```bash
# pdb usage
python -m pdb script.py
(Pdb) break module.py:42
(Pdb) continue
(Pdb) next
(Pdb) print(variable)
(Pdb) where  # Stack trace

# pudb TUI debugger
python -m pudb script.py

# debugpy remote debugging
python -m debugpy --listen 5678 script.py
```

**In-code Breakpoints**:
```python
# Start pdb from code
import pdb; pdb.set_trace()

# Python 3.7+ breakpoint()
breakpoint()
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Python Debug",
      "type": "debugpy",
      "request": "launch",
      "program": "${file}",
      "console": "integratedTerminal",
      "justMyCode": false
    }
  ]
}
```

#### Ruby
**Debuggers**: debug gem (built-in), byebug, pry-byebug

**CLI Commands**:
```bash
# Ruby 3.1+ built-in debugger
ruby -r debug script.rb
debugger  # In code

# byebug usage
gem install byebug
# In code
require 'byebug'
byebug

# pry-byebug usage
gem install pry-byebug
# In code
require 'pry-byebug'
binding.pry
```

#### PHP
**Debuggers**: Xdebug 3.3.x, phpdbg

**Xdebug Setup** (php.ini):
```ini
[xdebug]
zend_extension=xdebug
xdebug.mode=debug
xdebug.start_with_request=yes
xdebug.client_port=9003
```

**phpdbg Usage**:
```bash
phpdbg -e script.php
phpdbg> break script.php:42
phpdbg> run
```

#### Lua
**Debuggers**: ZeroBrane Studio, MobDebug

**MobDebug Usage**:
```lua
require("mobdebug").start()
-- Breakpoint
require("mobdebug").pause()
```

#### Shell (Bash)
**Debug Mode**:
```bash
# Debug mode execution
bash -x script.sh

# Toggle within script
set -x  # Enable debug mode
# ... code ...
set +x  # Disable debug mode
```

---

### Web & Mobile

#### JavaScript
**Debuggers**: Chrome DevTools, node --inspect

**Node.js Debugging**:
```bash
# Debug mode execution
node --inspect script.js
# or break at first line
node --inspect-brk script.js

# Connect Chrome DevTools
chrome://inspect
```

**In-code Breakpoint**:
```javascript
debugger;  // Break here
```

#### TypeScript
**Debuggers**: Chrome DevTools + Source Maps, VS Code

**tsconfig.json Source Map Setup**:
```json
{
  "compilerOptions": {
    "sourceMap": true,
    "inlineSources": true
  }
}
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "TypeScript Debug",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/src/index.ts",
      "preLaunchTask": "tsc: build - tsconfig.json",
      "outFiles": ["${workspaceFolder}/dist/**/*.js"],
      "sourceMaps": true
    }
  ]
}
```

#### Dart/Flutter
**Debuggers**: Flutter DevTools, Hot Reload

**CLI Commands**:
```bash
# Flutter debug mode execution
flutter run --debug

# Open DevTools
flutter pub global activate devtools
flutter pub global run devtools
```

#### Swift
**Debuggers**: LLDB (Xcode), Instruments

**LLDB Usage**:
```bash
lldb MyApp.app
(lldb) breakpoint set --name viewDidLoad
(lldb) run
(lldb) po variable  # Print Object
```

---

### Functional & Concurrency

#### Haskell
**Debuggers**: GHCi debugger, Debug.Trace

**GHCi Debugging**:
```haskell
-- In GHCi
:break module.function
:trace expression
:step
:continue
```

**Debug.Trace Usage**:
```haskell
import Debug.Trace

myFunction x = trace ("x = " ++ show x) $ x + 1
```

#### Elixir
**Debuggers**: IEx debugger, :observer.start()

**IEx Debugging**:
```elixir
# In code
require IEx
IEx.pry()

# Run Observer
:observer.start()
```

#### Julia
**Debuggers**: Debugger.jl, Infiltrator.jl

**Debugger.jl Usage**:
```julia
using Debugger

@enter myfunction(args)
# or
@bp  # Breakpoint
```

#### R
**Debuggers**: browser(), debug(), RStudio Debugger

**CLI Commands**:
```r
# Debug function
debug(my_function)
my_function(args)

# In-code breakpoint
browser()

# Disable debug mode
undebug(my_function)
```

---

### Enterprise & Data

#### C#
**Debuggers**: Visual Studio Debugger, Rider, vsdbg

**vsdbg Usage** (Linux/macOS):
```bash
# Install vsdbg
curl -sSL https://aka.ms/getvsdbgsh | bash /dev/stdin -v latest -l ~/.vsdbg
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": ".NET Core Debug",
      "type": "coreclr",
      "request": "launch",
      "program": "${workspaceFolder}/bin/Debug/net8.0/MyApp.dll",
      "args": [],
      "cwd": "${workspaceFolder}",
      "stopAtEntry": false
    }
  ]
}
```

#### SQL
**Debuggers**: EXPLAIN ANALYZE, pg_stat_statements

**PostgreSQL Debugging**:
```sql
-- Analyze query plan
EXPLAIN ANALYZE SELECT * FROM users WHERE id = 1;

-- Track slow queries
CREATE EXTENSION pg_stat_statements;
SELECT * FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 10;
```

---

## Complete Container Debugging Guide

### Docker Debugging

#### Basic Debugging Patterns
```bash
# 1. Access running container
docker exec -it <container_name> /bin/sh

# 2. Check logs
docker logs <container_name>
docker logs -f <container_name>  # Real-time

# 3. Check container state
docker inspect <container_name>
docker stats <container_name>
```

#### Remote Debugging by Language

**Java (JDWP)**:
```dockerfile
# Dockerfile
ENV JAVA_TOOL_OPTIONS='-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005'
EXPOSE 5005
```

```bash
# Run container
docker run -p 5005:5005 -p 8080:8080 myapp
```

**Python (debugpy)**:
```dockerfile
# Dockerfile
RUN pip install debugpy
ENV DEBUGPY_ENABLE=true
EXPOSE 5678
```

```python
# app.py
import debugpy

if os.environ.get('DEBUGPY_ENABLE'):
    debugpy.listen(("0.0.0.0", 5678))
    print("Debugger listening on port 5678")
```

```bash
# Run container
docker run -p 5678:5678 -p 8000:8000 myapp
```

**Node.js (--inspect)**:
```dockerfile
# Dockerfile
EXPOSE 9229
CMD ["node", "--inspect=0.0.0.0:9229", "app.js"]
```

```bash
# Run container
docker run -p 9229:9229 -p 3000:3000 myapp
```

**Go (Delve)**:
```dockerfile
# Dockerfile
RUN go install github.com/go-delve/delve/cmd/dlv@latest
EXPOSE 2345
CMD ["dlv", "debug", "--headless", "--listen=:2345", "--api-version=2", "--accept-multiclient"]
```

```bash
# Run container
docker run -p 2345:2345 -p 8080:8080 myapp
```

#### Multi-stage Build Debugging
```dockerfile
# Debug stage
FROM golang:1.22 AS debug
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY . .
CMD ["dlv", "debug", "--headless", "--listen=:2345"]

# Production stage
FROM golang:1.22 AS production
COPY . .
RUN go build -o app
CMD ["./app"]
```

```bash
# Debug build
docker build --target debug -t myapp:debug .
docker run -p 2345:2345 myapp:debug
```

---

### Kubernetes Debugging

#### Basic Debugging Commands
```bash
# 1. Check pod logs
kubectl logs <pod-name>
kubectl logs -f <pod-name>  # Real-time
kubectl logs <pod-name> -c <container-name>  # Specific container
kubectl logs <pod-name> --previous  # Previous container logs

# 2. Access pod
kubectl exec -it <pod-name> -- /bin/sh
kubectl exec -it <pod-name> -c <container-name> -- /bin/bash

# 3. Check pod state
kubectl describe pod <pod-name>
kubectl get pod <pod-name> -o yaml
```

#### Port Forwarding (Debugger Connection)
```bash
# Java JDWP
kubectl port-forward pod/<pod-name> 5005:5005

# Python debugpy
kubectl port-forward pod/<pod-name> 5678:5678

# Node.js --inspect
kubectl port-forward pod/<pod-name> 9229:9229

# Go Delve
kubectl port-forward pod/<pod-name> 2345:2345
```

#### Ephemeral Containers (K8s 1.23+)
```bash
# Add ephemeral container with debug tools
kubectl debug -it <pod-name> --image=busybox --target=<container-name>

# or use debug tool image
kubectl debug -it <pod-name> --image=nicolaka/netshoot --target=<container-name>
```

#### Network Debugging
```bash
# Check network policies
kubectl get networkpolicies

# Check service endpoints
kubectl get endpoints <service-name>

# Check DNS (inside pod)
kubectl exec -it <pod-name> -- nslookup <service-name>
```

#### Resource Debugging
```bash
# Check resource usage
kubectl top pod <pod-name>
kubectl top node <node-name>

# Check events
kubectl get events --sort-by='.lastTimestamp'
kubectl get events --field-selector involvedObject.name=<pod-name>
```

---

## Distributed Tracing

### OpenTelemetry 1.24.0+ Setup

#### Python Setup
```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

# Setup Tracer Provider
trace.set_tracer_provider(TracerProvider())
tracer = trace.get_tracer(__name__)

# Setup OTLP Exporter
otlp_exporter = OTLPSpanExporter(
    endpoint="http://localhost:4317",
    insecure=True
)
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(otlp_exporter)
)

# Usage example
with tracer.start_as_current_span("my-operation"):
    # Perform operation
    pass
```

#### TypeScript/Node.js Setup
```typescript
import { NodeTracerProvider } from '@opentelemetry/sdk-trace-node';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';

// Setup Tracer Provider
const provider = new NodeTracerProvider();
const exporter = new OTLPTraceExporter({
  url: 'http://localhost:4317',
});

provider.addSpanProcessor(new BatchSpanProcessor(exporter));
provider.register();

// Usage example
import { trace } from '@opentelemetry/api';

const tracer = trace.getTracer('my-service');
const span = tracer.startSpan('my-operation');
// Perform operation
span.end();
```

#### Go Setup
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Setup Tracer Provider
exporter, _ := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint("localhost:4317"),
    otlptracegrpc.WithInsecure(),
)

tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),
)
otel.SetTracerProvider(tp)

// Usage example
tracer := otel.Tracer("my-service")
ctx, span := tracer.Start(ctx, "my-operation")
defer span.End()
```

---

### Prometheus 2.48.x Integration

#### Python (prometheus-client 0.19.0)
```python
from prometheus_client import Counter, Histogram, Gauge, start_http_server

# Define metrics
request_count = Counter('http_requests_total', 'Total HTTP requests')
request_duration = Histogram('http_request_duration_seconds', 'HTTP request duration')
active_connections = Gauge('active_connections', 'Active connections')

# Usage example
request_count.inc()
with request_duration.time():
    # Perform operation
    pass
active_connections.set(42)

# Start metrics server (port 8000)
start_http_server(8000)
```

#### Go (prometheus/client_golang)
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestCount = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests",
    })

    requestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request duration",
    })
)

func init() {
    prometheus.MustRegister(requestCount)
    prometheus.MustRegister(requestDuration)
}

// Metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

---

### Cloud Debugger Integration

#### AWS X-Ray
```python
from aws_xray_sdk.core import xray_recorder
from aws_xray_sdk.ext.flask.middleware import XRayMiddleware

app = Flask(__name__)
XRayMiddleware(app, xray_recorder)

# Custom subsegment
@xray_recorder.capture('my_function')
def my_function():
    # Perform operation
    pass
```

#### GCP Cloud Debugger
```python
try:
    import googleclouddebugger
    googleclouddebugger.enable()
except ImportError:
    pass
```

---

## Performance Profiling

### CPU Profiling

#### Python (cProfile, py-spy)
```bash
# cProfile
python -m cProfile -o output.prof script.py
python -m pstats output.prof

# py-spy (production-safe)
py-spy record -o profile.svg -- python script.py
py-spy top --pid <pid>
```

#### Go (pprof)
```go
import _ "net/http/pprof"

// Automatically adds pprof endpoints to HTTP server
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
# Collect CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze profile
go tool pprof -http=:8080 cpu.prof
```

#### Rust (flamegraph)
```toml
[dependencies]
pprof = { version = "0.13", features = ["flamegraph"] }
```

```bash
# Generate flamegraph
cargo flamegraph
```

#### Java (JFR)
```bash
# Enable JFR
java -XX:+FlightRecorder -XX:StartFlightRecording=duration=60s,filename=recording.jfr MyApp

# Analyze JFR file (JDK Mission Control)
jmc recording.jfr
```

---

### Memory Profiling

#### Python (memory_profiler, tracemalloc)
```python
from memory_profiler import profile

@profile
def my_function():
    # Perform operation
    pass
```

```bash
python -m memory_profiler script.py
```

**tracemalloc (built-in)**:
```python
import tracemalloc

tracemalloc.start()
# Perform operation
snapshot = tracemalloc.take_snapshot()
top_stats = snapshot.statistics('lineno')

for stat in top_stats[:10]:
    print(stat)
```

#### Go (pprof heap)
```bash
# Collect heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Track memory allocations
go tool pprof http://localhost:6060/debug/pprof/allocs
```

#### C/C++ (Valgrind massif)
```bash
# Run Massif
valgrind --tool=massif ./myapp

# Visualize Massif
ms_print massif.out.<pid>
```

#### Rust (heaptrack)
```bash
# Run heaptrack
heaptrack ./target/release/myapp

# Analyze results
heaptrack_gui heaptrack.myapp.<pid>.gz
```

---

## Advanced Debugging Techniques

### Conditional Breakpoints

**Python (pdb)**:
```python
# Conditional breakpoint
(Pdb) break script.py:42, x > 100
```

**GDB**:
```bash
(gdb) break main.c:42 if x > 100
```

**LLDB**:
```bash
(lldb) breakpoint set --file main.c --line 42 --condition 'x > 100'
```

---

### Time Travel Debugging

**GDB Reverse Execution**:
```bash
(gdb) target record-full
(gdb) continue
# After error occurs
(gdb) reverse-continue  # Reverse execution
(gdb) reverse-step
```

**WinDbg Time Travel Debugging** (Windows):
```bash
# Collect TTD trace
ttd.exe -out trace.run myapp.exe

# Debug trace
windbg -z trace.run
```

---

### Dynamic Instrumentation

**DTrace (macOS/FreeBSD)**:
```bash
# Trace function calls
sudo dtrace -n 'pid$target::my_function:entry { printf("Called with arg=%d", arg0); }' -p <pid>
```

**SystemTap (Linux)**:
```bash
# Trace function calls
stap -e 'probe process("/path/to/binary").function("my_function") { println("Called") }'
```

**BPF/eBPF (Linux)**:
```bash
# bpftrace usage
bpftrace -e 'uprobe:/path/to/binary:my_function { printf("Called\n"); }'
```

---

## Best Practices Summary

### 1. Debugger Selection
- Use language-appropriate debugger (Python → debugpy, Go → Delve)
- Use safe profilers in production (py-spy, async-profiler)

### 2. Logging Strategy
- Use structured logging (JSON format)
- Set log levels appropriately (DEBUG, INFO, WARNING, ERROR)
- Add Correlation ID in distributed systems

### 3. Performance Considerations
- Include debug symbols (-g flag)
- Generate source maps (TypeScript, JavaScript)
- Perform warmup before profiling

### 4. Security
- Protect production debug ports with firewall
- Never log sensitive information
- Control debug mode via environment variables

### 5. Automation
- Add debug builds to CI/CD pipeline
- Automatic stack trace collection (Sentry, Rollbar)
- Automatic metrics collection (Prometheus, Datadog)

---

**End of Reference** | moai-essentials-debug v2.1.0
