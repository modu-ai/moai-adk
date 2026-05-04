---
id: SPEC-RESUME-MSG-001
status: draft
version: "0.1.0"
priority: Medium
labels: [resume-message, context-window, session, autocompact, workflow, wave-4, tier-3]
issue_number: null
scope: [.claude/rules/moai/workflow, .claude/skills/moai/workflows, .moai/state]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 4
tier: 3
---

# SPEC-RESUME-MSG-001: Resume Message 강화 / 확산

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 4 / Tier 3. Anthropic "Using Claude Code: Session Management and 1M Context" 권고를 적용하여 본 프로젝트의 모범 사례 (context-window-management.md)를 5+ long-running workflow로 확산.

---

## 1. Goal (목적)

본 프로젝트는 `/moai plan` → `/moai run` 전환에서 75% / 90% threshold + 구조화된 resume message 모범 사례를 보유한다 (`context-window-management.md`). 그러나 다른 long-running workflow (loop / fix / design / sync / Wave-style)는 동일한 보호를 받지 못한다. 본 SPEC은 모범 사례를 6+ workflow로 **확산**하고, `.moai/state/<session-id>.json` 영속화 schema를 정의한다.

### 1.1 배경

- Anthropic blog "Using Claude Code: Session Management and 1M Context" (https://claude.com/blog/using-claude-code-session-management-and-1m-context), § "What causes a bad autocompact?": "bad compacts can happen when the model can't predict the direction your work is going"
- Anthropic blog 동일 출처, § "When to start a new session": "when you start a new task, you should also start a new session"
- Anthropic blog 동일 출처, § "Compacting vs. launching a fresh session": three tools (`/compact` lossy summarization, `/clear` manual clean slate, `/rewind` double-Esc backtrack) — paraphrase; verbatim per-tool descriptions are spread across the section (see research.md §1.1)
- 본 프로젝트의 `context-window-management.md`은 plan/run 전환만 다룸 → loop / design / sync 미적용

### 1.2 비목표 (Non-Goals)

- 75% auto-detect Go 자동화 구현 (orchestrator self-monitoring 정책만)
- `.moai/state/<session-id>.json` schema의 hard rule 격상 (권장 수준)
- AskUserQuestion으로 /clear 트리거 (자연어 status announcement)
- 사용자 자동 paste 도구
- 다른 LLM (Codex, GLM)으로의 resume message 호환성

---

## 2. Scope (범위)

### 2.1 In Scope

- `context-window-management.md` 보강: 6+ workflow 적용 명시
- 표준 resume message format 보존 + 5종 확장 (workflow별)
- `.moai/state/` 디렉토리 정의 + per-session JSON schema (권장)
- `.claude/skills/moai/workflows/{plan,run,sync,loop,fix}.md` cross-ref 추가
- `.claude/skills/moai-team-design/SKILL.md` cross-ref 추가 (GAN loop)
- 75% / 90% threshold 보존
- 자동 저장 trigger는 orchestrator self-monitoring (Go 변경 없음)
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- 75% auto-detect Go 자동화 (orchestrator self-monitoring)
- `.moai/state/<session-id>.json` schema hard rule
- AskUserQuestion 기반 /clear 트리거
- 자동 paste 도구
- 다른 LLM 호환 resume message
- `.moai/state/` 자동 retention cleanup (사용자 책임)

---

## 3. Environment (환경)

- 런타임: orchestrator (MoAI main session)
- 영향 파일: `.claude/rules/moai/workflow/context-window-management.md`, `.claude/skills/moai/workflows/`, `.moai/state/`
- Threshold: 75% (state 저장 권장), 90% (/clear 강력 권고)
- 모델 context: Opus 4.7 (1M, 75% = 750K), Sonnet (200K, 75% = 150K)

---

## 4. Assumptions (가정)

- A1: 본 프로젝트의 기존 resume message 모범 사례가 효과적
- A2: 6 workflow가 long-running 우선순위
- A3: 75%는 Anthropic 1M context 기준 750K tokens
- A4: `.moai/state/<session-id>.json` 권장 schema는 후속 SPEC에서 정형화
- A5: 자동 저장은 orchestrator self-monitoring (Go 코드 변경 없음)
- A6: 사용자가 /clear 후 paste를 수동 수행 (자동화 X)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-RESUME-001**: THE FILE `.claude/rules/moai/workflow/context-window-management.md` SHALL document the standard resume message format and apply it to at least 6 long-running workflows.
- **REQ-RESUME-002**: THE STANDARD RESUME MESSAGE SHALL contain: `ultrathink` prefix, workflow name, SPEC-ID/scope, applied lessons, progress.md or `.moai/state/<session-id>.json` path, next-step command, post-completion target.
- **REQ-RESUME-003**: THE POLICY SHALL define the directory `.moai/state/` for per-session state persistence with recommended JSON schema.
- **REQ-RESUME-004**: THE POLICY SHALL preserve 75% and 90% threshold definitions (75% = state save + resume, 90% = /clear strongly advised).

### 5.2 Event-Driven Requirements

- **REQ-RESUME-005**: WHEN context usage approaches 75% in any covered workflow (plan, run, sync, loop, fix, design, Wave-style), THE WORKFLOW SHALL persist state to `.moai/state/<session-id>.json` and emit a structured resume message.
- **REQ-RESUME-006**: WHEN context usage exceeds 90%, THE WORKFLOW SHALL announce `/clear` recommendation via natural-language status (NOT AskUserQuestion).
- **REQ-RESUME-007**: WHEN user pastes a resume message in a new session after `/clear`, THE WORKFLOW SHALL recognize the `ultrathink` prefix + workflow name and auto-load referenced state.

### 5.3 State-Driven Requirements

- **REQ-RESUME-008**: WHILE a long-running workflow is below the 75% threshold, THE WORKFLOW SHALL NOT proactively save state (cost control).

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-RESUME-009**: WHERE a workflow is one of {plan, run, sync, loop, fix, design, Wave-style}, THE WORKFLOW SHALL include cross-reference to `context-window-management.md` in its skill body.
- **REQ-RESUME-010**: WHERE a Wave-style multi-SPEC delegation is in progress, THE RESUME MESSAGE SHALL include Wave N/M and completed SPEC list.
- **REQ-RESUME-011**: IF the GAN loop (design workflow) reaches iteration N/5 with non-passing score, THE RESUME MESSAGE SHALL include current iteration and last failure dimension.
- **REQ-RESUME-012**: WHERE `.moai/state/<session-id>.json` is missing or unreadable, THE WORKFLOW SHALL emit a degraded resume message (best-effort, with explicit "state unavailable" note).

### 5.5 Unwanted (Negative) Requirements

- **REQ-RESUME-013**: THE WORKFLOW SHALL NOT use AskUserQuestion to trigger `/clear` (status announcement is sufficient; user retains control).
- **REQ-RESUME-014**: THE RESUME MESSAGE SHALL NOT include API keys, secrets, or credentials.
- **REQ-RESUME-015**: THE POLICY SHALL NOT mandate `.moai/state/<session-id>.json` schema as hard rule (recommendation only).
- **REQ-RESUME-016**: THE WORKFLOW SHALL NOT auto-clean `.moai/state/` (retention is user responsibility).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 정책 문서 보강 | 6+ workflow 명시 | 명시 |
| Cross-ref 카운트 | grep `context-window-management.md` | >= 6 |
| Verbatim 인용 | grep Anthropic | >= 3 |
| `.moai/state/.gitkeep` | file existence | EXISTS |
| Format 확장 5종 | 문서 검토 | 모두 명시 |
| 75% / 90% threshold | 문서 검토 | 보존 |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Go 코드 변경 없음 (정책 + cross-ref만)
- C2: 75% auto-detect는 orchestrator self-monitoring (자동화 후속 SPEC)
- C3: `.moai/state/` retention 사용자 책임
- C4: AskUserQuestion 사용 금지 (status announcement 사용)
- C5: Template-First Rule 준수

End of spec.md (SPEC-RESUME-MSG-001 v0.1.0).
