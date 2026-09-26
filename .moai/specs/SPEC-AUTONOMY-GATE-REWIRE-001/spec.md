---
id: SPEC-AUTONOMY-GATE-REWIRE-001
title: "계약 기반 자율 하네스 A3 — contract 모드 게이트 재배선, Kickoff 자율 승인, contract decide·revoke"
version: "0.2.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "CLAUDE.md, .claude/rules/moai/core, .claude/rules/moai/workflow, .claude/skills/moai, internal/template/templates, internal/template, internal/contract, internal/cli"
lifecycle: spec-anchored
tags: "autonomy, contract-mode, gate-rewire, kickoff, autonomous-kickoff, contract-decide, contract-revoke, receipt-store, tamper-evidence, guided-preservation, template-mirror, lifecycle"
tier: L
depends_on: [SPEC-AUTONOMY-CONTRACT-001, SPEC-AUTONOMY-ESCALATION-001, SPEC-ALWAYS-LOADED-DIET-002]
related_specs: [SPEC-AUTONOMY-CONTRACT-001, SPEC-AUTONOMY-ESCALATION-001, SPEC-JEV-CORE-001, SPEC-ALWAYS-LOADED-DIET-002, SPEC-AUTONOMY-TIERS-001]
---

# SPEC-AUTONOMY-GATE-REWIRE-001 — contract 모드 게이트 재배선 + Kickoff 자율 승인 (카드 t1236, AUTONOMY-A3)

## HISTORY

| 날짜 | 버전 | 변경 | 작성 |
|---|---|---|---|
| 2026-09-26 | 0.2.0 | 리드 범위 추가 세 건(운영자 결정 2026-09-26, A1 분할). **(1) Kickoff 자율 승인** — contract 모드에서 계약 서명을 사람 TTY 서명자가 아니라 결정자 쌍(주 LLM + Jev)이 수행; 여섯 기계 전제조건, 합의 규칙, 작성자 배제, Jev 단독 금지(A5 전). **(2) `moai contract revoke <card>`** — A1 에서 넘어온 Go 코드; needs-decision 은 큐 상태가 아니라 A2 형식 에스컬레이션 기록 1건으로 표현(리드 결정); revoke 존재가 자율 Kickoff 활성의 선행 조건. **(3) `moai contract decide <card>` (가칭)** — 에이전트가 쓴 영수증은 위조 가능하므로 moai 가 Jev 를 직접 호출해 원문을 보관하고 `~/.moai` 아래 추가 전용 해시 체인 저장소에 영수증을 발급; A1 의 receipt 기반 서명은 그 저장소의 기록만 인정(A1 의존). 변조 **방지**가 아니라 변조 **흔적**이 목표임을 명시. **(4) 전제 변경** — t1235(A2) 도 run 진입 전 develop 병합 필수(push 직렬화 강제·에이전트발 `sign` 차단이 A2 소관); `depends_on` 에 A1·A2·t1175 SPEC 추가. 요구사항 21 → **25**(상한 맞춰 080·090 과 072·073 통합, 013 을 010 에 흡수). A2 스키마에 기대는 항목에 **[A2 스키마 재확인]** 표시 추가. 발견된 충돌 3건(A1 서명기 TTY 전용·에이전트 거부, `SPEC-JEV-CORE-001` REQ-JEVC-012 의 「운영자 게이트 입력 금지」, `mission-governor` 의 「NOT for: approval」)을 `research.md §9` 와 NC-7·NC-8 로 기록 | manager-spec |
| 2026-09-26 | 0.1.0 | 최초 작성 (plan phase). 기준 트리 develop `ca1d5dc43`. 계약 필드·설정 키·CLI 동사는 A1 초안(`SPEC-AUTONOMY-CONTRACT-001` design.md § Contract Schema, 커밋 `8f77d9a33`, plan-audit 이전)에 맞췄다. 그 초안에 기대는 REQ·AC 에는 **[A1 감사 통과본으로 재확인]** 표시를 달았다. 근거 조사 `research.md`, 설계 `design.md` | manager-spec |

## §A. 배경

계약 기반 자율 하네스는 여러 카드로 나뉜다.

