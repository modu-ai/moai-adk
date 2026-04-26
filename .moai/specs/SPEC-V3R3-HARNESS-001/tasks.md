---
id: SPEC-V3R3-HARNESS-001
artifact: tasks
version: "0.1.0"
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
---

# SPEC-V3R3-HARNESS-001 — Implementation Tasks

Decomposed task list grouped by milestone (M1-M5 from plan.md). Each task lists owner agent, file paths, exit gate, and AC reference. Priority order is M1 → M2 → M3 → M4 → M5.

## M1 — Meta-Harness Skill Authoring (Priority Critical)

### T-M1-01: Verify revfactory/harness availability + Apache 2.0 license

- **Owner**: manager-spec (this session) or expert-debug
- **Action**: Confirm the public repo `https://github.com/revfactory/harness` is reachable and the LICENSE file is Apache 2.0. Cache a snapshot of the 7-Phase definitions to `/tmp/harness-analysis/` for use in T-M1-02.
- **Exit**: License confirmed, snapshot saved.
- **AC**: AC-HARNESS-01 (attribution accuracy depends on this).

### T-M1-02: Author moai-meta-harness/SKILL.md (Template-First source)

- **Owner**: builder-skill
- **File**: `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md`
- **Action**:
  1. Top-of-file Apache 2.0 attribution block (per plan.md §3.2).
  2. Frontmatter with `name: moai-meta-harness`, `version: 0.1.0`, `triggers.phases: ["plan", "run", "sync"]`, `triggers.keywords: ["harness", "project-init", "meta-skill"]`, `metadata.generated_by: "moai-adk"`, `metadata.upstream_source: "revfactory/harness"`.
  3. Body documenting the 7-Phase workflow with the adaptation table from plan.md §3.1.
  4. Cross-references to existing MoAI agents (no new agent introduction).
  5. "Generated Harness Validation" section describing Sprint Contract handoff (REQ-HARNESS-009).
- **Exit**: SKILL.md passes frontmatter schema validation, Apache 2.0 block in first 20 lines.
- **AC**: AC-HARNESS-01, AC-HARNESS-07.

### T-M1-03: Build local mirror via make build

- **Owner**: manager-tdd (during run phase) or main session
- **File**: `.claude/skills/moai-meta-harness/SKILL.md` (auto-generated)
- **Action**: Run `make build` to regenerate `internal/template/embedded.go` and confirm the local mirror appears.
- **Exit**: `diff -q internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md .claude/skills/moai-meta-harness/SKILL.md` returns zero (modulo template variable substitution).
- **AC**: AC-HARNESS-01.

### T-M1-04: Add modules and reference files (optional but recommended)

- **Owner**: builder-skill
- **Files**:
  - `internal/template/templates/.claude/skills/moai-meta-harness/modules/seven-phase-workflow.md`
  - `internal/template/templates/.claude/skills/moai-meta-harness/modules/integration-with-moai-agents.md`
  - `internal/template/templates/.claude/skills/moai-meta-harness/examples.md`
  - `internal/template/templates/.claude/skills/moai-meta-harness/reference.md`
- **Action**: Split content beyond 500 lines into modules per `moai-foundation-core` Modular System rules. Each module ≤ 500 lines.
- **Exit**: SKILL.md ≤ 500 lines, modules referenced via `Details in modules/...` syntax.
- **AC**: AC-HARNESS-01.

## M2 — 16 Static Skills Removal (Priority High)

### T-M2-01: Author skills_removal_test.go

- **Owner**: manager-tdd (RED phase)
- **File**: `internal/template/skills_removal_test.go` (NEW)
- **Action**: Encode the verbatim 16-skill list from spec.md §3 in a Go slice. Test asserts none of the 16 directories exist under `internal/template/templates/.claude/skills/`. RED first (test fails because skills still present).
- **Exit**: Test compiles, fails with explicit "removed skill X reappeared" message for all 16 skills.
- **AC**: AC-HARNESS-02.

### T-M2-02: Remove 16 skill directories from template

