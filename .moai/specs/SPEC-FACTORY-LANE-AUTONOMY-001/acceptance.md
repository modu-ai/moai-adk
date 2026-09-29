# acceptance.md — SPEC-FACTORY-LANE-AUTONOMY-001 (card t1338)

> Verification layer. AC-FLA-001..017, Given-When-Then, binary-testable. Traceability: one AC per
> REQ-FLA-001..016 plus AC-FLA-017 (cross-fragment concurrency). GEARS obligations live in
> spec.md §B; this file carries no restatement of them.

## §D. AC Matrix

| AC | Verifies | Fragment | Severity |
|----|----------|----------|----------|
| AC-FLA-001 | REQ-FLA-001 | F1 | Must |
| AC-FLA-002 | REQ-FLA-002 | F1 | Must |
| AC-FLA-003 | REQ-FLA-003 | F1 | Must |
| AC-FLA-004 | REQ-FLA-004 | F1 | Must |
| AC-FLA-005 | REQ-FLA-005 | F1 | Must |
| AC-FLA-006 | REQ-FLA-006 | F2 | Must |
| AC-FLA-007 | REQ-FLA-007 | F2 | Must |
| AC-FLA-008 | REQ-FLA-008 | F2 | Must |
| AC-FLA-009 | REQ-FLA-009 | F3 | Must |
| AC-FLA-010 | REQ-FLA-010 | F3 | Must |
| AC-FLA-011 | REQ-FLA-011 | F3 | Must |
| AC-FLA-012 | REQ-FLA-012 | F4 | Must |
| AC-FLA-013 | REQ-FLA-013 | F4 | Must |
| AC-FLA-014 | REQ-FLA-014 | F4 | Must |
| AC-FLA-015 | REQ-FLA-015 | F4 | Must |
| AC-FLA-016 | REQ-FLA-016 | X-cut | Must |
| AC-FLA-017 | F1×F2×F3 joint | Concurrency | Should |

## §D.1 Scenarios

- **AC-FLA-001** (maps REQ-FLA-001) **Given** a lane with the probe available and the lead peer unregistered **When**
  the lane runs the availability probe **Then** the probe exits reporting `channel-unavailable`,
  and the lane's next pickup is served by `/moai:todo --auto` self-service mode (probe output +
  mode marker both observable).
- **AC-FLA-002** (maps REQ-FLA-002) **Given** the channel nominally up and a directed lead request sent **When** the
  bounded no-response timer expires with no ack **Then** a no-response observation is recorded
  with its timestamp, and the lane treats the channel as unavailable for the remainder of the
  bound period (log row observable).
- **AC-FLA-003** (maps REQ-FLA-003) **Given** a lane switches to self-service mode **When** the transition completes
  **Then** exactly one transition event exists in the queryable log, carrying lane id, trigger
  kind (`channel-unavailable` | `no-response`), timestamp, and the in-progress card id; a second
  switch appends a second event (count-by-lane query is the check).
- **AC-FLA-004** (maps REQ-FLA-004) **Given** a card in `picked` state whose owner lane is determined stalled by
  t1241's stall interface, with a non-empty `progress.md` **When** a resuming lane adopts the
  card **Then** the resuming lane's transcript records the `progress.md` + evidence read BEFORE
  its first work action, and its first recorded phase equals the card's recorded phase (not
  `plan`).
- **AC-FLA-005** (maps REQ-FLA-005) **Given** the resumption of AC-FLA-004 **When** the resuming lane writes its
  records **Then** the previous owner's `progress.md` content and evidence files are
  byte-unchanged (`git diff` / checksum before vs after), and the resuming lane's entries are
  appended.
- **AC-FLA-006** (maps REQ-FLA-006) **Given** two lanes and a queued sequential card carrying t1332 metadata
  (sequential) **When** both lanes reach pickup simultaneously **Then** exactly one lane picks it
  and the other proceeds to a different pick or waits; **and given** a parallel-classified card
  **When** two lanes pick concurrently **Then** both holds succeed (two observable outcomes).
- **AC-FLA-007** (maps REQ-FLA-007) **Given** a queued card with NO classification metadata **When** a lane in
  self-service mode evaluates it **Then** the lane does not autonomously multi-pick and behaves as
  the operator-picked single-dispatch path would (behavior marker observable).
- **AC-FLA-008** (maps REQ-FLA-008) **Given** the consumer reading a card whose metadata record lacks the axis field
  or carries an unknown value **When** pickup evaluates it **Then** the consumer exits 0 with the
  fallback classification and logs the tolerated-unknown condition — no error, no producer-schema
  code in the diff (`grep` for schema-definition symbols over the consumer package returns 0).
- **AC-FLA-009** (maps REQ-FLA-009) **Given** a lane whose card branch fails ANY of the triple (sync-audit not PASS /
  unresolved conflict / `HEAD^{tree} != HEAD^2^{tree}`) **When** the lane attempts merge
  preparation **Then** the merge is refused with the failing condition named; **and given** all
  three pass **When** the lane proceeds **Then** the recorded check output for all three exists
  before the `integration acquire` timestamp.
