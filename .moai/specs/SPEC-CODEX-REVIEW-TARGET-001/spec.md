---
id: SPEC-CODEX-REVIEW-TARGET-001
title: "codex native review/start 의 target 객체를 스키마 계약대로 채운다"
version: "0.3.0"
status: completed
created: 2026-09-01
updated: 2026-09-01
author: manager-spec
priority: P0
phase: "v3.1.5 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, mcp, review-start, base-branch, fail-open, issue-1632"
---

# SPEC-CODEX-REVIEW-TARGET-001 — native audit 이 codex 에 닿게 한다

카드: **t399** · 외부 이슈: **modu-ai/moai-adk#1632** (제보 Al-Lukyanets, 2026-08-24)

## HISTORY

- 2026-09-01 · v0.3.0 · manager-spec · plan-audit iter2 (PASS-WITH-DEBT 0.81, 회귀 없음) 의 blocking 3건 + optional 1건 수리. **정정 — v0.2.0 의 D2 주장은 과장이었다**: 그 항목은 "`worktree_base_branch` 우선을 철회하고 정렬"이라고 적었으나, 철회는 `spec.md`(§A.7 · REQ-CRT-003 · §F)와 `plan.md`(§C · 안티패턴 9)에만 닿았고 **`acceptance.md` AC-CRT-002 는 철회된 설계를 그대로 요구하고 있었다.** 같은 개정이 AC-CRT-004 의 설정 키 절은 제거했으므로 의도적 예외가 아니라 전파 누락이다. run-phase 가 AC 에서 검사를 쓰는 이상, 그 상태로 M2a 에 들어갔으면 §A.7 이 기각한 설계가 RED→GREEN 을 거쳐 그대로 착지했을 것이다. **D2-RESIDUAL** — AC-CRT-002 를 정렬된 사슬로 재작성(+ 두 값이 갈리는 픽스처에서만 두 설계가 구별된다는 [HARD] 절). **N1** — AC-CRT-010 의 RED 를 **관측된 것으로 적은 3곳**을 예측형으로 되돌림. 라이브 왕복은 이 트리에서 한 번도 실행되지 않았고, 그것을 관측으로 적은 것은 이 SPEC 이 고치려는 결함(관측 없는 주장)과 같은 형태다. **N2** — AC-CRT-006b 의 부재 단언에 대체 금지 절 추가, 뿌리인 REQ-CRT-005 에 대체 금지 + 양성 귀결 부여(현행 default 분기가 `uncommittedChanges` 로 떨어지므로 첫 절만으로는 붉지 않았다). **D8-RESIDUAL** — REQ-CRT-004 의 modifier `Where` → `When`(런타임 상태이지 capability gate 가 아니다). **N3(식별자 `006b`)는 현행 유지** — 짝 sub-criteria 표기는 허용된 관례이고, `011` 로 옮기면 쌍의 반쪽이 목록 끝으로 떨어진다.
- 2026-09-01 · v0.2.0 · manager-spec · plan-audit iter1 (FAIL 0.75, Tier M 임계 0.80) 수리. **D1** — 카드의 [HARD] 회귀선을 스키마 계약으로 대체한 것이 완화임을 인정하고 **AC-CRT-010**(라이브 baseBranch 왕복) 추가. 저장소 안 선례(`codex_live_protocol_probe_test.go` / `codex_review_gate_live_test.go`)를 §G 에 인용해 재사용임을 명시 — §A.3 의 재사용 규율을 검증 표면에도 적용. **D2** — `worktree_base_branch` 우선을 **철회**하고 GLM 경로(`resolveReviewMergeBase`)와 정렬(결정 (a), §A.7). **D3** — AC-CRT-006 을 직렬화/비직렬화 두 속성으로 분할. **D4** — AC-CRT-004 의 관측 필드를 후보별로 고정, plan §B (나) 대가에 `applyGateUnmet` 미경유 추가. **D5** — 해석 사슬의 중복 단계 병합 + 반환 이름의 ref 해석 가능성 확인 요구. **D6** — 도구 표면 요구를 **REQ-CRT-007** 로 승격, AC-CRT-008 매핑을 `§C 규율` 로 정정. **D7** — §A.4 좌표를 심볼 줄로 통일, §E 의 t284 출처를 succession.md(축)와 큐(카드 번호)로 분리. **D8** — REQ-CRT-004 를 canonical `shall not` 형으로 재작성. **D9** — DoD 에 #1632 응답 문언 구속 추가. **D10** — REQ-CRT-006 의 "byte-identical in shape" 를 shape-identical 로 고치고 기준을 `442da4f06` 로 고정.
- 2026-09-01 · v0.1.0 · manager-spec · 최초 작성. 카드 t399 (최우선, 외부 사용자 이슈). 측정 원천은 `.moai/reports/t399/discovery.md` 와 `.moai/reports/t399/schema/v2/ReviewStartParams.json` (codex-cli 0.150.1), 트리 `.claude/worktrees/t399` @ `442da4f06` (= `origin/develop`).

