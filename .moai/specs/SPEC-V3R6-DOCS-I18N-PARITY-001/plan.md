# Implementation Plan — SPEC-V3R6-DOCS-I18N-PARITY-001

## A. Context

Clear the 53-error docs-i18n parity baseline so
`DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` reports `Errors:   0`.
This is **docs-site content work only** — pure markdown edits under
`docs-site/content/{ko,en,ja,zh}/`. No Go code, no templates, no test files.

### Run-phase routing

[HARD] This is **manager-docs-implementable** run-phase work, NOT manager-develop.
The orchestrator MUST route the run phase to the `manager-docs` subagent
(`/moai run SPEC-V3R6-DOCS-I18N-PARITY-001` → manager-docs). There is no Go
implementation, no TDD/DDD cycle, no `cycle_type`. The "tests" are the bash
checker invocations enumerated in acceptance.md.

Per `.moai/docs/docs-site-i18n-rules.md` §17.5, the default executor for
docs-site changes is `manager-docs`. Editing many files across 4 locales MAY use
`isolation: worktree` if the runtime materializes it; not mandated (per CLAUDE.md
§14 worktree advisory + 2026-05-17 user policy).

## B. Known Issues / Constraints

- **Diagnosis correction (verified)**: C3 `SPEC-First` errors are in
  `zh/getting-started/introduction.md` + `zh/getting-started/update.md`, not in
  `contributing/_index.md`. Author fixes against the live checker output.
- **C2/C3 overlap**: `multi-llm/cg-mode.md` + `multi-llm/model-policy.md`
  (en/ja/zh) are `draft: true` stubs that error in BOTH C2 and C3. A single edit
  per locale file (add H1 + insert invariant glossary sentence) clears both.
- **CI is non-blocking**: the workflow is `continue-on-error: true` + warn-only.
  Success is `Errors: 0` from the checker, not a CI red→green flip.
- **Glossary check is case-sensitive `grep -Fq`**: terms must appear EXACTLY as
  `MoAI-ADK`, `moai-adk`, `Claude Code`, `SPEC-First` — these are distinct strings.
- **Check 3 H1 = `^#[[:space:]]+.+`**: a `## H2` does NOT satisfy it. The first
  body heading after frontmatter must be a single `#`.
- **`_index.md` H1 exemption**: Check 3 skips `_index.md` files, so the
  `multi-llm/_index.md` glossary fix does NOT also need an H1.
- **File-path parity (Check 1) must stay intact**: do not add/remove `.md` files;
  edit existing files in place only.

## C. Pre-flight

- `ls .moai/specs/` → confirmed `SPEC-V3R6-DOCS-I18N-PARITY-001/` is the only
  matching directory (no duplicate ID).
- HEAD `038f6e793`; 2 local-ahead commits, neither touches `docs-site/content`.
- Baseline checker run captured (53 errors). Re-run after each milestone.

## D. Constraints (authoring contract — `.moai/docs/docs-site-i18n-rules.md`)

- Canonical source = `ko`. en/ja/zh mirror ko.
- 4 fixed locales: ko / en / ja / zh.
- Frontmatter fields: `title`, `weight`, `draft` (match sibling files in the same
  directory). Use the exact schema observed in siblings, e.g.:
  ```yaml
  ---
  title: <localized title>
  weight: <integer matching sibling weights>
  draft: false
  ---
  ```
- Emphasis-paren spacing rule, Mermaid TD-only, no forbidden URLs, no emoji in body.
- Glossary proper nouns are invariant (untranslated) across all locales.

## E. Self-Verification (run after each milestone)

```bash
# Per-category counts
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep -c "no frontmatter block"   # C1 → target 0
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep -c "no H1 heading"          # C2 → target 0
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep -c "glossary term"          # C3 → target 0
# Final gate (both modes)
DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1 | grep "Errors:"                    # → Errors:   0
bash scripts/docs-i18n-check.sh; echo "exit=$?"                                            # strict → exit=0
```

