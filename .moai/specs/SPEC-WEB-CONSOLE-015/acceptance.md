# SPEC-WEB-CONSOLE-015 — Acceptance Criteria

Every criterion below is binary-testable. Where a criterion is satisfied by an absence, the
baseline is stated so a zero-hit grep counts as evidence rather than as a vacuous pass.

## §A Framing

**AC-WC15-001** (REQ-WC15-001) — Given the branch at merge, When
`git diff <base>..HEAD -- internal/web/events.go` is run and the `connect()` /
`startPolling()` / `es.onerror` block of `internal/web/assets/app.js` is diffed, Then no
transport-behaviour change appears: `EventSource` remains the primary channel, `POLL_MS`
remains 30000, and `startPolling()` remains reachable only from the failure and
no-`EventSource` branches.

**AC-WC15-002** (REQ-WC15-002) — Given the merged tree, When `internal/web` is grepped for
`Mutate(`, `WriteBestEffort(`, `acquireLock`, `SaveFactoryRegistry`, and `os.WriteFile` against
any `.moai/state` path, Then every hit is zero. Baseline: the same grep on the pre-change tree
also returns zero for these tokens in `internal/web`, so the criterion asserts preservation and
the test fails if a write is introduced.

## §B Axis 1 — recorded model, effort, context

**AC-WC15-010a** (REQ-WC15-010) — Given a `kanban.Record` with model and effort populated, When
it is marshalled, Then the JSON carries the new keys; and Given a record with both empty, When
it is marshalled, Then neither key appears (`omitempty`).

**AC-WC15-010b** (REQ-WC15-010, REQ-WC15-051) — Given a record JSON file **produced by the
pre-change writer** — that is, generated in the test by marshalling a `kanban.Record` with the
pre-change field set, not hand-authored — When it is unmarshalled by the post-change struct and
re-marshalled, Then the output is byte-identical to the input. The input is constrained to
writer-produced bytes so that key order and indentation match the marshaller's own output; a
hand-indented or key-reordered fixture would fail a correct implementation for reasons the
requirement does not care about.

**AC-WC15-011** (REQ-WC15-011) — Given a launch through `moai cc` in kanban mode with a
resolvable profile, When the launcher writes the session record, Then the record's model and
effort fields equal the values `EffectiveProfile` resolved for that session, asserted against
a fake profile source rather than a live launch.

**AC-WC15-012** (REQ-WC15-012) — Given a `kanban.Record` whose model field is empty, When the
role view model is built and rendered, Then the model cell carries the "not recorded" marker
and the rendered fragment contains no empty `<td>` for that column.

**AC-WC15-020** (REQ-WC15-020) — Given a statusline render for session `S` under project root
`P`, When the builder completes, Then `P/.moai/state/context-usage/S.json` exists and
`P/.moai/state/context-usage.json` does not. The second half asserts the **hard cut** and is
correct only because resolved decision G-3 (plan.md §G) chose a hard cut over a dual-write
window; under a dual-write window this criterion would fail by construction. It is written this
way by decision, not by presupposition.

**AC-WC15-021** (REQ-WC15-021) — Given the merged tree, When `internal/statusline` is scanned
for exported identifiers matching `^func Read.*ContextUsage`, Then the count is **exactly 1**;
and When `internal/web` is grepped for that identifier, Then the count is ≥ 1; and When
`internal/web` compiles, Then it declares no context-usage reader of its own. Baseline: on the
pre-change tree the exported count is **0** — `readContextUsage` (`context_usage.go:186`) is
unexported and no exported sibling exists — so both the "exactly 1" and the "≥ 1 in
internal/web" halves are new information, and the exclusivity half is mechanically pinned
rather than asserted.

**AC-WC15-022** (REQ-WC15-022) — Given the merged tree, When the grep
`grep -rn '"raw_pct"' internal/` is run — scope pinned to `internal/`, **including** `_test.go`
files — Then every hit lies inside `internal/statusline`. Baseline, measured on this tree at
the pre-change commit, is **4 hits in 4 files, of which 2 are outside `internal/statusline`**:

```
internal/statusline/context_usage.go:63       (struct field tag — stays)
internal/statusline/context_usage_test.go:150 (schema assertion — stays)
internal/cli/tokens.go:86                     (duplicate declaration — removed by REQ-025)
internal/cli/tokens_test.go:283               (its fixture — moves with it)
```

So the post-change assertion is 4 hits in 2 files, both under `internal/statusline`, and the
outside-statusline count goes **2 → 0**. A prior draft of this criterion claimed a pre-change
outside count of zero; that was false and would have mislabelled a removal test as a
preservation test.

**AC-WC15-023** (REQ-WC15-023) — Given two sessions `A` and `B` where only `A` has a
per-session context-usage file, When the console renders both role rows, Then `A` shows its own
percentage and `B` shows "not recorded" — specifically, `B` does not show `A`'s value. This is
the direct regression test for the observed last-writer-wins race.

