---
id: SPEC-V3R6-DOCS-I18N-PARITY-001
title: "docs-site 4-locale i18n parity baseline clearance"
version: "0.2.0"
status: implemented
created: 2026-05-31
updated: 2026-05-31
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs-site, i18n, parity, ci, 4-locale"
---

# SPEC-V3R6-DOCS-I18N-PARITY-001 — docs-site 4-locale i18n parity baseline clearance

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-31 | manager-spec | Initial plan-phase authoring. Reproduced checker: 53 ERRORs (4 frontmatter + 26 H1 + 23 glossary). |
| 0.1.1 | 2026-05-31 | manager-spec | D1 (plan-audit SHOULD-FIX) resolution: added AC-DIP-009 mapping orphaned REQ-DIP-008 (canonical-URL / Mermaid TD-only / forbidden-URL regression guard). AC matrix updated; no normative REQ text changed. |

## A. Context (WHY)

The `scripts/docs-i18n-check.sh` validator reports **53 ERRORs** from docs-site
4-locale (ko / en / ja / zh) structural divergence. The error count was confirmed
by running the checker directly on the current working tree (HEAD `038f6e793`):

```
=== Summary ===
Errors:   53
Warnings: 0
```

The 53 errors split into exactly three checker categories:

| Category | Checker section | Error count |
|----------|-----------------|-------------|
| C1 — frontmatter title presence | Check 2 | 4 |
| C2 — H1 heading existence | Check 3 | 26 |
| C3 — glossary term preservation | Check 4 | 23 |

This SPEC clears the entire 53-error baseline so the docs-i18n parity check
reports **0 errors**, enabling a future Phase-2 flip of the workflow to strict
mode (`DOCS_I18N_STRICT=1`) without immediately blocking routine work.

### A.1 Honest framing — the workflow is NOT currently blocking

The CI workflow `.github/workflows/docs-i18n-check.yml` is **advisory-only**:

- The job declares `continue-on-error: true`.
- The strictness resolver sets `strict=false` for both `pull_request` and
  `push` to `main` events (Phase-1 rollout), so the checker runs with
  `DOCS_I18N_STRICT=0` and exits 0 even with 53 errors present.

Therefore the workflow does NOT mechanically fail on the 53 errors today; it
surfaces them as a non-blocking drift comment. The genuine value of this SPEC is
clearing the baseline so the maintainer can flip Phase-2 strict mode (a separate,
out-of-scope change) and have the gate actually block regressions going forward.

This SPEC's success is therefore defined against the **checker output** (0 errors),
not against a red→green CI transition (the CI is already non-red).

### A.2 Diagnosis corrections discovered during reproduction

Running the checker revealed two deviations from the initial diagnosis estimate
that this SPEC's acceptance criteria are authored against (actual checker output,
not the estimate):

1. **C3 SPEC-First location**: The missing `SPEC-First` term is in
   `zh/getting-started/introduction.md` and `zh/getting-started/update.md`
   (NOT `contributing/_index.md`). The `contributing/_index.md` errors are for
   `MoAI-ADK` and `moai-adk` only.

2. **C2 / C3 overlap on multi-llm stubs**: `multi-llm/cg-mode.md` and
   `multi-llm/model-policy.md` are `draft: true` placeholder stubs in en/ja/zh
   ("This page is only available in Korean") that contribute errors to BOTH
   C2 (missing H1) and C3 (missing glossary terms). Resolving these stubs once
   per locale clears errors in both categories simultaneously.

## B. Scope (WHAT)

### B.1 Category breakdown (verified file sets)

**C1 — frontmatter missing (4 errors).** The ko source
`cost-optimization/prompt-caching.md` (added by SPEC-V3R6-PROMPT-CACHE-001) was
written as bare markdown — it begins directly with an `#` H1 and has NO YAML
frontmatter. All 4 locale copies lack a frontmatter block:

- `docs-site/content/{ko,en,ja,zh}/cost-optimization/prompt-caching.md`

**C2 — H1 heading missing (26 errors).** Two sub-sets:

- *Set A — 5 files missing H1 in ALL 4 locales (20 errors).* These have valid
  frontmatter and body content, but the body's first heading is `##` (H2, e.g.
  `## 개요`), and the checker's Check 3 requires an `#` (H1) after the
  frontmatter:
  - `advanced/harness-profiles.md`
  - `core-concepts/harness-engineering.md`
  - `db/migration-patterns.md`
  - `getting-started/profile.md`
  - `workflow-commands/moai-design.md`
  (each × 4 locales = 20)
