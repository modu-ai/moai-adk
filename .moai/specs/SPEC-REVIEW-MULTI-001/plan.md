---
id: SPEC-REVIEW-MULTI-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-REVIEW-MULTI-001

## 1. Overview

`/moai review` 워크플로우에 Anthropic Code Review 3-stage 패턴 (parallel detection → verification → severity ranking) 도입. 기존 expert-* agent를 재활용하고 verifier/ranker는 general-purpose Agent()로 구현.

## 2. Approach Summary

**전략**: PR Size-Aware Routing + Reuse Existing Agents.

1. PR LOC 측정 → 3-tier routing (small / medium / large)
2. Stage 1 4-agent 병렬 (read-only이므로 worktree 미적용)
3. Stage 2 finding당 verifier (general-purpose, 독립 컨텍스트)
4. Stage 3 ranker (general-purpose, project risk profile 반영)

## 3. Milestones

### M0 — Pre-flight Verification (Priority: Critical)

- [ ] 현재 `review.md` workflow verbatim 캡처
- [ ] `team/review.md` 분석 — 이미 구현된 부분 파악
- [ ] sample 5 PR 식별 (small / medium / large 각 카테고리)
- [ ] baseline single-agent review 결과 캡처 (finding count, false-positive rate, token cost)

**Exit Criteria**: 기존 review 동작 baseline 기록

### M1 — review.md Stage 1 (Parallel Detection) 구현 (Priority: Critical)

- [ ] `.claude/skills/moai/workflows/review.md`에 Phase 2 (Multi-Perspective Analysis) 재작성:
  - PR LOC 측정 단계 추가 (`git diff --shortstat`)
  - 50/1,000 LOC threshold 분기 명시
  - 4 detection agent 병렬 spawn 명시 (모두 `mode: "plan"` 읽기 전용):
    - expert-security: OWASP Top 10, secrets, injection, auth/session
    - expert-performance: latency, memory, complexity, hotspot
    - manager-quality: TRUST 5, style consistency, naming
    - expert-refactoring: code smell, duplication, maintainability
  - Large PR (>1,000 LOC): expert-debug 추가 spawn (또한 read-only)
  - finding aggregation 형식 표준화 (file:line + symptom + originating_agent)
  - HARD rule 준수: 4 detection agent 모두 reviewer role → `isolation: "worktree"` 사용 금지 (CLAUDE.md §14, REQ-RM-016)

**Exit Criteria**: Stage 1 절 완성, agent 4종 병렬 spawn pattern 명시 (read-only, no worktree)

### M1.1 — Read-Only / No-Worktree Isolation Matrix (Priority: Critical)

본 SPEC의 모든 review agent는 **읽기 전용 (reviewer role)** 이며 CLAUDE.md §14 [HARD] 규칙에 따라 worktree isolation 을 사용하지 않는다.

| Agent | Stage | Mode | Worktree isolation | Source REQ |
|-------|-------|------|--------------------|------------|
| expert-security | 1 (detection) | `plan` (read-only) | NEVER | REQ-RM-016 |
| expert-performance | 1 (detection) | `plan` (read-only) | NEVER | REQ-RM-016 |
| manager-quality | 1 (detection) | `plan` (read-only) | NEVER | REQ-RM-016 |
| expert-refactoring | 1 (detection) | `plan` (read-only) | NEVER | REQ-RM-016 |
| expert-debug (large PR) | 1 (detection) | `plan` (read-only) | NEVER | REQ-RM-016 |
| general-purpose verifier | 2 (verification) | `plan` (read-only) | NEVER | REQ-RM-016 |
| general-purpose ranker | 3 (severity) | `plan` (read-only) | NEVER | REQ-RM-016 |

No flag, configuration override, or opt-in path may relax this constraint. EC-5 in acceptance.md describes the rejection behavior.

**Exit Criteria**: 본 매트릭스가 plan.md에 명시되었고 review.md / team/review.md 구현이 이를 따른다.

