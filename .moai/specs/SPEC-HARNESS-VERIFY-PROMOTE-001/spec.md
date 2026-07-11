---
id: SPEC-HARNESS-VERIFY-PROMOTE-001
title: "Harness-generation offer promotion + mandatory verify skill + specialist-agent template rules"
version: "0.1.1"
status: in-progress
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai/workflows, internal/template/templates/.claude/skills/moai/workflows"
lifecycle: spec-anchored
tags: "harness, verify-skill, specialist-agent, promote-offer, run-skill-generator, template-mirror"
era: V3R6
tier: S
depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]
---

# SPEC-HARNESS-VERIFY-PROMOTE-001 — Harness-generation offer promotion + mandatory verify skill + specialist-agent template rules

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-11 | 0.1.0 | Initial plan-phase draft (Tier S, 9 REQ / 11 AC). Third SPEC of the 3-SPEC "Project-Harness Pipeline" Epic; `depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]` (the foundation SPEC that introduced `.moai/project/harness-spec.yaml` + the adaptive interview). Three changes: (1) PROMOTE the harness-generation offer from a buried Phase 4.2 menu option to the project interview's final question (retaining the Phase 4.2 menu as a fallback); (2) make `harness-builder.md` GENERATE ship a mandatory `harness-<name>-verify` companion skill (mirroring the official `/run-skill-generator` runnable-verification pattern) as a 6th artifact; (3) inject two short mandatory rule blocks (tool-priority decision tree + Skill-First execution) into every generated specialist agent, and state a 3-7-specialists PLAN guardrail. Doc-only (markdown); no Go code. All Template-First. 2 open clarifications tracked in plan.md. | manager-spec |
| 2026-07-11 | 0.1.1 | Plan-audit fix (D1-D6). D1: resolved both plan-phase clarifications — the promoted offer is a **post-project-type-confirmation proposal** in `meta-harness.md` (the Phase 5.1 handoff module, NOT the interview's final question); the verify skill is ALWAYS mandatory with a stub-if-no-recipe fallback; the clarification markers are struck from plan.md (resolutions recorded in progress.md §E.1). D2: Epic artifact order fixed — verify skill = artifact 6 (mandatory, THIS SPEC), MCP fragment = artifact 7 (optional, owned by `SPEC-HARNESS-MCP-PROVISION-001`, NOT `SPEC-PROJECT-HARNESS-BRIDGE-001`); removed the self-contradictory artifact-6 collision language. D3: rewrote REQ-HVP-001 + AC-HVP-002 — the Phase 4.2 "Generate harness" fallback menu lives in `doc-generation.md`, not `meta-harness.md`; AC-HVP-002 now asserts the meta-harness.md addition via a discriminating token. D4: added AC-HVP-012 measuring the injected specialist blocks stay bounded (≤8 lines / verbatim §D.2 shape). D5/D6: informal short-form Epic-SPEC refs replaced with formal IDs; AC-HVP-003 gains a baseline-0 note (vacuous-preserve guard). Now 9 REQ / 12 AC. | manager-spec |

## §A. Context and Intent

The `/moai project` flow gathers project intent through an adaptive interview
(introduced by `SPEC-PROJECT-HARNESS-BRIDGE-001`) and can generate a
project-specific harness. Three weaknesses reduce the quality and discoverability
of that generated harness:

1. **The harness-generation offer is buried.** Today the offer surfaces only at
   the project workflow's Phase 4.2 — as ONE next-step menu option among ~5
   choices (DB-sync / Create SPEC / Review / Generate harness / Done), AFTER all
   docs are already generated. That Phase 4.2 next-step menu lives in
   `project/doc-generation.md`; `project/meta-harness.md` is the Phase 5.1 handoff
   module that redirects to `harness-build-entry.md` → `harness-builder.md`. A user
   who wants a harness has to reach the tail of the flow and pick it out of a list.
   Now that the foundation SPEC makes the interview confirm the project type up
   front, the harness-generation proposal SHOULD be surfaced as a
   **post-project-type-confirmation proposal** in `meta-harness.md` ("이 프로젝트에
   <type> 개발 하네스를 생성할까요?" — "Generate a <type> development harness for
   this project?"), fired after the adaptive interview confirms the project type,
   while the Phase 4.2 menu option (in `doc-generation.md`) is RETAINED as a
   fallback (both entry points reachable).

2. **Generated harnesses ship no runnable verification loop.** `harness-builder.md`
   GENERATE emits 5 artifact types (thin command / Runner JS / specialist agents /
   companion skills / manifest.json). It does NOT mandate any verification / run
   skill. Anthropic's strongest official theme for generated harnesses is "give
   Claude a check it can run"; the official `/run-skill-generator` bundled skill
   realizes exactly this — it discovers how to build / launch / test the app once
   from a clean environment and commits the recipe to `.claude/skills/run-<name>/`.
   Every generated harness SHOULD ship the same runnable verification loop.

3. **Generated specialist agents lack execution discipline, and harnesses
   over-generate agents.** Two reusable prompt patterns (surfaced from a reviewed
   claude.ai system-prompt leak, absorbing ONLY these two) improve every generated
   specialist: (a) an explicit **tool-priority decision tree** — "use a tool when
   it is the category fit, not a style preference" (category-fit MCP → search →
   file tools → inline response), and (b) a **Skill-First execution rule** — "read
   the relevant companion SKILL.md before any file / code work". Separately,
   Anthropic guidance is to generate FEW trigger-rich agents (3-7 max) because
   over-generation degrades Claude's automatic sub-agent delegation.

**Design premise (Anthropic-verified).** A generated harness should (i) be offered
where the user is already engaged (interview end, not a tail menu), (ii) ship a
runnable check, and (iii) contain few, trigger-rich, discipline-bearing specialists.
This SPEC realizes those three for the project→harness pipeline.

**Boundary principle.** This is the THIRD (Tier S, doc-only) SPEC of the Epic. It
builds ON TOP of the confirmed project type + `harness-spec.yaml` contract that
`SPEC-PROJECT-HARNESS-BRIDGE-001` created; it does NOT re-implement the interview
or the artifact. It changes only the harness-generation surface
(`meta-harness.md` / `harness-build-entry.md` / `harness-builder.md`).

## §B. Scope Summary

**In scope**:
- PROMOTE the harness-generation offer to a post-project-type-confirmation harness
  proposal in `project/meta-harness.md` (the Phase 5.1 handoff module), and RETAIN
  the Phase 4.2 next-step menu option — hosted in `project/doc-generation.md` — as a
  fallback (both entry points reachable).
- Surface the same harness-generation proposal as the harness-entry interview's
  final-round offer in `harness-build-entry.md`.
- Make `harness-builder.md` GENERATE mandate a `harness-<name>-verify` companion
  skill (mirroring `/run-skill-generator`) as a 6th artifact of every generated
  harness set.
- Inject two short mandatory rule blocks — a tool-priority decision tree and a
  Skill-First execution rule — into every generated specialist agent body via the
  `harness-builder.md` specialist-generation template.
- State a 3-7-specialists-maximum guardrail in the `harness-builder.md` PLAN phase,
  requiring each specialist to be justified as a recurring same-instruction worker
  (else emit a skill instead).
- Mirror all edits into `internal/template/templates/...` (Template-First),
  `make build`, keep the neutrality CI guard green.

**Preserve**:
- The `harness-*` namespace-only invariant + the FROZEN guard rejecting writes to
  `.claude/agents/moai/`, `.claude/skills/moai-*/`, `.claude/rules/moai/`.
- The `/moai project` NO-SPEC scope guard (project flow never writes to
  `.moai/specs/**`).
- The `builder-harness` specialist-generation internals (how specialists are
  authored from the composed request) — unchanged except for the two injected
  rule blocks + the guardrail statement.
- The existing 5 generated artifact types (the verify skill is additive, a 6th).

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

### C.1 Promote the harness-generation offer

- **REQ-HVP-001** (Event-driven, with preservation clause): When the adaptive
  project interview (`SPEC-PROJECT-HARNESS-BRIDGE-001`) confirms the project type,
  the workflow shall surface the harness-generation proposal — "이 프로젝트에
  <type> 개발 하네스를 생성할까요?" — as a **post-project-type-confirmation harness
  proposal** in `project/meta-harness.md` (the Phase 5.1 handoff module, NOT the
  interview itself); AND the Phase 4.2 next-step menu option ("Generate harness") —
  which lives in `project/doc-generation.md` — shall be RETAINED as a fallback entry
  point (both entry points reachable). `doc-generation.md` is read-only for this
  SPEC; the fallback is retained by non-modification.
- **REQ-HVP-002** (Event-driven): When `harness-build-entry.md` runs its interview,
  the workflow shall surface the same harness-generation proposal as the
  interview's **final-round offer**.

### C.2 Mandatory verify companion skill (6th artifact)

- **REQ-HVP-003** (Ubiquitous): Every generated harness shall include a run / verify
  companion skill named `harness-<name>-verify` (under the `harness-*` namespace)
  that discovers and codifies the project's build / launch / test recipe from a
  clean environment — mirroring the official `/run-skill-generator` pattern. The
  verify skill is ALWAYS mandatory for every generated harness; when no build /
  launch / test recipe is discoverable, it is STILL emitted as a documented stub
  ("no recipe found") rather than omitted. This verify skill is the mandatory
  **artifact 6** of the generated harness set. The optional MCP fragment — the
  **artifact 7**, owned by `SPEC-HARNESS-MCP-PROVISION-001` — is a separate, distinct
  artifact (6 = mandatory verify skill; 7 = optional MCP fragment). The MCP fragment
  is NOT owned by `SPEC-PROJECT-HARNESS-BRIDGE-001`, whose §D deliverable is the
  `harness-spec.yaml` schema, not an MCP fragment.

### C.3 Specialist-agent template rule blocks

- **REQ-HVP-004** (Ubiquitous): The `harness-builder.md` specialist-agent
  generation template shall inject a short mandatory **tool-priority decision-tree**
  rule block into every generated specialist agent body — category-fit MCP →
  search → file tools → inline response ("use a tool when it is the category fit,
  not a style preference").
- **REQ-HVP-005** (Ubiquitous): The `harness-builder.md` specialist-agent
  generation template shall inject a short mandatory **Skill-First execution** rule
  block into every generated specialist agent body — "read the relevant companion
  SKILL.md before any file / code work". Each injected block shall stay short (a
  few lines) so generated agents are not bloated.

### C.4 Agent-count guardrail

- **REQ-HVP-006** (Ubiquitous): The `harness-builder.md` PLAN phase shall state a
  **3-7-specialists-maximum** guardrail; PLAN shall justify each specialist as a
  recurring same-instruction worker, else emit a companion skill instead of an
  agent (over-generation degrades Claude's automatic sub-agent delegation).

### C.5 Invariants (preservation)

- **REQ-HVP-007** (Unwanted behavior): The generated artifacts shall use the
  `harness-*` namespace prefix only; the harness-generation flow shall not write to
  `.claude/agents/moai/`, `.claude/skills/moai-*/`, or `.claude/rules/moai/` — the
  FROZEN guard rejection shall be preserved.
- **REQ-HVP-008** (Unwanted behavior): The `/moai project` and harness-generation
  flow shall not write to `.moai/specs/**`; the NO-SPEC scope guard shall be
  preserved.
- **REQ-HVP-009** (Ubiquitous): Every edit shall be made Template-First (in
  `internal/template/templates/...` FIRST, then mirrored byte-identically to the
  local `.claude/` copy, then compiled via `make build`); template neutrality (no
  internal SPEC IDs / dates / commit SHAs in `internal/template/templates/**`)
  shall be preserved.

## §D. Reference — generated-harness artifact set + specialist rule blocks (SSOT)

### D.1 Generated-harness artifact set (6 mandatory + 1 optional)

`harness-builder.md` GENERATE emits artifacts 1-6. Artifact 6 (verify skill) is added
by REQ-HVP-003; the first five are unchanged. Artifact 7 (optional MCP fragment) is
owned by a sibling Epic SPEC and is OUT OF SCOPE for THIS SPEC — it is listed only to
fix the Epic's canonical 6/7 order.

| # | Artifact | Namespace / path shape | Change |
|---|----------|------------------------|--------|
| 1 | Thin command | `.claude/commands/harness/<name>.md` | unchanged |
| 2 | Runner JS | `.claude/workflows/harness-<name>-run.js` | unchanged |
| 3 | Specialist agents (3-7) | `.claude/agents/harness/harness-<name>-*.md` | +2 injected rule blocks (D.2) |
| 4 | Companion skills | `.claude/skills/harness-<name>-*/SKILL.md` | unchanged |
| 5 | manifest.json | `.claude/commands/harness/<name>/manifest.json` | unchanged |
| 6 | **Verify skill (NEW — mandatory, THIS SPEC)** | `.claude/skills/harness-<name>-verify/SKILL.md` | mandatory; codifies build / launch / test recipe (`/run-skill-generator` pattern); stub-if-no-recipe |
| 7 | MCP fragment (optional) | (owned by `SPEC-HARNESS-MCP-PROVISION-001`) | OUT OF SCOPE for THIS SPEC — optional artifact 7, distinct from the mandatory verify skill (artifact 6) |

### D.2 Mandatory specialist-agent rule blocks (short — inject verbatim shape)

Every generated specialist agent body carries these two short blocks:

```markdown
## Tool Priority (category fit, not style preference)
1. Category-fit MCP tool — when the task IS the tool's category.
2. Search (Grep/Glob) — locate content/files.
3. File tools (Read/Edit/Write) — inspect/modify.
4. Inline response — when no tool is the category fit.

## Skill-First Execution
Before any file/code work, read the relevant companion SKILL.md.
```

These blocks are the absorbed reusable patterns; the remainder of the reviewed
claude.ai leak is consumer-app-specific and out of scope (§E).

The full machine-verifiable AC matrix (AC-HVP-001 … AC-HVP-012) lives in
`acceptance.md` (SSOT). Every REQ maps to at least one AC; preservation REQs
(REQ-HVP-007/008) map to a namespace-intact / NO-WRITE absence assertion; the
anti-bloat clause of REQ-HVP-005 maps to AC-HVP-012.

## §E. Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — SPEC-PROJECT-HARNESS-BRIDGE-001 / SPEC-HARNESS-MCP-PROVISION-001 Epic territory

- The adaptive clarity-scored interview, the extended interview axes, and the
  `.moai/project/harness-spec.yaml` artifact are owned by
  `SPEC-PROJECT-HARNESS-BRIDGE-001` (the foundation SPEC). This SPEC CONSUMES the
  confirmed project type + `harness-spec.yaml`; it does NOT re-author them. The
  `project/mode-detection.md` / `project/codebase-analysis.md` interview hosts are
  BRIDGE-001 scope and are NOT edited by this SPEC.
- The optional MCP fragment — **artifact 7**, owned by `SPEC-HARNESS-MCP-PROVISION-001`
  — is a distinct artifact from this SPEC's mandatory `harness-<name>-verify` skill
  (**artifact 6**). The Epic's canonical order is: artifact 6 = mandatory verify skill
  (THIS SPEC), artifact 7 = optional MCP fragment (`SPEC-HARNESS-MCP-PROVISION-001`).

### Out of Scope — builder-harness specialist-generation internals

- HOW `builder-harness` authors the actual specialist skills / agents from the
  composed request is unchanged except for the two injected rule blocks (§D.2) and
  the PLAN guardrail statement. This SPEC does NOT redesign the generation
  algorithm, the trigger-authoring logic, or the manifest schema.

### Out of Scope — Go code changes

- This SPEC is doc-only (markdown under `.claude/skills/...` and its template
  mirrors). No `internal/` / `pkg/` / `cmd/` Go source is modified. No Go parser
  for the verify skill or the specialist rule blocks is added.

### Out of Scope — the rest of the claude.ai system-prompt leak

- Only the two reusable patterns (tool-priority decision tree + Skill-First
  execution) are absorbed into generated specialist bodies. The remainder of the
  reviewed consumer-app system-prompt leak is consumer-specific and is NOT
  imported into any generated artifact or any MoAI file.

### Out of Scope — CHANGELOG / README / docs-site

- CHANGELOG.md is owned by manager-docs (sync-phase); README and docs-site
  4-locale updates for the promoted-offer / verify-skill behavior are a follow-up
  sync / docs concern.

## §F. Cross-References

- `.claude/skills/moai/workflows/project/meta-harness.md` — Phase 5.1 handoff /
  redirect host (promote-offer target; the post-project-type-confirmation proposal
  lands here).
- `.claude/skills/moai/workflows/project/doc-generation.md` — Phase 4.2 next-step
  menu host (the "Generate harness" fallback lives here; read-only for THIS SPEC —
  retained by non-modification).
- `.claude/skills/moai/workflows/harness-build-entry.md` — harness entry interview
  (final-round offer target).
- `.claude/skills/moai/workflows/harness-builder.md` — GENERATE contract (mandatory
  verify skill + specialist rule blocks + 3-7 PLAN guardrail).
- `SPEC-PROJECT-HARNESS-BRIDGE-001` — foundation SPEC (adaptive interview +
  `harness-spec.yaml` + confirmed project type this SPEC consumes). `depends_on`.
- Official `/run-skill-generator` bundled skill — the runnable-verification pattern
  the mandatory `harness-<name>-verify` skill mirrors.
- CLAUDE.local.md §24 (Harness Namespace) + §25 (Template Internal-Content
  Isolation) + §15 (16-language neutrality) — the namespace / neutrality / mirror
  discipline this SPEC's edits must respect.
- `plan.md` / `acceptance.md` — implementation plan + AC matrix (SSOT).
