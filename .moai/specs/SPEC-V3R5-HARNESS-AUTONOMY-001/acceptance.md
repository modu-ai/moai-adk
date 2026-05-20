# SPEC-V3R5-HARNESS-AUTONOMY-001 — Acceptance Criteria

> 14 binary AC-HRA-* + 6 Edge Cases + 5 Risk Mitigations + 1 Constraint AC (C-HRA-008 subagent-boundary grep). All ACs are binary PASS/FAIL with reproducible verification commands.

## 1. Acceptance Criteria (AC-HRA-001..014)

### AC-HRA-001 — Observation auto-capture on SubagentStop

**REQ**: REQ-HRA-006, REQ-HRA-007 · **Milestone**: M1

**Setup**: Fresh project with empty `.moai/harness/observations.yaml`. Trigger a subagent invocation that produces a diff containing `fmt.Errorf("...: %w", err)`.

**Verification**:
```bash
go test ./internal/harness/capture/... -run TestExtractorOnSubagentStop
test -s .moai/harness/observations.yaml
yq '.observations | length' .moai/harness/observations.yaml  # ≥ 1
yq '.observations[0].pattern' .moai/harness/observations.yaml  # contains "%w"
yq '.observations[0].count' .moai/harness/observations.yaml  # = 1
yq '.observations[0].status' .moai/harness/observations.yaml  # = "observation"
```

**PASS**: observation entry created with `status: observation`, `count: 1`, canonical `created:` / `updated:` field names per §1.7.

---

### AC-HRA-002 — Field naming policy enforcement (REQ-HRA-001)

**REQ**: REQ-HRA-001 · **Milestone**: M1

**Verification**:
```bash
# canonical field names required, _at suffix forbidden
yq '.observations[] | has("created")' .moai/harness/observations.yaml | grep -v true; [ $? -ne 0 ]
grep -c "created_at\|updated_at" .moai/harness/observations.yaml  # = 0
go test ./internal/harness/... -run TestHarnessSchemaCanonicalFieldNames
```

**PASS**: `harness_schema_test.go` passes; zero `_at` field name occurrences.

---

### AC-HRA-003 — Tier progression 1x → 3x → 5x → 10x

**REQ**: REQ-HRA-008 · **Milestone**: M2

**Verification**:
```bash
go test ./internal/harness/tier/... -run TestTierProgression
# Simulates 10 observations of same pattern; checks status transitions:
#   count=1  → observation
#   count=3  → heuristic
#   count=5  → rule
#   count=10 → high-confidence (Proposal generated)
```

**PASS**: 4 transitions verified in single test; Proposal YAML emitted at count=10 matches vision §6.4 schema.

---

### AC-HRA-004 — Anti-pattern instant FROZEN

**REQ**: REQ-HRA-014 · **Milestone**: M2

**Verification**:
```bash
go test ./internal/harness/tier/... -run TestAntiPatternInstantFreeze
# Simulates 1 observation with severity=critical (score drop 0.25)
# Verifies status='anti-pattern' immediately + flag FROZEN=true
```

**PASS**: anti-pattern entry created on first observation, FROZEN flag set, subsequent observations cannot reclassify (only human intervention).

---

### AC-HRA-005 — Proposal format vision §6.4 compliance

**REQ**: REQ-HRA-009 · **Milestone**: M2

**Verification**:
```bash
go test ./internal/harness/tier/... -run TestProposalGeneratorVisionFormat
# Verifies emitted Proposal has: question, header, options[4]
# First option label ends with "(권장)" or "(Recommended)"
# All option descriptions populated (no empty strings)
```

**PASS**: Proposal matches §6.4 verbatim structure; first option marked recommended.

---

### AC-HRA-006 — L1 Frozen Guard rejects harness-learner writes to Core paths (REQ-HRA-010, REQ-HRA-011, REQ-HRA-021)

**REQ**: REQ-HRA-010, REQ-HRA-011, REQ-HRA-021 · **Milestone**: M3.1

