# Tasks — SPEC-V3R2-HRN-003 Hierarchical Acceptance Scoring

> Run-phase task breakdown derived from plan.md milestones. Naming convention: `T-HRN003-NN`. TDD cycle per task: RED (failing test) → GREEN (minimal implementation) → REFACTOR (clarity). Owner roles use the agent catalog (`manager-spec`, `expert-backend`, `manager-docs`, `manager-quality`). All 19 REQs covered across M2-M5.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec | Initial T-HRN003-01 through T-HRN003-22 task list. TDD cycle structure throughout M2-M5. |

---

## 1. Task Index

| Task ID | Milestone | Owner | Priority | TDD phase | Depends on |
|---------|-----------|-------|----------|-----------|------------|
| T-HRN003-01 | M1 | manager-spec | Critical | — (plan-phase) | — |
| T-HRN003-02 | M1 | manager-spec | Critical | — (plan-phase) | T-01 |
| T-HRN003-03 | M2 | expert-backend | Critical | RED | T-02 |
| T-HRN003-04 | M2 | expert-backend | Critical | GREEN | T-03 |
| T-HRN003-05 | M2 | expert-backend | High | REFACTOR | T-04 |
| T-HRN003-06 | M2 | expert-backend | Critical | RED+GREEN | T-04 |
| T-HRN003-07 | M2 | expert-backend | Critical | RED+GREEN | T-04 |
| T-HRN003-08 | M2 | expert-backend | High | GREEN | T-04 |
| T-HRN003-09 | M3 | expert-backend | Critical | RED+GREEN | T-06 |
| T-HRN003-10 | M3 | expert-backend | Critical | RED+GREEN | T-09 |
| T-HRN003-11 | M3 | expert-backend | Critical | RED+GREEN | T-09 |
| T-HRN003-12 | M3 | expert-backend | High | RED+GREEN | T-07 |
| T-HRN003-13 | M3 | expert-backend | High | RED+GREEN | T-11 |
| T-HRN003-14 | M3 | expert-backend | High | RED+GREEN | T-09 |
| T-HRN003-15 | M3 | expert-backend | Medium | REFACTOR | T-10..T-14 |
| T-HRN003-16 | M4 | expert-backend | Critical | RED+GREEN | T-07 |
| T-HRN003-17 | M4 | expert-backend | High | RED+GREEN | T-16 |
| T-HRN003-18 | M4 | manager-docs | High | GREEN | T-09 |
| T-HRN003-19 | M4 | manager-docs | High | GREEN | T-09 |
| T-HRN003-20 | M5 | manager-quality | High | RED+GREEN | T-09, T-16 |
| T-HRN003-21 | M5 | manager-quality | Critical | GREEN | T-03..T-19 |
| T-HRN003-22 | M5 | manager-spec | High | GREEN | T-21 |

---

## 2. Task Detail

### M1 — Plan-phase artifacts

#### T-HRN003-01 — Author research.md with codebase audit

Owner: `manager-spec`
Priority: Critical
TDD phase: — (plan-phase, no test)

Description: Produce research.md capturing the as-is state of evaluator scoring across `internal/harness/`, evaluator-active body, `.moai/config/evaluator-profiles/`, SPC-001 parser, HRN-001 HarnessConfig substrate, HRN-002 Sprint Contract substrate. Compare against spec.md §5 (19 REQs) and produce gap analysis. Surface 3 drift reconciliations (`.md` profiles, body already cites §11.4.1, no gan_loop.go). Surface OQ1-OQ5 with proposed defaults. Surface D1-D10 plan-time decisions.

Inputs: spec.md, `internal/harness/*.go`, `.claude/agents/moai/evaluator-active.md`, `.claude/skills/moai-workflow-gan-loop/SKILL.md`, `.moai/config/evaluator-profiles/*.md`, dependency SPECs (CON-001, HRN-001, HRN-002, SPC-001).
Outputs: `.moai/specs/SPEC-V3R2-HRN-003/research.md` with ≥25 file:line anchors and §15 D1-D10.
Acceptance: research.md committed; ≥50 anchors verified by plan-auditor; OQ1-OQ5 surfaced with proposed defaults.
Status: DONE (this PR).

