#!/usr/bin/env bash
set -euo pipefail

usage() {
    printf 'usage: %s create <backup-dir> <project-root> <path>...\n' "$0" >&2
    printf '       %s verify <backup-dir>\n' "$0" >&2
    exit 2
}

hash_file() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    else
        printf 'sha256 tool unavailable\n' >&2
        return 127
    fi
}

create_backup() {
    backup_dir=$1
    project_root=$2
    shift 2
    [ "$#" -gt 0 ] || usage
    mkdir -p "$backup_dir"
    manifest="$backup_dir/manifest.tsv"
    : > "$manifest"
    for relative_path in "$@"; do
        source_path="$project_root/$relative_path"
        if [ ! -e "$source_path" ]; then
            printf 'MISSING\t%s\n' "$relative_path" >> "$manifest"
            continue
        fi
        mkdir -p "$(dirname "$backup_dir/$relative_path")"
        cp -R "$source_path" "$backup_dir/$relative_path"
    done
    while IFS= read -r -d '' backed_path; do
        relative_path="${backed_path#"$backup_dir/"}"
        [ "$relative_path" = "manifest.tsv" ] && continue
        printf 'FILE\t%s\t%s\n' "$(hash_file "$backed_path")" "$relative_path" >> "$manifest"
    done < <(find "$backup_dir" -type f -print0 | sort -z)
    printf 'PASS: backup manifest created (%s)\n' "$manifest"
}

verify_backup() {
    backup_dir=$1
    manifest="$backup_dir/manifest.tsv"
    [ -f "$manifest" ] || { printf 'FAIL: backup manifest missing\n' >&2; return 1; }
    failures=0
    entries=0
    while IFS=$'\t' read -r kind first second; do
        case "$kind" in
            FILE)
                entries=$((entries + 1))
                backed_path="$backup_dir/$second"
                if [ ! -f "$backed_path" ] || [ "$(hash_file "$backed_path")" != "$first" ]; then
                    printf 'FAIL: backup hash mismatch: %s\n' "$second" >&2
                    failures=$((failures + 1))
                fi
                ;;
            MISSING)
                printf 'FAIL: required backup input was missing: %s\n' "$first" >&2
                failures=$((failures + 1))
                ;;
            '') ;;
            *) printf 'FAIL: malformed backup manifest row\n' >&2; failures=$((failures + 1)) ;;
        esac
    done < "$manifest"
    [ "$entries" -gt 0 ] || { printf 'FAIL: backup manifest contains no files\n' >&2; return 1; }
    [ "$failures" -eq 0 ] || return 1
    printf 'PASS: backup manifest verified (%s files)\n' "$entries"
}

case "${1:-}" in
    create) shift; create_backup "$@" ;;
    verify) shift; [ "$#" -eq 1 ] || usage; verify_backup "$1" ;;
    *) usage ;;
esac
