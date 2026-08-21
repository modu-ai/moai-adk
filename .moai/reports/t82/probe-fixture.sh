#!/usr/bin/env bash
# Rebuilds the fixture used by codex-probe.md and codex-probe-p4.md.
#
# The probe reports recorded only the invocations, not how the tree under them
# was built, and the tree itself lived in a session-local scratchpad that does
# not survive the session. That made the measurements unreproducible by anyone
# else. This script is the missing half: run it, then run the commands the two
# reports record, and the same numbers come back.
#
# Usage:
#   bash probe-fixture.sh [target-dir]
#
# Default target: a mktemp -d directory, printed on exit.
# Requires: git, python3, and codex on PATH for the probes themselves.

set -euo pipefail

TARGET="${1:-$(mktemp -d)}"
FIXTURE="$TARGET/codexprobe"
FAKEHOME="$TARGET/fakehome"

mkdir -p "$FIXTURE/area/deep" "$FAKEHOME"

# Root AGENTS.md: a byte ruler. Each marker names the offset it sits at, so the
# last marker that survives truncation reads the effective cap directly off the
# output. 110 B per record (a 10-byte marker line plus a 100-byte pad line),
# carried past 40,000 B so a 32,768 B cap lands well inside the file.
python3 - "$FIXTURE" <<'PY'
import sys, os
fixture = sys.argv[1]
out, size = [], 0
while size < 40000:
    marker = "MARK%05d\n" % size
    out.append(marker); size += len(marker)
    pad = "p" * 99 + "\n"
    out.append(pad); size += len(pad)
with open(os.path.join(fixture, "AGENTS.md"), "w") as f:
    f.write("".join(out))
print("root AGENTS.md", os.path.getsize(os.path.join(fixture, "AGENTS.md")), "B")
PY

# The two nested levels. Kept tiny: what is measured about them is whether they
# are loaded at all, never how much of them survives.
printf '# Area doc\nMARKER_AREA\n' > "$FIXTURE/area/AGENTS.md"
printf '# Deep doc\nMARKER_DEEP\n' > "$FIXTURE/area/deep/AGENTS.md"

# codex resolves the project root through git, so an uninitialised directory
# measures something else entirely (only the CWD's own doc loads). The git init
# is part of the fixture, not incidental setup.
git -C "$FIXTURE" init -q .
git -C "$FIXTURE" add -A
git -C "$FIXTURE" -c user.email=probe@example.invalid -c user.name=probe commit -qm init

# The user-scope config used by the P4 differential. CODEX_HOME points codex at
# this instead of the real ~/.codex, so the measurement never touches the
# operator's own configuration.
printf 'project_doc_max_bytes = 4096\n' > "$FAKEHOME/config.toml"

cat <<EOF

Fixture built at: $FIXTURE
Fake CODEX_HOME:  $FAKEHOME

--- codex-probe.md (loading behaviour) ---

  # cap and tail-truncation, from the repo root
  cd "$FIXTURE" && codex debug prompt-input probe | grep -o 'MARK[0-9]*' | tail -1
  # expect MARK32670 — the next marker at offset 32780 exceeds the 32,768 B cap

  # merge scope: nested docs load only along the git-root -> CWD path
  cd "$FIXTURE" && codex debug prompt-input probe | grep -c 'MARKER_AREA\|MARKER_DEEP'
  # expect 0 — run from the root, the nested docs contribute nothing

  # the same run from inside the chain, with a small root so the budget is free
  printf '# Root doc\nMARKER_ROOT_HEAD\n' > "$FIXTURE/AGENTS.md"
  cd "$FIXTURE/area/deep" && codex debug prompt-input probe | grep -o 'MARKER_[A-Z_]*' | sort -u
  # expect MARKER_AREA, MARKER_DEEP, MARKER_ROOT_HEAD — all three
  # (rebuild the ruler with this script before continuing to the P4 rows)

--- codex-probe-p4.md (config scope, four-way differential) ---

  # row 1: no user config, project file present, project NOT trusted
  printf 'project_doc_max_bytes = 4096\n' > "$FIXTURE/.codex/config.toml"   # mkdir -p first
  cd "$FIXTURE" && codex debug prompt-input probe | grep -o 'MARK[0-9]*' | tail -1
  # expect MARK32670 — the project file is ignored, the default stands

  # row 2: user config only
  rm -rf "$FIXTURE/.codex"
  cd "$FIXTURE" && CODEX_HOME="$FAKEHOME" codex debug prompt-input probe | grep -o 'MARK[0-9]*' | tail -1
  # expect MARK04070 — the user value applies

  # row 3: user config plus an untrusted project file
  mkdir -p "$FIXTURE/.codex" && printf 'project_doc_max_bytes = 8192\n' > "$FIXTURE/.codex/config.toml"
  cd "$FIXTURE" && CODEX_HOME="$FAKEHOME" codex debug prompt-input probe | grep -o 'MARK[0-9]*' | tail -1
  # expect MARK04070 — still the user value; the project file is ignored

  # row 4: the same, with only the trust entry added to the user config
  printf 'project_doc_max_bytes = 4096\n\n[projects."%s"]\ntrust_level = "trusted"\n' "$FIXTURE" > "$FAKEHOME/config.toml"
  cd "$FIXTURE" && CODEX_HOME="$FAKEHOME" codex debug prompt-input probe | grep -o 'MARK[0-9]*' | tail -1
  # expect MARK08140 — the project value now wins

Rows 3 and 4 differ only in the trust entry, which is what isolates trust
registration as the discriminator rather than the file's presence or location.

Every one of these runs writes an empty stderr. That is itself a measurement:
both truncation and a disregarded cap are silent.
EOF
