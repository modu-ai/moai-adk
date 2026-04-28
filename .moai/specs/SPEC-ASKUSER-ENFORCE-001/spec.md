---
id: SPEC-ASKUSER-ENFORCE-001
version: "1.0.0"
status: draft
created: "2026-04-25"
updated: "2026-04-25"
author: GOOS
priority: high
issue_number: 0
depends_on: []
---

# SPEC-ASKUSER-ENFORCE-001: AskUserQuestion 의무화 + Socratic Interview 표준화 + ToolSearch 사전 로드 절차 정식화

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0.0 | 2026-04-25 | GOOS | Initial draft. AskUserQuestion-only 사용자 상호작용 채널 의무화 + Socratic interview 절차 표준화 + deferred tool ToolSearch 사전 로드 절차 정식화 + 5 핵심 지침 파일 일괄 정합성 보강 SPEC. |

---

## 1. 개요

본 SPEC은 MoAI orchestrator가 사용자와 상호작용하는 **유일한 합법 채널**이 `AskUserQuestion`임을 모든 핵심 지침 문서에 일관되게 명시하고, 사용자의 의도가 모호하거나 추가 맥락이 필요할 때 항상 **Socratic interview** 형식의 라운드제 질문을 사용하도록 절차를 표준화한다. 또한 `AskUserQuestion`이 Claude Code 환경에서 **deferred tool**(스키마가 기본 컨텍스트에 미적재)이라는 사실을 정식 지침으로 편입하여, 호출 직전에 `ToolSearch(query: "select:AskUserQuestion")`로 스키마를 사전 로드하는 단계를 누락 없이 수행하도록 강제한다.

본 SPEC은 새로운 코드를 만들지 않는다. 기존에 분산되어 있던 규범을 한 곳으로 모으고(canonical reference), 5개 핵심 지침 파일(CLAUDE.md, moai-constitution.md, agent-common-protocol.md, output-styles/moai/moai.md + 신규 askuser-protocol.md)을 정합성 있게 갱신한 뒤, Template-First 원칙에 따라 `internal/template/templates/*` 미러를 동기화하고 `make build`로 임베디드 산출물을 갱신한다.

## 2. 배경 및 문제

### 2.1 직접 원인 (Direct Cause)

`AskUserQuestion`은 Claude Code v2.x에서 **deferred tool**로 분류된다. 즉, agent 컨텍스트 초기화 시점에는 도구 스키마가 로드되지 않고, 호출 직전에 `ToolSearch(query: "select:AskUserQuestion")`를 통해 명시적으로 select해야 한다. 이 절차를 누락한 채 `AskUserQuestion`을 호출하면 `InputValidationError: tool not in schema` 형태의 런타임 오류가 발생한다.

문제는 현재 시점에서 본 절차가 `CLAUDE.md`, `.claude/rules/moai/core/moai-constitution.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/output-styles/moai/moai.md` 어디에도 명시되어 있지 않다는 점이다. orchestrator가 사용자에게 질문을 던지려는 순간, 어떤 사전 절차가 필요한지에 대한 단일 출처가 부재하다.

### 2.2 비대칭 프로토콜 (Asymmetric Protocol)

`agent-common-protocol.md`의 `User Interaction Boundary` 단락은 **subagent가 `AskUserQuestion`을 호출하는 것을 금지**하는 규범만 단방향으로 기술하고 있다. 그러나 그 반대편, 즉 **orchestrator만 호출 가능**하다는 사실과, 호출 시 **반드시 ToolSearch 사전 로드를 선행**해야 한다는 의무는 codify되어 있지 않다. 결과적으로 "subagent 금지"는 규범으로 존재하지만 "orchestrator의 사전 로드 의무"는 규범으로 존재하지 않는다.

### 2.3 Output Style의 절차 누락 (Output Style Gap)

`.claude/output-styles/moai/moai.md` §3 Stage 1 "Clarify"는 "AskUserQuestion으로 질문하라"고 지시하지만, 그 직전 단계인 "deferred tool ToolSearch select"를 명시하지 않는다. Output Style은 orchestrator의 turn-by-turn 동작 흐름을 정의하는 가장 가까운 절차서이므로, 이 누락은 운영상 가장 빈번한 오류 진입점이 된다.

### 2.4 Stage 1 트리거 매트릭스의 누락 (Trigger Matrix Gap)

