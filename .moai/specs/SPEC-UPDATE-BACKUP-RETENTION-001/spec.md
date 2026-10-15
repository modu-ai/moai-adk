---
id: SPEC-UPDATE-BACKUP-RETENTION-001
title: "moai update — backup retention that protects unique content inside a per-file budget, plus symlink-safe skill-folder residue cleanup"
version: "0.1.0"
status: draft
created: 2026-10-16
updated: 2026-10-16
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli/update"
lifecycle: spec-anchored
tier: L
tags: "cli, update, backup, retention, pre-clean, symlink, skill-residue, data-loss, card-t1733"
related_specs: [SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-UPDATE-REPAIR-RC30-001, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001]
---

# SPEC-UPDATE-BACKUP-RETENTION-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-10-16 | Initial draft (card t1733, Class C). Respecifies the backup retention policy and the retired-skill archive / link-removal path that card t1687 reverted to main's behavior (commit 8e43f064a), and carries the t1687 audit items F1, F13-F19, F21, F24-F26. Evidence and per-finding disposition: `research.md`. |

## §A Problem / Motivation

### A.1 Why this SPEC exists

`moai update` stays forbidden in this repository until this card lands (operator ruling, decision board record a041, recorded in `research.md` § Operator rulings). The reason is that the update's backup pruning, as it stands on main, deletes backups without regard to what they hold. Card t1687 had written a retention policy to prevent that; its audits found defects in it (F1, F13, F14, F15), and the operator split the card: t1687 landed the rest of its scope with the retention policy and the retired-skill archive and link-removal path **reverted to main's behavior**, and the repair moved here.

### A.2 What the tree does today (measured, tree `e669da81b`)

Every figure below is a measurement from this run; the commands and verbatim output are in `research.md` § Evidence ledger.

- The pruner counts backup runs and keeps the newest keepCount. Nothing else is kept. A real deploy puts a user skill file into a backup's `pre-clean/` directory; once five newer runs exist the pruner removes that run, and the user file's copies go from 1 to 0 (E1).
- A file backed up as v1, v2, v3, v4, v4 with keepCount 1 keeps only the newest run; the runs holding v1, v2 and v3 are gone (E2). Ten runs holding v1 followed by one holding v2 leave no copy of v1 (E3).
- On a clone of the mo.ai.kr backup roots taken on 2026-10-14, one prune with keepCount 5 empties the namespace root (79 directories to 0), leaves the config root at 20, and removes 1700 of 1951 distinct file versions across 996 paths (E6).
- The skill-residue cleanup reports retired skill folders and `.agents/skills` links and moves or removes nothing, so its link and archive hazards (F16, F17) are unreachable today (E4, E5). Reintroducing the archive and link removal without the missing root checks would make them reachable again.

The symptom the t1687 audit named for F1 and F14 (a cap that removes nothing) belongs to the reverted policy code, which no longer exists in this tree. The defect to fix on this tree is the opposite one: unconditional deletion.

### A.3 Finding disposition

`research.md` § Finding verification gives file:line evidence per row. "Requirement" means the finding is specified here.

| Finding | Subject | Disposition |
|---|---|---|
| F1, F14 | Cap did not shrink the real namespace root; redundant copies all kept | Requirement: REQ-UBR-005 (irredundant retention) and REQ-UBR-010 |
| F13 | Budget counted the current version as one slot | Requirement: REQ-UBR-004 |
| F15 | `pre-clean/` exempted from protection | Requirement: REQ-UBR-006 (operator ruling: protect; not reopened) |
| F16 | `.agents/skills` link root unchecked | Requirement: REQ-UBR-016, REQ-UBR-017 |
| F17 | `.moai/archive` parent chain unchecked | Requirement: REQ-UBR-015 |
| F18 | Bookkeeping recognized by name | Requirement: REQ-UBR-009 (in scope; debt closed cheaply while the policy is rewritten) |
| F19 | Default constant duplicated; missing symmetry and report tests | Requirement: REQ-UBR-007, REQ-UBR-021 (in scope) |
| F21 | Open zero-byte database can be removed | Requirement: REQ-UBR-019 (mechanism awaits the operator) |
| F24, F25 | Path-based rename races | Requirement: REQ-UBR-018 |
| F26 | FIFO blocks the unmanaged restore | Already fixed by card t1799 (commit 06fc853bf); no requirement |
| F20, F22, F23 | Registry row, rename alias, caller-less export | Already resolved on this tree; REQ-UBR-020 keeps the guard honest for new sites |

