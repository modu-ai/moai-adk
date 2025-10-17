# /alfred:2-build

RED-GREEN-REFACTOR 사이클로 기능을 구현합니다.

## Overview

TDD Implementation은 MoAI-ADK 3단계 워크플로우의 두 번째 단계입니다. **"테스트 없이는 구현 없음"** 원칙을 따라 SPEC을 TDD 방식으로 구현합니다.

### 담당 에이전트

- **code-builder** 💎: 수석 개발자
- **역할**: RED (테스트 작성) → GREEN (구현) → REFACTOR (리팩토링)
- **전문성**: 언어별 TDD 패턴, 코드 품질, TRUST 원칙 준수

---

## When to Use

다음과 같은 경우 `/alfred:2-build`를 사용합니다:

- ✅ `/alfred:1-spec`으로 SPEC 작성이 완료되었을 때
- ✅ SPEC의 요구사항이 명확하게 정의되어 있을 때
- ✅ 기존 코드를 리팩토링할 때 (테스트 먼저 작성)
- ✅ 버그 수정 시 (재현 테스트 먼저 작성)

---

## Command Syntax

### Basic Usage

```bash
# 단일 SPEC 구현
/alfred:2-build SPEC-ID

# 예시
/alfred:2-build AUTH-001
/alfred:2-build UPLOAD-003
```

### Advanced Usage

```bash
# 여러 SPEC 순차 구현
/alfred:2-build AUTH-001 AUTH-002

# 모든 pending SPEC 구현
/alfred:2-build all

# 특정 도메인의 모든 SPEC
/alfred:2-build AUTH-*
```

---

## Workflow (2단계)

### Phase 1: 분석 및 계획 수립

Alfred가 다음 작업을 수행합니다:

#### 1. SPEC 문서 분석

```bash
# SPEC 파일 읽기
cat .moai/specs/SPEC-AUTH-001/spec.md

# SPEC 메타데이터 확인
rg "^id:|^version:|^status:" .moai/specs/SPEC-AUTH-001/spec.md
```

Alfred가 분석하는 항목:

- **id**: SPEC 고유 ID (예: AUTH-001)
- **version**: SPEC 버전 (예: 0.0.1)
- **status**: SPEC 상태 (draft, active)
- **priority**: 우선순위 (high, medium, low)
- **EARS Requirements**: 요구사항 목록
- **Technical Design**: 기술 설계 (API, 데이터 모델)
- **Test Plan**: 테스트 전략

#### 2. 프로젝트 언어 감지

```bash
# package.json 확인 (TypeScript/JavaScript)
cat package.json

# requirements.txt 확인 (Python)
cat requirements.txt

# go.mod 확인 (Go)
cat go.mod

# Cargo.toml 확인 (Rust)
cat Cargo.toml
```

#### 3. TDD 구현 계획 생성

Alfred가 제시하는 계획 예시:

```markdown
📋 TDD 구현 계획 - SPEC-AUTH-001

**분석 결과**:
- SPEC ID: AUTH-001
- 제목: JWT 인증 시스템
- 언어: TypeScript
- 테스트 프레임워크: Vitest
- 요구사항: 5개 (Ubiquitous 1, Event-driven 2, Constraints 2)

**구현 전략**:

1. RED Phase (실패하는 테스트 작성)
   - 파일: tests/auth/service.test.ts
   - 테스트 케이스 5개:
     * should authenticate valid user
     * should reject invalid credentials
     * should generate JWT token with 15min expiry
     * should throw error on expired token
     * should validate email format

2. GREEN Phase (최소 구현)
   - 파일: src/auth/service.ts
   - 클래스: AuthService
   - 메서드: authenticate(), generateToken()

3. REFACTOR Phase (품질 개선)
   - 입력 검증 (Zod 스키마)
   - 에러 처리 개선
   - 의존성 주입 적용
   - 타입 안전성 강화

**예상 소요 시간**: 15-20분

진행하시겠습니까? (진행/수정/중단)
```

#### 4. 사용자 확인 대기

- **"진행"**: Phase 2로 이동
- **"수정 [내용]"**: 계획 재수립
- **"중단"**: 작업 취소

---

### Phase 2: TDD 구현 (RED-GREEN-REFACTOR)