**AC-WC15-024** (REQ-WC15-024) — Given the merged tree, When
`grep -rln "state/context-usage.json" .claude internal/template/templates` is run, Then it
returns **zero** files; and When the **four** doctrine files are grepped individually — the main
rule `.claude/rules/moai/workflow/context-window-management.md`, its detail companion
`context-window-management-detail.md`, and each one's `internal/template/templates/…` mirror —
Then each contains `context-usage/<session-id>.json`, and `diff` is empty **for each of the two
mirror pairs**. Baseline: the same `grep -rln` on the pre-change tree returns exactly those four
files, so a zero-hit result is evidence. The bare-path absence half asserts the **hard cut** and
is correct because resolved decision G-3 chose it; the four-file scope is the D6 correction —
an earlier draft named only the first pair and would have left the detail companion, which
carries the read procedure the main rule defers to, stale.

**AC-WC15-025** (REQ-WC15-025) — Two halves, both required.

*Removal half:* Given the merged tree, When `grep -rn 'context-usage' internal/cli/` is run,
Then it returns no filename constant and no record-schema declaration — specifically,
`tokensContextSnapshotFilename` and `tokensContextSnapshot` are absent from the package. Baseline:
on the pre-change tree the same grep returns `tokens.go:30` (the constant) and `tokens.go:79`
(the struct), so a zero-hit result is evidence rather than a vacuous pass.

*Positive half:* Given a state directory holding a per-session context-usage snapshot for
session `S` at the **new** path, When `moai tokens` is run for `S`, Then the emitted record
still embeds the context block with the snapshot's `raw_pct`, `stage`, and `band` values. This
half is the one nothing currently tests, and it is the half that matters: the migrated call
site's `readTokensContextSnapshot` (`tokens.go:393-397`) returns `nil` on **any** read error, so
a path the reader can no longer find produces no compile error, no runtime error, and no log
line — the block simply disappears. Only an assertion on the block's presence observes that
break.

## §C Axis 2 — todo section

**AC-WC15-030** (REQ-WC15-030) — Given a backlog file holding three items — one `queued` with
no SPEC, one `picked` with a SPEC id, one `dropped` — When the todo section is rendered, Then
**all three rows are present** (none filtered out), each carries its id and text, each carries a
**state badge element** whose rendered text is that item's state, and the picked row carries its
SPEC id. Both halves rest on resolved decision G-5 (plan.md §G), which chose the audit view —
all three states, with a badge — over a `queued`-only working view; the badge is asserted here
because it is part of that resolution and would otherwise go unobserved.

**AC-WC15-031a** (REQ-WC15-031) — Given a process whose working directory is a linked worktree
of a repository whose primary checkout holds N queued cards, When the console resolves the
queue root and loads it, Then it reads the primary checkout's file and reports N — not zero.

**AC-WC15-031b** (REQ-WC15-031) — Given the merged tree, When `internal/cli/todo.go` is read,
Then its queue-root resolution delegates to the exported `internal/kanban` function and
declares no git-common-dir resolution of its own.

**AC-WC15-031c** (REQ-WC15-031, REQ-WC15-002) — Given a process whose working directory is
**not resolvable to a git primary checkout** (no git metadata), and a project-local backlog file
present at that directory's `.moai/state/kanban/backlog.json`, and a home-based fallback root
that does **not** yet hold a queue file — the exact preconditions under which
`adoptLocalTodoQueue` performs its migration — When the **console's** exported resolver is
called and the queue is loaded, Then the project-local file is still at its original path with
its original modification time, no directory has been created under the fallback root, and no
file has been written there. And separately: When the **`moai todo`** command path is exercised
against the same preconditions, Then the adoption still occurs, unchanged.

This AC exercises the fallback branch **specifically**, because the git-resolvable path is
already pure (`resolveTodoQueueRoot` returns `filepath.Dir(dirs.CommonDir)` with no side effect)
and an AC that only exercised it would prove nothing. The write it guards is real and verified
in this tree: `fallbackTodoQueueRoot` (`internal/cli/todo.go:89-102`) calls `adoptLocalTodoQueue`
(`:115-139`), which performs `os.MkdirAll` (`:124`), `os.Rename` (`:128`), and `os.WriteFile`
(`:139`). Without the resolution/adoption split of REQ-WC15-031, a console launched in a
non-git directory would migrate the operator's backlog as a side effect of rendering a page.

**AC-WC15-032** (REQ-WC15-032) — Given a backlog file and its lock file, When every console
route is exercised, Then the lock file's modification time is unchanged and the backlog file's
content is byte-identical before and after.

**AC-WC15-033** (REQ-WC15-033) — Given a project root with no `.moai/state/kanban/backlog.json`,
and separately one whose backlog file contains malformed JSON, When the todo section is
requested, Then the response status is 200 and the body carries the empty-state marker.

**AC-WC15-034** (REQ-WC15-034) — Given a `GET /todo` response, When its markup is inspected,
Then the todo section sits inside an element carrying the **existing** `data-live="kanban"`
attribute that `refresh()` keys on; and When `internal/web/events.go` and
`internal/web/assets/app.js` are diffed against the base, Then `watchMap` (`events.go:25-32`)
and `EVENTS` (`app.js:637`) are **unchanged** — six entries each, the same six names. The second
half is the mechanical pin that the `/todo` route (resolved decision G-4) did not change the
event vocabulary; it is asserted rather than assumed because the route decision was taken
against a prediction that it would force one. Grounding for why reuse is correct and a seventh
event name would be wrong: spec.md §C.6.

