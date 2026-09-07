#!/bin/bash
# t504 experiment harness — SPEC-CODEX-SKILLCONFIG-SHAPE-001
# CODEX_HOME-isolated cell matrix per plan.md §C/§F. Serial, 0 model calls,
# no real-home writes (cell R reads only). Runs in its own process; does not
# touch the caller's session cwd.

set -u

EXP_ROOT="$(mktemp -d /tmp/t504-XXXXXXXX)"
mkdir -p "$EXP_ROOT/proj" "$EXP_ROOT/neutral" "$EXP_ROOT/cells"
mkdir -p "$EXP_ROOT/home-IV" "$EXP_ROOT/home-N" "$EXP_ROOT/home-V1" \
         "$EXP_ROOT/home-V2" "$EXP_ROOT/home-V3" "$EXP_ROOT/home-V4"

# ---- Pre-flight (plan.md §C) --------------------------------------------
codex --version > "$EXP_ROOT/version.txt" 2>&1
shasum -a 256 "$HOME/.codex/config.toml" > "$EXP_ROOT/hash-before.txt" 2>&1

# ---- Fixtures -------------------------------------------------------------
# Shared probe skill body (name t504probe, token T504MARKER in frontmatter
# description AND body).
SKILL_BODY='---
name: t504probe
description: T504MARKER probe skill for SPEC-CODEX-SKILLCONFIG-SHAPE-001.
---

# t504probe

T504MARKER body marker. This skill exists only for the t504 measurement.
'

# IV home: unique AGENTS.md marker, NO skills config.
printf '%s\n' 'T504IVMARKER isolated home marker for t504.' > "$EXP_ROOT/home-IV/AGENTS.md"
printf '# t504 cell IV: no skills config\n' > "$EXP_ROOT/home-IV/config.toml"

# N home: skill via the $CODEX_HOME/skills/ convention, NO skills.config.
mkdir -p "$EXP_ROOT/home-N/skills/t504probe"
printf '%s' "$SKILL_BODY" > "$EXP_ROOT/home-N/skills/t504probe/SKILL.md"
printf '# t504 cell N: directory convention only\n' > "$EXP_ROOT/home-N/config.toml"

# Neutral fixture target (V1/V2/V3/V4) — under EXP_ROOT, outside every
# known skill root.
mkdir -p "$EXP_ROOT/neutral/t504probe"
printf '%s' "$SKILL_BODY" > "$EXP_ROOT/neutral/t504probe/SKILL.md"

# V1: existing FILE at the neutral path — the moai-entry shape.
cat > "$EXP_ROOT/home-V1/config.toml" <<EOF
# t504 cell V1: neutral existing FILE
[[skills.config]]
path = "$EXP_ROOT/neutral/t504probe/SKILL.md"
EOF

# V2: nonexistent path — the lead-mandated invalid control.
cat > "$EXP_ROOT/home-V2/config.toml" <<EOF
# t504 cell V2: nonexistent path
[[skills.config]]
path = "$EXP_ROOT/neutral/absent/t504probe/SKILL.md"
EOF

# V3: existing DIRECTORY at the neutral path.
cat > "$EXP_ROOT/home-V3/config.toml" <<EOF
# t504 cell V3: neutral existing DIRECTORY
[[skills.config]]
path = "$EXP_ROOT/neutral/t504probe"
EOF

# V4: V1 plus enabled = false.
cat > "$EXP_ROOT/home-V4/config.toml" <<EOF
# t504 cell V4: existing FILE + enabled = false
[[skills.config]]
path = "$EXP_ROOT/neutral/t504probe/SKILL.md"
enabled = false
EOF

# ---- Cell runner ----------------------------------------------------------
run_cell() {
  cell="$1"; home="$2"
  if [ "$home" = "REAL" ]; then
    ( cd "$EXP_ROOT/proj" && codex debug prompt-input "hi" ) \
      > "$EXP_ROOT/cells/$cell.out" 2> "$EXP_ROOT/cells/$cell.err"
  else
    ( cd "$EXP_ROOT/proj" && CODEX_HOME="$home" codex debug prompt-input "hi" ) \
      > "$EXP_ROOT/cells/$cell.out" 2> "$EXP_ROOT/cells/$cell.err"
  fi
  echo $? > "$EXP_ROOT/cells/$cell.rc"
}

# ---- M2: isolated matrix, controls first (serial) -------------------------
run_cell IV "$EXP_ROOT/home-IV"
run_cell N  "$EXP_ROOT/home-N"
run_cell V2 "$EXP_ROOT/home-V2"
run_cell V1 "$EXP_ROOT/home-V1"
run_cell V3 "$EXP_ROOT/home-V3"
run_cell V4 "$EXP_ROOT/home-V4"

# ---- M3: real-environment read-only cell ----------------------------------
cp "$HOME/.codex/config.toml" "$EXP_ROOT/backup-config.toml"
run_cell R "REAL"

shasum -a 256 "$HOME/.codex/config.toml" > "$EXP_ROOT/hash-after.txt" 2>&1

# ---- Census (counts from files on disk, never pipe verdicts) --------------
summary() {
  for c in IV N V2 V1 V3 V4 R; do
    out="$EXP_ROOT/cells/$c.out"; err="$EXP_ROOT/cells/$c.err"
    ob=$(wc -c < "$out" | tr -d ' ')
    eb=$(wc -c < "$err" | tr -d ' ')
    rc=$(cat "$EXP_ROOT/cells/$c.rc")
    if [ "$c" = "IV" ]; then
      m=$(grep -c T504IVMARKER "$out" || true)
    else
      m=$(grep -c T504MARKER "$out" || true)
    fi
    moai=$(grep -c 'moai-' "$out" || true)
    printf '%s\tout=%s\terr=%s\trc=%s\tmarker=%s\tmoai=%s\n' "$c" "$ob" "$eb" "$rc" "$m" "$moai"
  done
}

{
  echo "EXP_ROOT=$EXP_ROOT"
  echo "version=$(cat "$EXP_ROOT/version.txt")"
  echo "hash_before=$(cat "$EXP_ROOT/hash-before.txt")"
  echo "hash_after=$(cat "$EXP_ROOT/hash-after.txt")"
  if diff -q "$EXP_ROOT/hash-before.txt" "$EXP_ROOT/hash-after.txt" >/dev/null 2>&1; then
    echo "notouch=IDENTICAL"
  else
    echo "notouch=DIFFERS"
  fi
  echo "---cells---"
  summary
  echo "---stderr-nonempty-cells---"
  for c in IV N V2 V1 V3 V4 R; do
    if [ -s "$EXP_ROOT/cells/$c.err" ]; then
      echo "[$c] stderr:"; head -5 "$EXP_ROOT/cells/$c.err"
    fi
  done
  echo "---iv-skills-section-head---"
  head -c 400 "$EXP_ROOT/cells/IV.out"
  echo
  echo "---v1-skills-section-head---"
  head -c 400 "$EXP_ROOT/cells/V1.out"
  echo
  echo "---r-complaint-sample---"
  grep -i -m 5 'skill' "$EXP_ROOT/cells/R.err" || true
  grep -i -m 5 'skills.config' "$EXP_ROOT/cells/R.out" | head -5 || true
} > "$EXP_ROOT/summary.txt" 2>&1

cat "$EXP_ROOT/summary.txt"
