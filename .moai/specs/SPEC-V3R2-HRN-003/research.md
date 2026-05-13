# Research — SPEC-V3R2-HRN-003 Hierarchical Acceptance Scoring

> Phase 0.5 deep codebase research preceding plan.md.
> Captures the as-is state of evaluator scoring across `internal/harness/`, `.claude/agents/moai/evaluator-active.md`, `.moai/config/evaluator-profiles/`, the SPC-001 hierarchical AC parser, the HRN-001 HarnessConfig surface, and the HRN-002 fresh-judgment substrate. Establishes the gap delta against spec.md §5 (19 REQs) and surfaces five Open Questions with proposed defaults.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec (HRN-003 plan author) | Initial research for HRN-003 plan phase — codebase audit confirms three drift reconciliations (a) profiles ship as `.md` not `.yaml` per spec.md §5.1 REQ-005 assumption; (b) evaluator-active body already cites §11.4.1 from HRN-002 M3 (lines 91-92); (c) `internal/harness/gan_loop.go` does NOT exist on main → REQ-011 wires via SKILL.md per HRN-002 D1 precedent. |

---

## 1. Research Goal

Establish ground truth for SPEC-V3R2-HRN-003 plan:

1. Audit current evaluator-scoring architecture across `internal/harness/`, the `.claude/agents/moai/evaluator-active.md` agent body, `.moai/config/evaluator-profiles/`, and the SPC-001 hierarchical AC parser surface.
2. Reconcile spec.md (authored 2026-04-23) against post-HRN-002 main state (commit `0ac27ee4e`, 2026-05-13).
3. Quantify gap delta against spec.md §5 (19 REQs across 5 EARS modalities).
4. Surface adjacent SPECs (CON-001, HRN-001, HRN-002, SPC-001) and the precise hook points each provides.
5. Derive milestone breakdown for plan.md (M1–M5 mirror of HRN-002 shape).
6. Identify Open Questions (OQ1..OQ5) with proposed defaults.

This research precedes plan.md per `.claude/skills/moai-workflow-spec` Phase 0.5 protocol.

---

## 2. As-Is — `internal/harness/` package

### 2.1 Existing files (as of 2026-05-13)

`ls -la /Users/goos/MoAI/moai-adk-go/internal/harness/` confirms 33 files. Relevant inventory:

- `internal/harness/types.go` (11.8KB) — `LogSchemaVersion`, `Event`, `EventType`, `Pattern`, `Tier`, `Promotion`, `Proposal`, `Decision`, `Session`, `CanaryResult`, `ContradictionItem`, `ContradictionReport`, `OversightOption`, `OversightProposal`. **No `Dimension`, `ScoreCard`, `Rubric`, `EvaluatorRunner`, or `SubCriterionScore` types.** Confirmed greenfield surface for HRN-003.
- `internal/harness/evaluator_leak.go` (2.6KB) — landed by HRN-002 M3. Exports `ErrPriorJudgmentLeak`, `DetectPriorJudgmentLeak()`. Out of scope for HRN-003 (REQ-011 fresh-respawn semantic flows through this validator).
- `internal/harness/evaluator_leak_test.go` (3.7KB) — leak detection regression test landed by HRN-002.
- `internal/harness/applier.go` + `chaining_rules.go` + `cleanup.go` + `interview.go` + `interview_writer.go` + `prefix_conflict.go` + `learner.go` + `observer.go` + `retention.go` — Phase 1-4 learning/safety pipeline; not in HRN-003 scope.
- `internal/harness/layer1.go` through `layer5.go` — CON-002 5-layer safety pipeline (FrozenGuard / Canary / ContradictionDetector / RateLimiter / HumanOversight). HRN-003 consumes (not modifies).
- `internal/harness/safety/` directory — sub-package; out of HRN-003 scope.

### 2.2 Files HRN-003 introduces (NEW)

Per spec.md §10 traceability code-side paths:

- `internal/harness/scorer.go` (NEW, ~250 LOC) — `Dimension` enum, `ScoreCard`, `DimensionScore`, `CriterionScore`, `SubCriterionScore` types; `EvaluatorRunner.Score()` function; aggregation (`min` default, `mean` opt-in); must-pass firewall.
- `internal/harness/rubric.go` (NEW, ~120 LOC) — `Rubric` struct with 4 anchor levels {0.25, 0.50, 0.75, 1.00}; rubric loader for `.md` profile files; rubric-anchor citation validator (REQ-009 enforcement).
- `internal/harness/scorer_test.go` (NEW, ~400 LOC) — unit tests covering AC fixtures (12 ACs).
- `internal/harness/rubric_test.go` (NEW, ~150 LOC) — rubric loader + anchor validator tests.

### 2.3 Files HRN-003 does NOT create

- `internal/harness/gan_loop.go` — DOES NOT EXIST on main. Confirmed via `ls /Users/goos/MoAI/moai-adk-go/internal/harness/gan_loop.go: No such file or directory`. HRN-002 D1 declared the orchestrator-level runner (`.claude/skills/moai-workflow-gan-loop/SKILL.md`) is the actual integration point; no Go-side runner module exists. HRN-003 inherits this decision: REQ-011 (Sprint Contract sub-criterion persistence) wires via SKILL.md text + Go-side scorer-emits-YAML helper, not via a new gan_loop.go module. See Decision D6 below.

### 2.4 Test coverage baseline

Existing `internal/harness/` tests (`applier_test.go`, `evaluator_leak_test.go`, `layer1-5_test.go`, etc.) currently green. HRN-003 must add ≥85% coverage on the new `scorer.go` + `rubric.go` files (constraint per spec.md §3 environment line — same target as HRN-001 §3).

---

## 3. As-Is — `.claude/agents/moai/evaluator-active.md` agent body

### 3.1 Frontmatter (preserved verbatim by HRN-003)

`.claude/agents/moai/evaluator-active.md:1-26` declares:
- `tools: Read, Grep, Glob, Bash, mcp__sequential-thinking__sequentialthinking` — no Agent tool, consistent with subagent role.
- `model: sonnet`
- `effort: high`
- `permissionMode: plan`
- `memory: project` — disk-level project memory, distinct from per-iteration LLM context (HRN-002 §11.4.1 clarification still load-bearing).
- `skills: moai-foundation-core, moai-foundation-quality`
- `SubagentStop` hook fires `evaluator-completion`.

[HARD per spec.md §7] HRN-003 makes NO frontmatter changes.

### 3.2 Body — pre-existing 4-dimension table

`.claude/agents/moai/evaluator-active.md:47-54` already declares the 4-dimension scoring table:

