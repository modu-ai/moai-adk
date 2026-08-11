# plan.md — SPEC-DESIGN-DOCS-V31-001

> Implementation plan. Milestones ordered by decision-reversibility: the decisions most likely to change (IA shape, token port, component vocabulary) lead; the mechanical steps (translation derivation, build verification) close.

---

## §A. Context

This SPEC governs a four-axis simultaneous renewal of `adk.mo.ai.kr` for `v3.1-rc.1`: information architecture, design system, Korean content depth, and sequential 4-locale translation. The full scope and design decisions are in `spec.md`. This plan decomposes the work into 8 milestones (M0–M7), with M4 (Korean content rewrite) decomposed into section-scoped sub-milestones because it is the single largest work item.

The plan honors three load-bearing user decisions (non-re-openable): (1) formal Tier L pipeline, (2) ko-first sequential i18n, (3) light-only design. It also documents two run-phase governance options for the orchestrator to surface at Implementation Kickoff Approval: the Epic-split escape valve (§G.1 of spec.md) and the sub-milestone decomposition of M4.

---

## §B. Known Issues

1. **Token-port cascade.** The live `moai-docs-layout.css`, `moai-docs-theme.css`, and `moai-design.css` contain hardcoded hex values that may diverge from the new token vocabulary. M1 includes an audit step that converts these to `var(--…)` references or verifies byte-compatibility. (Baseline: the live `moai-brand.css` header confirms `--color-bg: #f4f4f4` and `--color-primary: #3d7d5f` are already on the v2 baseline — so the PRIMARY/INK/BG trio is stable; the neutral-ramp delta is the real port surface.)
2. **Home page rewrite blast radius.** The current `layouts/index.html` is a flat section list. The handoff home is a hero + 3-card grid + section grid + book CTA. The rewrite touches `layouts/index.html`, `content/<locale>/_index.md` × 4 locales, and the partials the index composes. This is a structural rewrite, not a restyle.
3. **v3.1 feature pages do not exist yet.** 9 of the 12 features in `spec.md` §F.1 require NEW pages (not edits to existing pages). The IA milestone must finalize the section placement BEFORE these pages are authored, so M0 (IA freeze) gates M4.
4. **Sidebar badge coupling.** The NEW-badge mechanism requires coordinated edits to `layouts/partials/menu.html`, `_meta.yaml` (4 locales), and per-page frontmatter. The coupling is documented in spec.md §G.3; the risk is partial application (badge appears in one locale but not another). REQ-I18N-004 addresses this by requiring the flag to be preserved across locales.
5. **Sequential i18n wall-time.** ko → en → ja → zh with each locale fully verified before the next means the i18n phase (M5–M7) is the longest serial segment. Parallelization is explicitly forbidden by the user decision. The plan does NOT attempt to parallelize; instead it minimizes per-locale wall-time by pre-computing the translation manifest in M3. **Quantification (v0.2.0 D10):** the i18n phase is approximately **3× the wall-clock of a single locale derivation** — M6 (en + ja + zh, each fully verified) is ~3× M4's per-locale wall-clock because en/ja/zh each derive serially from the frozen `ko` source. The M3 manifest pre-computation reduces the per-locale constant (no re-planning between locales) but does NOT collapse the 3× multiplier. The user's ko-first decision is the binding constraint; the 3× is accepted as its cost. This number is recorded so the Implementation Kickoff Approval decision can weigh the Epic-split against a concrete wall-clock envelope.

---

## §C. Pre-flight (before M0)

- [ ] Confirm the design handoff staging path (`/tmp/moai-design-handoff/moai-adk/project/`) is present and complete (6 `.dc.html`, `_ds/`, `assets/`). If the path is stale or the user has re-staged it elsewhere, re-locate before M1.
- [ ] Read `02 Getting Started.dc.html` in full (the handoff README mandates this as the primary design capture).
- [ ] Read the `_ds/.../README.md` tone/copy/color/type sections.
- [ ] Run `cd docs-site && hugo --gc --minify` to capture the pre-SPEC build baseline (must exit 0).
- [ ] Snapshot the current `docs-site/static/*.css` byte-hash (M1 will diff against this).
- [ ] Confirm `moai brand.css` FROZEN comment carries SPEC-DESIGN-DOCSV2-001 (verified at research time — present).

