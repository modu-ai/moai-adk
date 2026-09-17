---
id: SPEC-GTD-CANON-BODY-001
title: "gtd canonical body migration — move the workflows/todo.md body into workflows/gtd.md and retire the todo body"
version: "0.3.0"
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
| 0.2.0 | 2026-09-18 | manager-spec | Plan-audit iteration 1 (FAIL 0.71) revisions D1-D14: baseline-delta test gate, script-based residual and body checks with observed RED and mutant controls, mirror `${CLAUDE_SKILL_DIR}` rule, catalog hash regeneration, per-verb GTD stage checks, historical set widened to CHANGELOG.md and top-level reports/, measured file count, comment citations made required, REQ-GCB-011 split. |
| 0.2.1 | 2026-09-18 | manager-spec | Plan-audit iteration 2 PASS 0.86; pre-run amendments N1 (bash-only guard + 113-line PASS floor), N2 (two-way flag set equality), N4 (per-package slot lease, exit 3/4 handling), N5 (M0 timeout is a Gap), N6 (6th/sixth-stage regex), N10 (single evidence file naming); cheap extras N3 (build-failed delta), N7 (exclude own review files), N8 (TestManifestHashFormat). |
| 0.3.0 | 2026-09-18 | manager-spec | Scope change from the lead: `TestGTDCanonicalSurfaceGolden` and `TestGTDAllTodoVerbsParity` now owned by t867 (moved from t854). Added REQ-GCB-014 and AC-GCB-012; both tests removed from the baseline set; failure causes measured (§C.4); parity repair left as an operator decision (plan.md §B.1); develop absorption before M0 (plan.md §D). |

## §A Context

Card **t867**. The `/moai:todo` → `/moai:gtd` rename moved the NAME but not the BODY. Measured at plan time in worktree `WT-gtd-canon` (base develop `f67d2193f`, SPEC commit `114737ea1`):

- CLI `moai gtd` is canonical and carries every todo verb plus the GTD stage verbs `capture`, `clarify`, `organize`, `reflect`, `engage` and the gate-response verb `answer` (`moai gtd --help`). `moai todo` is a compatibility alias over the same `backlog.db`.
- `.claude/commands/moai/gtd.md` is canonical; `.claude/commands/moai/todo.md` routes to `Skill("moai")` with `gtd` $ARGUMENTS.
- `.claude/skills/moai/workflows/gtd.md` is a 12-line stub that defers to `.claude/skills/moai/workflows/todo.md`. The real body (279 lines, 24,663 bytes, byte-identical in the template mirror) lives in `workflows/todo.md`.
- `.claude/skills/moai/SKILL.md` still carries `### todo - Backlog Queue` (line 169) pointing at `workflows/todo.md`.
- 21 files carry `moai todo` / `/moai:todo` / `workflows/todo.md`: 10 local, 10 template mirror, and 1 generated Codex TOML. With `todo.md` deleted in both trees, 19 files remain to clean.
- Eight Go files cite the deleted path. Four tests read it, one user-visible CLI help string (`internal/cli/todo.go:211`) names it, and three more files carry comment-only citations.
- Change set, measured: 19 doc edits (including the generated TOML) + 2 deletions + 8 Go files + `internal/template/catalog.yaml` + `.moai/reports/t867/verdict.md` = **31 paths**. That is above the Tier M 5-15 file guidance. It stays Tier M because the edits are mechanical retargeting with no new behavior (recorded under D14).
- Historical citations of `workflows/todo.md` measured outside this SPEC: `.moai/specs` + `.moai/reports` = 132 lines; `CHANGELOG.md` + top-level `reports/` = 13 lines in 6 files.

Operator decisions already made: delete the todo body outright (no pointer stub); retarget the four Go tests (approved scope expansion); keep the config key `workflow.todo.enabled`; update the comment-only Go citations (lane decision, final).

## §B Requirements (GEARS)

### REQ-GCB-001 — Canonical body lives in gtd.md

The `workflows/gtd.md` skill body shall carry the complete operational content of the former `workflows/todo.md` body. That means every section it held (What It Is, Commands, What the analyser may do, Reading the records, Picking the next card, Standing sources, Outside Kanban Mode, Boundaries, Cross-references), at a size no smaller than the migrated body minus the replaced stub, with every command example expressed as `moai gtd <verb>` / `/moai gtd`.

### REQ-GCB-002 — GTD stages documented from the CLI, not invented

The `workflows/gtd.md` skill body shall carry a `## GTD stages` section. It documents the stage verbs `capture`, `clarify`, `organize`, `reflect`, `engage` and the gate-response verb `answer`, each with the argument shape and every non-`--json` flag its `moai gtd <verb> --help` output lists. The section shall name no flag absent from that help output.

### REQ-GCB-003 — answer is not a stage; the queue boundary is stated

The `## GTD stages` section shall describe `answer` as the verb that answers a gate-blocked card, never as a sixth stage. It shall also state the queue boundary using the `moai gtd --help` wording "stay separate from the established development queue" and "explicitly approved Engage".

