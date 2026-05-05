# SPEC-V3R2-WF-005 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.2.0, `research.md` v0.1.0, `plan.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 1B)     | Initial G/W/T conversion of 12 ACs (AC-WF005-01 through AC-WF005-12)   |

---

## Scope

This document converts each of the 12 ACs from `spec.md` §6 into Given/When/Then format with happy-path + edge-case + test-mapping notation. Test-mapping uses `internal/template/lang_boundary_audit_test.go` (new, M1) and direct skill-content audit for content-level assertions; runtime assertions for the language detection mapping (REQ-WF005-006) are observed via `paths:` frontmatter loader behavior in Claude Code.

Notation:
- **Test mapping** identifies which Go test function (or manual verification step) covers the AC.
- AC-WF005-05 has the highest test importance (REQ-WF005-007 CI guard for `LANG_AS_SKILL_FORBIDDEN`).

---

## AC-WF005-01 — v3 skill tree contains no `moai-lang-*` directory

Maps to: REQ-WF005-002.

### Happy path

- **Given** the v3 skill tree at `.claude/skills/`
- **When** an inspector lists subdirectories or runs `ls .claude/skills/ | grep -E '^moai-lang-'`
- **Then** zero matches are returned
- **And** the embedded FS (`internal/template/templates/.claude/skills/`) shows the same — no `moai-lang-*` directories exist

### Edge case — embedded FS check

- **Given** `make build` has regenerated `internal/template/embedded.go`
- **When** the audit test walks the embedded FS
- **Then** no path matching `moai-lang-[a-z-]+` is found

### Test mapping

- Static directory audit: `TestNoLangSkillDirectory` in `internal/template/lang_boundary_audit_test.go` (M1) — fails if any `moai-lang-*` directory exists in the embedded FS.
- Manual verification: `ls .claude/skills/ | grep -c '^moai-lang-'` returns 0.

---

## AC-WF005-02 — `skill-authoring.md` documents the principle

Maps to: REQ-WF005-003.

### Happy path

- **Given** the run phase has completed M2 (principle section insertion)
- **When** an inspector reads `.claude/rules/moai/development/skill-authoring.md`
- **Then** a section titled `## Language Guidance Lives in Rules, Not Skills` exists
- **And** the section explicitly forbids the creation of `moai-lang-*` skills
- **And** the section references the canonical location `.claude/rules/moai/languages/`
- **And** the section cross-links to `CLAUDE.local.md` §15 for 16-language neutrality

### Edge case — embedded template parity

- **Given** the section has been added to `.claude/rules/moai/development/skill-authoring.md`
- **When** `make build` regenerates `internal/template/embedded.go`
- **Then** the embedded copy of `skill-authoring.md` (read via `EmbeddedTemplates()`) contains the identical section
- **And** `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` has been edited to match

### Test mapping

- Manual verification: post-M2, run `grep -n "Language Guidance Lives in Rules" .claude/rules/moai/development/skill-authoring.md` returns the heading line.
- Embedded parity: existing `internal/template/embed_test.go` exercises the embedded FS; the new audit test reads from the same FS to verify content.

---

## AC-WF005-03 — All 16 language rules have `paths:` frontmatter

Maps to: REQ-WF005-001, REQ-WF005-004.

### Happy path

- **Given** the 16 files at `.claude/rules/moai/languages/{cpp,csharp,elixir,flutter,go,java,javascript,kotlin,php,python,r,ruby,rust,scala,swift,typescript}.md`
- **When** the frontmatter of each file is parsed
- **Then** every file has a `paths:` field declaring the file glob pattern for auto-loading
- **And** each `paths:` value is a comma-separated string of glob patterns (e.g., `**/*.py,**/pyproject.toml`)
- **And** REQ-WF005-001 is satisfied: language guidance lives in `.claude/rules/moai/languages/` as rules

### Edge case — file count exactly 16

- **Given** the directory `.claude/rules/moai/languages/`
- **When** counting `.md` files
- **Then** exactly 16 files are present (no more, no fewer)
- **And** the filenames match the canonical list per CLAUDE.local.md §15