### A.4 Open decisions

Six decisions are not settled by committed authority. They are listed, with their routing, in `decision-index.md`. The three product-level ones (Q1, Q2, Q3) are carried as clarification markers in `plan.md` and block the plan-to-run Kickoff; the one needing a measurement (Q4) is resolved by AC-UBR-06; Q5 and Q6 are implementation-level and carry the published Default.

## §B Goals

1. A prune never destroys the last copy of a user file version that the operator's per-file budget says to keep, and never keeps more than that.
2. The pruner's behavior on the real mo.ai.kr backup roots is measured, not modelled.
3. The retired-skill archive and the `.agents` link removal, if reintroduced, cannot reach outside the project through a symbolic link at any component of their paths.
4. The carried optional items (F21, F24, F25) are closed or explicitly ruled on.

## §C Scope Exclusions

### Out of Scope — what the backup writers produce

- The names, layout, run marker, metadata and `.complete` formats the writers use for backup directories are unchanged.
- Which files a backup contains is unchanged. The pre-clean copy rule of the deploy step is the producer contract this SPEC consumes, not one it edits.

### Out of Scope — ordering of unmarked runs

- A run without a readable marker is still ordered by its directory's modification time. This SPEC does not change that ordering. A copy of the backup roots that resets modification times changes which runs are "newest"; the real-data measurement must therefore use a modification-time-preserving copy (see § G).

### Out of Scope — restore behavior

- `moai update --restore`, the section restore and the unmanaged-config restore keep their behavior. F26 is recorded as fixed elsewhere and is not respecified.

### Out of Scope — t1687 audit items the operator did not carry

- F8-F12, F27-F31, F34, F36, F37 and F38 were not named by the operator's carry decision. They are not respecified here; `research.md` lists what was and was not checked.

### Out of Scope — landing and live projects

- No `moai update` runs against any project in any tree for this SPEC; measurements use function-level harnesses on copies.
- The live mo.ai.kr project is never read for writing. Remote push and release batching stay with the leader.

## §D Requirements (GEARS)

Notation: GEARS (`Where` / `While` / `When` / `shall`). `N` is the per-file past-version budget (REQ-UBR-007). "Window" and "excess" are defined in REQ-UBR-001.

### D.1 Backup retention model

**REQ-UBR-001 (retention unit and window).** The backup pruner shall count backups in runs, the unit update retention already uses, and shall keep the newest keepCount runs whole, every directory of such a run in both backup roots, as the retention window; every older run is an excess run.
*Rationale:* runs are created and pruned together (SPEC-UPDATE-REPAIR-RC30-001); splitting a run would break the restore points it describes.

**REQ-UBR-002 (version identity).** The pruner shall identify a backed-up file version by the file's path inside its backup directory together with the SHA-256 of its bytes, shall consider regular files only, and shall ignore the `.template-defaults/` render and the bookkeeping files of REQ-UBR-009.
*Rationale:* a restore resolves a file by its path in the backup, so identical bytes at another path do not make a version redundant; `.template-defaults/` is regenerated from the binary and holds no user content.

**REQ-UBR-003 (budgeted protection).** When an excess run holds a version of a file that no window run holds, the pruner shall protect that version if it is among the N most recent such versions of that path, where the recency of a version is that of the newest excess run that holds it.
*Rationale:* operator ruling a263 (C): keep only the newest N past versions per file; allow cleanup of backups whose only unique content is older than that.

**REQ-UBR-004 (current version is not a past version).** The pruner shall not count a version held by a window run against the budget of its path, so that a path whose current version also sits in excess runs keeps N past versions and not N-1.
*Rationale:* finding F13; the budget is of past versions.

**REQ-UBR-005 (irredundant retention).** The pruner shall retain excess runs such that every protected version is held by at least one retained run, shall retain no excess run whose protected versions are all held by other retained runs, and shall remove every other excess run.
*Rationale:* finding F14; a copy of a version that another retained run already holds is redundant, so a cap that keeps all such copies keeps nothing back.

**REQ-UBR-006 (pre-clean is user content).** The pruner shall not exempt content under a backup's `pre-clean/` directory from REQ-UBR-003 to REQ-UBR-005; the `.template-defaults/` render shall remain the only exempt content. Because the budget is at least 1, the newest non-window version of every file under `pre-clean/` shall therefore have at least one retained copy.
*Rationale:* operator ruling (board record fd8b): the deploy step copies into `pre-clean/` only files the template does not ship at the same path, so the copy is the only recovery copy of a user file. The earlier premise "regenerable from the binary" is withdrawn.

