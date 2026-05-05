# Design Brief: MoAI Cockpit — Solo Developer Workflow Dashboard

> **Paste this entire file into Claude Design (claude.com).** It is self-contained and ready to use without modification. After pasting, also reference the additional context files in this same folder if needed.

---

## Reference Design System (fetch first inside Claude Design)

Before generating designs, fetch the referenced design system file and apply its tokens, components, and layout patterns:

```
Fetch this design file, read its readme, and implement the relevant aspects of the design.
https://api.anthropic.com/v1/design/h/-jhWdwlf2BjzXHPBHl4JxQ

Implement: incorporate the design system from the URL above into this project's UI, then layer the project-specific requirements below on top.
```

---

## What I'm Building

A **localhost-only, read-only, single-page web dashboard** that gives a solo software developer ambient awareness of their current work — without typing terminal commands.

The user runs one CLI command, a browser tab opens at `localhost:PORT/`, and that tab stays open in the corner of their monitor while they code in another window. The dashboard auto-refreshes via polling. The user never types into the dashboard — they only look at it.

Think: "the smartwatch on the developer's wrist that tells them everything important without making them stop coding."

---

## Who Uses It

A solo software developer working alone (no team, no collaborators) who:
- Spends most of their day in a code editor + terminal + browser combo
- Frequently runs the same 5–10 status-check commands (git, GitHub CLI, file listing) in the terminal — every few minutes — to track their own progress
- Loses focus every time they switch away from coding to type a status command
- Wants ambient information ("how is everything?") without active queries

Persona shorthand: **"the indie dev who pair-programs with an AI assistant."**

---

## The Single-Page Layout (5 panels + 1 utility strip)

Design a single full-screen dashboard page — no navigation, no second page, no modal. All five panels visible at once on a typical 1440×900 laptop screen (responsive down to 1280×720). On smaller screens, panels can stack vertically with the most important one (Workflow Tracker) at the top.

### Panel 1 — Workflow Tracker (top, full width, ★ most important)
What's happening right now. Shows:
- Currently active project specification ID (e.g., a short alphanumeric code)
- Current phase the user is in (one of: Plan / Run / Sync — three states)
- Last checkpoint or milestone description (one line of text from a progress file)
- A subtle "last updated" timestamp

Visual priority: this is the headline of the page. Largest type. Most attention.

### Panel 2 — Worktree Switchboard (left column, top)
Multiple parallel git working directories. Shows a compact table:
- Working directory name / branch name
- Last commit summary (1 line, truncated)
- Count of uncommitted file changes (badge with color coding: 0 = neutral, 1–5 = subtle, 6+ = warning)
- Age of last commit (relative: "2h ago", "3d ago")

3–8 rows typically. Each row is glance-readable.

### Panel 3 — CI/PR Glance Wall (right column, top)
Pull requests and their continuous integration status. Shows cards or a compact list:
- PR title (truncated)
- PR number + branch
- CI status badge (green = passing / yellow = pending / red = failing)
- Review status badge (approved / changes requested / awaiting review)
- "MERGED" marker for recently merged ones
- 1-line author/timestamp footer

3–10 PRs typically. Failing or blocking PRs should visually stand out without screaming.

### Panel 4 — Memory Surfboard (left column, bottom)
Session memory — the AI assistant's notes from previous work sessions. Shows a card list:
- Section header pulled from an index file
- 4–6 small cards: each card has a title (3–6 words), a one-line description (15–25 words), and a subtle tag (lessons / feedback / project)

Browseable, not interactive. Just scannable.

### Panel 5 — Drift Sentinel (right column, bottom)
Gentle warnings about things that are subtly wrong. Shows a list:
- Uncommitted changes that have sat for hours
- Working directories that haven't been touched in 14+ days
- Branches that were merged but not deleted

This is **not** a critical alert panel. The tone is "by the way, this might want attention." Empty state is celebrated ("All clear" with a subtle positive marker).

