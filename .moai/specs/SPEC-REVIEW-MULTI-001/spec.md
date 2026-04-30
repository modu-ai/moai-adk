---
id: SPEC-REVIEW-MULTI-001
status: draft
version: "0.1.0"
priority: High
labels: [review, multi-agent, parallel, verification, severity, wave-2, tier-1]
issue_number: null
scope: [review.md, team/review.md]
blockedBy: []
dependents: []
related_specs: [SPEC-V3R3-PATTERNS-001, SPEC-EVAL-LOOP-001]
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 2
tier: 1
---

# SPEC-REVIEW-MULTI-001: Code Review Multi-Agent 패턴 도입

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 2 / Tier 1. Anthropic "Code Review (with Claude)" 3-stage 패턴 (parallel detection → verification → severity ranking)을 `/moai review`에 도입.

---

## 1. Goal (목적)

Anthropic blog "Code Review (with Claude)"의 3-stage multi-agent 패턴 (parallel detection → verification → severity ranking)을 `/moai review` 워크플로우에 도입한다. 단일 agent 검토의 false-positive 문제를 verification stage로 해소하고, 병렬 detection으로 large PR (>1,000 LOC) coverage를 84% 수준 (Anthropic baseline)에 근접시킨다.

### 1.1 배경

- Anthropic: "We found that a single agent reviewing in isolation produced too many false positives. The verification step cut false positives by more than half."
- 현재 `/moai review`는 manager-quality 단일 위임 (4 perspective sequential) → 한 관점 지배, false-positive filter 없음
- Severity ranking이 manager-quality 내부에 묻혀 있어 일관성 부재

### 1.2 비목표 (Non-Goals)

- 신규 specialized review agent 생성 (기존 expert-* + general-purpose 조합 사용)
- review 외 다른 워크플로우의 multi-agent 일반화 (review 한정)
- review 결과의 자동 fix 적용 (review = 분석, fix는 별도 워크플로우)
- Anthropic baseline (84%/31%) 재현 의무 (참고 metric으로만 사용)
- chunked large-PR 처리 알고리즘 (별도 SPEC, 본 SPEC은 분기 정책만)

---

## 2. Scope (범위)

### 2.1 In Scope

- `.claude/skills/moai/workflows/review.md` 재작성:
  - Stage 1 (Parallel Detection): 4 agent 병렬 (expert-security, expert-performance, manager-quality, expert-refactoring)
  - Stage 2 (Verification): finding당 verifier agent (general-purpose)
  - Stage 3 (Severity Ranking): ranker agent (general-purpose)
  - Large PR (>1,000 LOC) 분기: depth 증가 (expert-debug 추가)
  - Small PR (<50 LOC) 분기: 기존 단일 agent 경로 유지 (token 절약)
- `.claude/skills/moai/team/review.md` 강화: 3-stage with team mode
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- 신규 review agent 신설 (general-purpose 활용)
- review 외 워크플로우의 multi-agent 일반화
- 자동 fix 적용
- chunked large-PR 처리 (token 한계 대응은 별도 SPEC)
- review report의 publish (GitHub PR comment posting 등)
- review 결과의 SPEC 자동 생성

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- Claude Code v2.1.111+
- 영향 디렉터리: `.claude/skills/moai/workflows/`, `.claude/skills/moai/team/`
- 템플릿 동기화: `internal/template/templates/`

---

## 4. Assumptions (가정)

- A1: 4 detection agent을 병렬로 spawn해도 read-only이므로 file race 없음
- A2: verifier는 general-purpose context에서 detection agent와 충분히 독립적
- A3: `git diff --shortstat`으로 PR LOC 측정 가능
- A4: small PR (<50 LOC) 단일 agent 경로 유지가 token 비용 효율적
- A5: ranker가 프로젝트 risk profile (.moai/project/tech.md, structure.md)을 인용 가능

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-RM-001**: THE `/moai review` WORKFLOW SHALL implement 3 distinct stages: Parallel Detection, Verification, Severity Ranking.
- **REQ-RM-002**: THE STAGE 1 (PARALLEL DETECTION) SHALL dispatch 4 agents simultaneously: expert-security, expert-performance, manager-quality, expert-refactoring.
- **REQ-RM-003**: THE STAGE 2 (VERIFICATION) SHALL spawn an independent verifier agent for each candidate finding from Stage 1.
- **REQ-RM-004**: THE STAGE 3 (SEVERITY RANKING) SHALL spawn a dedicated ranker agent that classifies verified findings into Critical / High / Medium / Low.
- **REQ-RM-005**: THE FINAL REVIEW REPORT SHALL group findings by severity in descending order with file:line references for each finding.

