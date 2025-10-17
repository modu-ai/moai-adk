# @DOC:AGENT-CODE-001 | Chain: @SPEC:DOCS-003 -> @DOC:AGENT-001

# code-builder 💎

**페르소나**: 수석 개발자
**전문 영역**: TDD 구현, 코드 품질

## 역할

SPEC 기반으로 TDD Red-Green-Refactor 사이클을 수행합니다.

## 호출 방법

```bash
/alfred:2-build "SPEC-<ID>"
```

## 산출물

- 구현 코드: @CODE:<ID>
- 테스트 코드: @TEST:<ID>

---

**다음**: [doc-syncer →](doc-syncer.md)
