# Card t1516 — Verdict: `factory next` excludes assigned rows of hold/queued queue items

- Branch: `WT-hold-lease-exclude` (worktree `agent-a4dd65a4b500c1203`)
- Commits: `6ff7074de` (fix), `1dea55194` (guard re-pin), `c9aecc230` (fixture re-pins); base `d05d1d5f0`
- Requirement source: the leader's dispatch text (card t1516), declared canon per leader ruling 2026-10-05 (see Gaps — the dispatch's cited evidence file is absent from disk)

## Claim

1. Bare `moai factory next` does NOT lease an `assigned` record row whose owning queue item's current state is `hold` or `queued`: arm (a) skips the row (exit 3, `no card is available`, when nothing else remains).
2. `moai factory next --card <id>` on the same shapes refuses: queue state `hold` → `held` (pre-existing state switch, now pinned with a placed row); queue state `queued` + assigned row → `recorded` (new keep-set clause; refusal decided before any write — no promotion, no record write, no lease).
3. A queued item with an own assigned row under bare `next` reaches a lease only through arm (c)'s promotion, which flips the item to `picked` first — the row is never leased while the item is queued.
4. The mid-flight self-resume edge (a lane re-leasing its own assigned card whose queue item is `picked`) is unchanged and pinned.
5. Three guards that pinned the superseded behavior are re-pinned to the new rule (AC-FAL-003 clause (iii) guard, AC-FAL-009 (ii) table's `nominated-queued-assigned` row, AC-QAS-010 `assigned_card_leased_during_wait` fixture).

## Evidence

RED (before the fix; `go test ./internal/cli/ -run 'TestFactoryNextArmASkipsHoldQueueItem|TestFactoryNextArmAQueuedQueueItemPromotesFirst|TestFactoryNextArmALeasesOwnAssignedPickedCard|TestFactoryNextNominateRefusesKeepSet' -count=1`), verbatim:

```
--- FAIL: TestFactoryNextNominateRefusesKeepSet (41.98s)
    --- FAIL: TestFactoryNextNominateRefusesKeepSet/queued-assigned (4.73s)
        factory_nominate_test.go:624: the nomination was not refused (want exit 4, token "recorded"): stdout="moai: worktree commits under global git identity F1 Test <f1@example.invalid> (read-only from ~/.gitconfig)\nt1 stage=run worktree=t1\nt1\tunknown\t\t\tpicked\t\tfactory card 1\n" stderr="note: open pull requests unavailable (exit status 1); link column left empty, landed check still ran\nnote: the landed check against origin/main could not answer for t1; those cards report unknown rather than no-link, because an unanswerable query is not evidence of not-landed\n"
--- FAIL: TestFactoryNextArmASkipsHoldQueueItem (2.68s)
    factory_nominate_test.go:686: an assigned row under a held queue item: expected an error, got nil
--- FAIL: TestFactoryNextArmAQueuedQueueItemPromotesFirst (4.31s)
    factory_nominate_test.go:717: t1 queue state = queued, want picked (promoted before the lease, never leased while queued)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	53.947s
```

(`held-assigned` and `TestFactoryNextArmALeasesOwnAssignedPickedCard` were green at RED — they pin already-correct behavior.)

GREEN (after the fix + re-pins), each observed in this run, this tree:

- Same selector set, `-count=1`: `ok github.com/modu-ai/moai-adk/internal/cli 56.729s`
- Amended family (`TestFactoryLeaseArmAExcludesHeldAssignedCard|TestFactoryNextArmA*|TestFactoryNextNominateRefusesKeepSet|TestFactoryNextNominateRefusalLeavesStateUnchanged`): `ok … 169.499s`
- AC-FAL-009 (ii) table (`TestFactoryLeaseSectionRecordWritesPerArm`): `ok … 44.181s` (after fixing my own amendment's row-presence assertion to an unchanged-row assertion)
- Full nomination/lease family, env-scrubbed (`-run 'TestFactoryNext|TestFactoryLease|TestQAS_'`): only `TestFactoryLeaseQueueLockStallBounded` + `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` fail — both classified pre-existing load flakes (below)
- `-race` on the same family, `-count=1 -timeout 30m`: zero `DATA RACE` lines; only `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` (the flake; 793.8ms / 1.926s vs the 500ms-after-release bound) — `TestFactoryLeaseQueueLockStallBounded` passed under `-race`, re-confirming flake classification
- Static: `go vet ./internal/cli/...` clean; `golangci-lint run ./internal/cli/...` → `0 issues.` (v2.1.6, go1.26.8, installed build); `go build ./...` OK; `GOOS=windows GOARCH=amd64 go build ./...` OK; `gofmt -l internal/cli/` empty

## Baseline-attribution

- Every result above was measured in this run against this tree: worktree `agent-a4dd65a4b500c1203`, branch `WT-hold-lease-exclude`, commits `6ff7074de`/`1dea55194`/`c9aecc230` on base `d05d1d5f0` (origin/develop tip carrying the t1525/t1526 merges).
- Pre-existing-flake attribution used a control tree: `git archive d05d1d5f0 | tar -x -C /tmp/t1516-baseline` (no worktree created), then the same selectors with the same env scrub:
  - `TestFactoryLeaseQueueLockStallBounded` fails on the scrubbed baseline too (`--- FAIL` nominated + bare, `/tmp/t1516-baseline-scrubbed3.log`).
  - `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits -count=3` fails on the scrubbed baseline identically (594.295ms / 562.352ms / 549.267ms vs my tree's 596.058ms, all over the 500ms `flMargin`), and passed on the baseline solo in a quieter window (`ok … 9.570s`). The fixture carries no assigned rows, so the t1516 gate code provably does not execute on its path.
- Unscrubbed runs were polluted by this lane session's ambient env (`MOAI_FACTORY_ROLE=lane`, `MOAI_FACTORY_WORKER=lane-27` observed in `env`): 21 additional tests failed with `todo add: moai add: refused — lane boundary: a lane session cannot mutate the queue …`. Root cause observed: `factoryLaneAdmission()` reads `config.EnvFactoryRole`, but `factoryAmbientEnvKeys` (internal/cli/factory_test.go) does not list it, so TestMain's ambient clear misses it. All 21 pass under the scrub (`unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER … && go test …`, single compound invocation).
- Tool provenance: golangci-lint is an installed third-party build (released artifact, no comparable repo coordinate for this tree — §2.2 carve-out); go vet/build/`-race` ran from the tree's own toolchain.

## Gaps

1. **The dispatch's cited evidence file is absent from disk.** The dispatch text cited `t1480/sync-audit-3.md` as evidence. In this tree `.moai/reports/t1480/` does not exist at all; the existing `.moai/reports/t1481/sync-audit-3.md` is card t1481's SPEC-FACTORY-DECISION-AUTO-001 sync re-audit (single H2 "Evaluation Report (sync 재감사 3회차, 수렴 규칙 적용)"; its findings are decision-gate N1/N2/C1 and a backlog.db/factory.db hold-race debt item — nothing about factory-next leasing assigned rows). Per the leader's ruling, the dispatch text itself is the canonical requirement; this verdict verifies against that text.
2. **Two pre-existing timing flakes were not driven to green here.** `TestFactoryLeaseQueueLockStallBounded` (3.3s lock-wait budget + 500ms margin) and `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` (500ms-after-release bound) fail on the base commit and on this branch alike under the machine's current multi-lane load, and pass in quiet windows. Classified pre-existing flake, not NEW; no code or test change made to them (out of card scope). CI owns the full-suite verdict.
3. **The 21 lane-env false reds are a harness gap, not fixed here.** `factoryAmbientEnvKeys` missing `config.EnvFactoryRole` will keep biting any lane session running `go test ./internal/cli/` locally. Adding the key is a one-line harness follow-up for the leader; out of t1516's scope (queue-mutation surface untouched).
4. **SPEC artifacts still cite the renamed guard.** `.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/{acceptance.md,plan.md,progress.md}` reference `TestFactoryLeaseArmAKeepsHeldAssignedCard` in `-run` commands (acceptance.md:603,779) and as mutant guard MU18. After the rename those patterns silently select one fewer test. Amending the SPEC artifacts is the leader's docs follow-up (this card is dispatch-canon; SPEC bodies are phase-owned).
5. **No push, no CI.** Per lane discipline the branch is unpushed; the full-suite and platform-matrix verdicts remain CI's.

## Residual-risk

- The pinned promote-then-lease path (queued item + own assigned row under bare `next`) depends on arm (c)'s pre-existing behavior of promoting queued items that already carry record rows (the claim then reports `raced`, and the next attempt's arm (a) leases the now-picked item). A future refactor that makes arm (c) skip items with record rows will flip that pinned outcome — the pin forces that decision to be conscious.
- Only leasing is gated. `unpick`/`hold` still leave assigned record rows behind; an operator re-pick (or `unhold` + pick) restores leaseability by design. The stale rows themselves are not cleaned up.
- The two timing flakes will keep failing intermittently on loaded machines and can mask future real regressions in this family; the documented disposition reduces, but does not remove, that cost.
