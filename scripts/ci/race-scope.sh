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
# What the answer rests on, stated exactly. The script's evidence is the IMPORT
# graph (`go list -test -deps`, test imports included) plus a textual scan; it is
# not a proof of independence. A shard is left out only when BOTH hold: (1) no
# changed package is among the dependencies of the packages that shard tests, and
# (2) nothing in the shard's module closure (its test files, and the sources of
# every module package they depend on) matches READ_PATTERN — file reads or
# walks, the working directory, commands, go:embed, testdata, "../". A shard
# that matches (2) is selected for ANY Go change, because tests that read files
# or run commands depend on things no import shows (audit round 2: rosterguard
# walks the repository's Go sources). The pattern is textual and broad on
# purpose; what it cannot recognise (a path built in another way, a generated
# file, behaviour switched on by the environment) is a residual risk that stays
# inside the advisory Race Test legs — the required Test job runs every package.
# Narrowing also needs every changed file to be a Go file in a package of this
# module. Everything else — a non-Go file, go.mod/go.sum, a workflow or this script, a
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

# Patterns that mean a package (or a test, or code a test calls) can depend on
# something an import does not show: files read or walked, the working directory,
# commands run, files embedded, testdata, a relative parent path. A match is
# deliberately broad — a false match only selects a shard that could have been
# skipped, never the reverse.
READ_PATTERN='filepath\.Walk|os\.(ReadFile|ReadDir|Open|OpenFile|Stat|Lstat|Getwd|Executable|DirFS)\(|ioutil\.|fs\.(WalkDir|ReadFile|ReadDir|Stat)\(|exec\.Command|runtime\.Caller\(|//go:embed|testdata|\.\./'

# reads_outside_imports <import paths…>: 0 when the targets' module closure (their
# tests and every module package they depend on) may depend on something outside the
# import graph, 1 when no such sign was found, 2 when the listing failed.
reads_outside_imports() {
    local rows ip dir files embeds f
    rows="$(go list -test -deps -f '{{.ImportPath}}|{{.Dir}}|{{join .GoFiles " "}} {{join .TestGoFiles " "}} {{join .XTestGoFiles " "}}|{{join .EmbedPatterns ","}}' "$@" 2> /dev/null)" || return 2
    [ -n "$rows" ] || return 2
    while IFS='|' read -r ip dir files embeds; do
        case "$ip" in
            "$mod" | "$mod"/* | "$mod "*) ;;
            *) continue ;;
        esac
        if [ -n "$embeds" ]; then
            echo "race-scope: $ip embeds files ($embeds)" >&2
            return 0
        fi
        for f in $files; do
            [ -f "$dir/$f" ] || continue
            if grep -Eq -- "$READ_PATTERN" "$dir/$f"; then
                echo "race-scope: $ip: $f matches a file/command/embed pattern" >&2
                return 0
            fi
        done
    done <<< "$rows"
    return 1
}

# needs_shard <label> <import paths…>: 0 select the shard, 1 leave it out. A shard is
# left out only when its closure shows no sign of reading outside its imports AND no
# changed package is among its dependencies; a shard that cannot be shown independent
# is selected for any Go change.
needs_shard() {
    local label="$1" rc
    shift
    reads_outside_imports "$@"
    rc=$?
    case $rc in
        0)
            echo "race-scope: $label may depend on more than its imports: selected" >&2
            return 0
            ;;
        1) ;;
        *) close_all "reading check of $label failed" ;;
    esac
    hits "$@"
    rc=$?
    case $rc in
        0 | 1) return $rc ;;
        *) close_all "dependency listing of $label failed" ;;
    esac
}

out=","
if printf '%s\n' "$test_ips" | grep -qxF -- "$cli_ip"; then
    if needs_shard cli "$cli_ip"; then out="${out}cli-1,cli-2,cli-3,cli-4,"; fi
fi
if printf '%s\n' "$test_ips" | grep -qxF -- "$rg_ip"; then
    if needs_shard rosterguard "$rg_ip"; then out="${out}rosterguard,"; fi
fi
if [ -n "$rest_ips" ]; then
    # shellcheck disable=SC2086 # one import path per word, by construction
    if needs_shard remainder $rest_ips; then out="${out}remainder,"; fi
fi

[ "$out" != "," ] || close_all "no shard depends on the changed packages (not narrowing to nothing)"
if [ "$out" = "$ALL" ]; then
    echo "race-scope: all six shards selected (none could be left out)" >&2
else
    echo "race-scope: narrowed to $out" >&2
fi
echo "shards=$out"
