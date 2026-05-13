# Plan — SPEC-V3R2-HRN-003 Hierarchical Acceptance Scoring

> Phase 1B implementation plan. Based on research.md §11 gap analysis: ~80% greenfield work — pre-existing substrate covers REQ-005 read-side (`.md` profiles already exist), REQ-006 partial (4-dim table + §11.4.1 cross-ref already in agent body), REQ-010 transitively (SPC-001 auto-wraps flat ACs), REQ-013 transitively (HRN-002 leak detection enforces fresh judgment), REQ-016 transitively (frontend.md already has UI sub-criteria). Plan organizes the remaining work into 5 milestones with M5 covering tests + zone-registry mirror entries (NOT a CON-002 amendment cycle — see Decision D8).

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec (HRN-003 plan author) | Initial plan covering M1–M5 milestones for HRN-003. Mirrors HRN-002 plan structure (PR #879). Drift reconciliations from spec.md v0.2.0 baked into §3 File Impact Analysis. |

---

## 1. Overview

Implement hierarchical 4-dimension × sub-criteria scoring in `internal/harness/scorer.go` + `internal/harness/rubric.go`. Augment `.claude/agents/moai/evaluator-active.md` body with a hierarchical JSON output schema and explicit rubric-citation requirement. Wire `EvaluatorConfig` (HRN-002 minimal substrate) with profile/aggregation/must-pass field additions. Consume the existing `.md` evaluator profile rubric tables via a tolerant Markdown parser. Persist per-sub-criterion state to `.moai/sprints/{spec-id}/contract.yaml` extending HRN-002 §11.4.1 substrate.

Out of scope: changing the 4 canonical dimensions, changing the 4 anchor levels, changing the pass_threshold floor 0.60, hierarchical AC parsing (SPC-001 owns), GAN loop runner module (HRN-002 D1 inherited), per-SPEC dimension weighting (all dimensions equal weight), `.yaml` parallel profile schema (Decision D1).

Source pattern: pattern-library E-1 (priority 9, ADOPT) + E-3 (ADOPT). Cite design-constitution §11.4.1 (HRN-002), §12 Mechanism 1 (Rubric Anchoring, FROZEN), §12 Mechanism 3 (Must-Pass Firewall, FROZEN), §5 (pass_threshold floor 0.60, FROZEN). Full SPEC: `.moai/specs/SPEC-V3R2-HRN-003/spec.md`.

## 2. Implementation Strategy

TDD per quality.yaml `development_mode: tdd`. Each REQ has at least one failing test fixture before the GREEN implementation. Standard Go cycle: RED test → GREEN minimal impl → REFACTOR for clarity. Coverage target ≥85% on `internal/harness/scorer.go` + `internal/harness/rubric.go` (matches HRN-001 §3 environment).

Sequencing rationale (mirrors HRN-002 plan §3):
- M1 lays out plan-phase artifacts (this PR).
- M2 lands type definitions (Dimension enum, ScoreCard, Rubric structs) so all downstream tasks can compile.
- M3 lands the scoring engine (aggregation, must-pass firewall, rubric citation, Sprint Contract persistence) — the meat of HRN-003.
- M4 lands integration points (profile loader extension, EvaluatorConfig fields, evaluator-active body augment, SKILL.md cross-reference).
- M5 lands quality gate (full test fixtures including 12 ACs, MX tags, optional zone-registry mirror entries per OQ1).

## 3. File Impact Analysis

### 3.1 NEW files (greenfield)

| File | Approx LOC | Purpose | Milestone |
|------|------------|---------|-----------|
| `internal/harness/scorer.go` | ~250 | Dimension enum, ScoreCard/DimensionScore/CriterionScore/SubCriterionScore types, EvaluatorRunner.Score(), aggregation logic, must-pass firewall, WriteContract() helper | M2 + M3 |
| `internal/harness/rubric.go` | ~120 | Rubric struct, Markdown table parser for `.md` profile files, anchor validator, citation enforcement | M2 + M4 |
| `internal/harness/scorer_test.go` | ~400 | Unit tests for all 12 ACs, aggregation min/mean, must-pass firewall, REQ-017 flat-prohibition integration | M5 |
| `internal/harness/rubric_test.go` | ~150 | Markdown parser tests, 4-dimension validator, 4-anchor validator, profile-file fixture loads | M5 |
| `internal/harness/testdata/profiles/malformed-5dim.md` | ~30 | Negative fixture for REQ-019 HRN_UNKNOWN_DIMENSION | M5 |
| `internal/harness/testdata/profiles/malformed-bypass.md` | ~30 | Negative fixture for REQ-018 HRN_MUSTPASS_BYPASS_PROHIBITED | M5 |
| `internal/harness/testdata/scorecards/2dim-3crit-2sub.json` | ~60 | Positive fixture for AC-HRN-003-02 (12 sub-criterion entries) | M5 |
| `.claude/rules/moai/core/zone-registry.md` (CONST-V3R2-154 entry) | ~10 | New zone-registry entry for 4-dimension enum FROZEN constraint (per OQ1 default) | M5 |
| `.claude/rules/moai/core/zone-registry.md` (CONST-V3R2-155 entry) | ~10 | New zone-registry entry for 4 anchor-level FROZEN constraint (per OQ1 default) | M5 |

### 3.2 MODIFIED files (additive)

| File | Approx LOC delta | Change scope | Milestone |
|------|------------------|--------------|-----------|
| `internal/config/types.go` | +30 | Extend `EvaluatorConfig` with `Profiles map[string]string`, `Aggregation string`, `MustPassDimensions []string` (additive vs HRN-002 minimal substrate) | M4 |
| `internal/config/loader.go` | +40 | Extend `LoadHarnessConfig()` with profile-loading step that calls `harness.ParseRubricMarkdown()` per profile name | M4 |
| `internal/config/errors.go` | +20 | 4 new sentinels: `ErrUnknownDimension`, `ErrRubricCitationMissing`, `ErrFlatScoreCardProhibited`, `ErrMustPassBypassProhibited` | M2 |
| `internal/config/loader_test.go` | +60 | Test fixtures loading the 4 shipping `.md` profiles + 2 malformed variants | M5 |
| `.claude/agents/moai/evaluator-active.md` | +35 | Insert "## Hierarchical Score Output (Phase 5)" section between line 77 and 79 (between flat output format and profile loading); declare structured JSON schema, rubric-citation requirement, cross-reference to `internal/harness/rubric.go` | M4 |
| `.claude/skills/moai-workflow-gan-loop/SKILL.md` | +25 | Insert "## Phase 3b: Hierarchical Scoring (REQ-HRN-003-004)" subsection in GAN Loop Execution Flow; cite `EvaluatorRunner.Score()` and Sprint Contract sub-criterion shape | M4 |
| `internal/template/templates/.claude/agents/moai/evaluator-active.md` | +35 | Template mirror per Template-First | M4 |
| `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md` | +25 | Template mirror per Template-First | M4 |

### 3.3 NOT MODIFIED (read-only references)

| File | Reason |
|------|--------|
| `.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.md` | Already on main with correct rubric structure; HRN-003 parser consumes existing format |
| `internal/template/templates/.moai/config/evaluator-profiles/*.md` | Template mirror byte-identical, no change needed |
| `.claude/rules/moai/design/constitution.md` | HRN-003 reads §5, §11.4.1, §12 Mechanism 1/3 — all FROZEN, no amendment |
| `internal/spec/ears.go`, `internal/spec/parser.go`, `internal/spec/lint.go` | SPC-001 surface; HRN-003 consumes `[]Acceptance` tree only |
| `internal/harness/evaluator_leak.go` | HRN-002 substrate; transitively satisfies REQ-013 |
| `internal/harness/gan_loop.go` | DOES NOT EXIST; HRN-003 inherits HRN-002 D1 (no Go-side runner) |

### 3.4 Template-First reminder

[HARD per CLAUDE.local.md §2] Any change to `.claude/agents/moai/evaluator-active.md` or `.claude/skills/moai-workflow-gan-loop/SKILL.md` MUST land in `internal/template/templates/.claude/agents/moai/evaluator-active.md` + `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md` first. `make build` regenerates `internal/template/embedded.go`. Local copies are mirrors, not sources of truth. The 4 evaluator profile `.md` files are already byte-identical between template and local — no Template-First action needed for read-side consumption (M4).

## 4. MX Tag Targets

Per CLAUDE.md §5 MX Tag Integration and `.claude/rules/moai/workflow/mx-tag-protocol.md`:

| Tag type | Count | Target locations |
|----------|-------|------------------|
| `@MX:NOTE` | 5 | (1) `internal/harness/scorer.go` Dimension enum block — "FROZEN at 4 values per spec.md REQ-001"; (2) `internal/harness/rubric.go` anchor levels — "FROZEN at {0.25, 0.50, 0.75, 1.00} per design-constitution §12 Mechanism 1"; (3) `internal/harness/scorer.go` aggregation function — "default min per REQ-007; mean opt-in per REQ-015"; (4) `.claude/agents/moai/evaluator-active.md` new hierarchical output section — "Cross-references internal/harness/rubric.go (SPEC-V3R2-HRN-003)"; (5) `.claude/skills/moai-workflow-gan-loop/SKILL.md` Phase 3b subsection — "Per design-constitution §12 Mechanism 1; enforces REQ-HRN-003-009 rubric citation" |
| `@MX:WARN` | 3 | (1) `internal/harness/scorer.go` Dimension enum — "FROZEN-zone constraint; addition of 5th dimension requires CON-002 amendment per zone-registry CONST-V3R2-154"; (2) `internal/harness/rubric.go` anchor validator — "FROZEN-zone constraint per CONST-V3R2-155"; (3) `internal/harness/scorer.go` must-pass firewall — "FROZEN per design-constitution §12 Mechanism 3; bypass requires CON-002 amendment" |
| `@MX:ANCHOR` | 2 | (1) `internal/harness/scorer.go EvaluatorRunner.Score()` (`fan_in=3` expected: scorer_test.go + integration with SKILL.md + future HRN-001 router); (2) `acceptance.md` AC-HRN-003-04 self-demonstrating hierarchical example block (`fan_in=3` expected: HRN-003 self + HRN-002 dogfood precedent + future SPC-002 catalog) |
| `@MX:TODO` | 0 | None at plan phase; M5 introduces transient TODOs only if rubric parser hits malformed real-world profiles during implementation |

All `@MX:WARN` tags require `@MX:REASON` sub-line per protocol.

## 5. Milestone Decomposition

### M1 — Plan-phase artifacts (this PR) — Priority Critical

Owner: `manager-spec`

Deliverables (this PR):
- spec.md v0.2.0 (refined from 2026-04-23 draft) — DONE.
- research.md (50 file:line anchors, §14 OQ1-OQ5, §15 D1-D10) — DONE.
- plan.md (this file) — IN PROGRESS.
- acceptance.md (12 ACs, ≥4 hierarchical depth-2 dogfooding) — pending.
- tasks.md (T-HRN003-NN naming, owner roles per task) — pending.
- spec-compact.md (~4KB summary) — pending.
- progress.md (frontmatter `plan_status: audit-ready`) — pending.

Exit criteria:
- plan-auditor independent verification PASS at iteration ≤2.
- ≥25 file:line anchors in research.md (50 confirmed).
- All 19 REQs in spec.md §5 mapped to ≥1 AC in acceptance.md.
- ≥4 ACs use hierarchical depth-2 schema.
- progress.md frontmatter `plan_status: audit-ready`.

### M2 — Type definitions + error sentinels — Priority Critical

Owner: `expert-backend`

File:line anchors:
- `internal/harness/scorer.go` (NEW) — `type Dimension int; iota constants Functionality | Security | Craft | Consistency`. `String()` + `IsValid()` methods. `MustPassDimensions []Dimension = []Dimension{Functionality, Security}` exported default.
- `internal/harness/scorer.go` (NEW) — `type ScoreCard struct { SchemaVersion string; SpecID string; Dimensions map[Dimension]DimensionScore; Verdict Verdict; Rationale string }`. `type DimensionScore struct { Aggregate float64; Criteria map[string]CriterionScore }`. `type CriterionScore struct { Aggregate float64; SubCriteria map[string]SubCriterionScore }`. `type SubCriterionScore struct { Score float64; RubricAnchor string; Evidence string; Dimension Dimension }`. `type Verdict string` enum: `pass | fail`.
- `internal/harness/rubric.go` (NEW) — `type Rubric struct { ProfileName string; Dimensions map[Dimension]DimensionRubric; PassThreshold float64; MustPass []Dimension; Aggregation string }`. `type DimensionRubric struct { Weight float64; PassThreshold float64; Anchors map[float64]string }`. `Validate()` method enforces 4-dim + 4-anchor + 0.60 floor + must-pass non-narrowing.
- `internal/config/errors.go` — Append 4 new sentinels: `ErrUnknownDimension`, `ErrRubricCitationMissing`, `ErrFlatScoreCardProhibited`, `ErrMustPassBypassProhibited`. Follow existing `ErrEvalMemoryFrozen` naming/message format.

Tasks:
- T2.1 RED: Add failing test for Dimension enum (REQ-001).
- T2.2 GREEN: Define Dimension enum with 4 iota constants.
- T2.3 REFACTOR: Add String() + IsValid() methods.
- T2.4 RED: Add failing test for ScoreCard tree shape (REQ-002).
- T2.5 GREEN: Define ScoreCard, DimensionScore, CriterionScore, SubCriterionScore types.
- T2.6 RED: Add failing test for Rubric struct + 4-anchor validation (REQ-003, REQ-013).
- T2.7 GREEN: Define Rubric struct + Validate() method enforcing anchors.
- T2.8 GREEN: Add 4 error sentinels.

mx_plan tags:
- `@MX:NOTE` on Dimension enum block.
- `@MX:WARN reason="FROZEN-zone constraint per CONST-V3R2-154"` on Dimension enum.
- `@MX:WARN reason="FROZEN-zone constraint per CONST-V3R2-155"` on Rubric anchor validator.

Exit criteria:
- `go test ./internal/harness/...` green for new test files (M2 RED tests now GREEN).
- `go vet ./internal/harness/...` clean.
- 4 new error sentinels exported and consumable.
- `Dimension(99).IsValid() == false`; only 4 canonical values pass.
- `Rubric.Validate()` rejects 5th anchor, rejects 5th dimension, rejects pass_threshold < 0.60.

### M3 — Scoring engine + Sprint Contract persistence — Priority Critical

Owner: `expert-backend`

File:line anchors:
- `internal/harness/scorer.go` — `EvaluatorRunner.Score(contract *SprintContract, artifact Artifact) (*ScoreCard, error)` (REQ-004). Iterates SPC-001 `[]Acceptance` tree. Per leaf invokes evaluator-active (mocked at unit-test layer; integration covered by harness leak test at M5). Aggregates per REQ-007.
- `internal/harness/scorer.go` — `aggregateMin([]float64) float64` + `aggregateMean([]float64) float64` (REQ-007, REQ-015).
- `internal/harness/scorer.go` — `applyMustPassFirewall(card *ScoreCard, mustPass []Dimension) Verdict` (REQ-008). If any must-pass dimension's `Aggregate < dimensionThreshold`, returns `fail` regardless of other dimensions.
- `internal/harness/scorer.go` — `WriteContract(contract *SprintContract, card *ScoreCard, path string) error` (REQ-011). Persists per-sub-criterion `status` field to `.moai/sprints/{spec-id}/contract.yaml`.
- `internal/harness/rubric.go` — `ValidateCitation(score SubCriterionScore, rubric *Rubric) error` (REQ-009). Returns `ErrRubricCitationMissing` if `RubricAnchor` is empty or not in `{0.25, 0.50, 0.75, 1.00}`.
- `internal/harness/scorer.go` — `EvaluatorRunner.Score()` rejects flat (non-hierarchical) call paths via type system: `Score()` only accepts `[]Acceptance` from SPC-001 parser, not raw `[]string`. Integration test at M5 verifies REQ-017 enforcement.

Tasks:
- T3.1 RED: Add failing test for EvaluatorRunner.Score() iterating 2-dim × 3-criterion × 2-subcrit fixture (AC-HRN-003-02).
- T3.2 GREEN: Implement Score() iteration + per-leaf invoke (mocked evaluator-active returns canned scores).
- T3.3 RED: Add failing test for aggregation min vs mean (AC-HRN-003-03).
- T3.4 GREEN: Implement aggregateMin + aggregateMean.
- T3.5 RED: Add failing test for must-pass firewall (AC-HRN-003-04).
- T3.6 GREEN: Implement applyMustPassFirewall.
- T3.7 RED: Add failing test for rubric citation rejection (AC-HRN-003-05).
- T3.8 GREEN: Implement ValidateCitation.
- T3.9 RED: Add failing test for must-pass bypass rejection (AC-HRN-003-11).
- T3.10 GREEN: Implement bypass rejection inside Rubric.Validate() (extends M2 validator).
- T3.11 RED: Add failing test for Sprint Contract persistence with sub-criterion status (AC-HRN-003-10).
- T3.12 GREEN: Implement WriteContract().
- T3.13 REFACTOR: Extract aggregation strategy into `Aggregator` interface for clarity.

mx_plan tags:
- `@MX:NOTE` on aggregation function.
- `@MX:WARN reason="FROZEN per design-constitution §12 Mechanism 3"` on must-pass firewall.
- `@MX:ANCHOR fan_in=3` on EvaluatorRunner.Score().

Exit criteria:
- All M3 RED tests now GREEN.
- AC-HRN-003-02, -03, -04, -05, -10, -11 PASS at unit-test level.
- Sprint Contract YAML produced by WriteContract() validates against existing `.claude/skills/moai-workflow-gan-loop/SKILL.md:210-240` shape (additive `status` field).
- Coverage on `internal/harness/scorer.go` ≥85%.

### M4 — Profile loader + EvaluatorConfig + agent body augment + SKILL.md — Priority High

Owner: `expert-backend` (Go side) + `manager-docs` (agent body + SKILL.md)

File:line anchors:
- `internal/harness/rubric.go` — `ParseRubricMarkdown(path string) (*Rubric, error)` (REQ-005). Extracts H2 "Evaluation Dimensions" table → `Weight` + `PassThreshold` per dim; H2 "Must-Pass Criteria" → `MustPass` slice; H2 "Scoring Rubric" → H3 per dimension → 2-column score table → anchor map. Tolerant to extra whitespace; strict on anchor scores.
- `internal/config/types.go:354-364` — Extend `EvaluatorConfig` struct with `Profiles map[string]string` (default: `{"default":".moai/config/evaluator-profiles/default.md", ...}`), `Aggregation string` (default: `"min"`), `MustPassDimensions []string` (default: `["Functionality", "Security"]`). Existing `MemoryScope string` field preserved verbatim.
- `internal/config/loader.go:223-262` — Extend `LoadHarnessConfig()` to call `cfg.LoadProfiles(profilesDir)` after MemoryScope validation. Loaded profiles cached in returned struct.
- `.claude/agents/moai/evaluator-active.md:78` (after current "## Output Format" section closes at line 77, before "## Evaluator Profile Loading" at line 79) — Insert new "## Hierarchical Score Output (Phase 5)" section. ~35 LOC. Declares structured JSON schema with `schema_version: "v1"`, `dimensions[].criteria[].sub_criteria[].{score, rubric_anchor, evidence, dimension}`. Mandates `rubric_anchor` field per sub-criterion. Cross-references `internal/harness/rubric.go`.
- `.claude/skills/moai-workflow-gan-loop/SKILL.md:118` (between Phase 3 Evaluator Scoring at lines 109-118 and Phase 4 Loop Decision at line 120) — Insert "### Phase 3b: Hierarchical Scoring (REQ-HRN-003-004)" subsection. ~25 LOC. Declares the scoring runner invokes `EvaluatorRunner.Score()` per leaf, aggregates per profile, applies must-pass firewall, persists to Sprint Contract via `WriteContract()`. Cross-references design-constitution §12 Mechanism 1/3.
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` — Template mirror.
- `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md` — Template mirror.

Tasks:
- T4.1 RED: Add failing test for ParseRubricMarkdown loading default.md (AC-HRN-003-07).
- T4.2 GREEN: Implement ParseRubricMarkdown.
- T4.3 REFACTOR: Extract table extractor helper for reuse.
- T4.4 RED: Add failing test for profile-loader integration in LoadHarnessConfig (AC-HRN-003-07).
- T4.5 GREEN: Extend LoadHarnessConfig with LoadProfiles step.
- T4.6 RED: Add failing test for unknown-dimension rejection (AC-HRN-003-08).
- T4.7 GREEN: Implement HRN_UNKNOWN_DIMENSION rejection in Rubric.Validate() (extends M2).
- T4.8 GREEN: Insert evaluator-active body augment section (template-first; mirror to local).
- T4.9 GREEN: Insert SKILL.md Phase 3b subsection (template-first; mirror to local).
- T4.10 GREEN: Run `make build` to regenerate `internal/template/embedded.go`. Verify byte-identical local mirrors.

mx_plan tags:
- `@MX:NOTE` on evaluator-active body new section: "Cross-references internal/harness/rubric.go (SPEC-V3R2-HRN-003)".
- `@MX:NOTE` on SKILL.md Phase 3b: "Per design-constitution §12 Mechanism 1; enforces REQ-HRN-003-009 rubric citation".

Exit criteria:
- `go test ./internal/harness/...` + `go test ./internal/config/...` green.
- All 4 shipping `.md` profiles load without error via `LoadHarnessConfig()`.
- AC-HRN-003-07, AC-HRN-003-08, AC-HRN-003-09 PASS.
- Template mirrors byte-identical after `make build`.
- evaluator-active body has new "## Hierarchical Score Output (Phase 5)" section with rubric-citation requirement.
- SKILL.md has new "### Phase 3b: Hierarchical Scoring" subsection.

### M5 — Test fixtures + MX tags + zone-registry mirror entries (per OQ1) — Priority High

Owner: `manager-quality` (T5.1-T5.5 fixtures + MX tag verify) + `manager-spec` (T5.6 zone-registry per OQ1 default)

File:line anchors:
- `internal/harness/testdata/profiles/malformed-5dim.md` (NEW) — fixture declaring a 5th dimension `Performance` for REQ-019 negative test.
- `internal/harness/testdata/profiles/malformed-bypass.md` (NEW) — fixture declaring `must_pass: false` on Security for REQ-018 negative test.
- `internal/harness/testdata/scorecards/2dim-3crit-2sub.json` (NEW) — positive fixture for AC-HRN-003-02 (12 sub-criterion entries).
- `internal/harness/testdata/integration/flat-scorecard-rejected_test.go` (NEW) — integration test for REQ-017 (`HRN_FLAT_SCORECARD_PROHIBITED`).
- `internal/harness/testdata/integration/strict-profile-thresholds_test.go` (NEW) — integration test for AC-HRN-003-12 (strict profile 0.85 threshold; 0.84 fails, 0.86 passes).
- `.claude/rules/moai/core/zone-registry.md` (per OQ1 default) — Append CONST-V3R2-154 (4-dim enum) + CONST-V3R2-155 (4-anchor levels). Both `zone: Frozen`, `canary_gate: true`.

Tasks:
- T5.1 RED: Add failing fixture for AC-HRN-003-02 (12 sub-criterion entries).
- T5.2 GREEN: Implement scorer logic to satisfy fixture (paired with M3 GREEN tasks).
- T5.3 RED: Add failing fixture for AC-HRN-003-08 (5th-dimension rejection).
- T5.4 GREEN: Already implemented in M4 T4.7; verify fixture passes.
- T5.5 GREEN: Add MX tags per §4 above. Verify with `moai mx --validate` (no orphan tags; all `@MX:REASON` sub-lines present for WARN and ANCHOR).
- T5.6 GREEN (per OQ1 default): Append CONST-V3R2-154 + CONST-V3R2-155 entries to `.claude/rules/moai/core/zone-registry.md`. Format follows CONST-V3R2-153 (HRN-002 §11.4.1) precedent.

Decision OQ1 fallback if user prefers deferral: skip T5.6 entirely; document deferral in progress.md "Iteration Log" + open follow-up issue. M5 then completes with T5.1-T5.5 only.

mx_plan tags:
- `@MX:ANCHOR fan_in=3` verification on `EvaluatorRunner.Score()` and acceptance.md AC-HRN-003-04.

Exit criteria:
- All 12 ACs PASS via `go test ./internal/harness/...` + integration tests.
- Coverage ≥85% on `internal/harness/scorer.go` + `internal/harness/rubric.go`.
- `moai mx --validate` clean.
- (per OQ1 default) zone-registry has CONST-V3R2-154 + 155 entries.
- progress.md frontmatter updated to `run_status: complete`.

## 6. Risk Register

Carried verbatim from research.md §12 with mitigation milestone owners:

| # | Risk | Severity | Owner | Mitigation milestone |
|---|------|----------|-------|----------------------|
| R1 | Markdown rubric table parser brittle | MEDIUM | expert-backend | M2 — tolerant parser; M5 — malformed fixtures |
| R2 | evaluator-active inconsistent JSON output | HIGH | manager-docs (agent body) + expert-backend (validator) | M4 — schema declared in body; M3 — REQ-009 strict enforcement; retry-on-reject up to 2 |
| R3 | Rubric authoring burden grows | MEDIUM | (deferred — out of HRN-003 scope) | Default profiles cover 80%+; opt-in per profile |
| R4 | Min-aggregation too strict | MEDIUM | expert-backend | M3 — `aggregation: mean` opt-in per REQ-015; lenient profile uses mean |
| R5 | Sub-criterion count unbounded | MEDIUM | (transitive — SPC-001 MaxDepth=3 caps) | — |
| R6 | Must-pass firewall surprises user | MEDIUM | expert-backend | M3 — explicit failure message in ScoreCard.Rationale citing failing must-pass dimension |
| R7 | Flat-criteria auto-wrap loses info | LOW | (transitive — SPC-001 preserves verbatim) | — |
| R8 | Profile drift template ↔ local | LOW | (already addressed — byte-identical on main) | — |
| R9 | Aggregation rule confusion | LOW | manager-docs | M4 — log effective rule in ScoreCard.Rationale |
| R10 | Profile schema drift `.md` ↔ Go struct | MEDIUM | expert-backend | M2 — Validate() runs on every load; M5 — CI fixture covers all 4 profiles |
| R11 | HRN-001 `EvaluatorConfig` field naming conflict | MEDIUM | manager-spec (coordinate at HRN-001 plan time) | M4 — Decision D7 additive merge convention |
| R12 | Sprint Contract sub-criterion shape breaks HRN-002 leak detection | LOW | (no interaction; leak detection scans substrings, not contract YAML) | — |
| R13 | OQ1 deferral leaves FROZEN constraints unregistered | MEDIUM | manager-spec | M5 — default register CONST-V3R2-154 + 155; fallback document deferral + open issue |

## 7. Open Questions

Replicated from research.md §14 with proposed defaults. plan-auditor should verify defaults against acceptance.md and surface any unresolved OQ as a blocker.

| # | Question | Proposed default | Authority |
|---|----------|------------------|-----------|
| OQ1 | Zone-registry mirror entries: M5 task or follow-up SPEC? | M5 task — register CONST-V3R2-154 (4-dim enum) + CONST-V3R2-155 (4-anchor levels) | research.md §14 OQ1 |
| OQ2 | Aggregation default min vs mean? | CONFIRM `min` per spec.md REQ-007 | research.md §14 OQ2 |
| OQ3 | Must-pass dimensions default set? | `[Functionality, Security]` exported as `DefaultMustPassDimensions`; floor-only `[Security]` (REQ-018 prevents narrowing below) | research.md §14 OQ3 |
| OQ4 | Structured JSON schema versioning? | YES — add `schema_version: "v1"` field, mirrors HRN-002 `LogSchemaVersion` pattern | research.md §14 OQ4 |
| OQ5 | Rubric-anchor citation enforcement: strict reject vs warn? | STRICT reject per REQ-009 + §12 Mechanism 1; retry-on-reject up to 2 per sub-criterion | research.md §14 OQ5 |

## 8. Acceptance Gate

Plan-phase gate:
- [ ] `plan-auditor` independent verification PASS at iteration ≤2.
- [ ] research.md ≥25 file:line anchors — achieved 50.
- [ ] Every REQ in spec.md §5 mapped to ≥1 AC in acceptance.md (19 REQs verified).
- [ ] Every leaf AC carries `(maps REQ-...)` tail.
- [ ] ≥4 of 12 ACs use hierarchical depth-2 schema (AC-03, AC-04, AC-05, AC-07 per Decision D10).
- [ ] AC-HRN-003-04 demonstrates depth-3 grandchildren `.a.i / .a.ii` (MaxDepth=3 boundary).
- [ ] tasks.md uses `T-HRN003-NN` naming with owner role per task.
- [ ] progress.md frontmatter `plan_status: audit-ready`.
- [ ] No spec.md modifications since v0.2.0 (read-only post-audit).

Run-phase gate (future PR):
- [ ] All 12 ACs PASS via `go test ./internal/harness/...`.
- [ ] Coverage ≥85% on `internal/harness/scorer.go` + `internal/harness/rubric.go`.
- [ ] All 4 shipping profile `.md` files load without error.
- [ ] evaluator-active body has hierarchical JSON schema section + rubric-citation requirement.
- [ ] SKILL.md Phase 3b subsection cites `EvaluatorRunner.Score()` and design-constitution §12.
- [ ] Template mirrors byte-identical after `make build`.
- [ ] (per OQ1 default) zone-registry has CONST-V3R2-154 + 155 entries.
- [ ] `moai mx --validate` clean.
- [ ] TRUST 5 — Tested / Readable / Unified / Secured / Trackable all green.

Sync-phase gate (future PR):
- [ ] CHANGELOG entry under `### Added` referencing E-1 (priority 9) + E-3 patterns + R1 §9 citation.
- [ ] docs-site 4-language sync if Hierarchical Scoring docs are publicly exposed (CLAUDE.local.md §17).

## 9. Cross-references

- **SPEC-V3R2-CON-001** — FROZEN/EVOLVABLE zone model + zone-registry.md infrastructure (HRN-003 M5 OQ1 default consumes).
- **SPEC-V3R2-HRN-001** — `HarnessConfig` struct (HRN-002 minimal substrate; HRN-003 M4 extends `EvaluatorConfig`).
- **SPEC-V3R2-HRN-002** — §11.4.1 fresh-judgment amendment (HRN-003 REQ-013 cross-ref); evaluator-active body cross-reference (HRN-003 M4 augment respects).
- **SPEC-V3R2-SPC-001** — Hierarchical AC parser (`internal/spec/ears.go`, `parser.go`, `lint.go`); HRN-003 scorer consumes `[]Acceptance` tree.
- **design-constitution §11.4.1** — Fresh judgment + Sprint Contract durability (HRN-002).
- **design-constitution §12 Mechanism 1** — Rubric Anchoring FROZEN (HRN-003 REQ-009 + REQ-013 source).
- **design-constitution §12 Mechanism 3** — Must-Pass Firewall FROZEN (HRN-003 REQ-008 + REQ-018 source).
- **design-constitution §5** — Pass threshold floor 0.60 FROZEN (HRN-003 REQ-014 source).
- **pattern-library E-1** (priority 9 ADOPT) — Agent-as-a-Judge per-sub-criterion scoring (HRN-003 implementation).
- **pattern-library E-3** (ADOPT) — Rubric-Anchored scoring (HRN-003 + HRN-002 jointly close).
- **major-v3-master §4.5 Layer 5** — Harness + Evaluator (HRN-003 phase context).
- **major-v3-master §11.5 HRN-003** — SPEC catalog entry.
- **CLAUDE.local.md §2** — Template-First discipline.
- **CLAUDE.local.md §18.12** — Plan-in-main doctrine (HRN-003 plan PR cuts from main).

End of plan.
