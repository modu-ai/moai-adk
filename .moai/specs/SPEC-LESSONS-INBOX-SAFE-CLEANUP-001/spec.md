---
id: SPEC-LESSONS-INBOX-SAFE-CLEANUP-001
title: "Lessons inbox safe improvement and cleanup: generation-checked consumer, locked atomic replace, provenance-backed removal of proven items only"
version: "0.2.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.3.0"
module: ".claude/skills/hns-lsel-curator, internal/hook, internal/cli"
lifecycle: spec-anchored
tags: "lessons-inbox, lsel, drain, checkpoint, generation, atomic-replace, lock, provenance-ledger, cleanup, retention-hold"
tier: L
---

# SPEC-LESSONS-INBOX-SAFE-CLEANUP-001

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-19 | manager-spec | Initial plan-phase artifacts for card t1901 (Tier L). Operator instruction relayed 2026-10-19 via ADK coordination; PR1841 retention-hold case folded in from the same-day operator follow-up (board 559a). |
| 0.2.0 | 2026-10-19 | manager-spec | Plan-audit round 1 repair (FAIL 0.74, audited 1c0cee7e78). D1 persistent blocked state (REQ-LIS-002, -004, AC-LIS-001); D2 named test seams replace text-coupled shims (design 2.6, AC-LIS-002, -003); D3 manifest-decided recovery at every crash point (REQ-LIS-005, design 2.4); D4 final-merge-under-lock spool handoff (REQ-LIS-010); D5 GENERATION_CHANGED and NOOP_AT_TAIL asserted (AC-LIS-002); D6 tightened hold-release predicate (REQ-LIS-016, AC-LIS-024); D7 prefix property restated for cutover (REQ-LIS-024, AC-LIS-023); D8 real-path pin step R3b; D9 ordering witness restated to tracked records; D10 lock and spool placed in the ignored state directory with gitignore assertion; D11 AC-LIS-008 RED made runnable; optional D12, D14, D15 applied, D13 partly. REQ 25 and AC 24 unchanged. |

## 1. Context

### 1.1 Problem

