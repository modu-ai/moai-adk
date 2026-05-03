# SPEC-V3R2-WF-005 Implementation Plan (Phase 1B)

> Implementation plan for Language Rules vs Skills Boundary Codification.
> Companion to `spec.md` v0.2.0 and `research.md` v0.1.0.
> Authored against branch `feature/SPEC-V3R2-WF-005-language-rules-boundary` at `/Users/goos/MoAI/moai-adk-go` (solo mode, no worktree).

## HISTORY

| Version | Date       | Author                        | Description                                                              |
|---------|------------|-------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 1B) | Initial implementation plan per `.claude/skills/moai/workflows/plan.md` Phase 1B |

---

## 1. Plan Overview

### 1.1 Goal restatement

Codify the v3.0.0 architectural decision stated in `spec.md` REQ-WF005-001..002:

- **Language guidance lives in rules**, not skills (`.claude/rules/moai/languages/*.md` × 16).
- **`moai-lang-*` skill creation is forbidden** under any circumstance.
- All existing `related-skills` frontmatter and body references to `moai-lang-*` (and the four other dead skill IDs `moai-infra-docker`, `moai-essentials-debug`, `moai-quality-testing`, `moai-quality-security`) are removed or substituted per REQ-WF005-005, REQ-WF005-008, REQ-WF005-015.

This is a **declaration-level** change (research.md §8): the `moai-lang-*` skills do not exist today, so removing references is documentation-accuracy work, not a runtime break. The CI guards (REQ-WF005-007, REQ-WF005-013, REQ-WF005-014) are forward-looking regression preventers.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED**: Write failing audit tests (REQ-WF005-002 + REQ-WF005-013 + REQ-WF005-014 enforcement) before any skill body or rule file is touched. Confirm tests fail in CI (sentinel-not-yet-asserted state).
- **GREEN**: Add the new "Language Guidance Lives in Rules" principle section to `skill-authoring.md` (REQ-WF005-003), then sweep through the 17 + 13 = ~28 files identified in research.md §3-§4 and remove or substitute dead-skill-ID references (REQ-WF005-005, REQ-WF005-008, REQ-WF005-015), making the audit tests pass.
- **REFACTOR**: Consolidate the substitution patterns if they share boilerplate; cross-link from edited files to the new principle section.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| "Language Guidance Lives in Rules" section (new) | `.claude/rules/moai/development/skill-authoring.md` (append after "## Tool Permissions by Category") | REQ-WF005-001, REQ-WF005-002, REQ-WF005-003, REQ-WF005-004, REQ-WF005-009 |
| `LANG_AS_SKILL_FORBIDDEN` audit test (new file) | `internal/template/lang_boundary_audit_test.go` | REQ-WF005-002, REQ-WF005-007, REQ-WF005-009 |
| `DEAD_LANG_SKILL_REFERENCE` audit subtest | same file | REQ-WF005-005, REQ-WF005-013 |
| `LANG_NEUTRALITY_VIOLATION` audit subtest | same file | REQ-WF005-014 |
| Skill SKILL.md frontmatter cleanup (3 files) | `.claude/skills/{moai-platform-auth,moai-framework-electron,moai-platform-chrome-extension}/SKILL.md` `related-skills:` field | REQ-WF005-005, REQ-WF005-008 |
| Skill body lang-reference cleanup (~12 files) | per research.md §3 (top 12 paths) | REQ-WF005-005, REQ-WF005-008, REQ-WF005-013 |
| Other dead-skill-ID cleanup (~13 files) | per research.md §4 | REQ-WF005-015 |
| Language rule cross-language reference fix (3 files) | `.claude/rules/moai/languages/{flutter,scala,cpp}.md` (lang-as-skill cross-refs) | REQ-WF005-005, REQ-WF005-008 |
| `.claude/skills/moai/workflows/run.md` language detection mapping fix | lines 338-352, 979 | REQ-WF005-008 |
| CHANGELOG entry | `CHANGELOG.md` (Unreleased section) | Trackable (TRUST 5) |

Embedded-template parity is a **HARD** requirement: every change to `.claude/skills/**/*.md` and `.claude/rules/**/*.md` must also be applied to the corresponding `internal/template/templates/.claude/...` source-of-truth file (per `CLAUDE.local.md` §2 Template-First Rule). `make build` regenerates `internal/template/embedded.go` after template edits.

### 1.4 Traceability Matrix (REQ → AC mapping)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC):

