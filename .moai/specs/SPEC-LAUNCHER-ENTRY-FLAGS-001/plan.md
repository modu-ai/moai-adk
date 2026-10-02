# SPEC-LAUNCHER-ENTRY-FLAGS-001 — Implementation Plan (v0.4.0, Tier L)

All measurements were taken on tree `a6d3e6fd4` (branch `WT-launcher-entry-flags`), against a binary built from that tree, `go run ./cmd/moai`, and targeted `go test` and grep runs recorded in `progress.md` §G. File counts are pattern-based upper bounds (`research.md` states each pattern); the run phase replaces them with a measured edit list at the start of each milestone.

## §A — Context

The factory entry already exists on the backend verbs (`moai cc|glm -f`, `-f lane`, `-f lane-<n>`, and `moai codex -f lane`). The operator's verdicts reshape it into two flags — `-f` (leader, unchanged) and a new `-l`/`--lane` (lane, no argument, automatic slot) — remove the lane spellings `-l` replaces, remove Kanban Mode entirely, and rename every surviving name that carried the word, including the Go package `internal/kanban`. Measurement shows the factory depends on part of what the tree names "kanban": six environment markers (one read from a live leader process, so their values are frozen), a settings-injection helper, shared launch helpers inside the 36 KB `kanban.go`, the leader/lane doctrine inside the `kanban-dispatch*.md` rules, a queue-polling sentence in the factory leader notice, a shared web live-update key, and a package that also holds the todo queue and factory slots. The plan removes the mode's code first, then renames what remains, sequenced so the factory is proven intact before any removal and re-proven after every step.

## §B — Decisions

### §B.1 Decided (decision-index.md Q1–Q17)

Operator answers via AskUserQuestion, lane-3, 2026-10-02 (Q1–Q15, Q17); one lane-orchestrator ruling on evidence (Q16).

| ID | Verdict | Lands in |
|----|---------|----------|
| OD-6 | (SUPERSEDED by OD-9) new tokens only, `-k` untouched | — |
| OD-3 | Backend is carried by the verb; no `--backend`, no default-backend setting | REQ-001 |
| OD-1 | Retire the short `-l` of `--leader`; keep `--leader <name>` | REQ-004 |
| OD-4 | Remove `-f lane`, `-f lane-<n>`, `moai codex -f lane` now (refusal naming the canonical form); `-f <N>` is a pin | REQ-005..007 |
| OD-2 | `-l` takes no argument; argument forms are refused | REQ-002 |
| OD-5 | Documentation scope is everything that mentions an entry command | REQ-022, REQ-023 |
| OD-7 | Refuse the explicit-name lane spelling | REQ-005 |
| OD-8 | Leader notice prints the lane command once (the count sentence is superseded by OD-16) | REQ-009 |
| OD-9 | Kanban Mode removed in this SPEC, whole extent, Tier L | REQ-010..016 |
| OD-10 | Keep the six factory-read marker names and values; no rename, no dual read; Go names change | REQ-015 |
| OD-10-LABEL | Remove the `MOAI_KANBAN_LABEL` stamp on the Codex lane child | REQ-012 |
| OD-11 | Rename everything that carries the word and survives, including the package `internal/kanban`; old→new mapping stated | REQ-017, REQ-018 |
| OD-12 | Rename the three `kanban-dispatch*.md` rules; strip kanban-only sections; docs pages renamed-with-redirect or deleted-with-redirect, four locales in one PR | REQ-019, REQ-022 |
| OD-13 | Rewrite `workflows/factory.md`, the `moai.md:210` sentence, and the Record chain-field docs to today's behavior | REQ-020 |
| OD-14 | Keep the foreman skill, `.claude/loop.md`, `moai todo --auto`, and the notice's queue sentence under factory names | REQ-019 |
| OD-15 | Raised as a lane-orchestrator ruling on evidence; CONFIRMED by the operator ("proceed as is"): edit the four constitution-slot sentences directly with `moai constitution validate` before and after each edit; do not touch registered Frozen clause strings; any validate failure or any need to touch a registered string is a blocker returned to the orchestrator | REQ-021 |
| OD-17 | Operator first chose Option A (a retired-rule-file cleanup step in `moai update` with five safeguards); shown the update-code facts (§B.2), the operator chose **Option X**: rely on the existing managed-root clean, no production change, no dedicated step; the fixture test is the proof | REQ-019, AC-020 |
| Q19 | All author choices in `spec.md` §D accepted as written (seven first, the remaining four afterward) | — |
| OD-16 | The leader notice drops the lane count entirely and states only the lane-start command | REQ-009 |

