#!/usr/bin/env bash
# baseline-latency.sh — interleaved paired A/B of the wall time of `moai hook pre-tool`
# (process start-up included) between two builds, for AC-SIPZ-007.
#
# Usage (from the repository root):
#   bash .moai/specs/SPEC-SELF-IMPROVE-PROTECTED-ZONE-001/evidence/baseline-latency.sh <binary-A> <binary-B> [N]
#
# Why paired and interleaved: two runs of the SAME build moved the identity-write median by
# 129 ms on this machine with no code change (evidence/latency-base-e497f6936.txt), so a bound
# on two separate runs is a coin flip. Here every iteration times A then B back to back, and the
# statistic is the MEDIAN OF THE PER-ITERATION DIFFERENCES (B - A), which cancels slow drift.
# The instrument is also run A against A (the control) — if the control's median difference is not
# well under the bound, the instrument cannot resolve the bound and the verdict is INCONCLUSIVE.
#
# Pass rule (stated, not tuned): the A/B median difference <= BOUND_MS for every class, and the
# A/A control median difference <= CONTROL_MS. BOUND_MS = 250 is 2.5% of the 10 s hook timeout in
# the shipped settings; CONTROL_MS = 100.
#
# The three classes use paths chosen so that none changes zone membership between builds:
#   non-identity-write   internal/hook/pre_tool.go, no agent_type      (guard must not apply)
#   identity-outside     docs/a.md, agent_type harness-learner          (outside every zone entry)
#   identity-inside      .claude/hooks/moai/x.sh, agent_type harness-learner (legacy-denied before and after)
# Output: one row per class: <class> <pair> n=<N> median_diff_ms=<d> verdict=<PASS|FAIL|INCONCLUSIVE>
# Exit: 0 all PASS; 1 any FAIL or INCONCLUSIVE; 2 usage.
set -u
A="${1:?usage: baseline-latency.sh <binary-A> <binary-B> [N]}"
B="${2:?usage: baseline-latency.sh <binary-A> <binary-B> [N]}"
N="${3:-21}"
BOUND_MS=250
CONTROL_MS=100
ROOT="$(cd "$(dirname "$0")/../../../.." && pwd -P)"
cd "$ROOT" || exit 2

ms_now() { python3 -c 'import time; print(int(time.time()*1000))'; }

one() { # <binary> <agent_type|-> <path>  -> elapsed ms on stdout
  local bin="$1" agent="$2" path="$3" agentjson="" t0 t1
  [ "$agent" != "-" ] && agentjson=",\"agent_type\":\"$agent\""
  t0="$(ms_now)"
  printf '%s' "{\"session_id\":\"lat\",\"hook_event_name\":\"PreToolUse\",\"cwd\":\"$ROOT\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"$path\",\"content\":\"x\"}$agentjson}" \
    | env CLAUDE_PROJECT_DIR="$ROOT" "$bin" hook pre-tool >/dev/null 2>&1
  t1="$(ms_now)"
  echo $((t1 - t0))
}

median_diff() { # <binary-X> <binary-Y> <agent> <path>
  local x="$1" y="$2" agent="$3" path="$4" i d diffs=""
  for i in $(seq 1 "$N"); do
    d=$(( $(one "$y" "$agent" "$path") - $(one "$x" "$agent" "$path") ))
    diffs="$diffs $d"
  done
  printf '%s\n' "$diffs" | tr ' ' '\n' | grep -v '^$' | sort -n | awk '{a[NR]=$1} END {print a[int((NR+1)/2)]}'
}

rc=0
for spec in "non-identity-write|-|internal/hook/pre_tool.go" \
            "identity-outside|harness-learner|docs/a.md" \
            "identity-inside|harness-learner|.claude/hooks/moai/x.sh"; do
  class="${spec%%|*}"; rest="${spec#*|}"; agent="${rest%%|*}"; path="${rest#*|}"
  ctl="$(median_diff "$A" "$A" "$agent" "$path")"
  ab="$(median_diff "$A" "$B" "$agent" "$path")"
  if [ "$ctl" -gt "$CONTROL_MS" ] || [ "$ctl" -lt "-$CONTROL_MS" ]; then verdict=INCONCLUSIVE
  elif [ "$ab" -le "$BOUND_MS" ]; then verdict=PASS; else verdict=FAIL; fi
  printf '%s\tA/A n=%s median_diff_ms=%s\tA/B n=%s median_diff_ms=%s\tbound=%s verdict=%s\n' "$class" "$N" "$ctl" "$N" "$ab" "$BOUND_MS" "$verdict"
  [ "$verdict" = PASS ] || rc=1
done
exit $rc
