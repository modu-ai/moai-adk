# SPEC-V3R3-DESIGN-PIPELINE-001 — Implementation Plan

## 1. Overview

This plan delivers the hybrid design pipeline (Path A / B1 / B2) and the DTCG 2025.10 validator across five phases. Each phase is priority-ordered (no time estimates per `agent-common-protocol.md` Time Estimation rule). Phases are gated by SPEC-V3R3-HARNESS-001 completion (Path B1/B2 require the meta-harness skill) and by inter-phase artifact handoff. Constitution §4 amendment is the lowest-risk change and is sequenced last to allow earlier phases to surface any unforeseen amendments.

## 2. Architectural Approach

### 2.1 Component Map

```
/moai design (workflow entry)
   └─> .claude/skills/moai/workflows/design.md  (Phase 4)
        ├── AskUserQuestion: Path A | B1 | B2
        ├── persist .moai/design/path-selection.json
        │
        ├── Path A: existing moai-workflow-design-import flow (preserved)
        │     └─> .moai/design/{tokens,components,...}
        │           └─> internal/design/dtcg/Validate (Phase 3)
        │                 └─> expert-frontend (token consumer)
        │
        ├── Path B1: meta-harness ─> .claude/skills/my-harness-figma-extractor/
        │     └─> figma-extractor produces tokens.json
        │           └─> internal/design/dtcg/Validate (Phase 3)
        │                 └─> expert-frontend
        │
        └── Path B2: meta-harness ─> .claude/skills/my-harness-pencil-mcp/
              └─> pencil-mcp produces tokens.json
                    └─> internal/design/dtcg/Validate (Phase 3)
                          └─> expert-frontend

DTCG Validator: internal/design/dtcg/   (Phase 3)
   ├── validator.go        (Validate API)
   ├── categories/         (color, dimension, font, typography, shadow, border, ...)
   ├── errors.go           (structured error reports)
   └── dtcg_test.go        (unit tests, ≥6 categories)

Skill Body Updates: (Phase 1)
   ├── .claude/skills/moai-workflow-design-import/SKILL.md   (3-path entry)
   ├── .claude/skills/moai-design-system/SKILL.md            (DTCG 2025.10 reference)

Constitution Amendment (Phase 5):
   └─> .claude/rules/moai/design/constitution.md  §4 additive rows + HISTORY
```

### 2.2 Data Flow

1. User invokes `/moai design`.
2. Workflow (Phase 4) loads brand context and presents AskUserQuestion (Path A first, marked "(권장)").
3. Selection persisted to `.moai/design/path-selection.json`.
4. Path A: existing import flow executes; tokens.json produced.
5. Path B1: meta-harness skill (from SPEC-V3R3-HARNESS-001) generates `my-harness-figma-extractor`; extractor runs; tokens.json produced.
6. Path B2: meta-harness generates `my-harness-pencil-mcp`; extractor runs; tokens.json produced.
7. `expert-frontend` invokes DTCG validator before code generation; invalid tokens block downstream work and surface error to orchestrator.
8. Brand context conflicts surface as warnings via AskUserQuestion; user resolves per-conflict.

### 2.3 Boundary Enforcement

The FROZEN zone matcher in Phase 3 is implemented in `internal/design/dtcg/frozen_guard.go` and uses path-prefix + section-heading hash to detect attempted modifications to constitution §2, §3.1, §3.2, §3.3, §5, §11, §12. Any pipeline write outside the §4 Phase Contracts row range fails the build with a structured error.

## 3. Phased Implementation

### Phase 1 — Skill Body Updates (Priority: P1 High)

**Goal**: Extend `moai-workflow-design-import/SKILL.md` and `moai-design-system/SKILL.md` bodies with Path B1/B2 routing guidance and DTCG 2025.10 reference. No behavioral change to Path A. No new code yet.

**Deliverables**:
- `.claude/skills/moai-workflow-design-import/SKILL.md` — Body extended with three sections: Path A (existing, preserved verbatim), Path B1 (Figma via meta-harness), Path B2 (Pencil via meta-harness). Reserved file paths preserved per constitution §3.2.
- `.claude/skills/moai-design-system/SKILL.md` — Body extended with DTCG 2025.10 token specification reference (categories, validation rules, expected schema), validator invocation guidance for `expert-frontend`.
- Template-First mirrors under `internal/template/templates/.claude/skills/moai-workflow-design-import/` and `.../moai-design-system/`.
- Skill frontmatter validated (no schema change; bodies only).

