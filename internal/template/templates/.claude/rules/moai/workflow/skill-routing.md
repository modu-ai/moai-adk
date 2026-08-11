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
| Render a data visualization (chart/dashboard) to HTML/SVG | `dataviz` |
| Author a design artifact hosted as a claude.ai web page (visual identity, landing page) | `artifact-design` |

Config coupling (report): the orchestrator reads `report.format` from the settings chain (`.moai/config/sections/report.yaml`, persisted via `internal/settings` — values `html+md` \| `md`) before rendering any report. When the format is `html+md` or `html`, an orchestrator-direct report request MUST route through `moai-domain-html-report`; when `md`, the orchestrator MUST NOT invoke the skill (markdown is the native output, the skill is idle).

[ZONE:Evolvable] [HARD] **Routing is intent-based, not keyword-based**: the orchestrator LLM reads skill descriptions semantically and matches the intent of a request against the described capability, so a single concise English intent statement suffices across all supported conversation languages (CLAUDE.md §2 — intent analysis is language-independent, never gated on English keyword matching). Multi-locale trigger-phrase enumeration in skill descriptions has no basis in this semantic-match mechanism and wastes the 1,536-char listing budget.

[ZONE:Evolvable] [HARD] **Anti-pattern (named)**: reaching for `artifact-design` when the task is a markdown→HTML **report render**. `artifact-design` calibrates visual identity for claude.ai-hosted web pages (landing pages, apps, shareable artifacts); `moai-domain-html-report` owns the report-render surface (six report modes, audience tiers, md-twin asymmetry). A report request that loads `artifact-design` instead of `moai-domain-html-report` is a routing miss. The corrective is intent-based: any request — in any language — whose intent is "produce a report/document as HTML" routes to `moai-domain-html-report`; `artifact-design` routes only "produce a hosted visual-identity page" intents.

## 2. Agent Obligation

Agents whose `tools:` include `Skill` load conditional skills per the "Conditional Skill Loading" section in their own body. The static `skills:` frontmatter preload stays at most 2 entries per agent (token diet — progressive disclosure does the rest). An agent loads a conditional skill when its body's stated trigger situation actually arises, not preemptively.

## 3. Rationale

The two loading mechanisms have different cost profiles:

- `skills:` frontmatter injects each listed skill's FULL body into the agent context at spawn — a fixed cost paid on every invocation, whether or not the skill is used.
- `Skill()` invocation loads on demand: only the ~100-token metadata line is always visible; the ~5K-token body is paid only when the skill is actually invoked.

Keeping the static preload minimal and routing the rest through explicit `Skill()` instructions converts a fixed per-spawn cost into a pay-per-use cost, while the orchestrator-side injection (section 1) preserves discoverability for domain skills the agent would not know to load.

## 4. Cross-references

- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors — cross-cutting agent obligations
- `.claude/rules/moai/development/agent-authoring.md` — agent frontmatter format (`skills:` YAML array, `tools:` CSV) and the Extension-Mechanism Context-Cost Ladder
- `CLAUDE.md` §4 — the retained agent catalog this protocol applies to
- `.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure — the 3-level token budget behind the on-demand cost profile

---

Classification: Evolvable operational rule — applies to all agent spawns and agent bodies with the Skill tool.
