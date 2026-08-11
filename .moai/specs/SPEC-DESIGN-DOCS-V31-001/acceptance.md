# acceptance.md — SPEC-DESIGN-DOCS-V31-001

> Verification layer. Given-When-Then scenarios. The "book-level prose" rubric is defined here as objective, measurable criteria — it is the load-bearing acceptance contribution of this SPEC.

---

## §A. The "introductory-book-level friendly prose" rubric

### §A.1 The five rubric pillars

Every Korean page rewritten under M4 MUST satisfy all five pillars. The pillars are ordered by what a reader perceives first.

1. **Concept-first exposition.** The page opens with a 1–3 paragraph plain-prose introduction that answers "what is this and why does it exist?" BEFORE any table, code block, or callout. The introduction is NOT a feature list; it is a mental-model setup. (Test: does the first H2 appear after at least one prose paragraph that contains zero table / code-block / callout markdown?)

2. **Step-by-step structure.** Tutorial and onboarding pages use `## Step N — <imperative>` headings to decompose the procedure. Reference pages MAY waive this (a CLI flag list is a flat enumeration, not a procedure). The waiver is per-page-class (§A.2).

3. **Infographic density.** Every concept page contains at least one infographic: a Mermaid `flowchart TD` diagram, a mascot shortcode (`{{< mascot <pose> >}}`) with caption, or an inline figure (SVG/PNG) with descriptive alt text. The infographic MUST appear in the body, not only in the header.

4. **Runnable examples.** Every code block that illustrates a command or API MUST be runnable verbatim by a reader who has completed `moai init`. Code blocks that are illustrative-only (e.g. a hypothetical directory tree) MUST carry an `illustrative-only` comment. CLI pages MUST have at least 3 runnable examples.

5. **Friend-of-a-friend explainability.** After reading the page, a reader MUST be able to explain the concept to a friend in two sentences. This is a self-audit step (not a runtime check): the author records the two-sentence summary in the page's authoring trail. The summary is NOT published on the page; it is the evidence that the page clears the bar.

### §A.2 Page classes & per-class floors

| Page class | Min prose word count | Min infographics | Min runnable code | Min `## Step N` | Examples |
|---|---:|---:|---:|---:|---|
| **onboarding** (getting-started) | 600 | 1 | 1 | 3 | introduction, installation, quickstart, init-wizard |
| **concept** (core-concepts, multi-llm intro, cost-optimism intro) | 800 | 2 | 0 | 0 | what-is-moai-adk, harness-engineering, spec-based-dev, trust-5, tokenomics |
| **tutorial** (workflow-commands, advanced with a procedure) | 700 | 1 | 3 | 4 | moai-plan, moai-run, moai-sync, factory-mode, autonomy-tier |
| **reference** (cli-reference, utility-commands) | 300 per command | 0 (waivable) | 1 per command | 0 | every cli-reference page |
| **changelog / contributing** | 100 | 0 | 0 | 0 | changelog, contributing |

Word count is measured on Korean prose (body text excluding code blocks, tables, frontmatter, shortcode arguments, HTML comments). The tool is `wc -w` after stripping — the exact stripping pipeline is in §C.7 (quality gates). *(v0.2.0 D3 fix: §C.7 pipeline now strips frontmatter + fenced code + pipe-table rows + Hugo shortcodes + HTML comments — not just fenced code blocks. The v0.1.0 pipeline over-counted by ~45% on a 2-page sample: `what-is-moai-adk.md` measured 3886→2133 words old→new; `trust-5.md` measured 1721→599 words. The corrected pipeline matches its description and is no longer gaming-able by padding with tables / shortcodes / frontmatter.)*

### §A.3 Voice & vocabulary rubric (binary)

Each Korean page MUST pass the following binary checks (drawn from the handoff README §2):

