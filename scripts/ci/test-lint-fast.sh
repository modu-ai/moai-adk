#!/bin/sh
# Verify the preflight invokes the v2 fast filter and propagates failures.
set -eu
# Git commands below must create and stage only in the temporary fixture.
unset GIT_CONFIG GIT_DIR GIT_COMMON_DIR GIT_WORK_TREE GIT_INDEX_FILE \
  GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_NAMESPACE \
  GIT_PREFIX GIT_SUPER_PREFIX GIT_CEILING_DIRECTORIES GIT_QUARANTINE_PATH
root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
cat > "$tmp/golangci-lint" <<'STUB'
#!/bin/sh
[ "$#" -eq 2 ] && [ "$1" = run ] && [ "$2" = --fast-only ] || exit 42
exit "${LINT_STUB_EXIT:-0}"
STUB
chmod +x "$tmp/golangci-lint"
PATH="$tmp:$PATH" make -s -C "$root" lint-fast
if PATH="$tmp:$PATH" LINT_STUB_EXIT=7 make -s -C "$root" lint-fast >/dev/null 2>&1; then
  echo 'lint-fast did not propagate a linter failure' >&2
  exit 1
fi
mkdir -p "$tmp/repo/a" "$tmp/repo/b"
git -C "$tmp/repo" init -q
cp "$root/lefthook.yml" "$tmp/repo/lefthook.yml"
printf 'package a\n' > "$tmp/repo/a/a.go"
printf 'package b\n' > "$tmp/repo/b/b.go"
git -C "$tmp/repo" add a/a.go b/b.go
(cd "$tmp/repo" && PATH="$tmp:$PATH" lefthook run pre-commit > "$tmp/hook.log" 2>&1)
if ! grep -q 'lint-staged' "$tmp/hook.log" || grep -q '(skip)' "$tmp/hook.log"; then
  echo 'pre-commit did not run the staged Go lint' >&2
  exit 1
fi
if (cd "$tmp/repo" && PATH="$tmp:$PATH" LINT_STUB_EXIT=7 lefthook run pre-commit >/dev/null 2>&1); then
  echo 'pre-commit did not propagate a linter failure' >&2
  exit 1
fi
echo 'lint-fast invocation and failure propagation: PASS'
