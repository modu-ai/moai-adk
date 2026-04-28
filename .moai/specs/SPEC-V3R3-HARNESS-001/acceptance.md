---
id: SPEC-V3R3-HARNESS-001
artifact: acceptance
version: "0.1.0"
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
---

# SPEC-V3R3-HARNESS-001 — Acceptance Criteria

This document defines the observable, testable acceptance criteria for SPEC-V3R3-HARNESS-001. Every criterion is bound to one or more EARS requirements via the AC ↔ REQ traceability matrix at the end.

## AC-HARNESS-01 — Meta-Harness Skill Exists with Apache 2.0 Attribution and 7-Phase Body

**Covers**: REQ-HARNESS-001, REQ-HARNESS-004, REQ-HARNESS-010

### Given
- The repository is checked out at the merge commit of SPEC-V3R3-HARNESS-001.
- `make build` has been executed successfully.

### When
- A reader opens `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` and `.claude/skills/moai-meta-harness/SKILL.md`.
- A new Claude Code session starts in a project initialized with `moai init`.

### Then
- Both files exist and are byte-identical (excluding template variable placeholders).
- The first 20 lines contain an Apache 2.0 attribution block crediting `https://github.com/revfactory/harness`.
- The body documents the 7-Phase workflow (Discovery, Analysis, Synthesis, Skeleton, Customization, Evaluation, Iteration).
- The body cross-references existing MoAI agents (manager-spec, manager-tdd, manager-ddd, expert-*, evaluator-active) — no new agent introduced.
- The frontmatter includes `triggers.phases: ["plan", "run", "sync"]` and `triggers.keywords: ["harness", "project-init", "meta-skill"]`.
- `moai doctor` reports the meta-harness skill as a valid static-core member.

### Edge Cases
- If a user has manually edited their local `.claude/skills/moai-meta-harness/SKILL.md` and runs `moai update`, the migrator overwrites it (template-managed) and prints a warning suggesting the user submit upstream PRs for changes.
- If `make build` was skipped, the local mirror differs from the template; CI catches this via existing template-sync test.

### Verification Method
- Manual review of attribution block.
- `grep -l "revfactory/harness" .claude/skills/moai-meta-harness/SKILL.md`
- Frontmatter parsed via existing `internal/template/skill_frontmatter_test.go`-style table-driven test.

---

## AC-HARNESS-02 — 16 Static Skills Verifiably Removed

**Covers**: REQ-HARNESS-002

### Given
- The repository is checked out at the merge commit.
- The verbatim 16-skill list is encoded in `internal/template/skills_removal_test.go`.

### When
- A user runs `ls .claude/skills/moai-* | wc -l` from the repository root.
- CI runs `go test ./internal/template/...`.

### Then
- The count is exactly **23** (22 base + 1 meta-harness).
- None of the 16 directories (`moai-domain-backend`, `moai-domain-frontend`, `moai-domain-database`, `moai-domain-db-docs`, `moai-domain-mobile`, `moai-framework-electron`, `moai-library-shadcn`, `moai-library-mermaid`, `moai-library-nextra`, `moai-tool-ast-grep`, `moai-platform-auth`, `moai-platform-deployment`, `moai-platform-chrome-extension`, `moai-workflow-research`, `moai-workflow-pencil-integration`, `moai-formats-data`) exist under `.claude/skills/` or `internal/template/templates/.claude/skills/`.
- `TestRemovedSkillsNotPresent` passes.

### Edge Cases
- A future PR accidentally re-adds one of the removed skills → CI fails with an explicit "removed skill X reappeared" error.
- A user-area skill named `my-harness-domain-backend` is permitted (different prefix).

### Verification Method
- Shell command: `ls .claude/skills/moai-* | wc -l` returns `23`.
- `go test -run TestRemovedSkillsNotPresent ./internal/template/...` exits 0.

---

## AC-HARNESS-03 — Namespace Separation Enforced (moai-* vs my-harness-*)

**Covers**: REQ-HARNESS-003

### Given
- A test fixture project containing both static (`moai-foundation-core`) and dynamic (`my-harness-test-skill`) skills.

### When
- The user runs `moai doctor`.
- The user runs `moai update`.

### Then
- `moai doctor` lists the static core skills as PASS.
- `moai doctor` lists `my-harness-test-skill` as INFO (user customization detected).
- `moai doctor` flags any unauthorized `moai-*` skill outside the 23-skill allowlist as WARN.
- After `moai update`, `.claude/skills/my-harness-test-skill/SKILL.md` is byte-identical to the pre-update state (no modification).
- `internal/template/templates/.claude/skills/` contains no entries matching `my-harness-*`.

