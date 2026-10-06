#!/bin/sh
# t1534 round-5 GREEN observation: the REPAIRED validator on the three
# round-5 fixtures (expect 0), plus the whole existing probe suite for
# regression. The base-fail P1 watch probe runs the repaired loop with the
# pr-view-fails stub and expects abort exit 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)"
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
cd "$HERE/matrix-squote" && sh "$VALIDATOR" >/dev/null 2>&1
echo "squote GREEN(expect 0)=$?"
cd "$HERE/matrix-readd" && sh "$VALIDATOR" >/dev/null 2>&1
echo "readd GREEN(expect 0)=$?"
cd "$HERE/matrix-include-suffix" && sh "$VALIDATOR" >/dev/null 2>&1
echo "include-suffix GREEN(expect 0)=$?"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$HERE/stub-base-fail/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$REPO_ROOT/scripts/ci-watch/run.sh" 99 feature/not-an-ssot-key >/tmp/t1534-r5-green-watch.txt 2>&1
echo "base-fail GREEN(expect 1)=$?"
grep -c "cannot resolve" /tmp/t1534-r5-green-watch.txt
exit 0
