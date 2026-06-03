---
id: SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001
title: "Harness Activation Wiring — 생성된 하네스 자동 트리거 배선 복원"
version: "0.2.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai/workflows/project/, .claude/skills/moai-meta-harness/, internal/cli/doctor_harness.go, internal/harness/"
lifecycle: spec-anchored
tags: "harness, activation, wiring, project-init, smoke-gate, marker-install"
era: V3R6
---

# SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001 — Harness Activation Wiring

## A. Overview

MoAI's meta-harness (`moai-meta-harness` skill + `project/meta-harness.md` workflow) generates a
project-specific agent team during `/moai project` Phase 5-6, emitting `.claude/agents/harness/*.md`
agents, companion skills, and `.moai/harness/` artifacts. The generation works — but the generated
harness **exists without ever auto-triggering**. A ground-truth diagnosis (recorded at
`.moai/docs/harness-delivery-strategy.md` §4 + §6.5) traced the failure to two systemic wiring breaks,
verified against a real generated harness and 4 sibling projects.

This SPEC restores the auto-trigger chain by wiring the **already-existing but orphaned** marker
installer (`internal/harness/layer3.go` `InjectMarker`) and harness-directory scaffolder
(`internal/harness/layer5.go` `ScaffoldHarnessDir`, which emits `main.md`) into the
post-generation flow, defining the `main.md` task-shape router structure, mandating `skills:`
frontmatter preload on generated agents, and adding a Phase-6 post-generation smoke gate that
fails when a generated harness is structurally incomplete.

### A.1 Ground-Truth Root Cause (verified, not hypothesized)

| Break | Evidence | Severity |
|-------|----------|----------|
| B1 — CLAUDE.md routing markers absent | `InjectMarker` (`internal/harness/layer3.go`) has **0 non-test callers**; this repo's CLAUDE.md and the template CLAUDE.md both have **0 `moai:harness-start` markers**; 5 sampled projects = 0 | SYSTEMIC, PRIMARY killer |
| B2 — `.moai/harness/main.md` entry point absent | `ScaffoldHarnessDir` (`internal/harness/layer5.go`, emits `main.md`) has **0 non-test callers**; the sampled MINK project has only `interview-results.md` + `design-extension.md` | Per-project (MINK), Phase-skeleton incomplete |
| B3 — generated-agent descriptions inadequate | **REFUTED** — the 9 sampled MINK harness agents all carry rich trigger text. Description quality is NOT the problem. | Not a defect |
| B4 — generated skills lack a guaranteed preload path | Generated agents do not declare `skills:` frontmatter, so the companion skill is not deterministically loaded when the agent is delegated | Contributing |

The auto-trigger chain dies at **Step 2 (marker precondition)** and **Step 3 (main.md entry)** per the
diagnosis's chain trace. Fixing B1 + B2 (+ B4 reinforcement) restores activation. The orphaned Go
functions are the dead-code installer the diagnosis names — they were nominally owned by
`SPEC-V3R3-PROJECT-HARNESS-001` (status completed) but its installer call path was never wired.

### A.2 Why This Matters

A generated harness that never auto-triggers is silent waste: the user invests a 16-question Socratic
interview, the meta-harness emits domain agents + skills, yet none of it loads when the user actually
works. The fix is high-leverage and low-risk because the mechanisms (`InjectMarker`,
`ScaffoldHarnessDir`, the `doctor harness` 5-layer check) already exist and are unit-tested — this
SPEC connects them, it does not invent them.

---

## B. Scope

This SPEC covers four wiring items derived from `.moai/docs/harness-delivery-strategy.md` §7 row 2
(Lane 1 + Lane 3 of the recommended Model C):

1. **Marker installation ownership** — wire the existing `InjectMarker` into the post-generation flow
   so the `<!-- moai:harness-start -->` / `<!-- moai:harness-end -->` CLAUDE.md routing block is
   installed when a harness is generated.
2. **`.moai/harness/main.md` generation as a task-shape router** — mandate that the post-generation flow
   emits `main.md` as a task-shape → specialist ROUTER manifest (the orchestrator entry point) and define
   its required structure.
3. **Generated-artifact self-activation** — generated agents MUST carry both trigger-shaped descriptions
   (already satisfied) AND `skills:` frontmatter preload referencing their companion harness skill.
4. **Phase-6 post-generation smoke gate** — a verification (Go test extension of the existing
   `doctor harness` 5-layer check) that fails when a generated harness is missing `main.md`, CLAUDE.md
   markers, non-empty agent descriptions, or contains dangling skill references.

### B.1 Existing Dependencies (referenced, NOT re-specified)

- **`moai update` protection already exists**: `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` (PR #1048,
  `internal/cli/update_namespace_protect.go` + `update.go isUserOwnedNamespace` + 14 tests) preserves and
  backs up `my-harness-*` skills, `.claude/agents/harness/`, and `.moai/harness/`. This SPEC does NOT
  re-specify update protection — it is an existing guarantee this SPEC relies on.
