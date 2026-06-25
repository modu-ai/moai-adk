---
name: moai-domain-html-report
description: >
  Markdown-to-single-file-HTML report renderer. Six modes (status, incident,
  plan, explainer, financial, pr) selected automatically by report type.
  Zero external JS/CSS framework dependencies — inline SVG charts, single
  font-CDN exception for Korean readability. Self-contained output for
  email attachment, print, and offline viewing.

when_to_use: >
  Use when a markdown report must be rendered into a single self-contained
  HTML file. Trigger phrases include "render this report as HTML", "weekly
  status report as one HTML file", "convert the financial statements to an
  HTML report", "incident report as HTML", "printable business plan HTML",
  and "email-ready HTML report".

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: true
metadata:
  version: "1.0.0"
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
- A single font-CDN `<link>` is the only external dependency, retained for Korean readability

**This skill does not replace the markdown output.** Markdown remains the single source of truth; HTML rendering is an additional branch that operates on it.

---

## Input

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `markdown` | yes | — | The markdown body to convert |
| `mode` | yes | — | `status` \| `incident` \| `plan` \| `explainer` \| `financial` \| `pr` |
| `slug` | no | auto-derived from the title | Output filename prefix |
| `output_path` | no | `<cwd>/reports/<slug>-<YYYYMMDD>.html` | Output path |
| `font_stack` | no | per-mode default | Font mapping override |

---

## Output

A single `.html` file at `<cwd>/reports/<slug>-<YYYYMMDD>.html`:

- Size: ≤ 50KB (excluding font-CDN traffic, before body compression)
- External dependencies: one font-CDN `<link>` + two `preconnect` hints (Korean fonts)
- Self-contained: opens directly in a browser, email-attachable, offline-capable

---

## After rendering — report back to the user

Once the `.html` file is written, the response MUST do two things:

1. **Summary** — print a concise summary of what was rendered: the mode, the report title, and the key sections or figures the file contains (a short paragraph or a few bullets). Do not paste the full HTML into the response.
2. **Open command** — surface the output file as a ready-to-run open command so the user can view it in one step:

   ```
   !open <output_path>
   ```

   The leading `!` runs the command in the session and opens the file in the default browser. Use the platform-appropriate opener: `!open <file>` on macOS, `!xdg-open <file>` on Linux, `!start <file>` on Windows.

Always provide the open command — a rendered report the user cannot locate or open has no value.

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

---

## Non-goals

- Does not replace the markdown default output — HTML is an additional rendering branch.
- Does not pull in external libraries such as React, Vue, a Tailwind CDN, Chart.js, or D3.
- Does not introduce a build step (webpack, vite, esbuild).
- Does not split output across multiple files — every artifact is a single `.html` file.
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
