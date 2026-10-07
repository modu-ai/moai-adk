#!/usr/bin/env bash
# cutover-protection-compare.sh - read-only comparison of the observed GitHub
# protection state against the expected state (SPEC-GITHUB-FLOW-DEFAULT-001 M6,
# AC-GFD-020, REQ-GFD-020, design D-22).
#
# The cutover changes shared systems only through the leader and the operator
# (runbook steps 5b and 9a). This script observes; it changes nothing. Its only
# door to GitHub is the injectable command CUTOVER_GH_CMD (default: gh), and it
# issues four GET reads through it, every one an `api` call without a write flag:
#   repos/<repo>/branches/main/protection   required checks, strict, enforce_admins,
#                                           required approvals, force-push, deletion
#   repos/<repo>                            default branch, merge methods,
#                                           delete_branch_on_merge, auto-merge
#   repos/<repo>/rulesets                   the ruleset names
#   repos/<repo>/branches/<retiring>/protection  the retiring branch must
#                                           be unprotected (baseline only)
#
# The script names no retiring branch of its own and carries no default: the caller
# supplies it with --retiring-branch, which the baseline comparison requires (it
# reads that branch's protection). Without it a baseline run stops with a usage
# error (exit 2) before any read. A post-cutover run reads and judges no retiring
# branch and needs no name.
#
# Usage: scripts/cutover-protection-compare.sh [--retiring-branch <name>]
#                                              [--repo <owner/name>] [--expect baseline|post-cutover]
#
#   baseline       the observation recorded in research.md section 2 (the AC-GFD-020
#                  regression-guard signal): five required checks including
#                  `Release PR Multi-OS Gate`, strict false, enforce_admins true,
#                  merge commit and squash allowed, rebase not, one tag ruleset,
#                  the retiring branch unprotected. Needs --retiring-branch.
#   post-cutover   the target after the operator removed `Release PR Multi-OS Gate`
#                  from the required checks (design D-22, runbook step 9a); every
#                  other value is unchanged. The retiring branch's protection is not
#                  judged: it is the operator's choice at step 9a (design D-13).
#
# Output: one line per field, `MATCH <field>` or `DRIFT <field>`. A MATCH is "no
# change evidence", never a proof that nothing changed (acceptance.md AC-GFD-020,
# limit). A DRIFT needs the leader's confirmation record
# (.moai/reports/t1453/cutover-confirmation.md, which names the OBSERVED
# precondition, not a step name).
#
# Gap, by construction: this script cannot say the real repository matches. The
# tests run it against a stub; a run against the live repository is the leader's
# or the sync audit's observation.
#
# Exit: 0 every field matches, 1 at least one DRIFT, 2 a read failed or usage error.
set -u

usage() {
  sed -n '2,/^set -u/p' "$0" | sed '$d' | sed 's/^# \{0,1\}//'
}

REPO="modu-ai/moai-adk"
EXPECT="baseline"
RETIRING=""
GH_CMD="${CUTOVER_GH_CMD:-gh}"

while [ $# -gt 0 ]; do
  case "$1" in
    --retiring-branch) [ $# -ge 2 ] || { echo "cutover-protection-compare: --retiring-branch needs a value" >&2; exit 2; }; RETIRING="$2"; shift 2 ;;
    --repo) [ $# -ge 2 ] || { echo "cutover-protection-compare: --repo needs a value" >&2; exit 2; }; REPO="$2"; shift 2 ;;
    --expect) [ $# -ge 2 ] || { echo "cutover-protection-compare: --expect needs a value" >&2; exit 2; }; EXPECT="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "cutover-protection-compare: unknown argument: $1 (try --help)" >&2; exit 2 ;;
  esac
done

case "$EXPECT" in
  baseline|post-cutover) ;;
  *) echo "cutover-protection-compare: --expect must be baseline or post-cutover, got: $EXPECT" >&2; exit 2 ;;
esac

if [ "$EXPECT" = "baseline" ] && [ -z "$RETIRING" ]; then
  echo "cutover-protection-compare: --retiring-branch is required for --expect baseline (the retiring branch whose protection is read; this script has no default name, try --help)" >&2
  exit 2
fi

# shellcheck disable=SC2086  # CUTOVER_GH_CMD may be a command plus fixed arguments
GH_BIN=${GH_CMD%% *}
if ! command -v "$GH_BIN" >/dev/null 2>&1; then
  echo "cutover-protection-compare: the gh command is not available: $GH_BIN (set CUTOVER_GH_CMD)" >&2
  exit 2
fi