---

## §0 지배 원칙 [HARD]

> **요청이 거절된 것과 리뷰가 통과한 것은 같은 값으로 보고될 수 없다.**

지금 `mode=native, target=baseBranch` 경로는 codex 가 요청을 스키마 위반으로 거절하는데도 fail-open 이 그것을 `inconclusive` 로 흡수한다. 이 SPEC 이 닫는 것은 **요청을 스키마대로 만드는 일** 하나다. `inconclusive` 를 통과와 구별해 보이는 일은 별개 축이며 카드 t284 소관이다(§E).

그 귀결로 이 SPEC 의 회귀 방어선은 두 층이다. 반환 여부를 세는 검사는 이 카드가 고치려는 결함 그 자체를 통과로 세므로 어느 층도 그것을 쓰지 않는다.

1. **계약 층** — 직렬화된 요청이 측정된 스키마의 required 집합을 만족한다. 결정적이고, codex 없이도 돈다.
2. **왕복 층** — 실 codex 가 그 요청을 거절하지 않는다(AC-CRT-010). 카드의 [HARD] 문언("native 경로가 실제로 codex 에 도달했는지 단언")이 지목하는 것이 이 층이다.

계약 층만으로 대신하지 않는 이유: 계약이 곧 codex 의 수락을 뜻한다는 것은 **스키마 문서의 주장이지 이 SPEC 의 측정이 아니다**. 그리고 왕복 층은 발명이 아니라 재사용이다 — 저장소에 라이브 `review/start` 선례와 skip 가드가 이미 있다(§G). §A.3 에서 해석기에 적용한 재사용 규율은 검증 표면에도 똑같이 적용된다.

---

## §A 배경 (측정)

### A.1 결함 — bare string 리프트는 네 variant 중 하나에서만 옳다

`internal/cli/mcp_codex.go` `coerceCodexReviewTarget` 는 문자열을 `{"type": s}` 로 올린다.

```go
if s, ok := v.(string); ok && s != "" {
    return map[string]any{"type": s}
}
```

측정된 프로토콜 계약 (codex-cli 0.150.1, `codex app-server generate-json-schema` 로 이 트리에서 산출):

| variant | required | bare string 리프트 결과 |
|---|---|---|
| `uncommittedChanges` | `[type]` | **적법** |
| `baseBranch` | `[branch, type]` | 위반 — `branch` 없음 |
| `commit` | `[sha, type]` | 위반 — `sha` 없음 |
| `custom` | `[instructions, type]` | 위반 — `instructions` 없음 |

즉 카드가 서술한 것은 `baseBranch` 사례이고, 일반형은 **"bare string 은 `uncommittedChanges` 에서만 well-formed 하다"** 이다.

### A.2 호출자는 브랜치 이름을 실을 수 없다

`internal/cli/mcp_server.go:255` 에서 `target` 은 두 값짜리 문자열 enum 이고 동반 브랜치 파라미터가 없다. `internal/cli/mcp_codex.go:1481` 이 그 문자열을 그대로 받는다. 따라서 교정은 "호출자가 객체를 넘기게 한다"가 될 수 없다 — **브랜치 이름은 서버가 해석해야 한다.**

### A.3 해석 원천은 이미 있다 (단순성 사다리 2단)

`internal/cli/mcp_review_material.go` `resolveReviewMergeBase` 가 `origin/HEAD → origin/main → main` 폴백 사슬을 이미 갖고 있고, GLM 경로가 그것을 쓴다.

새 해석기를 발명할 필요가 없다. 다만 `resolveReviewMergeBase` 가 돌려주는 것은 **merge-base SHA** 이고 codex `baseBranch` 가 요구하는 것은 **브랜치 이름**이므로, 같은 사슬을 이름 층위에서 읽는 형태가 필요하다.

`.moai/config/sections/git-strategy.yaml` 의 `git_strategy.worktree_base_branch`(reader: `internal/config/loader_worktree_base.go` `LoadWorktreeBaseBranch`)도 후보였으나 §A.7 의 결정으로 **채택하지 않는다**.

