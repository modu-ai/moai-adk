---
id: SPEC-CODEX-AUDIT-GATE-AXES-001
title: "codex 감사 게이트 잔여 2축 — 단일 codex_audit 의 required 차단, 영수증 없는 감사 PASS 거부"
version: "0.3.1"
status: implemented
created: 2026-09-18
updated: 2026-09-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, audit-gate, required-gate, fail-open, audit-receipt, subagent-stop, issue-1632, t686"
---

# SPEC-CODEX-AUDIT-GATE-AXES-001 — codex 감사 게이트 잔여 2축

카드: **t686** · 외부 이슈: **modu-ai/moai-adk#1632** 잔여분 · 카드 클래스: **C**

> 축 (c) `auth_provider: "unknown"` 은 v0.3.0 에서 카드 **t870** 으로 분리되었다(제보자 입력 대기). 이 SPEC 은 축 (a)·(b)만 다룬다.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-18 | manager-spec | 초안 — 3축 (a) 단일 codex_audit required 차단, (b) 도구 호출 증거 없는 판정, (c) auth_provider unknown 재현 우선 |
| 0.2.0 | 2026-09-18 | manager-spec | 운영자 결정(B.1=B-1, B.2=C-1, B.3 유지) + plan-audit iter1 FAIL 0.74 의 D1-D12 수리 |
| 0.3.0 | 2026-09-18 | manager-spec | 운영자 결정(축 (c) → t870 분리, K1 표면=S1 확정, N2 시작 표식+거부 기록+PreToolUse 소비자) + plan-audit iter2 FAIL 0.82 의 N1-N4, O1-O4 수리. 요구사항 17 → 16 |
| 0.3.1 | 2026-09-18 | manager-spec | plan-audit iter3 최종 문구 수정(운영자 Kickoff 승인, 추가 감사 없음): R1 판정 줄 누락 시 `unknown-spec` 거부 기록, K4 같은 역할의 유효 PASS 가 그 역할의 모든 거부 기록 해제 + 수동 삭제 허용, R2 누출 검사 강화, O7 로컬 전체 스위트 제거, O5·O10·게이트 미설정 잔여 위험 추가, O6 인용 줄 번호, O8·O9 판정 줄 규칙 |

## §A Context

### A.1 이슈에서 남은 것

#1632 의 #0, #1, #2, #5 는 develop 에 착지했다(이슈 댓글 기준). 이 SPEC 은 **#3 게이트 미집행**의 두 결함을 다룬다.

- **(a) 단일 `codex_audit` 경로의 required 미차단.** 다중 경로(`audit_multi`)는 카드 t580 의 운영자 결정("명시적 required 는 fail-closed")에 따라 overall verdict 를 fail 로 올린다. 단일 경로는 주석만 단다.
- **(b) 감사 도구를 한 번도 부르지 않은 판정.** 제보자 관측: plan-auditor 가 3회 반복 동안 codex 게이트를 건너뛰고 PASS 0.92 를 냈고, 게이트를 프롬프트로 강제하자 결함 6건이 나와 FAIL 0.79 로 뒤집혔다. sync-auditor 에서도 재발했다.

### A.2 실측 근거 (develop @ f67d2193f; plan-audit iter1·iter2 재확인)

