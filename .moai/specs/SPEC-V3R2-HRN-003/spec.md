---
id: SPEC-V3R2-HRN-003
title: "Hierarchical Acceptance Scoring (4-dimension × sub-criteria)"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 5 — Harness + Evaluator"
module: "internal/harness/scorer.go, .claude/agents/moai/evaluator-active.md, .moai/config/evaluator-profiles/, .moai/specs/SPEC-*/acceptance.md"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-HRN-001
  - SPEC-V3R2-HRN-002
  - SPEC-V3R2-SPC-001
related_problem: []
related_theme: "Layer 5 — Harness, Master §4.5, §Appendix C E-1 E-3, §11 SPEC-V3R2-HRN-003"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "evaluator, scoring, hierarchical, acceptance-criteria, rubric, 4-dimension, agent-as-judge, v3r2"
---

# SPEC-V3R2-HRN-003: Hierarchical Acceptance Scoring (4-dimension × sub-criteria)

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-23 | GOOS   | Initial draft (Wave 4 SPEC writer, round 2) |

---

## 1. Goal (목적)

Current MoAI evaluator-active produces flat `ScoreCard` structures with a single numeric score per artifact. R1 §9 Agent-as-a-Judge (Zhuge et al. 2024) introduces a hierarchical scoring shape where a single acceptance criterion decomposes into sub-criteria (365-sub-requirement tree in their benchmark). This shape aligns with how v3r2 SPECs already express nested acceptance (`AC-X-01.a`, `AC-X-01.b` — introduced in SPEC-V3R2-SPC-001).

Independently, MoAI's evaluator-profiles documented in design-constitution §11 include a 4-dimension scoring framework: **Functionality / Security / Craft / Consistency**. Each dimension currently returns a single flat score; this SPEC extends the scorer so that each dimension evaluates every sub-criterion independently, producing a hierarchical ScoreCard that preserves sub-criterion granularity.

This SPEC:

- Defines the canonical 4-dimension scoring framework: Functionality, Security, Craft, Consistency (already named in design-constitution §11 rubrics; this SPEC formalizes them).
- Defines the hierarchical ScoreCard struct: dimension → criterion → sub-criterion → score.
- Implements the scorer in `internal/harness/scorer.go` respecting the fresh-judgment semantics from SPEC-V3R2-HRN-002 (no cross-iteration reasoning leak).
- Publishes per-harness-level rubric files in `.moai/config/evaluator-profiles/` with anchored scores at 0.25 / 0.50 / 0.75 / 1.00 (pattern E-3 from pattern-library.md).
- Updates evaluator-active agent body with the hierarchical scoring contract.
- Wires the scorer to the Sprint Contract so that passed sub-criteria carry forward per HRN-002 §12 durability rule.

### 1.1 Background

Master §Appendix C (Pattern to SPEC mapping):
- E-1 Agent-as-a-Judge Intermediate scoring → HRN-002 + HRN-003 (ADOPT priority 9).
- E-3 Rubric-Anchored + Independent Re-eval → HRN-003 (ADOPT).

Pattern E-1 description (from pattern-library.md): "Agent-as-a-Judge scores per-criterion per sub-criterion (Agent-as-a-Judge §9 shape); rubric files in `.moai/config/evaluator-profiles/` per harness level."

Pattern E-3: "Every evaluation criterion has a concrete rubric with examples of scores at 0.25, 0.50, 0.75, and 1.0. evaluator-active MUST reference the rubric when assigning scores. Scores without rubric justification are invalid." (already in design-constitution §12 Mechanism 1 Rubric Anchoring).

Dimension choice comes from `.moai/config/evaluator-profiles/` existing files referencing four dimensions in rubric examples. This SPEC makes the 4-dimension schema canonical and required.

Hierarchical shape:

```
ScoreCard
├── Dimension: Functionality
│   ├── Criterion AC-X-01
│   │   ├── Sub-criterion 01.a → score 0.0-1.0 + rubric anchor + evidence
│   │   ├── Sub-criterion 01.b → ...
│   │   └── aggregate_score (mean or min per profile)
│   └── Criterion AC-X-02 → ...
├── Dimension: Security
│   └── ...
├── Dimension: Craft
│   └── ...
└── Dimension: Consistency
    └── ...
```