### A.4 [HARD] 기존 테스트는 있으나, 옳은 쪽 variant 에서만 잰다 — 카드 전제의 정정

카드와 discovery 는 "`coerceCodexReviewTarget` 을 덮는 테스트가 없다"고 적었다. 함수 이름 grep 기준으로는 참이지만 **행동 기준으로는 거짓**이며, 그 어긋남의 방향이 중요하다.

좌표 관례: 아래 표의 줄 번호는 전부 **`func` 선언 줄**이며, 기준 트리는 `442da4f06` 이다.

| 검사 | 파일 | 실제로 재는 것 |
|---|---|---|
| `TestCodexRPC_TargetIsTaggedObject_BareStringLifted` | `internal/cli/codex_review_rpc_test.go:63` | 리프트 동작 — 단 `uncommittedChanges` 로만 |
| `TestCodexRPC_BareStringTargetNotSentInReviewRequest` | `internal/cli/codex_review_rpc_test.go:96` | bare string 미직렬화 — 단 `uncommittedChanges` 로만 |
| `TestCodexAudit_NativeDispatchesReviewStart` | `internal/cli/mcp_codex_test.go:183` | native 가 review/start 를 3번째로 보냄 — 단 `uncommittedChanges` 로만 |
| `TestCodexAudit_AdversarialDispatchesTurnStart` | `internal/cli/mcp_codex_test.go:216` | `target: baseBranch` 를 넘기지만(`:231`) **adversarial 모드**라 target 이 전송되지 않는다 — target 형태에 대해 아무것도 단언하지 않는다 |

정정된 전제: **리프트 행동은 이미 고정돼 있고, 고정된 지점이 하필 리프트가 옳은 유일한 variant 다.** 저장소 안에서 `baseBranch` 문자열을 든 유일한 codex 검사는 그 값이 전송조차 되지 않는 경로에 있다. 이것은 "커버리지 없음"이 아니라 **variant 선택으로 만들어진 공허한 초록**이며, RED 를 세우는 방식도 달라진다: 새 검사는 없는 검사를 채우는 것이 아니라 **기존 검사가 보지 않는 입력**을 겨눈다.

### A.5 스크립트 스텁은 스키마를 검증하지 않는다

`withCodexSession` / `codexSessionScript` 는 요청과 무관하게 준비된 NDJSON 을 되돌려준다. 따라서 잘못된 `review/start` 도 스텁 앞에서는 `pass` 로 끝난다. 이 SPEC 의 단언은 스텁의 반환값이 아니라 **`sess.sent[2]` 에 실제로 직렬화된 바이트**를 대상으로 해야 한다.

### A.6 GLM 경로와의 비대칭

`internal/cli/mcp_glm.go:230` 의 같은 `baseBranch` 값은 정상 동작한다 — GLM 은 diff 를 수집해 보내므로 `resolveReviewMergeBase` 만으로 충분하기 때문이다. 즉 `audit_multi` 에서 `target=baseBranch` 를 쓰면 **GLM 은 리뷰하고 codex 는 거절당하는** 상태이며, 이것이 제보자가 본 화면의 구조적 이유다.

### A.7 [HARD] 해석 원천의 선택 — 새 비대칭을 만들지 않는다 (결정)

§A.6 이 현행 비대칭을 결함으로 규정했으므로, 이 SPEC 은 그 자리에 **새 비대칭을 만들지 않을 의무**를 진다. 후보는 둘이었다.

| 후보 | 결과 |
|---|---|
| `git_strategy.worktree_base_branch` 를 1순위로 | codex 는 설정값을, GLM 은 `origin/HEAD` 를 기준으로 리뷰한다. 두 값이 갈리는 트리에서 `audit_multi(target=baseBranch)` 의 두 백엔드가 **서로 다른 변경분**을 보고, 그 판정이 `mcp_convergence.go` 에서 하나로 수렴된다 |
| **`resolveReviewMergeBase` 와 정렬** (채택) | 두 백엔드가 같은 사슬을 읽으므로 같은 변경분을 리뷰한다 |

**결정: 정렬한다.** REQ-CRT-003 은 `resolveReviewMergeBase` 와 동일한 사슬을 쓰고 `worktree_base_branch` 는 읽지 않는다.

근거: 설정 키가 옳다면 그것은 **두 백엔드 모두에게** 옳다. 한쪽만 읽게 만드는 것은 이 SPEC 이 진단한 것과 같은 종류의 비대칭을 반대 방향으로 하나 더 만드는 일이며, 그 대가를 치르는 곳(수렴 판정)은 이 SPEC 이 손대지 않기로 한 t284 의 소관이다.

