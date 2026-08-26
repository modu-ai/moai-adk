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