Sub-criteria flatten to single-level when the SPEC's acceptance.md uses flat Given/When/Then (backward compat from SPEC-V3R2-SPC-001 §BC-V3R2-011 wrapping flat criteria as single-level children).

*Source: R1 §9 Agent-as-a-Judge (365-sub-requirement shape); pattern-library.md E-1 (priority 9), E-3 (rubric anchoring); design-principles.md Principle 4; design-constitution §11 (existing 4-dimension rubric references); major-v3-master.md §4.5 Layer 5, §11.5 HRN SPECs.*

### 1.2 Non-Goals

- Changing the 4 dimension names (Functionality, Security, Craft, Consistency are canonical for v3.0; new dimensions require CON-002 amendment).
- Implementing acceptance.md parsing/linting (SPEC-V3R2-SPC-001 and SPEC-V3R2-SPC-003 own the parser; this SPEC consumes the parsed tree).
- Changing the Sprint Contract file format (SPEC-V3R2-HRN-002 owns).
- Generating rubric content from scratch (this SPEC defines the schema; specific rubric texts are authored as part of each evaluator profile).
- Implementing the GAN loop runner (SPEC-V3R2-HRN-002 owns).
- Changing pass_threshold floor (stays 0.60 FROZEN from HRN-002).
- Per-dimension weighting configuration (all dimensions equal-weight; weighting deferred to Master §12 post-telemetry).
- Auto-generating sub-criteria from flat acceptance (the SPC-001 migrator wraps existing flat criteria).
- Scoring skill files or agent files (only artifact code produced by the GAN loop).

---

## 2. Scope (범위)

### 2.1 In Scope

- Define the canonical 4-dimension enum in `internal/harness/scorer.go`: `Functionality | Security | Craft | Consistency`.
- Define the hierarchical `ScoreCard` struct with nested Dimension → Criterion → SubCriterion scoring.
- Define the `Rubric` struct with anchored scores at 0.25, 0.50, 0.75, 1.00 per sub-criterion.
- Implement `EvaluatorRunner.Score(contract, artifact) ScoreCard` that:
  - iterates the hierarchical acceptance tree from SPEC-V3R2-SPC-001,
  - for each sub-criterion, evaluator-active emits `{score, rubric_anchor, evidence, dimension}`,
  - aggregates sub-criterion scores per criterion (aggregation rule: min by default; mean available via profile flag),
  - aggregates criterion scores per dimension (same rule).
- Author rubric template files `.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.yaml`:
  - each profile declares per-dimension rubric templates with 4 anchor levels,
  - strict profile has tighter 0.75 → 0.85 minimum pass bar for must-pass dimensions,
  - lenient profile softer for exploratory specs.
- Update `evaluator-active` agent body with the hierarchical scoring contract:
  - agent MUST emit structured JSON output carrying the hierarchical score tree,
  - agent MUST cite the rubric anchor for every sub-criterion score,
  - scores without rubric citation are rejected.
- Wire scorer to Sprint Contract: passed sub-criteria persist to `.moai/sprints/{spec-id}/contract.yaml` per HRN-002 §12 durability rule.
- Implement aggregation rules in `scorer.go`:
  - default: criterion aggregate = min of sub-criterion scores (any sub-criterion failure fails the criterion),
  - profile `mean_aggregate: true` allows mean of sub-criterion scores,
  - dimension aggregate = min of criterion aggregates (any criterion failure fails the dimension).
- Implement must-pass firewall per design-constitution §12 Mechanism 3: even perfect scores on other dimensions cannot compensate a failing must-pass dimension.
- Backward compat with flat acceptance: auto-wrap flat Given/When/Then as single-level children per BC-V3R2-011.
- Template-first: all file edits land in template tree first; `make build` regenerates; local tree byte-identical.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Defining specific rubric text for every domain (this SPEC defines schema; rubric authors extend per SPEC).
- Per-SPEC dimension weighting (all four dimensions equal weight).
- Adding dimensions beyond the 4 canonical names.
- Changing pass_threshold floor (stays 0.60 FROZEN per HRN-002).
- Cross-SPEC score aggregation (scope is per-SPEC).
- Historical score comparison (regression baseline is in design-constitution §12 Mechanism 2; implementation out of scope here).
- Score visualization (web dashboard deferred post-v3.0).
- Auto-refining sub-criteria (refinement is Sprint Contract's responsibility, not scorer's).
- Evaluating artifacts beyond code (e.g., scoring prose copy) — separate profile concerns.
- Per-iteration score delta reporting (future telemetry feature).

