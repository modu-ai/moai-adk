---
id: SPEC-GLM-EFFORT-TUNE-001
title: "GLM effort overlay configuration tune-up (P1/P2/P4) — design decisions"
version: "0.1.0"
status: draft
created: 2026-07-14
updated: 2026-07-14
author: manager-spec
priority: P2
phase: "v3.x config-tune"
module: "internal/template/glm_effort_overlay.go + .moai/config/sections/llm.yaml"
lifecycle: spec-anchored
tags: "glm, effort, overlay, config, template-mirror, reasoning-effort"
related_specs: [SPEC-MODEL-TIER-PLANTYPE-001]
---

# SPEC-GLM-EFFORT-TUNE-001 — design.md

This file records the two design decisions a reviewer is most likely to reverse. Per CLAUDE.md §7 Rule 1, these are placed first so review focuses here.

## §A. Decision P1 — exclude `builder-harness` from the coding-max override set (BOTH options recorded)

### §A.1 The decision

**Recommended: EXCLUDE `builder-harness` from `glmCodingMaxOverrideAgents`** (set becomes `{manager-develop}` only).

### §A.2 Option A (recommended) — EXCLUDE

**Consequence**: `builder-harness` falls under the standard collapse. Its Claude effort is `high` across all 3 tiers in both plan_type profiles (`model_policy.go:303,317`), so `CollapseClaudeEffortToGLM("high")` returns `reasoning-high` (NOT `reasoning-max`).

**Rationale**:
1. **Role mismatch with z.ai's recommendation**: z.ai's "reasoning_effort: max for coding tasks" guidance targets deep coding reasoning — the run-phase implementation work that `manager-develop` performs at Claude effort `xhigh`. `builder-harness` performs **artifact scaffolding** (agents/skills/plugins/hooks generation) — `agent-authoring.md:350` classifies its default effort as `high` with the explicit rationale "artifact scaffolding", NOT "coding/agentic". The two roles are not the same kind of "coding".
2. **Cost-reduction alignment**: CLAUDE.md §15 CG Mode states the goal as "60-70% cost reduction on implementation-heavy tasks". Forcing the harness agent to `reasoning-max` (the most expensive tier) for scaffolding work contradicts this goal — scaffolding is generation-volume + metadata-accuracy work, where `reasoning-high` is the economically appropriate tier.
3. **Override-set minimality**: the override is an EXCEPTION to the standard collapse. Minimality — the smallest set that captures z.ai's coding-task recommendation — is the safer default; adding members should require positive justification, not the other way around. `manager-develop` is the unambiguous primary beneficiary; `builder-harness` is a defensible-but-secondary beneficiary.

### §A.3 Option B (rejected) — INCLUDE (preserve the status quo)

**Consequence**: `builder-harness` stays in the override set; no P1 change.

**Rationale (the steel-manned case for INCLUDE)**:
1. **`builder-harness` DOES generate code**: its output includes dynamic specialist agent definitions, skill bodies, hook scripts — files that contain executable code (Go tests will run against them, shell scripts will execute). A "coding-task" classification is therefore defensible.
2. **Code generation correctness benefits from max reasoning**: a malformed generated agent file or a subtly wrong skill body is expensive to debug downstream; `reasoning-max` reduces that risk at the source.
3. **Simplicity (one less change)**: leaving the set as-is is a smaller diff and avoids the test rename.

### §A.4 Why Option A wins

Option B's rationales are real but secondary. The role-mismatch argument (A.2.1) is load-bearing: z.ai's "max for coding" recommendation is keyed to deep reasoning over an existing codebase, which is `manager-develop`'s profile; `builder-harness`'s scaffolding work is materially different. The cost-reduction argument (A.2.2) is the policy constraint that breaks the tie — the CG-mode goal explicitly targets cost reduction, and forcing the scaffolding agent to the most expensive tier works against it. Option B's "code-generation correctness" rationale is better served by the post-generation quality gates (`sync-auditor` 4-dimension scoring, `go test` against generated code) than by paying `reasoning-max` tax on every harness invocation.

A reviewer who rejects this tradeoff should flip to Option B — but the flip should be an explicit decision, not a silent preservation of the status quo.

## §B. Decision P2 — llm.yaml exposure mechanism (comments-only vs struct field + loader)

### §B.1 The decision

**Recommended: comments-only documentation block** for v0.1.0. Defer the real struct field + loader wiring to a follow-up if and when a user-override use case materializes.

### §B.2 Option A (recommended) — comments-only block

