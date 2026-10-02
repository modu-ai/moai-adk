#!/usr/bin/env bash
# cutover-precheck.sh - batch-boundary precheck for the develop -> main base-branch
# cutover (SPEC-GITHUB-FLOW-DEFAULT-001 M6, AC-GFD-019, REQ-GFD-019, design D-8
# step 3 and D-25).
#
# READ-ONLY. It runs `git rev-parse|rev-list|merge-base|worktree list` and the
# injected reader commands below, nothing else: it does not fetch, push, tag,
# merge, or write any file. It reads origin/develop as the remote-tracking ref
# stands; fetch before running it if a fresher read is wanted (the output prints
# the SHA it compared against).
#
# The cutover must not start while any of these is true:
#   unpushed-develop-commits  local develop has commits origin/develop lacks
#   live-integration-window   the release-integration window has a live holder
#   live-slot-holder          a slot lease has a live holder
#   active-lane-session       a live session sits in a card worktree
#   unmerged-picked-card      a picked card's branch tip is not an ancestor of
#                             origin/develop (not merged, or merged but not pushed)
#   reader-unavailable        a reader failed or answered in a shape this script
#                             does not recognise: unobserved is a violation, never
#                             a pass
#
# Usage: scripts/cutover-precheck.sh [--repo <path>] [--local-ref <ref>]
#                                    [--remote-ref <ref>] [--exclude-card <id>]
#
#   --exclude-card <id>  exempts exactly ONE card (design D-25: the cutover card
#                        itself, whose post-merge commits may not be ancestors of
#                        origin/develop). It is accepted once; the exclusion is
#                        printed by name; it never masks the unpushed, window,
#                        slot or any other card's condition.
#
# Readers (each a command string run with `sh -c`; defaults in brackets):
#   CUTOVER_INTEGRATION_STATUS_CMD  [moai integration status]
#   CUTOVER_SLOT_STATUS_CMD         [moai slot status]
#   CUTOVER_SESSION_LIST_CMD        [moai session list --json]
#   CUTOVER_PID_ALIVE_CMD           [kill -0]   (called as: <cmd> <pid>)
#   CUTOVER_GTD_LIST_CMD            [moai gtd list]
#   CUTOVER_CARD_BRANCH_CMD         [worktree lookup by card id] (called as: <cmd> <card-id>)
#
# Exit: 0 clear (PRECHECK_CLEAR), 1 one or more violations, 2 usage or environment error.
set -u

usage() {
  sed -n '2,/^set -u/p' "$0" | sed '$d' | sed 's/^# \{0,1\}//'
}

REPO="."
LOCAL_REF="develop"
REMOTE_REF="origin/develop"
EXCLUDE=""

