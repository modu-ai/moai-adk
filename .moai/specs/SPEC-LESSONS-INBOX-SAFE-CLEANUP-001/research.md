# research.md — SPEC-LESSONS-INBOX-SAFE-CLEANUP-001

Codebase and evidence research. Every figure is either measured on tree `8249bff8c3` by the command in `acceptance.md` §C (ledger id given) or labelled as an input not re-measured.

## 1. What the consumer does today (read, then reproduced)

`drain.sh` (sha256 in EL-S): reads the offset (`jq`), counts lines with `wc -l`, returns a no-op when `TOTAL_LINES <= OFFSET_BEFORE` (and rewrites `clusters.json` with `total_read: 0` and `offset_after` equal to the stale offset), then runs `tail -n +offset+1 | jq -s` with `offset_after = TOTAL_LINES` from the earlier count, writing clusters and then the checkpoint by redirection straight over the final paths. `session_drain.sh` wraps it with a mkdir lock (reaped by age), an archive of the previous `clusters.json`, and a fail-open exit 0 for every failure.

Reproduced (EL-C01..EL-C12): the nine ADK conditions, a tenth (a replaced file is consumed as the same lineage), a stress run (rows clustered 232-233 against 220 rows, three of three runs), and a wider form of condition 9 (the staged `clusters.json` is truncated as well, not only the checkpoint left empty).

What the existing tests cover (EL-RG): the normal fixture, lock contention as a skip, archive-before-overwrite, no-op, fail-open, unwritable log, and one mutant. None of the twelve reproduced conditions is among them, which is why both scripts pass on the unmodified tree.

## 2. Producer and rotation

`appendLessonsInboxStub` marshals one row and writes it with `O_APPEND` in a single `Write`; there is no lock and no id (EL-R1, EL-R3). `RotateLessonsInboxArchive` renames the live file at `inbox_lifecycle.go:136` with no lock (EL-R4) and the cap path is armed only when `.moai/state/lsel/` is absent, so on a curator machine the inbox is never rotated and grows without bound (23051 rows on the real file at measurement, EL-R2). The existing concurrency test asserts no panic, generation count and size bound; it does not assert that every appended row survives (read of `inbox_lifecycle_test.go` lines 382-422).

Consequence for design: an exclusive replacer is required to keep producers from appending to an inode that is about to be superseded; a shared/exclusive file lock on the Go side with a spool fallback keeps the producer fail-open without losing events.

## 3. Why a bounded window hash rather than a prefix hash

The curator inbox is never rotated, so a prefix hash costs the whole file at every session start, against the repository's Advisory-Check Discipline (cost independent of corpus size). The anchor hash (first line) plus the window hash of the last 8 consumed lines ending at a recorded byte offset is constant cost and catches both replacement shapes reproduced here (EL-C10 differs in the first line; a same-first-line replacement is caught by the window). Residual risk: `design.md` §2.2 and `plan.md` §E.

## 4. Real data, read-only (not an acceptance baseline; the file grows)

- Label distribution at one read (EL-L): UnknownFailure 13778, WorktreeGuardRefusal 8435, TimeoutError 210, OOMKilled 183, ExitError 148, SandboxViolation 130, PermissionDenied 113, ContextCancelled 50 (sum 23047), all of kind `tool_failure`, timestamps 2026-07-10T09:43:18Z to 2026-10-19T05:20:37Z. These are observation-time labels, not root causes.
- Identity (EL-R2): 23051 rows, none with an `id`, 22667 distinct byte-strings; byte-identical duplicate rows therefore exist and must be counted, not merged.
- Checkpoint: `{"offset":23339,"updated":"2026-10-15T04:30:35Z"}`, 288 rows beyond the file at 23051 rows.
- Processing evidence in the primary LSEL state (EL-R11): `clusters-history`, `clusters.json`, `drain-offset.json`, `last-drain.log` only; no proposals directory and no apply ledger. So no row has a resolvable change-and-verification chain today.
- Rows carrying no `v` field: 7654 of 23047 at one read (the earlier schema); the consumer ignores unknown fields.

## 5. ADK inputs: what was re-measured and what was not

| ADK claim (input) | Status |
|-------------------|--------|
| Nine synthetic conditions | re-measured on this tree (all reproduce) |
| Real file 22,989 lines, checkpoint 23,339 | not re-measured (file moved on); this tree measured 23047-23051 and the same checkpoint |
| Birthtime 2026-10-15T04:36:32Z, 19,804-line prefix equal to an older snapshot, 2,137 historical-only rows, replacement actor unknown | not re-measured; informs only the hypothesis that the checkpoint belongs to an earlier file generation, and is the reason this SPEC reports a mismatch instead of resetting |
| 18,116 rows with a truncated summary | not re-measured as stated; this tree measured 18158 at a later read (same order) |
| PermissionDenied advice recommends `bypassPermissions` | re-measured by reading the source (EL-P) |
| `inbox_lifecycle_test.go:383` concurrent-append test lacks a row-preservation assertion | re-read: confirmed |
| Go `harness execute` exists as a production apply path | not re-measured; not needed by this SPEC |
| PR1841 merge time, head SHA, review completion times, `reviews=[]` | third-party, not re-measured |

## 6. Template mirror and namespace

`find internal/template/templates -iname "*lsel*"` is empty and `namespace_protection_audit_test.go` forbids `hns-*` skills in the template FS (user-owned namespace). `internal/hook`, `internal/cli`, and the new package have no template mirror. That conclusion holds for source files only. The new runtime files do touch the template: `.gitignore:311` and `internal/template/templates/.gitignore:302` ignore only the exact inbox filename, `git check-ignore -v .moai/lessons-inbox.lock .moai/lessons-inbox.spool.jsonl` exits 1 (EL-D10), and the spool holds raw failure text. Both trees ignore `.moai/state/` (`.gitignore:401`, template `.gitignore:242`), so the files are placed under `.moai/state/lessons-inbox/` and the existing `TestTemplateGitignoreIgnoresLocalArtifacts` (one test function whose table lists path, expectation, reason) is extended with rows built from the exported path constants; that is the Template-First step. One lexical mention of the inbox exists in a distributed template (`navigator-audit.sh` comment) and is not touched.

## 7. Open measurement gaps carried into the plan

G1 the behavior of the producer lock bound under real hook load is unmeasured (AC-LIS-009 measures it on synthetic load only); G2 durability of `fsync` from shell; G3 whether an in-process write to the primary checkout path is accepted from a worktree session (R0 measures it); G4 the share of UnknownFailure rows that are guideline versus code defects (needs per-row evidence, which the 200-rune summaries often lack); G5 the stress fixture is probabilistic (three of three red on the current tree; the deterministic shim fixtures back it).
