# SPEC-LAUNCHER-ENTRY-FLAGS-001 — Design (Tier L, v0.4.0)

Design for the requirements in `spec.md` §B, on the facts in `research.md`. It states WHAT changes in each surface and WHY the shape is safe; it carries no function bodies. Every decision here follows a recorded verdict (`decision-index.md` Q1–Q17); the author choices that were not verdicts are listed in `spec.md` §D and repeated where they bite.

## §1 — Principles

1. **Reuse before rebuild.** `-l` is parsed where `-f` and the former `-l` short are parsed (`parseFactoryFlag`) and desugars onto the path the removed `-f lane` used (`NextFactoryLaneNumber`, then the shared claim). No parse, claim, capacity, discovery, or environment logic is duplicated.
2. **Add the new entry, prove the factory, remove the mode, then rename.** The kanban removal is sequenced after `-f`/`-l` is in place and after a factory safety net has been observed green AND observed red on a mutant (§7). Renames come after removals so no symbol is renamed and then deleted. No milestone deletes or renames a file the factory path still reaches without the net green before and after.
3. **Refuse, do not warn.** A removed spelling is refused with one line naming the replacement, before any branch resolves, nothing launched and nothing written — the repository precedent for retired flag spellings (`internal/cli/factory.go:235-251`, `:310-323`). No migration command, no deprecation window.
4. **Names are code; strings are contracts.** A Go identifier is free to rename; a string a process, a disk, or another version reads is frozen unless a verdict renames it. The six factory-read marker values are frozen (REQ-015); everything else that carried the word is renamed (REQ-017). §8 lists every persisted or wire name with its verdict.
5. **Separate "mode" from "package".** The package `internal/kanban` was never only the mode: it holds the todo queue, factory slots, locks, and landing code. This SPEC deletes the mode's code (REQ-011) and then renames the package that remains (REQ-017).

## §2 — The entry grammar

| Form | Behavior | Notes |
|------|----------|-------|
| `moai cc\|glm -f`, `--factory` | factory leader, unchanged | derived capacity (`internal/config/defaults.go:708`) |
| `moai cc\|glm\|codex -l`, `--lane` | factory lane, next free slot via the shared claim | codex: the supervising relaunch lane (`internal/cli/codex_launcher.go:897`, claim `:931`) |
| `-l <anything>`, `-l=<anything>` | refused: `-l` takes no argument; `--leader <name>` selects a leader | the short `-l` no longer means `--leader` |
| `-l` with `-f`, `-k`, or an operator `--name` | refused: one entry token per launch / `-l` already names the role | |
| `-f lane`, `-f lane-<n>`, `--factory=lane`, and `-f`/`--factory` with a lane-shaped `--name` | refused naming `-l` (cc, glm) | the lane branch of the dispatch table loses its `-f` trigger |
| `moai codex -f` in any shape | refused naming `moai codex -l` (lane) and `moai cc -f` / `moai glm -f` (leader) | the fixed Codex refusal line is rewritten; tests pinned to the old line are re-pinned; the `FACTORY_MODE_UNSUPPORTED_BACKEND:` prefix is kept (a Codex leader stays unsupported) |
| `-f <N>` | refused as today, message rewritten | pin; observed already refused (progress.md PV-16) |
| `-k` / `--kanban`, every shape | refused: kanban mode is retired; use `-f` to lead, `-l` to join | cc, glm, codex; the refusal lives in ONE file, `internal/cli/launcher_retired_entries.go` (§4.7) |
| `--leader <name>` | unchanged long form; composes with `-l` only | |

Dispatch: `parseLauncherEntry` stops merging a kanban parse and a factory parse; it parses `-f`, `-l`, and the retired-entry refusals only. The branch table reduces to leader, lane, and none, with the lane triggered by `-l`. The entry-parse type keeps the factory fields and loses the kanban ones (renamed `launcherEntryParse`, §4.7). Refusal text tokens (tests assert tokens, not full sentences): the `-k` line contains `retired`, `-f`, and `-l`; the `-f lane*` line contains `-l`; the codex `-f` line contains `moai codex -l` and `moai cc -f`; the `-l <arg>` line contains `-l`, `no argument`, and `--leader`; none contains a removed form except the one being refused. The `-k` line is the one place that must contain the word "kanban" (it says the mode is retired), which is why §4.7 confines it to one file.

## §3 — The leader notice (verdicts 8 and 17)

