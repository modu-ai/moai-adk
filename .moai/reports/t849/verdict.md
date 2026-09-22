# t849 verdict — calls.Load() 다이얼 계수 단정의 served 카운터 이관 + 봉쇄 전송 수리

- card: t849 · class B (Tier S, 테스트 전용) · branch `WT-served-counter` · base `origin/develop` `643abfb8c`
- 제품 코드 무변경 (변경 8파일 전부 `*_test.go`) · push 없음 · develop 병합은 리드 창 대기

## Claim

1. `oaiTLS` / `nativeTLS`가 돌려주던 TLS 다이얼 계수를 서버측 served 카운터(t839 패턴)로 이관했고, 소비 테스트 전수의 `calls.Load()` 단정이 `served.Load()`로 갱신됐다.
2. 핸들러 내 봉쇄 채널 전송 2곳(glm_test.go, replay_egress_test.go)이 select/default 비봉쇄 전송으로 수리됐다.
3. 판정 기준 3항목(대상 -race -count=20 PASS / `go test -race -count=1 ./internal/gateway/` ok / lint 0)이 모두 충족됐다.
4. (수리 동반) `appserver_integration_test.go`의 pyenv-shim 중첩 셸뱅 ENOEXEC 기존 결함을 테스트 픽스처 수정으로 수리했다 — 베이스부터 이 환경에서 패키지 게이트를 막던 기존 실패.

## Evidence

| # | 명령 | 결과 |
|---|---|---|
| 1 | `go test -race -count=20 -run '<대상 17개 테스트>' ./internal/gateway/` | `ok github.com/modu-ai/moai-adk/internal/gateway 7.838s`, exit 0 |
| 2 | `go test -race -count=1 ./internal/gateway/` | `ok github.com/modu-ai/moai-adk/internal/gateway 10.538s`, exit 0 |
| 3 | `golangci-lint run ./internal/gateway/` | `0 issues.`, exit 0 |
| 4 | `gofmt -l internal/gateway/` + `go vet ./internal/gateway/` | 출력 없음 + `VET_OK` |
| 5 | 선택자 매칭 검증: `-v` 실행 후 `=== RUN` 최상위 unique 수 | 17/17 (서브테스트 포함 23 RUN) |
| 6 | 베이스 재측정: `git archive HEAD`(=`643abfb8c`) → /tmp 트리에서 `TestAppServerSubprocessHTTPToolContinuation` | `--- FAIL ... app server start failed` (수리 전), exit 1 — 즉 내 변경과 무관한 기존 실패 |

로그 사본: 본 디렉터리 `t849-target-final.log` · `t849-pkg-final.log` · `t849-lint-final.log` · `t849-base-repro.log`

### 주요 변경 (8파일, 전부 테스트)

- `openai_test.go` — `oaiTLS`가 핸들러 진입 계수(`served`)를 돌려주도록 교체, t839 해설 주석을 헬퍼로 이동. `TestOpenAISubscriptionUsesStoreAuthorizedBoundary`의 로컬 served 카운터는 헬퍼 반환값으로 흡수(중복 제거). `wantCalls` → `wantServed` (5xx=3 재시도 단정은 유지 — 스텁이 매 시도를 serve하므로 served==3 성립).
- `anthropic_test.go` — `nativeTLS` 동일 교체 + `TestAnthropicNativeStatusAndRedirectBoundaries` `wantServed`.
- `openai_subscription_test.go`(4곳) · `glm_test.go` · `native_policy_test.go`(2곳) · `input_estimate_test.go`(2곳) · `replay_egress_test.go` — 변수명·메시지 `calls` → `served`.
- 봉쇄 전송 수리: `glm_test.go` `seen <- r`, `replay_egress_test.go` `sent <- string(b)` → select/default(선캡처 우선, 초과 드랍).
- `appserver_integration_test.go` — 픽스처를 2파일로 분해: `#!/<bash>` wrapper(실제 Mach-O 해석기) + `fake-codex-impl.py`(python 본문, 0600). wrapper는 `exec <python3> <impl> "$@"`.

### appserver 기존 결함 규명 (게이트를 막던 것)

- **원인**: `exec.LookPath("python3")`가 pyenv shim(`~/.pyenv/shims/python3`)을 고르는데, shim은 bash 스크립트(`#!/usr/bin/env bash`)다. 픽스처의 셸뱅이 스크립트를 가리키는 **중첩 셸뱅** → macOS 커널은 ENOEXEC, Go fork/exec는 사용자 공간 폴백이 없어 `fork/exec ...: exec format error`. zsh/bash는 ENOEXEC 폴백이 있어 수동 실행·셸 경유 실행은 성공(즉, CI 리눅스·셸 수동 재현은 통과하는 로컬 환경 특이 실패).
- **프로브 실측**: 임시 테스트로 `cmd.Start: fork/exec /private/var/.../fake-codex: exec format error` 직접 확인(프로브 파일은 커밋 전 삭제).
- **귀속**: 위 Evidence 6행 — 수정 전 pristine 베이스 트리에서 동일 FAIL. 내 변경(카운터 의미 교체)은 이 테스트 경로와 무관.
- **수리 검증**: `go test -race -count=3 -run TestAppServerSubprocessHTTPToolContinuation ./internal/gateway/` → `ok 5.583s`.

## Baseline-attribution

- 전부 이 실행, 이 워크트리(`.claude/worktrees/t849`, HEAD `643abfb8c` 분기 `WT-served-counter`)에서 측정.
- 베이스 비교는 `git archive HEAD`로 뽑은 pristine `643abfb8c` 트리(/tmp/t849-base)에서 별도 측정 — 워킹 사본 혼입 없음.

## Gaps

- `-race -count=20`은 대상 17개 테스트에만 적용 — 패키지 나머지는 `-count=1` 기준.
- `internal/codexbridge/lifecycle_subprocess_test.go`의 동일 결함군 3곳(`LookPath("python3")` + `#!`+python 픽스처)은 이번 카드 판정식(`./internal/gateway/`) 밖이라 **미수리** — 후속 카드 후보(같은 2파일 분해 수리면 해소).
- Windows 실측 없음(로컬 darwin만). bash 부재 환경에서는 신규 `t.Skip("bash unavailable")` 분기로 전락 — CI 매트릭스 판정은 CI 몫.

## Residual-risk

- `wantServed=3`(5xx 재시도)은 "스텁이 매 시도를 serve한다"는 전제 위에 섬 — 향후 upstreamSend가 핸들러 재진입 없이 재시도하도록 바뀌면 이 단정은 갱신 필요.
- appserver 수리는 bash 폴백 동작(셸뱅을 주석으로 해석)에 의존하는 pyenv 환경 특성 위에 섬 — python3가 정상 바이너리인 환경에서는 wrapper 없이 직접 exec되는 것과 동치라 무해.
- served 카운터는 "핸들러 진입 수"로 정의가 바뀐 것 — 다이얼 수 자체를 단정해야 하는 미래 테스트는 별도 계수가 필요.
