#!/bin/sh
# t1534 round-5 RED observation (P1): run the PRE-REPAIR watch loop (/tmp
# copy of the committed ef8ebd702 run.sh) with the pr-view-fails stub and a
# head branch that is not an SSoT key. Expect exit 0 — the false green
# (Lint=fail reclassified advisory, empty required list scored all-passed).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$HERE/stub-base-fail/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 \
CIWATCH_TIMEOUT_SECONDS=10 \
sh /tmp/t1534-pre-r5-watch/run.sh 99 feature/not-an-ssot-key
echo "base-fail RED(expect 0)=$?"
exit 0
