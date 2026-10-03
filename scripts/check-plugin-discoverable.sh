#!/bin/sh
# Discoverability check for the moai plugin (SPEC-PLUGIN-MARKETPLACE-001, AC-006 (b) to (e)).
#
# Usage: check-plugin-discoverable.sh <empty-config-home>
#
# Installs the marketplace of this repository from its local path into an empty scratch Claude config home
# and checks that the runtime's own component inventory lists every skill and command the template tree
# defines. The inventory is the only view that counts a command (a nested commands/<dir>/ layout installs
# fine and is listed as nothing, P-30), so this is the one script that runs the real Claude CLI. It starts
# exactly three of its verbs, each once, and nothing else: marketplace add <repository root>, install
# moai@moai-adk, details moai@moai-adk. It starts no other tool.
#
# Contract (REQ-025 carve-out):
#   - the argument must be an existing EMPTY directory, else it prints "refused: ..." and exits 2 before
#     anything starts, which keeps the script away from a real profile;
#   - every MOAI_*, CLAUDE_* and CODEX_* name in the caller's environment is unset by live enumeration
#     (between the scrub markers below), and the one name set afterwards is the config home. The scrub is
#     not decoration: the plugin cache variable moves the whole plugin tree out of the config home (P-49),
#     and a pin variable would redirect which binary the product resolves;
#   - the working directory is a scratch directory, and the marketplace source is a local path;
#   - the expected names are the directory names under plugins/moai/skills plus the command stems of the
#     TEMPLATE tree, never a literal, so a payload that nests or omits its commands cannot define its own
#     expectation.
#
# Exit status: 0 all expected names listed; 1 a name is missing, the counts or the MCP servers differ, the
# inventory cannot be parsed or a verb failed; 2 refused or unusable input.
# Output: "scrub: enumerated and unset <n> names", then on success a last line "ok: <N> names listed, 0 missing".
set -eu

fail() {
    printf '%s\n' "$*" >&2
    exit 1
}

[ $# -eq 1 ] || { echo "refused: expected exactly one argument, an existing empty directory" >&2; exit 2; }
if [ ! -d "$1" ] || [ -n "$(ls -A "$1")" ]; then
    echo "refused: $1 is not an existing empty directory" >&2
    exit 2
fi
home=$(cd "$1" && pwd -P)
repo=$(cd "$(dirname "$0")/.." && pwd -P)

# Local source: an absolute path to a directory that carries the marketplace manifest.
case $repo in /*) ;; *) echo "refused: the repository root $repo is not an absolute path" >&2; exit 2 ;; esac
for f in .claude-plugin/marketplace.json plugins/moai/.mcp.json; do
    [ -f "$repo/$f" ] || { echo "refused: $repo has no $f (run make plugin-emit)" >&2; exit 2; }
done

# scrub:begin
scrubbed=0
for name in $(awk 'BEGIN { for (n in ENVIRON) if (n ~ /^(MOAI|CLAUDE|CODEX)_/) print n }'); do
    unset "$name"
    scrubbed=$((scrubbed + 1))
done
echo "scrub: enumerated and unset $scrubbed names"
left=$(awk 'BEGIN { for (n in ENVIRON) if (n ~ /^(MOAI|CLAUDE|CODEX)_/) print n }')
[ -z "$left" ] || { echo "ABORT: the scrub left names behind: $left" >&2; exit 2; }
# scrub:end

command -v claude >/dev/null 2>&1 || { echo "refused: the Claude CLI is not on PATH" >&2; exit 2; }

work=$(mktemp -d)
trap 'cd /; rm -rf "$work"' EXIT
cd "$work"
CLAUDE_CONFIG_DIR=$home
export CLAUDE_CONFIG_DIR

# Expected names, from the tree and the template only.
{
    for d in "$repo"/plugins/moai/skills/*/; do
        [ -d "$d" ] || continue
        d=${d%/}
        echo "${d##*/}"
    done
    for f in "$repo"/internal/template/templates/.claude/commands/moai/*; do
        [ -f "$f" ] || continue
        b=${f##*/}
        b=${b%.tmpl}
        echo "${b%.md}"
    done
} | LC_ALL=C sort > "$work/expected"
[ -s "$work/expected" ] || { echo "refused: no expected names found under $repo" >&2; exit 2; }
expected_n=$(wc -l < "$work/expected" | tr -d ' ')

# MCP servers the payload declares (the generated file indents each server key by four spaces).
awk '/^    "[^"]+": *\{/ { sub(/^    "/, ""); sub(/".*$/, ""); print }' "$repo/plugins/moai/.mcp.json" | LC_ALL=C sort > "$work/mcp-expected"

# The three verbs.
add_out=$(claude plugin marketplace add "$repo" --json 2>&1) || fail "verb failed (marketplace add): $add_out"
case $add_out in *'"outcome":"ok"'*) ;; *) fail "marketplace add did not report ok: $add_out" ;; esac
install_out=$(claude plugin install moai@moai-adk --json 2>&1) || fail "verb failed (install): $install_out"
case $install_out in *'"outcome":"ok"'*) ;; *) fail "install did not report ok: $install_out" ;; esac
details=$(claude plugin details moai@moai-adk 2>&1) || fail "verb failed (details): $details"

# Parse the inventory text (it has no JSON form): "  Skills (N)  a, b, ..." and "  MCP servers (K)  x  (...)".
skills_line=$(printf '%s\n' "$details" | sed -n 's/^  Skills (\([0-9][0-9]*\))  \(.*\)$/\1 \2/p')
[ -n "$skills_line" ] || fail "cannot parse the inventory: no 'Skills (N)' line in: $details"
count=${skills_line%% *}
printf '%s\n' "${skills_line#* }" | tr ',' '\n' | sed 's/^ *//; s/ *$//' | LC_ALL=C sort > "$work/listed"

status=0
for name in $(LC_ALL=C comm -23 "$work/expected" "$work/listed"); do
    echo "missing: $name"
    status=1
done
if [ "$count" != "$expected_n" ]; then
    echo "mismatch: the inventory reports Skills ($count) but $expected_n names are expected"
    status=1
fi

mcp_line=$(printf '%s\n' "$details" | sed -n 's/^  MCP servers ([0-9][0-9]*)  \(.*\)$/\1/p')
if [ -n "$mcp_line" ]; then
    mcp_line=${mcp_line%%  (*}
else
    mcp_line=
    [ ! -s "$work/mcp-expected" ] || fail "cannot parse the inventory: no 'MCP servers (K)' line in: $details"
fi
printf '%s\n' "$mcp_line" | tr ',' '\n' | sed 's/^ *//; s/ *$//; /^$/d' | LC_ALL=C sort > "$work/mcp-listed"
if ! cmp -s "$work/mcp-expected" "$work/mcp-listed"; then
    echo "mismatch: the inventory lists MCP servers [$(tr '\n' ' ' < "$work/mcp-listed")] but .mcp.json declares [$(tr '\n' ' ' < "$work/mcp-expected")]"
    status=1
fi

[ "$status" -eq 0 ] || exit 1
echo "ok: $count names listed, 0 missing"
