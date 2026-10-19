# design.md — SPEC-LESSONS-INBOX-SAFE-CLEANUP-001

System design at the level the audit needs: components, contracts, and protocols. Function and type names are deferred to the run phase; names that appear here (state names, file names, verbs) are the user-visible or cross-card contract surface.

## 1. Component map

| Component | Language | Role | Milestone |
|-----------|----------|------|-----------|
| Consumer (`drain.sh`, `session_drain.sh`) | shell + jq | Fixed-boundary read, state classification, v2 checkpoint, pair commit, lock owner record. The `is_noise` filter and clustering program stay byte-identical | M1, M2 |
| `internal/inboxsafe` (new package) | Go | Lock helper (exclusive/shared + spool), generation manifest, backup, inventory, classifier, proof gate, archive, equation, cutover, rollback, ledger | M3-M5 |
| Hook producer (`failure_observer.go`) | Go | Adds the stable `id`; takes the shared producer lock with a bound and spool fallback | M3 |
| Rotation (`inbox_lifecycle.go`) | Go | Calls the exclusive helper around its rename chain; policy untouched | M3 |
| `moai inbox` verbs (`internal/cli`) | Go | `backup`, `inventory`, `classify`, `checkpoint bind`, `prove`, `archive`, `cutover`, `rollback`, `ledger`; existing `status`/`drain` unchanged | M4, M5 |
| E2E harness (`e2e_test.sh`) | shell | Drives the existing applier and verify scripts on a synthetic repository | M6 |

Language split rationale (Q2 default): the consumer is already shell and runs at session start; keeping it shell means it never depends on the installed `moai` build, which lags the tree (measured, EL-B). Everything that needs an OS-level lock shared with the Go producer, an exclusive rename under that lock, or cross-platform behavior is Go. The consumer does not take the OS lock: it validates optimistically (§3).

## 2. Consumer protocol (M1, M2)

### 2.1 States, exit codes, status record

| State | `drain.sh` exit | Meaning | Consumed? |
|-------|-----------------|---------|-----------|
| OK | 0 | rows read from a verified lineage, pair committed | yes |
| NOOP_AT_TAIL | 0 | verified lineage, offset equals the complete-line count | none to read |
| OFFSET_BEYOND_FILE | 3 | offset greater than the complete-line count | no |
| LINEAGE_MISMATCH | 4 | stored binding does not match the file | no |
| CHECKPOINT_CORRUPT | 5 | checkpoint empty or unparseable | no |
| COMMIT_FAILED | 6 | the pair commit did not complete (previous pair intact) | no (retry next run) |
| GENERATION_CHANGED | 7 | generation pointer moved during the read | no (snapshot discarded) |
| LOCK_HELD | 75 | another live consumer owns the lock | no |

`session_drain.sh` keeps its fail-open contract: it always exits 0 and the session start is never blocked. The record is `<state-dir>/consumer-status.json`, written by the same temp-then-rename step: `{schema:1, state, ts, exit_code, detail, checkpoint_unverified}`. A stderr notice is printed only when the state differs from the previous status record (so the real, still-mismatched checkpoint does not print on every session start); the record updates every run. This file is interface contract IC-1.

### 2.2 Checkpoint v2

`drain-offset.json` schema 2: `{schema:2, anchor_sha256, offset_lines, offset_bytes, window_lines, window_sha256, commit, updated}`.

- `anchor_sha256`: hash of the first complete line (the lineage anchor). A replaced file almost always changes it.
- `offset_bytes` + `window_sha256`: byte position of the end of the last consumed line and the hash of the last up to W consumed complete lines (W fixed at 8). Checking the window is O(W), independent of corpus size (Advisory-Check Discipline: constant cost). A file replaced by a different file sharing the first line and length is caught by the window.
- Legacy v1 `{offset, updated}`: readable. Offset greater than the file: OFFSET_BEYOND_FILE. Offset within the file: adopted once (Q1 default), `checkpoint_unverified: true` for that run, binding written by the commit. A v1 checkpoint is never "verified" retroactively.

Reasons a full-prefix hash was not chosen: it costs the whole file at every session start and the inbox on a curator machine is never rotated (research.md §3). Residual risk: two different files that agree on the first line, the byte offset, and the last 8 lines before the offset are indistinguishable; the probability is negligible for failure-observation rows that carry second-resolution timestamps, and is recorded in `plan.md` §E.

### 2.3 Fixed-boundary read

One open of the inbox copies `[offset_bytes, EOF)` into a snapshot file under the state directory; the snapshot is trimmed to the last newline (a trailing partial line stays unread). The row count, the cluster input, the new `offset_lines`, the new `offset_bytes`, and the new window hash all derive from the snapshot alone. The generation identity (anchor hash plus inode/size monotonicity, read before and after the copy) detects a swap during the copy: mismatch discards the snapshot and reports GENERATION_CHANGED. No separate `wc` exists, so the race of condition 6 cannot occur.

