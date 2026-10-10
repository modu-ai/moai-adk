---
id: SPEC-REVIEW-GATE-REPAIR-001
title: "Codex 리뷰 게이트 수리 — 프롬프트 진입 교착을 풀고 리뷰 기준을 하나로 모은다"
version: "0.1.0"
status: draft
created: 2026-10-10
updated: 2026-10-10
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: L
tags: "codex, review-gate, turn-end, prompt-entry, receipt, merge-base, card-t1620"
---

# SPEC-REVIEW-GATE-REPAIR-001 — Codex 리뷰 게이트의 프롬프트 진입 교착을 풀고 리뷰 기준을 하나로 모은다

카드: **t1620** (class C — 훅·영수증·리뷰 범위·설정 네 서브시스템을 가로지르는 설계 변경) · 작업 트리: `WT-review-gate-fix` @ `2aab5f797`

## HISTORY

- 2026-10-10 · v0.1.0 · manager-spec · 최초 작성(draft). Tier L 판정 근거는 plan.md §A. 측정 원천은 research.md(인용 전수 재검증, 기준선 셀 실행). 운영자 판정이 필요한 항목 Q1–Q6 은 decision-index.md 에 개방 상태로 두었다. Q1–Q6 이 기록되기 전에는 Implementation Kickoff 가 차단된다.

---

## §0 지배 원칙 [HARD]

> **턴 종료 게이트만 차단하고, 프롬프트 진입은 차단하지 않는다. 검토가 이뤄지지 않은 상태는 "통과"로 재사용되지 않는다.**

교착의 구조는 이렇다. 진입 단계에서 막힌 프롬프트는 모델에게 턴을 주지 않으므로, 리뷰 키가 움직일 방법이 없다(research.md §4). 모델이 조치할 수 있는 지점은 턴 종료뿐이다. 그래서 검토 결과는 진입 시 한 번 전달되고, 차단은 턴 종료에서만 일어나며, 그 차단도 상한을 가진다. 검토가 되지 않은 상태는 드러나야 하고, 통과로 세어지면 안 된다.

---

## §A 배경 (측정)

### A.1 재현 기준 — 카드 변경이 없는 레인이 24시간 잠겼다

- 관측: 레인 저장소 `.moai/worktrees/t1543/.moai/state/verify/snapshots/2aab5f797b75-7f1e2657f0831d9f.json` 의 `codex-review` 검사는 `verdict: fail`, `exit_code: 1`, `recorded_at: 2026-10-10T01:37:46.265725+09:00`, `tool_version: codex-cli 0.161.0` 이다. 키는 `2aab5f797…:7f1e2657f0831d9f` 로, 이 트리의 HEAD 와 같다.
- 상세 파일의 발견 사항 두 건은 `internal/homestate/card_transition.go:287` 과 `internal/cli/factory_card.go:2894` 를 가리킨다. 두 경로는 이 트리에 있다. 발견 사항이 옳은지는 이 SPEC 이 판단하지 않는다(research.md §4).
- 미관측: 레인 자체의 커밋 수. 이 세션의 git 가드는 다른 워크트리를 대상으로 한 git 호출을 거부했다. "카드가 아무것도 바꾸지 않았다"는 전제는 Gap 으로 남긴다.

### A.2 리뷰 기준은 두 해석기가 서로 다른 설정 키를 읽는다

- 턴 종료 게이트와 card 범위 리뷰는 `cardMergeBase` 로 기준을 정한다. 이 함수는 `IntegrationTarget` 을 읽고(`internal/cli/codex_review_scope.go:189-191`), 그 값은 git-flow 수동 프로필에서 `develop_branch`(`.moai/config/sections/git-strategy.yaml:16`, `develop`)다(`internal/config/loader_integration_branch.go:123`).
- 감사 도구는 `resolveReviewBase` 를 쓰며, 이 함수는 첫 단계로 `worktree_base_branch` 를 읽는다(`internal/cli/mcp_review_material.go:159`, 값 `main`, `git-strategy.yaml:5`). 이 단계는 커밋 `f82a078bb`(2026-10-03, 카드 t1426)가 추가했다.
- 토폴로지(git 측정): `merge-base(main, develop) = db0c514d3`, `db0c514d3..main` 139 커밋, `db0c514d3..develop` 149 커밋.
- 같은 저장소에서 두 해석기는 서로 다른 기준을 고른다. M2 가 필요한 이유다.

### A.3 완료된 SPEC 의 결정이 코드에서 뒤집혀 있다

