---
id: SPEC-ADVISOR-RUNG-001
title: "Executor-Advisor Escalation Rung + GLM Judgment Carve-Out — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "advisor-rung, escalation, moai-fix, moai-loop, glm-carve-out, workflow-reflex, plan"
---

# SPEC-ADVISOR-RUNG-001 — Plan

> plan.md is the derived execution plan. WHAT/WHY SSOT is spec.md. Tier S minimal envelope.

## §A Context

### §A.1 Problem summary

Stuck-state escalation in `/moai fix` and `/moai loop` has exactly two rungs: retry-the-same-executor, then interrupt-the-user. The per-spawn strong-model advisor primitive (archived-agent-rejection.md §C + model-policy.md `[1m]`-safe runtime args) exists but nothing triggers it conditionally. Separately, all-GLM mode (`team_mode: glm`) runs judgment gates and planning on GLM with no carve-out, while CG mode's CLAUDE.md §15 Avoid list exists precisely to keep that work off the worker-class model.

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Bash/Read, vci §2 attribution)

```
fix.md:146,184                      → "Level 3 (Review): User approval required" / "Level 3 fixes require AskUserQuestion approval"
fix.md:312-314                      → CI-loop path: "패치 실패 시 AskUserQuestion 경유 escalation"; "semantic 분류 또는 patch 실패 시 user escalation"
ci-autofix-protocol.md:36-50        → [ZONE:Frozen] 3-iteration ceiling; "iteration 4+ → MANDATORY BLOCKING AskUserQuestion" (CONST-V3R5-006)
loop.md:133                         → "Level 3 (Approval): AskUserQuestion required"
archived-agent-rejection.md:86-91   → §C rows 7-12 per-spawn Agent(general-purpose, model: opus, tools: <whitelist>) pattern
model-policy.md:96                  → per-spawn model parameter is a runtime arg, distinct from the frontmatter field that triggers the [1m] bug
team/glm.md:~99-104                 → mode table row: | glm | GLM-only | GLM | GLM |
grep -n "Avoid" team/glm.md         → 0 matches (no judgment carve-out)
CLAUDE.md §15                       → "Avoid: planning/architecture (needs Opus reasoning), security reviews, complex debugging" (CG mode only)
team/run.md:87                      → "## CG Mode (Claude Leader + GLM Teammates)"; :204 blocker report via SendMessage
```

Template mirrors verified present (2026-07-09): fix.md, loop.md, ci-autofix-protocol.md, team/glm.md, team/run.md — all under `internal/template/templates/`. Line numbers indicative; re-verify content anchors at run-phase.

### §A.3 Approach — two milestones

- **M1 — advisor rung (doc-layer)**: add the conditional advisor instruction to loop.md (iteration cycle, before user escalation) and fix.md (Level-3-class re-failure + CI-loop patch-failure path); add an Evolvable cross-reference note to ci-autofix-protocol.md (frozen clauses untouched) stating the advisor consult happens within the 3-iteration budget. Instruction content: trigger (N=2 consecutive same-diagnostic failures), spawn shape (read-only `Agent(general-purpose)`, per-spawn model/effort args per SPEC-MODEL-ROUTING-WIRE-001), evidence payload, re-seed semantics, escalate-to-user only after advisor rung fails.
- **M2 — GLM carve-out + CG advisor naming + template sync**: glm.md judgment carve-out section (reduced-assurance flag, CLAUDE.md §15 Avoid-list mirror); run.md §CG cross-ref + leader-review-as-advisor naming (blocker report → leader advisory diagnosis → re-delegation); template-first application + `make build`.

### §A.4 Tier evidence (S)

- Files affected: 5 live docs (+5 template mirrors) — surgical section-level edits, zero Go code.
- LOC estimate: <150 markdown lines total — well inside Tier S.
- Nominal Tier S "<5 files" band is exceeded only by counting mirrors; the logical edit surfaces are 5 and each edit is a bounded section insert. Recorded per the Tier honesty convention.

### §A.5 PRESERVE / EXTEND map

| Surface | Disposition |
|---------|-------------|
| ci-autofix-protocol.md [ZONE:Frozen] clauses (3-iteration ceiling, blocking AskUserQuestion) | PRESERVE verbatim; EXTEND with Evolvable advisor note only |
| fix.md Level 1-4 classification + static dispatch table (@MX:WARN guarded) | PRESERVE; EXTEND repeat-failure path |
| loop.md safety rules + Step structure (SPEC-LOOP-VERDICT-CONTRACT-001 surface) | PRESERVE; EXTEND after sibling run-phase lands (§B B10) |
| glm.md mode table + tmux mechanism | PRESERVE; EXTEND with carve-out section |
| run.md §CG | PRESERVE; EXTEND with cross-ref + advisor naming |
| CLAUDE.md §15 | PRESERVE (mirror source, unmodified) |

## §B Known Issues (filtered, Tier S)

