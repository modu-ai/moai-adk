# SPEC-WEB-TODO-QUEUE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored (Tier M set complete): `spec.md`, `plan.md`, `acceptance.md`, plus this
  `progress.md`.
- Tier: M (justification in `plan.md` §B). Threshold 0.80.
- SPEC ID regex check executed as Bash; output `PASS`.
- Budget: 8 requirements, 11 acceptance criteria — both under the Tier M ceilings of 16 / 16.
- Carved out of `SPEC-WEB-CONSOLE-015` per `.moai/reports/t207/spec-split-design.md` §4.
  Inherits resolved decisions G-4 and G-5; adds D-2 (read-through).
- Findings answered from `.moai/reports/t207/plan-audit-iter2-independent.md`: F3 (D-2 +
  AC-WTQ-007), F6 (REQ-WTQ-007 conditional wording), F11 (AC-WTQ-004 mechanical form), F13
  (GEARS form), F9 (AC-WTQ-001 executable scan form).
- Status: `draft`. Awaiting plan-audit.

## §E.2 Run-phase Evidence

Every row's evidence was captured in this run, in this tree
(`.claude/worktrees/t207`), at HEAD `acd529f1d` unless a row states otherwise.
Run-phase base: `f276b9742`.

### Milestones

| M | Commit | Content |
|---|---|---|
| M1 | `063a0a5c7` | queue-root resolution relocated to `internal/kanban`, split pure / adopting; `kanban.HomeDirFn` seam; read-through (D-2) |
| M2 | `d471057de` | `/todo` route, sixth nav row, `iconAt` case, four-locale strings |
| M3 | `acd529f1d` | read-only section: view model, rows, state badges, empty state |

### AC matrix

| AC | Status | Command | Actual output |
|---|---|---|---|
| AC-WTQ-001 | PASS | `grep -rnE 'Mutate\(\|acquireLock\|os\.WriteFile\|os\.MkdirAll\|os\.Rename' internal/web --include='*.go' \| grep -v '_test\.go' \| grep -vE '^[^:]+:[0-9]+:[[:space:]]*//'` + `go test -run TestConsoleRoutesLeaveBacklogUntouched ./internal/web/` | exactly the two allowlisted lines (`profile_crud.go:38`, `:64`) and nothing else; `--- PASS: TestConsoleRoutesLeaveBacklogUntouched (0.26s)` — backlog bytes identical, `backlog.lock` mtime unchanged after all six routes were exercised |
| AC-WTQ-002 | PASS | `go test -run 'TestTodoRouteServesPage\|TestTodoNavRowIsSixthAndCurrent\|TestTodoIconCaseExists' ./internal/web/`; `awk '/templ iconAt/,/^}$/' internal/web/icons.templ \| grep -c 'case "todo"'` | all three PASS (200; six rows in order `/ /kanban /specs /monitor /settings /todo` with `href="/todo" aria-current="page"`); `awk` prints `1` (baseline `0`) |
| AC-WTQ-003 | PASS | `go test -run TestTodoSectionListsAllThreeStates ./internal/web/` | `--- PASS (0.14s)` — 3 rows, none filtered; ids/texts present; `data-todo-state="queued\|picked\|dropped"` each rendering its state as its text; the picked row carries `SPEC-EXAMPLE-001` |
| AC-WTQ-004 | PASS | `grep -c 'ResolveGitDirs' internal/cli/todo.go` | `0` (baseline `1`) |
| AC-WTQ-005 | PASS | `go test -run TestTodoSectionFromWorktreeReadsPrimaryQueue ./internal/web/`; `go test -run TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary ./internal/kanban/` | both PASS — a console served from a linked worktree renders the primary's 3 rows, not zero |
| AC-WTQ-006 | PASS | `go test -run TestResolveTodoQueueRoot_PureFallbackWritesNothing ./internal/kanban/` | `--- PASS (0.04s)` — local file at its original path with its original mtime; fallback root absent (`IsNotExist`). Measured listing in §E.2 “E9” below |
| AC-WTQ-007 | PASS | `go test -run TestTodoSectionReadsThroughToProjectLocalQueue ./internal/web/`; `go test -run 'TestResolveTodoQueueRoot_ReadThroughToProjectLocal\|TestResolveTodoQueueRootAdopting_AdoptsLocalQueue' ./internal/kanban/` | all PASS — the section lists the project-local queue's 3 items while the disk stays untouched; the `moai todo` path against the same fixture reports the same 3 items (after adopting) |
| AC-WTQ-008 | PASS | `go test -run 'TestResolveTodoQueueRootAdopting_AdoptsLocalQueue' ./internal/kanban/`; `go test -timeout 900s -run 'TestTodoQueue_FallbackAdoptsExistingLocalQueue' ./internal/cli/` | both PASS — local file gone from its original path, fallback root holds the queue, item count and states unchanged |
| AC-WTQ-009 | PASS | `go test -run TestTodoSectionEmptyStates ./internal/web/` | `--- PASS (0.40s)` over three sub-cases (absent / empty / malformed JSON) — 200 with the `data-todo-empty` marker and no rows in all three |
| AC-WTQ-010 | PASS | `go test -run TestTodoSectionCarriesExistingKanbanMarker ./internal/web/`; `git diff f276b9742..HEAD -- internal/web/events.go internal/web/assets/app.js` | PASS (the section sits inside `data-live="kanban"`); the diff is **empty** — `watchMap` still 6 entries, `EVENTS` still `["spec","session","goal","verify","kanban","config"]` |
| AC-WTQ-011 | PASS | `grep -c '"nav\.todo"' internal/web/assets/i18n.js`; `go test -run TestI18n ./internal/web/` | `4` (baseline `0`); `ok github.com/modu-ai/moai-adk/internal/web` — no allowlist entry added |

