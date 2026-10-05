# t697 — GPT 게이트웨이 상류 5xx 탄성 — 판정서

- 카드: t697 (Class B, run → sync, plan 생략)
- 브랜치: `WT-502-prestream-retry` @ `1ea7b1127` (기점: 로컬 develop `44e56d017`)
- 재측정 범위: `internal/gateway` 패키지

## Claim

1. 스텁 상류 5xx 1회 후 200 → 게이트웨이가 재시도로 성공한다 (RED→GREEN 확인).
2. 상류 연결 실패 1회 후 복구 → 동일하게 성공한다.
3. 연속 5xx / 연속 연결 실패 → 시도 상한(3회) 소진 후, 생성지와 사유가 본문에 라벨된 502로 실패한다.
   - 게이트웨이 생성: `gateway upstream connection failed after N attempts: <원인>`
   - 상류 응답: `upstream returned <status> after N attempts`
4. 재시도는 스트림 시작 전 멱등 경계에 한정된다 — 응답이 시작된 뒤에는 절대 재시도하지 않는다 (다이얼 수 1로 관측).
5. 상류 4xx는 재시도하지 않는다 (다이얼 수 1, 기존 응답 유지).

## Evidence

명령: `go test ./internal/gateway/ -count=1 -run 'TestUpstreamRetry|TestUpstreamNoRetry|TestUpstreamFourHundred|TestOpenAIUpstreamRetryWiring'` — 최초 실행(구현 전):

```
--- FAIL: TestUpstreamRetryRecoversSingleFiveHundred
    retry did not recover from a single upstream 502: {"error":{"message":"Bad Gateway",...}}
--- FAIL: TestUpstreamRetryExhaustionLabelsUpstreamOrigin
    exhaustion message missing upstream origin label: "Bad Gateway"
--- FAIL: TestUpstreamRetryDialFailureRecovers
--- FAIL: TestUpstreamRetryDialExhaustionCarriesReason
--- FAIL: TestOpenAIUpstreamRetryWiring (recovers / labels_exhaustion)
```

같은 명령, 구현 후: 전부 PASS. (RED 단계에서 사건의 형태 — 사유 없는 게이트웨이 렌더 `Bad Gateway` — 그대로 재현.)

전수: `go test ./internal/gateway/ -count=1 -v` → **`--- PASS` 94건**, `--- FAIL` 1건(아래 Gaps).
`-race` 동일 결과(데이터 레이스 0).
`go vet ./internal/gateway/` clean. `gofmt -l internal/gateway/` empty. 커버리지: `go test -cover` → `coverage: 91.8% of statements` (목표 85%).
golangci-lint: 신규 프로덕션 파일의 errcheck 1건 수정(`_ = resp.Body.Close()`). 잔여 223건은 전부 기존 파일 소속(base 상태, 본 카드 범위 밖).

커넥션 신선도 관측: 스텁 리스너의 `DialTLSContext` 카운터로 각 시도의 새 다이얼을 직접 셈 — 502-1회-후-200 케이스에서 다이얼 수 2, 소진 케이스에서 3. (어댑터 transport는 기존 `DisableKeepAlives=true`로 요청마다 새 TCP를 열며, 테스트가 그 성질을 관측으로 고정했다.)

## Baseline-attribution

- 모든 측정은 본 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t697`)의 커밋 `1ea7b1127`(기점 `44e56d017`)에서, 2026-09-13 본 세션의 실행으로 관측한 출력이다.
- 귀속 통제: `TestAppServerSubprocessHTTPToolContinuation` 실패의 귀속을 확인하기 위해 `-overlay` 빌드로 base 소스(`44e56d017`의 anthropic.go/openai.go/openai_test.go + 신규 2파일 스텁) 상태에서 동일 테스트를 재실행 → **동일 실패** (`cmd.Start()` exec 단계, 서브프로세스 스폰 — 내 변경과 무관한 환경 원인).

## 변경 파일 (배차 지시에 따라 명시 — lane-7 t700과의 겹침 판정용)

- `internal/gateway/upstream.go` (신규) — `upstreamSend` 재시도 헬퍼 + 생성지 라벨 2종
- `internal/gateway/upstream_retry_test.go` (신규) — 회귀/경계 테스트 7건
- `internal/gateway/anthropic.go` — egress를 upstreamSend로 배선, 502 라벨
- `internal/gateway/openai.go` — 동일 (PKCE/API-key 양경로 클로저)
- `internal/gateway/anthropic_test.go` — 503 케이스 기대값 갱신(시도 3회, Retry-After 보존)
- `internal/gateway/openai_test.go` — 동일 갱신 + 소진 사유 전달 단언

t700(웨지 재뿌리 정책, plan 카드)은 코드 미변경이므로 실충돌 표면 없음. 병합 순서는 배차 지시대로 t697 먼저.

## Gaps

- `TestAppServerSubprocessHTTPToolContinuation`: 본 환경에서는 관측되지 않음(샌드박스 exec 제한, base에서도 동일). CI 깨끗한 러너의 판정이 이 테스트의 근거다.
- 구독(SendAuthorized/PKCE) 경로의 재시도: 공유 헬퍼 + API-key 경로 실소켓 테스트로 간접 커버. PKCE 경로의 실소켓 e2e는 본 카드에서 미작성.
- 라이브 게이트웨이(127.0.0.1:50879) 재기동은 카드 운영 항목상 운영자 결정 사항 — 본 카드에서 실행하지 않음.
- 패키지에 로거가 없어 "로그에서의 구별"은 응답 본문 메시지가 운반하며, 그 메시지가 클라이언트 디버그 로그에 그대로 남는다. 게이트웨이 자체 로그 설비 추가는 범위 밖으로 보류했다.

## Residual-risk

- 상류가 5xx를 "1회성"으로 주는 경우 최대 3회까지 요청이 상류에 도달한다 — 재시도 상한은 카드 요구(제한 재시도)이나, 과금 민감 상류에서 상한 값(3)의 적정성은 운영 관측 후 조정 여지가 있다.
- 재시도 간 지연 없음(즉시 재시도) — 5xx 폭풍 중 짧은 버스트가 상류에 추가된다. 상한 3회로 bounded라 폭풍 증폭은 제한적.
- 구현 중 발견·수리한 잠재결함: 재시도 시 소진된 `*bytes.Reader` 요청 본문으로 2회차가 실패하는 문제 — `req.GetBody`로 본문을 재생성해 수리했고, 다이얼-실패 회귀 테스트가 이 경로를 겨냥한다.
