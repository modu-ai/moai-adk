#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
docs="$root/.claude/skills/moai/workflows/sync/delivery.md"
if grep -q 'which golangci-lint &&' "$docs"; then
  printf 'FAIL: short-circuit lint command still swallows failures\n' >&2
  exit 1
fi

bin=$(mktemp -d)
trap 'rm -rf "$bin"' EXIT
cat > "$bin/golangci-lint" <<'EOF'
#!/usr/bin/env bash
exit 7
EOF
chmod +x "$bin/golangci-lint"

run_lint() {
  if ! command -v golangci-lint >/dev/null 2>&1; then
    echo SKIP
  elif golangci-lint run --timeout=5m; then
    echo PASS
  else
    status=$?
    echo "FAIL:$status" >&2
    return "$status"
  fi
}

set +e
PATH="$bin:$PATH" run_lint >/tmp/moai-lint-contract.out 2>/tmp/moai-lint-contract.err
status=$?
set -e
[[ "$status" -eq 7 ]]
grep -q 'FAIL:7' /tmp/moai-lint-contract.err
rm -f /tmp/moai-lint-contract.out /tmp/moai-lint-contract.err
printf 'PASS: installed lint failure preserves exit status 7\n'
