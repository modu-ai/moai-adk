# SPEC-V3R3-DESIGN-PIPELINE-001 — Acceptance Criteria

## 1. Acceptance-Requirement Traceability Matrix

| AC ID    | REQ-ID Coverage                                         | Test Type    | Phase |
|----------|---------------------------------------------------------|--------------|-------|
| AC-DPL-01 | REQ-DPL-001, REQ-DPL-002, REQ-DPL-003                   | Integration  | 1, 2, 4 |
| AC-DPL-02 | REQ-DPL-004                                             | Unit         | 3 |
| AC-DPL-03 | REQ-DPL-005, REQ-DPL-008                                | Integration  | 2 |
| AC-DPL-04 | REQ-DPL-006                                             | Integration  | 5 |
| AC-DPL-05 | REQ-DPL-009, REQ-DPL-011                                | Integration  | 3, 4 |
| AC-DPL-06 | REQ-DPL-010                                             | Integration  | 4 |
| AC-DPL-07 | REQ-DPL-007, REQ-DPL-012                                | Verification | 5 |

Matrix rows: 7. REQs covered: REQ-DPL-001 through REQ-DPL-012 (12 REQs total). Coverage ratio: 100% — every REQ appears in at least one AC.

---

## 2. Acceptance Criteria — Given/When/Then

### AC-DPL-01: Three-Path Workflow End-to-End

**Covers**: REQ-DPL-001 (Path A preservation), REQ-DPL-002 (Path B1 Figma routing), REQ-DPL-003 (Path B2 Pencil routing).

#### Scenario A1 — Path A (Claude Design) preserves existing behavior
- **Given** a project with `.moai/design/brief/BRIEF-001.md` and a Claude Design handoff bundle at `.moai/design/_handoff/`
- **When** the user invokes `/moai design` and selects Path A
- **Then** `.moai/design/tokens.json`, `.moai/design/components.json`, `.moai/design/assets/`, and `.moai/design/import-warnings.json` are produced exactly as in the pre-SPEC baseline
- **And** the produced `tokens.json` is automatically validated by the DTCG validator (REQ-DPL-010)
- **And** `.moai/design/path-selection.json` records `path: "A"`
- **And** no behavioral regression is observed against the pre-SPEC golden output (byte-equal except for the new validator-report sidecar)

#### Scenario B1 — Path B1 (Figma) generates dynamic figma-extractor
- **Given** the user has SPEC-V3R3-HARNESS-001 meta-harness skill installed
- **And** the user provides a Figma file ID and page selectors
- **When** the user invokes `/moai design` and selects Path B1
- **Then** `.claude/skills/my-harness-figma-extractor/SKILL.md` is generated with project-specific frontmatter (Figma file ID, page selectors, credential reference)
- **And** the figma-extractor produces `.moai/design/tokens.json` conforming to DTCG 2025.10
- **And** `.moai/design/path-selection.json` records `path: "B1"`

#### Scenario B2 — Path B2 (Pencil) generates dynamic pencil-mcp
- **Given** the user has SPEC-V3R3-HARNESS-001 meta-harness skill installed
- **And** the user has `.pen` files in the project
- **When** the user invokes `/moai design` and selects Path B2
- **Then** `.claude/skills/my-harness-pencil-mcp/SKILL.md` is generated with project-specific frontmatter (pen file paths, MCP server endpoint)
- **And** the pencil-mcp extractor produces `.moai/design/tokens.json` conforming to DTCG 2025.10
- **And** `.moai/design/path-selection.json` records `path: "B2"`

#### Edge Cases
- Selection persistence survives a `/moai design` re-invocation (idempotency).
- Missing meta-harness skill on Path B1/B2 selection produces a clear error pointing to SPEC-V3R3-HARNESS-001 prerequisite.
- Concurrent `/moai design` invocations on the same project produce a single `path-selection.json` (last-write-wins with audit log).

---

### AC-DPL-02: DTCG 2025.10 Validator Unit Tests (≥6 Categories, 100% Pass)

**Covers**: REQ-DPL-004.

