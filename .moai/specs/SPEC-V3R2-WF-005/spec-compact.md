# SPEC-V3R2-WF-005 Compact Reference

> Auto-extract from `spec.md` v0.2.0 — REQs + ACs + Files + Exclusions only.
> Use this file for fast plan-auditor scans and cross-SPEC referencing.

---

## Requirements (EARS)

### Ubiquitous Requirements

- **REQ-WF005-001**: Language-specific guidance for the 16 supported languages **shall** live in `.claude/rules/moai/languages/*.md` as rules, not as skills.
- **REQ-WF005-002**: The v3 skill tree **shall not** contain any skill directory matching `moai-lang-*`.
- **REQ-WF005-003**: `.claude/rules/moai/development/skill-authoring.md` **shall** include a "language guidance lives in rules" principle that forbids language-scoped skills.
- **REQ-WF005-004**: All 16 language rules **shall** continue to use `paths:` frontmatter for conditional loading (per R6 §4.2 confirmation of current state).
- **REQ-WF005-005**: References to non-existent skills (`moai-lang-*`, `moai-infra-docker`, `moai-essentials-debug`, `moai-quality-testing`, `moai-quality-security`) **shall** be removed from all skill `related-skills` frontmatter and body text.

### Event-Driven Requirements

- **REQ-WF005-006**: **When** a user's project is detected as Python, the language rule `.claude/rules/moai/languages/python.md` **shall** auto-load via `paths: "**/*.py,**/pyproject.toml"` match.
- **REQ-WF005-007**: **When** any PR introduces a `moai-lang-<name>/` skill directory, CI **shall** reject it with `LANG_AS_SKILL_FORBIDDEN`.
- **REQ-WF005-008**: **When** agent prompts or skill bodies reference `moai-lang-<name>`, the refactor commit **shall** replace the reference with a direct path reference to `.claude/rules/moai/languages/<name>.md`.

### State-Driven Requirements

- **REQ-WF005-009**: **While** the v3 language count is 16 (per CLAUDE.local.md §15), adding a 17th language **shall** create a new `.md` file under `.claude/rules/moai/languages/` with `paths:` frontmatter — never a new skill.
- **REQ-WF005-010**: **While** Dart/Flutter is referenced, the canonical name **shall** be "flutter" (CLAUDE.local.md §15), both for rule filename and for project_markers detection.

### Optional Requirements

- **REQ-WF005-011**: **Where** cross-language abstraction is genuinely needed (e.g., general API design), the guidance **shall** live in `moai-ref-*` skills (`moai-ref-api-patterns`, `moai-ref-owasp-checklist`) — not as `moai-lang-*` composite.
- **REQ-WF005-012**: **Where** future architectural research proves language-as-skill more effective, reversal **shall** require a new SPEC with migration plan covering all 16 languages atomically (no partial adoption).

### Complex Requirements (Unwanted Behavior / Composite)

- **REQ-WF005-013** (Unwanted Behavior): **If** a skill body (not frontmatter) mentions `moai-lang-<name>` as a sibling skill, **then** the content audit **shall** flag it with `DEAD_LANG_SKILL_REFERENCE` and propose the rule-path substitute.
- **REQ-WF005-014** (Unwanted Behavior): **If** the 16-language neutrality is violated (e.g., a skill body marks only "go" as "primary"), **then** CI **shall** reject with `LANG_NEUTRALITY_VIOLATION` per CLAUDE.local.md §15.
- **REQ-WF005-015** (Complex: State + Event): **While** a skill references a non-existent sibling (`moai-infra-docker`, `moai-essentials-debug`, `moai-quality-*`), **when** the cleanup commit runs, the system **shall** substitute:
  - `moai-infra-docker` → remove (no substitute; platform infra deferred)
  - `moai-essentials-debug` → replace with `expert-debug` agent delegation note
  - `moai-quality-testing` → replace with `moai-foundation-quality` + `moai-ref-testing-pyramid`
  - `moai-quality-security` → replace with `moai-foundation-quality` + `moai-ref-owasp-checklist`

---

## Acceptance Criteria

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

## Files to Modify

### To-be-created (1 file)

- `internal/template/lang_boundary_audit_test.go` (CI guard for REQ-WF005-002, 013, 014)

