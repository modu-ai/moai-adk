# Gateway fallback 설정·인수 방어 검증

## Claim

REQ-MG-022의 세션 설정·명시 인수 방어를 구현했다. `newGatewaySessionBinding`은 호출자 overlay를 복사한 뒤 `fallbackModel: []`를 설정한다. 전달된 `--settings`의 fallback 설정보다 이 세션 값이 우선한다. 호출자 map과 availableModels는 변경하지 않는다. `prepareGatewayOverlay`는 명시 `--fallback-model` 및 `--fallback-model=...`를 자식 시작 전에 오류로 거절한다. `--` 이후 인수와 JSON 설정 문자열은 옵션으로 오인하지 않는다.

## Evidence

RED — 구현 변경 전에 신규 시험을 실행했다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-picker-guard-cache go test ./internal/cli -run 'TestGatewaySessionDisablesInheritedFallback|TestGatewaySessionRejectsExplicitFallback|TestGatewayOverlayPreservesFallback' -count=1 -timeout 60s
```

```text
--- FAIL: TestGatewaySessionDisablesInheritedFallbackWithoutMutatingCaller (0.00s)
    gateway_session_test.go:177: fallbackModel = ["claude-opus-5"], want []
    gateway_session_test.go:177: fallbackModel = ["claude-opus-5"], want []
--- FAIL: TestGatewaySessionRejectsExplicitFallbackBeforeChildStart (0.00s)
    gateway_session_test.go:202: args=[--fallback-model claude-sonnet-5] err=unexpected start starts=1
    gateway_session_test.go:202: args=[--fallback-model=claude-opus-5] err=unexpected start starts=1
    gateway_session_test.go:202: args=[--fallback-model] err=unexpected start starts=1
    gateway_session_test.go:202: args=[--settings {} --fallback-model=] err=unexpected start starts=1
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.239s
FAIL
```

GREEN — settings/session 관련 기존 시험과 신규 시험을 실행했다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-picker-guard-cache go test ./internal/cli -run 'TestGatewaySession|TestGatewayOverlay|TestGatewaySettings' -count=1 -timeout 60s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	1.646s
```

커버리지 측정:

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/cli -run 'TestGatewaySession|TestGatewayOverlay|TestGatewaySettings' -count=1 -timeout 60s -coverprofile=/tmp/gateway-fallback-guard.cover && go tool cover -func=/tmp/gateway-fallback-guard.cover | rg 'gateway_settings.go|gateway_session.go'
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	1.252s	coverage: 6.6% of statements
github.com/modu-ai/moai-adk/internal/cli/gateway_session.go:31:			newGatewaySessionBinding		100.0%
github.com/modu-ai/moai-adk/internal/cli/gateway_session.go:94:			replaceGatewaySettingsArgs		100.0%
github.com/modu-ai/moai-adk/internal/cli/gateway_settings.go:17:		cleanupGatewaySettings			88.9%
github.com/modu-ai/moai-adk/internal/cli/gateway_settings.go:39:		prepareGatewayOverlay			95.6%
```

6.6%는 선택한 시험이 CLI 전체 패키지에서 실행한 문장의 비율이다. 위 네 함수의 개별 측정과 구별한다.

LSP 구현 전·후 각각 같은 명령을 실행했고 모두 exit 0, stdout/stderr 출력 없음이었다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache gopls check internal/cli/gateway_settings.go internal/cli/gateway_session.go
```

## Baseline-attribution

2026-09-11, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, `WT-unified-gateway`, HEAD `81c1d58f9`의 공유 변경 상태에서 실행했다. 편집 전 fetch 후 `origin/main...HEAD`는 `0 2879`였다. 이번 소유 변경은 gateway_settings.go, gateway_session.go, 두 대응 시험 파일 및 이 보고서뿐이다. 이전 실제 클라이언트 관측은 [fallback-runtime-observation.md](fallback-runtime-observation.md)를 읽어 근거로 삼았으나 이번 작업에서 재실행하지 않았다. 작은 기존 순회에 조건과 세션 값을 추가하여 별도 추상화는 만들지 않았다.

## Gaps

- 실제 Claude 클라이언트·공급자·계정·TUI 시험을 이번 변경으로 실행하지 않았다. 이번 PASS는 Go 세션 설정·인수 조립과 자식 시작 전 거절 범위다.
- 기존 실제 관측의 빈 fallback overlay 사례는 timeout이므로 무기한 fallback 부재나 정상 오류 종료를 입증하지 않는다. 초기 availableModels 제외 시 Opus로 바뀐 관측은 별도 계약 검토 대상이다.
- 실제 클라이언트의 동일 공급자 재시도, 401/429/503/529별 최종 사용자 오류, 재개 세션의 정책은 이번 시험으로 증명하지 않는다.
- 제품 활성화·OAuth·모델 한도·availableModels는 변경하지 않았다. 전체 REQ-MG-022 완료 판정이 아니다.
- 저장소 전체 시험 판정은 통합 브랜치 GitHub CI 소관이며 현재 PENDING이다. 이번 작업에서 CI run을 생성하거나 관측하지 않았다. commit/push/PR은 수행하지 않았다.

## Residual-risk

클라이언트 버전과 설정 출처 우선순위에 따라 실제 fallback 억제 결과가 달라질 수 있다. 설정 배열과 명시 CLI 옵션 방어의 로컬 계약은 통과했으나 실제 런타임 재시도·초기 모델 대체와 전체 요청 경로 검증이 남는다.
