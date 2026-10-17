---
id: SPEC-MOAI-DB-PATHS-001
title: "Confirmed database paths — a read-only path-reporting command, guess-proof instruction surfaces, and a ghost-file doctor check"
version: "0.1.0"
status: draft
created: 2026-10-17
updated: 2026-10-17
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

Three resolvers produce the stores, and they do not share one rule for the temporary-origin layout (read, not executed, at `93f4fc75c`; the run phase proves the claims by test):

- Queue: `factory.ResolveTodoQueueRoot` (`internal/factory/todo_root.go:81`) then `factory.StateDirForRoot` (`internal/factory/state_dir.go:~25-45`) — home layout `<moai-home>/db/<key>/todo`; a temp-origin launch without an absolute `MOAI_HOME` stays `<root>/.moai/state/todo`. The database is the sibling `backlog.db` of the queue file (`backlogSQLitePath`, `internal/factory/backlog_sqlite.go:~254`).
- Factory: `homestate.FactoryDBPath` (`internal/homestate/paths.go:345`) via `ProjectDir` (line 225) — home layout `<moai-home>/db/<key>/factory/factory.db`; temp-rooted layout `<canonical>/.moai/db/<key>/factory/factory.db`.
- Broker: `factorymsg.BrokerPath` (`internal/factorymsg/store.go:~148-155`) = `<factory-dir>/messages/<run-id>/broker.db`, run id validated by `safeID`.

