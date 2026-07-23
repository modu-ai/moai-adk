#!/usr/bin/env bash
# TUX v3 CLI E2E journeys (J1-J6) for moai-adk-go.
#
# Self-contained, re-runnable, macOS bash-3.2 compatible.
# - Builds a fresh binary from the repo (never uses the installed ~/go/bin/moai).
# - All mutating journeys run inside a throwaway sandbox under $SANDBOX
#   (default /tmp/moai-e2e). The dev checkout is never mutated.
# - Full command output is redirected to $RUN_DIR (bounded-output discipline);
#   the terminal only sees per-journey PASS/FAIL lines + a final matrix.
#
# Env overrides:
#   MOAI_E2E_SANDBOX  sandbox root        (default /tmp/moai-e2e)
#   MOAI_E2E_BIN      prebuilt binary     (default $SANDBOX/bin/moai, built if absent)
#   MOAI_E2E_RUN_DIR  log directory       (default <repo>/e2e/.runs/tux3-<YYYYMMDD>)
#   MOAI_E2E_KEEP=1   keep sandbox proj dirs after the run
#
# Exit code: 0 when all journeys PASS, 1 otherwise.

set -u

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SANDBOX="${MOAI_E2E_SANDBOX:-/tmp/moai-e2e}"
BIN="${MOAI_E2E_BIN:-$SANDBOX/bin/moai}"
RUN_DIR="${MOAI_E2E_RUN_DIR:-$ROOT/e2e/.runs/tux3-$(date +%Y%m%d)}"
KEEP="${MOAI_E2E_KEEP:-0}"

mkdir -p "$SANDBOX" "$RUN_DIR"
RESULTS="$RUN_DIR/results.tsv"
MATRIX="$RUN_DIR/matrix.tsv"
ASSERTS="$RUN_DIR/asserts.log"
: > "$RESULTS"; : > "$MATRIX"; : > "$ASSERTS"

# ---------------------------------------------------------------- helpers ----

log() { printf '%s\n' "$*"; }

# esc_count <file> -> number of lines containing ESC (0x1b)
esc_count() { LC_ALL=C grep -c "$(printf '\033')" "$1" 2>/dev/null || true; }

# run_to <log-name> <timeout-s> <workdir> <command string...>
# Runs the command string via bash -c with stdin </dev/null (non-TTY),
# stdout+stderr -> $RUN_DIR/<log-name>.log. Sets $RC (137/143 => timeout-kill).
RC=0
run_to() {
  local name="$1" tmo="$2" dir="$3"; shift 3
  local logf="$RUN_DIR/$name.log"
  ( cd "$dir" && bash -c "$*" ) </dev/null >"$logf" 2>&1 &
  local pid=$!
  ( sleep "$tmo"; kill -9 "$pid" 2>/dev/null ) &
  local wpid=$!
  wait "$pid" 2>/dev/null; RC=$?
  kill "$wpid" 2>/dev/null; wait "$wpid" 2>/dev/null
  return 0
}

# run_split <log-name> <timeout-s> <workdir> <command string...>
# Like run_to but keeps stdout/stderr separate (<name>.out / <name>.err).
run_split() {
  local name="$1" tmo="$2" dir="$3"; shift 3
  ( cd "$dir" && bash -c "$*" ) </dev/null >"$RUN_DIR/$name.out" 2>"$RUN_DIR/$name.err" &
  local pid=$!
  ( sleep "$tmo"; kill -9 "$pid" 2>/dev/null ) &
  local wpid=$!
  wait "$pid" 2>/dev/null; RC=$?
  kill "$wpid" 2>/dev/null; wait "$wpid" 2>/dev/null
  return 0
}

# Per-journey assertion accumulator.
J_FAILS=""
j_begin() { J_FAILS=""; }
ok()   { printf 'PASS  %s\n' "$1" >> "$ASSERTS"; }
bad()  { printf 'FAIL  %s\n' "$1" >> "$ASSERTS"; J_FAILS="${J_FAILS}${1}; "; }
# check <desc> <cond-exit-status>  (call as: [ ... ]; check "desc" $?)
check() { if [ "$2" -eq 0 ]; then ok "$1"; else bad "$1"; fi; }