이 트리에서는 `worktree_base_branch: develop` 과 `origin/HEAD → origin/develop` 이 **우연히 일치**한다(둘 다 측정). 일치는 관측이고 일반화는 성립하지 않으므로, 이 트리에서 증상이 안 보인다는 사실은 어느 쪽 후보의 근거도 되지 못한다.

**뒤집을 수 있는 결정이다.** 설정 키가 기준이어야 한다는 판단이 서면, 그것은 `resolveReviewMergeBase` 를 고쳐 **두 백엔드에 동시에** 적용하는 별도 카드다 — GLM 경로의 동작이 바뀌므로 이 카드의 범위가 아니다.

---

## §B 요구사항 (GEARS)

### REQ-CRT-001 — 전송되는 target 은 스키마 계약을 만족한다 (Ubiquitous)

The codex review request builder shall serialize a `target` object that satisfies the required-field set of its own variant as declared by the codex `ReviewStartParams` schema.

### REQ-CRT-002 — baseBranch 요청은 브랜치 이름을 싣는다 (event-driven)

When `codex_audit` is invoked with `mode=native` and `target=baseBranch`, the request builder shall include a non-empty `branch` field resolved by the server.

### REQ-CRT-003 — 해석 사슬은 GLM 경로와 같다 (state-driven)

While resolving the base branch name for a named project root, the resolver shall follow the same fallback chain `resolveReviewMergeBase` uses — the remote default head, then `origin/main`, then `main` — and shall return a name only after confirming that name resolves as a ref in that tree.

각 단계가 **반환하는 이름 자체의 해석 가능성**을 확인한다는 절이 이 요구의 절반이다. 확인 없이 이름을 돌려주면 codex 가 없는 브랜치로 리뷰를 시도하고, 그 실패는 이 SPEC 이 닫은 자리에서 다시 `inconclusive` 로 나타난다. 사슬의 형태(중복 없는 단계 목록)는 plan.md §C 가 소유한다.

### REQ-CRT-004 — 해석 실패는 조용히 다른 리뷰가 되지 않는다 (unwanted)

When no base branch can be resolved for the named project root, the codex audit shall not issue a `review/start`, and shall not substitute `uncommittedChanges`.

modifier 는 `When` 이다. GEARS 의 `Where` 는 capability gate / feature flag / static config 를 가리키는데, "base 를 해석할 수 없다"는 **런타임 상태**이지 정적 게이트가 아니다.

### REQ-CRT-005 — 미완성 객체는 직렬화되지 않는다 (unwanted + 양성 귀결)

The request builder shall not serialize a `target` object for a variant whose required fields it cannot populate, and shall not substitute another variant's `target` in its place.

When the request builder cannot populate a variant's required fields, the codex audit shall decline to issue the `review/start` and shall report the unsupported variant as the cause in the same named output field REQ-CRT-008 designates.

두 번째 문단이 이 요구의 **양성 귀결**이다. 금지만 두면 "무엇이 일어나면 안 되는지"는 고정되지만 "무엇이 일어나야 하는지"가 비고, 그 빈자리를 현행 구현의 default 분기(`internal/cli/mcp_codex.go:1004` → `uncommittedChanges`)가 조용히 채운다. 대체 금지 절과 양성 귀결이 함께 있어야 그 경로가 닫힌다.

### REQ-CRT-006 — uncommittedChanges 경로는 변하지 않는다 (Ubiquitous, 회귀)

The `uncommittedChanges` review request shall remain shape-identical to its form at `442da4f06`.

기준을 SHA 로 고정하는 이유: "pre-change" 는 움직이는 참조라 측정과 판독 사이에 뜻이 달라진다. 그리고 요구하는 것은 **형태 동일**(키 집합과 값)이지 바이트 동일이 아니다 — 직렬화 순서까지 묶으면 무관한 변경이 이 회귀선을 거짓으로 붉게 만든다.

### REQ-CRT-007 — 도구 표면이 서버 해석을 서술한다 (Ubiquitous)

The `codex_audit` tool description shall state that the `baseBranch` target's branch name is resolved server-side, and shall name the resolution source.

호출자는 브랜치를 공급할 수 없다(§A.2). 그 사실을 적지 않은 도구 설명은 공급할 수 없는 값을 공급하라고 암시한다.

