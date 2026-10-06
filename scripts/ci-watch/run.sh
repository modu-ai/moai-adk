#!/bin/sh
# scripts/ci-watch/run.sh — CI watch loop for SPEC-V3R3-CI-AUTONOMY-001 T2.
# Polls gh pr checks every 30 seconds, classifies required vs auxiliary,
# emits structured JSON handoff on required failure, exits 0 on all-pass.
#
# Usage: MOAI_CIWATCH_GH=gh sh run.sh <PR_NUMBER> [BRANCH]
# Environment overrides (for testing):
#   MOAI_CIWATCH_GH                 — gh binary path (default: gh)
#   MOAI_CIWATCH_REQUIRED_CHECKS_FILE — path to required-checks.yml SSoT
#   MOAI_CIWATCH_NO_SLEEP           — skip sleep between polls (testing)
#   CIWATCH_TIMEOUT_SECONDS         — wall-clock timeout in seconds (default: 1800)
#
# Exit codes:
#   0 — all required checks passed
#   1 — error (PR not found, gh auth, SSoT missing)
#   2 — required check(s) failed (JSON handoff written to stdout)
#   3 — hard timeout reached (30 min)
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)"
# shellcheck source=lib/_common.sh
. "$SCRIPT_DIR/lib/_common.sh"
# shellcheck source=lib/timeout.sh
. "$SCRIPT_DIR/lib/timeout.sh"
# shellcheck source=lib/classify.sh
. "$SCRIPT_DIR/lib/classify.sh"

# ─── argument parsing ─────────────────────────────────────────────────────────