#### Scenario — All six minimum categories validate correctly
- **Given** the Go package `internal/design/dtcg/`
- **When** `go test -race ./internal/design/dtcg/...` runs
- **Then** all unit tests pass on ubuntu-latest, macos-latest, windows-latest
- **And** the following 6 categories are covered with both positive and negative test cases:
  - `color` (hex, rgb, hsl, named, alpha)
  - `dimension` (px, rem, em, %, unitless)
  - `font` (composite font reference)
  - `typography` (composite typography token)
  - `shadow` (single + multi-layer shadow)
  - `border` (width + style + color composite)
- **And** test coverage ≥85% per `quality.yaml`
- **And** structured error reports (`ValidationError` with `Path`, `Rule`, `Got`, `Want` fields) are produced for every negative case
- **And** the validator pins to `DTCG-2025.10` spec snapshot documented in `internal/design/dtcg/SPEC.md`

#### Edge Cases
- Empty token map → valid (zero tokens to validate).
- Unknown token category → structured warning (not error).
- Cross-token alias references (typography → fontFamily + fontWeight) resolve correctly.
- Cyclic alias reference detected and reported as error.
- Non-UTF-8 token file rejected with clear error.

---

### AC-DPL-03: `/moai design` Workflow Path Branching with AskUserQuestion

**Covers**: REQ-DPL-005 (workflow branching), REQ-DPL-008 (AskUserQuestion path selection).

#### Scenario — AskUserQuestion presents three paths in correct order
- **Given** a project where `/moai design` has not yet been invoked
- **When** the user runs `/moai design`
- **Then** an AskUserQuestion is presented with exactly three options
- **And** the first option is `Path A (Claude Design)` marked `(권장)` with description explaining Claude Design handoff bundle prerequisite
- **And** the second option is `Path B1 (Figma)` with description explaining Figma file access prerequisite
- **And** the third option is `Path B2 (Pencil)` with description explaining `.pen` file prerequisite
- **And** option count does not exceed Claude Code's max-4-options limit
- **And** upon user selection, the workflow branches to the correct path before delegating to any specialist agent
- **And** `.moai/design/path-selection.json` is written before the branch executes (audit trail)

#### Edge Cases
- User cancels the AskUserQuestion → workflow exits cleanly without partial side effects.
- User selects "Other" (auto-fallback) → workflow displays explicit guidance to choose A/B1/B2.
- Brand context fails to load → workflow halts with diagnostic pointing to `.moai/project/brand/`.

---

### AC-DPL-04: `moai-workflow-pencil-integration` Removal — Design Workflow Operates Normally

**Covers**: REQ-DPL-006.

#### Scenario — Static skill removal preserves Pencil workflow capability via Path B2
- **Given** SPEC-V3R3-HARNESS-001 BC-V3R3-007 has removed `moai-workflow-pencil-integration` from the static skill set
- **When** the user invokes `/moai design` and selects Path B2 (Pencil)
- **Then** the workflow successfully generates `my-harness-pencil-mcp` via meta-harness
- **And** Pencil-based design import produces `.moai/design/tokens.json` correctly
- **And** no error or warning references `moai-workflow-pencil-integration`
- **And** `grep -r "moai-workflow-pencil-integration" .claude/` returns zero matches outside archival directories

#### Edge Cases
- Project upgraded from pre-BC-V3R3-007 with cached `moai-workflow-pencil-integration` references → migration runs cleanly via meta-harness regeneration.
- User attempts to manually re-create `moai-workflow-pencil-integration` skill → no special handling required; meta-harness path supersedes it; SKILL.md routing table no longer references the legacy name.

---

### AC-DPL-05: Brand Context FROZEN Protection — Build Fails on Violation

**Covers**: REQ-DPL-009 (brand context priority), REQ-DPL-011 (FROZEN zone protection).

