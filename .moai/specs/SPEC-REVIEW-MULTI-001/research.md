# Research — SPEC-REVIEW-MULTI-001 (Code Review Multi-Agent 패턴)

**SPEC**: SPEC-REVIEW-MULTI-001
**Wave**: 2 / Tier 1
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Code Review (with Claude)":

> "When a PR is opened, Code Review dispatches a team of agents. The agents look for bugs in parallel, verify bugs to filter out false positives, and rank bugs by severity."

> "On large PRs (over 1,000 lines changed), 84% get findings, averaging 7.5 issues. On small PRs under 50 lines, that drops to 31%, averaging 0.5 issues."

> "We found that a single agent reviewing in isolation produced too many false positives — flags that, on closer inspection, were intended behavior or non-issues. The verification step (a second agent re-examining each candidate finding) cut false positives by more than half."

> "Severity ranking is its own subagent because the criteria for 'critical' depend on the system, not just the symptom. The ranker reads the project's risk profile and the team's prior triage decisions."

### 1.2 검증 (claude-code-guide 결과)

claude-code-guide:
- **호환성**: ✅ 표준 multi-agent 오케스트레이션 패턴
- **권고 채택**: ACCEPT
- **고려사항**: 4 detection agent 병렬 실행은 worktree isolation 필수

---

## 2. 현재 상태 (As-Is)

### 2.1 review.md 워크플로우 분석

`.claude/skills/moai/workflows/review.md`:
- 현재 default flow: manager-quality 단일 위임 (Phase 2 "Multi-Perspective Analysis")
- manager-quality는 "review from all 4 perspectives sequentially"
- `--team` flag로 team mode (`team/review.md`) 사용 가능
- expert-security는 "Dependency Vulnerability Scan"에만 명시 (full perspective는 manager-quality 내부)

**격차**:
- 단일 agent (manager-quality)가 4 perspective 모두 sequential 처리 → 한 가지 관점 지배 위험
- false positive filter 없음 (verification 부재)
- severity ranking이 독립 단계로 분리되지 않음

### 2.2 Existing Agent Catalog

CLAUDE.md §4 Agent Catalog:

- expert-security: OWASP Top 10, secrets detection
- expert-performance: latency, memory, complexity
- manager-quality: TRUST 5 + style
- expert-refactoring: code smell, maintainability
- expert-debug: 복잡 디버깅 (대규모 PR 시 추가)

→ Stage 1 Parallel Detection에 모두 사용 가능.

### 2.3 Worktree Isolation Rules

CLAUDE.md §14 + worktree-integration.md:
- 4 detection agent을 동시에 spawn할 때 implementation teammate가 아니므로 (read-only review) `isolation: "worktree"` 미사용 권장
- 그러나 PR 컨텍스트에서 git worktree 분리는 안전성 향상 — 채택 검토

### 2.4 Team Mode Variant

`.claude/skills/moai/team/review.md` (이미 존재) 분석:
- 4 dedicated reviewers 병렬 패턴 일부 구현됨
- 그러나 verification + ranking 단계는 명시 없음 (추정)

---

## 3. 격차 분석

| 영역 | As-Is | To-Be | 격차 |
|------|-------|-------|------|
| Stage 1 (Parallel Detection) | manager-quality 단일 sequential | 4 agent 병렬 | 신규 오케스트레이션 |
| Stage 2 (Verification) | 없음 | dedicated verifier per finding | 신규 |
| Stage 3 (Severity Ranking) | manager-quality 내부 부분 처리 | dedicated ranker agent | 신규 분리 |
| Large PR 대응 (>1,000 LOC) | 일반 review와 동일 | depth 증가 (expert-debug 추가) | 신규 분기 |
| Anthropic 84% finding rate (large PR) | 측정 안 됨 | 측정 + 비교 | metric 도입 |

---

## 4. Stage Architecture 설계 초안

### 4.1 Stage 1 — Parallel Detection (4 agents)

| Agent | 검토 영역 |
|-------|----------|
| expert-security | OWASP Top 10, secrets, injection, auth/session |
| expert-performance | latency hotspot, memory leak, complexity |
| manager-quality | TRUST 5, style consistency, naming |
| expert-refactoring | code smell, duplication, maintainability |

병렬 실행 → 모두 finding list 반환.

### 4.2 Stage 2 — Verification (False Positive Filter)

각 finding에 대해 별도 verifier agent (general-purpose, 독립 컨텍스트)가:
- 재현 가능한 reproducer 또는 명확한 증거 검증
- 검증 실패 → finding drop
- 검증 성공 → next stage

