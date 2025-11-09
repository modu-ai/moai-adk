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

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/reference/tags/traceability.md
**Target Language:** Chinese (Simplified)
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/zh/reference/tags/traceability.md

**Content to Translate:**

# 추적성 시스템 완전 가이드

TAG 체인을 통한 SPEC-TEST-CODE-DOC 완전 추적성 구축입니다.

## :bullseye: 추적성 원칙

모든 기능은 다음 4계층을 모두 포함해야 합니다:

```
1. SPEC    (요구사항) ← Alfred가 작성
2. @TEST   (테스트)   ← tdd-implementer가 작성
3. @CODE   (구현)     ← tdd-implementer가 작성
4. @DOC    (문서)     ← doc-syncer가 작성
```

## 📊 TAG 체인 다이어그램

```
SPEC-001 ────┐
             │
             ├─→ @TEST:SPEC-001:*
             │
             ├─→ @CODE:SPEC-001:*
             │
             ├─→ @DOC:SPEC-001:*
             │
             └─→ 모두 상호 참조
```

## :link: 완전한 추적성 예시

### Step 1: SPEC 작성

```markdown
# SPEC-001: 사용자 인증

## 요구사항
- 로그인 기능
- 회원가입
- 비밀번호 초기화
```

### Step 2: @TEST 작성

```python
# tests/test_auth.py

# @TEST:SPEC-001:login_success
def test_login_success():
    user = login("user@example.com", "pass123")
    assert user is not None

# @TEST:SPEC-001:register_success
def test_register_success():
    user = register("new@example.com", "pass123")
    assert user.email == "new@example.com"

# @TEST:SPEC-001:password_reset_success
def test_password_reset():
    token = request_reset("user@example.com")
    assert token is not None
```

### Step 3: @CODE 작성

```python
# src/auth.py

# @CODE:SPEC-001:login
def login(email, password):
    """@TEST:SPEC-001:login_success 테스트 대상"""
    user = User.query.filter_by(email=email).first()
    if user and user.verify_password(password):
        return user
    return None

# @CODE:SPEC-001:register
def register(email, password):
    """@TEST:SPEC-001:register_success 테스트 대상"""
    if User.query.filter_by(email=email).first():
        return None
    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()
    return user

# @CODE:SPEC-001:password_reset
def request_password_reset(email):
    """@TEST:SPEC-001:password_reset_success 테스트 대상"""
    user = User.query.filter_by(email=email).first()
    if user:
        token = generate_reset_token(user)
        return token
    return None
```

### Step 4: @DOC 작성

````markdown
# 사용자 인증 API @DOC:SPEC-001:api_docs

이 문서는 @SPEC-001의 구현을 설명합니다.

## 로그인 엔드포인트

**구현**: @CODE:SPEC-001:login
**테스트**: @TEST:SPEC-001:login_success

```bash
POST /api/auth/login
````

## 회원가입 엔드포인트

**구현**: @CODE:SPEC-001:register **테스트**: @TEST:SPEC-001:register_success

```bash
POST /api/auth/register
```

## 비밀번호 초기화

**구현**: @CODE:SPEC-001:password_reset **테스트**: @TEST:SPEC-001:password_reset_success

```bash
POST /api/auth/reset-password
```

````

## ✅ 추적성 검증

### 자동 검증

```bash
# /alfred:3-sync에서 자동 실행
moai-adk tag-agent

1. SPEC 감지: SPEC-001 ✓
2. @TEST 확인: 3개 ✓
3. @CODE 확인: 3개 ✓
4. @DOC 확인: 1개 ✓
5. 체인 완성: 완료 ✅

결과: SPEC-001 추적성 100%
````

### 매뉴얼 검증

```bash
# 특정 SPEC 상태 확인
moai-adk status --spec SPEC-001

# 결과:
SPEC-001: 사용자 인증
├─ @TEST:SPEC-001:* (3개)
├─ @CODE:SPEC-001:* (3개)
├─ @DOC:SPEC-001:* (1개)
└─ ✅ 완성도: 100%
```

## 🚨 추적성 문제 진단

### Problem 1: 고아 TAG

```python
# :x: 문제: SPEC이 없는 TAG
@CODE:SPEC-999:orphan_function
def some_function():
    pass

# 해결: SPEC-999 생성 또는 TAG 제거
```

### Problem 2: 불완전한 체인

```python
# :x: 문제: TEST와 CODE는 있지만 DOC 없음
@TEST:SPEC-001:test
def test_feature():
    pass

@CODE:SPEC-001:feature
def feature():
    pass

# 해결: @DOC:SPEC-001 생성
```

### Problem 3: TAG 중복

```python
# :x: 문제: 같은 TAG가 여러 파일에 있음
# file1.py:
@CODE:SPEC-001:register

# file2.py:
@CODE:SPEC-001:register  # 중복!

# 해결: 태그 이름 변경 또는 통합
```

## 🔄 TAG 중복 제거 프로세스

```bash
# 1. 중복 스캔
/alfred:tag-dedup --scan-only

# 2. 계획 수립
/alfred:tag-dedup --dry-run

# 3. 백업 생성 후 적용
/alfred:tag-dedup --apply --backup

# 결과: 모든 TAG 고유, 체인 완성
```

## 📈 추적성 메트릭

### 계산 방식

```
추적성 = (TEST + CODE + DOC) / (SPEC × 3) × 100%

예:
- SPEC: 5개
- @TEST: 15개 (5 × 3)
- @CODE: 15개 (5 × 3)
- @DOC: 5개 (5 × 1)

추적성 = (15 + 15 + 5) / (5 × 3) × 100% = 100%
```

### 목표

| Level     | Traceability | Status       |
| --------- | ------------ | ------------ |
| Excellent | 100%         | ✅ 배포 가능 |
| Good      | 90%+         | ⚠️ 검토 필요 |
| Fair      | 70%+         | 🚨 개선 필요 |
| Poor      | \<70%        | :x: 배포 불가 |

______________________________________________________________________

**다음**: [TAG 타입](types.md) 또는 [TAG 개요](index.md)


**Instructions:**
- Translate the content above to Chinese (Simplified)
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