| Dimension | Weight | Criteria | FAIL Condition |
|-----------|--------|----------|----------------|
| Functionality | 40% | All SPEC acceptance criteria met | Any criterion FAIL |
| Security | 25% | OWASP Top 10 compliance | Any Critical/High finding |
| Craft | 20% | Test coverage >= 85%, error handling | Coverage below threshold |
| Consistency | 15% | Codebase pattern adherence | Major pattern violations |

**HRN-003 does NOT introduce these 4 dimensions** — they are already canonical in the agent body. spec.md §1 wording "formalizes" is therefore precise: the SPEC formalizes what is already de-facto, by binding the dimensions to a typed Go enum and YAML profile schema, and by extending each from a flat single score to a per-sub-criterion scoring tree.

### 3.3 Body — pre-existing §11.4.1 cross-reference (HRN-002 M3 landing)

`.claude/agents/moai/evaluator-active.md:91-92`:
```
<!-- @MX:NOTE: Cross-references design-constitution §11.4.1 (SPEC-V3R2-HRN-002) -->
Per design-constitution §11.4.1, evaluator judgment memory is ephemeral per iteration.
The orchestrator MUST respawn evaluator-active via a fresh `Agent()` call at each
GAN-loop iteration boundary; prior iteration's evaluator transcript MUST NOT appear
in the new spawn prompt.
```

This satisfies the fresh-respawn portion of REQ-006 already. HRN-003 only needs to add:
- (a) per-sub-criterion structured JSON output schema,
- (b) explicit rubric-anchor citation requirement,
- (c) cross-reference to `Rubric` schema in `internal/harness/rubric.go`.

The "augment, NOT introduce" framing in spec.md §5.1 REQ-006 v0.2.0 reflects this finding.

### 3.4 Body — output format section (insertion target)

`.claude/agents/moai/evaluator-active.md:57-77` declares the current flat output format:

```
## Evaluation Report
SPEC: {SPEC-ID}
Overall Verdict: PASS | FAIL

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
| Functionality (40%) | {n}/100 | PASS/FAIL/UNVERIFIED | {evidence} |
...
```

HRN-003 augments this with a hierarchical JSON schema sibling. Decision D2 below: HRN-003 adds a NEW subsection "## Hierarchical Score Output (Phase 5)" between line 77 and the "## Evaluator Profile Loading" section (line 79), preserving the existing flat output for backward compatibility.

### 3.5 Body — profile loading section (existing)

`.claude/agents/moai/evaluator-active.md:79-89` already declares profile loading mechanics:

> 1. Check if the SPEC file contains an `evaluator_profile` field in its frontmatter
> 2. If present: load `.moai/config/evaluator-profiles/{evaluator_profile}.md`
> 3. If absent: load `.moai/config/evaluator-profiles/{harness.default_profile}.md` (from harness.yaml)
> 4. If profile file not found: use built-in default weights

This confirms `.md` is the canonical profile format. HRN-003 must consume `.md` rubric tables (Decision D3 below).

---

## 4. As-Is — `.moai/config/evaluator-profiles/` directory

### 4.1 Inventory (drift surfaced)

`ls /Users/goos/MoAI/moai-adk-go/.moai/config/evaluator-profiles/` returns:

```
default.md (2.3KB) — landed 2026-05-10
strict.md  (2.8KB)
lenient.md (2.2KB)
frontend.md (3.8KB)
```

**All 4 profiles already exist as Markdown files** with rubric tables structured exactly as spec.md REQ-003 requires (4 anchor levels at 0.25 / 0.50 / 0.75 / 1.00 per dimension).

`internal/template/templates/.moai/config/evaluator-profiles/` contains identical 4 files at the template tree, byte-identical (2342/3798/2166/2776 bytes) — Template-First discipline already satisfied for HRN-003 read-side consumption.

### 4.2 Drift vs spec.md §5.1 REQ-005

spec.md (authored 2026-04-23) declares: *"4 profiles: `default.yaml`, `strict.yaml`, `lenient.yaml`, `frontend.yaml`; each profile shall include per-dimension rubric templates"*.

Reality on main (2026-05-13): files are `.md`, contain rubric tables, sit alongside the agent body's profile-loading instructions which already point at `.md`.