The `-k` back-compat position, measured against repository precedent: the retired-`cg` guard is a pre-Cobra guard for a retired VERB (`internal/cli/root.go:72-76`); `-k` is a FLAG parsed inside the launcher verbs, so the matching precedent is the parser-level refusal of retired flag spellings (`internal/cli/factory.go:235-251`, `:310-323`). The refusal names `-f` (lead) and `-l` (lane) — not a one-to-one replacement.

OD-15 evidence (re-measured, progress.md PV-26..PV-28, PV-38): zero registry hits for the four sentences; no `[ZONE:Frozen]` tag in the three kanban rules; `moai constitution validate` OK (97 of 101 entries). The nuance recorded, not blocking: `agent-common-protocol.md:27` sits in the section anchored `#user-interaction-boundary`, which holds the three Frozen clauses `CONST-V3R2-036..038` (registered strings: three short clauses the edit does not touch), and `:75` is in the body of `#language-handling` (`CONST-V3R2-039`, registered clause = the header sentence). Nothing contradicts the ruling; if `validate` reports drift after any one edit, the run phase returns a blocker instead of continuing.

### §B.2 Decided with evidence — OD-17 (stale renamed rule files in user projects)

#### OD-17 — Option X chosen by the operator after seeing the code facts

The operator first chose Option A: a retired-rule-file cleanup step in `moai update` removing the old installed `kanban-dispatch.md`, `kanban-dispatch-detail.md`, and `kanban-dispatch-mechanics.md`, with five safeguards: (1) respect the namespace-protection contract; (2) delete only on a released-content hash match, and back up and report a user-modified file instead of deleting it silently; (3) idempotent and fail-open; (4) a fixture RED/GREEN pair for unmodified, user-modified, and already-absent; (5) a named file list kept in one place and mirrored for Template-First. Reading the update code (progress.md PV-54..PV-59) showed the premise does not hold: every `moai update` already removes the whole MoAI-managed `.claude/rules/moai` root and first backs up every file the template does not carry, so once the template stops shipping the three paths they are removed and backed up with no new code; safeguard (1) holds trivially (the paths are not a user-owned namespace); safeguard (2)'s "retain a modified file" cannot coexist with the managed-root removal without changing that contract, and its hash list would be built from 65 committed revisions of the main file; safeguard (3) conflicts with the clean's tested abort-on-backup-failure (`TestCleanMoaiManagedPaths_BackupFailureAbortsRemoval`); and the `legacySkillIDs` analogue would be a new list plus step, since `defs.DeprecatedPaths` aborts the update on error and flips V2 detection. Shown those facts, the operator chose **Option X** (source: operator answer via AskUserQuestion, lane-3, 2026-10-02).

**What Option X means, stated plainly.** No production change and no dedicated retired-rule step. The existing managed-root clean does the work: pre-clean backup, then removal. Under X a user-modified copy of an old rule file is backed up and removed from the live tree — it is NOT retained — and the operator has accepted that. Deletion is not conditional on a hash match, and a failed backup aborts the removal and the update by design (an existing, tested data-protection property that X leaves untouched). The proof is the fixture test `TestUpdateRemovesRetiredRuleFilesWithBackup` (unmodified, user-modified, and absent cases, bystander files, second-run idempotence), exactly as REQ-019 and AC-020 describe. AC-018's allowlist of retired-name files is unchanged under X (no new Go file).

Why the old files matter at all: `kanban-dispatch.md` is always-loaded (26,352 B local, 26,030 B template); a copy that survived would keep injecting the removed `-k` doctrine into every turn of every user session. The other two are path-scoped (39,965 B and 18,909 B).

