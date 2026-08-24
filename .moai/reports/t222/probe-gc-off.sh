#!/bin/bash
# t222 probe: does `maintenance.auto=false` + `gc.auto=0` suppress the detached
# `git maintenance run --auto --detach` that every commit otherwise spawns?
set -e
d=$(mktemp -d /tmp/t222-gcoff.XXXXXX)
cd "$d"
export GIT_TRACE=1
{
  git init -b main .
  git config user.email a@b.c
  git config user.name t
  git config gc.auto 0
  git config maintenance.auto false
  echo x > f
  git add .
  git commit -m c1
  echo y > f
  git add .
  git commit -m c2
} 2>&1 | grep -ac "maintenance run --auto" || echo "0 detached maintenance spawns"
