---
title: TRUST 5원칙 완전 가이드
description: MoAI-ADK의 코드 품질 보증을 위한 TRUST 5원칙을 상세히 다룹니다
version: 0.3.0
updated: 2025-10-14
---

# TRUST 5원칙 완전 가이드

> **"완벽한 코드 품질은 TRUST에서 시작된다"**

TRUST 5원칙은 MoAI-ADK가 모든 코드에 적용하는 품질 보증 기준입니다.

---

## 목차

1. [TRUST 5원칙 개요](#trust-5원칙-개요)
2. [T - Test First (테스트 주도 개발)](#t---test-first-테스트-주도-개발)
3. [R - Readable (가독성)](#r---readable-가독성)
4. [U - Unified (통합 아키텍처)](#u---unified-통합-아키텍처)
5. [S - Secured (보안)](#s---secured-보안)
6. [T - Trackable (추적성)](#t---trackable-추적성)
7. [언어별 TRUST 구현](#언어별-trust-구현)
8. [TRUST 검증 도구](#trust-검증-도구)
9. [실전 예시](#실전-예시)
10. [트러블슈팅](#트러블슈팅)
11. [다음 단계](#다음-단계)

---

## TRUST 5원칙 개요

### TRUST란?

**TRUST**는 MoAI-ADK가 정의한 5가지 코드 품질 원칙의 약자입니다:

```
T - Test First (테스트 주도 개발)
R - Readable (가독성)
U - Unified (통합 아키텍처)
S - Secured (보안)
T - Trackable (추적성)
```

### TRUST의 목표

1. **85% 이상 테스트 커버리지** 보장
2. **린터/타입 체크 통과** 의무화
3. **복잡도 ≤ 10, 함수 ≤ 50 LOC** 제약
4. **보안 취약점 제로** 목표
5. **완전한 추적성** (@TAG 시스템)

### TRUST 검증 타이밍

| 타이밍 | 검증 항목 | 도구 |
|--------|---------|------|
| **개발 중** | 린터, 타입 체크 | ruff, mypy (Python) |
| **커밋 전** | 테스트, 커버리지 | pytest --cov |
| **PR 전** | TRUST 전체 검증 | @agent-trust-checker |
| **CI/CD** | 자동 검증 | GitHub Actions |

---

## T - Test First (테스트 주도 개발)

### 원칙

**SPEC → Test → Code 사이클**

모든 코드는 테스트로 시작하고, SPEC 요구사항을 검증합니다.

### SPEC 기반 TDD

```mermaid
graph LR
    A[SPEC 작성] --> B[TEST 작성<br>RED]
    B --> C[CODE 구현<br>GREEN]
    C --> D[REFACTOR]
    D --> E[품질 검증]

    A -->|@SPEC:ID| A
    B -->|@TEST:ID| B
    C -->|@CODE:ID| C
    E -->|TRUST Check| E
```

### 언어별 TDD 구현

#### Python: pytest + mypy

**테스트 프레임워크**: pytest
**타입 체크**: mypy
**커버리지**: pytest-cov

**예시**:

```python
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
import pytest
from src.auth.jwt_service import JWTService, AuthenticationError


class TestJWTService:
    """@SPEC:AUTH-001 JWT 인증 시스템 테스트"""

    def test_generate_token_creates_valid_jwt(self):
        """
        Ubiquitous: 시스템은 JWT 토큰 생성 기능을 제공해야 한다

        Given: JWTService 인스턴스
        When: generate_token 호출
        Then: 유효한 JWT 토큰 반환
        """
        service = JWTService()
        token = service.generate_token(user_id=1)

        assert token is not None
        assert isinstance(token, str)
        assert len(token.split(".")) == 3  # JWT 형식

    def test_invalid_token_raises_error(self):
        """
        Constraint: IF 잘못된 토큰이 제공되면, 접근을 거부해야 한다

        Given: JWTService 인스턴스
        When: 잘못된 토큰으로 verify_token 호출
        Then: AuthenticationError 발생
        """
        service = JWTService()

        with pytest.raises(AuthenticationError):
            service.verify_token("invalid.token.here")

    def test_token_expiry_within_15_minutes(self):
        """
        Constraint: 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다

        Given: JWTService 인스턴스
        When: 토큰 생성 및 디코딩
        Then: exp - iat <= 900초 (15분)
        """
        service = JWTService()
        token = service.generate_token(user_id=1)
        payload = service.decode_token(token)

        expiry_time = payload["exp"] - payload["iat"]
        assert expiry_time <= 900  # 15분 = 900초
```

**실행**:

```bash
# 테스트 실행 + 커버리지
pytest tests/ --cov=src --cov-report=term-missing --cov-fail-under=85

# 타입 체크
mypy src/ --strict
```

#### TypeScript: Vitest + strict typing

**테스트 프레임워크**: Vitest
**타입 체크**: TypeScript (strict)
**커버리지**: vitest --coverage

**예시**:

```typescript
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
import { describe, it, expect, beforeEach } from "vitest";
import { JWTService, AuthenticationError } from "@/auth/jwt-service";

describe("JWTService", () => {
  let service: JWTService;

  beforeEach(() => {
    service = new JWTService();
  });

  it("should generate valid JWT token", () => {
    /**
     * Ubiquitous: 시스템은 JWT 토큰 생성 기능을 제공해야 한다
     *
     * Given: JWTService 인스턴스
     * When: generateToken 호출
     * Then: 유효한 JWT 토큰 반환
     */
    const token = service.generateToken(1);

    expect(token).toBeDefined();
    expect(typeof token).toBe("string");
    expect(token.split(".").length).toBe(3); // JWT 형식
  });

  it("should reject invalid token", () => {
    /**
     * Constraint: IF 잘못된 토큰이 제공되면, 접근을 거부해야 한다
     *
     * Given: JWTService 인스턴스
     * When: 잘못된 토큰으로 verifyToken 호출
     * Then: AuthenticationError 발생
     */
    expect(() => {
      service.verifyToken("invalid.token.here");
    }).toThrow(AuthenticationError);
  });
});
```

**실행**:

```bash
# 테스트 실행 + 커버리지
vitest run --coverage

# 타입 체크
tsc --noEmit
```

#### Go: go test + 테이블 주도 테스트

**테스트 프레임워크**: go test (표준 라이브러리)
**커버리지**: go test -cover

**예시**:

```go
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
package auth

import "testing"

func TestGenerateToken(t *testing.T) {
    /*
     * Ubiquitous: 시스템은 JWT 토큰 생성 기능을 제공해야 한다
     *
     * Given: JWTService 인스턴스
     * When: GenerateToken 호출
     * Then: 유효한 JWT 토큰 반환
     */
    service := NewJWTService()
    token, err := service.GenerateToken(1)

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    if token == "" {
        t.Error("Expected non-empty token")
    }
}

func TestVerifyToken_InvalidToken(t *testing.T) {
    /*
     * Constraint: IF 잘못된 토큰이 제공되면, 접근을 거부해야 한다
     */
    tests := []struct {
        name  string
        token string
        want  error
    }{
        {"empty token", "", ErrInvalidToken},
        {"malformed token", "invalid", ErrInvalidToken},
        {"expired token", "expired.token.here", ErrExpiredToken},
    }

    service := NewJWTService()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := service.VerifyToken(tt.token)
            if err != tt.want {
                t.Errorf("VerifyToken() error = %v, want %v", err, tt.want)
            }
        })
    }
}
```

**실행**:

```bash
# 테스트 실행 + 커버리지
go test ./... -cover -coverprofile=coverage.out

# 커버리지 리포트
go tool cover -html=coverage.out
```

### 테스트 커버리지 목표

| 항목 | 목표 | 최소 |
|------|------|------|
| **전체 커버리지** | 90% | 85% |
| **브랜치 커버리지** | 85% | 80% |
| **함수 커버리지** | 95% | 90% |

---

## R - Readable (가독성)

### 원칙

**SPEC 정렬 클린 코드**

함수는 SPEC 요구사항을 직접 구현하고, 변수명은 SPEC 용어를 반영합니다.

### 코드 제약

| 항목 | 제약 | 이유 |
|------|------|------|
| **파일** | ≤ 300 LOC | 전체 파일을 한눈에 파악 가능 |
| **함수** | ≤ 50 LOC | 단일 책임 원칙 |
| **매개변수** | ≤ 5개 | 인지 부하 감소 |
| **복잡도** | ≤ 10 | 순환 복잡도 (Cyclomatic Complexity) |

### 언어별 SPEC 구현

#### Python: 타입 힌트 + Docstring

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/test_jwt_service.py
from typing import Dict, Any
from datetime import datetime, timedelta, timezone
from dataclasses import dataclass
import jwt


@dataclass
class TokenConfig:
    """
    토큰 설정

    @CODE:AUTH-001:DATA - 데이터 모델
    """
    secret_key: str = "secret"
    algorithm: str = "HS256"
    expiry_minutes: int = 15


class JWTService:
    """
    JWT 인증 서비스

    @SPEC:AUTH-001 요구사항:
    - Ubiquitous: JWT 토큰 생성 기능 제공
    - Event-driven: 유효한 자격증명으로 로그인 시 JWT 토큰 발급
    - Constraint: 액세스 토큰 만료시간 15분 이내

    TDD 이력:
    - v0.0.1 (2025-10-14): RED - 테스트 작성
    - v0.1.0 (2025-10-14): GREEN - 최소 구현
    - v0.1.1 (2025-10-14): REFACTOR - 코드 품질 개선
    """

    def __init__(self, config: TokenConfig | None = None):
        self.config = config or TokenConfig()

    def generate_token(self, user_id: int) -> str:
        """
        JWT 토큰 생성

        @SPEC:AUTH-001 Ubiquitous:
        시스템은 JWT 토큰 생성 기능을 제공해야 한다

        Args:
            user_id: 사용자 ID (양수)

        Returns:
            JWT 토큰 문자열

        Raises:
            ValueError: user_id가 0 이하일 때

        Example:
            >>> service = JWTService()
            >>> token = service.generate_token(1)
            >>> isinstance(token, str)
            True
        """
        # 가드절: 유효하지 않은 입력 처리
        if user_id <= 0:
            raise ValueError("user_id must be positive")

        # @CODE:AUTH-001:DOMAIN - 비즈니스 로직
        payload = self._create_payload(user_id)
        return jwt.encode(payload, self.config.secret_key, algorithm=self.config.algorithm)

    def _create_payload(self, user_id: int) -> Dict[str, Any]:
        """
        페이로드 생성 (내부 헬퍼)

        @SPEC:AUTH-001 Constraint:
        액세스 토큰 만료시간은 15분을 초과하지 않아야 한다

        Args:
            user_id: 사용자 ID

        Returns:
            JWT 페이로드
        """
        now = datetime.now(timezone.utc)
        return {
            "user_id": user_id,
            "iat": now,
            "exp": now + timedelta(minutes=self.config.expiry_minutes),
        }
```

#### TypeScript: 엄격한 인터페이스

```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth.test.ts

/**
 * 토큰 설정
 *
 * @CODE:AUTH-001:DATA - 데이터 모델
 */
interface TokenConfig {
  secretKey: string;
  algorithm: string;
  expiryMinutes: number;
}

/**
 * JWT 페이로드
 */
interface JWTPayload {
  userId: number;
  iat: number;
  exp: number;
}

/**
 * JWT 인증 서비스
 *
 * @SPEC:AUTH-001 요구사항:
 * - Ubiquitous: JWT 토큰 생성 기능 제공
 * - Event-driven: 유효한 자격증명으로 로그인 시 JWT 토큰 발급
 * - Constraint: 액세스 토큰 만료시간 15분 이내
 *
 * TDD 이력:
 * - v0.0.1 (2025-10-14): RED - 테스트 작성
 * - v0.1.0 (2025-10-14): GREEN - 최소 구현
 * - v0.1.1 (2025-10-14): REFACTOR - 코드 품질 개선
 */
export class JWTService {
  private config: TokenConfig;

  constructor(config?: Partial<TokenConfig>) {
    this.config = {
      secretKey: config?.secretKey ?? "secret",
      algorithm: config?.algorithm ?? "HS256",
      expiryMinutes: config?.expiryMinutes ?? 15,
    };
  }

  /**
   * JWT 토큰 생성
   *
   * @SPEC:AUTH-001 Ubiquitous:
   * 시스템은 JWT 토큰 생성 기능을 제공해야 한다
   *
   * @param userId - 사용자 ID (양수)
   * @returns JWT 토큰 문자열
   * @throws {Error} userId가 0 이하일 때
   */
  generateToken(userId: number): string {
    // 가드절: 유효하지 않은 입력 처리
    if (userId <= 0) {
      throw new Error("userId must be positive");
    }

    // @CODE:AUTH-001:DOMAIN - 비즈니스 로직
    const payload = this.createPayload(userId);
    return this.encodeToken(payload);
  }

  /**
   * 페이로드 생성 (내부 헬퍼)
   *
   * @SPEC:AUTH-001 Constraint:
   * 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
   */
  private createPayload(userId: number): JWTPayload {
    const now = Math.floor(Date.now() / 1000);
    return {
      userId,
      iat: now,
      exp: now + this.config.expiryMinutes * 60,
    };
  }
}
```

### 린터 규칙

#### Python: ruff

```toml
# pyproject.toml
[tool.ruff]
line-length = 100
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "I", "N", "W", "UP", "B", "A", "C4", "SIM"]
ignore = ["E501"]  # 라인 길이는 100으로 설정

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
```

**실행**:

```bash
# 린터 실행
ruff check src/

# 자동 수정
ruff check --fix src/

# 포맷팅
ruff format src/
```

#### TypeScript: Biome

```json
// biome.json
{
  "$schema": "https://biomejs.dev/schemas/1.5.0/schema.json",
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "complexity": {
        "noExcessiveCognitiveComplexity": {
          "level": "error",
          "options": {
            "maxComplexity": 10
          }
        }
      },
      "style": {
        "useConsistentArrayType": "error"
      }
    }
  },
  "formatter": {
    "enabled": true,
    "lineWidth": 100,
    "indentStyle": "space"
  }
}
```

**실행**:

```bash
# 린터 실행
biome check src/

# 자동 수정
biome check --apply src/
```

---

## U - Unified (통합 아키텍처)

### 원칙

**SPEC 기반 복잡도 관리**

각 SPEC은 복잡도 임계값을 정의하고, 초과 시 새로운 SPEC 또는 면제(Waiver)가 필요합니다.

### SPEC 기반 아키텍처

```
도메인 경계 = SPEC 경계
```

도메인 경계는 언어 관례가 아닌 SPEC에 의해 정의됩니다.

### 복잡도 제약

| 항목 | 제약 | 측정 도구 |
|------|------|---------|
| **순환 복잡도** | ≤ 10 | radon (Python), complexity-report (TS) |
| **인지 복잡도** | ≤ 15 | SonarQube |
| **중첩 깊이** | ≤ 4 | 린터 |
| **함수 길이** | ≤ 50 LOC | 린터 |

### 복잡도 측정 (Python)

```bash
# radon 설치
pip install radon

# 순환 복잡도 측정
radon cc src/ -a

# 결과 예시:
# src/auth/jwt_service.py
#     M 12:4 JWTService.generate_token - A (5)
#     M 24:4 JWTService._create_payload - A (2)
# Average complexity: A (3.5)
```

### 복잡도 리팩토링 예시

#### Before (복잡도 15)

```python
def process_payment(user_id: int, amount: float, method: str) -> bool:
    """결제 처리 (복잡도 15)"""
    if user_id <= 0:
        raise ValueError("Invalid user_id")

    if amount <= 0:
        raise ValueError("Invalid amount")

    if method == "card":
        if not validate_card():
            return False
        if not charge_card(amount):
            return False
        if not update_balance(user_id, -amount):
            rollback_card(amount)
            return False
        send_receipt(user_id)
        return True
    elif method == "paypal":
        if not validate_paypal():
            return False
        if not charge_paypal(amount):
            return False
        if not update_balance(user_id, -amount):
            rollback_paypal(amount)
            return False
        send_receipt(user_id)
        return True
    else:
        raise ValueError("Invalid method")
```

#### After (복잡도 5)

```python
# SPEC 분리: PAYMENT-001 (결제 처리), PAYMENT-002 (카드), PAYMENT-003 (PayPal)

# @CODE:PAYMENT-001 - 결제 처리 (복잡도 5)
def process_payment(user_id: int, amount: float, method: str) -> bool:
    """결제 처리 (리팩토링 후)"""
    validate_input(user_id, amount)

    processor = get_payment_processor(method)
    return processor.process(user_id, amount)


# @CODE:PAYMENT-002 - 카드 결제 (복잡도 3)
class CardProcessor:
    """카드 결제 프로세서"""

    def process(self, user_id: int, amount: float) -> bool:
        """카드 결제 처리"""
        if not self.validate():
            return False

        return self.charge_and_update(user_id, amount)


# @CODE:PAYMENT-003 - PayPal 결제 (복잡도 3)
class PayPalProcessor:
    """PayPal 결제 프로세서"""

    def process(self, user_id: int, amount: float) -> bool:
        """PayPal 결제 처리"""
        if not self.validate():
            return False

        return self.charge_and_update(user_id, amount)
```

---

## S - Secured (보안)

### 원칙

**SPEC 준수 보안**

모든 SPEC에 보안 요구사항, 데이터 민감도, 접근 제어를 명시적으로 정의합니다.

### SPEC 보안 요구사항

```markdown
## 보안 요구사항

### 데이터 민감도
- 사용자 비밀번호: CRITICAL (해시 저장)
- JWT 토큰: HIGH (암호화 전송)
- 사용자 ID: MEDIUM (인증 필요)

### 접근 제어
- JWT 생성: 인증된 사용자만
- JWT 검증: 모든 보호된 엔드포인트

### 보안 제어
- 입력 검증: 모든 사용자 입력
- SQL 인젝션 방지: 파라미터화된 쿼리
- XSS 방지: 출력 이스케이핑
```

### 언어별 보안 패턴

#### Python: 입력 검증 + 보안 라이브러리

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md
from typing import Dict, Any
import jwt
import secrets
from datetime import datetime, timedelta, timezone


class JWTService:
    """JWT 인증 서비스 (보안 강화)"""

    def __init__(self, secret_key: str | None = None):
        # @SPEC:AUTH-001 Security: 안전한 비밀 키 생성
        self.secret_key = secret_key or self._generate_secure_key()

    def _generate_secure_key(self) -> str:
        """
        암호학적으로 안전한 비밀 키 생성

        @SPEC:AUTH-001 Security:
        SHOULD use cryptographically secure random generator
        """
        return secrets.token_urlsafe(32)

    def generate_token(self, user_id: int) -> str:
        """
        JWT 토큰 생성

        @SPEC:AUTH-001 Security:
        - MUST validate user_id (positive integer)
        - SHOULD set expiry time (15 minutes)
        """
        # 입력 검증
        if not isinstance(user_id, int):
            raise TypeError("user_id must be integer")

        if user_id <= 0:
            raise ValueError("user_id must be positive")

        # 페이로드 생성 (안전한 데이터만 포함)
        payload = {
            "user_id": user_id,
            "iat": datetime.now(timezone.utc),
            "exp": datetime.now(timezone.utc) + timedelta(minutes=15),
        }

        # JWT 생성
        return jwt.encode(payload, self.secret_key, algorithm="HS256")

    def verify_token(self, token: str) -> Dict[str, Any]:
        """
        JWT 토큰 검증

        @SPEC:AUTH-001 Security:
        - MUST validate token format
        - MUST check expiry time
        - MUST reject invalid signatures
        """
        # 입력 검증
        if not isinstance(token, str):
            raise TypeError("token must be string")

        if not token:
            raise ValueError("token cannot be empty")

        try:
            # JWT 검증 (자동으로 만료 및 서명 확인)
            return jwt.decode(
                token,
                self.secret_key,
                algorithms=["HS256"],
                options={
                    "verify_signature": True,
                    "verify_exp": True,
                    "require": ["user_id", "iat", "exp"],
                },
            )
        except jwt.ExpiredSignatureError:
            raise AuthenticationError("Token expired")
        except jwt.InvalidTokenError as e:
            raise AuthenticationError(f"Invalid token: {e}")
```

#### 보안 스캔 도구

**Python: bandit**

```bash
# bandit 설치
pip install bandit

# 보안 스캔
bandit -r src/

# 결과 예시:
# [B101:assert_used] Use of assert detected. Consider removing assert statements
# Severity: Low   Confidence: High
```

**TypeScript: eslint-plugin-security**

```bash
# eslint-plugin-security 설치
npm install --save-dev eslint-plugin-security

# .eslintrc.json
{
  "plugins": ["security"],
  "extends": ["plugin:security/recommended"]
}

# 보안 스캔
eslint src/
```

---

## T - Trackable (추적성)

### 원칙

**SPEC-코드 추적성**

모든 코드 변경은 @TAG 시스템을 통해 SPEC ID와 특정 요구사항을 참조합니다.

### 3단계 워크플로우 추적

```
/alfred:1-spec → @SPEC:ID (.moai/specs/)
/alfred:2-build → @TEST:ID (tests/) + @CODE:ID (src/)
/alfred:3-sync → @DOC:ID (docs/) + TAG 검증
```

### 코드 스캔 기반 추적성

**원칙**: 중간 캐시 없이 코드를 직접 스캔

**검증 명령어**:

```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 특정 SPEC 추적
rg 'AUTH-001' -n

# TAG 체인 검증
rg '@SPEC:AUTH-001' -n .moai/specs/
rg '@TEST:AUTH-001' -n tests/
rg '@CODE:AUTH-001' -n src/
rg '@DOC:AUTH-001' -n docs/
```

### 추적성 매트릭스

| SPEC ID | SPEC | TEST | CODE | DOC | 상태 |
|---------|------|------|------|-----|------|
| AUTH-001 | ✅ | ✅ | ✅ | ✅ | 완료 |
| AUTH-002 | ✅ | ✅ | ❌ | ❌ | 구현 중 |
| AUTH-003 | ✅ | ❌ | ❌ | ❌ | 계획 |

---

## 언어별 TRUST 구현

### Python 전체 예시

```bash
# 프로젝트 구조
.
├── src/
│   └── auth/
│       └── jwt_service.py      # @CODE:AUTH-001
├── tests/
│   └── test_jwt_service.py     # @TEST:AUTH-001
├── .moai/
│   └── specs/
│       └── SPEC-AUTH-001/
│           └── spec.md         # @SPEC:AUTH-001
└── pyproject.toml

# TRUST 검증 전체 프로세스
# 1. Test First
pytest tests/ --cov=src --cov-report=term-missing --cov-fail-under=85

# 2. Readable
ruff check src/
mypy src/ --strict

# 3. Unified
radon cc src/ -a -nb

# 4. Secured
bandit -r src/

# 5. Trackable
rg '@(SPEC|TEST|CODE):AUTH-001' -n
```

### TypeScript 전체 예시

```bash
# 프로젝트 구조
.
├── src/
│   └── auth/
│       └── jwt-service.ts      # @CODE:AUTH-001
├── tests/
│   └── auth.test.ts            # @TEST:AUTH-001
├── .moai/
│   └── specs/
│       └── SPEC-AUTH-001/
│           └── spec.md         # @SPEC:AUTH-001
└── package.json

# TRUST 검증 전체 프로세스
# 1. Test First
vitest run --coverage

# 2. Readable
biome check src/
tsc --noEmit

# 3. Unified
complexity-report --format json src/

# 4. Secured
eslint src/ --plugin security

# 5. Trackable
rg '@(SPEC|TEST|CODE):AUTH-001' -n
```

---

## TRUST 검증 도구

### @agent-trust-checker

Alfred의 trust-checker 에이전트는 TRUST 5원칙을 자동으로 검증합니다.

**사용법**:

```bash
@agent-trust-checker "AUTH-001 품질 검증"
```

**검증 항목**:
1. **Test First**: pytest --cov 실행, 커버리지 85% 이상 확인
2. **Readable**: ruff check, mypy --strict 실행
3. **Unified**: radon cc 실행, 복잡도 ≤ 10 확인
4. **Secured**: bandit 실행, 보안 취약점 없음 확인
5. **Trackable**: TAG 체인 무결성 검증

**출력 예시**:

```markdown
## TRUST 검증 결과: AUTH-001

### T - Test First
✅ 테스트 통과: 15/15
✅ 커버리지: 92% (목표 85% 초과)

### R - Readable
✅ 린터 통과: 0 errors, 0 warnings
✅ 타입 체크 통과: mypy strict

### U - Unified
✅ 복잡도: 평균 5.2 (목표 10 이하)
✅ 함수 길이: 최대 45 LOC (목표 50 이하)

### S - Secured
✅ 보안 스캔: 0 issues found

### T - Trackable
✅ TAG 체인 완성: @SPEC → @TEST → @CODE → @DOC

## 종합 평가
✅ TRUST 5원칙 모두 통과
```

---

## 실전 예시

### 예시 1: Python - JWT 인증 시스템

#### TRUST 검증 프로세스

```bash
# 1. Test First
$ pytest tests/test_jwt_service.py --cov=src/auth --cov-report=term-missing

tests/test_jwt_service.py::TestJWTService::test_generate_token PASSED
tests/test_jwt_service.py::TestJWTService::test_verify_token PASSED
tests/test_jwt_service.py::TestJWTService::test_invalid_token PASSED
tests/test_jwt_service.py::TestJWTService::test_token_expiry PASSED

---------- coverage: platform darwin, python 3.13.0 ----------
Name                        Stmts   Miss  Cover   Missing
---------------------------------------------------------
src/auth/jwt_service.py        45      3    93%   23-25
---------------------------------------------------------
TOTAL                          45      3    93%

✅ 커버리지 93% (목표 85% 초과)

# 2. Readable
$ ruff check src/auth/jwt_service.py
All checks passed!

$ mypy src/auth/jwt_service.py --strict
Success: no issues found in 1 source file

✅ 린터 및 타입 체크 통과

# 3. Unified
$ radon cc src/auth/jwt_service.py -a

src/auth/jwt_service.py
    M 12:4 JWTService.generate_token - A (5)
    M 24:4 JWTService._create_payload - A (2)

Average complexity: A (3.5)

✅ 복잡도 평균 3.5 (목표 10 이하)

# 4. Secured
$ bandit -r src/auth/

[main]  INFO    profile include tests: None
[main]  INFO    profile exclude tests: None
[main]  INFO    running on Python 3.13.0
Run started:2025-10-14 14:30:00

Test results:
        No issues identified.

✅ 보안 취약점 없음

# 5. Trackable
$ rg '@(SPEC|TEST|CODE):AUTH-001' -n

.moai/specs/SPEC-AUTH-001/spec.md:19:# @SPEC:AUTH-001
tests/test_jwt_service.py:1:# @TEST:AUTH-001
src/auth/jwt_service.py:1:# @CODE:AUTH-001

✅ TAG 체인 완성
```

---

## 트러블슈팅

### 문제 1: 테스트 커버리지 부족

**증상**:

```bash
$ pytest tests/ --cov=src --cov-fail-under=85

FAILED tests/test_jwt_service.py::coverage
Coverage 78% < 85%
```

**해결**:

```bash
# 1. 커버리지 리포트 확인
pytest tests/ --cov=src --cov-report=html
open htmlcov/index.html

# 2. 누락된 테스트 케이스 추가
vi tests/test_jwt_service.py

# 3. 재테스트
pytest tests/ --cov=src --cov-report=term-missing
```

### 문제 2: 복잡도 초과

**증상**:

```bash
$ radon cc src/auth/payment.py

src/auth/payment.py
    M 12:4 process_payment - C (15)  ← 복잡도 15 (목표 10 초과)
```

**해결**:

```python
# 리팩토링: 복잡도 분해
# Before: 복잡도 15
def process_payment(user_id, amount, method):
    # 복잡한 if-else 체인

# After: 복잡도 5
def process_payment(user_id, amount, method):
    validate_input(user_id, amount)
    processor = get_payment_processor(method)
    return processor.process(user_id, amount)
```

---

## 다음 단계

1. **[3단계 워크플로우 가이드](./workflow.md)**: 1-spec → 2-build → 3-sync 실전 가이드
2. **[@TAG 시스템 가이드](./tag-system.md)**: TAG 체계와 추적성 관리 완전 가이드
3. **[SPEC-First TDD 가이드](./spec-first-tdd.md)**: SPEC 작성 방법과 TDD 구현 상세 가이드
4. **[Alfred SuperAgent 가이드](./alfred-superagent.md)**: Alfred 사용법 완전 가이드

---

**최종 업데이트**: 2025-10-14
**버전**: 0.3.0
**작성자**: MoAI-ADK Documentation Team
