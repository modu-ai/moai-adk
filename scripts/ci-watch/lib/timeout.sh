#!/bin/sh
# scripts/ci-watch/lib/timeout.sh — 30-min wall-clock guard for ci-watch loop
# Sources: lib/_common.sh must be sourced before this file.

# CIWATCH_TIMEOUT_SECONDS: maximum wall-clock seconds for the watch loop.
# Default: 1800 (30 minutes). Override via env for testing.
CIWATCH_TIMEOUT_SECONDS="${CIWATCH_TIMEOUT_SECONDS:-1800}"

# ciwatch_start_timer initializes the wall-clock start time.
# Must be called once at the start of the watch loop.
ciwatch_start_timer() {
    CIWATCH_LOOP_START="$(posix_now)"
    export CIWATCH_LOOP_START
}

# ciwatch_check_timeout returns 0 (OK) or exits with code 3 if the
# 30-min hard timeout has been reached.
ciwatch_check_timeout() {
    now="$(posix_now)"
    elapsed="$((now - CIWATCH_LOOP_START))"
    if [ "$elapsed" -ge "$CIWATCH_TIMEOUT_SECONDS" ]; then
        log_step "Hard timeout reached after ${elapsed}s (limit ${CIWATCH_TIMEOUT_SECONDS}s) — aborting watch"
        exit 3
    fi
}

# ciwatch_elapsed_seconds returns the number of seconds since ciwatch_start_timer.
ciwatch_elapsed_seconds() {
    now="$(posix_now)"
    printf '%d' "$((now - CIWATCH_LOOP_START))"
}
