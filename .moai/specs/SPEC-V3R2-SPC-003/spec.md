---
id: SPEC-V3R2-SPC-003
title: "SPEC linter (moai spec lint)"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 4 SPEC Writer
priority: P1 High
phase: "v3.0.0 — Phase 7 — Extension"
module: "internal/spec/, cmd/moai/spec.go"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-SPC-001
related_gap: []
related_problem: []
related_pattern:
  - X-1
related_principle:
  - P1
  - P12
related_theme: "Layer 2: SPEC & TAG"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "v3r2, spec, linter, ears-compliance, dag-validation, coverage"
---

# SPEC-V3R2-SPC-003: SPEC linter (moai spec lint)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | Wave 4 SPEC Writer | Initial draft |

---

## 1. Goal (목적)

Provide a deterministic, CI-integrable `moai spec lint` subcommand that validates every SPEC document under `.moai/specs/SPEC-*/spec.md` against the EARS format, the hierarchical acceptance schema (SPEC-V3R2-SPC-001), the frontmatter schema (CON-001 zone registry), and the inter-SPEC dependency DAG (no cycles).

Today, SPEC quality depends on author vigilance. Flat acceptance criteria, missing REQ→AC mappings, dangling dependencies, and cyclic `dependencies:` graphs pass silently because no automated gate rejects them. This SPEC establishes the gate.

`moai spec lint` exits non-zero on any violation so that CI and pre-commit hooks can block merges. Violations are grouped by severity (error / warning / info) and include file:line pointers.

## 2. Scope (범위)

### 2.1 In Scope

