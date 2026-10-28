#!/bin/sh
set -eu
unset GIT_CONFIG GIT_DIR GIT_COMMON_DIR GIT_WORK_TREE GIT_INDEX_FILE \
  GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_NAMESPACE \
  GIT_PREFIX GIT_SUPER_PREFIX GIT_CEILING_DIRECTORIES GIT_QUARANTINE_PATH
root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
# Keep the invoking user's Go caches outside the disposable HOME.
GOPATH=$(go env GOPATH)
GOMODCACHE=$(go env GOMODCACHE)
GOCACHE=$(go env GOCACHE)
export GOPATH GOMODCACHE GOCACHE
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
mkdir -p "$tmp/home" "$tmp/foreign"
export HOME="$tmp/home" GIT_CONFIG_NOSYSTEM=1
git -C "$tmp/foreign" init -q
git -C "$tmp/foreign" config --local user.name Original
git -C "$tmp/foreign" config --local user.email original@example.invalid
printf 'committed\n' > "$tmp/foreign/protected.txt"
git -C "$tmp/foreign" add protected.txt
git -C "$tmp/foreign" commit -qm baseline
printf 'uncommitted operator work\n' > "$tmp/foreign/protected.txt"
head=$(git -C "$tmp/foreign" rev-parse HEAD)
cp "$tmp/foreign/.git/config" "$tmp/config.before"
cp "$tmp/foreign/.git/index" "$tmp/index.before"
cp "$tmp/foreign/protected.txt" "$tmp/work.before"
printf '[user]\n\tname = External\n' > "$tmp/external.config"
cp "$tmp/external.config" "$tmp/external.before"
for script in test-git-fixture-isolation.sh test-lint-fast.sh; do
  GIT_DIR="$tmp/foreign/.git" GIT_WORK_TREE="$tmp/foreign" \
    GIT_INDEX_FILE="$tmp/foreign/.git/index" GIT_CONFIG="$tmp/external.config" sh "$root/scripts/ci/$script"
  [ "$(git -C "$tmp/foreign" rev-parse HEAD)" = "$head" ]
  cmp "$tmp/config.before" "$tmp/foreign/.git/config"
  cmp "$tmp/index.before" "$tmp/foreign/.git/index"
  cmp "$tmp/work.before" "$tmp/foreign/protected.txt"
  cmp "$tmp/external.before" "$tmp/external.config"
done
echo 'caller Git HEAD, config, index and dirty file preserved: PASS'
