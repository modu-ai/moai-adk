---
id: SPEC-V3R6-AGENTIC-LOOP-CONFIG-001
title: "Go-side loader registration for workflow.agentic_loop.max_iterations (prose-read → 기계적 상한)"
version: "0.1.0"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P2
phase: "v3.x"
module: "internal/config"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "config, workflow, agentic-loop, loader, distinctness-invariant, follow-up"
depends_on:
  - SPEC-MOAI-AGENTIC-LOOP-001
---

# SPEC-V3R6-AGENTIC-LOOP-CONFIG-001 — Implementation Plan

## §A. Context

### §A.1 Parent-SPEC scope decision

`SPEC-MOAI-AGENTIC-LOOP-001` (status: completed, push `9be283c32`) explicitly deferred Go-side registration in its §E "Out of Scope — Go code changes" (spec.md line 190). The parent owns the orchestration/skill-body prose; this SPEC owns the loader. The boundary is clean — no overlap.

### §A.2 Current ground-truth state (verified this session)

The following facts were verified by the orchestrator on 2026-07-08 using the domain's mechanical tools (grep + file reads, per verification-claim-integrity §1.1 surface 3 — defect claims verified by tool output, not text-pattern inference):

| Claim | Verification command | Result |
|---|---|---|
| `workflow.yaml` ships `agentic_loop.max_iterations: 10` | `Read .moai/config/sections/workflow.yaml` lines 6-7 | CONFIRMED |
| `workflow.yaml` ships `loop_prevention.max_iterations: 100` | `Read .moai/config/sections/workflow.yaml` lines 15-17 | CONFIRMED |
| Template `workflow.yaml` ships the `agentic_loop:` block (parity) | `grep agentic_loop template/.../workflow.yaml` | CONFIRMED (lines 7-11) |
| `types.go` has 4 distinct `MaxIterations` fields, none for agentic_loop | `grep MaxIterations types.go` | CONFIRMED (lines 295, 378, 768-769, 1034) |
| `defaults.go:436` sets `MaxIterations: 100` (loop_prevention only) | `Read defaults.go` line 436 | CONFIRMED |
| `schema_sections.go:194` registers `loop_prevention.max_iterations`, not `agentic_loop` | `Read schema_sections.go` line 194 | CONFIRMED |
| `agentic_loop` is Go-absent across `internal/config`, `internal/settings`, `internal/cli` | `grep -rn 'agentic_loop\|AgenticLoop' internal/{config,settings,cli}/` | CONFIRMED (no matches, excluding tests) |
| `WorkflowConfig` struct already exists at types.go:318 | `Read types.go` lines 309-355 | CONFIRMED — the struct has a nested-field convention (`AutoClear`, `LoopPrevention`, `Team`, etc.); the new `AgenticLoop` field follows the same pattern |

### §A.3 Existing-struct decision (HARD architectural finding)

**Blocker-status: NONE.** The `WorkflowConfig` struct (types.go:318-355) already exists and already hosts nested sub-structs (`AutoClear AutoClearConfig`, `LoopPrevention LoopPreventionConfig`, `Team TeamConfig`, etc.). The new field is a straightforward addition to this existing struct — no new top-level struct, no refactor of the four existing `MaxIterations` fields elsewhere in `types.go`.

This resolves the "Blocker 3" the task description anticipated: **a Workflow-section struct already exists that should host the field, rather than a new struct**. The plan treats this as a one-line addition to `WorkflowConfig` plus a new sibling config struct `AgenticLoopConfig` (mirroring the `AutoClearConfig` / `LoopPreventionConfig` pattern at types.go:367-380).

## §B. Known Issues / Pre-flight Audit (verification-claim-integrity)

### §B.1 Anti-aliasing hazard (the critical risk)

The single highest-severity risk is **field aliasasing**: an implementer under time pressure might be tempted to reuse the pre-existing `LoopPrevention.MaxIterations` field for the agentic-loop key (or vice versa) "since they're both called max_iterations". This would silently destroy the §A.4 distinctness invariant (spec.md).

**Mitigation**: REQ-ALC-005, REQ-ALC-006, REQ-ALC-008, REQ-ALC-010, and AC-DISTINCT-001 all enforce distinctness. The dedicated anti-aliasing test (AC-DISTINCT-002) parses a fixture with both keys set to distinct custom values and asserts both land in the correct separate Go fields. The run-phase TDD cycle is structured so the anti-aliasing test is written FIRST (RED), before any production struct change — the test fails until the new field exists and parses independently.

