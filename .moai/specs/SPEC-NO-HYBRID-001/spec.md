---
id: SPEC-NO-HYBRID-001
status: draft
version: "0.1.0"
priority: Medium
labels: [srp, anti-hybrid, audit, policy, anti-pattern, wave-4, tier-3]
issue_number: null
scope: [.claude/rules/moai/development, .claude/skills, .moai/reports]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 4
tier: 3
---

# SPEC-NO-HYBRID-001: Anti-Hybrid Tools 원칙 + Audit

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 4 / Tier 3. Anthropic "Seeing Like an Agent" 권고 (Hybrid tool confusion + Format-based control)를 본 프로젝트의 SRP 정책 + 1회성 audit로 변환.

---

## 1. Goal (목적)

본 프로젝트의 일부 워크플로우/도구가 다목적 (multi-mode)으로 운영되어 Anthropic의 "Hybrid tool confusion" 안티패턴 위험을 안고 있다. 본 SPEC은 Single Responsibility Principle (SRP)을 도구 차원에 적용하는 정책 문서를 신설하고, 본 프로젝트 인벤토리에 대한 1회성 audit를 수행한다.

### 1.1 배경

- Anthropic blog "Seeing Like an Agent": "Hybrid tool confusion: Adding parameters to serve multiple purposes simultaneously (asking for both a plan AND questions about the plan) confuses agent behavior."
- Anthropic blog "Seeing Like an Agent": "Format-based control: Attempting to constrain outputs through markdown formatting or structured instructions without proper tool support results in unreliable compliance."
- 본 프로젝트의 `/moai project` 등 일부 명령어는 init/analyze/generate/refresh 4 mode 처리

### 1.2 비목표 (Non-Goals)

- 발견된 다목적 도구의 실제 분리 작업 (별도 SPEC 후속)
- 자동 hybrid tool detection 도구 구현
- SRP 강제 LSP rule
- 팀 간 SRP 적용 일관성 (본 프로젝트만)
- Skill 형식의 정책 (rule 형식이 적합)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/rules/moai/development/single-responsibility.md` 신규 작성
- 5+ Anti-Pattern catalog (Anthropic verbatim 인용 + 본 프로젝트 적용)
- Audit 절차 5-step 정의
- `.moai/reports/srp-audit-2026-04-30.md` 1회성 audit 결과
- "3+ distinct modes" trigger 정의
- Format-based control 식별 가이드
- Cross-ref: `agent-authoring.md`, `skill-authoring.md`, `moai-foundation-cc` SKILL.md
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- 다목적 도구의 실제 분리 작업 (`/moai project` 분리는 후속 SPEC)
- 자동 hybrid tool detection LSP rule
- SRP 강제 enforcement
- ai-agent (외부) SRP 가이드
- 사용자 자기 도구 SRP audit 자동화
- Skill 형식 정책

---

## 3. Environment (환경)

- 런타임: 정책 문서 + 1회성 audit (코드 변경 없음)
- 영향 파일: `.claude/rules/moai/development/`, `.moai/reports/`, `.claude/skills/moai-foundation-cc/`
- Audit 대상: `.claude/skills/*/SKILL.md`, `.claude/agents/moai/*.md`, `internal/cli/*.go`, hook handlers, 슬래시 커맨드

---

## 4. Assumptions (가정)

- A1: 사용자가 SRP 적용을 환영 (혼동 감소)
- A2: 분리 작업은 별도 SPEC (본 SPEC은 정책 + audit만)
- A3: 5+ anti-pattern이 본 프로젝트에 실제로 존재
- A4: Anthropic 권고가 본 프로젝트에 직접 적용 가능
- A5: Audit는 1회성 — 분기별 또는 major release 시 재실행

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-SRP-001**: THE FILE `.claude/rules/moai/development/single-responsibility.md` SHALL exist and define the SRP policy for tools, commands, and workflows.
- **REQ-SRP-002**: THE POLICY SHALL include at least 5 anti-pattern examples with verbatim Anthropic quotes.
- **REQ-SRP-003**: THE POLICY SHALL define a 5-step audit procedure (Inventory → Multi-Mode Detection → Violation Assessment → Remediation Recommendation → Anti-Pattern Catalog).
- **REQ-SRP-004**: THE POLICY SHALL specify the trigger threshold of "3+ distinct execution modes" for SRP review.

### 5.2 Event-Driven Requirements

- **REQ-SRP-005**: WHEN a new tool, command, or workflow is designed, THE AUTHOR SHALL adhere to single responsibility principle as documented.
- **REQ-SRP-006**: WHEN a tool exposes 3+ distinct execution modes, THE AUTHOR SHALL evaluate splitting via SRP audit checklist.
- **REQ-SRP-007**: WHEN format-based output control is detected (e.g., "respond in JSON"), THE AUTHOR SHALL replace the pattern with structured tool support.

### 5.3 State-Driven Requirements

- **REQ-SRP-008**: WHILE the SRP policy is in force, THE PROJECT SHALL conduct a quarterly or per-major-release audit producing `.moai/reports/srp-audit-YYYY-MM-DD.md`.

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-SRP-009**: WHERE a multi-mode tool has documented cohesion (e.g., `/moai db init/refresh/verify/list` with shared state), THE TOOL MAY remain consolidated with explicit subcommand clarity.
- **REQ-SRP-010**: IF audit identifies 3+ violations of equal severity, THE REMEDIATION SHALL be sequenced into separate SPECs to bound scope.
- **REQ-SRP-011**: WHERE a violation involves backward break risk, THE REMEDIATION SHALL be deferred to a major version release.
- **REQ-SRP-012**: IF a remediation requires API change, THE FOLLOW-UP SPEC SHALL include migration guide and deprecation period.

### 5.5 Unwanted (Negative) Requirements

- **REQ-SRP-013**: THE POLICY SHALL NOT mandate splitting every multi-mode tool (cohesion is acceptable).
- **REQ-SRP-014**: THE POLICY SHALL NOT use parameter explosion as the sole trigger (semantic analysis required).
- **REQ-SRP-015**: THE AUDIT SHALL NOT modify any tool, command, or workflow as part of this SPEC (separation is follow-up).
- **REQ-SRP-016**: THE POLICY SHALL NOT enforce SRP via LSP rule (manual review only).

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| 정책 문서 존재 | file existence | EXISTS |
| Anti-pattern catalog | 항목 수 | >= 5 |
| Verbatim 인용 | grep | 2+ Anthropic 출처 |
| 5-step audit 절차 | 문서 검토 | 명시 |
| 1회성 audit report | file existence | `.moai/reports/srp-audit-2026-04-30.md` |
| Audit 항목 수 | report 검토 | >= 10 |
| Cross-ref | grep | >= 3 |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: Go 코드 변경 없음 (정책 + audit report만)
- C2: 분리 작업은 후속 SPEC
- C3: Audit는 1회성 (재실행은 별도 SPEC 또는 사용자 트리거)
- C4: `.moai/reports/` 위치 (CLAUDE.md §SPEC vs Report 권고)
- C5: Template-First Rule 준수

End of spec.md (SPEC-NO-HYBRID-001 v0.1.0).