j_end() { # j_end <journey> <evidence>
  local verdict="PASS"
  [ -n "$J_FAILS" ] && verdict="FAIL"
  printf '%s\t%s\t%s\t%s\n' "$1" "$verdict" "$2" "${J_FAILS:--}" >> "$RESULTS"
  log "[$1] $verdict ${J_FAILS:+-- $J_FAILS}"
}

# ------------------------------------------------------------------ build ----

if [ ! -x "$BIN" ]; then
  log "[setup] building fresh binary -> $BIN"
  mkdir -p "$(dirname "$BIN")"
  ( cd "$ROOT" && go build -o "$BIN" ./cmd/moai ) > "$RUN_DIR/setup-build.log" 2>&1
  if [ $? -ne 0 ]; then
    log "[setup] FAIL: go build failed, see $RUN_DIR/setup-build.log"
    exit 1
  fi
fi
"$BIN" --version > "$RUN_DIR/setup-version.log" 2>&1
log "[setup] binary: $BIN ($(head -1 "$RUN_DIR/setup-version.log"))"

P1="$SANDBOX/proj-j1"
rm -rf "$P1" "$SANDBOX/proj-j1b"

# ------------------------------------------------------------ J1: moai init ----

j_begin
run_to j1-init 180 "$SANDBOX" "NO_COLOR=1 '$BIN' init proj-j1 --non-interactive --language go --git-mode manual"
[ "$RC" -eq 0 ]; check "J1 init exit=0 (got $RC)" $?
[ -d "$P1/.moai" ];   check "J1 .moai scaffold present" $?
[ -d "$P1/.claude" ]; check "J1 .claude scaffold present" $?
[ "$(wc -l < "$RUN_DIR/j1-init.log" | tr -d ' ')" -ge 5 ]; check "J1 live-progress output present (>=5 lines)" $?
[ "$(esc_count "$RUN_DIR/j1-init.log")" -eq 0 ]; check "J1 NO_COLOR output ANSI-free" $?
j_end J1 "$RUN_DIR/j1-init.log"

# J1b (informational sub-journey): bare non-TTY init WITHOUT --non-interactive —
# TUX v3 plain-fallback guarantee. Timeout => hang => FAIL.
j_begin
run_to j1b-init-notty 120 "$SANDBOX" "NO_COLOR=1 '$BIN' init proj-j1b"
if [ "$RC" -eq 137 ] || [ "$RC" -eq 143 ]; then
  bad "J1b bare non-TTY init hung (killed by timeout)"
else
  [ "$RC" -eq 0 ]; check "J1b bare non-TTY init exit=0 (got $RC)" $?
  [ -d "$SANDBOX/proj-j1b/.moai" ]; check "J1b scaffold present" $?
  [ "$(esc_count "$RUN_DIR/j1b-init-notty.log")" -eq 0 ]; check "J1b ANSI-free" $?
fi
j_end J1b "$RUN_DIR/j1b-init-notty.log"

# ---------------------------------------------- J2: moai update change-preview ----

j_begin
if [ -d "$P1/.moai" ]; then
  # Hash snapshot (exclude runtime-state dirs: logs/state/cache/backups).
  snap() {
    ( cd "$P1" && find .moai .claude -type f \
        ! -path '.moai/logs/*' ! -path '.moai/state/*' ! -path '.moai/cache/*' \
        ! -path '.moai/backups/*' ! -path '.moai/archive/*' \
        -exec shasum -a 256 {} + | LC_ALL=C sort )
  }
  snap > "$RUN_DIR/j2-hash-before.txt"
  run_to j2-update-dryrun 180 "$P1" "NO_COLOR=1 '$BIN' update --templates-only --dry-run --yes"
  snap > "$RUN_DIR/j2-hash-after.txt"

  [ "$RC" -eq 0 ]; check "J2 update --dry-run exit=0 (got $RC)" $?
  grep -Eiq 'dry|preview|plan|would|sync|skip|template' "$RUN_DIR/j2-update-dryrun.log"
  check "J2 preview/diff summary rendered" $?
  diff "$RUN_DIR/j2-hash-before.txt" "$RUN_DIR/j2-hash-after.txt" > "$RUN_DIR/j2-hash-diff.txt"
  check "J2 dry-run made no destructive changes (hash-identical tree)" $?
  [ "$(esc_count "$RUN_DIR/j2-update-dryrun.log")" -eq 0 ]; check "J2 NO_COLOR output ANSI-free" $?