### Utility Strip — Quick Actions (bottom, slim, optional)
3–6 chips, each with a command name and a copy-to-clipboard button. The user clicks copy, then pastes into their terminal. The dashboard never executes anything itself.

Examples of chip labels: "Sync current branch", "Run quality gate", "Open project docs", "Show recent activity". Just labels and copy buttons — no other interaction.

---

## Visual & Interaction Constraints

### Tone & Mood
- **Calm, ambient, glanceable** — not noisy, not gamified, not "engagement-optimized"
- Low cognitive load — the user reads it for 3 seconds and looks away
- Status-as-information, not status-as-marketing
- A developer should be comfortable leaving this open in the corner of their screen for 8 hours

### Read-only Invariant (★ must hold across all designs)
- **Zero buttons that mutate state.** No "deploy", no "merge", no "fix", no "sync now"
- The only interactive elements are: link-style row clicks (open external thing in new tab), copy-to-clipboard chips, and a hover state for badges that explains what the badge means
- If a button looks like it could change something, it must instead either (a) copy a command to clipboard or (b) open an external link in a new tab

### Refresh Behavior
- The dashboard polls itself silently every 5–30 seconds depending on panel
- A subtle "last refreshed Ns ago" indicator per panel (small, neutral)
- No spinner during refresh — content updates in place
- No notifications, no popups, no sound

### Responsive
- Primary target: 1440×900 (laptop) at 100% zoom
- Functional minimum: 1280×720
- Below 1280px wide: panels stack vertically; Workflow Tracker first, then alternate Worktree → PR → Memory → Drift in a single column
- Above 1920px: max content width capped (e.g., 1800px) with centered margins

### Color & Typography
- Primary palette and typography: take from the referenced design system fetched at the top of this prompt
- If the referenced system is unavailable: default to a neutral developer-tool aesthetic (light or dark mode supported), monospace for code/IDs, sans-serif for prose
- Status colors must be color-blind-friendly (never rely on red-vs-green alone — use icon + text)
- Dark mode: support via system preference (prefers-color-scheme)

### Accessibility
- WCAG 2.1 AA contrast ratios
- All status badges include a textual label (not color-only)
- Keyboard navigable (tab through chips and rows)
- Screen reader friendly (semantic HTML, aria-live polite for refreshing regions)

---

## What This Is NOT

To prevent scope creep, the dashboard does NOT:
- Edit any files
- Execute any commands itself
- Show real-time logs streaming live (no console output panel)
- Display large code diffs (no diff viewer)
- Provide search across panels
- Allow customizing layout in v1
- Show notifications outside the page (no browser notifications)
- Connect to remote hosts (localhost only)
- Track multiple users (solo only)

These are explicit non-goals for the first release.

---

## Design Deliverables Requested

Please produce:

1. **Full dashboard mockup** — 1440×900 desktop view, light mode and dark mode
2. **Mobile/narrow stacking view** — 768px width showing the vertical fallback
3. **Component spec sheet** — each panel as a reusable component with: header pattern, row/card pattern, badge variants, empty state, error state, loading state
4. **Color & typography token sheet** — extracted from the referenced design system, mapped to semantic roles (primary text, secondary text, success badge, warning badge, error badge, surface, surface elevated, border, divider, etc.)
5. **Interaction states** — hover, focus, active for the only three interactive element types (row click, copy chip, badge tooltip)

---

## Optional Refinements (only if time permits)

- An empty-state design for first-run (no project specs yet, no PRs yet)
- A "stale data" treatment when a panel hasn't refreshed in 60+ seconds
- A subtle loading shimmer for the very first paint

---

## Acceptance Snapshot

The design is "done" when:
- All 5 panels render at 1440×900 without scrolling
- Every interactive element either copies-to-clipboard or opens-in-new-tab — never mutates
- A solo developer can look at the page for 3 seconds and answer: "what am I working on, and what needs my attention?"
- Color contrast and typography meet WCAG 2.1 AA
- Dark mode and light mode are both visually polished, not dark-mode-as-afterthought

Detailed acceptance criteria are in the companion `acceptance.md` file in this same folder.
