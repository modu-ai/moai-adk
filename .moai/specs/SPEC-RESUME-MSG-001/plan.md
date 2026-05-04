---
id: SPEC-RESUME-MSG-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-RESUME-MSG-001

## 1. Overview

본 프로젝트의 모범 사례 (`context-window-management.md` plan/run 전환)를 6+ long-running workflow (loop / fix / design / sync / Wave-style)로 확산. `.moai/state/<session-id>.json` 영속화 schema 정의. 자동화는 orchestrator self-monitoring (Go 변경 없음).

## 2. Approach Summary

**전략**: Spread-Existing-Pattern, Self-Monitoring-Only, Documentation-Driven.

1. `context-window-management.md` 보강: 6+ workflow 적용 명시
2. Format 5 확장 (loop / design / sync / run agent chain / Wave)
3. `.moai/state/<session-id>.json` 권장 schema
4. `.claude/skills/moai/workflows/{plan,run,sync,loop,fix}.md` cross-ref
5. design GAN loop cross-ref
6. Template-First 동기화

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] Anthropic blog "Using Claude Code: Session Management and 1M Context" verbatim 3 인용 재확인
- [ ] 본 프로젝트의 기존 모범 사례 (`context-window-management.md`) 검토
- [ ] 6+ long-running workflow 식별:
  - `/moai plan`
  - `/moai run`
  - `/moai sync`
  - `/moai loop`
  - `/moai fix`
  - `/moai design` (GAN loop)
  - Wave-style multi-SPEC delegation
- [ ] 75% / 90% threshold 본 프로젝트 적용 검증

**Exit Criteria**: 7 workflow 식별 + 모범 사례 baseline

### M1 — Policy Document 보강 (Priority: High)

- [ ] `.claude/rules/moai/workflow/context-window-management.md` 보강:
  - §Existing (75% / 90% threshold) 보존
  - §New: "Applies To" 목록에 6+ workflow 명시
  - §New: "Workflow-Specific Resume Format Extensions" — 5 확장 (loop / design / sync / run agent chain / Wave)
  - §New: ".moai/state/<session-id>.json Recommended Schema"
  - §New: "Auto-Save Trigger (Orchestrator Self-Monitoring)"
- [ ] 보강 후 문서 4-5KB 범위

**Exit Criteria**: 5 신규 절 추가, 기존 보존

### M2 — 5 Format Extension 정의 (Priority: High)

- [ ] **Ext1: /moai loop**
  ```
  ultrathink. /moai loop iteration <N>/<M> 이어서 진행. SPEC-<ID> 통과 기준: <criteria>.
  applied lessons: <files>.
  state: .moai/state/<session-id>.json
  다음 단계: <command>.
  완료 후: <next iteration or sync>.
  ```
- [ ] **Ext2: /moai design (GAN loop)**
  ```
  ultrathink. /moai design GAN iteration <N>/5 이어서 진행. 현재 score: <X.YZ>. 마지막 fail dimension: <name>.
  applied lessons: <files>.
  state: .moai/state/<session-id>.json
  다음 단계: <command>.
  완료 후: <pass or escalate>.
  ```
- [ ] **Ext3: /moai sync (multi-PR)**
  ```
  ultrathink. /moai sync 이어서 진행. pending PR count: <N>. 다음 SPEC: SPEC-<ID>.
  applied lessons: <files>.
  state: .moai/state/<session-id>.json
  다음 단계: <command>.
  완료 후: <next SPEC or finalize>.
  ```
- [ ] **Ext4: /moai run (agent chain)**
  ```
  ultrathink. /moai run agent chain 이어서 진행. completed phase: <N>/<M>. next agent: <name>.
  applied lessons: <files>.
  state: .moai/state/<session-id>.json
  다음 단계: <command>.
  완료 후: <next phase or sync>.
  ```