- `moai spec lint` CLI subcommand, operating on a single file (`moai spec lint .moai/specs/SPEC-X/spec.md`) or the full tree (`moai spec lint` with no args).
- EARS compliance checks: every requirement uses exactly one modality keyword (`SHALL` paired with WHEN/WHILE/WHERE/IF/Ubiquitous form), is testable, and has a unique `REQ-<DOMAIN>-<NNN>-<NNN>` ID.
- REQ ID uniqueness within a SPEC and recommended uniqueness across the project.
- AC→REQ coverage ≥ 100%: every REQ must be referenced by at least one acceptance leaf's `(maps REQ-...)` tail.
- Frontmatter schema validation: required fields `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `dependencies`, `bc_id`, `lifecycle`, `tags`; types enforced (id matches `SPEC-V3R2-<DOMAIN>-<NNN>`; version matches semver; bc_id always an array).
- Dependency DAG validation: no cycles among `dependencies:` fields; each dependency is a valid existing SPEC ID.
- Hierarchical AC structural validation (delegated to SPC-001 parser invariants): max depth 3, unique IDs per depth, REQ-map tail on every leaf.
- Zone registry cross-reference: if SPEC contains `related_rule: [CONST-V3R2-...]`, those IDs must exist in the zone registry (SPEC-V3R2-CON-001).
- Exclusions section required: every spec.md must contain a `Scope → Out of Scope` subsection with at least one entry (FROZEN per skill moai-workflow-spec).
- Output modes: human-readable table (default), JSON (`--json`), SARIF (`--sarif`) for CI annotations.

### 2.2 Out of Scope

- @MX TAG validation (→ SPEC-V3R2-SPC-002).
- SPEC content quality (grammar, clarity, ambiguity detection) — too subjective for a linter.
- Running tests or verifying that acceptance criteria actually pass against implementation (that is evaluator-active's job).
- Automatic fix-ups (`moai spec lint --fix`) — intentionally deferred; linter only reports.
- Non-SPEC artifact validation (reports, docs, decision records under `.moai/decisions/`).

## 3. Environment (환경)

Current moai state (v2.13.2):

- SPEC documents live at `.moai/specs/SPEC-<ID>/spec.md`; each SPEC directory also contains `plan.md` and `acceptance.md` per moai-workflow-spec skill (but this SPC-003 only lints spec.md body).
- Existing flat SPECs in `.moai/design/v3-legacy/specs/` demonstrate the 10-section body format with flat acceptance.
- No linter exists today; SPEC correctness is enforced by manual review or ad-hoc agent self-checks.
- `go-playground/validator/v10` referenced by legacy SPEC-V3-SCH-001; available dependency.
- CLI pattern: existing `cmd/moai/` uses cobra (or equivalent) subcommand tree — `spec` subcommand tree should be added.

References: skill moai-workflow-spec (EARS modality table + required sections); pattern-library.md §X-1 (markdown + YAML frontmatter validation).

## 4. Assumptions (가정)

- Linter runs locally in <5s for a project with 50 SPECs. No network calls.
- SPEC-V3R2-SPC-001 parser is available and produces the canonical acceptance tree; linter consumes that tree.
- SPEC-V3R2-CON-001 zone registry is queryable in-process; linter looks up `CONST-V3R2-NNN` IDs there.
- Cycle detection on the dependency DAG uses standard Tarjan or DFS; manageable for <100 SPECs.
- Frontmatter YAML parser is `gopkg.in/yaml.v3` (already in moai module graph).
- Pre-commit hook integration is opt-in via `moai doctor spec --install-pre-commit`.

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-SPC-003-001: The `moai spec lint` subcommand SHALL accept zero or more SPEC file paths; with no args, it SHALL lint every `.moai/specs/SPEC-*/spec.md`.
- REQ-SPC-003-002: The linter SHALL exit with status 0 iff no errors are reported; warnings and info messages SHALL NOT affect exit status by default.
- REQ-SPC-003-003: Every requirement in a SPEC SHALL use exactly one of the EARS modality forms: Ubiquitous (`The [system] SHALL`), Event-driven (`WHEN ..., the [system] SHALL`), State-driven (`WHILE ..., the [system] SHALL`), Optional (`WHERE ..., the [system] SHALL`), Unwanted (`IF ..., THEN the [system] SHALL NOT`), Complex (combining State + Event or State + Complex).
- REQ-SPC-003-004: Every REQ ID SHALL match the regex `^REQ-[A-Z]{2,5}-\d{3}-\d{3}$` and SHALL be unique within its SPEC file.
- REQ-SPC-003-005: Every REQ in the SPEC SHALL appear in at least one leaf acceptance node's `(maps REQ-...)` tail; SPECs failing this SHALL be reported with error `CoverageIncomplete` naming uncovered REQ IDs.
- REQ-SPC-003-006: The linter SHALL validate that `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `dependencies`, `bc_id`, `lifecycle`, `tags` frontmatter fields are present with correct types.
- REQ-SPC-003-007: The linter SHALL validate that every `dependencies: [...]` entry is a syntactically-valid SPEC ID and that the referenced SPEC directory exists on disk.
- REQ-SPC-003-008: The linter SHALL compute the dependency DAG across all linted SPECs and SHALL detect cycles; any detected cycle is a `DependencyCycle` error naming the cycle participants.
- REQ-SPC-003-009: Every spec.md SHALL contain a `Scope → Out of Scope` subsection with at least one non-empty entry; violations are `MissingExclusions` errors.
- REQ-SPC-003-010: Every `related_rule: [CONST-V3R2-NNN]` frontmatter entry SHALL reference a Rule ID that exists in the CON-001 zone registry; dangling references are `DanglingRuleReference` warnings.

### 5.2 Event-driven

- REQ-SPC-003-020: WHEN `moai spec lint --json` is invoked, the linter SHALL emit a machine-readable JSON array of findings to stdout, each with `file`, `line`, `severity`, `code`, `message` fields.
- REQ-SPC-003-021: WHEN `moai spec lint --sarif` is invoked, the linter SHALL emit SARIF 2.1.0-conformant output for CI annotation tools.
- REQ-SPC-003-022: WHEN the linter encounters a SPEC file that cannot be parsed (malformed YAML or markdown), it SHALL emit `ParseFailure` error referencing the offending file and SHALL continue with remaining files.