The three options that were weighed (X chosen; Y and Z not taken):

| Option | What it means | Trade-offs | Measured impact |
|--------|---------------|------------|-----------------|
| X (chosen) | Rely on the existing managed-root clean; add only the fixture proof of AC-020 | Zero production change; the three files leave the live tree and are backed up byte-identical (modified or not); the removal is NOT conditional on a hash match, modified copies are not retained, and a backup failure aborts the removal and the update by design | 0 production files; 1 new fixture test; AC-020 as written below |
| Y (not taken) | A dedicated retired-rule step BEFORE the managed clean: a list of the three paths and their released-content hashes in one Go file (the rules analogue of `legacySkillIDs`), classifying each present file as matching a released version or user-modified, reporting each by name with its backup location, fail-open at the step level | Adds the per-file named report and the hash classification the operator asked for; the managed clean still removes the files afterwards and still aborts on backup failure, so deletion is not conditional and nothing is retained; the hash list covers released revisions only (an unlisted build reads as user-modified, the safe direction) and must be regenerated from git history | 1 new Go file + generator script + hash data for 3 paths over their released revisions (65 committed revisions of the main file, 18 v3 tags) + guard test; the Go file joins the AC-018 allowlist of retired-name files; REQ-019 and AC-020 gain one clause (no ceiling change) |
| Z (not taken) | Y plus retention: carve the retired paths out of the managed-root removal when the file does not match a released version, so a user-modified copy stays in place | Meets "never delete a modified file" literally; changes the managed-root contract for every project, edits `backupThenRemove`/`ManagedCleanTargets` and the destructive-target registry and its drift guard; the always-loaded `kanban-dispatch.md` copy a user modified keeps injecting the removed `-k` doctrine every turn, which is the harm the rename must stop | `deploy.go`, `update_destructive_registry.go`, their guard test and the Y files; REQ-019's "none of the three paths remains" clause is revised for the modified case |

The managed clean's own abort on a failed backup is a tested data-protection property; the operator's choice of X leaves it as it is, so "a cleanup failure never aborts the update" is not adopted for the old rule files.

### §B.3 Items the operator has confirmed or accepted

Nothing is left unconfirmed: the OD-15 ruling was confirmed by the operator ("proceed as is"), and every author choice listed in `spec.md` §D is operator-accepted — rows 2–6, 9, and 10 first, and rows 7, 8, 11, and 12 afterward — as written (package name `internal/factory`; the free-slot line dropped from the notice; `GET /kanban` redirecting to `/factory`; the `origin-trail-chain` page; the home banner and image removed; `MOAI_FACTORY_WORKERS` value kept; the four factory-contract dependents reworded and the verify gate retained as specification; the retired-name literals confined to four files; the web live-update key named `factory`; the three removed docs pages redirecting to `advanced/factory-mode`; `FactoryFreeSlots` left caller-less). The acceptance was given in two steps (seven rows first, the other four afterward), so all eleven choices and the confirmed ruling are accepted.

### §B.4 Update targets

Entry grammar — parses: `factory.go:162-180,231-250,258-288,378-394,624-657`; `codex_factory.go:42-60,86-110,128-159`; `codex_launcher.go:700-728,812-880`. Prints: `cc.go:24,77-100,129-131`, `glm.go:40,85-108,131-132`, `codex_launcher.go:60,637-638,773-780`, `factory.go:88-90,241,244,263,273,284,319`, `factory_card.go:95,98,101`, `session_start_factory.go:207`, `session_start_factory_i18n.go:79-83,130-134,175-179,220-224`, `session_stale_run.go:92,103,114,125`. Comment-only: `factory_lane_relaunch.go:17`, `defaults.go:689,700`, `envkeys.go:357`, `bootstrap.go:370`.

Codex lane label — `codex_launcher.go:316-329` (scrub), `:959-961` (stamp), `:1037` (child env); tests `factory_m5_test.go:142-147`, `codex_factory_retire_test.go:52`, `codex_debug_trace_test.go:120`, `factory_m4_test.go:36`.

