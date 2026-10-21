# t1916 file-map survey (pre-implementation, read-only) — 2026-10-21T03:30Z

Directive: leader constraint message (03:26Z) items ① and ③ — measured, not inferred.
Tree: worktree `.moai/worktrees/t1916` @ `8249bff8c3` (branch `WT-factory-next-raced`), clean.

## A. Record-busy fix target map (this tree)

- `internal/cli/factory_card.go:469` — `var errFactoryRecordBusy = errors.New("the factory record stayed busy")`
- `internal/cli/factory_card.go:492-506` — `factoryClaimFailure`: marks claim-write errors as the busy-record outcome; its own comment (492-494) acknowledges a SQLite busy/locked error can mean host load ("never a lock race") yet still marks the busy outcome — the misclassification axis
- call sites: `1142` (`factoryNextRecordAndClaim`), `1170/1175/1203/1213` (`factoryNextClaim`), `1608` (`factoryNextNominate`, `case errors.Is(claimErr, errFactoryRecordBusy)`)
- busy-timeout DSN wiring context: `61`, `387`, `1510`
- refusal surface: `raced` refusal, exit 4 — test anchors `internal/cli/factory_lease_bound_test.go:109,151,298,443,464`
- status latency target: `internal/cli/factory_card.go:3335` (`newFactoryStatusCommand`), `3409` (`writeFactoryStatusText`) plus the read path behind them
- MCP relay of the same classification: `internal/cli/mcp_factory_card.go:123` (`handleFactoryNext`)

## B. PR #1859 measured state (overlap counterpart)

- state `OPEN`, head `8f5594f1b3`, updated `2026-10-21T03:25:23Z` (active — lane-11 implementing), 33 files
- code files: `internal/cli/factory_card.go`, `internal/cli/mcp_factory_card.go`, `internal/config/defaults.go`, `internal/config/types.go`
- test files: `factory_bundle_test.go`, `factory_card_eligible_pull_test.go`, `factory_lease_section_test.go`, `factory_m3_test.go`, `factory_nominate_test.go`, `factory_quota_test.go`, `factory_quota_worktrees_test.go`, `factory_self_dispatch_test.go`, `factory_serial_slot_stale_test.go`, `factory_t1533_test.go`, `factory_two_hub_cycle_test.go`, `review_observation_test.go`
- non-code: SPEC-CARD-ELIGIBLE-PULL-001 set, SPEC-FACTORY-RECORD-REPAIR-001/acceptance.md, SPEC-GLM-JEV-KEY-001, SPEC-PROGRESS-RECORD-IO-001, `CHANGELOG.md`, `workflow.yaml` (sections + template)

## C. Overlap verdict

- **FILE overlap: YES** — `factory_card.go` shared (and `mcp_factory_card.go` potentially, via the MCP classification relay)
- **Hunk-level today: no collision** — #1859's `factory_card.go` hunks sit in `factoryNextLeaseOnce` (old 378-431), `factoryCreateAndRenameWorktree` (829), `factoryNextSelectAndLease` (991-1044, 3 hunks), a +138-line insertion at 1288, and `newFactoryNextCommand` (1744+); my target regions (466-506, 1133-1220, 1608, 3335+) fall outside those old-base ranges
- **But the file is actively being reshaped** — #1859's line numbers are its own old base; its claim-path hunks (991-1044 reshaped, +138 inserted at 1288) bracket my claim-call-site region (1133-1220), so a same-file rebase is the expected integration shape after it lands
- **Directive ② disposition: implementation HOLD until #1859 lands** (leader coordinates order); read-only plan continues per directive ④

## D. Classification honesty (directive ③)

- The reissued card text labels this work `Class B` — a **label mismatch**: the operative contract (leader dispatches, 03:16Z reissue + 03:26Z constraint) is the original card scope WITH the SPEC binding — plan→run→sync, SPEC-FACTORY-LEASE-BUSY-001 authored in plan. No safety reduction vs the original t1743 dispatch; the session task list and this survey record the actual contract.
- File-separation rationale, measured: **t1903** (TUI SPEC) and **t1909** (codex removal) touch neither `factory_card.go` nor the lease-path tests — no coupling with this fix. **#1859** (self-dispatch gate · record SpecID eligibility · bundle predecessor) DOES share `factory_card.go` + `mcp_factory_card.go` — coupling is file-level, so sequencing (this card after #1859) is required, which is exactly what directive ② encodes.
- Contrast note (lane-observed, for the SPEC): the same lease path produces a clean structured refusal — verbatim `factory next: refused serial-slot: another serial card is in flight and holds the serial slot` (observed 03:10Z on t1743) — while record-busy outcomes get a single masked label; the misclassification branch is the defect.

## E. Watchdog / wait-record linkage

- watchdog snapshot: `.moai/state/watchdog/t1916.json` (first observation 03:25:29Z, lease renewed → 03:39:26Z)
- lane wait record: the primary-checkout copy froze at Update 2 (worktree isolation blocks editing shared-checkout paths); the worktree-local copy (`.moai/state/lane-60/wait-record-20261021.md`) is the live one from here