**AC-WC15-035** (REQ-WC15-030, REQ-WC15-050) — Given the merged tree, When `GET /todo` is
requested, Then the response is 200 and its rail carries **six** nav rows, the sixth linking to
`/todo` and marked `aria-current` on this route; and When `internal/web/icons.templ`'s `iconAt`
switch is read, Then it carries a `todo` case (a missing case renders a blank glyph, since
`navRow` calls `@iconAt(id, 16)` with the nav id); and When the i18n governance test runs, Then
`nav.todo` is present in all four locale maps of `internal/web/assets/i18n.js` with no allowlist
entry added. Baseline: the pre-change rail carries **five** rows
(`shell.templ` `rail()`: overview, kanban, specs, monitor, settings) and no `todo` icon case, so
each half is new information. This AC exists because the G-4 route decision brought the nav
surface into scope (spec.md §C.6).

## §D Axis 3 — per-lane factory progress

**AC-WC15-040** (REQ-WC15-040, REQ-WC15-041) — Given a lane launch labelled `lane-3`, When the
record is written and re-read, Then its role reads as the lane role and its lane-number field
reads 3. Baseline: on the pre-change tree the same launch produces a record with role `lane`
and no recoverable lane number, so a passing assertion here is new information.

**AC-WC15-042** (REQ-WC15-042) — Given a Class B card with no SPEC, When its lane's record is
written and re-read, Then the card-identifier field holds the card id and the SPEC field is
empty.

**AC-WC15-043** (REQ-WC15-043) — Given a factory registry mapping `lane-2` to PID `P`, an
active-sessions entry with PID `P` and session id `S`, and a kanban record for `S`, When the
factory view model is built, Then the `lane-2` row carries `S`'s record values; and When the
merged tree is grepped, Then no new file is created under `.moai/state/` by this join.

**AC-WC15-044** (REQ-WC15-044) — Given two registered lanes with complete join data, When the
factory section is rendered, Then each row contains the lane number, the card id, the SPEC id
where present, the session state, and the stage.

**AC-WC15-045** (REQ-WC15-045) — Given a lane whose stage came from `estimateStage` with
`estimated == true`, When its row is rendered, Then the row carries the estimated marker; and
Given `estimated == false`, Then it does not.

**AC-WC15-046** (REQ-WC15-046) — Given a project root with no `.moai/state/factory/workers.json`,
and separately one whose registry is malformed, When the factory section is requested, Then the
response status is 200 and the section renders zero lanes.

## §E Cross-cutting

**AC-WC15-050** (REQ-WC15-050) — Given the merged tree, When the existing i18n governance test
in `internal/web/i18n_governance_test.go` runs, Then every key added by this SPEC is present in
all four locale maps and the test passes with no allowlist entry added.

**AC-WC15-051** (REQ-WC15-051) — Covered by AC-WC15-010b for the write path, and additionally:
Given a records directory holding one pre-change record and one post-change record, When the
console renders the chain, Then both rows render, the pre-change row showing "not recorded" for
each new field.

**AC-WC15-052** (REQ-WC15-020, REQ-WC15-002) — Given a statusline render whose session id
contains a path separator (`a/b`), a parent traversal (`../x`), an absolute prefix (`/etc/x`),
or is empty, When the per-session context-usage write is attempted under project root `P`, Then
no file is created outside `P/.moai/state/context-usage/`, the write is refused rather than
redirected, and the render itself still completes (the persistence step is best-effort and must
not fail the statusline). Promoted from a §F edge-case note because REQ-WC15-020 turns an
externally-shaped value — a session id the statusline receives rather than mints — into a
filesystem path component; that is a write boundary, and a write boundary taking outside input
earns an assertion rather than prose.

## §F Edge cases

- A registered factory lane whose PID matches no active-sessions entry — the row renders with
  the lane number and an explicit unresolved marker, not a blank row and not a dropped row.
- Two lanes whose registry entries carry the same PID (a stale entry not yet pruned) — the join
  must not silently attribute one session's record to both lanes.
- A backlog item whose `spec_id` is a non-nil pointer to an empty string — distinguishable in
  the view from a nil pointer.

## §G Definition of Done

- [ ] Every REQ in spec.md §B maps to at least one AC above, and every AC maps to a REQ.
- [ ] Affected-package tests pass (`internal/web`, `internal/kanban`, `internal/cli`,
      `internal/statusline`); the full-suite verdict is read from CI, not from a local run.
- [ ] `go vet ./...` and `golangci-lint run` clean.
- [ ] `make build` run after **both** doctrine mirror-pair edits, and both mirror diffs empty.
- [ ] Coverage on changed packages does not regress from the pre-change baseline.
- [x] The six decisions in plan.md §G are resolved and every clarification marker is removed.
      Verified by grepping this SPEC directory for the clarification-marker token: zero hits.