### Storage-migration read seam (for whoever migrates the queue)

The backlog queue's storage is under operator review (JSON file today; a sqlite
store under `~/.moai` keyed by repo root was being considered, undecided at the
time of writing). The console's read path is isolated in **one function**:
`readTodoQueue` in `internal/web/todo_queue_read.go` — the only file in
`internal/web` that names a backlog-store symbol
(`kanban.ResolveTodoQueueRoot`, `kanban.NewBacklogStore`,
`kanban.BacklogPathForRoot`, `kanban.BacklogItem`). `buildTodo`
(`internal/web/todo_view.go`) calls it and touches no store symbol itself, and
`TestBacklogReadSeamIsSingleFile` (`internal/web/todo_queue_read_test.go`)
asserts the single-file property mechanically. A storage change lands in that
one function. No interface, factory, or plugin point was introduced, and no
sqlite code was written.

### E2 — cross-platform build

`go build ./...` → exit 0. `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.

### E3 — test suites and coverage

- `go test -count=1 -cover ./internal/web/... ./internal/kanban/...` → `ok internal/web 3.214s coverage: 65.8%`, `ok internal/kanban 11.915s coverage: 84.5%`
- `go test -count=1 -cover -timeout 900s ./internal/cli/` → `ok internal/cli 479.162s coverage: 78.8%`; all `./internal/cli/...` sub-packages `ok`

### E4 — lint

`golangci-lint run ./internal/web/... ./internal/kanban/... ./internal/cli/` → `0 issues.` No NEW findings (and no pre-existing ones to distinguish).

### E7 — RED evidence (captured before GREEN)

- M1: `go test ./internal/kanban/ -run TestResolveTodoQueueRoot` → `undefined: HomeDirFn`, `undefined: ResolveTodoQueueRoot`, `undefined: TodoQueueProjectKey` … `FAIL [build failed]`
- M2: `GET /todo status = 404, want 200`; `rail carries 0 navigation rows ([]), want 6`; `iconAt carries 0 'case "todo"' arms, want 1`; `i18n.js carries 0 "nav.todo" entries, want 4`
- M3: `no state badge for "queued"/"picked"/"dropped"`; `section renders 0 rows, want 3`; `console served from the worktree renders 0 rows, want the primary's 3`; `read-through renders 0 rows, want 3`

### E9 — the fallback pair, measured

A temporary probe (run, captured, then deleted) printed the disk before and after
the console resolved and rendered under AC-WTQ-006/007's preconditions:

```
[before] home tree under <tmp>/001: []
[before] local queue <tmp>/002/.moai/state/kanban/backlog.json mtime=2026-08-24T21:30:36.024418949+09:00 size=349
[after]  home tree under <tmp>/001: []
[after]  local queue <tmp>/002/.moai/state/kanban/backlog.json mtime=2026-08-24T21:30:36.024418949+09:00 size=349
[render] rows=3 empty-state=false states=[true true true]
```

Both halves together: nothing was written, and the section rendered the
project-local queue's three items while nothing was written.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-24
run_commit_sha: acd529f1d
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: exit 0
  windows_amd64: exit 0
total_run_phase_files: 12
m1_to_mN_commit_strategy: one commit per milestone (M1 063a0a5c7, M2 d471057de, M3 acd529f1d)
pushed: false   # unpushed by operator decision; this worktree holds the only copy
```

Run-phase findings reported to the orchestrator (not acted on):

1. `kanban.NewBacklogStore` removes a superseded legacy lock artifact
   (`backlog.json.lock`) beside the queue file as a constructor side effect
   (`internal/kanban/backlog_store.go:241`). The console reaches it through the
   render path prescribed by plan.md §D M3. It touches neither `backlog.json`
   nor the live `backlog.lock`, so AC-WTQ-001 passes as written, but
   REQ-WTQ-001's prose ("no write ... against the backlog queue") is arguably
   broader than the criterion. Left as found — the store belongs to
   SPEC-KANBAN-TODO-CLI-001.
2. Commit `977a8dc15` (a sibling SPEC's `progress.md`, another session) landed on
   this branch between this run's pre-flight read of `f276b9742` and its M1
   commit. It touches no file in this SPEC's scope; the milestones stack on top
   of it.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-24
sync_commit_sha: pending-backfill-SPEC-WEB-TODO-QUEUE-001
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-WEB-TODO-QUEUE-001' CHANGELOG.md → 0 (pipeline rc 1, a genuine zero-match; no duplicate, emission permitted)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 11; the CHANGELOG entry states 11"
b12_self_test_c: "every path named in the CHANGELOG entry verified present with ls before commit (13 paths, all present)"
changelog_entry_position: "[Unreleased] → ### Changed, appended immediately after the SPEC-SESSION-TELEMETRY-001 entry"
frontmatter_status_transitions:
  spec.md: "in-progress → completed (updated: already 2026-08-24, the sync date — unchanged)"
  plan.md: "n/a — carries no frontmatter block"
  acceptance.md: "n/a — carries no frontmatter block"
  progress.md: "n/a — carries no frontmatter block"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy that its own sync tests"
```