#### T-HRN003-02 — Author plan.md, acceptance.md, tasks.md, progress.md, spec-compact.md

Owner: `manager-spec`
Priority: Critical
TDD phase: — (plan-phase)

Description: Produce remaining 5 plan-phase artifacts. acceptance.md self-demonstrates hierarchical schema on ≥4 of 12 ACs; AC-HRN-003-04 reaches depth-3 (`.a.i / .a.ii`). progress.md frontmatter `plan_status: audit-ready`.
Inputs: research.md (T-HRN003-01).
Outputs: 5 markdown files in `.moai/specs/SPEC-V3R2-HRN-003/`.
Acceptance: All 19 REQs mapped; tasks.md uses T-HRN003-NN naming; plan-auditor PASS at iteration ≤2.
Status: DONE (this PR).

### M2 — Type definitions + error sentinels

#### T-HRN003-03 — RED: Add failing tests for Dimension enum + ScoreCard tree shape + Rubric struct

Owner: `expert-backend`
Priority: Critical
TDD phase: RED

Description: Author failing test cases in `internal/harness/scorer_test.go` (NEW) and `internal/harness/rubric_test.go` (NEW) that exercise the not-yet-existing Dimension enum, ScoreCard tree shape, and Rubric struct.

Test cases:
- `TestDimensionEnum_FrozenSet` — verify exactly 4 values; `Dimension(99).IsValid() == false` (REQ-001).
- `TestScoreCard_HierarchicalShape` — instantiate fixture with 2 dim × 3 crit × 2 sub-crit, assert 12 SubCriterion entries (REQ-002, AC-HRN-003-02).
- `TestRubric_AnchorLevelsValid` — instantiate Rubric with 4 canonical anchors, assert Validate() returns nil (REQ-003).
- `TestRubric_AnchorLevelsRejectFifth` — instantiate Rubric with 5 anchors, assert ErrInvalidConfig (REQ-013).
- `TestRubric_AnchorLevelsRejectNonCanonical` — anchors `{0.20, 0.40, 0.60, 0.80}`, assert ErrInvalidConfig.

Files: `internal/harness/scorer_test.go` (NEW, ~80 LOC for these RED cases); `internal/harness/rubric_test.go` (NEW, ~60 LOC).
Outputs: 5 RED tests; expected failures: package does not compile.
Acceptance: `go test ./internal/harness/... -run 'TestDimension|TestScoreCard|TestRubric'` reports compilation failure for missing types.

#### T-HRN003-04 — GREEN: Define Dimension enum + ScoreCard struct family

Owner: `expert-backend`
Priority: Critical
TDD phase: GREEN

Description: Create `internal/harness/scorer.go` (NEW) with minimal types:
- `type Dimension int` + iota constants `Functionality | Security | Craft | Consistency` (4 values, REQ-001).
- `type Verdict string` enum: `pass | fail`.
- `type SubCriterionScore struct { Score float64; RubricAnchor string; Evidence string; Dimension Dimension }`.
- `type CriterionScore struct { Aggregate float64; SubCriteria map[string]SubCriterionScore }`.
- `type DimensionScore struct { Aggregate float64; Criteria map[string]CriterionScore }`.
- `type ScoreCard struct { SchemaVersion string; SpecID string; Dimensions map[Dimension]DimensionScore; Verdict Verdict; Rationale string }`.
- `var DefaultMustPassDimensions = []Dimension{Functionality, Security}` (per OQ3 default).

Files: `internal/harness/scorer.go` (~70 LOC).
Outputs: package compiles; T-03 RED tests for ScoreCard shape now GREEN.
Acceptance: `go test ./internal/harness/... -run 'TestDimensionEnum_FrozenSet|TestScoreCard_HierarchicalShape'` PASS.

mx_plan: Add `@MX:NOTE` on Dimension enum: "FROZEN at 4 values per spec.md REQ-HRN-003-001". Add `@MX:WARN reason="FROZEN-zone constraint per CONST-V3R2-154 (zone-registry mirror entry, registered in M5 T-HRN003-22 per OQ1 default)"` on Dimension enum block.

