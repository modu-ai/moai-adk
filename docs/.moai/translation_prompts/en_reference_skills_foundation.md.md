Translate the following Korean markdown document to English.

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/reference/skills/foundation.md
**Target Language:** English
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/en/reference/skills/foundation.md

**Content to Translate:**

# Foundation Skills 상세 가이드

모든 MoAI-ADK 프로젝트의 기반이 되는 4개 기초 스킬입니다.

## 개요

| 스킬                      | 설명                | 버전 | 호출 방식                                 |
| ------------------------- | ------------------- | ---- | ----------------------------------------- |
| **moai-foundation-trust** | TRUST 5 원칙 검증   | 5.0  | `Skill("moai-foundation-trust")`          |
| **moai-foundation-tags**  | TAG 시스템 (추적성) | 3.2  | `Skill("moai-foundation-tags")`           |
| **moai-alfred-workflow**  | 4단계 워크플로우    | 4.0  | `Skill("moai-alfred-workflow")`           |
| **moai-alfred-ask-user**  | 사용자 상호작용     | 2.1  | `Skill("moai-alfred-ask-user-questions")` |

______________________________________________________________________

## 1. moai-foundation-trust

**TRUST 5 원칙 검증 및 적용**

### 5가지 원칙

#### 🧪 T - Test First

**테스트 주도 개발**

```
요구사항
    ↓
테스트 작성 (RED)
    ↓
최소 구현 (GREEN)
    ↓
코드 개선 (REFACTOR)
    ↓
테스트 커버리지 85%+ 검증
```

**검증 기준**:

- 커버리지 85% 이상
- 모든 테스트 통과
- 엣지 케이스 포함

#### <span class="material-icons">library_books</span> R - Readable

**읽기 쉬운 코드**

```python
# :x: 읽기 어려운 코드
def f(x):
    return sum([i*2 for i in x if i>0])

# ✅ 읽기 쉬운 코드
def double_positive_numbers(numbers):
    """양수를 2배로 만든 리스트 반환"""
    return [num * 2 for num in numbers if num > 0]
```

**검증 항목**:

- MyPy/type-checking 통과
- 명명 규칙 준수 (camelCase/snake_case)
- 함수 길이 50줄 이하
- 복잡도 10 이하

#### :bullseye: U - Unified

**일관된 구조**

```
Project Structure:
src/
  ├── models/       # 데이터 모델
  ├── api/         # API 엔드포인트
  ├── services/    # 비즈니스 로직
  ├── utils/       # 유틸리티
  └── config.py    # 설정

tests/
  ├── unit/        # 단위 테스트
  ├── integration/ # 통합 테스트
  └── e2e/        # E2E 테스트
```

**검증 항목**:

- 프로젝트 구조 일관성
- 명명 규칙 준수
- 임포트 구조 일관성

#### 🔒 S - Secured

**보안 보장**

```python
# :x: 보안 위험
user = User.query.filter_by(
    email=request.args.get('email')  # SQL injection 위험!
).first()

# ✅ 안전한 코드
from sqlalchemy import text
user = User.query.filter_by(
    email=request.args.get('email')  # SQLAlchemy ORM 자동 이스케이프
).first()
```

**검증 항목**:

- OWASP Top 10 취약점 없음
- 의존성 보안 검사 (Snyk, safety)
- 입력 검증 및 이스케이프
- 암호화된 저장소

#### 🏷️ T - Trackable

**완전한 추적성**

```
SPEC-001 (요구사항)
    ↓
@TEST:SPEC-001 (테스트)
    ↓
@CODE:SPEC-001 (구현)
    ↓
@DOC:SPEC-001 (문서)
    ↓
모두 상호 참조됨
```

**검증 항목**:

- 모든 구현에 TAG 포함
- TAG 체인 완성 (SPEC→TEST→CODE→DOC)
- 고아 TAG 없음

### TRUST 검증 자동화

```bash
# TRUST 5 검증 실행
Skill("moai-foundation-trust")

# 검증 결과
✅ Test First: 92% 커버리지
✅ Readable: MyPy 통과
✅ Unified: 구조 준수
✅ Secured: 보안 검사 통과
✅ Trackable: TAG 완성

:bullseye: TRUST 5: PASS ✅
```

______________________________________________________________________

## 2. moai-foundation-tags

**TAG 시스템 완전 가이드**

### TAG 문법

#### SPEC TAG

```
SPEC-001: 첫 번째 스펙
SPEC-002: 두 번째 스펙
```

#### TEST TAG

```
@TEST:SPEC-001:login_feature
@TEST:SPEC-001:password_validation
```

#### CODE TAG

```
@CODE:SPEC-001:register_user
@CODE:SPEC-001:validate_email
```

#### DOC TAG

```
@DOC:SPEC-001:api_documentation
@DOC:SPEC-001:deployment_guide
```

### TAG 체인 예시

