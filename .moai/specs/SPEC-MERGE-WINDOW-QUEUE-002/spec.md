---
id: SPEC-MERGE-WINDOW-QUEUE-002
title: "Merge-window queue integrity packet 2 — short-base record panic before window release, no-test marker wrongful refusal, cause-7 hold/release atomicity, measurement-failure exit code"
version: "0.3.0"
status: draft
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: "internal/factory (integration_remeasure.go, integration_merge_step.go), internal/cli (integration_remeasure.go, factory_merge.go, factory_card.go)"
lifecycle: spec-anchored
tags: "integration-window, remeasure, merge-step, sha-validation, no-test-files, cause-7-hold, exit-code, card-t1582"
tier: M
card: t1582
related_specs: [SPEC-MERGE-WINDOW-QUEUE-001, SPEC-FACTORY-LANE-AUTONOMY-001]
---

# SPEC-MERGE-WINDOW-QUEUE-002 — 병합 창 큐 무결성 수리 패킷 2

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 |
|------|------|--------|------|
| 0.1.0 | 2026-10-09 | manager-spec | 최초 초안 (카드 t1582, SPEC-MERGE-WINDOW-QUEUE-001의 착지 후 게이트 인계 4건 수리). RED 재현: R7-2·R7-3은 plan 단계에서 오버레이 테스트로 실측 관측 완료, item ①·②·③·R7-1 관측/재현 설계는 plan.md §Research에 기록. 7 REQ / 8 AC. |
| 0.2.0 | 2026-10-09 | manager-spec | plan-audit iteration 1 (FAIL 0.69 / Tier M 0.80, `.moai/reports/t1582/plan-audit.md`, D1-D6 차단) 수리. D5(MP-2): REQ-MWQ2-002/003/006의 두 trigger→behavior 쌍 묶음을 분할 — REQ-MWQ2-002(접두 렌더)+REQ-MWQ2-008(panic 대신 오류 경로), REQ-MWQ2-003(스윕 판정)+REQ-MWQ2-009(빈 스윕 무효), REQ-MWQ2-006(측정 실패 비제오)+REQ-MWQ2-010(경합 verdict 유지); 범위 문장은 Out of Scope로 귀속. 10 REQ / 8 AC. D1-D4·D6은 plan.md·acceptance.md 수리. |
| 0.3.0 | 2026-10-09 | manager-spec | plan-audit iteration 2 (FAIL 0.85, D13-D15·D17·D19 차단 + D18·D20 optional) 수리. D19: M1(ii) 렌더 열거를 실측 12곳으로 확장(factory 10 + cli 2 — 감사 표기 "14"에 대한 실측 좌표는 plan.md M1) + 총괄 규칙 + cli integration_remeasure.go M1 스코프 편입. D13: §4 잔존 구 E6 목록을 AC-MWQ2-007 SSOT 참조로 교체. D14: cause-7 시딩 설비 + merge-ready 측정 실패 RED 오버레이 테스트를 plan 단계에서 작성·실측(`internal/factory/mergestep_red_t1582_test.go`, `internal/cli/factory_merge_ready_red_t1582_test.go` — 관측은 acceptance.md AC-005/006) + AC-006 실행 계수 가드. D15/D17/D20·D18도 acceptance.md·plan.md 수리. RED 총 4건 관측. |

## 1. 배경과 문제 정의

카드 t1576(SPEC-MERGE-WINDOW-QUEUE-001의 착지 카드)은 리뷰 라운드 6-7에서 게이트가 계속 재지적하는 3건(R7-1, R7-2, R7-3)을 본 카드의 원 범위 밖으로 인계했고, 라운드 1-4의 분기 표면 조사가 항목 ①②③을 추가했다. 이 SPEC은 그 인계를 수리한다. 4개 카드 항목의 성격:

- **④-R7-3 (High, 최우선)** — 흡수 기록(`Base`)이 짧은 토큰(예: `"bad"`)인 수기 작성 레코드가 병합 단계의 cause-2 메시지 렌더에서 `record.Base[:12]` 슬라이스 panic을 내고, panic은 `releaseWindow`보다 앞서 단계를 죽여 **통합 창을 점유한 채로 남긴다**. plan 단계에서 오버레이 테스트로 실측 재현 완료: `runtime error: slice bounds out of range [:12] with length 3`, 창은 sess-b held 유지.
- **④-R7-2 (Medium)** — `[no test files]` 마커가 스트림 어디에 있든 전체 재측정을 반려한다. 실제 테스트 패키지 + 무테스트 보조 패키지의 정상 `go test -json ./...` 스윕이 거짓 거부된다. plan 단계 실측 재현 완료: `runner reported [no test files] — an empty sweep cannot stand for a re-measure`. 이는 REQ-MWQ-015의 "runner-reported empty sweep" 조항의 갱신을 수반한다 — "빈 스윕"의 정의가 마커의 존재가 아니라 **총 per-test pass 개수 0**으로 정밀화된다.
- **④-R7-1 (Medium)** — cause-7 경로(병합 실패 + abort 후 dirty)에서 hold 기록 쓰기와 창 잠금 해제가 서로 다른 mutation이다. 두 쓰기 사이에 status 갱신이나 대기자 승격이 끼어들면 원래 호출자의 release가 실패하고, hold는 남았는데 승격은 우회될 수 있다.
- **① (P1)** — "ValidateRemeasureRecord가 failure 이벤트를 무시한다"의 남은 같은-클래스 인스턴스 탐색. plan 단계 프루빙 결과: t1576이 수리한 fail 이벤트 거부(455-461행)는 유지되고, 남은 실질 재현은 **R7-2와 동일 지점(마커 상호작용)**이다. 나머지 후보는 관측 결과가 기록된 문서화된 잔여 또는 호출자 측정 포기(plan.md §Research §R1).
- **② (P2)** — `sh -c` 인자 경계: 실행 쪽(`shellJoinArgs`)과 분류 쪽(`shellSegments`/`shellFields`)의 정합성. plan 단계 round-trip 관측 결과 **경계 붕괴 미재현** — join이 보존한 인자 경계를 분류기도 읽어 돌린다(4/5 케이스 완전 일치, apostrophe 케이스는 escape 바이트 잔류가 분류 필드에 남지만 경계는 하나로 유지되어 도구·`-json` 판정에 무해). 현재 동작을 봉인 테스트로 고정한다.
- **③ (P2)** — "INVALID 판정을 렌더하면서 nil을 반환"의 남은 인스턴스. `moai integration remeasure` verb는 t1576 라운드 2에서 `remeasureVerdictError`로 수리됐다. 같은 표면의 `moai factory merge ready`는 재측정 레코드 부재/무효(측정 실패)로 REFUSED 판정을 렌더하면서도 exit 0으로 나간다 — 레인 스크립트가 판정 문서를 읽지 않으면 "측정 실패를 진행으로 읽는" t1576 라운드 2와 동일한 판독 위험.

## 2. 신뢰 모델과 범위

상위 SPEC-MERGE-WINDOW-QUEUE-001 spec.md §D의 신뢰 모델을 계승한다: 수기 작성 레코드는 검증기만으로 구별할 수 없고, 행위자는 협조적이며 위조는 방어 경계가 아니라 잔여 위험이다. 그래서 R7-3의 수리는 위조 방어가 아니라 **어떤 이상한 레코드가 들어와도 단계가 panic으로 죽어 창을 점유하지 않는다**는 무결성 보장이다.

## 3. 요구사항 (GEARS)

### C.1 검증기와 렌더의 SHA 무결성 (④-R7-3)

- **REQ-MWQ2-001** (Event-driven) — **When** the re-measure gate reads a record, the verifier shall refuse any record whose tree key or absorbed base commit is not a full 40-character hexadecimal SHA, as the record-invalid cause — before any message render that prefixes the value.
- **REQ-MWQ2-002** (Event-driven) — **When** the merge step's refusal paths render a SHA prefix, the prefixing shall be length-safe — a value shorter than the prefix length renders whole, never panics.
- **REQ-MWQ2-008** (Event-driven) — **When** any gate of the merge step encounters a malformed record value, the step shall refuse through its error path with a cause code instead of panicking.

### C.2 재측정 판정의 정밀화 (④-R7-2, ①)