if [ $# -lt 1 ]; then
    abort "Usage: $0 <PR_NUMBER> [BRANCH]" 1
fi

PR_NUMBER="$1"
BRANCH="${2:-main}"

GH="${MOAI_CIWATCH_GH:-gh}"
POLL_INTERVAL="${CIWATCH_POLL_INTERVAL:-30}"

# t1534 gate round: the SSoT keys are BASE-branch patterns (main,
# release/*) — the caller's $2 is the HEAD branch. Resolve the PR's actual
# base branch from gh for the SSoT lookup; the handoff keeps reporting the
# HEAD branch. (GH/POLL must be initialized BEFORE this block — the gate
# caught $GH used-when-unset here.)
# GATE-5 P1: a failing `pr view` must NOT fall back silently to the head
# branch — head branches (feature/*, WT-*) are not SSoT keys, the required
# list comes back EMPTY, and an empty required list scores "all required
# checks passed" (exit 0) with every failure reclassified advisory (gate
# repro: injected Lint=fail surfaced exit 0). Retry once, then abort.
PR_BASE="$("$GH" pr view "$PR_NUMBER" --json baseRefName --jq '.baseRefName' 2>/dev/null)" || PR_BASE=""
if [ -z "$PR_BASE" ]; then
    PR_BASE="$("$GH" pr view "$PR_NUMBER" --json baseRefName --jq '.baseRefName' 2>/dev/null)" || PR_BASE=""
fi
if [ -z "$PR_BASE" ]; then
    abort "cannot resolve PR #${PR_NUMBER} base branch after retry — refusing to watch with a guessed SSoT key (an empty required list scores all-passed)" 1
fi

# Verify gh is available.
if ! command -v "$GH" >/dev/null 2>&1; then
    abort "gh CLI not found at '$GH'. Install via: brew install gh && gh auth login" 1
fi

# t1534 gate round: jq is a REQUIRED dependency — the awk JSON fallbacks
# were each a measured defect source (compact arrays, quoted names,
# duplicate entries). Fail loudly rather than misclassify.
if ! command -v jq >/dev/null 2>&1; then
    abort "jq not found — the ci-watch loop requires jq for check classification (install: brew install jq)" 1
fi

# ─── SSoT override for testing ────────────────────────────────────────────────
if [ -n "${MOAI_CIWATCH_REQUIRED_CHECKS_FILE:-}" ]; then
    REQUIRED_CHECKS_FILE="$MOAI_CIWATCH_REQUIRED_CHECKS_FILE"
fi
if [ ! -f "$REQUIRED_CHECKS_FILE" ]; then
    abort "required-checks.yml not found at $REQUIRED_CHECKS_FILE" 1
fi

# GATE-6 P1: verify the base-branch KEY in the SSoT BEFORE watching — the
# loader still prints a newline for an unknown key, so a file-size check on
# the loaded list passes and the required list reads empty, which scores
# all-passed with every failure advisory (gate repro: base=develop +
# Lint=fail surfaced exit 0). A release/* base resolves to the release/*
# SSoT pattern key.
# GATE-7: key presence is judged by the YAML parser when yq is available —
# a QUOTED key (`"main":`) is valid YAML the raw grep could not see, and it
# aborted a perfectly watchable PR.
_ssot_has_key() {
    if command -v yq >/dev/null 2>&1; then
        yq -r '.branches | keys | .[]' "$REQUIRED_CHECKS_FILE" 2>/dev/null | grep -qxF "$1"
    else
        grep -qF "  $1:" "$REQUIRED_CHECKS_FILE"
    fi
}
SSOT_BRANCH=""
_ssot_has_key "$PR_BASE" && SSOT_BRANCH="$PR_BASE"
if [ -z "$SSOT_BRANCH" ]; then
    case "$PR_BASE" in
        release/*)
            _ssot_has_key "release/*" && SSOT_BRANCH="release/*"
            ;;
    esac
fi
if [ -z "$SSOT_BRANCH" ]; then
    abort "required-checks SSoT has no key for base branch '${PR_BASE}' — refusing to watch an unkeyed base (an empty required list scores all-passed)" 1
fi
if [ "$PR_BASE" != "$BRANCH" ]; then
    log_step "PR base branch '${PR_BASE}' differs from head '${BRANCH}' — classifying against SSoT key '${SSOT_BRANCH}'"
fi

# ─── start timer ──────────────────────────────────────────────────────────────
ciwatch_start_timer
log_step "Watching PR #${PR_NUMBER} on branch '${BRANCH}'"

# ─── classify helpers ─────────────────────────────────────────────────────────

# _check_bucket extracts the gh pr checks bucket for a named check.
# t1534 M3: `gh pr checks --json` supports ONLY name/state/bucket/link —
# the pre-M3 field list (status/conclusion/detailsUrl) made every poll abort
# with "Unknown JSON field: 'status'" (card repro). The bucket is the
# authoritative classification: pass|fail|pending|skipping|cancel.
_check_bucket() {
    check_name="$1"
    json_file="$2"
    # t1534 gate round: aggregate EVERY bucket entry for the name (real PR
    # JSON carries duplicate names — push + PR runs both publish) and emit
    # exactly ONE worst-case verdict: fail/cancel > pending/skipping > pass.
    # The pre-fix awk emitted per-line verdicts AND an END verdict (a
    # pass+fail pair scored exit 3 = timeout in the reviewer's repro).
    jq -r --arg n "$check_name" \
        '.[] | select(.name==$n) | .bucket // "pending"' "$json_file" \
        | awk '
            function worse(cur, cand) {
                # cancel aggregates as FAIL — a cancelled required check is
                # never a pass regardless of entry order (gate repro:
                # [pass, cancel] scored exit 0 with the pass-keeping order).
                if (cur == "fail" || cand == "fail" || cur == "cancel" || cand == "cancel") return "fail"
                # GATE-5: skipping aggregates as PENDING in every order — a
                # skipped run published no verdict, and [pass, skipping]
                # vs [skipping, pass] used to disagree on the verdict.
                if (cur == "pending" || cand == "pending" || cur == "skipping" || cand == "skipping") return "pending"
                if (cur == "") return cand
                return cur
            }
            { verdict = worse(verdict, $0) }
            END { print (verdict == "" ? "pending" : verdict) }
        '
}

# _check_link extracts the link for a named check from JSON array.
_check_link() {
    check_name="$1"
    json_file="$2"
    jq -r --arg n "$check_name" \
        '.[] | select(.name==$n) | .link // ""' "$json_file" \
        | head -1
}

# _all_check_names lists all check names from JSON array.
_all_check_names() {
    json_file="$1"
    jq -r '.[].name' "$json_file"
}

# (_check_status removed — t1534 M3: `gh pr checks --json` has no `status`
# field; classification runs on the bucket via _check_bucket.)

# ─── main poll loop ───────────────────────────────────────────────────────────

TMP_JSON="$(mktemp /tmp/ciwatch_checks_XXXXXXXXXX)"
TMP_SSOT="$(mktemp /tmp/ciwatch_ssot_XXXXXXXXXX)"
TMP_ALL="$(mktemp /tmp/ciwatch_all_XXXXXXXXXX)"
TMP_FN="$(mktemp /tmp/ciwatch_fn_XXXXXXXXXX)"
TMP_FL="$(mktemp /tmp/ciwatch_fl_XXXXXXXXXX)"
TMP_PAIR="$(mktemp /tmp/ciwatch_pair_XXXXXXXXXX)"
trap 'rm -f "$TMP_JSON" "$TMP_SSOT" "$TMP_ALL" "$TMP_FN" "$TMP_FL" "$TMP_PAIR"' EXIT

# t1534 M3: load the SSoT required contexts for this branch — yq when
# available, otherwise an awk scan over the branches.<key>.contexts block.
_load_required_contexts() {
    branch_key="$1"
    if command -v yq >/dev/null 2>&1; then
        # GATE P1 fix: a yq FAILURE (malformed YAML, unreadable file) must
        # abort — the pre-fix `|| true` turned an unreadable SSoT into an
        # empty list, which the loop read as "all required passed" (exit 0).
        # A successful yq run with an empty contexts list is still a legal
        # empty result (e.g. release/* per decision-index Q1).
        yq_out="$(yq -r ".branches[\"$branch_key\"].contexts // [] | .[]" "$REQUIRED_CHECKS_FILE" 2>/dev/null)" || {
            abort "yq failed to read $REQUIRED_CHECKS_FILE — refusing to treat an unreadable SSoT as 'all required passed'" 1
        }
        printf '%s\n' "$yq_out"
    else
        # GATE P1 fix: match the REAL SSoT shape — branch keys at indent 2
        # UNQUOTED (`  main:`) and contexts items at indent 6 (`      - `).
        # The first draft matched a quoted `"main":` at indent 4 and yielded
        # an empty list, which read as "all required passed" (exit 0).
        awk -v key="$branch_key" '
            /^branches:/ { inb = 1; next }
            inb && /^[A-Za-z_*]/ { inb = 0 }
            inb && $0 ~ ("^  " key ":") { inctx = 1; next }
            # the key block owns a contexts: line at indent 4 — it OPENS the
            # item list and must not terminate the scan (GATE P1 fix, second
            # shape: the indent-4 terminator rule matched `    contexts:`
            # itself and closed the scan before any item line).
            inctx && /^    contexts:/ { next }
            inctx && /^      - / {
                line = $0
                sub(/^ *- */, "", line)
                gsub(/"/, "", line)
                print line
            }
            inctx && /^    [A-Za-z_*]/ { inctx = 0 }
        ' "$REQUIRED_CHECKS_FILE"
    fi
}