---

## 3. Environment (환경)

- Runtime: moai-adk-go v3.0.0-beta.1+ (Phase 5)
- Claude Code v2.1.111+ (Opus 4.7 Adaptive Thinking required for rubric-anchored judgment; agent effort: xhigh per SPEC-V3R2-ORC-003 for evaluator-active)
- Canonical dimensions: `Functionality`, `Security`, `Craft`, `Consistency` (FROZEN enum for v3.0)
- Rubric anchor levels: 0.25, 0.50, 0.75, 1.00 (FROZEN per design-constitution §12 Mechanism 1)
- Evaluator profile files location: `.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.yaml`
- ScoreCard output format: structured JSON consumable by the GAN loop runner
- Sprint Contract integration: `.moai/sprints/{spec-id}/contract.yaml` per HRN-002
- Backward compat: flat acceptance trees from pre-v3r2 SPECs auto-wrap as single-level children (BC-V3R2-011 from Master §8)
- Must-pass firewall: design-constitution §12 Mechanism 3 (FROZEN)

---

## 4. Assumptions (가정)

- SPEC-V3R2-CON-001 has landed; FROZEN zones enforced.
- SPEC-V3R2-HRN-001 has landed; `HarnessConfig` struct and evaluator profile loading are live.
- SPEC-V3R2-HRN-002 has landed; evaluator memory is per-iteration; Sprint Contract carries state.
- SPEC-V3R2-SPC-001 has landed; hierarchical acceptance tree is parseable; flat criteria auto-wrap to single-level.
- The 4 dimensions (Functionality, Security, Craft, Consistency) are authoritative and frozen for v3.0; any new dimension requires CON-002 amendment.
- Aggregation default `min` is the safer choice (any sub-criterion failure fails the criterion); profiles can opt-in to `mean`.
- Rubric authoring is a separate PR per evaluator profile; this SPEC ships schema + test fixtures + one default profile template.
- evaluator-active agent reliably emits structured JSON when prompted (Claude Opus 4.7 structured-output fidelity; test fixtures validate).
- Must-pass firewall preserves even when the scorer aggregation produces high overall score but a must-pass dimension fails.
- No third-party evaluator replaces evaluator-active in v3r2.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-HRN-003-001 (Ubiquitous) — 4-차원 열거형**
The file `internal/harness/scorer.go` **shall** define the Dimension enum `{Functionality, Security, Craft, Consistency}` as a FROZEN 4-value set for v3.0.

**REQ-HRN-003-002 (Ubiquitous) — 계층 스코어카드 구조**
The `ScoreCard` struct **shall** contain nested fields: `Dimensions map[Dimension]DimensionScore`; each `DimensionScore` contains `Criteria map[string]CriterionScore` keyed by criterion ID; each `CriterionScore` contains `SubCriteria map[string]SubCriterionScore`; each `SubCriterionScore` contains fields `{Score float64, RubricAnchor string, Evidence string, Dimension Dimension}`.

**REQ-HRN-003-003 (Ubiquitous) — 루브릭 구조**
The `Rubric` struct **shall** declare 4 anchor levels at scores 0.25, 0.50, 0.75, 1.00 and **shall** include a descriptive text per anchor explaining what observed behavior earns that score.

**REQ-HRN-003-004 (Ubiquitous) — Score 함수**
The function `EvaluatorRunner.Score(contract *SprintContract, artifact Artifact) (*ScoreCard, error)` **shall** iterate the hierarchical acceptance tree, invoke evaluator-active per sub-criterion, collect `{Score, RubricAnchor, Evidence, Dimension}`, and aggregate per REQ-007.

**REQ-HRN-003-005 (Ubiquitous) — 평가 프로필 기본 셋**
The directory `.moai/config/evaluator-profiles/` **shall** contain at minimum 4 profiles: `default.yaml`, `strict.yaml`, `lenient.yaml`, `frontend.yaml`; each profile **shall** include per-dimension rubric templates with the 4 anchor levels.