else
  bad "J2 skipped: J1 sandbox project missing"
fi
j_end J2 "$RUN_DIR/j2-update-dryrun.log"

# -------------------------------------------------------- J3: moai doctor ----

j_begin
run_to j3-doctor-default 120 "$P1" "'$BIN' doctor"
RC_A=$RC
run_to j3-doctor-nocolor 120 "$P1" "NO_COLOR=1 '$BIN' doctor"
RC_B=$RC
[ "$RC_A" -eq "$RC_B" ]; check "J3 exit-code semantics consistent (default=$RC_A nocolor=$RC_B)" $?
grep -Eiq 'check|status|environment|config|git|go' "$RUN_DIR/j3-doctor-nocolor.log"
check "J3 section result tables present" $?
[ "$(esc_count "$RUN_DIR/j3-doctor-nocolor.log")" -eq 0 ]; check "J3 NO_COLOR+pipe ANSI-free (ESC count 0)" $?
j_end J3 "$RUN_DIR/j3-doctor-nocolor.log"

# ------------------------------------- J4: moai status + moai spec view ----

j_begin
run_to j4-status 60 "$P1" "NO_COLOR=1 '$BIN' status"
[ "$RC" -eq 0 ]; check "J4 status exit=0 (got $RC)" $?
[ "$(esc_count "$RUN_DIR/j4-status.log")" -eq 0 ]; check "J4 status ANSI-free" $?
grep -Eiq 'proj-j1|project' "$RUN_DIR/j4-status.log"; check "J4 status shows project name" $?
# Version renders as "ADK: moai-adk v3.0.0" (semver token, not the word "version").
grep -Eiq 'version|v[0-9]+\.[0-9]+\.[0-9]+' "$RUN_DIR/j4-status.log"; check "J4 status shows version" $?
grep -Eiq 'config|mode'     "$RUN_DIR/j4-status.log"; check "J4 status shows config summary" $?
! grep -q '```' "$RUN_DIR/j4-status.log"; check "J4 status: no markdown fences leaking" $?

# Fixture SPEC for spec view (AC-ID grammar per internal/spec parser).
FIX="$P1/.moai/specs/SPEC-E2E-DEMO-001"
mkdir -p "$FIX"
cat > "$FIX/spec.md" <<'EOF'
---
id: SPEC-E2E-DEMO-001
title: E2E demo fixture
status: draft
---

# SPEC-E2E-DEMO-001: E2E demo fixture

## Requirements

- REQ-E2E-001: The system SHALL render acceptance criteria as a tree.

## Acceptance Criteria

- AC-E2E-001-01: Given a sandbox project, When spec view runs non-TTY, Then plain tree output is produced (maps REQ-E2E-001)
  - AC-E2E-001-01.a: Given the same, When NO_COLOR=1, Then output contains zero ESC bytes
- AC-E2E-001-02: Given a sandbox project, When spec view runs, Then the exit code is 0 (maps REQ-E2E-001)
EOF

run_to j4-specview 60 "$P1" "NO_COLOR=1 '$BIN' spec view SPEC-E2E-DEMO-001"
[ "$RC" -eq 0 ]; check "J4 spec view exit=0 (got $RC)" $?
[ "$(esc_count "$RUN_DIR/j4-specview.log")" -eq 0 ]; check "J4 spec view ANSI-free (plain passthrough)" $?
grep -q 'AC-E2E-001-01' "$RUN_DIR/j4-specview.log"; check "J4 spec view renders AC tree nodes" $?
grep -Eq '└──|├──' "$RUN_DIR/j4-specview.log"; check "J4 spec view tree glyphs present" $?
! grep -q '```' "$RUN_DIR/j4-specview.log"; check "J4 spec view: no fences leaking" $?
j_end J4 "$RUN_DIR/j4-specview.log"

