# SPEC-CODEX-LANE-SLOTS-001 — Acceptance Criteria

All ACs are mechanically verifiable. Primary re-measurement unit: `go test -timeout 30m ./internal/kanban/...`; targeted touched-package tests and builds where named. Every lane-env-reading test pins all the lane env axes it reads.

## §D — AC Matrix

### AC-001 — Repro sequence: capacity-open run grows past the default bound

**Given** a factory run in a `t.TempDir()` root whose join context reads capacity-open (no operator-declared size) and an explicit-size parameter of 1 on the bounded claim, **When** a live claim already holds `lane-1` (pid kept alive by the test's alive probe) and a second automatic bounded claim arrives in the codex-lane shape (auto=true, no requested label), **Then** the claim succeeds with label `lane-2`, and the recorded run's workers show both live claims with distinct numbers.

- Verifies REQ-001, REQ-002; reproduces the operator sequence (`factory run tm3yoq has no free lane slots in 1..1` no longer occurs).
- Command: `go test ./internal/kanban/... -run TestClaimFactoryLaneWithin` (new test) exit 0.

### AC-002 — Explicit out-of-range request still refused

**Given** the same bounded claim context with allowed range 1..1, **When** an explicit `lane-9` request arrives, **Then** the claim is refused with the out-of-range refusal naming the allowed range, and no growth is applied to explicit requests.

- Verifies REQ-003.
- Command: `go test ./internal/kanban/... -run TestClaimFactoryLaneWithin` exit 0 (assertion inside the existing/new bounds test).

### AC-003 — Full explicitly-sized run still refuses (t1294 contract preserved)

**Given** a run whose recorded capacity is an explicit operator-declared count of 1 with `lane-1` live, **When** an automatic bounded claim arrives, **Then** the claim is refused with the no-free-slot refusal naming the declared range `1..1`.

- Verifies REQ-006.
- Command: `go test ./internal/kanban/... -run TestClaimFactoryLaneWithin` exit 0.

### AC-004 — Legacy-run refusal survives on the bounded path

**Given** a live record carrying a legacy label (`worker-1`) stamped with a run id, **When** a bounded automatic claim arrives, **Then** the whole claim is refused with the legacy-run error naming the legacy value, the run id, and the retire step.

- Verifies REQ-009.
- Command: `go test ./internal/kanban/... -run 'TestClaimFactory'` exit 0; the existing `TestClaimFactoryWorkerRefusesLiveLegacyClaim` (role_naming_m1_test.go:76) passes unmodified.

### AC-005 — Legacy label request refused; empty-run-id legacy ignored

**Given** the current legacy-label semantics, **When** an explicit request uses a legacy spelling and when a legacy record carries an empty run id, **Then** the first is refused naming the canonical replacement and the second is neither refused nor counted — both unchanged.

- Verifies REQ-010, REQ-011.
- Command: `go test ./internal/kanban/... -run 'TestClaimFactoryWorker|TestIsLegacyLeadLabel'` exit 0 with the existing tests unmodified in intent.

### AC-006 — Run capacity recorded at leader start

**Given** a leader start with an explicit size N and a leader start without one, **When** the run record is read back, **Then** the first records N and the second records the derived-capacity marker, and the codex lane join reads exactly what the record holds.

- Verifies REQ-004, REQ-005.
- Command: targeted test in the touched package (`internal/cli` or `internal/homestate`) exit 0, e.g. `go test ./internal/cli/... -run TestRecordFactoryRunStart` (new).

### AC-007 — Owner probe bounded per listing invocation

**Given** the owner-classification seam carrying a counting fake probe and a run table with multiple run rows, **When** `moai factory runs` classification executes, **Then** the real probe performs a bounded number of subprocess invocations per listing (batched, not one-per-row), and the counting test asserts the invocation bound.

- Verifies REQ-013.
- Command: `go test ./internal/homestate/... -run TestClassifyRuns` (new instrumented test) exit 0; before/after `moai factory runs` wall-time recorded in the evidence path per REQ-014.

### AC-008 — Whole-package re-measurement green

**Given** all milestones landed, **When** the owning package suite runs, **Then** it exits 0.

- Command: `go test -timeout 30m ./internal/kanban/...` exit 0; plus `go build ./...` and `go vet` on touched packages exit 0.

## §D.1 — Edge cases

- Concurrent bounded claims into a capacity-open run: exactly one claim per number, no duplicates (extends `TestClaimFactoryLaneWithinConcurrentOneSlot`'s shape to the grown range).
- Dead-claim pruning precedes growth: a dead `lane-2` claim frees its number; growth targets one past the highest **live** claim only.
- Derived-capacity marker never leaks into an explicit-capacity run through a join.

## §D.2 — Quality gates

- TRUST 5: package coverage maintained per `quality.yaml` targets; `gofmt` clean; `golangci-lint run` clean on touched packages.
- No `t.Setenv` with OTEL variables anywhere in the new tests; all temp dirs via `t.TempDir()`.

## §D.3 — Definition of Done

- AC-001..AC-008 all green with verbatim command output recorded in the run-phase evidence.
- The four plan-phase diagnosis findings (§B of plan.md) each carry a run-phase resolution note in progress evidence: attributed or explicitly handed to a follow-up card.
- Legacy-path tests pass without intent changes; the t1294 explicit-capacity contract holds.
