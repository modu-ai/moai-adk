---
spec_id: SPEC-V3R3-ARCH-007
title: Acceptance Criteria — Token Circuit Breaker
version: "1.0.0"
status: draft
created: 2026-04-25
related_spec: .moai/specs/SPEC-V3R3-ARCH-007/spec.md
---

# Acceptance Criteria — SPEC-V3R3-ARCH-007

## AC-ARCH007-01: runtime.yaml 신설 (local + template), schema 일치

**Given** the runtime.yaml schema is defined in spec.md REQ-ARCH007-001
**When** the implementation is applied
**Then** both `.moai/config/sections/runtime.yaml` and `internal/template/templates/.moai/config/sections/runtime.yaml` exist with identical content matching the schema

### Verification

```bash
test -f .moai/config/sections/runtime.yaml || echo "MISSING local runtime.yaml"
test -f internal/template/templates/.moai/config/sections/runtime.yaml || echo "MISSING template runtime.yaml"
diff -q .moai/config/sections/runtime.yaml internal/template/templates/.moai/config/sections/runtime.yaml \
  && echo "OK: synced" \
  || echo "DRIFT detected"

# Required keys
for key in pre_clear_threshold hard_clear_threshold per_agent_budget circuit_breaker; do
  grep -q "$key:" .moai/config/sections/runtime.yaml || echo "MISSING key: $key"
done
# Expected: only "OK: synced" line + no MISSING
```

Maps to: REQ-ARCH007-001, REQ-ARCH007-003, REQ-ARCH007-011

---

## AC-ARCH007-02: `internal/runtime/budget.go` Tracker 5 메서드 구현

**Given** `internal/runtime/budget.go` is created
**When** inspecting the Tracker type
**Then** the type exposes the 5 methods: `RecordCall`, `Usage`, `IsApproachingLimit`, `IsAtHardLimit`, `DetectStall`, plus the `PersistProgress` helper

### Verification

```bash
test -f internal/runtime/budget.go || echo "MISSING budget.go"
for method in "RecordCall" "Usage" "IsApproachingLimit" "IsAtHardLimit" "DetectStall" "PersistProgress"; do
  grep -E "func \\(t \\*Tracker\\) $method" internal/runtime/budget.go > /dev/null \
    && echo "OK: $method" \
    || echo "MISSING method: $method"
done
# Expected: 6 OK lines

# Unit tests
go test -count=1 -race ./internal/runtime/...
# Expected: PASS
```

Maps to: REQ-ARCH007-002, REQ-ARCH007-007

---

## AC-ARCH007-03: SessionStart 훅에서 runtime.yaml 로드

**Given** the SessionStart hook handler exists
**When** a fresh session starts
**Then** the runtime.yaml is loaded and a `Tracker` instance is initialized in the session context

### Verification

```bash
# Verify hook handler imports and uses runtime package
grep -rE "runtime\\.NewTracker|runtime\\.LoadRuntime" internal/cli/ \
  && echo "OK: SessionStart hook integrates runtime" \
  || echo "MISSING: hook integration"

# Manual smoke test:
# 1. moai init test-project
# 2. cd test-project
# 3. moai hook session-start (with mock JSON input)
# Expected: log line indicating runtime.yaml load + Tracker init
```

Maps to: REQ-ARCH007-004

---

## AC-ARCH007-04: 75% 도달 시 progress.md 자동 저장 + resume message

**Given** a Tracker recording usage approaching 75% of context budget
**When** RecordCall pushes usage past the pre_clear_threshold
**Then** PersistProgress writes a progress.md file and returns a resume message in the format specified in `.claude/rules/moai/workflow/context-window-management.md` §Resume message format

### Verification (unit test based)

```bash
# Test fixture: Tracker init with mock SPEC dir, force usage to 76%, verify file write + resume string
cat > /tmp/test_persist.go <<'EOF'
// (smoke test — see budget_test.go for canonical test)
EOF
go test -run TestPersistProgressAt75Pct -v ./internal/runtime/...
# Expected: PASS with assertion on file existence + resume message format
```