Today the notice prints `moai <entry> -f lane-<i>` once per declared lane (`internal/hook/session_start_factory.go:207`), states the count in `leaderManual` and `entryGuide`, and appends a free-slot line built from `FactoryFreeSlots` (`session_start_factory.go:229`; `internal/kanban/factory_slots.go:354-360`), whose only production caller is that line. The string table `session_start_factory_i18n.go` carries 28 references to `leaderFreeSlots`, `leaderSlotsNone`, `laneJoin`, `laneJoinNoCount`, `leaderManual`, `entryGuide` (measured, research.md §R10).

New notice, four locales: ONE sentence stating how to start a lane — to start a lane, enter `moai cc -l`, `moai glm -l`, or `moai codex -l` in a new terminal — and nothing that states, derives, or lists a number: no per-lane line, no count sentence, no free-slot list, no numbered lane label. Consequences, recorded so they are not discovered later:

- The free-slot line and its string-table fields (`leaderFreeSlots`, `leaderSlotsNone`) are removed from the notice builder. `FactoryFreeSlots` stays in the slots package (no deletion of seemingly-unused code without approval); it is left caller-less in production and the run phase reports that.
- The leader's `MOAI_FACTORY_WORKERS` value stays a number: it is the factory discriminator and block-cap signal (`internal/cli/launcher_blockcap_infinite.go:57`), not a display. Only its display is dropped.
- The queue sentence that names the foreman loop is kept under factory names (verdict 15): "the queue is polled and cards are picked by the operator or the factory foreman loop (bare `/loop`)" in en; the ko/ja/zh renderings (`칸반 포어맨`, `かんばんフォアマン`, `看板工头` at `session_start_factory_i18n.go:139,184,229`) become the native renderings of "factory foreman" (humanize pass in M11 scope; strings edited in M4).
- The stale-run rebind hint (`internal/hook/session_stale_run.go:92,103,114,125`) says `moai cc -l` in four locales and can no longer reclaim a specific number. The factory card rejoin errors (`internal/cli/factory_card.go:95,98,101`) name `-l`.
- Tests: `TestFactoryLeadNoticeWorkerCountDrivesLineCount` (observed passing today, progress.md PV-39) is replaced by a text-pinning pair — `TestFactoryLeadNoticePrintsLaneCommandOnce` and `TestFactoryLeadNoticeIsLaneCountIndependent` (the same notice renders byte-identical lane guidance for declared counts 1, 3, and 8, which a count-printing mutant cannot satisfy); `TestFactoryLeadNoticeAllSlotsClaimed` (passing today, PV-39) is deleted with the free-slot line; `TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide` and `TestFactoryGuideNamesWorkerJoinInEveryLocale` are re-pinned.

## §4 — Removing Kanban Mode

### §4.1 What is deleted

| Surface | Deleted | Evidence class |
|---------|---------|----------------|
| `internal/cli/kanban.go` | the kanban-only symbols (`research.md` §R4 left column) | 12 symbols |
| `internal/cli/cc.go`, `glm.go` | the kanban branches of the dispatch and the kanban help text | mixed files |
| `internal/cli/codex_launcher.go` | the kanban refusal constant's wording and the `MOAI_KANBAN_LABEL` stamp, scrub entry, and child-env line (`:316-329`, `:959-961`, `:1037`) | mixed; 4 constant references (progress.md PV-40) |
| `internal/cli/launcher_blockcap_infinite.go` | the `MOAI_KANBAN` / `MOAI_KANBAN_LABEL` clause (the `MOAI_FACTORY_WORKERS` clause stays) | mixed |
| `internal/hook/session_start_kanban.go`, `session_start_kanban_i18n.go` | whole files | 2 |
| `internal/hook/session_start.go` | the kanban notice block `:555-605` and the `kanban_notice` timing lap `:611` | mixed |
| `internal/hook/session_start_record.go` | the companion and kanban-leader branches of the role reader `:180-202` | mixed |
| `internal/kanban` | the board family: `board.go`, `board_store.go`, `board_recover.go`, `board_lock.go`, `board_lock_unix.go`, `board_lock_windows.go`, `board_lock_clear_unix.go`, `board_lock_clear_windows.go`, `column.go`, `reconcile.go` (10) and their 13 tests; the companion symbols in `bootstrap.go` and `role.go` | 10 files + symbols |
| `internal/config/envkeys.go` | constants `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL` | 3 constants |
| `internal/web` | the "Chain session board" panel and its view model and i18n strings | panel |
| Docs | per `spec.md` REQ-022 and §5 below | per page |

### §4.2 What is kept, and why

