#!/usr/bin/env bash
# 08-C T2 — Codex host-loading E2E (AC-08C-4, Issue #2014). The lane does NOT run this:
# it needs the operator's Codex API key and writes nothing outside the directories it
# creates. It may run from any directory (the package is found next to this script),
# so the operator can run it from a tree outside the checkout. Receipts and the raw
# outputs they cite land in $RECEIPT_DIR. Credential VALUES are never written to any
# file — the receipt records only that CODEX_API_KEY was set.
#
# Isolation: `codex exec` runs with HOME and CODEX_HOME pointing into the temporary
# work dir, so the operator's ~/.codex and ~/.agents are neither read for
# configuration nor written. Authentication therefore comes from CODEX_API_KEY only
# (learn.chatgpt.com/docs/config-file/environment-variables, 2026-11-09: CODEX_HOME
# "Sets the root for Codex state, including config, auth, …"; CODEX_API_KEY
# "Provides an API key to a non-interactive Codex process").
#
# Usage: export CODEX_API_KEY (presence is checked, the value is never read into any
# file), then run: tests/e2e/host-codex.sh [RECEIPT_DIR]
# Result fields follow schemas/e2e-receipt.schema.json (REQ-08C-01).
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RECEIPT_DIR="${1:-./e2e-receipts}"
mkdir -p "$RECEIPT_DIR"
RECEIPT_DIR="$(cd "$RECEIPT_DIR" && pwd)"
TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
WORK="$(mktemp -d /tmp/e2e-codex-XXXXXX)"
trap 'rm -rf "$WORK"' EXIT
SKILL_DIGEST="$(shasum -a 256 "$REPO/skills/moai/SKILL.md" | awk '{print $1}')"
CLI_SHA="$(command -v moai-cli >/dev/null 2>&1 && shasum -a 256 "$(command -v moai-cli)" | awk '{print $1}' || echo absent)"
# Pre-step: project the skill into the temporary project's .agents/skills and give
# every codex call, the version probe included, an isolated HOME / CODEX_HOME — the
# probe alone may leave state under $CODEX_HOME (#2014 agent-8 N4).
mkdir -p "$WORK/project/.agents/skills" "$WORK/home/.codex"
HOST_VERSION="$(HOME="$WORK/home" CODEX_HOME="$WORK/home/.codex" codex --version 2>/dev/null < /dev/null | head -1 || echo unknown)"
cp -R "$REPO/skills/moai" "$WORK/project/.agents/skills/moai"

receipt() { # receipt <name> <result> <detail>
  printf '{"ac":"%s","os":"%s","arch":"%s","host":"codex","host_exact_version":"%s","mode":"local","skill_digest":"%s","moai_cli_sha":"%s","result":"%s","ran_at":"%s","detail":"%s"}\n' \
    "$1" "$(uname -s)" "$(uname -m)" "$HOST_VERSION" "$SKILL_DIGEST" "$CLI_SHA" "$2" "$TS" "$3" \
    >> "$RECEIPT_DIR/receipts.jsonl"
}
# keep <file in $WORK>: copies a raw output next to the receipts and prints its path.
keep() {
  local dest="$RECEIPT_DIR/$STAMP-$1"
  cp "$WORK/$1" "$dest" 2>/dev/null || : > "$dest"
  printf '%s' "$dest"
}

# AC-08C-4 (INVOKE): `$moai` must explicitly invoke the skill and reach the SETUP
# route — verified from the JSONL structure, never from the prose. Every condition
# below must hold, otherwise the receipt is failed (#2014 agent-8 N3):
#   1. codex exec exited 0;
#   2. the stream has a `turn.completed` event and no `turn.failed` or `error` event;
#   3. the last message file is non-empty;
#   4. route evidence: `command_execution` items whose `command` names the projected
#      skill's references/routes.json and workflows/setup.md (the skill text says to
#      pick the route from references/routes.json, then load that route's workflow).
# Conditions 1-3 use the documented event types (thread.started, turn.started,
# turn.completed, turn.failed, item.*, error). Condition 4 rests on an UNVERIFIED
# assumption: that Codex reads those files through shell `command_execution` items
# carrying a `command` string (the shape of the docs' sample line), and that `$moai`
# makes it follow SKILL.md. Neither was read verbatim from an official source nor
# measured against a real stream. A wrong assumption turns a real pass into failed
# (a false negative); it can never turn a failure into passed. The raw stream is kept
# next to the receipt for that review. HOME is isolated, so no personal copy of the
# skill can produce the evidence. Requires CODEX_API_KEY (the isolated CODEX_HOME
# holds no saved login). Not run by the RC lane.
if ! command -v codex >/dev/null 2>&1; then
  receipt "AC-08C-4" not_run "codex binary not found; command not executed"
elif [ -z "$(printenv CODEX_API_KEY || true)" ]; then
  receipt "AC-08C-4" not_run "credential CODEX_API_KEY not set; command not executed"
else
  (cd "$WORK/project" && HOME="$WORK/home" CODEX_HOME="$WORK/home/.codex" codex exec --json --ephemeral \
      --ignore-user-config --sandbox read-only --skip-git-repo-check \
      -o "$WORK/last.md" \
      '$moai 프로젝트 설정 미리보기만 해줘' \
      < /dev/null > "$WORK/ac4.jsonl" 2> "$WORK/ac4.err")
  RC4=$?
  keep last.md > /dev/null
  keep ac4.err > /dev/null
  route_cmd() { # route_cmd <path suffix>: some command_execution item names the file
    jq -se --arg p "$1" \
      'any(.[] | select((.type? // "") | startswith("item.")) | .item? // empty
           | select(.type? == "command_execution") | .command? | tostring; contains($p))' \
      "$WORK/ac4.jsonl" >/dev/null 2>&1
  }
  if [ "$RC4" -eq 0 ] \
     && jq -se 'any(.[]; .type? == "turn.completed")' "$WORK/ac4.jsonl" >/dev/null 2>&1 \
     && jq -se 'all(.[]; (.type? != "turn.failed") and (.type? != "error"))' "$WORK/ac4.jsonl" >/dev/null 2>&1 \
     && [ -s "$WORK/last.md" ] \
     && route_cmd "references/routes.json" \
     && route_cmd "workflows/setup.md"; then
    receipt "AC-08C-4" passed "$(keep ac4.jsonl)"
  else
    receipt "AC-08C-4" failed "$(keep ac4.jsonl) (exit $RC4)"
  fi
fi

echo "receipts written to $RECEIPT_DIR/receipts.jsonl (AC-08C-4; needs CODEX_API_KEY)"
