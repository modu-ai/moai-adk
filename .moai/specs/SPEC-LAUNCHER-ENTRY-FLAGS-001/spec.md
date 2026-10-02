---
id: SPEC-LAUNCHER-ENTRY-FLAGS-001
title: "Launcher entry flags on the backend verbs (-f leader, -l lane), removal of Kanban Mode, and the factory-vocabulary rename of everything that carried the kanban name"
version: "0.8.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/hook, internal/kanban, internal/web, internal/config"
lifecycle: spec-anchored
tags: "launcher, factory, lane, leader, entry-flags, kanban-removal, rename, cc, glm, codex, template-first, docs-mirror"
tier: L
related_specs: [SPEC-FACTORY-MODE-001, SPEC-FACTORY-LANE-JOIN-SOCKET-001, SPEC-CODEX-LANE-SLOTS-001, SPEC-CODEX-FACTORY-RETIRE-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-CODEX-LAUNCHER-001, SPEC-ROLE-NAMING-CODE-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-BOOTSTRAP-001, SPEC-KANBAN-RENAME-001, SPEC-KANBAN-WORKTREE-001]
---

# SPEC-LAUNCHER-ENTRY-FLAGS-001

> Card: **t1399** (Class C design change). Evidence path for the card (lead-read, created at close, not now): `.moai/reports/t1399/verdict.md`.
> Tier **L** — measured: 13 milestones (M0–M11 with M5 split into M5a and M5b; the `manager-lead` threshold is at least 3 milestones AND at least 10 files) and a footprint of 190 non-test Go, templ, and script source files carrying the word "kanban" (63 of them the `internal/kanban` package), 179 files importing that package (66 production, one of them under `cmd`, and 113 test), 310 test files carrying the word, and 157 README, docs-site, rule, skill, template, instruction, and configuration files carrying the word (`research.md` §R1 states each pattern). The decision surface spans launcher grammar, a cross-process environment contract, a state package rename, a web screen, distributed rules and skills, and the constitution-slot files. Tier L work is coordinated by `manager-lead` in the run phase; each integration unit is mergeable alone — M2, M3, and M4 land together as one unit — and the factory safety net stays green across the whole sequence (design.md §7).

## §A — History

