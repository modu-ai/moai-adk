#!/usr/bin/env bash
# reflect_test.sh — SPEC-LSEL-LOCAL-EVOLUTION-001 M4 REFLECTION characterization test.
#
# Exercises the REAL reflect.sh mechanism inside an hermetic temp memory dir (no
# pollution of the host worktree or the real auto-memory store). Asserts AC-LSEL-016:
#   - ≥3 concrete topic files above the reflection threshold → 1 principle synthesized
#   - originals moved to memory/_archive/ (NOT deleted)
#   - each new principle carries a memory_type label
#   - retrieval probe for a related cue ranks the principle ABOVE the archived originals
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
REFLECT="$PROJECT_ROOT/.claude/skills/hns-lsel-curator/reflect.sh"

FAIL=0
log(){ printf '%s\n' "$*"; }
fail(){ log "FAIL: $*"; FAIL=1; }
pass(){ log "PASS: $*"; }

# --- RED guard: mechanism must exist ---
if [[ ! -f "$REFLECT" ]]; then
  log "RED: reflect.sh absent at $REFLECT (mechanism not built yet)"
  exit 1
fi

# =====================================================================
log "=== AC-LSEL-016 fixture: ≥3 concrete topics above threshold → 1 principle ==="
MEM="$(mktemp -d)"; trap 'rm -rf "$MEM"' EXIT

# Three concrete topic files whose accumulated importance (60+60+60=180) clears
# the default reflection threshold of 150. All share a concrete theme so the
# synthesized principle can be probed by a related cue.
for n in 1 2 3; do
cat > "$MEM/feedback_topic_${n}.md" <<MD
---
name: topic-${n}
description: concrete observation about bash timeout noise in drain stub ${n}
type: feedback
importance: 60
---

Concrete topic ${n}: recurring Bash UnknownFailure stub ${n}.
**Why:** timeout/sandbox noise floods the inbox.
**How to apply:** filter drain-side before clustering.
MD
done

# Run reflection (default threshold 150, min-topics 3).
bash "$REFLECT" --memory-dir "$MEM" --threshold 150 --min-topics 3 >/tmp/r1.out 2>&1
RC=$?
if [[ $RC -ne 0 ]]; then
  fail "reflect.sh exited non-zero ($RC): $(cat /tmp/r1.out)"
fi

# (a) exactly ONE principle file synthesized in the active set.
PRINCIPLE_COUNT=$(find "$MEM" -maxdepth 1 -name 'feedback_*principle*.md' | wc -l | tr -d ' ')
if [[ "$PRINCIPLE_COUNT" -eq 1 ]]; then
  pass "one principle file synthesized (got $PRINCIPLE_COUNT)"
else
  fail "expected 1 synthesized principle, got $PRINCIPLE_COUNT"
fi
PRINCIPLE="$(find "$MEM" -maxdepth 1 -name 'feedback_*principle*.md' | head -1)"

# (b) originals moved to _archive/ (NOT deleted).
ARCHIVED_COUNT=$(find "$MEM/_archive" -name 'feedback_topic_*.md' 2>/dev/null | wc -l | tr -d ' ')
if [[ "$ARCHIVED_COUNT" -eq 3 ]]; then
  pass "all 3 originals archived under memory/_archive/ (cold tier, not deleted)"
else
  fail "expected 3 originals in _archive/, got $ARCHIVED_COUNT"
fi
STILL_ACTIVE=$(find "$MEM" -maxdepth 1 -name 'feedback_topic_*.md' | wc -l | tr -d ' ')
if [[ "$STILL_ACTIVE" -eq 0 ]]; then
  pass "no original left in the active set (decay-weighted: cold tier owns them)"
else
  fail "$STILL_ACTIVE original(s) still in active set (should be archived only)"
fi
# audit-trail: originals must still EXIST on disk (archive preserves, never deletes).
TOTAL_ON_DISK=$(find "$MEM" -name 'feedback_topic_*.md' | wc -l | tr -d ' ')
if [[ "$TOTAL_ON_DISK" -ge 3 ]]; then
  pass "originals preserved on disk in _archive/ (audit trail intact, not deleted)"
else
  fail "originals were DELETED (only $TOTAL_ON_DISK remain on disk) — violates archive-not-delete"
fi

# (c) each new principle carries a memory_type label (CoALA taxonomy).
if grep -qi '^memory_type:' "$PRINCIPLE" 2>/dev/null; then
  pass "principle file carries a memory_type label"
else
  fail "principle file lacks a memory_type label (CoALA taxonomy requirement)"
fi

# (d) decay-weighted retrieval probe: a related cue returns the principle ranked
# ABOVE the archived originals. The active recall set (maxdepth 1) is consulted
# before the archive; the principle MUST be there and the originals MUST NOT.
CUE="bash timeout"
ACTIVE_HITS=$(find "$MEM" -maxdepth 1 -name 'feedback_*.md' -exec grep -li "$CUE" {} + 2>/dev/null | sort)
if printf '%s\n' "$ACTIVE_HITS" | grep -q "$PRINCIPLE"; then
  pass "retrieval probe: principle is in the active recall set (ranked above archive)"
else
  fail "retrieval probe: principle NOT in active set — principle must be retrievable by the cue"
fi
if printf '%s\n' "$ACTIVE_HITS" | grep -q 'feedback_topic_'; then
  fail "retrieval probe: an archived original leaked back into the active set (decay broken)"
else
  pass "retrieval probe: archived originals are NOT in the active set (decay-weighted)"
fi

# =====================================================================
log "=== AC-LSEL-016 edge case: below-threshold cohort → no-op (not a failure) ==="
MEM2="$(mktemp -d)"
# Only 2 topics, each importance 30 (sum 60 < 150) → below threshold.
for n in 1 2; do
cat > "$MEM2/feedback_low_${n}.md" <<MD
---
name: low-${n}
description: low-importance concrete observation ${n}
type: feedback
importance: 30
---

Low-importance topic ${n}.
MD
done
bash "$REFLECT" --memory-dir "$MEM2" --threshold 150 --min-topics 3 >/tmp/r2.out 2>&1
RC2=$?
if [[ $RC2 -eq 0 ]]; then
  pass "below-threshold cohort: reflect.sh exited 0 (no-op, not a failure)"
else
  fail "below-threshold cohort: reflect.sh exited non-zero ($RC2) — should be a clean no-op"
fi
SYNTH2=$(find "$MEM2" -maxdepth 1 -name 'feedback_*principle*.md' | wc -l | tr -d ' ')
if [[ "$SYNTH2" -eq 0 ]]; then
  pass "below-threshold cohort: no principle synthesized (correctly skipped)"
else
  fail "below-threshold cohort: $SYNTH2 principle(s) synthesized (should be 0)"
fi
LOW_STILL_ACTIVE=$(find "$MEM2" -maxdepth 1 -name 'feedback_low_*.md' | wc -l | tr -d ' ')
if [[ "$LOW_STILL_ACTIVE" -eq 2 ]]; then
  pass "below-threshold cohort: originals NOT archived (reflection did not fire)"
else
  fail "below-threshold cohort: originals touched despite no synthesis ($LOW_STILL_ACTIVE active)"
fi
rm -rf "$MEM2"

if [[ "$FAIL" -ne 0 ]]; then
  log "reflect_test: FAILED"
  exit 1
fi
log "reflect_test: PASS"
exit 0
