# sync-audit verdict — SPEC-DOCS-V313-CATCHUP-001 (card t274)

- **Auditor**: sync-auditor (independent, fresh context)
- **Audited tree**: `bed33bbde` (branch WT-v313-docs, run-phase terminal commit M5) — pinned; see §Race note for why pinning was required
- **Run-phase range**: `311d5498a..bed33bbde` (M1 `47270bd72` → M2 `e5808b61f` → M3 `8eeecbf0f` → M4 `5d68cdac9` → M5 `bed33bbde`), 60 files +898/−266
- **Date**: 2026-08-26
- **Verdict**: **PASS** — harmonic mean **0.937** (threshold 0.80); no blocking findings

## Evaluation Report

SPEC: SPEC-DOCS-V313-CATCHUP-001
Overall Verdict: **PASS**

### Dimension Scores

| Dimension | Score | Verdict | Evidence (re-measured this run, tree `bed33bbde` unless noted) |
|-----------|-------|---------|----------|
| Functionality (40%) | 97/100 | PASS | 9/9 ACs re-executed: AC-004 cells (a)–(f) all exact match (`git grep -c` pinned: README `🗿 v3.1.3` = 2×4 files, faq arrow-form = 1×4, moai-feedback `v3.1.3` = 1×4 + `v10.8.0` = 0×4, statusline/claude-cloud stale `v3.1.2` = 0×8); AC-002 per-locale `todo.enabled` 2 files ×4 locales (8 total), `awk 'FNR==352'` = 4, scrub/symlink(현지어)/inconclusive all 1+ per locale; AC-003 `opus / max` 0×8 files, `manager-lead` 3/4 per page; AC-006 Go/template/hook diff = 0 lines (both baselines `311d5498a`, `e07a6d0f4`); AC-007 structural diff = 4 new pages (63 lines each) + main.yaml 6 lines only; AC-008 unique IDs = 26; AC-009 H2 = 12×4 |
| Security (25%) | 95/100 | PASS | URL blacklist grep (axis 3) = 0 hits pinned to `bed33bbde` incl. README×4; secret-pattern scan of full docs diff: `api[_-]?key|secret|password|bearer|sk-` → 2 lines, both the same A5 scrubbing-contract prose ("sensitive values such as secrets, tokens") — feature description, not a credential; `token` 8 hits all model-token prose; claude-cloud install command reads `@v3.1.3` exactly (1×4 locales); no new command/injection surface introduced |
| Craft (20%) | 90/100 | PASS | 8-axis exit gate independently re-run (see §Axes); new page factual fidelity verified against CHANGELOG A1–A4 **and against the code** (`internal/codexadapter/events.go` "Six rows are adapted" comment + 6 `Adapted: true` rows; `output.go` inertKeys `systemMessage`/`continue`/`stopReason`; `ErrUnadapted` reject-not-swallow protocol; `.agents/` absent from `git ls-tree bed33bbde` — "never committed" claim holds); icon shortcode used, body-emoji 0 on new pages; F1 (below) is the only craft deduction |
| Consistency (15%) | 93/100 | PASS | 4-locale derivation mechanically identical where it must be: dispatcher-arg patterns = 5/5/5/5, `0.147.0` = 1×4, `systemMessage` = 1×4 (all pinned `bed33bbde`); hugo renders all four pages (77–82 KB index.html each); nav entry in main.yaml carries all 4 locale name maps; `school` icon SVG case pre-existed at `layouts/partials/menu.html:65` (last touched `77dc2043a`, prior card) — executor's claim verified; F1 marker-style drift on the new ko page is also a consistency deduction |

**Harmonic mean** = 4/(1/97 + 1/95 + 1/90 + 1/93) = **0.937**

### 8-axis gate re-run (AC-DVC-005)

| # | Axis | Auditor command | Observed |
|---|------|-----------------|----------|
| 1 | hugo build | `git archive bed33bbde docs-site \| tar -x -C /tmp/t274-audit` → `hugo --source … --minify --gc` | rc=0, WARN/ERROR = 0 lines, KO 185 / EN·JA·ZH 183 pages — byte-for-byte the executor's page counts |
| 2 | sitemap | `ls …/public/sitemap.xml` | exists (396 B sitemapindex, minified form; executor recorded 572 B unminified-equivalent — size delta is rendering form, axis is existence) |
| 3 | URL blacklist | `git grep -c 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' bed33bbde -- docs-site/content README×4 \| grep -v ':0$' \| wc -l` | **0** |
| 4 | Mermaid LR/RL | `git grep -c 'flowchart LR\|graph LR\|flowchart RL\|graph RL' bed33bbde -- docs-site/content \| grep -v ':0$' \| wc -l` | **0** |
| 5 | 4-locale parity | `bash scripts/docs-i18n-check.sh` (script byte-identical to `bed33bbde`, diff 0) | Errors 0 / Warnings 0, all 4 locales pass parity, frontmatter, H1, glossary |
| 6 | README heading parity | `git grep -c '^## ' bed33bbde -- README×4` | 12 / 12 / 12 / 12 |
| 7 | body emoji | `git grep -Pn '[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]' bed33bbde -- <4 new pages> \| wc -l`; plus full-diff emoji added vs removed | new pages **0**; full diff added 32 = removed 32 (**net 0**) — all 🗿 version-example substitutions, confirming the executor's exact claim |
| 8 | version-sync | AC-004 cells (a)–(f) + `git grep -c 'Release-v3\.1\.3' bed33bbde -- README×4` | all green; badges `Release-v3.1.3` = 1×4 |

