---
id: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001
title: "AskUserQuestion 의사결정 메모리 — 캡처·추론·적응형 추천 배치 (Standard tier)"
version: "0.1.0"
status: in-progress
created: 2026-06-24
updated: 2026-06-24
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/hook + .claude/rules/moai/core/askuser-protocol.md + .claude/agent-memory/"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "askuser, decision-memory, preference-inference, recommendation, advisory-hook, adaptive, template-neutrality"
---

# SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 — AskUserQuestion 의사결정 메모리 (Standard tier)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-24 | manager-spec | 초기 plan-phase 저술. Tier M. Standard tier 구현 (policy + advisory PostToolUse 캡처 훅 + 선호 메모리 계층 + 28일 TTL 감쇠). 적응형 추천 강도 (숙련도 기반 자동 분기). 5개 깊은 연구 각도 + 25개 학술 인용 통합. |

---

## §A. 배경 및 문제 정의

### §A.1 배경

유지보수자(GOOS행님)의 제안: "AskUserQuestion 사용자 의사결정을 로그로 기록하고, 사용자 의도를 추론하여, 과거 의사결정을 참조해 향후 AskUserQuestion 선택지에서 추천 옵션을 상단에 배치하라."

이 제안은 현재 `.claude/rules/moai/core/askuser-protocol.md`가 (1) 채널 독점, (2) Socratic 인터뷰 절차, (3) ToolSearch 사전 로드, (4) 옵션 서술 표준, (5) `(권장)` 라벨 규칙을 다루지만 — **의사결정의 지속성(persistence)**과 **과거 의사결정 기반 추천 배치**는 다루지 않는 간극을 메운다.

### §A.2 문제 정의

현재 AskUserQuestion은 **stateless**하다. 각 세션에서 동일한 결정(Tier 선택, 언어 선호, 에이전트 위임 선호, 노력 등급 등)을 반복 질문하며, `(권장)` 라벨은 정적 기본값에만 의존한다. 이는:

1. **사용자 피로** — 동일 결정 반복 질문은 질문 수 자체가 피로를 유발 (보조적 근거 unverified: CHI 2025 just-in-time vs batched 연구 라인)
2. **정보이익 미활용 (Fisher 정보)** — p(사용자가 A 선택)≈0.5인 결정 경계에서만 질문해야 하나, 현재는 모든 결정을 동등 질문
3. **추천의 비합리성** — `(권장)` 라벨이 "시스템이 밀고 싶은 것"이지 "이 사용자가 통계적으로 선호하는 것"이 아님

### §A.3 deep-research 근거 (요약; 상세는 research.md)

5개 각도(적응형 LLM 에이전트 / 선택 아키텍처 / 에이전트 코딩 도구 / 선호 추출 / just-in-time 추천+역효과)에서 25개 학술 논문 + 6개 산업 도구 교차검증. 4개 STRONG 원칙 도출:

