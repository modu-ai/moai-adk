# Acceptance Criteria — MoAI Cockpit Design Handoff

Pass/fail criteria for the design deliverables. Use as a checklist while reviewing Claude Design output.

---

## A. Structural Acceptance (must pass)

- [ ] All 5 panels (Workflow Tracker, Worktree Switchboard, CI/PR Glance Wall, Memory Surfboard, Drift Sentinel) are visible on a 1440×900 viewport without vertical scrolling
- [ ] Workflow Tracker is the visual headline (largest type, top of page, full width or dominant left)
- [ ] Panels 2–5 occupy a 2-column layout below the headline, with sensible aspect ratios
- [ ] Optional Quick Actions strip appears at the bottom and does not crowd content above
- [ ] At 768px width, panels stack vertically with Workflow Tracker first
- [ ] At 1920px+ width, content has a max-width cap and centered margins

## B. Read-Only Invariant (★ critical, no exceptions)

- [ ] Zero buttons in any panel mutate state (no "deploy", "merge", "fix", "sync now", "delete", "rerun")
- [ ] All interactive elements fall into exactly three categories: copy-to-clipboard chips, links opening in new tab, hover tooltips
- [ ] No element looks like a primary action button unless it is a copy chip or an external link

## C. Visual Polish (must pass)

- [ ] Color contrast meets WCAG 2.1 AA on all text and badge variants (verified contrast ratio ≥ 4.5:1 for normal text, ≥ 3:1 for large text and UI components)
- [ ] Status badges (success, warning, error, neutral) include both color AND icon AND text label — never color-only
- [ ] Typography hierarchy is readable: minimum 14px body, 1.5 line-height for prose, 1.2 for numerals
- [ ] Spacing follows a consistent base unit (8px scale: 8, 16, 24, 32, 48, 64 — no arbitrary 11px or 13px gaps)
- [ ] Light mode and dark mode are both designed (not dark-mode-as-color-inversion)
- [ ] Tokens from the referenced design system are applied consistently across all panels

## D. Information Density (must pass)

- [ ] Each panel renders gracefully at empty state (e.g., "no working directories yet" with subtle positive treatment)
- [ ] Each panel renders gracefully at high density (e.g., 10+ PRs, 8+ worktrees) without overflow chaos
- [ ] Each panel renders gracefully at error/loading state (subtle skeleton or muted placeholder, no full-page error screen)
- [ ] Truncation is consistent (ellipsis for inline text, line-clamp for multi-line, with hover tooltip showing full text)

## E. Refresh Behavior (visual treatment, not functional)

- [ ] A subtle "last refreshed Ns ago" indicator is present per panel (small, neutral, footer-aligned)
- [ ] No spinner or loading bar during refresh — content updates in place
- [ ] No browser notification, modal, or toast triggered by refreshes
- [ ] A "stale data" treatment exists for panels that haven't refreshed in 60+ seconds (subtle, not alarming)

## F. Interaction States (must pass)

- [ ] Hover state for rows in Worktree Switchboard and CI/PR Glance Wall (subtle background tint)
- [ ] Hover state for badges (tooltip explaining what the badge means)
- [ ] Focus state for all keyboard-tabbable elements (visible outline, WCAG 2.1 compliant)
- [ ] Active state for copy-to-clipboard chips (brief feedback that copy succeeded)

## G. Accessibility (must pass)

- [ ] All interactive elements are keyboard-reachable via tab order
- [ ] aria-live="polite" applied to refreshing regions (so screen readers don't announce every poll)
- [ ] Semantic HTML structure (h1 → h2 → h3 hierarchy, table for tabular data, nav for navigation)
- [ ] Color is never the sole carrier of meaning (always paired with icon + text label)

## H. Tone & Mood (subjective, must pass review)

- [ ] The page feels CALM (not noisy, not gamified, not engagement-optimized)
- [ ] A solo developer would be comfortable leaving this open in the corner of their screen for 8 hours
- [ ] The visual hierarchy answers "what am I working on?" within 3 seconds of glancing
- [ ] No marketing copy, watermarks, or "Powered by..." footers
- [ ] No red dots / unread badges that drive anxiety-checking

## I. Design Deliverables (artifacts produced)

- [ ] Full dashboard mockup at 1440×900 — light mode
- [ ] Full dashboard mockup at 1440×900 — dark mode
- [ ] Mobile/narrow stacking view at 768px
- [ ] Component spec sheet (one section per panel, with all states)
- [ ] Color & typography token sheet (mapped to semantic roles)
- [ ] Interaction state samples (hover, focus, active for the 3 interactive types)

## J. Optional Excellence (bonus, not required for pass)

- [ ] First-run empty state design (no specs yet, no PRs yet)
- [ ] Stale data treatment (60+ seconds since refresh)
- [ ] First-paint loading shimmer (subtle)
- [ ] Animated transitions between data updates (must be < 200ms, never blocking)

---

## Pass Threshold

The design **passes** if:
- All items in sections A, B, C, D, F, G are checked
- At least 80% of items in sections E, H, I are checked
- Section J is optional

If any item in section B (Read-Only Invariant) fails, the design must be regenerated regardless of other scores.