### REQ-GCB-004 — todo body deleted in both trees

The repository shall contain neither `.claude/skills/moai/workflows/todo.md` nor `internal/template/templates/.claude/skills/moai/workflows/todo.md`, and no pointer stub shall replace either file.

### REQ-GCB-005 — SKILL.md canonical gtd section

The `.claude/skills/moai/SKILL.md` Workflow Quick Reference (both trees) shall carry a gtd section that points detailed orchestration at `workflows/gtd.md`: `${CLAUDE_SKILL_DIR}/workflows/gtd.md` in the local tree, `.claude/skills/moai/workflows/gtd.md` in the template mirror. It shall document `/moai todo` and `moai todo` on a line carrying the marker `compat alias`. The Intent Router's backlog-language exemplar shall route to **gtd**.

### REQ-GCB-006 — Residual references cleaned

The 19 scoped surviving files (§C.1) shall contain no `workflows/todo.md` reference. Every remaining `moai todo` / `/moai todo` / `/moai:todo` line shall carry the marker phrase `compat alias` as a whole word.

### REQ-GCB-007 — Compatibility aliases keep working

The `moai todo` CLI alias and the `/moai:todo` slash command shall continue to reach the same queue behavior as `moai gtd` / `/moai:gtd` after the migration.

### REQ-GCB-008 — Mechanical consumers retargeted

When the todo body path is deleted, every Go consumer of that path or of its `moai todo` wording shall be retargeted to `workflows/gtd.md` and `moai gtd` wording, so that no test reads a deleted path and no help text or comment names one. That covers the four tests, the `moai todo` Long help string, and the comment citations.

### REQ-GCB-009 — Template-First and generated artifacts

The migration shall edit the `internal/template/templates/` mirror in the same change as each local file. The generated `internal/template/templates/.codex/agents/moai/manager-lead.toml` shall be regenerated by `make agents-emit` and the `internal/template/catalog.yaml` hashes by `make build` (`gen-catalog-hashes.go --all`); neither shall be hand-edited, and both regenerated files shall be committed. Local/mirror pairs byte-identical at base shall remain byte-identical.

### REQ-GCB-010 — Config key untouched

The configuration key `workflow.todo.enabled` shall not be renamed. The migration shall record its todo-named survival as a finding in `.moai/reports/t867/verdict.md`.

### REQ-GCB-011 — Historical records untouched

The migration shall not edit `CHANGELOG.md`, any file under top-level `reports/`, any file under `.moai/reports/` other than `.moai/reports/t867/`, or any file under `.moai/specs/` other than this SPEC's own directory.

### REQ-GCB-012 — Broken historical citations inventoried

When the deletion leaves citations of `workflows/todo.md` in the historical records named by REQ-GCB-011, the run/sync phase shall list every such citation (path:line) in `.moai/reports/t867/verdict.md`.

### REQ-GCB-013 — Scoped verification judged against a named baseline

While verifying this change locally, the implementer shall not run `go test ./...`. It shall run `./internal/cli/...` and `./internal/template/...` with an explicit `-timeout` and judge them as "no NEW failing test name versus the baseline set recorded at pre-edit HEAD". The five baseline failures owned by card t854 (§C.3) shall be recorded as a finding for the lead, not fixed by this SPEC. The two GTD-surface tests owned by this SPEC (REQ-GCB-014) are excluded from that baseline and are judged by name.

### REQ-GCB-014 — GTD surface tests owned by this card pass

When run-phase completes, `TestGTDCanonicalSurfaceGolden` (`./internal/template/`) and `TestGTDAllTodoVerbsParity` (`./internal/cli/`) shall each pass when run alone by exact name. The repair shall keep the `moai todo` and `/moai:todo` compat aliases working. It shall touch no Go code outside `internal/cli` and `internal/template`. It shall make no CLI behavior change beyond the one the operator selects for the parity decision in plan.md §B.1.

## §C Scope

### §C.1 In-scope documentation files (19 surviving + 2 deleted)

Local tree, each with a template mirror under `internal/template/templates/`:

- `.claude/skills/moai/workflows/gtd.md` (receives body)
- `.claude/skills/moai/workflows/todo.md` (deleted, both trees)
- `.claude/skills/moai/SKILL.md`
- `.claude/agents/moai/manager-lead.md`
- `.claude/rules/moai/workflow/kanban-dispatch.md`
- `.claude/rules/moai/workflow/kanban-dispatch-detail.md`
- `.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md`
- `.claude/skills/moai-kanban-foreman/SKILL.md`
- `.claude/skills/moai/workflows/project/doc-generation.md`
- `.moai/docs/todo-queue-storage.md` (content only; file name kept)

Template tree only: `internal/template/templates/.codex/agents/moai/manager-lead.toml` (generated).

### §C.2 In-scope Go and generated files (all required)

