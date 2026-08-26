# Acceptance — SPEC-CI-FLAKE-SERIES-001

측정 기준 트리: `.claude/worktrees/t278` @ `175d63f3f` (plan-phase 증거). 아래 RED 값은 이 트리와 CI attempt-1 로그(run ID 명시)에서 실제로 관측한 값이다. 각 AC는 **관측 가능한 출력**을 내는 명령으로 판정한다. 완료 축은 t261 verification-completeness: 검증은 관측으로, 추론으로 대체하지 않는다 — 특히 AC-CFS-007(CI 재현율)은 로컬 초록만으로 판정 불가다.

---

## §D AC 매트릭스

| AC | 요구사항 | RED (현재) | GREEN (목표) |
|---|---|---|---|
| AC-CFS-001 | REQ-CFS-001, 002 | 결정론적 재현 테스트가 구 종료 규칙에서 97/100 손실로 실패 | 재현 테스트 100/100 + 수정된 `TestConcurrentSendPoll` 무변식 통과 |
| AC-CFS-002 | REQ-CFS-001 | (연속성 축 — 현재도 로컬 초록) | `go test -race -count=20 ./internal/sessionmsg/` 연속 통과 |
| AC-CFS-003 | REQ-CFS-003, 005 | 합성 교대-비대칭 분포에서 현행 추정량(per-round 중앙값)이 오탐 실패 — 이 결함 자체가 단위 테스트로 핀됨 | 신규 추정량은 교대 비대칭 통과 **그리고** 기존 `TestCalibratedRatioSurvivesOffsetLoadStep`(load-step)도 통과 |
| AC-CFS-004 | REQ-CFS-004 | 균질 3x 합성 분포(mutant: `cpuUnit(6_000_000)`) — 현행 추정량 실패(검출됨) | 신규 추정량도 여전히 실패(검출 보존) |
| AC-CFS-005 | REQ-CFS-006 | 통계 결정 기록 부재 | `.moai/reports/t278/timing-statistic-decision.md`가 측정 데이터(§D.5 판정 명령)와 호출자 전수를 인용 |
| AC-CFS-006 | REQ-CFS-007 | 변이(동기 경로 `time.Sleep(150ms)`)에서 현행 단일-샘플 단언 실패 — 단, 무부하시 통과하는 것이 노이즈 취약 | p95 재구성 후: (a) 변이 주입 시 여전히 실패 (b) 변이 제거 시 `go test -race -count=5 -run 'TestConfigChange' ./internal/hook/` 통과 |
| AC-CFS-007 | REQ-CFS-009, 010, 011 | baseline 미측정(관측된 발화: 4건 — run 32774108273 a1, 32777242100, 32779472351 a1, 32815411885 a1) | `.moai/reports/t278/reproduction-rate.md`: (a) attempt-aware baseline 발화 수·분모(창 2026-08-10~) 측정 (b) 머지 후 N≥40(N = go_code=true workflow run 수 — §D.1 GREEN 조건과 동일 정의) 관측 **AND** ≥7일에서 3 테스트 재발 0 (run ID 목록) (c) 신뢰도 산술(같은 N 정의) + Gaps/Residual-risk |
| AC-CFS-008 | REQ-CFS-012 | 계열 보고서 부재 | `.moai/reports/t278/series-analysis.md` — 공통 인자 + 3 사례 대응 표 + 재사용 저작 규칙 명문화 |
| AC-CFS-009 | (범위 규율) | — | diff가 명명된 파일(`store_test.go`, `stoprule` 신규 테스트, `timing.go`, timing 테스트 2종, `config_change_test.go`)과 산출물 경로에만 한정 — 무관 테스트 경화 0건 |
| AC-CFS-010 | REQ-CFS-008 | 자가선언 측정 테스트 5종 전부 비병렬(현행 트리 실측: VIOLATION 0건 — 판정 명령의 RED는 mutant로 별도 관측 완료, §D.1) | 동일 판정 명령으로 VIOLATION 0건 유지 + 신규 측정 테스트의 마커 합류 |