- 2026-10-02 (v0.1.0): Card t1399 opened; first draft proposed a verb-less root entry with a pre-Cobra rewrite.
- 2026-10-02 (v0.2.0): First operator verdicts (source: operator answer via AskUserQuestion, lane-3, 2026-10-02): flags ride on the backend verbs; `-l` retired as the `--leader` short; `-f lane`, `-f lane-<n>`, `moai codex -f lane` removed now; `-l` takes no argument; documentation scope is everything that names an entry command.
- 2026-10-02 (v0.3.0): More verdicts (same source): explicit-name lane spelling refused; notice prints the lane command once; **Kanban Mode removed**, whole extent, in this SPEC, Tier L.
- 2026-10-02 (v0.4.0): Final verdicts (same source, one lane-orchestrator ruling marked as such): the six factory-read `MOAI_KANBAN_*` string values are KEPT; the `MOAI_KANBAN_LABEL` stamp is REMOVED; everything that carries "kanban" and survives is RENAMED including the Go package `internal/kanban`; the kanban-dispatch rules are renamed; `workflows/factory.md` and the chain-contract text are REWRITTEN to today's behavior; the foreman skill and loop driver are KEPT under factory names; the constitution-slot sentences are edited directly (orchestrator ruling); the leader notice drops the lane count. The cleanup of renamed distributed rule files already installed in user projects was raised as a last question (OD-17) and is settled in v0.6.0.
- 2026-10-02 (v0.5.0): OD-15 CONFIRMED by the operator ("proceed as is") with its stop condition kept. The operator first chose Option A for OD-17 (a retired-rule-file cleanup step in `moai update` with five safeguards); reading the update code showed that every `moai update` that runs the template sync already removes the whole MoAI-managed `.claude/rules/moai` root and backs up first every file the template does not carry, so the three old files are already removed and backed up once the template stops shipping them, and three of the five safeguards cannot be built as stated without changing that managed-root contract. That finding was returned to the operator, and REQ-019 and AC-020 were written to the observed contract.
- 2026-10-02 (v0.6.0): Final bookkeeping (same source). OD-17 / Q18 settled: the operator, shown the code facts, chose Option X — rely on the existing managed-root clean, with no production change and no dedicated retired-rule step; the fixture test `TestUpdateRemovesRetiredRuleFilesWithBackup` is the proof. Under X a user-modified copy of an old rule file is backed up and removed from the live tree, not retained, and this is accepted; AC-018's retired-name allowlist is unchanged. Q19 completed: rows 7, 8, 11, and 12 of §D, which the first acceptance had not covered, are accepted, so every row of §D is operator-accepted. No verdict was open at v0.6.0.
- 2026-10-02 (v0.7.0): Plan-audit iteration 1 (FAIL, 0.72; findings D1–D28) re-measured on this tree and addressed; no operator verdict is changed or reopened. Order and classification fixes: the constants `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, and `MOAI_KANBAN_LABEL` leave in M5b together with their last readers (the hook notice, the hook record role reader, the test-harness scrub list), not in M5a; `exportKanbanLaunchFacts` is factory-shared (the glm factory branches, the cc factory branches through `exportFactoryLaunchFacts`, and both Codex entries call it, and it is the only publisher of the frozen backend marker), so M5a keeps it and drops only the `MOAI_KANBAN_SPEC` write, and M7 collapses it into `exportFactoryLaunchFacts`; M2, M3, and M4 are one integration unit because the strings that teach the replacement forms land with the refusals. Coverage fixes: `--lane` long-form refusals in AC-003 and AC-005, a second and third lane-number fixture and a claim-row assertion in AC-001, a RED-now for the explicit-name rows in AC-006, a removed-form grep, `AGENTS.local.md`, and a file-name search in AC-023, eight observed-red mutants for the net in AC-015, a pinned positive-control form in AC-022, the base-ref guards named in AC-020; AC-025 becomes the docs exit gate (the windows cross-build moves into AC-018, so the ceiling of 25 criteria holds). The requirement lines are reformatted for the audit collector; REQ-003, REQ-004, REQ-019, and REQ-022 are reworded (no requirement added: 24 of 25). Three questions the audit marked as operator judgments are recorded in decision-index.md as Q20–Q22, non-gating, with the SPEC written to the smallest-footprint reading of each.
- 2026-10-02 (v0.8.0): Plan-audit iteration 2 (FAIL, 0.79; findings D29–D45) addressed; iteration 3 is the last audit. NEW operator verdict (source: operator answer via AskUserQuestion, lane-3, 2026-10-02; decision-index Q23): the role-declaration carrier in `internal/kanban/role.go` (`RoleDeclaration`, `DeclareRole`, `ResolveDeclaredRole`, and their helper `roleDeclarationDir`) is deleted together with Kanban Mode, and the operator's approval of that deletion satisfies `AGENTS.md` §5 (no deleting of seemingly-unused code without explicit approval) for this carrier; the on-disk `.moai/state/kanban-board/roles` directory is left in place, code removed and files untouched. No earlier verdict is changed. The deletion-order proof is replaced: v0.7.0's caller-grep table was wrong twice, so v0.8.0 proves the order by compilation — a committed runner (`probe/probe.go` with its stage data) replays M2 through M10 on a scratch copy of the Go tree and compiles after every stage on darwin and `GOOS=windows`, with `go vet ./...` (progress.md PV-73 to PV-90). Compiler output found and fixed these order errors: the five `board_lock*.go` files, the wait-budget block of `board_store.go`, and `ClearStaleReport` are a shared file-lock primitive (todo queue, integration lock, slot lease), so M6 re-homes them under `state_lock*` as its first step, before the board is deleted; `langEnglish` and `operatorLang` move out of the kanban notice file in M5b; five test helpers move before the test files that define them are deleted; `leadFlagShort` leaves in M3, not M2; `kanbanSocketDir` leaves in M6 and `codexKanbanRefusalDiag` in M5a; the file `role.go` stays because the factory reads four other symbols in it (the verdict deletes the carrier, not those); and the chain view model is read by the Overview screen as well as the factory panel (Q24, open). Requirement and criterion fixes: REQ-003 excepts the removed label; REQ-011 names the board state store and the role carrier and keeps the shared lock substrate; REQ-019 is split and REQ-025 carries the update behavior (25 of 25 requirements); REQ-021 takes the event-detected form; AC-003 gains three shapes and AC-005 the `--leader` refusal rows without a new criterion (25 of 25); AC-018 closes at M10 and AC-023 widens its removed-form grep to the docs spellings of the retired `-l` short. Q20–Q22 stay open and non-gating; Q24 is added, open and non-gating.

### §A.1 Verdicts (decision-index.md Q1–Q19 and Q23; the table mapping verdict number, plan.md OD number, and decision-index Q number is at the top of decision-index.md)

1. (Superseded by 9) New tokens only; `-k` untouched.
2. Grammar: leader `moai cc|glm -f`; lane `moai cc|glm|codex -l`; no verb-less root form; the verb carries the backend.
3. The short `-l` of `--leader <name>` is retired; the long `--leader <name>` stays.
4. Removed now, refused naming the canonical form: `moai cc|glm -f lane`, `-f lane-<n>`, `moai codex -f lane`; `-f <N>` is a pin (already refused).
5. `-l` takes no argument; `-l lane-<n>` and any argument form is refused naming `-l`; slot assignment stays automatic through the shared claim.
6. Documentation scope is everything that mentions an entry command.
7. The explicit-name lane spelling (`-f --name lane-<n>` and equivalents) is refused naming `-l`.
8. (Superseded in count by 17) The leader notice prints the lane command once.
9. Kanban Mode is removed — whole extent — in this SPEC; Tier L; `-k` refused in every shape, naming `-f`/`-l`.
10. Keep the names and string values of the six factory-read markers; no rename, no dual read.
11. Remove the `MOAI_KANBAN_LABEL` stamp on the Codex lane child.
12. Rename everything that carries "kanban" and survives, including the Go package, web, statusline, state names, and types, with a stated old→new mapping.
13. Rename the three `kanban-dispatch*.md` rules to factory names (both trees), strip kanban-only sections, re-point references and pinned tests; docs pages renamed with redirects or deleted, per page.
14. Rewrite `workflows/factory.md`, the `moai.md` chain-contract sentence, and the Record chain-field documentation to match today's behavior.
15. Keep the foreman skill, `.claude/loop.md`, `moai todo --auto`, and the leader notice's queue sentence under factory names.
16. Constitution-slot sentences edited directly with `moai constitution validate` before and after each edit. Raised as a lane-orchestrator ruling on evidence; CONFIRMED by the operator ("proceed as is"). Stop condition kept: any validate failure, or any need to touch a registered clause string (`CONST-V3R2-036..038` in the section anchored `#user-interaction-boundary`), is a blocker returned to the orchestrator.
17. The leader notice drops the lane COUNT entirely and states only the command to start a lane.
18. Stale renamed rule files in user projects (OD-17): the operator first chose Option A — a retired-rule-file cleanup step in `moai update` with five safeguards. The update code showed the premise does not hold and three safeguards cannot be built without changing the managed-root contract (§A.2 rows 18 and 21–23). Shown those facts, the operator chose **Option X**: rely on the existing managed-root clean (pre-clean backup, then removal), no production change, no dedicated step; the fixture test is the proof. Under X a user-modified copy of an old rule file is backed up and removed, not retained, and this is accepted.
19. All author choices listed in §D are accepted as written — rows 2–6, 9, and 10 first, rows 7, 8, 11, and 12 afterward — and the row 1 ruling is confirmed, so every row of §D is operator-accepted.
20. The role-declaration carrier in `internal/kanban/role.go` — `RoleDeclaration`, `DeclareRole`, `ResolveDeclaredRole`, and the helper `roleDeclarationDir`, whose only non-test caller is `board_store.go:200` inside the board family — is deleted together with Kanban Mode. The operator's approval of this deletion satisfies `AGENTS.md` §5 for it. The on-disk `.moai/state/kanban-board/roles` directory is left in place: code removed, files untouched (design.md §8). Compile-proved reading (progress.md PV-80): the file stays, because `RoleLeader`, `RoleLane`, `IsLegacyLeaderSpelling`, and `legacyLeaderSpelling` in it have factory readers; only the four carrier symbols leave.

