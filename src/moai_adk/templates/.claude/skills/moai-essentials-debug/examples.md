# MoAI Essentials Debug - Real-World Examples

## Table of Contents
1. [Async/Await Debugging (Python)](#asyncawait-debugging-python)
2. [Goroutine Leak Investigation (Go)](#goroutine-leak-investigation-go)
3. [Distributed Tracing with OpenTelemetry (Multi-service)](#distributed-tracing-with-opentelemetry-multi-service)
4. [Kubernetes Pod Crash Debugging](#kubernetes-pod-crash-debugging)
5. [Memory Leak Diagnosis (Rust)](#memory-leak-diagnosis-rust)
6. [Race Condition Detection (Go)](#race-condition-detection-go)
7. [Null Pointer Debugging (TypeScript)](#null-pointer-debugging-typescript)
8. [Database Query Performance (SQL)](#database-query-performance-sql)

---

## Async/Await Debugging (Python)

### Problem
FastAPI application experiencing intermittent timeouts and "Event loop is closed" errors.

### Stack Trace
```python
Traceback (most recent call last):
  File "app.py", line 87, in process_request
    result = await fetch_data()
  File "app.py", line 42, in fetch_data
    async with session.get(url) as response:
RuntimeError: Event loop is closed
```

### Investigation Steps

#### 1. Reproduce with Debugger
```python
# app.py
import asyncio
import debugpy

# Enable remote debugging
debugpy.listen(5678)
print("Waiting for debugger...")
debugpy.wait_for_client()

async def fetch_data():
    # Set breakpoint here
    breakpoint()
    async with session.get(url) as response:
        return await response.json()
```

#### 2. Run with Async Debugging
```bash
# Terminal 1: Start app
python app.py

# Terminal 2: Connect VSCode debugger (Remote Attach config)
# Step through async code
```

#### 3. VSCode launch.json
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Python: Async Debug",
      "type": "debugpy",
      "request": "attach",
      "connect": {
        "host": "localhost",
        "port": 5678
      },
      "justMyCode": false,
      "pathMappings": [
        {
          "localRoot": "${workspaceFolder}",
          "remoteRoot": "."
        }
      ]
    }
  ]
}
```

#### 4. Inspect Async Context
```python
# In debugger console
import asyncio
loop = asyncio.get_event_loop()
print(f"Loop running: {loop.is_running()}")
print(f"Loop closed: {loop.is_closed()}")

# Check pending tasks
tasks = asyncio.all_tasks(loop)
print(f"Pending tasks: {len(tasks)}")
for task in tasks:
    print(task)
```

### Root Cause
Multiple event loops created; old loop closed while tasks still pending.

### Fix
```python
# Before (problematic)
def main():
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    loop.run_until_complete(fetch_data())
    loop.close()  # ❌ Closes loop prematurely

# After (correct)
async def main():
    await fetch_data()

if __name__ == "__main__":
    asyncio.run(main())  # ✅ Proper lifecycle management
```

### Test Case
```python
# tests/test_fetch_data.py
import pytest
import asyncio

@pytest.mark.asyncio
async def test_fetch_data_no_event_loop_error():
    """Ensure fetch_data doesn't close event loop prematurely."""
    result = await fetch_data()
    assert result is not None

    # Verify loop still running
    loop = asyncio.get_event_loop()
    assert not loop.is_closed()
```

---

## Goroutine Leak Investigation (Go)

### Problem
Go service memory usage growing indefinitely over time.

### Stack Trace
```
goroutine 1543 [chan receive]:
main.processMessages()
    /app/worker.go:42 +0x120
created by main.startWorker
    /app/worker.go:25 +0x80

goroutine 1544 [chan receive]:
main.processMessages()
    /app/worker.go:42 +0x120
created by main.startWorker
    /app/worker.go:25 +0x80
... (1542 more goroutines)
```

### Investigation Steps

#### 1. Enable Goroutine Profiling
```go
// main.go
import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    // Enable pprof
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // ... rest of application
}
```

#### 2. Capture Goroutine Dump
```bash
# Get goroutine count
curl http://localhost:6060/debug/pprof/goroutine?debug=1 > goroutines.txt

# Analyze goroutines
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) top
(pprof) list processMessages
```

#### 3. Debug with Delve
```bash
# Attach to running process
dlv attach $(pgrep myapp)

(dlv) goroutines
Goroutine 1 - User: /app/main.go:15 main.main (0x10a4b20)
Goroutine 2 - User: /app/worker.go:42 main.processMessages (0x10a5c30)
... (1542 more)

(dlv) goroutine 2 bt
0  0x000000000043e3e5 in runtime.gopark
   at /usr/local/go/src/runtime/proc.go:363
1  0x000000000040b5b6 in runtime.chanrecv
   at /usr/local/go/src/runtime/chan.go:583
2  0x00000000010a5c30 in main.processMessages
   at /app/worker.go:42
```

#### 4. VSCode launch.json
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug with Goroutine Inspection",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "showLog": true,
      "logOutput": "debugger",
      "dlvToolPath": "/usr/local/bin/dlv"
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

### Root Cause
Worker goroutines waiting on channel that never gets closed when context cancelled.

```go
// Before (leaking)
func startWorker(ctx context.Context) {
    msgChan := make(chan Message)
    go func() {
        for msg := range msgChan {  // ❌ Never exits if channel not closed
            processMessage(msg)
        }
    }()
    // ... no channel close on ctx.Done()
}

// After (fixed)
func startWorker(ctx context.Context) {
    msgChan := make(chan Message)
    go func() {
        defer close(msgChan)
        for {
            select {
            case msg := <-msgChan:
                processMessage(msg)
            case <-ctx.Done():
                return  // ✅ Proper cleanup
            }
        }
    }()
}
```

### Test Case
```go
// worker_test.go
func TestNoGoroutineLeak(t *testing.T) {
    initialCount := runtime.NumGoroutine()

    ctx, cancel := context.WithCancel(context.Background())
    startWorker(ctx)

    time.Sleep(100 * time.Millisecond)
    cancel()
    time.Sleep(100 * time.Millisecond)

    finalCount := runtime.NumGoroutine()
    assert.Equal(t, initialCount, finalCount, "Goroutine leak detected")
}
```

---

## Distributed Tracing with OpenTelemetry (Multi-service)

### Problem
Request slow across microservices; need to identify bottleneck.

### Architecture
```
User → API Gateway → Auth Service → User Service → Database
                  → Product Service → Cache
```

### Investigation Steps

#### 1. Instrument Services with OpenTelemetry

**API Gateway (Python/FastAPI)**
```python
# gateway/main.py
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

# Setup tracing
provider = TracerProvider()
processor = BatchSpanProcessor(OTLPSpanExporter(endpoint="http://jaeger:4317"))
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

app = FastAPI()
FastAPIInstrumentor.instrument_app(app)

@app.get("/users/{user_id}")
async def get_user(user_id: int):
    tracer = trace.get_tracer(__name__)

    with tracer.start_as_current_span("gateway.get_user") as span:
        span.set_attribute("user.id", user_id)

        # Call auth service
        with tracer.start_as_current_span("gateway.verify_auth"):
            auth_result = await verify_auth(user_id)

        # Call user service
        with tracer.start_as_current_span("gateway.fetch_user_data"):
            user_data = await fetch_user_data(user_id)

        return user_data
```

**User Service (Go)**
```go
// user-service/main.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/trace"
)

func initTracing() {
    exporter, _ := otlptracegrpc.New(
        context.Background(),
        otlptracegrpc.WithEndpoint("jaeger:4317"),
        otlptracegrpc.WithInsecure(),
    )

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
    )
    otel.SetTracerProvider(tp)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tracer := otel.Tracer("user-service")

    ctx, span := tracer.Start(ctx, "user-service.get_user")
    defer span.End()

    userID := chi.URLParam(r, "userID")
    span.SetAttributes(attribute.String("user.id", userID))

    // Database query with tracing
    ctx, dbSpan := tracer.Start(ctx, "user-service.db_query")
    user, err := db.GetUser(ctx, userID)
    dbSpan.End()

    if err != nil {
        span.RecordError(err)
        http.Error(w, "User not found", 404)
        return
    }

    json.NewEncoder(w).Render(user)
}
```

#### 2. Deploy Jaeger for Trace Collection
```yaml
# docker-compose.yml
version: '3.8'
services:
  jaeger:
    image: jaegertracing/all-in-one:1.50
    ports:
      - "16686:16686"  # Jaeger UI
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
    environment:
      - COLLECTOR_OTLP_ENABLED=true
```

#### 3. Generate Load and Capture Traces
```bash
# Start services
docker-compose up -d

# Generate test traffic
for i in {1..100}; do
  curl http://localhost:8000/users/123
  sleep 0.1
done

# Open Jaeger UI
open http://localhost:16686
```

#### 4. Analyze Traces in Jaeger UI
1. Navigate to Jaeger UI
2. Select service: `api-gateway`
3. Find slow traces (> 500ms)
4. Click trace to see waterfall view
5. Identify bottleneck span

**Example Trace Timeline**
```
Total: 1,245ms
├─ gateway.get_user: 1,245ms
   ├─ gateway.verify_auth: 45ms
   │  └─ auth-service.verify_token: 42ms
   │     └─ redis.get: 15ms
   ├─ gateway.fetch_user_data: 1,180ms  ← BOTTLENECK
      └─ user-service.get_user: 1,175ms
         └─ user-service.db_query: 1,165ms  ← ROOT CAUSE
            └─ postgres.query: 1,160ms
```

### Root Cause
Database query missing index on `users.id`.

### Fix
```sql
-- Add missing index
CREATE INDEX idx_users_id ON users(id);

-- Verify with EXPLAIN
EXPLAIN ANALYZE SELECT * FROM users WHERE id = 123;
-- Before: Seq Scan on users (cost=0.00..1245.00 rows=1 width=100) (actual time=1160.234..1160.235 rows=1 loops=1)
-- After:  Index Scan using idx_users_id on users (cost=0.42..8.44 rows=1 width=100) (actual time=0.023..0.024 rows=1 loops=1)
```

### Test Case
```python
# tests/test_performance.py
import pytest
from opentelemetry import trace

@pytest.mark.asyncio
async def test_get_user_performance():
    """Ensure /users/{id} responds within 100ms."""
    tracer = trace.get_tracer(__name__)

    with tracer.start_as_current_span("test.get_user_performance") as span:
        start = time.time()
        response = await client.get("/users/123")
        duration = time.time() - start

        span.set_attribute("response.duration_ms", duration * 1000)

        assert response.status_code == 200
        assert duration < 0.1, f"Request took {duration}s, expected < 0.1s"
```

---

## Kubernetes Pod Crash Debugging

### Problem
Pod enters `CrashLoopBackOff` state with OOMKilled status.

### Symptoms
```bash
kubectl get pods
NAME                    READY   STATUS             RESTARTS   AGE
myapp-7d4f8c9b5-xyz     0/1     CrashLoopBackOff   5          10m

kubectl describe pod myapp-7d4f8c9b5-xyz
...
Last State:     Terminated
  Reason:       OOMKilled
  Exit Code:    137
```

### Investigation Steps

#### 1. Check Logs (Current and Previous)
```bash
# Current logs
kubectl logs myapp-7d4f8c9b5-xyz

# Logs before crash
kubectl logs myapp-7d4f8c9b5-xyz --previous

# Output:
[INFO] Application starting...
[INFO] Loading dataset (1M records)...
[WARNING] Memory usage: 450MB
[WARNING] Memory usage: 750MB
[ERROR] Out of memory
```

#### 2. Inspect Resource Limits
```bash
kubectl get pod myapp-7d4f8c9b5-xyz -o yaml | grep -A 10 resources
  resources:
    limits:
      memory: 512Mi  # ❌ Too low
    requests:
      memory: 256Mi
```

#### 3. Debug with Ephemeral Container
```bash
# Add debug container to running pod
kubectl debug -it myapp-7d4f8c9b5-xyz --image=ubuntu --target=myapp

# In debug container
apt-get update && apt-get install -y htop
htop

# Check memory usage patterns
cat /proc/meminfo
free -h
```

#### 4. Profile Memory Usage
```python
# Add memory profiling to application
from memory_profiler import profile

@profile
def load_dataset():
    # This function is consuming too much memory
    data = [process_record(r) for r in fetch_records()]  # ❌ Loads all into memory
    return data
```

```bash
# Run with memory profiler
python -m memory_profiler app.py

# Output:
Line #    Mem usage    Increment   Line Contents
================================================
     3    125.2 MiB    0.0 MiB     @profile
     4                             def load_dataset():
     5    875.4 MiB  750.2 MiB         data = [process_record(r) for r in fetch_records()]
     6    875.4 MiB    0.0 MiB         return data
```

### Root Cause
Loading entire dataset into memory at once; exceeds pod memory limit.

### Fix
```python
# Before (memory-intensive)
def load_dataset():
    data = [process_record(r) for r in fetch_records()]  # ❌ 750MB
    return data

# After (streaming)
def load_dataset():
    for record in fetch_records():  # ✅ Process one at a time
        yield process_record(record)
```

**Update Kubernetes Deployment**
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: myapp
        resources:
          limits:
            memory: 1Gi  # ✅ Increased limit
          requests:
            memory: 512Mi
```

### Test Case
```python
# tests/test_memory.py
import tracemalloc

def test_load_dataset_memory_efficient():
    """Ensure load_dataset uses < 100MB memory."""
    tracemalloc.start()

    list(load_dataset())  # Consume generator

    current, peak = tracemalloc.get_traced_memory()
    tracemalloc.stop()

    assert peak / 1024 / 1024 < 100, f"Peak memory {peak/1024/1024}MB exceeds 100MB"
```

---

## Memory Leak Diagnosis (Rust)

### Problem
Long-running Rust service memory usage increasing over time.

### Stack Trace
```rust
thread 'main' panicked at 'allocation error: Cannot allocate memory', src/main.rs:142:23
```

### Investigation Steps

#### 1. Enable Memory Profiling with Valgrind
```bash
# Build with debug symbols
cargo build --release

# Run with Valgrind
valgrind --leak-check=full --show-leak-kinds=all ./target/release/myapp

# Output:
HEAP SUMMARY:
    in use at exit: 1,048,576 bytes in 1,024 blocks
  total heap usage: 5,120 allocs, 4,096 frees, 52,428,800 bytes allocated

LEAK SUMMARY:
   definitely lost: 1,048,576 bytes in 1,024 blocks
```

#### 2. Use Rust-specific Tools
```bash
# Install cargo-flamegraph
cargo install flamegraph

# Generate flamegraph
cargo flamegraph --bin myapp

# Opens flamegraph.svg showing allocation hotspots
```

#### 3. Debug with rust-lldb
```bash
cargo build
rust-lldb target/debug/myapp

(lldb) breakpoint set -n main
(lldb) run
(lldb) memory read --size 8 --format x --count 10 $rsp
```

#### 4. Inspect with AddressSanitizer
```bash
# Rebuild with sanitizer
RUSTFLAGS="-Z sanitizer=address" cargo build --target x86_64-unknown-linux-gnu

# Run
./target/x86_64-unknown-linux-gnu/debug/myapp

# Output shows leak locations
```

### Root Cause
`Rc<RefCell<T>>` circular reference preventing drop.

```rust
// Before (memory leak)
use std::rc::Rc;
use std::cell::RefCell;

struct Node {
    data: String,
    next: Option<Rc<RefCell<Node>>>,
}

fn create_cycle() {
    let node1 = Rc::new(RefCell::new(Node {
        data: "Node 1".to_string(),
        next: None,
    }));

    let node2 = Rc::new(RefCell::new(Node {
        data: "Node 2".to_string(),
        next: Some(Rc::clone(&node1)),
    }));

    node1.borrow_mut().next = Some(Rc::clone(&node2));  // ❌ Cycle!
}  // node1 and node2 never dropped

// After (fixed with Weak)
use std::rc::{Rc, Weak};

struct Node {
    data: String,
    next: Option<Weak<RefCell<Node>>>,  // ✅ Use Weak to break cycle
}

fn create_no_cycle() {
    let node1 = Rc::new(RefCell::new(Node {
        data: "Node 1".to_string(),
        next: None,
    }));

    let node2 = Rc::new(RefCell::new(Node {
        data: "Node 2".to_string(),
        next: Some(Rc::downgrade(&node1)),  // ✅ Weak reference
    }));

    node1.borrow_mut().next = Some(Rc::downgrade(&node2));
}  // Properly cleaned up
```

### Test Case
```rust
// tests/memory_test.rs
#[test]
fn test_no_memory_leak() {
    let initial_allocs = ALLOCATOR.allocated();

    {
        create_no_cycle();
    }  // Scope ends, should free memory

    let final_allocs = ALLOCATOR.allocated();
    assert_eq!(initial_allocs, final_allocs, "Memory leak detected");
}
```

---

## Race Condition Detection (Go)

### Problem
Intermittent test failures; data corruption in concurrent writes.

### Symptoms
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation]
```

### Investigation Steps

#### 1. Enable Race Detector
```bash
# Run tests with race detector
go test -race ./...

