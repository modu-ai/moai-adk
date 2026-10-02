#!/bin/bash
# hook-probe.sh — behavioural probe for SPEC-FACTORY-STALE-RUN-HEAL-001.
#
# usage: bash hook-probe.sh <scenario> [moai-binary]
# The binary defaults to ./bin/moai-t1345 (bin/ is gitignored); build it from the tree first:
#   go build -o ./bin/moai-t1345 ./cmd/moai
#
# Builds a throw-away project + factory database in a temp directory OUTSIDE the
# repository (isolated HOME/MOAI_HOME, git-initialised project), seeds run rows,
# drives the real `moai hook ...` subcommands of <moai-binary> with a lane
# environment, prints what each hook turn said, and ends with ONE verdict line:
#   VERDICT: PASS   (exit 0)  — the post-implementation behaviour holds
#   VERDICT: FAIL <reason> (exit 1) — it does not (the RED-now state)
# It never writes inside the repository. Requires: bash, git, jq, sqlite3.
#
# Scenarios (each is one acceptance cell in acceptance.md):
#   control-healthy      positive control: a current-vocabulary lane on an active run binds
#   rebind                a current-vocabulary lane on a retired run, one other run active
#   unbind-then-rebind    the same lane with NO active run, then a run becomes active
#   ambiguous             the same lane, two other runs active
#   legacy-lines          a legacy-label lane: the three command-line rows of the notice table
#   session-start-silent  SessionStart (source clear) for the rebind-shaped lane
#   roundtrip             the printed legacy relaunch line is accepted by the verb (--dry-run)
#   dry-run-nonmutation   `moai factory relaunch --dry-run --from-run` writes nothing

SCEN="${1:-}"
BIN="${2:-./bin/moai-t1345}"
if [ -z "$SCEN" ] || [ ! -x "$BIN" ]; then
  echo "usage: bash hook-probe.sh <scenario> [moai-binary]   (default binary: ./bin/moai-t1345, run from the repo root)" >&2
  exit 2
fi
BIN="$(cd "$(dirname "$BIN")" && pwd)/$(basename "$BIN")"

T="$(mktemp -d)"
trap 'rm -rf "$T"' EXIT
PROJ="$T/proj"
mkdir -p "$T/home" "$PROJ/.moai"
export HOME="$T/home" MOAI_HOME="$T/home/.moai"
cd "$PROJ" || exit 2
git init -q . && git -c user.email=probe@example.invalid -c user.name=probe commit -q --allow-empty -m init
"$BIN" factory runs >/dev/null 2>&1
DB="$(find "$T/home" -name factory.db | head -1)"
FDIR="$(dirname "$DB")"
if [ -z "$DB" ]; then echo "VERDICT: FAIL no factory database was created"; exit 1; fi

seed() { # seed <run> <status> [pid] [start]
  sqlite3 "$DB" "INSERT OR REPLACE INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,created_at,updated_at) VALUES('$1','','test','$2','{}',${3:-0},'${4:-}','2026-10-02T00:00:00Z','2026-10-02T00:00:00Z');"
}
fail() { echo "VERDICT: FAIL $*"; exit 1; }

# hook <event> <label> <env-run> <session-id> -> decoded additionalContext on stdout
hook() {
  local ev="$1" label="$2" envrun="$3" sid="$4" out
  export CLAUDE_PROJECT_DIR="$PROJ" MOAI_KANBAN_ID="$envrun" MOAI_FACTORY_WORKERS=4 MOAI_FACTORY_WORKER="$label" MOAI_KANBAN_BACKEND=claude MOAI_SESSION_PID=$$
  if [ "$ev" = prompt ]; then
    out="$(printf '{"session_id":"%s","cwd":"%s","hook_event_name":"UserPromptSubmit","prompt":"please continue the card"}' "$sid" "$PROJ" | "$BIN" hook user-prompt-submit 2>/dev/null)"
  else
    out="$(printf '{"session_id":"%s","cwd":"%s","hook_event_name":"SessionStart","source":"clear"}' "$sid" "$PROJ" | "$BIN" hook session-start 2>/dev/null)"
  fi
  printf '%s' "$out" | jq -r '.hookSpecificOutput.additionalContext // ""' 2>/dev/null
}
show() { printf '%s:\n%s\n' "$1" "$2" | sed 's/^/  /'; }
has_line() { printf '%s\n' "$2" | grep -qxF -- "$1"; }          # has_line <exact line> <text>
broker_peers() { # broker_peers <run> -> "slot|session" rows, empty if no broker
  local b="$FDIR/messages/$1/broker.db"
  [ -f "$b" ] && sqlite3 "$b" "SELECT slot||'|'||session_uuid FROM peers;" 2>/dev/null
}

