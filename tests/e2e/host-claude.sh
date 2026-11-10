#!/usr/bin/env bash
# 08-C T2 — Claude Code host-loading E2E (AC-08C-3/5/6, Issue #2014). The lane does NOT
# run this: it needs the operator's existing Claude Code credentials and writes nothing
# outside the directories it creates. It may run from any directory (the package is
# found next to this script), so the operator can run it from a tree outside the
# checkout. Receipts and the raw outputs they cite land in $RECEIPT_DIR (default:
# ./e2e-receipts, created if missing). Credential VALUES are never written to any
# file — the receipt records only that ANTHROPIC_API_KEY was set.
#
# Docs grounding (code.claude.com/docs/en/headless + /skills, re-read 2026-11-09):
# --bare skips auto-discovery of hooks/skills/commands/agents/plugins/MCP/CLAUDE.md,
# but "A directory you name with --add-dir is a partial exception: bare mode loads
# skills from its .claude/skills/ folder" — so the explicit `/moai` at the start of
# the -p prompt is expanded. Caveats honored here: under --bare the model is not
# given the skill LIST (automatic invocation is impaired; only the explicit prompt
# invocation is measured), and a personal `~/.claude/skills/moai` would shadow the
# added-directory fixture ("personal over project / added directory") — this script
# therefore redirects HOME to the temporary work dir for the -p runs, which both
# removes the shadowing and guarantees no writes to the real user HOME.
# `--output-format stream-json` with `-p` requires `--verbose` (claude 2.1.319:
# "When using --print, --output-format=stream-json requires --verbose"; the headless
# docs' stream-json examples pass it too). Both -p runs read stdin from /dev/null so
# the host does not wait for piped input that never comes (#2291).
#
# Usage: export the existing Claude Code credentials into the environment
# (ANTHROPIC_API_KEY — presence is checked, the value is never read into any file),
# then run: tests/e2e/host-claude.sh [RECEIPT_DIR]
# Result fields follow schemas/e2e-receipt.schema.json (REQ-08C-01).
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RECEIPT_DIR="${1:-./e2e-receipts}"
mkdir -p "$RECEIPT_DIR"
RECEIPT_DIR="$(cd "$RECEIPT_DIR" && pwd)"
TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
WORK="$(mktemp -d /tmp/e2e-claude-XXXXXX)"
trap 'rm -rf "$WORK"' EXIT
SKILL_DIGEST="$(shasum -a 256 "$REPO/skills/moai/SKILL.md" | awk '{print $1}')"
CLI_SHA="$(command -v moai-cli >/dev/null 2>&1 && shasum -a 256 "$(command -v moai-cli)" | awk '{print $1}' || echo absent)"
HOST_VERSION="$(claude --version 2>/dev/null | head -1 || echo unknown)"

# Pre-step: project the skill into the temporary project for both host layouts, and
# give the -p runs an isolated HOME (no personal-skill shadowing, no real-HOME writes).
mkdir -p "$WORK/project/.claude/skills" "$WORK/project/.agents/skills" "$WORK/home"
cp -R "$REPO/skills/moai" "$WORK/project/.claude/skills/moai"
cp -R "$REPO/skills/moai" "$WORK/project/.agents/skills/moai"

receipt() { # receipt <name> <result> <detail>
  printf '{"ac":"%s","os":"%s","arch":"%s","host":"claude-code","host_exact_version":"%s","mode":"local","skill_digest":"%s","moai_cli_sha":"%s","result":"%s","ran_at":"%s","detail":"%s"}\n' \
    "$1" "$(uname -s)" "$(uname -m)" "$HOST_VERSION" "$SKILL_DIGEST" "$CLI_SHA" "$2" "$TS" "$3" \
    >> "$RECEIPT_DIR/receipts.jsonl"
}
# keep <file in $WORK>: copies a raw output next to the receipts (the EXIT trap removes
# $WORK) and prints the kept path, which the receipt cites.
keep() {
  local dest="$RECEIPT_DIR/$STAMP-$1"
  cp "$WORK/$1" "$dest" 2>/dev/null || : > "$dest"
  printf '%s' "$dest"
}

# AC-08C-5: static validity — local frontmatter validation, no credentials needed.
if claude plugin validate "$WORK/project/.claude/skills" > "$WORK/ac5.out" 2>&1; then
  receipt "AC-08C-5" passed "$(keep ac5.out)"