### To-be-modified (~28 source-of-truth files + ~28 template mirrors)

#### Principle section (1 file)

- `.claude/rules/moai/development/skill-authoring.md` (append "## Language Guidance Lives in Rules, Not Skills" section)

#### Frontmatter cleanup (3 SKILL.md files)

- `.claude/skills/moai-platform-auth/SKILL.md` (line 20: `related-skills` field)
- `.claude/skills/moai-framework-electron/SKILL.md` (line 19: `related-skills` field)
- `.claude/skills/moai-platform-chrome-extension/SKILL.md` (line 19: `related-skills` field)

#### Body cleanup of `moai-lang-*` references (~17 files)

- `.claude/skills/moai/workflows/run.md` (lines 338-352, 979)
- `.claude/skills/moai-foundation-core/modules/agents-reference.md` (lines 268, 278-282, 292)
- `.claude/skills/moai-framework-electron/SKILL.md` (lines 228, 230)
- `.claude/skills/moai-platform-chrome-extension/SKILL.md` (lines 278, 279)
- `.claude/skills/moai-platform-auth/SKILL.md` (line 225)
- `.claude/skills/moai-platform-deployment/SKILL.md` (lines 398, 399)
- `.claude/skills/moai-workflow-loop/SKILL.md` (lines 148, 149)
- `.claude/skills/moai-domain-frontend/SKILL.md` (line 119)
- `.claude/skills/moai-foundation-cc/reference/skill-examples.md` (lines 237, 427)
- `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md` (lines 201, 392)
- `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md` (lines 30, 439, 936)
- `.claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md` (line 113)
- `.claude/skills/moai-foundation-cc/references/examples.md` (line 175)
- `.claude/rules/moai/development/agent-authoring.md` (line 165)
- `.claude/rules/moai/languages/scala.md` (line 131)
- `.claude/rules/moai/languages/flutter.md` (lines 94, 95)
- `.claude/rules/moai/languages/cpp.md` (line 100)

#### Other dead-skill-ID cleanup (~13 files)

- `moai-essentials-debug` substitution: `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:621`; `.claude/rules/moai/languages/{r,typescript,flutter,cpp,csharp,elixir,javascript,ruby}.md` (8 files)
- `moai-quality-testing` substitution: `.claude/rules/moai/languages/kotlin.md:109`
- `moai-quality-security` substitution: `.claude/skills/moai-domain-backend/SKILL.md:110`; `.claude/skills/moai-workflow-project/references/overview.md:120`; `.claude/rules/moai/languages/flutter.md:97`
- `moai-infra-docker` removal: `.claude/rules/moai/languages/kotlin.md:110`; `.claude/rules/moai/languages/java.md:121`

#### Trackable

- `CHANGELOG.md` (`## [Unreleased]` section)

#### Embedded template parity

- All edits mirrored to `internal/template/templates/.claude/...` (per `CLAUDE.local.md` §2 Template-First Rule HARD constraint)

---

## Exclusions (Out of Scope — What NOT to Build)

Per `spec.md` §1.2 Non-Goals + §2.2 Out of Scope:

1. 16개 언어 rule의 내용 rewrite (16 language rules' body content is preserved verbatim)
2. `moai-lang-*` skills 신규 생성 (**금지** — explicitly forbidden by REQ-WF005-002)
3. `.claude/rules/moai/languages/` 디렉터리 구조 변경 (no directory restructuring)
4. 특정 언어의 rule을 특별 대우 (no language gets primacy; CLAUDE.local.md §15 16-language neutrality preserved)
5. 언어 자동 감지 로직 수정 (existing `project.yaml` + project_markers reused)
6. 비-언어 rules(core/development/workflow/design)의 classification 변경 (only languages classified)
7. `paths:` frontmatter 방식 변경 (current mechanism preserved per REQ-WF005-004)
8. 신규 언어 추가 (16개 고정; 17th-language gate is REQ-WF005-009)
9. Dart/Flutter 이름 변경 (canonical "flutter" preserved per REQ-WF005-010)
10. 언어 rule에서 skill body style(Quick Reference / Implementation Guide) 강제 (rules retain rule format, not skill format)
11. Agency-absorbed skills의 body에 언어별 section 강제 추가 (no per-language sections forced into skill bodies)

---

End of spec-compact.md.
