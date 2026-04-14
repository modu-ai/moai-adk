---
name: agency-client-interview
description: >
  Client discovery interview framework for AI Agency projects covering
  business objectives, audience analysis, brand context gathering, and
  structured question flows that adapt to user patterns over time.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "agency"
  status: "active"
  updated: "2026-04-02"
  evolution_count: "0"
  confidence_score: "0.00"
  base_context: "target-audience.md"
  dependencies: ""
  tags: "interview, discovery, context-gathering, client, onboarding, brief"
  related-skills: "agency-copywriting"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["interview", "brief", "discovery", "onboarding", "context", "client", "requirements"]
  agents: ["planner", "copywriter"]
  phases: ["plan"]
---

# Agency Client Interview Skill

Governs the client discovery process for AI Agency projects. Provides structured question flows, progressive disclosure of interview depth, and context file generation for `.agency/context/`.

---

## Static Zone

### Identity

**Purpose**: Define interview question frameworks, context gathering strategies, and structured output formats for populating `.agency/context/` files during project onboarding.

**Input Contract**:
- User request (natural language description of desired project)
- Existing `.agency/context/` files (may be empty on first run)

**Output Contract**:
- Populated `.agency/context/` files:
  - `brand-voice.md` — tone, style, language preferences
  - `target-audience.md` — customer personas, pain points
  - `visual-identity.md` — colors, fonts, imagery preferences
  - `tech-preferences.md` — framework, hosting, integrations
  - `quality-standards.md` — performance targets, accessibility level

**Owner**: `.agency/context/target-audience.md`

### Core Principles

> Derived from Brand Context. Never auto-modified. Manual editing only.

1. Empathetic questioning — understand the WHY before the WHAT
2. Progressive disclosure — start broad, drill down based on answers
3. Business-first — always connect technical decisions to business outcomes
4. No assumptions — ask, don't guess (especially for brand voice and audience)
5. Respect user's time — skip questions already answered in context files

### Default Interview Flow

**Phase 1: Business Context (3 questions)**
1. What is the primary business objective for this project?
2. Who is your target customer? (age, role, pain points)
3. What does success look like? (specific KPIs if possible)

**Phase 2: Brand Identity (3 questions)**
4. How would you describe your brand voice? (3-5 adjectives)
5. Are there reference sites or brands you admire?
6. Any design preferences or constraints? (colors, style, existing brand guidelines)

**Phase 3: Technical Scope (2 questions)**
7. What pages/sections do you need?
8. Any technical requirements? (framework, hosting, integrations)

**Phase 4: Quality Expectations (1 question)**
9. What matters most to you? (speed, design quality, mobile experience, SEO)

### Adaptive Behavior

- If `.agency/context/` files already have content, skip covered questions
- If user has completed 5+ projects, reduce to 3 key questions only
- Learn which questions users skip or answer briefly → reduce priority

---

## Dynamic Zone

> Interview questions, order, and depth evolve via user feedback.

### Rules

(No rules yet. Rules will be added as the system learns from user feedback.)

### Anti-Patterns

(No anti-patterns yet.)

### Heuristics

(No heuristics yet.)

---

## Evolution Log

- v1.0.0: Initial creation (Static Zone with default interview flow, empty Dynamic Zone)

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I know what the client wants from the initial message" | Initial messages are surface-level. Discovery interviews surface hidden requirements, constraints, and priorities. |
| "Business objectives are obvious for this type of project" | Assumed objectives lead to misaligned deliverables. Every client has unique success metrics. |
| "Brand context can be filled in later" | Brand context flows to every pipeline phase. Missing brand context produces generic, brandless output. |
| "The audience is everyone" | Universal audience means no audience. Defining segments enables targeted copy, design, and feature prioritization. |
| "I will ask all the questions at once to save time" | Question overload produces shallow answers. Structured flows with follow-ups surface deeper insights. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- BRIEF created without conducting a discovery interview
- Business objectives section contains generic statements ("increase engagement")
- Target audience described without specific segments or persona details
- Brand voice and visual identity sections left empty or placeholder
- Interview ended without confirming findings with the client

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Discovery interview conducted with structured question flow
- [ ] Business objectives are specific and measurable (not generic aspirations)
- [ ] At least one target audience segment defined with demographics and needs
- [ ] Brand voice, tone, and visual identity captured in `.agency/context/`
- [ ] Interview findings confirmed with client before proceeding to pipeline

<!-- moai:evolvable-end -->
