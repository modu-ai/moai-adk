#!/bin/bash
# t383 M3 — dangling-link reconciliation (REQ-MSR-009, REQ-MSR-010).
#
# Copies each .md target named by the ACTIVE store's own MEMORY.md but absent from
# that store, sourcing it from the legacy store.
#
# [HARD] Copy-only and never-overwrite:
#   - nothing in the legacy store is deleted, moved, or modified (only read);
#   - a file already present in the active store is left untouched (`cp -n` semantics,
#     implemented explicitly with a `[ -e ]` test so the skip is logged rather than silent);
#   - the index (MEMORY.md) is not edited at all.
#
# Requires bash. Run from anywhere. Writes only into the ACTIVE store, and only files
# the index already names.

set -u
export LC_ALL=C

D="$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory"
L="$HOME/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory"
M="$D/MEMORY.md"

DRY="${DRY_RUN:-0}"

copied=0
skipped_exists=0
missing_source=0

# Derive the target list from the LIVE index at run time (plan.md §F M3 step 3).
grep -o '](\([^)]*\.md\))' "$M" | sed 's/](//;s/)$//' | sort -u > /tmp/t383-recon-targets.$$

while IFS= read -r f; do
  # Only ever consider targets the index names AND that are absent from the active store.
  if [ -e "$D/$f" ]; then
    continue
  fi
  if [ ! -f "$L/$f" ]; then
    echo "NO_SOURCE      $f"
    missing_source=$((missing_source + 1))
    continue
  fi
  # never-overwrite: re-test immediately before writing (the check above is not the guard)
  if [ -e "$D/$f" ]; then
    echo "SKIP_EXISTS    $f"
    skipped_exists=$((skipped_exists + 1))
    continue
  fi
  if [ "$DRY" = "1" ]; then
    echo "WOULD_COPY     $f"
  else
    cp -n "$L/$f" "$D/$f"
    echo "COPIED         $f"
  fi
  copied=$((copied + 1))
done < /tmp/t383-recon-targets.$$

echo "---"
echo "copied:         $copied"
echo "skipped_exists: $skipped_exists"
echo "no_source:      $missing_source"
echo "dry_run:        $DRY"

rm -f /tmp/t383-recon-targets.$$