### §A.2 Measured current surface (tree `a6d3e6fd4`)

Entry grammar and kanban coupling (unchanged from v0.3.0; full tables in `research.md`):

| # | Observed fact | Evidence |
|---|---|---|
| 1 | `moai cc`/`moai glm` parse their flags by hand; `-f`, `-f lane`, `-f lane-<n>`, `-k` shapes are accepted; `-f lane` is accepted (`TestCCFactoryLaneJoinsDiscoveredLeader` passes). | `internal/cli/cc.go:23-135`; `internal/cli/factory.go:131-252`; ledger RED-4 |
| 2 | `-f <N>` is already refused on cc, glm, codex (exit 1); `-l` today is the `--leader` short; on `moai codex` only `-f lane` is accepted. | `internal/cli/factory.go:88-90,162-180,251`; `internal/cli/codex_launcher.go:700-728,773-780`; ledger RED-1..6 |
| 3 | `-f` with a lane-shaped `--name` selects the lane branch: `-f --name lane-2`, `-f -n lane-2`, and `--factory --name=lane-2` each route to the lane branch, and `-f --name leader-r7` routes to the leader (non-persistent probe, progress.md PV-60). `--lane` is not a launcher flag today: the cc and glm parse leaves `--lane` and its value in the passthrough rest without an error, and `moai codex --lane …` fails with the generic unknown-verb usage line (PV-60). | `internal/cli/factory.go:624-642`; `internal/cli/cc.go:203-205`; ledger RED-11, RED-12, RED-13 |
| 4 | `kanban.go` holds 27 function and type declarations, 9 kanban-only and 18 factory-shared; `exportKanbanLaunchFacts` is among the factory-shared (it publishes the frozen backend marker and is called from the glm factory leader and lane branches, from the cc factory branches through `exportFactoryLaunchFacts`, and from both Codex entries — PV-61); `kanban_settings.go` is called from the factory leader branch; the block-cap raise has separable kanban and factory clauses; the SessionStart hook has independent factory and kanban blocks. | `research.md` §R4–R5, §R16; `internal/cli/glm.go:271,304`; `internal/cli/cc.go:222,227,260`; `internal/cli/launcher_blockcap_infinite.go:41-60`; `internal/hook/session_start.go:511-542,555-605` |
| 5 | Six `MOAI_KANBAN*` markers are read by factory code (92 files, 31 production and 61 test, by constant name or literal — `grep -rlE` over the six constants and the six literals, progress.md PV-87); a joining lane reads `MOAI_KANBAN_ID` from a live leader process's environment by name. | `internal/config/envkeys.go:182-273`; `internal/discovery/leader_readers_darwin.go:47-62`; `internal/discovery/factory_discovery.go:171-178` |
| 6 | The board state store has no non-test caller outside `internal/kanban`, but five of its ten files are not board code: `board_lock.go`, `board_lock_unix.go`, `board_lock_windows.go`, `board_lock_clear_unix.go`, and `board_lock_clear_windows.go`, together with the wait-budget block of `board_store.go` (seven constants including `boardLockWaitBudget`, and the function `boardLockRetryWait`), are the file-lock substrate that the todo queue (`backlog_store.go:1254-1272`), the integration lock (`integration_lock_mutation*.go`), and the slot lease (`slot_lease*.go`) acquire through; deleting the ten files as listed at v0.7.0 fails the build with 25 compiler errors (progress.md PV-75). The package also holds the todo queue, factory slots, locks, and landing (63 non-test files: 5 board-only, 5 shared lock-substrate, 1 conditional, 3 mixed, 4 factory-specific, 45 other). | `research.md` §R3; progress.md PV-75 |

Persisted names and wire contracts (new in v0.4.0; each says whether the SPEC renames or freezes it):

| # | Name | Where it lives | Disposition |
|---|------|----------------|-------------|
| 7 | `MOAI_KANBAN_ID`, `_LEAD_ADDR`, `_LEAD_NAME`, `_SETTINGS_INJECTED`, `_BACKEND`, `_CARD` (string values) | process environment; one is read across processes by name | FROZEN (verdict 10); Go constant names renamed (design §4.7) |
| 8 | `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL` | kanban chain and companion markers; LABEL stamped on the Codex lane child | REMOVED (publishers and the stamp leave in M5a; the constants and their last readers leave in M5b) |
| 9 | Project-local queue and record directory: canonical name is already `todo` (`.moai/state/todo`, `~/.moai/db/<project-key>/todo/`); the old name `kanban` survives only as the legacy fallback literal `legacyStateDirName` that reads data written by older versions | disk | the canonical names are not renamed (already factory-neutral); the legacy literal is FROZEN (it exists to read old data) — `internal/kanban/state_dir.go:20-24,40-44,59-61` |
| 10 | Kanban board directory `.moai/state/kanban-board`, which also holds the role declarations under `roles/` (the carrier of verdict 20) | disk (written by board code and by `DeclareRole`, with no external caller) | the code that writes and reads it is removed; the directory and its files are left in place and ignored (REQ-016, verdict 20) — `internal/kanban/board.go:27`, `internal/kanban/role.go` |
| 11 | Leader socket address directories: `/tmp/moai-socket-kanban` (kanban leader) and `/tmp/moai-socket-factory` (factory leader) — a conventional address line, not a filesystem contract | environment value | kanban directory removed; factory directory unchanged — `internal/kanban/bootstrap.go:402-432` |
| 12 | Session record JSON keys (`session_id`, `spec_id`, `role`, `backend`, `entered_at`, `deepscan_dir`, `verify_rung`, `verify_reentries`, `lane`, `card_id`) — none carries "kanban" | disk, read by older and newer binaries | FROZEN (unchanged); records written by old kanban sessions (role `plan`/`run`/`sync`) stay readable | `internal/kanban/record.go:57-120` |
| 13 | Web live-update area key `kanban`, `data-live="kanban"`, i18n keys `kanban.*` (84 occurrences in `i18n.js`), route `/kanban`, nav and screen labels | web assets embedded in the binary (`go:embed`), so server and client ship together; the only browser-persisted key is `moai-console-lang` | RENAMED with the screen; `/kanban` kept as a GET redirect for bookmarks; a browser tab opened before an upgrade holds old script (residual, §G) — `internal/web/assets.go:24`; `internal/web/events.go:33-35,223-225`; `internal/web/assets/app.js:614`; `internal/web/app.go:163` |
| 14 | Statusline segment key is `backlog` (no kanban in any key); "Kanban" appears in comments | statusline config and output | no key to rename; comments reworded — `internal/statusline/types.go:350` |
| 15 | CLI and MCP descriptive text: `moai todo` Short, `moai gtd` Short, MCP `todo_add` description, `tokens --card` flag text; MCP tool NAMES carry no kanban | help text, MCP descriptions | RENAMED (text only) — `internal/cli/todo.go:220`; `internal/cli/gtd.go:56`; `internal/cli/mcp_todo.go:32`; `internal/cli/tokens.go:441` |
| 16 | Transient `--settings` file prefix `moai-kanban`; no sweeper or reader of the prefix was found | session-private temp file | RENAMED `moai-factory` — `internal/cli/kanban_settings.go:89,108` |
| 17 | Package path strings used at run time by the home-state tools and a seam test: `./internal/kanban` in `migrate_home_state.go:95,187,202,287`, `home_state_coverage.go:55,64`, `queue_path_seam_scan_test.go:41-44`, and a sample in `evidence_writer.go:73` | source-checkout tooling | RENAMED with the package |
| 18 | Distributed files: three always-loaded/path-scoped rules `kanban-dispatch*.md`, skill `moai-kanban-foreman` (catalog entry with a hash regenerated by `gen-catalog-hashes.go`, `Makefile:35`), `.claude/loop.md`; retired skills are archived on `moai update` through `legacySkillIDs` (`update_archive.go:45`); retired rule files need no list — the managed-root clean of rows 21–22 removes and backs them up | user projects | skill: add the old id to the archive list; rules: no new mechanism needed for removal and backup (OD-17, option X chosen) |