| # | 사실 | 위치 |
|---|------|------|
| F1 | 단일 경로 주석 함수는 verdict 가 inconclusive 이고 raw `workflow.audit.gates.codex` 가 `required` 일 때만 `gate_unmet` 을 채우며 verdict 는 바꾸지 않는다 | `internal/cli/mcp_codex.go:1699-1714`, 호출 :1671·:1693 |
| F2 | `gate_unmet` 은 additive + omitempty | `mcp_codex.go:284-297` |
| F3 | 다중 경로는 명시적 required 의 inconclusive 를 overall=fail 로 올리고 per-backend verdict 는 보존. 판별은 raw 설정값 | `internal/cli/mcp_convergence.go:774-810` |
| F4 | 다중 경로의 codex 호출은 단일 핸들러를 거치지 않는다 | `mcp_convergence.go:545-561`; 단일 핸들러 참조는 등록부 `mcp_server.go:279` 뿐 |
| F5 | codex Stop-hook 게이트는 RPC 를 직접 부르며 바이너리 부재·inconclusive 는 ALLOW | `internal/cli/codex_review_gate.go:78-102` |
| F6 | 다중 리뷰 Stop-hook 게이트는 opt-in(기본 off), 상태 파일이 없으면 ALLOW | `internal/cli/multi_review_gate.go:44-79` |
| F7 | 게이트 열거형 off / advisory / required | `internal/config/audit_models.go:32-47` |
| F8 | 런타임 plan-audit 게이트 `GateConfig.Invoke` 에는 프로덕션 호출자가 없다 | `internal/runtime/audit_gate.go:199`, 유일 호출 :307 |
| F9 | SubagentStop 은 배선돼 있고(timeout 5) 핸들러 출력은 최상위 `decision: "block"` 을 실을 수 있다. 입력 선언 필드: `cwd`, `agent_type`, `agent_id`, `agent_transcript_path`, `last_assistant_message`. 현재 핸들러는 차단하지 않는다 | `.claude/settings.json:201-207`, `internal/hook/subagent_stop.go:38`, `internal/hook/types.go:212,230,238-241,369-372` |
| F10 | SubagentStart 핸들러는 로그와 additionalContext 만 내며 아무것도 저장하지 않는다 | `internal/hook/subagent_start.go:58-76` |
| F11 | PreToolUse 는 `Agent|Task` 매처로 배선돼 있고, 핸들러가 Agent/Task 스폰에서 가드를 돌려 deny 를 반환할 수 있다(에이전트 모델 가드 선례) | `.claude/settings.json:69` PreToolUse `"matcher": "Agent|Task"`, `internal/hook/pre_tool.go:632-640`, `internal/hook/agent_model_guard.go:237` |
| F12 | MCP 도구 `project_root` 는 `EvalSymlinks` 로 정규화되고, 생략 시 서버 폴백으로 해석된다 | `internal/cli/mcp_project_root.go:179`, :100-104 |
| F13 | `codex_audit` 는 읽기 전용 힌트 주석을 달고 있고 카탈로그상 `WriteCapable: false` 다. `audit_multi` 는 같은 분류에서 이미 상태 파일을 쓴다 | `mcp_server.go:278`, `internal/mcp/catalog.go:51`, `mcp_convergence.go:749` |

### A.3 소비자 — "차단"이 누구에게 무엇을 의미하는가

단일 `codex_audit` 결과를 읽는 쪽은 plan-auditor 와 sync-auditor 에이전트(Claude 본문, 그리고 템플릿에서 방출되는 `internal/template/templates/.codex/agents/moai/{plan-auditor,sync-auditor}.toml`)다. F4·F5 에 따라 수렴 엔진과 codex Stop-hook 게이트는 이 핸들러의 출력을 소비하지 않는다.

### A.4 이 SPEC 이 덮어쓰는 선행 계약 (해당 SPEC 은 수정하지 않는다)

| 선행 SPEC | 위치 | 기존 계약 | 처분 |
|-----------|------|-----------|------|
| SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 (completed) | spec.md:163-165 REQ-CBR-009 | required 게이트에서 빈 출력 inconclusive 에 `GateUnmet` 주석 | 주석 조항 유지, verdict 조항 대체. 그 AC 테스트 `codex_blank_review_test.go:419-420` 의 `verdict == inconclusive` 단언이 `fail` 로 바뀐다 |
| SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 | spec.md:221-222 Out of Scope | applyGateUnmet 을 바꾸지 않음, 게이트 집행은 형제 카드 소관 | 이 SPEC 이 그 형제 카드다 |
| SPEC-CODEX-REVIEW-TARGET-001 (completed) | spec.md:207 Out of Scope | "판정 자체는 의도적으로 건드리지 않는다" | t580 결정과 REQ-CAG-001 이 닫는다 |