### 5.3 State-driven

- REQ-SPC-003-030: WHILE the `--strict` flag is set, WARNING-level findings SHALL be promoted to ERROR and cause non-zero exit.
- REQ-SPC-003-031: WHILE two SPECs declare the same `id:` frontmatter value, the linter SHALL report `DuplicateSPECID` as an ERROR citing both file paths.

### 5.4 Optional

- REQ-SPC-003-040: WHERE a SPEC's frontmatter contains `lint.skip: [CODE1, CODE2]`, the linter SHALL suppress findings matching those codes for that SPEC only.
- REQ-SPC-003-041: WHERE `moai spec lint --format table` is specified, the default human-readable output is explicitly selected (redundant with default, useful for script clarity).

### 5.5 Complex

- REQ-SPC-003-050: WHILE a requirement's text starts with "WHEN" but does not contain "SHALL" in its response clause, the linter SHALL report `ModalityMalformed` error naming the REQ ID.
- REQ-SPC-003-051: WHEN a REQ ID is referenced by an AC tail but does NOT exist in the Requirements section, the linter SHALL report `UnmappedRequirement` error; ELSE if the REQ exists but no AC references it, `CoverageIncomplete` applies.
- REQ-SPC-003-052: IF a SPEC has `breaking: true` AND `bc_id: []` (empty array), THEN the linter SHALL report `BreakingChangeMissingID` error; IF `breaking: false` AND `bc_id` is non-empty, the linter SHALL report `OrphanBCID` warning.

## 6. Acceptance Criteria

- AC-SPC-003-01: Given a valid SPEC with 12 REQs and 12 ACs each mapping uniquely, When `moai spec lint` runs, Then exit status is 0 and no findings are reported. (maps REQ-SPC-003-001, REQ-SPC-003-002, REQ-SPC-003-005)
- AC-SPC-003-02: Given a SPEC with REQ-X-001-007 declared but no AC tail references it, When lint runs, Then `CoverageIncomplete` error is reported naming REQ-X-001-007. (maps REQ-SPC-003-005)
- AC-SPC-003-03: Given a requirement text "WHEN the user logs in, the system creates a session" (missing SHALL), When lint runs, Then `ModalityMalformed` error is reported. (maps REQ-SPC-003-003, REQ-SPC-003-050)
- AC-SPC-003-04: Given two SPECs A and B where A depends on B and B depends on A, When lint runs across both, Then `DependencyCycle` error is reported naming A and B. (maps REQ-SPC-003-008)
- AC-SPC-003-05: Given a SPEC with duplicate REQ ID `REQ-X-001-005` appearing twice, When lint runs, Then error `DuplicateREQID` is reported naming both locations. (maps REQ-SPC-003-004)
- AC-SPC-003-06: Given a SPEC missing the `Out of Scope` subsection, When lint runs, Then `MissingExclusions` error is reported. (maps REQ-SPC-003-009)
- AC-SPC-003-07: Given a SPEC with `dependencies: [SPEC-NONEXISTENT]`, When lint runs, Then `MissingDependency` error is reported naming the missing SPEC. (maps REQ-SPC-003-007)
- AC-SPC-003-08: Given a SPEC with `related_rule: [CONST-V3R2-999]` where 999 is not in the zone registry, When lint runs, Then `DanglingRuleReference` warning is reported. (maps REQ-SPC-003-010)
- AC-SPC-003-09: Given `moai spec lint --json`, When lint runs, Then stdout contains a valid JSON array of finding objects. (maps REQ-SPC-003-020)
- AC-SPC-003-10: Given `moai spec lint --sarif`, When lint runs, Then stdout is SARIF 2.1.0-conformant JSON. (maps REQ-SPC-003-021)
- AC-SPC-003-11: Given `moai spec lint --strict` and a SPEC with 0 errors and 2 warnings, When lint runs, Then exit status is non-zero. (maps REQ-SPC-003-030)
- AC-SPC-003-12: Given two SPECs declaring `id: SPEC-V3R2-X-001`, When lint runs across both, Then `DuplicateSPECID` error is reported. (maps REQ-SPC-003-031)
- AC-SPC-003-13: Given a SPEC with frontmatter `lint.skip: [DanglingRuleReference]`, When lint runs and a dangling rule exists, Then no finding is emitted for that code. (maps REQ-SPC-003-040)
- AC-SPC-003-14: Given a SPEC with `breaking: true` and `bc_id: []`, When lint runs, Then `BreakingChangeMissingID` error is reported. (maps REQ-SPC-003-052)
- AC-SPC-003-15: Given a malformed SPEC file with broken YAML frontmatter, When lint runs across the tree, Then `ParseFailure` error is reported for that file and the linter continues with other files. (maps REQ-SPC-003-022)
- AC-SPC-003-16: Given a SPEC with hierarchical acceptance produced by SPC-001 parser (e.g., AC-X-01 with three children), When lint runs, Then coverage counts parent-level mappings as covering any REQ referenced by at least one leaf child. (maps REQ-SPC-003-005)

