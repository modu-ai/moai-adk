# Context — MoAI Cockpit Design Handoff

> Supplementary context for Claude Design. Use after reading prompt.md if more depth is needed.

---

## The Real Problem (research-backed)

Solo software developers using AI-assisted coding workflows lose **23 minutes 15 seconds** on average to recover focus after each interruption (Mark, UC Irvine). Each terminal command they type to check status (e.g., "is my pull request merged yet?", "are tests passing?", "what was I working on?") is a micro-interruption. Across an 8-hour day, RescueTime data suggests only 2 hours 48 minutes are actually focused work — the rest is context switching and recovery.

Tool-integration research suggests that **consolidating status queries into a single ambient surface can recover up to 30% of lost productivity.** This dashboard's primary success metric is exactly that: 30%+ reduction in terminal status-command roundtrips, measured by shell history grep before/after.

So this is not a "nice-to-have monitoring tool." It is a **deep-work preservation tool** disguised as a dashboard.

---

## Why a Dashboard, Specifically

Existing terminal tools (lazygit, gh-dash, k9s) are excellent within their own narrow scope but:
- They're modal — you switch into them and out of them, which is itself a context shift
- They cover one domain each (git, GitHub, Kubernetes) — not five
- They can't model project-specific concepts like "specifications", "phases", "quality gates"

Existing web dashboards are mostly hosted SaaS targeting teams of 10+, which carries unacceptable overhead for a solo developer (auth, accounts, sync, billing).

The unfilled niche: **localhost + single user + read-only + opinionated about a specific workflow methodology**. That is what this design serves.

---

## Visual Reference Anchors

Designs that inspire the right tone:
- **Vercel dashboard** — clean glanceable cards, calm chrome, status as data
- **Linear inbox** — dense information at glanceable scale, no decorative noise
- **Stripe dashboard** — trustworthy badges, semantic color use
- **Raycast extensions** — utility-first, minimal-but-warm

Designs that do NOT inspire the right tone:
- Marketing landing pages with hero animations
- "Engagement-optimized" social feeds with red dots and counts
- Overly skeumorphic dashboards (no glassmorphism, no neumorphism)
- Generic Bootstrap admin templates

---

## Key Domain Vocabulary (translate where helpful)

The user (solo developer) uses these terms internally:

| Internal term | What it means in plain language |
|---------------|--------------------------------|
| SPEC | A single project specification document — the unit of work |
| Phase | One of three project states: Plan (writing the spec), Run (implementing), Sync (documenting + creating the pull request) |
| Worktree | A parallel git working directory — one folder per branch the developer is actively working on |
| Memory | Notes the AI assistant kept across previous sessions (lessons learned, feedback, project state) |
| Drift | Subtle inconsistencies that accumulate (uncommitted changes, stale branches, unmerged work) |

In the dashboard, prefer **plain English labels** ("Project Spec", "Stage", "Working Directory", "Notes", "Loose Ends") over the internal jargon — but keep the internal codes (e.g., the spec ID itself) visible so the user can map back.

---

## Information Density Target

The dashboard packs a lot into one page. Calibrate density toward Linear/Stripe rather than Notion/Trello:
- ~120 characters readable per "row" without truncation
- Minimum 14px body text (16px preferred)
- 8px base spacing scale (multiples: 8, 16, 24, 32, 48, 64)
- Card padding ~16–24px, never less than 12px
- 1.5 line-height for body, 1.2 for badges and numerals

---

## Known Anti-patterns to Avoid

1. **"Notification anxiety design"** — red badges on every panel header drives frequent context checks. Use subtle indicators that reward calm reading.
2. **"Demo data smell"** — designs that look great with 12 perfectly-cropped fake PRs but break with 1 PR or 47 PRs. Show realistic edge cases.
3. **"Hover-to-discover"** — anything important must be visible without hovering. Hover is for tooltips and progressive disclosure of detail, not core content.
4. **"Branding the user's tool"** — this is a developer's personal cockpit, not a marketing surface. No watermark logos, no "Powered by..." footers, no marketing copy.
5. **"Lego brick uniformity"** — five identical-looking cards make scanning hard. The Workflow Tracker should visually dominate; subordinate panels should clearly nest under it in hierarchy.

---

## Design System Integration Note

The reference URL at the top of prompt.md (`https://api.anthropic.com/v1/design/h/-jhWdwlf2BjzXHPBHl4JxQ`) points to a design system that should drive color, typography, spacing, and component tokens. Within Claude Design:

1. Fetch and read the system first
2. Map its semantic tokens to the panel components below
3. If a token is missing for a need (e.g., a "warning" color is undefined), propose a token name and value, then use it consistently
4. If the design system implies a different layout philosophy than this brief, **prefer this brief's structural decisions** (5 panels, single page, read-only) and adapt the design system's surface treatment to fit

The brief defines structure; the design system defines surface.