- **AC-FLA-010** (maps REQ-FLA-010) **Given** the integration window held by lane A (live `acquire` record) **When**
  lane B runs acquire **Then** B is refused/waiting with the holder named, and B's merge lands
  only after A's `release` (window ledger order observable).
- **AC-FLA-011** (maps REQ-FLA-011) **Given** any lane-direct merge of a card branch into local develop **When** the
  merge commit is inspected **Then** an `integration acquire` hold record for that lane exists in
  the window ledger covering the merge timestamp — no out-of-window merge exists in the audit
  sample.
- **AC-FLA-012** (maps REQ-FLA-012) **Given** a card worktree whose merge commit has NOT reached origin/develop
  **When** disposal is requested **Then** the machine check runs `git fetch origin develop` +
  `git rev-list --count --left-right origin/develop...<merge>` and the disposal is REFUSED with
  the check output shown.
- **AC-FLA-013** (maps REQ-FLA-013) **Given** an L1-tier worktree path, and separately a worktree with a live
  anchored session **When** disposal is requested in each case **Then** the L1 refusal and the
  ANCHORED_SESSIONS_PRESENT refusal each fire exactly as before this SPEC (existing test suite
  for `done.go` passes unmodified).
- **AC-FLA-014** (maps REQ-FLA-014) **Given** the machine check confirms the merge commit on origin **When** disposal
  runs with `--auto` **Then** disposal completes with no operator interaction (exit 0 + removal
  observable), and the CI status was NOT consulted by the machine check (no CI-read call in the
  path).
- **AC-FLA-015** (maps REQ-FLA-015) **Given** a worktree whose card branch is unpushed or whose merge is local-only
  **When** any disposal path (manual or `--auto`) is attempted **Then** disposal is refused; the
  worktree still exists after the attempt (existence check).
- **AC-FLA-016** (maps REQ-FLA-016) **Given** the full implemented surface **When** audited for delegation routes
  **Then** the only card-delegation channel mutations are queue mutations; zero new
  messaging-based delegation calls exist in the diff, and card admission code paths are unchanged
  (`git diff` scope check).
- **AC-FLA-017** (maps REQ-FLA-003, REQ-FLA-006, REQ-FLA-009, REQ-FLA-010) **Given** three lanes — one in fallback self-service, two picking classified
  cards (one sequential group, one parallel) **When** the system runs a mixed workload **Then**
  the sequential group never shows two simultaneous holds, the parallel pair does, the fallback
  lane's transition is logged, and every merge in the run sits inside an acquired window (joint
  scenario log observable).

## §D.2 Edge Cases

- Lead heartbeat stale but channel technically up → the probe's heartbeat-age bound decides
  `unavailable` (AC-FLA-001 path); the bound value is configuration, not a magic constant.
- Disposal raced by an operator opening a session in the worktree → anchored-session guard wins
  over `--auto` (AC-FLA-013 outranks AC-FLA-014).
- t1332 lands with a schema richer than the consumer's minimal interface → unknown fields
  tolerated (AC-FLA-008); no consumer change REQUIRED.
- Merge commit reachable but origin fetch fails (network) → machine check cannot confirm →
  refuse (fail-closed, AC-FLA-012 path).
- Resumption racing the previous owner's liveness → t1241's stall determination is the sole
  trigger; a heartbeat-alive owner is never resumed (AC-FLA-004 precondition).

## §D.3 Indirect Verification

- REQ-FLA-016 (queue-as-channel) is verified negatively — by absence of new messaging-delegation
  routes in the diff (AC-FLA-016), not by a positive test.
- REQ-FLA-013 guard preservation is verified by the existing `internal/cli/worktree` test suite
  passing unmodified (AC-FLA-013).

## §D.4 Quality Gates (TRUST 5)

- **Tested**: affected-package coverage ≥ 85% (`quality.yaml` `test_coverage_target: 85`); new
  probe/consumer/disposal logic fully unit-tested; characterization tests for the `worktree done`
  guard behavior.
- **Readable/Unified**: `golangci-lint run` + `gofmt` clean on affected packages; lane lint uses
  the CI golangci version (house lesson, t1235/t1271).
- **Secured**: the machine check runs read-only git commands; no credentials in the transition
  log; the probe makes no network calls beyond `git fetch`.
- **Trackable**: Conventional Commits per milestone; card id t1338 in every commit message.

## §D.5 Definition of Done

1. AC-FLA-001..017 all PASS with observed command output (no claimed-without-run rows).
2. M0 gate satisfied before M1 (t1240 develop merge, mechanically confirmed).
3. Zero boundary-card overlap (spec.md §F exclusions hold against the final diff).
4. Operator-gate reduction table (plan.md) matches the implemented surface — no undocumented
   shrink.
5. Sync phase closes with `worktree done` guards and the `--auto` foreman contract unchanged.
