# plan.md — SPEC-LESSONS-INBOX-SAFE-CLEANUP-001

Derived implementation plan. Order: boundary table first (card instruction), then the decisions most likely to change (public contracts), then mechanical steps. No time estimates; priority and ordering only.

## §A Boundary table versus card t1653 (written first)

Card t1653 is live and in state hold (`moai todo show t1653`, 2026-10-19). Rule for every row: this card (t1901) owns **mechanism** (lock, stable id, generation check, atomic replace, backup, ledger, proof gate, cleanup); t1653 owns **semantics** (when, whether, and what it means). t1901 never decides when the offset is reset or what a proposal means. Items marked IC are interface contracts t1653 may build on without editing t1901 files, and the reverse.

| # | Concern | t1653 owns (semantics) | t1901 owns (mechanism) | Interface contract and collision rule |
|---|---------|------------------------|------------------------|----------------------------------------|
| 1 | Offset self-reset | The decision that an offset is wrong, and the value to rebind to | Detecting the mismatch (`OFFSET_BEYOND_FILE`, `LINEAGE_MISMATCH`), refusing to consume, and a verified bind primitive | **IC-1** `consumer-status.json` (read-only for t1653): schema `{schema:1, state, ts, exit_code, detail}`, closed state set of REQ-LIS-001. **IC-2** `moai inbox checkpoint bind --generation G --offset-lines N --expect-window-sha256 H`: explicit values only, verified against the file, atomic; no reset shortcut exists. t1901 never calls IC-2 on real data |
| 2 | Rotation | When to rotate, how many generations, oldest-generation eviction | One exclusive-lock acquisition around the existing rename chain (`RotateLessonsInboxArchive`), so a concurrent append cannot be lost | **IC-4** lock contract (`design.md` §3). t1653's rotation redesign calls the same exported exclusive helper instead of renaming bare. Policy constants (`DefaultInboxMaxBytes`, `DefaultInboxArchiveGenerations`) are untouched by t1901 |
| 3 | 91-proposal bundle | Bundling, ordering, and handling of the 91 proposals | Nothing | t1901 never reads or writes `.moai/harness/proposals/`. Proposal artifacts reach the mechanism only as evidence references (IC-3) |
| 4 | Improvement loop, one round | Running the loop on real data, mining, diff proposals, prediction and verified records | A synthetic end-to-end harness (AC-LIS-022) proving the pipeline segments connect; no loop round is run on real data | The harness uses the existing `lsel-apply.sh` and `verify.sh` as black boxes; neither file is edited |
| 5 | Apply ledger (`apply-ledger.jsonl`) and quarantined verification | The applier's ledger rows, verified markers, revert handling, approval policy | A separate provenance ledger (`.moai/state/lessons-cleanup/ledger.jsonl`) that references apply-ledger rows by path plus row sha256 | **IC-3** evidence reference `{path, row_sha256 | file_sha256, commit}`: the proof gate re-resolves it; t1901 never writes an apply-ledger row; t1653 never needs to write the cleanup ledger |
| 6 | `AGENTS.local.md` review | The whole review task and edits to the user-owned layer | Nothing | no shared file |
| 7 | Noise separation (`is_noise` filter and clustering/importance in `drain.sh`) | Which categories are noise, thresholds, scoring | Nothing in the filter; t1901 restructures only the read, state, and commit steps around it | Guard RG-2: the jq filter program text of `drain.sh` is byte-identical before and after (hash recorded in the evidence ledger at M1 start and checked at closing) |
| 8 | Offset protection (REQ-CAP-005) | The requirement itself | Strengthened: a mismatching offset is reported, never silently skipped, never auto-changed | Cited by REQ-LIS-002/-020 |
| 9 | Item state judgment (`triaged`, `proposed`, `approved`) | Who judges and when | Recording the evidence chain and mechanically resolving it (REQ-LIS-015) | The mechanism never advances an item by itself: state comes from resolvable evidence supplied by the judging owner |

File ownership for blockers: t1901 edits `.claude/skills/hns-lsel-curator/{drain.sh,session_drain.sh,*_test.sh}`, `internal/hook/{failure_observer.go,inbox_lifecycle.go}` and their tests, a new `internal/inboxsafe` package, `internal/cli/inbox.go` plus new `internal/cli/inbox_*.go` verb files, the existing test `internal/template/gitignore_local_artifacts_test.go` (rows only, Template-First), and this SPEC. t1653 owns `.claude/skills/hns-lsel-applier/**`, `.moai/hooks/lsel-apply.sh`, `internal/harness/**`, `internal/hook/post_tool_failure.go`, and `.moai/harness/**`. A needed edit across the line is a blocker report to the leader, never a quiet edit. Closing guard RG-2: `git diff --name-only <base>..<closing-commit>` intersected with the t1653-owned list is empty.