- `kanban_settings.go` (renamed): called from the factory leader branch (`internal/cli/cc.go:227`); its signal marker is read by the factory leader notice (`session_start_factory.go:241`).
- The shared helpers of `kanban.go` (research §R4 right column) in the renamed helper file, and the entry-parse type.
- The todo queue, GTD, classification, landing, integration lock, slot lease, settings-drift, state directory, status reader, record writer, factory slots and runtime — 49 files plus the three mixed files after their companion symbols are removed — all in the renamed package.
- The `Record` type, its writer, and its JSON keys (REQ-015/§8): factory leaders and lanes write session records; the chain fields `Rung`, `DeepScanDir`, `VerifyRung`, `VerifyReentries` and `revision.go` STAY — the sync-phase dedup gate step reads `verify_rung` by name (`.claude/skills/moai/workflows/sync/quality-gates-quality.md:146`), and no Go code in the repository writes the three chain fields (measured, progress.md PV-41), which the rewritten text states plainly (§6).
- The six factory-read markers and the discovery contract that reads `MOAI_KANBAN_ID` from a live leader process (`internal/discovery/leader_readers_darwin.go:47-62`).
- The always-loaded leader/lane doctrine, renamed `factory-dispatch*.md` and stripped of its kanban-only sections.

### §4.3 The six markers: Go names change, string values do not (verdict 10)

The six factory-read markers are published by the launchers, read by factory code in 95 files (34 production, 61 test), and — for `MOAI_KANBAN_ID` — read across a PROCESS boundary from a live leader's environment by name. Their string values and semantics are frozen; no dual read is added; no rename window exists. What changes is the Go constant names that carry them (table in §4.7). A string outside Go that spells them also stays: the Codex MCP server environment allowlist `internal/codexwiring/configtoml.go:21` lists `MOAI_KANBAN_ID` and `MOAI_KANBAN_BACKEND` literally and keeps both, and the rule/docs text that names the markers (the scrub command in the dispatch rule, `moai-mcp-tools-catalogue.md`) keeps the names while dropping the three removed ones. The frozen-value test (AC-016) pins all six values to literals written in the test file itself, so a constant rename can never move a value unnoticed.

### §4.4 The Codex lane child and `MOAI_KANBAN_LABEL` (verdict 11)

Measured: the stamp is written at `codex_launcher.go:959-961` and `:1037`, scrubbed at `:316-329`, pinned by `factory_m5_test.go:145-146`, and set by three other tests (`codex_factory_retire_test.go:52`, `codex_debug_trace_test.go:120`, `factory_m4_test.go:36`). The code comment at `:953-958` and the test comment at `factory_m5_test.go:142-144` say the factory card verbs read the label. The card verbs read the lane label from `MOAI_FACTORY_WORKER` (`internal/cli/factory_card.go:62`, `:88-91`) and the role from `MOAI_FACTORY_ROLE`; no production reader of the label exists in the repository's Go code (bounded search; readers outside Go are a Gap). The child therefore no longer receives the label, the scrub entry goes with the constant, and `TestSD_AC003_CodexRelaunchPerCard` is re-pinned to assert the label's absence and the presence of `MOAI_FACTORY_WORKER`, role, backend, and card id. A new test runs the card verbs with exactly the child's environment (worker, role, backend, card; no label) and shows admission and stage transitions work — the RED/GREEN pair that proves the Codex lane still launches and identifies itself without the label.

### §4.5 Web console (REQ-018; verdicts 9 and 12)

The "Chain session board" panel, its view model (`ChainRoles`, the role cards), and its strings are removed. The "Factory lanes" panel and the "SPEC pipeline" panel stay on the screen now served at `/factory`. The shell id, title ("Factory"), the live-update area key, the `data-live` markers (including the two on the Todo screen and the overview todo card), the i18n key family `kanban.*` (84 occurrences in `internal/web/assets/i18n.js`, no existing `factory.*` key — PV-42), and the icon id are renamed; the watch map in `events.go` (key at `:35`, path registrations `:223-225`) and the matching `app.js` list (`:614`) change together. `GET /kanban` redirects to `/factory` (author choice 4). The assets are embedded in the binary (`internal/web/assets.go:24`), so server and script always ship together; the one residual is a browser tab opened before the upgrade, which holds the old script until reload (spec.md §A.2 row 13). Old session records with roles `plan`/`run`/`sync` stay readable and are not rendered as a chain.

### §4.6 Rules, skills, agents, catalog (REQ-019/REQ-021)

