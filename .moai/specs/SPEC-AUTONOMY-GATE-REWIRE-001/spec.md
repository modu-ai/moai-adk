---
id: SPEC-AUTONOMY-GATE-REWIRE-001
title: "계약 기반 자율 하네스 A3 — contract 모드에서 Evolvable 게이트를 계약 서명 1회로 재배선"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "CLAUDE.md, .claude/rules/moai/core, .claude/rules/moai/workflow, .claude/skills/moai, internal/template/templates, internal/template"
lifecycle: spec-anchored
tags: "autonomy, contract-mode, gate-rewire, kickoff, contract-signing, guided-preservation, template-mirror, lifecycle"
tier: L
related_specs: [SPEC-AUTONOMY-CONTRACT-001, SPEC-ALWAYS-LOADED-DIET-002, SPEC-AUTONOMY-TIERS-001]
---

# SPEC-AUTONOMY-GATE-REWIRE-001 — contract 모드 게이트 재배선 (카드 t1236, AUTONOMY-A3)

## HISTORY

| 날짜 | 버전 | 변경 | 작성 |
|---|---|---|---|
| 2026-09-26 | 0.1.0 | 최초 작성 (plan phase). 기준 트리 develop `ca1d5dc43`. 계약 필드·설정 키·CLI 동사는 A1 초안(`SPEC-AUTONOMY-CONTRACT-001` design.md § Contract Schema, 커밋 `8f77d9a33`, plan-audit 이전)에 맞췄다. 그 초안에 기대는 REQ·AC 에는 **[A1 감사 통과본으로 재확인]** 표시를 달았다. 근거 조사 `research.md`, 설계 `design.md` | manager-spec |

## §A. 배경

계약 기반 자율 하네스는 여러 카드로 나뉜다.

| 카드 | 담당 | 이 SPEC 과의 관계 |
|---|---|---|
| t1234 (A1, `SPEC-AUTONOMY-CONTRACT-001`) | `contract.yaml` 스키마, `moai contract sign/show/verify`, 설정 `workflow.autonomy.*` | **선행.** 이 SPEC 은 A1 의 이름을 문서에서 가리키기만 한다. A1 은 「plan phase 에서 manager-spec 이 초안 `contract.yaml` 을 내는 것」을 A3 의 문서 변경으로 넘겼다(A1 spec.md § Out of Scope — Gate rewiring (A3)) |
| t1235 (A2) | 에스컬레이션 감지기, push 직렬화·두 번째 리뷰 강제 | **병행.** 이 SPEC 은 감지 결과를 어디로 보내는지만 문서화한다 |
| A4 | 종료 보고(`review.human: closure-report`) | 이 SPEC 은 Closure 단계가 그 보고를 낸다는 자리만 둔다 |

운영자 결정(2026-09-26): **A-Q1** 계약 서명이 Implementation Kickoff Approval 을 대체한다 · **A-Q2** develop push 는 계약 안에서 허용하되 직렬화 · **A-Q3** 두 번째 모델 리뷰 필수, 수행되지 않으면 Push 전에 멈춘다 · **A-Q4** 배포 기본값은 `guided`.

게이트 표(설계 문서, develop `ca1d5dc43` 실측 — 조항 위치와 등록 여부는 `research.md §2`·`§3` 에서 재확인):

| 게이트 | 현재 | contract 모드 처분 |
|---|---|---|
| G1 Implementation Kickoff Approval (plan→run) | 본문은 Frozen/HARD 로 적혀 있으나 zone-registry 미등록 | **계약 서명**으로 전환 |
| G2 소크라테스 인터뷰 (Rule 5) | Evolvable, `CONST-V3R2-013` | 계약 서명에 흡수 |
| G3 접근 승인 (Rule 1) | Evolvable, `CONST-V3R2-014` | 계약의 `approach` 필드에서 1회 승인 |
| G4 가정 확인 대기 | Evolvable, `CONST-V3R2-030` | 가정을 기록하고 진행, 계약과 모순되면 에스컬레이션 |
| G7/G8 plan-audit FAIL → 사용자 선택 | HARD, 규칙 전용 | 상한(`audit_retries`, 기본 2)까지 자동 수리 후 에스컬레이션 |
| G11 sync 확인 질문 | Evolvable | 제거 (계약이 문서 범위를 고정) |
| G5·G14·G20 등 Frozen, G17/G18 헌법 | Frozen | **건드리지 않는다** |