| 카드 | 담당 | 이 SPEC 과의 관계 |
|---|---|---|
| t1234 (A1, `SPEC-AUTONOMY-CONTRACT-001`) | `contract.yaml` 스키마, `moai contract sign/show/verify`, 설정 `workflow.autonomy.*` | **선행.** 이 SPEC 은 A1 의 이름을 가리키고, A1 에 **비대화형 영수증 기반 서명 경로**를 요청한다(리드가 A1 에 전달, `research.md §9.1`). revoke 는 A1 에서 빠져 이 카드로 넘어왔다 |
| t1235 (A2, `SPEC-AUTONOMY-ESCALATION-001`) | 에스컬레이션 감지기, needs-decision 기록, push 직렬화·두 번째 리뷰 강제, 에이전트발 `contract sign` 차단 | **선행(리드 결정 — run 진입 전 develop 병합 필수).** needs-decision 은 A2 소유이며 큐 상태가 아니라 에스컬레이션 기록이다(리드 결정). revoke 는 그 형식으로 기록 1건을 쓴다. A1 의 구속 순서 제약상 A2 가 먼저 착지해야 서명이 Kickoff 를 대체할 수 있다 |
| A4 | 종료 보고(`review.human: closure-report`) | Closure 단계가 그 보고를 낸다는 자리만 둔다 |
| A5 | Jev 보정(calibration) | Jev 단독 결정자(`decider: jev`)는 A5 전까지 금지 |

운영자 결정(2026-09-26): **A-Q1** 계약 서명이 Implementation Kickoff Approval 을 대체한다 · **A-Q2** develop push 는 계약 안에서 허용하되 직렬화 · **A-Q3** 두 번째 모델 리뷰 필수, 수행되지 않으면 Push 전에 멈춘다 · **A-Q4** 배포 기본값은 `guided` · **Kickoff 자율 승인** — contract 모드의 서명은 결정자 쌍이 수행하고, 여섯 전제조건 중 하나라도 빠지면 사람에게 간다.

게이트 표(설계 문서, develop `ca1d5dc43` 실측 — 조항 위치와 등록 여부는 `research.md §2`·`§3`):

| 게이트 | 현재 | contract 모드 처분 |
|---|---|---|
| G1 Implementation Kickoff Approval (plan→run) | 본문은 Frozen/HARD 로 적혀 있으나 zone-registry 미등록 | **계약 서명**으로 전환. 서명 주체는 설정의 결정자: `human`(기본) 또는 결정자 쌍 `llm+jev`(활성 조건 충족 시) |
| G2 소크라테스 인터뷰 (Rule 5) | Evolvable, `CONST-V3R2-013` | 계약 서명에 흡수 |
| G3 접근 승인 (Rule 1) | Evolvable, `CONST-V3R2-014` | 계약의 `approach` 필드에서 1회 승인 |
| G4 가정 확인 대기 | Evolvable, `CONST-V3R2-030` | 가정을 기록하고 진행, 계약과 모순되면 에스컬레이션 |
| G7/G8 plan-audit FAIL → 사용자 선택 | HARD, 규칙 전용 | 상한(`audit_retries`, 기본 2)까지 자동 수리 후 에스컬레이션 |
| G11 sync 확인 질문 | Evolvable | 제거 (계약이 문서 범위를 고정) |
| G5·G14·G20 등 Frozen, G17/G18 헌법 | Frozen | **건드리지 않는다**. 자율 승인은 결정을 **기록**할 뿐 사용자에게 묻지 않는다 |

## §B. 범위

세 층이다.

1. **문서 층(추가형)** — 기존 guided 경로 텍스트는 한 바이트도 바꾸지 않고, contract 모드 지시를 구분 마커 블록으로 덧붙인다. 편집 대상 집합은 `design.md §2` 가 정본이다.
2. **Go 코드 층** — `moai contract decide <card>`(가칭)와 `moai contract revoke <card>` 두 동사, 그리고 둘이 공유하는 moai 소유 영수증 저장소. `design.md §7`~`§9`.
3. **가드 테스트** — contract-mode 블록의 상시 검사.

