# t229 — M1 경계 상태 (세션 마감 시점)

운영자 마감 지시로 **M1 까지만** 착지시키고 세션을 종료한다. M2 는 착수하지 않았다.

| 항목 | 값 |
|---|---|
| 브랜치 | `WT-audit-verdict-converge` |
| 측정 트리 | M1 커밋 직전 스테이지 상태 (base `a7d1001ee`) |
| 측정 일자 | 2026-08-25 |

## M1 이 실제로 한 일 (실측)

`synthesizeReviewOutput(reviewText, method string)` 로 시그니처를 바꾸고, 판정 로직을 **신호 수집 + 집합 최댓값 채택** 구조로 재작성했다.

| 새 심볼 | 역할 |
|---|---|
| `codexVerdictSignal` | 신호 하나(출처 이름 + verdict). 출처 이름은 불일치를 운영자 언어로 설명하기 위해 있다 |
| `codexVerdictSignalsOf` | 본문이 담은 **모든** 신호를 모은다. 결정하지 않는다 — 넷째 신호는 여기에만 추가된다 |
| `codexVerdictRank` | `fail`(3) > `inconclusive`(2) > `pass`(1), 미상(0) |
| `adoptConservativeVerdict` | **P-CONS** — 집합의 최댓값 채택. 순서 의미론 없음 |
| `codexUnrecognizedVerdict` | 신호 집합이 **비었을 때**의 모드 기본값. native → `pass`, 그 외 → `inconclusive` |

[HARD] 구현이 리드 지침 2를 지켰다: **집합 연산**이며 순서 있는 대입이 아니다. `adoptConservativeVerdict` 주석이 그 이유를 코드에 남겨 두었다 — 순서 규칙으로 쓰면 넷째 신호에서 같은 구멍이 한 겹 옆으로 옮겨 간다는 것. 리드가 SPEC 에 남기라 한 조건부성이 코드 주석으로도 한 번 더 남았다(지침 3 충족).

`codexUnrecognizedVerdict` 주석에 **"인식기를 추가하는 일이 이 함수를 건드리게 만들면 안 된다"** 가 명시돼 있다 — 이 결함의 구조적 원인(판정이 한 CLI 버전 서식에 묶임)이 재발하는 경로를 코드에서 막았다.

## 실측 — 모드 분기가 실제로 배선됐다

```
T229-M1 | C1 adversarial=inconclusive native=pass
T229-M1 | C5 adversarial=inconclusive native=pass
```

**같은 본문이 모드에 따라 갈린다.** 착수 전 베이스라인에서 C1·C5 는 둘 다 `pass` 였다(8/8 RED). AC-CVS-001(미인식 서식은 adversarial 에서 `pass` 아님)과 AC-CVS-003(native 무불릿 정상 리뷰는 `pass` 보존)이 동시에 성립한다.

## [HARD] K3·K7 은 여전히 붉다 — 회귀가 아니라 설계대로다

```
T229-M1 | K1 adopted=fail         want=fail         GREEN
T229-M1 | K2 adopted=fail         want=fail         GREEN
T229-M1 | K3 adopted=pass         want=fail         RED
T229-M1 | K4 adopted=inconclusive want=inconclusive GREEN
T229-M1 | K5 adopted=fail         want=fail         GREEN
T229-M1 | K6 adopted=fail         want=fail         GREEN
T229-M1 | K7 adopted=pass         want=inconclusive RED
T229-M1 | K8 adopted=pass         want=pass         GREEN
```

착수 전 베이스라인과 **동일**하다(K3·K7 RED, 나머지 6행 GREEN). M1 은 이 두 행을 건드리지 않으며, 그것이 옳다:

- K3 = `Verdict: pass` + `FAIL 0.20 / 1.00` · K7 = `INCONCLUSIVE 0.50 / 1.00` + `Verdict: pass`
- 두 본문 모두 **`stated` 신호가 매치**되므로 신호 집합이 비지 않는다 → `codexUnrecognizedVerdict`(M1 이 고친 fall-through)를 **아예 타지 않는다**
- 붉은 이유는 점수 표기(`FAIL 0.20` / `INCONCLUSIVE 0.50`)가 **아직 신호로 인식되지 않아** 집합에 들어오지 못하고, 집합에 하나뿐인 `stated: pass` 가 최댓값이 되기 때문이다

[HARD] **이 두 행을 초록으로 뒤집는 것은 M2** — `codexScoredVerdict` 를 `codexVerdictSignalsOf` 에 추가해 점수 표기가 집합에 들어오는 순간, K3 의 집합은 {pass, fail} 이 되어 `fail` 이, K7 의 집합은 {inconclusive, pass} 가 되어 `inconclusive` 가 채택된다.