- **Owner**: manager-tdd (GREEN phase)
- **Files** (16 directories under `internal/template/templates/.claude/skills/`):
  - `moai-domain-backend/`
  - `moai-domain-frontend/`
  - `moai-domain-database/`
  - `moai-domain-db-docs/`
  - `moai-domain-mobile/`
  - `moai-framework-electron/`
  - `moai-library-shadcn/`
  - `moai-library-mermaid/`
  - `moai-library-nextra/`
  - `moai-tool-ast-grep/`
  - `moai-platform-auth/`
  - `moai-platform-deployment/`
  - `moai-platform-chrome-extension/`
  - `moai-workflow-research/`
  - `moai-workflow-pencil-integration/`
  - `moai-formats-data/`
- **Action**: `git rm -rf` each directory. Run `go test -run TestRemovedSkillsNotPresent ./internal/template/...` to confirm GREEN.
- **Exit**: Test passes. `ls internal/template/templates/.claude/skills/moai-* | wc -l` returns 23.
- **AC**: AC-HARNESS-02.

### T-M2-03: Rebuild and remove local mirror

- **Owner**: manager-tdd
- **Action**: Run `make build` to regenerate `internal/template/embedded.go`. Then remove the corresponding 16 directories under `.claude/skills/` (local mirror).
- **Exit**: `ls .claude/skills/moai-* | wc -l` returns 23.
- **AC**: AC-HARNESS-02.

### T-M2-04: Verify foundation/workflow/ref/design/copywriting/brand-design preserved

- **Owner**: manager-tdd
- **Action**: Sanity check that none of the FROZEN or non-removed skills were accidentally deleted. Compare against the `staticCoreAllowlist` from plan.md §3.5.
- **Exit**: All 22 base + 1 meta = 23 skills present.
- **AC**: AC-HARNESS-02.

## M3 — Namespace Separation Enforcement (Priority High)

### T-M3-01: Extend moai doctor with skills allowlist check

- **Owner**: manager-tdd
- **Files**:
  - `internal/cli/doctor/skills.go` (NEW or EXTENSION)
  - `internal/cli/doctor/skills_test.go` (NEW)
- **Action**: Hardcode the 23-skill allowlist. For each `.claude/skills/*/`, classify as PASS (in allowlist), WARN (moai- prefix outside allowlist), or INFO (my-harness- prefix). Table-driven test with all four cases.
- **Exit**: `moai doctor` against fixture project produces expected verdicts.
- **AC**: AC-HARNESS-03.

### T-M3-02: Extend moai update to skip my-harness paths

- **Owner**: manager-tdd
- **Files**:
  - `internal/cli/update/sync.go` (extension)
  - `internal/cli/update/preserve_my_harness_test.go` (NEW)
- **Action**: Add path-pattern skip rules: any path matching `.claude/skills/my-harness-*` or `.claude/agents/my-harness/` is excluded from update operations. Test fixture: project with `my-harness-test-skill/SKILL.md` mtime + content captured pre-update; assert byte-identical post-update.
- **Exit**: Test passes. User customization mtime preserved.
- **AC**: AC-HARNESS-03, AC-HARNESS-04 (user customization preservation).

### T-M3-03: Build-time test: no my-harness-* in template

- **Owner**: manager-tdd
- **File**: `internal/template/templates_test.go` (extension)
- **Action**: Assert `internal/template/templates/.claude/skills/` contains no `my-harness-*` entries.
- **Exit**: Test passes.
- **AC**: AC-HARNESS-03.

## M4 — moai update Migrator + Archive (Priority High)

### T-M4-01: Implement archive function

- **Owner**: manager-tdd (RED → GREEN)
- **Files**:
  - `internal/cli/update/archive.go` (NEW)
  - `internal/cli/update/archive_test.go` (NEW)
- **Action**:
  1. RED: Write table-driven test covering 16 skills × {present, absent, partially-modified}. Each case asserts archive directory equality via SHA-256 hash comparison.
  2. GREEN: Implement `Archive(skillID string) error` that copies `.claude/skills/<id>/` → `.moai/archive/skills/v2.16/<id>/` recursively.
  3. Idempotency: detect existing archive, skip with diagnostic.
  4. Symlink + permission edge cases handled.
- **Exit**: All test cases pass. Coverage ≥ 85%.
- **AC**: AC-HARNESS-04.

### T-M4-02: Wire archive into moai update flow

