---
id: SPEC-DISCOVERY-UNKNOWNS-001
title: "Unknowns framework Tier-1 for Context-First Discovery — Blind Spot Pass + decision-reversibility ordering + 4-quadrant lens"
version: "0.1.0"
status: draft
created: 2026-07-05
updated: 2026-07-05
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/core"
lifecycle: spec-anchored
tier: M
tags: "discovery, unknowns, blind-spot, gears, askuser, planning, context-first, doc"
---

# SPEC-DISCOVERY-UNKNOWNS-001 — Implementation Plan

## §A. Context

- **Work location**: `/Users/goos/MoAI/moai-adk-go` (main checkout, Route A Hybrid Trunk main-direct — Tier M default; no PR, no worktree).
- **Tier**: M (standard) — 4 doctrine surfaces × template mirror = 8 paired edit targets + `make build` + no-Go-change verification. plan-auditor PASS threshold 0.80.
- **SPEC artifacts**: `.moai/specs/SPEC-DISCOVERY-UNKNOWNS-001/{spec.md, plan.md, acceptance.md, progress.md}`.
- **Development mode**: DDD/TDD not applicable — this is a documentation/rule/agent-body SPEC with zero Go code. Run-phase is markdown authoring + `make build` re-embed + verification greps.
- **PRESERVE (do NOT restructure)**: the existing section structure, heading anchors, and cross-references of all 4 target files. Enhancements are additive subsections + single-line insertions, not rewrites.
- **EXTEND**: `askuser-protocol.md` (new Blind Spot Pass subsection + 4-quadrant lens paragraph in § Ambiguity Triggers); `manager-spec.md` (plan.md ordering guidance); `CLAUDE.md` §7 Rule 1 + Rule 5 (minimal insertions); `spec-workflow.md` § Plan Phase (one-line ref).

### §A.1 Scope SSOT

The authoritative scope is spec.md §B (three Tier-1 enhancements) + §C (REQ-DU-001..015) + §E (Out of Scope). This plan.md derives the execution order; the SPEC body governs WHAT/WHY.

### §A.2 Dogfooding note (T2 applied to this very plan)

The milestone ordering below is itself ordered by decision-reversibility (T2) — leading with the highest-change-likelihood decision (the new user-facing Blind Spot Pass rule text, most subject to annotation revision) and deferring the mechanical `make build` + verification to the bottom.

## §B. Known Issues (filtered for a doc/rule SPEC)

- **B4. Frontmatter canonical schema**: all 4 SPEC artifacts use `created:` / `updated:` / `tags:` (no snake_case aliases). Verified at author time.
- **B6. spec-lint heading convention**: spec.md exclusions use `### Out of Scope — <topic>` H3 sub-headings (not a bare `## Out of Scope` H2), each with `-` bullets, to satisfy `OutOfScopeRule` (`MissingExclusions`). Verified in spec.md §E.
- **B10. Untouched paths PRESERVE (CRITICAL)**: the shared checkout has ~52 uncommitted files from a PARALLEL session (SPEC-WEB-CONSOLE-011 / SPEC-HANDOFF-CTXGUIDE-001). Run-phase MUST touch ONLY the 4 target surfaces + their 4 template mirrors + this SPEC directory. NO `git add -A` / `git add .`; use pathspec-restricted staging only. Do NOT touch, stage, or modify any file outside the enumerated edit set.
- **B11. AskUserQuestion boundary**: run-phase is subagent-executed; if the exact insertion wording needs a user decision, return a structured blocker report — never prompt the user.
- **Template divergence (this SPEC)**: `askuser-protocol.md` diverges local-vs-template by ~1 hunk and `manager-spec.md` by ~5 hunks / 10 lines (all §25-neutrality strips of internal dates / SPEC-IDs / REQ tokens — verified via live diff). Neither divergence touches this SPEC's T1/T2 edit anchors (the manager-spec `Step 4` / `**plan.md**` T2 anchor is confirmed untouched); the paired edit targets identical anchors in each surface. `make build` re-embeds the mirror edits.

