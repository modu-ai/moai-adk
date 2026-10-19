# acceptance.md — SPEC-LESSONS-INBOX-SAFE-CLEANUP-001

Verification layer. Requirement obligations (GEARS) live in `spec.md` §2; this file owns the observable, binary-testable criteria in Given-When-Then form, the evidence ledger, and the quality gates. Criterion ids `AC-LIS-NNN`; requirement ids `REQ-LIS-NNN`.

## §A Discipline

- Classes: **RB** release-blocking, **RG** regression-guard (green today by design; never recorded as an adopted pass from a RED).
- Every RB criterion has a RED-now cell (evidence-ledger id, single-invocation read-only command, verbatim stdout, exit code recorded separately, tree pin) and a green path cell naming the milestone that flips it (verification-completeness §2). The RED must be red for the stated reason.
- **Document-level tree pin: `8249bff8c3`** (the base; `git rev-parse --short HEAD` at measurement, ledger EL-B). Source files under test are unchanged between the pin and the measurement; the only untracked files are under this SPEC directory (`repro/`). That pin binds every criterion carrying no pin of its own.
- RED kinds: **behavioral** (a synthetic fixture shows the defect) and **absent-component** (the component the criterion names does not exist yet; the cell shows the absence and the green path flips it by creating the component). Absent-component RED is lexical by nature; each such criterion therefore also names mutants that a lexical check could not catch.
- Exit codes of cells that wrap a command in `repro/exit-of.sh` are the wrapper's final `exit=N` line (the single-invocation form forbids `; echo $?`); `git grep` no-match is documented exit 1. `repro/conditions.sh` always exits 0 when the harness completed; the observation lines are the verdict.
- Failure and race injection uses the named seams of `design.md` §2.6 (`LSEL_HOOK_DIR`), never a patch of an implementation detail. Every GREEN cell that uses a seam asserts `hook_reached=true` (written by the consumer into `hook-trace.log`), so it cannot pass if the seam was not reached. The legacy shim fixtures 6-9 are RED observations of the legacy code only.
- Fixture root: `mktemp -d` (honours `TMPDIR`); measurements here ran with the plain form and, for the first pass, with `TMPDIR` set to the session scratch directory; outputs are path-free and identical.
- Go-test green cells require a non-empty swept set: the output must show at least the stated number of `--- PASS` lines and must not show `[no tests to run]` or `[no test files]`.
- **Closing check names the closing commit, not the base SHA.** At run-phase close each RED cell is re-run at the closing commit (recorded in `progress.md` §E.3 as `closing_commit`) and must show the GREEN output. The base-SHA form in §C is only the labelled RED control.
- Mutant probe: each RB criterion names at least one mutant that satisfies a weaker form of the criterion while violating the requirement; run-phase tests include that mutant as a negative case (the criterion must reject it).

## §B Criteria

### Consumer (M1, M2)

**AC-LIS-001 — Offset beyond the file is a named state, nothing consumed, nothing reset** (REQ-LIS-001, -002) RB
- Given a checkpoint offset of 5 and an inbox of 3 complete lines (fixture 2), When the consumer runs, then two lines are appended and it runs again (fixture 3), then two more and again (fixture 4), Then every run reports `OFFSET_BEYOND_FILE`, the checkpoint file is byte-unchanged, no clusters file is written, and lines 4-5 are not silently skipped by a later run. The blocked state is persistent: in fixture 4 the inbox (7 lines) has grown past the stored offset (5) and the state is still `OFFSET_BEYOND_FILE`; growth never clears it, only an explicit verified bind does (AC-LIS-020). A legacy checkpoint is adopted only if it was never observed blocked, so the real checkpoint (beyond the file today, EL-R2 context) cannot be adopted when the inbox later passes 23339 rows.
- RED-now: EL-C02, EL-C03, EL-C04. Red because: the run reports a plain no-op (`status_state=absent`, `read=0`), writes a clusters file that claims `clusters_offset_after=5`, and in fixture 4 lines 4-5 are lost for good (`lines_4_5_key_in_candidates=false`).
- GREEN (M1): `bash .moai/specs/SPEC-LESSONS-INBOX-SAFE-CLEANUP-001/repro/conditions.sh 2` prints `exit=3`, `clusters=absent`, `checkpoint_offset=5`, `status_state=OFFSET_BEYOND_FILE`; fixtures 3 and 4 print the same state lines (`checkpoint_offset=5`, `status_state=OFFSET_BEYOND_FILE`, `clusters=absent`; fixture 4 prints `inbox_lines=7` and `lines_4_5_key_in_candidates=n/a (no clusters.json)`). Fixture 4 is the sequential persistence test: it runs with the file already longer than the offset.
- Mutants that must be rejected: auto-reset to 0 (`checkpoint_offset=0`), reset to the line count (`checkpoint_offset=3`), the current silent no-op (`status_state=absent`), and **adopt-after-observation** — a consumer that treats a beyond-file checkpoint as a plain legacy checkpoint once the file is long enough (fixture 4 would print `status_state=OK` with `checkpoint_offset=7` and rows 4-5 skipped, i.e. the RED defect itself). A criterion asserting only `read=0` would pass the silent no-op; the state line is therefore asserted.

**AC-LIS-002 — One fixed boundary; GENERATION_CHANGED and NOOP_AT_TAIL are asserted outputs** (REQ-LIS-003, REQ-LIS-001) RB
- Given a consumer that reads the inbox, When a row is appended right after the snapshot copy (fixture 13, through the `after-snapshot-copy` seam), or repeated drains run under a continuous appender (fixture 11), Then the number of rows clustered equals the saved offset and no row is read twice or skipped: the appended row is not in run 1 and is read exactly once by run 2. When the inbox file is swapped between the copy and the post-check (fixture 16), Then the run reports `GENERATION_CHANGED`, commits nothing, and the checkpoint and clusters stay as before. When a lineage-verified inbox is fully drained and drained again (fixture 17), Then the second run reports `NOOP_AT_TAIL`.
- RED-now: EL-C06 (`read=3`, `checkpoint_offset=2`), EL-C07 (`run2_read=1`, `run2_singletons_discarded=1`: the last row re-read and dropped), EL-C11 (`sum_read=233` against 220 final lines, `invariant_sum_eq_offset_eq_lines=false`; two further runs gave 232 and 233, a later audit run 234, all red). Red because the count and the read are two separate opens. Fixtures 6 and 7 are RED observations of the legacy code only: their `wc` shim cannot fire in a conforming implementation, so they are never a GREEN cell. Fixtures 13, 16 and 17 are RED on this tree for the stated reason that the seam and the states do not exist: EL-C13 `run1_hook_reached=false`, EL-C16 `hook_reached=false` and `status_state=absent`, EL-C17 `run2_status_state=absent`.
- GREEN (M1), non-vacuous by construction: fixture 13 prints `run1_hook_reached=true`, `run1_read=2`, `run1_checkpoint_offset=2`, `run2_read=1`, `run2_checkpoint_offset=3`, `inbox_lines=3` (the third row appeared during the run, because the seam fired, and was consumed exactly once); fixture 11 prints `invariant_sum_eq_offset_eq_lines=true` on three consecutive runs; fixture 16 prints `hook_reached=true`, `status_state=GENERATION_CHANGED`, `checkpoint_offset=3` and `swapped_file_keys_in_candidates=false`; fixture 17 prints `run1_status_state=OK` and `run2_status_state=NOOP_AT_TAIL`.
- Mutants: keep the separate count (offset 2 with read 3; caught by fixture 11 and by the seam fixture 13, whose `run1_read` would differ from `run1_checkpoint_offset`); snapshot but compute the offset from a second count (fixture 11); skip the post-read generation check and consume the torn snapshot (fixture 16 prints `status_state=OK` and `swapped_file_keys_in_candidates=true`); report `NOOP_AT_TAIL` for a mismatching lineage (fixtures 2 and 10 print the mismatch states, never NOOP_AT_TAIL). A consumer that never reaches the seam is rejected by `hook_reached=true`. The consumer-versus-replacer half of the lock contract is exercised by fixture 16 (a swap during the read) and by the Go case in AC-LIS-017 (a consumer snapshot overlapping a real cutover).