### §B.2 Default-zero hazard

The default zero-value of `int` in Go is `0`. If the loader does not explicitly initialize `AgenticLoop.MaxIterations`, a YAML fixture omitting the block parses to `0`, not `10`. A ceiling of `0` would either (a) halt the pipeline immediately (if enforced), or (b) be interpreted as "no ceiling" (if used as an upper bound in a `for i := 0; i < max; i++` loop). Both are wrong.

**Mitigation**: REQ-ALC-003 (single const `DefaultAgenticLoopMaxIterations = 10` in defaults.go), REQ-ALC-004 (loader synthesizes the default on absent block), and REQ-ALC-002 (explicit invariant: absent block → `10`, NEVER zero, NEVER `100`).

### §B.3 Literal-hardcoding hazard (CLAUDE.local.md §14)

The literal `10` is tempting to inline in `types.go` (struct tag comment), `loader.go` (default branch), or `_test.go` (expected-value constant).

**Mitigation**: REQ-ALC-003 explicitly forbids the literal `10` outside `defaults.go`. Tests reference the const, not the literal.

## §C. Pre-flight Checks (Run-phase Entry)

The following MUST be true before run-phase M1 begins. Run-phase agent (manager-develop, cycle_type=tdd) MUST verify and record evidence:

1. `git status` clean (no unrelated staged changes).
2. `go test ./internal/config/... ./internal/settings/...` baseline PASS recorded.
3. `grep -rn 'AgenticLoop\|agentic_loop' internal/ | grep -v _test.go` baseline returns empty (pre-implementation, the key is Go-absent).
4. `.moai/specs/SPEC-V3R6-AGENTIC-LOOP-CONFIG-001/{spec,plan,acceptance}.md` exist and `status: draft`.

## §D. Constraints (restated for run-phase visibility)

- **CLAUDE.local.md §14**: no hardcoding; default is a named const in `defaults.go`.
- **CLAUDE.local.md §3**: file naming `snake_case.go`, error wrapping `fmt.Errorf("...: %w", err)`, all code/comments/godoc in English.
- **`internal/config/CLAUDE.md`**: non-strict YAML mode; unknown keys silently ignored; only type mismatches are `ConfigTypeError`.
- **Era integrity**: parent `SPEC-MOAI-AGENTIC-LOOP-001` MUST NOT be modified.
- **No snake_case aliases in frontmatter**: use `created`/`updated`, never `created_at`/`updated_at`.

## §E. Self-Verification (Plan-phase Audit-Ready)

Plan-phase audit-ready signal: the spec.md, plan.md, and acceptance.md files are internally consistent — every REQ-ALC-* in spec.md §B has a traceable acceptance criterion in acceptance.md §D; every file in the milestone plan below is named explicitly; the distinctness invariant (§A.4 / §B.1) has a dedicated AC (AC-DISTINCT-001) and a dedicated anti-aliasing test (AC-DISTINCT-002). The era is V3R6 (matches parent). The SPEC ID passed the Pre-Write Self-Check Protocol (decomposition printed in the authoring turn).

## §F. Milestones (priority-ordered, no time estimates)

### M1 — RED: anti-aliasing + default-fallback test suite (REQ-ALC-009, REQ-ALC-010, AC-DISTINCT-001/002)

**Priority**: P0 (highest — the distinctness invariant is the single most important guarantee of this SPEC).

