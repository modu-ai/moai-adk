# Research — SPEC-PARALLEL-COOK-001 (Parallel Sub-agent Cookbook)

**SPEC**: SPEC-PARALLEL-COOK-001
**Wave**: 3 / Tier 2 (검증 통과 — orchestration pattern document)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Subagents in Claude Code":

> "Three subagents working in parallel complete in roughly the time one would take."

Anthropic blog "Multi-Agent Systems":

> "Parallel orchestration excels when subtasks are independent. The orchestrator's job is to detect independence and dispatch fan-out — then aggregate results in a fan-in step."

> "Aggregation strategy is the bottleneck of parallel agents. If the fan-in agent must reread every output in full, you have lost the time advantage."

### 1.2 검증 (claude-code-guide 검증 완료)

claude-code-guide 에이전트가 Claude Code Agent() spawn 패턴 기준으로 본 권고를 검증함. 결론:

- **호환성**: ✅ 완전 지원 — `Agent()` 도구는 sub-agent 다중 spawn 지원, 본 프로젝트 CLAUDE.md §14 "Parallel Execution" 정책과 일치
- **표준 우회 불필요**: 별도 우회 없이 cookbook 문서화만으로 가치 전달 가능
- **권고 채택**: ACCEPT — Team mode 외 솔로 모드 sub-agent fan-out 표준 패턴 부재 → 신규 cookbook 문서 신설

---

## 2. 현재 상태 (As-Is)

### 2.1 기존 Team mode 패턴

`.claude/rules/moai/workflow/team-pattern-cookbook.md` 존재 (NOTICE.md harness import). 그러나 본 문서는 **Team API (TeamCreate, SendMessage, TaskCreate)** 기반의 multi-agent coordination이며, **단일 orchestrator + 여러 sub-agent fan-out** 패턴은 부분적 커버.

### 2.2 솔로 모드 sub-agent 사용 현황

orchestrator가 `Agent()` 도구로 sub-agent를 호출하는 케이스:

- 단일 호출 (sequential): `manager-ddd → manager-tdd → manager-quality` 순차
- 병렬 호출 (parallel): `Promise.all([Agent(researcher), Agent(analyst)])` 같은 패턴

**관찰**: 어떤 agent 페어가 병렬화 가능한지, 어떻게 fan-in을 책임지는지 표준 cookbook 부재. CLAUDE.md §14 "Parallel Execution Safeguards"가 일반 원칙은 제공하나 **8가지 표준 페어 매트릭스**는 부재.

### 2.3 worktree isolation 정책

CLAUDE.md §14 "Worktree Isolation Rules":

- implementer/tester/designer → `isolation: "worktree"` 필수
- researcher/analyst/reviewer → `isolation: "worktree"` 금지 (read-only)

**관찰**: 정책은 명시되어 있으나 구체적 페어링 시 어떤 worktree 구조를 사용해야 하는지 사례 부재.

### 2.4 fan-in aggregation 패턴

현재 분산되어 있음:
- orchestrator가 직접 결과 통합 (가장 흔함)
- 별도 reviewer/aggregator agent 호출 (드묾)
- shared state file (`.moai/specs/<ID>/progress.md`) 경유 (간헐적)

**관찰**: fan-in 책임 모델이 일관되지 않음 → 개별 PR마다 ad-hoc 결정.

---

## 3. 격차 분석 (Gap Analysis)

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|-------------|-------------|------|
| 솔로 fan-out cookbook | 부재 | 8 페어 매트릭스 | 신규 문서 |
| fan-in 책임 모델 | ad-hoc | 표준 3 패턴 (orchestrator / reviewer / shared-state) | 표준화 필요 |
| 페어 worktree 적용 | 일반 정책만 | 페어별 구체 가이드 | 매핑 필요 |
| 실패 격리 | 없음 | 단일 페어 실패 시 다른 페어 보존 | 패턴 정의 |
| aggregation 거부 룰 | 없음 | aggregation undefined 시 fan-out 거부 | 신규 룰 |
| 사례 베이스라인 | Team cookbook (multi-agent) | Solo cookbook (single orchestrator + N) | 신규 문서 |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 신규 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/rules/moai/development/parallel-subagent-patterns.md` | 신규 | 8 페어 매트릭스 + fan-in 책임 + 실패 격리 |
| `internal/template/templates/.claude/rules/moai/development/parallel-subagent-patterns.md` | 신규 | Template-First 동기화 |

### 4.2 Secondary 수정

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/rules/moai/workflow/team-pattern-cookbook.md` | 수정 (cross-ref 추가) | "솔로 패턴은 parallel-subagent-patterns.md 참조" 명시 |
| `CLAUDE.md` §14 | 수정 (cross-ref 추가) | "구체 페어 사례는 .claude/rules/moai/development/parallel-subagent-patterns.md" |
| `internal/template/templates/CLAUDE.md` | 동일 동기화 |  |

