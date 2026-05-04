<!-- Verifies REQ-BRAIN-004: proposal.md contains SPEC Decomposition Candidates section with 2-10 entries -->
<!-- Verifies REQ-BRAIN-011: NO tech-stack assumptions in proposal.md -->
<!-- Verifies REQ-BRAIN-008: 16-language neutrality enforced at ideation layer -->
---
name: moai-domain-ideation
description: >
  Ideation domain specialist: Lean Canvas assembly, SPEC decomposition list extraction,
  and Diverge-Converge pipeline for product proposal generation. Use during /moai brain
  Phase 2 (Diverge), Phase 4 (Converge), and Phase 6 (Proposal).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-05-04"
  modularized: "false"
  tags: "ideation, lean-canvas, diverge-converge, spec-decomposition, proposal, brain"
  related-skills: "moai-foundation-thinking, moai-domain-design-handoff, moai-domain-research"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["ideation", "lean canvas", "diverge", "converge", "brainstorm", "proposal", "decomposition", "brain"]
  agents: ["manager-brain"]
  phases: ["brain"]
---

<!-- @MX:ANCHOR: [AUTO] SPEC Decomposition Candidates grammar — canonical definition -->
<!-- @MX:REASON: Consumed by /moai plan --from-brain (high fan_in). Grammar MUST remain stable across brain workflow versions. -->

# Ideation Domain Specialist

Thin orchestrator for the Diverge-Converge ideation pipeline. Delegates creative framework execution to `moai-foundation-thinking` and adds artifact-shaping logic specific to the brain workflow: Lean Canvas section assembly and SPEC Decomposition List extraction.

## Quick Reference

Core responsibilities:
- Phase 2 (Diverge): Generate 5-15 divergent concept angles for the idea
- Phase 4 (Converge): Assemble Lean Canvas into `ideation.md` with all 9 blocks
- Phase 6 (Proposal): Produce `proposal.md` with SPEC Decomposition Candidates section

Key invariants:
- [HARD] SPEC Decomposition Candidates grammar: `- SPEC-{DOMAIN}-{NUM}: {scope}` (see anchor below)
- [HARD] No tech-stack assumptions (language/framework agnostic) in any artifact
- [HARD] Lean Canvas always includes all 9 blocks (missing blocks get placeholder text)
- [HARD] SPEC IDs use generic domain labels (e.g., SPEC-API-001, SPEC-AUTH-001) — never language-specific (e.g., SPEC-FASTAPI-001)

Foundation reuse:
- Diverge step: delegates to `moai-foundation-thinking` modules/diverge-converge.md (Diverge phase)
- Converge step: delegates to `moai-foundation-thinking` modules/diverge-converge.md (Converge phase)
- Critical eval: delegates to `moai-foundation-thinking` modules/critical-evaluation.md

---

## Phase 2: Diverge

### Input

- Clarity-scored idea from Phase 1 Discovery
- User context from AskUserQuestion rounds

### Process

Invoke `moai-foundation-thinking` Diverge-Converge framework (Diverge phase):

1. Generate 5-15 divergent angles for the idea. Each angle explores a different lens:
   - Core feature set angle (minimum viable product)
   - Target user segment angle (niche vs broad market)
   - Distribution channel angle (B2C, B2B, marketplace, API)
   - Revenue model angle (subscription, freemium, per-use, enterprise)
   - Technical differentiation angle (AI, real-time, offline-first, mobile-first)
   - Competitor gap angle (what existing tools fail to do)
   - Adjacent market angle (related problem space)

2. For each angle, produce a one-sentence concept label.

3. Cluster related angles by affinity (max 5 clusters).

### Output

In-memory concept map. NOT persisted to disk at this phase — convergence in Phase 4 determines what is written.

### Language Neutrality Rules

[HARD] During divergence, do NOT anchor any angle to a specific programming language or framework. Describe capabilities, not implementations:
- Correct: "real-time collaborative editing engine"
- Wrong: "Node.js WebSocket server with React frontend"

---

## Phase 4: Converge — Lean Canvas Assembly

### Input