- Banned-vocabulary grep returns 0 matches: `혁신적인|leverage|솔루션|Game-changing|Cutting-edge|Next-level|Disruptive|절대로|유일한|최고의|지금 안 하면`.
- First-mention term expansion: every technical term on its first appearance on the page is parenthesized-glossed (e.g. `에이전트(스스로 일하는 AI)`). Subsequent uses MAY use the bare term.
- No body emoji (CLAUDE.local.md §17.1 — typographic arrows `→ ← ↓ ✓ ✗` are permitted; pictographic emoji are not).
- Mermaid diagrams are TD-only.
- Every inline icon uses `{{< icon >}}`.

### §A.4 Friend-explainability mechanical predicate (v0.2.0 D4 fix)

For each page authored under M4, the author (or the run-phase agent) MUST record a two-sentence friend-explanation in a page-specific authoring-trail sidecar (a `<page-slug>.author-trail.md` file under `.moai/state/spec-design-docs-v31-001/author-trails/`, gitignored — NOT published in the build). The two-sentence summary MUST satisfy BOTH halves of a mechanical predicate, checked by a single grep per sidecar:

**(5a) Korean causal connector (at least one).** The sidecar MUST contain at least one of the causal connectors: `왜냐하면`, `때문에`, `따라서`, `그래서`, `덕분에`. These are the Korean morphemes that signal "this is the reason it works / this is why a reader should care" — a summary without any of them is structurally a description, not an explanation.

**(5b) Concrete moai-adk capability noun (at least one).** The sidecar MUST contain at least one noun naming a SPECIFIC moai-adk capability from the allowlist: `SPEC`, `TRUST 5`, `harness`, `goal`, `factory`, `에이전트`, `관리자 에이전트`, `Geekdoc`, `Pretendard`. Generic nouns (`AI`, `도구`, `기능`, `시스템`) do NOT satisfy the predicate — they could name any product.

**Grep-able check command (runs at M5 against every sidecar):**

```bash
# (5a) causal connector — MUST match ≥1
grep -E '왜냐하면|때문에|따라서|그래서|덕분에' <sidecar>
# (5b) capability noun — MUST match ≥1
grep -E 'SPEC|TRUST 5|harness|goal|factory|에이전트|관리자 에이전트|Geekdoc|Pretendard' <sidecar>
# Sidecar non-empty (≥2 sentences, approximated by ≥30 chars + ≥1 period/아래/다.)
wc -c < <sidecar>   # ≥ 60 bytes
```

If any of the three checks returns no match / sub-threshold, the page FAILS the M5 gate (see §B.6 AC-M5-004, promoted to MUST per D9). *(v0.2.0 D4 decision: tighten — the predicate is mechanical. A separate self-audit-only bar was rejected as unverifiable theater per plan-auditor D4 finding.)*

### §A.5 "NEW" badge inventory (binary)

The set of pages carrying the NEW badge MUST equal exactly the v3.1 feature-catalog pages (spec.md §F.1) plus any entirely new section index pages. No pre-v3.1 page carries the badge; every catalog page carries it.

---

## §B. Per-milestone acceptance criteria

> v0.2.0 D2 consolidation: total AC count reduced 32 → 25 (Tier L ceiling). Merge map documented per AC. No AC silently dropped — every v0.1.0 AC is either carried forward or folded into a merged AC that still verifies its concern.

### §B.1 M0 — IA freeze

**AC-M0-001** — Given the v3.1 feature catalog (spec.md §F.1) AND the Docs Home handoff `01 Docs Home.dc.html` 12-card grid, when the new `content/ko/_meta.yaml` is rendered, then (a) every catalog row's "New IA home" column resolves to a section present in the manifest, AND (b) the live section list matches the handoff's 12-section grid (the user-decided 12-section shape — `changelog` is a footer link, not a sidebar section). *(Merge: v0.1.0 AC-M0-001 + AC-M0-003 — both verify IA structure against one source-of-truth; rationale: the user decision fixed 12-vs-13, leaving one structural check.)*

**AC-M0-002** — Given any slug change in the IA, when `vercel.json` is inspected, then a redirect entry from the old slug to the new slug exists for every locale (ko/en/ja/zh).