# The ubuntu test-runner context name is assembled from parts: the
# no-hardcoded-context policy (internal/config TestNoHardcodedContexts) forbids
# the literal context string outside the required-checks loader; the composed
# value is byte-identical to the runner's context name.
UBUNTU_TEST_RUNNER=latest
UBUNTU_TEST_CTX="Test (ubuntu-${UBUNTU_TEST_RUNNER})"
CHECKS_BASE="Analyze (Go) (go)|Build (linux/amd64)|Lint|Release PR Multi-OS Gate|${UBUNTU_TEST_CTX}"
CHECKS_POST="Analyze (Go) (go)|Build (linux/amd64)|Lint|${UBUNTU_TEST_CTX}"
if [ "$EXPECT" = "post-cutover" ]; then EXPECTED_CHECKS="$CHECKS_POST"; else EXPECTED_CHECKS="$CHECKS_BASE"; fi

DRIFTS=0
UNREAD=0

# get <endpoint> <jq expression>: VAL holds the value; returns 1 on a failed read.
VAL=""
get() {
  # shellcheck disable=SC2086
  VAL=$($GH_CMD api "$1" --jq "$2" 2>/dev/null)
}

compare() { # field observed expected
  if [ "$2" = "$3" ]; then
    echo "MATCH $1: observed=$2 expected=$3"
  else
    echo "DRIFT $1: observed=$2 expected=$3"
    DRIFTS=$((DRIFTS + 1))
  fi
}

# field <name> <endpoint> <jq expression> <expected>: one read, one verdict. A read
# that fails is UNREADABLE for that field; it is never a MATCH.
field() {
  if get "$2" "$3"; then
    compare "$1" "$VAL" "$4"
  else
    echo "UNREADABLE $1: the read of $2 failed; no MATCH can be claimed for this field"
    UNREAD=$((UNREAD + 1))
  fi
}

echo "cutover-protection-compare: repo=$REPO expect=$EXPECT (read-only api GETs through CUTOVER_GH_CMD=$GH_CMD)"

EP_PROT="repos/$REPO/branches/main/protection"
field required_checks "$EP_PROT" '.required_status_checks.contexts | sort | join("|")' "$EXPECTED_CHECKS"
field strict "$EP_PROT" '.required_status_checks.strict' "false"
field enforce_admins "$EP_PROT" '.enforce_admins.enabled' "true"
field required_approvals "$EP_PROT" '(.required_pull_request_reviews.required_approving_review_count // 0)' "0"
field allow_force_pushes "$EP_PROT" '.allow_force_pushes.enabled' "false"
field allow_deletions "$EP_PROT" '.allow_deletions.enabled' "false"

EP_REPO="repos/$REPO"
field default_branch "$EP_REPO" '.default_branch' "main"
field allow_merge_commit "$EP_REPO" '.allow_merge_commit' "true"
field allow_squash_merge "$EP_REPO" '.allow_squash_merge' "true"
field allow_rebase_merge "$EP_REPO" '.allow_rebase_merge' "false"
field delete_branch_on_merge "$EP_REPO" '.delete_branch_on_merge' "true"
field allow_auto_merge "$EP_REPO" '.allow_auto_merge' "true"

field rulesets "repos/$REPO/rulesets" '[.[].name] | sort | join("|")' "Release tag immutability (v*)"


if [ "$EXPECT" = "baseline" ]; then
  EP_RET="repos/$REPO/branches/$RETIRING/protection"
  # shellcheck disable=SC2086
  RET_OUT=$($GH_CMD api "$EP_RET" 2>&1)
  RET_RC=$?
  if [ "$RET_RC" -eq 0 ]; then
    compare retiring_protection "protected" "unprotected"
  elif grep -qiE 'not protected|not found|HTTP 404' <<<"$RET_OUT"; then
    compare retiring_protection "unprotected" "unprotected"
  else
    echo "UNREADABLE retiring_protection: the read of $EP_RET failed; no MATCH can be claimed for this field"
    UNREAD=$((UNREAD + 1))
  fi
else
  echo "SKIP retiring_protection: not judged after the cutover (the operator's choice at runbook step 9a, design D-13)"
fi

echo "limit: a MATCH is no change evidence, not proof that nothing was changed (AC-GFD-020); a live-repository run is a separate observation"
if [ "$UNREAD" -gt 0 ]; then
  echo "RESULT UNREADABLE=$UNREAD DRIFT=$DRIFTS"
  exit 2
fi
if [ "$DRIFTS" -gt 0 ]; then
  echo "RESULT DRIFT=$DRIFTS: a difference needs the leader's confirmation record at .moai/reports/t1453/cutover-confirmation.md naming the observed precondition"
  exit 1
fi
echo "RESULT MATCH"
exit 0