**Verification** (8 sentinel matrix):
```bash
go test ./internal/hook/... -run TestPreToolFrozenGuardSentinelMatrix
# Per sentinel: spawn fake harness-learner PreToolUse payload writing to
# the deny path; assert JSON response includes:
#   status: "denied"
#   sentinel: HARNESS_FROZEN_<TYPE>_VIOLATION
#   agent: "harness-learner"
#   path: "<denied path>"
# Verify file NOT created on disk.
```

**PASS**: All 8 sentinels (AGENT/SKILL/RULE/COMMAND/HOOK/OUTPUTSTYLE/INSTRUCTION/CONFIG) reject correctly; my-harness-* path explicitly ALLOWED through skill exception.

---

### AC-HRA-007 — L1 latency budget <10ms p99 (R-HRA-S1)

**REQ**: REQ-HRA-010 / R-HRA-S1 · **Milestone**: M3.1

**Verification**:
```bash
go test -bench=BenchmarkFrozenGuardL1 -benchtime=100x ./internal/hook/...
# Assert: p99 < 10ms over 100 samples
# Assert: p50 < 1ms (early-exit path for non-harness agents)
```

**PASS**: benchmark gate green; latency report saved to `.moai/logs/harness/latency.jsonl`.

---

### AC-HRA-008 — L2 Canary VETO post-L5 (REQ-HRA-012)

**REQ**: REQ-HRA-012 · **Milestone**: M3.2

**Setup**: Fixture with 3 baseline projects; proposed rule causes regression in 1 (score drop 0.15).

**Verification**:
```bash
go test ./internal/harness/safety/... -run TestCanaryVetoPostL5
# Scenario:
#   1. L5 user-approves proposal
#   2. provisional file written (status: provisional)
#   3. Canary completes (~30s sim'd as immediate via test seam)
#   4. Score drop 0.15 detected (>0.10 threshold)
#   5. Automatic rollback: provisional file reverted to .bak
#   6. evolution-log.md appended with HARNESS_LEARNING_CANARY_VETO
#   7. 48h cooldown recorded in rate-limit-state.json
```

**PASS**: VETO triggers rollback + sentinel + cooldown; file system reverts atomically.

---

### AC-HRA-009 — L3 Contradiction blocker-report (REQ-HRA-013, REQ-HRA-024)

**REQ**: REQ-HRA-013, REQ-HRA-024 · **Milestone**: M3.3

**Verification**:
```bash
go test ./internal/harness/safety/... -run TestContradictionBlockerReport
# Scenario: proposed rule "use camelCase" vs existing "use snake_case"
# Expected output:
#   blocker_report.contradiction.existing_rule_id != ""
#   blocker_report.contradiction.proposed_change != ""
#   blocker_report.contradiction.recommendation != ""
# AND: NO AskUserQuestion invocation (subagent boundary)
```

**PASS**: structured blocker report emitted; subagent NEVER calls AskUserQuestion (verified via grep on test output).

---

### AC-HRA-010 — L4 Rate Limiter enforces 3/week + 24h cooldown + 50 active (REQ-HRA-015)

**REQ**: REQ-HRA-015 · **Milestone**: M3.4

**Verification**:
```bash
go test ./internal/harness/safety/... -run TestRateLimiterTripleConstraint
# Three sub-tests:
#   a. 4th evolution in 7-day window → defer
#   b. 2nd evolution <24h after last → defer
#   c. 51st active learning → defer
```

**PASS**: all 3 constraints trigger `defer` verdict; state persists across process restart.

---

### AC-HRA-011 — Proposal throttling 4 modes (REQ-HRA-016, REQ-HRA-017, REQ-HRA-018)

**REQ**: REQ-HRA-016/017/018 · **Milestone**: M4