## §C. 요구사항 (GEARS)

Tier L 상한(요구사항 25)에 맞춰 **25개**다.

### 모드 분기와 guided 보존

- **REQ-GR-001** (Ubiquitous) — The 편집 대상 문서 shall contract 모드 지시를 `<!-- moai:contract-mode-start id="<slug>" -->` 와 `<!-- moai:contract-mode-end -->` 가 각각 한 줄을 차지하는 블록 안에만 두며, 그 블록은 어떤 `moai:evolvable-start` ~ `moai:evolvable-end` 구간 안에도 놓이지 않는다.
- **REQ-GR-002** (Ubiquitous) — The 편집 대상 문서 shall, contract-mode 블록(마커 줄 포함)을 제거하면 기준 ref `BASE` 의 같은 파일과 바이트 동일하다. `BASE` 는 run phase 진입 시 흡수한 develop 커밋이다(`plan.md §C`).
- **REQ-GR-003** (Where) — **Where** `workflow.autonomy.mode` 가 `guided` 이거나, 키가 없거나, 무효한 값인 경우, the 오케스트레이터 지시문 shall 현행 guided 게이트를 그대로 적용한다 — contract 블록은 첫 문장에서 자신의 적용 조건을 `workflow.autonomy.mode: contract` 로 선언하고, contract 모드라도 Kickoff 결정자 설정이 없거나 `human` 이면 사람 서명 경로를 쓴다. **[A1 감사 통과본으로 재확인]** (설정 키 경로, 무효값 → guided 규칙, 결정자 키 이름)

### G1 — Kickoff 를 계약 서명으로

- **REQ-GR-010** (Where) — **Where** `workflow.autonomy.mode: contract` 인 경우, the plan→run 게이트 shall 해당 SPEC 의 서명된 `contract.yaml` 이다 — 오케스트레이터는 Implementation Kickoff Approval `AskUserQuestion` 을 내지 않고 `moai contract verify <SPEC-ID>` 의 exit 0(`state: signed-valid`)을 통과로 취급한다. 결정자가 `human` 이면 오케스트레이터는 서명 명령과 요약을 보고하고 턴을 닫는다(서명은 대화형 터미널에서 운영자가 한다). verify 가 exit 0 이 아니거나 `contract.yaml` 이 없으면 run phase 에 들어가지 않고 verify 의 `reasons` 를 담아 에스컬레이션한다 — guided 로 조용히 떨어지지 않는다. **[A1 감사 통과본으로 재확인]** (CLI 동사, exit 코드, `state` 값, reason 코드 집합)
- **REQ-GR-011** (Ubiquitous) — The 신규 SSOT 규칙 shall 「Implementation Kickoff Approval 이 통과했다/얻었다」를 전제로 하는 MoAI 규칙·스킬·에이전트·출력 스타일의 모든 참조가 contract 모드에서는 검증된 계약 서명으로 충족된다는 등가 조항을 담고, G1 **발화 지점**(`research.md §1.2` 의 E)은 각각 contract-mode 블록을 가지며 **참조 지점**(R)은 이 등가 조항으로 덮여 편집되지 않는다.
- **REQ-GR-015** (Where) — **Where** contract 모드인 경우, the plan 워크플로 shall manager-spec 위임 지시에 서명 전 초안 `contract.yaml`(`signature` 블록 없음)을 SPEC 산출물과 함께 내라는 지시를 포함한다. **[A1 감사 통과본으로 재확인]** (초안 필드 집합)

### Kickoff 자율 승인 (contract 모드 한정)

