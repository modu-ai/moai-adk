#!/usr/bin/env bash
# session_drain.sh — durable session-start drain wrapper (SPEC-LSEL-DRAIN-STALL-001 M1).
#
# WHY: the drain's only previous "schedule" was a session-scoped /loop recipe that
# died with its owning session (2026-08-04; the 3-week stall) and executed nothing
# even while alive (console.log-only). This wrapper is the durable machine trigger.
#
# Contract (REQ-LDS-001..005):
#   1. Acquire an exclusive drain lock on the state dir. Contention -> skip +
#      notice + exit 0 (concurrent session starts are NORMAL here, REQ-LDS-002).
#      Lock means: mkdir-based atomic lock — POSIX-universal with zero PATH
#      dependence (homebrew flock works on this machine and supports the fd form,
#      but a missing flock binary at hook time is indistinguishable from permanent
#      contention: every drain would skip forever, silently recreating the exact
#      stall this wrapper exists to end; AC-LDS-002 verifies behavior, the means
#      is free). A stale lock (holder crashed) is reaped when older than 120s —
#      the measured full-backlog drain is <1s, so this never fires on a live
#      drain. rmdir only ever removes an EMPTY dir; a corrupted non-empty lock
#      dir reads as contention forever and needs manual cleanup (never auto-deleted).
#   2. UNCONDITIONALLY archive any existing clusters.json to
#      <state-dir>/clusters-history/ BEFORE invoking drain.sh — drain.sh
#      overwrites clusters.json on BOTH the drain path and the empty-delta no-op
#      path, so a conditional archive would lose staged candidates (REQ-LDS-003,
#      spec section B.5). Newest 100 archives kept (retention is implementation
#      discretion; the duty is the copy before any overwrite).
#   3. Invoke drain.sh unchanged (the frozen mechanical core, byte-preserved).
#   4. Emit a one-line status to stderr: stubs read / candidates / new offset
#      (no-op stated as such, REQ-LDS-004).
#   5. FAIL-OPEN (REQ-LDS-005): any internal error degrades to a stderr notice
#      and exit 0 — this hook must NEVER block session start. The EXIT trap
#      forces exit 0 by construction. Runtime budget: measured <1s for a 1.1MB /
#      ~4.2k-line inbox; the wiring sets an explicit 30s hook timeout (live
#      SessionStart precedent, spec.md section E).
#
# Writes ONLY under --state-dir (the .moai/state/lsel surface). Never memory/.
# The inbox is append-only and never mutated. Wiring (settings.local.json) is a
# maintainer-machine local deliverable applied in M2 — a tracked settings.json
# entry would be wiped by every moai update (REQ-LDS-009).
#
# Usage:
#   session_drain.sh [--inbox <path>] [--state-dir <dir>]
#   (defaults: .moai/lessons-inbox.jsonl / .moai/state/lsel — hook-CWD relative)
set -uo pipefail   # NOTE: deliberately NO -e — every failure is handled and fails open.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DRAIN="$SCRIPT_DIR/drain.sh"

INBOX=".moai/lessons-inbox.jsonl"
STATE_DIR=".moai/state/lsel"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --inbox)     INBOX="$2"; shift 2;;
    --state-dir) STATE_DIR="$2"; shift 2;;
    -h|--help)   sed -n '2,42p' "$0"; exit 0;;
    *) echo "session_drain.sh: unknown argument: $1" >&2; exit 2;;
  esac
done
[[ -n "$INBOX" && -n "$STATE_DIR" ]] || { echo "usage: session_drain.sh --inbox <path> --state-dir <dir>" >&2; exit 2; }

# ---------------------------------------------------------------------------
# From here on: fail-open. The EXIT trap releases the lock (if held) and forces
# exit 0 — even an unbound-variable abort cannot leak a non-zero hook exit.
# ---------------------------------------------------------------------------
LOCK_DIR=""
on_exit() {
  if [[ -n "$LOCK_DIR" ]]; then rmdir "$LOCK_DIR" 2>/dev/null; fi
  exit 0
}
trap on_exit EXIT

