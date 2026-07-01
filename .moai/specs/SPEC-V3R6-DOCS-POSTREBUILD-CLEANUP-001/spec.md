---
id: SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001
title: "Docs-Site Post-Rebuild Cleanup — Deferred Defects (D2/D6/D7) + New Findings"
version: "0.1.2"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "docs-site/content, docs-site/static, docs-site/hugo.toml"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, cleanup, follow-up"
era: V3R6
tier: M
depends_on: [SPEC-V3R6-DOCS-V3-REBUILD-001]
related_specs: [SPEC-INVOCATION-MODEL-001]
---

# SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001 — Docs-Site Post-Rebuild Cleanup

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-07-02 | 0.1.0 | manager-spec | Initial plan-phase authoring. Follow-up to SPEC-V3R6-DOCS-V3-REBUILD-001's sync-auditor Tier-L review (4 MUST-FIX: D1-D4; D1/D3/D4 fixed in post-close commit 8f89a42b7). Picks up the 2 user-deferred defects in scope for this SPEC (D2, D6, D7 — D5 excluded, see Exclusions) plus 2 new findings independently verified in this session (stale version label, stale maintainer-path leak in `inventory.md` that the original sync-auditor review did not catch). |
| 2026-07-02 | 0.1.1 | manager-spec | Iteration-2 revision, applied after an independent plan-auditor iteration-1 FAIL (score 0.45, must-pass firewall failure on informal AC form). Ground Truth facts unchanged (re-verified accurate, not re-investigated) — this iteration is a format + scope-completeness pass only: (1) acceptance.md §D AC Matrix rewritten from informal command-assertion form into GEARS "shall" sentences for all 18 lettered entries; (2) REQ-DPC-006 scope extended to also correct the ko `moai-feedback.md` "자동 수집되는 정보" table's unsupported "Claude Code 버전" / "현재 SPEC" rows, with a corresponding new AC; (3) REQ-DPC-006's `Where` clause reworded to plain Ubiquitous form (page-scoping is not a genuine capability gate); (4) REQ-DPC-005's Unwanted clause reworded to the canonical `shall not` form; (5) an explicit `## Scope (WHAT)` subsection added under Overview to close the plan-auditor's Completeness gap; (6) plan.md's M3 milestone gained a concrete 16-canonical-page path enumeration (re-run of the same Ground-Truth-cited `docs-i18n-check.sh` command, not a new investigation). |
| 2026-07-02 | 0.1.2 | manager-spec | Iteration-3 revision. Applied D1 (SHOULD-FIX, still-open) + D3 (MINOR) after an independent orchestrator-spawned plan-auditor iteration-1 flagged D1 on the 0.1.0 base and the orchestrator adopted this SPEC; Ground Truth re-verified this session (hugo.toml L55 `v3.0.0-rc4` + L56 `2026-06-23`; latest tag `v3.0.0-rc5` dated 2026-07-01; L54 SSOT comment `# SSOT — Version 갱신은 이 두 줄만 동시에 수정` pairs the two lines). D1: REQ-DPC-005 extended to also bump `hugo.toml` `releaseDate` 2026-06-23 → 2026-07-01 (the rc5 tag date), honoring the L54 two-line SSOT convention; plan.md M1 gained step M1.2b + M5 gained M5.3b; acceptance.md gained AC-DPC-005d. D3: NFR-DPC-004 (ko-first, ko-only scope) was untraced in 0.1.1 — added AC-DPC-006f (new `feedback.repository`/`modu-ai/moai-adk` content grep = 0 in en/ja/zh, ground-truthed to 0 across all 4 locales this session). D2 (already fixed in 0.1.1) and Tier M (D4 accepted debt) untouched. |

## Overview (WHY)

SPEC-V3R6-DOCS-V3-REBUILD-001 (Tier L, `status: completed`, sync commit `44073782b`) rebuilt `docs-site/` to v3.0.0-rc4 accuracy. Its independent sync-auditor Tier-L review found 4 MUST-FIX content defects (D1-D4). D1/D3/D4 were fixed immediately in a post-close commit (`8f89a42b7`). D2, D5, D6, D7 were explicitly deferred by user decision to a follow-up SPEC. This SPEC is that follow-up.

This SPEC's scope is **D2 + D6 + D7**, plus **2 new findings** the orchestrator independently verified in this session while re-running the docs-i18n check and grepping for maintainer-path leaks:

- A **stale version label**: `docs-site/hugo.toml` still declares `v3.0.0-rc4` while the project has since tagged and released `v3.0.0-rc5` (2026-07-01).
- A **second maintainer-path leak** the original sync-auditor review did not catch: `getting-started/inventory.md` (all 4 locales) still shows the literal path `/Users/goos/MoAI/moai-adk-go` in an example output block — the same class of defect as D6 (`worktree/examples.md` + `worktree/faq.md`), just in a different file the original review missed.

**D5** (the non-Korean-locale untranslated-content backlog — `hooks-reference.md` ja/zh + 8 other pages entirely in Korean across en/ja/zh) is explicitly **OUT OF SCOPE**. It is tracked in memory file `project_docsite_untranslated_backlog.md` and the user has decided to defer it further; it is a much larger prose-translation effort than the mechanical/small-prose fixes in this SPEC.

## Scope (WHAT)

This SPEC touches exactly 3 directories (`docs-site/content/`, `docs-site/static/`, `docs-site/hugo.toml`) across 6 requirements: (1) restore a missing body-level H1 on 16 canonical pages × 4 locales, (2) remove a maintainer-path leak from 2 worktree pages × 4 locales, (3) remove a maintainer-path leak from `inventory.md` × 4 locales, (4) correct `robots.txt`'s sitemap domain, (5) bump the stale `hugo.toml` version string, and (6) correct + extend the ko-only `moai-feedback.md` page's content accuracy (including its "자동 수집되는 정보" diagnostic table). No Go code, template source, or feedback workflow/config schema file is touched — see Exclusions below for the complete out-of-scope enumeration.

## Ground Truth (observed evidence — verified in this session, not re-derived from prior claims)

| Fact | Observed value | Verification command (run 2026-07-02) |
|------|-----------------|------------------------------------------|
| Check-3 (H1 missing) file count | Exactly **16 pages × 4 locales = 64 files** (16 confirmed, not re-derived from the task prompt's hedge) | `DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh` |
| Check-3 per-section breakdown | `claude-code/foundations/` 6 · `claude-code/agentic/` 5 (agent-teams, best-practices, goal, scheduled-tasks, sub-agents) · `claude-code/extensibility/` 2 (hooks, skills) · `claude-code/context-memory/context-window.md` 1 · `advanced/decision-memory.md` 1 · `workflow-commands/moai-harness.md` 1 = 16 | Same run — script emits one `ERROR no H1 heading in <path>` line per file |
| Unrelated Check-4 finding (out of scope) | `glossary term 'Anthropic' missing in ja/claude-code/agentic/best-practices.md` | Same run — Check 4, distinct from Check 3 and unrelated to D2 |
| Maintainer path leak — worktree pages (D6) | `worktree/examples.md` 4 occurrences/locale (L26-27/47-48/142-143/433-434-ish, exact line varies by locale); `worktree/faq.md` 1 occurrence/locale (~L391-392) | `grep -n goos docs-site/content/{ko,en,ja,zh}/worktree/{examples,faq}.md` |
| Maintainer path leak — inventory.md (new finding) | 2 occurrences/locale (`Project Root: /Users/goos/MoAI/moai-adk-go` + `"project_root": "/Users/goos/MoAI/moai-adk-go"` inside a JSON example) | `grep -n goos docs-site/content/{ko,en,ja,zh}/getting-started/inventory.md` |
| robots.txt wrong domain (D7) | `Sitemap: https://cowork.mo.ai.kr/sitemap.xml` (a different, unrelated MoAI product site) | `cat docs-site/static/robots.txt` |
| hugo.toml stale version (new finding) | `docs-site/hugo.toml` L55 `version = "v3.0.0-rc4"`; latest git tag is `v3.0.0-rc5` (`.moai/config/sections/system.yaml` confirms `version: v3.0.0-rc5`, tagged 2026-07-01) | `grep -n 'version =' docs-site/hugo.toml`; `git tag --sort=-creatordate \| head -3`; `grep -n version .moai/config/sections/system.yaml` |
| No other stale rc string | 0 matches for `rc2`/`rc3`/`rc4` in any locale's homepage `_index.md` (the prior post-close commit already corrected the ko homepage's self-contradictory rc2 label) | `grep -rn 'rc4\|rc2\|rc3' docs-site/content/*/_index.md docs-site/hugo.toml` |
| Actual `/moai feedback` enhancement source (corrected sourcing) | `.claude/skills/moai/workflows/feedback.md` (orchestrator-direct, no subagent) + `internal/config/feedback_accessors.go` (`FeedbackRepository()` resolver) + `internal/template/templates/.moai/config/sections/feedback.yaml` (`feedback.repository`, default `modu-ai/moai-adk`). The `internal/loop/{feedback,feedback_channel,go_feedback}.go` files named in the originating task are the **unrelated** Ralph Engine loop-feedback system (test/lint/coverage/LSP-diagnostic tracking for `/moai loop`) — a naming collision, not the source for this SPEC. | Read of all 4 files above + `SPEC-INVOCATION-MODEL-001/spec.md` §C |
| ko `moai-feedback.md` pre-existing inaccuracy | The page's agent-chain diagram (L121-151) attributes GitHub issue creation to a "sync-auditor 에이전트", but the actual workflow (`feedback.md` L89, L176-177) is orchestrator-direct with **no subagent spawn** — issue creation has no retained-agent owner per `archived-agent-rejection.md` §C | Read of `docs-site/content/ko/utility-commands/moai-feedback.md` |

