#!/bin/sh
# Check that HOME-isolating fixture scripts preserve the caller's Go paths.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
mkdir -p "$tmp/bin" "$tmp/original-home"
cat > "$tmp/bin/go" <<'STUB'
#!/bin/sh
set -eu
case "$1" in
  env)
    [ "$HOME" = "$FAKE_ORIGINAL_HOME" ] || exit 40
    case "$2" in
      GOPATH) printf '%s\n' "$FAKE_ORIGINAL_HOME/go" ;;
      GOMODCACHE) printf '%s\n' "$FAKE_ORIGINAL_HOME/go/pkg/mod" ;;
      GOCACHE) printf '%s\n' "$FAKE_ORIGINAL_HOME/cache" ;;
      *) exit 41 ;;
    esac ;;
  test)
    [ "$HOME" != "$FAKE_ORIGINAL_HOME" ] || exit 42
    [ "$GOPATH" = "$FAKE_ORIGINAL_HOME/go" ] || exit 43
    [ "$GOMODCACHE" = "$FAKE_ORIGINAL_HOME/go/pkg/mod" ] || exit 44
    [ "$GOCACHE" = "$FAKE_ORIGINAL_HOME/cache" ] || exit 45
    printf '%s\n' "$HOME" >> "$FAKE_TEST_CALLS" ;;
  *) exit 46 ;;
esac
STUB
cat > "$tmp/bin/lefthook" <<'STUB'
#!/bin/sh
printf '%s\n' lint-staged
exit "${LINT_STUB_EXIT:-0}"
STUB
chmod +x "$tmp/bin/go" "$tmp/bin/lefthook"
export FAKE_ORIGINAL_HOME="$tmp/original-home" FAKE_TEST_CALLS="$tmp/calls"
HOME="$FAKE_ORIGINAL_HOME" PATH="$tmp/bin:$PATH" sh "$root/scripts/ci/test-factory-home-isolation.sh"
HOME="$FAKE_ORIGINAL_HOME" PATH="$tmp/bin:$PATH" sh "$root/scripts/ci/test-fixture-caller-isolation.sh"
[ "$(wc -l < "$FAKE_TEST_CALLS")" -eq 2 ]
echo 'fixture Go cache paths preserved before HOME changes: PASS'
