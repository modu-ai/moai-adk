---
title: 참여하기
weight: 110
draft: false
---

MoAI-ADK는 오픈소스 프로젝트로, 누구나 자유롭게 읽고, 쓰고, 고칠 수 있습니다.
작은 오타 수정부터 새로운 기능 제안까지 모든 형태의 기여를 환영합니다. 이
문서는 저장소를 처음 열어보는 분이 첫 번째 Pull Request(코드 변경 제안)까지
도달하는 길을 안내합니다.

MoAI-ADK 자체도 SPEC(기능 단위의 작업 명세서) 기반 3-phase 워크플로우와
TRUST 5(다섯 가지 품질 기준) 게이트를 거쳐 만들어집니다. 기여자가 제출하는
코드에도 같은 기준을 적용합니다. 이 기준은 고품질을 유지하기 위한 울타리이지,
진입 장벽이 아닙니다. 어려운 부분은 리뷰 과정에서 함께 다듬어 갑니다.


## 기여 흐름 한눈에 보기

기여는 다섯 단계의 흐름으로 요약됩니다. 각 단계가 왜 필요한지 짚고 나면 아래
절차가 자연스럽게 이어집니다.

1. **탐색** — 저장소 구조와 기존 이슈를 살펴둡니다.
2. **분기** — 작업용 브랜치를 만들어 메인 코드와 격리합니다.
3. **구현** — 코드를 고치고 테스트로 검증합니다.
4. **제출** — Conventional Commits(약속된 커밋 메시지 형식) 규칙에 맞춰 PR을 엽니다.
5. **리뷰** — 피드백을 반영하고 main 브랜치에 반영될 때까지 다듬습니다.

{{< icon bulb muted >}} 처음 기여하시는 분은 "good first issue" 라벨이 붙은
이슈부터 시작하면 좋습니다. 범위가 작게 정해져 있어 진입이 쉽습니다.


## 빠른 시작

아래 여덟 단계를 순서대로 따라가면 로컬 환경에서 코드를 고쳐 PR까지 만들 수
있습니다.

1. 저장소를 **Fork** (내 GitHub 계정으로 복사)합니다.
2. 기능 브랜치를 생성합니다: `git checkout -b feature/my-feature`
3. 테스트를 먼저 작성합니다. 새 코드는 TDD(테스트 주도 개발), 기존 코드는
   특성화 테스트(현재 동작을 고정하는 회귀 테스트)로 보호합니다.
4. 모든 테스트가 통과하는지 확인합니다: `make test`
5. 린트(코드 스타일 검사)가 통과하는지 확인합니다: `make lint`
6. 코드 포맷을 맞춥니다: `make fmt`
7. Conventional Commits 형식으로 커밋 메시지를 작성합니다.
8. Pull Request를 생성합니다.

PR 제목은 70자 이내로 짧고 분명하게 적습니다. 본문에는 변경 요약(Summary),
테스트 계획(Test Plan), 관련 이슈 참조(예: `Fixes #123`)를 포함합니다.


## 코드 품질 요구사항

TRUST 5 프레임워크의 **T**ested (검증됨) / **T**rackable (추적 가능) 기준이
그대로 적용됩니다. PR이 병합되려면 아래 표의 기준을 모두 충족해야 합니다.

| 항목 | 기준 |
|------|------|
| 테스트 커버리지 | **85%** 이상 |
| 린트 에러 | **0**개 |
| 타입 에러 | **0**개 |
| 커밋 메시지 | Conventional Commits 형식 |


## 커밋 메시지 형식

Conventional Commits는 커밋 제목 한 줄만 봐도 어떤 종류의 변경인지 알 수 있게
하는 약속입니다. 자동 CHANGELOG 생성과 버전 관리에 그대로 사용됩니다.

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
| `style` | 코드 포맷(기능 변경 없음) |
| `refactor` | 리팩토링(기능 변경 없음) |
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
- **make** — 빌드 명령 래퍼

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

PR은 기여자와 메인테이너가 코드를 함께 검토하는 자리입니다. 아래 두 가지를
염두에 두면 리뷰가 순조롭게 흘러갑니다.

### PR 작성 시

- 제목은 70자 이내로 짧고 분명하게 적습니다.
- 변경 내용을 요약하는 Summary 섹션을 넣습니다.
- 어떻게 검증했는지 보여주는 Test Plan 섹션을 넣습니다.
- 관련 이슈를 참조합니다 (예: `Fixes #123`).

### PR 체크리스트

- [ ] 테스트를 추가하거나 업데이트했습니다.
- [ ] 모든 테스트가 통과합니다 (`make test`).
- [ ] 린팅을 통과했습니다 (`make lint`).
- [ ] 커밋 메시지가 Conventional Commits 형식을 따릅니다.
- [ ] 필요한 경우 문서를 업데이트했습니다.


## 세션 안에서 바로 이슈 만들기

MoAI-ADK를 사용하다가 버그를 발견하거나 기능을 제안하고 싶을 때, GitHub 웹으로
나가지 않아도 됩니다. 세션 안에서 `/moai feedback`을 입력하면 현재 대화 맥락을
반영한 이슈 초안이 자동으로 작성되어 저장소 이슈 트래커로 제출됩니다. 버그
재현 스텝과 환경 정보까지 함께 전달되므로, 별도 부연 설명이 줄어듭니다.


## 커뮤니티

- **이슈 트래커**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — 버그 리포트와 기능 요청을 받습니다.
- **Discord**: [Discord 커뮤니티](https://discord.gg/Z7E7Mdc5aN) — 실시간 소통과 사용 팁 공유.
- **공식 문서**: [adk.mo.ai.kr](https://adk.mo.ai.kr) — 안내서와 참조 문서.


## 라이선스

[Apache License 2.0](https://github.com/modu-ai/moai-adk/blob/main/LICENSE) —
자유롭게 사용, 수정, 배포할 수 있습니다. 기여해 주신 코드도 동일한 라이선스로
배포됩니다.
