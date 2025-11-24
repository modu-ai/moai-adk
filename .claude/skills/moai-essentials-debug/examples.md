# moai-essentials-debug — Practical Examples

> **Version**: 2.1.0
> **Last Updated**: 2025-10-27

This document provides practical examples and debugging scenarios for stack trace analysis across programming languages.

---

## Python Debugging Examples

### Example 1: TypeError — Type Mismatch

**Error Message**:
```
Traceback (most recent call last):
  File "app.py", line 45, in <module>
    main()
  File "app.py", line 38, in main
    result = process_data(data)
  File "app.py", line 15, in process_data
    total = sum(items)
TypeError: unsupported operand type(s) for +: 'int' and 'str'
```

**Analysis**:
1. **Error Location**: `app.py:15` (`sum(items)`)
2. **Error Type**: `TypeError` — attempting to add integer and string
3. **Execution Path**: `main()` → `process_data()` → `sum()`
4. **Root Cause**: `items` list contains strings

**Debugging Steps**:
```python
# 1. Set breakpoint
import pdb; pdb.set_trace()

# 2. Check items content
(Pdb) print(items)
[1, 2, '3', 4, 5]  # '3' is a string!

# 3. Type validation
(Pdb) [type(x) for x in items]
[<class 'int'>, <class 'int'>, <class 'str'>, <class 'int'>, <class 'int'>]
```

**Solution**:
```python
# Option 1: Input data validation
def process_data(items):
    # Type check and conversion
    items = [int(x) if isinstance(x, str) else x for x in items]
    total = sum(items)
    return total

# Option 2: Type hints + mypy
def process_data(items: list[int]) -> int:
    total = sum(items)
    return total
```

---

### Example 2: ImportError — Module Not Installed

**Error Message**:
```
Traceback (most recent call last):
  File "script.py", line 3, in <module>
    import requests
ModuleNotFoundError: No module named 'requests'
```

**Analysis**:
1. **Error Location**: `script.py:3`
2. **Error Type**: `ModuleNotFoundError` — requests module not installed
3. **Root Cause**: Virtual environment not activated or dependency not installed

**Debugging Steps**:
```bash
# 1. Check virtual environment
which python
# /usr/bin/python (system Python — wrong!)

# 2. Activate virtual environment
source venv/bin/activate
which python
# /path/to/venv/bin/python (correct)

# 3. Check package installation
pip list | grep requests
# (not found)

# 4. Install dependency
pip install requests
```

**Solution**:
```bash
# Specify dependency in pyproject.toml or requirements.txt
# requirements.txt
requests==2.31.0

# Install
pip install -r requirements.txt
```

---

### Example 3: AttributeError — Missing Attribute

**Error Message**:
```
Traceback (most recent call last):
  File "app.py", line 28, in <module>
    result = user.get_profile()
AttributeError: 'NoneType' object has no attribute 'get_profile'
```

**Analysis**:
1. **Error Location**: `app.py:28`
2. **Error Type**: `AttributeError` — `user` is `None`
3. **Root Cause**: `user` object not created or None returned

**Debugging Steps**:
```python
# 1. Set breakpoint
breakpoint()

# 2. Check user
(Pdb) print(user)
None

# 3. Trace why user became None
(Pdb) where
  app.py(20)main()
  app.py(15)get_user()
  -> return None  # Here's the problem!

# 4. Check get_user() function
def get_user(user_id):
    user = database.find_user(user_id)
    if not user:
        return None  # Problem found!
    return user
```

**Solution**:
```python
# Option 1: None check
user = get_user(user_id)
if user is None:
    print("User not found")
    return
result = user.get_profile()

# Option 2: Optional type hint
from typing import Optional

def get_user(user_id: int) -> Optional[User]:
    user = database.find_user(user_id)
    return user

# Option 3: Exception handling
try:
    result = user.get_profile()
except AttributeError:
    print("User is None or has no get_profile method")
```

---

## TypeScript Debugging Examples

### Example 1: undefined Access

