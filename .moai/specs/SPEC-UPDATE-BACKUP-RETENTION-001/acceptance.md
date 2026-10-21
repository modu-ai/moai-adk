# Acceptance — SPEC-UPDATE-BACKUP-RETENTION-001

This file is the verification layer: Given / When / Then, binary-testable. The requirements it verifies are the GEARS `REQ-UBR-*` entries in `spec.md`.

## Conventions

- Class `release-blocking` criteria (AC-UBR-01, AC-UBR-02, AC-UBR-03 and AC-UBR-06) are main-versus-tree comparison tests. Each carries a RED-now cell for main and for the starting tree and a GREEN cell for the tree the repair is measured on, authored as a pair per `.claude/rules/moai/development/verification-completeness.md` §2. The RED-now cell names a ledger entry in `research.md` § Evidence ledger, or carries the four elements inline (AC-UBR-06(d)); each holds a single-invocation command, verbatim stdout, exit code and tree SHA. AC-UBR-04 and AC-UBR-05, though among the card's six completion criteria, are class `standard` regression guards: their starting observation is green by design (operator ruling Q3), so no RED can be reproduced, and a criterion whose RED cannot be reproduced loses release-blocking eligibility (`verification-completeness.md` §2.1).
- Tree pins. **main** = `ec5a4f392` (full SHA `ec5a4f39202f3c8b404301864b90596e6af2e16f`; measured by the reproduction report `/private/tmp/t1687-repro-report.md`, which extracted main with `git archive` and ran the same probe sources; for the pruner path the main code equals the start tree's apart from an added refusal early return, `git diff --stat ec5a4f392 e669da81b -- internal/cli/update/backup/backup.go` = 40 insertions, 0 deletions). **start** = `e669da81b3f865488dd40f840623bed69839f459` (the card tree before the repair; code identical at `3b133bedf` and at the plan-revision HEAD `8fe4aea6c1a99a5302ae12552b3deefe3bba76c2`, `git diff --stat e669da81b HEAD -- . ':!.moai/specs'` prints nothing; ledger E1-E6 re-run on it). **repaired** = the card branch HEAD at run-phase close, whose full SHA the run phase writes into progress.md §E.3 and cites in each GREEN observation (a branch name is never the pin).
- The test commands are run from the tree root exactly as written and carry no placeholder. They need the probe files recreated under `.moai/reports/t1733/probes/` (a gitignored local directory) from `research.md` Appendix A (A.1 and A.2 probe sources, A.3 overlay files, A.4 the fixed-mtime probe P5); the placement steps are in `research.md` § Evidence ledger. The one input a reader supplies is a project's two backup roots, copied with `cp -Rcp` (AC-UBR-06).
- The probes P1-P4b and the real-data probe only log (`t.Logf`); the verdict is read from the logged values named in each cell. The probe P5 of AC-UBR-06(d) also asserts, so its RED is a failing exit code. The run-phase tests add assertions that fail on the start tree (AC-UBR-24).
- Class `standard` criteria carry a green path and, where cheap, a RED-now observation; they are not release-blocking. A `standard` criterion marked regression guard (AC-UBR-04, AC-UBR-05) has no RED-now cell by design: it records a guard value that must stay as observed, and its mutant is the defect it keeps out.
- "Budget" is `state.backup_past_versions_keep` (default 3). "Window", "excess" and "exempt run" are as in REQ-UBR-001 and REQ-UBR-008.

## Release-blocking criteria and the card's regression guards

AC-UBR-01, AC-UBR-02, AC-UBR-03 and AC-UBR-06 are release-blocking. AC-UBR-04 and AC-UBR-05 sit between them by number and are class `standard` regression guards.

### AC-UBR-01 (P1) — a pre-clean user file survives rotation

Class: release-blocking. Covers: REQ-UBR-003, REQ-UBR-006.
Given a project whose real deploy clean step has copied a user skill file into a backup run's `pre-clean/` directory, and five newer backup runs exist,
When the pruner runs with keepCount 5 and budget 3,
Then at least one file with that user skill's bytes remains under the backup roots.
RED-now, main `ec5a4f392`: `user-file-copies-before=1 deleted=1 user-file-copies-after=0` (reproduction report, P1 row). RED-now, start `e669da81b`: ledger E1 (same line). Red for the stated reason: the pruner keeps the five newest runs and nothing else.
Command (start): `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run '^TestT1733ProbeP1$' ./internal/cli/update/backup/`
GREEN, repaired: `user-file-copies-after` is at least 1 (pre-revert reference `8e43f064a^` shows 1).
Mutant probe: a pruner that deletes nothing passes this criterion; AC-UBR-03 (ten copies leave one) and AC-UBR-06(b) reject it.

### AC-UBR-02 (P2) — the current version does not take a budget slot

Class: release-blocking. Covers: REQ-UBR-003, REQ-UBR-004.
Given one file backed up in five runs as v1, v2, v3, v4, v4 (oldest first),
When the pruner runs with keepCount 1 and budget 3,
Then the runs holding v1, v2 and v3 remain, the older run holding v4 is removed, and the newest run (v4) remains.
RED-now, main `ec5a4f392`: `deleted=4 ... v1-entry-survives=false v2-entry-survives=false v3-entry-survives=false` (reproduction report). RED-now, start: ledger E2 (same line). Red because the pruner keeps only the window.
Command (start): `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run '^TestT1733ProbeP2$' ./internal/cli/update/backup/`
GREEN, repaired: `deleted=1`, `v1-entry-survives=true v2-entry-survives=true v3-entry-survives=true` (pre-revert reference shows deleted=0 with all three true; the repair additionally removes the redundant older v4).
Mutant probe: a pruner that keeps every excess run passes the first three clauses; AC-UBR-03 and AC-UBR-06(b) reject it. A pruner that counts the current version against the budget keeps only v2 and v3 here and fails.

### AC-UBR-03 (P3) — redundant copies are not all kept

Class: release-blocking. Covers: REQ-UBR-005.
Given ten runs that each hold v1 of one file and a newer run that holds v2,
When the pruner runs with keepCount 1 and budget 3,
Then exactly one run holding v1 remains (the newest of the ten), nine are removed, and the run holding v2 remains.
RED-now, main `ec5a4f392`: `deleted=10 remaining=[update-2026-10-11T00-00-00Z] v1-holding-entries-remaining=0` (reproduction report). RED-now, start: ledger E3 (same line). Red because the present pruner keeps only the window; the audit's "keeps all ten" symptom belongs to the reverted policy code.
Command (start): `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run '^TestT1733ProbeP3$' ./internal/cli/update/backup/`
GREEN, repaired: `deleted=9`, `v1-holding-entries-remaining=1` and the remaining v1 holder is the newest of the ten (pre-revert reference shows 10 remaining, which this criterion rejects as redundant).
Mutant probe: keep-everything fails (ten remain); remove-everything fails (zero remain); keeping the oldest of the ten instead of the newest fails the "newest" clause.

### AC-UBR-04 (P4a) — content behind a symbolic `.agents/skills` is untouched (regression guard)

Class: standard (regression guard; the starting observation is green by design, operator ruling Q3; not release-blocking). Covers: REQ-UBR-015, REQ-UBR-016.
Given project A whose `.agents/skills` is a symbolic link to project B's `.agents/skills`, B holding a live link `moai-kanban-foreman`, and A holding the retired folder with only `SKILL.md`,
When the skill-residue cleanup runs on A with the successor shipped,
Then nothing in B changes (B's link survives), A's retired folder stays in `.claude/skills`, and the result carries only report rows.
Command: `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_residue.rel.json -run '^TestT1733ProbeP4a$' ./internal/cli/update/skillresidue/`
Guard value (start `e669da81b`, observed in this revision on `8fe4aea6c`, code identical; ledger E4): exit 0, and the log line `P4a err=<nil> actions=[{.claude/skills/moai-kanban-foreman reported-retired-skill} {.agents/skills/moai-kanban-foreman reported-retired-link}] b-link-survives=true a-folder-in-place=true`. The same line is the expected value on the repaired tree (no action other than the two report rows).
Main `ec5a4f392`: no guard value, the `skillresidue` package and `update_state_residue.go` do not exist on main (`git diff --stat ec5a4f392 e669da81b` shows both as insertions); the guard is adopted on the start tree.
Mutant probe: a cleanup that removes the link through `.agents/skills` (the reintroduced F16 path) makes `b-link-survives=false` and fails; the run-phase test asserts it (no such code exists on this tree, so the mutant is not executed in the plan phase). A cleanup that refuses every pass is not rejected by this criterion (a refusal is outside scope, decision-index Q6).

### AC-UBR-05 (P4b) — nothing is written outside the project through `.moai/archive` (regression guard)

Class: standard (regression guard; the starting observation is green by design, operator ruling Q3; not release-blocking). Covers: REQ-UBR-015, REQ-UBR-016.
Given project A whose `.moai/archive` is a symbolic link to a directory outside the project, and a retired folder in A,
When the cleanup runs on A,
Then nothing is written outside the project and the retired folder stays in `.claude/skills`.
Command: `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_residue.rel.json -run '^TestT1733ProbeP4b$' ./internal/cli/update/skillresidue/`
Guard value (start `e669da81b`, observed in this revision on `8fe4aea6c`, code identical; ledger E5): exit 0, and the log line `P4b err=<nil> actions=[{.claude/skills/moai-kanban-foreman reported-retired-skill}] content-now-outside-project=false a-folder-in-place=true`. The same line is the expected value on the repaired tree.
Main `ec5a4f392`: no guard value (package absent, as AC-UBR-04).
Mutant probe: a cleanup that reintroduces the archive move makes `content-now-outside-project=true` and fails; the run-phase test asserts it (no such code exists on this tree, so the mutant is not executed in the plan phase).

### AC-UBR-06 — real-data remeasure on the mo.ai.kr backup roots

Class: release-blocking. Covers: REQ-UBR-001, REQ-UBR-003 to REQ-UBR-006, REQ-UBR-010, REQ-UBR-011.
Given a copy of the mo.ai.kr project's `.moai-backups` and `.moai/backups` (made with `cp -Rcp` from the live project or its snapshot; no `moai update` run; the copy command recorded) in which the harness has stamped every run directory with a distinct, explicitly fixed modification time after the copy and before the prune (`os.Chtimes`; the stamp is the timestamp the directory name encodes, so a copy that does not preserve times cannot change the ordering, and no two runs tie, so the path-ascending tie-break is never reached),
When the harness runs the pruner on a fresh copy twice: with the production keepCount 5, and with keepCount 4,
Then
(a) the counted set is the backup runs and only those: a run is what `collectBackupRuns` defines (the directories that share one run marker, or an unmarked directory whose name has a writer shape: `YYYYMMDD_HHMMSS` under `.moai-backups`, `update-` or `v2-to-v3-` plus a `2006-01-02T15-04-05Z` stamp under `.moai/backups`; REQ-UBR-001, REQ-UBR-008). Unmanaged entries (directories of any other name, and loose files) are report-only: they are named in the report as a count and total size (REQ-UBR-012), take no part in the keepCount arithmetic, and are not removed. The number of runs present after the prune equals keepCount plus R, where R is the retained-excess run count the report states, and the report names each of those R runs with every directory it spans. Observed on the 10-14 snapshot (read-only `find` over both roots, and the P5 log): config root 20 directories = 5 timestamp-shaped runs + 15 unmanaged directories (13 `sync-*`, `backup`, `settings-backup`), plus 1 loose file; namespace root 79 directories = 47 `update-*` + 32 `v2-to-v3-*`, every one a run, plus 8 loose files; no `.moai-run` marker, so each run is one directory; runs before the prune = 84. After the prune the 15 unmanaged directories, the 1 loose file and the 8 loose files are all still present, and the run count is keepCount plus R with R measured, not a target;
(b) over the pre-prune census, computed independently by a brute-force oracle: every protected version has a retained holder; removing any retained non-exempt excess run would uncover a protected version (irredundancy); and the number of retained non-exempt excess runs is at most the number of distinct protected versions. The criterion does not require equality with any particular cover (irredundant covers of one key set differ in size; design D2);
(c) no path loses a version that is among its N most recent non-window versions;
(d) under the fixed-mtime stamps above, at keepCount 4, at least one file under the config root's `pre-clean/` copy remains. On the snapshot the run `20261008_203921` holds 48 regular files under `pre-clean/`, each a (path relative to the run, SHA-256) key that no other run in either root holds (read-only scratch computation, bookkeeping files excluded), and by the stamps the run is the oldest of the five config runs and ranks 9th of 84, so it is an excess run at keepCount 4 and at keepCount 5;
(e) the report states the number of runs removed and the number retained, and every retained excess run is named with its reason.
RED-now for (b) and (c), main `ec5a4f392` (reproduction report) and start `e669da81b` (ledger E6), keepCount 5: `REAL cfg-root dirs 20->20 ns-root dirs 79->0 deleted=79 pre-clean-dirs 1->1` and `REAL distinct-file-versions-before=1951 lost=1700 paths-with-lost-versions=996`. Red because nothing outside the window is protected: 1700 versions (996 paths) are lost, including versions inside their budget. Pre-revert reference `8e43f064a^`: `ns-root dirs 79->79 deleted=0`, `lost=0`. The ledger line `pre-clean-dirs 1->1` is not evidence that the pruner keeps the pre-clean run: the snapshot's directories all carry one copy time, so the path-ascending tie-break ranked `20261008_203921` first and kept it; (d) below is the cell that measures it with ordering fixed.
Command, (a)-(c) and (e) (start): `env T1733_CLONE=/private/tmp/t1733-real/moaikr go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run '^TestT1733ProbeRealData$' ./internal/cli/update/backup/` after `cp -Rcp <snapshot> /private/tmp/t1733-real/moaikr` (placement: `research.md` § Evidence ledger).
RED-now for (d), observed once in this revision (fresh copy made with `cp -Rcp` from the snapshot; tree `8fe4aea6c1a99a5302ae12552b3deefe3bba76c2`, code identical to start `e669da81b`; main `ec5a4f392` by code identity, the pruner's ordering and removal code being unchanged between them):
```
cp -Rcp <snapshot> /private/tmp/t1733-real-p5-k4b/moaikr
env T1733_CLONE=/private/tmp/t1733-real-p5-k4b/moaikr go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup_p5.rel.json -run '^TestT1733ProbeRealDataFixedMtime$' ./internal/cli/update/backup/
```
Stdout (the second command):
```
=== RUN   TestT1733ProbeRealDataFixedMtime
    zz_t1733_probe_p5_test.go:108: P5 keepCount=4 runs=84 distinct-run-mtimes=true pre-clean-run=20261008_203921 pre-clean-run-rank=9 pre-clean-run-is-oldest-config-run=true
    zz_t1733_probe_p5_test.go:110: P5 pre-clean-files before=48 after=0 deleted=80 pre-clean-run-exists-after=false
    zz_t1733_probe_p5_test.go:112: AC-UBR-06(d): pre-clean files before=48 after=0, want before>0 and after>=1
--- FAIL: TestT1733ProbeRealDataFixedMtime (11.76s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	12.262s
FAIL
```
Exit code: 1. Red because the pruner keeps only the newest keepCount runs and the pre-clean run ranks 9th of 84, so its directory and all 48 files under it are removed. The probe source and overlay are `research.md` Appendix A.3 and A.4. The same fixture at the production keepCount (`T1733_KEEP=5`, another fresh copy) printed `pre-clean-run-rank=9`, `after=0 deleted=79` and exit 1.
GREEN, repaired: the same command prints `after>=1` for `pre-clean-files`, `--- PASS` and exit 0. REQ-UBR-003 gives a stronger expectation, `after=48` (each of the 48 paths has a single version, outside the window, so each is protected); that figure is derived from the snapshot, not observed, and the run phase records the measured value in progress.md §E.2. For (a)-(c) and (e): they hold; `lost` counts only versions older than their path's budget (zero versions inside the budget); R and the number of namespace-root directories remaining are recorded in progress.md §E.2 as measured figures, not as targets.
Mutant probe: keep-everything passes (c) and (d) trivially and fails (b) irredundancy (removing a retained run uncovers nothing); remove-everything passes (a) and fails (b) coverage, (c) and (d). The oracle in (b) is a brute-force check of coverage, irredundancy and the bound, not a re-run of the production selection.
Note: R feeds no open decision (Q2 is ruled, Q4 closed); it is recorded for the operator's post-landing comparison (design D11).

## Standard criteria

### AC-UBR-07 — the budget bounds protection

Class: standard. Covers: REQ-UBR-003, REQ-UBR-004, REQ-UBR-007.
Given one path with five distinct versions each held only by excess runs, and the budget set to 3 and then to 1,
When the pruner runs,
Then exactly the 3 (then 1) most recent versions remain held by a retained run and the older versions are not.
RED-now: main and start keep none of the five (consequence of E2/E3: only the window survives).

### AC-UBR-08 — bookkeeping is recognized by what the writer wrote

Class: standard. Covers: REQ-UBR-009.
Given a config backup that holds a user file named `.complete` with content `user A` in an excess run, whose newer copy in the window reads `user B`, and a second excess run that differs from a retained run only by the writer's `backup_metadata.json`,
When the pruner runs,
Then the run holding `user A` is retained as unique content, and the run differing only by writer-produced bookkeeping is removed.
Residual (audit D8, deferred): a user file that imitates the writer's shape is read as bookkeeping. Closing it needs provenance (the backup kind and location of each file), a design choice larger than this card; it is recorded as an accepted residual in design D4 and spec.md risk 7, and not asserted here.

### AC-UBR-09 — the template render is exempt

Class: standard. Covers: REQ-UBR-002, REQ-UBR-006.
Given an excess run whose only content not held by the window is under `.template-defaults/`,
When the pruner runs,
Then that run is removed.

### AC-UBR-10 — the budget key ships once

Class: standard. Covers: REQ-UBR-007.
Given the tree after M1,
When the embedded `state.yaml`, the configuration struct and the defaults are read,
Then the key `state.backup_past_versions_keep` is present in the shipped template with value 3, the default is defined in exactly one place in non-test Go code, and a configured value of 0 or a negative number makes the pruner use 3.
RED-now (start `e669da81b`): ledger E7 and E8 — `grep -c` of the key in the template returns 0 (exit 1) and of the default constant in `internal/config/defaults.go` returns 0 (exit 1).
Green path: M1.

### AC-UBR-11 — unclassifiable content fails closed and never blocks

Class: standard. Covers: REQ-UBR-008, REQ-UBR-005, REQ-UBR-010.
Given an excess run that holds an unreadable file, an excess run that holds only a FIFO named like a file (no protected key), and an excess run holding a redundant readable copy,
When the pruner runs,
Then the unreadable run and the FIFO-only run are both retained as exempt runs and reported as retained for that reason (REQ-UBR-005's irredundancy clause does not remove them), the redundant readable run is removed, and the prune returns without opening the FIFO (bounded by a test timeout).

### AC-UBR-12 — each decision is reported

Class: standard. Covers: REQ-UBR-011, REQ-UBR-012.
Given a project with removable runs, a run retained for protected content, and non-timestamp entries under `.moai-backups`,
When the update housekeeping runs,
Then the output names every removed run, every retained excess run with its reason, and the unmanaged entries as a count and a total size, and no unmanaged entry is deleted.

### AC-UBR-13 — root containment still refuses

Class: standard. Covers: REQ-UBR-013.
Given a symbolic `.moai-backups` (then a symbolic `.moai/backups`, then a root under a symbolic parent),
When the pruner is asked to run,
Then nothing is removed from either root and the refusal is reported. The existing containment tests continue to pass.

### AC-UBR-14 — explicit wipe and negative keepCount

Class: standard. Covers: REQ-UBR-014.
Given runs holding protected versions,
When the pruner runs with keepCount 0, then with keepCount -1,
Then keepCount 0 removes every run and keepCount -1 removes none.

### AC-UBR-15 — retired folders and links are reported, never moved or removed

Class: standard. Covers: REQ-UBR-015, REQ-UBR-016.
Given a project with a retired skill folder holding only `SKILL.md`, its successor shipped, a regular `.moai` directory, and an `.agents/skills/<name>` link to `.claude/skills/<name>`,
When the cleanup runs twice,
Then the retired folder and the link are both still present after each run, no directory is created under `.moai/archive`, and both runs report the retired folder and link without acting on them.

### AC-UBR-16 — a swapped component cannot redirect the rename (F24)

Class: standard. Covers: REQ-UBR-017.
Given a skill folder that a `Shipped` callback replaces with a symbolic link to an outside directory holding `skill.md` just before the rename,
When the rename step runs,
Then the outside file is not renamed.

### AC-UBR-17 — a rename does not overwrite (F25)

Class: standard. Covers: REQ-UBR-017.
Given a user file created at the temporary name just before the first move of the rename,
When the rename step runs,
Then that file's bytes survive, and the step either completes under another name or restores the content under its own name and reports.

### AC-UBR-18 — an empty database file is reported, not removed (F21)

Class: standard (bound to M4's Exit). Covers: REQ-UBR-018.
Given a zero-byte SQLite file under `.moai/state` with no sidecar and a modification time older than the minimum age, opened by a process that has not written to it,
When the state-residue pass runs,
Then the file is still present, the process's next write succeeds, and the pass output contains the "empty DB file present" row naming the path; a zero-byte `backlog.json.lock` of the same age is still removed.
RED-now (start): the file is removed (`internal/cli/update_state_residue.go:49-57`); the run-phase test fails on the start tree for that reason. Not re-measured in the plan phase (research Gap 2).

### AC-UBR-19 — the pruner's site is registered

Class: standard (bound to M2's Exit). Covers: REQ-UBR-019.
Given the tree after M2,
When `go test -count=1 -run '^TestDestructiveTargetRegistry_CoversAllSites$' ./internal/cli/` runs,
Then it passes and every removal call site this SPEC adds or changes in the pruner has a registry row with a protection or an exemption reason.

### AC-UBR-20 — the changed move site is registered

Class: standard (bound to M3's Exit). Covers: REQ-UBR-019.
Given the tree after M3,
When the same registry test runs,
Then it passes, the held-back rename site has a registry row, and the pin test sees the changed move calls (it fails on a deliberately added unregistered alias, observed once).

### AC-UBR-21 — the key and the report rows are pinned by tests

Class: standard. Covers: REQ-UBR-020.
Given the tree after M5,
When the struct-YAML symmetry tests and the new cli tests run,
Then the symmetry tests include the state key, and the full-reinstall and template-sync paths each have a test that fails when the retention report rows are removed.

### AC-UBR-22 — quality gates

Class: standard.
Given the touched packages,
When `gofmt -l`, `go vet`, `golangci-lint run` and `go test -cover ./internal/cli/update/backup/... ./internal/cli/update/skillresidue/...` run,
Then the first three are clean and coverage is at least 85% for each package; permission-bit tests are skipped on Windows and FIFO tests are unix-only.

### AC-UBR-23 — Template-First holds

Class: standard.
Given the shipped template change,
When `make build` and the `internal/template` tests run,
Then the embedded files regenerate, the tests pass, and the tracked changes in the working tree are exactly the files this card's commits name: the template `state.yaml`, its embedded counterpart, `internal/config` and the touched Go packages.

### AC-UBR-24 — the probe tests are red before the fix, in the commit graph

Class: standard. Covers: ordering attribution (`verification-claim-integrity.md` §2.3).
Given the card branch after run-phase close,
When the commit graph is read,
Then each probe test of AC-UBR-01 to AC-UBR-03 is added, with assertions, in a commit that precedes the commit that makes it pass, progress.md §E.2 holds the verbatim RED output for each, and the guards of AC-UBR-04 and AC-UBR-05 are added in a commit that precedes the residue change.

## Edge cases

- Only one backup root exists; the other is absent. A run that spans both roots with one root containing the protected version: the whole run is kept or removed together.
- keepCount larger than the number of runs; budget larger than the number of versions.
- A run that holds only bookkeeping files and `.template-defaults/`: removable.
- Two copies of one logical file at different paths inside backups (`pre-clean/` and a namespace copy): distinct versions under the Q5 default.
- A file whose content is large: hashed by streaming; no whole-file buffering requirement.
- A removal that fails midway through a run: the run is reported as retained with the failure reason and counted among the retained excess runs.
- A backup directory whose name matches no writer shape and carries no marker: neither counted nor deleted (unchanged).

## Definition of Done

1. AC-UBR-01 to AC-UBR-24 are PASS with verbatim command output, or recorded as a Gap with the reason.
2. `decision-index.md` Q1, Q2 and Q3 carry the operator's verdicts (recorded 2026-10-16); the AC-UBR-06 figures are in progress.md §E.2.
3. The registry guard, build for linux, darwin and windows, gofmt, vet, lint and the coverage floor are clean for the touched packages.
4. progress.md §E.2 holds the RED outputs and the real-data measurement; the card branch is committed, unpushed.
5. No `moai update` was run in any tree, and no live project was written.
6. The post-landing verification of design D11 is handed to the operator; it is not a lane deliverable.
