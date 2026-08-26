---
id: SPEC-CI-FLAKE-SERIES-001
title: "CI 의존 테스트 플레이크 3종 계열 수리 — poller 조기탈출 TOCTOU / 페어링 ratio 통계 / 단일 샘플 벽시계 단언"
version: "0.2.0"
status: in-progress
created: 2026-08-26
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "post-v3.1.3 (next patch train)"
module: "internal/sessionmsg/store_test.go, internal/timing/timing.go, internal/timing/paired_test.go, internal/hook/config_change_test.go, .moai/reports/t278/"
lifecycle: spec-anchored
tags: "ci-flake, test-stability, toctou, paired-ratio, wall-clock-assertion, sessionmsg, timing, hook, reproduction-rate"
tier: M
era: V3R6
related_specs: [SPEC-CI-FLAKY-STABILIZE-001]
---

# SPEC: CI 의존 테스트 플레이크 3종 계열 수리

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-26 | manager-spec | 최초 작성 — 카드 t278 (t270·t271 흡수). plan-phase에서 CI 로그 attempt-1 포렌식 + 코드 정독으로 세 메커니즘 후보를 판별했고, 그 결과와 뒤집힌 카드 전제 2건을 §2에 고정 |
| 0.2.0 | 2026-08-26 | manager-spec | plan-audit iter1(PASS-WITH-DEBT 0.86) 반영 — D1: REQ-CFS-008에 AC-CFS-010 대응(관측 가능 판정 명령). 적용 중 발견: `TestMedianRunsWarmupPlusSamples`(timing_test.go:174)이 cpuUnit 호출 + t.Parallel 병존 — 규칙 스코프를 "지속시간·비 단언 측정 테스트(자가선언 5종)"로 정밀화(이 테스트는 tick-양성 검사로 부하 강건, 면제 명시). D2: 관측 단위 N을 "go_code=true workflow **run** 수"로 고정(REQ-CFS-010·AC-CFS-007·plan M4 일관). D3: related_specs에서 실재하지 않는 SPEC-V3R6-CI-PR-SPEEDUP-001 제거(ci.yml 주석 인용으로만 존재 — §2.5에 주석 근거 명시). D4: §2.2 "≥10개"→"≥11개" 정정. D5: plan §B.3 wg 대기 방식 구체화 |

## 1. 문제 — 측정된 형태

