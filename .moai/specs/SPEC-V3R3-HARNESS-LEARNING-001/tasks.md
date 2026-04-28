# SPEC-V3R3-HARNESS-LEARNING-001 — Tasks

Granular work items per phase. Each task references the REQ-IDs it implements. Dependencies on SPEC-V3R3-HARNESS-001 and SPEC-V3R3-PROJECT-HARNESS-001 are marked explicitly with `[DEP:HARNESS-001]` or `[DEP:PROJECT-HARNESS-001]`.

Task ID format: `T-P{phase}-{NN}` where `{phase}` is the phase number (1-5) and `NN` is the in-phase sequence.

---

## Phase 1 — Observer + Log Schema

### T-P1-01: Define JSONL log schema
- **Type**: Design
- **REQ-IDs**: REQ-HL-001
- **Dependencies**: None
- **Output**: Document the schema in `internal/harness/types.go` as Go structs with JSON tags (`Event`, `EventType` enum). Include schema version field for forward compat.
- **Done when**: Struct compiles; example marshal/unmarshal round-trip test passes.

### T-P1-02: Implement observer.go core
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-001
- **Dependencies**: T-P1-01
- **Output**: `internal/harness/observer.go` with `RecordEvent(eventType, subject, contextHash) error`. Uses `os.OpenFile(O_APPEND|O_CREATE|O_WRONLY)` for atomic writes.
- **Done when**: 100 sequential `RecordEvent` calls each complete in <100ms; output JSONL is `jq`-parseable.

### T-P1-03: Implement PostToolUse hook wrapper
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-001
- **Dependencies**: T-P1-02
- **Output**: `.claude/hooks/moai/handle-harness-observe.sh` that pipes stdin to `moai hook harness-observe`; `internal/cli/hook.go` adds `harness-observe` subcommand routing to observer.
- **Done when**: Hook fires on PostToolUse for `Bash`, `Edit`, `Write`, `Agent`, `AskUserQuestion`; verified by reading log after a sample session.

### T-P1-04: Implement log retention + pruning
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-011
- **Dependencies**: T-P1-02
- **Output**: `internal/harness/retention.go` with `PruneStaleEntries(retentionDays int) error`. Pruning runs on every `RecordEvent` (lazy) but skips if last prune was within 1 hour.
- **Done when**: Unit test with mock clock verifies entries older than retention are removed and archived to `.moai/harness/learning-history/archive/<YYYY-MM>.jsonl.gz`.

### T-P1-05: Phase 1 unit + integration tests
- **Type**: Test
- **REQ-IDs**: REQ-HL-001, REQ-HL-011
- **Dependencies**: T-P1-02, T-P1-03, T-P1-04
- **Output**: `internal/harness/observer_test.go` (unit), `internal/harness/integration_test.go` (sample session replay).
- **Done when**: `go test -race ./internal/harness/...` passes; coverage ≥ 85%.

---

## Phase 2 — Tier Classifier

### T-P2-01: Define Pattern + Tier types
- **Type**: Design
- **REQ-IDs**: REQ-HL-002
- **Dependencies**: T-P1-01
- **Output**: Add `Pattern`, `Tier`, `Promotion` types to `internal/harness/types.go`. Tier is enum: `Observation, Heuristic, Rule, AutoUpdate`.
- **Done when**: Types compile; tier transitions enumerable.

### T-P2-02: Implement pattern aggregator
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-002
- **Dependencies**: T-P2-01
- **Output**: `internal/harness/learner.go` `AggregatePatterns(logPath string) (map[string]*Pattern, error)` reads JSONL, groups by `(event_type, subject, context_hash)`, returns counts.
- **Done when**: Unit test with synthetic 1,000-event log produces expected pattern map.

### T-P2-03: Implement tier classifier
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-002
- **Dependencies**: T-P2-02
- **Output**: `internal/harness/learner.go` `ClassifyTier(p *Pattern, thresholds []int) Tier`. Confidence below 0.70 forces `Observation` regardless of count.
- **Done when**: Boundary tests at counts {0, 1, 2, 3, 4, 5, 9, 10, 11} produce correct tier; low-confidence override tested.

### T-P2-04: Implement Tier 2 description enrichment writer
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-003
- **Dependencies**: T-P2-03, [DEP:HARNESS-001] (target file structure must exist)
- **Output**: `internal/harness/applier.go` (initial scaffold) with `EnrichDescription(skillPath, heuristicNote string) error`. Modifies frontmatter `description` only; preserves all other fields and body.
- **Done when**: Unit test verifies description is updated, body byte-identical, no other frontmatter field changed.