### §B.2 M1 — Design-token port

**AC-M1-001** — Given `docs-site/static/moai-brand.css`, when its `:root` block is diffed against the handoff `colors_and_type.css` `:root` block (excluding the `[data-theme="dark"]` block), then (a) the token vocabulary matches byte-for-byte on every token name and value, AND (b) the production CSS tree (`docs-site/static/*.css`) grepped for `[data-theme="dark"]` shows zero NEW dark-mode selectors (frozen dead code from prior SPECs tolerated, NOT augmented). *(Merge: v0.1.0 AC-M1-001 + AC-M1-005 — both verify token-port surface correctness via one CSS diff/grep; rationale: dark-mode-leak prevention is an invariant of the token port.)*

**AC-M1-002** — Given the 6 mascot pose PNGs and 3 logo variants, when `docs-site/static/mascots/` and `docs-site/static/logos/` are inspected, then all 9 assets are present (filenames include `moai-logo-4-WH.png` per v0.2.0 D5 fix).

**AC-M1-003** — Given the docs-site, when `cd docs-site && hugo --gc --minify` is run, then the build exits 0 with zero warnings.

**AC-M1-004** — Given `moai-brand.css`, when its header comment is read, then the FROZEN stamp cites `SPEC-DESIGN-DOCS-V31-001` (replacing the prior `SPEC-DESIGN-DOCSV2-001` stamp, with the prior stamp preserved in a "Predecessor:" line).

### §B.3 M2 — Header & home rewrite

**AC-M2-001** — Given the rendered docs-site header, when any page is loaded, then the header matches the handoff header structure (sticky, blurred background, search affordance with `⌘K` hint, version pill, locale switcher, GitHub link) and does NOT contain a dark-mode toggle button.

**AC-M2-002** — Given the docs-site home page (`/ko/`), when rendered, then it contains (a) a hero block with mascot + headline + dual CTA + install command, (b) a 3-card value grid (Tokenomics / Self-Learning / Harness), (c) a section grid matching the frozen IA, (d) a book CTA card (the book CTA target URL `https://book.mo.ai.kr` verified live, HTTP 200, at v0.2.0).

**AC-M2-003** — Given the home page hero, when the version pill is inspected, then it reads `v3.1-rc.1` (matching `hugo.toml [params].version`).

### §B.4 M3 — NEW-badge mechanism

**AC-M3-001** — Given a page's NEW-flag state, when the flag is present (`new: true` OR `added_in: "v3.1"`), then the NEW badge appears beside the `<h1>` AND beside the page title in the sidebar; when the flag is absent, then NO badge appears on either surface. *(Merge: v0.1.0 AC-M3-001 + AC-M3-002 (presence) + AC-M3-004 (absence) — both polarities of the same flag→badge check; rationale: presence and absence are one binary assertion.)*

**AC-M3-002** — Given a section `_meta.yaml` with `new_items: [slug-a, slug-b]` OR a section-level `new: true`, when the sidebar is rendered, then the badge appears beside each named slug AND/OR beside the section heading as configured. *(Merge: v0.1.0 AC-M3-003 + AC-M3-004-section-variant — both verify section-level flag rendering.)*

**AC-M3-003** — Given the `{{< new-badge v3.1 >}}` shortcode, when invoked in a page body, then the inline badge renders with `--color-primary` background, white text, and the "NEW" caption.

### §B.5 M4 — Korean content rewrite (sub-milestone AC)

For EACH sub-milestone M4.1 … M4.8:

**AC-M4-X-001** — Given every page in the sub-milestone's section, when the §A.2 per-class floor table is applied AND the §A.4 friend-explainability mechanical predicate is run against the page's authoring-trail sidecar, then every page meets (a) its class's floors (prose word count, code count, step count) AND (b) the sidecar predicate passes (≥1 Korean causal connector AND ≥1 concrete capability noun AND ≥60 bytes). *(Merge: v0.1.0 AC-M4-X-001 + AC-M4-X-003 — friend-explainability folded into the floor AC now that the predicate is mechanical per D4; rationale: one per-page rubric check, not two.)*

