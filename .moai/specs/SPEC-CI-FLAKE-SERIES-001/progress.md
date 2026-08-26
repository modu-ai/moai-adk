# Progress — SPEC-CI-FLAKE-SERIES-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-26T16:10+09:00
- plan_verdict: PASS-WITH-DEBT 0.86 (Tier M threshold 0.80) — plan-audit iter-1, report `.moai/reports/t278/plan-audit-iter1.md`
- should_fix_resolution: D1–D5 all landed in SPEC v0.2.0 — D1/D2/D3/D4 authored by manager-spec; the final plan.md edits (D2 M4 consistency + D5 B3 wg specification) were transcribed by orchestrator lane-12 after manager-spec hit a backend usage-limit 429 mid-edit, strictly per the author's own HISTORY 0.2.0 rows (content decisions remain manager-spec's)
- kickoff_approval: GRANTED 2026-08-26 (operator; option "승인 — 리셋 후 착수"; progression mode = semi-autonomous)
- spawn_deferral: manager-develop spawn deferred to backend usage-limit reset (2026-08-26 17:22) via scheduled 17:25 wakeup — 429 killed the first re-delegation attempt

## §F Phase 4 Mode Selection

Input parameters:

- tier: M
- scope: 4–5 implementation files (store_test.go + new stoprule test, timing.go + 2 property tests, config_change_test.go) + report artifacts
- domain count: 1 (Go test/library code)
- file language mix: Go + markdown
- concurrency benefit: LOW (coding-heavy; sequential milestone dependency M1 → M2 → M3)
- Agent Teams prereqs: not operator-requested

Mode evaluation:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | non-trivial code changes, RED-first discipline required |
| serial | **SELECTED** | coding-heavy work; Anthropic coding-task parallelism caveat; single manager-develop carries M1→M3 |
| fanout | no | single-domain implementation, no research fan-out |
| sweep | no | not high-volume mechanical transformation |
| agent-team | no | experimental, explicit-request-only |

Decision: serial

Justification: implementation is coding-heavy with strictly sequential milestone dependencies (statistic decision in M1 feeds the M2 fixes feeds the M3 PR). Progression mode: **semi-autonomous** — orchestrator checkpoints with the operator at M1 end (statistic decision) and M2 end (fixes + local verification), per the kickoff approval.

## §E.2 Run-phase Evidence

### M1 — 조사 확정 · baseline 측정 · 통계 결정 (수정 0건, 2026-08-27, tree `d1289c5db`)

사전점검 (plan §C, 전부 이번 실행·이 트리에서 관측):
- `git rev-parse --short HEAD` → `d1289c5db` (branch `WT-ci-flake-series`) — 기대값 일치
- `grep -c "publish(" internal/timing/timing.go` → `2` (정의 timing.go:254 + 호출 timing.go:234) — ≥2 충족
- `gh run view 32774108273 --attempt 1 --log-failed | grep -c ConcurrentSendPoll` → `1` — 증거 접근성 확인
- `go test -race -count=1 ./internal/timing/ ./internal/sessionmsg/` → `ok internal/timing 1.798s` / `ok internal/sessionmsg 2.032s` — baseline 초록

