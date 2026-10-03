# t1432 — observed RED baseline for the heal lock (M7, commit T2)

SPEC-HARNESS-RETENTION-HARDEN-001 v0.4.1, plan.md §F.1 M7. Run-phase takeover from lane-12.

## Tree

- Measured tree: the working tree of the takeover worktree at HEAD `f040ffcd3` (merge of
  `WT-harness-retention-debt` @ `09a701ad2` into local `develop` @ `02f939b14`) plus the two
  uncommitted test edits this commit adds (`internal/harness/retention_heallock_test.go` new,
  `internal/harness/retention_owner_test.go` B11 guard). No production file differs from
  `7639c04c1`: `git diff --quiet 7639c04c1 HEAD -- internal/harness/retention.go internal/harness/observer.go internal/lockfile`
  exit 0 (observed).
- Slot lease `go-test-internal-harness` held for the package run.
- Deviation: the kanban environment was NOT scrubbed for the package run. The worktree-isolation
  guard refused every compound form (`unset ... && go test ...`, `env -u ... go test ...`) as
  "too complex to verify"; the plain `go test` form was the one it admitted. `grep -rl MOAI_KANBAN internal/harness/`
  printed nothing, so no test in this package reads those variables; recorded as a Gap, not hidden.

## Build

- `go vet ./internal/harness/` exit 0 (`VET_OK`).
- `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0 (`WINVET_OK`).
- `gofmt -l internal/harness/` printed nothing.

## RED run

Command: `go test -count=1 -v -timeout 300s -run "TestPrune[HCS]" ./internal/harness/` (50 `=== RUN` lines swept), exit 1, `FAIL github.com/modu-ai/moai-adk/internal/harness 2.187s`.

| Criterion | Test | Expected at base | Observed |
|---|---|---|---|
| AC-HRH-006 (b) | `TestPruneHealSerializesOnTheHealLock` | FAIL (E-036) | `--- FAIL` |
| AC-HRH-006 (b2) | `TestPruneHealWaitsForASharedHolder` | FAIL (E-047a) | `--- FAIL` |
| AC-HRH-015 | `TestPruneHealLockHeldPastTheBoundFailsClosed` | FAIL (E-037) | `--- FAIL` |
| AC-HRH-016 | `TestPruneHealLockHostileEntryFailsClosed` (4 subtests) | FAIL (E-038) | `--- FAIL` symlink, directory, fifo, not-owned |
| AC-HRH-006 (c) | `TestPruneCommonPathCreatesNoHealLock` (absent, healthy) | PASS (E-039) | `--- PASS` both |
| AC-HRH-006 (c) held | `TestPruneCommonPathIgnoresAHeldHealLock` | PASS (vacuous) | `--- PASS` |
| AC-HRH-006 (d) | `TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep` | PASS (vacuous) | `--- PASS` |
| AC-HRH-003 (c) | `TestPruneStateRemovalFailureInReadOnlyDirSkips` | PASS (E-048a) | `--- PASS` |
| AC-HRH-003 (b) | `TestPruneStateUnreplaceableInReadOnlyDirSkips` | PASS | `--- PASS` |
| swap test (B11 guard) | `TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal` | PASS | `--- PASS` |

Every RED cell matches its ledger cell (E-036 to E-038, E-047a); no Gap from a mismatch.

Verbatim failure lines (temporary-directory prefix `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/` shortened to `<tmp>/`):

```
=== RUN   TestPruneHealLockHeldPastTheBoundFailsClosed
    retention_heallock_test.go:201: want an error naming <tmp>/TestPruneHealLockHeldPastTheBoundFailsClosed527238699/001/usage-log.jsonl.prune-heal, got <nil>
    retention_heallock_test.go:204: the state-path entry was removed or replaced although the heal lock was not acquired: err=<nil>
    retention_heallock_test.go:210: log changed
    retention_heallock_test.go:213: archive directory exists: <nil>
    retention_heallock_test.go:216: want exactly one warning line naming <tmp>/TestPruneHealLockHeldPastTheBoundFailsClosed527238699/001/usage-log.jsonl.prune-heal, got ""
