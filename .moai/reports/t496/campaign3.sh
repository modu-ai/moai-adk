#!/bin/bash
# t496 trigger campaign — stage 3 (refinements).
#  - interrupt: SIGINT delivered to codex itself via `timeout -s INT 8`.
#  - perm2: workspace-write sandbox + write outside workspace + approval_policy
#    override, so an approval request is actually raised.
#  - compact2: ~2x input size for the auto-compaction precondition.
set -u
EXP_ROOT="${1:?usage: campaign3.sh EXP_ROOT}"
[ -d "$EXP_ROOT/home" ] || { echo "EXP_ROOT/home missing"; exit 1; }

# ---- R-interrupt2 -------------------------------------------------------------
( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout -s INT 8 codex exec --json \
    --skip-git-repo-check --dangerously-bypass-hook-trust \
    "Think carefully, then count slowly from 1 to 200, one number per line." ) \
  > "$EXP_ROOT/runs/interrupt2.jsonl" 2> "$EXP_ROOT/runs/interrupt2.err" < /dev/null
echo $? > "$EXP_ROOT/runs/interrupt2.rc"

# ---- R-perm2 -------------------------------------------------------------------
( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
    --skip-git-repo-check --sandbox workspace-write -c approval_policy='"on-request"' \
    --dangerously-bypass-hook-trust \
    "Use the shell to create a file at /tmp/t496-outside-perm.txt containing exactly: PERMOK2. Reply with whether it succeeded." ) \
  > "$EXP_ROOT/runs/perm2.jsonl" 2> "$EXP_ROOT/runs/perm2.err" < /dev/null
echo $? > "$EXP_ROOT/runs/perm2.rc"

# ---- R-compact2 -----------------------------------------------------------------
{
  echo "Below is filler text you must ignore entirely. Reply with exactly: T496COMPACT2OK"
  for i in $(seq 1 9000); do
    echo "Paragraph $i: The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. How vexingly quick daft zebras jump. Sphinx of black quartz judge my vow."
  done
} > "$EXP_ROOT/compact2-input.txt"
wc -c "$EXP_ROOT/compact2-input.txt" > "$EXP_ROOT/runs/compact2.size"
( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
    --skip-git-repo-check --dangerously-bypass-hook-trust "Reply with exactly: T496COMPACT2OK" ) \
  > "$EXP_ROOT/runs/compact2.jsonl" 2> "$EXP_ROOT/runs/compact2.err" \
  < "$EXP_ROOT/compact2-input.txt"
echo $? > "$EXP_ROOT/runs/compact2.rc"

# ---- census --------------------------------------------------------------------
{
  echo "--- rc ---"
  for r in interrupt2 perm2 compact2; do echo "$r rc=$(cat "$EXP_ROOT/runs/$r.rc")"; done
  echo "--- interrupt2: stream head+tail ---"
  head -3 "$EXP_ROOT/runs/interrupt2.jsonl"
  tail -2 "$EXP_ROOT/runs/interrupt2.jsonl"
  echo "--- perm2: approval-ish markers ---"
  grep -c -i 'approval\|permission' "$EXP_ROOT/runs/perm2.jsonl" || true
  grep -i -o '.\{0,60\}approval.\{0,60\}' "$EXP_ROOT/runs/perm2.jsonl" | head -5 || true
  tail -3 "$EXP_ROOT/runs/perm2.jsonl"
  echo "--- compact2: usage + compact markers ---"
  grep -o '"input_tokens":[0-9]*' "$EXP_ROOT/runs/compact2.jsonl" | tail -1
  grep -c -i 'compact' "$EXP_ROOT/runs/compact2.jsonl" || true
  grep -i -o '.\{0,80\}compact.\{0,80\}' "$EXP_ROOT/runs/compact2.jsonl" | head -5 || true
  tail -2 "$EXP_ROOT/runs/compact2.jsonl"
} > "$EXP_ROOT/campaign3-summary.txt" 2>&1

cat "$EXP_ROOT/campaign3-summary.txt"