**REQ-UBR-007 (budget setting).** The configuration shall provide the budget as `state.backup_past_versions_keep`, shipped with the default 3 and with that default defined in exactly one place, and the pruner shall treat a value below 1 as the default.
*Rationale:* operator ruling a263 (N is a configuration value, default 3). A budget below 1 would disable REQ-UBR-006.

**REQ-UBR-008 (fail closed on what cannot be classified).** Where an excess run holds an entry the pruner cannot read or classify, the pruner shall retain that run and report it as retained for that reason, and the pruner shall not open an entry that is not a regular file.
*Rationale:* an unclassifiable entry can make a run look redundant; keeping it errs toward the user's data. Not opening non-regular entries avoids blocking on a FIFO.

**REQ-UBR-009 (bookkeeping is recognized by what the writer wrote).** The pruner shall treat a file as bookkeeping only when it is the marker or metadata file that a backup writer produced at the root of its backup directory, and shall treat a user file that merely carries a bookkeeping name as user content.
*Rationale:* finding F18; matching by name alone misreads a user's `.moai/config/.complete`.

**REQ-UBR-010 (what the cap bounds).** After a prune, the number of window runs shall be at most keepCount, the excess runs the prune retains shall be exactly those REQ-UBR-005 requires, and the update shall report how many it retained and why.
*Rationale:* protection and a literal cap cannot both hold for every input, because a protected version may live only in an old run. This requirement fixes which of the two bounds is literal; the operator's ruling on the stricter alternative is open (`decision-index.md` Q2).

**REQ-UBR-011 (per-run reporting).** When a prune removes a run, retains an excess run, or retains a run under REQ-UBR-008, the update shall report each such run by its path, and for a retained run shall name the reason.
*Rationale:* operator ruling a263: each removal is reported.

**REQ-UBR-012 (unmanaged entries are reported, never deleted).** The update shall report the count and total size of the entries under `.moai-backups` whose names no backup writer produces, and shall not delete them.
*Rationale:* operator ruling a263 (non-timestamp entries: report only).

**REQ-UBR-013 (root containment).** Where a backup root is a symbolic link or leads outside the project, the pruner shall prune nothing and the update shall report the refusal.
*Rationale:* behavior that exists today; this requirement keeps it a guarded invariant of the rewritten pruner.

**REQ-UBR-014 (explicit wipe and negative).** Where keepCount is 0, the pruner shall remove every run without applying REQ-UBR-003 to REQ-UBR-006, and where keepCount is negative, the pruner shall remove nothing.
*Rationale:* both are existing semantics of the pruner's contract.

### D.2 Skill-folder residue

**REQ-UBR-015 (archive path is contained).** When a retired skill folder holds nothing besides its `SKILL.md` and the template ships its successor, the skill-residue cleanup shall move the folder to `.moai/archive/skills/retired-<stamp>/<name>`, and shall do so only when every component of that destination path inside the project is a real directory; otherwise the cleanup shall leave the folder where it is and report the refusal naming the offending path.
*Rationale:* finding F17; a symbolic link at `.moai/archive` otherwise carries the folder outside the project. A retired folder that holds other files stays where it is, as today.

**REQ-UBR-016 (link roots are checked).** Where `.agents` or `.agents/skills` is a symbolic link, the skill-residue cleanup shall refuse the pass, change nothing, and report the link, in the way it already refuses a symbolic `.moai`, `.claude` or `.claude/skills`.
*Rationale:* finding F16; the link comparison is by path text, so a link root lets another project's link read as this project's.

**REQ-UBR-017 (link removal is exact).** When the cleanup has archived a retired skill folder, it shall remove `.agents/skills/<name>` only if that entry is a symbolic link whose target resolves to the folder's former location `.claude/skills/<name>`, and shall leave any other link and any real folder in that place.
*Rationale:* finding F16; removing anything else could delete a user's own link or folder.

**REQ-UBR-018 (moves are bound to the project root).** The cleanup shall perform each move and removal through a handle bound to the project root, so that a path component replaced by a symbolic link after it was checked cannot redirect the operation outside the project, and a rename shall not overwrite a file that appears at its destination after that destination was checked.
*Rationale:* findings F24 and F25 (operator carry decision 242e). The documented threat model is a single process; this is defense in depth with no new dependency.

