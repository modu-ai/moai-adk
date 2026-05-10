# Design Brief: HabitScope — Outcome-Linked Habit Tracker

> NOTE: This is a worked example of the expected output of Phase 7 (Handoff).
> A real prompt.md is paste-ready for use in claude.com Design — no MoAI internals should appear.
> This example demonstrates correct structure: 5 sections, no tech-stack names, no internal references.

## 1. Goal

I need a complete visual design for a habit tracking web application — specifically the main dashboard and daily check-in interface.

The application helps knowledge workers connect their daily habits to measurable professional outcomes. The design should communicate:
- Progress visibility: Users see habit streaks and trend data at a glance
- Outcome correlation: The design makes visible whether tracked habits are producing results
- Calm productivity: The interface feels focused, not gamified — serious professionals use this tool

Target users: Remote software developers and knowledge workers, 25-40 years old, who care about professional development alongside personal habits. They value signal over noise and distrust flashy consumer apps.

---

## 2. References

For visual inspiration and style direction, please study these references:

- https://linear.app — Clean, high-information-density interface; excellent use of subtle color coding and typography hierarchy
- https://cal.com — Calm, professional booking tool; minimal chrome, generous whitespace
- https://notion.so — Document-first interface that prioritizes content; sidebar navigation model

Key aesthetic direction:
- **Minimal**: No decorative elements, no illustrations, no animations; the data IS the design
- **Dense but readable**: Fits a week's worth of habit data in the viewport without feeling crowded
- **Professional**: Muted color palette — dark mode primary, light mode secondary

---

## 3. Brand Voice (default — please customize)

> NOTE: This project does not yet have a defined brand voice. The placeholders below
> are generic suggestions. Before using this prompt in Claude Design, either:
> (a) Edit this section with your actual brand voice, OR
> (b) Run /moai brain brand-interview (when available) to define brand context

Brand personality: focused, evidence-driven, quietly confident

Voice guidelines:
- Concise labels; no motivational slogans
- Empty states use factual prompts ("No habits tracked this week") not cheerful encouragement
- Success states are understated ("7-day streak") not celebratory

Color palette: dark background (#0F0F0F) with single accent color (TBD — suggest teal or amber for habit completion states)
Typography: monospace for data/numbers (JetBrains Mono), sans-serif for prose (Inter or Geist)

---

## 4. Acceptance Criteria

The design MUST satisfy these non-negotiable requirements:

- [ ] WCAG 2.1 AA compliance (minimum 4.5:1 contrast ratio for all text)
- [ ] Mobile-first responsive: base breakpoint 375px, tablet 768px, desktop 1280px
- [ ] No more than 2 accent colors used in the entire design system
- [ ] Habit completion action requires exactly 1 tap/click (no modal confirmation)
- [ ] Week-at-a-glance view visible without scrolling on 1280px desktop
- [ ] Dark mode is the primary design variant; light mode is derived
- [ ] Typography scale: base 14px, max 24px (no oversized hero type)

---

## 5. Out of Scope

Do NOT design:
- Onboarding or signup flow (separate project)
- Mobile native app screens (web app only)
- Settings or account management pages
- Social features (team pods, sharing)
- Any animation or motion design
- Illustrations or photography
- Admin or analytics dashboards
