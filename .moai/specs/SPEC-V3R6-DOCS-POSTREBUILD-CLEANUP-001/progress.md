# Progress — SPEC-V3R6-DOCS-POSTREBUILD-CLEANUP-001

Lifecycle progress ledger. §E.1 is populated at plan-phase (manager-spec). §E.2/§E.3 are populated at run-phase (manager-develop); §E.4 at sync-phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

- **Phase**: plan (complete)
- **Tier**: M (judgment call — see plan.md §A for the file-count-vs-complexity rationale; raw file count ~79 nominally suggests Tier L, but per-file complexity is trivial/mechanical for 78 of 79 files).
- **Artifacts produced**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (4 files — standard Tier M set, no design.md/research.md).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | DOCS ✓ | POSTREBUILD ✓ | CLEANUP ✓ | 001 ✓ → PASS` (regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Frontmatter**: 12 canonical fields present; `status: draft`; `created`/`updated` (not `_at`); `tags` comma-separated string; `era: V3R6`; `tier: M`; `depends_on`/`related_specs` optional fields present.
- **Requirement count**: 6 REQ-DPC (001-006) + 4 NFR-DPC (001-004).
- **Ground truth basis** (observed 2026-07-02, live repo state): `scripts/docs-i18n-check.sh` fresh run confirmed 65 total errors (64 Check-3 H1-missing across 16 pages × 4 locales + 1 unrelated Check-4 glossary finding); `grep -n goos` confirmed 28 occurrences across 3 files × 4 locales (worktree/examples.md 4/locale, worktree/faq.md 1/locale, getting-started/inventory.md 2/locale); `robots.txt` confirmed wrong `cowork.mo.ai.kr` domain; `hugo.toml` confirmed stale `v3.0.0-rc4` vs current tag `v3.0.0-rc5` (dated 2026-07-01); no other stale rc references found.
- **Source-correction finding**: the task's originally-suggested REQ-DPC-006 source files (`internal/loop/{feedback,feedback_channel,go_feedback}.go`) are the unrelated Ralph Engine loop-feedback system (test/lint/coverage/LSP-diagnostic tracking for `/moai loop`), NOT the `/moai feedback` GitHub-issue workflow. The actual source is `.claude/skills/moai/workflows/feedback.md` + `internal/config/feedback_accessors.go` + `internal/template/templates/.moai/config/sections/feedback.yaml`, cross-verified against `SPEC-INVOCATION-MODEL-001/spec.md`. Also found: the current ko `moai-feedback.md` page already misattributes GitHub issue creation to a "sync-auditor 에이전트" (the actual workflow is orchestrator-direct, no subagent) — folded into REQ-DPC-006's scope.
- **Out of Scope**: present (D5 non-ko backlog, en/ja/zh feedback-page translation, Check-4 glossary defect, Go/template code changes, feedback workflow/config schema changes, full hugo build/deploy re-verification).
- **Plan-phase gaps (residual)**: (1) exact H1 text for each of the 64 files is not enumerated in the SPEC body — left to run-phase per-file frontmatter-title lookup (WHAT not HOW); (2) the ko `moai-feedback.md` new subsection's exact heading name ("피드백 설정" is a suggestion, not mandated) is left to the implementing agent's judgment.
- **Next phase**: run (M1 → M5 per plan.md §F). Requires Implementation Kickoff Approval (plan-to-implement HUMAN GATE) before run-phase entry.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
