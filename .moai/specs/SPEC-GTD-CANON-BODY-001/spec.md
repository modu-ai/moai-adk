---
id: SPEC-GTD-CANON-BODY-001
title: "gtd canonical body migration — move the workflows/todo.md body into workflows/gtd.md and retire the todo body"
version: "0.1.0"
status: draft
created: 2026-09-18
updated: 2026-09-18
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
tags: "gtd, todo, kanban, skill-body, compat-alias, template-first, card-t867"
tier: M
card: t867
related_specs: []
---

# SPEC-GTD-CANON-BODY-001 — gtd canonical body migration

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-18 | manager-spec | Initial plan-phase draft for kanban card t867 (Tier M, Class C). |

## §A Context

Card **t867**. The `/moai:todo` → `/moai:gtd` rename moved the NAME but not the BODY. Measured at plan time (worktree `WT-gtd-canon`, base develop `f67d2193f`):

- CLI `moai gtd` is canonical and carries every todo verb plus `capture`, `clarify`, `organize`, `reflect`, `engage`, `answer` (`moai gtd --help`). `moai todo` is a compatibility alias over the same `backlog.db`.
- `.claude/commands/moai/gtd.md` is canonical; `.claude/commands/moai/todo.md` routes `Skill("moai")` with `gtd $ARGUMENTS`.
- `.claude/skills/moai/workflows/gtd.md` is a 609-byte stub that defers to `.claude/skills/moai/workflows/todo.md`; the real 24,663-byte body (byte-identical in the template mirror) lives in `workflows/todo.md`.
- `.claude/skills/moai/SKILL.md` still carries `### todo - Backlog Queue` pointing at `workflows/todo.md`, while its Intent Router already routes `todo` to gtd.
- 21 files (10 local + 10 template mirror + 1 generated Codex TOML) carry `moai todo` / `/moai:todo` / `workflows/todo.md` references.
- Five Go tests pin the path `workflows/todo.md` or `moai todo` wording inside that body, and one user-visible CLI help string (`internal/cli/todo.go` Long help) names `workflows/todo.md`.

Operator decisions already made: the todo body file is deleted outright (no pointer stub); four Go tests are retargeted (approved scope expansion); the config key `workflow.todo.enabled` is not renamed.

## §B Requirements (GEARS)

### REQ-GCB-001 — Canonical body lives in gtd.md

The `workflows/gtd.md` skill body shall carry the complete operational content of the former `workflows/todo.md` body — every section it held (What It Is, Commands, What the analyser may do, Reading the records, Picking the next card, Standing sources, Outside Kanban Mode, Boundaries, Cross-references) — with every command example expressed as `moai gtd <verb>` / `/moai gtd`.

### REQ-GCB-002 — GTD five stages documented from the CLI, not invented

The `workflows/gtd.md` skill body shall document the `capture`, `clarify`, `organize`, `reflect`, `engage` stages and the `answer` verb, and each documented verb, argument shape, and flag shall exist in the corresponding `moai gtd <verb> --help` output of the binary built from the same tree.

### REQ-GCB-003 — Stage boundary with the queue stated

The `workflows/gtd.md` skill body shall state that captured GTD items stay separate from the development queue and that only an explicitly approved `engage` publishes into the shared `backlog.db`, consistent with the `moai gtd --help` long description.

### REQ-GCB-004 — todo body deleted in both trees

The repository shall contain neither `.claude/skills/moai/workflows/todo.md` nor `internal/template/templates/.claude/skills/moai/workflows/todo.md`, and no pointer stub shall replace either file.

### REQ-GCB-005 — SKILL.md canonical gtd section

The `.claude/skills/moai/SKILL.md` Workflow Quick Reference (both trees) shall carry a gtd section that points detailed orchestration at `workflows/gtd.md` and documents `/moai todo` and `moai todo` as compatibility aliases; the Intent Router's backlog-language exemplar shall route to **gtd**.

### REQ-GCB-006 — Residual references cleaned

