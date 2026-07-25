---
id: SPEC-ASTGREP-EDIT-001
title: "Progress — ast-grep wiring repair and moai ast-edit"
version: "0.1.0"
status: in-progress
created: 2026-07-25
updated: 2026-07-26
author: GOOS
priority: P1
phase: "v3.1.0 target"
module: "internal/cli, internal/astgrep, internal/hook/security"
lifecycle: spec-anchored
tags: "ast-grep, ast-edit, progress"
---

# Progress — SPEC-ASTGREP-EDIT-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-25
plan_commit_sha: "95faf94d0"

Plan-phase artifacts (spec.md, plan.md, acceptance.md) were authored in commit
`95faf94d0 docs(SPEC-ASTGREP-EDIT-001): plan-phase artifacts (Tier L, tdd)`.
Note: this SPEC shipped with 3 artifacts (spec.md + plan.md + acceptance.md)
despite Tier L classification (which nominally expects 5: + design.md +
research.md). This is a plan-phase gap, recorded here for visibility but out
of scope for this sync (per the sync-phase mandate — SPEC body content is
frozen and design.md/research.md are not created retroactively).

## §E.2 Run-phase Evidence

**Provenance note (user-approved recovery):** `progress.md` was absent at the
end of run-phase, which caused era misclassification (V2.x via heuristic H-1).
The user explicitly authorized manager-docs to author the full §E.2/§E.3
evidence retroactively in this single sync commit, reconstructing evidence
from the 6 landed run-phase commits plus the sync-phase observed baseline
(this run, this tree) — this is a deliberate, user-approved gap recovery, not
an undisclosed authorship-boundary crossing.

Six run-phase commits landed on `feat/SPEC-ASTGREP-EDIT-001`:

| Milestone | Requirement | Commit | Evidence |
|---|---|---|---|
| M1 | REQ-AGE-005 — `FindRulesConfig` resolves shipped ruleset | `b0d6d4215 fix(SPEC-ASTGREP-EDIT-001): resolve shipped ast-grep ruleset in FindRulesConfig` | `internal/hook/security/rules.go` `searchPaths` now includes `.moai/config/astgrep-rules/sgconfig.{yml,yaml}`; retired `moai-tool-ast-grep` skill paths removed (verified: `grep -c "moai-tool-ast-grep" internal/hook/security/rules.go` → 0) |
| M2 | REQ-AGE-001 / REQ-AGE-002 — `moai ast-edit` command skeleton, dry-run safe default | `b3c857dfe feat(SPEC-ASTGREP-EDIT-001): add moai ast-edit and fix the rewrite no-op` | `internal/cli/astedit.go` created (`NewAstEditCmd`), registered in `internal/cli/root.go` alongside `ast-grep`; `--dry` flag defaults to safe preview |
| M3 | REQ-AGE-003 — pattern-mode rewrite + `--update-all` no-op fix | same commit `b3c857dfe` | `internal/astgrep/analyzer.go` `PatternReplace` now passes `--update-all` to `sg run --rewrite` (previously the rewrite printed a diff and never touched disk); fixture-rewrite assertion added, not mock-arg-only |
| M4 | REQ-AGE-004 — rule-mode rewrite, `--rule` filter | `9bfb039b7 feat(SPEC-ASTGREP-EDIT-001): rule-mode tests, fix: assessment, dead-reference cleanup` | `applyRuleEdits` in `internal/cli/astedit.go`; tests `TestAstEditCmd_RuleModeAppliesFix` / `TestAstEditCmd_RuleFilterNarrowsToOneRule` |
| M5 | REQ-AGE-006 — shipped rule `fix:` assessment | same commit `9bfb039b7` | Disposition table recorded in `acceptance.md` § REQ-AGE-006 — all 4 shipped rule files remain detection-only (0 `fix:` fields added), each with an evidence-backed verdict |
| M6 | REQ-AGE-007 — dead skill reference cleanup | same commit `9bfb039b7` | `moai-foundation-cc/SKILL.md`, `moai-workflow-ddd/SKILL.md`, `moai-workflow-loop/SKILL.md` + template mirrors updated to reference `moai ast-grep` / `moai ast-edit` instead of the retired `moai-tool-ast-grep` skill |
| M7 | REQ-AGE-008 — template parity, docs, verification | `d24c4ff14 docs(SPEC-ASTGREP-EDIT-001): document ast-grep and ast-edit in 4 locales` | `docs-site/content/{en,ja,ko,zh}/cli-reference/{ast-grep.md,_meta.yaml}` added; `make build` run (catalog.yaml regenerated); template neutrality/namespace guards pass |
| close | — | `39cdd761c chore(SPEC-ASTGREP-EDIT-001): draft -> in-progress (run phase complete)` | frontmatter transition `draft → in-progress` |

Sync-phase observed baseline (this run, this tree):
- `go test ./...` → exit 0, 105 `ok` packages, 0 FAIL
- `golangci-lint run --timeout=5m` → exit 0, `0 issues.`
- `grep -c 'SPEC-ASTGREP-EDIT-001' CHANGELOG.md` → 0 (pre-emission; safe to append)

## §E.3 Run-phase Audit-Ready Signal

run_status: complete
run_complete_at: 2026-07-26
run_commit_sha: "39cdd761c"

All 8 REQ-AGE-001..008 requirements implemented across M1-M7. `go test ./...`
and `golangci-lint run` both clean at run-phase completion.

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete
sync_complete_at: 2026-07-26
sync_commit_sha: "pending-backfill-sync"

CHANGELOG entry appended under `[Unreleased]`. Frontmatter transitions
`in-progress → implemented → completed` applied atomically across spec.md,
plan.md, acceptance.md (this single sync commit). `sync_commit_sha` above is
a placeholder — a commit cannot know its own hash — backfilled in the
immediately following `chore(SPEC-ASTGREP-EDIT-001): backfill sync_commit_sha`
commit per the SHA placeholder backfill exemption (D3).

## §F Phase 4 Mode Selection

Decision: `sub-agent` (Mode 5)

Rationale: sequential single-agent sync-phase work (CHANGELOG emission,
progress.md authoring, frontmatter transitions) — no multi-domain fan-out or
high-volume mechanical transformation applies. Implementation Kickoff
Approval is not applicable to sync-phase delegation.