case "$SCEN" in
control-healthy)
  # Positive control: a current-vocabulary lane on an ACTIVE run binds as it always has.
  # It must PASS on the pre-change tree and after the change (REQ-SRH-009) — it proves the
  # fixture and driver can reach a PASS, so the FAIL verdicts above are not vacuous.
  seed runX active
  p1="$(hook prompt lane-3 runX s1)"
  show "prompt 1" "$p1"
  case "$p1" in *"factory messaging bound: run=runX slot=lane-3"*) ;; *) fail "healthy lane did not bind into runX";; esac
  [ "$(broker_peers runX)" = "lane-3|s1" ] || fail "runX broker holds no peer lane-3|s1"
  echo "VERDICT: PASS"; exit 0;;
rebind)
  seed runX retired; seed runY active
  p1="$(hook prompt lane-3 runX s1)"; p2="$(hook prompt lane-3 runX s1)"
  show "prompt 1" "$p1"; show "prompt 2" "$p2"
  show "peers in runY" "$(broker_peers runY)"
  [ -f "$FDIR/messages/runX/broker.db" ] && fail "the retired run's broker was opened"
  case "$p1" in *degraded*) fail "prompt 1 still says degraded";; esac
  case "$p1" in *"factory lane rebound:"*runX*runY*) ;; *) fail "prompt 1 carries no rebound notice naming runX and runY";; esac
  case "$p1" in *lane-3*) ;; *) fail "rebound notice does not name slot lane-3";; esac
  [ "$(broker_peers runY)" = "lane-3|s1" ] || fail "runY broker holds no peer lane-3|s1"
  case "$p2" in *factory\ lane*|*degraded*) fail "prompt 2 repeated a notice";; esac
  echo "VERDICT: PASS"; exit 0;;
unbind-then-rebind)
  seed runX retired
  p1="$(hook prompt lane-3 runX s1)"; p2="$(hook prompt lane-3 runX s1)"
  show "prompt 1 (no active run)" "$p1"; show "prompt 2 (no active run)" "$p2"
  seed runY active
  p3="$(hook prompt lane-3 runX s1)"
  show "prompt 3 (runY now active)" "$p3"
  case "$p1" in *degraded*) fail "prompt 1 says degraded instead of the one-time unbound notice";; esac
  case "$p1" in *"factory lane unbound:"*lane-3*runX*) ;; *) fail "prompt 1 carries no unbound notice naming lane-3 and runX";; esac
  has_line "moai factory relaunch" "$p1" && fail "unbound notice prints a command with no active run"
  case "$p2" in *factory\ lane*|*degraded*) fail "prompt 2 repeated the unbound notice";; esac
  case "$p3" in *"factory lane rebound:"*runY*) ;; *) fail "prompt 3 did not rebind into runY";; esac
  [ "$(broker_peers runY)" = "lane-3|s1" ] || fail "runY broker holds no peer lane-3|s1"
  echo "VERDICT: PASS"; exit 0;;