사용자가 "진행"하면 Alfred가 TDD 사이클을 실행합니다.

---

## RED-GREEN-REFACTOR Cycle

### 🔴 RED Phase: 실패하는 테스트 작성

#### 목표

- SPEC 요구사항을 테스트 코드로 변환
- 테스트 실행 시 **반드시 실패해야 함** (구현 전이므로)
- 명확한 실패 메시지 확인

#### 언어별 예시

##### TypeScript (Vitest)

```typescript
// @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

import { describe, it, expect, beforeEach } from 'vitest'
import { AuthService } from '@/auth/service'
import type { User } from '@/auth/types'

describe('@TEST:AUTH-001: JWT 인증 시스템', () => {
  let authService: AuthService

  beforeEach(() => {
    authService = new AuthService()
  })

  describe('유효한 자격증명 인증', () => {
    it('should authenticate valid user and return JWT token', async () => {
      // Arrange
      const email = 'user@example.com'
      const password = 'password123'

      // Act
      const result = await authService.authenticate(email, password)

      // Assert
      expect(result.success).toBe(true)
      expect(result.token).toBeDefined()
      expect(result.token).toMatch(/^eyJ/) // JWT 시작 패턴
      expect(result.tokenType).toBe('Bearer')
      expect(result.expiresIn).toBe(900) // 15분 = 900초
    })
  })

  describe('잘못된 자격증명 거부', () => {
    it('should reject invalid email', async () => {
      await expect(
        authService.authenticate('invalid-email', 'password123')
      ).rejects.toThrow('Invalid email format')
    })

    it('should reject wrong password', async () => {
      await expect(
        authService.authenticate('user@example.com', 'wrongpassword')
      ).rejects.toThrow('Invalid credentials')
    })
  })

  describe('토큰 만료 처리', () => {
    it('should generate token with 15 minute expiry', async () => {
      const result = await authService.authenticate('user@example.com', 'password123')
      const decoded = jwt.decode(result.token!) as jwt.JwtPayload

      const now = Math.floor(Date.now() / 1000)
      const expiry = decoded.exp!

      expect(expiry - now).toBeCloseTo(900, -1) // 15분 ±10초
    })
  })
})
```

**테스트 실행 (실패 확인)**:

```bash
$ bun test tests/auth/service.test.ts

❌ FAIL  tests/auth/service.test.ts
  @TEST:AUTH-001: JWT 인증 시스템
    유효한 자격증명 인증
      ✗ should authenticate valid user and return JWT token
        → Error: Cannot find module '@/auth/service'

Tests: 1 failed, 1 total
Time: 0.42s
```

##### Python (pytest)

```python
# @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

import pytest
from datetime import datetime, timedelta
from auth.service import AuthService, AuthResult

class TestAuthService:
    """@TEST:AUTH-001: JWT 인증 시스템"""

    @pytest.fixture
    def auth_service(self):
        """AuthService 인스턴스 생성"""
        return AuthService()

    def test_should_authenticate_valid_user(self, auth_service):
        """유효한 사용자 인증 및 JWT 토큰 발급"""
        # Arrange
        email = "user@example.com"
        password = "password123"

        # Act
        result = auth_service.authenticate(email, password)

        # Assert
        assert result.success is True
        assert result.token is not None
        assert result.token.startswith("eyJ")  # JWT 시작 패턴
        assert result.token_type == "Bearer"
        assert result.expires_in == 900  # 15분 = 900초

    def test_should_reject_invalid_email(self, auth_service):
        """잘못된 이메일 형식 거부"""
        with pytest.raises(ValueError, match="Invalid email format"):
            auth_service.authenticate("invalid-email", "password123")

    def test_should_reject_wrong_password(self, auth_service):
        """잘못된 비밀번호 거부"""
        with pytest.raises(ValueError, match="Invalid credentials"):
            auth_service.authenticate("user@example.com", "wrongpassword")

    def test_should_generate_token_with_15min_expiry(self, auth_service):
        """15분 만료 시간의 토큰 생성"""
        result = auth_service.authenticate("user@example.com", "password123")

        import jwt
        decoded = jwt.decode(result.token, options={"verify_signature": False})

        now = datetime.utcnow()
        expiry = datetime.fromtimestamp(decoded["exp"])

        delta = (expiry - now).total_seconds()
        assert 890 <= delta <= 910  # 15분 ±10초
```

