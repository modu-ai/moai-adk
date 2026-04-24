---
id: SPEC-V3R2-CON-003
title: "Rule tree consolidation pass (merge duplicates, relocate, frontmatter migration)"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 4 SPEC Writer
priority: P1 High
phase: "v3.0.0 — Phase 1 — Constitution & Foundation"
module: ".claude/rules/moai/, .moai/decisions/"
dependencies:
  - SPEC-V3R2-CON-001
related_gap: []
related_problem:
  - P-H10
  - P-H11
  - P-H12
  - P-H13
  - P-H14
  - P-R02
  - P-R03
related_pattern:
  - S-4
  - X-1
related_principle:
  - P2
  - P12
related_theme: "Layer 1: Constitution"
breaking: true
bc_id: [BC-V3R2-014, BC-V3R2-017]
lifecycle: spec-anchored
tags: "v3r2, constitution, consolidation, rule-tree, frontmatter, relocate"
---

# SPEC-V3R2-CON-003: Rule tree consolidation pass

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | Wave 4 SPEC Writer | Initial draft from r6-audit §4 recommendations |

---

## 1. Goal (목적)

Address the six concrete findings from r6-commands-hooks-style-rules.md §4 (rules audit): (1) `core/lsp-client.md` is misfiled as a rule — it is a SPEC decision record; (2) `workflow/workflow-modes.md` (195 LOC) and `workflow/spec-workflow.md` (217 LOC) overlap significantly on Plan-Run-Sync; (3) `workflow/team-protocol.md` (54 LOC) duplicates "Team File Ownership" content in `worktree-integration.md` (303 LOC); (4) `workflow/file-reading-optimization.md` is heuristic, not an enforceable rule; (5) 4 rule files use legacy `description + globs` frontmatter while rest use `paths:` CSV; (6) CLAUDE.md Section 8 duplicates `agent-common-protocol.md` AskUserQuestion rules.

Net effect: rule tree shrinks from 34 files to 31; frontmatter convention unifies to `paths:` CSV; constitutional sprawl (problem P-R02: ~800 lines across 4 files) is reduced by extracting duplication.

This is a **mechanical cleanup pass**, not a rule-text amendment. No HARD clause changes meaning; only file layout and frontmatter syntax shift.

## 2. Scope (범위)

### 2.1 In Scope

- Move `.claude/rules/moai/core/lsp-client.md` → `.moai/decisions/lsp-client-choice.md` (content unchanged; this is a decision record for SPEC-LSP-CORE-002, not an agent rule).
- Merge `workflow/workflow-modes.md` content into `workflow/spec-workflow.md` under a new subsection; delete the standalone file.
- Merge `workflow/team-protocol.md` content into `workflow/worktree-integration.md` under a new "Team File Ownership" subsection; delete the standalone file.
- Relocate `workflow/file-reading-optimization.md` to `.claude/skills/moai-foundation-context/references/file-reading-optimization.md` (heuristic guidance, belongs to skill-level reference).
- Migrate 4 files from `description + globs:` to `paths:` CSV frontmatter: `moai-constitution.md`, `coding-standards.md`, `team-protocol.md` (before the merge), `worktree-integration.md` (after absorbing team-protocol).
- Deduplicate CLAUDE.md §8 (User Interaction Architecture) against `agent-common-protocol.md` (User Interaction Boundary): keep `agent-common-protocol.md` as authoritative; rewrite CLAUDE.md §8 to reference it with at most one-line summaries per rule.
- Update all cross-references to moved/merged files (grep + update).
- Update `zone-registry.md` (from CON-001) pointers for moved and merged rules.
- Update `internal/template/templates/.claude/rules/moai/` — the source of truth — and run `make build` to regenerate embedded files per CLAUDE.local.md §2.

### 2.2 Out of Scope

- Zone assignment for new or moved rules (→ SPEC-V3R2-CON-001 zone registry is the source of truth; CON-003 only updates pointers).
- Amending the content of any HARD clause (→ SPEC-V3R2-CON-002 amendment protocol).
- Creating new rules or adding new clauses to existing files.
- Consolidating language rules under `.claude/rules/moai/languages/` (→ SPEC-V3R2-WF-005 decides rule vs skill boundary).
- Restructuring skills (→ SPEC-V3R2-WF-001).

## 3. Environment (환경)

Current rule-tree state (34 files, per r6-commands-hooks-style-rules.md §4):

