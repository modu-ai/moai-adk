---
title: TRUST 5원칙
description: MoAI-ADK의 품질 기준과 준수 방법
---

# TRUST 5원칙

TRUST는 MoAI-ADK가 지향하는 **5가지 핵심 품질 원칙**입니다. SPEC-First TDD 개발에서 코드 품질과 추적성을 보장하는 기준입니다.

> **T**est First + **R**eadable + **U**nified + **S**ecured + **T**rackable = **TRUST**

## 개요

### TRUST 5원칙 요약

| 원칙 | 의미 | 목표 | v0.1.8 달성율 |
|------|------|------|--------------|
| **T** | Test First | SPEC 기반 TDD | 92.9% |
| **R** | Readable | 요구사항 주도 가독성 | 100% |
| **U** | Unified | SPEC 기반 아키텍처 | 95% |
| **S** | Secured | 보안 by 설계 | 100% |
| **T** | Trackable | CODE-FIRST TAG 추적성 | 95% |

**전체 준수율**: 96.6% (목표 85% 대비 113.6% 초과 달성)

### 왜 TRUST인가?

1. **품질 보장**: 객관적이고 측정 가능한 품질 기준
2. **일관성**: 모든 언어와 프레임워크에 적용 가능
3. **자동화**: Claude Code 에이전트가 자동 검증
4. **추적성**: SPEC부터 코드까지 완전한 연결

## T - Test First (테스트 우선)

### 원칙

**"SPEC 없이는 테스트 없고, 테스트 없이는 코드 없다"**

모든 코드는 다음 순서로 작성됩니다:

```
SPEC → Test → Code
```

### SPEC 기반 테스트 작성

#### 1단계: SPEC 요구사항

```markdown
# SPEC-AUTH-001

## Requirements

### Ubiquitous Requirements
- 시스템은 이메일/비밀번호 기반 인증을 제공해야 한다

### Event-driven Requirements
- WHEN 유효한 자격증명으로 로그인하면, JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 401 에러를 반환해야 한다

### Constraints
- 토큰 만료시간은 15분을 초과하지 않아야 한다
- 비밀번호는 bcrypt로 해싱해야 한다
```

#### 2단계: RED - 실패하는 테스트

```typescript
// @TEST:AUTH-001: 인증 서비스 테스트
import { describe, test, expect, beforeEach } from 'vitest';
import { AuthService } from '../src/auth/service';

describe('AuthService - SPEC-AUTH-001', () => {
  let service: AuthService;

  beforeEach(() => {
    service = new AuthService();
  });

  // @SPEC:AUTH-001: 유효한 자격증명 인증
  test('should authenticate valid credentials', async () => {
    const result = await service.authenticate(
      'user@example.com',
      'password123'
    );

    expect(result.success).toBe(true);
    expect(result.token).toBeDefined();
    expect(result.token).toMatch(/^[\w-]+\.[\w-]+\.[\w-]+$/); // JWT 형식
  });

  // @SPEC:AUTH-001: 잘못된 자격증명 거부
  test('should reject invalid credentials', async () => {
    const result = await service.authenticate(
      'user@example.com',
      'wrongpassword'
    );

    expect(result.success).toBe(false);
    expect(result.error).toBe('Invalid credentials');
  });

  // CONSTRAINT:AUTH-001: 토큰 만료시간 검증
  test('should generate token with 15 minutes expiry', async () => {
    const result = await service.authenticate(
      'user@example.com',
      'password123'
    );

    const decoded = decodeToken(result.token);
    const expiryMinutes = (decoded.exp - decoded.iat) / 60;

    expect(expiryMinutes).toBeLessThanOrEqual(15);
  });
});
```

테스트 실행 → **실패 확인**:

```bash
npm test

# 출력:
FAIL  __tests__/auth/service.test.ts
  ✗ should authenticate valid credentials
    TypeError: service.authenticate is not a function
```

#### 3단계: GREEN - 최소 구현

```typescript
// @CODE:AUTH-001: 인증 서비스 구현
export class AuthService {
  async authenticate(email: string, password: string) {
    // 최소한의 구현으로 테스트 통과
    if (email === 'user@example.com' && password === 'password123') {
      return {
        success: true,
        token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
      };
    }

    return {
      success: false,
      error: 'Invalid credentials'
    };
  }
}
```

테스트 실행 → **통과 확인**:

```bash
npm test

# 출력:
PASS  __tests__/auth/service.test.ts
  ✓ should authenticate valid credentials (15ms)
  ✓ should reject invalid credentials (8ms)
  ✓ should generate token with 15 minutes expiry (12ms)

Test Suites: 1 passed, 1 total
Tests:       3 passed, 3 total
```

#### 4단계: REFACTOR - 품질 개선

```typescript
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
// Related: @CODE:AUTH-001:API, @CODE:AUTH-001:DATA

export class AuthService {
  constructor(
    private userRepository: UserRepository,
    private tokenService: TokenService,
    private passwordService: PasswordService
  ) {}

  async authenticate(email: string, password: string): Promise<AuthResult> {
    // @CODE:AUTH-001:API: 입력값 검증
    this.validateInput(email, password);

    // @CODE:AUTH-001:DATA: 사용자 조회
    const user = await this.userRepository.findByEmail(email);
    if (!user) {
      return this.failureResponse();
    }

    // @CODE:AUTH-001: bcrypt 비밀번호 검증
    const isValid = await this.passwordService.verify(
      password,
      user.passwordHash
    );
    if (!isValid) {
      return this.failureResponse();
    }

    // @CODE:AUTH-001: JWT 토큰 발급 (15분 만료)
    const token = await this.tokenService.generate(user, { expiresIn: '15m' });

    return {
      success: true,
      token
    };
  }

  private validateInput(email: string, password: string): void {
    if (!email || !password) {
      throw new ValidationError('Email and password are required');
    }

    if (!/^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$/.test(email)) {
      throw new ValidationError('Invalid email format');
    }
  }

  private failureResponse(): AuthResult {
    return {
      success: false,
      error: 'Invalid credentials'
    };
  }
}
```

### 언어별 테스트 프레임워크

#### TypeScript (Vitest) - 92.9% 성공률

```typescript
import { describe, test, expect } from 'vitest';

describe('UserService', () => {
  test('should create user', async () => {
    const service = new UserService();
    const user = await service.create({ email: 'test@example.com' });
    expect(user.id).toBeDefined();
  });
});
```

**실행**:

```bash
npm test                  # 모든 테스트
npm test -- --watch       # watch 모드
npm test -- --coverage    # 커버리지
```

**v0.1.8 성과**:
- Vitest 테스트 성공률: 92.9% (52/56 통과)
- 커버리지: 92.5%

#### Python (pytest + mypy)

```python
# @TEST:USER-001: 사용자 생성 테스트
import pytest
from src.user_service import UserService

def test_should_create_user():
    """@TEST:USER-001: 유효한 사용자 생성"""
    service = UserService()
    user = service.create(email='test@example.com')
    assert user.id is not None

def test_should_reject_invalid_email():
    """@CODE:USER-001:API: 잘못된 이메일 거부"""
    service = UserService()
    with pytest.raises(ValidationError):
        service.create(email='invalid-email')
```

**실행**:

```bash
pytest                          # 모든 테스트
pytest --cov=src tests/        # 커버리지
pytest -v                       # 상세 출력
mypy src/                      # 타입 검사
```

#### Java (JUnit)

```java
// @TEST:USER-001: 사용자 생성 테스트
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class UserServiceTest {
    Test
    void shouldCreateUser() {
        UserService service = new UserService();
        User user = service.create("test@example.com");
        assertNotNull(user.getId());
    }
}
```

#### Go (go test)

```go
// @TEST:USER-001: 사용자 생성 테스트
package user_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestShouldCreateUser(t *testing.T) {
    service := NewUserService()
    user, err := service.Create("test@example.com")

    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

#### Rust (cargo test)

```rust
// @TEST:USER-001: 사용자 생성 테스트
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn should_create_user() {
        let service = UserService::new();
        let user = service.create("test@example.com").unwrap();
        assert!(user.id.is_some());
    }
}
```

### 커버리지 85% 이상

MoAI-ADK는 최소 85% 테스트 커버리지를 요구합니다:

```bash
npm test -- --coverage

# 출력:
Coverage report
  Lines       : 92.5%
  Statements  : 91.8%
  Branches    : 88.3%
  Functions   : 95.2%

