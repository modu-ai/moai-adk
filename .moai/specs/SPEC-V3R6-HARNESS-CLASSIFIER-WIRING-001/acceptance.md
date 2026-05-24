---
id: SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001
title: "V3R4 Harness Classifier Runtime Wiring — Acceptance Criteria"
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Evolution Loop Closure"
module: ".claude/skills/moai/workflows/harness.md, internal/cli/hook.go (Option A) or internal/hook/post_tool.go (Option B)"
lifecycle: spec-anchored
tags: "harness, classifier, wiring, runtime, v3r6, tier-s-minimal, acceptance"
---

# Acceptance Criteria — SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001

## §A. Overview

Binary acceptance criteria for the V3R4 Harness Classifier Runtime Wiring. Each AC has a single binary outcome (PASS / FAIL) verifiable by a deterministic command or observable artifact.

REQ → AC mapping table:

| REQ ID | AC IDs | Mandatory? |
|--------|--------|------------|
| REQ-HCW-001 | AC-HCW-001, AC-HCW-005 | Yes |
| REQ-HCW-002 | AC-HCW-002 | Yes |
| REQ-HCW-003 | AC-HCW-003 | Yes |
| REQ-HCW-004 | AC-HCW-004 | Yes |
| REQ-HCW-005 | (none — Optional MAY, deferred) | No |
| (lint envelope) | AC-HCW-006 | Optional |

## §B. Mandatory Acceptance Criteria

### AC-HCW-001 (maps REQ-HCW-001)

**Statement**: After `/moai:harness status` invocation, the file `.moai/harness/learning-history/tier-promotions.jsonl` exists and `wc -l` returns a value `> 0`.

**Verification command**:
```bash
/moai:harness status  # via workflow body
test -f .moai/harness/learning-history/tier-promotions.jsonl && \
  test "$(wc -l < .moai/harness/learning-history/tier-promotions.jsonl)" -gt 0
echo "Exit: $?"  # 0 = PASS
```

**Pre-condition**: `.moai/harness/usage-log.jsonl` exists with at least 1 event (verified on this project: 97 events present).

**Binary outcome**: Exit code 0 from the test command = PASS. Any other = FAIL.

---

### AC-HCW-002 (maps REQ-HCW-002)

**Statement**: The count of unique patterns in `tier-promotions.jsonl` matches the count of unique patterns in `usage-log.jsonl`, using the keying convention from `REQ-HRN-FND-010` (key = `event_type + subject + context_hash`).

**Verification command**:
```bash
USAGE_UNIQUE=$(jq -r '"\(.event_type)|\(.subject)|\(.context_hash)"' .moai/harness/usage-log.jsonl | sort -u | wc -l)
PROMO_UNIQUE=$(jq -r '"\(.event_type)|\(.subject)|\(.context_hash)"' .moai/harness/learning-history/tier-promotions.jsonl | sort -u | wc -l)
test "$USAGE_UNIQUE" -eq "$PROMO_UNIQUE"
echo "Usage unique: $USAGE_UNIQUE, Promo unique: $PROMO_UNIQUE, Match: $?"  # 0 = PASS
```

**Binary outcome**: Counts match (test exit 0) = PASS. Mismatch = FAIL.

---

### AC-HCW-003 (maps REQ-HCW-003)

**Statement**: When a corrupt entry is injected into `usage-log.jsonl`, the `status` verb renders an error annotation in its output AND continues rendering remaining sections AND exits with code 0.

**Verification command**:
```bash
# Inject corrupt entry (malformed JSON)
echo "{this is not valid json" >> .moai/harness/usage-log.jsonl

# Invoke status verb and capture output + exit code
STATUS_OUTPUT=$(/moai:harness status 2>&1)
STATUS_EXIT=$?

# Cleanup (remove corrupt entry)
sed -i '' '$d' .moai/harness/usage-log.jsonl

# Verify all 3 conditions
echo "$STATUS_OUTPUT" | grep -q "classifier error\|⚠"  # condition 1: error annotation rendered
ANNOTATION_OK=$?
echo "$STATUS_OUTPUT" | grep -q "usage-log\|Tier Distribution"  # condition 2: other sections rendered
SECTIONS_OK=$?
test "$STATUS_EXIT" -eq 0  # condition 3: exit code 0
EXIT_OK=$?

test "$ANNOTATION_OK" -eq 0 -a "$SECTIONS_OK" -eq 0 -a "$EXIT_OK" -eq 0
echo "Final: $?"  # 0 = PASS
```

**Binary outcome**: All 3 conditions pass = PASS. Any single failure = FAIL.

**Cleanup contract**: The verification command MUST restore `usage-log.jsonl` to its pre-injection state (remove the corrupt entry). Failure to clean up is a verification-script defect, not an AC failure.

---

### AC-HCW-004 (maps REQ-HCW-004)

**Statement**: When `.moai/harness/harness.yaml` has `learning.enabled: false`, the `status` verb invocation does NOT modify `usage-log.jsonl` AND does NOT modify `tier-promotions.jsonl`.