- `.claude/rules/moai/core/` (5 files): `moai-constitution.md`, `agent-common-protocol.md`, `lsp-client.md`, `settings-management.md`, `conventions.md`.
- `.claude/rules/moai/workflow/` (8 files) including the 4 consolidation targets.
- `.claude/rules/moai/development/` (3 files).
- `.claude/rules/moai/design/` (1 file — FROZEN design/constitution.md).
- `.claude/rules/moai/languages/` (16 language rules).
- `.claude/rules/moai/core/lsp-client.md` (94 LOC) — a SPEC decision record, misfiled.

Frontmatter audit (r6 §4.3):
- Modern `paths: "**/*.py,**/*.go"` CSV format used by most files.
- 4 files use legacy `description + globs:` YAML array format: `moai-constitution.md`, `coding-standards.md`, `team-protocol.md`, `worktree-integration.md`.

CLAUDE.md Section 8 currently reproduces the same rules as `agent-common-protocol.md §User Interaction Boundary`, creating constitutional sprawl (problem P-R02).

Template-First discipline: every rule edit must be mirrored in `internal/template/templates/.claude/rules/moai/` with `make build` run before commit (CLAUDE.local.md §2).

## 4. Assumptions (가정)

- Rule content preserved verbatim during mechanical move/merge operations. No HARD clause is rephrased.
- `moai` CLI grep-based cross-reference updater is reliable (consumers search for literal file paths like `.claude/rules/moai/workflow/workflow-modes.md`; these references can be rewritten mechanically).
- The `paths:` CSV frontmatter syntax is stable in Claude Code 2.1.111+ (per coding-standards.md compatibility table).
- `.moai/decisions/` directory convention is appropriate for SPEC decision records (the lsp-client.md destination). The file already references SPEC-LSP-CORE-002.
- `make build` is available in the dev workflow to regenerate `internal/template/embedded.go`.
- Breaking changes BC-V3R2-014 (lsp-client move) and BC-V3R2-017 (workflow-modes + team-protocol merge) are communicated in the migration tool (SPEC-V3R2-MIG-001).

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-CON-003-001: The system SHALL ensure the final rule tree under `.claude/rules/moai/` contains exactly 31 files (down from 34) after the consolidation pass.
- REQ-CON-003-002: All moved and merged rule content SHALL preserve verbatim clause text; no HARD rule text is rephrased as part of this consolidation.
- REQ-CON-003-003: The consolidated `.claude/rules/moai/workflow/spec-workflow.md` SHALL contain a new subsection absorbing `workflow-modes.md` content in full.
- REQ-CON-003-004: The consolidated `.claude/rules/moai/workflow/worktree-integration.md` SHALL contain a new "Team File Ownership" subsection absorbing `team-protocol.md` content in full.
- REQ-CON-003-005: The 4 target files SHALL use `paths:` CSV frontmatter syntax after migration: `moai-constitution.md`, `coding-standards.md`, `team-protocol.md` (pre-merge), `worktree-integration.md` (post-merge).
- REQ-CON-003-006: CLAUDE.md Section 8 (User Interaction Architecture) SHALL be shortened to reference `agent-common-protocol.md §User Interaction Boundary` with at most one-line summaries per rule; the authoritative text lives in the rule file.
- REQ-CON-003-007: The relocated decision record at `.moai/decisions/lsp-client-choice.md` SHALL cite SPEC-LSP-CORE-002 in its HISTORY section so the decision's origin is discoverable.

### 5.2 Event-driven

- REQ-CON-003-010: WHEN the consolidation script runs, the system SHALL first compute a diff summary and present it via AskUserQuestion before making any file changes.
- REQ-CON-003-011: WHEN any file is moved, merged, or deleted during this pass, the system SHALL search the repository for references to the old path and rewrite them to the new path atomically.
- REQ-CON-003-012: WHEN the `zone-registry.md` (from CON-001) has entries whose `file` pointer points at a moved rule, the system SHALL update those entries to the new path in the same commit.
- REQ-CON-003-013: WHEN `make build` is invoked after the consolidation, the system SHALL succeed with zero errors and the embedded template tree SHALL reflect exactly 31 files under `.claude/rules/moai/`.

### 5.3 State-driven