ambiguous)
  seed runX retired; seed runY active; seed runZ active
  p1="$(hook prompt lane-3 runX s1)"; p2="$(hook prompt lane-3 runX s1)"
  show "prompt 1" "$p1"; show "prompt 2" "$p2"
  case "$p1" in *degraded*) fail "prompt 1 says degraded";; esac
  has_line "moai factory relaunch --provider cc --lane lane-3 --run runY" "$p1" || fail "no exact command line for runY"
  has_line "moai factory relaunch --provider cc --lane lane-3 --run runZ" "$p1" || fail "no exact command line for runZ"
  [ -z "$(broker_peers runY)$(broker_peers runZ)" ] || fail "a peer was registered although two runs are active"
  case "$p2" in *factory\ lane*|*degraded*) fail "prompt 2 repeated the ambiguity notice";; esac
  echo "VERDICT: PASS"; exit 0;;
legacy-lines)
  seed runX active
  a="$(hook prompt worker-69 runX sa)"; show "legacy, run active" "$a"
  has_line "moai factory relaunch --provider cc --from-run runX" "$a" || fail "row R6: no exact line '--from-run runX'"
  seed runX retired
  c="$(hook prompt worker-69 runX sc)"; show "legacy, run retired, no active run" "$c"
  has_line "moai factory relaunch --provider cc" "$c" && fail "row R7: a command is printed with no active run"
  seed runY active
  b="$(hook prompt worker-69 runX sb)"; show "legacy, run retired, one active run" "$b"
  has_line "moai factory relaunch --provider cc" "$b" || fail "row R8: no exact line 'moai factory relaunch --provider cc'"
  case "$a$b" in *'<'*|*'>'*) fail "a notice still carries an angle-bracket placeholder";; esac
  [ -z "$(broker_peers runY)" ] || fail "a legacy session was registered into runY"
  echo "VERDICT: PASS"; exit 0;;
session-start-silent)
  seed runX retired; seed runY active
  s="$(hook start lane-3 runX s1)"
  show "SessionStart (source clear)" "$(printf '%s' "$s" | head -3)"
  case "$s" in *degraded*) fail "SessionStart still says 'factory messaging degraded'";; esac
  [ -z "$(broker_peers runX)$(broker_peers runY)" ] || fail "SessionStart registered a peer"
  [ -f "$FDIR/messages/runX/broker.db" ] || [ -f "$FDIR/messages/runY/broker.db" ] && fail "SessionStart created a broker"
  echo "VERDICT: PASS"; exit 0;;
roundtrip)
  seed runX retired; seed runY active
  b="$(hook prompt worker-69 runX sb)"
  line="$(printf '%s\n' "$b" | grep -m1 '^moai factory relaunch ')"
  show "printed line" "$line"
  [ -n "$line" ] || fail "the notice prints no 'moai factory relaunch' line to feed the verb"
  args="${line#moai }"
  # shellcheck disable=SC2086
  got="$("$BIN" $args --dry-run 2>&1)"; rc=$?
  show "verb --dry-run" "$got (exit $rc)"
  [ $rc -eq 0 ] || fail "the verb rejected its own printed line"
  [ "$got" = "moai cc -f lane" ] || fail "dry-run printed '$got', want 'moai cc -f lane'"
  echo "VERDICT: PASS"; exit 0;;
dry-run-nonmutation)
  seed runX active 999999 x
  before="$(sqlite3 "$DB" "SELECT run_id,status,lead_pid FROM runs; SELECT count(*) FROM events;")"
  got="$("$BIN" factory relaunch --dry-run --provider cc --from-run runX 2>&1)"; rc=$?
  after="$(sqlite3 "$DB" "SELECT run_id,status,lead_pid FROM runs; SELECT count(*) FROM events;")"
  show "verb --dry-run --from-run runX" "$got (exit $rc)"
  [ $rc -eq 0 ] || fail "the verb exited $rc"
  [ "$got" = "moai cc -f lane" ] || fail "dry-run printed '$got', want 'moai cc -f lane'"
  [ "$before" = "$after" ] || fail "--dry-run changed the runs/events tables: [$before] -> [$after]"
  echo "VERDICT: PASS"; exit 0;;
*)
  echo "unknown scenario: $SCEN" >&2; exit 2;;
esac
