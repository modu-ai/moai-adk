# Plan — SPEC-GTD-CANON-BODY-001 (card t867)

## §A Context

Tier M, Class C. Body migration from `workflows/todo.md` to `workflows/gtd.md`, deletion of the todo body, residual-reference cleanup across 19 surviving documentation files, retargeting of 8 Go files, and regeneration of `manager-lead.toml` and `catalog.yaml`. Measured change set: 31 paths (spec.md §A). The work is doc-plus-test-retarget with no Go behavior change. Two check scripts in this SPEC directory (`check-residual.sh`, `check-gtd-body.sh`) are the mechanical gates, and each has an observed RED on the pre-edit tree (acceptance.md § Evidence ledger).

## §B Known Issues / Decisions

- **Decision (operator):** delete `workflows/todo.md` in both trees, no stub.
- **Decision (operator):** retarget the 4 Go tests.
- **Decision (lane, final):** the comment-only Go citations are updated (required, not discretionary).
- **Decision (orchestrator):** the pre-existing red tests (spec.md §C.3) are NOT fixed by this SPEC. Verification is a baseline delta over the same test-name set.
- **Plan-time finding:** `internal/cli/todo.go:211` names `workflows/todo.md` inside the Cobra Long help string. That is user-visible text, so fixing it is required.
- **Plan-time finding:** `internal/template/gtd_canonical_surface_test.go` is **already red on the base tree** (plan-audit iteration 1). It fails because `commands/moai/todo.md` lost the literal `arguments: gtd $ARGUMENTS` in cards t860/t861, which is unrelated to this SPEC. This SPEC does not touch that file. The test stays red and is carried in the baseline set.
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

- Template-First: every doc edit lands in the local file and its mirror in the same commit, then `make build`.
- `manager-lead.toml` is regenerated only via `make agents-emit`; `catalog.yaml` only via `make build`.
- No local `go test ./...`. Heavy package runs take `moai slot acquire --resource go-test --max-duration 40m` first.
- Template neutrality: no SPEC IDs, card ids, dates or SHAs inside `internal/template/templates/**`.
- Every allowed `moai todo` survivor line carries the literal marker `compat alias`.

## §F Milestones (ordered by decision-reversibility)

### M0 — Baseline capture (Priority High)

Record the failing test-name sets at pre-edit HEAD for both packages (AC-GCB-010 procedure), and store them at `.moai/reports/t867/baseline-{template,cli}.txt`. Record RED-now for both check scripts at `.moai/reports/t867/red-{residual,body}.txt`.

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

### M4 — Delete todo body and clean residual references (Priority Medium)

1. Delete `workflows/todo.md` in both trees.
2. Clean the remaining §C.1 files in both trees. Anchors like `workflows/todo.md § Standing sources` become `workflows/gtd.md § Standing sources`, CLI examples become `moai gtd <verb>`, and kept alias mentions carry `compat alias`.
3. `make agents-emit` to regenerate `manager-lead.toml`.

### M5 — Build, verify, record (Priority Medium)

1. `make build`: agents-emit-check, commands-emit-check, and `gen-catalog-hashes.go --all` (rewrites `internal/template/catalog.yaml` — commit it).
2. Run `check-residual.sh` and `check-gtd-body.sh` (PASS required); run the retargeted tests by name; run the baseline-delta package runs.
3. Write `.moai/reports/t867/verdict.md` with the `workflow.todo.enabled` finding, the §C.3 pre-existing failures as a lead finding, and the path:line inventory of historical `workflows/todo.md` citations across `.moai/specs`, `.moai/reports`, `CHANGELOG.md` and `reports/`.

## §G Anti-Patterns

- Leaving a pointer stub at `workflows/todo.md`.
- Writing `${CLAUDE_SKILL_DIR}` into the template mirror.
- Hand-editing `manager-lead.toml` or `catalog.yaml`, or leaving the regenerated catalog uncommitted.
- Documenting GTD flags absent from `--help`, or presenting `answer` as a stage.
- Making a check pass by weakening it (loosening the `compat alias` marker, deleting a test assertion, dropping a flag from the required list).
- Fixing or deleting the §C.3 pre-existing failing tests inside this SPEC.
- Editing historical SPEC, report, or CHANGELOG files to silence citations.

## §H Cross-References

- `.claude/rules/moai/workflow/kanban-dispatch.md` (queue doctrine consumer)
- `.claude/skills/moai/workflows/project/doc-generation.md` (standing-source consumer)
- `internal/cli/gtd.go`, `internal/cli/gtd_answer.go` (stage behavior source)
- `internal/template/skill_dir_token_guard_test.go`, `internal/template/catalog_tier_audit_test.go` (mirror token and catalog hash guards)
- `.moai/reports/plan-audit/SPEC-GTD-CANON-BODY-001-review-1.md` (iteration-1 audit and baseline observation)
