---
name: hns-oss-docs-readme-sync
description: >
  README 4-file synchronization procedure for the oss-docs harness: Korean
  README.ko.md as primary source, en/ja/zh derivation, the shared
  language-switcher header contract, section-order parity checklist, and the
  manual verification recipe (no linter exists for READMEs). Loaded by the
  content-author and locale-translator specialists for any README work.
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "2.0.0"
  category: "harness"
  status: "active"
  updated: "2026-08-17"
  tags: "oss-docs,readme,4-locale,translation,parity"
---

# README 4-File Sync Procedure

The GitHub-facing README set is 4 files at the repo root:

| File | Locale | Role |
|------|--------|------|
| `README.ko.md` | ko | **canonical / primary** — author here first |
| `README.md` | en | derived |
| `README.ja.md` | ja | derived |
| `README.zh.md` | zh | derived |

> **Chain history**: the canonical locale was English until 2026-08-17, when
> card t47 promoted the ko new-skeleton (feature-oriented section structure)
> to canonical per the operator decision — aligning the README chain with the
> docs-site chain, which was already ko-canonical. The former en-skeleton
> redesign reference (`.moai/reports/readme-docs-redesign-20260713.md`) is
> superseded by the ko skeleton and kept as history only.

## Procedure

1. **Author** the change in `README.ko.md` (Korean) only. Respect the
   canonical ko section skeleton; keep the file length in the range of the
   current set.
2. **Derive** en, ja, zh — same PR, one derived file per translator worker.
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
- [ ] H2 section ORDER matches ko (compare `grep '^## '` output order).
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
| Editing `README.md` first "because GitHub is English-facing" | README canonical is ko — author `README.ko.md`, then derive |
| Re-authoring an entire derived file for a 3-line canonical change | Minimal-diff derivation of the changed sections |
| "Improving" facts/figures during translation | Report the discrepancy; amend canonical first |
| Dropping the switcher header in a redesign | The 4-entry header is a HARD shared contract |