- **Owner**: manager-tdd
- **File**: `internal/cli/update/sync.go` (extension)
- **Action**: For each detected legacy skill (matching the 16-skill list), call `Archive()` then remove from `.claude/skills/`. Print one-line status: `archive: <id> → .moai/archive/skills/v2.16/<id>`. Final summary: `total: N skills archived, 1 skill installed, 0 user customizations modified`.
- **Exit**: Integration test fixture passes (16 skills archived, meta-harness installed).
- **AC**: AC-HARNESS-04, AC-HARNESS-05.

### T-M4-03: Implement moai migrate restore-skill subcommand

- **Owner**: manager-tdd
- **Files**:
  - `internal/cli/migrate/restore_skill.go` (NEW)
  - `internal/cli/migrate/restore_skill_test.go` (NEW)
- **Action**: New subcommand `moai migrate restore-skill <skill-id>`. Reverses archive: copies `.moai/archive/skills/v2.16/<id>/` → `.claude/skills/<id>/`. Errors if archive entry missing. Errors if target already exists (force flag required).
- **Exit**: Round-trip test (archive → restore) produces byte-identical content for all 16 skills.
- **AC**: AC-HARNESS-04.

### T-M4-04: Implement moai update --dry-run

- **Owner**: manager-tdd
- **Files**:
  - `internal/cli/update/dry_run.go` (NEW or extension of `sync.go`)
  - `internal/cli/update/dry_run_test.go` (NEW)
- **Action**: `--dry-run` flag prints planned archive + install + preserve operations to stdout without filesystem mutation. Verify zero mtime changes via post-run snapshot.
- **Exit**: Test asserts expected output line count + summary line.
- **AC**: AC-HARNESS-05.

### T-M4-05: Idempotency test (run moai update twice)

- **Owner**: manager-tdd
- **File**: `internal/cli/update/idempotency_test.go` (NEW)
- **Action**: Run `moai update` once → record archive state. Run again → assert archive state unchanged, no duplicate entries, no error.
- **Exit**: Test passes.
- **AC**: AC-HARNESS-04 (edge case).

## M5 — Migration Guide + Release Documentation (Priority Medium)

### T-M5-01: Author MIGRATION-v2.17.0.md

- **Owner**: manager-docs
- **File**: `.moai/release/MIGRATION-v2.17.0.md` (NEW)
- **Action**: Sections required:
  1. **What changed** (BC-V3R3-007 summary).
  2. **Removed skills** (verbatim 16-skill list from spec.md §3).
  3. **Automatic migration** (`moai update` walkthrough with example output).
  4. **Manual fallback** (for users who cannot run `moai update`).
  5. **Restore command** (`moai migrate restore-skill <skill-id>` syntax + worked example).
  6. **Deprecation timeline** (1 minor grace window: v2.17.x supports restore, v2.18.0 removes archive support).
  7. **Apache 2.0 attribution** (revfactory/harness credit + license link).
- **Exit**: Document reviewed against REQ-HARNESS-006 checklist.
- **AC**: AC-HARNESS-06.

### T-M5-02: Update CHANGELOG.md with v2.17.0 entry

- **Owner**: manager-docs
- **File**: `CHANGELOG.md`
- **Action**: Add v2.17.0 entry with `### Breaking Changes` section listing BC-V3R3-007. Link to MIGRATION-v2.17.0.md. Cite Apache 2.0 attribution.
- **Exit**: Diff reviewed by manager-docs.
- **AC**: AC-HARNESS-06.

### T-M5-03: Author RELEASE-NOTES-v2.17.0.md

- **Owner**: manager-docs
- **File**: `.moai/release/RELEASE-NOTES-v2.17.0.md`
- **Action**: First 30 lines must include prominent BC-V3R3-007 callout. Reference SPEC-V3R3-HARNESS-001, SPEC-V3R3-DESIGN-PIPELINE-001, SPEC-V3R3-PROJECT-HARNESS-001 (the v2.17.0 cluster).
- **Exit**: Visible BC-V3R3-007 callout confirmed.
- **AC**: AC-HARNESS-06.

### T-M5-04: Bump version files

