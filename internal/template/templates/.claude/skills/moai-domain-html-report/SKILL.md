---
name: moai-domain-html-report
description: >
  Markdown-to-single-file-HTML report renderer. Six modes (status, incident,
  plan, explainer, financial, pr) selected by report type, crossed with three
  audience tiers (expert, basic, learn) derived from the active output style.
  The basic and learn tiers enrich the HTML with mermaid flowcharts, worked
  examples, and plain-language primers; the expert tier stays dense. Zero
  external JS/CSS framework dependencies — inline SVG charts, a font-CDN
  exception for Korean readability, and a tier-gated mermaid-CDN exception.
  Self-contained output for email attachment, print, and offline viewing.

when_to_use: >
  Use when a markdown report must be rendered into a single self-contained
  HTML file. Trigger phrases include "render this report as HTML", "weekly
  status report as one HTML file", "convert the financial statements to an
  HTML report", "incident report as HTML", "printable business plan HTML",
  "email-ready HTML report", and "explain this as an HTML report with
  diagrams".

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: true
metadata:
  version: "1.1.0"
  category: "domain"
  status: "active"
---

# html-report — Single-File HTML Report Renderer

## Purpose and Scope

This skill is a terminal renderer that converts a markdown report into a single self-contained HTML file. It accepts any markdown body produced by a text, analysis, or reporting workflow and emits one `.html` file that opens directly in a browser, attaches to email, prints cleanly, and works offline.

**Core principles**:

- Zero external JS libraries (no Chart.js, D3, htmx)
- Zero external CSS frameworks (no Tailwind, Bootstrap)
- Inline SVG renders all charts directly
- A font-CDN `<link>` is permitted for Korean readability
- A mermaid-CDN `<script>` is permitted **only in the `basic` and `learn` audience tiers**, always paired with a no-JS fallback (see § Diagram Policy). The `expert` tier remains strictly zero-JS.

**This skill does not replace the markdown output.** Markdown remains the single source of truth; HTML rendering is an additional branch that operates on it.

### The asymmetry principle — HTML is rich, the markdown twin is lean

The two artifacts this skill produces serve **different readers and therefore carry different amounts of content**. They are not the same document in two syntaxes:

| Artifact | Reader | Content rule |
|----------|--------|--------------|
| `.html` | **the human** | **Enriched.** May carry MORE than the source markdown — plain-language primers, mermaid diagrams, worked examples, analogies, glossary callouts — scaled by the audience tier (§ Audience Tiers). |
| `.md` twin | **the agent** (context) | **Lean.** Carries ONLY the load-bearing facts: findings, decisions, numbers, tables, action items. Never the tier enrichment. |

[HARD] **Audience-tier enrichment scales the HTML and NEVER the markdown twin.** Raising the tier from `expert` to `learn` must not add a single primer, analogy, or worked example to the `.md` twin — that enrichment exists to teach a human, and it is pure token cost to an agent that already understands the domain. A `learn`-tier report and an `expert`-tier report of the same source produce **markdown twins of substantially the same size**; only their HTML differs.

---

## Input

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `markdown` | yes | — | The markdown body to convert |
| `mode` | yes | — | `status` \| `incident` \| `plan` \| `explainer` \| `financial` \| `pr` |
| `audience` | no | derived from the active output style | `expert` \| `basic` \| `learn` — see § Audience Tiers |
| `slug` | no | auto-derived from the title | Output filename prefix |
| `output_path` | no | `<cwd>/reports/<slug>-<YYYYMMDD>.html` | Output path |
| `font_stack` | no | per-mode default | Font mapping override |

`mode` and `audience` are **orthogonal**: `mode` picks the report's *structure* (which sections exist), `audience` picks its *depth* (how much explanation each section carries). Every mode renders at every tier.

---

## Output

Two files at `<cwd>/reports/<slug>-<YYYYMMDD>.{html,md}`:

**The `.html` file** — the human-facing artifact:

- Size: ≤ 50KB at the `expert` tier; ≤ 120KB at the `basic` / `learn` tiers (the enrichment budget — diagrams and examples cost bytes)
- External dependencies: one font-CDN `<link>` + two `preconnect` hints (Korean fonts), plus one mermaid-CDN `<script>` at the `basic` / `learn` tiers only
- Self-contained: opens directly in a browser, email-attachable, print-clean, and readable offline (diagrams degrade to their fallback — see § Diagram Policy)

