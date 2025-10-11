# SPEC-First TDD 워크플로우 가이드

<!-- @CODE:DOCS-002 | SPEC: .moai/specs/SPEC-DOCS-002/spec.md -->

> "명세 없으면 코드 없다. 테스트 없으면 구현 없다."

## 개요

**SPEC-First TDD**는 MoAI-ADK의 핵심 개발 방법론입니다. 명세(SPEC) 작성부터 시작하여 테스트 주도 개발(TDD)로 구현하고, 문서 동기화로 완성하는 3단계 워크플로우를 따릅니다.

### SPEC-First TDD의 철학

- **명세 우선**: 코드 작성 전 명확한 요구사항 정의
- **테스트 주도**: 실패하는 테스트부터 작성
- **점진적 개선**: RED → GREEN → REFACTOR 사이클
- **완벽한 추적성**: @TAG 시스템으로 SPEC부터 코드까지 연결

---

## Alfred SuperAgent의 역할

**Alfred**는 MoAI-ADK의 중앙 오케스트레이터로, 3단계 워크플로우를 조율합니다.

### Alfred의 책임

- **요청 분석**: 사용자 요청의 본질 파악
- **작업 라우팅**: 적절한 전문 에이전트에게 위임
- **품질 보장**: TRUST 5원칙 및 @TAG 체인 검증
- **결과 통합**: 각 단계 완료 후 통합 보고

### Alfred 워크플로우

```
사용자 요청
    ↓
Alfred 분석
    ↓
1-spec → 2-build → 3-sync
    ↓
품질 게이트 검증
    ↓
Alfred 최종 보고
```

---

## 3단계 워크플로우

### 1단계: `/alfred:1-spec` - SPEC 작성

**목적**: EARS 방식으로 명확한 요구사항 작성

**실행**:

```bash
/alfred:1-spec "JWT 인증 시스템"
```

**자동 수행 작업**:

1. 프로젝트 문서 분석 (product.md 등)
2. SPEC 후보 제안 및 사용자 승인
3. `.moai/specs/SPEC-{ID}/spec.md` 생성
4. EARS 구문으로 요구사항 작성
5. Git 브랜치 생성 (`feature/SPEC-{ID}`)
6. Draft PR 생성 (Team 모드)

**결과물**:

```markdown
# .moai/specs/SPEC-AUTH-001/spec.md
---
id: AUTH-001
version: 0.0.1
status: draft
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY
### v0.0.1 (2025-10-11)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성

## Requirements
### Ubiquitous Requirements
- 시스템은 JWT 기반 인증 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
```

**Git 상태** (Team 모드):

```bash
$ git branch
* feature/SPEC-AUTH-001

$ gh pr list
#42 [Draft] SPEC-AUTH-001: JWT 인증 시스템
```

---

### 2단계: `/alfred:2-build` - TDD 구현

**목적**: RED-GREEN-REFACTOR 사이클로 테스트 주도 구현

**실행**:

```bash
/alfred:2-build SPEC-AUTH-001
```

**자동 수행 작업**:

1. SPEC 문서 분석 및 구현 계획 수립
2. 사용자 승인 후 TDD 구현 시작
3. RED: 실패하는 테스트 작성
4. GREEN: 테스트 통과하는 최소 구현
5. REFACTOR: 코드 품질 개선
6. 각 단계별 Git 커밋 (TDD 이력 보존)

#### RED 단계: 🔴 실패하는 테스트 작성

```python
# tests/auth/service.test.py
# @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

import pytest
from auth.service import AuthService

def test_should_authenticate_valid_user():
    """유효한 사용자 인증 테스트"""
    # Arrange
    auth = AuthService()

    # Act
    result = auth.authenticate("user@example.com", "password123")

    # Assert
    assert result.success is True
    assert result.token is not None
    assert result.token_type == "Bearer"

def test_should_reject_invalid_credentials():
    """잘못된 자격증명 거부 테스트"""
    auth = AuthService()
    result = auth.authenticate("user@example.com", "wrongpassword")
    assert result.success is False
```

**테스트 실행**:

```bash
$ pytest tests/auth/
FAILED tests/auth/service.test.py::test_should_authenticate_valid_user
# ImportError: cannot import name 'AuthService'
```

