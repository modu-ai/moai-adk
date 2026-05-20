---
id: SPEC-V3R5-HARNESS-AUTONOMY-001
title: "Harness Autonomy — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P0
phase: "v3.5.0"
module: "internal/harness + .moai/harness + .claude/skills/moai-harness-learner + internal/hook"
lifecycle: spec-anchored
tags: "harness, autonomy, self-evolution, acceptance, mega-sprint, w3"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial acceptance criteria — 12 ACs + 6 edge cases + 5 risk-mitigation ACs. Mapped to 36 REQ-HRA-* in spec.md §8 traceability matrix. |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Iteration 2 revision per plan-auditor iter 1 BLOCKING + SHOULD defects. AC-HRA-004 precondition corrected (W1 ships zone-registry DATA SSOT only; W3 implements PreToolUse hook code per B1). AC-HRA-005 wording corrected (orchestrator AskUserQuestion via blocker report, not direct subagent invocation per B5). AC-HRA-008 notification text rephrased (B4 — removed "Override" wording, preserves Canary final-gate). AC-HRA-008b NEW (B4 — cooldown rejection assertion). AC-HRA-013 NEW (S3 — REQ-HRA-007 SCHEMA_DRIFT enforcement). AC-HRA-014 NEW (S6 — REQ-HRA-037 L1 latency NFR benchmark). Empirical baseline corrected: W1 ships data SSOT only, NOT PreToolUse hook. §1.5 brownfield reality corrected: `internal/harness/` is NOT absent (30+ files exist). Total: 14 ACs (12 + 008b + 013 + 014) + 6 EC + 5 R + 1 C-binary verification (C-HRA-008 per S5). |

---

## 1. 검증 개요 (Verification Overview)

본 문서는 SPEC-V3R5-HARNESS-AUTONOMY-001 의 acceptance criteria 를 정의한다.

검증 원칙:

- 모든 AC 는 binary pass/fail (관찰 가능한 증거 기반)
- TRUST 5 quality gate 통과 필수 (`internal/harness/` 신규 코드 coverage ≥ 85%; 기존 코드 PRESERVE per spec.md §1.5 REQ-HRA-038)
- Latency budgets enforced via benchmark tests
- plan-auditor 가 본 acceptance.md 의 모든 scenario 를 독립 검증

**Empirical baseline (binding for all ACs, iteration 2 corrected)** at main HEAD `7bd23bb69`:

- W1 deliverables: zone-registry 111 entries at `.claude/rules/moai/core/zone-registry.md` + `internal/constitution/validator.go` (**DATA SSOT only** per W1 EXCL-001; PreToolUse hook is W3 first-implementer per B1 resolution).
- 8 HARNESS_FROZEN_* sentinel catalog: **defined in Vision §3.4**, NOT W1.
- W2 deliverables complete: moai-foundation-quality preload on 4 expert agents.
- `.moai/harness/` runtime directory: absent (W3 will create).
- `internal/harness/` package directory: **EXISTS** (30+ files per spec.md §1.5 brownfield inventory). W3 EXTENDS `safety/` subdirectory; root layer*.go PRESERVED per REQ-HRA-038.
- `internal/hook/pre_tool.go`: EXISTS (20,548 bytes since 2026-05-18). W3 EXTENDS with 8 HARNESS_FROZEN_* catalog + harness-learner identity gate.

---

## 2. Given/When/Then 시나리오 (12 Core ACs)

### AC-HRA-001 — Lesson auto-capture via SubagentStop hook trigger

**Given**: A workflow event (`/moai run SPEC-XXX`) has just completed wherein manager-develop subagent finished its task and produced file changes to the current branch.

**When**: Claude Code emits the SubagentStop hook event. The `internal/hook/subagent_stop.go` handler dispatches to the harness-learner capture pipeline in background mode.

**Then**:

- Within 500ms p95, the harness-learner produces zero or more observation candidates from the diff input.
- For each candidate, an entry is appended to `.moai/harness/observations.yaml` with all canonical schema fields populated: `id`, `category`, `pattern`, `evidence[]`, `count`, `confidence`, `status`, `created`, `updated`.
- The file write uses `flock(2)` advisory lock; subsequent reads return the appended state.
- No LLM API call occurs during capture (assertion: total elapsed time < 500ms includes no remote round-trip).

**Verification**:

```bash
# Setup: simulate SubagentStop event with synthetic diff
go test -run TestCapture_SubagentStopTrigger -v ./internal/harness/capture/

# Assertion: observations.yaml contains expected entry
yq '.observations | length' .moai/harness/observations.yaml  # > 0
yq '.observations[0].status' .moai/harness/observations.yaml  # "observation"
yq '.observations[0].created' .moai/harness/observations.yaml  # YYYY-MM-DD ISO format
```

---

### AC-HRA-002 — Tier 1 → Tier 4 end-to-end progression

