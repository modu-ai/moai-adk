---
id: SPEC-UPDATE-BACKUP-RETENTION-001
title: "moai update — backup retention that protects unique content inside a per-file budget, and report-only skill-folder and database residue"
version: "0.2.0"
status: completed
created: 2026-10-16
updated: 2026-10-16
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli/update"
lifecycle: spec-anchored
tier: L
tags: "cli, update, backup, retention, pre-clean, skill-residue, data-loss, card-t1733"
---

# SPEC-UPDATE-BACKUP-RETENTION-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-10-16 | Initial draft (card t1733, Class C). Respecifies the backup retention policy that card t1687 reverted (commit 8e43f064a) and carries the t1687 audit items F1, F13-F19, F21, F24-F26. |
| 0.2.0 | 2026-10-16 | Revision after plan-audit FAIL (receipt rcpt-23e014c572cca1b636f63244, D1-D14) and the operator answers to Q1-Q3 (2026-10-16): zero-byte database is report-only; soft cap confirmed; archive and link removal are not reintroduced (report-only; authority basis deferred to card 3.3). Operator principle for 3.2: add no data-deleting behavior. Baseline corrected: the loss is pre-existing on main `ec5a4f392`, not introduced by t1687 (reproduction report). Revision log: progress.md § Revision log. Related SPECs (moved from frontmatter, D13): SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-UPDATE-REPAIR-RC30-001, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001. |

## §A Problem / Motivation

### A.1 Why this SPEC exists

`moai update` stays forbidden in this repository until this card lands (operator ruling, decision board record a041, recorded in `research.md` § Operator rulings). The reason is that the update's backup pruning deletes backups without regard to what they hold. Card t1687 had written a retention policy to prevent that; its audits found defects in it (F1, F13, F14, F15), and the operator split the card: t1687 landed the rest of its scope with the retention policy reverted, and the repair moved here.

### A.2 What the tree does today (measured)

Every figure below is a measurement; the commands and verbatim output are in `research.md` § Evidence ledger. The reference tree is `e669da81b` (code identical at the SPEC commit `3b133bedf`).

- **The defect is pre-existing on main.** The reproduction report (`/private/tmp/t1687-repro-report.md`) ran the same probes on main `ec5a4f392`, on t1687 `fb7e1933b` and on the pre-revert tree `8e43f064a^`: main and t1687 give identical output (P1 user-file copies 1 to 0; P2 no v1, v2 or v3 entry survives; P3 ten v1-holding entries to 0; namespace root 79 directories to 0; 1700 of 1951 distinct file versions lost on both), while the pre-revert tree keeps everything (79 kept, loss 0). Commit `8e43f064a` removed a protection main never had; it did not introduce the loss.
- The pruner counts backup runs and keeps the newest keepCount. Nothing else is kept. A real deploy puts a user skill file into a backup's `pre-clean/` directory; once five newer runs exist the pruner removes that run, and the user file's copies go from 1 to 0 (E1).
- A file backed up as v1, v2, v3, v4, v4 with keepCount 1 keeps only the newest run (E2). Ten runs holding v1 followed by one holding v2 leave no copy of v1 (E3).
- On a clone of the mo.ai.kr backup roots (10-14 snapshot, modification times preserved), one prune with keepCount 5 empties the namespace root (79 to 0), leaves the config root at 20, and removes 1700 of 1951 distinct file versions across 996 paths (E6).
- The skill-residue cleanup reports retired skill folders and `.agents/skills` links and moves or removes nothing (E4, E5). The operator has ruled that it stays that way (Q3).

The symptom the t1687 audit named for F1 and F14 (a cap that removes nothing) belongs to the reverted policy code, which no longer exists. The defect to fix is the opposite one: unconditional deletion.

### A.3 Finding disposition

`research.md` § Finding verification gives file:line evidence per row. "Requirement" means the finding is specified here.