**Error Message**:
```
TypeError: Cannot read properties of undefined (reading 'name')
    at processUser (app.ts:42:28)
    at Array.map (<anonymous>)
    at getUserNames (app.ts:35:18)
    at main (app.ts:10:5)
```

**Analysis**:
1. **Error Location**: `app.ts:42` (`user.name` access)
2. **Error Type**: `TypeError` — `user` is `undefined`
3. **Execution Path**: `main()` → `getUserNames()` → `map()` → `processUser()`
4. **Root Cause**: Array contains `undefined` element

**Code**:
```typescript
// app.ts
function processUser(user: User) {
  return user.name.toUpperCase();  // Error here!
}

function getUserNames(users: User[]): string[] {
  return users.map(processUser);
}

const users = [
  { id: 1, name: 'Alice' },
  undefined,  // Problem found!
  { id: 2, name: 'Bob' },
];
```

**Debugging Steps**:
```typescript
// 1. Set breakpoint (debugger keyword)
function processUser(user: User) {
  debugger;  // Break here
  return user.name.toUpperCase();
}

// In Chrome DevTools:
// user: undefined
```

**Solution**:
```typescript
// Option 1: Type guard
function processUser(user: User | undefined): string {
  if (!user) {
    return 'Unknown';
  }
  return user.name.toUpperCase();
}

// Option 2: Optional chaining
function processUser(user?: User): string {
  return user?.name?.toUpperCase() ?? 'Unknown';
}

// Option 3: Filtering
function getUserNames(users: (User | undefined)[]): string[] {
  return users
    .filter((user): user is User => user !== undefined)
    .map(user => user.name.toUpperCase());
}
```

---

### Example 2: Promise Rejection

**Error Message**:
```
UnhandledPromiseRejectionWarning: Error: Network request failed
    at fetchData (api.ts:15:11)
    at async processRequest (handler.ts:28:18)
    at async main (app.ts:12:3)
```

**Analysis**:
1. **Error Location**: `api.ts:15`
2. **Error Type**: `UnhandledPromiseRejectionWarning` — Promise rejection not handled
3. **Root Cause**: Error from `fetchData()` not caught

**Code**:
```typescript
// api.ts
async function fetchData(url: string): Promise<Data> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error('Network request failed');  // Error thrown here!
  }
  return response.json();
}

// handler.ts
async function processRequest(url: string) {
  const data = await fetchData(url);  // No error handling!
  return data;
}
```

**Debugging Steps**:
```typescript
// 1. Set breakpoint
async function fetchData(url: string): Promise<Data> {
  debugger;
  const response = await fetch(url);
  debugger;  // Check response
  // response.ok: false
  // response.status: 404
  if (!response.ok) {
    throw new Error('Network request failed');
  }
  return response.json();
}
```

**Solution**:
```typescript
// Option 1: try-catch
async function processRequest(url: string): Promise<Data | null> {
  try {
    const data = await fetchData(url);
    return data;
  } catch (error) {
    console.error('Failed to fetch data:', error);
    return null;
  }
}

// Option 2: Result type
type Result<T> = { success: true; data: T } | { success: false; error: Error };

async function fetchData(url: string): Promise<Result<Data>> {
  try {
    const response = await fetch(url);
    if (!response.ok) {
      return { success: false, error: new Error('Network request failed') };
    }
    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    return { success: false, error: error as Error };
  }
}
```

---

## Java Debugging Examples

### Example 1: NullPointerException

**Error Message**:
```
Exception in thread "main" java.lang.NullPointerException: Cannot invoke "User.getName()" because "user" is null
    at com.example.UserService.processUser(UserService.java:42)
    at com.example.UserService.processAllUsers(UserService.java:28)
    at com.example.Main.main(Main.java:15)
```

**Analysis**:
1. **Error Location**: `UserService.java:42` (`user.getName()` call)
2. **Error Type**: `NullPointerException` — `user` is `null`
3. **Execution Path**: `Main.main()` → `processAllUsers()` → `processUser()`
4. **Root Cause**: Attempting to call method on `null` user object