## 7. Constraints (제약)

- Linter performance: <5s for 50 SPECs on a dev laptop; <15s for 200 SPECs.
- Exit codes: 0 = success, 1 = errors, 2 = linter crash, 3 = invalid arguments.
- Output size: JSON and SARIF outputs must remain <10MB for a 200-SPEC project.
- Dependencies: Go stdlib + `gopkg.in/yaml.v3` + `go-playground/validator/v10` (already in moai deps).
- No network or filesystem writes; linter is read-only.
- Integration: `moai doctor spec --install-pre-commit` installs a git pre-commit hook that runs `moai spec lint --strict` on changed SPEC files.

## 8. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| False positive on EARS modality regex (natural language variants) | MEDIUM | MEDIUM | Conservative patterns with allowlist; lint.skip escape hatch per-SPEC |
| Cycle detection is slow on large dependency graphs | LOW | LOW | Tarjan O(V+E); budget accommodates 200 SPECs |
| Duplicate SPEC IDs across forks/worktrees | MEDIUM | MEDIUM | DuplicateSPECID error surfaces early; merge resolves |
| SARIF output format churn (version updates) | LOW | LOW | Pin SARIF 2.1.0; upgrade is a separate SPEC |
| Frontmatter schema drift vs new v3 SPECs | MEDIUM | MEDIUM | Schema defined in a single Go struct; CI covers canonical fixtures |

## 9. Dependencies

### 9.1 Blocked by

- SPEC-V3R2-CON-001 (zone registry for `CONST-V3R2-NNN` cross-reference validation).
- SPEC-V3R2-SPC-001 (hierarchical acceptance parser must produce canonical tree).

### 9.2 Blocks

- CI gates for merging PRs that touch SPECs.
- SPEC-V3R2-MIG-001 (migration tool runs linter post-migration to verify integrity).

### 9.3 Related

- SPEC-V3R2-SPC-002 (@MX TAG linting happens via `/moai mx --verify`, not this linter).
- skill moai-workflow-spec — authoritative source for required body sections.

## 10. Traceability

- Theme: Layer 2 SPEC & TAG (master-v3 §4).
- Principles: P1 SPEC as Constitutional Contract (linter enforces the contract); P12 File-First Primitives (linter reads markdown, no framework dependency).
- Patterns: X-1 Markdown + YAML Frontmatter validation (pattern-library.md §X-1).
- Wave 1 sources: R1 §7 MetaGPT (typed intermediate artifacts must be validated), R2 §A Top-5 Pattern 3 (markdown + YAML universal).
- Wave 2 sources: design-principles.md §P1, pattern-library.md §X-1.