The three rule files are renamed in the local `.claude` tree and in the template mirror; the kanban-only sections (the board, card classes by column, the phase `/clear` handoff, the companion launch mechanics) are removed, the factory doctrine stays. 52 files outside specs reference the path (research.md §R6): 16 Go/test/fixture files and the rest rules, skills, docs, and mirrors; four tests pin or read it (`workflow_rule_paths_pinned_test.go:31`, `docs_delegation_lane_flow_test.go:70-71`, `contract_mode_guided_test.go:679`, `init_headroom_export_test.go`). `cross-session-messaging-detail.md` is a declared fork pair: both copies are edited by hand and keep the `mirror-fork: intentional` marker. The skill `moai-kanban-foreman` is renamed `moai-factory-foreman` (directory, frontmatter name, catalog entry at `internal/template/catalog.yaml:41-43`, hash regenerated by `gen-catalog-hashes.go` via `Makefile:35`, three catalog tests' counts and comments); the old id is appended to `legacySkillIDs` (`internal/cli/update_archive.go:45`) so `moai update` archives the stale skill in user projects. `.claude/loop.md` keeps its path (the bare-`/loop` driver file name is the runtime's convention) and its content is reworded to the factory foreman. `moai todo --auto` and its CLI help keep their behavior and lose the word.

The three OLD rule files already installed in a user project are handled by the existing update flow, not by a list: every `moai update` that runs the template sync removes the whole `.claude/rules/moai` root (`internal/cli/update/deploy/deploy.go:75-78`, the clean at `:107-185`, the stage at `internal/cli/update_template_sync.go:388-421`), after copying every regular file the embedded template does not carry into the run's pre-clean backup `.moai-backups/<timestamp>/pre-clean/` and aborting the removal if that copy fails. Once the renamed template stops shipping the three old paths they fall into the "not carried" class. Observed against the real embedded template (progress.md PV-55): today a user-modified `kanban-dispatch.md` is NOT backed up (the template carries the path and the deploy rewrites it), while a user-modified file at a path the template does not carry is backed up byte-identical and leaves the live tree. The rule root is MoAI-managed, not a user-owned namespace (`IsUserOwnedNamespace` has no branch for `.claude/rules/`; `plan.go:152-259`), so namespace protection is not engaged and user files outside the managed roots are untouched. The `defs.DeprecatedPaths` route is not usable for these files: its v3 residue sweep deletes after backup regardless of modification, returns an error that aborts the update (`update.go:494-496`), and a registered path that exists counts as a V2 signal (`v2_detection.go:142-144`). The operator's original stricter safeguards (hash-match deletion, retention of modified copies, a named per-file report, step-level fail-open) were weighed against these facts (plan.md §B.2), and the operator chose to rely on the existing clean (OD-17, Option X): no production change, no list, no dedicated step. Under X a user-modified copy of an old rule file is backed up and removed, not retained, and the progress line names a count rather than paths; both are accepted. AC-020's fixture asserts exactly this contract.

### §4.7 The old→new name mapping

**Package, paths, and files**

| Old | New | Notes |
|-----|-----|-------|
| `internal/kanban` (import path, package `kanban`) | `internal/factory` (package `factory`) | 179 importing files, 1,689 qualified references (measured, PV-35); four non-LSP, non-TUI sites declare a local identifier named `factory` (`internal/cli/factory_handoff_recover.go:20`, `internal/hook/role_naming_m3_notice_test.go:129`, `internal/hook/stale_run_m1_test.go:118`, `internal/homestate/runtime_census_test.go:75`) — whether each also imports the package was not read (Gap); a shadow is a compile error and gets a local variable rename, not an import alias |
| `internal/kanban/kanban_helper_test.go` | `internal/factory/factory_helper_test.go` | |
| `internal/cli/kanban.go` | `internal/cli/factory_launch_helpers.go` | after the kanban-only symbols are deleted |
| `internal/cli/kanban_settings.go` (+ `_test`) | `internal/cli/factory_settings.go` (+ `_test`) | |
| `internal/cli/kanban_launch_facts_test.go`, `kanban_lead_name_test.go` | `factory_launch_facts_test.go`, `factory_lead_name_test.go` | factory-shared; kanban-only tests (`kanban_autonomy`, `_companion_name`, `_bootstrap`, `_dispatch`, `_help`) are deleted when their symbols are (classification per file is a run-phase read; Gap) |
| (new) | `internal/cli/launcher_retired_entries.go` | one of FOUR files allowed to carry the word (REQ-017): the `-k`/`--kanban` retired-entry refusal and its sentence |
| `internal/cli/update_archive.go` (existing) | gains the retired skill id `moai-kanban-foreman` in `legacySkillIDs` | the second allowed file: the identifier is the string the archive step looks for |
| `internal/factory/state_dir.go` (moved) | keeps `legacyStateDirName = "kanban"` and its two read paths | the third allowed file (REQ-015) |
| (new) `internal/web/legacy_routes.go` | registers `GET /kanban` → redirect to `/factory` | the fourth allowed file (REQ-018) |
| `internal/hook/session_start_kanban.go`, `_i18n.go` and 4 tests | deleted | notice builders and message table |
| run-time path strings `./internal/kanban` (`migrate_home_state.go:95,187,202,287`, `home_state_coverage.go:55,64`, `queue_path_seam_scan_test.go:41-44`, sample at `evidence_writer.go:73`) | `./internal/factory` | home-state coverage key `"kanban"` (`home_state_coverage.go:64`) becomes `"factory"` |