### Test mapping

- Manual verification (research.md §2 already verified at baseline): `grep -l "^paths:" .claude/rules/moai/languages/*.md | wc -l` returns 16.
- File count: `ls .claude/rules/moai/languages/*.md | wc -l` returns 16.

---

## AC-WF005-04 — Python project auto-loads `python.md` rule

Maps to: REQ-WF005-006.

### Happy path

- **Given** a user's project contains `pyproject.toml` and `*.py` files
- **When** Claude Code session starts in that project directory
- **Then** the rule loader matches the `paths: "**/*.py,**/pyproject.toml,**/requirements*.txt"` declaration in `.claude/rules/moai/languages/python.md`
- **And** the rule body content is loaded into the active context for any agent operating on Python files

### Edge case — multi-language monorepo

- **Given** a project contains both `pyproject.toml` (Python) and `package.json` with `"typescript"` (TypeScript)
- **When** Claude Code loads rules
- **Then** both `.claude/rules/moai/languages/python.md` and `.claude/rules/moai/languages/typescript.md` are loaded
- **And** both bodies are available in the active context

### Edge case — non-language file changes

- **Given** the user is editing `.md` documentation files only (no Python sources)
- **When** the rule loader evaluates path matches
- **Then** `python.md` may or may not load depending on the rule loader's project-level vs file-level scoping
- **And** REQ-WF005-006 specifically applies to project detection at session start, not per-file evaluation

### Test mapping

- Runtime verification: integration with Claude Code's rule loader (existing system; not re-tested by this SPEC).
- Static verification: research.md §2 confirmed `paths:` glob pattern is correctly declared in `python.md`.

---

## AC-WF005-05 — PR adding `moai-lang-rust/` skill is rejected

Maps to: REQ-WF005-007. **Highest test importance — this is the regression guard.**

### Happy path (current state — green CI)

- **Given** the current state of `.claude/skills/` (no `moai-lang-*` directories per AC-WF005-01)
- **When** `TestNoLangSkillDirectory` runs in CI
- **Then** the test PASSES
- **And** CI all-green status is achieved

### Failure scenario (regression detection)

- **Given** a future PR adds `.claude/skills/moai-lang-rust/SKILL.md` to the repository
- **When** that PR's CI runs `go test ./internal/template/ -run TestNoLangSkillDirectory`
- **Then** the test FAILS with output matching:
  - `t.Errorf("LANG_AS_SKILL_FORBIDDEN: %s exists; language guidance must live in .claude/rules/moai/languages/, not as a skill", path)`
- **And** the CI run reports the violation
- **And** the PR is blocked from merge until the directory is removed (and the rust guidance is moved to `.claude/rules/moai/languages/rust.md`, which already exists)

### Edge case — `moai-lang-` substring elsewhere

- **Given** a skill named `moai-domain-frontend` mentions `moai-lang-typescript` in body prose
- **When** `TestNoLangSkillDirectory` runs
- **Then** the test does NOT flag the substring (the test scans directory paths, not body content)
- **And** body prose mentions are caught by the separate `TestRelatedSkillsNoLangReference` (AC-WF005-12)

### Test mapping

- Primary: `TestNoLangSkillDirectory` in `internal/template/lang_boundary_audit_test.go` (M1).
- Subtests: one per any matching directory found (zero subtests at GREEN state).

---

## AC-WF005-06 — `related-skills` field with `moai-lang-typescript` is removed

Maps to: REQ-WF005-005, REQ-WF005-008.

### Happy path

- **Given** `.claude/skills/moai-platform-auth/SKILL.md:20` had `related-skills: "moai-platform-supabase, moai-platform-vercel, moai-lang-typescript, moai-domain-backend, moai-expert-security"` (per research.md §3)
- **When** the M3 cleanup commit runs
- **Then** the `moai-lang-typescript` token is removed from the `related-skills:` field
- **And** the resulting field reads `related-skills: "moai-platform-supabase, moai-platform-vercel, moai-domain-backend, moai-expert-security"`
- **And** `TestRelatedSkillsNoLangReference` passes for this file