## §C. Pre-flight Check List (run-phase, before any edit)

```bash
# 1. Baseline: confirm current branch + HEAD
git branch --show-current
git rev-parse HEAD

# 2. Confirm the 3 gaps still hold (no parallel session pre-empted them)
grep -rniE 'blind ?spot|unknown[ -]unknown' .claude/rules .claude/agents   # expect only manager-git.md:85 incidental
grep -niE 'unknown|blind' .claude/rules/moai/core/askuser-protocol.md        # expect no matches
grep -rniE 'reversib|most likely to change|change[ -]likelihood' .claude/agents/moai/manager-spec.md  # expect no matches

# 3. Confirm all 8 edit targets exist (4 local + 4 mirror)
ls .claude/rules/moai/core/askuser-protocol.md .claude/agents/moai/manager-spec.md CLAUDE.md .claude/rules/moai/workflow/spec-workflow.md
ls internal/template/templates/.claude/rules/moai/core/askuser-protocol.md internal/template/templates/.claude/agents/moai/manager-spec.md internal/template/templates/CLAUDE.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md

# 4. Build baseline (make build must already succeed before edits)
make build
```

## §D. Constraints (DO NOT VIOLATE)

- **Doc/rule/agent-body ONLY** — no `.go` file created or modified (REQ-DU-013). Verified via `git status --porcelain | grep '\.go$'` returning empty for this SPEC's changes.
- **Paired edits** — every content edit to a mirrored file is applied to BOTH local and template; `make build` afterward (REQ-DU-014).
- **§25 neutrality** — template-mirror inserts contain NO internal SPEC ID (`SPEC-DISCOVERY-UNKNOWNS-001`), internal dates, commit SHAs, or audit citations (REQ-DU-015). The unknowns framework prose itself is generic and safe to ship.
- **CLAUDE.md diet** — §7 Rule 5 gets a minimal combined addition (≤ ~2 lines: the T1 Blind Spot Pass ref + the T3 one-line lens framing + SSOT pointer); §7 Rule 1 gets one line (T2). No section-block expansion (REQ-DU-012).
- **Optionality preserved** — the Blind Spot Pass text states it is OPTIONAL, not a mandatory gate (REQ-DU-004).
- **Boundary preserved** — Blind Spot Pass text reaffirms `Agent(Explore)` never prompts the user; surfacing is via the orchestrator's AskUserQuestion (REQ-DU-005).
- **Working-tree hygiene** — pathspec-restricted staging only; do NOT touch parallel-session files (B10).
- **Forbidden commands** — no `git add -A`, no `git add .`, no `--no-verify`, no `--amend`, no force-push.

## §E. Self-Verification (deliverables for run-phase completion)

Run-phase (manager-develop) MUST report:
- **E1. AC PASS/FAIL matrix** — all 15 ACs (AC-DU-001..015) with the exact grep/build command + observed output per acceptance.md § D.
- **E2. make build result** — `make build` exit 0 (templates re-embedded).
- **E3. Template parity** — for each of the 4 files, the new phrase/section present in BOTH local and mirror (grep both).
- **E4. §25 neutrality grep** — `grep -rn 'SPEC-DISCOVERY-UNKNOWNS' internal/template/templates/` returns zero matches.
- **E5. No-Go-change assertion** — `git status --porcelain` shows no `.go` file in this SPEC's change set.
- **E6. Commit/push state** — Conventional Commit subjects (Route A main-direct), pathspec-restricted; new commit SHAs listed.

## §F. Milestones (decision-reversibility ordered — dogfoods T2)

> Ordered highest-change-likelihood first. Each milestone performs the PAIRED (local + template) edit for its surface; `make build` runs once in M4 after all mirror edits land.

### M1 — Blind Spot Pass technique + 4-quadrant lens (askuser-protocol.md) — HIGHEST change-likelihood