- SPEC-CODEX-REVIEW-TARGET-001(status: completed) §A.7 은 codex 네이티브 경로가 `worktree_base_branch` 를 읽지 않는다고 결정했다(`.moai/specs/SPEC-CODEX-REVIEW-TARGET-001/spec.md` 102–111행).
- 카드 t1426 이 이 결정을 뒤집었다. 감사 기준 테스트의 주석이 이를 명시한다: "Card t1426 reverses the spec.md §A.7 non-read clause"(`internal/cli/codex_review_target_test.go`, AC-CRT-002 절).
- 해당 SPEC 의 본문에는 t1426 을 가리키는 항목이 없다(`grep -c t1426` = 0). 결정의 기록은 코드와 테스트뿐이다. 운영자 판정은 Q1 에서 받는다.

### A.4 진입 차단은 진행 중 SPEC 의 명시 요구다

- SPEC-GATE-BOTTLENECK-001(status: in-progress)의 REQ-GBN-002: "a fresh FAIL at turn entry SHALL block. `async:true` SHALL NOT be used (it surrenders the block capability)." 진입 훅의 헤더 주석(`internal/cli/codex_review_delay.go` 1–14행)도 같은 설계를 적었다. M1 은 이 요구를 뒤집는다. 기록 처분은 Q6 이다.

### A.5 영수증과 재사용 규칙

- 영수증의 결합 필드는 다섯 개다: head, tree_digest, config_digest, command, tool_version(`internal/verify/receipt.go:70-76`). 유효기간은 24시간이다(`internal/cli/codex_review_cache.go:29`).
- 현재 턴 종료 게이트는 수용된 영수증이 pass 든 inconclusive 든 허용한다(`internal/cli/codex_review_gate.go:176-178`). 관측: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 가 통과하며, stderr 에 "reusing the cached inconclusive verdict for the unchanged tree" 를 남긴다(acceptance.md §C, 셀 B-2).

### A.6 설정 상태

- 추적 파일 `.moai/config/sections/workflow.yaml` 302행의 `codex.review_gate.enabled` 는 HEAD 에서 `true` 다. 작업 트리에는 리더의 미커밋 변경으로 `false` 가 있다(한 줄 차이). 이 SPEC 은 그 파일을 건드리지 않으며, 측정은 HEAD 기준으로 한다.

### A.7 카드의 두 주장 정정

- 카드가 적은 `internal/hook/codex_review_*.go` 경로는 이 트리에 없다. 네 파일은 모두 `internal/cli/` 에 있다(research.md §2).
- "두 층을 모두 켜면 doctor 경고가 뜬다"는 주장은 코드에서 관측되지 않았다. doctor 파일에 해당 검사가 없고, 문서 한 문장(`.claude/rules/moai/workflow/factory-dispatch-detail.md:157`)만 있다(research.md §5, Q5).

---

## §B 요구사항 (GEARS)

### REQ-RG-001 — 진입 훅은 프롬프트를 막지 않는다 (Ubiquitous, M1)

The codex review entry hook shall not return a block decision for any receipt state.

어떤 영수증 상태에서도 진입 훅은 차단 결정을 내지 않는다. 이 요구가 교착을 끊는다. 발견 사항의 전달은 REQ-RG-002 가 맡는다.

### REQ-RG-002 — 발견 사항은 키마다 한 번 모델에 전달된다 (event-driven, M1)

When a prompt is submitted and a fresh fail receipt exists for the current review key, the entry hook shall add the fail detail to the model context as additional context exactly once for that key, and shall record that delivery against that key.

이미 전달된 키는 프롬프트마다 반복하지 않는다. 리뷰 키(영수증 결합 필드의 조합)가 바뀌면 새 전달 기록이 생긴다.

### REQ-RG-003 — 차단은 턴 종료에서만 한다 (event-driven, M1)

When a turn ends with a fresh fail receipt for the current review key, and the stop has not been released by the stop-hook-active flag, the turn-end gate shall block the stop and shall give the fail detail as the reason.

현행 턴 종료 차단을 유지한다. 모델이 조치할 수 있는 유일한 지점이기 때문이다(§0).

### REQ-RG-004 — 연속 차단에는 상한이 있다 (unwanted, M1, 값은 Q2)

The turn-end gate shall not block more than C consecutive stops for the same review key; once that limit is reached it shall allow the stop and shall report the unreviewed state in a visible system message.

C 의 값은 운영자 판정 Q2 가 정한다. 이 SPEC 은 값을 고르지 않는다. 상한에 닿은 뒤의 허용은 침묵하지 않는다.