# Output:
==================
WARNING: DATA RACE
Write at 0x00c0001a0180 by goroutine 7:
  main.updateCounter()
      /app/counter.go:23 +0x45

Previous read at 0x00c0001a0180 by goroutine 6:
  main.getCounter()
      /app/counter.go:15 +0x38

Goroutine 7 (running) created at:
  main.main()
      /app/main.go:42 +0x120
==================
```

#### 2. Debug with Delve + Race Detector
```bash
# Build with race detection
go build -race -o myapp

# Debug
dlv exec ./myapp

(dlv) break counter.go:23
(dlv) continue
(dlv) print counter
(dlv) goroutines
```

### Root Cause
Unsynchronized access to shared counter.

```go
// Before (race condition)
var counter int

func updateCounter() {
    counter++  // ❌ Not atomic
}

func getCounter() int {
    return counter  // ❌ Unsynchronized read
}

// After (fixed with mutex)
import "sync"

var (
    counter int
    mu      sync.RWMutex
)

func updateCounter() {
    mu.Lock()
    defer mu.Unlock()
    counter++  // ✅ Protected
}

func getCounter() int {
    mu.RLock()
    defer mu.RUnlock()
    return counter  // ✅ Protected
}

// Alternative: Use atomic operations
import "sync/atomic"

