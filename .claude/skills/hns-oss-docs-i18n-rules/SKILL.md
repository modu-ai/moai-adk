---
name: hns-oss-docs-i18n-rules
description: >
  HARD i18n rules digest for the oss-docs harness specialists working on
  moai-adk-go README 4-locale set and the docs-site (adk.mo.ai.kr). Covers the
  canonical-locale chains, the 4-locale same-PR obligation, Mermaid TD-only,
  the no-emoji + icon-shortcode rule, emphasis-marker spacing, the URL
  blacklist, version SSOT, vercel.json redirect pattern, and the immutable
  Vercel binding. Loaded FIRST by every oss-docs specialist before any edit.
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "harness"
  status: "active"
  updated: "2026-07-13"
  tags: "oss-docs,i18n,4-locale,readme,docs-site,adk.mo.ai.kr"
---

# oss-docs HARD i18n Rules

SSOT: `.moai/docs/docs-site-i18n-rules.md` (+ CLAUDE.local.md §17.1 for the
design/icon regime). This skill is the working digest; on conflict, the SSOT
wins — EXCEPT the two stale items in § Known-Stale below, where reality wins.

## 1. Canonical-locale chains [HARD]

| Surface | Canonical | Derivation chain | Derived files |
|---------|-----------|------------------|---------------|
| docs-site | **ko** | ko → en → ja/zh, same PR | `docs-site/content/{en,ja,zh}/` |
| README | **en** (`README.md`) | en → ko/ja/zh, same PR | `README.ko.md`, `README.ja.md`, `README.zh.md` |

- Author in the canonical locale only; derive the rest. Never "fix" canonical
  content inside a translation — report the discrepancy back instead.

## 2. 4-locale simultaneous-update obligation [HARD]

Every canonical content change MUST land in all 4 locales in the same PR.
A canonical edit without its 3 derived counterparts is a locale-parity FAIL
(sprint contract `locale-parity` threshold 1.0, must_pass).

## 3. Mermaid TD-only [HARD]

- Allowed: `flowchart TD`, `graph TB`.
- Forbidden: `LR` / `RL` directions (`flowchart LR`, `graph LR`,
  `flowchart RL`, `graph RL`).
- Translation preserves diagram direction verbatim.

## 4. No emoji in body text [HARD]

- Use the icon shortcode instead: `{{</* icon <name> [variant] */>}}`
  (defined in `docs-site/layouts/shortcodes/icon.html`; variants:
  `ok|warn|danger|primary|muted`).
- Preserved (NOT emoji — do not strip): typographic symbols `→ ← ↓ ✓ ✗`, and
  branding emoji inside MoAI orchestrator-banner example code blocks.

## 5. Emphasis-marker spacing [HARD]

- Correct: `**바이브코딩** (Vibe Coding)` — parenthetical OUTSIDE the markers.
- Wrong: `**바이브코딩(Vibe Coding)**`.

## 6. URL blacklist [HARD]

Only `adk.mo.ai.kr` is a valid docs-site domain. Forbidden (all occurrences,
including link labels and translated prose):

- `docs.moai-ai.dev`
- `adk.moai.com`
- `adk.moai.kr`

## 7. Version SSOT [HARD]

`docs-site/hugo.toml` `params.version` / `params.releaseDate` is the single
version surface. Never hardcode divergent version/date strings into pages,
menus, or READMEs beyond what the release process syncs.

## 8. Moved pages need redirects [HARD]

Every docs-site page move/rename adds to `docs-site/vercel.json`:

1. Locale-aware: `/:locale(ko|en|ja|zh)/old-path → /:locale/new-path`
2. Non-locale fallback: `/old-path → /ko/new-path`

## 9. Vercel binding immutable [HARD]

The Vercel project binding and deployment config are never changed by this
harness (redirects array excepted). Push = production deploy at adk.mo.ai.kr,
which is why specialists NEVER commit/push — publishing is human-gated.

## Known-Stale items in the SSOT doc

The SSOT `.moai/docs/docs-site-i18n-rules.md` predates the current site and
carries 2 stale facts — reality wins:

1. It says the theme is **Hextra** → reality: **hugo-geekdoc**.
2. It says config is **hugo.yaml** → reality: **`docs-site/hugo.toml`**
   (defaultContentLanguage=ko).

Also: the scripts `docs-i18n-check.sh` and `gen_menu.py` referenced there DO
NOT exist. Never shell out to them — run the inlined checks in
Skill("hns-oss-docs-verify") instead.
