# Implementation Plan — SPEC-V3R6-DOCS-I18N-COMPLETION-001

## §A Context

Follow-up SPEC bundling 3 items deferred by SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001 (`status: completed`, sync commit `85391a770`): the D5 non-ko untranslated-content backlog, the en/ja/zh translation of REQ-DPC-006's new ko `moai-feedback.md` content, and the residual Check-4 glossary defect. Docs-content-only — no Go code, no `internal/template/templates/`, no `hugo.toml`/`static/`.

**Tier: L** (judgment call, documented — and a deliberate contrast with the predecessor SPEC's Tier M reasoning). File count alone (23 genuine Item-1 translation files + 3 Item-2 files + 1 Item-3 file = **27 touched files** — excluding the 3 explicitly-untouched `init-wizard.md` false-positive files, which are NOT part of the workload per REQ-DIC-002) exceeds the Tier-L ">15 files" threshold outright. Unlike the predecessor SPEC — where a similarly high raw file count (~79) was justified down to Tier M because 78/79 files needed only a single mechanical line insertion or token substitution — **this SPEC's dominant work (23 of 27 files) is genuine prose translation**, not mechanical replacement: ~15,121 Hangul characters across 23 files must be read, understood, and re-expressed as natural target-language prose while preserving structure, glossary terms, and factual grounding. This is the opposite complexity profile from the predecessor SPEC, and the LOC/complexity guidance in `spec-workflow.md` § SPEC Complexity Tier points toward Tier L for exactly this reason: high file count AND non-trivial per-file complexity, not high file count with trivial per-file complexity. Only Item 3 (1 file, mechanical glossary-attribution-sentence addition) and 2 of Item 2's 3 files (ja/zh, mechanical diagram-label correction, layered on top of their REQ-DIC-003 prose translation which all 3 Item-2 files require) carry a mechanical component; the remaining 23 Item-1 files require pure prose authorship, with no mechanical shortcut available.

**Artifact-set adaptation** (documented per the LEAN Tier judgment being an implementer call, not blind table lookup): despite Tier L nominally calling for a 5-file set (+ design.md + research.md), this SPEC omits both. No architecture or design decision is being made (design.md would be empty ceremony — this is prose translation, not a structural change), and no separate codebase research is needed beyond what is already captured inline in spec.md's Ground Truth table (there is no research.md-worthy codebase investigation left to do; the 3 grounding source files for Item 2 were already read and cited in spec.md). The 4-file set (spec.md + plan.md + acceptance.md + progress.md) is therefore the pragmatic artifact set for this SPEC, justified independently on its own merits — no architecture decision needs recording, and no separate codebase research remains outstanding. (Note: the predecessor SPEC, SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001, was Tier M, whose nominal artifact set never included design.md/research.md to begin with — it is not a comparable precedent for "adapting away" those files, and this SPEC's 4-file set is not justified by analogy to it.)

Hybrid Trunk 1-person OSS policy applies (CLAUDE.local.md §23): all tiers push directly to `main`; no PR required.

## §B Known Issues (drift inventory feeding this SPEC)

| Surface | Drift | Target REQ |
|---------|-------|-----------|
| 7 pages × en, 8 pages × ja, 8 pages × zh (23 files total — see spec.md Ground Truth) | Entire pages (or near-entire, per Hangul char count) still in Korean despite being under the en/ja/zh locale directories | REQ-DIC-001 |
| `getting-started/init-wizard.md` × 3 locales | FALSE POSITIVE — flagged by the raw Hangul grep but contains only an intentional native-script language-picker label (`Korean (한국어)`), already correctly surrounded by translated prose | REQ-DIC-002 (exclusion, do-not-touch) |
| `utility-commands/moai-feedback.md` × en/ja/zh | Missing translated equivalent of ko's new "피드백 설정" subsection (4 `/moai feedback` enhancements) | REQ-DIC-003 |
| `utility-commands/moai-feedback.md` × ja/zh only | Agent-chain diagram + role table misattribute GitHub issue creation to "sync-auditor" (en already correct; ko already fixed by the predecessor SPEC) | REQ-DIC-004 |
| `claude-code/agentic/best-practices.md` × ja only | References section (`## 参考資料`) is a bare link-only bullet, missing the "based on Anthropic's official documentation" attribution sentence present in ko/en/zh | REQ-DIC-005 |

## §C Pre-flight (run-phase entry preconditions)