### T-P2-05: Implement Tier 3 trigger injection writer (deferred-write mode)
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-004
- **Dependencies**: T-P2-04, [DEP:HARNESS-001]
- **Output**: `internal/harness/applier.go` `InjectTrigger(skillPath, keyword string) error`. Deduplicates triggers list; in Phase 2 the writer is gated behind a feature flag — actual file write deferred to Phase 4.
- **Done when**: Unit test verifies deduplication and idempotency; feature flag prevents actual writes in Phase 2.

### T-P2-06: Implement promotion event writer
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-002
- **Dependencies**: T-P2-03
- **Output**: `internal/harness/learner.go` `WritePromotion(p Promotion) error` appends to `.moai/harness/learning-history/tier-promotions.jsonl`.
- **Done when**: Promotion log line schema validated against §4.2 of plan.md.

### T-P2-07: Phase 2 unit tests
- **Type**: Test
- **REQ-IDs**: REQ-HL-002, REQ-HL-003, REQ-HL-004
- **Dependencies**: T-P2-02 through T-P2-06
- **Output**: `internal/harness/learner_test.go` covering aggregator, classifier, promotion writer; `internal/harness/applier_test.go` covering enrichment + trigger injection.
- **Done when**: All tests pass; coverage ≥ 85% for new files.

---

## Phase 3 — 5-Layer Safety Architecture

### T-P3-01: Implement L1 Frozen Guard
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-006
- **Dependencies**: None (independent of Phase 1/2)
- **Output**: `internal/harness/safety/frozen_guard.go` with hardcoded prefix list `[".claude/agents/moai/", ".claude/skills/moai-", ".claude/rules/moai/", ".moai/project/brand/"]`. Function: `IsFrozen(path string) bool`. Path normalization via `filepath.Clean` + symlink resolution via `filepath.EvalSymlinks`.
- **Done when**: Unit test enumerates all known violation patterns + golden USER paths; symlink + traversal cases covered; bypass attempt via config returns true (hardcoded).

### T-P3-02: Implement Frozen Guard violation logger
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-006
- **Dependencies**: T-P3-01
- **Output**: `internal/harness/safety/frozen_guard.go` `LogViolation(path, caller string) error` appends to `.moai/harness/learning-history/frozen-guard-violations.jsonl`; emits stderr warning.
- **Done when**: Triggered violation produces JSONL entry + visible stderr message.

### T-P3-03: Implement L2 Canary Check
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-007
- **Dependencies**: T-P1-02
- **Output**: `internal/harness/safety/canary.go` `EvaluateCanary(proposal Proposal, sessions []Session) (CanaryResult, error)`. Effectiveness score = weighted sum of (subcommand success rate, agent invocation success, completion rate). Reject if delta > threshold (default 0.10).
- **Done when**: Unit test injects baseline + degrading proposal; canary correctly rejects.

### T-P3-04: Implement L3 Contradiction Detector
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-008
- **Dependencies**: T-P2-05
- **Output**: `internal/harness/safety/contradiction.go` detects (a) overlapping trigger keywords across skills, (b) contradictory chaining-rules entries. Returns `ContradictionReport` for surfacing via L5.
- **Done when**: Unit test with crafted overlapping triggers returns report; non-conflicting case returns empty report.

### T-P3-05: Implement L4 Rate Limiter
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-008 (rate dimension)
- **Dependencies**: None
- **Output**: `internal/harness/safety/rate_limit.go` sliding-window counter persisted to `.moai/harness/learning-history/rate-limit-state.json`. Functions: `CheckLimit() (allowed bool, retryAfter time.Duration)`, `RecordUpdate()`.
- **Done when**: Unit test verifies max 3 updates per 7-day window + 24h cooldown between updates; persistence survives process restart.

### T-P3-06: Implement L5 Human Oversight bridge
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-005, REQ-HL-008 (surfacing)
- **Dependencies**: T-P3-04
- **Output**: `internal/harness/safety/oversight.go` returns proposal payload (problem statement + 2-4 options) for the orchestrator skill to surface via AskUserQuestion. Subagent itself does NOT call AskUserQuestion (per agent-common-protocol §User Interaction Boundary).
- **Done when**: Unit test verifies proposal payload structure matches AskUserQuestion option schema (max 4 options, recommended first).

### T-P3-07: Compose safety pipeline
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-005, REQ-HL-006, REQ-HL-007, REQ-HL-008
- **Dependencies**: T-P3-01 through T-P3-06
- **Output**: `internal/harness/safety/pipeline.go` `Evaluate(proposal Proposal) (Decision, error)` runs L1 → L2 → L3 → L4 → L5 in order, short-circuits on rejection.
- **Done when**: Integration test with synthetic proposal exercises all 5 layers in correct order; ordering enforced by test (mock layer counters).

### T-P3-08: Phase 3 unit + integration tests
- **Type**: Test
- **REQ-IDs**: REQ-HL-005, REQ-HL-006, REQ-HL-007, REQ-HL-008
- **Dependencies**: T-P3-01 through T-P3-07
- **Output**: Per-layer unit tests + composition integration test.
- **Done when**: `go test -race ./internal/harness/safety/...` passes; coverage ≥ 90% (safety code is critical).

