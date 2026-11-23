---
name: moai-html-report
description: >
  Renders the html file and the Claude artifact of every report moai writes:
  the html goes beside the md record in ~/.moai/reports/<project>/, and a
  private artifact is published once when the session can publish. Six modes
  (status, incident, plan, explainer, financial, pr) chosen by report type,
  crossed with three audience tiers (expert, basic, learn) taken from the
  active output style; the basic and learn tiers add plain-language leads,
  worked examples and mermaid diagrams to the html, never new facts. Charts
  are inline SVG; fonts and the mermaid script load from CDNs.

when_to_use: >
  Use whenever moai writes a report to ~/.moai/reports/<project>/: after the
  md record is written, render the html beside it and publish the artifact
  version once when the session offers the Artifact tool.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash(mkdir -p *)
user-invocable: false
---

# html-report — renders moai reports as an html file and a Claude artifact

This is an internal skill: it is not in the `/` menu (`user-invocable: false`),
and Claude loads it whenever a report is rendered. `allowed-tools` pre-allows
reading, writing and `mkdir -p`; Artifact publishing is deliberately not
pre-allowed, so every publish asks for approval.

## Purpose and Scope

Every report moai writes has one md record. This skill renders that md as one
html file beside it and, when the session can publish, as one private Claude
artifact.

- **The md is the record.** Every fact, number and decision of the report is
  in the md file in `~/.moai/reports/<project>/`. The html renders that md;
  the artifact is a copy of the html and never the record.
- **The html may explain, never add.** At the `basic` and `learn` tiers the
  html adds explanation — plain words, analogies, diagrams, worked examples —
  but never a fact the md does not hold.
- **External connections.** The html file links two font CDNs (Pretendard on
  jsDelivr, JetBrains Mono on Google Fonts) and, at the `basic` and `learn`
  tiers when it holds a diagram, one mermaid script from jsDelivr. Offline it
  shows the system fonts and each diagram's source text with its
  plain-language fallback. Charts are inline SVG and need nothing.
- **Scripts.** Reports load no external script except the one mermaid script
  in html files; small inline scripts are allowed (the explainer tabs); the
  artifact version carries no external script.
- No CSS or JS framework (Tailwind, Bootstrap, React, Chart.js, D3) and no
  build step.

### What goes in the md, what only in the html

| File | Reader | Content rule |
|------|--------|--------------|
| `.md` record | the next agent, and the user later | Every finding, decision, number, table and action item; the revision line; the artifact `url` once published; a diagram's mermaid source only when the diagram carries information the prose does not (a real state machine, a real dependency graph). |
| `.html` | the human | Renders the md. At `basic` / `learn` it adds plain-language leads, analogies, worked examples, glossary callouts and diagrams (§ Audience Tiers). |

[HARD] **Tier enrichment goes into the html and never into the md.** A
`learn`-tier report and an `expert`-tier report of the same facts have md
records of substantially the same size; only their html differs.

**Consumption rule (token discipline)**: whenever a report is needed as
context — an agent prompt, a follow-up analysis, a read of a past report in a
later session — read the md, never the html. Raw HTML costs three to five
times the tokens for the same content, plus the tier enrichment.

**HTML without an md**: when only an html file exists, extract its facts into
markdown first (strip tags, `<style>`, `<script>`, SVG markup and the tier
enrichment; turn `<table>` into markdown tables), write the extraction beside
it as its md record, and use that.

---

## Input

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `markdown` | yes | — | The md record, written first (§ Delivery) |
| `mode` | yes | — | `status` \| `incident` \| `plan` \| `explainer` \| `financial` \| `pr` |
| `audience` | no | from the output style | `expert` \| `basic` \| `learn` — see § Audience Tiers |
| `slug` | no | from the title | ASCII file-name slug (§ Delivery) |

`mode` and `audience` are **orthogonal**: `mode` picks the report's
*structure* (which sections exist), `audience` picks its *depth* (how much
explanation each section carries). Every mode renders at every tier.

---

## Delivery

**Folder.** `~/.moai/reports/<project>/`, where `<project>` is the top folder
name of the repository's primary checkout — in a worktree too:
`basename "$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")"`.
Outside git it is the current folder's name. Create the folder with
`mkdir -p` when it is missing.

**File names.** `<YYYY-MM-DD>-<ascii-slug>.md` and `.html` with the same
stem. The date comes from `date -u +%Y-%m-%d`. The slug is lowercase ASCII
words joined by `-`; a title in another script becomes short English words
(`결제 502 장애` → `payment-502-outage`). When a file of that name already
belongs to another report, add `-2`, `-3`, … to both names.

Every report goes out in this order:

