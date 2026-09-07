#!/bin/bash
# t496 trigger campaign — stage 2. Reuses the stage-1 EXP_ROOT (isolated home
# with all-event loggers already verified by P0). Runs the four trigger
# attempts: collaboration delegation, permission request, compaction,
# interrupt. Serial. Everything lands under EXP_ROOT (in /tmp).
set -u
EXP_ROOT="${1:?usage: campaign2.sh EXP_ROOT}"
[ -d "$EXP_ROOT/home" ] || { echo "EXP_ROOT/home missing"; exit 1; }

run_exec() {
  name="$1"; shift
  ( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
      --skip-git-repo-check --dangerously-bypass-hook-trust "$@" ) \
    > "$EXP_ROOT/runs/$name.jsonl" 2> "$EXP_ROOT/runs/$name.err" < /dev/null
  echo $? > "$EXP_ROOT/runs/$name.rc"
}

# ---- R-collab: SubagentStart / SubagentStop ---------------------------------
run_exec collab "Use the collaboration tool to spawn a subagent whose task is to compute 2+2. Wait for it to finish, then reply with the subagent's answer and nothing else."

# ---- R-perm: PermissionRequest ----------------------------------------------
# read-only sandbox + a write command => approval request path. NO
# bypass-approvals flag on this run (the request itself is the trigger).
( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
    --skip-git-repo-check --sandbox read-only --dangerously-bypass-hook-trust \
    "Use the shell to create a file named t496perm.txt in the current directory containing exactly: PERMOK. Then read it back and reply with its content." ) \
  > "$EXP_ROOT/runs/perm.jsonl" 2> "$EXP_ROOT/runs/perm.err" < /dev/null
echo $? > "$EXP_ROOT/runs/perm.rc"

# ---- R-compact: PreCompact / PostCompact -------------------------------------
# Very long stdin payload to push context toward auto-compaction. The --json
# stream is the precondition evidence: if compaction never happens, the
# verdict is trigger-not-achieved, not not-fired.
{
  echo "Below is filler text you must ignore entirely. Reply with exactly: T496COMPACTOK"
  for i in $(seq 1 4000); do
    echo "Paragraph $i: The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. How vexingly quick daft zebras jump. Sphinx of black quartz judge my vow."
  done
} > "$EXP_ROOT/compact-input.txt"
wc -c "$EXP_ROOT/compact-input.txt" > "$EXP_ROOT/runs/compact.size"
( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
    --skip-git-repo-check --dangerously-bypass-hook-trust "Reply with exactly: T496COMPACTOK" ) \
  > "$EXP_ROOT/runs/compact.jsonl" 2> "$EXP_ROOT/runs/compact.err" \
  < "$EXP_ROOT/compact-input.txt"
echo $? > "$EXP_ROOT/runs/compact.rc"

# ---- R-interrupt: Interrupt ---------------------------------------------------
( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 120 codex exec --json \
    --skip-git-repo-check --dangerously-bypass-hook-trust \
    "Count slowly from 1 to 100, one number per line, thinking about each number carefully before answering." ) \
  > "$EXP_ROOT/runs/interrupt.jsonl" 2> "$EXP_ROOT/runs/interrupt.err" < /dev/null &
IPID=$!
sleep 8
kill -INT $IPID 2>/dev/null
echo "sent-sigint-to=$IPID" > "$EXP_ROOT/runs/interrupt.signal"
wait $IPID
echo $? > "$EXP_ROOT/runs/interrupt.rc"

# ---- census --------------------------------------------------------------------
{
  echo "--- run exit codes ---"
  for r in collab perm compact interrupt; do
    echo "$r rc=$(cat "$EXP_ROOT/runs/$r.rc")"
  done
  echo "--- captures ---"
  for f in "$EXP_ROOT"/captures/*.jsonl; do
    printf '%s\t%d lines\t%d bytes\n' "$(basename "$f" .jsonl)" "$(wc -l < "$f" | tr -d ' ')" "$(wc -c < "$f" | tr -d ' ')"
  done
  echo "--- collab: tool calls in stream ---"
  grep -o '"type":"item.[a-z]*","item":{"id":"[^"]*","type":"[a-z_]*"' "$EXP_ROOT/runs/collab.jsonl" | head -10
  echo "--- perm: stream tail ---"
  tail -4 "$EXP_ROOT/runs/perm.jsonl"
  echo "--- compact: stream markers ---"
  grep -c -i 'compact' "$EXP_ROOT/runs/compact.jsonl" || true
  grep -i -m 3 -o '.\{0,80\}compact.\{0,80\}' "$EXP_ROOT/runs/compact.jsonl" || true
  tail -2 "$EXP_ROOT/runs/compact.jsonl"
  echo "--- interrupt: stream tail + signal ---"
  cat "$EXP_ROOT/runs/interrupt.signal"
  tail -3 "$EXP_ROOT/runs/interrupt.jsonl"
} > "$EXP_ROOT/campaign2-summary.txt" 2>&1

cat "$EXP_ROOT/campaign2-summary.txt"