### A.5 운영자 결정 (2026-09-18, 확정)

- **B-1 영수증**: MCP 서버가 호출마다 영수증을 기록, 감사 보고가 인용, required 일 때 서버가 아는 영수증 없는 PASS 는 거부.
- **B.3**: 차단 표현은 `verdict: fail` + `gate_unmet` + `isError: false`.
- **K1 = S1**: 영수증 검사는 SubagentStop 훅에서 실행.
- **N2**: SubagentStart 가 감사자 시작 표식을 기록하고, SubagentStop 이 첫 종료를 막으면서 거부를 상태 파일에 영속하며, 재진입에서는 기록·경고만 한다. 거부 파일은 run·sync 단계 진입의 기계 소비자가 읽고 막는다 — 소비자는 기존 PreToolUse `Agent|Task` 가드 경로(F11)다.
- **축 (c)**: 카드 t870 으로 분리.

## §B Requirements (GEARS)

### 축 (a) — 단일 codex_audit 의 required 차단

#### REQ-CAG-001
**Where** the audited tree's raw `workflow.audit.gates.codex` value is exactly `required`, **When** a `codex_audit` call would otherwise return a fail-open `inconclusive` verdict (codex binary absent, RPC failure, blank review output, unresolvable target, or any other no-verdict cause), the `codex_audit` tool shall return a blocking result: `verdict` equal to `fail`, a non-empty `gate_unmet` field, and `isError` false with the structured content intact.

#### REQ-CAG-002
**When** `codex_audit` returns the blocking result of REQ-CAG-001, the result's `summary` shall name both the unmet required gate and the original no-verdict cause.

#### REQ-CAG-003
**Where** the raw `workflow.audit.gates.codex` value is `off`, `advisory`, absent, or unreadable, the `codex_audit` tool result shall be byte-identical, after normalizing `build_commit` and `build_lag`, to the pre-change result for the same input, including after axis (b) lands.

#### REQ-CAG-004
The `codex_audit` tool shall not treat the distributed engine default (codex `required` applied when the key is absent) as an explicit opt-in.

#### REQ-CAG-005
**When** codex returns a real `pass` or `fail` verdict, the `codex_audit` tool shall leave `verdict` and `gate_unmet` unchanged regardless of the gate value.

#### REQ-CAG-006
The `audit_multi` convergence result and the codex Stop-hook review gate shall keep their verdict fields and allow/block decisions unchanged for identical inputs; the only permitted change on the convergence result is the additive receipt field of REQ-CAG-009.

#### REQ-CAG-007
The `codex_audit` tool description, the plan-auditor and sync-auditor agent bodies, and the `moai-ref-cross-model-audit` skill shall state that an explicitly `required` codex gate left without a verdict yields `verdict: fail` with `gate_unmet` on the single-backend tool as well as on the convergence result, and the skill shall not claim that the convergence engine reuses the single-backend handlers.

### 축 (b) — 영수증 없는 감사 PASS 거부 (B-1, 표면 S1)

#### REQ-CAG-008
The moai MCP server shall record a receipt for every `codex_audit` call and every `audit_multi` call in which codex participated, in the receipt store of the call's canonical tree root (the canonicalized `project_root` argument, or the server's fallback root when the argument is absent), with the schema defined in plan.md §B.5.

#### REQ-CAG-009
**Where** the raw `workflow.audit.gates.codex` value is `required`, the `codex_audit` result and the `audit_multi` result shall carry the identifier of the receipt recorded for that call; **Where** it is not `required`, the result shall not carry the receipt field.

