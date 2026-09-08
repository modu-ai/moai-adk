#!/usr/bin/env bash
# t459 R2: after a cap rotation, is the archived generation reachable to the
# real drain? Runs the actual .claude/skills/hns-lsel-curator/drain.sh against
# a /tmp fixture. Never touches this project's .moai/state/lsel/.
set -uo pipefail

DRAIN="$1"          # absolute path to drain.sh
FIX="$(mktemp -d)"
trap 'rm -rf "$FIX"' EXIT

INBOX="$FIX/lessons-inbox.jsonl"
STATE="$FIX/lsel-state"
mkdir -p "$STATE"

# 4 archived stubs (what a rotation moved out of reach) ...
for i in 1 2 3 4; do
  printf '{"timestamp":"2026-09-03T00:00:0%sZ","event_key":"test_fail:archived:","summary":"archived stub %s","source":"test:archived","v":1}\n' "$i" "$i"
done > "$INBOX.1"

# ... and 2 live stubs written after the rotation.
for i in 1 2; do
  printf '{"timestamp":"2026-09-03T01:00:0%sZ","event_key":"test_fail:live:","summary":"live stub %s","source":"test:live","v":1}\n' "$i" "$i"
done > "$INBOX"

echo "fixture: live=$(wc -l < "$INBOX") stubs, archived(.1)=$(wc -l < "$INBOX.1") stubs"
echo "--- drain.sh --inbox <live> ---"
"$DRAIN" --inbox "$INBOX" --state-dir "$STATE"
echo "--- drain exit=$? ---"

echo "--- offset after drain ---"
cat "$STATE/drain-offset.json" 2>/dev/null || echo "(no offset file)"

echo "--- did any archived stub reach clusters.json? ---"
if [ -f "$STATE/clusters.json" ]; then
  if grep -q 'archived' "$STATE/clusters.json"; then
    echo "REACHED: archived stubs appear in clusters.json"
  else
    echo "UNREACHABLE: no archived stub in clusters.json"
  fi
  echo "clusters.json:"; cat "$STATE/clusters.json"
else
  echo "(no clusters.json emitted)"
fi

echo "--- does drain.sh reference any archive generation? ---"
grep -c 'jsonl\.[0-9]' "$DRAIN"
