# SPEC-INBOX-DRAIN-GAP-001 — Progress

> Card t280 · Factory Mode lane-15 · Tier M · status: in-progress (run-phase M1 commit, 2026-09-02)

## §E.1 Plan-phase Audit-Ready Signal

_Pending plan-audit — plan-phase artifact set (spec.md / plan.md / acceptance.md / progress.md) created 2026-09-02 at worktree HEAD `131daa290`._

## §E.2 Run-phase Evidence

Run-phase evidence — manager-develop (lane-15), worktree `.claude/worktrees/t280`, branch `WT-inbox-drain-gap`, base `918bacd2c`. TDD discipline (RED→GREEN per milestone); mutant-teeth RED observations recorded per acceptance.md §A to `.moai/reports/t280/`.

### Milestone log

| Milestone | Scope | Tests | Evidence |
|---|---|---|---|
| M1 | Stub schema version field `v:1` + absence-tolerant reader (`InboxStubVersion`); cap constants pinned in `internal/config/defaults.go` (`DefaultInboxMaxBytes` 1 MiB, `DefaultInboxArchiveGenerations` 2) | RED build-failure observed (`m1-red.log`) → GREEN 7/7 PASS incl. pre-existing lessons-inbox baseline (`m1-green.log`) | `.moai/reports/t280/m1-red.log`, `.moai/reports/t280/m1-green.log` |
| M2 | `internal/hook/inbox_lifecycle.go`: write-time cap (`enforceInboxCap`, one stat steady-state NFC-2), shared rotation chain (`RotateLessonsInboxArchive`, delete-then-rename every link — NFC-5 Windows-safe, pre-era-idempotent), stand-down probe (`LSELDrainMarkerPresent`, ONE read-only os.Stat) | RED assertion-level 5 FAIL observed (`m2-red.log`) → GREEN 14/14 (`m2-green-final.log`) + `-race` ok (`m4-race.log`) | `.moai/reports/t280/m2-*.log` |
| M3 | `internal/cli/inbox.go`: `moai inbox status` (5 fields + regime token), `moai inbox drain` (shared rotation — no fork; refusal naming `session_drain.sh`), registered on rootCmd | RED build-failure (`m3-red.log`) → GREEN 6/6 (`m3-green.log`) | `.moai/reports/t280/m3-*.log` |
| M4 | Hardening + verification: `-race` formal, GOOS=windows vet+build, E4 scope guard, lint 0 issues, function-level coverage, boundary grep, consumer-vacuum re-check | all GREEN (see matrix below) | `.moai/reports/t280/m4-*.log` |

### AC matrix (E1 — all observed this run, tree `c543a0488`+M4, worktree `WT-inbox-drain-gap`)