### Edge case — body mention

- **Given** the same file's body line 225 references `moai-lang-typescript: TypeScript patterns for auth SDKs`
- **When** the M4 cleanup runs
- **Then** the body line is replaced with `.claude/rules/moai/languages/typescript.md - TypeScript patterns for auth SDKs (auto-loaded via paths frontmatter)`
- **And** `TestRelatedSkillsNoLangReference` continues to pass (body is not the test target; only frontmatter `related-skills` is)

### Edge case — all 3 frontmatter targets

- **Given** all 3 SKILL.md files identified in research.md §3 (`moai-platform-auth`, `moai-framework-electron`, `moai-platform-chrome-extension`)
- **When** M3 cleanup runs
- **Then** all 3 files have their `related-skills:` field free of `moai-lang-*` tokens
- **And** all 3 corresponding subtests in `TestRelatedSkillsNoLangReference` pass

### Test mapping

- Static frontmatter audit: `TestRelatedSkillsNoLangReference` walks all `.claude/skills/**/SKILL.md` files in the embedded FS, parses frontmatter, asserts `related-skills` value contains no `moai-lang-` substring.
- Pseudocode: For each SKILL.md, parse frontmatter YAML; if `metadata.related-skills` field exists; check `bytes.Contains(value, []byte("moai-lang-"))`; on true, emit `t.Errorf("DEAD_LANG_SKILL_REFERENCE: %s related-skills contains %q; substitute with .claude/rules/moai/languages/<name>.md", path, token)`.

---

## AC-WF005-07 — `moai-quality-testing` body reference is replaced

Maps to: REQ-WF005-015 + REQ-WF005-011.

### Happy path

- **Given** `.claude/rules/moai/languages/kotlin.md:109` had `moai-quality-testing - JUnit 5, MockK, TestContainers integration` (per research.md §4.2)
- **When** the M5b cleanup runs
- **Then** the line is replaced with `moai-foundation-quality + moai-ref-testing-pyramid - JUnit 5, MockK, TestContainers integration`
- **And** subsequent grep for `moai-quality-testing` in `.claude/rules/` returns zero matches

### Edge case — single substitution scope

- **Given** research.md §4.2 identified exactly 1 file (`kotlin.md`) with this reference
- **When** M5b runs
- **Then** exactly 1 line is changed
- **And** no false-positive substitutions occur in unrelated files

### Test mapping

- Manual verification: post-M5b, run `grep -rn "moai-quality-testing" .claude/skills .claude/rules` returns zero matches.
- Cross-reference verification: post-M5b, the substitute references (`moai-foundation-quality`, `moai-ref-testing-pyramid`) MUST point to existing skills. Verify with `ls .claude/skills/moai-foundation-quality .claude/skills/moai-ref-testing-pyramid` (or document if these are also pending creation, in which case substitute with rule-path or remove the line).

---

## AC-WF005-08 — `moai-essentials-debug` body reference is replaced

Maps to: REQ-WF005-015.

### Happy path

- **Given** the 9 files identified in research.md §4.1 (8 language rules + 1 sub-agent example)
- **When** the M5a cleanup runs
- **Then** each line containing `moai-essentials-debug` is replaced with `delegate to expert-debug agent for AI-powered debugging` (or the equivalent context-appropriate rephrasing)
- **And** subsequent grep for `moai-essentials-debug` in `.claude/skills/` and `.claude/rules/` returns zero matches

### Edge case — varied surrounding text

- **Given** different files use different phrasings (e.g., `moai-essentials-debug for AI-powered debugging`, `moai-essentials-debug for debugging TypeScript applications`, `moai-essentials-debug for debugging .NET applications`)
- **When** M5a runs
- **Then** each substitution preserves the language-specific suffix (e.g., "for debugging TypeScript applications" remains; only the prefix changes)
- **And** the substitution pattern `delegate to expert-debug agent for <suffix>` is applied consistently

