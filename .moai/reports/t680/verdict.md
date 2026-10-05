# t680 Verdict — stop-goal stdout 오염 (ISSUE #1703)

Date: 2026-09-13 · Lane: lane-1 · Branch: `WT-stop-goal-stdout` (develop `74d872aaf` 기점) · Tier S, Class B (run→sync)

## Claim

ISSUE #1703(stop-goal Stop 훅 stdout에 판정 JSON 뒤 CLI 헬프 텍스트 부착 → Claude Code 파싱 실패, block 결정 유실)의 원인을 확정하고, 카드 [HARD] 회귀(모든 종료 경로에서 stdout이 정확히 JSON 객체 하나)를 테스트로 확보한다.

## Root cause (확정)

카드의 초기 가설(SilenceUsage 부재 / RunE가 오류 반환)은 **반증**됐다 — `hook_stop_goal.go`는 v3.1.2와 develop 양쪽 모두 `SilenceUsage: true`이며 RunE는 모든 경로에서 nil을 반환한다.

실제 원인은 **2중 결함의 합성**이다:

1. **래퍼 double-fire (t220)**: v3.1.2 배포본 `handle-stop-goal.sh`는 티어별 `printf '%s' "$INPUT" | exec <bin> hook stop-goal` 형태. 파이프라인 우변의 `exec`는 서브셸만 교체하므로 래퍼가 다음 티어로 추락해 평가자를 티어 수만큼 발사한다.
2. **구형 바이너리의 부모-헬프 출력**: 티어2(`$HOME/go/bin/moai`)가 `stop-goal` 서브커맨드가 없는 오래된 바이너리면, cobra는 non-runnable 부모 `hook` 커맨드의 **헬프 전문을 stdout에** 출력하고 exit 0로 끝난다.

합성 결과: 티어1(신형, 판정 JSON) + 티어2(구형, 헬프)가 한 stdout에 결합 → `Extra data: line 2 column 1`.

수리 `c4e90cd58`(t220, #1629, 2026-08-24)은 **v3.1.2(4b2f203fe)에 미포함** — develop에만 존재. 즉 본 이슈는 "이미 수리된 결함의 스테일 배포 manifestation"이며, v3.1.2 사용자에게는 다음 릴리스가 유일한 전달 경로다.

## Evidence (본 run의 실측)

| # | Command | Observed |
|---|---------|----------|
| E1 | `/tmp/moai-t680`(develop 빌드) + armed goal, block 경로 | stdout 95바이트 단일 JSON `{"decision":"block",...}`, exit 0 |
| E2 | 동일, ceiling 경로(turns_used=30/30) | stdout 655바이트 단일 JSON `{"ceiling_exit":true,...}`, exit 0 |
| E3 | 설치 바이너리 v3.2.0-rc.10(`8050369e5`), block 경로 | stdout 95바이트 단일 JSON — 오염 없음 |
| E4 | `moai hook stop-goalx`(존재하지 않는 서브커맨드) | **stdout 3,777바이트 헬프 전문**(`hookCmd.Long` + `Usage: moai hook [command]`), stderr 0바이트, exit 0 — 오염 기제의 경험적 확인, 이슈 기술(~4,100자)과 동일 형태 |
| E5 | `git merge-base --is-ancestor` 3회 | 9ad54f5c1(goal engine) ∈ v3.1.2 · c4e90cd58(수리) ∉ v3.1.2 · 9ad54f5c1 ∈ 4b2f203fe |
| E6 | 신규 회귀 `TestStopGoalStdoutSingleJSONObjectOnEveryExitPath` | 5/5 PASS (block·ceiling_exit·allow·load_error·no_goal; stdout/stderr 분리 캡처) |
| E7 | 변이 프로브(JSON 뒤 헬프 한 줄 삽입) | block·ceiling 서브테스트 RED — "carries data after the first JSON object"; 원복 후 zero diff, GREEN 재확인 |
| E8 | `gofmt -l`(빈 출력) + `go vet ./internal/cli/`(무소식) + 선택자 실행 `TestStopGoalStdout*|TestGoalArm|TestStopGoalWrapper` → ok | 포맷·정적·관련 회귀 전부 통과 |

## 형제 훑기 (카드 [HARD])

- develop 래퍼 전수: 살아있는 `printf | exec` 파이프라인 형태는 0건(매치는 전부 t220 설명 주석).
- 형제 래퍼(handle-stop.sh 등)는 bare `exec <bin> hook <event>` 형태 — `exec`가 래퍼 프로세스 자체를 교체하므로 단일 발사가 구조적으로 보장되고, 티어 추락 경로가 없다.
- `moai hook <미지정 서브커맨드>`가 stdout에 헬프를 출력하는 cobra 동작 자체는 모든 버전에 존재한다. 단일 발사가 보장된 후에는 오염이 발생하려면 **첫 티어 바이너리 자체가 구형**이어야 하는데, 이는 래퍼가 고칠 수 없는 축(호출 대상이 곧 결함)이다 — 별도 대응 불필요로 판정.

## Changed files

- `internal/cli/hook_stop_goal_test.go` (신규) — 5경로 stdout 단일-JSON 회귀 + 분리 스트림 드라이버
- `.moai/reports/t680/verdict.md` (본 문서)

`internal/cli/hook_stop_goal.go`는 **무변경**(원인이 본 트리의 코드가 아니므로 — 변이 프로브는 원복 완료, E7).

## Gaps

- `internal/cli` 패키지 전량 스위트는 로컬 미실행(§4/CLAUDE.local.md — CI가 전체 판정). 선택자 범위(TestStopGoal*·TestGoalArm*)만 로컬 측정.
- E4의 unknown-subcommand 관측은 rc.10 바이너리 기준이며, fang 레이어의 모든 stdout 라우팅 경로를 망라하지 않는다.
- 이슈 사용자(lotto_mini) 환경의 티어2 바이너리 실체(구형 `~/go/bin/moai`)는 이슈 서술(JSON 1,386자 + 헬프 ~4,100자)로부터의 정합 추정이며 사용자 머신 미실측.

## Residual risk

- v3.1.2 사용자는 다음 릴리스(v3.2.0 계열) 배포 전까지 동일 증상에 계속 노출된다 — 본 카드의 트리 수리(develop)만으로는 전달되지 않는다. 릴리스가 실제 전달 행위다.
- 티어1이 구형 바이너리인 환경(래퍼와 무관하게 PATH가 오래된 moai를 가리킴)에서는 헬프-only stdout이 여전히 가능 — 위 "형제 훑기" 판정대로 래퍼 수리 밖의 축.

🗿 MoAI