### M2 — review.md Stage 2 (Verification) 구현 (Priority: High)

- [ ] Stage 2 절 신설:
  - finding당 verifier agent (general-purpose, `mode: "plan"` read-only) spawn
  - verifier prompt: §5.6의 4-step algorithm (Step 1 Reproducer / Step 2 AST·grep / Step 3 OWASP·CWE / Step 4 Confidence floor)을 verbatim 인용
  - verifier output schema (per §5.6): `{verified, total_confidence, step1, step2, step3, rationale, drop_reason?}`
  - drop된 finding은 final report의 "Dropped Findings" 섹션에 메타데이터로만 보존 (failed_steps + total_confidence 포함)
- [ ] independence 강제:
  - verifier는 originating agent와 다른 컨텍스트 (general-purpose)
  - verifier 프롬프트에서 originating_agent 필드 strip (Scenario 9)
- [ ] HARD rule 준수: verifier는 read-only (`mode: "plan"`), worktree isolation 사용 금지 (REQ-RM-016)

**Exit Criteria**: Stage 2 절 완성, §5.6 algorithm verbatim 포함, verifier independence + read-only 규칙 명시

### M3 — review.md Stage 3 (Severity Ranking) 구현 (Priority: High)

- [ ] Stage 3 절 신설:
  - ranker agent (general-purpose) spawn
  - ranker input: verified findings + .moai/project/tech.md (risk profile)
  - ranker output: Critical / High / Medium / Low 분류
  - 중복 finding 머지 (file:line + symptom signature)
  - 프로젝트 risk profile 기반 severity 조정 (auth/payment/public_api → Security finding +1 level)
- [ ] Final Report 형식 표준화:
  - severity별 group
  - 각 finding: file:line + symptom + originating_agent(s) + verification rationale + recommendation

**Exit Criteria**: Stage 3 절 완성, final report schema 명시

### M4 — Small PR Path (Token Optimization) (Priority: Medium)

- [ ] `<50 LOC` 분기:
  - 기존 manager-quality 단일 agent 경로 유지 (legacy fallback)
  - `--full` flag로 multi-agent 경로 opt-in 가능
  - Anthropic baseline (small PR 31% finding rate)는 작은 PR에서는 multi-agent 가치가 낮음을 시사

**Exit Criteria**: small PR 경로 명시, opt-in flag 명시

### M5 — Team Mode Variant (`team/review.md`) (Priority: Medium)

- [ ] `.claude/skills/moai/team/review.md` 강화:
  - 4 dedicated reviewer teammate (mode: "plan", read-only)
  - verifier teammate
  - ranker teammate
  - file ownership: 모두 read-only (review는 write 없음)
  - SendMessage 통신 패턴 명시
- [ ] `--team` flag 분기 명확화

**Exit Criteria**: team/review.md 3-stage with team mode

### M6 — Validation + Acceptance Sign-off (Priority: High)

- [ ] M0 baseline 5 PR을 새 review로 재실행
- [ ] acceptance.md 9 시나리오 PASS
- [ ] detection coverage > +80% (single-agent 대비)
- [ ] false-positive rate < 15%
- [ ] token cost < +200%, wall-clock < +50%
- [ ] plan-auditor PASS
- [ ] Template-First sync clean

**Exit Criteria**: acceptance.md PASS + plan-auditor PASS

## 4. Technical Approach

### 4.1 Stage 1 Pseudocode

```
loc = git_diff_shortstat()
agents = [expert-security, expert-performance, manager-quality, expert-refactoring]
if loc > 1000:
  agents.append(expert-debug)
elif loc < 50 and not opt_in_full:
  return single_agent_review(manager-quality)

findings = parallel_spawn(agents, prompt=review_prompt(diff))
candidates = aggregate(findings)  // file:line + symptom + originating_agent
```

### 4.2 Stage 2 Pseudocode