### 2.4 Pair commit with restart recovery

Order: (1) build `clusters.json.<cid>.tmp` and `drain-offset.json.<cid>.tmp` in the state directory; (2) verify both parse and satisfy the invariants (clusters `offset_after` equals the checkpoint `offset_lines`); (3) flush both; (4) write `commit.json` `{cid, phase:"PREPARED", expected sha256 of both}` by temp-and-rename; (5) rename clusters, rename checkpoint; (6) rewrite `commit.json` with phase `COMMITTED`; (7) append one line to `commits.jsonl` (slice from/to, hashes). Recovery at the start of every run: a `PREPARED` manifest with both temporaries present rolls forward; a `PREPARED` manifest with a missing or mismatching temporary rolls back (delete the temporaries, mark `ABORTED`); the checkpoint therefore never exceeds the clusters it describes, and no zero-byte final file can occur because no redirect targets a final path. Durability: the shell cannot call `fsync`; the flush step uses the best available tool and records `durability: best-effort` in the manifest when only `sync` is available (Gap G2). Process-crash atomicity does not depend on it; power-loss durability does.

### 2.5 Consumer lock

`drain.lock` stays a directory (mkdir is atomic everywhere), now holding an `owner` file `{pid, start_ts, host}`. A lock is reaped only when its owner process is not alive (`kill -0`); a lock directory without an owner file older than a fixed grace is reaped as before; a live owner is never reaped however old. This is the consumer half of IC-4.

## 3. Lock contract (IC-4)

Two primitives, one written contract (it is also copied into `SKILL.md` of the curator at M2):

| Lock | Type | Who | Held for |
|------|------|-----|----------|
| A `drain.lock` | mkdir directory with owner file | consumer, backup/freeze, cutover, bind | consumer pass; freeze; cutover |
| B `.moai/lessons-inbox.lock` | OS file lock: shared and exclusive (flock on Unix, LockFileEx on Windows, build-tagged) | producer takes shared around open-write-close; rotation and cutover take exclusive | one append / one replace |

Order: A before B; never wait for A while holding B. The consumer never takes B: it reads optimistically and validates with the generation identity (§2.3). Producer behavior on contention: bounded wait on B (fixed small bound), then append the row to `lessons-inbox.spool.jsonl` with `O_APPEND`; the exclusive holder merges the spool into the new generation before releasing B (REQ-LIS-010). Lock failure (error other than contention) fails open to the plain append, as today.

## 4. Identity

New rows: `id` = a random 128-bit value with a time prefix, produced without coordination. Existing rows: `lg:<generation_hash>:<line>:<row_sha256_prefix>` computed into the manifest, never written back. A frozen file's `generation_hash` is the sha256 of its bytes; a live file is identified by its anchor hash plus the cutover epoch recorded in the generation pointer. Same-content rows get distinct identities from their positions and are individually counted by the equation.

## 5. Safe cleanup protocol (M4, M5)

### 5.1 Item states and the proof gate

States: observed, triaged, proposed, approved, applied, independently_verified, archived, removed_from_active. Non-removal terminals: blocked, rejected, normal_refusal, environment_pending, verification_failed, reverted.

The gate never reads a state string as proof. Eligibility needs all of: recorded state in {independently_verified, archived}; the change commit resolves (`git cat-file -e` and ancestor of the cutover base); the verification record file exists and its sha256 equals the recorded hash; the approval reference exists; the apply-ledger row (when referenced) re-hashes to the recorded value and carries `verified:true`. A missing or unresolvable link makes the item `UNPROVEN` and keeps it. `PINNED_UNPROCESSED` (§6) is checked before eligibility.

### 5.2 Order of operations

freeze (lock A, copy, hash, read back) -> inventory (frozen copy only) -> classify -> prove -> if non-empty: archive (write, reopen, count, hash, ledger row) -> cutover (lock A then B exclusive; merge spool; read appended rows since the freeze; equation; build new generation in the same directory; verify; flush; write the commit manifest; rename; write the generation pointer; map or preserve the checkpoint; release) -> post-verify. A failure at any step restores the retained original generation byte-identical.

### 5.3 Checkpoint realignment

Verified checkpoint: `offset_lines_new = offset_lines_old - (removed rows whose position <= offset_lines_old)`, byte offset and window recomputed from the new generation, written in the same commit manifest as the inbox rename. Checkpoint in a mismatch state (the real one today): preserved verbatim and the state remains reported; the mechanism never invents a mapping. Deciding to rebind is IC-2 and belongs to the card that owns the semantics.

### 5.4 Classification

