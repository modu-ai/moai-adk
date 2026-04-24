---
id: SPEC-V3R2-WF-005
title: Language Rules vs Skills Boundary Codification
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P2 Medium
phase: "v3.0.0 — Phase 6 — Multi-Mode Workflow"
module: ".claude/rules/moai/languages/, .claude/skills/, .claude/rules/moai/development/skill-authoring.md"
dependencies:
  - SPEC-V3R2-WF-001
related_gap:
  - r4-skill-audit-lang-skills-absent
  - r6-frontmatter-consistency
related_theme: "Theme 6 — Workflow Consolidation"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "languages, rules, skills, boundary, paths-frontmatter, v3"
---

# SPEC-V3R2-WF-005: Language Rules vs Skills Boundary Codification

## HISTORY

| Version | Date       | Author | Description                                                          |
|---------|------------|--------|----------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — codify languages as rules (not skills) in v3           |

---

## 1. Goal (목적)

R4 audit는 "moai-lang-* skills referenced but absent"를 단일 최대 구조적 불일치로 지목한다: 16개 language rules(`go`, `python`, `typescript`, `javascript`, `rust`, `java`, `kotlin`, `csharp`, `ruby`, `php`, `elixir`, `cpp`, `scala`, `r`, `flutter`, `swift`)가 `.claude/rules/moai/languages/` 아래에 **rules**로 존재하지만, 여러 skill의 `related-skills` 필드와 문서는 `moai-lang-*` **skills**을 가리키고 있다. 본 SPEC은 **v3.0.0 결정: 언어별 가이드는 rules로 유지**를 명문화하고, `moai-lang-*` skill 창설을 거부하며, 모든 쟈랑 reference를 rules 경로로 수정한다.

### 1.1 배경

R4 §Section A: "16 language rules live in `.claude/rules/moai/languages/` instead. Either migrate rules to skills or document why not. This is the single biggest architectural inconsistency uncovered in this audit: skills reference siblings that do not exist at the skill tree but do exist as rules, blurring the skill/rule boundary." R6 §4.2: "All 16 language rules use `paths:` frontmatter for conditional loading. **Uniform and correct.** Keep." 언어 rules는 이미 `paths: "**/*.py,**/pyproject.toml"` 같은 path glob으로 자동 로드되어 rule 성격에 부합한다. Skills는 keyword-description 기반 auto-activation이 1차인 반면 rules는 파일 경로 매칭이 1차라는 본질적 차이가 있다. 본 SPEC은 이 차이를 architectural boundary로 공식화한다.

### 1.2 비목표 (Non-Goals)

- 16개 언어 rule의 내용 rewrite
- `moai-lang-*` skills 신규 생성 (**금지**)
- `.claude/rules/moai/languages/` 디렉터리 구조 변경
- 특정 언어의 rule을 특별 대우 (16개 언어 동등성 유지, CLAUDE.local.md §15)
- 언어 자동 감지 로직 수정 (기존 `project.yaml` + project_markers 재사용)
- 비-언어 rules(core/development/workflow/design)의 classification 변경

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `.claude/rules/moai/languages/*.md` 16개 파일의 canonical location 선언, `moai-lang-*` skill 창설 금지 rule.
- 모든 skill의 `related-skills` 필드에서 `moai-lang-*` reference 제거 또는 rules path로 교체.
- `.claude/rules/moai/development/skill-authoring.md`에 "language guidance lives in rules, not skills" 원칙 추가.
- R4 §Section A coverage gaps에 언급된 기타 missing skills(`moai-infra-docker`, `moai-essentials-debug`, `moai-quality-testing`, `moai-quality-security`) 판정:
  - `moai-infra-docker`: 신규 skill 창설하지 않음 (언어 무관 infra는 향후 `moai-platform-*`으로 확장 고려)
  - `moai-essentials-debug`: `expert-debug` agent로 route (skill 불필요)
  - `moai-quality-testing`, `moai-quality-security`: `moai-foundation-quality` + `moai-ref-testing-pyramid` + `moai-ref-owasp-checklist`로 redirect

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 16개 언어 rule의 body rewrite
- `paths:` frontmatter 방식 변경 (현재 방식 유지)
- 신규 언어 추가 (16개 고정)
- Dart/Flutter 이름 변경 (canonical name "flutter", CLAUDE.local.md §15 준수)
- 언어 rule에서 skill body style(Quick Reference / Implementation Guide) 강제
- Agency-absorbed skills의 body에 언어별 section 강제 추가

---

## 3. Environment (환경)