#### REQ-CAG-010
**When** a SubagentStart event arrives whose `agent_type` is `plan-auditor` or `sync-auditor`, the SubagentStart hook shall write a start marker keyed by `agent_id` that records the agent type, session, start time, and the auditor's canonical tree root (plan.md §B.5).

#### REQ-CAG-011
**Where** the auditor's tree has the raw `workflow.audit.gates.codex` value `required`, **When** a plan-auditor or sync-auditor SubagentStop arrives with `stop_hook_active` false and its final message either carries no parseable verdict line (plan.md §B.6) or carries a PASS verdict line whose cited receipts fail the check of plan.md §B.6 (start marker missing, no receipt cited, receipt unknown to the store, receipt for a different canonical tree root, or receipt created before the matching start marker), the SubagentStop hook shall return `decision: "block"` with a reason naming the failed condition and shall persist a rejection record for that auditor role — keyed by the cited SPEC, or by `unknown-spec` when no verdict line could be parsed.

#### REQ-CAG-012
**When** the same SubagentStop condition as REQ-CAG-011 (including the unparseable-verdict-line case) arrives with `stop_hook_active` true, the SubagentStop hook shall not block, shall keep the rejection record persisted (writing it if absent), and shall emit a `systemMessage` stating that the PASS is not accepted and that phase-entry spawns stay denied until a PASS with a valid receipt is recorded.

#### REQ-CAG-013
**When** a plan-auditor or sync-auditor PASS passes the check of plan.md §B.6 in a `required` tree, the SubagentStop hook shall remove every rejection record for that auditor role, including records for other SPECs and `unknown-spec` records of that role; manual deletion of a rejection record by a person shall also clear it.

#### REQ-CAG-014
**Where** the spawning tree has the raw `workflow.audit.gates.codex` value `required`, **When** an `Agent` or `Task` spawn of `manager-develop`, `manager-docs`, or `manager-git` is attempted while any rejection record exists in that tree's receipt store, the PreToolUse hook shall deny the spawn with the `AUDIT_RECEIPT_VIOLATION` sentinel and name each outstanding rejection's auditor role, SPEC, and cause.

#### REQ-CAG-015
The receipts, start markers, and rejection records shall be written only by the moai runtime (MCP server or hook), never inferred from text an agent writes; **Where** the raw `workflow.audit.gates.codex` value is not `required`, the SubagentStart, SubagentStop, and PreToolUse additions shall write no start marker or rejection record and shall emit no block, deny, or warning.

#### REQ-CAG-016
**When** the evidence for a check cannot be read in a `required` tree (receipt store record corrupt, final message empty or without a parseable verdict line, rejection record unparseable), the hooks shall not accept the PASS or allow the spawn on that basis and shall name the unreadable item — an empty or unparseable final message follows REQ-CAG-011 and REQ-CAG-012 with an `unknown-spec` rejection record; an absent rejection directory shall count as no outstanding rejection.

## §C Constraints and Residual Risk

- 모든 차단·거부는 raw 설정값 `workflow.audit.gates.codex == required` 에만 반응한다. 설정하지 않은 프로젝트의 도구 결과는 빌드 식별 필드 정규화 후 바이트 동일하다(영수증은 저장소에만 남는다).
- 템플릿 파일은 16개 프로그래밍 언어에 중립이며 SPEC ID·카드 ID·날짜·이슈 번호를 싣지 않는다.
- 게이트 기본값(배포 템플릿)은 바꾸지 않는다.

