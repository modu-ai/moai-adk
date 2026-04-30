# Research — SPEC-SKILL-TEST-001 (Skill Testing Framework)

**SPEC**: SPEC-SKILL-TEST-001
**Wave**: 2 / Tier 1
**Created**: 2026-04-30
**Author**: manager-spec

---

## 1. 출처 (Anthropic 공식 자료)

### 1.1 Verbatim 인용

Anthropic blog "Improving Skill Creator: Test, Measure, and Refine Agent Skills":

> "Testing turns a skill that seems to work into one you know works."

> "Skill-creator implements: Testing via evals, Measurement through benchmarks (pass rate, elapsed time, token usage), Refinement loops."

> "A/B comparison via comparator agents judging outputs without knowing which version they're evaluating."

> "Without testing, every skill is hopeful documentation. Testing is the difference between 'this should work' and 'this passes 9 of 10 sample prompts'."

> "We found that descriptions matter more than skill body content for trigger accuracy. A skill with a perfect body but a vague description fires at the wrong time or fails to fire at all."

### 1.2 검증 (claude-code-guide 결과)

claude-code-guide:
- **호환성**: ✅ Skill testing은 표준 관행
- **권고 채택**: ACCEPT
- **고려사항**: 100+ skill을 일괄 테스트할 인프라 필요

---

## 2. 현재 상태 (As-Is)

### 2.1 builder-skill agent 분석

`.claude/agents/moai/builder-skill.md` (105 lines):
- skill 생성/수정 담당
- testing framework 부재
- description 분석 메커니즘 없음
- A/B comparison 없음

### 2.2 Skill Inventory 추정

`.claude/skills/moai/`에 100+ skill 추정. 각 skill:
- frontmatter (name, description, triggers)
- body (markdown)
- 일부는 modules/ 하위 progressive disclosure 사용

### 2.3 Trigger Mechanism

skill frontmatter `triggers.keywords`로 활성화 결정. 그러나:
- 실제로 keyword가 trigger되는지 측정 메커니즘 없음
- false-positive (잘못 활성화) / false-negative (놓침) 측정 없음
- description의 trigger 정확도 평가 없음

### 2.4 비교: Anthropic skill-creator

Anthropic blog "Improving Skill Creator": Anthropic 자체 도구는:
- sample prompts × expected outcomes 정의
- pass rate / elapsed time / token usage 측정
- A/B comparator agent로 두 버전 블라인드 평가
- description optimization (false-trigger 최소화)

→ moai-adk-go에는 이 인프라 부재.

---

## 3. 격차 분석

| 영역 | As-Is | To-Be | 격차 |
|------|-------|-------|------|
| Eval test framework | 없음 | sample-prompts.yaml + expected-outcomes.yaml | 신규 인프라 |
| Pass rate measurement | 없음 | per-skill metric | 신규 measurement |
| Elapsed time / token usage | 없음 | per-test metric | 신규 |
| A/B comparator agent | 없음 | 블라인드 평가 | 신규 |
| Description optimization | 없음 | trigger accuracy 분석 + 제안 | 신규 |
| Test result persistence | 없음 | `.moai/research/skill-tests/` | 신규 |

---

## 4. 코드베이스 분석 (Affected Files)

### 4.1 Primary 수정 대상

| 파일 | 수정 유형 | 변경 사유 |
|------|----------|----------|
| `.claude/agents/moai/builder-skill.md` | 확장 | testing framework 통합 |
| 신규: `.claude/skills/moai/<skill>/tests/sample-prompts.yaml` | 신규 (per-skill) | 각 skill의 sample prompts |
| 신규: `.claude/skills/moai/<skill>/tests/expected-outcomes.yaml` | 신규 (per-skill) | 각 skill의 expected outcomes |
| 신규: `.moai/research/skill-tests/<skill-id>/` | 신규 디렉토리 | 테스트 결과 보존 |
| 신규: builder-skill 본문에 A/B comparator 절 | 신규 | 두 버전 블라인드 평가 |

### 4.2 Templates (Template-First)

- `internal/template/templates/.claude/agents/moai/builder-skill.md`: 동일 변경
- 신규 테스트 디렉토리는 per-skill이므로 template은 example skill에만 적용 (skill 작성자가 따라 만듦)

### 4.3 Skill Test Schema (제안)

