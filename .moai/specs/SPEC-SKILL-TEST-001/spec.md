---
id: SPEC-SKILL-TEST-001
status: draft
version: "0.1.0"
priority: High
labels: [skill, testing, framework, builder, evaluation, wave-2, tier-1]
issue_number: null
scope: [builder-skill.md, skill-tests-framework]
blockedBy: []
dependents: []
related_specs: [SPEC-EVAL-RUBRIC-001, SPEC-EVAL-LOOP-001]
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
wave: 2
tier: 1
---

# SPEC-SKILL-TEST-001: Skill Testing Framework

## HISTORY

- 2026-04-30 v0.1.0: 최초 작성. Wave 2 / Tier 1. Anthropic "Improving Skill Creator" 패턴 (eval tests + metrics + A/B comparator + description optimization)을 builder-skill agent에 통합.

---

## 1. Goal (목적)

Anthropic blog "Testing turns a skill that seems to work into one you know works"에 따라 builder-skill agent를 확장하여 skill testing framework를 도입한다. 각 skill에 sample prompts × expected outcomes 정의 + pass rate / elapsed time / token usage 측정 + A/B comparator 블라인드 평가 + description optimization 제안 메커니즘을 구현한다.

### 1.1 배경

- Anthropic: "Without testing, every skill is hopeful documentation."
- Anthropic: "Descriptions matter more than skill body content for trigger accuracy."
- 현재 moai-adk-go 100+ skill 중 어떤 것도 effectiveness 측정 안 됨
- builder-skill agent는 생성/수정만 담당, testing 책임 부재

### 1.2 비목표 (Non-Goals)

- 100+ legacy skill 일괄 테스트 (baseline 측정 1회 + 점진 마이그레이션)
- skill body 자동 재작성 (testing은 평가/제안만)
- description 자동 수정 적용 (제안만, 사용자 승인 필수)
- External Go binary로 testing 구현 (LLM 평가는 LLM이 수행)
- skill registry/marketplace 신설
- non-MoAI skill (외부 skill) 테스트 통합

---

## 2. Scope (범위)

### 2.1 In Scope

- builder-skill agent에 testing framework 통합:
  - skill 생성/수정 시 sample prompts 작성 prompt
  - eval test 실행 명령
  - A/B comparator 호출
  - description optimization 분석 + 제안
- Skill test directory schema:
  - `.claude/skills/<skill>/tests/sample-prompts.yaml`
  - `.claude/skills/<skill>/tests/expected-outcomes.yaml`
- Test result persistence: `.moai/research/skill-tests/<skill-id>/<run-id>.md`
- builder-skill 본문에 testing protocol 절 신설
- Template-First 동기화

### 2.2 Exclusions (What NOT to Build)

- 100+ legacy skill 일괄 테스트 (baseline 1회 + incremental만)
- skill body 자동 재작성 (manual edit)
- description 자동 수정 적용 (제안만)
- External binary test runner
- skill marketplace
- non-MoAI external skill 테스트
- Hook-level enforcement (testing은 LLM 책임)
- skill description 변경에 대한 automatic deployment

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- Claude Code v2.1.111+
- 영향 디렉터리: `.claude/agents/moai/`, `.claude/skills/`, `.moai/research/skill-tests/`
- 템플릿 동기화: `internal/template/templates/`
- gitignore: `.moai/research/skill-tests/` (local-only)

---

## 4. Assumptions (가정)

- A1: skill testing은 incremental (변경된 skill만) 충분
- A2: sample prompts (5개 권장) × expected outcomes로 pass rate 측정 가능
- A3: A/B comparator는 general-purpose Agent()로 블라인드 평가 가능
- A4: description optimization은 자동 적용 금지, 사용자 AskUserQuestion 승인 필수
- A5: 100+ legacy skill의 testing 마이그레이션은 점진적 (변경 시점에만 의무)

---

## 5. Requirements (EARS Format)

### 5.1 Ubiquitous Requirements

- **REQ-ST-001**: THE BUILDER-SKILL BODY SHALL contain a dedicated section titled "Skill Testing Framework" describing the test schema, metrics, and execution protocol.
- **REQ-ST-002**: EACH SKILL DIRECTORY THAT OPTS INTO TESTING SHALL contain a `tests/` subfolder with two YAML files: `sample-prompts.yaml` and `expected-outcomes.yaml`.
- **REQ-ST-003**: THE TEST METRICS SHALL include `pass_rate`, `elapsed_time_avg`, `token_consumption_avg` for each test run.
- **REQ-ST-004**: THE TEST RESULTS SHALL be persisted to `.moai/research/skill-tests/<skill-id>/<run-id>.md`.
- **REQ-ST-005**: THE A/B COMPARATOR AGENT SHALL judge two skill versions in blind mode (no version labels exposed in the comparator prompt).
- **REQ-ST-006**: THE DESCRIPTION OPTIMIZATION SHALL produce suggestions only; automatic application is prohibited.

