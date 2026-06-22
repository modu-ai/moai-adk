---
id: SPEC-DIVECC-EXTENSION-COST-LADDER-001
title: "Extension-Mechanism Context-Cost Ladder (Hooks → Skills → Plugins → MCP)"
version: "0.1.0"
status: in-progress
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/development"
lifecycle: spec-anchored
tags: "extension-mechanism, context-cost, decision-criterion, doctrine, dogfooding, divecc"
era: V3R6
tier: S
---

# SPEC-DIVECC-EXTENSION-COST-LADDER-001 — Extension-Mechanism Context-Cost Ladder

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | Initial plan-phase draft. Candidate N2 of Epic Dive-into-CC. Paper-claim premise; light grounding at authoring (see §B.2). Thematic pair with N3 (SPEC-DIVECC-DELEGATION-TOKEN-COST-001). |

---

## §A. Background

### A.1 Epic provenance (Dive-into-CC dogfooding)

This SPEC is candidate **N2** of the **Epic Dive-into-CC** (domain token `DIVECC`), a dogfooding exercise that applies findings from a reverse-engineering analysis of Claude Code internals to moai-adk's own harness. The Epic roadmap lives at `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md`.

Source body of work (one publication, two surfaces):

- **arXiv:2604.14228** — "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (Liu, Zhao, Shang, Shen, 2026, cs.SE).
- **github.com/VILA-Lab/Dive-into-Claude-Code** — companion repository + "Build Your Own AI Agent: A Design Guide".

### A.2 The paper claim this SPEC records

The paper observes that Claude Code's four extension mechanisms occupy distinct context-cost tiers — choosing an extension mechanism is fundamentally a **context-cost decision**, not only a capability decision. The ladder (paper claim):

| Mechanism | Context cost (paper claim) | Why |
|-----------|----------------------------|-----|
| **Hooks** | **zero** | Shell scripts executed out-of-band by the runtime; emit exit codes / structured JSON. No tokens are injected into the model's context window. |
| **Skills** | **low** | Progressive disclosure — metadata listed at session start (~100 tokens), body loaded on invocation (~5K), bundled files on demand. |
| **Plugins** | **medium** | Bundle agents/skills/commands/hooks; loading a plugin's agents and skills adds their combined footprint. |
| **MCP** | **high** | Each MCP server's tool schemas are injected into context; a server with many tools is the densest per-capability context cost (mitigated only by deferred/`alwaysLoad` gating). |

> **Provenance discipline (verification-claim-integrity §1.1)**: the four-tier ladder is recorded here as the **paper's claim** about Claude Code internals (cite arXiv:2604.14228), NOT as a moai-adk verification claim. The run-phase edit records it the same way — as an external grounding, attributed to the paper. The only moai-tree assertion in this SPEC is the §B.2 grounding observation about what `agent-authoring.md` currently contains, which is backed by an actual Read.

### A.3 Why this is worth documenting in moai-adk