**REQ-IDs covered**: REQ-DPL-001 (Path A preservation in body), REQ-DPL-002 (Path B1 routing in body), REQ-DPL-003 (Path B2 routing in body), REQ-DPL-012 (Template-First).

**Exit gate**: `make build && go test ./internal/template/...` passes; both skill body line counts under 500; bodies parse-valid via `internal/skill/parse_test.go`.

### Phase 2 — Workflow Routing (Priority: P1 High)

**Goal**: Implement `/moai design` AskUserQuestion path branching and persist selection. Path B1/B2 stub-route to meta-harness invocation (the meta-harness skill is supplied by SPEC-V3R3-HARNESS-001).

**Deliverables**:
- `.claude/skills/moai/workflows/design.md` — New routing section with three branches (Path A / B1 / B2), AskUserQuestion prompt template, `.moai/design/path-selection.json` writer, brand-context loader.
- Template-First mirror under `internal/template/templates/.claude/skills/moai/workflows/design.md`.
- AskUserQuestion option ordering: Path A first with "(권장)", Path B1 second, Path B2 third — each with prerequisite description.
- Brand-context conflict warning section.

**REQ-IDs covered**: REQ-DPL-005 (workflow branching), REQ-DPL-008 (AskUserQuestion path selection), REQ-DPL-009 (brand context priority), REQ-DPL-012 (Template-First).

**Exit gate**: Workflow file under 500 lines; brand-context loader cross-references `.moai/project/brand/` per constitution §3.1; AskUserQuestion options pass option-count limit (max 4) and recommended-first rule.

### Phase 3 — DTCG Validator (Priority: P0 Critical)

**Goal**: Implement Go-based DTCG 2025.10 token validator covering ≥6 categories. TDD discipline (RED → GREEN → REFACTOR). FROZEN zone matcher integrated.

**Deliverables**:
- `internal/design/dtcg/validator.go` — `Validate(tokens map[string]any) (*Report, error)` API.
- `internal/design/dtcg/categories/` package — per-category validators: `color.go`, `dimension.go`, `font.go`, `typography.go`, `shadow.go`, `border.go` (minimum 6); plus `font_family.go`, `font_weight.go`, `duration.go`, `cubic_bezier.go`, `number.go`, `stroke_style.go`, `transition.go`, `gradient.go` for full DTCG 2025.10 coverage.
- `internal/design/dtcg/errors.go` — structured error report types (`ValidationError`, `Report`).
- `internal/design/dtcg/frozen_guard.go` — path-prefix + heading-hash matcher for constitution §2, §3.1, §3.2, §3.3, §5, §11, §12.
- `internal/design/dtcg/dtcg_test.go` — unit tests covering all 6 minimum categories with positive (valid token) and negative (invalid token) cases each. Target coverage ≥85% per `quality.yaml`.
- `internal/design/dtcg/SPEC.md` — pinned DTCG 2025.10 spec snapshot reference.

**REQ-IDs covered**: REQ-DPL-004 (DTCG validator), REQ-DPL-011 (FROZEN zone protection).

**Exit gate**: `go test -race ./internal/design/dtcg/...` passes on macOS, Linux, Windows runners; coverage ≥85%; FROZEN guard rejects 5 known-violation paths in unit test.

### Phase 4 — expert-frontend Integration + Path A Validator Hookup (Priority: P1 High)

**Goal**: Wire DTCG validator into `expert-frontend` agent's token-consumption flow; preserve Path A behavior; integrate Path B1/B2 outputs.

**Deliverables**:
- `expert-frontend` agent prompt updated to invoke `internal/design/dtcg.Validate` on every consumed `tokens.json` before code generation.
- Validator failure → structured error returned to orchestrator with offending token path and rule violated.
- Path A integration test: existing Claude Design handoff bundle → tokens.json → validator → expert-frontend.
- Path B1 integration test (stub mode): meta-harness produces stub `my-harness-figma-extractor` skill with sample tokens.json → validator → expert-frontend.
- Path B2 integration test (stub mode): meta-harness produces stub `my-harness-pencil-mcp` skill → validator → expert-frontend.
- Brand-context warning surface: invalid token vs brand visual-identity mismatch → AskUserQuestion warning.