**Go identifiers**

| Old | New | Frozen value / note |
|-----|-----|---------------------|
| `EnvMoaiKanbanID` | `EnvFactoryRunID` | value `MOAI_KANBAN_ID` frozen |
| `EnvMoaiKanbanLeadAddr` | `EnvFactoryLeadAddr` | value `MOAI_KANBAN_LEAD_ADDR` frozen |
| `EnvMoaiKanbanLeadName` | `EnvFactoryLeadName` | value `MOAI_KANBAN_LEAD_NAME` frozen |
| `EnvMoaiKanbanSettingsInjected` | `EnvFactorySettingsInjected` | value `MOAI_KANBAN_SETTINGS_INJECTED` frozen |
| `EnvMoaiKanbanBackend` | `EnvFactoryBackend` | value `MOAI_KANBAN_BACKEND` frozen |
| `EnvMoaiKanbanCard` | `EnvFactoryCard` | value `MOAI_KANBAN_CARD` frozen |
| `EnvMoaiKanban`, `EnvMoaiKanbanSpec`, `EnvMoaiKanbanLabel` | deleted | no collision with existing `EnvMoaiFactoryWorker(s)`, `EnvFactoryRole`, `EnvFactoryClearPolicy`, `EnvFactoryAutoDispatch` (`envkeys.go:289-389`) |
| `kanbanEntryParse` (field `KanbanEnabled` deleted) | `launcherEntryParse` | |
| `kanbanBranch`, `resolveKanbanBranch` | deleted; the existing `resolveFactoryBranch` carries leader/lane/none | |
| `prepareKanbanSettings` | `prepareFactorySettings` | transient prefix below |
| `kanbanRoleFromEnv` (`session_start_record.go:180`) | `factoryRoleFromEnv` | |
| `kanbanFlagUsageError`, `kanbanUnsupportedBackendSentinel`, `codexKanbanRefusalDiag` | replaced by the retired-entry refusals in `launcher_retired_entries.go` | the sentinel `KANBAN_MODE_UNSUPPORTED_BACKEND` is retired; non-test consumers were not searched (Gap 10) |
| `Kanban(vm, k)`, `KanbanVM`, `KanbanRecord`, `handleKanban`, `buildKanban`, `kanbanViewModel` | `Factory`, `FactoryVM`, `FactoryRecord`, `handleFactory`, `buildFactory` | in `internal/web` |
| timing laps `"kanban_record"` (`session_start.go:475`), `"kanban_notice"` (`:611`) | `"factory_record"`; the notice lap is deleted with the notice | persisted only if a latency log keys on the name (unobserved, Gap 14) |

**Names that are strings or texts**