## §B Decisions most likely to change (review these first)

1. Public contracts that other cards read: the status record, the v2 checkpoint, the `moai inbox` verb family (names and flags), and the lock contract. Fix these in M1/M2 before any Go code consumes them.
2. Whether the consumer stays shell (planner default) and the store verbs are Go: decision-index Q2.
3. v1 checkpoint adoption (Q1), `id` without a schema-version bump (Q3), where the PR1841 case is recorded (Q4), second backup copy outside the repository tree (Q5), preview before real cutover (Q6).

## §C Milestones (dependency order)

Delegation: `manager-develop` owns run phase, one milestone per spawn, TDD (RED first, then GREEN). Priority High for all; ordering is the dependency order.

| M | Scope | Language | Flips | Depends on |
|---|-------|----------|-------|------------|
| M1 | Consumer core: fixed-boundary snapshot, named states and `consumer-status.json`, persistent blocked state, v2 checkpoint with lineage binding, v1 adoption (never of a blocked checkpoint), distinct exit codes for `drain.sh`, the named seams of `design.md` §2.6; the jq filter program stays byte-identical | shell + jq | AC-LIS-001, 002, 004, 006 | none |
| M2 | Consumer commit and lock: commit manifest pair, same-dir temp, verify, flush, rename, restart recovery; lock owner record, dead-owner reap, `LOCK_HELD`; wrapper stays fail-open and now emits a notice on state change | shell + jq | AC-LIS-003, 005 | M1 |
| M3 | Identity and producer lock: new-row `id`, generation manifest builder, shared exclusive/shared lock plus spool with the final-merge handoff, lock and spool under the ignored `.moai/state/lessons-inbox/` with `TestTemplateGitignoreIgnoresLocalArtifacts` extended, rotation under the exclusive lock | Go (`internal/inboxsafe`, `internal/hook`) | AC-LIS-007, 008, 009, 010 | M2 contract text (lock order) |
| M4 | Backup, inventory, classification, bind primitive: `moai inbox backup`, `inventory`, `classify`, `checkpoint bind` | Go | AC-LIS-011, 012, 013, 020 | M3 |
| M5 | Proof and cutover: `prove` (evidence resolution and retention hold), `archive`, multiset equation, `cutover`, `rollback`, hash-chained ledger, forbidden-pattern guards | Go | AC-LIS-014, 015, 016, 017, 018, 019, 021, 024 | M3, M4 |
| M6 | Synthetic end-to-end through the existing applier and verify scripts | shell + Go | AC-LIS-022 | M5 |
| M7 | Real safe path, separately gated steps R1..R5 (below) | tooling run, no new code | AC-LIS-023 and the closing checks | M6 green, plan audit PASS, kickoff met |

### M7 real-path steps (each gated, each with its own evidence file under `.moai/reports/t1901/`)

| Step | Action | Gate to proceed | Evidence file |
|------|--------|-----------------|---------------|
| R0 | Confirm the lane is the sole maintenance writer; record the other consumers (session-start drains in other sessions also write `clusters.json`); build the Go binary from the tree and invoke it by path | leader confirmation recorded; build commit equals tree HEAD | `r0-preconditions.md` |
| R1 | Freeze: take the consumer lock, copy the complete-line prefix and the checkpoint, hash, read back, release | backup verified (REQ-LIS-013) | `r1-freeze.md` |
| R2 | Read-only inventory and classification of the frozen copy only | row count equals frozen line count | `r2-inventory.md` |
| R3 | Proof gate over the inventory | gate output file exists; counts per state | `r3-prove.md` |
| R3b | Record the pinned PR1841 item in the provenance ledger (requires t1900 and t1868), citing the re-hashed sha256 of the two case files; runs before any cutover and also on a zero-removal run | both hashes equal `spec.md` §1.2, else `evidence_missing`; the pin row exists | `r3b-pin.md` |
| R4 | Only if the removal set is non-empty: archive, then preview, then cutover (Q6) | archive verified; equation balanced; exclusive lock held | `r4-cutover.md` |
| R5 | Post-verify. No cutover: prefix property to the end (first N bytes of the live inbox equal the frozen copy), checkpoint byte-identical. After a cutover: multiset equation, new-generation hash against the manifest, archive count and hash against the ledger, checkpoint mapping or verbatim preservation. Both cases: ledger chain verifies | all green | `r5-post.md` |