while [ $# -gt 0 ]; do
  case "$1" in
    --repo) [ $# -ge 2 ] || { echo "cutover-precheck: --repo needs a value" >&2; exit 2; }; REPO="$2"; shift 2 ;;
    --local-ref) [ $# -ge 2 ] || { echo "cutover-precheck: --local-ref needs a value" >&2; exit 2; }; LOCAL_REF="$2"; shift 2 ;;
    --remote-ref) [ $# -ge 2 ] || { echo "cutover-precheck: --remote-ref needs a value" >&2; exit 2; }; REMOTE_REF="$2"; shift 2 ;;
    --exclude-card)
      [ $# -ge 2 ] || { echo "cutover-precheck: --exclude-card needs a card id" >&2; exit 2; }
      if [ -n "$EXCLUDE" ]; then
        echo "cutover-precheck: --exclude-card is accepted once; the exemption covers exactly one card (design D-25)" >&2
        exit 2
      fi
      EXCLUDE="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "cutover-precheck: unknown argument: $1 (try --help)" >&2; exit 2 ;;
  esac
done

INTEGRATION_CMD="${CUTOVER_INTEGRATION_STATUS_CMD:-moai integration status}"
SLOT_CMD="${CUTOVER_SLOT_STATUS_CMD:-moai slot status}"
SESSION_CMD="${CUTOVER_SESSION_LIST_CMD:-moai session list --json}"
ALIVE_CMD="${CUTOVER_PID_ALIVE_CMD:-kill -0}"
GTD_CMD="${CUTOVER_GTD_LIST_CMD:-moai gtd list}"
BRANCH_CMD="${CUTOVER_CARD_BRANCH_CMD:-}"

git_r() { git -C "$REPO" "$@"; }

git_r rev-parse --git-dir >/dev/null 2>&1 || { echo "cutover-precheck: not a git repository: $REPO" >&2; exit 2; }
LOCAL_SHA=$(git_r rev-parse --verify --quiet "${LOCAL_REF}^{commit}") || { echo "cutover-precheck: cannot resolve $LOCAL_REF" >&2; exit 2; }
REMOTE_SHA=$(git_r rev-parse --verify --quiet "${REMOTE_REF}^{commit}") || { echo "cutover-precheck: cannot resolve $REMOTE_REF (run git fetch first)" >&2; exit 2; }

FAILS=0
ok() { printf 'ok   %s: %s\n' "$1" "$2"; }
note() { printf 'note %s\n' "$1"; }
fail() { printf 'FAIL %s: %s\n' "$1" "$2"; FAILS=$((FAILS + 1)); }

# run_reader <name> <command>: sets READER_OUT; returns non-zero (after printing a
# violation) when the reader failed. stderr is dropped from the parsed output.
READER_OUT=""
run_reader() {
  local name="$1" cmd="$2" rc
  READER_OUT=$(sh -c "$cmd" 2>/dev/null)
  rc=$?
  if [ "$rc" -ne 0 ]; then
    fail reader-unavailable "$name reader failed (exit $rc): $cmd"
    return 1
  fi
  return 0
}

echo "cutover-precheck: repo=$(git_r rev-parse --show-toplevel)"
echo "cutover-precheck: $LOCAL_REF=$LOCAL_SHA $REMOTE_REF=$REMOTE_SHA (the remote-tracking ref is read as it stands; this script does not fetch)"
if [ -n "$EXCLUDE" ]; then
  echo "exclusion: --exclude-card $EXCLUDE (exactly this card, design D-25; no other card, no other condition)"
else
  echo "exclusion: none"
fi

# 1. unpushed develop commits
UNPUSHED=$(git_r rev-list --count "$REMOTE_REF..$LOCAL_REF" 2>/dev/null) || UNPUSHED=""
case "$UNPUSHED" in
  ''|*[!0-9]*) fail reader-unavailable "git rev-list --count $REMOTE_REF..$LOCAL_REF did not return a count" ;;
  0) ok unpushed-develop-commits "0 commits in $REMOTE_REF..$LOCAL_REF" ;;
  *) fail unpushed-develop-commits "$UNPUSHED commit(s) in $REMOTE_REF..$LOCAL_REF are not pushed" ;;
esac

# 2. release-integration window
if run_reader integration "$INTEGRATION_CMD"; then
  if grep -q '^release-integration window: free' <<<"$READER_OUT"; then
    ok live-integration-window "window is free"
  elif grep -q '^release-integration window: held by a session that is gone' <<<"$READER_OUT"; then
    note "window: the recorded holder is gone (reclaimable); not a live holder"
  elif grep -q '^release-integration window: held$' <<<"$READER_OUT"; then
    holder=$(grep -m1 'holder:' <<<"$READER_OUT" | sed 's/^[[:space:]]*//')
    fail live-integration-window "the window has a live holder ${holder:+($holder)}"
  else
    fail reader-unavailable "integration status answered in an unrecognised shape: $(head -n1 <<<"$READER_OUT")"
  fi
fi

# 3. slot leases
if run_reader slot "$SLOT_CMD"; then
  live=0
  while IFS= read -r line; do
    case "$line" in
      "slot "*": held")
        name=${line#slot }; name=${name%: held}
        fail live-slot-holder "slot $name has a live holder"
        live=$((live + 1)) ;;
      "slot "*"(reclaimable)")
        note "slot: ${line#slot } (reclaimable; not a live holder)" ;;
    esac
  done <<<"$READER_OUT"
  [ "$live" -eq 0 ] && ok live-slot-holder "no live slot holder"
fi