**Verification command**:
```bash
# Snapshot pre-state
USAGE_HASH_BEFORE=$(shasum .moai/harness/usage-log.jsonl | awk '{print $1}')
PROMO_HASH_BEFORE=$(shasum .moai/harness/learning-history/tier-promotions.jsonl 2>/dev/null | awk '{print $1}' || echo "ABSENT")

# Temporarily disable learning
cp .moai/harness/harness.yaml .moai/harness/harness.yaml.bak
sed -i '' 's/learning:.*enabled:.*true/learning:\n  enabled: false/' .moai/harness/harness.yaml
# (or use yq if available for safer edit)

# Invoke status verb
/moai:harness status >/dev/null 2>&1

# Snapshot post-state
USAGE_HASH_AFTER=$(shasum .moai/harness/usage-log.jsonl | awk '{print $1}')
PROMO_HASH_AFTER=$(shasum .moai/harness/learning-history/tier-promotions.jsonl 2>/dev/null | awk '{print $1}' || echo "ABSENT")

# Restore config
mv .moai/harness/harness.yaml.bak .moai/harness/harness.yaml

# Verify both files unchanged
test "$USAGE_HASH_BEFORE" = "$USAGE_HASH_AFTER" -a "$PROMO_HASH_BEFORE" = "$PROMO_HASH_AFTER"
echo "Usage unchanged: $([ "$USAGE_HASH_BEFORE" = "$USAGE_HASH_AFTER" ] && echo YES || echo NO)"
echo "Promo unchanged: $([ "$PROMO_HASH_BEFORE" = "$PROMO_HASH_AFTER" ] && echo YES || echo NO)"
echo "Final: $?"  # 0 = PASS
```

**Binary outcome**: Both hashes match between before/after = PASS. Any modification = FAIL.

---

### AC-HCW-005 (maps REQ-HCW-001..005 envelope)

**Statement**: Full Go test suite passes with zero regression after wiring insertion.

**Verification command**:
```bash
go test ./internal/harness/... ./internal/hook/... ./internal/cli/...
echo "Exit: $?"  # 0 = PASS
```

**Binary outcome**: Exit code 0 (all tests pass) = PASS. Any test failure = FAIL.

**Baseline note**: This AC requires zero new test regressions. Pre-existing failures (e.g., sibling SPEC `TEMPLATE-MIRROR-DRIFT-001` baseline failures per L46) MUST be confirmed pre-existing via `git diff` attribution check before run-phase. If pre-existing, document attribution in `progress.md §Run-phase Evidence` and treat as PASS-WITH-DEBT per L46 attribution discipline.

---

## §C. Optional Acceptance Criterion

### AC-HCW-006 (Optional, lint envelope)

**Statement**: `golangci-lint run --timeout=2m` reports `0 issues.` AND `go vet ./...` reports zero output.

**Verification command**:
```bash
golangci-lint run --timeout=2m
LINT_EXIT=$?

VET_OUTPUT=$(go vet ./... 2>&1)
VET_EXIT=$?

test "$LINT_EXIT" -eq 0 -a -z "$VET_OUTPUT" -a "$VET_EXIT" -eq 0
echo "Final: $?"  # 0 = PASS
```

**Binary outcome**: Both lint and vet clean = PASS. Any issue = FAIL.

**Optional rationale**: AC-HCW-006 is Optional because workflow-body-only edits (Bash insertion in `.claude/skills/moai/workflows/harness.md`) do not exercise Go lint paths. If Option B is chosen (in-process `post_tool.go` edit), this AC effectively becomes mandatory because Go LOC ≥80 is introduced. If Option A is chosen and the hook subcommand is the only Go addition (~80-150 LOC), this AC remains Optional but recommended.

---

## §D. Verification Batch Strategy

All ACs are read-only verification (do not mutate the worktree state irreversibly — AC-003 and AC-004 use snapshot-and-restore patterns). They MAY be issued in parallel within a single orchestrator response turn per HARD batching rule (`.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution).

Recommended batch order (parallel):
- AC-HCW-001 (file existence + line count)
- AC-HCW-002 (jq aggregation comparison)
- AC-HCW-003 (corrupt entry injection + cleanup)
- AC-HCW-004 (config gate snapshot+restore)
- AC-HCW-005 (`go test`)
- AC-HCW-006 (`golangci-lint` + `go vet`)

Note: AC-003 and AC-004 each have their own state mutation + cleanup contract. They MAY interfere if issued concurrently against the same `usage-log.jsonl`. Run AC-003 and AC-004 sequentially (one at a time), but AC-001, 002, 005, 006 may run in parallel with the sequential AC-003 → AC-004 chain.

## §E. Definition of Done (DoD)

The SPEC is considered complete (status `implemented`) when:
1. All 5 mandatory ACs (HCW-001..005) PASS
2. AC-HCW-006 (Optional) PASS OR documented exception in `progress.md`
3. Workflow body §2.1 includes the new wiring (visually verifiable in `.claude/skills/moai/workflows/harness.md` diff)
4. New hook subcommand (Option A) OR PostToolUse batching (Option B) implemented and tested
5. `tier-promotions.jsonl` non-empty after first `status` invocation on this project (smoke test)
6. CHANGELOG `[Unreleased] ### Fixed` entry added per B12 standing-rule guard
7. 4 frontmatter status fields transition `draft → implemented` (spec.md, plan.md, acceptance.md, progress.md)

## §F. AC Decision Rule (Optional MAY without AC coverage)

REQ-HCW-005 is **Optional MAY** without dedicated AC coverage. Its absence from AC is itself the contract per the canonical decision rule:
- Mandatory REQs (HCW-001..004) MUST have direct AC coverage
- Optional MAY REQs (HCW-005) MAY exist without AC; their absence signals future-SPEC scope

Future SPECs (e.g., `SPEC-V3R6-HARNESS-INCREMENTAL-001`) may add ACs that exercise REQ-HCW-005's MAY clause (e.g., PostToolUse incremental cadence ACs).

## §G. Cross-references

- `spec.md §D` — EARS REQ-HCW-001..005 statements
- `plan.md §6` — 7-item verification batch
- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution — HARD batching rule
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — verification grouping pattern
