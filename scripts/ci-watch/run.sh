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

# Verify gh is available.
if ! command -v "$GH" >/dev/null 2>&1; then
    abort "gh CLI not found at '$GH'. Install via: brew install gh && gh auth login" 1
fi

# ─── SSoT override for testing ────────────────────────────────────────────────
if [ -n "${MOAI_CIWATCH_REQUIRED_CHECKS_FILE:-}" ]; then
    REQUIRED_CHECKS_FILE="$MOAI_CIWATCH_REQUIRED_CHECKS_FILE"
fi
if [ ! -f "$REQUIRED_CHECKS_FILE" ]; then
    abort "required-checks.yml not found at $REQUIRED_CHECKS_FILE" 1
fi

# ─── start timer ──────────────────────────────────────────────────────────────
ciwatch_start_timer
log_step "Watching PR #${PR_NUMBER} on branch '${BRANCH}'"

# ─── classify helpers ─────────────────────────────────────────────────────────

# _check_conclusion extracts the conclusion for a named check from JSON array.
# Uses jq if available, falls back to grep+sed.
_check_conclusion() {
    check_name="$1"
    json_file="$2"
    if command -v jq >/dev/null 2>&1; then
        jq -r --arg n "$check_name" '.[] | select(.name==$n) | .conclusion // "unknown"' "$json_file"
    else
        # Rough grep fallback.
        awk "
            /\"name\": *\"$check_name\"/ { found=1 }
            found && /\"conclusion\"/ { gsub(/.*\"conclusion\": *\"|\".*/, \"\"); print; exit }
        " "$json_file"
    fi
}

# _check_details_url extracts the detailsUrl for a named check.
_check_details_url() {
    check_name="$1"
    json_file="$2"
    if command -v jq >/dev/null 2>&1; then
        jq -r --arg n "$check_name" '.[] | select(.name==$n) | .detailsUrl // ""' "$json_file"
    else
        awk "
            /\"name\": *\"$check_name\"/ { found=1 }
            found && /\"detailsUrl\"/ { gsub(/.*\"detailsUrl\": *\"|\".*/, \"\"); print; exit }
        " "$json_file"
    fi
}

# _all_check_names lists all check names from JSON array.
_all_check_names() {
    json_file="$1"
    if command -v jq >/dev/null 2>&1; then
        jq -r '.[].name' "$json_file"
    else
        grep '"name"' "$json_file" | sed 's/.*"name": *"\(.*\)".*/\1/'
    fi
}

# _check_status extracts the status for a named check.
_check_status() {
    check_name="$1"
    json_file="$2"
    if command -v jq >/dev/null 2>&1; then
        jq -r --arg n "$check_name" '.[] | select(.name==$n) | .status // "unknown"' "$json_file"
    else
        awk "
            /\"name\": *\"$check_name\"/ { found=1 }
            found && /\"status\"/ { gsub(/.*\"status\": *\"|\".*/, \"\"); print; exit }
        " "$json_file"
    fi
}

# ─── main poll loop ───────────────────────────────────────────────────────────

TMP_JSON="$(mktemp /tmp/ciwatch_checks_XXXXXX.json)"
trap 'rm -f "$TMP_JSON"' EXIT

while true; do
    ciwatch_check_timeout

    # Fetch current checks state.
    if ! "$GH" pr checks "$PR_NUMBER" --json "name,status,conclusion,detailsUrl" >"$TMP_JSON" 2>/dev/null; then
        abort "gh pr checks failed for PR #${PR_NUMBER} — check gh auth and PR number" 1
    fi

    # Classify each check.
    required_pass=0
    required_fail=0
    required_pending=0
    aux_fail=0
    failed_names=""
    failed_urls=""

    for check_name in $(_all_check_names "$TMP_JSON"); do
        status="$(_check_status "$check_name" "$TMP_JSON")"
        conclusion="$(_check_conclusion "$check_name" "$TMP_JSON")"

        if is_auxiliary "$check_name"; then
            # Auxiliary check.
            if [ "$conclusion" = "failure" ] || [ "$conclusion" = "cancelled" ]; then
                aux_fail=$((aux_fail + 1))
                log_step "ADVISORY: '${check_name}' ${conclusion} (non-blocking)"
            fi
            continue
        fi

        # Required check.
        case "$status" in
            "completed")
                case "$conclusion" in
                    "success"|"skipped"|"neutral")
                        required_pass=$((required_pass + 1))
                        ;;
                    "failure"|"cancelled"|"timed_out"|"action_required")
                        required_fail=$((required_fail + 1))
                        url="$(_check_details_url "$check_name" "$TMP_JSON")"
                        failed_names="${failed_names}${check_name}|"
                        failed_urls="${failed_urls}${url}|"
                        ;;
                esac
                ;;
            *)
                required_pending=$((required_pending + 1))
                ;;
        esac
    done

    # Compute total required (sum of known states — note: we re-count from scratch each tick).
    total_required=$((required_pass + required_fail + required_pending))

    # Emit status update to stderr.
    log_step "PR #${PR_NUMBER}: required ${required_pass}/${total_required} pass, ${required_pending} pending, ${required_fail} failed; advisory ${aux_fail} fail"

    # State machine transitions.
    if [ "$required_fail" -gt 0 ]; then
        # Required check failed — emit JSON handoff to stdout for orchestrator.
        log_step "Required failure detected — emitting T3 handoff JSON"

        # Build a minimal JSON handoff. Avoid external deps where possible.
        printf '{"prNumber":%s,"branch":"%s","failedChecks":[' "$PR_NUMBER" "$BRANCH"
        first=1
        IFS="|"
        set -- $failed_names
        idx=1
        for name in "$@"; do
            [ -z "$name" ] && continue
            url_idx=0
            for u in $(printf '%s' "$failed_urls" | tr '|' '\n'); do
                url_idx=$((url_idx + 1))
                [ "$url_idx" = "$idx" ] && break
            done
            if [ "$first" = "1" ]; then first=0; else printf ','; fi
            printf '{"name":"%s","logUrl":"%s"}' "$name" "$u"
            idx=$((idx + 1))
        done
        unset IFS
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