moai-adk authors agents, skills, plugins, hooks, and MCP integrations as a routine part of harness development (the `builder-harness` agent's whole remit). Today the choice among these mechanisms is made on **capability** grounds ("can a hook do this? does it need a tool schema?") with no explicit **context-cost** lens. The paper's ladder gives a third decision axis. It parallels two taxonomies already in the rule tree:

- the **skill 3-level progressive-disclosure taxonomy** (`skill-authoring.md` § Progressive Disclosure — metadata / body / bundled), which is a *within-skill* cost axis; and
- the **dynamic-workflows purpose→effort taxonomy** (`dynamic-workflows.md` § Purpose-driven model+effort selection — read-only-extract→low … implement→xhigh), which is a *per-agent-spawn* cost axis.

This SPEC adds a third parallel axis: **mechanism→context-cost** (a *cross-mechanism* cost axis). The three axes together form a coherent cost-model layer for harness authoring.

---

## §B. Problem Statement & Grounding

### B.1 Problem

When `builder-harness` (or a maintainer) decides how to deliver a new capability, no rule names the context-cost difference between Hooks, Skills, Plugins, and MCP. The natural tendency is to reach for the most *capable* mechanism (MCP, a plugin) when a *cheaper* mechanism (a hook, a skill) would suffice, silently inflating the per-session context budget. There is no documented "prefer the cheapest mechanism that meets the capability requirement" criterion.

### B.2 Grounding (light — paper claim, single moai-tree observation)

This SPEC's premise is a paper claim about Claude Code internals, so it needs only light grounding: confirm the target moai-adk surface still reads as described before scoping the edit.

**Observed (Read of `.claude/rules/moai/development/agent-authoring.md`, 2026-06-22):**
- The file is `paths:`-scoped to `**/.claude/agents/**` (frontmatter line 2) and is the canonical agent-authoring rule.
- It documents agent frontmatter fields, namespace separation, per-spawn domain specialization, and the static-file-vs-per-spawn decision tree.
- It does **not** currently contain any "extension-mechanism context-cost ladder" or "mechanism→context-cost" axis. The closest adjacent content is the per-spawn-vs-static decision tree (§ Static Agent File vs Per-Spawn Specialization Decision Tree) — a *which-agent-form* axis, distinct from the *which-mechanism* axis this SPEC adds.

**Observed (Read of cross-reference anchors, 2026-06-22):**
- `skill-authoring.md` § Progressive Disclosure (the 3-level taxonomy) exists and is the canonical within-skill cost axis.
- `dynamic-workflows.md` § Purpose-driven model+effort selection (the purpose→effort taxonomy) exists and is the canonical per-spawn effort/cost axis.

Both cross-reference anchors the run-phase edit will cite are present. No moai-tree fact beyond these Read-backed observations is asserted.

---

## §C. Requirements (GEARS)

> Notation: GEARS (current). `<subject>` is generalized.

**REQ-ECL-001 (Ubiquitous)** — The `agent-authoring.md` rule **shall** document an extension-mechanism context-cost ladder classifying Hooks (zero), Skills (low), Plugins (medium), and MCP (high) as four distinct context-cost tiers.

**REQ-ECL-002 (Ubiquitous)** — The ladder **shall** state the decision criterion that choosing an extension mechanism is a context-cost decision: prefer the cheapest mechanism on the ladder that meets the capability requirement.

**REQ-ECL-003 (Ubiquitous)** — The ladder section **shall** attribute the four-tier cost claim to the source paper (arXiv:2604.14228) as an external grounding, NOT as a moai-adk verification claim.

**REQ-ECL-004 (Ubiquitous)** — The ladder section **shall** cross-reference the two existing parallel cost-axis taxonomies: the skill 3-level progressive-disclosure taxonomy (`skill-authoring.md` § Progressive Disclosure) and the dynamic-workflows purpose→effort taxonomy (`dynamic-workflows.md` § Purpose-driven model+effort selection).

**REQ-ECL-005 (Where capability gate)** — **Where** a builder-harness policy section is the more appropriate home for the decision criterion than `agent-authoring.md`, the run-phase implementation **may** additionally place a one-line cross-reference pointer in that policy section. The primary target remains `agent-authoring.md`.

**REQ-ECL-006 (Unwanted behavior)** — The ladder section **shall not** assert that moai-adk measured or verified the four-tier cost figures; the figures are the paper's claim and **shall not** be presented as moai-tree measurements.

**REQ-ECL-007 (Unwanted behavior)** — The run-phase implementation **shall not** modify any extension mechanism's runtime behavior, any hook script, any skill body, any plugin, or any MCP configuration. This SPEC is doctrine-only (documentation of a decision criterion).

---

## §D. Acceptance Criteria (inline — Tier S)

> Tier S: AC inline in spec.md §D (no separate acceptance.md). Each AC is observable/mechanical (grep / file-existence / cross-ref presence). These ACs bind the **run-phase** outcome; they are NOT satisfied at plan-phase.

**AC-ECL-001 — Ladder present with 4 tiers** (binds REQ-ECL-001)
- GIVEN the run-phase edit to `.claude/rules/moai/development/agent-authoring.md`
- WHEN `grep -i -E "hooks?.*zero|skills?.*low|plugins?.*medium|mcp.*high" .claude/rules/moai/development/agent-authoring.md` is run
- THEN at least the four tier labels (Hooks/zero, Skills/low, Plugins/medium, MCP/high) appear in the added ladder section.
- Verification: `grep -c -i "context-cost ladder\|extension-mechanism" .claude/rules/moai/development/agent-authoring.md` ≥ 1.

**AC-ECL-002 — Decision criterion stated** (binds REQ-ECL-002)
- GIVEN the added ladder section
- WHEN the section is read
- THEN it states the "choosing an extension mechanism is a context-cost decision / prefer the cheapest mechanism that meets the capability requirement" criterion.
- Verification: `grep -i "context-cost decision\|cheapest mechanism" .claude/rules/moai/development/agent-authoring.md` returns ≥ 1 match.

**AC-ECL-003 — Paper attribution present** (binds REQ-ECL-003, REQ-ECL-006)
- GIVEN the added ladder section
- WHEN `grep "2604.14228\|Dive into Claude Code" .claude/rules/moai/development/agent-authoring.md` is run
- THEN the paper is cited as the source of the cost claim, and the section contains no moai-tree verification claim for the figures (a "paper claim" / "as the paper observes" / equivalent attribution qualifier is present near the ladder).
- Verification: `grep -i "paper claim\|the paper\|2604.14228" .claude/rules/moai/development/agent-authoring.md` returns ≥ 1 match.

**AC-ECL-004 — Both parallel taxonomies cross-referenced** (binds REQ-ECL-004)
- GIVEN the added ladder section
- WHEN `grep -E "skill-authoring|Progressive Disclosure" .claude/rules/moai/development/agent-authoring.md && grep -E "dynamic-workflows|Purpose-driven" .claude/rules/moai/development/agent-authoring.md` is run
- THEN both cross-references resolve to ≥ 1 match each.
- Verification: two grep commands each return ≥ 1 match.

**AC-ECL-005 — No runtime/mechanism change** (binds REQ-ECL-007)
- GIVEN the run-phase commit diff
- WHEN `git show --stat <run-commit>` is inspected
- THEN no file under `.claude/hooks/`, `.claude/skills/*/SKILL.md` body, plugin manifests, or MCP config (`.mcp.json`, settings MCP entries) is modified; the only changed file is `.claude/rules/moai/development/agent-authoring.md` (plus optional builder-harness policy pointer per REQ-ECL-005, and the template mirror if `agent-authoring.md` is template-distributed).
- Verification: `git show --stat <run-commit>` file list contains only the rule file (+ optional mirror + optional policy pointer).

**AC-ECL-006 — Template-neutrality preserved (if mirrored)** (binds template isolation doctrine)
- GIVEN `agent-authoring.md` is template-distributed (it lives under `.claude/rules/`)
- WHEN the run-phase mirrors the edit to `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`
- THEN the mirrored content contains no forbidden internal-content class (no internal SPEC ID, no REQ/AC token, no commit SHA, no internal date) per `.moai/docs/template-internal-isolation-doctrine.md` — the ladder + paper citation + generic cross-references are all in the acceptable class.
- Verification: `go test ./internal/template/... -run TestTemplateNeutralityAudit` passes (run-phase gate).

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

- Authoring the ladder text itself into `agent-authoring.md`. This SPEC is plan-phase only; the rule edit happens at run-phase against the AC matrix in §D.
- Editing any hook script, skill body, plugin manifest, or MCP configuration. The ladder is a *decision criterion document*, not a mechanism change (REQ-ECL-007).

### Out of Scope — the sibling N3 delegation-token-cost surface

- Any edit to `.claude/output-styles/moai/moai.md` §4 (Delegation Decision) or `CLAUDE.local.md` §16 (self-check). Those surfaces are owned by the thematic-pair SPEC **SPEC-DIVECC-DELEGATION-TOKEN-COST-001** (N3). N2 covers the *extension-mechanism* cost ladder; N3 covers the *delegation* (Skill-vs-Agent) token-cost signal. The two are complementary halves of the same cost-model refinement but have disjoint target surfaces.

### Out of Scope — other Epic Dive-into-CC candidates

- N4 (observability loop), N5 (compaction naming), N6 (unified inventory), N7 (paper archival). Each is its own SPEC when promoted; this SPEC touches none of their surfaces.

### Out of Scope — cost-figure verification

- Independently measuring or benchmarking the per-mechanism context cost in the moai-adk tree. The four-tier ladder is recorded as the paper's claim (arXiv:2604.14228), not a moai-adk measurement; producing moai-tree cost benchmarks is explicitly out of scope (REQ-ECL-006).

---

## §G. Cross-References

- **arXiv:2604.14228** — "Dive into Claude Code" (the source of the extension-mechanism cost-ladder claim).
- `.moai/specs/SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` — Epic Dive-into-CC roadmap (§N2 candidate detail).
- **SPEC-DIVECC-DELEGATION-TOKEN-COST-001** (N3) — thematic pair; the delegation token-cost (~7×) half of the cost-model refinement. N2 ladder ↔ N3 token-cost are the two halves of the same cost-model refinement.
- `.claude/rules/moai/development/agent-authoring.md` — run-phase target (the ladder lands here).
- `.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure — the parallel within-skill 3-level cost axis (cross-referenced by the ladder).
- `.claude/rules/moai/workflow/dynamic-workflows.md` § Purpose-driven model+effort selection — the parallel per-spawn purpose→effort cost axis (cross-referenced by the ladder).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — provenance discipline (paper claim vs moai-tree verification claim).
