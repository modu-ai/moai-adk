# Git 워크플로우(요약)

브랜치 전략, 릴리스 절차, pre-commit 구성 등 상세 지침은 `@.moai/memory/operations.md` 3·4·5장에 정리되어 있습니다. 요약을 빠르게 상기하기 위한 문서입니다.

## 핵심 요약
- 브랜치: `main`(배포), `develop`(통합), `feature/*`, `hotfix/*`, `release/*`; 보호 브랜치는 force push 금지.
- 커밋: Conventional Commit(`type(scope): message`), 50자 이하, 명령형.
- PR: 공통 체크리스트로 테스트/보안/문서 검증, 리뷰어 1명 이상, CI 통과 필수.
- 릴리스: Feature Freeze → release 브랜치 → QA → main 병합 → 태깅 → `/moai:6-sync`.

### 참조
- 운영 메모 전문: `@.moai/memory/operations.md`
- 엔지니어링 표준: `@.moai/memory/engineering-standards.md`
