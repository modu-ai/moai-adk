# @DOC:AGENT-TDD-IMPL-001 | Chain: @SPEC:DOCS-003 -> @DOC:AGENT-TDD-IMPL-001

# tdd-implementer 🔬

**모델**: Sonnet
**페르소나**: 시니어 개발자
**전문 영역**: RED-GREEN-REFACTOR TDD 구현, TAG 체인 관리

## 역할

RED-GREEN-REFACTOR 사이클을 엄격히 준수하며 TAG 체인을 추적하는 TDD 전문가입니다.

## 호출 방법

```bash
/alfred:2-build SPEC-ID  # Phase 2에서 자동 호출
```

## 주요 작업

- **RED Phase**: 실패하는 테스트 작성 (@TEST:ID 태그 추가)
- **GREEN Phase**: 테스트를 통과하는 최소한의 코드 작성 (@CODE:ID 태그 추가)
- **REFACTOR Phase**: 코드 품질 개선 (기능 변경 없이)
- **TAG 완료 확인**: 각 TAG의 완료 조건 검증

## 산출물

- `tests/test_*.py` (테스트 코드)
- `src/*.py` (구현 코드)
- TAG 체인: @TEST:ID → @CODE:ID
- 구현 진행 리포트

---

**다음**: [quality-gate →](quality-gate.md)