**REQ-IDs covered**: REQ-DPL-001 (Path A validator hookup), REQ-DPL-010 (validator auto-invocation), REQ-DPL-009 (brand conflict warning).

**Exit gate**: 3 integration tests pass (one per path); brand-context conflict generates expected AskUserQuestion warning; `expert-frontend` agent definition mirrored to `internal/template/templates/`.

### Phase 5 — Constitution §4 Amendment + Pencil-Integration Removal Coordination (Priority: P1 High)

**Goal**: Apply additive amendment to design constitution §4 Phase Contracts; coordinate pencil-integration removal with SPEC-V3R3-HARNESS-001.

**Deliverables**:
- `.claude/rules/moai/design/constitution.md` — §4 Phase Contracts table receives two new rows:
  - `figma-extractor (Path B1) | BRIEF + Figma file ID + page selectors | tokens.json + components.json | Path B1`
  - `pencil-mcp (Path B2) | BRIEF + .pen files | tokens.json + components.json | Path B2`
- HISTORY entry appended: `2026-04-26 (SPEC-V3R3-DESIGN-PIPELINE-001): §4 Phase Contracts table extended with Path B1 (figma-extractor) and Path B2 (pencil-mcp) rows. Version 3.3.0 → 3.4.0.`
- Version bump in constitution footer: `Version: 3.4.0`.
- Template-First mirror to `internal/template/templates/.claude/rules/moai/design/constitution.md`.
- `moai-workflow-pencil-integration` skill removal coordinated with SPEC-V3R3-HARNESS-001 BC-V3R3-007 — verify no orphaned references in:
  - `.claude/skills/moai/SKILL.md` routing table
  - `.claude/agents/moai/expert-frontend.md`
  - workflow files under `.claude/skills/moai/workflows/`
- zone-registry CONST-V3R2-068 verified unaffected (still references constitution §3.2; §4 is not in CONST-V3R2-051..072 range).

**REQ-IDs covered**: REQ-DPL-006 (pencil-integration removal), REQ-DPL-007 (constitution §4 amendment).

**Exit gate**: `make build && go test ./...` passes; plan-auditor verifies §4 amendment is additive-only (diff shows only appended rows + HISTORY + version line); zone-registry validation passes; no orphaned references to `moai-workflow-pencil-integration` remain.

## 4. Data Schemas

### 4.1 `.moai/design/path-selection.json`

```json
{
  "ts": "2026-04-26T20:00:00Z",
  "path": "A | B1 | B2",
  "brand_context_loaded": true,
  "spec_id": "SPEC-V3R3-DESIGN-PIPELINE-001",
  "session_id": "uuid-v7"
}
```

### 4.2 DTCG Validator `Report`

```go
type Report struct {
    Valid       bool
    TokenCount  int
    Errors      []ValidationError
    Warnings    []ValidationWarning  // brand-context conflicts surface here
    SpecVersion string               // "DTCG-2025.10"
}

type ValidationError struct {
    Path    string  // "color.brand.primary"
    Rule    string  // "color value must be a valid CSS color or hex"
    Got     any
    Want    string
}
```

### 4.3 Constitution §4 Amended Rows (additive)

```markdown
| figma-extractor (Path B1) | BRIEF + Figma file ID + page selectors | tokens.json + components.json | Path B1 |
| pencil-mcp (Path B2) | BRIEF + .pen files | tokens.json + components.json | Path B2 |
```

## 5. Technical Approach Decisions

### 5.1 Why a Go validator (vs JSON Schema)

- DTCG 2025.10 has cross-token constraints (e.g., typography references fontFamily and fontWeight tokens by alias) that pure JSON Schema cannot express ergonomically.
- Go validator integrates directly with `expert-frontend` agent invocation pipeline (existing Go runtime).
- Unit tests with positive/negative cases per category provide auditable correctness evidence.

### 5.2 Why FROZEN zone matcher in `internal/design/dtcg/`

