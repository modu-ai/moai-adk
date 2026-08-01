---
name: moai-ref-ui-polish
description: >
  UI polish and interface-completion reference: the small visual details — concentric
  border radius, optical alignment, shadow-vs-border, motion easing, typography
  smoothing, tabular numbers, icon stroke weight, hit areas — that separate polished
  interfaces from generic ones. Agent-extending skill that amplifies frontend/UI domain
  work with production-grade "interface taste" rules.
  NOT for: backend logic, database design, DevOps, security audits, non-UI work.

when_to_use: >
  Use for UI polish and design-completion work: building UI components, reviewing
  frontend code, implementing animations, hover/active states, shadows, borders,
  typography, icons, micro-interactions, enter/exit animations, or any visual detail
  work. Amplifies frontend domain work (manager-develop, Agent(general-purpose) with
  frontend instructions) with interface-design taste rules. Implementation examples are
  Web/CSS; the design principles are platform-neutral (apply to native mobile/desktop UI too).

user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-07-28"
  tags: "ui, polish, design, animation, typography, motion, frontend, reference"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000
---

# UI Polish Reference

## Target Agents

- `manager-develop` - Applies polish rules during frontend/UI component implementation (`cycle_type=tdd` or `cycle_type=ddd` context)
- `/moai review` - UI-design review surface; equivalently available as a per-spawn `Agent(general-purpose)` frontend specialist per `archived-agent-rejection.md` §C

## Core Philosophy

Great interfaces are a collection of small details that compound into a great experience. AI agents frequently miss these details — an `ease-in` easing on an enter animation (should be `ease-out`), a solid border where a semi-transparent shadow reads better, or mismatched radii on nested elements. None of these is catastrophic; together they separate "polished" from "generic".

Before suggesting polish changes, identify the project's existing styling system (design tokens, spacing scale, motion library). Never introduce a second styling system for polish fixes — extend the existing one.

## Geometry and Alignment

| Principle | Rule | Rationale |
|-----------|------|-----------|
| Concentric Border Radius | `outerRadius = innerRadius + padding` | Mismatched radii on nested elements is the most common cause of an interface "feeling off" |
| Optical over Geometric Alignment | When geometric centering looks wrong, align optically | Buttons with icons, play triangles, and asymmetric icons need a manual nudge; geometric center is visually off-center for these shapes |

## Elevation and Structure

| Element | Use | Avoid |
|---------|-----|-------|
| Semi-transparent layered `box-shadow` | Depth, elevation, floating surfaces | Solid borders for depth (they read heavy and flat) |
| Borders | Structure, dividers, separators, selected/focus state | Using shadows for structural separation (ambiguous) |

## Motion

| Pattern | Rule | Common Mistake |
|---------|------|----------------|
| Enter animation easing | `ease-out` (decelerate) — element arrives calmly | `ease-in` on enter (element appears to slam into place) |
| Exit animation easing | `ease-in` (accelerates away), mirroring the enter curve (small fixed `translateY`) | Full-height collapse, or harsher-than-enter motion |
| Interruptible state changes | CSS `transition` (can be interrupted mid-animation) | `keyframes` for interactive states (cannot interrupt) |
| Staged entrances | Stagger semantic chunks ~100ms — only for infrequent staged entrances | Staggering routine, high-frequency interactions (feels sluggish) |
| Contextual icon animation | `opacity`/`scale`/`blur` cross-fade (scale 0.25→1, opacity 0→1, blur 4px→0) | Toggling `visibility` (jarring, no transition) |
| Press feedback | `scale(0.96)` on click — always 0.96 | Smaller than 0.95 (reads as a bug, not a press) |
| First-render enter | Skip with `initial={false}` on `AnimatePresence` (or equivalent) | Enter animation fires on page load (disorienting) |
| Motion restraint | No custom animation on high-frequency interactions; motion is never the only feedback channel | Animating everything (noise, performance cost, accessibility) |

### Motion Easing Values (Web/CSS)

| Case | Value |
|------|-------|
| With motion library (Framer Motion et al.) | `transition: { type: "spring", duration: 0.3, bounce: 0 }` |
| Without motion library (CSS) | `cubic-bezier(0.2, 0, 0, 1)` for the standard "decelerate" curve |
| Never | `transition: all` — always specify exact properties (`transition-property: scale, opacity`) |

## Typography