1. **The md record.** Write `<stem>.md` first, with every fact, number and
   decision and the revision line.
2. **The html beside it.** Render the md through the mode template
   (§ Render order) and write `<stem>.html` in the same folder.
3. **The Artifact publish, once.** When the session offers the Artifact tool,
   load the `artifact-design` skill, build the artifact version as
   `<stem>.artifact.html` beside the pair (`references/artifact-contract.md`)
   and publish it once: `icon` one generic word on the first publish, a
   one-sentence `description`, private. Give the private link with the two
   file paths, and write the link into the md as `Artifact: <url>`. Never
   share or delete the artifact. Do not probe the account or the environment
   first; the publish answer itself says whether it worked. With a PreToolUse
   hook on Edit or Write in the session, edit the file first and publish it as
   its own call. To republish the same report: in the same conversation,
   publish the same file path again; in a later conversation, pass the `url`
   the md keeps.
4. **The refusal line.** When the publish is refused or the session has no
   Artifact tool, give the two file paths and one line:
   `No artifact was made: <reason>.` A refusal never fails the report.

Sessions that cannot publish, for the reason line (not a check to run before
publishing):

- a session that uses an API key, a gateway token or a cloud-provider
  credential; a session on a GLM backend is an API-key session;
- Amazon Bedrock, Google Cloud's Agent Platform or Microsoft Foundry as the
  model provider;
- an organization with customer-managed encryption keys (CMEK), HIPAA or Zero
  Data Retention;
- a `-p` run, the Agent SDK, the GitHub Action or an MCP-server context
  (off by default there);
- a login with the Anthropic Console account option, which gives no claude.ai
  credentials;
- artifacts turned off: `CLAUDE_CODE_DISABLE_ARTIFACT=1`,
  `"enableArtifact": false`, or Artifact in `permissions.deny`.

**No browser on its own.** Show the paths and the link; open a file only when
the user asks: macOS `open "<path>"`, Windows from Git Bash
`cmd //c start "" "<path>"`, elsewhere give the path.

**Summary to the user.** Name the mode, the audience tier and where it came
from (the output style or an explicit `audience`), the report title, and the
key sections or figures in a short paragraph or a few bullets. Never paste the
html.

**The other report rules:**

- The md carries one revision line: the head commit the report was made at, or `UNVERSIONED`.
- A table of more than 200 rows goes to a file beside the report; the answer shows its row count and path.
- A report that lists findings gets a `.jsonl` file beside it, one finding per line, with the md's ids.
- A repair item that names a folder gets a `claude-cli://open` link: URL-encoded `q` of at most 5,000 characters, absolute `cwd` without `..`; the md holds it in a code block, the html as a link.
- A repair prompt over 5,000 characters, or one holding a value shaped like a credential, gets no link.
- A report folder inside a git work tree gets a `.gitignore` that ignores every file; an existing one is kept.
- With a PreToolUse hook on Edit or Write, change the file first, then publish it as its own call.

---

## Render order

There is no mustache engine: the agent reads the template and writes the
finished html itself.

1. **Pick the mode template**: `references/templates/<mode>.html.mustache`.
2. **Shape the data** from the md record into the fields the template names
   (§ Six Modes, and the comment above each template section).
3. **Escape every outside text** — PR bodies, logs, issue text, command
   output — (`&`, `<`, `>`, `"`, `'`) before it goes into any field.
4. **`*_html` fields get only HTML the agent built itself from the md.**
   Two-brace fields print escaped text; the three-brace fields (`*_html`,
   and `diagram_svg` in plan) take markup as it is, so outside text never
   goes into them unescaped.
5. **Compute the SVG coordinates** with the formulas in each chart's
   template comment.
