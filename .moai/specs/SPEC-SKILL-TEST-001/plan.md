---
id: SPEC-SKILL-TEST-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-SKILL-TEST-001

## 1. Overview

builder-skill agent에 testing framework 통합. sample prompts × expected outcomes + metrics + A/B comparator + description optimization 4-pillar 도입. legacy skill은 incremental 마이그레이션.

## 2. Approach Summary

**전략**: Builder-First (책임 명확화) + Incremental Adoption (점진 마이그레이션) + Suggestion-Only (자동 수정 금지).

1. builder-skill 본문에 Testing Framework 절 신설
2. Skill test schema 정의 (YAML)
3. Test result persistence schema 정의 (Markdown)
4. A/B comparator 블라인드 프로토콜 명시
5. Description optimization 제안 메커니즘
6. legacy skill incremental 마이그레이션 정책

## 3. Milestones

### M0 — Pre-flight Verification (Priority: Critical)

- [ ] 현재 builder-skill agent verbatim 캡처 (105 lines)
- [ ] 100+ skill 인벤토리 파악 (`find .claude/skills -name "SKILL.md" | wc -l`)
- [ ] sample skill 5개 선정 (다양한 카테고리: foundation, workflow, domain, agent)
- [ ] sample skill 5개의 frontmatter description + triggers 추출

**Exit Criteria**: skill inventory + 5 sample skill baseline

### M1 — Skill Testing Framework Section in builder-skill (Priority: Critical)

- [ ] `.claude/agents/moai/builder-skill.md`에 "Skill Testing Framework" 절 신설:
  - Testing protocol 5단계: prompts → outcomes → execute → measure → suggest
  - Test schema (sample-prompts.yaml, expected-outcomes.yaml) 명시
  - Metrics: pass_rate, elapsed_time_avg, token_consumption_avg
  - Test result file path: `.moai/research/skill-tests/<skill-id>/<run-id>.md`
  - A/B comparator 블라인드 프로토콜
  - Description optimization 제안 (자동 적용 금지)
- [ ] Template 동기화: `internal/template/templates/.claude/agents/moai/builder-skill.md`

**Exit Criteria**: Testing Framework 절 명시 + Template-First sync

### M2 — Skill Test Schema (YAML) Definition (Priority: High)

- [ ] sample-prompts.yaml schema:
  ```yaml
  skill_id: <skill-id>
  prompts:
    - id: P001
      text: "user query"
      intent: <intent classification>
  ```
- [ ] expected-outcomes.yaml schema:
  ```yaml
  skill_id: <skill-id>
  outcomes:
    - prompt_id: P001
      skill_should_trigger: true|false
      expected_keywords: [...]
      expected_modules_loaded: [...]  # optional
      expected_redirect: <other skill if should_trigger=false>
  ```
- [ ] schema validation rules (YAML structure 검증)
- [ ] sample template 작성: `.claude/skills/<sample-skill>/tests/sample-prompts.yaml`

**Exit Criteria**: schema 문서화 + sample template 작성

### M3 — Test Execution Protocol (Priority: High)

- [ ] builder-skill 본문에 execute 단계 명시:
  - sample prompt를 LLM에 입력
  - skill activation 여부 측정 (binary)
  - expected_keywords 매칭률 측정
  - elapsed_time / token_consumption 기록
  - pass criteria: activation 일치 + keyword 매칭 >= 70%
- [ ] Test result Markdown schema:
  ```markdown
  # Skill Test Run: <skill-id>
  Run ID: <uuid>
  Date: ISO-8601
  Skill version (commit hash): ...
  
  ## Summary
  - Pass rate: X/N
  - Elapsed time avg: X seconds
  - Token consumption avg: X tokens
  
  ## Per-prompt Results
  | Prompt ID | Activated | Keywords matched | Pass | Tokens | Time |
  |-----------|-----------|-----------------|------|--------|------|
  ...
  ```

**Exit Criteria**: execute protocol + result schema 명시

### M4 — A/B Comparator Blind Protocol (Priority: High)

- [ ] builder-skill 본문에 A/B 절 추가:
  - Input: 두 skill version의 output (current vs proposed)
  - Random shuffle: `output_X` / `output_Y` 라벨 (version A/B 매핑은 비공개)
  - Comparator prompt: "Which output better satisfies the user's intent? X / Y / inconclusive"
  - Comparator agent: general-purpose (단일 컨텍스트)
  - Result: winner = X (= current) / Y (= proposed) / null
- [ ] 블라인드 보장 audit: prompt 내부에 file path, version metadata 노출 금지

**Exit Criteria**: A/B 프로토콜 명시 + 블라인드 audit checklist

### M5 — Description Optimization Suggestion (Priority: Medium)

- [ ] builder-skill 본문에 description optimization 절:
  - 분석 input: skill description + sample prompts × expected_should_trigger
  - 분석: description의 trigger keyword가 sample prompts와 어떻게 매칭하는지
  - threshold: <70% 매칭 시 suggestion 트리거
  - Suggestion 형식: "Current description: ... / Suggested: ... / Rationale: ..."
  - 적용 절차: orchestrator → AskUserQuestion → user approval → frontmatter edit
