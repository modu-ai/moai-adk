#!/usr/bin/env bash
# alias-shape-scan.sh — measure the same-file identifier-aliasing axis.
#
# For each acceptance.md, report identifier pairs (X, Y), X != Y, where the
# numeric tails are equal AND X's alpha prefix is a proper prefix of Y's alpha
# prefix (e.g. AC-01 <~ AC-ORC-001-01). This is the mechanical shape of the
# short-form alias observed in SPEC-V3R2-ORC-001.
#
# Identifiers are fed to awk on stdin (NOT via -v: BSD awk rejects newlines in
# a -v assignment, which silently turns the whole scan into an error stream).
#
# Run from the repository root of this worktree.
for f in .moai/specs/*/acceptance.md; do
  hits=$(grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' "$f" | sort -u | awk '
    { a[++n] = $0 }
    END {
      for (i = 1; i <= n; i++) {
        pi = a[i]; sub(/[0-9]+$/, "", pi); ti = a[i]; sub(/^.*[^0-9]/, "", ti)
        for (j = 1; j <= n; j++) {
          if (i == j) continue
          pj = a[j]; sub(/[0-9]+$/, "", pj); tj = a[j]; sub(/^.*[^0-9]/, "", tj)
          if (ti == tj && length(pi) < length(pj) && substr(pj, 1, length(pi)) == pi)
            print a[i] " <~ " a[j]
        }
      }
    }')
  if [ -n "$hits" ]; then
    echo "=== $f"
    echo "$hits"
  fi
done