| Finding | Subject | Disposition |
|---|---|---|
| F1, F14 | Cap did not shrink the real namespace root; redundant copies all kept | Requirement: REQ-UBR-005 (irredundant retention) and REQ-UBR-010 |
| F13 | Budget counted the current version as one slot | Requirement: REQ-UBR-004 |
| F15 | `pre-clean/` exempted from protection | Requirement: REQ-UBR-006 (operator ruling: protect; not reopened) |
| F16 | `.agents/skills` link root unchecked | Operator ruling Q3: no link removal; REQ-UBR-016 (content reached through the link is never touched) |
| F17 | `.moai/archive` parent chain unchecked | Operator ruling Q3: no archive move; REQ-UBR-015 (report-only) |
| F18 | Bookkeeping recognized by name | Requirement: REQ-UBR-009 |
| F19 | Default constant duplicated; missing symmetry and report tests | Requirement: REQ-UBR-007, REQ-UBR-020 |
| F21 | Open zero-byte database can be removed | Operator ruling Q1: report only. Requirement: REQ-UBR-018 |
| F24, F25 | Path-based rename races | Requirement: REQ-UBR-017 |
| F26 | FIFO blocks the unmanaged restore | Already fixed by card t1799 (commit 06fc853bf); no requirement |
| F20, F22, F23 | Registry row, rename alias, caller-less export | Resolved on this tree; REQ-UBR-019 keeps the guard honest for changed sites |

### A.4 Decisions

All six decisions in `decision-index.md` are resolved or closed: Q1, Q2 and Q3 carry the operator's answers of 2026-10-16; Q4 is closed (no decision depends on it); Q5 carries the published Default; Q6 is closed by the Q3 ruling.

## §B Goals

1. A prune never destroys the last copy of a user file version that the operator's per-file budget says to keep, and never keeps more than that.
2. The pruner's behavior on the real mo.ai.kr backup roots is measured, not modelled.
3. The skill-folder and state-residue passes add no data-deleting behavior: they report, and the project root bounds every move they still make.
4. The carried optional items (F21, F24, F25) are closed or explicitly ruled on.

## §C Scope Exclusions

### Out of Scope — what the backup writers produce

- The names, layout, run marker, metadata and `.complete` formats the writers use for backup directories are unchanged.
- Which files a backup contains is unchanged. The pre-clean copy rule of the deploy step is the producer contract this SPEC consumes, not one it edits.

### Out of Scope — ordering of unmarked runs

- A run without a readable marker is still ordered by its directory's modification time. This SPEC does not change that ordering. A copy of the backup roots that resets modification times changes which runs are "newest"; the real-data measurement must therefore use a modification-time-preserving copy (see § G).

### Out of Scope — restore behavior

- `moai update --restore`, the section restore and the unmanaged-config restore keep their behavior. F26 is recorded as fixed elsewhere and is not respecified.

### Out of Scope — retired-skill archive and link removal

- Moving a retired skill folder to an archive and removing an `.agents/skills` link are not reintroduced (operator ruling Q3, 2026-10-16). The authority basis for such cleanup is deferred to card 3.3.
- Refusing the whole skill-residue pass when `.agents` or `.agents/skills` is a symbolic link is not added (decision-index Q6 closed).

### Out of Scope — t1687 audit items the operator did not carry

- F8-F12, F27-F31, F34, F36, F37 and F38 were not named by the operator's carry decision. They are not respecified here; `research.md` lists what was and was not checked.

### Out of Scope — landing and live projects

- No `moai update` runs against any project in any tree for this SPEC; measurements use function-level harnesses on copies.
- The live mo.ai.kr project is never read for writing. Remote push and release batching stay with the leader.

## §D Requirements (GEARS)

Notation: GEARS (`Where` / `While` / `When` / `shall`). `N` is the per-file past-version budget (REQ-UBR-007). "Window", "excess" and "exempt run" are defined in REQ-UBR-001 and REQ-UBR-008.

### D.1 Backup retention model