**Files**:
- `internal/config/loader_test.go` (or a new `loader_agentic_loop_test.go` — implementer's choice; one file is preferred for locality) — add table-driven tests:
  - Case (a): explicit custom value parses correctly.
  - Case (b): absent `agentic_loop:` block → `DefaultAgenticLoopMaxIterations` (the const, never the literal `10`).
  - Case (c): present `agentic_loop:` parent, absent `max_iterations` sub-key → `DefaultAgenticLoopMaxIterations`.
  - Case (d) (AC-DISTINCT-001): both blocks present with distinct values → both parse to distinct fields.
- `internal/config/agentic_loop_distinctness_test.go` — dedicated anti-aliasing test (AC-DISTINCT-002):
  - Parse a YAML fixture setting `agentic_loop.max_iterations: 7` and `loop_prevention.max_iterations: 99`.
  - Assert `cfg.Workflow.AgenticLoop.MaxIterations == 7` AND `cfg.Workflow.LoopPrevention.MaxIterations == 99`.
  - Negative assertion: setting one does NOT propagate to the other.

**Expected RED state**: all tests fail to compile (the `AgenticLoop` field does not exist yet) or fail at the assertion (field missing).

**Exit criteria**: tests exist, fail for the right reason (missing struct field), git-commit-attributed to the run-phase agent.

### M2 — GREEN: types + defaults + loader + schema registration (REQ-ALC-001..008)

**Priority**: P1.

**Files (file-by-file change plan)**:

1. **`internal/config/types.go`** (REQ-ALC-001, REQ-ALC-005, REQ-ALC-006):
   - Add a new struct `AgenticLoopConfig` mirroring the `AutoClearConfig` / `LoopPreventionConfig` pattern (types.go:367-380). Single field: `MaxIterations int \`yaml:"max_iterations"\``.
   - Add a new field to `WorkflowConfig` (types.go:318-355): `AgenticLoop AgenticLoopConfig \`yaml:"agentic_loop"\``, placed adjacent to `LoopPrevention` for locality (comment cross-referencing §A.4 distinctness invariant).
   - Godoc on `AgenticLoopConfig` explicitly names the distinctness-from-`LoopPrevention` invariant.

2. **`internal/config/defaults.go`** (REQ-ALC-003):
   - Add a new named const: `DefaultAgenticLoopMaxIterations = 10` (in the `const (...)` block near `DefaultMaxIterations = 5` at defaults.go:33, OR a standalone const — implementer's choice).
   - In `NewDefaultWorkflowConfig()` (defaults.go:424-506), add an `AgenticLoop: AgenticLoopConfig{MaxIterations: DefaultAgenticLoopMaxIterations}` entry — placed adjacent to the existing `LoopPrevention: LoopPreventionConfig{...}` entry at line 434 for locality.
   - The literal `10` appears EXACTLY ONCE in the const declaration; everywhere else references the const.

3. **`internal/config/loader.go` (or section-specific loader)** (REQ-ALC-002, REQ-ALC-004):
   - The workflow section is already loaded via `Loader.loadWorkflowSection` (or equivalent — the loader chain in `Loader.Load()` already includes `workflow.yaml` → `cfg.Workflow`). The YAML unmarshal of `workflow.yaml` into `WorkflowConfig` automatically picks up the new `AgenticLoop` field because yaml.v3 uses the struct tag — **no new loader code is strictly required** if the existing workflow loader already unmarshals the top-level `workflow:` block into `WorkflowConfig`. The implementer MUST verify this (read `loader.go` or `loader_workflow.go`) and, if the existing loader only unmarshals specific sub-blocks (not the whole `WorkflowConfig`), add the wiring.
   - If the existing loader is already a whole-section unmarshal into `WorkflowConfig` (most likely), M2 loader.go change is a no-op — the test (M1) verifies the field is populated.
   - REQ-ALC-004 (non-strict unknown-key policy) is inherited from the existing loader's `strictUnmarshalSection` policy — no new code.

4. **`internal/settings/schema_sections.go`** (REQ-ALC-007, REQ-ALC-008):
   - Add a new line immediately BEFORE the `loop_prevention` block — between `auto_clear.token_threshold` (line 192) and `loop_prevention.failure_pattern_detection` (line 193): `s(SectionWorkflow, "workflow", TypeInt, "workflow", "agentic_loop", "max_iterations"),`. This places `agentic_loop` ahead of `loop_prevention` and `team` (alphabetical: `agentic_loop` < `loop_prevention` < `team`) and does NOT split the `loop_prevention` block, which spans lines 193-195 as three consecutive entries (`failure_pattern_detection`, `max_iterations`, `max_retries_per_operation`). The earlier guidance "insert at approximately line 195 (immediately after line 194)" was a D3 defect — it would have placed the new entry between `loop_prevention.max_iterations` (line 194) and `loop_prevention.max_retries_per_operation` (line 195), splitting the block. Re-verify the exact line numbers before editing — the file may have drifted since this plan was authored (line numbers cited are as of 2026-07-08).
   - Confirm the path tokens are `... "agentic_loop", "max_iterations"` (distinct from `... "loop_prevention", "max_iterations"` at line 194) — REQ-ALC-008.

5. **`internal/config/audit_struct_yaml_symmetry_test.go`** (if the test enforces struct-yaml symmetry — see `internal/config/CLAUDE.md` "CI Guards"):
   - Add `WorkflowConfig.AgenticLoop` / `AgenticLoopConfig.MaxIterations` to the symmetryCases IF the existing test enumerates struct fields. If the test uses reflection, no change needed.
   - Implementer verifies by running `go test ./internal/config/... -run TestStructYAMLSymmetry` — if it fails, add the entry; if it passes, no change.

**Exit criteria**: M1 tests now PASS (GREEN). `go build ./...` succeeds. `grep -rn 'AgenticLoop\|agentic_loop' internal/ | grep -v _test.go` returns the new sites (types.go + defaults.go + schema_sections.go, possibly loader.go).

### M3 — REFACTOR + coverage verification

**Priority**: P2.

**Files**: none mandatory (refactor only if M2 introduced duplication).

**Tasks**:
- Run `go test -cover ./internal/config/... ./internal/settings/...` and record coverage delta. Coverage for `internal/config` MUST NOT decrease.
- Run `go vet ./...` and `golangci-lint run --timeout=2m` (per CLAUDE.local.md §4 lint baseline).
- If the implementer introduced any inline literal `10` in test fixtures, refactor to reference `DefaultAgenticLoopMaxIterations` (REQ-ALC-003 compliance audit).
- Optional: add a `// @MX:ANCHOR` tag on the `AgenticLoopConfig` godoc naming the distinctness invariant as a fan_in ≥ 1 contract (cross-reference to `LoopPreventionConfig`). Per `.claude/rules/moai/workflow/mx-tag-protocol.md` § When to Add Tags — ANCHOR for "public API boundary" / "external system integration point". The `agentic_loop` field is a public API boundary between the YAML file and the Go runtime.

**Exit criteria**: coverage non-decreasing; lint clean; no inline `10` literals outside defaults.go.

## §G. Anti-Patterns (run-phase agent MUST avoid)

- **AP-1**: Aliasing `AgenticLoop.MaxIterations` to `LoopPrevention.MaxIterations` (e.g. via a getter that falls back). HARD violation of §A.4.
- **AP-2**: Inlining `10` in `_test.go` expected-value constants. Use `DefaultAgenticLoopMaxIterations`.
- **AP-3**: Inlining `10` in a struct tag default comment. The comment may say "default is `DefaultAgenticLoopMaxIterations` (= 10)" — the literal is allowed only in the const declaration and in prose comments.
- **AP-4**: Modifying `.moai/specs/SPEC-MOAI-AGENTIC-LOOP-001/` (parent SPEC is completed/3-phase-closed; re-opening violates era integrity).
- **AP-5**: Adding an env-var override without (a) declaring it in `internal/config/envkeys.go` and (b) scoping it to this SPEC's requirements. REQ-MAL-023 requires NO env-var — adding one is scope-creep (see spec.md §E "Out of Scope — Env-var override").
- **AP-6**: Modifying `internal/template/templates/.moai/config/sections/workflow.yaml`. The template already ships the block (verified §A.2); this SPEC is Go-only.
- **AP-7**: Skipping the RED phase (M1) — writing the production struct first, then the tests. HARD violation of the TDD cycle_type=tdd contract.
- **AP-8**: Refactoring any of the 4 pre-existing `MaxIterations` fields (types.go:295/378/768/1034). Out of scope (spec.md §E "Out of Scope — Refactoring the existing MaxIterations fields").

## §H. Cross-References

- Parent SPEC: `.moai/specs/SPEC-MOAI-AGENTIC-LOOP-001/spec.md` REQ-MAL-023 (line 108) + §E line 190 (Out of Scope follow-up record).
- Config loader conventions: `internal/config/CLAUDE.md` (non-strict YAML, section-file layout, env-var naming, `Loader.Load()` chain).
- Hardcoding policy: `CLAUDE.local.md §14` (Go code 하드코딩 방지 — thresholds as const, single source).
- Test isolation: `CLAUDE.local.md §6` (`t.TempDir()`, no `~/.claude/settings.json` modification).
- GEARS notation: `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format.
- SPEC frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields, 8-value status enum).
- MX tag protocol: `.claude/rules/moai/workflow/mx-tag-protocol.md` (ANCHOR tag for public API boundary — optional M3 task).
- Verification-claim integrity: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (defect claims verified by domain tool, not text-pattern inference — applied in §A.2 above).