verifier는 finding을 작성한 agent와 다른 컨텍스트에서 평가 (independence 핵심).

### 4.3 Stage 3 — Severity Ranking

dedicated ranker agent:
- 검증된 findings를 severity (Critical / High / Medium / Low)로 분류
- 중복 findings 머지
- 프로젝트 risk profile 반영 (예: payment 도메인은 Security finding이 자동 High+)

---

## 5. 코드베이스 분석 (Affected Files)

### 5.1 Primary 수정 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/skills/moai/workflows/review.md` | 재작성 | 3-stage 오케스트레이션 |
| `.claude/skills/moai/team/review.md` | 강화 | 3-stage with team mode |

### 5.2 신규 또는 강화 가능한 agent

본 SPEC은 기존 agent 활용. 신규 agent 생성 없음.
- ranker는 `general-purpose` Agent()로 구현 (전용 agent 신설 불요)
- verifier도 `general-purpose` Agent()

### 5.3 Templates

- `internal/template/templates/.claude/skills/moai/workflows/review.md`: 동일 변경
- `internal/template/templates/.claude/skills/moai/team/review.md`: 동일 변경

---

## 6. 위험 및 가정

### 6.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| 4 detection + N verifier + 1 ranker → 토큰 비용 폭증 | High | High | 작은 PR (<50 LOC)는 기존 단일 agent 경로 유지, threshold 분기 |
| 병렬 worktree race | Medium | Medium | review는 read-only이므로 worktree 미적용 가능 |
| verifier가 부적절한 finding drop | Medium | High | drop 사유 logging, sampling으로 audit |
| ranker의 severity 일관성 부재 | Medium | Medium | rubric 명시 + 프로젝트 risk profile 반영 |
| Large PR (>1,000 LOC) 시 token 한계 | High | Critical | chunked review (파일 단위 분할), summary 단계 |

### 6.2 Assumptions

- A1: 4 detection agent이 병렬 실행 가능 (read-only이므로 file race 없음)
- A2: verifier가 detection agent와 다른 컨텍스트 (general-purpose)에서 충분히 독립적
- A3: PR LOC 측정으로 large/small 분기 가능 (`git diff --shortstat`)
- A4: Anthropic baseline (84% large, 31% small)을 우리 환경에서 재현하지 않음 (참고 metric)

---

## 7. 측정 계획

| Metric | 측정 방법 | 목표 |
|--------|----------|------|
| Detection coverage vs single-agent baseline | 동일 PR을 두 모드로 review | > 80% 더 많은 finding |
| False-positive rate | verifier에 의해 drop된 finding / 전체 finding | < 15% |
| Large PR (>1,000 LOC) finding rate | sample 5 PRs | ~84% (Anthropic baseline) |
| Small PR (<50 LOC) finding rate | sample 5 PRs | ~31% (Anthropic baseline) |
| Token cost vs single-agent | per-PR token measurement | < +200% (4x agent + verifier 고려) |
| Review duration | wall-clock time | < +50% (병렬 효과) |

---

## 8. 대안 검토

| 대안 | 채택 | 이유 |
|------|-----|------|
| 단일 manager-quality 유지 | ❌ | Anthropic 명시: "single agent produced too many false positives" |
| 4 detection만 (verification 생략) | ❌ | false-positive 핵심 문제 잔존 |
| 신규 reviewer agents 신설 (4개) | ❌ | 기존 expert-* agent 충분, 복잡도 증가 |
| Stage별 sequential (병렬 X) | ❌ | 시간 +300%, parallelism 핵심 가치 |
| 3-stage with general-purpose verifier/ranker | ✅ | 기존 agent + general-purpose 조합으로 가성비 우수 |

---

## 9. 참고 SPEC

- SPEC-V3R3-PATTERNS-001: multi-agent 패턴 일반화
- SPEC-EVAL-LOOP-001 (Wave 2 sibling): Generator-Verifier 패턴 (review의 verification stage와 유사)
- SPEC-AGENT-002: 에이전트 신설 가이드 (참고만)

---

## 10. Open Questions

- OQ1: small PR threshold = 50 LOC? Anthropic baseline 그대로 채택 → ✅
- OQ2: large PR threshold = 1,000 LOC 또는 다른 값? → 1,000 채택 (Anthropic baseline)
- OQ3: verifier agent를 SPEC당 1번 spawn하나 finding당 1번 spawn하나? → finding당 1번 (병렬 가능, 독립성 강화)
- OQ4: ranker agent에 어떤 input 제공? findings + project risk profile + (선택) 과거 triage 이력 → spec에 명시

---

End of research.md (SPEC-REVIEW-MULTI-001).
