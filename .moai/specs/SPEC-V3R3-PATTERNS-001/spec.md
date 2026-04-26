---
id: SPEC-V3R3-PATTERNS-001
version: "0.1.0"
status: draft
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: High
labels: [patterns, harness, agent-design, skill-design, apache2, cookbook, v3r3]
issue_number: null
title: "Pattern Cookbook — revfactory/harness Apache 2.0 6 reference docs 흡수"
breaking: false
bc_id: []
lifecycle: spec-anchored
related_specs: [SPEC-V3R3-DEF-007, SPEC-V3R3-ARCH-003]
---

# SPEC-V3R3-PATTERNS-001: Pattern Cookbook — revfactory/harness 흡수

## HISTORY

| Version | Date       | Author       | Description                                                                                       |
|---------|------------|--------------|---------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial draft. revfactory/harness (Apache 2.0) 6 reference docs를 6 rule 파일로 흡수. handoff §4.1. |

---

## 1. Goal (목적)

revfactory/harness 저장소(Apache 2.0 라이선스)의 6 reference docs를 MoAI-ADK의 `.claude/rules/moai/`로 흡수해 **agent 설계 / skill 작성 / 팀 운영 / QA 경계면 / orchestrator 선택**을 위한 권위 있는 cookbook을 확립한다.

본 SPEC은 외부 prior art를 정식 채택하는 작업이며 모든 산출 파일 상단에 Apache 2.0 attribution 의무를 명시한다.

### 1.1 배경

- handoff §4.1: revfactory/harness `skills/harness/references/*` 6 reference docs는 MoAI-ADK가 v3.0.0 R3에서 도입하려는 패턴 체계와 정확히 일치한다. 자력 작성보다 prior art 흡수가 품질·신뢰성·시간 모두에서 우월.
- 6 reference 항목 (handoff 명시):
  1. **agent-patterns** — Team / Sub-agent / Hybrid / Orchestrator / Specialist / Pipeline 6 architectural patterns
  2. **boundary-verification** — QA 경계면 교차 검증 + 7 실제 버그 사례
  3. **skill-ab-testing** — With-skill vs Baseline A/B methodology
  4. **team-pattern-cookbook** — 5 팀 운영 예시 (research/implementation/review/design/debug)
  5. **orchestrator-templates** — 3 templates (Team / Sub / Hybrid)
  6. **skill-writing-craft** — description craft + body structure + schema validation
- License: Apache 2.0. Attribution + NOTICE 의무. MoAI-ADK 자체 라이선스(MIT)와 호환 — Apache 2.0 → MIT 방향은 Apache 2.0 조건(notice 보존) 충족 시 허용.

### 1.2 비목표 (Non-Goals)

- 기존 MoAI-ADK rule 파일 수정/대체 (이번 SPEC은 추가만, 기존 frozen content 변경 없음)
- harness 저장소의 skill / agent 파일 자체를 흡수 (reference docs만 대상)
- 한국어 번역 추가 (cookbook은 영어 원본 + 발췌. 사용자 대화는 ko, agent prompt는 en 정책 유지)
- Apache 2.0 외 라이선스 코드 흡수 (이번 SPEC은 Apache 2.0 단일 source)

---

## 2. Scope

### 2.1 산출 파일 (6 rule files + 1 NOTICE)

#### 신규 6 rule 파일

| # | 경로 | 흡수 source | 핵심 |
|---|------|-------------|------|
| 1 | `.claude/rules/moai/development/agent-patterns.md` | harness `references/agent-patterns.md` | 6 patterns + 안티패턴 + selection guide |
| 2 | `.claude/rules/moai/quality/boundary-verification.md` | harness `references/boundary-verification.md` | 경계면 검증 방법론 + 7 bug case |
| 3 | `.claude/rules/moai/development/skill-ab-testing.md` | harness `references/skill-ab-testing.md` | A/B methodology + sample size + significance |
| 4 | `.claude/rules/moai/workflow/team-pattern-cookbook.md` | harness `references/team-pattern-cookbook.md` | 5 팀 예시 + role profile + ownership |
| 5 | `.claude/rules/moai/development/orchestrator-templates.md` | harness `references/orchestrator-templates.md` | 3 templates + when/how/recovery |
| 6 | `.claude/rules/moai/development/skill-writing-craft.md` | harness `references/skill-writing-craft.md` | description craft + 3-level disclosure + schema |