`CLAUDE.md` §7 Rule 5 "Context-First Discovery"는 4종 ambiguity trigger를 정의한다 — pronoun without referent, multi-interpretable verb, unclear boundary, conflict with existing state. 그리고 `CLAUDE.md` §8 "Ambiguity Triggers"가 동일 4종을 약간 다른 표현으로 다시 정의한다. 두 섹션이 모두 trigger를 명시하지만, **trigger 충족 직후의 첫 행동**이 무엇인지(=ToolSearch deferred tool 사전 로드 → AskUserQuestion 구성)에 대한 bridging step이 어느 쪽에도 없다.

### 2.5 Memory Dead Lesson (Memory Graduation Failure)

사용자의 자가 메모리 저장소(`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md`)에는 이미 본 SPEC과 동일한 사실관계가 "AskUserQuestion Enforcement — Deferred tool preload 의무 + 산문 질문 anti-pattern + 재발 방지 프로토콜"이라는 제목으로 기록되어 있다. 그러나 이 lesson은 canonical 지침 트리(`.claude/rules/moai/`)로 graduate되지 못한 채 dead lesson 상태로 남아 있다. `moai-constitution.md`의 Lessons Protocol에 따르면, 5회 이상 관찰된 패턴은 graduation 대상이 되어야 한다. 본 SPEC은 이 dead lesson의 정식 graduation을 포함한다.

### 2.6 종합 (Synthesis)

위 5축의 원인은 모두 동일한 표면 증상으로 수렴한다 — orchestrator가 사용자에게 질문해야 할 순간에 `AskUserQuestion` 대신 응답 본문에 산문 의문문을 넣어버리거나, `AskUserQuestion`을 시도하다가 deferred tool 스키마 미적재로 실패하는 현상이다. 본 SPEC은 단일 PR로 5축 원인 모두를 동시에 봉합한다.

## 3. 요구사항 (EARS)

### REQ-AUE-001 (Ubiquitous)

The system (MoAI orchestrator) **shall** present every user-facing question exclusively through an `AskUserQuestion` tool invocation, and free-form interrogative prose in the response body **shall not** be used as a question channel.

- 적용 범위: Stage 1 Clarify, 의사 결정 분기, Socratic interview 라운드, 분쟁 해소(merge strategy, rollback confirmation 등)
- 예외: AskUserQuestion이 기술적으로 불가능한 경우(현실에서는 거의 발생하지 않음), 또는 표현 형식이 의문문이지만 실제로는 상태 보고인 문장
- 검증: `.claude/rules/moai/core/agent-common-protocol.md` `User Interaction Boundary` 단락의 양방향 명문화 (orchestrator 의무 + subagent 금지)

### REQ-AUE-002 (Event-driven)

**When** a deferred tool call (예: `AskUserQuestion`)이 예정될 때, the system **shall** immediately precede the call with `ToolSearch(query: "select:<tool>[,<tool>...]")` to load the tool schema into the active context.

- 적용 대상: `AskUserQuestion`을 포함한 모든 deferred tool — 본 요구사항은 향후 추가될 deferred tool에도 일반화 적용되어야 한다.
- 호출 형태: `ToolSearch(query: "select:AskUserQuestion")` (단일) 또는 `ToolSearch(query: "select:AskUserQuestion,OtherDeferredTool")` (다중)
- 검증: `.claude/rules/moai/core/askuser-protocol.md` "ToolSearch Preload Procedure" 섹션 + `.claude/output-styles/moai/moai.md` §3 Stage 1의 "preload step" 신설

### REQ-AUE-003 (State-driven)

**While** a Stage 1 Clarify trigger is satisfied, the system **shall** execute Socratic interview rounds with the following structural constraints:

- 라운드당 최대 4개 질문, 질문당 최대 4개 옵션 (Claude Code AskUserQuestion 한계와 일치)
- 첫 옵션은 추천 옵션이며 라벨에 `(권장)` 또는 `(Recommended)` 접미를 포함한다.
- 모든 질문 텍스트, 옵션 라벨, 옵션 description은 사용자의 `conversation_language`(현재 프로젝트는 `ko`)로 작성되어야 한다.
- 후속 라운드는 직전 라운드의 답변을 기반으로 모호성을 좁혀야 하며, 단순 반복은 금지된다.
- 의도 명료도가 100%에 도달하기 전까지 라운드를 종료해서는 안 된다.

### REQ-AUE-004 (Ubiquitous)

The system **shall** populate every `AskUserQuestion` option's `description` field with sufficient detail for the user to evaluate implications and trade-offs without consulting external context — 라벨만 보고 결정을 내릴 수 없는 옵션은 description 부재로 간주된다.