```python
# @CODE:SPEC-001:user_registration
def register_user(email: str, password: str) -> User:
    """
    사용자 등록 @CODE:SPEC-001:register_user

    @TEST:SPEC-001:test_register_success 참조
    """
    # @CODE:SPEC-001:validate_email
    if not is_valid_email(email):
        raise ValueError("Invalid email")

    # @CODE:SPEC-001:hash_password
    hashed = hash_password(password)

    # @CODE:SPEC-001:create_user
    user = User(email=email, password_hash=hashed)
    db.session.add(user)
    db.session.commit()

    return user

# @TEST:SPEC-001:test_register_success
def test_register_success():
    """@TEST:SPEC-001:test_register_success"""
    user = register_user("test@example.com", "password123")
    assert user.email == "test@example.com"
    # @TEST:SPEC-001:verify_user_created 검증
    assert user.id is not None
```

### TAG 검증 규칙

| 규칙              | 설명                               |
| ----------------- | ---------------------------------- |
| **중복 금지**     | 같은 TAG가 여러 파일에 있으면 오류 |
| **고아 TAG 금지** | 대응 SPEC 없는 TAG는 제거          |
| **체인 완성**     | SPEC→TEST→CODE→DOC 모두 연결       |
| **명확한 식별**   | TAG는 고유하고 추적 가능해야 함    |

### TAG 스캔

```bash
# TAG 현황 조회
moai-adk status --spec SPEC-001

# TAG 검증
/alfred:3-sync auto SPEC-001

# TAG 중복 제거
/alfred:tag-dedup --dry-run
/alfred:tag-dedup --apply --backup
```

______________________________________________________________________

## 3. moai-alfred-workflow

**4단계 Alfred 워크플로우**

### Phase 1: 의도 파악 (Intent Understanding)

```
사용자 요청 → 명확성 평가
    ├─ 명확: Phase 2로 진행
    └─ 불명확: AskUserQuestion → 사용자 응답 → Phase 2로 진행
```

### Phase 2: 계획 수립 (Plan Creation)

```
Plan Agent 호출
    ↓
├─ 작업 분해 (Decomposition)
├─ 의존성 분석 (Dependency Analysis)
├─ 병렬화 기회 식별 (Parallelization)
├─ 파일 목록 명시 (File List)
└─ 시간 추정 (Time Estimation)
    ↓
사용자 승인 (AskUserQuestion)
    ↓
TodoWrite 초기화
```

### Phase 3: 작업 실행 (Execution)

```
RED Phase
├─ 테스트 작성
└─ 모두 실패 확인

GREEN Phase
├─ 최소 구현
└─ 모두 통과 확인

REFACTOR Phase
├─ 코드 개선
└─ 테스트 유지
```

### Phase 4: 보고 및 커밋 (Report & Commit)

```
작업 완료
    ↓
├─ 문서 생성 (생성 설정에 따라)
├─ Git 커밋 (자동)
├─ PR 생성 (팀 모드)
└─ 정리
```

______________________________________________________________________

## 4. moai-alfred-ask-user-questions

**사용자 상호작용 최적화**

### AskUserQuestion 사용법

```json
{
  "questions": [
    {
      "question": "어떤 인증 방식을 사용하시겠습니까?",
      "header": "Authentication Method",
      "multiSelect": false,
      "options": [
        {
          "label": "JWT",
          "description": "Stateless, REST API에 최적"
        },
        {
          "label": "OAuth 2.0",
          "description": "타사 서비스 통합"
        },
        {
          "label": "Session",
          "description": "기존 웹 애플리케이션"
        }
      ]
    }
  ]
}
```

### 최적 사용 시나리오

| 시나리오       | 사용 여부 | 설명              |
| -------------- | --------- | ----------------- |
| 명확한 요청    | :x:        | 바로 진행         |
| 모호한 요청    | ✅        | 명확히 하기       |
| 기술 결정      | ✅        | 선택 제시         |
| 아키텍처 선택  | ✅        | 트레이드오프 설명 |
| 영향 범위 확인 | ✅        | 사전 고지         |

### 규칙

```
- :x: 이모지 금지 (question, header, label, description)
- ✅ 최대 4개 옵션 (5개 이상은 여러 번 호출)
- ✅ 구조화된 형식 (header + options)
- ✅ 명확한 설명 (각 옵션마다)
```

______________________________________________________________________

## Foundation Skills 통합 예시

```
사용자 요청
    ↓
Skill("moai-alfred-workflow") - 4단계 워크플로우 적용
    ↓
Phase 1: 의도 파악
    └─→ Skill("moai-alfred-ask-user-questions") - 명확화
    ↓
Phase 2: 계획 수립
    └─→ TodoWrite 초기화
    ↓
Phase 3: 작업 실행 (TDD)
    └─→ Skill("moai-foundation-trust") - TRUST 5 검증
    └─→ Skill("moai-foundation-tags") - TAG 추가
    ↓
Phase 4: 보고 및 커밋
    └─→ Git 커밋 (자동)
    ↓
완료!
```

______________________________________________________________________

## Foundation Skills FAQ

### "TRUST 5가 엄격한가요?"

→ **매우 엄격합니다**. 85% 미만 커버리지는 배포 불가능합니다.

### "TAG를 항상 모두 추가해야 하나요?"

→ **네, 추적성을 위해 필수입니다**. 고아 TAG는 자동 제거됩니다.

### "4단계 워크플로우를 생략할 수 있나요?"

→ **아니요, 항상 따릅니다**. Phase 1은 필수입니다.

______________________________________________________________________

**다음**: [Languages Skills](languages.md) 또는 [Alfred Skills](alfred.md)


**Instructions:**
- Translate the content above to English
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
