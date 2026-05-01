---
id: SPEC-CONTEXT-INJ-001
status: draft
version: "0.1.1"
priority: Medium
labels: [context-injection, memory, sub-agent, orchestrator, progress, wave-3, tier-2]
issue_number: null
scope: [.claude/rules/moai/development, .claude/skills/moai-foundation-core, CLAUDE.md]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 3
tier: 2
---

# SPEC-CONTEXT-INJ-001: Memory Persistence 단순화 / Context Injection

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 3 / Tier 2. Anthropic Managed Agents Memory 권고를 본 프로젝트 sub-agent context 모델에 맞춰 명시 주입(orchestrator-injected) 정책으로 변환.
- 2026-04-30 v0.1.1: BLOCKING 결함 수정 — 토큰 예산 단위를 "5KB"에서 "5000 tokens"으로 통일 (1 token ≈ 4 chars 변환 공식 명시, 측정 도구 명시).

---

## 1. Goal (목적)

Claude Code sub-agent가 매번 fresh context로 spawn되는 본 프로젝트 환경에서, "auto memory mount" 환상을 포기하고 orchestrator가 sub-agent prompt 작성 시 SPEC progress, related memory, domain lessons를 명시 주입하는 표준 정책을 신설한다. **5000 tokens** 토큰 cap + 우선순위 정책으로 context overflow를 방지한다.

> **단위 표기 (Single Source of Truth)**: 본 SPEC은 토큰 예산을 **5000 tokens**으로 통일한다. 변환 공식: 1 token ≈ 4 characters (Anthropic Tokenizer 기준, English/Korean/Mixed 평균치). 측정 도구: `tiktoken` (cl100k_base, Claude는 동일 BPE family) 또는 `wc -m` × 0.25 근사. 과거 표기 "5KB"는 본 SPEC에서 사용하지 않으며, 모든 후속 문서는 tokens 단위로 통일한다.

### 1.1 배경

- Anthropic blog "Claude Managed Agents Memory": "Memory on Managed Agents mounts directly onto a filesystem... Stores can be shared across multiple agents with different access scopes."
- Anthropic blog "Harnessing Claude's Intelligence": "Memory on Managed Agents is automatic; on Claude Code sub-agents, it is the orchestrator's responsibility to inject relevant context at spawn time."
- 본 프로젝트의 sub-agent는 매번 fresh context → 표준화된 명시 주입 정책 부재
- progress.md, recent feedback, domain lessons 주입이 ad-hoc 진행됨

### 1.2 비목표 (Non-Goals)