**Git 커밋**:

```bash
🔴 RED: SPEC-AUTH-001 테스트 작성 (실패 확인)

@TEST:AUTH-001
```

#### GREEN 단계: 🟢 테스트 통과하는 최소 구현

```python
# src/auth/service.py
# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.py

from dataclasses import dataclass

@dataclass
class AuthResult:
    success: bool
    token: str | None
    token_type: str

class AuthService:
    """사용자 인증 서비스 - 최소 구현"""

    def authenticate(self, username: str, password: str) -> AuthResult:
        # 최소 구현: 테스트 통과만 목표
        if password == "password123":
            return AuthResult(
                success=True,
                token="dummy_token",
                token_type="Bearer"
            )
        return AuthResult(success=False, token=None, token_type="")
```

**테스트 실행**:

```bash
$ pytest tests/auth/
PASSED tests/auth/service.test.py::test_should_authenticate_valid_user
PASSED tests/auth/service.test.py::test_should_reject_invalid_credentials
```

**Git 커밋**:

```bash
🟢 GREEN: SPEC-AUTH-001 최소 구현 (테스트 통과)

@CODE:AUTH-001
- AuthService 클래스 구현
- 기본 인증 로직 (테스트 통과 목표)
```

#### REFACTOR 단계: 🔄 품질 개선

```python
# src/auth/service.py (리팩토링 완료)
# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.py

import bcrypt
import jwt
from datetime import datetime, timedelta
from typing import Protocol

# @CODE:AUTH-001:DATA - 데이터 모델
@dataclass
class AuthResult:
    success: bool
    token: str | None
    token_type: str

# @CODE:AUTH-001:DOMAIN - 인터페이스
class UserRepository(Protocol):
    def find_by_username(self, username: str) -> User | None:
        ...

# @CODE:AUTH-001:DOMAIN - 비즈니스 로직
class AuthService:
    """사용자 인증 서비스 - 프로덕션 구현"""

    def __init__(self, user_repo: UserRepository, secret_key: str):
        self._user_repo = user_repo
        self._secret_key = secret_key

    def authenticate(self, username: str, password: str) -> AuthResult:
        """사용자 인증"""
        # 가드절: 사용자 조회
        user = self._user_repo.find_by_username(username)
        if not user:
            return self._failed_auth()

        # 가드절: 비밀번호 검증
        if not self._verify_password(password, user.password_hash):
            return self._failed_auth()

        # JWT 토큰 생성
        token = self._generate_jwt(user.id)
        return AuthResult(success=True, token=token, token_type="Bearer")

    def _verify_password(self, plain: str, hashed: bytes) -> bool:
        """비밀번호 검증 (bcrypt)"""
        return bcrypt.checkpw(plain.encode(), hashed)

    def _generate_jwt(self, user_id: int) -> str:
        """JWT 토큰 생성 (15분 만료)"""
        payload = {
            "sub": user_id,
            "exp": datetime.utcnow() + timedelta(minutes=15)
        }
        return jwt.encode(payload, self._secret_key, algorithm="HS256")

    def _failed_auth(self) -> AuthResult:
        """인증 실패 응답"""
        return AuthResult(success=False, token=None, token_type="")
```

**테스트 커버리지 확인**:

```bash
$ pytest --cov=src/auth --cov-report=term-missing
---------- coverage: platform darwin, python 3.11 -----------
Name                    Stmts   Miss  Cover   Missing
-----------------------------------------------------
src/auth/service.py        28      2    93%   45, 52
-----------------------------------------------------
TOTAL                      28      2    93%
```

**Git 커밋**:

```bash
♻️ REFACTOR: SPEC-AUTH-001 품질 개선 (커버리지 93%)

@CODE:AUTH-001
- 의존성 주입 패턴 적용
- bcrypt 비밀번호 해싱
- JWT 토큰 생성 (15분 만료)
- 가드절 패턴 적용
- 테스트 커버리지 93% 달성
```

---

### 3단계: `/alfred:3-sync` - 문서 동기화

**목적**: Living Document 생성 및 TAG 체인 검증

**실행**:

```bash
/alfred:3-sync
```

**자동 수행 작업**:

1. 코드 변경사항 분석
2. TAG 체인 검증 (@SPEC → @TEST → @CODE)
3. Living Document 자동 생성
4. PR 상태 Draft → Ready 전환 (Team 모드)
5. CI/CD 확인 후 자동 머지 (Team 모드, --auto-merge)

**TAG 체인 검증**:

```bash
$ rg '@(SPEC|TEST|CODE):AUTH-001' -n

.moai/specs/SPEC-AUTH-001/spec.md:7:# @SPEC:AUTH-001: JWT 인증 시스템
tests/auth/service.test.py:1:# @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md
src/auth/service.py:1:# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md
```

**Living Document 생성**:

```markdown
# docs/features/auth/jwt-authentication.md
<!-- @DOC:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md -->

# JWT 인증 시스템

이 문서는 @SPEC:AUTH-001에 정의된 JWT 인증 시스템을 설명합니다.

## 개요
- **SPEC**: AUTH-001
- **버전**: 0.0.1
- **상태**: 구현 완료

## 사용 방법
...
```

**Git 상태** (Team 모드):

```bash
$ gh pr view 42
#42 SPEC-AUTH-001: JWT 인증 시스템
  ✅ Ready for review
  ✅ All checks passed
  ✅ Auto-merge enabled
```

---

## 모드별 차이점

### Personal 모드

**특징**: 로컬 Git 워크플로우

```bash
# 1단계: SPEC 작성
/alfred:1-spec "새 기능"
→ feature/SPEC-{ID} 브랜치 생성 (main/develop 기반)
→ SPEC 문서 작성 및 커밋

# 2단계: TDD 구현
/alfred:2-build SPEC-{ID}
→ RED → GREEN → REFACTOR 커밋

# 3단계: 문서 동기화
/alfred:3-sync
→ Living Document 생성
→ TAG 체인 검증
→ 로컬 머지 (develop 또는 main)
```

### Team 모드

**특징**: GitHub PR 자동화

```bash
# 1단계: SPEC 작성
/alfred:1-spec "새 기능"
→ feature/SPEC-{ID} 브랜치 생성 (develop 기반)
→ SPEC 문서 작성 및 커밋
→ Draft PR 자동 생성 ✨

# 2단계: TDD 구현
/alfred:2-build SPEC-{ID}
→ RED → GREEN → REFACTOR 커밋
→ PR 업데이트 (자동)

# 3단계: 문서 동기화
/alfred:3-sync --auto-merge
→ Living Document 생성
→ TAG 체인 검증
→ PR Ready 전환 ✨
→ CI/CD 확인
→ PR 자동 머지 (squash) ✨
→ develop 체크아웃
→ 다음 작업 준비 완료 ✅
```

---

## 실전 예제: TODO App 기능 추가

### 시나리오: TODO 항목에 우선순위 추가

#### Step 1: SPEC 작성

```bash
$ /alfred:1-spec "TODO 항목에 우선순위(high, medium, low) 필드 추가"

# Alfred 응답:
SPEC 후보 제안:
- id: TODO-PRIORITY-001
- 제목: TODO 우선순위 필드 추가
- 도메인: TODO, PRIORITY

진행하시겠습니까? (진행/수정/중단)

$ 진행

# 자동 생성:
# .moai/specs/SPEC-TODO-PRIORITY-001/spec.md
```

**생성된 SPEC**:

```markdown
---
id: TODO-PRIORITY-001
version: 0.0.1
status: draft
---

# @SPEC:TODO-PRIORITY-001: TODO 우선순위 필드 추가

## HISTORY
### v0.0.1 (2025-10-11)
- **INITIAL**: TODO 항목에 우선순위 필드 추가 명세 작성

## Requirements
### Ubiquitous Requirements
- 시스템은 TODO 항목에 우선순위(high, medium, low) 필드를 제공해야 한다

### Event-driven Requirements
- WHEN TODO를 생성하면, 시스템은 기본 우선순위를 'medium'으로 설정해야 한다
- WHEN 우선순위를 변경하면, 시스템은 변경사항을 즉시 저장해야 한다

### Constraints
- 우선순위는 'high', 'medium', 'low' 중 하나여야 한다
- IF 유효하지 않은 우선순위가 입력되면, 시스템은 에러를 반환해야 한다
```

**Git 상태** (Team 모드):

