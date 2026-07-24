---
id: SPEC-SUBAGENT-NESTING-DOCTRINE-001
title: "Subagent-nesting doctrine correction + auditor read-only nesting pilot — Progress"
version: "0.1.0"
status: draft
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: P2
phase: "v3.0.2 target"
module: ".claude"
lifecycle: spec-anchored
tags: "doctrine, subagent-nesting, claude-code, agent-authoring, sync-auditor"
tier: M
---

# Progress — SPEC-SUBAGENT-NESTING-DOCTRINE-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID `SPEC-SUBAGENT-NESTING-DOCTRINE-001` self-check: `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` → PASS (executed Bash; verbatim `PASS`).
- Artifacts authored (4): spec.md + plan.md + acceptance.md + progress.md; directory-structured (no flat file).
- Frontmatter: all 12 canonical fields present across all 4 files; `created`/`updated` (not `_at`); `tags` comma-separated string (not `labels`); `status: draft`.
- Notation: GEARS (Ubiquitous / When / While / Where / Unwanted) throughout §B.
- Out of Scope: 6 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (spec.md §E), incl. `### Out of Scope — plan-auditor nesting pilot`.
- Scope: single SPEC, two milestones (M1 doc-correction + M2 opt-in env-gated pilot) — coupled, split not recommended.
- Ground truth: orchestrator-verified (spec.md §A); 7 M1 surfaces + 2 M2 surfaces confirmed to exist with template mirrors (2026-07-24).
- Clarifications RESOLVED (plan finalization): D5 → M2 pilot scope = `sync-auditor` only (`plan-auditor` deferred to a future SPEC, spec.md §E Out of Scope — plan-auditor nesting pilot); D6 → M1 + M2 both ship in v3.0.2 (shipped default flat, `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` env opt-in only). Both clarification markers resolved; none remain (0 open).
- @MX targets: none (doctrine prose + agent frontmatter, no Go production code).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
