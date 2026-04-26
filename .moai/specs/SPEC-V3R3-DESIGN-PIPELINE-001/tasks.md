# SPEC-V3R3-DESIGN-PIPELINE-001 — Task Breakdown

This task list decomposes the 5-phase implementation plan into discrete, ownership-assigned, dependency-ordered work items. Tasks are priority-tagged (P0 Critical, P1 High, P2 Medium); no time estimates per `agent-common-protocol.md` Time Estimation rule. Each task references its REQ-IDs and AC-IDs and notes its file ownership for parallel/sequential execution decisions.

## Legend

- **P0 Critical**: Blocking; must complete before dependent tasks start.
- **P1 High**: Required for SPEC completion; no compromise.
- **P2 Medium**: Quality / documentation / polish; should complete before release.
- **Owner**: Suggested agent (`expert-backend`, `expert-frontend`, `manager-tdd`, `manager-docs`, `plan-auditor`).
- **Dep**: Task IDs this depends on.
- **Files**: Primary files modified (Template-First mirror always implied).

---

## Phase 1 — Skill Body Updates (P1 High)

### T1-01 — Extend `moai-workflow-design-import/SKILL.md` body with 3-path entry sections

- **REQ-IDs**: REQ-DPL-001, REQ-DPL-002, REQ-DPL-003
- **AC-IDs**: AC-DPL-01
- **Owner**: `expert-frontend`
- **Dep**: —
- **Files**:
  - `.claude/skills/moai-workflow-design-import/SKILL.md` (body update)
  - `internal/template/templates/.claude/skills/moai-workflow-design-import/SKILL.md` (mirror)
- **Subtasks**:
  - Preserve existing Path A section verbatim (no behavioral change to handoff bundle import).
  - Add Path B1 section: meta-harness invocation, `my-harness-figma-extractor` generation contract, Figma credential reference convention.
  - Add Path B2 section: meta-harness invocation, `my-harness-pencil-mcp` generation contract, `.pen` file location convention.
  - Cross-reference SPEC-V3R3-HARNESS-001 as the meta-harness provider.
  - Cross-reference reserved file paths per design constitution §3.2.
  - Verify body length under 500 lines (per `coding-standards.md`).
- **Exit gate**: `make build && go test ./internal/template/...` green.

### T1-02 — Extend `moai-design-system/SKILL.md` body with DTCG 2025.10 reference

- **REQ-IDs**: REQ-DPL-004 (referenced from skill body for `expert-frontend` invocation guidance)
- **AC-IDs**: AC-DPL-02, AC-DPL-06
- **Owner**: `expert-frontend`
- **Dep**: —
- **Files**:
  - `.claude/skills/moai-design-system/SKILL.md` (body update)
  - `internal/template/templates/.claude/skills/moai-design-system/SKILL.md` (mirror)
- **Subtasks**:
  - Add DTCG 2025.10 section: spec snapshot reference (`internal/design/dtcg/SPEC.md`), supported categories (color, dimension, font, fontFamily, fontWeight, duration, cubicBezier, number, strokeStyle, border, transition, shadow, gradient, typography).
  - Add validator invocation guidance: `expert-frontend` MUST call `internal/design/dtcg.Validate` before code generation; failure surfaces structured error.
  - Cross-reference brand context priority per design constitution §3.1.
  - Verify body length under 500 lines.
- **Exit gate**: skill body parses cleanly via `internal/skill/parse_test.go`.

### T1-03 — Template-First mirror verification (Phase 1)

- **REQ-IDs**: REQ-DPL-012
- **AC-IDs**: AC-DPL-07
- **Owner**: `manager-tdd`
- **Dep**: T1-01, T1-02
- **Files**: `internal/template/templates/.claude/skills/moai-{workflow-design-import,design-system}/SKILL.md`
- **Subtasks**:
  - Run `make build` to regenerate `internal/template/embedded.go`.
  - `diff -rq` between working copy and template; expect zero diff.
  - Update `internal/template/commands_audit_test.go` if any new path is introduced (none expected in Phase 1).
- **Exit gate**: `go test ./internal/template/...` green; embedded files identical to source.

---