#### T-HRN003-05 — REFACTOR: Add String() + IsValid() methods on Dimension

Owner: `expert-backend`
Priority: High
TDD phase: REFACTOR

Description: Add `(d Dimension) String() string` mapping to canonical names and `(d Dimension) IsValid() bool` checking range.
Files: `internal/harness/scorer.go` (~20 LOC).
Acceptance: `Dimension(99).IsValid() == false`; `Functionality.String() == "Functionality"`.

#### T-HRN003-06 — RED+GREEN: Define Rubric struct + Validate() with 4-anchor/4-dimension/0.60 floor enforcement

Owner: `expert-backend`
Priority: Critical
TDD phase: RED+GREEN combined

Description: Create `internal/harness/rubric.go` (NEW) with:
- `type Rubric struct { ProfileName string; Dimensions map[Dimension]DimensionRubric; PassThreshold float64; MustPass []Dimension; Aggregation string }`.
- `type DimensionRubric struct { Weight float64; PassThreshold float64; Anchors map[float64]string }`.
- `(r *Rubric) Validate() error` enforcing: (a) exactly 4 dimensions present (REQ-012); (b) exactly 4 anchor levels per dimension at `{0.25, 0.50, 0.75, 1.00}` (REQ-003, REQ-013); (c) `r.PassThreshold >= 0.60` AND each `DimensionRubric.PassThreshold >= 0.60` (REQ-014); (d) `r.Aggregation` in `{"min", "mean"}` (REQ-007); (e) MustPass set MAY be wider than `[Security]` but MAY NOT be narrower (REQ-018, per OQ3 default).

Files: `internal/harness/rubric.go` (~120 LOC); `internal/harness/rubric_test.go` (extend, +40 LOC).
Acceptance: T-03 RED tests for Rubric now GREEN; new tests cover floor rejection and dimension count rejection.

mx_plan: Add `@MX:WARN reason="FROZEN-zone constraint per CONST-V3R2-155 (zone-registry mirror entry registered in M5 T-HRN003-22 per OQ1 default)"` on anchor validator. Add `@MX:NOTE` on `DefaultMustPassDimensions`: "Floor per OQ3 default + design-constitution §12 Mechanism 3".

#### T-HRN003-07 — RED+GREEN: ParseRubricMarkdown for `.md` profile files

Owner: `expert-backend`
Priority: Critical
TDD phase: RED+GREEN combined

Description: Implement `internal/harness/rubric.go` `ParseRubricMarkdown(path string) (*Rubric, error)` (REQ-005). Parses Markdown structure of `.moai/config/evaluator-profiles/default.md` (anatomy in research.md §4.3): H2 "Evaluation Dimensions" table → Weight + PassThreshold per dim; H2 "Must-Pass Criteria" → MustPass slice; H2 "Scoring Rubric" → H3 per dimension → 2-column score table → anchor map.

RED test: `TestParseRubricMarkdown_DefaultProfile` — load `.moai/config/evaluator-profiles/default.md`, assert returned `*Rubric` has 4 dimensions, MustPass=[Functionality, Security], 4 anchors per dimension, PassThreshold floor honored.

GREEN: Implement parser using `goldmark` (already a project dep) or manual line-based extraction (lighter dep). Tolerant to whitespace; strict on anchor scores (must normalize `"1.0"` → `1.00`).

Files: `internal/harness/rubric.go` (+~80 LOC); `internal/harness/rubric_test.go` (+~50 LOC); 4-profile fixture path setup.
Acceptance: AC-HRN-003-07.a + .b PASS; all 4 shipping profiles (`default.md`, `strict.md`, `lenient.md`, `frontend.md`) load without error.

#### T-HRN003-08 — GREEN: Add 4 new error sentinels in `internal/config/errors.go`

Owner: `expert-backend`
Priority: High
TDD phase: GREEN

