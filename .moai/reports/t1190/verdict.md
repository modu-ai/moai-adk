# t1190 — Claude Code 명령 버전 표 정정

## Claim

- `docs-site/content/{en,ko,ja,zh}/claude-code/foundations/commands.md`의 `/workflows` 버전은 `v2.1.154+`로 정정했다.
- `/background` (`/bg`)는 `v2.1.141`에 존재했다는 근거만 확인돼, 표에서 도입 버전 미확인으로 표시했다. 기존 `v2.1.139+` 표기는 입증하지 못했다.

## Evidence

- 공식 저장소의 태그별 CHANGELOG 원문: `https://raw.githubusercontent.com/anthropics/claude-code/v2.1.154/CHANGELOG.md`의 `## 2.1.154` 아래에 `Introducing dynamic workflows`와 `Run /workflows to view your runs`가 함께 나온다. `v2.1.153` 변경 기록에는 `/workflows` 항목이 없다.
- `https://raw.githubusercontent.com/anthropics/claude-code/v2.1.141/CHANGELOG.md`의 `## 2.1.141` 아래에는 `/bg`가 현재 권한 모드를 보존하도록 고쳤다는 항목이 나온다. `v2.1.139` 변경 기록에는 agent view 도입만 있고 `/background` 또는 `/bg` 명령 도입 항목은 없다.
- `git -C .claude/worktrees/t1190 diff --check` → 출력 없음, 종료 0.
- `hugo --quiet --destination /tmp/moai-t1190-hugo 2>/tmp/moai-t1190-hugo.stderr` → `hugo_exit=0`, `warn_count=0`.
- 빌드 산출물의 `en`, `ko`, `ja`, `zh` 네 페이지 모두 존재했고 `v2.1.154+`와 각 언어의 도입 버전 미확인 문구가 확인됐다.

## Baseline-attribution

- 이 턴에 `origin/develop`를 fetch하고, MoAI가 생성한 `.claude/worktrees/t1190`에서 `origin/develop`를 병합한 뒤 측정했다. 시작 HEAD는 `302cd9e04`였다.
- 비교한 공식 변경 기록은 이 턴에 받은 `v2.1.139`, `v2.1.141`, `v2.1.153`, `v2.1.154` 태그의 원문이다.

## Gaps

- 공식 CHANGELOG에는 `/background`와 `/bg`의 도입 버전이 명시되지 않았다. `v2.1.139`에 기능이 없었다고 단정할 근거도 없다.
- 과거 Claude Code 바이너리를 실행해 명령을 재현하지 않았다.

## Residual-risk

- 릴리스 기록에 빠진 명령 도입이 더 이른 버전에 있을 수 있다. 그래서 최소 버전 대신 확인 가능한 상한과 미확인 상태를 표시했다.
