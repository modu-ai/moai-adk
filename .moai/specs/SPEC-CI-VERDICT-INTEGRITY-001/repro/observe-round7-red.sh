#!/bin/sh
# t1534 round-7 RED observation: PRE-REPAIR copies (committed c5f5ff229)
# against the round-7 findings. Expect: merge-all-legs 1, merge-overwrite 1,
# comment-boundary 1, quoted-key watch 1 (spurious abort), and the
# green-gone negative probe 0 on the pre-repair (green wrongly allowed).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
cd "$HERE/merge-all-legs" && sh /tmp/t1534-pre-r7-validator.sh >/dev/null 2>&1
echo "merge-all-legs RED(expect 1)=$?"
cd "$HERE/merge-overwrite" && sh /tmp/t1534-pre-r7-validator.sh >/dev/null 2>&1
echo "merge-overwrite RED(expect 1)=$?"
cd "$HERE/comment-boundary" && sh /tmp/t1534-pre-r7-validator.sh >/dev/null 2>&1
echo "comment-boundary RED(expect 1)=$?"
cd "$HERE/merge-green-gone" && sh /tmp/t1534-pre-r7-validator.sh >/dev/null 2>&1
echo "green-gone RED(expect 0 = wrongly allowed)=$?"
cd "$HERE/quoted-key"
MOAI_CIWATCH_GH="$HERE/stub-skipping/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh /tmp/t1534-pre-r7-watch/run.sh 99 >/tmp/t1534-r7-red-watch.txt 2>&1
echo "quoted-key watch RED(expect 1)=$?"
grep -c "no key for base branch" /tmp/t1534-r7-red-watch.txt
exit 0
