# Acceptance Criteria — SPEC-V3R2-HRN-003 Hierarchical Acceptance Scoring

> Self-demonstrating: this file uses the **hierarchical** AC format introduced by SPC-001 (parent → `.a/.b/.c` children inheriting parent Given). 4 of 12 ACs are authored as parent/child trees (AC-HRN-003-03, -04, -05, -07), recursively proving SPC-001's schema works for HRN-003's own paperwork. AC-HRN-003-04 reaches depth-3 grandchildren `.a.i / .a.ii` exercising the MaxDepth=3 boundary.
>
> Flat ACs from spec.md §6 remain canonical and unchanged; this file augments them with hierarchical Given/When/Then breakdown for each REQ as plan-phase deliverable.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec (HRN-003 plan author) | Initial hierarchical AC authoring covering all 19 REQs across 5 EARS modalities. Self-demonstrates SPC-001 schema on AC-03/04/05/07. |

---

## 1. Format Conventions (this file)

- Top-level: `AC-HRN-003-NN` (e.g., `AC-HRN-003-01`).
- Depth-1 children: `.a / .b / .c` (lowercase letters).
- Depth-2 grandchildren: `.a.i / .a.ii` (lowercase Roman).
- Maximum depth = 3 (top + 2 child levels), enforced by `internal/spec/ears.go:21` `MaxDepth = 3`.
- Each leaf carries `(maps REQ-...)`. Intermediate nodes MAY omit when all leaves carry a tail.
- Children inherit parent's Given when the child's own Given is empty (`internal/spec/ears.go:77-81`).

## 1.1 Drift Reconciliation Notes (post-HRN-002 merge, 2026-05-13)