- The validator is the only Go component that touches design constitution paths during validation.
- Centralizing the matcher next to the validator keeps the boundary check colocated with the consumer.
- Hardcoded prefix + heading-hash list — no config injection point — cannot be bypassed.

### 5.3 Why path-selection persistence

- Audit trail for plan-auditor and post-merge review.
- Idempotency: re-invocations of `/moai design` on the same project surface previous selection for confirmation.
- Enables future analytics (which paths are popular per project type) without telemetry leaving the machine.

### 5.4 Why additive constitution amendment (vs new section)

- §4 Phase Contracts already has the right structure (input/output/required-when columns).
- Additive rows are minimal-impact and preserve existing tooling that parses §4 (e.g., plan-auditor's phase-contract validator).
- New section would require zone-registry expansion and additional FROZEN/EVOLVABLE classification — out of scope.

### 5.5 Why Path A is the recommended default

- Claude Design handoff bundle is the most stable path with the longest production history.
- New users without Figma/Pencil credentials can immediately succeed with Path A.
- Path B1/B2 require credentials and per-project configuration that meta-harness must collect.

## 6. Dependencies on Other SPECs

This plan **cannot start Phase 4** until SPEC-V3R3-HARNESS-001 Phase 4 (meta-harness skill installation + 16-skill removal) is merged:

1. **SPEC-V3R3-HARNESS-001** — Provides `moai-meta-harness` skill that generates `my-harness-figma-extractor` and `my-harness-pencil-mcp`. Without HARNESS-001, Path B1/B2 cannot operate beyond stub mode.

Phase 1, 2, 3 can proceed independently (skill body updates, workflow routing, validator). Phase 4 integration tests for Path B1/B2 require HARNESS-001 to be at least at stub-mode completeness. Phase 5 pencil-integration removal MUST be coordinated with HARNESS-001 BC-V3R3-007 to avoid orphaned references.

## 7. Risks (Plan-Level)

| Risk | Severity | Mitigation |
|------|----------|------------|
| HARNESS-001 slips past v2.17 | High | This SPEC's Phase 4 stub mode for Path B1/B2 enables Phase 1/2/3/5 to land on schedule; Path B1/B2 production behavior defers to next minor. |
| DTCG 2025.10 spec changes during implementation | Medium | Pin spec snapshot in `internal/design/dtcg/SPEC.md`; version-tag categories. |
| Constitution §4 amendment introduces parser regression | Medium | Plan-auditor validates additive-only diff; zone-registry CONST-V3R2-068 untouched. |
| Cross-platform JSONL line-ending differences in path-selection.json | Low | Always emit `\n` (LF), never `\r\n`; verified by IT on Windows runner. |
| Brand-context conflict UX surprises user | Medium | Warning-not-block; explicit AskUserQuestion per conflict; documented in `.moai/harness/main.md` template comment. |
| `expert-frontend` agent already uses non-DTCG token format | Low | Phase 4 integration test exercises both old and new token paths; old format treated as unvalidated (warning-only) until Phase 5 hardening. |

## 8. Verification Strategy

- Unit coverage target: ≥85% per `quality.yaml` standard for `internal/design/dtcg/`.
- Integration test count: ≥3 (one per path A/B1/B2) plus FROZEN-guard regression suite.
- Cross-platform: ubuntu-latest, macos-latest, windows-latest in CI.
- Manual verification: design constitution diff reviewed by plan-auditor; FROZEN zone integrity confirmed; pencil-integration references grep-clean.
- DTCG conformance: validator output checked against the DTCG 2025.10 reference vectors (positive + negative samples) committed under `internal/design/dtcg/testdata/`.

## 9. Rollout

- Default `/moai design` behavior: Path A "(권장)" — no behavioral change for existing users.
- Path B1/B2 opt-in via `/moai design` AskUserQuestion selection — requires meta-harness skill from HARNESS-001.
- v2.17.0 release notes MUST highlight:
  - DTCG 2025.10 validator gate (auto-invoked by `expert-frontend`).
  - Path B1/B2 availability (Figma + Pencil via meta-harness).
  - `moai-workflow-pencil-integration` removal (BC-V3R3-007 coordinated migration).
- Post-release: `manager-docs` adds design pipeline section to docs-site (per CLAUDE.local.md §17 4-locale rule).