**REQ-UBR-001 (retention unit and window).** The backup pruner shall count backups in runs, the unit update retention already uses, and shall keep the newest keepCount runs whole, every directory of such a run in both backup roots, as the retention window; every older run is an excess run.
*Rationale:* runs are created and pruned together (SPEC-UPDATE-REPAIR-RC30-001); splitting a run would break the restore points it describes.

**REQ-UBR-002 (version identity).** The pruner shall identify a backed-up file version by the file's path inside its backup directory together with the SHA-256 of its bytes, shall consider regular files only, and shall ignore the `.template-defaults/` render and the bookkeeping files of REQ-UBR-009.
*Rationale:* a restore resolves a file by its path in the backup, so identical bytes at another path do not make a version redundant; `.template-defaults/` is regenerated from the binary and holds no user content.

**REQ-UBR-003 (budgeted protection).** When an excess run holds a version of a file that no window run holds, the pruner shall protect that version if it is among the N most recent such versions of that path, where the recency of a version is that of the newest excess run that holds it.
*Rationale:* operator ruling a263 (C): keep only the newest N past versions per file; allow cleanup of backups whose only unique content is older than that.

**REQ-UBR-004 (current version is not a past version).** The pruner shall not count a version held by a window run against the budget of its path, so that a path whose current version also sits in excess runs keeps N past versions and not N-1.
*Rationale:* finding F13; the budget is of past versions.

**REQ-UBR-005 (irredundant retention).** The pruner shall retain excess runs such that every protected version is held by at least one retained run, shall retain no excess run whose protected versions are all held by other retained runs unless it is an exempt run (REQ-UBR-008), and shall remove every other excess run.
*Rationale:* finding F14; a copy of a version that another retained run already holds is redundant, so a cap that keeps all such copies keeps nothing back. The last copy of a protected version is never removed (operator ruling Q2).

**REQ-UBR-006 (pre-clean is user content).** The pruner shall not exempt content under a backup's `pre-clean/` directory from REQ-UBR-003 to REQ-UBR-005; the `.template-defaults/` render shall remain the only exempt content. Because the budget is at least 1, the newest non-window version of every file under `pre-clean/` shall therefore have at least one retained copy.
*Rationale:* operator ruling (board record fd8b): the deploy step copies into `pre-clean/` only files the template does not ship at the same path, so the copy is the only recovery copy of a user file.

**REQ-UBR-007 (budget setting).** The configuration shall provide the budget as `state.backup_past_versions_keep`, shipped with the default 3 and with that default defined in exactly one place, and the pruner shall treat a value below 1 as the default.
*Rationale:* operator ruling a263 (N is a configuration value, default 3). A budget below 1 would disable REQ-UBR-006.

**REQ-UBR-008 (fail closed on what cannot be classified).** Where an excess run holds an entry the pruner cannot read or classify, including an entry that is not a regular file, the pruner shall retain that run unconditionally as an exempt run, shall treat it as a holder of every key it does hold, shall report it as retained for that reason, and shall not open an entry that is not a regular file.
*Rationale:* an unclassifiable entry can make a run look redundant; keeping it errs toward the user's data. The disposition is single: an exempt run is retained whether or not its readable content is redundant (this reconciles REQ-UBR-005, REQ-UBR-010, design D2 step 4 and AC-UBR-11; audit D2). Not opening non-regular entries avoids blocking on a FIFO.

**REQ-UBR-009 (bookkeeping is recognized by what the writer wrote).** The pruner shall treat a file as bookkeeping only when it is the marker or metadata file that a backup writer produced at the root of its backup directory, and shall treat a user file that merely carries a bookkeeping name as user content.
*Rationale:* finding F18; matching by name alone misreads a user's `.moai/config/.complete`.

