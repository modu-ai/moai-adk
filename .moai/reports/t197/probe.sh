#!/bin/bash
# t197 measurement probe. Run with: bash .moai/reports/t197/probe.sh
# Read-only. Every step prints its command, then its output, then its rc.
set +e

run() {
  echo "\$ $1"
  eval "$1"
  echo "rc=$?"
  echo
}

echo "### shell"
run 'echo "bash=${BASH_VERSION}"'

echo "### baseline"
run 'git rev-parse --short HEAD'
run 'git status --short'
run 'git merge-base --is-ancestor 7b217da7c HEAD'

echo "### M-1 no moai codex"
run '/tmp/moai-t197 --help 2>&1 | grep -ci "codex"; ( exit ${PIPESTATUS[1]} )'
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
run 'python3 -c "import json;d=json.load(open(\"/tmp/t197-doctor.json\"));print(json.dumps(d[\"checks\"][\"auth.credentials\"],indent=1))"'
run 'python3 .moai/reports/t197/authshape.py'

echo "### M-4 wiring surface"
run 'grep -n "HooksRelPath =\|ConfigRelPath =" internal/codexwiring/codexwiring.go'
run 'find internal/template/templates/.codex/agents/moai -name "*.toml" -type f | sort; ( exit ${PIPESTATUS[0]} )'
run 'find internal/template/templates/.codex/agents/moai -name "*.toml" -type f | wc -l; ( exit ${PIPESTATUS[0]} )'
run 'grep -n "must be one of: claude, codex, both" internal/cli/init.go'
run 'grep -n "func checkCodexWiring" internal/cli/doctor_codex.go'
run 'grep -n "codexHarnessFlagValue = " internal/cli/hook_harness_codex.go'
run 'ls -a .codex'

echo "### M-5 CODEX_HOME readers"
run 'grep -rn "CODEX_HOME" internal/ --include="*.go"'

echo "### M-6 launcher convergence"
run 'grep -n "func unifiedLaunch\|func unifiedLaunchDefault" internal/cli/launcher.go'
run 'grep -n "return unifiedLaunch(" internal/cli/cc.go internal/cli/cg.go internal/cli/glm.go'
run 'grep -n "func spawnLaunch" internal/cli/spawn.go'

echo "### M-7 runner implementations"
run 'grep -rn ") run(" internal/cli/*.go internal/cli/*_test.go | grep -v glm; ( exit ${PIPESTATUS[0]} )'
