# 메모리 시스템 인덱스(요약)

이 폴더는 프로젝트 메모리의 "목차"만 보관함. 세부 지침·체크리스트는 모두 `.moai/memory/`에 위치하며, 필요할 때마다 `@.moai/memory/...` 경로로 임포트해 사용함.

## 우선순위 계층
- 조직 정책 → 프로젝트 메모리(`CLAUDE.md`, `.moai/memory/*`) → 개인 메모리(`~/.claude/CLAUDE.md`)
- `.claude/memory/`에는 요약과 링크만 유지해 초기 토큰 사용량을 최소화함.

## 핵심 참조 문서
- 프로젝트 허브: `@CLAUDE.md`
- 운영·협업 전문: `@.moai/memory/operations.md`
- 엔지니어링 표준: `@.moai/memory/engineering-standards.md`
- 공통 체크리스트: `@.claude/memory/shared_checklists.md`
- 헌법 및 거버넌스: `@.moai/memory/constitution.md`
- 스택별 가이드: `@.moai/memory/backend-*.md`, `@.moai/memory/frontend-*.md`

## 단계별 추천 임포트
- SPECIFY: `@.moai/memory/common.md`
- PLAN: `@.moai/memory/constitution.md`
- TASKS/IMPLEMENT: 기술 스택별 문서(`backend-*.md`, `frontend-*.md`)
- SYNC: 헌법/체크리스트 업데이트 여부 확인

## 운영 팁
1. 새 문서를 추가할 때는 `.moai/memory/`에 작성 후 이 README에 링크만 추가함.
2. 오래된 내용은 `.moai/memory/decisions/` 기록과 함께 정리함.
3. `.claude/hooks/moai/context_selector.py`가 Top‑K 추천(3~5개)을 출력하므로, 콘솔 안내를 참고해 필요한 문서를 임포트.

세부 규칙은 `@.moai/memory/constitution.md`와 기술 스택별 문서를 기준으로 유지·갱신함.
