#!/bin/bash
# audit-spec-sync-drift.sh
#
# Interim audit script for spec.md ↔ progress.md status drift detection.
# Implements a bash subset of SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M2 deliverable
# (`moai spec audit`) per plan-phase v0.1.1.
#
# Era classification per SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 design.md §C.2:
#   - V3R6 modern (drift = action required)
#   - V3R5 (sync_section convention era — drift = SHOULD-FIX)
#   - V3R4 / V3R3 / V3R2 (early era — drift = INFO)
#   - V2.x / legacy (no SPEC- prefix or pre-V3R2 — era-final grandfather, no action)
#
# Exit code: always 0 (non-blocking; observability tool only).
# Output:    grouped report on stdout. Empty body if zero drift.
#
# Origin: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 plan-phase PASS 0.88 (commits
# `0616823dc` + `2c4513930` v0.1.1). Interim implementation pending Go-based
# `moai spec audit` CLI subcommand at run-phase M2.

set -uo pipefail

SPEC_ROOT=".moai/specs"

if [ ! -d "$SPEC_ROOT" ]; then
  echo "ERROR: $SPEC_ROOT directory not found (run from project root)" >&2
  exit 0   # non-blocking: tool is opt-in
fi

# Counters
total=0
drift_modern=0
drift_v3r5=0
drift_v3r4_or_older=0
era_final=0

# Output buffers (per-severity)
modern_lines=""
v3r5_lines=""
older_lines=""

for spec_dir in "$SPEC_ROOT"/SPEC-*/; do
  [ -d "$spec_dir" ] || continue
  spec_id=$(basename "$spec_dir")
  spec_file="$spec_dir/spec.md"
  prog_file="$spec_dir/progress.md"

  [ -f "$spec_file" ] || continue
  total=$((total + 1))

  # Read spec.md status (frontmatter only — first 30 lines)
  spec_status=$(awk '/^---$/{c++; if(c==2) exit} c==1 && /^status:/{print; exit}' "$spec_file" \
                 | sed 's/^status: *//; s/^"//; s/"$//')

  # Read progress.md status (frontmatter only — first 30 lines, if file exists)
  prog_status=""
  if [ -f "$prog_file" ]; then
    prog_status=$(awk '/^---$/{c++; if(c==2) exit} c==1 && /^status:/{print; exit}' "$prog_file" \
                   | sed 's/^status: *//; s/^"//; s/"$//')
  fi

  # Era classification
  case "$spec_id" in
    SPEC-V3R6-*) era="V3R6" ;;
    SPEC-V3R5-*) era="V3R5" ;;
    SPEC-V3R4-*) era="V3R4" ;;
    SPEC-V3R3-*) era="V3R3" ;;
    SPEC-V3R2-*) era="V3R2" ;;
    *)           era="legacy" ;;
  esac

  # Drift detection
  drift_kind=""
  if [ -z "$prog_status" ] && [ "$era" != "V3R6" ] && [ "$era" != "V3R5" ]; then
    # Pre-V3R6: progress.md absent is normal era-final pattern
    era_final=$((era_final + 1))
    continue
  fi

  if [ "$spec_status" = "$prog_status" ] || [ -z "$prog_status" ]; then
    # No drift OR no progress.md to compare (V2.x/V3R2-4 era-final)
    if [ "$era" = "legacy" ] || [ "$era" = "V3R2" ] || [ "$era" = "V3R3" ] || [ "$era" = "V3R4" ]; then
      era_final=$((era_final + 1))
    fi
    continue
  fi

  # Drift present: classify by era
  drift_line="  $spec_id [$era] spec.md=$spec_status vs progress.md=$prog_status"
  case "$era" in
    V3R6)
      drift_modern=$((drift_modern + 1))
      modern_lines="${modern_lines}${drift_line}"$'\n'
      ;;
    V3R5)
      drift_v3r5=$((drift_v3r5 + 1))
      v3r5_lines="${v3r5_lines}${drift_line}"$'\n'
      ;;
    *)
      drift_v3r4_or_older=$((drift_v3r4_or_older + 1))
      older_lines="${older_lines}${drift_line}"$'\n'
      ;;
  esac
done

# Render report
echo "=== SPEC Sync Drift Audit ==="
echo "Total SPECs scanned: $total"
echo "Era-final (grandfather, no action required): $era_final"
echo ""
echo "[MODERN-ERA drift (V3R6) — action target]: $drift_modern"
if [ -n "$modern_lines" ]; then
  echo -n "$modern_lines"
fi
echo ""
echo "[V3R5 drift — SHOULD-FIX]: $drift_v3r5"
if [ -n "$v3r5_lines" ]; then
  echo -n "$v3r5_lines"
fi
echo ""
echo "[Early-era drift (V3R2-R4) — INFO only, era-final grandfather]: $drift_v3r4_or_older"
if [ -n "$older_lines" ]; then
  echo -n "$older_lines"
fi
echo ""

# Summary
if [ "$drift_modern" -gt 0 ]; then
  echo "ACTION: $drift_modern modern-era drift(s) detected — recommended fix: orchestrator-direct atomic chore commit per SPEC (status backfill + version bump + HISTORY entry)."
  echo "        Reference: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 plan-phase v0.1.1 (commits 0616823dc + 2c4513930)."
elif [ "$drift_v3r5" -gt 0 ]; then
  echo "ACTION: $drift_v3r5 V3R5 drift(s) detected — SHOULD-FIX. Era classification grandfather applies if no audit-ready signal section in progress.md."
else
  echo "OK: no modern-era drift detected. $era_final SPECs era-final (grandfather)."
fi

# Non-blocking exit
exit 0