Leader notice — `session_start_factory.go:185-250` (builder; per-lane line `:207`, free-slot line `:229`), `session_start_factory_i18n.go:20-100,130-134,175-179,220-224` (28 table references), `session_stale_run.go:92-125`, tests in `session_start_factory_test.go` (`:16,44,87,338`) and `stale_run_gate_test.go:143`.

Removal and rename targets — `design.md` §4.1 and §4.7; `research.md` §R3–§R8.

### §B.5 Documentation and mirror inventory (measured upper bounds; replaced by a measured edit list at M10/M11)

| Surface | Files | Notes |
|---------|-------|-------|
| `.claude`, template mirrors, `AGENTS.md`, `CLAUDE.md`, `.moai/docs`, `.moai/config` carrying the word | part of the 156 (combined pattern, claude-code docs excluded) | of the rule pairs, `cross-session-messaging-detail.md` is a declared fork (edit both by hand, keep `mirror-fork: intentional`); the `kanban-dispatch*.md` pairs are not enrolled in the mirror guard (AC-020 compares them by command); the `.toml` Codex agents are covered by `make embed-check` (`Makefile:77`) |
| README ×4, docs-site content/data/i18n/layouts/static, `vercel.json` | the rest of the 156 | four-locale same-commit rule; native-idiom policy; `vercel.json` reversed (design.md §5) |
| Go non-test with the word | 189 | 12 whole-file deletions among them |
| Tests with the word | 308 | 13 board tests and 12 kanban-only launcher/hook tests are deletions; the rest are re-pins or renames |
| Always-loaded slots (`AGENTS.md`, `CLAUDE.md`, `factory-dispatch.md`, `moai-constitution.md`, `agent-common-protocol.md`) | edits expected to shrink them | the rule-loading-budget record (counts and bytes before and after) is written at M10 |
| Excluded | `CHANGELOG.md`, `.moai/release-notes/**`, generated codemaps, frozen testdata fixtures, completed SPECs, mirrored `claude-code/` docs | partial-supersession annotations at sync |

## §C — Technical approach

See `design.md`: reuse the entry parse and the shared claim for `-l`; refuse removed spellings at the parser; add a factory safety net first; delete kanban-only symbols and files; keep factory-shared helpers; freeze the marker values and rename their Go constants; rename the surviving identifiers, then the package, then the web screen, then the rules and skills, then the docs.

## §D — Milestones (priority labels, no durations; each independently verifiable and mergeable; coordinated by manager-lead in the run phase)

Tier L coordination: 12 milestones (at least 3) and a measured file footprint far above 10, so the run phase is led by `manager-lead` (`CLAUDE.md` §4 item 7). Ordering is by decision-reversibility and safety: grammar first (the most likely to change on review), then proof that the factory survives, then removal, then the mechanical renames, then text and mirrors last. Every milestone is mergeable alone: its commit leaves the factory net green and the build passing, and the milestone that follows does not need the one before it to be unmerged.

Priority High — grammar and safety net:

- **M0 — Kickoff.** `moai constitution validate` baseline recorded. No code. Every verdict is recorded and no decision is outstanding, so no milestone is gated on an operator answer; entry to M1 follows the plan-audit PASS under the autonomous Kickoff form.
- **M1 — Safety net and characterization (tests only).** Author, observe green, then observe RED on mutants: the factory net (leader launch and record; lane launch, claim and markers; settings injection; block-cap factory clause; session records; factory SessionStart notices; live-leader discovery); the enterable-pair matrix; the leader/`cg`/`gpt`/bare-`moai` characterization (AC-009); the lane-marker golden (AC-004), committed alone BEFORE any change; the frozen-marker-value test (AC-016) and the pre-existing-artifact tolerance test (AC-017), both green today. Mergeable alone: no production change.
- **M2 — cc/glm entry grammar.** `-l`, the `-l` refusals, the `-f lane*` and explicit-name refusals, the `-f <N>` message, retirement of the `-l` short, gates re-keyed to `-l`, comments. `-k` still parses. AC-001, 003, 004, 005, 006, 008.
- **M3 — codex entry.** `-l` as the relaunch-lane trigger; every `-f` shape refused with the new line; byte-pinned refusal tests re-pinned; reachability of the interactive Codex lane branch measured (`codex_factory.go:128-229`). AC-002, AC-007.
- **M4 — Leader notice and notice strings.** One lane-start sentence, no count, no per-lane lines, no free-slot line; the stale-run hint; the factory card errors; the foreman sentence reworded in four locales; hook tests re-pinned or replaced. AC-010. After M4 the `-f`/`-l` entry is complete and mergeable with kanban still present.

