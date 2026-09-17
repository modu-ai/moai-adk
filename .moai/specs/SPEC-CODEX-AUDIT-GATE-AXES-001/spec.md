---
id: SPEC-CODEX-AUDIT-GATE-AXES-001
title: "codex 감사 게이트 잔여 3축 — 단일 codex_audit 의 required 차단, 도구 호출 증거 없는 판정 거부, auth_provider unknown 재현"
version: "0.1.0"
status: draft
created: 2026-09-18
updated: 2026-09-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, audit-gate, required-gate, fail-open, auth-provider, issue-1632, t686"
---

# SPEC-CODEX-AUDIT-GATE-AXES-001 — codex 감사 게이트 잔여 3축

카드: **t686** · 외부 이슈: **modu-ai/moai-adk#1632** 잔여분 · 카드 클래스: **C** (정책 결정 1건 포함)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-18 | manager-spec | 초안 — 3축 (a) 단일 codex_audit required 차단, (b) 도구 호출 증거 없는 판정, (c) auth_provider unknown 재현 우선 |

## §A Context

### A.1 이슈에서 남은 것

#1632 은 다섯 항목을 보고했고, 그중 #0(native baseBranch), #1(구조화 findings), #2(판정 합성), #5(가짜 타임아웃 기록)는 이미 develop 에 착지했다(이슈 댓글 기준). 남은 것은 **#3 게이트 미집행**과, 유지자가 "환경에서 재확인하지 않았다"고 명시한 **#4 auth_provider unknown** 이다. #3 은 성격이 다른 두 결함으로 나뉜다.

- **(a) 단일 `codex_audit` 경로의 required 미차단.** 다중 경로(`audit_multi`)는 카드 t580 에서 "명시적 required 이면 fail-closed" 운영자 결정을 받아 실제로 overall verdict 를 fail 로 올린다. 단일 경로는 주석만 단다 — verdict 는 `inconclusive` 그대로다.
- **(b) 감사 도구를 한 번도 부르지 않은 판정.** 제보자의 관측: plan-auditor 가 3회 반복 동안 codex 게이트를 건너뛰고 PASS 0.92 를 냈고, 게이트를 프롬프트로 강제하자 결함 6건이 나와 FAIL 0.79 로 뒤집혔다. sync-auditor 에서도 재발했다. 현재 어떤 메커니즘도 이를 거부하지 않는다.
- **(c) `codex_setup` 의 `auth_provider: "unknown"`.** dbca3f710(t234)에서 auth_mode 없는 레거시 auth.json 분기를 추가했으나 제보자 구성으로 확인된 적이 없다.

### A.2 실측 근거 (develop @ f67d2193f)

| # | 사실 | 위치 |
|---|------|------|
| F1 | 단일 경로 주석 함수는 verdict 가 inconclusive 이고 raw `workflow.audit.gates.codex` 가 `required` 일 때만 `gate_unmet` 을 채우며 verdict 는 바꾸지 않는다 | `internal/cli/mcp_codex.go` `applyGateUnmet` (L1699-1714), 호출 L1671(바이너리 부재)·L1693(RPC 후) |
| F2 | `gate_unmet` 필드는 additive + omitempty | `mcp_codex.go` `ReviewOutput.GateUnmet` (L284-297) |
| F3 | 다중 경로는 명시적 required 의 inconclusive 를 overall=fail 로 올리되 per-backend verdict 는 inconclusive 로 보존한다. 판별은 raw 설정값이며 배포 기본값(codex required)은 opt-in 으로 치지 않는다 | `internal/cli/mcp_convergence.go` `enforceRequiredGateUnmet` (L774-810) |
| F4 | 다중 경로의 codex 호출은 단일 핸들러를 거치지 않는다(`performCodexAudit` 가 RPC seam 을 직접 호출) | `mcp_convergence.go` L545-561 |
| F5 | codex Stop-hook 게이트는 단일 핸들러를 거치지 않고 RPC 를 직접 부르며, 바이너리 부재·inconclusive 는 ALLOW | `internal/cli/codex_review_gate.go` `HandleCodexReviewGate` |
| F6 | 다중 리뷰 Stop-hook 게이트는 opt-in(기본 off)이고, 수렴 결과 상태 파일이 없으면 ALLOW 한다 — 감사 도구를 부르지 않은 세션이 정확히 이 분기로 통과한다 | `internal/cli/multi_review_gate.go` 결정 순서 (4) |
| F7 | 게이트 열거형 off / advisory / required | `internal/config/audit_models.go` L32-47 |
| F8 | auth 분류는 2단: auth.json 구조 판독 → `codex login status` 양 스트림의 전체-줄 문법(`logged in using (chatgpt|api key)`), 판독 불가는 `unknown` | `mcp_codex.go` `classifyCodexAuthFile` (L1927 부근), `classifyCodexAuth` (L2061 부근) |
| F9 | 이슈 본문·댓글에 제보자의 auth.json 형태나 `codex login status` 출력은 **없다**. 확인되는 것은 codex-cli 0.149.0, Linux, 바이너리 `/root/.local/bin/codex` 뿐 | `gh issue view 1632` |