1. `git status --porcelain docs-site/` clean (no uncommitted drift from a parallel session).
2. `git rev-parse HEAD` matches the plan-phase baseline (no intervening docs-site commit landed since plan-phase Ground Truth was captured).
3. Re-run `grep -rlP '[가-힣]' docs-site/content/{en,ja,zh}/` at the start of Item 1's first milestone to reconfirm the 26-file list (23 genuine + 3 false-positive) is unchanged since plan-phase.
4. Re-run `grep -n "sync-auditor" docs-site/content/{ko,en,ja,zh}/utility-commands/moai-feedback.md` at the start of Item 2's milestone to reconfirm the ja/zh-only (not en) misattribution finding is unchanged.
5. Re-run `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` at the start of Item 3's milestone to reconfirm the sole residual Check-4 finding is unchanged.

## §D Constraints

- Docs-content-only: no changes to `internal/`, `cmd/`, `pkg/`, or `internal/template/templates/`.
- No changes to any `ko/`-locale file (canonical baseline for both Item 1 and Item 2 — NFR-DIC-006).
- No changes to `docs-site/hugo.toml` or `docs-site/static/` (out of this SPEC's scope; already corrected by the predecessor SPEC).
- Do NOT alter the `Korean (한국어)` label in `getting-started/init-wizard.md` — REQ-DIC-002 explicitly carves this out as a false positive, not a defect.
- Every translated sentence must be traceable to the `ko/` source (Item 1) or the 3 named grounding files (Item 2) — no invented behavior (NFR-DIC-003).
- Preserve all 7 canonical glossary terms verbatim in translated output (NFR-DIC-002) — spot-check via `scripts/docs-i18n-check.sh` Check 4 after each locale's translation milestone.
- Preserve Mermaid diagram direction (`TD`/`TB` only, never `LR`) when translating diagram node labels (NFR-DIC-004).
- 99-files-per-locale count must not change (NFR-DIC-001) — no file added or removed, only content translated in place.
- hugo-geekdoc theme/CSS/layout untouched — content-only edits.

## §E Self-Verification

Plan-phase self-verification is recorded in `progress.md` §E.1. Run-phase and sync-phase evidence are populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4), per the Status Transition Ownership Matrix.

## §F Milestones

Priority-based ordering; no time estimates. Sequenced from lowest-risk/fastest (Item 3, Item 2) to highest-volume (Item 1, split by locale rather than by mixed file-batches — see rationale below), closing with a full verification sweep.

**Milestone-granularity rationale (Item 1 split by locale, not by file-count batch)**: Item 1's 23 genuine-translation files are split into 3 per-locale milestones (M3=en, M4=ja, M5=zh) rather than cross-locale file-count batches (e.g., "batch of 8 files spanning 2-3 locales"). Rationale: (a) each locale's translation work benefits from staying in one target-language "voice" across a session rather than context-switching between Japanese and Chinese prose mid-milestone; (b) per-locale glossary/terminology consistency is easier to self-verify when one locale is fully drafted before moving to the next; (c) it produces a natural commit boundary per locale (`feat(SPEC-...): M3 en translation` etc.), matching the predecessor SPEC's per-milestone-commit convention; (d) if a run is interrupted (context-window threshold, session boundary), a locale-complete checkpoint is a cleaner resume point than a partially-translated cross-locale batch.

### M1 · Check-4 glossary fix (Priority: High — fastest, lowest risk, do first)

- **M1.1** Re-verify `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` still shows exactly 1 residual error (the ja Anthropic finding) before editing.
- **M1.2** Edit `docs-site/content/ja/claude-code/agentic/best-practices.md`'s `## 参考資料` section: replace (or augment) the bare bullet-list link with a natural Japanese attribution sentence containing the literal term "Anthropic", structurally mirroring the ko/en/zh pattern ("本ガイドは Anthropic の公式 [Best practices for Claude Code](...) ドキュメントを基に作成されています。" or an equivalent natural rendering) (REQ-DIC-005).
- **M1.3** Re-run `scripts/docs-i18n-check.sh` (default strict mode) and confirm Check 4 now reports 0 errors for this file.

### M2 · REQ-DPC-006 en/ja/zh moai-feedback.md translation (Priority: High — small prose, ko source already fully read)

- **M2.1** Re-read ko's "피드백 설정" subsection (L79-104) to reconfirm content is unchanged since plan-phase.
- **M2.2** Add a translated "Feedback Settings" (en) / equivalent H2 subsection to `content/en/utility-commands/moai-feedback.md`, with the 4 enhancement H3s translated as natural English prose (en already has the correct agent-chain diagram — no REQ-DIC-004 fix needed for en).
- **M2.3** Add a translated equivalent H2 subsection to `content/ja/utility-commands/moai-feedback.md` AND fix the agent-chain diagram/table (L141/L150-ish) to replace "sync-auditor エージェント" / "sync-auditor" with "manager-docs エージェント" / "manager-docs" (REQ-DIC-003 + REQ-DIC-004).
- **M2.4** Add a translated equivalent H2 subsection to `content/zh/utility-commands/moai-feedback.md` AND fix the agent-chain diagram/table (L143/L152-ish) to replace "sync-auditor Agent" / "sync-auditor" with "manager-docs Agent" / "manager-docs" (REQ-DIC-003 + REQ-DIC-004).
- **M2.5** Verify: `grep -c 'feedback.repository\|modu-ai/moai-adk' docs-site/content/{en,ja,zh}/utility-commands/moai-feedback.md` each ≥ 1; `grep -c 'sync-auditor' docs-site/content/{en,ja,zh}/utility-commands/moai-feedback.md` each `0`.

