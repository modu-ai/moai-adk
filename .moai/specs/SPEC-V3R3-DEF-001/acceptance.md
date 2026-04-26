---
spec_id: SPEC-V3R3-DEF-001
title: Acceptance Criteria — ORC Dependency Cycle Resolution
version: "1.0.0"
status: draft
created: 2026-04-25
related_spec: .moai/specs/SPEC-V3R3-DEF-001/spec.md
---

# Acceptance Criteria — SPEC-V3R3-DEF-001

## AC-DEF001-01: ORC-001 ~ ORC-005 dependencies가 단방향 DAG

**Given** the 5 ORC SPEC files exist under `.moai/specs/SPEC-V3R2-ORC-*/spec.md`
**When** the cycle resolution is applied
**Then** every ORC SPEC's `dependencies` field references only lower-numbered ORC SPECs and CON-001

### Verification

```bash
for n in 001 002 003 004 005; do
  echo "=== ORC-$n ==="
  awk '/^dependencies:/,/^[a-z_]+:/' .moai/specs/SPEC-V3R2-ORC-$n/spec.md | grep "ORC-"
done
# Expected output (DAG invariant):
# ORC-001: only CON-001 (no ORC- references)
# ORC-002: ORC-001 only
# ORC-003: ORC-001, ORC-002
# ORC-004: ORC-001, ORC-002
# ORC-005: ORC-001, ORC-004

# Cycle check (any ORC-N depending on ORC-M where M > N is a violation)
python3 -c "
import re
for n in range(1, 6):
    path = f'.moai/specs/SPEC-V3R2-ORC-{n:03d}/spec.md'
    with open(path) as f:
        content = f.read()
    deps_section = re.search(r'^dependencies:\n((?:  - .*\n)+)', content, re.MULTILINE)
    if deps_section:
        for line in deps_section.group(1).strip().split('\n'):
            m = re.search(r'ORC-(\d+)', line)
            if m and int(m.group(1)) >= n:
                print(f'VIOLATION: ORC-{n:03d} depends on ORC-{m.group(1)}')
print('OK')
"
# Expected: only 'OK' line
```

Maps to: REQ-DEF001-001, REQ-DEF001-004

---

## AC-DEF001-02: MIG-001 dependencies에 SPEC-V3R2-WF-001 포함

**Given** SPEC-V3R2-MIG-001/spec.md exists
**When** the cycle resolution is applied
**Then** the `dependencies` field includes `SPEC-V3R2-WF-001`

### Verification

```bash
grep -A 6 "^dependencies:" .moai/specs/SPEC-V3R2-MIG-001/spec.md | grep "SPEC-V3R2-WF-001"
# Expected: matching line found
```

Maps to: REQ-DEF001-002

---

## AC-DEF001-03: WF-001 dependencies에 MIG-001 미포함 (단방향 보장)

**Given** SPEC-V3R2-WF-001/spec.md exists
**When** the cycle resolution is applied
**Then** the `dependencies` field does NOT include `SPEC-V3R2-MIG-001`

### Verification

```bash
grep -A 5 "^dependencies:" .moai/specs/SPEC-V3R2-WF-001/spec.md | grep "MIG-001" && echo "FAIL: cycle detected" || echo "OK"
# Expected: OK
```

Maps to: REQ-DEF001-003

---

## AC-DEF001-04: 7개 SPEC 본문 무수정 (HISTORY 외)

**Given** baseline checkout of the 7 affected SPECs
**When** the cycle resolution is applied
**Then** `git diff` shows changes ONLY in HISTORY section and (for MIG-001) the `dependencies` field — no other body changes

### Verification

```bash
for s in SPEC-V3R2-ORC-001 SPEC-V3R2-ORC-002 SPEC-V3R2-ORC-003 SPEC-V3R2-ORC-004 SPEC-V3R2-ORC-005 SPEC-V3R2-WF-001 SPEC-V3R2-MIG-001; do
  echo "=== $s ==="
  git diff --stat .moai/specs/$s/spec.md
  # Manually inspect: changes confined to HISTORY section + (MIG-001 only) dependencies field
done
```

Maps to: REQ-DEF001-007

---

## AC-DEF001-05: plan-auditor 결과 D-CRIT-001 RESOLVED

**Given** the cycle resolution is applied to all 7 SPECs
**When** plan-auditor is invoked on the SPEC graph
**Then** the audit report no longer lists D-CRIT-001 as an open defect

### Verification

```
Agent(subagent_type: "plan-auditor", prompt: "Validate ORC SPEC dependency graph for cycles. Check D-CRIT-001 status.")
# Expected: report contains "D-CRIT-001: RESOLVED" or equivalent verdict
```

Maps to: REQ-DEF001-006

---

## AC-DEF001-06: 각 수정된 SPEC에 HISTORY 항목 추가

**Given** the cycle resolution is applied
**When** inspecting each of the 7 affected SPEC files
**Then** each spec.md has a new HISTORY row with date 2026-04-25 and reference to SPEC-V3R3-DEF-001

### Verification

```bash
for s in SPEC-V3R2-ORC-001 SPEC-V3R2-ORC-002 SPEC-V3R2-ORC-003 SPEC-V3R2-ORC-004 SPEC-V3R2-ORC-005 SPEC-V3R2-WF-001 SPEC-V3R2-MIG-001; do
  if grep -q "SPEC-V3R3-DEF-001" .moai/specs/$s/spec.md; then
    echo "OK: $s"
  else
    echo "MISSING HISTORY: $s"
  fi
done
# Expected: 7 OK lines
```

Maps to: REQ-DEF001-005

---

## Edge Cases

### EC-1: ORC SPEC에 이미 invariant 위반 dependency 존재

If grep reveals an existing higher-number-references-lower violation in the current state, the cycle resolution MUST report the violation explicitly and STOP — do not silently rewrite. Operator must manually approve or reject.

### EC-2: MIG-001이 이미 WF-001을 dependency로 보유

If MIG-001 already lists WF-001, the resolution MUST skip insertion (no duplicate) and proceed with HISTORY annotation only.

### EC-3: WF-001이 어떤 형태로든 MIG-001 참조

If WF-001's dependencies (or related_spec) lists MIG-001, the resolution MUST flag this as a cycle and STOP — operator review required to break the cycle.

---

## Definition of Done

- [ ] AC-DEF001-01: ORC DAG invariant verified
- [ ] AC-DEF001-02: MIG-001 dependencies includes WF-001
- [ ] AC-DEF001-03: WF-001 dependencies excludes MIG-001
- [ ] AC-DEF001-04: 7개 SPEC body unchanged (HISTORY/dependencies 외)
- [ ] AC-DEF001-05: plan-auditor D-CRIT-001 RESOLVED
- [ ] AC-DEF001-06: HISTORY 항목 7개 추가 검증
- [ ] Edge cases EC-1/2/3 처리 로직 검증
