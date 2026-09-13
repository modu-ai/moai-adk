# t654 A5-M3 — 경로별 context 재판정·표시 증거 (AS-020 / AC-MG-010)

## Claim

provider별 capability 선언이 재판정됐다 — GLM은 text-only(`Images: false`)·명목 200K, Claude는
이미지 수용(`Images: true`)·명목 1M. 요청 경계에서 `!entry.Capabilities.Images` 경로의 이미지 입력은
명시 400으로 거절되고 upstream 요청 계수는 0이다. gpt 모드의 `CLAUDE_CODE_MAX_CONTEXT_TOKENS` env는
경로 유효 한도(선택 가능 경로 중 최솟값)를 싣는다 — 명목 창과 유효 한도는 다른 값으로 구분된다.

## Evidence

### 1. 서버 capability 게이트 — RED

명령: `go test ./internal/gateway/ -run 'TestGatewayContextPathCapabilityGate' -count=1`
기준 트리: `aec09a5d6` (게이트 구현 전). 출력:

```
    context_paths_test.go:38: image input on a text-only route answered 200, want explicit 400
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	0.570s
```

로그: `.moai/reports/t654/as5-m3-context-red.log`

### 2. cli capability 선언 — RED

명령: `go test ./internal/cli/ -run 'TestGatewayNativeModelCapabilitiesPerProvider' -count=1`
기준 트리: `aec09a5d6` (선언 변경 전). 출력:

```
    gateway_launch_assembly_test.go:141: glm row glm-tier-high declares image acceptance; GLM is text-only (AS-020)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.948s
```

로그: `.moai/reports/t654/as5-m3-capabilities-red.log`

### 3. GREEN

명령: `go test ./internal/gateway/ -count=1 -skip 'TestAppServerSubprocessHTTPToolContinuation'`
기준 트리: 커밋 예정 워킹 트리. 출력:

```
ok  	github.com/modu-ai/moai-adk/internal/gateway	6.914s
```

로그: `.moai/reports/t654/as5-m3-context-green.log`. skip한 2하위시험은 baseline에서도 실패하는
환경 의존 하위프로세스 시작 실패("app server start failed")로, baseline server.go 재실행으로
사전 존재를 확인했다(내 변경 무관).
cli 측: `go test ./internal/cli/ -run 'TestGatewayNativeModelCapabilitiesPerProvider' -count=1` →
`ok`, `go vet ./internal/gateway/ ./internal/cli/` → exit 0.

## Baseline-attribution

모든 출력은 2026-09-14, worktree `.claude/worktrees/t654`, base `6de8dd489`에서 직접 실행.

## Gaps

- GLM 200K·Claude 1M은 공식 문서의 **명목값**이다(`research.md` §19.3 취지). 실제 대형 입력·이미지의
  수용 보장 실측은 실계정 권한이 필요하므로 **Gap** — 다른 모델 숫자로 대체하지 않는다. AC-MG-010의
  "872k metadata도 실제 수용 시험과 별개로 표기" 원칙에 따라 catalog 값은 수용 보장으로 표시되지
  않는다(gatewayAuthDisplay/출력 정책 표시가 그 경계를 운반).
- 실계정 이미지 입력 양성(GLM 거절 관측·Claude 수용 관측)은 라이브 창 대기.

## Residual-risk

- 이미지 탐지는 Messages 형식의 `messages[].content[].type == "image"` 블록 스캔이다. 미래 형식
  변화(새 이미지 운반 필드)가 생기면 게이트가 빗나갈 수 있다 — 음성 변형이 픽스처로 고정돼 있어
  회귀 시 적색으로 드러난다.
