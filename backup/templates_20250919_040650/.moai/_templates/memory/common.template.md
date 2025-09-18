# $PROJECT_NAME 프로젝트 공통 운영 메모

> 팀 전체가 공유하는 핵심 원칙과 체크 항목을 요약합니다. 상세 규정은 CLAUDE 메모리(@.claude/memory/*)를 참조하세요.

## 1. 기본 원칙
- 3단계 응답 구조 준수: 탐색 → 계획 → 구현 (@.claude/memory/three_phase_process.md)
- Small & Safe 변경 루프 반복: 문제 정의 → 작고 안전한 변경 → 리뷰 → 리팩터링 (@.claude/memory/project_guidelines.md)
- 모든 작업은 한국어로 커뮤니케이션하며, 가정·대안 비교(최소 2가지)를 Issue/PR/ADR에 기록
- 테스트/TDD/보안/PR 규칙은 공통 체크리스트(@.claude/memory/shared_checklists.md)로 관리

## 2. 기본 체크리스트
- [ ] 변경 전 관련 파일을 처음부터 끝까지 읽고 정의·참조·테스트·문서를 전역 검색
- [ ] 공통 체크리스트(@.claude/memory/shared_checklists.md) 항목을 모두 통과
- [ ] TDD 사이클(Red → Green → Refactor)로 구현하고 커버리지 ≥ 80% 유지 (@.claude/memory/tdd_guidelines.md)
- [ ] 보안 규칙(ISMS-P) 준수: 입력 검증, 시크릿 미노출, 최소 권한 (@.claude/memory/security_rules.md)
- [ ] 변경 사항을 CLAUDE.md(프로젝트 메모리)와 문서에 반영하고 `/moai:6-sync`로 동기화

## 3. 운영 가이드
- Git 브랜치/PR 규칙: @.claude/memory/git_workflow.md, @.claude/memory/git_commit_rules.md
- 협업 규약 & 회의/문서 템플릿: @.claude/memory/team_conventions.md
- 마스터 원칙 요약(Refactoring/Clean Code/TDD/API 패턴): @.claude/memory/software_principles.md
- Bash/CLI 도구 모음: @.claude/memory/bash_commands.md

## 4. 유지보수 메모
- 이 문서는 `moai init` 시 자동 생성되며 팀 합의에 따라 업데이트합니다.
- 언어/프레임워크별 메모(`frontend-*.md`, `backend-*.md`)와 함께 사용하고, 필요 없는 문서는 삭제 가능합니다.
- 새 템플릿을 추가할 때는 `.moai/_templates/memory/` 구조와 README 지침을 따르세요.