### D.3 State residue

**REQ-UBR-019 (an open database survives housekeeping).** The state-residue pass shall not remove a zero-byte database file in a way that makes a process holding it open fail on its next write.
*Rationale:* finding F21. The mechanism (report-only, or a narrower removal list) is an operator ruling (`decision-index.md` Q1); this requirement fixes the outcome.

### D.4 Cross-cutting

**REQ-UBR-020 (destructive-site registry).** Every call site this SPEC adds or changes that removes or moves user-visible files shall have a row in the destructive-target registry stating its protection or exemption, and the registry drift guard shall pass.
*Rationale:* card t1687 failed an audit for exactly this omission (F20).

**REQ-UBR-021 (test surface for the key and the report rows).** The shipped template and the configuration struct shall agree on `state.backup_past_versions_keep` under the struct-YAML symmetry tests, and the full-reinstall path and the template-sync path shall each have a test that asserts the retention report rows.
*Rationale:* finding F19 (ii) and (iii).

## §E Non-Functional Constraints

- No new external dependency; the module's existing toolchain (Go 1.26.8, `os.Root` with `Rename`, `Remove`, `Lstat`, `Mkdir`) is sufficient.
- Cross-platform: darwin, linux and windows. Tests that depend on POSIX permission bits, FIFOs or hard links are skipped where the platform lacks them.
- A prune reads each regular file of the backup roots at most once and opens no non-regular entry; its cost grows linearly with backup bytes (measured baseline on the 519 MB namespace clone: one prune plus two version censuses in 9 to 15 s across two runs, E6).
- Commands and test harnesses stay bounded in output; scratch files live under the default temporary directory, never in the tree or the home directory.
- The update lock already serializes `moai update`; this SPEC adds no new concurrency.

## §F Success Criteria

1. The probe-shaped tests P1-P4 (acceptance AC-UBR-01 to AC-UBR-05) are red on tree `e669da81b` for the stated reasons and green after the change.
2. The real-data criterion (AC-UBR-06) holds on a modification-time-preserving clone of the mo.ai.kr backup roots: the window is at or under keepCount, every retained excess run is named with the protected version it holds, no path loses a version inside its budget, and at least one pre-clean user-file copy remains.
3. The destructive-target registry guard, the symmetry and report-row tests, `gofmt`, `go vet`, `golangci-lint`, the 85% coverage floor and the Template-First build hold for the touched packages (AC-UBR-20 to AC-UBR-23).
4. Operator rulings on `decision-index.md` Q1, Q2 and Q3 are recorded before run-phase entry.

## §G Risk Register

| # | Risk | Mitigation |
|---|------|------------|
| 1 | Unmarked runs are ordered by directory modification time; a copy that resets it changes the window. | The real-data harness copies with modification times preserved and records the copy command. |
| 2 | Protection plus a budget can leave many excess runs on a project with many distinct file versions. | REQ-UBR-005 keeps the retained set irredundant; REQ-UBR-010 requires it to be reported; the operator decides on a stricter cap (Q2) once AC-UBR-06 reports the real number. |
| 3 | A user who symlinks `.agents/skills` on purpose now gets a refusal row instead of skill cleanup. | The refusal names the path; behavior matches the existing treatment of the other roots (design D5). |
| 4 | `os.Root` operations can fail on filesystems without hard-link or rename support. | The cleanup fails closed: it leaves the content under its own name and reports (design D5). |
| 5 | Hashing every backup file is the dominant cost on large roots. | Linear and bounded; measured baseline in E6; version censuses reuse one walk (design D2). |
| 6 | The registry guard counts call expressions, so an aliased `os.Rename` is invisible to it. | REQ-UBR-020 and the pin test already used for `moveOne`. |

## §H Cross-References

- `research.md` — evidence ledger, per-finding verification, probe sources.
- `design.md` — retention algorithm, options considered, report shape, residue changes.
- `plan.md` — milestones ordered by reversibility, assumptions, clarification markers.
- `acceptance.md` — AC-UBR-01 to AC-UBR-24, probe-shaped tests, real-data criterion.
- `decision-index.md` — unresolved decisions and their routing.
- `.claude/rules/moai/development/spec-frontmatter-schema.md`, `.claude/rules/moai/development/verification-completeness.md`.
- Producer contract for `pre-clean/`: `internal/cli/update/deploy/deploy.go:106-115`.