---

## §D. Constraints

(From spec.md §D — restated here only for plan-local reference; the authoritative list is spec.md §D.)

- Light-only (HARD, CLAUDE.local.md §17.1).
- Mermaid TD-only.
- `{{< icon >}}` shortcode; no body emoji.
- `adk.mo.ai.kr` URL whitelist.
- Korean canonical SSOT for content.
- Hugo build exit 0, zero warnings, at every milestone boundary.
- No handoff runtime JS shipped.

---

## §E. Self-Verification

Each milestone's self-verification block lives in `acceptance.md`. The plan-level verification obligation is the cumulative gate at M7 (final build + 4-locale parity + NEW-badge sunset readiness). The `progress.md` §E.2 / §E.3 evidence is populated by `manager-develop` at run-phase.

---

## §F. Milestones

### M0 — IA freeze & redirect map (Priority: High, reversibility: HIGH)

**Scope.** Freeze the new integrated 12-section section list. The 12-vs-13 question is RESOLVED at v0.2.0 (user decision): the handoff's 12-card grid is canonical — the live 13th section `changelog` MOVES TO THE SITE FOOTER (it becomes a footer link, not a sidebar section), and `cost-optimization` retains its own card per the handoff home grid. Emit the canonical `content/ko/_meta.yaml` with the 12-section weight order and the `new_items:` lists. Emit the `vercel.json` redirect entries for the `changelog` slug change (sidebar → footer link) and any other slug change. This milestone is gated HIGH-reversibility because every downstream milestone depends on the IA being frozen.

**Exit criteria.** (1) New `_meta.yaml` (ko) committed with the 12-section list (changelog absent from sidebar) and per-section `new_items`. (2) `vercel.json` redirect diff committed (records the changelog-sidebar→footer redirect and any other slug change; MAY be empty if no other slugs changed). (3) spec.md §F.1 v3.1 feature catalog table cross-checked: every row's "New IA home" column resolves to a real section in the frozen IA.

### M1 — Design-token port & static-asset onboarding (Priority: High, reversibility: HIGH)

**Scope.** Port the v2-renewal tokens from `colors_and_type.css` into `docs-site/static/moai-brand.css` (replace `:root` block) and `docs-site/static/moai-docs-tokens.css`. Copy the 6 mascot pose PNGs into `docs-site/static/mascots/` (filename convention `MoAI-Mascot-<Pose>.png` — already matched by the existing `mascot` shortcode). Copy the 3 logo variants into `docs-site/static/logos/`. Audit `moai-docs-layout.css`, `moai-docs-theme.css`, `moai-design.css` for hardcoded hexes and convert to `var(--…)` references where the value is token-equivalent. Re-stamp the `moai-brand.css` FROZEN header to cite `SPEC-DESIGN-DOCS-V31-001`. Do NOT port the `[data-theme="dark"]` block.

**Exit criteria.** (1) `moai-brand.css` `:root` block matches the handoff `colors_and_type.css` `:root` block (minus the dark block) byte-for-byte on the token vocabulary. (2) Mascot/logo assets present at the documented paths. (3) `hug --gc --minify` still exits 0 (no token-reference breakage). (4) The FROZEN header re-stamp carries this SPEC's ID.

### M2 — Component vocabulary & header/home rewrite (Priority: High, reversibility: HIGH)

**Scope.** Port the component vocabulary (callout, card, pill, terminal, code-card) from the handoff into the existing shortcodes and into new component partials as needed. Rewrite `layouts/partials/site-header.html` to the sticky-blurred-header + search + version-pill + locale-switcher + GitHub pattern (omit the dark-mode toggle per §G/light-only). Rewrite `layouts/index.html` to the Docs Home hero + 3-card grid + section-grid + book-CTA layout. Author the ko `_index.md` to match.

**Exit criteria.** (1) Header partial matches handoff across all 6 prototypes (shared header). (2) Home page renders the hero + 3-card + section-grid + book-CTA structure. (3) Visual diff against `01 Docs Home.dc.html` screenshot confirms structural fidelity (tolerance: section ordering + grid shape; NOT pixel-perfection on inline styles).

