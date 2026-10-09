# SPEC-MERGE-WINDOW-QUEUE-002 — Progress

> Card t1582 · created 2026-10-09 by manager-spec (plan phase) · branch `WT-p1-p2-t1576` · base develop tip `f7606c7bc`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md + plan.md + acceptance.md (Tier M set) + decision-index.md (`interview.decision_gate: on`) + progress.md.
- SPEC id regex check (Bash) → `PASS`; uniqueness: `ls .moai/specs/ | grep -i MERGE-WINDOW` → SPEC-MERGE-WINDOW-QUEUE-001 만 존재 (002 자유).
- RED 재현 (plan 단계 실측, 트리 `f7606c7bc`, env-scrub 복합 형태):
  - ④-R7-2: `TestRedT1582MixedSweepWithNoTestPackageCounts` — `runner reported [no test files] — an empty sweep cannot stand for a re-measure` (거짓 거부 관측).
  - ④-R7-3 단위: `TestRedT1582VerifierAdmitsAShortBaseSHA` — 검증기가 3바이트 Base 통과 관측.
  - ④-R7-3 e2e: `TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease` — `runtime error: slice bounds out of range [:12] with length 3`, release 전 사망 + 창 held 관측.
  - item ②: round-trip 관측 — 경계 붕괴 **미재현**(4/5 일치, apostrophe는 escape 바이트 잔류·경계 유지) → 봉인 테스트 GREEN.
  - item ①: probe 관측 — non-test 명령의 실패 의미 출력이 exit 0으로 valid (spec.md §D 잔여, 수리 대상 아님).
  - item ③·R7-1: 재현 설계 확정 — 관측은 run 단계 M3/M4 첫 행위 (plan.md §F).
- RED 오버레이 테스트 파일: `internal/factory/remeasure_red_t1582_test.go`, `internal/cli/shelljoin_red_t1582_test.go` (트리에 작성, 미커밋 — run 단계 M1-M4가 GREEN으로 뒤집는 대상).
- 10 REQ / 8 AC (v0.2.0 — D5 분할: 002/008, 003/009, 006/010). Out of Scope 경계: 보호구역·todo 발행·nominate(리더 큐), non-test 명령 잔여, 호출자 리다이렉션, `merge gate` verdict 설계, complete의 release-failure 출력.
- plan-audit iteration 1: FAIL 0.69 / Tier M 0.80 (`.moai/reports/t1582/plan-audit.md`) — D1-D6 아티팩트 수리 완료(v0.2.0, plan.md·acceptance.md·spec.md).
- plan-audit iteration 2: FAIL 0.85 (동일 파일 Iteration 2 섹션) — 라운드 2 수리 완료(v0.3.0): D19(M1 렌더 12곳 실측 열거+총괄 규칙+cli 스코프 — 감사 prose "14"는 자기 목록 10+2의 오산술, 리더 독립 grep과 본인 측정 일치)·D13(§4 목록 교체)·D14(cause-7 시딩 설비 `internal/factory/mergestep_red_t1582_test.go` + merge-ready RED `internal/cli/factory_merge_ready_red_t1582_test.go` plan 단계 작성·실측, AC-005/006 셀 verbatim 기록, AC-006 실행 계수 가드 `--- PASS` ≥ 2)·D15(스크럽 변수 3종 전체 목록 명기+축약형 전면 제거)·D17(DoD tracked+untracked 수집 대조)·D20(GNU grep `-r`+exit 구분)·D18(version 0.3.0 정합). RED 총 4건 실측: R7-2·R7-3 단위/e2e·merge-ready ③.
- 429 재개 기록: 라운드 2 중 429(요청 한도) 2회 — 트랜스크립트 재개로 잔여 목록(디스크 상태 대조) 수행, 부분 상태는 위 행들이 증언.
- plan_status: audit-ready + ceiling-exception approved (§G 참조)

## §G Plan-audit Ceiling Exception Record

- 판정 궤적: iteration 1 FAIL 0.69 → 2 FAIL 0.85 → 3(천장) FAIL **0.91** — 단조 상승, STOP 신호 없음. 영수증 3건: `rcpt-a221eb42d027db89e2c7fac2`·`rcpt-33b7845356c311a4cddf2a1c`·`rcpt-7b69c79f30c6981fbbcc0aa1` (codex required 게이트 3/3 응답, 매 라운드 축소하는 형식 지적).
- 최종 잔여: acceptance.md 형식 hunk 2건(D21 DoD 허용 목록 내부 모순·D22 RED stdout 장부) — 코드 결함 0, 코드 변경 0줄.
- **운영자 결정: 천장 예외 승인·run 진입** — 근거 "0.91 단조·STOP 없음·잔여 형식 2건 수리 완료·코드 0줄". 리더 경유 전달(2026-10-09 21시경, msg), lane-26 기록.
- 판정 후 조치: D21·D22 hunk를 감사자의 required-fix 명세대로 수리(acceptance.md 20:30, §D.3 증거 장부 E-LEDGER-001~004 신설 포함) — 천장으로 재감사 없음, 본 행이 그 상태의 기록.
- decision record: decided_by=operator+leader evidence_refs=.moai/reports/t1582/plan-audit.md iter-3 ladder_path=리더-경유 운영자 결정(keep-set 인접 게이트의 정식 처분) — Kickoff 재개.

