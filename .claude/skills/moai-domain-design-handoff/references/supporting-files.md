# Supporting Files Templates — references.md, acceptance.md, context.md, checklist.md

Verbatim templates for the 4 supporting files in the 5-file handoff package.

## Step 2: Assemble references.md

Populate from research.md's Sources section. Select 3-5 URLs that represent:

1. Existing competitors (to show what the user wants to improve on)
2. Design inspiration from adjacent products (visual quality reference)
3. User experience patterns relevant to the target use case

Format:

```markdown
# Design References

## Competitor Analysis

{URL}: {product name} — {what the design should improve on or learn from}

## Visual Inspiration

{URL}: {product name} — {specific visual quality to emulate: e.g., "clean typography", "card layout", "mobile-first nav"}

## UX Pattern References

{URL}: {product name or pattern} — {specific interaction pattern relevant to the design}
```

If research.md has fewer than 3 URLs (e.g., WebSearch failed), include a note:

```markdown
*Note: Limited references available due to research tool availability. Add 2-3 URLs of products you admire.*
```

## Step 3: Assemble acceptance.md

Design acceptance criteria derived from Lean Canvas + product type:

```markdown
# Design Acceptance Criteria

These criteria must be met for the design to be considered complete.

## Accessibility
- [ ] WCAG 2.1 AA compliance (minimum contrast ratio 4.5:1 for normal text)
- [ ] Interactive elements have visible focus states
- [ ] Alt text descriptions provided for all images and icons

## Responsiveness
- [ ] Mobile-first design (base breakpoint: 375px)
- [ ] Tablet layout defined (768px breakpoint)
- [ ] Desktop layout defined (1280px breakpoint)

## Brand Alignment
- [ ] Color palette consistent with brand voice section of prompt.md
- [ ] Typography consistent and readable
- [ ] Visual hierarchy reflects product priority (UVP communicated first)

## Content Completeness
- [ ] Hero section includes: headline, subheadline, primary CTA
- [ ] Core features visually communicated (minimum 3 features)
- [ ] Social proof element present (testimonial, stat, or logo row)

## Technical Constraints
- [ ] No animations or complex interactions in v1 (static design only)
- [ ] Design system uses reusable components (cards, buttons, inputs)
```

Customize the checklist based on the specific product type identified in proposal.md.

## Step 4: Assemble context.md

Extended context for the design session — NOT pasted into the prompt, kept as reference:

```markdown
# Extended Context: {product name}

> This file supplements prompt.md with additional context for your design session.
> It is NOT meant to be pasted into Claude Design — use prompt.md for that.

## Product Background

{Full Lean Canvas summary from ideation.md — all 9 blocks}

## SPEC Roadmap Context

{List of SPEC decomposition candidates from proposal.md — helps designer understand scope and what is out of scope for v1}

## Research Findings Summary

{Executive summary from research.md — key market insights and competitive dynamics}

## Brand Context

{If brand present: full brand-voice.md content}
{If brand absent: placeholder and instructions to populate .moai/project/brand/}
```

## Step 5: Assemble checklist.md

Human self-check before pasting prompt.md into claude.com Design:

```markdown
# Pre-Paste Checklist

Before pasting prompt.md into Claude Design, verify:

## Content Review
- [ ] Goal section accurately describes what you want designed
- [ ] References section has 3-5 URLs you have checked and are relevant
- [ ] Brand Voice section reflects your actual brand (not placeholder)
- [ ] Acceptance Criteria section reflects your real quality bar

## MoAI-Internal Cleanup (auto-verified)
- [ ] No SPEC- identifiers in prompt.md
- [ ] No .moai/ path references in prompt.md
- [ ] No /moai commands in prompt.md

## Scope Verification
- [ ] Out of Scope section lists things you explicitly do NOT want
- [ ] The "Goal" section is scoped to one page/view (not the entire product)

## Session Readiness
- [ ] You have a claude.com account with Design access
- [ ] You have reviewed IDEA-NNN/proposal.md and know what this design supports
- [ ] You are prepared to provide feedback on the generated design

After design is complete:
- Copy the Claude Design output to a local bundle directory
- Run: /moai design --path A --bundle <path-to-bundle>
```
