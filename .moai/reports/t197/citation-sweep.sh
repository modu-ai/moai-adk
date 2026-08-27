#!/bin/bash
# Citation GATE for t197.
#
#   bash .moai/reports/t197/citation-sweep.sh   # rc 0 = clean, rc 1 = drift
#
# This replaces a earlier script of the same name that LISTED citations and
# always exited 0. An audit found the obvious consequence: three citations whose
# line ranges were valid but whose content did not show the claim sat in front
# of it and it reported "sweep done". A check that cannot fail is a report.
#
# So this one judges. Every check below can exit non-zero, and the script's rc
# is the verdict. It reads only; it changes nothing.
#
# WHAT IT PROVES
#   C1  every live commit pin in prose equals the pin the transcript recorded
#   C2  every cited transcript range is in bounds
#   C3  every citation is bound to a manifest row (a declared claim)
#   C4  the cited range CONTAINS that row's evidence token
#   C5  no manifest row is stale (its citation still exists)
#
# WHAT IT DOES NOT PROVE
#   That the prose around a citation is a fair reading of the matched line.
#   C4 proves the evidence is present, not that the sentence describes it
#   correctly. Saying so here is deliberate: the script it replaced was
#   described in spec.md as confirming citation integrity, which was more than
#   it did.
#
# Run from the repository root.

set -uo pipefail

# Everything below is relative to this root. It defaults to the working
# directory (run from the repository root), and the self-test overrides it with
# a throwaway copy so that driving faults through the gate never touches a
# tracked file. Two concurrent self-test runs in the shared tree used to
# clobber each other's backups; a per-run root removes the shared state rather
# than sequencing access to it.
cd "${MOAI_CITATION_ROOT:-.}" || { echo "FATAL: cannot enter ${MOAI_CITATION_ROOT:-.}"; exit 2; }

TRANSCRIPT=.moai/reports/t197/probe-output.txt
MANIFEST=.moai/reports/t197/citation-manifest.txt
SPEC_L=.moai/specs/SPEC-CODEX-LAUNCHER-001
SPEC_I=.moai/specs/SPEC-CODEX-INIT-001
REPORTS=.moai/reports/t197

FAILURES=0
CHECKS=0

fail() { echo "FAIL: $*"; FAILURES=$((FAILURES + 1)); }
ok()   { echo "ok:   $*"; }

for f in "$TRANSCRIPT" "$MANIFEST"; do
  [ -r "$f" ] || { echo "FATAL: cannot read $f"; exit 2; }
done

TMP="$(mktemp -d)" || { echo "FATAL: mktemp -d failed"; exit 2; }
trap 'rm -rf "$TMP"' EXIT

NLINES=$(wc -l < "$TRANSCRIPT" | tr -d ' ')

# A citation is a LIVE claim about the current state. History rows and verdict
# files name old values on purpose; flagging those would train the reader to
# ignore the gate, so they are excluded by pattern.
NOT_LIVE='verdict-|낡은 핀|^[^:]*:[0-9]+:\| [0-9]+\.[0-9]+\.[0-9]+ \|'
SCOPE="$SPEC_L $SPEC_I $REPORTS"

