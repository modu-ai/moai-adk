---
id: SPEC-DIVECC-DELEGATION-TOKEN-COST-001
title: "Delegation Token-Cost (~7×) Signal — Prefer Skill Injection over Agent Spawn"
version: "0.1.0"
status: draft
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/output-styles/moai"
lifecycle: spec-anchored
tags: "delegation, token-cost, skill-vs-agent, decision-criterion, doctrine, dogfooding, divecc"
era: V3R6
tier: S
---

# SPEC-DIVECC-DELEGATION-TOKEN-COST-001 — Delegation Token-Cost (~7×) Signal

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | Initial plan-phase draft. Candidate N3 of Epic Dive-into-CC. Paper-claim premise; light grounding pinned the target surfaces precisely (see §B.2 — the ROADMAP's "CLAUDE.md §16" was imprecise; the actual surfaces are moai.md §4 + CLAUDE.local.md §16). Thematic pair with N2 (SPEC-DIVECC-EXTENSION-COST-LADDER-001). |

---

## §A. Background

### A.1 Epic provenance (Dive-into-CC dogfooding)

This SPEC is candidate **N3** of the **Epic Dive-into-CC** (domain token `DIVECC`), a dogfooding exercise applying findings from a reverse-engineering analysis of Claude Code internals to moai-adk's own harness. The Epic roadmap lives at `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md`.

Source body of work (one publication, two surfaces):

- **arXiv:2604.14228** — "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (Liu, Zhao, Shang, Shen, 2026, cs.SE).
- **github.com/VILA-Lab/Dive-into-Claude-Code** — companion repository + "Build Your Own AI Agent: A Design Guide".

### A.2 The paper claim this SPEC records

The paper observes a token-cost asymmetry between Claude Code's two principal delegation mechanisms:

- A **Skill** (`SkillTool`) injects its content into the **current** context window — cheap. The model continues in the same conversation; only the skill body's tokens are added.
- An **Agent** (`AgentTool`) spawns an **isolated** context window — the spawned sub-agent re-establishes its own working context from scratch. The paper measures this isolated-context overhead at roughly **~7× the token cost** of a Skill injection for comparable work.

The decision insight (paper claim): **choosing between a Skill and an Agent is a token-cost decision, not only a capability decision.** Prefer Skill injection when shared context is acceptable; spawn an Agent only when isolation is genuinely needed (independence / bias-prevention / parallel fan-out / read-only investigation that should not pollute the main context).

> **Provenance discipline (verification-claim-integrity §1.1)**: the ~7× figure is recorded here as the **paper's measurement of Claude Code internals** (cite arXiv:2604.14228), NOT as a moai-adk verification claim. moai-adk has not independently benchmarked the multiplier. The run-phase edit records it the same way — attributed to the paper. The only moai-tree assertions in this SPEC are the §B.2 grounding observations about what the target surfaces currently contain, each backed by an actual Read.

### A.3 Why this is worth documenting in moai-adk

moai-adk's delegation decision today weighs **quality / independence / bias** (the three questions in `moai.md` §4) — but carries no **token-cost** axis. The result: the natural reach is for an `Agent()` spawn (isolated, "clean") even when a Skill injection (shared context, ~7× cheaper) would meet the need. Adding the token-cost signal lets the orchestrator prefer the cheaper mechanism when isolation is not actually required. This parallels the sibling N2 SPEC's extension-mechanism context-cost ladder: N2 covers *which extension mechanism* (Hooks/Skills/Plugins/MCP), N3 covers *Skill-vs-Agent at delegation time*. Both are halves of one cost-model refinement.

---

## §B. Problem Statement & Grounding

### B.1 Problem

The canonical Delegation Decision surface (`moai.md` §4) lists three weighing questions — specialist domain? / agent exists? / does delegation beat direct work on **quality / independence / bias**? — and a Forced Delegation table. None of these carries a token-cost lens. An orchestrator that has decided "delegate" then reaches for `Agent()` by default, even for work a Skill could inject into the current context at ~7× lower token cost. There is no documented "prefer Skill injection when shared context is acceptable; spawn an Agent only when isolation is genuinely needed" directive.

### B.2 Grounding (light — paper claim; target-surface pinning corrected)

This SPEC's premise is a paper claim about Claude Code internals, so it needs only light grounding: confirm the target moai-adk surfaces still read as described, and **pin the correct target file(s)** rather than carry the ROADMAP's reference forward unverified.

> **ROADMAP imprecision corrected.** The ROADMAP named the target as `.claude/output-styles/moai/moai.md` §4 + "CLAUDE.md §16 self-check (four self-check questions)". Grounding shows the §16 self-check is in **`CLAUDE.local.md` §16**, NOT `CLAUDE.md` §16 (`CLAUDE.md` §16 is "Context Search Protocol", unrelated). The pinned targets below replace the ROADMAP reference.

**Observed (Read of `.claude/output-styles/moai/moai.md` §4 "Delegation Decision (§24 Self-Check)", lines 101–138, 2026-06-22):**
- §4 lists **three** self-check questions: (1) Is this a specialist domain? (2) Does the specialist agent exist in the catalog? (3) Does delegation beat direct work on **quality, independence, bias**?
- It has a Forced Delegation Table and Volume Triggers. It carries **no token-cost axis**.
- `moai.md` IS template-distributed — `internal/template/templates/.claude/output-styles/moai/moai.md` exists (confirmed via `ls`). So a §4 edit MUST be mirrored to the template tree + `make build` + neutrality check.
- This is the **primary, template-managed** target.

**Observed (Read of `CLAUDE.local.md` §16 "오케스트레이터 자가 점검", lines 632–657, 2026-06-22):**
- §16 carries the **four** self-check questions (the ROADMAP's "four self-check questions"): (1) specialist domain? (2) catalog exists? (3) delegation beats direct on quality/independence/bias? (4) decomposable into read-only sub-agent parallel spawn?
- It is **`CLAUDE.local.md`** — explicitly local-only, maintainer-only, **NOT template-managed** (per `CLAUDE.local.md` header + §2 separation rules; `CLAUDE.local.md` is in the Local-Only Files list and a forbidden internal-content reference class for templates).
- **Scoping note (LOAD-BEARING)**: editing `CLAUDE.local.md` §16 is a *local-only maintainer convenience* edit — it does NOT propagate to user projects and is NOT subject to template mirroring. The run-phase MAY add the token-cost signal to `CLAUDE.local.md` §16 for maintainer-side coherence, but the **canonical, user-distributed** home for the directive is `moai.md` §4. Do NOT attempt to mirror a `CLAUDE.local.md` edit into the template tree (that would violate template isolation doctrine — `CLAUDE.local.md` references are a forbidden content class).

**Conclusion (pinned targets):**
- **Primary target (REQUIRED)**: `.claude/output-styles/moai/moai.md` §4 — add the token-cost signal as a 4th weighing consideration / decision note. Template-managed → mirror + build + neutrality check.
- **Secondary target (OPTIONAL, local-only)**: `CLAUDE.local.md` §16 — MAY add the same signal for maintainer coherence; NOT mirrored, NOT user-distributed.

---

## §C. Requirements (GEARS)

> Notation: GEARS (current). `<subject>` is generalized.

**REQ-DTC-001 (Ubiquitous)** — The `moai.md` §4 Delegation Decision surface **shall** carry a token-cost signal stating that a Skill injects into the current context (cheap) while an Agent spawns an isolated context (~7× the token cost per the paper claim).

**REQ-DTC-002 (Ubiquitous)** — The token-cost signal **shall** state the directive: prefer Skill injection when shared context is acceptable; spawn an Agent only when isolation is genuinely needed (independence / bias-prevention / parallel fan-out / read-only investigation).

**REQ-DTC-003 (Ubiquitous)** — The token-cost signal **shall** attribute the ~7× figure to the source paper (arXiv:2604.14228) as an external grounding, NOT as a moai-adk verification claim.

**REQ-DTC-004 (Ubiquitous)** — The run-phase implementation **shall** mirror the `moai.md` §4 edit to the template tree (`internal/template/templates/.claude/output-styles/moai/moai.md`), run `make build`, and verify template neutrality, because `moai.md` is template-distributed.

**REQ-DTC-005 (Where capability gate)** — **Where** maintainer-side coherence warrants it, the run-phase **may** additionally add the same token-cost signal to `CLAUDE.local.md` §16 (the local-only 4-question self-check). This edit is local-only and **shall not** be mirrored to the template tree.

**REQ-DTC-006 (Unwanted behavior)** — The token-cost signal **shall not** present the ~7× figure as a moai-adk measurement; it is the paper's measurement of Claude Code internals.

**REQ-DTC-007 (Unwanted behavior)** — The run-phase implementation **shall not** alter the existing three (or four) weighing questions' meaning, the Forced Delegation Table, or the Volume Triggers. The token-cost signal is an ADDITIVE axis, not a replacement of the existing quality/independence/bias weighing.

**REQ-DTC-008 (Unwanted behavior)** — The run-phase implementation **shall not** edit the sibling N2 extension-mechanism ladder surface (`.claude/rules/moai/development/agent-authoring.md`); that surface is owned by SPEC-DIVECC-EXTENSION-COST-LADDER-001.

---

## §D. Acceptance Criteria (inline — Tier S)

> Tier S: AC inline in spec.md §D (no separate acceptance.md). Each AC is observable/mechanical (grep / file-existence / cross-ref presence). These ACs bind the **run-phase** outcome; they are NOT satisfied at plan-phase.

**AC-DTC-001 — Token-cost signal present in moai.md §4** (binds REQ-DTC-001)
- GIVEN the run-phase edit to `.claude/output-styles/moai/moai.md`
- WHEN `grep -n -i "7×\|7x\|isolated context\|Skill injects\|token cost" .claude/output-styles/moai/moai.md` is run within the §4 Delegation Decision region
- THEN the Skill-injects-current-context (cheap) vs Agent-spawns-isolated-context (~7×) asymmetry appears in §4.
- Verification: `grep -i "7×\|7x" .claude/output-styles/moai/moai.md` returns ≥ 1 match in the §4 region.

**AC-DTC-002 — Directive stated** (binds REQ-DTC-002)
- GIVEN the added signal
- WHEN the §4 region is read
- THEN it states "prefer Skill injection when shared context is acceptable; spawn an Agent only when isolation is genuinely needed".
- Verification: `grep -i "prefer Skill\|isolation is genuinely needed\|only when isolation" .claude/output-styles/moai/moai.md` returns ≥ 1 match.

**AC-DTC-003 — Paper attribution present** (binds REQ-DTC-003, REQ-DTC-006)
- GIVEN the added signal
- WHEN `grep "2604.14228\|Dive into Claude Code" .claude/output-styles/moai/moai.md` is run
- THEN the ~7× figure is attributed to the paper, with no moai-tree verification claim for the figure (a "paper claim" / "per the paper" / equivalent qualifier is present near the figure).
- Verification: `grep -i "paper\|2604.14228" .claude/output-styles/moai/moai.md` returns ≥ 1 match near the token-cost signal.

**AC-DTC-004 — Template mirror + neutrality** (binds REQ-DTC-004)
- GIVEN `moai.md` is template-distributed
- WHEN the run-phase mirrors the §4 edit to `internal/template/templates/.claude/output-styles/moai/moai.md` and runs `make build`
- THEN the mirrored content matches the local edit AND contains no forbidden internal-content class (no internal SPEC ID, no REQ/AC token, no commit SHA, no internal date, no `CLAUDE.local.md` reference) per `.moai/docs/template-internal-isolation-doctrine.md`.
- Verification: `go test ./internal/template/... -run TestTemplateNeutralityAudit` passes (run-phase gate); `diff <(sed -n '/Delegation Decision/,/Checkpoint/p' .claude/output-styles/moai/moai.md) <(sed -n '/Delegation Decision/,/Checkpoint/p' internal/template/templates/.claude/output-styles/moai/moai.md)` shows the token-cost signal in both.

**AC-DTC-005 — Existing weighing preserved** (binds REQ-DTC-007)
- GIVEN the run-phase edit
- WHEN the §4 region is read
- THEN the existing three weighing questions (specialist domain / agent exists / quality·independence·bias), the Forced Delegation Table, and the Volume Triggers are all still present and unchanged in meaning.
- Verification: `grep -c "specialist domain\|Forced Delegation\|Volume Trigger" .claude/output-styles/moai/moai.md` ≥ 3.

**AC-DTC-006 — Local-only edit not mirrored** (binds REQ-DTC-005)
- GIVEN the OPTIONAL `CLAUDE.local.md` §16 edit (if the run-phase chooses to make it)
- WHEN `git show --stat <run-commit>` is inspected
- THEN any `CLAUDE.local.md` change is NOT accompanied by a corresponding template-tree change (CLAUDE.local.md is local-only, never mirrored); the template diff contains only the `moai.md` mirror.
- Verification: the run-commit's template-tree diff contains `moai.md` but NOT any `CLAUDE.local.md`-derived content.

**AC-DTC-007 — N2 surface untouched** (binds REQ-DTC-008)
- GIVEN the run-phase commit diff
- WHEN `git show --stat <run-commit>` is inspected
- THEN `.claude/rules/moai/development/agent-authoring.md` is NOT modified by this SPEC's run-phase (that surface is owned by N2).
- Verification: `git show --stat <run-commit>` file list does not include `agent-authoring.md`.

---

## §E. Lifecycle Progress Markers

> Plan-phase emits the §E section skeleton (placeholder headings only). §E.2–§E.4 are populated by manager-develop (run) and manager-docs (sync), not by this plan-phase author. See progress.md for the canonical skeleton.

- **§E.1 Plan-phase Audit-Ready Signal** — see progress.md §E.1.
- §E.2 Run-phase Evidence — _pending run-phase_.
- §E.3 Run-phase Audit-Ready Signal — _pending run-phase_.
- §E.4 Sync-phase Audit-Ready Signal — _pending sync-phase_.

---

## §F. Out of Scope

This section bounds what this SPEC does NOT cover (satisfies `OutOfScopeRule` / avoids `MissingExclusions`).

### Out of Scope — run-phase implementation

- Authoring the token-cost signal text itself into `moai.md` §4 (or `CLAUDE.local.md` §16). This SPEC is plan-phase only; the edits happen at run-phase against the AC matrix in §D.
- Changing the meaning of the existing three/four weighing questions, the Forced Delegation Table, or the Volume Triggers (REQ-DTC-007). The token-cost signal is purely additive.

### Out of Scope — the sibling N2 extension-mechanism ladder surface

- Any edit to `.claude/rules/moai/development/agent-authoring.md`. That surface (the extension-mechanism context-cost ladder Hooks/Skills/Plugins/MCP) is owned by the thematic-pair SPEC **SPEC-DIVECC-EXTENSION-COST-LADDER-001** (N2). N3 covers the *Skill-vs-Agent* delegation-time token-cost signal; N2 covers the *which-extension-mechanism* ladder. Disjoint target surfaces.

### Out of Scope — template mirroring of CLAUDE.local.md

- Mirroring any `CLAUDE.local.md` §16 edit into the template tree. `CLAUDE.local.md` is local-only / maintainer-only and is a forbidden internal-content reference class for templates (per `CLAUDE.local.md` §2 + §25). The only template-mirrored edit in this SPEC is the `moai.md` §4 edit (REQ-DTC-004/005).

### Out of Scope — token-cost-figure verification

- Independently benchmarking the Skill-vs-Agent token multiplier in the moai-adk tree. The ~7× figure is recorded as the paper's claim (arXiv:2604.14228), not a moai-adk measurement (REQ-DTC-006). Producing a moai-tree benchmark is out of scope.

### Out of Scope — other Epic Dive-into-CC candidates

- N4 (observability loop), N5 (compaction naming), N6 (unified inventory), N7 (paper archival). Each is its own SPEC when promoted; this SPEC touches none of their surfaces.

---

## §G. Cross-References

- **arXiv:2604.14228** — "Dive into Claude Code" (the source of the Skill-vs-Agent ~7× token-cost claim).
- `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` — Epic Dive-into-CC roadmap (§N3 candidate detail). Note: the ROADMAP's "CLAUDE.md §16" target reference is corrected by this SPEC's §B.2 grounding to `CLAUDE.local.md` §16 + `moai.md` §4.
- **SPEC-DIVECC-EXTENSION-COST-LADDER-001** (N2) — thematic pair; the extension-mechanism context-cost ladder half of the cost-model refinement. N2 ladder ↔ N3 token-cost are the two halves of the same cost-model refinement.
- `.claude/output-styles/moai/moai.md` §4 (Delegation Decision) — primary run-phase target (template-managed).
- `CLAUDE.local.md` §16 (오케스트레이터 자가 점검) — optional secondary run-phase target (local-only, NOT mirrored).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — provenance discipline (paper claim vs moai-tree verification claim).
- `.moai/docs/template-internal-isolation-doctrine.md` — template neutrality gate for the `moai.md` mirror; the doctrine that makes `CLAUDE.local.md` a forbidden-mirror class.