| Old | New |
|-----|-----|
| temp `--settings` file prefix `moai-kanban` (`kanban_settings.go:89,108`) | `moai-factory` |
| error texts `kanban backlog …`, `kanban: landing …`, `kanban: git log …` (`internal/kanban/backlog_store.go:1000-1013`, `backlog_sqlite.go:197-206`, `landing_verdict.go`, `landing_evidence.go`, `autodone_scan.go`) | `todo queue …`, `factory: landing …`, `factory: git log …` (tests that match the text are re-pinned) |
| web: route `/kanban`, shell id and title, SSE key `kanban`, `data-live="kanban"`, i18n `kanban.*`, icon `"kanban"` | `/factory`, `factory`, `factory`, `data-live="factory"`, `factory.*`, `"factory"` |
| statusline label "Kanban backlog" (`internal/statusline/types.go:252,350`) | "Backlog" (the segment key is already `backlog`) |
| CLI/MCP help text: `moai todo` Short (`todo.go:220`), `moai gtd` Short (`gtd.go:56`), MCP `todo_add` (`mcp_todo.go:32`), `tokens --card` flag text (`tokens.go:441`) | the same sentences without the word |
| skill `moai-kanban-foreman` | `moai-factory-foreman` |
| rules `kanban-dispatch.md`, `-detail.md`, `-mechanics.md` | `factory-dispatch.md`, `-detail.md`, `-mechanics.md` |
| docs pages `advanced/kanban-mode`, `multi-llm/kanban-mode`, `core-concepts/kanban-board-terms` (×4 locales) | per §5 |
| home banner strings `home_kanban_*`, CSS class `cw-home-kanban*`, image `kanban-five-sessions.*` | removed (§5) |
| leader notice foreman sentence (en `session_start_factory_i18n.go:89`; ko/ja/zh `:139,184,229`) | "factory foreman" renderings |

## §5 — Documentation and redirects (REQ-022)

Per-page disposition (author choices 5 and 6; each page is renamed-with-redirect or deleted-with-redirect, in en, ko, ja, and zh in one commit, honoring the native-idiom policy and the humanize step for ko/ja/zh):

| Page (each locale) | Disposition | Redirect |
|--------------------|-------------|----------|
| `advanced/kanban-mode` (25 KB; also carries factory-mode text) | factory text merged into the rewritten `advanced/factory-mode`; the chain-lineage and `moai chain` sections move to a new page `advanced/origin-trail-chain` (the command is live and has no other page); the old page is deleted | `/advanced/kanban-mode` → `/advanced/factory-mode` |
| `multi-llm/kanban-mode` | deleted | → `/advanced/factory-mode` |
| `core-concepts/kanban-board-terms` | deleted (the board is gone) | → `/advanced/factory-mode` |
| `advanced/factory-mode` (11 KB) | rewritten for `-f`/`-l` (the page is currently shadowed by a redirect) | none |

**`docs-site/vercel.json` is reversed, not appended.** It already redirects `/:locale(ko|en|ja|zh)/advanced/factory-mode` and `/advanced/factory-mode` to the kanban page, and `/multi-llm/factory-mode` likewise (`:182-201`). Those four rules are removed (they would shadow the canonical page and, beside the new rules, form a redirect cycle); the new rules follow the file's two-rule pattern (a `:locale(ko|en|ja|zh)` rule and a bare rule to `/ko/…`). The existing `manager-kanban` → `manager-lead` pair (`:202-211`) is unaffected. Menu entries (`docs-site/data/menu/main.yaml:159-162,493-496,725-728`), the three `_meta.yaml` section listings per locale, the home banner (`layouts/index.html:56-76`), the i18n strings (`home_kanban_*` in `i18n/{en,ko,ja,zh-cn}.yaml`), the banner CSS in `static/moai-docs-layout.css`, and the image pair are removed or repointed. The third-party mirrored Claude Code docs under `docs-site/content/*/claude-code/` are out of scope: their native-language uses of the word (a Chinese "monitoring dashboard", a Korean analogy for the Agent Teams task list) are not this mode. READMEs lose the "What's New in v3.1 — Kanban Mode" section and carry the `-f`/`-l` entry forms. `.claude/rules/local/gitflow-lane-protocol.md` (local-only, not mirrored) names the removed `-f lane` lane form and the old rule path and is edited in place.

## §6 — The `workflows/factory.md` rewrite (REQ-020)

The skill documents a different feature from the one the tree ships. What it asserts today, against what the tree observably holds:

| Today's text | Observed |
|--------------|----------|
| `-f` seeds ONE session's plan→run→verify→sync chain, optionally targeting a SPEC | `-f` starts the multi-lane leader and accepts no argument (`factory.go:162-180`, PV-16) |
| a goal preset named `factory_chain` drives it | the name occurs only in `factory.md` itself — 2 lines per tree (`.claude/skills/moai/workflows/factory.md:64,66`) — and in no code or configuration |
| the record lives under `.moai/state/factory/` | `RecordPath` resolves `RuntimeStateDirForRoot(root)/<session>.json` (`record.go:189-191`) — `.moai/state/todo` (or the legacy directory if only that exists) |
| three record fields are "written by orchestrator" | no Go code in the repository writes `deepscan_dir`, `verify_rung`, or `verify_reentries` (PV-41); the sync gate step reads `verify_rung` (`quality-gates-quality.md:146`) |
| `moai cg` rejects Factory Mode with the sentinel | `moai cg` is a retired verb and returns the retirement error (PV-5) |
| "Factory Mode raises the block cap because the chain's goal is armed mid-session" | the factory clause raises it from `MOAI_FACTORY_WORKERS`; no chain goal exists |
| a verify exit gate is part of the factory entry | no Go code names the gate (PV-41); it is a skill-level procedure the orchestrator elects to run |