### 5.2 Event-Driven Requirements

- **REQ-ST-007**: WHEN builder-skill creates a new skill, THE BUILDER SHALL prompt the user (via orchestrator AskUserQuestion) to provide at least 3 sample test prompts.
- **REQ-ST-008**: WHEN builder-skill modifies an existing skill, THE BUILDER SHALL check for an existing `tests/` subfolder; if absent, prompt the user to create one (via orchestrator AskUserQuestion).
- **REQ-ST-009**: WHEN builder-skill executes a skill test, THE METRICS (pass_rate, elapsed_time, token_consumption) SHALL be recorded in the test result file.
- **REQ-ST-010**: WHEN two skill versions exist (current `v_current` vs proposed `v_proposed`), THE BUILDER SHALL spawn an A/B comparator agent (general-purpose) with both outputs presented in randomized order without version labels.
- **REQ-ST-011**: WHEN description optimization analysis runs, THE BUILDER SHALL emit suggestions to the orchestrator who SHALL surface them via AskUserQuestion for user approval before any frontmatter modification.

### 5.3 State-Driven Requirements

- **REQ-ST-012**: WHILE skill is in test phase, THE TEST RESULTS SHALL be persisted incrementally so partial runs can be resumed.
- **REQ-ST-013**: WHILE A/B comparator is evaluating, THE COMPARATOR SHALL NOT have access to skill version metadata (file paths, frontmatter, modification times).

### 5.4 Conditional Requirements

- **REQ-ST-014**: WHERE skill description matches less than 70% of sample prompts (trigger keyword overlap heuristic), THEN BUILDER-SKILL SHALL emit a description optimization suggestion.
- **REQ-ST-015**: IF the test result pass_rate is less than 80%, THEN BUILDER-SKILL SHALL flag the skill as `needs_refinement` in the result metadata.
- **REQ-ST-016**: WHERE the same skill has been tested multiple times, THE LATEST test run SHALL be used for current pass_rate; historical runs are retained for trend analysis.
- **REQ-ST-017**: IF the user does NOT provide sample prompts when prompted, THEN THE BUILDER-SKILL SHALL emit a non-blocking warning and proceed with the skill creation/modification (testing remains opt-in initially).
- **REQ-ST-018**: WHERE A/B comparator returns inconclusive (`winner: null`), THEN THE BUILDER-SKILL SHALL retain the current version and surface the inconclusive result to the orchestrator.
- **REQ-ST-019**: WHERE the skill has external dependencies (other skills, MCP tools), THE TESTING SHALL include availability checks before scoring.

### 5.5 Unwanted (Negative) Requirements

- **REQ-ST-020**: THE BUILDER-SKILL SHALL NOT automatically modify skill frontmatter (description, triggers, name) without explicit user approval.
- **REQ-ST-021**: THE A/B COMPARATOR SHALL NOT receive any version label, file path, or modification metadata that would break blind evaluation.
- **REQ-ST-022**: THE TESTING FRAMEWORK SHALL NOT auto-test all 100+ legacy skills in a single run; testing is incremental and opt-in for legacy skills.
- **REQ-ST-023**: THE TEST RESULTS DIRECTORY (`.moai/research/skill-tests/`) SHALL be gitignored to prevent local test artifacts from polluting commits.

---

## 6. Success Criteria

| Criterion | Measurement | Target |
|-----------|-------------|--------|
| Test framework adoption rate (modified skills) | (skills with tests) / (modified skills) | > 50% in first quarter |
| Pass rate threshold (passing skills) | per-skill | >= 80% |
| Elapsed time per test | per-test | < 30 seconds avg |
| Token consumption per test | per-test | < 5,000 tokens avg |
| A/B comparator blind compliance | prompt audit | 100% (no leakage) |
| Description optimization acceptance | manual review of 10 suggestions | > 70% accepted |
| False-trigger reduction | controlled test before vs after | -30% |
| Template-First sync | clean diff | `make build` |

---

## 7. Acceptance References

See `acceptance.md` for Given-When-Then scenarios and Definition of Done.

---

## 8. Constraints

- C1: skill body 자동 재작성 금지 (manual edit only)
- C2: description/frontmatter 자동 수정 금지 (제안만, AskUserQuestion 승인 필수)
- C3: A/B comparator의 블라인드성 보장 (version label 노출 금지)
- C4: legacy skill 일괄 테스트 금지 (incremental만)
- C5: test results는 gitignored (`.moai/research/skill-tests/`)
- C6: testing은 LLM 평가, external binary 사용 금지

End of spec.md (SPEC-SKILL-TEST-001 v0.1.0).
