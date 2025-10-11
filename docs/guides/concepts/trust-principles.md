# TRUST 5원칙 가이드

<!-- @CODE:DOCS-002 | SPEC: .moai/specs/SPEC-DOCS-002/spec.md -->

> "AI 시대의 일관된 코드 품질 보장"

## 개요

**TRUST 5원칙**은 MoAI-ADK가 모든 코드에 적용하는 품질 기준입니다. Test First, Readable, Unified, Secured, Trackable의 앞글자를 따서 만든 이 원칙은 AI와 협업하는 개발 환경에서도 높은 코드 품질을 유지하도록 설계되었습니다.

### TRUST의 의의

- **일관성**: 모든 주요 프로그래밍 언어에 적용 가능한 통합 기준
- **검증 가능성**: 자동화 도구로 검증 가능한 명확한 지표
- **AI 친화적**: AI 에이전트가 이해하고 적용할 수 있는 구조화된 규칙
- **추적성**: @TAG 시스템과 결합하여 완벽한 코드 추적성 확보

---

## T - Test First (테스트 주도 개발)

### SPEC → Test → Code 사이클

```
@SPEC:ID (요구사항) → @TEST:ID (테스트) → @CODE:ID (구현) → @DOC:ID (문서)
```

MoAI-ADK의 TDD는 SPEC에서 시작하여 코드로 완성되는 완전한 사이클을 따릅니다.

### 언어별 테스트 프레임워크

| 언어 | 테스트 도구 | 커버리지 도구 | 목표 |
|------|------------|--------------|------|
| **Python** | pytest | pytest-cov | ≥85% |
| **TypeScript** | Vitest, Jest | c8, istanbul | ≥85% |
| **Java** | JUnit 5 | JaCoCo | ≥85% |
| **Go** | go test | built-in | ≥85% |
| **Rust** | cargo test | tarpaulin | ≥85% |

### TDD 3단계: RED-GREEN-REFACTOR

#### 1. 🔴 RED: 실패하는 테스트 작성

```python
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
import pytest
from auth.service import AuthService

def test_should_authenticate_valid_user():
    """TEST-LOGIN-001: 유효한 사용자 인증 테스트"""
    # Arrange
    auth = AuthService()
    username = "test@example.com"
    password = "validPassword123"

    # Act
    result = auth.authenticate(username, password)

    # Assert
    assert result.success is True
    assert result.token is not None
    assert result.token_type == "Bearer"
```

#### 2. 🟢 GREEN: 최소 구현

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.py
from dataclasses import dataclass

@dataclass
class AuthResult:
    success: bool
    token: str | None
    token_type: str

class AuthService:
    """CODE-LOGIN-001: 사용자 인증 서비스"""

    def authenticate(self, username: str, password: str) -> AuthResult:
        # 최소 구현: 테스트 통과만 목표
        return AuthResult(
            success=True,
            token="dummy_token",
            token_type="Bearer"
        )
```

#### 3. 🔄 REFACTOR: 품질 개선

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.py
import bcrypt
import jwt
from datetime import datetime, timedelta

class AuthService:
    """CODE-LOGIN-001: 사용자 인증 서비스 (리팩토링 완료)"""

    def __init__(self, user_repo, secret_key: str):
        self._user_repo = user_repo
        self._secret_key = secret_key

    def authenticate(self, username: str, password: str) -> AuthResult:
        # 실제 구현
        user = self._user_repo.find_by_username(username)
        if not user:
            return AuthResult(success=False, token=None, token_type="")

        if not bcrypt.checkpw(password.encode(), user.password_hash):
            return AuthResult(success=False, token=None, token_type="")

        token = self._generate_jwt(user.id)
        return AuthResult(success=True, token=token, token_type="Bearer")

    def _generate_jwt(self, user_id: int) -> str:
        payload = {
            "sub": user_id,
            "exp": datetime.utcnow() + timedelta(minutes=15)
        }
        return jwt.encode(payload, self._secret_key, algorithm="HS256")
```

---

## R - Readable (가독성)