---

## §D.1 AC 상세

### AC-CFS-001 — 종료 규칙 TOCTOU 제거 (결정론적 RED→GREEN)

**판정 명령** (워크트리 안):
```bash
go test -race -count=3 -run 'TestPollerStopRule' -v ./internal/sessionmsg/   # 신규 재현 테스트(명칭은 run-phase 확정)
go test -race -count=3 -run 'TestConcurrentSendPoll' -v ./internal/sessionmsg/
```

**RED 근거 (이미 관측됨)**: CI run 32774108273 attempt 1 — `store_test.go:385: received 97 messages, want 100` (0.07s, 에러 0건). 로컬에서는 스케줄링 창이 열리지 않아 재현되지 않으므로, 재현 테스트는 채널 핸드셰이크로 그 인터리빙을 **강제**한다(plan.md §B.1). 수정 전 트리에서 재현 테스트가 97/100으로 실패함을 먼저 관측(RED-now)하고 수정 후 100/100(GREEN)을 관측한다.

**Mutant (규칙 회그 감시)**: 재생성 규칙을 구 형태(select-즉시-탈출)로 되돌리면 재현 테스트가 다시 RED가 됨을 1회 관측해 테스트가 규칙을 실제로 검사함을 증명(관측 기록을 progress.md에).

### AC-CFS-002 — sessionmsg 연속성

**판정 명령**: `go test -race -count=20 ./internal/sessionmsg/` → exit 0. (기존 테스트 주석이 명시하는 acceptance 관례 준수. 로컬은 필요조건일 뿐 — CI 판정은 AC-CFS-007.)

### AC-CFS-003 / AC-CFS-004 — 보정 통계의 이중 속성 (순수 단위 판정)

**판정 명령**: `go test -race -count=5 ./internal/timing/` — 신규 속성 테스트 2종 + 기존 핀 테스트 전부 포함.

- **교대 비대칭 합성 분포**: 참 비 1.00x, 라운드 절반이 fn-우위·절반이 ref-우위(2026-08-24 관측의 통계적 형태 — median(fn)/median(ref)≈1.09x인데 per-round 중앙값 ≥2.4x). 신규 추정량: 통과해야 함. **Falsifier arm**: per-round 단독 추정량은 이 분포에서 실패함을 같은 테스트가 증명(이 팔이 없으면 추정량이 데이터를 무시해도 통과).
- **균질 3x 분포** (`cpuUnit(6_000_000)` 또는 합성 3x 샘플): 신규 추정량 실패(검출). 현행 추정량도 실패함(RED-now 셀 — 검출력이 현재 있었고 유지됨을 보여주는 기준점).
- **기존 load-step 핀**: `TestCalibratedRatioSurvivesOffsetLoadStep` 변경 없이 통과 (REQ-CFS-005).

### AC-CFS-005 — 통계 결정 기록

**판정 명령**: `.moai/reports/t278/timing-statistic-decision.md` 존재 + 다음을 인용: (a) AND-gate 대안 비교표, (b) 측정 근거 — 로컬 `-race -count=10 ./internal/timing/ -run TestAssertPairedHealthyEndToEnd`의 ratio 분포와(또는) 확보된 CI 초록 summary 라인 수(채널 격차 명시), (c) `grep -rn "AssertPaired(" --include="*_test.go" internal/` 호출자 전수와 영향 판정.

### AC-CFS-006 — hook 단언 p95 재구성

**판정 명령**:
```bash
go test -race -count=5 -run 'TestConfigChange_RT005ReloadIntegration' -v ./internal/hook/   # 변이 전후 각 1회 + 최종
```
(1) 변이 주입(동기 경로 `time.Sleep(150*time.Millisecond)`) → 실패 관측. (2) 변이 제거 → 통과 관측. (3) 측정 방식이 다중 샘플 p95(≥20표본)임을 코드에서 확인 — 단일 벽시계 최댓값 형태가 남아 있으면 FAIL.

