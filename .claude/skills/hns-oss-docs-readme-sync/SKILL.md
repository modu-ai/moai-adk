---
name: hns-oss-docs-readme-sync
description: >
  README 4-file synchronization procedure for the oss-docs harness: English
  README.md as primary source, ko/ja/zh derivation, the shared
  language-switcher header contract, section-order parity checklist, and the
  manual verification recipe (no linter exists for READMEs). Loaded by the
  content-author and locale-translator specialists for any README work.
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "harness"
  status: "active"
  updated: "2026-07-13"
  tags: "oss-docs,readme,4-locale,translation,parity"
---

# README 4-File Sync Procedure

The GitHub-facing README set is 4 files at the repo root, ~1100 lines each:

| File | Locale | Role |
|------|--------|------|
| `README.md` | en | **canonical / primary** — author here first |
| `README.ko.md` | ko | derived |
| `README.ja.md` | ja | derived |
| `README.zh.md` | zh | derived |

SSOT design reference for redesign work:
`.moai/reports/readme-docs-redesign-20260713.md` (README full redesign draft
+ docs-site 12→11 section restructure + 6 new docs × 4 locales).

## Procedure

1. **Author** the change in `README.md` (English) only. Respect the section
   design from the SSOT report; keep the ~1100-line budget in mind.
2. **Derive** ko, ja, zh — same PR, one derived file per translator worker.
   Translate the changed sections minimally; do not rewrite untouched prose.
3. **Preserve verbatim** across all 4 files: code blocks, command names,
   badges, version strings, file paths, tables' structure, Mermaid direction,
   and the switcher header (below).
4. **Verify parity** (checklist below) before returning.

## Language-switcher header contract [HARD]

All 4 files share the same switcher header near the top, linking the sibling
files with the label set exactly:

```
English · 한국어 · 日本語 · 中文
```

- The current file's own label renders as plain text; the other 3 are links
  to the sibling README files.
- Never reorder, drop, or re-label the 4 entries.

## Section-order parity checklist

- [ ] `grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md` —
      identical H2 counts across the 4 files.
- [ ] H2 section ORDER matches en (compare `grep '^## '` output order).
- [ ] H3 counts per section match for sections you touched.
- [ ] Table row counts match in touched sections.
- [ ] Code-block count matches (` ```` grep -c '^```' ```` ` is even and equal).
- [ ] Switcher header present and correct in all 4.
- [ ] URL blacklist clean: `grep -n 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' README*.md` → no matches.

## Manual verification recipe

No linter exists for the README set — verification is the manual recipe above
plus rendering sanity: preview the markdown (GitHub-flavored) for the touched
sections and confirm Mermaid blocks declare `TD`/`TB` only. The runnable
docs-site checks live in Skill("hns-oss-docs-verify"); the README-specific
checks are the greps above.

## Anti-patterns

| Anti-pattern | Correct approach |
|--------------|------------------|
| Editing README.ko.md first "because the user speaks Korean" | README canonical is en — author `README.md`, then derive |
| Re-authoring an entire derived file for a 3-line canonical change | Minimal-diff derivation of the changed sections |
| "Improving" facts/figures during translation | Report the discrepancy; amend canonical first |
| Dropping the switcher header in a redesign | The 4-entry header is a HARD shared contract |