## §B. 범위

이 SPEC 은 문서 층의 **추가형 개정**과, 그 개정을 지키는 템플릿 테스트 하나다. 기존 guided 경로 텍스트는 한 바이트도 바꾸지 않고, contract 모드 지시는 구분 마커로 감싼 블록으로만 덧붙인다. 편집 대상 집합은 `design.md §2` 가 정본이다.

## §C. 요구사항 (GEARS)

Tier L 상한(요구사항 25)에 맞춰 21개로 묶었다.

### 모드 분기와 guided 보존

- **REQ-GR-001** (Ubiquitous) — The 편집 대상 문서 shall contract 모드 지시를 `<!-- moai:contract-mode-start id="<slug>" -->` 와 `<!-- moai:contract-mode-end -->` 가 각각 한 줄을 차지하는 블록 안에만 두며, 그 블록은 어떤 `moai:evolvable-start` ~ `moai:evolvable-end` 구간 안에도 놓이지 않는다.
- **REQ-GR-002** (Ubiquitous) — The 편집 대상 문서 shall, contract-mode 블록(마커 줄 포함)을 제거하면 기준 ref `BASE` 의 같은 파일과 바이트 동일하다. `BASE` 는 run phase 진입 시 흡수한 develop 커밋이다(`plan.md §C`).
- **REQ-GR-003** (Where) — **Where** `workflow.autonomy.mode` 가 `guided` 이거나, 키가 없거나, 무효한 값인 경우, the 오케스트레이터 지시문 shall 현행 guided 게이트를 그대로 적용한다 — contract 블록은 첫 문장에서 자신의 적용 조건을 `workflow.autonomy.mode: contract` 로 선언한다. **[A1 감사 통과본으로 재확인]** (설정 키 경로, 무효값 → guided 규칙)

### G1 — Kickoff 를 계약 서명으로

- **REQ-GR-010** (Where) — **Where** `workflow.autonomy.mode: contract` 인 경우, the plan→run 게이트 shall 해당 SPEC 의 서명된 `contract.yaml` 이다 — 오케스트레이터는 Implementation Kickoff Approval `AskUserQuestion` 을 내지 않고, `moai contract verify <SPEC-ID>` 의 exit 0(`state: signed-valid`)을 그 게이트의 통과로 취급하며, verify 가 exit 0 이 아니거나 `contract.yaml` 이 없으면 run phase 에 진입하지 않고 verify 의 `reasons` 를 담아 에스컬레이션한다 — guided 로 조용히 떨어지지 않는다. **[A1 감사 통과본으로 재확인]** (CLI 동사, exit 코드, `state` 값, reason 코드 집합)
- **REQ-GR-011** (Ubiquitous) — The 신규 SSOT 규칙 shall 「Implementation Kickoff Approval 이 통과했다/얻었다」를 전제로 하는 MoAI 규칙·스킬·에이전트·출력 스타일의 모든 참조가 contract 모드에서는 검증된 계약 서명으로 충족된다는 등가 조항을 담고, G1 **발화 지점**(`research.md §1.2` 의 E)은 각각 contract-mode 블록을 가지며 **참조 지점**(R)은 이 등가 조항으로 덮여 편집되지 않는다.
- **REQ-GR-013** (When) — **When** plan-audit 가 통과하고 contract 모드일 때, the 오케스트레이터 shall 서명 명령(`moai contract sign <SPEC-ID>`)과 서명 요약을 보고하고 턴을 닫는다 — 서명은 대화형 터미널에서 운영자가 하는 행위이며 오케스트레이터가 대신하지 않는다. **[A1 감사 통과본으로 재확인]** (서명 명령, 대화형 터미널 요건)
- **REQ-GR-015** (Where) — **Where** contract 모드인 경우, the plan 워크플로 shall manager-spec 위임 지시에 서명 전 초안 `contract.yaml`(`signature` 블록 없음)을 SPEC 산출물과 함께 내라는 지시를 포함한다. **[A1 감사 통과본으로 재확인]** (초안 필드 집합)

### G2 / G3 / G4 — 인터뷰·접근 승인·가정 대기

