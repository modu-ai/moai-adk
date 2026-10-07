#!/usr/bin/env bash
# Explicit sync backup helper; never registered as an automatic hook.
set -euo pipefail

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
usage() {
    printf 'usage: %s create <new-backup-dir> <project-root> <relative-path>...\n' "$0" >&2
    printf '       %s verify <backup-dir>\n' "$0" >&2
    exit 2
}

relative_path() {
    case "$1" in
        ''|/*|*/|*//*|*$'\t'*|*$'\n'*|*$'\r'*) fail "unsupported relative path: $1" ;;
    esac
    case "/$1/" in */../*|*/./*) fail "unsafe relative path: $1" ;; esac
    [ "$1" != manifest.tsv ] || fail 'manifest.tsv is reserved'
}

# Check every component, not just the leaf: copies and hashes never follow links.
no_links() {
    local base=$1 rest=$2 component
    while [ -n "$rest" ]; do
        component=${rest%%/*}
        base="$base/$component"
        [ ! -L "$base" ] || fail "symbolic link is unsupported: $base"
        if [ "$rest" = "$component" ]; then break; fi
        rest=${rest#*/}
    done
}

hash_file() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum < "$1" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 < "$1" | awk '{print $1}'
    else
        fail 'SHA-256 tool unavailable (sha256sum or shasum required)'
    fi
}

create_backup() {
    [ "$#" -ge 3 ] || usage
    local backup_dir=$1 project_root=$2 path source links digest
    shift 2
    [ ! -e "$backup_dir" ] && [ ! -L "$backup_dir" ] || fail 'backup directory already exists'
    project_root=$(cd "$project_root" && pwd -P)
    # Validate all inputs before copying anything. Missing inputs remain explicit.
    for path in "$@"; do
        relative_path "$path"
        no_links "$project_root" "$path"
        source="$project_root/$path"
        if [ -d "$source" ]; then
            links=$(find "$source" -type l -print) || fail 'cannot inspect backup input'
            [ -z "$links" ] || fail "symbolic link in backup input: $path"
        elif [ -e "$source" ] && [ ! -f "$source" ]; then
            fail "unsupported backup input: $path"
        fi
    done
    mkdir -p "$(dirname "$backup_dir")"
    mkdir "$backup_dir"
    backup_dir=$(cd "$backup_dir" && pwd -P)
    for path in "$@"; do
        source="$project_root/$path"
        case "$backup_dir/" in "$source/"*) fail 'backup cannot be inside an input directory' ;; esac
        if [ ! -e "$source" ]; then
            printf 'MISSING\t%s\n' "$path" >> "$backup_dir/manifest.tsv"
            continue
        fi
        [ ! -e "$backup_dir/$path" ] || fail "overlapping backup inputs: $path"
        mkdir -p "$(dirname "$backup_dir/$path")"
        cp -R "$source" "$backup_dir/$path"
    done
    touch "$backup_dir/manifest.tsv"
    while IFS= read -r -d '' source; do
        path=${source#"$backup_dir/"}
        [ "$path" != manifest.tsv ] || continue
        relative_path "$path"
        digest=$(hash_file "$source") || fail "cannot hash: $path"
        printf 'FILE\t%s\t%s\n' "$digest" "$path" >> "$backup_dir/manifest.tsv"
    done < <(find "$backup_dir" -type f -print0)
    printf 'PASS: backup manifest created (%s)\n' "$backup_dir/manifest.tsv"
}

verify_backup() {
    [ "$#" -eq 1 ] || usage
    local backup_dir=$1 manifest row kind digest path extra actual links seen=$'\n' entries=0
    [ -d "$backup_dir" ] && [ ! -L "$backup_dir" ] || fail 'backup directory missing or linked'
    manifest="$backup_dir/manifest.tsv"
    [ -f "$manifest" ] && [ ! -L "$manifest" ] || fail 'backup manifest missing or linked'
    links=$(find "$backup_dir" -type l -print) || fail 'cannot inspect backup'
    [ -z "$links" ] || fail 'symbolic link in backup'
    while IFS= read -r row || [ -n "$row" ]; do
        IFS=$'\t' read -r kind digest path extra <<< "$row"
        case "$kind" in
            MISSING) fail "required backup input was missing: $digest" ;;
            FILE) ;;
            *) fail 'malformed backup manifest row' ;;
        esac
        [ "$row" = "FILE"$'\t'"$digest"$'\t'"$path" ] || fail 'malformed backup manifest fields'
        [[ "$digest" =~ ^[0-9a-f]{64}$ ]] || fail 'malformed SHA-256 digest'
        relative_path "$path"
        no_links "$backup_dir" "$path"
        case "$seen" in *$'\n'"$path"$'\n'*) fail "duplicate manifest path: $path" ;; esac
        seen="$seen$path"$'\n'
        [ -f "$backup_dir/$path" ] || fail "backup file missing: $path"
        actual=$(hash_file "$backup_dir/$path") || fail "cannot hash: $path"
        [ "$actual" = "$digest" ] || fail "backup hash mismatch: $path"
        entries=$((entries + 1))
    done < "$manifest"
    [ "$entries" -gt 0 ] || fail 'backup manifest contains no files'
    while IFS= read -r -d '' actual; do
        path=${actual#"$backup_dir/"}
        [ "$path" != manifest.tsv ] || continue
        relative_path "$path"
        case "$seen" in *$'\n'"$path"$'\n'*) ;; *) fail "unrecorded backup file: $path" ;; esac
    done < <(find "$backup_dir" -type f -print0)
    printf 'PASS: backup manifest verified (%s files)\n' "$entries"
}

case "${1:-}" in
    create) shift; create_backup "$@" ;;
    verify) shift; verify_backup "$@" ;;
    *) usage ;;
esac
