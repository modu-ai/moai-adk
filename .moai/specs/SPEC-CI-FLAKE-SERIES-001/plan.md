# Plan — SPEC-CI-FLAKE-SERIES-001

## §A 맥락

카드 t278 (t270·t271 흡수). 작업 트리 `.claude/worktrees/t278`, 브랜치 `WT-ci-flake-series`, base `175d63f3f`. plan-phase에서 수행한 로그 포렌식·코드 정독의 결과와 뒤집힌 전제는 `spec.md` §2에 고정됐다 — run-phase는 그 조사를 **다시 하지 않고** §C 사전점검으로 현행성만 확인한 뒤 시작한다.

### Tier 판정 — M

| 축 | 값 | Tier M 상한 | 여유 |
|---|---|---|---|
| REQ | 12 (REQ-CFS-001..012) | 16 | 4 |
| AC | 9 (AC-CFS-001..009) | 16 | 7 |
| 마일스톤 | 4 | — | — |

- **S 아님**: 3개 패키지 + 공유 라이브러리(internal/timing) 통계 변경 + CI 관측 의존 완료 판정 + 보고서 4종. 단일 파일 편집이 아니다.
- **L 아님**: 파일 6-8개, 새 아키텍처 없음, 되돌리기 어려운 인터페이스 없음, constitutional 범위 아님.

### 카드 등급