### Edge Cases
- A user creates a skill named `moai-custom-foo` (wrong prefix) → doctor WARN.
- A user creates a skill named `my-harness-foo` with invalid frontmatter → doctor INFO + frontmatter validation runs.

### Verification Method
- Integration test: `internal/cli/doctor/namespace_test.go` covers all four cases (valid static, valid user, invalid moai-, invalid my-harness-).
- Diff fixture: `diff -rq fixture/before fixture/after` after `moai update` shows zero changes under `my-harness-*`.

---

## AC-HARNESS-04 — Archive Directory Preserves Removed Skills + User Customization Inviolate

**Covers**: REQ-HARNESS-007, REQ-HARNESS-008

### Given
- A test fixture project containing all 16 removed skills (each with original `SKILL.md`, `modules/`, `examples.md`, `reference.md`).
- The same fixture also contains user customizations: `.moai/harness/main.md`, `.claude/agents/my-harness/test-agent.md`, `.claude/skills/my-harness-test/SKILL.md`.

### When
- The user runs `moai update`.

### Then
- `.moai/archive/skills/v2.16/<skill-id>/` exists for each of the 16 removed skills.
- Each archive entry preserves the full original directory structure (`SKILL.md`, `modules/`, `examples.md`, `reference.md`).
- A second invocation of `moai update` does not re-archive (idempotent).
- `.moai/harness/main.md` is byte-identical to pre-update.
- `.claude/agents/my-harness/test-agent.md` is byte-identical to pre-update.
- `.claude/skills/my-harness-test/SKILL.md` is byte-identical to pre-update.
- `moai migrate restore-skill moai-domain-backend` reproduces `.claude/skills/moai-domain-backend/` with byte-identical content (excluding runtime artifacts like `.moai/cache/`).

