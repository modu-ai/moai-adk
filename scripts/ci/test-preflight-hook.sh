#!/bin/sh
# Exercise the actual hook entrypoint without running the full preflight.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
cat > "$tmp/make" <<'STUB'
#!/bin/sh
[ "$#" -eq 1 ] && [ "$1" = preflight ] || exit 2
for name in GIT_CONFIG GIT_DIR GIT_COMMON_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_NAMESPACE GIT_PREFIX GIT_SUPER_PREFIX GIT_CEILING_DIRECTORIES GIT_QUARANTINE_PATH; do
  eval "value=\${$name+x}"
  [ -z "$value" ] || exit 3
done
[ "${GIT_AUTHOR_NAME:-}" = preserved ] || exit 4
[ "${GIT_CONFIG_GLOBAL:-}" = preserved ] || exit 5
exit "${MAKE_STUB_EXIT:-0}"
STUB
chmod +x "$tmp/make"
PATH="$tmp:$PATH" GIT_CONFIG=bad GIT_DIR=bad GIT_COMMON_DIR=bad GIT_WORK_TREE=bad GIT_INDEX_FILE=bad \
  GIT_OBJECT_DIRECTORY=bad GIT_ALTERNATE_OBJECT_DIRECTORIES=bad GIT_NAMESPACE=bad \
  GIT_PREFIX=bad GIT_SUPER_PREFIX=bad GIT_CEILING_DIRECTORIES=bad GIT_QUARANTINE_PATH=bad \
  GIT_AUTHOR_NAME=preserved GIT_CONFIG_GLOBAL=preserved sh "$root/scripts/ci/preflight-hook.sh"
if PATH="$tmp:$PATH" GIT_AUTHOR_NAME=preserved GIT_CONFIG_GLOBAL=preserved MAKE_STUB_EXIT=7 sh "$root/scripts/ci/preflight-hook.sh"; then
  echo 'preflight failure was not propagated' >&2
  exit 1
fi
echo 'preflight hook environment and failure propagation: PASS'
