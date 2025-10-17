# @DOC:START-QUICK-001 | Chain: @SPEC:DOCS-003 -> @DOC:START-001

# Quick Start

5분 안에 MoAI-ADK로 첫 기능을 구현해보세요.

## 1️⃣ SPEC 작성 (1-spec)

Alfred에게 SPEC 작성을 요청합니다:

```bash
# Claude Code에서 실행
/alfred:1-spec "사용자 로그인 기능 구현"
```

Alfred는 spec-builder 에이전트를 호출하여 EARS 방식의 SPEC을 생성합니다:

```markdown
# .moai/specs/SPEC-AUTH-001/spec.md

## @SPEC:AUTH-001 Overview

### EARS Requirements

**Ubiquitous Requirements**:
- REQ-AUTH-001-001: 시스템은 사용자 이메일/비밀번호 로그인을 지원해야 한다

**Event-driven Requirements**:
- REQ-AUTH-001-002: WHEN 사용자가 올바른 자격증명 입력하면,
  시스템은 JWT 토큰을 발급해야 한다
```

## 2️⃣ TDD 구현 (2-build)

SPEC 기반으로 TDD 구현을 요청합니다:

```bash
/alfred:2-build "SPEC-AUTH-001"
```

code-builder 에이전트가 RED-GREEN-REFACTOR 사이클을 수행합니다:

### 🔴 RED: 테스트 작성

```python
# tests/test_auth.py
# @TEST:AUTH-001 | Chain: @SPEC:AUTH-001 -> @TEST:AUTH-001

def test_should_authenticate_valid_user():
    """유효한 자격증명으로 인증 성공"""
    result = authenticate("user@example.com", "password123")
    assert result.success is True
    assert result.token is not None
```

### 🟢 GREEN: 최소 구현

```python
# src/auth.py
# @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 -> @CODE:AUTH-001

def authenticate(email: str, password: str) -> AuthResult:
    """사용자 인증 처리"""
    # 최소 구현
    if email and password:
        return AuthResult(success=True, token="jwt_token")
    return AuthResult(success=False)
```

### 🔄 REFACTOR: 품질 개선

```python
# src/auth.py
# @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 -> @CODE:AUTH-001

import bcrypt
import jwt

def authenticate(email: str, password: str) -> AuthResult:
    """사용자 인증 처리 (보안 강화)"""
    user = User.find_by_email(email)
    if user and bcrypt.verify(password, user.password_hash):
        token = jwt.encode({"sub": user.id}, SECRET_KEY)
        return AuthResult(success=True, token=token)
    return AuthResult(success=False, error="Invalid credentials")
```

## 3️⃣ 문서 동기화 (3-sync)

구현이 완료되면 문서를 동기화합니다:

```bash
/alfred:3-sync
```

doc-syncer 에이전트가 다음을 수행합니다:

1. **TAG 체인 검증**: @SPEC → @CODE → @TEST → @DOC 연결 확인
2. **API 문서 생성**: docstring 기반 자동 생성
3. **README 업데이트**: 새로운 기능 반영

---

## 완성된 TAG 체인

```
@SPEC:AUTH-001 (사용자 인증 요구사항)
  ├─ @CODE:AUTH-001 (인증 구현 코드)
  ├─ @TEST:AUTH-001 (인증 테스트)
  └─ @DOC:AUTH-001 (인증 API 문서)
```

---

## 다음 단계

기본 워크플로우를 이해했다면:

1. [첫 프로젝트 만들기](first-project.md) - TODO 앱 전체 구현
2. [워크플로우 심화](../workflow.md) - 3단계 워크플로우 상세 가이드
3. [Configuration](../configuration/config-json.md) - Personal vs Team 모드 설정

---

**다음**: [First Project →](first-project.md)
