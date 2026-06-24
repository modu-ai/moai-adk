---
id: SPEC-V3R6-SKILL-DECISION-HEURISTICS-001
title: "Skill Decision Heuristics + Anti-Pattern Provenance Binding for 4 High-Traffic Skills"
version: "0.2.0"
status: completed
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
priority: P3
phase: "v3.0.0"
module: ".claude/skills"
lifecycle: spec-anchored
tags: "skills,craft,heuristics,harness"
era: V3R6
tier: S
---

# SPEC-V3R6-SKILL-DECISION-HEURISTICS-001

## §A. Problem Statement

The repo artifact `github.com/wquguru/harness-books` (`.codex/skills/harness-book-best-practice/SKILL.md`) employs two borrowable skill-craft devices that moai-adk skills currently lack:

1. **Closing "## Decision Heuristics" section** — 3-5 compact "if the question is X, default to Y" rules that give an agent FAST default-routing without loading the full skill body. The device respects the preview-pane scroll limit (Issue anthropics/claude-code#33062, ~12 visible lines) and gives a sub-100ms routing answer where the full skill body would cost a full load + read.
2. **Inline anti-pattern entries bound to the SPECIFIC past failure that motivated them** (e.g. "Do not reintroduce these settings, which were removed because <concrete reason>"). The WHY is load-bearing: a future agent that re-encounters the same wrong default has a one-line regression-resistant rationale at the point of decision, instead of having to rediscover the failure.

### The Gap (verified 2026-06-18)

- (a) The 4 highest-traffic moai-adk skills do NOT consolidate their existing body guidance into a fast-routing "Decision Heuristics" block at the end. An agent that loads the skill description must read the entire body to extract a default — costing tokens and latency on every routing decision.
- (b) The moai-adk skills use evolvable rationalization / red-flag / verification blocks (the conceptual anti-pattern content) but those entries cite the abstract rule more than the SPECIFIC past failure that motivated them. The concrete failure history (dates, SPEC-IDs, recurrence counts) lives in the memory system at `~/.claude/projects/{hash}/memory/` — the skills do not bind to it, so a future regression encounters a bare "do not X" with no provenance to weigh.

### Honest Scope Note on AP-* Codes

The SPEC source author's clause "(b) For each existing AP-* anti-pattern code already present in that skill" was empirically verified: **none of the 4 target skills currently carry inline `AP-*` codes in their bodies** (`grep -c "AP-" SKILL.md` = 0 for all four). The `AP-*` codes (AP-SRN-001..004, AP-D-001..005, AP-V-001..004, AP-GWT-001..006, AP-VBP-001..003, AP-25.1..25.3) live in the RULE files (`.claude/rules/moai/workflow/*.md`), not in these skill bodies.

This SPEC therefore reframes deliverable (b) honestly: the provenance binding applies to **the skills' existing evolvable rationalization/red-flag rows** (the conceptual anti-pattern content that IS present) — each relevant row gains a one-line past-failure provenance pulled from the memory system. Where the memory system has a specific date/SPEC-ID, the binding cites it verbatim; where it does not, the binding uses the pending-provenance form rather than inventing a date. This respects the no-unobserved-claim invariant (`.claude/rules/moai/core/verification-claim-integrity.md`).

## §B. Background & Cross-References

### Source Artifact

- Repo: `github.com/wquguru/harness-books`
- File: `.codex/skills/harness-book-best-practice/SKILL.md`
- Borrowed devices: closing Decision Heuristics section; inline past-failure provenance on each anti-pattern entry.
- License posture: idea/pattern only, no source text copied. Consistent with `.claude/rules/moai/NOTICE.md` precedent (pattern cookbooks imported as rules, not verbatim text).

### Memory Provenance Sources (verified present 2026-06-18)

| Anti-pattern family | Memory file(s) carrying date + SPEC-ID provenance |
|---|---|
| AP-V-004 (lsof filename/process-filter false-positive) | `feedback_v0b_bg_pty_host_false_positive.md` (2026-06-07, SPEC-WEB-CONSOLE-009); `lesson_v0_stale_orphan_bg_pty_host_false_positive.md` (2026-06-15, IMP-06); `project_sprint11_lifecycle_sync_gate_m4_complete_m5_handoff.md` (2026-05-30, commit `d0cfbbfe6`) |
| AP-SRN-004 (Wave→Round retired terminology) | `project_sprint10_paste_ready_improvement_chore.md` (2026-05-25, commit `64310df3f`); `project_session_handoff_track1_committed.md` (D3 finding) |
| AP-D-001..005 (paste-ready diet violations) | `project_phase1b_5th_v3a_resolved_diet_doctrine_handoff.md` (iter-N); `project_session_handoff_align_completed.md` (2026-05-30 alignment) |
| AP-V-001/002/003 (V0 abort-gate doctrine) | `project_session_handoff_align_completed.md`; `project_sec_harden_003_plan_handoff.md` (L_v0_fail_scope_decision_without_spawn) |

### Project Documents Alignment

- `.moai/project/tech.md` — Go module; skill authoring at `.claude/skills/`. No conflict.
- `.moai/project/structure.md` — skills directory layout. Additive change only (new closing section + inline one-liners).
- CLAUDE.local.md §2 Template-First Rule — LOCAL `.claude/skills/` edits are scoped; template-source mirror sync (`internal/template/templates/.claude/skills/`) is a flagged follow-up, NOT a blocker for this Tier-S SPEC.

## §C. Requirements (GEARS notation)

### REQ-SDH-001 — Closing Decision Heuristics section (Ubiquitous, all 4 skills)

The [<skill-author>] **shall** append a closing `## Decision Heuristics` section to each of the 4 target SKILL.md files (moai-foundation-core, moai-workflow-spec, moai-foundation-cc, moai-meta-harness).

### REQ-SDH-002 — 3-5 heuristics per section (Ubiquitous)

The [<skill-author>] **shall** distill between three (3) and five (5) "if X, default to Y" heuristics per Decision Heuristics section, each traceable to existing body content of that skill (the distillation cites the section each heuristic compresses).

### REQ-SDH-003 — ≤ 12 visible lines (State-driven, preview-pane scroll limit)

**While** the preview-pane scroll limit (Issue anthropics/claude-code#33062) constrains visible content to approximately twelve (12) lines, the [<skill-author>] **shall** keep each Decision Heuristics section within that bound (heading + blank line + 3-5 compact rule lines + optional one-line note).

### REQ-SDH-004 — No new doctrine invented (Unwanted behavior)

The [<skill-author>] **shall not** introduce new doctrine, new rules, or new policy in a Decision Heuristics section. Each heuristic MUST be a faithful compression of guidance already present in that skill's body, with the source section cited inline.

### REQ-SDH-005 — Anti-pattern provenance binding (Event-detected, 3 evolvable-bearing skills only)

**When** a target skill's existing evolvable rationalization or red-flag row corresponds to a known anti-pattern family documented in the memory system, **and** that target skill carries at least one `moai:evolvable-start` marker, the [<skill-author>] **shall** append a one-line past-failure provenance of the form `AP-X — recurred on <DATE> in <SPEC-ID>` (or `AP-X — observed recurrence, provenance pending in memory` where the memory lacks a specific date/SPEC). This REQ applies to `moai-foundation-core`, `moai-workflow-spec`, and `moai-foundation-cc` only. It is **N/A for `moai-meta-harness`** because that skill has zero evolvable markers (§F A-3) — provenance binding has no evolvable row to append to, and no content is fabricated.

### REQ-SDH-006 — No fabricated dates (Unwanted behavior)

The [<skill-author>] **shall not** invent a date, SPEC-ID, or commit SHA for a provenance binding. Where the memory system does not carry a specific date/SPEC for an anti-pattern family, the pending-provenance form (`observed recurrence, provenance pending in memory`) MUST be used verbatim.

### REQ-SDH-007 — Frontmatter lint preserved (State-driven)

**While** the 4 target skills must continue to pass the skill frontmatter lint, the [<skill-author>] **shall not** introduce any new snake_case YAML aliases, break existing frontmatter fields, or alter the frontmatter block — edits are strictly additive to the body.

### REQ-SDH-008 — Bodies preserved (Unwanted behavior)

The [<skill-author>] **shall not** rewrite the existing body content of the 4 target skills. The only permitted additions are: (a) the closing Decision Heuristics section, and (b) the one-line provenance bindings appended to existing evolvable rows.

### REQ-SDH-009 — Template-source sync flagged as follow-up (Capability gate)

**Where** the moai-adk skills ALSO live as template source under `internal/template/templates/.claude/skills/`, the [plan-phase artifacts] **shall** flag the template-source mirror sync as a known follow-up step (separate from this SPEC's run-phase scope) so that CLAUDE.local.md §2 Template-First Rule is respected without over-scoping a Tier-S SPEC. This REQ is a plan-phase documentation obligation only — it is verified by grep of plan.md §I, NOT by any run-phase deliverable.

## §D. Constraints

- **C-1 (Localization)**: Decision Heuristics sections are written in English (per `agent_prompt_language: en`); user-facing surface is unaffected (skill bodies are agent-facing).
- **C-2 (Scroll limit)**: Binary threshold — a Decision Heuristics section with **≤ 13 lines PASSES; ≥ 14 lines is a VIOLATION**. (12-line SHOULD target from Issue anthropics/claude-code#33062 + 1-line readability tolerance = 13-line PASS ceiling; the 14th line crosses into violation. This single threshold governs C-2 here, plan.md M3.3, and acceptance.md AC-SDH-003 uniformly — no ambiguous 14-line gap.)
- **C-3 (Provenance honesty)**: The no-unobserved-claim invariant binds. Pending-provenance form is the ONLY acceptable fallback when memory lacks a specific date/SPEC; fabricating a date is a REQ-SDH-006 violation that blocks acceptance.
- **C-4 (Template scope)**: This SPEC edits LOCAL `.claude/skills/` only. The template-source mirror sync is a flagged follow-up in plan.md, not a deliverable of this SPEC.
- **C-5 (Additive only)**: No deletions, no rewrites, no reorderings of existing body content. Diff is strictly append-only.
- **C-6 (Tier-S envelope)**: 4 files, single domain (skill craft), no Go code change, no rule-file change. Tier-S minimal envelope.

## §E. Open Questions

None at plan-phase. The honest scope note (§A) resolves the "for each existing AP-* code" ambiguity by reframing deliverable (b) onto the skills' existing evolvable rows.

## §F. Assumptions

- **A-1**: The memory system at `~/.claude/projects/{hash}/memory/` remains the authoritative provenance source; no new memory is written by this SPEC.
- **A-2**: Issue anthropics/claude-code#33062 preview-pane scroll limit (~12 visible lines) remains the binding ergonomic constraint for Decision Heuristics length.
- **A-3 (honest per-skill evolvable baseline, verified 2026-06-18 via `grep -c 'moai:evolvable-start' SKILL.md`)**: Exactly **3 of 4** target skills carry evolvable blocks — `moai-foundation-core` (3 markers: rationalizations / red-flags / verification), `moai-workflow-spec` (3 markers), and `moai-foundation-cc` (3 markers). The 4th target, **`moai-meta-harness`, carries ZERO `moai:evolvable-start` markers and has no `rationalizations` / `red-flags` / `verification` content in any form** (its body ends with plain `## Out of Scope` and `## Works Well With` sections). Therefore deliverable (b) — the provenance/red-flag one-liner binding — **applies to the 3 evolvable-bearing skills only; it is N/A for `moai-meta-harness`**. For `moai-meta-harness` this SPEC delivers **deliverable (a) only** (the closing Decision Heuristics section). No evolvable content is fabricated; the marker contract is preserved. This corrects the iter-1 §F assumption that wrongly claimed all 4 skills carry evolvable blocks (a verification-claim-integrity defect — the SPEC must match the real file state per `.claude/rules/moai/core/verification-claim-integrity.md`).

## §G. Dependencies

- None blocking. The memory system is already populated; the skills already exist; the rule files that host the canonical `AP-*` codes are not modified.

## §H. Out of Scope (What NOT to Build)

### §H.1 Out of Scope — excluded work items

- **EX-1**: Template-source mirror sync (`internal/template/templates/.claude/skills/`). Out of scope — flagged as follow-up in plan.md per CLAUDE.local.md §2.
- **EX-2**: Other moai-adk skills beyond the 4 named targets. A later SPEC may extend the pattern if the 4-skill pilot validates the device.
- **EX-3**: Rewriting any skill body, frontmatter, or existing rationalization/red-flag/verification content. This SPEC is append-only.
- **EX-4**: Editing the canonical `AP-*` code definitions in `.claude/rules/moai/workflow/*.md`. The codes live there; this SPEC only binds provenance in the skill bodies that reference the conceptual anti-pattern.
- **EX-5**: Adding NEW `AP-*` codes. This SPEC binds provenance to existing anti-pattern families; it does not mint new codes.
- **EX-6**: Changing the evolvable-block delimiters (`<!-- moai:evolvable-start -->` / `<!-- moai:evolvable-end -->`). The provenance one-liners are appended INSIDE the existing evolvable blocks, preserving the marker contract.
- **EX-7**: Run-phase implementation. This is plan-phase only; do NOT proceed to `/moai run` without explicit user approval (Implementation Kickoff Approval, CLAUDE.local.md §19.1).
- **EX-8**: Fabricating evolvable blocks in `moai-meta-harness/SKILL.md`. That skill has zero `moai:evolvable-start` markers today (verified 2026-06-18); deliverable (b) is N/A for it. See §F assumption A-3 for the honest per-skill breakdown.

## §I. Risks

- **R-1 (Low) — Heuristic drift**: A future agent might treat a Decision Heuristics rule as the full policy and skip the body. Mitigation: each heuristic ends with a `(<- §<section>)` pointer back to the body section it compresses.
- **R-2 (Low) — Provenance staleness**: The memory system accumulates new incidents; the provenance one-liners will age. Mitigation: pending-provenance form is honest about coverage gaps; a periodic refresh SPEC (not this one) can backfill.
- **R-3 (Low) — Scroll-limit violation**: An agent authoring a 5-heuristic section may exceed 12 lines. Mitigation: REQ-SDH-003 + AC-3 enforce the bound at acceptance.
- **R-4 (Low) — Template drift**: LOCAL edits land first; template mirror sync is a follow-up. Risk: brief window where local and template diverge. Mitigation: REQ-SDH-009 + plan.md follow-up flag make the divergence explicit and time-bounded.

## §J. Non-Functional Requirements

- **NFR-1 (Performance)**: Decision Heuristics section must be parseable in a single read without scrolling (preview-pane ergonomic).
- **NFR-2 (Maintainability)**: Each heuristic + each provenance one-liner must be a single line (no multi-line prose) to preserve grep-ability and visual scan-ability.
- **NFR-3 (Honesty)**: No fabricated dates. The pending-provenance form is the canonical fallback.
- **NFR-4 (Reversibility)**: All edits are additive; `git revert` of the SPEC's commits restores the prior skill bodies byte-for-byte.

## HISTORY

- **2026-06-18** — Plan-phase artifacts authored (spec.md + plan.md + acceptance.md + progress.md §E skeleton) by manager-spec. Tier S, V3R6 era, Sprint 15 harness-books application cohort (P3).
- **2026-06-18 (iter-2)** — Plan-auditor iter-1 FAIL 0.74 defect fixes: D1 §H `Exclusions`→`Out of Scope` rename + `### §H.1 Out of Scope —` h3 sub-section (resolves `MissingExclusions` ERROR; linter requires the `Out of Scope` token at h3, not only h2); D2 §F assumption corrected — 3 of 4 skills carry evolvable blocks, `moai-meta-harness` carries ZERO (verified `grep -c 'moai:evolvable-start' = 0`), so REQ-SDH-005 now scopes deliverable (b) to the 3 evolvable-bearing skills and is N/A for `moai-meta-harness` (deliverable (a) only; no fabricated content — verification-claim-integrity); D3 acceptance.md REQ-SDH-009 mapped to real AC-SDH-008 (was "(meta)"); D4 line-count threshold standardized to binary ≤13 PASS / ≥14 VIOLATION across C-2 / M3.3 / AC-SDH-003; D5 REQ-SDH-009 subject rewritten from `[<plan.md>]` (a file) to `[plan-phase artifacts]` (system actor). REQ count unchanged (9); AC count 7→8.
