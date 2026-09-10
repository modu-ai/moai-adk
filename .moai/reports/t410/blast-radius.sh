#!/usr/bin/env bash
# t410 blast radius: of the SPEC-IDs the drift table currently flags DRIFT, how
# many are body-only close attributions — i.e. a lifecycle-close commit exists on
# the default branch whose BODY declares a close for the SPEC-ID while its
# SUBJECT does not name that id, so the walker's subject-scoped
# commitMatchesSPECID filter cannot see it.
#
# Two predicates, reported separately, because the loose one over-counts:
#
#   LOOSE  body merely CONTAINS the id. Over-counts: a `depends_on:
#          SPEC-X (completed)` line is a mention, not a close (measured
#          false hit: a83934d55 for SPEC-AUTONOMY-TIERS-001).
#   TIGHT  the body LINE naming the id also carries a close signal on that
#          same line (close / completed / Mx verdict). This is the number to
#          act on.
#
# Read-only. Takes the drift output on stdin.
set -uo pipefail

BRANCH="${1:-main}"
total=0
loose=0
tight=0
declare -a TIGHT_HITS
declare -a LOOSE_ONLY

while read -r id _rest; do
  case "$id" in SPEC-*) ;; *) continue ;; esac
  total=$(( total + 1 ))
  while read -r sha; do
    [ -n "$sha" ] || continue
    subject=$(git show -s --format='%s' "$sha")
    body=$(git show -s --format='%b' "$sha")
    # subject already carries the id -> the walker can see it; not our shape.
    case "$subject" in *"$id"*) continue ;; esac
    case "$body" in *"$id"*) ;; *) continue ;; esac
    case "$subject" in
      *"phase close"*|*"phase-close"*|*"Mx-phase"*|*"4-phase"*|*"3-phase"*|*"Close out"*) ;;
      *) continue ;;
    esac
    loose=$(( loose + 1 ))
    # TIGHT: the body line naming the id must itself declare a close.
    if printf '%s\n' "$body" | grep -F "$id" | grep -Eqi 'close|completed|Mx verdict'; then
      tight=$(( tight + 1 ))
      TIGHT_HITS+=("$id  <-  ${sha:0:9}  $subject")
    else
      LOOSE_ONLY+=("$id  <-  ${sha:0:9}  $subject")
    fi
    break
  done < <(git log "$BRANCH" --format='%H' --no-merges --grep="$id" -50)
done

echo "drift rows examined:        $total"
echo "LOOSE body-only close hits: $loose"
echo "TIGHT (line declares close): $tight"
echo "--- TIGHT hits (act on these) ---"
printf '%s\n' "${TIGHT_HITS[@]+"${TIGHT_HITS[@]}"}"
echo "--- LOOSE-only (mention, not a close — excluded) ---"
printf '%s\n' "${LOOSE_ONLY[@]+"${LOOSE_ONLY[@]}"}"
