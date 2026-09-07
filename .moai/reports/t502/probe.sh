#!/bin/bash
# t502 gate probe — a re-runnable INSTRUMENT, not a replay of one matrix.
#
# Answers exactly one question per invocation:
#   "is skill <NAME> exposed to codex, in project <PROJECT>, under <CODEX_HOME>?"
# and optionally asserts the answer, so a later phase can use it as a check.
#
# The four things a later phase will vary are inputs, not literals:
#   --codex-home   the isolation root holding config.toml
#   --entry-path   the [[skills.config]] path value        (optional)
#   --enabled      the [[skills.config]] enabled value     (optional)
#   --skill        the probe skill / marker name
#
# THE END-TO-END CASE this exists for: the t502 verb writes a config itself.
# Point --codex-home at the home whose config.toml the verb produced and pass
# NEITHER --entry-path NOR --enabled; the script then reads that config exactly
# as written and never rewrites it. Run once before the verb (--expect exposed)
# and once after (--expect gated) to measure marker 1 -> 0 end to end.
#
# Verified against the known cells of the 2026-09-07 matrix — see --selftest.
#
# Boundaries: never writes outside --codex-home / --project, never touches the
# real ~/.codex (it only sha256's it, before and after, and prints both).

set -u

usage() {
  cat <<'USAGE'
usage:
  probe.sh fixture --root <dir> --mirror symlink|copy [--skill NAME]
      Build an isolated project fixture in <dir> and print the paths a probe
      run needs (PROJECT, LITERAL_PATH, RESOLVED_PATH). <dir> must not exist.

  probe.sh probe --codex-home <dir> --project <dir> [--skill NAME]
                 [--entry-path PATH] [--enabled true|false]
                 [--expect exposed|gated]
      Run one cell and print a one-line verdict. With --entry-path/--enabled
      the script WRITES a config.toml into --codex-home first; without them it
      uses the config already there verbatim (the verb-produced case).
      With --expect it exits 0 on a match, 3 on a mismatch.

  probe.sh selftest [--root <dir>]
      Build both mirror fixtures and re-measure the four load-bearing cells
      (symlink/literal, copy/literal, copy/resolved, bare control), asserting
      the instrument still detects BOTH directions. Exits non-zero if not.

marker convention:
  exposed  <=> grep -c '<MARKER>' <captured stdout file> >= 1
  gated    <=> that count is 0, WITH rc=0 and empty stderr
  MARKER defaults to the uppercased skill name (t502probe -> T502PROBE_MARKER);
  it is written into the probe SKILL.md frontmatter description AND body.
  A count of 0 on a nonzero rc is NOT "gated" — it is a failed run, and the
  script reports it as ERROR rather than as a verdict.
USAGE
}

SKILL_NAME="t502probe"
MARKER=""
ROOT=""; MIRROR=""; CODEX_HOME_ARG=""; PROJECT=""; ENTRY_PATH=""; ENABLED=""; EXPECT=""

SUB="${1:-}"; [ -n "$SUB" ] && shift
while [ $# -gt 0 ]; do
  case "$1" in
    --root)        ROOT="$2"; shift 2 ;;
    --mirror)      MIRROR="$2"; shift 2 ;;
    --codex-home)  CODEX_HOME_ARG="$2"; shift 2 ;;
    --project)     PROJECT="$2"; shift 2 ;;
    --skill)       SKILL_NAME="$2"; shift 2 ;;
    --marker)      MARKER="$2"; shift 2 ;;
    --entry-path)  ENTRY_PATH="$2"; shift 2 ;;
    --enabled)     ENABLED="$2"; shift 2 ;;
    --expect)      EXPECT="$2"; shift 2 ;;
    -h|--help)     usage; exit 0 ;;
    *) echo "unknown flag: $1" >&2; usage >&2; exit 2 ;;
  esac
done
[ -n "$MARKER" ] || MARKER="$(printf '%s' "$SKILL_NAME" | tr '[:lower:]-' '[:upper:]_')_MARKER"