Description: Append 4 sentinels to existing `ErrEvalMemoryFrozen` (HRN-002 line 44):
- `ErrUnknownDimension` — `"HRN_UNKNOWN_DIMENSION: profile declares dimension outside canonical set {Functionality, Security, Craft, Consistency}"` (REQ-019).
- `ErrRubricCitationMissing` — `"HRN_RUBRIC_CITATION_MISSING: sub-criterion score missing rubric_anchor field or non-canonical anchor (per design-constitution §12 Mechanism 1)"` (REQ-009).
- `ErrFlatScoreCardProhibited` — `"HRN_FLAT_SCORECARD_PROHIBITED: ScoreCard must be hierarchical; flat shape rejected per SPEC-V3R2-HRN-003 REQ-017"` (REQ-017).
- `ErrMustPassBypassProhibited` — `"HRN_MUSTPASS_BYPASS_PROHIBITED: profile attempts to narrow must-pass set below floor [Security] (per design-constitution §12 Mechanism 3)"` (REQ-018).

Files: `internal/config/errors.go` (+~25 LOC).
Acceptance: All 4 sentinels exported; `errors.Is(wrapped, ErrUnknownDimension)` works.

### M3 — Scoring engine + Sprint Contract persistence

#### T-HRN003-09 — RED+GREEN: EvaluatorRunner.Score() with 2dim×3crit×2subcrit fixture

Owner: `expert-backend`
Priority: Critical
TDD phase: RED+GREEN combined

Description: Implement `internal/harness/scorer.go` `type EvaluatorRunner struct { Rubric *Rubric }` + `(r *EvaluatorRunner) Score(contract *SprintContract, artifact Artifact) (*ScoreCard, error)` (REQ-004). Iterates `[]Acceptance` from SPC-001 parser. Per leaf invokes evaluator-active (mocked at unit-test layer via `EvaluatorInvoker` interface).

RED test: `TestEvaluatorRunner_Score_2dim3crit2subcrit` — load fixture `internal/harness/testdata/scorecards/2dim-3crit-2sub.json`, mock evaluator returning canned scores, assert returned ScoreCard has exactly 12 SubCriterion entries (AC-HRN-003-02).

GREEN: Implement Score() with recursive AC tree traversal + mocked invoker.

Files: `internal/harness/scorer.go` (+~100 LOC); `internal/harness/scorer_test.go` (+~80 LOC); `internal/harness/testdata/scorecards/2dim-3crit-2sub.json` (NEW).
Acceptance: AC-HRN-003-02 PASS at unit-test level.

mx_plan: Add `@MX:ANCHOR fan_in=3` on `EvaluatorRunner.Score()`: "Fan-in expected from scorer_test + SKILL.md Phase 3b cross-ref + future HRN-001 router".

#### T-HRN003-10 — RED+GREEN: Aggregation min vs mean

Owner: `expert-backend`
Priority: Critical
TDD phase: RED+GREEN combined

Description: Implement `aggregateMin([]float64) float64` and `aggregateMean([]float64) float64` (REQ-007, REQ-015). Wire into Score() per profile's `Aggregation` field. Per-dimension override supported via `DimensionRubric.Aggregation` optional field.

RED tests:
- `TestAggregateMin_TwoSubScores` — input `{0.8, 0.5}`, expect `0.5` (AC-HRN-003-03.a).
- `TestAggregateMean_TwoSubScores` — input `{0.8, 0.5}`, expect `0.65` (AC-HRN-003-03.b).
- `TestAggregate_PerDimensionOverride` — Functionality=mean, Security=min, both in same ScoreCard (AC-HRN-003-03.c).

Files: `internal/harness/scorer.go` (+~40 LOC); `internal/harness/scorer_test.go` (+~50 LOC).
Acceptance: AC-HRN-003-03.a, .b, .c all PASS.

mx_plan: Add `@MX:NOTE` on `aggregateMin`: "default min per REQ-HRN-003-007".

#### T-HRN003-11 — RED+GREEN: Must-pass firewall

Owner: `expert-backend`
Priority: Critical
TDD phase: RED+GREEN combined

Description: Implement `applyMustPassFirewall(card *ScoreCard, mustPass []Dimension, dimensionThresholds map[Dimension]float64) Verdict` (REQ-008). If any must-pass dimension's `Aggregate < dimensionThreshold`, returns `Verdict("fail")` regardless of other dimensions; populates `card.Rationale` with specific failing dimension and threshold per AC-HRN-003-04.a.ii.