### 코드 제약 기준

- **파일 크기**: ≤ 300 LOC (Lines of Code)
- **함수 크기**: ≤ 50 LOC
- **매개변수**: ≤ 5개
- **복잡도**: ≤ 10 (Cyclomatic Complexity)

### 의도를 드러내는 이름

❌ **나쁜 예**:
```typescript
function calc(a: number, b: number): number {
  return a * b * 0.1;
}
```

✅ **좋은 예**:
```typescript
function calculateDiscountedPrice(
  originalPrice: number,
  discountRate: number
): number {
  return originalPrice * discountRate;
}
```

### 가드절 우선 사용

❌ **나쁜 예** (중첩 조건):
```typescript
function processPayment(amount: number, user: User) {
  if (user.isActive) {
    if (amount > 0) {
      if (user.balance >= amount) {
        // 결제 처리
        return processTransaction(amount, user);
      } else {
        throw new Error("잔액 부족");
      }
    } else {
      throw new Error("유효하지 않은 금액");
    }
  } else {
    throw new Error("비활성 사용자");
  }
}
```

✅ **좋은 예** (가드절):
```typescript
function processPayment(amount: number, user: User) {
  // 가드절로 조기 리턴
  if (!user.isActive) {
    throw new Error("비활성 사용자");
  }

  if (amount <= 0) {
    throw new Error("유효하지 않은 금액");
  }

  if (user.balance < amount) {
    throw new Error("잔액 부족");
  }

  // 핵심 로직
  return processTransaction(amount, user);
}
```

### 언어별 린터/포매터

| 언어 | 린터 | 포매터 | 특징 |
|------|------|--------|------|
| **Python** | ruff, pylint | black, ruff | 빠른 속도, 엄격한 규칙 |
| **TypeScript** | Biome, ESLint | Biome, Prettier | 통합 도구, 설정 간편 |
| **Go** | golint, staticcheck | gofmt, goimports | 표준 도구, 일관성 강제 |
| **Rust** | clippy | rustfmt | 강력한 정적 분석 |
| **Java** | Checkstyle, PMD | google-java-format | 엔터프라이즈 표준 |

---

## U - Unified (통합 아키텍처)

### SPEC 기반 복잡도 관리

각 SPEC은 복잡도 임계값을 정의합니다. 초과 시 새로운 SPEC 또는 명확한 근거가 있는 면제(Waiver)가 필요합니다.

**복잡도 임계값 예시**:
```markdown
### Constraints (제약사항)
- 단일 모듈의 클래스 개수는 5개를 초과하지 않아야 한다
- 함수의 Cyclomatic Complexity는 10을 초과하지 않아야 한다
- IF 복잡도 초과가 불가피하면, ADR(Architecture Decision Record)로 근거를 문서화해야 한다
```

### 모듈화 및 경계 정의

언어별 모듈 경계는 SPEC이 정의합니다:

| 언어 | 모듈 단위 | 예시 |
|------|----------|------|
| **Python** | 패키지 (\_\_init\_\_.py) | `auth/`, `payment/` |
| **TypeScript** | 인터페이스 + 배럴 | `index.ts` 내보내기 |
| **Java** | 패키지 | `com.moai.auth` |
| **Go** | 패키지 | `package auth` |
| **Rust** | 크레이트 + 모듈 | `mod auth;` |

### 일관된 패턴 사용

**의존성 주입 (Python)**:
```python
# @CODE:AUTH-001:DOMAIN
class AuthService:
    def __init__(self, user_repo: UserRepository):
        self._user_repo = user_repo  # 의존성 주입
```

**의존성 주입 (TypeScript)**:
```typescript
// @CODE:AUTH-001:DOMAIN
class AuthService {
  constructor(private readonly userRepo: UserRepository) {}
}
```

---

## S - Secured (보안)

### SPEC 보안 요구사항 정의

모든 SPEC에 보안 요구사항을 명시적으로 정의합니다.

