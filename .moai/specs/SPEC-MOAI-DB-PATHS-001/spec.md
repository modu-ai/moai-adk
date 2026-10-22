---
id: SPEC-MOAI-DB-PATHS-001
title: "Confirmed database paths — a read-only path-reporting command, guess-proof instruction surfaces, and a ghost-file doctor check"
version: "0.1.5"
status: completed
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli + internal/homestate + internal/template/templates"
lifecycle: spec-anchored
tags: "db-paths, project-key, ghost-files, sqlite, instruction-surfaces, doctor, t1827"
tier: M
related_specs: [SPEC-HOME-STATE-ROLLOUT-001, SPEC-TODO-QUEUE-HOME-CANON-001, SPEC-TODO-STALE-STORE-001, SPEC-TODO-SURFACE-POLISH-001, SPEC-TODO-HOME-TEMP-GUARD-001, SPEC-FACTORY-LEASE-RENEW-001]
---

# SPEC-MOAI-DB-PATHS-001 — Confirmed database paths (card t1827)

## HISTORY

- 2026-10-17 v0.1.0 — manager-spec, plan-phase authoring (card t1827, Class C). Measurement tree: `93f4fc75c8dafc1c436458940efa8957fd25ae12` (branch `WT-db-paths-where`). Every file and line reference below was read in that tree; each is a moving coordinate and the run phase re-reads it before editing.
- 2026-10-17 v0.1.1 — manager-spec, plan-audit iteration 1 revision (verdict FAIL 0.69, defects D1-D13). Queue-path contract unified on the canonical directory with a disclosed pre-adoption state (D1, D8, D12); closure gate rewritten as an explicit allow-list (D2, D11); generated Codex TOML moved to the `make agents-emit` path and catalog-hash regeneration recorded (D3); five statements required on README and docs-site, codemaps deferral recorded (D4); `--field` set and JSON directory key defined (D5); the 36-file grep command pinned in acceptance.md with the root `reports/` exclusion classified (D6); wording, count, and coverage repairs (D7, D9, D10, D13). Requirement count stays 16, acceptance-criterion count stays 16.
- 2026-10-17 v0.1.2 — manager-spec, plan-audit iteration 2 revision (verdict FAIL 0.75; D2 unresolved as a criterion). N1: the foreman watch rewrite breaks the existing foreman script tests in `internal/factory`; both test files are now deliverables, edited to provision a `moai` stub on the fixture PATH. N2: the SPEC artifacts are `A` (added) entries on the change-set diff, not `M`; AC-DBP-016 re-checked against the real diff letters. O1: REQ-DBP-007 first clause reworded. O2: the swept count of AC-DBP-011 states the determinism condition. Counts unchanged (16/16).
- 2026-10-18 v0.1.4 — manager-spec, sync 중 문구 정정, 근거 F1 재현 (card t1827). The sync audit (`.moai/reports/t1827/sync-audit.md`, finding F1, reproduced) showed `moai paths` roots at CLAUDE_PROJECT_DIR first (`internal/cli/paths.go:114`, `resolveProjectDir()`) while `moai factory runs` and `recover-resume` open the factory at `os.Getwd()` (`internal/cli/factory_handoff_recover.go:80-84`). The resolver-parity wording of REQ-DBP-002 is narrowed to those facts; no requirement is added or removed and no behavior changes (leader ruling: the factory verbs are outside this card's allow-list). Counts unchanged (16/16).
- 2026-10-18 v0.1.5 — manager-spec, command-path correction, factory recover-resume → factory handoff recover-resume (audit F7, card t1827). REQ-DBP-002 and § H named a command that does not exist; the verb is registered under `factory handoff` (`internal/cli/factory_handoff_recover.go:24`). No requirement added or removed; counts unchanged (16/16).

## A. Background and problem

### A.1 The failure

On 2026-10-17 eight zero-byte `*.db` files were found at paths no code path writes (the operator deleted them):

| # | Guessed path | Why it is a guess |
|---|---|---|
| 1 | `~/.moai/todo.db` | no resolver names it |
| 2 | `~/.moai/backlog.db` | no resolver names it |
| 3 | `~/.moai/db/todo.db` | no `<project-key>` segment |
| 4 | `~/.moai/db/factory.db` | no `<project-key>` segment |
| 5 | `~/.moai/db/<key>/backlog.db` | the `todo/` segment is missing |
| 6 | `<root>/.moai/state/backlog.db` | not a queue location in either layout |
| 7 | `<root>/.moai/state/factory.db` | not a factory location in either layout |
| 8 | `<root>/.moai/state/factory/factory.db` | the error fallback of `FactoryRegistryPath` only (A.3); never the live path |

Mechanism, observed in this run: `sqlite3 <missing>.db '.tables'` and `sqlite3 <missing>.db 'select 1;'` both exit 0 and leave a zero-byte `<missing>.db` behind; `sqlite3 -readonly <missing>.db '.tables'` exits 1 and creates nothing. An agent that guesses a path and "just looks" therefore manufactures a file that later tooling must classify.

### A.2 Why agents guess

Instruction surfaces name `~/.moai/db/<project-key>/todo/backlog.db` but never say how `<project-key>` is computed, and no CLI prints a confirmed path. Observed in the tree at `93f4fc75c`:

- `internal/homestate/paths.go:30` `ProjectKey` = `<basename of the canonical root, non-[A-Za-z0-9._-] mapped to '-'>-<first 4 bytes of sha256(canonical root), hex>` (lines 43-58). A linked worktree shares the key of the primary checkout via `CanonicalProjectRoot` (lines 72-77, 116-161). For this repository: `moai-adk-go-1bd3d038`.
- Real paths here: `~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` and `~/.moai/db/moai-adk-go-1bd3d038/factory/factory.db`.
- No CLI verb prints a database path. Observed with a build of the pinned tree: `moai todo list --json` and `moai factory status --json` emit no path-valued key (the one `.db` string in the todo output is card t1827's own text), and `moai factory --help` lists no path verb.
- `.claude/skills/moai-factory-foreman/SKILL.md` lines 108-124 recompute the key by hand in shell (`sha256sum | cut -c1-8`). That is a second copy of `ProjectKey` that cannot follow the temp layout or the separate-git-dir rules and will drift from the Go function.

### A.3 The layouts a command must not re-derive

Three resolvers produce the stores, and they do not share one rule for the temporary-origin layout (read at `93f4fc75c`; the run phase proves the claims by test):

- Queue: the directory is `factory.StateDirForRoot(factory.ResolveTodoQueueRoot(launch))` (`internal/factory/todo_root.go:81`, `internal/factory/state_dir.go:26-41`) — home layout `<moai-home>/db/<key>/todo`; a temp-origin launch without an absolute `MOAI_HOME`, or a home that cannot be resolved, gives `<root>/.moai/state/todo`. The database is the sibling `backlog.db` of the queue file (`backlogSQLitePath`, `internal/factory/backlog_sqlite.go:~255`). This canonical directory is where the adopting `moai todo` verbs leave the queue (`todoBacklogPath` calls `BacklogPathForRootAdopting`, `internal/cli/todo.go:62`, `internal/factory/state_dir.go:246-252`).
- The pure `BacklogPathForRoot` (`internal/factory/state_dir.go:238`, used by the read-only verbs at `internal/cli/todo.go:117,130`) is a different value before adoption: while the canonical directory holds no queue and a legacy directory does (project-local `todo`, `kanban`, or a legacy home directory), `resolveStateDir(root, false)` returns that legacy directory (`state_dir.go:115-155`), and the first adopting verb then relocates the queue. The command therefore reports the canonical directory and discloses the legacy one separately (REQ-DBP-007); it does not use `BacklogPathForRoot` for the reported path.
- Factory: `homestate.FactoryDBPath` (`internal/homestate/paths.go:345`) via `ProjectDir` (line 225) — home layout `<moai-home>/db/<key>/factory/factory.db`; temp-rooted layout `<canonical>/.moai/db/<key>/factory/factory.db`.
- Broker: `factorymsg.BrokerPath` (`internal/factorymsg/store.go:~148-155`) = `<factory-dir>/messages/<run-id>/broker.db`, run id validated by `safeID`.

The queue and factory resolvers apply different temporary-origin predicates: the queue uses `TempOriginReason` (`internal/factory/temp_origin.go:75`), the factory uses `rootLayout` (`internal/homestate/paths.go:276-278`), which adds a git-directory exclusion. For a git-directory root inside a temp root the two can disagree, so the command reports the queue layout under a queue-specific name (§ B.1, plan § A.1) and reports the factory path exactly as its resolver returns it.

Two traps sit beside them: `homestate.OpenFactory` and `factory.FactoryRegistryPath` (`internal/factory/factory_slots.go:55`) call `EnsureProjectLayout`, which creates directories and writes `project.json`; and `FactoryRegistryPath` falls back to `<root>/.moai/state/factory/factory.db` when the home path cannot be resolved (ghost #8 above). A path-reporting command must use neither.

### A.4 Instruction surfaces that name a database path

Measured with a repository grep (the exact command, the stdout, and the exit code are ledger entries L-10 and L-11 in acceptance.md): at the pinned tree the grep excluding `.moai/specs`, `.moai/reports`, root `reports/`, and `*.go` returns 36 files that mention `backlog.db`, `factory.db`, `broker.db`, `todo.db`, or `project-key`. The same grep without the root `reports/` exclusion returns 40: the four extra files are historical report artifacts under root `reports/`, classified in §C and never edited. §C classifies all 36; Go sources are outside the grep (`*.go` is excluded) and are listed in §C separately.

## B. Requirements (GEARS)

### B.1 The path-reporting command

- **REQ-DBP-001**: The CLI shall provide a top-level command `moai paths` that reports, for the project resolved from the current launch directory, the project key, the project state directory, the queue database, the factory database, and the factory broker database.
- **REQ-DBP-002**: The command shall obtain every reported path from the resolution chain the owning verb itself uses once the project root is fixed, shall resolve that project root CLAUDE_PROJECT_DIR first (then the launch directory), shall report the todo queue at the canonical location the adopting todo verbs leave it in, and shall not recompute the project key or join path segments of its own.
- **REQ-DBP-003**: The command shall never create a file or directory, open a database through any SQLite driver, take a lock, or trigger a migration, adoption, or layout-ensuring call; it shall learn existence and size by `stat` and directory read only.
- **REQ-DBP-004**: For each reported database the command shall state whether it exists, its size in bytes, and whether its `-wal` and `-shm` companions exist, with their sizes.
- **REQ-DBP-005**: Where `--json` is given, the command shall write exactly one JSON object to stdout and nothing else, carrying a key for every `--field` name below, the `todo-dir` value being the directory part of the queue database path; where `--field <name>` is given with a name among `project-key`, `project-dir`, `todo-dir`, `todo-db`, and `factory-db`, it shall write only that field's bare value followed by a newline, and for any other or unresolved name it shall exit non-zero with empty stdout; otherwise it shall write a human-readable listing.
- **REQ-DBP-006**: While launched inside a linked worktree, the command shall report the same project key and the same paths it reports from the primary checkout.
- **REQ-DBP-007**: When the queue resolves to the project-local layout, the command shall report the project-local paths and name the queue layout; when a todo queue exists only in a legacy directory, it shall report the canonical queue location with the state `pre-adoption` and name the legacy directory without reading or moving it; when a path cannot be resolved, it shall report that field as unresolved, exit non-zero, and not substitute a guessed or fallback path.
- **REQ-DBP-008**: When `--run <run-id>` is given, the command shall report that run's broker path, and when the run id is not valid it shall exit non-zero without reporting a path; without `--run` it shall report the broker path template and the count of run directories found under the factory `messages/` directory.
- **REQ-DBP-009**: The `moai todo` and `moai factory` help texts and the `moai handoff` help text that names a database path shall point to `moai paths` as the source of the confirmed path, and the `moai todo` long help shall keep the canonical-store string that `internal/cli/todo_help_store_test.go` pins.

### B.2 The instruction surfaces

- **REQ-DBP-010**: Each surface of the twelve-file set S12 defined in §C that names a queue, factory, or broker database path shall carry five statements: (1) the path is obtained from `moai paths`; (2) the forbidden guessed-path list of §A.1; (3) a guessed path is never opened with `sqlite3` because opening creates an empty file; (4) a direct read is made only on a snapshot copy of the database together with its `-wal` and `-shm` files; (5) the CLI is preferred (`moai todo list --json`, `moai factory status`).
- **REQ-DBP-011**: Where an S12 surface has a template twin, the surface set shall receive the insertion in the template file and the local file by separate edits, and each divergence that already exists between twins shall remain exactly as it is.
- **REQ-DBP-012**: The factory foreman skill's queue-watch shell shall obtain the queue directory from `moai paths` instead of recomputing the project key, and shall stop with a visible error rather than fall back to a computed path when the command fails.
- **REQ-DBP-013**: Each of the four README files and the eight docs-site pages listed in §C shall carry the five statements of REQ-DBP-010 in compact form in its locale prose, with the command spelling, the forbidden path literals, and the quoted commands identical in all four locales.
- **REQ-DBP-014**: A repository test shall fail when an embedded template text file that names a `<project-key>` todo, factory, or broker database path lacks the §B.2 pointer sentence and is not on a reasoned allowlist.

### B.3 The doctor check

- **REQ-DBP-015**: Where `moai doctor` runs, a check named `DB Path Ghosts` shall report at warning severity every zero-byte regular `*.db` file found directly in the moai home, its `db/` directory, the project state directory and its `todo/` and `factory/` subdirectories, `<root>/.moai`, `<root>/.moai/state`, and its `todo/` and `factory/` subdirectories, excluding the canonical database paths reported by `moai paths`; the check shall be read-only, shall never delete or move a file, and shall never fail the doctor run.
- **REQ-DBP-016**: When the `DB Path Ghosts` check is registered, the doctor registry shall list its name in the binary-lag doctor-name contract and the doctor golden outputs shall contain its row.

## C. Surface inventory (observed)

Classification of the 36 non-Go grep hits (command in acceptance.md ledger L-10) plus the four root `reports/` files (L-11). "Twin" = a template copy under `internal/template/templates/`.

**S12 — the twelve-file set REQ-DBP-010 binds:** the five twin pairs of the "Agent and skill instructions" and "Canonical storage doc" rows below (ten files: `manager-todo.md`, the foreman `SKILL.md`, `.claude/skills/moai/SKILL.md`, `workflows/gtd.md`, `todo-queue-storage.md`, each local and template), the template-only `manager-todo.toml`, and the local-only `.moai/workflows/gtd.md`.

| Group | Files | Twin | Disposition |
|---|---|---|---|
| Agent and skill instructions | `.claude/agents/moai/manager-todo.md` (L31); `.claude/skills/moai-factory-foreman/SKILL.md` (L108-156); `.claude/skills/moai/SKILL.md` (L179); `.claude/skills/moai/workflows/gtd.md` (L25-30, 132) | yes, all four | edit both trees (REQ-DBP-010, -011, -012). Byte-identical pairs: `manager-todo.md`, `workflows/gtd.md`, the foreman skill; divergent pair: `.claude/skills/moai/SKILL.md` |
| Codex agent | `internal/template/templates/.codex/agents/moai/manager-todo.toml` (L28) | template only; no local `.codex/agents/moai/manager-todo.toml` exists | generated file (header: "regenerate, do not edit"); regenerated from `manager-todo.md` by `make agents-emit`, never hand-edited |
| Canonical storage doc | `.moai/docs/todo-queue-storage.md` (L4-38, 120-124) | yes; the twins already differ at lines 25 and 38 | edit both; keep the divergence |
| Local legacy workflow | `.moai/workflows/gtd.md` (L24-29, 66, 126) | none (the template carries only `.moai/workflows/README.md`) | edit local file only |
| Generated codemaps | `.moai/project/codemaps/{data-flow,modules,overview}.md` | none | deferred to the sync-phase regeneration, no hand edit — decision F.5 |
| README | `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` (L66 each) | n/a | REQ-DBP-013 — five statements, compact form |
| docs-site | `content/{en,ko,ja,zh}/advanced/factory-mode.md` (L122), `content/{en,ko,ja,zh}/utility-commands/moai-todo.md` (L63, L252) | n/a | REQ-DBP-013 — five statements, compact form |
| docs-site, inspected, no change | `content/{en,ko,ja,zh}/advanced/moai-web-console.md` (L106) | n/a | lists watched paths for a console; instructs nobody to open a file |
| Historical, untouched | `.moai/learning/2026-09-03-moai-todo.{md,html}`, `CHANGELOG.md` (changed only by the sync phase) | n/a | history is not rewritten |
| Historical, untouched (root `reports/`) | `reports/hooks-audit-20260911-01a08e35/cards-{before,after}.json`, `reports/moai-factory-mcp-matrix-live-20260924.{md,html}` (4 files; in the 40-file count, not in the 36) | n/a | dated report artifacts; not instruction surfaces; never edited |
| Comments only, untouched | `internal/web/screens.templ` (L234) | n/a | no reader is instructed by it |
| Test seed, untouched | `internal/testguard/home_access_seed.txt` | n/a | generated by the home-access guard |
| Go help strings (outside the grep: `*.go` excluded) | `internal/cli/todo.go` (L103, L224), `internal/cli/handoff.go` (L40), `internal/cli/factory_handoff_recover.go` (the `factory` command `Short` at L20; a `Long` is added) | n/a | REQ-DBP-009 |
| Go comments only, untouched (outside the grep) | `internal/cli/doctor_todo_store.go`, `internal/factory/state_dir.go`, `internal/homestate/paths.go` | n/a | no reader is instructed by them |
| Generated artifacts the change regenerates | `internal/template/catalog.yaml` (entry hashes of the foreman skill directory and `manager-todo.md`, rewritten by the `make build` step `gen-catalog-hashes.go --all`); `internal/cli/testdata/doctor-{dark,light,nocolor}.golden` (M5) | n/a | regenerated, committed with the change that caused them; never hand-edited |

Row count check: agent and skill files 8, Codex TOML 1, storage doc 2, local legacy workflow 1, codemaps 3, README 4, docs-site edited 8, docs-site inspected 4, historical 3, comments-only 1, test seed 1 = 36 non-Go files; the root `reports/` row adds 4, which is the 40 of the grep without the `reports/` exclusion. The Go rows and the generated-artifacts row are outside the grep.

## D. Constraints

- **Ordering with card t1641** (gtd to todo rename): t1641 edits several §C files and starts after this card lands. This SPEC therefore (a) writes every new sentence with the canonical `moai todo` spelling, (b) does not edit any existing `gtd` word, and (c) keys the drift guard (REQ-DBP-014) on the pointer sentence, which contains neither `gtd` nor `todo`.
- Template-First: template trees are the source; local `.claude/` and `.moai/` files are updated by hand where the twin is not byte-identical (§C). Generated files (the Codex TOML, `catalog.yaml`, the doctor goldens) follow their regeneration verbs and are never hand-edited.
- Instruction text is English; locale prose appears only in README and docs-site, canonical locale `ko`, no emoji (hns-oss-docs-i18n-rules).
- Tests never touch the real home: they set an absolute `MOAI_HOME` under `t.TempDir()`; the home-access guard (`internal/testguard`) must stay green.
- The command is read-only by construction; its tests include a positive control that proves a file-creating call would be detected.

## E. Out of Scope

### Out of Scope — the gtd to todo rename

- Renaming `moai gtd`, rewriting `gtd` words in the §C files, or the alias-warning release. That is card t1641.

### Out of Scope — other project-keyed stores

- The search cache `sessions.db` (`homestate.SearchDBPath`), the contract receipt store, the decisions directory, and the run directory. They are not named by the card and can be added to `moai paths` later without changing its contract.

### Out of Scope — changing where state lives

- Any change to `ProjectKey`, `CanonicalProjectRoot`, `StateDirForRoot`, `resolveStateDir`, the temp-origin guard, or the adoption policy. This SPEC reports locations; it does not move them. The `internal/factory` package gains one new source file holding a pure read-only wrapper over the existing resolvers, its test, and edits to two existing foreman-script test files that provision a `moai` stub (plan § B M3 step 7); no existing non-test file of the package changes.

### Out of Scope — removing ghost files

- Deleting or repairing zero-byte files. The operator already removed the eight; the doctor check only reports.

### Out of Scope — alias verbs and tool surfaces

- `moai todo where`, `moai factory where`, and an MCP tool for paths. One spelling keeps the guard marker single; an MCP tool would also require a catalogue update in `moai-mcp-tools.md`.

### Out of Scope — hand-editing generated codemaps

- Inserting the five statements into `.moai/project/codemaps/*.md` by hand. Decision F.5 defers them to the sync-phase regeneration and records why.

## F. Decisions recorded at plan time

1. **Command name: `moai paths`** (top-level). It serves three stores with one resolver call chain, it is not inside the surface t1641 renames, and the project key belongs to neither todo nor factory. Discoverability from the two verbs that own the stores is provided by REQ-DBP-009. Rejected: `moai todo where` plus `moai factory where` (two implementations, two spellings in every instruction block, and the todo half sits in the renamed namespace).
2. **Doctor check: include**, as the last, lowest-priority milestone. Instructions are a probabilistic fix; a ghost file reappearing is the only mechanical signal that they failed. The existing `Todo Ghost Inventory` check (`internal/cli/doctor_todo_ghost.go`) is the precedent; the cost is one binary-lag registration and a golden regeneration (REQ-DBP-016). The milestone can be dropped without affecting the command or the instruction work (acceptance.md AC-DBP-016 states the dropped-M5 variant).
3. **Tier M, disclosed deviation**: estimated 400-700 lines of Go and tests (below 1000) and 16 requirements and 16 acceptance criteria (both at the Tier M ceiling of 16). The affected-file count is well above the Tier M guidance of 5-15 (12 S12 files, 12 README and docs-site pages, about 9 Go files, 3 doctor goldens, `catalog.yaml`), but only through mechanical twin and locale edits; Tier L's `design.md` and `research.md` would restate this document and the ceilings on requirements and criteria would not change the work. The deviation selects the Tier M threshold 0.80 rather than 0.85 and is accepted knowingly; if a later revision adds a requirement or a criterion, tier up instead.
4. **Shared block, not shared file**: the five statements of REQ-DBP-010 are written once in full in `.moai/docs/todo-queue-storage.md` and carried in compact form (the pointer sentence plus four short lines) in each agent-facing surface, because an agent does not necessarily open the doc. README and docs-site carry the same five statements in locale prose (REQ-DBP-013).
5. **Codemaps: documented deferral, reason recorded.** The card lists codemaps among the surfaces that name a database path. The three codemap files are generated by `/moai codemaps` from the code; a block inserted by hand is erased by the next regeneration, and a regeneration cannot produce the five statements because the generator derives text from code, not from this SPEC. The disposition is therefore: no hand edit in the run phase; the sync phase regenerates the codemaps (already a sync deliverable) and records in its verdict whether each of the three files still names a database path and whether it then points at `moai paths`; a codemap that names a path without the pointer after regeneration is reported to the operator as follow-up card input, not repaired by hand. This is a deviation from the card's "every surface" wording, taken because the alternative would be a requirement that cannot stay true. It is checked at sync, not at run-phase close.
6. **Queue-path contract: canonical directory, disclosed legacy state.** Two contracts were possible: report where the queue is read from right now (the pure resolver), or where it will live (the canonical directory). The canonical directory was chosen because (a) the foreman watch derives its polling directory from `moai paths --field todo-dir` and must keep watching the same place after adoption relocates a legacy queue, whereas a legacy directory vanishes at the relocation; (b) it is the value the adopting verbs leave behind, so an agent never gets a path that is about to change; (c) a reader who needs the data now is served by the disclosed `queue_state` and `legacy_dir` fields instead of a different `path`. The cost is that a database path can be reported with `exists: false` while a legacy copy exists, which is exactly what `pre-adoption` states.

## G. Cross-references

- SPEC-HOME-STATE-ROLLOUT-001 — home-state layout and project key origin.
- SPEC-TODO-QUEUE-HOME-CANON-001, SPEC-TODO-STALE-STORE-001, SPEC-TODO-SURFACE-POLISH-001 — canonical queue location, stale-store disclosure, ghost inventory check.
- SPEC-TODO-HOME-TEMP-GUARD-001 — the temp-origin project-local layout.
- SPEC-FACTORY-LEASE-RENEW-001 — `factoryRootOf`, the factory root rule the command reuses.
- `.claude/rules/moai/development/verification-completeness.md` — observed-failure completion, used in acceptance.md.

## H. Root-rule alignment (recorded at sync)

- **Root-rule divergence between `moai paths` and the factory verbs (audit finding F1) — RESOLVED by card t1862.** The two factory verbs that opened the factory at the working directory (`os.Getwd()`) were aligned to the moai-paths root: both now open the factory through `factoryCardRoot()`, the same rule `moai paths`, the todo verbs, and every other factory-opening surface apply. The boundary (a different cwd plus `CLAUDE_PROJECT_DIR` naming a different project) is pinned by `internal/cli/factory_root_alignment_test.go` on the t1862 card branch. No divergence remains.