| AC | Status | Verification command | Observed output |
|---|---|---|---|
| AC-IBX-001 | PASS | `go test ./internal/hook/ -run TestInboxCap_RotatesOverCapInbox -count=1 -v` | `--- PASS: TestInboxCap_RotatesOverCapInbox (0.00s)` (m2-green.log) |
| AC-IBX-002 | PASS | `go test ./internal/hook/ -run TestInboxStandDown_MarkerPresentNoRotation -count=1 -v` | `--- PASS: TestInboxStandDown_MarkerPresentNoRotation (0.00s)` (m2-green.log) — incl. lane-adopted state-dir byte-snapshot assert |
| AC-IBX-003 | PASS | `go test ./internal/hook/ -run TestInboxRetention_MaxTwoGenerations -count=1 -v` | `--- PASS: TestInboxRetention_MaxTwoGenerations (0.00s)` (m2-green.log) |
| AC-IBX-004 | PASS | `go test ./internal/cli/ -run TestInboxStatus -count=1 -v` | `--- PASS: TestInboxStatus_ReportsAllFields` + `--- PASS: TestInboxStatus_CuratorRegime` (m3-green.log) |
| AC-IBX-005 | PASS | `go test ./internal/cli/ -run TestInboxDrain_StandardInstallRotates -count=1 -v` | `--- PASS: TestInboxDrain_StandardInstallRotates (0.00s)` (m3-green.log) |
| AC-IBX-006 | PASS | `go test ./internal/cli/ -run TestInboxDrain_CuratorRefusal -count=1 -v` | `--- PASS: TestInboxDrain_CuratorRefusal (0.00s)` (m3-green.log) |
| AC-IBX-007 | PASS | `go test ./internal/hook/ -run 'TestLessonsInboxStub_GoldenJSONCarriesSchemaVersion\|TestInboxStubVersion_AbsenceReadsAsV1' -count=1 -v` | `--- PASS` ×2 (m1-green.log) |
| AC-IBX-008 | PASS | `go test ./internal/hook/ -run TestInboxCap_FailOpenOnRotationFailure -count=1 -v` | `--- PASS: TestInboxCap_FailOpenOnRotationFailure (0.00s)` (m2-green.log) |
| AC-IBX-009 | PASS | `git diff --name-only 918bacd2c..HEAD \| grep 'internal/cli/update/'` / `\| grep 'internal/template/'` | grep exit 1 both (0 matches) — full list in m4-scope-guard.log |
| AC-IBX-010 | PASS | `go test ./internal/hook/ -run TestInboxCap_ConcurrentAppendsCrossCapBoundary -count=1 -race` | `ok github.com/modu-ai/moai-adk/internal/hook 1.671s` (m4-race.log) |

### Mutant teeth (acceptance §A — every contract observed RED under its scratch mutation, then reverted)

| Contract | Mutation | RED observation | Evidence |
|---|---|---|---|
| AC-IBX-007 (marshal) | `Version` field omitted from stub literal | `--- FAIL ... version = 0, want 1` | m1-mutant1-version-drop.log |
| AC-IBX-007 (reader) | absence default `1` → `0` | `--- FAIL ... InboxStubVersion(pre-upgrade line) = 0, want 1` | m1-mutant2-reader-default.log |
| AC-IBX-001 | `enforceInboxCap` no-op | `--- FAIL ... archive generation .1 not created` | m2-mutant1-cap-noop.log |
| §B boundary edge | cap check `>=` → `>` (exclusive) | `--- FAIL ... rotation did not fire at the exact-cap boundary` | m2-mutant2-boundary.log |
| AC-IBX-002 / NFC-4 | stand-down branch disabled | `--- FAIL ... stand-down violated: 1 archive generations created` | m2-mutant3-standdown-off.log |
| AC-IBX-003 | `DefaultInboxArchiveGenerations` 2 → 3 | `--- FAIL ... 3 archive generations (bound 2)` | m2-mutant4-retention.log |
| AC-IBX-008 | rotation shift swallows real errors | `--- FAIL ... not logged as a warning` + `live file was rotated despite sabotaged chain` | m2-mutant5-failopen-swallow.log |
| NFC-4 (marker creation) | probe adds `os.MkdirAll(state/lsel)` | `--- FAIL ... cap path created .moai/state` | m2-mutant6-nfc4-mkdir.log |
| AC-IBX-004 | `cap_distance_bytes` field dropped | `--- FAIL ... status output missing "cap_distance_bytes:"` | m3-mutant1-status-field.log |
| AC-IBX-005 | drain never rotates | `--- FAIL ... drain did not rotate` | m3-mutant2-drain-skip.log |
| AC-IBX-006 | refusal branch disabled | `--- FAIL ... drain on a curator machine must exit non-zero` | m3-mutant3-refusal-off.log |