## Phase 2 — Workflow Routing (P1 High)

### T2-01 — Implement `/moai design` workflow path branching in `design.md`

- **REQ-IDs**: REQ-DPL-005, REQ-DPL-008
- **AC-IDs**: AC-DPL-03
- **Owner**: `expert-frontend`
- **Dep**: T1-01 (skill body must reference workflow correctly)
- **Files**:
  - `.claude/skills/moai/workflows/design.md` (workflow body)
  - `internal/template/templates/.claude/skills/moai/workflows/design.md` (mirror)
- **Subtasks**:
  - Implement AskUserQuestion path selector with three options:
    - Option 1: `Path A (Claude Design)` marked `(권장)`, description: "Claude Design handoff bundle import (most stable, recommended for new users)".
    - Option 2: `Path B1 (Figma)`, description: "Figma file via dynamic figma-extractor (requires Figma credentials)".
    - Option 3: `Path B2 (Pencil)`, description: "Pencil .pen files via dynamic pencil-mcp (requires .pen files in project)".
  - Implement workflow branching: Path A → existing import, Path B1 → meta-harness `figma-extractor` generation, Path B2 → meta-harness `pencil-mcp` generation.
  - Implement `.moai/design/path-selection.json` writer (audit trail).
  - Implement brand-context loader from `.moai/project/brand/` (constitution §3.1 priority).
  - Implement brand-conflict warning surface (token vs visual-identity.md).
- **Exit gate**: workflow body under 500 lines; AskUserQuestion options pass max-4 limit; recommended-first rule enforced.

### T2-02 — Persist path-selection.json schema test

- **REQ-IDs**: REQ-DPL-005
- **AC-IDs**: AC-DPL-03
- **Owner**: `manager-tdd`
- **Dep**: T2-01
- **Files**:
  - `internal/design/pipeline/path_selection.go` (new)
  - `internal/design/pipeline/path_selection_test.go` (new)
- **Subtasks**:
  - Define `PathSelection` struct (path: A|B1|B2, brand_context_loaded: bool, spec_id, ts, session_id).
  - Implement JSON writer with stable field ordering.
  - Implement reader for idempotency (re-invocation surfaces previous selection).
  - Unit tests: positive write/read, malformed JSON, missing fields.
- **Exit gate**: `go test ./internal/design/pipeline/...` green; coverage ≥85%.

---

## Phase 3 — DTCG Validator (P0 Critical)

### T3-01 — Scaffold `internal/design/dtcg/` package + `SPEC.md` spec snapshot

- **REQ-IDs**: REQ-DPL-004
- **AC-IDs**: AC-DPL-02
- **Owner**: `expert-backend`
- **Dep**: —
- **Files**:
  - `internal/design/dtcg/dtcg.go` (package doc, `Validate` API stub)
  - `internal/design/dtcg/SPEC.md` (DTCG 2025.10 snapshot reference)
  - `internal/design/dtcg/errors.go` (`ValidationError`, `ValidationWarning`, `Report` types)
- **Subtasks**:
  - Define `Report{Valid bool, TokenCount int, Errors []ValidationError, Warnings []ValidationWarning, SpecVersion string}`.
  - Define `ValidationError{Path, Rule string, Got any, Want string}`.
  - Document DTCG 2025.10 snapshot URL + commit hash in `SPEC.md`.
  - Add package-level Godoc.
- **Exit gate**: package compiles; `go vet ./internal/design/dtcg/...` clean.

### T3-02 — Implement `color` category validator (RED → GREEN)

- **REQ-IDs**: REQ-DPL-004
- **AC-IDs**: AC-DPL-02
- **Owner**: `manager-tdd`
- **Dep**: T3-01
- **Files**:
  - `internal/design/dtcg/categories/color.go`
  - `internal/design/dtcg/categories/color_test.go`
