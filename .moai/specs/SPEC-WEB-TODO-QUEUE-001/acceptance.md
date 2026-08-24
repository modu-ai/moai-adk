# SPEC-WEB-TODO-QUEUE-001 — Acceptance Criteria

Every criterion below names a command and an expected result. Where a criterion is satisfied by an
absence, its **measured pre-change baseline** is stated, so a zero-hit scan counts as evidence
rather than as a vacuous pass. Baselines were measured in this tree at `dfbf828a6`; each is quoted
with the command that produced it.

Eleven criteria, one or more per requirement, every requirement covered.

## §A Read-only posture

**AC-WTQ-001** (REQ-WTQ-001) — Given the merged tree, When

```bash
grep -rnE 'Mutate\(|acquireLock' internal/web --include='*.go' \
  | grep -v '_test\.go' | grep -vE '^[^:]+:[0-9]+:[[:space:]]*//' | wc -l
```

is run, Then the count is `0`; and Given a project root holding a backlog file and its lock file,
When every console route is exercised in one test run, Then the backlog file's bytes are identical
before and after and the lock file's modification time is unchanged.

Baseline: the same command on the pre-change tree prints `0`, so the grep half asserts
**preservation** and fails only if a write is introduced. The unqualified form of this scan
(`grep -rnE 'Mutate\(|acquireLock|os\.WriteFile' internal/web`) prints `23` in this tree — every
hit a `_test.go` file or an anti-pattern comment — which is why the executable form above is
scoped to production Go with comment lines excluded rather than left as prose about `.moai/state`
paths. The runtime half is the non-vacuous half: it observes behaviour a grep cannot.

## §B Route, navigation, and content

**AC-WTQ-002** (REQ-WTQ-002, REQ-WTQ-008) — Given the merged tree, When `GET /todo` is requested,
Then the response status is 200 and its rail carries **six** navigation rows, the sixth linking to
`/todo` and carrying `aria-current` on this route; and When

```bash
awk '/templ iconAt/,/^}$/' internal/web/icons.templ | grep -c 'case "todo"'
```

is run, Then the count is `1`.

Baseline, both halves measured on the pre-change tree:
`sed -n '/templ rail/,/^}/p' internal/web/shell.templ | grep -c '@navRow'` prints `5`
(overview, kanban, specs, monitor, settings at `shell.templ:130-134`), and the `awk` command above
prints `0`. Each half is therefore new information.

**AC-WTQ-003** (REQ-WTQ-003) — Given a backlog file holding three items — one `queued` with no
SPEC id, one `picked` with a SPEC id, one `dropped` — When the todo section is rendered, Then all
three rows are present with none filtered out, each row carries its identifier and its text, each
row carries a state-badge element whose rendered text is that item's state, and the picked row
carries its SPEC id.

The badge is asserted because it is half of resolved decision G-5 (plan.md §F): a `queued`-only
list cannot answer "where did card X go", and an unasserted badge leaves the other half of the
decision unobserved.

**AC-WTQ-011** (REQ-WTQ-008) — Given the merged tree, When

```bash
grep -c '"nav\.todo"' internal/web/assets/i18n.js
```

is run, Then the count is `4` — one per locale map (`en`, `ko`, `ja`, `zh`); and When
`go test -run TestI18n ./internal/web/...` runs the governance test
(`internal/web/i18n_governance_test.go`), Then it passes with no allowlist entry added for any
string this SPEC introduces.

Baseline: the same `grep -c` prints `0` on the pre-change tree.

## §C Queue-root resolution

**AC-WTQ-004** (REQ-WTQ-004) — Given the merged tree, When

```bash
grep -c 'ResolveGitDirs' internal/cli/todo.go
```

is run, Then the count is `0`, the resolution having moved to the shared package that
`internal/cli` and `internal/web` both import.

Baseline: the same command prints `1` on the pre-change tree (`todo.go:68`). This is the mechanical
form of "delegates rather than keeping its own copy" — a criterion phrased as "when the file is
read" would be a human judgement, not a check.

**AC-WTQ-005** (REQ-WTQ-004) — Given a process whose working directory is a linked worktree of a
repository whose primary checkout holds N queued cards (N > 0), When the console resolves the queue
root and loads it, Then it reads the primary checkout's file and reports N, not zero.

**AC-WTQ-006** (REQ-WTQ-004, REQ-WTQ-001) — Given a working directory that git cannot resolve to a
primary checkout, a project-local backlog file at that directory's `.moai/state/kanban/backlog.json`,
and a home-based fallback root holding no queue file — the exact preconditions under which
`adoptLocalTodoQueue` migrates (`internal/cli/todo.go:118-123`) — When the entry point the console
consumes is called and the queue is loaded, Then the project-local file is still at its original
path with its original modification time, no directory has been created under the fallback root,
and no file has been written there.

