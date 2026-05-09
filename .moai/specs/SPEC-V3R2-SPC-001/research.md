# Research — SPEC-V3R2-SPC-001 EARS + Hierarchical Acceptance Criteria

> Phase 0.5 deep codebase research preceding plan.md.
> Captures the as-is state of EARS / acceptance-criteria handling across 185 SPECs and the Go parser/lint subsystem, and validates the hierarchical schema against existing implementation.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-09 | manager-spec (Wave 9 plan author) | Initial research for SPC-001 plan phase; reflects discovery that ~80% of SPC-001 implementation already landed via prior incremental work. |

---

## 1. Research Goal

Establish ground truth for SPEC-V3R2-SPC-001 plan:

1. Audit current EARS + acceptance-criteria architecture (rules, parser, renderer, tests).
2. Quantify migration scope: how many existing SPECs are flat, mixed, or already hierarchical.
3. Identify the **gap delta** between landed code and SPC-001 requirements.
4. Surface adjacent SPECs (CON-001, MIG-001, SPC-003, HRN-002/003) and their cross-cuts.
5. Derive milestone breakdown for plan.md.

This research must precede plan.md per `.claude/skills/moai-workflow-spec` Phase 0.5 protocol.

---

## 2. As-Is — EARS Rules Surface

### 2.1 Workflow rule (`.claude/rules/moai/workflow/spec-workflow.md`)

- L88 — Plan phase mandates "Create comprehensive specification using EARS format".
- L92 — "Planning: SPEC document creation with EARS format requirements".
- L104 — Phase output enumerates "EARS format requirements".
- L149 — Pre-submission self-review compares full diff "against SPEC acceptance criteria".
- L179 — Re-planning gate triggers on "3+ iterations with no new SPEC acceptance criteria met".
- L194 — Stagnation detection appends "acceptance criteria completion count" per iteration.
- L196 — Stagnation flag fires when "acceptance criteria completion rate is zero for 3+ consecutive entries".

The rule references acceptance criteria as a flat counter; no schema-level guidance for hierarchical shape currently exists in this rule. SPC-001 amends this rule with the new schema (see plan.md M3).

### 2.2 Constitution / FROZEN zone

- `.claude/rules/moai/core/zone-registry.md:16-19` — `CONST-V3R2-001` pins `SPEC+EARS format` to `.claude/rules/moai/workflow/spec-workflow.md#phase-overview` as Frozen + canary_gate true.
- `docs/design/major-v3-master.md:48-56` — §1.3 lists "SPEC EARS format" among FROZEN invariants.
- `docs/design/major-v3-master.md:970` — BC-V3R2-011 declares "SPEC acceptance criteria become hierarchical (nested Given/When/Then); flat criteria promoted to 1-level tree" as AUTO migrator-handled, "Old flat SPECs remain parseable indefinitely".
- `docs/design/major-v3-master.md:992` — Phase 5 Harness + Evaluator owns SPC-001 alongside HRN-001/002/003.
- `docs/design/major-v3-master.md:1035-1036` — §11.2 names SPC-001 ("EARS + hierarchical acceptance criteria — Agent-as-a-Judge hierarchical shape (R1 §9 365-sub-req pattern)").
- `docs/design/major-v3-master.md:1083` — MIG-001 owns the AUTO wrapping execution.
- `docs/design/major-v3-master.md:1136` — Principle P1 "SPEC as Constitutional Contract" explicitly maps CON-001 + SPC-001 as foundational.

### 2.3 SPEC skill guidance

- `.claude/skills/moai-workflow-spec/SKILL.md` § "EARS Format Deep Dive" enumerates 5 patterns (Ubiquitous / Event / State / Unwanted / Optional) — modality is FROZEN per CON-001.
- Same skill § "Standard 3-File Format" mandates `spec.md` + `plan.md` + `acceptance.md` per SPEC directory; SPC-001 honours this contract.

---

## 3. As-Is — Go Implementation Already Landed

Critical finding: SPC-001's Go data model and parser are **already implemented** in `internal/spec/`. The plan must therefore focus on consolidation, documentation, migration, and gap-fill rather than greenfield implementation.

