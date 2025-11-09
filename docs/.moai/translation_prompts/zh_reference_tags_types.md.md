Translate the following Korean markdown document to Chinese (Simplified).

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/reference/tags/types.md
**Target Language:** Chinese (Simplified)
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/zh/reference/tags/types.md

**Content to Translate:**

# TAG 타입 상세 가이드

4가지 TAG 타입의 문법, 용도, 예시입니다.

## 1. SPEC TAG

**요구사항 정의**

### 문법

```
SPEC-{숫자}: 설명
```

### 위치

```
.moai/specs/SPEC-001/
├── spec.md         ← SPEC-001
├── requirements.md
└── tests.md
```

### 예시

```markdown
# SPEC-001: 사용자 인증 시스템

사용자 로그인, 회원가입, 비밀번호 관리 기능을 구현한다.
```

### 생성 방법

```bash
# Alfred가 자동 생성
/alfred:1-plan "사용자 인증 시스템"

# 결과: SPEC-001 자동 생성
```

______________________________________________________________________

## 2. @TEST TAG

**테스트 코드 식별**

### 문법

```
@TEST:SPEC-{숫자}:{테스트명}
```

### 위치

```
tests/
├── test_auth.py
│   └─ @TEST:SPEC-001:login_success
│   └─ @TEST:SPEC-001:login_failure
│   └─ @TEST:SPEC-001:password_reset
```

### 예시

```python
# tests/test_auth.py

# @TEST:SPEC-001:login_success
def test_user_login_with_valid_credentials():
    """사용자가 유효한 인증정보로 로그인 성공"""
    user = login("user@example.com", "password123")
    assert user.is_authenticated == True

# @TEST:SPEC-001:login_failure
def test_user_login_with_invalid_password():
    """사용자가 잘못된 비밀번호로 로그인 실패"""
    result = login("user@example.com", "wrong_password")
    assert result.error == "Invalid password"

# @TEST:SPEC-001:password_reset
def test_password_reset():
    """사용자가 비밀번호 초기화 가능"""
    reset_token = request_password_reset("user@example.com")
    assert reset_token is not None
```

### 명명 규칙

| 테스트 타입 | 예시                                           |
| ----------- | ---------------------------------------------- |
| 정상 케이스 | @TEST:SPEC-001:success, @TEST:SPEC-001:valid   |
| 오류 케이스 | @TEST:SPEC-001:failure, @TEST:SPEC-001:invalid |
| 엣지 케이스 | @TEST:SPEC-001:empty, @TEST:SPEC-001:boundary  |

______________________________________________________________________

## 3. @CODE TAG

**구현 코드 식별**

### 문법

```
@CODE:SPEC-{숫자}:{기능명}
```

### 위치

```
src/
├── auth.py
│   └─ @CODE:SPEC-001:login
│   └─ @CODE:SPEC-001:register
│   └─ @CODE:SPEC-001:validate_password
```

### 예시

```python
# src/auth.py

# @CODE:SPEC-001:login
def login(email: str, password: str) -> User:
    """
    사용자 로그인

    @TEST:SPEC-001:login_success 참조
    """
    user = User.query.filter_by(email=email).first()
    if not user or not user.verify_password(password):
        raise AuthenticationError("Invalid credentials")
    return user

# @CODE:SPEC-001:register
def register(email: str, password: str) -> User:
    """
    사용자 등록

    @TEST:SPEC-001:register_success 참조
    """
    if User.query.filter_by(email=email).first():
        raise ValidationError("Email already registered")

    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()
    return user

# @CODE:SPEC-001:validate_password
def validate_password(password: str) -> bool:
    """비밀번호 검증"""
    return len(password) >= 8
```

### 세부 TAG

큰 함수는 세부 TAG로 분할 가능:

```python
# @CODE:SPEC-001:register
def register(email: str, password: str) -> User:
    # @CODE:SPEC-001:validate_email
    if not is_valid_email(email):
        raise ValidationError("Invalid email")

    # @CODE:SPEC-001:hash_password
    password_hash = hash_password(password)

    # @CODE:SPEC-001:create_user
    user = User(email=email, password_hash=password_hash)

    # @CODE:SPEC-001:save_user
    db.session.add(user)
    db.session.commit()

    return user
```

______________________________________________________________________

## 4. @DOC TAG

**문서 식별**

### 문법

```
@DOC:SPEC-{숫자}:{문서명}
```

### 위치

```
docs/
├── api/
│   └─ auth.md @DOC:SPEC-001:api_documentation
├── deployment/
│   └─ auth-deploy.md @DOC:SPEC-001:deployment_guide
├── migration/
│   └─ 001_create_users.sql @DOC:SPEC-001:database_migration
```

### 예시

````markdown
# 사용자 인증 API @DOC:SPEC-001:api_documentation

이 문서는 SPEC-001의 API 명세입니다.

## POST /api/auth/login @DOC:SPEC-001:login_endpoint

사용자 로그인 엔드포인트

**요청**: @CODE:SPEC-001:login 구현 참조
**테스트**: @TEST:SPEC-001:login_success 참조

### 요청 본문
```json
{
  "email": "user@example.com",
  "password": "password123"
}
````

### 응답

```json
{
  "status": "success",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name"
  }
}
```

## DELETE /api/auth/logout @DOC:SPEC-001:logout_endpoint

사용자 로그아웃

````

---

## TAG 통계

### 검증 명령어

```bash
# SPEC-001의 TAG 통계
moai-adk status --spec SPEC-001

# 출력 예시
SPEC-001: 사용자 인증 시스템
├─ @TEST 태그: 5개
├─ @CODE 태그: 8개
├─ @DOC 태그: 3개
└─ ✅ 완성도: 100%
````

______________________________________________________________________

## TAG 작성 팁

### ✅ Good

```python
# @CODE:SPEC-001:validate_email
def validate_email(email):
    pass
```

### :x: Bad

```python
# @CODE:SPEC-001
def validate_email(email):
    pass

# 문제: 함수 이름 미포함
```

### ✅ Good

```python
# @TEST:SPEC-001:email_validation_success
def test_valid_email():
    pass
```

### :x: Bad

```python
# @TEST:SPEC-001:test
def test_valid_email():
    pass

# 문제: 너무 범용적
```

______________________________________________________________________

**다음**: [추적성 시스템](traceability.md) 또는 [TAG 개요](index.md)


**Instructions:**
- Translate the content above to Chinese (Simplified)
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
