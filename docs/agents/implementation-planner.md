# @DOC:AGENT-IMPL-PLANNER-001 | Chain: @SPEC:DOCS-003 -> @DOC:AGENT-IMPL-PLANNER-001

# implementation-planner 📋

**모델**: Sonnet
**페르소나**: 테크니컬 아키텍트
**전문 영역**: SPEC 분석, 구현 전략 수립, 라이브러리 선정

## 역할

SPEC 문서를 분석하여 최적의 구현 전략과 라이브러리 버전을 결정합니다.

## 호출 방법

```bash
/alfred:2-build SPEC-ID  # Phase 1에서 자동 호출
```

## 주요 작업

- SPEC 문서 분석 및 요구사항 추출
- 라이브러리 및 도구 선정 (WebFetch 활용)
- TAG 체인 설계 및 순서 결정
- 단계별 구현 계획 수립
- 리스크 식별 및 대응 방안

## 산출물

- 구현 계획서 (Implementation Plan)
- TAG 체인 설계
- 라이브러리 버전 목록
- 단계별 구현 로드맵

---

**다음**: [tdd-implementer →](tdd-implementer.md)
