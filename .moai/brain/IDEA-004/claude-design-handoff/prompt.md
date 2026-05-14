# Design Brief: Self-Evolving Harness System Documentation Site

## 1. Goal

I need a complete visual design for a documentation and product landing experience for a self-evolving harness system aimed at developer tooling power-users.

Specifically, I need the landing page (above-the-fold hero, three-feature explainer, principle-based safety section, social proof block, footer) and one inner documentation page template (sidebar navigation, content area with mermaid diagrams and code blocks, on-page table of contents).

The design should communicate three value propositions:
- Autonomous improvement that the user can actually trust, grounded in explicit principles rather than opaque heuristics
- A safety architecture with five distinct layers, each visible and inspectable, gated by a human checkpoint before any change is applied
- A library of specialist skills that auto-organizes and compounds in capability over time

Target users: solo developer power-users running multiple long-lived coding projects, project tech leads managing two to five technical teams, and tooling maintainers who want their development environment to self-improve without micro-management. They are technical, value substance over marketing fluff, and judge products by depth of mechanism rather than surface polish.

---

## 2. References

For visual inspiration and style direction, please study these references:

- https://anthropic.com — restrained typography, generous whitespace, principle-first hierarchy. The way Anthropic communicates safety architecture is a direct stylistic model.
- https://langchain.com — technical depth without sacrificing approachability; sidebar-driven documentation layout with mermaid diagrams inline.
- https://www.cursor.com — product landing energy for developer tools; how to make autonomous capabilities feel trustworthy rather than threatening.
- https://docs.astral.sh/uv/ — modern documentation site aesthetic for Rust/Python tooling community; clean code blocks, minimal chrome, fast feel.
- https://linear.app — opinionated visual confidence; the design itself communicates that the team has strong opinions on quality.

Key aesthetic direction:
- Principle-grounded: every feature on the landing page should reference an explicit principle or invariant, not a vague benefit. The visual hierarchy should let principles breathe.
- Layered transparency: the five-layer safety stack must be visually expressible. A horizontal or vertical strata diagram should anchor that section.
- Technical confidence: monospace accents for identifiers, generous whitespace, restrained color (one accent color used sparingly, neutral grayscale dominant).

---

## 3. Brand Voice (default — please customize)

> NOTE: This project does not yet have a defined brand voice. The placeholders below
> are generic suggestions. Before using this prompt in Claude Design, either:
> (a) Edit this section with your actual brand voice, OR
> (b) Run a brand interview to define brand context

Brand personality: technical, principle-grounded, modest about claims and rigorous about safeguards. The voice respects the reader's intelligence and assumes domain familiarity.

Voice guidelines:
- Direct, declarative sentences. No marketing intensifiers ("revolutionary", "game-changing", "cutting-edge").
- Show, do not tell. Specifications, mechanisms, and gates are described concretely (e.g., "five-layer safety stack with a human checkpoint" not "robust safety").
- Identifiers and technical terms use monospace styling inline.
- Honest about limitations. Where mechanisms are aspirational or under validation, say so.

Color palette: neutral and modern. Dominant grayscale (white background, near-black text, mid-gray dividers). One restrained accent color (suggest a deep blue or muted teal). Avoid purple-to-blue gradients and generic SaaS rainbow effects.

Typography: clean sans-serif for body and headings (Inter, Geist, or DM Sans family). Monospace for inline identifiers and code blocks (JetBrains Mono or Fira Code).

---

## 4. Acceptance Criteria

The design must satisfy these non-negotiable requirements:

- WCAG 2.1 AA compliance — text contrast ratio at least 4.5:1 for normal text, 3:1 for large text. Interactive elements have visible focus rings.
- Responsive across three breakpoints: mobile (375px base), tablet (768px), desktop (1280px+). Sidebar collapses to slide-in drawer on mobile.
- The five-layer safety stack must be one continuous visual element (single diagram), not five disconnected feature cards. The layering itself communicates the design.
- Hero section includes a headline, a one-sentence subheadline that names the differentiator concretely, and a single primary call-to-action. No carousel.
- Three-feature explainer uses parallel structure (same number of bullets per feature, same heading depth, same visual rhythm).
- Code blocks in documentation pages support both syntax highlighting and a copy-to-clipboard affordance.
- Mermaid diagrams render cleanly in both light and dark mode (if dark mode is implemented; light mode is required, dark mode is optional for v1).
- No animations on initial v1. Transitions only on hover and focus states.

---

## 5. Out of Scope

Do not design:

- Pricing pages or pricing tiers (the product is open-source and free)
- Sign-up or authentication flows (no user account model exists)
- Marketing email templates or transactional email designs
- Mobile app screens (no mobile app)
- Onboarding wizards or interactive product tours
- Marketing campaign landing pages targeted at non-developer audiences
- Animated hero illustrations or video backgrounds (static design only in v1)
- Multiple competing call-to-action buttons in the hero (one primary CTA only)
