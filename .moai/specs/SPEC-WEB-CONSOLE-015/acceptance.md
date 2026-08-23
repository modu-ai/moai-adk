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

**AC-WC15-010b** (REQ-WC15-010, REQ-WC15-051) — Given a record JSON file written before this
change (no model, effort, lane, or card key), When it is unmarshalled and re-marshalled, Then
the output is byte-identical to the input.

**AC-WC15-011** (REQ-WC15-011) — Given a launch through `moai cc` in kanban mode with a
resolvable profile, When the launcher writes the session record, Then the record's model and
effort fields equal the values `EffectiveProfile` resolved for that session, asserted against
a fake profile source rather than a live launch.

**AC-WC15-012** (REQ-WC15-012) — Given a `kanban.Record` whose model field is empty, When the
role view model is built and rendered, Then the model cell carries the "not recorded" marker
and the rendered fragment contains no empty `<td>` for that column.

**AC-WC15-020** (REQ-WC15-020) — Given a statusline render for session `S` under project root
`P`, When the builder completes, Then `P/.moai/state/context-usage/S.json` exists and
`P/.moai/state/context-usage.json` does not.

**AC-WC15-021** (REQ-WC15-021) — Given the merged tree, When `internal/web` is compiled, Then
it references exactly one exported `internal/statusline` context-usage reader, and
`grep -c` for that identifier in `internal/web` is ≥ 1.

**AC-WC15-022** (REQ-WC15-022) — Given the merged tree, When every package outside
`internal/statusline` is grepped for a struct declaring the field tag `"raw_pct"`, Then the
count is zero. Baseline: the pre-change tree also returns zero outside `internal/statusline`,
so this asserts the split did not spawn a copy.

**AC-WC15-023** (REQ-WC15-023) — Given two sessions `A` and `B` where only `A` has a
per-session context-usage file, When the console renders both role rows, Then `A` shows its own
percentage and `B` shows "not recorded" — specifically, `B` does not show `A`'s value. This is
the direct regression test for the observed last-writer-wins race.

**AC-WC15-024** (REQ-WC15-024) — Given the merged tree, When
`.claude/rules/moai/workflow/context-window-management.md` and its template mirror are grepped,
Then both contain `context-usage/<session-id>.json`, neither contains the bare
`.moai/state/context-usage.json`, and `diff` between the two files is empty.

## §C Axis 2 — todo section

**AC-WC15-030** (REQ-WC15-030) — Given a backlog file holding three items — one `queued` with
no SPEC, one `picked` with a SPEC id, one `dropped` — When the todo section is rendered, Then
the fragment contains each item's id and text, its state, and the SPEC id for the picked item.

**AC-WC15-031a** (REQ-WC15-031) — Given a process whose working directory is a linked worktree
of a repository whose primary checkout holds N queued cards, When the console resolves the
queue root and loads it, Then it reads the primary checkout's file and reports N — not zero.

**AC-WC15-031b** (REQ-WC15-031) — Given the merged tree, When `internal/cli/todo.go` is read,
Then its queue-root resolution delegates to the exported `internal/kanban` function and
declares no git-common-dir resolution of its own.

**AC-WC15-032** (REQ-WC15-032) — Given a backlog file and its lock file, When every console
route is exercised, Then the lock file's modification time is unchanged and the backlog file's
content is byte-identical before and after.

**AC-WC15-033** (REQ-WC15-033) — Given a project root with no `.moai/state/kanban/backlog.json`,
and separately one whose backlog file contains malformed JSON, When the todo section is
requested, Then the response status is 200 and the body carries the empty-state marker.

**AC-WC15-034** (REQ-WC15-034) — Given the rendered todo section, When its markup is inspected,
Then it sits inside an element carrying the `data-live="kanban"` attribute the existing
`refresh()` path keys on.

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

## §F Edge cases

- A registered factory lane whose PID matches no active-sessions entry — the row renders with
  the lane number and an explicit unresolved marker, not a blank row and not a dropped row.
- A session id containing a path separator or `..` — the per-session context-usage path must
  reject it rather than resolve outside the state directory.
- Two lanes whose registry entries carry the same PID (a stale entry not yet pruned) — the join
  must not silently attribute one session's record to both lanes.
- A backlog item whose `spec_id` is a non-nil pointer to an empty string — distinguishable in
  the view from a nil pointer.

## §G Definition of Done

- [ ] Every REQ in spec.md §B maps to at least one AC above, and every AC maps to a REQ.
- [ ] Affected-package tests pass (`internal/web`, `internal/kanban`, `internal/cli`,
      `internal/statusline`); the full-suite verdict is read from CI, not from a local run.
- [ ] `go vet ./...` and `golangci-lint run` clean.
- [ ] `make build` run after the doctrine-rule mirror edit, and the mirror diff is empty.
- [ ] Coverage on changed packages does not regress from the pre-change baseline.
- [ ] The six open decisions in plan.md §G are resolved and the
      `[NEEDS CLARIFICATION]` markers removed before run-phase entry.
