#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
git -C "$tmp" init -q
git -C "$tmp" config user.name Fixture
git -C "$tmp" config user.email fixture@example.invalid
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