var counter int64

func updateCounter() {
    atomic.AddInt64(&counter, 1)  // ✅ Atomic
}

func getCounter() int64 {
    return atomic.LoadInt64(&counter)  // ✅ Atomic
}
```

### Test Case
```go
// counter_test.go
func TestConcurrentCounterAccess(t *testing.T) {
    counter = 0

    var wg sync.WaitGroup
    iterations := 1000
    goroutines := 10

    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < iterations; j++ {
                updateCounter()
            }
        }()
    }

    wg.Wait()
    expected := iterations * goroutines
    assert.Equal(t, int64(expected), getCounter(), "Counter mismatch indicates race condition")
}
```

---

## Null Pointer Debugging (TypeScript)

### Problem
Production error: `Cannot read properties of undefined (reading 'name')`.

### Stack Trace
```
TypeError: Cannot read properties of undefined (reading 'name')
    at getUserName (user.service.ts:15:23)
    at processUser (user.controller.ts:42:10)
    at async Router.handle (express.js:234:15)
```

### Investigation Steps

#### 1. Add Source Maps for Debugging
```json
// tsconfig.json
{
  "compilerOptions": {
    "sourceMap": true,
    "inlineSourceMap": false,
    "outDir": "./dist",
    "rootDir": "./src"
  }
}
```

#### 2. Debug with Node Inspector
```bash
# Run with inspector
node --inspect -r ts-node/register src/app.ts

