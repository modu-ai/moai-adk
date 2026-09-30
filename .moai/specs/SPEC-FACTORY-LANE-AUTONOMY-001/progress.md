# progress.md — SPEC-FACTORY-LANE-AUTONOMY-001 (card t1338)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-29
tier: L
artifacts: 5 (spec.md, plan.md, acceptance.md, design.md, research.md) + progress.md
req_count: 16
ac_count: 17
needs_clarification_markers: 0
baseline: worktree .moai/worktrees/t1338, branch WT-lane-autonomy-umbrella, base develop 145c3d98c
run_entry_gate: M0 — t1240 (SPEC-FACTORY-SELF-DISPATCH-001) develop merge confirmed mechanically before M1
```

**Plan-audit trajectory** (reports under `.moai/reports/t1338/`, first lines name the serving
auditor model `glm-5.3-flash` per the GLM-lane attribution rule): iter1 `plan-audit-iter1.md`
**PASS 0.97** (Tier L threshold 0.85; 0 blocking / 0 major / 3 minor / 2 advisory; iteration
1/3 — loop closed at first PASS). The 3 minors were polished post-audit (plan.md H1 id typo
AUTOMY→AUTONOMY; research.md autoEvidencePath :36→:32, autoLiveness :47→:51 — each re-verified
by the author against `internal/cli/todo_auto.go`). Carried advisories: D3's "landed-enough"
predicate remains documented discipline until t1241's SPEC text is pinnable at M3 entry (F4);
plan commit `dade0e534` carries no `Authored-By-Agent:` trailer (session attribution reminder
takes precedence; lint INFO OwnershipTransitionUnmeasured, non-strict).

**Phase 2/6 research skip rationale (FO-PLAN-1 note)**: the plan-research fan-out script
(`plan-research-fanout`) exists, but a SINGLE-Explorer pass was chosen for this card. Recorded
decision of the dispatching orchestrator, not a silent deviation: the domain is one coherent
factory-operations subsystem, the 4 boundary SPECs/documents already document the layered surfaces
(t1240/t1241 branch-resident, F1 + t1306 landed in develop), and the carried reconnaissance
(research.md R1-R5) covers every surface the 4 fragments touch. A multi-lens fan-out would have
re-derived the same boundary evidence at ~4x the read cost with no new decision input.

## §E.2 Run-phase Evidence

### M1 — Messaging-fallback detection + self-service switch (2026-09-29, owner: manager-develop)

Surfaces: new package `internal/factorylane` (probe + directed-request observations +
fallback-transition log), CLI verbs `moai factory messaging probe|request|ack`,
`moai factory fallback [declare|restore]`, `moai factory handoff adopt`; config constants
`DefaultFactoryNoResponseMinutes` / `DefaultFactoryFallbackBoundMinutes`
(`internal/config/defaults.go`; heartbeat-age bound reuses `DefaultSessionMsgAgentOfflineMinutes`
— single source). Milestones M2 done (below); M3-M5 pending.

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-FLA-001 | PASS (M1 Go surface) | `go test ./internal/factorylane/ -run TestProbe -count=1` + `go test ./internal/cli/ -run TestFactoryMessagingProbe -count=1` | `ok github.com/modu-ai/moai-adk/internal/factorylane` + `ok ... internal/cli` (probe verbs report `channel: available` / `channel: channel-unavailable` + `--json` verdict; lane switch act = `/moai:todo --auto` doctrine wiring) |
| AC-FLA-002 | PASS (M1 Go surface) | `go test ./internal/factorylane/ -run 'TestSweep|TestAck|TestProbe' -count=1` | `ok ... internal/factorylane` (timer expiry records `no_response_at`; channel unavailable until `requested_at + bound`; ack before expiry stays available) |
| AC-FLA-003 | PASS (M1 Go surface) | `go test ./internal/factorylane/ -run TestDeclareFallback -count=1` + `go test ./internal/cli/ -run TestFactoryFallback -count=1` | `ok` both (exactly one event per switch — second declare refused, count stays 1; restore enables a counted second switch; count-by-lane query prints events) |
| AC-FLA-004 | PASS (M1 Go surface) | `go test ./internal/cli/ -run TestAdopt -count=1` | `ok ... internal/cli` (adopt prints progress.md content + evidence BEFORE work, derives `recorded phase: run`/`sync` from markers, refuses non-picked and nothing-recorded) |
| AC-FLA-005 | PASS (M1 Go surface) | `go test ./internal/cli/ -run TestAdoptAppendsResumption -count=1` | `ok ... internal/cli` (previous owner's progress.md + evidence SHA-256 unchanged; `resumption.jsonl` appended alongside) |

M1 E2-E6 outputs (builds, coverage, boundary grep, lint, commit list) are carried verbatim in the
run-phase completion report returned to the orchestrator and pinned to the M1 commit SHA in git
history; M0 gate evidence: `d43e50bb3` ancestor + SPEC-FACTORY-SELF-DISPATCH-001 present (§F
record). RED-before-GREEN evidence per TDD cycle (5 cycles) is likewise in the completion report
(verbatim failing-test output captured before each GREEN).

### M2 — Classified pickup consumption (2026-09-30, owner: manager-develop)

Surfaces: consumer decision core `internal/factorylane/pickup.go` — the DECLARED MINIMAL
CONSUMPTION INTERFACE per design.md D2 (`Classifier` seam, `NormalizeClassification` tolerance
point, `PlanPickup` rules, `Hold` read from the F1 lease model) — and the CLI verb
`moai factory pickup plan` (`internal/cli/factory_pickup.go`, registered beside
messaging/fallback). t1332 producer GATED: absent as a SPEC at M2 entry (plan.md §F M2 gate), so
every card evaluates through the REQ-FLA-007 fallback default; `StaticClassifier` and the CLI's
flag adapter are the documented wiring points t1332's future reader replaces — no producer schema
defined here (AC-FLA-008 grep 0, below). Exclusivity rides the existing lease records
(`homestate.IsLeaseHoldingState` holders) — no new lock, integration window untouched.

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-FLA-006 | PASS (M2 Go surface) | `go test ./internal/factorylane/ -run TestPickup -count=1` + same with `-race` | `ok github.com/modu-ai/moai-adk/internal/factorylane` (sequential group held by another lane → denied, `WaitOn` names the holder; free group → allowed exclusive `MultiPick=false`; parallel → allowed `MultiPick=true`; two-lane concurrent contention over one sequential group ends with exactly 1 winner, parallel pair → 2 winners; `-race` clean) + CLI edge `ok ... internal/cli` (`--axis sequential --group alpha` prints the classified decision) |
| AC-FLA-007 | PASS (M2 Go surface) | `go test ./internal/factorylane/ -run TestPickupUnclassified -count=1` + `go test ./internal/cli/ -run TestFactoryPickupPlanFallsBackWithoutMetadata -count=1` | `ok` both (unclassified card → `Allowed=true` with `Fallback=true` behavior marker and `MultiPick=false` — the operator-picked single-dispatch behavior, no autonomous multi-pick; reason names REQ-FLA-007) |
| AC-FLA-008 | PASS (M2 Go surface) | `go test ./internal/factorylane/ -run 'TestNormalize|TestPickupUnknownAxis|TestStaticClassifier' -count=1` + producer-schema symbol grep | `ok` (unknown axis token → `Known=false`, decision exits nil-error with `ToleratedUnknown=true` + tolerated log; absent metadata → fallback classification) + `grep -rnE 'MetadataSchema\|SchemaDefinition\|type [A-Za-z]*Metadata struct\|ClassificationSchema\|ProducerSchema\|BacklogItemMetadata' internal/factorylane/ internal/cli/factory_pickup.go` → **0 matches** (no producer-schema code in the consumer) |

M2 E2-E8 outputs (both builds, coverage, boundary grep, lint 0 issues, commit list, verbatim RED
per cycle, the AC-FLA-008 grep, and the `TestSD_AC021_LegacySpellingsRefused` guard result) are
carried in the run-phase completion report returned to the orchestrator; suite evidence persisted
at `.moai/state/verify/t1338-m2/` under the M2 commit SHA in git history. E3 GAP (recorded like
M1's): the `internal/cli` FULL-suite aggregate coverage line was NOT obtained — the full-suite run
under the slot lease hit go test's default 10m timeout and was killed mid-package (`panic: test
timed out after 10m0s`, 601.8s, **0** `--- FAIL` lines — no test failure, an elapsed-bound kill;
the mid-dump `coverage: 24.8%` line is the partial-run value and is not an aggregate). The
substantive measure is the per-package aggregate `go test ./internal/factorylane/ -cover` →
**87.4%** (includes pickup.go) plus the lane-local targeted run `go test ./internal/cli/ -run
'TestFactory|TestSD_AC021' -count=1` → `ok` (116.5s). The full-suite aggregate remains CI's
(observed on `origin/develop` at batch push).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owner: manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owner: manager-docs>_

## §F Phase 4 Mode Selection

**Kickoff gate record**: Implementation Kickoff Approval PASSED in-lane (2026-09-29, operator
answered the lane session's AskUserQuestion directly): run entry approved + progression axis =
autonomous (goal armed after this log). M0 evidence at entry: `d43e50bb3` ancestor of
`origin/develop` AND `SPEC-FACTORY-SELF-DISPATCH-001` present in `origin/develop` tree (both
literal plan.md tests, observed this run). Pre-run baseline absorb: `origin/develop` `8ea2febe2`
merged into the branch as `fce80341f` (incl. t1240 F2 + t1306) so M1-M5 build on landed
self-lease surfaces. Phase 1 plan-audit iter2 scheduled at entry (post-audit polish changed
plan.md/research.md bytes → skip-eligibility condition 3 fails → re-execute per the single
authoritative skip contract).

**Input parameters**: tier L; scope ~15-25 files across internal/cli/factory,
internal/cli/worktree, internal/sessionmsg, internal/kanban; domain count 1 (one coherent
factory-operations subsystem, Go-dominant + doctrine doc touches); concurrency benefit LOW
(coding-heavy); agent-teams prereqs not requested.

**Mode evaluation**:

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file feature work, no trivial single-line shape |
| fanout | no | coding-heavy — Anthropic coding-task parallelism caveat |
| sweep | no | semantic new code, multi-rule, inter-file dependency |
| serial | **YES** | per-milestone manager-develop spawns, Section A-E template |

**Decision: serial**

**Justification**: coding-heavy implementation across a cohesive subsystem; milestones M1-M5
share one package graph and one writer tree (single card worktree), so sequential per-milestone
delegation is the simpler correct envelope. Tier L auto-routing to manager-lead was considered
(the ≥3-milestone AND ≥10-file predicate is met on paper) and declined per the §B.2 boundary
default toward the simpler mode: no cross-domain fan-out is warranted — the fragments consume
one boundary surface each and are implemented against one tree. Boundary case note: file count
and milestone count sit above their thresholds while domain count sits below; the simpler-mode
tie-breaker resolves to serial.