The 20 surviving scoped files (§C.1) and the generated `manager-lead.toml` shall contain no `workflows/todo.md` reference, and every remaining `moai todo` / `/moai todo` / `/moai:todo` occurrence shall be an allowed survivor as defined in acceptance.md § Allowed-survivor rule.

### REQ-GCB-007 — Compatibility aliases keep working

The `moai todo` CLI alias and the `/moai:todo` slash command shall continue to reach the same queue behavior as `moai gtd` / `/moai:gtd` after the migration.

### REQ-GCB-008 — Mechanical consumers retargeted

When the todo body path is deleted, every mechanical consumer of that path or of its `moai todo` wording (Go tests and the user-visible `moai todo` / `moai gtd` Long help string) shall be retargeted to `workflows/gtd.md` and `moai gtd` wording, so that no test reads a deleted path and no help text names one.

### REQ-GCB-009 — Template-First and generated artifacts

The migration shall edit the `internal/template/templates/` mirror in the same change as each local file, and the generated `internal/template/templates/.codex/agents/moai/manager-lead.toml` shall be regenerated by `make agents-emit`, never hand-edited.

### REQ-GCB-010 — Config key untouched

The configuration key `workflow.todo.enabled` shall not be renamed; the migration shall record its todo-named survival as a finding in `.moai/reports/t867/verdict.md`.

### REQ-GCB-011 — Historical records untouched, broken citations inventoried

The migration shall not edit any file under `.moai/specs/` (other than this SPEC's own directory) or `.moai/reports/` (other than `.moai/reports/t867/`); When the deletion leaves citations of `workflows/todo.md` in those historical records, the run/sync phase shall list every such citation (path:line) in `.moai/reports/t867/verdict.md`.

### REQ-GCB-012 — Scoped verification only

While verifying this change locally, the implementer shall not run `go test ./...`; verification shall be scoped to `./internal/cli/...` and `./internal/template/...` plus `go vet` and `make build`.

## §C Scope

### §C.1 In-scope files (each has a template mirror under `internal/template/templates/`)

- `.claude/skills/moai/workflows/gtd.md` (receives body)
- `.claude/skills/moai/workflows/todo.md` (deleted)
- `.claude/skills/moai/SKILL.md`
- `.claude/agents/moai/manager-lead.md` (+ generated `.codex/agents/moai/manager-lead.toml` in the template tree only)
- `.claude/rules/moai/workflow/kanban-dispatch.md`
- `.claude/rules/moai/workflow/kanban-dispatch-detail.md`
- `.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md`
- `.claude/skills/moai-kanban-foreman/SKILL.md`
- `.claude/skills/moai/workflows/project/doc-generation.md`
- `.moai/docs/todo-queue-storage.md` (content only; file name kept)

### §C.2 In-scope Go files

- `internal/cli/todo_skill_doc_test.go`, `internal/cli/todo_landed_doc_test.go`, `internal/cli/doc_json_shape_test.go`, `internal/template/backlog_json_disclosure_mirror_test.go` (operator-approved retarget)
- `internal/cli/todo.go` Long help string near line 211 (user-visible text naming the deleted path — required by REQ-GCB-008)
- Comment-only citations (implementer discretion, recommended): `internal/cli/todo_drop.go:16`, `internal/cli/todo_edit_move.go:8`, `internal/cli/todo_edit_move.go:100`, `internal/cli/todo_test.go:665`, `internal/template/backlog_json_disclosure_mirror_test.go:5`

## §D Exclusions

### Out of Scope — renames beyond the body

- Renaming the config key `workflow.todo.enabled` or `TodoEnabled()` (finding only)
- Renaming `moai todo` Go source files (`internal/cli/todo*.go`) or the `.moai/docs/todo-queue-storage.md` file name
- Removing the `moai todo` CLI alias, `.claude/commands/moai/todo.md`, or the published `moai-todo` skill

### Out of Scope — historical and external surfaces

- Editing `.moai/specs/**` or `.moai/reports/**` records other than this SPEC and `.moai/reports/t867/`
- docs-site and README pages (plan-time grep found zero `workflows/todo.md` citations there)
- Changing GTD CLI behavior; the body documents existing behavior only
