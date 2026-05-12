# Acceptance Criteria — SPEC-V3R2-SPC-001 EARS + Hierarchical Acceptance Criteria

> Self-demonstrating: this file uses the **hierarchical** AC format that SPC-001 introduces (parent → `.a/.b/.c` children inheriting parent Given). At least three ACs (AC-SPC-001-01, AC-SPC-001-09, AC-SPC-001-14) are authored as parent/child trees.
>
> Flat ACs from spec.md §6 remain canonical and unchanged; this file augments them with hierarchical Given/When/Then breakdown for each REQ as plan-phase deliverable.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-09 | manager-spec (Wave 9 plan author) | Initial hierarchical AC authoring covering 18 REQs across 5 EARS modalities. Self-demonstrates SPC-001 schema. |
| 0.1.1 | 2026-05-10 | manager-spec (Batch 3 fix) | Synchronized with spec.md v0.1.1 (no content changes, version alignment only). |

---

## 1. Format Conventions (this file)

- Top-level: `AC-SPC-001-NN` (e.g., `AC-SPC-001-01`).
- Depth-1 children: `.a / .b / .c` (lowercase letters).
- Depth-2 grandchildren: `.a.i / .a.ii` (lowercase Roman).
- Maximum depth = 3 (top + 2 child levels), enforced by `internal/spec/ears.go:21` `MaxDepth = 3`.
- Each leaf carries `(maps REQ-...)`. Intermediate nodes MAY omit when all leaves carry a tail.
- Children inherit parent's Given when the child's own Given is empty (`internal/spec/ears.go:77-81`).

## 2. REQ ↔ AC Traceability Matrix

| REQ ID | EARS modality | Mapped AC(s) | Notes |
|--------|---------------|--------------|-------|
| REQ-SPC-001-001 | Ubiquitous | AC-SPC-001-01, AC-SPC-001-14 | Tree shape per node |
| REQ-SPC-001-002 | Ubiquitous | AC-SPC-001-15 | Top-level ID regex |
| REQ-SPC-001-003 | Ubiquitous | AC-SPC-001-05, AC-SPC-001-14 | Child IDs + max depth 3 |
| REQ-SPC-001-004 | Ubiquitous | AC-SPC-001-12, AC-SPC-001-13 | Leaf REQ-mapping rule |
| REQ-SPC-001-005 | Ubiquitous | AC-SPC-001-01 | Acceptance struct fields |
| REQ-SPC-001-006 | Ubiquitous | AC-SPC-001-03 | Given inheritance |
| REQ-SPC-001-007 | Ubiquitous | AC-SPC-001-01, AC-SPC-001-16 | Markdown-list tree representation |
| REQ-SPC-001-010 | Event-driven | AC-SPC-001-02 | Auto-wrap synthesis |
| REQ-SPC-001-011 | Event-driven | AC-SPC-001-01 | Indentation = child attach |
| REQ-SPC-001-012 | Event-driven | AC-SPC-001-04 | Duplicate ID halts parser |
| REQ-SPC-001-013 | Event-driven | AC-SPC-001-06 | `moai spec view` tree output |
| REQ-SPC-001-020 | State-driven | AC-SPC-001-02, AC-SPC-001-11 | MIG-001 wrap behaviour |
| REQ-SPC-001-021 | State-driven | AC-SPC-001-05 | MaxDepth violation |
| REQ-SPC-001-030 | Optional | AC-SPC-001-07 | `acceptance_format: flat` opt-out |
| REQ-SPC-001-031 | Optional | AC-SPC-001-08 | `--shape-trace` output |
| REQ-SPC-001-040 | Complex | AC-SPC-001-09 | Mixed tree accepted |
| REQ-SPC-001-041 | Complex | AC-SPC-001-10 | Dangling REQ warning |
| REQ-SPC-001-042 | Complex | AC-SPC-001-13, AC-SPC-001-17 | Coverage counts parents+leaves |

Coverage check: 18 REQs declared in spec.md §5 (REQ-SPC-001-001 through 042, sparse numbering), every REQ has ≥1 mapped AC entry above. Total ACs in this file: 17 (14 carried from spec.md §6 + 3 plan-phase additions for full traceability and self-demonstration).

---

## 3. Acceptance Criteria (hierarchical)

### 3.1 Tree-shape representation