### REQ-CRT-008 — 해석 실패의 원인은 관측 가능한 필드에 실린다 (Ubiquitous)

The codex audit shall report an unresolvable base branch in a named output field, distinguishable from every other fail-open cause.

REQ-CRT-004 는 무엇이 일어나면 안 되는지를 고정하고, 이 요구는 무엇이 관측 가능해야 하는지를 고정한다. 어느 필드인지는 plan.md §B 의 (가)/(나) 결정이 정하며, AC 는 후보별로 그 필드를 명시한다.

---

## §C AC 형태에 대한 구속 [HARD]

- AC 는 **직렬화된 요청 바이트**를 관측한다. 반환된 verdict 값은 단독으로는 어떤 AC 의 근거도 되지 못한다.
- `inconclusive` 를 통과로 세는 단언은 금지한다. 관측할 것은 **"이 요청이 스키마를 만족한 채 나갔다"** 는 양성 사실이다.
- 신규 AC 는 **현행 구현에 대해 실패해야 한다**(RED). RED 는 서술이 아니라 관측으로 기록한다 — 변경 전 트리에서 실패 출력을 얻어 남긴다.
- §A.4 가 보인 대로, 기존 검사가 초록인 것은 교정의 근거가 아니다.

전체 AC 목록과 Given-When-Then 은 `acceptance.md` 에 있다.

---

## §D 실행 순서 구속

`baseBranch` 해석기는 `uncommittedChanges` 회귀 방어선(AC-CRT-003)이 초록으로 고정된 뒤에 도입한다. 두 경로가 같은 함수를 지나므로, 회귀선이 없는 상태의 수정은 고친 것과 깨뜨린 것을 구별하지 못한다.

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — 백엔드 참여 여부의 판정 노출 (카드 t284)

- 카드 t399 가 명시한 2축 중 두 번째("백엔드가 아예 안 뜬 것과 떠서 동의한 것을 구별한다")는 이 SPEC 이 설계하지 않는다.
- 근거는 두 출처로 나뉘며, 하나가 둘을 다 담고 있지 않다.
  - **축의 존재와 처분** — `.moai/reports/t229/succession.md`: 그 축을 **카드↔SPEC 범위 갭**으로 기록하고 운영자 승인(2026-08-26)으로 "별도 신규 카드"에 이관했다. **이 문서는 카드 번호를 담고 있지 않다.**
  - **카드 번호와 본문** — `moai todo` 큐 판독: t284 가 그 카드이며, 본문이 (1) on-target 백엔드 수 노출 (2) 참여자 2 미만이면 `disagreement_flag=false` 금지 (3) 대표 mutant("수를 세되 여전히 false 를 내보내는 구현")를 이미 명세한다.
- 기제 위치도 갈린다: 이 SPEC 은 `mcp_codex.go` 의 요청 조립부를, t284 는 `mcp_convergence.go` 의 `DisagreementFlag` 유도부(`internal/cli/mcp_convergence.go:168-227`)를 건드린다. 같은 파일도 같은 함수도 아니다.
- 따라서 중복 설계 금지 지시는 유지한다. 이 SPEC 은 t284 를 **선행 조건으로 두지도 않는다** — 요청이 스키마를 만족하는 일은 판정 노출과 독립적으로 옳다.

### Out of Scope — `commit` / `custom` 을 호출자 표면에 노출하기

- REQ-CRT-005 는 이 두 variant 를 **안전성 축에서만** 범위에 넣는다: 미완성 객체를 내보내지 않는다.
- `codex_audit` 의 `target` enum 에 `commit` / `custom` 값을 **추가하지 않는다**. 제보된 결함은 기능 부재가 아니라 잘못된 요청이며, 값 추가는 `sha` / `instructions` 파라미터까지 끌고 오는 기능 확장이다.
- 이 결정을 명시적으로 적는 이유: 카드가 "`baseBranch` 로 조용히 좁히지 말라"고 지시했기 때문이다. 좁히지 않되, 좁히지 않는 방식이 기능 추가가 아니라 안전 요구라는 것이 여기서의 답이다.

### Out of Scope — `gates.codex: required` 의 강제 (제보 #3)

- `internal/cli/mcp_codex.go` `applyGateUnmet` 이 fail-open `inconclusive` 에 `GateUnmet` 주석을 단다(커밋 `5cfefe2cc`, 카드 t234). 판정 자체는 의도적으로 건드리지 않는다 — 간극을 **보이게** 만들 뿐 차단하지 않는다.
- "강제하거나 키 이름을 바꾸라"는 제보자의 요구는 여전히 열린 **정책** 질문이며, 이 SPEC 의 작업이 아니다. 관련은 있으나 별개다.