1. **통합(consolidation) > 축적(accumulation)** — Mem0 (토큰 −90%, p95 지연 −91%), Generative Agents reflection
2. **just-in-time 결정경계 질문** — Fisher 정보 I=p(1−p) p=0.5 최대 (`[verified]` Murphy "Probabilistic Machine Learning" Ch.3); 보조적 근거 unverified: Pep ICML 2026 정적 질문 수렴 함정 (동일 연구 라인)
3. **안정 특성 vs 일시 상태 분리** — MemGPT 3계층 (core/recall/archival), OpenAI Agents SDK
4. **추천 + 투명한 이유 + 쉬운 opt-out = 단일 번들** — Sinha 투명성 (liking 2.79→3.51, p<.01); Beshears 메타분석 (Cohen's d=0.546); Loughrey (알고리즘-통제-사용자 역전 위험)

**상충 증거 (design.md §B 양면 문서화 의무)**: "추천 효과적"(d≈0.55) vs "과도 추천→자율성 침식+필터 버블+프라이버시 역화" — 버퍼: 회복 제어(opt-out/reset/exploration). "오래된 데이터 감쇠 필요"(Copilot 28일) vs "순진 시간 감쇠가 지속 신호 상실"(Koren) — 버퍼: 일시/지속 분리 + 멱법칙 감쇠.

---

## §B. 범위 (5개 컴포넌트, Standard tier)

| # | 컴포넌트 | 책임 |
|---|---------|------|
| C1 | **선호 메모리 계층** (`user_decision_memory`) | 기존 기술 교훈 `feedback` 메모리와 **분리**. upsert(append 금지). 3계층 검색: core(항상 로드된 사용자 프로필) / recall(최근 N 세션) / archival(검색). 추론 사실은 `inferred` 태그 + 정정 루프. |
| C2 | **askuser-protocol 추천 규칙** | (a) 발화 시점 = 정보이익 최대 결정 경계(추정 p≈0.5); (b) 질문 순서 = 정보이익 내림차순; (c) 추천 옵션 = 통계적 다수 합리적 기본값(시스템이 밀고 싶은 것 아님) + 서술은 추천이 성립하는 전제조건 명시. |
| C3 | **PostToolUse advisory 캡처 훅** (`user_decision_capture`) | AskUserQuestion tool_result → 선호 메모리 upsert. **ADVISORY / warn-first / fail-open** (절대 exit 2 차단 금지). 회복 턴에서 exit 0 (Recovery-Signal Carve-Out 준수). 자가 편집은 명시적 사용자 신호에서만 발화 (MemGPT discipline). |
| C4 | **감쇠 정책** | 멱법칙 감쇠(초기 가파른 후 평탄, Ebbinghaus/Murre) + 재사용 시 신선도 부양(간격 반복) + 도메인별 신선도 예산 + 28일 미사용 만료 + 사용시 리셋(Copilot). 일시/지속 분리로 핵심 선호 보존(Koren). |
| C5 | **회복 제어 + 맥락적 개인화 게이트** | "이번 세션 개인화 비활성화" 토글, 민감 작업(보안 코드, 일회성 쿼리)에 맥락적 강도 저하, "N일 전 데이터 기반" 공개, 정정 루프("이 추론이 틀리면 알려주세요"), 의도적 탐색/우연성(필터 버블 방지 — 보조적 근거 unverified: Iyendo serendipity 연구 라인; `[verified]` Loughrey autonomy erosion이 핵심 근거). |

---

## §C. GEARS 요구사항

> **GEARS 정준 표기** (current MoAI standard). `<subject>`는 시스템/컴포넌트/에이전트/함수/아티팩트 중 어느 명사도 가능. 레거시 `IF/THEN` 금지 — `When <event-detected>` 사용.

### REQ-ADM-001 (C1, 통합 원칙) — upsert 전용 선호 메모리

The preference memory layer **shall** store each user decision as a single upserted entry keyed by (domain, decision_key), replacing any prior entry for the same key rather than appending.

**Rationale**: 통합 > 축적 원칙(Mem0 토큰 −90% / Generative Agents reflection). append-only는 토큰 비용과 검색 노이즈를 선형 증가시킨다.

### REQ-ADM-002 (C1, 계층 분리) — 기술 교훈과의 분리

The preference memory layer **shall** be stored in a separate namespace (`memory/user_decisions/`) from the existing technical-lesson feedback memory (`memory/feedback_*.md` + `MEMORY.md`), so that user-decision facts and engineering-lesson facts remain independently queryable.

**Rationale**: 안정 특성 vs 일시 상태 분리(MemGPT 3계층, OpenAI Agents SDK cookbook). 혼합 계층은 회수 정밀도를 저하시킨다.

### REQ-ADM-003 (C1, 스키마) — 정준 엔트리 스키마

Each preference memory entry **shall** carry the fields: `fact`, `source_citation` (file:line or session_id), `valid_time`, `last_used`, `scope` (stable|transient), `domain`, and `confidence` (observed|inferred).

**Rationale**: 정정 루프(REQ-ADM-018)와 감쇠(REQ-ADM-011)는 스키마 필드에 의존한다. 미검증 추론 사실은 반드시 `confidence: inferred`로 태깅된다(verification-claim-integrity.md 준수).

### REQ-ADM-004 (C1, 검색 계층) — 3계층 검색

The orchestrator **shall** retrieve preference memory via a 3-tier cascade: (1) core tier (always-loaded user profile, 최근 빈번 사용 stable 사실); (2) recall tier (최근 N 세션 사실); (3) archival tier (전체 검색, 지연 허용).

**Rationale**: MemGPT core/recall/archival 3계층; OpenAI Agents SDK cookbook. 3계층은 항상-로드 비용을 최소화하면서 드문 사실 회수를 허용한다.

### REQ-ADM-005 (C2, 발화 시점) — 정보이익 정렬

**When** the orchestrator estimates a decision uncertainty p≈0.5 (maximum Fisher information I=p(1−p)) for an upcoming AskUserQuestion round, the orchestrator **shall** emit that question; **while** p is near 0 or 1 (near-certain), the orchestrator **shall** auto-handle the decision with the statistically-majority option and omit the question.

**Rationale**: just-in-time 결정경계 원칙 — Fisher 정보 I=p(1−p)는 p=0.5에서 최대 (`[verified]` Murphy "Probabilistic Machine Learning" Ch.3). 보조적 근거(unverified): Pep ICML 2026 정적 질문 수렴 함정 (미래 날짜 인용 — 동일 연구 라인으로 취급, 검증 제한).

### REQ-ADM-006 (C2, 질문 순서) — 정보이익 내림차순

**Where** multiple questions are batched in one AskUserQuestion call, the orchestrator **shall** order them by estimated per-question information gain, highest first.

**Rationale**: 정보이익 내림차순 정렬은 Fisher 정보 이론의 직접적 적용 (`[verified]` Murphy Ch.3 — 높은 정보이익 질문이 의사결정에 더 많은 정보를 제공). 보조적 근거(unverified): CHI 2025 just-in-time > batched (질문 수 자체가 피로 원인 — 구체 논문 검증 제한, 일반 연구 라인). 높은 정보이익 질문을 먼저 배치하면 사용자가 낮은 가치 질문을 만나기 전에 핵심 의사결정을 완료할 수 있다.

### REQ-ADM-007 (C2, 추천 옵션) — 통계적 다수 합리적 기본값

The recommended option (first option with `(권장)` label) **shall** be the statistically-majority rational default observed in the preference memory for the current domain, NOT a system-pushed preference; **where** no sufficient observation exists (cold-start, <N observations), the orchestrator **shall** fall back to the existing static default and disclose "based on static default, N observations needed for personalization" in the option description.

**Rationale**: Beshears 메타분석 기본값 효과(d=0.546)는 합리적 기본값에서 성립; 시스템 밀어넣기는 Loughrey 자율성 침식 위험. cold-start 공개는 verification-claim-integrity를 만족한다(미관측 추천 금지).

### REQ-ADM-008 (C2, 전제조건 서술) — 추천 성립 전제 명시

The recommended option's `description` **shall** state the precondition under which the recommendation holds, enabling the user to reject it trivially when the precondition is violated.

**Rationale**: 투명성 + 쉬운 opt-out 번들(Sinha liking 2.79→3.51; Beshears autonomy risk ∝ automation). 전제가 서술되지 않은 추천은 기형적 설계이다.

### REQ-ADM-009 (C3, advisory/fail-open) — 캡처 실패 시 exit 0

**When** the PostToolUse capture hook fails to parse the AskUserQuestion tool_result, or fails to upsert the preference memory, or encounters any internal error, the hook **shall** exit 0 (allow the turn to continue) and log a warning to `.moai/logs/hook-stderr.log`.

**Rationale**: advisory/fail-open 원칙. 기존 `sync-phase-quality-gate.sh` / `status-transition-ownership.sh` 만이 exit 2 차단 권한을 가진다. 캡처 훅은 선호 데이터 품질에 영향을 주지만 워크플로 진행을 차단할 수 없다.

### REQ-ADM-010 (C3, Recovery-Signal Carve-Out) — 회복 턴 exit 0 (SHOULD, doctrine-honest)

**While** a recovery turn is detected (detection mechanism is owned by a future runtime-layer SPEC per `runtime-recovery-doctrine.md §4` — the current advisory hook cannot parse `stopReason`), the capture hook **should** exit 0 without performing capture, to avoid placing recovery turns in the error→block→retry death-spiral.

> **Modality note (doctrine-honest SHOULD)**: 본 SPEC은 회복 턴에서의 *행동*(exit 0 + 캡처 미실행)만 정의한다. 회복 턴 *탐지 메커니즘*(stopReason 파싱)은 현재 advisory 훅이 수행할 수 없으며, `runtime-recovery-doctrine.md §4` + AP-RR-006에 따라 future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`로 이연된다. 따라서 본 REQ는 `should`(SHOULD)이지 `shall`(HARD)이 아니다. 탐지 메커니즘을 대체하는 proxy를 조작해선 안 된다 (AP-RR-006).

**Rationale**: `.claude/rules/moai/workflow/runtime-recovery-doctrine.md §4` Recovery-Signal Carve-Out + AP-RR-006 ("current hooks do NOT parse stopReason; no mechanical enforcement is possible without a runtime-layer hook that parses stopReason, deferred to a future SPEC"). 회복 턴에서 캡처 차단은 death-spiral의 교과서적 형태이나, 현재 layer에서는 documentation-only policy guidance(SHOULD)로만 정의할 수 있다.

### REQ-ADM-011 (C4, 멱법칙 감쇠) — 일시/지속 분리 감쇠

The decay policy **shall** apply a power-law decay function (initial steep, then plateau; Ebbinghaus/Murre) to transient-scope entries, while stable-scope entries **shall** be exempt from pure time-decay and instead refresh their `last_used` timestamp on each reuse (spaced-repetition boost).

**Rationale**: "오래된 데이터 감쇠 필요"(Copilot 28일) vs "순진 시간 감쇠가 지속 신호 상실"(Koren temporal dynamics) 상충 — 버퍼: 일시/지속 분리 + 멱법칙. 순진 time-decay는 핵심 선호를 잃는다.

### REQ-ADM-012 (C4, 28일 TTL) — 미사용 만료 + 사용시 리셋

**Where** a transient-scope entry has not been reused within 28 days, the decay policy **shall** expire (soft-delete) the entry; **when** the entry is reused, the policy **shall** reset its age counter to zero and boost its confidence weight.

**Rationale**: Copilot Memory 28-day TTL; 간격 반복(reset-on-use). 28일은 Copilot의 산업 검증 임계값이며, 사용시 리셋은 재사용되는 선호가 만료되지 않음을 보장한다.

### REQ-ADM-013 (C5, 회복 제어 토글) — 이번 세션 개인화 비활성화

**Where** the user signals "disable personalization this session" (via a `/moai preference toggle` command or an explicit instruction), the orchestrator **shall** suppress preference-based recommendation placement and uncertainty-based question omission for the remainder of the session.

**Rationale**: 회복 제어 번들(Sinha/Beshears/Loughrey). opt-out이 없는 추천은 자율성 침식이다.

### REQ-ADM-014 (C5, 맥락적 강도 저하) — 민감 작업 게이트

**While** the orchestrator is operating on a sensitive domain (security review, one-off exploratory query, or a domain flagged `cold-start`), the orchestrator **shall** reduce recommendation strength to neutral (no `(권장)` placement based on inferred preference) and disclose the reduction.

**Rationale**: 맥락적 개인화 게이트 — Loughrey `[verified]` (알고리즘-통제-사용자 역전 위험; 2차 욕구 침식) + Beshears `[verified]` (autonomy risk ∝ automation). 보조적 근거(unverified): Iyendo overspecialization 방지 (serendipity 연구 라인 — 구체 논문 검증 제한). 보안 코드에서 비전문가 추천은 위험하다.

### REQ-ADM-015 (C5, 데이터 신선도 공개) — "N일 전 데이터 기반"

**When** a recommendation draws on preference memory data, the option description **shall** disclose the data age ("based on N-day-old data") so the user can calibrate trust.

**Rationale**: 투명성 번들(Sinha). 신선도 공개 없는 추천은 verification-claim-integrity를 위반한다(관측되지 않은 신선도 주장).

### REQ-ADM-016 (C5, 정정 루프) — 추론 정정 채널

The orchestrator **shall** offer a correction channel ("이 추론이 틀리면 알려주세요" / "tell me if this inference is wrong") at each inferred-preference disclosure; **when** the user corrects, the orchestrator **shall** immediately upsert the corrected fact with `confidence: observed` and decrement the inferred entry's weight.

**Rationale**: 정정 루프(MemGPT self-editing discipline; Sinha transparency). 정정 불가능한 추론은 블랙박스이다.

### REQ-ADM-017 (적응형 추천 강도) — 숙련도 기반 자동 분기

**Where** the orchestrator infers high user proficiency (expert user — via session count, decision consistency, or explicit self-rating), the orchestrator **shall** apply weak recommendation strength (info-centric, autonomy-first — disclose inferred preference without `(권장)` label override); **where** the orchestrator infers low proficiency (general user), the orchestrator **shall** apply strong recommendation strength (default-like, `(권장)` label with transparent reason).

**Rationale**: 적응형 추천 강도 철학(사용자 확인). 전문가에게 강 추천은 info-centric 작업에서 자율성 침식(Loughrey); 일반 사용자에게 약 추천은 결정 피로 가중(Beshears). 자동 분기는 양쪽을 모두 만족한다.

### REQ-ADM-018 (verification-claim-integrity) — 추천 = 관측 메모리 증거 매핑

**Where** the orchestrator emits a recommendation that claims to reflect a past user decision, the recommendation **shall** map to an observed preference memory entry (`confidence: observed` or `inferred` with disclosed basis); the orchestrator **shall not** emit a recommendation based on an unobserved assumption.

**Rationale**: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (결함/부채/드리프트 주장은 도메인 도구로 검증). "추천이 과거 의사결정을 반영한다"는 주장은 관측된 메모리 증거에 매핑되어야 한다.

---

## §D. 비기능 요구사항 (제약)

| ID | 제약 | 근거 |
|----|------|------|
| NFR-ADM-001 | 캡처 훅 지연 ≤ 50ms (p95) | advisory 훅이 워크플로 병목이 되어서는 안 됨 |
| NFR-ADM-002 | 선호 메모리 core 계층 크기 ≤ 4KB (항상 로드 비용) | Mem0 토큰 절감 원칙 — core는 session-start에 로드됨 |
| NFR-ADM-003 | 추천 배치 결정 지연 ≤ 10ms (p95) | 인라인 결정 — AskUserQuestion 발화 직전에 평가 |
| NFR-ADM-004 | 감쇠 스캔 주기 = 1일 1회 (백그라운드) | 매 턴 감쇠 계산은 비용 과다 |
| NFR-ADM-005 | 회복 제어 토글 = 세션 단위 (영구 아님) | Loughrey 자율성 — 세션마다 재활성화 가능 |
| NFR-ADM-006 | 템플릿 중립성 — template-shipped 산출물은 내부 SPEC ID/REQ 토큰 포함 금지 | `.moai/docs/template-internal-isolation-doctrine.md §25` |

---

## §E. 범위 외 (Out of Scope)

> 본 SECTION은 `OutOfScopeRule` lint를 만족한다: "out of scope" 리터럴 + `### Out of Scope —` H3 + `-` 불릿.

### Out of Scope — Complete tier 기능 (향후 별도 SPEC)

- **Power-law refinement의 동적 파라미터 학습** — 본 SPEC은 고정 초기 파라미터만 사용; 사용자별 감쇠 곡선 학습은 "complete" tier로 이월
- **다중 사용자 프로필 분리** (팀 환경) — 본 SPEC은 단일 사용자(`~/.claude/projects/{hash}/memory/`) 가정
- **선호 메모리의 외부 동기화** (클라우드 백업, 팀 공유) — 프라이버시 민감; 별도 SPEC에서 프라이버시 모델 설계 후
- **강화학습 기반 추천 정책 학습** — offline structure learning + online Bayesian information-gain selection 전체 파이프라인은 "complete" tier (보조적 근거 unverified: Pep ICML 2026 연구 라인)

### Out of Scope — 기존 SPEC 영역 (중복 방지)

- **AskUserQuestion 채널 독점 / Socratic 인터뷰 절차 / ToolSearch 사전 로드** — `SPEC-ASKUSER-ENFORCE-001` (implemented, legacy v2.x) 소관. 본 SPEC은 의사결정 지속성 + 추천 배치만 다룬다.
- **기술 교훈 메모리 시스템** (`feedback_*.md`, `MEMORY.md` 인덱스) — 기존 자동 메모리 시스템 소관. 본 SPEC은 **분리된** `user_decisions/` 계층을 추가할 뿐 기존 계층을 수정하지 않는다.
- **메모리 서브시스템 dead-config 제거** — `SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001` 소관.
- **스킬 본문 결정 휴리스틱** — `SPEC-V3R6-SKILL-DECISION-HEURISTICS-001` 소관 (스킬 크래프트 장치). 본 SPEC은 orchestrator 수준 의사결정 메모리이다.

### Out of Scope — 구현 세부 (run-phase로 이월)

- **Go 코드 함수명/클래스 구조/API 스키마** — `internal/hook/` 확장, `internal/cli/hook.go runHarnessObserve` 확장, `internal/cli/preference/` 신규 패키지의 구체적 시그니처는 plan.md §F 마일스톤 범위 명세까지만; 실제 구현은 run-phase manager-develop 위임
- **askuser-protocol.md 템플릿 카피의 정확한 diff 헝크** — plan.md가 수정 영역을 명시하되, 바이트 단위 diff는 run-phase 산출물
- **PostToolUse stdin JSON 파싱 구현 상세** — `internal/hook/post_tool.go`의 토큰 추출 로직은 run-phase

### Out of Scope — 산업 도구 직접 통합

- **Cursor rules / Windsurf memories / GitHub Copilot Memory / Aider config의 외부 프로토콜** — 본 SPEC은 MoAI 내부 메모리 모델만 정의; 외부 도구와의 가져오기/내보내기는 별도 SPEC

---

## §F. 위험 및 완화

| 위험 | 완화 |
|------|------|
| 캡처 훅이 race condition으로 선호 메모리 손상 | advisory/fail-open + 원자적 upsert (REQ-ADM-009); 다중 세션 race는 `agent-common-protocol.md §Pre-Spawn Sync Check` 준수 |
| 추천이 filter bubble 유발 | REQ-ADM-014 민감 작업 게이트 + REQ-ADM-013 회복 토글 + 의도적 우연성 주입 (보조적 근거 unverified: Iyendo serendipity; `[verified]` Loughrey autonomy가 핵심) |
| 추론된 선호가 잘못되어 사용자 불편 | REQ-ADM-016 정정 루프 + REQ-ADM-018 verification-claim-integrity (관측되지 않은 추론 주장 금지) |
| advisory 훅이 exit 2로 워크플로 차단 | REQ-ADM-009 fail-open 강제 + 기존 exit-2 훅(sync-phase-quality-gate.sh, status-transition-ownership.sh)만 차단 권한 보유 명시 |
| 템플릿 중립성 위반 (내부 SPEC ID가 template-shipped 카피로 누출) | design.md §D 템플릿 중립성 분할 매트릭스 + CI guard `internal/template/internal_content_leak_test.go` |
| 28일 TTL이 핵심 stable 선호를 만료시킴 | REQ-ADM-011 일시/지속 분리 — stable은 pure time-decay 면제, last_used 갱신 |
| 회복 턴에서 캡처가 death-spiral 유발 | REQ-ADM-010 Recovery-Signal Carve-Out (SHOULD, doctrine-honest; runtime-recovery-doctrine.md §4 + AP-RR-006 준수 — 탐지 메커니즘은 future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001로 이연) |

---

## §G. 관련 SPEC 교차 참조

- `SPEC-ASKUSER-ENFORCE-001` (implemented, legacy) — AskUserQuestion 채널 독점 + Socratic 인터뷰 절차. 본 SPEC은 그 위에 의사결정 지속성 계층을 추가한다.
- `SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001` (completed) — 메모리 서브시스템 dead-config 제거. 본 SPEC의 `user_decisions/` 계층은 이 정리 이후의 깨끗한 상태에서 추가된다.
- `SPEC-V3R6-SKILL-DECISION-HEURISTICS-001` (completed) — 스킬 본문 결정 휴리스틱 장치. 본 SPEC은 orchestrator 수준이며, 스킬 수행 시 발생하는 사용자 의사결정을 캡처한다.
- `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` (completed) — runtime-recovery-doctrine. 본 SPEC REQ-ADM-010이 §4 Recovery-Signal Carve-Out을 준수한다 (SHOULD, doctrine-honest — 탐지 메커니즘은 future SPEC으로 이연).
- `SPEC-EVIDENCE-CLAIM-INVARIANT-001` (completed) — verification-claim-integrity doctrine. 본 SPEC REQ-ADM-018이 §1.1 surface 3을 준수한다.

---

## §H. 성공 기준 (요약; 상세는 acceptance.md)

1. **advisory/fail-open 준수** — 캡처 훅이 어떤 오류 상황에서도 exit 0 (AC-ADM-009)
2. **Recovery-Signal Carve-Out 준수 (SHOULD, doctrine-honest)** — 회복 턴에서 캡처 미실행 + exit 0 행동 정의; 탐지 메커니즘은 future SPEC으로 이연 (AC-ADM-010, S3 Major)
3. **verification-claim-integrity 준수** — 모든 추천이 관측된 메모리 엔트리에 매핑 (AC-ADM-018)
4. **stable/transient 분리 검증** — stable 엔트리가 28일 TTL에 만료되지 않음 (AC-ADM-011)
5. **템플릿 중립성 경계 준수** — template-shipped 카피에 내부 SPEC ID/REQ 토큰 누출 0 (AC-ADM-NFR-006)
6. **GEARS 준수** — 모든 요구사항이 GEARS 패턴 사용, 레거시 IF/THEN 0 (plan-auditor gate)
