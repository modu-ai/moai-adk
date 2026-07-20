---
description: "Project Phase 15/7 — project-specific harness generation entry point (redirects to the v4 harness Builder) and 5-Layer Activation wiring"
user-invocable: false
metadata:
  parent: moai-workflow-project
  phase: "Phase 15: Harness Generation Entry (v4 Builder redirect); Phase 16: 5-Layer Activation"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

## Phase 15: Project-Specific Harness Generation (v4 Builder Redirect)

Purpose: when the completion flow (Phase 14) offers project-specific harness generation, this phase hands off to the current v4 harness Builder instead of running a standalone interview here. The former 16-question / 4-round static Socratic interview and its direct invocation of the deprecated `moai-meta-harness` skill (status: deprecated, superseded by the v4 Builder) are retired. The v4 Builder's own Context-First Discovery (domain / goal / constraints / scope extraction) replaces the fixed question set with an adaptive interview scoped to the actual request, so this phase no longer duplicates that logic.

[HARD] Do NOT invoke the deprecated `moai-meta-harness` skill or reproduce its static question set here. Route to the v4 flow instead.

### 5.0 Promoted Offer — Post-Project-Type-Confirmation Harness Proposal

Purpose: surface the harness-generation offer WHERE the user is already engaged — immediately after the adaptive project interview confirms the project type — instead of leaving it buried as one option in the tail-end Phase 14 next-step menu. This **promoted offer** is a **post-project-type-confirmation** proposal: once the interview has confirmed the project type, the orchestrator surfaces the harness-generation proposal

> "이 프로젝트에 `<type>` 개발 하네스를 생성할까요?"
> ("Generate a `<type>` development harness for this project?")

via `AskUserQuestion`, BEFORE the completion flow reaches the Phase 14 next-step menu. This is a live flow step: it fires as the project-type confirmation lands, not as a documented aside.

`<type>` resolves from the `.moai/project/harness-spec.yaml` `domain` field (written by `project/doc-generation.md` Phase 8) when present; otherwise it falls back to the mode-detection confirmed project type. State both sources so the prompt always renders a concrete `<type>`.

[HARD] The Phase 14 next-step "Generate harness" menu option — hosted in `project/doc-generation.md` (read-only for this promotion, retained by non-modification) — is RETAINED as a fallback. Both entry points remain reachable: a user who declines this post-project-type-confirmation proposal still reaches the Phase 14 menu later. Promoting the offer does NOT remove the fallback.

On acceptance, proceed to § 5.1 (Entry — hand off to the v4 Builder entry workflow). On decline, continue the completion flow with the Phase 14 fallback still available.

### 5.1 Entry — Hand off to the v4 Builder entry workflow

Compose a natural-language harness-creation request from `.moai/project/harness-spec.yaml` (the machine-readable interview output written by `project/doc-generation.md` Phase 8) together with the completed project documentation (`product.md` / `structure.md` / `tech.md`) and the user's stated intent from Phase 14, then hand off to `${CLAUDE_SKILL_DIR}/workflows/harness-build-entry.md` (the same entry point the `/moai:harness <request>` invocation uses). Carrying `harness-spec.yaml` forward means the interview data is no longer discarded — its recorded `domain` / `goal` / `constraints` / `scope` (plus the extended `verification` / `external_systems` / `ui_surface` / `team_sharing` axes) reach harness generation as pre-satisfied context. That workflow runs its own Context-First Discovery, harness `<name>` derivation, and orchestrator-issued approval gate — none of which are duplicated here.

### 5.2 Generation — Orchestrator-direct Builder

On approval, the entry workflow transitions directly into the orchestrator-direct Builder (`${CLAUDE_SKILL_DIR}/workflows/harness-builder.md`), which runs the 4 signal-driven phases (ANALYZE / PLAN / GENERATE / ACTIVATE) and emits the 5 canonical artifact types (thin-wrapper entry command, Runner Workflow, specialist sub-agent definitions, companion Progressive-Disclosure skills, `manifest.json`).

[HARD] The GENERATE phase MUST run the FROZEN guard (`EnsureAllowed`) as the **first check**
before any write attempt. Paths in `.claude/agents/moai/`, `.claude/skills/moai-*/`,
`.claude/skills/moai/`, or `.claude/rules/moai/` are permanently FROZEN and must be
rejected immediately.

[HARD] If the Builder's generation fails mid-way, its own cleanup handling removes all
partial artifacts written so far — see `harness-builder.md` for the full failure-handling contract.

