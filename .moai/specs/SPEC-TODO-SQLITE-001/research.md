# SPEC-TODO-SQLITE-001 — Research

All measurements taken 2026-08-27 against worktree `WT-todo-sqlite@d29b8942e`
(Go cross-platform, templates embedded via `//go:embed`). Items marked VERIFIED cite a
command actually run in this planning session; corrections to the dispatch's premises
are listed in §6.

## 1. Current state of the queue store (VERIFIED)

Queue file, primary checkout `<primary>/.moai/state/kanban/backlog.json`:

- Measured live twice: 81 items / `last_seq: 309` at dispatch time, 82 items /
  `last_seq: 310` minutes later — the queue is LIVE (new cards kept arriving while
  this SPEC was authored). By state: 56 queued, 15 picked, 11 dropped; 1 finding.
- Top-level keys `{version: 1, last_seq, items[], findings[]}`.
- Directory census: 133 files (~130 session-registry `<uuid>.json`, 198–214 B each),
  636 KB total.

Store implementation (`internal/kanban/backlog_store.go`, ~20 KB):

- `Load()` reads the WHOLE file lock-free (`atomicfile.ReadFile` + unmarshal); missing
  file = empty queue (fail-open contract); malformed file = error, no repair.
- `Mutate(fn)` acquires the sibling `backlog.lock` (`BoardLock`, 25 ms × 40 retries
  ≈ 1 s window) across load→mutate→write, landing via temp + atomic rename.
- Ids issued INSIDE the locked region from the persisted high-water mark `last_seq`
  (never max-present-id; done deletes rows).
- Exported surface callers see: `NewBacklogStore(path)`, `BacklogPathForRoot(root)`
  (THE canonical join), `QueuedCount()`, `QueuedBacklogCountForRoot(root)`, `Path()`,
  `LockPath()`, `Load()`, `Mutate()`, `Add()`; record types `BacklogRecord{Version,
  LastSeq, Items, Findings}`, `BacklogItem{ID,Text,AddedAt,SpecID *,State}`,
  `BacklogFinding{SubjectID,RelatedID,Relation,Source,Score,Note,At}`.

## 2. What shares the directory, and what does not (VERIFIED)

Two producers live in `.moai/state/kanban/`: the queue file (+ its locks) and the
session registry (`record.go`: `stateDirSegments = {".moai","state","kanban"}`,
`<uuid>.json` written best-effort by the launcher, read by orchestrator/dedup gate).

Sibling stores inventoried:

| Sibling | Touches the shared dir? | Notes |
|---|---|---|
| `board_store.go` / `board.go` | NO | board lives in `.moai/state/kanban-board/` (`boardDirSegments`), deliberately distinct (AP-24) |
| `revision.go` | NO | reads deep-scan results dir |
| `status_read.go` | NO | git blob reads of specs on branches |
| `prlink.go` | NO | pure analysis, "writes nothing, ever" |

Conclusion: the rename moves ONE directory containing queue + registry; the board is
unaffected. Registry files ride the rename as data, keeping their JSON-on-disk model
(their access pattern is filename-keyed random access; SQL adds nothing there).

## 3. Path-reference inventory (VERIFIED — supersedes the dispatch's list)

Production Go literals constructing/consuming the path. Methodology note (corrected at
plan-audit): a single-pattern grep for the joined literal `state/kanban` misses
segment-wise joins that spell the path as separate quoted arguments
(`".moai", "state", "kanban"`); the authoritative inventory below comes from BOTH
patterns:

| File | Shape | Action required |
|---|---|---|
| `internal/kanban/backlog_store.go:250` | `filepath.Join(root,".moai","state","kanban","backlog.json")` (`BacklogPathForRoot`) | THE central constant; becomes todo-dir based |
| `internal/kanban/record.go:43` | `stateDirSegments` | rename segment list |
| `internal/kanban/todo_root.go:119` | inline join (fallback branch) | same rename |
| `internal/web/events.go:30` | `watchMap["kanban"] = ".moai/state/kanban"` | swap watched path only; SSE key stays |
| `internal/web/viewmodel_ops.go:509` | hand-join for `*.json` registry glob | swap |
| `internal/statusline/backlog.go:50` | hand-join (`resolveBoardRoot` context) | swap + engine read via NEW cost-budgeted path |
| `internal/cli/kanban.go:330` | `companionRegistryPath` → join … `"companions.json"` | swap (companion name registry lives in the renamed directory) |
| `internal/cli/kanban.go:368` | `leadRegistryPath` → join … `"leads.json"` | swap (lead name registry, same dir) |
| `internal/cli/todo.go:48` | comment literal naming the path | update with code sites |
| `internal/cli/todo.go:80` | user-visible Long help text naming the path verbatim | update (user-visible string, not just code) |
| `internal/cli/graph.go:~57-66` | consumes via `NewBacklogStore(todoBacklogPath(...))` | no direct literal (see correction §6.3) |
| `internal/hook/session_start_kanban.go` / `_factory.go` / `session_start_record.go` | go through `kanban.*` helpers incl. `RecordPath` | no direct literal |
| `internal/kanban/board_test.go:123-124`, `factory_slots_test.go:128`, `backlog_store_test.go` et al. | test assertions/comments | update expectations deliberately |

Site count summary (consistent everywhere it is cited): **8 structural construction
or value sites** — `backlog_store.go:250`, `record.go:43`, `todo_root.go:119`,
`events.go:30`, `viewmodel_ops.go:509`, `statusline/backlog.go:50`,
`cli/kanban.go:330`, `cli/kanban.go:368` — plus individually enumerated prose,
help-text, and comment literals (`todo.go:48` comment, `todo.go:80` user-visible Long
help, `screens_templ.go:1242` generated-file comment, `statusline/backlog.go:41`,
`record.go:49`, `todo_root.go:110`, `session_start_kanban.go:194`).

Mechanical guard that survives: `TestNoBareJoinBacklogPathSurvives`
(`todo_root_convention_test.go`) walks cli/web/statusline packages for hand-built joins
of `backlog.json` — any new construction MUST go through `BacklogPathForRoot`, which
keeps this SPEC's sweep honest.

Template references (Template-First targets, VERIFIED at exactly three):
`internal/template/templates/.claude/skills/moai/SKILL.md:170`,
`.../skills/moai/workflows/todo.md:17`,
`.../skills/moai-kanban-foreman/SKILL.md:95` (inside a shell snippet — functional, not
prose). Local mirror copies of the same three exist under `.claude/skills/`; per
CLAUDE.local.md §2 the managed roots are redeployed wholesale by `moai update`, so the
template edit + `make build` is the durable fix and dogfood machines sync from it.
No references in `.moai/docs/`, CLAUDE.md, AGENTS.md, or hook scripts.

Gitignore (VERIFIED): `**/.moai/state/` and `.moai/state/` both ignored — the database
and its `-wal`/`-shm` siblings need NO gitignore changes.

## 4. Consumers (VERIFIED — reconciliation of lead's list)

