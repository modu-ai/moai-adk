# moai-essentials-debug — 실전 예제

> **Version**: 2.1.0  
> **Last Updated**: 2025-10-27

이 문서는 언어별 스택 트레이스 분석 실전 예제와 디버깅 시나리오를 제공합니다.

---

## Python 디버깅 예제

### 예제 1: TypeError — 타입 불일치

**에러 메시지**:
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

**분석**:
1. **에러 위치**: `app.py:15` (`sum(items)`)
2. **에러 타입**: `TypeError` — 정수와 문자열을 더하려고 시도
3. **실행 경로**: `main()` → `process_data()` → `sum()`
4. **근본 원인**: `items` 리스트에 문자열이 포함됨

**디버깅 단계**:
```python
# 1. 브레이크포인트 설정
import pdb; pdb.set_trace()

# 2. items 내용 확인
(Pdb) print(items)
[1, 2, '3', 4, 5]  # '3'이 문자열!

# 3. 타입 검증
(Pdb) [type(x) for x in items]
[<class 'int'>, <class 'int'>, <class 'str'>, <class 'int'>, <class 'int'>]
```

**해결 방법**:
```python
# Option 1: 입력 데이터 검증
def process_data(items):
    # 타입 체크 및 변환
    items = [int(x) if isinstance(x, str) else x for x in items]
    total = sum(items)
    return total

# Option 2: 타입 힌트 + mypy
def process_data(items: list[int]) -> int:
    total = sum(items)
    return total
```

---

### 예제 2: ImportError — 모듈 미설치

**에러 메시지**:
```
Traceback (most recent call last):
  File "script.py", line 3, in <module>
    import requests
ModuleNotFoundError: No module named 'requests'
```

**분석**:
1. **에러 위치**: `script.py:3`
2. **에러 타입**: `ModuleNotFoundError` — requests 모듈이 설치되지 않음
3. **근본 원인**: 가상환경 활성화 안 됨 또는 의존성 미설치

**디버깅 단계**:
```bash
# 1. 가상환경 확인
which python
# /usr/bin/python (시스템 Python — 잘못됨!)

# 2. 가상환경 활성화
source venv/bin/activate
which python
# /path/to/venv/bin/python (올바름)

# 3. 패키지 설치 확인
pip list | grep requests
# (없음)

# 4. 의존성 설치
pip install requests
```

**해결 방법**:
```bash
# pyproject.toml 또는 requirements.txt에 의존성 명시
# requirements.txt
requests==2.31.0

# 설치
pip install -r requirements.txt
```

---

### 예제 3: AttributeError — 속성 없음

**에러 메시지**:
```
Traceback (most recent call last):
  File "app.py", line 28, in <module>
    result = user.get_profile()
AttributeError: 'NoneType' object has no attribute 'get_profile'
```

**분석**:
1. **에러 위치**: `app.py:28`
2. **에러 타입**: `AttributeError` — `user`가 `None`임
3. **근본 원인**: `user` 객체가 생성되지 않았거나 None 반환

**디버깅 단계**:
```python
# 1. 브레이크포인트 설정
breakpoint()

# 2. user 확인
(Pdb) print(user)
None

# 3. user가 None이 된 이유 추적
(Pdb) where
  app.py(20)main()
  app.py(15)get_user()
  -> return None  # 여기서 None 반환!

# 4. get_user() 함수 확인
def get_user(user_id):
    user = database.find_user(user_id)
    if not user:
        return None  # 문제 발견!
    return user
```

**해결 방법**:
```python
# Option 1: None 체크
user = get_user(user_id)
if user is None:
    print("User not found")
    return
result = user.get_profile()

# Option 2: Optional 타입 힌트
from typing import Optional

def get_user(user_id: int) -> Optional[User]:
    user = database.find_user(user_id)
    return user

# Option 3: 예외 처리
try:
    result = user.get_profile()
except AttributeError:
    print("User is None or has no get_profile method")
```

---

## TypeScript 디버깅 예제

