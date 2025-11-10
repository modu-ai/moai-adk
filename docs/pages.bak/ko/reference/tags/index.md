# TAG 시스템 완전 참고서

MoAI-ADK의 추적성 시스템의 핵심인 TAG 시스템입니다.

## 🎯 목적

CODE-FIRST 원칙으로 SPEC, TEST, CODE, DOC를 모두 연결하여 **완전한 추적성**을 보장합니다.

```
SPEC-001 (요구사항)
    ↓
@TEST:SPEC-001 (테스트)
    ↓
@CODE:SPEC-001 (구현)
    ↓
@DOC:SPEC-001 (문서)
    ↓
상호 참조 (완전한 추적성)
```

## 📋 TAG 종류

| TAG         | 위치         | 용도        | 예시              |
| ----------- | ------------ | ----------- | ----------------- |
| **SPEC-ID** | .moai/specs/ | 요구사항    | SPEC-001          |
| **@TEST**   | tests/       | 테스트 코드 | @TEST:SPEC-001:\* |
| **@CODE**   | src/         | 구현 코드   | @CODE:SPEC-001:\* |
| **@DOC**    | docs/        | 문서        | @DOC:SPEC-001:\*  |

## ✅ TAG 작성 규칙

### SPEC TAG

```
SPEC-001: 첫 번째 스펙
SPEC-002: 두 번째 스펙
SPEC-N: N번째 스펙
```

### @TEST TAG

```python
# @TEST:SPEC-001:login_success
def test_login_success():
    pass

# @TEST:SPEC-001:login_failure
def test_login_failure():
    pass
```

### @CODE TAG

```python
# @CODE:SPEC-001:register_user
def register_user(email, password):
    pass

# @CODE:SPEC-001:validate_email
def validate_email(email):
    pass
```

### @DOC TAG

```markdown
# API 문서 @DOC:SPEC-001:api

이것은 SPEC-001의 API 문서입니다.
```

## 🔍 TAG 검증 규칙

| 규칙       | 설명                              | 위반 시 |
| ---------- | --------------------------------- | ------- |
| **고유성** | 같은 TAG가 중복되면 안 됨         | 오류    |
| **완성성** | SPEC→TEST→CODE→DOC 모두 있어야 함 | 경고    |
| **일관성** | TAG 형식 일관성                   | 오류    |
| **추적성** | 상호 참조 가능                    | 경고    |

## 🚀 TAG 스캔 및 검증

```bash
# TAG 현황 조회
moai-adk status

# 특정 SPEC TAG 상세 조회
moai-adk status --spec SPEC-001

# TAG 검증 실행
/alfred:3-sync auto SPEC-001

# TAG 중복 제거
/alfred:tag-dedup --dry-run
/alfred:tag-dedup --apply --backup
```

## 📚 상세 가이드

- **[TAG 타입](types.md)** - 각 TAG 타입 상세 설명
- **[추적성 시스템](traceability.md)** - TAG 체인과 완전성 검증

______________________________________________________________________

**다음**: [TAG 타입](types.md) 또는 [추적성 시스템](traceability.md)