```yaml
# tests/sample-prompts.yaml (per skill)
skill_id: moai-foundation-core
prompts:
  - id: P001
    text: "MoAI의 TRUST 5 framework 설명해줘"
    intent: explain_concept
  - id: P002
    text: "/moai run SPEC-001 어떻게 실행해?"
    intent: workflow_command
  ...

# tests/expected-outcomes.yaml
skill_id: moai-foundation-core
outcomes:
  - prompt_id: P001
    skill_should_trigger: true
    expected_keywords: ["TRUST 5", "Tested", "Readable"]
    expected_modules_loaded: ["modules/trust-5-framework.md"]
  - prompt_id: P002
    skill_should_trigger: false  # /moai run is workflow, not foundation
    expected_redirect: "moai-workflow-spec or moai-foundation-core based on context"
  ...
```

---

## 5. 위험 및 가정

### 5.1 Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| 100+ skill을 모두 테스트하면 토큰 비용 폭증 | Critical | Critical | 변경된 skill만 incremental test, 전체 baseline은 1회만 |
| sample prompts 작성 부담 → 채택률 저조 | High | High | builder-skill에 prompts 자동 생성 시도 (LLM-driven) |
| expected outcomes의 객관성 부족 | Medium | High | keyword + module load + boolean trigger 결합 |
| A/B comparator의 블라인드성 보장 어려움 | Medium | Medium | 두 버전을 random shuffle, version label 제거 |
| Description optimization이 frontmatter를 자동 수정 → 의도치 않은 변경 | High | Medium | 자동 적용 금지, 제안만, 사용자 승인 필요 |

### 5.2 Assumptions

- A1: skill testing은 incremental (변경된 skill만)으로 충분
- A2: sample prompts × expected outcomes로 trigger accuracy 측정 가능
- A3: A/B comparator는 general-purpose Agent()로 구현 가능
- A4: description optimization은 자동 적용 금지, 사용자 승인 필요
- A5: 100+ skill 일괄 테스트는 baseline 측정 시 1회만 (이후 incremental)

---

## 6. 측정 계획

| Metric | 측정 방법 | 목표 |
|--------|----------|------|
| Skill test framework 채택률 (변경된 skill 중) | (with tests) / (modified skills) | > 50% in 첫 quarter, 100% mandatory in 다음 |
| Pass rate (per skill) | passing tests / total tests | >= 80% (passing skill 기준) |
| Elapsed time per test | per-test measurement | < 30 seconds avg |
| Token consumption per test | per-test measurement | < 5,000 tokens avg |
| Description optimization accuracy | manual review of 10 suggestions | > 70% accepted |
| False-trigger rate (before vs after optimization) | controlled test | -30% reduction |

---

## 7. 대안 검토

| 대안 | 채택 | 이유 |
|------|-----|------|
| 모든 skill 일괄 테스트 (100+ skill 매 변경마다) | ❌ | 토큰 비용 폭증, 무한 루프 |
| Test 자체 skip (manual review만) | ❌ | Anthropic 권고 핵심 어김 |
| Incremental test (변경된 skill만) + baseline 1회 | ✅ | 비용/효과 balance |
| External test runner (Go binary) | ❌ | LLM 평가는 LLM이 해야 (외부 binary로는 불가) |
| Test results를 git commit에 포함 | ❌ | local-only `.moai/research/skill-tests/` (gitignored) |

---

## 8. 참고 SPEC

- SPEC-AGENT-001 / SPEC-AGENT-002: 에이전트 신설
- SPEC-EVAL-RUBRIC-001 (Wave 2 sibling): rubric anchoring (test outcome rubric과 유사)
- SPEC-EVAL-LOOP-001: feedback-loop (skill optimization도 iterative)

---

## 9. Open Questions

- OQ1: Test 의무화 범위 (신규 skill만? 변경 skill 모두? 100+ legacy skill까지?) → plan에서 결정
- OQ2: A/B comparator의 input은 하나의 prompt인가, prompt set 전체인가?
- OQ3: Description optimization 자동 적용 vs 제안만? → 제안만 (사용자 승인) 권장
- OQ4: Pass rate threshold 80%는 모든 skill에 일괄? skill complexity별 차별?

---

End of research.md (SPEC-SKILL-TEST-001).
