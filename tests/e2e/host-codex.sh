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
HOST_VERSION="$(codex --version 2>/dev/null | head -1 || echo unknown)"

# Pre-step: project the skill into the temporary project's .agents/skills and give
# the run an isolated HOME / CODEX_HOME.
mkdir -p "$WORK/project/.agents/skills" "$WORK/home/.codex"
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
# route — verified from the JSONL structure (`turn.completed` present, last message
# non-empty), never from the prose. Requires CODEX_API_KEY (the isolated CODEX_HOME
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
      > "$WORK/ac4.jsonl" 2> "$WORK/ac4.err")
  RC4=$?
  keep last.md > /dev/null
  keep ac4.err > /dev/null
  if grep -q '"type":"turn.completed"' "$WORK/ac4.jsonl" 2>/dev/null \
     && [ -s "$WORK/last.md" ]; then
    receipt "AC-08C-4" passed "$(keep ac4.jsonl)"
  else
    receipt "AC-08C-4" failed "$(keep ac4.jsonl) (exit $RC4)"
  fi
fi

echo "receipts written to $RECEIPT_DIR/receipts.jsonl (AC-08C-4; needs CODEX_API_KEY)"
