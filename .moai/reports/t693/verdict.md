# t693 판정서 — develop CI 적색 3축 수리

- baseline HEAD: `74d872aaf` (= origin/develop, CI run 34746128869 적색)
- 작업 트리: `.claude/worktrees/t693`, 브랜치 `WT-develop-ci-red`
- 커밋: 축1 `c2b1af1f3` · 축2 `aaad584cf` · 축3 `8d92fcee4`

---

## 1. Claim (주장)

| 축 | 주장 | 상태 |
|---|---|---|
| 1 | Lint 잡 적색은 커밋된 t649 증거 테스트 4개 파일의 gofmt 위반이 원인이며, 해당 4개만 gofmt로 수리했다 | **수리 완료** |
| 2 | Test 잡 적색은 `TestProductionGatewayFactoryRejectsUnauthenticatedRequest`가 호스트의 `codex` 바이너리 존재에 의존하는 환경 의존 결함이며, hermetic 스텁으로 수리했다 | **수리 완료** |
| 3 | Race 잡 적색(`TestRegressionNativeEgressPreservesOpaqueItemBytes/true`, 502 calls=1)은 race 잡에서 단 1회 관측된 부하 의존 flaky이다. 테스트 하네스의 데이터 레이스는 제거했으나, 502 발화 기제는 기계적으로 확정하지 못했다 | **부분 수리 + 원인 미확정** |

## 2. Evidence (증거)

### 축 1 — gofmt 4파일 (커밋 `c2b1af1f3`)

```
$ gofmt -l .moai/reports/t649/        # RED (74d872aaf)
.moai/reports/t649/context-window-evidence/e2e/context_probe_test.go
.moai/reports/t649/independent-evidence/audit_gateway_test.go
.moai/reports/t649/independent-evidence/audit_test.go
.moai/reports/t649/reaudit-evidence/audit_gateway_test.go

$ gofmt -w <위 4개 파일> && gofmt -l .moai/reports/t649/   # GREEN
(빈 출력 — 로그: .moai/reports/t693/axis1-gofmt.log)
```

diff는 순수 포맷 재배열(gofmt는 토큰을 변경하지 않음). 4파일 274+/84−.

### 축 2 — hermetic codex 스텁 (커밋 `aaad584cf`)

RED (codex를 PATH에서 제거하고 재관찰):

```
$ env PATH="/usr/bin:/bin:/opt/homebrew/bin" go test -count=1 \
    -run TestProductionGatewayFactoryRejectsUnauthenticatedRequest ./internal/cli/
--- FAIL: TestProductionGatewayFactoryRejectsUnauthenticatedRequest (0.00s)
    gateway_product_binding_test.go:83: gateway private configuration or verified dependencies unavailable
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.162s
```

근본 원리: `installedGPTBrokerForGateway` (`internal/cli/gateway_product_binding.go:211`)의 `exec.LookPath("codex")` 실패 → `errGatewayFactory` → 테스트가 팩토리 구성 단계(:83)에서 실패. CI ubuntu에는 codex가 없다.

수리: 테스트가 `t.TempDir()`에 실행 가능한 `codex` 스텁(Windows는 `codex.bat`)을 만들고 `t.Setenv("PATH", ...)`로 앞단에 붙인다. 401 단정은 자격 증명 resolve 이전에 일어나므로 스텁으로 계약 충족.

GREEN (동일한 codex 없는 PATH):

```
ok  	github.com/modu-ai/moai-adk/internal/cli	0.935s
```

동일 파일 인접 테스트 7개 선별 회귀 실행: `ok ... 0.866s` (`.moai/reports/t693/axis2-green.log`)

### 축 3 — replay egress race 502 (커밋 `8d92fcee4`)

CI 원문 (race job 103694333029):

```
--- FAIL: TestRegressionNativeEgressPreservesOpaqueItemBytes/true (0.03s)
    replay_egress_test.go:49: probe failed status=502 calls=1
```

로컬 재현 시도 5레시피 — 전부 미재현(각 로그는 `.moai/reports/t693/axis3-repro-*.log`):

| 레시피 | 명령 | 결과 |
|---|---|---|
| 1a | `go test -count=5 -race -run TestRegression... ./internal/gateway/` | 통과 1.676s |
| 1b | `go test -count=5 -race ./internal/gateway/` (전체 패키지) | 통과 (범위 밖 TestAppServer만 실패) |
| 2 | `GOMAXPROCS=2 go test -count=5 -race ...` | 통과 1.504s |
| 3 | gateway 전체 race ×5 ⇄ `./internal/cli/` race 동시 실행 | gateway: 미재현 / cli: 10분 타임아웃 panic(과부하 유발물) |
| 4 | GOMAXPROCS=4 게이트웨이 race 바이너리 3개 ×count=3 동시 | 미재현 |