**테스트 실행 (실패 확인)**:

```bash
$ pytest tests/auth/service.test.py -v

============================= test session starts ==============================
collected 4 items

tests/auth/service.test.py::TestAuthService::test_should_authenticate_valid_user FAILED
tests/auth/service.test.py::TestAuthService::test_should_reject_invalid_email FAILED

================================= FAILURES =====================================
E   ModuleNotFoundError: No module named 'auth.service'

========================= 4 failed in 0.18s =================================
```

##### Java (JUnit 5)

```java
// @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

package com.moai.auth;

import org.junit.jupiter.api.*;
import static org.junit.jupiter.api.Assertions.*;

class AuthServiceTest {
    // @TEST:AUTH-001: JWT 인증 시스템

    private AuthService authService;

    @BeforeEach
    void setUp() {
        authService = new AuthService();
    }

    @Test
    @DisplayName("유효한 사용자 인증 및 JWT 토큰 발급")
    void shouldAuthenticateValidUserAndReturnJwtToken() {
        // Arrange
        String email = "user@example.com";
        String password = "password123";

        // Act
        AuthResult result = authService.authenticate(email, password);

        // Assert
        assertTrue(result.isSuccess());
        assertNotNull(result.getToken());
        assertTrue(result.getToken().startsWith("eyJ"));
        assertEquals("Bearer", result.getTokenType());
        assertEquals(900, result.getExpiresIn());
    }

    @Test
    @DisplayName("잘못된 이메일 형식 거부")
    void shouldRejectInvalidEmail() {
        assertThrows(IllegalArgumentException.class, () -> {
            authService.authenticate("invalid-email", "password123");
        }, "Invalid email format");
    }

    @Test
    @DisplayName("잘못된 비밀번호 거부")
    void shouldRejectWrongPassword() {
        assertThrows(AuthenticationException.class, () -> {
            authService.authenticate("user@example.com", "wrongpassword");
        }, "Invalid credentials");
    }
}
```

##### Go (go test)

```go
// @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

package auth_test

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/moai/auth"
)

// TestAuthService @TEST:AUTH-001: JWT 인증 시스템
func TestAuthService_Authenticate_ValidUser(t *testing.T) {
	// Arrange
	service := auth.NewAuthService()
	email := "user@example.com"
	password := "password123"

	// Act
	result, err := service.Authenticate(email, password)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "eyJ", result.Token[:3]) // JWT 시작 패턴
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, 900, result.ExpiresIn)
}

func TestAuthService_Authenticate_InvalidEmail(t *testing.T) {
	service := auth.NewAuthService()

	_, err := service.Authenticate("invalid-email", "password123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestAuthService_Authenticate_WrongPassword(t *testing.T) {
	service := auth.NewAuthService()

	_, err := service.Authenticate("user@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}
```

#### Git 커밋 (RED Phase)

```bash
# 한국어 (locale: ko)
git add tests/
git commit -m "🔴 RED: SPEC-AUTH-001 테스트 작성

@TAG:AUTH-001-red

- 유효한 사용자 인증 테스트
- 잘못된 자격증명 거부 테스트
- 토큰 만료 처리 테스트
- 테스트 실패 확인 완료"

# 영어 (locale: en)
git commit -m "🔴 RED: SPEC-AUTH-001 test cases written

@TAG:AUTH-001-red

- Valid user authentication test
- Invalid credentials rejection test
- Token expiry handling test
- Test failure confirmed"
```

---

### 🟢 GREEN Phase: 테스트 통과하는 최소 구현

#### 목표

- **최소한의 코드**로 모든 테스트 통과
- 완벽함 추구 금지 (REFACTOR에서 개선)
- "작동하는 코드" 우선

#### 언어별 예시

##### TypeScript (최소 구현)

```typescript
// @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.ts

import jwt from 'jsonwebtoken'

export interface AuthResult {
  success: boolean
  token?: string
  tokenType?: string
  expiresIn?: number
}

export class AuthService {
  async authenticate(email: string, password: string): Promise<AuthResult> {
    // 최소 구현: 하드코딩으로 테스트 통과
    if (!email.includes('@')) {
      throw new Error('Invalid email format')
    }

    if (password !== 'password123') {
      throw new Error('Invalid credentials')
    }

    const token = jwt.sign(
      { userId: 'dummy-user-id' },
      'secret',
      { expiresIn: '15m' }
    )

    return {
      success: true,
      token,
      tokenType: 'Bearer',
      expiresIn: 900
    }
  }
}
```

