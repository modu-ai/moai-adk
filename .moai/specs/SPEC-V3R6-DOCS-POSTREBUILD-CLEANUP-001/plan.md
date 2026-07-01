# Implementation Plan — SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001

## §A Context

Follow-up cleanup batch to SPEC-V3R6-DOCS-V3-REBUILD-001 (Tier L, completed, sync commit `44073782b`, post-close fix `8f89a42b7`). Picks up 2 user-deferred sync-auditor findings (D2, D6, D7 — D5 excluded) plus 2 new findings the orchestrator independently verified this session (stale `hugo.toml` version, a second maintainer-path leak in `inventory.md`). Docs-content-only — no Go code, no `internal/template/templates/` changes.

**Tier: M** (judgment call, documented). Raw file count is high (~79 files touched: 64 for REQ-DPC-001, 12 for REQ-DPC-002/003, 1 each for REQ-DPC-004/005, 1 for REQ-DPC-006), which nominally exceeds the Tier-L ">15 files" guidance band. However, per-file complexity is trivial for 78 of 79 files (single-line H1 insertion or a single-token path substitution, mechanically identical across locales) — no architecture decisions, no external research dependency, no cross-cutting IA change. Only REQ-DPC-006 (1 file, ko-only) requires genuine prose authoring grounded in source-code reading. The LOC thresholds in the Tier table are explicitly guidance, not enforcement (`spec-workflow.md` § SPEC Complexity Tier); given the low-risk, low-complexity, docs-only profile, Tier M (3-file artifact set: spec.md + plan.md + acceptance.md, no design.md/research.md) is the appropriate classification, consistent with the task's own framing ("Tier L... seems unlikely given the mechanical nature of 4 of 5 requirements").

Hybrid Trunk 1-person OSS policy applies (CLAUDE.local.md §23): all tiers push directly to `main`; no PR required.

## §B Known Issues (drift inventory feeding this cleanup batch)

| Surface | Drift | Target REQ |
|---------|-------|-----------|
| 16 pages × 4 locales (see spec.md Ground Truth breakdown) | Missing body-level H1 (frontmatter `title:` renders as H1 via hugo-geekdoc, but raw markdown source lacks one) | REQ-DPC-001 |
| `worktree/examples.md` (×4 locale) | 4 occurrences/locale of literal `/Users/goos/MoAI/moai-project` | REQ-DPC-002 |
| `worktree/faq.md` (×4 locale) | 1 occurrence/locale of literal `/Users/goos/MoAI/moai-project` | REQ-DPC-002 |
| `getting-started/inventory.md` (×4 locale) | 2 occurrences/locale of literal `/Users/goos/MoAI/moai-adk-go` (missed by the original sync-auditor review) | REQ-DPC-003 |
| `docs-site/static/robots.txt` | `Sitemap: https://cowork.mo.ai.kr/sitemap.xml` (wrong, unrelated domain) | REQ-DPC-004 |
| `docs-site/hugo.toml` L55-56 | `version = "v3.0.0-rc4"` + `releaseDate = "2026-06-23"`; current tag `v3.0.0-rc5` dated 2026-07-01 (L54 SSOT comment mandates both lines bump together) | REQ-DPC-005 |
| `content/ko/utility-commands/moai-feedback.md` | Missing description of 4 SPEC-INVOCATION-MODEL-001 enhancements; agent-chain diagram misattributes issue creation to "sync-auditor 에이전트" (actual: orchestrator-direct, no subagent); "자동 수집되는 정보" table (L72-78) lists 2 unsupported diagnostics ("Claude Code 버전", "현재 SPEC") not collected by the actual workflow | REQ-DPC-006 |

## §C Pre-flight (run-phase entry preconditions)

1. `git status --porcelain docs-site/` clean (no uncommitted drift from a parallel session).
2. `git rev-parse HEAD` matches the plan-phase baseline (no intervening docs-site commit landed since plan-phase Ground Truth was captured).
3. Re-run `DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh` at the start of REQ-DPC-001's milestone to reconfirm the 64-file H1 list is unchanged since plan-phase.
4. Re-read `.claude/skills/moai/workflows/feedback.md` + the 3 Go/config sources at the start of REQ-DPC-006's milestone (in case the implementation drifted between plan-phase and run-phase) — do not rely solely on the plan-phase Ground Truth snapshot for prose content.

