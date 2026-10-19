#!/usr/bin/env bash
# codex-review-gate.sh - fail-closed reader of the native Codex review of a pull request.
#
# SPEC-CI-CODEX-REVIEW-001. The reviewer is the ChatGPT subscription's native Codex GitHub
# review; this script only reads what that review leaves on a pull request and refuses
# everything it cannot recognise. It holds no key and posts nothing. The native
# issue-comment summary and receipt are checked alongside formal reviews and findings.
#
# Subcommands:
#   read-native-review --reviews F --comments G [--issue-comments I] --head-sha S --draft true|false [--wait-closed true|false]
#   fetch-pages --endpoint E --out F [--expect-count N]
#   gate-files-changed --file FILES --head-sha S
#
# Exit classes (pairwise distinct; a caller treats any exit it does not recognise as a refusal):
#   0  pass      trusted current-head native artifacts are clean, no finding exists
#   1  fail      a named refusal
#   10 pending   no native review yet and the caller's wait window is open
#   11 NOT RUN   a draft pull request, or no native review when the wait window has ended;
#                neither a failure nor a pass, and the merge is withheld
# A non-zero exit prints one line to stderr beginning CODEX_GATE_REFUSED: followed by the reason.
#
# Portable to bash 3.2 (macOS): no associative arrays, no mapfile, no case-conversion expansions.

set -u

# Recognition contract (SPEC design 3.1): exactly one identity, matched on all three parts.
IDENT_LOGIN='chatgpt-codex-connector[bot]'
IDENT_ID=199175422
IDENT_TYPE='Bot'
SHA_RE='^[0-9a-fA-F]{40}$'

refuse() {
  printf 'CODEX_GATE_REFUSED: %s\n' "$2" >&2
  exit "$1"
}

# The decision, one jq program. Input lists are the NDJSON files of the pull request's reviews and
# review comments. Fields read: review id, user.login, user.id, user.type, commit_id, state,
# submitted_at; comment id, user.login, user.id, user.type, commit_id, original_commit_id.
# Output: one line, "<exit code> TAB <reason or summary>".
DECIDE_JQ='
def ident:
  (.user | type) == "object"
  and .user.login == $login and .user.id == $uid and .user.type == $utype;