- 런타임: Claude Code rule loader (paths-based), skill loader (keyword-based)
- 영향 디렉터리:
  - 수정: `.claude/rules/moai/development/skill-authoring.md` (language-boundary principle 추가)
  - 수정: `.claude/skills/<any>/SKILL.md` frontmatter `related-skills` 필드 (lang reference 제거)
  - 참조: `.claude/rules/moai/languages/*.md` (16개)
- 외부 레퍼런스: R4 §Section A, R6 §4.2 (paths frontmatter uniform), CLAUDE.local.md §15 (16-language neutrality)

---

## 4. Assumptions (가정)

- 16개 언어 rule은 이미 `paths:` frontmatter로 자동 로드되어 runtime behavior는 변경되지 않는다.
- Skill loader와 rule loader는 별개 메커니즘 (keyword vs paths)이며 v3에서 동일하게 유지된다.
- `related-skills` frontmatter 필드의 lang-skill reference 제거는 skill auto-activation에 영향 없다 (reference-less).
- CLAUDE.local.md §15의 16-language neutrality는 본 SPEC 결정에 의해 영향받지 않는다.
- `moai-infra-docker` 등 존재하지 않는 referenced skill은 dead reference로서 R4 audit가 이미 식별했다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-WF005-001**
Language-specific guidance for the 16 supported languages **shall** live in `.claude/rules/moai/languages/*.md` as rules, not as skills.

**REQ-WF005-002**
The v3 skill tree **shall not** contain any skill directory matching `moai-lang-*`.

**REQ-WF005-003**
`.claude/rules/moai/development/skill-authoring.md` **shall** include a "language guidance lives in rules" principle that forbids language-scoped skills.

**REQ-WF005-004**
All 16 language rules **shall** continue to use `paths:` frontmatter for conditional loading (per R6 §4.2 confirmation of current state).

**REQ-WF005-005**
References to non-existent skills (`moai-lang-*`, `moai-infra-docker`, `moai-essentials-debug`, `moai-quality-testing`, `moai-quality-security`) **shall** be removed from all skill `related-skills` frontmatter and body text.

### 5.2 Event-Driven Requirements

**REQ-WF005-006**
**When** a user's project is detected as Python, the language rule `.claude/rules/moai/languages/python.md` **shall** auto-load via `paths: "**/*.py,**/pyproject.toml"` match.

**REQ-WF005-007**
**When** any PR introduces a `moai-lang-<name>/` skill directory, CI **shall** reject it with `LANG_AS_SKILL_FORBIDDEN`.

**REQ-WF005-008**
**When** agent prompts or skill bodies reference `moai-lang-<name>`, the refactor commit **shall** replace the reference with a direct path reference to `.claude/rules/moai/languages/<name>.md`.

### 5.3 State-Driven Requirements

**REQ-WF005-009**
**While** the v3 language count is 16 (per CLAUDE.local.md §15), adding a 17th language **shall** create a new `.md` file under `.claude/rules/moai/languages/` with `paths:` frontmatter — never a new skill.

**REQ-WF005-010**
**While** Dart/Flutter is referenced, the canonical name **shall** be "flutter" (CLAUDE.local.md §15), both for rule filename and for project_markers detection.

### 5.4 Optional Requirements

**REQ-WF005-011**
**Where** cross-language abstraction is genuinely needed (e.g., general API design), the guidance **shall** live in `moai-ref-*` skills (`moai-ref-api-patterns`, `moai-ref-owasp-checklist`) — not as `moai-lang-*` composite.

**REQ-WF005-012**
**Where** future architectural research proves language-as-skill more effective, reversal **shall** require a new SPEC with migration plan covering all 16 languages atomically (no partial adoption).

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-WF005-013 (Unwanted Behavior)**
**If** a skill body (not frontmatter) mentions `moai-lang-<name>` as a sibling skill, **then** the content audit **shall** flag it with `DEAD_LANG_SKILL_REFERENCE` and propose the rule-path substitute.

**REQ-WF005-014 (Unwanted Behavior)**
**If** the 16-language neutrality is violated (e.g., a skill body marks only "go" as "primary"), **then** CI **shall** reject with `LANG_NEUTRALITY_VIOLATION` per CLAUDE.local.md §15.