#### Scenario A — Brand context wins on conflict
- **Given** `.moai/project/brand/visual-identity.md` defines `primary color: #1a73e8`
- **And** Path B1/B2 produces a `tokens.json` with `color.brand.primary: #ff0000`
- **When** the validator runs (Phase 4 hookup)
- **Then** a `ValidationWarning` is surfaced with category `brand-conflict`
- **And** the warning is presented to the user via AskUserQuestion (brand-vs-token resolution)
- **And** brand context is NOT silently overridden

#### Scenario B — FROZEN zone modification attempt fails build
- **Given** any pipeline component that attempts to write to a path matching:
  - `.claude/rules/moai/design/constitution.md` outside §4 Phase Contracts row range
  - `.moai/project/brand/brand-voice.md`
  - `.moai/project/brand/visual-identity.md`
  - `.moai/project/brand/target-audience.md`
- **When** the build executes (`go test ./internal/design/dtcg/... -run FrozenGuard`)
- **Then** the build fails with a structured error identifying the violated FROZEN zone
- **And** the error message names the constitution section (§2, §3.1, §3.2, §3.3, §5, §11, §12)
- **And** the change is NOT persisted to disk
- **And** the FROZEN guard cannot be bypassed by configuration (no config flag exists for it)

#### Edge Cases
- Modification of constitution §4 Phase Contracts (additive amendment) → allowed (verified by AC-DPL-07).
- Modification of FROZEN section §3.1 Brand Context → rejected with `FROZEN_VIOLATION_3_1`.
- Modification of constitution §11 GAN Loop contract → rejected with `FROZEN_VIOLATION_11`.
- 5 known-violation paths in unit test (`TestFrozenGuard_KnownViolations`) all rejected.

---

### AC-DPL-06: GAN Loop Integration — Score Variance ≤ ±0.05

**Covers**: REQ-DPL-010 (validator auto-invocation in `expert-frontend`).

#### Scenario — DTCG validator auto-invocation does not destabilize evaluator-active scoring
- **Given** a baseline GAN loop run (pre-SPEC) producing an evaluator-active score for a sample project
- **And** the same project running through Path A with the new DTCG validator gate
- **When** evaluator-active scores the post-validator output
- **Then** the score delta versus baseline is within ±0.05 (per design constitution §11 improvement_threshold)
- **And** valid tokens pass the validator gate without delaying expert-frontend code generation
- **And** invalid tokens block code generation with a structured error report identifying offending token paths
- **And** the validator runs in <100ms for typical token sets (≤500 tokens)

#### Edge Cases
- Token count > 1000 → validator completes in <500ms (informational benchmark, not gate).
- Validator panic → caught by recover; expert-frontend receives a clear error, not a process crash.
- Validator on non-Path-A token format (legacy) → warning-only mode until Phase 5 hardening.

---

### AC-DPL-07: Constitution §4 Additive Amendment + zone-registry Untouched

**Covers**: REQ-DPL-007 (constitution §4 amendment), REQ-DPL-012 (Template-First).

#### Scenario A — §4 Phase Contracts table extended additively
- **Given** the pre-SPEC constitution at version 3.3.0
- **When** the SPEC-V3R3-DESIGN-PIPELINE-001 implementation lands
- **Then** the diff against the pre-SPEC constitution shows ONLY the following changes:
  - HISTORY entry appended at the top (1 row)
  - §4 Phase Contracts table receives 2 new rows (figma-extractor, pencil-mcp)
  - Footer version bumped from `Version: 3.3.0` to `Version: 3.4.0`
  - Footer "Last Updated" line updated
- **And** no other section content is modified
- **And** all FROZEN zones (§2, §3.1, §3.2, §3.3, §5, §11, §12) are byte-identical to the pre-SPEC version