- **`InjectMarker` + `ScaffoldHarnessDir` already exist** (`internal/harness/layer3.go`, `layer5.go`) with
  isolation unit tests (`layer3_test.go`, `layer5_test.go`). This SPEC wires their call path; it does NOT
  rewrite the functions' core behavior.
- **`doctor harness` 5-layer check already exists** (`internal/cli/doctor_harness.go` `runHarnessCheck`)
  with L1-L5 status. This SPEC extends it with smoke-gate semantics, not a new diagnosis engine.

---

## C. Requirements (GEARS)

### C.1 Marker Installation (B1)

REQ-HAW-001 — **When** a harness generation flow completes successfully (all generated agents + skills +
`.moai/harness/` files written), the harness activation wiring **shall** install the CLAUDE.md routing
block by invoking the existing `InjectMarker` mechanism with the generating SPEC ID, the project domain,
and the harness import paths.

REQ-HAW-002 — The harness activation wiring **shall** produce exactly one paired
`<!-- moai:harness-start ... -->` / `<!-- moai:harness-end -->` block per CLAUDE.md (idempotent
re-install replaces the existing block rather than appending a duplicate).

REQ-HAW-003 — **Where** the marker-install mechanism is reached from a CLI or orchestrator path, the
mechanism **shall not** invoke `AskUserQuestion` or any `mcp__askuser__*` tool (subagent boundary
C-HRA-008 / `internal/cli/CLAUDE.md`).

REQ-HAW-004 — **When** the marker installer cannot write CLAUDE.md (file absent, permission error), the
harness activation wiring **shall** surface a structured error and **shall not** silently report success.

### C.2 main.md Router Generation (B2)

REQ-HAW-005 — **When** a harness generation flow runs, the harness activation wiring **shall** emit
`.moai/harness/main.md` as the orchestrator entry point (the existing `ScaffoldHarnessDir` already
produces this file; the wiring **shall** ensure that scaffolder is actually invoked in the flow).

REQ-HAW-006 — The generated `.moai/harness/main.md` **shall** be structured as a task-shape → specialist
ROUTER manifest containing: (a) a project domain summary, (b) a routing table mapping observable
task-shapes to the generated harness specialist agents, and (c) a Linked Files section enumerating the
companion extension files.

REQ-HAW-007 — **While** `main.md` is the documented activation entry point (per
`moai-meta-harness/SKILL.md` § Trigger Mechanics auto-load condition 1), the harness activation wiring
**shall** ensure `main.md` is present whenever any `.claude/agents/harness/*.md` agent has been
generated (no agents-without-entry-point skeleton).

### C.3 Generated-Artifact Self-Activation (B4)

REQ-HAW-008 — The harness generation flow **shall** emit each generated `.claude/agents/harness/*.md`
agent with a `skills:` frontmatter entry that preloads the agent's companion harness skill, so the skill
loads deterministically when the agent is delegated. **When** a harness generation flow completes, the
post-generation smoke gate (REQ-HAW-010..014) **shall** fail if any generated agent omits the `skills:`
frontmatter key — i.e., the emission contract is enforced at runtime by the gate, not assumed (closes the
auto-discovery failure mode where a `skills:`-less agent would otherwise pass silently). See AC-HAW-015.

REQ-HAW-009 — Each generated `.claude/agents/harness/*.md` agent **shall** carry a non-empty,
trigger-shaped `description` frontmatter field (this is already satisfied per B3 REFUTED; the requirement
codifies it so the smoke gate can assert it).

### C.4 Phase-6 Post-Generation Smoke Gate (B1+B2+B3+B4 catch)

REQ-HAW-010 — **When** a harness generation flow completes, the post-generation smoke gate **shall** fail
(non-OK status) if `.moai/harness/main.md` is absent.

REQ-HAW-011 — **When** a harness generation flow completes, the post-generation smoke gate **shall** fail
if the project CLAUDE.md does not contain exactly one paired
`<!-- moai:harness-start -->` / `<!-- moai:harness-end -->` block.

REQ-HAW-012 — **When** a harness generation flow completes, the post-generation smoke gate **shall** fail
if any generated `.claude/agents/harness/*.md` agent has an empty `description` frontmatter field.

REQ-HAW-013 — **When** a harness generation flow completes, the post-generation smoke gate **shall** fail
if any generated agent declares a `skills:` preload reference to a harness skill directory that does not
exist on disk (dangling skill reference).

REQ-HAW-013b — **When** a harness generation flow completes, the post-generation smoke gate **shall** fail
if any generated `.claude/agents/harness/*.md` agent OMITS the `skills:` frontmatter key entirely (the
runtime enforcement of REQ-HAW-008's emission contract — distinct from REQ-HAW-013, which catches a
`skills:` key that is present but points to a non-existent skill dir).

