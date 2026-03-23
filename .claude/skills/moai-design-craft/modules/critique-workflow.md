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

## What Critique Is Not

- Not a code review (that is `/moai review` without `--critique`)
- Not a design pattern extraction audit (that is `/moai review --design`; WCAG is handled by moai-domain-uiux)
- Not a performance review
- Not an opportunity to redesign — critique evaluates the built thing against stated intent, not against a better design that occurred to you afterward
