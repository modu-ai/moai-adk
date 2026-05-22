# prompt.md — 5-Section Template + Brand Branches

Canonical verbatim template for `prompt.md` Phase 7 assembly.

## 5-Section Template

`prompt.md` MUST follow this exact structure:

```markdown
# Design Brief: {product name from proposal.md}

## 1. Goal

{2-3 sentences describing what needs to be designed.}

I need a complete visual design for a {product description} — specifically the {scope: landing page / dashboard / mobile app / web app / etc.}.

The design should communicate: {top 3 value propositions from Lean Canvas UVP block}

Target users: {Customer Segments from Lean Canvas, 1-2 sentences}

---

## 2. References

For visual inspiration and style direction, please study these references:

{List of URLs from references.md — 3-5 URLs to existing products with brief style notes}

Key aesthetic direction:
- {style adjective 1}: {brief explanation}
- {style adjective 2}: {brief explanation}
- {style adjective 3}: {brief explanation}

---

## 3. Brand Voice

{EITHER brand_present branch OR brand_absent branch — see below}

---

## 4. Acceptance Criteria

The design MUST satisfy these non-negotiable requirements:

{Concise list from acceptance.md — typically 5-8 items}

---

## 5. Out of Scope

Do NOT design:

{Explicit exclusions — typically 3-5 items}
```

## Section 3 — Brand Voice Two Branches

### Branch A: Brand present (`brand_present = true`)

```markdown
## 3. Brand Voice

The brand personality is: {from brand-voice.md tone/personality fields}

Voice guidelines:
{Extract 3-5 most actionable brand voice rules from brand-voice.md}

Color palette (from brand identity):
{If visual-identity.md present: list primary colors with hex codes}
{If visual-identity.md absent: "Brand colors TBD — please use {tone}-appropriate palette"}

Typography:
{If visual-identity.md present: font families}
{If visual-identity.md absent: "Typography TBD — sans-serif for readability"}
```

### Branch B: Brand absent (`brand_present = false`)

```markdown
## 3. Brand Voice (default — please customize)

> NOTE: This project does not yet have a defined brand voice. The placeholders below
> are generic suggestions. Before using this prompt in Claude Design, either:
> (a) Edit this section with your actual brand voice, OR
> (b) Run /moai brain brand-interview (when available) to define brand context

Brand personality: professional, approachable, modern

Voice guidelines:
- Clear and concise language; no jargon
- Action-oriented CTAs
- Friendly but credible tone

Color palette: neutral, modern (grays, whites, one accent color — TBD)
Typography: clean sans-serif (Inter, Geist, or similar)
```