### AC-CFS-007 — 재현율 verdict (관측 의존 — 로컬 대체 불가)

**Baseline (RED-now 셀)**: 2026-08-24~25에 관측된 발화 4건 (각각 spec.md §1의 run ID·attempt·job·타임스탬프). run-phase M1이 창 2026-08-10~ 전체를 attempt-aware로 sweep해 분모(run 수)와 함께 확정한다.

**판정 명령 (sweep — 스크립트로 `.moai/reports/t278/`에 산출)**:
```bash
gh run list --workflow ci.yml -L <창> --json databaseId,conclusion
# 각 run: gh api repos/modu-ai/moai-adk/actions/runs/<id>/attempts → 실패 attempt마다:
gh run view <id> --attempt <n> --log-failed | grep -e 'TestConcurrentSendPoll' -e 'TestAssertPairedHealthyEndToEnd' -e 'TestConfigChange_RT005ReloadIntegration'
```

**GREEN 조건 (모두 관측값)**: (a) baseline 절 완성 (b) 수정 머지 후 **N≥40 — N은 `go_code=true` workflow run 수**(job-인스턴스 수가 아님; 각 run에서 `Test (ubuntu-latest)` + `Race Test` 양쪽 모두 attempt-aware 열거) **그리고** **≥7 역일** 경과 (c) 그 전 구간에서 3 테스트의 재발 0건(run ID 목록으로 적립) (d) 동일 N 정의로 `(1-p̂)^N` 신뢰도 산술과 Gaps/Residual-risk 명시. **재발 1건이면 GREEN 불가** — 해당 테스트의 수정이 재검토된다(진짜 회귀인지 동일 노이즈인지 로그로 판별 후 기록).

### AC-CFS-008 — 계열 보고서

**판정 명령**: `.moai/reports/t278/series-analysis.md` 존재 + 3 사례 각각 (단일 관측 / 침투 창 / 계약의 통계) 대응 표 + "시간·스케줄링 의존 판정의 3원칙(긍정 조건·계약 통계·변이 증명)" 명문화 포함.

### AC-CFS-009 — 범위

**판정 명령**: `git -C <worktree> diff --name-only <base>..<head>` — 경로가 spec.md §4의 "안" 목록(테스트/라이브러리 4-5 파일 + `.moai/reports/t278/` + SPEC 4종)에 속함. store.go 본체·CI workflow·무관 테스트 포함 시 FAIL (단 §G의 전제 뒤집힘 보고가 진행된 경우 예외로 기록).

---

## §E 증거 인용 색인

| 증거 | 위치 |
|---|---|
| #1 fast-job 발화 | run 32774108273 attempt 1, 2026-08-24T20:33Z — `store_test.go:385` |
| #1 race-job 재발 | run 32777242100, 2026-08-24T21:07Z (최신 attempt 실패 — attempt 열거로 재확인) |
| #2 발화 + 요약통계 | run 32779472351 attempt 1, 2026-08-24T21:35Z — `timing.go:233` 라인 전체 |
| #2 역사 관측 | `timing_test.go:18-24`, `paired_step_test.go:20-24` (2.32x/2.72x/4.64x, 1.82x @ run 32687843472 a1) |
| #3 발화 | run 32815411885 attempt 1, 2026-08-25T06:12Z — `config_change_test.go:51` |
| attempt-은폐 실측 | 최근 60 run `--status failure` 3건 vs 그 창의 플레이크 발화 4건 (spec.md §2.5) |

## §F 판정 요약 양식 (progress.md §E용)

각 AC는 판정 시점·명령·출력 요지·트리 SHA를 함께 기록한다. AC-CFS-007은 5절 보고(Claim/Evidence/Baseline-attribution/Gaps/Residual-risk) 형식으로만 완료 선언한다.