Reconciliation (per spec.md §10.1 v0.2.0 amendment):
- Adopt `.md` as the canonical format.
- HRN-003 introduces a Go parser that reads the existing `.md` rubric tables (extracts dimension headings, anchor scores from the table, and threshold metadata).
- A `.yaml` parallel schema is OUT of scope (would require migrating the 4 existing profiles + the agent body's load instructions; net new surface area for no new value).

This is locked in spec.md HISTORY row 0.2.0; no further OQ.

### 4.3 default.md anatomy (parser target)

`.moai/config/evaluator-profiles/default.md` structure (verified):

- H1: profile name
- H2: "Evaluation Dimensions" → 3-column table {Dimension | Weight | Pass Threshold}
- H2: "Must-Pass Criteria" → bullet list (Functionality + Security)
- H2: "Hard Thresholds" → bullet list (Security FAIL = Overall FAIL, Coverage <85% = Craft FAIL)
- H2: "Evaluation Rules" → 4 bullet rules
- H2: "Scoring Rubric" → H3 per dimension, each H3 contains a 2-column table {Score | Description} with rows for 1.00 / 0.75 / 0.50 / 0.25.

The Go parser in `internal/harness/rubric.go` (NEW) reads this structure via Markdown table extraction. HRN-003 introduces a tolerant parser that normalizes anchor-score strings (`"1.00"` / `"0.75"` / `"0.50"` / `"0.25"`) to `float64` and rejects deviations.

### 4.4 strict.md vs default.md delta

`strict.md` (2.8KB) differs from `default.md` primarily in:
- Higher pass thresholds (e.g., Functionality 100% vs default 100% — same; Security must include Medium severity, not just Critical/High; Craft coverage ≥90% vs default 85%).
- Stricter must-pass criteria (Security: zero Medium+ findings).

This satisfies spec.md REQ-007 + AC-12 strict-profile semantics. HRN-003 parser must extract `pass_threshold` from the Pass Threshold column of the Dimensions table, NOT from a separate YAML key.

### 4.5 frontend.md UI dimensions (REQ-016)

`frontend.md` (3.8KB) extends the Craft dimension rubric with UI-specific sub-criteria (viewport responsiveness, accessibility, animation smoothness) — already present, satisfying REQ-016 in shape.

---

## 5. As-Is — HRN-001 HarnessConfig integration surface

### 5.1 HarnessConfig struct (current state)

`internal/config/types.go:349-365` (verified):

```go
// HarnessConfig는 harness.yaml 최상위 설정 구조체입니다.
// HRN-002 run-phase minimal substrate: memory_scope 필드 검증만 포함합니다.
// HRN-001 run-phase에서 routing/profile 확장 예정입니다.
type HarnessConfig struct {
    DefaultProfile string         `yaml:"default_profile"`
    Evaluator      EvaluatorConfig `yaml:"evaluator"`
}

// EvaluatorConfig는 evaluator 하위 설정 구조체입니다.
// @MX:NOTE: FROZEN at per_iteration per design-constitution §11.4.1 (SPEC-V3R2-HRN-002)
type EvaluatorConfig struct {
    MemoryScope string `yaml:"memory_scope"`
}
```

The struct is HRN-002's "minimal substrate". HRN-001's full harness routing struct (Levels, ModeDefaults, AutoDetection, Escalation, EffortMapping) is NOT YET LANDED — HRN-001 spec.md status is `draft`, plan-phase artifacts not yet written.

### 5.2 Implications for HRN-003

HRN-003's `EvaluatorConfig` extension lands additively:
- `EvaluatorConfig.Profiles map[string]string` (NEW, REQ-005) — maps profile name to `.md` file path; parser reads each.
- `EvaluatorConfig.Aggregation string` (NEW, REQ-007 + REQ-015) — `"min"` (default) or `"mean"` per profile.
- `EvaluatorConfig.MustPassDimensions []string` (NEW, REQ-008) — defaults to `["Functionality", "Security"]` per design-constitution §12 Mechanism 3 + the existing default.md "Must-Pass Criteria" section.

Decision D7 below: HRN-003 lands the EvaluatorConfig extensions WITHOUT depending on HRN-001's full struct. HRN-001 plan-phase author can later merge the dimension-level fields under their broader `LevelConfig.evaluator_profile` key without conflict.

### 5.3 Loader extension surface

`internal/config/loader.go:223-262` (verified) holds `LoadHarnessConfig()`. HRN-003 extends this loader with a profile-loading step:

```go
// pseudo-code, lands in M3
func (cfg *HarnessConfig) LoadProfiles(profilesDir string) (map[string]*Rubric, error) {
    rubrics := make(map[string]*Rubric)
    for _, name := range []string{"default", "strict", "lenient", "frontend"} {
        path := filepath.Join(profilesDir, name+".md")
        rubric, err := harness.ParseRubricMarkdown(path)
        if err != nil { return nil, err }
        if err := rubric.Validate(); err != nil { return nil, err }  // REQ-012 + REQ-013 + REQ-014
        rubrics[name] = rubric
    }
    return rubrics, nil
}
```

REQ-019 (HRN_UNKNOWN_DIMENSION on 5th-dimension declaration) is enforced inside `Rubric.Validate()`.

### 5.4 Error sentinels (NEW)

`internal/config/errors.go` currently declares `ErrEvalMemoryFrozen` (HRN-002 sentinel, line 44). HRN-003 adds:

- `ErrUnknownDimension` — REQ-012, REQ-019 (`HRN_UNKNOWN_DIMENSION`)
- `ErrRubricCitationMissing` — REQ-009 (`HRN_RUBRIC_CITATION_MISSING`)
- `ErrFlatScoreCardProhibited` — REQ-017 (`HRN_FLAT_SCORECARD_PROHIBITED`)
- `ErrMustPassBypassProhibited` — REQ-008, REQ-018 (`HRN_MUSTPASS_BYPASS_PROHIBITED`)

All 4 follow the existing `ErrEvalMemoryFrozen` naming/message pattern.

---

## 6. As-Is — HRN-002 Sprint Contract integration surface

### 6.1 Sprint Contract storage (post-HRN-002)

`.moai/config/sections/design.yaml:13-29` declares `sprint_contract.artifact_dir: ".moai/sprints"`. HRN-002 §11.4.1 added the FROZEN rule that this directory carries the durable cross-iteration state.

`.moai/sprints/` directory currently empty (no in-flight GAN runs); architecture supports both `{team-id}/contract.yaml` and `{spec-id}/contract.yaml` aliases per HRN-002 REQ-016.

### 6.2 Sprint Contract YAML shape (consumed by HRN-003)

`.claude/skills/moai-workflow-gan-loop/SKILL.md:210-240` (verified) declares the JSON shape: `sprint_id`, `iteration`, `priority_dimension`, `acceptance_checklist[]`, `test_scenarios[]`, `pass_conditions{}`, `negotiation_history[]`, `created_at`. Status enum: `pending | passed | failed`.

HRN-003 REQ-011 extends `acceptance_checklist[]` items with a per-sub-criterion `status` field (enum: `passed | failed | refined | new`). This is an additive shape change — backward-compatible for v2 contracts that lack sub-criterion granularity.

### 6.3 Persistence path (Decision D6)

The scorer (Go side, `internal/harness/scorer.go`) emits a `WriteContract()` helper that takes the in-memory `*ScoreCard` + Sprint Contract path and writes the updated YAML. The orchestrator-level GAN loop runner (SKILL.md text) calls this helper between iterations. No `gan_loop.go` orchestrator-side runner module is created — the loop is the orchestrator's interpretive responsibility, the persistence helper is Go-side.

### 6.4 Fresh-respawn contract (REQ-013 cross-reference to HRN-002)

HRN-003 REQ-013 ("Judgment memory volatility") restates HRN-002 REQ-013 verbatim. The enforcement is HRN-002's leak-detection test (`internal/harness/evaluator_leak.go` + `evaluator_leak_test.go`). HRN-003 adds NO new enforcement for REQ-013; it is satisfied transitively via the HRN-002 substrate.

---

## 7. As-Is — SPC-001 hierarchical AC parser API surface

### 7.1 Acceptance struct (consumed by HRN-003 scorer)

`internal/spec/ears.go:21` declares `MaxDepth = 3`. The `Acceptance` struct (per SPC-001 REQ-005) has fields:

```go
type Acceptance struct {
    ID             string
    Given          string
    When           string
    Then           string
    RequirementIDs []string
    Children       []Acceptance
}
```

`internal/spec/ears.go:24-26` declares `topLevelIDPattern = ^AC-[A-Z0-9]+-[0-9]+-[0-9]+$`.

### 7.2 Parser API (consumed by HRN-003)

`internal/spec/parser.go` (per SPC-001 acceptance.md AC-15/16/17) exposes:
- `extractACLines()` (lines 90-117) → indent-tracked AC lines with depth.
- `buildTree()` → `[]Acceptance` tree with `Children` populated.
- `Acceptance.ValidateID()` (lines 28-33) → topLevelIDPattern enforcement.
- `internal/spec/parser.go:200-227` → flat-AC auto-wrap path (per REQ-SPC-001-010 / REQ-SPC-001-020).
- `internal/spec/lint.go:394-403` → `collectAllREQIDs(criteria)` traverses tree and returns flat REQ-ID set.

HRN-003 scorer iterates `[]Acceptance` recursively. The leaf detection rule is `len(node.Children) == 0`. For each leaf:
- The leaf's `RequirementIDs` map back to spec.md §5 REQs.
- The scorer invokes evaluator-active per leaf (per REQ-004 + REQ-009).
- The leaf's score becomes a `SubCriterionScore`.
- Parent aggregation (min default, mean opt-in per REQ-007) combines children's scores.

### 7.3 Auto-wrap interaction with REQ-010

When SPEC's acceptance.md is flat (no children), SPC-001's parser auto-wraps each top-level AC as a synthesized single-child `.a` (parser.go:200-227). HRN-003 scorer therefore always traverses a hierarchical tree — flat input becomes single-child synthesized hierarchy. REQ-010 ("flat criteria auto-wrap as single-level children") is satisfied by SPC-001 infrastructure; HRN-003 just consumes the post-wrap shape.

### 7.4 BC-V3R2-011 backward compat

HRN-003 imposes NO new constraints on legacy v2 SPECs. Auto-wrapped flat ACs produce a single `SubCriterionScore` per `CriterionScore`; aggregation collapses trivially (min/mean of a single value = the value). This satisfies REQ-010 + spec.md §7 [HARD] backward-compat guarantee.

---

## 8. As-Is — Pattern library E-1 + E-3 verbatim excerpts

### 8.1 E-1 Agent-as-a-Judge Intermediate Trajectory Scoring

Source: `.moai/design/v3-redesign/synthesis/pattern-library.md:279-285` (verified):

> ### E-1 Agent-as-a-Judge Intermediate Trajectory Scoring
>
> - **Source**: R1 §9 Agent-as-a-Judge (Zhuge et al. 2024, *arXiv:2410.10934*).
> - **Description**: Extend LLM-as-a-Judge with agentic capabilities (tool use, memory, multi-step reasoning). Scores intermediate trajectories, not only final outputs. Matches human evaluation reliability at 97% cost savings.
> - **When to apply**: evaluator-active at thorough harness — scores trajectory, not just final artifact. Hierarchical requirements (Zhuge's 365 sub-requirements per 55 tasks) map to moai's acceptance criteria per SPEC.
> - **Trade-offs**: Judge memory across iterations cascades errors (R1 §9 anti-pattern flag) — must scope memory per-iteration.
> - **v3 disposition**: **ADOPT (priority 9)** — evaluator-active already in this shape; v3 formalizes hierarchical acceptance criteria and enforces per-iteration memory scope.

HRN-003 implements the per-sub-criterion scoring half of E-1 (HRN-002 implemented the per-iteration memory scope half). Together HRN-002 + HRN-003 close the E-1 pattern.

### 8.2 E-3 Rubric-Anchored Scoring with Independent Re-evaluation

Source: `.moai/design/v3-redesign/synthesis/pattern-library.md:295-301` (verified):

> ### E-3 Rubric-Anchored Scoring with Independent Re-evaluation
>
> - **Source**: design-constitution.md §12 (Mechanism 1 & 4), aligns with R1 Reflexion + Constitutional AI critique-then-revise.
> - **Description**: Every evaluation criterion has concrete rubric with examples at 0.25, 0.50, 0.75, 1.0. Every 5th project undergoes independent re-evaluation: scores must be within 0.10 of each other; divergence triggers calibration review.
> - **When to apply**: evaluator-active and plan-auditor both reference rubrics in `.moai/config/evaluator-profiles/`. Independent re-evaluation automatic every 5th SPEC.
> - **Trade-offs**: Rubric authoring is expensive one-time cost.
> - **v3 disposition**: **ADOPT** — already FROZEN in design subsystem; v3 populates evaluator-profiles per harness level (default, strict, lenient, frontend).

HRN-003 implements the rubric-anchor citation enforcement portion (REQ-009, REQ-013). The Independent Re-evaluation portion (Mechanism 4) is FROZEN per design-constitution §12 and is OUT of HRN-003 scope (separate observability work).

### 8.3 Top-10 priority confirmation

`pattern-library.md:395-397` ranks E-1 as #9 in the top-10 v3 priority patterns, citing the per-iteration memory scope close (HRN-002) AND hierarchical acceptance criteria enhancement (HRN-003) as paired deliverables. HRN-003 P1 priority in spec.md frontmatter aligns.

---

## 9. As-Is — Design constitution §11 + §12 anchors

### 9.1 §12 Mechanism 1 (Rubric Anchoring) — FROZEN, REQ-009 + REQ-013 source

`.claude/rules/moai/design/constitution.md:357-359` (verified):

> ### Mechanism 1: Rubric Anchoring
>
> Every evaluation criterion has a concrete rubric with examples of scores at 0.25, 0.50, 0.75, and 1.0. evaluator-active MUST reference the rubric when assigning scores. Scores without rubric justification are invalid.

Locked-in FROZEN clause is the evidence base for HRN-003 REQ-003 + REQ-009 + REQ-013. Anchor levels {0.25, 0.50, 0.75, 1.00} cannot be evolved.

### 9.2 §12 Mechanism 3 (Must-Pass Firewall) — FROZEN, REQ-008 + REQ-018 source

`.claude/rules/moai/design/constitution.md:365-367` (verified):

> ### Mechanism 3: Must-Pass Firewall
>
> Must-pass criteria cannot be compensated by high scores in other areas. A project with perfect nice-to-have scores but a failing must-pass criterion still fails. This is FROZEN and cannot be evolved.

Locked-in FROZEN clause is the evidence base for HRN-003 REQ-008 + REQ-018. The default must-pass set is `[Functionality, Security]` per `.moai/config/evaluator-profiles/default.md` "Must-Pass Criteria" section. OQ3 below: should this set be configurable per profile?

### 9.3 §11.4.1 (HRN-002 Evaluator Memory Scope) — FROZEN, REQ-013 cross-ref

`.claude/rules/moai/design/constitution.md:341-349` (verified) — added by HRN-002 commit `0ac27ee4e`. HRN-003 REQ-013 ("Judgment memory volatility") cross-references this clause. No re-amendment needed; HRN-003 inherits.

### 9.4 §5 Layer 5 (Pass threshold floor 0.60) — FROZEN, REQ-014 source

`.claude/rules/moai/design/constitution.md:42-44` (FROZEN zone declarations) lists "Pass threshold floor (minimum 0.60, cannot be lowered by evolution)". HRN-003 REQ-014 inherits and enforces at the loader (every profile's `pass_threshold` ≥ 0.60).

---

## 10. As-Is — Zone registry mirror context

### 10.1 CONST-V3R2-153 (HRN-002 §11.4.1 mirror)

`.claude/rules/moai/core/zone-registry.md:570` declares CONST-V3R2-153 as the HRN-002 §11.4.1 FROZEN mirror entry:

```yaml
- id: CONST-V3R2-153
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#1141-evaluator-memory-scope-principle-4"
  clause: "§11.4.1 Evaluator Memory Scope (Principle 4) ..."
  canary_gate: true
```

### 10.2 HRN-003 candidate mirror entries (next IDs CONST-V3R2-154+)

HRN-003 introduces multiple FROZEN constraints that COULD warrant zone-registry mirror entries:

1. 4-dimension enum `{Functionality, Security, Craft, Consistency}` (REQ-001, §12 Mechanism 1 implicitly).
2. 4 rubric anchors `{0.25, 0.50, 0.75, 1.00}` (REQ-003, §12 Mechanism 1 explicitly — already mirrored?).
3. Must-pass firewall (REQ-008, §12 Mechanism 3 explicitly — already mirrored?).
4. Pass threshold floor 0.60 (REQ-014, §5 — already mirrored as CONST-V3R2-059).

Verified in zone-registry.md: CONST-V3R2-059 "Pass threshold floor (minimum 0.60, cannot be lowered by evolution)" already exists. The 4-dimension enum, 4 rubric anchors, and must-pass firewall do NOT have explicit mirror entries (they are clauses inside §11/§12 that the registry treats holistically as CONST-V3R2-056/057, not per-mechanism).

OQ1 below: should HRN-003 M5 register 1-3 additional CONST-V3R2-NNN entries (e.g., 154 = 4-dimension enum, 155 = 4 anchor levels), OR defer to a follow-up CON-002 amendment SPEC that consolidates Mechanism 1/3 sub-clauses?

---

## 11. Gap Analysis — HRN-003 vs Already Landed

| Capability | spec.md REQ | Implementation status | Gap |
|------------|-------------|------------------------|-----|
| `Dimension` enum (4-value) | REQ-001 | NOT LANDED | M2 task — `internal/harness/scorer.go` |
| Hierarchical `ScoreCard` struct | REQ-002 | NOT LANDED | M2 task — `internal/harness/scorer.go` |
| `Rubric` struct + 4 anchors | REQ-003 | NOT LANDED — but `.md` profile tables already encode them | M2 task — `internal/harness/rubric.go` |
| `EvaluatorRunner.Score()` | REQ-004 | NOT LANDED | M3 task — `internal/harness/scorer.go` |
| 4 evaluator profile files | REQ-005 | ALREADY LANDED — `.md` format on main as of 2026-05-10 | — (read-side only; M3 parser consumes) |
| evaluator-active body 4-dim contract | REQ-006 | PARTIALLY LANDED — body has 4-dim table (lines 47-54) + §11.4.1 cross-ref (lines 91-92) | M4 task — augment with hierarchical JSON schema + rubric-citation requirement |
| Aggregation rules (min/mean) | REQ-007, REQ-015 | NOT LANDED | M3 task — `internal/harness/scorer.go` |
| Must-pass firewall | REQ-008, REQ-018 | DOC-LANDED in §12 Mechanism 3; NOT CODE-LANDED | M3 task — `internal/harness/scorer.go` |
| Rubric citation enforcement | REQ-009 | NOT LANDED | M3 task — `internal/harness/rubric.go` validator |
| Flat AC auto-wrap (BC-V3R2-011) | REQ-010 | ALREADY LANDED — SPC-001 parser auto-wraps | — (transitive) |
| Sprint Contract sub-criterion persistence | REQ-011 | NOT LANDED — Sprint Contract YAML shape exists but lacks sub-criterion granularity | M3 task — Go-side `WriteContract()` helper + SKILL.md text update |
| 4-dim FROZEN enforcement | REQ-012, REQ-019 | NOT LANDED | M2 task — Rubric.Validate() + loader rejects 5th dimension |
| Anchor levels FROZEN | REQ-013 | NOT LANDED | M2 task — Rubric.Validate() rejects non-{0.25, 0.50, 0.75, 1.00} anchors |
| Pass threshold floor 0.60 | REQ-014 | DOC-LANDED in §5; NOT CODE-LANDED for HRN-003 profiles | M3 task — Rubric.Validate() rejects pass_threshold < 0.60 |
| Frontend profile UI dims | REQ-016 | ALREADY LANDED — frontend.md has UI sub-criteria | — (transitive) |
| Flat ScoreCard prohibition | REQ-017 | NOT LANDED | M3 task — CI integration test |
| Must-pass bypass prohibition | REQ-018 | NOT LANDED | M3 task — Rubric.Validate() |

Summary: HRN-003 is ~80% greenfield. Pre-existing substrate:
- 4 evaluator profile `.md` files exist (REQ-005 transitively satisfied for read-side).
- evaluator-active body has 4-dim table + §11.4.1 cross-ref (REQ-006 partially satisfied).
- SPC-001 parser auto-wraps flat ACs (REQ-010 transitively satisfied).
- frontend.md has UI sub-criteria (REQ-016 transitively satisfied).
- HRN-002 leak detection enforces fresh judgment (REQ-013 transitively satisfied).
- Pass threshold floor 0.60 + must-pass firewall + rubric anchoring are FROZEN in design-constitution (REQ-008 + REQ-013 + REQ-014 doc-landed).

Plan therefore focuses on:
1. **M1** — plan-phase artifacts (this PR).
2. **M2** — Type definitions: `Dimension` enum, `ScoreCard`, `Rubric` structs in `internal/harness/scorer.go` + `rubric.go`. Validators (REQ-012, REQ-013, REQ-014) inline.
3. **M3** — Aggregation logic + must-pass firewall + rubric citation enforcement + Sprint Contract sub-criterion persistence (REQ-007, REQ-008, REQ-009, REQ-011, REQ-015, REQ-017, REQ-018).
4. **M4** — Profile loader extension (consume `.md` files via Markdown table parser) + `EvaluatorConfig` field additions in `internal/config/types.go` + evaluator-active body augment + SKILL.md update + zone-registry CONST entries (REQ-005, REQ-006, REQ-019).
5. **M5** — Tests + MX tags + completion gate (acceptance fixtures, ≥85% coverage, integration test for REQ-017).

Note: HRN-003 does NOT require CON-002 paperwork unless OQ1 resolves to "register new CONST entries". HRN-003 modifies NO FROZEN-zone clauses (§11.4.1 already amended by HRN-002; HRN-003 only reads §12 Mechanism 1/3 + §5 floor). Therefore M5 is lighter than HRN-002 M5.

---

## 12. Risks Identified During Research

| # | Risk | Severity | Mitigation milestone |
|---|------|----------|-----------------------|
| R1 | Markdown rubric table parser brittle to author formatting variations (extra columns, alignment whitespace, missing rows) | MEDIUM | M2 — tolerant parser with explicit error messages; M5 — fixture covering all 4 existing profiles + 3 malformed variants |
| R2 | evaluator-active inconsistent structured JSON output across iterations | HIGH | M4 — schema declared in agent body; M3 — REQ-009 strict rubric-anchor enforcement; retry-on-reject pattern up to 2 retries per sub-criterion |
| R3 | Rubric authoring burden per SPEC grows | MEDIUM | Default profile templates cover 80%+ of SPECs; rubric authoring is opt-in via profile selection |
| R4 | Min-aggregation too strict for exploratory SPECs | MEDIUM | Profile `aggregation: mean` opt-in per REQ-015; lenient profile uses mean by default |
| R5 | Sub-criterion count grows unbounded on complex SPECs | MEDIUM | SPC-001 MaxDepth=3 caps depth; HRN-003 inherits |
| R6 | Must-pass firewall surprises user when overall score is high | MEDIUM | M3 — output the firewall trigger message explicitly in ScoreCard verdict; cite the failing must-pass dimension in Evidence field |
| R7 | Flat-criteria auto-wrap loses information | LOW | SPC-001 parser preserves Given/When/Then verbatim on auto-wrap; HRN-003 inherits |
| R8 | Profile drift between template and local `.moai/config/evaluator-profiles/` | LOW | Already addressed by Template-First — current 4 files are byte-identical between template and local |
| R9 | Aggregation rule confusion (min vs mean) | LOW | Default min documented; profile explicitly opt-in mean; log the effective rule in ScoreCard rationale |
| R10 | Evaluator profile schema drift between `.md` rubric tables and Go struct | MEDIUM | M2 — Rubric.Validate() runs on every loader invocation; M5 — CI test parses all 4 shipping profiles |
| R11 | HRN-001 lands later with conflicting `EvaluatorConfig` field name (e.g., `Profile string` vs HRN-003's `Profiles map[string]string`) | MEDIUM | M4 — coordinate with HRN-001 plan-phase author; favor Decision D7 additive merge |
| R12 | Sprint Contract sub-criterion shape change breaks HRN-002 leak detection test | LOW | REQ-011 is additive (new `status` field per item); HRN-002 leak detection scans for `Score:`/`Feedback:`/`Verdict:` substrings, not contract YAML structure — no interaction |
| R13 | OQ1 (zone-registry mirror) deferral leaves HRN-003 FROZEN constraints undocumented in registry | MEDIUM | Default proposal (M5 task): register 2 entries (CONST-V3R2-154 4-dim enum, CONST-V3R2-155 4 anchor levels); fallback (defer to follow-up SPEC): document deferral in M5 evidence + open follow-up issue |

---

## 13. Cross-Reference Summary (file:line anchors, ≥25)

1. `internal/harness/types.go:1-10` — Package header; existing types listed.
2. `internal/harness/types.go:65-95` — `Tier`, `Pattern` types (Phase 2 learning); HRN-003 does NOT extend.
3. `internal/harness/types.go:152-208` — `Proposal`, `Decision`, `Session` (Phase 3 safety); HRN-003 does NOT extend.
4. `internal/harness/evaluator_leak.go:15` — `ErrPriorJudgmentLeak` sentinel (HRN-002 substrate).
5. `internal/harness/evaluator_leak.go:40-61` — `DetectPriorJudgmentLeak()` validator — HRN-003 inherits, does NOT modify.
6. `internal/harness/evaluator_leak_test.go` — leak detection regression test (HRN-002 substrate).
7. `internal/harness/applier.go:167-176` — `Apply()` with safety evaluator; CON-002 surface, unrelated naming.
8. `internal/harness/scorer.go` — DOES NOT EXIST; HRN-003 M2 creates.
9. `internal/harness/rubric.go` — DOES NOT EXIST; HRN-003 M2 creates.
10. `internal/harness/scorer_test.go` — DOES NOT EXIST; HRN-003 M5 creates.
11. `internal/harness/rubric_test.go` — DOES NOT EXIST; HRN-003 M5 creates.
12. `internal/harness/gan_loop.go` — DOES NOT EXIST; HRN-003 does NOT create (Decision D6 inherits HRN-002 D1).
13. `internal/config/types.go:349-365` — `HarnessConfig`, `EvaluatorConfig` (HRN-002 minimal substrate).
14. `internal/config/types.go:354` — `Evaluator EvaluatorConfig` field (HRN-003 extension target).
15. `internal/config/loader.go:223-262` — `LoadHarnessConfig()` (HRN-003 M4 extension target).
16. `internal/config/errors.go:44` — `ErrEvalMemoryFrozen` sentinel (HRN-002 pattern; HRN-003 follows for 4 new sentinels).
17. `internal/config/errors.go:48-66` — `ValidationError` struct (HRN-003 reuses for new error wrapping).
18. `internal/spec/ears.go:21` — `MaxDepth = 3` constant (SPC-001).
19. `internal/spec/ears.go:24-26` — `topLevelIDPattern` (SPC-001).
20. `internal/spec/parser.go:90-117` — `extractACLines()` (SPC-001 indent-tracker — HRN-003 consumes).
21. `internal/spec/parser.go:200-227` — flat-AC auto-wrap path (SPC-001; satisfies HRN-003 REQ-010 transitively).
22. `internal/spec/lint.go:394-403` — `collectAllREQIDs()` (SPC-001 tree traversal — HRN-003 reuse pattern).
23. `.claude/agents/moai/evaluator-active.md:1-26` — Frontmatter (HRN-003 preserves verbatim).
24. `.claude/agents/moai/evaluator-active.md:47-54` — 4-dimension table (HRN-003 inherits, does NOT introduce).
25. `.claude/agents/moai/evaluator-active.md:57-77` — Current flat output format (HRN-003 M4 augments with hierarchical JSON schema sibling section).
26. `.claude/agents/moai/evaluator-active.md:79-89` — Profile loading section (HRN-003 inherits; confirms `.md` is canonical).
27. `.claude/agents/moai/evaluator-active.md:91-92` — §11.4.1 cross-reference (HRN-002 M3 landing; HRN-003 inherits).
28. `.claude/agents/moai/evaluator-active.md:94-105` — Sprint Contract Negotiation + Intervention Modes (HRN-003 reuses Phase 2.0 thorough hook).
29. `.moai/config/evaluator-profiles/default.md:1-67` — Default profile rubric tables (HRN-003 parser target).
30. `.moai/config/evaluator-profiles/default.md:14-17` — Must-Pass Criteria (HRN-003 REQ-008 reads `[Functionality, Security]`).
31. `.moai/config/evaluator-profiles/default.md:33-67` — 4 dimension rubric tables × 4 anchor scores each (HRN-003 REQ-003 schema source).
32. `.moai/config/evaluator-profiles/strict.md` — Strict profile (HRN-003 REQ-007 + AC-12 source).
33. `.moai/config/evaluator-profiles/lenient.md` — Lenient profile.
34. `.moai/config/evaluator-profiles/frontend.md` — Frontend profile (HRN-003 REQ-016 transitively satisfied).
35. `internal/template/templates/.moai/config/evaluator-profiles/{default,strict,lenient,frontend}.md` — Template mirror; byte-identical.
36. `.claude/rules/moai/design/constitution.md:42-44` — FROZEN pass threshold floor 0.60 (HRN-003 REQ-014 source).
37. `.claude/rules/moai/design/constitution.md:341-349` — §11.4.1 (HRN-002 amendment; HRN-003 REQ-013 cross-ref).
38. `.claude/rules/moai/design/constitution.md:357-359` — §12 Mechanism 1 Rubric Anchoring (HRN-003 REQ-009 + REQ-013 source).
39. `.claude/rules/moai/design/constitution.md:365-367` — §12 Mechanism 3 Must-Pass Firewall (HRN-003 REQ-008 source).
40. `.claude/rules/moai/core/zone-registry.md:570` — CONST-V3R2-153 (HRN-002 mirror; HRN-003 follows pattern for OQ1).
41. `.moai/design/v3-redesign/synthesis/pattern-library.md:279-285` — E-1 verbatim text.
42. `.moai/design/v3-redesign/synthesis/pattern-library.md:295-301` — E-3 verbatim text.
43. `.moai/design/v3-redesign/synthesis/pattern-library.md:395-397` — Top-10 priority #9 E-1 confirmation.
44. `.claude/skills/moai-workflow-gan-loop/SKILL.md:210-240` — Sprint Contract YAML shape (HRN-003 REQ-011 extension target).
45. `.claude/skills/moai-workflow-gan-loop/SKILL.md:147-158` — Stagnation detection (HRN-003 inherits HRN-002 D1).
46. `.moai/specs/SPEC-V3R2-HRN-001/spec.md:286-287` — HRN-001 cites HRN-003 as downstream consumer (per-level evaluator profile).
47. `.moai/specs/SPEC-V3R2-HRN-002/spec.md:73-83` — §11.4.1 verbatim text (HRN-003 REQ-013 source).
48. `.moai/specs/SPEC-V3R2-SPC-001/spec.md:130-146` — AC-SPC-001-01 through AC-SPC-001-17 (HRN-003 acceptance.md self-demonstration follows pattern).
49. `.moai/specs/SPEC-V3R2-HRN-003/spec.md` — This SPEC, v0.2.0 frontmatter + 19 REQs + 12 ACs + Drift Reconciliation §10.1.
50. `.moai/sprints/` — empty directory; storage target for HRN-003 REQ-011 sub-criterion persistence.

Total file:line anchors: 50 (>25 minimum required).

---

## 14. Open Questions (OQ1..OQ5)

### OQ1 — Zone-registry mirror entries: M5 task or follow-up SPEC?

**Question**: HRN-003 introduces several FROZEN constraints (4-dimension enum, 4 rubric anchors, must-pass dimension defaults). Should M5 register 1-3 new CONST-V3R2-NNN entries in `.claude/rules/moai/core/zone-registry.md`, or defer registry extension to a follow-up CON-002 amendment SPEC?

**Proposed default**: Register 2 entries in M5:
- **CONST-V3R2-154** — `clause: "4-dimension scoring enum {Functionality, Security, Craft, Consistency} FROZEN per SPEC-V3R2-HRN-003 REQ-001"`, `zone: Frozen`, `canary_gate: true`, `file: internal/harness/scorer.go`, `anchor: "#dimension-enum"`.
- **CONST-V3R2-155** — `clause: "4 rubric anchor levels {0.25, 0.50, 0.75, 1.00} FROZEN per SPEC-V3R2-HRN-003 REQ-003 + design-constitution §12 Mechanism 1"`, `zone: Frozen`, `canary_gate: true`, `file: internal/harness/rubric.go`, `anchor: "#anchor-levels"`.

The must-pass firewall (REQ-008) already has design-constitution §12 Mechanism 3 as an authoritative source; CONST-V3R2-057 covers it implicitly. Pass threshold floor 0.60 (REQ-014) already mirrors as CONST-V3R2-059. No additional entries needed.

**Rationale**: The 4-dimension enum + 4 anchor levels are NEW v3 invariants introduced by code (not just by constitution prose), so they merit explicit machine-readable registry entries that future amendment proposals must check. CON-002 is NOT triggered because HRN-003 modifies NO existing FROZEN clauses — the new entries are additive registrations, not amendments to existing clauses. Cost: ~30 LOC in zone-registry.md, minimal risk.

**Trade-off if deferred**: Future contributors editing `internal/harness/scorer.go` Dimension enum will not see the FROZEN constraint at the registry level; they will only see the `@MX:WARN` tag in the source. Acceptable risk but suboptimal.

### OQ2 — Aggregation default `min` vs `mean`

**Question**: spec.md REQ-007 declares default `min(sub_scores)` with profile-level opt-in `mean`. Confirm this is the right default?

**Proposed default**: CONFIRM `min` as default, per spec.md REQ-007. Rationale: any sub-criterion failure should fail the criterion (consistent with §12 Mechanism 3 Must-Pass Firewall philosophy — partial pass is dangerous). Lenient profile opts into `mean` per REQ-015 to allow exploratory SPECs to weigh partial successes.

**Trade-off**: `min` is harsh on novel SPECs where 1 of 5 sub-criteria fails for irrelevant reasons. Mitigated by `lenient.md` profile + per-dimension `aggregation: mean` opt-in.

### OQ3 — Must-pass dimensions default set: hard-coded or configurable?

**Question**: spec.md REQ-008 + REQ-018 enforce must-pass firewall but don't fully specify which dimensions are must-pass by default. design-constitution §12 Mechanism 3 declares it FROZEN but the specific dimension set is profile-defined.

**Proposed default**: Default must-pass set is `[Functionality, Security]`, sourced from the existing `default.md` profile's "Must-Pass Criteria" section. Profiles MAY widen this (e.g., `strict.md` adds Craft when coverage <90%) but MAY NOT narrow it below the design-constitution-implied minimum of `[Security]`. The narrowing constraint is REQ-018 / `HRN_MUSTPASS_BYPASS_PROHIBITED`.

**Trade-off**: Hard-coding `[Functionality, Security]` as the floor in `Rubric.Validate()` couples Go code to profile authoring conventions. Mitigated by making the floor a `var DefaultMustPassDimensions = []Dimension{Functionality, Security}` exported constant; profile parser cross-checks against it.

### OQ4 — Structured JSON schema versioning strategy

**Question**: REQ-002 defines `ScoreCard` Go struct; REQ-006 declares evaluator-active emits structured JSON. Should the JSON schema have an explicit version field for forward-compatibility?

**Proposed default**: YES. Add `schema_version: "v1"` field to the top-level ScoreCard JSON output. Mirrors `LogSchemaVersion = "v1"` pattern from `internal/harness/types.go:11`. When v3.1 introduces a 5th dimension (requires CON-002 amendment), the schema version bumps to `"v2"` and Go-side parser dispatches accordingly. Cost: 1 line in struct + 1 line in serializer.

**Trade-off**: Versioning unused fields is YAGNI for v3.0. Counter: HRN-002 introduced `LogSchemaVersion` for the same reason; consistency wins.

### OQ5 — Rubric-anchor citation enforcement: strict reject vs warn

**Question**: REQ-009 declares the scorer rejects scores without rubric_anchor citation with `HRN_RUBRIC_CITATION_MISSING` error. Should the rejection be strict (fail iteration) or lenient (warn + continue using a default anchor)?

**Proposed default**: STRICT reject. The evaluator-active body (post-M4 augment) makes citation MANDATORY. A missing citation is an evaluator bug, not a user error. The retry-on-reject pattern (up to 2 retries per sub-criterion) provides resilience. This satisfies REQ-009 verbatim and matches §12 Mechanism 1 ("Scores without rubric justification are invalid").

**Trade-off**: Strict mode makes evaluator-active behavior brittle if Claude Opus 4.7 occasionally emits malformed JSON. Mitigated by retry pattern + integration test fixtures covering 5+ malformed-emit scenarios.

---

## 15. Decisions Driving Plan.md

D1. **`.md` profile format adopted; `.yaml` parallel schema OUT of scope.** Confirmed in spec.md HISTORY 0.2.0. The Go-side rubric parser (`internal/harness/rubric.go`) reads existing `.md` rubric tables. No migration of the 4 existing profiles.

D2. **evaluator-active body is augmented (not introduced).** The 4-dimension table + §11.4.1 cross-reference already exist. M4 task adds: hierarchical JSON output schema subsection, rubric-citation requirement subsection, and `Rubric` schema cross-reference. No frontmatter change. No structural rewrite.

D3. **Evaluator profile parser consumes `.md`, not `.yaml`.** Markdown table extraction in `internal/harness/rubric.go`. Tolerant to whitespace; strict on anchor scores (must be exactly {0.25, 0.50, 0.75, 1.00}).

D4. **Hierarchical AC scoring builds on SPC-001 parser.** No new SPC-001 changes required; HRN-003 consumes `[]Acceptance` tree from `internal/spec/parser.go` and traverses recursively. Flat ACs handled transitively via SPC-001's auto-wrap.

D5. **Sprint Contract sub-criterion persistence is additive YAML extension.** REQ-011 adds a `status` field per `acceptance_checklist[]` item (enum: `passed | failed | refined | new`). Backward-compatible for v2 contracts. Persistence helper lives in `internal/harness/scorer.go` (Go-side `WriteContract()`); orchestrator-level GAN loop calls it.

D6. **No `internal/harness/gan_loop.go` runner module.** HRN-002 D1 declared the orchestrator-level runner (`.claude/skills/moai-workflow-gan-loop/SKILL.md`) is the actual integration point. HRN-003 inherits this. The Go-side scorer is a library called by the orchestrator, not a standalone loop.

D7. **EvaluatorConfig field additions are additive vs HRN-001.** HRN-001 plan-phase author can later introduce `LevelConfig.evaluator_profile` fields without conflict. Convention: HRN-003 fields live inside `EvaluatorConfig` (`Profiles`, `Aggregation`, `MustPassDimensions`); HRN-001 fields live inside `LevelConfig` and reference `EvaluatorConfig.Profiles` by name.

D8. **CON-002 paperwork NOT required.** HRN-003 modifies NO FROZEN-zone clauses. §11.4.1 already amended by HRN-002; §12 Mechanism 1/3 + §5 floor are READ ONLY. M5 is therefore lighter than HRN-002 M5 (no Canary, no FrozenGuard evidence, no HumanOversight gate). Per OQ1 default: M5 registers 2 additive zone-registry entries (CONST-V3R2-154, CONST-V3R2-155) which is documentation, not amendment.

D9. **Plan-in-main (no SPEC worktree).** Per CLAUDE.local.md §18.12 BODP and PR #822 doctrine, plan-phase artifacts work on a feature branch from main. Plan PR is `plan/SPEC-V3R2-HRN-003` cut from `main` HEAD `0ac27ee4e` (HRN-002 M5 merge). Run-phase work creates a fresh worktree at run start per spec-workflow.md Step 2.

D10. **Acceptance.md self-demonstrates hierarchical schema on ≥4 of 12 ACs.** Per task brief HARD constraint and SPC-001 dogfooding precedent. AC-HRN-003-03, AC-HRN-003-04, AC-HRN-003-05, AC-HRN-003-07 will use depth-2 `.a/.b/.c` children. AC-HRN-003-04 includes depth-3 grandchildren `.a.i/.a.ii` to exercise MaxDepth=3.

---

End of research.