**REQ-UBR-010 (what the cap bounds, soft cap).** After a prune, the window shall hold at most keepCount runs; the excess runs the prune retains shall be exactly the runs REQ-UBR-005 requires plus the exempt runs of REQ-UBR-008 plus any run whose removal failed; the number of non-exempt retained excess runs shall not exceed the number of distinct protected versions; and the update shall report how many excess runs it retained and why.
*Rationale:* the operator ruled a soft cap (Q2, 2026-10-16): protection and a literal cap cannot both hold for every input, because a protected version may live only in an old run. The card's earlier wording "both backup roots at or under their cap" is replaced by "the window at keepCount plus the reported irredundant excess"; when no excess run is needed the total is at or under keepCount (audit D4).

**REQ-UBR-011 (per-run reporting).** When a prune removes a run, retains an excess run, or retains an exempt run, the update shall report each such run by its path, and for a retained run shall name the reason.
*Rationale:* operator ruling a263: each removal is reported.

**REQ-UBR-012 (unmanaged entries are reported, never deleted).** The update shall report the count and total size of the entries under `.moai-backups` whose names no backup writer produces, and shall not delete them.
*Rationale:* operator ruling a263 (non-timestamp entries: report only).

**REQ-UBR-013 (root containment).** Where a backup root is a symbolic link or leads outside the project, the pruner shall prune nothing and the update shall report the refusal.
*Rationale:* behavior that exists today; this requirement keeps it a guarded invariant of the rewritten pruner.

**REQ-UBR-014 (explicit wipe and negative).** Where keepCount is 0, the pruner shall remove every run without applying REQ-UBR-003 to REQ-UBR-006, and where keepCount is negative, the pruner shall remove nothing.
*Rationale:* both are existing semantics of the pruner's contract.

### D.2 Skill-folder residue (report-only)

**REQ-UBR-015 (retired skill folders are reported, not moved or removed).** When a retired skill folder is present, the skill-residue cleanup shall report it and shall not move, rename or remove the folder or any entry inside it.
*Rationale:* operator ruling Q3 and the 3.2 principle (no new data-deleting behavior). This is the behavior of the tree today (E4, E5); the requirement pins it so that a later reintroduction of removal fails a test (the archive path and its F17 parent-chain defect are therefore unreachable).

**REQ-UBR-016 (content behind a link root is left untouched).** Where `.agents`, `.agents/skills` or `.moai/archive` is a symbolic link, the skill-residue cleanup shall not remove, move or write any entry reached through that link, and the pass shall report the link entries it finds without acting on them.
*Rationale:* operator ruling Q3; finding F16 (a link root lets another project's link read as this project's). Under report-only the link content survives by construction; the requirement keeps that invariant testable (AC-UBR-04, AC-UBR-05). Refusing the whole pass is not added (decision-index Q6).

**REQ-UBR-017 (the held-back rename is bound to the project root).** When the cleanup applies the held-back `skill.md` to `SKILL.md` rename, it shall perform the rename through a handle bound to the project root, so that a path component replaced by a symbolic link after it was checked cannot redirect the operation outside the project, and the rename shall not overwrite a file that appears at its destination after that destination was checked.
*Rationale:* findings F24 and F25 (operator carry decision 242e). The rename exists today and deletes nothing; this narrows it. The documented threat model is a single process; this is defense in depth with no new dependency.

### D.3 State residue

**REQ-UBR-018 (an empty database file is reported, never removed).** When the state-residue pass finds a zero-byte `*.db` file under `.moai/state`, the pass shall report "empty DB file present" with its path and shall not remove it.
*Rationale:* operator ruling Q1 (2026-10-16) and finding F21: a process that opened the file without writing takes an I/O error on its next write if the file is unlinked. The zero-byte `backlog.json.lock` removal in the same function is outside this requirement and unchanged.

### D.4 Cross-cutting

**REQ-UBR-019 (destructive-site registry).** Every call site this SPEC adds or changes that removes or moves user-visible files shall have a row in the destructive-target registry stating its protection or exemption, and the registry drift guard shall pass.
*Rationale:* card t1687 failed an audit for exactly this omission (F20).

**REQ-UBR-020 (test surface for the key and the report rows).** The shipped template and the configuration struct shall agree on `state.backup_past_versions_keep` under the struct-YAML symmetry tests, and the full-reinstall path and the template-sync path shall each have a test that asserts the retention report rows.
*Rationale:* finding F19 (ii) and (iii).

