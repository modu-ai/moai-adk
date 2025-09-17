# 프로젝트 가이드라인(요약)

MoAI-ADK의 개발 원칙과 워크플로우 전문은 `@.moai/memory/operations.md`에서 확인합니다. 이 문서는 빠른 참조용 요약본이며, 세부 규칙은 반드시 `.moai/memory` 버전을 참고하세요.

## 핵심 요약
- Spec-First + TDD: `/moai:2-spec` → `/moai:5-dev` 명령 흐름을 따른다.
- Small & Safe 루프: 문제 정의 → 작고 안전한 변경 → 리뷰 → 리팩터링을 반복하고 가정·대안 비교를 기록한다.
- Constitution 5원칙 준수: `@.moai/memory/constitution.md` 확인.
- 체크리스트: `@.moai/memory/operations.md` 4장 참조(테스트/보안/문서/동기화 항목).

### 참조
- 운영 메모 전문: `@.moai/memory/operations.md`
- 헌법 전문: `@.moai/memory/constitution.md`
- 공통 체크리스트: `@.claude/memory/shared_checklists.md`