## Requirements (GEARS)

### H1 heading restoration (D2)

- **REQ-DPC-001** (Event-driven): **When** `scripts/docs-i18n-check.sh` Check 3 flags a content page as missing a body-level H1, the docs-site content **shall** carry an explicit `# <localized title>` heading as the first content line after YAML frontmatter, where the heading text matches the frontmatter `title:` field's language and content, across all 4 locales, for each of the 16 canonical pages identified by the fresh Check-3 run recorded in Ground Truth above (NOT the prompt's original unverified per-section hedge).

### Maintainer-path-leak fixes (D6 + new finding)

- **REQ-DPC-002** (Ubiquitous): The published `worktree/examples.md` and `worktree/faq.md` pages, across all 4 locales, **shall** use a generic placeholder path (`/path/to/your-project`) in place of the literal maintainer development path `/Users/goos/MoAI/moai-project`.
- **REQ-DPC-003** (Ubiquitous): The published `getting-started/inventory.md` page, across all 4 locales, **shall** use the same generic placeholder path (`/path/to/your-project`) in place of the literal maintainer development path `/Users/goos/MoAI/moai-adk-go`.

### robots.txt domain correction (D7)

- **REQ-DPC-004** (Ubiquitous): `docs-site/static/robots.txt` **shall** reference the sitemap at the docs site's own canonical domain (`https://adk.mo.ai.kr/sitemap.xml`) and **shall not** reference the unrelated `cowork.mo.ai.kr` domain.

### Version label bump (new finding)

- **REQ-DPC-005** (Ubiquitous): `docs-site/hugo.toml` `params.version` **shall** reflect the current release tag (`v3.0.0-rc5`), and `docs-site/hugo.toml` `params.releaseDate` **shall** reflect that tag's release date (`2026-07-01`) — honoring the hugo.toml inline SSOT convention (L54 comment `# SSOT — Version 갱신은 이 두 줄만 동시에 수정`) that mandates the `version` and `releaseDate` lines be updated together, never one without the other; and `docs-site/hugo.toml` and any locale's homepage `_index.md` **shall not** reference any stale `rc2`/`rc3`/`rc4` version string.

### ko `/moai feedback` content accuracy (new finding, ko-first)

- **REQ-DPC-006** (Ubiquitous): The ko-locale `utility-commands/moai-feedback.md` page **shall** accurately describe the 4 workflow enhancements shipped by SPEC-INVOCATION-MODEL-001 (guaranteed diagnostic collection — MoAI version + OS — plus best-effort diagnostic collection — Go toolchain version + orchestrator-passed error context; a duplicate-issue candidate-report step; a `gh`-failure graceful fallback that saves a local draft; and a configurable target repository via `feedback.repository`, default `modu-ai/moai-adk`), grounded in the actual current implementation of `.claude/skills/moai/workflows/feedback.md` + `internal/config/feedback_accessors.go` + `internal/template/templates/.moai/config/sections/feedback.yaml`; **shall** correct the page's pre-existing agent-chain diagram inaccuracy (which currently attributes GitHub issue creation to a "sync-auditor 에이전트" when the actual workflow is orchestrator-direct with no subagent spawn); **and shall** correct the page's pre-existing "자동 수집되는 정보" table (current L72-78) so it lists only the diagnostics the workflow actually collects — the guaranteed pair (MoAI version, OS) and the best-effort pair (Go toolchain version, orchestrator-passed error context) — removing the two unsupported rows ("Claude Code 버전" and "현재 SPEC") that no current source (`.claude/skills/moai/workflows/feedback.md`, `internal/config/feedback_accessors.go`, `internal/template/templates/.moai/config/sections/feedback.yaml`) substantiates.

