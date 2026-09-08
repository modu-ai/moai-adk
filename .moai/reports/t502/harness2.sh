#!/bin/bash
# t502 round 2 — discriminating HOW the gate matches, plus dir-shape and
# relative-notation cells. Reuses round 1's EXP_ROOT fixtures.
#
# Round 1 found BOTH the literal (.agents/…) and the resolved (.claude/…)
# notation gate under a symlink mirror, while only the literal one gates under
# a copy mirror. Two explanations predict that equally: lexical normalization,
# or full realpath canonicalization. These cells separate them.

set -u
EXP_ROOT="${1:?usage: harness2.sh EXP_ROOT}"
[ -d "$EXP_ROOT" ] || { echo "EXP_ROOT missing: $EXP_ROOT"; exit 1; }

S_LIT="$EXP_ROOT/proj-S/.agents/skills/t502probe/SKILL.md"
C_LIT="$EXP_ROOT/proj-C/.agents/skills/t502probe/SKILL.md"

# An alias symlink that reaches the SAME canonical skill directory by a path
# sharing no component with either mirror path.
ln -sfn "$EXP_ROOT/proj-S/.claude/skills/t502probe" "$EXP_ROOT/alias-dir"

# Non-normalized spelling of the literal mirror path (same file, `.` + `..`).
E1_PATH="$EXP_ROOT/proj-S/.agents/skills/./t502probe/../t502probe/SKILL.md"
# Reached through the unrelated alias link.
E2_PATH="$EXP_ROOT/alias-dir/SKILL.md"
# /private prefix: on macOS /tmp is itself a symlink to /private/tmp.
E3_PATH="/private${S_LIT}"
E8_PATH="/private${C_LIT}"

mk_home() {   # mk_home <name> <path-or-NONE> <enabled>
  h="$EXP_ROOT/home-$1"; mkdir -p "$h"
  if [ "$2" = "NONE" ]; then
    printf '# t502 cell %s: no skills.config\n' "$1" > "$h/config.toml"
  else
    printf '# t502 cell %s\n[[skills.config]]\npath = "%s"\nenabled = %s\n' "$1" "$2" "$3" > "$h/config.toml"
  fi
}

mk_home EC  NONE - # round-2 control on proj-S
mk_home ECc NONE - # round-2 control on proj-C
mk_home E1  "$E1_PATH" false
mk_home E2  "$E2_PATH" false
mk_home E3  "$E3_PATH" false
mk_home E4  "$EXP_ROOT/proj-S/.agents/skills/t502probe" false          # dir shape, symlink mirror
mk_home E5  "$EXP_ROOT/proj-C/.agents/skills/t502probe" false          # dir shape, copy mirror
mk_home E6  ".agents/skills/t502probe/SKILL.md" false                  # relative notation
mk_home E8  "$E8_PATH" false

run_cell() {   # run_cell <cell> <projdir>
  ( cd "$2" && CODEX_HOME="$EXP_ROOT/home-$1" codex debug prompt-input "hi" ) \
    > "$EXP_ROOT/cells/$1.out" 2> "$EXP_ROOT/cells/$1.err"
  echo $? > "$EXP_ROOT/cells/$1.rc"
}

for c in EC E1 E2 E3 E4; do run_cell "$c" "$EXP_ROOT/proj-S"; done
for c in ECc E5 E6 E8; do run_cell "$c" "$EXP_ROOT/proj-C"; done

shasum -a 256 "$HOME/.codex/config.toml" > "$EXP_ROOT/hash-after2.txt" 2>&1

{
  echo "---round2 paths---"
  echo "E1=$E1_PATH"
  echo "E2=$E2_PATH   (alias -> $(readlink "$EXP_ROOT/alias-dir"))"
  echo "E3=$E3_PATH"
  echo "E6=.agents/skills/t502probe/SKILL.md (relative)"
  echo "E8=$E8_PATH"
  echo "---round2 cells---"
  for c in EC E1 E2 E3 E4 ECc E5 E6 E8; do
    out="$EXP_ROOT/cells/$c.out"; err="$EXP_ROOT/cells/$c.err"
    ob=$(wc -c < "$out" | tr -d ' '); eb=$(wc -c < "$err" | tr -d ' ')
    rc=$(cat "$EXP_ROOT/cells/$c.rc")
    m=$(grep -c T502MARKER "$out" || true)
    n=$(grep -c 't502probe' "$out" || true)
    printf '%s\tout=%s\terr=%s\trc=%s\tmarker=%s\tname=%s\n' "$c" "$ob" "$eb" "$rc" "$m" "$n"
  done
  echo "---round2 stderr-nonempty---"
  for c in EC E1 E2 E3 E4 ECc E5 E6 E8; do
    if [ -s "$EXP_ROOT/cells/$c.err" ]; then echo "[$c]"; head -5 "$EXP_ROOT/cells/$c.err"; fi
  done
  echo "---notouch-recheck---"
  cat "$EXP_ROOT/hash-before.txt"
  cat "$EXP_ROOT/hash-after2.txt"
} > "$EXP_ROOT/summary2.txt" 2>&1

cat "$EXP_ROOT/summary2.txt"