**The `.md` twin** — the agent-facing artifact (below).

### Markdown twin (agent-context artifact)

Alongside every `.html` file, write a **markdown twin** at the same path with the `.md` extension (`<slug>-<YYYYMMDD>.md`). The HTML file is the human-viewing artifact; the markdown twin is the machine-context artifact, and per § The asymmetry principle it is deliberately **leaner than the HTML**, not merely the same content with tags stripped.

**What the twin contains** — the load-bearing facts only:

- The findings, decisions, numbers, and conclusions
- Tables (as markdown tables) and any figures the numbers depend on
- Action items, owners, and open questions
- The mermaid source of a diagram **only when the diagram encodes information the prose does not** (a real state machine, a real dependency graph). A diagram that merely re-illustrates a sentence for a beginner is enrichment — it is omitted.

**What the twin OMITS** — everything the audience tier added for the human:

- Plain-language primers and jargon glossaries
- Analogies and motivating narratives
- Worked step-by-step examples that re-derive a stated result
- Self-check questions, callout boxes, decorative diagrams
- All HTML tags, inline CSS, `<script>` blocks, and SVG chart markup

**Consumption rule (token discipline)**: whenever a report is needed as context — an `Agent()` spawn prompt, a follow-up analysis turn, a cross-session Read of a past report — use the `.md` twin, NEVER the `.html` file. Raw HTML wastes tokens on tags, style blocks, and SVG paths that carry no information the markdown does not already have (typically 3-5x the tokens for identical content), and the tier enrichment on top of that is pure cost to an agent.

**Legacy HTML without a twin**: when only an `.html` file exists, extract the load-bearing facts into markdown first (strip tags, `<style>`, `<script>`, SVG chart markup, and the tier enrichment; convert `<table>` to markdown tables) and inject the extraction — not the raw HTML — into the agent prompt or context. Write the extraction next to the HTML as its `.md` twin so the cost is paid once.

---

## After rendering — report back to the user

Once the `.html` file and its `.md` twin are written, the response MUST do two things:

1. **Summary** — print a concise summary of what was rendered: the mode, the **audience tier** (and what it was derived from — the active output style, or an explicit `audience` argument), the report title, and the key sections or figures the file contains (a short paragraph or a few bullets). Do not paste the full HTML into the response.
2. **Auto-open** — immediately open the rendered file in the user's default browser by running the platform-appropriate opener via the Bash tool. Do NOT ask the user to type `! open` themselves; run the opener directly so the report appears in one step on macOS, Windows, and Linux alike:

   ```bash
   case "$(uname -s)" in
     Darwin) open "<output_path>" ;;
     Linux)  xdg-open "<output_path>" >/dev/null 2>&1 || echo "Open manually: <output_path>" ;;
     MINGW*|MSYS*|CYGWIN*) start "" "<output_path>" ;;
     *) echo "Open manually: <output_path>" ;;
   esac
   ```

   macOS uses `open`, Linux uses `xdg-open` (fall back to printing the absolute path when no opener/display is available — headless or WSL environments), Windows Git-Bash/MSYS uses `start`. If the opener command fails or the permission is denied, print the absolute path so the user can open the file manually.

Always auto-open the report (or, failing that, print its absolute path) — a rendered report the user cannot locate or open has no value.

---

## Audience Tiers

The report adapts its **depth** to the reader. The tier is derived from the active output style unless an explicit `audience` argument overrides it.

### Resolving the tier

Read `outputStyle` from the settings chain — `.claude/settings.local.json` (highest) → `.claude/settings.json` → `~/.claude/settings.json` → hardcoded default — and map it:

| Active output style | Audience tier | Reader |
|---------------------|---------------|--------|
| `MoAI` | `expert` | An engineer who knows the domain and wants the signal, fast |
| `MoAI-Easy` | `basic` | Someone who codes occasionally; the jargon still costs them effort |
| `MoAI-Learn` | `learn` | Someone who wants to genuinely understand the concept, not just the outcome |
| (any other / unresolvable) | `expert` | Safe default — never enrich unasked |

An explicit `audience` argument always wins over the derived value.

### What each tier renders

