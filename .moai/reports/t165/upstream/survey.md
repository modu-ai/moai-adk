# diagram-design absorption survey

Source: /tmp/diagram-design — github.com/cathrynlavery/diagram-design, v2.6.1, MIT.
441 files on disk incl. .git (~412 tracked). Root: README.md (40,876 B), CONTRIBUTING.md,
SECURITY.md, CODE_OF_CONDUCT.md, THIRD_PARTY_LICENSES.md, LICENSE, .maintainer-policy.json.

Layout: skills/diagram-design/{SKILL.md, references/, assets/, scripts/}, scripts/ (repo
tooling), commands/ (Claude), prompts/ (Pi), docs/{adr,screenshots,superpowers}, .github/,
.claude-plugin/ .codex-plugin/ .factory-plugin/ .agents/ (host manifests).

## 1. skills/diagram-design/references/ — 52 files, 660 KB

### Design system / skin (1)
| File | Bytes | Contents |
|---|---|---|
| style-guide.md | 8,677 | The single source of truth for the "skin": semantic-role color tokens (paper/paper-2/ink/muted/soft/rule/rule-solid/accent/accent-tint/link) each with light+dark defaults; brand palette mapping (jet-black #2d3142, silver, white-smoke, atomic-tangerine #eb6c36, blue-slate); light→dark inversion rule (same alphas, RGB flipped); 5-color series palette (radar-only); fixed terminal sub-skin (9 `terminal-*` tokens, no pure black); typography roles (Instrument Serif title/callout, Geist node names, Geist Mono technical); stroke/radius/4px-grid tokens; node-type→treatment table (focal/backend/store/external/input/optional/security); customization constraints (WCAG AA, 1-accent, no rainbow, warm-neutral paper). Changing this file changes every future diagram. |

### Primitives (4)
| File | Bytes | Contents |
|---|---|---|
| primitive-icons.md | 106,768 | Generated catalog (by build-icons.py) of 87 icons normalized to 24×24 viewBox + currentColor, with copy-paste SVG snippets — compute, people, network, data, k8s, action, devops + brand outlines/silhouettes. |
| primitive-terminal.md | 3,915 | CLI-chrome full-page variant: fake terminal window (titlebar 3 dots, `$` prompt), all-mono type ramp bumped 1–2px, 1:1 token swap to terminal-* skin; explicitly a second fixed skin not touched by onboarding. |
| primitive-annotation.md | 1,708 | Editorial marginalia: italic Instrument Serif text + dashed Bézier leader + landing dot; max 2 per diagram; solid leaders forbidden (read as flow arrows). |
| primitive-sketchy.md | 1,700 | Hand-drawn variant: feTurbulence+feDisplacementMap filter (baseFrequency 0.02, scale 1.5) applied to shapes group only — text stays outside the filter for legibility; tuning table; "not for dark variants / dense labels". |

### Semantics (1)
| File | Bytes | Contents |
|---|---|---|
| semantic-patterns.md | 10,296 | The behavior axis: 7 semantic patterns (fan-in queue/bottleneck, stage framework w/ semantic slots, unstructured→structured artifact, paired policy-evaluation traces, secure paved road, governance/control catalog, compensating security layers), each with selection triggers, required primitives, complexity budget, anti-patterns, static fallback, and routing to the nearest visual type. Pattern selects semantics; the visual type owns layout. |

### Workflow references (8)
| File | Bytes | Contents |
|---|---|---|
| output-spec.md | 11,194 | The "four dials" output contract: format (html/svg/png/html+png), 9 size presets (doc-inline 960×600 … print-letter, fit) each 4px-divisible with per-class type ramps, detail (faithful ≤24 / balanced ≤12 / simplified ≤7 nodes), audience (engineer/mixed/executive), plus the fidelity ledger (what was merged/collapsed/dropped). |
| onboarding.md | 13,804 | Generate a skin from a design source: URL (agent-browser/fetch), installed skill, or local folder → extract palette+fonts → map to semantic roles → propose style-guide.md diff → write with approval → offer named profile save. |
| profiles.md | 14,221 | Client profiles: named style-guide snapshots in ~/.diagram-design/profiles/, strict slug grammar, metadata header, project-root `.diagram-design` marker with marker-first resolution, schema check + default backfill on load. |
| animation.md | 11,815 | Motion contract: 4 modes (none/reveal/step/loop) via data-motion-mode; static-first enhancement contract; semantic motion primitives; keyboard controls; reduced-motion/color/a11y rules. |
| import-drawio.md | 13,066 | Draw.io import pipeline: run drawio_extract.py first (handles compressed containers), digest, four dials, pick target type, build semantic model, fidelity ledger. |
| import-mermaid.md | 8,913 | Mermaid import: same shape; source text treated as untrusted data. |
| export.md | 7,894 | SVG/PNG export via Playwright; PNG at device_scale_factor=2; sizing rules; refuses gallery/index.html and SVG-less files. |
| doctor.md | 5,193 | Environment diagnostics contract: required checks, warn/fail classification, compact summary + optional JSON output, strictly read-only. |

### Chart/type specs (39 type-*.md) — each: layout conventions, complexity budget, anti-patterns, examples
architecture 5,001 · bar 13,576 (incl. dumbbell + slopegraph variants) · data-flow 21,284 (deterministic layout formulas) · db-schema 4,865 · dependency 4,187 · deployment 4,184 · dp-integration 22,247 · dp-security-matrix 19,086 · er 1,495 · fishbone 5,710 (bone math) · flowchart 1,079 · gantt 2,468 · high-level 26,685 (chevron/branching rules) · it-state 29,984 · journey 5,075 (sentiment curve geometry) · kanban 4,302 (card-state vocabulary) · layers 1,535 · line 17,407 · loop 12,609 (deterministic geometry) · medallion 22,226 · nested 1,392 · org-chart 3,082 · polar 7,767 (angle=category, radius=quantity) · process 28,588 · pyramid 1,783 · quadrant 6,254 (incl. consultant 2×2 special) · radar 5,592 (series palette) · sankey 8,952 (scale rule) · scatter 2,299 · sequence 6,686 (message kinds, alt/opt/loop fragments) · state 1,112 · story-map 4,827 · swimlane 991 · timeline 1,065 · tree 1,248 · treemap 7,798 (honest-data rule) · uml-class 4,857 (arrowhead semantics) · venn 1,654 · wardley 5,748.
Largest (high-level, it-state, process, dp-integration, dp-security-matrix, medallion, data-flow, loop, bar, line) carry full "Inputs — the parameter contract / Layout formulas — deterministic geometry" sections, i.e. numeric layout algorithms, not vibes.

## 2. scripts/ inventories

Root scripts/ — 34 files, 1.1 MB. Skill scripts/ — 3 files (shipped to installed agents):
- skills/.../scripts/drawio_extract.py (31,364 B) — draw.io IR extractor (4 container formats incl. base64-deflate compressed).
- skills/.../scripts/mermaid_extract.py (47,650 B) — Mermaid IR extractor, all grammars.
- skills/.../scripts/self_check.py (16,772 B) — distilled output checker agents run on their own generated diagrams.

### verify-*.py (15) — one-line property each
| Script | Bytes | Checked property |
|---|---|---|
| verify-sankey.py | 68,304 | Ribbon width = volume and volume is conserved: inflow=outflow per node, ribbon width vs printed numbers, no silent evaporation. |
| verify-slopegraph.py | 33,443 | Both axes share one linear transform (shared scale), values sit exactly on-scale (no jitter), labels match endpoints. |
| verify-polar.py | 32,627 | Machine-readable polar contract: angle=ordered cyclic categories, radius=linear quantity, geometry/label/URL invariants. |
| verify-mermaid-import.py | 30,237 | Mermaid import: extractor over all grammars, multi-block MD, adversarial labels, trust boundaries, resource caps, doc/command wiring. |
| verify-treemap.py | 20,750 | Cell's drawn-area share matches its printed value (relative error), and labels/markers stay inside their cell. |
| verify-docs-sync.py | 19,153 | 10 classes of routing/docs drift: SKILL.md description keeps every type's lexical hook, gallery reaches every example, README tree files exist, reference links resolve, command/prompt↔reference parity, plugin manifests repeat the description verbatim. |
| verify-semantic-motion.py | 17,820 | Semantic-pattern routing table integrity + animated-example structure + SKILL.md 40 KB byte cap; never executes JS. |
| verify-drawio-import.py | 20,203 | Drives real extractor against the fixture in all 4 container formats; references stay in sync. |
| verify-motion.py | 21,921 | Motion contract with zero deps: mode/action vocab, controller byte-equality vs template-motion.html (ADR 0001's second enforcement). |
| verify-dumbbell.py | 10,801 | The two bar-chart dumbbell rules as executable math: domain resolution never uses observed extremes (exhaustive sign partition, no ÷0 on all-zero) + connected-pair geometry. |
| verify-plugin-package.py | 18,519 | Synchronized version bumps across the 3 native manifests, no drift/deletion/unsafe marketplace paths, version advances from base ref. |
| verify-doctor.py | 13,297 | The doctor itself: read-only env checks, warn/fail classification, --strict/--json contract. |
| verify-screenshot-freshness.py | 3,807 | Fails when any canonical example or committed PNG drifts from docs/screenshots/manifest.json digests. |
| verify-sequence-oauth.py | 5,476 | Structural + skin-lint gates for the sequence combined-fragment (alt/opt/loop) work and its oauth examples. |
| verify-geometry.py | 5,224 | No label mask is clipped by a node declared later in the document (paint-order rule; node ≥60×40, mask 20–200×8–14 heuristics). |

### Tooling scripts
- lint-skin.py (27,040 B) — lints new/changed examples against the current skin: colors, fonts, accessible-SVG contract (resolving accessible name, title/desc placement), SHA-256 pin of the motion controller, rejects remote assets / CSS @import / non-fragment url() / executable attributes; `--all --baseline` skips the 20 legacy files listed in lint-skin-baseline.txt.
- lint-render.py (33,899 B) — renders examples in headless Chromium and checks *paint*, not source: viewport clipping measured by screenshot-diffing with overflow released in stages, collapsed SVGs, page overflow, JS errors; no golden images; network cut at the resolver; `--fonts` allows only the two Google Fonts hosts over HTTPS; `--self-test` asserts 23 cases (over half must-not-flag).
- build-icons.py (28,271 B) — fetches Tabler (MIT), Simple Icons (CC0), log-z/logos (MIT), Devicon (MIT) SVGs; normalizes to 24×24 + currentColor (strips embedded `<style>`, rewrites class fills); emits primitive-icons.md + icons.html. Generated files are committed; end users never run it.
- fix-mojibake.py (2,219 B) — repairs cp1252↔UTF-8 mojibake sequences (em-dash→'â€"' etc.) in place without touching legitimate UTF-8.
- render-canonical-screenshots.py (2,582 B) — renders the 39 canonical minimal-light examples to PNG via Playwright and records source+PNG sha256 into docs/screenshots/manifest.json (renderer: playwright-chromium, scale 2.0, capture first-svg, fonts.ready gate).
- screenshot_catalog.py (1,584 B) — shared helpers (slugs, paths, sha256, PNG dimensions) for the screenshot manifest.
- bump-plugin-version.py — synchronized semver bump across manifests (driven by .maintainer-policy.json).
- 16 test-*.py (263 KB total) — adversarial suites for the checkers themselves (sankey 56,577 B; slopegraph 27,131 B; treemap 26,424 B; polar 20,443 B; motion 16,838 B; docs-sync 17,307 B; plugin-package 16,711 B; lint-a11y 26,461 B; plus small ones for geometry, dumbbell, drawio, doctor, self-check, build-icons ×2, fix-mojibake).

## 3. docs/adr/ — 8 ADRs

- **0001 Static by default; one pinned controller for motion** (v2.3): output is script-free (`data-motion-mode="none"`) unless motion is requested; when it is, exactly one `<script data-diagram-controls>` whose body must byte-match the reviewed `assets/template-motion.html` — enforced by SHA-256 in lint-skin.py AND string equality in verify-motion.py AND self_check.py. Absorption relevance: motion policy becomes a review-once-and-propagate artifact; hand-edited controllers fail every gate.
- **0002 Semantic patterns never expand the visual-type taxonomy** (v2.3, amended v2.6): behavior is a separate axis; the 7 patterns route to the nearest existing visual type; the visual-type count moves only for a genuinely new layout grammar, and two counters (verify-semantic-motion.py, verify-docs-sync.py) hardcode it — amendments moved it 27→28→38→39, always in the same PR as the ADR. Relevance: taxonomy growth is a governed, test-enforced decision, not drift.
- **0003 `reveal` is the only sanctioned autoplay** (v2.3): reveal may run once on load then stays complete; never restarts on re-entry; step is user-initiated, loop is CSS-scoped, reduced-motion/no-JS always shows the complete static frame. No autoplay heuristic needed in verify-motion.py because the pinned controller is the only code that can start a run.
- **0004 SKILL.md byte cap and the trigger-rich description** (v2.3, raised): the frontmatter description must name every visual type (lexical routing hooks — verified by verify-docs-sync.py); MAX_SKILL_BYTES = 40,000 (verify-semantic-motion.py); when near the cap, cut body prose or move to references/ — never the description; cap measured with core.autocrlf=false pinned in CI.
- **0005 Label placement is verified geometrically, not by review** (v2.3): 9 shipped examples had label masks clipped by later-painted nodes and every prose gate passed; verify-geometry.py uses document paint order as the criterion + adversarial tests in both polarities; threshold history (mask width cap 120→200 for CJK/mono plates; height cap kept at 14 to avoid misclassifying ~80 shipped rects) is recorded as a caution.
- **0006 Client profiles use marker-first resolution** (v2.4): named full style-guide snapshots in ~/.diagram-design/profiles/; a project-root `.diagram-design` marker selects a validated slug read directly (no copy races, survives plugin updates); rejected alternatives (in-install storage, token-merge, central path index) are listed with reasons.
- **0007 Ten new layout grammars (28→38)** (v2.5.10): the bar for new types demonstrated via coverage audits vs the Mermaid taxonomy + practitioner requests; a per-type table of "the grammar that did not exist / why the nearest type fails"; an explicit rejected list (C4, mindmap, pie, use-case, git graph…) with reasons; consequences incl. §6 elbow exemptions for 4 types, ER scope narrowing vs db-schema, the byte cap becoming binding, and canonical screenshots shipping later.
- **0008 Native host manifests share one plugin root** (v2.5.14): Claude/Codex/Factory each get the smallest native manifest; every marketplace resolves to the repo root; all hosts reuse skills/diagram-design/ + commands/; the package verifier rejects drift/deletion/unsafe paths/non-advancing versions. Adding a host = native metadata + gate coverage + docs + synchronized bump, never copying the skill.

## 4. commands/ (5, Claude) + prompts/ (4, Pi)

Thin routers that delegate to references ("treat that reference as the source of truth — don't reimplement"):
- **export-diagram**: HTML → .svg + .png next to source (PNG @2×); flags --svg-only/--png-only/--scale {1,2,3}/--output; refuses missing file, the multi-SVG gallery, SVG-less files; surfaces Playwright install instructions verbatim and never auto-installs. (prompts/ version adds Pi skill-location resolution.)
- **import-mermaid**: redraw .mmd/.mermaid/fenced-Markdown as an editorial diagram; always runs the installed skill's mermaid_extract.py first; four dials (format/size/detail/audience) + --type/--diagram/--variant; refuses to render Mermaid or carry over its layout/theme/fonts; treats source text and digest as untrusted; reports the fidelity ledger.
- **import-drawio**: same shape for .drawio/.drawio.png/.drawio.svg (compressed containers — never read raw); --page selection; zoning rules (faithful >9 nodes ⇒ zone; >24 ⇒ split); never carries over source coordinates/colors/fonts.
- **doctor**: one-shot read-only environment diagnostics per references/doctor.md; --strict/--json; never installs or edits.
- **profile** (commands only): save/load/list/show/update/reset/delete client profiles per references/profiles.md; marker-first resolution; confirm-before-overwrite; verifies writes by re-reading. prompts/ has doctor, export-diagram, import-mermaid, profile (no import-drawio for Pi).

## 5. skills/diagram-design/assets/ — 143 HTML files, 2.1 MB

- **136 example-*.html**: 39 canonical types × 3 variants = 117, plus 19 specials — datalake ×3, high-level-vertical ×3, sequence-oauth ×3 (combined-fragment showcase), slopegraph ×3 (bar/line variant), import-drawio, import-mermaid, quadrant-consultant, loop-terminal, and 3 animated: queue-animated (28.2 KB), policy-trace-animated (25.5 KB), paved-road-animated (28.3 KB).
- **The 39-type catalogue** (matches references/type-*.md 1:1): architecture, IT current-state, flowchart, sequence, state machine, ER/data model, timeline, swimlane, quadrant, radar/spider, loop/flywheel, nested, tree, org chart, layer stack, Venn, pyramid/funnel, bar chart, treemap, line chart, Gantt, scatter, high-level, process, medallion, data flow, DP integration, DP security matrix, sankey, fishbone, Wardley map, kanban, user journey, deployment, dependency graph, UML class, story map, database schema, polar. (Dumbbell and slopegraph are bar/line *variants*, not types.)
- **3-variant convention**: `<type>.html` = minimal light, `<type>-dark.html` = minimal dark (token inversion), `<type>-full.html` = full-editorial (title block, legend, richer chrome; typically ~35% larger). Dark files are +2–3 KB over light; full are +3–5 KB.
- **5 templates**: template.html (3,253 B, minimal light), template-dark.html (3,130), template-full.html (22,518), template-motion.html (19,863 — contains the reviewed pinned controller; the only sanctioned script), template-terminal.html (4,934).
- **icons.html** (108,775 B): visual gallery of the 87 normalized icons (24×24, currentColor) from primitive-icons.md.
- **index.html** (15,351 B): the tabbed gallery flipping all 39 types across light/dark/full-editorial; published to GitHub Pages.

## 6. Plugin manifests — multi-host packaging

One skill, four hosts, zero duplication (ADR 0008): `.claude-plugin/{marketplace.json, plugin.json}` (Claude Code; the canonical 40-type-naming description, v2.6.1, owner Cathryn Lavery), `.codex-plugin/plugin.json` (adds `skills: ./skills/` + an `interface` block with displayName, category "Productivity", capabilities Read/Write, default prompts, brandColor #b5522a), `.factory-plugin/{plugin.json, marketplace.json}` (Factory Droid; minimal native manifest), and `.agents/plugins/marketplace.json` (Pi; local source path with installation/authentication policy). Every marketplace entry resolves `"source": "./"` — the repository root — so all hosts share skills/diagram-design/ and commands/ verbatim. verify-plugin-package.py (driven by .maintainer-policy.json: semver, require_synchronized, require_unique_across_open_prs, bump via bump-plugin-version.py) rejects manifest drift, deletions, unsafe marketplace paths, or a release that doesn't advance all three versions together.

## 7. Root README.md (40,876 B)

- **Claimed features**: 39 visual types × 3 static variants, standalone offline HTML (no build/JS/external images), brand onboarding from URL/skill/folder with automatic contrast checks and a fidelity receipt, named client profiles per project, semantic patterns + optional accessible motion, draw.io/Mermaid import at chosen size/detail/audience, SVG/PNG export, live gallery on GitHub Pages.
- **"Skins" concept**: the skin is style-guide.md — one file of semantic-role tokens that every diagram draws from; regenerate it via onboarding.md (extract palette+fonts → map to semantic roles → approved diff), persist variants as profiles; terminal and sketchy are opt-in second registers, not skins.
- **Validation story**: 16 test-*.py suites (263 KB) adversarially test the checkers themselves; repo-wide gates listed in "Contributing / skin lint" (lint-skin --all --baseline, verify-semantic-motion --markdown-only/--example-only, verify-motion --shipped, lint-render --all, verify-geometry/treemap/docs-sync/drawio/mermaid, self_check); 38 CI steps in .github/workflows/ci.yml across Linux/Windows/macOS; .maintainer-policy.json lists 29 local gate commands.
- **Determinism/pinning**: output opens double-clicked, offline, with no network beyond Google Fonts; no golden images anywhere; lint-render cuts network at the browser resolver (covers WebSockets), pins Playwright + its Chromium build instead of latest; default font-metrics run is deterministic/machine-independent on fallback faces (--fonts measures real webfonts locally); screenshots carry source+PNG sha256 digests in docs/screenshots/manifest.json with a pinned renderer spec; motion timing is deterministic with a byte-pinned controller.
- Also notable: "What loads when" table (progressive disclosure — one type reference per request), "It's working if…" acceptance list, and a "When *not* to use this skill" section.

## 8. scripts/fixtures/ + scripts/vendor/icons/

- **fixtures/ (4 files)**: sample-architecture.drawio (5,173 B — driven through all four draw.io container formats by verify-drawio-import.py), sample-flowchart.mmd (452 B), sample-adversarial.mmd (726 B — adversarial labels/trust boundaries for the Mermaid gates), sample-readme-with-mermaid.md (532 B — multi-block Markdown extraction case).
- **vendor/icons/ (5 sets, ~102 SVGs)**: tabler/ (58 stroked UI/tech icons, MIT), simple/ (25 CC0 brand silhouettes: airflow, hive, nifi, superset, gitea, trino, jupyter…), logz/ (6 MIT website logos — mysql, redis, starrocks… — 100×100 with embedded styles that build-icons.py strips and rewrites to currentColor), devicon/ (5, incl. rstudio-plain, spss-plain), url/ (8 fetched-by-URL exceptions: dagster, hop, pentaho, sas, stata…). All redistributions are normalized and documented in THIRD_PARTY_LICENSES.md with upstream links.

## DISTINCTIVE MECHANISMS

1. **Per-diagram-type geometric verifiers** — each quantitative type gets an executable invariant, because "every way of breaking the claim renders perfectly": sankey volume conservation (scripts/verify-sankey.py), treemap relative area fidelity + label fit (scripts/verify-treemap.py), slopegraph shared-scale/no-jitter (scripts/verify-slopegraph.py), polar encoding contract (scripts/verify-polar.py), dumbbell domain-resolution math (scripts/verify-dumbbell.py).
2. **Paint-order-aware label-clip checker** — verify-geometry.py reports a label mask overlapped by a node *declared later in the document*, using the fixed §5 paint order as the legality criterion, with documented shape-heuristic thresholds and their revision history (ADR 0005).
3. **Single-source skin with semantic-role tokens** — style-guide.md is the only place hex lives; a light→dark inversion rule (same alphas, RGB flipped); onboarding.md regenerates the whole skin from a URL/skill/folder with contrast gates and a fidelity receipt (skills/diagram-design/references/style-guide.md, onboarding.md).
4. **Client profiles with marker-first resolution** — named style-guide snapshots in ~/.diagram-design/profiles/ + a project `.diagram-design` marker selecting a slug read in place; survives plugin updates and avoids parallel-workspace races; rejected alternatives documented (references/profiles.md, ADR 0006, commands/profile.md).
5. **Static-by-default + byte-pinned single motion controller** — one reviewed `<script data-diagram-controls>` in template-motion.html, enforced three ways (SHA-256 in lint-skin.py, string equality in verify-motion.py, again in shipped self_check.py); reveal-once-only autoplay policy (ADR 0001, 0003; references/animation.md).
6. **Two-axis taxonomy: semantic patterns vs visual types** — 7 behavior patterns route onto the 39 layout grammars; the type count is hardcoded in two verifiers and moves only via ADR amendment (references/semantic-patterns.md, ADR 0002/0007, verify-semantic-motion.py).
7. **Checkers-that-test-the-checkers** — 16 adversarial test-*.py suites (263 KB) covering both polarities (must-flag and must-not-flag), plus lint-render.py --self-test with 23 cases and a byte-identical-DOM assertion (scripts/test-*.py).
8. **Rendered-paint linting without golden images** — lint-render.py screenshots each SVG as-authored and again with overflow released, diffing pixels to detect clipping that getBoundingClientRect cannot see; staged releases so wrappers can't mask spill; network cut at the resolver; Playwright/Chromium pinned (scripts/lint-render.py).
9. **Docs-as-code drift gates** — verify-docs-sync.py's ten drift classes include: SKILL.md frontmatter description must retain every type's lexical routing hook, plugin manifests must repeat that description verbatim (four copies of one sentence), gallery↔example bidirectional reachability, README architecture-tree file existence (scripts/verify-docs-sync.py, ADR 0004).
10. **SKILL.md byte cap + trigger-rich description trade-off made explicit** — 40,000-byte cap with "cut prose, never the description" priority order, both machine-enforced (ADR 0004; verify-semantic-motion.py).
11. **Icon build pipeline with license normalization** — build-icons.py pulls four icon sets under three licenses, normalizes to 24×24 currentColor (stripping embedded `<style>` and rewriting class fills), commits both artifacts, and THIRD_PARTY_LICENSES.md maps every redistribution to its upstream license; vendored override sets for icons upstream doesn't carry (scripts/build-icons.py, scripts/vendor/icons/, THIRD_PARTY_LICENSES.md).
12. **Importers with IR extractors, resource caps, and fidelity ledgers** — drawio_extract.py (31 KB; 4 container formats incl. compressed) and mermaid_extract.py (47 KB; all grammars, adversarial inputs); four-dial output contract (format/size/detail/audience, references/output-spec.md) and a mandatory fidelity ledger reporting what was merged/collapsed/dropped; source text treated as untrusted.
13. **Screenshot freshness manifest** — render-canonical-screenshots.py + screenshot_catalog.py + verify-screenshot-freshness.py keep 41 canonical PNGs digest-bound (source sha256 + PNG sha256 + pinned renderer spec) so README images can't silently drift from examples (docs/screenshots/manifest.json).
14. **ADR-governed design decisions** — 8 short ADRs with status/amendments, explicitly positioned as "read before relitigating; add one when you settle a new policy," and cross-referenced from verifiers that enforce them (docs/adr/).
15. **Single-root multi-host packaging with synchronized version gate** — 4 host manifest families resolving to one repo root, byte-identical shared metadata, and verify-plugin-package.py + .maintainer-policy.json enforcing synchronized semver bumps, unique versions across open PRs, and a pinned local gate list (ADR 0008, .maintainer-policy.json).

Smaller but notable: legacy-skin baseline exemptions (scripts/lint-skin-baseline.txt, 20 files), mojibake repair tooling (scripts/fix-mojibake.py + test), a shipped distilled self-checker for installed agents (skills/diagram-design/scripts/self_check.py), a first-run style-guide gate (SKILL.md §0), and design-doc + plan artifacts retained under docs/superpowers/ for the polar-chart type.