**REQ-HRN-003-006 (Ubiquitous) — 에이전트 바디 수정**
The file `.claude/agents/moai/evaluator-active.md` body **shall** declare the hierarchical scoring contract including: (a) 4-dimension enumeration, (b) rubric-anchored scoring requirement (Mechanism 1), (c) structured JSON output format, (d) fresh-iteration respawn per HRN-002.

### 5.2 Event-Driven (이벤트 기반)

**REQ-HRN-003-007 (Event-Driven) — 집계 규칙**
**When** the scorer aggregates sub-criterion scores to a criterion score, the default rule **shall** be `min(sub_scores)`; when the profile declares `aggregate: mean`, the rule **shall** be `mean(sub_scores)`.

**REQ-HRN-003-008 (Event-Driven) — Must-Pass 방화벽**
**When** a dimension is declared `must_pass: true` in the evaluator profile and the dimension's aggregate score is below its dimension-specific `pass_threshold`, the overall ScoreCard verdict **shall** be `fail` regardless of other dimensions' scores (design-constitution §12 Mechanism 3, FROZEN).

**REQ-HRN-003-009 (Event-Driven) — 루브릭 미인용 거부**
**When** evaluator-active emits a sub-criterion score without a `rubric_anchor` field referencing one of the 4 anchor labels (0.25 / 0.50 / 0.75 / 1.00), the scorer **shall** reject the score with error `HRN_RUBRIC_CITATION_MISSING` and require re-evaluation.

**REQ-HRN-003-010 (Event-Driven) — 평면 기준 자동 래핑**
**When** a SPEC's acceptance.md contains flat Given/When/Then criteria (no nested sub-criteria), the scorer's tree builder **shall** auto-wrap each flat criterion as a single-child tree node with the parent criterion's ID and sub-criterion ID `.01` (backward compat per BC-V3R2-011).

**REQ-HRN-003-011 (Event-Driven) — Sprint Contract 승계**
**When** the scorer completes an iteration, it **shall** persist the hierarchical ScoreCard to the Sprint Contract at `.moai/sprints/{spec-id}/contract.yaml` with per-sub-criterion `{status: passed|failed|refined|new}` per HRN-002 §12.

### 5.3 State-Driven (상태 기반)

**REQ-HRN-003-012 (State-Driven) — 차원 고정**
**While** the v3.0.0 minor cycle is active, only the 4 canonical dimensions **shall** be valid; any profile declaring a fifth dimension **shall** fail loader validation with error `HRN_UNKNOWN_DIMENSION`.

**REQ-HRN-003-013 (State-Driven) — 앵커 레벨 고정**
**While** the runtime is active, rubric anchors **shall** be exactly the 4 values `{0.25, 0.50, 0.75, 1.00}`; no profile **shall** introduce additional anchors.

**REQ-HRN-003-014 (State-Driven) — 통과 임계 floor 준수**
**While** the runtime is active, every profile's `pass_threshold` (per dimension or overall) **shall** be ≥ 0.60 (FROZEN floor from HRN-002 REQ-012 + design-constitution §5).

### 5.4 Optional (선택)

**REQ-HRN-003-015 (Optional) — 평균 집계 옵션**
**Where** a profile declares `aggregate: mean` in a specific dimension's configuration, the scorer **may** apply mean aggregation to that dimension only while using default min aggregation for other dimensions in the same profile.

**REQ-HRN-003-016 (Optional) — 프론트엔드 프로필 확장**
**Where** the profile `frontend.yaml` is active, the rubric templates **may** include UI-specific sub-criteria (viewport responsiveness, accessibility, animation smoothness) in the `Craft` dimension.

### 5.5 Unwanted Behavior

**REQ-HRN-003-017 (Unwanted Behavior) — 평면 스코어 재도입 금지**
**If** the GAN loop runner invokes a version of EvaluatorRunner.Score that returns a flat (non-hierarchical) ScoreCard, **then** CI integration test **shall** fail with error `HRN_FLAT_SCORECARD_PROHIBITED`.

**REQ-HRN-003-018 (Unwanted Behavior) — Must-Pass 우회 금지**
**If** a profile attempts to disable the must-pass firewall by setting `must_pass: false` on a dimension that design-constitution §12 Mechanism 3 declares must-pass (FROZEN Security + Functionality by default), **then** the loader **shall** reject with error `HRN_MUSTPASS_BYPASS_PROHIBITED`.