| REQ ID | Category | Mapped AC(s) |
|--------|----------|--------------|
| REQ-WF005-001 | Ubiquitous | AC-WF005-03 (paths frontmatter declaration confirms language-as-rules canonical location) |
| REQ-WF005-002 | Ubiquitous | AC-WF005-01 |
| REQ-WF005-003 | Ubiquitous | AC-WF005-02 |
| REQ-WF005-004 | Ubiquitous | AC-WF005-03 |
| REQ-WF005-005 | Ubiquitous | AC-WF005-06, AC-WF005-12 |
| REQ-WF005-006 | Event-Driven | AC-WF005-04 |
| REQ-WF005-007 | Event-Driven | AC-WF005-05 |
| REQ-WF005-008 | Event-Driven | AC-WF005-06 |
| REQ-WF005-009 | State-Driven | AC-WF005-09 |
| REQ-WF005-010 | State-Driven | AC-WF005-10 |
| REQ-WF005-011 | Optional | AC-WF005-07 (substitute pattern for cross-language abstraction is `moai-ref-*`) |
| REQ-WF005-012 | Optional | AC-WF005-09 (atomic reversal gate is the dual to "adding 17th language as rule") |
| REQ-WF005-013 | Complex (Unwanted) | AC-WF005-12 |
| REQ-WF005-014 | Complex (Unwanted) | AC-WF005-11 |
| REQ-WF005-015 | Complex (State+Event) | AC-WF005-07, AC-WF005-08 |

Coverage: **15/15 REQs mapped, 12/12 ACs validated** (some ACs map to multiple REQs).

---

## 2. Milestone Breakdown (M1-M5)

Each milestone is **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule).

### M1: Test scaffolding (RED phase) — Priority P0

Reference: `internal/template/commands_audit_test.go:1-60` (audit-test scaffold pattern).

Owner role: `expert-backend` (Go test) or direct manager-tdd execution.

Scope:
1. Create `internal/template/lang_boundary_audit_test.go` mirroring `commands_audit_test.go` structure.
2. Test func `TestNoLangSkillDirectory` walks `.claude/skills/` in the embedded FS, asserts no directory matches `^moai-lang-[a-z-]+$`. On match: `t.Errorf("LANG_AS_SKILL_FORBIDDEN: %s exists; language guidance must live in .claude/rules/moai/languages/, not as a skill", path)`.
3. Test func `TestRelatedSkillsNoLangReference` walks all `.claude/skills/**/SKILL.md` in the embedded FS, parses frontmatter (YAML), inspects `metadata.related-skills` field. On any `moai-lang-` substring: `t.Errorf("DEAD_LANG_SKILL_REFERENCE: %s related-skills contains %q; substitute with .claude/rules/moai/languages/<name>.md", path, token)`.
4. Test func `TestLanguageNeutrality` walks all `.claude/skills/**/SKILL.md` and `.claude/rules/**/*.md`, scans body bytes (excluding fenced code blocks), asserts no language-primacy regex matches. On match: `t.Errorf("LANG_NEUTRALITY_VIOLATION: %s body line %d contains %q; per CLAUDE.local.md §15, all 16 languages must be treated equally", path, line, match)`.
5. Run `go test ./internal/template/ -run "TestNoLangSkillDirectory|TestRelatedSkillsNoLangReference|TestLanguageNeutrality"` — confirm RED for `TestRelatedSkillsNoLangReference` (3 SKILL.md files have `moai-lang-*` in `related-skills`); GREEN for `TestNoLangSkillDirectory` (no `moai-lang-*` directories exist today) and `TestLanguageNeutrality` (no obvious primacy violations at baseline).

Verification gate before advancing to M2: at least 3 of `TestRelatedSkillsNoLangReference` subtests fail with the documented sentinel message; the directory-existence test passes (the `moai-lang-*` skill creation has not been attempted today, per research.md §3 — only references exist).

[HARD] No implementation code in M1 outside of test files.

### M2: "Language Guidance Lives in Rules" principle in skill-authoring.md (GREEN, part 1) — Priority P0

Owner role: `manager-docs`.