- **REQ-GR-016** (Where) — **Where** 결정자 설정이 `llm+jev` 인 경우, the Kickoff 자율 승인 shall 다음이 **모두** 참일 때만 활성이고, 하나라도 거짓이면 사람 서명 경로로 간다: (i) `moai contract revoke` 와 `moai contract decide` 가 사용 가능, (ii) A1 의 영수증 기반 서명 경로가 이 SPEC 의 영수증 저장소 기록만 받아들임, (iii) A1 이 구속 순서로 요구한 A2 의 push 직렬화·두 번째 리뷰 정지·에이전트발 대화형 `sign` 거부가 착지, (iv) `SPEC-JEV-CORE-001` REQ-JEVC-012 의 개정이 착지(NC-7). 결정자 값 `jev`(Jev 단독)와 `llm`(LLM 단독)은 이 SPEC 에서 유효하지 않으며 `human` 으로 취급한다 — Jev 단독은 A5 보정 전까지 금지된다. 활성 순서는 **revoke → moai 발급 영수증(decide + 저장소) → A2 push 직렬화 → A2 에이전트발 `sign` 차단 → 그 뒤에야 자율 Kickoff** 이며, 이 SPEC 은 A2 의 두 부분이 무엇을 어떻게 막는지 정하지 않고 존재 여부만 조건으로 쓴다. **[A1 감사 통과본으로 재확인]** (결정자 설정 키, 영수증 기반 서명 경로) **[A2 스키마 재확인]** (push 직렬화 강제, 에이전트발 `sign` 차단의 제공 형태)
- **REQ-GR-017** (When) — **When** `moai contract decide <card>` 가 실행될 때, the 명령 shall 판단을 열기 전에 여섯 전제조건을 기계적으로 평가하고, 하나라도 성립하지 않으면 Jev 를 호출하지 않고 결과 `human` 영수증을 남긴다: (a) plan-audit 판정이 PASS 이고 판정 파일의 점수가 Tier 문턱(S 0.75 / M 0.80 / L 0.85) 이상, (b) SPEC 파일의 `[NEEDS CLARIFICATION` 표지 0건, (c) `moai contract verify` 가 서명 부재 외 사유 없이 계약 완결(수락 기준 해시·소유권·행동·예산 존재), (d) 계약 `actions` 에 되돌릴 수 없는 행동(main·release·tag 계열) 없음, (e) 해당 카드의 열린 에스컬레이션 기록 0건, (f) 계약 `ownership.write` 경로와 헌법 Frozen 항목의 파일 집합의 교집합이 공집합. **[A1 감사 통과본으로 재확인]** (verify 사유 코드, 금지 행동 토큰, glob 의미론) **[A2 스키마 재확인]** (열린 에스컬레이션 판독)
- **REQ-GR-018** (Ubiquitous) — The Kickoff 결정 shall 두 신호의 합의로 정해진다: 주 LLM 판단(승인/거절/사람에게 에스컬레이션, 계약 줄별 사유)과 Jev 신호(「이 계약으로 시작해도 되는가」, 선택지와 기준 포함). 둘 다 승인이면 결과 `approve`; 어느 한쪽이 거절 또는 에스컬레이션이면 `on_disagree` 설정에 따라 `human`(기본) 또는 `reject`; Jev 가 사용 불가이거나 신뢰도가 `jev_min_confidence`(기본 0.50) 미만이면 「측정되지 않음」으로 `human`. **SPEC 작성 에이전트는 결정자가 될 수 없다** — 판단 제출자의 에이전트 유형이 `manager-spec` 이거나 그 식별자가 SPEC 의 plan-phase 커밋 `Authored-By-Agent:` 트레일러와 같으면 결과는 `human`(사유 `author-decider-conflict`)이다. **[A1 감사 통과본으로 재확인]** (`on_disagree`·`jev_min_confidence` 키 이름)
- **REQ-GR-019** (When) — **When** `moai contract decide <card> --spec <SPEC-ID> --judgement <file|->` 가 실행될 때, the 명령 shall 영수증을 기록하면 결과(`approve`/`human`/`reject`)와 무관하게 exit 0, 저장소 무결성 검증 실패로 아무것도 쓰지 않으면 exit 1, 사용법·입력 형식·I/O 오류로 아무것도 쓰지 않으면 exit 2 를 내고, 결과는 `--json` 출력의 `outcome` 필드로 전달한다. 명령은 계약에 서명하지 않고, `contract.yaml`·SPEC 파일·큐·git 상태를 바꾸지 않으며, 워크트리 안에 쓰지 않고, LLM 을 호출하지 않는다(주 LLM 판단은 입력으로 받을 뿐이다). 이름 `decide` 는 가칭이다.
- **REQ-GR-024** (Ubiquitous) — The Kickoff·revoke 영수증 shall moai 가 소유하는 워크트리 밖 저장소 `~/.moai/db/<project-key>/contract/receipts.jsonl` 에 추가 전용으로 기록되며, 각 기록은 직전 기록의 해시를 담아 체인을 이루고, 결정자 신원(에이전트 유형·모델·세션), SPEC 작성자 신원, 입력 파일 해시(`contract.yaml`·`acceptance.md`·plan-audit 판정 파일·판단 파일), 각 결정자의 답과 사유, Jev 원시 요청·응답 본문과 그 해시, 신뢰도, 전제조건 평가 결과, 최종 결과를 담는다. 카드 증거 경로 `.moai/reports/<card>/kickoff-receipt.json` 은 사본이며 권위가 없다. 목표는 변조 **흔적**이다 — 단일 사용자 로컬 환경에서 변조 **방지**는 불가능하고, 이 SPEC 은 그것을 주장하지 않는다. 영수증 형식의 서명 측 소비는 A1 이 정한다. **[A1 감사 통과본으로 재확인]** (영수증을 소비하는 `sign` 경로의 필드 요구)