- 검증 기준: 옵션 description은 (a) 해당 옵션 선택 시의 즉각적 결과, (b) 부수 효과/리스크, (c) 가능한 경우 정량 정보(예: "토큰 30K 절감", "응답 시간 +200ms")를 포함해야 한다.
- bias prevention: description은 옵션을 추천하거나 폄하하는 어조를 사용해서는 안 되며, 추천 신호는 라벨의 `(권장)` 접미로만 전달한다.

### REQ-AUE-005 (Event-driven)

**When** a subagent attempts to invoke `AskUserQuestion` (직접 호출 시도 또는 자유 서술 형태의 사용자 질문 출력), the orchestrator **shall** reject the attempt and convert it into a structured blocker report.

- subagent 측 동작: AskUserQuestion 호출 코드를 prompt body에 포함시키지 않으며, 사용자 질문이 필요한 상황에서는 "missing inputs" 섹션을 포함한 blocker report를 반환한다.
- orchestrator 측 동작: blocker report 수신 시 자체 컨텍스트에서 `AskUserQuestion` 라운드를 실행한 후, 결과를 다음 subagent 호출 prompt에 주입하여 재위임한다.
- 검증: `.claude/rules/moai/core/agent-common-protocol.md` 의 `User Interaction Boundary` 단락이 양방향(orchestrator 의무 + subagent 금지)으로 codify되어야 한다.

### REQ-AUE-006 (Event-driven)

**When** any one of the four ambiguity triggers is satisfied — (a) pronoun or demonstrative without clear referent, (b) multi-interpretable action verb without specified scope, (c) unclear boundaries (how far / how much / which files / where to stop), (d) potential conflict with existing state (uncommitted changes, in-progress branches, overlapping work) — the system **shall** immediately initiate a Stage 1 Clarify round.

- trigger 정의의 단일 출처: `CLAUDE.md` §7 Rule 5와 §8 `Ambiguity Triggers`의 합집합으로 정의되며, 두 섹션은 동일 4종을 동일 어휘로 표현해야 한다(중복 표현 금지, 일관성 강제).
- 첫 행동 시퀀스: trigger 충족 → ToolSearch select(REQ-AUE-002) → AskUserQuestion 라운드 구성(REQ-AUE-003) → 응답 수집 → 의도 명료도 평가 → 라운드 반복 또는 종료
- 면제 조건은 `CLAUDE.md` §7 Rule 5 Exceptions와 §8 Exceptions의 합집합으로 정의된다 (single-line typo, 명시적 reproduction이 제공된 bug fix, 경로가 지정된 파일 read, 모든 인자가 제공된 command 호출, 동일 세션 내 사전 확인된 작업의 연속).

### REQ-AUE-007 (Unwanted Behavior)

**If** the user expresses a strong preference for free-form text answers, **then** the system **shall not** circumvent the AskUserQuestion channel by presenting a free-form interrogative.

- 근거: `AskUserQuestion`은 모든 질문에 자동으로 "Other" 옵션을 제공한다. 자유 서술이 필요한 사용자는 "Other"를 선택한 뒤 자유 서술 답변을 제출할 수 있다.
- 운영 규범: orchestrator는 이 사실을 사용자에게 안내할 의무가 없으며, 자유 서술이 필요해 보이는 상황에서도 항상 AskUserQuestion 라운드로 시작한다.

### REQ-AUE-008 (Ubiquitous)

The system **shall** document the deferred-tool ToolSearch procedure and AskUserQuestion-only enforcement in **all** of the following five canonical guidance files, with consistent terminology and cross-references:

1. `CLAUDE.md` — §7 Rule 5 "Context-First Discovery" 및 §8 "User Interaction Architecture" 섹션
2. `.claude/rules/moai/core/moai-constitution.md` — `MoAI Orchestrator` 단락
3. `.claude/rules/moai/core/agent-common-protocol.md` — `User Interaction Boundary` 단락
4. `.claude/output-styles/moai/moai.md` — §3 Stage 1 "Clarify" 및 §10 "Output Rules" 섹션
5. `.claude/rules/moai/core/askuser-protocol.md` — **신규 생성**, canonical reference 역할