- **Owner**: manager-docs (during release prep)
- **Files**:
  - `.moai/config/sections/system.yaml`
  - `internal/template/templates/.moai/config/config.yaml`
  - `README.md` (Version line)
  - `README.ko.md` (Version line)
- **Action**: All to v2.17.0. Single commit `docs(release): v2.17.0 prep`.
- **Exit**: All four files reflect v2.17.0.
- **AC**: AC-HARNESS-06 (indirect — release artifacts must agree on version).

### T-M5-05: Update SPEC status to completed

- **Owner**: manager-docs (post-merge)
- **File**: `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md` frontmatter `status:` field
- **Action**: Change `status: draft` → `status: completed`. Add HISTORY entry with merge commit hash.
- **Exit**: status field updated.
- **AC**: AC-HARNESS-06.

## Cross-Cutting Tasks

### T-X-01: Validate plan-auditor PASS verdict before /moai run

- **Owner**: plan-auditor (auto-invoked by `/moai run` Phase 0.5)
- **Action**: Audit spec.md / plan.md / acceptance.md for EARS compliance, traceability, frontmatter schema, FROZEN zone preservation.
- **Exit**: PASS verdict logged to `.moai/reports/plan-audit/SPEC-V3R3-HARNESS-001-2026-04-26.md`.
- **AC**: All ACs (gate condition).

### T-X-02: Conventional Commit messages with Korean body

- **Owner**: All implementation agents
- **Action**: Each commit:
  - `spec(harness): SPEC-V3R3-HARNESS-001 — Meta-Harness + 16 정적 skills 제거 (BC-V3R3-007)` (this commit, SPEC creation)
  - `feat(skill): moai-meta-harness 신설 (revfactory/harness Apache 2.0 흡수)` (M1)
  - `feat(template): 16 정적 skills 제거 (BC-V3R3-007)` (M2)
  - `feat(doctor): namespace allowlist 추가 (moai-* / my-harness-* 분리)` (M3)
  - `feat(update): archive + restore-skill 마이그레이터` (M4)
  - `docs(release): v2.17.0 BC-V3R3-007 마이그레이션 가이드` (M5)
- **Exit**: All commits follow Conventional Commits, body in Korean per `.moai/config/sections/language.yaml`.

### T-X-03: Run TRUST 5 quality gate before merge

- **Owner**: manager-quality
- **Action**: 
  - Tested: `go test -race -cover ./...` (≥85% coverage on new files)
  - Readable: `golangci-lint run`
  - Unified: `gofmt -l` returns empty
  - Secured: archive path-traversal test in `archive_test.go`
  - Trackable: commit messages reviewed
- **Exit**: All five gates pass.

### T-X-04: PR creation with BC callout

- **Owner**: manager-git (post-implementation)
- **Action**: PR title: `feat(v3R3): SPEC-V3R3-HARNESS-001 — Meta-Harness Skill (BC-V3R3-007)`. Body: BC summary, AC checklist (7 items), test plan, link to MIGRATION-v2.17.0.md.
- **Exit**: PR created against main, base branch verified.

## Task Summary

| Milestone | Tasks | Owners |
|-----------|-------|--------|
| M1 (Meta-Harness Authoring) | 4 | manager-spec, builder-skill, manager-tdd |
| M2 (16 Skills Removal) | 4 | manager-tdd |
| M3 (Namespace Separation) | 3 | manager-tdd |
| M4 (Migrator + Archive) | 5 | manager-tdd |
| M5 (Documentation) | 5 | manager-docs |
| Cross-Cutting | 4 | plan-auditor, all agents, manager-quality, manager-git |
| **Total** | **25 tasks** | |

## Dispatch Order (priority-based, no time estimates)

1. M1 first (meta-harness skill must exist before everything else).
2. M2 in parallel with M3 (independent; M3 only requires the 23-skill count to be stable, achievable after M2).
3. M4 after M2 + M3 (migrator needs the 16-skill list and namespace rules locked).
4. M5 after M1-M4 (documentation ratifies the implementation).
5. T-X-01 (plan-auditor) runs BEFORE M1 (Plan Audit Gate Phase 0.5 of /moai run).
6. T-X-02, T-X-03 run continuously per commit / pre-merge.
7. T-X-04 (PR) runs last, after M5 completes.
