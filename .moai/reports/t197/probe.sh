#!/bin/bash
# t197 measurement probe.
#
#   bash .moai/reports/t197/probe.sh
#
# Self-contained: it builds the binary it measures and generates the doctor
# JSON it reads, into its own temp dir. A failed prerequisite aborts the run
# instead of letting later steps report against a stale artifact.
#
# NOT read-only. Two steps mutate outside the repo:
#   - the build writes a binary into a temp dir (removed on exit)
#   - invoking `codex` makes it attempt to create PATH aliases under its own
#     home; on this machine that attempt fails and codex warns, but the attempt
#     is real. Everything else is a read.
#
# Repo working tree: untouched.

set -uo pipefail

WORK="$(mktemp -d)" || { echo "FATAL: mktemp -d failed" >&2; exit 1; }
[ -n "$WORK" ] && [ -d "$WORK" ] || { echo "FATAL: mktemp -d gave no usable dir" >&2; exit 1; }
trap 'rm -rf "$WORK"' EXIT
MOAI="$WORK/moai"
DOCTOR_JSON="$WORK/doctor.json"

FAILURES=0

fail() { echo "PREREQUISITE FAILED: $*" >&2; exit 1; }

# run <command>            — command must succeed (rc 0)
# run_rc <expected> <cmd>  — command must exit with <expected>
#
# Both print the command, its output, and its OWN rc. An unexpected rc is
# recorded and reported at the end, so a failed measurement can never hide
# behind a green transcript.
run_rc() {
  local expected="$1"; shift
  echo "\$ $1"
  eval "$1"
  local rc=$?
  echo "rc=$rc"
  if [ "$rc" != "$expected" ]; then
    echo "UNEXPECTED: rc=$rc, expected $expected" >&2
    echo "UNEXPECTED: rc=$rc, expected $expected"
    FAILURES=$((FAILURES + 1))
  fi
  echo
  return 0
}

run() { run_rc 0 "$1"; }

echo "### 0. environment + prerequisites"
run 'echo "bash=${BASH_VERSION}"'
run 'command -v codex'
run 'codex --version'

echo "# build the binary under measurement (into the probe temp dir)"
run 'go build -o "$MOAI" ./cmd/moai'
[ -x "$MOAI" ] || fail "go build produced no binary at $MOAI"

echo "# generate the doctor JSON this probe reads (timed, rc captured)"
run 'START=$(date +%s); codex doctor --json > "$DOCTOR_JSON" 2>/dev/null; DRC=$?; END=$(date +%s); echo "codex doctor --json: rc=$DRC elapsed=$((END-START))s"; ( exit $DRC )'
[ -s "$DOCTOR_JSON" ] || fail "codex doctor --json produced no output at $DOCTOR_JSON"

echo "### baseline"
run 'git rev-parse --short HEAD'
run 'git status --short'
run 'git merge-base --is-ancestor 7b217da7c HEAD'

echo "### M-1 no moai codex"
run_rc 1 '"$MOAI" --help 2>&1 | grep -ci "codex"; ( exit ${PIPESTATUS[1]} )'   # rc 1 = no match, which IS the finding
run '"$MOAI" --help 2>&1 | sed -n "/LAUNCH COMMANDS/,/PROJECT COMMANDS/p" | sed "s/ *$//"; ( exit ${PIPESTATUS[0]} )'
run 'grep -rn "Use: *\"codex" internal/cli/*.go'

echo "### M-2 auth stream"
run 'codex login status'
run 'codex login status 2>/dev/null | wc -c; ( exit ${PIPESTATUS[0]} )'
run 'codex login status 2>&1 1>/dev/null | wc -c; ( exit ${PIPESTATUS[0]} )'
run 'sed -n "273,284p" internal/cli/mcp_codex.go'
run 'sed -n "1326,1343p" internal/cli/mcp_codex.go'
run 'grep -rn "AuthProvider" internal/web/*.go internal/cli/web.go | grep -v _test; ( exit ${PIPESTATUS[0]} )'

echo "### M-2b structured auth source"
run 'codex doctor 2>&1 | grep -iE "auth|login|account"; ( exit ${PIPESTATUS[1]} )'
run 'python3 -c "import json,sys;d=json.load(open(sys.argv[1]));print(json.dumps(d[\"checks\"][\"auth.credentials\"],indent=1))" "$DOCTOR_JSON"'
run 'python3 -c "import json,sys;d=json.load(open(sys.argv[1]));print(chr(10).join(k+\": \"+v[\"summary\"] for k,v in d[\"checks\"].items() if \"rollout\" in k or \"thread\" in k))" "$DOCTOR_JSON"'
run 'CODEX_HOME="${CODEX_HOME:-$HOME/.codex}" python3 .moai/reports/t197/authshape.py'

echo "### M-3 codex CLI surface (full Commands block + count)"
run 'codex --help 2>&1 | sed -n "/^Commands:/,/^Arguments:/p" | sed "s/ *$//"; ( exit ${PIPESTATUS[0]} )'
run 'codex --help 2>&1 | sed -n "/^Commands:/,/^Arguments:/p" | grep -cE "^  [a-z]"; ( exit ${PIPESTATUS[1]} )'

echo "### M-4 wiring surface"
run 'grep -n "HooksRelPath =\|ConfigRelPath =" internal/codexwiring/codexwiring.go'
run 'find internal/template/templates/.codex/agents/moai -name "*.toml" -type f | sort; ( exit ${PIPESTATUS[0]} )'
run 'find internal/template/templates/.codex/agents/moai -name "*.toml" -type f | wc -l; ( exit ${PIPESTATUS[0]} )'
run 'grep -n "must be one of: claude, codex, both" internal/cli/init.go'
run 'grep -n "func checkCodexWiring" internal/cli/doctor_codex.go'
run 'grep -n "codexHarnessFlagValue = " internal/cli/hook_harness_codex.go'
run_rc 1 'ls -a .codex'   # rc 1 = no wiring in this repo, which IS the finding

echo "### M-5 CODEX_HOME readers"
run_rc 1 'grep -rn "CODEX_HOME" internal/ --include="*.go"'   # rc 1 = zero matches, which IS the finding

echo "### M-6 launcher convergence"
run 'grep -n "func unifiedLaunch\|func unifiedLaunchDefault" internal/cli/launcher.go'
run 'grep -n "return unifiedLaunch(" internal/cli/cc.go internal/cli/cg.go internal/cli/glm.go'
run 'grep -n "func spawnLaunch" internal/cli/spawn.go'

echo "### M-7 runner implementations"
run 'grep -rn ") run(" internal/cli/*.go internal/cli/*_test.go | grep -v glm | sort -u; ( exit ${PIPESTATUS[0]} )'

echo "### probe result"
if [ "$FAILURES" -ne 0 ]; then
  echo "PROBE FAILED: $FAILURES command(s) exited unexpectedly — this transcript is NOT a valid baseline"
  exit 1
fi
echo "PROBE OK: every command exited as expected"
