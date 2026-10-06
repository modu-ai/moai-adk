#!/bin/sh
# t1534 round-6 RED observation: run the PRE-REPAIR copies (committed
# a236ac8b3) against the four round-6 findings. Expect: validator 3 exit 1
# (phantom), watch 1 exit 0 (false green).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
STUB="$HERE/stub-base-develop/gh"
chmod +x "$STUB" 2>/dev/null || true
cd "$HERE/matrix-block" && sh /tmp/t1534-pre-r6-validator.sh >/dev/null 2>&1
echo "block RED(expect 1)=$?"
cd "$HERE/matrix-merge" && sh /tmp/t1534-pre-r6-validator.sh >/dev/null 2>&1
echo "merge RED(expect 1)=$?"
cd "$HERE/job-id" && sh /tmp/t1534-pre-r6-validator.sh >/dev/null 2>&1
echo "job-id RED(expect 1)=$?"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$STUB" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh /tmp/t1534-pre-r6-watch/run.sh 99 >/tmp/t1534-r6-red-watch.txt 2>&1
echo "unkeyed-base RED(expect 0)=$?"
grep -c "ADVISORY" /tmp/t1534-r6-red-watch.txt
exit 0