RED tests:
- `TestMustPassFirewall_SecurityFails` — Functionality=1.00, Craft=1.00, Consistency=1.00, Security=0.55, expect fail + rationale citing Security (AC-HRN-003-04.a.i, .a.ii).
- `TestMustPassFirewall_FunctionalityFails` — Security=0.95, Functionality=0.55, expect fail + rationale citing Functionality (AC-HRN-003-04.b).
- `TestMustPassFirewall_AllPassing` — all dims ≥0.60, expect pass (AC-HRN-003-04.c).

Files: `internal/harness/scorer.go` (+~50 LOC); `internal/harness/scorer_test.go` (+~70 LOC).
Acceptance: AC-HRN-003-04.a.i, .a.ii, .b, .c all PASS.

mx_plan: Add `@MX:WARN reason="FROZEN per design-constitution §12 Mechanism 3; bypass requires CON-002 amendment"` on must-pass firewall function.

#### T-HRN003-12 — RED+GREEN: ValidateCitation for rubric-anchor enforcement

Owner: `expert-backend`
Priority: High
TDD phase: RED+GREEN combined

Description: Implement `(r *Rubric) ValidateCitation(score SubCriterionScore) error` (REQ-009). Returns `ErrRubricCitationMissing` if `RubricAnchor` is empty OR not in `{"0.25", "0.50", "0.75", "1.00"}`. Wires into Score() — every leaf score is validated before aggregation; failed citations trigger retry-on-reject (max 2, per OQ5 default).

RED tests:
- `TestValidateCitation_MissingField` — empty `RubricAnchor`, expect ErrRubricCitationMissing (AC-HRN-003-05.a).
- `TestValidateCitation_NonCanonicalAnchor` — `RubricAnchor="0.65"`, expect ErrRubricCitationMissing (AC-HRN-003-05.b).
- `TestValidateCitation_ValidAnchor` — `RubricAnchor="0.75"`, score=0.75, expect nil (AC-HRN-003-05.c).

Files: `internal/harness/rubric.go` (+~30 LOC); `internal/harness/rubric_test.go` (+~50 LOC).
Acceptance: AC-HRN-003-05.a, .b, .c all PASS.

#### T-HRN003-13 — RED+GREEN: Must-pass bypass rejection in Rubric.Validate()

Owner: `expert-backend`
Priority: High
TDD phase: RED+GREEN combined

Description: Extend Rubric.Validate() (T-06) to enforce `MustPass` floor `[Security]` per OQ3 default. If profile sets `MustPass: [Craft]` (excludes Security), reject with `ErrMustPassBypassProhibited` (REQ-018).

RED test: `TestValidate_MustPassBypassRejected` — load fixture `internal/harness/testdata/profiles/malformed-bypass.md`, expect `ErrMustPassBypassProhibited` (AC-HRN-003-11).

Files: `internal/harness/rubric.go` (+~15 LOC); `internal/harness/rubric_test.go` (+~30 LOC); `internal/harness/testdata/profiles/malformed-bypass.md` (NEW, ~30 LOC).
Acceptance: AC-HRN-003-11 PASS.

#### T-HRN003-14 — RED+GREEN: WriteContract() for Sprint Contract sub-criterion persistence

Owner: `expert-backend`
Priority: High
TDD phase: RED+GREEN combined

Description: Implement `internal/harness/scorer.go` `WriteContract(contract *SprintContract, card *ScoreCard, path string) error` (REQ-011). Reads existing contract YAML at path (or creates if absent), extends `acceptance_checklist[]` items with `status` field per item (enum: `passed | failed | refined | new`), writes back atomically. Backward-compatible: items without `status` key are treated as `pending` per existing HRN-002 SKILL.md shape.

RED test: `TestWriteContract_AddsSubCriterionStatus` — Given fixture ScoreCard with 12 sub-criterion entries, write to temp file, re-parse, assert `status` field present per item; assert HRN-002 leak detection (`evaluator_leak.go` `DetectPriorJudgmentLeak()`) does NOT flag the resulting YAML (AC-HRN-003-10).

Files: `internal/harness/scorer.go` (+~60 LOC); `internal/harness/scorer_test.go` (+~50 LOC).
Acceptance: AC-HRN-003-10 PASS.