CI 이력 대조 (기계 확인): run 34736881337 (`8eade2f89`, 이 커밋에도 동일 테스트 존재 — `git merge-base --is-ancestor ce79ef7ca 8eade2f89` 확인)의 race 잡에서는 이 테스트가 **통과**, 축 2 테스트만 실패. 즉 축 3은 race 잡 조건에서 드물게 발화하는 flaky.

분석(코드 판독 기반): 502는 `SendAuthorized`가 비인증 오류를 반환한 경로(`internal/gateway/openai.go:189`)에서 산출. 소요 0.03s이므로 1s `WriteTimeout` 타이머 발화는 배제되고, 30ms 내 `sendCtx`를 취소할 수 있는 유일한 동적 요인은 5ms 폴링 워처의 `Status()` 실패·불일치 시 즉시 cancel하는 구조다. 다만 안정적인 tempdir에서 `Status()`가 어떻게 오동작하는지는 확정하지 못했다.

수리한 것(테스트 하네스 결함 — 검사상 실재하는 데이터 레이스):

- 핸들러 고루틴이 `sent` 공유 변수에 쓰고 테스트 고루틴이 읽는 비동기화 접근을 버퍼드 채널 전달로 교체.
- probe 실패 시 수집한 응답 본문을 실패 메시지에 포함 → 다음 CI 재발 시 502 단락이 운반하는 진단이 즉시 확보된다.

제품 코드는 변경하지 않았다 — 워처의 fail-closed는 의도된 보안 설계이며(`send.go` @MX:WARN), 502 유발 주체가 제품 결함이라는 기계적 증거가 없었다.

GREEN:

```
$ go test -count=3 -race -run TestRegressionNativeEgressPreservesOpaqueItemBytes ./internal/gateway/
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.698s
```

전체 패키지 race 1회: replay 테스트 통과, 실패는 범위 밖 `TestAppServerSubprocessHTTPToolContinuation`뿐(CI는 해당 테스트에서 녹색 — 로컬 환경 관측치, 무시 지시 준수).

크로스 컴파일: `GOOS=windows GOARCH=amd64 go vet ./internal/gateway/ ./internal/cli/` → `VET_OK`.

## 3. Baseline-attribution (baseline 귀속)

모든 명령은 이번 실행에서 작업 트리 `.claude/worktrees/t693` (HEAD `74d872aaf`에서 시작, 커밋 후 `c2b1af1f3`→`aaad584cf`→`8d92fcee4`)에 대해 직접 실행했고, 출력은 그대로 위에 인용됐다. 타 패키지·타 시점 수치 반입 없음.

## 4. Gaps (미검증)

- **축 3의 502 발화 기제 미확정**: 5개 로컬 재현 레시피 전부 실패. 워처 cancel 가설은 코드 구조상 유일한 동적 후보이나 기계적으로 증명하지 못했다. CI race 잡에서 재발 여부로만 검증 가능하다 — 재발 시 새 실패 로그가 응답 본문을 운반하도록 진단을 심어 두었다.
- **축 1·2의 CI 녹색은 미관측**: 로컬 GREEN만 관찰했다. `make fmt-check` 전체와 CI 잡 판정은 push 후 CI 몫이다(로컬 전체 스위트 금지 규율).
- 축 3에서 `sent` 레이스를 race detector가 실제로 감지한 적은 없다(로컬·CI 모두) — 제거 근거는 코드 검사상 실재이지 탐지기 출력이 아니다.
- 재현 레시피 3의 cli 10분 타임아웃 panic은 과부하 유발물로, 제품 결함 근거가 아니다.

## 5. Residual-risk (잔여 위험)

- race 잡의 502 flaky는 재발할 수 있다. 재발 시 이번 커밋의 실패 메시지(`body=` 포함)가 원인 단서를 제공한다. 재발 패턴이 워처 cancel로 확정되면 그때 제품 경로(예: Status 일시 오류의 재시도 허용)를 근거와 함께 수리하는 후속 카드가 권장된다.
- `TestAppServerSubprocessHTTPToolContinuation`은 이 머신에서 재현되는 별개 환경 관측치다(본 카드 범위 밖).
- gofmt 대상 4파일은 t649 증거물이라 앞으로도 압축 스타일로 추가 작성될 경우 같은 적색이 재발할 수 있다(작성 측 규율 필요).