측정 기준 트리: 본 워크트리 `.claude/worktrees/t278` @ `175d63f3f` (origin/main, 2026-08-26 체크아웃). 세 플레이크 모두 **로컬 darwin에서는 재현되지 않고**(카드 기록: #1 `-race -count=20` 전부 PASS, #2·#3 `-race -count=5` 전부 PASS) **CI ubuntu 러너에서만 산발적으로 발화**한다. 발화는 초 단위 안에 끝난다 — 60초 데드라인이나 2초 비동기 대기가 아니라 테스트 자체의 판정 논리가 조기에 결론 내린 형태다.

### 1.1 Flake #1 — `TestConcurrentSendPoll` (internal/sessionmsg)

관측 1 (카드 기록 + 이번 포렌식 재확보):

```console
$ gh run view 32774108273 --attempt 1 --log-failed
Test (ubuntu-latest)  Run tests with coverage (fast — no race detector)
  2026-08-24T20:33:19.8564953Z --- FAIL: TestConcurrentSendPoll (0.07s)
  2026-08-24T20:33:19.8566145Z     store_test.go:385: received 97 messages, want 100 (loss or duplication)
```

- PR #1601 (docs-only — `WT-project-pipeline`)의 attempt 1. **필수 머지 게이트 `Test (ubuntu-latest)` job에서 발화** → 머지 차단 + 수동 재실행 비용.
- 실패 소요 0.07s — poller의 60s deadline(`store_test.go:351`)과 무관. `poll:` / `send` 에러 로그가 **한 건도 없다**(전체 실패 로그에서 위 2줄이 유일한 테스트 출력).

관측 2 (이번 plan-phase에서 신규 발견 — 카드에 없던 재발):

```console
$ gh run view 32777242100 --log-failed
Race Test  Run race detector across all packages
  2026-08-24T21:07:31.3326266Z --- FAIL: TestConcurrentSendPoll (0.27s)
```

- 같은 날 35분 뒤 **Race Test job(advisory)에서도 재발**. 두 job 모두에서 터진다.

### 1.2 Flake #2 — `TestAssertPairedHealthyEndToEnd` (internal/timing)

```console
$ gh run view 32779472351 --attempt 1 --log-failed
Race Test  UNKNOWN STEP
  2026-08-24T21:35:42.6609240Z --- FAIL: TestAssertPairedHealthyEndToEnd (0.13s)
  2026-08-24T21:35:42.6611086Z     timing.go:233: paired-cpu-1x: n=20 median=1.502ms p95=4.479ms worst=5.355ms avg=2.812ms | refUnit=1.375ms ratio=2.47x (maxUnits=2.00x, steadyCeiling=10s, budget=30s)
  2026-08-24T21:35:42.6615423Z     timing.go:237: paired-cpu-1x: measured latency is 2.47x the reference unit (median 1.502491ms), above the 2.00x calibrated bound — ...this is a code regression, not machine load (load inflates the reference equally)
```

- PR #1600 (`WT-server-version`, internal/cli만 변경)의 Race Test job attempt 1. ref와 fn은 **바이트 동일한 `cpuUnit(2_000_000)`**(`paired_test.go:68`) — 참 비용 비는 1.00x, 상한 2.00x.
- **결정적 판별자**: 같은 로그 라인에서 median(fn)=1.502ms, median(ref)=refUnit=1.375ms → **median 비는 1.09x(건강)** 인데, 라운드별 비의 중앙값(per-round paired ratio)만 2.47x. 노이즈가 라운드 교대(`measurePaired`의 ref-first/fn-first 교차, `timing.go:342-348`)에 위상고정돼 한쪽으로 쏠린 흔적.
- 이 플레이크는 유서(有書)가 있다: `timing_test.go:18-24` 주석과 `paired_step_test.go:20-24`가 이전 관측(2.32x / 2.72x / 4.64x — #1591 페어링 이전; `TestBranchGuard_Latency` 1.82x, run 32687843472 attempt 1)을 기록한다. #1591의 페어링(per-round 비)은 load-step 취약성을 닫았지만, 2026-08-24 관측은 **반대 방향**(per-round 비가 교대 비대칭에 취약)의 잔여 실패 모드다.

### 1.3 Flake #3 — `TestConfigChange_RT005ReloadIntegration` (internal/hook)

```console
$ gh run view 32815411885 --attempt 1 --log-failed
Race Test  Run race detector across all packages
  2026-08-25T06:12:45.1059931Z --- FAIL: TestConfigChange_RT005ReloadIntegration (0.15s)
  2026-08-25T06:12:45.1060749Z     config_change_test.go:51: synchronous return took 123.195919ms, want ≤ 100ms (REQ-HAE-002)
```

- PR #1653 (docs-only)의 Race Test job attempt 1.
- 실패한 것은 `WaitForAsync`(2s 예산, `testutil/wait_async.go:40-57` — 타임아웃이면 테스트가 2s 이상 걸린다)이 아니라 **단일 샘플 벽시계 단언**(`config_change_test.go:50-52`). 측정 구간은 `h.Handle`의 동기 경로 — `slog.Info` 1회 + `context.WithTimeout` + `go func()` spawn뿐(`config_change.go:56-84`)으로 µs 규모다. 123ms는 연산 비용이 아니라 **그 고루틴이 받은 ~120ms 스케줄러 선점**(t.Parallel 테스트 + 전체 패키지 -race + 2 vCPU 러너)이다.
- 계약 자체는 p95다: REQ-HAE-002 / AC-HAE-003은 "10-동시 부하에서 p95 ≤ 100ms"이고 그 측정기는 `BenchmarkConfigChange_AsyncReturn`(`config_change_test.go:271-313`, p95 산출)이다. 테스트는 이를 **per-sample 최댓값**으로 시행하고 있다 — 범주 착오.

## 2. 조사 결과 — 코드 정독 + 로그 포렌식 (plan-phase 수행)

카드는 Class B(원인 미상)로 배차됐다. plan-phase에서 세 실패의 attempt-1 로그를 확보하고 아래 코드를 정독했다: `store_test.go:302-397`, `store.go:206-427`, `lock.go:69-114`, `timing.go:1-100·180-238·289-356`, `paired_test.go`, `paired_step_test.go`, `config_change.go`, `config_change_test.go`, `testutil/wait_async.go`, `.github/workflows/ci.yml:85-240`.

### 2.1 Flake #1 — 메커니즘 확정: poller 종료 조건의 TOCTOU (테스트 측 결함)

poller 루프(`store_test.go:352-371`)의 탈출 논리:

```go
res, err := s.Poll(receiver.AgentID, nil)        // (1) pending 리스팅 — 이 시각의 스냅샷
...
if len(res.Messages) == 0 && res.Remaining == 0 { // (2) "drain 완료" 관측
    select {
    case <-sendersDone:                           // (3) 송신 완료 확인 — 관측 '이후'의 시각
        stop = true
    ...
```

치명적 창: (1)의 `listEnvelopes`(pending 리스팅, `store.go:383`) 시각과 (3)의 select 시각 사이에 마지막 Send들의 pending rename(`store.go:262-264`)이 끼어들 수 있다. `sendersDone`은 `wgSend.Wait()` 후 close되므로 (3)에서 채널이 닫혀 보여도, 그것은 (1)의 리스팅 이후에 닫힌 것인지 보증하지 않는다. 두 poller가 모두 (1)에서 빈 pending을 보고(마지막 3건의 Send가 아직 rename 전) (3)에서 닫힌 채널을 보면 — 두 poller 모두 탈출, 3건은 영원히 미수령. **97/100 손실 + 에러 0건 + 0.07s 종료**의 전부가 이 하나의 인터리빙으로 설명된다.

store 구현 자체는 결함 후보에서 **제외**된다:
- Send/Poll 모두 동일 per-mailbox lock(`lock.go:69-114`: in-process mutex + NB-flock)으로 직렬화 — claim/Remaining 부기 경쟁 불가.
- claim 순서는 claimed-쓰기 → pending-삭제(`store.go:401-412`)로, 어느 순간에도 메시지가 양쪽 디렉터리에서 동시에 부재가 되지 않는다.
- TTL 스윕(24h/10m, `config/defaults.go:401-404`)은 0.07s 창에서 발동 불가.
- Poll 에러 경로(클로저 중도 실패 → `PollResult{}` 폐기)는 손실을 만들 수 있지만 **반드시 `poll:` 에러를 남긴다** — CI 로그에 해당 에러가 0건이므로 배제됨.

### 2.2 Flake #2 — 메커니즘 후보 2계열, 판별 데이터는 통계 자체에 있음

관측된 요약 통계(median 1.502 / refUnit 1.375 / per-round 중앙값 2.47)가 성립하려면 20개 라운드 비 중 **≥11개**가 2.47 이상이어야 하고(`medianFloat`은 `sorted[len/2]` = 20개 중 11번째 값 — 11개 미만이면 중앙값이 그 아래로 내려감; 반대 방향 비들이 양쪽 중앙값을 거의 같게 유지) — 이는 노이즈가 **라운드 교대 순서에 위상고정**된 계통적 비대칭일 때만 나온다. 후보:
- (i) 교대 순서와 위상고정된 스케줄링 비대칭(-race 하 fake-clock/타이밍 코드 개입, 코어 마이그레이션) — 요약 통계와 정합.
- (ii) CFS 할당량(cgroup throttle, 100ms 주기)이 in-flight 샘플 하나를 동결 — 양쪽에 무작위로 떨어지므로 (i)만큼 중앙값을 밀기 어려움.

현재 추정량의 구조적 사실이 설계의 핵심이다: `measurePaired`의 per-round 비 중앙값은 **load-step에는 강하고**(`paired_step_test.go`가 핀) **교대 비대칭에는 약하며**, ratio-of-medians은 그 반대(관측 라인에서 1.09x로 건강). 두 추정량이 상호 보완적이다.

### 2.3 Flake #3 — 메커니즘 확정: 단일 샘플 벽시계 단언의 범주 착오

동기 경로 비용은 µs 단위(§1.3). 123.195919ms는 측정 창 안의 선점이다. 계약(REQ-HAE-002, AC-HAE-003 p95)과 시행(단일 샘플 최댓값)의 불일치가 결함이며, 수정은 시행을 계약 의미론(p95, 다중 샘플)에 맞추는 것이다.

### 2.4 카드 전제 판정표

| # | 카드 전제 | 판정 | 근거 |
|---|---|---|---|
| 1 | #1의 원인은 "store.go claim/Remaining 부기의 race — 마지막 poller의 claimed 메시지가 다른 poller에 drain-complete로 보임" | **전복** — 결함은 store가 아니라 `store_test.go:361-370`의 종료 규칙 TOCTOU | lock.go 직렬화 + claim 순서 정독; CI 로그 `poll:` 에러 0건으로 store 에러경로 배제; 0/0 관측 시각과 select 시각의 분리가 97 손실을 무에러로 설명 |
| 2 | #2는 "ratio가 실패 시에만 노출돼 건강 baseline이 없음" | **전복** — `publish()`(`timing.go:254-265`)이 #1591(t162)부터 **모든 실행의 ratio를 job summary(GITHUB_STEP_SUMMARY)에 기록** | publish_test.go 존재 + timing.go:241-253 주석("records the healthy figure on green runs too"); 관측 채널은 이미 존재하며 M1이 이를 수집한다 |
| 3 | #2의 paired 설계는 "부하가 reference를 동등히 부풀린다" 가정 — CI 비대칭 스로틀링이 이를 깰 수 있음 | **유지·정밀화** — 깨는 방향이 확인됨: per-round 비가 교대 위상고정 비대칭에 취약 | §1.2 판별자(ratio-of-medians 1.09x vs per-round 2.47x) |
| 4 | #3 실패(0.15s) | **유지·확정** — 실패 단언이 `config_change_test.go:51`의 벽시계 검사로 특정됨 | §1.3 로그 + WaitForAsync 구조 대조 |
| 5 | 공통 패턴: "로컬 darwin 초록 + CI 단발 적신호 + 초 단위 실패 → CI 러너 환경 특성" | **유지** — 계열 인자는 §3 | 세 사례 모두 ubuntu 2-vCPU 공유 러너의 풀스위트 병렬 부하下的 스케줄링/시간 의존성 |

### 2.5 CI 관측 채널의 구조 (관측 가능성의 전제)

- `test` job "Run tests with coverage (fast — no race detector)": `go test -coverprofile=... ./...`, **no -race, 필수 머지 게이트** — flake #1 발화 시 머지 차단.
- `test-race` job "Race Test": `go test -race -count=1 ./...`, advisory(비필수) — #1(재발)·#2·#3 발화 위치.
- **`gh run list --status failure`는 attempt-1 실패를 숨긴다**(재실행이 초록이면 목록에서 사라진다 — 실측: 최근 60 run 중 실패 3건만 보였으나 그 창에서 플레이크 발화는 4건). 모든 발화 계수는 **attempt-aware** sweep이어야 한다(REQ-CFS-009).
- job 분리(test / test-race)의 설계 배경은 `.github/workflows/ci.yml` 주석의 `SPEC-V3R6-CI-PR-SPEEDUP-001` 인용으로만 전승된다 — 해당 이름의 SPEC 문서는 저장소에 존재한 적이 없음(audit D3 실측). 본 SPEC은 ci.yml 주석을 근거로 인용한다.

## 3. 공통 인자 — 계열 분석

세 플레이크의 공통형: **시간·스케줄링 의존 판정을 단일 관측으로 내리는 테스트가, 경합 중인 2-vCPU 공유 러너의 스케줄링 비결정성을 신호와 구분하지 못한다.**

| 사례 | 단일 관측 | 노이즈가 침투한 창 | 계약이 실제로 요구하는 것 |
|---|---|---|---|
| #1 | "이번 poll은 0/0이다" → drain 완료로 간주 | 관측 시각 ↔ 종료 확인 시각 | **긍정 조건**: 송신 완료 이후에 시작한 poll의 0/0 |
| #2 | 라운드별 비의 중앙값 1개 | 측정 창 전체의 교대 위상 | 부하에 강인한 비 통계(노이즈 이분법: step ↔ 교대 비대칭) |
| #3 | Handle 1회의 벽시계 | start ↔ elapsed 측정 | p95(다중 샘플) ≤ 100ms |

수리 원칙(각 사례에 적용): ① 부재(absence-of-work)가 아니라 긍정 조건을 기다린다 ② 단언의 통계가 계약의 통계와 같다 ③ 수정은 변이(mutant)로 "진짜 회귀는 여전히 잡는다"를 증명한다.

## 4. 범위

**안**: `internal/sessionmsg/store_test.go` 종료 규칙 + 그 결정론적 재현 테스트, `internal/timing` 보정 통계(공유 라이브러리 — 모든 AssertPaired 호출자에 적용) + 속성 테스트 2종(기존 load-step 핀 유지 + 교대 비대칭 신규 핀), `internal/hook/config_change_test.go` 단언 재구성, `.moai/reports/t278/` 산출물 일체.

**밖**: store.go 본체 변경(§2.1에서 결함 제외 — 조사가 뒤집으면 그때 보고), 무관 테스트 경화, CI workflow 편집(관측만), BenchmarkConfigChange_AsyncReturn 변경, timing의 다른 두 arm(SteadyCeiling/Budget) 재조정, TestBranchGuard_Latency 등 다른 AssertPaired 호출자 테스트의 본체(공유 통계 변경의 수혜는 따라온다).

## 5. 요구사항 (GEARS)

도메인 토큰: **CFS**.

- **REQ-CFS-001** (Event-driven) — **When** a poller in `TestConcurrentSendPoll` observes an empty drain (`len(res.Messages) == 0 && res.Remaining == 0`), the test shall treat that observation as terminal only if a poll **started after all senders completed** also observes an empty drain; observing `sendersDone` closed after an empty poll shall trigger a re-poll whose collected result decides termination (관측 이후 창이 닫히는 TOCTOU 제거 — 긍정 조건 대기).
- **REQ-CFS-002** (Event-driven) — **When** the historical early-exit interleaving is forced deterministically (channel-handshake: 0/0 poll 관측 → 송신측 잔여 3건 발신 + `sendersDone` close → poller의 종료 확인 재개), a regression test shall fail on the pre-fix stop rule (97/100) and pass on the post-fix rule (100/100) — 재현 불가 환경(CI)의 인터리빙을 인위적 순서 강제로 로컬 판정 가능하게 핀.
- **REQ-CFS-003** (Event-driven) — **When** the paired measurement window exhibits order-alternating scheduling asymmetry (교대 위상고정 노이즈), the calibrated arm of `AssertPaired` shall not report a regression on byte-identical ref/fn — 핀 테스트가 합성 분포로 이 속성을 강제한다.
- **REQ-CFS-004** (Event-driven) — **When** fn is uniformly ≥ `MaxUnits`× the reference cost (참 회귀), the calibrated arm shall still fail — 검출력 보존의 변이 증명.
- **REQ-CFS-005** (Ubiquitous) — The calibrated-arm statistic shall continue to survive the offset load-step property already pinned by `TestCalibratedRatioSurvivesOffsetLoadStep` (교대 비대칭 해소가 load-step 취약성 재발로 되돌아가는 것 금지).
- **REQ-CFS-006** (Ubiquitous) — The statistic choice shall be recorded in a decision document citing measured data (초록 실행의 job-summary ratio 분포 + 합성 분포 단위 테스트), not asserted a priori.
- **REQ-CFS-007** (Event-driven) — **When** `TestConfigChange_RT005ReloadIntegration` enforces the REQ-HAE-002 latency contract, it shall evaluate a noise-robust multi-sample statistic aligned with the contract's p95 semantics; single-sample wall-clock maximum enforcement is prohibited — and **when** the sync path carries a true ≥100ms regression (변이: 동기 경로 sleep), the assertion shall still fail.
- **REQ-CFS-008** (Ubiquitous) — 지속시간·비(ratio) 단언을 내리는 측정 테스트 — `timing` 패키지에서 "Deliberately NOT parallel — measuring test" 자가선언 마커를 가진 테스트(현행 5종: `TestMeasurePairedEqualSampleCounts`, `TestAssertPairedHealthyEndToEnd`, `TestMeasureCalibratedRatioHealthy`, `TestMeasureCalibratedRatioTripsAt4x`, `TestAssertHealthyEndToEnd`) — shall not run with `t.Parallel()`, and 신규 측정 테스트는 같은 마커를 달아 이 규칙의 집합에 명시적으로 합류한다. 기존 `timing_test.go:14` HARD 규칙의 명문화. **면제**: tick-양성(tick-positivity)·호출-횟수만 단언하는 테스트(예: `TestMedianRunsWarmupPlusSamples` — cpuUnit을 호출하나 측정값의 양수성만 검사, 부하에 강건)는 이 규칙의 대상이 아니다. 판정 명령은 AC-CFS-010.
- **REQ-CFS-009** (Ubiquitous) — All CI occurrence counting (baseline과 사후 관측 모두) shall be attempt-aware: every workflow run's **every failed attempt** is inspected (`gh run view <id> --attempt <n> --log-failed`), because rerun-green hides attempt-1 failures from `gh run list`.
- **REQ-CFS-010** (Ubiquitous) — The completion verdict shall rest on observed CI evidence: post-merge, **≥40 attempt-aware observations** — 관측 단위 N = `go_code=true`인 **workflow run** 의 수(각 run에서 affected 두 job을 모두 열거; job-인스턴스 수가 아님 — 신뢰도 산술 `(1-p̂)^N`의 N도 이 정의) — AND **≥7 calendar days**, with zero recurrences of the three tests, recorded with run IDs; local darwin passes are necessary but not sufficient. 잔여 위험(통계적 검정력)은 verdict에 숫자로 명시한다.
- **REQ-CFS-011** (Event-driven) — **When** run-phase begins, the pre-fix occurrence count shall be measured over a named window (attempt-aware sweep) and recorded — the verdict's denominator and confidence arithmetic require it.
- **REQ-CFS-012** (Event-driven) — **When** the fixes land, a series analysis shall extract the common factor (§3) with the three cases as instances and name the reusable authoring rule, landing at `.moai/reports/t278/series-analysis.md`.

## 6. 위험과 중요 결정

- **R1 — 통계 변경의 파급**: AssertPaired는 `TestBranchGuard_Latency`(internal/hook) 등 다른 호출자가 있다. REQ-CFS-005의 핀 테스트가 퇴행을 막는다. M1이 호출자 전수를 조사한다.
- **R2 — AND-gate 후보의 검출력 공백**: 두 추정량이 모두 넘어야 실패하는 형태(교대 비대칭과 load-step에 각각 면역)는 leading option이나, 한쪽만 움직이는 참 회귀(비현실적 형태)를 놓친다 — 잔여 위험으로 기록. M1 데이터가 다른 형태를 지지하면 기각.
- **R3 — 재현율 검정력**: baseline 발화율이 낮아(관측 4건/수십 run) 0-in-N의 통계적 확신은 유한하다. REQ-CFS-010은 관측된 증거 + 명시적 잔여 위험(verification-claim-integrity 5-절 보고)으로 판정한다 — 허위 확신 금지.
- **R4 — 부하 규율**: 로컬 검증은 3 패키지 한정·소수 -count만(계획 §C). 풀스위트는 CI가 판정한다.

## 7. 완료 정의 (요약)

acceptance.md의 AC 전부 GREEN — 특히 AC-CFS-007(재현율 verdict, 관측된 CI 증거)과 AC-CFS-008(계열 보고서). `moai spec audit` 무관련 위반 0. 산출물: 수정 4-5 파일 + `.moai/reports/t278/` 4종(forensics.md, timing-statistic-decision.md, reproduction-rate.md, series-analysis.md).
