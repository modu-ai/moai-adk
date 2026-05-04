# MoAI Developer Console

_Started 2026-05-04 04:25 UTC_

---

## User



## Assistant

_[tool: local_ls]_

I'll explore the design folder to understand the project context first.

_[tool: local_ls]_

_[tool: local_read]_

This looks like a complex project with extensive research and audits. Let me get a clearer picture before asking questions.

_[tool: local_ls]_

_[tool: local_ls]_

Let me skim the most relevant high-level docs to understand what MoAI v3 is.

_[tool: local_read]_

_[tool: local_read]_

_[tool: snip]_

I have a clear picture now. The user has an extensive design system (모두의AI / MoAI) plus deep research and audits for **MoAI-ADK v3** — a SPEC-governed development orchestrator. The project title is "MoAI-Web Dashboard". 

But the request is quite open-ended — there's no explicit ask. Let me ask scoping questions before diving in. The two attached skills (Hi-fi design + Interactive prototype) plus the design system suggest the user wants a high-fidelity, interactive web dashboard for MoAI, but I need to know exactly what surface.

_[tool: questions_v2]_

## User

Questions answered:
- what_dashboard: MoAI-ADK v3 developer console — SPEC/Plan/Run/Sync workflow, agents, hooks, harness state
- audience: 개발자
- fidelity: Multi-screen interactive prototype (tabs, modals, navigation flows)
- key_widgets: SPEC list & EARS requirement viewer, Plan/Run/Sync pipeline timeline (DAG), Sprint Contract / evaluator status, Agent roster + activity (manager/expert/builder), Hook & harness state monitor, TRUST 5 quality gates, Sandbox / permission / safety panel, Skill & rule library browser
- layout_vibe: 0
- theme: Both with toggle
- data_density: Balanced (modern SaaS density)
- interactions: Tab/route switching with smooth transitions, Live updating numbers / streaming logs, Filterable/sortable tables, Expandable detail panels (click row → side drawer), DAG / timeline animation, Modal forms (create SPEC, etc.), Just polished hover/press states is fine
- tweaks_interest: Light/dark theme toggle, Density switch (cozy / compact / dense), Layout variants (sidebar vs topbar, etc.), Live data on/off (seeded vs static), Mascot presence on/off
- mascot: Use it sparingly (empty states, onboarding hint)

## Assistant

Got it. Building the **MoAI-ADK v3 developer console** — multi-screen interactive prototype with sidebar layout, light+dark toggle, balanced density, and rich tweaks. Let me set up the project.

_[tool: update_todos]_

_[tool: list_files]_

_[tool: read_file]_

_[tool: list_files]_

_[tool: list_files]_

_[tool: list_files]_

_[tool: snip]_

Now copying needed assets and building the dashboard.

_[tool: copy_files]_

_[tool: copy_starter_component]_

Now let me build the dashboard. I'll create the main HTML, app shell, mock data, and feature screens. Given size constraints, I'll split into multiple JSX files.

_[tool: write_file]_

_[tool: snip]_

_[tool: write_file]_

_[tool: snip]_

_[tool: write_file]_

_[tool: write_file]_

<details>

_[tool: snip]_

<details>

_[tool: snip]_

_[tool: write_file]_

_[tool: write_file]_

_[tool: write_file]_

_[tool: done]_

<details>

The MoAI-ADK v3 Console prototype is ready. Highlights:

**7 fully wired screens** — Overview, SPEC, Pipeline, Agents, Hooks & Harness, Safety & Eval, Skills & Rules — connected via sidebar nav with badges + live indicator.

**Working interactions**
- ⌘K command palette (keyboard nav)
- ⌘N "New SPEC" modal with EARS template
- ⌘J theme toggle (light / dark)
- SPEC drawer with EARS / Contract / @MX tabs
- Agent drawer with frontmatter + recent calls
- Live-streaming pipeline log (toggle in Tweaks)

**Tweaks** — theme, density (cozy/compact/dense), layout (sidebar/topbar), live-log on/off, mascot toggle.

Built strictly on the 모두의AI design system: Pretendard + JetBrains Mono, deep-teal coreonly, signature gradient on hero/CTAs, `#09110f` ink (no pure black), Lucide icons.

_[tool: update_todos]_

_[tool: fork_verifier_agent]_

Open `MoAI-ADK Console.html` to explore. Try ⌘K for the command palette, click any SPEC card to open the detail drawer, and toggle the Tweaks panel for theme/density/layout variants.