## §D Constraints

- Docs-content-only: no changes to `internal/`, `cmd/`, or `internal/template/templates/` (this is a `docs-site/`-scoped SPEC).
- No changes to `.claude/skills/moai/workflows/feedback.md` or `.moai/config/sections/feedback.yaml` (REQ-DPC-006 only updates the docs page describing existing behavior).
- 4-locale parity mandatory for REQ-DPC-001/002/003 within the same milestone (NFR-DPC-001/002); REQ-DPC-006 is ko-only by explicit scope (NFR-DPC-004).
- No invented behavior in REQ-DPC-006 — every claim about the 4 `/moai feedback` enhancements must be traceable to the actual read source (NFR-DPC-003).
- Do not touch the pre-existing Check-4 glossary-term defect (`ja/claude-code/agentic/best-practices.md`) — out of scope, unrelated to this SPEC's REQs.
- Do not touch D5 (non-ko untranslated backlog) — explicitly deferred.
- hugo-geekdoc theme/CSS/layout untouched — content-only edits.

## §E Self-Verification

Plan-phase self-verification is recorded in `progress.md` §E.1. Run-phase and sync-phase evidence are populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4), per the Status Transition Ownership Matrix.

## §F Milestones

Priority-based ordering; no time estimates. Sequenced from lowest-risk/fastest to highest-effort, with verification as the closing milestone.

### M1 · Mechanical single-line fixes (Priority: High — fastest, lowest risk, do first)

- **M1.1** Fix `docs-site/static/robots.txt`: change `Sitemap: https://cowork.mo.ai.kr/sitemap.xml` → `Sitemap: https://adk.mo.ai.kr/sitemap.xml` (REQ-DPC-004).
- **M1.2** Bump `docs-site/hugo.toml` L55 `params.version` from `v3.0.0-rc4` to `v3.0.0-rc5` (REQ-DPC-005).
- **M1.2b** Bump `docs-site/hugo.toml` L56 `params.releaseDate` from `2026-06-23` to `2026-07-01` (the v3.0.0-rc5 tag date) — same SSOT unit as M1.2 per the L54 comment (`# SSOT — Version 갱신은 이 두 줄만 동시에 수정`); the `version` and `releaseDate` lines MUST be bumped together, never one without the other (REQ-DPC-005).
- **M1.3** Re-run `grep -n 'rc4\|rc2\|rc3' docs-site/hugo.toml docs-site/content/*/_index.md` to confirm no other stale rc references remain (the prior post-close fix already corrected the ko homepage; this step re-verifies, does not reintroduce drift) (REQ-DPC-005).

### M2 · Maintainer-path-leak fixes (Priority: High — mechanical, locale-symmetric)

- **M2.1** Replace `/Users/goos/MoAI/moai-project` with `/path/to/your-project` in `worktree/examples.md` × 4 locales (4 occurrences/locale) (REQ-DPC-002).
- **M2.2** Replace `/Users/goos/MoAI/moai-project` with `/path/to/your-project` in `worktree/faq.md` × 4 locales (1 occurrence/locale) (REQ-DPC-002).
- **M2.3** Replace `/Users/goos/MoAI/moai-adk-go` with `/path/to/your-project` in `getting-started/inventory.md` × 4 locales (2 occurrences/locale, including inside a JSON example block — preserve JSON validity) (REQ-DPC-003).
- **M2.4** `grep -rn goos docs-site/content/{ko,en,ja,zh}/` returns zero matches in these 3 files across all 4 locales.

### M3 · H1 heading restoration (Priority: Medium — highest file count, mechanical per-file edit)

