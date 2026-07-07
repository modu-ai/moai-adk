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
related_specs:
  - SPEC-MOAI-AGENTIC-LOOP-001
---

# SPEC-V3R6-AGENTIC-LOOP-CONFIG-001

## HISTORY

- **2026-07-08** — Plan-phase artifact authored (manager-spec). Created as the deferred Go-side follow-up of `SPEC-MOAI-AGENTIC-LOOP-001` §E "Out of Scope — Go code changes" (spec.md line 190). Parent SPEC closed the prose/skill-body surface (`/moai` agentic completion loop router, 6-mode wiring, REQ-MAL-001..027); this SPEC closes the loader surface. No prior drafts.

## §A. Overview

### §A.1 Problem

`SPEC-MOAI-AGENTIC-LOOP-001` REQ-MAL-023 documents a pipeline-level iteration ceiling key `workflow.agentic_loop.max_iterations` (default 10) in `.moai/config/sections/workflow.yaml`. The orchestrator prose-reads this key today. The Go config loader (`internal/config`) has no struct field for it, no default, and no `internal/settings/schema_sections.go` registration — the key is Go-invisible. A parser-validated enforcement ceiling requires a mechanically-parsed Go field; prose-read cannot be mechanically trusted.

### §A.2 Scope

**In scope** (this SPEC):

- Add a typed Go field for `workflow.agentic_loop.max_iterations` to `internal/config/types.go`.
- Set a single-source-of-truth default const in `internal/config/defaults.go` (per CLAUDE.local.md §14).
- Wire the loader so the `agentic_loop:` block is parsed.
- Register the key in `internal/settings/schema_sections.go`.
- Tests, including a dedicated distinctness-invariant test.

**Out of scope** (see §E).

### §A.3 Parent-SPEC provenance

This SPEC is the deferred Go-side follow-up of `SPEC-MOAI-AGENTIC-LOOP-001`:

- **REQ-MAL-023** (parent spec.md line 108): _"The loop **shall** enforce a max-iteration ceiling on pipeline iterations (default 10), configurable via a documented `workflow.agentic_loop.max_iterations` key in `.moai/config/sections/workflow.yaml` (prose-read by the orchestrator; Go-side key registration is out of scope — see §E)."_
- **Parent §E "Out of Scope — Go code changes"** (spec.md line 190): _"`workflow.agentic_loop.max_iterations` Go-side config-key registration (`internal/config`) — recorded as follow-up; this SPEC documents the key and reads it as prose only."_

This SPEC delivers that follow-up: **prose-read → 기계적 상한 (mechanical ceiling)**.

### §A.4 Critical-distinctness invariant (HARD)

`agentic_loop.max_iterations` (pipeline-level iterations, default 10) is **DISTINCT** from the pre-existing `loop_prevention.max_iterations: 100` (per-operation / diagnostic fix-loop bound). They occupy separate YAML blocks, carry separate semantics, and **MUST NOT collide** in the Go struct layer:

| YAML key | YAML block | Default | Semantics | Go struct (target state) |
|---|---|---|---|---|
| `agentic_loop.max_iterations` | `workflow.agentic_loop` | **10** | Pipeline-level completion-loop iteration ceiling (REQ-MAL-023) | `WorkflowConfig.AgenticLoop.MaxIterations` (NEW) |
| `loop_prevention.max_iterations` | `workflow.loop_prevention` | **100** | Per-operation diagnostic fix-loop bound (pre-existing) | `WorkflowConfig.LoopPrevention.MaxIterations` (existing) |

The Go loader MUST preserve this distinctness — the two keys MUST parse to **separate Go fields with separate defaults** and MUST NOT be merged, aliased, or fallback-chained into a single field. This is codified as a dedicated acceptance criterion (see acceptance.md AC-DISTINCT-001).

## §B. Requirements (GEARS)

### §B.1 Core loader requirements

- **REQ-ALC-001** (Ubiquitous) — The `internal/config` package **shall** expose a typed Go field that mechanically parses the `workflow.agentic_loop.max_iterations` YAML key. The field SHALL be reachable as `WorkflowConfig.AgenticLoop.MaxIterations` (or a structurally equivalent path with distinct semantics from `LoopPrevention.MaxIterations`).