# ------------------------------------------------- J5: moai --help banner ----

j_begin
run_split j5-help 60 "$P1" "NO_COLOR=1 '$BIN' --help"
[ "$RC" -eq 0 ]; check "J5 --help exit=0 (got $RC)" $?
[ "$(esc_count "$RUN_DIR/j5-help.out")" -eq 0 ]; check "J5 help ANSI-free under NO_COLOR" $?
[ ! -s "$RUN_DIR/j5-help.err" ]; check "J5 stderr empty (stdout/stderr separation)" $?
grep -q 'USAGE'            "$RUN_DIR/j5-help.out"; check "J5 USAGE header present" $?
grep -q 'COMMANDS'         "$RUN_DIR/j5-help.out"; check "J5 COMMANDS group header present" $?
grep -q 'PROJECT COMMANDS' "$RUN_DIR/j5-help.out"; check "J5 PROJECT COMMANDS group header present" $?
grep -q 'TOOLS'            "$RUN_DIR/j5-help.out"; check "J5 TOOLS group header present" $?
! grep -Eq '█|�block|_____|\\\\ /|ASCII' "$RUN_DIR/j5-help.out"; check "J5 no large ASCII-art logo" $?
BANNER_LINES=$(awk '/USAGE/{exit} {n++} END{print n+0}' "$RUN_DIR/j5-help.out")
[ "$BANNER_LINES" -le 12 ]; check "J5 compact banner (<=12 lines before USAGE, got $BANNER_LINES)" $?
! grep -Eiq 'warning|deprecat' "$RUN_DIR/j5-help.out"; check "J5 stdout free of warnings" $?
j_end J5 "$RUN_DIR/j5-help.out"

# --------------------------------------- J6: regression matrix + GOOS build ----

j_begin
printf 'surface\tcombo\texit\tesc_lines\n' >> "$MATRIX"
for surface in doctor status help; do
  case "$surface" in
    help) CMDSTR="'$BIN' --help" ;;
    *)    CMDSTR="'$BIN' $surface" ;;
  esac
  EXITS=""
  for combo in nocolor plain; do
    name="j6-$surface-$combo"
    if [ "$combo" = "nocolor" ]; then
      run_to "$name" 120 "$P1" "NO_COLOR=1 $CMDSTR"
    else
      run_to "$name" 120 "$P1" "$CMDSTR"
    fi
    E=$(esc_count "$RUN_DIR/$name.log")
    printf '%s\t%s\t%s\t%s\n' "$surface" "$combo" "$RC" "$E" >> "$MATRIX"
    EXITS="$EXITS $RC"
    [ "$E" -eq 0 ]; check "J6 $surface/$combo piped non-TTY ANSI-free (esc=$E)" $?
  done
  set -- $EXITS
  [ "$1" = "$2" ]; check "J6 $surface exit consistent across combos ($1 vs $2)" $?
done

run_to j6-goos-windows 300 "$ROOT" "GOOS=windows GOARCH=amd64 go build ./..."
[ "$RC" -eq 0 ]; check "J6 GOOS=windows GOARCH=amd64 build (got $RC)" $?
j_end J6 "$RUN_DIR/matrix.tsv"

# --------------------------------------------------------------- teardown ----

if [ "$KEEP" != "1" ]; then
  rm -rf "$SANDBOX/proj-j1" "$SANDBOX/proj-j1b"
  log "[teardown] sandbox project dirs removed (binary kept at $BIN)"
fi

log ""
log "==== journey results ($RESULTS) ===="
cat "$RESULTS"
log "==== J6 matrix ($MATRIX) ===="
cat "$MATRIX"

grep -q "$(printf '\t')FAIL" "$RESULTS" && exit 1
exit 0
