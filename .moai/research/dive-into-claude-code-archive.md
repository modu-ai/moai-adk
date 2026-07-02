---
name: dive-into-claude-code-archive
description: Durable archive of the VILA-Lab "Dive into Claude Code" paper (arXiv:2604.14228) — Claude Code v2.1.88 internals authority reference + consolidation of 4 scattered in-repo citation surfaces (SPEC-DIVECC-PAPER-ARCHIVE-001 / Epic Dive-into-CC N7)
created: 2026-06-22
updated: 2026-07-02
author: manager-develop (N7 run-phase)
related_spec: SPEC-DIVECC-PAPER-ARCHIVE-001
related_epic: Epic Dive-into-CC (DIVECC)
---

# "Dive into Claude Code" Paper Archive

> **Purpose**: a single durable archive entry for the academic paper that Epic Dive-into-CC reverse-engineers and applies to moai-adk's own harness doctrine. Epic Dive-into-CC's prior candidates (N2 / N3 / N5) each cited this paper at a different in-repo surface, but no single durable archive entry held the paper's full picture in one place. This file consolidates them.
>
> **Framing boundary (read first)**: the citation is **VERIFIED-by-citation** in sibling N5 (`SPEC-DIVECC-COMPACTION-LAYER-NAMING-001`). This archive does **NOT** re-fetch or re-verify the arXiv paper — it treats the citation as established fact and records what moai-adk already consumes from it. Every figure and taxonomy recorded below (the "98.4% / 1.6%" thesis, the 5-layer compaction names) is framed as **the paper's own claim** (a citation of arXiv:2604.14228), NOT as a moai-adk measurement or behavioral assertion. This boundary aligns with `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (a defect / claim is a citation until the moai-tree itself verifies it). See §5 below for the full framing-boundary statement.

---

## §1. Full Bibliographic Citation

| Field | Value |
|-------|-------|
| **Title** | "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" |
| **Authors** | VILA-Lab (Liu, Zhao, Shang, Shen, 2026) |
| **arXiv ID** | arXiv:2604.14228 |
| **Subject class** | cs.SE (Software Engineering) |
| **Year** | 2026 |
| **Companion repository** | github.com/VILA-Lab/Dive-into-Claude-Code (includes a companion "Build Your Own AI Agent: A Design Guide") |
| **Subject of analysis** | Claude Code **v2.1.88** internals (reverse-engineering of the agent harness) |

The paper is the academic reverse-engineering of Claude Code's internals at version v2.1.88. moai-adk is a harness built **on top of** Claude Code; it CONSUMES the Claude Code internals this paper describes and cannot modify the native runtime loop. This archive records the paper as moai-adk's authority reference for Claude Code v2.1.88 architecture/internals.

---

## §2. Central Thesis (the paper's own claim)

> **"98.4% infrastructure, 1.6% AI"** — the paper's central figure: the agent loop itself is a trivial `while`-loop, and the differentiator is the **harness** (the infrastructure surrounding the loop), NOT the model call.

This "98.4% / 1.6%" split is recorded here as **the paper's own thesis** (a citation of arXiv:2604.14228), NOT as a moai-adk measurement. The Epic Dive-into-CC dogfooding programme takes this thesis as its external grounding: if the harness is where the leverage is, then moai-adk's harness doctrine is where moai-adk should invest — which is exactly what the Epic's candidates do (they apply the paper's harness observations to moai-adk's own rules). The paper's open-direction enumeration (the design-space taxonomy in §3) is likewise the paper's own claim, not a moai-adk assertion.

---

## §3. CC-internals Consumed by moai-adk

The following five paper claims are the CC-internals content moai-adk has **already consumed** across the four in-repo citation surfaces (§4). Each row names the surface that consumes it.

