#!/usr/bin/env bash
# lsel-apply.sh — SPEC-LSEL-LOCAL-EVOLUTION-001 M3 playback-only APPLY consumer.
#
# Reads an already-approved decision.json (path passed as $1) and plays back the
# associated diff.patch onto the working tree. The script consumes a queue of
# approved decisions; it does not invent work. It:
#   1. Validates the target against the frozen allowlist (REQ-LSEL-001 / AC-001).
#   2. Mechanically refuses execution-meta targets that lack a synchronous-approval
#      marker in decision.json (REQ-LSEL-002 / AC-005 D3 self-amending-handcuffs).
#   3. Applies diff.patch via `git apply`, appends an apply-ledger.jsonl row, and
#      commits the change as a single lsel-<proposal-id>-tagged Conventional
#      Commit on the current (feature) branch (REQ-LSEL-004 / AC-013).
#
# No-arg invocation is a clean no-op (returns exit 0) so an empty approved-queue
# does not derail a loop pass.
#
# Exit codes: 0 = applied or no-op; 2 = frozen-path refuse; 3 = execution-meta
# refuse (no approval marker); 1 = other error.
set -euo pipefail

DECISION="${1:-}"

resolve_root() {
  # Prefer an explicit override, then $CLAUDE_PROJECT_DIR, then the git toplevel.
  if [[ -n "${LSEL_PROJECT_ROOT:-}" ]]; then echo "$LSEL_PROJECT_ROOT"; return; fi
  if [[ -n "${CLAUDE_PROJECT_DIR:-}" && -d "${CLAUDE_PROJECT_DIR:-}" ]]; then echo "$CLAUDE_PROJECT_DIR"; return; fi
  git rev-parse --show-toplevel 2>/dev/null || pwd
}
PROJECT_ROOT="$(resolve_root)"
ALLOWLIST="$PROJECT_ROOT/.claude/lsel/frozen-allowlist.json"
REJECT_LOG="$PROJECT_ROOT/.moai/logs/lsel-reject.log"
LEDGER="$PROJECT_ROOT/.moai/state/lsel/apply-ledger.jsonl"

# --- no-arg / empty-queue no-op ---
if [[ -z "$DECISION" ]]; then
  echo "lsel-apply: no decision.json argument — no-op (playback-only)" >&2
  exit 0
fi
if [[ ! -f "$DECISION" ]]; then
  echo "lsel-apply: decision not found: $DECISION" >&2
  exit 1
fi
if [[ ! -f "$ALLOWLIST" ]]; then
  echo "lsel-apply: frozen allowlist not found at $ALLOWLIST" >&2
  exit 1
fi

mkdir -p "$(dirname "$REJECT_LOG")" "$(dirname "$LEDGER")"

# --- portable field extraction (no jq dependency) ---
json_str_field() {
  # $1 key, $2 file → first string value
  grep -m1 "\"$1\"" "$2" 2>/dev/null \
    | sed 's/.*"'"$1"'"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' \
    | head -1
}
has_sync_approval() {
  # synchronous-approval marker = both the key and an approved decision present
  grep -q '"synchronous_approval"' "$1" && \
    grep -q '"decision"[[:space:]]*:[[:space:]]*"approved"' "$1"
}

PROPOSAL_ID="$(json_str_field proposal_id "$DECISION")"
TARGET="$(json_str_field target_surface "$DECISION")"
PROPOSAL_DIR="$(dirname "$DECISION")"
DIFF_PATCH="$PROPOSAL_DIR/diff.patch"
TS="$(date -u +%FT%TZ)"

if [[ -z "$PROPOSAL_ID" || -z "$TARGET" ]]; then
  echo "lsel-apply: decision.json missing proposal_id or target_surface" >&2
  exit 1
fi

