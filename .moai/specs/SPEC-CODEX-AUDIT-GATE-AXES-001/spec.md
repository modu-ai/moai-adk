---
id: SPEC-CODEX-AUDIT-GATE-AXES-001
title: "codex 감사 게이트 잔여 3축 — 단일 codex_audit 의 required 차단, 영수증 없는 PASS 거부, auth_provider unknown 재현"
version: "0.2.0"
status: draft
created: 2026-09-18
updated: 2026-09-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, audit-gate, required-gate, fail-open, audit-receipt, auth-provider, issue-1632, t686"
---

# SPEC-CODEX-AUDIT-GATE-AXES-001 — codex 감사 게이트 잔여 3축

카드: **t686** · 외부 이슈: **modu-ai/moai-adk#1632** 잔여분 · 카드 클래스: **C**

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-18 | manager-spec | 초안 — 3축 (a) 단일 codex_audit required 차단, (b) 도구 호출 증거 없는 판정, (c) auth_provider unknown 재현 우선 |
| 0.2.0 | 2026-09-18 | manager-spec | 운영자 결정 반영(B.1=B-1 영수증, B.2=C-1 리드가 제보자에게 요청·축 (c) 입력 대기, B.3 유지) + plan-audit iter1 FAIL 0.74 수리. D2 AC-CAG-006 범위 재설정, D3 덮어쓰는 선행 계약 2건 명시(§A.4)와 codex_blank_review_test.go 추가, D4 영수증 필드는 required 일 때만 결과에 싣도록 REQ-CAG-003 과 정합(REQ-CAG-009), D5 B-1 의 기계 검사 표면 실측·권장(plan.md §B.4), D6 provider 리터럴 `apiKey`, D8 골든 정규화·선행 커밋, D10 REQ-CAG-015 문구, D11 AC-CAG-004 전제 단언, D12 cross-model-audit 스킬 문서를 REQ-CAG-007 에 포함, D7 B-2 행 선언 필드 인용 |

## §A Context

### A.1 이슈에서 남은 것

#1632 은 다섯 항목을 보고했고, 그중 #0(native baseBranch), #1(구조화 findings), #2(판정 합성), #5(가짜 타임아웃 기록)는 이미 develop 에 착지했다(이슈 댓글 기준). 남은 것은 **#3 게이트 미집행**과, 유지자가 "환경에서 재확인하지 않았다"고 명시한 **#4 auth_provider unknown** 이다. #3 은 성격이 다른 두 결함으로 나뉜다.

- **(a) 단일 `codex_audit` 경로의 required 미차단.** 다중 경로(`audit_multi`)는 카드 t580 의 운영자 결정("명시적 required 는 fail-closed")에 따라 overall verdict 를 fail 로 올린다. 단일 경로는 주석만 단다 — verdict 는 `inconclusive` 그대로다.
- **(b) 감사 도구를 한 번도 부르지 않은 판정.** 제보자 관측: plan-auditor 가 3회 반복 동안 codex 게이트를 건너뛰고 PASS 0.92 를 냈고, 게이트를 프롬프트로 강제하자 결함 6건이 나와 FAIL 0.79 로 뒤집혔다. sync-auditor 에서도 재발했다. 현재 이를 거부하는 메커니즘은 없다.
- **(c) `codex_setup` 의 `auth_provider: "unknown"`.** dbca3f710(t234)에서 auth_mode 없는 레거시 auth.json 분기를 추가했으나 제보자 구성으로 확인된 적이 없다. **입력 대기** 상태다(§A.5).

### A.2 실측 근거 (develop @ f67d2193f, plan-audit iter1 이 F1-F8 재확인)

