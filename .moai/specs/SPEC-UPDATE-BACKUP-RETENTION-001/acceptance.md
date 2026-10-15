# Acceptance — SPEC-UPDATE-BACKUP-RETENTION-001

This file is the verification layer: Given / When / Then, binary-testable. The requirements it verifies are the GEARS `REQ-UBR-*` entries in `spec.md`.

## Conventions

- Class `release-blocking` criteria (AC-UBR-01 to AC-UBR-06, the card's completion criteria) carry a RED-now cell and a green-path cell, authored as a pair per `.claude/rules/moai/development/verification-completeness.md` §2. The RED-now cell names a ledger entry in `research.md` § Evidence ledger; each entry holds the four elements (the single-invocation command, its verbatim stdout, its exit code, the tree SHA). Document-level tree pin: `e669da81b3f865488dd40f840623bed69839f459`.
- Class `standard` criteria carry a green path and, where cheap, a RED-now observation; they are not release-blocking.
- "Budget" is `state.backup_past_versions_keep` (default 3). "Window" and "excess" are as in REQ-UBR-001.
- Re-executing a RED-now command requires materializing the probe sources of `research.md` Appendix A into any scratch directory and pointing a `go test -overlay` JSON at them; the commands are recorded with the paths used in this run.

## Release-blocking criteria

### AC-UBR-01 (P1) — a pre-clean user file survives rotation

Class: release-blocking. Covers: REQ-UBR-003, REQ-UBR-006.
Given a project whose real deploy clean step has copied a user skill file into a backup run's `pre-clean/` directory, and five newer backup runs exist,
When the pruner runs with keepCount 5 and budget 3,
Then at least one file with that user skill's bytes remains under the backup roots.
RED-now: ledger E1 — copies before 1, deleted 1, copies after 0. Red for the stated reason: the pruner keeps the five newest runs and nothing else.
Green path: M2 flips it by protecting `pre-clean/` content within the budget.
Mutant probe: a pruner that deletes nothing passes this criterion; AC-UBR-03 (ten copies leave one) and AC-UBR-06(b) reject it.

### AC-UBR-02 (P2) — the current version does not take a budget slot

Class: release-blocking. Covers: REQ-UBR-003, REQ-UBR-004.
Given one file backed up in five runs as v1, v2, v3, v4, v4 (oldest first),
When the pruner runs with keepCount 1 and budget 3,
Then the runs holding v1, v2 and v3 remain, the older run holding v4 is removed, and the newest run (v4) remains.
RED-now: ledger E2 — `remaining=[update-2026-10-05T00-00-00Z]`, `v1-entry-survives=false v2-entry-survives=false v3-entry-survives=false`. Red because the pruner keeps only the window.
Green path: M1 selection step (table test) and M2 integration.
Mutant probe: a pruner that keeps every excess run passes this criterion; AC-UBR-03 rejects it. A pruner that counts the current version against the budget keeps only v2 and v3 here and fails.

### AC-UBR-03 (P3) — redundant copies are not all kept

Class: release-blocking. Covers: REQ-UBR-005.
Given ten runs that each hold v1 of one file and a newer run that holds v2,
When the pruner runs with keepCount 1 and budget 3,
Then exactly one run holding v1 remains (the newest of the ten), nine are removed, and the run holding v2 remains.
RED-now: ledger E3 — `deleted=10`, `v1-holding-entries-remaining=0`. Red because the present pruner keeps only the window; the audit's "keeps all ten" symptom belongs to the reverted policy code and is not what this tree does.
Green path: M1 selection step and M2.
Mutant probe: keep-everything fails here (ten remain); remove-everything fails here (zero remain); keeping the oldest of the ten instead of the newest fails the "newest" clause.

### AC-UBR-04 (P4a) — a symbolic `.agents/skills` is refused

Class: release-blocking. Covers: REQ-UBR-016.
Given project A whose `.agents/skills` is a symbolic link to project B's `.agents/skills`, B holding a live link `moai-kanban-foreman` that points at `../../.claude/skills/moai-kanban-foreman`, and A holding the retired folder with only `SKILL.md`,
When the skill-residue cleanup runs on A with the successor shipped,
Then nothing in B changes, A's retired folder stays in `.claude/skills`, and the result reports a refusal naming `.agents/skills`.
RED-now: ledger E4 — actions `reported-retired-skill` and `reported-retired-link`, no refusal. Red because the root refusal list stops at `.claude/skills`, so B's link is read as A's. (`b-link-survives=true` today only because no removal path exists on this tree.)
Green path: M3.
Mutant probe: a cleanup that refuses every pass passes AC-UBR-04 and AC-UBR-05; AC-UBR-15 (an ordinary project archives its retired folder) rejects it.

### AC-UBR-05 (P4b) — a symbolic `.moai/archive` is refused

Class: release-blocking. Covers: REQ-UBR-015.
Given project A whose `.moai/archive` is a symbolic link to a directory outside the project, and a retired folder in A,
When the cleanup runs on A,
Then nothing is written outside the project, the retired folder stays in `.claude/skills`, and the result reports a refusal naming `.moai/archive`.
RED-now: ledger E5 — actions `reported-retired-skill` only, no refusal. Red because no archive-chain check exists; the harm itself is unreachable today (`content-now-outside-project=false`).
Green path: M3.
Mutant probe: checking only the leaf `archive` and not `archive/skills` passes this fixture; the AC-UBR-15 variant with a symbolic `.moai/archive/skills` rejects it.

### AC-UBR-06 — real-data remeasure on the mo.ai.kr backup roots

Class: release-blocking. Covers: REQ-UBR-001, REQ-UBR-003 to REQ-UBR-006, REQ-UBR-010, REQ-UBR-011.
Given a modification-time-preserving copy of the mo.ai.kr project's `.moai-backups` and `.moai/backups` (copied read-only from the live project; no `moai update` run; the copy command recorded),
When the harness runs the pruner on the copy twice: with the production keepCount 5, and with keepCount 4 so that the oldest config run (the one holding `pre-clean/`) becomes an excess run,
Then (a) the number of window runs is at most keepCount; (b) every retained excess run is named in the report together with at least one protected version it holds, and their count equals the irredundant cover the harness computes independently from the pre-prune state; (c) no path loses a version that is among its N most recent non-window versions; (d) under keepCount 4, at least one file under the retained `pre-clean/` copy remains; (e) the report states the number of runs removed and retained.
RED-now: ledger E6 (10-14 snapshot of the same roots): namespace root 79 directories to 0, config root 20 to 20, 1700 of 1951 distinct file versions lost across 996 paths, `pre-clean` directories 1 to 1 (the pre-clean run is inside the window only because it is one of five). Red against (b), (c) and the keepCount 4 form of (d) by construction: nothing is protected.
Green path: M6 after M2.
Mutant probe: keep-everything passes (b)-(d) trivially and fails (a); remove-everything passes (a) and fails (c) and (d). The independent cover computation in (b) must be a brute-force check of coverage and irredundancy, not a re-run of the production selection.
Note: the retained-excess count is the figure the operator uses for decision-index Q2 and Q4. This criterion is drafted for the soft cap; a ruling for a hard cap changes (a).

## Standard criteria

### AC-UBR-07 — the budget bounds protection

Class: standard. Covers: REQ-UBR-003, REQ-UBR-004, REQ-UBR-007.
Given one path with five distinct versions each held only by excess runs, and the budget set to 3 and then to 1,
When the pruner runs,
Then exactly the 3 (then 1) most recent versions remain held by a retained run and the older versions are not.

### AC-UBR-08 — bookkeeping is recognized by what the writer wrote

Class: standard. Covers: REQ-UBR-009.
Given a config backup that holds a user file named `.complete` with content `user A` in an excess run, whose newer copy in the window reads `user B`, and a second excess run that differs from a retained run only by the writer's `backup_metadata.json`,
When the pruner runs,
Then the run holding `user A` is retained as unique content, and the run differing only by writer-produced bookkeeping is removed.

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
RED-now: ledger E7 and E8 — `grep -c` of the key in the template returns 0 (exit 1) and of the default constant in `internal/config/defaults.go` returns 0 (exit 1).
Green path: M1.

### AC-UBR-11 — unclassifiable content fails closed and never blocks

Class: standard. Covers: REQ-UBR-008.
Given an excess run that holds an unreadable file and an excess run that holds a FIFO named like a file,
When the pruner runs,
Then both runs are retained and reported as retained for that reason, and the prune returns without opening the FIFO (bounded by a test timeout).

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

### AC-UBR-15 — an ordinary retired skill is archived, once

Class: standard. Covers: REQ-UBR-015.
Given a project with a retired skill folder holding only `SKILL.md`, its successor shipped, and a regular `.moai` directory,
When the cleanup runs twice,
Then the first run moves the folder to `.moai/archive/skills/retired-<stamp>/<name>`, the second run changes nothing, and a retired folder holding any other file stays where it is. A variant with a symbolic `.moai/archive/skills` is refused like AC-UBR-05.

### AC-UBR-16 — only the exact link is removed

Class: standard. Covers: REQ-UBR-017.
Given the archive in AC-UBR-15 with three variants of `.agents/skills/<name>`: a link to `.claude/skills/<name>`, a link elsewhere, and a real folder,
When the cleanup runs,
Then only the first is removed; the other two are left and are not reported as removed.

### AC-UBR-17 — a swapped component cannot redirect a move (F24)

Class: standard. Covers: REQ-UBR-018.
Given a skill folder that a `Shipped` callback replaces with a symbolic link to an outside directory holding `skill.md` just before the rename,
When the rename step runs,
Then the outside file is not renamed.

### AC-UBR-18 — a rename does not overwrite (F25)

Class: standard. Covers: REQ-UBR-018.
Given a user file created at the temporary name just before the first move of the rename,
When the rename step runs,
Then that file's bytes survive, and the step either completes under another name or restores the content under its own name and reports.

### AC-UBR-19 — an open zero-byte database survives housekeeping (F21)

Class: standard (regression guard; mechanism per decision-index Q1). Covers: REQ-UBR-019.
Given a zero-byte SQLite file under `.moai/state` opened by a process that has not written to it, with an old modification time,
When the state-residue pass runs,
Then the process's next write succeeds. The probe is the t1687 audit's A3-P3 (iter3), re-created in a test.

### AC-UBR-20 — destructive sites are registered

Class: standard. Covers: REQ-UBR-020.
Given the tree after M2 and M3,
When `go test -count=1 -run '^TestDestructiveTargetRegistry_CoversAllSites$' ./internal/cli/` runs,
Then it passes, every removal or move call site added or changed by this SPEC has a registry row with a protection or an exemption reason, and the pin test sees the new move calls (it fails on a deliberately added unregistered alias, observed once).

### AC-UBR-21 — the key and the report rows are pinned by tests

Class: standard. Covers: REQ-UBR-021.
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
Then the embedded files regenerate, the tests pass, and the working tree shows no tracked change beyond the intended ones.

### AC-UBR-24 — the probe tests are red before the fix, in the commit graph

Class: standard. Covers: ordering attribution (`verification-claim-integrity.md` §2.3).
Given the card branch after run-phase close,
When the commit graph is read,
Then each probe test of AC-UBR-01 to AC-UBR-05 is added in a commit that precedes the commit that makes it pass, and progress.md §E.2 holds the verbatim RED output for each.

## Edge cases

- Only one backup root exists; the other is absent. A run that spans both roots with one root containing the protected version: the whole run is kept or removed together.
- keepCount larger than the number of runs; budget larger than the number of versions.
- A run that holds only bookkeeping files and `.template-defaults/`: removable.
- Two copies of one logical file at different paths inside backups (`pre-clean/` and a namespace copy): distinct versions under the Q5 default.
- A file whose content is large: hashed by streaming; no whole-file buffering requirement.
- A removal that fails midway through a run: the run is reported as retained with the failure reason.
- A backup directory whose name matches no writer shape and carries no marker: neither counted nor deleted (unchanged).

## Definition of Done

1. AC-UBR-01 to AC-UBR-24 are PASS with verbatim command output, or recorded as a Gap with the reason.
2. `decision-index.md` Q1, Q2 and Q3 carry operator verdicts; Q4 carries the AC-UBR-06 figure.
3. The registry guard, build for linux, darwin and windows, gofmt, vet, lint and the coverage floor are clean for the touched packages.
4. progress.md §E.2 holds the RED outputs and the real-data measurement; the card branch is committed, unpushed.
5. No `moai update` was run in any tree, and no live project was written.