=== RUN   TestPruneHealLockHostileEntryFailsClosed/symlink
    retention_heallock_test.go:302: want an error naming <tmp>/TestPruneHealLockHostileEntryFailsClosedsymlink2838471276/001/usage-log.jsonl.prune-heal, got <nil>
    retention_heallock_test.go:305: the state-path entry was removed or replaced although the heal lock was unusable: err=<nil>
    retention_heallock_test.go:311: log changed
    retention_heallock_test.go:314: want exactly one warning line naming <tmp>/TestPruneHealLockHostileEntryFailsClosedsymlink2838471276/001/usage-log.jsonl.prune-heal, got ""
=== RUN   TestPruneHealLockHostileEntryFailsClosed/directory
    retention_heallock_test.go:302: want an error naming <tmp>/TestPruneHealLockHostileEntryFailsCloseddirectory751617534/001/usage-log.jsonl.prune-heal, got <nil>
    retention_heallock_test.go:305: the state-path entry was removed or replaced although the heal lock was unusable: err=<nil>
    retention_heallock_test.go:311: log changed
    retention_heallock_test.go:314: want exactly one warning line naming <tmp>/TestPruneHealLockHostileEntryFailsCloseddirectory751617534/001/usage-log.jsonl.prune-heal, got ""
=== RUN   TestPruneHealLockHostileEntryFailsClosed/fifo
    retention_heallock_test.go:302: want an error naming <tmp>/TestPruneHealLockHostileEntryFailsClosedfifo2350474361/001/usage-log.jsonl.prune-heal, got <nil>
    retention_heallock_test.go:305: the state-path entry was removed or replaced although the heal lock was unusable: err=<nil>
    retention_heallock_test.go:311: log changed
    retention_heallock_test.go:314: want exactly one warning line naming <tmp>/TestPruneHealLockHostileEntryFailsClosedfifo2350474361/001/usage-log.jsonl.prune-heal, got ""
=== RUN   TestPruneHealLockHostileEntryFailsClosed/not-owned
    retention_heallock_test.go:302: want an error naming <tmp>/TestPruneHealLockHostileEntryFailsClosednot-owned2911238329/001/usage-log.jsonl.prune-heal, got <nil>
    retention_heallock_test.go:305: the state-path entry was removed or replaced although the heal lock was unusable: err=<nil>
    retention_heallock_test.go:311: log changed
    retention_heallock_test.go:314: want exactly one warning line naming <tmp>/TestPruneHealLockHostileEntryFailsClosednot-owned2911238329/001/usage-log.jsonl.prune-heal, got ""
=== NAME  TestPruneHealWaitsForASharedHolder
    retention_heallock_test.go:426: the pruner returned (<nil>) while another descriptor held a shared heal lock: it did not request an exclusive lock
    retention_heallock_test.go:430: the state-path entry was removed or replaced while a shared heal lock was held: err=<nil>
=== NAME  TestPruneHealSerializesOnTheHealLock
    retention_heallock_test.go:116: the pruner returned (<nil>) while another descriptor held the heal lock: it did not wait
    retention_heallock_test.go:120: the state-path entry was removed or replaced while the heal lock was held by another descriptor: err=<nil>
```

## Sentinel and ignore baselines

- E-041 (ninth sentinel): `grep -c -F "heal lock gives no exclusion on Windows" internal/harness/retention.go` printed `0`.
- E-052 (negative sentinel): `grep -c -F "heal window remains" internal/harness/retention.go` printed `0`, exit 1.
- E-042: `git check-ignore -v .moai/harness/usage-log.jsonl.prune-heal` exit 1, empty output.

## Ordering (Definition of Done 9)

At M7 the bounded listing `7639c04c1..HEAD -- internal/harness/retention.go internal/harness/retention_heal_unix.go internal/harness/retention_heal_windows.go` is expected empty; it is evaluated at M10 and an empty listing is a Gap, never a pass.
