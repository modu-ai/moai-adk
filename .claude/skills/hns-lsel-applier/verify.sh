#!/usr/bin/env bash
# verify.sh — SPEC-LSEL-LOCAL-EVOLUTION-001 M4 mechanical VERIFY core.
#
# The post-apply VERIFY stage for the LSEL APPLY engine. After lsel-apply.sh
# commits a proposal, verify.sh runs the proposal's verify_command with a
# timeout-retry-once policy and, on a second non-timeout failure, auto-fires
# `git revert lsel-<proposal-id>` and marks the proposal's feedback_*.md
# verified:false.
#
# DESIGN SPLIT (REQ-LSEL-013 / AC-LSEL-015): VERIFY has TWO layers.
#   (a) MECHANICAL — this script: runs verify_command + timeout-retry-once +
#       auto-revert + ledger `verified` marker. Bash-testable.
#   (b) MANDATORY `/moai gate` SUPERSET — invoked MODEL-SIDE (the applier SKILL
#       instructs the orchestrator/model to run `/moai gate` after the mechanical
#       apply+verify). A bash hook cannot invoke a Claude Code slash command
#       directly; `/moai gate` (lint+format+type+test) is documented as MANDATORY
#       in hns-lsel-applier/SKILL.md. Proposer-authored verify alone is circular
#       (report §11 mustFix B#6); the gate superset is the independent check.
#
# RETRY POLICY (2 attempts total):
#   attempt 1:
#     SUCCESS            → verified:true, stop.
#     TIMEOUT-class      → attempt 2 (flaky tolerance; clause 2).
#     NON-TIMEOUT-FAIL   → attempt 2 (allows a "second" non-timeout failure per
#                          clause 3 before revert fires).
#   attempt 2:
#     SUCCESS            → verified:true.
#     TIMEOUT            → verified:false + git revert.
#     NON-TIMEOUT-FAIL   → verified:false + git revert (clause 3: the second
#                          non-timeout failure auto-fires revert).
#
# Exit codes: 0 = verified true OR verified-false-but-revert-clean; 1 = error.
# The `verified` ledger marker is the load-bearing signal, not the exit code.
set -euo pipefail

PROPOSAL_DIR=""
REPO_ROOT=""
FEEDBACK_FILE=""
TIMEOUT_SECS="${VERIFY_TIMEOUT_SECS:-30}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --proposal-dir) PROPOSAL_DIR="$2"; shift 2 ;;
    --repo-root) REPO_ROOT="$2"; shift 2 ;;
    --feedback-file) FEEDBACK_FILE="$2"; shift 2 ;;
    --timeout) TIMEOUT_SECS="$2"; shift 2 ;;
    *) echo "verify: unknown arg: $1" >&2; exit 1 ;;
  esac
done

resolve_root() {
  if [[ -n "${LSEL_PROJECT_ROOT:-}" ]]; then echo "$LSEL_PROJECT_ROOT"; return; fi
  if [[ -n "${CLAUDE_PROJECT_DIR:-}" && -d "${CLAUDE_PROJECT_DIR:-}" ]]; then echo "$CLAUDE_PROJECT_DIR"; return; fi
  git rev-parse --show-toplevel 2>/dev/null || pwd
}
[[ -z "$REPO_ROOT" ]] && REPO_ROOT="$(resolve_root)"
LEDGER="$REPO_ROOT/.moai/state/lsel/apply-ledger.jsonl"

if [[ -z "$PROPOSAL_DIR" || ! -d "$PROPOSAL_DIR" ]]; then
  echo "verify: --proposal-dir required (pointing at the proposal triple)" >&2
  exit 1
fi
PROPOSAL_MD="$PROPOSAL_DIR/proposal.md"
if [[ ! -f "$PROPOSAL_MD" ]]; then
  echo "verify: proposal.md not found at $PROPOSAL_MD" >&2
  exit 1
fi

# --- portable YAML-frontmatter field extraction (single-line value only) ---
fm_field() {
  # $1 key, $2 file → first frontmatter value (line: "key: value")
  awk -v k="$1" '
    /^---$/ { c++; next }
    c==1 && $0 ~ "^"k":" { sub("^"k":[[:space:]]*",""); print; exit }
  ' "$2"
}

PROPOSAL_ID="$(fm_field proposal_id "$PROPOSAL_MD")"
VERIFY_CMD="$(fm_field verify_command "$PROPOSAL_MD")"
if [[ -z "$PROPOSAL_ID" ]]; then
  echo "verify: proposal.md frontmatter missing proposal_id" >&2
  exit 1
fi
if [[ -z "$VERIFY_CMD" ]]; then
  echo "verify: proposal.md frontmatter missing verify_command" >&2
  exit 1
fi

mkdir -p "$(dirname "$LEDGER")"
TS="$(date -u +%FT%TZ)"

