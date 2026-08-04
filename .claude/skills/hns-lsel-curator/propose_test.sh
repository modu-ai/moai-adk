#!/usr/bin/env bash
# propose_test.sh — AC-LSEL-011 PROPOSE shadow payload schema validation
# (SPEC-LSEL-LOCAL-EVOLUTION-001 M2).
#
# Validates that a shadow proposal at .moai/state/lsel/proposals/<id>/
# contains {proposal.md, diff.patch, self-critique.md} AND proposal.md
# carries the full payload schema AND retrieval-before-propose is evidenced
# AND a self-critique with an unresolved objection blocks APPROVE.
#
# RED: run before the sample proposal exists → fails (files absent).
# GREEN: after .moai/state/lsel/proposals/lsel-001/ is authored → passes.
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$PROJECT_ROOT"

PROPOSALS_DIR=".moai/state/lsel/proposals"
SAMPLE_ID="lsel-001"
SAMPLE_DIR="$PROPOSALS_DIR/$SAMPLE_ID"

FAIL=0
log() { printf '%s\n' "$*"; }
fail() { log "FAIL: $*"; FAIL=1; }
pass() { log "PASS: $*"; }

log "=== AC-LSEL-011 PROPOSE shadow payload schema validation ==="

# (a) the trio of files exists
for f in proposal.md diff.patch self-critique.md; do
  if [[ -f "$SAMPLE_DIR/$f" ]]; then
    pass "$f present"
  else
    fail "$SAMPLE_DIR/$f absent"
  fi
done

# (b) proposal.md carries the full payload schema (8 required keys)
if [[ -f "$SAMPLE_DIR/proposal.md" ]]; then
  for key in proposal_id target_surface rationale WHY-not-just-WHAT prediction verify_command blast_radius memory_type; do
    if grep -q "^${key}:" "$SAMPLE_DIR/proposal.md"; then
      pass "schema key '$key' present"
    else
      fail "schema key '$key' missing from proposal.md"
    fi
  done
  # memory_type value restricted to semantic|procedural|episodic
  mt=$(grep "^memory_type:" "$SAMPLE_DIR/proposal.md" | head -1 | sed 's/.*: *//')
  case "$mt" in
    semantic|procedural|episodic) pass "memory_type '$mt' valid" ;;
    *) fail "memory_type '$mt' not in {semantic,procedural,episodic}" ;;
  esac
fi

# (c) retrieval-before-propose evidenced (retrieval_evidence block names feedback_*.md files)
if [[ -f "$SAMPLE_DIR/proposal.md" ]]; then
  if grep -q "retrieval_evidence:" "$SAMPLE_DIR/proposal.md" && \
     grep -A4 "retrieval_evidence:" "$SAMPLE_DIR/proposal.md" | grep -q "feedback_"; then
    pass "retrieval-before-propose evidenced (retrieval_evidence block cites feedback_*.md)"
  else
    fail "retrieval-before-propose NOT evidenced in proposal.md"
  fi
fi

# (d) self-critique with an UNRESOLVED objection blocks APPROVE
#     The sample self-critique.md MUST carry an UNRESOLVED objection, and the
#     proposal status MUST be 'blocked' (not 'approved'). This proves the gate fires.
if [[ -f "$SAMPLE_DIR/self-critique.md" ]]; then
  if grep -qi "UNRESOLVED" "$SAMPLE_DIR/self-critique.md"; then
    pass "self-critique carries an UNRESOLVED objection"
  else
    fail "self-critique.md has no UNRESOLVED objection (gate-fire not demonstrated)"
  fi
fi
if [[ -f "$SAMPLE_DIR/proposal.md" ]]; then
  if grep -q "^status: *blocked" "$SAMPLE_DIR/proposal.md"; then
    pass "proposal status=blocked (unresolved objection blocks APPROVE)"
  else
    fail "proposal is NOT blocked despite the unresolved-objection requirement"
  fi
fi

if [[ "$FAIL" -ne 0 ]]; then
  log "propose_test: schema validation FAILED"
  exit 1
fi
log "propose_test: PASS"
exit 0