Closed class set: guideline-fix candidate, code-or-test-fix candidate, operations-or-environment, normal-refusal, unclassified-needs-evidence. Initial routing hypotheses from the observation labels (hypotheses, not facts; evidence per row decides): WorktreeGuardRefusal -> normal-refusal with a guideline candidate; SandboxViolation, PermissionDenied, ContextCancelled -> normal-refusal or operations (PermissionDenied also lists the advice-string code candidate, EL-P); TimeoutError, OOMKilled -> operations-or-environment pending evidence; ExitError -> code-or-test candidate pending reproduction; UnknownFailure -> unclassified-needs-evidence, sub-indexed by non-identifying clues, never bulk-labelled noise. The output lists dispositions for follow-up cards and never carries a removal decision.

## 6. Provenance ledger and the worked PR1841 retention-hold case

Ledger: `.moai/state/lessons-cleanup/ledger.jsonl`, append-only; each row `{seq, ts, op, refs, counts, prev_row_sha256}`; `ledger verify` recomputes the chain. Companion `items.jsonl` holds item records `{item_id, class, state, evidence[], requires[]}`. Evidence refs are `{path, sha256 | row_sha256, commit}` (IC-3).

**Retention hold (REQ-LIS-016).** An item may carry `requires: [{card, kind: "landed_and_verified", evidence}]`. The gate evaluates every requirement the same mechanical way (commit reachable from the cutover base AND a verification record that resolves). Any requirement unmet means state `PINNED_UNPROCESSED`: the item is kept active, excluded from every removal set, and a `pin` ledger row records the dependency list. The hold overrides the recorded state, so a hand-edited `independently_verified` cannot release it.

**Worked example: PR1841 (third-party facts, not re-measured; `case-pr1841-premerge-review-REPORT.md` lines 38-60).** The pull request was merged at 05:13:31Z with `reviews=[]` while the native code review completed at 05:14:11Z and the security review at 05:15:04Z; a later pass does not retroactively cover the missing precondition. The case type is resolved only by a guideline (card t1900) plus a code gate (card t1868). The item therefore carries `requires: [{card:"t1900"}, {card:"t1868"}]`. Trail states: observed (the case report, hash `9b1ddc40...22b41`, plus the REST comment hash `c8ae797c...e1e9b` as supporting evidence) -> triaged (class: code-or-test-fix candidate plus guideline candidate) -> pinned. It is released only when both cards have landed and an independent verification record for each resolves; until then every cutover leaves it active, and the ledger shows the pin row. The named mutant `remove-before-t1900-t1868` (AC-LIS-024) is a removal criterion that skips `requires` or accepts landed-but-unverified; it must fail the case "both landed, one unverified".

How the case is recorded (a ledger/items entry referencing the report hashes, or an inbox row appended through the producer path) is decision-index Q4. In either option the lane does not hand-write inbox rows outside the mechanism.

## 7. Real safe path (M7)

R0 preconditions and build-by-path; R1 freeze under lock A (the other sessions' session-start drains also take A, so the freeze never interleaves with them); R2 read-only inventory/classification of the frozen copy; R3 prove; R4 only for a non-empty removal set; R5 post-verify. Prefix property: after each step `sha256(live[0:N]) == frozen_sha256` for the frozen length N; the only permitted change to the live inbox between steps is appended rows. Other sessions keep appending throughout (producers hold shared B for microseconds); only cutover takes exclusive B, and the spool covers any producer that waits longer than its bound. The single-writer statement means single maintenance writer: the producers' appends are the one mutation the lock contract is built for.

Expected outcome, stated as acceptable: the primary LSEL state holds no proposals directory and no apply ledger (EL-R11), so zero items currently have a resolvable chain; the real run is expected to end `zero_removal` with a complete inventory, classification, and ledger. That outcome is a pass of AC-LIS-023, not a gap.

## 8. Interface contracts

| IC | Surface | Reader/caller | Guarantee |
|----|---------|---------------|-----------|
| IC-1 | `consumer-status.json` | t1653 loop, session notices | closed state set, written every run |
| IC-2 | `moai inbox checkpoint bind` | the owner of offset semantics | explicit verified values only; no reset shortcut |
| IC-3 | evidence reference `{path, sha256, commit}` | t1653 apply ledger, verdict files | re-resolved by the gate; never written by t1901 |
| IC-4 | lock contract (A and B, order) | producer, rotation, t1653 rotation redesign | exclusive/shared helper exported from `internal/inboxsafe` |

## 9. Alternatives considered

Full-prefix hash per run (cost grows with the file); a database for items (new dependency, no need); `flock` in the shell consumer (absent on stock macOS, the very failure the wrapper's comments describe); rewriting legacy rows to add ids (violates immutability); moving the consumer to Go (couples session start to the installed build).
