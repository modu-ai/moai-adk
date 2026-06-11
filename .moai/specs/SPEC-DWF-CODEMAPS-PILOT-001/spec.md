---
id: SPEC-DWF-CODEMAPS-PILOT-001
title: "Dynamic-Workflow Pilot: Read-Only Per-Package Codemaps Extraction Fan-Out"
version: "0.1.0"
status: implemented
created: 2026-06-05
updated: 2026-06-11
author: GOOS
priority: P2
phase: "v3.0.0"
module: ".claude/workflows"
lifecycle: exploratory
tier: S
tags: "dynamic-workflow, codemaps, pilot, fan-out, falsification"
---

# SPEC-DWF-CODEMAPS-PILOT-001 — Dynamic-Workflow Pilot: Read-Only Per-Package Codemaps Extraction Fan-Out

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-05 | GOOS | Initial draft — plan-phase artifacts (Tier S, GEARS) |

## A. Context and Rationale

A 16-agent dynamic-workflow adoption analysis (ground-truth + adversarial verification) evaluated 9 candidates for applying Claude Code's dynamic-workflow primitive to moai-adk-go. The result was 0 RECOMMEND, 2 CONDITIONAL, 7 REJECT. The user selected the highest-scoring CONDITIONAL candidate (DWF-05, fit 0.52) as a **low-risk pilot to learn the dynamic-workflow primitive mechanics** — the script-orchestrates-many-agents primitive (`.claude/workflows/*.js` using `agent()` / `parallel()` / `pipeline()` coordination concepts) described in `.claude/rules/moai/workflow/dynamic-workflows.md`.

The pilot orchestrates ONLY the **read-only per-package dependency-graph / public-surface extraction** half of `/moai codemaps` as a fan-out: one read-only agent per Go package, aggregating in script variables, returning structured graphs to the orchestrator.

**Verified scale** (do not re-derive): `go list ./... | wc -l` = 97 real Go packages; ~27,800 Go source files in the tree. This is a borderline "dozens-to-hundreds" fan-out — workflow-eligible per the dynamic-workflows.md routing heuristic, but it MUST be justified over the deterministic alternative (see REQ-DCP-003, the falsification test).

The pilot's value is the **learning plus the falsification verdict** — whichever way the verdict lands. The pilot is NOT committed to shipping a reusable workflow; shipping is contingent on the falsification test passing.

### A.1 What this SPEC is and is not

- This IS a SPEC (a forward-looking pilot that produces a persisted artifact + a documented verdict), so it correctly lives in `.moai/specs/`.
- This is NOT a code-implementation SPEC for production Go. No `internal/` or `pkg/` Go source changes are in scope.
- The existing `/moai codemaps` pipeline (`.claude/skills/moai/workflows/codemaps.md`) is NOT modified by this SPEC.

## B. Scope

### B.1 In scope

1. A pilot dynamic-workflow that fans out one read-only agent per Go package to extract that package's dependency graph + public surface (exported types / functions / interfaces) and any per-package architectural observation an LLM can add that `go list -deps -json` cannot.
2. A falsification test comparing per-package LLM synthesis output against the deterministic baseline (`go list -deps -json ./...` + `go doc`).
3. A documented verdict (value proven OR value not proven) and a how-to note capturing the primitive mechanics learned.
4. The persisted deliverable in its recommended shape (see § E Design Decision).

### B.2 Out of scope — see § Exclusions

The following are explicitly excluded (full enumeration in § Exclusions below):

- The 5-doc cohesive authoring (overview / modules / dependencies / entry-points / data-flow).
- The Phase 5 `AskUserQuestion` next-steps round.
- Any production Go change under `internal/` or `pkg/`.
- Any template-suite landing under `internal/template/templates/`.
- Modification of the existing `/moai codemaps` pipeline (`.claude/skills/moai/workflows/codemaps.md`).

## C. GEARS Requirements

> GEARS notation (current). `<subject>` is generalized to any noun. MUST-PASS requirements are flagged inline.