**테스트 실행 (통과 확인)**:

```bash
$ bun test tests/auth/service.test.ts

✓ tests/auth/service.test.ts (4 tests) 234ms
  ✓ @TEST:AUTH-001: JWT 인증 시스템
    ✓ should authenticate valid user and return JWT token
    ✓ should reject invalid email
    ✓ should reject wrong password
    ✓ should generate token with 15 minute expiry

Tests: 4 passed (4 total)
Time: 0.45s
```

##### Python (최소 구현)

```python
# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.py

from dataclasses import dataclass
import jwt
from datetime import datetime, timedelta

@dataclass
class AuthResult:
    success: bool
    token: str | None
    token_type: str
    expires_in: int

class AuthService:
    """사용자 인증 서비스 - 최소 구현"""

    def authenticate(self, email: str, password: str) -> AuthResult:
        # 최소 구현: 하드코딩으로 테스트 통과
        if "@" not in email:
            raise ValueError("Invalid email format")

        if password != "password123":
            raise ValueError("Invalid credentials")

        # JWT 토큰 생성
        payload = {
            "user_id": "dummy-user-id",
            "exp": datetime.utcnow() + timedelta(minutes=15)
        }
        token = jwt.encode(payload, "secret", algorithm="HS256")

        return AuthResult(
            success=True,
            token=token,
            token_type="Bearer",
            expires_in=900
        )
```

**테스트 실행 (통과 확인)**:

```bash
$ pytest tests/auth/service.test.py -v

============================= test session starts ==============================
collected 4 items

tests/auth/service.test.py::TestAuthService::test_should_authenticate_valid_user PASSED
tests/auth/service.test.py::TestAuthService::test_should_reject_invalid_email PASSED
tests/auth/service.test.py::TestAuthService::test_should_reject_wrong_password PASSED
tests/auth/service.test.py::TestAuthService::test_should_generate_token_with_15min_expiry PASSED

============================== 4 passed in 0.32s ================================
```

#### Git 커밋 (GREEN Phase)

```bash
# 한국어
git add src/
git commit -m "🟢 GREEN: SPEC-AUTH-001 최소 구현

@TAG:AUTH-001-green

- AuthService 클래스 구현
- authenticate() 메서드 추가
- JWT 토큰 생성 로직
- 모든 테스트 통과 (4/4)"

# 영어
git commit -m "🟢 GREEN: SPEC-AUTH-001 minimal implementation

@TAG:AUTH-001-green

- AuthService class implemented
- authenticate() method added
- JWT token generation logic
- All tests passing (4/4)"
```

---

### ♻️ REFACTOR Phase: 코드 품질 개선

#### 목표

- 코드 품질 개선 (가독성, 유지보수성)
- TRUST 원칙 적용
- **테스트는 그대로 유지** (변경 금지)
- 리팩토링 후 테스트 재실행 (회귀 방지)

#### 언어별 예시

##### TypeScript (프로덕션 구현)

