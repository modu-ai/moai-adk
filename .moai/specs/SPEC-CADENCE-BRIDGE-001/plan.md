---
id: SPEC-CADENCE-BRIDGE-001
title: "AUTOMATE Bridge — Sanctioned Cadence Recipes — Implementation Plan"
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
tags: "cadence, automate, native-loop, cron, read-only, workflow-reflex, plan"
---

# SPEC-CADENCE-BRIDGE-001 — Plan

> plan.md is the derived execution plan. WHAT/WHY SSOT is spec.md. Tier S minimal envelope.

## §A Context

### §A.1 Problem summary

The AUTOMATE element of Loop Engineering is absent: the runtime ships a native `/loop` interval scheduler and Cron tools, MoAI ships read-only entry points (`/moai gate`, `/moai review --lean`), and SPEC 3 of this Epic introduces a persisted ceiling-exit backlog — but nothing composes them. goal-directive.md documents the native-`/loop`-vs-`/moai loop` distinction and stops there. All discovery is user-initiated or PR-gated (CI watch).

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Bash/Read, vci §2 attribution)

```
goal-directive.md:~19-31          → comparison table row "/loop (Claude Code native) | A fixed time interval elapses ..."
                                    + note "...are distinct commands — ... They are not interchangeable."
grep -rn "CronCreate" .claude/rules/moai/ .claude/skills/moai/  → 0
grep -rniE "\bcron\b"  (same scope)                              → 0
gate.md header                     → "Lightweight pre-commit quality gate ... Fast validation (<30s) without full code review"
review.md:43                       → "--lean: ... Read-only and advisory: applies no fixes, modifies no files, renders no PASS/FAIL verdict"
fix.md:144-148                     → Level 1 (Immediate) / Level 2 (Safe) / Level 3 (Review) / Level 4 (Manual) classification
fix.md:312                         → CI watch activates "After /moai sync PR creation" (PR-gated discovery only)
SPEC-LOOP-VERDICT-CONTRACT-001 REQ-LVC-005 → .moai/state/loop-verdict-<id>.json persistence (sibling, plan-phase)
goal-directive.md frontmatter      → paths: "**/goal-directive.md" (conditional loading — relevant to D1)
```

Template mirrors verified present (2026-07-09): goal-directive.md. A NEW rule file must be created template-first. Line numbers indicative; re-verify content anchors at run-phase.

### §A.3 Approach — two milestones

- **M1 — cadence-bridge rule (template-first)**: author `.claude/rules/moai/workflow/cadence-bridge.md` (per D1) containing: the recipe catalog (drift watcher `/loop 30m /moai gate`; nightly `/moai review --lean`; periodic loop-verdict backlog re-discovery), the catalog-level HARD read-only invariant (REQ-CDB-002), the discovery-to-queue contract (REQ-CDB-003), and an eligibility table stating which `/moai` entry points are cadence-safe and why (validation-only / advisory-only / prose-reader).
- **M2 — goal-directive cross-ref + template sync**: add the bridge cross-reference to goal-directive.md's comparison section (distinctness note preserved); mirror both surfaces; `make build`.

### §A.4 Tier evidence (S)

- Files affected: 1 new rule + 1 edited rule (+2 template mirrors) = 4 file operations, zero Go code.
- LOC estimate: ~120-180 markdown lines — well inside Tier S.
- Single-domain (workflow doctrine), no config, no code: minimal envelope holds cleanly.

### §A.5 PRESERVE / EXTEND map

| Surface | Disposition |
|---------|-------------|
| goal-directive.md comparison table + distinctness note | PRESERVE meaning; EXTEND with one cross-reference |
| gate.md / review.md / fix.md | PRESERVE (cited as-is; no edits) |
| `.claude/rules/moai/workflow/cadence-bridge.md` | CREATE (template-first) |
| Implementation Kickoff Approval gate + AskUserQuestion monopoly | PRESERVE (cited invariants) |
| SPEC-LOOP-VERDICT-CONTRACT-001 verdict schema | PRESERVE (consumed by reference) |

## §B Known Issues (filtered, Tier S)

- **B6 spec-lint heading convention**: not applicable to rule files, but the new rule must follow rules-file conventions (Version/Status footer, cross-references section) per existing `.claude/rules/moai/workflow/` siblings.
- **B10 Scope discipline**: sibling SPEC 3 owns loop.md/moai.md; sibling SPEC 6 owns loop-snapshots doc markers — this SPEC touches neither. goal-directive.md is not on any sibling's surface list (verified against SPECs 1-4, 6 plan scopes).
- **B8 Working-tree hygiene**: specific-path `git add` only.
- **Template-first for NEW file**: create under `internal/template/templates/.claude/rules/moai/workflow/` FIRST, then `make build`, then verify the live copy (CLAUDE.local.md §2 Template-First rule).

## §C Pre-flight checklist

