# SPEC-LAUNCHER-ENTRY-FLAGS-001 — Acceptance Criteria (v0.4.0, Tier L)

All criteria are mechanically verifiable. Tier L ceiling: 25 requirements and 25 criteria; this SPEC carries 24 requirements and 25 criteria (at the ceiling — no criterion may be added without merging or splitting the SPEC). Every requirement `REQ-001`..`REQ-024` is cited by at least one criterion (hand count in `progress.md` PV-47; the lint does not check coverage — PV-14, PV-46).

Primary re-measurement unit: the targeted `internal/cli`, `internal/hook`, `internal/discovery`, `internal/web`, `internal/template`, and `internal/config` tests named below, plus the greps, file checks, and the cross-compile named per criterion. The full `internal/cli` suite is NOT a verification step of this SPEC; a heavy package suite or `go vet ./...` takes a resource slot lease first (`moai slot acquire`), because a full run on a loaded machine measures the machine.

**Classification discipline.** **Release-blocking** criteria carry an executed RED-now observation (§E Evidence Ledger: command, verbatim stdout, exit code, tree SHA `a6d3e6fd4`) and name the milestone that flips them green. **Regression-guard** criteria preserve behavior that is correct today and are never recorded as passes before their green lands. A test named below that does not exist yet is authored RED-first by its milestone. A `-run` selector that sweeps zero tests is not a pass and a SKIP is not a pass — every Go command below is read with `-v`, the swept count and the skip count are established before the verdict, and the two template guards that SKIP on this tree (`TestContractModeConstitutionDriftNotIncreased`, `TestContractModeAlwaysLoadedBudget`) are never cited. A zero-output grep or `find` is read only beside its RED-now cell, which is its positive control: the same command returned a non-empty list on the pre-change tree.

Every behavior test runs through the existing launch seam (the pattern of `factory_join_discovery_test.go`: the launch is intercepted before the process replacement), so no real session starts and no real home state is written.

AC-020's update-fixture clause is the proof of OD-17 (Option X, chosen by the operator): the existing managed-root clean (spec.md §A.2 rows 21–22) is relied on with no production change and no dedicated retired-rule step, so the fixture asserts exactly what that clean does — backup, then removal, with a user-modified copy backed up and removed rather than retained. AC-018's allowlist of retired-name files is unchanged under Option X (no new Go file).

## §D — AC Matrix

### AC-001 — `-l` on cc and glm joins as the next free lane through the shared claim (release-blocking)

**Given** a running factory whose registry holds a live `lane-1` claim and the launch seam standing in for the process replacement, **When** `moai cc -l` and `moai glm -l` (and `--lane`) are launched, **Then** each launch carries lane label `lane-2`, claimed through the same shared lane-slot claim `-f lane` used, with the lane role marker set and the backend recorded as the verb's (`claude`, `glm`).

- Verifies REQ-001. Flipped green by **M2**. RED-now: ledger **RED-1**.
- Command: `go test ./internal/cli -run '^TestLaneEntryJoinsNextFreeSlot$' -v -count=1` — exit 0; the verbose output reports a PASS result for this test and for its cc and glm subtests (swept count 1 parent, 2 subtests).

### AC-002 — `-l` on codex starts the supervising relaunch lane (release-blocking)

**Given** the Codex launcher with its lease and child-start steps intercepted, **When** `moai codex -l` (and `--lane`) is launched, **Then** the supervising per-card relaunch loop starts exactly as `moai codex -f lane` started it, claiming its label automatically through the shared claim, and no Codex leader is started.

- Verifies REQ-001 (codex). Flipped green by **M3**. RED-now: ledger **RED-2**.
- Command: `go test ./internal/cli -run '^TestCodexLaneEntryStartsRelaunchLane$' -v -count=1` — exit 0; one PASS result for exactly this test (swept count 1).

### AC-003 — `-l` refusals launch nothing and write nothing (release-blocking)

**Given** a clean project root, **When** `moai cc`, `moai glm`, and `moai codex` are launched with `-l lane-2`, `-l 3`, `-l=lane-2`, `-l leader-2`, `-l -f`, `-l -k`, and `-l --name lane-2`, **Then** each exits non-zero with exactly one diagnostic line — for an argument, naming `-l` as taking no argument and `--leader <name>` as the leader selector; for a combination, naming the one-entry-token rule or that `-l` already names the role — and no factory run record, lane claim, or settings write exists afterwards.

- Verifies REQ-002. Flipped green by **M2** (cc, glm) and **M3** (codex). RED-now: ledger **RED-3**, **RED-10** (red for the right reason: both fail today with removed-form text — "applies to a factory lane join (-f lane / -f lane-<n>)" and "requires a leader label" — naming neither `-l` as the canonical entry nor the one-entry-token rule).
- Command: `go test ./internal/cli -run '^TestLaneEntryRefusals$' -v -count=1` — exit 0; a PASS result for this test and for each subtest (swept count 1 parent, 7 shapes times 3 verbs = 21 subtests).

### AC-004 — Lane markers equal today's `-f lane` markers, no new marker (release-blocking)

**Given** a golden capture of the lane environment produced by today's `-f lane` launch (captured and committed alone BEFORE the change lands), **When** the lane is launched through `moai cc -l`, `moai glm -l`, and `moai codex -l`, **Then** the `MOAI_FACTORY_*` and `MOAI_KANBAN_*` keys and values equal the golden (label aside), and the key set contains no name outside the constants defined in `internal/config/envkeys.go`.

