#!/bin/bash
# t528 — corpus-wide `moai spec view <ID>` hard-error sweep.
#
# NOT `--acceptance`: that flag does not exist (spec_view.go registers only
# --shape-trace), and sweeping with it returns 807/807 "Unknown flag", a figure
# about argument parsing rather than the collector. `moai spec view <ID>` IS the
# acceptance view. See specview-before-20260908.md.
# $1 = binary, $2 = output file. Counts lines ending in a `parse error:` failure.
BIN="$1"; OUT="$2"
: > "$OUT"
n=0; hard=0
while IFS= read -r p; do
  id=$(basename "$(dirname "$p")")
  n=$((n+1))
  err=$("$BIN" spec view "$id" 2>&1 >/dev/null)
  rc=$?
  if [ $rc -ne 0 ]; then
    printf '%s\tRC=%d\t%s\n' "$id" "$rc" "$(printf '%s' "$err" | tr '\n' ' ')" >> "$OUT"
    hard=$((hard+1))
  fi
done < <(sed 's|^\.\./\.\./||' .moai/reports/t528/probe/before/filelist.txt)
echo "SWEPT=$n  NONZERO_EXIT=$hard"
echo "PARSE_ERROR_LINES=$(grep -c 'parse error:' "$OUT" || true)"
