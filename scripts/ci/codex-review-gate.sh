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
def full_sha: type == "string" and test("^[0-9A-Fa-f]{40}$");
def receipt_head_lines:
  (.body // "") | split("\n") | map(select(startswith("**Reviewed commit:** ")));
def receipt_head:
  receipt_head_lines as $lines
  | if ($lines | length) == 1 then ($lines[0] | ltrimstr("**Reviewed commit:** `") | rtrimstr("`")) else "" end;
def receipt_prefix_matches:
  . as $v
  | ($v | type) == "string" and ($v | test("^[0-9A-Fa-f]{10,40}$")) and ($head | startswith($v | ascii_downcase));
def receipt_kind:
  (.body // "") as $body
  | (($body | split("\n") | any(startswith("Codex Review:")))) as $code
  | (($body | contains("Codex Security Review")) or ($body | contains("Security review completed."))) as $security
  | if $code and $security then "ambiguous"
    elif $code then "code"
    elif $security then "security"
    else "unknown" end;
# Require the exact clean first sentence at line start; sign-off prose after that sentence is not part of the protocol.
# Identity, type, full-head state, both completed legs, and corroborating findings remain separate gates.
def code_clean_declaration:
  . == "Codex Review: Didn\u0027t find any major issues."
  or startswith("Codex Review: Didn\u0027t find any major issues. ");
def receipt_pending_for_running_leg($kind; $status; $error):
  $status == "running"
  and (
    $error == ("native-receipt-missing:" + $kind + "_review")
    or (($error | type) == "string" and ($error | startswith("stale-review:")))
  );
def receipt_check($kind):
  . as $all
  | ($all | map(select(.kind == $kind))) as $items
  | ($items | if length == 0 then null else max_by(.id) end) as $latest
  | ($items | map(select(.reviewed_head | receipt_prefix_matches)) | length) as $current_count
  | ($latest.body // "") as $body
  | ($body | split("\n")) as $lines
  | ([$lines[] | select(startswith("Codex Review:"))] | length) as $code_line_count
  | [$lines[] | select(startswith("**Reviewed commit:** "))] as $head_lines
  | ($latest | receipt_head) as $reviewed_head
  | (if $kind == "code"
     then [$lines[] | select(code_clean_declaration)] | length
     else [$lines[] | select(. == "Security review completed. No security issues were found in this pull request.")] | length
     end) as $clean_count
  | if $latest == null then "native-receipt-missing:\($kind)_review"
    elif $current_count > 1 then "native-receipt-ambiguous:\($kind)_review"
    elif $kind == "code" and $code_line_count != 1 then "native-receipt-ambiguous:\($kind)_review"
    elif ($body | [scan("<!-- codex-review-receipt:v1 -->")] | length) != 1 or $lines[0] != "<!-- codex-review-receipt:v1 -->" then "native-receipt-malformed:\($kind)_review"
    elif ($head_lines | length) != 1 or ($reviewed_head | test("^[0-9A-Fa-f]{10,40}$") | not) then "native-receipt-malformed:\($kind)_review"
    elif ($reviewed_head | receipt_prefix_matches | not) then ("stale-review:" + $reviewed_head[0:10])
    elif $clean_count != 1 then "native-receipt-not-clean:\($kind)_review"
    else null end;
def leading_block:
  split("\n")
  | reduce .[] as $line ([true, []];
      if .[0] and ($line | startswith("<!--") and endswith("-->")) then [true, .[1] + [$line]]
      elif .[0] then [false, .[1]]
      else . end)
  | .[1] | join("\n");
# Only trusted Codex issue comments are candidates; untrusted marker text is not parsed.
($issues | map(select(ident))) as $native
| ($native | map(select((.body // "") | contains("<!-- codex-review-state:v1")))) as $states
| ($native | map(select((.body // "") | contains("<!-- codex-review-receipt:v1 -->")))) as $receipts
| ($states[0] // null) as $state_comment
| ($state_comment.body // "") as $body
| ($body | leading_block) as $lead
| [$lead | split("\n")[] | select(startswith("<!-- codex-review-state:v1"))] as $state_lines
| [$lead | split("\n")[] | select(startswith("<!-- codex-security-review:v1"))] as $security_lines
| ($body | [scan("<!-- codex-review-state:v1")] | length) as $state_occurrences
| ($body | [scan("<!-- codex-security-review:v1")] | length) as $security_occurrences
| (try ($state_lines[0] | capture("^<!-- codex-review-state:v1 (?<json>\\{.*\\}) -->$").json | fromjson) catch null) as $state
| (try ($security_lines[0] | capture("^<!-- codex-security-review:v1 (?<json>\\{.*\\}) -->$").json | fromjson) catch null) as $security_marker
| ($state.code_review // null) as $code
| ($state.security_review // null) as $security
| ($states | map(select((.id | type) != "number" or (.id | floor) != .id or (.body | type) != "string" or (.created_at | timestamp | not) or (.updated_at | timestamp | not))) | length) as $state_bad
| ($states | map(.id) | group_by(.) | map(select(length > 1)) | length) as $state_id_dups
| ($receipts | map(select((.id | type) != "number" or (.id | floor) != .id or (.body | type) != "string" or (.created_at | timestamp | not) or (.updated_at | timestamp | not))) | length) as $receipt_bad
| ($receipts | map(.id) | group_by(.) | map(select(length > 1)) | length) as $receipt_id_dups
| ($receipts | map(. + {kind: receipt_kind, reviewed_head: receipt_head})) as $typed_receipts
| ($typed_receipts | map(select(.kind == "ambiguous")) | length) as $receipt_type_ambiguous
| ($typed_receipts | map(select(.kind == "unknown")) | length) as $receipt_type_unknown
| ($typed_receipts | receipt_check("code")) as $code_receipt_error
| ($typed_receipts | receipt_check("security")) as $security_receipt_error
| ($reviews | map(select(ident))) as $cand
| ($cand | map(select(((.id | type) != "number") or ((.id | floor) != .id) or ((.commit_id | type) != "string") or ((.state | type) != "string")))) as $bad
| ($cand | map(select((.id | type) == "number")) | group_by(.id) | map(select(length > 1)) | .[0]) as $dup
| ($cand | map(select(is_head(.commit_id)))) as $bound
| ($bound | map(select(.state == "CHANGES_REQUESTED") | .id) | max) as $cr
| ($bound | map(select(.state == "COMMENTED") | .id) | max) as $commented
| ($bound | max_by(.id)) as $n
| ($cand | map(select((.id | type) == "number")) | max_by(.id)) as $newest
| ($comments | map(select(ident)
    | {bound: (is_head(.commit_id) or is_head(.original_commit_id)),
       unbound: (((.commit_id | type) != "string" or .commit_id == "") and ((.original_commit_id | type) != "string" or .original_commit_id == "")),
       id: idtext})
    | map(select(.bound or .unbound)) | .[0]) as $find
| if ($bad | length) > 0 then {c: 1, r: ("review-entry-invalid:" + ($bad[0] | idtext))}
  elif $dup != null then {c: 1, r: ("review-entry-invalid:duplicate-id:" + ($dup[0] | idtext))}
  elif $state_id_dups > 0 then {c: 1, r: "native-artifact-invalid"}
  elif $receipt_bad > 0 or $receipt_id_dups > 0 then {c: 1, r: "native-artifact-invalid"}
  elif ($states | length) > 1 then {c: 1, r: ("state-comment-duplicated:" + ($states | map(.id | tostring) | join(",")))}
  elif $state_comment != null and $state_bad > 0 then {c: 1, r: "state-comment-invalid"}
  elif $state_comment != null then
    if $state_occurrences > 1 then {c: 1, r: "state-marker-duplicated"}
    elif ($state_lines | length) != 1 then {c: 1, r: "state-marker-misplaced"}
    elif ($state | type) != "object" or ($code | type) != "object" or ($security | type) != "object" then {c: 1, r: "state-marker-malformed"}
    elif ($code.status | type) != "string" or ($security.status | type) != "string" then {c: 1, r: "state-marker-malformed"}
    elif ($code.commit_sha | type) == "string" and ($code.commit_sha | full_sha) and (is_head($code.commit_sha) | not) then {c: 1, r: ("stale-review:" + $code.commit_sha[0:10])}
    elif ($security.commit_sha | type) == "string" and ($security.commit_sha | full_sha) and (is_head($security.commit_sha) | not) then {c: 1, r: ("stale-review:" + $security.commit_sha[0:10])}
    elif ($security_occurrences > 1) then {c: 1, r: "security-marker-duplicated"}
    elif $security_occurrences == 1 and (($security_lines | length) != 1) then {c: 1, r: "security-marker-misplaced"}
    elif $security_occurrences == 1 and ($security_marker | type) != "object" then {c: 1, r: "security-marker-malformed"}
    elif $security_occurrences == 1 and (($security_marker.status | type) != "string" or ($security_marker.headSha | full_sha | not)) then {c: 1, r: "security-marker-malformed"}
    elif $security_occurrences == 1 and ($security_marker.headSha | full_sha) and (is_head($security_marker.headSha) | not) then {c: 1, r: ("stale-review:" + $security_marker.headSha[0:10])}
    elif $receipt_type_ambiguous > 0 then {c: 1, r: "native-receipt-type-ambiguous"}
    elif $receipt_type_unknown > 0 then {c: 1, r: "native-receipt-type-unknown"}
    elif $code.status == "failed" then {c: 1, r: "review-failed:code_review"}
    elif $security.status == "failed" then {c: 1, r: "review-failed:security_review"}
    elif $code.status != "completed" and $code.status != "running" then {c: 1, r: ("review-not-complete:code_review:" + $code.status)}
    elif $security.status != "completed" and $security.status != "running" then {c: 1, r: ("review-not-complete:security_review:" + $security.status)}
    elif $find != null then {c: 1, r: ("finding:" + $find.id)}
    elif $cr != null then {c: 1, r: "changes-requested"}
    elif $commented != null then {c: 1, r: "review-not-conclusive:COMMENTED"}
    elif $code_receipt_error != null and (receipt_pending_for_running_leg("code"; $code.status; $code_receipt_error) | not) then {c: 1, r: $code_receipt_error}
    elif $security_receipt_error != null and (receipt_pending_for_running_leg("security"; $security.status; $security_receipt_error) | not) then {c: 1, r: $security_receipt_error}
    elif $code.status == "running" and ($code.commit_sha | full_sha | not) then {c: 1, r: "review-not-complete:code_review:commit_sha"}
    elif $security.status == "running" and ($security.commit_sha | full_sha | not) then {c: 1, r: "review-not-complete:security_review:commit_sha"}
    elif ($code | has("has_any_findings") | not) then {c: 1, r: "review-not-complete:code_review:has_any_findings"}
    elif ($security | has("has_any_findings") | not) then {c: 1, r: "review-not-complete:security_review:has_any_findings"}
    elif ($code.has_any_findings != null and ($code.has_any_findings | type) != "boolean") then {c: 1, r: "review-not-complete:code_review:has_any_findings"}
    elif ($security.has_any_findings != null and ($security.has_any_findings | type) != "boolean") then {c: 1, r: "review-not-complete:security_review:has_any_findings"}
    elif ($code | has("has_unresolved_findings") | not) then {c: 1, r: "review-not-complete:code_review:has_unresolved_findings"}
    elif ($security | has("has_unresolved_findings") | not) then {c: 1, r: "review-not-complete:security_review:has_unresolved_findings"}
    elif ($code.has_unresolved_findings != null and ($code.has_unresolved_findings | type) != "boolean") then {c: 1, r: "review-not-complete:code_review:has_unresolved_findings"}
    elif ($security.has_unresolved_findings != null and ($security.has_unresolved_findings | type) != "boolean") then {c: 1, r: "review-not-complete:security_review:has_unresolved_findings"}
    elif $code.status != "running" and ($code.has_any_findings | type) != "boolean" then {c: 1, r: "review-not-complete:code_review:has_any_findings"}
    elif $security.status != "running" and ($security.has_any_findings | type) != "boolean" then {c: 1, r: "review-not-complete:security_review:has_any_findings"}
    elif $code.status != "running" and ($code.has_unresolved_findings | type) != "boolean" then {c: 1, r: "review-not-complete:code_review:has_unresolved_findings"}
    elif $security.status != "running" and ($security.has_unresolved_findings | type) != "boolean" then {c: 1, r: "review-not-complete:security_review:has_unresolved_findings"}
    elif $code.has_unresolved_findings == true then {c: 1, r: "finding:unresolved:code_review"}
    elif $security.has_unresolved_findings == true then {c: 1, r: "finding:unresolved:security_review"}
    elif $code.status == "completed" and ($code.commit_sha | full_sha | not) then {c: 1, r: "review-not-complete:code_review:commit_sha"}
    elif $security.status == "completed" and ($security.commit_sha | full_sha | not) then {c: 1, r: "review-not-complete:security_review:commit_sha"}
    elif $security_occurrences == 0 and ($code.status != "running" and $security.status != "running") then {c: 1, r: "security-marker-missing"}
    elif $security_occurrences == 1 and $security_marker.status == "running" then
      (if $closed then {c: 1, r: "review-not-complete:running"} else {c: 10, r: "pending:running"} end)
    elif $security_occurrences == 1 and ($code.status != "running" and $security.status != "running") and $security_marker.status != "completed" then {c: 1, r: ("security-marker-disagrees:" + ($security_marker.status // "null"))}
    elif $security_occurrences == 1 and ($code.status == "running" or $security.status == "running") and $security_marker.status != "completed" and $security_marker.status != "running" then {c: 1, r: ("security-marker-disagrees:" + ($security_marker.status // "null"))}
    elif ($code.status == "running" or $security.status == "running") then
      (if $closed then {c: 1, r: "review-not-complete:running"} else {c: 10, r: "pending:running"} end)
    elif ($code.status != "completed") then {c: 1, r: ("review-not-complete:code_review:" + $code.status)}
    elif ($security.status != "completed") then {c: 1, r: ("review-not-complete:security_review:" + $security.status)}
    elif (is_head($code.commit_sha) | not) or (is_head($security.commit_sha) | not) then {c: 1, r: "native-summary-stale"}
    elif $security_occurrences == 0 and ($security_lines | length) == 0 then {c: 1, r: "security-marker-missing"}
    elif $security_marker.status != "completed" then {c: 1, r: ("security-marker-disagrees:" + ($security_marker.status // "null"))}
    elif ($security_marker.headSha | full_sha | not) then {c: 1, r: "security-marker-malformed"}
    elif (is_head($security_marker.headSha) | not) then {c: 1, r: ("stale-review:" + $security_marker.headSha[0:10])}
    else
      {c: 0, r: ("native-review-clean: state-comment " + ($state_comment.id | tostring) + " updated " + $state_comment.updated_at + " markers " + ([{code_review: {status: $code.status, commit_sha: $code.commit_sha, has_unresolved_findings: $code.has_unresolved_findings}, security_review: {status: $security.status, commit_sha: $security.commit_sha, has_unresolved_findings: $security.has_unresolved_findings}, security_marker: {status: $security_marker.status, headSha: $security_marker.headSha}}] | tojson | @base64))}
    end
  elif $find != null then
    {c: 1, r: ((if $find.bound then "finding:" else "finding-unbound:" end) + $find.id)}
  elif $cr != null then {c: 1, r: "changes-requested"}
  elif $commented != null then {c: 1, r: "review-not-conclusive:COMMENTED"}
  elif ($cand | length) == 0 then
    (if $closed then {c: 11, r: "not-run:review-absent"} else {c: 10, r: "pending:no-review"} end)
  elif ($bound | length) == 0 then {c: 1, r: ("stale-review:" + ($newest.commit_id | .[0:10]))}
  elif $n.state != "APPROVED" then {c: 1, r: ("review-not-complete:" + $n.state)}
  elif (($n.submitted_at | type) != "string") or (($n.submitted_at | length) == 0) then {c: 1, r: "review-not-complete:no-submitted-at"}
  elif $closed then {c: 11, r: "not-run:review-absent"}
  else {c: 10, r: "pending:no-state-comment"}
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
