# Research — SPEC-WT-DOC-001 (Worktree Shared State 명시)

**SPEC**: SPEC-WT-DOC-001
**Wave**: 3 / Tier 2 (검증 ⚠️ 우회 가능 — 문서화로 충분)
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Multi-Agent Coordination Patterns" (Shared State 절):

> "Reactive loops are a behavioral problem requiring first-class termination conditions."

> "Agents operate autonomously, reading from and writing to a shared database, file system, or document. The pattern works when the schema is well-defined and concurrent access is bounded. It fails when agents implicitly contend for the same key."

> "Shared state without termination becomes a synchronization graveyard. Define when state transitions terminate, who is the writer of last resort, and what the consistency model is."

### 1.2 검증 (claude-code-guide 검증 완료)

claude-code-guide 에이전트가 본 프로젝트의 git worktree isolation 메커니즘 기준으로 검증함. 결론:

- **호환성**: ⚠️ Worktree isolation으로 동시성 격리는 강력함. shared state는 명시적 cross-worktree merge 시점에만 발생
- **표준 우회**: 추가 인프라 불필요. **`.claude/rules/moai/workflow/worktree-integration.md` 확장만으로 충분**
- **권고 채택**: ACCEPT — 단, 코드 변경 없이 문서 정책 강화

---

## 2. 현재 상태 (As-Is)

### 2.1 Worktree integration rule 위치

`.claude/rules/moai/workflow/worktree-integration.md`:
- worktree 선택 결정 트리
- isolation 정책 (implementer/tester/designer = worktree, researcher/analyst = direct)
- branch naming 규칙

**관찰**: 동시성/cross-worktree shared state 정책 부재.

### 2.2 shared state 사용처

현재 .moai/specs/<ID>/ 디렉토리 내 파일 중 shared state로 사용되는 후보:

| 파일 | 사용 패턴 | 동시성 위험 |
|------|----------|-----------|
| `progress.md` | run 중 갱신 | 단일 worktree 내 sequential, cross-worktree에서는 분기 후 merge 시 | 
| `acceptance.md` | plan 후 read-only (보통) | 변경 가능성 낮음 |
| `spec.md` | plan 단계 작성, 이후 read-only | 변경 가능성 낮음 |
| `_status.json` | hook이 갱신 | 단일 worktree 내 race 가능 |

### 2.3 cross-worktree state synchronization 사례

CLAUDE.local.md §18.11에서 "v2.14.0 case study"로 stacked PR + squash merge 이슈가 기록됨. 이는 worktree간 history divergence가 불완전 merge 시 발생.

**관찰**:
- worktree 자체는 isolated이지만 PR merge 후 main에 병합될 때 shared state 변경 발생
- 그동안의 termination 조건 / writer-of-last-resort 정책 부재

### 2.4 SPEC-LOOP-TERM-001 (Wave 2) 가용성

Wave 2의 SPEC-LOOP-TERM-001은:
- termination 조건 표준 schema (max_iterations, improvement_threshold, escalation_after)
- 일반 evaluator-active loop에 적용

**관찰**: 본 SPEC은 LOOP-TERM-001의 termination schema 의미를 worktree shared state context로 확장 인용.

---

## 3. 격차 분석 (Gap Analysis)

| 영역 | 현재 (As-Is) | 목표 (To-Be) | 격차 |
|------|-------------|-------------|------|
| 동시 쓰기 정책 | undefined | per-file ownership 명시 | 정책 신설 |
| cross-worktree merge 정책 | git default | "PR merge to main, no direct cross-worktree write" 명시 | 정책 명시 |
| shared state termination | undefined | LOOP-TERM-001 schema 인용 | 의미 명시 |
| anti-pattern 카탈로그 | 부재 | 5개 anti-pattern | 신규 |
| worktree-integration.md 규모 | 기본만 | shared state 절 추가 | 확장 |
| consistency model | undefined | "eventual via PR merge, no direct cross" | 명시 |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 수정 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/rules/moai/workflow/worktree-integration.md` | 수정 (확장) | shared state 절 + 5 anti-pattern 추가 |
| `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | 수정 (동일) | Template-First 동기화 |

### 4.2 Secondary 수정

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `CLAUDE.md` §14 | 수정 (cross-ref) | "shared state 정책은 worktree-integration.md §<NEW>" |
| `.claude/rules/moai/core/moai-constitution.md` | 검토 (변경 없음 권장) | constitutional level은 변경 없이 worktree rule이 흡수 |