**Code**:
```java
// UserService.java
public class UserService {
    public void processUser(User user) {
        String name = user.getName();  // NPE occurs here!
        System.out.println("Processing: " + name);
    }

    public void processAllUsers(List<User> users) {
        for (User user : users) {
            processUser(user);
        }
    }
}
```

**Debugging Steps**:
```java
// 1. Set breakpoint (IntelliJ/Eclipse)
// Set breakpoint at UserService.java:42

// 2. Check variable
// user: null

// 3. Check call stack
// processUser() ← processAllUsers() ← main()

// 4. Check users list
// users: [User@123, null, User@456]  // null found!
```

**Solution**:
```java
// Option 1: Null check
public void processUser(User user) {
    if (user == null) {
        System.out.println("User is null");
        return;
    }
    String name = user.getName();
    System.out.println("Processing: " + name);
}

// Option 2: Optional<T> (Java 8+)
public void processUser(Optional<User> userOpt) {
    userOpt.ifPresent(user -> {
        String name = user.getName();
        System.out.println("Processing: " + name);
    });
}

public void processAllUsers(List<User> users) {
    users.stream()
        .filter(Objects::nonNull)  // Filter null
        .forEach(this::processUser);
}

// Option 3: @NonNull annotation (Lombok, Checker Framework)
import lombok.NonNull;

public void processUser(@NonNull User user) {
    String name = user.getName();
    System.out.println("Processing: " + name);
}
```

---

### Example 2: ClassNotFoundException

**Error Message**:
```
Exception in thread "main" java.lang.ClassNotFoundException: com.example.database.DatabaseDriver
    at java.base/java.net.URLClassLoader.findClass(URLClassLoader.java:445)
    at java.base/java.lang.ClassLoader.loadClass(ClassLoader.java:587)
    at java.base/java.lang.Class.forName0(Native Method)
    at java.base/java.lang.Class.forName(Class.java:467)
    at com.example.Main.main(Main.java:12)
```

**Analysis**:
1. **Error Location**: `Main.java:12` (`Class.forName()` call)
2. **Error Type**: `ClassNotFoundException` — class not found
3. **Root Cause**: JDBC driver JAR not in classpath

**Debugging Steps**:
```bash
# 1. Check classpath
echo $CLASSPATH

# 2. Check JAR file
ls lib/
# mysql-connector-java-8.0.33.jar (exists)

# 3. Check compile and run command
javac -cp ".:lib/*" Main.java
java -cp ".:lib/*" Main  # Need lib/* in classpath!
```

**Solution**:
```bash
# Option 1: Specify classpath
java -cp ".:lib/*" com.example.Main

# Option 2: Use Maven/Gradle
# pom.xml (Maven)
<dependency>
    <groupId>mysql</groupId>
    <artifactId>mysql-connector-java</artifactId>
    <version>8.0.33</version>
</dependency>

# build.gradle (Gradle)
dependencies {
    implementation 'mysql:mysql-connector-java:8.0.33'
}

# Option 3: Add Class-Path to Manifest
# MANIFEST.MF
Class-Path: lib/mysql-connector-java-8.0.33.jar
```

---

## Go Debugging Examples

### Example 1: nil Pointer Dereference

**Error Message**:
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x10a4f20]

goroutine 1 [running]:
main.processUser(...)
    /Users/dev/project/main.go:42
main.main()
    /Users/dev/project/main.go:15 +0x120
```

**Analysis**:
1. **Error Location**: `main.go:42` (`processUser` function)
2. **Error Type**: `panic: nil pointer dereference`
3. **Root Cause**: Accessing `nil` pointer

**Code**:
```go
// main.go
type User struct {
    Name string
    Age  int
}

func processUser(user *User) {
    fmt.Printf("Name: %s\n", user.Name)  // Panic here!
}

func main() {
    var user *User  // nil pointer
    processUser(user)
}
```

**Debugging Steps**:
```bash
# 1. Debug with Delve
dlv debug main.go
(dlv) break main.go:42
(dlv) continue

# 2. Check variable
(dlv) print user
nil

