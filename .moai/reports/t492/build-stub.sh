#!/usr/bin/env bash
# t492 C2 — rebuild verification-claim-integrity.md with the §2.1 procedure blocks relocated.
# Keeps 1-53 (the [HARD] clause + "pin every ref is not the corrective"),
#       77-119 (classification-is-not-remedy, R1-R4 + cost table, exemption marker),
#       153-EOF (§2.2, §2.3, §3, trailer).
# Drops  54-76  (the four tests)  and  120-152 (instances, detection limits, divergence figure).
# Inserts the [HARD] companion pointer where the tests were.
set -eu
ROOT="$1"
S="$ROOT/.claude/rules/moai/core/verification-claim-integrity.md"
R="$ROOT/.moai/reports/t492"
TMP="$R/.stub.new"

sed -n '1,53p'   "$S" >  "$TMP"
cat "$R/stub-pointer.md" >> "$TMP"
sed -n '77,119p' "$S" >> "$TMP"
sed -n '153,$p'  "$S" >> "$TMP"

mv "$TMP" "$S"
wc -c "$S"
