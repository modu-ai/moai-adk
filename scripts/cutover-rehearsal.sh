#!/usr/bin/env bash
# cutover-rehearsal.sh - rehearse the develop -> main convergence and its rollback
# in a SCRATCH clone (SPEC-GITHUB-FLOW-DEFAULT-001 M6, AC-GFD-018, REQ-GFD-018,
# design D-8).
#
# It touches no real remote. The scratch working repository is created here by a
# local `git clone` of --source, its own remotes are local paths only (the script
# refuses to continue otherwise), and the "origin" it pushes to is a bare
# repository created inside --workdir. git is run with GIT_ALLOW_PROTOCOL=file and
# without global/system configuration, so a network transport or a URL rewrite rule
# cannot be reached even by accident. It never calls gh.
#
# What it rehearses (all inside the scratch clone):
#   1. start       develop and main at their source tips; the refs are written to
#                  pre-merge-refs.txt BEFORE anything moves (REQ-GFD-018: the way
#                  back is recorded before the merge)
#   2. absorb      develop absorbs origin/main with a merge commit (runbook step 4)
#   3. convergence main merges develop with a merge commit, never a squash
#                  (runbook step 5b)
#   4. identity    tree(main) == tree(develop tip before the merge), and the
#                  develop tip is an ancestor of main (a squash fails here)
#   5. revert      the rollback that works against a protected main: a revert of
#                  the merge commit; tree(main) returns to the starting main tree
#   6. rollback    the recorded pre-merge refs are restored and the scratch origin
#                  is returned to them; refs identical to the start (scratch-only:
#                  the real main forbids force-push, so the real rollback is 5)
#
# Usage:
#   scripts/cutover-rehearsal.sh --source <repo> --workdir <new-or-empty-dir>
#                                [--develop <rev>] [--main <rev>]
#   scripts/cutover-rehearsal.sh --check-remotes-only <repo>
#
#   --source   a local repository carrying refs/remotes/origin/{develop,main}
#              (or give --develop/--main as revisions reachable from its refs)
#   --develop  default refs/source/origin/develop (the source's origin/develop)
#   --main     default refs/source/origin/main
#   --check-remotes-only  only run the local-path remote check on <repo>
#
# Exit: 0 PASS, 1 a rehearsal assertion failed, 2 usage/environment error,
#       3 a non-local remote was found (refused).
set -u

usage() {
  sed -n '2,/^set -u/p' "$0" | sed '$d' | sed 's/^# \{0,1\}//'
}

# Hermetic and network-proof: only the file transport, no user configuration.
export GIT_ALLOW_PROTOCOL=file
export GIT_CONFIG_NOSYSTEM=1
export GIT_CONFIG_GLOBAL=/dev/null
export GIT_TERMINAL_PROMPT=0
export GIT_MERGE_AUTOEDIT=no
export GIT_EDITOR=true

die_usage() { echo "cutover-rehearsal: $*" >&2; exit 2; }

