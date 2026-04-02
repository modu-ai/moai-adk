---
name: agency-copywriting
description: >
  Copy rules, tone, structure, and anti-patterns for AI Agency website content.
  Enforces brand voice consistency, section-level copy structure, and JSON output
  contracts for automated page generation workflows.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "3.2.0"
  category: "agency"
  status: "active"
  updated: "2026-04-02"
  evolution_count: "0"
  confidence_score: "0.70"
  base_context: "brand-voice.md"
  dependencies: "agency-design-system, agency-frontend-patterns"
  tags: "copywriting, brand-voice, tone, content-structure, agency, web-copy"
  related-skills: "agency-design-system, agency-frontend-patterns, agency-client-interview"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["copy", "copywriting", "brand voice", "tone", "headline", "tagline", "web copy", "content structure", "microcopy", "CTA", "value proposition"]
  agents: ["copywriter"]
  phases: ["plan", "run"]
---

# Agency Copywriting Skill

Governs all textual content produced for AI Agency websites. Ensures brand-aligned tone, structured section output, and elimination of common AI-generated copy anti-patterns.

---

## Static Zone

### Identity

**Purpose**: Define and enforce copy rules, tone guidelines, structural patterns, and anti-patterns for AI Agency website content generation.

**Input Contract**:
- Brand brief (from `.agency/context/brand-voice.md`)
- Target audience description (from `.agency/context/target-audience.md`)
- Page type (landing, about, services, case-study, blog, contact)

**Output Contract**:
JSON-structured copy per section with the following shape:
```json
{
  "page_type": "landing",
  "sections": [
    {
      "id": "hero",
      "headline": "string",
      "subheadline": "string",
      "body": "string",
      "cta_primary": "string",
      "cta_secondary": "string | null"
    }
  ],
  "metadata": {
    "tone_profile": "string",
    "word_count": "number",
    "reading_level": "string"
  }
}
```

**Owner**: copywriter agent

### Core Principles (Brand Context -- manual change only)

1. **Specificity Over Generality**: Every claim must be grounded in a concrete detail. Replace "we help businesses grow" with a specific mechanism or outcome.

2. **Active Voice, Present Tense**: Default to active constructions. Use present tense for capabilities, past tense only for case study results.

3. **Reader-Centric Framing**: The subject of every headline should be the reader or their outcome, not the agency. "You" before "We."

4. **Conversational Authority**: Tone sits between casual blog and formal whitepaper. Confident but not stiff. Knowledgeable but not jargon-heavy.

5. **Brevity as Respect**: Every word must earn its place. If a sentence can lose a clause without losing meaning, cut it.

6. **Evidence-Backed Claims**: Quantify where possible. "3x faster" over "significantly faster." Cite sources or case studies for performance claims.

7. **No Filler Phrases**: Eliminate "In today's fast-paced world", "leveraging cutting-edge", "at the end of the day", and similar empty constructions.

---

## Dynamic Zone

### Rules

_No rules discovered yet. Rules are added through the evolution process._

Format for future rules:
```
### R-001 (v1 | confidence: 0.XX | evidence: N)
[Rule description]
- Context: [When this rule applies]
- Evidence: [F-XXX references]
- Added: vX.Y.Z
```

### Anti-Patterns

_No anti-patterns discovered yet. Anti-patterns are added through the evolution process._

Format for future anti-patterns:
```
### AP-001 (confidence: 0.XX | evidence: N)
AVOID: [What to avoid]
- Why: [Reason]
- Evidence: [F-XXX references]
```

### Heuristics

_No heuristics discovered yet. Heuristics are added through the evolution process._

Format for future heuristics:
```
### H-001 (weight: 0.X | evidence: N)
[Soft guideline]
- Context: [When applicable]
- Note: [Caveats]
```

---

## Evolution Log

_No evolution entries yet._

Format for future entries:
```
### Evolution E-001
- Date: YYYY-MM-DD
- Type: rule_added | rule_updated | anti_pattern_added | heuristic_added
- Source: F-XXX (feedback reference)
- Change: [Description of what changed]
- Confidence delta: +/- 0.XX
```

---

Version: 3.2.0
Last Updated: 2026-04-02
