# SPEC-CODEX-LANE-SLOTS-001 — Acceptance Criteria

All ACs are mechanically verifiable. Primary re-measurement unit: `go test -timeout 30m ./internal/kanban/...`; targeted touched-package tests and builds where named. Every lane-env-reading test pins all the lane env axes it reads.

Classification discipline: **release-blocking** criteria carry an executed RED-now observation (§ Evidence Ledger below) and name the milestone that flips them green. **Regression-guard** criteria preserve already-correct behavior — no red is reproducible on the pre-implementation tree — and are never recorded as passes before their green lands (undecidable disposition).

## §D — AC Matrix

### AC-001 — Repro sequence: capacity-open run grows past the default bound (release-blocking)

**Given** a factory run in a `t.TempDir()` root holding a live `lane-1` claim recorded through the unbounded claude-shape join, **When** a bounded automatic claim in the codex-lane shape (auto=true, no requested label, bound 1) arrives for the same run, **Then** the claim succeeds with label `lane-2` — the refusal `factory run tm3yoq has no free lane slots in 1..1` no longer occurs.

- Verifies REQ-001, REQ-002. Flipped green by **M2**. RED-now: ledger **RED-1**.
- Command: `go test ./internal/kanban/... -run '^TestClaimFactoryLaneWithinGrowsCapacityOpen$'` exit 0.

### AC-002 — Explicit out-of-range request still refused (regression-guard)

**Given** the same bounded claim context with allowed range 1..1, **When** an explicit `lane-9` request arrives, **Then** the claim is refused with the out-of-range refusal naming the allowed range, and no growth is applied to explicit requests.

- Verifies REQ-003. Flipped green by **M2** (same tests as AC-001). No RED-now: the refusal is correct behavior today.
- Command: `go test ./internal/kanban/... -run '^TestClaimFactoryLaneWithinBounds$'` exit 0 (existing bounds test, assertion on the explicit-refusal case).

### AC-003 — Full explicitly-sized run still refuses (regression-guard)

**Given** a run whose recorded capacity is an explicit operator-declared count of 1 with `lane-1` live, **When** an automatic bounded claim arrives, **Then** the claim is refused with the no-free-slot refusal naming the declared range `1..1`.

- Verifies REQ-006 (t1294 contract preserved). Flipped green by **M2**. No RED-now: today every bounded run behaves as explicitly-sized, so the refusal exists.
- Command: `go test ./internal/kanban/... -run '^TestClaimFactoryLaneWithinExplicitCapacityFull$'` exit 0.

### AC-004 — Legacy-run refusal survives on the bounded path (regression-guard)

**Given** a live record carrying a legacy label (`worker-1`) stamped with a run id, **When** a bounded automatic claim arrives, **Then** the whole claim is refused with the legacy-run error naming the legacy value, the run id, and the retire step.

- Verifies REQ-009. Flipped green by **M4**. No RED-now: the refusal exists today (`TestClaimFactoryWorkerRefusesLiveLegacyClaim`, role_naming_m1_test.go:76, passes unmodified).
- Command: `go test ./internal/kanban/... -run '^(TestClaimFactoryLaneWithinRefusesLiveLegacy|TestClaimFactoryWorkerRefusesLiveLegacyClaim)$'` exit 0.

### AC-005 — Legacy label request refused; empty-run-id legacy ignored (regression-guard)

**Given** the current legacy-label semantics, **When** an explicit request uses a legacy spelling and when a legacy record carries an empty run id, **Then** the first is refused naming the canonical replacement and the second is neither refused nor counted — both unchanged.

- Verifies REQ-010, REQ-011. Flipped green by **M4**. No RED-now: unchanged behavior.
- Command: `go test ./internal/kanban/... -run '^(TestClaimFactoryWorkerDeadLegacyClaimIsStale|TestClaimFactoryWorkerIgnoresLegacyClaimOfNoRun|TestIsLegacyLeadLabelDetection)$'` exit 0.

### AC-006 — Run capacity recorded at leader start (release-blocking)

**Given** a leader start with an explicit size N and a leader start without one, **When** the run record is read back, **Then** the first records N and the second records the derived-capacity marker, and the codex lane join reads exactly what the record holds.