## F. Milestones (priority-ordered, no time estimates)

### M1 — C1: frontmatter for prompt-caching (4 files)
- Prepend a YAML frontmatter block to
  `docs-site/content/{ko,en,ja,zh}/cost-optimization/prompt-caching.md`.
- READ a sibling in the same directory (`cost-optimization/_index.md` per locale,
  or another `cost-optimization/*.md`) to copy the exact `title`/`weight`/`draft`
  schema. The file already has an `#` H1, so only frontmatter is added.
- Localize `title` per locale (ko already has the H1 text "프롬프트 캐싱…"; en/ja/zh
  use their localized titles).
- Verify: C1 count → 0.

### M2 — C2: H1 headings
- *Set A (5 files × 4 locales)*: insert a `# <title>` H1 immediately after the
  frontmatter (before the existing intro paragraph / first `##`), mirroring the
  frontmatter `title` value in each locale:
  `advanced/harness-profiles.md`, `core-concepts/harness-engineering.md`,
  `db/migration-patterns.md`, `getting-started/profile.md`,
  `workflow-commands/moai-design.md`.
- *Set B (2 stub files × en/ja/zh)*: add an `#` H1 to the `draft: true` stubs
  `multi-llm/cg-mode.md`, `multi-llm/model-policy.md` (en/ja/zh). (Combined with M3
  for these files — single edit adds H1 + glossary terms.)
- Verify: C2 count → 0.

### M3 — C3: glossary parity
- *multi-llm stubs (en/ja/zh × cg-mode, model-policy, `_index.md`)*: insert the
  invariant terms verbatim. The minimal honest approach: add a short locale-prefixed
  sentence that names the proper nouns, e.g. an en line
  `MoAI-ADK supports the z.ai GLM backend alongside Claude Code.` Each file must
  contain exactly the terms the checker flags for it (`MoAI-ADK`, `Claude Code`,
  `moai-adk` as applicable). Note `_index.md` needs terms only (H1-exempt).
- *contributing/_index.md (en/ja/zh)*: ensure `MoAI-ADK` and `moai-adk` appear verbatim.
- *zh getting-started*: ensure `SPEC-First` appears verbatim in
  `introduction.md` and `update.md` (ko has `### SPEC-First 워크플로우`; insert the
  equivalent zh heading/sentence containing the literal `SPEC-First`).
- Verify: C3 count → 0.

### M4 — Green gate (both modes)
- Full run: `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` → `Errors:   0`.
- Strict run: `bash scripts/docs-i18n-check.sh` → exit 0 + final OK line.
- Confirm Check 1 file-path parity unchanged (78 files per locale, no add/remove).
- Hybrid Trunk Tier M → main-direct commit (no PR unless `--pr`).

## G. Anti-Patterns to Avoid

- Translating a glossary term (e.g. localizing `MoAI-ADK`) — they are invariant;
  the checker does case-sensitive `grep -Fq` for the verbatim string.
- Using `##` where the checker needs `#` (H1) — Check 3 only matches `^#[ ]+`.
- Adding/removing `.md` files (breaks Check 1 parity).
- Touching the 2 local-ahead parallel-session commits.
- Editing `internal/template/templates/` (this is live docs, not templates).
- Relaxing `scripts/docs-i18n-check.sh` or removing a glossary term to "pass".

## H. Cross-References

- `scripts/docs-i18n-check.sh` — the contract being satisfied.
- `.github/workflows/docs-i18n-check.yml` — advisory workflow (Phase-1 warn-only).
- `.moai/docs/docs-site-i18n-rules.md` — 4-locale authoring rules (canonical ko,
  invariant glossary, Mermaid TD-only, forbidden URLs).
- CLAUDE.local.md §17 — docs-site i18n doctrine.
- Sibling SPEC-V3R6-CI-FLAKY-STABILIZE-001 — Go-test flakies (disjoint scope).