Rewrite targets (verdict 14):

| # | Observed fact | Evidence |
|---|---|---|
| 19 | `workflows/factory.md` documents a single-session plan→run→verify→sync chain entered by `--factory`/`-f` through a goal preset named `factory_chain`; that preset name occurs in no code or config (2 lines, all in the document); the state record is placed at `.moai/state/factory/` while the code resolves it under the todo state directory; the three chain fields are documented as orchestrator-written and no production code path writes them. | `.claude/skills/moai/workflows/factory.md:3-29,64-66,112,118-135`; `internal/kanban/state_dir.go:25-37`; `internal/kanban/record.go:57-120` |
| 20 | The same chain contract is referenced by four further files: `moai.md:210`, `run.md:52`, `run/mode-orchestration.md:86` ("Verify Exit Gate (factory contract)"), `sync/quality-gates-quality.md:133` ("a sync entered from a factory chain"). | those files, both trees |

Leader-notice facts (verdict 17): the notice builds `leaderManual` (count twice), `entryGuide` (count), and a free-slot line from `FactoryFreeSlots(root, lanes, …)`, whose only production caller is that line; `MOAI_FACTORY_WORKERS` carries the count and is also the factory discriminator and block-cap signal. — `internal/hook/session_start_factory_i18n.go:35-50,75-99`; `internal/hook/session_start_factory.go:185-250`; `internal/kanban/factory_slots.go:354-360`.

Constitution evidence (verdict 16): none of the four kanban sentences' distinctive phrases occurs in the zone registry; none of the three kanban rules carries a zone tag; the sentence at `agent-common-protocol.md:75` sits in the body of the registered Evolvable Language Handling clause, whose registered text is the header sentence only; `moai constitution validate` is OK on this tree (97 of 101 entries). — progress.md PV-26..PV-28, PV-38.

Update-flow facts (verdict 18; progress.md PV-54..PV-59), measured on tree `a6d3e6fd4`:

| # | Observed fact | Evidence |
|---|---|---|
| 21 | Every `moai update` that runs the template sync includes a stage "Removing old MoAI-managed files" that removes the whole `.claude/rules/moai` root, among other managed roots, before redeploying the template. | `internal/cli/update/deploy/deploy.go:75-78` (managed targets), `:107-185` (the clean); `internal/cli/update_template_sync.go:388-421` (the stage) |
| 22 | Before removing a root, the clean copies every regular file the embedded template does NOT carry at the same relative path into the run's pre-clean backup `.moai-backups/<timestamp>/pre-clean/`, aborts the removal if that backup fails, and reports "backed up N unmanaged file(s)". A file the template DOES carry is not backed up (the deploy rewrites it). Observed against the real embedded template: a user-modified copy of a path the template carries today was not in the backup; a user-modified file at a path it does not carry was backed up byte-identical, and both left the live tree. | `deploy.go:90-107,394-418`; `deploy_preclean_backup_test.go:47-101` (passing, PV-56); scratch probe PV-55 |
| 23 | `.claude/rules/moai/**` is MoAI-managed, not a user-owned namespace: `IsUserOwnedNamespace` has no branch for it, and `IsMoaiManaged` covers `rules`. The existing deprecated-path sweep (`defs.DeprecatedPaths`) cannot carry the retired rules as designed: it deletes after backup regardless of modification, its error aborts the update (`internal/cli/update.go:494-496`), and a registered path that exists counts as a V2 signal (`internal/cli/v2_detection.go:142-144`, overridden only by a confirmed v3 version). | `internal/cli/update/plan/plan.go:152-259`; `internal/cli/update_residue_cleanup.go:50-160`; PV-57 |

### §A.3 Cross-check — can every backend/role pair still be entered after the removals and renames?

| Role | Claude | GLM | Codex |
|------|--------|-----|-------|
| Leader | `moai cc -f` | `moai glm -f` | never enterable (refused before and after) |
| Lane | `moai cc -l` | `moai glm -l` | `moai codex -l` |

No pair enterable today is lost. The renames touch Go identifiers, file names, and package paths, not the entry tokens. The discovery path that reads `MOAI_KANBAN_ID` from a live leader process stays valid: the string value is frozen (REQ-015), the constant that names it is renamed, and AC-016 pins both the value and a discovery read. The Codex lane identifies itself through `MOAI_FACTORY_WORKER` (read by the card verbs at `factory_card.go:62,88-91`) and `MOAI_FACTORY_ROLE`, neither of which the LABEL removal touches (AC-013). What is lost: operator-chosen lane numbering in every spelling, and the kanban three-companion chain, which has no one-to-one factory equivalent.