- [ ] **Ext5: Wave-style multi-SPEC**
  ```
  ultrathink. Wave <N>/4 이어서 진행. completed SPECs: <list>. 다음 SPEC: SPEC-<ID>.
  applied lessons: <files>.
  progress.md 경로: .moai/specs/SPEC-<ID>/progress.md
  다음 단계: <command>.
  완료 후: <next SPEC or finalize wave>.
  ```

**Exit Criteria**: 5 확장 모두 명시, 표준 format 보존

### M3 — `.moai/state/<session-id>.json` Schema (Priority: High)

- [ ] 권장 schema (hard rule 아님):
  ```json
  {
    "session_id": "uuid-or-timestamp",
    "workflow": "plan|run|sync|loop|fix|design|wave",
    "started_at": "ISO-8601",
    "last_saved_at": "ISO-8601",
    "context_usage_percent": 76,
    "active_spec": "SPEC-METRICS-001",
    "applied_lessons": ["feedback_x.md", "lessons.md#5"],
    "current_step": "M3",
    "next_action": "implement workflow extension",
    "extension_data": {
      "iteration": "3/5",
      "score": 0.72,
      "wave": "4/4"
    }
  }
  ```
- [ ] `.moai/state/.gitkeep` 추가
- [ ] retention 사용자 책임 (7일 권장 — manual cleanup)

**Exit Criteria**: schema + placeholder + retention 권장

### M4 — Auto-Save Trigger (Self-Monitoring) (Priority: High)

- [ ] orchestrator self-monitoring 정책:
  - 75% 도달 → state 저장 + resume message 자연어 안내
  - 90% 도달 → /clear 강력 권고
  - **Detection 신호** (orchestrator가 직접 추정):
    - Cumulative output bytes
    - System reminder volume
    - Large tool result count
    - Agent() invocation count
  - **Under-estimate 권장**: premature /clear < missed /clear
- [ ] AskUserQuestion 사용 금지 명시 (status announcement, not question)

**Exit Criteria**: self-monitoring 정책 + AskUserQuestion 금지 명시

### M5 — Cross-Reference + Skill Integration (Priority: High)

- [ ] cross-ref 추가 (6+ skill):
  - `.claude/skills/moai/workflows/plan.md`: cross-ref 추가
  - `.claude/skills/moai/workflows/run.md`: cross-ref 추가
  - `.claude/skills/moai/workflows/sync.md`: cross-ref 추가
  - `.claude/skills/moai/workflows/loop.md`: cross-ref 추가
  - `.claude/skills/moai/workflows/fix.md`: cross-ref 추가
  - `.claude/skills/moai-team-design/SKILL.md`: GAN loop cross-ref
  - (Wave-style은 SPEC 자체 — 별도 문서 X, plan/run에 포함)
- [ ] cross-ref 텍스트 표준:
  ```
  Resume protocol: .claude/rules/moai/workflow/context-window-management.md
  ```

**Exit Criteria**: 6+ cross-ref 추가

### M6 — Template-First Sync + Documentation (Priority: Medium)

- [ ] `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` 동기화
- [ ] `internal/template/templates/.claude/skills/moai/workflows/*.md` 동기화 (M5 cross-ref 반영)
- [ ] `internal/template/templates/.claude/skills/moai-team-design/SKILL.md` 동기화
- [ ] `internal/template/templates/.moai/state/.gitkeep` 신설
- [ ] `make build` 실행 → embedded.go 재생성
- [ ] CHANGELOG entry under Unreleased

**Exit Criteria**: Template-First sync clean

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] 5 format extension 검증
- [ ] cross-ref 6+ 검증 (grep)
- [ ] verbatim 3+ 검증 (Anthropic 인용)
- [ ] `.moai/state/.gitkeep` 검증
- [ ] AskUserQuestion 금지 명시 검증
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 표준 Resume Message 핵심 구조

```
ultrathink. <workflow_name> 이어서 진행. <SPEC-ID/SCOPE>부터 <approach>.
applied lessons: <files>.
progress.md 경로: .moai/specs/SPEC-<ID>/progress.md (또는 .moai/state/<session-id>.json)
다음 단계: <command>.
완료 후: <next>.
```