def timestamp: type == "string" and test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([.][0-9]+)?Z$");
def idtext: if (.id | type) == "number" then (.id | tostring) else "?" end;
def is_head($v): ($v | type) == "string" and ($v | ascii_downcase) == $head;
def has_sha($v): ($v | type) == "string" and ($v | length) > 0;
# Native GitHub reviews publish a mutable full-SHA summary and a clean receipt.
# Neither receipt prose nor a short SHA is sufficient on its own.
($issues | map(select(ident))) as $native
| ($native | map(select((.body // "") | contains("<!-- codex-pull-request-review-summary -->"))) | sort_by(.updated_at, .id)) as $summaries
| ($native | map(select((.body // "") | startswith("<!-- codex-review-receipt:v1 -->"))) | sort_by(.created_at, .id)) as $receipts
| ($summaries[-1] // null) as $summary
| ($receipts[-1] // null) as $receipt
| (try ($summary.body | capture("<!-- codex-review-state:v1 (?<state>[^\\n]+) -->") | .state | fromjson | .code_review) catch null) as $state
| ($native | group_by(.id) | map(select(length > 1)) | length) as $native_dups
| ($native | map(select((.id | type) != "number" or (.id | floor) != .id or (.body | type) != "string" or (.created_at | timestamp | not) or (.updated_at | timestamp | not))) | length) as $native_bad
| ($reviews | map(select(ident))) as $cand
| ($cand | map(select(
      ((.id | type) != "number") or ((.id | floor) != .id)
      or ((.commit_id | type) != "string") or ((.state | type) != "string")))) as $bad
| ($cand | map(select((.id | type) == "number")) | group_by(.id) | map(select(length > 1)) | .[0]) as $dup
| ($cand | map(select(is_head(.commit_id)))) as $bound
| ($bound | map(select(.state == "CHANGES_REQUESTED") | .id) | max) as $cr
| ($bound | map(select(.state == "APPROVED") | .id) | max) as $ap
| ($comments | map(select(ident)
    | {bound: (is_head(.commit_id) or is_head(.original_commit_id)),
       unbound: ((has_sha(.commit_id) or has_sha(.original_commit_id)) | not),
       id: idtext})
    | map(select(.bound or .unbound)) | .[0]) as $find
| ($bound | max_by(.id)) as $n
| ($cand | max_by(.id)) as $newest
| if ($bad | length) > 0 then {c: 1, r: ("review-entry-invalid:" + ($bad[0] | idtext))}
  elif $dup != null then {c: 1, r: ("review-entry-invalid:duplicate-id:" + ($dup[0] | idtext))}
  elif $native_dups > 0 or $native_bad > 0 then {c: 1, r: "native-artifact-invalid"}
  elif ($summaries | length) > 0 or ($receipts | length) > 0 then
    if $find != null then {c: 1, r: ("finding:" + $find.id)}
    elif $cr != null and ($ap == null or $cr > $ap) then {c: 1, r: "changes-requested"}
    elif $summary == null or $receipt == null then {c: 1, r: "native-artifact-incomplete"}
    elif ($state | type) != "object" then {c: 1, r: "native-summary-invalid"}
    elif (is_head($state.commit_sha) | not) then {c: 1, r: "native-summary-stale"}
    elif $state.status != "completed" then {c: 1, r: "native-review-not-complete"}
    elif $state.has_any_findings != false or $state.has_unresolved_findings != false then {c: 1, r: "native-findings-or-unknown"}
    elif ($receipt.body | contains("Codex Review: Didn\u0027t find any major issues. :+1:")) | not then {c: 1, r: "native-receipt-not-clean"}
    elif ($receipt.body | contains("**Reviewed commit:** `" + $head[0:10] + "`")) | not then {c: 1, r: "native-receipt-stale"}
    else {c: 0, r: ("native-review-clean: issue-summary " + ($summary.id | tostring) + " receipt " + ($receipt.id | tostring) + " bound to " + $head + " artifacts " + ([$summary | {id,updated_at,body}, $receipt | {id,updated_at,body}] | tojson | @base64))}
    end
  elif ($cand | length) == 0 then
    (if $closed then {c: 11, r: "not-run:review-absent"} else {c: 10, r: "pending:no-review"} end)
  elif ($bound | length) == 0 then {c: 1, r: ("stale-review:" + ($newest.commit_id | .[0:10]))}
  elif $cr != null and ($ap == null or $cr > $ap) then {c: 1, r: "changes-requested"}
  elif $find != null then
    {c: 1, r: ((if $find.bound then "finding:" else "finding-unbound:" end) + $find.id)}
  elif $n.state == "COMMENTED" then {c: 1, r: "review-not-conclusive:COMMENTED"}
  elif $n.state != "APPROVED" then {c: 1, r: ("review-not-complete:" + $n.state)}
  elif (($n.submitted_at | type) != "string") or (($n.submitted_at | length) == 0) then
    {c: 1, r: "review-not-complete:no-submitted-at"}
  else {c: 0, r: ("native-review-clean: review " + ($n.id | tostring) + " state APPROVED bound to " + $head[0:10])}
  end
| "\(.c)\t\(.r)"
'

cmd_read_native_review() {
  local reviews="" comments="" issues="" head="" draft="" closed="false"
  while [ $# -gt 0 ]; do
    case "$1" in
      --issue-comments) issues="${2-}"; shift 2 || refuse 1 'usage:--issue-comments' ;;
      --reviews)     reviews="${2-}";  shift 2 || refuse 1 'usage:--reviews' ;;
      --comments)    comments="${2-}"; shift 2 || refuse 1 'usage:--comments' ;;
      --head-sha)    head="${2-}";     shift 2 || refuse 1 'usage:--head-sha' ;;
      --draft)       draft="${2-}";    shift 2 || refuse 1 'usage:--draft' ;;
      --wait-closed) closed="${2-}";   shift 2 || refuse 1 'usage:--wait-closed' ;;
      *) refuse 1 "usage:unknown-flag:$1" ;;
    esac
  done

  # Step 0: flags.
  case "$draft" in true | false) ;; *) refuse 1 'draft-flag-unusable' ;; esac
  case "$closed" in true | false) ;; *) refuse 1 'wait-closed-flag-unusable' ;; esac
  if ! [[ "$head" =~ $SHA_RE ]]; then refuse 1 'head-sha-unusable'; fi

  # Step 1: a draft is NOT RUN before any file is read; nothing here depends on whether the
  # native review runs on a draft.
  if [ "$draft" = "true" ]; then refuse 11 'not-run:draft'; fi

  # Step 2: both lists must exist and hold one JSON object per line.
  if [ -z "$reviews" ] || [ ! -f "$reviews" ] || [ -z "$comments" ] || [ ! -f "$comments" ]; then
    refuse 1 'input-invalid'
  fi
  local f
  for f in "$reviews" "$comments"; do
    if ! jq -e -s 'all(.[]; type == "object")' "$f" >/dev/null 2>&1; then refuse 1 'input-invalid'; fi
  done

  # Optional only for compatibility with callers reading formal reviews. The workflow
  # always supplies the complete, count-bound issue-comment list.
  if [ -n "$issues" ]; then
    [ -f "$issues" ] || refuse 1 'input-invalid'
    jq -e -s 'all(.[]; type == "object")' "$issues" >/dev/null 2>&1 || refuse 1 'input-invalid'
  else
    issues=/dev/null
  fi
  local lower_head result code reason
  lower_head="$(printf '%s' "$head" | tr 'A-F' 'a-f')"
  result="$(jq -n -r --slurpfile reviews "$reviews" --slurpfile comments "$comments" --slurpfile issues "$issues" \
    --arg head "$lower_head" --arg login "$IDENT_LOGIN" --argjson uid "$IDENT_ID" \
    --arg utype "$IDENT_TYPE" --argjson closed "$closed" "$DECIDE_JQ" 2>/dev/null)" \
    || refuse 1 'input-invalid'
  IFS=$'\t' read -r code reason <<EOF
$result
EOF
  case "$code" in
    0)
      printf '%s\n' "$reason"
      exit 0
      ;;
    1 | 10 | 11) refuse "$code" "$reason" ;;
    *) refuse 1 'decision-unreadable' ;;
  esac
}

