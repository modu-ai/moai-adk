# Acceptance Criteria — SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001

All criteria are testable via shell command or observable artifact. `<L>` denotes the 4 locales `{ko,en,ja,zh}`. Commands assume CWD = repo root unless noted.

## §D AC Matrix (per requirement)

Each AC below is authored as a GEARS "shall" sentence (Ubiquitous / Event-driven `When` / State-driven `While`), followed by an em-dash and the literal verification command + expected value. The mechanical evidence (command + value) is preserved verbatim from the original AC authoring; only the sentence wrapper changed.

### REQ-DPC-001 — H1 heading restoration (D2)

- **AC-DPC-001a**: When `scripts/docs-i18n-check.sh` Check 3 runs against the docs-site content, the content **shall** report 0 "no H1 heading" errors — verified via `DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh 2>&1 | grep -c 'no H1 heading'` returning `0` (down from the plan-phase baseline of 64).
- **AC-DPC-001b**: For each of the 16 canonical pages, the inserted H1 text (first content line after frontmatter) **shall** match the frontmatter `title:` value for that locale — verified via `head -n 5 <file>` spot-check for at least 4 representative files (1 per locale) across at least 2 of the 16 pages.
- **AC-DPC-001c**: While the H1 insertion is applied across all 4 locales, the per-locale file count **shall** remain unchanged at 99 — verified via `find docs-site/content/<L> -name '*.md' | wc -l` returning `99` for each of ko/en/ja/zh, confirming no file was added or removed, only content inserted.

### REQ-DPC-002 — worktree page path-leak fix (D6)

- **AC-DPC-002a**: The published `worktree/examples.md` page **shall not** contain the maintainer path leak, across all 4 locales — verified via `grep -c goos docs-site/content/<L>/worktree/examples.md` returning `0` for all 4 locales.
- **AC-DPC-002b**: The published `worktree/faq.md` page **shall not** contain the maintainer path leak, across all 4 locales — verified via `grep -c goos docs-site/content/<L>/worktree/faq.md` returning `0` for all 4 locales.
- **AC-DPC-002c**: The published `worktree/examples.md` page **shall** use the generic placeholder path at the same occurrence count as the original leak — verified via `grep -c '/path/to/your-project' docs-site/content/<L>/worktree/examples.md` returning `4` for all 4 locales, confirming a 1:1 mechanical substitution, not a deletion.

### REQ-DPC-003 — inventory.md path-leak fix (new finding)

- **AC-DPC-003a**: The published `getting-started/inventory.md` page **shall not** contain the maintainer path leak, across all 4 locales — verified via `grep -c goos docs-site/content/<L>/getting-started/inventory.md` returning `0` for all 4 locales.
- **AC-DPC-003b**: The published `getting-started/inventory.md` page **shall** use the generic placeholder path — verified via `grep -c '/path/to/your-project' docs-site/content/<L>/getting-started/inventory.md` returning `2` for all 4 locales.
- **AC-DPC-003c**: When the path substitution is applied inside the JSON example block, the block **shall** remain valid JSON syntax — verified by extracting the fenced code block containing `"project_root": "/path/to/your-project"` and confirming it starts with `{` and ends with `}` with balanced braces (no broken quoting from the substitution).

### REQ-DPC-004 — robots.txt domain correction (D7)

- **AC-DPC-004a**: `docs-site/static/robots.txt` **shall** reference the sitemap at the docs site's own canonical domain — verified via `cat docs-site/static/robots.txt` showing `Sitemap: https://adk.mo.ai.kr/sitemap.xml` exactly.
- **AC-DPC-004b**: `docs-site/static/robots.txt` **shall not** reference the unrelated `cowork.mo.ai.kr` domain — verified via `grep -c cowork docs-site/static/robots.txt` returning `0`.

### REQ-DPC-005 — hugo.toml version bump (new finding)