**Given**: An empty `.moai/harness/observations.yaml`. The harness-learner is configured with default tier_thresholds `[1, 3, 5, 10]`.

**When**: A synthetic test fixture injects 10 occurrences of the same pattern across 10 separate SubagentStop events, each event triggering the tier engine to increment `count` on the matching observation entry.

**Then**:

- After event 1: `count: 1`, `status: observation`
- After event 3: `count: 3`, `status: heuristic` (transition emitted to evolution-log.md)
- After event 5: `count: 5`, `status: rule`
- After event 10: `count: 10`, `status: high-confidence`, AND the 5-Layer Safety pipeline is triggered.
- evolution-log.md contains 4 tier-progression entries (one per transition).

**Verification**:

```bash
go test -run TestIntegration_TierProgression1to4 -v ./internal/harness/
```

---

### AC-HRA-003 — 5-Layer Safety unit tests (each layer PASS/FAIL scenarios)

**Given**: A Tier 4 proposal at the input of the 5-Layer Safety pipeline.

**When**: Each layer is unit-tested independently with both PASS and FAIL fixtures via `internal/harness/safety/{l1,l2,l3,l4,l5}_test.go`.

**Then**:

| Layer | PASS path verified | FAIL path verified | Budget verified |
|-------|--------------------|--------------------|------------------|
| L1 Frozen Guard | Proposal to `.claude/skills/my-harness-*/` allowed | Proposal to `.claude/agents/moai/` blocked with HARNESS_LEARNING_FROZEN_BLOCKED | < 10ms p99 |
| L2 Canary | Score delta ≤ 0.10 on 3 SPECs → continue | Score drop > 0.10 → CANARY_VETO | Async dispatch confirmed |
| L3 Contradiction | No conflict with existing rules → continue | Conflict → L3-conflict blocker report emitted to orchestrator (translated to AskUserQuestion per plan.md §3.3a) | < 1s |
| L4 Rate Limiter | Within 3/week + 24h cooldown + ≤50 active → continue | Exceeded any → HARNESS_LEARNING_RATELIMIT_EXCEEDED | < 100ms |
| L5 Human Oversight | User selects Apply → proceeds to write | User selects Reject → status=anti-pattern | User-paced |

**Verification**:

```bash
go test -run 'TestL[1-5]_' -v ./internal/harness/safety/...

# Latency benchmarks
go test -bench 'BenchmarkL[1-4]' ./internal/harness/safety/...
```

---

### AC-HRA-004 (B1 corrected) — Frozen Guard violation block (8 sentinel simulation) using zone-registry as data

**Given (corrected precondition)**: zone-registry exists at `.claude/rules/moai/core/zone-registry.md` with 111 entries (W1 deliverable, main HEAD `7bd23bb69`). The 8 HARNESS_FROZEN_* sentinel catalog is **defined in Vision §3.4** (NOT W1). W3 implements the L1 PreToolUse hook code in `internal/hook/pre_tool.go` extension (W1 EXCL-001: W1 explicitly disclaimed this hook). The hook reads the zone-registry as **input data** at init time.

**When**: A test fixture simulates 8 distinct Tier 4 proposals via the harness-learner subagent, each attempting to write to one of the protected paths:

1. `.claude/agents/moai/expert-security.md` (expects HARNESS_FROZEN_AGENT_VIOLATION)
2. `.claude/skills/moai-foundation-quality/SKILL.md` (expects HARNESS_FROZEN_SKILL_VIOLATION)
3. `.claude/rules/moai/core/agent-common-protocol.md` (expects HARNESS_FROZEN_RULE_VIOLATION)
4. `.claude/commands/moai/run.md` (expects HARNESS_FROZEN_COMMAND_VIOLATION)
5. `.claude/hooks/moai/handle-pre-tool.sh` (expects HARNESS_FROZEN_HOOK_VIOLATION)
6. `.claude/output-styles/moai/moai.md` (expects HARNESS_FROZEN_OUTPUTSTYLE_VIOLATION)
7. `CLAUDE.md` (expects HARNESS_FROZEN_INSTRUCTION_VIOLATION)
8. `.moai/config/sections/harness.yaml` (expects HARNESS_FROZEN_CONFIG_VIOLATION)

**Then**:

- For each of 8 attempts, the W3 L1 hook (consuming zone-registry data) matches the corresponding Vision §3.4 catalog sentinel, AND wraps it in `HARNESS_LEARNING_FROZEN_BLOCKED` with the catalog sentinel as `cause` field.
- The proposal is rejected; L2-L5 are NOT invoked (verified via assertion that no Canary dispatch goroutine started, no rate-limiter mutation, no orchestrator blocker report emitted for L5).
- evolution-log.md records the failed attempt with `final_status: rejected`, `layer_results.L1: blocked`.

**Verification**:

```bash
# Setup: ensure zone-registry exists at .claude/rules/moai/core/zone-registry.md (W1 deliverable, 111 entries)
test -f .claude/rules/moai/core/zone-registry.md
wc -l .claude/rules/moai/core/zone-registry.md  # expect >= 111 entries (header + N entries)

# Run the 8-sentinel matrix test against W3 L1 hook implementation
go test -run TestL1_8SentinelMatrix -v ./internal/harness/safety/

# Verify W3 L1 hook code exists in pre_tool.go extension
grep -q "HARNESS_FROZEN_AGENT_VIOLATION" internal/hook/pre_tool.go
```

---

### AC-HRA-005 (B5 wording corrected) — User rejection → permanent block (anti-pattern)

**Given**: A Tier 4 proposal has reached L5 user oversight. The harness-learner subagent emits a structured L5 blocker report to the orchestrator (per REQ-HRA-018 + C-HRA-008 — subagent NEVER invokes AskUserQuestion directly). The orchestrator (in test mode, via mock orchestrator + mock AskUserQuestion translator) parses the blocker report and runs AskUserQuestion, returning the user selection.

**When**: User selects `Reject permanently` from the 4-option choice (Apply / Apply-with-modification / Defer / Reject-permanently). The orchestrator re-delegates to harness-learner with the user choice injected into the spawn prompt.

**Then**:

- The observation entry's `status` field transitions to `anti-pattern` in `observations.yaml`.
- A copy of the observation (with full evidence) is written to `.moai/harness/anti-patterns.yaml` with FROZEN flag (no further evolution allowed).
- evolution-log.md records `final_status: rejected_permanent`.
- Subsequent SubagentStop events that would re-detect the same pattern do NOT create new observation entries (anti-pattern lookup short-circuits the capture pipeline).

**Verification**:

```bash
go test -run TestL5_RejectPermanent_ToAntiPattern -v ./internal/harness/safety/
```

---

### AC-HRA-006 — Proposal Throttling 4-mode behavior

**Given**: `workflow.yaml` `harness.proposal.mode` is set successively to `immediate`, `batch`, `quiet`, and a mute fixture per AC. The test harness counts the number of AskUserQuestion blocker reports emitted in each mode.

**When**: 5 synthetic Tier 4 proposals are produced in a single test session.

**Then**:

| Mode | Expected AskUserQuestion blocker count | Expected behavior |
|------|----------------------------------------|-------------------|
| `immediate` | 5 (one per proposal, may span multiple rounds of 4+1) | Immediate emit on Tier 4 attainment |
| `batch` (window=manual) | 0 during session; 2 rounds (4+1) after `moai harness apply --batch-flush` | Queued in proposal-queue.yaml |
| `quiet` (current time in [18,9] Asia/Seoul, fixture forces in-window) | 0 (all deferred) | Deferred entries in proposal-queue.yaml |
| `mute` (all 5 proposals' categories in mute.categories[]) | 0 | Logged to evolution-log.md with `status: muted` |

**Verification**:

```bash
go test -run TestThrottling_FourModes -v ./internal/harness/throttle/
```

---

### AC-HRA-007 — Cold-start regression: seed inject → first SPEC quality

**Given**: A synthetic empty project where `.moai/harness/observations.yaml` does not exist. A test seed file is placed at `.claude/skills/moai-meta-harness/seeds/test/go-error-handling.yaml` containing one SEED entry (e.g., `fmt.Errorf with %w wrapping`).

**When**: The harness-learner is invoked for the first time in this project. The cold-start seed loader detects empty observations and loads the seed file.

**Then**:

- `.moai/harness/observations.yaml` is created.
- The seed entry is injected with `status: rule`, `count: 5` synthetic, `confidence: 0.85` per seed file.
- The first subsequent SPEC's manager-develop invocation can reference the seed pattern via observations.yaml lookup (no actual SPEC quality measurement in this AC; we verify the load path works).
- evolution-log.md records a `seed-injected` event with the seed file path and entry IDs.

**Verification**:

```bash
go test -run TestColdStart_SeedInject -v ./internal/harness/seeds/
```

NOTE: Actual SPEC quality measurement is W4 scope (production seed library + measurement methodology). W3 verifies only the schema + load + inject path.

---

### AC-HRA-008 (B4 wording corrected) — Canary Veto Policy (E5) provisional apply + auto-rollback

**Given**: A Tier 4 proposal has passed L1, L3, L4, AND received L5 user approval BEFORE L2 Canary completes (provisional apply path per Vision §6.5 E5).

**When**: The L2 Canary returns FAIL (score drop > 0.10 detected on one of the last 3 SPECs).

**Then**:

- Before L2 completion: the my-harness-* file was written with `evolution_status: provisional` (verified by reading the file and confirming it contains the proposal payload).
- After L2 FAIL: the file content is reverted to its pre-evolution snapshot (stored at `.moai/harness/revert/<evolution-id>/`).
- evolution-log.md records two entries: one for `provisional_apply` and one for `vetoed_by_canary` (with `result: vetoed_by_canary`).
- A 48h cooldown entry is added to the rate-limiter for re-submission gating.
- An orchestrator notification blocker is emitted with the **B4-corrected text** (Canary final-gate preserved — "Override" wording removed): "Canary가 regression을 감지하여 provisional 변경이 자동 롤백되었습니다. 다음 옵션: (a) deeper review (새 proposal 생성 — fresh tier 요구사항 적용, 48h cooldown 후), (b) 거부 (영구 anti-pattern으로 분류)". The path "(a) immediate user-side override" is NOT offered — the cooldown is final.

**Verification**:

```bash
go test -run TestCanaryVeto_ProvisionalRollback -v ./internal/harness/safety/

# B4 wording assertion: verify notification text does NOT contain "Override"
go test -run TestCanaryVeto_NotificationTextNoOverride -v ./internal/harness/safety/
```

---

### AC-HRA-008b (B4 NEW) — Cooldown rejection after Canary veto

**Given**: A Canary-veto event has just occurred at time T, recording a 48h cooldown entry in the L4 rate-limiter (per AC-HRA-008 + REQ-HRA-021).

**When**: At time T + 1h (within cooldown window), a user invokes:

```bash
./moai harness apply <evolution_id>
```

**Then**:

- Exit code: 1 (reject)
- Sentinel emitted (stderr or JSON output `entries[].sentinel`): `HARNESS_LEARNING_RATELIMIT_EXCEEDED`
- The proposal is NOT re-applied; the my-harness-* file remains at its pre-evolution state.
- evolution-log.md records the rejected apply attempt with `final_status: deferred`, `reason: cooldown_active`.

**Verification**:

```bash
go test -run TestCanaryVeto_CooldownRejection -v ./internal/harness/safety/

# Assert exit code via Bash:
./moai harness apply EVOL-TEST-001 2>/tmp/stderr.log
EXIT=$?
[ $EXIT -eq 1 ] && grep -q "HARNESS_LEARNING_RATELIMIT_EXCEEDED" /tmp/stderr.log && echo "B4 cooldown rejection: PASS"
```

**Edge case**: At time T + 48h + 1s (post-cooldown), the same `moai harness apply <evolution_id>` succeeds (EC-HRA-006 cross-reference).

---

### AC-HRA-009 — 6 CLI verb surface

**Given**: The W3 build of `moai` CLI is compiled with `make build`.

**When**: Each of the 6 verbs is invoked:

```bash
./moai harness status --format json
./moai harness apply PROPOSAL-TEST-001
./moai harness rollback EVOL-TEST-001
./moai harness disable
./moai harness mute error-handling
./moai harness mute-list
./moai harness unmute error-handling
./moai harness verify --determinism
./moai harness --help
```

**Then**:

- `status` verb exits 0 and produces JSON conforming to the schema in plan.md §6.3.
- `apply` verb exits 1 on non-existent proposal-id (or 0 if a synthetic fixture proposal is queued).
- `rollback` verb exits 1 on non-existent evolution-id (or 0 with successful revert).
- `disable` verb exits 0 after confirmation; sets `learning.enabled: false` in harness.yaml.
- `mute`/`mute-list`/`unmute` verbs manage workflow.yaml `harness.proposal.mute.categories[]`.
- `verify --determinism` verb exits 0 and prints "Determinism verification deferred to W4 (PROJECT-MEGA-001)".
- `--help` lists all 6 verbs (8 sub-verbs counting mute variants) in the Cobra command tree.

**Verification**:

```bash
go test -run 'TestHarnessCLI_.*' -v ./internal/cli/...
./moai harness --help | grep -E '(status|apply|rollback|disable|mute|verify)'
```

---

### AC-HRA-010 — evolution-log.md append-only with structured entries

**Given**: A series of 5 evolution events: 2 successful applies, 1 user-rejected, 1 Canary-vetoed (with rollback), 1 rate-limit-deferred.

**When**: All 5 events are processed sequentially. After each event, evolution-log.md is inspected.

**Then**:

- After event 1: 1 entry present
- After event 2: 2 entries present (event 1 unchanged)
- After event 3: 3 entries present (events 1+2 unchanged)
- After event 4: 5 entries present (event 4 = provisional + 5 = vetoed; rollback creates NEW reverse-evolution entry, original record preserved)
- After event 5: 6 entries present (event 5 = rate-limit-deferred, recorded with `final_status: deferred`)
- Each entry contains all required fields: `evolution_id`, `timestamp`, `proposal_id`, `layer_results` (L1/L2/L3/L4/L5 per-layer verdict), `final_status`, `affected_files`.
- No earlier entry has been modified or deleted (verified via byte-comparison of the prefix).

**Verification**:

```bash
go test -run TestEvolutionLog_AppendOnly -v ./internal/harness/
```

---

### AC-HRA-011 — Anti-pattern auto-flag on critical failure

**Given**: A test fixture simulates a critical-failure trigger condition: a SPEC quality score dropped by 0.25 (above the 0.20 threshold) between two consecutive iterations of the same SPEC, AND the harness-learner is invoked with the corresponding diff.

**When**: The capture pipeline processes the diff and the tier engine evaluates the failure pattern.

**Then**:

- An entry is written to `.moai/harness/anti-patterns.yaml` IMMEDIATELY (single occurrence triggers the anti-pattern, not waiting for 10x accumulation).
- The entry has `status: anti-pattern` and FROZEN flag set.
- The entry contains the full evidence: `spec_id`, `commit`, `before`, `after`, `context`, `score_delta: -0.25`.
- Subsequent attempts to evolve this pattern (e.g., via a different trigger path) are rejected with `HARNESS_LEARNING_TIER_VIOLATION` sentinel (per plan.md §7).

**Verification**:

```bash
go test -run TestAntiPattern_AutoFlagCriticalFailure -v ./internal/harness/tier/
```

---

### AC-HRA-012 — observations.yaml schema canonical field names

**Given**: A populated `observations.yaml` with at least 3 entries.

**When**:

```bash
# Verify each entry uses canonical field names per Vision §6.2 D3 resolution
yq '.observations[].created' .moai/harness/observations.yaml | grep -v null
yq '.observations[].updated' .moai/harness/observations.yaml | grep -v null

# Verify no snake_case aliases
! yq '.observations[].created_at' .moai/harness/observations.yaml | grep -v null
! yq '.observations[].updated_at' .moai/harness/observations.yaml | grep -v null

# Verify Go decode succeeds against canonical schema
go test -run TestObservations_CanonicalSchemaDecode -v ./internal/harness/tier/
```

**Then**:

- All entries use `created:` (NOT `created_at:`)
- All entries use `updated:` (NOT `updated_at:`)
- Go YAML decode succeeds without `HARNESS_LEARNING_SCHEMA_DRIFT` sentinel.
- The schema aligns with `spec-frontmatter-schema.md` canonical field names (per Vision §6.2 D3 resolution + §6.7 schema separation note).

---

### AC-HRA-013 (S3 NEW) — REQ-HRA-007 SCHEMA_DRIFT enforcement on non-canonical tier_thresholds

**Given**: A test fixture mutates `.moai/config/sections/harness.yaml` setting `learning.tier_thresholds: [1, 3, 5, 11]` (non-canonical — 11 instead of 10).

**When**:

```bash
./moai harness status --format json 2>/tmp/stderr.log
EXIT=$?
```

**Then**:

- Exit code: 1
- stderr contains `HARNESS_LEARNING_SCHEMA_DRIFT` sentinel
- stderr message indicates the offending field path (`learning.tier_thresholds`) and the canonical expected value (`[1, 3, 5, 10]`)
- The learning loop refuses to start (verified by absence of `.moai/harness/observations.yaml` writes following invocation)

**Verification**:

```bash
go test -run TestHarnessConfig_TierThresholdsCanonical -v ./internal/config/
go test -run TestHarnessStatus_SchemaDriftRejection -v ./internal/cli/

# Direct CLI assertion
cp .moai/config/sections/harness.yaml /tmp/harness.yaml.backup
sed -i.bak 's/- 10/- 11/' .moai/config/sections/harness.yaml
./moai harness status --format json 2>/tmp/stderr.log
EXIT=$?
cp /tmp/harness.yaml.backup .moai/config/sections/harness.yaml  # restore
[ $EXIT -eq 1 ] && grep -q "HARNESS_LEARNING_SCHEMA_DRIFT" /tmp/stderr.log && echo "S3 SCHEMA_DRIFT enforcement: PASS"
```

**Edge case (multiple deviations)**: If multiple fields under `learning.*` are non-canonical, the error message lists all deviations, but exit code remains 1.

---

### AC-HRA-014 (S6 NEW) — REQ-HRA-037 L1 latency NFR benchmark

**Given**: W3 L1 Frozen Guard implementation in `internal/hook/pre_tool.go` extension exists. `internal/harness/safety/` has a benchmark file `BenchmarkL1FrozenGuard` that exercises L1 with 10,000 randomized path inputs against a zone-registry-derived 8-pattern matcher.

**When**:

```bash
go test -bench BenchmarkL1FrozenGuard -benchtime 10000x -benchmem ./internal/harness/safety/ > /tmp/bench.out
```

**Then**:

- The benchmark report includes p50 / p95 / p99 latency measurements (statistical via `testing/quick` or `golang.org/x/perf/benchstat` parsing).
- p99 latency MUST be ≤ 10ms (REQ-HRA-037).
- If p99 ≥ 15ms, the test FAILS (BLOCKING per REQ-HRA-037).
- The benchmark MUST NOT allocate heap on the hot path beyond the registry-load amortized cost (verified via `-benchmem` zero per-op allocs after init).

**Verification**:

```bash
go test -bench BenchmarkL1FrozenGuard -benchtime 10000x ./internal/harness/safety/ \
  | tee /tmp/bench.out

# Parse p99 latency from benchstat output and assert ≤ 10ms
go test -run TestL1FrozenGuard_P99Latency -v ./internal/harness/safety/
```

**Edge case (CI variability)**: If CI hardware variability exceeds 5ms on consecutive runs, the gate retries up to 3 times; persistent failure is fatal.

---

## 3. Edge Cases (보강 시나리오)

### EC-HRA-001 — SubagentStop hook fires on every subagent including read-only

**Scenario**: A read-only agent (e.g., manager-strategy in plan mode) completes its task. SubagentStop hook fires.

**Expected behavior**: The harness-learner is invoked, BUT the capture pipeline produces 0 observations (read-only agents have no diff input to analyze). The hook does NOT raise an error or block the workflow.

**Acceptance**: Test fixture with read-only agent → hook invoked → observations.yaml unchanged.

---

### EC-HRA-002 — Concurrent lesson capture race (parallel subagents)

**Scenario**: 3 parallel subagents complete simultaneously (e.g., team mode 3-teammate execution). All 3 SubagentStop hooks fire within 100ms of each other.

**Expected behavior**: All 3 capture events succeed. observations.yaml is written 3 times sequentially due to `flock(2)` advisory lock. No data corruption.

**Acceptance**: Goroutine-based fixture spawns 3 parallel captures → observations.yaml contains all 3 entries (verified by line count + content match).

---

### EC-HRA-003 — Rate limiter cross-week boundary

**Scenario**: 3 evolutions applied during Sunday (the last day of week N). On Monday (week N+1), a 4th evolution proposal arrives.

**Expected behavior**: The L4 Rate Limiter resets its 7-day rolling window. The 4th proposal is accepted (assuming other layers pass), NOT rejected with HARNESS_LEARNING_RATELIMIT_EXCEEDED.

**Acceptance**: Time-elapsed fixture (mocked clock) → week boundary crossed → 4th proposal accepted at L4.

---

### EC-HRA-004 — Quiet hours timezone DST transition

**Scenario**: A user's local timezone observes DST. Quiet hours are configured for `[18, 9]` local time.

**Expected behavior**: For locales without DST (e.g., Asia/Seoul, default), the behavior is deterministic. For DST locales, W3 falls back to fixed UTC offset at session start (documented limitation per plan.md §5.3). The fixture verifies Asia/Seoul behavior; DST locale handling is a documented limitation.

**Acceptance**: Asia/Seoul fixture with `quiet.hours: [18, 9]` → proposals at 19:00 KST deferred, at 10:00 KST emitted.

---

### EC-HRA-005 — Active learning count exceeds 50 (archival)

**Scenario**: observations.yaml accumulates 51 entries (status ∈ {observation, heuristic, rule, high-confidence}).

**Expected behavior**: Per REQ-HRA-017, the oldest entry (by `created:` field) with status=`observation` is archived (status transitions to `archived`, entry remains in file with status flag). Active count drops to 50.

**Acceptance**: 51-entry fixture → archival runs → 1 entry transitions to `archived` → active count = 50.

---

### EC-HRA-006 — Canary veto AFTER 48h cooldown (re-submission path)

**Scenario**: A proposal was Canary-vetoed at time T. At time T + 48h + 1 second, the same pattern is re-detected and reaches Tier 4 again.

**Expected behavior**: Per REQ-HRA-021, the 48h cooldown has elapsed. The re-proposal is accepted at L4 (rate limiter does not reject for this specific entry). The pipeline proceeds normally.

**Acceptance**: Time-elapsed fixture (T + 48h+1s) → re-proposal at L4 succeeds.

---

## 4. Risk Mitigation ACs

### R-HRA-001 — R2: Over-aggressive evolution mitigated by L1+L5 gate

**Scenario**: A malicious or buggy harness-learner attempts to apply 100 Tier 4 proposals in rapid succession to `.claude/agents/moai/expert-security.md`.

**Expected behavior**: All 100 attempts blocked at L1 with `HARNESS_LEARNING_FROZEN_BLOCKED` sentinel. No file write occurs. Rate limiter also independently rejects after 3 attempts (defense in depth).

**Acceptance**: Stress fixture with 100 attempts → 0 file writes → evolution-log.md records 100 `final_status: rejected` entries.

---

### R-HRA-002 — R4: Lesson capture latency budget enforced

**Scenario**: A synthetic 10MB diff is fed to the capture pipeline.

**Expected behavior**: p95 latency < 500ms (REQ-HRA-002). p99 latency < 1s.

**Acceptance**: `BenchmarkLessonCapture` reports p95 < 500ms over 100 runs.

```bash
go test -bench BenchmarkLessonCapture -benchtime 100x ./internal/harness/capture/
```

---

### R-HRA-003 — R7: W3 mechanism complexity — incremental layer activation

**Scenario**: Each of the 5 layers can be unit-tested independently without invoking the others.

**Expected behavior**: Per-layer unit tests pass without requiring the full pipeline orchestration.

**Acceptance**: `internal/harness/safety/{l1,l2,l3,l4,l5}_test.go` each contain at least PASS/FAIL fixtures and run independently in `go test ./internal/harness/safety/l1_test.go ./internal/harness/safety/l1.go` (per-file granularity verified).

---

### R-HRA-004 — R8: Cold-start regression mitigated by seed

**Scenario**: A new project has empty observations.yaml. The harness-learner is invoked for the first SPEC.

**Expected behavior**: Without seed inject, the my-harness-* skill body would have low depth (TIER 1 only). With seed inject (REQ-HRA-023), the skill starts at TIER 3 starting point.

**Acceptance**: Synthetic empty-project fixture with 1 seed file → observations.yaml after first invocation contains the seed entry at `status: rule` (TIER 3 equivalent).

---

### R-HRA-005 — R9: Tier 4 fatigue — batch + mute reduce AskUserQuestion calls

**Scenario**: 5 Tier 4 proposals arrive in a single session.

**Expected behavior**:

| Mode | AskUserQuestion calls |
|------|----------------------|
| `immediate` | 5 (or 2 multi-round = 4+1) |
| `batch` (window=manual, awaiting `apply --batch-flush`) | 0 in session |
| `mute` (all 5 categories muted) | 0 |

**Acceptance**: Counter assertion in `TestThrottling_FourModes` — mode-dependent call counts match expected values.

---

## 5. Definition of Done (DoD)

본 SPEC 은 다음 조건 **모두** 만족 시 완료:

### 5.1 Milestone 별 완료 조건

| Milestone | DoD |
|-----------|-----|
| M1 — Lesson Capture | AC-HRA-001 PASS, R-HRA-002 PASS (latency benchmark), EC-HRA-001 PASS (read-only agent), EC-HRA-002 PASS (concurrent capture) |
| M2 — Tier Engine | AC-HRA-002 PASS (1→4 progression), AC-HRA-011 PASS (anti-pattern), AC-HRA-012 PASS (schema canonical), AC-HRA-013 PASS (S3 SCHEMA_DRIFT enforcement), EC-HRA-005 PASS (archival) |
| M3 — 5-Layer Safety | AC-HRA-003 PASS (per-layer), AC-HRA-004 PASS (8 sentinel; B1 zone-registry data input), AC-HRA-005 PASS (reject permanent via orchestrator blocker), AC-HRA-008 PASS (Canary veto), AC-HRA-008b PASS (B4 cooldown rejection), AC-HRA-014 PASS (S6 L1 latency NFR), R-HRA-001 PASS (over-aggressive), R-HRA-003 PASS (incremental), EC-HRA-003 PASS (week boundary), EC-HRA-006 PASS (48h cooldown), C-HRA-008 PASS (S5 subagent boundary static grep) |
| M4 — Throttling + CLI | AC-HRA-006 PASS (4 modes), AC-HRA-009 PASS (6 verbs), R-HRA-005 PASS (fatigue), EC-HRA-004 PASS (timezone) |
| M5 — Cold-Start Seed | AC-HRA-007 PASS (seed inject — STUB DetectProjectType per S10), R-HRA-004 PASS (cold-start mitigation) |
| M6 — Integration Test | AC-HRA-010 PASS (append-only log), all 14 ACs + 6 ECs + 5 Rs + 1 C-binary run as integration suite |

### 5.2 통합 완료 조건

- All 14 ACs (AC-HRA-001..012 + AC-HRA-008b + AC-HRA-013 + AC-HRA-014) binary PASS
- All 6 EC cases (EC-HRA-001..006) covered by test fixtures
- All 5 R mitigations (R-HRA-001..005) verified
- C-HRA-008 (S5 binary verification) PASS — static grep returns zero AskUserQuestion matches in `internal/harness/` + `internal/hook/`
- REQ-HRA-038 PASS — `internal/harness/layer{1,2,3,5}_test.go` pre/post test counts identical (PRESERVE per §1.5 + REQ-HRA-038)
- spec.md, plan.md, acceptance.md status transitions to `completed` after lifecycle close

### 5.3 TRUST 5 Quality Gate

| Pillar | 검증 방법 | 통과 기준 |
|--------|-----------|-----------|
| Tested | `go test -cover ./internal/harness/...` | ≥ 85% coverage |
| Readable | `golangci-lint run ./internal/harness/... ./internal/cli/harness_*.go` | zero warnings |
| Unified | `gofmt -l ./internal/harness/` | empty output |
| Secured | Code review: no shell injection, no path traversal, lock acquire pattern correct | manual review |
| Trackable | `git log --oneline feat/SPEC-V3R5-HARNESS-AUTONOMY-001..` | Conventional Commits, milestone-aligned |

### 5.4 회귀 방지 (Regression Prevention)

- 8 HARNESS_FROZEN_* sentinel catalog (Vision §3.4-defined) — W3 implementation matches Vision §3.4 enumeration exactly; W1 has no hook to compare against per W1 EXCL-001. Verification command: `grep -c HARNESS_FROZEN .moai/research/harness-autonomy-vision-2026-05-18.md` shall return ≥ 8.
- W1's `internal/constitution/validator.go` UNCHANGED (verified by git diff)
- `.moai/config/sections/harness.yaml` `tier_thresholds: [1, 3, 5, 10]` UNCHANGED (REQ-HRA-007 enforced)
- spec-frontmatter-schema.md canonical field names respected throughout W3 code (no `created_at`/`updated_at`/`labels`)

---

## 6. plan-auditor 검증 포인트 (Independent Review Criteria)

본 acceptance.md 는 plan-auditor 의 다음 차원에서 검증된다:

### D1 — Brief Quality (목표 ≥ 0.85)

- 14 ACs (12 + AC-HRA-008b + AC-HRA-013 + AC-HRA-014) binary verifiable (PASS/FAIL clear)
- 6 edge cases enumerated
- 5 risk mitigation ACs covering Vision §8 risks R2/R4/R7/R8/R9
- 1 binary constraint verification (C-HRA-008 — S5 subagent boundary static grep)
- All 38 REQ-HRA-* (36 + REQ-HRA-037 + REQ-HRA-038) mapped to ≥1 AC (per spec.md §8 traceability matrix)
- Iteration 2 BLOCKING resolutions: B1 W1 deliverable corrected, B2 brownfield inventory + consolidation strategy (b), B3 seed dual-path resolved, B4 Canary cooldown finality, B5 L3+L5 unified blocker pattern

### D2 — Phase Decomposition

- Milestone-aligned DoD (M1..M6)
- Each milestone independently testable
- AC coverage per milestone documented

### D3 — Risk Management

- Vision §8 risks (R2/R3/R4/R7/R8/R9) mapped to mitigations in acceptance.md §4
- W3-specific risks (R11/R12/R13/R14 per plan.md §9.3) covered by ECs or integration test
- Latency budgets enforced via benchmarks

### D4 — Frontmatter Compliance

- 12-field canonical schema 준수
- snake_case alias 없음 (`created`/`updated`/`tags` 사용)

### D5 — Exclusion Discipline

- spec.md §4 EXCL-HRA-001..010 cross-referenced; this acceptance.md does NOT test deferred items.
- W4 boundary clearly maintained (e.g., AC-HRA-007 verifies seed load path, NOT seed library content quality)

### D6 — Lint Baseline

- 본 SPEC 자체의 lint warnings ≤ baseline (no new findings)
- Test fixtures generate observations.yaml entries that pass schema validation

---

## 7. Scope Boundaries

### 7.1 Out of Scope

See `spec.md` §4 for the canonical exclusion list (EXCL-HRA-001 through EXCL-HRA-010). Brief reminder:

- Determinism verification — W4 scope (W3 ships `verify --determinism` placeholder only)
- Seed library content quality measurement — W4 scope (W3 verifies load path only)
- Backward migration tool — explicit non-goal
- LLM-based capture — deferred
- Cross-project sharing — deferred

본 acceptance.md 는 W3 mechanism의 binary AC verification에 집중하며, 다음은 본 acceptance.md scope 외:

- Production seed library quality (W4)
- meta-harness 7-Phase orchestration test (W4)
- Project-specific my-harness generation test (W4)

---

## 8. 후속 검증 (Post-Completion Verification)

본 SPEC 완료 후 다음 검증 자동 실행:

1. **W4 SPEC 작성 시**: SPEC-V3R5-PROJECT-MEGA-001 의 seed library content + meta-harness 7-Phase 가 본 SPEC 의 schema + load hook + tier engine 을 정상 사용 (smoke test)

2. **v3.5.0 release 시**:
   - `moai doctor` 명령에 harness autonomy 상태 확인 항목 통합 (옵션, 후속 SPEC)
   - End-user 시나리오: 신규 프로젝트 `moai init` → `/moai project` → 첫 SPEC `/moai run` → SubagentStop hook 발화 → observations.yaml 생성 확인

3. **harness 진화 로그 정합성**:
   - evolution-log.md 누적 분석 (월간/분기별)
   - 비정상 패턴 (rate-limit 빈번 도달, Canary veto 빈번) 발생 시 follow-up SPEC 검토