**REQ-HRN-003-019 (Unwanted Behavior) — 새로운 차원 추가 금지**
**If** an evaluator profile YAML declares a dimension other than the 4 canonical names, **then** loader **shall** fail with `HRN_UNKNOWN_DIMENSION` (REQ-012).

---

## 6. Acceptance Criteria (수용 기준 요약)

Detailed Given-When-Then scenarios are in `acceptance.md`.

Core criteria:

- **AC-HRN-003-01**: `go test ./internal/harness/...` passes with ≥ 85% coverage including scorer aggregation logic tests.
- **AC-HRN-003-02**: Running `EvaluatorRunner.Score` on a fixture SPEC with 2 dimensions × 3 criteria × 2 sub-criteria produces a ScoreCard with 2 × 3 × 2 = 12 sub-criterion score entries.
- **AC-HRN-003-03**: Setting aggregate `min` and providing sub-scores {0.8, 0.5} yields criterion score 0.5; setting `mean` yields 0.65.
- **AC-HRN-003-04**: Must-pass Security dimension with score 0.55 (below 0.60 floor) fails the overall ScoreCard regardless of other dimensions scoring 1.00.
- **AC-HRN-003-05**: evaluator-active output lacking `rubric_anchor` field triggers `HRN_RUBRIC_CITATION_MISSING`.
- **AC-HRN-003-06**: Flat acceptance criteria from a pre-v3r2 SPEC auto-wrap as single-level children without loss.
- **AC-HRN-003-07**: The 4 evaluator profiles (default, strict, lenient, frontend) exist at `.moai/config/evaluator-profiles/` and parse successfully.
- **AC-HRN-003-08**: Adding a fifth dimension `Performance` to a profile YAML triggers loader error `HRN_UNKNOWN_DIMENSION`.
- **AC-HRN-003-09**: evaluator-active body declares the 4-dimension contract and structured JSON output format.
- **AC-HRN-003-10**: Sprint Contract file `.moai/sprints/{spec-id}/contract.yaml` contains sub-criterion states after one iteration (status: passed|failed|refined|new per sub-criterion).
- **AC-HRN-003-11**: Attempting to set `must_pass: false` on Security or Functionality in a profile triggers `HRN_MUSTPASS_BYPASS_PROHIBITED`.
- **AC-HRN-003-12**: Strict profile sets dimension pass_threshold 0.85 for must-pass Security; scoring 0.84 fails; scoring 0.86 passes.

---

## 7. Constraints (제약)

- [HARD] FROZEN 4-dimension set (REQ-001, REQ-012).
- [HARD] FROZEN 4 rubric anchors at {0.25, 0.50, 0.75, 1.00} (REQ-003, REQ-013).
- [HARD] FROZEN pass_threshold floor 0.60 (REQ-014, cross-ref HRN-002).
- [HARD] FROZEN must-pass firewall (REQ-008, REQ-018) per design-constitution §12 Mechanism 3.
- [HARD] Fresh-judgment semantics preserved (REQ-006 cites HRN-002) — no cross-iteration reasoning leak.
- [HARD] Template-First (CLAUDE.local.md §2).
- [HARD] Backward compat with flat acceptance (REQ-010) — no breaking migration required for v2 SPECs.
- [HARD] Rubric citation required for every score (REQ-009, Mechanism 1 compliance).
- [HARD] No frontmatter change on evaluator-active (only body text); no new agent tools.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk                                                              | Impact | Mitigation                                                                                                  |
|-------------------------------------------------------------------|--------|-------------------------------------------------------------------------------------------------------------|
| Evaluator-active inconsistent structured JSON output              | HIGH   | REQ-009 strict rubric-anchor enforcement; retry-on-reject pattern up to 2 retries per sub-criterion         |
| Rubric authoring burden per SPEC grows                            | MEDIUM | Default profile templates cover 80%+ of SPECs; rubric authoring is opt-in via profile selection             |
| Min-aggregation too strict for exploratory SPECs                  | MEDIUM | Profile `aggregate: mean` opt-in per dimension (REQ-015); lenient profile uses mean by default              |
| Sub-criterion count grows unbounded on complex SPECs              | MEDIUM | SPEC-V3R2-SPC-003 lint can cap sub-criterion depth at 3 levels                                              |
| Must-pass firewall surprises user when overall score is high       | MEDIUM | Output the firewall trigger message explicitly in ScoreCard verdict; cite the failing must-pass dimension   |
| Flat-criteria auto-wrap loses information                         | LOW    | Single-level wrap preserves ID + Given/When/Then as-is; test fixtures verify lossless roundtrip             |
| Profile drift between template and local `.moai/config/evaluator-profiles/` | MEDIUM | `diff -r` CI gate per CLAUDE.local.md §2                                                                    |
| Aggregation rule confusion (min vs mean)                          | LOW    | Default min is documented; profile explicitly opt-in mean; log the effective rule in ScoreCard rationale    |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE zone model)
- **SPEC-V3R2-HRN-001** (HarnessConfig struct and evaluator profile loader)
- **SPEC-V3R2-HRN-002** (Evaluator memory scope amendment — fresh judgment per iteration)
- **SPEC-V3R2-SPC-001** (Hierarchical acceptance criteria parser)