- Phase 2 diverged concept map
- User's original idea
- Optional: brand context from `.moai/project/brand/brand-voice.md`

### Process

Invoke `moai-foundation-thinking` Diverge-Converge framework (Converge phase) to reduce 5-15 angles to the single most defensible product concept.

Then assemble the Lean Canvas.

### Lean Canvas — 9 Blocks

Populate each block for the converged concept. Every block MUST be present, even if sparse. Empty blocks get placeholder: `[TBD — to be refined with user research]`.

```
## Lean Canvas

### Problem
[Top 3 problems this product solves for the target customer]

### Customer Segments
[Specific user personas — who has the problem most acutely?]

### Unique Value Proposition
[Single, clear, compelling message — why this over alternatives]

### Solution
[Top 3 features / capabilities that address the problems]

### Channels
[How the product reaches customers: direct, marketplace, viral, partnerships]

### Revenue Streams
[How value is monetized: subscription, freemium, per-use, enterprise license, API]

### Cost Structure
[Main cost drivers: infrastructure, people, acquisition, support]

### Key Metrics
[The numbers that tell you the product is succeeding — leading and lagging indicators]

### Unfair Advantage
[What is genuinely hard for competitors to copy? Network effect, data, brand, IP, team]
```

### Language Neutrality in Solution Block

[HARD] The Solution block describes WHAT the product does, not HOW it is built:
- Correct: "High-throughput transformation engine that processes 10K events/sec"
- Wrong: "Python Pandas pipeline running on Airflow DAGs"

### Output

Write `ideation.md` to `.moai/brain/IDEA-NNN/`:

```markdown
# Idea: {user's original idea, verbatim}
*Session: {date}*

## Lean Canvas

[9 blocks as specified above]

```

---

## Phase 5 Append: Critical Evaluation

After Phase 5 executes (managed by `moai-foundation-thinking` critical-evaluation.md), append the evaluation report to the existing `ideation.md`:

```markdown
## Evaluation Report

### Strengths
[Evidence-backed strengths from critical evaluation]

### Weaknesses
[Identified gaps, assumptions, and risks]

### First Principles Validation
[First principles breakdown per moai-foundation-thinking/modules/first-principles.md]

### Verdict
[Proceed / Proceed with caveats / Revisit / Abandon — with rationale]
```

---

## Phase 6: Proposal — SPEC Decomposition List

### Input

- `ideation.md` with Lean Canvas + Evaluation Report
- User's confirmation to proceed

### Process

Translate the converged product concept into actionable SPEC candidates. Each candidate represents a discrete, independently-implementable unit of work.

#### SPEC ID Naming Convention

[HARD] SPEC domain labels MUST be generic capability terms, never technology/language names:

| Correct (capability-based) | Wrong (technology-based) |
|---------------------------|--------------------------|
| `SPEC-AUTH-001`           | `SPEC-OAUTH2-001`        |
| `SPEC-API-001`            | `SPEC-FASTAPI-001`       |
| `SPEC-PIPELINE-001`       | `SPEC-AIRFLOW-001`       |
| `SPEC-UI-001`             | `SPEC-REACT-001`         |
| `SPEC-DB-001`             | `SPEC-POSTGRES-001`      |
| `SPEC-NOTIFY-001`         | `SPEC-FIREBASE-001`      |
| `SPEC-SEARCH-001`         | `SPEC-ELASTICSEARCH-001` |

#### Decomposition Heuristics

Suggest 2-10 SPEC candidates. Each candidate should:
1. Represent a cohesive capability boundary (not too granular, not too broad)
2. Be independently implementable without hard dependencies on sibling SPECs (except declared dependencies)
3. Represent 1-3 weeks of focused work (typical SPEC scope)
4. Address one of the Lean Canvas Solution blocks or a key infrastructure concern

If the idea is very small (single capability), 2-3 candidates is appropriate.
If the idea is large, suggest 7-10 candidates and note that ordering matters.

#### Edge Case: 0 or 1 candidates

