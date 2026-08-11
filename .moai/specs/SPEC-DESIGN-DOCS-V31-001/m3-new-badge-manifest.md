# M3 NEW-badge Target Manifest — SPEC-DESIGN-DOCS-V31-001

> **Purpose.** Input artifact for M3 (NEW-badge mechanism & sidebar wiring).
> Maps every v3.1 feature in `spec.md` §F.1 to its target section + page slug,
> the page status (NEW / REWRITE / MIGRATE / MENTION), and the badge scope
> (page-level frontmatter `added_in: "v3.1"` vs section-level `_meta.yaml`
> `new_items:` list). M3 consumes this to know where to render badges; M4
> (Korean content rewrite) consumes the same target slugs to know where to
> author.
>
> **Source of truth.** `spec.md` §F.1 (12 v3.1 features). Verified against
> `docs-site/content/ko/` actual page inventory at M0 execution time
> (2026-08-11).
>
> **Scope binding.** M0 owns IA freeze only. This manifest is **M3 prep** —
> the page slugs are the IA decisions M0 freezes; the NEW-badge *mechanism*
> (shortcode, menu.html wiring, page-header rendering) is M3 work and is NOT
> implemented by this manifest. M0 does not populate `new_items:` lists.

---

## §A. The 12 v3.1 features — target slug map

Page status legend:
- **NEW** — page does not exist; M4 authors it from scratch.
- **REWRITE** — page exists; M4 rewrites content (badge applies if v3.1-origin).
- **MIGRATE** — page exists at a different path; M4 moves it to the new IA home.
- **MENTION** — no dedicated page; an existing page gets a section/paragraph.

Badge scope legend:
- **page** — frontmatter `added_in: "v3.1"` on this single page; sidebar +
  page-header both render the badge (the dual mechanism in spec.md §G.3).
- **section** — page slug listed in its section's `_meta.yaml` `new_items:`
  array; M3's extended `menu.html` reads this for sidebar badge rendering
  without requiring every page to carry frontmatter.

