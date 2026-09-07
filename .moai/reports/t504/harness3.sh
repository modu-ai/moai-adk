#!/bin/bash
# t504 D-series — lead-directed supplementary cells (2026-09-07).
# Question: does an [[skills.config]] entry with enabled=false turn OFF a
# skill that WOULD load via the live $CODEX_HOME/skills/ convention?
#   D2f/D2d: controls — entry with enabled=true (file/dir path shape)
#   D1f/D1d: cells     — entry with enabled=false (file/dir path shape)
# The skill physically lives at $CODEX_HOME/skills/t504probe/SKILL.md in
# every cell (the live convention location); the entry points at that same
# file or its directory. No other extensions per the lead's instruction.

set -u
EXP_ROOT="${1:?usage: harness3.sh EXP_ROOT}"
[ -d "$EXP_ROOT" ] || { echo "EXP_ROOT missing: $EXP_ROOT"; exit 1; }

SKILL_BODY='---
name: t504probe
description: T504MARKER probe skill for SPEC-CODEX-SKILLCONFIG-SHAPE-001.
---

# t504probe

T504MARKER body marker. This skill exists only for the t504 measurement.
'

for h in home-D2f home-D1f home-D2d home-D1d; do
  mkdir -p "$EXP_ROOT/$h/skills/t504probe"
  printf '%s' "$SKILL_BODY" > "$EXP_ROOT/$h/skills/t504probe/SKILL.md"
done

# D2f: live skill + entry(file shape) enabled=true — control.
cat > "$EXP_ROOT/home-D2f/config.toml" <<EOF
# t504 cell D2f: live skill + entry(FILE) enabled = true
[[skills.config]]
path = "$EXP_ROOT/home-D2f/skills/t504probe/SKILL.md"
enabled = true
EOF

# D1f: live skill + entry(FILE shape) enabled=false.
cat > "$EXP_ROOT/home-D1f/config.toml" <<EOF
# t504 cell D1f: live skill + entry(FILE) enabled = false
[[skills.config]]
path = "$EXP_ROOT/home-D1f/skills/t504probe/SKILL.md"
enabled = false
EOF

# D2d: live skill + entry(DIRECTORY shape) enabled=true — control.
cat > "$EXP_ROOT/home-D2d/config.toml" <<EOF
# t504 cell D2d: live skill + entry(DIRECTORY) enabled = true
[[skills.config]]
path = "$EXP_ROOT/home-D2d/skills/t504probe"
enabled = true
EOF

# D1d: live skill + entry(DIRECTORY shape) enabled=false.
cat > "$EXP_ROOT/home-D1d/config.toml" <<EOF
# t504 cell D1d: live skill + entry(DIRECTORY) enabled = false
[[skills.config]]
path = "$EXP_ROOT/home-D1d/skills/t504probe"
enabled = false
EOF

run_cell() {
  cell="$1"; home="$2"
  ( cd "$EXP_ROOT/proj" && CODEX_HOME="$home" codex debug prompt-input "hi" ) \
    > "$EXP_ROOT/cells/$cell.out" 2> "$EXP_ROOT/cells/$cell.err"
  echo $? > "$EXP_ROOT/cells/$cell.rc"
}

run_cell D2f "$EXP_ROOT/home-D2f"
run_cell D1f "$EXP_ROOT/home-D1f"
run_cell D2d "$EXP_ROOT/home-D2d"
run_cell D1d "$EXP_ROOT/home-D1d"

{
  echo "---D-series---"
  for c in D2f D1f D2d D1d; do
    out="$EXP_ROOT/cells/$c.out"; err="$EXP_ROOT/cells/$c.err"
    ob=$(wc -c < "$out" | tr -d ' ')
    eb=$(wc -c < "$err" | tr -d ' ')
    rc=$(cat "$EXP_ROOT/cells/$c.rc")
    m=$(grep -c T504MARKER "$out" || true)
    n=$(grep -c 't504probe' "$out" || true)
    printf '%s\tout=%s\terr=%s\trc=%s\tmarker=%s\tname=%s\n' "$c" "$ob" "$eb" "$rc" "$m" "$n"
  done
  echo "---D-stderr---"
  for c in D2f D1f D2d D1d; do
    if [ -s "$EXP_ROOT/cells/$c.err" ]; then
      echo "[$c] stderr:"; head -5 "$EXP_ROOT/cells/$c.err"
    fi
  done
  echo "---D2f-marker-context---"
  grep -o '.\{60\}T504MARKER.\{60\}' "$EXP_ROOT/cells/D2f.out" | head -2
  echo "---notouch-recheck---"
  shasum -a 256 "$HOME/.codex/config.toml"
} > "$EXP_ROOT/summary3.txt" 2>&1

cat "$EXP_ROOT/summary3.txt"