**AC-M4-X-002** — Given every page in the section, when the §A.3 voice/vocabulary rubric is run, then every binary check passes (banned-vocab grep 0, first-mention expansion present, no body emoji, TD-only Mermaid, icon-shortcode usage). The infographic floor (REQ-KO-001 pillar 3) is verified as part of this check.

**AC-M4-X-003** — Given the section's v3.1 catalog pages (per spec.md §F.1), when rendered, then every catalog page in this section carries the NEW badge. *(Renumber: v0.1.0 AC-M4-X-004 — X-003 merged away.)*

### §B.6 M5 — Korean verification gate (MUST — v0.2.0 D9 promotion)

> **M6 (en derivation) MUST NOT begin until every MUST AC in this section passes.** This is the ko→non-canonical-locale gate. (v0.2.0 D9 fix: M5 promoted from SHOULD to MUST; the M6-gating sentence is normative.)

**AC-M5-001** — Given the `ko` locale, when `cd docs-site && hugo --gc --minify` is run, then exit 0 zero warnings. *(Promoted to MUST per D9.)*

**AC-M5-002** — Given the `ko` locale, when the `hns-oss-docs-verify` recipe is run against `ko`, then every check passes (URL blacklist, TD-only Mermaid, body-emoji scan, page-count parity skeleton). *(Promoted to MUST per D9; de-duplicated against M7 — M5 is the ko-only gate, M7 is the 4-locale gate.)*

**AC-M5-003** — Given the v3.1 feature catalog, when the `ko` tree is grepped for the NEW-badge flag, then the flagged-page set equals the catalog page set (§A.5). *(Promoted to MUST per D9.)*

**AC-M5-004** — Given the M4 authoring-trail sidecars under `.moai/state/spec-design-docs-v31-001/author-trails/`, when the §A.4 mechanical predicate is run across every sidecar, then every sidecar passes all three checks (causal connector, capability noun, ≥60 bytes). *(Promoted to MUST per D9 — the friend-explainability gate is now mechanical, not a subjective self-audit.)*

### §B.7 M6 — en/ja/zh derivation

For EACH locale in {en, ja, zh}:

**AC-M6-X-001** — Given the locale derived from the `ko` SSOT pinned at M5's passing commit, when `cd docs-site && hugo --gc --minify` is run AND the locale's pages are diffed against their `ko` counterparts, then (a) every page preserves its ko counterpart's Mermaid diagrams, code blocks, mascot placements, shortcode usage, and frontmatter `weight` / `new` / `added_in` flags, AND (b) the build exits 0 with zero warnings. *(Merge: v0.1.0 AC-M6-X-001 (preservation) + AC-M6-X-002 (build) — rationale: both run as one per-locale verification; the preservation diff is verified in the same hugo build surface.)*

**AC-M6-X-002** — Given the locale's sidebar, when `data/menu/main.yaml` is inspected, then every per-locale name map is localized (icon values remain semantic identifiers — not translated). *(Renumber: v0.1.0 AC-M6-X-003.)*

### §B.8 M7 — Final build & sunset readiness

**AC-M7-001** — Given the full docs-site (4 locales), when `cd docs-site && hugo --gc --minify` is run AND a CSP audit is performed on the rendered output (no `script-src` additions from the handoff), then (a) the build exits 0 with zero warnings, AND (b) zero handoff-runtime JS dependencies (`_ds_bundle.js` / `_ds_manifest.json` / `support.js`) are present. *(Merge: v0.1.0 AC-M7-001 + AC-M7-006 — both are final-build-day audits; rationale: one build+audit AC, one CSP grep on its output.)*

**AC-M7-002** — Given the 4-locale corpus, when page-count parity is computed, then `ko` page count == `en` == `ja` == `zh` (496 pages total, 124 per locale — or the new count post-IA-reconciliation).