- **M3.1** For each of the 16 canonical pages × 4 locales (64 files total), insert `# <localized title>` as the first content line after YAML frontmatter, where the title text matches the frontmatter `title:` field's language/content for that locale (REQ-DPC-001). The 16 canonical pages (locale-relative path, identical across ko/en/ja/zh) — re-derived at plan-phase from the same `DOCS_I18N_STRICT=0 scripts/docs-i18n-check.sh` command already cited in spec.md Ground Truth, not a new investigation:
  1. `advanced/decision-memory.md`
  2. `claude-code/agentic/agent-teams.md`
  3. `claude-code/agentic/best-practices.md`
  4. `claude-code/agentic/goal.md`
  5. `claude-code/agentic/scheduled-tasks.md`
  6. `claude-code/agentic/sub-agents.md`
  7. `claude-code/context-memory/context-window.md`
  8. `claude-code/extensibility/hooks.md`
  9. `claude-code/extensibility/skills.md`
  10. `claude-code/foundations/claude-directory.md`
  11. `claude-code/foundations/commands.md`
  12. `claude-code/foundations/features-overview.md`
  13. `claude-code/foundations/how-claude-code-works.md`
  14. `claude-code/foundations/interactive-mode.md`
  15. `claude-code/foundations/tools-reference.md`
  16. `workflow-commands/moai-harness.md`

  This static list gives run-phase a checklist to check off against; per acceptance.md EC-3, if a run-phase re-verification finds the list has drifted (e.g., a parallel session already fixed some files), re-derive the actual current state from the script rather than blindly applying this snapshot.
- **M3.2** Re-run `scripts/docs-i18n-check.sh` Check 3 in isolation (or `DOCS_I18N_STRICT=0` full run) to confirm 0 remaining H1-missing errors (the 1 residual Check-4 finding is expected and out of scope).

### M4 · ko `/moai feedback` content update (Priority: Medium — genuine prose, ko-only)

- **M4.1** Re-read `.claude/skills/moai/workflows/feedback.md`, `internal/config/feedback_accessors.go`, `internal/template/templates/.moai/config/sections/feedback.yaml` to reconfirm the plan-phase Ground Truth is still current (implementation may have drifted since plan-phase).
- **M4.2** Add a new ko subsection (e.g., "피드백 설정") to `content/ko/utility-commands/moai-feedback.md` describing the 4 enhancements: guaranteed diagnostics (MoAI version, OS) + best-effort diagnostics (Go version, orchestrator-passed error context); duplicate-issue candidate-report step; `gh`-failure graceful fallback with local draft save; configurable `feedback.repository` target (default `modu-ai/moai-adk`).
- **M4.3** Correct the existing "자동 수집되는 정보" table (current L72-78 region) to list only the diagnostics the workflow actually collects — replace the unsupported "Claude Code 버전" row with the actual best-effort "Go 툴체인 버전" diagnostic, and remove the unsupported "현재 SPEC" row entirely (no current source collects it); keep "MoAI-ADK 버전", "OS 정보", and "오류 로그" (marking the latter as best-effort/orchestrator-passed, not guaranteed) (REQ-DPC-006).
- **M4.4** Correct the existing agent-chain diagram/table (L121-151 region) to remove the "sync-auditor 에이전트가 GitHub 이슈 생성" misattribution and reflect the actual orchestrator-direct, no-subagent execution model.
- **M4.5** Verify no invented behavior: cross-check every claim added under M4.2/M4.3 against the actual read source from M4.1.

### M5 · Full verification sweep (Priority: High — closing gate)

- **M5.1** `scripts/docs-i18n-check.sh` (default strict mode): expect exit 1 with exactly 1 residual error (the out-of-scope Check-4 glossary finding) — 0 Check-3 (H1) errors.
- **M5.2** `grep -rn goos docs-site/content/{ko,en,ja,zh}/` returns zero matches.
- **M5.3** `grep -n 'rc4\|rc2\|rc3' docs-site/hugo.toml` returns zero matches (only rc5 present).
- **M5.3b** `grep -c 'releaseDate = "2026-07-01"' docs-site/hugo.toml` returns `1` AND `grep -c '2026-06-23' docs-site/hugo.toml` returns `0` — confirming the SSOT-paired releaseDate bump landed alongside the version bump (REQ-DPC-005).
- **M5.4** `cat docs-site/static/robots.txt` confirms the `adk.mo.ai.kr` sitemap domain.
- **M5.5** `hugo --minify` (in `docs-site/`) completes with zero warnings and emits the `v3.0.0-rc5` version string.
- **M5.6** `git status --porcelain` scoped to `docs-site/` shows only the expected file set touched by M1-M4 (no unrelated drift).