#### Scenario B — zone-registry CONST-V3R2-068 unaffected
- **Given** `.claude/rules/moai/core/zone-registry.md` containing CONST-V3R2-051..072 entries
- **When** the constitution amendment lands
- **Then** CONST-V3R2-068 (the design constitution §3.2 reserved file paths entry) remains unchanged
- **And** no new CONST-V3R2-NNN entries need to be added (since §4 is not part of the registry's Frozen-zone enumeration; §4 is EVOLVABLE per design constitution §2)
- **And** `moai constitution list --zone frozen` output is identical pre/post amendment

#### Scenario C — Template-First mirror is consistent
- **Given** the modified `.claude/rules/moai/design/constitution.md`
- **When** `make build` runs
- **Then** `internal/template/templates/.claude/rules/moai/design/constitution.md` is byte-identical to the working copy
- **And** `moai init` on a fresh project deploys the v3.4.0 constitution
- **And** the same Template-First check passes for:
  - `.claude/skills/moai-workflow-design-import/SKILL.md`
  - `.claude/skills/moai-design-system/SKILL.md`
  - `.claude/skills/moai/workflows/design.md`

#### Edge Cases
- Plan-auditor diff check rejects any non-additive change to §4 (e.g., row deletion, column rename) → expected fail.
- Version bump skipped in error → plan-auditor flags missing version increment.
- Template-First mirror missing → `make build` fails with clear error pointing to the missing mirror.

---

## 3. Definition of Done

All seven acceptance criteria above MUST pass before the SPEC is marked `completed`:

- [ ] AC-DPL-01: Three-path workflow integration tests pass (3 scenarios + 3 edge cases)
- [ ] AC-DPL-02: DTCG validator unit tests cover ≥6 categories with positive + negative cases, coverage ≥85%, all 3 OS targets green
- [ ] AC-DPL-03: AskUserQuestion path selection ordering verified (Path A first with `(권장)`)
- [ ] AC-DPL-04: `moai-workflow-pencil-integration` references absent from non-archival directories
- [ ] AC-DPL-05: FROZEN guard rejects 5+ known-violation paths in unit test; brand-conflict warnings surface correctly
- [ ] AC-DPL-06: evaluator-active GAN-loop score variance ≤±0.05 vs baseline
- [ ] AC-DPL-07: Constitution diff is additive-only; zone-registry untouched; Template-First mirrors consistent
- [ ] All HARD constraints from spec.md §7 verified (Template-First, FROZEN immutability, no network egress, ≥6 DTCG categories, AskUserQuestion ordering, pencil-integration coordination, English-only instructions)
- [ ] `make build && go test -race ./...` passes on macOS, Linux, Windows
- [ ] Plan-auditor sign-off on §4 amendment additive-only diff

## 4. Quality Gate Criteria

Per TRUST 5 framework (constitution §6):

- **Tested**: ≥85% unit coverage in `internal/design/dtcg/`; ≥3 integration tests across paths; FROZEN-guard regression suite.
- **Readable**: Validator API surface ≤10 exported symbols; per-category files ≤200 lines.
- **Unified**: Go formatting via `gofmt`; consistent error wrapping (`fmt.Errorf("%w", err)`).
- **Secured**: No network egress from `internal/design/dtcg/`; FROZEN guard cannot be bypassed via config; no credential storage.
- **Trackable**: Conventional commits referencing SPEC-V3R3-DESIGN-PIPELINE-001; CHANGELOG entry under v2.17.0; `path-selection.json` audit trail.

## 5. Verification Commands

```bash
# Phase 3 validator
go test -race -cover ./internal/design/dtcg/...

# Phase 4 cross-platform integration
go test -tags=integration ./internal/design/dtcg/... ./internal/design/pipeline/...

# Phase 5 constitution diff (additive-only)
diff -u .claude/rules/moai/design/constitution.md.pre-spec .claude/rules/moai/design/constitution.md | grep -E '^[+-]' | grep -v '^[+-]{3}'
# Expected output: only HISTORY row, 2 §4 rows, version line, last-updated line

# Pencil-integration removal verification
grep -r "moai-workflow-pencil-integration" .claude/ --exclude-dir=archive
# Expected: 0 matches

# Template-First mirror check
make build && diff -rq .claude/skills/moai-workflow-design-import/ internal/template/templates/.claude/skills/moai-workflow-design-import/

# zone-registry untouched
moai constitution list --zone frozen --format json > /tmp/post.json
diff /tmp/pre.json /tmp/post.json
# Expected: no diff
```