# 4. active lane sessions: a live pid whose cwd is inside a card worktree
if run_reader sessions "$SESSION_CMD"; then
  expected=$(grep -o '"pid"' <<<"$READER_OUT" | wc -l | tr -d ' ')
  parsed=$(awk '
    /"pid":/ { v=$0; gsub(/[^0-9]/, "", v); pid=v }
    /"cwd":/ { v=$0; sub(/^[^:]*:[ \t]*"/, "", v); sub(/",?[ \t]*$/, "", v); cwd=v }
    /^[ \t]*}/ { if (pid != "" && cwd != "") print pid "\t" cwd; pid=""; cwd="" }
  ' <<<"$READER_OUT")
  got=0
  [ -n "$parsed" ] && got=$(wc -l <<<"$parsed" | tr -d ' ')
  if [ "$expected" != "$got" ]; then
    fail reader-unavailable "session list: $expected pid fields but $got parsed entries (expected one field per line, as moai prints it)"
  else
    lanes=0
    while IFS=$'\t' read -r pid cwd; do
      [ -n "$pid" ] || continue
      case "$cwd" in
        */.moai/worktrees/*|*/.claude/worktrees/*) ;;
        *) continue ;;
      esac
      sh -c "$ALIVE_CMD $pid" >/dev/null 2>&1 || continue
      rest=${cwd#*/worktrees/}; card=${rest%%/*}
      if [ -n "$EXCLUDE" ] && [ "$card" = "$EXCLUDE" ]; then
        echo "EXCLUDED card $card (--exclude-card): session pid $pid in $cwd is the cutover card's own lane (design D-25)"
        continue
      fi
      fail active-lane-session "$card (pid $pid, cwd $cwd)"
      lanes=$((lanes + 1))
    done <<<"$parsed"
    [ "$lanes" -eq 0 ] && ok active-lane-session "no live session inside a card worktree"
  fi
fi

# 5. picked cards: the tip must be an ancestor of the remote-tracking develop
default_card_branch() {
  git_r worktree list --porcelain 2>/dev/null | awk -v id="$1" '
    /^worktree / { path = substr($0, 10); n = split(path, a, "/"); base = a[n] }
    /^branch /   { if (base == id) { b = $2; sub("^refs/heads/", "", b); print b; exit } }'
}
card_branch() {
  if [ -n "$BRANCH_CMD" ]; then
    sh -c "$BRANCH_CMD \"\$1\"" cutover "$1" 2>/dev/null | head -n1
  else
    default_card_branch "$1"
  fi
}

if run_reader queue "$GTD_CMD"; then
  picked=$(awk -F'\t' '$2 == "picked" { print $1 }' <<<"$READER_OUT")
  open=0
  excluded=0
  for card in $picked; do
    br=$(card_branch "$card")
    tip=""
    [ -n "$br" ] && tip=$(git_r rev-parse --verify --quiet "${br}^{commit}")
    landed="no"
    [ -n "$tip" ] && git_r merge-base --is-ancestor "$tip" "$REMOTE_REF" 2>/dev/null && landed="yes"
    if [ -n "$EXCLUDE" ] && [ "$card" = "$EXCLUDE" ]; then
      echo "EXCLUDED card $card (--exclude-card): picked card, branch ${br:-unknown} tip ${tip:-unknown}, ancestor of $REMOTE_REF: $landed (design D-25)"
      excluded=$((excluded + 1))
      continue
    fi
    if [ -z "$br" ] || [ -z "$tip" ]; then
      fail unmerged-picked-card "$card has no resolvable branch (card-to-branch reader gave '${br}'); unobserved is not landed"
      open=$((open + 1))
    elif [ "$landed" = "no" ]; then
      fail unmerged-picked-card "$card tip ${tip} of $br is not an ancestor of $REMOTE_REF (not merged, or merged but not pushed)"
      open=$((open + 1))
    fi
  done
  [ "$open" -eq 0 ] && ok unmerged-picked-card "no non-excluded picked card is unmerged or unpushed ($(wc -w <<<"$picked" | tr -d ' ') picked read, $excluded excluded)"
fi

if [ "$FAILS" -eq 0 ]; then
  echo "PRECHECK_CLEAR"
  exit 0
fi
echo "PRECHECK_VIOLATIONS=$FAILS"
exit 1