### Out of Scope — 제보 #1 / #2 / #4 (이미 착지)

- #1 구조화 findings: `codexFindingsOf` 로 착지(`4fe2c54c0`, t234).
- #2 verdict 합성: SPEC-CODEX-VERDICT-SYNTH-001 로 착지(`410da655f`, t229).
- #4 legacy `auth.json` 형태: `dbca3f710` 로 착지(t234). 제보자 환경에서 여전히 `unknown` 인지는 **미측정**이며 이 SPEC 이 재측정하지 않는다.

### Out of Scope — 제보 #5 (즉시 timeout 레코드)

- 재현이 시도된 적이 없다. `internal/cli/codex_task.go:67` 이 그 문자열의 유일한 기록자라는 것 외에 관측이 없다.
- 미측정 항목을 명세로 끌어들이면 근거 없는 요구가 되므로 범위 밖에 둔다. 결함이 없다는 뜻이 아니라, **아직 아무도 재지 않았다**는 뜻이다.

---

## §F 제약

- 측정된 스키마는 codex-cli **0.150.1** 기준이다. 제보자는 0.149.0 을 썼다. 0.149.0 에서만 다른 차이가 있다면 이 트리에서는 보이지 않는다(잔여 위험). AC-CRT-010 의 라이브 왕복도 실행 머신의 codex 판번호에서만 성립한다.
- fail-open 계약은 유지한다: codex 부재·오류는 여전히 `inconclusive` 다. REQ-CRT-004 / 008 이 바꾸는 것은 **codex 가 멀쩡한데 트리에서 base 를 못 찾은 경우**의 처리이며, 그 처리 형태(도구 오류 vs 원인을 명명한 inconclusive)는 plan.md §B 가 후보를 제시하고 run 이 고정한다.
- `internal/cli/codex_review_gate.go:90` 은 `uncommittedChanges` 를 하드코딩하므로 이 변경의 영향을 받지 않는다 — 다만 REQ-CRT-006 회귀선이 그 경로도 함께 지킨다.
- **codex 가 `baseBranch` 값을 어떻게 해석하는지는 스키마에 없다.** 설명문("Review changes between the current branch and the given base branch")은 로컬 브랜치인지 임의 revision 인지를 말하지 않는다. REQ-CRT-003 이 ref 해석 가능성을 요구하고 AC-CRT-010 이 실물 수용 여부를 재지만, codex 내부의 해석 규칙 자체는 이 SPEC 이 관측하지 못한다.
- §A.7 의 정렬 결정에 따라 `git_strategy.worktree_base_branch` 는 이 경로에서 읽히지 않는다. 그 키를 기준으로 삼고 싶다면 **두 백엔드를 함께** 바꾸는 별도 카드다.

---

## §G 참조

- `.moai/reports/t399/discovery.md` — 이 SPEC 의 일차 측정 기록
- `.moai/reports/t399/schema/v2/ReviewStartParams.json` — 프로토콜 계약 원문
- `.moai/reports/t229/succession.md` — 축 2 가 갈라진 경위(카드 번호는 큐에서)
- `internal/cli/mcp_codex.go` · `internal/cli/mcp_server.go` · `internal/cli/mcp_review_material.go`

**라이브 왕복 선례 (AC-CRT-010 이 재사용하는 것 — 새 발명이 아님):**

- `internal/cli/codex_live_protocol_probe_test.go:507` `TestCodexLive_ReviewStartEmitsTurnStarted` — 실 codex 세션을 열어 `review/start` 를 라이브로 보내고(`:533`), `turn/started` 를 보는 즉시 세션을 끊어 리뷰 turn 전체를 청구서에서 뺀다. 부속 키트: `probeLiveCodex` · `probeSeedRepo`(`:548`, 임시 git repo 생성) · `probeInstallRunner` · `probeWriteTranscript`.
- `internal/cli/codex_review_gate_live_test.go:33` — skip 조건 3종 확립본(바이너리 부재 / `--version` 비정상 / `MOAI_SKIP_LIVE_CODEX=1`). codex 없는 CI 가 그대로 통과한다.

`probeSeedRepo` 는 **미커밋 변경**을 심으므로 `uncommittedChanges` 용이다. `baseBranch` 는 base 브랜치 + 그로부터 갈라진 HEAD 가 필요하므로 픽스처 한 변형이 든다 — 키트의 확장이지 대체가 아니다.
