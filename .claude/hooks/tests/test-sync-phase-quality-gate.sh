#!/usr/bin/env bash
set -euo pipefail

hook_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
root=$(mktemp -d)
trap 'rm -rf "$root"' EXIT
mkdir -p "$root/bin"

cat > "$root/bin/go" <<'EOF'
#!/usr/bin/env bash
case "$1 $2" in
  "vet ./...") exit 7 ;;
  "build ./...") exit 0 ;;
  *) exit 0 ;;
esac
EOF
chmod +x "$root/bin/go"

git -C "$root" init -q
git -C "$root" config user.email test@example.com
git -C "$root" config user.name test
printf 'module example.test\n\ngo 1.23\n' > "$root/go.mod"
printf 'package main\n' > "$root/main.go"
git -C "$root" add go.mod main.go
git -C "$root" commit -qm 'docs(t624): sync-phase quality-gate fixture'

run_gate() {
  (cd "$root" && PATH="$root/bin:$PATH" CLAUDE_PROJECT_DIR="$root" \
    bash "$hook_dir/moai/sync-phase-quality-gate.sh" <<< '{}')
}

first=$(run_gate)
grep -q '"decision":"block"' <<< "$first"
second=$(run_gate)
grep -q '"decision":"block"' <<< "$second"

head=$(git -C "$root" rev-parse HEAD)
[[ "$(cat "$root/.moai/state/sync-quality-gate.last")" == "$head fail" ]]
[[ "$first" == "$second" ]]

printf 'PASS: failed gate is re-delivered for the same HEAD\n'