Priority Medium — kanban removal, only after M1–M4 are merged and the net is green:

- **M5a — Launcher removal.** `-k` refused in every shape (cc, glm, codex) from `launcher_retired_entries.go`; delete the kanban-only symbols of `kanban.go` and the kanban branches of `cc.go`/`glm.go`; the block-cap kanban clause; the Codex lane `MOAI_KANBAN_LABEL` stamp, scrub, and constant, with the Codex-lane RED/GREEN pair; the `MOAI_KANBAN`/`MOAI_KANBAN_SPEC` constants; re-run the net. AC-011, 012 (symbol part), 013, 015.
- **M5b — Hook removal.** Delete the two kanban hook files, the notice block and timing lap in `session_start.go`, and the companion/kanban-leader branches of the record role reader; re-run the net. AC-014, 017.
- **M6 — Board family.** Delete the 10 board files and 13 tests and the companion symbols in `bootstrap.go`/`role.go`; re-run the net. AC-012 (file part).

Priority Medium — renames, after removals so nothing is renamed and then deleted:

- **M7 — Identifiers and files inside packages.** The six marker constants (values frozen), `launcherEntryParse`, `prepareFactorySettings`, `factoryRoleFromEnv`, the launcher helper and settings files, error texts, the transient prefix, the statusline label, CLI/MCP help text; the `kanban_*_test.go` files by class. AC-016 re-run, AC-018 (identifier part), AC-024.
- **M8 — Package rename.** `internal/kanban` → `internal/factory`: import path in 179 files, 1,689 qualified references, run-time path strings, the home-state coverage key, local-variable shadows; `go vet ./...` under a resource slot lease. AC-018 (Go part), AC-025.
- **M9 — Web console.** Remove the chain session board panel and its view model and strings; rename route, screen, area key, `data-live`, i18n keys, icon; add the `/kanban` → `/factory` redirect; regenerate templates (`make templ-generate`). AC-019, AC-017.

Priority Low — mechanical text and mirrors, last (session-loaded files are edited at the end of the task, `cache-aware-execution.md` directive 3):

- **M10 — Rules, skills, agents, loop, catalog, factory.md, constitution slots.** Rename and strip the three rules in both trees; rename the foreman skill, catalog entry and hash, add the old id to `legacySkillIDs`, reword `.claude/loop.md`; rewrite `workflows/factory.md` and its four dependents and the `Record` comments (design.md §6); edit the four constitution sentences with `moai constitution validate` before and after each; re-point the 52 rule-path references and 4 pinned tests; author the update fixture of AC-020 (unmodified, user-modified, already-absent; real embedded template; observed RED on the pre-rename template for the stated reason, then GREEN once the template stops shipping the three paths) and add NO retired-rule step and NO production change for the old rule files (OD-17, option X); record always-loaded counts and bytes before and after; stop and return a blocker on any `moai constitution validate` failure or any need to touch a registered clause string; `make embed-check`. AC-020, 021, 022.
- **M11 — Help residue, READMEs, docs-site, final cross-build (sync-phase owner).** README ×4 and docs-site (content, menu, i18n, layouts, `vercel.json` reversed, banner, image) in four locales in one commit; ko/ja/zh pass the humanize step; `.moai/docs` and `gitflow-lane-protocol.md`; AC-023, 024 (final), 025. The criteria close at sync-audit.

Verification gate at each milestone: its targeted tests and greps; the factory net from M1 (M5a onward, re-run after every step); `GOOS=windows GOARCH=amd64 go build ./...` at M5a, M5b, M6, M7, M8, M9, M11; `moai constitution validate` at M10; resource slot lease before `go vet ./...` and any `internal/cli` package-wide run.

