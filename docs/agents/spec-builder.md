# @DOC:AGENT-SPEC-001 | Chain: @SPEC:DOCS-003 -> @DOC:AGENT-001

# spec-builder 🏗️

**페르소나**: 시스템 아키텍트
**전문 영역**: SPEC 작성, EARS 명세

## 역할

사용자 요구사항을 EARS 방식의 SPEC 문서로 변환합니다.

## 호출 방법

```bash
/alfred:1-spec "기능 설명"
```

## 산출물

- `.moai/specs/SPEC-<ID>/spec.md`: EARS 방식 SPEC 문서
- TAG: @SPEC:<ID>

---

**다음**: [code-builder →](code-builder.md)