```typescript
// @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.ts
//
// TDD History:
// - RED: 4 test cases written
// - GREEN: Minimal implementation (hardcoded)
// - REFACTOR: Production-ready implementation with DI, validation, error handling

import jwt from 'jsonwebtoken'
import { z } from 'zod'

// @CODE:AUTH-001:DATA - 데이터 모델
export interface AuthResult {
  success: boolean
  token?: string
  tokenType?: string
  expiresIn?: number
  error?: string
}

export interface User {
  id: string
  email: string
  passwordHash: string
}

// @CODE:AUTH-001:DATA - 입력 검증 스키마
const AuthInputSchema = z.object({
  email: z.string().email('Invalid email format'),
  password: z.string().min(8, 'Password must be at least 8 characters')
})

// @CODE:AUTH-001:DOMAIN - 저장소 인터페이스
export interface UserRepository {
  findByEmail(email: string): Promise<User | null>
}

// @CODE:AUTH-001:DOMAIN - 비즈니스 로직
export class AuthService {
  private readonly JWT_EXPIRY = '15m'
  private readonly JWT_EXPIRY_SECONDS = 900

  constructor(
    private readonly userRepo: UserRepository,
    private readonly jwtSecret: string
  ) {}

  async authenticate(email: string, password: string): Promise<AuthResult> {
    try {
      // 입력 검증
      const input = AuthInputSchema.parse({ email, password })

      // 사용자 조회
      const user = await this.findUser(input.email)

      // 비밀번호 검증
      await this.verifyPassword(input.password, user.passwordHash)

      // JWT 토큰 생성
      return this.generateAuthResult(user.id)

    } catch (error) {
      return this.handleAuthError(error)
    }
  }

  private async findUser(email: string): Promise<User> {
    const user = await this.userRepo.findByEmail(email)
    if (!user) {
      throw new Error('Invalid credentials')
    }
    return user
  }

  private async verifyPassword(plain: string, hash: string): Promise<void> {
    const bcrypt = await import('bcryptjs')
    const valid = await bcrypt.compare(plain, hash)
    if (!valid) {
      throw new Error('Invalid credentials')
    }
  }

  private generateAuthResult(userId: string): AuthResult {
    const token = jwt.sign(
      { userId },
      this.jwtSecret,
      { expiresIn: this.JWT_EXPIRY }
    )

    return {
      success: true,
      token,
      tokenType: 'Bearer',
      expiresIn: this.JWT_EXPIRY_SECONDS
    }
  }

  private handleAuthError(error: unknown): AuthResult {
    const message = error instanceof Error ? error.message : 'Authentication failed'

    return {
      success: false,
      error: message
    }
  }
}
```

##### Python (프로덕션 구현)

```python
# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.py
#
# TDD History:
# - RED: 4 test cases written
# - GREEN: Minimal implementation (hardcoded)
# - REFACTOR: Production-ready implementation with DI, validation, error handling

from dataclasses import dataclass
from typing import Protocol, Optional
from datetime import datetime, timedelta
import bcrypt
import jwt
from email_validator import validate_email, EmailNotValidError

# @CODE:AUTH-001:DATA - 데이터 모델
@dataclass
class AuthResult:
    success: bool
    token: Optional[str] = None
    token_type: str = ""
    expires_in: int = 0
    error: Optional[str] = None

@dataclass
class User:
    id: str
    email: str
    password_hash: bytes

# @CODE:AUTH-001:DOMAIN - 저장소 인터페이스
class UserRepository(Protocol):
    """사용자 저장소 인터페이스"""

    def find_by_email(self, email: str) -> Optional[User]:
        """이메일로 사용자 조회"""
        ...

# @CODE:AUTH-001:DOMAIN - 비즈니스 로직
class AuthService:
    """사용자 인증 서비스 - 프로덕션 구현"""

    JWT_EXPIRY_MINUTES = 15
    JWT_EXPIRY_SECONDS = 900

    def __init__(self, user_repo: UserRepository, jwt_secret: str):
        """
        Args:
            user_repo: 사용자 저장소
            jwt_secret: JWT 시크릿 키
        """
        self._user_repo = user_repo
        self._jwt_secret = jwt_secret

    def authenticate(self, email: str, password: str) -> AuthResult:
        """
        사용자 인증

        Args:
            email: 사용자 이메일
            password: 비밀번호

        Returns:
            AuthResult: 인증 결과
        """
        try:
            # 입력 검증
            validated_email = self._validate_email(email)
            self._validate_password(password)

            # 사용자 조회
            user = self._find_user(validated_email)

            # 비밀번호 검증
            self._verify_password(password, user.password_hash)

            # JWT 토큰 생성
            return self._generate_auth_result(user.id)

        except (ValueError, EmailNotValidError) as e:
            return AuthResult(success=False, error=str(e))

    def _validate_email(self, email: str) -> str:
        """이메일 형식 검증"""
        try:
            validated = validate_email(email)
            return validated.email
        except EmailNotValidError:
            raise ValueError("Invalid email format")

    def _validate_password(self, password: str) -> None:
        """비밀번호 길이 검증"""
        if len(password) < 8:
            raise ValueError("Password must be at least 8 characters")

    def _find_user(self, email: str) -> User:
        """사용자 조회 (가드절)"""
        user = self._user_repo.find_by_email(email)
        if not user:
            raise ValueError("Invalid credentials")
        return user

    def _verify_password(self, plain: str, hashed: bytes) -> None:
        """비밀번호 검증 (bcrypt)"""
        if not bcrypt.checkpw(plain.encode(), hashed):
            raise ValueError("Invalid credentials")

    def _generate_auth_result(self, user_id: str) -> AuthResult:
        """JWT 토큰 생성"""
        payload = {
            "sub": user_id,
            "exp": datetime.utcnow() + timedelta(minutes=self.JWT_EXPIRY_MINUTES)
        }
        token = jwt.encode(payload, self._jwt_secret, algorithm="HS256")

        return AuthResult(
            success=True,
            token=token,
            token_type="Bearer",
            expires_in=self.JWT_EXPIRY_SECONDS
        )
```