```bash
git branch --show-current && git rev-parse HEAD
grep -rn "CronCreate\|cadence-bridge" .claude/rules/moai/ .claude/skills/moai/ | head -5   # still absent pre-edit
ls .moai/specs/SPEC-LOOP-VERDICT-CONTRACT-001/                                              # sibling status check
grep -n "not interchangeable" .claude/rules/moai/workflow/goal-directive.md                 # anchor intact
ls internal/template/templates/.claude/rules/moai/workflow/ | head -20                      # mirror dir baseline
```

## §D Constraints + open decisions

Constraints: see spec.md §Constraints (read-only HARD; no Go scheduler; Kickoff Approval unaffected; schema consumed not defined).

**D1 — rule placement.** New `.claude/rules/moai/workflow/cadence-bridge.md` (RECOMMENDED) vs extending goal-directive.md. Recommendation rationale: (a) goal-directive.md carries a `paths: "**/goal-directive.md"` conditional-loading frontmatter — cadence recipes deserve their own discoverable loading surface; (b) goal-directive.md is already long and single-purpose (/goal semantics); (c) a separate file mirrors cleanly and gives future recipes a home. Extending goal-directive.md remains viable if the run-phase finds the new file would be <40 lines (not expected).

**D2 — backlog record surface.** When no TaskList ledger is live: recommended = a doctrine-defined markdown/JSON backlog note under `.moai/reports/cadence/<date>.md` (reports are the analyze-what-exists namespace, gitignored-local per plan-audit precedent) — orchestrator-written, no Go loader (mirrors SPEC 3's D2 pattern). Alternative: `.moai/state/cadence-backlog.json`. Run-phase picks one and states it in the rule; the contract (persist + surface next session + never auto-execute) is invariant either way.

**D3 — Cron tools citation depth.** The rule cites Cron tools (CronCreate etc.) as runtime primitives at the level the official docs support, with native `/loop` as the primary composition example. Do NOT invent Cron tool signatures beyond what official Claude Code docs state (anti-hallucination; verify with WebFetch at run-phase if deeper citation is needed).

## §E Self-Verification (run-phase deliverables)

Per manager-develop-prompt-template.md §E (Tier S minimal form), vci 5-section format:
- E1: AC matrix (acceptance.md §D) with verbatim grep outputs.
- E2: `make build` exit 0 after template edits.
- E5: template-neutrality guard unaffected (no internal SPEC IDs in the templated rule body — cite "the loop-verdict persistence contract" generically in the TEMPLATE copy if needed, or keep the SPEC-ID citation live-only per §25 isolation; decide at run-phase against `internal_content_leak_test.go`).
- E6: commit SHAs + push state.

> Note: the template copy of cadence-bridge.md MUST NOT carry internal SPEC IDs (CLAUDE.local.md §25 / template-internal-isolation doctrine). The live copy may cite SPEC-LOOP-VERDICT-CONTRACT-001; the template copy references the mechanism generically ("the loop ceiling-exit verdict file"). This live/template divergence is sanctioned by §25 and must be recorded in E5.

## §F Milestones (priority-ordered; no time estimates)

| Milestone | Scope | REQs | Exit criterion |
|-----------|-------|------|----------------|
| M1 — cadence-bridge rule | new rule file (template-first): recipes + HARD invariant + queue contract + eligibility table | REQ-CDB-001..003 | AC-CDB-001..003, AC-CDB-005 PASS |
| M2 — cross-ref + template sync | goal-directive.md cross-reference; mirrors; make build | REQ-CDB-004..005 | AC-CDB-004, AC-CDB-006 PASS |

Dependency note: none blocking. Landing SHOULD follow SPEC-LOOP-VERDICT-CONTRACT-001 so recipe 3 references a landed schema; if it lands first, recipe 3 cites the schema as pending with a forward link.

## §G Anti-Patterns (do NOT)

- Adding ANY write-capable subcommand to the recipe catalog or eligibility table.
- Phrasing the read-only constraint per-recipe instead of as a catalog-level invariant (future recipes would escape it).
- Letting a cadence discovery auto-trigger `/moai fix` beyond Level-1-no-commit, or any run-phase entry.
- Inventing Cron tool API signatures not present in official docs.
- Leaking internal SPEC IDs into the template copy of the new rule (§25 neutrality guard).
- Merging the native-/loop and /moai-loop concepts — the bridge composes, the distinctness note stands.

## §H Cross-References

- spec.md (SSOT), acceptance.md (AC matrix), progress.md (§E skeleton).
- SPEC-LOOP-VERDICT-CONTRACT-001 (verdict-file producer; REQ-LVC-005).
- `.claude/rules/moai/workflow/goal-directive.md` (comparison table anchor).
- `.moai/docs/template-internal-isolation-doctrine.md` §25 (template neutrality for the new rule's mirror).
- CLAUDE.local.md §2 Template-First rule (new-file procedure).