---

## Phase 4 — Auto-Update Applier + CLI + Coordinator Skill

### T-P4-01: Implement applier with snapshot creation
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-005, REQ-HL-009 (rollback dependency)
- **Dependencies**: T-P3-07, T-P2-04, T-P2-05
- **Output**: `internal/harness/applier.go` (full version) `Apply(decision Decision) error`. Creates snapshot at `.moai/harness/learning-history/snapshots/<ISO-DATE>/` with `manifest.json` BEFORE writing changes.
- **Done when**: Unit test verifies snapshot precedes write; rollback restores byte-identical state.

### T-P4-02: Implement `/moai harness status`
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-009
- **Dependencies**: T-P2-06, T-P3-05
- **Output**: `internal/cli/harness.go` `cmdStatus` reports tier distribution, last update, rate-limit window, pending proposals, observer state.
- **Done when**: CLI integration test verifies output format on fresh + populated projects.

### T-P4-03: Implement `/moai harness apply`
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-005, REQ-HL-009
- **Dependencies**: T-P4-01, T-P3-06
- **Output**: `internal/cli/harness.go` `cmdApply` loads next pending proposal, returns it for orchestrator-mediated AskUserQuestion, applies on approval.
- **Done when**: CLI integration test simulates approval flow end-to-end.

### T-P4-04: Implement `/moai harness rollback <date>`
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-009
- **Dependencies**: T-P4-01
- **Output**: `internal/cli/harness.go` `cmdRollback` reads snapshot manifest, restores files, logs rollback event.
- **Done when**: Integration test verifies byte-identical restoration; nonexistent date returns exit 1 with helpful message.

### T-P4-05: Implement `/moai harness disable`
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-009, REQ-HL-010
- **Dependencies**: None
- **Output**: `internal/cli/harness.go` `cmdDisable` sets `learning.enabled: false` in `.moai/config/sections/harness.yaml` (preserves all other keys via YAML round-trip).
- **Done when**: Integration test verifies config update + observer no-op after disable.

### T-P4-06: Create coordinator skill `moai-harness-learner`
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-005
- **Dependencies**: T-P4-01, T-P4-03
- **Output**: `.claude/skills/moai-harness-learner/SKILL.md` with Quick Reference + Implementation Guide sections (per moai-foundation-core §5 Progressive Disclosure). Skill triggers learner runs and surfaces Tier 4 proposals.
- **Done when**: Skill file passes `commands_audit_test.go`-style frontmatter validation (CSV `allowed-tools`, etc.).

### T-P4-07: Add config defaults to template
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-010
- **Dependencies**: None
- **Output**:
  - `.moai/config/sections/harness.yaml` `learning:` section as defined in plan.md §4.3.
  - `internal/template/templates/.moai/config/sections/harness.yaml` byte-identical mirror.
- **Done when**: Both files exist; mirror enforcer test passes.

### T-P4-08: Update `internal/template/templates/` and rebuild embedded
- **Type**: Implementation
- **REQ-IDs**: REQ-HL-010
- **Dependencies**: T-P4-07
- **Output**: Run `make build` to regenerate `internal/template/embedded.go`; commit alongside templates per CLAUDE.local.md §2 Template-First Rule.
- **Done when**: `moai init test-project` deploys the new harness.yaml without manual intervention.

### T-P4-09: Phase 4 CLI integration tests
- **Type**: Test
- **REQ-IDs**: REQ-HL-009
- **Dependencies**: T-P4-02 through T-P4-05
- **Output**: `internal/cli/harness_test.go` covering all 4 verbs across macOS/Linux/Windows.
- **Done when**: CI matrix passes on all three OS targets.

---

## Phase 5 — Integration Tests + Documentation

### T-P5-01: IT-01 — 100-event session replay
- **Type**: Test
- **REQ-IDs**: REQ-HL-001, REQ-HL-002
- **Dependencies**: All Phase 1 + Phase 2 tasks complete
- **Output**: `test/integration/harness/it01_replay_test.go` replays a recorded 100-event session and verifies tier distribution.
- **Done when**: Test passes deterministically.

### T-P5-02: IT-02 — Tier 3 promotion + snapshot
- **Type**: Test
- **REQ-IDs**: REQ-HL-004
- **Dependencies**: T-P4-01
- **Output**: `test/integration/harness/it02_tier3_test.go` verifies frontmatter modification + snapshot creation.
- **Done when**: Test passes.

### T-P5-03: IT-03 — Tier 4 with AskUserQuestion gate
- **Type**: Test
- **REQ-IDs**: REQ-HL-005
- **Dependencies**: T-P4-03
- **Output**: `test/integration/harness/it03_tier4_test.go` simulates approval flow via mock orchestrator.
- **Done when**: Test passes.

