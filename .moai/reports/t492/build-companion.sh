#!/usr/bin/env bash
# t492 C2 — rebuild verification-claim-integrity-detail.md with the relocated §2.1 material
# inserted before the trailing classification block.
set -eu
ROOT="$1"
D="$ROOT/.claude/rules/moai/core/verification-claim-integrity-detail.md"
R="$ROOT/.moai/reports/t492"
TMP="$R/.companion.new"

head -n 105 "$D"                                    >  "$TMP"
cat "$R/companion-insert-header.md"                 >> "$TMP"
cat "$R/moved-block-A-four-tests.md"                >> "$TMP"
printf '\n'                                         >> "$TMP"
cat "$R/moved-block-B-instances-limits-divergence.md" >> "$TMP"
tail -n 5 "$D"                                      >> "$TMP"

mv "$TMP" "$D"
wc -c "$D"
