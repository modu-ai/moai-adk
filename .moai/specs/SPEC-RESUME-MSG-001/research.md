# Research — SPEC-RESUME-MSG-001 (Resume Message 강화 / 확산)

**SPEC**: SPEC-RESUME-MSG-001
**Wave**: 4 / Tier 3 (장기/폴리싱)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용 (Anthropic blog "Using Claude Code: Session Management and 1M Context")

> "Bad autocompacts occur when the model can't predict the direction your work is going."

> "When you start a new task, you should also start a new session."

> "Three primary tools: /compact (lossy), /clear (manual but clean), /rewind (Esc Esc to jump back)."

### 1.2 Anthropic의 세션 관리 권고 핵심

- **Bad autocompact 회피**: 모델이 다음 단계를 예측할 수 있도록 **자기-기술적 (self-descriptive) 상태**를 명시
- **Task boundary = session boundary**: 작업 단위가 바뀌면 새 세션 시작이 깨끗
- **/clear is preferred over /compact**: clean state가 lossy compact보다 안전
- **State persistence across /clear**: 외부 파일 (progress.md)에 의존

### 1.3 본 프로젝트의 모범 사례 (이미 보유)

`.claude/rules/moai/workflow/context-window-management.md`이 이미 정착:
- 75% / 90% threshold 명시
- 구조화된 resume message 포맷:
  ```
  ultrathink. Wave <N> 이어서 진행. SPEC-<ID>부터 <approach>.
  applied lessons: <files>.
  progress.md 경로: .moai/specs/SPEC-<ID>/progress.md
  다음 단계: <command>.
  완료 후: <next SPEC or /moai sync>.
  ```

본 SPEC의 사명: 이 모범 사례를 모든 long-running workflow로 **확산**.

---

## 2. 현재 상태 (As-Is)

### 2.1 본 프로젝트의 resume message 보유 워크플로우

**보유**:
- `/moai plan` → `/moai run` 전환 시 (context-window-management.md §Resume message format)

**미보유** (확산 대상):
- `/moai loop` (iterative fix loop)
- `/moai fix` (auto-fix)
- `/moai design` (GAN loop, 5 iterations 가능)
- `/moai sync` (multi-PR aggregate)
- `/moai run` 자체 (multi-task agent chain)
- Wave-style multi-SPEC 작업 (Wave 1-4 패턴)

### 2.2 격차 분석

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|--------------|--------------|------|
| Resume message 표준 | plan/run 전환만 | 모든 long-running workflow | 확산 |
| `.moai/state/<session-id>.json` | 부재 | 75% 도달 시 자동 저장 | 신규 자동화 |
| Auto-load on paste | 부재 | 사용자가 paste 시 자동 인식 | 신규 정책 |
| 75% / 90% threshold 적용 워크플로우 수 | 1 (plan/run) | 5+ workflow | 확산 |
| Resume pattern docs | 1 위치 | 모든 workflow skill cross-ref | 확산 |

### 2.3 표준 resume message format (보유)

본 SPEC은 기존 format을 유지 + 확장:

```
ultrathink. <workflow_name> 이어서 진행. <SPEC-ID 또는 SCOPE>부터 <approach 요약>.
applied lessons: <memory file names>.
progress.md 경로: .moai/specs/SPEC-<ID>/progress.md (또는 .moai/state/<session-id>.json)
다음 단계: <one-line command>.
완료 후: <next SPEC or workflow>.
```

확장 후보 (long-running workflow별):

- `/moai loop`: iteration N/M 명시 + 통과 기준
- `/moai design`: GAN iteration N/5 + 현재 score
- `/moai sync`: pending PR count + 다음 SPEC
- `/moai run` agent chain: completed phase + next agent
- Wave: Wave N/4 + completed SPEC list

---

## 3. 코드베이스 분석 (Affected Files)

