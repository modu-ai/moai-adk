#!/usr/bin/env bash
# t492 — replicates internal/config/token_budget_guard.go: alwaysLoadedSurface + measureAlwaysLoaded.
#
# Surface = every .claude/rules/moai/**/*.md whose frontmatter carries no `paths:` key
# (sorted), followed by the 3 fixed slots (CLAUDE.md, AGENTS.md, output-styles/moai/moai.md).
# Tokens per file = floor(bytes/4); total = sum of the per-file floors (NOT floor of the sum).
#
# Usage: bash measure-surface.sh <repo-root> <out-tsv>
set -u
ROOT="$1"
OUT="$2"

has_paths() {
  awk '
    NR==1 {
      line=$0; sub(/[ \t\r]+$/, "", line)
      if (line != "---") { print "no"; exit }
      next
    }
    {
      line=$0; sub(/[ \t\r]+$/, "", line)
      if (line == "---") { print "no"; exit }
      if (index($0, "paths:") == 1) { print "yes"; exit }
    }
    END { print "no" }
  ' "$1" | head -1
}

: > "$OUT"
total=0
count=0

find "$ROOT/.claude/rules/moai" -type f -name '*.md' 2>/dev/null | LC_ALL=C sort > "$OUT.all"

while IFS= read -r f; do
  if [ "$(has_paths "$f")" = "yes" ]; then
    continue
  fi
  b=$(wc -c < "$f" | tr -d ' ')
  t=$((b / 4))
  printf '%s\t%s\t%s\n' "$b" "$t" "${f#"$ROOT"/}" >> "$OUT"
  total=$((total + t))
  count=$((count + 1))
done < "$OUT.all"

for f in "$ROOT/CLAUDE.md" "$ROOT/AGENTS.md" "$ROOT/.claude/output-styles/moai/moai.md"; do
  if [ -f "$f" ]; then
    b=$(wc -c < "$f" | tr -d ' ')
  else
    b=0
  fi
  t=$((b / 4))
  printf '%s\t%s\t%s\n' "$b" "$t" "${f#"$ROOT"/}" >> "$OUT"
  total=$((total + t))
  count=$((count + 1))
done

rm -f "$OUT.all"
echo "SCRIPT_TOTAL_TOKENS=$total SCRIPT_ENTRIES=$count"
