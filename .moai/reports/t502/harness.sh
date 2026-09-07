#!/bin/bash
# t502 gate-path-shape harness — SPEC-CODEX-SKILL-DISABLE-001 plan-phase measurement.
#
# Question: for a skill exposed through moai's .agents/skills MIRROR, does the
# [[skills.config]] enabled=false gate bind the LITERAL path codex loaded from
# (the symlink path) or the symlink-RESOLVED target path — and does whichever
# notation binds survive the copy-fallback mirror shape?
#
# Precedent: .moai/reports/t504/harness.sh + harness3.sh (D-series). t504
# measured the $CODEX_HOME/skills/ convention with a REAL directory; this
# harness measures the cwd .agents/skills/ convention with a SYMLINK and with
# a real COPY, which is what internal/template/skill_mirror.go actually creates.
#
# CODEX_HOME-isolated, serial, 0 model calls, no writes to the real ~/.codex.

set -u

EXP_ROOT="$(mktemp -d /tmp/t502-XXXXXXXX)"
mkdir -p "$EXP_ROOT/cells"

# ---- Pre-flight -----------------------------------------------------------
codex --version > "$EXP_ROOT/version.txt" 2>&1
shasum -a 256 "$HOME/.codex/config.toml" > "$EXP_ROOT/hash-before.txt" 2>&1

SKILL_BODY='---
name: t502probe
description: T502MARKER probe skill for SPEC-CODEX-SKILL-DISABLE-001.
---

# t502probe

T502MARKER body marker. This skill exists only for the t502 measurement.
'

# ---- Project fixtures -----------------------------------------------------
# proj-S: the SYMLINK mirror shape (skill_mirror.go MirrorModeSymlink).
#   .claude/skills/t502probe/SKILL.md  (canonical, real)
#   .agents/skills/t502probe -> ../../.claude/skills/t502probe  (relative link)
mkdir -p "$EXP_ROOT/proj-S/.claude/skills/t502probe" "$EXP_ROOT/proj-S/.agents/skills"
printf '%s' "$SKILL_BODY" > "$EXP_ROOT/proj-S/.claude/skills/t502probe/SKILL.md"
( cd "$EXP_ROOT/proj-S/.agents/skills" && ln -s ../../.claude/skills/t502probe t502probe )

# proj-C: the COPY-fallback mirror shape (skill_mirror.go MirrorModeCopy).
#   both .claude/skills/<n>/SKILL.md and .agents/skills/<n>/SKILL.md are real
#   files; no symlink anywhere in the path.
mkdir -p "$EXP_ROOT/proj-C/.claude/skills/t502probe" "$EXP_ROOT/proj-C/.agents/skills/t502probe"
printf '%s' "$SKILL_BODY" > "$EXP_ROOT/proj-C/.claude/skills/t502probe/SKILL.md"
printf '%s' "$SKILL_BODY" > "$EXP_ROOT/proj-C/.agents/skills/t502probe/SKILL.md"

S_LIT="$EXP_ROOT/proj-S/.agents/skills/t502probe/SKILL.md"   # literal (through the link)
S_RES="$EXP_ROOT/proj-S/.claude/skills/t502probe/SKILL.md"   # symlink-resolved target
C_LIT="$EXP_ROOT/proj-C/.agents/skills/t502probe/SKILL.md"   # literal (real file)
C_RES="$EXP_ROOT/proj-C/.claude/skills/t502probe/SKILL.md"   # the canonical twin

# ---- Homes ----------------------------------------------------------------
mk_home() {   # mk_home <name> <path-or-NONE> <enabled>
  h="$EXP_ROOT/home-$1"; mkdir -p "$h"
  if [ "$2" = "NONE" ]; then
    printf '# t502 cell %s: no skills.config\n' "$1" > "$h/config.toml"
  else
    printf '# t502 cell %s\n[[skills.config]]\npath = "%s"\nenabled = %s\n' "$1" "$2" "$3" > "$h/config.toml"
  fi
}