### 4.3 Templates (Template-First Rule 준수)

모든 변경 `internal/template/templates/`에 동기화 후 `make build`.

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Cookbook이 prescriptive하여 새 페어 발굴 저해 | Medium | Medium | 문서 도입부에 "표준 8 페어는 출발점, 새 페어 추가 가이드 포함" 명시 |
| Team cookbook과 의미 충돌 | Low | High | 명시적 cross-ref + 차이 (Team = multi-agent coord, Solo = single orchestrator fan-out) 강조 |
| 페어 매트릭스가 곧 outdated | Medium | Low | living document 마킹, 분기별 검토 |
| 실패 격리 패턴 오해 → 모든 페어 동시 abort | Low | High | 실패 격리 코드 예시 (Promise.allSettled 패턴) 포함 |

### 5.2 Assumptions

- A1: 본 프로젝트의 sub-agent 호출은 항상 `Agent()` 도구 경유
- A2: orchestrator는 `Promise.all` 또는 동등한 동시 dispatch 패턴 사용 가능
- A3: 8 페어는 모두 본 프로젝트 catalog의 agent 조합 (manager-/expert-/builder-)
- A4: cookbook 문서는 Skill 시스템과 별도 (rules/development/ 위치 적합)

---

## 6. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| 표준 페어 수 | 문서 작성 후 카운트 | >= 8 |
| 각 페어 worktree 가이드 | 페어별 명시 | 8/8 |
| fan-in 책임 명시 | 페어별 fan-in agent 지정 | 8/8 |
| 실패 격리 코드 예시 | 문서 내 코드 블록 | >= 1 |
| aggregation 거부 사례 | anti-pattern 절 | >= 3 |

---

## 7. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|----------|------|
| Team cookbook에 솔로 절 추가 | ❌ | 두 패턴은 의미적으로 다름 (multi-agent vs single-orchestrator), 분리가 명료 |
| Skill 형태로 작성 (`.claude/skills/moai/parallel-cook/`) | ❌ | 패턴 문서는 rules가 적합 (skill은 동작 단위) |
| CLAUDE.md §14 확장만 | ❌ | CLAUDE.md는 원칙 수준, 8 페어 매트릭스는 별도 문서 필요 |
| Wave 4로 연기 | ❌ | Tier 2 우선순위, Anthropic 권고 즉시 가치 |

---

## 8. 참고 SPEC

- SPEC-AUTO-DELEG-001 (Wave 1): 자동 delegation 패턴 — 본 SPEC의 fan-out 결정 로직과 보완
- SPEC-V3R2-ORC-005 (team formalization): Team mode 공식화 — solo 패턴과 분리 의미
- SPEC-TEAM-001: Dynamic team generation — solo와 team의 boundary 명시

---

## 9. Open Questions (Plan 단계 해결 대상)

- OQ1: 8 페어가 표준이지만 새 페어 발굴 시 등재 절차는? → plan.md M5에서 결정
- OQ2: fan-in 책임을 orchestrator로 강제 vs reviewer agent 권장 중 default는? → plan.md
- OQ3: cookbook 위반 (aggregation undefined) 시 hook으로 차단할지, 텍스트 가이드만 할지?
- OQ4: 페어 매트릭스에 expert-debug 포함 여부 (디버그는 본질적으로 sequential일 수 있음)

---

End of research.md (SPEC-PARALLEL-COOK-001).