# A line inside a fenced code block is QUOTED OUTPUT, not a live claim. The
# gate's own self-test record quotes the faults it injected — `deadbeef1`, an
# out-of-range citation — and re-judging those as live claims made the gate
# reject a correct tree. Fenced lines are therefore filtered out of both scans.
#
# The cost is real and is the reason this is written down: a citation a person
# writes inside a fence is invisible to the gate. Prose is where a claim
# belongs, so that trade is acceptable — but a fence is now a place where a
# citation stops being checked.
fenced_lines() { # <file> -> "<file>:<lineno>" for every line inside a fence
  awk -v f="$1" '
    /^[[:space:]]*(```|~~~)/ { infence = !infence; print f ":" NR; next }
    infence { print f ":" NR }
  ' "$1"
}

: > "$TMP/fenced"
for d in $SCOPE; do
  for f in $(find "$d" -name '*.md' -type f | sort); do
    fenced_lines "$f" >> "$TMP/fenced"
  done
done

# stdin: "file:lineno:text" -> drops the entries whose file:lineno is fenced.
# Matching is on the first two colon-separated fields only; a substring match
# would also drop line 200 when line 20 is fenced.
drop_fenced() {
  awk -F: -v setfile="$TMP/fenced" '
    BEGIN { while ((getline l < setfile) > 0) skip[l] = 1 }
    { key = $1 ":" $2 } !(key in skip)
  '
}

# ---------------------------------------------------------------- C1: pins
echo "=== C1  commit pins vs the transcript's own recorded HEAD ==="
# The transcript records its HEAD as the line following the rev-parse command.
PIN_RECORDED=$(grep -A1 '^\$ git rev-parse --short HEAD$' "$TRANSCRIPT" | sed -n '2p' | tr -d ' ')
if [ -z "$PIN_RECORDED" ]; then
  fail "the transcript does not record a HEAD pin — C1 cannot run"
else
  ok "transcript records HEAD = $PIN_RECORDED"
fi
CHECKS=$((CHECKS + 1))

# 7b217da7c is the t88 ancestor the probe checks, not a baseline pin.
grep -rnE '\b[0-9a-f]{9}\b' $SCOPE --include='*.md' \
  | grep -vE "$NOT_LIVE" \
  | grep -v '7b217da7c' \
  | drop_fenced > "$TMP/pinlines" || true

while IFS= read -r line; do
  [ -n "$line" ] || continue
  where=$(printf '%s' "$line" | cut -d: -f1,2)
  for pin in $(printf '%s' "$line" | grep -oE '\b[0-9a-f]{9}\b' | sort -u); do
    CHECKS=$((CHECKS + 1))
    if [ "$pin" = "$PIN_RECORDED" ]; then
      ok "$where cites $pin"
    else
      fail "$where cites $pin but the transcript recorded $PIN_RECORDED"
    fi
  done
done < "$TMP/pinlines"

# ------------------------------------------- C2-C4: transcript line citations
echo
echo "=== C2-C4  transcript line citations (bounds, binding, support) ==="
echo "transcript is $NLINES lines"

grep -rnE 'L[0-9]+-[0-9]+' $SCOPE --include='*.md' \
  | grep -vE "$NOT_LIVE" | drop_fenced > "$TMP/citelines" || true

: > "$TMP/seen"

while IFS= read -r entry; do
  [ -n "$entry" ] || continue
  file=$(printf '%s' "$entry" | cut -d: -f1)
  lno=$(printf '%s' "$entry" | cut -d: -f2)
  text=$(printf '%s' "$entry" | cut -d: -f3-)

  for range in $(printf '%s' "$text" | grep -oE 'L[0-9]+-[0-9]+' | sort -u); do
    a=${range#L}; a=${a%%-*}
    b=${range##*-}
    label="$file:$lno $range"
    printf '%s~%s\n' "$file" "$range" >> "$TMP/seen"

    # C2 bounds
    CHECKS=$((CHECKS + 1))
    if [ "$a" -lt 1 ] || [ "$b" -lt "$a" ] || [ "$b" -gt "$NLINES" ]; then
      fail "$label out of bounds (transcript has $NLINES lines)"
      continue
    fi

    # C3 binding: a manifest row for this (file, range) whose claim_regex
    # matches the citing line.
    # EVERY matching row is checked, not just the first. Stopping at the first
    # match let a second, stricter row sit in the manifest governing nothing —
    # the gate's own self-test caught that, which is the whole reason the
    # self-test exists.
    : > "$TMP/rows"
    while IFS= read -r mrow; do
      case "$mrow" in ''|'#'*) continue ;; esac
      mfile=$(printf '%s' "$mrow" | awk -F' ~ ' '{print $1}')
      mrange=$(printf '%s' "$mrow" | awk -F' ~ ' '{print $2}')
      mclaim=$(printf '%s' "$mrow" | awk -F' ~ ' '{print $3}')
      [ "$mfile" = "$file" ] || continue
      [ "$mrange" = "$range" ] || continue
      if printf '%s' "$text" | grep -qE -- "$mclaim"; then
        printf '%s\n' "$mrow" >> "$TMP/rows"
      fi
    done < "$MANIFEST"

    CHECKS=$((CHECKS + 1))
    if [ ! -s "$TMP/rows" ]; then
      fail "$label has no manifest row whose claim matches this line — declare it in $MANIFEST"
      continue
    fi
    ok "$label bound"

    # C4 support: the cited range must contain each bound row's evidence token.
    while IFS= read -r row; do
      support=$(printf '%s' "$row" | awk -F' ~ ' '{print $4}')
      CHECKS=$((CHECKS + 1))
      if sed -n "${a},${b}p" "$TRANSCRIPT" | grep -qE -- "$support"; then
        ok "$label supported by /$support/"
      else
        fail "$label does NOT contain /$support/ — the range is valid but does not show the claim"
      fi
    done < "$TMP/rows"
  done
done < "$TMP/citelines"

# ------------------------------------------------------- C5: stale manifest
echo
echo "=== C5  manifest rows that no longer have a citation ==="
sort -u "$TMP/seen" > "$TMP/seen.u" 2>/dev/null || : > "$TMP/seen.u"
while IFS= read -r mrow; do
  case "$mrow" in ''|'#'*) continue ;; esac
  mfile=$(printf '%s' "$mrow" | awk -F' ~ ' '{print $1}')
  mrange=$(printf '%s' "$mrow" | awk -F' ~ ' '{print $2}')
  CHECKS=$((CHECKS + 1))
  if grep -qxF "$mfile~$mrange" "$TMP/seen.u"; then
    ok "$mfile $mrange still cited"
  else
    fail "$mfile $mrange is in the manifest but nothing cites it any more (stale row)"
  fi
done < "$MANIFEST"

echo
if [ "$FAILURES" = 0 ]; then
  echo "CITATION GATE PASS — $CHECKS checks, 0 failures"
  exit 0
fi
echo "CITATION GATE FAIL — $CHECKS checks, $FAILURES failure(s)"
exit 1