This criterion exercises the fallback branch **specifically**. The git-resolvable branch is already
pure (`todo.go:68-70` returns `filepath.Dir(dirs.CommonDir)` with no side effect), so a criterion
that only exercised it would prove nothing. The write it guards is real and measured:
`fallbackTodoQueueRoot` (`:89-102`) calls `adoptLocalTodoQueue` (`:115-139`), which performs
`os.MkdirAll` (`:124`), `os.Rename` (`:128`), and `os.WriteFile` (`:139`).

**AC-WTQ-007** (REQ-WTQ-005) — Given the same preconditions as AC-WTQ-006 with the project-local
file holding N items (N > 0), When the todo section is rendered, Then it lists exactly those N
items — not an empty queue; and When `moai todo` is run against the same preconditions on a fresh
copy of that fixture, Then it reports the same N items.

This is the read-through half of decision D-2 (spec.md §A.3). AC-WTQ-006 asserts the disk is
untouched; this criterion asserts what is rendered while it is untouched. Without both, an
implementation can satisfy every other criterion and still ship the divergence the SPEC exists to
close.

**AC-WTQ-008** (REQ-WTQ-004) — Given the same preconditions as AC-WTQ-006, When the `moai todo`
command path is exercised, Then the adoption still occurs: the project-local file is gone from its
original path, the fallback root holds the queue file, and its item count and states are unchanged
from before.

Baseline: this is today's behaviour on the pre-change tree, so the criterion asserts that the
relocation narrowed the adoption's **call site** and not its behaviour.

## §D Robustness and live refresh

**AC-WTQ-009** (REQ-WTQ-006) — Given a project root with no `.moai/state/kanban/backlog.json`, and
separately one whose backlog file is empty, and separately one whose backlog file contains
malformed JSON, When `GET /todo` is requested for each, Then the response status is 200 in all
three cases and the body carries the empty-state marker.

**AC-WTQ-010** (REQ-WTQ-007) — Given a `GET /todo` response, When its markup is inspected, Then the
todo section sits inside an element carrying the **existing** `data-live="kanban"` attribute that
`refresh()` keys on (`app.js:644`); and When

```bash
git diff <base>..HEAD -- internal/web/events.go internal/web/assets/app.js
```

is run, Then `watchMap` (`events.go:25-32`) and `EVENTS` (`app.js:637`) are unchanged — six entries
each, the same six names.

Baseline, measured on the pre-change tree:
`sed -n '/^var watchMap/,/^}/p' internal/web/events.go | grep -c '":'` prints `6`, and
`sed -n '637p' internal/web/assets/app.js` prints
`var EVENTS = ["spec", "session", "goal", "verify", "kanban", "config"];`. The unchanged-diff half
is the mechanical pin that the `/todo` route did not change the event vocabulary; it is asserted
rather than assumed, because the route decision was taken against a prediction that it would force
one (spec.md §A.4).

The worktree-served half of REQ-WTQ-007 carries no criterion of its own by design: it states a
**limitation** — no live event is guaranteed — and asserting the absence of an event would assert
a race rather than a behaviour. What is testable there is already covered: correctness on load
(AC-WTQ-005) and the unchanged watch set (this criterion).

## §E Traceability

| Requirement | Criteria |
|---|---|
| REQ-WTQ-001 | AC-WTQ-001, AC-WTQ-006 |
| REQ-WTQ-002 | AC-WTQ-002 |
| REQ-WTQ-003 | AC-WTQ-003 |
| REQ-WTQ-004 | AC-WTQ-004, AC-WTQ-005, AC-WTQ-006, AC-WTQ-008 |
| REQ-WTQ-005 | AC-WTQ-007 |
| REQ-WTQ-006 | AC-WTQ-009 |
| REQ-WTQ-007 | AC-WTQ-010 |
| REQ-WTQ-008 | AC-WTQ-002, AC-WTQ-011 |

Every criterion names at least one requirement; every requirement is covered by at least one
criterion.

## §F Definition of Done

- All eleven criteria pass.
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test ./internal/web/... ./internal/kanban/... ./internal/cli/...` passes; the full-suite
  verdict comes from CI, not from a local full run.
- `golangci-lint run` reports no new findings against the pre-change baseline.
- `templ generate` has been run and the regenerated `_templ.go` files are committed.
- No file under `internal/template/templates/` is touched (spec.md §C.4).