| Element | `expert` | `basic` | `learn` |
|---------|----------|---------|---------|
| Section prose | Dense, terse | Dense + a one-paragraph plain-language lead per section | Same as basic + why-it-matters framing |
| Jargon | Used bare | **First use is defined inline** — `함수 (function)` style, term followed by a plain-language gloss | Same as basic + a glossary callout box |
| Diagrams | Inline SVG charts only (as today) | + **one mermaid flowchart** of the report's main flow | + **multiple mermaid diagrams** — flow, sequence, and/or state — one per concept that has structure worth seeing |
| Examples | None (numbers speak) | **One worked example** per key claim, with concrete inputs and outputs | Same as basic + a step-by-step walkthrough that derives the result, not just states it |
| Analogies | None | Sparingly, where a concept is genuinely unfamiliar | Freely — an everyday analogy per new concept |
| Closing | Action items | Action items + "what to check yourself" | Action items + self-check questions the reader can answer to confirm they understood |
| HTML size budget | ≤ 50KB | ≤ 120KB | ≤ 120KB |
| **`.md` twin** | **lean** | **lean — identical rule** | **lean — identical rule** |

The last row is the invariant, restated because it is the one that is easy to violate: **no tier adds anything to the markdown twin.** Enrichment is an HTML-only concern.

### Authoring the enrichment (basic / learn)

- **Explain, then state.** At `basic` / `learn`, a section that opens with a raw metric is a miss. Open with one sentence saying what the metric *is* and why the reader should care, then give the number.
- **Define every term on first use.** Everyday-language gloss first, canonical English term in parentheses: `배포 (deployment) — 만든 코드를 실제 사용자에게 내보내는 일`. After the first definition the bare term is fine.
- **Prefer a diagram to a paragraph** when the content is a flow, a sequence, or a state machine. That is exactly what these tiers exist for.
- **Ground every example.** A worked example uses real inputs from the report, not `foo` / `bar`.
- **Never pad.** Enrichment means *more understanding*, not *more words*. A section a beginner already grasps needs no primer.

---

## Diagram Policy

Charts and diagrams follow two different rules depending on what they are.

### Inline SVG charts (all tiers)

Quantitative charts — bar, variance, timeline — are hand-authored **inline SVG**, exactly as today. They work everywhere: browser, email, print, offline. This is unchanged and applies at every tier.

### Mermaid diagrams (`basic` / `learn` tiers only)

Structural diagrams — flowcharts, sequences, state machines — are rendered with **mermaid**, and mermaid needs JavaScript. To keep the single-file, offline-capable promise, mermaid is emitted in a **hybrid form**: the CDN renders it richly in a browser, and a no-JS fallback keeps it readable everywhere else.

Emit all three parts together:

1. **The mermaid source**, in a `<pre class="mermaid">` block — this is what the CDN renders, and it stays human-readable as plain text when it does not.
2. **One mermaid-CDN `<script type="module">`** — placed once per document, at the end of `<body>`, initialized with the design-token palette so diagrams match the report (`--clay` accent on `--ivory` background).
3. **A `<noscript>` fallback** — either a hand-authored inline SVG of the same diagram, or, when the diagram is simple enough that its source reads clearly, a short prose summary of the flow. Never leave `<noscript>` empty.

```html
<pre class="mermaid">
flowchart TD
  A[Markdown source] --> B{Audience tier}
  B -->|expert| C[Dense HTML]
  B -->|basic / learn| D[Enriched HTML + diagrams]
  C --> E[Lean .md twin]
  D --> E
</pre>

<noscript>
  <!-- inline SVG of the same flow, or a prose summary -->
  <p>Flow: the markdown source branches on audience tier — expert renders dense HTML,
     basic/learn render enriched HTML with diagrams. Both paths emit the same lean .md twin.</p>
</noscript>

<script type="module">
  import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs";
  mermaid.initialize({
    startOnLoad: true,
    theme: "base",
    themeVariables: {
      primaryColor:      "#FAF9F5",  /* --ivory  */
      primaryTextColor:  "#141413",  /* --slate  */
      primaryBorderColor:"#D97757",  /* --clay   */
      lineColor:         "#87867F",  /* --g500   */
      secondaryColor:    "#E3DACC",  /* --oat    */
      tertiaryColor:     "#F0EEE6"   /* --g100   */
    }
  });
</script>
```

### Degradation matrix (what the reader actually sees)