✓ Coverage threshold met: 85%
```

## R - Readable (가독성)

### 원칙

**"요구사항이 코드에 드러나야 한다"**

코드는 SPEC 용어를 사용하고, 의도를 명확히 표현해야 합니다.

### 요구사항 주도 가독성

#### Before (의도 불명확)

```typescript
// ❌ 나쁜 예: 의도를 알 수 없음
function proc(d: any) {
  if (d.t === 1) {
    return d.v * 1.1;
  }
  return d.v;
}
```

#### After (SPEC 기반)

```typescript
// ✅ 좋은 예: SPEC 용어와 의도 명확
// @CODE:PRICING-001: 프리미엄 회원 할인
function calculatePrice(order: Order): number {
  // @SPEC:PRICING-001: 프리미엄 회원은 10% 할인
  if (order.isPremiumMember()) {
    return order.basePrice * 0.9; // 10% discount
  }

  return order.basePrice;
}
```

### 함수 크기 제한

**규칙**:
- 함수: ≤ 50 LOC (Lines of Code)
- 파일: ≤ 300 LOC
- 매개변수: ≤ 5개
- 복잡도: ≤ 10 (Cyclomatic Complexity)

#### 리팩토링 예시

**Before (62 LOC, 복잡도 15)**:

```typescript
// ❌ 너무 긴 함수
function processOrder(order: Order) {
  if (!order.user) throw new Error('No user');
  if (!order.items.length) throw new Error('No items');
  if (order.total < 0) throw new Error('Invalid total');

  let total = 0;
  for (const item of order.items) {
    if (item.quantity <= 0) throw new Error('Invalid quantity');
    total += item.price * item.quantity;
    if (item.discount) {
      total -= item.discount;
    }
  }

  if (order.user.isPremium) {
    total *= 0.9;
  }

  if (order.coupon) {
    if (order.coupon.isValid()) {
      total -= order.coupon.amount;
    }
  }

  // ... 더 많은 로직
}
```

**After (4개 함수, 각 ≤ 15 LOC)**:

```typescript
// ✅ 작고 명확한 함수들
function processOrder(order: Order): OrderResult {
  validateOrder(order);
  const subtotal = calculateSubtotal(order.items);
  const discount = calculateDiscount(order, subtotal);
  const total = subtotal - discount;

  return { total, discount };
}

function validateOrder(order: Order): void {
  if (!order.user) throw new ValidationError('User required');
  if (!order.items.length) throw new ValidationError('Items required');
}

function calculateSubtotal(items: OrderItem[]): number {
  return items.reduce((sum, item) => {
    validateItem(item);
    return sum + item.price * item.quantity - (item.discount || 0);
  }, 0);
}

function calculateDiscount(order: Order, subtotal: number): number {
  let discount = 0;

  // @SPEC:PRICING-001: 프리미엄 회원 10% 할인
  if (order.user.isPremium) {
    discount += subtotal * 0.1;
  }

  // @SPEC:PRICING-002: 쿠폰 할인 적용
  if (order.coupon?.isValid()) {
    discount += order.coupon.amount;
  }

  return discount;
}
```

### SPEC 용어 사용

코드의 모든 이름은 SPEC에서 사용된 용어를 따릅니다:

```markdown
# SPEC-PAYMENT-001

## Terms
- **Payment Method**: 결제 수단 (카드, 계좌이체 등)
- **Transaction**: 결제 거래
- **Settlement**: 정산
```

```typescript
// ✅ SPEC 용어 사용
class PaymentService {
  async processTransaction(paymentMethod: PaymentMethod): Promise<Transaction> {
    const transaction = await this.createTransaction(paymentMethod);
    await this.executeSettlement(transaction);
    return transaction;
  }
}

// ❌ SPEC 용어 미사용
class PayService {
  async proc(pm: any): Promise<any> {
    const tx = await this.create(pm);
    await this.exec(tx);
    return tx;
  }
}
```

## U - Unified (통합)

### 원칙

**"SPEC이 아키텍처를 결정한다"**

언어나 프레임워크가 아닌, SPEC 요구사항이 시스템 구조를 정의합니다.

### SPEC 기반 복잡도 관리

각 SPEC은 복잡도 임계값을 정의합니다:

```markdown
# SPEC-AUTH-001

## Complexity Constraints
- 순환 복잡도: ≤ 10
- 의존성 깊이: ≤ 3
- 함수 크기: ≤ 50 LOC
```

초과 시:
1. 새로운 SPEC 작성
2. Waiver 문서 작성 (예외 승인)

### 언어 간 일관성

동일한 SPEC을 여러 언어로 구현할 때도 일관된 구조:

#### TypeScript 구현

```typescript
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
export interface AuthService {
  authenticate(email: string, password: string): Promise<AuthResult>;
  validateToken(token: string): Promise<boolean>;
}
```

#### Python 구현

```python
# @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
from abc import ABC, abstractmethod