# 3. Check call stack
(dlv) stack
0  0x00000000010a4f20 in main.processUser
   at main.go:42
1  0x00000000010a5040 in main.main
   at main.go:15
```

**Solution**:
```go
// Option 1: Nil check
func processUser(user *User) {
    if user == nil {
        fmt.Println("User is nil")
        return
    }
    fmt.Printf("Name: %s\n", user.Name)
}

// Option 2: Use value type
func processUser(user User) {  // Not a pointer
    fmt.Printf("Name: %s\n", user.Name)
}

// Option 3: Constructor pattern
func NewUser(name string, age int) *User {
    return &User{Name: name, Age: age}
}

func main() {
    user := NewUser("Alice", 30)  // Always non-nil
    processUser(user)
}
```

---

### Example 2: Goroutine Leak

**Problem Description**: Goroutines not terminating, causing memory leak

**Code**:
```go
// main.go
func worker(ch chan int) {
    for {
        val := <-ch  // Waits forever if channel not closed
        fmt.Println(val)
    }
}

func main() {
    ch := make(chan int)
    go worker(ch)

    ch <- 1
    ch <- 2
    // Not closing ch before exit → goroutine leak!
}
```

**Debugging Steps**:
```bash
# 1. Check goroutines with pprof
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 2. Check goroutine count
(pprof) top
Showing nodes accounting for 1000 goroutines
      flat  flat%   sum%        cum   cum%
      1000 100%   100%       1000 100%  main.worker

# 3. Check with Delve
dlv debug main.go
(dlv) goroutines
[1000 goroutines]

(dlv) goroutine 5
Goroutine 5 - User: main.worker
   main.go:10 (0x10a4f20) (Waiting)
```

**Solution**:
```go
// Option 1: Close channel
func main() {
    ch := make(chan int)
    go worker(ch)

    ch <- 1
    ch <- 2
    close(ch)  // Close channel
    time.Sleep(time.Second)  // Wait for worker to finish
}

func worker(ch chan int) {
    for val := range ch {  // Terminates when channel closed
        fmt.Println(val)
    }
}