- Verifies REQ-004, REQ-005. Flipped green by **M1**. RED-now: ledger **RED-2**.
- Command: `go test ./internal/cli/... -run '^TestRecordFactoryRunStartCapacity$'` (new test in `internal/cli` — the recording seam is `recordFactoryRunStart`, readback through the same package's join path) exit 0.

### AC-007 — Owner probe bounded per listing invocation, platform variants pinned (release-blocking)

**Given** the owner-classification seam carrying a counting fake probe and a run table with multiple run rows, **When** `moai factory runs` classification executes, **Then** the per-listing probe work is bounded (batched, not one probe per row), the counting test asserts the invocation bound, and the test pins the platform-appropriate `platformProcessFingerprint` variant — the darwin variant on this tree, the unix (`ps`) and windows variants via their build-tagged test files.

- Verifies REQ-013. Flipped green by **M3**. RED-now: ledger **RED-4**.
- Command: `go test ./internal/homestate/... -run '^TestClassifyRunsBatchesOwnerProbe$'` (new instrumented test) exit 0; before/after `moai factory runs` wall-time recorded in the evidence path per REQ-014.

### AC-008 — Whole-package re-measurement green (regression-guard)

**Given** all milestones landed, **When** the owning package suite runs, **Then** it exits 0.

- Gate for the whole card; flipped green by **M4**. No RED-now: the suite is green on the base tree (6.633s measured by the auditor) — that green carries no defect-fixing information until the release-blocking greens land.
- Command: `go test -timeout 30m ./internal/kanban/...` exit 0; plus `go build ./...` and `go vet` on touched packages exit 0.

### AC-009 — Shared run-scoped lane pool across backends (release-blocking)

**Given** a run holding a live claim recorded through the unbounded claude-shape join (`lane-1`) and a bounded codex-shape join arriving after the growth fix, **When** both claims are live in the same run's workers set, **Then** both hold distinct numbers drawn from the one run-scoped pool, and the codex-shape claim followed the identical selection, growth, and refusal rules as the claude-shape claim (the claim engine records no backend and applies no backend-specific rule).

- Verifies REQ-007, REQ-008. Flipped green by **M2**. RED-now: ledger **RED-1** (shared with AC-001 — the observed refusal IS the pool-sharing failure: a codex-shape claim cannot draw a distinct number while a claude-shape claim holds `lane-1`).
- Command: `go test ./internal/kanban/... -run '^TestClaimFactoryLaneWithinSharedPoolDistinctNumbers$'` exit 0 (assertions on the mixed-shape fixture inside the new test).

### AC-010 — Per-step pre-exec timing report (release-blocking)

**Given** the codex lane launch path instrumented with per-step timing, **When** the timing report is emitted, **Then** it contains one timing line per pre-exec step, each naming the step (join gate, active-run resolution, codex pre-exec init, exec handoff at minimum) and its wall-time.

- Verifies REQ-012. Flipped green by **M3**. RED-now: ledger **RED-3**.
- Command: `go test ./internal/cli/... -run '^TestCodexLaneLaunchTimingNamesSteps$'` exit 0 (targeted launcher test asserting the report names each pre-exec step), plus the REQ-014 evidence: the slow-launch attribution measurement recorded with command and verbatim output.

## § Evidence Ledger (RED-now observations, executed on the pre-implementation tree)

Carrier note: fenced blocks, not table cells — shell metacharacters mangle inside table rows.

### RED-1 — AC-001 / AC-009: the root defect reproduces from the existing exported API

Committed plan-phase probe `redprobe_t1378_main.go` (this directory, `//go:build ignore`, invisible to `./...`) calling only exported `kanban.ClaimFactoryLane` (unbounded claude-shape) then `kanban.ClaimFactoryLaneWithin` (bounded codex-shape, 1..1) on one synthetic temp root it creates and removes, same run id `tm3yoq`. Re-executed at the iteration-3 tree:

```
$ go run .moai/specs/SPEC-CODEX-LANE-SLOTS-001/redprobe_t1378_main.go
RED-1 claude-shape claim: lane-1
RED-1 codex-shape claim error: factory run tm3yoq has no free lane slots in 1..1
exit code: 0 (probe completes; the red is the refusal verbatim on stdout)
```

Original observation (pre-implementation tree `850aefca9`, 2026-09-30, scratch form of the same probe, identical output): the claude-shape claim drew `lane-1` and the codex-shape claim refused with the same message, exit 0.

### RED-2 — AC-006: the runs record carries no lane-capacity datum

Same invocation as RED-1, second observation line:

```
RED-2 runs table columns: run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,created_at,updated_at
```

No capacity / lane-count column exists — the join bound has no run-state authority today. Original observation on tree `850aefca9` produced the identical column list.

### RED-3 — AC-010: no pre-exec timing instrumentation on the codex launch path

```
$ grep -rn "time.Since\|elapsed\|Elapsed" internal/cli/codex_factory.go internal/cli/codex_launcher.go internal/cli/factory.go
(exit code 1 — zero matches)
```

Positive control (the grep mechanism works; `time.Since` exists elsewhere in the package):

```
$ grep -rln "time.Since" internal/cli/*.go | head -3
internal/cli/chain.go
internal/cli/codex_audit_launch_test.go
internal/cli/codex_audit_mcp_test.go
(exit code 0)
```

### RED-4 — AC-007: the process-identity probe is per-pid across build-tagged variants

Single invocation, explicit file list, no pipe (re-executed at the iteration-3 tree):

```
$ grep -rn "platformProcessFingerprint" internal/homestate/process_fingerprint_unix.go internal/homestate/process_fingerprint_darwin.go internal/homestate/process_fingerprint_windows.go
internal/homestate/process_fingerprint_darwin.go:11:func platformProcessFingerprint(pid int) (string, bool) {
internal/homestate/process_fingerprint_unix.go:11:func platformProcessFingerprint(pid int) (string, bool) {
internal/homestate/process_fingerprint_windows.go:10:func platformProcessFingerprint(pid int) (string, bool) {
(exit code 0)
```

The seam takes a single pid; every classification of every run row pays one probe call. (The original `850aefca9` observation additionally showed the caller line `profile_lease.go:207` via a `*.go` glob; the de-piped form scopes to the three variant definitions, which is the load-bearing evidence.) Platform split for AC-007's variant pinning: `process_fingerprint_unix.go` is `//go:build !windows && !darwin` (one `ps -o lstart= -p <pid>` subprocess per probe), `process_fingerprint_darwin.go` is `//go:build darwin` (`unix.SysctlKinfoProc` per probe — no subprocess), `process_fingerprint_windows.go` is `//go:build windows`.

### Measurement attribution

- Original execution: 2026-09-30 on the pre-implementation worktree at tree SHA `850aefca9` (`git rev-parse --short HEAD` observed immediately before and after; nothing committed). The iteration-1 scratch probe directory was removed after capture.
- Re-execution (iteration 3, MP-8 re-executability): same day, tree SHA `4a7a2b0a4` observed immediately before and after; nothing committed. RED-1/RED-2 re-run through the committed probe (`go run .moai/specs/SPEC-CODEX-LANE-SLOTS-001/redprobe_t1378_main.go`, `//go:build ignore`, output byte-identical to the original); RED-4 re-run de-piped with an explicit file list. RED-3 is a read-only grep pair, unchanged by the re-anchor and still citable from the original execution.
- The probe runs against a SYNTHETIC state root only — a temp directory it creates (`os.MkdirTemp`) and removes; the live factory store (including run `tm3yoq`) is never read or written.

## §D.1 — Edge cases

- Concurrent bounded claims into a capacity-open run: exactly one claim per number, no duplicates (extends `TestClaimFactoryLaneWithinConcurrentOneSlot`'s shape to the grown range).
- Dead-claim pruning precedes growth: a dead `lane-2` claim frees its number; growth targets one past the highest **live** claim only.
- Derived-capacity marker never leaks into an explicit-capacity run through a join.

## §D.2 — Quality gates

- TRUST 5: package coverage maintained per `quality.yaml` targets; `gofmt` clean; `golangci-lint run` clean on touched packages.
- No `t.Setenv` with OTEL variables anywhere in the new tests; all temp dirs via `t.TempDir()`.

## §D.3 — Definition of Done

- All release-blocking ACs (001, 006, 007, 009, 010) green with their RED-now cells flipped, verbatim command output recorded in the run-phase evidence.
- All regression-guard ACs (002, 003, 004, 005, 008) green and never recorded as passes before the release-blocking greens land.
- The four plan-phase diagnosis findings (§B of plan.md) each carry a run-phase resolution note in progress evidence: attributed or explicitly handed to a follow-up card.
- Legacy-path tests pass without intent changes; the t1294 explicit-capacity contract holds.