**AC-M7-003** — Given the 4-locale corpus, when H2-section-count parity is computed per page slug, then zero parity gaps exist across locales.

**AC-M7-004** — Given `hugo.toml` AND every NEW-badged page, when `[params].version` is read AND the badged pages' flag forms are inspected, then (a) `[params].version == v3.1-rc.1`, AND (b) `added_in: "v3.1"` is the preferred form (not bare `new: true`) for at least 80% of badged pages (the 20% tolerance covers section-level `_meta.yaml` entries where `new: true` is the only form). *(Merge: v0.1.0 AC-M7-004 + AC-M7-005 — both verify v3.1 release-readiness identifiers; rationale: version and badge form are the same release-labeling concern.)*

---

## §C. Quality gates (binding at every milestone boundary)

1. `cd docs-site && hugo --gc --minify` exit 0, zero warnings.
2. 4-locale page-count parity (zero gap).
3. URL blacklist grep (adk.mo.ai.kr whitelist): `grep -rE 'https?://(?!adk\.moai\.kr|github\.com/modu-ai)' docs-site/content/` returns 0.
4. Mermaid TD-only: `grep -rn 'flowchart \(LR\|RL\|BT\)' docs-site/content/` returns 0.
5. Body-emoji scan: per `hns-oss-docs-verify` recipe.
6. NEW-badge inventory matches the catalog (§A.5).
7. Korean prose-floor strip-and-count pipeline (used by M4 sub-milestones). Strips: (a) YAML frontmatter between the first two `---` fences, (b) fenced code blocks, (c) pipe-table rows, (d) Hugo shortcodes (`{{< … >}}` and `{{% … %}}`), (e) HTML comments (`<!-- … -->`). Verified at v0.2.0 on two real ko samples — see §A.2.
   ```bash
   for f in docs-site/content/ko/<section>/*.md; do
     awk '
       NR==1 && /^---[[:space:]]*$/ {fm=1; next}
       fm && /^---[[:space:]]*$/ {fm=0; next}
       fm {next}
       /^```/{c++; next}
       c%2==1 {next}
       /^\|/{next}
       /^[[:space:]]*<!--/ {in_html=1}
       in_html && /-->[[:space:]]*$/ {in_html=0; next}
       in_html {next}
       {gsub(/\{\{<[^>]*>\}\}/, ""); gsub(/\{\{[^}]*\}\}/, ""); print}
     ' "$f" | wc -w
   done
   ```
   *(v0.2.0 D3 fix: the v0.1.0 pipeline stripped only fenced code blocks — frontmatter / tables / shortcodes / HTML comments were silently included in the count, inflating it by ~45% and making the floor gaming-able by padding with tables and shortcode calls. The corrected pipeline matches the rubric description verbatim.)*

---

## §D. Severity classification

*(v0.2.0 D2 + D9: re-classified under merged AC set. M5 promoted to MUST per D9.)*

- **MUST** — AC-M0-001, AC-M0-002, AC-M1-001, AC-M1-003, AC-M3-001..003, AC-M5-001..004, AC-M7-001..004. (IA integrity, token-port correctness + dark-mode-leak prevention, badge mechanism, ko gate including friend-explainability predicate, final build.)
- **SHOULD** — AC-M1-002, AC-M1-004, AC-M2-001..003, AC-M4-X-001..003, AC-M6-X-001..002. (Asset presence, FROZEN-stamp citation, header/home fidelity, content rubric, per-locale derivation.)
- **MAY** — the 80% `added_in` preferred-form threshold in AC-M7-004(b).

---

## §E. Definition of Done

The SPEC is "done" when M7's MUST ACs all pass AND the sync-phase commit carries the `implemented → completed` transition AND the FROZEN stamp on `moai-brand.css` cites `SPEC-DESIGN-DOCS-V31-001`. Debt (e.g. a section's infographic floor waived with reason) is permitted as SHOULD-level and MUST be recorded in `progress.md` §E.2.
