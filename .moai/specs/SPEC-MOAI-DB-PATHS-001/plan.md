# SPEC-MOAI-DB-PATHS-001 — Plan (card t1827)

Sections are ordered by decision reversibility: the decisions most likely to change come first (the command contract, the instruction block text), the mechanical sweeps last. Milestones are priority-ordered; no time estimates.

## A. Decisions most likely to change

### A.1 The command contract (public interface; hardest to change after release)

`moai paths` (top-level, `GroupID: "tools"`, registered in `internal/cli/root.go` beside the other `rootCmd.AddCommand` calls).

| Flag | Meaning |
|---|---|
| (none) | human listing, one `name: value` line per field |
| `--json` | one JSON object on stdout; notes and warnings go to stderr only |
| `--field <name>` | bare value of `project-key`, `project-dir`, `todo-dir`, `todo-db`, `factory-db`, then newline; unknown name exits non-zero with empty stdout |
| `--run <run-id>` | the broker path for that run (id validated by the broker's own rule); without it, a path template and a count of run directories |

JSON shape (names are the acceptance contract, AC-DBP-001):

```text
project_key, layout ("home" | "project-local"), project_dir,
todo:    { path, exists, size_bytes, wal: {exists,size_bytes}, shm: {exists,size_bytes} }
factory: { path, exists, size_bytes, wal: {...}, shm: {...} }
broker:  { path_template, existing_runs }          // or { path, exists, ... } when --run is given
```

Exit codes: 0 when every requested path resolved (absence of a database is data, not an error); non-zero when a path is unresolved, `--run` is invalid, or `--field` is unknown.

Why `moai paths` and not `moai todo where` plus `moai factory where` (SPEC § F.1): one resolver call chain for three stores, a spelling that survives the card t1641 rename, and a single pointer sentence for every instruction surface.

### A.2 The resolver seam (what the command may call)

Allowed, because each is pure (no file creation): `factory.ResolveTodoQueueRoot`, `factory.BacklogPathForRoot` (pure; the adopting variant is forbidden), the sibling `backlog.db` derivation (`backlogSQLitePath`, `internal/factory/backlog_sqlite.go`, currently unexported — M1 adds one exported pure wrapper next to it rather than re-deriving the suffix rule), `homestate.FactoryDBPath`, `homestate.ProjectKey`, `homestate.CanonicalProjectRoot`, `factorymsg.BrokerPath`.

Forbidden, because each creates or opens something: `homestate.OpenFactory*`, `homestate.EnsureProjectLayout`, `factory.FactoryRegistryPath` (calls `EnsureProjectLayout`; also carries the ghost-#8 fallback), `factory.BacklogPathForRootAdopting`, `factorymsg.Open*`, any `sql.Open`. AC-DBP-003 enforces both the behavior (before/after walk with a positive control) and the source (forbidden-identifier test).

Root selection: the launch directory is `resolveProjectDir()` (`internal/cli/session.go`), the same call the todo and factory verbs make (`factoryCardRoot`, `todo_disclosure.go`), then `ResolveTodoQueueRoot` so a linked worktree converges on the primary checkout (REQ-DBP-006).

### A.3 The instruction block (text most likely to be reworded)

Written once in full in `.moai/docs/todo-queue-storage.md` under a new section "Finding the database files", and carried in compact form in every S12 surface. The five fixed lead phrases are the test's contract (AC-DBP-009); everything after the lead phrase may be reworded without touching the test.

```text
Find the database files with `moai paths`; never guess them.
Forbidden guessed paths: `~/.moai/todo.db`, `~/.moai/backlog.db`, `~/.moai/db/todo.db`, `~/.moai/db/factory.db`, `~/.moai/db/<key>/backlog.db`, `<root>/.moai/state/backlog.db`, `<root>/.moai/state/factory.db`, `<root>/.moai/state/factory/factory.db`.
Never `sqlite3`-open a guessed path: opening a missing file creates an empty one.
Direct reads only on a snapshot copy of `.db`, `.db-wal`, and `.db-shm` together.
Prefer the CLI: `moai todo list --json`, `moai factory status`.
```

Constraints on this text: it is English; it uses the canonical `moai todo` spelling (card t1641 ordering); it never writes a bare `.moai/db` (the deprecated-path text guard in `internal/cli/deprecated_paths_text_reference_test.go` flags `.moai/db` unless preceded by `~/`, so the temp-layout spelling `<canonical>/.moai/db/<key>` must not appear in any template file); and the list is illustrative — the rule is that only `moai paths` output is valid.

### A.4 Doctor check scope (M5; droppable as a unit)

Detects zero-byte regular `*.db` files directly in: the moai home, `<home>/db`, the project state directory and its `todo/` and `factory/`, `<root>/.moai`, `<root>/.moai/state`, and its `todo/` and `factory/`; excludes the canonical database paths; warning severity only. All eight observed ghosts fall inside this set (SPEC § A.1). The run phase greps the tree for every legitimate `*.db` writer under those directories before finalizing the list and records the result in progress.md §E.2; a legitimate zero-byte writer found there becomes an exclusion with a reason, not a silent miss.

## B. Milestones

### M1 — `moai paths` and its pure derivation (Priority High)

Deliverables: `internal/cli/paths.go` and `internal/cli/paths_test.go`; one exported pure helper in `internal/factory` for the queue database path; registration in `root.go`. Criteria: AC-DBP-001 to AC-DBP-007.

Order inside the milestone: RED tests first (parity AC-DBP-002 and layout AC-DBP-006 before the happy path, because they are the ones a recomputing implementation fails), then the command, then the positive-control walker of AC-DBP-003. Add `@MX:ANCHOR` or `@MX:NOTE` per the MX protocol where fan-in warrants.

### M2 — Help-text pointers (Priority High)

Deliverables: `internal/cli/todo.go` (`Long`, lines ~224-246 and the stderr message at ~103), `internal/cli/handoff.go` (~40), the `factory` command's help (`internal/cli/factory_handoff_recover.go:20` carries only `Short`; add a `Long`). Criteria: AC-DBP-008. Constraint: keep the string pinned by `todo_help_store_test.go`.

### M3 — Instruction surfaces and the drift guard (Priority High)

Deliverables: the S12 edits (SPEC § C), the foreman watch rewrite, `TestDBPathsBlock_Surfaces`, `TestTemplateDBPathSurfacesCarryPointer`, `TestForemanWatch*`. Criteria: AC-DBP-009 to AC-DBP-011, AC-DBP-013.

Per-file procedure, which the run phase follows verbatim:

1. Re-read the file in this tree and re-locate the insertion point by content, not by the line numbers in SPEC § C (they are coordinates at `93f4fc75c`).
2. Insert the compact block as new lines near the existing database mention; never rewrite an existing line in the two divergent twin pairs (AC-DBP-010).
3. Apply the same insertion to the twin by a second, separate edit.
4. For the foreman skill: replace the hand-computed key block with `d=$(moai paths --field todo-dir) || { echo "moai paths failed" >&2; exit 1; }` plus the surrounding explanation; delete the paragraph that describes recomputing the key (`SKILL.md` lines ~148-154) and keep the WAL-watching rationale.
5. The template-only `manager-todo.toml` is edited on the template side only; the local-only `.moai/workflows/gtd.md` on the local side only.
6. `make build` regenerates the embedded manifest; commit the regenerated files with the change that caused them.

The codemaps (`.moai/project/codemaps/*.md`) are generated; the sync phase regenerates them. They are not hand-edited here.

### M4 — Four-locale user-facing pages (Priority Medium)

Deliverables: one added sentence in each of `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` (line ~66 region) and in each locale of `docs-site/content/*/advanced/factory-mode.md` and `docs-site/content/*/utility-commands/moai-todo.md`. Criteria: AC-DBP-012. Canonical locale `ko`; load `hns-oss-docs-i18n-rules` first; run the `hns-oss-docs-verify` recipe. `moai-web-console.md` is inspected and left unchanged (SPEC § C).

### M5 — Doctor ghost check (Priority Low; droppable as a unit)

Deliverables: `internal/cli/doctor_db_ghosts.go` and its test, registration in `doctor.go` (the `moaiChecks` list, beside `todoGhostInventoryCheckName`), the name in the contract map of `internal/cli/binary_lag_test.go`, regenerated `internal/cli/testdata/doctor-{dark,light,nocolor}.golden` via `UPDATE_GOLDEN=1`. Criteria: AC-DBP-014, AC-DBP-015. Follow the structure of `doctor_todo_ghost.go` (read-only, three states, never deletes).

## C. Technical approach notes

- The command is a thin renderer over resolvers; its only logic is `stat` for existence and size, one `ReadDir` for the run count, and layout naming. Estimated 150-250 lines plus 300-450 lines of tests.
- Layout naming: `layout` is `project-local` exactly when the queue resolver returns the project-local state directory (temp-origin without an absolute `MOAI_HOME`); the factory path uses `homestate.FactoryDBPath`, whose temp-rooted spelling differs from the queue's (SPEC § A.3). The command reports each path as its own resolver returns it and does not assume the two share a root.
- Snapshot-copy instruction (element 4) is guidance for agents; the command does not copy anything.
- Run-phase prompts for `manager-develop` use the Tier M Section A-E template; the working tree for the run phase is the lane's card worktree, one writer per tree.

## D. Risks

| Risk | Mitigation |
|---|---|
| The temp-origin layouts of queue and factory differ (SPEC § A.3) and a recomputing command reports the wrong one | AC-DBP-002 judges parity against files the real verbs created, in both layouts |
| Template text trips the deprecated-path guard (`.moai/db`) | Constraint in § A.3; AC-DBP-013 re-runs `TestDeprecatedPaths_NoTemplateTextReference` |
| Card t1641 edits the same files and starts after this card | Canonical `moai todo` spelling; no existing `gtd` word touched; guard keyed on a sentence containing neither word (SPEC § D) |
| An older installed binary lacks `moai paths`, so the foreman watch fails | The watch stops with a visible error (REQ-DBP-012); the binary-lag doctor pair already reports the mismatch; the card's landing sequence rebuilds and reinstalls |
| The twin divergence is silently "fixed" by copying one file over the other | AC-DBP-010 requires zero removed lines in the two divergent pairs |
| New tests touch the real home | Absolute `MOAI_HOME` under `t.TempDir()`; `internal/testguard` must stay green; no `os.UserHomeDir` or `~/` literals in new tests without a seed entry |
| The doctor list misses or over-reports a legitimate zero-byte database | § A.4 pre-finalization grep, recorded in progress.md; exclusions carry reasons |
| `moai paths` becomes a second place that must change when a store moves | REQ-DBP-002 forces it through the same resolvers; AC-DBP-002 fails if a resolver and the command disagree |

## E. Open items (none block the run phase)

- Whether to later add the search-cache `sessions.db` and the contract receipt store to `moai paths` (SPEC § E, out of scope now; additive to the JSON shape).
- Whether an MCP tool for the same report is wanted (would need the `moai-mcp-tools.md` catalogue entry).
- The pre-existing twin divergences (`SKILL.md` lines 27 and 31, `todo-queue-storage.md` lines 25 and 38) and the template-only status of `manager-todo.toml` are observed and deliberately left alone.