산출물 (커밋됨):
- `.moai/reports/t278/forensics.md` — 3 flake 증거 로그 발췌(run ID + 출력 라인 번호)를 run-phase 트리에서 재확보해 고정. 신규: pre-#1591 발화 2건(32429213275 a1/a2, 양측 2.72x) 발견·분리 계정, `.../runs/<id>/attempts` REST 404(537/537) 실측, RT005 INFO 노출 노이즈(7 run, `--- FAIL:` 아님) 계류 정리.
- `.moai/reports/t278/reproduction-rate.md` — baseline 절: 창 2026-08-10~08-26, 537 run(535 go_code=true 관측 + 2 cancelled), 19 multi-attempt run / 556 attempt 검사, 현재-결함 발화 4건(#1591 이후 166 run 기준 p̂=4/166≈0.0241, 테스트별 2/1/1), 검정력 산술 ((1-p̂)^40≈0.377 — N≥40이 증명이 아님을 숫자로 명시). 로컬 초록 ratio 참조분포 10회(0.99–1.01x) + CI 요약 채널 격차 기록.
- `.moai/reports/t278/timing-statistic-decision.md` — **AND-gate 채택** (per-round 중앙값 AND ratio-of-medians 모두 MaxUnits 초과 시 실패). 근거: post-#1591 유일 발화가 2.47x/1.09x 형태, 로컬 초록 양측 1.00x, load-step 핀(per-round 1.00x)·균질 4x 핀(양측) 모두 호환, pre-#1591 양측 형태는 여전히 검출. fallback(절사+반복증가)은 산술적으로 기각 — 20 라운드 중 ≥11개 쏠림은 20% 절사까지 불변, 45%+ 절사만 이동시키나 검출력 파괴. 호출자 전수 3곳(paired_test.go:61, observer_test.go:251 `TestRecordEvent100Sequential`, pre_tool_branch_guard_integration_test.go:207 `TestBranchGuard_Latency`).
- `.moai/reports/t278/sweep-attempts.sh` (v2) + `refetch-jobs.sh` — attempt-aware sweep 스크립트(재실행 가능 판정 명령, AC-CFS-007).

M1 데이터가 spec 전제에 대한 정밀화 (E5 상세는 최종 보고):
- plan §D-M1 item 2의 `gh api .../attempts` 경로는 이 repo에서 404 — `run_attempt` 필드 + `gh run view --attempt N --json`으로 대체 관측(스크립트 v2에 주석 고정). spec §2.5 서술의 정정 후보.
- 대상-외 확산 증거: 32687843472 a1(8/24)에서 `TestBranchGuard_Latency` 1.82x 양측 FAIL + 같은 run Race Test에서 `TestReadCardStatus_DoesNotSearchBranchSet` FAIL — 3종 밖 flake 계열 존재. 본 SPEC 범위 외이나 M4 계열 분석 맥락으로 기록.

미해결 GAP: CI 초록 ratio의 기계적 수집 불가(채널 격차 — GITHUB_STEP_SUMMARY는 웹 전용, 초록 패키지의 t.Log는 비-verbose go test가 폐기). AC-CFS-007 관측 창에서 웹 summary 수동 열람분을 지속 기록한다.

### M2 — 수정 (RED-first, #1 → #3 → #2 순서, 2026-08-27, tree `def99739d` → `5aedc1cd3` → `324883ebb`)

사전점검 (Section C, 전부 이번 실행·이 트리 `0239bddd4`에서 관측): `git rev-parse --short HEAD` → `0239bddd4` ✓ · `go build ./...` → exit 0 ✓ · `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 ✓ · `go test -race -count=1 ./internal/timing/ ./internal/sessionmsg/ ./internal/hook/` → `ok internal/timing 1.828s` / `ok internal/sessionmsg 1.829s` / `ok internal/hook 40.218s` (baseline 초록) ✓ · `golangci-lint run --timeout=2m` → `0 issues` (baseline — NEW-vs-preexisting 분류 기준) ✓

**Fix #1 — poller 종료 규칙 (commit `def99739d`; stoprule_test.go 신규 + store_test.go)**

종료 규칙을 `pollUntilDrained` 헬퍼로 추출해 `TestConcurrentSendPoll`과 재현 테스트 `TestPollerStopRule`이 같은 코드를 실행 (규칙이 한 곳에만 존재 → mutant probe가 실제 테스트의 규칙을 검사). 핸드셰이크 게이트(`beforeStopCheck` 훅)는 0/0 관측 직후·sendersDone 확인 직전에 poller를 hold — CI 경합 창을 결정론적으로 강제.

- RED (구규칙, E8 verbatim): `go test -race -count=1 -run 'TestPollerStopRule' -v ./internal/sessionmsg/` →
  `stoprule_test.go:117: received 97 messages, want 100 (stop rule treated a pre-completion empty observation as terminal)` / `--- FAIL: TestPollerStopRule (0.17s)` — CI 관측(run 32774108273 a1)과 동일한 97/100.
- GREEN: 규칙 수정(REQ-CFS-001 — 닫힌 채널 관측 시 재poll 1회, 그 재poll의 0/0만 종언, 수령 합산) 후 `go test -race -count=3 -run 'TestPollerStopRule|TestConcurrentSendPoll' -v` → 두 테스트 3회 연속 PASS (100/100).
- Mutant probe (AC-CFS-001): 규칙을 구형(select-즉시-탈출)으로 1회 되돌림 → `--- FAIL: TestPollerStopRule (0.17s): received 97 messages, want 100` 재관측 → 복원 → `go test -race -count=3 -run 'TestPollerStopRule|TestConcurrentSendPoll' ./internal/sessionmsg/` → `ok 2.291s` (GREEN 재확인) + `go vet ./internal/sessionmsg/` → VET_OK.
- 커밋에 `status: draft → in-progress` frontmatter 전환 동반 (첫 run-phase 코드 커밋).

**Fix #3 — RT005 단언 p95 재구성 (commit `5aedc1cd3`; config_change_test.go만, production 무변경)**

20샘플(핸들러 매회 신규 생성)·샘플 간 `testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)` 배수·nearest-rank p95(20중 19번째 = 두 번째로 큰 값) ≤ 100ms 단언. 기능 불변식(Continue/SystemMessage)은 첫 샘플에서 1회 점검.

- Mutant RED (AC-CFS-006, E8 verbatim): 동기 경로에 `time.Sleep(150*time.Millisecond)` 주입 → `go test -race -count=1 -run 'TestConfigChange_RT005ReloadIntegration' -v ./internal/hook/` → `config_change_test.go:93: synchronous return p95 151.163958ms over 20 samples, want ≤ 100ms (REQ-HAE-002)` / `--- FAIL: (3.44s)`.
- Mutant 제거 후 GREEN: `go test -race -count=5 -run 'TestConfigChange' ./internal/hook/` → `ok 4.093s` (5/5) + `go vet ./internal/hook/` → VET_OK. `git diff --stat`로 production 파일(config_change.go) 잔여 변경 0 확인.

**Fix #2 — 페어링 보정 AND-gate (commit `324883ebb`; timing.go + paired_asym_test.go 신규)**

M1 결정(AND-gate) 구현: `AssertPaired` → `reportPaired`가 두 수치(per-round 중앙값 + ratio-of-medians)를 로그·보고하고 `CheckRatioAnd`(신규 순수 헬퍼)가 **양측 모두** MaxUnits 초과 시에만 calibrated arm 실패. `CheckRatio` 시그니처·의미 불변(기존 핀 paired_step_test.go:56/:79 무수정 통과) — p95/worst arm은 `checkAbsolute`로 공유 추출. 페어링 로그 라인에 두 수치 병기(`ratio=X.XXx per-round / X.XXx medians`).

- RED (현행 추정량, E8 verbatim): 신규 속성 테스트의 기대("AND-gate는 교대 비대칭을 통과")를 구형 단일 수치 추정량에 대고 관측 — `go test -race -count=1 -run 'TestCalibratedGate...|TestPairedAndGate...' -v ./internal/timing/` → `paired_asym_test.go:63: calibrated gate tripped on alternation-locked asymmetry: [measured latency is 2.50x the reference unit (median 400µs), above the 2.00x calibrated bound — ...]` / `--- FAIL: TestCalibratedGateSurvivesAlternationLockedAsymmetry (0.00s)` (균질 3x 검출 arm은 PASS — 현행 추정량의 검출력 기준점).
- 속성 테스트 (합성 분포 — 측정 아님, paired_step_test.go 핀과 동일 부류로 t.Parallel):
  - (a) 교대 위상고정 비대칭(참 비 1.00x, 절반 fn-우위 2.50x/절반 ref-우위 0.40x → per-round 중앙값 2.50x·medians 1.00x): `CheckRatioAnd` 오류 0건 통과 **+ falsifier arm** — 구형 `CheckRatio`(per-round 단독)가 동일 분포에서 calibrated 오류 1건 냄을 테스트가 상시 단언(반위 패턴, paired_step_test.go:61-63 준용).
  - (b) 균질 3x(load 변동 하 전 라운드 3배): `CheckRatioAnd` 여전히 calibrated 오류 1건 (REQ-CFS-004 검출력 보존).
- GREEN (핀 전체 + 호출자 3곳, 전부 이번 실행·이 트리): `TestCalibratedRatioSurvivesOffsetLoadStep`·`TestPairedRatioStillCatchesRealCostGrowth` 무변경 통과; `TestAssertPairedHealthyEndToEnd` PASS (`ratio=1.00x per-round / 1.12x medians`); `internal/harness -run TestRecordEvent100Sequential` PASS (`0.99x / 0.98x`); `internal/hook -run TestBranchGuard_Latency` PASS (`1.00x / 1.01x`, maxUnits 1.50x).

**§E 로컬 검증 배터리 (plan §E, 최종 트리 `324883ebb`에서, 전부 이번 실행)**

- `go test -race -count=20 ./internal/sessionmsg/` → `ok 57.477s` ✓ (AC-CFS-002)
- `go test -race -count=5 ./internal/timing/` → `ok 3.146s` ✓
- `go test -race -count=1 ./internal/hook/` → `ok 168.898s` ✓
- `go test -race -count=5 -run 'TestConfigChange' ./internal/hook/` → `ok 4.216s` ✓
- `go build ./...` → exit 0 ✓ · `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 ✓
- `golangci-lint run --timeout=2m` → `0 issues` (baseline 0 → NEW 0) ✓
- `go vet ./internal/{sessionmsg,hook,timing}/` → 각 VET_OK ✓