신규 디렉토리 1개: `.claude/rules/moai/quality/` (Pattern #2 산출 위해 생성)

#### Apache 2.0 attribution 산출

- `.claude/rules/moai/NOTICE.md` (또는 `LICENSE-third-party.md`) — 흡수된 source 일람, Apache 2.0 텍스트 사본, attribution 보존

### 2.2 Template-First 동기화

- Local: `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/`
- Template mirror: `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/rules/moai/`
- 양쪽 동시 갱신 필수. `make build` 후 `go test ./internal/template/...` 회귀 확인.

### 2.3 Frontmatter 정책

각 rule 파일은 conditional loading 가능하도록 frontmatter 추가:

- `agent-patterns.md` → `paths: ".claude/agents/**/*.md,.claude/rules/moai/development/agent-authoring.md"` (agent 작성 시 auto-load)
- `boundary-verification.md` → `paths: "**/*_test.go,**/test/**"` (테스트 작성 시 auto-load)
- `skill-ab-testing.md` → `paths: ".claude/skills/**/*.md"` (skill 작성/평가 시 auto-load)
- `team-pattern-cookbook.md` → `paths: "workflow.yaml,.claude/rules/moai/workflow/team-protocol.md"` (team 모드 시 auto-load)
- `orchestrator-templates.md` → `paths: ".claude/rules/moai/core/moai-constitution.md,CLAUDE.md"` (오케스트레이터 컨텍스트)
- `skill-writing-craft.md` → `paths: ".claude/skills/**/SKILL.md"` (skill body 작성 시 auto-load)

---

## 3. EARS Requirements

### REQ-PAT-001 (Ubiquitous — agent patterns)

WHEN agent designer references pattern selection during agent authoring,
THE `.claude/rules/moai/development/agent-patterns.md` SHALL provide 6 architectural patterns (Team / Sub-agent / Hybrid / Orchestrator / Specialist / Pipeline) with: pattern definition, when-to-use criteria, anti-pattern examples, and Apache 2.0 attribution at file head.

### REQ-PAT-002 (Ubiquitous — boundary verification)

WHEN QA engineer plans boundary tests OR test agent receives `boundary-verification.md` via paths frontmatter,
THE document SHALL describe 7 documented bug case patterns with: bug symptom, root cause, boundary condition that should have been tested, and verification strategy template.

### REQ-PAT-003 (Ubiquitous — skill A/B testing)

WHEN skill author optimizes skill quality OR proposes a new skill,
THE `.claude/rules/moai/development/skill-ab-testing.md` SHALL provide A/B methodology including: with-skill vs baseline comparison protocol, evaluation criteria, statistical significance guidance, and sample size recommendations.

### REQ-PAT-004 (Ubiquitous — team patterns)

WHEN team mode setup is needed OR `workflow.yaml` is being authored,
THE `.claude/rules/moai/workflow/team-pattern-cookbook.md` SHALL describe 5 team examples (research / implementation / review / design / debug) with: role profile composition, file ownership pattern, communication protocol, and shutdown sequence.

### REQ-PAT-005 (Ubiquitous — orchestrator templates)

WHEN orchestrator template is selected during workflow design,
THE `.claude/rules/moai/development/orchestrator-templates.md` SHALL provide 3 templates (Team-orchestrator / Sub-orchestrator / Hybrid-orchestrator) with: when to use each, how to spawn, error recovery flow, and escalation rules.

### REQ-PAT-006 (Ubiquitous — skill writing craft)

WHEN skill body is authored OR skill description is reviewed for triggering accuracy,
THE `.claude/rules/moai/development/skill-writing-craft.md` SHALL guide on: description craft (when to trigger / when to skip), 3-level progressive disclosure body structure, frontmatter schema validation rules.

### REQ-PAT-007 (Ubiquitous — Apache 2.0 compliance)

WHILE any of the 6 cookbook rule files are present in repository,
THE NOTICE artifact (`.claude/rules/moai/NOTICE.md`) SHALL list all 6 source references with: original repository URL (revfactory/harness), license (Apache 2.0), attribution clause, and date of import.

### REQ-PAT-008 (Ubiquitous — Template-First sync)

WHILE the 6 cookbook rule files exist in `.claude/rules/moai/`,
THE corresponding files SHALL also exist (byte-identical) in `internal/template/templates/.claude/rules/moai/` so that `moai init` produces the same cookbook for new projects.

---

## 4. Exclusions (What NOT to Build)

- **언어 번역**: 영어 원문만 흡수. 한국어 번역 추가는 별도 SPEC.
- **harness 저장소의 skill / agent 본체 흡수**: 이번 SPEC은 reference docs 6개만. skill/agent 본체 흡수는 별도 SPEC.
- **MoAI-ADK 기존 패턴 문서 대체**: `.claude/rules/moai/development/agent-authoring.md`, `skill-authoring.md`은 그대로 유지. cookbook은 보충 자료.
- **harness 저장소의 다른 reference (7번째 이후)**: handoff에서 명시한 6개만. 추가 흡수는 별도 SPEC.
- **harness clone 자동화**: 이번 SPEC은 1회성 흡수. 추후 upstream sync 자동화는 별도 SPEC.
- **frontmatter `paths` 글로브 외 동적 로딩**: 표준 Claude Code rules 메커니즘만 사용.

---

## 5. Constraints

- [HARD] Apache 2.0 attribution: 모든 6 산출 파일 상단에 `<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->` 주석.
- [HARD] Template-First: local + template mirror 동시 갱신, `make build` 후 회귀 테스트.
- [HARD] 16-language neutral: cookbook 예시는 특정 언어(Go/Python/JS) 편향 금지. patterns은 언어 중립적 표현.
- [HARD] 기존 frozen rule 파일 (CLAUDE.md, moai-constitution.md, agent-common-protocol.md, design/constitution.md) 수정 금지.
- [HARD] 영어 instruction 정책 (coding-standards.md §Language Policy) 준수.
- 산출물은 `.claude/rules/moai/` 하위. SKILL.md / agent body 직접 수정 없음.

---

## 6. Acceptance Hooks

- AC-001 ~ AC-005 → see `acceptance.md`
- Implementation order: see `plan.md`
- Task breakdown: see `tasks.md`

---

## 7. References

- handoff §4.1 — revfactory/harness reference adoption strategy
- `.claude/rules/moai/development/agent-authoring.md` — 기존 agent 작성 가이드 (보존)
- `.claude/rules/moai/development/skill-authoring.md` — 기존 skill 작성 가이드 (보존)
- `.claude/rules/moai/workflow/team-protocol.md` — 기존 team 운영 프로토콜 (보존)
- SPEC-V3R3-DEF-007 — Convention Compliance Sweep (related: skill frontmatter baseline)
- SPEC-V3R3-ARCH-003 — Expert tool uplift (related: agent body 일관성 baseline)
- Apache License 2.0 — https://www.apache.org/licenses/LICENSE-2.0
- revfactory/harness — https://github.com/revfactory/harness