### G2 / G3 / G4 — 인터뷰·접근 승인·가정 대기

- **REQ-GR-020** (While) — **While** 서명된 계약 범위 안의 작업이 진행 중인 동안, the 오케스트레이터 shall 소크라테스 인터뷰(Rule 5)와 접근 승인(Rule 1)을 계약 서명으로 충족된 것으로 취급하고(접근은 계약의 `approach` 필드가 1회 승인한 것이다), 가정은 진행 기록(`progress.md`)에 적고 확인을 기다리지 않고 진행한다. **[A1 감사 통과본으로 재확인]** (`approach` 필드)
- **REQ-GR-022** (When) — **When** 기록한 가정이나 새로 드러난 모호함이 계약의 `acceptance`·`invariants`·`ownership`·`approach` 와 모순될 때, the 오케스트레이터 shall 진행을 멈추고 `escalate_on` 의 해당 종류(`contradictory-evidence` 등)로 에스컬레이션한다. **[A1 감사 통과본으로 재확인]** (필드명, `escalate_on` 토큰)

### G7 / G8 — plan-audit FAIL

- **REQ-GR-030** (When) — **When** contract 모드에서 plan-auditor 가 FAIL 을 내거나 SPEC 품질 게이트(Phase 15)가 WARNING/FAIL 을 낼 때, the 오케스트레이터 shall 수리 위임과 재감사를 상한까지 자동으로 반복하고, 상한에 이르면 사용자 선택 질문 대신 에스컬레이션 보고를 낸다. 상한은 초안 `contract.yaml` 의 `budget.audit_retries`, 없으면 `workflow.autonomy.escalation.budget_default.audit_retries`(기본 2)다 — 이 시점에는 계약이 아직 서명 전이다. **[A1 감사 통과본으로 재확인]** (`budget.audit_retries`, `budget_default` 키)

### G11 — sync 확인 질문

- **REQ-GR-040** (Where) — **Where** contract 모드인 경우, the sync 워크플로 shall 문서 범위 승인(gate-sync-2), 다음 단계 질문, 「현재 브랜치에서 동기화할까」 질문을 내지 않고(문서 범위는 계약이 고정한다), 실패 경로(테스트 실패·보안 critical·호환성 파괴·CI 미러 실패)는 질문 대신 에스컬레이션 보고로 라우팅한다.

### 유지되는 게이트와 에스컬레이션

