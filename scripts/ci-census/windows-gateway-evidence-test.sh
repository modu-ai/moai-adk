#!/usr/bin/env bash
set -euo pipefail
base="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fixture_dir="$(mktemp -d)"
trap 'rm -rf "$fixture_dir"' EXIT
# Synthetic event streams only; no Go/Windows/client invocation.
cat > "$fixture_dir/expected.json" <<'JSON'
[{"Package":"fixture/auth","Test":"TestACL"},{"Package":"fixture/gateway","Test":"TestLifecycle"}]
JSON
jq -cn --slurpfile wanted "$fixture_dir/expected.json" '$wanted[0][] | {Action:"run",Package,Test},{Action:"pass",Package,Test},{Action:"pass",Package}' > "$fixture_dir/good.json"
bash "$base/windows-gateway-evidence.sh" "$fixture_dir/good.json" "$fixture_dir/expected.json"
for mutation in missing skipped failed pass_without_run truncated malformed; do
 case "$mutation" in
 missing) jq -c 'select(.Test != "TestACL")' "$fixture_dir/good.json" > "$fixture_dir/bad.json";;
 skipped) jq -c 'if .Test=="TestACL" and .Action=="pass" then .Action="skip" else . end' "$fixture_dir/good.json" > "$fixture_dir/bad.json";;
 failed) jq -c 'if .Package=="fixture/auth" and .Test==null then .Action="fail" else . end' "$fixture_dir/good.json" > "$fixture_dir/bad.json";;
 pass_without_run) jq -c 'select(.Action != "run")' "$fixture_dir/good.json" > "$fixture_dir/bad.json";;
 truncated) jq -c 'select(.Test != null)' "$fixture_dir/good.json" > "$fixture_dir/bad.json";;
 malformed) printf '{broken\n' > "$fixture_dir/bad.json";;
 esac
 if bash "$base/windows-gateway-evidence.sh" "$fixture_dir/bad.json" "$fixture_dir/expected.json" > "$fixture_dir/result" 2>&1; then
  echo "FAIL accepted $mutation";exit 1
 fi
 echo "REJECTED $mutation"
done
if bash "$base/windows-gateway-evidence.sh" "$fixture_dir/absent" "$fixture_dir/expected.json" > "$fixture_dir/result" 2>&1; then exit 1;fi
echo 'REJECTED absent_stream'