| Context | `expert` | `basic` / `learn` |
|---------|----------|-------------------|
| Browser, online | SVG charts | SVG charts + **rendered mermaid diagrams** |
| Browser, offline | SVG charts | SVG charts + `<noscript>` fallback (SVG or prose) |
| Email client (JS stripped) | SVG charts | SVG charts + `<noscript>` fallback |
| Print | SVG charts | SVG charts + fallback (mermaid does not render to print reliably) |

The `expert` tier's strict zero-JS guarantee is **untouched** — the mermaid exception is tier-gated and never fires there.

### Diagram selection

| Content shape | Diagram |
|---------------|---------|
| A process with branches or decisions | `flowchart` |
| An ordered exchange between actors / systems | `sequenceDiagram` |
| A thing that occupies one of several states | `stateDiagram-v2` |
| A quantity compared across categories or time | **inline SVG chart** (not mermaid) |
| A one-liner with no structure | prose (no diagram — resist the urge) |

---

## Six Modes

### Implemented modes

| Mode | Structure sections |
|------|--------------------|
| **`status`** | 4 metric cards · highlights · completed table · velocity SVG bar chart · carryover |
| **`incident`** | TL;DR dark banner · timeline · log excerpts in `<details>` · code diff panel · impact table · action checklist |
| **`plan`** | summary KPI strip · vertical milestone timeline · data-flow SVG · slice table · risk grid · success metrics |
| **`explainer`** | side nav · collapsible `<details>` steps · tabbed code blocks (vanilla JS) · FAQ accordion · callout boxes |
| **`financial`** | 4 KPI cards · income-statement table (item / current / prior / delta / delta-%) · variance SVG horizontal bar chart · notes panel |
| **`pr`** | TL;DR · PR meta row (files / +− / branch) · before/after two-column cards · file tour `<details>` · key points · test checklist · rollout steps |

#### Per-mode input fields

The main fields each template fills (template-internal variable names):

| Mode | Key input fields |
|------|------------------|
| `status` | `{{title}}`, `{{#metrics}}`, `{{#highlights}}`, `{{#completed_rows}}`, `{{#chart_bars}}` |
| `incident` | `{{inc_id}}`, `{{severity}}`, `{{title}}`, `{{#tl_entries}}`, `{{#impact_rows}}`, `{{#actions}}` |
| `plan` | `{{title}}`, `{{#kpis}}`, `{{#milestones}}`, `{{diagram_svg}}`, `{{#slices}}`, `{{#risks}}`, `{{#metrics}}` |
| `explainer` | `{{title}}`, `{{lead}}`, `{{#steps}}`, `{{#config_tabs}}`, `{{#faq_items}}` |
| `financial` | `{{title}}`, `{{period}}`, `{{#kpis}}`, `{{#statement_rows}}`, `{{chart_height}}`, `{{#variance_bars}}` |
| `pr` | `{{pr_ref}}`, `{{title}}`, `{{author}}`, `{{branch}}`, `{{files_changed}}`, `{{additions}}`, `{{deletions}}`, `{{#focus_items}}`, `{{#test_items}}`, `{{#rollout_steps}}` |

---

## Korean Font Policy

This skill permits a single font-CDN `<link>` as the only external dependency, in service of Korean readability.

System-font-only rendering would fracture consistency across operating systems (macOS: Apple SD Gothic Neo, Windows: Malgun Gothic), so a font CDN is required for predictable Korean typography.

### Per-mode font mapping

| Mode | sans (body) | serif (heading) | mono (code) |
|------|-------------|-----------------|-------------|
| `status` / `financial` / `pr` | Pretendard | Pretendard 700 | JetBrains Mono |
| `incident` | Pretendard | Pretendard 700 | JetBrains Mono |
| `plan` | Pretendard | Noto Serif KR | JetBrains Mono |
| `explainer` | Noto Sans KR | Noto Serif KR | JetBrains Mono |
| `editorial` | Pretendard | Chosunilbo Myungjo | JetBrains Mono |
| `legal` | KoPubWorld Batang | KoPubWorld Batang Bold | JetBrains Mono |

CDN URLs and the `preconnect` pattern live in [`references/fonts.md`](references/fonts.md).

---

## Design Tokens (CSS variable contract)

Every mode declares the same 8 CSS variables at `:root`.