- *Set B — 2 stub files missing H1 only in en/ja/zh (6 errors).* These are
  `draft: true` placeholder stubs; ko has full content with an H1, en/ja/zh do not:
  - `multi-llm/cg-mode.md` (en/ja/zh)
  - `multi-llm/model-policy.md` (en/ja/zh)
  (2 files × 3 locales = 6)

**C3 — glossary term parity (23 errors).** Terms present verbatim in the ko
source must appear verbatim (case-sensitive, via `grep -Fq`) in en/ja/zh:

- `Claude Code` (9): en/ja/zh × {`multi-llm/_index.md`, `multi-llm/cg-mode.md`, `multi-llm/model-policy.md`}
- `MoAI-ADK` (9): en/ja/zh × {`contributing/_index.md`, `multi-llm/_index.md`, `multi-llm/model-policy.md`}
- `moai-adk` (3): en/ja/zh × {`contributing/_index.md`}
- `SPEC-First` (2): `zh/getting-started/introduction.md`, `zh/getting-started/update.md`

The glossary list is the maintainer-canonical set defined in
`scripts/docs-i18n-check.sh` (`GLOSSARY_TERMS`): MoAI-ADK, SPEC-First, EARS,
TRUST 5, Claude Code, Anthropic, moai-adk. These are product / proper nouns that
the parity rule requires verbatim in every locale.

### B.2 Requirements (GEARS)

REQ-DIP-001 (Ubiquitous): The docs-i18n checker shall report **0 frontmatter-missing
errors** (Check 2) for all four locale copies of `cost-optimization/prompt-caching.md`.

REQ-DIP-002 (Ubiquitous): The docs-i18n checker shall report **0 H1-missing
errors** (Check 3) for every `.md` file in every locale.

REQ-DIP-003 (Ubiquitous): The docs-i18n checker shall report **0 glossary-term
parity errors** (Check 4) for every canonical glossary term present in a ko source file.

REQ-DIP-004 (Event-driven): **When** a maintainer runs
`DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh`, the checker shall print
`Errors:   0` in its Summary section.

REQ-DIP-005 (Event-driven): **When** a maintainer runs the checker in strict mode
(`bash scripts/docs-i18n-check.sh`, i.e. `DOCS_I18N_STRICT=1`), the checker shall
exit with status code 0.

REQ-DIP-006 (State-driven): **While** an en/ja/zh page exists as a `draft: true`
placeholder stub, the page shall still contain an `#` H1 heading and every
canonical glossary term that its ko counterpart contains, so that the parity
checker passes without requiring the stub to be fully translated.

REQ-DIP-007 (Ubiquitous): The four-locale file-path parity (Check 1) shall remain
intact — no locale shall gain or lose a `.md` file relative to the ko canonical set.

REQ-DIP-008 (State-driven): **While** editing any docs-site content file, the editor
shall preserve the canonical-URL and Mermaid TD-only rules of the docs-site i18n
authoring contract, and shall not introduce any forbidden URL.

## C. Exclusions (What NOT to Build)

- **Go test CI flakies** — handled by the sibling SPEC
  (SPEC-V3R6-CI-FLAKY-STABILIZE-001). This SPEC touches docs-site content only;
  no Go code, no `internal/`, no `_test.go`.
- **Flipping the workflow to strict mode** — changing
  `.github/workflows/docs-i18n-check.yml` to set `DOCS_I18N_STRICT=1` (Phase-2)
  is a separate maintainer decision, out of scope here. This SPEC only clears the
  baseline so that flip becomes safe.
- **Full translation of placeholder stubs** — en/ja/zh `draft: true` stubs are
  brought to parity (H1 + invariant glossary terms inserted) but are NOT
  fully translated from ko. Full translation is a larger, separate docs effort.
- **Adding NEW documentation pages or content** beyond the parity fixes — no new
  `.md` files, no new sections unrelated to the 53 errors.
- **`internal/template/templates/`** — no template-source changes; this is
  live docs-site content (`docs-site/content/`), not deployed user-project templates.
- **Modifying `scripts/docs-i18n-check.sh` or the `GLOSSARY_TERMS` list** — the
  checker rules are treated as the fixed contract to satisfy, not to relax.
- **The 2 local-ahead commits** (`038f6e793`, `3dcd58677`, parallel session) —
  neither touches `docs-site/content`; do not rebase, revert, or touch them.

## D. Acceptance Criteria Reference

See `acceptance.md` for the binary, command-verifiable AC-DIP-NNN criteria. Each
AC asserts a concrete checker error count for its category, plus a final
all-categories-green gate.