# Or debug tests
node --inspect-brk node_modules/.bin/jest --runInBand
```

#### 3. VSCode Debugging
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "node",
      "request": "launch",
      "name": "TypeScript: Debug",
      "runtimeArgs": ["-r", "ts-node/register"],
      "args": ["${file}"],
      "cwd": "${workspaceFolder}",
      "sourceMaps": true
    }
  ]
}
```

#### 4. Set Breakpoint and Inspect
```typescript
// user.service.ts
export function getUserName(user: User): string {
    // Set breakpoint here
    debugger;
    return user.name;  // user is undefined!
}

// In debugger console:
// > user
// undefined
// > user?.name
// undefined
```

### Root Cause
Async data not awaited; function called before user fetched.

```typescript
// Before (bug)
async function processUser(userId: string) {
    const user = fetchUser(userId);  // ❌ Missing await
    const name = getUserName(user);  // user is Promise<User>, not User
    console.log(name);
}

// After (fixed)
async function processUser(userId: string) {
    const user = await fetchUser(userId);  // ✅ Await promise
    if (!user) {
        throw new Error(`User ${userId} not found`);
    }
    const name = getUserName(user);
    console.log(name);
}

// Better: Use optional chaining
function getUserName(user: User | undefined): string {
    return user?.name ?? 'Unknown';  // ✅ Safe access
}
```