All mutations reverted before commit; `grep -rn MUTANT internal/hook/ internal/cli/ internal/config/` over owned files = 0 (remaining hits are graph_refresh_test.go, another SPEC's naming).

### Lane-adopted NFC-4 strengthening (plan-audit N-D2 — recorded, NOT an AC formula change)

No current AC mechanically checks the read-only discipline on `.moai/state/lsel/`, so the lane added inside the AC-IBX-002 test implementation: (a) a recursive byte-snapshot (path + mode + size + sha256 content) of the state dir before vs after over-cap capped appends, asserted `reflect.DeepEqual`; (b) an assertion that the cap path never CREATES `.moai/state` on marker-absent installs (creating it would stand the cap down forever after). Teeth proven by m2-mutant6 (snapshot catches creation) and m2-mutant3 (stand-down branch). Residual: a PURE READ of a file under the state dir would not change the snapshot — excluded by the implementation contract (one `os.Stat`, no other call in `LSELDrainMarkerPresent`, auditable in `internal/hook/inbox_lifecycle.go`), not by the snapshot.

### E2–E7 evidence summary

- **E2 (local part)**: `GOOS=windows GOARCH=amd64 go vet ./internal/hook/ ./internal/cli/ ./internal/config/` exit 0; `GOOS=windows ... go build` same → m4-windows-vet.log. NFC-5 final verdict remains the CI windows matrix — NOT claimed locally.
- **E3**: cross-package profile (`-coverpkg=./internal/hook` over hook+cli inbox batches): `RotateLessonsInboxArchive` 88.9%, `enforceInboxCap` 100%, `LSELDrainMarkerPresent` 100%, `LessonsInboxPath` 100%, `LessonsInboxArchiveGens` 100%, `InboxStubVersion` 100% (m4-coverage-cross.log + m4-coverage.log). New code ≥ 85% standard at function level. The bare `3.8%` package figure is a `-run`-filtered artifact, not the package's coverage — cited nowhere as a package baseline.
- **E4**: m4-scope-guard.log — changed files are exactly: 4 Go source + 2 Go test + root.go + defaults.go + SPEC artifacts + evidence logs. Zero `internal/cli/update/`, zero template tree.
- **E5**: m4-race.log — single + wide inbox batch under `-race`, both `ok`.
- **E6**: AC-IBX-002 doubles as the NFC-4 parity proof (zero archives + byte-snapshot equality + unbounded live growth under marker).
- **E7**: `golangci-lint run --timeout=2m ./internal/hook/... ./internal/cli/... ./internal/config/...` → `0 issues.` (m4-lint.log).
- **C-HRA-008**: `grep -rn 'AskUserQuestion\|mcp__askuser'` over owned files → 0 matches.
- **Consumer-vacuum premise re-check (acceptance §D.3)**: `grep -rln lessons-inbox internal/ --include='*.go' | grep -v _test.go` → only this SPEC's own files (inbox.go, inbox_lifecycle.go, failure_observer.go, defaults.go comment). No shipped Go consumer. Premise holds.

### Notes for sync-phase handoff

- **N-D1 advisory (NOT fixed — acceptance.md body is out of run-phase scope)**: acceptance.md §B first edge cites "the t259 figures in spec.md §B"; the measured 1.1MB / ~4.2k-line figures actually live at plan.md §B (lines 23, 56). Cross-reference label only; content is correct.
- **Behavior note**: a one-shot `moai inbox drain` moves live → `.1` and does NOT fabricate a fresh live file; the collector's next append recreates it (O_CREATE). `moai inbox status` handles the absent inbox (size 0, lines 0). AC-IBX-005 requires no fresh live file.
- **Verification discipline**: per lane contract, ALL package runs used targeted `-run` filters (internal/cli is a 600s-floor package — bare package runs prohibited). Full-suite verdict is CI's.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-03
run_commit_sha: "6e55a6166" # backfilled per the D3 placeholder pattern — terminal run-phase commit (M4)
run_status: complete
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "n/a — lane does not push; develop push is the lead's batch concern"
l44_post_push_fetch: "n/a — same; remote landing verified by lead after batch push"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  macos: "local darwin green (hook/cli/config targeted batches, m4-*-batch.log)"
  windows: "local GOOS=windows vet+build exit 0 (m4-windows-vet.log); CI windows matrix is the NFC-5 final verdict"
total_run_phase_files: 7
m1_to_m4_commit_strategy: "4 milestone commits (M1..M3 landed, M4 this) + 1 run_commit_sha backfill commit"
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