### Edge case — sub-agent-examples.md edge case

- **Given** `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:621` has `skills: moai-essentials-debug, moai-core-code-reviewer, moai-context7-integration`
- **When** M5a runs
- **Then** the YAML schema example is updated to either remove `moai-essentials-debug` or substitute with an existing agent reference (`expert-debug`)
- **And** the demonstration of YAML array syntax is preserved

### Test mapping

- Manual verification: post-M5a, run `grep -rn "moai-essentials-debug" .claude/skills .claude/rules` returns zero matches.
- Cross-reference verification: `expert-debug` agent exists per `.claude/agents/expert-debug.md` (verified at baseline).

---

## AC-WF005-09 — Adding 17th language as skill is rejected

Maps to: REQ-WF005-009, REQ-WF005-012.

### Happy path

- **Given** a future PR proposes adding language support for "Crystal" by creating `.claude/skills/moai-lang-crystal/SKILL.md`
- **When** CI runs `TestNoLangSkillDirectory`
- **Then** the test FAILS with `LANG_AS_SKILL_FORBIDDEN: .claude/skills/moai-lang-crystal/ exists; language guidance must live in .claude/rules/moai/languages/, not as a skill`
- **And** the PR is blocked
- **And** the PR author is directed to either (a) follow REQ-WF005-009 by creating `.claude/rules/moai/languages/crystal.md` instead, or (b) follow REQ-WF005-012 by submitting a new SPEC with an atomic migration plan covering all 17 languages

### Edge case — atomic reversal scenario

- **Given** a hypothetical SPEC-V4-LANG-AS-SKILL is submitted that proposes migrating ALL 16 (or 17) languages to skills atomically
- **When** that SPEC is reviewed
- **Then** the SPEC must include a complete migration plan, a rollback plan, and unanimous coverage (no partial adoption)
- **And** REQ-WF005-012 sets the gate; the audit test would be removed only as part of that approved migration

### Test mapping

- Primary: `TestNoLangSkillDirectory` (M1) — the same test that catches AC-WF005-05 catches the 17th-language case.
- Manual verification: PR review process documented in `skill-authoring.md` section "Language Guidance Lives in Rules, Not Skills" (M2).

---

## AC-WF005-10 — `flutter.md` filename is "flutter" (not "dart")

Maps to: REQ-WF005-010.

### Happy path

- **Given** the directory `.claude/rules/moai/languages/`
- **When** an inspector lists files matching the Flutter language rule
- **Then** the file is named `flutter.md` (per `ls` output and CLAUDE.local.md §15)
- **And** the file's `paths:` frontmatter declares `**/*.dart,**/pubspec.yaml,**/pubspec.lock` (using `.dart` source extension because that is the actual file suffix)
- **And** project_markers detection logic uses the canonical name "flutter" for this language

### Edge case — body content references "Dart"

- **Given** `flutter.md` body legitimately discusses Dart language features (since Flutter uses Dart)
- **When** auditing the file
- **Then** body mentions of "Dart" are acceptable as language content, not a violation of REQ-WF005-010
- **And** REQ-WF005-010 specifically governs the canonical filename and project_markers key, not body content

### Edge case — no `dart.md` file exists

- **Given** the languages directory
- **When** searching for `dart.md`
- **Then** no such file exists (only `flutter.md`)
- **And** any future attempt to add `dart.md` triggers a 17th-language gate per AC-WF005-09

### Test mapping

- Manual verification: `ls .claude/rules/moai/languages/flutter.md` returns the file; `ls .claude/rules/moai/languages/dart.md` returns "No such file or directory".
- Static verification: research.md §2 confirmed canonical filename at baseline.

---

## AC-WF005-11 — Skill body claiming "go is primary" triggers `LANG_NEUTRALITY_VIOLATION`

Maps to: REQ-WF005-014.

### Happy path (current state — green CI)