# is_local_path <url>: an absolute filesystem path of an existing directory. No
# scheme, no host:path form.
is_local_path() {
  case "$1" in
    *://*) return 1 ;;
    /*) [ -d "$1" ] ;;
    *) return 1 ;;
  esac
}

# assert_local_remotes <repo>: every configured remote URL (fetch and push) is a
# local path and no URL rewrite rule exists. Exits 3 on the first violation.
assert_local_remotes() {
  local repo="$1" line key url
  if [ -n "$(git -C "$repo" config --get-regexp '^url\..*\.(insteadof|pushinsteadof)$' 2>/dev/null)" ]; then
    echo "REFUSED: $repo has a url.*.insteadOf rewrite rule; remotes could resolve to a non-local address" >&2
    exit 3
  fi
  while IFS= read -r line; do
    [ -n "$line" ] || continue
    key=${line%% *}
    url=${line#* }
    if ! is_local_path "$url"; then
      echo "REFUSED: non-local remote ${key#remote.} = $url (the rehearsal only runs against local-path remotes)" >&2
      exit 3
    fi
  done < <(git -C "$repo" config --get-regexp '^remote\..*\.(url|pushurl)$' 2>/dev/null)
}

SOURCE=""
WORKDIR=""
DEVELOP_REV=""
MAIN_REV=""
CHECK_ONLY=""

while [ $# -gt 0 ]; do
  case "$1" in
    --source) [ $# -ge 2 ] || die_usage "--source needs a value"; SOURCE="$2"; shift 2 ;;
    --workdir) [ $# -ge 2 ] || die_usage "--workdir needs a value"; WORKDIR="$2"; shift 2 ;;
    --develop) [ $# -ge 2 ] || die_usage "--develop needs a value"; DEVELOP_REV="$2"; shift 2 ;;
    --main) [ $# -ge 2 ] || die_usage "--main needs a value"; MAIN_REV="$2"; shift 2 ;;
    --check-remotes-only) [ $# -ge 2 ] || die_usage "--check-remotes-only needs a repository"; CHECK_ONLY="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) die_usage "unknown argument: $1 (try --help)" ;;
  esac
done

if [ -n "$CHECK_ONLY" ]; then
  git -C "$CHECK_ONLY" rev-parse --git-dir >/dev/null 2>&1 || die_usage "not a git repository: $CHECK_ONLY"
  assert_local_remotes "$CHECK_ONLY"
  echo "remotes of $CHECK_ONLY are all local paths"
  exit 0
fi

[ -n "$SOURCE" ] || die_usage "--source is required (try --help)"
[ -n "$WORKDIR" ] || die_usage "--workdir is required (try --help)"
git -C "$SOURCE" rev-parse --git-dir >/dev/null 2>&1 || die_usage "--source is not a git repository: $SOURCE"
SRC=$(cd "$SOURCE" && pwd -P) || die_usage "cannot resolve --source: $SOURCE"

if [ -e "$WORKDIR" ]; then
  [ -d "$WORKDIR" ] || die_usage "--workdir exists and is not a directory: $WORKDIR"
  [ -z "$(ls -A "$WORKDIR" 2>/dev/null)" ] || die_usage "--workdir is not empty: $WORKDIR (the rehearsal needs a fresh scratch directory)"
fi
# Resolve the absolute scratch path BEFORE creating anything, so a refusal never
# leaves a directory behind (least of all inside the source).
WPARENT=$(cd "$(dirname "$WORKDIR")" 2>/dev/null && pwd -P) || die_usage "the parent of --workdir does not exist: $WORKDIR"
WORK="$WPARENT/$(basename "$WORKDIR")"
case "$WORK/" in "$SRC"/*) die_usage "--workdir must be outside --source ($WORK is inside $SRC)" ;; esac
case "$SRC/" in "$WORK"/*) die_usage "--source must be outside --workdir" ;; esac

# The source must carry what the defaults read; name what is missing.
if [ -z "$DEVELOP_REV" ] && ! git -C "$SRC" rev-parse --verify --quiet refs/remotes/origin/develop >/dev/null; then
  die_usage "the source has no refs/remotes/origin/develop (pass --develop <rev>)"
fi
if [ -z "$MAIN_REV" ] && ! git -C "$SRC" rev-parse --verify --quiet refs/remotes/origin/main >/dev/null; then
  die_usage "the source has no refs/remotes/origin/main (pass --main <rev>)"
fi
DEVELOP_REV=${DEVELOP_REV:-refs/source/origin/develop}
MAIN_REV=${MAIN_REV:-refs/source/origin/main}

mkdir -p "$WORK" || die_usage "cannot create --workdir: $WORKDIR"
REPO="$WORK/work"
BARE="$WORK/origin.git"
RECORD="$WORK/pre-merge-refs.txt"
g() { git -C "$REPO" "$@"; }

fail() {
  echo "REHEARSAL-ASSERT-FAILED $1"
  echo "REHEARSAL-RESULT FAIL $1"
  exit 1
}
must() { "$@" >/dev/null 2>&1 </dev/null || fail "command failed: $*"; }
tree_of() { g rev-parse --verify --quiet "$1^{tree}"; }
parents_of() { g rev-list --parents -n1 "$1" | cut -d' ' -f2- | tr ' ' ','; }

# ---- scratch setup: a local clone, a scratch bare origin, local-path remotes only
git clone --quiet --no-checkout --no-hardlinks -o source "$SRC" "$REPO" >/dev/null 2>&1 || fail "git clone of the local source failed"
must git init --quiet --bare "$BARE"
must git -C "$REPO" remote add origin "$BARE"
assert_local_remotes "$REPO"
g config user.name "cutover-rehearsal"
g config user.email "cutover-rehearsal@example.invalid"
g config commit.gpgsign false
g config gc.auto 0
g config core.hooksPath /dev/null

echo "REHEARSAL-EVIDENCE begin"
echo "remotes:"
g remote -v | sed 's/^/  /'
echo "tools: $(git --version) (GIT_ALLOW_PROTOCOL=file, no global/system git configuration, no gh)"

# Pull the source's refs into a private namespace: a local-path fetch only.
must git -C "$REPO" fetch --quiet --no-tags source '+refs/remotes/origin/*:refs/source/origin/*' '+refs/heads/*:refs/source/heads/*'
DEV_SHA=$(g rev-parse --verify --quiet "${DEVELOP_REV}^{commit}") || die_usage "cannot resolve --develop $DEVELOP_REV in the scratch clone"
MAIN_SHA=$(g rev-parse --verify --quiet "${MAIN_REV}^{commit}") || die_usage "cannot resolve --main $MAIN_REV in the scratch clone"

must git -C "$REPO" checkout --quiet -B develop "$DEV_SHA"
must git -C "$REPO" branch --force main "$MAIN_SHA"
must git -C "$REPO" push --quiet origin develop main

# The way back, recorded before anything moves.
{
  echo "main $MAIN_SHA"
  echo "develop $DEV_SHA"
} >"$RECORD"
START_MAIN_TREE=$(tree_of "$MAIN_SHA")
START_DEV_TREE=$(tree_of "$DEV_SHA")
echo "start develop=$DEV_SHA main=$MAIN_SHA develop-tree=$START_DEV_TREE main-tree=$START_MAIN_TREE"
echo "divergence (develop...main, left-right): $(g rev-list --count --left-right "$DEV_SHA...$MAIN_SHA" | tr '\t' ' ')"

# ---- runbook step 4: develop absorbs origin/main with a merge commit
g checkout --quiet develop
if ! g merge --no-ff --quiet -m "Merge origin/main into develop (cutover rehearsal)" origin/main >/dev/null 2>&1; then
  g merge --abort >/dev/null 2>&1
  fail "develop could not absorb origin/main without a conflict"
fi
ABSORB=$(g rev-parse HEAD)
if [ "$ABSORB" = "$DEV_SHA" ]; then
  echo "absorb develop=$ABSORB parents=none tree=$(tree_of "$ABSORB") (main is already an ancestor of develop; nothing to absorb)"
else
  [ "$(parents_of "$ABSORB")" = "$DEV_SHA,$MAIN_SHA" ] || fail "the absorbing commit is not a merge of develop and main"
  echo "absorb develop=$ABSORB parents=$(parents_of "$ABSORB") tree=$(tree_of "$ABSORB")"
fi
must git -C "$REPO" push --quiet origin develop
DEV_TIP=$(g rev-parse develop)

# ---- runbook step 5b: main merges develop with a merge commit (never a squash)
g checkout --quiet main
if ! g merge --no-ff --quiet -m "Merge develop into main (cutover rehearsal)" develop >/dev/null 2>&1; then
  g merge --abort >/dev/null 2>&1
  fail "main could not merge develop"
fi
CONV=$(g rev-parse HEAD)
[ "$(parents_of "$CONV")" = "$MAIN_SHA,$DEV_TIP" ] || fail "the convergence commit is not a merge commit of main and develop"
echo "convergence main=$CONV parents=$(parents_of "$CONV") tree=$(tree_of "$CONV")"
must git -C "$REPO" push --quiet origin main

# ---- runbook step 6: tree identity and ancestry
MAIN_TREE=$(tree_of main)
DEV_TIP_TREE=$(tree_of "$DEV_TIP")
if [ "$MAIN_TREE" = "$DEV_TIP_TREE" ]; then eq=yes; else eq=no; fi
echo "tree-identity main-tree=$MAIN_TREE develop-tree=$DEV_TIP_TREE equal=$eq"
[ "$eq" = yes ] || fail "tree identity: main tree differs from the develop tip tree"
if g merge-base --is-ancestor "$DEV_TIP" main; then anc=yes; else anc=no; fi
echo "ancestry develop-tip-is-ancestor-of-main=$anc"
[ "$anc" = yes ] || fail "ancestry: the develop tip is not an ancestor of main"

# ---- rollback path 1 (the real one): revert the merge commit on main
if ! g revert --no-edit -m 1 "$CONV" >/dev/null 2>&1; then
  g revert --abort >/dev/null 2>&1
  fail "reverting the convergence merge failed"
fi
REVERT=$(g rev-parse HEAD)
REVERT_TREE=$(tree_of main)
if [ "$REVERT_TREE" = "$START_MAIN_TREE" ]; then eq=yes; else eq=no; fi
echo "revert main=$REVERT tree=$REVERT_TREE start-main-tree=$START_MAIN_TREE equal=$eq"
[ "$eq" = yes ] || fail "revert: main tree after the revert differs from the starting main tree"
must git -C "$REPO" push --quiet origin main

# ---- rollback path 2 (scratch only): restore the recorded refs, read from the record
g checkout --quiet --detach
[ "$(g remote get-url origin)" = "$BARE" ] || fail "origin is not the scratch bare repository; refusing to force-update it"
assert_local_remotes "$REPO"
while read -r name sha; do
  [ -n "$name" ] || continue
  must git -C "$REPO" update-ref "refs/heads/$name" "$sha"
  must git -C "$REPO" push --quiet --force origin "refs/heads/$name:refs/heads/$name"
done <"$RECORD"
RB_MAIN=$(g rev-parse refs/heads/main)
RB_DEV=$(g rev-parse refs/heads/develop)
OR_MAIN=$(git -C "$BARE" rev-parse refs/heads/main)
OR_DEV=$(git -C "$BARE" rev-parse refs/heads/develop)
if [ "$RB_MAIN" = "$MAIN_SHA" ] && [ "$RB_DEV" = "$DEV_SHA" ] && [ "$OR_MAIN" = "$MAIN_SHA" ] && [ "$OR_DEV" = "$DEV_SHA" ]; then ident=yes; else ident=no; fi
echo "rollback main=$RB_MAIN develop=$RB_DEV origin-main=$OR_MAIN origin-develop=$OR_DEV identical-to-start=$ident"
[ "$ident" = yes ] || fail "rollback: the restored refs differ from the starting refs"

echo "REHEARSAL-EVIDENCE end"
echo "REHEARSAL-RESULT PASS"
exit 0