- REQ-CON-003-020: WHILE a CON-003 consolidation is in progress, concurrent constitution amendments (CON-002) SHALL be blocked to prevent merge conflicts on the same rule files.
- REQ-CON-003-021: WHILE the local project has uncommitted changes to `.claude/rules/moai/`, the consolidation script SHALL refuse to run and instruct the user to commit or stash first.

### 5.4 Optional

- REQ-CON-003-030: WHERE the configuration key `constitution.consolidation.dry_run: true` is set, the system SHALL produce the diff summary and target-tree listing without modifying any file.
- REQ-CON-003-031: WHERE `MOAI_CONSOLIDATION_SKIP_TEMPLATE_BUILD=1` is set (CI-only), the system SHALL skip the `make build` step at the end of consolidation; the developer is responsible for running it.

### 5.5 Complex

- REQ-CON-003-040: WHILE migrating a file from `description + globs:` frontmatter to `paths:` CSV, IF the source `globs:` list contains a single element AND the element is a language-specific glob (e.g., `**/*.py`), THEN the migrator SHALL convert it to a single-item CSV and preserve the `description:` field as a comment on the first body line (not lost); ELSE the migrator emits a multi-element CSV.
- REQ-CON-003-041: IF CLAUDE.md §8 deduplication produces a one-line summary that diverges in meaning from the authoritative clause in `agent-common-protocol.md`, THEN the pass SHALL halt and surface the divergence via AskUserQuestion; ELSE the shortened §8 is written.
- REQ-CON-003-042: WHEN the CLAUDE.md total character count after consolidation EXCEEDS 40,000, the system SHALL emit a warning (not a hard block) referencing coding-standards.md §File Size Limits and suggesting further extraction to `.claude/rules/moai/`.

## 6. Acceptance Criteria

- AC-CON-003-01: Given the pre-consolidation rule tree with 34 files, When the CON-003 pass runs to completion, Then `.claude/rules/moai/` recursively contains exactly 31 files. (maps REQ-CON-003-001)
- AC-CON-003-02: Given the original `workflow-modes.md` content, When compared against the new subsection inside `spec-workflow.md` after merge, Then every HARD clause and bullet is present verbatim. (maps REQ-CON-003-002, REQ-CON-003-003)
- AC-CON-003-03: Given the original `team-protocol.md` content, When compared against the new subsection inside `worktree-integration.md`, Then every HARD clause and bullet is present verbatim. (maps REQ-CON-003-002, REQ-CON-003-004)
- AC-CON-003-04: Given the 4 legacy-frontmatter files, When CON-003 migration completes, Then YAML `paths:` CSV keys are present in all 4 and no `globs:` key remains. (maps REQ-CON-003-005)
- AC-CON-003-05: Given CLAUDE.md with pre-CON-003 §8, When the pass rewrites §8, Then §8 contains no standalone rule text exceeding 5 lines; each rule is a one-line summary linking to `agent-common-protocol.md`. (maps REQ-CON-003-006)
- AC-CON-003-06: Given `core/lsp-client.md` exists pre-pass, When CON-003 runs, Then `.claude/rules/moai/core/lsp-client.md` no longer exists, `.moai/decisions/lsp-client-choice.md` exists with identical content plus a SPEC-LSP-CORE-002 HISTORY pointer. (maps REQ-CON-003-007)
- AC-CON-003-07: Given cross-references to moved file paths elsewhere in the repo, When CON-003 completes, Then `grep -r` finds zero references to the old paths (`workflow/workflow-modes.md`, `workflow/team-protocol.md`, `core/lsp-client.md`, `workflow/file-reading-optimization.md`). (maps REQ-CON-003-011)
- AC-CON-003-08: Given a zone-registry entry pointing at a consolidated file, When CON-003 completes, Then the entry's `file` field points to the new location. (maps REQ-CON-003-012)
- AC-CON-003-09: Given `make build` is invoked post-consolidation, When it runs, Then it succeeds and the embedded tree has exactly 31 rule files. (maps REQ-CON-003-013)
- AC-CON-003-10: Given `constitution.consolidation.dry_run: true`, When the pass runs, Then no file is modified and a diff summary is emitted to stdout. (maps REQ-CON-003-030)
- AC-CON-003-11: Given uncommitted changes exist in `.claude/rules/moai/`, When the consolidation script is invoked, Then it refuses with an instructive error. (maps REQ-CON-003-021)
- AC-CON-003-12: Given a legacy-frontmatter file has `globs: ["**/*.py"]` and `description: "Python rules"`, When the migrator runs, Then the output has `paths: "**/*.py"` and the description is preserved as a body-first-line comment. (maps REQ-CON-003-040)
- AC-CON-003-13: Given CLAUDE.md §8 summarization produces text whose meaning diverges from `agent-common-protocol.md`, When the divergence detector runs, Then the pass halts and asks the user via AskUserQuestion. (maps REQ-CON-003-041)
- AC-CON-003-14: Given the post-consolidation CLAUDE.md exceeds 40,000 characters, When the pass completes, Then a warning is emitted citing coding-standards.md §File Size Limits. (maps REQ-CON-003-042)