- Verifies REQ-003. Flipped green by **M2**/**M3**. RED-now: ledger **RED-1**. Ordering note (verification-claim-integrity §2.3): the golden lands in its own commit that precedes the change's commit; a golden committed with the implementation leaves the ordering unverifiable. The golden's codex row excludes `MOAI_KANBAN_LABEL`, which M5a removes (AC-013); the golden is re-pinned there in the same commit.
- Command: `go test ./internal/cli -run '^TestLaneEntryEnvParity$' -v -count=1` — exit 0; one PASS result (swept count 1).

### AC-005 — Lane options compose with `-l`; the `-l` short of `--leader` is gone (release-blocking)

**Given** the launch seam, **When** `moai cc -l` is launched with `--leader leader-2`, `--leader=leader-2`, `--clear-policy each`, `--no-auto-dispatch`, `--factory-run <id>`, `-p work`, `-w feat-x`, and a `-- --print` passthrough, **Then** each composes as it composed with `-f lane` (discovery targets `leader-2`; the policy, dispatch, and run selections reach the lane markers; tokens after `--` are not inspected); **and When** `-l leader-2` (the retired short form) is launched, **Then** it is refused per AC-003.

- Verifies REQ-004. Flipped green by **M2**. RED-now: ledger **RED-1** (composition), **RED-3** (retired short).
- Command: `go test ./internal/cli -run '^TestLaneEntryComposesWithLaneOptions$' -v -count=1` — exit 0; one PASS result (swept count 1); and the existing test re-pinned from `-l leader-2` to `--leader leader-2`: `go test ./internal/cli -run '^TestFactoryLaneJoinLeadTargeting$' -v -count=1` — exit 0, the subtest count established; `grep -c '"-l", "leader' internal/cli/factory_join_discovery_test.go` — count 0 (today the retired short is still used there).

### AC-006 — cc/glm `-f lane*` and the explicit-name lane spelling refused naming `-l` (release-blocking)

**Given** a clean project root, **When** `moai cc` and `moai glm` are launched with `-f lane`, `-f lane-2`, `--factory lane`, `-f=lane`, `--factory=lane-2`, `-f --name lane-2`, `-f -n lane-2`, and `--factory --name=lane-2`, **Then** each exits non-zero with one line naming `-l` as the lane entry and no removed form, and no run record, claim, or settings write exists afterwards; **and When** `-f --name leader-r7` (a non-lane name) is launched, **Then** it launches as a named leader exactly as today.

- Verifies REQ-005. Flipped green by **M2**. RED-now: ledger **RED-4** (the `-f lane` launch is accepted today). The explicit-name rows have no RED-now yet — their current acceptance is derived from code reading (`factory.go:624-642`, `cc.go:203-205`; plan.md §G gap 4); M2's first act is to observe the CURRENT behavior of those three spellings through the launch seam and record it in `progress.md` §E.2 as their RED-now. Until then those rows are not release-blocking-eligible.
- Command: `go test ./internal/cli -run '^TestFactoryLaneSpellingsRefused$' -v -count=1` — exit 0; one PASS result for this test and 16 refusal subtests (eight spellings times cc and glm) plus 2 launch subtests (the non-lane name on cc and glm).

### AC-007 — codex `-f`, in every shape, refused naming `moai codex -l` (release-blocking)

**Given** the Codex launcher, **When** `moai codex` is launched with `-f`, `-f lane`, `-f lane-2`, `-f 3`, and `--factory lane`, **Then** each exits non-zero with one line that names `moai codex -l` for a lane and `moai cc -f` / `moai glm -f` for a leader and names no removed form, and nothing is leased, claimed, or written.

- Verifies REQ-006. Flipped green by **M3**. RED-now: ledger **RED-5** (the refusal line today says "`moai codex -f lane` is the only Codex factory entry").
- Command: `go test ./internal/cli -run '^TestCodexFactoryFlagRefusals$' -v -count=1` — exit 0; one PASS result for this test and five subtests. The byte-pinned refusal tests (`TestCodexFactoryEntryParsingUsesLaneOnly`, `TestCodexFactoryLegacyEntryIsRefused`, `TestSD_AC004_CodexOtherFactoryShapesRefused`) are re-pinned to the new line in the same milestone and run with an anchored selector, swept count 3.

### AC-008 — `-f <N>` stays refused; the message names no removed form (release-blocking for the message, regression-guard for the refusal)

**Given** the current refusal of a numeric `-f`, **When** `moai cc -f 3` and `moai glm -f 3` are launched, **Then** each still exits 1 with one line, nothing launched or written, and the line names the bare leader form `-f` and the lane entry `-l` and no removed form.

- Verifies REQ-007. Flipped green by **M2**. RED-now: ledger **RED-6** (the refusal exists, exit 1; its text names `-f lane` and `-f lane-2`).
- Command: `go test ./internal/cli -run '^TestFactoryCountShapeStillRefused$' -v -count=1` — exit 0; one PASS result for this test and two subtests.

### AC-009 — Leader entry, `cg`, `gpt`, and bare `moai` unchanged (regression-guard)

**Given** the current behavior, **When** `moai cc -f` and `moai glm -f` (and `--factory`) are launched, **Then** the run is recorded with the leader markers and the derived capacity as today and `moai codex -f` has no leader entry; **When** `moai cg` runs, **Then** it prints the retirement error and exits 1; **When** `moai gpt` runs, **Then** it fails as an unknown command; **When** `moai` runs with no arguments, **Then** it prints the banner and help and exits 0.

- Verifies REQ-008. Guard held throughout. No RED-now. Baseline observed on tree `a6d3e6fd4`: `TestCCFactoryEntryRecordsFailOpenRunMetadata` and the two CG tests pass (progress.md PV-4, PV-17); `moai gpt` and `moai cg` exit 1 and bare `moai` exits 0 (PV-1, PV-5). The glm leader and bare `moai` are covered by characterization tests authored FIRST in M1.
- Command: `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestLeaderEntryUnchanged|TestCGRetiredEntryAndModeHaveZeroEffects|TestCGRetirementCompleteEntryShapesAndCounters|TestBareMoaiPrintsBannerAndHelp)$' -v -count=1` — exit 0; five PASS results (swept count 5).

### AC-010 — The leader notice states how to start a lane and no number (release-blocking)

**Given** the factory leader SessionStart notice in each of the four locales rendered for declared lane counts 1, 3, and 8, **When** it is rendered, **Then** it contains the lane-start sentence naming `moai cc -l`, `moai glm -l`, and `moai codex -l` exactly once, its lane guidance is byte-identical across the three counts, it contains no per-lane numbered line, no free-slot list, no count sentence, and no removed form, and its entry guide names `-l` and `-f` only.

- Verifies REQ-009. Flipped green by **M4**. RED-now: ledger **RED-7** (one numbered `-f lane-<i>` line per lane is built at `session_start_factory.go:207`) and **RED-7b** (the free-slot line is built at `:229` and `:240`).
- Command: `go test ./internal/hook -run '^(TestFactoryLeadNoticePrintsLaneCommandOnce|TestFactoryLeadNoticeIsLaneCountIndependent|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide|TestFactoryGuideNamesWorkerJoinInEveryLocale|TestUnbindNoticeRebindLinePresence)$' -v -count=1` — exit 0; five PASS results (the first two are new; the next three are observed passing on tree `a6d3e6fd4`, PV-12, and re-pinned by M4); `TestFactoryLeadNoticeWorkerCountDrivesLineCount` and `TestFactoryLeadNoticeAllSlotsClaimed` (both passing today, PV-39) are replaced by the first two and by deletion, so a `-run` for them sweeps zero tests afterwards and is not cited.

### AC-011 — `-k` is refused in every shape (release-blocking)

**Given** a clean project root, **When** `moai cc`, `moai glm`, and `moai codex` are launched with `-k`, `-k SPEC-X-001`, `-k 3`, `-k --name plan`, `-k --name lane-1`, `--kanban`, and `-k=3`, **Then** each exits non-zero with one line stating that kanban mode is retired and naming `-f` (lead) and `-l` (join as a lane) — on codex naming `moai codex -l` and `moai cc -f` / `moai glm -f` — launches nothing, and writes no record; **and When** `-k` appears after the `--` pass-through marker, **Then** it is forwarded untouched.

- Verifies REQ-010. Flipped green by **M5a**. RED-now: ledger **RED-K1** (the cc and glm `-k` launches are accepted today), **RED-K2** (the codex refusal today names `moai cc -k`).
- Command: `go test ./internal/cli -run '^TestKanbanEntryRefused$' -v -count=1` — exit 0; one PASS result for this test, 21 subtests (seven shapes times three verbs), and one passthrough subtest.

### AC-012 — The kanban launcher, notice, and board code are gone; the queue, slots, and locks remain (release-blocking)

**Given** the final tree, **When** the Go sources are searched for the kanban chain and companion identifiers and for the board API, and the board family's files are listed, **Then** no source carries the identifiers or the board API (tests included, because a test referencing a deleted API does not compile), the ten board-family files do not exist, the todo queue, factory slots, integration lock, and slot lease files do exist, and the module builds.

- Verifies REQ-011. Flipped green by **M5a** (identifiers), **M5b**, and **M6** (board API and files). RED-now: ledger **RED-K3** (15 non-test files), **RED-K4** (24 files carry the board API today — the 10 board files, the board tests, six further kanban tests such as `admission_test.go` and `status_read_test.go`, and a comment in `integration_lock_mutation.go:24`; the run phase classifies each at M6).
- Command: `grep -rlE 'enterKanbanMode|enterKanbanCompanionMode|parseKanbanFlag|rejectKanbanOnCG|EnvMoaiKanban\b|EnvMoaiKanbanSpec|CompanionRoles|SplitCompanionLabel|kanbanLeaderNotice|kanbanCompanionNotice|kanbanBootstrapNotice' internal cmd --exclude='*_test.go'` — empty output, exit 1; `grep -rlE '\b(LoadBoard|WriteBoardState|AcquireBoardLock|RecoverBoard|ParseColumn|TransitionIntoRun|BoardState)\b' internal cmd --include='*.go'` — empty output, exit 1; `find internal -name 'board_store*.go' -o -name 'board_lock*.go' -o -name 'board_recover*.go' -o -name board.go -o -name column.go -o -name reconcile.go` — empty output, exit 0 (today it lists 18 paths: ten production files and eight of their tests — the other board tests are matched by the symbol grep); `find internal -name 'factory_slots.go' -o -name 'backlog_store.go' -o -name 'integration_lock.go' -o -name 'slot_lease.go'` — four lines, exit 0; `go build ./...` — exit 0.

### AC-013 — The Codex lane child carries no `MOAI_KANBAN_LABEL` and still launches and identifies itself (release-blocking)

**Given** the Codex relaunch loop with the child start intercepted, **When** it starts a card session, **Then** the child environment carries the lane role, the lane label under `MOAI_FACTORY_WORKER`, the backend, and the card id and carries no `MOAI_KANBAN_LABEL`; **and Given** exactly that child environment, **When** the factory card verbs run (`stage`, `next`-admission), **Then** they resolve the lane label and admission from `MOAI_FACTORY_WORKER` and `MOAI_FACTORY_ROLE` alone and the card reaches merge-ready.

- Verifies REQ-012. Flipped green by **M5a**. RED-now: ledger **RED-8** (the stamp is written; 4 references to the label constant in the launcher) — the existing `TestSD_AC003_CodexRelaunchPerCard` asserts the label PRESENT (`factory_m5_test.go:145-147`), so it is re-pinned to assert it absent. Mutants rejected: removing the stamp but also dropping `MOAI_FACTORY_WORKER` fails the second test; keeping the stamp under another constant name fails the child-env key-set assertion.
- Command: `go test ./internal/cli -run '^(TestSD_AC003_CodexRelaunchPerCard|TestCodexLaneChildEnvOmitsLabelMarker|TestFactoryCardVerbsResolveLaneFromWorkerMarker)$' -v -count=1` — exit 0; three PASS results (swept count 3); `grep -rl 'MOAI_KANBAN_LABEL' internal cmd --include='*.go' --exclude='*_test.go'` — empty output, exit 1 (today: `internal/config/envkeys.go`).

### AC-014 — SessionStart emits no kanban notice; factory notices unchanged (release-blocking)

**Given** a SessionStart input for a session with no kanban markers and for a factory leader and lane, **When** the hook runs, **Then** no kanban leader, companion, or bootstrap text is emitted and the factory leader, lane, lane-rule, and stale-run notices are emitted as before (the lane-start sentence per AC-010).

- Verifies REQ-013. Flipped green by **M5b**. RED-now: ledger **RED-K5** (the hook still calls the kanban notice builder at `session_start.go:587` and `:599`).
- Command: `grep -c 'kanbanBootstrapNotice' internal/hook/session_start.go` — count 0, exit 1; `find internal/hook -name 'session_start_kanban*'` — empty output, exit 0 (today it lists the two source files and four tests); `go test ./internal/hook -run '^(TestSessionStartEmitsNoKanbanNotice|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide)$' -v -count=1` — exit 0; two PASS results.

### AC-015 — The factory survives every removal and rename, and every backend/role pair that is enterable today still is (regression-guard with a mutant probe)

**Given** the factory safety net authored in M1, **When** it runs after M2 through M9, **Then** the factory leader launches and records its run, a lane launches and claims a slot with its markers, the transient `crossSessionInbound` settings injection is applied, the Stop-hook block-cap raise reaches a factory session through the factory clause, a leader and a lane write their session record, the factory SessionStart notices are emitted, a lane discovers a live leader from its marker, and the enterable-pair matrix holds — Claude and GLM leader (`-f`), Claude, GLM and Codex lane (`-l`) launch, and the Codex leader is refused; **and Given** each of the following scratch mutants, **When** the net runs against it, **Then** the net fails on that mutant: (a) the settings-injection call removed, (b) the factory block-cap clause removed, (c) a factory marker publish removed, (d) the discovery read removed.

- Verifies REQ-014. Guard held from M1 onward. No RED-now: green by design. Baseline observed on tree `a6d3e6fd4`: `TestCCFactoryLaneJoinsDiscoveredLeader` and `TestGLMFactoryLaneJoinsDiscoveredLeader` (PV-4), the three hook notice tests (PV-12), `TestCCFactoryEntryRecordsFailOpenRunMetadata` (PV-17), `TestPrepareKanbanSettingsWritesTransientFile` (PV-22), and the two discovery tests (PV-43) pass; the block-cap, session-record, and enterable-pair nets are authored FIRST in M1. Observed-failure completion: the four mutant reds are recorded in `progress.md` §E.2 before M5a starts — a net whose red has never been seen has proven nothing.
- Command: `go test ./internal/cli -run '^(TestCCFactoryEntryRecordsFailOpenRunMetadata|TestCCFactoryLaneJoinsDiscoveredLeader|TestGLMFactoryLaneJoinsDiscoveredLeader|TestPrepareFactorySettingsWritesTransientFile|TestFactoryNetLeaderLaunch|TestFactoryNetLaneLaunch|TestFactoryNetBlockCap|TestFactoryEntryMatrix)$' -v -count=1` (swept count 8; the lane tests are re-pinned from `-f lane` to `-l` by M2 and the settings test is renamed at M7); `go test ./internal/hook -run '^(TestFactoryNetSessionRecord|TestFactoryNetSessionStartNotices)$' -v -count=1` (swept count 2); `go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1` (swept count 2) — each exit 0.

### AC-016 — The six factory-read marker values and the legacy state-directory name are frozen (regression-guard)

**Given** the OD-10 verdict, **When** the leader and a lane are launched and the constants are read, **Then** the launcher publishes `MOAI_KANBAN_ID`, `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_KANBAN_SETTINGS_INJECTED`, `MOAI_KANBAN_BACKEND`, and `MOAI_KANBAN_CARD` under exactly those names and with today's semantics, the six renamed Go constants equal those literals (written in the test file itself), a lane discovers a live leader from `MOAI_KANBAN_ID`, the Codex MCP environment allowlist still lists the names it lists today, and the legacy project-local state directory is still read from `.moai/state/kanban`.

- Verifies REQ-015. Guard held from M1 (the value test is green on today's constant names and is mechanically re-pointed at M7). Baseline observed on tree `a6d3e6fd4`: the nine constant literals at `envkeys.go:182-273` (ledger **BASE-1**), the allowlist literal at `configtoml.go:21` (**BASE-2**), `legacyStateDirName = "kanban"` at `state_dir.go:23` (**BASE-3**), and the discovery tests pass (PV-43). Mutant rejected: renaming a value in the constant but not in the discovery reader passes a launch-only test — the value test and the discovery test reject it.
- Command: `go test ./internal/config -run '^TestFactoryMarkerValuesFrozen$' -v -count=1` (swept count 1); `go test ./internal/discovery -run '^(TestDiscoverLeaderVerifiesLiveLeader|TestDiscoverLeaderDeclinesUnparseableRunID)$' -v -count=1` (swept count 2); `go test ./internal/codexwiring -run '^TestMCPServerEnvVarsKeepFactoryMarkers$' -v -count=1` (swept count 1); `go test ./internal/factory -run '^TestLegacyStateDirStillRead$' -v -count=1` (swept count 1; before M8 the package path is `./internal/kanban`) — each exit 0.

### AC-017 — Pre-existing kanban artifacts do not break any reader (regression-guard, authored RED-first on a mutant)

**Given** a project holding a session record with role `plan`, a `.moai/state/kanban-board` directory, and a surviving session environment carrying `MOAI_KANBAN` and `MOAI_KANBAN_LABEL`, **When** `moai doctor`, the web console builders, the statusline, and the SessionStart hook run, **Then** none fails, each degrades the unreadable or unknown artifact to absence, and the hook emits no kanban notice.

- Verifies REQ-016. Guard held from M1 (green today: the artifacts are understood by the current readers); the test is authored FIRST in M1 and observed RED on a mutant that makes one reader fail on an unknown role, so its red is seen before M5b removes the readers' knowledge of those artifacts. Behavior of a live old session is unexercised (research.md §R14 gap 8).
- Command: `go test ./internal/cli ./internal/hook ./internal/web ./internal/statusline -run '^TestPreexistingKanbanArtifactsTolerated$' -v -count=1` — exit 0; one PASS result in each package that defines it, the per-package swept count established.

### AC-018 — No source, file name, or import path carries the word except four named retired-name files (release-blocking)

**Given** the final tree, **When** the non-test Go, templ, and script sources under `internal` and `cmd` are searched for the word in any letter case and in its Korean, Japanese, and Chinese spellings — with the six frozen marker names allowed — and file names and import paths are searched, **Then** the search returns exactly the four named retired-name files, no file or directory is named with the word, no file imports the old package path, and the module builds and vets.

- Verifies REQ-017. Flipped green by **M7** (identifiers, files, texts) and **M8** (package path). RED-now: ledger **RED-N1** (189 files), **RED-N2** (22 names), **RED-N3** (179 importers).
- Command: `grep -rlP '(?i)(?<!moai_)kanban|moai_kanban(?!_(id|lead_addr|lead_name|settings_injected|backend|card)\b)|칸반|かんばん|カンバン|看板' internal cmd --include='*.go' --include='*.templ' --include='*.js' --exclude='*_test.go' --exclude-dir=testdata --exclude-dir=node_modules` — exactly four lines, `internal/cli/launcher_retired_entries.go`, `internal/cli/update_archive.go`, `internal/factory/state_dir.go`, `internal/web/legacy_routes.go` (order unspecified), exit 0; `find internal cmd -iname '*kanban*'` — empty output, exit 0; `grep -rl '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go'` — empty output, exit 1; `go build ./...` and `go vet ./...` (the second under a slot lease) — each exit 0. A mutant that leaves the word in a fifth file adds a line and fails the exact-four reading; a mutant that renames the package but leaves one import path fails the build.

### AC-019 — The web console drops the chain session board, serves the factory screen at `/factory`, and keeps the Todo screen live (release-blocking)

**Given** the web console, **When** the screen at `/factory` and the Todo screen render and `GET /kanban` is requested, **Then** the chain session board is absent, the factory lanes panel and the SPEC pipeline panel render, the live-update area key, the `data-live` markers, the i18n keys, and the view-model and handler names carry factory names, `/kanban` redirects to `/factory`, and the todo queue keeps its live updates.

- Verifies REQ-018. Flipped green by **M9**. RED-now: ledger **RED-W1** (the chain board panel exists), **RED-W2** (the `/kanban` route is registered).
- Command: `grep -c 'Chain session board' internal/web/screens.templ` — count 0, exit 1; `go test ./internal/web -run '^(TestFactoryScreenOmitsChainBoard|TestFactoryScreenStillShowsFactoryLanes|TestLegacyKanbanRouteRedirects|TestTodoScreenStillLive|TestWebLiveKeyContract)$' -v -count=1` — exit 0; five PASS results (the last asserts `events.go` and the `app.js` list agree on the key set); `go build ./internal/web` — exit 0, with the regenerated `_templ.go` committed (`make templ-generate` then `git status --short internal/web` empty).

### AC-020 — The rules, skill, loop, catalog, and archive list carry factory names; the doctrine is still loadable (release-blocking)

**Given** the OD-12 and OD-14 verdicts, **When** the local `.claude` tree, the template mirror, the catalog, and the archive list are inspected, **Then** `factory-dispatch.md`, `factory-dispatch-detail.md`, and `factory-dispatch-mechanics.md` exist in both trees with the lane spawn authority, isolation, verification-load, and integration-window doctrine present and the kanban-only sections absent, the old three paths are absent from both trees, the skill is named `moai-factory-foreman` in its directory, frontmatter, and catalog entry with a matching hash, `.claude/loop.md` exists, the old skill id is in `legacySkillIDs`, and the pinned rule-path tests pass; **and Given** a fixture project, built under a temporary directory with the real embedded template, that carries the three old rule files — one byte-identical to a shipped version, one user-modified, one already absent — plus bystander files outside the managed roots (a user skill and a file under `.claude/rules/local/`), **When** the update's managed clean runs and then runs a second time, **Then** each run returns no error, none of the three paths remains under `.claude/rules/moai/workflow/`, a byte-identical copy of each file that was present exists in the run's pre-clean backup, the absent file leaves no backup entry, the progress output reports the backup, the bystander files are byte-identical and present, and the second run changes nothing.

- Verifies REQ-019. Flipped green by **M10**. RED-now: ledger **RED-C1** (the new rule paths do not exist), **RED-C2** (the foreman skill is a catalog entry under the old name, 2 lines), **RED-C3** (the real template still carries the three old paths, so the managed clean treats them as template-managed and does not back up a user-modified copy — observed in progress.md PV-55: a user-modified copy of a carried path was absent from the backup while a not-carried path was present; the fixture test is authored RED-first at M10 against the pre-rename template for exactly that reason, then flips when the template stops shipping the paths). Mutants rejected: a clean that removes without backing up passes "absent" and fails the byte-identical backup; one that backs up only hash-matching files passes the unmodified case and fails the modified case; one that removes `.claude/rules` instead of `.claude/rules/moai` fails the bystander check. This criterion is the observed contract of the managed-root clean (spec.md §A.2 rows 21–22), which the operator chose to rely on (OD-17, Option X): no dedicated retired-rule step is added, a user-modified copy is backed up and removed (not retained), and the fixture mutants above are the guard against a clean that stops backing up.
- Command: `ls .claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` — six lines, exit 0; `ls .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` — "No such file" twice, exit non-zero; `go test ./internal/cli -run '^TestUpdateRemovesRetiredRuleFilesWithBackup$' -v -count=1` — exit 0; one PASS result for this test and three case subtests (unmodified, user-modified, absent) plus one bystander subtest (swept count 1 parent, 4 subtests); `grep -rl 'Lane spawn authority' .claude/rules/moai/workflow internal/template/templates/.claude/rules/moai/workflow` — at least two paths, one per tree, exit 0; `grep -c 'moai-kanban-foreman' internal/template/catalog.yaml` — count 0, exit 1; `grep -c 'moai-factory-foreman' internal/template/catalog.yaml` — count 2, exit 0; `grep -c 'moai-kanban-foreman' internal/cli/update_archive.go` — count 1, exit 0; `go test ./internal/template -run '^(TestWorkflowRulePathsPinned|TestContractModeLocalTemplateParity|TestDeclaredRuleMirrorForks|TestRuleTemplateMirrorDrift|TestAllSkillsInCatalog|TestCatalogHashCoversSkillSubfiles)$' -v -count=1` — exit 0; six PASS results (observed on tree `a6d3e6fd4`, PV-23; the two SKIP guards are not cited); `go test ./internal/cli -run '^TestLegacySkillIDsNotEmbedded$' -v -count=1` — exit 0, swept count 1; `make embed-check` — exit 0.

### AC-021 — `workflows/factory.md` and its dependents assert only what the tree observably holds (release-blocking)

**Given** the rewritten skill and its four dependents, **When** the document-versus-behavior test parses the skill, **Then** every entry form it presents as accepted is accepted by the launcher parse seam and every form it presents as refused is refused, the record path it states equals the resolved `RecordPath` for the project, it names no goal preset, no `.moai/state/factory/` path, no `moai cg` sentinel, and no chain head, and the four dependents no longer name a "factory contract" or a "factory chain".

- Verifies REQ-020. Flipped green by **M10**. RED-now: ledger **RED-F1** (`factory_chain` appears twice), **RED-F2** (`chain head` appears once in `moai.md`), **RED-F3** (`.moai/state/factory` appears in the skill). Mutants rejected: a rewrite that keeps `-f` documented as taking a SPEC argument fails the accepted/refused test; one that states the old record path fails the path comparison.
- Command: `go test ./internal/cli -run '^TestFactorySkillAssertionsMatchBehavior$' -v -count=1` — exit 0; one PASS result (swept count 1); `grep -c 'factory_chain' .claude/skills/moai/workflows/factory.md internal/template/templates/.claude/skills/moai/workflows/factory.md` — both counts 0, exit 1; `grep -c 'chain head' .claude/skills/moai/workflows/moai.md internal/template/templates/.claude/skills/moai/workflows/moai.md` — both counts 0, exit 1; `grep -c '\.moai/state/factory' .claude/skills/moai/workflows/factory.md` — count 0, exit 1; `grep -rn 'factory contract\|factory chain' .claude/skills/moai/workflows/run.md .claude/skills/moai/workflows/run/mode-orchestration.md .claude/skills/moai/workflows/sync/quality-gates-quality.md` — empty output, exit 1.

### AC-022 — The constitution validates before and after each slot edit (regression-guard)

**Given** the baseline `moai constitution validate` on tree `a6d3e6fd4`, **When** it runs after each of the four constitution-slot sentence edits (M10), **Then** it still reports no drift or violation, the zone registry file and the registered clause strings (including the Frozen clauses `CONST-V3R2-036..038` in the section anchored `#user-interaction-boundary`) are unchanged, and any validate result other than OK, or any edit that would need to change a registered string, stopped the run phase and was returned to the orchestrator as a blocker instead of continuing.

- Verifies REQ-021. Guard held. No RED-now. Baseline observed: `constitution validate: OK — no drift or violations detected (97 of 101 entries checked)`, exit 0 (progress.md PV-26, re-run PV-38); the registered Frozen clause strings present in `agent-common-protocol.md`: ledger **BASE-5** (progress.md PV-58). The stop condition is a process obligation recorded in `progress.md` §E.2 (each validate run's output verbatim, and "no blocker" or the blocker returned); a mutant that edits a Frozen clause string passes a grep for the word "kanban" and is rejected by the unchanged-string count below and by `moai constitution validate` itself.
- Command: `moai constitution validate` (built from the tree under test) — exit 0 and the OK line, run before the first and after each of the four edits, the four after-runs recorded in `progress.md` §E.2; `git diff --stat -- .claude/rules/moai/core/zone-registry.md` — empty output; `grep -c -e 'Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.' -e 'preload `AskUserQuestion` via `ToolSearch`' .claude/rules/moai/core/agent-common-protocol.md` — count 2, exit 0 (the same as BASE-5); `grep -c -i 'kanban' CLAUDE.md .claude/rules/moai/core/moai-constitution.md .claude/rules/moai/core/agent-common-protocol.md` — all counts 0, exit 1.

### AC-023 — No README, docs-site, instruction, or configuration file carries the word; the redirects are reversed and valid; the four locales and the mirror pairs agree (release-blocking)

**Given** the OD-5, OD-9, and OD-12 verdicts, **When** the four READMEs, `AGENTS.md`, `CLAUDE.md`, `.claude`, the template mirror, `docs-site` content, data, i18n, layouts, and static assets, `.moai/docs`, and `.moai/config` are searched (the mirrored Claude Code docs excluded), **Then** no file matches; **and** `vercel.json` carries no rule whose source is a factory-mode page, carries a locale rule and a bare rule from each of the three removed pages, and no destination is a removed page; the three removed pages are absent in all four locales; `advanced/factory-mode` and `cli-reference/launchers` present the lane entry in all four locales; and each local-versus-template pair carries the same doctrine.

- Verifies REQ-022, REQ-023 (documentation part). Flipped green by **M11** (closes at sync-audit). RED-now: ledger **RED-D1** (156 files), **RED-D2** (4 shadowing redirects), **RED-D3** (no redirect from the removed pages), **RED-D4** (no factory-mode page presents the lane entry).
- Command: `grep -rIlP '(?i)(?<!moai_)kanban|moai_kanban(?!_(id|lead_addr|lead_name|settings_injected|backend|card)\b)|칸반|かんばん|カンバン|看板' README.md README.ko.md README.ja.md README.zh.md AGENTS.md CLAUDE.md .claude internal/template/templates docs-site/content docs-site/data docs-site/i18n docs-site/layouts docs-site/static .moai/docs .moai/config --exclude-dir=agent-memory --exclude-dir=worktrees --exclude-dir=node_modules --exclude-dir=claude-code` — empty output, exit 1; `grep -cE '"source": "[^"]*(advanced|multi-llm)/factory-mode"' docs-site/vercel.json` — count 0, exit 1; `grep -cE '"source": "[^"]*(advanced/kanban-mode|multi-llm/kanban-mode|core-concepts/kanban-board-terms)"' docs-site/vercel.json` — count 6, exit 0; `grep -cE '"destination": "[^"]*(advanced/kanban-mode|multi-llm/kanban-mode|core-concepts/kanban-board-terms)"' docs-site/vercel.json` — count 0, exit 1; `ls docs-site/content/en/advanced/kanban-mode.md docs-site/content/ko/advanced/kanban-mode.md docs-site/content/ja/advanced/kanban-mode.md docs-site/content/zh/advanced/kanban-mode.md` — "No such file" four times, exit non-zero (the same for the other two pages); `grep -c -e 'cc -l' -e 'glm -l' docs-site/content/en/advanced/factory-mode.md docs-site/content/ko/advanced/factory-mode.md docs-site/content/ja/advanced/factory-mode.md docs-site/content/zh/advanced/factory-mode.md` — every count at least 1, exit 0 (the same on the four `cli-reference/launchers.md` pages); `make embed-check` — exit 0; the single-commit property is read from the commit graph (`git log --name-only` over the change shows the four locale files of a page together).

### AC-024 — Help, usage, notices, and refusal text name `-f` and `-l` only (release-blocking)

**Given** the Go sources and their tests, **When** the sources are searched for the removed forms and the help and notice tests run, **Then** no non-test source in the four packages names `-k`, `-f lane`, `-f lane-<n>`, `moai codex -f lane`, or the `-l, --leader` short (the retired-entry refusals name `-k` only in `launcher_retired_entries.go`), and the help, leader notice, lane notice, stale-run hint, and factory card errors name `-l` and `-f` only.

- Verifies REQ-023 (strings part). Flipped green by **M2**, **M3**, **M4**, **M5a**, and **M11**. RED-now: ledger **RED-S1** (16 files today).
- Command: `grep -rlE -e '-k --name|moai (cc|glm) -k|(-f|--factory)[ =]lane|-l, --leader' internal cmd --include='*.go' --exclude='*_test.go' --exclude-dir=testdata --exclude=launcher_retired_entries.go` — empty output, exit 1 (the retired-entry file is excluded because its `-k` sentence is the one place a removed form is named, and AC-011 asserts that sentence); `go test ./internal/cli -run '^TestLauncherHelpDocumentsLaneEntry$' -v -count=1` (swept count 1) and `go test ./internal/hook -run '^(TestUnbindNoticeRebindLinePresence|TestFactoryGuideNamesWorkerJoinInEveryLocale)$' -v -count=1` (swept count 2) — each exit 0.

### AC-025 — windows/amd64 cross-build (regression-guard)

**Given** the final tree, **When** it is cross-compiled for windows/amd64, **Then** it builds.

- Verifies REQ-024. Guard held at M5a, M5b, M6, M7, M8, M9, and M11. No RED-now. Baseline observed on tree `a6d3e6fd4`: ledger **BASE-4** (exit 0, no output).
- Command: `GOOS=windows GOARCH=amd64 go build ./...` — exit 0.

## §E — Evidence Ledger (RED-now and baselines, pinned to tree `a6d3e6fd4`)

Every entry is one read-only single invocation, its stdout verbatim (or its bounded head and tail with the line count where the output exceeds the 50-line ceiling, per the file-redirect contract), its exit code, and the tree it ran on. Binary entries ran through `go run ./cmd/moai` built from the same tree HEAD. The help renderer title-cases the first letters of the rendered error line (`--Leader`, `-F/--Factory`); the source strings are lowercase, and the tests assert on source strings. The tree `a6d3e6fd4` was re-read as HEAD with only this SPEC directory untracked before the measurements of this revision. Long listings are recorded in `progress.md` and cited by id.

```
RED-1  command : go run ./cmd/moai cc -l
       stdout  : (empty; 0 bytes)
       exit    : 1
       stderr  : "--Leader requires a leader label."  (informational)
       tree    : a6d3e6fd4
       why red : today `-l` is only the short of --leader and needs a value; no `-l` lane entry exists (AC-001, AC-004, AC-005)

RED-2  command : go run ./cmd/moai codex -l
       stdout  : (empty; 0 bytes)
       exit    : 1
       stderr  : "--Leader requires a leader label."  (informational)
       tree    : a6d3e6fd4
       why red : no `-l` lane entry on codex (AC-002)

RED-3  command : go run ./cmd/moai cc -l lane-2
       stdout  : (empty; 0 bytes)
       exit    : 1
       stderr  : "--Leader applies to a factory lane join (-f lane / -f lane-<n>); a factory leader names itself, not a target."  (informational)
       tree    : a6d3e6fd4
       why red : the refusal exists but teaches a removed form and does not state that `-l` takes no argument (AC-003, AC-005)

RED-4  command : go test ./internal/cli -run '^TestCCFactoryLaneJoinsDiscoveredLeader$' -v -count=1
       stdout  : === RUN   TestCCFactoryLaneJoinsDiscoveredLeader
                 --- PASS: TestCCFactoryLaneJoinsDiscoveredLeader (3.61s)
                 PASS
                 ok  	github.com/modu-ai/moai-adk/internal/cli	4.636s
       exit    : 0
       tree    : a6d3e6fd4
       why red : the test launches `runCC(ccCmd, {"-f","lane"})` and it succeeds, so `-f lane` is accepted today; AC-006 asserts it is refused

RED-5  command : go run ./cmd/moai codex -f 3
       stdout  : (empty; 0 bytes)
       exit    : 1
       stderr  : "FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex -f lane is the only Codex factory entry; use 'moai cc -f' or 'moai glm -f' for the factory leader"
       tree    : a6d3e6fd4
       why red : the refusal line names the removed `moai codex -f lane` (AC-007); it also shows `-f <N>` is already refused on codex

RED-6  command : go run ./cmd/moai cc -f 3
       stdout  : (empty; 0 bytes)
       exit    : 1
       stderr  : "-F/--Factory takes no argument (the factory leader), the role token -f lane, which joins this session to a running factory as the next free lane, or a lane label (e.g. -f lane-2) that launches exactly that one lane, got \"3\"."
       tree    : a6d3e6fd4
       why red : the refusal exists (pin) but its text names `-f lane` and `-f lane-2` (AC-008)

RED-7  command : grep -n 'FactoryLaneLabel(i)' internal/hook/session_start_factory.go
       stdout  : 207:		launch = append(launch, "moai "+entry+" -f "+kanban.FactoryLaneLabel(i))
       exit    : 0
       tree    : a6d3e6fd4
       why red : the leader notice builds one numbered `-f lane-<i>` launch line per lane; AC-010 asserts the lane-start sentence appears once and no numbered line exists

RED-7b command : grep -n 'leaderFreeSlots\|FactoryFreeSlots' internal/hook/session_start_factory.go
       stdout  : 229:	slots := kanban.FactoryFreeSlots(root, lanes, kanban.FactoryProcessAlive)
                 240:	context = append(context, fmt.Sprintf(m.leaderFreeSlots, slotLine))
       exit    : 0
       tree    : a6d3e6fd4
       why red : the notice builds a free-slot list from the declared lane count; AC-010 asserts no free-slot list and count-independent guidance

RED-8  command : grep -c 'EnvMoaiKanbanLabel' internal/cli/codex_launcher.go
       stdout  : 4
       exit    : 0
       tree    : a6d3e6fd4
       why red : the Codex lane child is stamped with the label (scrub entry, stamp, child-env line, constant use); AC-013 asserts the label absent from the child

RED-10 command : go run ./cmd/moai cc -f -l
       stdout  : (empty; 0 bytes)
       exit    : 1
       stderr  : "--Leader requires a leader label."  (informational)
       tree    : a6d3e6fd4
       why red : the combination is refused only because `-l` needs a value, not by the one-entry-token rule AC-003 requires

RED-K1 command : go test ./internal/cli -run '^(TestCC_KanbanFlagStrippedBeforeLaunch|TestGLM_KanbanFlagParity)$' -v -count=1
       stdout  : === RUN   TestCC_KanbanFlagStrippedBeforeLaunch
                 --- PASS: TestCC_KanbanFlagStrippedBeforeLaunch (0.01s)
                 === RUN   TestGLM_KanbanFlagParity
                 --- PASS: TestGLM_KanbanFlagParity (0.03s)
                 PASS
                 ok  	github.com/modu-ai/moai-adk/internal/cli	1.251s
       exit    : 0
       tree    : a6d3e6fd4
       why red : the cc and glm `-k` entries launch (the tests pass); AC-011 asserts they are refused

RED-K2 command : go run ./cmd/moai codex -k
       stdout  : (empty; 0 bytes)
       exit    : 1
       stderr  : "KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead"
       tree    : a6d3e6fd4
       why red : the refusal names `moai cc -k`/`moai glm -k` as the replacement; AC-011 requires it to state kanban is retired and name `-f`/`-l`

RED-K3 command : grep -rlE 'enterKanbanMode|enterKanbanCompanionMode|parseKanbanFlag|rejectKanbanOnCG|EnvMoaiKanban\b|EnvMoaiKanbanSpec|CompanionRoles|SplitCompanionLabel|kanbanLeaderNotice|kanbanCompanionNotice|kanbanBootstrapNotice' internal cmd --exclude='*_test.go'
       stdout  : internal/config/envkeys.go
                 internal/cli/launcher_blockcap_infinite.go
                 internal/cli/cc.go
                 internal/cli/glm.go
                 internal/cli/factory.go
                 internal/cli/kanban_settings.go
                 internal/cli/codex_launcher.go
                 internal/cli/kanban.go
                 internal/cli/ptycaptest/harness.go
                 internal/kanban/bootstrap.go
                 internal/kanban/role.go
                 internal/hook/session_start_factory.go
                 internal/hook/session_start.go
                 internal/hook/session_start_kanban.go
                 internal/hook/session_start_record.go
       exit    : 0
       tree    : a6d3e6fd4
       why red : 15 non-test files carry the kanban chain and companion identifiers; AC-012 asserts none

RED-K4 command : grep -rlE '\b(LoadBoard|WriteBoardState|AcquireBoardLock|RecoverBoard|ParseColumn|TransitionIntoRun|BoardState)\b' internal cmd --include='*.go'
       stdout  : 24 paths, all under internal/kanban: status_read_test.go, board_recover_test.go, admission_test.go, board_store.go,
                 fix3_wedge_test.go, board_lock_clear_unix.go, board.go, board_recover.go, board_test.go, integration_lock_mutation.go,
                 fix2_probe_test.go, board_store_test.go, board_lock_join_test.go, board_lock_errno_test.go, board_lock_clear_windows_test.go,
                 board_lock_cross_test.go, column_test.go, f1_traversal_test.go, board_coverage_test.go, reconcile_test.go,
                 kanban_helper_test.go, board_lock_test.go, board_lock.go, column.go
       exit    : 0
       tree    : a6d3e6fd4
       why red : the board API and its tests exist; AC-012 asserts none (integration_lock_mutation.go:24 is a comment to reword)

RED-K5 command : grep -n 'kanbanBootstrapNotice' internal/hook/session_start.go
       stdout  : 587:	if notice := kanbanBootstrapNoticeForSource(input.Source, kanbanRoot, input.SessionID, langEnglish); notice != "" {
                 599:		operatorNotice := kanbanBootstrapNotice(kanbanRoot, input.SessionID, operatorLang(h.cfg))
       exit    : 0
       tree    : a6d3e6fd4
       why red : the hook still builds the kanban notice; AC-014 asserts count 0

RED-N1 command : grep -rlP '(?i)(?<!moai_)kanban|moai_kanban(?!_(id|lead_addr|lead_name|settings_injected|backend|card)\b)|칸반|かんばん|カンバン|看板' internal cmd --include='*.go' --include='*.templ' --include='*.js' --exclude='*_test.go' --exclude-dir=testdata --exclude-dir=node_modules
       stdout  : 189 lines (count read by a separate `| wc -l` invocation, progress.md PV-45); head: internal/factorymsg/store.go, internal/factorymsg/factory_run_retire.go, internal/core/git/checkout.go; tail: internal/hook/session_start_record.go, internal/session/session_pid.go, cmd/t657-merge/main.go
       exit    : 0
       tree    : a6d3e6fd4
       why red : 189 files carry the word; AC-018 asserts exactly four named retired-name files

RED-N2 command : find internal cmd -iname '*kanban*'
       stdout  : 22 lines (count by `| wc -l`, PV-45): internal/kanban, internal/hook/session_start_kanban{,_i18n,_test,_i18n_test,_todo_test,_surface_test}.go, internal/cli/kanban{,_settings,_autonomy_test,_companion_name_test,_bootstrap_test,_launch_facts_test,_help_test,_lead_name_test,_dispatch_test,_settings_test}.go, internal/kanban/kanban_helper_test.go, and 4 template paths (the foreman skill directory and 3 rules)
       exit    : 0
       tree    : a6d3e6fd4
       why red : 22 names carry the word; AC-018 asserts none

RED-N3 command : grep -rlE '"github.com/modu-ai/moai-adk/internal/kanban"' internal cmd --include='*.go'
       stdout  : 179 lines (count by `| wc -l`, PV-35), first three internal/factorymsg/factory_run_retire.go, internal/factorymsg/store.go, internal/web/todo_section_test.go, last cmd/t657-merge/main.go
       exit    : 0
       tree    : a6d3e6fd4
       why red : 179 files import the old package path; AC-018 asserts none

RED-W1 command : grep -c 'Chain session board' internal/web/screens.templ
       stdout  : 1
       exit    : 0
       tree    : a6d3e6fd4
       why red : the chain session board panel exists; AC-019 asserts count 0

RED-W2 command : grep -c '"/kanban"' internal/web/app.go
       stdout  : 1
       exit    : 0
       tree    : a6d3e6fd4
       why red : the route is registered as a screen; AC-019 asserts `/kanban` is a redirect only (in the legacy-route file) and the screen is at `/factory`

RED-C1 command : ls .claude/rules/moai/workflow/factory-dispatch.md .claude/skills/moai-factory-foreman/SKILL.md internal/factory
       stdout  : (empty; 0 bytes)
       exit    : 1
       stderr  : "ls: ...factory-dispatch.md: No such file or directory" ×3 (one per path)
       tree    : a6d3e6fd4
       why red : none of the new names exists; AC-020 asserts the six rule paths exist (and AC-018 the package path)

RED-C2 command : grep -n -i 'foreman' internal/template/catalog.yaml internal/cli/update_archive.go
       stdout  : internal/template/catalog.yaml:41:            - name: moai-kanban-foreman
                 internal/template/catalog.yaml:43:              path: templates/.claude/skills/moai-kanban-foreman/
       exit    : 0
       tree    : a6d3e6fd4
       why red : the skill is a catalog entry under the old name and not on the archive list; AC-020 asserts the reverse

RED-C3 command : ls internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md
       stdout  : three long-format `ls` lines, in this order: kanban-dispatch-detail.md (39965 bytes), kanban-dispatch-mechanics.md (18909 bytes), kanban-dispatch.md (26030 bytes), each under internal/template/templates/.claude/rules/moai/workflow/
       exit    : 0
       tree    : a6d3e6fd4
       why red : the real template carries the three old paths, so the managed clean treats them as template-managed and does not back up a user-modified copy (observed in progress.md PV-55 on a scratch probe, file removed after the run: carried path absent from the backup, not-carried path present, byte-identical); AC-020's fixture asserts the opposite once the template stops shipping them

RED-F1 command : grep -c 'factory_chain' .claude/skills/moai/workflows/factory.md
       stdout  : 2
       exit    : 0
       tree    : a6d3e6fd4
       why red : the skill names a goal preset that exists in no code or configuration (progress.md PV-41); AC-021 asserts count 0

RED-F2 command : grep -c 'chain head' .claude/skills/moai/workflows/moai.md
       stdout  : 1
       exit    : 0
       tree    : a6d3e6fd4
       why red : the dependent sentence (`moai.md:210`) names a plan-phase chain head; AC-021 asserts count 0

RED-F3 command : grep -c '.moai/state/factory' .claude/skills/moai/workflows/factory.md
       stdout  : 1
       exit    : 0
       tree    : a6d3e6fd4
       why red : the skill places the record under a directory the code does not use (`RecordPath` resolves `.moai/state/todo`); AC-021 asserts count 0

RED-D1 command : grep -rIlP '(?i)(?<!moai_)kanban|moai_kanban(?!_(id|lead_addr|lead_name|settings_injected|backend|card)\b)|칸반|かんばん|カンバン|看板' README.md README.ko.md README.ja.md README.zh.md AGENTS.md CLAUDE.md .claude internal/template/templates docs-site/content docs-site/data docs-site/i18n docs-site/layouts docs-site/static .moai/docs .moai/config --exclude-dir=agent-memory --exclude-dir=worktrees --exclude-dir=node_modules --exclude-dir=claude-code
       stdout  : 156 lines (count by `| wc -l`, PV-45); head: README.md, README.zh.md, README.ja.md; tail: .moai/docs/audit-artifact-convention.md, .moai/docs/hook-stdin-fail-closed.md, .moai/docs/jamo-integrity-guide.md
       exit    : 0
       tree    : a6d3e6fd4
       why red : 156 files carry the word; AC-023 asserts none

RED-D2 command : grep -cE '"source": "[^"]*(advanced|multi-llm)/factory-mode"' docs-site/vercel.json
       stdout  : 4
       exit    : 0
       tree    : a6d3e6fd4
       why red : four rules (`vercel.json:183,188,193,198`) redirect the factory-mode pages to the kanban pages and would shadow the canonical page; AC-023 asserts count 0

RED-D3 command : grep -cE '"source": "[^"]*(advanced/kanban-mode|multi-llm/kanban-mode|core-concepts/kanban-board-terms)"' docs-site/vercel.json
       stdout  : 0
       exit    : 1
       tree    : a6d3e6fd4
       why red : no redirect leaves the three pages that M11 removes; AC-023 asserts count 6

RED-D4 command : grep -c -e 'cc -l' -e 'glm -l' docs-site/content/en/advanced/factory-mode.md docs-site/content/ko/advanced/factory-mode.md docs-site/content/ja/advanced/factory-mode.md docs-site/content/zh/advanced/factory-mode.md
       stdout  : docs-site/content/ja/advanced/factory-mode.md:0
                 docs-site/content/zh/advanced/factory-mode.md:0
                 docs-site/content/en/advanced/factory-mode.md:0
                 docs-site/content/ko/advanced/factory-mode.md:0
       exit    : 1
       tree    : a6d3e6fd4
       why red : no factory-mode page presents the lane entry (AC-023)

RED-S1 command : grep -rlE -e '-k --name|moai (cc|glm) -k|(-f|--factory)[ =]lane|-l, --leader' internal cmd --include='*.go' --exclude='*_test.go' --exclude-dir=testdata --exclude=launcher_retired_entries.go
       stdout  : internal/config/envkeys.go
                 internal/config/defaults.go
                 internal/cli/factory_card.go
                 internal/cli/glm.go
                 internal/cli/cc.go
                 internal/cli/factory.go
                 internal/cli/factory_lane_relaunch.go
                 internal/cli/codex_launcher.go
                 internal/cli/codex_factory.go
                 internal/cli/kanban.go
                 internal/kanban/bootstrap.go
                 internal/hook/session_start_factory_i18n.go
                 internal/hook/session_start_factory.go
                 internal/hook/session_stale_run.go
                 internal/hook/session_start_kanban.go
                 internal/hook/session_start_kanban_i18n.go
       exit    : 0
       tree    : a6d3e6fd4
       why red : 16 files name a removed form; AC-024 asserts none outside the retired-entry refusal file

BASE-1 command : grep -n 'Env.* = "MOAI_KANBAN' internal/config/envkeys.go
       stdout  : the nine literals at envkeys.go:182,187,195,205,216,225,238,253,273 (values MOAI_KANBAN, _SPEC, _ID, _LABEL, _SETTINGS_INJECTED, _LEAD_ADDR, _BACKEND, _CARD, _LEAD_NAME; recorded in PV-48)
       exit    : 0
       tree    : a6d3e6fd4
       baseline: AC-016 guard

BASE-2 command : grep -n 'mcpServerEnvVarsValue' internal/codexwiring/configtoml.go
       stdout  : 21:	mcpServerEnvVarsValue = `["MOAI_HOME", "MOAI_KANBAN_ID", "MOAI_SESSION_PID", "MOAI_KANBAN_BACKEND", "MOAI_FACTORY_WORKER", "MOAI_FACTORY_WORKERS", "CLAUDE_PROJECT_DIR", "CLAUDE_CODE_SESSION_ID"]`
       exit    : 0
       tree    : a6d3e6fd4
       baseline: AC-016 guard

BASE-3 command : grep -n 'legacyStateDirName = ' internal/kanban/state_dir.go
       stdout  : 23:const legacyStateDirName = "kanban"
       exit    : 0
       tree    : a6d3e6fd4
       baseline: AC-016 guard

BASE-5 command : grep -c -e 'Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.' -e 'preload `AskUserQuestion` via `ToolSearch`' .claude/rules/moai/core/agent-common-protocol.md
       stdout  : 2
       exit    : 0
       tree    : a6d3e6fd4
       baseline: AC-022 guard (the registered Frozen clause strings CONST-V3R2-036..038 present, unedited)

BASE-4 command : GOOS=windows GOARCH=amd64 go build ./...
       stdout  : (empty; 0 bytes)
       exit    : 0
       tree    : a6d3e6fd4
       baseline: AC-025 guard
```

Green paths: M2 flips RED-1, RED-3, RED-4, RED-6, RED-10; M3 flips RED-2, RED-5; M4 flips RED-7, RED-7b; M5a flips RED-K1, RED-K2, RED-K3 (identifiers), RED-8; M5b flips RED-K5; M6 flips RED-K4; M7 flips RED-N1 (identifiers), RED-N2 (file names), RED-S1 (partly); M8 flips RED-N1 (remainder), RED-N3; M9 flips RED-W1, RED-W2; M10 flips RED-C1, RED-C2, RED-C3, RED-F1..F3; M11 flips RED-D1..D4, RED-S1 (rest). After each flip the command is re-run on the milestone tree and the new output is recorded in `progress.md` §E.2, never in this file.

Mutant probe (verification-completeness §2): a mutant that adds `-l` but leaves `-f lane` accepted passes AC-001..005 — AC-006 and AC-024 reject it. A mutant that adds `-l` but still accepts `-l lane-2` passes AC-001 — AC-003 rejects it. A mutant that keeps the `-l` short for `--leader` passes AC-001 — AC-005's retired-short row rejects it. A mutant that leaves `-f --name lane-<n>` accepted passes AC-001..005 — AC-006 rejects it. A mutant that refuses `-k` on cc but not on glm or codex passes AC-001..010 — AC-011's 21 subtests reject it. A mutant that deletes the kanban files but also the factory settings injection or the factory block-cap clause passes AC-012 — AC-015's net fails on it, which is why the net is probed with those exact mutants. A mutant that renames a marker value in the launcher but not in the discovery reader passes AC-004 — AC-016 rejects it. A mutant that prints the lane count only when it exceeds one passes a notice test at count 1 — AC-010's count-independence at 1, 3, and 8 rejects it. A mutant that removes the Codex label stamp and also stops publishing the worker marker passes the "no label" assertion — AC-013's card-verb test rejects it. A mutant that renames the package directory but leaves the package clause passes a file-name search — AC-018's build rejects it. A mutant that edits only the local rule copies passes AC-020 on the local tree — AC-020 lists both trees and AC-023 searches both. A mutant that appends the three new redirects without removing the four shadowing ones passes AC-023's count-6 row — its count-0 row for the factory-mode sources rejects it. A mutant that rewrites `factory.md` to describe `-f` as taking a SPEC argument passes the grep rows of AC-021 — the accepted/refused test rejects it.

## §F — Edge cases

- `moai cc -k --help` and `moai cc -l --help`: the launcher help, exit 0, nothing launched; the help names no removed form.
- `moai cc -- -k`: `-k` after the pass-through marker is forwarded untouched.
- `moai cc -l` with no running factory: the existing `NO_ACTIVE_FACTORY` refusal (with verified leader discovery first), unchanged text.
- `moai cc -l` outside a git working tree: the existing lane-join precondition refusal, unchanged text.
- `moai codex -l` while no card is leasable: the relaunch loop's existing no-card stop, unchanged.
- `moai cc -f` while a Codex-led run is active: the existing refusal that a Codex-led run cannot be adopted by another leader, unchanged.
- A surviving session started before the upgrade, with `MOAI_KANBAN` still set: it receives no kanban notice and no companion record (AC-017); it is not killed or migrated.
- `moai cc --name lane-3` alone (no `-f`, no `-l`): a plain named session, not a lane join — unchanged by this change.
- A browser tab opened before the upgrade keeps the old web script and old area key until reload; the server no longer emits the old key (residual risk, not a criterion).

## §G — Definition of Done

1. Every release-blocking criterion has its RED-now ledger entry flipped green on the milestone tree, with the new command and verbatim output recorded in `progress.md` §E.2; AC-006's explicit-name rows first record their observed RED-now.
2. Every regression-guard criterion passes at the end of the run phase on the tree it is recorded against; AC-015's four mutant reds and AC-017's mutant red are recorded before M5a starts.
3. AC-020's update fixture (OD-17, Option X) has its RED observed on the pre-rename template before the rename commit, and no production file is added or changed for the old rule files; AC-018's allowlist is unchanged.
4. No `-run` selector verdict is read without its swept count; no zero-test selector and no SKIP is recorded as a pass.
5. The pattern-based inventories in plan.md §B.5 and research.md are replaced by a measured edit list at the start of each milestone.
6. The AC-004 golden is committed before the implementation commit.
7. Each milestone is mergeable alone: its commit leaves the factory net green and the build passing.
8. Sync phase: READMEs and docs-site updated in four locales in one commit; partial-supersession annotations recorded on the affected completed SPECs (including the SPEC-KANBAN-* family); AC-023 closes at sync-audit.
