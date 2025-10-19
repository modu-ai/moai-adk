---
name: moai-essentials-debug
description: Advanced debugging with stack trace analysis, error pattern detection, and fix suggestions
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Debugger Pro

## What it does

Advanced debugging support with stack trace analysis, common error pattern detection, and actionable fix suggestions.

## When to use

- "에러 해결해줘", "이 오류 원인은?", "스택 트레이스 분석"
- Automatically invoked on runtime errors (via debug-helper sub-agent)
- "왜 안 돼?", "NullPointerException 해결"

## How it works

**Stack Trace Analysis**:
```python
# Error example
jwt.exceptions.ExpiredSignatureError: Signature has expired

# Alfred Analysis
📍 Error Location: src/auth/service.py:142
🔍 Root Cause: JWT token has expired
💡 Fix Suggestion:
   1. Implement token refresh logic
   2. Check expiration before validation
   3. Handle ExpiredSignatureError gracefully
```

**Common Error Patterns**:
- `NullPointerException` → Optional usage, guard clauses
- `IndexError` → Boundary checks
- `KeyError` → `.get()` with defaults
- `TypeError` → Type hints, input validation
- `ConnectionError` → Retry logic, timeouts

**Debugging Checklist**:
- [ ] Reproducible?
- [ ] Log messages?
- [ ] Input data?
- [ ] Recent changes?
- [ ] Dependency versions?

**Language-specific Tips**:
- **Python**: Logging, type guards
- **TypeScript**: Type guards, null checks
- **Java**: Optional, try-with-resources

## Language-specific Debugging Tools

| Language | Tool | Command | Use Case |
|----------|------|---------|----------|
| **Python** | pdb | `python -m pdb script.py` | Interactive debugging, breakpoints |
| **Python** | logging | `logging.debug()` | Trace execution flow |
| **TypeScript** | Chrome DevTools | `node --inspect app.ts` | Breakpoints, profiling |
| **TypeScript** | console | `console.log(), console.table()` | Quick debugging |
| **Java** | jdb | `jdb -attach 5005` | Remote debugging |
| **Java** | IntelliJ Debugger | IDE built-in | GUI debugging |
| **Go** | Delve | `dlv debug ./cmd/main.go` | Step-through debugging |
| **Go** | pprof | `go tool pprof http://localhost:6060/debug/pprof/heap` | Memory profiling |
| **Rust** | gdb | `gdb ./target/debug/app` | Low-level debugging |
| **Rust** | println! | `println!("{:?}", variable)` | Variable inspection |

## Examples

### Example 1: Python NullPointerException → Option with Guard Clause

**❌ Before (Runtime Error)**:
```python
# @CODE:ERROR-HANDLING-001
def get_user_email(user_data):
    # 위험: user_data가 None이면 AttributeError 발생
    return user_data['email'].lower()

# 실행 시:
# TypeError: 'NoneType' object is not subscriptable
# File "app.py", line 5, in get_user_email
```

**✅ After (Safe with Guard Clause)**:
```python
# @CODE:ERROR-HANDLING-001: Optional handling
def get_user_email(user_data: Optional[Dict]) -> Optional[str]:
    """@CODE:ERROR-HANDLING-001: 사용자 이메일 안전 조회"""
    # 가드절: None 확인
    if not user_data:
        return None

    # 안전한 접근
    email = user_data.get('email')
    return email.lower() if email else None

# 테스트:
assert get_user_email(None) is None              # ✅ PASS
assert get_user_email({'email': 'USER@EXAMPLE.COM'}) == 'user@example.com'  # ✅ PASS
```

**디버깅 과정**:
```bash
# 1단계: 원인 파악
$ python -m pdb app.py
(Pdb) l
(Pdb) p user_data  # None 확인

# 2단계: 수정
$ git diff
-    return user_data['email'].lower()
+    if not user_data: return None

# 3단계: 테스트
$ pytest tests/test_error_handling.py -v
```

### Example 2: TypeScript N+1 Query + Memory Leak (Database)

**❌ Before (Performance Issue)**:
```typescript
// @CODE:PERF-N-PLUS-ONE
async function getUsersWithOrders(userIds: number[]): Promise<UserOrder[]> {
    const users = [];

    // N+1 문제: userIds 배열 크기만큼 DB 쿼리 실행
    for (const userId of userIds) {
        // 문제 1: Loop 내 쿼리
        const user = await db.query('SELECT * FROM users WHERE id = $1', [userId]);

        // 문제 2: 각 사용자마다 추가 쿼리
        const orders = await db.query('SELECT * FROM orders WHERE user_id = $1', [userId]);

        users.push({ user: user[0], orders });
    }

    return users;
}

// 성능: 10명 사용자 = 1 + 10 쿼리 = 약 5.2초 ⏱️
// 메모리: 연결 누수 (await 없이 실행 가능)
```

**✅ After (Optimized with Eager Loading)**:
```typescript
// @CODE:PERF-N-PLUS-ONE: JOIN을 사용한 최적화
async function getUsersWithOrders(userIds: number[]): Promise<UserOrder[]> {
    // 최적화: 단일 JOIN 쿼리로 모든 데이터 조회
    const query = `
        SELECT
            u.id, u.name, u.email,
            json_agg(json_build_object(
                'id', o.id,
                'amount', o.amount
            )) as orders
        FROM users u
        LEFT JOIN orders o ON u.id = o.user_id
        WHERE u.id = ANY($1)
        GROUP BY u.id
    `;

    const result = await db.query(query, [userIds]);

    return result.rows.map(row => ({
        user: { id: row.id, name: row.name, email: row.email },
        orders: row.orders || []
    }));
}

// 성능: 10명 사용자 = 1 쿼리만 실행 = 약 0.3초 ⏱️ (94% 개선!)
// 메모리: 연결 재사용 + 명시적 정리
```