**AC-LIS-003 — Clusters and checkpoint commit as one pair; at every crash point the manifest decides and the pair is wholly old or wholly new before anything is consumed** (REQ-LIS-005) RB
- Given a failure injected at the rename boundary (fixture 14) and a SIGKILL of the consumer at each of five named points (fixture 15: `temps-written`, `after-prepared`, `between-renames`, `after-renames`, `after-committed`), Then no zero-byte final file exists, the checkpoint is never ahead of the clusters it describes in any interrupted state, a non-crash failure reports `COMMIT_FAILED` and restores the previous pair, and the next clean run completes or restores the pair per the recovery table of `design.md` §2.4 and ends with a coherent pair at offset 3.
- RED-now: EL-C08 (legacy behavior: `run1_checkpoint=EMPTY-FILE` while `run1_read=3` is recorded in clusters, wrapper exit 0 hiding it) and EL-C14, EL-C15 (the seam is absent: `hook_reached=false` and `crash_hook_reached=false`, so a conforming implementation cannot pass without reaching the seam). Fixtures 8 and 9 inject by matching the legacy `jq` program text and are RED observations only; they are not GREEN cells.
- GREEN (M2): fixture 14 prints `hook_reached=true`, `status_state=COMMIT_FAILED`, `checkpoint_zero_byte=false`, `checkpoint_ahead_of_clusters=false`, `zero_byte_final=false` and, after the failure, `clusters_offset=absent` with `checkpoint_offset=absent` (previous pair restored; no prior pair existed). Fixture 15 prints for each of the five points `crash_hook_reached=true`, `crash_zero_byte_final=false`, `crash_checkpoint_ahead_of_clusters=false`, and after the recovery run `recovered_pair_consistent=true`, `recovered_checkpoint_offset=3`, `recovered_zero_byte_final=false`, `recovered_rows_accounted=true`; at `between-renames` the crash state itself is `crash_clusters_offset=3` with `crash_checkpoint_offset=absent` (the checkpoint behind the clusters, never ahead), and the recovery rolls forward. The five-point matrix passes only if all five lines of each block hold.
- Mutants: write the checkpoint before the clusters (`crash_checkpoint_ahead_of_clusters=true` at `between-renames`); roll back after the first rename without restoring the clusters (the `between-renames` block ends with clusters ahead and checkpoint behind, `recovered_pair_consistent=false`); skip the verification step (a malformed temporary is renamed into place; `crash_zero_byte_final=true` or a failed recovery); redirect straight over the final file (the legacy behavior, caught by EL-C08's shape in fixture 14: `zero_byte_final=true`).

**AC-LIS-004 — An empty or unparseable checkpoint is a named state and destroys nothing** (REQ-LIS-006) RB
- Given staged candidates in `clusters.json` and a checkpoint that became a zero-byte file (fixture 12), When the consumer runs with one new row, Then the state is `CHECKPOINT_CORRUPT`, `clusters.json` is byte-identical to before, and the previous committed checkpoint is available as a restorable copy.
- RED-now: EL-C12 (and EL-C09). Red because: `run2_clusters=EMPTY-FILE`, `clusters_unchanged_across_run2=false` — the staged candidates were truncated (the wrapper's pre-overwrite archive still holds a copy, so the loss is to the active file, not to history).
- GREEN (M1): fixture 12 prints `run2_status_state=CHECKPOINT_CORRUPT`, `clusters_unchanged_across_run2=true`, and `clusters_bytes_after_run2` equal to the byte count of the run-1 clusters file; the unit test `corrupt_checkpoint_json` (garbage JSON instead of empty) prints `ok`.
- Mutant: treat a corrupt checkpoint as offset 0 (re-reads everything, caught by `run2_read` appearing); truncate clusters before parsing (current behavior).

**AC-LIS-005 — Lock contention is a recorded state; a dead owner is reaped, a live owner is not** (REQ-LIS-001, -007) RB
- Given a fresh drain lock directory (fixture 5), When the consumer runs, Then the run is skipped with `LOCK_HELD` in the status record, the checkpoint and clusters are unchanged; Given a lock whose recorded owner process is dead, the consumer reaps it and drains; Given a lock whose owner is a live process, the consumer never reaps it however old the lock is.
- RED-now: EL-C05. Red because: the skip leaves only a stderr line and `status_state=absent`; staleness is by age alone (the written contract of `session_drain.sh` reaps at 120 s regardless of the owner).
- GREEN (M2): fixture 5 prints `status_state=LOCK_HELD`, `checkpoint_offset=0`, `clusters=absent`, `exit=0`; test cases `lock_dead_owner_reaped` (expects `read=3`) and `lock_live_owner_kept` (the owner is a running `sleep`; expects `LOCK_HELD`) print `ok`.
- Mutant: reap by age only (kills the live-owner case); never reap (kills the dead-owner case).

**AC-LIS-006 — A replaced file of at least the checkpoint length is detected as a different lineage; a legacy checkpoint is adopted once, flagged** (REQ-LIS-004) RB
- Given a checkpoint taken against file A and the inbox replaced by a different file B with at least as many lines (fixture 10), Then the consumer reports `LINEAGE_MISMATCH` and consumes nothing; Given a legacy offset-only checkpoint within the file, the consumer adopts it once, records `checkpoint_unverified: true` in the status record, and binds the lineage.
- RED-now: EL-C10. Red because: `run2_read=2` from the new file, `run2_candidate_keys` holds the new file's key, `run2_lineage_check=none-performed`.
- GREEN (M1): fixture 10 prints `run2_status_state=LINEAGE_MISMATCH`, `run2_checkpoint_offset=3`, and `run2_candidate_keys` still equals run 1's (the new file's key `tool_failure:Write:PermissionDenied` absent); the test `legacy_v1_adopt` prints `ok` and the status record contains `"checkpoint_unverified":true` once, then `false`; the test `no_adopt_after_blocked` (a checkpoint recorded OFFSET_BEYOND_FILE, then the file grown past the offset, then a run) prints `ok` with the state still blocked and the checkpoint unchanged.
- Mutants: compare only line counts (fixture 10 alone does not kill it when B is longer, so the test `lineage_window_swap` uses a file B that shares A's first line and length but differs in the last window and must report the mismatch); compare only the first line (killed by that same case, because B shares A's first line).

### Identity and shared lock (M3)

**AC-LIS-007 — New rows carry a stable id; existing rows are never rewritten** (REQ-LIS-008) RB
- Given N concurrent producers, When each appends, Then every new row has a non-empty `id`, all ids are distinct, legacy rows are byte-identical, and the legacy readers (`jq` consumer, `InboxStubVersion`, the `moai inbox status` line count) are unaffected.
- RED-now: EL-R1 (no `json:"id"` in `failure_observer.go`, exit 1); absent-component.
- GREEN (M3): `go test -v -run '^TestInboxSafe_ProducerID$' ./internal/hook/...` shows at least 3 `--- PASS` (unique under 8 concurrent producers; legacy bytes unchanged; reader tolerance) and `ok`.
- Mutant: an id derived from the timestamp alone (collides within a second; the 8-producer case fails); rewriting the whole file to add ids to old rows (the legacy-bytes case fails).

**AC-LIS-008 — Every legacy row has a deterministic identity; same-content rows stay distinct** (REQ-LIS-009) RB
- Given a frozen file with byte-identical rows at different positions, When the manifest is built twice, Then identities are equal across builds, pairwise distinct across the identical rows, and the row count equals the line count.
- RED-now: EL-D11 (`ls internal/inboxsafe` exits 1: the package that builds the manifest does not exist; absent-component, and it flips when M3 creates the package). The real-data counts of EL-R2 (`with_id` 0, `distinct_row_bytes` below `rows`: byte-identical duplicate rows exist and nothing distinguishes them) are context for why the criterion matters, not a cell: old rows are immutable by design (REQ-LIS-008), so that observation never flips and is not used as a RED.
- GREEN (M3): `go test -v -run '^TestInboxSafe_Manifest$' ./internal/inboxsafe/...` shows at least 3 `--- PASS`.
- Mutant: identity = row hash only (merges the duplicates; the distinctness case fails).

**AC-LIS-009 — The producer loses no event and does not block while an exclusive replacer holds the lock** (REQ-LIS-010) RB
- Given an exclusive holder, When producers append, Then each producer returns within the fixed bound, the events are held in a spool, and after the holder releases every event is present exactly once. Given a producer that waits for the lock and is released by a gate in the lock helper exactly between the replacer's final spool merge and its release (the deterministic race), Then its event is written to a fresh spool after the atomic close, is not stranded, and is merged exactly once by the next lock holder. Given the lock and spool paths the code uses, Then the repository `.gitignore` and the distributed template `.gitignore` both ignore them.
- RED-now: EL-R3 (no lock token in the producer path, exit 1) and EL-D10 (a spool or lock placed beside the inbox is matched by no ignore rule: `git check-ignore -v` exits 1); absent-component for the lock/spool, behavioral for the ignore hazard.
- GREEN (M3): `go test -v -run '^TestInboxSafe_ProducerSpool$' ./internal/inboxsafe/...` shows at least 5 `--- PASS` (no loss, bound honoured, exactly once, after-final-merge handoff through the gate, closed-spool recovery after a crash between close and merge); `go test -v -run '^TestTemplateGitignoreIgnoresLocalArtifacts$' ./internal/template/...` shows `--- PASS` with the two new rows (path constants of `internal/inboxsafe`) present in its table, and `git check-ignore -v` on both paths exits 0.
- Mutants: producer blocks until release (the bound case fails); producer drops on contention (the no-loss case fails); spool merged twice (the exactly-once case fails); replacer releases without a final merge or closes the spool before the last merge (the after-final-merge case fails: the event is stranded or duplicated); lock/spool constants pointing beside the inbox (the extended gitignore test fails).

**AC-LIS-010 — Rotation runs under the same exclusive lock with unchanged policy** (REQ-LIS-011) RB
- Given appenders racing a rotation, When the chain rotates, Then no appended row is lost across the rename, and the generation count and size cap are as before.
- RED-now: EL-R4 (`os.Rename(inboxPath, …)` at `inbox_lifecycle.go:136`, no lock around it, exit 0 from the grep); behavioral for the race is covered by the existing `TestInboxCap_ConcurrentAppendsCrossCapBoundary`, which asserts no panic, generation count and size but not row preservation (that is the gap).
- GREEN (M3): `go test -v -run '^(TestInboxSafe_RotationLock|TestInboxCap_ConcurrentAppendsCrossCapBoundary)$' ./internal/hook/...` shows at least 4 `--- PASS`, including the new multiset-preservation case and the unchanged existing cases.
- Mutant: lock acquired after the rename (the race case fails); constants changed (the existing cap cases fail).

### Analysis and backup (M4)

**AC-LIS-011 — Classification covers every row once, with evidence, and never bulk-labels or decides removal** (REQ-LIS-012) RB
- Given a frozen synthetic inbox with all eight observation labels, When the classifier runs, Then each row has exactly one class from the closed set, counts sum to the row count, every class carries frequency and evidence basis, truncated summaries are flagged, the UnknownFailure rows are not all in one non-evidence class, and the output has a disposition list and no removal field.
- RED-now: EL-R5 (no classification code in `internal/cli/inbox.go`, exit 1); absent-component. Real label distribution for context: EL-L.
- GREEN (M4): `go test -v -run '^TestInboxSafe_Classify$' ./internal/inboxsafe/...` shows at least 4 `--- PASS`.
- Mutant: every UnknownFailure row to `noise` (the evidence-basis case fails); a `removable` field in the output (the schema case fails).

**AC-LIS-012 — The frozen backup is byte-identical, hashed, read back, owner-only; live files untouched** (REQ-LIS-013) RB
- Given a live synthetic inbox that grows during the copy, When the backup runs, Then the copy is a complete-line prefix, its sha256 equals the recorded hash after a read-back, permissions are owner-only, the live files' prefix hashes equal the backup hash, and the live checkpoint is byte-unchanged.
- RED-now: EL-R6 (no `backup` verb, exit 1); absent-component.
- GREEN (M4): `go test -v -run '^TestInboxSafe_Backup$' ./internal/inboxsafe/...` at least 3 `--- PASS`.
- Mutant: copy that ends mid-line (the complete-line case fails); hash computed before the read-back (the tampered-copy case fails).

**AC-LIS-013 — Inventory equals the frozen copy** (REQ-LIS-014) RB
- Given a frozen copy, When the inventory runs, Then its row count equals the copy's line count and every row has identity, position, hash, and class.
- RED-now: EL-R6 (no `inventory` verb); absent-component.
- GREEN (M4): `go test -v -run '^TestInboxSafe_Inventory$' ./internal/inboxsafe/...` at least 2 `--- PASS`.
- Mutant: an inventory that skips unparseable rows (the count equation fails; unparseable rows must be counted and preserved).

**AC-LIS-020 — The bind primitive accepts explicit verified values only** (REQ-LIS-020) RB
- Given a file and a claimed generation, offset, and window hash, When the bind runs, Then it succeeds only if all three verify, writes atomically, and any call without explicit values or with a reset-style flag is refused.
- RED-now: EL-R6 (no `checkpoint` verb); absent-component.
- GREEN (M4): `go test -v -run '^TestInboxSafe_Bind$' ./internal/cli/...` at least 5 `--- PASS` (verified bind; wrong window refused; missing value refused; `--reset`/`--to-tail` unknown flag refused; a verified bind removes the persistent blocked record in the same atomic step, and nothing else — file growth, a later consumer run — clears it).
- Mutants: a default offset of 0 when none is given (the missing-value case fails); a bind that leaves the blocked record in place (the next consumer run stays blocked after a verified bind); a consumer path that clears the blocked record on its own (the growth case fails).

### Proof, archive, cutover, ledger (M5)

**AC-LIS-014 — Removal eligibility needs a mechanically resolved evidence chain** (REQ-LIS-015) RB
- Given synthetic items in each state and with each evidence link broken in turn (change commit not an ancestor, verification record missing, hash mismatch, approval reference missing), When the proof gate runs, Then only items in `independently_verified`/`archived` with all links resolved are eligible, every other state (blocked, rejected, normal_refusal, environment_pending, verification_failed, reverted, unresolved) is preserved, and a fixture with zero eligible items yields an empty removal set and exit 0.
- RED-now: EL-R6 (no `prove` verb); absent-component.
- GREEN (M5): `go test -v -run '^TestInboxSafe_Prove$' ./internal/inboxsafe/...` at least 12 `--- PASS` (one per state and one per broken link).
- Mutant: a gate that trusts the recorded state string without resolving the references (the broken-link cases fail); a gate that treats a missing evidence file as "no objection" (the missing-record case fails).

**AC-LIS-015 — The archive is opened, counted, and hashed before any cutover** (REQ-LIS-017) RB
- Given a non-empty removal set, When the archive step runs, Then the archive is read back, its row count equals the removal set size, its hash is in the ledger, and the cutover refuses to start without that record.
- RED-now: EL-R6 (no `archive` verb); absent-component.
- GREEN (M5): `go test -v -run '^TestInboxSafe_Archive$' ./internal/inboxsafe/...` at least 3 `--- PASS`.
- Mutant: cutover that does not check for the archive record (the refusal case fails).

**AC-LIS-016 — The multiset equation holds and an imbalance refuses the cutover** (REQ-LIS-018) RB
- Given an inbox with byte-identical duplicate rows, When a cutover is planned, Then rows before plus rows appended equals new active plus archive counting duplicates individually; removing one of two identical rows and keeping both in the archive is refused.
- RED-now: EL-R6 (no `cutover` verb); absent-component.
- GREEN (M5): `go test -v -run '^TestInboxSafe_Equation$' ./internal/inboxsafe/...` at least 4 `--- PASS`.
- Mutant: a set-based check (passes a dropped duplicate; the duplicate case fails).

**AC-LIS-017 — Cutover is atomic, appends survive it, the checkpoint is mapped exactly or preserved** (REQ-LIS-019) RB
- Given producers appending during a cutover, When it completes, Then every appended row is present once in order, the live path never shows a partial file, the verified checkpoint advances by exactly the number of removed rows before it (and the next consumer run neither re-reads nor skips), and a checkpoint in a mismatch state is preserved verbatim with the state still reported.
- RED-now: EL-R6 (no `cutover` verb); absent-component.
- GREEN (M5): `go test -v -run '^TestInboxSafe_Cutover$' ./internal/inboxsafe/...` at least 6 `--- PASS`, including the case that a consumer snapshot overlapping the cutover (triggered through the consumer seam `after-snapshot-copy`) ends `GENERATION_CHANGED` with nothing committed.
- Mutant: remap by percentage or by clamping to the new length (the exact-mapping case fails); cutover that resets a mismatching checkpoint (the preserved-verbatim case fails).

**AC-LIS-018 — Any failure rolls back to the old generation byte-identical** (REQ-LIS-019) RB
- Given a failure injected at each cutover step, When the cutover aborts, Then the live inbox and the checkpoint hash equal their pre-cutover hashes, the archive is retained, and the ledger records the abort.
- RED-now: EL-R6 (no `rollback`); absent-component.
- GREEN (M5): `go test -v -run '^TestInboxSafe_Rollback$' ./internal/inboxsafe/...` at least 6 `--- PASS` (one per injected step).
- Mutant: a rollback that regenerates the old file from the archive plus active instead of restoring the retained original (the byte-identity case fails on ordering).

**AC-LIS-019 — The three forbidden patterns are guarded** (REQ-LIS-022) RB
- Given the tooling, Then (a) no code path truncates the inbox using a checkpoint value, (b) no code path unlinks an inbox row source, the backed-up original, or the archive (the only permitted removal is the superseded temporary file of the same operation), (c) the tooling output never contains a whole-log validity verdict and unparseable rows are counted individually and preserved.
- RED-now: EL-R7 (no `TestInboxSafe_` tests exist, exit 1); absent-component.
- GREEN (M5): `go test -v -run '^TestInboxSafe_Forbidden$' ./internal/inboxsafe/...` at least 3 `--- PASS`: a behavioral case per pattern plus a source scan that fails on `os.Remove`, `os.RemoveAll`, `os.Truncate`, `O_TRUNC`, a shell `>` or `: >` redirect onto the inbox or checkpoint path, and a rename over the original generation outside the cutover, anywhere outside the allowlisted temporary-file cleanup.
- Named mutants (each must fail its case): `truncate-with-old-offset` (an implementation that truncates the inbox to the checkpoint offset), `delete-original` (an implementation that unlinks the backup after a successful cutover), `whole-log-invalid` (a report field `log_valid:false` after one bad row).

**AC-LIS-021 — The ledger is append-only, hash-chained, and gives every removed item a resolvable trail** (REQ-LIS-021) RB
- Given a completed synthetic cleanup, When the ledger is verified, Then each removed item shows item, change commit, verification record, and archive hash, all resolvable, and altering or deleting any earlier row makes verification fail.
- RED-now: EL-R6 (no `ledger` verb); absent-component.
- GREEN (M5): `go test -v -run '^TestInboxSafe_Ledger$' ./internal/inboxsafe/...` at least 4 `--- PASS`.
- Mutant: an unchained ledger (the tamper case fails).

**AC-LIS-024 — A pinned item stays active until its dependencies have landed and been independently verified** (REQ-LIS-016) RB
- Given a synthetic item standing for the PR1841 case (a premerge-review violation resolved only by a guideline card and a code-gate card) with those two dependencies unlanded or unverified, When the proof gate and the cutover run, Then the item is reported `PINNED_UNPROCESSED`, is excluded from the removal set, stays in the active generation, and a ledger entry records the pin with its dependency list. When only one dependency is satisfied, it is still pinned. It becomes eligible only when, for each of t1900 and t1868, all five checks of `design.md` §6 hold: a landing commit carrying the card id is reachable from the cutover base; the verification record names that same commit; its verdict is `pass` (and `verified:true` for an apply-ledger-style row); its verifier differs from the author of the commit and from the card's own self-check; its artifact exists and re-hashes to the recorded value.
- RED-now: EL-R10 (no pin concept in the inbox CLI or package, exit 1); absent-component.
- GREEN (M5): `go test -v -run '^TestInboxSafe_Pin$' ./internal/inboxsafe/...` at least 9 `--- PASS`, one per case: neither landed; one landed and verified, one not landed; both landed, one with no verification record; both landed, one record with a failing verdict (`verified:false`); both landed, one record for a different commit than the landing commit; both landed, one record that is the card's own self-check or names the commit author as verifier; both landed, one record whose artifact hash does not match; both landed and every check passing (eligible); and the pin survives a hand-edited recorded state `independently_verified`.
- **Named mutant `remove-before-t1900-t1868`** and its bypass family, each of which must fail its own case: ignore `requires` entirely; accept landed-but-unverified; accept a failing verdict; accept a record for another commit; accept the card's own self-check as independent verification; accept a record whose artifact hash mismatches; release on a hand-edited state. This is the PR1841 worked example (`design.md` §6).

### Proof runs (M6, M7)

**AC-LIS-022 — Synthetic end-to-end: consumer, proposal, apply, verify, prove, archive, cutover** (REQ-LIS-023) RB
- Given a synthetic git repository, a synthetic inbox with a repeated failure, and a staged proposal and approved decision, When the chain runs through the existing `lsel-apply.sh` and `verify.sh`, Then the apply ledger row and the verified marker exist, the proof gate accepts exactly that item through them, the archive and cutover remove exactly it, an item with a failed verification (forced revert) is preserved, and the provenance ledger shows the full trail.
- RED-now: EL-R8 (no end-to-end script, exit 1); absent-component.
- GREEN (M6): `bash .claude/skills/hns-lsel-curator/e2e_test.sh` prints `ok` for each of at least 5 named steps ending `ALL E2E TESTS PASSED`.
- Mutant: a proof gate that accepts the item with `verified:false` in the apply ledger (the preserved-failed case fails).

**AC-LIS-023 — Real safe path: gated steps, zero removal is a valid result, live files keep their frozen prefix** (REQ-LIS-024) RB
- Given the real inbox and checkpoint, When steps R0-R3b run (and R4-R5's cutover branch only if the removal set is non-empty), Then each step leaves its evidence file and refuses to start without the previous step's ledger row; the PR1841 case is recorded as a pinned item in the provenance ledger (step R3b, before any cutover) citing the re-hashed sha256 of the two case files; and the recorded outcome states the number of removed items (which may be 0) with the evidence behind each preserved state. Verification at R5 depends on the run: **no cutover occurred** (zero removals, the expected outcome) — the first N bytes of the live inbox hash to the frozen copy's recorded hash and the live checkpoint is byte-identical to its frozen copy; **a cutover occurred** — the multiset equation (before plus appended equals new active plus archive plus pending spool) holds, the new generation's sha256 equals the commit manifest's, the archive row count and hash equal the ledger's, and the checkpoint is either exactly mapped or preserved verbatim; the prefix-byte check is not applied after a cutover because removing rows legitimately changes the prefix.
- RED-now: EL-R9 (the cleanup state directory does not exist, exit 1); absent-component. Real-state context: EL-R11 (the primary LSEL state holds no proposals directory and no apply ledger).
- GREEN (M7): `r5-post.md` records, via the tool's `verify` output, `mode: no_cutover` with `prefix_hash_equal: true` and `checkpoint_byte_identical: true`, or `mode: cutover` with `equation_balanced: true`, `generation_hash_matches_manifest: true`, `archive_count_matches_ledger: true`, `checkpoint_mapped_or_preserved: true`; in both modes `ledger_chain: ok`, `pin_recorded: true` (R3b, with the two case-file hashes equal to `spec.md` §1.2) and `removed_items: <n>` equal to the cutover record (or `zero_removal`).
- Ordering witness (verification-claim-integrity §2.3): `.moai/reports/` and `.moai/state/` are gitignored, so the commit graph cannot witness the real-world order of real steps from them. The criterion therefore claims only what can be verified: (1) mechanically, each step refuses to start without the previous step's ledger row with its hash (freeze before inventory, archive before cutover); (2) in the tracked record, after each step the lane commits one `progress.md` §E.2 line (step id, evidence-file sha256, ledger sequence number) in its own commit, and the closing check reads the graph order of those commits (`git log --format=%H -- progress.md`, one commit per step, in step order).
- Mutants: a run that writes the live inbox before the frozen copy is verified (the mechanical refusal fails it: the freeze row is absent); a step record committed together with the next step's record (the graph shows two steps in one commit; the closing check fails); a cutover performed with an empty removal set; applying the prefix-byte check after a non-empty cutover (a correct run would fail it — the criterion forbids this form); a R3b that pins from a case file whose hash differs from `spec.md` §1.2 without recording `evidence_missing`.

## §C Evidence ledger (RED-now observations)

Pin for every entry: tree `8249bff8c3` plus the untracked `repro/` files of this SPEC. Time of measurement: 2026-10-19. Entries that read the real inbox are read-only and the file grows by append, so only the asserted facts (not the exact counts) are stable.

**EL-B** — `git rev-parse --short HEAD` -> `8249bff8c3`, exit 0. `bash .moai/specs/SPEC-LESSONS-INBOX-SAFE-CLEANUP-001/repro/exit-of.sh git merge-base --is-ancestor 7705b365e2 HEAD` -> `exit=0` (the installed build commit is an ancestor of HEAD; it lags). The first command printed `8249bff8c3` then `exit=0` through the same wrapper.

**EL-S** — `shasum -a 256 .claude/skills/hns-lsel-curator/drain.sh .claude/skills/hns-lsel-curator/session_drain.sh internal/hook/failure_observer.go internal/hook/inbox_lifecycle.go`, exit 0:
```
bddacab789ac52fc677c784d6af8464376146c4284354dad0e06f367f81c7a41  .claude/skills/hns-lsel-curator/drain.sh
307beb001cc00826164267a11bd1a47e30773e6dfb6bf57139babe06cad972a3  .claude/skills/hns-lsel-curator/session_drain.sh
6aa8e5c8aa23bfa7ebeb7413931b1cb189074f2ad00f20a5ded941b58b7a86da  internal/hook/failure_observer.go
726a6aeda37a06d21b6564f4db0868c85ccabdf713526e937a20972bffeb1785  internal/hook/inbox_lifecycle.go
```

**EL-C01** (control, regression guard RG-1) — `bash .moai/specs/SPEC-LESSONS-INBOX-SAFE-CLEANUP-001/repro/conditions.sh 1`, exit 0:
```
condition=1
exit=0
read=3
candidates=1
candidate_keys=["tool_failure:Bash:ExitError"]
singletons_discarded=0
clusters_offset_after=3
checkpoint_offset=3
status_state=absent
inbox_lines=3
stderr=drain: read 3 stubs (offset 0→3) — discarded 0 noise + 0 singletons; 1 candidate cluster(s) emitted to clusters.json|
```

**EL-C02** — `... conditions.sh 2`, exit 0:
```
condition=2
exit=0
read=0
candidates=0
candidate_keys=[]
singletons_discarded=0
clusters_offset_after=5
checkpoint_offset=5
status_state=absent
inbox_lines=3
stderr=drain: no-op (offset=5, inbox=3 lines) — nothing new to drain|
```

**EL-C03** — `... conditions.sh 3`, exit 0:
```
condition=3
exit=0
read=0
candidates=0
candidate_keys=[]
singletons_discarded=0
clusters_offset_after=5
checkpoint_offset=5
status_state=absent
inbox_lines=5
stderr=drain: no-op (offset=5, inbox=5 lines) — nothing new to drain|
```

**EL-C04** — `... conditions.sh 4`, exit 0:
```
condition=4
exit=0
read=2
candidates=1
candidate_keys=["tool_failure:Agent:ContextCancelled"]
singletons_discarded=0
clusters_offset_after=7
checkpoint_offset=7
status_state=absent
inbox_lines=7
stderr=drain: read 2 stubs (offset 5→7) — discarded 0 noise + 0 singletons; 1 candidate cluster(s) emitted to clusters.json|
lines_4_5_key_in_candidates=false
```

**EL-C05** — `... conditions.sh 5`, exit 0:
```
condition=5
exit=0
clusters=absent
checkpoint_offset=0
status_state=absent
inbox_lines=3
stderr=session_drain: lock contention — another session holds the drain lock; skipping this drain (no-op)|
```

**EL-C06** — `... conditions.sh 6`, exit 0:
```
condition=6
exit=0
read=3
candidates=1
candidate_keys=["tool_failure:Bash:ExitError"]
singletons_discarded=0
clusters_offset_after=2
checkpoint_offset=2
status_state=absent
inbox_lines=3
stderr=drain: read 3 stubs (offset 0→2) — discarded 0 noise + 0 singletons; 1 candidate cluster(s) emitted to clusters.json|
```

**EL-C07** — `... conditions.sh 7`, exit 0:
```
condition=7
run1_exit=0
run1_read=3
run1_candidates=1
run1_candidate_keys=["tool_failure:Bash:ExitError"]
run1_singletons_discarded=0
run1_clusters_offset_after=2
run1_checkpoint_offset=2
run1_status_state=absent
run1_inbox_lines=3
run1_stderr=drain: read 3 stubs (offset 0→2) — discarded 0 noise + 0 singletons; 1 candidate cluster(s) emitted to clusters.json|
run2_exit=0
run2_read=1
run2_candidates=0
run2_candidate_keys=[]
run2_singletons_discarded=1
run2_clusters_offset_after=3
run2_checkpoint_offset=3
run2_status_state=absent
run2_inbox_lines=3
run2_stderr=drain: read 1 stubs (offset 2→3) — discarded 0 noise + 1 singletons; 0 candidate cluster(s) emitted to clusters.json|
```

**EL-C08** — `... conditions.sh 8`, exit 0:
```
condition=8
run1_exit=0
run1_read=3
run1_candidates=1
run1_candidate_keys=["tool_failure:Bash:ExitError"]
run1_singletons_discarded=0
run1_clusters_offset_after=3
run1_checkpoint=EMPTY-FILE bytes=0
run1_status_state=absent
run1_inbox_lines=3
run1_stderr=session_drain: drain.sh exited non-zero (tooling or inbox problem; see <fx>/state/last-drain.log) — offset NOT advanced — degraded to no-op (fail-open; next session start retrie
run1_checkpoint_bytes=0
```

**EL-C09** — `... conditions.sh 9`, exit 0:
```
condition=9
run1_exit=0
run1_read=3
run1_candidates=1
run1_candidate_keys=["tool_failure:Bash:ExitError"]
run1_singletons_discarded=0
run1_clusters_offset_after=3
run1_checkpoint=EMPTY-FILE bytes=0
run1_status_state=absent
run1_inbox_lines=3
run1_stderr=session_drain: drain.sh exited non-zero (tooling or inbox problem; see <fx>/state/last-drain.log) — offset NOT advanced — degraded to no-op (fail-open; next session start retrie
run1_checkpoint_bytes=0
run2_exit=0
run2_clusters=EMPTY-FILE bytes=0
run2_checkpoint=EMPTY-FILE bytes=0
run2_status_state=absent
run2_inbox_lines=3
run2_stderr=session_drain: drain.sh exited non-zero (tooling or inbox problem; see <fx>/state/last-drain.log) — offset NOT advanced — degraded to no-op (fail-open; next session start retrie
run2_checkpoint_bytes=0
```

**EL-C10** — `... conditions.sh 10`, exit 0:
```
condition=10
run1_exit=0
run1_read=3
run1_candidates=1
run1_candidate_keys=["tool_failure:Bash:ExitError"]
run1_singletons_discarded=0
run1_clusters_offset_after=3
run1_checkpoint_offset=3
run1_status_state=absent
run1_inbox_lines=3
run1_stderr=drain: read 3 stubs (offset 0→3) — discarded 0 noise + 0 singletons; 1 candidate cluster(s) emitted to clusters.json|
run2_exit=0
run2_read=2
run2_candidates=1
run2_candidate_keys=["tool_failure:Write:PermissionDenied"]
run2_singletons_discarded=0
run2_clusters_offset_after=5
run2_checkpoint_offset=5
run2_status_state=absent
run2_inbox_lines=5
run2_stderr=drain: read 2 stubs (offset 3→5) — discarded 0 noise + 0 singletons; 1 candidate cluster(s) emitted to clusters.json|
run2_lineage_check=none-performed
```

**EL-C11** — `... conditions.sh 11`, exit 0 (first of three runs; the other two printed `sum_read=232` and `sum_read=233`, all `invariant ... =false`):
```
condition=11
sum_read=233
final_checkpoint_offset=220
final_inbox_lines=220
invariant_sum_eq_offset_eq_lines=false
status_state=absent
```

**EL-C12** — `... conditions.sh 12`, exit 0:
```
condition=12
run1_exit=0
run1_read=3
run1_candidates=1
run1_candidate_keys=["tool_failure:Bash:ExitError"]
run1_singletons_discarded=0
run1_clusters_offset_after=3
run1_checkpoint_offset=3
run1_status_state=absent
run1_inbox_lines=3
run1_stderr=
run2_exit=0
run2_clusters=EMPTY-FILE bytes=0
run2_checkpoint=EMPTY-FILE bytes=0
run2_status_state=absent
run2_inbox_lines=4
run2_stderr=session_drain: drain.sh exited non-zero (tooling or inbox problem; see <fx>/state/last-drain.log) — offset NOT advanced — degraded to no-op (fail-open; next session start retrie
clusters_unchanged_across_run2=false
clusters_bytes_after_run2=0
```

**EL-C13** (seam race; RED because the seam does not exist) — `bash .moai/specs/SPEC-LESSONS-INBOX-SAFE-CLEANUP-001/repro/conditions.sh 13`, exit 0:
```
condition=13
run1_hook_reached=false
run1_read=2
run1_clusters_offset=2
run1_checkpoint_offset=2
run1_pair_consistent=true
run1_checkpoint_ahead_of_clusters=false
run1_zero_byte_final=false
run2_read=0
run2_clusters_offset=2
run2_checkpoint_offset=2
run2_pair_consistent=true
run2_checkpoint_ahead_of_clusters=false
run2_zero_byte_final=false
run2_status_state=absent
inbox_lines=2
```

**EL-C14** (commit failure through the seam; RED because the seam does not exist) — `... conditions.sh 14`, exit 0:
```
condition=14
hook_reached=false
wrapper_exit=0
clusters_offset=3
checkpoint_offset=3
pair_consistent=true
checkpoint_ahead_of_clusters=false
zero_byte_final=false
status_state=absent
checkpoint_zero_byte=false
```

**EL-C15** (crash matrix; one of five points shown — the other four, `temps-written`, `after-prepared`, `after-renames`, `after-committed`, printed the identical block on this tree) — `... conditions.sh 15 between-renames`, exit 0:
```
condition=15
crash_hook_reached=false
crash_clusters_offset=3
crash_checkpoint_offset=3
crash_pair_consistent=true
crash_checkpoint_ahead_of_clusters=false
crash_zero_byte_final=false
recovered_clusters_offset=3
recovered_checkpoint_offset=3
recovered_pair_consistent=true
recovered_checkpoint_ahead_of_clusters=false
recovered_zero_byte_final=false
recovered_status_state=absent
recovered_rows_accounted=true
```

**EL-C16** (swap during read; RED because neither the seam nor `GENERATION_CHANGED` exists) — `... conditions.sh 16`, exit 0:
```
condition=16
hook_reached=false
status_state=absent
clusters_offset=5
checkpoint_offset=5
pair_consistent=true
checkpoint_ahead_of_clusters=false
zero_byte_final=false
swapped_file_keys_in_candidates=false
```

**EL-C17** (`NOOP_AT_TAIL`; RED because the state does not exist) — `... conditions.sh 17`, exit 0:
```
condition=17
run1_exit=0
run1_read=3
run1_candidates=1
run1_candidate_keys=["tool_failure:Bash:ExitError"]
run1_singletons_discarded=0
run1_clusters_offset_after=3
run1_checkpoint_offset=3
run1_status_state=absent
run1_inbox_lines=3
run1_stderr=drain: read 3 stubs (offset 0→3) — discarded 0 noise + 0 singletons; 1 candidate cluster(s) emitted to clusters.json|
run2_exit=0
run2_read=0
run2_candidates=0
run2_candidate_keys=[]
run2_singletons_discarded=0
run2_clusters_offset_after=3
run2_checkpoint_offset=3
run2_status_state=absent
run2_inbox_lines=3
run2_stderr=drain: no-op (offset=3, inbox=3 lines) — nothing new to drain|
```

**EL-D10** (a lock or spool placed beside the inbox is unignored) — `bash .moai/specs/SPEC-LESSONS-INBOX-SAFE-CLEANUP-001/repro/exit-of.sh git check-ignore -v .moai/lessons-inbox.lock .moai/lessons-inbox.spool.jsonl`: stdout `exit=1` (no ignore rule matches either path). Supporting reads: `git grep -n lessons-inbox -- .gitignore internal/template/templates/.gitignore` printed `.gitignore:311:.moai/lessons-inbox.jsonl` and `internal/template/templates/.gitignore:302:.moai/lessons-inbox.jsonl`; `git check-ignore -v .moai/state/lessons-cleanup/ledger.jsonl` printed `.gitignore:401:.moai/state/`; `git grep -n -E "^\.moai/state/" -- internal/template/templates/.gitignore` printed `242:.moai/state/`.

**EL-D11** (the package that builds the manifest is absent) — `bash .moai/specs/SPEC-LESSONS-INBOX-SAFE-CLEANUP-001/repro/exit-of.sh ls internal/inboxsafe`:
```
ls: internal/inboxsafe: No such file or directory
exit=1
```

**EL-L** — `jq -s -c '{n: length, categories: (group_by(.event_key | split(":") | last) | map({k: (.[0].event_key | split(":") | last), n: length}) | sort_by(-.n)), kinds: (group_by(.event_key | split(":") | first) | map({k: (.[0].event_key | split(":") | first), n: length})), ts_min: (map(.timestamp) | min), ts_max: (map(.timestamp) | max)}' /Users/goos/MoAI/moai-adk-go/.moai/lessons-inbox.jsonl` (read-only), exit 0:
```
{"n":23047,"categories":[{"k":"UnknownFailure","n":13778},{"k":"WorktreeGuardRefusal","n":8435},{"k":"TimeoutError","n":210},{"k":"OOMKilled","n":183},{"k":"ExitError","n":148},{"k":"SandboxViolation","n":130},{"k":"PermissionDenied","n":113},{"k":"ContextCancelled","n":50}],"kinds":[{"k":"tool_failure","n":23047}],"ts_min":"2026-07-10T09:43:18Z","ts_max":"2026-10-19T05:20:37Z"}
```

**EL-R1** — `bash .moai/specs/SPEC-LESSONS-INBOX-SAFE-CLEANUP-001/repro/exit-of.sh git grep --count json:.id. -- internal/hook/failure_observer.go`: stdout `exit=1`.

**EL-R2** — `jq -s -c '{rows: length, with_id: (map(select(has("id"))) | length), distinct_row_bytes: (map(tojson) | unique | length)}' /Users/goos/MoAI/moai-adk-go/.moai/lessons-inbox.jsonl` (read-only), exit 0 (the harness surfaced no error):
```
{"rows":23051,"with_id":0,"distinct_row_bytes":22667}
```

**EL-R3** — `bash .../repro/exit-of.sh git grep --count -E "Flock|LockFileEx|[.]lock" -- internal/hook/failure_observer.go internal/hook/inbox_lifecycle.go`: stdout `exit=1`.

**EL-R4** — `bash .../repro/exit-of.sh git grep -n os.Rename.inboxPath -- internal/hook/inbox_lifecycle.go`:
```
internal/hook/inbox_lifecycle.go:136:	if err := os.Rename(inboxPath, inboxGenPath(inboxPath, 1)); err != nil {
exit=0
```

**EL-R5** — `bash .../repro/exit-of.sh git grep -n -i classif -- internal/cli/inbox.go internal/cli/inbox_*.go internal/inboxsafe`: stdout `exit=1`.

**EL-R6** — `bash .../repro/exit-of.sh git grep --count -E "Use:[[:space:]]+.(backup|inventory|classify|prove|archive|cutover|rollback|ledger|checkpoint|pin)." -- internal/cli/inbox.go internal/cli/inbox_*.go internal/inboxsafe`: stdout `exit=1`.

**EL-R7** — `bash .../repro/exit-of.sh git grep --count TestInboxSafe_ -- internal`: stdout `exit=1`.

**EL-R8** — `bash .../repro/exit-of.sh git ls-files --error-unmatch .claude/skills/hns-lsel-curator/e2e_test.sh`:
```
error: pathspec '.claude/skills/hns-lsel-curator/e2e_test.sh' did not match any file(s) known to git
Did you forget to 'git add'?
exit=1
```

**EL-R9** — `bash .../repro/exit-of.sh ls /Users/goos/MoAI/moai-adk-go/.moai/state/lessons-cleanup` (read-only):
```
ls: /Users/goos/MoAI/moai-adk-go/.moai/state/lessons-cleanup: No such file or directory
exit=1
```

**EL-R10** — `bash .../repro/exit-of.sh git grep --count -i pinned -- internal/cli/inbox.go internal/cli/inbox_*.go internal/inboxsafe`: stdout `exit=1`.

**EL-R11** — `bash .../repro/exit-of.sh find /Users/goos/MoAI/moai-adk-go/.moai/state/lsel -maxdepth 1` (read-only), exit 0 wrapper-recorded:
```
/Users/goos/MoAI/moai-adk-go/.moai/state/lsel
/Users/goos/MoAI/moai-adk-go/.moai/state/lsel/drain-offset.json
/Users/goos/MoAI/moai-adk-go/.moai/state/lsel/last-drain.log
/Users/goos/MoAI/moai-adk-go/.moai/state/lsel/clusters-history
/Users/goos/MoAI/moai-adk-go/.moai/state/lsel/clusters.json
exit=0
```

**EL-P** — `bash .../repro/exit-of.sh git grep -n bypassPermissions -- internal/hook/post_tool_failure.go`:
```
internal/hook/post_tool_failure.go:219:		return "PermissionDenied: Tool was denied permission. Check settings.json permissions or run in bypassPermissions mode."
exit=0
```

**EL-RG** — `bash .../repro/exit-of.sh bash .claude/skills/hns-lsel-curator/drain_test.sh` ended with `ALL DRAIN TESTS PASSED` then `exit=0`; `... session_drain_test.sh` ended with `ALL SESSION DRAIN TESTS PASSED` then `exit=0` (seven `ok` lines each section as in the existing scripts).

## §C.1 Traceability (requirement to criterion)

| Requirement | Criteria |
|-------------|----------|
| REQ-LIS-001 | AC-LIS-001, AC-LIS-005 |
| REQ-LIS-002 | AC-LIS-001 |
| REQ-LIS-003 | AC-LIS-002 |
| REQ-LIS-004 | AC-LIS-006 |
| REQ-LIS-005 | AC-LIS-003 |
| REQ-LIS-006 | AC-LIS-004 |
| REQ-LIS-007 | AC-LIS-005 |
| REQ-LIS-008 | AC-LIS-007 |
| REQ-LIS-009 | AC-LIS-008 |
| REQ-LIS-010 | AC-LIS-009 |
| REQ-LIS-011 | AC-LIS-010 |
| REQ-LIS-012 | AC-LIS-011 |
| REQ-LIS-013 | AC-LIS-012 |
| REQ-LIS-014 | AC-LIS-013 |
| REQ-LIS-015 | AC-LIS-014 |
| REQ-LIS-016 | AC-LIS-024 |
| REQ-LIS-017 | AC-LIS-015 |
| REQ-LIS-018 | AC-LIS-016 |
| REQ-LIS-019 | AC-LIS-017, AC-LIS-018 |
| REQ-LIS-020 | AC-LIS-020 |
| REQ-LIS-021 | AC-LIS-021 |
| REQ-LIS-022 | AC-LIS-019 |
| REQ-LIS-023 | AC-LIS-022 |
| REQ-LIS-024 | AC-LIS-023 |
| REQ-LIS-025 | RG-2 (§D); interface contracts IC-1..IC-4 are exercised by AC-LIS-001 (IC-1), AC-LIS-020 (IC-2), AC-LIS-014 (IC-3), AC-LIS-009 and AC-LIS-010 (IC-4) |

## §D Regression guards (green today by design; not RED-adopted)

- **RG-1** — fixture 1 (EL-C01) keeps `read=3`, `candidates=1`, `checkpoint_offset=3` after every milestone; at M1+ it also prints `status_state=OK`.
- **RG-2** — boundary: `git diff --name-only 8249bff8c3..<closing-commit>` contains no path of the t1653-owned list (`plan.md` §A), and the sha256 of the `drain.sh` jq filter program text equals the pre-implementation value recorded at M1 start. Vacuous today, which is why it is a regression guard and not an adopted criterion.
- **RG-3** — `drain_test.sh` and `session_drain_test.sh` keep ending `exit=0` with all `ok` lines (EL-RG), extended with the new cases.

## §E Edge cases (each covered by an AC or a named test)

1. Trailing partial line at snapshot time: left unread (AC-LIS-002).
2. Checkpoint exactly at the tail: `NOOP_AT_TAIL` only if the lineage verified (AC-LIS-001/-006).
3. Same-content rows: distinct identities, counted individually in the equation (AC-LIS-008, -016).
4. Spool non-empty at exclusive release: merged before release (AC-LIS-009).
5. Cutover with a checkpoint in a mismatch state: preserved verbatim (AC-LIS-017).
6. Unparseable row: counted, preserved, never a whole-log verdict (AC-LIS-019).
7. Evidence reference to a commit not reachable from the cutover base: not eligible (AC-LIS-014).
8. Another session's consumer writing `clusters.json` during the freeze: the freeze holds the consumer lock (AC-LIS-012, R1).

## §F Quality gates and Definition of Done

- Tested: changed Go packages at or above 85% statement coverage (`go test -cover` scoped to the touched packages); every cell above quoted verbatim in the run-phase `progress.md` §E.2.
- Readable/Unified: gofmt clean; `golangci-lint run --timeout=2m` baseline-only on the touched packages; shell passes `bash -n` and `shellcheck` where installed.
- Secured: backups and ledger owner-only; no secret is written; row text (up to 200 runes of raw failure output) is never copied into reports, only counts, hashes, and identities.
- Trackable: conventional commits with card id t1901 in every message; the tracked baselines and step records (`progress.md` §E.2 lines, RED observations) are committed separately from the implementation they measure, in the order the claim needs (verification-claim-integrity §2.3). The gitignored evidence files and the ledger cannot witness order in the commit graph; they are cited by sha256 from the tracked records.
- Done when: all 24 criteria are green at the closing commit; RG-1..3 hold; R5 evidence exists; the real-path outcome is recorded as either a verified cutover or `zero_removal`; the provenance ledger verifies; `decision-index.md` has no open row that blocks the autonomous Kickoff.