### 5.2 Event-Driven Requirements

- **REQ-RM-006**: WHEN `/moai review` is invoked, THE WORKFLOW SHALL measure PR LOC via `git diff --shortstat` to determine routing.
- **REQ-RM-007**: WHEN PR size is less than 50 LOC, THE WORKFLOW SHALL execute single-agent review (legacy manager-quality path) for token efficiency, with multi-agent path opt-in via `--full` flag.
- **REQ-RM-008**: WHEN PR size is between 50 and 1,000 LOC, THE WORKFLOW SHALL execute the 3-stage multi-agent review by default.
- **REQ-RM-009**: WHEN PR size exceeds 1,000 LOC, THE WORKFLOW SHALL execute the 3-stage review AND additionally include expert-debug in Stage 1 detection agents.
- **REQ-RM-010**: WHEN all Stage 1 detection agents complete, THE WORKFLOW SHALL aggregate findings into a unified candidate list before invoking Stage 2.

### 5.3 State-Driven Requirements

- **REQ-RM-011**: WHILE Stage 1 detection agents are running in parallel, THE ORCHESTRATOR SHALL collect their results without enforcing inter-agent communication.
- **REQ-RM-012**: WHILE Stage 2 verification is in progress, THE ORCHESTRATOR SHALL track which findings are verified vs dropped, with rationale captured for each drop.

### 5.4 Conditional Requirements

- **REQ-RM-013**: WHERE Stage 2 verification fails for a finding (no reproducer or weak evidence), THEN THE FINDING SHALL be dropped from the candidate list with the drop reason logged.
- **REQ-RM-014**: WHERE the project's `.moai/project/tech.md` declares a sensitive domain (auth, payment, public_api), THEN THE RANKER SHALL elevate Security findings by one severity level (e.g., Medium → High).
- **REQ-RM-015**: IF the same finding is reported by multiple Stage 1 agents, THEN THE RANKER SHALL deduplicate by file:line + symptom signature, retaining all originating agents in metadata.
- **REQ-RM-016**: WHERE the SPEC opts into worktree isolation via review flag `--isolated`, THEN Stage 1 agents SHALL run in `isolation: "worktree"` (read-only).
- **REQ-RM-017**: IF Stage 1 detection agents collectively produce zero candidate findings, THEN THE WORKFLOW SHALL skip Stage 2 and Stage 3, producing a "no findings" report.

### 5.5 Unwanted (Negative) Requirements

- **REQ-RM-018**: THE STAGE 1 DETECTION AGENTS SHALL NOT communicate with each other during parallel execution; their context isolation is essential for independence.
- **REQ-RM-019**: THE STAGE 2 VERIFIER SHALL NOT be the same agent that produced the finding (independence requirement).
- **REQ-RM-020**: THE WORKFLOW SHALL NOT proceed past Stage 2 with unverified findings included in the final report.

---

## 6. Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Detection coverage vs single-agent baseline | sample 5 PRs comparison | > 80% more findings |
| False-positive rate (verifier-dropped) | dropped / total candidates | < 15% |
| Large PR (>1,000 LOC) finding rate | sample 5 large PRs | reasonable (Anthropic baseline ~84%) |
| Token cost vs single-agent | per-PR token | < +200% (4x agents + verifiers + ranker) |
| Wall-clock time vs single-agent | parallel stage 1 timing | < +50% (parallelism benefit) |
| Template-First sync | `make build` diff | clean |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: 신규 specialized review agent 신설 금지 (general-purpose 활용)
- C2: review 외 워크플로우 변경 금지
- C3: small PR (<50 LOC) 단일 agent 경로 유지 (token 절약)
- C4: Stage 1 detection agent 간 통신 금지 (independence 핵심)
- C5: verifier는 originating agent와 다른 컨텍스트여야 함

End of spec.md (SPEC-REVIEW-MULTI-001 v0.1.0).