| # | 사실 | 위치 |
|---|------|------|
| F1 | 단일 경로 주석 함수는 verdict 가 inconclusive 이고 raw `workflow.audit.gates.codex` 가 `required` 일 때만 `gate_unmet` 을 채우며 verdict 는 바꾸지 않는다 | `internal/cli/mcp_codex.go:1699-1714` `applyGateUnmet`, 호출 :1671(바이너리 부재)·:1693(RPC 후) |
| F2 | `gate_unmet` 필드는 additive + omitempty | `mcp_codex.go:284-297` |
| F3 | 다중 경로는 명시적 required 의 inconclusive 를 overall=fail 로 올리되 per-backend verdict 는 보존한다. 판별은 raw 설정값이며 배포 기본값은 opt-in 으로 치지 않는다 | `internal/cli/mcp_convergence.go:774-810` `enforceRequiredGateUnmet` |
| F4 | 다중 경로의 codex 호출은 단일 핸들러를 거치지 않는다 | `mcp_convergence.go:545-561` `performCodexAudit`; 단일 핸들러의 유일한 참조는 등록부 `mcp_server.go:279` |
| F5 | codex Stop-hook 게이트는 RPC 를 직접 부르며 바이너리 부재·inconclusive 는 ALLOW | `internal/cli/codex_review_gate.go:78-102` |
| F6 | 다중 리뷰 Stop-hook 게이트는 opt-in(기본 off)이고, 수렴 결과 상태 파일이 없으면 ALLOW | `internal/cli/multi_review_gate.go:44-79` |
| F7 | 게이트 열거형 off / advisory / required | `internal/config/audit_models.go:32-47` |
| F8 | auth 분류는 2단: auth.json 구조 판독 → `codex login status` 양 스트림의 전체-줄 문법. 보고하는 provider 값은 `chatgpt` / `apiKey` / `unknown` 이다(`api key` 는 2단 문법의 줄 토큰일 뿐 보고값이 아니다) | `mcp_codex.go:136-138`(보고값), :1927 `classifyCodexAuthFile`, :2024(줄 문법), :2061 `classifyCodexAuth` |
| F9 | 이슈 본문·댓글에 제보자의 auth.json 형태나 `codex login status` 출력은 없다 | `gh issue view 1632` |
| F10 | 런타임 plan-audit 게이트 `GateConfig.Invoke` 에는 프로덕션 호출자가 없다(유일한 호출은 같은 파일의 `TeamModeInvoke` 자기 위임). run-phase 의 plan-audit 게이트는 오케스트레이터 산문이다 | `internal/runtime/audit_gate.go:199`, :307 |

### A.3 소비자 — "차단"이 누구에게 무엇을 의미하는가

단일 `codex_audit` 결과를 읽는 쪽은 plan-auditor 와 sync-auditor 에이전트(Claude 본문, 그리고 템플릿에서 방출되는 Codex 에이전트 `internal/template/templates/.codex/agents/moai/{plan-auditor,sync-auditor}.toml`)다. 이들은 `verdict` 로 판정한다. F4·F5 에 따라 수렴 엔진과 codex Stop-hook 게이트는 이 핸들러의 출력을 소비하지 않는다.

### A.4 이 SPEC 이 덮어쓰는 선행 계약 (해당 SPEC 은 수정하지 않는다)

| 선행 SPEC | 위치 | 기존 계약 | 이 SPEC 의 처분 |
|-----------|------|-----------|-----------------|
| SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 (completed) | spec.md:163-165 REQ-CBR-009 | required 게이트에서 빈 출력 inconclusive 에 `GateUnmet` 주석을 단다 | **주석 조항은 유지, verdict 조항은 대체.** REQ-CBR-009 는 verdict 를 말하지 않지만 그 AC 테스트(`codex_blank_review_test.go:419-420`)가 `verdict == inconclusive` 를 고정한다. REQ-CAG-001 이 required 일 때 그 verdict 를 `fail` 로 대체한다 |
| SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 | spec.md:221-222 Out of Scope | "applyGateUnmet 을 있는 그대로 소비하고 바꾸지 않는다", 게이트 집행은 형제 카드 소관 | 이 SPEC 이 그 형제 카드(t686)다. 경계 선언은 이 SPEC 착지로 해소된다 |
| SPEC-CODEX-REVIEW-TARGET-001 (completed) | spec.md:207 Out of Scope | "판정 자체는 의도적으로 건드리지 않는다" | 당시 열린 정책 질문으로 남겼던 항목을 t580 운영자 결정과 이 SPEC 의 REQ-CAG-001 이 닫는다 |

### A.5 운영자 결정 (2026-09-18, 확정)

