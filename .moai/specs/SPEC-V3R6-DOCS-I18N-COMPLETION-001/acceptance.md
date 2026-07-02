# Acceptance Criteria — SPEC-V3R6-DOCS-I18N-COMPLETION-001

All criteria are testable via shell command, observable artifact, or (where noted) manual cross-check by the implementing agent. `<L>` denotes locale; commands assume CWD = repo root unless noted.

## §D AC Matrix (per requirement)

Each AC below is authored as a GEARS "shall" sentence, followed by an em-dash and the literal verification command + expected value.

### REQ-DIC-001 — Item 1 genuine-translation completion (23 files)

- **AC-DIC-001a**: When `grep -rlP '[가-힣]' docs-site/content/en/` runs against the translated content, it **shall** return exactly 1 file (`getting-started/init-wizard.md`) — down from the plan-phase baseline of 8, confirming all 7 genuine en files are translated.
- **AC-DIC-001b**: When the same grep runs against `docs-site/content/ja/`, it **shall** return exactly 1 file — down from the plan-phase baseline of 9, confirming all 8 genuine ja files (including `hooks-reference.md`) are translated.
- **AC-DIC-001c**: When the same grep runs against `docs-site/content/zh/`, it **shall** return exactly 1 file — down from the plan-phase baseline of 9, confirming all 8 genuine zh files are translated.
- **AC-DIC-001d** (structural preservation, manual cross-check, full coverage per D8): For **each of the 23** translated (locale, file) pairs — NOT a representative sample — the translated version **shall** preserve the same H2/H3 heading count and code-fence count as its `ko/` source — verified via `grep -c '^#\{2,3\} '` and `grep -c '^```'` compared between the ko source and the translated output, run individually for all 23 pairs. Full-coverage verification is adopted here (rather than a 1-file/locale sample) because the check is mechanically cheap — a simple shell loop over the 23-file list — so there is no cost/risk trade-off that would justify sampling.
- **AC-DIC-001e** (glossary preservation): When `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` Check 4 runs after all Item-1 translation is complete, it **shall** report 0 NEW glossary-term-missing findings attributable to the 23 translated files (the only finding permitted at this checkpoint, before M1's fix lands, is the pre-existing ja `Anthropic` finding — see REQ-DIC-005's own AC).
- **AC-DIC-001f** (verification-claim-integrity per-file traceability, manual cross-check, incremental — new, D7): For each of the 23 translated (locale, file) pairs, before the implementing agent moves to the next file, every factual claim in the translated output (version numbers, command names, file paths, behavioral descriptions) **shall** be manually cross-checked against the corresponding `ko/` source paragraph, confirming no new claim is introduced that is not traceable to ko — recorded as an incremental per-file entry in `progress.md` §E.2 as each file is completed (e.g. `harness-engineering.md (en): traced — no invented claims found`, or the discrepancy found and its resolution). This is NOT satisfied by a single end-of-SPEC spot check across all 23 files; the per-file record in §E.2 is the evidence artifact. Cross-references NFR-DIC-003 (no unobserved-claim) and plan.md M3/M4/M5's per-file self-verification obligation.

### REQ-DIC-002 — init-wizard.md exclusion (false-positive carve-out)

- **AC-DIC-002a**: The en/ja/zh `getting-started/init-wizard.md` page **shall** retain the `Korean (한국어)` label unchanged — verified via `grep -c 'Korean (한국어)' docs-site/content/<L>/getting-started/init-wizard.md` returning `4` for en/ja/zh (unchanged from the plan-phase baseline of 4 occurrences per locale).
- **AC-DIC-002b** (reworded per D9 — direct operationally-testable claim, cross-referenced to M6.1): The en/ja/zh `getting-started/init-wizard.md` file **shall** continue to be matched by `grep -rlP '[가-힣]' docs-site/content/<L>/getting-started/init-wizard.md` (non-empty match, unchanged) in every locale — this is the accepted, documented exception established by REQ-DIC-002, not a residual defect. The binary pass condition is cross-referenced to plan.md M6.1's closing assertion: `grep -rlP '[가-힣]' docs-site/content/{en,ja,zh}/` **shall** return exactly `3` total matches (one `init-wizard.md` per locale — matching the corrected 27-touched-file arithmetic in plan.md §A), confirming `init-wizard.md`'s continued appearance is the sole, expected, accounted-for exception rather than an independent grep target of its own.