If the idea seems too atomic for SPEC decomposition:
- 0 candidates: Add placeholder section: `### SPEC Decomposition Candidates` with note "Idea scope is atomic — consider direct /moai plan instead of /moai brain decomposition"
- 1 candidate: Acceptable, no special handling required

### Output

Write `proposal.md` to `.moai/brain/IDEA-NNN/`:

```markdown
# Proposal: {product name or concept label}
*Generated: {date} | Idea: IDEA-NNN*

## Product Summary

{2-3 sentence summary derived from Lean Canvas UVP + Solution blocks}

## Target User

{From Lean Canvas Customer Segments block}

## Core Problems Solved

{From Lean Canvas Problem block, formatted as numbered list}

## Proposed Solution

{From Lean Canvas Solution block — capabilities only, no tech stack}

## SPEC Decomposition Candidates

{2-10 bullets, each matching the canonical grammar below}

- SPEC-{DOMAIN}-001: {one-line scope description}
- SPEC-{DOMAIN}-002: {one-line scope description}
...

## Recommended Execution Order

{Numbered list of SPEC IDs in dependency order, with brief rationale}

## Out of Scope (v0.1)

{Explicit exclusions deferred to later SPECs or a v0.2 phase}

## Notes

{Any caveats, open questions, or assumptions from the evaluation}

```

### Grammar Invariant (ANCHOR)

<!-- @MX:ANCHOR: [AUTO] Canonical SPEC Decomposition Candidates bullet grammar -->
<!-- @MX:REASON: Consumed by /moai plan --from-brain parser (high fan_in: all brain-originated plan sessions). Changing this grammar breaks the parser silently. -->

The `### SPEC Decomposition Candidates` section MUST follow this exact grammar:

```
- SPEC-{DOMAIN}-{NUM}: {scope}
```

Where:
- `{DOMAIN}` is uppercase alphanumeric (e.g., `AUTH`, `API`, `UI`, `DB`, `NOTIFY`)
- `{NUM}` is zero-padded 3 digits (e.g., `001`, `002`, `010`)
- `{scope}` is a plain English one-line description (no backticks, no nested lists)
- One bullet per line, no sub-bullets

The `/moai plan --from-brain` parser uses this regex: `^- SPEC-[A-Z][A-Z0-9]+-[0-9]{3}: .+$`

Any bullet NOT matching this pattern is excluded from the suggestion list (surfaced as a warning, not an error).

---

## Works Well With

- `moai-foundation-thinking`: Diverge-Converge, Critical Evaluation, First Principles modules
- `moai-domain-research`: Feeds research.md content into Converge phase context
- `moai-domain-design-handoff`: Consumes proposal.md product summary for prompt.md context section
- `moai-workflow-brain`: Orchestrates this skill across phases 2, 4, 5 (append), and 6

---

## Common Rationalizations

| Rationalization | Reality |
|----------------|---------|
| "The user mentioned Python, so SPEC-PYTHON-001 is clearer" | Technology names in SPEC IDs create language lock-in. Use SPEC-API-001 — the technology choice happens at /moai plan time. |
| "The Solution block needs a tech stack to be concrete" | Solution describes WHAT the system does. HOW is deferred to architecture phase. "Processes 10K events/sec" is concrete without naming a framework. |
| "5 SPEC candidates is too few for a complex idea" | Start with 5-7 high-level candidates. /moai plan will decompose each one further if needed. |
| "I should skip the Lean Canvas for a simple idea" | Every brain invocation produces a Lean Canvas. The Customer Segments block alone is worth the exercise — it forces explicit user definition. |

## Verification

- [ ] Phase 2 diverge produced 5-15 distinct angles (not minor variations of the same angle)
- [ ] Phase 4 Lean Canvas has all 9 blocks present (none omitted or collapsed)
- [ ] Solution block contains no technology/framework names (capability-only language)
- [ ] Phase 6 proposal.md contains `### SPEC Decomposition Candidates` heading
- [ ] All SPEC candidates match grammar: `- SPEC-{DOMAIN}-{NUM}: {scope}`
- [ ] SPEC domain labels are capability-based (no technology names)
- [ ] proposal.md has no tech-stack assumptions outside of "Notes" section