- **B.1 = B-1 영수증.** MCP 서버가 `codex_audit` / `audit_multi` 호출마다 영수증을 기록하고, 감사 보고가 그것을 인용하며, 게이트가 `required` 일 때 서버가 아는 영수증을 인용하지 않은 PASS 는 거부한다.
- **B.2 = C-1.** 리드가 제보자에게 auth 형태를 요청한다(레인은 게시하지 않는다). 축 (c)는 **입력 대기**. SPEC 은 (a)+(b)로 sync 할 수 있고, 회신이 없으면 (c)는 기록된 gap 으로 닫는다.
- **B.3 유지.** 차단 표현은 `verdict: fail` + `gate_unmet` + `isError: false`.

남은 결정은 plan.md §B.4 의 **거부 검사가 실제로 실행되는 기계 표면** 1건이다.

## §B Requirements (GEARS)

### 축 (a) — 단일 codex_audit 의 required 차단

#### REQ-CAG-001
**Where** the audited tree's raw `workflow.audit.gates.codex` value is exactly `required`, **When** a `codex_audit` call would otherwise return a fail-open `inconclusive` verdict (codex binary absent, RPC failure, blank review output, unresolvable target, or any other no-verdict cause), the `codex_audit` tool shall return a blocking result: `verdict` equal to `fail`, a non-empty `gate_unmet` field, and `isError` false with the structured content intact.

#### REQ-CAG-002
**When** `codex_audit` returns the blocking result of REQ-CAG-001, the result's `summary` shall name both the unmet required gate and the original no-verdict cause, so that a reader can distinguish "codex reviewed the change and failed it" from "the required gate was left unmet".

#### REQ-CAG-003
**Where** the raw `workflow.audit.gates.codex` value is `off`, `advisory`, absent, or unreadable, the `codex_audit` tool result shall be byte-identical, after normalizing the build-identity fields `build_commit` and `build_lag`, to the pre-change result for the same input — including after axis (b) lands, because the receipt field of REQ-CAG-009 is not emitted for these gate values.

#### REQ-CAG-004
The `codex_audit` tool shall not treat the distributed engine default (codex `required` applied when the key is absent) as an explicit opt-in; only a value the project wrote counts, matching the discriminator the convergence engine already uses.

#### REQ-CAG-005
**When** codex returns a real `pass` or `fail` verdict, the `codex_audit` tool shall leave `verdict` and `gate_unmet` unchanged regardless of the gate value.

#### REQ-CAG-006
The `audit_multi` convergence result and the codex Stop-hook review gate shall keep their verdict fields and allow/block decisions unchanged by this SPEC for identical inputs; the only permitted change on the convergence result is the additive receipt field of REQ-CAG-009 under an explicit `required` codex gate.

#### REQ-CAG-007
The `codex_audit` tool description, the plan-auditor and sync-auditor agent bodies, and the cross-model audit reference skill (`moai-ref-cross-model-audit`) shall state that an explicitly `required` codex gate left without a verdict yields `verdict: fail` with `gate_unmet` on the single-backend tool as well as on the convergence result, and the skill shall not claim that the convergence engine reuses the single-backend handlers.

### 축 (b) — 영수증 없는 PASS 거부 (운영자 결정 B-1)

#### REQ-CAG-008
The moai MCP server shall record a receipt for every `codex_audit` call and every `audit_multi` call in which codex participated, in a runtime-owned store under the audited project root; each receipt shall carry a server-generated identifier, the audited tree, the creation time, and the codex verdict and `gate_unmet` value of that call.

#### REQ-CAG-009
**Where** the raw `workflow.audit.gates.codex` value is `required`, the `codex_audit` result and the `audit_multi` result shall carry the identifier of the receipt recorded for that call; **Where** it is not `required`, the result shall not carry the receipt field, so that REQ-CAG-003 holds.

#### REQ-CAG-010
**Where** the raw `workflow.audit.gates.codex` value is `required`, **When** a plan-auditor or sync-auditor completes with a PASS verdict that cites no receipt identifier, cites an identifier absent from the store, or cites a receipt for a different audited tree or one created before that auditor started, the enforcement surface selected at Kickoff (plan.md §B.4) shall reject the PASS and name which of these conditions failed.

#### REQ-CAG-011
The receipt evidence consumed by REQ-CAG-010 shall be written only by the moai runtime (MCP server or hook), never inferred from text the auditing agent writes, so that an agent cannot satisfy the check without having made the call.

