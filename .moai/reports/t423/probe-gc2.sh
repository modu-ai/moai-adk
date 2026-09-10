#!/bin/bash
# t423 mechanism probe v2: two loose objects under objects/17/ with gc.auto=1
# (threshold ceil(1/256)=1, trigger is count > 1 => need 2). If the detached
# maintenance gc fires, .git/objects/pack gains files AFTER `git commit`
# returns — while the Go test's t.TempDir RemoveAll is already deleting.
set -u
T=$(mktemp -d /tmp/t423-probe2.XXXXXX)
cd "$T"
git init -q .
git config gc.auto 1
git config user.email p@e.com && git config user.name P

found=()
for i in $(seq 1 400000); do
  h=$(printf "content-%d" $i | git hash-object --stdin)
  case "$h" in 17*) found+=("$i");; esac
  [ ${#found[@]} -ge 2 ] && break
done
echo "seeds=${found[*]} (2 blob hashes starting with 17)"
for n in "${found[@]}"; do printf "content-%s" "$n" > "seed_$n.txt"; done
git add -A
git commit -q -m base
echo "objects_17_after_base=$(ls .git/objects/17 2>/dev/null | wc -l | tr -d ' ')"

# Second commit — spawns `git maintenance run --auto --quiet --detach`.
printf "tail\n" >> "seed_${found[0]}.txt"
git add -A
git commit -q -m churn
echo "commit returned; now polling pack dir for 5s..."
for j in $(seq 1 50); do
  n=$(ls .git/objects/pack/ 2>/dev/null | wc -l | tr -d ' ')
  if [ "$n" != "0" ]; then
    echo "PACK WRITTEN after ~$((j*100))ms:"
    ls -la .git/objects/pack/
    exit 0
  fi
  sleep 0.1
done
echo "no pack written within 5s"
ls .git/objects/17 2>/dev/null | head -3