### Non-functional constraints

- **NFR-DPC-001** (4-locale parity): REQ-DPC-001's H1 additions **shall** land in all 4 locales within the same milestone; the 99-file-per-locale baseline (verified by SPEC-V3R6-DOCS-V3-REBUILD-001 M3) **shall not** change (no file added or removed, only H1 lines inserted into existing files).
- **NFR-DPC-002** (mechanical symmetry): REQ-DPC-002 and REQ-DPC-003's path substitutions **shall** be identical mechanical replacements across all 4 locales — NOT prose translation or locale-specific rewording beyond the substituted path token.
- **NFR-DPC-003** (no unobserved-claim / verification-claim-integrity): REQ-DPC-006's content **shall** be verified against the actual Go source and workflow-skill files cited above before writing; the page **shall not** describe any enhancement behavior absent from the read source.
- **NFR-DPC-004** (ko-first sequencing): REQ-DPC-006 **shall** be completed for the ko locale as the first content-authoring milestone in this SPEC's run-phase; en/ja/zh translations of the new content are explicitly out of scope for this SPEC (see Exclusions).

## Exclusions

### Out of Scope — Non-Korean untranslated content backlog (D5)

- The pre-existing non-ko untranslated-content backlog (`hooks-reference.md` ja/zh + 8 other pages entirely in Korean across en/ja/zh), tracked in memory file `project_docsite_untranslated_backlog.md`, is explicitly deferred per user decision. This SPEC does not touch it.

### Out of Scope — en/ja/zh prose updates to moai-feedback.md

- REQ-DPC-006 covers the ko locale only. Translating the new "피드백 설정" subsection (or equivalent) into en/ja/zh is deferred to a follow-up SPEC or a later milestone outside this one.

### Out of Scope — Check-4 glossary-term pre-existing defect

- The `scripts/docs-i18n-check.sh` Check 4 finding (`glossary term 'Anthropic' missing in ja/claude-code/agentic/best-practices.md`) is a pre-existing, unrelated defect discovered incidentally during this SPEC's plan-phase verification run. It is **NOT** fixed by this SPEC. `scripts/docs-i18n-check.sh` (default strict mode, exit-on-error) **will still exit 1** after this SPEC's changes land, due solely to this residual Check-4 finding — this is expected and is not a regression introduced by this SPEC. Tracked for a future follow-up.

### Out of Scope — Go/template code changes

- No changes to `internal/`, `cmd/`, or `internal/template/templates/`. This SPEC touches ONLY `docs-site/content/`, `docs-site/static/`, and `docs-site/hugo.toml`.

### Out of Scope — feedback workflow skill / config schema changes

- `.claude/skills/moai/workflows/feedback.md` and `.moai/config/sections/feedback.yaml` (and their template mirrors) are **NOT** modified by this SPEC. REQ-DPC-006 only updates the docs PAGE that describes already-existing behavior; the config already exists (per SPEC-INVOCATION-MODEL-001), only its docs description is being added.

### Out of Scope — hugo build/deploy verification beyond the version string

- Full `hugo --minify` build verification, Vercel preview checks, and the §17.6 build/deploy checklist are not re-run in full by this SPEC (SPEC-V3R6-DOCS-V3-REBUILD-001 already validated the build pipeline). This SPEC verifies only that `hugo --minify` still completes cleanly and emits the corrected version string after these targeted edits.

## Dependencies and References

- SPEC-V3R6-DOCS-V3-REBUILD-001 — the source SPEC; its sync-auditor Tier-L review is the origin of D1-D7. D1/D3/D4 fixed in its post-close commit `8f89a42b7`; D2/D6/D7 are picked up here; D5 remains deferred.
- SPEC-INVOCATION-MODEL-001 — source of the 4 `/moai feedback` workflow enhancements described by REQ-DPC-006 (`.claude/skills/moai/workflows/feedback.md`, `internal/config/feedback_accessors.go`, `internal/template/templates/.moai/config/sections/feedback.yaml`).
- `project_docsite_untranslated_backlog.md` (auto-memory) — tracks D5, explicitly out of scope here.
- `.moai/docs/docs-site-i18n-rules.md` §17 — 4-locale parity, URL blacklist, build/deploy checklist (SSOT for the docs-site NFRs this SPEC's edits must remain consistent with).

## Acceptance Criteria

See `acceptance.md` for per-requirement testable/verifiable acceptance criteria and Given-When-Then scenarios.