while true; do
    ciwatch_check_timeout

    # Fetch current checks state. t1534 M3: `gh pr checks --json` supports
    # ONLY name/state/bucket/link — the pre-M3 field list asked for
    # status/conclusion/detailsUrl, which made every poll abort instantly
    # with "Unknown JSON field: 'status'" (the card's repro).
    if ! "$GH" pr checks "$PR_NUMBER" --json "name,state,bucket,link" >"$TMP_JSON" 2>/dev/null; then
        abort "gh pr checks failed for PR #${PR_NUMBER} — check gh auth and PR number" 1
    fi

    _load_required_contexts "$SSOT_BRANCH" > "$TMP_SSOT"
    # GATE-6 P1: key presence is verified ONCE before the loop (above) —
    # the loader prints a bare newline for an unknown key, so a file-size
    # check here passed an empty required list; release/* bases are
    # resolved to the release/* pattern key there too.

    # Classify SSoT required checks. t1534 M3: iterate the SSoT list
    # newline-safely (the pre-M3 loop word-split $(_all_check_names),
    # shredding contexts containing spaces) and treat a required check
    # ABSENT from the response as pending — pre-M3, iterating only the
    # returned names made a not-yet-published required check silently
    # count as passed.
    required_pass=0
    required_fail=0
    required_pending=0
    total_required=0
    aux_fail=0
    failed_names=""
    failed_links=""

    while IFS= read -r check_name; do
        [ -n "$check_name" ] || continue
        total_required=$((total_required + 1))
        bucket="$(_check_bucket "$check_name" "$TMP_JSON")"
        case "$bucket" in
            pass)
                required_pass=$((required_pass + 1))
                ;;
            fail|cancel)
                required_fail=$((required_fail + 1))
                link="$(_check_link "$check_name" "$TMP_JSON")"
                failed_names="${failed_names}${check_name}|"
                failed_links="${failed_links}${link}|"
                ;;
            *)
                # pending, skipping, or absent from the response ("") — the
                # check has not published a verdict yet: still pending.
                required_pending=$((required_pending + 1))
                ;;
        esac
    done < "$TMP_SSOT"

    # Auxiliary checks: every RESPONDED check the SSoT does not require on
    # this branch — is_required (classify.sh) was unused pre-M3 and now
    # governs the required/auxiliary split.
    _all_check_names "$TMP_JSON" > "$TMP_ALL"
    while IFS= read -r check_name; do
        [ -n "$check_name" ] || continue
        is_required "$check_name" "$SSOT_BRANCH" && continue
        bucket="$(_check_bucket "$check_name" "$TMP_JSON")"
        if [ "$bucket" = "fail" ] || [ "$bucket" = "cancel" ]; then
            aux_fail=$((aux_fail + 1))
            log_step "ADVISORY: '${check_name}' failed (non-blocking)"
        fi
    done < "$TMP_ALL"

    # Compute total required (sum of known states — note: we re-count from scratch each tick).
    total_required=$((required_pass + required_fail + required_pending))

    # Emit status update to stderr.
    log_step "PR #${PR_NUMBER}: required ${required_pass}/${total_required} pass, ${required_pending} pending, ${required_fail} failed; advisory ${aux_fail} fail"

    # State machine transitions.
    if [ "$required_fail" -gt 0 ]; then
        # Required check failed — emit JSON handoff to stdout for orchestrator.
        log_step "Required failure detected — emitting T3 handoff JSON"

        # Build the T3 handoff JSON (ci-watch-protocol schema: failedChecks
        # name/link pairs + auxiliaryFailCount + totalRequired). t1534 M3:
        # the name/link pairs travel via tr + paste + while-read — the
        # pre-M3 `set -- $failed_names` word-split re-shredded required
        # contexts containing spaces ("Build (linux/amd64)" → "Build" +
        # "(linux/amd64)").
        printf '%s' "$failed_names" | tr '|' '\n' | sed '/^$/d' > "$TMP_FN"
        printf '%s' "$failed_links" | tr '|' '\n' | sed '/^$/d' > "$TMP_FL"
        paste "$TMP_FN" "$TMP_FL" > "$TMP_PAIR"

        printf '{"prNumber":%s,"branch":"%s","totalRequired":%s,"failedChecks":[' \
            "$PR_NUMBER" "$BRANCH" "$total_required"
        first=1
        while IFS="$(printf '\t')" read -r name link; do
            [ -n "$name" ] || continue
            if [ "$first" = "1" ]; then first=0; else printf ','; fi
            printf '{"name":"%s","logUrl":"%s"}' "$name" "$link"
        done < "$TMP_PAIR"
        printf '],"auxiliaryFailCount":%s}\n' "$aux_fail"
        exit 2
    fi

    if [ "$required_pending" -eq 0 ] && [ "$required_fail" -eq 0 ]; then
        # All required checks passed.
        if [ "$aux_fail" -gt 0 ]; then
            log_step "ADVISORY: ${aux_fail} auxiliary check(s) failed (non-blocking)"
        fi
        log_step "All required checks passed for PR #${PR_NUMBER}"
        exit 0
    fi

    # Still pending — sleep and loop.
    if [ -n "${MOAI_CIWATCH_NO_SLEEP:-}" ]; then
        # Testing mode: exit after one tick to avoid infinite loop.
        # If there are pending checks in test mode, treat as pass for simplicity.
        log_step "NO_SLEEP mode: exiting after single poll tick (pending=${required_pending})"
        exit 0
    fi

    sleep "$POLL_INTERVAL"
done