### 4.3 변경 없음 (no code change)

본 SPEC은 **문서 정책 강화 only**. Go 코드 변경 없음, hook 변경 없음, agent body 변경 없음 (cross-ref 외).

### 4.4 추가될 anti-pattern 5개 (계획 초안)

1. **Direct cross-worktree write**: worktree A에서 worktree B의 `progress.md` 직접 쓰기 → 금지
2. **Concurrent SPEC modification**: 두 worktree에서 동일 SPEC의 spec.md 동시 변경 → PR conflict
3. **Untermenated reactive loop**: shared state를 무한히 재읽어 loop → SPEC-LOOP-TERM-001 schema 위반
4. **Schema drift**: progress.md 형식이 worktree마다 다름 → fan-in 실패
5. **Implicit shared mutable**: `.moai/state/` 임시 파일을 cross-worktree에서 읽기 → undefined

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| 문서 정책만으로 강제 불가 | High | Medium | living document + plan-auditor가 anti-pattern 감지 (별도 SPEC) |
| 기존 worktree workflow 깨짐 | Low | High | 정책은 prescriptive하지만 enforcement는 점진 |
| LOOP-TERM-001 의존성 의미 충돌 | Low | Medium | LOOP-TERM-001을 인용만 (요건 추가 없음) |
| anti-pattern 5개로 커버 부족 | Medium | Low | living document 확장 가능 |
| 사용자 학습 곡선 | High | Low | examples + decision tree 포함 |

### 5.2 Assumptions

- A1: worktree-integration.md가 표준 reference로 자리잡음
- A2: SPEC-LOOP-TERM-001의 termination schema는 worktree shared state context에 의미 확장 가능
- A3: 5개 anti-pattern은 본 프로젝트의 실제 incident에서 도출 (CLAUDE.local.md §18.11 참조)
- A4: 코드 변경 없는 문서 SPEC은 빠른 머지 가능
- A5: cross-worktree state는 PR merge 외 방법 없음 (Git semantics)

---

## 6. 측정 계획 (Baseline + Validation)

| Metric | Baseline 측정 방법 | 목표 |
|--------|-------------------|------|
| anti-pattern 카운트 | 문서 검토 | >= 5 |
| 결정 트리 명시 | 문서 내 flowchart | EXISTS |
| LOOP-TERM-001 인용 | 명시적 cross-ref | EXISTS |
| 4개 cross-ref (CLAUDE.md, agents 등) | 검토 | 1+ |
| Template-First sync | `make build` diff | clean |

---

## 7. 대안 검토 (Alternatives Considered)

| 대안 | 채택 여부 | 이유 |
|------|----------|------|
| 별도 새 rule 파일 (`shared-state-policy.md`) | ❌ | worktree-integration.md에 흡수가 응집도 높음 |
| Go 코드로 cross-worktree write 차단 | ❌ | over-engineering, Git에서 자연스러운 isolation 활용 |
| Hook 기반 정책 enforcement | ❌ (별도 SPEC 후보) | 문서 + 사람 review가 1차, 자동화는 검증 후 |
| MoAI Constitution 격상 | ❌ | constitutional level은 변경 없이 worktree rule 흡수 |
| Wave 4로 연기 | ❌ | 본 SPEC은 문서만이라 빠른 가치 |

---

## 8. 참고 SPEC

- SPEC-LOOP-TERM-001 (Wave 2): termination schema 인용 — 본 SPEC의 의미적 기반
- SPEC-PARALLEL-COOK-001 (이번 wave sibling): solo fan-out 패턴 — worktree shared state 사례 참조
- SPEC-MEM-SCOPE-001 (이번 wave sibling): memory scope의 audit log — shared state의 concurrency 보호 사례
- SPEC-CI-MULTI-LLM-001: CI worktree 활용 사례

---

## 9. Open Questions (Plan 단계 해결 대상)

- OQ1: anti-pattern을 plan-auditor가 자동 감지 가능 (텍스트 매칭)? 아니면 review-only?
- OQ2: shared state termination을 frontmatter 필드로 노출할지 (e.g., `shared_state_termination: <ref>`)?
- OQ3: cross-worktree merge 시 progress.md 충돌 해결 정책 (last-writer / merge-on-PR / manual)?
- OQ4: 정책 위반 시 panic vs warning 강도?

---

End of research.md (SPEC-WT-DOC-001).