### 예제 1: undefined 접근

**에러 메시지**:
```
TypeError: Cannot read properties of undefined (reading 'name')
    at processUser (app.ts:42:28)
    at Array.map (<anonymous>)
    at getUserNames (app.ts:35:18)
    at main (app.ts:10:5)
```

**분석**:
1. **에러 위치**: `app.ts:42` (`user.name` 접근)
2. **에러 타입**: `TypeError` — `user`가 `undefined`
3. **실행 경로**: `main()` → `getUserNames()` → `map()` → `processUser()`
4. **근본 원인**: 배열에 `undefined` 요소가 포함됨

**코드**:
```typescript
// app.ts
function processUser(user: User) {
  return user.name.toUpperCase();  // 여기서 에러!
}

function getUserNames(users: User[]): string[] {
  return users.map(processUser);
}

const users = [
  { id: 1, name: 'Alice' },
  undefined,  // 문제 발견!
  { id: 2, name: 'Bob' },
];
```

**디버깅 단계**:
```typescript
// 1. 브레이크포인트 설정 (debugger 키워드)
function processUser(user: User) {
  debugger;  // 여기서 중단
  return user.name.toUpperCase();
}

// Chrome DevTools에서:
// user: undefined
```

**해결 방법**:
```typescript
// Option 1: 타입 가드
function processUser(user: User | undefined): string {
  if (!user) {
    return 'Unknown';
  }
  return user.name.toUpperCase();
}

// Option 2: Optional 체이닝
function processUser(user?: User): string {
  return user?.name?.toUpperCase() ?? 'Unknown';
}

// Option 3: 필터링
function getUserNames(users: (User | undefined)[]): string[] {
  return users
    .filter((user): user is User => user !== undefined)
    .map(user => user.name.toUpperCase());
}
```

---

### 예제 2: Promise rejection

**에러 메시지**:
```
UnhandledPromiseRejectionWarning: Error: Network request failed
    at fetchData (api.ts:15:11)
    at async processRequest (handler.ts:28:18)
    at async main (app.ts:12:3)
```

**분석**:
1. **에러 위치**: `api.ts:15`
2. **에러 타입**: `UnhandledPromiseRejectionWarning` — Promise 거부가 처리되지 않음
3. **근본 원인**: `fetchData()`에서 발생한 에러가 catch되지 않음

**코드**:
```typescript
// api.ts
async function fetchData(url: string): Promise<Data> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error('Network request failed');  // 여기서 에러 발생!
  }
  return response.json();
}

// handler.ts
async function processRequest(url: string) {
  const data = await fetchData(url);  // 에러 처리 없음!
  return data;
}
```

**디버깅 단계**:
```typescript
// 1. 브레이크포인트 설정
async function fetchData(url: string): Promise<Data> {
  debugger;
  const response = await fetch(url);
  debugger;  // response 확인
  // response.ok: false
  // response.status: 404
  if (!response.ok) {
    throw new Error('Network request failed');
  }
  return response.json();
}
```

**해결 방법**:
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

// Option 2: Result 타입
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

## Java 디버깅 예제

### 예제 1: NullPointerException

**에러 메시지**:
```
Exception in thread "main" java.lang.NullPointerException: Cannot invoke "User.getName()" because "user" is null
    at com.example.UserService.processUser(UserService.java:42)
    at com.example.UserService.processAllUsers(UserService.java:28)
    at com.example.Main.main(Main.java:15)
```

**분석**:
1. **에러 위치**: `UserService.java:42` (`user.getName()` 호출)
2. **에러 타입**: `NullPointerException` — `user`가 `null`
3. **실행 경로**: `Main.main()` → `processAllUsers()` → `processUser()`
4. **근본 원인**: `user` 객체가 `null`인 채로 메서드 호출 시도

**코드**:
```java
// UserService.java
public class UserService {
    public void processUser(User user) {
        String name = user.getName();  // 여기서 NPE 발생!
        System.out.println("Processing: " + name);
    }
    
    public void processAllUsers(List<User> users) {
        for (User user : users) {
            processUser(user);
        }
    }
}
```