`spec.md` was authored 2026-04-23 against constitution v3.3.0. Between authorship and plan delivery (2026-05-13), SPEC-V3R2-HRN-002 (PR #879, main `0ac27ee4e`) advanced infrastructure that HRN-003 builds on. The spec.md HISTORY row 0.2.0 documents three drift reconciliations:

1. **Profile format**: spec.md REQ-005 assumed `.yaml` profile files; reality on main as of 2026-05-13 is `.md` profile files (`default.md`, `strict.md`, `lenient.md`, `frontend.md`) already shipping with rubric tables structured per design-constitution §12 Mechanism 1. AC-HRN-003-07 (this file) verifies against the `.md` reality.
2. **evaluator-active body**: spec.md REQ-006 assumed introducing a 4-dimension contract; reality is the body already declares the 4-dimension table (lines 47-54) and cites §11.4.1 (lines 91-92, landed by HRN-002 M3). AC-HRN-003-09 (this file) verifies the augment scope (hierarchical JSON schema + rubric-citation requirement), NOT the introduction.
3. **`internal/harness/gan_loop.go`**: file does NOT exist on main; HRN-003 inherits HRN-002 D1 (orchestrator-level runner via SKILL.md, not a Go-side runner module). AC-HRN-003-10 (this file) verifies Sprint Contract persistence via Go-side `WriteContract()` helper, NOT a runner module.

These reconciliations are bounded by the three items above; no other spec.md deviation expected.

## 2. REQ ↔ AC Traceability Matrix

| REQ ID | EARS modality | Mapped AC(s) | Notes |
|--------|---------------|--------------|-------|
| REQ-HRN-003-001 | Ubiquitous | AC-HRN-003-01, AC-HRN-003-08 | 4-dimension enum FROZEN |
| REQ-HRN-003-002 | Ubiquitous | AC-HRN-003-02 | Hierarchical ScoreCard tree shape |
| REQ-HRN-003-003 | Ubiquitous | AC-HRN-003-07 | 4 anchor levels in Rubric struct |
| REQ-HRN-003-004 | Ubiquitous | AC-HRN-003-02, AC-HRN-003-03 | EvaluatorRunner.Score() |
| REQ-HRN-003-005 | Ubiquitous | AC-HRN-003-07 | 4 evaluator profile `.md` files |
| REQ-HRN-003-006 | Ubiquitous | AC-HRN-003-09 | evaluator-active body augment |
| REQ-HRN-003-007 | Event-Driven | AC-HRN-003-03 | Aggregation rules (min default, mean opt-in) |
| REQ-HRN-003-008 | Event-Driven | AC-HRN-003-04, AC-HRN-003-12 | Must-pass firewall |
| REQ-HRN-003-009 | Event-Driven | AC-HRN-003-05 | Rubric citation rejection |
| REQ-HRN-003-010 | Event-Driven | AC-HRN-003-06 | Flat AC auto-wrap (transitive) |
| REQ-HRN-003-011 | Event-Driven | AC-HRN-003-10 | Sprint Contract sub-criterion persistence |
| REQ-HRN-003-012 | State-Driven | AC-HRN-003-08 | 4-dimension FROZEN at runtime |
| REQ-HRN-003-013 | State-Driven | AC-HRN-003-07 | 4-anchor FROZEN at runtime |
| REQ-HRN-003-014 | State-Driven | AC-HRN-003-12 | pass_threshold floor 0.60 |
| REQ-HRN-003-015 | Optional | AC-HRN-003-03 | Per-dimension mean aggregation |
| REQ-HRN-003-016 | Optional | AC-HRN-003-07 | Frontend profile UI dimensions (transitive — frontend.md exists) |
| REQ-HRN-003-017 | Unwanted | AC-HRN-003-02 | Flat ScoreCard prohibition |
| REQ-HRN-003-018 | Unwanted | AC-HRN-003-11 | Must-pass bypass prohibition |
| REQ-HRN-003-019 | Unwanted | AC-HRN-003-08 | Unknown dimension rejection |

Coverage check: 19 REQs declared in spec.md §5 (REQ-HRN-003-001 through 019), every REQ has ≥1 mapped AC entry above. Total ACs in this file: 12 (matches spec.md §6 count exactly).

---

## 3. Acceptance Criteria (hierarchical)

### 3.1 Test suite + coverage gate (flat)

- AC-HRN-003-01: Given the implementation post-M5, When a developer runs `go test ./internal/harness/...`, Then all tests pass and coverage on `internal/harness/scorer.go` + `internal/harness/rubric.go` is ≥85%. The Dimension enum exposes exactly 4 values (`Functionality`, `Security`, `Craft`, `Consistency`); calling `Dimension(99).IsValid()` returns false. (maps REQ-HRN-003-001)

### 3.2 Hierarchical ScoreCard tree shape (flat with negative cross-ref)

- AC-HRN-003-02: Given a fixture SPEC with 2 dimensions × 3 criteria × 2 sub-criteria, When `EvaluatorRunner.Score(contract, artifact)` is invoked, Then the returned `*ScoreCard` contains exactly 12 `SubCriterionScore` entries (2 × 3 × 2) accessible via `card.Dimensions[d].Criteria[c].SubCriteria[s]`. Conversely, When any code path attempts to construct a flat (non-hierarchical) ScoreCard at the type level (e.g., direct `[]float64` instead of nested `Dimensions map`), Then the type system rejects compilation; runtime cross-check returns `ErrFlatScoreCardProhibited` (`HRN_FLAT_SCORECARD_PROHIBITED`) per REQ-HRN-003-017. (maps REQ-HRN-003-002, REQ-HRN-003-004, REQ-HRN-003-017)

### 3.3 Aggregation rules (hierarchical — 3 children)

<!-- @MX:NOTE: Self-demonstrating hierarchical AC depth-2; aggregateMin/aggregateMean coverage -->
- AC-HRN-003-03: Given a single criterion with 2 sub-criterion scores `{0.8, 0.5}` and a `Rubric` with 4 valid anchor levels
  - AC-HRN-003-03.a: When the active profile declares `aggregation: min` (default per REQ-HRN-003-007), Then the criterion's `Aggregate` field equals `0.5` (the minimum). The `EvaluatorRunner.Score()` rationale records "aggregation: min". (maps REQ-HRN-003-007)
  - AC-HRN-003-03.b: When the active profile declares `aggregation: mean` per REQ-HRN-003-015, Then the criterion's `Aggregate` field equals `0.65` (the mean: (0.8 + 0.5) / 2). The `EvaluatorRunner.Score()` rationale records "aggregation: mean". (maps REQ-HRN-003-007, REQ-HRN-003-015)
  - AC-HRN-003-03.c: When the active profile declares per-dimension override (e.g., `Functionality.aggregation: mean` while other dimensions inherit profile-level `min`), Then `Functionality` criteria use mean aggregation while `Security`, `Craft`, `Consistency` criteria use min aggregation in the same ScoreCard. (maps REQ-HRN-003-015)

### 3.4 Must-pass firewall (hierarchical depth-3 — exercises MaxDepth=3 boundary)

<!-- @MX:ANCHOR fan_in=3 -->
<!-- @MX:REASON: "self-demonstrating hierarchical AC reference cited by HRN-003 self + HRN-002 dogfood precedent + future SPC-002 catalog" -->
- AC-HRN-003-04: Given an active profile declaring `MustPass: [Functionality, Security]` and a fixture ScoreCard with `Functionality.Aggregate = 1.00`, `Craft.Aggregate = 1.00`, `Consistency.Aggregate = 1.00`
  - AC-HRN-003-04.a: When `Security.Aggregate = 0.55` (below 0.60 floor)
    - AC-HRN-003-04.a.i: When `applyMustPassFirewall(card, mustPass)` is invoked, Then it returns `Verdict("fail")` regardless of the other 3 dimensions scoring 1.00. (maps REQ-HRN-003-008)
    - AC-HRN-003-04.a.ii: When the ScoreCard is serialized, Then `card.Verdict == "fail"` AND `card.Rationale` contains the substring `"must-pass dimension Security failed (0.55 < 0.60)"` to satisfy R6 user-friendliness mitigation. (maps REQ-HRN-003-008)
  - AC-HRN-003-04.b: When `Security.Aggregate = 0.95` (above floor) AND `Functionality.Aggregate = 0.55` (below floor on the second must-pass dimension), Then `applyMustPassFirewall()` returns `Verdict("fail")` and `Rationale` cites `Functionality` as the failing must-pass dimension. (maps REQ-HRN-003-008)
  - AC-HRN-003-04.c: When all 4 dimensions score `≥ 0.60` AND must-pass dimensions both score `≥ 0.60`, Then `applyMustPassFirewall()` returns `Verdict("pass")` and the overall ScoreCard verdict reflects each dimension's individual aggregate. (maps REQ-HRN-003-008)

### 3.5 Rubric citation enforcement (hierarchical — 3 children)

- AC-HRN-003-05: Given evaluator-active emits a sub-criterion score JSON object
  - AC-HRN-003-05.a: When the JSON object lacks the `rubric_anchor` field entirely, Then `ValidateCitation()` returns `ErrRubricCitationMissing` wrapping `HRN_RUBRIC_CITATION_MISSING` and the scorer requests re-evaluation (up to 2 retries per OQ5 default). (maps REQ-HRN-003-009)
  - AC-HRN-003-05.b: When the JSON object's `rubric_anchor` field is non-empty but contains a value outside `{"0.25", "0.50", "0.75", "1.00"}` (e.g., `"0.65"`), Then `ValidateCitation()` returns `ErrRubricCitationMissing` with message naming the invalid anchor value. (maps REQ-HRN-003-009, REQ-HRN-003-013)
  - AC-HRN-003-05.c: When the JSON object's `rubric_anchor` field equals one of the 4 canonical anchors AND the `score` field equals that anchor value (e.g., anchor `"0.75"` and score `0.75`), Then `ValidateCitation()` returns nil and the SubCriterionScore is accepted. (maps REQ-HRN-003-009)

### 3.6 Flat AC auto-wrap (transitive — flat AC verifies SPC-001 substrate)

- AC-HRN-003-06: Given a pre-v3r2 SPEC with flat acceptance criteria (no `.a/.b` children declared), When `EvaluatorRunner.Score()` is invoked on it, Then SPC-001's parser auto-wraps each top-level AC as a synthesized single-child `.a` per BC-V3R2-011 (`internal/spec/parser.go:200-227`); the scorer treats each as a `CriterionScore` with exactly 1 `SubCriterionScore` entry; aggregation collapses trivially (min/mean of a single value = the value); no information loss occurs. (maps REQ-HRN-003-010)

### 3.7 Profile loading + 4 anchor levels (hierarchical — 3 children)

- AC-HRN-003-07: Given the 4 evaluator profile `.md` files at `.moai/config/evaluator-profiles/` post-M4
  - AC-HRN-003-07.a: When `harness.ParseRubricMarkdown(".moai/config/evaluator-profiles/default.md")` is invoked, Then the returned `*Rubric` contains exactly 4 `Dimensions` entries (Functionality, Security, Craft, Consistency); each `DimensionRubric.Anchors` map has exactly 4 entries with keys `{0.25, 0.50, 0.75, 1.00}`. (maps REQ-HRN-003-003, REQ-HRN-003-005, REQ-HRN-003-013)
  - AC-HRN-003-07.b: When the same parser is invoked on `strict.md`, `lenient.md`, and `frontend.md` (all 4 profiles ship on main as of 2026-05-13), Then each returns a valid `*Rubric` with no errors; `frontend.md` produces a `Craft` dimension whose rubric anchor descriptions reference UI-specific sub-criteria (viewport, accessibility, animation) per REQ-HRN-003-016. (maps REQ-HRN-003-005, REQ-HRN-003-016)
  - AC-HRN-003-07.c: When `LoadHarnessConfig()` is invoked with the shipping `.moai/config/sections/harness.yaml`, Then it loads all 4 profiles into `cfg.Evaluator.Profiles` map (per OQ1 default field naming) AND validates each via `Rubric.Validate()` AND returns no error. (maps REQ-HRN-003-005)

### 3.8 Unknown dimension rejection (flat)

- AC-HRN-003-08: Given a malformed evaluator profile `.md` file at `internal/harness/testdata/profiles/malformed-5dim.md` declaring a 5th dimension `Performance` in the Evaluation Dimensions table, When `harness.ParseRubricMarkdown()` is invoked on it, Then the parser returns `ErrUnknownDimension` wrapping `HRN_UNKNOWN_DIMENSION`; the loader rejects the profile; the SchemaVersion field of any in-flight ScoreCard is unaffected (REQ-HRN-003-001 + REQ-HRN-003-012 enforcement). (maps REQ-HRN-003-001, REQ-HRN-003-012, REQ-HRN-003-019)

### 3.9 evaluator-active body augment (flat)

- AC-HRN-003-09: Given `.claude/agents/moai/evaluator-active.md` post-M4, When the reader scans the body for "## Hierarchical Score Output (Phase 5)", Then the new section exists between the current "## Output Format" closing line (line 77) and "## Evaluator Profile Loading" heading (line 79). The new section declares: (a) structured JSON schema with `schema_version: "v1"` (per OQ4 default), (b) mandatory `rubric_anchor` field per sub-criterion (REQ-HRN-003-009 enforcement contract), (c) cross-reference to `internal/harness/rubric.go`. The pre-existing 4-dimension table (lines 47-54) remains unchanged; the §11.4.1 cross-reference (lines 91-92, landed by HRN-002 M3) remains unchanged; no frontmatter modification. (maps REQ-HRN-003-006)

### 3.10 Sprint Contract sub-criterion persistence (flat)

- AC-HRN-003-10: Given an `*ScoreCard` with 12 sub-criterion entries from a completed iteration, When `WriteContract(contract, card, ".moai/sprints/SPEC-V3R2-HRN-003/contract.yaml")` is invoked, Then the file is written with the existing `acceptance_checklist[]` items extended with a new `status` field per item (enum: `passed | failed | refined | new`). When the GAN loop subsequently runs another iteration, Then HRN-002's leak detection validator (`internal/harness/evaluator_leak.go`) does NOT flag the contract YAML as a prior-judgment leak (the contract carries criterion state, not evaluator rationale — additive shape change is backward-compatible). (maps REQ-HRN-003-011)

### 3.11 Must-pass bypass prohibition (flat)

- AC-HRN-003-11: Given a malformed evaluator profile `.md` file at `internal/harness/testdata/profiles/malformed-bypass.md` attempting to set `must_pass: false` on the `Security` dimension via the "Must-Pass Criteria" section, When `harness.ParseRubricMarkdown()` is invoked followed by `Rubric.Validate()`, Then the validator returns `ErrMustPassBypassProhibited` wrapping `HRN_MUSTPASS_BYPASS_PROHIBITED` (per OQ3 default: `[Security]` is the floor; profiles MAY widen but MAY NOT narrow below). The error message cites design-constitution §12 Mechanism 3 as the FROZEN authority. (maps REQ-HRN-003-018)

### 3.12 Strict profile threshold (flat)

- AC-HRN-003-12: Given the shipping `strict.md` profile (loaded via `LoadHarnessConfig()`) declaring `Security.PassThreshold = 0.85` (above the 0.60 FROZEN floor per REQ-HRN-003-014), When a fixture ScoreCard is scored with `Security.Aggregate = 0.84`, Then the must-pass firewall (REQ-HRN-003-008) triggers and the overall verdict is `fail`. When the same fixture is scored with `Security.Aggregate = 0.86`, Then the must-pass firewall does NOT trigger and the overall verdict reflects the other dimensions' aggregates. The `Rubric.Validate()` method rejects any profile attempting to declare `Security.PassThreshold < 0.60` with `ErrInvalidConfig` citing the FROZEN floor. (maps REQ-HRN-003-008, REQ-HRN-003-014)

---

## 4. Edge Cases

| # | Scenario | Expected behaviour | Anchor |
|---|----------|--------------------|--------|
| E1 | Profile `.md` file uses tabs instead of spaces in rubric tables | Parser normalizes tabs→spaces before extraction; rubric tables parse successfully | `internal/harness/rubric.go` `ParseRubricMarkdown()` |
| E2 | Profile `.md` file declares anchor scores in non-canonical case (e.g., `"1.0"` vs `"1.00"`) | Parser normalizes to canonical 2-decimal form; `"1.0"` accepts as `1.00`; `"1.000"` rejects | `internal/harness/rubric.go` anchor normalizer |
| E3 | Sub-criterion score equals an anchor value but evaluator omits the citation | `ValidateCitation()` rejects; retry-on-reject pattern (max 2 retries per OQ5 default) | `internal/harness/rubric.go ValidateCitation()` |
| E4 | Acceptance.md has 4-level depth (`AC-X-01.a.i.x`) | SPC-001 parser rejects with `MaxDepthExceeded` per REQ-SPC-001-021; HRN-003 scorer never sees the malformed input | `internal/spec/ears.go:96-105` |
| E5 | Profile declares `aggregation: median` (unknown value) | Parser rejects with `ErrInvalidConfig` citing valid values `{min, mean}` | `internal/harness/rubric.go Validate()` |
| E6 | ScoreCard JSON output is valid against schema but evaluator-active emits zero sub-criterion scores | Empty SubCriteria map per CriterionScore is treated as `Aggregate = 0.0`; must-pass firewall triggers if dimension is must-pass | `internal/harness/scorer.go` aggregation default for empty input |
| E7 | Same SubCriterion ID appears twice in a single CriterionScore (parser bug or hand-authored fixture) | SPC-001 parser detects via `DuplicateAcceptanceID` per REQ-SPC-001-012; HRN-003 scorer never processes duplicates | `internal/spec/parser.go` |
| E8 | Active profile declares per-dimension `MustPass = [Craft]` (overriding default `[Functionality, Security]` with a NARROWER set excluding Security) | Validator rejects with `ErrMustPassBypassProhibited`; profiles may widen the must-pass set but may not narrow below `[Security]` floor (per OQ3 default) | `internal/harness/rubric.go Validate()` |

---

## 5. Quality Gate Criteria (Definition of Done)

### 5.1 Plan-phase DoD (this PR)

- [ ] All 12 ACs above carry at least one `(maps REQ-...)` reference on every leaf.
- [ ] At least 4 ACs (AC-HRN-003-03, -04, -05, -07) self-demonstrate the hierarchical schema with explicit `.a/.b/.c` children.
- [ ] AC-HRN-003-04 demonstrates depth-2 nesting (`.a.i`, `.a.ii`) to prove the MaxDepth=3 schema works.
- [ ] Every REQ in spec.md §5 (REQ-HRN-003-001 through REQ-HRN-003-019) appears at least once in §2 traceability matrix.
- [ ] Plan-auditor PASS at iteration ≤2.

### 5.2 Run-phase DoD (future PR per tasks.md)

- [ ] `internal/harness/scorer.go` + `internal/harness/rubric.go` exist with public types per spec.md §5.1.
- [ ] `go test ./internal/harness/...` green; coverage ≥85% on the two new files.
- [ ] `go test ./internal/config/...` green (covers AC-HRN-003-07.c via LoadHarnessConfig integration).
- [ ] All 4 shipping profile `.md` files load without error via `ParseRubricMarkdown()`.
- [ ] AC-HRN-003-01 through AC-HRN-003-12 all PASS via fixtures or integration tests.
- [ ] `.claude/agents/moai/evaluator-active.md` has new "## Hierarchical Score Output (Phase 5)" section (line 78-ish, between current line 77 and 79).
- [ ] `.claude/skills/moai-workflow-gan-loop/SKILL.md` has new "### Phase 3b: Hierarchical Scoring" subsection (between Phase 3 and Phase 4).
- [ ] Template mirrors byte-identical after `make build`.
- [ ] (per OQ1 default) `.claude/rules/moai/core/zone-registry.md` has new CONST-V3R2-154 (4-dim enum) + CONST-V3R2-155 (4 anchor levels) entries.
- [ ] `moai mx --validate` clean; all `@MX:WARN` tags carry `@MX:REASON` sub-line.

### 5.3 Sync-phase DoD

- [ ] CHANGELOG entry under `### Added` cites pattern-library E-1 (priority 9) + E-3 + R1 §9 Agent-as-a-Judge.
- [ ] CHANGELOG entry documents the additive `EvaluatorConfig` field extensions (`Profiles`, `Aggregation`, `MustPassDimensions`).
- [ ] zone-registry.md cross-link complete (per OQ1 default).
- [ ] docs-site 4-language sync if §12 Mechanism 1/3 is publicly exposed (CLAUDE.local.md §17).

---

## 6. Self-Demonstration Notice

This file uses the new hierarchical AC schema (depth 0 → 1 → 2) on AC-HRN-003-03, -04, -05, -07. Specifically:

- **AC-HRN-003-03** has 3 depth-1 children covering `min`, `mean`, and per-dimension override aggregation cases.
- **AC-HRN-003-04** has 3 depth-1 children (`.a/.b/.c`); AC-04.a has 2 depth-2 grandchildren (`.a.i / .a.ii`) covering firewall trigger + rationale message — this exercises the MaxDepth=3 boundary.
- **AC-HRN-003-05** has 3 depth-1 children covering missing field, invalid value, and valid citation.
- **AC-HRN-003-07** has 3 depth-1 children covering single-profile parse, all-4-profile parse, and full LoadHarnessConfig integration.

The remaining 8 ACs (AC-01, -02, -06, -08, -09, -10, -11, -12) are flat — demonstrating that flat and hierarchical co-exist within the same SPEC (REQ-SPC-001-040 → AC-SPC-001-09 from the SPC-001 plan-phase precedent).

Recursive proof: HRN-003's acceptance.md uses the hierarchical AC schema that SPC-001 introduced and HRN-002 dogfooded. If the SPC-001 parser (`internal/spec/parser.go`) successfully reads this file with `go test ./internal/spec/...` against the live fixture, then the SPC-001 schema works for HRN-003's own paperwork. This is the same recursive proof HRN-002's acceptance.md provides for itself.

End of acceptance.