```css
:root {
  /* palette */
  --ivory: #FAF9F5;   /* background warm off-white */
  --paper: #FFFFFF;   /* card / panel background */
  --slate: #141413;   /* body text warm black */
  --clay:  #D97757;   /* accent / link terracotta */
  --clay-d:#B85C3E;   /* clay hover state */
  --oat:   #E3DACC;   /* secondary background / divider light tan */
  --olive: #788C5D;   /* secondary accent sage green */

  /* fonts */
  --sans:  "Pretendard", system-ui, -apple-system, sans-serif;
  --serif: "Pretendard", ui-serif, Georgia, serif;
  --mono:  "JetBrains Mono", ui-monospace, "SF Mono", monospace;

  /* layout */
  --max-width:    860px;
  --radius-panel: 12px;
  --radius-row:   8px;
  --border:       1.5px solid var(--g300);
}
```

Greyscale: `--g100: #F0EEE6`, `--g300: #D1CFC5`, `--g500: #87867F`, `--g700: #3D3D3A`

Full contrast verification and print tokens: [`references/design-tokens.md`](references/design-tokens.md)

---

## Recommended chain pattern

This renderer sits at the end of a text-production pipeline. The markdown source may come from any upstream text, analysis, or reporting skill.

```
[text skill] → (optional review / humanize step) → html-report (mode selection)
```

Minimum chain (fast rendering):

```
[text skill] → html-report (mode selection)
```

---

## Usage examples

**Example 1: weekly status report**
```
Render the executive summary result as an HTML report for Hanul Engineering week 11.
```

**Example 2: financial statements**
```
Convert the financial-statement result into an HTML report.
```

**Example 3: incident report**
```
Summarize the payment-gateway 502 outage as an HTML incident report. Severity is SEV-2.
```

**Example 4: PR description document**
```
Turn the realtime notification channel integration pull request into an HTML review document.
```

**Example 5: tier derived from the active output style**
```
Render the caching-layer design as an HTML report.
```
With `MoAI-Easy` active, this resolves to the `basic` tier: each section opens with a plain-language lead, a mermaid flowchart shows the cache read/write path, and every key claim carries a worked example. With `MoAI` active, the same request resolves to `expert` and renders dense. The `.md` twin is the same lean artifact either way.

**Example 6: explicit tier override**
```
Render the incident report as HTML for the expert audience — the on-call engineers already know the system.
```
The explicit `audience: expert` wins over the derived tier, so no primers or diagrams are added even under `MoAI-Learn`.

---

## Non-goals

- Does not replace the markdown default output — HTML is an additional rendering branch.
- Does not pull in external libraries such as React, Vue, a Tailwind CDN, Chart.js, or D3. The only sanctioned external dependencies are the font CDN (all tiers) and the mermaid CDN (`basic` / `learn` tiers, always with a `<noscript>` fallback — § Diagram Policy). Charting stays inline SVG at every tier; mermaid never replaces a chart.
- Does not introduce a build step (webpack, vite, esbuild).
- Does not split the human artifact across multiple files — the report is a single `.html`. The `.md` twin is a *different artifact for a different reader*, not a second half of the report.
- Does not enrich the markdown twin. Audience-tier depth is an HTML-only concern (§ The asymmetry principle).
- External design-system theming (Tailwind-CDN-based brand-token application) is out of scope for the bundled templates here, which are strictly zero-dependency. The `design_system` parameter is not honored by these templates.

---

## References

### Design documents
- [`references/design-tokens.md`](references/design-tokens.md) — CSS variable contract, palette, accessibility
- [`references/fonts.md`](references/fonts.md) — font mapping, CDN URLs, preconnect pattern

### Templates
- [`references/templates/status.html.mustache`](references/templates/status.html.mustache) — status mode
- [`references/templates/incident.html.mustache`](references/templates/incident.html.mustache) — incident mode
- [`references/templates/plan.html.mustache`](references/templates/plan.html.mustache) — plan mode
- [`references/templates/explainer.html.mustache`](references/templates/explainer.html.mustache) — explainer mode
- [`references/templates/financial.html.mustache`](references/templates/financial.html.mustache) — financial mode
- [`references/templates/pr.html.mustache`](references/templates/pr.html.mustache) — pr mode

Design reference: [Thariq Shihipar, "The Unreasonable Effectiveness of HTML"](https://thariqs.github.io/html-effectiveness/) — the origin of the single-file, zero-dependency HTML approach.
