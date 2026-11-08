#!/usr/bin/env bash
# 08-C T2 — Claude Code host-loading E2E (AC-08C-3/5/6, Issue #2014). The lane does NOT
# run this: it needs the operator's existing Claude Code credentials and writes nothing
# outside the directories it creates. Run from the repository root; receipts land in
# $RECEIPT_DIR (default: ./e2e-receipts, created if missing). Credential VALUES are
# never written to any file — the receipt records only that ANTHROPIC_API_KEY was set.
#
# Usage: export the existing Claude Code credentials into the environment
# (ANTHROPIC_API_KEY — presence is checked, the value is never read into any file),
# then run: tests/e2e/host-claude.sh [RECEIPT_DIR]
# Result fields follow schemas/e2e-receipt.schema.json (REQ-08C-01).
set -uo pipefail

RECEIPT_DIR="${1:-./e2e-receipts}"
mkdir -p "$RECEIPT_DIR"
TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
WORK="$(mktemp -d /tmp/e2e-claude-XXXXXX)"
trap 'rm -rf "$WORK"' EXIT
SKILL_DIGEST="$(shasum -a 256 skills/moai/SKILL.md | awk '{print $1}')"
CLI_SHA="$(command -v moai-cli >/dev/null 2>&1 && shasum -a 256 "$(command -v moai-cli)" | awk '{print $1}' || echo absent)"
HOST_VERSION="$(claude --version 2>/dev/null | head -1 || echo unknown)"

# Pre-step: project the skill into the temporary project for both host layouts.
mkdir -p "$WORK/project/.claude/skills" "$WORK/project/.agents/skills"
cp -R skills/moai "$WORK/project/.claude/skills/moai"
cp -R skills/moai "$WORK/project/.agents/skills/moai"

receipt() { # receipt <name> <result> <detail-file>
  printf '{"ac":"%s","os":"%s","arch":"%s","host":"claude-code","host_exact_version":"%s","mode":"local","skill_digest":"%s","moai_cli_sha":"%s","result":"%s","ran_at":"%s","detail":"%s"}\n' \
    "$1" "$(uname -s)" "$(uname -m)" "$HOST_VERSION" "$SKILL_DIGEST" "$CLI_SHA" "$2" "$TS" "$3" \
    >> "$RECEIPT_DIR/receipts.jsonl"
}

# AC-08C-5: static validity — local frontmatter validation, no credentials needed.
if claude plugin validate "$WORK/project/.claude/skills" > "$WORK/ac5.out" 2>&1; then
  receipt "AC-08C-5" passed "$WORK/ac5.out"
else
  receipt "AC-08C-5" failed "$WORK/ac5.out"
fi

# AC-08C-3 (INVOKE): the skill must actually LOAD and reach the SETUP route — verified
# from the JSON result, never from a file merely existing. Requires ANTHROPIC_API_KEY
# (`claude --bare` reads no OAuth credentials). Not run by the RC lane; the operator
# runs this script once outside the checkout and files the receipt.
if [ -z "$(printenv ANTHROPIC_API_KEY || true)" ]; then
  receipt "AC-08C-3" not_run "credential ANTHROPIC_API_KEY not set; command not executed"
else
  if (cd "$WORK/project" && claude --bare \
      -p '/moai 프로젝트 설정 미리보기만 해줘' \
      --add-dir "$WORK/project" \
      --output-format json \
      --permission-mode dontAsk --permission-prompts none \
      > "$WORK/ac3.json" 2> "$WORK/ac3.err"); then
    # Expected: .result is a JSON string mentioning the SETUP route and the loaded
    # skill; assert the structure, never the prose.
    if jq -e '.result | type == "string"' "$WORK/ac3.json" >/dev/null 2>&1; then
      receipt "AC-08C-3" passed "$WORK/ac3.json"
    else
      receipt "AC-08C-3" failed "$WORK/ac3.json"
    fi
  else
    receipt "AC-08C-3" failed "$WORK/ac3.err"
  fi
fi

# AC-08C-6 (PERMISSION/HA-08): an apply request under dontAsk must come back denied and
# leave the target tree byte-identical. Same credential precondition as AC-08C-3.
if [ -z "$(printenv ANTHROPIC_API_KEY || true)" ]; then
  receipt "AC-08C-6" not_run "credential ANTHROPIC_API_KEY not set; command not executed"
else
  BEFORE="$(find "$WORK/project" -type f -print0 2>/dev/null | sort -z | xargs -0 shasum -a 256)"
  (cd "$WORK/project" && claude --bare \
      -p '/moai 설정 적용까지 진행해줘' \
      --add-dir "$WORK/project" \
      --output-format stream-json \
      --permission-mode dontAsk --permission-prompts none \
      > "$WORK/ac6.jsonl" 2> "$WORK/ac6.err" || true)
  AFTER="$(find "$WORK/project" -type f -print0 2>/dev/null | sort -z | xargs -0 shasum -a 256)"
  if [ "$(jq -c 'select(.type=="result") | .permission_denials' "$WORK/ac6.jsonl" 2>/dev/null)" != "null" ] \
     && [ "$BEFORE" = "$AFTER" ]; then
    receipt "AC-08C-6" passed "$WORK/ac6.jsonl"
  else
    receipt "AC-08C-6" failed "$WORK/ac6.jsonl"
  fi
fi

echo "receipts written to $RECEIPT_DIR/receipts.jsonl (AC-08C-3/5/6; AC-08C-3/6 need ANTHROPIC_API_KEY)"