- **REQ-GR-020** (While) — **While** 서명된 계약 범위 안의 작업이 진행 중인 동안, the 오케스트레이터 shall 소크라테스 인터뷰(Rule 5)와 접근 승인(Rule 1)을 계약 서명으로 충족된 것으로 취급하고(접근은 계약의 `approach` 필드가 1회 승인한 것이다), 가정은 진행 기록(`progress.md`)에 적고 확인을 기다리지 않고 진행한다. **[A1 감사 통과본으로 재확인]** (`approach` 필드)
- **REQ-GR-022** (When) — **When** 기록한 가정이나 새로 드러난 모호함이 계약의 `acceptance`·`invariants`·`ownership`·`approach` 와 모순될 때, the 오케스트레이터 shall 진행을 멈추고 `escalate_on` 의 해당 종류(`contradictory-evidence` 등)로 에스컬레이션한다. **[A1 감사 통과본으로 재확인]** (필드명, `escalate_on` 토큰)

### G7 / G8 — plan-audit FAIL

- **REQ-GR-030** (When) — **When** contract 모드에서 plan-auditor 가 FAIL 을 내거나 SPEC 품질 게이트(Phase 15)가 WARNING/FAIL 을 낼 때, the 오케스트레이터 shall 수리 위임과 재감사를 상한까지 자동으로 반복하고, 상한에 이르면 사용자 선택 질문 대신 에스컬레이션 보고를 낸다. 상한은 초안 `contract.yaml` 의 `budget.audit_retries`, 없으면 `workflow.autonomy.escalation.budget_default.audit_retries`(기본 2)다 — 이 시점에는 계약이 아직 서명 전이다. **[A1 감사 통과본으로 재확인]** (`budget.audit_retries`, `budget_default` 키)

### G11 — sync 확인 질문

- **REQ-GR-040** (Where) — **Where** contract 모드인 경우, the sync 워크플로 shall 문서 범위 승인(gate-sync-2), 다음 단계 질문, 「현재 브랜치에서 동기화할까」 질문을 내지 않고(문서 범위는 계약이 고정한다), 실패 경로(테스트 실패·보안 critical·호환성 파괴·CI 미러 실패)는 질문 대신 에스컬레이션 보고로 라우팅한다.

### 유지되는 게이트와 에스컬레이션

- **REQ-GR-050** (Ubiquitous) — The contract-mode 블록 shall Frozen 게이트(질문 채널 독점, CI autofix 3회 후 질문, 컨텍스트 한도 `/clear`)와 헌법 조항을 완화·재해석·재서술하지 않으며, SSOT 는 유지 게이트를 명시한다: sync-auditor must-pass, `main`/release 동작은 계약으로 허용될 수 없음(금지 토큰), 카드 선택은 운영자, goal 상한, 파괴적 명령 확인. **[A1 감사 통과본으로 재확인]** (행동 어휘, 금지 토큰)
- **REQ-GR-052** (Ubiquitous) — The 에스컬레이션 보고 shall Report-Before-Ask 게이트의 보고 형식(출처별 발견·정량·근거 경로)을 재사용하며, 보고를 남기고 멈추는 것이지 `AskUserQuestion` 이 아니다. 감지와 카드 상태 표기는 t1235 소관이다.

### 한 호흡 생명주기

- **REQ-GR-060** (Ubiquitous) — The run 워크플로 문서 shall contract 모드 생명주기의 Discovery → RED → GREEN → Qualification 을, the sync 워크플로 문서 shall Closure → Integration → Push 를 이 순서로 담는다.
- **REQ-GR-062** (Ubiquitous) — The 각 단계 shall 자신의 증거가 기록되어야만 다음 단계로 넘어간다 — Discovery 는 계약의 `reobserve` 목록 재관측과 재현, RED 는 구현 전에 커밋된 실패 테스트, Qualification 은 범위 테스트·lint·커버리지·변이·두 번째 모델 리뷰(`audit_multi`, 모델은 `review.second_model`), Closure 는 상태 전이와 `.moai/reports/<card>/verdict.md`, Integration 은 develop 흡수 후 병합 트리 재측정. **[A1 감사 통과본으로 재확인]** (`reobserve`, `review.second_model`)
- **REQ-GR-063** (While) — **While** Push 단계에 있는 동안, the develop push shall 계약 `actions` 에 `push-develop` 이 있고 설정 `workflow.autonomy.contract.push_develop` 가 참이며 통합 창을 잡은 상태에서만 직렬로 수행되고, `workflow.autonomy.contract.second_review: required` 인데 두 번째 모델 리뷰 증거가 없으면 Push 전에 멈추고 에스컬레이션한다. **[A1 감사 통과본으로 재확인]** (행동 토큰, 설정 키, 창 요건 함의 규칙)

