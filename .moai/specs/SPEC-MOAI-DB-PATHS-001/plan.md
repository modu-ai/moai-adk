# SPEC-MOAI-DB-PATHS-001 — Plan (card t1827)

Sections are ordered by decision reversibility: the decisions most likely to change come first (the command contract, the instruction block text), the mechanical sweeps last. Milestones are priority-ordered; no time estimates.

## A. Decisions most likely to change

### A.1 The command contract (public interface; hardest to change after release)

`moai paths` (top-level, `GroupID: "tools"`, registered in `internal/cli/root.go` beside the other `rootCmd.AddCommand` calls).

| Flag | Meaning |
|---|---|
| (none) | human listing, one `name: value` line per field; the first five names are the `--field` names below, followed by `queue-state`, `legacy-dir`, `queue-layout`, and the broker lines |
| `--json` | one JSON object on stdout; notes and warnings go to stderr only |
| `--field <name>` | bare value of one of the five names in the table below, then newline; an unknown name, or a name whose value is unresolved, exits non-zero with empty stdout |
| `--run <run-id>` | the broker path for that run (id validated by the broker's own rule); without it, a path template and a count of run directories |

Valid `--field` names and the JSON key each one equals (this table is the AC-DBP-004 contract):

| `--field` name | JSON key | Value |
|---|---|---|
| `project-key` | `project_key` | the project key |
| `project-dir` | `project_dir` | the project state directory |
| `todo-dir` | `todo.dir` | the canonical queue directory; exactly the directory part of `todo.path` |
| `todo-db` | `todo.path` | the queue database file |
| `factory-db` | `factory.path` | the factory database file |

JSON shape (names are the acceptance contract, AC-DBP-001):

```text
project_key, project_dir, queue_layout ("home" | "project-local"),
todo:    { dir, path, exists, size_bytes, wal: {exists,size_bytes}, shm: {exists,size_bytes},
           queue_state ("canonical" | "pre-adoption"), legacy_dir (string | null) }
factory: { path, exists, size_bytes, wal: {...}, shm: {...} }
broker:  { path_template, existing_runs }          // or { path, exists, ... } when --run is given
```

An unresolved field is `null` in the JSON and makes the process exit non-zero. `queue_layout` describes the queue resolver only; the factory path is reported exactly as its own resolver returns it (spec § A.3, last paragraph). `legacy_dir` is non-null exactly when `queue_state` is `pre-adoption`.

Exit codes: 0 when every requested path resolved (absence of a database is data, not an error); non-zero when a path is unresolved, `--run` is invalid, or `--field` is unknown.

Why `moai paths` and not `moai todo where` plus `moai factory where` (SPEC § F.1): one resolver call chain for three stores, a spelling that survives the card t1641 rename, and a single pointer sentence for every instruction surface.

### A.2 The resolver seam (what the command may call) — one resolver per store

**The queue directory has exactly one resolver, and it is the canonical one:** `factory.StateDirForRoot(factory.ResolveTodoQueueRoot(launch))`, with `launch` from `resolveProjectDir()` (`internal/cli/session.go`), the call the todo verbs and the `factoryCardRoot()` path make (`factoryCardRoot`, `todo_disclosure.go`) — a rule every factory-opening surface shares (card t1862 aligned `factory runs` and `recover-resume` to it). `moai paths` does not call `BacklogPathForRoot`: that pure form returns a LEGACY directory (project-local `todo`, `kanban`, or legacy home) while the canonical directory holds no queue yet and a legacy one does (`resolveStateDir(root, false)`, `internal/factory/state_dir.go:115-155`), and the first adopting verb then relocates the queue (`BacklogPathForRootAdopting`, `state_dir.go:246-252`, reached by `todoBacklogPath` at `internal/cli/todo.go:62`). Reporting the pure form would hand an agent, and the foreman watch's `--field todo-dir`, a directory about to disappear. The decision and its cost are SPEC § F.6.

**One new pure helper, in a NEW file** `internal/factory/queue_location.go` (no existing `internal/factory` file is edited): `QueueLocationForRoot(root)` returns the canonical directory (`StateDirForRoot(root)`), the queue database path (the canonical directory joined with `backlog.json`, passed through the existing `backlogSQLitePath` rule so the `.json`-to-`.db` derivation is not re-implemented), the layout (`project-local` exactly when the canonical directory equals `projectStateDirForRoot(root)`, else `home`), and the legacy directory (the directory `resolveStateDir(root, false)` returns when its legacy flag is true, else empty). It performs only `stat` and directory reads. The legacy directory is disclosed, never opened.

Other allowed calls, because each is pure (no file creation): `homestate.FactoryDBPath`, `homestate.ProjectKey`, `homestate.CanonicalProjectRoot`, `factorymsg.BrokerPath`.

Forbidden, because each creates, opens, or relocates something: `factory.BacklogPathForRoot` for the reported path (can return a legacy directory), `factory.BacklogPathForRootAdopting` (relocates, and calls `EnsureProjectLayout`), `homestate.OpenFactory*`, `homestate.EnsureProjectLayout`, `factory.FactoryRegistryPath` (calls `EnsureProjectLayout`; also carries the ghost-#8 fallback), `factorymsg.Open*`, any `sql.Open`. AC-DBP-003 enforces both the behavior (before/after walk with a positive control) and the source (forbidden-identifier test); AC-DBP-002 enforces the queue seam (pre-adoption, post-adoption, and resolver-failure cases).

Root selection: `ResolveTodoQueueRoot` makes a linked worktree converge on the primary checkout (REQ-DBP-006).

### A.3 The instruction block (text most likely to be reworded)

Written once in full in `.moai/docs/todo-queue-storage.md` under a new section "Finding the database files", and carried in compact form in every S12 surface. The five fixed lead phrases are the test's contract (AC-DBP-009); everything after the lead phrase may be reworded without touching the test.

```text
Find the database files with `moai paths`; never guess them.
Forbidden guessed paths: `~/.moai/todo.db`, `~/.moai/backlog.db`, `~/.moai/db/todo.db`, `~/.moai/db/factory.db`, `~/.moai/db/<key>/backlog.db`, `<root>/.moai/state/backlog.db`, `<root>/.moai/state/factory.db`, `<root>/.moai/state/factory/factory.db`.
Never `sqlite3`-open a guessed path: opening a missing file creates an empty one.
Direct reads only on a snapshot copy of `.db`, `.db-wal`, and `.db-shm` together.
Prefer the CLI: `moai todo list --json`, `moai factory status`.
```

The five lead phrases, verbatim, each matched as a single-line substring and each appearing exactly once per S12 file (so a surface keeps every lead phrase on one physical line, no wrapping inside the phrase):

1. `Find the database files with `
2. `Forbidden guessed paths:`
3. `Never `sqlite3`-open a guessed path`
4. `Direct reads only on a snapshot copy`
5. `Prefer the CLI:`

The user-facing form (README in four locales, eight docs-site pages; REQ-DBP-013) carries the same five statements in locale prose, with five language-neutral anchors in code spans that the AC-DBP-012 test matches in every page: `moai paths`, `~/.moai/db/todo.db`, `sqlite3`, `.db-wal`, `moai todo list --json`. The anchors are inserted as sentences inside the existing section next to the original mention, never as a new heading (the `hns-oss-docs-verify` section-count parity must hold).

Constraints on this text: it is English in S12 surfaces; it uses the canonical `moai todo` spelling (card t1641 ordering); it never writes a bare `.moai/db` (the deprecated-path text guard in `internal/cli/deprecated_paths_text_reference_test.go` flags `.moai/db` unless preceded by `~/`, so the temp-layout spelling `<canonical>/.moai/db/<key>` must not appear in any template file); and the list is illustrative — the rule is that only `moai paths` output is valid.

### A.4 Doctor check scope (M5; droppable as a unit)

Detects zero-byte regular `*.db` files directly in: the moai home, `<home>/db`, the project state directory and its `todo/` and `factory/`, `<root>/.moai`, `<root>/.moai/state`, and its `todo/` and `factory/`; excludes the canonical database paths; warning severity only. All eight observed ghosts fall inside this set (SPEC § A.1). The run phase greps the tree for every legitimate `*.db` writer under those directories before finalizing the list and records the result in progress.md §E.2; a legitimate zero-byte writer found there becomes an exclusion with a reason, not a silent miss.

## B. Milestones

### M1 — `moai paths` and its pure derivation (Priority High)

Deliverables: `internal/cli/paths.go` and `internal/cli/paths_test.go`; the new `internal/factory/queue_location.go` with its test; registration in `root.go`. Criteria: AC-DBP-001 to AC-DBP-007.

Order inside the milestone: RED tests first (queue parity AC-DBP-002 including the pre-adoption subtest, and layout AC-DBP-006, before the happy path, because they are the ones a recomputing or pure-resolver implementation fails), then the command, then the positive-control walker of AC-DBP-003. Add `@MX:ANCHOR` or `@MX:NOTE` per the MX protocol where fan-in warrants.

### M2 — Help-text pointers (Priority High)

Deliverables: `internal/cli/todo.go` (`Long`, lines ~224-246 and the stderr message at ~103), `internal/cli/handoff.go` (~40), the `factory` command's help (`internal/cli/factory_handoff_recover.go:20` carries only `Short`; add a `Long`). Criteria: AC-DBP-008. Constraint: keep the string pinned by `todo_help_store_test.go`.

### M3 — Instruction surfaces and the drift guard (Priority High)

Deliverables: the S12 edits (SPEC § C), the foreman watch rewrite, the two existing foreman-script test files `internal/factory/foreman_queue_statement_test.go` and `internal/factory/foreman_queue_watch_test.go` (edited, step 7), `TestDBPathsBlock_Surfaces`, `TestTemplateDBPathSurfacesCarryPointer`, `TestForemanWatch_FailsVisibly`. Criteria: AC-DBP-009 to AC-DBP-011, AC-DBP-013.

Per-file procedure, which the run phase follows verbatim:

1. Re-read the file in this tree and re-locate the insertion point by content, not by the line numbers in SPEC § C (they are coordinates at `93f4fc75c`).
2. Insert the compact block as new lines near the existing database mention; never rewrite an existing line in the two divergent twin pairs (AC-DBP-010). Keep each lead phrase of § A.3 on one physical line.
3. Apply the same insertion to the twin by a second, separate edit.
4. For the foreman skill: replace the hand-computed key block with `d=$(moai paths --field todo-dir) || { echo "moai paths failed" >&2; exit 1; }` plus the surrounding explanation; delete the paragraph that describes recomputing the key (`SKILL.md` lines ~147-154) and keep the WAL-watching rationale. Behavior across adoption: the directory is canonical, so on a project whose queue is still legacy the first poll reads `missing` and the first adopting `moai todo` verb produces one `backlog changed` line — benign, and the reason the canonical contract was chosen (SPEC § F.6).
5. `manager-todo.toml` is a GENERATED file (its first line reads "Generated by the MoAI agent dual-publication emitter; regenerate, do not edit"; neutral source `.claude/agents/moai/manager-todo.md`). Edit the neutral `.md` in both trees (step 2), then run `make agents-emit` (`Makefile:38`; it runs the golden-emission test of `./internal/template/agentemit` with `AGENTEMIT_UPDATE=1`) to regenerate the TOML, and verify with `make agents-emit-check` (read-only; it never regenerates). Never hand-edit the TOML. AC-DBP-009 judges the TOML after regeneration. The local-only `.moai/workflows/gtd.md` is edited on the local side only.
6. `make build` runs the read-only `agents-emit-check`, `commands-emit-check`, `embed-manifest-check` and then `go run ./internal/template/scripts/gen-catalog-hashes.go --all`, which rewrites the entry hashes in `internal/template/catalog.yaml` for the changed foreman skill directory and `manager-todo.md`; commit that regenerated file with the change that caused it. `embed_manifest_gen.go` lists `go:embed` file names, not content hashes, and no S12 edit adds or removes a file, so it stays unchanged and `make embed-manifest-check` must stay green; if it goes red a file was added or removed by mistake. If `make commands-emit-check` goes red, return a blocker report; do not widen the AC-DBP-016 allow-list silently.

7. Existing foreman-script tests. `TestForemanQueueWatchResolvesCanonicalStateDir` (`foreman_queue_statement_test.go`, ~80-136) runs the extracted resolution prologue under `sh` and compares `$d` with `StateDirForRoot`; the `TestForemanQueueWatch_*` family (`foreman_queue_watch_test.go`: `watchFixture` ~134-154, `armWatch` ~160-191; four tests, only `TestForemanQueueWatch_FiresOnMutation` has `local` and `template` subtests) runs the extracted script under `sh` with only `MOAI_HOME` set. After step 4 the script calls `moai paths --field todo-dir`, so with no `moai` on the fixture PATH `$d` is never assigned and every one of them goes red. Both files are edited: the fixture builds a directory holding an executable `moai` shell STUB, prepends it to the subprocess `PATH` (the original `PATH` stays behind it so `sh`, `git`, `cksum` still resolve), and the stub prints the fixture's `StateDirForRoot` for exactly `paths --field todo-dir` and exits 2 for anything else. A STUB, not a tree-built binary, for two reasons: it adds one trivial `exec` before the loop, so the time-bound `TestForemanQueueWatch_*` family (a 2-second settle plus a `watchWindow` deadline in `armWatch`) stays deterministic, whereas building the CLI inside a package test would add build time and a cross-package dependency to a test that must not be slow; and the real command's agreement with the resolver is already proved by AC-DBP-002 and AC-DBP-004, so the stub only has to prove the script consumes the command output unchanged. The statement test keeps its anti-regression assertions (no `state/todo` literal, not the retired path) and changes its equality check to: the stub was invoked with exactly `paths --field todo-dir` and `$d` equals its output. A second stub variant that exits non-zero is the `TestForemanWatch_FailsVisibly` subject (AC-DBP-011).

The codemaps (`.moai/project/codemaps/*.md`) are generated and are not hand-edited here (SPEC § F.5); the sync phase regenerates them and records the pointer check in its verdict.

### M4 — Four-locale user-facing pages (Priority Medium)

Deliverables: the five statements in compact form in locale prose in each of `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` (line ~66 region) and in each locale of `docs-site/content/*/advanced/factory-mode.md` and `docs-site/content/*/utility-commands/moai-todo.md`, carrying the five anchors of § A.3; `TestDBPathsBlock_UserFacingPages` (12 files x 5 anchors). Criteria: AC-DBP-012. Canonical locale `ko`; load `hns-oss-docs-i18n-rules` first; run the `hns-oss-docs-verify` recipe. `moai-web-console.md` is inspected and left unchanged (SPEC § C).

### M5 — Doctor ghost check (Priority Low; droppable as a unit)

Deliverables: `internal/cli/doctor_db_ghosts.go` and its test, registration in `doctor.go` (the `moaiChecks` list, beside `todoGhostInventoryCheckName`), the name in the contract map of `internal/cli/binary_lag_test.go`, regenerated `internal/cli/testdata/doctor-{dark,light,nocolor}.golden` via `UPDATE_GOLDEN=1`. Criteria: AC-DBP-014, AC-DBP-015. Follow the structure of `doctor_todo_ghost.go` (read-only, three states, never deletes). Dropping M5 changes the AC-DBP-016 variant that applies (acceptance.md).

## C. Technical approach notes

- The command is a thin renderer over resolvers; its only logic is `stat` for existence and size, one `ReadDir` for the run count, and layout naming. Estimated 150-250 lines plus 300-450 lines of tests.
- Layout naming: `queue_layout` is `project-local` exactly when the canonical queue directory equals the project state directory (temp-origin without an absolute `MOAI_HOME`, or a home that cannot be resolved); the factory path uses `homestate.FactoryDBPath`, whose temp-rooted spelling can differ from the queue's (SPEC § A.3). The command reports each path as its own resolver returns it and does not assume the two share a root; the single `queue_layout` value is deliberately queue-only.
- Snapshot-copy instruction (element 4) is guidance for agents; the command does not copy anything.
- Run-phase prompts for `manager-develop` use the Tier M Section A-E template; the working tree for the run phase is the lane's card worktree, one writer per tree.

## D. Risks

| Risk | Mitigation |
|---|---|
| The temp-origin layouts of queue and factory differ (SPEC § A.3) and a recomputing command reports the wrong one | AC-DBP-002 judges parity against files the real verbs created, in both layouts |
| The command reports a legacy directory that the first adopting verb then relocates | The pure `BacklogPathForRoot` is forbidden for the reported path; AC-DBP-002's pre-adoption subtest fails a pure-resolver implementation |
| Adoption is refused (a filesystem that cannot rename), so the verbs keep serving the legacy directory while the command reports the canonical one | `queue_state: pre-adoption` and `legacy_dir` keep reporting where the data is until the relocation succeeds; recorded as an edge case in acceptance.md |
| Template text trips the deprecated-path guard (`.moai/db`) | Constraint in § A.3; AC-DBP-013 re-runs `TestDeprecatedPaths_NoTemplateTextReference` |
| Card t1641 edits the same files and starts after this card | Canonical `moai todo` spelling; no existing `gtd` word touched; guard keyed on a sentence containing neither word (SPEC § D) |
| An older installed binary lacks `moai paths`, so the foreman watch fails | The watch stops with a visible error (REQ-DBP-012); the binary-lag doctor pair already reports the mismatch; the card's landing sequence rebuilds and reinstalls |
| The twin divergence is silently "fixed" by copying one file over the other | AC-DBP-010 requires zero removed lines in the two divergent pairs (four files) |
| A hand edit to the generated Codex TOML drifts from the emitter | M3 step 5 routes through `make agents-emit`; `make agents-emit-check` is part of `make build` and part of AC-DBP-009 |
| The foreman watch rewrite turns the existing foreman-script tests red because no `moai` is on their fixture PATH | M3 step 7 edits both test files to provision a `moai` stub; both files are on the AC-DBP-016 allow-list; AC-DBP-011 runs them and counts them swept |
| A test enumerating the registered root commands pins the command set, so adding `moai paths` turns it red | The run phase searches the tests for such a pin before M1 closes; if one exists it returns a blocker report naming the file, and the allow-list is amended in a SPEC revision rather than widened silently |
| New tests touch the real home | Absolute `MOAI_HOME` under `t.TempDir()`; `internal/testguard` must stay green; no `os.UserHomeDir` in new tests; every `~/` string literal in a new test file (the planned tests need `~/.moai/db/todo.db`-style literals) carries a `// home-guard:allow <reason>` comment with a non-empty reason on its own line or the line directly above (`internal/testguard/scan.go`: the marker without a reason exempts nothing); `internal/testguard/home_access_seed.txt` is never edited, so the AC-DBP-016 allow-list holds |
| The doctor list misses or over-reports a legitimate zero-byte database | § A.4 pre-finalization grep, recorded in progress.md; exclusions carry reasons |
| `moai paths` becomes a second place that must change when a store moves | REQ-DBP-002 forces it through the same resolution chains once the project root is fixed; AC-DBP-002 fails if a resolver and the command disagree. The factory verbs share the same root rule (card t1862 aligned `factory runs` and `recover-resume` to `factoryCardRoot()`) |
| A user-facing page carries the anchors but not the statements | Residual risk accepted: the anchor test proves presence, review judges meaning; the compact wording is short enough for the section review in `hns-oss-docs-verify` |

## E. Open items (none block the run phase)

- Whether to later add the search-cache `sessions.db` and the contract receipt store to `moai paths` (SPEC § E, out of scope now; additive to the JSON shape).
- Whether an MCP tool for the same report is wanted (would need the `moai-mcp-tools.md` catalogue entry).
- The pre-existing twin divergences (`SKILL.md` lines 27 and 31, `todo-queue-storage.md` lines 25 and 38) and the template-only status of `manager-todo.toml` are observed and deliberately left alone.
- RESOLVED (card t1862): the audit-F1 divergence — `factory runs` and `recover-resume` rooting at the working directory while `moai paths` roots at CLAUDE_PROJECT_DIR first — is closed; both verbs route through `factoryCardRoot()` with the boundary test that line planned (real command invoked with CLAUDE_PROJECT_DIR and the working directory naming different repos).
- Codemaps: regenerated at sync, pointer checked there (SPEC § F.5); a codemap that still names a path without the pointer becomes follow-up card input.