잔여 위험(설계가 막지 못하는 것):
- 훅이 꺼진 환경(`disableAllHooks`, `allowManagedHooksOnly`)에서는 S1 과 PreToolUse 소비자 모두 무력하다.
- 오케스트레이터가 `manager-develop` / `manager-docs` / `manager-git` 을 스폰하지 않고 직접 구현·문서화·PR 을 하면 소비자를 우회한다.
- 거부 기록은 트리 단위라서, 같은 트리의 다른 SPEC 작업 스폰도 막는다(카드당 워크트리 운영에서는 영향이 작다).
- 거부 기록은 같은 역할의 다음 유효 PASS 가 자동으로 모두 해제하며(`unknown-spec` 포함), 사람이 파일을 지워도 해제된다. 역할 단위 해제이므로, 한 SPEC 의 유효 PASS 가 같은 역할의 다른 SPEC 거부를 함께 지운다. 수동 삭제는 어디에도 기록되지 않는다.
- 스폰 거부는 `subagent_type` 의 정확한 값(`manager-develop` / `manager-docs` / `manager-git`)에만 걸린다. 구현을 `general-purpose` 나 다른 이름공간의 에이전트로 스폰하면 거부되지 않는다(`internal/hook/agent_model_guard.go:97` 의 스폰 추출 방식).
- primary 체크아웃 세션이 워크트리 SPEC 을 `project_root` 로 감사하면(예: 칸반 리드가 레인 트리를 감사), 감사자 시작 표식의 `tree_root`(primary)와 영수증의 `tree_root`(워크트리)가 달라 항상 "다른 트리"로 거부된다. 감사자는 감사 대상 트리 안에서 실행해야 한다.
- 페이로드 필드는 선언만 확인했다. M1 실측에서 없으면 S2+S3 로 되돌린다(plan.md §F M1).

미검증(Gaps):
- 이 저장소의 `.moai/config/sections/workflow.yaml` 은 `workflow.audit.gates` 를 설정하지 않는다. 따라서 축 (b)의 required 경로는 이 저장소의 실제 세션에서 발동하지 않으며, 테스트 픽스처(`t.TempDir()` 트리에 게이트를 쓴 설정)로만 검증된다.
- SubagentStart / SubagentStop 런타임 페이로드는 M1 실측 전까지 관측되지 않았다.

## §D Exclusions (What NOT to Build)

이 SPEC 의 범위 밖 항목은 아래와 같다.

### Out of Scope — 분리된 축과 이미 닫힌 항목

- 축 (c) `codex_setup` 의 `auth_provider: "unknown"` — 카드 t870 소관(제보자 입력 대기)
- #0 native baseBranch target, #1 구조화 findings, #2 verdict 합성, #5 가짜 타임아웃 기록 — 이미 develop 착지

### Out of Scope — 다중 경로와 기존 게이트

- `audit_multi` 수렴 판정 변경 — 영수증 필드 추가(REQ-CAG-009) 외에는 건드리지 않는다
- codex Stop-hook 게이트가 required 를 읽도록 하는 변경, 다중 리뷰 Stop-hook 의 기본값 변경
- 사용되지 않는 `internal/runtime` plan-audit 게이트(`GateConfig.Invoke`) 배선
- `gates.claude` / `gates.glm` 에 대한 단일 도구 차단과 영수증
- 선행 SPEC(§A.4) 문서의 수정

### Out of Scope — 이슈 운영

- 이슈 #1632 댓글과 종료 — 리드의 몫
- `required` 키 이름 변경 — 운영자 결정으로 "required 는 차단" 확정

## §E Traceability

| 요구사항 식별자 | 인수 기준 (acceptance.md) |
|-----|--------------------|
| REQ-CAG-001, 002 | AC-CAG-001, AC-CAG-002 |
| REQ-CAG-003, 004 | AC-CAG-003, AC-CAG-004 |
| REQ-CAG-005 | AC-CAG-005 |
| REQ-CAG-006 | AC-CAG-006 |
| REQ-CAG-007 | AC-CAG-007 |
| REQ-CAG-008, 009 | AC-CAG-008 |
| REQ-CAG-010 | AC-CAG-009 |
| REQ-CAG-011 | AC-CAG-010 |
| REQ-CAG-012, 013 | AC-CAG-011 |
| REQ-CAG-014 | AC-CAG-012 |
| REQ-CAG-015 | AC-CAG-013 |
| REQ-CAG-016 | AC-CAG-014 |