### M3 — NEW-badge mechanism & sidebar wiring (Priority: High, reversibility: MEDIUM)

**Scope.** Implement `layouts/shortcodes/new-badge.html`. Extend `layouts/partials/menu.html` to read the page-level `new` / `added_in` flag and the section-level `_meta.yaml` `new_items:` list, rendering the badge. Extend `layouts/_default/single.html` to render the badge beside the `<h1>`. Apply the flag to every page identified in the v3.1 feature catalog. Pre-compute the translation manifest (which pages each locale will receive) so M5–M7 can run without re-planning.

**Exit criteria.** (1) Shortcode renders the badge inline with token-driven styling. (2) Sidebar renders the badge beside every flagged page in `ko`. (3) Page-header renders the badge beside the `<h1>` on every flagged page. (4) The translation manifest is committed as a run-phase artifact (under `.moai/state/` or inside `progress.md` — NOT in the docs-site tree).

### M4 — Korean content rewrite (Priority: High, reversibility: LOW — largest milestone)

**Scope.** Rewrite every `content/ko/**/*.md` page (124 pages) to introductory-book-grade prose per `acceptance.md` §A. This milestone is decomposed into section-scoped sub-milestones so each can be committed independently. Recommended decomposition (section page-counts in parentheses):

- **M4.1** getting-started (8) — onboarding path, highest reader impact, first impression.
- **M4.2** core-concepts (7) — concept-first exposition, defines the shared vocabulary.
- **M4.3** workflow-commands (7) + the new `/moai goal` page + the new factory-mode page.
- **M4.4** utility-commands (12) + cli-reference (21) — reference pages, waivable infographic floor.
- **M4.5** advanced (24, largest) — including new navigator / hierarchical-team / multi-model-audit / autonomy-tier / harness-learning pages.
- **M4.6** claude-code (30, largest-after-advanced) — including CC 2.1.219 rewrite.
- **M4.7** multi-llm (3) + cost-optimization (2) — including profile-matrix rewrite.
- **M4.8** guides (3) + worktree (4) + contributing (1) + changelog (1).

**Critical-path call-out.** M4 is the dominant cost. The recommended sub-milestone order above front-loads the pages a first-time reader meets (getting-started, core-concepts) so the highest-value content is rewritten first even if M4 is later split or paused. Each sub-milestone's AC block is in `acceptance.md` §B.

**Epic-split escape valve (v0.2.0 D8 — quantified trigger).** The single-SPEC path is the committed one (user decision). The Epic-split into a child SPEC `SPEC-DESIGN-DOCS-V31-CONTENT-KO-001` is a documented fallback that fires when ANY of these concrete triggers holds:

- **(T1) Per-sub-milestone turn overshoot.** Any single M4.{x} sub-milestone exceeds **N=40 turns** (the per-milestone soft ceiling; tracked in `progress.md` §E.2 turn-counter column).
- **(T2) Cumulative run-phase turn overshoot.** The cumulative turn count across all of M4.1–M4.8 exceeds **T=200 turns** (≈2.5× the per-milestone ceiling × 8 sub-milestones, with slack).
- **(T3) Repeated sub-milestone failure.** Any single sub-milestone fails its §B.5 AC block **twice with the same root cause** (root-cause match judged by the orchestrator from the manager-develop blocker reports; two different root causes do NOT trigger).