## §F Phase 4 Mode Selection

- Input parameters: tier=M, scope=4 소스 파일+4 RED 오버레이+테스트 재작성 1(internal/factory 2·internal/cli 2), domain count=1(Go CLI/factory 표면), file language mix=Go 100%, concurrency benefit=LOW(coding-heavy — 마일스킨 의존: M1 검증기→M2 판정→M3 exit→M4 원자화 순차), agent-team prereqs=미요청.
- Mode evaluation: direct 미선정(다중 파일·다중 마일스톤), fanout 미선정(coding-heavy — Anthropic 병렬화 주의), sweep 미선정(기계 균일 변환 아님), agent-team 미선정(명시 요청 없음).
- Decision: `serial`
- Justification: 단일 표면의 순차 의존 수리 패킷 — 마일스톤 M1→M4가 서로의 산출 위에 서고 같은 파일군을 건드리므로 단일 manager-develop 위임이 재위임 위험과 쓰기 경합을 최소화한다(Tier M 전체 Section A-E 템플릿 적용).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §J Lane run-entry record (card t1582, run tmnboq, lane-5)

- Lease: `factory next --card t1582` was refused once with `serial-slot` (holder t1568 live until 2026-10-09T16:15:03Z, observed 16:08:21Z). Leader ruling bb3b2d04 (all run tmnboq cards parallel; withdraws the serial order 44610f49) was verified on disk at 16:33Z: `moai factory status` shows `mode=parallelizable` for the run. The lease was then granted; the card is `picked` and the tree was entered at 16:33Z.
- Watchdog first observation (16:33:03Z, no prior snapshot, fail-open): HEAD `8673c2a95`, integration window `free`, evidence mtime max 1791549653.
- RED baseline on the pinned tree `8673c2a95` (`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -v -run '^TestRedT1582' ./internal/factory/ ./internal/cli/`; raw log kept in the lane scratchpad): FAIL ×4 with the stated reasons — `TestRedT1582MixedSweepWithNoTestPackageCounts`, `TestRedT1582VerifierAdmitsAShortBaseSHA`, `TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease`, `TestRedT1582MergeReadyMeasurementFailureStillExitsZero`; PASS ×5 (control and sealed items). Committed before any implementation as `d7f4fcf0c`.
- Pre-spawn sync (orchestrator rule, lane-local): `git rev-list --count --left-right origin/main...HEAD` = `139 95` (diverged). The implementation spawn is HELD.
- Absorb probe (`git merge-tree --write-tree HEAD origin/main`, tree unchanged): exit 1. Conflicts in `.claude/rules/moai/workflow/context-window-management.md`, its template mirror, `internal/bugreport/spool.go`, and `internal/bugreport/spool_bump_serialize_test.go`. origin/main has 36 commits touching `internal/factory` or `internal/cli` that this branch lacks (the t1538 factory-recovery series): a cross-card overlap with this card's target packages.

decision record: decided_by=lane-5 (card-pick, run tmnboq) evidence_refs=card=t1582;class=C;mode=parallelizable@2026-10-09T16:33Z;prior_hold=t1568(lease expired 16:25:01Z);leader_ruling=bb3b2d04(supersedes 44610f49);pr=no-link;landed=none ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9.3)

wait record: id=w-t1582-20261009T1638Z waiting_on=leader reason=cross-card base decision — pinned base vs absorbing origin/main (4 conflicts + t1538 overlap); implementation spawn held recheck=one-shot 5 min (local 01:43 KST) plus standing cron 8a51a689 (:07/:27/:47)

decision record: decided_by=lane-5-watchdog evidence_refs=board:d-20261009T163954Z-0030;resolves:w-t1582-20261009T1638Z;baseline=8673c2a95;red=d7f4fcf0c;run_record=f9e272a7c;waiver=pre-spawn-139/95-impl-spawn-only;mode=parallelizable@20261009T1643Z ladder_path=step2-board-ruling-option-B
