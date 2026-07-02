---
id: SPEC-V3R6-DOCS-I18N-COMPLETION-001
title: "Docs-Site i18n Completion — Non-KO Translation Backlog + Feedback Page Parity + Glossary Fix"
version: "0.1.1"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, translation, glossary, follow-up"
era: V3R6
tier: L
depends_on: [SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001]
related_specs: [SPEC-V3R6-DOCS-V3-REBUILD-001, SPEC-INVOCATION-MODEL-001]
---

# SPEC-V3R6-DOCS-I18N-COMPLETION-001 — Docs-Site i18n Completion

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-07-02 | 0.1.0 | manager-spec | Initial plan-phase authoring. Bundles 3 deferred follow-up items from SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001 (`status: completed`, sync commit `85391a770`, `progress.md` §E.4 backfill `fde79556a`): (D5) the non-ko untranslated-content backlog tracked in memory file `project_docsite_untranslated_backlog.md`, (2) the deferred en/ja/zh translation of REQ-DPC-006's new ko `moai-feedback.md` subsection, and (3) the pre-existing, out-of-scope Check-4 glossary defect (`ja/claude-code/agentic/best-practices.md` missing the term "Anthropic"). All three items were re-verified fresh in this session (not carried over from the memory file's 2026-07-01 snapshot) — see Ground Truth below. |
| 2026-07-02 | 0.1.1 | manager-spec | Iteration-2 revision responding to an independent plan-auditor FAIL verdict (score 0.77, Tier L threshold 0.85 — all 4 MUST-PASS criteria passed; the FAIL was score-driven from AC-matrix rigor gaps, NOT from incorrect Ground Truth, which the auditor independently re-ran and confirmed accurate). Fixes 9 defects: **D1/D2** corrected the double-counted file-count arithmetic (26 flagged + 3 + 1 = 30 → 23 genuine Item-1 files + 3 Item-2 + 1 Item-3 = 27 touched files; the 3 untouched `init-wizard.md` false-positives are excluded from the workload count) in plan.md §A and acceptance.md's Quality Gate Criteria. **D3/D4** relabeled REQ-DIC-002 and REQ-DIC-004 from the incorrect "(Ubiquitous, ...)" GEARS pattern to the correct "(Unwanted — 'shall not' ...)" pattern (sentences were already correct; only the parenthetical label was wrong). **D5** removed a false analogy in plan.md's artifact-set rationale (the predecessor SPEC was Tier M, whose nominal set never included design.md/research.md — nothing was "adapted away" there); the 4-file set is now justified on independent merits only. **D6** added AC-DIC-006a/b for NFR-DIC-004 (Mermaid diagram-direction preservation), scoped after a fresh grep to the 2 actual files containing Mermaid diagrams in Item 1 + Item 2 (`core-concepts/harness-engineering.md`, `utility-commands/moai-feedback.md`). **D7** added AC-DIC-001f requiring a manual per-file cross-check of every translated file's factual claims against its ko source, recorded incrementally in `progress.md` §E.2 per file (not deferred to an end-of-SPEC spot check), plus a matching per-file self-verification obligation added to plan.md M3/M4/M5. **D8** expanded AC-DIC-001d from a 1-file/locale sample to full coverage of all 23 translated (locale, file) pairs (the structural check is mechanically cheap). **D9** reworded AC-DIC-002b to state the operationally-testable claim directly, cross-referenced to M6.1's corrected exact-count assertion (3). |

## Overview (WHY)

SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001 explicitly deferred 3 items as out of scope, each tracked separately: D5 (the non-ko untranslated-content backlog, tracked in auto-memory since 2026-07-01), the en/ja/zh translation of REQ-DPC-006's new ko-only `moai-feedback.md` content (NFR-DPC-004, ko-first sequencing), and the residual Check-4 glossary finding (explicitly named "out of scope... tracked for a future follow-up" in that SPEC's Exclusions). This SPEC is that follow-up, bundling all 3 into one closing SPEC so the docs-site i18n backlog reaches a clean state: `scripts/docs-i18n-check.sh` (default strict mode) should exit 0 after this SPEC closes, and the non-ko Korean-content backlog should be fully translated.

## Scope (WHAT)

This SPEC touches ONLY `docs-site/content/{en,ja,zh}/` — no ko-locale file, no Go source, no `internal/template/templates/`, no `docs-site/hugo.toml` or `docs-site/static/`. Three items:

1. **Item 1 (D5 backlog)**: Translate the non-ko untranslated-content backlog — re-verified this session at **26 flagged files**, of which **23 require genuine prose translation** and **3 are false positives** (see Ground Truth finding below) — from the canonical `ko/` source into natural en/ja/zh prose.
2. **Item 2 (REQ-DPC-006 en/ja/zh translation)**: Translate ko's new "피드백 설정" subsection (added by SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001 REQ-DPC-006) into en/ja/zh `utility-commands/moai-feedback.md`, and independently-verified-per-locale-fix the "sync-auditor" agent-chain misattribution where it still exists (ja/zh; en already correct).
3. **Item 3 (Check-4 glossary defect)**: Add the missing glossary term "Anthropic" to `ja/claude-code/agentic/best-practices.md`'s References section, matching the ko/en/zh sibling pages' sentence pattern.

## Ground Truth (observed evidence — verified in this session, 2026-07-02, not re-derived from the memory file's prior snapshot)

| Fact | Observed value | Verification command (run this session) |
|------|-----------------|------------------------------------------|
| Item 1 — flagged file count (fresh re-run) | Exactly **26 files**: `en` 8, `ja` 9, `zh` 9 | `grep -rlP '[가-힣]' docs-site/content/en/ docs-site/content/ja/ docs-site/content/zh/` |
| Item 1 — flagged file list | `en`: `advanced/{catalog-system,harness-profiles}.md`, `core-concepts/{constitution,harness-engineering}.md`, `getting-started/{init-wizard,profile,windows-guide}.md`, `guides/ci-autonomy.md` (8). `ja`/`zh`: same 8 + `advanced/hooks-reference.md` (9 each) | Same command, full path list captured |
| **Item 1 — false-positive finding (non-obvious, independently discovered this session)** | `getting-started/init-wizard.md` (all 3 locales) is flagged by the raw Hangul grep, but its ONLY Hangul content is the literal language-picker label `Korean (한국어)` appearing 4× per file — a native-script rendering of a language NAME inside a CLI mockup (`? Select conversation language: ... Korean (한국어) - English/한국어로 커밋/etc.`), directly analogous to `Japanese (日本語)` / `Chinese (中文)` in the same block. This is **NOT untranslated prose** — ja/zh's surrounding prose in the same file (`韓国語でコミット` / `用韩语提交` etc.) is already correctly translated; only the intentional native-script label remains, identically in en/ja/zh. Translating or removing `한국어` here would be a REGRESSION (it would break the multi-script language-picker illustration), not a fix. | `grep -noP '.*[가-힣].*' docs-site/content/{en,ja,zh}/getting-started/init-wizard.md` — 4 matching lines per locale, each showing the `Korean (한국어)` label pattern with correctly-localized surrounding text |
| Item 1 — genuine translation scope (after excluding the false positive) | **23 files**: en 7, ja 8, zh 8 (each locale's flagged set minus `init-wizard.md`) | Derived from the above two facts |
| Item 1 — per-file Hangul character volume (genuine files only) | `catalog-system.md` 468/locale · `harness-profiles.md` 761/locale · `hooks-reference.md` 698/locale (ja/zh only) · `constitution.md` 620/locale · `harness-engineering.md` 1116/locale · `profile.md` 439/locale (H1 itself is Korean) · `windows-guide.md` 664/locale · `ci-autonomy.md` 495/locale. Total ≈ 15,121 Hangul characters across the 23 genuine files — a substantial prose-translation volume, not a mechanical fix. | `grep -oP '[가-힣]' <file> \| wc -l` per file, all 26 files individually counted this session |
| Item 1 — total line count (genuine 23 files) | 3,725 raw lines summed across the 26 flagged files (includes the 3 false-positive instances, which contribute negligible lines) | `wc -l` on all 26 files |
| Check 1 (4-locale file-count parity) | `ko` 99 / `en` 99 / `ja` 99 / `zh` 99 — unchanged from the SPEC-V3R6-DOCS-V3-REBUILD-001 M3 baseline; this SPEC must not change this count | `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` Check 1 output |
| Check 4 (glossary) — sole residual finding | `glossary term 'Anthropic' missing in ja/claude-code/agentic/best-practices.md (present in ko)` — **exactly 1** error, 0 warnings; Check 3 (H1) is 0 (confirming SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001's REQ-DPC-001 fix is intact) | `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh 2>&1` full output this session — `Errors: 1`, `Warnings: 0` |
| Item 3 — ko/en/zh sentence pattern | ko L227: `이 가이드는 Anthropic의 공식 [...] 문서를 바탕으로 작성되었습니다.` / en L227: `This guide is based on Anthropic's official [...] documentation.` / zh L227: `此指南基于 Anthropic 的官方文档 [...] 编写。` — all under an H2 heading (`## 참고` / `## References` / `## 参考`) | `grep -n "Anthropic" docs-site/content/{ko,en,zh}/claude-code/agentic/best-practices.md` |
| Item 3 — ja divergent structure (root cause of the missing term) | `ja` has 222 lines vs 227 in ko/en/zh (5-line deficit). Its equivalent section is `## 参考資料` (not `## 参考`) containing ONLY a bare bullet-list link — `- [Best practices for Claude Code(公式ドキュメント)](https://code.claude.com/docs/en/best-practices)` — with **no prose sentence** naming "Anthropic" at all (unlike ko/en/zh's one-sentence attribution). This is why Check-4 flags it: the term is not merely mistranslated, it is structurally absent because the section was rendered as a link-only bullet instead of the attribution-sentence pattern used by the other 3 locales. | `grep -n "^## " docs-site/content/{ko,ja}/claude-code/agentic/best-practices.md`; `tail -8 docs-site/content/ja/claude-code/agentic/best-practices.md` |
| Item 3 — glossary check mechanism | `scripts/docs-i18n-check.sh` Check 4 requires the literal case-sensitive string `"Anthropic"` (one of 7 `GLOSSARY_TERMS`: `MoAI-ADK`, `SPEC-First`, `EARS`, `TRUST 5`, `Claude Code`, `Anthropic`, `moai-adk`) to appear verbatim in every locale where `ko` contains it — no requirement on surrounding sentence structure | `sed -n '28,50p' scripts/docs-i18n-check.sh` (GLOSSARY_TERMS array + Check 4 loop) |
| Item 2 — ko new subsection (REQ-DPC-006 content, source-of-truth for translation) | `## 피드백 설정` (L79) with 4 H3 subsections: `진단 정보: 보장 항목 + best-effort 항목` (L83), `중복 이슈 후보 확인` (L87), `` `gh` 인증 실패 시 로컬 임시 저장 `` (L91), `피드백 대상 저장소 설정` (L101) — full text read this session, ~25 lines | `sed -n '55,104p' docs-site/content/ko/utility-commands/moai-feedback.md` |
| Item 2 — en/ja/zh currently lack this subsection | `grep -n "sync-auditor" docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md` and heading-list diff confirm en/ja/zh have NO `## 피드백 설정`-equivalent H2 section at all (their H2 sequence jumps directly from "How It Works" to "Feedback Types" — no "Feedback Settings" section exists) | `grep -n "^#\{2,3\} "` on all 4 locale files, side-by-side heading comparison |
| Item 2 — sync-auditor misattribution: **independently confirmed per-locale, NOT assumed symmetric** | `ja` (L141, L150) and `zh` (L143, L152) STILL carry the "sync-auditor" misattribution in their Mermaid diagram + role table. `en` does **NOT** — en L141/L150 already correctly say `manager-docs Agent` / `manager-docs`, matching the already-fixed ko content. This is an asymmetric finding: only 2 of 3 non-ko locales need the misattribution fix. | `grep -n "sync-auditor" docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md` — 0 matches for ko and en, 2 matches each for ja and zh |
| Item 2 — feedback.repository content absent in en/ja/zh (pre-check) | `grep -c 'feedback.repository\|modu-ai/moai-adk' docs-site/content/{en,ja,zh}/utility-commands/moai-feedback.md` → 0 for all 3 (confirms the new content genuinely does not exist yet, matching AC-DPC-006f from the predecessor SPEC) | Same grep, this session |
| Item 2 — grounding sources (re-verified present) | `.claude/skills/moai/workflows/feedback.md`, `internal/config/feedback_accessors.go`, `internal/template/templates/.moai/config/sections/feedback.yaml` — all 3 exist and are unchanged since the predecessor SPEC's grounding read | `ls -la` on all 3 paths this session |
| Predecessor SPEC status | `SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001` frontmatter `status: completed`; sync commit `85391a770`; `progress.md` §E.4 backfill commit `fde79556a` (HEAD at plan-phase start) | `git log --oneline -5`; `cat .moai/specs/SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001/spec.md \| head -15` |
| Unrelated parallel SPEC (do not touch) | `.moai/specs/SPEC-DEAD-CONFIG-001/` exists as an untracked directory in `git status` — confirmed unrelated (config dead-code cleanup, no docs-site overlap) | `git status --short` |

## Requirements (GEARS)

### Item 1 — Non-ko untranslated-content backlog (D5)

- **REQ-DIC-001** (Event-driven): **When** a reader opens the en/ja/zh version of any of the 23 genuinely-untranslated pages (en: `advanced/catalog-system.md`, `advanced/harness-profiles.md`, `core-concepts/constitution.md`, `core-concepts/harness-engineering.md`, `getting-started/profile.md`, `getting-started/windows-guide.md`, `guides/ci-autonomy.md`; ja/zh: the same 7 + `advanced/hooks-reference.md`), the page **shall** present natural target-language prose translated from the corresponding `ko/` canonical source, preserving the document's heading structure, code fences, Mermaid diagrams (direction unchanged), inline links, and frontmatter `title`/`description` fields' informational content (translated, not copied verbatim from ko).
- **REQ-DIC-002** (Unwanted — "shall not", exclusion — false-positive carve-out): The en/ja/zh `getting-started/init-wizard.md` page's native-script language-picker labels (`Korean (한국어)`, alongside `Japanese (日本語)`, `Chinese (中文)`) **shall not** be translated, removed, or otherwise altered — these are intentional literal script renderings of language names inside a CLI selection mockup, not untranslated prose, and altering them would break the multi-script illustration.

### Item 2 — REQ-DPC-006 en/ja/zh translation of moai-feedback.md

- **REQ-DIC-003** (Ubiquitous): The en/ja/zh `utility-commands/moai-feedback.md` pages **shall** include a translated equivalent of the ko "피드백 설정" subsection (added by SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001 REQ-DPC-006), describing all 4 enhancements — guaranteed diagnostics (MoAI version + OS) plus best-effort diagnostics (Go toolchain version + orchestrator-passed error context); the duplicate-issue candidate-report step; the `gh`-failure graceful fallback with local draft save; and the configurable `feedback.repository` target (default `modu-ai/moai-adk`) — as natural target-language prose (not a literal word-for-word rendering of the Korean), grounded strictly in `.claude/skills/moai/workflows/feedback.md`, `internal/config/feedback_accessors.go`, and `internal/template/templates/.moai/config/sections/feedback.yaml`.
- **REQ-DIC-004** (Unwanted — "shall not" — independently verified per-locale, ja/zh only): The ja and zh `utility-commands/moai-feedback.md` agent-chain diagram and role table **shall not** attribute GitHub issue creation to a "sync-auditor" agent; they **shall** instead reflect the orchestrator-direct, no-subagent execution model, mirroring the already-corrected ko and en content. (The en page already carries this fix and requires no change under this REQ.)

### Item 3 — Check-4 glossary defect (ja best-practices.md)

- **REQ-DIC-005** (Ubiquitous): The ja `claude-code/agentic/best-practices.md` References section **shall** contain the term "Anthropic" verbatim, expressed as a natural Japanese attribution sentence consistent with the ko/en/zh sibling pages' pattern ("based on Anthropic's official ... documentation"), added to (or replacing) the existing link-only bullet-list section.

### Non-functional constraints

- **NFR-DIC-001** (4-locale file-count parity): This SPEC's edits **shall not** change the 99-file-per-locale baseline (verified by SPEC-V3R6-DOCS-V3-REBUILD-001 M3 and re-confirmed in this session's Ground Truth) — content is translated into existing files only; no file is added or removed.
- **NFR-DIC-002** (glossary term preservation): Every one of the 7 canonical glossary terms (`MoAI-ADK`, `SPEC-First`, `EARS`, `TRUST 5`, `Claude Code`, `Anthropic`, `moai-adk`) that appears in a `ko/` source file being translated **shall** appear verbatim (untranslated, case-sensitive) in the corresponding translated en/ja/zh output.
- **NFR-DIC-003** (no unobserved-claim / verification-claim-integrity): Every translated sentence **shall** be traceable to content actually present in the `ko/` canonical source (Item 1) or the 3 grounding source files named in REQ-DIC-003 (Item 2) — no invented behavior, no embellishment beyond what the source states.
- **NFR-DIC-004** (Mermaid TD-only preservation): Where translated content includes a Mermaid diagram, the diagram's direction (`flowchart TD` / `graph TB`) **shall** remain unchanged; node labels are translated, diagram syntax and direction are not (per `.moai/docs/docs-site-i18n-rules.md` §17.2).
- **NFR-DIC-005** (no maintainer-path-leak reintroduction): Translated content **shall not** introduce any literal maintainer development path (e.g., `/Users/goos/...`) — a regression class already remediated by SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001 REQ-DPC-002/003.
- **NFR-DIC-006** (ko-baseline immutability): No `ko/`-locale file **shall** be modified by this SPEC — `ko` is the canonical baseline for both Item 1 and Item 2, and is already correct.

## Exclusions

### Out of Scope — SPEC-DEAD-CONFIG-001

- `.moai/specs/SPEC-DEAD-CONFIG-001/` is a separate, unrelated, parallel in-progress SPEC (config dead-code cleanup). This SPEC does not touch it.

### Out of Scope — ko-locale content

- `ko` is the canonical, already-correct baseline for both Item 1 (D5 backlog) and Item 2 (REQ-DPC-006 content). No `ko/` file is read-write touched by this SPEC (only read, for translation grounding).

### Out of Scope — init-wizard.md language-picker labels

- Per REQ-DIC-002, the `Korean (한국어)` native-script label in `getting-started/init-wizard.md` (all 3 non-ko locales) is intentional and correct as-is. This SPEC explicitly does NOT "fix" it — doing so would be a regression, not a fix.

### Out of Scope — Go source, template source, and any file outside docs-site/

- No changes to `internal/`, `cmd/`, `pkg/`, or `internal/template/templates/`. This SPEC touches ONLY `docs-site/content/{en,ja,zh}/` markdown files.

### Out of Scope — docs-site/hugo.toml and docs-site/static/

- Unlike the predecessor SPEC (which touched `hugo.toml` version strings and `static/robots.txt`), this SPEC's scope is content-translation-only. Neither file is touched here.

### Out of Scope — file-count parity changes

- No file is added or removed under `docs-site/content/`. The 99-files-per-locale baseline (NFR-DIC-001) is preserved exactly.

### Out of Scope — other Check-4 glossary findings

- At plan-phase, `scripts/docs-i18n-check.sh` reports exactly 1 residual error (the ja `Anthropic` finding fixed by REQ-DIC-005). If translating the 23 Item-1 files introduces a NEW glossary-term omission (a risk this SPEC's NFR-DIC-002 aims to prevent), that would need to be caught at run-phase verification and fixed within this same SPEC's scope (it is not a pre-existing, separately-tracked defect) — but no such finding is known or presumed to exist at plan-phase.

### Out of Scope — full hugo build/deploy §17.6 checklist

- SPEC-V3R6-DOCS-V3-REBUILD-001 already validated the full build/deploy pipeline. This SPEC verifies only that `hugo --minify` still completes cleanly after these targeted content edits — not the full Vercel-preview / language-switcher / sitemap checklist.

## Dependencies and References

- SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001 — the immediate predecessor SPEC; source of all 3 deferred items (D5 backlog, REQ-DPC-006 ko-first sequencing, Check-4 glossary defect).
- SPEC-V3R6-DOCS-V3-REBUILD-001 — root origin of the docs-site v3 rebuild; established the 99-file-per-locale baseline (M3) that NFR-DIC-001 preserves.
- SPEC-INVOCATION-MODEL-001 — original source of the 4 `/moai feedback` workflow enhancements described in ko's "피드백 설정" subsection, now being translated into en/ja/zh by REQ-DIC-003.
- `project_docsite_untranslated_backlog.md` (auto-memory) — original 2026-07-01 discovery of the D5 backlog; this SPEC's Ground Truth independently re-verifies (not blindly trusts) that snapshot and finds the init-wizard.md false positive it did not surface.
- `.moai/docs/docs-site-i18n-rules.md` §17 — 4-locale parity rules, glossary-preservation convention, Mermaid TD-only rule (SSOT for this SPEC's NFRs).
- `scripts/docs-i18n-check.sh` — the mechanical verification tool for Checks 1-4 (file parity, frontmatter title, H1 heading, glossary term preservation).

## Acceptance Criteria

See `acceptance.md` for per-requirement testable/verifiable acceptance criteria and Given-When-Then scenarios.