When any of T1/T2/T3 fires, the orchestrator surfaces the split recommendation via `AskUserQuestion` with three options: **split-now** (open `SPEC-DESIGN-DOCS-V31-CONTENT-KO-001` and migrate the remaining sub-milestones), **continue-with-debt** (proceed single-SPEC, record the threshold breach as SHOULD-level debt), **abort** (pause the run-phase for re-planning). The orchestrator — not the agents — owns this trigger evaluation. The split, when taken, is an authorized deviation; it does NOT require a new plan-phase (the child SPEC inherits this SPEC's plan.md as its scope SSOT, per spec.md §G.1).

### M5 — Korean verification gate (Priority: High, reversibility: LOW — MUST gate per v0.2.0 D9)

**Scope.** Verify the Korean locale meets every i18n gate BEFORE any translation begins: `hugo build` exit 0, body-emoji scan clean, TD-only Mermaid, URL blacklist clean, NEW-badge presence on the catalog pages, the §A.4 friend-explainability mechanical predicate passing on every authoring-trail sidecar. This milestone is a GATE, not a work item — its output is a pass/fail decision. **M6 (en derivation) MUST NOT begin until every MUST AC in §B.6 passes; this is the ko→en gate.** (v0.2.0 D9 fix: M5 promoted from SHOULD to MUST; the M6-gating sentence is normative.)

**Exit criteria.** Every MUST AC in `acceptance.md` §B.6 + §C passes for `ko`. If any AC fails, M4 is revisited before M6 begins. The gate is NOT de-duplicated against M7: M5 is the ko-only gate (catches ko-authoring debt before any translation spends), M7 is the 4-locale parity gate (catches inter-locale drift).

### M6 — en + ja + zh sequential derivation (Priority: Medium, reversibility: LOW)

**Scope.** Derive `en` from the verified `ko` SSOT; verify `en` passes every i18n gate; then derive `ja` from the same SSOT; verify `ja`; then derive `zh`; verify `zh`. Each locale is a serial sub-milestone (M6.1 en, M6.2 ja, M6.3 zh) with its own gate. Parallelization is FORBIDDEN by REQ-I18N-003.

**Exit criteria.** All three locales pass the i18n gate. 4-locale page-count parity holds. 4-locale H2-section-count parity holds (zero gaps).

### M7 — Final build, 4-locale parity, NEW-badge sunset readiness (Priority: High)

**Scope.** Final `hugo --gc --minify` across all 4 locales. Bump `hugo.toml [params].version` from `v3.0.2` to `v3.1-rc.1` and `releaseDate` to the release date. Run the `hns-oss-docs-verify` recipe. Confirm the NEW-badge `added_in: "v3.1"` form is used on every catalog page (sunset-sweep-ready). Prepare the sync-phase commit.

**Exit criteria.** (1) `hugo --gc --minify` exit 0 zero warnings. (2) Version bumped. (3) NEW-badge inventory matches the v3.1 catalog exactly (no missing, no extra). (4) 4-locale parity holds. (5) Sync commit ready.

---

## §G. Anti-patterns

- **AP-1 — Parallel i18n.** Translating en/ja/zh in parallel with ko-authoring violates REQ-I18N-003 and risks deriving from a stale or in-flight Korean draft.
- **AP-2 — Pixel-perfect prototype port.** Attempting to reproduce the `.dc.html` inline styles verbatim in Hugo templates produces unmaintainable markup. The prototypes are a visual target; the production code uses the token vocabulary and component partials.
- **AP-3 — Shipping `_ds_bundle.js`.** Vendoring the prototype runtime into the Hugo site expands the CSP surface and adds a runtime failure mode. REQ-BL-003 forbids it.
- **AP-4 — One-shot M4.** Attempting to author all 124 Korean pages in a single run-phase iteration without sub-milestone commits. The decomposition in §F M4 is binding.
- **AP-5 — Dark-mode token leakage.** Porting the `[data-theme="dark"]` block "because it's in the handoff" violates REQ-DS-003 and CLAUDE.local.md §17.1.
- **AP-6 — Hardcoded NEW badges.** Hand-writing `<span class="new-badge">NEW</span>` in page bodies instead of using the shortcode + flag mechanism. This bypasses the sunset sweep.

---

## §H. Cross-references

- spec.md — §C requirements, §F.1 v3.1 feature catalog, §G design decisions.
- acceptance.md — §A book-depth rubric, §B per-milestone AC, §C quality gates.
- research.md — §A handoff analysis, §B content baseline, §C v3.1 feature verification, §D token DIFF.
- `.moai/docs/docs-site-i18n-rules.md` — i18n SSOT.
- `hns-oss-docs-i18n-rules` / `hns-oss-docs-structure-map` / `hns-oss-docs-verify` — harness skills bound at run-phase.