### REQ-RG-005 — 리뷰 기준은 하나의 해석 함수가 정한다 (Ubiquitous, M2)

The turn-end gate, the card-scope review tool, and the audit tools shall obtain the review base from one shared resolution function.

두 해석 경로가 같은 입력에 같은 기준을 내야 한다. 함수를 하나로 묶는 것만으로는 부족하고, 모든 호출 경로가 그 함수를 거쳐야 한다.

### REQ-RG-006 — 기준은 HEAD 에 가장 가까운 공통 조상이다 (state-driven, M2)

While selecting among merge-base candidates of the configured base refs, the shared resolution function shall select the candidate with the fewest commits between it and HEAD.

카드가 main 에서 잘렸다면 merge-base(main, HEAD) 가, develop 에서 잘렸다면 merge-base(develop, HEAD) 가 선택된다. 후보 집합(로컬 ref 와 원격 추적 ref 를 함께 볼지)은 plan.md §C 가 정한다.

### REQ-RG-007 — 두 키가 다른 ref 를 가리키면 판정은 inconclusive 다 (event-detected, M2, 결정 대기 Q1)

When the configured integration target and the configured worktree base branch name different refs, the shared resolution function shall return an inconclusive result that names both refs.

카드 M2 의 문언을 그대로 옮긴 요구다. 이 저장소에서 두 키는 항상 다르므로(§A.2), 문언대로 구현하면 모든 판정이 inconclusive 가 된다. 이 결과를 받아들일지는 Q1 이 정한다. Q1 이 기록되기 전까지 이 요구는 확정되지 않았다.

### REQ-RG-008 — 기준 후보가 없으면 다른 기준으로 넘어가지 않는다 (unwanted, M2)

If none of the configured base refs resolves in the tree, the shared resolution function shall return an inconclusive result and shall not substitute another base.

### REQ-RG-009 — 재사용은 pass 와 fail 에만 허용한다 (Ubiquitous, M3)

The turn-end gate shall reuse a stored review receipt for an unchanged review key only when the stored verdict is pass or fail.

### REQ-RG-010 — inconclusive 뒤에는 재검토를 시작한다 (event-driven, M3, 간격은 Q4)

When a turn ends and the stored verdict for the current review key is inconclusive, the turn-end gate shall start a review in the background, subject to the in-flight marker, and shall not allow the stop as if a verdict existed.

재검토의 간격(즉시 재시작인지 쿨다운인지)은 Q4 가 정한다. 진행 중 표식에 의한 중복 방지는 기존 메커니즘을 쓴다.

### REQ-RG-011 — 실행 단위의 스위치는 하나다 (Ubiquitous, M4, 형태는 Q3)

The codex review gate shall be governed by one run-wide switch that applies to every lane of the run, and that switch shall not be a per-tree tracked setting.

스위치의 형태(환경 값 또는 주 체크아웃의 미추적 재정의)는 Q3 가 정한다.

### REQ-RG-012 — 셸 층과 Go 층은 같은 스위치 값을 읽는다 (Ubiquitous, M4)

The shell wrapper and the Go handler shall read the same switch value for the same run.

### REQ-RG-013 — 현재 pass 영수증은 재검토 없이 허용한다 (Ubiquitous, 회귀)

While a fresh pass receipt exists for the current review key, the entry hook and the turn-end gate shall allow without starting a new review.

### REQ-RG-014 — 기존의 비검토 건너뜀은 결과가 바뀌지 않는다 (Ubiquitous, 회귀)

The codex review gate shall keep its current outcome for disabled gates, active stop-hook releases, trees without reviewable changes, and absent reviewer binaries.

---

## §C AC 와 검증 구속

- 모든 인수 기준은 acceptance.md 에 있다. 이 SPEC 은 plan 시점에 릴리스 게이트를 하나도 갖지 않는다. 이유: 행위 요구를 단언하는 테스트가 현재 트리에 없고(research.md §5, 테스트 이름 확인), 기존 행위 테스트는 결함을 고정할 뿐이다(acceptance.md §A). 릴리스 게이트로 승격하려면 run 단계가 구현 전 트리에서 RED 를 원문으로 기록해야 한다.
- 검증 명령은 항상 작업 트리 루트에서, 단일 호출로 실행한다. 실행된 테스트 수를 함께 기록하며, 0개는 측정 불가로 보고한다.

## §D 실행 순서 구속

