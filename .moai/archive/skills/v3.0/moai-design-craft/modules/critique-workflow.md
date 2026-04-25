# Design Critique Workflow

Post-build design critique is a structured review of what was actually built versus the original design intent. It is not a style review — it evaluates alignment between implementation and purpose.

## Trigger

Activated by `/moai review --critique`. Sync phase auto-trigger is planned for a future release.

## Three-Step Process

### Step 1: Observe (What is it doing?)

Set aside the SPEC and design direction temporarily. Look at the built interface as a user would encounter it:

- What does the layout communicate about priority?
- What is the user's eye drawn to first?
- What interactions are immediately available? Which require discovery?
- What does the empty state communicate?
- What does the error state communicate?

Write down observations as factual statements, not judgments. "The primary action is below the fold on mobile" — not "the layout is bad."

### Step 2: Diagnose (Where is the drift?)

Compare observations against `.moai/design/system.md` Design Direction:

- Which craft principles were upheld?
- Which craft principles were violated?
- Is vocabulary alignment maintained in labels and copy?
- Does the interaction contract match what was specified?

Classify each finding:
- **Cosmetic drift**: Visual deviation that does not affect user comprehension or task completion
- **Structural drift**: Layout or interaction pattern that conflicts with the mental model
- **Intent drift**: Feature does something different from the stated design intent

### Step 3: Decide (Patch or Rebuild?)

| Finding Type | Default Decision | Override Condition |
|-------------|-----------------|-------------------|
| Cosmetic drift | Patch | Multiple cosmetic issues converge into structural drift |
| Structural drift | Rebuild affected component | If isolated to one component with clear fix |
| Intent drift | Rebuild the flow | Never patch intent drift |

**Patch**: Minimal change to correct the specific violation. Do not touch adjacent code.

**Rebuild**: Scrap the implementation and return to the Design Direction statement. Build again from intent, using observations from the first build to avoid the same failure.

## Critique Report Format

```markdown
## Design Critique: [SPEC-ID or Feature Name]

### Observations
- [Factual observation 1]
- [Factual observation 2]

### Drift Analysis
- [COSMETIC] [Description] — Patch: [specific fix]
- [STRUCTURAL] [Description] — Rebuild: [component name]
- [INTENT] [Description] — Rebuild: [flow name]

### Decision
[Patch / Rebuild] — [one sentence justification]

### Design Direction Reminder
[Quote the relevant Design Direction from system.md]
```

## Hard Rules

Design quality gates that apply to all frontend output:

- No cards by default — use sections, columns, dividers, lists, and media blocks instead
- No cards in the hero — the hero is a visual anchor, not a container
- No boxed or center-column hero when brief calls for full bleed
- No more than one dominant idea per section
- No headline should overpower the brand on branded pages
- No filler copy — if deleting 30% of the copy improves the page, keep deleting
- No split-screen hero unless text sits on calm, unified side
- No more than two typefaces without clear reason
- No more than one accent color unless product has a strong system
- No default typography stacks (Inter, Roboto, Arial, system) — use expressive, purposeful fonts
- No flat, single-color backgrounds — use gradients, images, or subtle patterns
- No decorative gradients or abstract backgrounds as the main visual idea

### Hero Budget

The first viewport should contain only:
- The brand (hero-level signal)
- One headline
- One short supporting sentence
- One CTA group
- One dominant image

If a sticky/fixed header exists, it counts against the hero. Combined header + hero must fit within the initial viewport. When using 100vh/100svh heroes, subtract persistent UI chrome: calc(100svh - header-height).

### App UI Defaults

For dashboards, admin tools, and operational workspaces:
- Default to Linear-style restraint: calm surface hierarchy, strong typography, few colors
- Dense but readable information, minimal chrome
- Cards only when the card IS the interaction
- Organize around: primary workspace, navigation, secondary context/inspector, one accent for action/state

## Rejection Criteria

Reject frontends that exhibit:

- Generic SaaS card grid as first impression
- Beautiful image with weak brand presence
- Strong headline with no clear action
- Busy imagery behind text
- Sections repeating the same mood statement
- Carousels with no narrative purpose
- App UI made of stacked cards instead of layout
- Pill clusters, stat strips, icon rows, or boxed promos competing for attention
- Detached labels, floating badges, promo stickers on top of hero media

## What Critique Is Not

- Not a code review (that is `/moai review` without `--critique`)
- Not a design pattern extraction audit (that is `/moai review --design`; WCAG is handled by moai-domain-uiux)
- Not a performance review
- Not an opportunity to redesign — critique evaluates the built thing against stated intent, not against a better design that occurred to you afterward