- **REQ-GR-050** (Ubiquitous) — The contract-mode 블록과 두 CLI 동사 shall Frozen 게이트(질문 채널 독점, CI autofix 3회 후 질문, 컨텍스트 한도 `/clear`)와 헌법 조항을 완화·재해석·재서술하지 않으며 — 자율 승인과 revoke 는 결정을 **기록**할 뿐 사용자에게 묻지 않는다 — SSOT 는 유지 게이트를 명시한다: sync-auditor must-pass, `main`/release 동작은 계약으로 허용될 수 없음(금지 토큰), 카드 선택은 운영자, goal 상한, 파괴적 명령 확인. **[A1 감사 통과본으로 재확인]** (행동 어휘, 금지 토큰)
- **REQ-GR-052** (Ubiquitous) — The 에스컬레이션 보고 shall Report-Before-Ask 게이트의 보고 형식(출처별 발견·정량·근거 경로)을 재사용하며, 보고를 남기고 멈추는 것이지 `AskUserQuestion` 이 아니다. needs-decision 은 큐 상태가 아니라 A2 형식의 에스컬레이션 기록 존재로 표현되며, 감지와 기록 형식은 t1235 소관이다. **[A2 스키마 재확인]**

### 한 호흡 생명주기

- **REQ-GR-060** (Ubiquitous) — The run 워크플로 문서 shall contract 모드 생명주기의 Discovery → RED → GREEN → Qualification 을, the sync 워크플로 문서 shall Closure → Integration → Push 를 이 순서로 담는다.
- **REQ-GR-062** (Ubiquitous) — The 각 단계 shall 자신의 증거가 기록되고 해당 카드에 열린 에스컬레이션 기록(revoke 가 쓴 것 포함)이 없을 때만 다음 단계로 넘어간다 — Discovery 는 계약의 `reobserve` 목록 재관측과 재현, RED 는 구현 전에 커밋된 실패 테스트, Qualification 은 범위 테스트·lint·커버리지·변이·두 번째 모델 리뷰(`audit_multi`, 모델은 `review.second_model`), Closure 는 상태 전이와 `.moai/reports/<card>/verdict.md`, Integration 은 develop 흡수 후 병합 트리 재측정. **[A1 감사 통과본으로 재확인]** (`reobserve`, `review.second_model`) **[A2 스키마 재확인]** (열린 기록 판독)
- **REQ-GR-063** (While) — **While** Push 단계에 있는 동안, the develop push shall 계약 `actions` 에 `push-develop` 이 있고 설정 `workflow.autonomy.contract.push_develop` 가 참이며 통합 창을 잡은 상태에서만 직렬로 수행되고, `workflow.autonomy.contract.second_review: required` 인데 두 번째 모델 리뷰 증거가 없으면 Push 전에 멈추고 에스컬레이션한다. **[A1 감사 통과본으로 재확인]** (행동 토큰, 설정 키, 창 요건 함의 규칙)

### moai contract revoke

- **REQ-GR-091** (When) — **When** `moai contract revoke <card>` 가 실행될 때, the 명령 shall 그 카드의 영수증 저장소에 `approve` 결과가 있으면 revoke 기록을 체인에 추가하고 A2 형식의 에스컬레이션 기록을 정확히 1건 써서 카드를 needs-decision 으로 되돌리며 exit 0 을 낸다; 이미 revoke 되어 그 뒤 새 `approve` 가 없으면 아무것도 쓰지 않고 exit 0(멱등); 시작된 run 이 없으면(`approve` 영수증 없음) exit 1; 사용법·I/O·저장소 무결성 오류는 exit 2 를 낸다. 진행 중인 run 은 다음 단계 경계에서 멈춘다(REQ-GR-062). **[A2 스키마 재확인]** (에스컬레이션 기록 경로·형식·revoke 종류)
- **REQ-GR-092** (Unwanted) — The `moai contract revoke` shall not 워크트리를 지우거나, 브랜치를 지우거나 이름을 바꾸거나, push 하거나, 큐를 바꾸거나, `contract.yaml`·서명·SPEC 파일을 고치거나, 프로세스를 종료하거나, git 쓰기 명령을 실행한다.

### 로컬·템플릿 동등과 중립성

- **REQ-GR-070** (Ubiquitous) — The 각 contract-mode 블록 shall 로컬 사본과 템플릿 사본에서 바이트 동일하고(신규 SSOT 는 통째로 동일), 블록 밖의 기존 로컬·템플릿 분기는 `BASE` 대비 변하지 않는다.
- **REQ-GR-072** (Unwanted) — The contract-mode 블록과 신규 SSOT shall not SPEC ID, REQ/AC 토큰, 카드 id, 내부 날짜, 커밋 SHA 를 담으며, 템플릿 테스트 패키지는 블록의 짝 맞춤·비중첩·evolvable 구간 밖 배치·금지 클래스 부재를 상시 검사하고 검사 대상 블록이 0개면 실패한다.