## §B — Requirements

### §B.1 The lane entry and the leader entry

- **REQ-001** (Event-driven): **When** `moai cc`, `moai glm`, or `moai codex` is launched with `-l` or `--lane` and no argument, the launcher shall start the session as a factory lane on the backend the verb names, joining the running factory and claiming the next free lane slot through the existing lane-slot claim; on `moai codex` the lane is the supervising per-card relaunch lane that `moai codex -f lane` starts today.
- **REQ-002** (Event-driven): **When** `-l` or `--lane` carries any argument (a lane label `lane-<n>`, a number, a leader name, or an `=`-joined value), or is combined with another entry token (`-f`, `-k`) or an operator `--name`, the launcher shall refuse with one line — naming `-l` as taking no argument and `--leader <name>` as the leader selector for an argument, or the one-entry-token rule for a combination — shall launch nothing, and shall write no run, claim, or settings record.
- **REQ-003** (Ubiquitous): A `-l` launch shall publish exactly the lane environment marker set that a `-f lane` launch publishes today — no marker added, none dropped, and no value changed except the lane label — as pinned by the golden of AC-004, with one exception: `MOAI_KANBAN_LABEL`, which REQ-012 removes from the Codex lane child, is outside that set (the golden carries it and the comparison excludes exactly that key).
- **REQ-004** (Ubiquitous): The long `--leader <name>` (space or `=` form) shall keep its meaning as the leader-session selector of a lane join and shall compose with `-l` or `--lane` only; the short `-l` shall no longer be accepted as its spelling; and `--clear-policy`, `--no-auto-dispatch`, and `--factory-run` shall compose with `-l` or `--lane` exactly as they compose with `-f lane` today.
- **REQ-005** (Event-driven): **When** `moai cc` or `moai glm` is launched with `-f` or `--factory` followed by `lane` or a lane label `lane-<n>` (space or `=`-joined), or together with an operator `--name` (or `-n`) whose value has the lane shape `lane-<n>`, the launcher shall refuse with one line naming `-l` as the lane entry, shall launch nothing, and shall write no run, claim, or settings record.
- **REQ-006** (Event-driven): **When** `moai codex` is launched with `-f` or `--factory` in any shape, including `lane`, the launcher shall refuse with one line that names `moai codex -l` for a lane and `moai cc -f` / `moai glm -f` for a leader, and that names no removed form.
- **REQ-007** (Event-driven): **When** `moai cc` or `moai glm` is launched with `-f` followed by a number, the launcher shall keep refusing exactly as it does today (exit 1, one line, nothing launched or written), and the line shall name the bare leader form `-f` and the lane entry `-l` and shall name no removed form.
- **REQ-008** (Ubiquitous): `moai cc -f` and `moai glm -f` (also `--factory`) with no argument shall start the factory leader exactly as today — the same leader markers and the derived capacity — `moai codex` shall continue to have no leader entry, bare `moai` shall print the banner and help and exit 0, `moai cg` shall keep returning the retirement error, and `moai gpt` shall remain an unknown command.

### §B.2 The leader notice

- **REQ-009** (Event-driven): **When** a factory leader session starts, the SessionStart notice shall state how to start a lane — to start a lane, enter `moai cc -l` (or `moai glm -l`, `moai codex -l`) in a new terminal — in one place, and shall carry no lane count, no per-lane launch line, no numbered lane label, and no free-slot list, in each of the four locales.

### §B.3 Kanban Mode removal