### Executor-claim verification highlights

- **BSD sed multi-file quirk (AC-002 note)**: independently reproduced — `sed -n '352p' A B C D | grep -c 'analyze'` → **1** (BSD sed continues line numbering across files), while per-file `sed` → 1/1/1/1 and `awk 'FNR==352'` → **4**. The executor's reasoning holds and the awk form is the correct equivalent. The acceptance criterion's literal command form is platform-fragile, not wrongly judged.
- **Historical-citation preservation**: `grep -c 'v3\.1\.1부터\|v3\.1\.1에' README.ko.md` = 5 now, and **5 against `git show e07a6d0f4:README.ko.md`** (baseline re-measured by the auditor, not carried from §E.2).
- **A12 locale landing**: the English keyword `symlink` only greps in `en`; ko/ja/zh land the notification in localized terms (`심볼릭` / `シンボリック` / `符号链接`) — all 8 files (2 pages × 4 locales) verified, plus diff content read: matches CHANGELOG A12 (symlink→copy fallback + completion-summary notice for both `moai update` and `moai init`).

### Findings

- **F1** [MINOR] [optional] `docs-site/content/ko/advanced/codex-dual-harness.md:23` (+3 more sites in the same file) — emphasis-marker style: `**결정적으로(deterministically, 같은 입력에 언제나 같은 출력)**` puts the parenthetical **inside** the marker; SPEC §3 states `**단어** (Word)` — parenthetical outside. Existing ko pages are overwhelmingly outside-marker (`multi-model-audit.md`, `config-sections.md`, `moai-feedback.md`: 0 inside-marker sites; `profile-matrix.md`: 1). Rendering and meaning unaffected. Required fix (optional, cosmetic): move 4 parentheticals outside their markers.
- **F2** [MINOR] [optional] progress.md §E.2 M5 gate table records sitemap as "572 bytes"; auditor's pinned minified build produced a 396 B sitemapindex. Existence-axis green either way; the byte figure is form-dependent, not a defect.

No BLOCKING findings. No SHOULD-FIX findings. Both findings are optional per the finding-consumption discipline.

### Race note (measurement integrity, not a SPEC defect)

Mid-audit, the branch moved: `bed33bbde` → `0044c7a83` ("Merge remote-tracking branch 'origin/main' into WT-v313-docs", pulling #1648/#1655/#1656/#1657). The auditor's first unpinned scope checks (`git diff --name-only 311d5498a..HEAD` showing `internal/graph/**`, foreign SPEC dirs, `main.yaml +18`) were contaminated by that merge. **All verdict-bearing measurements above were re-executed pinned to `bed33bbde`** (`git grep bed33bbde`, `git show bed33bbde:`, `git archive bed33bbde`). Post-merge working-tree greps that overlapped t274-owned files happened to agree with the pinned results, but the verdict rests on the pinned set. Downstream (sync/PR): the merge commit is already part of the branch history, so the PR diff against main will show only the card's commits — no action needed; do not rebase it away.

### Recommendations

- Proceed to sync close + PR (Route B; repo-local all-tier PR policy). The branch's origin/main merge is benign for the PR diff.
- F1 is safe to fold into any later docs pass; it does not gate this close.

---

## Auditor's 5-section evidence summary

- **Claim**: run-phase `bed33bbde` satisfies all 9 acceptance criteria; 8-axis exit gate green with zero new violations; executor's §E.2 PASS claims are accurate.
- **Evidence**: every AC's judgement command re-executed this run (pinned to `bed33bbde` after the mid-audit branch move); verbatim outputs in the tables above; hugo build re-run from a `git archive` extraction (rc=0, 0 warnings).
- **Baseline-attribution**: all greps via `git grep/show/diff` keyed to commit `bed33bbde`; hugo build against `/tmp/t274-audit` extracted from `git archive bed33bbde` (with `enableGitInfo=false` applied **only to the /tmp copy** — the repo's `hugo.toml` is untouched); historical-citation baseline re-read from `e07a6d0f4`.
- **Gaps**: (1) `docs-i18n-check.sh` was executed on the post-merge working tree (script itself verified byte-identical to `bed33bbde`), not against the pinned extraction — its zero-error result plus the pinned per-locale greps and the successful pinned hugo build jointly cover the parity axis. (2) AC-DVC-001 is procedural (pre-flight re-observation); the auditor verified its recorded evidence exists and spot-checked the CHANGELOG structure rather than re-observing all 26 cells from scratch. (3) zh/ja new-page prose read less deeply than ko/en (mechanical fidelity counters + structure verified; full prose review limited).
- **Residual-risk**: the sitemap lists only locale-index entries (existing pages like `autonomy-tier` are also absent from it) — pre-existing site-wide behavior outside this SPEC, worth a separate card if sitemap completeness ever matters; the post-merge branch state means any *future* re-measurement must pin `bed33bbde` (or the eventual merge SHA) to stay attributable.