### Edge Cases
- User has modified one of the 16 skills locally before running `moai update` → the modified version is archived (preserves user's edits, not upstream).
- Archive directory already exists (manual creation) → `moai update` overwrites with diagnostic warning.
- `moai migrate restore-skill <id>` invoked for a skill that was never archived → exits non-zero with clear error.

### Verification Method
- Integration test: `internal/cli/update/archive_integration_test.go` runs the full fixture migration, then asserts directory equality via `filepath.Walk` + SHA-256 hash comparison.
- Manual byte-for-byte verification: `diff -rq .moai/archive/skills/v2.16/moai-domain-backend/ <git-show>:.claude/skills/moai-domain-backend/` returns zero.

---

## AC-HARNESS-05 — moai update --dry-run Preview Shows 16 Archive Operations

**Covers**: REQ-HARNESS-005

### Given
- The fixture from AC-HARNESS-04 (all 16 skills present).

### When
- The user runs `moai update --dry-run`.

### Then
- Stdout contains exactly 16 lines of the form `archive: .claude/skills/<skill-id> → .moai/archive/skills/v2.16/<skill-id>` (one per removed skill).
- Stdout contains one line `install: .claude/skills/moai-meta-harness (from template)`.
- Stdout ends with a summary: `total: 16 skills archived, 1 skill installed, 0 user customizations modified`.
- No filesystem mutation occurs (verified by file mtime stability).

### Edge Cases
- Project has zero of the 16 removed skills → output reads `total: 0 skills archived, 1 skill installed (or already present)`.
- Project has only some of the 16 → only the present ones are listed; absent ones are skipped silently.
- Project has user customization under `my-harness-*` → output line `preserve: .claude/skills/my-harness-test (user customization)`.

### Verification Method
- Integration test: `internal/cli/update/dry_run_test.go` captures stdout and asserts on expected line count and order.

---

## AC-HARNESS-06 — BC-V3R3-007 Announced in CHANGELOG, Release Notes, and Migration Guide

**Covers**: REQ-HARNESS-006

### Given
- The merge commit of SPEC-V3R3-HARNESS-001.

### When
- A reader opens `CHANGELOG.md`, `.moai/release/RELEASE-NOTES-v2.17.0.md`, and `.moai/release/MIGRATION-v2.17.0.md`.

### Then
- `CHANGELOG.md` v2.17.0 entry contains a `### Breaking Changes` section listing BC-V3R3-007 with the verbatim 16-skill list.
- `RELEASE-NOTES-v2.17.0.md` opens with a prominent BC-V3R3-007 callout (visible in the first 30 lines).
- `MIGRATION-v2.17.0.md` includes:
  - The verbatim 16-skill list (matching §3 of spec.md).
  - Automatic migration command: `moai update`.
  - Manual fallback steps (for users who can't run `moai update`).
  - Restore command syntax: `moai migrate restore-skill <skill-id>`.
  - Deprecation timeline: 1 minor release grace window (v2.17.x → v2.18.0 final removal).
  - Apache 2.0 attribution to revfactory/harness.

### Edge Cases
- A user reading only CHANGELOG.md must see a link to MIGRATION-v2.17.0.md.
- All three documents must agree on the 16-skill list (no drift).

### Verification Method
- Documentation review checklist run by manager-docs.
- Automated check: `scripts/check-migration-docs.sh` parses the 16-skill list from spec.md §3 and verifies it appears verbatim in MIGRATION-v2.17.0.md.

---

## AC-HARNESS-07 — Generated Harness Validation Hook (evaluator-active Sprint Contract)

**Covers**: REQ-HARNESS-009

### Given
- The meta-harness skill body documents the validation hook.
- The evaluator profile referenced by the meta-harness exists.

### When
- A reader inspects `moai-meta-harness/SKILL.md` for the validation hook description.
- A reader inspects the evaluator profile for the +60% target citation.

### Then
- `moai-meta-harness/SKILL.md` contains a section titled "Generated Harness Validation" describing the Sprint Contract handoff to evaluator-active.
- The evaluator profile (e.g., `.moai/config/evaluator-profiles/harness-generation.yaml` or equivalent) cites the +60% A/B effectiveness data from revfactory/harness as the design target.
- The Sprint Contract reference points to design constitution §11.5.

### Edge Cases
- The +60% citation must be traceable to revfactory/harness public documentation (not a fabricated number).
- evaluator-active profile must already be loadable without errors before this SPEC merges (existing infrastructure).

### Verification Method
- Manual review of skill body section.
- Grep: `grep -r "revfactory/harness" .moai/config/evaluator-profiles/` returns at least one match in the harness-relevant profile.
- Cross-reference: design constitution §11.5 citation is checked by manager-docs review.

---

## Definition of Done

The SPEC is complete when **all** of the following are true:

1. All 7 acceptance criteria above pass (AC-HARNESS-01 through AC-HARNESS-07).
2. All 10 EARS requirements (REQ-HARNESS-001 through REQ-HARNESS-010) are covered (see traceability matrix below).
3. TRUST 5 quality gates pass:
   - **Tested**: 85%+ coverage on new Go code (`internal/cli/update/archive*.go`, `internal/template/skills_removal_test.go`).
   - **Readable**: Skill body passes Hextra/Markdown lint, attribution block clearly demarcated.
   - **Unified**: All Go code formatted with `gofmt`, all Markdown matches existing skill style.
   - **Secured**: No archive path traversal vulnerabilities; archive operations confined to `.moai/archive/skills/v2.16/`.
   - **Trackable**: Conventional Commits with Korean body, BC-V3R3-007 cited in commit messages.
4. Template-First verification: `make build` is run before commit; `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` is the source of truth.
5. Plan audit gate (Phase 0.5 of `/moai run`): plan-auditor PASS verdict.
6. CHANGELOG.md, RELEASE-NOTES-v2.17.0.md, MIGRATION-v2.17.0.md cross-link consistently.
7. depends_on (SPEC-V3R3-HARNESS-LEARNING-001) is in `draft` or `approved` status (not blocked).
8. No file under `.claude/agents/moai/` modified by this SPEC.
9. FROZEN zone (`.claude/rules/moai/design/constitution.md` §2) untouched.
10. PR description references BC-V3R3-007 prominently.

---

## AC ↔ REQ Traceability Matrix

| AC ID | Description | Covers REQ-IDs |
|-------|-------------|----------------|
| AC-HARNESS-01 | Meta-harness skill exists + Apache 2.0 + 7-Phase + frontmatter | REQ-HARNESS-001, REQ-HARNESS-004, REQ-HARNESS-010 |
| AC-HARNESS-02 | 16 skills verifiably removed | REQ-HARNESS-002 |
| AC-HARNESS-03 | Namespace separation enforced (moai-*/my-harness-*) | REQ-HARNESS-003 |
| AC-HARNESS-04 | Archive structure + user customization preserved | REQ-HARNESS-007, REQ-HARNESS-008 |
| AC-HARNESS-05 | moai update --dry-run preview | REQ-HARNESS-005 |
| AC-HARNESS-06 | BC-V3R3-007 in CHANGELOG/RELEASE-NOTES/MIGRATION | REQ-HARNESS-006 |
| AC-HARNESS-07 | Validation hook + evaluator-active Sprint Contract | REQ-HARNESS-009 |

**Coverage**: 10 REQs ↔ 7 ACs, 100% (every REQ covered). Total matrix rows: 7.