- **REQ-010** (Event-driven): **When** `-k` or `--kanban` appears before the pass-through marker on `moai cc`, `moai glm`, or `moai codex`, in any shape (bare, with a SPEC identifier, with a number, with `--name <role>` or `--name lane-<n>`, or `=`-joined), the launcher shall refuse with one line stating that kanban mode is retired and naming `-f` (lead a factory run, on `moai cc` or `moai glm`) and `-l` (join as a lane) as the entries to use, shall launch nothing, and shall write no record, by the same parser-level refusal pattern that already refuses retired flag spellings.
- **REQ-011** (Ubiquitous): The source tree shall contain no kanban chain-seeding mode entry, companion mode entry, companion label parser or registry, `-k` flag parser, kanban notice builder, kanban leader socket path, role-declaration carrier (verdict 20), or kanban board state store (board state, board recovery, columns, reconciliation, and the board's own lock entry points), no launcher shall publish `MOAI_KANBAN` or `MOAI_KANBAN_SPEC`, and the todo queue, factory slots, integration lock, slot lease, landing, and settings-drift code in `internal/kanban` shall remain together with the file-lock substrate they acquire through, which shall be re-homed under a neutral name before the board is deleted (design.md §4.1).
- **REQ-012** (Event-driven): **When** the Codex relaunch loop starts a card session, the child environment shall carry the lane role, the lane label under `MOAI_FACTORY_WORKER`, the backend, and the card id, shall carry no `MOAI_KANBAN_LABEL`, and the factory card verbs run in that child shall resolve the lane label and admission from `MOAI_FACTORY_WORKER` and `MOAI_FACTORY_ROLE` alone.
- **REQ-013** (Ubiquitous): The SessionStart hook shall emit no kanban leader, companion, or bootstrap notice; the factory leader, lane, lane-rule, and stale-run notices shall be emitted exactly as before except for the REQ-009 text.
- **REQ-014** (Ubiquitous): With the kanban code removed and the renames applied, the factory leader and lane shall continue to launch, claim a slot, discover a live leader, receive the transient `crossSessionInbound` settings injection, receive the raised Stop-hook block cap, write their session record, and receive their SessionStart notices, exactly as before.
- **REQ-015** (Ubiquitous): The string values of `MOAI_KANBAN_ID`, `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_KANBAN_SETTINGS_INJECTED`, `MOAI_KANBAN_BACKEND`, and `MOAI_KANBAN_CARD` and their semantics shall remain byte-identical, every reader — including live-leader discovery — shall read them under those values, the Go constants that name them shall carry factory names (design.md §4.7), and the legacy project-local state-directory literal that reads data written by older versions shall remain.
- **REQ-016** (Ubiquitous): A project holding pre-existing kanban artifacts — session records with role `plan`/`run`/`sync`, a `.moai/state/kanban-board` directory (its `roles/` declarations included), kanban markers in a surviving session environment — shall not make `moai doctor`, the web console, the statusline, or any hook fail; an unreadable artifact shall degrade to absence.

### §B.4 Renames, rewrites, and mirrors

- **REQ-017** (Ubiquitous): The Go package `internal/kanban` shall be named `internal/factory` and no import path, no run-time package-path string, no file or directory name, and no non-test Go, templ, or script source under `internal` and `cmd` shall carry the word kanban in any letter case except the six marker string values wherever they are written (REQ-015) and four retired-name literals, each confined to its own named file (design.md §4.7): the retired-entry refusals of REQ-010, the retired skill identifier in the update archive list, the legacy state-directory name of REQ-015, and the legacy web route redirect of REQ-018; the types, functions, files, constants, texts, and temporary-file prefix in design.md §4.7 shall carry their new names.
- **REQ-018** (Ubiquitous): The web console shall present no chain session board and shall carry no chain view model or builder, so the Overview screen's chain-stopped attention row, which reads that view model, leaves with it (Q24, open and non-gating: the smallest-footprint reading); the screen formerly served at `/kanban` shall be served at `/factory` with its factory lanes and SPEC pipeline panels; a GET of `/kanban` shall redirect to `/factory`; the live-update area key, the `data-live` markers, the i18n keys, the view-model and handler names shall carry factory names; and the todo card and the Todo screen shall keep their live updates.
- **REQ-019** (Ubiquitous): The three `kanban-dispatch*.md` rule files shall be renamed `factory-dispatch.md`, `factory-dispatch-detail.md`, and `factory-dispatch-mechanics.md` in both the local `.claude` tree and the template mirror with their kanban-only sections stripped and the factory leader/lane doctrine retained; the foreman skill shall be renamed `moai-factory-foreman` with its catalog entry, hash, and tests, and the retired skill id shall be listed for archive on `moai update`; `.claude/loop.md`, `moai todo --auto`, and the leader notice's queue sentence shall name the factory foreman; and every reference and pinned test — including the tests that read the foreman skill by path (`foreman_queue_watch_test.go`, `foreman_queue_statement_test.go`, and `backlog_json_disclosure_mirror_test.go`) — shall be re-pointed.
- **REQ-025** (Event-driven): **When** `moai update` runs the template sync on a project that still carries the three old rule files — each unmodified, user-modified, or already absent — after the rename, the update shall complete, none of the three paths shall remain under `.claude/rules/moai/workflow/`, a byte-identical copy of each file that was present shall exist in the update's pre-clean backup, files outside the managed roots (a user skill, a file under `.claude/rules/local/`) shall be untouched, and a second update shall change nothing further; a user-modified copy of an old rule file is removed from the live tree, not retained, and is preserved in the backup.
- **REQ-020** (Ubiquitous): `workflows/factory.md` shall be rewritten, and the `moai.md` chain-contract bullet, the `run.md` table row, the `mode-orchestration.md` verify-gate section title, and the `quality-gates-quality.md` applicability sentence shall be reworded, so that each assertion they make about entry tokens, the record location and fields, the backend rules, and the goal preset is one the tree observably holds (design.md §6 lists the assertions).
- **REQ-021** (Event-driven): **When** `moai constitution validate` reports anything other than OK, or an edit is found to require changing a registered clause string, the run phase shall stop and return a blocker to the orchestrator instead of continuing; the kanban sentences in the constitution-slot files shall be edited directly, `moai constitution validate` shall report no drift or violation before and after each edit, and no clause registered in the zone registry shall change.
- **REQ-022** (Ubiquitous): Every README, docs-site page, menu and chrome entry, redirect, image (three tracked files), and instruction file (`AGENTS.md`, `CLAUDE.md`, and the tracked maintainer file `AGENTS.local.md` included) that names kanban mode or a removed entry form shall be updated or removed per design.md §5, and no tracked file or directory outside `internal` and `cmd` shall carry the word in its name (the completed SPEC directories under `.moai/specs` and `.moai/reports` excepted) — each kanban-named docs page renamed with a redirect in all four locales or deleted with a redirect — the four locales shall change in one commit, and each local-versus-template pair shall be edited per file in the same change set, honoring the declared-fork marker.
- **REQ-023** (Ubiquitous): The help and usage text, the leader and lane notices, the stale-run hint, the factory card errors, and the refusal lines shall name `-f` and `-l` only and shall name no removed form (`-k`, `-f lane`, `-f lane-<n>`, `moai codex -f lane`, `-f <N>`, the `-l` leader short).
- **REQ-024** (Ubiquitous): The change shall add no OS-specific code path and shall build for windows/amd64, reusing the existing per-OS launch path unchanged.

## §C — Acceptance Criteria (summary)

Full Given-When-Then scenarios, the RED-now evidence ledger, and the Definition of Done live in `acceptance.md`. Tier L ceiling: 25 requirements and 25 criteria; this SPEC carries 25 requirements and 25 criteria (both at the ceiling).

| AC | Verifies | Class | Mechanical check |
|----|----------|-------|------------------|
| AC-001 | REQ-001 (cc, glm) | release-blocking | new `internal/cli` test; RED-now: RED-1 |
| AC-002 | REQ-001 (codex) | release-blocking | new test; RED-now: RED-2 |
| AC-003 | REQ-002 | release-blocking | new test (15 shapes: eight `-l`, seven `--lane`; times three verbs); RED-now: RED-3, RED-10, RED-12, RED-13 |
| AC-004 | REQ-003 | release-blocking | new test; RED-now: RED-1 |
| AC-005 | REQ-004 | release-blocking | new test (composition rows, retired-short rows, `--leader` selector refusal rows); RED-now: RED-1, RED-3, RED-12, RED-14 |
| AC-006 | REQ-005 | release-blocking | new test; RED-now: RED-4, RED-11 |
| AC-007 | REQ-006 | release-blocking | new test; RED-now: RED-5 |
| AC-008 | REQ-007 | release-blocking (message) / guard (refusal) | new test; RED-now: RED-6 |
| AC-009 | REQ-008 | regression-guard | existing + characterization tests |
| AC-010 | REQ-009 | release-blocking | hook tests; RED-now: RED-7, RED-7b |
| AC-011 | REQ-010 | release-blocking | new test; RED-now: RED-K1, RED-K2 |
| AC-012 | REQ-011 | release-blocking | symbol grep + file absence + shared lock substrate present; RED-now: RED-K3, RED-K4 |
| AC-013 | REQ-012 | release-blocking (with mutant) | new tests (M5a) + grep (M5b); RED-now: RED-8, RED-8b |
| AC-014 | REQ-013 | release-blocking | grep + hook test; RED-now: RED-K5 |
| AC-015 | REQ-014 | regression-guard (with eight observed-red mutants) | factory safety-net tests plus the enterable-pair matrix; baseline PV-4, PV-12, PV-17, PV-22, PV-43 |
| AC-016 | REQ-015 | regression-guard | frozen-value test + discovery test + allowlist and legacy-directory tests; baseline BASE-1..BASE-3, PV-43 |
| AC-017 | REQ-016 | regression-guard (authored RED-first on a mutant) | tolerance test; green today by design |
| AC-018 | REQ-017, REQ-024 | release-blocking (REQ-017) / regression-guard (REQ-024) | greps + build + vet + windows cross-build; closes at M10; RED-now: RED-N1, RED-N2, RED-N3; baseline BASE-4 |
| AC-019 | REQ-018 | release-blocking | grep + web tests; RED-now: RED-W1, RED-W2 |
| AC-020 | REQ-019, REQ-025 | release-blocking | file checks + template tests + foreman path-pinned tests + base-ref guards run not-skipped + update fixture (unmodified, user-modified, absent); RED-now: RED-C1, RED-C2, RED-C3, RED-C4 |
| AC-021 | REQ-020 | release-blocking | doc-versus-behavior test + greps; RED-now: RED-F1, RED-F2, RED-F3 |
| AC-022 | REQ-021 | regression-guard | `moai constitution validate` before and after each edit, registered Frozen clause strings unchanged, stop condition; baseline PV-26, PV-38, PV-58 |
| AC-023 | REQ-022, REQ-023 (docs part) | release-blocking | word grep, removed-form grep, and file-name search + redirect checks + mirror-pair checks; RED-now: RED-D1, RED-D2, RED-D3, RED-D4, RED-D5 |
| AC-024 | REQ-023 (strings part) | release-blocking | grep over Go sources + help tests; RED-now: RED-S1 |
| AC-025 | REQ-022 (docs exit gate) | release-blocking (new page, one-commit property) / regression-guard (build, parity) | hugo build, locale and README parity, new-page existence in four locales, one-commit-per-page property; RED-now: RED-D6; baselines BASE-6, BASE-7 |

## §D — Decisions: all decided, and the author readings

Every verdict is recorded (§A.1, decision-index.md Q1–Q19 and Q23) and every requirement above is written to a verdict. The last one, OD-17 / Q18, is settled as Option X: the operator first chose a dedicated retired-rule-file cleanup step in `moai update`; the update code showed the removal and backup already happen in the managed-root clean (§A.2 rows 21–23) and that three of the five safeguards (delete only on a released-content hash match, retain user-modified files, never abort on cleanup failure) cannot be built as stated without changing that contract; shown those facts, the operator chose to rely on the existing clean with no production change. A user-modified copy of an old rule file is therefore backed up and removed, not retained, and that is accepted. REQ-019 and AC-020 already state the observed contract; AC-018's allowlist of retired-name files is unchanged under X because no Go file is added. Three questions the plan audit marked as operator judgments are not decided here (decision-index.md Q20 the release-note scope, Q21 the supersession wording for the two draft kanban SPECs, Q22 the update backup line): each is written to its smallest-footprint reading, and none gates the run phase. A fourth, decision-index.md Q24 (the Overview screen's chain-stopped attention row, which leaves with the chain view model it reads), was found by the compile proof of v0.8.0 and is treated the same way.