- AC-SPC-001-01: Given a SPEC with hierarchical acceptance markdown
  - AC-SPC-001-01.a: Given a top-level entry `AC-X-01` with two indented child entries `AC-X-01.a` and `AC-X-01.b`, When the parser at `internal/spec/parser.go:182-185` builds the tree, Then it produces `Acceptance{ID: "AC-X-01", Children: [{ID: "AC-X-01.a"}, {ID: "AC-X-01.b"}]}`. (maps REQ-SPC-001-001, REQ-SPC-001-005, REQ-SPC-001-011)
  - AC-SPC-001-01.b: Given the same input, When the parser exposes fields via `internal/spec/ears.go:11-18`, Then `ID`, `Given`, `When`, `Then`, `RequirementIDs`, and `Children` are populated for every node. (maps REQ-SPC-001-005, REQ-SPC-001-007)
  - AC-SPC-001-01.c: Given a markdown list-item indented at 2+ spaces below its parent, When the parser determines depth via `internal/spec/ears.go:68-74` `Depth()`, Then the deeper item attaches as a child of the closest preceding shallower item. (maps REQ-SPC-001-007, REQ-SPC-001-011)

### 3.2 Auto-wrap of flat legacy SPECs

- AC-SPC-001-02: Given a flat legacy SPEC with `AC-HOOKS-001-01: Given X, When Y, Then Z (maps REQ-HOOKS-001-001)`
  - AC-SPC-001-02.a: When the parser at `internal/spec/parser.go:200-227` encounters a top-level node with no `.letter` suffix and no indented children, Then it synthesises a single child `AC-HOOKS-001-01.a` with verbatim Given/When/Then/REQ tail. (maps REQ-SPC-001-010, REQ-SPC-001-020)
  - AC-SPC-001-02.b: When `ParseAcceptanceCriteria` returns the tree, Then `result.Criteria[0].Children` has length 1 and `result.Criteria[0].Children[0].RequirementIDs == ["HOOKS-001-001"]`. (maps REQ-SPC-001-010)

### 3.3 Given inheritance

- AC-SPC-001-03: Given a parent acceptance `AC-X-02` with Given clause "user authenticated" and a child `AC-X-02.a` with empty Given, When the parser invokes `internal/spec/ears.go:77-81` `InheritGiven(parent)`, Then `AC-X-02.a.Given` resolves to "user authenticated". (maps REQ-SPC-001-006)

### 3.4 Duplicate ID detection

- AC-SPC-001-04: Given a SPEC where two acceptance entries share the same ID at the same depth (e.g., two `AC-X-03.a` siblings), When the parser runs, Then it emits `*spec.DuplicateAcceptanceID{ID: "AC-X-03.a", Depth: 1}` (`internal/spec/errors.go:5-13`) and halts construction of that subtree. (maps REQ-SPC-001-012)

### 3.5 Max-depth enforcement

- AC-SPC-001-05: Given a SPEC with an acceptance at 4-segment dotted form `AC-X-01.a.i.x` (depth = 4 — beyond MaxDepth), When `ValidateDepth()` (`internal/spec/ears.go:103-119`) traverses the tree, Then it returns `*spec.MaxDepthExceeded{ID: "AC-X-01.a.i.x", Depth: 3, Max: 2}` and parsing fails fatally. (maps REQ-SPC-001-021, REQ-SPC-001-003)

### 3.6 `moai spec view` tree rendering

- AC-SPC-001-06: Given an existing hierarchical SPEC parsed by `internal/cli/spec_view.go:41-65`, When the user runs `moai spec view SPEC-V3R2-SPC-001`, Then the output indents children under parents with tree glyphs (e.g., `├─` `└─`) such that visual nesting matches `Acceptance.Children` traversal. (maps REQ-SPC-001-013)

### 3.7 Flat-format opt-out via frontmatter

- AC-SPC-001-07: Given a SPEC with YAML frontmatter `acceptance_format: flat`, When indented child markdown appears below an `AC-X-NN` entry, Then `internal/spec/parser.go:117` flat-format branch ignores the indentation and produces sibling entries only (no Children attached). (maps REQ-SPC-001-030)

### 3.8 `--shape-trace` flag

- AC-SPC-001-08: Given the `moai spec view SPEC-XXX --shape-trace` invocation, When the renderer iterates the parsed tree, Then each node's output line includes its `depth` (0/1/2) and `parent_id` (empty for top-level, parent's ID otherwise) fields. (maps REQ-SPC-001-031)