#### Git 커밋 (REFACTOR Phase)

```bash
# 한국어
git add src/
git commit -m "♻️ REFACTOR: SPEC-AUTH-001 품질 개선

@TAG:AUTH-001-refactor

- 의존성 주입 패턴 적용
- Zod 입력 검증 스키마
- bcrypt 비밀번호 해싱
- 에러 처리 개선
- 가드절 패턴 적용
- 함수 분리 (단일 책임)
- 테스트 통과 유지 (4/4)"

# 영어
git commit -m "♻️ REFACTOR: SPEC-AUTH-001 quality improvement

@TAG:AUTH-001-refactor

- Dependency injection pattern
- Zod input validation schema
- bcrypt password hashing
- Enhanced error handling
- Guard clause pattern
- Function separation (SRP)
- Tests still passing (4/4)"
```

---

## Locale-based Git Commit Messages

Alfred는 `.moai/config.json`의 `locale` 설정에 따라 커밋 메시지를 생성합니다.

### 지원 언어

| Locale | 언어 | RED | GREEN | REFACTOR |
|--------|------|-----|-------|----------|
| `ko` | 한국어 | 🔴 RED: 테스트 작성 | 🟢 GREEN: 구현 | ♻️ REFACTOR: 개선 |
| `en` | English | 🔴 RED: Test written | 🟢 GREEN: Implementation | ♻️ REFACTOR: Improvement |
| `ja` | 日本語 | 🔴 RED: テスト作成 | 🟢 GREEN: 実装 | ♻️ REFACTOR: 改善 |
| `zh` | 中文 | 🔴 RED: 测试编写 | 🟢 GREEN: 实现 | ♻️ REFACTOR: 改进 |

### 커밋 메시지 구조

```
[이모지] [단계]: [SPEC-ID] [설명]

@TAG:[SPEC-ID]-[단계]

[상세 내용]
```

**예시**:

```bash
# 한국어 (ko)
🔴 RED: SPEC-AUTH-001 테스트 작성

@TAG:AUTH-001-red

- 유효한 사용자 인증 테스트
- 잘못된 자격증명 거부 테스트
```

```bash
# 영어 (en)
🔴 RED: SPEC-AUTH-001 test cases written

@TAG:AUTH-001-red

- Valid user authentication test
- Invalid credentials rejection test
```

---

## TRUST Principles Verification

### 자동 검증 (Alfred 수행)

Alfred는 REFACTOR 완료 후 자동으로 TRUST 원칙을 검증합니다:

#### T - Test First

```bash
# 테스트 커버리지 확인
bun test --coverage
pytest --cov=src --cov-report=term-missing
go test -cover ./...

# 목표: ≥85%
```

#### R - Readable

```bash
# 린터 실행
biome check src/
ruff check src/
golangci-lint run

# 복잡도 체크
radon cc src/ -a  # Python
eslint src/ --rule 'complexity: [error, 10]'  # TypeScript
```

#### U - Unified

```bash
# 타입 체크
tsc --noEmit  # TypeScript
mypy src/  # Python
go vet ./...  # Go
```

#### S - Secured

```bash
# 보안 스캔
npm audit
bandit -r src/
gosec ./...
```

#### T - Trackable

```bash
# TAG 체인 검증
rg '@(SPEC|TEST|CODE):AUTH-001' -n
```

### TRUST 검증 보고서

