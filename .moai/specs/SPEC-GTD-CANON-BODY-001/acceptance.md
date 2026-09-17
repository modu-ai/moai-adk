# Acceptance — SPEC-GTD-CANON-BODY-001 (card t867)

All commands run from the worktree root. `<wt>` = `.claude/worktrees/t867`.

## Allowed-survivor rule

A line in a scoped file (§C.1 of spec.md, both trees, plus the generated `manager-lead.toml`) that matches `moai todo|/moai todo|/moai:todo` is an **allowed survivor** only when that same line contains the case-insensitive token `compat` (e.g. "compatibility alias", "compat alias"). Any other match is a residual defect. `workflows/todo.md` has no allowed survivors anywhere outside `.moai/specs/` and `.moai/reports/`.

Scoped file list used below (`$FILES`):

```
.claude/agents/moai/manager-lead.md
.claude/rules/moai/workflow/kanban-dispatch.md
.claude/rules/moai/workflow/kanban-dispatch-detail.md
.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
.claude/skills/moai-kanban-foreman/SKILL.md
.claude/skills/moai/SKILL.md
.claude/skills/moai/workflows/gtd.md
.claude/skills/moai/workflows/project/doc-generation.md
.moai/docs/todo-queue-storage.md
```
each also under `internal/template/templates/`, plus `internal/template/templates/.codex/agents/moai/manager-lead.toml`.

## §D AC Matrix

### AC-GCB-001 — gtd.md carries the full body (REQ-GCB-001)

Given the migrated tree, When `grep -cE '^## (What It Is|Commands|Reading the records|Picking the next card|Standing sources|Outside Kanban Mode|Boundaries|Cross-references)$' .claude/skills/moai/workflows/gtd.md` runs, Then it prints `8`, and `grep -c '^### What the analyser may do$'` on the same file prints `1`; the template mirror gives identical counts and `cmp` of the two copies exits 0.

### AC-GCB-002 — GTD stages documented and backed by --help (REQ-GCB-002, REQ-GCB-003)

Given the built binary, When `grep -cE 'moai gtd (capture|clarify|organize|reflect|engage|answer)' .claude/skills/moai/workflows/gtd.md` runs, Then every one of the six verbs appears at least once; And for each `--<flag>` named next to a stage verb in gtd.md, `bin/moai gtd <verb> --help` output contains that flag; And gtd.md contains a sentence stating captured items stay separate from the queue and only an approved `engage` publishes into it.

### AC-GCB-003 — todo body deleted, no stub (REQ-GCB-004)

Given the migrated tree, When `test ! -e .claude/skills/moai/workflows/todo.md && test ! -e internal/template/templates/.claude/skills/moai/workflows/todo.md && echo GONE` runs, Then it prints `GONE`.

### AC-GCB-004 — SKILL.md gtd section (REQ-GCB-005)

Given the migrated tree, When `grep -n '### todo - Backlog Queue' .claude/skills/moai/SKILL.md` runs, Then it prints nothing (exit 1); And `grep -c 'workflows/gtd.md' .claude/skills/moai/SKILL.md` prints at least `1`; And a line in the gtd section containing `compat` names both `/moai todo` and `moai todo`; And the backlog-language Intent Router line routes to `**gtd**`. The template mirror satisfies the same checks.

### AC-GCB-005 — no residual workflows/todo.md citation (REQ-GCB-006, REQ-GCB-008)

Given the migrated tree, When `grep -rn 'workflows/todo\.md' .claude internal cmd pkg .moai/docs Makefile` runs, Then it prints nothing (exit 1). (Comment-only Go citations left by discretion would appear here; the implementer either updates them or records each in the verdict as a deliberate survivor with justification — a silent survivor fails this AC.)

### AC-GCB-006 — residual todo wording is alias documentation only (REQ-GCB-006)

Given `$FILES`, When `grep -nE 'moai todo|/moai todo|/moai:todo' $FILES | grep -viE 'compat'` runs, Then it prints nothing (exit 1).

### AC-GCB-007 — compat aliases keep working (REQ-GCB-007)

Given the built binary and an isolated queue root (tests' `todoFixture`), When `go test ./internal/cli/ -run 'TestGTDAllTodoVerbsParity|TestGTDFiveStageCLIUsesSameSQLite|TestTodoBareInvocationLists' -count=1` runs, Then it reports `ok`; And `go test ./internal/template/ -run TestGTDCanonicalSurfaceGolden -count=1` reports `ok` (confirms `commands/moai/todo.md` and `.agents/skills/moai-todo/SKILL.md` still dispatch `gtd $ARGUMENTS`); And `bin/moai todo --help` exits 0.

### AC-GCB-008 — retargeted tests read gtd.md (REQ-GCB-008)

Given the migrated tree, When `grep -nE 'workflows", "todo\.md|workflows/todo\.md' internal/cli/todo_skill_doc_test.go internal/cli/todo_landed_doc_test.go internal/cli/doc_json_shape_test.go internal/template/backlog_json_disclosure_mirror_test.go` runs, Then it prints nothing; And `go test ./internal/cli/ -run 'TestTodoDoctrine_MirrorParityAndStatedColumnCount|TestTodoListJSONShapeMatchesDoc' -count=1` and the todo_skill_doc test and `go test ./internal/template/ -run TestBacklogJSONDisclosure -count=1` report `ok`; And `bin/moai todo --help` output does not contain `workflows/todo.md`.

### AC-GCB-009 — generated TOML regenerated, build clean (REQ-GCB-009)

Given the migrated tree, When `make build` runs, Then it exits 0 (its `agents-emit-check` and `commands-emit-check` prerequisites pass, proving `manager-lead.toml` matches its `.md` source and published skills match command sources).

### AC-GCB-010 — scoped test and vet gate (REQ-GCB-012)

Given the migrated tree, When `go vet ./internal/cli/... ./internal/template/...` and `go test ./internal/cli/... ./internal/template/...` run, Then both exit 0; And no `go test ./...` invocation appears in the §E.2 evidence.

### AC-GCB-011 — config key and historical records (REQ-GCB-010, REQ-GCB-011)

Given the change set, When `git diff --name-only <base>...HEAD -- .moai/specs .moai/reports` runs, Then every listed path is under `.moai/specs/SPEC-GTD-CANON-BODY-001/` or `.moai/reports/t867/`; And `git diff <base>...HEAD -- internal/config .moai/config internal/template/templates/.moai/config | grep -c 'todo'` prints `0`; And `.moai/reports/t867/verdict.md` contains a `workflow.todo.enabled` finding and a citation list whose entry count equals the line count of `grep -rn 'workflows/todo\.md' .moai/specs .moai/reports --exclude-dir=SPEC-GTD-CANON-BODY-001 --exclude-dir=t867`.

## Edge Cases

- A `moai todo` line that documents the alias but lacks the word `compat` fails AC-GCB-006 — reword it rather than widening the rule.
- A section anchor (`workflows/todo.md § Standing sources`) must be retargeted to the gtd.md section of the same name; AC-GCB-001 guarantees the heading exists.
- The `.agents/skills/moai` deploy-time mirror stops carrying `workflows/todo.md`; AC-GCB-010's template package tests cover the mirror code path.

## Definition of Done

- AC-GCB-001 through AC-GCB-011 PASS with verbatim command output in progress.md §E.2
- Template-First parity (local and mirror edited together) and `make build` clean
- `.moai/reports/t867/verdict.md` written with the config-key finding and historical citation inventory
- Card id `t867` in every commit message