**다음 세션에 대한 지시**: 이 두 행이 붉은 것을 보고 기대값을 관측 동작(`pass`)에 맞춰 낮추지 말 것. 기대값은 P-CONS 에서 도출된 것이고, 낮추면 두 행의 검출력이 사라진다 — 그것이 이 카드가 없애려는 실패 형태 그 자체다(acceptance.md §C AC-CVS-006 `[HARD]`).

## 검증 상태 — 정직한 기록

| 검증 | 명령 | 결과 |
|---|---|---|
| 빌드 | `go build ./internal/cli/...` | **rc=0** |
| 정적 분석 | `go vet ./internal/cli/` | **rc=0** |
| 대상 테스트 | `go test ./internal/cli/ -run 'TestSynthesizeReviewOutput\|Fallthrough\|Verdict\|Blind' -count=1 -timeout 1200s` | **ok 36.937s** |
| M1 경계 실측 | `go test ./internal/cli/ -run TestT229PostM1State -v -count=1` | **ok 1.119s** (위 표) |
| **전 패키지 스위트** | `go test ./internal/cli/ -count=1 -timeout 1200s` | **FAIL @ 1032.479s** (아래) |

### [HARD] 전 패키지 스위트 — FAIL 이고, 래퍼는 그것을 숨겼다

마감 직전에 결과가 도착했다. 백그라운드 래퍼는 **`[exited with code 0]`** 을 보고했으나 출력 본문은 `FAIL github.com/modu-ai/moai-adk/internal/cli 1032.479s` 다. **래퍼의 종료 코드를 스위트 판정으로 읽으면 안 된다** — 이 저장소에 이미 기록된 함정(`feedback_bg_exitcode_direct_verify`)이 그대로 재현됐다.

**실패한 테스트 이름은 확보하지 못했다.** 출력 파일이 마지막 10줄만 남기고 잘려 `--- FAIL:` 행이 사라졌다. 따라서 원인을 **주장하지 않는다.**

확인한 것과 확인하지 못한 것을 갈라 적는다:

| 항목 | 상태 |
|---|---|
| 이 run 의 시작 시각 | 04:02 — **`mcp_convergence_test.go` 픽스처 수정 이전** |
| 그 픽스처 테스트(`TestPerformCodexAudit_ReusesExistingCodexHandler_NoReimpl_AP_AMM_1`) | 현재 트리에서 **PASS** (단독 실측) |
| `TestResolveRulesDir` (출력 꼬리에 이름이 보였던 것) | 현재 트리에서 **PASS** (단독 실측) |
| 실패 테스트 정체 | **미확인** |
| 픽스처 수정 후 스위트 상태 | **미검증** |

가장 그럴듯한 설명은 픽스처 테스트가 그 run 시점에 붉었고 이후 수정됐다는 것이다 — M1 이 fall-through 를 바꾸면서 `"codex:ok, no findings"` 본문이 `pass` → `inconclusive` 로 갈렸을 것이기 때문이다. **그러나 관측하지 않았으므로 확정하지 않는다.**

부수 관측: 1032초는 이 패키지의 평소 범위(약 300~540초)를 크게 넘는다. 같은 시각 백그라운드 run 이 여럿 돌고 있었으므로 **머신 부하 아티팩트**로 보이며, 코드 지연 신호로 읽지 말 것.

[HARD] **다음 세션의 첫 작업**: 깨끗한 상태에서 `go test ./internal/cli/ -count=1 -timeout 1200s` 를 **다시 돌리고, 래퍼 종료 코드가 아니라 출력 본문의 `ok`/`FAIL` 행을 읽어** 판정할 것. 붉으면 `--- FAIL:` 행에서 이름을 확보한 뒤 M2 착수 전에 해소한다.

미측정 항목(마감으로 생략): 커버리지(`go test -cover`), `golangci-lint`, `GOOS=windows go vet`.

## 남은 일

| # | 항목 |
|---|---|
| 1 | **전 패키지 스위트 재실행 + 결과 확인** (위 미완 항목) |
| 2 | M2 — `codexScoredVerdict` 신설. **착수 게이트: AC-CVS-006 테스트가 먼저 존재해야 한다**(현재 없음). 신호는 `codexVerdictSignalsOf` 에만 추가하고 `adoptConservativeVerdict`·`codexUnrecognizedVerdict` 는 건드리지 않는다 |
| 3 | M3 — `SynthesisNote` (`ReviewOutput` + `PerBackendVerdict`, `omitempty`) + `converge()` 의 `disagreement_flag`/`residual_risk_note` 반영. `overall_verdict` 불변 |
| 4 | M4 — 회귀 고정 (기존 테스트 native 명시 호출 확장, 라이브 프로브 픽스처, `codex_task` 출력 불변) |
| 5 | 착지 후 리드에게 t234 (= #1632) 착수 가능 신호 |