The rewritten skill asserts exactly these, each checked by AC-021:

- **A1** `-f` on `moai cc`/`moai glm` starts the factory leader; `-l` on `moai cc`/`moai glm`/`moai codex` joins a lane; `moai codex` has no leader entry; no other entry token exists.
- **A2** Neither flag takes an argument; a SPEC identifier or a count after `-f`, and any label after `-l`, is refused.
- **A3** A lane joins a running factory and claims the next free lane slot automatically; slot numbers are not chosen by the operator.
- **A4** Leaders and lanes write a session record at `RuntimeStateDirForRoot(root)/<session-id>.json`; the launcher writes `session_id`, `spec_id`, `role`, `backend`, `entered_at`, `lane`, `card_id`; the three chain fields are defined in the record and are written by no code in this repository.
- **A5** `moai cg` is retired and `moai gpt` does not exist; the retired entries refuse as `AC-008`/`AC-009` observe.
- **A6** The leader raises the Stop-hook block cap through the factory clause of the launcher; no goal preset is named.
- **A7** The verify exit gate (`workflows/run/mode-orchestration.md` § Verify Exit Gate) and the sync dedup gate (`quality-gates-quality.md` Step 0.55.0) are specified in their own sections; no launcher flag enters or arms them. The four dependents (`moai.md:210`, `run.md:52`, `mode-orchestration.md:86` title and opening paragraph, `quality-gates-quality.md:133` applicability sentence) are reworded to stop naming a "factory contract" or a "factory chain"; the gates' own procedures are unchanged. The `Record` code comments (`record.go:47-56,81`, "Kanban state record", "entered Kanban Mode", "the chain heads at plan-phase") are reworded to the same facts.

Choices inside A7 (not verdicts, Kickoff-visible): that the verify gate is retained as specification, and that the dedup gate's trigger sentence becomes "a sync whose run-phase recorded a verify result" (behavior-neutral: the gate defaults to RUN, and today nothing records one).

## §7 — The factory safety net and the sequencing proof

M1 authors, BEFORE any production change, tests that exercise the factory path through the existing launch seam (`installFactoryLaunchSeam`) and observes them green: leader launch and run record; lane launch with the shared claim and markers; settings injection; block-cap raise on the factory clause; session record write for a leader and a lane; SessionStart factory notices; live-leader discovery from the marker; the enterable-pair matrix (§ AC-015). Each is then observed RED on a mutant — a scratch change that removes the settings-injection call, the factory block-cap clause, a marker publish, or the discovery read. An instrument whose red has never been seen has proven nothing. Existing tests already green on this tree and reused: `TestCCFactoryEntryRecordsFailOpenRunMetadata`, `TestCCFactoryLaneJoinsDiscoveredLeader`, `TestGLMFactoryLaneJoinsDiscoveredLeader`, `TestPrepareKanbanSettingsWritesTransientFile` (renamed at M7), `TestDiscoverLeaderVerifiesLiveLeader`, `TestDiscoverLeaderDeclinesUnparseableRunID` (PV-43), and the hook notice tests.

Sequence, each step mergeable alone with the net green before and after: M1 net, M2 cc/glm grammar, M3 codex grammar, M4 notice → the `-f`/`-l` entry is complete with kanban still present; M5a launcher `-k` refusal and removal, M5b hook removal, M6 board family → kanban mode is gone; M7 identifier and file renames inside packages (string values untouched), M8 package rename, M9 web, M10 rules/skills/loop/catalog/constitution/factory.md, M11 docs. Renames follow removals so no symbol is renamed and then deleted; the package rename (M8) is separated from the identifier renames (M7) so a compile failure points at one cause; the web screen (M9) follows M8 because it imports the package.

