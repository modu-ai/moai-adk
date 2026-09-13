# SPEC-DOCS-TODO-TEMP-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-DOCS-TODO-TEMP-GUARD-001
tier: S
status: draft
card: t575
worktree: .claude/worktrees/t575
branch: WT-docs-todo-locales
authoring_only: true
artifacts_emitted:
  - .moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/spec.md
  - .moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/plan.md
  - .moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/acceptance.md
  - .moai/specs/SPEC-DOCS-TODO-TEMP-GUARD-001/progress.md
grounding_evidence:
  truth_table: spec.md §2 (code-cited: internal/kanban/todo_root.go, state_dir.go, temp_origin.go)
  false_sentences: docs-site/content/<locale>/utility-commands/moai-todo.md lines 59 and 248, all 4 locales (verified present)
  heading_parity_baseline: 29 per file (verified equal across ko/en/ja/zh)
  tracking_owner_gap: confirmed — .moai/docs/followup-candidates.md absent; only CHANGELOG line 440 mentions the doc follow-up
  spec_id_check: "ID=SPEC-DOCS-TODO-TEMP-GUARD-001 regex ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ → PASS (verbatim Bash output); uniqueness → no existing directory"
premise_correction: card quotes the stale path ~/.moai/todo/<key>/; current pages say ~/.moai/db/<key>/todo/backlog.db (post-t621 rekey) — defect class unchanged
doc_pages_edited: false   # authoring only per delegation prompt
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