**Verification**:
```bash
go test ./internal/harness/throttle/... -run TestThrottleAllModes
# Per mode:
#   immediate → dispatch within 1s
#   batch     → queued, dispatch at window boundary (≤max_per_window)
#   quiet     → suppressed during hours, dispatched at hour exit
#   mute      → category-suppressed, count still increments
```

**PASS**: All 4 modes behave per spec; state persists to `.moai/harness/throttle-state.json`.

---

### AC-HRA-012 — R11 timeout 7-day auto-defer (REQ-HRA-028, R-HRA-T1)

**REQ**: REQ-HRA-028 · **Milestone**: M4

**Verification**:
```bash
go test ./internal/harness/throttle/... -run TestProposalTimeout7Day
# Inject 7-day-stale proposal in queue
# Assert: auto-deferred to quiet mode + HARNESS_LEARNING_PROPOSAL_TIMEOUT logged
```

**PASS**: stale proposals auto-defer; sentinel logged.

---

### AC-HRA-013 — Seed schema + load hook contract (REQ-HRA-020, REQ-HRA-033)

**REQ**: REQ-HRA-020, REQ-HRA-033 · **Milestone**: M5 (stub only per D11)

**Verification**:
```bash
go test ./internal/harness/seeds/... -run TestSeedLoaderContractWithDummy
# Loads .claude/skills/moai-meta-harness/seeds/_dummy.yaml
# Asserts: returned []Seed has 1 element with tier=3
# Asserts: Load() signature stable for W4 consumer
```

**PASS**: contract test green; full seed library deferred to W4 per EXCL-HRA-002.

---

### AC-HRA-014 — CLI 6 verbs + 2 helpers (REQ-HRA-005)

**REQ**: REQ-HRA-005 · **Milestone**: M6

**Verification**:
```bash
go test ./internal/cli/... -run TestHarnessCLIAllVerbs
# Per verb: status, apply, rollback, disable, mute, verify, mute-list, unmute
# Assert: --json flag present, exit codes per plan §2.6 table
# Idempotency (R-HRA-I1): rollback called twice on same ID → second call exit 0 with "already rolled back"
```

**PASS**: 8 verbs all return structured JSON; rollback idempotent.

---

## 2. Edge Cases (6 total)

### EC-HRA-001 — Bypass env honored only from `moai update` chain (REQ-HRA-025)

```bash
# Direct env set without CLI chain → still rejected
MOAI_FROZEN_GUARD_BYPASS=moai-update-internal go test -run TestL1RejectsBypassWithoutCLIChain
# Real CLI chain → allowed
go test -run TestMoaiUpdateAllowedCorePath
```

### EC-HRA-002 — Malformed observation YAML degrades gracefully (REQ-HRA-026)

```bash
# Insert a malformed entry (missing 'created' field)
go test -run TestSchemaDriftSkipMalformed
# Assert: loader emits HARNESS_LEARNING_SCHEMA_DRIFT, skips entry, continues with valid entries
```

### EC-HRA-003 — Process crash mid-VETO recovery

```bash
go test -run TestCanaryVetoCrashRecovery
# Simulate: crash after .bak created but before rename
# On startup: cleanup.go detects orphan .provisional + .bak → rolls back
```

### EC-HRA-004 — Disabled sentinel halts auto-capture (REQ-HRA-019)

```bash
touch .moai/harness/disabled
go test -run TestDisabledHaltsCapture
# Trigger SubagentStop; assert: no new observation written
# Run `moai harness rollback EVO-001`; assert: still works (CLI not gated)
```

### EC-HRA-005 — my-harness-* skill path explicitly ALLOWED through skill deny

```bash
go test -run TestSkillExceptionMyHarness
# harness-learner writes to .claude/skills/my-harness-go-spec/file.md → ALLOW
# harness-learner writes to .claude/skills/moai-foundation-core/file.md → DENY
```

### EC-HRA-006 — Tier 4 dispatch during quiet hours

```bash
# Set quiet.hours: [18, 9], current time 22:00
go test -run TestTier4DuringQuietHours
# Assert: proposal queued, dispatched at 09:00 next day
```

