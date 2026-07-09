---
id: SPEC-CADENCE-BRIDGE-001
title: "AUTOMATE Bridge — Sanctioned Cadence Recipes Composing Native /loop and Cron with Read-Only /moai Entry Points"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
era: V3R6
tier: S
related_specs: [SPEC-LOOP-VERDICT-CONTRACT-001]
tags: "cadence, automate, native-loop, cron, read-only, drift-watcher, backlog-rediscovery, workflow-reflex"
---

# SPEC-CADENCE-BRIDGE-001 — AUTOMATE Bridge: Sanctioned Cadence Recipes

## Epic Context

**Epic**: Workflow-Reflex (6-SPEC epic derived from the 3-lens workflow audit: model-tier routing / Loop Engineering / Harness Engineering). This SPEC is **5 of 6**.

- **Dependency notes**: No blocking dependency. **Recommended after SPEC-LOOP-VERDICT-CONTRACT-001 (3 of 6)** — the backlog re-discovery recipe reads the `.moai/state/loop-verdict-*.json` leftovers that SPEC 3 introduces (REQ-LVC-005); authoring order is soft (the recipe can cite the schema as pending), landing order SHOULD follow SPEC 3.
- **Tier**: S (minimal envelope) — see plan.md §A.4 for evidence.
- **era**: V3R6 (modern 3-phase close: plan→run→sync).

## Traceability (audit findings provenance)

| Finding ID | Severity | Summary |
|------------|----------|---------|
| L1 | HIGH | The Loop Engineering AUTOMATE element ("cadence finds work") is entirely absent: every `/moai loop` is user-initiated one-shot; Claude Code ships a native `/loop` interval scheduler and Cron tools (CronCreate etc.), goal-directive.md documents native `/loop` as "distinct... not interchangeable" with `/moai loop` but builds no bridge; zero cron/CronCreate references across `.claude/rules/` + `.claude/skills/moai/`; CI watch only activates after a PR exists |

---

## User Story

**As a** user who wants drift, over-engineering, and leftover backlog items surfaced without remembering to ask,
**I want** a sanctioned set of cadence recipes that compose Claude Code's native `/loop <interval>` scheduler (and Cron tools where appropriate) with strictly read-only `/moai` entry points — with a hard contract that scheduled runs never write, never commit, and never enter run-phase,
**so that** the harness gains the AUTOMATE element (cadence finds work; humans decide what to do with it) without ever converting scheduled convenience into unattended write automation.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

All gaps measured 2026-07-09 by this agent via Bash/Read. Line numbers indicative; content anchors are authoritative.

### GAP-1 — AUTOMATE element entirely absent (L1)

- **Measured source**: `.claude/rules/moai/workflow/goal-directive.md` § Comparing Autonomous-Continuation Approaches (table + note, observed near lines 19-31): the `/loop` (Claude Code native) row — "A fixed time interval elapses (re-runs the prompt/command on a schedule)" — and the note: *"the Claude Code native `/loop` (time-interval scheduler) and MoAI's `/moai loop` (diagnostic-driven Ralph Engine) are distinct commands ... They are not interchangeable."*; `grep -rn "CronCreate" .claude/rules/moai/ .claude/skills/moai/` → 0 matches; `grep -rniE "\bcron\b"` (same scope) → 0 matches; `.claude/skills/moai/workflows/fix.md` Related Skills (lines 312-314): CI watch activates "After `/moai sync` PR creation" — i.e. only once a PR already exists.
- **Observed pattern**: The distinctness between native `/loop` and `/moai loop` is documented, but no surface anywhere composes them: there is no sanctioned recipe for "run a read-only MoAI check on a schedule", no guidance on which `/moai` entry points are cadence-safe, and no contract for what a scheduled run does when it finds work. Every discovery pathway in the harness is user-initiated (one-shot subcommands) or PR-gated (CI watch). Work that nobody asks about — drift, over-engineering creep, ceiling-exit leftovers — is found by nothing.

### Aggregate defect claim

**Cadence never finds work: the scheduler primitives exist, the read-only entry points exist, and no sanctioned bridge composes them.** This SPEC authors the bridge as doctrine — named recipes, a HARD read-only safety constraint, and a discovery-to-queue contract — without building any scheduler mechanics in Go.

---

## Requirements (GEARS notation)

> **Subject convention**: generalized subjects ("the cadence-bridge rule", "the scheduled invocation", "the cadence run"). No legacy `IF/THEN` modality.

### REQ-CDB-001 — Ubiquitous — sanctioned recipe catalog

The cadence-bridge rule (placement per plan.md §D D1; recommended: new `.claude/rules/moai/workflow/cadence-bridge.md`) SHALL define a sanctioned recipe catalog composing native `/loop <interval>` (and Cron tools where appropriate) with READ-ONLY `/moai` entry points, containing at least these three named recipes: (1) **drift watcher** — `/loop 30m /moai gate` (lint/format/type-check/test validation, <30s class); (2) **lean review** — nightly `/moai review --lean` (read-only, advisory, over-engineering-only scan); (3) **backlog re-discovery** — periodic read of `.moai/state/loop-verdict-*.json` ceiling-exit leftovers (SPEC-LOOP-VERDICT-CONTRACT-001 REQ-LVC-005) surfacing unresolved items.

### REQ-CDB-002 — Unwanted behavior — scheduled-write prohibition (HARD)