### Test Case
```typescript
// user.service.test.ts
describe('getUserName', () => {
    it('should handle undefined user gracefully', () => {
        const result = getUserName(undefined);
        expect(result).toBe('Unknown');
    });

    it('should return user name when defined', () => {
        const user = { id: '123', name: 'Alice' };
        const result = getUserName(user);
        expect(result).toBe('Alice');
    });
});
```

---

## Database Query Performance (SQL)

### Problem
Dashboard loads slowly; query taking 5+ seconds.

### Investigation Steps

#### 1. Enable Query Logging
```sql
-- PostgreSQL: Enable slow query log
ALTER SYSTEM SET log_min_duration_statement = 1000;  -- 1 second
SELECT pg_reload_conf();

-- Check logs
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;
```

#### 2. Analyze Query with EXPLAIN
```sql
-- Original slow query
EXPLAIN ANALYZE
SELECT u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at > '2024-01-01'
GROUP BY u.id, u.name
ORDER BY order_count DESC;

-- Output:
Sort  (cost=15234.56..15235.56 rows=400 width=16) (actual time=5234.123..5234.234 rows=1000 loops=1)
  ->  HashAggregate  (cost=15210.00..15214.00 rows=400 width=16) (actual time=5233.456..5233.789 rows=1000 loops=1)
        ->  Hash Left Join  (cost=1234.00..12345.00 rows=286500 width=8) (actual time=123.456..4567.890 rows=500000 loops=1)
              ->  Seq Scan on users u  (cost=0.00..1245.00 rows=50000 width=8) (actual time=0.012..234.567 rows=50000 loops=1)
                    Filter: (created_at > '2024-01-01'::date)
              ->  Hash  (cost=800.00..800.00 rows=30000 width=8) (actual time=123.444..123.444 rows=30000 loops=1)
                    ->  Seq Scan on orders o  (cost=0.00..800.00 rows=30000 width=8) (actual time=0.010..89.123 rows=30000 loops=1)
Planning Time: 1.234 ms
Execution Time: 5234.567 ms  ← SLOW!
```