Alfred가 생성하는 보고서 예시:

```markdown
✅ TRUST 검증 완료 - SPEC-AUTH-001

### T - Test First
- ✅ 테스트 커버리지: 92% (목표 85% 초과)
- ✅ 모든 테스트 통과: 4/4

### R - Readable
- ✅ 린터 통과: 0 issues
- ✅ 파일 크기: 156 LOC (≤300)
- ✅ 함수 크기: 평균 18 LOC (≤50)
- ✅ 복잡도: 최대 6 (≤10)

### U - Unified
- ✅ 타입 체크 통과
- ✅ 의존성 주입 적용
- ✅ 인터페이스 기반 설계

### S - Secured
- ✅ 보안 스캔: 0 vulnerabilities
- ✅ bcrypt 해싱 적용
- ✅ 입력 검증 구현

### T - Trackable
- ✅ @SPEC:AUTH-001 존재
- ✅ @TEST:AUTH-001 존재
- ✅ @CODE:AUTH-001 존재
- ✅ TAG 체인 무결성 확인

**TRUST 점수**: 5/5 ✅
```

---

## Best Practices

### 1. RED Phase 작성 요령

✅ **권장사항**:

- SPEC 요구사항을 빠짐없이 테스트로 변환
- AAA 패턴 (Arrange-Act-Assert) 사용
- 명확한 테스트 이름 (`should_xxx`, `test_should_xxx`)
- 실패 메시지가 명확한지 확인

❌ **피해야 할 것**:

- 구현을 먼저 생각하며 테스트 작성
- 애매한 테스트 이름 (`test1`, `testAuth`)
- 너무 많은 assertion (테스트 분리 권장)

### 2. GREEN Phase 구현 요령

✅ **권장사항**:

- 테스트만 통과하는 최소 코드
- 하드코딩도 괜찮음
- "작동하는 코드" 우선

❌ **피해야 할 것**:

- 완벽한 구현 시도 (REFACTOR에서)
- 과도한 추상화
- 테스트에 없는 기능 추가

### 3. REFACTOR Phase 개선 요령

✅ **권장사항**:

- 테스트를 자주 실행하며 개선
- 작은 단위로 리팩토링
- TRUST 원칙 준수
- TDD History 주석 추가

❌ **피해야 할 것**:

- 테스트 코드 수정 (회귀 위험)
- 한 번에 여러 가지 변경
- 리팩토링 중 새 기능 추가

---

## Common Pitfalls

### ❌ Pitfall 1: GREEN Phase에서 과도한 최적화

**잘못된 예**:

```typescript
// GREEN Phase에서 완벽한 구현 시도
class AuthService {
  // 캐싱, 로깅, 메트릭, 재시도 로직...
  // → REFACTOR에서 해야 할 일
}
```

**올바른 예**:

```typescript
// GREEN Phase: 최소 구현
class AuthService {
  authenticate(email, password) {
    if (password === 'password123') {
      return { success: true, token: 'dummy' }
    }
    throw new Error('Invalid credentials')
  }
}
```

### ❌ Pitfall 2: 테스트 없이 코드 수정

**잘못된 예**:

```bash
# 구현 먼저, 테스트는 나중에
/alfred:2-build AUTH-001
→ GREEN Phase부터 시작 (RED 건너뜀)
```

**올바른 예**:

```bash
# 항상 RED → GREEN → REFACTOR 순서
/alfred:2-build AUTH-001
→ RED: 테스트 작성 및 실패 확인
→ GREEN: 최소 구현
→ REFACTOR: 품질 개선
```

### ❌ Pitfall 3: REFACTOR 중 테스트 변경

**잘못된 예**:

```python
# REFACTOR 중 테스트 수정
def test_authenticate():
    # 테스트 로직 변경 → 회귀 위험!
```

**올바른 예**:

```python
# REFACTOR: 테스트는 그대로, 코드만 개선
# 테스트 실패 시 → 코드를 다시 수정
```

---

## Troubleshooting

### Issue 1: 테스트 실패 (GREEN Phase)

**증상**:

```bash
$ bun test
❌ FAIL: Expected token to be defined, but got undefined
```

**해결**:

1. 테스트 코드 재확인 (요구사항과 일치하는지)
2. 구현 코드 디버깅
3. 필요 시 RED Phase로 돌아가기

