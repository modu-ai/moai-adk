# Design Acceptance Criteria

These criteria must be met for the design to be considered complete.

## Accessibility

- [ ] WCAG 2.1 AA compliance: all text passes minimum 4.5:1 contrast ratio (normal text) and 3:1 (large text)
- [ ] All interactive elements have visible focus states
- [ ] Color is never the sole method of conveying information (e.g., habit completion state uses both color AND icon)

## Responsiveness

- [ ] Mobile-first design (base breakpoint: 375px)
- [ ] Tablet layout defined (768px breakpoint)
- [ ] Desktop layout defined (1280px breakpoint)
- [ ] Week-at-a-glance view visible without horizontal scrolling on 1280px desktop

## Brand Alignment

- [ ] Maximum 2 accent colors used across the entire design
- [ ] Dark mode is the primary variant with 90%+ of design decisions made in dark mode first
- [ ] Typography uses monospace for data/numbers and sans-serif for prose

## Content Completeness

- [ ] Dashboard view: Habit list, streak indicators, and weekly grid all visible in primary viewport
- [ ] Daily check-in: Single-tap/click habit completion (no confirmation modal required)
- [ ] Outcome indicator: Visual representation of habit-to-outcome correlation (even if simplified for v0.1)
- [ ] Empty state: Clear empty state design for new users with 0 habits

## Technical Constraints

- [ ] No animations or complex interactions in v1 design (static design only)
- [ ] Component reuse: buttons, cards, and input components should be visually consistent
- [ ] Design tokens: Colors and spacing should use a clear scale (4px base grid recommended)
- [ ] Mobile touch targets: minimum 44x44px for all interactive elements