**디버깅 단계**:
```java
// 1. 브레이크포인트 설정 (IntelliJ/Eclipse)
// UserService.java:42에 브레이크포인트

// 2. 변수 확인
// user: null

// 3. 호출 스택 확인
// processUser() ← processAllUsers() ← main()

// 4. users 리스트 확인
// users: [User@123, null, User@456]  // null 발견!
```

**해결 방법**:
```java
// Option 1: Null 체크
public void processUser(User user) {
    if (user == null) {
        System.out.println("User is null");
        return;
    }
    String name = user.getName();
    System.out.println("Processing: " + name);
}

// Option 2: Optional<T> 사용 (Java 8+)
public void processUser(Optional<User> userOpt) {
    userOpt.ifPresent(user -> {
        String name = user.getName();
        System.out.println("Processing: " + name);
    });
}

public void processAllUsers(List<User> users) {
    users.stream()
        .filter(Objects::nonNull)  // null 필터링
        .forEach(this::processUser);
}

// Option 3: @NonNull 어노테이션 (Lombok, Checker Framework)
import lombok.NonNull;

public void processUser(@NonNull User user) {
    String name = user.getName();
    System.out.println("Processing: " + name);
}
```

---

### 예제 2: ClassNotFoundException

**에러 메시지**:
```
Exception in thread "main" java.lang.ClassNotFoundException: com.example.database.DatabaseDriver
    at java.base/java.net.URLClassLoader.findClass(URLClassLoader.java:445)
    at java.base/java.lang.ClassLoader.loadClass(ClassLoader.java:587)
    at java.base/java.lang.Class.forName0(Native Method)
    at java.base/java.lang.Class.forName(Class.java:467)
    at com.example.Main.main(Main.java:12)
```

**분석**:
1. **에러 위치**: `Main.java:12` (`Class.forName()` 호출)
2. **에러 타입**: `ClassNotFoundException` — 클래스를 찾을 수 없음
3. **근본 원인**: JDBC 드라이버 JAR이 classpath에 없음

**디버깅 단계**:
```bash
# 1. Classpath 확인
echo $CLASSPATH

# 2. JAR 파일 확인
ls lib/
# mysql-connector-java-8.0.33.jar (있음)

# 3. 컴파일 및 실행 명령 확인
javac -cp ".:lib/*" Main.java
java -cp ".:lib/*" Main  # classpath에 lib/* 포함 필요!
```

**해결 방법**:
```bash
# Option 1: Classpath 명시
java -cp ".:lib/*" com.example.Main

# Option 2: Maven/Gradle 사용
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

# Option 3: Manifest에 Class-Path 추가
# MANIFEST.MF
Class-Path: lib/mysql-connector-java-8.0.33.jar
```

---

## Go 디버깅 예제

### 예제 1: nil pointer dereference

**에러 메시지**:
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x10a4f20]

goroutine 1 [running]:
main.processUser(...)
    /Users/dev/project/main.go:42
main.main()
    /Users/dev/project/main.go:15 +0x120
```

**분석**:
1. **에러 위치**: `main.go:42` (`processUser` 함수)
2. **에러 타입**: `panic: nil pointer dereference`
3. **근본 원인**: `nil` 포인터에 접근 시도

**코드**:
```go
// main.go
type User struct {
    Name string
    Age  int
}

func processUser(user *User) {
    fmt.Printf("Name: %s\n", user.Name)  // 여기서 panic!
}

func main() {
    var user *User  // nil 포인터
    processUser(user)
}
```

**디버깅 단계**:
```bash
# 1. Delve로 디버깅
dlv debug main.go
(dlv) break main.go:42
(dlv) continue

# 2. 변수 확인
(dlv) print user
nil

# 3. 호출 스택 확인
(dlv) stack
0  0x00000000010a4f20 in main.processUser
   at main.go:42