**디버깅 및 검증**:
```bash
# 1단계: 성능 프로파일링
$ node --inspect app.ts
$ open chrome://inspect

# 2단계: 데이터베이스 쿼리 로그 확인
$ QUERY_DEBUG=1 npm test
// 개선 전: 11개 쿼리
// 개선 후: 1개 쿼리

# 3단계: 성능 벤치마크
$ npm run benchmark
Before: 5.2s (10 users)
After:  0.3s (10 users) ✅ 1633% 빠름
```

### Example 3: Java NullPointerException → Optional Pattern

**❌ Before (Risky)**:
```java
// @CODE:OPTIONAL-HANDLING
public String getUserCity(int userId) {
    User user = userRepository.findById(userId);
    // 위험: user가 null이면 NullPointerException
    return user.getAddress().getCity().toUpperCase();

    // java.lang.NullPointerException
    //   at UserService.getUserCity(UserService.java:15)
}

// 실행: getUserCity(999)  ❌ Crash
```

**✅ After (Safe with Optional)**:
```java
// @CODE:OPTIONAL-HANDLING: Optional 패턴
public String getUserCity(int userId) {
    return userRepository.findById(userId)           // Optional<User>
        .map(User::getAddress)                       // Optional<Address>
        .map(Address::getCity)                       // Optional<String>
        .map(String::toUpperCase)                    // Optional<String>
        .orElse("UNKNOWN");                          // String (기본값)
}

// 테스트:
assertEquals("SEOUL", getUserCity(1));              // ✅ PASS
assertEquals("UNKNOWN", getUserCity(999));          // ✅ PASS (안전)
```

### Example 4: Go goroutine Deadlock 디버깅

**❌ Before (Deadlock)**:
```go
// @CODE:CONCURRENCY-DEADLOCK
func processOrders(orderChan chan Order) {
    for order := range orderChan {
        // 문제: goroutine 누수 + 데드락 위험
        go func(o Order) {
            result := processOrder(o)
            // 채널 닫혔는데 쓰기 시도 → panic
            orderChan <- result
        }(order)
    }

    close(orderChan)  // ❌ 데드락: goroutine이 아직 실행 중
}

// 실행: goroutine 300+ 개 생성 후 데드락
```

**✅ After (Safe with WaitGroup)**:
```go
// @CODE:CONCURRENCY-DEADLOCK: WaitGroup으로 동기화
func processOrders(orders []Order) []Result {
    var wg sync.WaitGroup
    resultChan := make(chan Result, len(orders))

    for _, order := range orders {
        wg.Add(1)

        go func(o Order) {
            defer wg.Done()  // 명시적 완료

            result := processOrder(o)
            resultChan <- result
        }(order)
    }

    // 모든 goroutine 완료 대기
    go func() {
        wg.Wait()
        close(resultChan)  // ✅ 모든 작업 완료 후 닫기
    }()

    // 결과 수집
    var results []Result
    for r := range resultChan {
        results = append(results, r)
    }
    return results
}

// 테스트:
result := processOrders(orders)
assert len(result) == len(orders)  // ✅ PASS
```

### Example 5: Rust Borrow Checker Error 디버깅

**❌ Before (Compilation Error)**:
```rust
// @CODE:BORROW-CHECKER
fn update_user_name(mut user: User) {
    let name_ref = &mut user.name;
    *name_ref = "Alice".to_string();

    // 문제: 다중 가변 참조
    let name_ref2 = &mut user.name;  // ❌ error[E0499]
    *name_ref2 = "Bob".to_string();   // 이미 borrowed

    println!("{}", user.name);  // ❌ compile error
}

// Rust Compiler Error:
// error[E0499]: cannot borrow `user.name` as mutable more than once
```

**✅ After (Correct Ownership)**:
```rust
// @CODE:BORROW-CHECKER: 명시적 스코프 관리
fn update_user_name(mut user: User) {
    {
        let name_ref = &mut user.name;
        *name_ref = "Alice".to_string();
    }  // name_ref 드롭됨 - 이제 안전

    // 이제 새로운 가변 참조 가능
    let name_ref2 = &mut user.name;
    *name_ref2 = "Bob".to_string();

    println!("{}", user.name);  // ✅ "Bob" 출력
}

// 테스트:
#[test]
fn test_update_name() {
    let user = User { name: "Alice".to_string() };
    update_user_name(user);
    // ✅ 컴파일 성공 + 런타임 안전
}
```

## Keywords

"에러 해결", "디버깅", "스택 트레이스", "NullPointerException", "TypeError", "N+1 쿼리", "데드락", "메모리 누수", "오류 분석", "근본 원인", "error pattern", "stack trace analysis", "runtime debugging"

## Reference

- Language debugging guides: `.moai/memory/development-guide.md#language-별-디버깅`
- Error handling patterns: CLAUDE.md#예외-처리
- Common error patterns: `.moai/memory/development-guide.md#common-error-patterns`

## Works well with

- moai-essentials-review (코드 품질 검증)
- moai-essentials-refactor (리팩토링)
- moai-essentials-perf (성능 분석)