### 5.3 Post-Build

After the Builder's ACTIVATE phase completes (including its reference-integrity smoke gate), proceed to Phase 16 (5-Layer Activation) below to wire the auto-trigger chain for the newly generated harness.

## Phase 16: 5-Layer Activation

Purpose: a generated harness only auto-triggers when its activation chain is
installed. Phase 15 emits the agents, skills, and `.moai/harness/` files; Phase 7
**wires the auto-trigger chain** so the generated harness actually loads when the
user works. Without Phase 16 the generated artifacts are silent waste — the
interview answers are captured and the domain agents emitted, yet nothing loads.

[HARD] Phase 16 runs ONLY after Phase 15 generation completes successfully (all
generated agents + skills + `.moai/harness/` files written). If generation failed
or ran its cleanup handling, Phase 16 MUST NOT run.

### 7.1 The Five Activation Layers

The generated harness activates through five layers. Phase 15 satisfies L1, L2,
L4, L5; Phase 16 installs the L3 marker and ensures the L5 `main.md` entry point
exists, then verifies all five with the smoke gate (7.3):

| Layer | Mechanism | Owner |
|-------|-----------|-------|
| L1 | `hns-*` skill frontmatter triggers (paths / keywords / agents / phases) | Phase 15 (generation) |
| L2 | `.moai/config/sections/workflow.yaml` `harness:` section | Phase 15 (generation) |
| L3 | `CLAUDE.md` `<!-- moai:harness-start -->` ~ `<!-- moai:harness-end -->` marker block | **Phase 16 (install)** |
| L4 | `.claude/skills/moai/workflows/{plan,run,sync,design}.md` static `@.moai/harness/` import line | Phase 15 (already present in workflow files) |
| L5 | `.moai/harness/main.md` task-shape router (the CLAUDE.md @import entry point) | **Phase 16 (install ensures present)** |

### 7.2 Install Invocation (orchestrator instruction)

The orchestrator runs the harness activation wiring by invoking the
`moai harness install` CLI surface with the generating SPEC ID and the project
domain. The command (a) scaffolds `.moai/harness/` so `main.md` exists (L5
entry point), and (b) injects the CLAUDE.md routing marker block (L3). It is
idempotent — re-running replaces the existing block rather than appending a
duplicate.

```bash
moai harness install --spec-id <SPEC-PROJ-INIT-NNN> --domain <domain>
# add --design-extension when the request calls for a full-custom design extension
```

The command takes positional flag inputs and never invokes `AskUserQuestion`
(subagent boundary). On a CLAUDE.md write failure (file absent / read-only) it
returns a structured error and does NOT report success — surface that error to
the user.

### 7.3 Phase 15 Post-Generation Smoke Gate

After the install runs, the orchestrator runs the post-generation smoke gate by
invoking the extended `doctor harness` 5-layer diagnosis:

```bash
moai doctor harness   # (or: moai doctor, which includes the Harness 5-Layer check)
```

The gate FAILs (non-OK status) when a generated harness is structurally
incomplete — covering:

- `.moai/harness/main.md` absent (L5 entry point missing).
- CLAUDE.md does not contain exactly one paired
  `<!-- moai:harness-start -->` / `<!-- moai:harness-end -->` block (L3 marker).
- a generated `.claude/agents/harness/*.md` agent has an empty `description`.
- a generated agent's `skills:` preload references a `hns-*` (or legacy
  `harness-*`) skill directory that does not exist on disk (dangling skill
  reference).
- a generated agent OMITS the `skills:` frontmatter key entirely (the runtime
  enforcement of the `skills:` preload emission contract — a `skills:`-less agent
  would otherwise pass silently and reproduce the auto-discovery failure mode
  this phase exists to close).

When the gate FAILs, surface the failing layers to the user — a structurally
incomplete harness must be regenerated or repaired before it can auto-trigger.

### 7.4 Retrofit Note (existing incomplete harnesses)

A harness generated **before** this activation wiring existed has its agents and
skills but lacks the CLAUDE.md marker (L3) and may lack `main.md` (L5), so it
never auto-triggers. To retrofit such a project, re-run the harness generation
flow (re-run `/moai project` Phase 15-7) OR run `moai harness install --spec-id
<SPEC-ID> --domain <domain>` directly against the project root. The install is
idempotent, so running it on an already-wired harness is safe.
