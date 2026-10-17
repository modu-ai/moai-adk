# Design — SPEC-UPDATE-BACKUP-RETENTION-001

Design intent for the run phase. Names below describe roles; the run phase chooses identifiers. Evidence for every statement about the present tree is in `research.md`.

## D0 Reference behavior

The reference baseline for the repair is the retention behavior at `8e43f064a^` (the deleted `internal/cli/update/backup/retention.go`): on the 10-14 mo.ai.kr snapshot it keeps all 79 namespace runs and loses 0 of 1951 distinct file versions (`research.md` baseline table; reproduction report `/private/tmp/t1687-repro-report.md`). The defect on main `ec5a4f392` is pre-existing, so the repair is judged against that reference, not against "main". The reference keeps everything; the repair must keep the protected set and still remove runs that protect nothing (the operator's budget ruling a263 C), so on the real roots the repaired result lies between the reference (nothing removed) and main (1700 versions lost): zero protected versions lost, and some excess runs removed. The reference code is read from git history for comparison only; it is not restored (it carried F1, F13, F14, F15).

## D1 Shape

```
CleanupOldBackups call sites (clean-install, template-sync)
        |
        v
  root containment check  --refused-->  report refusal, prune nothing   (REQ-UBR-013)
        |
        v
  run collection (unchanged)  ->  runs, newest first                      (REQ-UBR-001)
        |
        v
  window = first keepCount runs ; excess = the rest
        |
        v
  census  (one read of every regular file; key = path-in-backup + SHA-256)  (REQ-UBR-002)
        |
        v
  selection  (pure function of the census, keepCount, N)                   (REQ-UBR-003..006, 008)
        |
        v
  removal of non-selected excess runs ; report rows                        (REQ-UBR-011)
```

The selection step takes no filesystem handle, so the F13 and F14 shapes are table-testable; the census and the removal are the only filesystem steps. This is a structural choice for testability, not a requirement.

## D2 Retention algorithm

Inputs: the ordered runs, keepCount, budget N.

1. keepCount 0: remove every run (explicit wipe). keepCount negative: remove nothing. `len(runs) <= keepCount`: nothing to do.
2. Census. Walk every directory of every run. For each regular file, skip `.template-defaults/` content and recognized bookkeeping (D4); hash the bytes by streaming; record `key = (path inside the backup directory, SHA-256)`. Window runs contribute to the set `W` of window keys. Excess runs contribute, per key, the list of excess runs that hold it. A non-regular entry or an unreadable file marks its run `exempt` (REQ-UBR-008) and is never opened; the run's readable keys are still recorded.
3. Candidates. For each path, take the keys that have at least one excess holder and are not in `W` (REQ-UBR-004). Rank them by recency, where a key's recency is its newest excess holder run (REQ-UBR-003). The first N per path are the protected keys.
4. Cover. Exempt runs are retained first and unconditionally; they count as holders of every key they hold (REQ-UBR-008). For the remaining uncovered protected keys, pick greedily the non-exempt run covering the most uncovered keys, ties to the newest run. A final irredundancy pass drops any *chosen non-exempt* run whose protected keys are all held by other retained runs, exempt runs included as holders. The pass never drops an exempt run (this is the single disposition of audit D2: a FIFO-only run has no protected key, is retained, and the pass does not re-examine it).
5. Removal. Every excess run not retained is removed whole, all its directories in both roots. A removal failure turns that run into "retained, removal failed" in the report; such a run joins the exempt set for the oracle of AC-UBR-06.

Cost: one streaming read of each regular file; memory proportional to the number of distinct keys (1951 on the 519 MB clone, E6); the cover step is polynomial in the number of protected keys and runs.

Bound used by REQ-UBR-010 and AC-UBR-06(b): every retained non-exempt excess run is, by irredundancy, the sole retained holder of at least one protected key, and distinct runs have distinct such keys, so the number of non-exempt retained excess runs is at most the number of distinct protected keys. That bound is checkable without knowing a minimum cover, which is NP-hard and not required: irredundant covers of one key set can differ in size (keys {a,b,c}, runs {abc}, {a}, {b}, {c} admits irredundant covers of size 1 and of size 3), so the criterion asserts coverage, irredundancy and the bound, not equality with a particular cover (audit D3).

The cover is deliberately not "newest holder per key": that model (the t1687 audit's estimate of 30 of 42 retained on the namespace root) keeps one run per key even when one run holds many keys.

Retained-excess figure (audit D5): the oracle computes the count from the pre-prune census alone, so it needs no implementation and can be measured at any time on a clone. It is not needed for a decision (Q2 is ruled); AC-UBR-06 prints it as a measured output.

## D3 Options considered

| Decision | Options | Chosen | Why |
|---|---|---|---|
| Cap composition (Q2) | A soft cap: window plus irredundant protected runs, reported. B hard cap with carried-forward versions. C hard cap with a limit on retained excess runs. | A (operator ruling 2026-10-16) | A keeps every existing backup directory intact and adds no write into backups; B writes new content into backups; C discards data the operator ruled to protect. |
| Version key (Q5) | Path inside the backup; normalized project path | Path inside the backup | Restore resolves by that path; normalizing could let a `pre-clean/` copy stand in for a namespace copy the restore would not read. |
| Cover algorithm | Newest holder per key; greedy cover plus irredundancy pass; exact minimum cover | Greedy plus irredundancy | Newest-holder retains avoidable runs; exact minimum is NP-hard and unnecessary. |
| Bookkeeping identification (F18) | Name at the directory root; name plus the shape the writer produced; backup kind from `backup_metadata.json` | Name plus writer shape | See D4. |
| Symbolic `.agents` (Q6) | Refuse the pass; keep report-only | Keep report-only (closed by the Q3 ruling) | A refusal removes nothing but changes a shipped command's behavior the operator did not ask for. |
| Skill-folder moves (Q3) | Reintroduce archive and link removal; report-only | Report-only (operator ruling) | The 3.2 principle: add no data-deleting behavior; authority basis deferred to card 3.3. |
| Held-back rename (F24, F25) | Path-based rename with re-checks; project-root-bound handle | Project-root-bound handle | Re-checking narrows the race window and leaves it open; the handle removes it for the escape case. |
| Budget below 1 | Disable protection; use the default | Use the default | Disabling would contradict REQ-UBR-006's floor. |

## D4 Bookkeeping recognition (F18)

A file at the root of a backup directory is bookkeeping when its name is one of `.moai-run`, `.complete`, `backup_metadata.json` and its content has the shape the writer produces: the run marker is one line that parses as a run id (`run_marker.go`); `.complete` begins with the `stamp=` or the `inventory=` key lines the namespace writer and the preserve-inventory writer emit (`update_namespace_protect.go:238-246`, `update_preserve_inventory.go:352-357`); `backup_metadata.json` decodes as the metadata record with its `BackupType` set (`backup.go:136`, `:198-208`). Anything else with such a name is user content and enters the census.

Residual (audit D8, deferred, optional): a user file that imitates a writer's shape (a `.moai/config/.complete` whose first line is `inventory=<n>`; config backups copy `.moai/config/` to the directory top level, `backup.go:94-104`) is treated as bookkeeping and not censused, so its last copy is not protected. The author's judgment: provenance (backup kind and location) would close it but is a design choice larger than this card; it is recorded as risk 7 and left to the operator. No fixture asserts the unsafe outcome.

A config backup also overwrites a user file named `backup_metadata.json` with the writer's own metadata before it can be censused; that is a writer-side collision outside this SPEC (`spec.md` § C) and a Gap in `research.md`.

## D5 Skill-folder residue (report-only)

- Retired folders and links are reported, never moved or removed (REQ-UBR-015, REQ-UBR-016). The pass keeps its present shape: `reported-retired-skill` and `reported-retired-link` rows.
- Roots. The pass keeps refusing a symbolic `.moai`, `.claude` or `.claude/skills` as today; `.agents` and `.agents/skills` are not added (Q6 closed). Content reached through a symbolic `.agents/skills` or `.moai/archive` is never written because no write path exists.
- Held-back rename (REQ-UBR-017). The lowercase `skill.md` to `SKILL.md` rename is the one move that remains. It runs through a handle bound to the project root (`os.Root`: `Lstat`, `Rename`), so a component swapped for a symbolic link after the check cannot lead outside the project. A temporary name is reserved by exclusive creation, and the final step links the temporary name to `SKILL.md` (which fails if `SKILL.md` exists) and then removes the temporary name — the removal is of the pass's own temporary name, not user content. If the filesystem refuses the link, the step restores the content under its own name and reports; it never overwrites.
- Registry. The single move site stays one registered call and every changed move or removal call goes through a named function with its own registry row; the pin test counts references so an alias is visible (the F22 and F35 lesson).

## D6 State residue (report-only for databases)

The state-residue pass keeps its walk; for a zero-byte `*.db` with no `-wal` / `-shm` sidecar that is older than the minimum age, it stops calling remove and appends a report row "empty DB file present: <path>" (REQ-UBR-018). The zero-byte `backlog.json.lock` removal is untouched. Whether the doctor command lists the same finding was not verified and is not required here.

## D7 Configuration and template surface

`state.backup_past_versions_keep` (integer, default 3) in the embedded `state.yaml`, in the state configuration struct, and in the shipped-key inventory used by `TestShippedConfigKeysHaveReaders`. The default is a single exported constant in `internal/config/defaults.go` (the F19 (i) fix); the backup package receives the budget as a parameter and defines no constant of its own. The callers read the configured value, apply the below-1 rule once, and pass it down. Template-First: edit the template, run `make build`.

## D8 Reporting

Rows go through the update ledger as reference rows, in the same style as the skill-residue rows. Indicative content (tests assert the tokens, not the exact prose): one row per removed run with its path; one row per retained excess run with its path and the count of protected versions it holds, or "unclassifiable entry: <relative path>", or "removal failed: <error>"; a summary row with window size, removed count and retained-excess count; one row for the unmanaged entries under `.moai-backups` with count and total size. On a refused root the existing "skipped old backup pruning" row stays.

## D9 Test strategy

- RED first: the probe tests P1 to P3 (Appendix A of `research.md`, adapted to the new function signatures, with assertions added) are committed failing, each before its fix commit; P4a and P4b are committed as regression guards that pass on the starting tree.
- Selection step: table tests including the F13 shape (a current version also held by excess runs), the F14 shape (repeated copies), a path with more than N versions, an exempt run (FIFO-only, no protected key), a run that is the only holder of two protected keys of different paths (cover case), a redundant retained run (irredundancy pass), and the {abc},{a},{b},{c} shape (two different irredundant covers must both pass the oracle).
- Oracle for AC-UBR-06(b): a brute-force checker that, from the pre-prune census, verifies every protected key has a retained holder, that removing any retained non-exempt excess run would uncover a protected key, and that the retained non-exempt count is at most the protected-key count. It does not compare to a particular cover.
- Skill residue: P4a and P4b fixtures, the F24 swap fixture and the F25 pre-create fixture reuse the `Shipped` callback and the `moveFile` test seam that the present tests use.
- Platform: permission-bit and FIFO tests are guarded as `requireUnprivileged` and the unix build tag already are.

## D10 Failure modes

| Failure | Behavior |
|---|---|
| Hashing fails for one file | The run is exempt and retained; reported. |
| Root becomes a symbolic link between containment check and removal | Out of the single-process threat model (`backup.go` TOCTOU note); unchanged. |
| `os.Root` operation unsupported on the filesystem | The rename fails closed: content stays, report row. |
| A removal fails midway through a run | The run is reported as retained with the error; the next update re-evaluates. |
| Budget value unreadable or below 1 | The default 3 is used. |

## D11 Post-landing real-world verification (operator act, not a lane act)

`moai update` is forbidden in this repository until card t1733 lands, and no lane runs it. After landing, the operator verifies the repair on a real project with these steps. They are an operator procedure; no lane, auditor or agent performs them.

1. Back up the project's backup roots before anything else: copy `.moai-backups` and `.moai/backups` with modification times preserved (macOS APFS: `cp -Rcp`; elsewhere `cp -a`) to a directory outside the project, and note the directory counts and a recursive file count of each root.
2. Make a second copy of the whole project (a throwaway copy) and run `moai update` there, never in the original project.
3. In the copy, list `.moai-backups` and `.moai/backups` again. Compare with the AC-UBR-06 figures recorded in progress.md §E.2: the config root keeps its window of five runs plus any reported retained excess; the namespace root is no longer emptied (main empties it, 79 to 0); the update summary names every removed run and every retained excess run with its reason.
4. Confirm at least one pre-clean user-file copy remains in the copy's backup roots (a file under a `pre-clean/` directory whose bytes match a user file the template does not ship).
5. Confirm the retained-excess count reported by the update equals the number of run directories present beyond the window in the copy, and does not exceed the number of distinct protected versions.
6. If any step fails, keep the original project untouched and the step-1 backup, and report the numbers to the card.

The procedure uses no machine-local path of this card; the 10-14 snapshot figures (79 to 0, 1700 of 1951) are the baseline on main for comparison.
