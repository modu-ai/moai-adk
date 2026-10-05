# t839 Verdict — CI Race 잡 internal/gateway 600s 교착 수리

Card: t839 · Branch: `WT-gateway-test-hang` · Base: develop `9f5978d67` · Date: 2026-09-14

## Claim

1. CI Race 잡(run 34765755281, head `4da5d1c4e`)의 gateway 600s 타임아웃의 범인은 `TestOpenAISubprocessHTTPToolContinuation`이 아니라 **`TestOpenAISubscriptionUsesStoreAuthorizedBoundary`** 다. (리드 최초 가설 기각 — 리드 부관의 아티팩트 판독과 본 레인 판독 일치)
2. 기제: `-race`+부하에서 첫 egress 시도가 서버 측 처리 후 클라이언트 측 전송 오류로 끝나면 `upstreamSend`가 재시도하고, 재시작된 핸들러가 `seen`(cap 1) 채널 봉쇄 전송에서 영구 대기 → `calls.Load()` 팽창(3)으로 `logout crossed boundary` 오판 Fatal → cleanup의 `httptest.Server.Close()`가 봉쇄된 핸들러를 영구 대기 → 600s 패키지 타임아웃. 실제 로그아웃 경계 로직(401)은 올바르게 작동했다 — 고장은 테스트 장부다.
3. 봉쇄 전송 제거 + 재시작 안전 계수로 수리. 형제 2곳(`TestOpenAIAPIKeyEndpointAndPublicResponse`, `TestAnthropicNativeMessagesPreserveAndFilter`) 동일 패턴 동반 수리.

## Evidence

- **CI 판독**: `gh run view --job 103746357668 --log` → `FAILED PKG .../internal/gateway`, `600.095s`. 덤프는 잡 로그가 아닌 아티팩트에 존재: `test-stream-ci-race-ubuntu-latest`(artifact id 10320263298) → `panic: test timed out after 10m0s / running tests: TestOpenAISubscriptionUsesStoreAuthorizedBoundary (9m59s)`, 테스트 goroutine `FailNow`@`openai_test.go:248` → `httptest.(*Server).Close` `[sync.WaitGroup.Wait, 8 minutes]`, 핸들러 goroutine `[chan send, 9 minutes]`@`openai_test.go:227`. 타임라인: 15:33:00 시작 → 15:34:00 `logout crossed boundary` → 15:34:05 `httptest.Server blocked in Close` → 15:42:58 panic.
- **RED (결정론 재현)**: 스크래치 재현 테스트(운용의 실제 `upstreamSend` + 수리 전 봉쇄 핸들러 패턴, `ResponseHeaderTimeout` 50ms < 핸들러 지연 80ms로 재시작 강제) → 30s·90s 바운드 모두 패키지 타임아웃, CI와 동일 형태의 덤프(`httptest.Server blocked in Close` + `[chan send]`). 로그: `.moai/reports/t839/red-repro.log`
- **수리** (diff 요약, 3개 테스트):
  - 핸들러 채널 전송을 `select { case seen <- v: default: }` 로 변경 — 첫 포착 우선, 초과 드랍. cleanup 교착 원천 제거.
  - `io.WriteString` 실패 `t.Error` 제거(`_, _ =`) — 버려진 재시작 연결의 쓰기 실패는 정상.
  - 경계 테스트의 `calls.Load() != 1`(TLS 다이얼 계수 — 재시작에 의해 오염) → 서버 측 `served` 계수로 교체: 로그아웃 전 값 대비 로그아웃 후 불변 단언. 재시작 노이즈에 불변.
- **GREEN**: 수리 후 재현 테스트 0.19s 통과(봉쇄 소실), 대상 4테스트 `-race` 전부 PASS (`.moai/reports/t839/green-fix.log`), gateway 패키지 전체 plain 9.467s ok / `-race` 13.443s ok (진짜 python3 `/opt/homebrew/bin/python3` PATH 고정, `-count=1`).
- **Lint**: `golangci-lint run internal/gateway/...` → `0 issues.` · `gofmt -l` → 0건 · `go vet` → 통과.

## Baseline-attribution

- 모든 측정은 본 레인 워크트리 `.claude/worktrees/t839`(branch `WT-gateway-test-hang`, base develop `9f5978d67`)에서, 이번 세션에 실행한 명령과 출력으로 확정. CI 판독은 run 34765755281(head `4da5d1c4e`)의 잡 로그와 아티팩트 원문.
- 리드 전달 가설(python3 pyenv shim → CI 전용 JSON-RPC 진입)의 전반부는 로컬에서 확인(shim이면 즉시 실패)하나, 진짜 python3 단독 재현은 0.94s PASS — 행 가설로는 불성립. 아티팩트 덤프가 위 2번 기제로 대체 확정.

## Gaps

- CI에서 재시작을 유발한 정확한 전송 오류의 신원(WriteTimeout 1s 초과 vs ResponseHeaderTimeout 등)은 미확정 — `-race`+러너 부하의 타이밍 의존. 다만 "서버 처리 후 클라이언트 오류 → 재시작 → 핸들러 재진입"은 결정론 재현으로 증명.
- 로컬 `-race` 13.4s는 CI 러너 부하 프로필과 다름 — 원래의 간헐적 재시작 조건 자체는 로컬에서 자연 재현되지 않음(결정론 주입으로 대체).
- 다른 2 head(600.012/600.144s)의 아티팩트는 미판독 — 같은 가족 테스트 붉음은 리드 측 보고 인용.

## Residual-risk

- 수리는 테스트 장부만 건드린다(운용 코드 무변경) — `upstreamSend`의 재시작 설계와 `SendAuthorized`의 WriteTimeout은 유지. 재시작이 실제로 유발되는 러너 부하에서 여전히 첫 Send가 502로 귀결될 수 있으나, 이제 즉시 실패로 드러나고 cleanup 교착은 없다.
- `TestOpenAISubscriptionNonStreamCollectsSSE` 등 타 head 붉음 가족은 본 수리 후 로컬 `-race` PASS — CI 판정은 develop push 후 리드 판독 몫.

## CI 개선 제안 (리드 판정용 — 미적용)

1. **census에 범인 노출**: Race 잡은 `go test -json … > test-stream.json` 으로 덤프가 잡 로그에 안 남는다. `scripts/ci-census/test-census.sh`가 스트림에서 `panic: test timed out` 직후의 `running tests:` 블록을 FAILED PKG 행과 함께 출력하면 아티팩트 다운로드 없이 범인이 즉시 보인다. (이번 사건에서 덤프는 아티팩트에 존재 — 부재가 아니라 요약 누락이었다)
2. **패키지 타임아웃 단축**: `go test -timeout`을 600s 미만(예: 300s)으로 명시해 덤프 도달을 앞당기는 안. 다만 census 개선이 선행되면 급하지 않다.