- **B10 Scope discipline / sibling collision**: SPEC-LOOP-VERDICT-CONTRACT-001 edits loop.md Steps 1/4/9. Sequence this SPEC's loop.md hunk AFTER that SPEC's run-phase (or after explicit orchestrator coordination); re-read loop.md at run-phase start.
- **B2 Cross-SPEC policy**: ci-autofix-protocol.md carries CONST-V3R5-006/-010 frozen constitution references — the advisor note must cite, not restate or weaken, them.
- **B8 Working-tree hygiene**: specific-path `git add` only; the tree carries unrelated modified files.

## §C Pre-flight checklist

```bash
git branch --show-current && git rev-parse HEAD
git log --oneline -3 -- .claude/skills/moai/workflows/loop.md   # sibling SPEC-LOOP-VERDICT run-phase landed?
grep -n "MANDATORY BLOCKING" .claude/rules/moai/workflow/ci-autofix-protocol.md  # frozen anchor intact
grep -n "Avoid" .claude/skills/moai/team/glm.md                 # still 0 pre-edit
diff <(ls internal/template/templates/.claude/skills/moai/team/) <(ls .claude/skills/moai/team/) | head -5
```

## §D Constraints + open decisions

Constraints: see spec.md §Constraints (frozen clauses untouched; subagent boundary; doc-layer only; sibling sequencing; cost-safety).

**D1 — advisor doctrine placement.** Inline sections in loop.md + fix.md (RECOMMENDED — Tier S lean; the two surfaces have different trigger vocabularies) vs a new shared rule file `.claude/rules/moai/workflow/advisor-rung.md` cited by both. Inline is recommended; if run-phase finds the two inline sections drifting toward duplication >30 lines, fall back to the shared-rule option and cite it from both.

**D2 — ci-autofix-protocol.md touch scope.** Add a short Evolvable subsection ("Advisor consult within the iteration budget") cross-referencing fix.md (RECOMMENDED) vs leaving ci-autofix-protocol.md untouched and documenting the CI-path advisor only in fix.md. Recommended: the small Evolvable note — the protocol file is where a future reader of the frozen ceiling will look.

**D3 — same-diagnostic identity.** How "SAME diagnostic" is judged: recommended = same file + same rule/error-code (or same failing test name) across consecutive iterations, judged by the orchestrator from the failure evidence — a prose definition, no mechanical hash. Exact wording is run-phase discretion; it MUST be deterministic enough to audit post-hoc.

**D4 — advisor model/effort default.** Recommended: consult `moai route <tier> run` is NOT the right source (that routes the executor); the advisor default is the strong-model per-spawn arg (`model: "opus"`-class, high effort) per archived-agent-rejection.md §C precedent, stated as a SHOULD with explicit user override. Alternative: introduce an advisor row into workflow_agents taxonomy — rejected for Tier S (config change out of scope).

## §E Self-Verification (run-phase deliverables)

Per manager-develop-prompt-template.md §E (Tier S minimal form), vci 5-section format:
- E1: AC matrix (acceptance.md §D) with verbatim grep outputs.
- E2: `make build` exit 0 after template edits; `go build ./...` unaffected.
- E4: frozen-anchor grep — "MANDATORY BLOCKING AskUserQuestion" still present verbatim post-edit.
- E5: template-neutrality guard unaffected (no SPEC IDs leaked into `internal/template/templates/`).
- E6: commit SHAs + push state.

## §F Milestones (priority-ordered; no time estimates)

| Milestone | Scope | REQs | Exit criterion |
|-----------|-------|------|----------------|
| M1 — advisor rung | loop.md + fix.md advisor instruction; ci-autofix Evolvable note | REQ-ADV-001..003 | AC-ADV-001..004 PASS |
| M2 — GLM carve-out + naming + template sync | glm.md carve-out; run.md §CG cross-ref + advisor naming; mirrors + make build | REQ-ADV-004..006 | AC-ADV-005..007 PASS |

Dependency note: M1 loop.md hunk sequences after SPEC-LOOP-VERDICT-CONTRACT-001 run-phase (or coordinated). Whole SPEC sequences after SPEC-MODEL-ROUTING-WIRE-001 (depends_on).

## §G Anti-Patterns (do NOT)

- Editing any [ZONE:Frozen] clause in ci-autofix-protocol.md — the advisor note is additive Evolvable prose only.
- Making the advisor rung fire on first failure or per-iteration — cost-safety inversion.
- Giving the advisor Write/Edit tools or letting its diagnosis authorize semantic auto-patching (CONST-V3R5-010).
- Pinning the advisor model in any frontmatter — per-spawn runtime args only ([1m] hazard).
- Weakening or bypassing the user-escalation contracts "because the advisor already looked".
- Editing CLAUDE.md §15 — it is the mirror SOURCE, not an edit target.

## §H Cross-References

- spec.md (SSOT), acceptance.md (AC matrix), progress.md (§E skeleton).
- SPEC-MODEL-ROUTING-WIRE-001 (prerequisite: per-spawn model/effort doctrine + moai route).
- SPEC-LOOP-VERDICT-CONTRACT-001 (loop.md surface coordination).
- `.claude/rules/moai/workflow/archived-agent-rejection.md` §C (per-spawn specialist pattern).
- `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format (the boundary REQ-ADV-005 extends).