fail_open() {  # $1 = short reason; the drain is skipped, session start proceeds
  echo "session_drain: $1 — degraded to no-op (fail-open; next session start retries)" >&2
}

# --- step 0: preconditions ----------------------------------------------------
if [[ ! -f "$INBOX" ]]; then
  echo "session_drain: inbox absent ($INBOX) — no-op (fresh clone or no observations yet)" >&2
  exit 0
fi
if [[ ! -x "$DRAIN" ]]; then
  fail_open "drain.sh missing or not executable at $DRAIN"
  exit 0
fi
if ! mkdir -p "$STATE_DIR" 2>/dev/null; then
  fail_open "cannot create state dir $STATE_DIR"
  exit 0
fi

# --- step 1: exclusive drain lock (REQ-LDS-001/002) ----------------------------
LOCK_DIR="$STATE_DIR/drain.lock"
if ! mkdir "$LOCK_DIR" 2>/dev/null; then
  if [[ -n "$(find "$LOCK_DIR" -maxdepth 0 -mmin +2 -print 2>/dev/null)" ]] \
     && rmdir "$LOCK_DIR" 2>/dev/null \
     && mkdir "$LOCK_DIR" 2>/dev/null; then
    :  # stale lock reaped and re-acquired
  else
    echo "session_drain: lock contention — another session holds the drain lock; skipping this drain (no-op)" >&2
    LOCK_DIR=""  # not ours — never release another holder's lock
    exit 0
  fi
fi

# --- step 2: unconditional archive-before-overwrite (REQ-LDS-003) --------------
CLUSTERS_FILE="$STATE_DIR/clusters.json"
HISTORY_DIR="$STATE_DIR/clusters-history"
if [[ -f "$CLUSTERS_FILE" ]]; then
  if ! mkdir -p "$HISTORY_DIR" 2>/dev/null; then
    fail_open "cannot create $HISTORY_DIR — drain skipped (refusing to overwrite unarchived candidates)"
    exit 0
  fi
  ARCHIVE="$HISTORY_DIR/clusters-$(date -u +%Y%m%dT%H%M%SZ)-$$.json"
  if ! cp -p "$CLUSTERS_FILE" "$ARCHIVE" 2>/dev/null; then
    fail_open "cannot archive clusters.json — drain skipped (refusing to overwrite unarchived candidates)"
    exit 0
  fi
  # retention (discretionary, acceptance section E): keep the newest 100 archives
  ls -t "$HISTORY_DIR" 2>/dev/null | tail -n +101 | while IFS= read -r stale; do
    rm -f "$HISTORY_DIR/$stale"
  done
fi

# --- step 3: the frozen mechanical core (drain.sh, byte-preserved) -------------
if ! "$DRAIN" --inbox "$INBOX" --state-dir "$STATE_DIR"; then
  fail_open "drain.sh exited non-zero (tooling or inbox problem) — offset NOT advanced"
  exit 0
fi

# --- step 4: one-line status (REQ-LDS-001; no-op stated per REQ-LDS-004) -------
READ_N="$(jq -r '.total_read // 0' "$CLUSTERS_FILE" 2>/dev/null)" || READ_N=""
CAND_N="$(jq -r '(.candidates // []) | length' "$CLUSTERS_FILE" 2>/dev/null)" || CAND_N=""
OFF_N="$(jq -r '.offset_after // 0' "$CLUSTERS_FILE" 2>/dev/null)" || OFF_N=""
if [[ -n "$READ_N" && -n "$CAND_N" && -n "$OFF_N" ]]; then
  if [[ "$READ_N" == "0" ]]; then
    echo "session_drain: no-op read=0 candidates=$CAND_N offset=$OFF_N (offset already at inbox tail)" >&2
  else
    echo "session_drain: read=$READ_N candidates=$CAND_N offset=$OFF_N" >&2
  fi
else
  fail_open "cannot parse drain status from $CLUSTERS_FILE"
fi

exit 0
