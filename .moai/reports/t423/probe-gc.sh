#!/bin/bash
# t423 mechanism probe: with gc.auto=1, does `git commit` in a small fixture
# repo spawn a DETACHED gc that writes .git/objects/pack after the parent
# returns? That is the exact race TestGitDiffNameCount_Predicate hits when
# t.TempDir RemoveAll runs cleanup.
set -u
T=$(mktemp -d /tmp/t423-probe.XXXXXX)
cd "$T"
git init -q .
git config gc.auto 1
git config user.email p@e.com && git config user.name P

# Get a loose object into objects/17/: find content whose blob hash starts
# with "17" (too_many_loose_objects counts only the objects/17 fan-out dir,
# threshold ceil(gc.auto/256); gc.auto=1 => a single object triggers).
found=""
for i in $(seq 1 200000); do
  h=$(printf "content-%d" $i | git hash-object --stdin)
  case "$h" in 17*) found=$i; break;; esac
done
echo "seed_index=$found (blob hash starts with 17)"
printf "content-%s" "$found" > seed.txt
git add -A
git commit -q -m base
echo "objects_17_count_after_base_commit: $(ls .git/objects/17 2>/dev/null | wc -l | tr -d ' ')"

# Second commit — `git commit` triggers `gc --auto` internally.
printf "more\n" >> seed.txt
git add -A
( for j in $(seq 1 100); do pgrep -x git -l >> "$T/pgrep.log" 2>/dev/null || true; sleep 0.01; done ) &
watcher=$!
git commit -q -m churn
wait $watcher 2>/dev/null || true
echo "--- git processes seen during second commit:"
sort -u "$T/pgrep.log" 2>/dev/null | head -8
echo "--- pack dir 1s after commit returned:"
sleep 1
ls -la .git/objects/pack/ 2>/dev/null
echo "PROBE_TDIR=$T"
