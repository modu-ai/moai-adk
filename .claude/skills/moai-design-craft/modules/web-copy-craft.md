# Web Copy Craft

Guidelines for writing website text (headlines, body copy, CTAs) that reads like a human copywriter wrote it — not an AI. These principles are language-neutral and apply to any locale.

## The Problem

AI-generated web copy follows predictable patterns that erode user trust:
- Uniform sentence length and rhythm
- Overuse of vague intensifiers ("innovative", "powerful", "seamless")
- Formulaic structure (problem → solution → call-to-action) without variation
- Translation-style phrasing that sounds unnatural in the target language

This module provides a framework to detect and eliminate these patterns.

## Principle: Concrete Over Abstract

Replace abstract adjectives with specific, verifiable facts.

| Weak (AI-typical) | Strong (human-typical) |
|---|---|
| A powerful analytics platform | Tracks 50M events per second |
| Innovative solution | First to support offline sync |
| Comprehensive guide | 47 tested recipes in 12 categories |
| Seamless integration | One line of code, zero config |
| Trusted by leading companies | Used by 3 of the top 5 banks |

**Rule**: If a word can be deleted without losing meaning, delete it. If it cannot be verified, replace it with a number or a name.

## Headline Formulas

Five headline structures that work across languages and cultures. Choose the structure that matches the section's purpose.

### 1. Number Anchor

Leads with a specific quantity. Creates immediate credibility.

```
Pattern: [Number] + [concrete outcome]
Example: "3 lines of code to deploy anywhere"
Example: "12 templates. Zero config."
```

Best for: Feature lists, tutorials, comparison sections.

### 2. Reversal

Contradicts the reader's assumption. Creates curiosity.

```
Pattern: [Expected behavior] + [unexpected outcome]
Example: "The more you optimize, the slower it gets"
Example: "Less data, better predictions"
```

Best for: Hero sections, blog headers, problem-framing sections.

### 3. Direct Question

Addresses the reader's actual situation. Creates recognition.

```
Pattern: [Situation the reader is in] + ?
Example: "Still deploying on Fridays?"
Example: "Spending more on tools than on people?"
```

Best for: Hero sections, email subject lines, CTA lead-ins.

### 4. Empathy Hook

Names the reader's pain without judgment. Creates connection.

```
Pattern: [Pain point they would say out loud]
Example: "When the dashboard says green but users say broken"
Example: "You shipped it. Nobody came."
```

Best for: Problem sections, testimonial intros, onboarding screens.

### 5. Declaration

States a position. Creates authority.

```
Pattern: [Opinionated statement about how things should be]
Example: "Monitoring should not require a PhD"
Example: "Your deploy pipeline is a product"
```

Best for: Brand positioning, hero taglines, about pages.

## Body Copy Rules

### Vary Sentence Rhythm

Monotonous rhythm is the strongest signal of AI-generated text. Break it deliberately.

```
Bad (uniform):
"This tool helps you monitor your systems. It provides real-time alerts.
It integrates with your existing stack. It scales automatically."

Good (varied):
"Real-time alerts. Integrates with what you already use.
And when traffic spikes — it just scales. No config, no panic."
```

**Technique**: Alternate between short fragments (under 5 words) and longer explanatory sentences (15-20 words). Never let three consecutive sentences share the same structure.

### Eliminate Filler Patterns

Common AI padding phrases to remove on sight:

- "In today's fast-paced world" — delete entirely
- "It's important to note that" — delete, start with the note itself
- "Let's dive into" — delete, present the content directly
- "Whether you're a beginner or expert" — delete unless the page genuinely serves both
- "Take your X to the next level" — replace with specific outcome
- "Unlock the full potential" — replace with what actually becomes possible
- "Embark on a journey" — delete

### End With Weight, Not Politeness

Do not close sections with:
- "Feel free to reach out" — weak
- "We hope this was helpful" — unnecessary
- "Don't hesitate to contact us" — redundant

Instead, end with:
- A fact that lingers
- A one-line summary
- A direct action ("Start building. It's free.")

## CTA Copy

### Rules

1. **Verb-first**: Start with an action verb ("Start", "Build", "Deploy", "Get")
2. **Outcome-oriented**: Name what the user gets, not what they do ("Get the report" not "Click here")
3. **One CTA per viewport**: Multiple competing CTAs reduce conversion
4. **Contrast the cost**: Pair the CTA with low commitment language ("Free", "No credit card", "2 minutes")

### CTA Patterns

| Purpose | Pattern | Example |
|---------|---------|---------|
| Primary conversion | [Verb] + [outcome] | "Start building" |
| Low commitment | [Verb] + [outcome] + [reassurance] | "Try free for 14 days" |
| Information request | [Verb] + [what they receive] | "Get the playbook" |
| Contact | [Verb] + [specificity] | "Talk to an engineer" |

## Anti-AI Checklist

After writing any web page copy, verify:

- No three consecutive sentences end with the same grammatical pattern
- No vague intensifier appears without a specific fact nearby ("powerful" must be followed by evidence)
- No filler opening phrase exists ("In today's...", "It's worth noting...")
- Headlines use one of the five formulas above (or a deliberate variation)
- CTA buttons start with a verb and name an outcome
- The page makes at least one specific, verifiable claim (number, name, or date)
- Closing text is a statement or action, not a courtesy phrase

## Integration With Design Craft

This module complements the Intent-First process:

- **Design Direction** (intent-first.md) establishes *why* the page exists and *who* reads it
- **Web Copy Craft** (this module) ensures the text matches that intent with human-quality writing
- **Design Critique** (critique-workflow.md) audits copy alongside visual elements

When expert-frontend or team-designer generates a web page, this module's rules apply to all visible text: hero headlines, section titles, feature descriptions, empty states, error messages, button labels, and footer copy.
