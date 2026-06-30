#!/bin/bash
# IGGDA Independent-Audit Preservation Regression Guard (M4, D5).
#
# Verifies the independent-audit preservation invariants remain documented in
# orchestration-mode-selection.md §J.
# The fresh-context spawn pattern itself is architecturally enforced by the
# agent catalog (Agent(subagent_type: "plan-auditor") IS a fresh context by
# definition — subagents run in isolated contexts). This guard verifies the
# DOCUMENTATION CONTRACT holds: the invariants are present, the self-audit vs
# independent-audit disambiguation is present, and the FAIL/INCONCLUSIVE halt
# is documented.
#
# This is a documentation-contract guard, NOT a runtime enforcement. The
# runtime enforcement is the agent catalog architecture (fresh-context spawn)
# + the Stop hook driver (iggda-phase-driver.sh halts on MUST-FIX drift).
#
# Exit codes:
#   0 = all invariants documented (PASS)
#   1 = one or more invariants missing (FAIL — regression detected)
#
# Run: bash .claude/hooks/moai/iggda-audit-preservation-guard.sh
# (or via the orchestrator's verification-batch at run-phase completion)

set -u

RULE_FILE="${CLAUDE_PROJECT_DIR:-.}/.claude/rules/moai/workflow/orchestration-mode-selection.md"
FAIL=0

if [ ! -f "$RULE_FILE" ]; then
    echo "FAIL: $RULE_FILE not found" >&2
    exit 1
fi

# Fresh-context spawn documented for both auditors.
# The §J.1 section must reference BOTH plan-auditor AND sync-auditor fresh-context
# spawn (NOT continuation of implementer turn).
if ! grep -q "plan-auditor.*Phase 1.*sync-auditor.*Phase 3\|sync-auditor.*Phase 3.*plan-auditor.*Phase 1" "$RULE_FILE" 2>/dev/null; then
    if ! grep -q "plan-auditor (Phase 1)" "$RULE_FILE" || ! grep -q "sync-auditor (Phase 3)" "$RULE_FILE"; then
        echo "FAIL: §J.1 fresh-context spawn for both plan-auditor (Phase 1) and sync-auditor (Phase 3) not documented" >&2
        FAIL=1
    fi
fi

# FAIL/INCONCLUSIVE verdict halts auto-advance (hard stop).
if ! grep -qi "FAIL/INCONCLUSIVE.*halt\|halts auto-advance" "$RULE_FILE"; then
    echo "FAIL: §J.2 FAIL/INCONCLUSIVE halt not documented" >&2
    FAIL=1
fi

# Self-audit vs independent-audit disambiguation present.
if ! grep -qi "Self-audit.*Independent audit\|self-audit.*independent audit" "$RULE_FILE"; then
    echo "FAIL: §J.3 self-audit vs independent-audit disambiguation not documented" >&2
    FAIL=1
fi

# The two audit types documented as COMPLEMENTARY (not interchangeable).
if ! grep -qi "COMPLEMENTARY" "$RULE_FILE"; then
    echo "FAIL: §J.3 complementary (not interchangeable) framing not documented" >&2
    FAIL=1
fi

if [ "$FAIL" -eq 0 ]; then
    echo "PASS: IGGDA independent-audit preservation invariants documented (§J.1 fresh-context spawn, §J.2 FAIL halt, §J.3 disambiguation, complementary framing)"
    exit 0
else
    echo "FAIL: IGGDA independent-audit preservation regression detected — one or more §J invariants missing from $RULE_FILE" >&2
    exit 1
fi
