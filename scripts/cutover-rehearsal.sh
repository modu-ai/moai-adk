#!/usr/bin/env bash
# cutover-rehearsal.sh - rehearse the convergence of the retiring branch
# into the target base branch, and its rollback, in a SCRATCH clone
# (SPEC-GITHUB-FLOW-DEFAULT-001 M6, AC-GFD-018, REQ-GFD-018, design D-8).
#
# It touches no real remote. The scratch working repository is created here by a
# local `git clone` of --source, its own remotes are local paths only (the script
# refuses to continue otherwise), and the "origin" it pushes to is a bare
# repository created inside --workdir. git is run with GIT_ALLOW_PROTOCOL=file and
# without global/system configuration, so a network transport or a URL rewrite rule
# cannot be reached even by accident. It never calls gh.
#
# The script names no branch of its own and carries no default: the caller supplies
# the retiring branch with --retiring-branch and the target base branch
# with --target-branch, and without either the script stops with a usage error
# (exit 2) naming the missing argument.
#
# What it rehearses (all inside the scratch clone):
#   1. start       the retiring and target branches at their source tips; the refs
#                  are written to pre-merge-refs.txt BEFORE anything moves
#                  (REQ-GFD-018: the way back is recorded before the merge)
#   2. absorb      the retiring branch absorbs origin/<target> with a merge commit
#                  (runbook step 4)
#   3. convergence the target branch merges the retiring branch with a merge commit,
#                  never a squash (runbook step 5b)
#   4. identity    tree(target) == tree(retiring tip before the merge), and the
#                  retiring tip is an ancestor of the target (a squash fails here)
#   5. revert      the rollback that works against a protected target: a revert of
#                  the merge commit; tree(target) returns to the starting target tree
#   6. rollback    the recorded pre-merge refs are restored and the scratch origin
#                  is returned to them; refs identical to the start (scratch-only:
#                  the real target forbids force-push, so the real rollback is 5)
#
# Usage:
#   scripts/cutover-rehearsal.sh --source <repo> --workdir <new-or-empty-dir>
#                                --retiring-branch <name> --target-branch <name>
#                                [--retiring-rev <rev>] [--target-rev <rev>]
#   scripts/cutover-rehearsal.sh --check-remotes-only <repo>
#
#   --source   a local repository carrying refs/remotes/origin/<retiring-branch> and
#              refs/remotes/origin/<target-branch> (or give --retiring-rev and
#              --target-rev as revisions reachable from its refs)
#   --retiring-branch  REQUIRED: the retiring branch
#   --target-branch    REQUIRED: the target base branch
#   --retiring-rev     default refs/source/origin/<retiring-branch> (the source's
#                      remote-tracking ref of the retiring branch)
#   --target-rev       default refs/source/origin/<target-branch>
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
RETIRING=""
TARGET=""
RETIRING_REV=""
TARGET_REV=""
CHECK_ONLY=""