The core new user-facing Discovery rule text — most subject to annotation revision (the "data model / new interface" analogue). Paired edit (local + template mirror, §25-neutral):
- Add a new `## Blind Spot Pass` (or `### Blind Spot Pass` under a suitable H2) subsection to `askuser-protocol.md` defining: the trigger (unfamiliar domain + suspected unknown-unknowns), the mechanism (`Agent(Explore)` read-only domain scan → surface likely unknown-unknowns via `AskUserQuestion`), the OPTIONAL nature (not a mandatory gate), and the boundary reaffirmation (Explore never prompts the user directly).
- Add the Known-Knowns / Known-Unknowns / Unknown-Knowns / Unknown-Unknowns 4-quadrant taxonomy paragraph to § Ambiguity Triggers and Exceptions, with the explicit routing: suspected Unknown-Unknowns → run a Blind Spot Pass.
- Covers: REQ-DU-001, REQ-DU-002, REQ-DU-003, REQ-DU-004, REQ-DU-005, REQ-DU-010, REQ-DU-011.

### M2 — Plan decision-reversibility ordering guidance (manager-spec.md + CLAUDE.md §7 Rule 1) — HIGH change-likelihood

Behavior-shaping wording decisions. Paired edits:
- `manager-spec.md`: amend the plan.md authoring guidance (§ Step 4 `**plan.md**` line + the plan.md section-structure guidance) to instruct that plan.md milestones/sections lead with the decisions most likely to change (data-model changes, new type interfaces, user-facing/UX flows) and defer mechanical/refactoring steps to the bottom.
- `CLAUDE.md` §7 Rule 1 (Approach-First): add one line instructing that Approach-First reports present the highest change-likelihood decisions first.
- Covers: REQ-DU-007, REQ-DU-008, REQ-DU-009.

### M3 — Entry-point wires (spec-workflow.md Plan Phase + CLAUDE.md §7 Rule 5) — LOW change-likelihood (mechanical)

Single-line references into the entry points. Paired edits:
- `spec-workflow.md` § Plan Phase: one-line reference to the Blind Spot Pass technique (SSOT = askuser-protocol.md).
- `CLAUDE.md` §7 Rule 5: minimal combined addition (≤ ~2 lines) — the T1 Blind Spot Pass reference + the T3 one-line 4-quadrant lens framing + SSOT pointer. Diet-respecting.
- Covers: REQ-DU-006, REQ-DU-012.

### M4 — Template-embed consolidation + verification — BOTTOM (purely mechanical)

- Run `make build` once to regenerate the embedded template FS from all M1-M3 mirror edits.
- Verify template parity (E3), §25 neutrality (E4), no-Go-change (E5), and run the full AC-DU matrix (E1).
- Covers: REQ-DU-013, REQ-DU-014, REQ-DU-015.

## §G. Anti-Patterns (avoid in run-phase)

- Rewriting existing sections instead of additive insertion (breaks anchors + cross-references; PRESERVE per §A).
- Editing local without the template mirror (or vice-versa) — breaks Template-First parity (REQ-DU-014).
- Leaking `SPEC-DISCOVERY-UNKNOWNS-001` / dates / SHAs into template content (violates §25 / REQ-DU-015).
- Expanding CLAUDE.md §7 Rule 5 into a taxonomy block (violates the diet / REQ-DU-012).
- Promoting the Blind Spot Pass to a mandatory gate (violates REQ-DU-004).
- Having Explore or a subagent prompt the user (violates REQ-DU-005 boundary).
- `git add -A` / touching parallel-session files (B10 working-tree hazard).
- Running `make build` per-milestone (wasteful; once at M4 after all mirror edits suffices).

## §H. Cross-References

- spec.md §A.4 — the 4-surface paired-edit table (local ↔ mirror).
- acceptance.md § D — the AC-DU-001..015 matrix (verification SSOT).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § B (B4/B6/B10/B11) — known-issue categories applied above.
- `CLAUDE.local.md` §2 (Template-First) + §25 (neutrality) — the paired-edit + neutrality doctrine.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (Route A) — main-direct, no PR/worktree for Tier M.