배차 등급 Class B(결함, 원인 미상). plan-phase 조사로 세 메커니즘 중 2개는 확정·1개는 후보 2계열로 좁혔다(#2.1, #2.2). 통계 선택(REQ-CFS-006)에 데이터 의존 설계 결정이 남아 있으므로 SPEC을 세워 진행한다 — 속도가 아니라 결정의 존재가 근거다(결함이라 해도 회귀 방지 속성 테스트라는 설계 판단을 동반).

## §B 수정 설계 (M1 결정 전 확정된 부분)

### B1 — flake #1: 종료 규칙 긍정화 (`internal/sessionmsg/store_test.go`)

현재(`store_test.go:361-370`): 0/0 관측 → `select`로 `sendersDone` 확인(관측보다 나중 시각) → 닫혔으면 즉시 탈출.

수정: 0/0 관측 후 `sendersDone`이 닫혀 보이면 **재poll 1회**를 수행하고, 그 재poll(송신 전원 완료 이후에 시작 = 모든 rename 이후의 리스팅)도 0/0이어야 탈출한다. 재poll의 메시지는 수령 목록에 합산한다. 재수행 불가 창이 원천적으로 사라진다 — `close(sendersDone)`는 `wgSend.Wait()` 뒤에만 실행되므로 모든 Send 반환(따라서 모든 pending rename 완료) 이후이고, 그 이후에 시작한 poll의 빈 관측은 종언(terminal)이다.

결정론적 재현 테스트(REQ-CFS-002, 신규 파일 `internal/sessionmsg/stoprule_test.go` 또는 store_test.go 내): 채널 핸드셰이크로 치명 인터리빙을 강제한다 —

1. 송신 고루틴: 97건 Send → "sent97" 신호 → poller의 "sawEmpty" 신호 대기.
2. poller: 0/0에 도달하면 "sawEmpty" 신호 → 송신측 "done-closed" 신호 대기(select 직전에서 봉쇄).
3. 송신: 잔여 3건 Send + done close → "done-closed" 신호.
4. poller 재개: 구규칙이면 즉시 탈출(97 ≠ 100 RED), 신규칙이면 재poll로 3건 수령(100 GREEN).

동시성 원본 테스트(`TestConcurrentSendPoll`)는 불변식(100 유일·0손실·0중복) 그대로 두고 종료 규칙만 교체한다. sleep 없이 결정론적이다.

### B2 — flake #2: 보정 통계 (internal/timing) — M1 데이터가 선택

Leading option **AND-gate**: calibrated arm 실패 조건을 "per-round 비 중앙값 > MaxUnits **AND** ratio-of-medians > MaxUnits"로.

- 참 균질 회귀(3x): 두 추정량 모두 상회 → 실패 ✓ (REQ-CFS-004 변이로 증명)
- load-step(핀된 속성): ratio-of-medians만 상회 → 통과 ✓ (REQ-CFS-005)
- 교대 비대칭(2026-08-24 관측 형태): per-round만 상회(1.09x vs 2.47x) → 통과 ✓ (REQ-CFS-003)

Fallback option: per-round 비의 절사 중앙값(trimmed) + Iterations 증가 — 단 관측 형태는 라운드의 절반이 쏠린 것이라 절사로는 회복 불가 가능성이 높고(§1.2 판별자), M1 합성 분포 테스트가 기각하면 채택하지 않는다.

변경 위치: `AssertPaired`/`report`(`timing.go:215-238`)가 두 수치를 계산해 `CheckRatio` 호출부에서 AND 판정. 순수 함수 `CheckRatio` 시그니처는 가능한 한 유지(기존 핀 테스트 `paired_step_test.go:56`가 직접 호출). 호출자(`TestBranchGuard_Latency` 등) 전수 조사 결과를 M1 결정 문서에 기록.

속성 테스트 2종(신규, `internal/timing/paired_asym_test.go` 가칭):
- 교대 비대칭 합성 분포: 짝수 라운드 fn+δ/ref 기준, 홀수 라운드 반전 — 참 비 1.00x, per-round 중앙값은 관측 형태처럼 폭등. 신규 추정량은 통과, per-round 단독 추정량은 실패해야 한다(falsifier — 추정량이 데이터를 무시해도 통과하는 테스트가 되지 않기 위한 것, `paired_step_test.go:61-63`의 반위(反僞) 패턴 준용).
- 균질 3x 분포: 신규 추정량 실패(검출력).

### B3 — flake #3: 단언 재구성 (`internal/hook/config_change_test.go`)

`Handle`을 N=20회(핸들러 매회 신규 생성) 측정해 elapsed 표본 20개를 만들고 p95 ≤ 100ms를 단언한다. 각 표본은 Handle 1회 호출 후 `testutil.WaitForAsync(t, h.waitGroup(), 2*time.Second)`(기존 RT005 종결 패턴, `config_change_test.go:63`)로 해당 핸들러의 비동기 고루틴을 배수한 뒤 다음 표본으로 진행한다 — 표본 간 고루틴 누수·간섭 없이 20개 표본 전원이 종결된 상태에서 판정한다. nearest-rank p95(`timing.go`와 동일 의미론으로 로컬 계산 또는 기존 percentile 도구 재사용)는 20표본에서 두 번째로 큰 값 — 단일 선점 표본 1개가 p95를 침범하지 않는다(선점 2개 이상이 동시에 들어와야 하는데 그 결합 확률은 관측된 단발 형태보다 훨씬 낮다). `WaitForAsync` 종결 검증(긍정 대기)은 그대로 유지. `BenchmarkConfigChange_AsyncReturn`은 무변경(범위 밖).

변이: 동기 경로에 `time.Sleep(150ms)` 주입 시 p95 단언 RED — 검출력 증명 후 변이 제거.

## §C 사전 점검 (run-phase 진입 시 재측정 — 값이 다르면 멈춰 보고)

| 명령 | 기대값 |
|---|---|
| `git -C <worktree> rev-parse --short HEAD` | `175d63f3f` 기반 (이후 커밋이 base를 앞지르면 fetch 후 재확인) |
| `grep -c "publish(" internal/timing/timing.go` | ≥ 2 (정의 + 호출 — publish 회귀 여부) |
| `gh run view 32774108273 --attempt 1 --log-failed \| grep -c ConcurrentSendPoll` | ≥ 1 (증거 접근성 — 로그 만료 시 spec.md 인용으로 대체하고 M1 sweep이 재수집) |
| `go test -race -count=1 ./internal/timing/ ./internal/sessionmsg/` | PASS (baseline — 이미 초록이어야 함) |

## §D 마일스톤

### M1 — 조사 확정·baseline 측정·통계 결정 (수정 0건)

1. `.moai/reports/t278/forensics.md`: spec.md §1-§2의 증거(3 run attempt-1 로그 발췌 + 행 번호)를 run-phase 트리 SHA와 함께 고정.
2. **attempt-aware baseline sweep** (창: 2026-08-10 ~ run-phase 시작일): `gh run list --workflow ci.yml -L <창>` → 각 run의 attempts 열거(`gh api repos/modu-ai/moai-adk/actions/runs/<id>/attempts`) → 실패 attempt마다 `--log-failed`에서 3 테스트명 grep → 발화 수·run ID·job·시각 표. 산출: `.moai/reports/t278/reproduction-rate.md`(baseline 절).
3. **초록 ratio 분포 수집**: job summary(GITHUB_STEP_SUMMARY)의 `paired-cpu-1x` 라인 ≥10개 — 웹 summary 접근이 불가하면 대안: 해당 job 재실행 없이 과거 초록 run의 summary URL 목록을 run ID와 함께 기록(수동 열람분 인용) 또는 Race Test job에서 직접 측정 1회(`go test -race -run TestAssertPairedHealthyEndToEnd -count=10 ./internal/timing/` — 단일 패키지·부하 규율 내)로 로컬 참조치만 확보하고 CI 초록 분포는 AC-CFS-007 관측 창에서 지속 수집(채널 격차를 결정 문서에 명시).
4. **통계 결정**: B2의 AND-gate vs fallback을 1-3 데이터로 판정. 산출: `.moai/reports/t278/timing-statistic-decision.md`(인용 데이터 + 호출자 전수 + 기각 사유).
5. AssertPaired 호출자 전수 조사(`grep -rn "AssertPaired(" --include="*_test.go" internal/`).

종료 조건: 산출물 4-5건 확정 + §C 사전점검 통과.

### M2 — 수정 (RED-first, flake#1 → #3 → #2 순서)

1. **#1**: 결정론적 재현 테스트 작성 → 구트리에서 RED 확인(97/100) → 종료 규칙 수정 → GREEN(100/100) → 기존 동시성 테스트 무변식 통과 확인.
2. **#3**: p95 재구성 → 변이(150ms) RED 확인 → 변이 제거 후 GREEN.
3. **#2**: M1 결정에 따라 통계 변경 → 속성 테스트 2종(교대 비대칭 falsifier 포함) → 기존 `TestCalibratedRatioSurvivesOffsetLoadStep` 통과 확인 → `TestAssertPairedHealthyEndToEnd` 통과.
4. 패키지별 로컬 검증(§E).

종료 조건: AC-CFS-001..006의 로컬 판정 전부 GREEN + 변이 증명 기록.

### M3 — PR + CI 관측 창 개시

manager-git 경유 PR(제목에 t278). PR 브랜치 push 시 두 job 관측. 머지 후 **관측 창**: attempt-aware sweep을 주기적 수행(최소 일 1회 + 알려진 머지 직후), `.moai/reports/t278/reproduction-rate.md`의 post 절에 run ID별로 적립.

종료 조건: 머지 + 관측 기록 개시(창 완료는 M4).

### M4 — 계열 보고서·verdict·종결

1. `.moai/reports/t278/series-analysis.md`(공통 인자 §3의 확정판 + 재사용 저작 규칙: "시간·스케줄링 의존 판정은 단일 관측으로 내리지 않는다 — 긍정 조건·계약 통계·변이 증명").
2. `reproduction-rate.md` 완결: baseline 대비 + REQ-CFS-010의 N≥40(**N = go_code=true workflow run 수** — spec.md v0.2.0 D2 고정 정의, job-instance 아님)·≥7일 달성 여부 + 신뢰도 산술((1-p̂)^N) + Gaps/Residual-risk.
3. sync-phase(CHANGELOG/README 필요시) + sync-audit + 종결.

종료 조건: AC 전부 GREEN (AC-CFS-007·008 포함) — 단 관측 창 부족 시 PENDING 상태로 잔여 일수를 명시하고 판정 보류(허위 완료 선언 금지).

## §E 검증 명령 (부하 규율 — 풀스위트 금지, 3 패키지 한정)

```bash
# sessionmsg — 단일 패키지, 기존 acceptance 관례 -count=20
go test -race -count=20 ./internal/sessionmsg/
# timing — 측정 테스트 밀집, -count=5
go test -race -count=5 ./internal/timing/
# hook — 패키지 전체는 -count=1 1회 + 대상만 -count=5
go test -race -count=1 ./internal/hook/
go test -race -count=5 -run 'TestConfigChange' ./internal/hook/
```

CI 전체 스위트 판정은 PR의 두 job에 맡긴다(로컬 풀스위트 금지 규율). 모든 로컬 실행은 워크트리 안에서.

## §F 산출물 경로

| 산출물 | 경로 |
|---|---|
| forensics 고정 | `.moai/reports/t278/forensics.md` |
| 통계 결정 기록 | `.moai/reports/t278/timing-statistic-decision.md` |
| 재현율 baseline+post | `.moai/reports/t278/reproduction-rate.md` |
| 계열 분석 | `.moai/reports/t278/series-analysis.md` |
| SPEC 4종 | `.moai/specs/SPEC-CI-FLAKE-SERIES-001/{spec,plan,acceptance,progress}.md` |

## §G 리스크 대응

| 위험 | 대응 |
|---|---|
| M1이 spec.md §2.1의 store-무결점 판정을 뒤집는 증거를 얻으면 | 수정 대상이 store.go로 이동 — REQ-CFS-001/002를 그에 맞게 개정하고 progress.md에 전제 뒤집힘을 기록 후 사용자 보고 |
| AND-gate가 초록 분포 데이터와 충돌 | fallback 또는 신규 옵션 — 결정 문서에 기각 사유와 함께 기록 |
| 관측 창(≥7일)이 카드 주기를 넘긴다 | M4를 PENDING으로 두고 종결 조건만 명시 — 관측은 무인 sweep(스크립트)로 지속 가능 |
| CI 로그 만료(90일) | baseline sweep은 run-phase 초반에 우선 수행 |