1. Q1–Q6 이 운영자 판정으로 기록되기 전에는 run 단계로 넘어가지 않는다.
2. 각 마일스톤은 RED(실패 확인) → GREEN → 회귀 순으로 진행한다. RED 출력은 progress.md §E.2 에 원문으로 남긴다.
3. 훅 래퍼와 설정은 `.claude/` 와 `internal/template/templates/` 두 면에 있다. 두 면을 같은 변경 세트에서 함께 갱신한다.

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — 영수증 결합 기준의 변경 (F1)

- 영수증을 HEAD 대신 (merge base, 내용 digest) 에 묶는 변경은 이 SPEC 에서 하지 않는다. Follow-up F1 로 추적한다.
- 현재 결합 필드 다섯 개의 구성 변경도 하지 않는다.

### Out of Scope — 가시성, 중복 제거, kick 정리 (F2)

- 차단 사유의 중복 제거와 길이 상한, 미검토 상태의 systemMessage 표시 확장, kick 의 pid·liveness 기반 중복 제거와 레인 전체 동시성 상한, 마커 누수 정리, 추적 diff 가 비면 kick 하지 않는 규칙, 진입 훅 timeout 의 5초 정렬은 이 SPEC 밖이다. Follow-up F2 로 추적한다.

### Out of Scope — factory 저장소 경로 (F3)

- `internal/factory/todo_dispatch_current.go:53` 과 `internal/factory/todo_runtime.go:364` 의 두 주장은 이 SPEC 의 diff 에 없다. 별도 후보로 추적한다.

### Out of Scope — 감사 도구의 나머지 결함

- `audit_multi` 의 codex 레그가 target 을 무시하는 문제, claude 자식 프로세스 env 의 denylist 방식, codex 샌드박스 정책 미전달, 작업 수명 정리 누락, factory 메시지의 null 반환은 카드 항목 밖이다. 연구 입력 lens-3 의 별도 후보로 남긴다.

### Out of Scope — doctor 진단 신설

- 두 리뷰 계층이 동시에 켜졌을 때 doctor 경고를 추가하는 일은 Q5 판정 전까지 범위 밖이다.

### Out of Scope — 두 리뷰 계층 중 하나의 폐기

- 턴 종료 게이트나 card-review 단계 가운데 하나를 없애는 결정은 이 SPEC 이 하지 않는다. 중복 검토는 D10 후보 결함으로만 기록한다.

### Out of Scope — GLM 턴 종료 게이트 신설

- 현재 GLM 턴 종료 게이트는 존재하지 않는다(research.md §3). 신설은 범위 밖이다.

---

## Follow-up

- **F1 — 영수증 결합 기준 변경.** 영수증을 HEAD 대신 (merge base, 내용 digest) 에 묶는다(카드 항목 4). 별도 후속 카드로 추적한다. 요구사항을 붙이지 않는다.
- **F2 — 가시성과 운영 정리.** 카드 항목 6–10: 미검토 상태를 systemMessage 로 드러내기, 차단 사유의 중복 제거와 길이 상한, kick 을 pid·liveness 로 중복 제거하고 레인 전체 상한을 두며 마커 약 64개 정리, 추적 diff 가 비면 kick 하지 않기, 진입 훅 timeout 을 5초에서 15초 예산과 맞추기. 요구사항을 붙이지 않는다.
- **F3 — factory 후보 두 건.** (a) `internal/factory/todo_dispatch_current.go:53` 의 백로그 SQLite 열기가 JSON 큐 이관 전에 빈 SQLite 를 만들 수 있다. (b) `internal/factory/todo_runtime.go:364` 의 할당 트랜잭션에서 외부 저장소 변경 뒤 내부 커밋(약 368행)이 실패하면 복구되지 않는다. 두 주장의 검증 결과는 research.md §7 에 있다. 요구사항을 붙이지 않는다.

---

## §F 제약

- 이 SPEC 의 산출물은 `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/` 안에만 둔다.
- 작업 트리의 `.moai/config/sections/workflow.yaml` 미커밋 변경은 건드리지 않는다.
- 기준 SHA 는 `2aab5f797` 이다. 설치된 빌드(4f8aba061)는 트리보다 새 커밋이므로 트리의 증거로 인용하지 않는다.

## §G 참조

- research.md — 인용 재검증 등록부, 재현 근거, 전제 확인
- design.md — 상태 모델과 판정 표
- plan.md — 마일스톤과 기술 접근
- acceptance.md — 인수 기준 매트릭스와 기준선 셀
- decision-index.md — 운영자 판정 항목 Q1–Q6
- `.moai/specs/SPEC-CODEX-REVIEW-TARGET-001/spec.md` §A.7 · `.moai/specs/SPEC-GATE-BOTTLENECK-001/spec.md` REQ-GBN-002