The lessons inbox (`.moai/lessons-inbox.jsonl`) is an append-only stream of failure observations. Its only consumer, the LSEL curator drain (`drain.sh` + `session_drain.sh`), tracks progress with a bare line-count checkpoint (`drain-offset.json`). The checkpoint carries no identity of the file it counted, the consumer reads in two non-atomic steps, the two outputs (clusters, checkpoint) are written by shell redirection straight over the live files, and nothing joins the producer (the hook's `O_APPEND` writer), the consumer, and any cleanup under one lock contract. Nothing in the tree can say which observations were processed, so nothing can be removed safely.

### 1.2 Measured on this tree (base `8249bff8c3`)

Every row below was produced in this run by the command shown in `acceptance.md` §C (evidence ledger); sources are unchanged between the base and the measurement (untracked files exist only under this SPEC directory).

| Fact | Measured value |
|------|----------------|
| The nine ADK drain conditions (ADK input) reproduce on synthetic fixtures | all nine reproduce; `acceptance.md` EL-C01..EL-C10 |
| A tenth condition (this SPEC): a replaced inbox of at least checkpoint length is consumed as the same lineage | reproduced, EL-C10 (`run2_read=2` from a different file) |
| Condition 9 is wider than the ADK wording | after a zero-byte checkpoint, the next run also leaves `clusters.json` at zero bytes (`run2_clusters=EMPTY-FILE`) |
| Existing curator tests on the unmodified tree | `drain_test.sh` and `session_drain_test.sh` both end `exit=0` with all `ok` lines (EL-RG); they cover only the normal fixture |
| Real inbox rows at measurement time (read-only `jq`) | 23051 rows, 0 carry an `id` field, 22667 distinct row byte-strings (so at least 384 byte-identical duplicate rows) |
| Real checkpoint | `{"offset":23339,"updated":"2026-10-15T04:30:35Z"}`, ahead of the file by 288 at 23051 rows (the file grows by append, so this number moves) |
| Real label distribution (observation-time labels, not root causes) | 13778 UnknownFailure, 8435 WorktreeGuardRefusal, 210 TimeoutError, 183 OOMKilled, 148 ExitError, 130 SandboxViolation, 113 PermissionDenied, 50 ContextCancelled; sum 23047 at that read |
| Producer lock | none: no `Flock`/`LockFileEx`/`.lock` token in `failure_observer.go` or `inbox_lifecycle.go` (EL-R3); the rotation chain renames the live file with no lock (`inbox_lifecycle.go:136`, EL-R4) |
| Template mirror | none for the curator skill and the Go code: `find internal/template/templates -iname "*lsel*"` is empty and `hns-*` skills are a user-owned namespace barred from the template FS (`namespace_protection_audit_test.go`). The new runtime files are different: `.gitignore:311` and `internal/template/templates/.gitignore:302` ignore only the exact name `.moai/lessons-inbox.jsonl`, and `git check-ignore -v` on `.moai/lessons-inbox.lock` and `.moai/lessons-inbox.spool.jsonl` exits 1 (EL-D10); the spool holds raw failure text. Both trees already ignore `.moai/state/` (`.gitignore:401`, template `.gitignore:242`), so the lock and spool live under `.moai/state/lessons-inbox/` (REQ-LIS-010) and the existing `TestTemplateGitignoreIgnoresLocalArtifacts` (`internal/template/gitignore_local_artifacts_test.go:29`) is extended; Template-First applies to that test, not to a new ignore line |
| Seam fixtures 13-17 | injection points defined by the consumer contract (design 2.6) replace the legacy text-coupled shims for every GREEN; on this tree no seam exists, so each prints `hook_reached=false` (EL-C13..EL-C17) |
| Installed `moai` build | `moai_cp/20261007_195927-1888-g7705b365e2`; `git merge-base --is-ancestor 7705b365e2 HEAD` returns exit 0, so the installed build is a strict ancestor of HEAD (it lags). Run-phase measurements of new Go verbs MUST invoke a build made from the tree, by path |

Third-party inputs (ADK, read-only analyses of other trees; numbers NOT re-measured here unless a row above says so):

| Input | sha256 |
|-------|--------|
| `.moai/reports/t1901/inputs/adk-lessons-review-report.md` | `635a5839f413b352463b3def91e1ea579c1f3c4e08a8e289ddadb5aeba9365b4` |
| `.moai/reports/t1901/inputs/lessons-analysis-report.md` | `cfedea6183e3e2996f03a55b717a94a08abd4290c0be08d68c46f1ab90ae6d0f` |
| `.moai/reports/t1901/inputs/case-pr1841-premerge-review-REPORT.md` | `9b1ddc40255b24a39b0a6ed1bb424a395b34604b695d9b048a403272bbf22b41` |
| `.moai/reports/t1901/inputs/case-pr1841-native-codex-rest-comment.json` | `c8ae797c1d05bebd858ae2718adf3252366ba0e5330b20c140090d18c27d1e9b` |

ADK figures not re-measured: 18,116 truncated summaries (this tree measured 18158 at a later read, same order), 10,966 repeated-summary extra rows, 2,137 historical-only rows, birthtime and prefix-lineage findings of the real file, and every PR/CI fact in the PR1841 case (merge time, head SHA, review completion times, `reviews=[]`). Those are inputs; none is an acceptance baseline.

### 1.3 Worked case: PR1841 (retention-hold example)

Per the operator follow-up relayed by the leader, the pull request 1841 was merged before the native code and security reviews completed (third-party facts, `case-pr1841-premerge-review-REPORT.md` lines 38-60: merge 05:13:31Z, code review completion 05:14:11Z, security 05:15:04Z, `reviews=[]`). The case type is resolved only by a guideline (card t1900) plus a code gate (card t1868). Until both have landed AND been independently verified, the case must stay unprocessed and out of any removal set. This SPEC turns that instruction into a mechanical rule (REQ-LIS-016) and a named mutant (AC-LIS-024).

### 1.4 Boundary with card t1653

Card t1653 (state hold) owns the improvement loop semantics: when the offset is reset, rotation rules, the 91-proposal bundle, noise separation, mining and diff proposals, quarantined verification, approvals, the apply ledger and rollback, and the `AGENTS.local.md` review. This card owns mechanism only. The boundary table is `plan.md` §A, written first; interface contracts IC-1..IC-4 let t1653 build on the mechanism without editing it.

## 2. Requirements (GEARS)

### 2.1 Consumer states, boundary, checkpoint

- **REQ-LIS-001** (Ubiquitous): The inbox consumer shall classify every run into exactly one named state from {OK, NOOP_AT_TAIL, OFFSET_BEYOND_FILE, LINEAGE_MISMATCH, CHECKPOINT_CORRUPT, LOCK_HELD, COMMIT_FAILED, GENERATION_CHANGED} and record it, with a timestamp and the exit code, in `<state-dir>/consumer-status.json`; the state NOOP_AT_TAIL shall be reported only when the checkpoint offset equals the complete-line count of a lineage-verified file.
- **REQ-LIS-002** (Event-driven): When the checkpoint offset exceeds the complete-line count of the inbox, the consumer shall report OFFSET_BEYOND_FILE, record that observation as a persistent blocked state, consume nothing, leave the checkpoint and clusters files byte-unchanged, and shall not reset the offset to 0 or to the current line count; the blocked state shall survive any later growth of the inbox, including growth past the stored offset, and shall be cleared only by an explicit verified bind (REQ-LIS-020), never by the file growing; the same persistence applies to LINEAGE_MISMATCH and CHECKPOINT_CORRUPT.
- **REQ-LIS-003** (Event-driven): When the consumer reads the inbox, it shall take one fixed-boundary snapshot of complete lines and derive the row count, the cluster input, and the offset advance from that snapshot alone, so the saved offset equals the number of rows actually clustered; a trailing partial line shall remain unread; where the inbox generation identity read before the snapshot differs from the one read after it, the consumer shall discard the snapshot, commit nothing, and report GENERATION_CHANGED.
- **REQ-LIS-004** (Ubiquitous): The consumer checkpoint shall bind the lineage it counted (first-line anchor hash, byte offset, and a hash of the last complete-line window ending at the offset); where the stored binding does not match the file, the consumer shall report LINEAGE_MISMATCH and consume nothing; a legacy offset-only checkpoint shall be adopted once, flagged unverified in the status record, and bound from then on only if it lies within the file and no blocked state has ever been recorded for it; a checkpoint with a recorded blocked state shall never be adopted.
- **REQ-LIS-005** (Event-driven): When the consumer writes its clusters file and its checkpoint, it shall commit them as one pair through a commit manifest (same-directory temporary files, content verification, flush, rename), such that no interruption leaves a zero-byte or partial final file, the checkpoint is never ahead of the clusters it describes in any interrupted state, and at every interruption point the commit manifest alone decides whether the next run completes the pair or restores the retained previous pair, so the pair is wholly old or wholly new before that run consumes anything.
- **REQ-LIS-006** (Event-driven): When the checkpoint is empty or unparseable, the consumer shall report CHECKPOINT_CORRUPT, leave the clusters file untouched, and retain the previous committed checkpoint as a restorable copy.
- **REQ-LIS-007** (Ubiquitous): The consumer lock shall record its owner (process id and start time); a lock whose owner is dead shall be reaped, a lock held by a live owner shall be reported as LOCK_HELD in the status record, and the lock order between the consumer lock and the producer lock shall be fixed by one written contract.

### 2.2 Identity and shared lock

- **REQ-LIS-008** (Ubiquitous): Every row the hook producer writes after this SPEC lands shall carry a stable event `id` that needs no coordination between producers; the producer shall never rewrite an existing row.
- **REQ-LIS-009** (Ubiquitous): A generation manifest shall give every existing row a deterministic identity composed of its generation hash, its line position, and its row hash; two rows with identical content at different positions shall keep different identities and shall never be merged.
- **REQ-LIS-010** (State-driven): While an exclusive replacer (cutover or rotation) holds the producer lock, the hook producer shall neither lose an event nor block beyond a fixed bound; events that could not take the lock shall be held in a spool; the replacer shall merge the spool a final time while still holding the lock and close the spool atomically before releasing, and an event written after the close shall be merged exactly once by the next lock holder; the lock and the spool shall live in the ignored runtime directory `.moai/state/lessons-inbox/`, and the repository and distributed-template ignore files shall be asserted to cover both paths.
- **REQ-LIS-011** (Event-driven): When the existing rotation chain replaces the live inbox, it shall do so under the same exclusive producer lock, with no change to when it rotates or how many generations it keeps.

### 2.3 Analysis

- **REQ-LIS-012** (Event-driven): When the classifier runs over a frozen copy of the inbox, it shall assign every row exactly one class from a closed set (guideline-fix candidate, code-or-test-fix candidate, operations-or-environment, normal-refusal, unclassified-needs-evidence) together with its evidence basis and frequency, shall treat the observation-time labels as labels and not as root causes, shall never bulk-label an unknown category as noise, shall emit a disposition list for follow-up cards, and shall never decide that any row is removable.

### 2.4 Safe cleanup path

- **REQ-LIS-013** (Event-driven): When a cleanup begins, the tooling shall first write a byte-identical frozen copy of the inbox prefix (complete lines) and of the checkpoint, record their sha256, read the copies back and re-hash them, and set owner-only permissions, leaving the live files untouched.
- **REQ-LIS-014** (Ubiquitous): The inventory of a frozen copy shall list every row with its identity, position, hash, and class, and its row count shall equal the line count of the frozen copy.
- **REQ-LIS-015** (Ubiquitous): A row shall be eligible for removal only when its recorded state is `independently_verified` or `archived` AND every evidence reference of its chain resolves mechanically (the change commit is reachable from the cutover base, the verification record exists and its hash matches, the approval reference exists); rows in the states blocked, rejected, normal_refusal, environment_pending, verification_failed, reverted, or in any unresolved state shall be preserved, and a run that proves zero eligible rows shall be a valid outcome.
- **REQ-LIS-016** (State-driven): While an item depends on a card that has not landed and been independently verified, the cleanup shall keep that item active as pinned-unprocessed and exclude it from every removal set, whatever its recorded state says; a dependency counts as satisfied only when a landing commit carrying that card's id is reachable from the cutover base and a verification record names that same commit, carries a passing verdict, and names a verifier other than the author or the self-check of the card's own work, for every card the item requires.
- **REQ-LIS-017** (Event-driven): When a removal set is non-empty, the tooling shall write the removed rows to an archive, open it, count it, and hash it before any cutover, and record the archive hash in the ledger.
- **REQ-LIS-018** (Ubiquitous): Before and after every cutover the tooling shall verify the multiset equation (rows before plus rows appended since the freeze equals rows in the new active generation plus rows in the archive, counting byte-identical duplicates individually) and shall refuse the cutover on any imbalance.
- **REQ-LIS-019** (Event-driven): When a cutover runs, it shall replace the live inbox atomically under the exclusive producer lock, shall realign the consumer checkpoint by an exact removed-rows-before-offset mapping only where the checkpoint is verified (and otherwise preserve it verbatim), and, on any failure at any step, shall restore the previous generation and checkpoint byte-identical, retaining the archive.
- **REQ-LIS-020** (Event-driven): When an owner supplies a checkpoint value, the tooling shall accept only an explicit generation, line offset, and expected window hash, shall verify them against the file, and shall refuse any shortcut that chooses the value itself (reset to zero, jump to tail, or similar).
- **REQ-LIS-021** (Ubiquitous): The provenance ledger shall be append-only and hash-chained, and every removed item shall have a recorded trail item, then change, then verification, then archive, each link resolvable.
- **REQ-LIS-022** (Event-driven): When an operation would truncate the inbox using a checkpoint offset, permanently delete an inbox row source, the backed-up original or the archive, or declare the whole log invalid, the tooling shall refuse the operation; unparseable rows shall be counted individually and preserved.

### 2.5 Proof and boundary

- **REQ-LIS-023** (Event-driven): When the synthetic end-to-end run executes, the chain consumer, proposal, apply, verify, proof gate, archive, cutover shall run on synthetic data through the existing applier and verify scripts, and the resulting ledger trail shall be the only evidence the proof gate accepts.
- **REQ-LIS-024** (Event-driven): When the real-path steps run, they shall run as separately gated steps (freeze, read-only inventory and classification of the frozen copy, proof gate, and only if the removal set is non-empty archive then cutover), each leaving its own evidence file, with this lane as the single maintenance writer; while no cutover has occurred, the live inbox's first N bytes shall hash to the frozen copy's recorded hash after every step, and after a cutover the multiset equation, the new-generation hash, the archive count and the checkpoint mapping shall verify instead; the PR1841 case shall be recorded as a pinned item in the provenance ledger before any cutover; and each step shall refuse to start without the previous step's recorded hash, the step order being witnessed by one tracked record per step in separate commits.
- **REQ-LIS-025** (Ubiquitous): The mechanism shall expose the interface contracts IC-1..IC-4 of `plan.md` §A and shall not edit the improvement-loop semantics owned by card t1653.

## 3. Constraints

- Mechanism only: this SPEC never decides when an offset is reset, what a proposal means, or whether an item is processed; those inputs arrive as explicit evidence.
- Existing rows are immutable; no tool rewrites a row in place.
- Fail-open at session start is preserved: the wrapper still exits 0, and the state is recorded instead.
- Resource discipline for run phase: no `go test ./...`; scoped `-run` and package lists; heavy runs in a named slot (`plan.md` §D).
- No time estimates; priority and ordering only.

## 4. Exclusions

### Out of Scope — loop semantics owned by card t1653

- Deciding when the real offset is reset or rebound (this SPEC supplies the verified bind primitive only).
- The 91-proposal bundle, noise separation policy, mining, diff proposals, approval policy, rollback policy of applied changes, and the `AGENTS.local.md` review.
- Rotation policy (when to rotate, how many generations to keep, eviction of the oldest generation); only the lock around the existing rename chain is in scope.

### Out of Scope — follow-up cards recommended by the planner

- Rewriting the PermissionDenied advice string at `internal/hook/post_tool_failure.go:219` (it recommends `bypassPermissions`; verified by reading the source, EL-P). A one-line code-and-test fix with a different subsystem; the classifier lists it as a disposition for a follow-up card.
- The worktree-guard refusal guidance (8435 rows) and the unknown-category root-cause work: they are classified here, fixed elsewhere.
- Retention of `clusters-history` (100 newest files) and `last-drain.log`; they are regenerable from the inbox plus the commit log and are not observations.
- The guideline and code-gate fixes for the PR1841 premerge-review case (cards t1900 and t1868).

### Out of Scope — irreversible or external-shared acts

- Permanent deletion of the backed-up original, the archive, or the frozen copies: an operator act, never performed by this tooling.
- Any push, pull request, release, or settings change.
- Re-training or fine-tuning of any model.

## 5. Premise ledger

| Premise | Status |
|---------|--------|
| Nine drain conditions reproduce on this tree | measured (EL-C01..EL-C09) |
| Real inbox has no `id`, has byte-duplicate rows, checkpoint ahead of file | measured read-only (EL-R2, §1.2) |
| No proven processed item exists on the real data today | NOT measured; the expected real outcome (zero removals) is a hypothesis the real-path steps test, and the SPEC accepts zero removals as a pass |
| PR1841 facts, t1900/t1868 status | third-party/relayed, not re-measured |
| Real apply-ledger or proposals exist in the primary state dir | measured absent: the primary `.moai/state/lsel` holds `clusters-history`, `clusters.json`, `drain-offset.json`, `last-drain.log` only |