## 7. Constraints (제약)

- Breaking changes BC-V3R2-014 and BC-V3R2-017 are declared; SPEC-V3R2-MIG-001 rewrites external references to the old paths.
- No HARD rule text is rephrased; mechanical move/merge only. If a merge surfaces an apparent rule-text improvement, it is logged but deferred to SPEC-V3R2-CON-002 amendment.
- Template-First discipline (CLAUDE.local.md §2) applies: changes land first in `internal/template/templates/.claude/rules/moai/`, then `make build` regenerates embedded; then local `.claude/rules/moai/` is synced.
- The consolidation must be idempotent: re-running the script on an already-consolidated tree is a no-op (detected by file-count + frontmatter audit).
- CI check: zero grep hits for `workflow/workflow-modes.md`, `workflow/team-protocol.md`, `core/lsp-client.md`, `workflow/file-reading-optimization.md` anywhere in the repository after consolidation.
- No changes to `.claude/rules/moai/languages/` (reserved for SPEC-V3R2-WF-005 language-boundary decision).

## 8. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cross-reference rewriting misses a consumer | MEDIUM | MEDIUM | Post-pass CI grep check; `moai doctor rules` reports dangling links |
| Subtle meaning change during merge | LOW | HIGH | Verbatim-clause requirement (REQ-CON-003-002); AC-CON-003-02/03 verify with diff |
| `paths:` CSV migration breaks Claude Code loading | LOW | HIGH | Minimum CC 2.1.111 requirement checked by `moai doctor`; rollback script preserves pre-pass backup |
| Template drift during mechanical moves | MEDIUM | MEDIUM | Template-First rule enforced; CI `diff -rq` between template and local |
| CLAUDE.md §8 summarization loses rule enforcement | LOW | HIGH | REQ-CON-003-041 divergence detector; ask-user gate before writing |
| lsp-client.md move breaks future SPEC-LSP-CORE-002 linkers | LOW | LOW | HISTORY pointer in relocated file; update `docs/design/` references |

## 9. Dependencies

### 9.1 Blocked by

- SPEC-V3R2-CON-001 (zone registry must exist so that CON-003 can update the `file:` pointers for relocated rules).

### 9.2 Blocks

- SPEC-V3R2-MIG-001 (user migrator rewrites external references to old rule paths).
- SPEC-V3R2-WF-005 (language rules vs skills boundary may further relocate language rules; must happen after CON-003 stabilizes the tree).

### 9.3 Related

- r6-commands-hooks-style-rules.md §4 (findings P-H10..P-H14).
- Problem catalog P-R02 Constitutional sprawl, P-R03 CLAUDE.md/common-protocol duplication.
- SPEC-V3R2-CON-002 amendment protocol — NOT invoked by this pass (mechanical only).

## 10. Traceability

- Theme: Layer 1 Constitution (master-v3 §4).
- Principles: P2 Constitutional Governance with FROZEN/EVOLVABLE zones (design-principles.md §P12 — maintaining zone clarity via tree hygiene); P12 File-First Primitives (rules live in markdown files; their relocation is a file operation).
- Problems addressed: P-H10 lsp-client.md misfiled; P-H11 workflow-modes/spec-workflow overlap; P-H12 team-protocol/worktree-integration overlap; P-H13 file-reading-optimization heuristic; P-H14 frontmatter inconsistency; P-R02 constitutional sprawl; P-R03 CLAUDE.md/agent-common-protocol duplication.
- Patterns: S-4 FROZEN + Graduation (tree hygiene supports graduation clarity); X-1 Markdown + YAML Frontmatter (frontmatter migration is an X-1 schema unification).
- Wave 1 sources: R6 §4 rules audit (35 + findings).
- Wave 2 sources: problem-catalog.md Cluster 5 adjacency and §3 severity distribution.
