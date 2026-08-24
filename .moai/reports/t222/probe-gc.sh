#!/bin/bash
# t222 probe: does the kanban specFixtureRepo git sequence spawn a detached
# auto-gc / maintenance process that could write .git/objects after the test?
set -e
d=$(mktemp -d /tmp/t222-gcprobe.XXXXXX)
cd "$d"
export GIT_TRACE=1
{
  git init -b main .
  git config user.email a@b.c
  git config user.name t
  mkdir -p .moai/specs/S && echo x > .moai/specs/S/spec.md
  git add .
  git commit -m c1
  for b in plan/S feat/S-run sync/S; do
    git branch "$b"
    git checkout "$b"
    echo "$b" > .moai/specs/S/spec.md
    git add .
    git commit -m "$b"
  done
  git checkout main
} 2>&1 | grep -aiE "gc|maintenance|commit-graph|run_command" | sort | uniq -c | sort -rn | head -20
echo "--- objects files:"
find .git/objects -type f | head
echo "--- gc.auto config:"
git config --get gc.auto || echo "(unset -> default 6700)"
echo "--- git version:"
git --version
echo "--- probe dir: $d"