### REQ-DIC-003 — moai-feedback.md en/ja/zh content translation

- **AC-DIC-003a**: When a reader opens `docs-site/content/{en,ja,zh}/utility-commands/moai-feedback.md`, each page **shall** describe the configurable target-repository behavior — verified via `grep -c 'feedback.repository\|modu-ai/moai-adk' docs-site/content/<L>/utility-commands/moai-feedback.md` ≥ 1 for en/ja/zh.
- **AC-DIC-003b**: Each page **shall** describe the guaranteed + best-effort diagnostic collection — verified via keyword-presence grep per locale: en `grep -ci 'go version\|go toolchain'` ≥ 1; ja `grep -c 'Go ツールチェーン\|go version'` ≥ 1; zh `grep -c 'Go 工具链\|go version'` ≥ 1 (locale-native terminology acceptable, not a literal English match requirement).
- **AC-DIC-003c**: Each page **shall** describe the duplicate-issue candidate-report step — verified via `grep -ci 'dup\|duplicate\|重复\|重複'` ≥ 1 per locale (locale-native term acceptable).
- **AC-DIC-003d** (verification-claim-integrity check, manual cross-check): Every factual claim added under AC-DIC-003a/b/c **shall** be traced by the implementing agent to a line in `.claude/skills/moai/workflows/feedback.md`, `internal/config/feedback_accessors.go`, or `internal/template/templates/.moai/config/sections/feedback.yaml` — no invented behavior. Recorded in `progress.md` §E.2, not grep-automatable.

### REQ-DIC-004 — sync-auditor misattribution fix (ja/zh only)

- **AC-DIC-004a**: The ja `utility-commands/moai-feedback.md` page **shall not** contain the string "sync-auditor" — verified via `grep -c 'sync-auditor' docs-site/content/ja/utility-commands/moai-feedback.md` returning `0` (down from the plan-phase baseline of 2 occurrences).
- **AC-DIC-004b**: The zh `utility-commands/moai-feedback.md` page **shall not** contain the string "sync-auditor" — verified via `grep -c 'sync-auditor' docs-site/content/zh/utility-commands/moai-feedback.md` returning `0` (down from the plan-phase baseline of 2 occurrences).
- **AC-DIC-004c** (regression guard — en unaffected): The en `utility-commands/moai-feedback.md` page **shall** continue to show `0` occurrences of "sync-auditor" (it was already correct at plan-phase; this AC confirms no accidental reintroduction) — verified via `grep -c 'sync-auditor' docs-site/content/en/utility-commands/moai-feedback.md` returning `0`.

### REQ-DIC-005 — Check-4 glossary fix (ja best-practices.md)

- **AC-DIC-005a**: The ja `claude-code/agentic/best-practices.md` page **shall** contain the term "Anthropic" verbatim — verified via `grep -c 'Anthropic' docs-site/content/ja/claude-code/agentic/best-practices.md` returning ≥ `1` (up from the plan-phase baseline of `0`).
- **AC-DIC-005b**: When `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` Check 4 runs, it **shall** report 0 errors for `ja/claude-code/agentic/best-practices.md` — verified via the script's Check-4 section output.
- **AC-DIC-005c**: When `scripts/docs-i18n-check.sh` runs in its default strict mode (no `DOCS_I18N_STRICT=0` override) after M1 lands, it **shall** exit `0` for this specific finding (the previously-guaranteed exit `1` from the predecessor SPEC's Definition of Done no longer applies once this SPEC's M1 completes) — verified via `bash scripts/docs-i18n-check.sh; echo $?` returning `0`, PROVIDED Item 1's translation (M3-M5) has not introduced any new Check-4 finding (see AC-DIC-001e).

### NFR-DIC-004 — Mermaid diagram-direction preservation (new section — D6 fix)