class AuthService(ABC):
    @abstractmethod
    async def authenticate(self, email: str, password: str) -> AuthResult:
        pass

    @abstractmethod
    async def validate_token(self, token: str) -> bool:
        pass
```

#### Go 구현

```go
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
type AuthService interface {
    Authenticate(email, password string) (*AuthResult, error)
    ValidateToken(token string) (bool, error)
}
```

#### Java 구현

```java
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
public interface AuthService {
    CompletableFuture<AuthResult> authenticate(String email, String password);
    CompletableFuture<Boolean> validateToken(String token);
}
```

#### Rust 구현

```rust
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
pub trait AuthService {
    async fn authenticate(&self, email: &str, password: &str) -> Result<AuthResult>;
    async fn validate_token(&self, token: &str) -> Result<bool>;
}
```

**동일한 SPEC, 동일한 구조, 언어별 관례 준수**

## S - Secured (보안)

### 원칙

**"모든 SPEC에 보안 요구사항을 명시한다"**

보안은 완료 후 추가가 아니라, TDD 단계에서 구현합니다.

### SPEC 보안 요구사항

```markdown
# SPEC-AUTH-001

## Security Requirements

### Authentication
- 비밀번호는 bcrypt로 해싱해야 한다 (cost factor: 12)
- 토큰은 HS256 알고리즘으로 서명해야 한다
- 비밀 키는 환경 변수로 관리해야 한다

### Input Validation
- 이메일 형식 검증 (정규식)
- 비밀번호 최소 8자, 특수문자 포함
- SQL Injection 방지

### Error Handling
- 인증 실패 시 구체적인 실패 이유 노출 금지
- 일반적인 "Invalid credentials" 메시지 사용
```

### Winston Logger (v0.1.8)

**민감정보 자동 마스킹**:

```typescript
import { logger } from './utils/winston-logger';

// ✅ 자동 마스킹
logger.info('User login', {
  email: 'user@example.com',
  password: 'secret123',        // → ***REDACTED***
  token: 'eyJhbGci...',         // → ***REDACTED***
  apiKey: 'sk_live_abc123'      // → ***REDACTED***
});

// @TAG 통합 로깅
logger.logWithTag('AUTH-001', 'Authentication successful', {
  userId: '12345',
  action: 'login'
});
```

**마스킹 대상 (27개)**:

- **필드**: password, token, apiKey, secret, accessKey, privateKey, credentials 등 15개
- **패턴**: email, credit card, SSN, phone, IP, URL credentials 등 12개

### 입력 검증

```typescript
// @CODE:AUTH-001:API: 입력값 보안 검증
class AuthService {
  private validateEmail(email: string): void {
    // 이메일 형식 검증
    const emailRegex = /^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$/;
    if (!emailRegex.test(email)) {
      throw new ValidationError('Invalid email format');
    }

    // 이메일 길이 제한
    if (email.length > 255) {
      throw new ValidationError('Email too long');
    }
  }