The verifier prompt MUST encode the §5.6 4-step algorithm verbatim (steps 1–4 + confidence floor). The pseudocode below shows the orchestrator-side loop; the algorithmic decision is delegated to the verifier prompt.

```
verified = []
dropped = []
for finding in candidates:
  # verifier_prompt(finding, diff) renders the §5.6 algorithm with step1..step4
  # plus the confidence floor (>= 0.50) and the strip of originating_agent metadata
  result = spawn(general-purpose, mode="plan",
                 prompt=verifier_prompt_with_algorithm(finding_minus_originating_agent, diff_hunks))
  # result schema:
  #   {verified, total_confidence, step1, step2, step3, rationale, drop_reason?}
  if result.verified == true:
    verified.append(finding | result)  # carry confidence breakdown forward
  else:
    dropped.append(finding | result.drop_reason | failed_steps)
```

Notes:
- `mode="plan"` enforces read-only verifier (REQ-RM-016).
- `finding_minus_originating_agent` removes the `originating_agent` field before serialization (REQ-RM-019, Scenario 9).
- `verifier_prompt_with_algorithm` is a prompt template that quotes §5.6 Step 1–4 verbatim and asks the verifier to emit the exact `verifier_output` schema.

### 4.3 Stage 3 Pseudocode

```
ranked = spawn(general-purpose, prompt=ranker_prompt(verified, risk_profile))
// ranker dedupes, classifies severity, applies risk-based elevation
return final_report(ranked, dropped_metadata)
```

### 4.4 Final Report Schema (Markdown)

```markdown
# Code Review Report

## Summary
- PR size: X LOC
- Findings: N (verified) / M (dropped)
- Severity breakdown: Critical X / High Y / Medium Z / Low W

## Critical Findings
### F-001: <symptom>
- Files: src/...:line
- Originating agents: [expert-security]
- Verification: <reasoning>
- Recommendation: <fix>

## High / Medium / Low Findings
... (same structure)

## Dropped Findings (metadata only)
- Total dropped: M
- Drop reasons: <distribution>
```

## 5. Risks and Mitigations

| Risk | P | I | Mitigation |
|------|---|---|-----------|
| Token cost > +200% (4 detection + N verifier + 1 ranker) | High | High | small PR 분기로 단일 경로 유지, --full opt-in 강제 |
| Verifier independence 부족 (general-purpose context overlap) | Medium | Medium | originating agent identity verifier에 노출 금지 |
| Ranker severity 일관성 부재 | Medium | Medium | risk profile 명시 + sample-based calibration |
| Large PR token 한계 | High | Critical | M0에서 large PR token 측정, chunked review는 별도 SPEC |
| 4 detection agent 간 finding 중복 → ranker 부담 | Medium | Low | dedup 알고리즘 file:line + symptom signature |

## 6. Dependencies

- 선행 SPEC: 없음 (기존 agent 활용)
- 동반 SPEC: SPEC-EVAL-LOOP-001 (Generator-Verifier 패턴 친화)
- 도구: `make build`, plan-auditor, sample PR 5개

## 7. Open Questions Resolution

- **OQ1**: small PR threshold = 50 LOC → ✅ 채택 (Anthropic baseline)
- **OQ2**: large PR threshold = 1,000 LOC → ✅ 채택
- **OQ3**: verifier finding당 1번 spawn → ✅ 채택 (병렬 가능, 독립성 강화)
- **OQ4**: ranker input은 findings + risk profile → ✅ 채택, 과거 triage는 별도 SPEC

## 8. Rollout Plan

1. M1-M3 구현 후 dogfooding (본 SPEC 자체 review)
2. M0 baseline 5 PR 재실행 → 비교 보고서 작성
3. CHANGELOG entry + minor release
4. 부정적 시그널 (token > +250% 또는 false-positive 증가) 시 small PR 경로 default로 회귀

End of plan.md.
