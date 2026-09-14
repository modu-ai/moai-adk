# t654 A5-M1 — launcher 생산 통합 증거 (AC-MG-026 (a))

## Claim

세 launcher의 launch 조립이 provider 전용 catalog·picker 구성·인증 방식 표시·App Server transport를
하나의 조립 경로에서 전달하며, 조립이 provider별 auth 방식 표시(`authMethod`)와 App Server 출력 정책
표시(`outputPolicy: app-server`)를 overlay로 넘긴다. 대기 오류 게이트 두 곳(`gpt.go:54`,
`launcher.go:142`)은 검증 게이트 통과 전 트리에서 대조군으로 유지된다.

## Evidence

### 1. 조립 auth 표시 — RED (구현 전)

명령: `go test ./internal/cli/ -run 'TestGatewayLaunchAssemblyCarriesAuthDisplay' -count=1`
기준 트리: `6de8dd489` (변경 전). 출력 발췌:

```
--- FAIL: TestGatewayLaunchAssemblyCarriesAuthDisplay (0.00s)
    --- FAIL: TestGatewayLaunchAssemblyCarriesAuthDisplay/gpt (0.00s)
        gateway_launch_assembly_test.go:87: authMethod display = <nil>, want "subscription" ...
    --- FAIL: TestGatewayLaunchAssemblyCarriesAuthDisplay/claude (0.00s)
        gateway_launch_assembly_test.go:87: authMethod display = <nil>, want "subscription" ...
    --- FAIL: TestGatewayLaunchAssemblyCarriesAuthDisplay/glm (0.00s)
        gateway_launch_assembly_test.go:87: authMethod display = <nil>, want "existing-credential" ...
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.578s
```

전체 로그: `.moai/reports/t654/as5-m1-auth-display-red.log`

### 2. 조립 auth 표시 — GREEN (구현 후)

명령: `go test ./internal/cli/ -run 'TestGatewayLaunchAssemblyCarriesAuthDisplay|TestGatewayLaunchTransportGateControl' -count=1`
기준 트리: `6de8dd489` 워킹 트리 (구현 후, 커밋 `5672f5029` 내용). 출력:

```
ok  	github.com/modu-ai/moai-adk/internal/cli	1.160s
```

### 3. 대기 오류 게이트 — 대조군 RED 증거 (기계 판정 명령)

명령: `grep -rn 'awaiting transport verification' internal/cli --include='*.go' | grep -v _test`
기준 트리: `6de8dd489`. 출력 (exit 0 — 리터럴 2건 존재 = 게이트 폐쇄 대조군):

```
internal/cli/launcher.go:142:		return errors.New("GPT gateway launch is awaiting transport verification; use moai gpt status to inspect login")
internal/cli/gpt.go:54:			return errors.New("GPT gateway launch is awaiting transport verification; use moai gpt status to inspect login")
```

로그: `.moai/reports/t654/as5-m1-launch-red.log`. 같은 상태를 소스 스캔 시험
`TestGatewayLaunchTransportGateControl`이 고정한다(파일당 리터럴 1건 + nil-binding gpt launch가
대기 오류로 끝남 + nil-Launch `moai gpt`가 대기 오류로 끝남). 이 시험은 GREEN(대조군 계약)이며,
AS-014~AS-022가 전수 PASS한 뒤 리터럴 제거와 함께 gate-open 형태로 뒤집히는 것이 (a)의 GREEN 행위다.

### 4. 회귀

명령: `go test ./internal/cli/ -count=1 -timeout 40m` — 배경 실행(부하 시 패키지 전량 수십 분).
명령: `go vet ./internal/cli/` → exit 0. `golangci-lint run internal/cli/` → `0 issues.`

## Baseline-attribution

위 모든 출력은 2026-09-14, worktree `.claude/worktrees/t654`, branch `WT-gateway-launchers`,
base `6de8dd489`에서 본 세션이 직접 실행해 관측했다. 구현 커밋: `5672f5029`.

## Gaps

- AC-MG-026 (a)의 GREEN(대기 오류 리터럴 0건 + 실제 launch 진행)은 AS-014~AS-022 전수 PASS를
  Given으로 하므로 이 창에서 판정하지 않는다 — 창 대기. 실계정 실증(AS-017·019·021)이 t851 이후
  라이브 창에서 열린다.
- `moai gpt` 실제 PTY 실행(대기 오류 미발생 대조군 실행)은 라이브 창 대기다.

## Residual-risk

- 게이트 개방 시 `TestGatewayLaunchTransportGateControl`의 소스 스캔 단언(파일당 1건)이 함께
  뒤집혀야 한다 — 개방 커밋이 이 시험을 삭제·수정하지 않고 리터럴만 지우면 적색으로 낙오되어
  무단 개방을 막는다(의도된 안전망).
- overlay의 `authMethod`/`outputPolicy` 키는 Claude Code가 소비하지 않는 표시 표면이다. 사용자
  가시 렌더링(클라이언트 UI 표시)은 AS-014/AS-021 실계정 판정에서 확인한다.