```bash
$ git log --oneline -1
a1b2c3d 🔴 RED: SPEC-TODO-PRIORITY-001 명세 작성

$ gh pr list
#43 [Draft] SPEC-TODO-PRIORITY-001: TODO 우선순위 필드 추가
```

#### Step 2: TDD 구현

```bash
$ /alfred:2-build SPEC-TODO-PRIORITY-001

# Alfred 응답:
구현 계획:
- RED: TODO 모델 및 API 테스트 작성
- GREEN: 우선순위 필드 추가 (최소 구현)
- REFACTOR: 타입 안전성 및 검증 로직 개선

진행하시겠습니까? (진행/수정/중단)

$ 진행
```

**RED: 테스트 작성**:

```python
# tests/todo/model.test.py
# @TEST:TODO-PRIORITY-001 | SPEC: .moai/specs/SPEC-TODO-PRIORITY-001/spec.md

def test_should_create_todo_with_default_priority():
    todo = Todo(title="새 할일")
    assert todo.priority == "medium"

def test_should_update_priority():
    todo = Todo(title="할일", priority="low")
    todo.priority = "high"
    assert todo.priority == "high"

def test_should_reject_invalid_priority():
    with pytest.raises(ValueError):
        Todo(title="할일", priority="invalid")
```

```bash
$ pytest tests/todo/
FAILED (ImportError: cannot import name 'priority')

$ git log --oneline -1
b2c3d4e 🔴 RED: SPEC-TODO-PRIORITY-001 테스트 작성 (실패 확인)
```

**GREEN: 최소 구현**:

```python
# src/todo/model.py
# @CODE:TODO-PRIORITY-001 | SPEC: .moai/specs/SPEC-TODO-PRIORITY-001/spec.md

from dataclasses import dataclass, field

@dataclass
class Todo:
    title: str
    priority: str = "medium"  # 기본값 추가

    def __post_init__(self):
        if self.priority not in ["high", "medium", "low"]:
            raise ValueError(f"Invalid priority: {self.priority}")
```

```bash
$ pytest tests/todo/
PASSED (3/3)

$ git log --oneline -1
c3d4e5f 🟢 GREEN: SPEC-TODO-PRIORITY-001 최소 구현 (테스트 통과)
```

**REFACTOR: 품질 개선**:

```python
# src/todo/model.py (리팩토링)
# @CODE:TODO-PRIORITY-001 | SPEC: .moai/specs/SPEC-TODO-PRIORITY-001/spec.md

from dataclasses import dataclass
from enum import Enum

# @CODE:TODO-PRIORITY-001:DATA - 우선순위 타입
class Priority(str, Enum):
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"

# @CODE:TODO-PRIORITY-001:DATA - TODO 모델
@dataclass
class Todo:
    title: str
    priority: Priority = Priority.MEDIUM

    def __post_init__(self):
        if isinstance(self.priority, str):
            self.priority = Priority(self.priority)
```

```bash
$ pytest --cov=src/todo
Coverage: 95%

$ git log --oneline -1
d4e5f6g ♻️ REFACTOR: SPEC-TODO-PRIORITY-001 품질 개선 (커버리지 95%)
```

#### Step 3: 문서 동기화

```bash
$ /alfred:3-sync --auto-merge

# Alfred 수행 작업:
1. TAG 체인 검증
   ✅ @SPEC:TODO-PRIORITY-001
   ✅ @TEST:TODO-PRIORITY-001
   ✅ @CODE:TODO-PRIORITY-001

2. Living Document 생성
   ✅ docs/features/todo/priority.md

3. PR 상태 전환
   ✅ Draft → Ready for review
   ✅ CI/CD 통과 확인
   ✅ PR 자동 머지 (squash)

4. develop 체크아웃
   ✅ 다음 작업 준비 완료
```

**최종 Git 이력**:

```bash
$ git log --oneline --graph
*   e5f6g7h Merge pull request #43 from feature/SPEC-TODO-PRIORITY-001
|\
| * d4e5f6g ♻️ REFACTOR: SPEC-TODO-PRIORITY-001 품질 개선 (커버리지 95%)
| * c3d4e5f 🟢 GREEN: SPEC-TODO-PRIORITY-001 최소 구현 (테스트 통과)
| * b2c3d4e 🔴 RED: SPEC-TODO-PRIORITY-001 테스트 작성 (실패 확인)
| * a1b2c3d 🔴 RED: SPEC-TODO-PRIORITY-001 명세 작성
|/
* f6g7h8i (develop) 이전 작업...
```