### A.3 소비자 — "차단"이 누구에게 무엇을 의미하는가

단일 `codex_audit` 결과를 읽는 쪽은 plan-auditor 와 sync-auditor(에이전트 본문이 `verdict` 로 판정)뿐이다. F4·F5 에 따라 수렴 엔진과 codex Stop-hook 게이트는 이 핸들러의 출력을 소비하지 않는다. 따라서 (a)의 변경은 두 감사자 에이전트가 받는 도구 결과에만 닿는다.

## §B Requirements (GEARS)

### 축 (a) — 단일 codex_audit 의 required 차단

#### REQ-CAG-001
**Where** the audited tree's raw `workflow.audit.gates.codex` value is exactly `required`, **When** a `codex_audit` call would otherwise return a fail-open `inconclusive` verdict (codex binary absent, RPC failure, unresolvable target, or any other no-verdict cause), the `codex_audit` tool shall return a blocking result: `verdict` equal to `fail`, a non-empty `gate_unmet` field, and `isError` false with the structured content intact.

#### REQ-CAG-002
**When** `codex_audit` returns the blocking result of REQ-CAG-001, the result's `summary` shall name both the unmet required gate and the original no-verdict cause, so that a reader can distinguish "codex reviewed the change and failed it" from "the required gate was left unmet".

#### REQ-CAG-003
**Where** the raw `workflow.audit.gates.codex` value is `off`, `advisory`, absent, or unreadable, the `codex_audit` tool shall not alter its fail-open `inconclusive` output in any field — the serialized result shall be byte-identical to the pre-change output for the same input.

#### REQ-CAG-004
The `codex_audit` tool shall not treat the distributed engine default (codex `required` applied when the key is absent) as an explicit opt-in; only a value the project wrote counts, matching the discriminator the convergence engine already uses.

#### REQ-CAG-005
**When** codex returns a real `pass` or `fail` verdict, the `codex_audit` tool shall leave `verdict` and `gate_unmet` unchanged regardless of the gate value.

#### REQ-CAG-006
The `audit_multi` convergence result and the codex Stop-hook review gate shall remain unchanged by the axis (a) change — their per-backend verdicts, overall verdicts, and allow/block decisions for identical inputs shall match the pre-change behavior.

#### REQ-CAG-007
The `codex_audit` tool description and the plan-auditor / sync-auditor agent bodies shall state that an explicitly `required` codex gate left without a verdict yields `verdict: fail` with `gate_unmet` on the single-backend tool as well as on the convergence result.

### 축 (b) — 도구 호출 증거 없는 판정

#### REQ-CAG-008
**Where** the raw `workflow.audit.gates.codex` value is `required`, **When** a plan-audit or sync-audit verdict of PASS is produced for which no codex audit invocation (`codex_audit` or `audit_multi` with codex participating) is recorded, the audit enforcement mechanism selected at the Implementation Kickoff gate (plan.md §B) shall prevent that verdict from being accepted as PASS and shall name the missing invocation as the reason.

#### REQ-CAG-009
The invocation evidence consumed by REQ-CAG-008 shall be recorded by the moai runtime itself (MCP server or hook), not by text the auditing agent writes, so that an agent cannot satisfy the check without having made the call.

