---
name: moai-domain-copywriting
description: >
  Brand-aligned copywriting specialist for marketing and product text.
  Enforces brand voice, anti-AI-slop rules, concrete numbers, and JSON section
  structure for downstream agent consumption. Covers hero, features,
  social_proof, cta, and footer sections with A/B variant output.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-04-20"
  tags: "copywriting, brand, marketing, voice, cta, headline, anti-slop"
  related-skills: "moai-domain-brand-design, moai-workflow-gan-loop"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["copy", "copywriting", "headline", "cta", "marketing", "microcopy", "tagline", "landing page copy", "value proposition", "brand voice"]
  agents: ["expert-frontend"]
  phases: ["run"]
---

# moai-domain-copywriting

Brand-aligned copywriting skill for marketing and product websites. Absorbed from agency-copywriting (v3.2.0). Enforces anti-AI-slop rules, requires brand voice context, and outputs structured JSON per section.

---

## Quick Reference

### Entry Conditions

Before generating copy, verify all three conditions are met:

1. Brand voice context is loaded: `.moai/project/brand/brand-voice.md` exists or is provided inline.
2. Target page or section scope is explicitly stated (landing page, about, pricing, etc.).
3. Anti-AI-slop checklist is active (see below).

If `brand-voice.md` does not exist, stop and instruct the user to run the brand interview via `/moai design` or `manager-spec`.

### Output Format

All copy output is structured JSON with the following top-level sections:

```
{
  "page_type": "<landing|about|services|pricing|contact>",
  "sections": {
    "hero": { "primary": {...}, "variant_a": {...} },
    "problem": { "primary": {...}, "variant_a": {...} },
    "solution": { "primary": {...}, "variant_a": {...} },
    "features": { "primary": {...}, "variant_a": {...} },
    "cta": { "primary": {...}, "variant_a": {...} },
    "pricing": { "primary": {...}, "variant_a": {...} }
  },
  "metadata": {
    "tone_profile": "<string>",
    "word_count": <number>,
    "reading_level": "<string>"
  }
}
```

Each section must include at least one A/B variant (`variant_a`). The `primary` field is the recommended default.

---

## Implementation Guide

### Section Structure

**hero**: The first visible block. Subject must be the reader outcome, not the company name.
- `headline`: Max 12 words. Present tense. Outcome-first.
- `subheadline`: One sentence. Expands the headline mechanism. Max 25 words.
- `cta_primary`: Verb + noun. Max 5 words. ("Start your free trial", "See how it works")
- `cta_secondary`: Optional. Softer alternative. ("Learn more", "Watch demo")

**problem**: Acknowledges the reader's pain point before proposing a solution.
- `headline`: Frame the cost of the status quo. Specific, not abstract.
- `body`: 2-3 sentences. Use second-person ("you", "your team").

**solution**: Introduces your product or service as the answer.
- `headline`: Direct claim with mechanism. ("We automate X so you can Y.")
- `body`: 3-4 sentences. State what it is, how it works, and one concrete outcome.

**features**: Scannable list of capabilities.
- Each feature: `title` (max 5 words) + `description` (max 2 sentences) + optional `metric`.
- Maximum 6 features per section.

**cta**: Final conversion prompt.
- `headline`: Urgency or clarity without manipulation. Avoid "limited time" cliches.
- `button_text`: Same rules as hero `cta_primary`.
- `supporting_text`: Trust signal (money-back guarantee, no credit card, etc.)

**pricing**: Only include if explicitly requested in scope.
- `headline`: Framing statement. ("Simple pricing. No surprises.")
- Plans array: `name`, `price`, `billing_period`, `highlights` (max 5 bullet strings).

---

### Anti-AI-Slop Checklist

Every piece of generated copy must pass this checklist before delivery:

**Forbidden patterns** — Reject any copy containing:
- Exclamation marks used more than once per section
- "In today's fast-paced world" or equivalent opening
- "Cutting-edge", "state-of-the-art", "revolutionary", "game-changing"
- "Leverage" as a verb (replace with "use", "apply", "deploy")
- "Seamless", "frictionless", "best-in-class" without quantified evidence
- Passive voice where active is grammatically equivalent
- Filler clauses: "at the end of the day", "the fact of the matter is"
- Generic benefits without concrete mechanism ("helps you succeed")
- Third-person self-reference in hero section ("Our company believes...")

**Required patterns** — Every section must have:
- At least one concrete number, percentage, or time reference
- Reader-centric framing (subject is "you" or reader outcome, not "we")
- Active voice for all capability claims
- Specificity: replace "grow your business" with a measurable outcome

**Tone calibration** — Load from `brand-voice.md`:
- Respect the tone spectrum defined (e.g., "playful" vs "authoritative")
- Match vocabulary preferences (avoid jargon if `jargon_level: low`)
- Preserve brand-specific terminology exactly as defined

---

### A/B Variant Rules

Each section must include one alternative version (`variant_a`) that differs in:
- Headline framing angle (problem-first vs. solution-first)
- CTA verb choice
- Tone register (slightly more or less formal)

Do not generate more than two variants unless the scope explicitly requires it.

---

### Brand Voice Integration

Load `.moai/project/brand/brand-voice.md` and apply:

1. `tone`: Overall register. If missing, default to "confident and direct".
2. `vocabulary_preferences`: Preferred and avoided terms. Enforce strictly.
3. `audience_familiarity`: Determines assumed knowledge level. Adjust jargon accordingly.
4. `example_phrases`: Use as stylistic anchors. Mimic sentence rhythm, not content.

If any `_TBD_` markers are present in `brand-voice.md`, stop and request completion before proceeding.

---

### Integration with moai-domain-brand-design

When both copywriting and design work are in scope (path B of `/moai design`):

1. Complete copy sections first and output structured JSON.
2. Pass the JSON copy output to `moai-domain-brand-design` as the content contract.
3. Design tokens and layout must accommodate copy length constraints (headline character counts, CTA button text length).

---

## Advanced Patterns

### Reading Level Targeting

Adjust copy complexity to the target audience's reading level:

- Technical audience (B2B SaaS developers): Grade 10-12. Allow domain terminology.
- Business audience (SMB owners): Grade 8-10. Avoid acronyms without expansion.
- Consumer audience (general public): Grade 6-8. Short sentences, concrete imagery.

Include the calculated reading level in the output metadata `reading_level` field using the Flesch-Kincaid grade level scale.

### Microcopy Guidelines

Form labels, error messages, and UI strings follow stricter rules:
- Labels: noun phrases, 1-3 words, no punctuation
- Placeholders: example values, not instructions ("jane@company.com" not "Enter email")
- Error messages: what went wrong + how to fix it, no blame language
- Success messages: confirm the action, state the next step

### Social Proof Integration

When testimonials or case study data are available in brand context:
- Quote exactly, do not paraphrase customer quotes
- Include company name, role, and metric (where permitted)
- Place after the problem section (reaffirms pain) or after features (reaffirms solution)

---

## Works Well With

- `moai-domain-brand-design`: Visual design must accommodate copy constraints
- `moai-workflow-gan-loop`: GAN loop evaluates copy quality in Design Quality and Completeness dimensions
- `expert-frontend`: Receives the JSON copy output for implementation
- `evaluator-active`: Scores copy accuracy against original `brand-voice.md`

---

Source: Absorbed from agency-copywriting v3.2.0 on 2026-04-20.
REQ coverage: REQ-SKILL-001, REQ-SKILL-002, REQ-SKILL-003
Version: 1.0.0
