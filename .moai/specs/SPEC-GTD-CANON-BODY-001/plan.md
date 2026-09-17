# Plan — SPEC-GTD-CANON-BODY-001 (card t867)

## §A Context

Tier M, Class C. Body migration from `workflows/todo.md` to `workflows/gtd.md`, deletion of the todo body, residual-reference cleanup across 21 files, and retargeting of the Go tests that pin the deleted path. Development mode per `quality.yaml`; this is a doc-plus-test change, so the cycle is characterization-first: retargeted tests must fail against the pre-migration tree (path absent) and pass after it.

## §B Known Issues / Decisions

- **Decision (operator):** delete `workflows/todo.md` in both trees, no stub.
- **Decision (operator):** retarget 4 Go tests; comment-only Go citations are implementer discretion.
- **Plan-time finding:** `internal/cli/todo.go` ~line 211 names `workflows/todo.md` inside the Cobra Long help string — user-visible text, not a comment. Treated as required (REQ-GCB-008).
- **Plan-time finding:** `internal/template/gtd_canonical_surface_test.go` does not read `workflows/todo.md`; it pins `commands/moai/todo.md` and `.agents/skills/moai-todo/SKILL.md` as thin compat paths. It must stay green unchanged.
- **Plan-time finding:** `internal/template/catalog.yaml`, `Makefile`, `internal/template/skill_mirror*.go`, and `internal/template/commandemit/*` contain no enumeration of `workflows/todo.md` (commandemit's `todo.md` mentions refer to `commands/moai/todo.md`, which is retained). The deploy-time `.agents/skills/moai` mirror walks the directory, so a deleted file simply stops being mirrored — the run phase confirms this with the scoped test run.
- **Finding to record:** `workflow.todo.enabled` / `TodoEnabled()` keep the todo name (REQ-GCB-010).
- **Historical citations:** plan-time `grep -rlE 'workflows/todo\.md' .moai/specs .moai/reports` returned 49 files; the run phase enumerates path:line into the verdict.

## §C Pre-flight

- `git -C <wt> branch --show-current` → `WT-gtd-canon`
- `cmp` the local and template `workflows/todo.md` → identical (measured at plan time)
- Build the binary once and capture `moai gtd --help` plus `moai gtd <stage> --help` for the six stage verbs as the documentation source for REQ-GCB-002

## §D Constraints

- Template-First: every doc edit lands in the local file and its `internal/template/templates/` mirror in the same commit; `make build` afterwards.
- `manager-lead.toml` is regenerated only via `make agents-emit`.
- No local `go test ./...`; scoped packages only.
- Template neutrality (no SPEC IDs, card ids, dates, or commit SHAs inside `internal/template/templates/**`).

## §F Milestones (ordered by decision-reversibility)

### M1 — gtd.md canonical body (Priority High)

Decisions most likely to be revised by review: the section structure of the merged body and how the five GTD stages sit next to the queue verbs.

1. Rewrite `workflows/gtd.md` (local + mirror): carry every todo.md section; retarget all command examples to `moai gtd` / `/moai gtd`; add a GTD stages section (capture → clarify → organize → reflect → engage, plus `answer`) sourced strictly from the captured `--help` output, including the capture/queue separation and the approved-engage publish rule; add one compat-alias paragraph naming `/moai todo` and `moai todo`.
2. Keep every table row the retargeted tests read: `| \`moai gtd landed ` / `| \`moai gtd pr ` rows with the "carries <word> tab-separated columns" prose, a `moai gtd history` mention, and the JSON fence whose keys `TestTodoListJSONShapeMatchesDoc` compares.

### M2 — SKILL.md gtd section (Priority High)

1. Replace `### todo - Backlog Queue` with a gtd section (local + mirror) pointing at `${CLAUDE_SKILL_DIR}/workflows/gtd.md`, verbs in `moai gtd` form, compat aliases documented, enablement paragraph keeping the `workflow.todo.enabled` key name.
2. Retarget the Intent Router backlog-language exemplar to route to **gtd**.

### M3 — Retarget Go consumers (Priority High)

1. `internal/cli/todo_skill_doc_test.go`: path → `workflows/gtd.md`, count `moai gtd history`.
2. `internal/cli/todo_landed_doc_test.go`: path → `workflows/gtd.md`, row prefixes → `moai gtd landed` / `moai gtd pr`.
3. `internal/cli/doc_json_shape_test.go`: doc surfaces → `workflows/gtd.md`.
4. `internal/template/backlog_json_disclosure_mirror_test.go`: mirrored file list → `workflows/gtd.md`.
5. `internal/cli/todo.go` Long help: `workflows/todo.md` → `workflows/gtd.md`.
6. Discretionary comments: `todo_drop.go:16`, `todo_edit_move.go:8,100`, `todo_test.go:665`, `backlog_json_disclosure_mirror_test.go:5`.

### M4 — Delete todo body and clean residual references (Priority Medium)

1. Delete `.claude/skills/moai/workflows/todo.md` and its template mirror.
2. Clean the remaining files of §C.1 (local + mirror): `manager-lead.md`, `kanban-dispatch.md`, `kanban-dispatch-detail.md`, `main-checkout-branch-guard-detail.md`, `moai-kanban-foreman/SKILL.md`, `project/doc-generation.md`, `todo-queue-storage.md`. Section anchors such as `workflows/todo.md § Standing sources` become `workflows/gtd.md § Standing sources`; CLI examples become `moai gtd <verb>`; any intentionally kept `moai todo` mention must satisfy the allowed-survivor rule.
3. `make agents-emit` to regenerate `manager-lead.toml`.

### M5 — Build, verify, record (Priority Medium)

1. `make build` (runs agents-emit-check, commands-emit-check).
2. `go vet ./internal/cli/... ./internal/template/...`; `go test ./internal/cli/... ./internal/template/...`.
3. Residual grep and compat checks per acceptance.md.
4. Write `.moai/reports/t867/verdict.md`: the `workflow.todo.enabled` finding and the path:line list of historical `workflows/todo.md` citations.

## §G Anti-Patterns

- Leaving a pointer stub at `workflows/todo.md` (contradicts the operator decision).
- Hand-editing `manager-lead.toml`.
- Documenting GTD flags or behavior not present in `--help`.
- Making the grep pass by deleting a test's assertion instead of retargeting it.
- Editing historical SPEC or report files to silence citations.

## §H Cross-References

- `.claude/rules/moai/workflow/kanban-dispatch.md` (queue doctrine consumer)
- `.claude/skills/moai/workflows/project/doc-generation.md` (standing-source consumer)
- `internal/cli/gtd.go`, `internal/cli/gtd_answer.go` (stage behavior source)
- `internal/cli/gtd_compat_test.go` (compat parity tests)