**예시**:
```markdown
### Security Requirements
- 시스템은 모든 비밀번호를 bcrypt(cost factor 12)로 해싱해야 한다
- 시스템은 JWT 토큰 만료시간을 15분으로 제한해야 한다
- IF SQL 쿼리에 사용자 입력이 포함되면, 시스템은 파라미터화된 쿼리를 사용해야 한다
```

### 언어별 보안 도구

| 언어 | 보안 도구 | 주요 검사 항목 |
|------|----------|---------------|
| **Python** | bandit, safety | 취약한 함수, 의존성 취약점 |
| **TypeScript** | npm audit, Snyk | 패키지 취약점, XSS |
| **Java** | OWASP Dependency-Check | 의존성 취약점, 인젝션 |
| **Go** | gosec | 취약한 암호화, 랜덤 |
| **Rust** | cargo audit | 크레이트 취약점 |

### 보안 by 설계

보안 제어는 완료 후 추가하는 것이 아니라 **TDD 단계에서 구현**합니다.

**예시: 입력 검증 (RED 단계)**:
```python
# @TEST:AUTH-001
def test_should_reject_sql_injection_attempt():
    auth = AuthService()
    malicious_input = "admin' OR '1'='1"

    with pytest.raises(ValidationError):
        auth.authenticate(malicious_input, "password")
```

---

## T - Trackable (추적성)

### @TAG 시스템을 통한 완벽한 추적성

```
@SPEC:AUTH-001 (SPEC 문서)
    ↓
@TEST:AUTH-001 (테스트 코드)
    ↓
@CODE:AUTH-001 (구현 코드)
    ↓
@DOC:AUTH-001 (문서)
```

### CODE-FIRST 원칙

TAG의 진실은 **코드 자체**에만 존재합니다. 중간 캐시나 데이터베이스 없이 직접 스캔합니다.

**TAG 검증 명령어**:
```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 특정 도메인 TAG 조회
rg "@SPEC:AUTH" -n .moai/specs/

# 고아 TAG 탐지
rg '@CODE:AUTH-001' -n src/          # CODE는 있는데
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC이 없으면 고아
```

### TAG 서브 카테고리

구현 세부사항은 `@CODE:ID` 내부에 주석으로 표기:

```python
# @CODE:AUTH-001:API - REST API 엔드포인트
@app.post("/auth/login")
def login(credentials: LoginRequest):
    pass

# @CODE:AUTH-001:DOMAIN - 비즈니스 로직
class AuthService:
    pass

# @CODE:AUTH-001:DATA - 데이터 모델
class User(BaseModel):
    pass
```

---

## TRUST 체크리스트

### 구현 전 필수 검증

```markdown
### ✅ T - Test First
- [ ] SPEC 문서 작성 완료
- [ ] RED: 실패하는 테스트 작성 및 확인
- [ ] GREEN: 테스트 통과하는 최소 구현
- [ ] REFACTOR: 코드 품질 개선
- [ ] 테스트 커버리지 ≥85%

### ✅ R - Readable
- [ ] 파일 크기 ≤300 LOC
- [ ] 함수 크기 ≤50 LOC
- [ ] 매개변수 ≤5개
- [ ] 복잡도 ≤10
- [ ] 린터/포매터 통과

### ✅ U - Unified
- [ ] SPEC 기반 모듈 경계 정의
- [ ] 일관된 아키텍처 패턴 적용
- [ ] 의존성 방향 명확

### ✅ S - Secured
- [ ] SPEC에 보안 요구사항 명시
- [ ] 입력 검증 구현
- [ ] 보안 도구 스캔 통과
- [ ] 민감 데이터 암호화

### ✅ T - Trackable
- [ ] @TAG 시스템 적용
- [ ] TAG 체인 무결성 검증
- [ ] 고아 TAG 없음 확인
```

---

## 실제 코드 리뷰 시나리오

### 시나리오 1: Python 인증 서비스

**코드 리뷰 체크포인트**:

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.py

