#!/bin/sh
# t1534 round-8 RED observation: PRE-REPAIR copies (committed 8d1b66b51)
# against the round-8 findings. flow-contexts runs under a yq-less PATH
# (homebrew stripped). bad-link captures the handoff logUrl. Expect:
# flow-contexts watch exit 0 (false green), logUrl pointing at the PASS
# run, name-comment 1, no-space-matrix 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
NOYQ_PATH="/usr/bin:/bin"
cd "$HERE/flow-contexts"
PATH="$NOYQ_PATH" MOAI_CIWATCH_GH="$HERE/stub-main-lint-fail/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh /tmp/t1534-pre-r8-watch/run.sh 99 >/tmp/t1534-r8-red-flow.txt 2>&1
echo "flow-contexts RED(expect 0)=$?"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$HERE/stub-bad-link/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh /tmp/t1534-pre-r8-watch/run.sh 99 >/tmp/t1534-r8-red-link.txt 2>&1
echo "bad-link RED exit(expect 2)=$?"
grep -o "http://example.invalid/[a-z]*" /tmp/t1534-r8-red-link.txt | head -1
cd "$HERE/name-comment" && sh /tmp/t1534-pre-r8-validator.sh >/dev/null 2>&1
echo "name-comment RED(expect 1)=$?"
cd "$HERE/no-space-matrix" && sh /tmp/t1534-pre-r8-validator.sh >/dev/null 2>&1
echo "no-space-matrix RED(expect 1)=$?"
exit 0
