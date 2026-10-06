#!/bin/sh
# Read-only hash of the protected real set (SPEC-PLUGIN-MARKETPLACE-001, AC-025 (d), design.md section 6).
# Prints one line: PROTECTED-SET <sha256> entries=<n>
#
# The hash covers directory ENTRIES (directories and empty directories are lines), so an empty directory
# created by a leaked plugin install changes it, plus the content of settings.json, installed_plugins.json
# and config.toml. A root that does not exist contributes "ABSENT <path>", so creating it changes the hash.
#
# Usage: protected-set-hash.sh [--save-roots FILE | --roots-file FILE] [--dump FILE]
#   --save-roots FILE  write the roots derived from the caller's environment to FILE and exit; call it before
#                      any environment scrub so the roots follow the caller's own CLAUDE_CONFIG_DIR / CODEX_HOME
#   --roots-file FILE  hash exactly the roots in FILE (default: derive them now)
#   --dump FILE        also write the sorted entry list, so a difference can be shown entry by entry
set -eu

tab=$(printf '\t')
save_roots=""
roots_file=""
dump_file=""
while [ $# -gt 0 ]; do
    case $1 in
        --save-roots) save_roots=${2:?--save-roots needs a file}; shift 2 ;;
        --roots-file) roots_file=${2:?--roots-file needs a file}; shift 2 ;;
        --dump) dump_file=${2:?--dump needs a file}; shift 2 ;;
        *) printf 'protected-set-hash: unknown argument: %s\n' "$1" >&2; exit 2 ;;
    esac
done

write_roots() {
    claude=${CLAUDE_CONFIG_DIR:-${HOME:?HOME is required}/.claude}
    codex=${CODEX_HOME:-${HOME:?HOME is required}/.codex}
    printf 'ROOT\t3\t%s\n' "$claude/plugins" "$codex/plugins"
    printf 'ROOT\t2\t%s\n' "$codex/.tmp/marketplaces"
    printf 'ROOT\t1\t%s\n' "$HOME/.moai"
    # Subtrees that change on their own are excluded by declaration; <codex>/tmp is deliberately not a root.
    printf 'EXCLUDE\t%s\n' "$claude/plugins/synced" "$claude/plugins/.trash" "$codex/.tmp/marketplaces/.staging"
    printf 'FILE\t%s\n' "$claude/settings.json" "$HOME/.claude/settings.json" "$codex/config.toml"
}

if [ -n "$save_roots" ]; then
    write_roots > "$save_roots"
    exit 0
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT INT TERM
if [ -z "$roots_file" ]; then
    roots_file=$work/roots
    write_roots > "$roots_file"
fi

sha256() {
    if command -v shasum >/dev/null 2>&1; then shasum -a 256; else sha256sum; fi | awk '{print $1}'
}

# 1. entries under every root, minus the declared exclusions
: > "$work/found"
while IFS=$tab read -r kind depth path; do
    [ "$kind" = ROOT ] || continue
    if [ -d "$path" ]; then
        find "$path" -maxdepth "$depth" -print >> "$work/found" 2>/dev/null || true
    else
        printf 'ABSENT %s\n' "$path" >> "$work/found"
    fi
done < "$roots_file"
awk -F'\t' '$1 == "EXCLUDE" { print $2 }' "$roots_file" > "$work/exclude"
awk 'NR == FNR { ex[$0] = 1; next }
     { for (e in ex) if ($0 == e || index($0, e "/") == 1) next; print }' "$work/exclude" "$work/found" > "$work/entries"

# 2. content of the named files and of every installed_plugins.json inside the roots
{
    cat "$work/entries"
    awk -F'\t' '$1 == "FILE" { print $2 }' "$roots_file" | while IFS= read -r f; do
        if [ -f "$f" ]; then printf 'CONTENT %s %s\n' "$(sha256 < "$f")" "$f"; else printf 'ABSENT-FILE %s\n' "$f"; fi
    done
    grep '/installed_plugins\.json$' "$work/entries" | while IFS= read -r f; do
        [ -f "$f" ] && printf 'CONTENT %s %s\n' "$(sha256 < "$f")" "$f"
    done || true
} | LC_ALL=C sort > "$work/sorted"

[ -z "$dump_file" ] || cp "$work/sorted" "$dump_file"
n=$(wc -l < "$work/sorted" | tr -d ' ')
printf 'PROTECTED-SET %s entries=%s\n' "$(sha256 < "$work/sorted")" "$n"