#### T-HRN003-15 — REFACTOR: Extract Aggregator interface for clarity

Owner: `expert-backend`
Priority: Medium
TDD phase: REFACTOR

Description: Refactor aggregateMin / aggregateMean into an `Aggregator` interface for testability + extensibility (without weakening Decision D7). Keep public function signatures stable.
Files: `internal/harness/scorer.go` (~+30 LOC, -10 LOC net).
Acceptance: All M3 tests still GREEN; no behavior change.

### M4 — Profile loader integration + EvaluatorConfig + agent body augment + SKILL.md update

#### T-HRN003-16 — RED+GREEN: Extend EvaluatorConfig + LoadHarnessConfig with profile loading

Owner: `expert-backend`
Priority: Critical
TDD phase: RED+GREEN combined

Description: Extend `internal/config/types.go:354-364` `EvaluatorConfig` with 3 new fields:
- `Profiles map[string]string` (default: maps profile name to `.moai/config/evaluator-profiles/{name}.md` paths).
- `Aggregation string` (default: `"min"`).
- `MustPassDimensions []string` (default: `["Functionality", "Security"]`).

Extend `internal/config/loader.go:223-262` `LoadHarnessConfig()` to call `cfg.LoadProfiles(profilesDir)` after the existing MemoryScope validation. New method `(cfg *HarnessConfig) LoadProfiles(profilesDir string) (map[string]*harness.Rubric, error)` invokes `harness.ParseRubricMarkdown()` per profile and validates each.

RED test: `TestLoadHarnessConfig_LoadsAllFourProfiles` — load shipping `.moai/config/sections/harness.yaml`, assert all 4 profiles parse + validate successfully (AC-HRN-003-07.c).

Files: `internal/config/types.go` (+~30 LOC); `internal/config/loader.go` (+~40 LOC); `internal/config/loader_test.go` (+~50 LOC).
Acceptance: AC-HRN-003-07.c PASS; existing HRN-001/HRN-002 loader tests still GREEN.

#### T-HRN003-17 — RED+GREEN: Unknown-dimension rejection + malformed profile fixture

Owner: `expert-backend`
Priority: High
TDD phase: RED+GREEN combined

Description: Implement REQ-019 enforcement inside Rubric.Validate() (already started in T-06; this task ensures `HRN_UNKNOWN_DIMENSION` sentinel is wired). Add fixture `internal/harness/testdata/profiles/malformed-5dim.md` declaring a 5th dimension `Performance`.

RED test: `TestParseRubricMarkdown_RejectsFifthDimension` — load fixture, expect `ErrUnknownDimension` (AC-HRN-003-08).

Files: `internal/harness/rubric.go` (+~10 LOC, may already exist from T-06); `internal/harness/rubric_test.go` (+~30 LOC); `internal/harness/testdata/profiles/malformed-5dim.md` (NEW, ~30 LOC).
Acceptance: AC-HRN-003-08 PASS.

#### T-HRN003-18 — GREEN: Augment evaluator-active body with hierarchical JSON output section

Owner: `manager-docs`
Priority: High
TDD phase: GREEN (no test; doc change)

Description: Insert new "## Hierarchical Score Output (Phase 5)" section in `.claude/agents/moai/evaluator-active.md` between current line 77 (close of "## Output Format" section) and line 79 ("## Evaluator Profile Loading" heading). ~35 LOC. Section content:

- Declare structured JSON schema with `schema_version: "v1"` (per OQ4 default).
- Schema example showing `dimensions[].criteria[].sub_criteria[].{score, rubric_anchor, evidence, dimension}`.
- Mandate `rubric_anchor` field per sub-criterion; cite REQ-HRN-003-009.
- Cross-reference `internal/harness/rubric.go` for the validator implementation.

Template-First: insert into `internal/template/templates/.claude/agents/moai/evaluator-active.md` first, then mirror to local. Run `make build` to regenerate `internal/template/embedded.go`.