**REQ-WF005-015 (Complex: State + Event)**
**While** a skill references a non-existent sibling (`moai-infra-docker`, `moai-essentials-debug`, `moai-quality-*`), **when** the cleanup commit runs, the system **shall** substitute:
- `moai-infra-docker` → remove (no substitute; platform infra deferred)
- `moai-essentials-debug` → replace with `expert-debug` agent delegation note
- `moai-quality-testing` → replace with `moai-foundation-quality` + `moai-ref-testing-pyramid`
- `moai-quality-security` → replace with `moai-foundation-quality` + `moai-ref-owasp-checklist`

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-WF005-01**: Given the v3 skill tree When inspected Then no directory matches `moai-lang-*` (maps REQ-WF005-002).
- **AC-WF005-02**: Given `.claude/rules/moai/development/skill-authoring.md` When inspected Then "language guidance lives in rules" principle section is present (maps REQ-WF005-003).
- **AC-WF005-03**: Given all 16 language rules When frontmatter is parsed Then each has `paths:` declaration (maps REQ-WF005-004).
- **AC-WF005-04**: Given a Python project When Claude Code session starts Then `.claude/rules/moai/languages/python.md` auto-loads (maps REQ-WF005-006).
- **AC-WF005-05**: Given a PR adding `.claude/skills/moai-lang-rust/` When CI runs Then `LANG_AS_SKILL_FORBIDDEN` rejection (maps REQ-WF005-007).
- **AC-WF005-06**: Given a skill's `related-skills` listing `moai-lang-typescript` When refactor commit runs Then the entry is removed (maps REQ-WF005-005, REQ-WF005-008).
- **AC-WF005-07**: Given a skill body referencing `moai-quality-testing` When audit runs Then reference is replaced with `moai-foundation-quality` + `moai-ref-testing-pyramid` per REQ-WF005-015.
- **AC-WF005-08**: Given a skill body referencing `moai-essentials-debug` When audit runs Then reference is replaced with "delegate to expert-debug agent" note (maps REQ-WF005-015).
- **AC-WF005-09**: Given a PR adding a 17th language as skill When CI runs Then `LANG_AS_SKILL_FORBIDDEN` rejection (maps REQ-WF005-009).
- **AC-WF005-10**: Given `.claude/rules/moai/languages/flutter.md` When opened Then filename is "flutter" (not "dart") (maps REQ-WF005-010).
- **AC-WF005-11**: Given a skill body claiming "go is primary; other languages are planned" When CI runs Then `LANG_NEUTRALITY_VIOLATION` fires (maps REQ-WF005-014).
- **AC-WF005-12**: Given a skill body mentioning `moai-lang-python` in prose When audit runs Then `DEAD_LANG_SKILL_REFERENCE` warning fires with rule-path suggestion (maps REQ-WF005-013).

---

## 7. Constraints (제약)

- 16-language neutrality: 어떤 언어도 "primary" 라벨 금지 (CLAUDE.local.md §15).
- `paths:` frontmatter 방식 유지 (R6 §4.2).
- 9-direct-dep 정책 준수.
- Canonical Dart name은 "flutter" (CLAUDE.local.md §15).
- 본 SPEC은 ``moai-lang-*`` reference 제거만 담당; 언어 rule 내용은 건드리지 않음.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| 사용자 기대: skill처럼 keyword 기반으로 언어 guidance 활성화 | UX 기대 차이 | paths-based auto-load는 rule의 원래 모델; 문서화로 기대 설정 |
| `related-skills` 필드에서 lang reference 제거 후 다른 skill activation 저해 | skill 매칭 저하 | lang은 rule이므로 skill activation과 무관; 실측으로 검증 |
| Future research가 lang-as-skill을 선호할 가능성 | 본 SPEC 번복 압력 | REQ-WF005-012의 atomic reversal 조건으로 gate |
| 17개 이상 언어 확장 시 rule 파일 폭증 | 유지보수 부담 | paths-based loader는 open-set 확장에 내성적 |
| 비-lang 존재하지 않는 skill reference(`moai-infra-docker` 등)가 dead ref로 방치 | 문서 혼란 | REQ-WF005-015의 명시적 substitute 규정 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-WF-001: 24-skill 카탈로그 확정 후에야 dead reference 제거가 가능.

### 9.2 Blocks

- 없음 (문서/rule 레벨 변경).

### 9.3 Related

- SPEC-V3R2-MIG-002: Hook cleanup과 함께 rule consolidation 연동 가능.
- R6 §4 rules audit; CLAUDE.local.md §15 language neutrality.

---

## 10. Traceability (추적성)

- REQ 총 15개: Ubiquitous 5, Event-Driven 3, State-Driven 2, Optional 2, Complex 3.
- AC 총 12개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: R4 §Section A coverage gaps; R6 §4.2 paths frontmatter; CLAUDE.local.md §15 16-language neutrality.
- BC 영향: 없음.
- 구현 경로 예상:
  - `.claude/rules/moai/development/skill-authoring.md` (language-boundary principle 추가)
  - 48개 skill의 `related-skills` 필드 audit + lang reference 제거
  - CI 체크 추가 (`LANG_AS_SKILL_FORBIDDEN`, `LANG_NEUTRALITY_VIOLATION`, `DEAD_LANG_SKILL_REFERENCE`)

---

End of SPEC.