---

## 3. Risk Mitigations (cross-ref plan §6)

| ID | Risk | Mitigation verification |
|----|------|-----------------------|
| R-HRA-01 | L1 latency creep | `BenchmarkFrozenGuardL1` gates CI at p99<10ms; deny-pattern cap=16 enforced |
| R-HRA-02 | VETO mid-revert crash | `TestCanaryVetoCrashRecovery` (EC-HRA-003) + startup cleanup invariant |
| R-HRA-03 | Subagent boundary leakage | C-HRA-008 below |
| R-HRA-04 | Brownfield test regression | Pre-EXTEND: `go test ./internal/harness/safety/... -v` baseline saved; CI gate on no test removal |
| R-HRA-05 | W4 API mismatch with seed loader | `internal/harness/seeds/CONTRACT.md` checked in; signature change requires SPEC bump |

---

## 4. Constraint AC

### C-HRA-008 — Subagent boundary grep enforcement (R-HRA-Q2)

**REQ**: REQ-HRA-029 (subagent MUST NOT invoke AskUserQuestion)

**Verification**:
```bash
# Hard grep gate — fails build if violated
test $(grep -c "AskUserQuestion" .claude/agents/moai/harness-learner.md 2>/dev/null) -eq 0
test $(grep -rc "AskUserQuestion" internal/harness/safety/*.go | awk -F: '{s+=$2} END {print s}') -eq 0
# Subagent definition + 5-layer pipeline code: zero AskUserQuestion references
```

**PASS**: harness-learner agent definition AND internal/harness/safety/ Go sources contain ZERO `AskUserQuestion` literal references. Blocker reports are the only user-input channel from subagent.

---

## 5. Traceability Matrix (AC → REQ)

| AC | REQs covered |
|----|--------------|
| AC-HRA-001 | REQ-HRA-006, REQ-HRA-007 |
| AC-HRA-002 | REQ-HRA-001 |
| AC-HRA-003 | REQ-HRA-008 |
| AC-HRA-004 | REQ-HRA-014 |
| AC-HRA-005 | REQ-HRA-009 |
| AC-HRA-006 | REQ-HRA-010, REQ-HRA-011, REQ-HRA-021 |
| AC-HRA-007 | REQ-HRA-010 (NFR R-HRA-S1, REQ-HRA-022) |
| AC-HRA-008 | REQ-HRA-012, REQ-HRA-023 |
| AC-HRA-009 | REQ-HRA-013, REQ-HRA-024 |
| AC-HRA-010 | REQ-HRA-015 |
| AC-HRA-011 | REQ-HRA-016, REQ-HRA-017, REQ-HRA-018 |
| AC-HRA-012 | REQ-HRA-028 |
| AC-HRA-013 | REQ-HRA-020, REQ-HRA-033 |
| AC-HRA-014 | REQ-HRA-005, R-HRA-I1 |
| EC-HRA-001 | REQ-HRA-025 |
| EC-HRA-002 | REQ-HRA-026 |
| EC-HRA-003 | REQ-HRA-012 (recovery aspect) |
| EC-HRA-004 | REQ-HRA-019 |
| EC-HRA-005 | REQ-HRA-010 (allowlist exception) |
| EC-HRA-006 | REQ-HRA-017 |
| C-HRA-008 | REQ-HRA-029, REQ-HRA-030, REQ-HRA-038 (R-HRA-Q2) |

Total REQs covered by binary tests: **REQ-HRA-001..030 + REQ-HRA-033 + REQ-HRA-038** = 32/38. Remaining 6 REQs (031, 032, 034, 035, 036, 037) are documentation / invariant assertions verified by plan-auditor + structural tests (e.g., REQ-HRA-035 sentinel catalog enumeration validated by C-HRA-008-style grep on plan.md §3).

---

End of acceptance.md (draft v0.1.0, awaiting plan-auditor iter1).