### 3.1 Primary 수정 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/rules/moai/workflow/context-window-management.md` | 보강 | resume message 확산 정책 추가 |
| `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | 보강 | Template-First |

### 3.2 Secondary 수정 (cross-ref 추가)

| 파일 | 수정 유형 | 변경 사유 |
|------|-----------|-----------|
| `.claude/skills/moai/workflows/plan.md` | cross-ref | plan workflow에서 resume pattern 인용 |
| `.claude/skills/moai/workflows/run.md` | cross-ref | run workflow에서 resume pattern 인용 |
| `.claude/skills/moai/workflows/sync.md` | cross-ref | sync workflow에서 resume pattern 인용 |
| `.claude/skills/moai/workflows/loop.md` | cross-ref | loop workflow에서 resume pattern 인용 |
| `.claude/skills/moai/workflows/fix.md` | cross-ref | fix workflow에서 resume pattern 인용 |
| `.claude/skills/moai-team-design/SKILL.md` | cross-ref | design GAN loop에서 resume pattern 인용 |

### 3.3 신규 디렉토리

| 경로 | 목적 |
|------|------|
| `.moai/state/` | 세션 상태 영속화 (이미 일부 사용 중) |
| `.moai/state/<session-id>.json` | per-session state (75% 도달 시 자동 저장) |

### 3.4 Auto-save trigger (orchestrator 책임)

본 SPEC은 정책 + 표준 format 확산 중심. 자동 저장 detection은 orchestrator의 self-monitoring:
- 75% 도달 → state 저장 + resume message 생성 (사용자에게 자연어 안내)
- 90% 도달 → /clear 강력 권고

자동화 (Go 코드)는 본 SPEC scope 외 — 후속 SPEC 후보.

---

## 4. 위험 및 가정

### 4.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| Orchestrator가 75% 정확히 detect 못함 | High | Medium | under-estimate 권장 (premature /clear < missed /clear) |
| Resume message format 워크플로우별 fragmentation | Medium | Medium | 표준 format은 동일, 확장 부분만 differentiate |
| `.moai/state/` 누적 → disk overflow | Low | Low | 7일 retention 권장 |
| 사용자가 resume message paste 시 다른 LLM 혼동 | Low | Low | "ultrathink." prefix가 explicit |
| Cross-ref 누락된 workflow에서 정책 미적용 | Medium | Medium | 6 workflow 명시 + audit checklist |

### 4.2 Assumptions

- A1: 본 프로젝트의 기존 resume message 모범 사례가 사용자에게 효과적
- A2: 6 workflow가 long-running 우선순위
- A3: 75%는 Anthropic 1M context model 기준 750K tokens
- A4: `.moai/state/<session-id>.json` 형식은 후속 SPEC에서 정형화
- A5: 자동 저장은 orchestrator self-monitoring (Go 코드 변경 없음)

---

## 5. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| Resume policy 보강 | 문서 diff | 6+ workflow 인용 |
| Cross-ref 카운트 | grep `context-window-management.md` | >= 6 |
| Verbatim 인용 | grep | 3+ 출처 |
| State 디렉토리 | `.moai/state/.gitkeep` | EXISTS |
| Format 확장 5종 | 문서 검토 | 모두 명시 |
| 75% / 90% threshold | 문서 검토 | 보존 |
| Template-First sync | `make build` diff | clean |

---

## 6. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|-----------|------|
| Go 코드로 75% auto-detect 자동화 | ❌ (현재 SPEC 외) | scope 폭발, 후속 SPEC |
| 워크플로우별 별도 resume format | ❌ | fragmentation, 표준 format + 확장이 더 적합 |
| `.moai/sessions/` 위치 | ❌ | 이미 `.moai/state/` 사용 중 |
| Resume message를 MEMORY.md에 영속화 | ❌ | per-session 데이터 (transient) |
| 75% threshold 50%로 하향 | ❌ | premature /clear 폐해, Anthropic 75% 권고 유지 |

---

## 7. 참고 SPEC

- SPEC-CONTEXT-INJ-001 (Wave 3): context injection — 본 SPEC의 자매
- SPEC-MEMO-001: 기존 메모리 시스템 — resume message가 memory file 인용
- SPEC-CRON-PATTERN-001 (이번 wave sibling): Pattern P5 (memory hygiene) — `.moai/state/` 정리에 활용 가능

---

## 8. Open Questions (Plan 단계 해결 대상)

- OQ1: `.moai/state/<session-id>.json` schema 정의는 본 SPEC scope? → 권장 schema 명시, hard rule 후속
- OQ2: 자동 저장 trigger (75% detect) 정확한 알고리즘? → 본 SPEC은 정책만, Go 자동화는 후속
- OQ3: Wave-style 작업의 resume format은 어떻게 differentiate? → 확장 5종 중 하나
- OQ4: `.moai/state/` retention 정책? → 7일 권장 (manual cleanup)
- OQ5: AskUserQuestion 호출 vs 자연어 안내 (75% 도달 시)? → 자연어 (status announcement, not question)

---

End of research.md (SPEC-RESUME-MSG-001).