- **REQ-ALC-002** (Ubiquitous) — The `internal/config` loader **shall** populate the REQ-ALC-001 field when parsing `.moai/config/sections/workflow.yaml`. An absent `agentic_loop:` block in the YAML MUST fall back to the default `10` (REQ-ALC-003), NEVER to zero, NEVER to `LoopPrevention.MaxIterations`'s `100`.

- **REQ-ALC-003** (Ubiquitous, single-source-of-truth) — The default value of `agentic_loop.max_iterations` **shall** be `10`, declared exactly once as a named const in `internal/config/defaults.go` (e.g. `DefaultAgenticLoopMaxIterations = 10`). The literal `10` MUST NOT be hardcoded in any other file (per CLAUDE.local.md §14 — no hardcoding). Tests reference the const, not the literal.

- **REQ-ALC-004** (Capability gate) — **Where** the `agentic_loop:` YAML block is absent or the `max_iterations` sub-key is omitted, the loader **shall** synthesize the default (REQ-ALC-003) without raising an error. YAML-non-strict semantics apply: unknown sibling keys under `agentic_loop:` are silently ignored (consistent with the loader's `strictUnmarshalSection` policy documented in `internal/config/CLAUDE.md`).

- **REQ-ALC-005** (State-driven) — **While** the distinctness invariant (§A.4) is in force, the loader **shall** parse `agentic_loop.max_iterations` and `loop_prevention.max_iterations` to **two distinct Go struct fields**. The two fields MUST NOT be aliases of one another; setting one to a custom value MUST NOT propagate to the other.

- **REQ-ALC-006** (Unwanted) — The loader **shall not** merge `agentic_loop.max_iterations` into the pre-existing `LoopPreventionConfig.MaxIterations` field. A regression that aliases the two is a HARD violation of §A.4 and is detected by AC-DISTINCT-001 (see acceptance.md).

### §B.2 Schema registration requirements

- **REQ-ALC-007** (Ubiquitous) — `internal/settings/schema_sections.go` **shall** register the key `workflow.agentic_loop.max_iterations` as `TypeInt`, in a form consistent with the adjacent `workflow.loop_prevention.max_iterations` registration at line 194.

- **REQ-ALC-008** (Unwanted) — The `internal/settings/schema_sections.go` registration **shall not** reuse the `loop_prevention` path token for the `agentic_loop` key. The two registrations MUST carry distinct path-token sequences (`... "agentic_loop", "max_iterations"` vs `... "loop_prevention", "max_iterations"`).

### §B.3 Test requirements

- **REQ-ALC-009** (Ubiquitous) — The loader change **shall** include table-driven tests covering: (a) explicit custom value parses correctly; (b) absent block falls back to default `10`; (c) absent sub-key with present parent block falls back to default `10`; (d) the distinctness invariant (§A.4) holds under all input combinations.

- **REQ-ALC-010** (Ubiquitous) — The test suite **shall** include a dedicated anti-aliasing test that parses a YAML fixture setting both keys to distinct custom values (e.g. `agentic_loop.max_iterations: 7` and `loop_prevention.max_iterations: 99`) and asserts both parse to the correct separate Go fields with no cross-contamination.

### §B.4 Template parity

- **REQ-ALC-011** (Unwanted) — The implementation **shall not modify** `internal/template/templates/.moai/config/sections/workflow.yaml`. The template already ships the `agentic_loop:` block with `max_iterations: 10` (template workflow.yaml lines 11-12, with a preceding distinctness-invariant comment at lines 7-10 — re-verified 2026-07-08). Template-First Rule (CLAUDE.local.md §2) applies: this SPEC is Go-only. The template's literal default `10` is the user-facing SSOT that the Go const `DefaultAgenticLoopMaxIterations = 10` mirrors (REQ-ALC-003); any future default-value change MUST update template YAML and Go const atomically (a parity test would enforce this; see §E out-of-scope).

## §C. Constraints

- **CLAUDE.local.md §14 (HARD)** — No hardcoding. The default `10` is a single named const in `defaults.go`. The literal `10` is forbidden in any other `*.go` file. Env-var names (if any env override is added — none required by REQ-MAL-023) MUST be constants in `internal/config/envkeys.go`.
- **verification-claim-integrity §1.1 surface 3** — Defect/debt claims in this SPEC are verified by mechanical grep (`agentic_loop` Go-absence verified pre-authoring; see plan.md §B). No text-only inference of a defect.
- **Loader non-strict semantics** — Per `internal/config/CLAUDE.md`, the loader uses `yaml.v3` in non-strict mode; unknown keys are silently ignored. The new `agentic_loop:` parsing MUST follow the same policy (REQ-ALC-004 restates this).
- **Template-First Rule (CLAUDE.local.md §2 HARD)** — The template `workflow.yaml` already ships the `agentic_loop:` block (verified). This SPEC does NOT change template YAML; it only adds Go-side parsing of an already-shipped user-facing key. If a future SPEC changes the default value, the template YAML and the Go const MUST be updated atomically (a parity test would enforce this; see §E out-of-scope).
- **Era integrity** — Parent SPEC `SPEC-MOAI-AGENTIC-LOOP-001` is `status: completed`, 3-phase closed (push `9be283c32`). This SPEC MUST NOT modify the parent. The parent's `era: V3R6` is inherited.

## §D. Success Criteria

1. `go build ./internal/config/... ./internal/settings/...` succeeds with zero new warnings.
2. `go test ./internal/config/... ./internal/settings/...` passes including the new tests (REQ-ALC-009, REQ-ALC-010).
3. `grep -rn 'agentic_loop\|AgenticLoop' internal/ | grep -v _test.go` returns at least the new types.go + defaults.go + loader.go + schema_sections.go sites (i.e. the key is now mechanically visible).
4. AC-DISTINCT-001 (acceptance.md) passes — the distinctness invariant is mechanically enforced by a dedicated test.
5. Coverage for the modified package(s) is equal to or higher than before the change (go test -cover).

## §E. Exclusions (Out of Scope)

The following are explicitly out of scope for this SPEC. Reports/analysis of existing behavior belong in `.moai/reports/`; the items below are deferred or rejected scope.

### Out of Scope — Runtime enforcement of the iteration ceiling

- This SPEC registers the Go field and parses the key. It does NOT add runtime code that **consumes** `AgenticLoop.MaxIterations` to halt a running pipeline at the ceiling. Runtime enforcement (e.g. a counter in the orchestrator's pipeline loop, a Stop-hook iteration guard) is owned by a follow-up runtime SPEC. Parent `SPEC-MOAI-AGENTIC-LOOP-001` REQ-MAL-023..027 continue to prose-enforce the ceiling at the orchestrator layer.
- A Stop-hook iteration-counter guard was explicitly named as a follow-up SPEC in the parent's §E "Out of Scope — Go code changes"; this SPEC does NOT deliver it.

### Out of Scope — Env-var override

- REQ-MAL-023 does not require an environment-variable override (e.g. `MOAI_AGENTIC_LOOP_MAX_ITERATIONS`). Adding one is a separate concern; if added later, the env-var name MUST be a constant in `internal/config/envkeys.go` per CLAUDE.local.md §14. This SPEC adds no env-var.

### Out of Scope — Template YAML change

- The template `internal/template/templates/.moai/config/sections/workflow.yaml` already ships the `agentic_loop:` block (lines 7-11, verified). This SPEC does NOT modify the template. A parity test asserting template `10` == Go const `10` is RECOMMENDED but NOT required by this SPEC (see "Out of Scope — Parity CI" below).

### Out of Scope — Parity CI

- A CI guard asserting `template workflow.yaml agentic_loop.max_iterations == DefaultAgenticLoopMaxIterations const` is valuable but out of scope. The symmetry audit tests `TestStructYAMLSymmetry_*` and `TestAuditLoaderCompleteness` (per `internal/config/CLAUDE.md` CI Guards) MAY surface related drift but are not required to be extended by this SPEC. Run-phase MAY add a parity test if the implementer judges it warranted; plan.md does not mandate it.

### Out of Scope — Refactoring the existing MaxIterations fields

- `internal/config/types.go` has `MaxIterations int \`yaml:"max_iterations"\`` at 4 sites (lines 295, 378, 768-769, 1034) belonging to loop_prevention / ralph / design / harness structs. This SPEC does NOT refactor, rename, or deduplicate those fields. The new `AgenticLoopConfig.MaxIterations` is a fifth, structurally distinct field.

### Out of Scope — Agent body / skill body / CLAUDE.md changes

- No `.claude/agents/**/*.md`, `.claude/skills/**/*.md`, or `CLAUDE.md` changes. The prose surface (`/moai` skill body, parent SPEC orchestration prose) is owned by `SPEC-MOAI-AGENTIC-LOOP-001` (completed).