#### REQ-CAG-012
**Where** the raw `workflow.audit.gates.codex` value is not `required`, the axis (b) enforcement shall not block, warn, or alter any audit verdict.

#### REQ-CAG-013
**When** the axis (b) enforcement cannot read its receipt store (missing directory, corrupt record) or the auditor's final message, it shall stay silent for non-required gates and, for a `required` gate, shall report the unreadable evidence as a named gap and shall not accept the PASS.

### 축 (c) — auth_provider unknown (입력 대기)

#### REQ-CAG-014
The axis (c) characterization test shall reproduce `auth_provider: "unknown"` from an input shape observed on the reporter's configuration (codex-cli 0.149.0) and shall fail against the unchanged classifier; the codex auth classification shall not be changed until that failure is observed.

#### REQ-CAG-015
**When** the codex auth input carries recognizable credential material in the observed shape of REQ-CAG-014, the classification shall report the matching provider value (`chatgpt` or `apiKey`) instead of `unknown`.

#### REQ-CAG-016
**When** the codex auth input is genuinely unparseable, carries conflicting provider signals, or carries no credential material, the classification shall continue to report `unknown`.

#### REQ-CAG-017
The classification shall not retain, log, return, or persist any credential value — only the provider kind.

## §C Constraints

- 모든 차단·거부 동작은 raw 설정값 `workflow.audit.gates.codex == required` 에만 반응한다. 설정하지 않은 프로젝트의 도구 결과는 빌드 식별 필드 정규화 후 바이트 단위로 보존한다(영수증 기록은 저장소에만 남는다).
- 템플릿 파일은 16개 프로그래밍 언어에 중립이어야 하며 SPEC ID·카드 ID·날짜·이슈 번호를 싣지 않는다.
- 게이트 기본값(배포 템플릿)은 바꾸지 않는다.

## §D Exclusions (What NOT to Build)

이 SPEC 의 범위 밖 항목은 아래와 같다. 세 축 밖의 변경은 하지 않는다.

### Out of Scope — 다중 경로와 Stop-hook 게이트

- `audit_multi` 수렴 판정(`enforceRequiredGateUnmet`, 참여 백엔드 수, `disagreement_flag` 의미) 변경 — 영수증 필드 추가(REQ-CAG-009) 외에는 건드리지 않는다
- codex Stop-hook 게이트(`workflow.codex.review_gate`)가 required 설정을 읽도록 하는 변경 — 바이너리 부재 시 ALLOW 는 그대로 둔다
- 다중 리뷰 Stop-hook 게이트의 opt-in 기본값 변경
- 사용되지 않는 `internal/runtime` plan-audit 게이트(`GateConfig.Invoke`)를 배선하는 일

### Out of Scope — 이슈의 이미 닫힌 항목과 다른 게이트

- #0 native baseBranch target, #1 구조화 findings, #2 verdict 합성, #5 가짜 타임아웃 기록 — 이미 develop 착지
- `gates.claude` / `gates.glm` 에 대한 단일 도구 차단과 영수증
- 선행 SPEC(§A.4) 문서 자체의 수정

### Out of Scope — 이슈 운영

- 이슈 #1632 착지 댓글, 제보자에게 auth 형태를 묻는 댓글, 이슈 종료 — 모두 리드의 몫
- `required` 키 이름 변경 — 운영자 결정이 "required 는 차단"으로 확정

## §E Traceability

| 요구사항 식별자 | 인수 기준 (acceptance.md) |
|-----|--------------------|
| REQ-CAG-001, 002 | AC-CAG-001, AC-CAG-002 |
| REQ-CAG-003, 004 | AC-CAG-003, AC-CAG-004 |
| REQ-CAG-005 | AC-CAG-005 |
| REQ-CAG-006 | AC-CAG-006 |
| REQ-CAG-007 | AC-CAG-007 |
| REQ-CAG-008, 009 | AC-CAG-008 |
| REQ-CAG-010, 011 | AC-CAG-009, AC-CAG-010 |
| REQ-CAG-012 | AC-CAG-011 |
| REQ-CAG-013 | AC-CAG-012 |
| REQ-CAG-014 | AC-CAG-013 |
| REQ-CAG-015 | AC-CAG-014 |
| REQ-CAG-016, 017 | AC-CAG-015, AC-CAG-016 |
