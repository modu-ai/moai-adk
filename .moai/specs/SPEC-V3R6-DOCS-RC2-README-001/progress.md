---
id: SPEC-V3R6-DOCS-RC2-README-001
title: "Progress — v3.0.0-rc2 README + CHANGELOG factual-alignment"
version: "0.3.0"
status: in-progress
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

**Run-phase** (status: draft → in-progress on M1 commit). Plan-phase complete; spec.md v0.3.0, plan.md, acceptance.md authored + iter-3 plan-audit PASS-WITH-DEBT 0.88.

Implementation Kickoff Approval (§19.1): **GRANTED** 2026-06-19 (plan-auditor verdict surfaced + plan summary presented; user approved run-phase entry; skip-eligible NO since 0.88 < 0.90 → explicit approval required).
Phase 0.95 Mode Selection: **sub-agent (Mode 5)** — see §E.0 below.
Rebase onto origin/main: divergence `13 2` → `0 2` (DSR 13 commits absorbed, scope independent). Backup: `backup/pre-rebase-docs-rc2-e8771c318`.

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
- **2026-06-19 (iter-3 plan-audit fix)** — iter-3 plan-audit fix applied per plan-auditor iter-2 FAIL 0.74 verdict (2 BLOCKING + 2 SHOULD-FIX). D1' skills 32→31 LIVE (root cause = DSR rebase `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` deleted `moai-design-system` post-authoring → the iter-1 count of 32 became stale carry-over in spec.md §B Ground Truth, a `verification-claim-integrity.md` §2 attribution violation). D2' AC-KO-001b line-anchored on `README.ko.md:111` (KO D6 discipline parity — L58/126/560/1215 carry the CORRECT `16개` token, so the KO whole-file grep was vacuously satisfied). D3 precise Go LOC dropped from §B (pipeline over-counts non-deterministically; 191,248 LIVE). D5 AC-CL-001 internal contradiction split into two binary ACs (AC-CL-001a + AC-CL-001b). Frontmatter version 0.2.0 → 0.3.0; status unchanged (draft). spec.md / plan.md / acceptance.md updated in lockstep; README.md / README.ko.md / CHANGELOG.md / CLAUDE.md NOT touched (run-phase deliverables).
- **2026-06-19** — rebase onto origin/main: divergence `13 2` → `0 2` (clean ahead). DSR 13 commits (SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001, e453bcc86~935b2fc07) absorbed via `git rebase origin/main`; scope independent (DSR touched README.md/README.ko.md/CHANGELOG.md/CLAUDE.md 0 times). Backup branch `backup/pre-rebase-docs-rc2-e8771c318`; reflog preserves pre-rebase HEAD (ecd270369). Root cause of D1' drift traced here: DSR M1 deleted moai-design-system → skills 32→31 between spec authoring and now.
- **2026-06-19** — plan-auditor iter-3 re-audit: **PASS-WITH-DEBT 0.88** (Tier M 0.80 cleared; Clarity 0.90 / Completeness 0.95 / Testability 0.88 / Traceability 0.95). All 4 iter-2 defects CONFIRMED FIXED via LIVE re-derivation. 2 MINOR nits remain (AC-EN-010 body H4 sequence; AC-KO-001b commentary token) — non-blocking debt. skip-eligible NO (0.88 < 0.90). Report: `.moai/reports/plan-audit/SPEC-V3R6-DOCS-RC2-README-001-2026-06-19-iter3.md`.
- **2026-06-19** — Implementation Kickoff Approval (§19.1) **GRANTED** by user. plan-auditor iter-3 verdict + plan summary surfaced via AskUserQuestion; user selected "run-phase 진입 승인". Phase 0.5 arc: iter-1 FAIL 0.74 → iter-2 FAIL 0.74 → iter-3 PASS-WITH-DEBT 0.88 (monotonic +0.14).
- **2026-06-19** — Phase 0.95 Mode Selection: **sub-agent (Mode 5)** — sequential manager-develop M1→M6 (single coupled repo-root docs domain, 4 files sharing one §B baseline). Full evaluation: §E.0.

## §D. Preserved List (Abort Recovery)

_(populated at run-phase if a run is aborted mid-milestone; empty at plan-phase)_

## §E.0 — Phase 0.95 Mode Selection

**Decision: sub-agent (Mode 5)**

Input parameters:
- tier: M
- scope: 4 files (README.md, README.ko.md, CHANGELOG.md, CLAUDE.md §4)
- domain count: 1 (repo-root docs factual-alignment)
- file language mix: 100% markdown (zero Go code)
- concurrency benefit: LOW (single tightly-coupled doc domain; EN/KO/CHANGELOG/CLAUDE.md share one §B ground-truth baseline — 8 agents, 31 skills, 100K+ LOC, 16 langs, 17 commands)
- Agent Teams prereqs: NOT all met (single domain, 4 files — below the 3-domain OR 10-file scope gate)

Mode evaluation:
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | 20 REQs / 30 ACs / 6 milestones — non-trivial |
| 2 background | no | Write/Edit required — not read-only |
| 3 agent-team | no | 1 domain, 4 files — below scope gate |
| 4 parallel | no | doc editing, not multi-domain research; coding-task parallelism caveat applies |
| 5 sub-agent | **yes** | sequential manager-develop M1→M6; single coupled doc domain favors sequential over parallel |
| 6 workflow | no | 4 files, not ≥30 mechanical-uniform |

Justification: SPEC A is a single-domain repo-root docs factual-alignment task across 4 tightly-coupled markdown files sharing one §B baseline. Sequential manager-develop per-milestone delegation (M1→M6) is the correct fit: the EN/KO mirror + CHANGELOG + CLAUDE.md edits must stay consistent with one baseline, so parallel fan-out would risk cross-file ground-truth drift (the very hazard this SPEC exists to fix). Mode 5 is the Anthropic-recommended default for non-research-parallel work.

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
