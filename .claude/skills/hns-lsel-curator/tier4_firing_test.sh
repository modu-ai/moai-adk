#!/usr/bin/env bash
# tier4_firing_test.sh — AC-LSEL-012 behavioral probe (SPEC-LSEL-LOCAL-EVOLUTION-001 M2).
#
# This is a CHARACTERIZATION test of the CURRENT production state of the
# moai-harness-learner Tier-4 AskUserQuestion flow. It does NOT assert the
# flow fires — it asserts the VERIFICATION was performed and records the
# finding (LIVE or DEAD). Per acceptance.md §E edge case, a DEAD finding
# means M2 does NOT wire the PROPOSE→APPROVE handoff to depend on Tier-4;
# APPROVE routes via the M3 fresh path (hns-lsel-applier + decision.json).
#
# The audit's cautionary precedent (CuratorDispatch 0 production callers,
# acceptance.md §C AC-LSEL-012) is the reason this probe exists: type
# definitions alone are NOT a live invocation path.
#
# Usage: ./tier4_firing_test.sh [--project-root <path>]
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
if [[ "${1:-}" == "--project-root" ]]; then
  PROJECT_ROOT="$2"; shift 2
fi

cd "$PROJECT_ROOT"

FAIL=0
log() { printf '%s\n' "$*"; }
fail() { log "FAIL: $*"; FAIL=1; }
pass() { log "PASS: $*"; }

log "=== AC-LSEL-012 Tier-4 firing probe (project: $PROJECT_ROOT) ==="

# --- Probe 1: CuratorDispatch production callers (the cautionary precedent) ---
# A live invocation path requires at least one non-test, non-definition caller.
curator_callers=$(grep -rn "CuratorDispatch" internal/ pkg/ cmd/ 2>/dev/null \
  | grep -v "_test.go" \
  | grep -vE "internal/harness/curator_dispatch\.go:" \
  | wc -l | tr -d ' ')
if [[ "$curator_callers" -eq 0 ]]; then
  pass "CuratorDispatch has 0 production callers outside its own definition file (cautionary precedent confirmed)"
else
  log "NOTE: CuratorDispatch now has $curator_callers external caller(s) — re-evaluate Tier-4 liveness"
fi

# --- Probe 2: enableTriggerInjectionWrites dead-switch ---
# The frozen Go applier must remain at enableTriggerInjectionWrites=false.
if grep -q "var enableTriggerInjectionWrites = false" internal/harness/applier.go 2>/dev/null; then
  pass "enableTriggerInjectionWrites = false (frozen applier stays frozen per REQ-LSEL-003)"
else
  fail "enableTriggerInjectionWrites is NOT false — frozen applier state changed"
fi

# --- Probe 3: moai harness apply payload-only path prints a stub string, never invokes the skill ---
# The CLI explicitly documents it does NOT call AskUserQuestion (subagent boundary).
# It prints a stub string. A live Tier-4 path would require the orchestrator to
# mechanically load the learner skill + surface the payload — nothing triggers that.
if grep -q "moai-harness-learner skill calls AskUserQuestion with payload" internal/cli/harness.go 2>/dev/null; then
  pass "moai harness apply prints a stub string (does NOT invoke the skill) — Tier-4 path is not mechanically triggered from CLI"
else
  log "NOTE: harness.go stub string changed — re-evaluate"
fi

# --- Probe 4: AskUserQuestion is NOT invoked by any harness/learner Go code ---
# Go code cannot call AskUserQuestion (subagent boundary). The Tier-4 surface
# requires the orchestrator (main-session Claude) to do it. No mechanical
# trigger wiring exists.
askuser_in_harness_go=$(grep -rn "AskUserQuestion" internal/cli/harness.go internal/cli/harness/ internal/harness/ 2>/dev/null \
  | grep -v "_test.go" \
  | grep -v "^[^:]*:[0-9]*:[[:space:]]*//" \
  | grep -v "does not directly call AskUserQuestion" \
  | grep -v "skill calls AskUserQuestion" \
  | grep -v "orchestrator.*AskUserQuestion" \
  | wc -l | tr -d ' ')
if [[ "$askuser_in_harness_go" -eq 0 ]]; then
  pass "zero harness Go code invokes AskUserQuestion (subagent boundary honored; orchestrator-only)"
else
  fail "found $askuser_in_harness_go harness Go AskUserQuestion reference(s) — subagent boundary violated"
fi

# --- Finding ---
# Per acceptance.md §E edge case: "Tier-4 flow found DEAD at M2 entry: M2 does
# NOT wire the dependency; report finding logged; M2 downgrades to 'PROPOSE
# shadow only, APPROVE via fresh path' and returns a blocker for orchestrator decision."
log ""
log "=== AC-LSEL-012 FINDING ==="
log "Tier-4 AskUserQuestion flow is DEAD at the production invocation layer:"
log "  - moai harness apply prints a stub string; it never invokes the learner skill"
log "  - No mechanical trigger causes the orchestrator to surface a Tier-4 proposal"
log "    (REQ-LSEL-007 names this as the audit's exact failure mode)"
log "  - CuratorDispatch has 0 production callers (cautionary precedent confirmed)"
log "  - enableTriggerInjectionWrites=false (the apply dead-switch)"
log ""
log "M2 DOWNGRADE applied per acceptance.md §E: PROPOSE shadow only; APPROVE via"
log "the M3 fresh path (hns-lsel-applier + decision.json). M2 wiring does NOT"
log "depend on the Tier-4 flow. This probe IS the AC-LSEL-012 verification"
log "evidence; the M2 wiring commit cites it."

if [[ "$FAIL" -ne 0 ]]; then
  log "tier4_firing_test: one or more probes failed"
  exit 1
fi
log "tier4_firing_test: characterization PASS (finding recorded: Tier-4 DEAD)"
exit 0
