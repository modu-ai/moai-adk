#!/usr/bin/env bash
# AC-GCB-001 / AC-GCB-002 / AC-GCB-009(parity) body check (read-only).
# Usage (from repo root):  bash check-gtd-body.sh <moai-binary> [gtd.md path]
# Prints one "CHECK <id> PASS|FAIL <detail>" line per assertion.
# Exit: 0 all PASS, 1 any FAIL, 2 usage/tool error.
# Must run under bash: zsh does not word-split the flag loops and passes vacuously.
[ -n "${BASH_VERSION:-}" ] || { echo "USAGE_ERROR: run with bash (bash check-gtd-body.sh ...)"; exit 2; }
set -u

BIN="${1:-}"
DOC="${2:-.claude/skills/moai/workflows/gtd.md}"
if [ -z "$BIN" ] || [ ! -x "$BIN" ]; then echo "USAGE_ERROR: executable moai binary required"; exit 2; fi
if [ ! -r "$DOC" ]; then echo "USAGE_ERROR: cannot read $DOC"; exit 2; fi

fails=0
check() { # id, condition-exit-status, detail
  if [ "$2" -eq 0 ]; then echo "CHECK $1 PASS $3"; else echo "CHECK $1 FAIL $3"; fails=$((fails+1)); fi
}

# --- AC-GCB-001: sections carried + body-size floor ---------------------------
# Floor = 279 lines (workflows/todo.md at base f67d2193f) - 12 lines (gtd.md stub at base) = 267.
FLOOR=267
for h in 'What It Is' 'Commands' 'Reading the records' 'Picking the next card' \
         'Standing sources' 'Outside Kanban Mode' 'Boundaries' 'Cross-references' 'GTD stages'; do
  n=$(grep -cxF "## $h" "$DOC")
  [ "$n" -eq 1 ]; check "001-h2:$h" $? "count=$n want=1"
done
n=$(grep -cxF '### What the analyser may do' "$DOC")
[ "$n" -eq 1 ]; check "001-h3:analyser" $? "count=$n want=1"
lines=$(wc -l < "$DOC" | tr -d ' ')
[ "$lines" -ge "$FLOOR" ]; check "001-floor" $? "lines=$lines floor=$FLOOR"

# --- AC-GCB-002: GTD stages section --------------------------------------------
SECTION=$(awk '/^## GTD stages$/{f=1;next} /^## /{f=0} f' "$DOC")
[ -n "$SECTION" ]; check "002-section-present" $? "bytes=${#SECTION}"

# usage shapes, one per verb, copied from `moai gtd <verb> --help` USAGE lines
for u in 'moai gtd capture <text>' 'moai gtd clarify <gtd-id>' 'moai gtd organize <gtd-id>' \
         'moai gtd reflect' 'moai gtd engage <gtd-id>' 'moai gtd answer <t-id> <text>'; do
  n=$(printf '%s\n' "$SECTION" | grep -cF "$u")
  [ "$n" -ge 1 ]; check "002-usage:$u" $? "count=$n want>=1"
  verb=$(printf '%s' "$u" | awk '{print $3}')
  "$BIN" gtd "$verb" --help 2>/dev/null | grep -qF "$u"; check "002-usage-in-help:$verb" $? "help USAGE carries '$u'"
done

# required flags: every flag of each verb's --help except --json and --help
REQUIRED='capture:--event,--sensitivity,--source,--source-authorized
clarify:--authority,--disposition,--evidence,--outcome,--trusted
organize:--class,--context,--depends-on,--part-of,--review-at
reflect:--rebuild-projection
engage:--approve,--dependencies-ready,--dispatch,--fresh,--lane,--pick,--resources,--run-id'
req_total=0
while IFS=: read -r verb flags; do
  help=$("$BIN" gtd "$verb" --help 2>/dev/null)
  # set equality, both directions: frozen list == flags parsed from help (minus --json/--help)
  want=$(printf '%s' "$flags" | tr ',' '\n' | sort -u)
  got=$(printf '%s\n' "$help" | grep -oE '^ +--[a-z][a-z-]*' | tr -d ' ' | grep -vxE -- '--json|--help' | sort -u)
  [ -n "$got" ] && [ "$want" = "$got" ]; check "002-flag-set-equal:$verb" $? "frozen=[$(printf '%s' "$want" | tr '\n' ' ')] help=[$(printf '%s' "$got" | tr '\n' ' ')]"
  for f in $(printf '%s' "$flags" | tr ',' ' '); do
    req_total=$((req_total+1))
    printf '%s\n' "$help" | grep -qE -- "(^|[^a-z-])$f([^a-z-]|$)"; check "002-flag-in-help:$verb$f" $? "help lists $f"
    printf '%s\n' "$SECTION" | grep -qE -- "(^|[^a-z-])$f([^a-z-]|$)"; check "002-flag-documented:$verb$f" $? "section names $f"
  done
done <<EOF
$REQUIRED
EOF
[ "$req_total" -eq 23 ]; check "002-required-count" $? "required=$req_total want=23 (non-vacuous)"

# every flag token the section names must exist in the union of the six verbs' help
UNION=$(for v in capture clarify organize reflect engage answer; do "$BIN" gtd "$v" --help 2>/dev/null; done)
doc_flags=$(printf '%s\n' "$SECTION" | grep -oE -- '--[a-z][a-z-]*' | sort -u)
nflags=$(printf '%s\n' "$doc_flags" | grep -c .)
[ "$nflags" -ge 23 ]; check "002-doc-flag-count" $? "distinct=$nflags want>=23"
for f in $doc_flags; do
  printf '%s\n' "$UNION" | grep -qE -- "(^|[^a-z-])$f([^a-z-]|$)"; check "002-no-invented-flag:$f" $? "present in help union"
done

# queue boundary phrases, verbatim from `moai gtd --help`
for p in 'stay separate from the established development queue' 'explicitly approved Engage'; do
  printf '%s\n' "$SECTION" | grep -qF "$p"; check "002-phrase:$p" $? "section carries phrase"
  "$BIN" gtd --help 2>/dev/null | tr -s ' \n' ' ' | grep -qF "$p"; check "002-phrase-in-help:$p" $? "help carries phrase"
done

# answer is the gate-response verb, not a sixth stage
printf '%s\n' "$SECTION" | grep -qF 'gate-blocked'; check "002-answer-gate" $? "section describes answer via 'gate-blocked'"
n=$(printf '%s\n' "$SECTION" | grep -ciE '(six|sixth|6|6th) (gtd )?stages?|answer[^.]*(is|as) (a|an|the) [^.]*stage([^a-z]|$)')
[ "$n" -eq 0 ]; check "002-answer-not-stage" $? "six/sixth-stage mentions=$n want=0"

# --- AC-GCB-009 parity: local/mirror pairs byte-identical at base stay identical -
for p in .claude/skills/moai/workflows/gtd.md \
         .claude/rules/moai/workflow/kanban-dispatch.md \
         .claude/rules/moai/workflow/kanban-dispatch-detail.md \
         .claude/skills/moai-kanban-foreman/SKILL.md \
         .claude/skills/moai/workflows/project/doc-generation.md \
         .moai/docs/todo-queue-storage.md; do
  cmp -s "$p" "internal/template/templates/$p"; check "009-cmp:$p" $? "local vs mirror"
done

echo "fails=$fails"
[ "$fails" -eq 0 ]
