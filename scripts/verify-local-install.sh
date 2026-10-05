#!/bin/sh
# Verify a local MoAI installation without scanning binary strings.
set -eu

source_bin=${1:-bin/moai}
installed_bin=${2:-${HOME:?HOME is required}/go/bin/moai}
expected_commit=${3:-}

if [ ! -x "$source_bin" ]; then
    printf 'local-install-check: built binary is missing or not executable: %s\n' "$source_bin" >&2
    exit 1
fi

if [ ! -x "$installed_bin" ]; then
    printf 'local-install-check: installed binary is missing or not executable: %s\n' "$installed_bin" >&2
    exit 1
fi

if ! cmp -s "$source_bin" "$installed_bin"; then
    printf 'local-install-check: FAIL — installed binary differs from %s\n' "$source_bin" >&2
    exit 1
fi

version_output=$("$installed_bin" version) || {
    version_exit=$?
    printf 'local-install-check: FAIL — installed binary version exited %s\n' "$version_exit" >&2
    exit "$version_exit"
}

if [ -z "$version_output" ]; then
    printf 'local-install-check: FAIL — installed binary returned an empty version\n' >&2
    exit 1
fi

if [ -z "$expected_commit" ]; then
    if expected_commit=$(git rev-parse --short HEAD 2>/dev/null); then
        :
    elif [ -x /Library/Developer/CommandLineTools/usr/bin/git ] \
        && expected_commit=$(/Library/Developer/CommandLineTools/usr/bin/git rev-parse --short HEAD 2>/dev/null); then
        :
    else
        printf 'local-install-check: FAIL — current commit is unavailable; pass the expected short commit as argument 3\n' >&2
        exit 1
    fi
fi

case "$expected_commit" in
    ''|*[!0123456789abcdefABCDEF]*)
        printf 'local-install-check: FAIL — expected commit is not hexadecimal: %s\n' "$expected_commit" >&2
        exit 1
        ;;
esac

if [ "${#expected_commit}" -lt 7 ]; then
    printf 'local-install-check: FAIL — expected commit is shorter than 7 characters: %s\n' "$expected_commit" >&2
    exit 1
fi

case "$version_output" in
    *"$expected_commit"*) ;;
    *)
        printf 'local-install-check: FAIL — installed binary does not report expected commit %s\n' "$expected_commit" >&2
        exit 1
        ;;
esac

printf '%s\n' "$version_output"
printf 'local-install-check: OK — %s is byte-identical to %s and reports commit %s\n' \
    "$installed_bin" "$source_bin" "$expected_commit"