- **Subtasks**:
  - RED: write failing tests for hex (#fff, #ffffff, #ffff, #ffffffff), rgb, hsl, named colors, invalid inputs.
  - GREEN: implement minimal validator that passes all tests.
  - REFACTOR: extract common color-parsing helper if duplication emerges.
- **Exit gate**: 100% category coverage; positive + negative cases ≥6 each.

### T3-03 — Implement `dimension` category validator

- **REQ-IDs**: REQ-DPL-004
- **AC-IDs**: AC-DPL-02
- **Owner**: `manager-tdd`
- **Dep**: T3-01
- **Files**:
  - `internal/design/dtcg/categories/dimension.go`
  - `internal/design/dtcg/categories/dimension_test.go`
- **Subtasks**:
  - RED: tests for px, rem, em, %, unitless, invalid units.
  - GREEN: implement.
  - REFACTOR: share unit-parsing helper across dimension and duration if applicable.
- **Exit gate**: 100% category coverage.

### T3-04 — Implement `font` + `fontFamily` + `fontWeight` category validators

- **REQ-IDs**: REQ-DPL-004
- **AC-IDs**: AC-DPL-02
- **Owner**: `manager-tdd`
- **Dep**: T3-01
- **Files**:
  - `internal/design/dtcg/categories/font.go`, `font_family.go`, `font_weight.go`
  - corresponding `_test.go` files
- **Subtasks**:
  - RED: tests for each category with positive/negative cases (composite font reference, family list, numeric weight, named weight).
  - GREEN: implement.
- **Exit gate**: 3 categories validated; alias resolution stub in place for typography composite.

### T3-05 — Implement `typography` composite category (alias resolution)

- **REQ-IDs**: REQ-DPL-004
- **AC-IDs**: AC-DPL-02
- **Owner**: `manager-tdd`
- **Dep**: T3-04
- **Files**:
  - `internal/design/dtcg/categories/typography.go`
  - `internal/design/dtcg/categories/typography_test.go`
  - `internal/design/dtcg/alias.go` (alias resolver, shared)
- **Subtasks**:
  - RED: tests for typography token referencing fontFamily + fontWeight aliases; cyclic alias detection.
  - GREEN: implement alias resolver with cycle detection.
  - REFACTOR: ensure alias resolver is reusable for other composite types (transition, gradient).
- **Exit gate**: typography validation passes; cyclic alias rejected with clear error.

### T3-06 — Implement `shadow` + `border` category validators

- **REQ-IDs**: REQ-DPL-004
- **AC-IDs**: AC-DPL-02
- **Owner**: `manager-tdd`
- **Dep**: T3-02 (color), T3-03 (dimension)
- **Files**:
  - `internal/design/dtcg/categories/shadow.go`, `border.go`
  - corresponding `_test.go` files
- **Subtasks**:
  - RED: tests for single-layer shadow, multi-layer shadow, border (width + style + color composite), invalid inputs.
  - GREEN: implement (composes color + dimension validators).
- **Exit gate**: 6 minimum categories complete (color, dimension, font, typography, shadow, border).

### T3-07 — Implement remaining DTCG 2025.10 categories (P2 Medium)

- **REQ-IDs**: REQ-DPL-004
- **AC-IDs**: AC-DPL-02
- **Owner**: `manager-tdd`
- **Dep**: T3-06
- **Files**:
  - `internal/design/dtcg/categories/{duration,cubic_bezier,number,stroke_style,transition,gradient}.go`
  - corresponding `_test.go` files
- **Subtasks**:
  - Implement remaining 6 categories per DTCG 2025.10 spec.
  - Tests for each.
- **Exit gate**: full DTCG 2025.10 category coverage achieved (informational; AC-DPL-02 only requires 6).

### T3-08 — Implement FROZEN zone matcher

- **REQ-IDs**: REQ-DPL-011
- **AC-IDs**: AC-DPL-05
- **Owner**: `expert-backend`
- **Dep**: T3-01
- **Files**:
  - `internal/design/dtcg/frozen_guard.go`
  - `internal/design/dtcg/frozen_guard_test.go`
- **Subtasks**:
  - Hardcode FROZEN zone path + section list (constitution §2, §3.1, §3.2, §3.3, §5, §11, §12; brand context files).
  - Implement `IsFrozen(path string) bool` with prefix + section-heading-hash match.
  - Implement `BlockWrite(path, reason string) error` returning structured violation error.
  - Tests: 5+ known-violation paths rejected; allowed paths (e.g., §4 Phase Contracts row range) accepted; no config bypass exists.
- **Exit gate**: `go test ./internal/design/dtcg/...` green; FROZEN guard cannot be bypassed by any code path.

### T3-09 — Implement top-level `Validate` API + integration with categories

- **REQ-IDs**: REQ-DPL-004, REQ-DPL-010
- **AC-IDs**: AC-DPL-02, AC-DPL-06
- **Owner**: `expert-backend`
- **Dep**: T3-02, T3-03, T3-04, T3-05, T3-06, T3-08
- **Files**:
  - `internal/design/dtcg/validator.go` (replaces `dtcg.go` stub)
  - `internal/design/dtcg/validator_test.go`
- **Subtasks**:
  - Implement `Validate(tokens map[string]any) (*Report, error)` dispatching to category validators.
  - Integrate FROZEN guard pre-check.
  - Implement structured error aggregation (offending token path, rule violated).
  - Performance: <100ms for ≤500-token sets (benchmarked).
  - Coverage: ≥85% per `quality.yaml`.
- **Exit gate**: `go test -race -cover ./internal/design/dtcg/...` green on macOS, Linux, Windows; coverage ≥85%.

### T3-10 — DTCG conformance test data

- **REQ-IDs**: REQ-DPL-004
- **AC-IDs**: AC-DPL-02
- **Owner**: `expert-backend`
- **Dep**: T3-09
- **Files**:
  - `internal/design/dtcg/testdata/positive/*.json`
  - `internal/design/dtcg/testdata/negative/*.json`
- **Subtasks**:
  - Curate ≥6 positive token files (one per minimum category).
  - Curate ≥6 negative token files (each violates a specific rule).
  - Wire into `validator_test.go` table-driven test.
- **Exit gate**: all positive cases validate; all negative cases produce expected `ValidationError`.

---

## Phase 4 — expert-frontend Integration + Path A Validator Hookup (P1 High)

### T4-01 — Update `expert-frontend` agent prompt to invoke DTCG validator

- **REQ-IDs**: REQ-DPL-010
- **AC-IDs**: AC-DPL-06
- **Owner**: `expert-backend`
- **Dep**: T3-09
- **Files**:
  - `.claude/agents/moai/expert-frontend.md` (prompt update)
  - `internal/template/templates/.claude/agents/moai/expert-frontend.md` (mirror)
- **Subtasks**:
  - Insert validator-invocation step before any code generation step.
  - Document failure handling: structured error returned to orchestrator; code generation blocked.
  - Cross-reference `internal/design/dtcg/` package.
- **Exit gate**: agent prompt under existing length limits; Template-First mirror consistent.

### T4-02 — Path A integration test (existing import + validator gate)

- **REQ-IDs**: REQ-DPL-001, REQ-DPL-010
- **AC-IDs**: AC-DPL-01, AC-DPL-06
- **Owner**: `manager-tdd`
- **Dep**: T4-01, T2-01
- **Files**:
  - `internal/design/pipeline/path_a_test.go`
  - `internal/design/pipeline/testdata/path_a_handoff/`
- **Subtasks**:
  - Set up `t.TempDir()` with sample Claude Design handoff bundle.
  - Run existing import flow → verify `tokens.json`, `components.json`, `assets/`, `import-warnings.json` produced (golden output).
  - Verify validator runs on `tokens.json` (no error for valid sample).
  - Negative path: corrupt `tokens.json` → validator blocks downstream consumption.
- **Exit gate**: integration test green on 3 OS targets.

### T4-03 — Path B1 integration test (stub mode)

- **REQ-IDs**: REQ-DPL-002, REQ-DPL-010
- **AC-IDs**: AC-DPL-01
- **Owner**: `manager-tdd`
- **Dep**: T4-01, T2-01
- **Files**:
  - `internal/design/pipeline/path_b1_test.go`
- **Subtasks**:
  - Stub meta-harness skill produces `my-harness-figma-extractor/SKILL.md` (depends on SPEC-V3R3-HARNESS-001 stub).
  - Verify generated skill frontmatter contains Figma file ID + page selectors + credential reference.
  - Verify generated `tokens.json` validates against DTCG 2025.10.
  - Verify `.moai/design/path-selection.json` records `path: "B1"`.
- **Exit gate**: integration test green; HARNESS-001 stub mode sufficient.

### T4-04 — Path B2 integration test (stub mode)

- **REQ-IDs**: REQ-DPL-003, REQ-DPL-010
- **AC-IDs**: AC-DPL-01, AC-DPL-04
- **Owner**: `manager-tdd`
- **Dep**: T4-01, T2-01
- **Files**:
  - `internal/design/pipeline/path_b2_test.go`
- **Subtasks**:
  - Stub meta-harness produces `my-harness-pencil-mcp/SKILL.md`.
  - Verify generated skill frontmatter contains `.pen` file paths + MCP server endpoint.
  - Verify generated `tokens.json` validates against DTCG 2025.10.
  - Verify legacy `moai-workflow-pencil-integration` references absent.
- **Exit gate**: integration test green; pencil-integration removal verified.

### T4-05 — Brand-conflict warning surface

- **REQ-IDs**: REQ-DPL-009
- **AC-IDs**: AC-DPL-05 (Scenario A)
- **Owner**: `manager-tdd`
- **Dep**: T4-02
- **Files**:
  - `internal/design/pipeline/brand_conflict.go`
  - `internal/design/pipeline/brand_conflict_test.go`
- **Subtasks**:
  - Implement brand-vs-token comparator (read `.moai/project/brand/visual-identity.md`, compare against `tokens.json`).
  - Surface conflicts as `ValidationWarning` with `category: "brand-conflict"`.
  - Integration with workflow AskUserQuestion (workflow surfaces warning to user).
  - Tests: matching colors → no warning; mismatching colors → warning with both values.
- **Exit gate**: brand-context priority enforced; warnings surface correctly.

### T4-06 — GAN loop score-variance benchmark

- **REQ-IDs**: REQ-DPL-010
- **AC-IDs**: AC-DPL-06
- **Owner**: `expert-backend`
- **Dep**: T4-02
- **Files**:
  - `internal/design/pipeline/gan_variance_test.go` (benchmark test)
- **Subtasks**:
  - Replay a baseline GAN-loop session (pre-SPEC) and capture evaluator-active score.
  - Replay the same session post-SPEC with validator gate enabled.
  - Assert `|delta| ≤ 0.05` per design constitution §11 improvement_threshold.
  - Performance assertion: validator <100ms for typical token sets.
- **Exit gate**: variance within ±0.05; performance within budget.

---

## Phase 5 — Constitution §4 Amendment + Pencil-Integration Removal (P1 High)

### T5-01 — Apply additive amendment to `.claude/rules/moai/design/constitution.md`

- **REQ-IDs**: REQ-DPL-007
- **AC-IDs**: AC-DPL-07 (Scenario A)
- **Owner**: `manager-docs`
- **Dep**: T1-01, T1-02, T2-01 (so the amendment matches actual implementation)
- **Files**:
  - `.claude/rules/moai/design/constitution.md`
  - `internal/template/templates/.claude/rules/moai/design/constitution.md` (mirror)
- **Subtasks**:
  - Append HISTORY entry: `2026-04-26 (SPEC-V3R3-DESIGN-PIPELINE-001): §4 Phase Contracts table extended with Path B1 (figma-extractor) and Path B2 (pencil-mcp) rows. Version 3.3.0 → 3.4.0.`
  - Add 2 new rows to §4 Phase Contracts table:
    - `figma-extractor (Path B1) | BRIEF + Figma file ID + page selectors | tokens.json + components.json | Path B1`
    - `pencil-mcp (Path B2) | BRIEF + .pen files | tokens.json + components.json | Path B2`
  - Bump footer version: `Version: 3.3.0` → `Version: 3.4.0`.
  - Update footer "Last Updated: 2026-04-26".
  - DO NOT modify any FROZEN section (§2, §3.1, §3.2, §3.3, §5, §11, §12).
- **Exit gate**: plan-auditor diff check — only HISTORY + 2 §4 rows + version + last-updated changed.

### T5-02 — Verify zone-registry CONST-V3R2-068 unaffected

- **REQ-IDs**: REQ-DPL-007
- **AC-IDs**: AC-DPL-07 (Scenario B)
- **Owner**: `plan-auditor`
- **Dep**: T5-01
- **Files**: `.claude/rules/moai/core/zone-registry.md` (read-only)
- **Subtasks**:
  - Run `moai constitution list --zone frozen --format json` before T5-01 → save baseline.
  - Run after T5-01 → compare to baseline.
  - Expect zero diff (CONST-V3R2-051..072 untouched; §4 is EVOLVABLE per design constitution §2).
- **Exit gate**: zone-registry frozen-zone enumeration identical pre/post amendment.

### T5-03 — Coordinate `moai-workflow-pencil-integration` removal with SPEC-V3R3-HARNESS-001

- **REQ-IDs**: REQ-DPL-006
- **AC-IDs**: AC-DPL-04
- **Owner**: `expert-backend`
- **Dep**: SPEC-V3R3-HARNESS-001 BC-V3R3-007 task complete
- **Files**:
  - `.claude/skills/moai/SKILL.md` (routing table — remove pencil-integration if present)
  - `.claude/agents/moai/expert-frontend.md` (prompt — remove pencil-integration reference if present)
  - `.claude/skills/moai/workflows/*.md` (any workflow referencing pencil-integration)
- **Subtasks**:
  - `grep -r "moai-workflow-pencil-integration" .claude/` → list all references.
  - Replace each reference with Path B2 / `my-harness-pencil-mcp` references.
  - Verify HARNESS-001 16-skill removal migration archives the legacy skill correctly.
- **Exit gate**: `grep -r "moai-workflow-pencil-integration" .claude/ --exclude-dir=archive` returns zero matches.

### T5-04 — Template-First mirror verification (Phase 5)

- **REQ-IDs**: REQ-DPL-012
- **AC-IDs**: AC-DPL-07 (Scenario C)
- **Owner**: `manager-tdd`
- **Dep**: T5-01, T5-03
- **Files**: All Template-First mirrors under `internal/template/templates/`
- **Subtasks**:
  - Run `make build`.
  - `diff -rq` between working copy and template for all modified files:
    - `.claude/skills/moai-workflow-design-import/SKILL.md`
    - `.claude/skills/moai-design-system/SKILL.md`
    - `.claude/skills/moai/workflows/design.md`
    - `.claude/rules/moai/design/constitution.md`
    - `.claude/agents/moai/expert-frontend.md`
  - Expect zero diff for each pair.
- **Exit gate**: all Template-First mirrors consistent; `make build && go test ./...` green.

---

## Phase 6 — Documentation + Release Coordination (P2 Medium)

### T6-01 — Update CHANGELOG.md with v2.17.0 entry

- **REQ-IDs**: —
- **AC-IDs**: —
- **Owner**: `manager-docs`
- **Dep**: All Phase 1–5 tasks complete
- **Files**: `CHANGELOG.md`
- **Subtasks**:
  - Add v2.17.0 section (or extend existing draft) with:
    - feat(design): SPEC-V3R3-DESIGN-PIPELINE-001 — Hybrid Design Pipeline (Path A/B1/B2 + DTCG 2025.10 validator).
    - BREAKING (BC-V3R3-007 coordinated): `moai-workflow-pencil-integration` removed; replaced by Path B2 dynamic generation.
- **Exit gate**: CHANGELOG entry follows existing format; bilingual (English + Korean) per CLAUDE.local.md §17 guidance.

### T6-02 — Update docs-site (4-locale) design pipeline reference

- **REQ-IDs**: —
- **AC-IDs**: —
- **Owner**: `manager-docs`
- **Dep**: T6-01
- **Files**:
  - `docs-site/content/ko/v2/design/pipeline.md` (canonical source)
  - `docs-site/content/en/v2/design/pipeline.md`
  - `docs-site/content/ja/v2/design/pipeline.md`
  - `docs-site/content/zh/v2/design/pipeline.md`
- **Subtasks**:
  - Per CLAUDE.local.md §17.3: ko canonical first, then en/ja/zh in same PR (or 48h grace if >5000 words).
  - Document Path A/B1/B2 selection, DTCG 2025.10 validator, brand context priority, FROZEN zone protection.
  - Per §17.6: `cd docs-site && hugo --minify` succeeds; Vercel preview verified.
- **Exit gate**: 4 locale files synced; hugo build green; preview verified.

### T6-03 — Plan-auditor sign-off on §4 amendment

- **REQ-IDs**: REQ-DPL-007
- **AC-IDs**: AC-DPL-07
- **Owner**: `plan-auditor`
- **Dep**: T5-01, T5-02
- **Files**: review report at `.moai/reports/plan-audit/SPEC-V3R3-DESIGN-PIPELINE-001-2026-04-26.md`
- **Subtasks**:
  - Verify §4 diff is additive-only (no row deletion, no column rename, no FROZEN section modification).
  - Verify zone-registry CONST-V3R2-068 unchanged.
  - Verify HISTORY entry follows existing format.
  - Verify version bump correct (3.3.0 → 3.4.0).
  - Issue PASS/FAIL verdict.
- **Exit gate**: PASS verdict recorded.

---

## Cross-Cutting Tasks

### TX-01 — TRUST 5 quality gate sweep

- **Owner**: `manager-quality`
- **Dep**: All Phase 1–5 tasks complete
- **Subtasks**:
  - **Tested**: ≥85% coverage in `internal/design/dtcg/`; ≥3 integration tests; FROZEN-guard regression suite.
  - **Readable**: Validator API ≤10 exported symbols; per-category files ≤200 lines.
  - **Unified**: `gofmt`, `golangci-lint run` clean.
  - **Secured**: No network egress from `internal/design/dtcg/`; FROZEN guard cannot be bypassed; no credential storage.
  - **Trackable**: Conventional commits with Korean body; SPEC-V3R3-DESIGN-PIPELINE-001 references; CHANGELOG entry.
- **Exit gate**: all 5 dimensions pass.

### TX-02 — Cross-platform CI green

- **Owner**: `manager-tdd`
- **Dep**: All Phase 1–5 tasks complete
- **Subtasks**:
  - `go test -race ./...` green on ubuntu-latest, macos-latest, windows-latest.
  - `make build` green on all 3 OS.
- **Exit gate**: CI all-green.

### TX-03 — SPEC status update

- **Owner**: `manager-docs`
- **Dep**: TX-01, TX-02, T6-03
- **Files**:
  - `.moai/specs/SPEC-V3R3-DESIGN-PIPELINE-001/spec.md` (frontmatter `status: draft` → `completed`)
- **Subtasks**:
  - Update `status` field to `completed`.
  - Update `updated_at` to merge date.
  - Append HISTORY entry: completion timestamp + merged_pr + merged_commit.
- **Exit gate**: SPEC marked completed; ready for v2.17.0 release prep.

---

## Task Summary

| Phase | Task Count | Priority Mix |
|-------|------------|--------------|
| Phase 1 | 3 | 3× P1 |
| Phase 2 | 2 | 2× P1 |
| Phase 3 | 10 | 1× P0 (TX), 8× P1, 1× P2 (T3-07 remaining categories) |
| Phase 4 | 6 | 6× P1 |
| Phase 5 | 4 | 4× P1 |
| Phase 6 | 3 | 3× P2 |
| Cross-cutting | 3 | 3× P1 |
| **Total** | **31** | — |

## Execution Notes

- Phase 1, 2, 3 can proceed in parallel (no inter-phase dependencies until Phase 4 integration).
- Phase 4 requires Phase 1, 2, 3 outputs — sequential.
- Phase 5 amendment can proceed in parallel with Phase 4 verification (different files, no overlap).
- Phase 6 documentation strictly after Phase 5 completion to capture final state.
- T5-03 (pencil-integration removal) requires SPEC-V3R3-HARNESS-001 BC-V3R3-007 task complete — coordinate sequencing with HARNESS-001 owner.
- All file modifications under `.claude/skills/`, `.claude/agents/`, `.claude/rules/`, `.moai/config/sections/` MUST have Template-First mirrors per CLAUDE.local.md §2 [HARD] rule.