**Author choices within the verdicts (all ACCEPTED by the operator as written, 2026-10-02, in two steps: rows 2–6, 9, and 10 first, then rows 7, 8, 11, and 12; the row 1 ruling is confirmed):**

| # | Choice | Why it is the author's, not the operator's |
|---|--------|---------------------------------------------|
| 1 | Verdict 16 was a lane-orchestrator ruling on evidence | confirmed by the operator ("proceed as is"); marked as originating from a ruling in decision-index.md Q16 |
| 2 | The new package name is `internal/factory` (package `factory`) | verdict 12 asks for a name from the factory vocabulary; no `internal/factory` or `package factory` exists |
| 3 | The leader notice also drops the free-slot line | the line is built from the declared lane count and lists lane numbers; verdict 17 says "no number" |
| 4 | `GET /kanban` redirects to `/factory` | bookmark compatibility for a read-only route; no key or script compat is added |
| 5 | The chain-lineage and `moai chain` sections of the kanban docs page move to a new page `advanced/origin-trail-chain` (four locales) | the `moai chain` command is live and has no other docs page (measured) |
| 6 | The home banner and the five-sessions image — three tracked files, `assets/images/kanban-five-sessions.svg`, `assets/images/kanban-five-sessions.png`, and `docs-site/static/images/profile/kanban-five-sessions.png` — are removed | the visual shows the five-column kanban board |
| 7 | The `MOAI_FACTORY_WORKERS` value stays a number | it is the factory discriminator and block-cap signal; only its display is dropped |
| 8 | The four dependents of the chain contract (`moai.md`, `run.md`, `mode-orchestration.md`, `quality-gates-quality.md`) are reworded and the verify exit gate is retained as specification | they name "the factory contract" the rewrite retires; leaving them would contradict the tree; no launcher flag enters or arms the gate (PV-41) |
| 9 | The retired-name literals are confined to four named files (the `-k` retirement refusal, the update archive list, the legacy state-directory name, the legacy web route) | REQ-010, REQ-015, REQ-018, and REQ-019 each need the word to appear once; confining it keeps the AC-018 search exact |
| 10 | The web live-update key is named `factory` | the key covers the queue and lane state files together; verdict 12 asks for a factory-vocabulary name |
| 11 | The three removed docs pages all redirect to `advanced/factory-mode` | no one-to-one successor exists for the board-terms page or the multi-LLM page |
| 12 | `FactoryFreeSlots` stays in the slots package, caller-less | AGENTS.md §5 forbids deleting seemingly-unused code without approval |