1  0x00000000010a5040 in main.main
   at main.go:15
```

**해결 방법**:
```go
// Option 1: Nil 체크
func processUser(user *User) {
    if user == nil {
        fmt.Println("User is nil")
        return
    }
    fmt.Printf("Name: %s\n", user.Name)
}

// Option 2: 값 타입 사용
func processUser(user User) {  // 포인터 아님
    fmt.Printf("Name: %s\n", user.Name)
}

// Option 3: 생성자 패턴
func NewUser(name string, age int) *User {
    return &User{Name: name, Age: age}
}

func main() {
    user := NewUser("Alice", 30)  // 항상 non-nil
    processUser(user)
}
```

---

### 예제 2: Goroutine leak

**문제 설명**: 고루틴이 종료되지 않고 계속 실행되어 메모리 누수 발생

**코드**:
```go
// main.go
func worker(ch chan int) {
    for {
        val := <-ch  // 채널이 닫히지 않으면 영원히 대기
        fmt.Println(val)
    }
}

func main() {
    ch := make(chan int)
    go worker(ch)
    
    ch <- 1
    ch <- 2
    // ch를 닫지 않고 종료 → goroutine leak!
}
```

**디버깅 단계**:
```bash
# 1. pprof로 고루틴 확인
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 2. 고루틴 수 확인
(pprof) top
Showing nodes accounting for 1000 goroutines
      flat  flat%   sum%        cum   cum%
      1000 100%   100%       1000 100%  main.worker

# 3. Delve로 고루틴 확인
dlv debug main.go
(dlv) goroutines
[1000 goroutines]

(dlv) goroutine 5
Goroutine 5 - User: main.worker
   main.go:10 (0x10a4f20) (Waiting)
```

**해결 방법**:
```go
// Option 1: 채널 닫기
func main() {
    ch := make(chan int)
    go worker(ch)
    
    ch <- 1
    ch <- 2
    close(ch)  // 채널 닫기
    time.Sleep(time.Second)  // worker가 종료될 시간
}

func worker(ch chan int) {
    for val := range ch {  // 채널이 닫히면 종료
        fmt.Println(val)
    }
}