Two traps sit beside them: `homestate.OpenFactory` and `factory.FactoryRegistryPath` (`internal/factory/factory_slots.go:55`) call `EnsureProjectLayout`, which creates directories and writes `project.json`; and `FactoryRegistryPath` falls back to `<root>/.moai/state/factory/factory.db` when the home path cannot be resolved (ghost #8 above). A path-reporting command must use neither.

### A.4 Instruction surfaces that name a database path

Measured with a repository grep (command and counts in acceptance.md baseline): 36 non-Go files outside `.moai/specs` and `.moai/reports` mention `backlog.db`, `factory.db`, `broker.db`, `todo.db`, or `project-key`. §C classifies them.

## B. Requirements (GEARS)

### B.1 The path-reporting command

- **REQ-DBP-001**: The CLI shall provide a top-level command `moai paths` that reports, for the project resolved from the current launch directory, the project key, the project state directory, the queue database, the factory database, and the factory broker database.
- **REQ-DBP-002**: The command shall obtain every reported path from the resolver the owning verb itself uses (the todo queue-root and queue-path resolvers, `homestate.FactoryDBPath`, `factorymsg.BrokerPath`) and shall not recompute the project key or join path segments of its own.
- **REQ-DBP-003**: The command shall never create a file or directory, open a database through any SQLite driver, take a lock, or trigger a migration, adoption, or layout-ensuring call; it shall learn existence and size by `stat` and directory read only.
- **REQ-DBP-004**: For each reported database the command shall state whether it exists, its size in bytes, and whether its `-wal` and `-shm` companions exist, with their sizes.
- **REQ-DBP-005**: Where `--json` is given, the command shall write exactly one JSON object to stdout and nothing else; where `--field <name>` is given, it shall write only that field's bare value followed by a newline; otherwise it shall write a human-readable listing.
- **REQ-DBP-006**: While launched inside a linked worktree, the command shall report the same project key and the same paths it reports from the primary checkout.
- **REQ-DBP-007**: When the project resolves to the project-local layout, the command shall report the project-local paths and name the layout; when any path cannot be resolved, it shall report that field as unresolved, exit non-zero, and not substitute a guessed or fallback path.
- **REQ-DBP-008**: When `--run <run-id>` is given, the command shall report that run's broker path, and when the run id is not valid it shall exit non-zero without reporting a path; without `--run` it shall report the broker path template and the count of run directories found under the factory `messages/` directory.
- **REQ-DBP-009**: The `moai todo` and `moai factory` help texts and the `moai handoff` help text that names a database path shall point to `moai paths` as the source of the confirmed path, and the `moai todo` long help shall keep the canonical-store string that `internal/cli/todo_help_store_test.go` pins.

### B.2 The instruction surfaces

- **REQ-DBP-010**: Each deployed or local instruction surface in the §C inventory that names a queue, factory, or broker database path shall carry five statements: (1) the path is obtained from `moai paths`; (2) the forbidden guessed-path list of §A.1; (3) a guessed path is never opened with `sqlite3` because opening creates an empty file; (4) a direct read is made only on a snapshot copy of the database together with its `-wal` and `-shm` files; (5) the CLI is preferred (`moai todo list --json`, `moai factory status`).
- **REQ-DBP-011**: When an inventory surface has a template twin, the change shall be applied to the template file and the local file by separate insertions, and the divergences that already exist between the twins shall be left exactly as they are.
- **REQ-DBP-012**: The factory foreman skill's queue-watch shell shall obtain the queue directory from `moai paths` instead of recomputing the project key, and shall stop with a visible error rather than fall back to a computed path when the command fails.
- **REQ-DBP-013**: The README (four locales) and the docs-site pages in §C shall mention `moai paths` in the same change in all four locales, with the same command spelling.
- **REQ-DBP-014**: A repository test shall fail when an embedded template text file that names a `<project-key>` todo, factory, or broker database path lacks the §B.2 pointer sentence and is not on a reasoned allowlist.

### B.3 The doctor check

- **REQ-DBP-015**: Where `moai doctor` runs, a check named `DB Path Ghosts` shall report at warning severity every zero-byte regular `*.db` file found directly in the moai home, its `db/` directory, the project state directory and its `todo/` and `factory/` subdirectories, `<root>/.moai`, `<root>/.moai/state`, and its `todo/` and `factory/` subdirectories, excluding the canonical database paths reported by `moai paths`; the check shall be read-only, shall never delete or move a file, and shall never fail the doctor run.
- **REQ-DBP-016**: When the `DB Path Ghosts` check is registered, the change shall also register its name in the binary-lag doctor-name contract and regenerate any doctor golden output in the same commit.

## C. Surface inventory (observed)

Classification of the 36 grep hits (command in acceptance.md baseline). "Twin" = a template copy under `internal/template/templates/`.

| Group | Files | Twin | Disposition |
|---|---|---|---|
| Agent and skill instructions | `.claude/agents/moai/manager-todo.md` (L31); `.claude/skills/moai-factory-foreman/SKILL.md` (L108-156); `.claude/skills/moai/SKILL.md` (L179); `.claude/skills/moai/workflows/gtd.md` (L25-30, 132) | yes, all four | edit both trees (REQ-DBP-010, -011, -012) |
| Codex agent | `internal/template/templates/.codex/agents/moai/manager-todo.toml` (L28) | template only; no local `.codex/agents/moai/manager-todo.toml` exists | edit template side only |
| Canonical storage doc | `.moai/docs/todo-queue-storage.md` (L4-38, 120-124) | yes; the twins already differ at lines 25 and 38 | edit both; keep the divergence |
| Local legacy workflow | `.moai/workflows/gtd.md` (L24-29, 66, 126) | none (the template carries only `.moai/workflows/README.md`) | edit local file only |
| Generated codemaps | `.moai/project/codemaps/{data-flow,modules,overview}.md` | none | sync-phase regeneration or a one-line pointer; no hand rewrite of generated bodies |
| README | `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` (L66 each) | n/a | REQ-DBP-013 |
| docs-site | `content/{en,ko,ja,zh}/advanced/factory-mode.md` (L122), `content/{en,ko,ja,zh}/utility-commands/moai-todo.md` (L63, L252) | n/a | REQ-DBP-013 |
| docs-site, inspected, no change | `content/{en,ko,ja,zh}/advanced/moai-web-console.md` (L106) | n/a | lists watched paths for a console; instructs nobody to open a file |
| Go help strings | `internal/cli/todo.go` (L103, L224), `internal/cli/handoff.go` (L40) | n/a | REQ-DBP-009 |
| Historical, untouched | `.moai/learning/2026-09-03-moai-todo.{md,html}`, `CHANGELOG.md` (changed only by the sync phase) | n/a | history is not rewritten |
| Comments only, untouched | `internal/web/screens.templ` (L234), `internal/cli/doctor_todo_store.go`, `internal/factory/state_dir.go`, `internal/homestate/paths.go` | n/a | no reader is instructed by them |
| Test seed, untouched | `internal/testguard/home_access_seed.txt` | n/a | generated by the home-access guard |

## D. Constraints

- **Ordering with card t1641** (gtd to todo rename): t1641 edits several §C files and starts after this card lands. This SPEC therefore (a) writes every new sentence with the canonical `moai todo` spelling, (b) does not edit any existing `gtd` word, and (c) keys the drift guard (REQ-DBP-014) on the pointer sentence, which contains neither `gtd` nor `todo`.
- Template-First: template trees are the source; local `.claude/` and `.moai/` files are updated by hand where the twin is not byte-identical (§C).
- Instruction text is English; locale prose appears only in README and docs-site, canonical locale `ko`, no emoji (hns-oss-docs-i18n-rules).
- Tests never touch the real home: they set an absolute `MOAI_HOME` under `t.TempDir()`; the home-access guard (`internal/testguard`) must stay green.
- The command is read-only by construction; its tests include a positive control that proves a file-creating call would be detected.

## E. Out of Scope

### Out of Scope — the gtd to todo rename

- Renaming `moai gtd`, rewriting `gtd` words in the §C files, or the alias-warning release. That is card t1641.

### Out of Scope — other project-keyed stores

- The search cache `sessions.db` (`homestate.SearchDBPath`), the contract receipt store, the decisions directory, and the run directory. They are not named by the card and can be added to `moai paths` later without changing its contract.

### Out of Scope — changing where state lives

- Any change to `ProjectKey`, `CanonicalProjectRoot`, `StateDirForRoot`, the temp-origin guard, or the adoption policy. This SPEC reports locations; it does not move them.

### Out of Scope — removing ghost files

- Deleting or repairing zero-byte files. The operator already removed the eight; the doctor check only reports.

### Out of Scope — alias verbs and tool surfaces

- `moai todo where`, `moai factory where`, and an MCP tool for paths. One spelling keeps the guard marker single; an MCP tool would also require a catalogue update in `moai-mcp-tools.md`.

## F. Decisions recorded at plan time

1. **Command name: `moai paths`** (top-level). It serves three stores with one resolver call chain, it is not inside the surface t1641 renames, and the project key belongs to neither todo nor factory. Discoverability from the two verbs that own the stores is provided by REQ-DBP-009. Rejected: `moai todo where` plus `moai factory where` (two implementations, two spellings in every instruction block, and the todo half sits in the renamed namespace).
2. **Doctor check: include**, as the last, lowest-priority milestone. Instructions are a probabilistic fix; a ghost file reappearing is the only mechanical signal that they failed. The existing `Todo Ghost Inventory` check (`internal/cli/doctor_todo_ghost.go`) is the precedent; the cost is one binary-lag registration and a golden regeneration (REQ-DBP-016). The milestone can be dropped without affecting the command or the instruction work.
3. **Tier M**: estimated 400-700 lines of Go and tests (below 1000) and 16 requirements (within the Tier M ceiling of 16); the affected-file count exceeds 15 only through mechanical twin and locale edits, so Tier L's design.md and research.md would restate this document.
4. **Shared block, not shared file**: the five statements of REQ-DBP-010 are written once in full in `.moai/docs/todo-queue-storage.md` and carried in compact form (the pointer sentence plus four short lines) in each agent-facing surface, because an agent does not necessarily open the doc.

## G. Cross-references

- SPEC-HOME-STATE-ROLLOUT-001 — home-state layout and project key origin.
- SPEC-TODO-QUEUE-HOME-CANON-001, SPEC-TODO-STALE-STORE-001, SPEC-TODO-SURFACE-POLISH-001 — canonical queue location, stale-store disclosure, ghost inventory check.
- SPEC-TODO-HOME-TEMP-GUARD-001 — the temp-origin project-local layout.
- SPEC-FACTORY-LEASE-RENEW-001 — `factoryRootOf`, the factory root rule the command reuses.
- `.claude/rules/moai/development/verification-completeness.md` — observed-failure completion, used in acceptance.md.
