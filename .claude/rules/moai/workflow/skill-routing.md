---
description: "Conditional routing of domain skills into agent spawns"
paths: ".claude/agents/**,.claude/skills/**,.moai/config/sections/delegation.yaml"
---

# Skill Routing Protocol

Canonical rule for dynamic skill chaining: how the orchestrator routes domain skills into agent spawns, and how agents load conditional skills on demand.

## 1. Orchestrator Obligation

[ZONE:Evolvable] [HARD] Before spawning an implementation or review `Agent()`, the orchestrator MUST match the mission's domain against the available `moai-ref-*` / `moai-domain-*` skill descriptions and inject an explicit instruction into the spawn prompt for each matched skill (0-3 matches maximum):

```
At start, invoke Skill("<name>") for <reason>.
```

Examples:
- Backend API implementation → `At start, invoke Skill("moai-ref-api-patterns") for REST endpoint and error-handling conventions.`
- Security-sensitive review → `At start, invoke Skill("moai-ref-owasp-checklist") for the OWASP Top 10 review baseline.`
- React/Next.js work → `At start, invoke Skill("moai-ref-react-patterns") for component and state-management patterns.`

When no skill description matches the mission's domain, inject nothing — zero matches is a valid outcome, not a gap.

### §1.1 — Orchestrator-Direct Skill Routing (non-spawn)

[ZONE:Evolvable] [HARD] The §1 obligation binds agent **spawns**. When the orchestrator performs a task **directly** (no `Agent()` spawn) whose output shape matches a `moai-domain-*` skill, the orchestrator MUST load that skill via `Skill()` before producing the artifact. Discovery is mandatory when the task shape matches; the skill body is paid only on invocation (progressive disclosure), so there is no cost to loading it preemptively and failing to match.

| Orchestrator-direct task | Mandatory skill |
|---|---|
| Render a report / markdown → HTML artifact | `moai-domain-html-report` (mode by report type — status/incident/plan/explainer/financial/pr; audience tier from active output style) |
| Humanize / post-edit AI text (de-AI, 윤문) | `moai-domain-humanize` |
| Generate an SVG infographic / architecture diagram | `moai-domain-svg-infographic` |
| Reproduce or capture the look of an existing reference design (screenshot, image set, or URL) — extract it into a Design DNA profile, or generate an artifact from one | `moai-domain-design-dna` |
| Render a data visualization (chart/dashboard) to HTML/SVG | `dataviz` |
| Author a design artifact hosted as a claude.ai web page (visual identity, landing page) | `artifact-design` |

Config coupling (report): the orchestrator reads `report.format` from the settings chain (`.moai/config/sections/report.yaml`, persisted via `internal/settings` — values `html+md` | `md` | `artifact`, where `artifact` names the artifact-publication path) before rendering any report. When the format is `html+md` or `html`, an orchestrator-direct report request MUST route through `moai-domain-html-report`; when `md`, the orchestrator MUST NOT invoke the skill (markdown is the native output, the skill is idle); when `artifact`, the report still renders through `moai-domain-html-report` — under the artifact page contract — and its delivery step references `artifact-design`'s publication contract.

[ZONE:Evolvable] [HARD] **Routing is intent-based, not keyword-based**: the orchestrator LLM reads skill descriptions semantically and matches the intent of a request against the described capability, so a single concise English intent statement suffices across all supported conversation languages (.claude/skills/moai/SKILL.md — intent analysis is language-independent, never gated on English keyword matching). Multi-locale trigger-phrase enumeration in skill descriptions has no basis in this semantic-match mechanism and wastes the 1,536-char listing budget.

[ZONE:Evolvable] [HARD] **Anti-pattern (named), restated as the ownership split**: report **content and rendering** — the six report modes, audience tiers, md-twin asymmetry — are owned by `moai-domain-html-report`; the artifact **publication contract** is owned by `artifact-design` (which calibrates visual identity for claude.ai-hosted web pages — landing pages, apps, shareable artifacts). What remains a routing miss is loading `artifact-design` *instead of* `moai-domain-html-report` for a markdown→HTML **report render**; loading `artifact-design` at the `format=artifact` **delivery step** for the publication contract is correct routing, not a miss. The corrective is intent-based: any request — in any language — whose intent is "produce a report/document as HTML" routes to `moai-domain-html-report`; `artifact-design` routes "produce a hosted visual-identity page" intents and supplies the publication contract when a rendered report is published as an artifact.

## 4. Cross-references

- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors — cross-cutting agent obligations
- `.claude/rules/moai/development/agent-authoring.md` — agent frontmatter format (`skills:` YAML array, `tools:` CSV) and the Extension-Mechanism Context-Cost Ladder
- `.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure — the 3-level token budget behind the on-demand cost profile
- `skill-routing-detail.md` — the lazy companion. Load it for § Why the two loading mechanisms differ · § 2. Agent Obligation · § 3. Rationale (the per-mechanism cost profile).

---

Classification: Evolvable operational rule — applies to all agent spawns and agent bodies with the Skill tool.