| # | Feature | Target section | Target page slug | Page status | Badge scope | Notes |
|---|---|---|---|---|---|---|
| 1 | `/moai goal` (infinite-duration, REAL bounds) | `workflow-commands` | `workflow-commands/moai-goal.md` | **NEW** | page + section | A `moai-goal.md` exists today under `utility-commands/`; the workflow-commands page is a distinct NEW page covering the goal directive (the utility-commands page stays as a command ref). Cross-reference, do not duplicate. |
| 2 | Factory Mode | `advanced` (+ `workflow-commands` mention) | `advanced/factory-mode.md` | **NEW** | page + section | Mention in `workflow-commands/moai-run.md` (existing, REWRITE-add-section). |
| 3 | BAS Navigator (3-tier codemap sync) | `advanced` (+ `utility-commands` command ref) | `advanced/bas-navigator.md` | **NEW** | page + section | Command ref in `utility-commands/moai-codemaps.md` (existing, REWRITE-add-section). |
| 4 | `manager-lead` hierarchical team | `advanced` | `advanced/manager-lead.md` | **NEW** | page + section | |
| 5 | Multi-model audit convergence | `advanced` | `advanced/multi-model-audit.md` | **NEW** | page + section | |
| 6 | `MOAI_AUTONOMY_TIER` mode tiers | `advanced` (+ `cost-optimization` mention) | `advanced/autonomy-tier.md` | **NEW** | page + section | Mention in `cost-optimization/prompt-caching.md` (existing, REWRITE-add-paragraph). **Shared page** with feature #10 — see below. |
| 7 | `profile` matrix (`max/medium/low`) | `multi-llm` (+ `cli-reference` command ref) | `multi-llm/profile-matrix.md` | **MIGRATE** | page (no badge — feature is v3.0.x-era content reorganization, not a v3.1-new capability) | **MIGRATE**: page exists today at `advanced/profile-matrix.md`; M4 moves it to `multi-llm/`. Command ref `cli-reference/profile.md` (existing, REWRITE). |
| 8 | Per-agent model enforcement | `multi-llm` (mention, no dedicated page until SPEC completes) | `multi-llm/model-policy.md` | **MENTION** | section only (no new page) | Existing `model-policy.md` gets a v3.1 paragraph. SPEC-AGENT-MODEL-ENFORCE-001 is `in-progress` at M0 authoring — badge placement deferred until that SPEC completes (spec.md §F.1 row 8 caveat). |
| 9 | Dynamic Workflows / `ultracode` | `advanced` (existing page, rewrite) | `advanced/ultracode-workflows.md` | **REWRITE** | page (badge applies — v3.1 behavior change is material: CC 2.1.219 nesting, depth-3 default) | Page exists today; M4 rewrites for CC 2.1.219 alignment. |
| 10 | Stop-chain / per-edit hook consolidation | `advanced` (autonomy-tier page mentions) | `advanced/autonomy-tier.md` (shared with #6) | **NEW** (shared) | page + section (same page as #6) | The autonomy-tier page is the single home for BOTH `MOAI_AUTONOMY_TIER` (feature #6) AND the stop-chain/per-edit-hook consolidation narrative — they are co-located because autonomy tiers and the stop-chain semantics are inseparable at the user-facing level. M4 authors one page, lists it once in `new_items:`. |
| 11 | Agent body diet + parallel batching | `claude-code` (mention) | `claude-code/agentic/*.md` (TBD subpage) | **MENTION** | none (no badge — indirect user-facing change surfaces as claude-code narrative, not a discrete v3.1 feature page) | M4 decides which claude-code subpage carries the narrative; no dedicated page. |
| 12 | CC 2.1.219 upstream alignment | `claude-code` (rewrite) | `claude-code/foundations/*.md` + section-wide | **REWRITE** | section (badge applies at section level — multi-page rewrite, not one page) | M4 rewrites the claude-code section for CC 2.1.219 (nesting enabled by default, subagent depth semantics, deprecated `mode` parameter, etc.). Badge scope is section-level because the change spans the section. |

---

## §B. Aggregated NEW-badge targets (M3 wiring input)

M3 reads this section to know: "where do I add `added_in: "v3.1"` frontmatter, and where do I populate `_meta.yaml` `new_items:` lists?"

### §B.1 Page-level frontmatter targets (`added_in: "v3.1"`)

8 pages — M3 adds `added_in: "v3.1"` to the frontmatter of each, M4 authors/rewrites the body:

1. `content/<locale>/workflow-commands/moai-goal.md` (NEW)
2. `content/<locale>/advanced/factory-mode.md` (NEW)
3. `content/<locale>/advanced/bas-navigator.md` (NEW)
4. `content/<locale>/advanced/manager-lead.md` (NEW)
5. `content/<locale>/advanced/multi-model-audit.md` (NEW)
6. `content/<locale>/advanced/autonomy-tier.md` (NEW — shared by features #6 + #10)
7. `content/<locale>/advanced/ultracode-workflows.md` (REWRITE — existing page)

(`profile-matrix.md` is intentionally absent from this list — see §A row 7 rationale.)

### §B.2 Section-level `new_items:` targets (in `<section>/_meta.yaml`)

M3 populates these after M4 creates the pages. M0 leaves these UNPOPULATED — `new_items:` lists require the target pages to exist before they can be referenced, and page creation is M4 work.

**`advanced/_meta.yaml`** (5 new items — the largest delta):
```yaml
new_items:
  - factory-mode
  - bas-navigator
  - manager-lead
  - multi-model-audit
  - autonomy-tier
```

**`workflow-commands/_meta.yaml`** (1 new item):
```yaml
new_items:
  - moai-goal
```

**`claude-code/_meta.yaml`** (section-level badge, no per-page `new_items:` for the CC 2.1.219 rewrite — M3 implements a `section_is_new:` flag for this case if it's in scope; otherwise the section's `_index.md` carries `added_in: "v3.1"`):

This case (a multi-page rewrite that's "new" at the section level but not at every individual page) is the **M3 schema design question** M3 must resolve when implementing the mechanism. Two viable approaches:
- **(a)** add `added_in: "v3.1"` to the section's `_index.md` frontmatter (treats the section landing page as the badge anchor);
- **(b)** extend `_meta.yaml` with a section-level `added_in: "v3.1"` field that `menu.html` reads at the section-summary level (broader scope, renders the badge beside the section name in the sidebar).

M3 owns this design decision; M0 records the question, does not pre-decide it.

---

## §C. IA-adjacent findings (orphan page flag — NOT an M0 blocker)

M0's job is to freeze the IA and remove changelog from the sidebar. While
inspecting the page inventory for this manifest, two observations surfaced
that are **NOT M0 scope** but should be tracked:

### §C.1 `profile-matrix.md` lives at `advanced/`, not `multi-llm/`

Spec §F.1 row 7 says the profile matrix's IA home is `multi-llm`. The page
exists today at `docs-site/content/ko/advanced/profile-matrix.md`. M4 must
**migrate** the page to `docs-site/content/ko/multi-llm/profile-matrix.md`
across all 4 locales — this is a content-level migration (move + sidebar entry
update in `data/menu/main.yaml` if one exists today), gated by M4.

**M3 implication**: the badge (if any — see §A row 7) attaches to the
`multi-llm/profile-matrix.md` path post-migration. M3 must run AFTER the M4
migration of this page lands, OR be aware the slug will change mid-SPEC.

### §C.2 `moai-goal.md` exists at `utility-commands/`, target is `workflow-commands/`

Spec §F.1 row 1 says the `/moai goal` IA home is `workflow-commands (new page)`.
A `moai-goal.md` already exists at `utility-commands/moai-goal.md` (a
utility-class command ref). The NEW page at `workflow-commands/moai-goal.md`
is the goal-directive narrative page (workflow-class), distinct from the
existing utility-class page. The two coexist by class:
- `workflow-commands/moai-goal.md` — narrative page covering the goal
  directive, autonomy, progression-mode axis (NEW, badge-eligible).
- `utility-commands/moai-goal.md` — command ref (existing, no badge, REWRITE
  to add a cross-reference link to the new workflow page).

**M4 implication**: author the workflow-commands page; update the
utility-commands page to cross-reference it. Both pages stay.

---

## §D. `_meta.yaml` schema readiness for `new_items:`

The orchestrator's M0 task asked to "verify the schema supports `new_items:`"
for M3. Findings:

- The current `_meta.yaml` files use a free-form YAML map keyed by section
  slug, with each section carrying `title:` + `weight:` sub-keys. Example:
  ```yaml
  advanced:
    title: "Advanced"
    weight: 100
  ```
- Hugo's data loader parses `_meta.yaml` as arbitrary YAML; adding a
  `new_items:` array sub-key is **YAML-valid** and Hugo-geekdoc does NOT
  reject unknown sub-keys (verified by inspection — the current files carry
  only `title:` + `weight:` but nothing in the geekdoc partial or Hugo
  config restricts additional keys).
- The `menu.html` partial does NOT read `new_items:` today (it ranges over
  `data/menu/main.yaml` exclusively — see `layouts/partials/menu.html` line
  10: `{{- range hugo.Data.menu.main.main }}`). M3 must extend `menu.html`
  to consult `_meta.yaml`'s `new_items:` if section-level badges are to
  render in the sidebar.
- **Conclusion**: schema supports `new_items:`; M3 implementation is a
  `menu.html` extension, NOT a `_meta.yaml` schema change. No M0 blocker.

---

## §E. AC-M0-* traceability

This manifest satisfies M0 plan.md exit criterion (3): "spec.md §F.1 v3.1
feature catalog table cross-checked: every row's 'New IA home' column
resolves to a real section in the frozen IA."

Cross-check result: **12/12 rows resolve**. The frozen IA has 12 sections
(getting-started, core-concepts, workflow-commands, utility-commands,
cli-reference, claude-code, multi-llm, cost-optimization, guides, worktree,
advanced, contributing); every §F.1 row's target section is in this set.
The `changelog` section is no longer a target (footer-only per M0 user
decision), which does NOT affect §F.1 because no v3.1 feature was assigned
to changelog.

---

## §F. Hand-off to M3

M3 reads this manifest and:
1. Implements `layouts/shortcodes/new-badge.html` (badge shortcode).
2. Extends `layouts/partials/menu.html` to read both page-level
   `added_in:` frontmatter AND section-level `_meta.yaml` `new_items:`,
   rendering the badge in the sidebar.
3. Extends `layouts/_default/single.html` to render the badge beside `<h1>`
   on pages with `added_in: "v3.1"`.
4. Populates `new_items:` lists in each section's `_meta.yaml` AFTER M4
   creates the target pages (or coordinates with M4 milestone ordering).
5. Resolves the §B.2 design question (section-level badge mechanism for the
   claude-code CC 2.1.219 rewrite).

M3 does NOT re-decide the IA — the slugs in this manifest are frozen by M0.