## §E Non-Functional Constraints

- No new external dependency; the module's existing toolchain (Go 1.26.8, `os.Root` with `Rename`, `Lstat`, `Mkdir`) is sufficient.
- Cross-platform: darwin, linux and windows. Tests that depend on POSIX permission bits, FIFOs or hard links are skipped where the platform lacks them.
- A prune reads each regular file of the backup roots at most once and opens no non-regular entry; its cost grows linearly with backup bytes (measured baseline on the 519 MB namespace clone: one prune plus two version censuses in 9 to 15 s, E6).
- Commands and test harnesses stay bounded in output; scratch files live under the default temporary directory, never in the tree or the home directory.
- The update lock already serializes `moai update`; this SPEC adds no new concurrency.

## §F Success Criteria

1. The probe-shaped tests P1-P3 (AC-UBR-01 to AC-UBR-03) are red on main `ec5a4f392` and on tree `e669da81b` for the stated reasons and green on the tree the repair is measured on; P4a and P4b (AC-UBR-04, AC-UBR-05) are green regression guards on `e669da81b` and stay green.
2. The real-data criterion (AC-UBR-06) holds on a modification-time-preserving clone of the mo.ai.kr backup roots: every protected version keeps a retained holder, every retained excess run is justified, no path loses a version inside its budget, and at least one pre-clean user-file copy remains.
3. The destructive-target registry guard, the symmetry and report-row tests, `gofmt`, `go vet`, `golangci-lint`, the 85% coverage floor and the Template-First build hold for the touched packages (AC-UBR-19 to AC-UBR-23).
4. The operator rulings on Q1, Q2 and Q3 are recorded (`decision-index.md`, 2026-10-16); no open clarification marker remains.

## §G Risk Register

| # | Risk | Mitigation |
|---|------|------------|
| 1 | Unmarked runs are ordered by directory modification time; a copy that resets it changes the window. | The real-data harness copies with modification times preserved and records the copy command. |
| 2 | Protection plus a budget can leave many excess runs on a project with many distinct file versions. | REQ-UBR-005 keeps the retained set irredundant; REQ-UBR-010 bounds it by the protected-version count and requires it to be reported; the operator ruled the soft cap (Q2). |
| 3 | A symbolic `.agents/skills` still reads another project's link as this project's in the report. | Report-only: nothing is removed through it (REQ-UBR-016); a refusal is a separate card (Q6). |
| 4 | `os.Root` operations can fail on filesystems without hard-link or rename support. | The rename fails closed: it leaves the content under its own name and reports (design D5). |
| 5 | Hashing every backup file is the dominant cost on large roots. | Linear and bounded; measured baseline in E6; version censuses reuse one walk (design D2). |
| 6 | The registry guard counts call expressions, so an aliased `os.Rename` is invisible to it. | REQ-UBR-019 and the pin test already used for `moveOne`. |
| 7 | A user file that imitates a bookkeeping writer's shape is read as bookkeeping and is not censused (audit D8). | Accepted residual, recorded in design D4; no writer produces such a file and the audits did not find the hazard; revisit if provenance (backup kind from `backup_metadata.json`) is wanted. |

## §H Cross-References

- `research.md` — evidence ledger, per-finding verification, probe sources, baseline on main.
- `design.md` — retention algorithm, reference behavior, report shape, residue changes, post-landing verification procedure.
- `plan.md` — milestones ordered by reversibility, assumptions.
- `acceptance.md` — AC-UBR-01 to AC-UBR-24, main-versus-tree probe cells, real-data criterion.
- `decision-index.md` — decisions and their resolution.
- `.claude/rules/moai/development/spec-frontmatter-schema.md`, `.claude/rules/moai/development/verification-completeness.md`.
- Producer contract for `pre-clean/`: `internal/cli/update/deploy/deploy.go:99-115`.
