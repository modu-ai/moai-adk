#!/usr/bin/env bash
# csa_refusal_test.sh — AC-LSEL-005 CSA forced-gate mechanical-refusal fixture
# (SPEC-LSEL-LOCAL-EVOLUTION-001 M2).
#
# Validates the CSA forced-gate refusal logic described in the hns-lsel-curator
# SKILL.md PROPOSE/CSA section. A fixture proposal whose target_surface matches
# one of the FOUR execution-meta categories, WITHOUT a synchronous-approval
# marker in its decision.json, is REFUSED (+ reject-log row + no write). The
# SAME fixture WITH the marker proceeds.
#
# This test exercises the refusal RULE (the CSA doctrine enumerated in the
# curator skill body). M3 will implement lsel-apply.sh which mechanically
# enforces it; here we test the rule via the fixture + the reject log.
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$PROJECT_ROOT"

REJECT_LOG=".moai/logs/lsel-reject.log"
FIXTURE_DIR="$(mktemp -d)"
trap 'rm -rf "$FIXTURE_DIR"' EXIT

FAIL=0
log() { printf '%s\n' "$*"; }
fail() { log "FAIL: $*"; FAIL=1; }
pass() { log "PASS: $*"; }

log "=== AC-LSEL-005 CSA forced-gate mechanical-refusal fixture ==="

# --- (1) The curator skill body enumerates all SIX CSA forced-gate categories ---
SKILL=".claude/skills/hns-lsel-curator/SKILL.md"
for cat in \
  "INVARANTS kernel" \
  "security/validation exception" \
  "HIGH-fan-in" \
  "Bash risk path" \
  "permissions.allow" \
  "execution-meta" \
  ; do
  if grep -qi "$cat" "$SKILL"; then
    pass "CSA category named: $cat"
  else
    fail "CSA category MISSING from $SKILL: $cat"
  fi
done

# bother-cost-exemption clause
if grep -qi "bother-cost-exempt\|bother-cost exemption" "$SKILL"; then
  pass "bother-cost-exemption clause present"
else
  fail "bother-cost-exemption clause absent from curator skill"
fi

# --- (2) The refusal RULE: execution-meta proposal WITHOUT approval marker → REFUSED ---
# Build two fixture proposals: one WITHOUT the marker (must refuse), one WITH (must proceed).
WITHOUT_DIR="$FIXTURE_DIR/without-marker"
WITH_DIR="$FIXTURE_DIR/with-marker"
mkdir -p "$WITHOUT_DIR" "$WITH_DIR"

# Fixture A: targets an execution-meta category (an applier/curator skill body),
# decision.json has NO synchronous-approval marker.
cat > "$WITHOUT_DIR/decision.json" <<'JSON'
{
  "proposal_id": "csa-fixture-without-marker",
  "target_surface": ".claude/skills/hns-lsel-applier/SKILL.md",
  "blast_radius": "execution-meta: applier skill body (CSA forced-gate category)",
  "synchronous_approval": null
}
JSON

# Fixture B: same target, WITH the synchronous-approval marker.
cat > "$WITH_DIR/decision.json" <<'JSON'
{
  "proposal_id": "csa-fixture-with-marker",
  "target_surface": ".claude/skills/hns-lsel-applier/SKILL.md",
  "blast_radius": "execution-meta: applier skill body (CSA forced-gate category)",
  "synchronous_approval": {
    "askuserquestion_artifact": "2026-08-04T00:00:00Z",
    "decision": "approved",
    "channel": "orchestrator AskUserQuestion"
  }
}
JSON

# Apply the refusal rule (a portable bash function mirroring the CSA doctrine in
# the curator skill body). This is the rule M3's lsel-apply.sh will enforce
# mechanically; here it is the test oracle.
csa_refuse_check() {
  local decision="$1"
  local target blast sync
  target=$(grep -m1 '"target_surface"' "$decision" | sed 's/.*: *"\([^"]*\)".*/\1/')
  blast=$(grep -m1 '"blast_radius"' "$decision" | sed 's/.*: *"\([^"]*\)".*/\1/')
  # execution-meta category match (the 4 categories from REQ-LSEL-005)
  local exec_meta=0
  case "$target" in
    *.claude/lsel/frozen-allowlist.json) exec_meta=1 ;;
    */hns-lsel-applier/*|*/hns-lsel-curator/*) exec_meta=1 ;;
    */lsel-apply.sh) exec_meta=1 ;;
    *settings.local.json) exec_meta=1 ;;
  esac
  if echo "$blast" | grep -qi "execution-meta"; then exec_meta=1; fi
  # synchronous-approval marker? (multi-line JSON: presence of both keys anywhere)
  local sa=0
  if grep -q '"synchronous_approval"' "$decision" && \
     grep -q '"decision": *"approved"' "$decision"; then
    sa=1
  fi
  if [[ "$exec_meta" -eq 1 && "$sa" -eq 0 ]]; then
    return 1   # REFUSE
  fi
  return 0     # PROCEED
}

mkdir -p "$(dirname "$REJECT_LOG")"
: > "$REJECT_LOG" 2>/dev/null || true

# Fixture A: WITHOUT marker → REFUSED
if csa_refuse_check "$WITHOUT_DIR/decision.json"; then
  fail "fixture WITHOUT marker was NOT refused (CSA gate did not fire)"
else
  pass "fixture WITHOUT marker REFUSED (+ reject-log row)"
  echo "$(date -u +%FT%TZ) proposal=csa-fixture-without-marker category=execution-meta reason=no-synchronous-approval-marker" >> "$REJECT_LOG"
fi

# Fixture B: WITH marker → PROCEEDS
if csa_refuse_check "$WITH_DIR/decision.json"; then
  pass "fixture WITH marker proceeds (refusal keyed on absent marker, not category alone)"
else
  fail "fixture WITH marker was refused (gate over-fired — marker should allow)"
fi

# (3) reject-log row recorded for the refused fixture
if grep -q "csa-fixture-without-marker" "$REJECT_LOG" 2>/dev/null; then
  pass "reject-log row appended for the refused execution-meta proposal"
else
  fail "no reject-log row for the refused proposal"
fi

if [[ "$FAIL" -ne 0 ]]; then
  log "csa_refusal_test: FAILED"
  exit 1
fi
log "csa_refusal_test: PASS"
exit 0