# --- (1) frozen-allowlist hard-reject (REQ-LSEL-001 / AC-LSEL-001) ---
frozen_match() {
  python3 - "$ALLOWLIST" "$TARGET" <<'PY'
import json, re, sys
allow, target = sys.argv[1], sys.argv[2]
try:
    data = json.load(open(allow))
except Exception:
    sys.exit(2)
for pat in data.get("frozen_patterns", []):
    try:
        if re.search(pat, target):
            sys.exit(0)
    except re.error:
        continue
sys.exit(1)
PY
}
if frozen_match; then
  echo "$TS proposal=$PROPOSAL_ID target=$TARGET category=frozen-path reason=allowlist-hard-reject" >> "$REJECT_LOG"
  echo "lsel-apply: REFUSED (frozen-path) target=$TARGET proposal=$PROPOSAL_ID" >&2
  exit 2
fi

# --- (2) execution-meta forced-gate (REQ-LSEL-002 / AC-LSEL-005 D3) ---
exec_meta_match() {
  python3 - "$ALLOWLIST" "$TARGET" <<'PY'
import json, os, re, sys
allow, target = sys.argv[1], sys.argv[2]
try:
    data = json.load(open(allow))
except Exception:
    sys.exit(2)
norm = os.path.normpath(target)
for pat in data.get("execution_meta", []):
    p = os.path.normpath(pat)
    # treat execution_meta entries as path prefixes / substring matches
    if norm == p or norm.startswith(p + os.sep) or norm.startswith(p) or p in norm:
        sys.exit(0)
sys.exit(1)
PY
}
if exec_meta_match; then
  if ! has_sync_approval "$DECISION"; then
    echo "$TS proposal=$PROPOSAL_ID target=$TARGET category=execution-meta reason=no-synchronous-approval-marker" >> "$REJECT_LOG"
    echo "lsel-apply: REFUSED (execution-meta, no synchronous-approval marker) target=$TARGET proposal=$PROPOSAL_ID" >&2
    exit 3
  fi
  CATEGORY="execution-meta"
else
  CATEGORY="routine"
fi

# --- (3) apply the diff.patch (playback of an already-approved decision) ---
if [[ ! -f "$DIFF_PATCH" ]]; then
  echo "lsel-apply: no diff.patch alongside decision.json at $DIFF_PATCH" >&2
  exit 1
fi
cd "$PROJECT_ROOT"
if ! git apply --whitespace=nowarn "$DIFF_PATCH"; then
  echo "lsel-apply: git apply failed for $DIFF_PATCH" >&2
  exit 1
fi

# Stage ONLY the paths the patch touches (working-tree hygiene: never git add -A).
PATCH_PATHS="$(grep -E '^\+\+\+ b/' "$DIFF_PATCH" | sed 's|^+++ b/||')"
if [[ -z "$PATCH_PATHS" ]]; then
  echo "lsel-apply: diff.patch declared no target paths" >&2
  exit 1
fi
# shellcheck disable=SC2086
git add $PATCH_PATHS

# --- (4) ledger row (AC-LSEL-013) — placeholder SHA backfilled after commit ---
printf '{"proposal_id":"%s","target_surface":"%s","ts":"%s","result":"applied","commit_sha":"pending-backfill","category":"%s"}\n' \
  "$PROPOSAL_ID" "$TARGET" "$TS" "$CATEGORY" >> "$LEDGER"

# --- (5) commit as one lsel-<proposal-id>-tagged Conventional Commit (REQ-LSEL-004) ---
if ! git commit -q -m "feat(lsel-${PROPOSAL_ID}): LSEL apply — ${TARGET}

lsel-${PROPOSAL_ID}
Playback via lsel-apply.sh (decision.json + diff.patch).

🗿 MoAI"; then
  echo "lsel-apply: git commit failed for proposal=$PROPOSAL_ID" >&2
  exit 1
fi
COMMIT_SHA="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"

# Backfill the real commit SHA into the last ledger row.
python3 - "$LEDGER" "$COMMIT_SHA" <<'PY'
import sys
ledger, sha = sys.argv[1], sys.argv[2]
with open(ledger) as f:
    lines = f.read().splitlines()
if lines:
    lines[-1] = lines[-1].replace('"commit_sha":"pending-backfill"', f'"commit_sha":"{sha}"')
with open(ledger, 'w') as f:
    f.write('\n'.join(lines) + '\n')
PY

echo "lsel-apply: applied proposal=$PROPOSAL_ID target=$TARGET commit=$COMMIT_SHA category=$CATEGORY"