- **AC-DPC-005a**: `docs-site/hugo.toml` **shall** declare the current release version — verified via `grep -n 'version = "v3.0.0-rc5"' docs-site/hugo.toml` matching L55 (or its shifted line number if the file changed elsewhere).
- **AC-DPC-005b**: `docs-site/hugo.toml` **shall not** contain any stale `rc2`/`rc3`/`rc4` version string — verified via `grep -n 'rc4\|rc2\|rc3' docs-site/hugo.toml` returning zero matches.
- **AC-DPC-005c**: No locale's homepage `_index.md` **shall** contain a stale `rc2`/`rc3`/`rc4` version string — verified via `grep -rn 'rc4\|rc2\|rc3' docs-site/content/*/_index.md` returning zero matches, confirming the prior post-close homepage fix is still intact and no new drift was introduced.
- **AC-DPC-005d**: `docs-site/hugo.toml` **shall** declare the release date matching the `v3.0.0-rc5` tag (2026-07-01), honoring the L54 SSOT convention that pairs `version` with `releaseDate` — verified via `grep -c 'releaseDate = "2026-07-01"' docs-site/hugo.toml` returning `1`, AND **shall not** retain the stale `2026-06-23` date — verified via `grep -c '2026-06-23' docs-site/hugo.toml` returning `0`.

### REQ-DPC-006 — ko moai-feedback.md content accuracy (new finding)

- **AC-DPC-006a**: When a reader opens `docs-site/content/ko/utility-commands/moai-feedback.md`, the page **shall** describe all 4 `/moai feedback` enhancements — (1) guaranteed diagnostics (MoAI version + OS), (2) best-effort diagnostics (Go version + orchestrator-passed error context), (3) duplicate-issue candidate-report step, (4) `gh`-failure graceful fallback with local draft save — verified via keyword-presence grep: `grep -c 'moai version\|uname' <file>` ≥ 1, `grep -c 'go version' <file>` ≥ 1, `grep -c '중복\|dedupe\|duplicate' <file>` ≥ 1 (Korean or transliterated term acceptable), `grep -c 'gh auth\|인증되지 않' <file>` ≥ 1.
- **AC-DPC-006b**: The page **shall not** misattribute GitHub issue creation to a "sync-auditor" agent — verified via `grep -c 'sync-auditor' docs-site/content/ko/utility-commands/moai-feedback.md` returning `0`.
- **AC-DPC-006c**: The page **shall** document the configurable target-repository behavior — verified via `grep -c 'feedback.repository\|modu-ai/moai-adk' docs-site/content/ko/utility-commands/moai-feedback.md` ≥ 1.
- **AC-DPC-006d** (verification-claim-integrity check): When any factual claim is added under AC-DPC-006a, the implementing agent **shall** trace it to a line in `.claude/skills/moai/workflows/feedback.md`, `internal/config/feedback_accessors.go`, or `internal/template/templates/.moai/config/sections/feedback.yaml` — no invented behavior. This is a manual cross-check performed by the implementing agent and recorded in `progress.md` §E.2, not a grep-automatable AC.
- **AC-DPC-006e** (iteration-2 addition — table accuracy): The page's "자동 수집되는 정보" table **shall not** list a "Claude Code 버전" row or a "현재 SPEC" row, and **shall** list the actual best-effort "Go 툴체인 버전" diagnostic in their place — verified via `grep -c 'Claude Code 버전\|현재 SPEC' docs-site/content/ko/utility-commands/moai-feedback.md` returning `0` AND `grep -c 'Go 툴체인\|Go 버전\|go version' docs-site/content/ko/utility-commands/moai-feedback.md` ≥ 1.
- **AC-DPC-006f** (iteration-3 addition — traces NFR-DPC-004, ko-first / ko-only scope boundary): The new REQ-DPC-006 configuration content **shall** land in the ko locale only — verified via `grep -c 'feedback.repository\|modu-ai/moai-adk' docs-site/content/{en,ja,zh}/utility-commands/moai-feedback.md` returning `0` for each of en/ja/zh (all 3 confirmed at `0` at plan-phase), confirming the en/ja/zh translation of the new content is deferred per Exclusions while ko carries it (AC-DPC-006c asserts the ko side ≥ 1). The ko-first run-phase *sequencing* sub-clause of NFR-DPC-004 is realized by plan.md milestone ordering (M4 is the sole content-authoring milestone), not a static end-state.

## Given-When-Then Scenarios

### Scenario 1 — H1 restoration does not break hugo-geekdoc's rendered title

