#!/usr/bin/env bash
# race-scope.sh — which advisory Race Test shards does this pull request need?
#
# Prints exactly one stdout line, `shards=,<token>,<token>,…,` (leading and
# trailing comma, so a workflow can test `contains(shards, ',cli-1,')`), drawn
# from the six shard tokens of .github/workflows/ci.yml:
#   cli-1 cli-2 cli-3 cli-4   go test -race ./internal/cli      (4 name slices)
#   rosterguard               go test -race ./internal/harness/rosterguard
#   remainder                 every other package that has tests
#
# Usage (from the module root):  race-scope.sh --changed <file with one path per line>
#
# The script narrows ONLY when it can prove the narrowing, and otherwise closes
# to all six. Narrowing needs every changed file to be a Go file in a package of
# this module; a shard is selected when a changed package is among the
# dependencies (test imports included) of the packages that shard tests.
# Everything else — a non-Go file, go.mod/go.sum, a workflow or this script, a
# testdata or vendor path, a main package (tests build and run the command
# binary), a directory that is no longer a package, an empty or unreadable
# list, a failing `go list` — prints all six. It always exits 0: the answer,
# never a failure, is how it signals "cannot judge", and each closed-to-all
# answer names its reason on stderr.
set -uo pipefail

ALL=",cli-1,cli-2,cli-3,cli-4,rosterguard,remainder,"
CLI_PKG="internal/cli"
RG_PKG="internal/harness/rosterguard"

close_all() {
    echo "race-scope: running all six shards: $1" >&2
    echo "shards=$ALL"
    exit 0
}

changed=""
while [ $# -gt 0 ]; do
    case "$1" in
        --changed)
            [ $# -ge 2 ] || close_all "--changed needs a file"
            changed="$2"
            shift 2
            ;;
        *) close_all "unknown argument: $1" ;;
    esac
done
[ -n "$changed" ] && [ -r "$changed" ] || close_all "no readable changed-file list"

count=0
changed_dirs=""
while IFS= read -r f || [ -n "$f" ]; do
    f="${f#"${f%%[![:space:]]*}"}"
    f="${f%"${f##*[![:space:]]}"}"
    [ -n "$f" ] || continue
    count=$((count + 1))
    case "$f" in
        /* | ../* | */../*) close_all "path is not repository-relative: $f" ;;
    esac
    case "$f" in
        *.go) ;;
        *) close_all "non-Go file in the diff: $f" ;;
    esac
    case "/$f" in
        */testdata/* | */vendor/*) close_all "testdata or vendor path: $f" ;;
    esac
    changed_dirs="$changed_dirs
$(dirname "$f")"
done < "$changed"
[ "$count" -gt 0 ] || close_all "empty changed-file list"

command -v go > /dev/null 2>&1 || close_all "go is not available"
[ -f go.mod ] || close_all "no go.mod in $(pwd)"
mod="$(go list -m 2> /dev/null)" || close_all "go list -m failed"
[ -n "$mod" ] || close_all "module path unreadable"

# One row per package: import path | package name | T when it has test files.
pkgs="$(go list -f '{{.ImportPath}}|{{.Name}}|{{if or .TestGoFiles .XTestGoFiles}}T{{else}}-{{end}}' ./... 2> /dev/null)" \
    || close_all "go list ./... failed"
[ -n "$pkgs" ] || close_all "go list ./... returned no packages"

# Changed directories -> import paths; every one must be a package of the module.
changed_ips=""
while IFS= read -r d; do
    [ -n "$d" ] || continue
    if [ "$d" = "." ]; then ip="$mod"; else ip="$mod/$d"; fi
    row="$(printf '%s\n' "$pkgs" | awk -F'|' -v ip="$ip" '$1 == ip {print; exit}')"
    [ -n "$row" ] || close_all "no package at $d (deleted, renamed, or outside the module)"
    case "$row" in
        *"|main|"*) close_all "main package $d (tests build the command binary)" ;;
    esac
    changed_ips="$changed_ips
$ip"
done << EOF
$changed_dirs
EOF

# Packages that have tests, split into the shard groups.
test_ips="$(printf '%s\n' "$pkgs" | awk -F'|' '$3 == "T" {print $1}')"
cli_ip="$mod/$CLI_PKG"
rg_ip="$mod/$RG_PKG"
rest_ips="$(printf '%s\n' "$test_ips" | grep -vx -- "$cli_ip" | grep -vx -- "$rg_ip")"

# deps_of <import paths…>: the packages those targets (and their tests) need,
# one per line, with go list's test-variant decorations removed.
deps_of() {
    go list -test -deps -f '{{.ImportPath}}' "$@" 2> /dev/null | sed 's/ \[.*$//' | grep -v '\.test$' | sort -u
}

hits() { # hits <import paths…>: 0 when a changed package is among the deps, 1 when none is, 2 on failure
    local deps c
    deps="$(deps_of "$@")" || return 2
    [ -n "$deps" ] || return 2
    while IFS= read -r c; do
        [ -n "$c" ] || continue
        if grep -qxF -- "$c" <<< "$deps"; then
            return 0
        fi
    done <<< "$changed_ips"
    return 1
}

out=","
if printf '%s\n' "$test_ips" | grep -qxF -- "$cli_ip"; then
    hits "$cli_ip"
    case $? in
        0) out="${out}cli-1,cli-2,cli-3,cli-4," ;;
        1) ;;
        *) close_all "dependency listing of $CLI_PKG failed" ;;
    esac
fi
if printf '%s\n' "$test_ips" | grep -qxF -- "$rg_ip"; then
    hits "$rg_ip"
    case $? in
        0) out="${out}rosterguard," ;;
        1) ;;
        *) close_all "dependency listing of $RG_PKG failed" ;;
    esac
fi
if [ -n "$rest_ips" ]; then
    # shellcheck disable=SC2086 # one import path per word, by construction
    hits $rest_ips
    case $? in
        0) out="${out}remainder," ;;
        1) ;;
        *) close_all "dependency listing of the remainder packages failed" ;;
    esac
fi

[ "$out" != "," ] || close_all "no shard depends on the changed packages (not narrowing to nothing)"
echo "race-scope: narrowed to $out" >&2
echo "shards=$out"
