# Design — SPEC-UPDATE-BACKUP-RETENTION-001

Design intent for the run phase. Names below describe roles; the run phase chooses identifiers. Evidence for every statement about the present tree is in `research.md`.

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
2. Census. Walk every directory of every run. For each regular file, skip `.template-defaults/` content and recognized bookkeeping (D4); hash the bytes by streaming; record `key = (path inside the backup directory, SHA-256)`. Window runs contribute to the set `W` of window keys. Excess runs contribute, per key, the list of excess runs that hold it. A non-regular entry or an unreadable file marks its run `unclassifiable` and is never opened.
3. Candidates. For each path, take the keys that have at least one excess holder and are not in `W` (REQ-UBR-004: a window-held version is not a past version). Rank them by recency, where a key's recency is its newest excess holder run (REQ-UBR-003). The first N per path are the protected keys.
4. Cover. Choose excess runs so that every protected key has a retained holder. Unclassifiable runs are retained first (REQ-UBR-008) and count as holders. For the remaining uncovered keys, pick greedily the run covering the most uncovered keys, ties to the newest run. A final pass drops any chosen run whose protected keys are all held by other retained runs (REQ-UBR-005 irredundancy).
5. Removal. Every excess run not retained is removed whole, all its directories in both roots. A removal failure turns that run into "retained, removal failed" in the report.

Cost: one streaming read of each regular file; memory proportional to the number of distinct keys (1951 on the 519 MB clone, E6); the cover step is polynomial in the number of protected keys and runs.