# ---- fixture --------------------------------------------------------------
# Reproduces the two shapes internal/template/skill_mirror.go actually creates:
#   symlink -> MirrorModeSymlink (.agents/skills/<n> -> ../../.claude/skills/<n>)
#   copy    -> MirrorModeCopy, and shape-identical to MirrorModeSkipped
build_fixture() {
  root="$1"; mode="$2"; name="$3"; marker="$4"
  proj="$root/proj"
  mkdir -p "$proj/.claude/skills/$name" "$proj/.agents/skills"
  printf -- '---\nname: %s\ndescription: %s probe skill for the t502 gate instrument.\n---\n\n# %s\n\n%s body marker.\n' \
    "$name" "$marker" "$name" "$marker" > "$proj/.claude/skills/$name/SKILL.md"
  case "$mode" in
    symlink) ( cd "$proj/.agents/skills" && ln -s "../../.claude/skills/$name" "$name" ) ;;
    copy)    mkdir -p "$proj/.agents/skills/$name"
             cp "$proj/.claude/skills/$name/SKILL.md" "$proj/.agents/skills/$name/SKILL.md" ;;
    *) echo "--mirror must be symlink or copy" >&2; return 2 ;;
  esac
  echo "PROJECT=$proj"
  echo "LITERAL_PATH=$proj/.agents/skills/$name/SKILL.md"
  echo "RESOLVED_PATH=$proj/.claude/skills/$name/SKILL.md"
  echo "MARKER=$marker"
}

# ---- probe ----------------------------------------------------------------
# Prints: verdict=exposed|gated|ERROR marker=N name=N rc=N out=N err=N
run_probe() {
  home="$1"; proj="$2"; name="$3"; marker="$4"; epath="$5"; enab="$6"
  mkdir -p "$home"
  if [ -n "$epath" ] || [ -n "$enab" ]; then
    if [ -z "$epath" ] || [ -z "$enab" ]; then
      echo "verdict=ERROR reason=--entry-path and --enabled must be given together" ; return 2
    fi
    # enabled is NOT optional to codex: an entry lacking it is a hard start
    # failure (t504 finding F1). Always emit it.
    printf '# t502 probe cell\n[[skills.config]]\npath = "%s"\nenabled = %s\n' "$epath" "$enab" > "$home/config.toml"
  elif [ ! -f "$home/config.toml" ]; then
    printf '# t502 probe cell: no skills.config\n' > "$home/config.toml"
  fi

  cap="$home/.probe-capture"; mkdir -p "$cap"
  ( cd "$proj" && CODEX_HOME="$home" codex debug prompt-input "hi" ) \
    > "$cap/out" 2> "$cap/err"
  rc=$?
  # Counts are taken from files on disk, never from a pipe.
  m=$(grep -c "$marker" "$cap/out" || true)
  n=$(grep -c "$name" "$cap/out" || true)
  ob=$(wc -c < "$cap/out" | tr -d ' '); eb=$(wc -c < "$cap/err" | tr -d ' ')

  if [ "$rc" -ne 0 ] || [ "$eb" -ne 0 ]; then
    verdict=ERROR
  elif [ "$m" -ge 1 ]; then
    verdict=exposed
  else
    verdict=gated
  fi
  printf 'verdict=%s marker=%s name=%s rc=%s out=%s err=%s capture=%s\n' \
    "$verdict" "$m" "$n" "$rc" "$ob" "$eb" "$cap/out"
  PROBE_VERDICT="$verdict"
  [ "$verdict" = "ERROR" ] && return 1
  return 0
}

notouch_before() { shasum -a 256 "$HOME/.codex/config.toml" 2>&1; }

