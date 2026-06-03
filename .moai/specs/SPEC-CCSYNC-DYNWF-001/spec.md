---
id: SPEC-CCSYNC-DYNWF-001
title: "Dynamic Workflows doctrine alignment (doc seams)"
version: "0.2.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/workflow, CLAUDE.md, .claude/skills/moai-domain-research"
lifecycle: spec-anchored
tags: "documentation, dynamic-workflows, deep-research, ccsync, doc-seam"
tier: S
---

# SPEC-CCSYNC-DYNWF-001 — Dynamic Workflows doctrine alignment (doc seams)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial plan-phase authoring (Tier S, doc-seam alignment) |

## A. Context

The Claude Code Dynamic Workflows feature is already documented in MoAI doctrine across three rule files:
`.claude/rules/moai/workflow/dynamic-workflows.md`, `.claude/rules/moai/workflow/goal-directive.md`, and the
`§ Parallel Execution` section of `.claude/rules/moai/core/moai-constitution.md`. A documentation-seam review of the
feature's public capability surface against the current doctrine identified four small gaps where the existing
guidance is incomplete or where a behavior is real but unstated.

This SPEC closes those four seams. It is **documentation-only**: no Go code, no behavior change, no new test
behavior. The work is confined to prose additions in four target documents, three of which are template-distributed
and therefore subject to the 16-language neutrality contract.

## B. Problem Statement

Four documentation seams remain between the Dynamic Workflows feature surface and MoAI doctrine:

1. **Determinism note missing.** The workflow runtime rejects wall-clock and random-number calls in the workflow
   script body because non-determinism breaks the resume cache, yet `dynamic-workflows.md § How a Workflow Runs`
   does not state this constraint. An author writing a workflow script would not learn the rule until the runtime
   rejected it.

2. **`/deep-research` not wired to the research path.** The bundled `/deep-research` workflow is the canonical
   multi-source fan-out + cross-check + claim-vote path for research-heavy questions, but `CLAUDE.md § 10`
   (Web Search Protocol) and the `moai-domain-research` skill body do not cross-reference it. A reader following
   the Web Search Protocol has no pointer to the heavier fan-out option when single-pass WebSearch is insufficient.

3. **No orchestrator-facing routing heuristic.** The three-primitive comparison table in `dynamic-workflows.md`
   describes WHAT each primitive is, but offers no concise decision heuristic for WHICH to pick, with rough
   quantitative anchors.

4. **`ultracode` resume pairing unstated.** `goal-directive.md` already documents that a `/goal` must be
   re-issued after a new session; the parallel fact — that `ultracode` auto-orchestration mode also resets on a
   new session and is NOT restored by the `ultrathink.` resume-message opener — is unstated.

## C. Requirements (GEARS)

### REQ-1 — Determinism note (Ubiquitous)

The `.claude/rules/moai/workflow/dynamic-workflows.md` `§ How a Workflow Runs` section **shall** state that the
workflow script body must not call wall-clock or random-number functions because non-determinism breaks resume
caching, and that any timestamp or random value the workflow needs must be injected via the script's input
arguments or stamped onto results after the run returns. The added statement **shall** be generic and
language-neutral (no session references, no commit SHAs, no internal dates).

### REQ-2 — `/deep-research` wiring (Ubiquitous, two targets)

The `CLAUDE.md` `§ 10` (Web Search Protocol) **shall** cross-reference the bundled `/deep-research <question>`
workflow as the multi-source fan-out + cross-check + claim-vote path for research-heavy questions where single-pass
WebSearch is insufficient.

The `moai-domain-research` skill body (`.claude/skills/moai-domain-research/SKILL.md`) **shall** cross-reference the
same `/deep-research` path for the same purpose.

Both cross-references **shall** carry three facts: (a) `/deep-research` requires the WebSearch tool; (b) a workflow
run spends meaningfully more tokens than a single-pass search (cost-awareness caveat); (c) the AskUserQuestion
boundary holds — the orchestrator collects and refines the research question before launch, never mid-run.

### REQ-3 — Routing heuristic (Where, capability gate)

Where an orchestrator is choosing among the three runtime primitives, the
`.claude/rules/moai/workflow/dynamic-workflows.md` document **shall** provide a concise decision heuristic
distinguishing dynamic workflow vs Agent Teams vs sequential subagents, with rough quantitative anchors (dynamic
workflow when fanning out over dozens-to-hundreds of mostly read-only items; Agent Teams for a small number of
coordinated long-running peers; sequential subagents as the default for coding-heavy run-phase work). The heuristic
**shall** reuse and not contradict the existing three-primitive table in the same document.

### REQ-4 — `ultracode` resume pairing (Event-driven)

When a session is resumed via a paste-ready resume message, the
`.claude/rules/moai/workflow/dynamic-workflows.md` document (or `session-handoff.md`) **shall** note that
`ultracode` auto-orchestration mode resets on a new session and is NOT restored by the `ultrathink.`
resume-message opener (which restores reasoning effort only); a resumed session that needs `ultracode` must
explicitly re-issue `/effort ultracode`. The note **shall** be parallel in intent to the existing re-set `/goal`
note in `goal-directive.md`.

## D. Exclusions (What NOT to Build)

- **No workflow-run audit-trail persistence.** This SPEC does NOT add any mechanism to persist workflow-run
  results or audit trails to `.moai/reports/` or elsewhere.
- **No team-mode ↔ workflow spawn-boundary note.** The interaction between Agent Teams mode and the workflow
  spawn boundary is out of scope.
- **No Go code change.** No file under `internal/`, `pkg/`, or `cmd/` is touched. No new Go test asserting new
  behavior is added (existing template mirror-drift and neutrality tests must merely remain green).
- **No saved workflows authored.** This SPEC does NOT create any saved workflow under `.claude/workflows/`.
- **No new section restructuring.** Existing section headings in the target files are preserved; additions are
  made within or immediately adjacent to the named sections, not by reorganizing the documents.

## E. Affected Artifacts

| Artifact | Template-distributed? | REQ |
|----------|----------------------|-----|
| `.claude/rules/moai/workflow/dynamic-workflows.md` | Yes (mirror under `internal/template/templates/`) | REQ-1, REQ-3, REQ-4 |
| `CLAUDE.md` | Yes (mirror under `internal/template/templates/`) | REQ-2 |
| `.claude/skills/moai-domain-research/SKILL.md` | Yes (mirror under `internal/template/templates/`) | REQ-2 |
| `.claude/rules/moai/workflow/goal-directive.md` OR `session-handoff.md` | Yes (mirror) — only if REQ-4 lands there instead of dynamic-workflows.md | REQ-4 (alternate location) |

All listed artifacts have a byte-parity mirror under `internal/template/templates/<same-path>`; run-phase edits
are subject to the Template-First mirror obligation documented in `plan.md § D`.

Run-phase insertion anchors (heading-anchored per `plan.md § D.4`, NOT line numbers):

- `dynamic-workflows.md` REQ-1 → within `## How a Workflow Runs`
- `dynamic-workflows.md` REQ-3 → within/under `## When to Use a Dynamic Workflow`
- `dynamic-workflows.md` REQ-4 → within `## MoAI Integration Notes` (augment the existing `ultracode` bullet)
- `CLAUDE.md` REQ-2 → within `## 10. Web Search Protocol`
- `moai-domain-research/SKILL.md` REQ-2 → within `## Works Well With`