REQ-HAW-014 — The post-generation smoke gate **shall** extend the existing `doctor harness` 5-layer
diagnosis (`internal/cli/doctor_harness.go`) rather than introduce a parallel diagnosis engine, and
**shall** preserve the existing L1-L5 status semantics.

### C.5 Process Constraints (Template-First + Prefix Stability)

REQ-HAW-015 — **Where** this SPEC modifies any `.claude/**` template asset (the meta-harness skill body,
the `project/meta-harness.md` workflow, or generated-agent emission templates), the change **shall** be
made in `internal/template/templates/.claude/...` first, then `make build`, then synced — per
CLAUDE.local.md §2 Template-First Rule.

REQ-HAW-016 — The harness generation flow **shall** continue to emit user-generated skills under the
existing protected `my-harness-*` prefix and **shall not** rename to or emit under a `harness-*` prefix
(the `my-harness-*` → `harness-*` namespace migration is a separate, non-blocking concern — see
Exclusions).

---

## D. Acceptance Criteria Reference

Detailed, individually verifiable acceptance criteria (AC-HAW-001 .. AC-HAW-014) live in
`acceptance.md`. Each AC binds to one or more REQ above and is grep- or test-verifiable.

---

## E. Exclusions (What NOT to Build)

This section is mandatory. The following are explicitly OUT OF SCOPE for this SPEC:

- **EX-1 — `my-harness-*` → `harness-*` prefix migration**: The namespace rename is a SEPARATE,
  NON-BLOCKING concern owned by a future `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` **(planned — not yet
  created; do not treat as an existing SPEC)** (doctrine-code split per commit `66a3d53be` "Phase 1
  doctrine-only"; `meta-harness SKILL.md:168` confirms `harness-*` generation has no protection yet).
  Generation MUST stay on the protected `my-harness-*` prefix. This SPEC specifies NO prefix rename.
- **EX-2 — `moai update` namespace protection**: Already implemented by
  `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` (PR #1048). NOT re-specified here; referenced as an existing
  dependency only.
- **EX-3 — Modifications to external user projects (MINK et al.)**: Existing generated harnesses live in
  USER projects, not this repo. This SPEC documents a retrofit path (re-run the harness generation flow,
  or a migration note) but scopes NO changes to those external projects.
- **EX-4 — Dynamic Workflow Lane 2 (`builder-harness` artifact_type=workflow, `.claude/workflows/`
  namespace)**: Deferred per `.moai/docs/harness-delivery-strategy.md` §7 row 4 (maximum risk / minimum
  payoff). Not in this SPEC.
- **EX-5 — Harness learning/evolution loop (Phase 7 iteration)**: Owned by
  `SPEC-V3R3-HARNESS-LEARNING-001`. Not in this SPEC.
- **EX-6 — Rewriting `InjectMarker` / `ScaffoldHarnessDir` core logic**: This SPEC wires the call path and
  may adjust `main.md` body structure (REQ-HAW-006), but does NOT rewrite the marker-injection or
  scaffolding algorithms.
- **EX-7 — Editing `.claude/rules/moai/core/askuser-protocol.md`**: A parallel session is editing this
  file; this SPEC touches it not at all.
- **EX-8 — Modifying any other SPEC's `spec.md` / `plan.md` / `acceptance.md` body**.

---

## F. HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial plan-phase authoring. Tier M. Wires the orphaned `InjectMarker` + `ScaffoldHarnessDir` installers, defines `main.md` router structure, mandates `skills:` preload on generated agents, adds Phase-6 smoke gate extending `doctor harness`. Grounded in `.moai/docs/harness-delivery-strategy.md` §4 + §6.5 ground-truth diagnosis. |

---

## G. Cross-References

- Diagnosis: `.moai/docs/harness-delivery-strategy.md` §4 (broken wiring chain), §6 (Model C recommendation), §6.5 (ground-truth sequencing correction), §7 (SPEC candidate table row 2)
- Orphaned installers: `internal/harness/layer3.go` (`InjectMarker`), `internal/harness/layer5.go` (`ScaffoldHarnessDir`)
- Smoke gate home: `internal/cli/doctor_harness.go` (`runHarnessCheck`, L1-L5), `internal/cli/doctor_harness_test.go`
- Generation flow: `.claude/skills/moai/workflows/project/meta-harness.md` (Phase 6 invocation; Phase 7 5-Layer Activation body absent — this SPEC adds it), `.claude/skills/moai-meta-harness/SKILL.md` (§ Trigger Mechanics, § Namespace Separation)
- Namespace contract: `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy
- Existing protection dependency: `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001`
- Template-first: CLAUDE.local.md §2
