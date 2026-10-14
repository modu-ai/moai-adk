#!/bin/sh
# Scratch probe for plan section B3 step 4a (SPEC-LOCAL-MAIN-FLOW-001, finding I3-1).
# Runs the planned primary merge command on three ignored-content shapes in new scratch repos.
# Usage: sh probe-shapes.sh   (takes no arguments; the scratch root is created by mktemp)
ROOT=$(mktemp -d) || exit 9

probe() {
  S="$1"; PAT="$2"; CARD_PATH="$3"; ADDFLAG="$4"; LOCAL_PATH="$5"; MARK="$6"
  D="$ROOT/shape-$S"
  mkdir "$D" || return 9
  cd "$D" || return 9
  git init -q -b main
  git config user.email probe@example.invalid
  git config user.name probe
  printf '%s\n' "$PAT" > .gitignore
  git add .gitignore
  git commit -q -m base
  git checkout -q -b card
  mkdir -p "$(dirname "$CARD_PATH")"
  printf 'card\n' > "$CARD_PATH"
  git add $ADDFLAG "$CARD_PATH"
  git commit -q -m card
  git checkout -q main
  mkdir -p "$(dirname "$LOCAL_PATH")"
  printf '%s\n' "$MARK" > "$LOCAL_PATH"

  echo "== shape $S (ignore '$PAT'; local '$LOCAL_PATH' = '$MARK'; card writes '$CARD_PATH')"
  echo "core.ignorecase=$(git config core.ignorecase)"
  echo "card branch has: $(git ls-tree -r --name-only card | tr '\n' ' ')"
  echo "T (paths the merge adds): $(git diff --name-only main card | tr '\n' ' ')"
  echo "IGN (ignored paths present): $(git ls-files --others --ignored --exclude-standard | tr '\n' ' ')"
  I=""
  for t in $(git diff --name-only main card); do
    for i in $(git ls-files --others --ignored --exclude-standard); do
      [ "$t" = "$i" ] && I="$I $t"
    done
  done
  echo "I (exact intersection, B5):${I:- (empty)}"
  echo "status before: [$(git status --porcelain | tr '\n' ' ')]"
  echo "local content before: [$(grep -rl "$MARK" . --exclude-dir=.git | tr '\n' ' ')]"

  git -c merge.autoStash=false merge --no-ff --no-overwrite-ignore -q -m probe card > "$ROOT/merge-$S.out" 2>&1
  RC=$?
  echo "merge exit: $RC"
  echo "merge output: $(tr '\n' ' ' < "$ROOT/merge-$S.out")"
  echo "status after: [$(git status --porcelain | tr '\n' ' ')]"
  echo "local content after: [$(grep -rl "$MARK" . --exclude-dir=.git | tr '\n' ' ')]"
  echo "tree after (excluding .git): $(find . -path ./.git -prune -o -print | sort | tr '\n' ' ')"
  echo "card content now at '$CARD_PATH': $(cat "$CARD_PATH" 2>/dev/null || echo '(not a readable file)')"
  cd "$ROOT" || return 9
}

probe A 'runtime' 'runtime/x' '-f' 'runtime' 'local-A'
probe B 'runtime/' 'runtime' '' 'runtime/secret' 'local-B'
probe C 'foo.txt' 'Foo.txt' '-f' 'foo.txt' 'local-C'
