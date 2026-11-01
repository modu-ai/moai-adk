#!/usr/bin/env bash
set -euo pipefail

hook_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
root=$(mktemp -d)
# The fake go's call journal lives outside the repository so writing it never
# changes the dirty state the gate keys its record on.
journal_dir=$(mktemp -d)
journal="$journal_dir/go-calls.log"
trap 'rm -rf "$root" "$journal_dir"' EXIT
mkdir -p "$root/bin"

# The gate calls `go -C <module root> vet ./...` (Go 1.20+ form), so the fake
# strips a leading -C <dir> before it decides which step it is answering, and
# journals every call so the test can prove the intended vet failure happened.
cat > "$root/bin/go" <<EOF
#!/usr/bin/env bash
printf '%s\n' "\$*" >> "$journal"
if [ "\${1:-}" = "-C" ]; then shift 2; fi
case "\${1:-} \${2:-}" in
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

# The gate's blocking/advisory mode reads MOAI_SYNC_GATE_BLOCKING and
# MOAI_AUTONOMY_TIER (fully-autonomous resolves to advisory). Both are removed
# so the result never depends on the session that happens to run this test.
run_gate() {
  (cd "$root" && env -u MOAI_AUTONOMY_TIER -u MOAI_SYNC_GATE_BLOCKING \
    PATH="$root/bin:$PATH" CLAUDE_PROJECT_DIR="$root" \
    bash "$hook_dir/moai/sync-phase-quality-gate.sh" <<< '{}')
}

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

# `|| true`: a gate that exits non-zero must reach the assertions below, which
# say what went wrong, instead of ending the script silently under set -e.
first=$(run_gate || true)

# Premise: the fixture itself made the intended `go vet ./...` call in the
# module root; without it a missing block would mean the fake was never hit.
[[ -f "$journal" ]] || fail "premise: the gate never called go at all"
grep -Eq -- '^-C .+ vet \./\.\.\.$' "$journal" ||
  fail "premise: the gate never ran go vet ./... through the fake go; journal: $(tr '\n' ';' < "$journal")"

grep -q '"decision":"block"' <<< "$first" || fail "first run did not block; output: $first"
second=$(run_gate || true)
grep -q '"decision":"block"' <<< "$second" || fail "second run did not block; output: $second"

# The record is "<HEAD> fail <WORKTREE_ID>" (three fields).
head=$(git -C "$root" rev-parse HEAD)
record=$(cat "$root/.moai/state/sync-quality-gate.last" 2>/dev/null || echo "<missing>")
[[ "$record" == "$head fail "* ]] || fail "state record is '$record', want '$head fail <worktree id>'"
[[ "$first" == "$second" ]] || fail "the re-delivered payload differs from the first"

printf 'PASS: failed gate is re-delivered for the same HEAD\n'