### 3.9 Mixed tree (auto-wrap + explicit children)

- AC-SPC-001-09: Given a SPEC §6 containing two top-level entries — `AC-X-04` with explicit children `.a` and `.b`, and `AC-X-05` with no children
  - AC-SPC-001-09.a: When the parser processes the section, Then `AC-X-04` retains its explicit `[.a, .b]` children list. (maps REQ-SPC-001-040)
  - AC-SPC-001-09.b: When the parser processes the section, Then `AC-X-05` is auto-wrapped with a synthesised single child `AC-X-05.a` carrying identical Given/When/Then. (maps REQ-SPC-001-040, REQ-SPC-001-010)
  - AC-SPC-001-09.c: When the parser produces the result, Then no warning is emitted for the mixed shape. (maps REQ-SPC-001-040)

### 3.10 Dangling REQ-mapping warning

- AC-SPC-001-10: Given an acceptance `AC-X-06.a: ... (maps REQ-X-042)` where `REQ-X-042` is not declared in the Requirements section, When the parser cross-references via `internal/spec/lint.go:294-403`, Then a `*spec.DanglingRequirementReference` warning (`internal/spec/errors.go:27-36`) is emitted but parsing continues and the tree is built. (maps REQ-SPC-001-041)

### 3.11 Migration tool 1-child wrap

- AC-SPC-001-11: Given the migration tool `SPEC-V3R2-MIG-001` runs against a legacy SPEC with 8 flat ACs and 8 REQs, When the post-migration spec.md is re-parsed by `internal/spec/parser.go`, Then the tree contains 8 top-level nodes each with exactly one synthesised `.a` child and no Given/When/Then content is lost. (maps REQ-SPC-001-020)

### 3.12 Missing REQ tail on a leaf

- AC-SPC-001-12: Given a top-level acceptance with no children and no `(maps REQ-...)` tail, When the parser auto-wraps it (synthesised `.a` leaf inherits no mapping), Then `ValidateRequirementMappings()` (`internal/spec/ears.go:145-165`) emits a `*spec.MissingRequirementMapping{ACID: "AC-X-07.a"}` warning targeting the synthesised leaf. (maps REQ-SPC-001-004)

### 3.13 Parent omits REQ tail when all children carry it

- AC-SPC-001-13: Given a parent `AC-X-08` with no REQ tail and three children `.a/.b/.c` each carrying distinct `(maps REQ-X-001)` / `(maps REQ-X-002)` / `(maps REQ-X-003)` tails, When `ValidateRequirementMappings()` runs, Then no warning is emitted because every leaf has its own mapping. (maps REQ-SPC-001-004, REQ-SPC-001-042)

### 3.14 DevAI-shape performance

<!-- @MX:ANCHOR fan_in=4 -->
<!-- @MX:REASON: "self-demonstrating hierarchical AC reference cited by SPC-002/SPC-003/HRN-002/HRN-003" -->
- AC-SPC-001-14: Given a SPEC with 365 leaf acceptance nodes distributed across 55 top-level parents (Agent-as-a-Judge DevAI shape per spec.md §1)
  - AC-SPC-001-14.a: When the parser at `internal/spec/parser.go:36-58` processes the SPEC, Then it succeeds within a 500ms wall-clock budget on a baseline laptop. (maps REQ-SPC-001-001, REQ-SPC-001-003)
  - AC-SPC-001-14.b: When the resulting tree is traversed, Then `Acceptance.CountLeaves()` (`internal/spec/ears.go:89-99`) returns 365. (maps REQ-SPC-001-001)
  - AC-SPC-001-14.c: When `go test -bench BenchmarkParse365Leaves ./internal/spec/...` runs, Then the benchmark report shows mean parse time below 500_000_000 ns/op (500 ms). (maps REQ-SPC-001-001, REQ-SPC-001-003)

### 3.15 Top-level ID regex

- AC-SPC-001-15: Given an acceptance ID `AC-XYZ-007-42`, When `Acceptance.ValidateID()` (`internal/spec/ears.go:28-33`) runs the `topLevelIDPattern` regex, Then the ID is accepted. Conversely given `AC-X-07` (3 segments), Then `ValidateID()` returns an error. (maps REQ-SPC-001-002)

### 3.16 Markdown list-item nesting

