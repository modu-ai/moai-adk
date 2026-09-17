# Plan — SPEC-GTD-CANON-BODY-001 (card t867)

## §A Context

Tier M, Class C. Body migration from `workflows/todo.md` to `workflows/gtd.md`, deletion of the todo body, residual-reference cleanup across 19 surviving documentation files, retargeting of 8 Go files, and regeneration of `manager-lead.toml` and `catalog.yaml`. Measured change set: 31 paths (spec.md §A). The work is doc-plus-test-retarget with no Go behavior change. Two check scripts in this SPEC directory (`check-residual.sh`, `check-gtd-body.sh`) are the mechanical gates, and each has an observed RED on the pre-edit tree (acceptance.md § Evidence ledger).

## §B Known Issues / Decisions

- **Decision (operator):** delete `workflows/todo.md` in both trees, no stub.
- **Decision (operator):** retarget the 4 Go tests.
- **Decision (lane, final):** the comment-only Go citations are updated (required, not discretionary).
- **Decision (orchestrator):** the five pre-existing red tests owned by card t854 (spec.md §C.3) are NOT fixed by this SPEC. Verification is a baseline delta over that five-name set.
- **Decision (lead, scope change):** `TestGTDCanonicalSurfaceGolden` and `TestGTDAllTodoVerbsParity` are owned by this SPEC (REQ-GCB-014, spec.md §C.4). They must pass by name and are never part of the baseline delta.
- **Plan-time finding:** `internal/cli/todo.go:211` names `workflows/todo.md` inside the Cobra Long help string. That is user-visible text, so fixing it is required.
- **Measured (`e4cc628e9`):** `internal/template/gtd_canonical_surface_test.go:24` fails because the literal it requires (`arguments: gtd $ARGUMENTS`, line 23) predates the t860/t861 wording ``with arguments: `gtd` $ARGUMENTS``. The repair aligns the test literal to the landed wording (spec.md §C.4). It needs no command-body change and no operator decision.

### §B.1 Operator decision — parity repair (RESOLVED: option A, 2026-09-18, via lead)

**Decision:** option A. `answer` stays gtd-only; add `"answer"` to `gtdWant` in `internal/cli/gtd_compat_test.go` only; the CLI is unchanged. The options as they were presented are kept below for the record.

`internal/cli/gtd_compat_test.go:61` fails because `moai gtd` has `answer` (commit `1b644372d`) and the test's `gtdWant` does not. `moai todo` has no `answer` subcommand. Both options stay inside `internal/cli` and keep `moai todo` / `/moai:todo` working. Choosing between them changes (or preserves) the CLI surface, so this SPEC does not decide.

| Option | Change | Consequence |
|---|---|---|
| A — `answer` stays gtd-only | Add `"answer"` to `gtdWant` only (test file). No Go production change. | The CLI surface is unchanged. `answer` joins capture/clarify/organize/reflect/engage as a gtd-only verb, so the todo alias keeps exactly the historical todo verb set. `moai todo answer …` stays unavailable. How a multi-word `moai todo answer t1 text` is treated today (refused as a mistyped verb, or added as a card via the phrase fallthrough) was **not measured**; run-phase must observe it in an isolated `todoFixture` queue before landing, and the verdict records it. |
| B — expose `answer` on the todo alias | Register `answer` on the `moai todo` command in `internal/cli`, and add `"answer"` to both `todoWant` and `gtdWant`. | A CLI behavior change: the compat alias gains a verb it never had. The parity loop then also asserts identical Use/Short/flags/help for `answer` on both roots. `workflows/gtd.md` must list `answer` among the verbs the alias shares, and the alias stops being "the historical todo verbs". |