### C.1 Extraction / authoring split (AskUserQuestion boundary)

- **REQ-DCP-001** (Ubiquitous): The pilot workflow **shall** cover ONLY the read-only per-package extraction stage — dependency-graph + public-surface extraction per Go package — and **shall not** include the 5-doc cohesive authoring stage (`codemaps.md` Phase 2-3) nor any next-steps interaction.
- **REQ-DCP-002** (Ubiquitous): The 5-doc cohesive authoring (overview / modules / dependencies / entry-points / data-flow) and the Phase 5 `AskUserQuestion` next-steps round **shall** remain OUTSIDE the workflow, owned by the orchestrator and manager-docs, because workflow agents cannot prompt the user mid-run (asymmetric interaction boundary inherited from subagents). [MUST-PASS]
  - Cross-reference: `.claude/rules/moai/workflow/dynamic-workflows.md` § MoAI Integration Notes + § When NOT to Use a Workflow; `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary.
- **REQ-DCP-007** (Where, capability gate): **Where** a workflow agent lacks an input it needs, the agent **shall** return a structured blocker report to the orchestrator rather than prompting the user.

### C.2 The `go list` value-justification (falsification test)

- **REQ-DCP-003** (Ubiquitous): The pilot **shall** demonstrate, with concrete side-by-side evidence on a sampled subset of packages, whether per-package LLM synthesis adds value BEYOND what `go list -deps -json ./...` + `go doc` already produce. [MUST-PASS]
- **REQ-DCP-004** (Where, capability gate): **Where** the falsification test shows per-package LLM synthesis adds material value over the deterministic baseline, the pilot **shall** record a "value proven" verdict and recommend shipping the validated pattern. [MUST-PASS — PASS state]
- **REQ-DCP-005** (Where, capability gate): **Where** the falsification test shows per-package LLM synthesis does NOT add material value over the deterministic baseline, the pilot **shall** record a documented "primitive not worth shipping for this op" negative verdict, retain the how-to learning note, and recommend the deterministic `go list -deps -json` path for this operation. [MUST-PASS — PASS state]
  - REQ-DCP-004 and REQ-DCP-005 are mutually exclusive and jointly exhaustive: exactly one fires per run, and BOTH are PASS states. A run is a FAIL only if NO verdict is recorded with evidence.

### C.3 Determinism (workflow-script constraint)

- **REQ-DCP-006** (Ubiquitous): The fan-out script body **shall not** read the wall clock nor draw random numbers; the package list **shall** be injected as `args`, and any timestamp **shall** be stamped onto the result AFTER the run returns. [MUST-PASS]
  - Cross-reference: `.claude/rules/moai/workflow/dynamic-workflows.md` § How a Workflow Runs (deterministic-script constraint — resume caching keys on deterministic outputs).
- **REQ-DCP-008** (When, event-driven): **When** the workflow is launched, the orchestrator **shall** have already passed GATE-2 and collected all preferences via `AskUserQuestion`, because the workflow takes no mid-run user input.

### C.4 Artifact placement (non-template)

- **REQ-DCP-009** (Ubiquitous): The pilot workflow script **shall** live in the local, user-owned `.claude/workflows/` directory and **shall not** be added to `internal/template/templates/` in any form. [MUST-PASS]
  - Rationale and source: `.claude/rules/moai/workflow/dynamic-workflows.md` states verbatim that "MoAI does not ship any saved workflows by default; the user-owned `.claude/workflows/` directory is not template-managed." Per CLAUDE.local.md § 24 (harness namespace) and § 25 (template internal-content isolation), shipping a workflow script in the template suite would violate the user-owned-artifact boundary and the template neutrality contract.
- **REQ-DCP-010** (When, event-driven): **When** any pilot artifact is created, the author **shall** verify no file under `internal/template/templates/` references or contains the pilot workflow script.

## D. Acceptance Criteria Summary

Full Given-When-Then enumeration is in `acceptance.md`. The MUST-PASS criteria are AC-DCP-002 (extraction/authoring split holds), AC-DCP-003 (falsification test executed with evidence), AC-DCP-004/005 (a verdict recorded — either outcome PASS), AC-DCP-006 (determinism), and AC-DCP-009 (non-template placement). A run that produces a defensible negative verdict is a PASS, not a failure.

## E. Design Decision — Deliverable Shape (recommendation)

The pilot deliverable shape was chosen among three options:

- **(a) Saved local `.claude/workflows/codemaps-extract.js` reusable script** — persists a runnable script. Higher commitment; only justified if the falsification test passes.
- **(b) Documented & validated pattern entry in `dynamic-workflows.md` § pattern catalog** — fills the strategy-doc gap (`.moai/docs/autonomous-workflow-strategy.md` § 6.5 "codebase sweep — mx / codemaps" is currently described but not built/validated). Low commitment, high reuse-of-learning.
- **(c) On-demand authoring only (no persisted artifact)** — produces only the verdict + a how-to note, no runnable artifact.

**Recommendation: option (b) as PRIMARY, with (a) conditional on a "value proven" verdict.**

Justification (lowest risk, highest learning, honors the non-template constraint):

- Option (b) captures the primitive mechanics learned as a durable, validated pattern entry regardless of which way the verdict lands — so the pilot retains value even on a negative verdict (REQ-DCP-005). The strategy doc's § 6.5 already sketches this pattern; the pilot's job is to validate or refute it and record the outcome.
- Option (a) is layered ON TOP of (b) ONLY when the falsification test passes (REQ-DCP-004). Saving a runnable `.claude/workflows/codemaps-extract.js` before the value is proven would over-commit; the script lands only after evidence justifies it.
- Both (a) and (b) honor REQ-DCP-009: the pattern-catalog entry lives in `dynamic-workflows.md` (an existing rule file, not a template), and the optional script lives in local `.claude/workflows/` — neither lands in `internal/template/templates/`.
- Option (c) is rejected as PRIMARY because it discards the reusable pattern learning; the how-to note alone is not discoverable at the rule layer.

## Exclusions (What NOT to Build)

[HARD] This SPEC explicitly does NOT build:

1. **The 5-doc cohesive authoring** (overview.md / modules.md / dependencies.md / entry-points.md / data-flow.md). Cohesive cross-linked authoring stays in the existing `/moai codemaps` Phase 2-3 (orchestrator + manager-docs owned). The workflow covers extraction only.
2. **The Phase 5 `AskUserQuestion` next-steps round.** Workflow agents cannot prompt the user mid-run; the interactive next-steps round stays outside the workflow, in the orchestrator turn after the run returns.
3. **Any production Go code change** in `internal/` or `pkg/`. No new Go package, no edit to existing Go source. The pilot is workflow-script + docs only.
4. **Any template-suite artifact.** The pilot workflow script MUST NOT appear under `internal/template/templates/`. `.claude/workflows/` is user-owned and not template-managed.
5. **Modification of the existing `/moai codemaps` pipeline** (`.claude/skills/moai/workflows/codemaps.md`). The pilot is additive and parallel; it does not rewire the shipped pipeline.
6. **A general-purpose "always use a workflow for codemaps" mandate.** The pilot informs a verdict; it does not pre-commit MoAI to the workflow path for codemaps.

## F. Cross-References

- `.claude/rules/moai/workflow/dynamic-workflows.md` — primitive semantics, deterministic-script constraint, non-template `.claude/workflows/` statement, MoAI integration notes.
- `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary — workflow/subagent cannot prompt user.
- `.claude/skills/moai/workflows/codemaps.md` — existing pipeline (Phase 1 Explore extraction; Phase 2-3 manager-docs cohesive authoring; Phase 5 AskUserQuestion).
- `.moai/docs/autonomous-workflow-strategy.md` § 6.5 / § 6.6 — codebase-sweep parallel-barrier pattern + pattern-selection guide.
- CLAUDE.local.md § 24 (harness namespace) + § 25 (template internal-content isolation) — user-owned vs template-managed boundary.
