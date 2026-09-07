#!/bin/bash
# t504 supplementary cells — SPEC-CODEX-SKILLCONFIG-SHAPE-001 round 2.
# Round 1 discovered codex 0.153.4 REQUIRES the `enabled` field inside
# [[skills.config]] (V1/V2/V3 all hard-errored on its absence), which
# confounded the intended controls. These cells redo the matrix with
# enabled = true, plus the real-config enabled census and the R-cell
# moai-hit identification. Same EXP_ROOT as round 1 (passed as $1).

set -u
EXP_ROOT="${1:?usage: harness2.sh EXP_ROOT}"
[ -d "$EXP_ROOT" ] || { echo "EXP_ROOT missing: $EXP_ROOT"; exit 1; }

SKILL_BODY='---
name: t504probe
description: T504MARKER probe skill for SPEC-CODEX-SKILLCONFIG-SHAPE-001.
---

# t504probe

T504MARKER body marker. This skill exists only for the t504 measurement.
'

mkdir -p "$EXP_ROOT/home-V5" "$EXP_ROOT/home-V6" "$EXP_ROOT/home-V7"

# V5: existing FILE + enabled = true — the true RQ1 cell.
cat > "$EXP_ROOT/home-V5/config.toml" <<EOF
# t504 cell V5: neutral existing FILE + enabled = true
[[skills.config]]
path = "$EXP_ROOT/neutral/t504probe/SKILL.md"
enabled = true
EOF

# V6: nonexistent path + enabled = true — the clean invalid control.
cat > "$EXP_ROOT/home-V6/config.toml" <<EOF
# t504 cell V6: nonexistent path + enabled = true
[[skills.config]]
path = "$EXP_ROOT/neutral/absent/t504probe/SKILL.md"
enabled = true
EOF

# V7: existing DIRECTORY + enabled = true.
cat > "$EXP_ROOT/home-V7/config.toml" <<EOF
# t504 cell V7: neutral existing DIRECTORY + enabled = true
[[skills.config]]
path = "$EXP_ROOT/neutral/t504probe"
enabled = true
EOF

run_cell() {
  cell="$1"; home="$2"
  ( cd "$EXP_ROOT/proj" && CODEX_HOME="$home" codex debug prompt-input "hi" ) \
    > "$EXP_ROOT/cells/$cell.out" 2> "$EXP_ROOT/cells/$cell.err"
  echo $? > "$EXP_ROOT/cells/$cell.rc"
}

run_cell V5 "$EXP_ROOT/home-V5"
run_cell V6 "$EXP_ROOT/home-V6"
run_cell V7 "$EXP_ROOT/home-V7"

{
  echo "---round2-cells---"
  for c in V5 V6 V7; do
    out="$EXP_ROOT/cells/$c.out"; err="$EXP_ROOT/cells/$c.err"
    ob=$(wc -c < "$out" | tr -d ' ')
    eb=$(wc -c < "$err" | tr -d ' ')
    rc=$(cat "$EXP_ROOT/cells/$c.rc")
    m=$(grep -c T504MARKER "$out" || true)
    printf '%s\tout=%s\terr=%s\trc=%s\tmarker=%s\n' "$c" "$ob" "$eb" "$rc" "$m"
  done
  echo "---round2-stderr---"
  for c in V5 V6 V7; do
    if [ -s "$EXP_ROOT/cells/$c.err" ]; then
      echo "[$c] stderr:"; head -5 "$EXP_ROOT/cells/$c.err"
    fi
  done
  echo "---v5-marker-context---"
  grep -o '.\{80\}T504MARKER.\{80\}' "$EXP_ROOT/cells/V5.out" | head -3
  echo "---real-config-enabled-census---"
  total=$(grep -c '^\[\[skills.config\]\]' "$HOME/.codex/config.toml" || true)
  enabled=$(grep -A3 '^\[\[skills.config\]\]' "$HOME/.codex/config.toml" | grep -c '^\s*enabled\s*=' || true)
  echo "real_entries=$total with_enabled_within_3_lines=$enabled"
  echo "---r-cell-moai-hit-identification---"
  grep -n 'moai-' "$EXP_ROOT/cells/R.out" | head -3
} > "$EXP_ROOT/summary2.txt" 2>&1

cat "$EXP_ROOT/summary2.txt"