- **Given** the current state of `.claude/skills/**/SKILL.md` and `.claude/rules/**/*.md` body content
- **When** `TestLanguageNeutrality` runs in CI
- **Then** the test PASSES (no language-primacy phrases match at baseline)
- **And** CI all-green status is achieved

### Failure scenario (regression detection)

- **Given** a future PR adds the line "Go is the primary language; other languages are planned" to a skill body
- **When** that PR's CI runs `go test ./internal/template/ -run TestLanguageNeutrality`
- **Then** the test FAILS with output matching:
  - `t.Errorf("LANG_NEUTRALITY_VIOLATION: %s body line %d contains %q; per CLAUDE.local.md §15, all 16 languages must be treated equally", path, line, match)`
- **And** the CI run reports the violation
- **And** the PR is blocked from merge until the language-primacy phrasing is removed

### Edge case — per-language sections allowed

- **Given** a skill body has `## Python Tooling`, `## TypeScript Tooling`, etc., as section headings
- **When** the regex evaluates these headings
- **Then** the headings are NOT flagged (they are per-language section structure, not primacy claims)
- **And** the test PASSES on these headings

### Edge case — code blocks excluded

- **Given** a skill body contains a fenced code block with example showing `language: go` in YAML
- **When** the regex evaluation runs
- **Then** content within the fenced code block is excluded from the scan (per research.md §6.2 "excluding code blocks")
- **And** documentation-only mentions of language identifiers do not falsely trigger the test

### Test mapping

- Primary: `TestLanguageNeutrality` in `internal/template/lang_boundary_audit_test.go` (M1).
- Subtests: one per affected file (zero subtests at GREEN state).

---

## AC-WF005-12 — Skill body mentioning `moai-lang-python` triggers `DEAD_LANG_SKILL_REFERENCE`

Maps to: REQ-WF005-013.

### Happy path (post-M4 — green CI)

- **Given** the post-M4 state of `.claude/skills/**/SKILL.md` and `.claude/rules/**/*.md` body content (all `moai-lang-*` references substituted per plan.md §2 M4)
- **When** `TestRelatedSkillsNoLangReference` runs in CI
- **Then** the test PASSES (no `moai-lang-` substring in any `related-skills` field)
- **And** the body-level audit (manual `grep -rn`) returns zero non-test matches

### Failure scenario (regression detection — frontmatter)

- **Given** a future PR adds `related-skills: "moai-lang-python, moai-domain-backend"` to a SKILL.md frontmatter
- **When** that PR's CI runs `go test ./internal/template/ -run TestRelatedSkillsNoLangReference`
- **Then** the test FAILS with `DEAD_LANG_SKILL_REFERENCE: <path> related-skills contains "moai-lang-python"; substitute with .claude/rules/moai/languages/python.md`
- **And** the PR is blocked

### Failure scenario (regression detection — body prose)

- **Given** a future PR adds the line `See moai-lang-python for additional patterns` to a skill body
- **When** auditing
- **Then** the audit warning fires (DEAD_LANG_SKILL_REFERENCE) with rule-path suggestion
- **And** the PR is required to substitute with `.claude/rules/moai/languages/python.md (auto-loaded via paths frontmatter)`

Note: The test func `TestRelatedSkillsNoLangReference` primarily targets frontmatter `related-skills:` field. Body-prose matches may be caught by an extended scan in the same test or by the manual `grep` step in M4. Plan §2 M4 verification step requires zero non-test `moai-lang-` matches.

### Edge case — substitute reference must point to existing rule

- **Given** the substitute pattern is `.claude/rules/moai/languages/<name>.md`
- **When** the substitution is applied
- **Then** the target file MUST exist (verify with `ls .claude/rules/moai/languages/<name>.md`)
- **And** `<name>` MUST be one of the 16 canonical language names per CLAUDE.local.md §15

### Test mapping

- Primary: `TestRelatedSkillsNoLangReference` in `internal/template/lang_boundary_audit_test.go` (M1).
- Manual verification: post-M4, `grep -rn "moai-lang-" .claude/skills .claude/rules` returns zero non-test matches.