### Issue 2: TRUST 검증 실패

**증상**:

```bash
❌ TRUST 검증 실패
- Readable: 복잡도 15 (목표 ≤10)
```

**해결**:

```typescript
// 함수 분리로 복잡도 감소
// Before: 복잡도 15
function authenticate(email, password) {
  if (...) {
    if (...) {
      if (...) {
        // 중첩 조건
      }
    }
  }
}

// After: 복잡도 6
function authenticate(email, password) {
  validateEmail(email)
  validatePassword(password)
  const user = findUser(email)
  verifyPassword(password, user.hash)
  return generateToken(user.id)
}
```

### Issue 3: TAG 체인 끊김

**증상**:

```bash
$ rg '@CODE:AUTH-001' -n
# 결과 없음 (CODE가 없음)
```

**해결**:

1. `/alfred:2-build AUTH-001` 재실행
2. TAG BLOCK 주석 추가 확인
3. TAG 검증: `rg '@(SPEC|TEST|CODE):AUTH-001' -n`

---

## Real-world Example: TODO App

### 시나리오: TODO 항목 추가 기능 구현

#### Step 1: SPEC 확인

```markdown
# .moai/specs/SPEC-TODO-001/spec.md

## Requirements
- 시스템은 TODO 항목 추가 기능을 제공해야 한다
- WHEN 사용자가 할 일을 입력하고 추가 버튼을 클릭하면, 시스템은 새로운 TODO 항목을 생성해야 한다
- IF 입력이 비어있으면, 시스템은 항목 추가를 거부해야 한다
```

#### Step 2: TDD 구현

**RED Phase**:

```typescript
// tests/todo.test.ts
describe('@TEST:TODO-001: TODO 항목 추가', () => {
  it('should add new todo item', () => {
    const manager = new TodoManager()
    const todo = manager.addTodo('Buy milk')

    expect(todo.id).toBeDefined()
    expect(todo.text).toBe('Buy milk')
    expect(todo.completed).toBe(false)
  })

  it('should reject empty todo', () => {
    const manager = new TodoManager()

    expect(() => manager.addTodo('')).toThrow('TODO text cannot be empty')
  })
})
```

**GREEN Phase**:

```typescript
// src/todo.ts
export class TodoManager {
  private todos: Todo[] = []

  addTodo(text: string): Todo {
    if (!text.trim()) {
      throw new Error('TODO text cannot be empty')
    }

    const todo = {
      id: crypto.randomUUID(),
      text: text.trim(),
      completed: false
    }

    this.todos.push(todo)
    return todo
  }
}
```

**REFACTOR Phase**:

```typescript
// src/todo.ts (개선 버전)
import { z } from 'zod'

const TodoSchema = z.object({
  text: z.string().min(1, 'TODO text cannot be empty').trim()
})

export class TodoManager {
  private readonly todos: Map<string, Todo> = new Map()

  addTodo(text: string): Todo {
    // 입력 검증
    const { text: validatedText } = TodoSchema.parse({ text })

    // TODO 생성
    const todo: Todo = {
      id: crypto.randomUUID(),
      text: validatedText,
      completed: false,
      createdAt: new Date()
    }

    this.todos.set(todo.id, todo)
    return todo
  }

  getTodos(): ReadonlyArray<Todo> {
    return Array.from(this.todos.values())
  }
}
```

#### Step 3: TRUST 검증

```bash
✅ T - Test: 커버리지 100%
✅ R - Readable: 복잡도 4
✅ U - Unified: 타입 안전
✅ S - Secured: 입력 검증
✅ T - Trackable: TAG 체인 완성
```

---

## Next Steps

TDD 구현이 완료되면 다음 단계로 진행합니다:

1. **[Stage 3: Document Sync](/guides/workflow/3-sync)** - `/alfred:3-sync` 실행
2. **[TRUST Principles](/guides/concepts/trust-principles)** - 품질 원칙 상세
3. **[TAG System](/guides/concepts/tag-system)** - 추적성 시스템 심화

---

<div style="text-align: center; margin-top: 40px;">
  <p><strong>테스트 없이는 구현 없음</strong> 💎</p>
  <p>TDD로 완벽한 품질을 만드세요!</p>
</div>