---

## 베스트 프랙티스

### SPEC 작성 시

✅ **권장사항**:

- EARS 구문을 엄격히 따르기
- 측정 가능한 기준 명시
- 제약사항 명확히 정의
- 관련 SPEC 참조 (related_specs)

❌ **피해야 할 것**:

- 모호한 표현 ("사용자 친화적")
- 측정 불가능한 기준 ("빠르게")
- 주체 불명확 ("처리되어야 한다")

### TDD 사이클 팁

✅ **권장사항**:

- RED: 테스트 먼저, 코드는 나중
- GREEN: 최소 구현, 완벽함 추구 금지
- REFACTOR: 품질 개선, 테스트는 그대로
- 각 단계별 Git 커밋 (이력 보존)

❌ **피해야 할 것**:

- GREEN 단계에서 과도한 최적화
- REFACTOR 없이 다음 기능으로 이동
- 테스트 없이 코드 수정

### 문서 동기화 타이밍

✅ **권장사항**:

- TDD 완료 후 즉시 실행
- PR 머지 전 TAG 체인 검증
- CI/CD 통과 확인

❌ **피해야 할 것**:

- 여러 SPEC 누적 후 한꺼번에 동기화
- TAG 검증 없이 PR 머지
- 실패한 테스트 그대로 두기

---

## 문제 해결 시나리오

### 시나리오 1: 테스트 실패

**문제**:

```bash
$ pytest tests/
FAILED tests/auth/service.test.py::test_authenticate
AssertionError: assert result.token is None
```

**해결**:

1. RED 단계로 돌아가기
2. 테스트 케이스 재검토
3. SPEC 요구사항과 테스트 일치 확인
4. 필요 시 SPEC 업데이트 (HISTORY 기록)

### 시나리오 2: TAG 체인 끊김

**문제**:

```bash
$ rg '@(SPEC|TEST|CODE):AUTH-002' -n
.moai/specs/SPEC-AUTH-002/spec.md:7:# @SPEC:AUTH-002
src/auth/oauth.py:1:# @CODE:AUTH-002
# @TEST:AUTH-002 없음!
```

**해결**:

1. `/alfred:2-build SPEC-AUTH-002` 재실행
2. RED 단계에서 테스트 작성
3. TAG 체인 검증: `rg '@TEST:AUTH-002' -n`

### 시나리오 3: PR 충돌

**문제** (Team 모드):

```bash
$ gh pr view 43
#43 SPEC-TODO-PRIORITY-001
  ❌ Merge conflict with develop
```

**해결**:

```bash
# 1. develop 최신화
$ git checkout develop
$ git pull origin develop

# 2. 기능 브랜치로 이동
$ git checkout feature/SPEC-TODO-PRIORITY-001

# 3. develop 머지
$ git merge develop

# 4. 충돌 해결
$ git status
# 충돌 파일 수정 후
$ git add .
$ git commit -m "Resolve merge conflict with develop"

# 5. PR 업데이트
$ git push origin feature/SPEC-TODO-PRIORITY-001
```

---

## Quick Start

### 최소 명령어로 전체 사이클 실행

```bash
# 1. SPEC 작성
$ /alfred:1-spec "새 기능 설명"
# 승인 → "진행"

# 2. TDD 구현
$ /alfred:2-build SPEC-{ID}
# 승인 → "진행"

# 3. 문서 동기화 + 자동 머지 (Team 모드)
$ /alfred:3-sync --auto-merge

# 완료! 🎉
```

---

## 관련 문서

- [EARS 요구사항 작성 가이드](./ears-guide.md)
- [TRUST 5원칙 가이드](./trust-principles.md)
- [TAG 시스템 가이드](./tag-system.md)
- [개발 가이드](../../.moai/memory/development-guide.md)
- [Alfred SuperAgent 가이드](../../../CLAUDE.md)

---

**작성일**: 2025-10-11
**버전**: v1.0.0
**TAG**: @CODE:DOCS-002