- [ ] 자동 적용 금지 강제: builder-skill agent는 frontmatter 수정 직접 수행 금지

**Exit Criteria**: optimization 절 명시 + 자동 적용 금지 정책 명시

### M6 — Legacy Skill Migration Policy (Priority: Medium)

- [ ] builder-skill 본문에 migration 절:
  - legacy skill (testing 없음)은 modify 시점에만 의무 (즉시 100+ migration 금지)
  - bulk migration script는 별도 SPEC (본 SPEC scope 외)
  - migration trigger: skill body 수정 또는 frontmatter 수정 시
- [ ] CHANGELOG에 명시: "skill testing is opt-in until 2026-Q3, then mandatory for all modifications"

**Exit Criteria**: migration policy 명시

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] sample 5 skill에 testing framework 적용 dogfooding
- [ ] acceptance.md 9 시나리오 PASS
- [ ] M0 baseline 대비 false-trigger reduction 측정
- [ ] plan-auditor PASS
- [ ] Template-First sync clean

**Exit Criteria**: acceptance.md PASS + plan-auditor PASS

## 4. Technical Approach

### 4.1 Skill Test Schema Example (foundation-core)

```yaml
# .claude/skills/moai-foundation-core/tests/sample-prompts.yaml
skill_id: moai-foundation-core
version: "1.0"
prompts:
  - id: P001
    text: "TRUST 5 framework 어떻게 동작해?"
    intent: explain_concept
  - id: P002
    text: "MoAI의 SPEC-First DDD 알려줘"
    intent: explain_workflow
  - id: P003
    text: "/moai run 실행 방법"
    intent: workflow_command
```

```yaml
# .claude/skills/moai-foundation-core/tests/expected-outcomes.yaml
skill_id: moai-foundation-core
outcomes:
  - prompt_id: P001
    skill_should_trigger: true
    expected_keywords: ["TRUST 5", "Tested", "Readable", "Unified"]
  - prompt_id: P002
    skill_should_trigger: true
    expected_keywords: ["SPEC", "DDD", "EARS"]
  - prompt_id: P003
    skill_should_trigger: false  # workflow command → moai-workflow-spec
    expected_redirect: "moai-workflow-spec"
```

### 4.2 A/B Comparator Pseudocode

```
v_current = current skill output
v_proposed = proposed skill output
shuffle = random({A: v_current, B: v_proposed} OR {A: v_proposed, B: v_current})

prompt = """
Compare two outputs (A and B) for user query: "{query}"
Output A: {shuffle.A}
Output B: {shuffle.B}
Which better satisfies the intent? Answer: A / B / inconclusive + rationale.
DO NOT speculate about which version came first.
"""

result = spawn(general-purpose, prompt)
winner = unshuffle(result.choice)  // map A/B back to current/proposed
```

### 4.3 Description Optimization Suggestion Schema

```markdown
## Description Optimization Suggestion for moai-foundation-core

Current: "Foundational principles ... TRUST 5, SPEC-First DDD, ..."
Trigger keyword overlap with sample prompts: 65% (below 70% threshold)

Suggested: "Foundational principles for MoAI development: TRUST 5 quality framework,
SPEC-First DDD workflow, agent delegation patterns, token optimization."

Rationale:
- Sample P003 ("/moai run 실행 방법") incorrectly triggers this skill 60% of time
- Adding "for MoAI development" + workflow context narrows trigger scope
- Predicted false-trigger reduction: -25%

Apply this suggestion? (User approval required via AskUserQuestion)
```

## 5. Risks and Mitigations

| Risk | P | I | Mitigation |
|------|---|---|-----------|
| 100+ skill 일괄 testing → 토큰 폭증 | High | Critical | incremental only (M6) |
| sample prompts 작성 부담으로 채택률 저조 | High | High | 첫 quarter는 opt-in, 자동 prompts 생성 시도 |
| A/B 블라인드 보장 실패 | Medium | High | M4 audit checklist + version label 제거 강제 |
| description 자동 수정 → 의도치 않은 변경 | High | High | C2 강제: AskUserQuestion 승인 필수 |
| Test result file race | Low | Medium | run_id (UUID) 격리 |
| legacy skill migration 부담 | High | Medium | M6 점진 정책, bulk migration은 별도 SPEC |

## 6. Dependencies

- 선행 SPEC: 없음
- 동반 SPEC: SPEC-EVAL-RUBRIC-001 (rubric anchoring 패턴 친화)
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1**: Test 의무화 범위 → ✅ modified skill만 (legacy 일괄 금지, 점진 마이그레이션)
- **OQ2**: A/B comparator input → ✅ prompt 단위 (한 sample prompt × 두 version output)
- **OQ3**: Description optimization 자동 적용 → ❌ 금지 (제안만, AskUserQuestion 승인)
- **OQ4**: Pass rate threshold 80% → ✅ 채택 (skill complexity별 차별은 향후 SPEC)

## 8. Rollout Plan

1. M1 builder-skill body 강화 후 sample 5 skill에 적용 (dogfooding)
2. opt-in 첫 quarter (2026-Q2)
3. mandatory for modifications 다음 quarter (2026-Q3)
4. bulk legacy migration script: 별도 SPEC (Wave 3 후보)

End of plan.md.