Scope:
1. Append a new section to `.claude/rules/moai/development/skill-authoring.md` titled `## Language Guidance Lives in Rules, Not Skills`. Body template:

   ```markdown
   ## Language Guidance Lives in Rules, Not Skills

   Per SPEC-V3R2-WF-005, the 16 supported languages live as **rules** under
   `.claude/rules/moai/languages/*.md`, never as skills.

   - **No `moai-lang-<name>` skill** may be created. Any PR adding such a
     skill directory triggers `LANG_AS_SKILL_FORBIDDEN` in CI.
   - **Canonical location**: `.claude/rules/moai/languages/<name>.md` for all
     16 supported languages: `cpp`, `csharp`, `elixir`, `flutter`, `go`,
     `java`, `javascript`, `kotlin`, `php`, `python`, `r`, `ruby`, `rust`,
     `scala`, `swift`, `typescript`. Canonical Dart name is `flutter` per
     `CLAUDE.local.md` §15.
   - **Loading mechanism**: each language rule uses `paths:` frontmatter for
     conditional loading (e.g., `paths: "**/*.py,**/pyproject.toml"`).
     Path-based loading is the structurally correct primary mechanism for
     language-scoped guidance; keyword-based skill activation is the wrong
     abstraction for files-on-disk language detection.
   - **Adding a 17th language**: create a new `.md` file under
     `.claude/rules/moai/languages/` with a `paths:` frontmatter; never a new
     skill. A reversal of this decision requires a new SPEC with an atomic
     migration plan covering all languages (no partial adoption).
   - **Cross-language abstraction**: when guidance applies across languages
     (general API design, security checklists), use the `moai-ref-*` skills
     (`moai-ref-api-patterns`, `moai-ref-owasp-checklist`) — not a
     `moai-lang-*` composite.
   - **CI guard**: `internal/template/lang_boundary_audit_test.go` enforces
     this principle.

   See `.claude/rules/moai/languages/*.md` (16 files) for the canonical
   per-language guidance, and `CLAUDE.local.md` §15 for the 16-language
   neutrality contract.
   ```

2. Mirror the edit into `internal/template/templates/.claude/rules/moai/development/skill-authoring.md`.
3. Run `make build` to regenerate `internal/template/embedded.go`.

Reference: `skill-authoring.md` insertion anchor is the natural append point per research.md §5 (after "## Tool Permissions by Category"). The new section title MUST appear verbatim so future `Grep`-based audits can locate it.

Verification: `grep -n "Language Guidance Lives in Rules" .claude/rules/moai/development/skill-authoring.md` returns the heading line.

### M3: Frontmatter `related-skills:` cleanup in 3 SKILL.md files (GREEN, part 2) — Priority P0

Owner role: `manager-docs`.

Scope:
1. For each of `.claude/skills/{moai-platform-auth,moai-framework-electron,moai-platform-chrome-extension}/SKILL.md` (per research.md §3 frontmatter targets):
   - Locate the `related-skills:` field in the YAML frontmatter (lines 19-20 area).
   - Remove every `moai-lang-<name>` token. Substitute with `.claude/rules/moai/languages/<name>.md` reference if the relationship is genuinely cross-cutting; otherwise drop the entry.
2. Mirror each edit into `internal/template/templates/.claude/skills/{moai-platform-auth,moai-framework-electron,moai-platform-chrome-extension}/SKILL.md`.
3. Run `make build`.

Verification: `go test ./internal/template/ -run TestRelatedSkillsNoLangReference` turns GREEN (3 of 3 frontmatter subtests pass).

[HARD] M3 must NOT modify other frontmatter fields, body content, or unrelated `related-skills` entries (e.g., `moai-domain-frontend`, `moai-platform-supabase`). Insert-only or remove-only.

### M4: Body cleanup of dead `moai-lang-*` references (GREEN, part 3) — Priority P0

Owner role: `manager-docs`.

Scope (per research.md §3):
1. `.claude/skills/moai/workflows/run.md` (lines 338-352, 979): replace each `moai-lang-<name>` mapping target with the rule-path reference `.claude/rules/moai/languages/<name>.md` OR rephrase as "load `.claude/rules/moai/languages/<name>.md` (auto-loaded via paths frontmatter)". The mapping itself (file marker → language) is preserved.
2. `.claude/skills/moai-foundation-core/modules/agents-reference.md` (lines 268, 278-282, 292): catalog table cleanup. Replace `moai-lang-<name>` rows with rule-path references or remove rows that no longer apply.
3. `.claude/skills/moai-framework-electron/SKILL.md` (lines 228, 230): body prose; substitute with rule-path references.
4. `.claude/skills/moai-platform-chrome-extension/SKILL.md` (lines 278, 279): body prose; substitute with rule-path references.
5. `.claude/skills/moai-platform-auth/SKILL.md` (line 225): body prose; substitute.
6. `.claude/skills/moai-platform-deployment/SKILL.md` (lines 398, 399): substitute.
7. `.claude/skills/moai-workflow-loop/SKILL.md` (lines 148, 149): substitute.
8. `.claude/skills/moai-domain-frontend/SKILL.md` (line 119): substitute.
9. `.claude/skills/moai-foundation-cc/reference/skill-examples.md` (lines 237, 427): substitute.
10. `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md` (lines 201, 392): substitute.
11. `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md` (lines 30, 439, 936): substitute.
12. `.claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md` (line 113): substitute.
13. `.claude/skills/moai-foundation-cc/references/examples.md` (line 175): substitute.
14. `.claude/rules/moai/development/agent-authoring.md` (line 165): replace the YAML schema example skill ID `moai-lang-go` with an existing skill ID such as `moai-domain-backend`.
15. `.claude/rules/moai/languages/scala.md` (line 131): replace `moai-lang-java` reference with `.claude/rules/moai/languages/java.md`.
16. `.claude/rules/moai/languages/flutter.md` (lines 94, 95): replace `moai-lang-swift` and `moai-lang-kotlin` with rule-path references.
17. `.claude/rules/moai/languages/cpp.md` (line 100): replace `moai-lang-rust` with rule-path reference.
18. Mirror all edits to `internal/template/templates/.claude/...`.
19. Run `make build` after the full sweep.

[HARD] M4 must NOT modify language rule body content beyond the specific lines identified above. The 16 language rules' core guidance (Tooling, Conventions, Best Practices) is preserved per `spec.md` §1.2 / §2.2 Out of Scope.

Verification: a final `grep -rn "moai-lang-" .claude/skills/ .claude/rules/` returns zero matches (or only matches inside fenced code blocks within the new audit test, which are acceptable).

### M5: Other dead-skill-ID cleanup + CHANGELOG + MX tags (GREEN, part 4 + Trackable) — Priority P1

Owner role: `manager-docs`.

Scope (per research.md §4):

#### M5a: `moai-essentials-debug` substitution (9 files)

In each file (per research.md §4.1), replace the bullet/line containing `moai-essentials-debug` with: `delegate to expert-debug agent for AI-powered debugging`.

Targets:
- `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:621`
- `.claude/rules/moai/languages/r.md:133`
- `.claude/rules/moai/languages/typescript.md:126`
- `.claude/rules/moai/languages/flutter.md:98`
- `.claude/rules/moai/languages/cpp.md:103`
- `.claude/rules/moai/languages/csharp.md:110`
- `.claude/rules/moai/languages/elixir.md:95`
- `.claude/rules/moai/languages/javascript.md:130`
- `.claude/rules/moai/languages/ruby.md:131`

#### M5b: `moai-quality-testing` substitution (1 file)

Replace `moai-quality-testing` with `moai-foundation-quality` + `moai-ref-testing-pyramid`.

Targets:
- `.claude/rules/moai/languages/kotlin.md:109`

#### M5c: `moai-quality-security` substitution (3 files)

Replace `moai-quality-security` with `moai-foundation-quality` + `moai-ref-owasp-checklist`.

Targets:
- `.claude/skills/moai-domain-backend/SKILL.md:110`
- `.claude/skills/moai-workflow-project/references/overview.md:120`
- `.claude/rules/moai/languages/flutter.md:97`

#### M5d: `moai-infra-docker` removal (2 files)

Remove the line entirely (no substitute; platform infra deferred per spec.md §2.1).

Targets:
- `.claude/rules/moai/languages/kotlin.md:110`
- `.claude/rules/moai/languages/java.md:121`

#### M5e: CHANGELOG + MX tags + final verification

1. Add CHANGELOG entry under `## [Unreleased]`:
   ```
   ### Changed
   - SPEC-V3R2-WF-005: Codified that language guidance lives as rules (`.claude/rules/moai/languages/*.md`), not as skills. Added "Language Guidance Lives in Rules, Not Skills" section to `.claude/rules/moai/development/skill-authoring.md`. Removed dead `moai-lang-*` references from ~17 skills/rules. Substituted `moai-essentials-debug`, `moai-quality-testing`, `moai-quality-security`, and `moai-infra-docker` references per REQ-WF005-015. CI guard `lang_boundary_audit_test.go` enforces forward-looking compliance.
   ```

2. Insert MX tags per §6 below.
3. Mirror all edits to `internal/template/templates/.claude/...`.
4. Run `make build` to regenerate `internal/template/embedded.go`.
5. Run full `go test ./...` from repo root. Verify ALL audit tests pass + 0 cascading failures (per `CLAUDE.local.md` §6 HARD rule "After fixing ANY test, run the FULL test suite to catch cascading failures").
6. Update `progress.md` for SPEC-V3R2-WF-005 with `run_complete_at` and `run_status: implementation-complete` after M1-M5d land.

[HARD] No new documents are created in `.moai/specs/` or `.moai/reports/` during M5 — this is a SPEC implementation phase, not a planning phase.

---

## 3. File:line Anchors (concrete edit targets)

### 3.1 To-be-modified (existing files)

| File | Anchor | Edit type | Reason |
|------|--------|-----------|--------|
| `.claude/rules/moai/development/skill-authoring.md` | After "## Tool Permissions by Category" section | Append `## Language Guidance Lives in Rules, Not Skills` section | M2 / REQ-WF005-001..004, 009 |
| `.claude/skills/moai-platform-auth/SKILL.md:20,225` | `related-skills` field + body line | Remove `moai-lang-typescript`; substitute body | M3 + M4 / REQ-WF005-005, 008 |
| `.claude/skills/moai-framework-electron/SKILL.md:19,228,230` | `related-skills` field + 2 body lines | Remove `moai-lang-typescript`/`moai-lang-javascript` from frontmatter; substitute body | M3 + M4 |
| `.claude/skills/moai-platform-chrome-extension/SKILL.md:19,278,279` | `related-skills` + 2 body lines | Same as above | M3 + M4 |
| `.claude/skills/moai/workflows/run.md:338-352,979` | language detection mapping + Phase 0.9 example | Substitute `moai-lang-<name>` with rule-path references | M4 / REQ-WF005-008 |
| `.claude/skills/moai-foundation-core/modules/agents-reference.md:268,278-282,292` | catalog table rows | Substitute or remove rows | M4 |
| `.claude/skills/moai-platform-deployment/SKILL.md:398,399` | body | Substitute | M4 |
| `.claude/skills/moai-workflow-loop/SKILL.md:148,149` | body | Substitute | M4 |
| `.claude/skills/moai-domain-frontend/SKILL.md:119` | body | Substitute | M4 |
| `.claude/skills/moai-foundation-cc/reference/skill-examples.md:237,427` | body | Substitute | M4 |
| `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md:201,392` | body | Substitute | M4 |
| `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:30,439,621,936` | body (621 is `moai-essentials-debug`) | Substitute | M4 + M5a |
| `.claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md:113` | body | Substitute | M4 |
| `.claude/skills/moai-foundation-cc/references/examples.md:175` | body | Substitute | M4 |
| `.claude/rules/moai/development/agent-authoring.md:165` | YAML schema example | Replace `moai-lang-go` with `moai-domain-backend` | M4 |
| `.claude/rules/moai/languages/scala.md:131` | "Related Skills" body | Replace `moai-lang-java` with rule-path | M4 |
| `.claude/rules/moai/languages/flutter.md:94,95,97,98` | "Related Skills" body | Replace 4 dead refs (M4: 94,95; M5c: 97; M5a: 98) | M4 + M5a + M5c |
| `.claude/rules/moai/languages/cpp.md:100,103` | "Related Skills" body | Replace 2 dead refs (M4: 100; M5a: 103) | M4 + M5a |
| `.claude/rules/moai/languages/r.md:133` | "Related Skills" body | Substitute `moai-essentials-debug` | M5a |
| `.claude/rules/moai/languages/typescript.md:126` | "Related Skills" body | Substitute `moai-essentials-debug` | M5a |
| `.claude/rules/moai/languages/csharp.md:110` | "Related Skills" body | Substitute `moai-essentials-debug` | M5a |
| `.claude/rules/moai/languages/elixir.md:95` | "Related Skills" body | Substitute `moai-essentials-debug` | M5a |
| `.claude/rules/moai/languages/javascript.md:130` | "Related Skills" body | Substitute `moai-essentials-debug` | M5a |
| `.claude/rules/moai/languages/ruby.md:131` | "Related Skills" body | Substitute `moai-essentials-debug` | M5a |
| `.claude/rules/moai/languages/kotlin.md:109,110` | "Related Skills" body | Substitute `moai-quality-testing` (M5b); remove `moai-infra-docker` (M5d) | M5b + M5d |
| `.claude/rules/moai/languages/java.md:121` | "Related Skills" body | Remove `moai-infra-docker` line | M5d |
| `.claude/skills/moai-domain-backend/SKILL.md:110` | body | Substitute `moai-quality-security` | M5c |
| `.claude/skills/moai-workflow-project/references/overview.md:120` | body | Substitute `moai-quality-security` | M5c |
| `CHANGELOG.md` | `## [Unreleased]` section | Add entry | M5e |

### 3.2 To-be-created (new files)

| File | Reason |
|------|--------|
| `internal/template/lang_boundary_audit_test.go` | REQ-WF005-002 + REQ-WF005-013 + REQ-WF005-014 CI guards (M1) |

### 3.3 NOT to be touched (preserved by reference)

The following files are referenced by tests but MUST NOT be modified by this SPEC's run phase. They define the rhythm WF-005 is *codifying*.

- `.claude/rules/moai/languages/<name>.md` (16 files) — Tooling, Version, Conventions, Best Practices content is preserved verbatim. Only the "Related Skills" section trailing lines per §3.1 above are edited.
- `internal/template/commands_audit_test.go:1-60` — audit-test scaffold pattern (preserve as scaffold reference; not modified).
- `CLAUDE.local.md` §15 — 16-language neutrality contract (preserved verbatim; the new principle in skill-authoring.md cross-links to it).
- All 16 `paths:` frontmatter declarations — confirmed correct per research.md §2 (REQ-WF005-004 is satisfied at baseline).

### 3.4 Reference citations (file:line)

Per `spec.md` §10 traceability and research.md §10, the following anchors are load-bearing and cited verbatim throughout this plan:

1. `.claude/rules/moai/development/skill-authoring.md:1-269` (insertion anchor)
2. `.claude/rules/moai/development/agent-authoring.md:165` (YAML schema example fix target)
3. `.claude/rules/moai/languages/python.md:1-2` (paths frontmatter exemplar)
4. `.claude/rules/moai/languages/flutter.md:94-98` (cross-language references + dead skill IDs)
5. `.claude/rules/moai/languages/scala.md:131` (cross-language reference)
6. `.claude/rules/moai/languages/cpp.md:100,103` (cross-language reference + dead skill ID)
7. `.claude/rules/moai/languages/kotlin.md:109,110` (dead skill IDs M5b + M5d)
8. `.claude/rules/moai/languages/java.md:121` (dead skill ID M5d)
9. `.claude/skills/moai/workflows/run.md:338-352,979` (language detection mapping fix)
10. `.claude/skills/moai-foundation-core/modules/agents-reference.md:268,278-282,292` (catalog cleanup)
11. `.claude/skills/moai-platform-auth/SKILL.md:20,225` (frontmatter + body)
12. `.claude/skills/moai-framework-electron/SKILL.md:19,228,230` (frontmatter + body)
13. `.claude/skills/moai-platform-chrome-extension/SKILL.md:19,278,279` (frontmatter + body)
14. `internal/template/commands_audit_test.go:1-60` (audit-test scaffold)
15. `CLAUDE.local.md:§15` (16-language neutrality)
16. `.claude/rules/moai/workflow/spec-workflow.md:172-204` (Plan Audit Gate)

Total: **16 distinct file:line anchors** (exceeds the §Hard-Constraints minimum of 10 for plan.md).

---

## 4. Technology Stack Constraints

Per `spec.md` §2.2 Out of Scope, **no new technology** is introduced:

- No new Go modules or external libraries.
- No new directory structure (uses existing `.claude/skills/`, `.claude/rules/`, `internal/template/`).
- No language-specific tooling beyond the existing Go test framework.
- Embedded-template machinery (`go:embed` in `internal/template/embed.go`) is reused as-is.

The only additive surfaces are:

- One new Go test file (`internal/template/lang_boundary_audit_test.go`) using the standard `testing` package.
- One new section heading in `skill-authoring.md`.
- One CHANGELOG entry.
- ~28 line-level substitutions across skills/rules.

---

## 5. Risk Analysis & Mitigations

Extends `spec.md` §8 risks with concrete file-path mitigations.

| Risk | Probability | Impact | Mitigation Anchor |
|------|-------------|--------|-------------------|
| Substitution introduces broken cross-references (rule-path that doesn't exist) | M | M | M4 substitution pattern uses verified rule-path `.claude/rules/moai/languages/<name>.md` for the 16 confirmed files (research.md §2). Run-phase agent verifies file existence before substitution. |
| User confused by `--related-skills` field showing rule-path instead of skill ID | L | L | M2 principle section explicitly documents the convention (rule paths are the canonical reference target). FAQ entry in skill-authoring.md if needed in M5. |
| `moai-lang-*` reference removal breaks `Skill()` invocation in agent prompts | L | M | These references are dead today (no `moai-lang-*` skill exists), so `Skill("moai-lang-<name>")` already fails silently. M4 cleanup makes the failure mode explicit (rule-path reference). No runtime regression. |
| 16-language neutrality test (`TestLanguageNeutrality`) generates false positives on legitimate per-language guidance | M | L | Regex set in M1 targets primacy phrases ("X is primary", "only Y is supported") not per-language sections. Per-language headings ("## Python Tooling") are exempt. Test §6.2 in research.md documents the verb allowlist. |
| `flutter.md` cross-language references (`moai-lang-swift`, `moai-lang-kotlin`) substitution breaks the cross-platform documentation | L | M | Substitute with rule-path references that already work today (`.claude/rules/moai/languages/swift.md`, `.../kotlin.md` exist per research.md §2). Documentation accuracy improves, not degrades. |
| Language rule "Related Skills" sections become inconsistent across the 16 files | M | L | M5 sweep applies a uniform substitution pattern (delegate-to-agent for debug, ref-skill for testing/security). Pattern documented in M5 task body. |
| Embedded-template mismatch (skill edited but `internal/template/templates/` not synced) | M | H | M2/M3/M4/M5 each end with `make build` step. CI runs `go test ./...` which exercises embedded FS. Per `CLAUDE.local.md` §2 Template-First Rule HARD constraint. |
| Future SPEC proposes lang-as-skill before reading WF-005 | M | M | REQ-WF005-012 atomic-reversal gate documented in `skill-authoring.md` principle section (M2). plan-auditor enforces SPEC structure consistency. |
| `agent-authoring.md:165` example substitution breaks YAML schema demonstration | L | L | `moai-domain-backend` is an existing skill ID (verified by `ls .claude/skills/`); substitution preserves the YAML array syntax demonstration. |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and `.claude/skills/moai/workflows/plan.md` mx_plan MANDATORY rule.

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/template/lang_boundary_audit_test.go:TestNoLangSkillDirectory` | `@MX:ANCHOR fan_in=16 - SPEC-V3R2-WF-005 REQ-WF005-002,007,009 enforcer; guards 16 language rules against being migrated to skills. Touching this test signature affects the contract for all 16 supported languages.` | The test *is* the contract enforcer. Future PRs that propose lang-as-skill will fail or pass this test — high downstream impact. |
| `.claude/rules/moai/development/skill-authoring.md` `## Language Guidance Lives in Rules, Not Skills` section | `@MX:ANCHOR fan_in=N - Language-as-rules canonical decision; cross-referenced by all skill authors. Changes here affect every future language-related SPEC.` | The principle is the single source of truth for language vs skill classification. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/template/lang_boundary_audit_test.go` package-level doc comment | `@MX:NOTE - Audit suite for SPEC-V3R2-WF-005 language-as-rules contract. Three tests: TestNoLangSkillDirectory (REQ-WF005-002), TestRelatedSkillsNoLangReference (REQ-WF005-013), TestLanguageNeutrality (REQ-WF005-014).` | Documents the test scope so future maintainers do not delete or reduce coverage. |
| `.claude/rules/moai/development/skill-authoring.md` new section | `@MX:NOTE - Language-as-rules per SPEC-V3R2-WF-005; do not propose moai-lang-* skills. See REQ-WF005-012 for atomic reversal gate.` | Carries the WHY for future readers. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `.claude/skills/moai/workflows/run.md` Phase 0.9 language detection section | `@MX:WARN @MX:REASON - Future PRs may be tempted to revert to moai-lang-* skill references here. The current rule-path mapping (lines 338-352 post-M4) MUST remain pointing to .claude/rules/moai/languages/<name>.md. Any moai-lang-* substring fails TestRelatedSkillsNoLangReference.` | Most likely point of future regression — language detection feature is high-traffic. |
| `.claude/rules/moai/development/agent-authoring.md:165` YAML schema example | `@MX:WARN @MX:REASON - Example skill ID must be an existing skill (e.g., moai-domain-backend). Reverting to moai-lang-go or any moai-lang-* fails the audit test.` | The example is illustrative but easily reverted. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC)

This SPEC is a documentation/audit publication, not a feature build. No `@MX:TODO` markers are planned — all work converges to GREEN within M1-M5. Any `@MX:TODO` introduced during implementation must be resolved before final M5 commit (per `.claude/rules/moai/workflow/mx-tag-protocol.md` GREEN-phase resolution rule).

### 6.5 MX tag count summary

- @MX:ANCHOR: 2 targets
- @MX:NOTE: 2 targets
- @MX:WARN: 2 targets
- @MX:TODO: 0 targets
- **Total**: 6 MX tag insertions planned across 5 distinct files

---

## 7. Solo Mode Discipline

[HARD] All run-phase work for SPEC-V3R2-WF-005 executes in:

```
/Users/goos/MoAI/moai-adk-go
```

Branch: `feature/SPEC-V3R2-WF-005-language-rules-boundary` (already checked out per session context).

[HARD] No worktree is used for this SPEC (per user directive: solo mode, no worktree). All Read/Write/Edit tool invocations use absolute paths under the main project root.

[HARD] `make build` and `go test ./...` execute from the repo root: `cd /Users/goos/MoAI/moai-adk-go && make build && go test ./...`.

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md:172-204`). The plan is **audit-ready** only if all are PASS.

- [x] **C1: Frontmatter v0.2.0 schema** — `spec.md` frontmatter has all 9 required fields (`id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number`). Verified by reading `spec.md:1-22`.
- [x] **C2: HISTORY entry for v0.2.0** — `spec.md:30-31` HISTORY table has v0.2.0 row with description.
- [x] **C3: 15 EARS REQs across 5 categories** — `spec.md:99-156` (Ubiquitous 5, Event-Driven 3, State-Driven 2, Optional 2, Complex 3).
- [x] **C4: 12 ACs all map to REQs (100% coverage)** — `spec.md:162-173`. Each AC explicitly cites the REQ(s) it maps to. Plan §1.4 traceability matrix confirms 15/15 REQ → AC mapping.
- [x] **C5: BC scope clarity** — `spec.md:19-20` (`breaking: false`, `bc_id: []`) + research.md §8 (declaration-level analysis).
- [x] **C6: File:line anchors ≥10** — research.md §10 (38 anchors), this plan.md §3.4 (16 anchors).
- [x] **C7: Exclusions section present** — `spec.md:42-49` Non-Goals + `spec.md:65-72` Out of Scope (7 entries each).
- [x] **C8: TDD methodology declared** — this plan §1.2 + `.moai/config/sections/quality.yaml`.
- [x] **C9: mx_plan section** — this plan §6 (6 MX tag insertions across 4 categories).
- [x] **C10: Risk table with mitigations** — `spec.md:189-195` + this plan §5 (9 risks, file-anchored mitigations).
- [x] **C11: Solo mode path discipline** — this plan §7 (3 HARD rules, no worktree per user directive).
- [x] **C12: No implementation code in plan documents** — verified self-check: this plan, research.md, acceptance.md, tasks.md contain only natural-language descriptions, regex patterns, file paths, and YAML/Markdown templates. No Go function bodies or executable code.
- [x] **C13: Acceptance.md G/W/T format with edge cases** — verified in acceptance.md §1-12.
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — verified in tasks.md §M1-M5 (manager-tdd / expert-backend / manager-docs assignments).
- [x] **C15: Cross-SPEC consistency** — WF-001 blocked-by dependency status: completed (per spec.md §9.1); WF-005 is independent of WF-003/WF-004.

All 15 criteria PASS → plan is **audit-ready**.

---

## 9. Implementation Order Summary

Run-phase agent executes in this order (P0 first, dependencies resolved):

1. **M1 (P0)**: Create `lang_boundary_audit_test.go` with 3 tests. Confirm RED for `TestRelatedSkillsNoLangReference` (3 frontmatter violations); GREEN for `TestNoLangSkillDirectory` and `TestLanguageNeutrality` at baseline.
2. **M2 (P0)**: Add "Language Guidance Lives in Rules" section to `skill-authoring.md` (+ embedded template parity + `make build`).
3. **M3 (P0)**: Frontmatter `related-skills:` cleanup in 3 SKILL.md files (+ template parity + `make build`). Confirm `TestRelatedSkillsNoLangReference` GREEN.
4. **M4 (P0)**: Body cleanup of dead `moai-lang-*` references across ~17 files (+ template parity + `make build`). Confirm `grep -rn "moai-lang-" .claude/skills .claude/rules` returns zero non-test matches.
5. **M5 (P1)**: Other dead-skill-ID cleanup (~13 files), CHANGELOG entry, MX tags per §6, final `make build` + `go test ./...`. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete`.

Total milestones: 5. Total file edits (existing): ~28. Total file creations (new): 1 (audit test). Total CHANGELOG entries: 1.

---

End of plan.md.