### M3 · Item 1 — en translation (Priority: Medium — 7 files)

**Per-file self-verification obligation (D7 — applies to every file below, and identically to M4/M5)**: Before moving to the next file, the implementing agent MUST manually cross-check every factual claim in the just-translated output (version numbers, command names, file paths, behavioral descriptions) against the corresponding `ko/` source paragraph, confirm no new claim is introduced that is not traceable to ko, and record the result as a per-file line in `progress.md` §E.2 (e.g. `harness-engineering.md (en): traced — no invented claims found` or, if a discrepancy is found, the discrepancy and its resolution). This check is NOT deferred to a single end-of-milestone or end-of-SPEC spot check — see acceptance.md AC-DIC-001f.

- **M3.1** Translate `content/en/advanced/catalog-system.md` from `content/ko/advanced/catalog-system.md`.
- **M3.2** Translate `content/en/advanced/harness-profiles.md` from ko.
- **M3.3** Translate `content/en/core-concepts/constitution.md` from ko.
- **M3.4** Translate `content/en/core-concepts/harness-engineering.md` from ko.
- **M3.5** Translate `content/en/getting-started/profile.md` from ko (including its H1, which is currently Korean).
- **M3.6** Translate `content/en/getting-started/windows-guide.md` from ko.
- **M3.7** Translate `content/en/guides/ci-autonomy.md` from ko.
- **M3.8** Do NOT touch `content/en/getting-started/init-wizard.md` (REQ-DIC-002 — already correct, its 12 remaining Hangul chars are the intentional language-picker label).
- **M3.9** Verify: `grep -rlP '[가-힣]' docs-site/content/en/` returns ONLY `getting-started/init-wizard.md`.

### M4 · Item 1 — ja translation (Priority: Medium — 8 files)

**Per-file self-verification obligation (D7 — same discipline as M3, per file, not batched)**: For each of the 8 files below, cross-check the translated output's factual claims against the ko source paragraph-by-paragraph before moving to the next file, and record the per-file result in `progress.md` §E.2 (see acceptance.md AC-DIC-001f).

- **M4.1** Translate the same 7 pages as M3 into `content/ja/` from ko, PLUS `content/ja/advanced/hooks-reference.md` (ja/zh-only backlog item; en is already translated per the predecessor discovery). Apply the per-file self-verification obligation above to each of the 8 files individually as it completes.
- **M4.2** Verify: `grep -rlP '[가-힣]' docs-site/content/ja/` returns ONLY `getting-started/init-wizard.md`.

### M5 · Item 1 — zh translation (Priority: Medium — 8 files)

**Per-file self-verification obligation (D7 — same discipline as M3/M4, per file, not batched)**: For each of the 8 files below, cross-check the translated output's factual claims against the ko source paragraph-by-paragraph before moving to the next file, and record the per-file result in `progress.md` §E.2 (see acceptance.md AC-DIC-001f).

- **M5.1** Translate the same 8 pages as M4 into `content/zh/` from ko. Apply the per-file self-verification obligation above to each of the 8 files individually as it completes.
- **M5.2** Verify: `grep -rlP '[가-힣]' docs-site/content/zh/` returns ONLY `getting-started/init-wizard.md`.

### M6 · Final verification sweep (Priority: Medium — closing milestone)

- **M6.1** `grep -rlP '[가-힣]' docs-site/content/{en,ja,zh}/` returns exactly 3 files total (one `init-wizard.md` per locale) — down from the plan-phase baseline of 26.
- **M6.2** `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` (and separately, the default strict-mode invocation) both report **0 errors, 0 warnings** — the docs-site i18n backlog is fully clean for the first time since SPEC-V3R6-DOCS-V3-REBUILD-001.
- **M6.3** `find docs-site/content/<L> -name '*.md' | wc -l` for ko/en/ja/zh each still returns `99` (NFR-DIC-001 unchanged).
- **M6.4** `cd docs-site && hugo --minify` completes with zero warnings.
- **M6.5** `grep -rn goos docs-site/content/{en,ja,zh}/` (spot-check for accidental path-leak reintroduction) returns zero matches across all newly-translated files.