- **REQ-MWQ2-003** (Event-driven) — **When** a `go test -json` stream reports a no-test package (a package-level skip whose output event carries the no-test marker), the classifier shall judge the sweep by its total per-test pass count.
- **REQ-MWQ2-009** (Event-driven) — **When** the total per-test pass count of a `go test -json` stream is zero, the classifier shall invalidate the record as an empty sweep.
- **REQ-MWQ2-004** (Event-driven) — **When** any fail event (per-test or package-level) appears in the joined command's output, the classifier shall invalidate the record — the segments' outputs remain the verdict the exit code cannot carry (the t1576 round-1 repair, promoted to contract here as the same-class sweep's anchor).

### C.3 분류-실행 경계 정합 (②)

- **REQ-MWQ2-005** (Where) — **Where** the re-measure verb joins the caller's argv into the command string, the classifiers shall re-read every argument boundary the join preserved — the joined line's tool recognition (`go test`) and `-json` request detection shall hold on the re-parsed fields, whatever quoting the join applied.

### C.4 측정 실패 판정의 exit 코드 (③)

- **REQ-MWQ2-006** (Event-driven) — **When** the merge-readiness pre-check refuses on a measurement failure (the re-measure record absent or invalid), the verb shall exit non-zero.
- **REQ-MWQ2-010** (Event-driven) — **While** a merge-readiness refusal is a window-contest verdict (waiting, another holder's window), the verb shall keep the zero-exit verdict SPEC-FACTORY-LANE-AUTONOMY-001 assigned.

### C.5 cause-7 원자성 (④-R7-1)

- **REQ-MWQ2-007** (Event-driven) — **When** the merge step reaches the cause-7 outcome (merge failed, the abort left the worktree dirty), the hold write and the window release shall run inside one serialized mutation — no status refresh or waiter promotion may interleave between them, and the original caller's release shall not fail after the hold landed.

### Out of Scope — 이 SPEC이 수리하지 않는 것

- **보호구역 POSIX 역슬래시 symlink 우회** (`internal/hook/protected_zone_path.go`, t1576 R5-2/R6-1) — 보호구역 계통(t1566 착지·t1570·t1566 PR) 소관. 이 카드에서 수정하면 열린 PR과 충돌한다.
- **todo 발행 충돌 탐지·todo-auto 레인 제외** (`todo_issuance.go`, `todo_auto_lane.go`, t1576 R5-4/R6-2/R9-1/R11-1) — 리더 큐 소관.
- **nominate 재시도의 hub 선행관계 순환** (`factory_card.go` nominate 표면, t1576 R6-3) — 리더 큐 소관.
- **non-`go test` 명령의 실패 의미 출력이 exit 0으로 valid가 되는 것** — 상위 SPEC spec.md §D가 명명한 잔여 위험(`cat`/`true`/lint류 도구는 인식 가능한 구조 보고가 없어 exit 0이면 valid). 제거는 candidate-CI 형태(SPEC-CANDIDATE-CI-001)의 소관이며 plan 단계 관측 테스트로 기록만 남긴다(`TestRedT1582ProbeNonTestCommandFailureSemanticsRideExitZero`).
- **호출자가 리다이렉션으로 버린 세그먼트의 실패** — `go test -json ./failing > /dev/null; go test -json ./passing`처럼 실패 세그먼트의 출력을 호출자가 지운 것은 도구가 볼 수 없는 측정 포기이며 결함이 아니다.
- **`moai factory merge gate`의 verdict-비제로-exit 설계** — REFUSED/PROCEED 판정은 SPEC-FACTORY-LANE-AUTONOMY-001이 문서화한 의도(`factory_merge.go` 상단 계약 주석)이며, REQ-MWQ2-006이 건드리는 것은 `merge ready`의 측정 실패 클래스뿐이다.
- **`moai factory complete`의 release 실패 시 안내 출력과 nil 반환** — 병합과 전이가 실제 성공한 뒤의 자원 정리 실패이며, 출력이 안내("moai integration release by hand")를 수반하므로 t1576 라운드 2의 "측정 실패 위장" 클래스가 아니다. 관측 기록은 plan.md §Research §R5.

## 4. 품질 기준 (요약)

- 수리 후 영향 식별자의 **전체 테스트 가족**이 관측 출력으로 GREEN: `./internal/factory/`와 `./internal/cli/`의 영향 계열 — 정규식 `^(TestIntegrationMerge|TestIntegrationRem|TestRemeasure|TestClassify|TestShellJoin|TestMWQ19|TestRedT1582|TestMergeStep|TestFactoryMergeReady)` (acceptance.md AC-MWQ2-007이 SSOT).
- 동일 클래스 스윕: 각 수리가 건드린 표면에서 새 텍스트가 같은 클래스 인스턴스를 다시 만들지 않았는가의 스캔 기록(plan.md §F M5).
- 검증은 lane-local: 전체 스위트는 돌리지 않고 영향 계열만 — 장시간 실행에는 슬롯 임대(`moai slot acquire --resource go-test-heavy --max-duration 45m`)을 먼저 잡는다.
- 도구 근원: moai CLI를 부르는 측정은 반드시 본 트리 빌드(`make build` 후 `./bin/moai` 경로 호출)로 수행한다 — 설치본이 판정자가 되지 않는다.