| Rule | Implementation (Web/CSS) | When |
|------|--------------------------|------|
| Font smoothing | `-webkit-font-smoothing: antialiased` on root layout | macOS targets (sharper rendering) |
| Tabular numbers | `font-variant-numeric: tabular-nums` | Dynamically updating numbers (counters, timers, prices) — prevents layout shift |
| Heading wrapping | `text-wrap: balance` | Headings (prevents orphan words, evens line lengths) |
| Body wrapping | `text-wrap: pretty` | Body paragraphs (avoids orphan widows) |

## Imagery

| Rule | Value |
|------|-------|
| Image outline | `1px` subtle outline at low opacity |
| Outline color (light mode) | pure black — `oklch(0 0 0 / 0.1)` |
| Outline color (dark mode) | pure white — `oklch(1 0 0 / 0.1)` |
| Never | Tinted neutral outlines (read as a color choice, not a separation cue) |

## Interaction

| Rule | Value |
|------|-------|
| Minimum hit area (touch/mobile) | 44 × 44 px |
| Minimum hit area (dense desktop) | 40 × 40 px (extend with pseudo-element if the visual is smaller) |
| Hit area overlap | Never let hit areas overlap |
| `will-change` | Only `transform`, `opacity`, `filter` — and only when first-frame stutter is observed; never `will-change: all` |

## Icons

| Rule | Detail |
|------|--------|
| Stroke matches text weight | `1.5px` stroke beside regular (400) text; `2px` beside semibold (600) text |
| One stroke weight per set | Never mix icon libraries in one interface |
| State via color, not assets | Icons use `currentColor`; states come from CSS `color`/`opacity` — never separate asset files |
| Outline vs fill | Outline variant is default; fill variant marks the active state |

## Review Modes

| Mode | Coverage | Finding Cap |
|------|----------|-------------|
| `quick` | Primary user path, high-traffic states; HIGH/MEDIUM issues only | 5 |
| `full` | Entire scope across all categories (Typography, Surfaces, Motion, Icons, Interaction) | 15 |

### Severity

| Level | Meaning |
|-------|---------|
| HIGH | Makes interaction inaccessible, misleading, unreadable, or repeatedly disruptive |
| MEDIUM | Noticeable usability or consistency problem |
| LOW | Isolated polish issue (full mode only) |

### Verdict

| Verdict | Condition |
|---------|-----------|
| Block | Any HIGH finding remains |
| Needs changes | Only MEDIUM or LOW findings remain |
| Approve | No actionable findings remain |

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The framework handles animations correctly by default" | Framework defaults are generic. Easing, timing, and stagger must be specified per interaction, not inherited. |
| "A solid border is simpler than a layered shadow" | A solid border communicates structure; depth needs transparency. Using a border for elevation reads heavy and flat. |
| "Geometric centering is correct, optical adjustment is bikeshedding" | Geometric center is visually off for asymmetric shapes (icons, play triangles). Optical alignment is the difference users feel but cannot name. |
| "Motion is optional polish, the interface works without it" | Motion is feedback. Without it, state changes are ambiguous (did it register?). But motion must be restrained — animating everything is noise. |
| "Hit area equals visible area" | A 20px icon needs a 44px hit area on touch. Overlapping or undersized hit areas are the most common mobile usability defect. |
| "`transition: all` is convenient" | It transitions properties you did not intend (layout, color), causing unexpected animation. Always specify exact properties. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Enter animation uses `ease-in` (should be `ease-out`)
- Nested elements with mismatched border radii (no concentric relationship)
- Solid border used for depth/elevation instead of semi-transparent shadow
- `transition: all` instead of specific properties
- Hit area smaller than 40×40px, or overlapping hit areas
- Dynamically updating numbers without `tabular-nums` (layout shifts on each update)
- Icon stroke weight mismatched with adjacent text weight
- Separate icon asset files per state instead of `currentColor` recoloring
- Page-load enter animation fires without a skip (`initial={false}`)
- `will-change: all`, or `will-change` set permanently instead of only on stutter

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Enter/exit animations use `ease-out`; exits are softer than enters
- [ ] Nested elements follow `outerRadius = innerRadius + padding`
- [ ] Depth uses semi-transparent shadow; structure uses border
- [ ] Every `transition` specifies exact properties (no `transition: all`)
- [ ] Hit areas meet 44×44px (touch) / 40×40px (desktop) minimum, no overlap
- [ ] Dynamic numbers use `tabular-nums`; headings use `text-wrap: balance`
- [ ] Icons use one stroke weight matching text; states via `currentColor`, not separate assets
- [ ] First render skips enter animation (`initial={false}` or equivalent)
- [ ] High-frequency interactions have no custom animation; motion is not the only feedback

<!-- moai:evolvable-end -->
