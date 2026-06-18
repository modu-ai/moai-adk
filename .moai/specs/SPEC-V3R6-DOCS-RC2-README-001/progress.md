---
id: SPEC-V3R6-DOCS-RC2-README-001
title: "Progress — v3.0.0-rc2 README + CHANGELOG factual-alignment"
version: "0.1.0"
status: draft
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "repo-root docs"
lifecycle: spec-anchored
tags: "docs, readme, changelog, progress, v3r6, rc2"
era: V3R6
tier: M
---

# Progress — SPEC-V3R6-DOCS-RC2-README-001

## §A. Current Phase

**Plan-phase** (status: draft). Plan-phase artifacts authored: spec.md, plan.md, acceptance.md, progress.md.

Run-phase (M1..M6) NOT started. Entry to run-phase requires Implementation Kickoff Approval per CLAUDE.local.md §19.1.

## §B. Milestone Status

| Milestone | Scope | Status | Commit SHA |
|-----------|-------|--------|------------|
| M1 | README.md (EN) quantitative fixes | pending (plan-phase) | — |
| M2 | README.md Mermaid + /moai db + expert-frontend + lifecycle | pending (plan-phase) | — |
| M3 | README.ko.md mirror + What's New rewrite | pending (plan-phase) | — |
| M4 | CHANGELOG rc2 promotion + mis-label cleanup | pending (plan-phase) | — |
| M5 | CLAUDE.md builder-harness path fix | pending (plan-phase) | — |
| M6 | Self-verification (grep AC suite) | pending (plan-phase) | — |

## §C. Decision Log

- **2026-06-19** — SPEC created at plan-phase. Decomposition self-check PASS. Ground-truth baseline verified against filesystem (`v3.0.0-rc2` confirmed, builder-harness path confirmed, 7 MoAI agents confirmed).
- **2026-06-19** — `era: V3R6` set explicitly in frontmatter to avoid H-6 unclassified fallback (per lifecycle-sync-gate.md § Frontmatter Era Field Semantics scenario 3).
- **2026-06-19** — Template counterpart decision (CLAUDE.md template source) DEFERRED to run-phase per plan.md §G.1.

## §D. Preserved List (Abort Recovery)

_(populated at run-phase if a run is aborted mid-milestone; empty at plan-phase)_

## §E.1 Plan-phase Audit-Ready Signal

- [x] SPEC ID decomposition self-check printed (spec.md §J): `SPEC ✓ | V3R6 ✓ | DOCS ✓ | RC2 ✓ | README ✓ | 001 ✓ → PASS`
- [x] All 12 canonical frontmatter fields present in spec.md.
- [x] `era: V3R6` set in frontmatter.
- [x] §H Exclusions contains 6 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (satisfies `OutOfScopeRule`).
- [x] Every REQ cites a DRIFT-ID; every DRIFT-ID cites a file:line.
- [x] No Go code edits in any milestone.
- [x] No template edits in plan-phase.
- [x] `v3.0.0-rc2` wording consistent; no `v3.0.0 stable` claim.
- [x] No superseded-SPEC citation (REQ-X-002).
- [x] Draft SPECs not asserted finalized (REQ-X-003).
- [x] 4 plan-phase artifacts present: spec.md, plan.md, acceptance.md, progress.md.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_

sync_commit_sha: ""