```bash
# 90% trigger
go test -run TestHardLimitWarning -v ./internal/runtime/...
# Expected: PASS, warning emitted, no /clear invocation
```

Maps to: REQ-ARCH007-005, REQ-ARCH007-012

---

## AC-ARCH007-05: stall detection 60s + retry max 3

**Given** a Tracker with stall_detection_seconds=60 and retry_max=3
**When** no RecordCall arrives for the agent within 60s
**Then** DetectStall returns true; after 3 stall events, fallback recommendation `split_into_waves` is emitted

### Verification

```bash
go test -run TestDetectStall -v ./internal/runtime/...
go test -run TestRetryMaxFallback -v ./internal/runtime/...
# Expected: PASS, fallback recommendation logged
```

Maps to: REQ-ARCH007-006

---

## AC-ARCH007-06: /clear 자동 트리거 부재 검증

**Given** the implementation is complete
**When** searching the runtime package for any auto-clear invocation
**Then** no code path triggers `/clear` automatically (HARD constraint)

### Verification

```bash
grep -rE "exec\\.Command\\(.*clear|os/exec.*clear|/clear" internal/runtime/ \
  && echo "FAIL: /clear invocation detected" \
  || echo "OK: no /clear auto-trigger"

# Also: grep for accidental "clear" string usage
grep -rn "\"/clear\"" internal/runtime/ && echo "FAIL: literal /clear found"
# Expected: no FAIL
```

Maps to: REQ-ARCH007-009

---

## AC-ARCH007-07: BC-V3R3-006 warning-first 동작 검증

**Given** an agent exceeds its per_agent_budget
**When** the Tracker records the over-budget call
**Then** a WARN-level log message is emitted naming the agent and budget; the agent execution is NOT blocked

### Verification

```bash
go test -run TestPerAgentBudgetOverWarning -v ./internal/runtime/...
# Expected: PASS, warning emitted, no error returned

# CHANGELOG entry
grep "BC-V3R3-006" CHANGELOG.md && echo "OK: BC-V3R3-006 documented" || echo "MISSING"
```

Maps to: REQ-ARCH007-008

---

## Edge Cases

### EC-1: runtime.yaml 부재
If `.moai/config/sections/runtime.yaml` is missing at SessionStart, Tracker MUST initialize with built-in defaults (REQ-ARCH007-011) without erroring out.

```bash
go test -run TestDefaultsWhenConfigMissing -v ./internal/runtime/...
```

### EC-2: SPEC 디렉터리 부재 시 progress.md 저장 시도
If `.moai/specs/<SPEC-ID>/` does not exist when PersistProgress is called, the function MUST silent-skip (no error), log a debug message.

### EC-3: per_agent_budget에 미정의 agent
If RecordCall is called with an agent name not in per_agent_budget, the Tracker MUST use `default` budget value.

### EC-4: 동시 호출 race
Tracker MUST be goroutine-safe. Multiple RecordCall from different goroutines for the same agent MUST not corrupt usage counter.

```bash
go test -race -run TestConcurrentRecordCall -v ./internal/runtime/...
```

---

## Definition of Done

- [ ] AC-ARCH007-01: runtime.yaml local + template 동기화
- [ ] AC-ARCH007-02: Tracker 5 메서드 + PersistProgress 구현, unit test PASS
- [ ] AC-ARCH007-03: SessionStart 통합 grep 통과 + 수동 smoke test
- [ ] AC-ARCH007-04: 75% trigger 동작 + 90% warning 동작
- [ ] AC-ARCH007-05: stall detection + retry max 통과
- [ ] AC-ARCH007-06: /clear 자동 트리거 부재 grep 검증
- [ ] AC-ARCH007-07: warning-first BC-V3R3-006 동작 + CHANGELOG entry
- [ ] Edge cases EC-1/2/3/4 모두 처리
- [ ] `make build && make install` 성공, `go test ./...` 회귀 통과