// Option 2: Use Context
func worker(ctx context.Context, ch chan int) {
    for {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return  // Terminate on context cancellation
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    ch := make(chan int)
    go worker(ctx, ch)

    ch <- 1
    ch <- 2
    cancel()  // Terminate goroutine
    time.Sleep(time.Second)
}
```

---

## Rust Debugging Examples

### Example 1: Borrow Checker Error

**Error Message**:
```
error[E0502]: cannot borrow `data` as mutable because it is also borrowed as immutable
  --> src/main.rs:8:5
   |
6  |     let first = &data[0];
   |                 -------- immutable borrow occurs here
7  |
8  |     data.push(4);
   |     ^^^^^^^^^^^^ mutable borrow occurs here
9  |
10 |     println!("First: {}", first);
   |                           ----- immutable borrow later used here
```

**Analysis**:
1. **Error Location**: `main.rs:8` (`data.push(4)`)
2. **Error Type**: Borrow checker violation — attempting mutable borrow while immutable reference exists
3. **Root Cause**: `first` holds immutable reference to `data` while trying to mutate `data`

**Code**:
```rust
// main.rs
fn main() {
    let mut data = vec![1, 2, 3];
    let first = &data[0];  // Immutable reference

    data.push(4);  // Error! Attempting mutable reference

    println!("First: {}", first);
}
```

**Debugging Steps**:
```bash
# 1. Check error with rust-analyzer (VSCode)
# Error message includes lifetime information

# 2. Check compiler explanation
cargo build --explain E0502
```

**Solution**:
```rust
// Option 1: Adjust reference scope
fn main() {
    let mut data = vec![1, 2, 3];
    let first = data[0];  // Copy value

    data.push(4);  // OK

    println!("First: {}", first);
}

// Option 2: Release reference first
fn main() {
    let mut data = vec![1, 2, 3];
    {
        let first = &data[0];
        println!("First: {}", first);
    }  // first dropped here

    data.push(4);  // OK
}

// Option 3: Clone
fn main() {
    let mut data = vec![1, 2, 3];
    let first = data.get(0).cloned();  // Option<i32>

    data.push(4);  // OK

    if let Some(first) = first {
        println!("First: {}", first);
    }
}
```

---

### Example 2: Panic in Thread

**Error Message**:
```
thread 'main' panicked at 'index out of bounds: the len is 3 but the index is 5', src/main.rs:5:23
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
```

**Analysis**:
1. **Error Location**: `main.rs:5` (index access)
2. **Error Type**: `panic: index out of bounds`
3. **Root Cause**: Accessing index beyond vector bounds

**Code**:
```rust
// main.rs
fn main() {
    let data = vec![1, 2, 3];
    let value = data[5];  // panic!
    println!("Value: {}", value);
}
```

**Debugging Steps**:
```bash
# 1. Check full stack trace with RUST_BACKTRACE
RUST_BACKTRACE=1 cargo run
# or
RUST_BACKTRACE=full cargo run

# 2. Debug with rust-lldb
rust-lldb target/debug/myapp
(lldb) breakpoint set --file main.rs --line 5
(lldb) run
(lldb) frame variable
(Vec<i32>) data = vec![1, 2, 3]
(lldb) print data.len()
3
```

**Solution**:
```rust
// Option 1: Use get() method
fn main() {
    let data = vec![1, 2, 3];
    match data.get(5) {
        Some(value) => println!("Value: {}", value),
        None => println!("Index out of bounds"),
    }
}

// Option 2: Bounds check
fn main() {
    let data = vec![1, 2, 3];
    let index = 5;

    if index < data.len() {
        let value = data[index];
        println!("Value: {}", value);
    } else {
        println!("Index out of bounds");
    }
}

// Option 3: Use unwrap_or
fn main() {
    let data = vec![1, 2, 3];
    let value = data.get(5).unwrap_or(&0);  // Default value 0
    println!("Value: {}", value);
}
```

---

## Container Debugging Scenarios

### Scenario 1: Container Exits Immediately

**Problem**:
```bash
$ docker ps -a
CONTAINER ID   IMAGE     STATUS
abc123         myapp     Exited (1) 2 seconds ago
```

**Debugging Steps**:
```bash
# 1. Check logs
docker logs abc123
# Error: Database connection failed

# 2. Access container shell without restart
docker run --rm -it --entrypoint /bin/sh myapp

# 3. Check environment variables
env | grep DB
# DB_HOST=localhost  # Problem found! localhost not correct in container

# 4. Check network
ping db-host
# ping: db-host: Name or service not known
```

**Solution**:
```bash
# Option 1: Modify environment variable
docker run -e DB_HOST=db-container myapp

# Option 2: Configure network with Docker Compose
# docker-compose.yml
version: '3'
services:
  app:
    image: myapp
    environment:
      DB_HOST: db
    depends_on:
      - db
  db:
    image: postgres:15
```

---

### Scenario 2: Kubernetes Pod CrashLoopBackOff

**Problem**:
```bash
$ kubectl get pods
NAME                READY   STATUS             RESTARTS
myapp-pod-abc123    0/1     CrashLoopBackOff   5
```

**Debugging Steps**:
```bash
# 1. Check pod description
kubectl describe pod myapp-pod-abc123
# Events:
#   Warning  BackOff  kubelet  Back-off restarting failed container

# 2. Check logs
kubectl logs myapp-pod-abc123
# panic: open /config/app.yaml: no such file or directory

# 3. Check previous container logs
kubectl logs myapp-pod-abc123 --previous

# 4. Check ConfigMap
kubectl get configmap myapp-config -o yaml
# (ConfigMap missing or incorrectly mounted)

# 5. Check volume mount
kubectl get pod myapp-pod-abc123 -o yaml | grep -A 10 volumeMounts
```

**Solution**:
```yaml
# 1. Create ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  app.yaml: |
    database:
      host: db-service
      port: 5432

# 2. Mount ConfigMap to Pod
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
spec:
  containers:
  - name: myapp
    image: myapp:latest
    volumeMounts:
    - name: config
      mountPath: /config
  volumes:
  - name: config
    configMap:
      name: myapp-config
```

---

## Distributed System Debugging Scenarios

### Scenario: Timeout Between Microservices

**Problem**: Timeout occurs when Service A → Service B call

**Debugging Steps**:

**1. Collect traces with OpenTelemetry**:
```python
# Service A
from opentelemetry import trace

tracer = trace.get_tracer(__name__)

with tracer.start_as_current_span("call-service-b") as span:
    response = requests.get("http://service-b/api/data", timeout=5)
    span.set_attribute("http.status_code", response.status_code)
```

**2. Analyze trace in Jaeger UI**:
```
Trace: request-123
├─ Service A: call-service-b (50ms)
│  └─ HTTP GET http://service-b/api/data
│     ├─ DNS lookup: 10ms
│     ├─ TCP connect: 15ms
│     └─ Waiting for response: 5000ms ← Timeout!
└─ Service B: process-request (4950ms)
   ├─ Database query: 4900ms ← Bottleneck!
   └─ Response serialization: 50ms
```

**3. Identify root cause**:
- Service B's database query takes 4.9 seconds
- Service A's timeout set to 5 seconds, causing race condition

**Solution**:
```python
# Option 1: Optimize Service B query
# Add index
CREATE INDEX idx_user_email ON users(email);

# Option 2: Increase Service A timeout
response = requests.get("http://service-b/api/data", timeout=10)

# Option 3: Add caching
from redis import Redis

cache = Redis(host='redis', port=6379)

def get_data():
    cached = cache.get('data')
    if cached:
        return cached

    data = expensive_database_query()
    cache.setex('data', 300, data)  # 5 minute cache
    return data
```

---

## Performance Debugging Scenarios

### Scenario: Slow API Response

**Problem**: API endpoint response time exceeds 3 seconds

**Debugging Steps**:

**1. CPU profiling with py-spy** (Python):
```bash
py-spy record -o profile.svg --pid <pid>
```

**2. Analyze Flamegraph**:
```
main() [100%]
├─ process_request() [95%]
│  ├─ load_users() [80%]  ← Bottleneck!
│  │  └─ database.query() [78%]
│  └─ serialize_response() [15%]
└─ logging() [5%]
```

**3. Analyze database query**:
```sql
EXPLAIN ANALYZE SELECT * FROM users WHERE status = 'active';
-- Seq Scan on users (cost=0.00..1234.56)
-- No index!
```

**4. Solution**:
```sql
-- Add index
CREATE INDEX idx_users_status ON users(status);

-- Re-run query
EXPLAIN ANALYZE SELECT * FROM users WHERE status = 'active';
-- Index Scan using idx_users_status (cost=0.28..45.67)
-- Response time: 3s → 50ms
```

---

## Summary: Debugging Checklist

### 1. Reproduce
- [ ] Create minimal reproducible example (MRE)
- [ ] Document consistent reproduction steps
- [ ] Record environment info (OS, language version, dependencies)

### 2. Isolate
- [ ] Narrow problem scope with binary search
- [ ] Check recent changes (git diff, git log)
- [ ] Validate input data and edge cases

### 3. Investigate
- [ ] Read stack trace from bottom to top
- [ ] Add logging at key decision points
- [ ] Set breakpoint before error location
- [ ] Check variable state in debugger

### 4. Hypothesize
- [ ] Establish theory about root cause
- [ ] Identify 2-3 most likely causes
- [ ] Design experiment to validate hypothesis

### 5. Fix
- [ ] Implement minimal fix first
- [ ] Add regression test (RED → GREEN)
- [ ] Refactor if needed (REFACTOR step)
- [ ] Update documentation

### 6. Verify
- [ ] Run full test suite
- [ ] Explicitly test edge cases
- [ ] Verify fix in production-like environment
- [ ] Monitor for recurrence

---

**End of Examples** | moai-essentials-debug v2.1.0