Files: `internal/template/templates/.claude/agents/moai/evaluator-active.md`, `.claude/agents/moai/evaluator-active.md`, `internal/template/embedded.go` (regenerated).
Acceptance: AC-HRN-003-09 PASS (new section exists at correct location, contains required elements, frontmatter unchanged, body lines 47-54 + 91-92 unchanged).

mx_plan: Add `@MX:NOTE` on the new section: "Cross-references internal/harness/rubric.go (SPEC-V3R2-HRN-003)".

#### T-HRN003-19 — GREEN: Insert Phase 3b subsection in moai-workflow-gan-loop SKILL.md

Owner: `manager-docs`
Priority: High
TDD phase: GREEN

Description: Insert new "### Phase 3b: Hierarchical Scoring (REQ-HRN-003-004)" subsection in `.claude/skills/moai-workflow-gan-loop/SKILL.md` between Phase 3 Evaluator Scoring (lines 109-118) and Phase 4 Loop Decision (line 120). ~25 LOC. Section content:

- Declare scoring runner invokes `EvaluatorRunner.Score()` per leaf.
- Aggregation per profile (`min` default, `mean` opt-in).
- Must-pass firewall application.
- Sprint Contract persistence via `WriteContract()`.
- Cross-reference design-constitution §12 Mechanism 1 (rubric anchoring) + Mechanism 3 (must-pass).

Template-First mirror.

Files: `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md`, `.claude/skills/moai-workflow-gan-loop/SKILL.md`, `internal/template/embedded.go` (regenerated).
Acceptance: SKILL.md Phase 3b cited in evaluator integration tests.

mx_plan: Add `@MX:NOTE` on Phase 3b: "Per design-constitution §12 Mechanism 1; enforces REQ-HRN-003-009 rubric citation".

### M5 — Test fixtures + MX tags + zone-registry mirror entries

#### T-HRN003-20 — RED+GREEN: Strict profile threshold integration test + flat-ScoreCard-prohibited integration test

Owner: `manager-quality`
Priority: High
TDD phase: RED+GREEN combined

Description: Add 2 integration tests:
1. `internal/harness/testdata/integration/strict-profile-thresholds_test.go` — covers AC-HRN-003-12 (strict.md PassThreshold=0.85; 0.84 fails, 0.86 passes).
2. `internal/harness/testdata/integration/flat-scorecard-rejected_test.go` — covers REQ-017 (`HRN_FLAT_SCORECARD_PROHIBITED`); attempts to construct flat ScoreCard via reflection, asserts compile-time + runtime rejection.

Files: 2 new integration test files (~60 LOC each).
Acceptance: AC-HRN-003-12 PASS; REQ-017 enforcement verified.

#### T-HRN003-21 — GREEN: Verify @MX tags + run `moai mx --validate`

Owner: `manager-quality`
Priority: Critical
TDD phase: GREEN

Description: Verify all `@MX` tags placed in M2-M4 are present and well-formed. Per plan.md §4 inventory:
- 5 `@MX:NOTE` tags (scorer.go Dimension enum, rubric.go anchor levels, scorer.go aggregation, evaluator-active.md new section, SKILL.md Phase 3b).
- 3 `@MX:WARN` tags with `@MX:REASON` sub-lines (scorer.go Dimension enum FROZEN, rubric.go anchor validator FROZEN, scorer.go must-pass firewall FROZEN).
- 2 `@MX:ANCHOR fan_in=3` tags (scorer.go EvaluatorRunner.Score(), acceptance.md AC-HRN-003-04 self-demonstrating example).

Run `moai mx --validate` to confirm no orphan tags + all `@MX:REASON` sub-lines present.

Files: verification only; no new files unless missing tags surface.
Acceptance: `moai mx --validate` exit 0; tag inventory matches plan.md §4.

#### T-HRN003-22 — GREEN: Append CONST-V3R2-154 + CONST-V3R2-155 to zone-registry (per OQ1 default)

Owner: `manager-spec`
Priority: High
TDD phase: GREEN

Description: Append 2 new zone-registry mirror entries to `.claude/rules/moai/core/zone-registry.md` per OQ1 default (research.md §14):