**Consequence**: a YAML comment block is added under the existing `glm:` key in both `llm.yaml` surfaces. No `LLMConfig` struct field is added; no loader code is touched.

**Rationale**:
1. **The goal is documentation transparency** (P2's actual finding: "the overlay is invisible in the config file a user reads"). Comments achieve this goal fully.
2. **No runtime behavior change** (constraint C-2). Comments cannot introduce a parallel runtime path; the Go overlay unambiguously remains the SSOT.
3. **No CI-guard binding**. Comments do not trigger `YAML_SECTION_NO_LOADER` (no new YAML key) or `CONFIG_STRUCT_YAML_MISMATCH` (no new struct field). The 5-step procedure for adding a new YAML section (settings-management.md § MoAI Configuration) does not apply.
4. **Forward-compatible**. If a later SPEC adds a real `glm.reasoning_effort_default` user-override field, the comments-only block can be promoted to a real key in a single, well-scoped change at that time — the comments do not preclude the struct-field path, they just do not require it now.
5. **Proportionality**. The documentation goal does not justify a struct + loader + CI-guard + audit-test change. YAGNI applies.

### §B.3 Option B (rejected for v0.1.0) — real `LLMConfig.GLMReasoning*` struct field + loader wiring

**Consequence**: a new struct field is added to `internal/config/types.go` `LLMConfig` (e.g. `GLMReasoningMapping string` or a richer struct), the loader (`loadLLMSection` in `internal/config/loader.go` or a sibling) is updated to populate it, the YAML key is added to both `llm.yaml` surfaces, and `audit_struct_yaml_symmetry_test.go` `symmetryCases` is extended.

**Rationale (the steel-manned case for Option B)**:
1. **Real user-override extension point**: a struct field lays the groundwork for users to override the overlay's collapse mapping from `llm.yaml` (e.g. "I want builder-harness at reasoning-max regardless of the default"). Comments cannot do this.
2. **Machine-readable parity with the Go overlay**: a populated struct field is testable in Go; comments are only testable by grep.
3. **Aligns with the "config is the user-facing surface" doctrine** — if the mapping is meaningful to users, it should be a real config key, not a comment.

### §B.4 Why Option A wins for v0.1.0

Option B's rationale (B.3.1) is about a FUTURE use case (user override of the overlay). The current SPEC's goal (P2) is visibility — "expose the mapping so a user can read it" — which Option A achieves. Option B is the right answer to a different question ("allow the user to override the overlay"), and that question is NOT in this SPEC's scope. Building the struct + loader + CI-guard apparatus now, preemptively, for a use case nobody has requested, violates the Simplicity decision ladder (moai-constitution.md § Agent Core Behaviors #4 — "Does this need to be built at all? YAGNI").

If a later SPEC introduces user-overridable reasoning-effort mappings, it will promote the comment block to a real key. That SPEC — not this one — owns the struct + loader + CI-guard work. This SPEC's exposure block is forward-compatible with that promotion (the comment block can be the seed for the real key's docstring).

### §B.5 Forward-compatibility note

If Option B is later adopted, the `llm.yaml` exposure block added by this SPEC becomes the docstring for the new key. No rewrite is needed — only a key-name promotion (e.g. wrapping the comment block's content under a real `reasoning_effort_mapping:` key). This is recorded so a future SPEC author does not need to re-derive the exposure from scratch.

## §C. Cross-cutting design notes

### §C.1 The 3-state framing (P4) is a correction, not a new model

The 3 reachable GLM reasoning states (`thinking-off` / `reasoning-high` / `reasoning-max`) are ALREADY named constants in `glm_effort_overlay.go:26-33`. P4 does not introduce a new model; it corrects the prose framing in comments and docs that drifted to "2-tier high/max". The code constants are the SSOT for the framing — comments and docs align to them, not the reverse.

### §C.2 The honesty caveat is non-negotiable

Per `verification-claim-integrity.md` §1.1 surface 3 + constraint C-3, the P2 exposure block's wording about wire-effectiveness MUST be "implemented + wired; live validation pending" — NEVER "validated", "guaranteed", or "works". This is the MODEL-TIER-PLANTYPE-001 honesty caveat carried forward. The exposure block makes the overlay VISIBLE; it does not make a fresh claim about the overlay's wire-effectiveness.

### §C.3 P3 is the next GLM-overlay SPEC

The live wire-effectiveness validation (P3) is owned by SPEC-MODEL-TIER-PLANTYPE-001 follow-up #1. This SPEC's progress.md §E.4 should record a pointer to that follow-up so the chain is not lost.