**Cross-check — what stays enterable (the verdicts' item e).** After every removal and rename: Claude and GLM leader (`-f`), Claude, GLM and Codex lane (`-l`); Codex leader refused. AC-015's matrix test launches each through the seam and reads the markers; AC-016 reads the leader marker through the discovery reader; AC-013 shows the Codex lane identifies itself without the label. A lane launched by the new binary discovers a leader started by an older binary because the marker values are frozen.

## §8 — Persisted and wire names: renamed or frozen

| Name | Verdict | Evidence that renaming would (not) break data or versions |
|------|---------|-----------------------------------------------------------|
| six marker string values | FROZEN | read across a process boundary and by older binaries; rename would break cross-version discovery |
| `MOAI_KANBAN`, `_SPEC`, `_LABEL` | REMOVED | publishers and the only stamp removed; no factory reader (PV-40); surviving old sessions keep them in their own environment and nothing reads them |
| legacy state-directory literal `kanban` (`state_dir.go:23`, read paths `:57,85`) | FROZEN | it exists to read data older versions wrote under `.moai/state/kanban` |
| queue/state directory `todo` (project-local and `~/.moai/db/<key>/todo`) | already neutral, unchanged | |
| `.moai/state/kanban-board` | REMOVED (leftover ignored) | no non-test reader outside the package (PV-25) |
| session record JSON keys and role values | FROZEN (no key carries the word) | old records with roles `plan`/`run`/`sync` remain readable (REQ-016) |
| leader socket address dirs | kanban dir REMOVED; `/tmp/moai-socket-factory` unchanged | `internal/kanban/bootstrap.go:402-432`; an address line, not a filesystem contract |
| backend values `claude`/`glm`/`gpt` | FROZEN | |
| web route, SSE key, `data-live`, i18n keys, icon id | RENAMED | embedded assets ship with the server; `/kanban` redirects; browser-persisted key `moai-console-lang` unaffected; residual stale tab |
| Go package path, identifiers, files, error texts, temp prefix | RENAMED | no persisted form found; error texts pinned by tests are re-pinned |
| `.claude/loop.md` path | unchanged (runtime convention); content reworded | |
| distributed rule files | RENAMED; stale old copies in user projects | removed and backed up by the managed-root clean (§4.6); no new mechanism (OD-17, Option X chosen) |
| distributed skill | RENAMED; old id archived by `legacySkillIDs` | |

Every row is decided; the distributed rule files are settled by OD-17 (Option X).

## §9 — Risks

| Risk | Mitigation |
|------|-----------|
| A removal or rename breaks a factory dependency the searches missed | the M1 net with mutant probes, re-run per step; the factory-read markers are measured; M5a/M5b/M6/M7/M8 are separate milestones |
| A rename moves a frozen string | AC-016 pins the six values to literals in the test file; the Go grep (AC-018) allows the frozen literals only where it expects them |
| `go vet ./...` over 179 importers is heavy | taken under a resource slot lease; the package rename is mechanical and compile-checked |
| A local variable named `factory` shadows the renamed package | compile-detected; five known sites (§4.7) |
| Operators on `-k` and the removed lane spellings break | the refusal names the replacement; a release note at sync lists the removals; docs state that a lane carries a card whole where a companion owned a phase |
| The pinned Codex refusal line is byte-compared by completed SPECs' tests | M3 re-pins them; partial-supersession recorded at sync |
| Always-loaded edits drift from pinned tests; two template guards SKIP on this tree | the pinned tests that run are re-run; the two skipped guards are not relied on |
| Constitution slot edit touches a section holding Frozen clauses | `agent-common-protocol.md:27` sits in the section anchored `#user-interaction-boundary` (registered Frozen clauses `CONST-V3R2-036..038`, whose registered strings the edit does not touch); `moai constitution validate` runs before and after each of the four edits (AC-022) |
| Four-locale docs diverge; template mirrors drift; Codex TOML embed drift | per-file pair edits; AC-023 parity greps; `make embed-check`; one commit |
| Redirect cycle or dead redirect target | `vercel.json` rules reversed (§5); AC-023 checks no rule destination is a removed page and no source/destination pair cycles |
| Deleting user files in user projects (the three old rule files, possibly user-modified) | removal stays inside the MoAI-managed `.claude/rules/moai` root; a user-owned namespace and `.claude/rules/local/` are never in it; every file the template does not carry is copied byte-identical to the pre-clean backup before removal and a failed backup aborts the removal (PV-55, PV-56); the fixture of AC-020 asserts bystander files survive, the backup is byte-identical for modified and unmodified copies, an absent file is a no-op, and a second run changes nothing; what is NOT bounded, and the operator has accepted under Option X: a modified copy leaves the live tree (recoverable from the backup) rather than being retained, and deletion is not hash-conditional |
| A parallel session writes the same tree during a long run | the run phase works in its own worktree under the one-writer-per-tree rule |