---

## Quality Gate Hooks (cross-reference)

### TRUST 5 framework alignment

- **Tested**: AC-WF005-05/11/12 enforce CI guards via Go test; AC-WF005-06/07/08 enforce content cleanup via static audit. All 12 ACs have explicit test mapping.
- **Readable**: The new "Language Guidance Lives in Rules" section (M2) uses consistent template wording; substitution patterns in M4/M5 follow uniform conventions documented in plan.md §2.
- **Unified**: Embedded-template parity (`make build` after every skill/rule edit per `CLAUDE.local.md` §2 Template-First Rule) ensures source and embedded copies remain in sync.
- **Secured**: No new attack surface — this SPEC only adds documentation sections, removes dead references, and adds a static audit test. No runtime behavior change.
- **Trackable**: CHANGELOG entry (M5e) + commit messages (per `CLAUDE.local.md` §4 Conventional Commits) provide audit trail.

### LSP quality gates

Per `.moai/config/sections/quality.yaml` `lsp_quality_gates`:

- **plan phase**: `require_baseline: true` — captured at this SPEC's plan completion (current commit at session start).
- **run phase**: `max_errors: 0`, `max_type_errors: 0`, `max_lint_errors: 0`, `allow_regression: false` — `lang_boundary_audit_test.go` must compile cleanly; no new lint warnings introduced.
- **sync phase**: `max_warnings: 10`, `require_clean_lsp: true` — verified post-M5e by running `golangci-lint run` and `go vet ./...`.

### Pre-submission self-review checklist

Per `.claude/rules/moai/workflow/spec-workflow.md` Pre-submission Self-Review:

- [ ] Full diff reviewed against this acceptance.md before commit.
- [ ] Asked "Is there a simpler approach?" — answer documented in plan.md §1.3 (this is the simpler approach: declarative principle + static audit + sweep substitution, not a runtime refactor or skill-class introduction).
- [ ] Asked "Would removing any changes still satisfy the SPEC?" — minimum set is M1 (audit test) + M2 (principle section) + M3 (frontmatter cleanup) + M5e (CHANGELOG); M4 (body sweep) and M5a-d (other dead-skill cleanup) are necessary for the spec.md §2.1 In Scope contract.

---

## Definition of Done (DoD)

The implementation is "done" when ALL of the following are true:

1. `lang_boundary_audit_test.go` exists with 3 test functions (`TestNoLangSkillDirectory`, `TestRelatedSkillsNoLangReference`, `TestLanguageNeutrality`), and `go test ./internal/template/ -run TestNoLangSkillDirectory` and the related tests pass.
2. `.claude/rules/moai/development/skill-authoring.md` contains a `## Language Guidance Lives in Rules, Not Skills` section with the principle, canonical location, paths-loading explanation, and atomic-reversal gate (REQ-WF005-012).
3. The 3 SKILL.md files (`moai-platform-auth`, `moai-framework-electron`, `moai-platform-chrome-extension`) have their `related-skills:` frontmatter free of any `moai-lang-*` tokens.
4. `grep -rn "moai-lang-" .claude/skills .claude/rules` returns zero non-test matches (test files in `internal/template/` are excluded from this grep).
5. `grep -rn "moai-essentials-debug\|moai-quality-testing\|moai-quality-security\|moai-infra-docker" .claude/skills .claude/rules` returns zero matches.
6. `internal/template/templates/.claude/...` mirrors all skill/rule edits (embedded template parity).
7. `make build` regenerates `internal/template/embedded.go` cleanly.
8. Full repository test suite passes: `go test ./...` returns 0 (per `CLAUDE.local.md` §6 HARD rule).
9. CHANGELOG `## [Unreleased]` section has the SPEC-V3R2-WF-005 entry.
10. MX tags per plan.md §6 inserted in all 5 target locations (2 ANCHOR + 2 NOTE + 2 WARN = 6 tags).
11. `progress.md` updated with `run_complete_at` and `run_status: implementation-complete`.

---

End of acceptance.md.