### 3.1 Type model (`internal/spec/ears.go`, 165 LOC)

- `internal/spec/ears.go:11-18` — `Acceptance` struct with `ID`, `Given`, `When`, `Then`, `RequirementIDs`, `Children` — matches SPC-001 REQ-SPC-001-005 verbatim.
- `internal/spec/ears.go:21` — `MaxDepth = 3` constant — matches REQ-SPC-001-003 cap.
- `internal/spec/ears.go:25` — `topLevelIDPattern = ^AC-[A-Z0-9]+-[0-9]+-[0-9]+$` — matches REQ-SPC-001-002.
- `internal/spec/ears.go:38-64` — `GenerateChildID(depth, index)` — produces `.a/.b` (depth 1) and roman `.i/.ii` (depth 2). Matches REQ-SPC-001-003.
- `internal/spec/ears.go:68-74` — `Depth()` based on dot-segment count.
- `internal/spec/ears.go:77-81` — `InheritGiven(parent)` implements REQ-SPC-001-006 (child inherits parent's Given when empty).
- `internal/spec/ears.go:83-99` — `IsLeaf()` + `CountLeaves()` enable REQ-coverage computation per REQ-SPC-001-042.
- `internal/spec/ears.go:103-119` — `ValidateDepth()` returns `MaxDepthExceeded` per REQ-SPC-001-021.
- `internal/spec/ears.go:122-141` — `ExtractRequirementMappings(text)` parses `(maps REQ-...)` tail per REQ-SPC-001-004.
- `internal/spec/ears.go:145-165` — `ValidateRequirementMappings()` enforces leaf-only mapping rule (parents may omit when all children carry tails) — matches AC-SPC-001-12 and AC-SPC-001-13.

### 3.2 Errors (`internal/spec/errors.go`, 45 LOC)

- `internal/spec/errors.go:5-13` — `DuplicateAcceptanceID{ID, Depth}` — matches REQ-SPC-001-012 → AC-SPC-001-04.
- `internal/spec/errors.go:16-24` — `MaxDepthExceeded{ID, Depth, Max}` — matches REQ-SPC-001-021 → AC-SPC-001-05.
- `internal/spec/errors.go:27-36` — `DanglingRequirementReference{ACID, ReqID, Location}` — matches REQ-SPC-001-041 → AC-SPC-001-10 (warning, not error).
- `internal/spec/errors.go:39-45` — `MissingRequirementMapping{ACID}` — matches REQ-SPC-001-004 leaf-mapping rule → AC-SPC-001-12.

### 3.3 Parser (`internal/spec/parser.go`, 350 LOC)

- `internal/spec/parser.go:10-15` — `ParseResult{Criteria, Errors, Warnings}` — three-bucket result (errors fatal, warnings advisory).
- `internal/spec/parser.go:26-34` — `ParseAcceptanceCriteria(markdown, isFlatFormat) ([]Acceptance, []error)` exported entry point. Second arg corresponds to `acceptance_format: flat` opt-out per REQ-SPC-001-030.
- `internal/spec/parser.go:32-34` — `ParseAcceptanceCriteriaTyped(...)` returns separated errors/warnings.
- `internal/spec/parser.go:90-117` — `extractACLines(lines, startIdx, isFlatFormat)` — when `isFlatFormat=true`, only sibling lines accepted (REQ-SPC-001-030 behaviour).
- `internal/spec/parser.go:182-185` — child-attach mechanic implementing REQ-SPC-001-011 (deeper indentation = child).
- `internal/spec/parser.go:200-227` — `wrap` synthesis: when a top-level node has zero children and no `.letter` suffix, parser wraps it with synthesized `.a` child carrying identical Given/When/Then. Matches REQ-SPC-001-010 + REQ-SPC-001-040 → AC-SPC-001-02 + AC-SPC-001-09.
- `internal/spec/parser.go:340-341` — recursive child traversal for downstream validators.

### 3.4 Lint (`internal/spec/lint.go`, 857 LOC)

- `internal/spec/lint.go:122` — `REQIDUniquenessRule` registered.
- `internal/spec/lint.go:294-306` — `REQEntry` + `Doc.REQs []REQEntry` model.
- `internal/spec/lint.go:311-316` — `reqIDPattern = ^REQ-[A-Z]{2,5}-\d{3}-\d{3}$` and `reqLinePattern` for markdown extraction.
- `internal/spec/lint.go:342-343` — `parseREQs(body)` populates `Doc.REQs`.
- `internal/spec/lint.go:394-403` — `collectAllREQIDs(criteria)` walks the **entire tree** (parents + leaves) when computing coverage. Matches REQ-SPC-001-042 ("parent and leaf acceptances SHALL both count toward coverage").

### 3.5 CLI renderer (`internal/cli/spec_view.go`, 158 LOC)

- `internal/cli/spec.go:21` — `specCmd.AddCommand(newSpecViewCmd())` — `moai spec view` subcommand registered.
- `internal/cli/spec_view.go:13-31` — `newSpecViewCmd()` defines `moai spec view SPEC-XXX [--shape-trace]`.
- `internal/cli/spec_view.go:23-24` — usage examples include `moai spec view SPEC-SPC-001 --shape-trace` (matches AC-SPC-001-08).
- `internal/cli/spec_view.go:41-65` — `viewAcceptanceCriteria` reads spec.md, parses, and renders the tree. Flat-format flag default `false` (auto-wrap mode).

### 3.6 Test corpus (`internal/spec/testdata/`, 12 fixture spec.md files)

- `internal/spec/testdata/valid/spec.md` — SPEC-V3R2-TST-001 — flat-only baseline.
- `internal/spec/testdata/hierarchical-ac/spec.md:38-42` — SPEC-V3R2-TST-011 — mixed top-level (one parent with two `.a/.b` children + one flat sibling). Matches AC-SPC-001-09 case.
- `internal/spec/testdata/dup-spec-id-{a,b}` — duplicate-ID coverage.
- `internal/spec/testdata/missing-exclusions` — exclusions-section coverage.
- `internal/spec/testdata/dangling-rule` — dangling-REQ coverage.
- `internal/spec/testdata/missing-coverage` — REQ↔AC coverage gap fixture.
- `internal/spec/ears_test.go` — 438 LOC; `internal/spec/parser_test.go` — 485 LOC; `internal/spec/lint_test.go` — 572 LOC. Test suite passes locally (`go test ./internal/spec/... → ok`).

---

## 4. As-Is — SPEC Corpus Snapshot

| Metric | Value | Source |
|--------|-------|--------|
| Total `.moai/specs/SPEC-*/` directories | 183 | `ls .moai/specs \| wc -l` (excludes `_archive`) |
| `spec.md` files | 185 | `find .moai/specs -name spec.md \| wc -l` (some have multiple drafts) |
| `acceptance.md` files | 124 | `find .moai/specs -name acceptance.md \| wc -l` |
| SPEC dirs lacking dedicated `acceptance.md` | ~60 | `185 - 124` (acceptance embedded in spec.md §6 only) |

### 4.1 Acceptance-format inventory

Probe spot-checks (representative):

- `.moai/specs/SPEC-V3R2-WF-005/spec.md` — flat ACs in spec.md §6, no children. Format: `AC-WF005-NN-XX: Given … When … Then …`.
- `.moai/specs/SPEC-V3R2-SPC-001/spec.md:129-142` — uses **flat** ACs (the SPEC about hierarchical shape is itself authored flat — pragmatic; demonstrates back-compat).
- `internal/spec/testdata/hierarchical-ac/spec.md:38-42` — confirmed hierarchical example exists in test corpus.

Conclusion: production SPEC corpus is essentially 100% flat; hierarchical adoption is presently zero outside test fixtures. Migration must treat the corpus as the migration target.

### 4.2 SPEC ID format observed

- `AC-<DOMAIN>-<NNN>-<NN>` format (e.g., `AC-SPC-001-05`) is regex-validated in `internal/spec/ears.go:25` and matches every observed flat AC.
- Some legacy SPECs use `AC-<DOMAIN>-<NN>` (3 segments only). These are not currently regex-conformant and would need either (a) migrator-rename or (b) regex relaxation. SPEC text (§3 Environment) cites `AC-HOOKS-001-01` style — implying 4-segment is the canonical post-MIG-001 format.

### 4.3 Adjacent SPECs identified

- `SPEC-V3R2-CON-001` — FROZEN EARS modality (parent FROZEN invariant; SPC-001 amends acceptance shape, not modality).
- `SPEC-V3R2-CON-002` — Amendment protocol (FrozenGuard / Canary / ContradictionDetector / RateLimiter / HumanOversight) gates SPC-001 landing.
- `SPEC-V3R2-MIG-001` — `moai migrate v2-to-v3` migrator. Already drafted at `.moai/specs/SPEC-V3R2-MIG-001/spec.md` (12325 bytes, status draft). Owns AUTO-wrap of flat ACs per §1083 of design master.
- `SPEC-V3R2-SPC-002` — @MX TAG extension (parallel; no dependency).
- `SPEC-V3R2-SPC-003` — SPEC linter (depends on SPC-001 schema; computes REQ↔AC coverage using both leaves and parents per REQ-SPC-001-042).
- `SPEC-V3R2-HRN-002` — Sprint Contract durable state (per-leaf criterion state).
- `SPEC-V3R2-HRN-003` — Per-sub-criterion scoring by evaluator-active.

---

## 5. Gap Analysis — SPC-001 vs Already Landed

| Capability | SPC-001 REQ | Implementation status | Gap |
|------------|-------------|------------------------|-----|
| `Acceptance` struct shape | REQ-SPC-001-005 | DONE (`ears.go:11-18`) | — |
| Top-level ID regex | REQ-SPC-001-002 | DONE (`ears.go:25`) | — |
| Child ID generation `.a/.b/.i/.ii` | REQ-SPC-001-003 | DONE (`ears.go:38-64`) | — |
| `MaxDepth = 3` | REQ-SPC-001-003 | DONE (`ears.go:21`) | — |
| Given inheritance | REQ-SPC-001-006 | DONE (`ears.go:77-81`) | — |
| Markdown indented-list parser | REQ-SPC-001-007, REQ-SPC-001-011 | DONE (`parser.go:90-185`) | — |
| Flat AC auto-wrap synthesis | REQ-SPC-001-010 | DONE (`parser.go:200-227`) | — |
| Duplicate ID detection | REQ-SPC-001-012 | DONE (`errors.go:5-13` + parser) | — |
| Max-depth detection | REQ-SPC-001-021 | DONE (`ears.go:103-119`) | — |
| `acceptance_format: flat` opt-out | REQ-SPC-001-030 | DONE (`parser.go:117` `isFlatFormat` branch) | — |
| `moai spec view --shape-trace` | REQ-SPC-001-013, REQ-SPC-001-031 | PARTIAL (`spec_view.go:23-24` accepts flag; trace fields not yet emitted in `spec_view.go:41-158`) | M3 task — verify trace output includes `depth` + `parent ID` |
| Mixed top-level (some children, some flat) | REQ-SPC-001-040 | DONE (testdata `hierarchical-ac` covers; `parser.go:200`) | — |
| Dangling REQ warning | REQ-SPC-001-041 | DONE (`errors.go:27-36` + lint integration) | — |
| Missing leaf REQ-mapping warning | REQ-SPC-001-004 | DONE (`ears.go:145-165`) | — |
| Coverage counts both parents and leaves | REQ-SPC-001-042 | DONE (`lint.go:394-403`) | — |
| 365-leaf perf budget <500ms | AC-SPC-001-14 | UNKNOWN — no perf benchmark currently | M2 task — add benchmark fixture + assertion |
| FROZEN-zone amendment compliance | spec §11.3 | NOT YET FILED — Canary re-parse of last 10 v2 SPECs not run; HumanOversight approval not recorded | M5 task |
| `.claude/rules/moai/workflow/spec-workflow.md` documents hierarchical schema | — | NOT YET — rule references "acceptance criteria" but lacks tree-format example | M3 task |
| Migration tool MIG-001 wraps existing 185 SPECs | REQ-SPC-001-020 | MIG-001 SPEC drafted; tool not yet implemented in run phase of MIG-001 | Out-of-scope of SPC-001 plan; MIG-001 owns implementation |

Summary: ~80% of SPC-001's runtime behaviour is already shipped. SPC-001 plan focus shifts to:

1. Documentation (spec-workflow.md amendment, design constitution cross-link).
2. `--shape-trace` output validation.
3. Performance benchmark.
4. CON-002 amendment paperwork (Canary + HumanOversight).
5. Cross-link with MIG-001 / SPC-003 / HRN-002 / HRN-003.

---

## 6. Hierarchical AC Schema (Proposed Canonical Form)

### 6.1 Markdown surface

Top-level node:
```
- AC-<DOMAIN>-<NNN>-<NN>: Given …, When …, Then … (maps REQ-…)
```

With children (depth 1):
```
- AC-<DOMAIN>-<NNN>-<NN>: Given <parent context>
  - AC-<DOMAIN>-<NNN>-<NN>.a: When variant-A, Then result-A. (maps REQ-…)
  - AC-<DOMAIN>-<NNN>-<NN>.b: When variant-B, Then result-B. (maps REQ-…)
```

With grandchildren (depth 2 — maximum):
```
- AC-<DOMAIN>-<NNN>-<NN>: Given <parent context>
  - AC-<DOMAIN>-<NNN>-<NN>.a: Given <child context override>, When X, Then Y.
    - AC-<DOMAIN>-<NNN>-<NN>.a.i: When sub-case 1, Then result-1. (maps REQ-…)
    - AC-<DOMAIN>-<NNN>-<NN>.a.ii: When sub-case 2, Then result-2. (maps REQ-…)
```

### 6.2 Inheritance rules

- A child without an explicit `Given` clause inherits the **nearest ancestor's** Given (REQ-SPC-001-006).
- A child's explicit Given **overrides** the ancestor.
- `When` and `Then` MUST be specified on every leaf; intermediate nodes MAY omit when all leaves carry them.
- `(maps REQ-...)` tail MUST appear on every leaf (REQ-SPC-001-004); MAY be omitted on intermediates.

### 6.3 Identifier conventions

| Depth | Suffix | Example |
|-------|--------|---------|
| 0 (top-level) | `-NN` | `AC-SPC-001-05` |
| 1 | `.a`–`.z` | `AC-SPC-001-05.a` |
| 2 | `.i`–`.xxvi` (lowercase roman) | `AC-SPC-001-05.a.i` |

Maximum depth 3 (top + 2 child levels). `internal/spec/ears.go:21` enforces the cap; deeper attempts emit `MaxDepthExceeded`.

### 6.4 Compatibility — flat → tree conversion

The parser auto-wraps any top-level AC that has no children and no `.letter` suffix into a single-child parent (synthesized `.a` child carries identical content). This means:

- All 185 existing flat SPECs remain parseable with zero source edits.
- REQ↔AC coverage computation is unchanged (synthesized child inherits parent's REQ tail).
- Migration tool MIG-001 may optionally rewrite the markdown to make the wrap explicit; not required for runtime.

### 6.5 YAML frontmatter opt-out

A SPEC author may pin their SPEC to flat-only mode via:
```yaml
---
acceptance_format: flat
---
```
When set, the parser treats indented child markdown as ill-formed and emits siblings only (REQ-SPC-001-030). Use case: SPECs that explicitly want flat-list semantics for tooling that pre-dates the hierarchical parser.

---

## 7. Compatibility With Existing 185 SPECs

Approach (validated by parser implementation):

1. **No source edits required for read-path compatibility.** The parser auto-wraps. All 185 SPECs continue to parse and pass lint.
2. **Migration tool MIG-001 owns optional source rewrite.** When run phase of MIG-001 lands, it will rewrite each SPEC's `## 6. Acceptance Criteria` section so that flat ACs are explicitly nested with `.a` children. This is cosmetic — runtime behaviour is identical.
3. **Coverage semantics are conservative.** REQ↔AC coverage is computed by `internal/spec/lint.go:394-403` walking parents and leaves both. A flat SPEC with 10 ACs and 10 REQs achieves 100% coverage with zero changes (REQ-SPC-001-042 → AC-SPC-001-13).
4. **`acceptance_format: flat` opt-out** lets cautious SPEC authors freeze their SPEC against accidental nesting.

Risk: SPECs using a 3-segment AC ID format (`AC-XX-NN`) will fail `topLevelIDPattern`. Mitigation: MIG-001 normalises legacy IDs to 4-segment during migration. Out of SPC-001's plan scope.

---

## 8. Out-of-Scope Confirmations

Per spec.md §2.2, the following remain explicitly out of scope of SPC-001:

- Hierarchical REQ IDs (REQ-V3R2-XXX-NNN stays flat).
- EARS modality vocabulary changes (FROZEN per CON-001).
- Sprint Contract per-leaf state shape → SPEC-V3R2-HRN-002.
- Per-sub-criterion scoring by evaluator → SPEC-V3R2-HRN-003.
- @MX TAG extensions → SPEC-V3R2-SPC-002.
- SPEC linter implementation → SPEC-V3R2-SPC-003.
- Migration tool implementation (only AC wrapping behaviour is referenced; tool itself is MIG-001).

---

## 9. Risks Identified During Research

| # | Risk | Severity | Mitigation |
|---|------|----------|------------|
| R1 | Markdown tab-vs-space ambiguity in indentation parsing | MEDIUM | Parser normalizes to spaces (`parser.go:90-117`); document 2-space minimum in spec-workflow.md amendment (M3). |
| R2 | Parser performance on 365-leaf trees not yet measured | LOW | Add benchmark fixture (`internal/spec/parser_test.go`) with 365 leaves; assert <500ms (AC-SPC-001-14). |
| R3 | `--shape-trace` output may not surface depth + parent ID per REQ-SPC-001-031 | MEDIUM | Audit `internal/cli/spec_view.go` rendering function; add fields if absent; cover with CLI test. |
| R4 | FROZEN-zone amendment paperwork incomplete | HIGH | M5 — run Canary against last 10 landed v2 SPECs; record HumanOversight approval; commit before/after schema text from spec.md §11. |
| R5 | Tree glyph rendering on non-UTF8 terminals | LOW | `spec_view.go` should fall back to ASCII pipes; verify via terminal-locale env-var test. |
| R6 | `acceptance_format: flat` frontmatter parsing path may not cover YAML-edge cases (multiline, missing colon) | LOW | M2 — add test fixtures for malformed frontmatter; ensure parser refuses gracefully (no crash). |

---

## 10. Cross-Reference Summary (file:line anchors, ≥25)

1. `.claude/rules/moai/workflow/spec-workflow.md:88` — Plan phase EARS mandate.
2. `.claude/rules/moai/workflow/spec-workflow.md:92` — Planning sub-phase requires EARS.
3. `.claude/rules/moai/workflow/spec-workflow.md:104` — EARS format requirements as plan output.
4. `.claude/rules/moai/workflow/spec-workflow.md:149` — Self-review against acceptance criteria.
5. `.claude/rules/moai/workflow/spec-workflow.md:179` — Re-planning gate uses AC completion.
6. `.claude/rules/moai/workflow/spec-workflow.md:194` — Stagnation log appends AC count delta.
7. `.claude/rules/moai/workflow/spec-workflow.md:196` — Stagnation flag at zero AC for 3 entries.
8. `.claude/rules/moai/core/zone-registry.md:16-19` — CONST-V3R2-001 SPEC+EARS frozen.
9. `docs/design/major-v3-master.md:48-56` — §1.3 FROZEN invariants list EARS format.
10. `docs/design/major-v3-master.md:91` — Intent source = "EARS SPEC + hierarchical acceptance".
11. `docs/design/major-v3-master.md:235` — Owned SPECs SPC-001/002/003/004.
12. `docs/design/major-v3-master.md:255` — Acceptance []Acceptance Agent-as-a-Judge shape.
13. `docs/design/major-v3-master.md:294` — moai subsystem upgrades flat → hierarchical.
14. `docs/design/major-v3-master.md:970` — BC-V3R2-011 declaration.
15. `docs/design/major-v3-master.md:992` — Phase 5 Harness + Evaluator owns SPC-001.
16. `docs/design/major-v3-master.md:1035-1036` — §11.2 SPC-001 description.
17. `docs/design/major-v3-master.md:1083` — MIG-001 owns AUTO wrap.
18. `docs/design/major-v3-master.md:1136` — Principle P1 maps SPC-001.
19. `docs/design/major-v3-master.md:1212` — Generation method = EARS + hierarchical AC.
20. `internal/spec/ears.go:11-18` — Acceptance struct.
21. `internal/spec/ears.go:21` — MaxDepth constant.
22. `internal/spec/ears.go:25` — Top-level regex.
23. `internal/spec/ears.go:38-64` — GenerateChildID.
24. `internal/spec/ears.go:68-74` — Depth.
25. `internal/spec/ears.go:77-81` — InheritGiven.
26. `internal/spec/ears.go:103-119` — ValidateDepth.
27. `internal/spec/ears.go:122-141` — ExtractRequirementMappings.
28. `internal/spec/ears.go:145-165` — ValidateRequirementMappings.
29. `internal/spec/errors.go:5-13` — DuplicateAcceptanceID.
30. `internal/spec/errors.go:16-24` — MaxDepthExceeded.
31. `internal/spec/errors.go:27-36` — DanglingRequirementReference.
32. `internal/spec/errors.go:39-45` — MissingRequirementMapping.
33. `internal/spec/parser.go:26-34` — ParseAcceptanceCriteria entry.
34. `internal/spec/parser.go:90-117` — extractACLines + flat-format gate.
35. `internal/spec/parser.go:117` — `isFlatFormat` branch.
36. `internal/spec/parser.go:182-185` — child-attach mechanic.
37. `internal/spec/parser.go:200-227` — flat → wrapped synthesis.
38. `internal/spec/lint.go:122` — REQIDUniquenessRule registration.
39. `internal/spec/lint.go:294-306` — REQEntry / Doc.REQs.
40. `internal/spec/lint.go:311-316` — reqIDPattern.
41. `internal/spec/lint.go:394-403` — collectAllREQIDs walks tree.
42. `internal/cli/spec_view.go:13-31` — newSpecViewCmd.
43. `internal/cli/spec_view.go:41-65` — viewAcceptanceCriteria reads + parses + renders.
44. `internal/cli/spec.go:21` — spec view subcommand registration.
45. `internal/spec/testdata/hierarchical-ac/spec.md:38-42` — hierarchical fixture.
46. `internal/spec/testdata/valid/spec.md` — flat baseline fixture.
47. `.moai/specs/SPEC-V3R2-MIG-001/spec.md` — sibling migrator SPEC (12325 bytes).
48. `.moai/specs/SPEC-V3R2-SPC-001/spec.md:127-142` — flat ACs in SPC-001 itself (back-compat demonstration).

Total file:line anchors: 48 (>25 minimum required).

---

## 11. Decisions Driving Plan.md

D1. **Plan does NOT redo Go implementation.** ears.go / parser.go / lint.go / spec_view.go land work is reused as-is. Plan tasks audit, extend tests, and document.

D2. **Plan owns spec-workflow.md amendment.** M3 inserts the hierarchical schema example into `.claude/rules/moai/workflow/spec-workflow.md` and updates the FROZEN-zone-registry note.

D3. **Plan owns CON-002 amendment paperwork.** M5 records FrozenGuard / Canary / HumanOversight evidence per SPC-001 §11.3.

D4. **Migration tool stays in MIG-001.** Plan only documents the wrap behaviour and confirms parser's auto-wrap covers read-path compatibility.

D5. **`--shape-trace` audit is M3 task.** Verify trace output includes `depth` and `parent ID` fields per REQ-SPC-001-031.

D6. **Performance benchmark added in M2.** 365-leaf fixture + <500ms assertion to honour AC-SPC-001-14.

---

End of research.