Lead-reported consumers checked one by one: `internal/cli/todo*.go` (thin wiring; all
nine verbs funnel through `BacklogStore.Load/Mutate/Add` — the store is the choke point
the swap hits); `internal/cli/graph.go` (real consumer via the store, injection-seamed
resolver — see correction §6.3); `internal/web/todo_queue_read.go` (the console's
SINGLE declared read seam; its header doc explicitly anticipates this swap: "a JSON
file today, under review ... swapped by changing this function and nothing else");
`internal/statusline/backlog.go` (hand-rolled join + snippet-parse of the whole 118 KB
file EVERY render — the one genuinely hot reader).

Additional consumers the lead did not list: `internal/web/events.go` (fsnotify watch
map) and `internal/web/viewmodel_ops.go` (registry glob for the ops view); both caught
by the §3 sweep, both trivial path-value swaps.

## 5. The global side (dispatch premise E — answered with measurements)

`~/.moai/todo/` holds 200 key directories of shape `<basename>-<sha256[:4]>`
(`TodoQueueProjectKey`), 1.5 MB total, containing 197 legacy `backlog.json` files at
the NESTED path `<key>/.moai/state/kanban/backlog.json`. There is NO aggregate index
anywhere — the keys are independent per-project fallback queues, most dormant.

Scope answer: everything is root-parametric, so the SAME lazy rules apply per root on
first open; stale roots keep working via the documented fallback READ indefinitely; no
global index, no bulk migrator, no schema accommodation. Explicitly designed-away.

## 6. Dispatch-premise corrections (say-so-per-dispatch section)

1. `screens_templ.go` does NOT construct the path — its only hit (line 1242) is a
   Korean COMMENT describing `watchMap["kanban"]`. The functional literal lives in
   `events.go`. (Whether the comment's `.templ` source carries the same text is a
   run-phase grep-away.)
2. "Any hook scripts/templates referencing the path in prose/docs" — NONE exist beyond
   the three skill templates above; hooks, `.moai/docs/`, CLAUDE.md, AGENTS.md: zero
   hits.
3. `graph.go:64` IS a genuine consumer, but exclusively through
   `kanban.NewBacklogStore(todoBacklogPath(...))` with an injectable resolver — no
   string literal, no bespoke read logic. Listed here because the dispatch phrased it
   as a string-reference site ("String references … 15 files"). CORRECTED AT PLAN-AUDIT:
   this document's OWN first inventory also undercounted — its sweep matched only the
   joined literal `state/kanban`, missing segment-wise joins spelled as separate quoted
   arguments. The plan-auditor surfaced `internal/cli/kanban.go:330/:368`
   (companions.json / leads.json hand-joins into the renamed directory); a dual-pattern
   re-sweep (joined literal + `".moai", "state", "kanban"` segments) is now the §3
   methodology, and the authoritative totals are the 8 structural sites + enumerated
   prose/help/comment literals listed there. The dispatch's "15 files" figure mixed
   production code with test files and comments; it was neither right nor wrong so much
   as measured by a single pattern.
4. Premise E's "~200 accumulated project keys … design query patterns accordingly":
   verified AND resolved as non-actionable (§5). Nothing accumulates on the global
   side that a query pattern would serve.
5. Live-queue premise CONFIRMED by direct observation (item count moved 81→82 during
   authoring) — migration machinery must tolerate arbitrary shape evolution
   (`last_seq`, findings tuples, dropped-heavy mixes), not today's exact census.

## 7. Driver decision inputs (VERIFIED facts feeding design.md)

- `.github/workflows/ci.yml` lines 415–424: cross-build jobs run `CGO_ENABLED: "0"`
  with a `goos` matrix — a CGO driver cannot compile there; this alone excludes
  `mattn/go-sqlite3` in this repository.
- `go.mod`: no existing cgo-bearing dependency on the direct-require list; repo makes
  no use of `import "C"` anywhere (grep). Adding modernc.org/sqlite introduces this
  repo's first large pure-Go translation dependency (~+MB-order binary weight on a
  measured 62 MB baseline — precise delta is an M1 deliverable, not assumed here).
- Locking substrate precedent: dedicated windows implementations already exist for the
  file-lock family (`board_lock_windows.go`, `lock_alive_windows.go`,
  `factory_alive_windows.go`) and stay in service unchanged; the engine adds its own
  portable locking beneath the kept outer file lock.

## 8. Open questions

None — all dispatch-flagged axes resolved above and in design.md. The one judgement
call worth restating for reviewers: keeping the legacy file lock as outer serializer
during and after cutover (REQ-TOSQ-008) trades a little purity for mechanical identity
with the proven concurrency protocol; justification in design.md §4.