case "$SUB" in
  fixture)
    [ -n "$ROOT" ] && [ -n "$MIRROR" ] || { usage >&2; exit 2; }
    [ -e "$ROOT" ] && { echo "--root must not exist: $ROOT" >&2; exit 2; }
    mkdir -p "$ROOT" || exit 2
    build_fixture "$ROOT" "$MIRROR" "$SKILL_NAME" "$MARKER"
    ;;

  probe)
    [ -n "$CODEX_HOME_ARG" ] && [ -n "$PROJECT" ] || { usage >&2; exit 2; }
    HB="$(notouch_before)"
    run_probe "$CODEX_HOME_ARG" "$PROJECT" "$SKILL_NAME" "$MARKER" "$ENTRY_PATH" "$ENABLED"
    st=$?
    HA="$(notouch_before)"
    echo "notouch_before=$HB"
    echo "notouch_after=$HA"
    [ "$HB" = "$HA" ] && echo "notouch=IDENTICAL" || { echo "notouch=DIFFERS"; exit 4; }
    [ $st -ne 0 ] && exit $st
    if [ -n "$EXPECT" ]; then
      if [ "$PROBE_VERDICT" = "$EXPECT" ]; then
        echo "expect=$EXPECT result=MATCH"
      else
        echo "expect=$EXPECT result=MISMATCH (got $PROBE_VERDICT)"; exit 3
      fi
    fi
    ;;

  selftest)
    # An instrument that only ever reports one direction proves nothing. This
    # asserts BOTH: the bound notation gates, and a control still exposes.
    [ -n "$ROOT" ] || ROOT="$(mktemp -d /tmp/t502-selftest-XXXXXXXX)/lab"
    HB="$(notouch_before)"
    fail=0
    for mode in symlink copy; do
      fx="$ROOT/$mode"; rm -rf "$fx"; mkdir -p "$fx"
      eval "$(build_fixture "$fx" "$mode" "$SKILL_NAME" "$MARKER" | grep -E '^(PROJECT|LITERAL_PATH|RESOLVED_PATH)=')"
      # control: no entry -> must be exposed
      run_probe "$fx/home-ctl" "$PROJECT" "$SKILL_NAME" "$MARKER" "" "" > "$fx/ctl.txt"
      c=$(sed 's/ .*//;s/verdict=//' "$fx/ctl.txt")
      # bound notation (literal mirror path) + enabled=false -> must be gated
      run_probe "$fx/home-lit" "$PROJECT" "$SKILL_NAME" "$MARKER" "$LITERAL_PATH" false > "$fx/lit.txt"
      l=$(sed 's/ .*//;s/verdict=//' "$fx/lit.txt")
      printf '%s\tcontrol=%s\tliteral_false=%s\n' "$mode" "$c" "$l"
      [ "$c" = "exposed" ] || { echo "SELFTEST FAIL: $mode control not exposed (adoption gate)"; fail=1; }
      [ "$l" = "gated" ]   || { echo "SELFTEST FAIL: $mode literal notation did not gate"; fail=1; }
    done
    # the asymmetry the matrix turns on: resolved notation is INERT under copy
    fx="$ROOT/copy"
    PROJECT="$fx/proj"; RESOLVED_PATH="$fx/proj/.claude/skills/$SKILL_NAME/SKILL.md"
    run_probe "$fx/home-res" "$PROJECT" "$SKILL_NAME" "$MARKER" "$RESOLVED_PATH" false > "$fx/res.txt"
    r=$(sed 's/ .*//;s/verdict=//' "$fx/res.txt")
    printf 'copy\tresolved_false=%s\t(expected exposed — the inert notation)\n' "$r"
    [ "$r" = "exposed" ] || { echo "SELFTEST FAIL: resolved notation gated under copy — matrix asymmetry gone"; fail=1; }
    HA="$(notouch_before)"
    echo "notouch_before=$HB"
    echo "notouch_after=$HA"
    [ "$HB" = "$HA" ] || { echo "SELFTEST FAIL: real ~/.codex/config.toml changed"; fail=1; }
    echo "ROOT=$ROOT"
    [ $fail -eq 0 ] && echo "SELFTEST PASS" || { echo "SELFTEST FAIL"; exit 1; }
    ;;

  *) usage >&2; exit 2 ;;
esac
