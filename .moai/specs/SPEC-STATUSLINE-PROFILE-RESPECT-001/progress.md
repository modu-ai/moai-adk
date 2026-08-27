---
id: SPEC-STATUSLINE-PROFILE-RESPECT-001
title: "Progress — statusline opt-out honored end-to-end + subtree profile resolution"
version: "0.2.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "progress"
lifecycle: spec-anchored
tags: "progress, statusline, forge, opt-out, launch-ledger"
tier: M
---

# Progress — SPEC-STATUSLINE-PROFILE-RESPECT-001

## §A. Current Phase

**plan-phase complete; kickoff gate PASSED (2026-08-27)** — decisions D1-D5
folded, artifacts at v0.2.0. Status: draft (the `draft → in-progress`
transition belongs to manager-develop on the first run-phase commit). Run
phase may begin.

## §B. Artifact Status

| Artifact | Path | Status |
|----------|------|--------|
| spec.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/spec.md` | draft v0.2.0 (11 REQs; REQ-009 DEFERRED) |
| plan.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/plan.md` | draft v0.2.0 (M0-M4, M6, M7; M5 DEFERRED; 0 open markers) |
| acceptance.md | `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/acceptance.md` | draft v0.2.0 (11 ACs; AC-009 DEFERRED; seam placement pinned §D) |
| progress.md | this file | §E skeleton emitted |

## §C. Milestone Tracker

| Milestone | Scope | Status | Evidence |
|-----------|-------|--------|----------|
| M1 | Profile subtree resolution, READ path (REQ-006/007/008) | pending | — |
| M2 | Segment-gated refresh spawn (REQ-001/003) | pending | — |
| M3 | Explicit-override spawn early-out (REQ-002) | pending | — |
| M4 | Display-honesty characterization (REQ-005) | pending | — |
| M5 | Write-side ledger normalization (REQ-009) | **DEFERRED** → follow-up card (kickoff D1) | — |
| M6 | Test seams + verification closure (REQ-010/011) | pending | — |
| M7 | Operational edit: `forge: none` in this repo's statusline.yaml (NOT code) | pending | — |

## §D. Blockers / Open Decisions

None open — kickoff gate passed 2026-08-27. Resolutions on record:

- D1/D2 — ancestor-walk READ-path only (`ResolveLaunchProfileForProject` miss
  path); write-side normalization (REQ-009/AC-009/M5) DEFERRED to follow-up
  card "launch-ledger write-side subtree normalization".
- D2(policy) — anonymous-session default-off REJECTED; explicit config only.
- New — operational `forge: none` edit accepted → plan.md M7 (§2.3
  update-wipe caveat noted inline).
- D3 — spawn-counter seam pinned: post-freshness, pre-`isSelfInvocable`
  (acceptance.md §D preamble).
- D4 — AC-005 fixture wording corrected (absent/corrupt cache ⇒ Available=false).
- D5 — REQ-004 scoped to the recognized-override branch (none→github).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex check: executed 2026-08-27, output `PASS` for
  `SPEC-STATUSLINE-PROFILE-RESPECT-001` (pattern
  `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`); no existing SPEC directory collision.
- Frontmatter: 12 canonical fields present in spec.md per
  `.claude/rules/moai/development/spec-frontmatter-schema.md` (tier: M added as
  optional field; status: draft — the only transition manager-spec performs).
- Out of Scope: 5 `### Out of Scope — <topic>` H3 sub-headings, each with `-`
  bullets (satisfies `OutOfScopeRule`).
- Requirements notation: GEARS (capability-gate `Where`, state-driven `While`,
  event-driven `When`, unwanted `shall not`, ubiquitous). No `IF/THEN`.
- Open clarification markers: 0 — the single marker (plan.md M5, decision D1)
  was resolved at the 2026-08-27 kickoff gate and replaced by a recorded
  verbatim decision digest; AC-009/REQ-009 carry the DEFERRED outcome for
  traceability.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
