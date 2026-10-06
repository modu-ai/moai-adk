#!/bin/sh
# t1534 round-5 RED observation: run the PRE-REPAIR validator (/tmp copy of
# the committed ef8ebd702 validator) against the three round-5 fixtures.
# Expect exit 1 (phantom) on all three — the defect the repairs remove.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
cd "$HERE/matrix-squote" && sh /tmp/t1534-pre-r5-validator.sh >/dev/null 2>&1
echo "squote RED(expect 1)=$?"
cd "$HERE/matrix-readd" && sh /tmp/t1534-pre-r5-validator.sh >/dev/null 2>&1
echo "readd RED(expect 1)=$?"
cd "$HERE/matrix-include-suffix" && sh /tmp/t1534-pre-r5-validator.sh >/dev/null 2>&1
echo "include-suffix RED(expect 1)=$?"
exit 0