- AC-SPC-001-16: Given a markdown source where a `- AC-X-09` line is followed by a `  - AC-X-09.a` line (2-space indent), When `extractACLines` (`internal/spec/parser.go:90-117`) extracts entries, Then the second line's `indentLevel` is 1 greater than the first's, signalling child relationship to `buildTree`. (maps REQ-SPC-001-007)

### 3.17 REQ↔AC coverage walks tree

- AC-SPC-001-17: Given a hierarchical SPEC with parents that omit REQ tails and leaves that carry them, When `internal/spec/lint.go:394-403` `collectAllREQIDs(criteria)` computes coverage, Then the returned set includes every REQ declared on **both** parent and leaf nodes; a SPEC achieves 100% coverage when every declared REQ appears at least once across the tree. (maps REQ-SPC-001-042)

---

## 4. Edge Cases

| # | Scenario | Expected behaviour | Anchor |
|---|----------|--------------------|--------|
| E1 | Tab-indented children | Parser normalises tabs to spaces; minimum 2-space equivalent for child attach | `internal/spec/parser.go:90-117` |
| E2 | Empty Acceptance Criteria section | Parser returns empty tree, no error | `internal/spec/parser.go:48-50` |
| E3 | Acceptance section heading variation (`## 6.` vs `## Acceptance Criteria`) | Parser tolerates both via `findACSectionStart` | `internal/spec/parser.go:36-50` |
| E4 | Roman numeral overflow (>26 children at depth 2) | Parser emits error: `invalid index for level 2 child` | `internal/spec/ears.go:58-60` |
| E5 | Mixed lowercase / uppercase letter children (`.a` then `.B`) | Parser treats as different IDs (uppercase fails ID regex) | `internal/spec/ears.go:25` |
| E6 | Inline children on a single line (no markdown nesting) | Not supported; parser relies on indent. Author MUST use newline + indent. | `internal/spec/parser.go:90-185` |
| E7 | YAML `acceptance_format` value other than `flat` (e.g., `tree`, `nested`) | Parser ignores unknown values and defaults to auto-wrap behaviour | `internal/spec/parser.go:117` |

---

## 5. Quality Gate Criteria (Definition of Done)

### 5.1 Plan-phase DoD (this PR)

- [ ] All 17 ACs above carry at least one `(maps REQ-...)` reference.
- [ ] At least three ACs (AC-SPC-001-01, AC-SPC-001-09, AC-SPC-001-14) self-demonstrate the hierarchical schema with explicit `.a/.b/.c` children.
- [ ] Every REQ in spec.md §5 (REQ-SPC-001-001 through REQ-SPC-001-042) appears at least once in §2 traceability matrix.
- [ ] Plan-auditor PASS at iteration ≤2.

### 5.2 Run-phase DoD (future PR per tasks.md)

- [ ] `go test ./internal/spec/...` green (existing 4 test files preserved).
- [ ] `go test -bench BenchmarkParse365Leaves` <500ms (AC-SPC-001-14.c).
- [ ] `moai spec view --shape-trace` test fixture asserts depth + parent_id fields (AC-SPC-001-08).
- [ ] CON-002 Canary evidence committed (10 v2 SPECs re-parsed without warning).
- [ ] HumanOversight approval recorded in landing PR description.

### 5.3 Sync-phase DoD

- [ ] `.claude/rules/moai/workflow/spec-workflow.md` documents hierarchical schema.
- [ ] `.claude/skills/moai-workflow-spec/SKILL.md` body updated with hierarchical AC subsection.
- [ ] `.claude/rules/moai/core/zone-registry.md` CONST-V3R2-001 entry cross-links SPC-001.
- [ ] CHANGELOG entry under `### Changed` cites BC-V3R2-011.

---

## 6. Self-Demonstration Notice

This file uses the new hierarchical AC schema (depth 0 → 1) on AC-SPC-001-01, AC-SPC-001-02, AC-SPC-001-09, AC-SPC-001-14. Specifically:

- AC-SPC-001-01 has 3 depth-1 children covering type/struct/markdown invariants.
- AC-SPC-001-02 has 2 depth-1 children covering parser-side wrapping behaviour.
- AC-SPC-001-09 has 3 depth-1 children covering mixed-tree, auto-wrap, and warning behaviours.
- AC-SPC-001-14 has 3 depth-1 children covering perf budget, leaf count, and benchmark assertion.

The remaining ACs are flat — demonstrating that flat and hierarchical co-exist within the same SPEC (REQ-SPC-001-040 → AC-SPC-001-09).

End of acceptance.