## §E — Risks

See `design.md` §9. Deleting user files in user projects is the risk behind the operator's original five safeguards; how it is bounded under option X, which the operator chose: the removal happens only inside the MoAI-managed `.claude/rules/moai` root, never in a user-owned namespace (`IsUserOwnedNamespace` is false for the three paths; the fixture asserts a user skill and a `.claude/rules/local/` file survive); every file the template does not carry is copied byte-identical to the pre-clean backup before the root is removed, and a failed backup aborts the removal (observed, PV-55, PV-56); the update is idempotent (a second run finds the paths absent). What these do NOT bound, and the operator has accepted: a user-modified copy of an old rule file is backed up and removed from the live tree (recoverable from the backup) rather than retained; the report names a count, not paths; deletion is not conditional on a hash match. Plan-level additions: a long single-writer run across 12 milestones (the run phase works in its own worktree and each milestone is committed alone); the package rename is the largest mechanical diff (M8 contains nothing else); the docs milestone is the largest four-locale surface (the one-commit rule, the humanize step, and per-locale parity greps).

## §F — Cross-platform and environment markers

- No OS-specific code is added; the POSIX replace-process and Windows spawn-and-wait launch paths are reused unchanged (`launch_exec_posix.go`, `launch_exec_windows.go:1-40`). `--spawn` still requires tmux as today (`cc.go:49-50`).
- Factory markers `MOAI_FACTORY_*` are untouched. The six factory-read `MOAI_KANBAN*` marker values are frozen; the lane marker set (AC-004) is pinned against a golden captured from today's code before any change.

## §G — Gaps (not observed; none is asserted as fact in spec.md)

`research.md` §R14 lists the measurement gaps (17 items). Plan-level gaps in addition:

1. Which of the eight `kanban_*_test.go` files in `internal/cli` and the four `session_start_kanban_*_test.go` files test kanban-only symbols (deleted) and which test factory-shared helpers (renamed) was not read per file; M5a/M5b classify them.
2. CORRECTED in v0.5.0: the earlier statement that no stale-rule cleanup exists rested on a bounded search for the wrong terms and missed the managed-root clean stage (spec.md §A.2 rows 21–22). Still unobserved: whether every update mode reaches that stage (a template-only or binary-only update, and the "Up to date · Skipping sync" early return were not traced), and what exactly the progress line names about backed-up files beyond the count. Under option X the user is told a count, not the paths; whether that is enough was accepted by the operator but not measured with users.
3. The content of the `multi-llm/kanban-mode` and `core-concepts/kanban-board-terms` pages was not re-read; their delete-with-redirect disposition rests on their titles and the board's removal.
4. `-f --name lane-<n>` as a lane entry today is derived from code reading (`factory.go:624-642`, `cc.go:203-205`); AC-006's first act in M2 is to observe the CURRENT behavior and record it as that row's RED-now.
5. `moai codex -f lane` was not run (it leases cards); the interactive Codex lane branch's reachability is unverified.
6. Whether `moai constitution validate` anchors depend on line numbers in the edited files was not tested; AC-022 measures it after each edit.
7. `go vet ./...` across the module and the windows cross-build were not run in the plan phase; their baselines are measured at M1.
8. Non-Latin native-language prose around the word in READMEs and docs-site pages (`칸반`, `かんばん`, `看板`) was counted, not read.
9. Twenty-four files under `internal/kanban` carry the board API (RED-K4): the ten board files, the board tests, six further tests (`admission_test.go`, `status_read_test.go:324`, `fix2_probe_test.go`, `fix3_wedge_test.go`, `f1_traversal_test.go`, `kanban_helper_test.go`) and a comment. Whether those six tests also exercise kept behavior (for example the status reader) and whether `status_read.go` reads the board directory were not read per file; M6 classifies each before deleting or editing it, and AC-012's symbol search drives the result.

## §H — Cross-references

- `spec.md` §A.2 (measured facts), §D (decisions and author choices); `research.md` (inventory); `design.md` (design, the old→new mapping, the factory.md rewrite assertions, the sequencing proof); `acceptance.md`; `decision-index.md`.