- Tests retargeted: `internal/cli/todo_skill_doc_test.go`, `internal/cli/todo_landed_doc_test.go`, `internal/cli/doc_json_shape_test.go`, `internal/template/backlog_json_disclosure_mirror_test.go`
- Help string: `internal/cli/todo.go:211`
- Comment citations: `internal/cli/todo_drop.go:16`, `internal/cli/todo_edit_move.go:8`, `internal/cli/todo_edit_move.go:100`, `internal/cli/todo_test.go:665`, `internal/template/backlog_json_disclosure_mirror_test.go:5`
- Generated: `internal/template/catalog.yaml` (via `make build`)
- Evidence: `.moai/reports/t867/verdict.md`

### §C.3 Pre-existing baseline failures owned by card t854 (finding, not fixed)

The plan auditor observed these at `114737ea1`, recorded in `.moai/reports/plan-audit/SPEC-GTD-CANON-BODY-001-review-1.md`. The run phase re-measures them on the absorbed tree at M0:

- `./internal/template/...`: `TestRealSetCodexShape`, `TestBoundaryFlagsRecorded`
- `./internal/cli/...`: `TestCGRetirementCompleteEntryShapesAndCounters`, `TestInitRegroup_SecondGroupGolden`, `TestAgentWiringOptions`, plus a 10-minute package timeout under load

Card t854 owns these five; this SPEC does not fix them.

### §C.4 GTD surface tests owned by this card (REQ-GCB-014)

Ownership moved to t867 from card t854 because both tests overlap this card's GTD surface. I measured each on this tree at `e4cc628e9`, running it alone with `-timeout 5m -count=1 -v -run '^<Test>$'`:

- `TestGTDCanonicalSurfaceGolden`: `internal/template/gtd_canonical_surface_test.go:24`, output `templates/.claude/commands/moai/todo.md is not a thin gtd compatibility path`.
  - Cause: the assertion (line 23) requires the literal `arguments: gtd $ARGUMENTS`. The template body now reads ``invoke `Skill("moai")` with arguments: `gtd` $ARGUMENTS`` (line 8), with backticks around `gtd`. Cards t860/t861 introduced that wording; the published `.agents/skills/moai-todo/SKILL.md` line 8 carries the same wording.
  - The test calls `t.Fatalf` on the first path, so the second todo path and the `publishedSkillNames == 17` check were **not reached** in this measurement.
  - Repair class: test-literal alignment inside `internal/template`. No CLI behavior change and no command-body change: reverting the bodies would undo t860 and break the `commandemit` golden.
- `TestGTDAllTodoVerbsParity`: `internal/cli/gtd_compat_test.go:61`, output `gtd verbs = [add analyze answer auto-done capture clarify done drop edit engage export-json history landed list move next organize pr reflect relate undone undrop unpick unrelate why], want [add analyze auto-done capture clarify done drop edit engage export-json history landed list move next organize pr reflect relate undone undrop unpick unrelate why]`.
  - Cause: `moai gtd` gained the `answer` verb (commit `1b644372d`, cards t863/t864), but the test's `gtdWant` (lines 51-54) was not updated.
  - The `t.Fatalf` at line 61 stops the test, so the todo-surface check (line 63) and the per-verb help/flag parity loop were **not reached**.
  - `moai todo answer --help` (scratch binary built from this tree) exits 0 and prints the todo root Long help, so `answer` is not a todo subcommand.
  - Repair class: **undecided**; see plan.md §B.1.

### §C.5 Local command-body divergence (finding only)

The local `.claude/commands/moai/todo.md` line 7 still reads `Use Skill("moai") with arguments: gtd $ARGUMENTS`, while the template reads the harness-neutral form. This SPEC does not change it; the verdict records it as a finding.

## §D Exclusions

### Out of Scope — renames beyond the body

- Renaming the config key `workflow.todo.enabled` or `TodoEnabled()` (finding only)
- Renaming `moai todo` Go source files (`internal/cli/todo*.go`) or the `.moai/docs/todo-queue-storage.md` file name
- Removing the `moai todo` CLI alias, `.claude/commands/moai/todo.md`, or the published `moai-todo` skill

### Out of Scope — pre-existing test failures

- Fixing any of the five t854-owned tests in §C.3 (the two GTD-surface tests in §C.4 are in scope)
- Changing the local `.claude/commands/moai/todo.md` dispatch line (§C.5)

### Out of Scope — runtime wording and generated project docs

- Runtime-emitted `moai todo` wording in Go strings (`internal/hook/session_start_kanban*.go`, `internal/web/assets/i18n.js`, `internal/statusline/*`, and the `moai todo` usage examples in `internal/cli/todo.go` Long help other than the `workflows/todo.md` path reference). These stay because `moai todo` remains a valid compat alias.
- `.moai/project/codemaps/*` wording (for example `data-flow.md`)

### Out of Scope — historical and external surfaces

- Editing `CHANGELOG.md`, top-level `reports/**`, `.moai/reports/**` (other than `.moai/reports/t867/`), or `.moai/specs/**` (other than this SPEC)
- docs-site and README pages (plan-time grep found zero `workflows/todo.md` citations there)
- Changing GTD CLI behavior; the body documents existing behavior only