다섯 파일은 동일 사실에 대해 동일 어휘를 사용해야 한다 — "AskUserQuestion is the **only** user-facing question channel", "AskUserQuestion is a **deferred tool**, requires `ToolSearch(query: \"select:AskUserQuestion\")` preload", "Stage 1 Socratic interview: ≤4 questions per round, ≤4 options per question, first option marked `(권장)` / `(Recommended)`".

### REQ-AUE-009 (Event-driven)

**When** any of the five canonical guidance files listed in REQ-AUE-008 is modified, the system **shall** synchronize the corresponding mirror under `internal/template/templates/*` in the same PR and verify that `make build` regenerates `internal/template/embedded.go` without diff for non-target files.

- Template-First 원칙: `CLAUDE.local.md` §2 "Template-First Rule" 준수
- 동기화 대상 미러:
    - `internal/template/templates/CLAUDE.md`
    - `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`
    - `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
    - `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` (신규)
    - `internal/template/templates/.claude/output-styles/moai/moai.md`
- 빌드 검증: `make build` 실행 후 `git diff internal/template/embedded.go`가 위 5개 파일 변경분만 반영함을 확인

## 4. 범위

### 4.1 In-Scope

- 5 핵심 지침 파일 본문 갱신 (CLAUDE.md, moai-constitution.md, agent-common-protocol.md, output-styles/moai/moai.md + 신규 askuser-protocol.md)
- `internal/template/templates/*` 5 미러 파일 동기화
- `internal/template/embedded.go` 재생성 (make build)
- memory dead lesson `feedback_askuserquestion_enforcement.md`의 정식 graduation — 본 SPEC ID로 SUPERSEDED 처리
- 정합성 검증 스크립트 (각 파일에 필수 마커 문자열 존재 여부를 grep으로 검증) — 별도 신규 코드 추가 없이 acceptance.md의 Evidence로 통합

### 4.2 Out-of-Scope

- `AskUserQuestion` 도구 자체의 구현 변경 — Claude Code 런타임 영역
- 신규 CLI 플래그 추가 — 본 SPEC은 지침 정합성 보강만 다룬다
- 신규 agent definition 또는 신규 skill 생성 — 본 SPEC은 기존 지침 재배열에 한정한다
- SPEC-WF-AUDIT-GATE-001 (plan→run audit gate) 와의 통합 작업 — 두 SPEC은 독립 실행되며, 두 SPEC이 모두 머지된 이후 본 SPEC이 audit gate를 통과하는지 self-dogfood 검증을 수행한다
- 기존 `AskUserQuestion` 사용 패턴의 강제 변경 — 본 SPEC은 의무를 추가하지만 기존 정상 사용 패턴은 그대로 유효하다 (백워드 호환)

## 5. 위험 및 완화

| Risk | 시나리오 | 완화 |
|------|----------|------|
| R-AUE-1 | 동시 다중 파일 수정 중 일관성 깨짐 (예: CLAUDE.md만 갱신, output-styles 누락) | 단일 PR + acceptance.md AC-AUE-008-x의 grep 기반 자동 일관성 체크 |
| R-AUE-2 | ToolSearch 절차 추가가 다른 deferred tool에도 회귀 영향 | REQ-AUE-002를 deferred tool 일반에 적용 — `AskUserQuestion-only` 한정이 아닌 일반화 표현 사용 |
| R-AUE-3 | 본 SPEC 자체가 SPEC-WF-AUDIT-GATE-001 audit을 통과하지 못함 | 정성 작성 + self-audit (plan-auditor 8 criteria 자가 검증) + acceptance.md의 모든 AC가 Evidence를 명시 |
| R-AUE-4 | Template 미러 동기화 누락 | acceptance.md AC-AUE-009-x의 빌드 검증으로 강제 — `make build` 후 `embedded.go` diff 확인 |
| R-AUE-5 | Memory dead lesson SUPERSEDED 처리 누락 | tasks.md Phase G의 명시적 task로 분리, DoD에 grep evidence 포함 |
| R-AUE-6 | conversation_language 변경 시 Socratic interview 라운드의 라벨 동기화 누락 | REQ-AUE-003의 `conversation_language` 동적 참조 명시 — 하드코딩 금지 |

## 6. HISTORY

- 2026-04-25: SPEC-ASKUSER-ENFORCE-001 v1.0.0 created. AskUserQuestion-only 사용자 상호작용 채널 의무화 + Socratic interview 절차 표준화 + deferred tool ToolSearch 사전 로드 절차 정식화. 5축 root cause 분석 (direct cause / asymmetric protocol / output style gap / trigger matrix gap / memory dead lesson) 기반.