| # | CC-internals content (paper claim) | Consumed by surface |
|---|-------------------------------------|---------------------|
| 1 | **5-layer graduated-compaction taxonomy** (see §3.1 below) | #2 (context-window-management.md) + #3 (runtime-recovery-doctrine.md §1) |
| 2 | **AI-agent-system design-space taxonomy** + open-direction enumeration (the "98.4% infrastructure" central thesis and its design-space framing) | #2 / #3 (as the framing that motivates the compaction-layer naming) |
| 3 | **query-loop / withheld-recoverable-error framing** — the input-governance preamble the query loop performs BEFORE the model call (memory prefetch → snip → microcompact → context-collapse → autocompact), during which recoverable errors are withheld and routed to layered recovery | #3 (runtime-recovery-doctrine.md §1) |
| 4 | **delegation-mechanism cost signal** — a Skill injects into the current context (cheap); an Agent spawns an isolated context that re-establishes from scratch (expensive). The paper reports **agent teams in plan mode at roughly ~7× the tokens of a single session** — a related but distinct comparison; the Skill-over-Agent cost gap is a moai extrapolation of that isolated-context principle (additionally supported by Anthropic's ~15× multi-agent figure), not a skill-vs-agent benchmark by the paper | #1 (moai.md §4 Token-Cost Axis) |
| 5 | **extension-mechanism context-cost ladder** — Hooks (zero) → Skills (low) → Plugins (medium) → MCP (high); the cheapest mechanism that meets the capability requirement should be preferred | #4 (agent-authoring.md Extension-Mechanism Context-Cost Ladder) |

### §3.1 The 5-layer graduated-compaction taxonomy (consume-not-implement boundary)

The paper names Claude Code's graduated-compaction mechanism as **five escalating layers** that progressively reduce the live input before each model call, in escalation order:

```
Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact
```

[ZONE:Frozen-spirit] **moai-adk CONSUMES the `Budget Reduction` → ... → Auto-Compact graduated-compaction layers; it does NOT implement them.** These five layers are Claude Code runtime internals — the interception lives inside Claude Code's `queryLoop()`, out of moai-adk's reach. moai-adk is a harness ON TOP of Claude Code; it consumes the graduated compaction (its `/clear` discipline and model-specific thresholds sit atop the runtime mechanism), but it cannot reimplement or modify the native compaction loop. Recording these layer names is **provenance enrichment** — it names the runtime mechanism that moai-adk's context-window-management `/clear` thresholds sit on top of, and changes no moai-adk behavior. This consume-not-implement framing is the same boundary that sibling N5 (`SPEC-DIVECC-COMPACTION-LAYER-NAMING-001`) established for these layer names.

> **Convergent second source**: the same five-layer graduated-compaction concept appears independently in `github.com/wquguru/harness-books` book1 ch03 (`memory prefetch → snip → microcompact → context-collapse → autocompact` input-governance sequence). The two sources map onto each other (the paper's leading `Budget Reduction` layer corresponds to the budget-reduction step book1 folds into its `memory prefetch` preamble). The convergence is recorded in `runtime-recovery-doctrine.md` §1 — two independent citations for one concept that moai-adk consumes, not implements.

---

## §4. In-repo Citation Surfaces

The paper is cited at **four canonical surfaces** in the moai-adk tree. Each surface consumes a distinct slice of the paper (per §3). All four cite the same publication (arXiv:2604.14228); there is no inter-surface inconsistency.

| # | Citation surface | Added by Epic candidate | What it consumes |
|---|------------------|-------------------------|------------------|
| 1 | `.claude/output-styles/moai/moai.md` | N3 (DELEGATION-TOKEN-COST) | delegation-mechanism cost signal (Skill-over-Agent, a moai extrapolation; the paper's ~7× figure is the agent teams in plan mode vs single session comparison) — §4 Token-Cost Axis |
| 2 | `.claude/rules/moai/workflow/context-window-management.md` | N5 (COMPACTION-LAYER-NAMING) | the 5-layer graduated-compaction taxonomy (the `/clear` thresholds sit atop these layers) |
| 3 | `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §1 | N5 (COMPACTION-LAYER-NAMING) | "convergent second source" — the 5-layer naming corroborated against book1 ch03; the query-loop / withheld-recoverable-error framing |
| 4 | `.claude/rules/moai/development/agent-authoring.md` | N2 (EXTENSION-COST-LADDER) | extension-mechanism context-cost ladder (Hooks/Skills/Plugins/MCP) |

A reader who lands on any one of these four surfaces can reach this durable archive via the consolidation cross-reference line added to `runtime-recovery-doctrine.md` §5 (Cross-References) by this SPEC (SPEC-DIVECC-PAPER-ARCHIVE-001 / N7) — that line maps "4 scattered citations → 1 durable archive".

---

## §5. Provenance & Framing Boundary

[ZONE:Frozen-spirit] This archive records the paper's claims as **the paper's own claims** — citations of arXiv:2604.14228 — NOT as moai-adk measurements or behavioral assertions. The distinction is load-bearing:

- The "98.4% infrastructure, 1.6% AI" figure (§2) is the paper's thesis, not a moai-adk benchmark of its own tree.
- The 5-layer graduated-compaction taxonomy (§3.1) names a **Claude Code runtime** mechanism that moai-adk **consumes, does not implement** — the interception lives inside Claude Code's `queryLoop()`, out of moai-adk's reach.
- The paper's ~7× figure (§3 row 4) is the **agent teams in plan mode vs single session** comparison (the paper's own claim, not a moai-adk benchmark); the Skill-over-Agent delegation-cost gap that moai.md §4 applies is a moai extrapolation of that isolated-context principle, not something the paper measured for the skill-vs-agent case.
- The extension-mechanism cost ladder (§3 row 5) is the paper's claim about Claude Code's internals, recorded as external grounding for moai-adk's decision criterion, not a moai-adk measurement (the agent-authoring.md surface states this explicitly).

This framing aligns with `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3: a defect / debt / drift / external claim is a **hypothesis (a citation) until the moai-tree's own dedicated tooling verifies it**. None of the figures above were independently re-measured in the moai-adk tree by this SPEC (and re-measuring them is explicitly out of scope — REQ-PA-008). The citation itself was established by sibling N5 (VERIFIED-by-citation); this archive consolidates, it does not re-verify.

---

## §6. References

- **arXiv:2604.14228** — "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (VILA-Lab, 2026, cs.SE) — the archived paper.
- **github.com/VILA-Lab/Dive-into-Claude-Code** — companion repository + "Build Your Own AI Agent: A Design Guide".
- `github.com/wquguru/harness-books` book1 ch03 — convergent second source for the graduated-compaction layer sequence (`memory prefetch → snip → microcompact → context-collapse → autocompact`).
- `.moai/research/gears-paper-validation.md` — precedent paper-archive entry (house-style model for this file).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — provenance discipline (paper-citation vs moai-tree behavioral assertion).
- `.moai/specs/SPEC-DIVECC-PAPER-ARCHIVE-001/` — the N7 SPEC that authored this archive.
- `.moai/specs/SPEC-DIVECC-COMPACTION-LAYER-NAMING-001/` — sibling N5 (citation provenance + consume-not-implement boundary model).

---

**End of archive entry.**
