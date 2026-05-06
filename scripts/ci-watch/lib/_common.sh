#!/bin/sh
# scripts/ci-watch/lib/_common.sh — POSIX sh helpers for ci-watch
# Mirrors the Wave 1 ci-mirror pattern; do NOT use bash-isms.

# log_step prints a timestamped step message to stderr.
log_step() { printf '[ci-watch] %s\n' "$1" >&2; }

# abort prints a fatal error message and exits with the given code (default 99).
abort() { printf '[ci-watch][FATAL] %s\n' "$1" >&2; exit "${2:-99}"; }

# posix_now returns the current Unix timestamp (seconds since epoch).
posix_now() { date +%s; }

# status_dump prints a key=value line to stderr for structured logging.
status_dump() { printf '[ci-watch][STATUS] %s=%s\n' "$1" "$2" >&2; }