The cover is deliberately not "newest holder per key": that model (the t1687 audit's estimate of 30 of 42 retained on the namespace root) keeps one run per key even when one run holds many keys. The greedy cover plus the irredundancy pass retains no run that another retained run makes unnecessary. It is not guaranteed to be a minimum cover (minimum set cover is NP-hard); REQ-UBR-005 asks for irredundancy, which the pass guarantees, and AC-UBR-06(b) checks coverage and irredundancy with an independent brute-force oracle rather than comparing to a minimum.

## D3 Options considered

| Decision | Options | Chosen in this draft | Why |
|---|---|---|---|
| Cap composition (Q2) | A soft cap: window plus irredundant protected runs, reported. B hard cap: protected versions carried forward into a retained location. C hard cap with a limit on retained excess runs. | A | A keeps every existing backup directory intact and adds no write into backups; B writes new content into backups and needs its own registry rows, provenance and restore story; C discards data the operator ruled to protect. The operator may overrule (Q2); the selection step is the same, only step 5 and the report differ. |
| Version key (Q5) | Path inside the backup; normalized project path | Path inside the backup | Restore resolves by that path; normalizing could let a `pre-clean/` copy stand in for a namespace copy the restore would not read. |
| Cover algorithm | Newest holder per key; greedy cover plus irredundancy pass; exact minimum cover | Greedy plus irredundancy | Newest-holder retains avoidable runs; exact minimum is NP-hard and unnecessary for the operator's goal. |
| Bookkeeping identification (F18) | Name at the directory root; name plus the shape the writer produced; cross-check against `backup_metadata.json` `BackedUpItems` | Name plus writer shape | See D4. |
| Symbolic `.agents` (Q6) | Refuse the pass; refuse only link and archive steps | Refuse the pass | Matches the three existing roots; one refusal path. |
| Moves (F24, F25) | Path-based rename with re-checks; project-root-bound handle | Project-root-bound handle | Re-checking narrows the race window and leaves it open; the handle removes it for the escape case. |
| Budget below 1 | Disable protection; use the default | Use the default | Disabling would contradict REQ-UBR-006's floor and the card's completion text. |

## D4 Bookkeeping recognition (F18)

A file at the root of a backup directory is bookkeeping when its name is one of `.moai-run`, `.complete`, `backup_metadata.json` and its content has the shape the writer produces: the run marker is one line that parses as a run id (`run_marker.go`); `.complete` begins with the `stamp=` or the `inventory=` key lines the namespace writer and the preserve-inventory writer emit (`update_namespace_protect.go:238-246`, `update_preserve_inventory.go:352-357`); `backup_metadata.json` decodes as the metadata record with its `BackupType` set (`backup.go:136`, `:198-208`). Anything else with such a name is user content and enters the census. Residual: a user file that imitates a writer's shape is treated as bookkeeping; no writer produces such a file and the hazard is not found in the audits.

A config backup copies `.moai/config/` content to the backup directory's top level, so a user file named `backup_metadata.json` there is overwritten by the writer's own metadata before it can be censused; that is a writer-side collision outside this SPEC (`spec.md` § C, writers unchanged) and is recorded as a Gap in `research.md`.

## D5 Skill-folder residue

- Roots. The refusal list gains `.agents` and `.agents/skills` (REQ-UBR-016, Q6 default). The list is checked once per pass with `Lstat`, as today.
- Archive (REQ-UBR-015). Destination `.moai/archive/skills/retired-<stamp>/<name>`. Each component of the destination path that exists is `Lstat`-checked and must be a real directory; a symbolic component refuses the move and the report names it. A free destination name is chosen with `-N` suffixes if the stamp directory already holds `<name>`.
- Link removal (REQ-UBR-017). After a successful archive, `.agents/skills/<name>` is removed only when `Lstat` shows a symbolic link and its target, resolved against the link's own directory (now known to be a real directory inside the project), equals the folder's former location.
- Bound handle (REQ-UBR-018). The pass opens a handle on the project root and performs `Lstat`, `MkdirAll`, `Rename` and `Remove` through it, so a component swapped for a symbolic link after the check cannot lead outside the project; the per-component check above is the policy layer, the handle is the defense against a swap between check and use. A temporary name for the case-insensitive `skill.md` rename is reserved by exclusive creation, and the final step links the temporary name to `SKILL.md` (which fails if `SKILL.md` exists) and then removes the temporary name. If the filesystem refuses the link, the step restores the content under its own name and reports; it never overwrites.
- Registry. The single move site stays the one registered call (`moveOne`), and every new removal or move call goes through a named function with its own registry row; the pin test counts references so an alias is visible (the F22 and F35 lesson).

Consequence recorded for users: a user who links `.agents/skills` on purpose gets one `refused-symlinked-root` row and no skill-folder cleanup (Q6 Default).

## D6 Configuration and template surface

`state.backup_past_versions_keep` (integer, default 3) in the embedded `state.yaml`, in the state configuration struct, and in the shipped-key inventory used by `TestShippedConfigKeysHaveReaders`. The default is a single exported constant in `internal/config/defaults.go` (the F19 (i) fix); the backup package receives the budget as a parameter and defines no constant of its own. The callers read the configured value, apply the below-1 rule once, and pass it down. Template-First: edit the template, run `make build`.

## D7 Reporting

Rows go through the update ledger as reference rows, in the same style as the skill-residue rows. Indicative content (tests assert the tokens, not the exact prose): one row per removed run with its path; one row per retained excess run with its path and the count of protected versions it holds, or "unclassifiable entry: <relative path>", or "removal failed: <error>"; a summary row with window size, removed count and retained-excess count; one row for the unmanaged entries under `.moai-backups` with count and total size. On a refused root the existing "skipped old backup pruning" row stays.

## D8 Test strategy

- RED first: the probe tests P1 to P4b (Appendix A of `research.md`, adapted to the new function signatures) are committed failing, each before its fix commit.
- Selection step: table tests including the F13 shape (a current version also held by excess runs), the F14 shape (repeated copies), a path with more than N versions, an unclassifiable run, a run that is the only holder of two protected keys of different paths (cover case), and a redundant retained run (irredundancy pass).
- Oracle for AC-UBR-06(b): a brute-force checker that, from the pre-prune census, verifies every protected key has a retained holder and that removing any retained excess run would uncover a protected key.
- Skill residue: P4a and P4b fixtures, the F24 swap fixture and the F25 pre-create fixture reuse the `Shipped` callback and the `moveFile` test seam that the present tests use.
- Platform: permission-bit and FIFO tests are guarded as `requireUnprivileged` and the unix build tag already are.

## D9 Failure modes

| Failure | Behavior |
|---|---|
| Hashing fails for one file | The run is `unclassifiable` and retained; reported. |
| Root becomes a symbolic link between containment check and removal | Out of the single-process threat model (`backup.go` TOCTOU note); unchanged. |
| `os.Root` operation unsupported on the filesystem | The residue step fails closed: content stays, report row. |
| A removal fails midway through a run | The run is reported as retained with the error; the next update re-evaluates. |
| Budget value unreadable or below 1 | The default 3 is used. |