6. **Delete every template comment and every unused section** (an empty
   list's section with its heading).
7. **Write the html** (§ Delivery step 2).
8. **Build the artifact version** when publishing
   (`references/artifact-contract.md` § 1).

### Final check before the publish

- The `<title>` is a name of two to four words.
- Every external host is on the allow list: the html file uses
  `cdn.jsdelivr.net` (Pretendard, mermaid), `fonts.googleapis.com` and
  `fonts.gstatic.com`; the artifact version uses `fonts.googleapis.com` and
  `fonts.gstatic.com` only.
- No `{{` is left in either file.
- The size fits: html at most 50 KB (`expert`) or 120 KB (`basic`, `learn`);
  the artifact version at most 16 MB with its `data:` URIs.
- One preview of the artifact version when the session offers a page-check
  tool, and no more.

---

## Language

All report text and labels are in the user's conversation language, phrased
the way a native writer of that language would write them (Korean in 해요체);
code, identifiers, paths and commands stay as they are. The `lang` field of
every template is that language's code (`ko`, `en`, `ja`, …). The templates'
fixed labels (Highlights, Timeline, Root Cause, Sources, Generated, …) are
English: write each one in the conversation language when filling the
template. The md record follows the same rule.

---

## Audience Tiers

The report adapts its **depth** to the reader.

### Resolving the tier

1. An explicit `audience` argument.
2. The output style name the session already shows.
3. `outputStyle` in `$CLAUDE_CONFIG_DIR/settings.json` (`~/.claude/settings.json`
   when the variable is not set).
4. The default `expert`.

| Active output style | Audience tier | Reader |
|---------------------|---------------|--------|
| `MoAI` | `expert` | An engineer who knows the domain and wants the signal, fast |
| `MoAI-Easy` | `basic` | Someone who codes occasionally; the jargon still costs them effort |
| `MoAI-Learn` | `learn` | Someone who wants to genuinely understand the concept, not just the outcome |
| (any other / unresolvable) | `expert` | Safe default — never enrich unasked |

### What each tier renders

| Element | `expert` | `basic` | `learn` |
|---------|----------|---------|---------|
| Section prose | Dense, terse | Dense + a one-paragraph plain-language lead per section | Same as basic + why-it-matters framing |
| Jargon | Used bare | **First use is defined inline** — `함수 (function)` style, term followed by a plain-language gloss | Same as basic + a glossary callout box |
| Diagrams | Inline SVG charts only | + **one mermaid flowchart** of the report's main flow | + **several mermaid diagrams** — flow, sequence, and/or state — one per concept that has structure worth seeing |
| Examples | None (numbers speak) | **One worked example** per key claim, with concrete inputs and outputs | Same as basic + a step-by-step walkthrough that derives the result, not just states it |
| Analogies | None | Sparingly, where a concept is genuinely unfamiliar | Freely — an everyday analogy per new concept |
| Closing | Action items | Action items + "what to check yourself" | Action items + self-check questions the reader can answer to confirm they understood |
| HTML size budget | ≤ 50KB | ≤ 120KB | ≤ 120KB |
| **`.md` record** | **no enrichment** | **no enrichment — identical rule** | **no enrichment — identical rule** |

The last row is the invariant: **no tier adds anything to the md record.**

### Authoring the enrichment (basic / learn)

- **Explain, then state.** At `basic` / `learn`, a section that opens with a raw metric is a miss. Open with one sentence saying what the metric *is* and why the reader should care, then give the number.
- **Define every term on first use.** Everyday-language gloss first, canonical English term in parentheses: `배포 (deployment) — 만든 코드를 실제 사용자에게 내보내는 일`. After the first definition the bare term is fine.
- **Prefer a diagram to a paragraph** when the content is a flow, a sequence, or a state machine.
- **Ground every example.** A worked example uses real inputs from the md record, not `foo` / `bar`, and adds no new fact.
- **Never pad.** Enrichment means *more understanding*, not *more words*. A section a beginner already grasps needs no primer.

---

## Diagram Policy

### Inline SVG charts (all tiers)

Quantitative charts — bar, variance, timeline — are hand-authored **inline
SVG**. They work in a browser, in email, in print and offline, at every tier.

### Mermaid diagrams in the html file (`basic` / `learn` tiers only)

Structural diagrams — flowcharts, sequences, state machines — use **mermaid**,
which needs JavaScript. Emit all three parts:

1. **The source** in a `<pre class="mermaid">` block. It stays readable as
   plain text when the script does not run.
2. **A visible fallback** right after it: a `<div class="mmd-fallback">`
   outside any `<noscript>`, holding a short plain-language summary of the
   diagram or a hand-drawn inline SVG. Never leave it empty.
3. **One mermaid script, once per document** at the end of `<body>`:
   `https://cdn.jsdelivr.net/npm/mermaid@11.4.1/dist/mermaid.min.js` — the
   exact version, the UMD build, a classic `<script src>` — followed by a
   small inline script that renders the diagrams and hides the fallbacks only
   after they rendered.

```html
<pre class="mermaid">
flowchart TD
  A[md record] --> B{Audience tier}
  B -->|expert| C[Dense html]
  B -->|basic / learn| D[html with explanation and diagrams]
</pre>
<div class="mmd-fallback">
  <p>Flow: the md record is rendered by audience tier. Expert gives a dense
     html; basic and learn add explanation and diagrams. The md stays the same.</p>
</div>

<!-- once, at the end of <body> -->
<script src="https://cdn.jsdelivr.net/npm/mermaid@11.4.1/dist/mermaid.min.js"></script>
<script>
  if (window.mermaid) {
    mermaid.initialize({
      startOnLoad: false,
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
    mermaid.run({ querySelector: "pre.mermaid" }).then(function () {
      document.querySelectorAll(".mmd-fallback").forEach(function (el) {
        el.hidden = true;
      });
    });
  }
</script>
```

### What the reader sees

| Context | `expert` | `basic` / `learn` |
|---------|----------|-------------------|
| Browser, online | SVG charts | SVG charts + rendered diagrams (fallback hidden) |
| Browser, offline or CDN blocked | SVG charts | SVG charts + diagram source text + visible fallback |
| JavaScript off (some mail clients) | SVG charts | Same as offline |
| Print | SVG charts | What the screen shows: the rendered diagram when it rendered, otherwise the source and the fallback |

The `expert` tier adds no mermaid diagram.

### Mermaid in the artifact version

The artifact version keeps each `pre.mermaid` block and carries no mermaid
script and no `.mmd-fallback` block: the artifact viewer renders `pre.mermaid`
itself (`references/artifact-contract.md` § 7). Inline SVG charts and diagram
selection are unchanged.

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

| Mode | Structure sections |
|------|--------------------|
| **`status`** | 4 metric cards · highlights · shipped table · velocity SVG bar chart · carryover |
| **`incident`** | TL;DR dark banner · timeline · log excerpts in `<details>` · code diff panel · impact table · action checklist |
| **`plan`** | summary strip · vertical milestone timeline · data-flow SVG · risk grid · success metrics |
| **`explainer`** | side nav · collapsible `<details>` steps · tabbed code blocks (small inline script) · FAQ accordion · callout box |
| **`financial`** | 4 KPI cards · income-statement table (item / current / prior / change / change %) · variance SVG horizontal bar chart · notes panel |
| **`pr`** | TL;DR · PR meta row (files / +− / branch) · background with before/after cards · file tour `<details>` · review focus · test checklist · rollout steps |

### Fields per mode

The names each template uses. Every template also takes `lang`, `title` (the
two-to-four-word name, which is also the `<title>`), `footer_sources` and
`generated_at`. Three-brace HTML fields are marked **(html)**; the comment
above each template section gives the item fields and formulas.

| Mode | Fields |
|------|--------|
| `status` | `date_range`, `repository`, `metrics` [label, value, delta, delta_type, warn], `highlights` [text_html **(html)**], `shipped_table` [id, link, title, author, risk, risk_label], `velocity_chart` {caption, y1, y2, y3, bars [x, y, height, fill, cx, label_y, text_fill, is_peak, value, label]}, `carryover` {in_review, blocked, slipped [text, who]}, `has_in_review`, `has_blocked`, `has_slipped` |
| `incident` | `inc_id`, `severity`, `sev_class`, `status_label`, `status_class`, `duration`, `detected_at`, `owner`, `tldr_html` **(html)**, `tl_entries` [time, dot_class, body_html **(html)**, log_excerpt], `root_cause_html` **(html)**, `diff_path`, `diff_lines` [cls, text], `impact_rows` [label, value], `actions` [done, initials, assignee, desc_html **(html)**, deadline] |
| `plan` | `eyebrow`, `goal_html` **(html)**, `summary_cells` [label, value, accent], `milestones_intro`, `milestones` [when, title, desc, done, tags], `dataflow_intro`, `diagram_svg` **(html)**, `diagram_caption`, `risks` [risk, severity, sev_class, mitigation], `success_metrics` [label, target, current, met] |
| `explainer` | `eyebrow`, `tldr_html` **(html)**, `nav_links` [href, label, l2], `nav_files` [string], `has_nav_files`, `walkthrough_title`, `walkthrough_intro`, `steps` [num, title, where, body_html **(html)**, open], `config_title`, `config_intro`, `config_tabs` [label, pane_id, code_html **(html)**, first], `callout_html` **(html)**, `faq_items` [question, answer_html **(html)**] |
| `financial` | `eyebrow`, `period`, `kpis` [label, value, delta, delta_type], `statement_rows` [label, current, prior, change_abs, change_pct, change_cls, is_header, is_subtotal, is_total], `chart_height`, `chart_height_minus_20`, `caption`, `variance_bars` [label, value_str, pct_str, label_y, bar_x, bar_rect_y, bar_width, bar_fill, val_x, val_y, val_fill], `notes` [string], `has_notes` |
| `pr` | `eyebrow`, `pr_ref`, `file_count`, `adds`, `dels`, `branch_from`, `branch_to`, `author`, `tldr_html` **(html)**, `background_html` **(html)**, `ba_before` [text], `ba_after` [text], `tour_intro`, `files` [path, path_short, badge_class, badge_label, adds, dels, why_html **(html)**, open], `focus_items` [num, title, desc_html **(html)**], `tests` [done, label, note], `rollout_intro`, `rollout_steps` [when, pct_label, desc] |

`has_*` fields are true when their list is not empty; they replace
list-length sections, which only some mustache engines read.

---

## Fonts

- **html file, every mode:** Pretendard (jsDelivr) for body text and
  headings, JetBrains Mono (Google Fonts) for code.
- **Artifact version:** the Noto Sans family of the report's language from
  Google Fonts (Noto Sans KR, JP, SC or TC; Noto Sans for any other
  language) and JetBrains Mono, because the artifact page takes stylesheets
  only from Google Fonts.

URLs, the preconnect pattern and the token values are in
[`references/fonts.md`](references/fonts.md).

---

## Design tokens and width

The palette, font, layout and print tokens and the contrast table are in
[`references/design-tokens.md`](references/design-tokens.md), the single
source for every template. The warm palette (ivory ground, clay accent) stays
by user decision. Each template sets its own `--max-width`:

| Mode | `--max-width` |
|------|---------------|
| `status` | 860px |
| `incident` | 860px |
| `plan` | 1100px |
| `explainer` | 1100px (prose column `--content-max` 680px) |
| `financial` | 920px |
| `pr` | 1040px |

Print rules (`@media print`) live in each template under
[`references/templates/`](references/templates/).

---

## Recommended chain pattern

This renderer sits at the end of a report: the md record comes from any moai
route or agent that wrote it.

```
[md record written] → html-report (mode selection) → [artifact publish when offered]
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
With `MoAI-Easy` active, this resolves to the `basic` tier: each section opens with a plain-language lead, a mermaid flowchart shows the cache read/write path, and every key claim carries a worked example. With `MoAI` active, the same request resolves to `expert` and renders dense. The md record is the same either way.

**Example 6: explicit tier override**
```
Render the incident report as HTML for the expert audience — the on-call engineers already know the system.
```
The explicit `audience: expert` wins over the derived tier, so no primers or diagrams are added even under `MoAI-Learn`.

**Example 7: Korean weekly status report**
```
이번 주 상태 보고서를 html 로 만들어 주세요.
```
`status` mode, `lang="ko"`, every label and sentence in Korean 해요체.

**Example 8: Korean incident report**
```
결제 502 장애를 장애 보고서로 정리해 주세요. 심각도는 SEV-2 예요.
```
`incident` mode; the `<title>` is a short name such as `결제 502 장애`, and the incident id goes in the body.

---

## Non-goals

- Does not replace the md record: the md is written first and stays the record.
- The artifact is a private copy of the html, published once when the session can publish; this skill never shares or deletes it.
- Does not pull in a CSS or JS framework (React, Vue, Tailwind, Chart.js, D3). Charting stays inline SVG at every tier; mermaid never replaces a chart.
- Does not introduce a build step (webpack, vite, esbuild) or a mustache engine.
- Does not split the human-facing report across several files: the report is one `.html` beside its `.md`.
- Does not enrich the md record. Audience-tier depth is an html-only concern.

---

## References

### Design documents
- [`references/artifact-contract.md`](references/artifact-contract.md) — the Claude Artifact page contract: no skeleton, title rule, dark-mode blocks, safe-area padding, allowed hosts, fonts by language, `pre.mermaid`, things not to use, publish arguments
- [`references/fonts.md`](references/fonts.md) — font mapping, CDN URLs, preconnect pattern
- [`references/design-tokens.md`](references/design-tokens.md) — palette, dark values, font and layout tokens, print block, contrast table
- [`references/craft-fundamentals.md`](references/craft-fundamentals.md) — spacing, numerals, neutrals, copy, structure
- The host's `artifact-design` skill — load it before every publish

### Templates
- [`references/templates/status.html.mustache`](references/templates/status.html.mustache) — status mode
- [`references/templates/incident.html.mustache`](references/templates/incident.html.mustache) — incident mode
- [`references/templates/plan.html.mustache`](references/templates/plan.html.mustache) — plan mode
- [`references/templates/explainer.html.mustache`](references/templates/explainer.html.mustache) — explainer mode
- [`references/templates/financial.html.mustache`](references/templates/financial.html.mustache) — financial mode
- [`references/templates/pr.html.mustache`](references/templates/pr.html.mustache) — pr mode

Design reference: [Thariq Shihipar, "The Unreasonable Effectiveness of HTML"](https://thariqs.github.io/html-effectiveness/) — the origin of the single-file HTML approach.