```yaml
- id: CONST-V3R2-154
  zone: Frozen
  file: internal/harness/scorer.go
  anchor: "#dimension-enum"
  clause: "4-dimension scoring enum {Functionality, Security, Craft, Consistency} FROZEN per SPEC-V3R2-HRN-003 REQ-001"
  canary_gate: true

- id: CONST-V3R2-155
  zone: Frozen
  file: internal/harness/rubric.go
  anchor: "#anchor-levels"
  clause: "4 rubric anchor levels {0.25, 0.50, 0.75, 1.00} FROZEN per SPEC-V3R2-HRN-003 REQ-003 + design-constitution §12 Mechanism 1"
  canary_gate: true
```

Format follows CONST-V3R2-153 (HRN-002 §11.4.1) precedent.

OQ1 fallback: if user prefers deferral, skip this task; document deferral in progress.md "Iteration Log" + open follow-up issue. M5 then completes with T-20 + T-21 only.

Files: `.claude/rules/moai/core/zone-registry.md` (+~25 LOC).
Acceptance: 2 new CONST entries present; format byte-identical to CONST-V3R2-153 pattern; AC-HRN-003-08 + must-pass tests cross-reference these entries in their failure messages.

---

## 3. Task Dependency Graph

```
T-HRN003-01 (research.md)
    └─ T-HRN003-02 (plan/AC/tasks/progress/compact)   [DONE this PR]
            ├─ T-HRN003-03 (RED: Dimension/ScoreCard/Rubric)
            │      └─ T-HRN003-04 (GREEN: types)
            │             ├─ T-HRN003-05 (REFACTOR: methods)
            │             ├─ T-HRN003-06 (RED+GREEN: Rubric.Validate())
            │             │      ├─ T-HRN003-09 (RED+GREEN: Score())
            │             │      │      ├─ T-HRN003-10 (aggregation)
            │             │      │      ├─ T-HRN003-11 (must-pass firewall)
            │             │      │      ├─ T-HRN003-14 (WriteContract)
            │             │      │      └─ T-HRN003-18 (evaluator-active body)
            │             │      └─ T-HRN003-13 (must-pass bypass rejection)
            │             ├─ T-HRN003-07 (RED+GREEN: ParseRubricMarkdown)
            │             │      ├─ T-HRN003-12 (ValidateCitation)
            │             │      ├─ T-HRN003-16 (LoadHarnessConfig integration)
            │             │      │      └─ T-HRN003-17 (HRN_UNKNOWN_DIMENSION)
            │             │      └─ T-HRN003-19 (SKILL.md Phase 3b)
            │             └─ T-HRN003-08 (error sentinels)
            │                    [enables T-09..T-17 error wrapping]
            └─ T-HRN003-15 (REFACTOR: Aggregator interface) [parallel after T-10..T-14]
                   └─ T-HRN003-20 (integration tests: strict + flat-prohibited)
                          └─ T-HRN003-21 (MX tag verify)
                                 └─ T-HRN003-22 (zone-registry CONST-V3R2-154/155 per OQ1)
```

---

## 4. Owner Role Reference

| Owner | Responsibilities |
|-------|------------------|
| `manager-spec` | Plan-phase artifact authoring (T-01, T-02); zone-registry CONST entries (T-22 per OQ1 default). |
| `expert-backend` | Go-level implementation: types (T-03..T-08); scoring engine (T-09..T-15); profile loader integration (T-16, T-17). |
| `manager-docs` | Documentation amendments: evaluator-active body augment (T-18); SKILL.md Phase 3b (T-19). Both via Template-First. |
| `manager-quality` | Run-phase quality gate: integration tests (T-20); MX tag verification (T-21). |

---

## 5. Run-Phase Entry Criteria

This tasks.md becomes actionable when:
- [ ] Plan PR (this) merges to main.
- [ ] plan-auditor PASS recorded.
- [ ] `expert-backend` (or `manager-tdd` delegating to it) assigned to lead M2-M4 execution in a fresh worktree (`moai worktree new SPEC-V3R2-HRN-003 --base origin/main`).
- [ ] HRN-001 status is at minimum `draft` (HRN-003 M4 may need to coordinate `EvaluatorConfig` field naming with HRN-001 plan-phase author per Decision D7); HRN-001 does NOT need to be merged before HRN-003 run starts.

End of tasks.
