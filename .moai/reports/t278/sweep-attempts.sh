#!/bin/zsh
# t278 M1 attempt-aware baseline sweep — SPEC-CI-FLAKE-SERIES-001
# REQ-CFS-009: every failed attempt of every ci.yml run is inspected, because
# rerun-green hides attempt-1 failures from `gh run list --status failure`.
#
# NOTE on the API: the REST path /actions/runs/<id>/attempts returns 404 on
# this repo (measured 2026-08-27 — 537/537 runs), so attempts are enumerated
# from the run's `run_attempt` field (the highest attempt number, 1-based)
# and each attempt's conclusion is read via `gh run view --attempt N --json`.
#
# v2 inputs (from sweep v1, already on disk):
#   /tmp/t278-runs.tsv            databaseId \t conclusion \t createdAt \t event
#   /tmp/t278-sweep/<id>.jobs     job name \t conclusion (latest attempt)
# v2 outputs:
#   /tmp/t278-sweep/<id>.att<N>   conclusion of attempt N (only when N>1 or failed)
#   /tmp/t278-sweep/<id>.attempt<N>.hits  grep of the 3 flaky test names
#   /tmp/t278-sweep/<id>.done2
set -u

process_run() {
  local id=$1
  local dir=/tmp/t278-sweep
  [ -f "$dir/$id.done2" ] && return 0

  local maxatt
  maxatt=$(gh api "repos/modu-ai/moai-adk/actions/runs/$id" --jq '.run_attempt' 2>/dev/null)
  case "$maxatt" in
    ''|*[!0-9]*) maxatt=1 ;;
  esac

  local att=1
  while [ "$att" -le "$maxatt" ]; do
    local c
    if [ "$maxatt" -eq 1 ]; then
      c=$(awk -F'\t' -v id="$id" '$1==id {print $2}' /tmp/t278-runs.tsv)
    else
      c=$(gh run view "$id" --attempt "$att" --json conclusion --jq '.conclusion' 2>/dev/null)
      printf '%s\n' "$c" > "$dir/$id.att$att"
    fi
    case "$c" in
      success|skipped|Success|""|null|cancelled) ;;
      *)
        gh run view "$id" --attempt "$att" --log-failed 2>/dev/null \
          | grep -e 'TestConcurrentSendPoll' -e 'TestAssertPairedHealthyEndToEnd' -e 'TestConfigChange_RT005ReloadIntegration' \
          > "$dir/$id.attempt${att}.hits"
        ;;
    esac
    att=$((att + 1))
  done
  touch "$dir/$id.done2"
}

while IFS=$'\t' read -r id concl created event; do
  process_run "$id" &
  while [ "$(jobs -r | wc -l)" -ge 8 ]; do sleep 0.3; done
done < /tmp/t278-runs.tsv
wait
echo "SWEEP_V2_DONE $(ls /tmp/t278-sweep/*.done2 2>/dev/null | wc -l | tr -d ' ') runs processed"