### 9.2 Blocks

- **SPEC-V3R2-WF-003** (Multi-mode router — thorough harness mode requires hierarchical scoring)
- **SPEC-V3R2-MIG-001** (migrator rewrites legacy SPEC acceptance trees; this SPEC consumes the rewritten output)

### 9.3 Related

- **SPEC-V3R2-SPC-003** (SPEC linter — validates acceptance hierarchy)
- **SPEC-V3R2-EVAL-001** (v3-legacy evaluator profile — content authors extend here)
- **SPEC-V3R2-CON-002** (Amendment protocol — adding a 5th dimension would pass through graduation)

---

## 10. Traceability (추적성)

- REQ-to-AC mapping: REQ-001 → AC-01, AC-08; REQ-002 → AC-02; REQ-003 → AC-07; REQ-004 → AC-02, AC-03; REQ-005 → AC-07; REQ-006 → AC-09; REQ-007 → AC-03; REQ-008 → AC-04, AC-12; REQ-009 → AC-05; REQ-010 → AC-06; REQ-011 → AC-10; REQ-012 → AC-08; REQ-013 → AC-07 (rubric-anchor test); REQ-014 → AC-12; REQ-015 → AC-03; REQ-016 → frontend profile regression; REQ-017 → AC-02 type test; REQ-018 → AC-11; REQ-019 → AC-08.
- Total REQ count: 19 (Ubiquitous 6, Event-Driven 5, State-Driven 3, Optional 2, Unwanted 3)
- Expected AC count: 12
- Wave 1/2 sources:
  - `r1-ai-harness-papers.md` §9 Agent-as-a-Judge (hierarchical 365-sub-req shape)
  - `pattern-library.md` E-1 (ADOPT priority 9, intermediate scoring + fresh memory), E-3 (Rubric-Anchored + Independent Re-eval)
  - `design-principles.md` Principle 4 (Evaluator Fresh, Contract State Durable)
  - `design-constitution.md` §11 Rubric Anchoring, §12 Mechanism 1 (Rubric Anchoring), Mechanism 3 (Must-Pass Firewall — FROZEN), Mechanism 4 (Independent Re-evaluation — FROZEN)
  - `major-v3-master.md` §4.5 Layer 5, §Appendix C (E-1 + E-3 mapping), §11.5 HRN-003
- Code-side paths:
  - `internal/harness/scorer.go` (new, REQ-001..004, REQ-007, REQ-008, REQ-009, REQ-010)
  - `internal/harness/scorer_test.go` (new, AC regression fixtures)
  - `internal/harness/rubric.go` (new, REQ-003)
  - `internal/config/types.go` (modified, EvaluatorProfile struct extensions, REQ-012, REQ-013, REQ-014)
  - `.moai/config/evaluator-profiles/default.yaml` (new or modified, REQ-005)
  - `.moai/config/evaluator-profiles/strict.yaml` (new or modified, REQ-005)
  - `.moai/config/evaluator-profiles/lenient.yaml` (new or modified, REQ-005)
  - `.moai/config/evaluator-profiles/frontend.yaml` (new or modified, REQ-005, REQ-016)
  - `.claude/agents/moai/evaluator-active.md` (modified, REQ-006)
  - `internal/harness/gan_loop.go` (modified — wires scorer, cross-ref HRN-002 REQ-009)
  - `internal/template/templates/...` (template-first mirrors)

---

End of SPEC.