Scope note: a fresh grep of the ko source tree (`grep -n 'mermaid\|flowchart\|graph TD\|graph TB\|graph LR' docs-site/content/ko/**/*.md`) run at revision time confirms exactly **2 files** in Item 1 + Item 2's scope contain a Mermaid diagram: `core-concepts/harness-engineering.md` (Item 1 — 2 diagrams: `graph TB` at L21, `graph TD` at L86) and `utility-commands/moai-feedback.md` (Item 2 — 2 diagrams, both `flowchart TD`, at L59-60/150-151 in ko; L59-60/124-125 in en/ja, L61-62/126-127 in zh). No other file among the 23 Item-1 genuine-translation files or the 3 Item-2 files contains a Mermaid block. Item 3's `best-practices.md` also contains 1 Mermaid diagram, but Item 3's edit is scoped only to the References section (REQ-DIC-005) and never touches the diagram — it is therefore out of scope for the two ACs below, per the task's explicit "Item 1 + Item 2" scoping.

- **AC-DIC-006a** (harness-engineering.md diagram-direction parity — Item 1, en/ja/zh): For each locale `<L>` in {en, ja, zh}, when `grep -c 'graph TB' docs-site/content/ko/core-concepts/harness-engineering.md` and the identical grep against `docs-site/content/<L>/core-concepts/harness-engineering.md` run, both **shall** return `1` (direction-parity, unchanged); the same parity check **shall** hold for `grep -c 'graph TD'`, both returning `1`. Additionally, `grep -c 'graph LR\|flowchart LR' docs-site/content/<L>/core-concepts/harness-engineering.md` **shall** return `0` for every locale (absence check — no direction drift introduced during translation).
- **AC-DIC-006b** (moai-feedback.md diagram-direction parity — Item 2, en/ja/zh, especially ja/zh where REQ-DIC-004 edits diagram labels): For each locale `<L>` in {en, ja, zh}, when `grep -c 'flowchart TD' docs-site/content/ko/utility-commands/moai-feedback.md` and the identical grep against `docs-site/content/<L>/utility-commands/moai-feedback.md` run, both **shall** return `2` (2 diagrams, direction unchanged even where ja/zh's node/label text is edited for the sync-auditor→manager-docs fix). Additionally, `grep -c 'flowchart LR\|graph LR' docs-site/content/<L>/utility-commands/moai-feedback.md` **shall** return `0` for every locale (absence check).

## Given-When-Then Scenarios

### Scenario 1 — init-wizard.md is correctly recognized as a non-defect, not "fixed"

**Given** the raw Hangul grep flags `getting-started/init-wizard.md` (all 3 non-ko locales) as part of the 26-file D5 backlog
**When** the implementing agent inspects the file's actual Hangul content (not just the grep hit count) and finds it is solely the `Korean (한국어)` language-picker label, correctly surrounded by already-translated prose
**Then** the agent does NOT translate or remove the label — the file is explicitly excluded from Item 1's genuine-translation scope (REQ-DIC-002), and the final verification (M6.1) expects exactly 3 residual Hangul-flagged files (1 per locale), not 0

### Scenario 2 — ja/zh moai-feedback.md fix mirrors the ko fix without assuming symmetry

**Given** the task description hypothesizes the sync-auditor misattribution "may or may not" exist in en/ja/zh
**When** the implementing agent independently greps each of en/ja/zh (not just assumes all 3 need the fix, and not just fixes the locale mentioned first)
**Then** the agent discovers en already lacks the misattribution (no fix needed) while ja and zh both still have it (fix needed in exactly 2 files, not 3), and applies the fix only where the grep confirms it is present

### Scenario 3 — ja Anthropic fix addresses structural absence, not mistranslation

**Given** the ja `best-practices.md` References section is NOT a mistranslation of the ko sentence but a structurally different bare-link-only bullet section with no attribution sentence at all (5 lines shorter than ko/en/zh)
**When** the implementing agent adds a natural Japanese attribution sentence containing "Anthropic" (rather than trying to find-and-replace a mistranslated word that does not exist)
**Then** the fix closes the Check-4 gap by adding new content structurally consistent with the ko/en/zh pattern, not by "correcting" an existing (nonexistent) mistranslation

## Edge Cases

- **EC-1**: If a run-phase re-verification (plan.md §C.3) finds the 26-file D5 list has drifted since plan-phase (e.g., a parallel session already translated some files), the implementing agent must re-derive the actual current state via a fresh grep rather than blindly applying the plan-phase snapshot, and note the delta in `progress.md`.
- **EC-2**: If translating any of the 23 genuine files accidentally introduces a NEW Check-4 glossary-term omission (e.g., a translator drops "MoAI-ADK" mid-sentence), M6.2's "0 errors, 0 warnings" closing criterion will catch it — the implementing agent must fix it within this same SPEC's scope before declaring M6 complete (per spec.md Exclusions' "Out of Scope — other Check-4 glossary findings" carve-out, which only excludes PRE-EXISTING findings, not ones this SPEC's own edits introduce).
- **EC-3**: The `profile.md` page's H1 heading itself is Korean in en/ja/zh (per spec.md Ground Truth) — the H1 text must also be translated, not just the body prose; a translation that leaves the H1 untranslated while translating the body is incomplete.
- **EC-4**: If a translated page contains a Mermaid diagram, verify the diagram still renders after translation (node label text translated, `flowchart TD`/`graph TB` direction and syntax untouched) — do not spot-check only prose paragraphs and skip diagram blocks. Concretely, this applies to exactly 2 files (`core-concepts/harness-engineering.md`, `utility-commands/moai-feedback.md`) per the NFR-DIC-004 scope note; see AC-DIC-006a/b for the literal verification commands.

## Quality Gate Criteria

- `scripts/docs-i18n-check.sh` (default strict mode) exits `0` after all milestones complete — first time since SPEC-V3R6-DOCS-V3-REBUILD-001 that the docs-site i18n backlog is fully clean.
- `grep -rlP '[가-힣]' docs-site/content/{en,ja,zh}/` returns exactly 3 files (1 `init-wizard.md` per locale) — an accepted, documented exception, not a residual defect.
- `find docs-site/content/<L> -name '*.md' | wc -l` returns `99` for all 4 locales (unchanged).
- `hugo --minify` (in `docs-site/`) completes with zero warnings.
- No `goos` string is introduced into any of the **27 touched files** across en/ja/zh (23 Item-1 translation files + 3 Item-2 files + 1 Item-3 file; disambiguated from the 26 raw-grep-*flagged* files per plan-phase Ground Truth — the 3 `init-wizard.md` false-positive files are explicitly excluded from the touched-file set per REQ-DIC-002 and are NOT edited by this SPEC).
- `git status --porcelain` scoped to `docs-site/` after all milestones shows only files enumerated in plan.md §F (no drive-by edits to `ko/`, `hugo.toml`, `static/`, or any `internal/` path).

## Definition of Done

- [ ] All 5 REQs (REQ-DIC-001 through REQ-DIC-005) have PASS evidence for their AC matrix rows above, INCLUDING AC-DIC-001f's incremental per-file traceability record (progress.md §E.2) and the NFR-DIC-004 diagram-direction AC-DIC-006a/b.
- [ ] `scripts/docs-i18n-check.sh` (default strict mode) exits `0` — zero residual errors, zero warnings.
- [ ] `grep -rlP '[가-힣]' docs-site/content/{en,ja,zh}/` returns exactly 3 files (the accepted init-wizard.md exception per locale).
- [ ] `hugo --minify` builds cleanly.
- [ ] 99-files-per-locale count unchanged across ko/en/ja/zh.
- [ ] en/ja/zh `moai-feedback.md` each document the 4 `/moai feedback` enhancements with no invented behavior and no residual "sync-auditor" misattribution.
- [ ] ja `best-practices.md` contains "Anthropic" verbatim.
- [ ] No Go code, template source, `hugo.toml`, `static/`, or ko-locale file was modified.
- [ ] `.moai/specs/SPEC-DEAD-CONFIG-001/` untouched.
- [ ] `progress.md` §E.2/§E.3 populated by manager-develop at run-phase completion; §E.4 by manager-docs at sync-phase completion.