#### 3. Identify Missing Indexes
```sql
-- Check existing indexes
SELECT tablename, indexname, indexdef
FROM pg_indexes
WHERE tablename IN ('users', 'orders');

-- Missing:
-- 1. Index on users.created_at
-- 2. Index on orders.user_id
```

#### 4. Add Indexes and Re-test
```sql
-- Add indexes
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- Re-run EXPLAIN ANALYZE
EXPLAIN ANALYZE
SELECT u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at > '2024-01-01'
GROUP BY u.id, u.name
ORDER BY order_count DESC;

-- Output (improved):
Sort  (cost=234.56..235.56 rows=400 width=16) (actual time=45.123..45.234 rows=1000 loops=1)
  ->  HashAggregate  (cost=210.00..214.00 rows=400 width=16) (actual time=44.456..44.789 rows=1000 loops=1)
        ->  Hash Left Join  (cost=123.00..200.00 rows=5000 width=8) (actual time=12.345..40.567 rows=5000 loops=1)
              ->  Index Scan using idx_users_created_at on users u  (cost=0.42..100.00 rows=5000 width=8) (actual time=0.023..15.234 rows=5000 loops=1)
                    Index Cond: (created_at > '2024-01-01'::date)
              ->  Hash  (cost=80.00..80.00 rows=3000 width=8) (actual time=12.321..12.321 rows=3000 loops=1)
                    ->  Index Scan using idx_orders_user_id on orders o  (cost=0.42..80.00 rows=3000 width=8) (actual time=0.012..8.123 rows=3000 loops=1)
Execution Time: 45.567 ms  ← 100x FASTER!
```

### Root Cause
Missing indexes on frequently queried columns (`users.created_at`, `orders.user_id`).

### Test Case
```python
# tests/test_query_performance.py
import pytest
import time

def test_dashboard_query_performance(db):
    """Ensure dashboard query completes within 100ms."""
    query = """
        SELECT u.name, COUNT(o.id) as order_count
        FROM users u
        LEFT JOIN orders o ON u.id = o.user_id
        WHERE u.created_at > '2024-01-01'
        GROUP BY u.id, u.name
        ORDER BY order_count DESC
    """

    start = time.time()
    result = db.execute(query)
    duration = time.time() - start

    assert duration < 0.1, f"Query took {duration}s, expected < 0.1s"
    assert len(result) > 0, "Query returned no results"
```

---

## Summary

These real-world examples demonstrate:
1. **Async debugging** with Python debugpy and asyncio inspection
2. **Goroutine leak detection** with Delve and pprof
3. **Distributed tracing** with OpenTelemetry across microservices
4. **Kubernetes debugging** with ephemeral containers and memory profiling
5. **Rust memory leaks** using Valgrind and Weak references
6. **Go race conditions** with built-in race detector
7. **TypeScript null pointers** with optional chaining and type guards
8. **SQL performance** with EXPLAIN ANALYZE and index optimization

Each example follows the **Reproduce → Isolate → Investigate → Fix → Verify** workflow and includes regression tests to prevent recurrence.