## §E — Constraints

- One entry token per launch: `-l` with `-f` or `-k` is refused.
- The kanban removal and the renames land only after the `-f`/`-l` entry is in place and a factory safety net has been observed green AND observed red on a mutant (M1); the net re-runs after every removal and rename milestone.
- The six marker string values are frozen; the Go package rename and the identifier renames never change a persisted or wire string except where §A.2 says RENAMED.
- Slot assignment is automatic through the existing shared claim; this SPEC does not re-implement slot selection, capacity policy, leader discovery, or the lane-join gate.
- A symbol, constant, or file leaves in the milestone that removes its LAST remaining reader, never earlier; the order is proved by compilation — a cumulative scratch-copy build after every milestone, on darwin and `GOOS=windows`, with `go vet ./...` for the test files (design.md §7.1, progress.md PV-73 to PV-90) — and not by a caller search, which was wrong twice.
- Integration units: M0–M1 (tests only); M2+M3+M4 (the entry unit — the refusals and the strings that teach the replacement forms land together, so no merged state teaches a refused form); then M5a, M5b, M6, M7, M8, M9, M10, and M11 each alone (M6 holds two commits, the lock re-home first and the deletions second; M10 holds two, the update fixture test first and the rename second). Every milestone commit builds and keeps the net green. M7–M9 are Go-wide mechanical changes: before each lands, the run-phase coordinator asks the factory leader to hold merges of other lane branches that touch Go sources, and the lane holding the integration window re-measures the old import path and the build on the merged tree (plan.md §D).
- Removal follows repository precedent: a refusal naming the canonical form, not a warning window; there is no migration command.
- All instruction and doc text follows `language.yaml`: English for instruction files; READMEs and docs-site per the four-locale rule and native-idiom policy; `AGENTS.md`, `CLAUDE.md`, and the dispatch rule are always-loaded slots and this change is expected to shrink them.
- Running the full `internal/kanban`/`internal/factory` or `internal/cli` suites takes a resource slot lease first; verification is scoped to the change.

## §F — Out of Scope

### Out of Scope — verb-less root entry

- No `moai -f` / `moai -l` root form, no pre-Cobra argv rewrite, and no `--backend` token or default-backend setting.

### Out of Scope — lane slot claim, capacity, and join

- Slot selection, bump-to-next-free, run capacity policy (SPEC-CODEX-LANE-SLOTS-001), leader discovery and the lane-join gate (SPEC-FACTORY-LANE-JOIN-SOCKET-001), and the factory record are not changed in behavior.

### Out of Scope — behavior of the todo queue, integration lock, slot lease, landing, and settings-drift code

- These move with the package rename and have their text renamed; their behavior is not changed.

### Out of Scope — Codex leader

- A Codex factory leader stays refused; no Codex leader entry is introduced.

### Out of Scope — a migration command or warning window

- No `moai migrate` kanban verb and no deprecation window for `-k`.

### Out of Scope — leftover on-disk kanban state

- The `.moai/state/kanban-board` directory and its `roles/` declaration files in existing projects are neither migrated nor deleted by this SPEC (verdict 20): the code that wrote and read them is removed, the files are left untouched, and no reader depends on them.

### Out of Scope — historical and generated artifacts

- Existing `CHANGELOG.md` entries, `.moai/release-notes/**`, the generated `.moai/project/codemaps/**` (stale after the rename until `/moai codemaps` runs; the integration release hook issues a debt card for stale codemaps), the frozen `internal/cli/testdata/**` rollout fixtures, and the completed SPECs under `.moai/specs/**` that mention kanban mode or the removed forms are not hand-edited in this change; the sync phase writes the ordinary `[Unreleased]` CHANGELOG entry (the sync workflow's D1 deliverable) and no release-note file (decision-index.md Q20); partial-supersession annotations on the affected completed SPECs (including the SPEC-KANBAN-* family and SPEC-WEB-TODO-QUEUE-001's `data-live` key contract) are recorded at sync, and the two draft kanban SPECs are left as they are (Q21).

## §G — Cross-References

- `research.md` — inventory, marker coupling, persisted-name table, symbol searches.
- `design.md` — entry grammar, removal design, the old→new mapping, the docs mapping, the factory.md rewrite assertions, the sequencing proof.
- `plan.md` §B — the verdicts, the OD-17 evidence and the option weighed, and the confirmed readings; §D — milestones M0–M11.
- `decision-index.md` — the numbering table (verdict / OD / Q), the recorded verdicts (Q1–Q19 and Q23), and the four open non-gating questions (Q20–Q22 and Q24).
- `probe/` — the committed compile-proof runner (`probe.go`), its per-stage data (`patches/`), and the entry-parse probe (`entry_probe.go`, `entry_probe_test.go.txt`); progress.md PV-73 to PV-90 records what each prints.
- Partially superseded by this SPEC, recorded at sync (completed SPECs are not edited in the plan phase): `SPEC-FACTORY-LANE-JOIN-SOCKET-001` REQ-008 (the `-l` short), `SPEC-FACTORY-SELF-DISPATCH-001` REQ-SD-003/-004 (the `moai codex -f lane` spelling and the pinned refusal), `SPEC-CODEX-FACTORY-RETIRE-001`, the completed `SPEC-KANBAN-*` family (kanban mode and the package name), and `SPEC-WEB-TODO-QUEUE-001` (its `data-live="kanban"` key, declared a frontend-visible contract at `internal/web/events.go:33-34`, is renamed here). `SPEC-KANBAN-BOOTSTRAP-001` and `SPEC-KANBAN-WORKTREE-001` are `draft` (read from each spec.md) and describe the mode this SPEC removes; their disposition is Q21, and `SPEC-TODO-SQLITE-001` REQ-TOSQ-018 (the legacy state-directory literal appears only in its fallback reader) is kept: the literal stays in one file, `state_dir.go`, which moves with the package, and no test pins its path (PV-67).
