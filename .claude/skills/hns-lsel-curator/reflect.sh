#!/usr/bin/env bash
# reflect.sh — SPEC-LSEL-LOCAL-EVOLUTION-001 M4 mechanical REFLECTION core.
#
# The periodic REFLECTION stage for the LSEL curator. Given a memory directory
# holding concrete `feedback_*.md` topic files, it fires when their accumulated
# importance clears the reflection threshold and synthesizes ONE abstract-
# principle `feedback_*.md`. The originals are MOVED to `memory/_archive/`
# (cold tier, NEVER deleted — archive preserves the audit trail per the design
# report "삭제 말고 보관" rule). The new principle carries a `memory_type` label
# (CoALA taxonomy: semantic → feedback).
#
# DESIGN (report §10 P4 + Vectorize 4 levers): reflection is THRESHOLD-FIRED,
# not wall-clock-fired (accumulated importance ~150, not a monthly cron). It is
# the decay/consolidation step the dominant failure mode ("un-refined
# accumulation → wrong-lesson retrieval") attacks.
#
# The synthesized principle is the MECHANICAL core (deterministic). The model-
# mediated layer (hns-lsel-curator SKILL) performs the real semantic synthesis;
# this script performs the mechanical relocation + principle-file scaffolding so
# decay-weighted retrieval is testable without an LLM in the loop.
#
# Exit codes: 0 = reflected OR clean no-op (below threshold); 1 = error.
set -euo pipefail

MEM=""
THRESHOLD="${REFLECT_THRESHOLD:-150}"
MIN_TOPICS="${REFLECT_MIN_TOPICS:-3}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --memory-dir) MEM="$2"; shift 2 ;;
    --threshold) THRESHOLD="$2"; shift 2 ;;
    --min-topics) MIN_TOPICS="$2"; shift 2 ;;
    *) echo "reflect: unknown arg: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$MEM" || ! -d "$MEM" ]]; then
  echo "reflect: --memory-dir required (must exist)" >&2
  exit 1
fi

# --- collect concrete topic files (active set, NOT _archive/) ---
# (portable: avoid mapfile/readarray, which are bash 4.0+ and absent on macOS
# system bash 3.2 that /usr/bin/env bash may resolve to.)
TOPICS=()
while IFS= read -r _f; do TOPICS+=("$_f"); done < <(find "$MEM" -maxdepth 1 -type f -name 'feedback_*.md' | sort)

# --- read importance from each topic's frontmatter (default 0 if absent) ---
fm_int() {
  awk -v k="importance" '
    /^---$/ { c++; next }
    c==1 && $0 ~ "^"k":" { sub("^"k":[[:space:]]*",""); gsub(/[^0-9].*/,""); print; exit }
  ' "$1"
}

TOTAL=0
COUNT=0
declare -a SOURCES=()
for f in "${TOPICS[@]}"; do
  imp="$(fm_int "$f")"; [[ -z "$imp" ]] && imp=0
  TOTAL=$((TOTAL + imp))
  COUNT=$((COUNT + 1))
  SOURCES+=("$f")
done

# --- threshold gate (accumulated importance, NOT wall clock) ---
if [[ "$COUNT" -lt "$MIN_TOPICS" ]] || [[ "$TOTAL" -lt "$THRESHOLD" ]]; then
  echo "reflect: no-op (count=$COUNT total_importance=$TOTAL < threshold=$THRESHOLD/min-topics=$MIN_TOPICS)"
  exit 0
fi

# --- synthesize ONE principle file (deterministic mechanical synthesis) ---
mkdir -p "$MEM/_archive"
TS="$(date -u +%FT%TZ)"
PRINCIPLE="$MEM/feedback_consolidated_principle_$(date +%s).md"

# Aggregate the concrete descriptions for the synthesis body.
DESCRIPTIONS=""
for f in "${SOURCES[@]}"; do
  desc="$(awk -v k="description" '
    /^---$/ { c++; next }
    c==1 && $0 ~ "^"k":" { sub("^"k":[[:space:]]*",""); print; exit }
  ' "$f")"
  [[ -z "$desc" ]] && desc="(no description)"
  DESCRIPTIONS+="  - $(basename "$f"): $desc"$'\n'
done

{
  printf -- '---\n'
  printf 'name: consolidated-principle\n'
  printf 'description: >\n'
  printf '  Consolidated principle synthesized from %s concrete topic files\n' "$COUNT"
  printf '  (accumulated importance %s >= reflection threshold %s). Originals archived\n' "$TOTAL" "$THRESHOLD"
  printf '  to memory/_archive/ — decay-weighted retrieval surfaces this principle first.\n'
  printf 'type: feedback\n'
  # CoALA taxonomy label — this principle is semantic knowledge (a feedback topic),
  # not a procedural skill (which would live in a hns-* skill body).
  printf 'memory_type: semantic\n'
  printf 'importance: %s\n' "$TOTAL"
  printf 'synthesized_at: %s\n' "$TS"
  printf 'source_count: %s\n' "$COUNT"
  printf -- '---\n\n'
  printf '# Consolidated principle\n\n'
  printf 'Synthesized from %s concrete topic files whose accumulated importance\n' "$COUNT"
  printf '(%s) cleared the reflection threshold (%s). The shared theme (drawn from\n' "$TOTAL" "$THRESHOLD"
  printf 'the source descriptions below) is the retrieval cue this principle answers:\n\n'
  printf '%s\n' "$DESCRIPTIONS"
  printf '\nOriginals moved to memory/_archive/ (cold tier, NOT deleted — archive\n'
  printf 'preserves the audit trail per the design report archive-not-delete rule).\n'
} > "$PRINCIPLE"

# --- relocate originals to _archive/ (decay-weighted: cold tier) ---
for f in "${SOURCES[@]}"; do
  mv "$f" "$MEM/_archive/$(basename "$f")"
done

echo "reflect: synthesized $PRINCIPLE (count=$COUNT total=$TOTAL threshold=$THRESHOLD); archived $COUNT originals"
exit 0
