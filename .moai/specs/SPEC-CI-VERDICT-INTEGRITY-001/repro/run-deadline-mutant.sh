#!/bin/sh
# t1534 repro E26 — deadline-absence probe, BEHAVIORAL form (AC-CI-005; leader-
# sanctioned post-ceiling repair P1-2). The audit's string-presence branch was
# fooled by two mutants: `&& 'deadline' == 'deadline'` (string with no time
# comparison -> false "repaired") and `&& false` (unmergeable guard -> false
# "proceeds"). This version EVALUATES the extracted condition as an executable
# guard under two clock inputs and requires BOTH observed decisions:
#
#   pre-deadline  (current_time=1000, deadline=1140)  -> MERGE_PROCEEDS
#   post-deadline (current_time=1200, deadline=1140)  -> MERGE_HELD
#
# Exit semantics:
#   0  RED   — pre-deadline proceeds AND post-deadline proceeds (deadline never
#             evaluated; the current tree's state)
#   2  FLIP  — pre-deadline proceeds AND post-deadline held (M1 contract holds)
#   3  BAD   — guard blocks a pre-deadline merge, or produces neither decision
#              (an always-withholding or untranslatable guard; never a pass)
#   9  ERROR — extraction or translation failure (loud, never silent)
#
# The extracted condition is mapped term-by-term: the three known output terms
# become shell tests; bare `deadline`/`current_time` tokens become references
# to the clock inputs so a repaired condition's own deadline term executes.
set -eu
WF=".github/workflows/auto-merge.yml"

COND="$(awk '
  index($0, "name: Auto merge PR") { grab = 1; next }
  grab && /^ *if: */ {
    line = $0
    sub(/^ *if: */, "", line)
    gsub(/^ +| +$/, "", line)
    if (line == "" || line ~ /^[>|][+-]?[0-9]*$/) { fold = 1; next }
    print line; exit
  }
  grab && fold {
    if (match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/) { exit }
    sub(/^ */, ""); print
  }
' "$WF")"

if [ -z "$(printf '%s' "$COND" | tr -d '[:space:]')" ]; then
  echo "repro extraction failed: merge-step if: condition not found in $WF — the workflow structure changed; update this probe's anchor." >&2
  exit 9
fi

echo "merge-step condition (extracted):"
printf '%s\n' "$COND"

# Map the known output terms to shell tests; expose the clock inputs to any
# deadline term the repair adds.
GUARD="$(printf '%s\n' "$COND" \
  | sed -e "s/steps\.checks\.outputs\.should_merge == 'true'/[ \"\$checks_should_merge\" = \"true\" ]/g" \
        -e "s/steps\.head\.outputs\.ok == 'true'/[ \"\$head_ok\" = \"true\" ]/g" \
        -e "s/steps\.coderabbit\.outputs\.ok == 'true'/[ \"\$coderabbit_ok\" = \"true\" ]/g" \
        -e 's/deadline/\$deadline/g' \
        -e 's/current_time/\$current_time/g')"

unmapped="$(printf '%s' "$GUARD" | grep -c 'steps\.' || true)"
[ -n "$unmapped" ] || unmapped=0
if [ "$unmapped" -gt 0 ]; then
  echo "repro translation incomplete: $unmapped unmapped steps.* term(s) — update this probe's term map." >&2
  exit 9
fi

eval_guard() {
  current_time="$1"
  deadline="$2"
  GUARD_SCRIPT="$(mktemp /tmp/t1534-e26-XXXXXX)"
  printf 'checks_should_merge=true\nhead_ok=true\ncoderabbit_ok=true\ncurrent_time=%s\ndeadline=%s\nif %s\nthen\necho MERGE_PROCEEDS\nelse\necho MERGE_HELD\nfi\n' \
    "$current_time" "$deadline" "$GUARD" > "$GUARD_SCRIPT"
  bash -e -o pipefail "$GUARD_SCRIPT"
}

pre_out="$(eval_guard 1000 1140)"
post_out="$(eval_guard 1200 1140)"
echo "pre-deadline  (current_time=1000): $pre_out"
echo "post-deadline (current_time=1200): $post_out"

if [ "$pre_out" = "MERGE_PROCEEDS" ] && [ "$post_out" = "MERGE_HELD" ]; then
  echo "M1 contract holds: the guard evaluates the clock and withholds post-deadline"
  exit 2
fi
if [ "$pre_out" = "MERGE_PROCEEDS" ] && [ "$post_out" = "MERGE_PROCEEDS" ]; then
  echo "DEFECT: merge proceeds at merge_at=1141 (declared_deadline=1140) — deadline never evaluated"
  exit 0
fi
echo "GUARD MISBEHAVES: pre-deadline merge must proceed before any post-deadline withholding can count"
exit 3