else
  receipt "AC-08C-5" failed "$(keep ac5.out)"
fi

# AC-08C-3 (INVOKE): the skill must actually LOAD and reach the SETUP route — verified
# from the event stream, never from the prose or a file merely existing. The skill's
# own instructions say to pick the route from references/routes.json and then load
# that route's workflow, so a run that loaded the skill and reached SETUP shows tool
# calls naming the fixture's references/routes.json and workflows/setup.md. A tool
# call counts even if dontAsk denied it: the attempt itself shows the skill text was
# followed. A plain answer without those calls fails. Requires ANTHROPIC_API_KEY
# (`claude --bare` reads no OAuth credentials). Not run by the RC lane.
if [ -z "$(printenv ANTHROPIC_API_KEY || true)" ]; then
  receipt "AC-08C-3" not_run "credential ANTHROPIC_API_KEY not set; command not executed"
else
  (cd "$WORK/project" && HOME="$WORK/home" claude --bare \
      -p '/moai 프로젝트 설정 미리보기만 해줘' \
      --add-dir "$WORK/project" \
      --output-format stream-json --verbose \
      --permission-mode dontAsk --permission-prompts none \
      < /dev/null > "$WORK/ac3.jsonl" 2> "$WORK/ac3.err")
  RC3=$?
  # Path suffixes, not absolute paths: /tmp resolves to /private/tmp on macOS and the
  # host may report either form. HOME is isolated, so no personal copy can match.
  if [ "$RC3" -eq 0 ] \
     && jq -se 'any(.[]; .type == "result" and .is_error == false)' "$WORK/ac3.jsonl" >/dev/null 2>&1 \
     && jq -se --arg p ".claude/skills/moai/references/routes.json" \
          'any(.[] | .. | objects | select(.type? == "tool_use") | .input | tostring; contains($p))' \
          "$WORK/ac3.jsonl" >/dev/null 2>&1 \
     && jq -se --arg p ".claude/skills/moai/workflows/setup.md" \
          'any(.[] | .. | objects | select(.type? == "tool_use") | .input | tostring; contains($p))' \
          "$WORK/ac3.jsonl" >/dev/null 2>&1; then
    receipt "AC-08C-3" passed "$(keep ac3.jsonl)"
  else
    keep ac3.err > /dev/null
    receipt "AC-08C-3" failed "$(keep ac3.jsonl) (exit $RC3)"
  fi
fi

# AC-08C-6 (PERMISSION/HA-08): an apply request under dontAsk must come back with at
# least one recorded permission denial and leave the target tree byte-identical. A run
# with no result event (for example a non-zero exit with empty output) or an empty
# `permission_denials` list fails. Same credential precondition as AC-08C-3.
if [ -z "$(printenv ANTHROPIC_API_KEY || true)" ]; then
  receipt "AC-08C-6" not_run "credential ANTHROPIC_API_KEY not set; command not executed"
else
  BEFORE="$(find "$WORK/project" -type f -print0 2>/dev/null | sort -z | xargs -0 shasum -a 256)"
  (cd "$WORK/project" && HOME="$WORK/home" claude --bare \
      -p '/moai 설정 적용까지 진행해줘' \
      --add-dir "$WORK/project" \
      --output-format stream-json --verbose \
      --permission-mode dontAsk --permission-prompts none \
      < /dev/null > "$WORK/ac6.jsonl" 2> "$WORK/ac6.err")
  RC6=$?
  AFTER="$(find "$WORK/project" -type f -print0 2>/dev/null | sort -z | xargs -0 shasum -a 256)"
  if jq -se 'any(.[]; .type == "result" and ((.permission_denials | type) == "array") and ((.permission_denials | length) > 0))' \
       "$WORK/ac6.jsonl" >/dev/null 2>&1 \
     && [ "$BEFORE" = "$AFTER" ]; then
    receipt "AC-08C-6" passed "$(keep ac6.jsonl)"
  else
    keep ac6.err > /dev/null
    receipt "AC-08C-6" failed "$(keep ac6.jsonl) (exit $RC6)"
  fi
fi

echo "receipts written to $RECEIPT_DIR/receipts.jsonl (AC-08C-3/5/6; AC-08C-3/6 need ANTHROPIC_API_KEY)"