Run-phase applies option A (M3.8). The unmeasured `moai todo answer t1 x` behavior is observed against an isolated queue under AC-GCB-013. A card-adding result is a finding and is not fixed.
- **Plan-time finding:** `make build` runs `gen-catalog-hashes.go --all` (Makefile:35). Changing the `moai` skill tree, `moai-kanban-foreman` and `manager-lead` therefore changes `internal/template/catalog.yaml` hashes, which `TestCatalogHashCoversSkillSubfiles` pins. The regenerated catalog must be committed.
- **Plan-time finding:** the template mirror `SKILL.md` must not contain `${CLAUDE_SKILL_DIR}` (`TestSkillTreeHasNoClaudeSkillDirToken`). The mirror writes `.claude/skills/moai/workflows/<x>.md` (mirror SKILL.md lines 126/134/142/150).
- **Plan-time finding:** `Makefile`, `internal/template/skill_mirror*.go` and `internal/template/commandemit/*` enumerate no `workflows/todo.md`. The `todo.md` that `commandemit` mentions is `commands/moai/todo.md`, which is retained.
- **Finding to record:** `workflow.todo.enabled` / `TodoEnabled()` keep the todo name.

## §C Pre-flight

- `git -C <wt> branch --show-current` → `WT-gtd-canon`
- `cmp` of local vs template `workflows/todo.md` → identical (measured)
- `go build -o <scratch>/moai ./cmd/moai`; the GTD stage help is the source for REQ-GCB-002 (flag set frozen in `check-gtd-body.sh`)
- **M0 baseline** (acceptance.md AC-GCB-010): record the failing test-name sets for `./internal/template/...` and `./internal/cli/...` at pre-edit HEAD, under a resource slot, with `-timeout 30m`

## §D Constraints

- **Before run-phase starts:** absorb local `develop` into `WT-gtd-canon`. At plan time local develop was `27220fb94` with the t783 merge pending; re-read `git rev-parse develop` at absorption. Measure the M0 baseline on the absorbed tree, never on the pre-absorption tree. Scope judgments use `git merge-base develop HEAD` as their left end.
- Template-First: every doc edit lands in the local file and its mirror in the same commit, then `make build`.
- `manager-lead.toml` is regenerated only via `make agents-emit`; `catalog.yaml` only via `make build`.
- No local `go test ./...`. Each heavy package run takes its own lease (`moai slot acquire --resource go-test --max-duration 45m`) and releases it before the next package. Acquire exit 3 or 4 means wait and retry, bounded; if the lease stays unavailable, record a Gap. A package never runs unleased (acceptance.md AC-GCB-010).
- Both check scripts run under bash only; they refuse other shells with exit 2.
- Template neutrality: no SPEC IDs, card ids, dates or SHAs inside `internal/template/templates/**`.
- Every allowed `moai todo` survivor line carries the literal marker `compat alias`.

## §F Milestones (ordered by decision-reversibility)

### M0 — Baseline capture (Priority High)

Record the failing test-name sets at pre-edit HEAD for both packages (AC-GCB-010 procedure) as `.moai/reports/t867/m0-{template,cli}.{log,exit,fail-names}`. An M0 timeout is a Gap: re-run the package once, and never treat a truncated baseline as authoritative. Record RED-now for both check scripts at `.moai/reports/t867/m0-check-{residual,body}.txt`.

### M1 — gtd.md canonical body (Priority High)

1. Rewrite `workflows/gtd.md` in both trees, byte-identical: carry every todo.md section and retarget command examples to `moai gtd` / `/moai gtd`.
2. Add `## GTD stages`. It names the stage order capture → clarify → organize → reflect → engage. Per verb it gives the usage shape and every non-`--json` flag exactly as `check-gtd-body.sh` lists them (23 required flags). It describes `answer` as answering a gate-blocked card, never as a stage, and carries the two queue-boundary phrases verbatim.
3. Add one compat paragraph whose `moai todo` / `/moai todo` lines each carry `compat alias`.
4. Keep what the retargeted tests read: `| \`moai gtd landed ` / `| \`moai gtd pr ` rows with "carries <word> tab-separated columns", a `moai gtd history` mention, and the JSON fence keys.

### M2 — SKILL.md gtd section (Priority High)

1. Replace `### todo - Backlog Queue` with a gtd section. Local points at `${CLAUDE_SKILL_DIR}/workflows/gtd.md`; mirror points at `.claude/skills/moai/workflows/gtd.md`. Verbs are in `moai gtd` form, the compat alias line carries `compat alias`, and the enablement paragraph keeps the `workflow.todo.enabled` key name.
2. Retarget the Intent Router backlog-language exemplar to route to **gtd**.