### T-P5-04: IT-04 — Frozen Guard rejection
- **Type**: Test
- **REQ-IDs**: REQ-HL-006
- **Dependencies**: T-P3-02
- **Output**: `test/integration/harness/it04_frozen_test.go` attempts write to `.claude/skills/moai-foundation-core/SKILL.md`, verifies block + violation log.
- **Done when**: Test passes; violation log contains expected entry.

### T-P5-05: IT-05 — Rate limiter blocks 4th update
- **Type**: Test
- **REQ-IDs**: REQ-HL-008
- **Dependencies**: T-P3-05
- **Output**: `test/integration/harness/it05_rate_test.go` performs 4 sequential updates within mock 7-day window; verifies 4th rejected.
- **Done when**: Test passes.

### T-P5-06: IT-06 — Rollback byte-identical restoration
- **Type**: Test
- **REQ-IDs**: REQ-HL-009
- **Dependencies**: T-P4-04
- **Output**: `test/integration/harness/it06_rollback_test.go` snapshot → modify → rollback → verify SHA256 match with snapshot.
- **Done when**: Test passes.

### T-P5-07: IT-07 — `learning.enabled: false` disables observer + applier
- **Type**: Test
- **REQ-IDs**: REQ-HL-010
- **Dependencies**: T-P4-05
- **Output**: `test/integration/harness/it07_disable_test.go` sets enabled: false, verifies no observer writes, no applier writes.
- **Done when**: Test passes.

### T-P5-08: User-facing documentation
- **Type**: Documentation
- **REQ-IDs**: REQ-HL-009, REQ-HL-010
- **Dependencies**: All Phase 4 tasks
- **Output**: `.moai/harness/README.md` documenting CLI verbs, config keys, tier thresholds, snapshot locations, rollback procedure.
- **Done when**: Document reviewed by manager-docs; section added to docs-site (deferred PR per §17 4-locale rule).

### T-P5-09: Cross-platform CI verification
- **Type**: Test
- **REQ-IDs**: All
- **Dependencies**: T-P5-01 through T-P5-07
- **Output**: GitHub Actions workflow runs full integration suite on ubuntu-latest, macos-latest, windows-latest.
- **Done when**: Green CI on all three OS targets.

### T-P5-10: 1-week real-developer trial
- **Type**: Validation
- **REQ-IDs**: All
- **Dependencies**: T-P5-09
- **Output**: Solo developer runs harness for 7 days; reports `/moai harness status` output and Frozen Guard violation count (target: 0).
- **Done when**: Trial complete; Definition of Done in `acceptance.md` checked off.

---

## Dependency Summary

| Task | Depends On | External SPEC |
|------|------------|---------------|
| T-P2-04, T-P2-05 | Phase 2 | [DEP:HARNESS-001] (target skill files must exist) |
| Phase 1 onset | — | [DEP:HARNESS-001] + [DEP:PROJECT-HARNESS-001] (both must merge before Phase 1 begins) |
| T-P5-04 (IT-04) | Phase 3 | [DEP:HARNESS-001] (validates boundary against MOAI area created by HARNESS-001) |

## Task Count Summary

| Phase | Tasks |
|-------|-------|
| Phase 1 | 5 |
| Phase 2 | 7 |
| Phase 3 | 8 |
| Phase 4 | 9 |
| Phase 5 | 10 |
| **Total** | **39** |

## REQ-ID Coverage Map (per task)

| REQ-ID | Tasks |
|--------|-------|
| REQ-HL-001 | T-P1-01, T-P1-02, T-P1-03, T-P1-05, T-P5-01 |
| REQ-HL-002 | T-P2-01, T-P2-02, T-P2-03, T-P2-06, T-P2-07, T-P5-01 |
| REQ-HL-003 | T-P2-04, T-P2-07 |
| REQ-HL-004 | T-P2-05, T-P2-07, T-P5-02 |
| REQ-HL-005 | T-P3-06, T-P3-07, T-P3-08, T-P4-01, T-P4-03, T-P5-03 |
| REQ-HL-006 | T-P3-01, T-P3-02, T-P3-08, T-P5-04 |
| REQ-HL-007 | T-P3-03, T-P3-08 |
| REQ-HL-008 | T-P3-04, T-P3-05, T-P3-06, T-P3-07, T-P3-08, T-P5-05 |
| REQ-HL-009 | T-P4-02, T-P4-03, T-P4-04, T-P4-05, T-P4-09, T-P5-06, T-P5-08 |
| REQ-HL-010 | T-P4-05, T-P4-07, T-P4-08, T-P5-07, T-P5-08 |
| REQ-HL-011 | T-P1-04, T-P1-05 |

Coverage: 11/11 REQ-IDs mapped to ≥1 task each (100%).
