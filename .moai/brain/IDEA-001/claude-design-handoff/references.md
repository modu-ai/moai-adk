# References — MoAI Cockpit Design Handoff

External references that informed the brief. Optional reading for the designer.

---

## User-Provided Design System (★ primary visual reference)

- [Claude Design Asset (referenced in prompt.md)](https://api.anthropic.com/v1/design/h/-jhWdwlf2BjzXHPBHl4JxQ) — fetch inside Claude Design and apply its tokens / components / layout patterns

---

## Tutorials & Templates (Implementation Reference, not design)

These influence the technical layer (polling cadence, partial updates) which constrains design decisions like "what's a refresh moment look like":

- [go-templ-htmx-example-app (alekLukanen)](https://github.com/alekLukanen/go-templ-htmx-example-app)
- [gohtmxapp (josephspurrier)](https://github.com/josephspurrier/gohtmxapp) — closest reference for a `localhost:8080/dashboard` route
- [go-htmx-data-dashboard (webdevfuel)](https://github.com/webdevfuel/go-htmx-data-dashboard) — data-heavy dashboard pattern
- [go-htmx-template (Piszmog)](https://github.com/Piszmog/go-htmx-template) — most modern starter (Go 1.25+)
- [Tailbits — Setting up HTMX and Templ for Go](https://tailbits.com/blog/setting-up-htmx-and-templ-for-go)
- [Go-Blueprint Docs — HTMX and Templ](https://docs.go-blueprint.dev/advanced-flag/htmx-templ/)

## Polling & Partial Update Patterns

- [HTMX hx-trigger every (polling)](https://github.com/bigskysoftware/htmx/blob/master/www/content/attributes/hx-trigger.md)
- [HTMX Hypermedia APIs vs Data APIs (table polling)](https://github.com/bigskysoftware/htmx/blob/master/www/content/essays/hypermedia-apis-vs-data-apis.md)
- [HTMX progress bar polling pattern](https://github.com/bigskysoftware/htmx/blob/master/www/content/essays/paris-2024-olympics-htmx-network-automation.md)
- [a-h/templ Fragments for partial updates](https://github.com/a-h/templ/blob/main/docs/docs/03-syntax-and-usage/19-fragments.md)
- [a-h/templ Layout & forms](https://github.com/a-h/templ/blob/main/docs/docs/03-syntax-and-usage/11-forms.md)

## Competitive / Comparable Tools (for tone calibration)

- [gh-dash — terminal GitHub dashboard](https://www.gh-dash.dev/)
- [gh-dash on Terminal Trove](https://terminaltrove.com/gh-dash/)
- [lazygit — terminal git UI](https://github.com/jesseduffield/lazygit)
- [gh-pr-dashboard (devactivity.com)](https://devactivity.com/insights/boost-developer-productivity-with-gh-pr-dashboard-your-unified-pr-command-center/)
- [Best Terminal Tools for Developers in 2026 (DEV)](https://dev.to/raxxostudios/best-terminal-tools-for-developers-in-2026-4jn1)
- [The Best TUI Apps for Linux Developers 2026](https://www.thetechbasket.com/best-tui-apps/)

## Productivity Research (justifies the read-only single-page constraint)

- [Mitigating Context Switching in Software Development (Jellyfish)](https://jellyfish.co/library/developer-productivity/context-switching/)
- [Reducing context switching in development workflows (Graphite)](https://graphite.com/guides/reducing-context-switching-development-workflows)
- [The Hidden Cost of Context Switching for Developers (Crownest)](https://www.crownest.dev/blog/hidden-cost-context-switching-developers)
- [Context Switching: The Silent Killer of Developer Productivity (Hatica)](https://www.hatica.io/blog/context-switching-killing-developer-productivity/)
- [The True Cost of Context Switching in Developer Workflows (Axolo)](https://axolo.co/blog/p/cost-context-switching-developer-workflow)
- [Context Switching is Killing Your Productivity (Software.com)](https://www.software.com/devops-guides/context-switching)
- [Context switching: How to reduce productivity killers (Atlassian)](https://www.atlassian.com/work-management/project-management/context-switching)
- [Reduce Context Switching | Developer Productivity (GitScrum Docs)](https://docs.gitscrum.com/en/best-practices/minimize-context-switching-developer-tools/)
- [The Context-Switching Problem: Why I Built a Tracker That Lives in My Terminal (DEV)](https://dev.to/tejas1233/the-context-switching-problem-why-i-built-a-tracker-that-lives-in-my-terminal-4dpe)

## Visual Tone Anchors (subjective, brief-author selections)

- Vercel dashboard
- Linear inbox
- Stripe dashboard
- Raycast extensions
