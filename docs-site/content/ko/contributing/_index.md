---
title: 참여하기
weight: 110
draft: false
---

MoAI-ADK는 오픈소스 프로젝트이며 기여를 환영합니다! 이 가이드는 프로젝트에 기여하는 방법을 설명합니다.

## 빠른 시작

1. 저장소를 **Fork** 합니다
2. 기능 브랜치 생성: `git checkout -b feature/my-feature`
3. 테스트 작성 (새 코드는 TDD, 기존 코드는 특성화 테스트)
4. 모든 테스트 통과 확인: `make test`
5. 린팅 통과 확인: `make lint`
6. 코드 포맷: `make fmt`
7. Conventional Commit 메시지로 커밋
8. Pull Request 생성

## 코드 품질 요구사항

| 항목 | 기준 |
|------|------|
| 테스트 커버리지 | **85%** 이상 |
| 린트 에러 | **0**개 |
| 타입 에러 | **0**개 |
| 커밋 메시지 | Conventional Commits 형식 |

## 커밋 메시지 형식

```
<type>(<scope>): <description>

[선택적 본문]

[선택적 푸터]
```

### 타입

| 타입 | 설명 |
|------|------|
| `feat` | 새로운 기능 |
| `fix` | 버그 수정 |
| `docs` | 문서 변경 |
| `style` | 코드 포맷 (기능 변경 없음) |
| `refactor` | 리팩토링 (기능 변경 없음) |
| `perf` | 성능 개선 |
| `test` | 테스트 추가/수정 |
| `chore` | 빌드/도구 변경 |
| `revert` | 이전 커밋 되돌리기 |

### 예시

```
feat(template): add SessionEnd hook to settings.json generator
fix(cli): prevent race condition in hook execution
test(settings): add TestEnsureGlobalSettingsEnv test cases
docs(readme): update agent count and statistics
```

## 개발 환경 설정

### 필수 도구

- **Go 1.26+** — 핵심 개발 언어
- **Git** — 버전 관리
- **make** — 빌드 명령

### 주요 명령어

```bash
make build        # 프로젝트 빌드
make test         # 테스트 실행
make test-race    # Race condition 감지 테스트
make lint         # 린터 실행
make fmt          # 코드 포맷
make install      # 로컬 설치
make clean        # 빌드 산출물 정리
```

## Pull Request 가이드

### PR 작성 시

- 명확하고 간결한 제목 (70자 이내)
- 변경 내용 요약 (Summary 섹션)
- 테스트 계획 (Test Plan 섹션)
- 관련 이슈 참조 (예: `Fixes #123`)

### PR 체크리스트

- [ ] 테스트 추가/업데이트
- [ ] 모든 테스트 통과 (`make test`)
- [ ] 린팅 통과 (`make lint`)
- [ ] 커밋 메시지 Conventional Commits 형식
- [ ] 문서 업데이트 (필요 시)

## 커뮤니티

- **이슈 트래커**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — 버그 리포트, 기능 요청
- **공식 문서**: [adk.mo.ai.kr](https://adk.mo.ai.kr)

## 라이선스

[Apache License 2.0](https://github.com/modu-ai/moai-adk/blob/main/LICENSE) — 자유롭게 사용, 수정, 배포할 수 있습니다.