# Symlink-mirror cells
mk_home S0     NONE     -        # bare control: does the mirror expose at all?
mk_home Split  "$S_LIT" true     # positive control, literal notation
mk_home Spres  "$S_RES" true     # positive control, resolved notation
mk_home Slit   "$S_LIT" false    # does the LITERAL notation gate?
mk_home Sres   "$S_RES" false    # does the RESOLVED notation gate?

# Copy-mirror cells
mk_home C0     NONE     -        # bare control
mk_home Cplit  "$C_LIT" true     # positive control, literal notation
mk_home Cpres  "$C_RES" true     # positive control, canonical-twin notation
mk_home Clit   "$C_LIT" false    # literal notation under copy mode
mk_home Cres   "$C_RES" false    # canonical-twin notation under copy mode

# ---- Cell runner ----------------------------------------------------------
run_cell() {   # run_cell <cell> <home> <projdir>
  ( cd "$3" && CODEX_HOME="$2" codex debug prompt-input "hi" ) \
    > "$EXP_ROOT/cells/$1.out" 2> "$EXP_ROOT/cells/$1.err"
  echo $? > "$EXP_ROOT/cells/$1.rc"
}

for c in S0 Split Spres Slit Sres; do
  run_cell "$c" "$EXP_ROOT/home-$c" "$EXP_ROOT/proj-S"
done
for c in C0 Cplit Cpres Clit Cres; do
  run_cell "$c" "$EXP_ROOT/home-$c" "$EXP_ROOT/proj-C"
done

shasum -a 256 "$HOME/.codex/config.toml" > "$EXP_ROOT/hash-after.txt" 2>&1

# ---- Census (counts from files on disk, never pipe verdicts) --------------
{
  echo "EXP_ROOT=$EXP_ROOT"
  echo "version=$(cat "$EXP_ROOT/version.txt")"
  echo "S_LIT=$S_LIT"
  echo "S_RES=$S_RES"
  echo "C_LIT=$C_LIT"
  echo "C_RES=$C_RES"
  echo "link=$(ls -l "$EXP_ROOT/proj-S/.agents/skills/t502probe")"
  echo "hash_before=$(cat "$EXP_ROOT/hash-before.txt")"
  echo "hash_after=$(cat "$EXP_ROOT/hash-after.txt")"
  if diff -q "$EXP_ROOT/hash-before.txt" "$EXP_ROOT/hash-after.txt" >/dev/null 2>&1; then
    echo "notouch=IDENTICAL"
  else
    echo "notouch=DIFFERS"
  fi
  echo "---cells---"
  for c in S0 Split Spres Slit Sres C0 Cplit Cpres Clit Cres; do
    out="$EXP_ROOT/cells/$c.out"; err="$EXP_ROOT/cells/$c.err"
    ob=$(wc -c < "$out" | tr -d ' '); eb=$(wc -c < "$err" | tr -d ' ')
    rc=$(cat "$EXP_ROOT/cells/$c.rc")
    m=$(grep -c T502MARKER "$out" || true)
    n=$(grep -c 't502probe' "$out" || true)
    printf '%s\tout=%s\terr=%s\trc=%s\tmarker=%s\tname=%s\n' "$c" "$ob" "$eb" "$rc" "$m" "$n"
  done
  echo "---stderr-nonempty---"
  for c in S0 Split Spres Slit Sres C0 Cplit Cpres Clit Cres; do
    if [ -s "$EXP_ROOT/cells/$c.err" ]; then echo "[$c]"; head -5 "$EXP_ROOT/cells/$c.err"; fi
  done
  echo "---S0-marker-context (which path does codex report?)---"
  grep -o '.\{80\}T502MARKER.\{40\}' "$EXP_ROOT/cells/S0.out" | head -3
  echo "---C0-marker-context---"
  grep -o '.\{80\}T502MARKER.\{40\}' "$EXP_ROOT/cells/C0.out" | head -3
  echo "---S0-roots-table---"
  grep -n -i -m 40 'r[0-9]\{1,2\}[:=) ]' "$EXP_ROOT/cells/S0.out" | head -40
} > "$EXP_ROOT/summary.txt" 2>&1

cat "$EXP_ROOT/summary.txt"