# --- fetch-pages ----------------------------------------------------------------------------
# The only subcommand that reads from the API, and the only place pagination lives. An explicit
# loop over the documented per_page and page parameters is used instead of a piped --paginate: a
# piped call hides a mid-pagination failure, and a test cannot reproduce its behaviour.
ENDPOINT_RE='^repos/[A-Za-z0-9._-]+/[A-Za-z0-9._-]+/(pulls/[0-9]+/(files|reviews|comments)|issues/[0-9]+/comments)$'
PER_PAGE=100
MAX_PAGE=30   # 30 x 100 = 3000, the files listing's own limit; a full 30th page is never complete

FP_OUT=""
FP_TMP=""

# fetch_refuse removes every list this call produced, so a caller that ignores the exit status
# finds nothing to read.
fetch_refuse() {
  if [ -n "$FP_OUT" ]; then rm -f "$FP_OUT"; fi
  if [ -n "$FP_TMP" ]; then rm -f "$FP_TMP"; fi
  refuse 1 "$1"
}

cmd_fetch_pages() {
  local endpoint="" expect="" have_expect=0
  while [ $# -gt 0 ]; do
    case "$1" in
      --endpoint)     endpoint="${2-}"; shift 2 || refuse 1 'usage:--endpoint' ;;
      --out)          FP_OUT="${2-}";   shift 2 || refuse 1 'usage:--out' ;;
      --expect-count) expect="${2-}"; have_expect=1; shift 2 || refuse 1 'usage:--expect-count' ;;
      *) refuse 1 "usage:unknown-flag:$1" ;;
    esac
  done

  # Step 0: the endpoint allow-list, before any gh call, so no other listing can be read through here.
  if ! [[ "$endpoint" =~ $ENDPOINT_RE ]]; then refuse 1 "endpoint-not-allowed:${endpoint}"; fi
  if [ -z "$FP_OUT" ]; then refuse 1 'usage:--out'; fi
  FP_TMP="${FP_OUT}.tmp.$$"
  rm -f "$FP_OUT" "$FP_TMP"
  if [ "$have_expect" -eq 1 ] && ! [[ "$expect" =~ ^[0-9]+$ ]]; then fetch_refuse 'expect-count-unusable'; fi

  local kind page=1 total=0 n resp rc
  kind="${endpoint##*/}"
  : > "$FP_TMP"
  while [ "$page" -le "$MAX_PAGE" ]; do
    rc=0
    resp="$(gh api "${endpoint}?per_page=${PER_PAGE}&page=${page}" 2>/dev/null)" || rc=$?
    if [ "$rc" -ne 0 ]; then fetch_refuse "page-error:${page}"; fi
    # An array of objects, each carrying its key: the id of a review or review comment, the filename
    # of a changed file.
    if ! printf '%s' "$resp" | jq -e --arg kind "$kind" '
        type == "array" and all(.[];
          type == "object" and
          (if $kind == "files" then (.filename | type == "string" and length > 0)
           else ((.id | type) == "number") end))' >/dev/null 2>&1; then
      fetch_refuse "page-invalid:${page}"
    fi
    n="$(printf '%s' "$resp" | jq 'length')" || fetch_refuse "page-invalid:${page}"
    printf '%s' "$resp" | jq -c '.[]' >> "$FP_TMP" || fetch_refuse "page-invalid:${page}"
    total=$((total + n))
    # A page with fewer than 100 entries ends the collection; a page of exactly 100 asks for the
    # next one, so a list whose size is a multiple of 100 ends on an empty page.
    if [ "$n" -lt "$PER_PAGE" ]; then break; fi
    if [ "$page" -ge "$MAX_PAGE" ]; then fetch_refuse 'page-ceiling'; fi
    page=$((page + 1))
  done

  # Count proof, where the pull request object carries a count for this list.
  local proof='none'
  if [ "$have_expect" -eq 1 ]; then
    proof='yes'
    if [ "$total" -ne "$expect" ]; then fetch_refuse "count-mismatch:${total}:${expect}"; fi
  fi

  # Entry identity: a key that repeats is the signature of a list that shifted between two page
  # reads, which a count comparison alone cannot see.
  local dup
  dup="$(jq -r -s --arg kind "$kind" '
      map(if $kind == "files" then .filename else .id end) | group_by(.)
      | map(select(length > 1)) | .[0][0] // empty' "$FP_TMP" 2>/dev/null)" || fetch_refuse 'page-invalid:0'
  if [ -n "$dup" ]; then fetch_refuse "page-duplicate:${dup}"; fi

  mv "$FP_TMP" "$FP_OUT" || fetch_refuse 'out-unwritable'
  printf 'fetched:%s count-proof:%s\n' "$total" "$proof"
  exit 0
}

