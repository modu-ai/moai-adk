# Acceptance — SPEC-FACTORY-LEASE-BUSY-001

Stateless artifact (no `status:` field). Every criterion is binary-testable and independently
executable. The implementation milestones are gated behind plan.md §F M0 (PR #1859 landed): the
criteria below are authored now and verified only after the hold lifts.

## §A AC Matrix

- **AC-FLB-001** (REQ-FLB-001/002/003/004) Cause-class token table.
  Given a nominated lease whose claim fails per induced class — (a) a stale-version/holder
  mismatch surviving the retries, (b) a record write held busy past every retry (`flHoldRecord`),
  (c) a claim deadline expired under a still-running write (`factoryLeaseClaimDeadlineOverride`
  shortened + a held write) — when `factory next --run <run> --card <card>` returns, then stderr
  carries exactly that class's reason token: (a) `refused raced`; (b) the record-contention token;
  (c) the claim-deadline token; (b) and (c) carry neither `raced` nor each other's token, and every
  arm exits 4.
  RED-now: on the current tree the (b) and (c) arms report `refused raced` (pinned at
  `factory_lease_bound_test.go` 151, 298) — the class-specific assertions fail here, for the stated
  reason (the misclassification this SPEC fixes), not for an unrelated pre-existing failure.
  Green path: M1 authors the assertions, M2 flips them green.
- **AC-FLB-002** (REQ-FLB-004/005) `raced` exclusivity under load.
  Given the (c) arm of AC-FLB-001, when the refusal renders, then the line does not contain
  `refused raced`, and the detail names the claim deadline and the host-load cause.
- **AC-FLB-003** (REQ-FLB-006) One-line contract.
  Given any refused lease of any class, when the refusal renders, then the stderr line matches
  `factory next: refused <token>: <detail>` with no embedded newline (joined inner errors
  flattened, the card-t1696 contract).
- **AC-FLB-004** (REQ-FLB-007) Exit-code and bare-form stability.
  Given each refusal class — the three claim classes plus serial-slot, quota-hold, and
  foreign-worktree — when the command exits, then every refused nomination exits 4; and given a
  bare-form (unnominated) lease whose claim fails per class (b) or (c), when the verb exits, then
  it keeps its current exit behavior and its error text names the true cause class.
- **AC-FLB-005** (REQ-FLB-010) Contrast-case preservation.
  Given a serial-blocked nominee, when refused, then stderr is byte-identical to the observed
  verbatim line `factory next: refused serial-slot: another serial card is in flight and holds the
  serial slot` and the exit code is 4.
- **AC-FLB-006** (REQ-FLB-008) Compensation invariance.
  Given a nominated lease refused after its promotion for each of the three claim classes, when the
  refusal lands, then the promoted queue item is restored to `queued`, and where another holder owns
  the card the holder's state is untouched (mirroring the current assertions at
  `factory_lease_bound_test.go` 155-157 and 302-307).
- **AC-FLB-007** (REQ-FLB-009) Slot-read classification.
  Given a serial nominee whose slot read reports busy, when the lease path classifies the outcome,
  then the outcome is attributed to the record-contention class and is not reported as `raced`.
- **AC-FLB-008** (REQ-FLB-011) Baseline ordering (attribution witness).
  Given the status read-path change, when the branch's commit graph is read, then the baseline
  artifact `.moai/reports/t1916/status-latency-baseline.md` (N ≥ 20 samples, recipe, host 1-minute
  load at or below the core count, p50/p95) exists in its own commit that precedes the first commit
  modifying the read path.
- **AC-FLB-009** (REQ-FLB-012) Latency outcome.
  Given the AC-FLB-008 baseline and a post-change measurement taken with the same recipe under the
  same load condition, when the 95th percentiles are compared, then exactly one recorded outcome
  holds: post-change p95 ≤ 50% of baseline p95 (defect fixed), or the baseline record shows
  baseline p95 < 200 ms (defect not reproduced; no improvement claimed).
- **AC-FLB-010** (regression) Scoped suite + builds.
  Given the updated lease-path tests, when
  `go test ./internal/cli/ -run 'TestFactoryLease|TestFactoryNext|TestFactoryNominate' -count=1 -v`
  runs, then the matched-test list is recorded, its count is greater than zero (empty-sweep guard —
  a green with a zero-count swept set asserts nothing), every matched test passes, and the
  443/464 queue-lock pins appear in the matched list untouched and passing; `go build ./...` and
  `GOOS=windows GOARCH=amd64 go build ./...` exit 0. The unanchored `-run` pattern is a deliberate
  prefix-family sweep — exact-branch anchoring would exclude the family's longer names
  (`TestFactoryLeaseRecordStallBounded` and siblings) — and the lint engine's `VacuousTestAssertion`
  warning on it is answered by the counted matched list, not by re-anchoring.

## §B REQ → AC traceability

| REQ | AC | REQ | AC |
|---|---|---|---|
| REQ-FLB-001 | AC-FLB-001 | REQ-FLB-007 | AC-FLB-004 |
| REQ-FLB-002 | AC-FLB-001 | REQ-FLB-008 | AC-FLB-006 |
| REQ-FLB-003 | AC-FLB-001 | REQ-FLB-009 | AC-FLB-007 |
| REQ-FLB-004 | AC-FLB-001, AC-FLB-002 | REQ-FLB-010 | AC-FLB-005 |
| REQ-FLB-005 | AC-FLB-002 | REQ-FLB-011 | AC-FLB-008 |
| REQ-FLB-006 | AC-FLB-003 | REQ-FLB-012 | AC-FLB-009 |

## §C Edge cases

- **Joined busy + deadline error**: the claim-deadline class takes precedence (spec § Constraints);
  AC-FLB-001's (c) arm covers it by construction.
- **Bare-form lease**: keeps its current exit and plain-error surface (AC-FLB-004, second clause) —
  the reclassification changes its error text's cause attribution, not its exit.
- **MCP relay**: `handleFactoryNext` surfaces the same classification strings; a relay-side
  re-classification is a defect (plan §D) — covered by AC-FLB-001 run through the MCP handler
  argument in review.
- **Latency floor escape**: if the M4 baseline already shows p95 < 200 ms, AC-FLB-009 records
  "defect not reproduced" — a named pass exit, not a silent skip.

## §D Quality gates

- TRUST 5 per `quality.yaml`: coverage ≥ 85% on touched packages; LSP 0 errors / 0 type errors /
  0 lint errors (zero-tolerance run gates); gofmt clean; Conventional Commits with the card id.
- TDD: verbatim RED output per behavior change (E8) before its GREEN.

## §E Definition of Done

1. All ten ACs PASS with attributed evidence (command + verbatim output + HEAD, this run, this
   tree) in progress.md §E.2; the M0 gate is recorded as open before any implementation commit.
2. AC-FLB-008's ordering is witnessed by the commit graph, not by a commit message assertion.
3. decision-index Q1 and Q2 are dispositioned (operator answer, or Q2 resolved by the M4/M5 data);
   Q4 stands DEFAULT-APPLIED.
4. manager-develop's E1-E8 self-verification is complete; manager-docs owns §E.4 and the close.
