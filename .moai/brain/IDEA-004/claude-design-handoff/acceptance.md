# Design Acceptance Criteria

These criteria must be met for the design to be considered complete.

## Accessibility

- [ ] WCAG 2.1 AA compliance — minimum contrast ratio 4.5:1 for normal text, 3:1 for large text and graphical objects
- [ ] All interactive elements have visible focus states distinguishable from hover states
- [ ] Alt text provided for all informational images and icons; decorative images marked as such
- [ ] Heading hierarchy is sequential (no skipped levels), with one H1 per page
- [ ] Color is never the sole means of conveying information (e.g., safety-stack layers must be labeled, not just color-coded)
- [ ] Keyboard navigation works for sidebar, on-page table of contents, and code-block copy actions

## Responsiveness

- [ ] Mobile-first design with base breakpoint at 375px width
- [ ] Tablet layout defined at 768px breakpoint with sidebar visible on tablet+
- [ ] Desktop layout defined at 1280px breakpoint with three-column layout (sidebar | content | on-page TOC)
- [ ] Sidebar collapses to slide-in drawer triggered by hamburger icon on mobile
- [ ] No horizontal scroll at any breakpoint
- [ ] Touch targets at least 44×44 pixels on mobile

## Brand Alignment

- [ ] Color palette uses one accent color sparingly; neutral grayscale dominant
- [ ] No purple-to-blue gradients, no generic SaaS rainbow effects, no Inter-plus-purple cliché
- [ ] Typography pairing limited to two families maximum (one sans-serif + one monospace)
- [ ] Visual hierarchy reflects the principle-first product positioning (safety section is prominent, not tucked away)
- [ ] Tone of voice in design copy is direct and declarative; no marketing intensifiers

## Content Completeness

- [ ] Landing page hero includes: headline, one-sentence subheadline naming the differentiator, single primary CTA, no carousel
- [ ] Three-feature explainer with parallel structure (same bullet count, same heading depth across all three)
- [ ] Five-layer safety stack rendered as one continuous visual element (single diagram), not five disconnected cards
- [ ] Social proof block (testimonial, stat, or attribution line) present somewhere above the fold of the third screen
- [ ] Documentation page template includes: sidebar nav with collapsible sections, breadcrumb at top, content area with mermaid diagram support, on-page TOC on right (desktop only)
- [ ] Code blocks have syntax highlighting and a copy-to-clipboard affordance
- [ ] Footer includes: legal links, repository link, version indicator, locale switcher placeholder

## Technical Constraints

- [ ] Static design only for v1 — no animations on initial page load, no video backgrounds, no parallax
- [ ] Transitions limited to hover and focus state changes (≤200ms duration)
- [ ] Design system documents reusable components: button (primary/secondary/ghost), card, code-block, sidebar-link, breadcrumb, callout-box
- [ ] Mermaid diagram styling defined for both light mode and (optional) dark mode
- [ ] Light mode is required; dark mode is optional for v1 but if included must use system preference detection
- [ ] No external CDN dependencies for fonts in production build (fonts must be self-hostable)

## Information Architecture

- [ ] Documentation sidebar hierarchy maximum three levels deep
- [ ] Landing page sections in this order: hero → three-feature explainer → five-layer safety → social proof → final CTA → footer
- [ ] Documentation page on-page TOC visible only at desktop breakpoint (≥1024px)
- [ ] Search-affordance placeholder positioned in sidebar header (functionality not in scope for design, but the slot must exist)