# --- gate-files-changed ---------------------------------------------------------------------
GATE_SCRIPT_PATH='scripts/ci/codex-review-gate.sh'
FILES_CEILING=3000

FILES_JQ='
def lc: ascii_downcase;
def is_gate($p): ($p | lc) as $l | ($l | startswith(".github/workflows/")) or ($l == $script);
def bound_ref:
  if (.contents_url | type) == "string"
  then (((.contents_url | capture("[?&]ref=(?<r>[^&#]*)")) // {r: ""}) | .r | lc)
  else "" end;
([.[] | (.filename, .previous_filename) | select(type == "string") | select(is_gate(.))] | .[0]) as $gp
| ([.[] | select((.status // "") != "removed") | select(bound_ref != $head) | .filename] | .[0]) as $ub
| if $gp != null then "1\tgate-path-changed:" + $gp
  elif $ub != null then "1\tfile-entry-unbound:" + $ub
  else "0\tfiles-clear" end
'

cmd_gate_files_changed() {
  local file="" head=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --file)     file="${2-}"; shift 2 || refuse 1 'usage:--file' ;;
      --head-sha) head="${2-}"; shift 2 || refuse 1 'usage:--head-sha' ;;
      *) refuse 1 "usage:unknown-flag:$1" ;;
    esac
  done

  if ! [[ "$head" =~ $SHA_RE ]]; then refuse 1 'head-sha-unusable'; fi
  if [ -z "$file" ] || [ ! -f "$file" ]; then refuse 1 'file-list-unusable'; fi
  local count
  count="$(jq -s 'length' "$file" 2>/dev/null)" || refuse 1 'file-entry-invalid'
  if [ "$count" -eq 0 ]; then refuse 1 'file-list-unusable'; fi
  if ! jq -e -s 'all(.[]; type == "object" and (.filename | type == "string" and length > 0))' "$file" >/dev/null 2>&1; then
    refuse 1 'file-entry-invalid'
  fi
  if [ "$count" -ge "$FILES_CEILING" ]; then refuse 1 'file-list-at-api-ceiling'; fi

  local lower_head result code reason
  lower_head="$(printf '%s' "$head" | tr 'A-F' 'a-f')"
  result="$(jq -s -r --arg head "$lower_head" --arg script "$GATE_SCRIPT_PATH" "$FILES_JQ" "$file" 2>/dev/null)" \
    || refuse 1 'file-entry-invalid'
  IFS=$'\t' read -r code reason <<EOF
$result
EOF
  if [ "$code" = "0" ]; then
    printf '%s: %s entries checked\n' "$reason" "$count"
    exit 0
  fi
  refuse 1 "$reason"
}

main() {
  if ! command -v jq >/dev/null 2>&1; then refuse 1 'jq-missing'; fi
  local sub="${1-}"
  if [ $# -gt 0 ]; then shift; fi
  case "$sub" in
    read-native-review) cmd_read_native_review "$@" ;;
    fetch-pages) cmd_fetch_pages "$@" ;;
    gate-files-changed) cmd_gate_files_changed "$@" ;;
    *) refuse 1 "unknown-subcommand:${sub}" ;;
  esac
}

main "$@"