// Option 2: Context 사용
func worker(ctx context.Context, ch chan int) {
    for {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return  // Context 취소 시 종료
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
    cancel()  // 고루틴 종료
    time.Sleep(time.Second)
}
```

---

## Rust 디버깅 예제

### 예제 1: Borrow checker 에러

**에러 메시지**:
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

**분석**:
1. **에러 위치**: `main.rs:8` (`data.push(4)`)
2. **에러 타입**: Borrow checker 위반 — 불변 참조가 있는데 가변 참조 시도
3. **근본 원인**: `first`가 `data`의 불변 참조를 보유한 상태에서 `data`를 변경하려 함

**코드**:
```rust
// main.rs
fn main() {
    let mut data = vec![1, 2, 3];
    let first = &data[0];  // 불변 참조
    
    data.push(4);  // 에러! 가변 참조 시도
    
    println!("First: {}", first);
}
```

**디버깅 단계**:
```bash
# 1. rust-analyzer로 에러 확인 (VSCode)
# 에러 메시지에 lifetime 정보 포함

# 2. 컴파일러 설명 확인
cargo build --explain E0502
```

**해결 방법**:
```rust
// Option 1: 참조 사용 범위 조정
fn main() {
    let mut data = vec![1, 2, 3];
    let first = data[0];  // 값 복사
    
    data.push(4);  // OK
    
    println!("First: {}", first);
}

// Option 2: 참조를 먼저 해제
fn main() {
    let mut data = vec![1, 2, 3];
    {
        let first = &data[0];
        println!("First: {}", first);
    }  // first가 여기서 드롭됨
    
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

### 예제 2: Panic in thread

**에러 메시지**:
```
thread 'main' panicked at 'index out of bounds: the len is 3 but the index is 5', src/main.rs:5:23
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace
```

**분석**:
1. **에러 위치**: `main.rs:5` (인덱스 접근)
2. **에러 타입**: `panic: index out of bounds`
3. **근본 원인**: 벡터 범위를 벗어난 인덱스 접근

**코드**:
```rust
// main.rs
fn main() {
    let data = vec![1, 2, 3];
    let value = data[5];  // panic!
    println!("Value: {}", value);
}
```

**디버깅 단계**:
```bash
# 1. RUST_BACKTRACE로 전체 스택 트레이스 확인
RUST_BACKTRACE=1 cargo run
# 또는
RUST_BACKTRACE=full cargo run

# 2. rust-lldb로 디버깅
rust-lldb target/debug/myapp
(lldb) breakpoint set --file main.rs --line 5
(lldb) run
(lldb) frame variable
(Vec<i32>) data = vec![1, 2, 3]
(lldb) print data.len()
3
```

**해결 방법**:
```rust
// Option 1: get() 메서드 사용
fn main() {
    let data = vec![1, 2, 3];
    match data.get(5) {
        Some(value) => println!("Value: {}", value),
        None => println!("Index out of bounds"),
    }
}

// Option 2: 범위 체크
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

// Option 3: unwrap_or 사용
fn main() {
    let data = vec![1, 2, 3];
    let value = data.get(5).unwrap_or(&0);  // 기본값 0
    println!("Value: {}", value);
}
```

---

## 컨테이너 디버깅 시나리오

### 시나리오 1: 컨테이너가 즉시 종료됨

**문제**:
```bash
$ docker ps -a
CONTAINER ID   IMAGE     STATUS                     
abc123         myapp     Exited (1) 2 seconds ago
```

**디버깅 단계**:
```bash
# 1. 로그 확인
docker logs abc123
# Error: Database connection failed

# 2. 컨테이너 재시작하지 않고 셸 접속
docker run --rm -it --entrypoint /bin/sh myapp

# 3. 환경 변수 확인
env | grep DB
# DB_HOST=localhost  # 문제 발견! 컨테이너 내에서는 localhost가 아님

# 4. 네트워크 확인
ping db-host
# ping: db-host: Name or service not known
```

**해결 방법**:
```bash
# Option 1: 환경 변수 수정
docker run -e DB_HOST=db-container myapp

# Option 2: Docker Compose로 네트워크 설정
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

### 시나리오 2: Kubernetes Pod CrashLoopBackOff

**문제**:
```bash
$ kubectl get pods
NAME                READY   STATUS             RESTARTS
myapp-pod-abc123    0/1     CrashLoopBackOff   5
```

**디버깅 단계**:
```bash
# 1. Pod 설명 확인
kubectl describe pod myapp-pod-abc123
# Events:
#   Warning  BackOff  kubelet  Back-off restarting failed container

# 2. 로그 확인
kubectl logs myapp-pod-abc123
# panic: open /config/app.yaml: no such file or directory

# 3. 이전 컨테이너 로그 확인
kubectl logs myapp-pod-abc123 --previous

# 4. ConfigMap 확인
kubectl get configmap myapp-config -o yaml
# (ConfigMap이 없거나 잘못 마운트됨)

# 5. 볼륨 마운트 확인
kubectl get pod myapp-pod-abc123 -o yaml | grep -A 10 volumeMounts
```

**해결 방법**:
```yaml
# 1. ConfigMap 생성
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  app.yaml: |
    database:
      host: db-service
      port: 5432

# 2. Pod에 ConfigMap 마운트
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

## 분산 시스템 디버깅 시나리오

### 시나리오: 마이크로서비스 간 타임아웃

**문제**: Service A → Service B 호출 시 타임아웃 발생

**디버깅 단계**:

**1. OpenTelemetry로 트레이스 수집**:
```python
# Service A
from opentelemetry import trace

tracer = trace.get_tracer(__name__)

with tracer.start_as_current_span("call-service-b") as span:
    response = requests.get("http://service-b/api/data", timeout=5)
    span.set_attribute("http.status_code", response.status_code)
```

**2. Jaeger UI에서 트레이스 분석**:
```
Trace: request-123
├─ Service A: call-service-b (50ms)
│  └─ HTTP GET http://service-b/api/data
│     ├─ DNS lookup: 10ms
│     ├─ TCP connect: 15ms
│     └─ Waiting for response: 5000ms ← 타임아웃!
└─ Service B: process-request (4950ms)
   ├─ Database query: 4900ms ← 병목!
   └─ Response serialization: 50ms
```

**3. 근본 원인 식별**:
- Service B의 데이터베이스 쿼리가 4.9초 소요
- Service A의 타임아웃이 5초로 설정되어 있어 경합 상태 발생

**해결 방법**:
```python
# Option 1: Service B의 쿼리 최적화
# 인덱스 추가
CREATE INDEX idx_user_email ON users(email);

# Option 2: Service A의 타임아웃 증가
response = requests.get("http://service-b/api/data", timeout=10)

# Option 3: 캐싱 추가
from redis import Redis

cache = Redis(host='redis', port=6379)

def get_data():
    cached = cache.get('data')
    if cached:
        return cached
    
    data = expensive_database_query()
    cache.setex('data', 300, data)  # 5분 캐시
    return data
```

---

## 성능 디버깅 시나리오

### 시나리오: 느린 API 응답

**문제**: API 엔드포인트 응답 시간이 3초 이상

**디버깅 단계**:

**1. py-spy로 CPU 프로파일링** (Python):
```bash
py-spy record -o profile.svg --pid <pid>
```

**2. Flamegraph 분석**:
```
main() [100%]
├─ process_request() [95%]
│  ├─ load_users() [80%]  ← 병목!
│  │  └─ database.query() [78%]
│  └─ serialize_response() [15%]
└─ logging() [5%]
```

**3. 데이터베이스 쿼리 분석**:
```sql
EXPLAIN ANALYZE SELECT * FROM users WHERE status = 'active';
-- Seq Scan on users (cost=0.00..1234.56)
-- 인덱스 없음!
```

**4. 해결 방법**:
```sql
-- 인덱스 추가
CREATE INDEX idx_users_status ON users(status);

-- 쿼리 재실행
EXPLAIN ANALYZE SELECT * FROM users WHERE status = 'active';
-- Index Scan using idx_users_status (cost=0.28..45.67)
-- 응답 시간: 3초 → 50ms
```

---

## 요약: 디버깅 체크리스트

### 1. 재현 (Reproduce)
- [ ] 최소 재현 예제 (MRE) 작성
- [ ] 일관된 재현 단계 문서화
- [ ] 환경 정보 기록 (OS, 언어 버전, 의존성)

### 2. 격리 (Isolate)
- [ ] 이진 탐색으로 문제 범위 좁히기
- [ ] 최근 변경사항 확인 (git diff, git log)
- [ ] 입력 데이터 및 엣지 케이스 검증

### 3. 조사 (Investigate)
- [ ] 스택 트레이스를 아래에서 위로 읽기
- [ ] 주요 결정 지점에 로깅 추가
- [ ] 에러 위치 이전에 브레이크포인트 설정
- [ ] 디버거에서 변수 상태 확인

### 4. 가설 (Hypothesize)
- [ ] 근본 원인에 대한 이론 수립
- [ ] 가장 가능성 높은 원인 2-3개 식별
- [ ] 가설 검증 실험 설계

### 5. 수정 (Fix)
- [ ] 최소한의 수정 먼저 구현
- [ ] 회귀 테스트 추가 (RED → GREEN)
- [ ] 필요시 리팩토링 (REFACTOR 단계)
- [ ] 문서 업데이트

### 6. 검증 (Verify)
- [ ] 전체 테스트 스위트 실행
- [ ] 엣지 케이스 명시적 테스트
- [ ] 프로덕션 유사 환경에서 수정 검증
- [ ] 재발 모니터링

---

**End of Examples** | moai-essentials-debug v2.1.0
