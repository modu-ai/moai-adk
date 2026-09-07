#!/bin/bash
# t496 trigger campaign — stage 4 (final refinement round).
#  - compact3: filler sized to just under the measured 1,048,576-char cap
#    (compact1 703K chars -> 192,872 input tokens; the cap allows ~1.4x more,
#    i.e. ~274K tokens at the same density).
#  - perm3: read-only sandbox + approval_policy=on-request + in-workspace write
#    (perm1 read-only+default-policy auto-failed without a request; perm2
#    workspace-write let /tmp through without one).
set -u
EXP_ROOT="${1:?usage: campaign4.sh EXP_ROOT}"
[ -d "$EXP_ROOT/home" ] || { echo "EXP_ROOT/home missing"; exit 1; }

# ---- R-perm3 -------------------------------------------------------------------
( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
    --skip-git-repo-check --sandbox read-only -c approval_policy='"on-request"' \
    --dangerously-bypass-hook-trust \
    "Use the shell to create a file named t496perm3.txt in the current directory containing exactly: PERMOK3. Reply with whether it succeeded." ) \
  > "$EXP_ROOT/runs/perm3.jsonl" 2> "$EXP_ROOT/runs/perm3.err" < /dev/null
echo $? > "$EXP_ROOT/runs/perm3.rc"

# ---- R-compact3 -----------------------------------------------------------------
{
  echo "Below is filler text you must ignore entirely. Reply with exactly: T496COMPACT3OK"
  for i in $(seq 1 5630); do
    echo "Paragraph $i: The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. How vexingly quick daft zebras jump. Sphinx of black quartz judge my vow."
  done
} > "$EXP_ROOT/compact3-input.txt"
wc -c "$EXP_ROOT/compact3-input.txt" > "$EXP_ROOT/runs/compact3.size"
( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
    --skip-git-repo-check --dangerously-bypass-hook-trust "Reply with exactly: T496COMPACT3OK" ) \
  > "$EXP_ROOT/runs/compact3.jsonl" 2> "$EXP_ROOT/runs/compact3.err" \
  < "$EXP_ROOT/compact3-input.txt"
echo $? > "$EXP_ROOT/runs/compact3.rc"

# ---- census --------------------------------------------------------------------
{
  echo "--- rc ---"
  for r in perm3 compact3; do echo "$r rc=$(cat "$EXP_ROOT/runs/$r.rc")"; done
  echo "--- perm3: approval markers ---"
  grep -c -i 'approval\|permission' "$EXP_ROOT/runs/perm3.jsonl" || true
  tail -3 "$EXP_ROOT/runs/perm3.jsonl"
  echo "--- compact3: usage + compact markers ---"
  grep -o '"input_tokens":[0-9]*' "$EXP_ROOT/runs/compact3.jsonl" | tail -1
  grep -c -i 'compact' "$EXP_ROOT/runs/compact3.jsonl" || true
  grep -i -o '.\{0,100\}compact.\{0,100\}' "$EXP_ROOT/runs/compact3.jsonl" | grep -v T496COMPACT3OK | head -5 || true
  tail -2 "$EXP_ROOT/runs/compact3.jsonl"
} > "$EXP_ROOT/campaign4-summary.txt" 2>&1

cat "$EXP_ROOT/campaign4-summary.txt"