### Storage-migration seam — where a queue-storage change lands

Carried forward from §E.2 into the audit-ready signal, because whoever migrates the queue reads
this section and not the run record.

| What | Where |
|---|---|
| The seam | `readTodoQueue` — `internal/web/todo_queue_read.go` |
| Its consumer | `buildTodo` — `internal/web/todo_view.go` (touches no store symbol itself) |
| The pin | `TestBacklogReadSeamIsSingleFile` — `internal/web/todo_queue_read_test.go` |
| Store symbols named | `kanban.ResolveTodoQueueRoot`, `kanban.NewBacklogStore`, `kanban.BacklogPathForRoot`, `kanban.BacklogItem` — in that one file and nowhere else in `internal/web` |

The queue's storage was under operator review when this SPEC landed (a JSON file today; a sqlite
store under `~/.moai` keyed by repo root was being considered, undecided). **No sqlite code was
written and no abstraction layer was introduced** — no interface, no factory, no plugin point. A
single concrete reader needs no seam beyond its own signature, so the whole preparation for a
storage change is the containment: change `readTodoQueue` and nothing else in `internal/web` has
to move, and the test fails if a later edit reaches around it.

The isolation was a separate commit — `aef1b51f3`, `refactor(SPEC-WEB-TODO-QUEUE-001): isolate
the console's backlog read behind one function (t207)` — landing after the M3 render commit, so
the property is attributable to that commit rather than assumed across the milestone set.

### Documentation review — what was searched, found and changed

Searched: `docs-site/content/{en,ko,ja,zh}/` and the four READMEs, for the console's route table,
its area count, its live-indicator count, the mermaid rail/read-only node labels, and the
`moai web` command-table row. Commands and their results are quoted in the sync report; the
finding is that the published pages **did** describe the nav set, and did so as five areas.

Changed, at sync (12 files):

| File (× 4 locales) | What was falsified |
|---|---|
| `docs-site/content/<loc>/advanced/moai-web-console.md` | `description` five areas; the callout's "four observation areas"; "the rail stacks the five areas"; the 5-row route table; "the four observation areas show a live indicator"; both mermaid node labels (`Rail`, `Read`) |
| `docs-site/content/<loc>/cli-reference/web.md` | "the four screens are read-only"; the route table missing `/todo` |
| `README{,.ko,.ja,.zh}.md` | the "Overview·Specs·Monitor·Settings" enumeration; "five screens"; the `moai web` command-table row |

Added, not merely corrected: a per-locale section describing the screen (all three states listed,
read-only, primary-checkout resolution from a worktree, the empty state at 200). Nothing else was
written — no new page, no new nav entry, no invented behaviour.

`hugo --logLevel warn` in `docs-site/` → exit 0, no warnings (184 KO / 182 EN / 182 JA / 182 ZH
pages). This is a build signal only; no page was rendered and read in a browser.

### Run-phase finding 1 — carried forward, not closed

`kanban.NewBacklogStore` removes a superseded legacy lock artifact (`backlog.json.lock`) as a
constructor side effect (`internal/kanban/backlog_store.go`), and the console reaches it through
the render path. AC-WTQ-001 passes as written — the observation is that `backlog.json`'s bytes and
`backlog.lock`'s mtime are both unchanged after all six routes are exercised — but REQ-WTQ-001's
prose ("no write … against the backlog queue") is broader than the criterion it is measured by.
The store belongs to `SPEC-KANBAN-TODO-CLI-001`; nothing was changed here. Recorded as a known
imprecision between a requirement's prose and its criterion, not as a defect this SPEC closes.

### Gaps — what sync did NOT observe

- **No coverage-regression figure.** The pre-change package rates for `internal/web` and
  `internal/kanban` were never measured on this tree, so no delta exists and none is stated. The
  run-phase absolute rates (65.8% / 84.5%) are absolute rates, not deltas.
- **No test was re-run at sync.** Every AC result in §E.2 is the run-phase measurement at
  `acd529f1d`; sync re-ran nothing (the lead's instruction, and `internal/cli` alone takes
  400-900s). The `aef1b51f3` refactor landed after `acd529f1d`, so the §E.2 AC evidence predates
  it; what pins the post-refactor state is `TestBacklogReadSeamIsSingleFile`, itself measured in
  that commit and not re-run here.
- **No CI verdict.** The branch is unpushed by operator decision; no CI has ever run on it.
- **The docs were not read in a rendered browser.** The build is warning-free; layout and the
  mermaid render were not visually confirmed.