**Given** a page like `content/ko/advanced/decision-memory.md` currently has a frontmatter `title:` but no body H1 (hugo-geekdoc renders the frontmatter title as the page's H1 in the browser)
**When** an explicit `# <title>` line is inserted as the first content line after frontmatter
**Then** `hugo --minify` still builds cleanly (no duplicate-heading warning) and the rendered page shows exactly one H1 (the explicit one now matches what hugo-geekdoc previously synthesized from frontmatter)

### Scenario 2 — Maintainer path substitution preserves example correctness

**Given** `worktree/examples.md` currently shows `$ cd /Users/goos/MoAI/moai-project` followed by output referencing the same path
**When** the literal path is replaced with `/path/to/your-project` in both the command and its corresponding output line
**Then** the example remains internally consistent (the substituted path appears identically in both the command and the output referencing it) and no `goos` string remains anywhere in the file

### Scenario 3 — ko feedback page content matches actual Go source, not the task's initially-suggested (incorrect) source

**Given** the task initially pointed at `internal/loop/{feedback,feedback_channel,go_feedback}.go` as the enhancement source, but those implement the unrelated Ralph Engine loop-feedback system
**When** the implementing agent re-reads `.claude/skills/moai/workflows/feedback.md` (the actual `/moai feedback` workflow definition) before writing new ko content
**Then** the new "피드백 설정" subsection accurately reflects the workflow's actual behavior (orchestrator-direct execution, `feedback.repository` config resolution, dedupe candidate-report, `gh`-fallback local-draft save) and does not describe any Ralph Engine loop-feedback concept (test/lint/coverage stagnation detection) by mistake

## Edge Cases

- **EC-1**: A page in the 16-page H1 list has a frontmatter `title:` containing special Markdown characters (e.g., a backtick or pipe) — the inserted H1 must escape or otherwise render correctly as Markdown, not break the heading syntax.
- **EC-2**: The `inventory.md` JSON example block's `"project_root"` value substitution must not accidentally alter surrounding JSON keys or introduce a trailing-comma/quote error.
- **EC-3**: If a run-phase re-verification (plan.md §C.3/§C.4) finds that the H1-missing file list or the feedback workflow content has drifted since plan-phase (e.g., a parallel session already fixed some files), the implementing agent must re-derive the actual current state rather than blindly applying the plan-phase snapshot, and note the delta in `progress.md`.
- **EC-4**: The Check-4 glossary-term finding (`ja/best-practices.md`) must NOT be touched even though it appears in the same `scripts/docs-i18n-check.sh` output as the in-scope Check-3 findings — mixing an out-of-scope fix into this SPEC's commits would blur its scope boundary.

## Quality Gate Criteria

- `scripts/docs-i18n-check.sh` (default strict mode) exits 1 with **exactly 1** residual error (the out-of-scope Check-4 finding) — not 0 (that would incorrectly imply the out-of-scope defect was also fixed) and not >1 (that would indicate a regression or incomplete REQ-DPC-001 fix).
- `hugo --minify` (in `docs-site/`) completes with zero warnings.
- No `goos` string remains in any of the 3 touched files across all 4 locales.
- No `rc2`/`rc3`/`rc4` string remains in `docs-site/hugo.toml` or any locale's `_index.md`.
- `git status --porcelain` scoped to `docs-site/` after all milestones shows only files enumerated in plan.md §F (no drive-by edits to `.claude/skills/moai/workflows/feedback.md`, `.moai/config/sections/feedback.yaml`, or any `internal/` path).

## Definition of Done

- [ ] All 6 REQs (REQ-DPC-001 through REQ-DPC-006) have PASS evidence for their AC matrix rows above.
- [ ] `scripts/docs-i18n-check.sh` shows exactly 1 residual (out-of-scope) error, down from the plan-phase baseline of 65.
- [ ] `hugo --minify` builds cleanly with the `v3.0.0-rc5` version string.
- [ ] No maintainer-path leak (`goos`) remains in the 3 touched files across 4 locales.
- [ ] `docs-site/static/robots.txt` and `docs-site/hugo.toml` reflect the corrected domain and version.
- [ ] ko `moai-feedback.md` accurately describes the 4 `/moai feedback` enhancements with no invented behavior and no residual "sync-auditor" misattribution.
- [ ] ko `moai-feedback.md`'s "자동 수집되는 정보" table lists only the diagnostics actually collected (no "Claude Code 버전" or "현재 SPEC" row).
- [ ] No Go code, template source, or feedback workflow/config schema file was modified.
- [ ] D5 (non-ko untranslated backlog) and the Check-4 glossary defect remain untouched (correctly out of scope).
- [ ] `progress.md` §E.2/§E.3 populated by manager-develop at run-phase completion; §E.4 by manager-docs at sync-phase completion.
