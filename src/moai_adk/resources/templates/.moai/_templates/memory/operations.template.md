# $PROJECT_NAME 운영 메모

> 프로젝트 헌법(Constitution)과 실무 규칙을 모아둔 운영 지침입니다. 세부 제약은 `.claude/memory` 요약본이 아닌 이 문서를 참고하세요.

## 1. 프로젝트 핵심 원칙
- Spec-First + TDD: 요구사항(@REQ) → 계획(@DESIGN) → 작업(@TASK) → 구현(@FEATURE/@TEST) 순서를 지키고 `/moai:2-spec`~`/moai:5-dev` 커맨드를 활용함.
- Small & Safe 루프: “문제 정의 → 작고 안전한 변경 → 리뷰 → 리팩터링”을 반복하고 영향도·가정·대안을 Issue/PR/ADR에 기록함.
- 헌법 준수: `@.moai/memory/constitution.md`의 5대 원칙(단순성/아키텍처/테스트/관찰가능성/버전 관리)을 매 변경마다 확인함.

## 2. 협업 & 커뮤니케이션
- 회의: 아젠다 24시간 전 공유 → 정시 시작/종료 → 액션 아이템(담당·기한) 기록 → 당일 회의록 공유.
- 슬랙/디스코드 채널: `#general`, `#dev-discuss`, `#code-review`, `#deployment`, `#monitoring`, `#random` 등으로 역할을 분리함.
- 메시지 규칙: 핵심 요약 + 🔴/🟡/🟢 긴급도, 관련 태그(@REQ/@BUG 등) 명시, 스레드로 후속 논의를 묶음.
- 언어: 모든 대화/문서는 한국어, 외부 공유 시 번역본을 추가.

## 3. Git 워크플로우 요약
- 브랜치: `main`(배포), `develop`(통합), `feature/*`, `hotfix/*`, `release/*`; 보호 브랜치는 force push 금지.
- 커밋: Conventional Commit(`type(scope): message`), 50자 이하, 명령형, 마침표 금지. 자세한 규칙은 `@.moai/memory/engineering-standards.md`의 “공통 규칙” 절을 참고함.
- PR: 본 문서의 공통 체크리스트로 테스트/보안/문서 영향 확인, 리뷰어 최소 1명, CI 통과 필수.
- 릴리스: Git Flow 기반. Feature Freeze → release 브랜치 → QA → main 병합 → 태깅 → `/moai:6-sync`로 문서 동기화.

## 4. 공통 체크리스트
- [ ] 변경 전 관련 파일 전체 읽기 + 정의·참조·테스트·문서를 전역 검색함.
- [ ] 가정과 최소 두 가지 대안(장점/단점/위험) 정리.
- [ ] 테스트: 새 코드에는 새 테스트, 버그 수정은 회귀 테스트(먼저 실패) 작성, 커버리지 ≥ 80%.
- [ ] 보안: 입력 검증·인코딩·파라미터 바인딩, 민감 정보 로그 금지, 최소 권한.
- [ ] 문서 동기화: CLAUDE.md, `.moai/memory` 문서 갱신 후 `/moai:6-sync` 실행.

## 5. CLI & 자동화 요약
- 검색/목록: `rg --files`, `rg -n`, `fd -e`.
- 테스트/품질: Python → `pytest --cov`, JS → `pnpm test`, 공통 → `pre-commit run --all-files`.
- 컨테이너: `docker build -t $PROJECT_NAME:latest .`, `docker compose up -d`, `docker logs -f`.
- 배포: SSH(`ssh user@host`), systemd(`sudo systemctl restart service`), 쿠버네티스(`kubectl rollout undo`).
- 데이터: PostgreSQL `psql -U user -d db`, MySQL `mysql -u user -p`, Redis `redis-cli --scan --pattern "*"`.

## 6. 참고 문서
- 프로젝트 헌법: `@.moai/memory/constitution.md`
- 공통 체크리스트(원문): `@.claude/memory/shared_checklists.md`
- 보안 규칙: `@.claude/memory/security_rules.md`
- TDD 가이드: `@.claude/memory/tdd_guidelines.md`
- Git 상세 절차, 협업 규범, CLI 레퍼런스: 축약된 `.claude/memory` 문서 대신 이 파일을 활용함.