A scheduled invocation SHALL NOT write, commit, or push: scheduled runs are restricted to read-only entry points or at most Level-1 fixes (fix.md Level 1 "Immediate: no approval required" class — formatting/import-sorting working-tree edits) with NO commit and NO push, and a scheduled invocation SHALL NOT enter run-phase — the Implementation Kickoff Approval (plan→run HUMAN GATE) remains human-only and is never satisfiable by a cadence run.

### REQ-CDB-003 — Event-driven (When) — discovery-to-queue contract

**When** a cadence run FINDS work (gate failures, lean-review findings, unresolved verdict leftovers), the cadence run SHALL persist the discovery to a queue surface (the active TaskList when a session ledger is live, otherwise a doctrine-defined backlog record per plan.md §D D2) and surface it to the user at the next interactive session — and SHALL NOT auto-execute any remediation.

### REQ-CDB-004 — Ubiquitous — distinctness preservation and bridge cross-reference

The autonomous-continuation comparison in goal-directive.md SHALL retain the native-`/loop`-vs-`/moai loop` distinctness note unchanged in meaning, EXTENDED with a cross-reference to the cadence-bridge rule as the sanctioned composition surface (the bridge composes the two; it does not merge them).

### REQ-CDB-005 — Capability gate (Where) — template-first boundary

**Where** an edited or created surface belongs to the template-distributed tree (`internal/template/templates/` mirror verified present 2026-07-09 for goal-directive.md; a NEW rule file MUST be added template-first per the Template-First rule), the run-phase SHALL apply edits template-first (edit template source, `make build`) or identically in both trees.

---

## Constraints

1. **Read-only cadence (HARD)** — REQ-CDB-002 binds every recipe, present and future: the rule MUST state the constraint as a catalog-level invariant, not per-recipe fine print.
2. **No scheduler in Go** — this SPEC is doctrine-only; no Go scheduler, no daemon, no hook that fires on wall-clock time. Native `/loop` and Cron tools are the runtime's own primitives, consumed as-is.
3. **Kickoff Approval unaffected (HARD)** — nothing in the bridge weakens the Implementation Kickoff Approval gate or the AskUserQuestion channel monopoly; a cadence discovery is an input to a human decision, never a decision.
4. **Verdict-file schema consumed, not defined** — the backlog re-discovery recipe cites SPEC-LOOP-VERDICT-CONTRACT-001's schema; this SPEC neither defines nor modifies it.
5. **GEARS notation; era V3R6; 12 canonical frontmatter fields.**

---

## Out of Scope

> Per the `OutOfScopeRule` lint, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — Go scheduler implementation

- Any Go-side scheduler, daemon, timer loop, or cron-registration code. The bridge composes runtime-native primitives (`/loop`, Cron tools) via doctrine only.

### Out of Scope — background write automation

- Scheduled write/commit/push pipelines of any kind, including "safe" auto-commit of Level-1 fixes. The HARD constraint (REQ-CDB-002) is the boundary; widening it is a different SPEC with a different risk profile.

### Out of Scope — auto-run of write-capable subcommands

- Scheduling `/moai run`, `/moai sync`, `/moai fix` (beyond the Level-1-no-commit carve-out), `/moai loop`, or any subcommand that mutates git state. Only validation/advisory entry points are cadence-eligible.

### Out of Scope — /goal and native /loop semantics

- The semantics of `/goal`, native `/loop`, Cron tools, and `/moai loop` themselves are unchanged; goal-directive.md's comparison table gains only a cross-reference.

### Out of Scope — loop-verdict schema and consumers

- The `.moai/state/loop-verdict-*.json` schema and any mechanical (Go) consumer of it — owned by SPEC-LOOP-VERDICT-CONTRACT-001 and its follow-ups. The re-discovery recipe is a prose-level reader.

---

## Cross-References

- **CREATE (doc)**: `.claude/rules/moai/workflow/cadence-bridge.md` (recommended placement per plan.md §D D1) + template mirror.
- **EXTEND base (doc)**: `.claude/rules/moai/workflow/goal-directive.md` § Comparing Autonomous-Continuation Approaches (cross-reference only; template mirror verified).
- **Read-only entry points cited**: `.claude/skills/moai/workflows/gate.md` (lint+format+type-check+test, <30s, validation-only); `.claude/skills/moai/workflows/review.md` `--lean` ("Read-only and advisory: applies no fixes, modifies no files", observed near line 43); fix.md Level 1 classification (near line 145).
- **Producer consumed**: SPEC-LOOP-VERDICT-CONTRACT-001 REQ-LVC-005 (`.moai/state/loop-verdict-<id>.json` persistence).
- **Gate preserved**: Implementation Kickoff Approval (CLAUDE.local.md §19.1 / orchestration-mode-selection.md header) — cited as untouched invariant.
- **Epic**: Workflow-Reflex 5 of 6. Siblings: SPEC-HARNESS-RATCHET-REWIRE-001 (1), SPEC-MODEL-ROUTING-WIRE-001 (2), SPEC-LOOP-VERDICT-CONTRACT-001 (3), SPEC-ADVISOR-RUNG-001 (4), SPEC-OBSERVE-HYGIENE-001 (6).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + progress). Workflow-Reflex Epic 5 of 6. Sanctioned cadence recipes + HARD read-only constraint + discovery-to-queue contract. Tier S. |
