#!/bin/bash
# t485 passive index.lock watcher (card t485, branch WT-indexlock-origin)
#
# Observation only: this script runs ZERO git commands, so it can never create
# or contend for an index.lock itself. It polls lock existence with stat(1),
# and on a sighting captures lsof(1) on the lock path plus a full ps(1)
# snapshot, attributing the live holder PID at that instant.
#
# Self-bounding (cleanup-guaranteed): exits cleanly after MAX_POLLS iterations;
# additionally run it under an external `timeout` as a backstop.
#
# Usage: watch-indexlock.sh <outdir> [max_polls] [interval_seconds]
#   default max_polls=4800, interval=0.25  ->  ~20 minutes per run

set -u
OUT="${1:?usage: watch-indexlock.sh <outdir> [max_polls] [interval_s]}"
MAX_POLLS="${2:-4800}"
INTERVAL="${3:-0.25}"
PRIMARY_LOCK="/Users/goos/MoAI/moai-adk-go/.git/index.lock"
WT_LOCK_GLOB="/Users/goos/MoAI/moai-adk-go/.git/worktrees/*/index.lock"

mkdir -p "$OUT"
EVENTS="$OUT/events.jsonl"
LOG="$OUT/watch.log"
echo "$(date -u +%FT%TZ) watcher start pid=$$ polls_max=$MAX_POLLS interval=$INTERVAL" >> "$LOG"

polls=0
hits=0
last_ps=0

while [ "$polls" -lt "$MAX_POLLS" ]; do
  polls=$((polls + 1))

  # One stat call per poll covers the primary checkout and every linked
  # worktree at once. Prints "size|mtime_epoch|path" lines for existing locks.
  found=$(stat -f '%z|%m|%N' "$PRIMARY_LOCK" $WT_LOCK_GLOB 2>/dev/null)

  if [ -n "$found" ]; then
    ts=$(date -u +%FT%TZ)
    epoch=$(date +%s)
    hits=$((hits + 1))

    psfile=""
    if [ $((epoch - last_ps)) -ge 8 ]; then
      psfile="$OUT/ps-$epoch.txt"
      ps -Ao pid,ppid,lstart,etime,command > "$psfile" 2>&1
      last_ps=$epoch
    fi

    while IFS= read -r line; do
      [ -z "$line" ] && continue
      size="${line%%|*}"; rest="${line#*|}"; mtime="${rest%%|*}"; path="${rest#*|}"
      tag=$(basename "$(dirname "$path")")
      lsoffile="$OUT/lsof-$epoch-$tag.txt"
      lsof "$path" > "$lsoffile" 2>&1
      holders=""
      if head -1 "$lsoffile" 2>/dev/null | grep -q "^COMMAND"; then
        # Real sighting: the lock fd is open. The holder lives only tens of
        # ms, so the co-temporal process table must be captured WHILE lsof
        # scans (concurrent background ps), not after it returns. Only git-
        # capable processes are kept (a non-git process cannot hold a git
        # index lock), which keeps each snapshot a few KB.
        eps="$OUT/eps-$epoch-$tag.txt"
        ps -Ao pid,ppid,lstart,command | awk 'NR==1 || /git|statusline|moai/' > "$eps" 2>&1 &
        psjob=$!
        for hpid in $(tail -n +2 "$lsoffile" | awk '{print $2}' | sort -u); do
          holders="$holders$hpid,"
          hf="$OUT/holder-$epoch-$tag-$hpid.txt"
          for attempt in 1 2 3; do
            ps -p "$hpid" -o pid=,ppid=,lstart=,command= >> "$hf" 2>&1
            if [ -s "$hf" ]; then
              pp=$(ps -p "$hpid" -o ppid= 2>/dev/null | tr -d ' ')
              if [ -n "$pp" ] && [ "$pp" != "1" ] && [ "$pp" != "0" ]; then
                ps -p "$pp" -o pid=,ppid=,lstart=,command= >> "$hf" 2>&1
              fi
              break
            fi
            sleep 0.05
          done
        done
        wait $psjob
        holders="${holders%,}"
      fi
      printf '{"ts":"%s","epoch":%s,"path":"%s","size":%s,"mtime_epoch":%s,"holders":"%s","lsof_file":"%s","ps_file":"%s"}\n' \
        "$ts" "$epoch" "$path" "$size" "$mtime" "$holders" "$lsoffile" "$psfile" >> "$EVENTS"
    done <<< "$found"

    echo "$ts HIT polls=$polls hits=$hits" >> "$LOG"
  fi

  if [ $((polls % 480)) -eq 0 ]; then
    echo "$(date -u +%FT%TZ) alive polls=$polls hits=$hits" >> "$LOG"
  fi

  sleep "$INTERVAL"
done

echo "$(date -u +%FT%TZ) watcher end polls=$polls hits=$hits" >> "$LOG"
