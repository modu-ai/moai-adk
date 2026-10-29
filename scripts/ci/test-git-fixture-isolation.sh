#!/bin/sh
set -eu
# A caller may itself be inside a Git hook. Fixture setup must not inherit
# the caller's repository, index or object store.
unset GIT_CONFIG GIT_DIR GIT_COMMON_DIR GIT_WORK_TREE GIT_INDEX_FILE \
  GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_NAMESPACE \
  GIT_PREFIX GIT_SUPER_PREFIX GIT_CEILING_DIRECTORIES GIT_QUARANTINE_PATH
root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
git -C "$tmp" init -q
git -C "$tmp" config --local user.name Fixture
git -C "$tmp" config --local user.email fixture@example.invalid
printf 'protected\n' > "$tmp/protected.txt"
git -C "$tmp" add protected.txt
git -C "$tmp" commit -qm baseline
git -C "$tmp" worktree add -qb protected-linked "$tmp/linked"
gitdir=$(git -C "$tmp/linked" rev-parse --absolute-git-dir)
head=$(git -C "$tmp/linked" rev-parse HEAD)
cp "$gitdir/index" "$tmp/index.before"
cp "$tmp/.git/config" "$tmp/config.before"
GIT_DIR="$gitdir" GIT_WORK_TREE="$tmp/linked" GIT_INDEX_FILE="$gitdir/index" \
  GIT_OBJECT_DIRECTORY="$tmp/.git/objects" GIT_ALTERNATE_OBJECT_DIRECTORIES="$tmp/.git/objects" \
  go test "$root/scripts/i18n-validator" "$root/internal/git/convention" "$root/internal/workflow" "$root/internal/gitenv" -count=1
[ "$(git -C "$tmp/linked" rev-parse HEAD)" = "$head" ]
cmp "$gitdir/index" "$tmp/index.before"
cmp "$tmp/.git/config" "$tmp/config.before"
[ -z "$(git -C "$tmp/linked" status --porcelain)" ]
echo 'linked worktree HEAD, index, config and files preserved: PASS'