# --- timeout wrapper: returns 124 on timeout, else the command's exit code ---
TIMEOUT_BIN=""
if command -v timeout >/dev/null 2>&1; then TIMEOUT_BIN=timeout
elif command -v gtimeout >/dev/null 2>&1; then TIMEOUT_BIN=gtimeout
fi
run_with_timeout() {
  # $1 = command string. Sets globals LAST_RC and LAST_CLASS.
  if [[ -n "$TIMEOUT_BIN" ]]; then
    set +e
    # Wrap the command in the timeout binary so a hung verify_command is killed
    # (exit 124) rather than blocking the loop. eval runs in the repo-root cwd.
    ( cd "$REPO_ROOT" && eval "$TIMEOUT_BIN $TIMEOUT_SECS $1" ) >/tmp/lsel-verify.out 2>&1
    LAST_RC=$?
    set -e
  else
    # Portable fallback: background + deadline poll.
    local dead pid
    rm -f /tmp/lsel-verify.exit
    ( cd "$REPO_ROOT" && eval "$1" >/tmp/lsel-verify.out 2>&1; echo $? >/tmp/lsel-verify.exit ) &
    pid=$!
    dead=$(( $(date +%s) + TIMEOUT_SECS ))
    while [[ $(date +%s) -lt $dead ]]; do
      kill -0 "$pid" 2>/dev/null || break
      sleep 0.2
    done
    if kill -0 "$pid" 2>/dev/null; then
      kill -TERM "$pid" 2>/dev/null; sleep 0.3; kill -KILL "$pid" 2>/dev/null
      wait "$pid" 2>/dev/null
      LAST_RC=124
    else
      wait "$pid" 2>/dev/null
      LAST_RC="$(cat /tmp/lsel-verify.exit 2>/dev/null || echo 1)"
    fi
  fi
  if   [[ "$LAST_RC" -eq 0 ]];   then LAST_CLASS=success
  elif [[ "$LAST_RC" -eq 124 ]]; then LAST_CLASS=timeout
  else                                LAST_CLASS=non-timeout-fail
  fi
}

# --- 2-attempt retry policy ---
OUTCOME=""
ATTEMPTS=0
run_with_timeout "$VERIFY_CMD"; ATTEMPTS=$((ATTEMPTS+1))
if [[ "$LAST_CLASS" == "success" ]]; then
  OUTCOME=true
else
  # First attempt failed (timeout OR non-timeout): retry exactly once.
  run_with_timeout "$VERIFY_CMD"; ATTEMPTS=$((ATTEMPTS+1))
  if [[ "$LAST_CLASS" == "success" ]]; then
    OUTCOME=true
  else
    OUTCOME=false
  fi
fi

append_ledger() {
  # $1 verified  $2 attempts  $3 last_class
  printf '{"stage":"verify","proposal_id":"%s","verified":%s,"attempts":%s,"last_class":"%s","ts":"%s"}\n' \
    "$PROPOSAL_ID" "$1" "$2" "$3" >> "$LEDGER"
}

if [[ "$OUTCOME" == "true" ]]; then
  append_ledger true "$ATTEMPTS" "$LAST_CLASS"
  echo "verify: PASS proposal=$PROPOSAL_ID verified=true attempts=$ATTEMPTS"
  exit 0
fi

# --- OUTCOME=false: auto git revert + verified:false marker ---
# Locate the lsel-<proposal-id> commit via the apply ledger (fallback: git log grep).
COMMIT_SHA="$(grep -m1 "\"proposal_id\":\"$PROPOSAL_ID\"" "$LEDGER" 2>/dev/null \
  | sed 's/.*"commit_sha":"\([^"]*\)".*/\1/' || true)"
if [[ -z "$COMMIT_SHA" || "$COMMIT_SHA" == "pending-backfill" ]]; then
  COMMIT_SHA="$(git -C "$REPO_ROOT" log --grep "lsel-$PROPOSAL_ID" --format='%H' | head -1)"
fi
REVERTED=0
if [[ -n "$COMMIT_SHA" ]]; then
  if git -C "$REPO_ROOT" revert --no-edit "$COMMIT_SHA" >/tmp/lsel-verify-revert.out 2>&1; then
    REVERTED=1
  else
    echo "verify: WARN git revert $COMMIT_SHA failed for proposal=$PROPOSAL_ID" >&2
  fi
else
  echo "verify: WARN no lsel-$PROPOSAL_ID commit found to revert" >&2
fi

append_ledger false "$ATTEMPTS" "$LAST_CLASS"

# Mark the proposal's feedback_*.md verified:false (audit trail for retrieval).
if [[ -n "$FEEDBACK_FILE" ]]; then
  mkdir -p "$(dirname "$FEEDBACK_FILE")"
  if [[ -f "$FEEDBACK_FILE" ]]; then
    printf '\nverified: false\n' >> "$FEEDBACK_FILE"
  else
    printf -- '---\nverified: false\n---\n' > "$FEEDBACK_FILE"
  fi
fi

echo "verify: FAIL proposal=$PROPOSAL_ID verified=false attempts=$ATTEMPTS reverted=$REVERTED"
exit 0