### 로컬·템플릿 동등과 중립성

- **REQ-GR-070** (Ubiquitous) — The 각 contract-mode 블록 shall 로컬 사본과 템플릿 사본에서 바이트 동일하고(신규 SSOT 는 통째로 동일), 블록 밖의 기존 로컬·템플릿 분기는 `BASE` 대비 변하지 않는다.
- **REQ-GR-072** (Unwanted) — The contract-mode 블록과 신규 SSOT shall not SPEC ID, REQ/AC 토큰, 카드 id, 내부 날짜, 커밋 SHA 를 담는다(템플릿 중립성 금지 클래스).
- **REQ-GR-073** (Ubiquitous) — The 템플릿 테스트 패키지 shall contract-mode 블록의 짝 맞춤·비중첩·evolvable 구간 밖 배치·금지 클래스 부재를 상시 검사하는 가드를 가지며, 검사 대상 블록이 0개면 실패한다.

### always-loaded 예산

- **REQ-GR-080** (Unwanted) — The 이 SPEC 의 편집 shall not always-loaded 파일을 파일당 40,000자 이상으로 키우거나, always-loaded 편집 파일 합계를 `design.md §4` 의 상한보다 많이 늘린다.

### 제외 파일

- **REQ-GR-090** (Unwanted) — The 구현 shall not `moai-constitution.md`, `zone-registry.md`, `.claude/agents/**`, `.claude/output-styles/**`, `ci-autofix-protocol.md`, `context-window-management.md`, `agent-common-protocol.md` 와 그 템플릿 사본을 수정한다.

## §D. 범위 밖

### Out of Scope — Frozen 게이트와 헌법

- 질문 채널 독점(G5), CI autofix 3회 후 질문(G14), 컨텍스트 한도에서의 사용자 `/clear`(G20), 헌법 조항(G17/G18)과 그 개정 절차.
- `moai-constitution.md` Agent Core Behaviors 1 의 본문 — G4 처분은 SSOT 의 등가 조항으로만 표현한다(`plan.md §B` NC-3).
- zone-registry 등록 추가·수정. G1 을 등록하거나 orchestration-mode-selection 머리의 인라인 `[ZONE:Frozen]` 태그를 바꾸지 않는다.

### Out of Scope — A1·A2·A4 소관

- `contract.yaml` 스키마, `moai contract` CLI, `workflow.autonomy.*` 설정 키와 기본값의 구현 — t1234.
- 에스컬레이션 감지기, `escalate_on` 6종의 판정, push 직렬화와 두 번째 리뷰 규칙의 도구 호출 시점 강제 — t1235.
- 종료 보고(`closure-report`)의 형식 — A4.

### Out of Scope — 동음이의 게이트

- `e2e.md` 의 `--autofix` 「Kickoff Approval (one-time gate)」와 `harness-build-entry.md` 의 Builder 승인 게이트. 이름이 겹칠 뿐 plan→run 게이트가 아니다.

### Out of Scope — 에이전트·출력 스타일 편집

- `.claude/agents/**`(템플릿 수정 시 `make agents-emit` 로 Codex 사본을 재생성해야 하는 층)와 출력 스타일. 두 층의 Kickoff 언급은 참조 지점이며 REQ-GR-011 등가 조항이 덮는다. 초안 `contract.yaml` 지시(REQ-GR-015)는 manager-spec 에이전트 본문이 아니라 오케스트레이터의 위임 지시에 싣는다(`plan.md §B` NC-5).

### Out of Scope — 로컬 전용 하네스

- `.claude/agents/harness/workflow-specialist.md` — 템플릿 미러가 없는 사용자 소유 하네스.

### Out of Scope — 이미 한도를 넘은 규칙 파일

- `spec-workflow.md`(40,799자), `session-handoff-examples.md`(41,616자)는 참조 지점이며 편집하지 않는다 — t1175 의 「악화 금지」를 따른다.