Ordering witness (verification-claim-integrity §2.3): `.moai/reports/` and `.moai/state/` are gitignored, so they cannot carry the order. Each step refuses to start without the previous step's ledger row (mechanical), and after each step the lane commits one `progress.md` §E.2 line (step id, evidence-file sha256, ledger sequence number) in its own commit (tracked), so the commit graph orders the records.

If R3 proves zero eligible rows, R4 does not run and R5 records `zero_removal`. That is an acceptable, complete outcome of the card. If the real checkpoint is still beyond the file at cutover, it is preserved verbatim and the mismatch state stays reported (REQ-LIS-019).

## §D Resource discipline (run phase)

- Never `go test ./...`. Scoped, anchored selectors only, one exact top-level test per area with table subtests (for example `go test -v -run '^TestInboxSafe_Prove$' ./internal/inboxsafe/...`, `go test -v -run '^(TestInboxSafe_ProducerID|TestInboxSafe_RotationLock)$' ./internal/hook/...`, `go test -v -run '^TestInboxSafe_Bind$' ./internal/cli/...`); an unanchored selector can sweep a different set than intended.
- Heavy runs (the `internal/hook` and `internal/cli` packages) only inside a named slot: `moai slot acquire --resource go-test --max-duration <limit>`, then `moai slot release`.
- Shell tests (`drain_test.sh`, `session_drain_test.sh`, the new fixtures) are cheap and run unslotted.
- Cross-platform: new Go files that use file locks carry build tags; `GOOS=windows GOARCH=amd64 go build ./internal/...` for the touched packages.
- Worktree guard: the verb family resolves the project root itself and performs writes in-process; plain commands only in cells (no heredocs, no redirects, no compound commands). Whether an in-process write to the primary checkout path is accepted from a worktree session is not measured at plan time (Gap G3); R0 measures it on a harmless scratch root first.

## §E Risks

| Risk | Mitigation |
|------|-----------|
| The new consumer prints a mismatch notice on every session start for the real, still-mismatched checkpoint | Notice only on state change (status record dedupes); the status file always updates |
| Installed `moai` lags the tree | Verbs invoked by built path; the curator never depends on an installed build for consumption (Q2) |
| Adding `id` changes hook output shape | Additive field, readers ignore unknown fields; Q3 default keeps `v` at 1; AC-LIS-007 checks legacy readers |
| Producer lock adds latency to every failure hook | Shared lock, bounded wait, spool on timeout (REQ-LIS-010); the bound is measured in AC-LIS-009 |
| Real cutover hits a checkpoint that cannot be mapped | Preserved verbatim; no mapping invented |
| Two different files agree on the first line, byte offset, and last 8 consumed lines (undetected lineage swap) | Accepted residual risk: second-resolution timestamps make it negligible for these rows; the window width is a constant that can widen without a schema change |
| A synthetic proof passes while a real item is mis-proven | Gate re-resolves evidence mechanically; mutants in AC-LIS-014 and AC-LIS-024 |

## §F Pre-flight (run phase, before M1)

1. `git rev-parse --short HEAD` and branch (expect the card worktree branch); `git rev-list --count --left-right origin/main...HEAD` for the sync check.
2. Re-run `repro/conditions.sh 1..10` on the unmodified tree and compare with the RED ledger (`acceptance.md` §C); any difference means the base moved: re-measure, re-pin.
3. Record the sha256 of the `drain.sh` jq filter program text (RG-2).
4. Baseline `drain_test.sh` and `session_drain_test.sh` both `exit=0`.

## §G Anti-patterns

- Deciding an offset value, or adding a `--reset`, `--to-tail`, or `--force-zero` flag.
- Treating an exit code of 0 as the verdict where the status record says otherwise.
- Citing the base SHA as the closing check (the base-SHA form is only the labelled RED control; closing cells name the closing commit).
- Widening scope into the PermissionDenied advice or rotation policy.
- Hand-writing an inbox row to record a case (the mechanism writes rows, Q4).

## §H Cross-references

`spec.md` §2 requirements; `acceptance.md` criteria and evidence ledger; `design.md` component design and lock contract; `research.md` measurements; `decision-index.md` open decisions; `.claude/rules/moai/development/verification-completeness.md`; `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 (tool provenance) and §2.3 (ordering).