### 4.2 Workflow 확장 vs 표준 format 관계

- 표준 format = 모든 workflow 공통 7 line
- 확장 = workflow별 1-2 line 추가 (extension_data)
- 사용자 paste 시 prefix `ultrathink. <workflow>`로 자동 인식

### 4.3 75% / 90% Threshold 적용

| Threshold | Action | Channel |
|-----------|--------|---------|
| 75% | state 저장 + resume message 안내 | 자연어 status |
| 90% | /clear 강력 권고 | 자연어 status (NOT AskUserQuestion) |

### 4.4 Detection 신호 (orchestrator self-monitoring)

**Unit conversions (Anthropic standard)**:
- 1 token ≈ 4 characters
- 1 KiB (1024 bytes) ≈ 256 tokens
- 75% of 1M context (Opus 4.7) ≈ 750K tokens ≈ 3 MiB cumulative output
- 75% of 200K context (Sonnet) ≈ 150K tokens ≈ 600 KiB cumulative output

```
# bytes → estimated tokens: divide by 4 (≈ 4 chars/token)
estimated_tokens = (
    cumulative_output_bytes / 4         # text output (1 KiB ≈ 256 tokens)
  + system_reminder_volume * 200        # per rule-file injection (large rule MD ≈ 200 tokens)
  + large_tool_results * 1000           # 5 KiB+ Read/Bash result (≈ 1000 tokens each)
  + agent_invocations * 5000            # Agent() return context (typical 5K tokens)
)

if estimated_tokens / context_window > 0.75:
    save_state(); emit_resume_message()
if estimated_tokens / context_window > 0.90:
    announce("/clear 권장")
```

**Sanity check**: 1 MiB cumulative output → ~262K tokens (well above Sonnet 200K limit, ~26% of Opus 1M limit). 75% threshold for Opus 1M is reached around ~3 MiB cumulative output.

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Orchestrator가 75% 정확히 detect 못함 | High | Medium | under-estimate 권장 (premature /clear < missed) |
| Resume message format fragmentation | Medium | Medium | 표준 format 동일, extension_data만 differentiate |
| `.moai/state/` 누적 → disk overflow | Low | Low | 7일 retention 권장 (manual) |
| 사용자 paste 시 다른 LLM 혼동 | Low | Low | "ultrathink." prefix가 explicit |
| Cross-ref 누락 워크플로우 미적용 | Medium | Medium | 6 workflow 명시 + audit checklist |
| Self-monitoring 부정확 → 자동 저장 누락 | High | Medium | 사용자가 75% statusline에서 수동 인지 가능 |

## 6. Dependencies

- 선행 SPEC: 없음 (standalone)
- 의존 입력: 본 프로젝트의 기존 `context-window-management.md`
- sibling SPEC: SPEC-CONTEXT-INJ-001 (Wave 3) — context injection 정합
- sibling SPEC: SPEC-CRON-PATTERN-001 (이번 wave) — Pattern P5 연계
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (`.moai/state/<session-id>.json` schema 본 SPEC scope): 권장 schema 명시 (hard rule 후속)
- **OQ2** (75% auto-detect 알고리즘): self-monitoring heuristic, Go 자동화는 후속 SPEC
- **OQ3** (Wave-style format differentiate): Ext5 명시
- **OQ4** (`.moai/state/` retention): 7일 권장 (manual)
- **OQ5** (AskUserQuestion vs 자연어): 자연어 (status announcement) — context-window-management.md 기존 정책 보존

## 8. Rollout Plan

1. M1-M6 구현 후 정책 문서 review
2. 1 long-running workflow (e.g., /moai loop) dogfooding
3. resume message paste-and-resume 검증 (manual)
4. CHANGELOG + v2.x.0 minor release
5. 후속 SPEC: 75% auto-detect Go 자동화 + `.moai/state/` schema hard rule

End of plan.md (SPEC-RESUME-MSG-001).