### M3 — Retarget Go consumers (Priority High)

1. `todo_skill_doc_test.go`: path → `workflows/gtd.md`, count `moai gtd history`.
2. `todo_landed_doc_test.go`: path → `workflows/gtd.md`, row prefixes → `moai gtd landed` / `moai gtd pr`.
3. `doc_json_shape_test.go`: doc surfaces → `workflows/gtd.md`.
4. `backlog_json_disclosure_mirror_test.go`: file list (line 25) and comment (line 5) → `workflows/gtd.md`.
5. `todo.go:211` Long help → `workflows/gtd.md`.
6. Comments (required): `todo_drop.go:16`, `todo_edit_move.go:8`, `todo_edit_move.go:100`, `todo_test.go:665`.
7. `internal/template/gtd_canonical_surface_test.go:23`: align the thin-path literal to the landed dispatch wording (``arguments: `gtd` $ARGUMENTS``). Then confirm by name that both todo paths and the `publishedSkillNames == 17` check are reached and pass. Those two checks were not reached at plan time. The commit message for this step must cite the commits whose wording it follows: `1dcaad954` (t860) and `61582178d` (t861).
8. `internal/cli/gtd_compat_test.go`: add `"answer"` to `gtdWant` only (option A). No production Go change, and `todoWant` unchanged.
9. Observe `moai todo answer t1 x` against an isolated queue (AC-GCB-013) and record the refused-or-added result in the verdict.

### M4 — Delete todo body and clean residual references (Priority Medium)

1. Delete `workflows/todo.md` in both trees.
2. Clean the remaining §C.1 files in both trees. Anchors like `workflows/todo.md § Standing sources` become `workflows/gtd.md § Standing sources`, CLI examples become `moai gtd <verb>`, and kept alias mentions carry `compat alias`.
3. `make agents-emit` to regenerate `manager-lead.toml`.

### M5 — Build, verify, record (Priority Medium)

1. `make build`: agents-emit-check, commands-emit-check, and `gen-catalog-hashes.go --all` (rewrites `internal/template/catalog.yaml` — commit it).
2. Run `check-residual.sh` and `check-gtd-body.sh` (PASS required); run the retargeted tests by name; run the baseline-delta package runs, writing `.moai/reports/t867/m5-{template,cli}.{log,exit,fail-names}`.
3. Write `.moai/reports/t867/verdict.md` with the `workflow.todo.enabled` finding, the §C.3 pre-existing failures as a lead finding, and the path:line inventory of historical `workflows/todo.md` citations across `.moai/specs`, `.moai/reports`, `CHANGELOG.md` and `reports/`.

## §G Anti-Patterns

- Leaving a pointer stub at `workflows/todo.md`.
- Writing `${CLAUDE_SKILL_DIR}` into the template mirror.
- Hand-editing `manager-lead.toml` or `catalog.yaml`, or leaving the regenerated catalog uncommitted.
- Documenting GTD flags absent from `--help`, or presenting `answer` as a stage.
- Making a check pass by weakening it (loosening the `compat alias` marker, deleting a test assertion, dropping a flag from the required list).
- Fixing or deleting the five t854-owned §C.3 failing tests inside this SPEC.
- Making the two t867-owned tests pass by deleting or weakening their assertions (for example removing the thin-path check or the verb-set comparison).
- Editing historical SPEC, report, or CHANGELOG files to silence citations.

## §H Cross-References

- `.claude/rules/moai/workflow/kanban-dispatch.md` (queue doctrine consumer)
- `.claude/skills/moai/workflows/project/doc-generation.md` (standing-source consumer)
- `internal/cli/gtd.go`, `internal/cli/gtd_answer.go` (stage behavior source)
- `internal/template/skill_dir_token_guard_test.go`, `internal/template/catalog_tier_audit_test.go` (mirror token and catalog hash guards)
- `.moai/reports/plan-audit/SPEC-GTD-CANON-BODY-001-review-1.md` (iteration-1 audit and baseline observation)