### 범위 제한

- **REQ-GR-080** (Unwanted) — The 이 SPEC 의 편집 shall not always-loaded 파일을 파일당 40,000자 이상으로 키우거나 always-loaded 편집 파일 합계를 `design.md §4` 의 상한보다 많이 늘리며, `moai-constitution.md`, `zone-registry.md`, `.claude/agents/**`, `.claude/output-styles/**`, `ci-autofix-protocol.md`, `context-window-management.md`, `agent-common-protocol.md` 와 그 템플릿 사본, 그리고 `backlog.db` 큐 스키마를 수정하지 않는다.

## §D. 범위 밖

### Out of Scope — Frozen 게이트와 헌법

- 질문 채널 독점(G5), CI autofix 3회 후 질문(G14), 컨텍스트 한도에서의 사용자 `/clear`(G20), 헌법 조항(G17/G18)과 그 개정 절차.
- `moai-constitution.md` Agent Core Behaviors 1 의 본문 — G4 처분은 SSOT 의 등가 조항으로만 표현한다(`plan.md §B` NC-3).
- zone-registry 등록 추가·수정. G1 을 등록하거나 orchestration-mode-selection 머리의 인라인 `[ZONE:Frozen]` 태그를 바꾸지 않는다.

### Out of Scope — A1·A2·A4·A5 소관

- `contract.yaml` 스키마, `moai contract sign/show/verify`, `workflow.autonomy.*` 설정 키의 구현, **영수증 기반 비대화형 서명 경로의 구현** — t1234 (이 SPEC 은 요청만 기록한다).
- 에스컬레이션 감지기, `escalate_on` 6종의 판정, needs-decision 기록 형식, push 직렬화와 두 번째 리뷰 규칙의 도구 호출 시점 강제, 에이전트발 대화형 `sign` 거부 — t1235.
- 종료 보고(`closure-report`)의 형식 — A4.
- Jev 보정과 Jev 단독 결정자(`decider: jev`) — A5. 이 SPEC 은 그 값을 활성화하지 않는다.

### Out of Scope — 큐 상태 추가

- `needs-decision` 을 큐(`backlog.db`)의 네 번째 상태로 추가하는 일. 큐 스키마는 세 상태로 고정되어 있고(`internal/kanban/backlog_schema_freeze_test.go`), 리드 결정으로 needs-decision 은 에스컬레이션 기록이다.

### Out of Scope — 변조 방지

- 단일 사용자 로컬 환경에서 영수증 위조를 막는 암호학적 보장(키 서명, 외부 공증). 이 SPEC 은 변조 흔적(해시 체인, Jev 원문 보관, 저장소에 없는 영수증 거부)만 제공한다.

### Out of Scope — 동음이의 게이트

- `e2e.md` 의 `--autofix` 「Kickoff Approval (one-time gate)」와 `harness-build-entry.md` 의 Builder 승인 게이트. 이름이 겹칠 뿐 plan→run 게이트가 아니다.

### Out of Scope — 에이전트·출력 스타일 편집

- `.claude/agents/**`(템플릿 수정 시 `make agents-emit` 로 Codex 사본을 재생성해야 하는 층)와 출력 스타일. 두 층의 Kickoff 언급은 참조 지점이며 REQ-GR-011 등가 조항이 덮는다. `mission-governor` 를 결정자로 재사용하려면 그 에이전트의 「NOT for: approval」과 충돌하므로 NC-8 에서 정한다.

### Out of Scope — 로컬 전용 하네스

- `.claude/agents/harness/workflow-specialist.md` — 템플릿 미러가 없는 사용자 소유 하네스.

### Out of Scope — 이미 한도를 넘은 규칙 파일

- `spec-workflow.md`(40,799자), `session-handoff-examples.md`(41,616자)는 참조 지점이며 편집하지 않는다 — t1175 의 「악화 금지」를 따른다.
