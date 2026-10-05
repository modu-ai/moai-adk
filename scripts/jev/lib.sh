#!/bin/sh
# Shared helpers for the local-only Jev scripts. Dev-only; not distributed.
#
# Credential posture: the key lives in ~/.moai/.env.typesafe (chmod 600) and
# nowhere else. Absence is not an error — callers degrade to the mechanical
# measurement and exit 0, the same fail-open shape glm_audit uses.

JEV_ENV_FILE="${JEV_ENV_FILE:-$HOME/.moai/.env.typesafe}"
JEV_API="${JEV_API:-https://api.typesafe.ai/v1/systemone}"
JEV_MODEL="${JEV_MODEL:-jev-latest}"

# jev_load_key: sets TYPESAFE_API_KEY when the env file carries one.
# Returns 1 when no key is available, printing nothing.
jev_load_key() {
    if [ -n "${TYPESAFE_API_KEY:-}" ]; then
        return 0
    fi
    [ -f "$JEV_ENV_FILE" ] || return 1
    # shellcheck disable=SC1090
    . "$JEV_ENV_FILE"
    [ -n "${TYPESAFE_API_KEY:-}" ] || return 1
    export TYPESAFE_API_KEY
    return 0
}

# jev_post <json-body>: POSTs to the API, prints the raw response on stdout.
# Prints the HTTP status to stderr when it is not 200 so a failure is never
# silent, and returns curl's own exit status.
jev_post() {
    body="$1"
    tmp_status="$(mktemp)"
    tmp_rc="$(mktemp)"
    printf '%s' "$body" | {
        curl -sS -X POST "$JEV_API" \
            -H "Authorization: Bearer $TYPESAFE_API_KEY" \
            -H "Content-Type: application/json" \
            -w '%{http_code}' -o /dev/stdout \
            --data-binary @- 2>"$tmp_status"
        # curl's exit code must survive the pipeline: the pipe's status is
        # the reader block's, so curl records its own rc for jev_post to
        # return (documented contract, line 28).
        echo $? >"$tmp_rc"
    } | {
        # curl -w appends the status to stdout; split it off.
        out="$(cat)"
        code="$(printf '%s' "$out" | tail -c 3)"
        # Strip the status from the LAST line only: an unaddressed sed strips
        # three characters from EVERY body line of a multi-line response
        # (codex review gate reproduction — "ok": true became "ok": t).
        printf '%s' "$out" | sed '$s/...$//'
        [ "$code" = "200" ] || printf 'jev: HTTP %s\n' "$code" >&2
    }
    rc="$(cat "$tmp_rc")"
    [ -s "$tmp_status" ] && cat "$tmp_status" >&2
    rm -f "$tmp_status" "$tmp_rc"
    return "$rc"
}