# ✅ T - Test First: 테스트 커버리지 90%
# ✅ R - Readable: 함수 크기 35 LOC, 복잡도 7
# ✅ U - Unified: 의존성 주입 패턴 사용
# ✅ S - Secured: bcrypt 해싱, JWT 만료 15분
# ✅ T - Trackable: @TAG 체인 완성

from typing import Optional
import bcrypt
import jwt
from datetime import datetime, timedelta

class AuthService:
    """사용자 인증 서비스"""

    def __init__(self, user_repo: UserRepository, secret_key: str):
        self._user_repo = user_repo
        self._secret_key = secret_key

    def authenticate(self, username: str, password: str) -> AuthResult:
        """사용자 인증"""
        # 가드절: 조기 리턴
        user = self._user_repo.find_by_username(username)
        if not user:
            return self._failed_auth()

        if not self._verify_password(password, user.password_hash):
            return self._failed_auth()

        token = self._generate_jwt(user.id)
        return AuthResult(success=True, token=token, token_type="Bearer")

    def _verify_password(self, plain: str, hashed: bytes) -> bool:
        """비밀번호 검증"""
        return bcrypt.checkpw(plain.encode(), hashed)

    def _generate_jwt(self, user_id: int) -> str:
        """JWT 토큰 생성"""
        payload = {
            "sub": user_id,
            "exp": datetime.utcnow() + timedelta(minutes=15)
        }
        return jwt.encode(payload, self._secret_key, algorithm="HS256")

    def _failed_auth(self) -> AuthResult:
        """인증 실패 응답"""
        return AuthResult(success=False, token=None, token_type="")
```

### 시나리오 2: TypeScript 파일 업로드

**코드 리뷰 체크포인트**:

```typescript
// @CODE:UPLOAD-001 | SPEC: SPEC-UPLOAD-001.md | TEST: tests/upload/service.test.ts

// ✅ T - Test First: 테스트 커버리지 88%
// ✅ R - Readable: 함수 크기 평균 20 LOC
// ✅ U - Unified: 인터페이스 기반 설계
// ✅ S - Secured: 파일 타입 검증, 크기 제한
// ✅ T - Trackable: @TAG 체인 완성

import { z } from 'zod';

// 파일 타입 검증 스키마
const FileSchema = z.object({
  name: z.string().min(1).max(255),
  size: z.number().min(1).max(10 * 1024 * 1024), // 10MB
  type: z.enum(['image/jpeg', 'image/png', 'application/pdf']),
});

export class UploadService {
  constructor(private readonly storage: StorageProvider) {}

  async upload(file: File): Promise<UploadResult> {
    // 가드절: 파일 검증
    const validation = FileSchema.safeParse(file);
    if (!validation.success) {
      return this.failedUpload('유효하지 않은 파일');
    }

    // 파일 업로드
    const url = await this.storage.save(file);
    return { success: true, url };
  }

  private failedUpload(reason: string): UploadResult {
    return { success: false, url: null, error: reason };
  }
}
```

---

## 자동화 도구 통합 가이드

### CI/CD 파이프라인 예시 (GitHub Actions)

```yaml
# .github/workflows/trust-check.yml
name: TRUST 5원칙 검증

on: [push, pull_request]

jobs:
  trust-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      # T - Test First
      - name: Run Tests
        run: |
          pytest --cov=src --cov-report=term-missing
          # 커버리지 85% 이상 강제
          coverage report --fail-under=85

      # R - Readable
      - name: Lint Check
        run: |
          ruff check src/
          black --check src/

      # S - Secured
      - name: Security Scan
        run: |
          bandit -r src/
          safety check

      # T - Trackable
      - name: TAG Validation
        run: |
          # 고아 TAG 탐지
          ./scripts/check-orphan-tags.sh
```

---

## 관련 문서

- [EARS 요구사항 작성 가이드](./ears-guide.md)
- [TAG 시스템 가이드](./tag-system.md)
- [SPEC-First TDD 워크플로우](./spec-first-tdd.md)
- [개발 가이드](../../.moai/memory/development-guide.md)

---

**작성일**: 2025-10-11
**버전**: v1.0.0
**TAG**: @CODE:DOCS-002