#### REQ-CAG-010
**Where** the raw `workflow.audit.gates.codex` value is not `required`, the axis (b) mechanism shall not block, warn, or alter any audit verdict.

#### REQ-CAG-011
**When** the axis (b) mechanism cannot read its own evidence store (missing directory, corrupt record, unreadable config), it shall fail open for non-required gates and shall report the unreadable evidence as a named gap rather than as a PASS for required gates.

### 축 (c) — auth_provider unknown

#### REQ-CAG-012
The axis (c) characterization test shall reproduce `auth_provider: "unknown"` from an input shape observed on the reporter's configuration (codex-cli 0.149.0) and shall fail against the unchanged classifier; the codex auth classification shall not be changed until that failure is observed.

#### REQ-CAG-013
**When** the codex auth input carries recognizable credential material in the observed shape of REQ-CAG-012, the classification shall report the matching provider (`chatgpt` or `api key`) instead of `unknown`.

#### REQ-CAG-014
**When** the codex auth input is genuinely unparseable, carries conflicting provider signals, or carries no credential material, the classification shall continue to report `unknown`.

#### REQ-CAG-015
The classification shall not read, log, return, or persist any credential value — only the provider kind.

## §C Constraints

- 모든 변경은 raw 설정값 `workflow.audit.gates.codex == required` 에만 반응한다. 설정하지 않은 프로젝트의 동작은 바이트 단위로 보존한다.
- 템플릿 파일은 16개 프로그래밍 언어에 중립이어야 하며 SPEC ID·카드 ID·날짜를 싣지 않는다(템플릿 중립성 가드).
- 게이트 기본값(배포 템플릿)은 바꾸지 않는다. 템플릿이 `gates.codex: required` 를 새로 싣는 일은 없다.

## §D Exclusions (What NOT to Build)

이 SPEC 의 범위 밖 항목은 아래와 같다. 세 축 밖의 변경은 하지 않는다.

### Out of Scope — 다중 경로와 Stop-hook 게이트

- `audit_multi` 수렴 엔진(`enforceRequiredGateUnmet`, 참여 백엔드 수 노출, `disagreement_flag` 의미) 변경 — 별도 카드 소관
- codex Stop-hook 게이트(`workflow.codex.review_gate`)가 required 설정을 읽도록 하는 변경 — 바이너리 부재 시 ALLOW 동작은 그대로 둔다
- 다중 리뷰 Stop-hook 게이트의 opt-in 기본값 변경

### Out of Scope — 이슈의 이미 닫힌 항목

- #0 native baseBranch target, #1 구조화 findings, #2 verdict 합성, #5 가짜 타임아웃 기록 — 이미 develop 착지
- `gates.claude` / `gates.glm` 의 단일 도구 차단 — 이 SPEC 은 codex 게이트만 다룬다

### Out of Scope — 이슈 운영

- 이슈 #1632 착지 댓글 작성과 이슈 종료 — 리드의 몫
- 제보자에게 auth 구성을 묻는 댓글 — 필요 시 리드가 판단(plan.md §B.3)
- `required` 키 이름 변경(advisory 로 개명) — 운영자 결정 t580 이 "required 는 차단"으로 확정했으므로 채택하지 않는다

## §E Traceability

| REQ | AC (acceptance.md) |
|-----|--------------------|
| REQ-CAG-001, 002 | AC-CAG-001, AC-CAG-002 |
| REQ-CAG-003, 004 | AC-CAG-003, AC-CAG-004 |
| REQ-CAG-005 | AC-CAG-005 |
| REQ-CAG-006 | AC-CAG-006 |
| REQ-CAG-007 | AC-CAG-007 |
| REQ-CAG-008, 009 | AC-CAG-008, AC-CAG-009 |
| REQ-CAG-010, 011 | AC-CAG-010, AC-CAG-011 |
| REQ-CAG-012 | AC-CAG-012 |
| REQ-CAG-013 | AC-CAG-013 |
| REQ-CAG-014, 015 | AC-CAG-014, AC-CAG-015 |