- Managed Agents Memory mount 자동화 (Claude Code sub-agent 미지원)
- Go 코드 기반 자동 주입 helper 신규 구현 (정책 표준화가 우선)
- progress.md schema의 hard rule 격상 (권장 수준 유지)
- 토큰 cap을 넘는 context 주입 자동 truncation 도구 (orchestrator 책임)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/rules/moai/development/context-injection.md` 신규 작성
- 5000 tokens cap per agent invocation 명시 (단위: tokens — see §1)
- 우선순위 정책: `progress.md` > `recent feedback (MEMORY.md 발췌)` > `domain lessons`
- progress.md 권장 schema (작성 가이드)
- 주입 마커 (`<!-- injected-context -->` ... `<!-- /injected-context -->`)
- manager-* / expert-* agent body에 cross-ref 추가
- `moai-foundation-core` SKILL.md Token Budget 절 보강
- CLAUDE.md cross-ref 추가
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- Go 자동 주입 helper 구현
- progress.md schema 강제
- 5000-token cap 자동 측정/차단 도구
- truncation 자동화
- LSP 검증 (텍스트 정책 한정)
- cross-session context restore (별도 SPEC)

---

## 3. Environment (환경)

- 런타임: orchestrator (MoAI main session)
- Claude Code v2.1.111+, Opus 4.7
- 영향 파일: `.claude/rules/moai/development/`, `.claude/skills/moai-foundation-core/`, `.claude/agents/moai/manager-*.md`, `CLAUDE.md`
- 의존 입력: `.moai/specs/<ID>/progress.md` (선택), `~/.claude/projects/<hash>/memory/MEMORY.md` (선택)

---

## 4. Assumptions (가정)

- A1: `Agent()` spawn prompt는 텍스트 형태이며 orchestrator가 자유 구성 가능
- A2: 5000 tokens cap은 sub-agent 토큰 예산 200K (Sonnet/Opus standard) 기준 ~2.5%, 1M (Opus 4.7) 기준 ~0.5%로 합리적 fit
- A3: 우선순위 정책은 일반 SPEC 90% 이상에 적합 (예외는 수동 조정)
- A4: progress.md 부재 SPEC에서는 silent skip
- A5: 정책 위반은 review로 감지 (자동 차단 부재)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-CI-001**: THE FILE `.claude/rules/moai/development/context-injection.md` SHALL exist and define the orchestrator's injection policy for sub-agent spawn prompts.
- **REQ-CI-002**: THE INJECTION POLICY SHALL specify a token budget cap of 5000 tokens per sub-agent invocation.
- **REQ-CI-003**: THE INJECTION POLICY SHALL define a 3-tier priority order: SPEC progress.md > recent feedback (MEMORY.md excerpts) > domain lessons.
- **REQ-CI-004**: THE INJECTION POLICY SHALL specify the marker convention `<!-- injected-context -->` ... `<!-- /injected-context -->` for separating injected text from task description.

### 5.2 Event-Driven Requirements

- **REQ-CI-005**: WHEN the orchestrator delegates to a manager-* or expert-* sub-agent for a SPEC-bound task, THE ORCHESTRATOR SHALL include relevant context per the injection policy.
- **REQ-CI-006**: WHEN `.moai/specs/<SPEC-ID>/progress.md` exists for the active SPEC, THE ORCHESTRATOR SHALL inject its content into the sub-agent prompt subject to the 5000-token cap.
- **REQ-CI-007**: WHEN injected context approaches the 5000-token cap, THE ORCHESTRATOR SHALL apply truncation per the priority order, retaining the highest priority entries first.

### 5.3 State-Driven Requirements

- **REQ-CI-008**: WHILE the active session has unrecorded context (in-flight progress), THE ORCHESTRATOR MAY inject in-memory summary derived from the current conversation, marked clearly as "in-flight".

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-CI-009**: WHERE progress.md is absent for the active SPEC, THE ORCHESTRATOR SHALL silently skip progress injection (no warning).
- **REQ-CI-010**: WHERE the sub-agent is research-only (researcher, analyst), THE INJECTION POLICY MAY be relaxed (priority order need not apply strictly).
- **REQ-CI-011**: IF the cumulative injected context exceeds the 5000-token cap after applying priority order, THEN THE ORCHESTRATOR SHALL truncate the lowest-priority entries first and emit a non-blocking note in the prompt.
- **REQ-CI-012**: WHERE the sub-agent's spawn prompt template is templated (e.g., from a workflow file), THE TEMPLATE SHALL include the injection marker placeholder.

### 5.5 Unwanted (Negative) Requirements

- **REQ-CI-013**: THE ORCHESTRATOR SHALL NOT inject raw memory file contents that exceed the 5000-token cap (truncation per REQ-CI-011 mandatory).
- **REQ-CI-014**: THE ORCHESTRATOR SHALL NOT inject agent-private memory of a different agent without that agent's documented consent (cross-scope read prohibition).
- **REQ-CI-015**: THE INJECTION POLICY SHALL NOT mandate progress.md schema (recommendation only).
- **REQ-CI-016**: THE INJECTION POLICY SHALL NOT inject API keys, secrets, or credentials regardless of source priority.

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 정책 문서 존재 | file existence | EXISTS |
| 우선순위 명시 | 3-tier order in document | EXISTS |
| 5000-token cap 명시 | token budget statement (tokens unit) | EXISTS |
| Cross-ref (CLAUDE.md, agents) | cross-ref count | >= 5 |
| 진행 중 SPEC 적용 | sample 3 SPEC sub-agent 호출 | progress 주입 100% |
| 5000-token cap 준수 | sample 5 호출 측정 (tiktoken cl100k_base) | <= 5000 tokens / 100% |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Go 코드 변경 없음 (정책 표준화만, 자동화는 후속 SPEC)
- C2: progress.md schema는 권장 수준 (강제 X)
- C3: 5000 tokens cap은 default; 사용자 설정 가능 여지 명시 (`.moai/config/sections/observability.yaml.context_injection.cap_tokens`, integer, unit: tokens)
- C4: 정책 위반은 review 단계에서 감지 (LSP 검증 부재)
- C5: Template-First Rule 준수

End of spec.md (SPEC-CONTEXT-INJ-001 v0.1.0).
