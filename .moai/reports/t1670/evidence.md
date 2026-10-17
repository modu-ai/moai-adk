# t1670 — local main atomic re-sync window (premise re-check → covered)

Lane: lane-17 · run tmvgvn · tree `.moai/worktrees/t1670` (branch WT-t1616-10-10, base local main 67e462e80) · leased 2026-10-17.
Card text: re-validation and fast-forward of the local-main re-sync in ONE factory lock section (t1616 F2), plus the failed-merge abort-ownership window (F1 residual). Completion: a test shows re-validation and move in the same lock section · external-actor contention test reproduces the window 0 times. Source: t1616 `sync-audit-789e472fc.md` F1·F2; accepted residuals listed again in `sync-audit-3f3504fb7.md` line 12 ("F2 (t1670), the F1 non-atomic abort window, local_main_resync.go:228 (t1670)").

## Premise check (verification-claim-integrity.md §1.1 surface 4 — the card's premise was re-measured before any plan work)

| Card premise | Tree at 67e462e80 | Covering change |
|---|---|---|
| F2 — re-validation not serialized with the fast-forward (`local_main_resync.go` had no mutation-lock call) | `internal/cli/local_main_resync.go:232-250`: `localMainResyncRevalidate` and `git merge --ff-only` both run inside `factory.RunIntegrationMutation(repoRoot, …)`; comment names "card t1619, codex review P1" | card t1619 |
| F2 test — re-validation and move in the same lock section | `internal/cli/local_main_resync_gap_test.go:16` `TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand` — a rival `--force` takeover attempted from inside the gap (seam `localMainResyncBeforeFastForward`) cannot land | card t1619 |
| F1 same-SHA — a MERGE_HEAD equal to the pin was treated as this call's own merge and aborted | `internal/factory/integration_merge_step.go:616-654`: foreign MERGE_HEAD refuses; a same-SHA MERGE_HEAD is aborted only when this call ran git and got exit 1 (`mergeCallDidNotStop`), otherwise refused without abort | card t1616 (later rounds) |
| F1 tests | `internal/factory/integration_merge_step_overlap_test.go:415,447,486,552` (`PreexistingMergeHeadRefusesWithoutAbort`, `RaceMergeHeadAfterProbeIsNotAborted`, `RaceSameShaMergeHeadAfterProbeIsNotAborted`, `StartFailureSameShaMergeIsNotAborted`) | card t1616 |

Measured by the lane in this tree (TMPDIR=GOTMPDIR=/private/tmp/rsync-tmp):

- `go test -count=1 -run 'TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand' ./internal/cli/` → `ok  github.com/modu-ai/moai-adk/internal/cli 9.649s`
- `go test -count=1 -v -run '<the four F1 tests>' ./internal/factory/` → `--- PASS` ×4 (0.33s, 0.42s, 0.52s, 0.42s), `ok  github.com/modu-ai/moai-adk/internal/factory 2.034s`

## What is NOT closable by code

The F1 "non-atomic abort window": git has no atomic merge-abort, so between this call's exit-1 merge and its `git merge --abort` an actor that does NOT take the integration flock can still touch the integration worktree. Every MoAI actor takes the lock (the abort runs inside the same `withIntegrationLockMutation` section as the merge), so the remaining window is reachable only by an actor outside the protocol. A test of "zero reproductions against an external actor" cannot be written against an actor that ignores the lock; the card's second completion clause is satisfiable only for lock-honouring actors, which the tests above already cover.

## Decision record (lane judgment, auto-semantics §6 terminal step)

decision record: decided_by=lane-17(orchestrator) evidence_refs=internal/cli/local_main_resync.go:232-250;internal/cli/local_main_resync_gap_test.go:16;internal/factory/integration_merge_step.go:616-654;internal/factory/integration_merge_step_overlap_test.go:415,447,486,552;test runs above ladder_path=①disk evidence → lane judgment

- The card's premise (F2 open, F1 same-SHA open) is false on the current tree; both were closed by cards t1619 and t1616 after this card was issued (2026-10-10).
- No code change. The card is landed as an evidence-only record so that the queue reflects the covered state; an independent sync audit re-verifies the coverage claim before landing.
- The non-atomic abort window stays an accepted residual (out-of-protocol actor only), as the t1616 audits already recorded.