**AC-CFS-010 판정 (이번 실행)**: timing 측정 테스트 5종(마커 보유) 전부 t.Parallel 없음 — VIOLATION 0건 (awk func-boundary 스캔 + `grep -c "Deliberately NOT parallel"` = paired_test.go 2 + timing_test.go 3 = 5). 신규 timing 테스트 2종은 합성 분포 순수 판정(측정 없음)으로 마커 집합 합류 대상 아님 — REQ-CFS-008의 면제 논리(tick-양성·호출-횟수·순수 통계 테스트)와 동일 근거. hook p95 테스트는 측정하나 REQ-CFS-008의 선언 범위(timing 패키지 마커 집합) 밖이고, p95의 계약 자체가 "병행 부하 하 p95"이므로 t.Parallel 유지가 수정 의도와 일치.

커밋 목록 (브랜치 WT-ci-flake-series, push 없음 — B9 override):
- `def99739d` fix(SPEC-CI-FLAKE-SERIES-001): M2 stop-rule re-poll closes poller TOCTOU (t278)
- `5aedc1cd3` fix(SPEC-CI-FLAKE-SERIES-001): M2 RT005 latency assertion rebuilt as 20-sample p95 (t278)
- `324883ebb` fix(SPEC-CI-FLAKE-SERIES-001): M2 paired calibrated arm enforces AND-gate (t278)

