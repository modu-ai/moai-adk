# Implementation Plan — SPEC-V3R6-WORKFLOW-DOCS-001

Tier M · card t273 · branch `WT-workflow-docs` (base `db1362739`) · touches `docs-site/` + 4 README files only.

## §A Context

Scope is fixed by `.moai/reports/t273/gap-map.md` GAP-1..GAP-4 (see spec.md §A). Execution vehicle: the `hns-oss-docs` harness (Skill `hns-oss-docs-run`: scope → author ko canonical → translate derived locales → verify). All RED baselines in acceptance.md were measured on this tree on 2026-08-25.

## §B Known Issues

1. **Page-count baseline correction (report to lead)**: the delegation quoted "131→132 per locale"; measured baseline is **150 pages per locale** (`find docs-site/content -name '*.md' | cut -d/ -f3 | sort | uniq -c` → 150 en / 150 ja / 150 ko / 150 zh). Expected post-change count is **152 per locale** (2 new pages × 4 locales). AC-WFD-010 uses the measured figures.
2. **Phase label assumption**: `phase: "v3.1.3 target"` assumes the next unreleased version after the v3.1.2 tag measured on this tree (`git describe --tags --abbrev=0` → v3.1.2; `docs-site/hugo.toml` `version = "v3.1.2"`). If v3.1.3 ships before this SPEC closes, the label is repaired in place (non-transition frontmatter correction, owner manager-spec).
3. **Nav leaf entries need no icon**: `data/menu/main.yaml` leaf entries carry `name:` (4-locale map) + `ref:` only; icons live at section level. The lead's icon constraint is satisfied trivially — no new icon is introduced. If an icon is ever needed, reuse SVG cases `school` or `flash_on` (`layouts/partials/menu.html:58/:65`).
4. **Translation drift of protocol tokens**: numbers and identifiers (0.75/0.80/0.85, 8/16/25, 2/3/5 files, cap 10, `workers.json`, `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, `Class A/B/C`) must survive translation locale-verbatim — the locale-parity ratchet plus the AC token greps enforce this.

## §C Pre-flight (run-phase entry)

- [ ] Plan Audit Gate verdict for this SPEC (task #7) is PASS or skip-eligible, AND Implementation Kickoff Approval obtained (mandatory, score-independent).
- [ ] M0 nav proposal approved by team lead (gates M4 only).
- [ ] `hugo` available locally for the verify build; `.locale-parity-baseline` present (measured: exists, 58 lines).
- [ ] Working tree clean except intended paths; commit staging by explicit pathspec only.

## §D Constraints

Per spec.md §C — i18n HARD rules (ko canonical chains, 4-locale same-PR, TD-only Mermaid, no body emoji, URL whitelist, version SSOT, no vercel.json redirects for additions); canonical-fidelity to the three rule files; nav gating (M4 after lead approval); touch boundary `docs-site/` + READMEs only.

## §E Self-Verification (run-phase exit)

1. **AC matrix**: run all 11 acceptance.md commands; every AC GREEN with output captured verbatim to the run evidence (progress.md §E.2).
2. **Regression gates (already passing today — must stay passing)**: `hns-oss-docs-verify` §1–§6 — warning-free `hugo --minify --gc` + sitemap; URL blacklist 0; Mermaid LR/RL 0; 4-locale file-existence loop clean + per-page section-count ratchet `comm -23` empty; README H2 parity 12/12/12/12 (`grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md`); body-emoji scan clean; version-string sync (no new stale displays introduced — historical citations exempt).
3. **New-page parity**: the 2 new pages must have IDENTICAL H2-and-deeper counts across 4 locales (they enter the ratchet as new pages; any divergence = NEW divergence = FAIL, they are not in the 58-line baseline).
4. **Canon spot-check**: the spec-lifecycle page's phase table names manager-spec / manager-develop / manager-docs; the Tier table carries 0.75/0.80/0.85; the factory page carries cap 10 + workers.json. Cross-read against the canonical rule files.

## §F Milestones

Ordered by decision-reversibility: the naming/split-boundary decisions come first; mechanical nav registration and verification come last. No time estimates — priority labels only.

- **M0 (Priority High, parallel — starts immediately, gates M4 only)**: Report the nav proposal to the team lead (SendMessage): 2 new pages — `advanced/factory-mode.md` (advanced `_meta.yaml`, adjacent to kanban-mode) and `core-concepts/spec-lifecycle.md` (core-concepts `_meta.yaml`, adjacent to spec-based-dev) — `main.yaml` leaf entries with 4-locale name maps, no new icon. Page authoring (M1–M3) proceeds without waiting; M4 blocks on approval.
- **M1 (Priority High) — GAP-3 spec-lifecycle page**: Author `core-concepts/spec-lifecycle.md` in ko first (page NAME decision finalizes here: `spec-lifecycle.md`), covering the spec.md REQ-WFD-005 content list; derive en/ja/zh; add bidirectional cross-links with `spec-based-dev.md` incl. the division-of-labor sentence. Flips AC-WFD-006/007/008, AC-WFD-011 (with M4 for nav).
- **M2 (Priority High) — GAP-2 factory-mode page**: Decide the split boundary (what moves out of kanban-mode.md's factory section vs. what stays in the summary); author `advanced/factory-mode.md` in ko per REQ-WFD-003; derive en/ja/zh; reduce kanban-mode.md's factory section to summary + link ×4 locales. Flips AC-WFD-003/004/005.
- **M3 (Priority Medium) — GAP-1 + GAP-4 card classes**: Add the card-class section (normative heading tokens per spec.md §C.4) to `advanced/kanban-mode.md` ×4 locales, ko first; add the compact card-class table to the README kanban section ×4 files, `README.ko.md` first. Flips AC-WFD-001/002.
- **M4 (Priority High, GATED on M0 approval) — nav registration**: `_meta.yaml` entries ×4 locales ×2 pages + `data/menu/main.yaml` 2 leaf entries (4-locale name maps + `ref`). No icon additions. Flips AC-WFD-009.
- **M5 (Priority High) — verify exit gate**: Run plan §E in full (AC matrix + regression gates + new-page parity + canon spot-check); record evidence; hand to sync phase.

## §G Anti-Patterns

- Authoring derived locales before the ko canonical settles (retranslation churn).
- "Fixing" canonical ko content inside a translation — report the discrepancy instead.
- Adding the new pages to `.locale-parity-baseline` — that admits new debt; the pages must land in parity.
- Restating Origin-Trail internals or session-handoff material in the new pages (out of scope).
- Inventing a menu icon or touching `layouts/partials/menu.html`.
- Editing any file under `internal/` or `internal/template/templates/`.
- Hardcoding version displays into the new pages (version SSOT is hugo.toml).

## §H Cross-References

- spec.md (this SPEC) / acceptance.md (AC evidence table)
- `.moai/reports/t273/gap-map.md` (scope source)
- Skills: `hns-oss-docs-i18n-rules` (HARD digest), `hns-oss-docs-readme-sync` (README 4-file procedure), `hns-oss-docs-structure-map` (nav schemas), `hns-oss-docs-verify` (exit gate)
- Harness runner: `hns-oss-docs-run`
