# SPEC-HARNESS-DETACHED-PRUNE-001 — Plan

> Tier M. 8 REQ / 7 AC. Every code anchor below was measured on this tree at HEAD `d05d1d5f0` (branch `WT-harness-prune-detached`, clean). Line numbers are dated pointers — re-verify at run start, never trust across absorbs.

## §A Approach Summary

Four milestones in TDD RED→GREEN order. M1 removes the two synchronous prune calls from the Observer and updates the tests that coupled record to prune. M2 builds the spawn gate (one function in `internal/harness`) plus the platform-split detached-exec files and the hidden child verb. M3 wires the gate from the four hook observe handlers through one wrapper and lands the behavioral tests (gate suppression, fail-open, child prune, double-spawn collapse) against the injectable seam. M4 is verification and evidence. The decisions most likely to change (observer behavior, the new gate surface, the child argv) land first; the mechanical sweep and evidence close last.

## §B File Map

| # | File | Action | What |
|---|------|--------|------|
| 1 | `internal/harness/observer.go` | edit | remove the two `_ = o.retention.PruneStaleEntries(defaultRetentionDays)` blocks (:91-94, :139-142) and their comments; the `retention` field stays (REQ-DP-001 keeps the constructor and field — the gate needs the paths) |
| 2 | `internal/harness/retention.go` | edit | the `@MX:ANCHOR` caller comment at :62 updated when callers change (observer no longer calls; the gate and the child verb do) |
| 3 | `internal/harness/retention_spawn.go` | new | the gate primitive: `MaybeSpawnRetentionPruner(logPath string) error` — one lock-free stamp read reusing the `readStampFile`/`stampIsFresh` pair, stale-or-absent → invoke the injectable spawn seam, fresh → return nil; the seam is a package-level `var retentionSpawnFn func(exe string, args []string) error`-style field replaced by a recording fake in tests (the `ownerCheck` precedent, `retention.go:78-81`) |
| 4 | `internal/harness/retention_spawn_unix.go` | new | `//go:build !windows` — the detached spawn: `exec.Command` + `SysProcAttr{Setsid: true}`, no `Wait`, fail-open error return |
| 5 | `internal/harness/retention_spawn_windows.go` | new | `//go:build windows` — the flag combination REQ-DP-006 canonicalizes: detached, own process group, no console flash (`CREATE_NEW_PROCESS_GROUP` plus exactly one of `DETACHED_PROCESS`/`CREATE_NO_WINDOW`), finalized at run start per C4 against the pinned Go version's exported constants |
| 6 | `internal/harness/retention_spawn_test.go` | new | the gate tests: `TestMaybeSpawnRetentionPruner` (stale stamp → seam invoked once with the child argv; absent stamp → invoked; fresh stamp → not invoked — `TestSpawnGateSuppressesOnFreshStamp`), `TestSpawnFailureFailOpen` (seam returns error → gate returns it, caller logs at exit 0), `TestDetachedChildPrunes` (the real child entry driven in-process: the verb's run function against a temp log performs a real prune), `TestDetachedChildDoubleSpawnCollapses` (two children, one stamp file: the second finds the fresh stamp under the lock and exits without rewriting) |
| 7 | `internal/cli/hook.go` | edit | the hidden verb registration (beside the `harness-observe` family, :140-165) + ONE unexported record-then-gate wrapper; the four `obs.RecordExtendedEvent(evt)` call sites (:843, :933, :1071, :1250) each delegate to it |
| 8 | `internal/cli/hook_test.go` (or sibling) | edit | `TestHookRetentionPruneVerb` (the verb is registered, hidden from help, and performs a prune against a temp log) + `TestHarnessObserveGateWiring` (the wrapper is on the observe path: with a stale stamp fixture, one spawn is attempted through the seam; with a fresh stamp, none) |
| 9 | `internal/harness/observer_test.go` | edit | sweep for record-prune coupling: the `TestPruneStaleEntries*` family (:287-400) already prunes explicitly and survives unchanged; any test asserting prune-on-record is updated to prune explicitly through `Retention` (REQ-DP-001) |
| 10 | `internal/harness/integration_test.go` | edit | same sweep — the `RecordEvent` call sites (:59, :133, :171) record fresh events a 30-day prune would not touch, so coupling is unlikely; update only on observed coupling |

PRESERVE (untouched): `Retention.PruneStaleEntries`'s signature and body, `pruneExclusive`/`pruneLocked`/`openStateFile`/`healStateEntry`, `retention_open_unix.go`/`retention_open_windows.go`, `retention_heal_*.go`, the `retention_archive_atomic_test.go` suite (t1467 M1), the hook shell-script wiring (`.claude/hooks/**` — the change is inside the moai binary, no settings.json change), `.moai/config/sections/harness.yaml`, `internal/template/templates/**` (C3).

## §C Milestones

### M1 — observer stops pruning (RED: LEDGER-DP-A/B → GREEN)

Files 1, 2, 9, 10. Remove the two synchronous calls; update the anchor comment; sweep the two test files for record-prune coupling.

**Exit (count-first):** `go test -list '^(TestRecordExtendedEventDoesNotPrune|TestRecordEventDoesNotPrune)$' ./internal/harness` lists the new test, THEN `go test -run` it, exit 0; `grep -c "PruneStaleEntries" internal/harness/observer.go` → `0`; `go test ./internal/harness` exit 0 (the whole package is the re-verification set — F5's drift net); `go vet ./internal/harness` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

### M2 — spawn gate + detached-exec files + hidden verb (RED: LEDGER-DP-C/E/F/G → GREEN)

Files 3, 4, 5, 6 (gate half), 7 (verb registration half). The gate primitive with its injectable seam; the platform files; the hidden verb's run function calling `Retention.PruneStaleEntries` (never the internals directly — D5).

**Exit (count-first):** `go test -list '^(TestMaybeSpawnRetentionPruner|TestSpawnGateSuppressesOnFreshStamp|TestSpawnFailureFailOpen|TestDetachedChildPrunes|TestDetachedChildDoubleSpawnCollapses)$' ./internal/harness` lists exactly 5, THEN `go test -run` them, exit 0; `ls internal/harness/retention_spawn_unix.go internal/harness/retention_spawn_windows.go` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go vet ./internal/harness` exit 0; `go run ./cmd/moai hook retention-prune --help` exits 0 with the verb's help (the LEDGER-DP-G red flips when registration lands).

### M3 — CLI wiring + behavioral tests (RED: LEDGER-DP-C/D + cli zero-match → GREEN)

Files 6 (remaining), 7 (wrapper half), 8. The four handlers delegate to the wrapper; gate suppression, fail-open, and wiring tests.

**Exit (count-first):** `go test -list '^(TestHookRetentionPruneVerb|TestHarnessObserveGateWiring)$' ./internal/cli` lists exactly 2, THEN `go test -run` them, exit 0; `go test -race ./internal/harness -run '^(TestMaybeSpawnRetentionPruner|TestSpawnGateSuppressesOnFreshStamp|TestSpawnFailureFailOpen|TestDetachedChildPrunes|TestDetachedChildDoubleSpawnCollapses)$'` exit 0 (the seam and the double-spawn path are the concurrency surface); `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness internal/cli/hook.go` (excluding tests/comments) → 0 matches (B3 boundary); `go vet ./internal/cli ./internal/harness` exit 0.

### M4 — verification + evidence

RED→GREEN matrix for all 7 ACs recorded into `progress.md` §E.1 with verbatim outputs; narrow selectors only (no whole-repo `go test ./...` — lane discipline); `golangci-lint run` with the CI-pinned v2.1.6 version; `GOOS=windows GOARCH=amd64 go vet ./internal/harness ./internal/cli` exit 0; `go run ./cmd/moai spec lint SPEC-HARNESS-DETACHED-PRUNE-001 --strict` exit 0.

**Exit:** lint strict 0 findings; all AC GREEN cells carry command + verbatim output + exit code; `go test -cover ./internal/harness` ≥ 85% (TRUST 5 Tested); §E.1 `audit_ready: true`.

## §D Verification Plan

- Every acceptance criterion adopts the two-cell discipline (acceptance.md §A): a RED-now cell measured on `d05d1d5f0` and the GREEN command naming the milestone that flips it. GREEN commands carrying alternation live in acceptance.md's fenced ledger (LEDGER-DP-GREEN-A), not in table cells — the escaped form is a literal pipe in Go regexp (the iter1 D1 defect); LEDGER-DP-FORM carries the grep-BRE form-control that keeps M3's expected-zero boundary grep from reading as a dead pattern.
- Re-verification set per milestone: `go vet` on changed packages, `golangci-lint run` (CI-pinned v2.1.6), the targeted `-race` suite of M3, and `GOOS=windows GOARCH=amd64 go build ./...` + `go vet`.
- The gate's spawn seam is asserted by construction (fake injected, invocation recorded) — no test spawns a real child (REQ-DP-007); `TestDetachedChildPrunes` drives the verb's run function in-process, not a detached process.

## §E Risks and Orderings

- M1 before M2 (the observer's removal is the behavior contract the gate replaces; landing M2 first would leave two prune paths transiently).
- M2 before M3 (the wiring consumes the gate primitive and the verb).
- The platform files carry the C4 build-tag split (`//go:build !windows` / `//go:build windows`); the Windows creation-flags combination (file 5) is finalized at run start against the pinned Go version's `syscall` constants — REQ-DP-006 canonicalizes the semantic contract (detached, own process group, no console flash, one console-suppression mode), so the run-start check is constant availability, not a design change.
- The parent no-wait property (REQ-DP-005) deliberately carries no AC: no input on the pre-implementation tree turns it red (nothing spawns before M2 — LEDGER-DP-C), and REQ-DP-007 forecloses the runtime experiment that would observe reparenting; it is asserted by construction (no `Wait` call on the gate path) and declared as such in acceptance.md §D (iter1 D4 disposition).
- The `retention.go:62` anchor-comment update (file 2) must land in the SAME commit as the observer change (file 1) — a stale caller comment is the exact defect class the anchor exists to prevent.
