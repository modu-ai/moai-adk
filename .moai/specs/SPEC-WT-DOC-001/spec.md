---
id: SPEC-WT-DOC-001
status: draft
version: "0.1.0"
priority: Low
labels: [worktree, shared-state, documentation, anti-pattern, termination, wave-3, tier-2]
issue_number: null
scope: [.claude/rules/moai/workflow/worktree-integration.md, CLAUDE.md]
blockedBy: []
dependents: []
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 3
tier: 2
---

# SPEC-WT-DOC-001: Worktree Shared State 명시

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 3 / Tier 2. Anthropic Multi-Agent Coordination Patterns의 Shared State 권고를 본 프로젝트 worktree-integration.md 정책 강화로 흡수. 코드 변경 없는 documentation-only SPEC.

---

## 1. Goal (목적)

본 프로젝트의 git worktree isolation 메커니즘 위에서 발생할 수 있는 shared state (e.g., `.moai/specs/<ID>/progress.md`) 동시 쓰기 / cross-worktree merge 시나리오에 대한 명시적 정책을 `.claude/rules/moai/workflow/worktree-integration.md`에 추가한다. SPEC-LOOP-TERM-001 (Wave 2)의 termination schema를 의미적으로 확장 인용하고, 5개 anti-pattern을 카탈로그화한다.

### 1.1 배경

- Anthropic blog "Multi-Agent Coordination Patterns": "Shared state without termination becomes a synchronization graveyard. Define when state transitions terminate, who is the writer of last resort, and what the consistency model is."
- 본 프로젝트의 `.moai/specs/<ID>/progress.md` 등이 shared state로 사용되나 동시 쓰기 정책 부재
- v2.14.0 case study (CLAUDE.local.md §18.11): squash merge로 cross-worktree state 손실 사례 — 정책 부재가 원인
- claude-code-guide ⚠️: Worktree isolation으로 충분, 문서화만 필요 → 코드 변경 없는 SPEC

### 1.2 비목표 (Non-Goals)

- Go 코드로 cross-worktree write 차단 (over-engineering)
- Hook 기반 enforcement (별도 SPEC 후보)
- MoAI Constitution 격상 (worktree rule이 흡수)
- 새로운 rule 파일 신설 (worktree-integration.md에 흡수)
- 동시성 자동 conflict resolution
- 강제 schema 적용 (권장 수준)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/rules/moai/workflow/worktree-integration.md`에 다음 절 추가:
  - "Shared State Policy" 절 (per-file ownership, writer-of-last-resort)
  - "Concurrency Model" 절 (consistency: eventual via PR merge)
  - "Anti-Patterns" 절 (5개 anti-pattern 카탈로그)
  - "Termination Conditions" 절 (SPEC-LOOP-TERM-001 의미 확장 인용)
- `CLAUDE.md` §14에 cross-ref 추가
- Template-First 동기화 (`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`)

### 2.2 Exclusions (What NOT to Build)

- Go 코드로 cross-worktree write 차단
- Hook 기반 anti-pattern enforcement
- 새 rule 파일 신설
- MoAI Constitution 변경
- 자동 conflict resolution
- 강제 schema 적용
- worktree 자동 생성/삭제 도구

---

## 3. Environment (환경)

- 영향 파일: `.claude/rules/moai/workflow/worktree-integration.md`, `CLAUDE.md`, `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`
- 의존: SPEC-LOOP-TERM-001 (Wave 2 산출물 — termination schema 인용 source)
- 변경 형식: 문서 only (코드 변경 0)

---

## 4. Assumptions (가정)

- A1: worktree-integration.md가 본 프로젝트의 표준 reference
- A2: SPEC-LOOP-TERM-001의 termination schema가 worktree shared state context에 의미 확장 가능
- A3: 5개 anti-pattern은 본 프로젝트의 실제 incident에서 도출 (CLAUDE.local.md §18.11)
- A4: 코드 변경 없는 문서 SPEC은 빠른 머지 가능
- A5: cross-worktree state는 PR merge 외 방법 없음 (Git semantics)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-WD-001**: THE FILE `.claude/rules/moai/workflow/worktree-integration.md` SHALL include a "Shared State Policy" section defining per-file ownership and writer-of-last-resort rules.
- **REQ-WD-002**: THE FILE SHALL include a "Concurrency Model" section explicitly stating: "consistency is eventual, achieved via PR merge to main; no direct cross-worktree file writes".
- **REQ-WD-003**: THE FILE SHALL include an "Anti-Patterns" section listing at least 5 anti-patterns with concrete examples.
- **REQ-WD-004**: THE FILE SHALL include a "Termination Conditions" section that cross-references SPEC-LOOP-TERM-001 schema for shared state loop termination.

### 5.2 Event-Driven Requirements

- **REQ-WD-005**: WHEN multiple agents in different worktrees access `.moai/specs/<ID>/progress.md` for the same SPEC, THE ACCESS PATTERN SHALL follow worktree isolation rules — each worktree maintains its own copy until PR merge.
- **REQ-WD-006**: WHEN cross-worktree state synchronization is needed, THE PATTERN SHALL be PR merge to main, not direct file access from one worktree to another.
- **REQ-WD-007**: WHEN worktrees diverge during feature development, THE MERGE STRATEGY at PR time SHALL preserve all SPEC creations (no overwriting of SPEC files).

### 5.3 State-Driven Requirements

- **REQ-WD-008**: WHILE multiple worktrees exist for parallel SPEC development, THE PROJECT SHALL document each worktree's SPEC ownership in `.moai/.moai-worktree-registry.json` (existing convention).

### 5.4 Conditional (WHERE / IF) Requirements

- **REQ-WD-009**: WHERE a SPEC requires shared state with reactive loops (e.g., evaluator feedback loop), THE TERMINATION SCHEMA SHALL reference SPEC-LOOP-TERM-001 fields (max_iterations, improvement_threshold, escalation_after).
- **REQ-WD-010**: IF an anti-pattern is detected during code review, THEN THE PR SHALL be requested to refactor according to the cookbook before merge.
- **REQ-WD-011**: WHERE `CLAUDE.md §14` references parallel execution safeguards, THE CROSS-REFERENCE to worktree-integration.md "Shared State Policy" SHALL be present.

### 5.5 Unwanted (Negative) Requirements

- **REQ-WD-012**: THE POLICY SHALL NOT mandate Go code or hook-based enforcement (text guide only).
- **REQ-WD-013**: THE POLICY SHALL NOT prohibit shared state usage (only undefined / undocumented usage).
- **REQ-WD-014**: THE POLICY SHALL NOT introduce new rule files beyond worktree-integration.md (single source of truth).
- **REQ-WD-015**: THE POLICY SHALL NOT mandate automatic conflict resolution algorithms.

---

## 6. Success Criteria (성공 기준)

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Anti-pattern count | document section count | >= 5 |
| Termination cross-ref | SPEC-LOOP-TERM-001 인용 | EXISTS |
| Shared State Policy 절 | section exists | EXISTS |
| Concurrency Model 절 | section exists | EXISTS |
| CLAUDE.md cross-ref | check | EXISTS |
| Template-First sync | `make build` diff | clean |
| Code change | git diff (Go files) | 0 lines |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: 코드 변경 0 (documentation-only SPEC)
- C2: 새 rule 파일 신설 금지 (worktree-integration.md에 흡수)
- C3: SPEC-LOOP-TERM-001 인용은 의미적 확장만 (요건 추가 금지)
- C4: anti-pattern은 실제 incident 기반 (가공 금지)
- C5: Template-First Rule 준수

End of spec.md (SPEC-WT-DOC-001 v0.1.0).