while [ $# -gt 0 ]; do
  case "$1" in
    --source) [ $# -ge 2 ] || die_usage "--source needs a value"; SOURCE="$2"; shift 2 ;;
    --workdir) [ $# -ge 2 ] || die_usage "--workdir needs a value"; WORKDIR="$2"; shift 2 ;;
    --retiring-branch) [ $# -ge 2 ] || die_usage "--retiring-branch needs a value"; RETIRING="$2"; shift 2 ;;
    --target-branch) [ $# -ge 2 ] || die_usage "--target-branch needs a value"; TARGET="$2"; shift 2 ;;
    --retiring-rev) [ $# -ge 2 ] || die_usage "--retiring-rev needs a value"; RETIRING_REV="$2"; shift 2 ;;
    --target-rev) [ $# -ge 2 ] || die_usage "--target-rev needs a value"; TARGET_REV="$2"; shift 2 ;;
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
[ -n "$RETIRING" ] || die_usage "--retiring-branch is required (the retiring branch; this script has no default name, try --help)"
[ -n "$TARGET" ] || die_usage "--target-branch is required (the target base branch; this script has no default name, try --help)"
# The names reach refs and merge commands: a malformed or option-like name is a usage error.
git check-ref-format --branch "$RETIRING" >/dev/null 2>&1 || die_usage "--retiring-branch is not a valid branch name: $RETIRING"
git check-ref-format --branch "$TARGET" >/dev/null 2>&1 || die_usage "--target-branch is not a valid branch name: $TARGET"
[ "$RETIRING" != "$TARGET" ] || die_usage "--retiring-branch and --target-branch name the same branch: $RETIRING"
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
if [ -z "$RETIRING_REV" ] && ! git -C "$SRC" rev-parse --verify --quiet "refs/remotes/origin/$RETIRING" >/dev/null; then
  die_usage "the source has no refs/remotes/origin/$RETIRING (pass --retiring-rev <rev>)"
fi
if [ -z "$TARGET_REV" ] && ! git -C "$SRC" rev-parse --verify --quiet "refs/remotes/origin/$TARGET" >/dev/null; then
  die_usage "the source has no refs/remotes/origin/$TARGET (pass --target-rev <rev>)"
fi
RETIRING_REV=${RETIRING_REV:-refs/source/origin/$RETIRING}
TARGET_REV=${TARGET_REV:-refs/source/origin/$TARGET}

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
echo "branches retiring=$RETIRING target=$TARGET (supplied by the caller; the script names neither)"
echo "remotes:"
g remote -v | sed 's/^/  /'
echo "tools: $(git --version) (GIT_ALLOW_PROTOCOL=file, no global/system git configuration, no gh)"

# Pull the source's refs into a private namespace: a local-path fetch only.
must git -C "$REPO" fetch --quiet --no-tags source '+refs/remotes/origin/*:refs/source/origin/*' '+refs/heads/*:refs/source/heads/*'
RET_SHA=$(g rev-parse --verify --quiet "${RETIRING_REV}^{commit}") || die_usage "cannot resolve --retiring-rev $RETIRING_REV in the scratch clone"
TGT_SHA=$(g rev-parse --verify --quiet "${TARGET_REV}^{commit}") || die_usage "cannot resolve --target-rev $TARGET_REV in the scratch clone"

must git -C "$REPO" checkout --quiet -B "$RETIRING" "$RET_SHA"
must git -C "$REPO" branch --force "$TARGET" "$TGT_SHA"
must git -C "$REPO" push --quiet origin "$RETIRING" "$TARGET"

# The way back, recorded before anything moves.
{
  echo "$TARGET $TGT_SHA"
  echo "$RETIRING $RET_SHA"
} >"$RECORD"
START_TGT_TREE=$(tree_of "$TGT_SHA")
START_RET_TREE=$(tree_of "$RET_SHA")
echo "start retiring=$RET_SHA target=$TGT_SHA retiring-tree=$START_RET_TREE target-tree=$START_TGT_TREE"
echo "divergence (retiring...target, left-right): $(g rev-list --count --left-right "$RET_SHA...$TGT_SHA" | tr '\t' ' ')"

# ---- runbook step 4: the retiring branch absorbs origin/<target> with a merge commit
g checkout --quiet "$RETIRING"
if ! g merge --no-ff --quiet -m "Merge origin/$TARGET into $RETIRING (cutover rehearsal)" "origin/$TARGET" >/dev/null 2>&1; then
  g merge --abort >/dev/null 2>&1
  fail "the retiring branch could not absorb origin/$TARGET without a conflict"
fi
ABSORB=$(g rev-parse HEAD)
if [ "$ABSORB" = "$RET_SHA" ]; then
  echo "absorb retiring=$ABSORB parents=none tree=$(tree_of "$ABSORB") (the target is already an ancestor of the retiring branch; nothing to absorb)"
else
  [ "$(parents_of "$ABSORB")" = "$RET_SHA,$TGT_SHA" ] || fail "the absorbing commit is not a merge of the retiring and target branches"
  echo "absorb retiring=$ABSORB parents=$(parents_of "$ABSORB") tree=$(tree_of "$ABSORB")"
fi
must git -C "$REPO" push --quiet origin "$RETIRING"
RET_TIP=$(g rev-parse "$RETIRING")

# ---- runbook step 5b: the target merges the retiring branch with a merge commit (never a squash)
g checkout --quiet "$TARGET"
if ! g merge --no-ff --quiet -m "Merge $RETIRING into $TARGET (cutover rehearsal)" "$RETIRING" >/dev/null 2>&1; then
  g merge --abort >/dev/null 2>&1
  fail "the target branch could not merge the retiring branch"
fi
CONV=$(g rev-parse HEAD)
[ "$(parents_of "$CONV")" = "$TGT_SHA,$RET_TIP" ] || fail "the convergence commit is not a merge commit of the target and retiring branches"
echo "convergence target=$CONV parents=$(parents_of "$CONV") tree=$(tree_of "$CONV")"
must git -C "$REPO" push --quiet origin "$TARGET"

# ---- runbook step 6: tree identity and ancestry
TGT_TREE=$(tree_of "$TARGET")
RET_TIP_TREE=$(tree_of "$RET_TIP")
if [ "$TGT_TREE" = "$RET_TIP_TREE" ]; then eq=yes; else eq=no; fi
echo "tree-identity target-tree=$TGT_TREE retiring-tree=$RET_TIP_TREE equal=$eq"
[ "$eq" = yes ] || fail "tree identity: the target tree differs from the retiring tip tree"
if g merge-base --is-ancestor "$RET_TIP" "$TARGET"; then anc=yes; else anc=no; fi
echo "ancestry retiring-tip-is-ancestor-of-target=$anc"
[ "$anc" = yes ] || fail "ancestry: the retiring tip is not an ancestor of the target"

# ---- rollback path 1 (the real one): revert the merge commit on the target
if ! g revert --no-edit -m 1 "$CONV" >/dev/null 2>&1; then
  g revert --abort >/dev/null 2>&1
  fail "reverting the convergence merge failed"
fi
REVERT=$(g rev-parse HEAD)
REVERT_TREE=$(tree_of "$TARGET")
if [ "$REVERT_TREE" = "$START_TGT_TREE" ]; then eq=yes; else eq=no; fi
echo "revert target=$REVERT tree=$REVERT_TREE start-target-tree=$START_TGT_TREE equal=$eq"
[ "$eq" = yes ] || fail "revert: the target tree after the revert differs from the starting target tree"
must git -C "$REPO" push --quiet origin "$TARGET"

# ---- rollback path 2 (scratch only): restore the recorded refs, read from the record
g checkout --quiet --detach
[ "$(g remote get-url origin)" = "$BARE" ] || fail "origin is not the scratch bare repository; refusing to force-update it"
assert_local_remotes "$REPO"
while read -r name sha; do
  [ -n "$name" ] || continue
  must git -C "$REPO" update-ref "refs/heads/$name" "$sha"
  must git -C "$REPO" push --quiet --force origin "refs/heads/$name:refs/heads/$name"
done <"$RECORD"
RB_TGT=$(g rev-parse "refs/heads/$TARGET")
RB_RET=$(g rev-parse "refs/heads/$RETIRING")
OR_TGT=$(git -C "$BARE" rev-parse "refs/heads/$TARGET")
OR_RET=$(git -C "$BARE" rev-parse "refs/heads/$RETIRING")
if [ "$RB_TGT" = "$TGT_SHA" ] && [ "$RB_RET" = "$RET_SHA" ] && [ "$OR_TGT" = "$TGT_SHA" ] && [ "$OR_RET" = "$RET_SHA" ]; then ident=yes; else ident=no; fi
echo "rollback target=$RB_TGT retiring=$RB_RET origin-target=$OR_TGT origin-retiring=$OR_RET identical-to-start=$ident"
[ "$ident" = yes ] || fail "rollback: the restored refs differ from the starting refs"

echo "REHEARSAL-EVIDENCE end"
echo "REHEARSAL-RESULT PASS"
exit 0