  private validatePassword(password: string): void {
    // 비밀번호 정책
    if (password.length < 8) {
      throw new ValidationError('Password must be at least 8 characters');
    }

    if (!/[A-Z]/.test(password)) {
      throw new ValidationError('Password must contain uppercase letter');
    }

    if (!/[0-9]/.test(password)) {
      throw new ValidationError('Password must contain number');
    }

    if (!/[!@#$%^&*]/.test(password)) {
      throw new ValidationError('Password must contain special character');
    }
  }
}
```

## T - Trackable (추적성)

### 원칙

**"모든 코드는 SPEC으로 추적 가능해야 한다"**

CODE-FIRST TAG 시스템으로 요구사항부터 코드까지 완전한 연결을 유지합니다.

### CODE-FIRST TAG 시스템

**핵심 철학**: TAG의 진실은 코드 자체에만 존재
- 중간 캐시 없음 (indexes 디렉토리 제거)
- 코드 직접 스캔 (rg '@TAG' -n)
- 실시간 검증

### @TAG 체계

**TAG 체인 (4 Core)**: 요구사항부터 검증까지
```
@SPEC → @TEST → @CODE → @DOC
```

**@CODE 서브카테고리**: 구현 세부 사항
```
@CODE → @CODE → @CODE → @CODE
```

### SPEC-코드 추적성

#### TAG BLOCK 예시

```typescript
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
// Related: @CODE:AUTH-001:API, @CODE:AUTH-001:DATA

/**
 * @CODE:AUTH-001:API: 사용자 인증 서비스
 */
export class AuthService {
  /**
   * @CODE:AUTH-001: 이메일/비밀번호 인증
   * @CODE:AUTH-001:DATA: 사용자 데이터 조회 및 검증
   */
  async authenticate(email: string, password: string): Promise<AuthResult> {
    // 구현...
  }
}

// @TEST:AUTH-001: 인증 테스트
describe('AuthService', () => {
  test('@TEST:AUTH-001: should authenticate valid user', () => {
    // 테스트...
  });
});
```

### 3단계 워크플로우 추적

```
/moai:1-spec  → @SPEC, @SPEC, @CODE 생성
             → SPEC 문서에 TAG BLOCK

/moai:2-build → @TEST, @CODE 서브카테고리 (API, UI, DATA 등) 적용
             → 코드에 TAG BLOCK 삽입

/moai:3-sync  → 코드 직접 스캔 (rg '@TAG' -n)
             → TAG 체인 검증
             → sync-report.md 생성
```

### 코드 스캔 기반 검증

```bash
# TAG 검색
rg "@CODE:AUTH-001" -n src/
rg "@TEST:AUTH-001" -n tests/

# TAG 체인 검증
rg "@SPEC:AUTH-001||@CODE:AUTH-001|@TEST:AUTH-001" -n .

# 고아 TAG 감지
rg "@DOC:[A-Z]+-\d+" -n . | @agent-tag-agent "고아 TAG 감지"
```

### TAG 검증 리포트

```markdown
# TAG 검증 리포트 (2025-09-30)

## 통계
- 총 TAG: 149개
- 완결 체인: 32개
- 불완전 체인: 2개
- 고아 TAG: 0개

## 완결된 체인
✅ AUTH-001: REQ → DESIGN → TASK → TEST
✅ PAYMENT-002: REQ → DESIGN → TASK → TEST
✅ PROFILE-003: REQ → DESIGN → TASK → TEST

## 불완전한 체인
⚠️ NOTIFICATION-004: @TEST 누락
⚠️ REPORT-005: @SPEC 누락

## 권장 사항
- NOTIFICATION-004: 테스트 추가 필요
- REPORT-005: 설계 문서 작성 필요
```

## TRUST 준수율 측정

### 자동 검증 도구

```bash
# TRUST 검증 실행
@agent-trust-checker "TRUST 원칙 검증"

# 또는 CLI
moai status --trust
```

### 리포트 생성

```markdown
# TRUST 준수율 리포트 (v0.1.8)

## 전체 준수율: 96.6%

### T - Test First: 92.9%
✓ 테스트 성공률: 92.9% (Vitest 52/56)
✓ 테스트 커버리지: 92.5%
✓ TDD 사이클 준수: 90%

### R - Readable: 100%
✓ 함수 크기: 100% (모두 ≤50 LOC)
✓ 파일 크기: 100% (모두 ≤300 LOC)
✓ 복잡도: 98% (≤10)

### U - Unified: 95%
✓ SPEC 기반 설계: 98%
✓ 언어 간 일관성: 92%

### S - Secured: 100%
✓ 입력 검증: 100%
✓ Winston logger: 97.92% coverage
✓ 민감정보 마스킹: 100%

### T - Trackable: 95%
✓ CODE-FIRST TAG 시스템: 100%
✓ TAG 체인 완결: 94%
✓ SPEC-코드 연결: 91%
```

### 개선 가이드

준수율이 낮은 항목에 대한 자동 제안:

```
⚠️ Vitest 테스트 실패: 4건

권장 사항:
1. 실패 원인 분석 및 수정
2. 테스트 격리 확인
3. 비동기 처리 검토

관련 자료:
- /concepts/spec-first-tdd
- /concepts/workflow
```

## 다음 단계

### TRUST 실천

1. **[SPEC-First TDD](/concepts/spec-first-tdd)**: 방법론 학습
2. **[TAG 시스템](/concepts/tag-system)**: 추적성 관리
3. **[워크플로우](/concepts/workflow)**: 3단계 사이클 이해

### 품질 검증

1. **[trust-checker 에이전트](/claude/agents)**: 자동 검증
2. **[moai status](/cli/status)**: TRUST 준수율 확인
3. **[CI/CD 통합](/advanced/ci-cd)**: 자동화 파이프라인

## 참고 자료

- **개발 가이드**: `.moai/memory/development-guide.md`
- **TRUST 검증**: `moai status --trust`
- **품질 게이트**: `.moai/config.json` 설정