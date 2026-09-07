#!/bin/bash
# t496 campaign finish — real-home zero-write evidence via control window.
# delta(P,O) minus delta(C,P): anything that changed during the campaign that
# was NOT already churning in the 20s control window is campaign-attributable.
set -u
EXP_ROOT="${1:?usage: campaign5-finish.sh EXP_ROOT}"
HC="$EXP_ROOT/homecheck"

( cd "$HOME/.codex" && find . -type f -print | sort ) > "$HC/O.list"
( cd "$HOME/.codex" && find . -type f -print0 | xargs -0 stat -f '%m %N' | sort -k2 ) > "$HC/O.mtimes"

{
  echo "--- new/removed files: control window (C->P) ---"
  diff "$HC/C.list" "$HC/P.list" || true
  echo "--- new/removed files: campaign window (P->O) ---"
  diff "$HC/P.list" "$HC/O.list" || true
  echo "--- new/removed files NOT explained by control window ---"
  comm -13 <(diff "$HC/C.list" "$HC/P.list" | grep '^> ' | sed 's/^> //' | sort) \
           <(diff "$HC/P.list" "$HC/O.list" | grep '^> ' | sed 's/^> //' | sort) || true
  echo "--- mtime-changed files in campaign window (P->O), full list ---"
  join -v2 -j2 <(sort -k2 "$HC/P.mtimes") <(sort -k2 "$HC/O.mtimes") | sort -k2 || true
  echo "(count: $(join -v2 -j2 <(sort -k2 "$HC/P.mtimes") <(sort -k2 "$HC/O.mtimes") | wc -l | tr -d ' '))"
  echo "--- of those, how many were ALSO churning in control window (C->P) ---"
  comm -12 <(join -v1 -j2 <(sort -k2 "$HC/C.mtimes") <(sort -k2 "$HC/P.mtimes") | sort -k2 | cut -d' ' -f2- | cut -d' ' -f2- | sort) \
           <(join -v2 -j2 <(sort -k2 "$HC/P.mtimes") <(sort -k2 "$HC/O.mtimes") | sort -k2 | cut -d' ' -f2- | cut -d' ' -f2- | sort) | wc -l | tr -d ' '
  echo "--- campaign-window mtime churn NOT in control churn (candidates) ---"
  comm -23 <(join -v2 -j2 <(sort -k2 "$HC/P.mtimes") <(sort -k2 "$HC/O.mtimes") | sort -k2 | cut -d' ' -f2- | cut -d' ' -f2- | sort) \
           <(join -v1 -j2 <(sort -k2 "$HC/C.mtimes") <(sort -k2 "$HC/P.mtimes") | sort -k2 | cut -d' ' -f2- | cut -d' ' -f2- | sort) || true
  echo "--- key-file integrity (config.toml, auth.json, hooks.json) ---"
  ls -la "$HOME/.codex/config.toml" "$HOME/.codex/auth.json" "$HOME/.codex/hooks.json" 2>&1
  shasum -a 256 "$HOME/.codex/config.toml" "$HOME/.codex/auth.json" "$HOME/.codex/hooks.json" 2>&1
  echo "--- campaign captures under EXP (tmp only) ---"
  ls -la "$EXP_ROOT/captures/"
} > "$HC/zero-write-verdict.txt" 2>&1

cat "$HC/zero-write-verdict.txt"
